package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mailit-dev/mailit/internal/observability"
)

// MetricsMiddleware records HTTP request count, duration, and in-flight gauge.
func MetricsMiddleware(m *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.HTTPRequestsInFlight.Inc()
			defer m.HTTPRequestsInFlight.Dec()

			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start).Seconds()

			// Use chi route pattern to avoid label cardinality explosion.
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unmatched"
			}

			m.HTTPRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(sw.status)).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
		})
	}
}

// statusWriter captures the HTTP response status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}
