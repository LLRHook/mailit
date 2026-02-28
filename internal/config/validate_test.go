package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validConfig returns a Config that passes all validation checks.
func validConfig() *Config {
	return &Config{
		Auth: AuthConfig{
			JWTSecret: "this-is-a-secret-that-is-at-least-32-chars-long!!",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Password: "secret",
			DBName:   "mailit",
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
		},
		SMTPOutbound: SMTPOutboundConfig{
			Hostname: "mail.example.com",
		},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.Auth.JWTSecret = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth.jwt_secret is required")
}

func TestValidate_ShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.Auth.JWTSecret = "short"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth.jwt_secret must be at least 32 characters")
}

func TestValidate_MissingDatabaseHost(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Host = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.host is required")
}

func TestValidate_MissingDatabasePassword(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Password = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.password is required")
}

func TestValidate_MissingDatabaseName(t *testing.T) {
	cfg := validConfig()
	cfg.Database.DBName = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.dbname is required")
}

func TestValidate_MissingRedisAddr(t *testing.T) {
	cfg := validConfig()
	cfg.Redis.Addr = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis.addr is required")
}

func TestValidate_MissingSMTPHostname(t *testing.T) {
	cfg := validConfig()
	cfg.SMTPOutbound.Hostname = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smtp_outbound.hostname is required")
}

func TestValidate_InvalidDKIMHex(t *testing.T) {
	cfg := validConfig()
	cfg.DKIM.MasterEncryptionKey = "not-valid-hex"
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dkim.master_encryption_key must be valid hex")
}

func TestValidate_ShortDKIMKey(t *testing.T) {
	cfg := validConfig()
	cfg.DKIM.MasterEncryptionKey = "0123456789abcdef" // 8 bytes, need 32
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dkim.master_encryption_key must be at least 32 bytes")
}

func TestValidate_ValidDKIMKey(t *testing.T) {
	cfg := validConfig()
	cfg.DKIM.MasterEncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{} // All required fields missing
	err := cfg.Validate()
	require.Error(t, err)

	msg := err.Error()
	// Should report all missing fields at once.
	assert.Contains(t, msg, "auth.jwt_secret is required")
	assert.Contains(t, msg, "database.host is required")
	assert.Contains(t, msg, "database.password is required")
	assert.Contains(t, msg, "database.dbname is required")
	assert.Contains(t, msg, "redis.addr is required")
	assert.Contains(t, msg, "smtp_outbound.hostname is required")

	// All 6 errors present.
	assert.Equal(t, 6, strings.Count(msg, "\n  - "))
}
