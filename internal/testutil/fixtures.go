package testutil

import (
	"time"

	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/model"
)

var (
	FixedTime = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	TestTeamID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	TestUserID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func NewTestUser() *model.User {
	return &model.User{
		ID:            TestUserID,
		Email:         "test@example.com",
		PasswordHash:  "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012",
		Name:          "Test User",
		EmailVerified: true,
		CreatedAt:     FixedTime,
		UpdatedAt:     FixedTime,
	}
}

func NewTestTeam() *model.Team {
	return &model.Team{
		ID:        TestTeamID,
		Name:      "Test Team",
		Slug:      "test-team",
		CreatedAt: FixedTime,
		UpdatedAt: FixedTime,
	}
}

func NewTestTeamMember() *model.TeamMember {
	return &model.TeamMember{
		ID:        uuid.New(),
		TeamID:    TestTeamID,
		UserID:    TestUserID,
		Role:      model.RoleOwner,
		CreatedAt: FixedTime,
	}
}

func NewTestDomain() *model.Domain {
	privKey := "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"
	return &model.Domain{
		ID:             uuid.New(),
		TeamID:         TestTeamID,
		Name:           "example.com",
		Status:         model.DomainStatusPending,
		DKIMPrivateKey: &privKey,
		DKIMSelector:   "mailit",
		OpenTracking:   false,
		ClickTracking:  false,
		TLSPolicy:      "opportunistic",
		CreatedAt:      FixedTime,
		UpdatedAt:      FixedTime,
	}
}

func NewTestEmail() *model.Email {
	html := "<p>Hello</p>"
	text := "Hello"
	return &model.Email{
		ID:          uuid.New(),
		TeamID:      TestTeamID,
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Test Subject",
		HTMLBody:    &html,
		TextBody:    &text,
		Status:      model.EmailStatusQueued,
		Tags:        []string{},
		Headers:     model.JSONMap{},
		Attachments: model.JSONArray{},
		RetryCount:  0,
		CreatedAt:   FixedTime,
		UpdatedAt:   FixedTime,
	}
}

func NewTestAPIKey() *model.APIKey {
	return &model.APIKey{
		ID:         uuid.New(),
		TeamID:     TestTeamID,
		Name:       "Test Key",
		KeyHash:    "abc123hash",
		KeyPrefix:  "re_1234abcd...",
		Permission: model.PermissionFull,
		CreatedAt:  FixedTime,
	}
}

func NewTestAudience() *model.Audience {
	return &model.Audience{
		ID:        uuid.New(),
		TeamID:    TestTeamID,
		Name:      "Test Audience",
		CreatedAt: FixedTime,
		UpdatedAt: FixedTime,
	}
}

func NewTestContact(audienceID uuid.UUID) *model.Contact {
	first := "John"
	last := "Doe"
	return &model.Contact{
		ID:         uuid.New(),
		AudienceID: audienceID,
		Email:      "john@example.com",
		FirstName:  &first,
		LastName:   &last,
		CreatedAt:  FixedTime,
		UpdatedAt:  FixedTime,
	}
}

func NewTestTemplate() *model.Template {
	desc := "A test template"
	return &model.Template{
		ID:          uuid.New(),
		TeamID:      TestTeamID,
		Name:        "Test Template",
		Description: &desc,
		CreatedAt:   FixedTime,
		UpdatedAt:   FixedTime,
	}
}

func NewTestTemplateVersion(templateID uuid.UUID) *model.TemplateVersion {
	subject := "Hello {{name}}"
	html := "<p>Hello {{name}}</p>"
	text := "Hello {{name}}"
	return &model.TemplateVersion{
		ID:         uuid.New(),
		TemplateID: templateID,
		Version:    1,
		Subject:    &subject,
		HTMLBody:   &html,
		TextBody:   &text,
		Variables:  model.JSONArray{},
		Published:  false,
		CreatedAt:  FixedTime,
	}
}

func NewTestWebhook() *model.Webhook {
	return &model.Webhook{
		ID:            uuid.New(),
		TeamID:        TestTeamID,
		URL:           "https://example.com/webhook",
		Events:        []string{"email.sent", "email.bounced"},
		SigningSecret: "whsec_test_secret_123",
		Active:        true,
		CreatedAt:     FixedTime,
		UpdatedAt:     FixedTime,
	}
}

func NewTestBroadcast() *model.Broadcast {
	from := "newsletter@example.com"
	subject := "Test Broadcast"
	html := "<p>Hello subscribers</p>"
	audienceID := uuid.New()
	return &model.Broadcast{
		ID:          uuid.New(),
		TeamID:      TestTeamID,
		Name:        "Test Broadcast",
		AudienceID:  &audienceID,
		Status:      model.BroadcastStatusDraft,
		FromAddress: &from,
		Subject:     &subject,
		HTMLBody:    &html,
		CreatedAt:   FixedTime,
		UpdatedAt:   FixedTime,
	}
}

func NewTestSuppressionEntry() *model.SuppressionEntry {
	return &model.SuppressionEntry{
		ID:        uuid.New(),
		TeamID:    TestTeamID,
		Email:     "suppressed@example.com",
		Reason:    "hard_bounce",
		CreatedAt: FixedTime,
	}
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string { return &s }

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool { return &b }

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int { return &i }
