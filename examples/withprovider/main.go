package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/plugin/opentelemetry/logging/logrus"
	"gorm.io/plugin/opentelemetry/provider"
	"gorm.io/plugin/opentelemetry/tracing"
)

func PrintTraceID(ctx context.Context) {
	fmt.Println("trace:", TraceURL(trace.SpanFromContext(ctx)))
}
func TraceURL(span trace.Span) string {
	switch {
	case os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT") != "":
		return fmt.Sprintf("http://localhost:16686/trace/%s", span.SpanContext().TraceID())
	default:
		return fmt.Sprintf("http://localhost:16686/trace/%s", span.SpanContext().TraceID())
	}
}

func main() {
	ctx := context.Background()

	serviceName := "examples"

	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint("host.docker.internal:4317"),
		provider.WithInsecure(),
	)

	defer p.Shutdown(context.Background())

	logger := logger.New(
		logrus.NewWriter(),
		logger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger})
	if err != nil {
		panic(err)
	}

	if err = db.Use(tracing.NewPlugin()); err != nil {
		panic(err)
	}

	tracer := otel.Tracer("gorm.io/plugin/opentelemetry")

	ctx, span := tracer.Start(ctx, "root")
	defer span.End()

	var num int
	if err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error; err != nil {
		panic(err)
	}

	PrintTraceID(ctx)

}
