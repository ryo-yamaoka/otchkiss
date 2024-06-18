# Otchkiss

[![CI](https://github.com/ryo-yamaoka/otchkiss/actions/workflows/go.yml/badge.svg)](https://github.com/ryo-yamaoka/otchkiss/actions/workflows/go.yml)
[![license: Apache](https://img.shields.io/badge/license-Apache-blue.svg?style=flat-square)](LICENSE)

## Features

Otchkiss is a simple load testing library.

* Easy to make single node load testing.
* Easy to display result some statistics by default format.
* Flexible displaying results by user template.

## Usage

`go get github.com/ryo-yamaoka/otchkiss`

See [./sample/main.go](./sample/main.go) for a sample implementation.

```
[Setting]
* warm up time:   3s
* duration:       3s
* max concurrent: 0
* max RPS:        0

[Request]
* total:      25
* succeeded:  25
* failed:     0
* error rate: 0 %
* RPS:        8.3

[Latency]
* max: 800.0 ms
* min: 1.0 ms
* avg: 390.0 ms
* med: 400.0 ms
* 99th percentile: 700.0 ms
* 90th percentile: 600.0 ms

[Histogram]
0s-88.888888ms             4%   █████▏                      1
88.888888ms-177.777777ms   8%   ██████████▏                 2
177.777777ms-266.666666ms  12%  ███████████████▏            3
266.666666ms-355.555555ms  16%  ████████████████████▏       4
355.555555ms-444.444444ms  20%  █████████████████████████▏  5
444.444444ms-533.333333ms  16%  ████████████████████▏       4
533.333333ms-622.222222ms  12%  ███████████████▏            3
622.222222ms-711.111111ms  8%   ██████████▏                 2
711.111111ms-800ms         4%   █████▏                      1
```

### Command line options

When you useing `otchkiss.New()` or `setting.FromDefaultFlag()`, will be parsed following command line parameters.

This eliminates the need to write the parsing process.

* `-p`: Specify the number of parallels executions. `0` means unlimited (default: `1`, it's not concurrently)
* `-d`: Running duration, ex: 300s or 5m etc... (default: `5s`)
* `-w`: Exclude from results for a given time after startup, ex: 300s or 5m etc... (default: `5s`)
* `-r`: Specify the max request per second. 0 means unlimited (default: `1`)

## Development

* Lint: `make lint`
    * install lint tools: `make install-tools`
* Test: `make test`

## Author

[Ryo Yamaoka](https://github.com/ryo-yamaoka)

* [X](https://twitter.com/mountainhill14)
* [misskey.io](https://misskey.io/@r_yamaoka)
