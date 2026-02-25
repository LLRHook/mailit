package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestWebhookDeliverHandler_ProcessTask_InvalidPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// We can't easily create a real Dispatcher without setting up repos,
	// but we can test the payload parsing path.
	h := NewWebhookDeliverHandler(nil, logger)

	task := asynq.NewTask(TaskWebhookDeliver, []byte("bad json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestWebhookDeliverHandler_ProcessTask_NilDispatcher_Panics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewWebhookDeliverHandler(nil, logger)

	webhookEventID := uuid.New()
	payload, _ := json.Marshal(WebhookDeliverPayload{WebhookEventID: webhookEventID})
	task := asynq.NewTask(TaskWebhookDeliver, payload)

	// A nil dispatcher causes a nil pointer dereference when Deliver is called.
	assert.Panics(t, func() {
		_ = h.ProcessTask(context.Background(), task)
	})
}
