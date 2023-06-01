package provider

import (
	"context"

	"github.com/sirupsen/logrus"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type OtelProvider interface {
	Shutdown(ctx context.Context) error
}

type otelProvider struct {
	traceExp      sdktrace.SpanExporter
	metricsPusher sdkmetric.Exporter
}

func (p *otelProvider) Shutdown(ctx context.Context) error {
	var err error

	if p.traceExp != nil {
		if err = p.traceExp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	if p.metricsPusher != nil {
		if err = p.metricsPusher.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return err
}

// NewOpenTelemetryProvider Initializes an otlp trace and metrics provider
func NewOpenTelemetryProvider(opts ...Option) OtelProvider {
	var (
		err       error
		traceExp  sdktrace.SpanExporter
		metricExp sdkmetric.Exporter
	)

	ctx := context.TODO()

	cfg := newConfig(opts)

	if !cfg.enableTracing && !cfg.enableMetrics {
		return nil
	}

	// resource
	res := newResource(cfg)

	// propagator
	otel.SetTextMapPropagator(cfg.textMapPropagator)

	// Tracing
	if cfg.enableTracing {
		// trace client
		var traceClientOpts []otlptracegrpc.Option
		if cfg.exportEndpoint != "" {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithInsecure())
		}

		traceClient := otlptracegrpc.NewClient(traceClientOpts...)

		// trace exporter
		traceExp, err = otlptrace.New(ctx, traceClient)
		if err != nil {
			logrus.Fatalf("failed to create otlp trace exporter: %s", err)
			return nil
		}

		// trace processor
		bsp := sdktrace.NewBatchSpanProcessor(traceExp)

		// trace provider
		tracerProvider := cfg.sdkTracerProvider
		if tracerProvider == nil {
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
				sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(bsp),
			)
		}

		otel.SetTracerProvider(tracerProvider)
	}

	// Metrics
	if cfg.enableMetrics {
		metricsClientOpts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithAggregationSelector(sdkmetric.DefaultAggregationSelector),
		}
		if cfg.exportEndpoint != "" {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithInsecure())
		}

		// metrics exporter
		metricExp, err = otlpmetricgrpc.New(ctx,
			metricsClientOpts...,
		)
		handleInitErr(err, "Failed to create the collector metric exporter")

		// metrics pusher
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		)
		otel.SetMeterProvider(meterProvider)

		err = runtimemetrics.Start()
		handleInitErr(err, "Failed to start runtime metrics collector")
	}

	return &otelProvider{
		traceExp:      traceExp,
		metricsPusher: metricExp,
	}
}

func newResource(cfg *config) *resource.Resource {
	if cfg.resource != nil {
		return cfg.resource
	}

	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithProcessPID(),
		resource.WithTelemetrySDK(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithDetectors(cfg.resourceDetectors...),
		resource.WithAttributes(cfg.resourceAttributes...),
	)
	if err != nil {
		otel.Handle(err)
		return resource.Default()
	}
	return res
}

func handleInitErr(err error, message string) {
	if err != nil {
		logrus.Fatalf("%s: %v", message, err)
	}
}
