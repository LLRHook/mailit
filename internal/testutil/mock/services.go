package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
)

// --- AuthService ---

type MockAuthService struct{ mock.Mock }

func (m *MockAuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}
func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

// --- EmailService ---

type MockEmailService struct{ mock.Mock }

func (m *MockEmailService) Send(ctx context.Context, teamID uuid.UUID, req *dto.SendEmailRequest) (*dto.SendEmailResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SendEmailResponse), args.Error(1)
}
func (m *MockEmailService) BatchSend(ctx context.Context, teamID uuid.UUID, req *dto.BatchSendEmailRequest) (*dto.BatchSendEmailResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BatchSendEmailResponse), args.Error(1)
}
func (m *MockEmailService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.EmailResponse], error) {
	args := m.Called(ctx, teamID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[dto.EmailResponse]), args.Error(1)
}
func (m *MockEmailService) Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error) {
	args := m.Called(ctx, teamID, emailID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EmailResponse), args.Error(1)
}
func (m *MockEmailService) Update(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID, req map[string]interface{}) (*dto.EmailResponse, error) {
	args := m.Called(ctx, teamID, emailID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EmailResponse), args.Error(1)
}
func (m *MockEmailService) Cancel(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*dto.EmailResponse, error) {
	args := m.Called(ctx, teamID, emailID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EmailResponse), args.Error(1)
}

// --- DomainService ---

type MockDomainService struct{ mock.Mock }

func (m *MockDomainService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateDomainRequest) (*dto.DomainResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DomainResponse), args.Error(1)
}
func (m *MockDomainService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.DomainResponse], error) {
	args := m.Called(ctx, teamID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[dto.DomainResponse]), args.Error(1)
}
func (m *MockDomainService) Get(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error) {
	args := m.Called(ctx, teamID, domainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DomainResponse), args.Error(1)
}
func (m *MockDomainService) Update(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID, req *dto.UpdateDomainRequest) (*dto.DomainResponse, error) {
	args := m.Called(ctx, teamID, domainID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DomainResponse), args.Error(1)
}
func (m *MockDomainService) Delete(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) error {
	return m.Called(ctx, teamID, domainID).Error(0)
}
func (m *MockDomainService) Verify(ctx context.Context, teamID uuid.UUID, domainID uuid.UUID) (*dto.DomainResponse, error) {
	args := m.Called(ctx, teamID, domainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DomainResponse), args.Error(1)
}

// --- APIKeyService ---

type MockAPIKeyService struct{ mock.Mock }

func (m *MockAPIKeyService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAPIKeyRequest) (*dto.APIKeyResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.APIKeyResponse), args.Error(1)
}
func (m *MockAPIKeyService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.APIKeyResponse], error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.APIKeyResponse]), args.Error(1)
}
func (m *MockAPIKeyService) Delete(ctx context.Context, teamID uuid.UUID, apiKeyID uuid.UUID) error {
	return m.Called(ctx, teamID, apiKeyID).Error(0)
}

// --- AudienceService ---

type MockAudienceService struct{ mock.Mock }

func (m *MockAudienceService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateAudienceRequest) (*dto.AudienceResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AudienceResponse), args.Error(1)
}
func (m *MockAudienceService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.AudienceResponse], error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.AudienceResponse]), args.Error(1)
}
func (m *MockAudienceService) Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.AudienceResponse, error) {
	args := m.Called(ctx, teamID, audienceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AudienceResponse), args.Error(1)
}
func (m *MockAudienceService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) error {
	return m.Called(ctx, teamID, audienceID).Error(0)
}

// --- ContactService ---

type MockContactService struct{ mock.Mock }

func (m *MockContactService) Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateContactRequest) (*dto.ContactResponse, error) {
	args := m.Called(ctx, teamID, audienceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ContactResponse), args.Error(1)
}
func (m *MockContactService) List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.ContactResponse], error) {
	args := m.Called(ctx, teamID, audienceID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[dto.ContactResponse]), args.Error(1)
}
func (m *MockContactService) Get(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) (*dto.ContactResponse, error) {
	args := m.Called(ctx, teamID, audienceID, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ContactResponse), args.Error(1)
}
func (m *MockContactService) Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID, req *dto.UpdateContactRequest) (*dto.ContactResponse, error) {
	args := m.Called(ctx, teamID, audienceID, contactID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ContactResponse), args.Error(1)
}
func (m *MockContactService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, contactID uuid.UUID) error {
	return m.Called(ctx, teamID, audienceID, contactID).Error(0)
}

// --- ContactPropertyService ---

type MockContactPropertyService struct{ mock.Mock }

func (m *MockContactPropertyService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateContactPropertyRequest) (*dto.ContactPropertyResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ContactPropertyResponse), args.Error(1)
}
func (m *MockContactPropertyService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.ContactPropertyResponse], error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.ContactPropertyResponse]), args.Error(1)
}
func (m *MockContactPropertyService) Update(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID, req *dto.UpdateContactPropertyRequest) (*dto.ContactPropertyResponse, error) {
	args := m.Called(ctx, teamID, propertyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ContactPropertyResponse), args.Error(1)
}
func (m *MockContactPropertyService) Delete(ctx context.Context, teamID uuid.UUID, propertyID uuid.UUID) error {
	return m.Called(ctx, teamID, propertyID).Error(0)
}

// --- TopicService ---

type MockTopicService struct{ mock.Mock }

func (m *MockTopicService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTopicRequest) (*dto.TopicResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TopicResponse), args.Error(1)
}
func (m *MockTopicService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.TopicResponse], error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.TopicResponse]), args.Error(1)
}
func (m *MockTopicService) Update(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID, req *dto.UpdateTopicRequest) (*dto.TopicResponse, error) {
	args := m.Called(ctx, teamID, topicID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TopicResponse), args.Error(1)
}
func (m *MockTopicService) Delete(ctx context.Context, teamID uuid.UUID, topicID uuid.UUID) error {
	return m.Called(ctx, teamID, topicID).Error(0)
}

// --- SegmentService ---

type MockSegmentService struct{ mock.Mock }

func (m *MockSegmentService) Create(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, req *dto.CreateSegmentRequest) (*dto.SegmentResponse, error) {
	args := m.Called(ctx, teamID, audienceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SegmentResponse), args.Error(1)
}
func (m *MockSegmentService) List(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID) (*dto.ListResponse[dto.SegmentResponse], error) {
	args := m.Called(ctx, teamID, audienceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.SegmentResponse]), args.Error(1)
}
func (m *MockSegmentService) Update(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID, req *dto.UpdateSegmentRequest) (*dto.SegmentResponse, error) {
	args := m.Called(ctx, teamID, audienceID, segmentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SegmentResponse), args.Error(1)
}
func (m *MockSegmentService) Delete(ctx context.Context, teamID uuid.UUID, audienceID uuid.UUID, segmentID uuid.UUID) error {
	return m.Called(ctx, teamID, audienceID, segmentID).Error(0)
}

// --- TemplateService ---

type MockTemplateService struct{ mock.Mock }

func (m *MockTemplateService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateTemplateRequest) (*dto.TemplateResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TemplateResponse), args.Error(1)
}
func (m *MockTemplateService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.TemplateResponse], error) {
	args := m.Called(ctx, teamID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[dto.TemplateResponse]), args.Error(1)
}
func (m *MockTemplateService) Get(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateDetailResponse, error) {
	args := m.Called(ctx, teamID, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TemplateDetailResponse), args.Error(1)
}
func (m *MockTemplateService) Update(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID, req *dto.UpdateTemplateRequest) (*dto.TemplateResponse, error) {
	args := m.Called(ctx, teamID, templateID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TemplateResponse), args.Error(1)
}
func (m *MockTemplateService) Delete(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) error {
	return m.Called(ctx, teamID, templateID).Error(0)
}
func (m *MockTemplateService) Publish(ctx context.Context, teamID uuid.UUID, templateID uuid.UUID) (*dto.TemplateResponse, error) {
	args := m.Called(ctx, teamID, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TemplateResponse), args.Error(1)
}

// --- BroadcastService ---

type MockBroadcastService struct{ mock.Mock }

func (m *MockBroadcastService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateBroadcastRequest) (*dto.BroadcastResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BroadcastResponse), args.Error(1)
}
func (m *MockBroadcastService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[dto.BroadcastResponse], error) {
	args := m.Called(ctx, teamID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[dto.BroadcastResponse]), args.Error(1)
}
func (m *MockBroadcastService) Get(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error) {
	args := m.Called(ctx, teamID, broadcastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BroadcastResponse), args.Error(1)
}
func (m *MockBroadcastService) Update(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID, req *dto.UpdateBroadcastRequest) (*dto.BroadcastResponse, error) {
	args := m.Called(ctx, teamID, broadcastID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BroadcastResponse), args.Error(1)
}
func (m *MockBroadcastService) Delete(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) error {
	return m.Called(ctx, teamID, broadcastID).Error(0)
}
func (m *MockBroadcastService) Send(ctx context.Context, teamID uuid.UUID, broadcastID uuid.UUID) (*dto.BroadcastResponse, error) {
	args := m.Called(ctx, teamID, broadcastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BroadcastResponse), args.Error(1)
}

// --- WebhookService ---

type MockWebhookService struct{ mock.Mock }

func (m *MockWebhookService) Create(ctx context.Context, teamID uuid.UUID, req *dto.CreateWebhookRequest) (*dto.WebhookResponse, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WebhookResponse), args.Error(1)
}
func (m *MockWebhookService) List(ctx context.Context, teamID uuid.UUID) (*dto.ListResponse[dto.WebhookResponse], error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ListResponse[dto.WebhookResponse]), args.Error(1)
}
func (m *MockWebhookService) Get(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) (*dto.WebhookResponse, error) {
	args := m.Called(ctx, teamID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WebhookResponse), args.Error(1)
}
func (m *MockWebhookService) Update(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID, req *dto.UpdateWebhookRequest) (*dto.WebhookResponse, error) {
	args := m.Called(ctx, teamID, webhookID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WebhookResponse), args.Error(1)
}
func (m *MockWebhookService) Delete(ctx context.Context, teamID uuid.UUID, webhookID uuid.UUID) error {
	return m.Called(ctx, teamID, webhookID).Error(0)
}

// --- InboundEmailService ---

type MockInboundEmailService struct{ mock.Mock }

func (m *MockInboundEmailService) List(ctx context.Context, teamID uuid.UUID, params *dto.PaginationParams) (*dto.PaginatedResponse[model.InboundEmail], error) {
	args := m.Called(ctx, teamID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[model.InboundEmail]), args.Error(1)
}
func (m *MockInboundEmailService) Get(ctx context.Context, teamID uuid.UUID, emailID uuid.UUID) (*model.InboundEmail, error) {
	args := m.Called(ctx, teamID, emailID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundEmail), args.Error(1)
}

// --- LogService ---

type MockLogService struct{ mock.Mock }

func (m *MockLogService) List(ctx context.Context, teamID uuid.UUID, level string, params *dto.PaginationParams) (*dto.PaginatedResponse[model.Log], error) {
	args := m.Called(ctx, teamID, level, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaginatedResponse[model.Log]), args.Error(1)
}

// --- MetricsService ---

type MockMetricsService struct{ mock.Mock }

func (m *MockMetricsService) Get(ctx context.Context, teamID uuid.UUID, period string) (*dto.MetricsResponse, error) {
	args := m.Called(ctx, teamID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MetricsResponse), args.Error(1)
}
func (m *MockMetricsService) IncrementCounter(ctx context.Context, teamID uuid.UUID, eventType string) error {
	return m.Called(ctx, teamID, eventType).Error(0)
}

// --- SettingsService ---

type MockSettingsService struct{ mock.Mock }

func (m *MockSettingsService) GetUsage(ctx context.Context, teamID uuid.UUID) (*dto.UsageResponse, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UsageResponse), args.Error(1)
}
func (m *MockSettingsService) GetTeam(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeamResponse), args.Error(1)
}
func (m *MockSettingsService) UpdateTeam(ctx context.Context, teamID uuid.UUID, req *dto.UpdateTeamRequest) error {
	return m.Called(ctx, teamID, req).Error(0)
}
func (m *MockSettingsService) GetSMTPConfig() *dto.SMTPConfigResponse {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*dto.SMTPConfigResponse)
}
func (m *MockSettingsService) InviteMember(ctx context.Context, teamID uuid.UUID, req *dto.InviteMemberRequest) (*model.TeamInvitation, error) {
	args := m.Called(ctx, teamID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TeamInvitation), args.Error(1)
}
func (m *MockSettingsService) AcceptInvite(ctx context.Context, req *dto.AcceptInviteRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

// --- TrackingService ---

type MockTrackingService struct{ mock.Mock }

func (m *MockTrackingService) HandleOpen(ctx context.Context, linkID uuid.UUID) error {
	return m.Called(ctx, linkID).Error(0)
}
func (m *MockTrackingService) HandleClick(ctx context.Context, linkID uuid.UUID) (string, error) {
	args := m.Called(ctx, linkID)
	return args.String(0), args.Error(1)
}
func (m *MockTrackingService) HandleUnsubscribe(ctx context.Context, linkID uuid.UUID) error {
	return m.Called(ctx, linkID).Error(0)
}
