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

// --- local mocks for broadcast handler ---

type mockBroadcastRepo struct{ mock.Mock }

func (m *mockBroadcastRepo) Create(ctx context.Context, broadcast *model.Broadcast) error {
	return m.Called(ctx, broadcast).Error(0)
}
func (m *mockBroadcastRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Broadcast, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Broadcast), args.Error(1)
}
func (m *mockBroadcastRepo) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Broadcast, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Broadcast), args.Error(1)
}
func (m *mockBroadcastRepo) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Broadcast, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Broadcast), args.Int(1), args.Error(2)
}
func (m *mockBroadcastRepo) Update(ctx context.Context, broadcast *model.Broadcast) error {
	return m.Called(ctx, broadcast).Error(0)
}
func (m *mockBroadcastRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockContactRepo struct{ mock.Mock }

func (m *mockContactRepo) Create(ctx context.Context, contact *model.Contact) error {
	return m.Called(ctx, contact).Error(0)
}
func (m *mockContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *mockContactRepo) GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Contact, error) {
	args := m.Called(ctx, audienceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *mockContactRepo) GetByAudienceAndEmail(ctx context.Context, audienceID uuid.UUID, email string) (*model.Contact, error) {
	args := m.Called(ctx, audienceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *mockContactRepo) List(ctx context.Context, audienceID uuid.UUID, limit, offset int) ([]model.Contact, int, error) {
	args := m.Called(ctx, audienceID, limit, offset)
	return args.Get(0).([]model.Contact), args.Int(1), args.Error(2)
}
func (m *mockContactRepo) Update(ctx context.Context, contact *model.Contact) error {
	return m.Called(ctx, contact).Error(0)
}
func (m *mockContactRepo) ListBySegmentID(ctx context.Context, segmentID uuid.UUID, limit, offset int) ([]model.Contact, int, error) {
	args := m.Called(ctx, segmentID, limit, offset)
	return args.Get(0).([]model.Contact), args.Int(1), args.Error(2)
}
func (m *mockContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockAudienceRepo struct{ mock.Mock }

func (m *mockAudienceRepo) Create(ctx context.Context, audience *model.Audience) error {
	return m.Called(ctx, audience).Error(0)
}
func (m *mockAudienceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Audience, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Audience), args.Error(1)
}
func (m *mockAudienceRepo) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Audience, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Audience), args.Error(1)
}
func (m *mockAudienceRepo) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Audience, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.Audience), args.Error(1)
}
func (m *mockAudienceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestBroadcastSendHandler_ProcessTask_NotQueued(t *testing.T) {
	broadcastRepo := new(mockBroadcastRepo)
	contactRepo := new(mockContactRepo)
	audienceRepo := new(mockAudienceRepo)
	emailRepo := new(mockEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := &BroadcastSendHandler{
		broadcastRepo:       broadcastRepo,
		contactRepo:         contactRepo,
		audienceRepo:        audienceRepo,
		emailRepo:           emailRepo,
		templateVersionRepo: nil,
		asynqClient:         nil,
		logger:              logger,
	}

	broadcastID := uuid.New()
	teamID := uuid.New()

	broadcast := &model.Broadcast{
		ID:     broadcastID,
		TeamID: teamID,
		Status: model.BroadcastStatusDraft,
	}

	broadcastRepo.On("GetByID", mock.Anything, broadcastID).Return(broadcast, nil)

	payload, _ := json.Marshal(BroadcastSendPayload{BroadcastID: broadcastID, TeamID: teamID})
	task := asynq.NewTask(TaskBroadcastSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	contactRepo.AssertNotCalled(t, "List")
}

func TestBroadcastSendHandler_ProcessTask_NoAudience(t *testing.T) {
	broadcastRepo := new(mockBroadcastRepo)
	contactRepo := new(mockContactRepo)
	audienceRepo := new(mockAudienceRepo)
	emailRepo := new(mockEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := &BroadcastSendHandler{
		broadcastRepo:       broadcastRepo,
		contactRepo:         contactRepo,
		audienceRepo:        audienceRepo,
		emailRepo:           emailRepo,
		templateVersionRepo: nil,
		asynqClient:         nil,
		logger:              logger,
	}

	broadcastID := uuid.New()
	teamID := uuid.New()

	broadcast := &model.Broadcast{
		ID:         broadcastID,
		TeamID:     teamID,
		Status:     model.BroadcastStatusQueued,
		AudienceID: nil,
	}

	broadcastRepo.On("GetByID", mock.Anything, broadcastID).Return(broadcast, nil)

	payload, _ := json.Marshal(BroadcastSendPayload{BroadcastID: broadcastID, TeamID: teamID})
	task := asynq.NewTask(TaskBroadcastSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audience")
}

func TestBroadcastSendHandler_ProcessTask_InvalidPayload(t *testing.T) {
	broadcastRepo := new(mockBroadcastRepo)
	contactRepo := new(mockContactRepo)
	audienceRepo := new(mockAudienceRepo)
	emailRepo := new(mockEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := &BroadcastSendHandler{
		broadcastRepo:       broadcastRepo,
		contactRepo:         contactRepo,
		audienceRepo:        audienceRepo,
		emailRepo:           emailRepo,
		templateVersionRepo: nil,
		asynqClient:         nil,
		logger:              logger,
	}

	task := asynq.NewTask(TaskBroadcastSend, []byte("bad json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestBroadcastSendHandler_ProcessTask_BroadcastNotFound(t *testing.T) {
	broadcastRepo := new(mockBroadcastRepo)
	contactRepo := new(mockContactRepo)
	audienceRepo := new(mockAudienceRepo)
	emailRepo := new(mockEmailRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := &BroadcastSendHandler{
		broadcastRepo:       broadcastRepo,
		contactRepo:         contactRepo,
		audienceRepo:        audienceRepo,
		emailRepo:           emailRepo,
		templateVersionRepo: nil,
		asynqClient:         nil,
		logger:              logger,
	}

	broadcastID := uuid.New()
	teamID := uuid.New()

	broadcastRepo.On("GetByID", mock.Anything, broadcastID).Return(nil, assert.AnError)

	payload, _ := json.Marshal(BroadcastSendPayload{BroadcastID: broadcastID, TeamID: teamID})
	task := asynq.NewTask(TaskBroadcastSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetching broadcast")
}
