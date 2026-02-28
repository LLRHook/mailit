package config

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Validate checks the configuration for required fields and invalid values.
// It collects all failures into a single error so the operator sees every
// problem at once.
func (c *Config) Validate() error {
	var errs []string

	// Auth
	if c.Auth.JWTSecret == "" {
		errs = append(errs, "auth.jwt_secret is required")
	} else if len(c.Auth.JWTSecret) < 32 {
		errs = append(errs, "auth.jwt_secret must be at least 32 characters")
	}

	// Database
	if c.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}
	if c.Database.Password == "" {
		errs = append(errs, "database.password is required")
	}
	if c.Database.DBName == "" {
		errs = append(errs, "database.dbname is required")
	}

	// Redis
	if c.Redis.Addr == "" {
		errs = append(errs, "redis.addr is required")
	}

	// SMTP Outbound
	if c.SMTPOutbound.Hostname == "" {
		errs = append(errs, "smtp_outbound.hostname is required")
	}

	// DKIM master encryption key (optional, but validated if set)
	if c.DKIM.MasterEncryptionKey != "" {
		decoded, err := hex.DecodeString(c.DKIM.MasterEncryptionKey)
		if err != nil {
			errs = append(errs, "dkim.master_encryption_key must be valid hex")
		} else if len(decoded) < 32 {
			errs = append(errs, "dkim.master_encryption_key must be at least 32 bytes (64 hex chars)")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
