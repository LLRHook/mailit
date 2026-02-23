package smtp

import (
	"crypto/tls"
	"log/slog"
	"time"

	gosmtp "github.com/emersion/go-smtp"
)

// ServerConfig holds the configuration for the inbound SMTP server.
type ServerConfig struct {
	ListenAddr      string
	Domain          string
	MaxMessageBytes int64
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	TLSCert         string // path to TLS certificate (optional)
	TLSKey          string // path to TLS private key (optional)
}

// NewServer creates a new inbound SMTP server backed by the given Backend.
func NewServer(cfg ServerConfig, backend *Backend, logger *slog.Logger) *gosmtp.Server {
	s := gosmtp.NewServer(backend)

	s.Addr = cfg.ListenAddr
	s.Domain = cfg.Domain
	s.MaxMessageBytes = cfg.MaxMessageBytes
	s.ReadTimeout = cfg.ReadTimeout
	s.WriteTimeout = cfg.WriteTimeout
	s.AllowInsecureAuth = true // No auth required for inbound mail

	// Configure optional TLS (STARTTLS support for inbound).
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			logger.Error("failed to load TLS certificate for inbound SMTP",
				"cert", cfg.TLSCert,
				"key", cfg.TLSKey,
				"error", err,
			)
		} else {
			s.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			logger.Info("TLS enabled for inbound SMTP server")
		}
	}

	return s
}
