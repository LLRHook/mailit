package observability

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PgxTracer implements pgx v5's QueryTracer interface to create OpenTelemetry
// spans for database queries.
type PgxTracer struct {
	tracer trace.Tracer
}

// NewPgxTracer creates a new PgxTracer.
func NewPgxTracer() *PgxTracer {
	return &PgxTracer{
		tracer: otel.Tracer("mailit/pgx"),
	}
}

type pgxSpanKey struct{}

// TraceQueryStart creates a span when a database query begins.
func (t *PgxTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = t.tracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.statement", data.SQL),
			attribute.Int("db.args_count", len(data.Args)),
		),
	)
	return ctx
}

// TraceQueryEnd ends the span when the query completes.
func (t *PgxTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if data.Err != nil {
		span.RecordError(data.Err)
	}
	span.End()
}
