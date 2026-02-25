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

// --- local mock for DomainDNSRecordRepository ---

type mockDNSRecordRepo struct{ mock.Mock }

func (m *mockDNSRecordRepo) Create(ctx context.Context, record *model.DomainDNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *mockDNSRecordRepo) ListByDomainID(ctx context.Context, domainID uuid.UUID) ([]model.DomainDNSRecord, error) {
	args := m.Called(ctx, domainID)
	return args.Get(0).([]model.DomainDNSRecord), args.Error(1)
}
func (m *mockDNSRecordRepo) Update(ctx context.Context, record *model.DomainDNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *mockDNSRecordRepo) DeleteByDomainID(ctx context.Context, domainID uuid.UUID) error {
	return m.Called(ctx, domainID).Error(0)
}

func TestDomainVerifyHandler_ProcessTask_NoRecords(t *testing.T) {
	domainRepo := new(mockDomainRepo)
	dnsRecordRepo := new(mockDNSRecordRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewDomainVerifyHandler(domainRepo, dnsRecordRepo, logger)

	domainID := uuid.New()
	teamID := uuid.New()
	domain := &model.Domain{
		ID:        domainID,
		TeamID:    teamID,
		Name:      "example.com",
		Status:    model.DomainStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	domainRepo.On("GetByID", mock.Anything, domainID).Return(domain, nil)
	dnsRecordRepo.On("ListByDomainID", mock.Anything, domainID).Return([]model.DomainDNSRecord{}, nil)

	payload, _ := json.Marshal(DomainVerifyPayload{DomainID: domainID, TeamID: teamID})
	task := asynq.NewTask(TaskDomainVerify, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	domainRepo.AssertExpectations(t)
}

func TestDomainVerifyHandler_ProcessTask_InvalidPayload(t *testing.T) {
	domainRepo := new(mockDomainRepo)
	dnsRecordRepo := new(mockDNSRecordRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewDomainVerifyHandler(domainRepo, dnsRecordRepo, logger)

	task := asynq.NewTask(TaskDomainVerify, []byte("invalid json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestDomainVerifyHandler_ProcessTask_DomainNotFound(t *testing.T) {
	domainRepo := new(mockDomainRepo)
	dnsRecordRepo := new(mockDNSRecordRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	h := NewDomainVerifyHandler(domainRepo, dnsRecordRepo, logger)

	domainID := uuid.New()
	teamID := uuid.New()

	domainRepo.On("GetByID", mock.Anything, domainID).Return(nil, assert.AnError)

	payload, _ := json.Marshal(DomainVerifyPayload{DomainID: domainID, TeamID: teamID})
	task := asynq.NewTask(TaskDomainVerify, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetching domain")
}

func TestIsCriticalRecord(t *testing.T) {
	assert.True(t, isCriticalRecord(RecordTypeSPF))
	assert.True(t, isCriticalRecord(RecordTypeDKIM))
	assert.True(t, isCriticalRecord(RecordTypeMX))
	assert.False(t, isCriticalRecord(RecordTypeDMARC))
	assert.False(t, isCriticalRecord(RecordTypeReturnPath))
}
