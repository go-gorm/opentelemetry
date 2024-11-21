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

func TestLoggerBase(t *testing.T) {
	ctx := context.Background()

	buf := new(bytes.Buffer)

	shutdown := stdoutProvider(ctx)
	defer shutdown()

	writer := NewWriter(
		WithTraceErrorSpanLevel(slog.LevelWarn),
		WithOutput(buf),
		WithRecordStackTraceInSpan(true),
	)

	writer.log.InfoContext(ctx, "log from origin slog")
	assert.True(t, strings.Contains(buf.String(), "log from origin slog"))
	buf.Reset()

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	writer.log.WarnContext(ctx, "hello world")
	assert.True(t, strings.Contains(buf.String(), "trace_id"))
	assert.True(t, strings.Contains(buf.String(), "span_id"))
	assert.True(t, strings.Contains(buf.String(), "trace_flags"))
	buf.Reset()

	span.End()

	ctx, child := tracer.Start(ctx, "child")
	logger := logger.New(
		NewWriter(),
		logger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)
	logger.Info(ctx, "Info %s", "this is a info log")
	logger.Warn(ctx, "warn %s", "this is a warn log")
	logger.Error(ctx, "error %s", "this is a error log")
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

	logger.SetOutput(buf)
	logger.log.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.SetLvel(slog.LevelDebug)

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
