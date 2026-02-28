package observability

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer serves the /metrics endpoint for Prometheus scraping.
type MetricsServer struct {
	server *http.Server
}

// NewMetricsServer creates a metrics HTTP server on the given address.
func NewMetricsServer(addr string, gatherer prometheus.Gatherer) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))

	return &MetricsServer{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// ListenAndServe starts the metrics server.
func (s *MetricsServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the metrics server.
func (s *MetricsServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
