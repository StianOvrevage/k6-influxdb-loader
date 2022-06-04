package main

import "time"

type Config struct {
	Filename string `yaml:"metrics_filename"`
	InfluxDB struct {
		URL       string `yaml:"url"`
		Bucket    string `yaml:"bucket"`
		Org       string `yaml:"org"`
		BatchSize uint   `yaml:"batchsize"`
		CreateDB  bool   `yaml:"create_db"`
	} `yaml:"influxdb"`
	Internal struct {
		ChanSize int `yaml:"chan_size"`
		Threads  int `yaml:"threads"`
	} `yaml:"internal"`
	Metrics []string `yaml:"metrics"`
}

type Metrics struct {
	SkippedMetricNotPoint int
	PointsBatched         int
	IgnoredMetricCount    map[string]int
	JSONUnmarshal         struct {
		Dur      time.Duration `json:"-"`
		Duration string
		Calls    int64
	}
	Runt    time.Duration `json:"-"`
	Runtime string
}

type K6Metric struct {
	Type string `json:"type"`
	Data struct {
		Name       string        `json:"name"`
		Type       string        `json:"type"`
		Contains   string        `json:"contains"`
		Tainted    bool          `json:"tainted"`
		Thresholds []interface{} `json:"thresholds"`
		Submetrics interface{}   `json:"submetrics"`
		Sub        struct {
			Name   string      `json:"name"`
			Parent string      `json:"parent"`
			Suffix string      `json:"suffix"`
			Tags   interface{} `json:"tags"`
		} `json:"sub"`
		Metric string `json:"metric"`
	} `json:"data"`
}

type K6Point struct {
	Type string `json:"type"`
	Data struct {
		Time  time.Time `json:"time"`
		Value float32   `json:"value"`
		Tags  struct {
			ExpectedResponse string `json:"expected_response"` // Or bool?
			Group            string `json:"group"`
			Method           string `json:"method"`
			Name             string `json:"name"`
			Proto            string `json:"proto"`
			Scenario         string `json:"scenario"`
			Status           string `json:"status"` // Or int?
			TLSVersion       string `json:"tls_version"`
			URL              string `json:"url"`
		} `json:"tags"`
	} `json:"data"`
	Metric string `json:"metric"`
}
