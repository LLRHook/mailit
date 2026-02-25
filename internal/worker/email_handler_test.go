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

// --- local mocks to avoid import cycle with testutil/mock ---

type mockEmailRepo struct{ mock.Mock }

func (m *mockEmailRepo) Create(ctx context.Context, email *model.Email) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockEmailRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Email, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Email), args.Error(1)
}
func (m *mockEmailRepo) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Email, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Email), args.Error(1)
}
func (m *mockEmailRepo) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Email, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Email), args.Int(1), args.Error(2)
}
func (m *mockEmailRepo) Update(ctx context.Context, email *model.Email) error {
	return m.Called(ctx, email).Error(0)
}

type mockEmailEventRepo struct{ mock.Mock }

func (m *mockEmailEventRepo) Create(ctx context.Context, event *model.EmailEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *mockEmailEventRepo) ListByEmailID(ctx context.Context, emailID uuid.UUID) ([]model.EmailEvent, error) {
	args := m.Called(ctx, emailID)
	return args.Get(0).([]model.EmailEvent), args.Error(1)
}

type mockDomainRepo struct{ mock.Mock }

func (m *mockDomainRepo) Create(ctx context.Context, domain *model.Domain) error {
	return m.Called(ctx, domain).Error(0)
}
func (m *mockDomainRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Domain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *mockDomainRepo) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Domain, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *mockDomainRepo) GetByTeamAndName(ctx context.Context, teamID uuid.UUID, name string) (*model.Domain, error) {
	args := m.Called(ctx, teamID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *mockDomainRepo) GetVerifiedByName(ctx context.Context, name string) (*model.Domain, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *mockDomainRepo) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Domain, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Domain), args.Int(1), args.Error(2)
}
func (m *mockDomainRepo) Update(ctx context.Context, domain *model.Domain) error {
	return m.Called(ctx, domain).Error(0)
}
func (m *mockDomainRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockSuppressionRepo struct{ mock.Mock }

func (m *mockSuppressionRepo) Create(ctx context.Context, entry *model.SuppressionEntry) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *mockSuppressionRepo) GetByTeamAndEmail(ctx context.Context, teamID uuid.UUID, email string) (*model.SuppressionEntry, error) {
	args := m.Called(ctx, teamID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SuppressionEntry), args.Error(1)
}
func (m *mockSuppressionRepo) ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.SuppressionEntry, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.SuppressionEntry), args.Int(1), args.Error(2)
}
func (m *mockSuppressionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockSender struct{ mock.Mock }

func (m *mockSender) SendEmail(ctx context.Context, msg *OutboundMessage) ([]RecipientResult, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]RecipientResult), args.Error(1)
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestEmailSendHandler_ProcessTask_Success(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	webhookCalled := false
	webhookDispatch := func(ctx context.Context, teamID uuid.UUID, eventType string, payload interface{}) {
		webhookCalled = true
	}

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, webhookDispatch, nil, newDiscardLogger())

	emailID := uuid.New()
	teamID := uuid.New()
	html := "<p>Hello</p>"
	email := &model.Email{
		ID:          emailID,
		TeamID:      teamID,
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Test",
		HTMLBody:    &html,
		Status:      model.EmailStatusQueued,
		Headers:     model.JSONMap{},
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "recipient@example.com").Return(nil, nil)
	domainRepo.On("GetByTeamAndName", mock.Anything, teamID, "example.com").Return(nil, assert.AnError)
	emailRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Email")).Return(nil)

	results := []RecipientResult{
		{Recipient: "recipient@example.com", Success: true, Code: 250, Message: "OK"},
	}
	sender.On("SendEmail", mock.Anything, mock.AnythingOfType("*worker.OutboundMessage")).Return(results, nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	payload, _ := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	task := asynq.NewTask(TaskEmailSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	assert.True(t, webhookCalled)
	emailRepo.AssertExpectations(t)
	sender.AssertExpectations(t)
}

func TestEmailSendHandler_ProcessTask_CancelledEmail(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusCancelled,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)

	payload, _ := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	task := asynq.NewTask(TaskEmailSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	sender.AssertNotCalled(t, "SendEmail")
}

func TestEmailSendHandler_ProcessTask_AllSuppressed(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:          emailID,
		TeamID:      teamID,
		FromAddress: "sender@example.com",
		ToAddresses: []string{"suppressed@example.com"},
		Status:      model.EmailStatusQueued,
		Headers:     model.JSONMap{},
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "suppressed@example.com").Return(
		&model.SuppressionEntry{Email: "suppressed@example.com", Reason: "hard_bounce"},
		nil,
	)
	emailRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *model.Email) bool {
		return e.Status == model.EmailStatusFailed
	})).Return(nil)

	payload, _ := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	task := asynq.NewTask(TaskEmailSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	sender.AssertNotCalled(t, "SendEmail")
	emailRepo.AssertExpectations(t)
}

func TestEmailSendHandler_ProcessTask_AlreadySent(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	emailID := uuid.New()
	teamID := uuid.New()
	email := &model.Email{
		ID:     emailID,
		TeamID: teamID,
		Status: model.EmailStatusSent,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)

	payload, _ := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	task := asynq.NewTask(TaskEmailSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	sender.AssertNotCalled(t, "SendEmail")
}

func TestEmailSendHandler_ProcessTask_WithDKIM(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	emailID := uuid.New()
	teamID := uuid.New()
	domainID := uuid.New()
	html := "<p>Hello</p>"
	privKey := "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"

	email := &model.Email{
		ID:          emailID,
		TeamID:      teamID,
		DomainID:    &domainID,
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Test",
		HTMLBody:    &html,
		Status:      model.EmailStatusQueued,
		Headers:     model.JSONMap{},
	}

	domain := &model.Domain{
		ID:             domainID,
		Name:           "example.com",
		DKIMSelector:   "mailit",
		DKIMPrivateKey: &privKey,
	}

	emailRepo.On("GetByID", mock.Anything, emailID).Return(email, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "recipient@example.com").Return(nil, nil)
	domainRepo.On("GetByID", mock.Anything, domainID).Return(domain, nil)
	emailRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Email")).Return(nil)

	results := []RecipientResult{
		{Recipient: "recipient@example.com", Success: true, Code: 250, Message: "OK"},
	}
	sender.On("SendEmail", mock.Anything, mock.MatchedBy(func(msg *OutboundMessage) bool {
		return msg.DKIMDomain == "example.com" && msg.DKIMSelector == "mailit"
	})).Return(results, nil)
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.EmailEvent")).Return(nil)

	payload, _ := json.Marshal(EmailSendPayload{EmailID: emailID, TeamID: teamID})
	task := asynq.NewTask(TaskEmailSend, payload)

	err := h.ProcessTask(context.Background(), task)
	assert.NoError(t, err)
	sender.AssertExpectations(t)
}

func TestEmailSendHandler_ProcessTask_InvalidPayload(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	task := asynq.NewTask(TaskEmailSend, []byte("invalid json"))

	err := h.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshalling")
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "example.com"},
		{"user@subdomain.example.com", "subdomain.example.com"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		result := extractDomain(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestPtrToString(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", ptrToString(&s))
	assert.Equal(t, "", ptrToString(nil))
}

func TestJsonMapToStringMap(t *testing.T) {
	m := model.JSONMap{"key": "value", "num": 42}
	result := jsonMapToStringMap(m)
	assert.Equal(t, "value", result["key"])
	assert.Empty(t, result["num"])

	assert.Nil(t, jsonMapToStringMap(nil))
}

func TestFilterSuppressed(t *testing.T) {
	emailRepo := new(mockEmailRepo)
	eventRepo := new(mockEmailEventRepo)
	domainRepo := new(mockDomainRepo)
	suppressionRepo := new(mockSuppressionRepo)
	sender := new(mockSender)

	h := NewEmailSendHandler(emailRepo, eventRepo, domainRepo, suppressionRepo, sender, nil, nil, newDiscardLogger())

	teamID := uuid.New()
	ctx := context.Background()
	log := newDiscardLogger()

	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "good@example.com").Return(nil, nil)
	suppressionRepo.On("GetByTeamAndEmail", mock.Anything, teamID, "bad@example.com").Return(
		&model.SuppressionEntry{
			Email:     "bad@example.com",
			Reason:    "hard_bounce",
			CreatedAt: time.Now(),
		}, nil,
	)

	result := h.filterSuppressed(ctx, teamID, []string{"good@example.com", "bad@example.com"}, log)
	assert.Equal(t, []string{"good@example.com"}, result)
}
