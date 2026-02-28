package observability

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// AsynqMetricsMiddleware wraps asynq task processing with Prometheus counters,
// gauges, and histograms.
func AsynqMetricsMiddleware(m *Metrics) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			m.TasksInFlight.Inc()
			defer m.TasksInFlight.Dec()

			start := time.Now()
			err := next.ProcessTask(ctx, task)
			duration := time.Since(start).Seconds()

			result := "success"
			if err != nil {
				result = "error"
			}

			m.TasksProcessedTotal.WithLabelValues(task.Type(), result).Inc()
			m.TaskDuration.WithLabelValues(task.Type()).Observe(duration)

			return err
		})
	}
}
