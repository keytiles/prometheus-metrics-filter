#!/bin/bash

echo 'building docker image...'
docker build --file docker/service-for-scylladb.dockerfile -t keytiles/prometheus-metrics-scylladb-filter:1.0.0 .

echo 'all done'