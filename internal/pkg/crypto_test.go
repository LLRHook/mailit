package pkg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAPIKey(t *testing.T) {
	t.Run("returns correct prefix", func(t *testing.T) {
		plaintext, _, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(plaintext, "re_"), "key should start with 're_'")
	})

	t.Run("custom prefix", func(t *testing.T) {
		plaintext, _, _, err := GenerateAPIKey("sk_live_")
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(plaintext, "sk_live_"), "key should start with 'sk_live_'")
	})

	t.Run("key length is prefix + 64 hex chars", func(t *testing.T) {
		prefix := "re_"
		plaintext, _, _, err := GenerateAPIKey(prefix)
		require.NoError(t, err)
		// 32 random bytes -> 64 hex chars + prefix length
		assert.Equal(t, len(prefix)+64, len(plaintext))
	})

	t.Run("hash is deterministic for same input", func(t *testing.T) {
		plaintext, hash1, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)

		hash2 := HashAPIKey(plaintext)
		assert.Equal(t, hash1, hash2, "hash from GenerateAPIKey should match HashAPIKey")
	})

	t.Run("hash is 64-char hex string (SHA-256)", func(t *testing.T) {
		_, hash, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		assert.Len(t, hash, 64, "SHA-256 hash should be 64 hex chars")
	})

	t.Run("key prefix is truncated correctly", func(t *testing.T) {
		prefix := "re_"
		plaintext, _, keyPrefix, err := GenerateAPIKey(prefix)
		require.NoError(t, err)

		// keyPrefix should be prefix + first 8 chars of random part + "..."
		expectedPrefix := plaintext[:len(prefix)+8] + "..."
		assert.Equal(t, expectedPrefix, keyPrefix)
		assert.True(t, strings.HasSuffix(keyPrefix, "..."))
	})

	t.Run("each call generates unique key", func(t *testing.T) {
		key1, _, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		key2, _, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		assert.NotEqual(t, key1, key2, "each call should generate a unique key")
	})

	t.Run("each call generates unique hash", func(t *testing.T) {
		_, hash1, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		_, hash2, _, err := GenerateAPIKey("re_")
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash2, "each call should generate a unique hash")
	})
}

func TestHashAPIKey(t *testing.T) {
	t.Run("consistent hashing", func(t *testing.T) {
		key := "re_abc123def456"
		h1 := HashAPIKey(key)
		h2 := HashAPIKey(key)
		assert.Equal(t, h1, h2)
	})

	t.Run("different keys produce different hashes", func(t *testing.T) {
		h1 := HashAPIKey("key1")
		h2 := HashAPIKey("key2")
		assert.NotEqual(t, h1, h2)
	})

	t.Run("hash length is 64 hex chars", func(t *testing.T) {
		h := HashAPIKey("any-key-value")
		assert.Len(t, h, 64)
	})

	t.Run("empty string hashes consistently", func(t *testing.T) {
		h1 := HashAPIKey("")
		h2 := HashAPIKey("")
		assert.Equal(t, h1, h2)
		assert.Len(t, h1, 64)
	})

	t.Run("hash is lowercase hex", func(t *testing.T) {
		h := HashAPIKey("test")
		for _, c := range h {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"expected lowercase hex char, got: %c", c)
		}
	})
}

func TestGenerateWebhookSecret(t *testing.T) {
	t.Run("returns 64-char hex string", func(t *testing.T) {
		secret, err := GenerateWebhookSecret()
		require.NoError(t, err)
		assert.Len(t, secret, 64, "32 bytes -> 64 hex chars")
	})

	t.Run("is lowercase hex", func(t *testing.T) {
		secret, err := GenerateWebhookSecret()
		require.NoError(t, err)
		for _, c := range secret {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"expected lowercase hex char, got: %c", c)
		}
	})

	t.Run("each call generates unique secret", func(t *testing.T) {
		s1, err := GenerateWebhookSecret()
		require.NoError(t, err)
		s2, err := GenerateWebhookSecret()
		require.NoError(t, err)
		assert.NotEqual(t, s1, s2)
	})
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("returns string of correct length", func(t *testing.T) {
		s, err := GenerateRandomString(16)
		require.NoError(t, err)
		assert.Len(t, s, 32, "16 bytes -> 32 hex chars")
	})

	t.Run("different lengths produce different sized outputs", func(t *testing.T) {
		s8, err := GenerateRandomString(8)
		require.NoError(t, err)
		assert.Len(t, s8, 16)

		s32, err := GenerateRandomString(32)
		require.NoError(t, err)
		assert.Len(t, s32, 64)
	})

	t.Run("zero length returns empty string", func(t *testing.T) {
		s, err := GenerateRandomString(0)
		require.NoError(t, err)
		assert.Equal(t, "", s)
	})

	t.Run("each call generates unique output", func(t *testing.T) {
		s1, err := GenerateRandomString(16)
		require.NoError(t, err)
		s2, err := GenerateRandomString(16)
		require.NoError(t, err)
		assert.NotEqual(t, s1, s2)
	})

	t.Run("output is valid hex", func(t *testing.T) {
		s, err := GenerateRandomString(32)
		require.NoError(t, err)
		for _, c := range s {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"expected hex char, got: %c", c)
		}
	})
}
