package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

// TeamRepository defines persistence operations for teams.
type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Team, error)
	GetBySlug(ctx context.Context, slug string) (*model.Team, error)
}

// TeamMemberRepository defines persistence operations for team members.
type TeamMemberRepository interface {
	Create(ctx context.Context, member *model.TeamMember) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.TeamMember, error)
	GetByTeamAndUser(ctx context.Context, teamID, userID uuid.UUID) (*model.TeamMember, error)
}

// EmailRepository defines persistence operations for emails.
type EmailRepository interface {
	Create(ctx context.Context, email *model.Email) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Email, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Email, error)
	List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Email, int, error)
	Update(ctx context.Context, email *model.Email) error
}

// EmailEventRepository defines persistence operations for email events.
type EmailEventRepository interface {
	Create(ctx context.Context, event *model.EmailEvent) error
	ListByEmailID(ctx context.Context, emailID uuid.UUID) ([]model.EmailEvent, error)
}

// DomainRepository defines persistence operations for domains.
type DomainRepository interface {
	Create(ctx context.Context, domain *model.Domain) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Domain, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Domain, error)
	GetByTeamAndName(ctx context.Context, teamID uuid.UUID, name string) (*model.Domain, error)
	GetVerifiedByName(ctx context.Context, name string) (*model.Domain, error)
	List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Domain, int, error)
	Update(ctx context.Context, domain *model.Domain) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DomainDNSRecordRepository defines persistence operations for domain DNS records.
type DomainDNSRecordRepository interface {
	Create(ctx context.Context, record *model.DomainDNSRecord) error
	ListByDomainID(ctx context.Context, domainID uuid.UUID) ([]model.DomainDNSRecord, error)
	Update(ctx context.Context, record *model.DomainDNSRecord) error
	DeleteByDomainID(ctx context.Context, domainID uuid.UUID) error
}

// APIKeyRepository defines persistence operations for API keys.
type APIKeyRepository interface {
	Create(ctx context.Context, key *model.APIKey) error
	GetByHash(ctx context.Context, keyHash string) (*model.APIKey, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.APIKey, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, keyHash string, usedAt time.Time) error
}

// AudienceRepository defines persistence operations for audiences.
type AudienceRepository interface {
	Create(ctx context.Context, audience *model.Audience) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Audience, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Audience, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Audience, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ContactRepository defines persistence operations for contacts.
type ContactRepository interface {
	Create(ctx context.Context, contact *model.Contact) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Contact, error)
	GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Contact, error)
	GetByAudienceAndEmail(ctx context.Context, audienceID uuid.UUID, email string) (*model.Contact, error)
	List(ctx context.Context, audienceID uuid.UUID, limit, offset int) ([]model.Contact, int, error)
	Update(ctx context.Context, contact *model.Contact) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ContactPropertyRepository defines persistence operations for contact properties.
type ContactPropertyRepository interface {
	Create(ctx context.Context, property *model.ContactProperty) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ContactProperty, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.ContactProperty, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.ContactProperty, error)
	Update(ctx context.Context, property *model.ContactProperty) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TopicRepository defines persistence operations for topics.
type TopicRepository interface {
	Create(ctx context.Context, topic *model.Topic) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Topic, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Topic, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Topic, error)
	Update(ctx context.Context, topic *model.Topic) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SegmentRepository defines persistence operations for segments.
type SegmentRepository interface {
	Create(ctx context.Context, segment *model.Segment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Segment, error)
	GetByAudienceAndID(ctx context.Context, audienceID, id uuid.UUID) (*model.Segment, error)
	ListByAudienceID(ctx context.Context, audienceID uuid.UUID) ([]model.Segment, error)
	Update(ctx context.Context, segment *model.Segment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TemplateRepository defines persistence operations for templates.
type TemplateRepository interface {
	Create(ctx context.Context, template *model.Template) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Template, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Template, error)
	List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Template, int, error)
	Update(ctx context.Context, template *model.Template) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TemplateVersionRepository defines persistence operations for template versions.
type TemplateVersionRepository interface {
	Create(ctx context.Context, version *model.TemplateVersion) error
	GetLatestByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error)
	GetPublishedByTemplateID(ctx context.Context, templateID uuid.UUID) (*model.TemplateVersion, error)
	ListByTemplateID(ctx context.Context, templateID uuid.UUID) ([]model.TemplateVersion, error)
	Publish(ctx context.Context, versionID uuid.UUID) error
}

// BroadcastRepository defines persistence operations for broadcasts.
type BroadcastRepository interface {
	Create(ctx context.Context, broadcast *model.Broadcast) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Broadcast, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Broadcast, error)
	List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.Broadcast, int, error)
	Update(ctx context.Context, broadcast *model.Broadcast) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookRepository defines persistence operations for webhooks.
type WebhookRepository interface {
	Create(ctx context.Context, webhook *model.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Webhook, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.Webhook, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.Webhook, error)
	Update(ctx context.Context, webhook *model.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookEventRepository defines persistence operations for webhook events.
type WebhookEventRepository interface {
	Create(ctx context.Context, event *model.WebhookEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookEvent, error)
	Update(ctx context.Context, event *model.WebhookEvent) error
	ListByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]model.WebhookEvent, int, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// SuppressionRepository defines persistence operations for the suppression list.
type SuppressionRepository interface {
	Create(ctx context.Context, entry *model.SuppressionEntry) error
	GetByTeamAndEmail(ctx context.Context, teamID uuid.UUID, email string) (*model.SuppressionEntry, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.SuppressionEntry, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// InboundEmailRepository defines persistence operations for inbound emails.
type InboundEmailRepository interface {
	Create(ctx context.Context, email *model.InboundEmail) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InboundEmail, error)
	GetByTeamAndID(ctx context.Context, teamID, id uuid.UUID) (*model.InboundEmail, error)
	List(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]model.InboundEmail, int, error)
	Update(ctx context.Context, email *model.InboundEmail) error
}

// LogRepository defines persistence operations for logs.
type LogRepository interface {
	Create(ctx context.Context, log *model.Log) error
	List(ctx context.Context, teamID uuid.UUID, level string, limit, offset int) ([]model.Log, int, error)
}

// SettingsRepository defines read operations for the settings page.
type SettingsRepository interface {
	GetUsageCounts(ctx context.Context, teamID uuid.UUID) (*dto.UsageResponse, error)
	GetTeamWithMembers(ctx context.Context, teamID uuid.UUID) (*dto.TeamResponse, error)
	UpdateTeamName(ctx context.Context, teamID uuid.UUID, name string) error
}

// TeamInvitationRepository defines persistence operations for team invitations.
type TeamInvitationRepository interface {
	Create(ctx context.Context, invitation *model.TeamInvitation) error
	GetByToken(ctx context.Context, token string) (*model.TeamInvitation, error)
	ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]model.TeamInvitation, error)
	MarkAccepted(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MetricsRepository defines persistence operations for email metrics.
type MetricsRepository interface {
	Upsert(ctx context.Context, m *model.EmailMetrics) error
	ListByTeam(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) ([]model.EmailMetrics, error)
	AggregateTotals(ctx context.Context, teamID uuid.UUID, periodType string, from, to time.Time) (*model.EmailMetrics, error)
}
