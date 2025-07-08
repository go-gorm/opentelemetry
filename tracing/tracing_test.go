package tracing

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Test struct {
	opts    []Option
	do      func(ctx context.Context, db *gorm.DB)
	require func(t *testing.T, spans []sdktrace.ReadOnlySpan)
}

func TestOtel(t *testing.T) {
	tests := []Test{
		{
			do: func(ctx context.Context, db *gorm.DB) {
				var num int
				err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error
				require.NoError(t, err)
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				require.Equal(t, 1, len(spans))
				require.Equal(t, "gorm.Row", spans[0].Name())
				require.Equal(t, trace.SpanKindClient, spans[0].SpanKind())

				m := attrMap(spans[0].Attributes())

				sys, ok := m[semconv.DBSystemNameKey]
				require.True(t, ok)
				require.Equal(t, "sqlite", sys.AsString())

				stmt, ok := m[semconv.DBQueryTextKey]
				require.True(t, ok)
				require.Equal(t, "SELECT 42", stmt.AsString())

				operation, ok := m[semconv.DBOperationNameKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
		},
		{
			do: func(ctx context.Context, db *gorm.DB) {
				var num int
				_ = db.WithContext(ctx).Raw("SELECT foo_bar").Scan(&num).Error
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				require.Equal(t, 1, len(spans))
				require.Equal(t, "gorm.Row", spans[0].Name())

				span := spans[0]
				status := span.Status()
				require.Equal(t, codes.Error, status.Code)
				require.Equal(t, "no such column: foo_bar", status.Description)

				m := attrMap(span.Attributes())

				sys, ok := m[semconv.DBSystemNameKey]
				require.True(t, ok)
				require.Equal(t, "sqlite", sys.AsString())

				stmt, ok := m[semconv.DBQueryTextKey]
				require.True(t, ok)
				require.Equal(t, "SELECT foo_bar", stmt.AsString())

				operation, ok := m[semconv.DBOperationNameKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
		},
		{
			opts: []Option{WithoutQueryVariables()},
			do: func(ctx context.Context, db *gorm.DB) {
				err := db.Exec("CREATE TABLE foo (id int)").Error
				require.NoError(t, err)
				var num int
				param := 42
				err = db.WithContext(ctx).Table("foo").Select("id", param).Where("id = ?", param).Scan(&num).Error
				require.NoError(t, err)
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				for _, s := range spans {
					fmt.Printf("span=%#v\n", s)
				}
				require.Equal(t, 2, len(spans))
				require.Equal(t, "select foo", spans[1].Name())
				require.Equal(t, trace.SpanKindClient, spans[1].SpanKind())

				m := attrMap(spans[1].Attributes())

				sys, ok := m[semconv.DBSystemNameKey]
				require.True(t, ok)
				require.Equal(t, "sqlite", sys.AsString())

				stmt, ok := m[semconv.DBQueryTextKey]
				require.True(t, ok)
				require.Equal(t, "SELECT id FROM `foo` WHERE id = ?", stmt.AsString())

				operation, ok := m[semconv.DBOperationNameKey]
				require.True(t, ok)
				require.Equal(t, "select", operation.AsString())
			},
		},
		{
			do: func(ctx context.Context, db *gorm.DB) {
				var num int
				db.Config.Dialector = &postgres.Dialector{
					Config: &postgres.Config{
						DSN: "test.dsn",
					},
				}
				err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error
				require.NoError(t, err)
			},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				m := attrMap(spans[0].Attributes())
				require.Equal(t, "test.dsn", m[semconv.ServerAddressKey].AsString())
			},
		},
		{
			do: func(ctx context.Context, db *gorm.DB) {
				var num int
				db.Config.Dialector = &postgres.Dialector{
					Config: &postgres.Config{
						DSN: "test.dsn",
					},
				}
				err := db.WithContext(ctx).Raw("SELECT 42").Scan(&num).Error
				require.NoError(t, err)
			},
			opts: []Option{WithoutServerAddress()},
			require: func(t *testing.T, spans []sdktrace.ReadOnlySpan) {
				m := attrMap(spans[0].Attributes())
				require.Equal(t, "", m[semconv.ServerAddressKey].AsString())
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			sr := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

			db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
			require.NoError(t, err)

			err = db.Use(NewPlugin(append(test.opts, WithTracerProvider(provider))...))
			require.NoError(t, err)

			test.do(context.TODO(), db)

			spans := sr.Ended()
			test.require(t, spans)
		})
	}
}

func TestOtelPlugin_ContextRestoration(t *testing.T) {
	provider := noop.NewTracerProvider()
	p := &otelPlugin{provider: provider}
	p.tracer = provider.Tracer("test")

	origCtx := context.WithValue(context.Background(), "foo", "bar")
	db := &gorm.DB{
		Statement: &gorm.Statement{Context: origCtx},
		Config:    &gorm.Config{},
	}

	// before should wrap context
	before := p.before("test-span")
	before(db)
	cw, ok := db.Statement.Context.(contextWrapper)
	require.True(t, ok)
	require.Equal(t, origCtx, cw.parent)

	// after should restore context
	after := p.after()
	after(db)
	require.Equal(t, origCtx, db.Statement.Context)

	origCtx = context.Background()
	db = &gorm.DB{
		Statement: &gorm.Statement{Context: origCtx},
		Config:    &gorm.Config{},
	}

	// after should not panic if context is not a contextWrapper
	after = p.after()
	require.NotPanics(t, func() { after(db) })
	require.Equal(t, origCtx, db.Statement.Context)
}

func attrMap(attrs []attribute.KeyValue) map[attribute.Key]attribute.Value {
	m := make(map[attribute.Key]attribute.Value, len(attrs))
	for _, kv := range attrs {
		m[kv.Key] = kv.Value
	}
	return m
}

// TestExtractServerAddress tests the extractServerAddress function with various DSN formats
func TestExtractServerAddress(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		// Common DSN formats
		{
			name:     "ClickHouse URL with credentials",
			dsn:      "clickhouse://user:password@127.0.0.1:9000/database",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "ClickHouse URL without credentials",
			dsn:      "clickhouse://127.0.0.1:9000/database",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "HTTP URL with credentials",
			dsn:      "http://user:password@127.0.0.1:8123/database",
			expected: "127.0.0.1:8123",
		},
		{
			name:     "HTTPS URL with credentials",
			dsn:      "https://user:password@127.0.0.1:8443/database",
			expected: "127.0.0.1:8443",
		},
		{
			name:     "TCP URL with query parameters",
			dsn:      "tcp://127.0.0.1:9000?database=default&username=user&password=secret",
			expected: "127.0.0.1:9000",
		},

		// Uncommon DSNs
		{
			name:     "Traditional format with credentials",
			dsn:      "user:password@127.0.0.1:9000/database",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "Simple host:port format",
			dsn:      "127.0.0.1:9000",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "Host with database",
			dsn:      "127.0.0.1:9000/database",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "Host with query parameters",
			dsn:      "127.0.0.1:9000?database=test&timeout=10s",
			expected: "127.0.0.1:9000",
		},
		{
			name:     "Complex format with all parts",
			dsn:      "user:password@127.0.0.1:9000/database?param1=value1&param2=value2",
			expected: "127.0.0.1:9000",
		},

		{
			name:     "Empty DSN",
			dsn:      "",
			expected: "",
		},
		{
			name:     "IPv6 address with credentials",
			dsn:      "clickhouse://user:pass@[::1]:9000/db",
			expected: "[::1]:9000",
		},
		{
			name:     "IPv6 address without credentials",
			dsn:      "tcp://[2001:db8::1]:9000?database=test",
			expected: "[2001:db8::1]:9000",
		},
		{
			name:     "Host without port",
			dsn:      "clickhouse://user:pass@localhost/db",
			expected: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractServerAddress(tt.dsn)
			require.Equal(t, tt.expected, result)
		})
	}
}
