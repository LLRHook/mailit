package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metric collectors for MailIt.
type Metrics struct {
	// HTTP
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Email
	EmailsSentTotal       *prometheus.CounterVec
	EmailsQueuedTotal     *prometheus.CounterVec
	EmailSendDuration     prometheus.Histogram

	// SMTP
	SMTPConnectionsTotal *prometheus.CounterVec

	// Worker
	TasksProcessedTotal *prometheus.CounterVec
	TasksInFlight       prometheus.Gauge
	TaskDuration        *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics with the given registerer.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		// HTTP
		HTTPRequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mailit",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "path", "status"}),
		HTTPRequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "mailit",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),
		HTTPRequestsInFlight: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "mailit",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed.",
		}),

		// Email
		EmailsSentTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mailit",
			Subsystem: "email",
			Name:      "sent_total",
			Help:      "Total number of emails sent.",
		}, []string{"status"}),
		EmailsQueuedTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mailit",
			Subsystem: "email",
			Name:      "queued_total",
			Help:      "Total number of emails queued.",
		}, []string{"type"}),
		EmailSendDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: "mailit",
			Subsystem: "email",
			Name:      "send_duration_seconds",
			Help:      "Time to deliver an email via SMTP.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
		}),

		// SMTP
		SMTPConnectionsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mailit",
			Subsystem: "smtp",
			Name:      "connections_total",
			Help:      "Total SMTP connections attempted.",
		}, []string{"mx_host", "result"}),

		// Worker
		TasksProcessedTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "mailit",
			Subsystem: "worker",
			Name:      "tasks_processed_total",
			Help:      "Total number of tasks processed.",
		}, []string{"task_type", "result"}),
		TasksInFlight: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: "mailit",
			Subsystem: "worker",
			Name:      "tasks_in_flight",
			Help:      "Number of tasks currently being processed.",
		}),
		TaskDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "mailit",
			Subsystem: "worker",
			Name:      "task_duration_seconds",
			Help:      "Task processing duration in seconds.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60},
		}, []string{"task_type"}),
	}
}
