package logrus

import (
	"errors"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/overview.md#json-formats
const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
	logEventKey   = "log"
)

var (
	logSeverityTextKey = attribute.Key("otel.log.severity.text")
	logMessageKey      = attribute.Key("otel.log.message")
)

type TraceHookConfig struct {
	recordStackTraceInSpan bool
	enableLevels           []logrus.Level
	errorSpanLevel         logrus.Level
}

type TraceHook struct {
	cfg *TraceHookConfig
}

func NewTraceHook(cfg *TraceHookConfig) *TraceHook {
	return &TraceHook{cfg: cfg}
}

func (h *TraceHook) Levels() []logrus.Level {
	return h.cfg.enableLevels
}

func (h *TraceHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	span := trace.SpanFromContext(entry.Context)
	if !span.IsRecording() {
		return nil
	}

	// attach span context to log entry data fields
	entry.Data[traceIDKey] = span.SpanContext().TraceID()
	entry.Data[spanIDKey] = span.SpanContext().SpanID()
	entry.Data[traceFlagsKey] = span.SpanContext().TraceFlags()

	// attach log to span event attributes
	attrs := []attribute.KeyValue{
		logMessageKey.String(entry.Message),
		logSeverityTextKey.String(OtelSeverityText(entry.Level)),
	}
	span.AddEvent(logEventKey, trace.WithAttributes(attrs...))

	// set span status
	if entry.Level <= h.cfg.errorSpanLevel {
		span.SetStatus(codes.Error, entry.Message)
		span.RecordError(errors.New(entry.Message), trace.WithStackTrace(h.cfg.recordStackTraceInSpan))
	}

	return nil
}
