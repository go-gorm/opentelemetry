module github.com/uptrace/opentelemetry-go-extra/example/prometheus/metrics

go 1.20

require (
	github.com/prometheus/client_golang v1.15.1
	go.opentelemetry.io/contrib/instrumentation/runtime v0.42.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/jaeger v1.16.0
	go.opentelemetry.io/otel/exporters/prometheus v0.39.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.opentelemetry.io/otel/sdk/metric v0.39.0
	go.opentelemetry.io/otel/trace v1.16.0
	gorm.io/driver/sqlite v1.5.1
	gorm.io/gorm v1.25.1
	gorm.io/plugin/opentelemetry v0.1.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.0 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)

replace gorm.io/plugin/opentelemetry => ./../../..
