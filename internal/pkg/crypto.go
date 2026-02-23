package pkg

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey generates a new API key with the given prefix (e.g., "re_").
// It returns the plaintext key, its SHA-256 hash, and a truncated prefix for display.
func GenerateAPIKey(prefix string) (plaintext string, hash string, keyPrefix string, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", "", fmt.Errorf("generating random bytes: %w", err)
	}

	plaintext = prefix + hex.EncodeToString(bytes)
	hash = HashAPIKey(plaintext)
	keyPrefix = plaintext[:len(prefix)+8] + "..."
	return plaintext, hash, keyPrefix, nil
}

// HashAPIKey creates a SHA-256 hash of an API key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// GenerateRandomString generates a cryptographically secure random hex string.
// The returned string will be 2*length characters long.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateWebhookSecret generates a signing secret for webhooks.
func GenerateWebhookSecret() (string, error) {
	return GenerateRandomString(32)
}
