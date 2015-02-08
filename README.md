# Tapglue backend 

This repository contains the implementation of tapglues backend.

[![wercker status](https://app.wercker.com/status/37a8675b2ae12075851f297ce6a36ead/s/master "wercker status")](https://app.wercker.com/project/bykey/37a8675b2ae12075851f297ce6a36ead)
[![codecov.io](https://codecov.io/github/Tapglue/backend/coverage.svg?token=OHlqgNOv66&branch=HEAD)](https://codecov.io/github/Tapglue/backend?branch=HEAD)

## Documentation

See [Documentation](https://github.com/tapglue/backend/wiki) for entities, api design and more.

## System Requirements

- go (latest)
- redis 2.8 or newer

## Installation

Following steps are need to download and install this project.

### Getting started

Download the git repository to get started.

```shell
$ mkdir -p $GOPATH/src/github.com/tapglue/backend
$ git clone https://github.com/tapglue/backend.git
$ cd backend
```

### Dependencies

All dependecies should be fecthed correctly by running:

```shell
$ go get github.com/tapglue/backend
```

or, if you cloned it locally in your GOPATH

```shell
$ cd $GOPATH/src/github.com/tapglue/backend
$ go get ./...
```

### Server configuration

Configure the server including ports and database settings in the [config.json](config.json).

```json
{
  "env": "dev",
  "use_artwork": false,
  "listenHost": ":8082",
  "newrelic": {
    "key": "",
    "name": "dev - tapglue",
    "enabled": false
  },
  "redis": {
    "hosts": [
      "127.0.0.1:6379"
    ],
    "password": "",
    "db": 0,
    "pool_size": 30
  }
}
```

### Start server

```shell
$ go run -race backend.go
```

## Tests

```shell
$ cd core
$ go test -check.v
$ cd ../server
$ go test -check.v
```

## Coverage

```shell
$ bin/coverage/*.sh
```

## Benchmarks

```shell
$ cd core
$ go test -bench=. -benchmem
$ cd ../server
$ go test -bench=. -benchmem
```

## Profilling

```shell
$ bin/ab/*.sh
```

## Code commit

Before doing a commit, please run the following in the ```$GOPATH/src/github.com/tapglue/backend```  
```shell
goimports -w ./.. && golint ./... && go vet ./...
```

## Deployment

TBD
