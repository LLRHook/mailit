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
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/model"
)

func TestBounceHandler_ProcessTask_HardBounce(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	suppressionRepo := new(mockSuppressionRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewBounceHandler(emailRepo, eventRepo, suppressionRepo, logger)

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusSent,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "bounce@example.com").Return(nil, nil)
	suppressionRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.SuppressionEntry")).Return(nil)
	emailRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *model.Email) bool {
		return e.Status == model.EmailStatusBounced
	})).Return(nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	bouncePayload := BounceProcessPayload{
		EmailID:   emailID,
		Code:      550,
		Message:   "User not found",
		Recipient: "bounce@example.com",
	}
	payload, _ := json.Marshal(bouncePayload)
	task := asynq.NewTask(TaskBounceProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	suppressionRepo.AssertExpectations(t)
	emailRepo.AssertExpectations(t)
}

func TestBounceHandler_ProcessTask_SoftBounce(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	suppressionRepo := new(mockSuppressionRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewBounceHandler(emailRepo, eventRepo, suppressionRepo, logger)

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusSent,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	bouncePayload := BounceProcessPayload{
		EmailID:   emailID,
		Code:      421,
		Message:   "Try again later",
		Recipient: "temp@example.com",
	}
	payload, _ := json.Marshal(bouncePayload)
	task := asynq.NewTask(TaskBounceProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	suppressionRepo.AssertNotCalled(t, "Create")
}

func TestBounceHandler_ProcessTask_Complaint(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	suppressionRepo := new(mockSuppressionRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewBounceHandler(emailRepo, eventRepo, suppressionRepo, logger)

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusSent,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "complainer@example.com").Return(nil, nil)
	suppressionRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.SuppressionEntry")).Return(nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	bouncePayload := BounceProcessPayload{
		EmailID:   emailID,
		Code:      550,
		Message:   "Message flagged as spam by recipient",
		Recipient: "complainer@example.com",
	}
	payload, _ := json.Marshal(bouncePayload)
	task := asynq.NewTask(TaskBounceProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	suppressionRepo.AssertExpectations(t)
}

func TestBounceHandler_ProcessTask_InvalidPayload(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	suppressionRepo := new(mockSuppressionRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewBounceHandler(emailRepo, eventRepo, suppressionRepo, logger)

	task := asynq.NewTask(TaskBounceProcess, []byte("invalid json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestBounceHandler_ProcessTask_AlreadySuppressed(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	suppressionRepo := new(mockSuppressionRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewBounceHandler(emailRepo, eventRepo, suppressionRepo, logger)

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusSent,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "already@example.com").Return(
		&model.SuppressionEntry{Email: "already@example.com", Reason: "hard_bounce"}, nil,
	)
	emailRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Email")).Return(nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	bouncePayload := BounceProcessPayload{
		EmailID:   emailID,
		Code:      550,
		Message:   "User not found",
		Recipient: "already@example.com",
	}
	payload, _ := json.Marshal(bouncePayload)
	task := asynq.NewTask(TaskBounceProcess, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	// Create should NOT be called because already suppressed
	suppressionRepo.AssertNotCalled(t, "Create")
}

func TestClassifyBounce(t *testing.T) {
	tests := []struct {
		code     int
		message  string
		expected string
	}{
		{550, "User not found", BounceTypeHard},
		{421, "Try again later", BounceTypeSoft},
		{550, "Message flagged as spam", BounceTypeComplaint},
		{550, "Blocked by abuse filter", BounceTypeComplaint},
		{452, "Mailbox full", BounceTypeSoft},
		{300, "Unknown code", BounceTypeSoft},
	}

	for _, tt := range tests {
		result := classifyBounce(tt.code, tt.message)
		assert.Equal(t, tt.expected, result, "code=%d message=%s", tt.code, tt.message)
	}
}
