
build:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/k6-influxdb-loader-linux-amd64 main.go models.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/k6-influxdb-loader-linux-arm64 main.go models.go

	GOOS=darwin GOARCH=amd64 go build -o dist/k6-influxdb-loader-macos-amd64 main.go models.go
	GOOS=darwin GOARCH=arm64 go build -o dist/k6-influxdb-loader-macos-arm64 main.go models.go

	GOOS=windows GOARCH=amd64 go build -o dist/k6-influxdb-loader-windows-amd64.exe main.go models.go
