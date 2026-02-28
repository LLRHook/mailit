package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TracingHandler wraps a slog.Handler and injects trace_id and span_id from
// the OpenTelemetry span context into every log record.
type TracingHandler struct {
	inner slog.Handler
}

// NewTracingHandler wraps the given handler with trace context injection.
func NewTracingHandler(inner slog.Handler) *TracingHandler {
	return &TracingHandler{inner: inner}
}

func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *TracingHandler) Handle(ctx context.Context, record slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, record)
}

func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TracingHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return &TracingHandler{inner: h.inner.WithGroup(name)}
}
