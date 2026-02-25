package engine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyBounce(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		message   string
		wantType  BounceType
		wantPerm  bool
	}{
		{
			name:     "550 hard bounce - mailbox not found",
			code:     550,
			message:  "User unknown",
			wantType: BounceHard,
			wantPerm: true,
		},
		{
			name:     "554 hard bounce - transaction failed",
			code:     554,
			message:  "Transaction failed",
			wantType: BounceHard,
			wantPerm: true,
		},
		{
			name:     "551 hard bounce - user not local",
			code:     551,
			message:  "User not local",
			wantType: BounceHard,
			wantPerm: true,
		},
		{
			name:     "553 hard bounce - mailbox name not allowed",
			code:     553,
			message:  "Mailbox name not allowed",
			wantType: BounceHard,
			wantPerm: true,
		},
		{
			name:     "421 soft bounce - service not available",
			code:     421,
			message:  "Service not available",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "450 soft bounce - mailbox unavailable",
			code:     450,
			message:  "Mailbox unavailable",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "451 soft bounce - local error",
			code:     451,
			message:  "Requested action aborted: local error in processing",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "452 soft bounce - insufficient storage",
			code:     452,
			message:  "Insufficient system storage",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "552 quota full becomes soft bounce",
			code:     552,
			message:  "Mailbox full - quota exceeded",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "552 over quota becomes soft bounce",
			code:     552,
			message:  "User over quota",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "552 storage becomes soft bounce",
			code:     552,
			message:  "Insufficient storage",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "552 without quota keyword stays hard bounce",
			code:     552,
			message:  "Message too large",
			wantType: BounceHard,
			wantPerm: true,
		},
		{
			name:     "spam complaint in 550 message",
			code:     550,
			message:  "This message was identified as spam",
			wantType: BounceComplaint,
			wantPerm: true,
		},
		{
			name:     "abuse complaint",
			code:     550,
			message:  "Reported as abuse by recipient",
			wantType: BounceComplaint,
			wantPerm: true,
		},
		{
			name:     "unsolicited complaint",
			code:     554,
			message:  "Rejected: unsolicited email",
			wantType: BounceComplaint,
			wantPerm: true,
		},
		{
			name:     "complaint keyword in 4xx message",
			code:     450,
			message:  "Blocked for spam, please retry later",
			wantType: BounceComplaint,
			wantPerm: true,
		},
		{
			name:     "complaint keyword overrides code range",
			code:     421,
			message:  "Your message was flagged as complaint content",
			wantType: BounceComplaint,
			wantPerm: true,
		},
		{
			name:     "unknown code defaults to soft bounce",
			code:     199,
			message:  "Something unexpected",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "zero code defaults to soft bounce",
			code:     0,
			message:  "Connection error",
			wantType: BounceSoft,
			wantPerm: false,
		},
		{
			name:     "600+ code defaults to soft bounce",
			code:     600,
			message:  "Unknown error",
			wantType: BounceSoft,
			wantPerm: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ClassifyBounce(tt.code, tt.message)
			assert.Equal(t, tt.wantType, info.Type, "bounce type")
			assert.Equal(t, tt.wantPerm, info.Permanent, "permanent flag")
			assert.Equal(t, tt.code, info.Code, "code preserved")
			assert.Equal(t, tt.message, info.Message, "message preserved")
		})
	}
}

func TestClassifyDSN(t *testing.T) {
	t.Run("valid multipart/report DSN message", func(t *testing.T) {
		boundary := "boundary123"
		rawMessage := fmt.Sprintf(
			"From: mailer-daemon@example.com\r\n"+
				"To: sender@example.com\r\n"+
				"Subject: Delivery Status Notification\r\n"+
				"Content-Type: multipart/report; report-type=delivery-status; boundary=%s\r\n"+
				"\r\n"+
				"--%s\r\n"+
				"Content-Type: text/plain\r\n"+
				"\r\n"+
				"Your message could not be delivered.\r\n"+
				"--%s\r\n"+
				"Content-Type: message/delivery-status\r\n"+
				"\r\n"+
				"Reporting-MTA: dns; example.com\r\n"+
				"\r\n"+
				"Final-Recipient: rfc822;bob@example.com\r\n"+
				"Action: failed\r\n"+
				"Status: 5.1.1\r\n"+
				"Diagnostic-Code: smtp; 550 5.1.1 User unknown\r\n"+
				"--%s--\r\n",
			boundary, boundary, boundary, boundary,
		)

		info, err := ClassifyDSN([]byte(rawMessage))
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "bob@example.com", info.Recipient)
		assert.Equal(t, 550, info.Code)
		assert.True(t, info.Permanent)
		assert.Equal(t, BounceHard, info.Type)
	})

	t.Run("DSN with soft bounce status 4.2.2", func(t *testing.T) {
		boundary := "softboundary"
		rawMessage := fmt.Sprintf(
			"From: mailer-daemon@example.com\r\n"+
				"To: sender@example.com\r\n"+
				"Subject: Delivery Status Notification\r\n"+
				"Content-Type: multipart/report; report-type=delivery-status; boundary=%s\r\n"+
				"\r\n"+
				"--%s\r\n"+
				"Content-Type: text/plain\r\n"+
				"\r\n"+
				"Delivery delayed.\r\n"+
				"--%s\r\n"+
				"Content-Type: message/delivery-status\r\n"+
				"\r\n"+
				"Reporting-MTA: dns; example.com\r\n"+
				"\r\n"+
				"Final-Recipient: rfc822;alice@example.com\r\n"+
				"Action: delayed\r\n"+
				"Status: 4.2.2\r\n"+
				"--%s--\r\n",
			boundary, boundary, boundary, boundary,
		)

		info, err := ClassifyDSN([]byte(rawMessage))
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "alice@example.com", info.Recipient)
		assert.Equal(t, BounceSoft, info.Type)
		assert.False(t, info.Permanent)
	})

	t.Run("missing Content-Type header", func(t *testing.T) {
		rawMessage := "From: test@example.com\r\n\r\nNo content type\r\n"
		_, err := ClassifyDSN([]byte(rawMessage))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Content-Type")
	})

	t.Run("wrong Content-Type (not multipart/report)", func(t *testing.T) {
		rawMessage := "From: test@example.com\r\nContent-Type: text/plain\r\n\r\nNot a DSN.\r\n"
		_, err := ClassifyDSN([]byte(rawMessage))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multipart/report")
	})

	t.Run("wrong report-type", func(t *testing.T) {
		rawMessage := "From: test@example.com\r\n" +
			"Content-Type: multipart/report; report-type=feedback-report; boundary=b1\r\n" +
			"\r\n" +
			"--b1\r\n" +
			"Content-Type: text/plain\r\n" +
			"\r\n" +
			"feedback.\r\n" +
			"--b1--\r\n"
		_, err := ClassifyDSN([]byte(rawMessage))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delivery-status")
	})

	t.Run("no delivery-status part found", func(t *testing.T) {
		boundary := "nobouncepart"
		rawMessage := fmt.Sprintf(
			"From: test@example.com\r\n"+
				"Content-Type: multipart/report; report-type=delivery-status; boundary=%s\r\n"+
				"\r\n"+
				"--%s\r\n"+
				"Content-Type: text/plain\r\n"+
				"\r\n"+
				"Some text.\r\n"+
				"--%s--\r\n",
			boundary, boundary, boundary,
		)
		_, err := ClassifyDSN([]byte(rawMessage))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no message/delivery-status part found")
	})

	t.Run("invalid raw message", func(t *testing.T) {
		_, err := ClassifyDSN([]byte("not a valid email at all"))
		// Should fail either on parsing or missing Content-Type
		assert.Error(t, err)
	})
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		substrs []string
		want    bool
	}{
		{
			name:    "contains first substring",
			s:       "hello world",
			substrs: []string{"hello", "foo"},
			want:    true,
		},
		{
			name:    "contains second substring",
			s:       "hello world",
			substrs: []string{"foo", "world"},
			want:    true,
		},
		{
			name:    "contains none",
			s:       "hello world",
			substrs: []string{"foo", "bar", "baz"},
			want:    false,
		},
		{
			name:    "empty string",
			s:       "",
			substrs: []string{"foo"},
			want:    false,
		},
		{
			name:    "empty substrs",
			s:       "hello",
			substrs: []string{},
			want:    false,
		},
		{
			name:    "empty substring matches everything",
			s:       "hello",
			substrs: []string{""},
			want:    true,
		},
		{
			name:    "case sensitive",
			s:       "Hello World",
			substrs: []string{"hello"},
			want:    false,
		},
		{
			name:    "partial match",
			s:       "unsubscribe",
			substrs: []string{"subscribe"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.s, tt.substrs...)
			assert.Equal(t, tt.want, got)
		})
	}
}
