package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"

	"github.com/mailit-dev/mailit/internal/handler"
	"github.com/mailit-dev/mailit/internal/server/middleware"
)

type Config struct {
	Addr           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	JWTSecret      string
	APIKeyPrefix   string
	CORSOrigins    []string
	RateLimitCfg   middleware.RateLimitConfig
	Redis          *redis.Client
	APIKeyLookup   middleware.APIKeyLookup
	APIKeyLastUsed middleware.APIKeyLastUsedUpdate
	Handlers       *handler.Handlers
	Logger         *slog.Logger
}

func New(cfg Config) *http.Server {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (no auth)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth middleware
	authMw := middleware.Auth(cfg.JWTSecret, cfg.APIKeyPrefix, cfg.APIKeyLookup, cfg.APIKeyLastUsed)
	rateLimitMw := middleware.RateLimit(cfg.Redis, cfg.RateLimitCfg)
	sendLimitMw := middleware.SendRateLimit(cfg.Redis, cfg.RateLimitCfg)
	batchLimitMw := middleware.BatchRateLimit(cfg.Redis, cfg.RateLimitCfg)

	// IP-based rate limits for public auth endpoints.
	registerLimitMw := middleware.IPRateLimit(cfg.Redis, 5, time.Minute)
	loginLimitMw := middleware.IPRateLimit(cfg.Redis, 10, time.Minute)

	h := cfg.Handlers

	// Public routes (auth) with stricter IP-based rate limits.
	r.With(registerLimitMw).Post("/auth/register", h.Auth.Register)
	r.With(loginLimitMw).Post("/auth/login", h.Auth.Login)
	r.With(loginLimitMw).Post("/auth/accept-invite", h.Settings.AcceptInvite)

	// Public tracking routes (no auth)
	r.Get("/track/open/{id}", h.Tracking.TrackOpen)
	r.Get("/track/click/{id}", h.Tracking.TrackClick)
	r.Post("/unsubscribe", h.Tracking.Unsubscribe)

	// Authenticated API routes
	r.Group(func(r chi.Router) {
		r.Use(authMw)
		r.Use(rateLimitMw)

		// Emails
		r.With(sendLimitMw).Post("/emails", h.Email.Send)
		r.With(batchLimitMw).Post("/emails/batch", h.Email.BatchSend)
		r.Get("/emails", h.Email.List)
		r.Get("/emails/{emailId}", h.Email.Get)
		r.Patch("/emails/{emailId}", h.Email.Update)
		r.Post("/emails/{emailId}/cancel", h.Email.Cancel)

		// Domains
		r.Post("/domains", h.Domain.Create)
		r.Get("/domains", h.Domain.List)
		r.Get("/domains/{domainId}", h.Domain.Get)
		r.Patch("/domains/{domainId}", h.Domain.Update)
		r.Delete("/domains/{domainId}", h.Domain.Delete)
		r.Post("/domains/{domainId}/verify", h.Domain.Verify)

		// API Keys
		r.Post("/api-keys", h.APIKey.Create)
		r.Get("/api-keys", h.APIKey.List)
		r.Delete("/api-keys/{apiKeyId}", h.APIKey.Delete)

		// Audiences
		r.Post("/audiences", h.Audience.Create)
		r.Get("/audiences", h.Audience.List)
		r.Get("/audiences/{audienceId}", h.Audience.Get)
		r.Delete("/audiences/{audienceId}", h.Audience.Delete)

		// Contacts
		r.Post("/audiences/{audienceId}/contacts", h.Contact.Create)
		r.Get("/audiences/{audienceId}/contacts", h.Contact.List)
		r.Get("/audiences/{audienceId}/contacts/export", h.Contact.Export)
		r.Post("/audiences/{audienceId}/contacts/import", h.ContactImport.Import)
		r.Get("/audiences/{audienceId}/contacts/import/{jobId}", h.ContactImport.GetImportStatus)
		r.Get("/audiences/{audienceId}/contacts/{contactId}", h.Contact.Get)
		r.Patch("/audiences/{audienceId}/contacts/{contactId}", h.Contact.Update)
		r.Delete("/audiences/{audienceId}/contacts/{contactId}", h.Contact.Delete)

		// Contact Properties
		r.Post("/contact-properties", h.ContactProperty.Create)
		r.Get("/contact-properties", h.ContactProperty.List)
		r.Patch("/contact-properties/{propertyId}", h.ContactProperty.Update)
		r.Delete("/contact-properties/{propertyId}", h.ContactProperty.Delete)

		// Topics
		r.Post("/topics", h.Topic.Create)
		r.Get("/topics", h.Topic.List)
		r.Patch("/topics/{topicId}", h.Topic.Update)
		r.Delete("/topics/{topicId}", h.Topic.Delete)

		// Segments
		r.Post("/audiences/{audienceId}/segments", h.Segment.Create)
		r.Get("/audiences/{audienceId}/segments", h.Segment.List)
		r.Patch("/audiences/{audienceId}/segments/{segmentId}", h.Segment.Update)
		r.Delete("/audiences/{audienceId}/segments/{segmentId}", h.Segment.Delete)

		// Templates
		r.Post("/templates", h.Template.Create)
		r.Get("/templates", h.Template.List)
		r.Get("/templates/{templateId}", h.Template.Get)
		r.Patch("/templates/{templateId}", h.Template.Update)
		r.Delete("/templates/{templateId}", h.Template.Delete)
		r.Post("/templates/{templateId}/publish", h.Template.Publish)

		// Broadcasts
		r.Post("/broadcasts", h.Broadcast.Create)
		r.Get("/broadcasts", h.Broadcast.List)
		r.Get("/broadcasts/{broadcastId}", h.Broadcast.Get)
		r.Patch("/broadcasts/{broadcastId}", h.Broadcast.Update)
		r.Delete("/broadcasts/{broadcastId}", h.Broadcast.Delete)
		r.Post("/broadcasts/{broadcastId}/send", h.Broadcast.Send)

		// Webhooks
		r.Post("/webhooks", h.Webhook.Create)
		r.Get("/webhooks", h.Webhook.List)
		r.Get("/webhooks/{webhookId}", h.Webhook.Get)
		r.Patch("/webhooks/{webhookId}", h.Webhook.Update)
		r.Delete("/webhooks/{webhookId}", h.Webhook.Delete)

		// Inbound Emails
		r.Get("/inbound/emails", h.InboundEmail.List)
		r.Get("/inbound/emails/{emailId}", h.InboundEmail.Get)

		// Logs
		r.Get("/logs", h.Log.List)

		// Metrics
		r.Get("/metrics", h.Metrics.Get)

		// Settings
		r.Get("/settings/usage", h.Settings.GetUsage)
		r.Get("/settings/team", h.Settings.GetTeam)
		r.Patch("/settings/team", h.Settings.UpdateTeam)
		r.Get("/settings/smtp", h.Settings.GetSMTP)
		r.Post("/settings/team/invite", h.Settings.InviteMember)
	})

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}
