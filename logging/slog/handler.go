package slog

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

type traceConfig struct {
	recordStackTraceInSpan bool
	errorSpanLevel         slog.Level
}

type traceHandler struct {
	slog.Handler
	tcfg *traceConfig
}

func NewTraceHandler(w io.Writer, opts *slog.HandlerOptions, traceConfig *traceConfig) *traceHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &traceHandler{
		slog.NewJSONHandler(w, opts),
		traceConfig,
	}
}

func (t *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.Handler.Enabled(ctx, level)
}

func (t *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	// trace span add
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		record.Add(traceIDKey, span.SpanContext().TraceID())
	}
	if span.SpanContext().SpanID().IsValid() {
		record.Add(spanIDKey, span.SpanContext().SpanID())
	}
	if span.SpanContext().TraceFlags().IsSampled() {
		record.Add(traceFlagsKey, span.SpanContext().TraceFlags())
	}

	// non recording spans do not support modifying
	if !span.IsRecording() {
		return t.Handler.Handle(ctx, record)
	}

	// set span status
	if record.Level >= t.tcfg.errorSpanLevel {
		span.SetStatus(codes.Error, "")
		span.RecordError(errors.New(record.Message), trace.WithStackTrace(t.tcfg.recordStackTraceInSpan))
	}

	return t.Handler.Handle(ctx, record)
}

func (t *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return t.Handler.WithAttrs(attrs)
}

func (t *traceHandler) WithGroup(name string) slog.Handler {
	return t.Handler.WithGroup(name)
}
