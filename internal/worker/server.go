package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Config holds configuration for the asynq worker server.
type Config struct {
	RedisAddr     string
	RedisPassword string
	Concurrency   int
	Queues        map[string]int // queue name -> priority weight
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		Concurrency:   20,
		Queues: map[string]int{
			QueueCritical: 6,
			QueueDefault:  3,
			QueueLow:      1,
		},
	}
}

// Handlers holds all task handler instances that will be registered with the mux.
type Handlers struct {
	EmailSend      *EmailSendHandler
	EmailBatchSend *BatchEmailSendHandler
	BroadcastSend  *BroadcastSendHandler
	DomainVerify   *DomainVerifyHandler
	Bounce         *BounceHandler
	Inbound        *InboundHandler
	Cleanup        *CleanupHandler
	WebhookDeliver   *WebhookDeliverHandler
	MetricsAggregate *MetricsAggregateHandler
}

// NewServer creates and configures a new asynq Server.
func NewServer(cfg Config, logger *slog.Logger) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	}

	queues := cfg.Queues
	if queues == nil {
		queues = DefaultConfig().Queues
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = DefaultConfig().Concurrency
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: concurrency,
		Queues:      queues,
		Logger:      newAsynqLogger(logger),
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error("task processing failed",
				"task_type", task.Type(),
				"error", err,
			)
		}),
	})

	return srv
}

// NewMux creates an asynq ServeMux with all task handlers registered.
func NewMux(h Handlers) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	if h.EmailSend != nil {
		mux.HandleFunc(TaskEmailSend, h.EmailSend.ProcessTask)
	}
	if h.EmailBatchSend != nil {
		mux.HandleFunc(TaskEmailBatchSend, h.EmailBatchSend.ProcessTask)
	}
	if h.BroadcastSend != nil {
		mux.HandleFunc(TaskBroadcastSend, h.BroadcastSend.ProcessTask)
	}
	if h.DomainVerify != nil {
		mux.HandleFunc(TaskDomainVerify, h.DomainVerify.ProcessTask)
	}
	if h.Bounce != nil {
		mux.HandleFunc(TaskBounceProcess, h.Bounce.ProcessTask)
	}
	if h.Inbound != nil {
		mux.HandleFunc(TaskInboundProcess, h.Inbound.ProcessTask)
	}
	if h.Cleanup != nil {
		mux.HandleFunc(TaskCleanupExpired, h.Cleanup.ProcessTask)
	}
	if h.WebhookDeliver != nil {
		mux.HandleFunc(TaskWebhookDeliver, h.WebhookDeliver.ProcessTask)
	}
	if h.MetricsAggregate != nil {
		mux.HandleFunc(TaskMetricsAggregate, h.MetricsAggregate.ProcessTask)
	}

	return mux
}

// asynqLogger adapts slog.Logger to asynq's Logger interface.
type asynqLogger struct {
	logger *slog.Logger
}

func newAsynqLogger(logger *slog.Logger) *asynqLogger {
	return &asynqLogger{logger: logger}
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug("asynq", "msg", args)
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info("asynq", "msg", args)
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn("asynq", "msg", args)
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error("asynq", "msg", args)
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Error("asynq fatal", "msg", args)
}
