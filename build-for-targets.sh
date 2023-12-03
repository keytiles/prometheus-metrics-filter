#!/bin/bash

echo 'building for all targets...'
env GOOS=linux GOARCH=amd64 go build -o bin/prometheus-metrics-filter.linux.amd64
env GOOS=linux GOARCH=arm go build -o bin/prometheus-metrics-filter.linux.arm
env GOOS=windows GOARCH=amd64 go build -o bin/prometheus-metrics-filter.windows.amd64.exe
env GOOS=windows GOARCH=arm go build -o bin/prometheus-metrics-filter.windows.arm.exe

echo 'all done'