package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"

	"gopkg.in/yaml.v2"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

var json jsoniter.API

func init() {
	json = jsoniter.ConfigCompatibleWithStandardLibrary
}

func main() {
	start := time.Now()

	// TODO: make config file cmd argument
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Panicf("ERR Error opening config file config.yaml: %v", err)
	}

	var config Config
	yaml.UnmarshalStrict([]byte(configFile), &config)

	fmt.Printf("INFO Config: %+v\n", config)

	var metrics Metrics
	metrics.IgnoredMetricCount = make(map[string]int)

	// Read source file
	file, err := os.Open(config.Filename)
	if err != nil {
		log.Panicf("ERR Error opening file %v: %v", config.Filename, err)
	}
	defer file.Close()

	var scanner *bufio.Scanner
	reader := bufio.NewReader(file)
	testBytes, err := reader.Peek(2)
	if err != nil {
		log.Panicf("Error peeking from file: %v", err)
	}

	if testBytes[0] == 31 && testBytes[1] == 139 {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			log.Panicf("Error creating gzip reader: %v", err)
		}
		scanner = bufio.NewScanner(gzipReader)
	} else {
		scanner = bufio.NewScanner(reader)
	}

	// InfluxDB connection
	client := influxdb2.NewClientWithOptions(config.InfluxDB.URL, "",
		influxdb2.DefaultOptions().SetBatchSize(config.InfluxDB.BatchSize))

	defer client.Close()

	writeAPI := client.WriteAPI(config.InfluxDB.Org, config.InfluxDB.Bucket)

	// TODO: make this work with influxdb 1 and 2
	//if config.InfluxDB.CreateDB {
	//	q := client.NewQuery("CREATE DATABASE k6", "", "")
	//	if response, err := client.Query(q); err == nil && response.Error() == nil {
	//		fmt.Printf("Create DB results: %+v\n", response.Results)
	//	}
	//}

	// Create channel for points feeding
	pointsCh := make(chan *write.Point, config.Internal.ChanSize)

	var writerWg sync.WaitGroup
	for t := 0; t < config.Internal.Threads; t++ {
		writerWg.Add(1)
		go func() {
			for p := range pointsCh {
				writeAPI.WritePoint(p)
			}
			writerWg.Done()
		}()
	}

	// Processing
	pointRegexp, _ := regexp.Compile("{\"type\":\"Point\"")
	metricRegexp, _ := regexp.Compile("{\"type\":\"Metric\"")

	var line string

	var p *write.Point
	for scanner.Scan() {
		line = scanner.Text()

		if pointRegexp.MatchString(line) {
			point, duration, err := unmarshalPoint(line)
			if err != nil {
				log.Panic(err)
			}
			metrics.JSONUnmarshal.Dur += duration
			metrics.JSONUnmarshal.Calls++

			if !stringInSlice(point.Metric, config.Metrics) {
				metrics.IgnoredMetricCount[point.Metric]++
				continue
			}

			//fmt.Printf("Queueing %v : %v : %v\n", point.Data.Time.Format("2006-01-02T15:04:05.999999"), point.Metric, point.Data.Value)

			p = influxdb2.NewPoint(
				point.Metric,
				map[string]string{
					"url": point.Data.Tags.URL,
				},
				map[string]interface{}{
					"value": point.Data.Value,
				},
				point.Data.Time,
			)

			metrics.PointsBatched++
			pointsCh <- p
		}

		if metricRegexp.MatchString(line) {
			metrics.SkippedMetricNotPoint++
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	close(pointsCh)
	writeAPI.Flush()

	// Wait for writes complete
	writerWg.Wait()
	metrics.Runt = time.Since(start)

	metrics.Runtime = fmt.Sprintf("%v", metrics.Runt)
	metrics.JSONUnmarshal.Duration = fmt.Sprintf("%v", metrics.JSONUnmarshal.Dur)

	b, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Metrics: %+v\n", string(b))
}

func unmarshalPoint(jsonStr string) (point *K6Point, duration time.Duration, err error) {
	start := time.Now()
	if err := json.Unmarshal([]byte(jsonStr), &point); err != nil {
		return &K6Point{}, time.Since(start), err
	}
	return point, time.Since(start), nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
