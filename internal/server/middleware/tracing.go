package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates spans per HTTP request with route and status attributes.
func TracingMiddleware() func(http.Handler) http.Handler {
	tracer := otel.Tracer("mailit/http")
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract incoming trace context.
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}

			spanName := fmt.Sprintf("%s %s", r.Method, route)
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.route", route),
					attribute.String("http.target", r.URL.Path),
				),
			)
			defer span.End()

			// Inject trace context into response headers.
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			span.SetAttributes(attribute.Int("http.status_code", sw.status))
		})
	}
}
