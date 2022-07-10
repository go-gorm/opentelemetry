package main

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gorm.io/plugin/opentelemetry/tracing"
)

func ConfigureOpentelemetry(ctx context.Context) func() {
	switch {
	case os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT") != "":
		return configureJaeger(ctx)
	default:
		return configureStdout(ctx)
	}
}

func configureJaeger(ctx context.Context) func() {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	provider.RegisterSpanProcessor(bsp)

	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}

func configureStdout(ctx context.Context) func() {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)

	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	provider.RegisterSpanProcessor(bsp)

	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}

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

	shutdown := ConfigureOpentelemetry(ctx)
	defer shutdown()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sr := tracetest.NewSpanRecorder()

	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	if err := db.Use(tracing.NewPlugin(tracing.WithTracerProvider(provider))); err != nil {
		panic(err)
	}

	tracer := otel.Tracer("app_or_package_name")

	ctx, span := tracer.Start(ctx, "root")
	defer span.End()

	var num int
	if err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error; err != nil {
		panic(err)
	}

	PrintTraceID(ctx)
}
