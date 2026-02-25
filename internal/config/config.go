package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds the complete application configuration.
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Auth         AuthConfig         `mapstructure:"auth"`
	SMTPOutbound SMTPOutboundConfig `mapstructure:"smtp_outbound"`
	SMTPInbound  SMTPInboundConfig  `mapstructure:"smtp_inbound"`
	DKIM         DKIMConfig         `mapstructure:"dkim"`
	Workers      WorkersConfig      `mapstructure:"workers"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Webhooks     WebhooksConfig     `mapstructure:"webhooks"`
	DNS          DNSConfig          `mapstructure:"dns"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Storage      StorageConfig      `mapstructure:"storage"`
	Suppression  SuppressionConfig  `mapstructure:"suppression"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	HTTPAddr        string        `mapstructure:"http_addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	CORSOrigins     []string      `mapstructure:"cors_origins"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
}

// DSN returns a PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret    string        `mapstructure:"jwt_secret"`
	JWTExpiry    time.Duration `mapstructure:"jwt_expiry"`
	APIKeyPrefix string        `mapstructure:"api_key_prefix"`
	BcryptCost   int           `mapstructure:"bcrypt_cost"`
}

// SMTPOutboundConfig holds outbound SMTP delivery settings.
type SMTPOutboundConfig struct {
	Hostname       string        `mapstructure:"hostname"`
	Port           int           `mapstructure:"port"`
	HELODomain     string        `mapstructure:"helo_domain"`
	TLSPolicy      string        `mapstructure:"tls_policy"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	SendTimeout    time.Duration `mapstructure:"send_timeout"`
	MaxRecipients  int           `mapstructure:"max_recipients"`
}

// SMTPInboundConfig holds inbound SMTP server settings.
type SMTPInboundConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ListenAddr      string        `mapstructure:"listen_addr"`
	Domain          string        `mapstructure:"domain"`
	MaxMessageBytes int           `mapstructure:"max_message_bytes"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}

// DKIMConfig holds DKIM signing settings.
type DKIMConfig struct {
	Selector            string `mapstructure:"selector"`
	KeyBits             int    `mapstructure:"key_bits"`
	MasterEncryptionKey string `mapstructure:"master_encryption_key"`
}

// WorkersConfig holds background worker settings.
type WorkersConfig struct {
	Concurrency int            `mapstructure:"concurrency"`
	Queues      map[string]int `mapstructure:"queues"`
	RetryDelays []string       `mapstructure:"retry_delays"`
}

// ParseRetryDelays parses the string retry delays into time.Duration values.
func (w WorkersConfig) ParseRetryDelays() ([]time.Duration, error) {
	delays := make([]time.Duration, 0, len(w.RetryDelays))
	for _, s := range w.RetryDelays {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("invalid worker retry delay %q: %w", s, err)
		}
		delays = append(delays, d)
	}
	return delays, nil
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	Enabled    bool          `mapstructure:"enabled"`
	DefaultRPS int           `mapstructure:"default_rps"`
	SendRPS    int           `mapstructure:"send_rps"`
	BatchRPS   int           `mapstructure:"batch_rps"`
	Window     time.Duration `mapstructure:"window"`
}

// WebhooksConfig holds webhook delivery settings.
type WebhooksConfig struct {
	SigningAlgorithm string   `mapstructure:"signing_algorithm"`
	Timeout          time.Duration `mapstructure:"timeout"`
	MaxRetries       int      `mapstructure:"max_retries"`
	RetryDelays      []string `mapstructure:"retry_delays"`
}

// ParseRetryDelays parses the string retry delays into time.Duration values.
func (w WebhooksConfig) ParseRetryDelays() ([]time.Duration, error) {
	delays := make([]time.Duration, 0, len(w.RetryDelays))
	for _, s := range w.RetryDelays {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("invalid webhook retry delay %q: %w", s, err)
		}
		delays = append(delays, d)
	}
	return delays, nil
}

// DNSConfig holds DNS resolution settings.
type DNSConfig struct {
	Resolver string        `mapstructure:"resolver"`
	Timeout  time.Duration `mapstructure:"timeout"`
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// StorageConfig holds attachment storage settings.
type StorageConfig struct {
	Type      string   `mapstructure:"type"`
	LocalPath string   `mapstructure:"local_path"`
	S3        S3Config `mapstructure:"s3"`
}

// S3Config holds S3-compatible storage settings.
type S3Config struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// SuppressionConfig holds suppression list settings.
type SuppressionConfig struct {
	AutoAddHardBounces bool `mapstructure:"auto_add_hard_bounces"`
	AutoAddComplaints  bool `mapstructure:"auto_add_complaints"`
}

// defaults returns the default configuration as a flat map using koanf's "."
// delimiter for nested keys.
func defaults() map[string]interface{} {
	return map[string]interface{}{
		// Server
		"server.http_addr":        ":8080",
		"server.read_timeout":     "30s",
		"server.write_timeout":    "30s",
		"server.shutdown_timeout": "10s",
		"server.cors_origins":    []string{"http://localhost:3000", "http://localhost:3001"},

		// Database
		"database.host":              "localhost",
		"database.port":              5432,
		"database.user":              "mailit",
		"database.password":          "",
		"database.dbname":            "mailit",
		"database.sslmode":           "disable",
		"database.max_open_conns":    25,
		"database.max_idle_conns":    5,
		"database.conn_max_lifetime": "5m",
		"database.auto_migrate":      true,

		// Redis
		"redis.addr":      "localhost:6379",
		"redis.password":  "",
		"redis.db":        0,
		"redis.pool_size": 10,

		// Auth
		"auth.jwt_secret":     "",
		"auth.jwt_expiry":     "24h",
		"auth.api_key_prefix": "re_",
		"auth.bcrypt_cost":    12,

		// SMTP Outbound
		"smtp_outbound.hostname":        "",
		"smtp_outbound.port":            25,
		"smtp_outbound.helo_domain":     "",
		"smtp_outbound.tls_policy":      "opportunistic",
		"smtp_outbound.connect_timeout": "30s",
		"smtp_outbound.send_timeout":    "5m",
		"smtp_outbound.max_recipients":  50,

		// SMTP Inbound
		"smtp_inbound.enabled":           true,
		"smtp_inbound.listen_addr":       ":25",
		"smtp_inbound.domain":            "",
		"smtp_inbound.max_message_bytes": 26214400,
		"smtp_inbound.read_timeout":      "60s",
		"smtp_inbound.write_timeout":     "60s",

		// DKIM
		"dkim.selector":              "mailit",
		"dkim.key_bits":              2048,
		"dkim.master_encryption_key": "",

		// Workers
		"workers.concurrency": 20,

		// Rate Limit
		"rate_limit.enabled":     true,
		"rate_limit.default_rps": 10,
		"rate_limit.send_rps":    2,
		"rate_limit.batch_rps":   1,
		"rate_limit.window":      "1s",

		// Webhooks
		"webhooks.signing_algorithm": "hmac-sha256",
		"webhooks.timeout":           "30s",
		"webhooks.max_retries":       5,

		// DNS
		"dns.resolver":  "system",
		"dns.timeout":   "10s",
		"dns.cache_ttl": "5m",

		// Logging
		"logging.level":  "info",
		"logging.format": "json",
		"logging.output": "stdout",

		// Storage
		"storage.type":       "local",
		"storage.local_path": "./data/attachments",

		// Suppression
		"suppression.auto_add_hard_bounces": true,
		"suppression.auto_add_complaints":   true,
	}
}

// Load reads the configuration from defaults, an optional YAML file, and
// environment variables (prefix MAILIT_). Later sources override earlier ones.
func Load(path string) (*Config, error) {
	k := koanf.New(".")

	// 1. Load defaults.
	if err := k.Load(confmap.Provider(defaults(), "."), nil); err != nil {
		return nil, fmt.Errorf("loading defaults: %w", err)
	}

	// 2. Load YAML file if provided and exists.
	if path != "" {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("loading config file %s: %w", path, err)
		}
	}

	// 3. Overlay environment variables.
	//    MAILIT_SERVER_HTTP_ADDR -> server.http_addr
	if err := k.Load(env.Provider("MAILIT_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "MAILIT_")),
			"_", ".",
		)
	}), nil); err != nil {
		return nil, fmt.Errorf("loading env variables: %w", err)
	}

	// 4. Unmarshal into the Config struct.
	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		Tag: "mapstructure",
	}); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
