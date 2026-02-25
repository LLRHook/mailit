package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMessage_SingleTextPart(t *testing.T) {
	msg := &OutgoingMessage{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello",
		TextBody: "This is plain text.",
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "From: sender@example.com")
	assert.Contains(t, body, "To: recipient@example.com")
	assert.Contains(t, body, "Subject: Hello")
	assert.Contains(t, body, "Content-Type: text/plain; charset=utf-8")
	assert.Contains(t, body, "Content-Transfer-Encoding: quoted-printable")
	assert.Contains(t, body, "Mime-Version: 1.0")
	assert.Contains(t, body, "This is plain text.")
	// Should NOT contain multipart
	assert.NotContains(t, body, "multipart/alternative")
	assert.NotContains(t, body, "multipart/mixed")
}

func TestBuildMessage_SingleHTMLPart(t *testing.T) {
	msg := &OutgoingMessage{
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "Hello HTML",
		HTMLBody: "<h1>Hello</h1>",
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "Content-Type: text/html; charset=utf-8")
	assert.Contains(t, body, "Content-Transfer-Encoding: quoted-printable")
	assert.Contains(t, body, "<h1>Hello</h1>")
	assert.NotContains(t, body, "multipart/alternative")
}

func TestBuildMessage_MultipartAlternative(t *testing.T) {
	msg := &OutgoingMessage{
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "Dual Content",
		TextBody: "Plain text version",
		HTMLBody: "<p>HTML version</p>",
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "multipart/alternative")
	assert.Contains(t, body, "text/plain; charset=utf-8")
	assert.Contains(t, body, "text/html; charset=utf-8")
	assert.Contains(t, body, "Plain text version")
	assert.Contains(t, body, "<p>HTML version</p>")
}

func TestBuildMessage_MultipartMixed(t *testing.T) {
	msg := &OutgoingMessage{
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "With Attachments",
		TextBody: "See attached.",
		HTMLBody: "<p>See attached.</p>",
		Attachments: []MessageAttachment{
			{
				Filename:    "test.txt",
				Content:     []byte("file content here"),
				ContentType: "text/plain",
			},
		},
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "multipart/mixed")
	assert.Contains(t, body, "multipart/alternative")
	assert.Contains(t, body, "See attached.")
	assert.Contains(t, body, "<p>See attached.</p>")
	assert.Contains(t, body, "test.txt")
	assert.Contains(t, body, "Content-Transfer-Encoding: base64")
	assert.Contains(t, body, "Content-Disposition: attachment")
}

func TestBuildMessage_MultipartMixed_HTMLOnly(t *testing.T) {
	msg := &OutgoingMessage{
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Subject:  "HTML with Attachment",
		HTMLBody: "<p>Body</p>",
		Attachments: []MessageAttachment{
			{
				Filename:    "doc.pdf",
				Content:     []byte("pdf-data"),
				ContentType: "application/pdf",
			},
		},
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "multipart/mixed")
	assert.Contains(t, body, "text/html; charset=utf-8")
	assert.Contains(t, body, "doc.pdf")
	// Should NOT have multipart/alternative since only HTML is present
	assert.NotContains(t, body, "multipart/alternative")
}

func TestBuildMessage_HeadersPresent(t *testing.T) {
	msg := &OutgoingMessage{
		From:      "sender@example.com",
		To:        []string{"to@example.com"},
		Cc:        []string{"cc1@example.com", "cc2@example.com"},
		ReplyTo:   "reply@example.com",
		Subject:   "Headers Test",
		TextBody:  "body",
		MessageID: "abc-123@example.com",
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "To: to@example.com")
	assert.Contains(t, body, "Cc: cc1@example.com, cc2@example.com")
	assert.Contains(t, body, "Reply-To: reply@example.com")
	assert.Contains(t, body, "Message-Id: <abc-123@example.com>")
	assert.Contains(t, body, "X-Custom-Header: custom-value")
	assert.Contains(t, body, "Date:")
	// Bcc should NOT appear in headers
	assert.NotContains(t, body, "Bcc")
}

func TestBuildMessage_DefaultAttachmentContentType(t *testing.T) {
	msg := &OutgoingMessage{
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		Subject: "Attachment default type",
		TextBody: "body",
		Attachments: []MessageAttachment{
			{
				Filename: "unknown.bin",
				Content:  []byte("binary data"),
				// ContentType intentionally empty
			},
		},
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, "application/octet-stream")
}

func TestBuildMessage_EmptyBody(t *testing.T) {
	msg := &OutgoingMessage{
		From:    "sender@example.com",
		To:      []string{"to@example.com"},
		Subject: "Empty",
	}

	raw, err := BuildMessage(msg)
	require.NoError(t, err)
	body := string(raw)

	// Default to text/plain with empty body
	assert.Contains(t, body, "Content-Type: text/plain; charset=utf-8")
}

func TestCollectRecipients(t *testing.T) {
	tests := []struct {
		name string
		msg  *OutgoingMessage
		want []string
	}{
		{
			name: "deduplication across To, Cc, Bcc",
			msg: &OutgoingMessage{
				To:  []string{"alice@example.com", "bob@example.com"},
				Cc:  []string{"bob@example.com", "charlie@example.com"},
				Bcc: []string{"alice@example.com", "dave@example.com"},
			},
			want: []string{"alice@example.com", "bob@example.com", "charlie@example.com", "dave@example.com"},
		},
		{
			name: "case normalization",
			msg: &OutgoingMessage{
				To:  []string{"Alice@Example.COM"},
				Cc:  []string{"alice@example.com"},
				Bcc: []string{"ALICE@EXAMPLE.COM"},
			},
			want: []string{"alice@example.com"},
		},
		{
			name: "whitespace trimming",
			msg: &OutgoingMessage{
				To: []string{"  alice@example.com  ", "bob@example.com "},
			},
			want: []string{"alice@example.com", "bob@example.com"},
		},
		{
			name: "empty strings filtered",
			msg: &OutgoingMessage{
				To:  []string{"alice@example.com", "", "  "},
				Cc:  []string{""},
				Bcc: []string{},
			},
			want: []string{"alice@example.com"},
		},
		{
			name: "all lists empty",
			msg: &OutgoingMessage{
				To:  []string{},
				Cc:  []string{},
				Bcc: []string{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectRecipients(tt.msg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGroupByDomain(t *testing.T) {
	tests := []struct {
		name       string
		recipients []string
		want       map[string][]string
	}{
		{
			name:       "group by domain",
			recipients: []string{"alice@example.com", "bob@example.com", "charlie@other.com"},
			want: map[string][]string{
				"example.com": {"alice@example.com", "bob@example.com"},
				"other.com":   {"charlie@other.com"},
			},
		},
		{
			name:       "domain is lowercased",
			recipients: []string{"alice@Example.COM"},
			want: map[string][]string{
				"example.com": {"alice@Example.COM"},
			},
		},
		{
			name:       "invalid address without @ is skipped",
			recipients: []string{"invalid-address", "valid@example.com"},
			want: map[string][]string{
				"example.com": {"valid@example.com"},
			},
		},
		{
			name:       "empty list",
			recipients: []string{},
			want:       map[string][]string{},
		},
		{
			name:       "single recipient",
			recipients: []string{"user@domain.com"},
			want: map[string][]string{
				"domain.com": {"user@domain.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupByDomain(tt.recipients)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSmtpError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: 0,
			wantMsg:  "",
		},
		{
			name:     "550 SMTP error",
			err:      errors.New("550 5.1.1 User unknown"),
			wantCode: 550,
			wantMsg:  "5.1.1 User unknown",
		},
		{
			name:     "421 SMTP error",
			err:      errors.New("421 Service not available"),
			wantCode: 421,
			wantMsg:  "Service not available",
		},
		{
			name:     "250 success code",
			err:      errors.New("250 OK"),
			wantCode: 250,
			wantMsg:  "OK",
		},
		{
			name:     "timeout error",
			err:      errors.New("i/o timeout"),
			wantCode: 421,
			wantMsg:  "i/o timeout",
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp: connection refused"),
			wantCode: 421,
			wantMsg:  "dial tcp: connection refused",
		},
		{
			name:     "unknown error format",
			err:      errors.New("something went wrong"),
			wantCode: 0,
			wantMsg:  "something went wrong",
		},
		{
			name:     "short error message",
			err:      errors.New("ab"),
			wantCode: 0,
			wantMsg:  "ab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := parseSmtpError(tt.err)
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}

func TestEncodeSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		check   func(t *testing.T, result string)
	}{
		{
			name:    "ASCII subject stays unchanged",
			subject: "Hello World",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "Hello World", result)
			},
		},
		{
			name:    "empty subject stays empty",
			subject: "",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "", result)
			},
		},
		{
			name:    "numbers and symbols in ASCII",
			subject: "Order #12345 - Confirmation!",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "Order #12345 - Confirmation!", result)
			},
		},
		{
			name:    "non-ASCII characters are encoded",
			subject: "Hallo Welt - Gruesse",
			check: func(t *testing.T, result string) {
				// Pure ASCII, should stay as is
				assert.Equal(t, "Hallo Welt - Gruesse", result)
			},
		},
		{
			name:    "unicode subject is RFC 2047 encoded",
			subject: "Bonjour le monde!",
			check: func(t *testing.T, result string) {
				// Pure ASCII, should stay as is
				assert.Equal(t, "Bonjour le monde!", result)
			},
		},
		{
			name:    "subject with emoji gets encoded",
			subject: "Hello \U0001F600 World",
			check: func(t *testing.T, result string) {
				assert.True(t, strings.HasPrefix(result, "=?utf-8?q?") || strings.HasPrefix(result, "=?UTF-8?Q?"),
					"expected RFC 2047 encoding, got: %s", result)
			},
		},
		{
			name:    "subject with accented characters gets encoded",
			subject: "Cafe\u0301 au lait",
			check: func(t *testing.T, result string) {
				assert.True(t, strings.Contains(result, "=?utf-8?q?") || strings.Contains(result, "=?UTF-8?Q?"),
					"expected RFC 2047 encoding, got: %s", result)
			},
		},
		{
			name:    "Japanese subject gets encoded",
			subject: "\u3053\u3093\u306b\u3061\u306f",
			check: func(t *testing.T, result string) {
				assert.True(t, strings.Contains(result, "=?utf-8?q?") || strings.Contains(result, "=?UTF-8?Q?"),
					"expected RFC 2047 encoding for Japanese, got: %s", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeSubject(tt.subject)
			tt.check(t, result)
		})
	}
}

func TestStatusFromBounce(t *testing.T) {
	tests := []struct {
		name       string
		bounceType BounceType
		want       string
	}{
		{"hard bounce", BounceHard, "failed"},
		{"soft bounce", BounceSoft, "deferred"},
		{"complaint", BounceComplaint, "failed"},
		{"empty/unknown type", BounceType(""), "deferred"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusFromBounce(BounceInfo{Type: tt.bounceType})
			assert.Equal(t, tt.want, got)
		})
	}
}
