# prometheus-metrics-filter

Acts as a proxy between Prometheus and a Metrics endpoint. This service is capable of scraping given Prometheus Metrics endpoint and then filter the results based on rules before returned.

# Why did we build this?

We are running ScyllaDB and Scylla Monitoring Stack. We needed table level metrics in Scylla but when enabled it is really blowing up the number of exposed metrics - exactly the way you are
warned by Scylla developers: https://github.com/scylladb/scylladb/blob/master/docs/dev/metrics.md#per-table-metrics

However there are possibilities to configure Prometheus to drop certain metrics but this is not easy to configure it in Scylla Monitoring Stack (we do not have direct config possibility). If we
would do so that would increase the maintenance efforts on Scylla Monitoring Stack side drastically in terms of updating/upgrading - better not to touch that...

To mitigate the problem it seemed to be much easier to create a simple proxy service which can be "injected" between Prometheus and the Metrics endpoint and can do filtering based on rules.
