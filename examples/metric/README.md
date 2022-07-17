# Metrics example

## Setup Basic Dependence and Run demo
```shell
cd examples/metric
docker-compose up -d
```
## View Metrics
You can then navigate to http://localhost:9090 to access the Prometheus.
Dockers starts 2 services:

- https://localhost:8088/metrics - Prometheus target that exports OpenTelemetry metrics.
- http://localhost:9090/graph - Prometheus configured to scrape our target.

You can use Prometheus to query counters:
```
go_sql_connections_max_idle
```

### Screenshots![prometheus.png](static/prometheus.png)