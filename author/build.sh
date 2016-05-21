#!/bin/bash

VER=$1
if [ -z "$1" ] ; then
  echo 'Missing version'
  exit 1
fi

REV=$(git rev-parse --verify HEAD)
PACKAGE="github.com/moznion/resque_exporter"

go get -v gopkg.in/yaml.v2
go get -v gopkg.in/redis.v3
go get -v github.com/mkideal/cli
go get -v github.com/prometheus/client_golang/prometheus

for GOOS in darwin linux; do
  for GOARCH in 386 amd64; do
    export GOOS
    export GOARCH
    go build -v -ldflags "-X ${PACKAGE}.rev=$REV -X ${PACKAGE}.ver=$VER" \
      -o bin/resque_exporter-$GOOS-$GOARCH cmd/resque_exporter/resque_exporter.go
  done
done

