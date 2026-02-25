package worker

import "github.com/hibiken/asynq"

// TaskEnqueuer abstracts the asynq.Client so it can be mocked in tests.
type TaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
