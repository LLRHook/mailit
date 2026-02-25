package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any MAILIT_ environment variables that could interfere.
	for _, env := range os.Environ() {
		if len(env) > 7 && env[:7] == "MAILIT_" {
			if idx := strings.IndexByte(env, '='); idx > 0 {
				key := env[:idx]
				t.Setenv(key, os.Getenv(key)) // register for cleanup
				_ = os.Unsetenv(key)
			}
		}
	}

	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Server defaults.
	assert.Equal(t, ":8080", cfg.Server.HTTPAddr)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:3001"}, cfg.Server.CORSOrigins)

	// Database defaults.
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "mailit", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "mailit", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.True(t, cfg.Database.AutoMigrate)

	// Redis defaults.
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
	assert.Equal(t, 10, cfg.Redis.PoolSize)

	// Auth defaults.
	assert.Equal(t, "", cfg.Auth.JWTSecret)
	assert.Equal(t, "re_", cfg.Auth.APIKeyPrefix)
	assert.Equal(t, 12, cfg.Auth.BcryptCost)

	// SMTP Outbound defaults.
	assert.Equal(t, 25, cfg.SMTPOutbound.Port)
	assert.Equal(t, "opportunistic", cfg.SMTPOutbound.TLSPolicy)
	assert.Equal(t, 50, cfg.SMTPOutbound.MaxRecipients)

	// SMTP Inbound defaults.
	assert.True(t, cfg.SMTPInbound.Enabled)
	assert.Equal(t, ":25", cfg.SMTPInbound.ListenAddr)
	assert.Equal(t, 26214400, cfg.SMTPInbound.MaxMessageBytes)

	// DKIM defaults.
	assert.Equal(t, "mailit", cfg.DKIM.Selector)
	assert.Equal(t, 2048, cfg.DKIM.KeyBits)

	// Workers defaults.
	assert.Equal(t, 20, cfg.Workers.Concurrency)

	// Rate limit defaults.
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 10, cfg.RateLimit.DefaultRPS)
	assert.Equal(t, 2, cfg.RateLimit.SendRPS)
	assert.Equal(t, 1, cfg.RateLimit.BatchRPS)

	// Webhooks defaults.
	assert.Equal(t, "hmac-sha256", cfg.Webhooks.SigningAlgorithm)
	assert.Equal(t, 5, cfg.Webhooks.MaxRetries)

	// DNS defaults.
	assert.Equal(t, "system", cfg.DNS.Resolver)

	// Logging defaults.
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)

	// Storage defaults.
	assert.Equal(t, "local", cfg.Storage.Type)
	assert.Equal(t, "./data/attachments", cfg.Storage.LocalPath)

	// Suppression defaults.
	assert.True(t, cfg.Suppression.AutoAddHardBounces)
	assert.True(t, cfg.Suppression.AutoAddComplaints)
}

func TestLoad_EnvOverrides(t *testing.T) {
	// The env transformer replaces ALL underscores with dots, so
	// MAILIT_DATABASE_HOST -> database.host (works because each segment is one word).
	// Multi-word koanf keys like "http_addr" cannot be targeted with a single
	// underscore because it becomes a dot separator. Only test keys whose
	// segments are single words.
	t.Setenv("MAILIT_DATABASE_HOST", "db.example.com")
	t.Setenv("MAILIT_LOGGING_LEVEL", "debug")
	t.Setenv("MAILIT_DKIM_SELECTOR", "custom")

	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "custom", cfg.DKIM.Selector)

	// Verify defaults are still set for keys we didn't override.
	assert.Equal(t, ":8080", cfg.Server.HTTPAddr)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "re_", cfg.Auth.APIKeyPrefix)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loading config file")
}

func TestDatabaseConfig_DSN(t *testing.T) {
	db := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "mailit",
		Password: "secret",
		DBName:   "mailit_db",
		SSLMode:  "require",
	}

	dsn := db.DSN()
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=mailit")
	assert.Contains(t, dsn, "password=secret")
	assert.Contains(t, dsn, "dbname=mailit_db")
	assert.Contains(t, dsn, "sslmode=require")
}

func TestWorkersConfig_ParseRetryDelays(t *testing.T) {
	t.Run("valid delays", func(t *testing.T) {
		w := WorkersConfig{
			RetryDelays: []string{"30s", "1m", "5m", "30m"},
		}
		delays, err := w.ParseRetryDelays()
		require.NoError(t, err)
		require.Len(t, delays, 4)
	})

	t.Run("invalid delay", func(t *testing.T) {
		w := WorkersConfig{
			RetryDelays: []string{"30s", "invalid"},
		}
		_, err := w.ParseRetryDelays()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("empty delays", func(t *testing.T) {
		w := WorkersConfig{
			RetryDelays: []string{},
		}
		delays, err := w.ParseRetryDelays()
		require.NoError(t, err)
		assert.Empty(t, delays)
	})
}

func TestWebhooksConfig_ParseRetryDelays(t *testing.T) {
	t.Run("valid delays", func(t *testing.T) {
		w := WebhooksConfig{
			RetryDelays: []string{"10s", "1m", "10m"},
		}
		delays, err := w.ParseRetryDelays()
		require.NoError(t, err)
		require.Len(t, delays, 3)
	})

	t.Run("invalid delay", func(t *testing.T) {
		w := WebhooksConfig{
			RetryDelays: []string{"not-a-duration"},
		}
		_, err := w.ParseRetryDelays()
		assert.Error(t, err)
	})
}

