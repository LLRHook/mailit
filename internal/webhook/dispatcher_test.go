package webhook

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	t.Run("produces consistent HMAC for same inputs", func(t *testing.T) {
		payload := []byte(`{"type":"email.sent","data":{"id":"123"}}`)
		secret := "whsec_test_secret_key"
		timestamp := int64(1700000000)

		sig1 := Sign(payload, secret, timestamp)
		sig2 := Sign(payload, secret, timestamp)

		assert.Equal(t, sig1, sig2, "same inputs should produce same signature")
		assert.Len(t, sig1, 64, "HMAC-SHA256 hex digest should be 64 chars")
	})

	t.Run("different payload produces different signature", func(t *testing.T) {
		secret := "whsec_test"
		timestamp := int64(1700000000)

		sig1 := Sign([]byte(`{"a":"1"}`), secret, timestamp)
		sig2 := Sign([]byte(`{"a":"2"}`), secret, timestamp)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("different secret produces different signature", func(t *testing.T) {
		payload := []byte(`{"data":"test"}`)
		timestamp := int64(1700000000)

		sig1 := Sign(payload, "secret1", timestamp)
		sig2 := Sign(payload, "secret2", timestamp)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("different timestamp produces different signature", func(t *testing.T) {
		payload := []byte(`{"data":"test"}`)
		secret := "whsec_test"

		sig1 := Sign(payload, secret, 1000)
		sig2 := Sign(payload, secret, 2000)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("empty payload produces valid signature", func(t *testing.T) {
		sig := Sign([]byte{}, "secret", 1000)
		assert.Len(t, sig, 64)
	})
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"type":"email.delivered","data":{"id":"abc-123"}}`)
	secret := "whsec_verification_test"
	timestamp := int64(1700000000)

	t.Run("valid signature returns true", func(t *testing.T) {
		sig := Sign(payload, secret, timestamp)
		assert.True(t, VerifySignature(payload, secret, timestamp, sig))
	})

	t.Run("wrong signature returns false", func(t *testing.T) {
		assert.False(t, VerifySignature(payload, secret, timestamp, "invalid_signature_value"))
	})

	t.Run("wrong secret returns false", func(t *testing.T) {
		sig := Sign(payload, secret, timestamp)
		assert.False(t, VerifySignature(payload, "wrong_secret", timestamp, sig))
	})

	t.Run("wrong timestamp returns false", func(t *testing.T) {
		sig := Sign(payload, secret, timestamp)
		assert.False(t, VerifySignature(payload, secret, timestamp+1, sig))
	})

	t.Run("tampered payload returns false", func(t *testing.T) {
		sig := Sign(payload, secret, timestamp)
		tampered := []byte(`{"type":"email.delivered","data":{"id":"xyz-789"}}`)
		assert.False(t, VerifySignature(tampered, secret, timestamp, sig))
	})

	t.Run("empty payload with matching signature", func(t *testing.T) {
		emptyPayload := []byte{}
		sig := Sign(emptyPayload, secret, timestamp)
		assert.True(t, VerifySignature(emptyPayload, secret, timestamp, sig))
	})
}

func TestSubscribesToEvent(t *testing.T) {
	tests := []struct {
		name      string
		events    []string
		eventType string
		want      bool
	}{
		{
			name:      "wildcard matches any event",
			events:    []string{"*"},
			eventType: "email.sent",
			want:      true,
		},
		{
			name:      "wildcard matches another event",
			events:    []string{"*"},
			eventType: "email.bounced",
			want:      true,
		},
		{
			name:      "exact match",
			events:    []string{"email.sent", "email.delivered"},
			eventType: "email.sent",
			want:      true,
		},
		{
			name:      "exact match second event",
			events:    []string{"email.sent", "email.delivered"},
			eventType: "email.delivered",
			want:      true,
		},
		{
			name:      "no match",
			events:    []string{"email.sent", "email.delivered"},
			eventType: "email.bounced",
			want:      false,
		},
		{
			name:      "empty events list",
			events:    []string{},
			eventType: "email.sent",
			want:      false,
		},
		{
			name:      "nil events list",
			events:    nil,
			eventType: "email.sent",
			want:      false,
		},
		{
			name:      "wildcard among specific events",
			events:    []string{"email.sent", "*"},
			eventType: "email.bounced",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subscribesToEvent(tt.events, tt.eventType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateRetryTime(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "attempt 1 - 30 seconds",
			attempt:  1,
			expected: 30 * time.Second,
		},
		{
			name:     "attempt 2 - 2 minutes",
			attempt:  2,
			expected: 2 * time.Minute,
		},
		{
			name:     "attempt 3 - 10 minutes",
			attempt:  3,
			expected: 10 * time.Minute,
		},
		{
			name:     "attempt 4 - 30 minutes",
			attempt:  4,
			expected: 30 * time.Minute,
		},
		{
			name:     "attempt 5 - 2 hours",
			attempt:  5,
			expected: 2 * time.Hour,
		},
		{
			name:     "attempt 0 - uses first backoff (30s)",
			attempt:  0,
			expected: 30 * time.Second,
		},
		{
			name:     "negative attempt - uses first backoff",
			attempt:  -1,
			expected: 30 * time.Second,
		},
		{
			name:     "attempt beyond max - uses last backoff (2h)",
			attempt:  10,
			expected: 2 * time.Hour,
		},
		{
			name:     "attempt 100 - still uses last backoff (2h)",
			attempt:  100,
			expected: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRetryTime(now, tt.attempt)
			expected := now.Add(tt.expected)
			assert.Equal(t, expected, result)
		})
	}
}

func TestToJSONMap(t *testing.T) {
	t.Run("struct to JSONMap", func(t *testing.T) {
		input := struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Count int    `json:"count"`
		}{
			Name:  "Alice",
			Email: "alice@example.com",
			Count: 42,
		}

		m, err := toJSONMap(input)
		require.NoError(t, err)
		assert.Equal(t, "Alice", m["name"])
		assert.Equal(t, "alice@example.com", m["email"])
		// JSON numbers become float64 through marshal/unmarshal cycle.
		assert.Equal(t, float64(42), m["count"])
	})

	t.Run("map to JSONMap", func(t *testing.T) {
		input := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		m, err := toJSONMap(input)
		require.NoError(t, err)
		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, float64(123), m["key2"])
	})

	t.Run("nested struct to JSONMap", func(t *testing.T) {
		input := struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}{
			Data: struct {
				ID string `json:"id"`
			}{ID: "abc-123"},
		}

		m, err := toJSONMap(input)
		require.NoError(t, err)
		nested, ok := m["data"].(map[string]interface{})
		require.True(t, ok, "nested should be a map")
		assert.Equal(t, "abc-123", nested["id"])
	})

	t.Run("unmarshalable input returns error", func(t *testing.T) {
		// Channels cannot be marshaled to JSON.
		ch := make(chan int)
		_, err := toJSONMap(ch)
		assert.Error(t, err)
	})

	t.Run("nil input produces nil JSONMap without error", func(t *testing.T) {
		// json.Marshal(nil) produces "null", json.Unmarshal("null", &m) sets m to nil.
		m, err := toJSONMap(nil)
		assert.NoError(t, err)
		assert.Nil(t, m)
	})

	t.Run("string input returns error", func(t *testing.T) {
		// A bare string like "hello" marshals to `"hello"`, which cannot unmarshal to JSONMap.
		_, err := toJSONMap("hello")
		assert.Error(t, err)
	})
}
