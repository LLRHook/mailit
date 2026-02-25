package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs matching DTO validate tags.

type testLoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type testSendEmailRequest struct {
	From    string   `validate:"required,email"`
	To      []string `validate:"required,min=1,dive,email"`
	Subject string   `validate:"required"`
}

type testCreateDomainRequest struct {
	Name string `validate:"required,fqdn"`
}

type testCreateWebhookRequest struct {
	URL    string   `validate:"required,url"`
	Events []string `validate:"required,min=1"`
}

func TestValidate(t *testing.T) {
	t.Run("valid struct passes", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "alice@example.com",
			Password: "securepassword",
		}
		err := Validate(req)
		assert.NoError(t, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		req := testLoginRequest{}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("invalid email format", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "not-an-email",
			Password: "securepassword",
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("password too short", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "alice@example.com",
			Password: "short",
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("valid send email request", func(t *testing.T) {
		req := testSendEmailRequest{
			From:    "sender@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Hello",
		}
		err := Validate(req)
		assert.NoError(t, err)
	})

	t.Run("send email with invalid recipient", func(t *testing.T) {
		req := testSendEmailRequest{
			From:    "sender@example.com",
			To:      []string{"not-an-email"},
			Subject: "Hello",
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("send email with empty To list", func(t *testing.T) {
		req := testSendEmailRequest{
			From:    "sender@example.com",
			To:      []string{},
			Subject: "Hello",
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("valid domain request", func(t *testing.T) {
		req := testCreateDomainRequest{
			Name: "example.com",
		}
		err := Validate(req)
		assert.NoError(t, err)
	})

	t.Run("invalid domain request - empty name", func(t *testing.T) {
		req := testCreateDomainRequest{
			Name: "",
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("valid webhook request", func(t *testing.T) {
		req := testCreateWebhookRequest{
			URL:    "https://example.com/webhook",
			Events: []string{"email.sent"},
		}
		err := Validate(req)
		assert.NoError(t, err)
	})

	t.Run("webhook request with invalid URL", func(t *testing.T) {
		req := testCreateWebhookRequest{
			URL:    "not-a-url",
			Events: []string{"email.sent"},
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("webhook request with empty events", func(t *testing.T) {
		req := testCreateWebhookRequest{
			URL:    "https://example.com/webhook",
			Events: []string{},
		}
		err := Validate(req)
		assert.Error(t, err)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "bad-email",
			Password: "short",
		}
		err := Validate(req)
		require.Error(t, err)

		errors := ValidationErrors(err)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors, "Email")
		assert.Contains(t, errors, "Password")
	})
}

func TestValidationErrors(t *testing.T) {
	t.Run("extracts field to tag mapping", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "",
			Password: "",
		}
		err := Validate(req)
		require.Error(t, err)

		errors := ValidationErrors(err)
		assert.Equal(t, "required", errors["Email"])
		assert.Equal(t, "required", errors["Password"])
	})

	t.Run("returns empty map for non-validation errors", func(t *testing.T) {
		err := assert.AnError
		errors := ValidationErrors(err)
		assert.Empty(t, errors)
	})

	t.Run("min tag appears for short password", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "alice@example.com",
			Password: "short",
		}
		err := Validate(req)
		require.Error(t, err)

		errors := ValidationErrors(err)
		assert.Equal(t, "min", errors["Password"])
	})

	t.Run("email tag appears for bad email", func(t *testing.T) {
		req := testLoginRequest{
			Email:    "not-email",
			Password: "longpassword",
		}
		err := Validate(req)
		require.Error(t, err)

		errors := ValidationErrors(err)
		assert.Equal(t, "email", errors["Email"])
	})

	t.Run("nil error returns empty map", func(t *testing.T) {
		errors := ValidationErrors(nil)
		assert.Empty(t, errors)
	})
}
