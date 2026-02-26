package tracing

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"regexp"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"gorm.io/plugin/opentelemetry/metrics"
)

var (
	firstWordRegex   = regexp.MustCompile(`^\w+`)
	cCommentRegex    = regexp.MustCompile(`(?is)/\*.*?\*/`)
	lineCommentRegex = regexp.MustCompile(`(?im)(?:--|#).*?$`)
	sqlPrefixRegex   = regexp.MustCompile(`^[\s;]*`)

	dbRowsAffected = attribute.Key("db.rows_affected")
)

type otelPlugin struct {
	provider               trace.TracerProvider
	tracer                 trace.Tracer
	attrs                  []attribute.KeyValue
	excludeQueryVars       bool
	excludeMetrics         bool
	serverAddressProvider  func(dialector gorm.Dialector) string
	recordStackTraceInSpan bool
	queryFormatter         func(query string) string
}

func NewPlugin(opts ...Option) gorm.Plugin {
	p := &otelPlugin{}
	for _, opt := range opts {
		opt(p)
	}

	if p.provider == nil {
		p.provider = otel.GetTracerProvider()
	}
	p.tracer = p.provider.Tracer("gorm.io/plugin/opentelemetry")

	return p
}

func (p otelPlugin) Name() string {
	return "otelgorm"
}

type gormHookFunc func(tx *gorm.DB)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

func (p otelPlugin) Initialize(db *gorm.DB) (err error) {
	if !p.excludeMetrics {
		if sqlDB, err := db.DB(); err == nil {
			metrics.ReportDBStatsMetrics(sqlDB)
		}
	}

	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		{cb.Create().Before("gorm:create"), p.before("gorm.Create"), "before:create"},
		{cb.Create().After("gorm:create"), p.after(), "after:create"},

		{cb.Query().Before("gorm:query"), p.before("gorm.Query"), "before:select"},
		{cb.Query().After("gorm:query"), p.after(), "after:select"},

		{cb.Delete().Before("gorm:delete"), p.before("gorm.Delete"), "before:delete"},
		{cb.Delete().After("gorm:delete"), p.after(), "after:delete"},

		{cb.Update().Before("gorm:update"), p.before("gorm.Update"), "before:update"},
		{cb.Update().After("gorm:update"), p.after(), "after:update"},

		{cb.Row().Before("gorm:row"), p.before("gorm.Row"), "before:row"},
		{cb.Row().After("gorm:row"), p.after(), "after:row"},

		{cb.Raw().Before("gorm:raw"), p.before("gorm.Raw"), "before:raw"},
		{cb.Raw().After("gorm:raw"), p.after(), "after:raw"},
	}

	var firstErr error

	for _, h := range hooks {
		if err := h.callback.Register("otel:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}

	return firstErr
}

type contextWrapper struct {
	context.Context
	parent context.Context
}

func (p *otelPlugin) before(spanName string) gormHookFunc {
	return func(tx *gorm.DB) {
		parentCtx := tx.Statement.Context
		ctx, span := p.tracer.Start(tx.Statement.Context, spanName, trace.WithSpanKind(trace.SpanKindClient))
		tx.Statement.Context = contextWrapper{ctx, parentCtx}
		if p.serverAddressProvider != nil {
			serverAddrAttr := semconv.ServerAddressKey.String(p.serverAddressProvider(tx.Config.Dialector))
			span.SetAttributes(serverAddrAttr)
		}
	}
}

func (p *otelPlugin) after() gormHookFunc {
	return func(tx *gorm.DB) {
		defer func() {
			if c, ok := tx.Statement.Context.(contextWrapper); ok {
				// recover previous context
				tx.Statement.Context = c.parent
			}
		}()

		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			return
		}
		defer span.End(trace.WithStackTrace(p.recordStackTraceInSpan))

		attrs := make([]attribute.KeyValue, 0, len(p.attrs)+4)
		attrs = append(attrs, p.attrs...)

		if sys := dbSystem(tx); sys.Valid() {
			attrs = append(attrs, sys)
		}

		vars := tx.Statement.Vars

		var query string
		if p.excludeQueryVars {
			query = tx.Statement.SQL.String()
		} else {
			query = tx.Dialector.Explain(tx.Statement.SQL.String(), vars...)
		}

		formatQuery := p.formatQuery(query)
		attrs = append(attrs, semconv.DBQueryText(formatQuery))
		operation := dbOperation(formatQuery)
		attrs = append(attrs, semconv.DBOperationName(operation))
		if tx.Statement.Table != "" {
			attrs = append(attrs, semconv.DBCollectionName(tx.Statement.Table))
			// add attr `db.query.summary`
			dbQuerySummary := operation + " " + tx.Statement.Table
			attrs = append(attrs, semconv.DBQuerySummary(dbQuerySummary))

			// according to semconv, we should update the span name here if `db.query.summary`is available
			// Use `db.query.summary` as span name directly here instead of keeping the original span name like `gorm.Query`,
			// as we cannot access the original span name here.
			span.SetName(dbQuerySummary)
		}
		if tx.Statement.RowsAffected != -1 {
			attrs = append(attrs, dbRowsAffected.Int64(tx.Statement.RowsAffected))
		}

		span.SetAttributes(attrs...)
		switch tx.Error {
		case nil,
			gorm.ErrRecordNotFound,
			driver.ErrSkip,
			io.EOF, // end of rows iterator
			sql.ErrNoRows:
			// ignore
		default:
			span.RecordError(tx.Error)
			span.SetStatus(codes.Error, tx.Error.Error())
		}
	}
}

func (p *otelPlugin) formatQuery(query string) string {
	if p.queryFormatter != nil {
		return p.queryFormatter(query)
	}
	return query
}

func dbSystem(tx *gorm.DB) attribute.KeyValue {
	switch tx.Dialector.Name() {
	case "mysql":
		return semconv.DBSystemNameMySQL
	case "mssql":
		return semconv.DBSystemNameMicrosoftSQLServer
	case "postgres", "postgresql":
		return semconv.DBSystemNamePostgreSQL
	case "sqlite":
		return semconv.DBSystemNameSqlite
	case "sqlserver":
		return semconv.DBSystemNameMicrosoftSQLServer
	case "clickhouse":
		return semconv.DBSystemNameClickhouse
	case "spanner":
		return semconv.DBSystemNameGCPSpanner
	default:
		return attribute.KeyValue{}
	}
}

func dbOperation(query string) string {
	s := cCommentRegex.ReplaceAllString(query, "")
	s = lineCommentRegex.ReplaceAllString(s, "")
	s = sqlPrefixRegex.ReplaceAllString(s, "")
	return strings.ToLower(firstWordRegex.FindString(s))
}
