module gorm.io/plugin/example

go 1.20

replace gorm.io/plugin/opentelemetry => ./../..

require (
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/jaeger v1.16.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
	gorm.io/driver/sqlite v1.5.1
	gorm.io/gorm v1.25.1
	gorm.io/plugin/opentelemetry v0.1.2
)

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)
