package logrus

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/gorm/logger"
)

func stdoutProvider(ctx context.Context) func() {
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

func TestLogger(t *testing.T) {
	ctx := context.Background()
	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := logger.New(
		NewWriter(),
		logger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      logger.Warn,
			Colorful:      false,
		},
	)

	logger.Info(ctx, "log from origin logrus")

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	logger.Info(ctx, "hello %s", "world")

	span.End()

	ctx, child := tracer.Start(ctx, "child")

	logger.Warn(ctx, "foo %s", "bar")

	child.End()

	ctx, errSpan := tracer.Start(ctx, "error")

	logger.Error(ctx, "error %s", "this is a error")

	logger.Info(ctx, "no trace context")

	errSpan.End()
}
