module gorm.io/plugin/opentelemetry

go 1.18

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	go.opentelemetry.io/contrib/instrumentation/runtime v0.38.0
	go.opentelemetry.io/contrib/propagators/b3 v1.13.0
	go.opentelemetry.io/contrib/propagators/jaeger v1.13.0
	go.opentelemetry.io/contrib/propagators/opencensus v0.39.0
	go.opentelemetry.io/contrib/propagators/ot v1.13.0
	go.opentelemetry.io/otel v1.13.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.13.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.13.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.13.0
	go.opentelemetry.io/otel/metric v0.36.0
	go.opentelemetry.io/otel/sdk v1.13.0
	go.opentelemetry.io/otel/sdk/metric v0.36.0
	go.opentelemetry.io/otel/trace v1.13.0
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.5
)

require (
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.12 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.36.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
	google.golang.org/grpc v1.52.3 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
