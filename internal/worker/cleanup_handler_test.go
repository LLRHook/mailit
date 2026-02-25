package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/model"
)

// --- local mocks for cleanup handler ---

type mockWebhookEventRepo struct{ mock.Mock }

func (m *mockWebhookEventRepo) Create(ctx context.Context, event *model.WebhookEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *mockWebhookEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebhookEvent), args.Error(1)
}
func (m *mockWebhookEventRepo) Update(ctx context.Context, event *model.WebhookEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *mockWebhookEventRepo) ListByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]model.WebhookEvent, int, error) {
	args := m.Called(ctx, webhookID, limit, offset)
	return args.Get(0).([]model.WebhookEvent), args.Int(1), args.Error(2)
}
func (m *mockWebhookEventRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

type mockLogRepo struct{ mock.Mock }

func (m *mockLogRepo) Create(ctx context.Context, log *model.Log) error {
	return m.Called(ctx, log).Error(0)
}
func (m *mockLogRepo) List(ctx context.Context, teamID uuid.UUID, level string, limit, offset int) ([]model.Log, int, error) {
	args := m.Called(ctx, teamID, level, limit, offset)
	return args.Get(0).([]model.Log), args.Int(1), args.Error(2)
}

func TestCleanupHandler_ProcessTask_Success(t *testing.T) {
	webhookEventRepo := new(mockWebhookEventRepo)
	logRepo := new(mockLogRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewCleanupHandler(webhookEventRepo, logRepo, logger)

	webhookEventRepo.On("DeleteOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(int64(5), nil)

	task := asynq.NewTask(TaskCleanupExpired, nil)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	webhookEventRepo.AssertExpectations(t)
}

func TestCleanupHandler_ProcessTask_WebhookCleanupError(t *testing.T) {
	webhookEventRepo := new(mockWebhookEventRepo)
	logRepo := new(mockLogRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewCleanupHandler(webhookEventRepo, logRepo, logger)

	webhookEventRepo.On("DeleteOlderThan", mock.Anything, mock.AnythingOfType("time.Time")).Return(int64(0), errors.New("db error"))

	task := asynq.NewTask(TaskCleanupExpired, nil)

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleanup completed with 1 errors")
	webhookEventRepo.AssertExpectations(t)
}
