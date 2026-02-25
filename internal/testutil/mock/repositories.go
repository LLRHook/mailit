package mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/model"
)

// --- UserRepository ---

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

// --- TeamRepository ---

type MockTeamRepository struct{ mock.Mock }

func (m *MockTeamRepository) Create(ctx context.Context, team *model.Team) error {
	return m.Called(ctx, team).Error(0)
}
func (m *MockTeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Team, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}
func (m *MockTeamRepository) GetBySlug(ctx context.Context, slug string) (*model.Team, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

// --- TeamMemberRepository ---

type MockTeamMemberRepository struct{ mock.Mock }

func (m *MockTeamMemberRepository) Create(ctx context.Context, member *model.TeamMember) error {
	return m.Called(ctx, member).Error(0)
}
func (m *MockTeamMemberRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.TeamMember, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.TeamMember), args.Error(1)
}
func (m *MockTeamMemberRepository) GetByTeamAndUser(ctx context.Context, teamID, userID uuid.UUID) (*model.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TeamMember), args.Error(1)
}

// --- EmailRepository ---

type MockEmailRepository struct{ mock.Mock }

func (m *MockEmailRepository) Create(ctx context.Context, email *model.Email) error {
	return m.Called(ctx, email).Error(0)
}
func (m *MockEmailRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Email, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Email), args.Error(1)
}
func (m *MockEmailRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Email, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Email), args.Error(1)
}
func (m *MockEmailRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Email, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Email), args.Int(1), args.Error(2)
}
func (m *MockEmailRepository) Update(ctx context.Context, email *model.Email) error {
	return m.Called(ctx, email).Error(0)
}

// --- EmailEventRepository ---

type MockEmailEventRepository struct{ mock.Mock }

func (m *MockEmailEventRepository) Create(ctx context.Context, event *model.EmailEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *MockEmailEventRepository) ListByEmailID(ctx context.Context, emailID uuid.UUID) ([]model.EmailEvent, error) {
	args := m.Called(ctx, emailID)
	return args.Get(0).([]model.EmailEvent), args.Error(1)
}

// --- DomainRepository ---

type MockDomainRepository struct{ mock.Mock }

func (m *MockDomainRepository) Create(ctx context.Context, domain *model.Domain) error {
	return m.Called(ctx, domain).Error(0)
}
func (m *MockDomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Domain, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *MockDomainRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Domain, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *MockDomainRepository) GetByTeamAndName(ctx context.Context, teamID uuid.UUID, name string) (*model.Domain, error) {
	args := m.Called(ctx, teamID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *MockDomainRepository) GetVerifiedByName(ctx context.Context, name string) (*model.Domain, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Domain), args.Error(1)
}
func (m *MockDomainRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Domain, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Domain), args.Int(1), args.Error(2)
}
func (m *MockDomainRepository) Update(ctx context.Context, domain *model.Domain) error {
	return m.Called(ctx, domain).Error(0)
}
func (m *MockDomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- DomainDNSRecordRepository ---

type MockDomainDNSRecordRepository struct{ mock.Mock }

func (m *MockDomainDNSRecordRepository) Create(ctx context.Context, record *model.DomainDNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockDomainDNSRecordRepository) ListByDomainID(ctx context.Context, domainID uuid.UUID) ([]model.DomainDNSRecord, error) {
	args := m.Called(ctx, domainID)
	return args.Get(0).([]model.DomainDNSRecord), args.Error(1)
}
func (m *MockDomainDNSRecordRepository) Update(ctx context.Context, record *model.DomainDNSRecord) error {
	return m.Called(ctx, record).Error(0)
}
func (m *MockDomainDNSRecordRepository) DeleteByDomainID(ctx context.Context, domainID uuid.UUID) error {
	return m.Called(ctx, domainID).Error(0)
}

// --- APIKeyRepository ---

type MockAPIKeyRepository struct{ mock.Mock }

func (m *MockAPIKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	return m.Called(ctx, key).Error(0)
}
func (m *MockAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.APIKey), args.Error(1)
}
func (m *MockAPIKeyRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.APIKey, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.APIKey), args.Error(1)
}
func (m *MockAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockAPIKeyRepository) UpdateLastUsed(ctx context.Context, keyHash string, usedAt time.Time) error {
	return m.Called(ctx, keyHash, usedAt).Error(0)
}

// --- AudienceRepository ---

type MockAudienceRepository struct{ mock.Mock }

func (m *MockAudienceRepository) Create(ctx context.Context, audience *model.Audience) error {
	return m.Called(ctx, audience).Error(0)
}
func (m *MockAudienceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Audience, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Audience), args.Error(1)
}
func (m *MockAudienceRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Audience, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Audience), args.Error(1)
}
func (m *MockAudienceRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Audience, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.Audience), args.Error(1)
}
func (m *MockAudienceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- ContactRepository ---

type MockContactRepository struct{ mock.Mock }

func (m *MockContactRepository) Create(ctx context.Context, contact *model.Contact) error {
	return m.Called(ctx, contact).Error(0)
}
func (m *MockContactRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *MockContactRepository) GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Contact, error) {
	args := m.Called(ctx, audienceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *MockContactRepository) GetByAudienceAndEmail(ctx context.Context, audienceID uuid.UUID, email string) (*model.Contact, error) {
	args := m.Called(ctx, audienceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contact), args.Error(1)
}
func (m *MockContactRepository) List(ctx context.Context, audienceID uuid.UUID, limit, offset int) ([]model.Contact, int, error) {
	args := m.Called(ctx, audienceID, limit, offset)
	return args.Get(0).([]model.Contact), args.Int(1), args.Error(2)
}
func (m *MockContactRepository) Update(ctx context.Context, contact *model.Contact) error {
	return m.Called(ctx, contact).Error(0)
}
func (m *MockContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- ContactPropertyRepository ---

type MockContactPropertyRepository struct{ mock.Mock }

func (m *MockContactPropertyRepository) Create(ctx context.Context, property *model.ContactProperty) error {
	return m.Called(ctx, property).Error(0)
}
func (m *MockContactPropertyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ContactProperty, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ContactProperty), args.Error(1)
}
func (m *MockContactPropertyRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.ContactProperty, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ContactProperty), args.Error(1)
}
func (m *MockContactPropertyRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.ContactProperty, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.ContactProperty), args.Error(1)
}
func (m *MockContactPropertyRepository) Update(ctx context.Context, property *model.ContactProperty) error {
	return m.Called(ctx, property).Error(0)
}
func (m *MockContactPropertyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- TopicRepository ---

type MockTopicRepository struct{ mock.Mock }

func (m *MockTopicRepository) Create(ctx context.Context, topic *model.Topic) error {
	return m.Called(ctx, topic).Error(0)
}
func (m *MockTopicRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Topic, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Topic), args.Error(1)
}
func (m *MockTopicRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Topic, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Topic), args.Error(1)
}
func (m *MockTopicRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Topic, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.Topic), args.Error(1)
}
func (m *MockTopicRepository) Update(ctx context.Context, topic *model.Topic) error {
	return m.Called(ctx, topic).Error(0)
}
func (m *MockTopicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- SegmentRepository ---

type MockSegmentRepository struct{ mock.Mock }

func (m *MockSegmentRepository) Create(ctx context.Context, segment *model.Segment) error {
	return m.Called(ctx, segment).Error(0)
}
func (m *MockSegmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Segment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Segment), args.Error(1)
}
func (m *MockSegmentRepository) GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Segment, error) {
	args := m.Called(ctx, audienceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Segment), args.Error(1)
}
func (m *MockSegmentRepository) ListByAudienceID(ctx context.Context, audienceID uuid.UUID) ([]model.Segment, error) {
	args := m.Called(ctx, audienceID)
	return args.Get(0).([]model.Segment), args.Error(1)
}
func (m *MockSegmentRepository) Update(ctx context.Context, segment *model.Segment) error {
	return m.Called(ctx, segment).Error(0)
}
func (m *MockSegmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- TemplateRepository ---

type MockTemplateRepository struct{ mock.Mock }

func (m *MockTemplateRepository) Create(ctx context.Context, template *model.Template) error {
	return m.Called(ctx, template).Error(0)
}
func (m *MockTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Template), args.Error(1)
}
func (m *MockTemplateRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Template, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Template), args.Error(1)
}
func (m *MockTemplateRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Template, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Template), args.Int(1), args.Error(2)
}
func (m *MockTemplateRepository) Update(ctx context.Context, template *model.Template) error {
	return m.Called(ctx, template).Error(0)
}
func (m *MockTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- TemplateVersionRepository ---

type MockTemplateVersionRepository struct{ mock.Mock }

func (m *MockTemplateVersionRepository) Create(ctx context.Context, version *model.TemplateVersion) error {
	return m.Called(ctx, version).Error(0)
}
func (m *MockTemplateVersionRepository) GetLatestByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error) {
	args := m.Called(ctx, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TemplateVersion), args.Error(1)
}
func (m *MockTemplateVersionRepository) GetPublishedByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error) {
	args := m.Called(ctx, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TemplateVersion), args.Error(1)
}
func (m *MockTemplateVersionRepository) ListByTemplateID(ctx context.Context, templateID uuid.UUID) ([]model.TemplateVersion, error) {
	args := m.Called(ctx, templateID)
	return args.Get(0).([]model.TemplateVersion), args.Error(1)
}
func (m *MockTemplateVersionRepository) Publish(ctx context.Context, versionID uuid.UUID) error {
	return m.Called(ctx, versionID).Error(0)
}

// --- BroadcastRepository ---

type MockBroadcastRepository struct{ mock.Mock }

func (m *MockBroadcastRepository) Create(ctx context.Context, broadcast *model.Broadcast) error {
	return m.Called(ctx, broadcast).Error(0)
}
func (m *MockBroadcastRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Broadcast, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Broadcast), args.Error(1)
}
func (m *MockBroadcastRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Broadcast, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Broadcast), args.Error(1)
}
func (m *MockBroadcastRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Broadcast, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.Broadcast), args.Int(1), args.Error(2)
}
func (m *MockBroadcastRepository) Update(ctx context.Context, broadcast *model.Broadcast) error {
	return m.Called(ctx, broadcast).Error(0)
}
func (m *MockBroadcastRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- WebhookRepository ---

type MockWebhookRepository struct{ mock.Mock }

func (m *MockWebhookRepository) Create(ctx context.Context, webhook *model.Webhook) error {
	return m.Called(ctx, webhook).Error(0)
}
func (m *MockWebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Webhook), args.Error(1)
}
func (m *MockWebhookRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Webhook, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Webhook), args.Error(1)
}
func (m *MockWebhookRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Webhook, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).([]model.Webhook), args.Error(1)
}
func (m *MockWebhookRepository) Update(ctx context.Context, webhook *model.Webhook) error {
	return m.Called(ctx, webhook).Error(0)
}
func (m *MockWebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- WebhookEventRepository ---

type MockWebhookEventRepository struct{ mock.Mock }

func (m *MockWebhookEventRepository) Create(ctx context.Context, event *model.WebhookEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *MockWebhookEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.WebhookEvent), args.Error(1)
}
func (m *MockWebhookEventRepository) Update(ctx context.Context, event *model.WebhookEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *MockWebhookEventRepository) ListByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]model.WebhookEvent, int, error) {
	args := m.Called(ctx, webhookID, limit, offset)
	return args.Get(0).([]model.WebhookEvent), args.Int(1), args.Error(2)
}
func (m *MockWebhookEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

// --- SuppressionRepository ---

type MockSuppressionRepository struct{ mock.Mock }

func (m *MockSuppressionRepository) Create(ctx context.Context, entry *model.SuppressionEntry) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *MockSuppressionRepository) GetByTeamAndEmail(ctx context.Context, teamID uuid.UUID, email string) (*model.SuppressionEntry, error) {
	args := m.Called(ctx, teamID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SuppressionEntry), args.Error(1)
}
func (m *MockSuppressionRepository) ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.SuppressionEntry, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.SuppressionEntry), args.Int(1), args.Error(2)
}
func (m *MockSuppressionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// --- InboundEmailRepository ---

type MockInboundEmailRepository struct{ mock.Mock }

func (m *MockInboundEmailRepository) Create(ctx context.Context, email *model.InboundEmail) error {
	return m.Called(ctx, email).Error(0)
}
func (m *MockInboundEmailRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InboundEmail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundEmail), args.Error(1)
}
func (m *MockInboundEmailRepository) GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.InboundEmail, error) {
	args := m.Called(ctx, teamID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundEmail), args.Error(1)
}
func (m *MockInboundEmailRepository) List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.InboundEmail, int, error) {
	args := m.Called(ctx, teamID, limit, offset)
	return args.Get(0).([]model.InboundEmail), args.Int(1), args.Error(2)
}
func (m *MockInboundEmailRepository) Update(ctx context.Context, email *model.InboundEmail) error {
	return m.Called(ctx, email).Error(0)
}

// --- LogRepository ---

type MockLogRepository struct{ mock.Mock }

func (m *MockLogRepository) Create(ctx context.Context, log *model.Log) error {
	return m.Called(ctx, log).Error(0)
}
func (m *MockLogRepository) List(ctx context.Context, teamID uuid.UUID, level string, limit, offset int) ([]model.Log, int, error) {
	args := m.Called(ctx, teamID, level, limit, offset)
	return args.Get(0).([]model.Log), args.Int(1), args.Error(2)
}

// --- MetricsRepository ---

type MockMetricsRepository struct{ mock.Mock }

func (m *MockMetricsRepository) Upsert(ctx context.Context, metrics *model.EmailMetrics) error {
	return m.Called(ctx, metrics).Error(0)
}
func (m *MockMetricsRepository) ListByTeam(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) ([]model.EmailMetrics, error) {
	args := m.Called(ctx, teamID, periodType, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.EmailMetrics), args.Error(1)
}
func (m *MockMetricsRepository) AggregateTotals(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) (*model.EmailMetrics, error) {
	args := m.Called(ctx, teamID, periodType, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.EmailMetrics), args.Error(1)
}
