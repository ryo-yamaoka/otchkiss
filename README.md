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

See [./sample/main.go](./sample/main.go) for a sample.

```
$ go run ./sample/...

[Setting]
* warm up time:   5s
* duration:       1s
* max concurrent: 1

[Request]
* total:      90
* succeeded:  90
* failed:     0
* error rate: 0 %
* RPS:        90

[Latency]
* max: 11.0 ms
* min: 10.0 ms
* avg: 10.9 ms
* med: 11.0 ms
* 99th percentile: 11.0 ms
* 90th percentile: 11.0 ms
```

## Development

* Lint: `make lint`
    * install lint tools: `make install-tools`
* Test: `make test`

## Author

[Ryo Yamaoka](https://github.com/ryo-yamaoka)

* [Twitter](https://twitter.com/mountainhill14)
