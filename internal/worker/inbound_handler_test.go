package worker

import (
	"context"
	"encoding/json"
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

// --- local mock for InboundEmailRepository ---

type mockInboundEmailRepo struct{ mock.Mock }

func (m *mockInboundEmailRepo) Create(ctx context.Context, email *model.InboundEmail) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockInboundEmailRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.InboundEmail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundEmail), args.Error(1)
}
func (m *mockInboundEmailRepo) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.InboundEmail, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundEmail), args.Error(1)
}
func (m *mockInboundEmailRepo) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.InboundEmail, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.InboundEmail), args.Int(1), args.Error(2)
}
func (m *mockInboundEmailRepo) Update(ctx context.Context, email *model.InboundEmail) error {
	return m.Called(ctx, email).Error(0)
}

func TestInboundHandler_ProcessTask_Success(t *testing.T) {
	inboundRepo := new(mockInboundEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	webhookCalled := false
	var capturedEventType string
	webhookDispatch := func(ctx context.Context, teamID uuid.UUID, eventType string, payload interface{}) {
		webhookCalled = true
		capturedEventType = eventType
	}

	h := NewInboundHandler(inboundRepo, webhookDispatch, logger)

	inboundEmailID := uuid.New()
	teamID := uuid.New()
	subject := "Hello"
	inbound := &model.InboundEmail{
		ID:          inboundEmailID,
		TeamID:      teamID,
		FromAddress: "sender@external.com",
		ToAddresses: []string{"inbox@example.com"},
		Subject:     &subject,
		Processed:   false,
		CreatedAt:   time.Now(),
	}

	inboundRepo.On("GetByID", mock.Anything, inboundEmailID).Return(inbound, nil)
	inboundRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *model.InboundEmail) bool {
		return e.Processed == true
	})).Return(nil)

	payload, _ := json.Marshal(InboundProcessPayload{InboundEmailID: inboundEmailID})
	task := asynq.NewTask(TaskInboundProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	assert.True(t, webhookCalled)
	assert.Equal(t, "email.received", capturedEventType)
	inboundRepo.AssertExpectations(t)
}

func TestInboundHandler_ProcessTask_AlreadyProcessed(t *testing.T) {
	inboundRepo := new(mockInboundEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewInboundHandler(inboundRepo, nil, logger)

	inboundEmailID := uuid.New()
	inbound := &model.InboundEmail{
		ID:        inboundEmailID,
		Processed: true,
	}

	inboundRepo.On("GetByID", mock.Anything, inboundEmailID).Return(inbound, nil)

	payload, _ := json.Marshal(InboundProcessPayload{InboundEmailID: inboundEmailID})
	task := asynq.NewTask(TaskInboundProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	inboundRepo.AssertNotCalled(t, "Update")
}

func TestInboundHandler_ProcessTask_InvalidPayload(t *testing.T) {
	inboundRepo := new(mockInboundEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewInboundHandler(inboundRepo, nil, logger)

	task := asynq.NewTask(TaskInboundProcess, []byte("bad json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestInboundHandler_ProcessTask_NotFound(t *testing.T) {
	inboundRepo := new(mockInboundEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewInboundHandler(inboundRepo, nil, logger)

	inboundEmailID := uuid.New()

	inboundRepo.On("GetByID", mock.Anything, inboundEmailID).Return(nil, assert.AnError)

	payload, _ := json.Marshal(InboundProcessPayload{InboundEmailID: inboundEmailID})
	task := asynq.NewTask(TaskInboundProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetching inbound email")
}

func TestInboundHandler_ProcessTask_NoWebhookDispatch(t *testing.T) {
	inboundRepo := new(mockInboundEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewInboundHandler(inboundRepo, nil, logger)

	inboundEmailID := uuid.New()
	teamID := uuid.New()
	inbound := &model.InboundEmail{
		ID:          inboundEmailID,
		TeamID:      teamID,
		FromAddress: "sender@external.com",
		ToAddresses: []string{"inbox@example.com"},
		Processed:   false,
		CreatedAt:   time.Now(),
	}

	inboundRepo.On("GetByID", mock.Anything, inboundEmailID).Return(inbound, nil)
	inboundRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.InboundEmail")).Return(nil)

	payload, _ := json.Marshal(InboundProcessPayload{InboundEmailID: inboundEmailID})
	task := asynq.NewTask(TaskInboundProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	inboundRepo.AssertExpectations(t)
}

func TestBuildInboundWebhookPayload(t *testing.T) {
	id := uuid.New()
	teamID := uuid.New()
	subject := "Test Subject"

	payload := buildInboundWebhookPayload(id, teamID, "from@example.com", []string{"to@example.com"}, &subject)

	assert.Equal(t, id.String(), payload["inbound_email_id"])
	assert.Equal(t, teamID.String(), payload["team_id"])
	assert.Equal(t, "from@example.com", payload["from"])
	assert.Equal(t, "Test Subject", payload["subject"])
	assert.NotEmpty(t, payload["timestamp"])
}

func TestBuildInboundWebhookPayload_NilSubject(t *testing.T) {
	id := uuid.New()
	teamID := uuid.New()

	payload := buildInboundWebhookPayload(id, teamID, "from@example.com", []string{"to@example.com"}, nil)

	_, hasSubject := payload["subject"]
	assert.False(t, hasSubject)
}
