package slog

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger})
	if err != nil {
		panic(err)
	}

	db.Logger.Info(ctx, "log from origin logrus")

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	db.Logger.Info(ctx, "hello %s", "world")

	span.End()

	ctx, child := tracer.Start(ctx, "child")

	db.Logger.Warn(ctx, "foo %s", "bar")

	child.End()

	ctx, errSpan := tracer.Start(ctx, "error")

	db.Logger.Error(ctx, "error %s", "this is a error")

	db.Logger.Info(ctx, "no trace context")

	errSpan.End()
}

func TestLogger1(t *testing.T) {
	ctx := context.Background()

	buf := new(bytes.Buffer)

	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := logger.New(
		NewWriter(
			WithTraceErrorSpanLevel(slog.LevelWarn),
			WithRecordStackTraceInSpan(true),
		), logger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      logger.Warn,
			Colorful:      false,
		})

	logger.Info(ctx, "log from origin slog")
	assert.True(t, strings.Contains(buf.String(), "log from origin slog"))
	buf.Reset()

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	assert.True(t, strings.Contains(buf.String(), "trace_id"))
	assert.True(t, strings.Contains(buf.String(), "span_id"))
	assert.True(t, strings.Contains(buf.String(), "trace_flags"))
	buf.Reset()

	span.End()

	ctx, child := tracer.Start(ctx, "child")

	child.End()

	_, errSpan := tracer.Start(ctx, "error")

	errSpan.End()
}

func TestLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewWriter(
		WithTraceErrorSpanLevel(slog.LevelWarn),
		WithRecordStackTraceInSpan(true),
	)

	/*	// output to buffer
		logger.SetOutput(buf)

		logger.Debug("this is a debug log")
		assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

		logger.SetLevel(klog.LevelDebug)
	*/
	logger.log.Debug("this is a debug log msg")
	assert.True(t, strings.Contains(buf.String(), "this is a debug log"))
}

func TestLogOption(t *testing.T) {
	buf := new(bytes.Buffer)

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	logger := NewWriter(
		WithLevel(lvl),
		WithOutput(buf),
		WithHandlerOptions(&slog.HandlerOptions{
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.MessageKey {
					msg := a.Value.Any().(string)
					msg = strings.ReplaceAll(msg, "log", "new log")
					a.Value = slog.StringValue(msg)
				}
				return a
			},
		}),
	)

	logger.log.Debug("this is a debug log")
	assert.True(t, strings.Contains(buf.String(), "this is a debug new log"))

	dir, _ := os.Getwd()
	assert.True(t, strings.Contains(buf.String(), dir))
}
