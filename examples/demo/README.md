Stdout exporter (default):

go run .

Jaeger exporter:

docker: `docker-compose up -d`

run: `OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces go run .`
