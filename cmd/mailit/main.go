package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"

	"github.com/mailit-dev/mailit/internal/config"
	"github.com/mailit-dev/mailit/internal/engine"
	"github.com/mailit-dev/mailit/internal/handler"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/server"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
	smtppkg "github.com/mailit-dev/mailit/internal/smtp"
	"github.com/mailit-dev/mailit/internal/webhook"
	"github.com/mailit-dev/mailit/internal/worker"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	configPath := ""

	switch os.Args[1] {
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		serveCmd.StringVar(&configPath, "config", "config/mailit.yaml", "config file path")
		_ = serveCmd.Parse(os.Args[2:])
		runServe(configPath)
	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		migrateCmd.StringVar(&configPath, "config", "config/mailit.yaml", "config file path")
		up := migrateCmd.Bool("up", false, "run migrations up")
		down := migrateCmd.Bool("down", false, "roll back last migration")
		_ = migrateCmd.Parse(os.Args[2:])
		runMigrate(configPath, *up, *down)
	case "setup":
		setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
		setupCmd.StringVar(&configPath, "config", "config/mailit.yaml", "config file path")
		_ = setupCmd.Parse(os.Args[2:])
		runSetup(configPath)
	case "version":
		fmt.Printf("mailit %s\n", Version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("MailIt - Self-hosted email platform")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mailit serve   [--config path]             Start API server, workers, and SMTP")
	fmt.Println("  mailit migrate [--config path] --up/--down Run database migrations")
	fmt.Println("  mailit setup   [--config path]             First-run setup (admin + DKIM)")
	fmt.Println("  mailit version                             Print version")
}

func runServe(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Set up structured logging.
	logger := setupLogger(cfg.Logging)
	slog.SetDefault(logger)

	logger.Info("starting mailit", "version", Version)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Connect to PostgreSQL.
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		logger.Error("invalid database config", "error", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.Database.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error("connecting to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("pinging database", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to database")

	// Connect to Redis.
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("connecting to redis", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to redis")

	// Health handler (uses pool and redis for liveness checks).
	healthHandler := handler.NewHealthHandler(
		pool,
		handler.PingFunc(func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		}),
	)

	// Run auto-migrations if enabled.
	if cfg.Database.AutoMigrate {
		logger.Info("running auto-migrations")
		connStr := dsnToURL(cfg.Database)
		m, err := migrate.New("file://db/migrations", connStr)
		if err != nil {
			logger.Error("initializing migrations", "error", err)
			os.Exit(1)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Error("running migrations", "error", err)
			os.Exit(1)
		}
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logger.Error("closing migration source", "error", srcErr)
		}
		if dbErr != nil {
			logger.Error("closing migration db", "error", dbErr)
		}
		logger.Info("migrations complete")
	}

	// Set up Asynq client and worker server.
	asynqRedisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer func() { _ = asynqClient.Close() }()

	asynqSrv := asynq.NewServer(asynqRedisOpt, asynq.Config{
		Concurrency: cfg.Workers.Concurrency,
		Queues:      cfg.Workers.Queues,
		Logger:      newAsynqLogger(logger),
	})

	// --- Repositories ---
	userRepo := postgres.NewUserRepository(pool)
	teamRepo := postgres.NewTeamRepository(pool)
	teamMemberRepo := postgres.NewTeamMemberRepository(pool)
	emailRepo := postgres.NewEmailRepository(pool)
	emailEventRepo := postgres.NewEmailEventRepository(pool)
	domainRepo := postgres.NewDomainRepository(pool)
	dnsRecordRepo := postgres.NewDomainDNSRecordRepository(pool)
	apiKeyRepo := postgres.NewAPIKeyRepository(pool)
	audienceRepo := postgres.NewAudienceRepository(pool)
	contactRepo := postgres.NewContactRepository(pool)
	contactPropertyRepo := postgres.NewContactPropertyRepository(pool)
	topicRepo := postgres.NewTopicRepository(pool)
	segmentRepo := postgres.NewSegmentRepository(pool)
	templateRepo := postgres.NewTemplateRepository(pool)
	templateVersionRepo := postgres.NewTemplateVersionRepository(pool)
	broadcastRepo := postgres.NewBroadcastRepository(pool)
	webhookRepo := postgres.NewWebhookRepository(pool)
	webhookEventRepo := postgres.NewWebhookEventRepository(pool)
	suppressionRepo := postgres.NewSuppressionRepository(pool)
	inboundEmailRepo := postgres.NewInboundEmailRepository(pool)
	logRepo := postgres.NewLogRepository(pool)
	metricsRepo := postgres.NewMetricsRepository(pool)
	trackingLinkRepo := postgres.NewTrackingLinkRepository(pool)
	importJobRepo := postgres.NewContactImportJobRepository(pool)
	settingsRepo := postgres.NewSettingsRepository(pool)
	invitationRepo := postgres.NewTeamInvitationRepository(pool)

	// --- Engine ---
	dnsResolver := engine.NewDNSResolver(cfg.DNS.Resolver, cfg.DNS.Timeout)
	smtpSender := engine.NewSender(engine.SenderConfig{
		Hostname:       cfg.SMTPOutbound.Hostname,
		HeloDomain:     cfg.SMTPOutbound.HELODomain,
		TLSPolicy:      cfg.SMTPOutbound.TLSPolicy,
		ConnectTimeout: cfg.SMTPOutbound.ConnectTimeout,
		SendTimeout:    cfg.SMTPOutbound.SendTimeout,
		MaxRecipients:  cfg.SMTPOutbound.MaxRecipients,
	}, dnsResolver, logger)
	emailSenderAdapter := engine.NewWorkerAdapter(smtpSender)

	// --- Webhook Dispatcher ---
	dispatcher := webhook.NewDispatcher(
		webhookRepo,
		webhookEventRepo,
		asynqClient,
		webhook.DispatcherConfig{
			Timeout:    cfg.Webhooks.Timeout,
			MaxRetries: cfg.Webhooks.MaxRetries,
		},
		logger,
	)

	webhookDispatchFn := func(ctx context.Context, teamID uuid.UUID, eventType string, payload interface{}) {
		if err := dispatcher.Dispatch(ctx, teamID, eventType, payload); err != nil {
			logger.Error("webhook dispatch failed", "error", err)
		}
	}

	// --- Services ---
	services := &service.Services{
		Auth:            service.NewAuthService(userRepo, teamRepo, teamMemberRepo, cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry, cfg.Auth.BcryptCost),
		Email:           service.NewEmailService(emailRepo, suppressionRepo, asynqClient, rdb),
		Domain:          service.NewDomainService(domainRepo, dnsRecordRepo, asynqClient, cfg.DKIM.Selector, cfg.DKIM.MasterEncryptionKey),
		APIKey:          service.NewAPIKeyService(apiKeyRepo, cfg.Auth.APIKeyPrefix),
		Audience:        service.NewAudienceService(audienceRepo),
		Contact:         service.NewContactService(contactRepo, audienceRepo),
		ContactProperty: service.NewContactPropertyService(contactPropertyRepo),
		Topic:           service.NewTopicService(topicRepo),
		Segment:         service.NewSegmentService(segmentRepo, audienceRepo),
		Template:        service.NewTemplateService(templateRepo, templateVersionRepo),
		Broadcast:       service.NewBroadcastService(broadcastRepo, asynqClient),
		Webhook:         service.NewWebhookService(webhookRepo),
		InboundEmail:    service.NewInboundEmailService(inboundEmailRepo),
		Log:             service.NewLogService(logRepo),
		Metrics: service.NewMetricsService(metricsRepo),
		Settings: service.NewSettingsService(
			settingsRepo,
			invitationRepo,
			userRepo,
			teamMemberRepo,
			service.SMTPDisplayConfig{
				Host:       cfg.SMTPOutbound.Hostname,
				Port:       587,
				Encryption: "STARTTLS",
			},
			cfg.Auth.JWTSecret,
			cfg.Auth.JWTExpiry,
			cfg.Auth.BcryptCost,
		),
	}

	metricsIncrementFn := func(ctx context.Context, teamID uuid.UUID, eventType string) {
		if err := services.Metrics.IncrementCounter(ctx, teamID, eventType); err != nil {
			logger.Error("metrics increment failed", "error", err, "team_id", teamID, "event_type", eventType)
		}
	}

	// Build tracking service (needs webhookDispatchFn and metricsIncrementFn).
	services.Tracking = service.NewTrackingService(
		trackingLinkRepo,
		emailRepo,
		emailEventRepo,
		contactRepo,
		audienceRepo,
		webhookDispatchFn,
		metricsIncrementFn,
	)

	// --- Handlers ---
	handlers := handler.NewHandlers(services, importJobRepo, audienceRepo, asynqClient)

	// --- API Key auth closures ---
	apiKeyLookup := func(ctx context.Context, keyHash string) (*middleware.AuthContext, error) {
		key, err := apiKeyRepo.GetByHash(ctx, keyHash)
		if err != nil {
			return nil, err
		}
		return &middleware.AuthContext{
			TeamID:     key.TeamID,
			Permission: key.Permission,
			AuthMethod: "api_key",
		}, nil
	}

	apiKeyLastUsed := func(ctx context.Context, keyHash string, usedAt time.Time) {
		_ = apiKeyRepo.UpdateLastUsed(ctx, keyHash, usedAt)
	}

	// --- HTTP Server ---
	httpServer := server.New(server.Config{
		Addr:           cfg.Server.HTTPAddr,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		JWTSecret:      cfg.Auth.JWTSecret,
		APIKeyPrefix:   cfg.Auth.APIKeyPrefix,
		CORSOrigins:    cfg.Server.CORSOrigins,
		RateLimitCfg: middleware.RateLimitConfig{
			Enabled:    cfg.RateLimit.Enabled,
			DefaultRPS: cfg.RateLimit.DefaultRPS,
			SendRPS:    cfg.RateLimit.SendRPS,
			BatchRPS:   cfg.RateLimit.BatchRPS,
			Window:     cfg.RateLimit.Window,
		},
		Redis:          rdb,
		APIKeyLookup:   apiKeyLookup,
		APIKeyLastUsed: apiKeyLastUsed,
		Handlers:       handlers,
		HealthHandler:  healthHandler,
		Logger:         logger,
	})

	// --- Worker Mux ---
	workerHandlers := worker.Handlers{
		EmailSend:      worker.NewEmailSendHandler(emailRepo, emailEventRepo, domainRepo, suppressionRepo, trackingLinkRepo, emailSenderAdapter, webhookDispatchFn, metricsIncrementFn, cfg.Server.BaseURL, logger),
		EmailBatchSend: worker.NewBatchEmailSendHandler(asynqClient, logger),
		BroadcastSend:  worker.NewBroadcastSendHandler(broadcastRepo, contactRepo, audienceRepo, emailRepo, templateVersionRepo, asynqClient, logger),
		DomainVerify:   worker.NewDomainVerifyHandler(domainRepo, dnsRecordRepo, logger),
		Bounce:         worker.NewBounceHandler(emailRepo, emailEventRepo, suppressionRepo, logger),
		Inbound:        worker.NewInboundHandler(inboundEmailRepo, webhookDispatchFn, logger),
		Cleanup:        worker.NewCleanupHandler(webhookEventRepo, logRepo, logger),
		WebhookDeliver:   worker.NewWebhookDeliverHandler(dispatcher, logger),
		MetricsAggregate: worker.NewMetricsAggregateHandler(pool, metricsRepo, logger),
		ContactImport:    worker.NewContactImportHandler(importJobRepo, contactRepo, logger),
	}
	mux := worker.NewMux(workerHandlers)

	// --- Inbound SMTP server (optional) ---
	var smtpServer *gosmtp.Server
	if cfg.SMTPInbound.Enabled {
		attachmentStorage := service.NewLocalAttachmentStorage(cfg.Storage.LocalPath)
		smtpBackend := smtppkg.NewBackend(
			domainRepo,
			inboundEmailRepo,
			attachmentStorage,
			asynqClient,
			int64(cfg.SMTPInbound.MaxMessageBytes),
			logger,
		)
		smtpServer = smtppkg.NewServer(smtppkg.ServerConfig{
			ListenAddr:      cfg.SMTPInbound.ListenAddr,
			Domain:          cfg.SMTPInbound.Domain,
			MaxMessageBytes: int64(cfg.SMTPInbound.MaxMessageBytes),
			ReadTimeout:     cfg.SMTPInbound.ReadTimeout,
			WriteTimeout:    cfg.SMTPInbound.WriteTimeout,
		}, smtpBackend, logger)
	}

	// Run all servers concurrently using errgroup.
	g, gctx := errgroup.WithContext(ctx)

	// HTTP server.
	g.Go(func() error {
		logger.Info("starting HTTP server", "addr", cfg.Server.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	// Asynq worker server.
	g.Go(func() error {
		logger.Info("starting worker server", "concurrency", cfg.Workers.Concurrency)
		if err := asynqSrv.Run(mux); err != nil {
			return fmt.Errorf("asynq worker: %w", err)
		}
		return nil
	})

	// Inbound SMTP server.
	if smtpServer != nil {
		g.Go(func() error {
			logger.Info("starting inbound SMTP server", "addr", cfg.SMTPInbound.ListenAddr)
			if err := smtpServer.ListenAndServe(); err != nil {
				return fmt.Errorf("smtp server: %w", err)
			}
			return nil
		})
	}

	// Graceful shutdown goroutine.
	g.Go(func() error {
		<-gctx.Done()
		logger.Info("shutting down...")
		healthHandler.SetReady(false)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer shutdownCancel()

		// Shutdown HTTP server.
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("http server shutdown", "error", err)
		}

		// Shutdown Asynq worker server.
		asynqSrv.Shutdown()

		// Shutdown inbound SMTP server.
		if smtpServer != nil {
			if err := smtpServer.Close(); err != nil {
				logger.Error("smtp server shutdown", "error", err)
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("mailit stopped")
}

func runMigrate(configPath string, up, down bool) {
	if !up && !down {
		fmt.Fprintln(os.Stderr, "Error: specify --up or --down")
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	connStr := dsnToURL(cfg.Database)

	m, err := migrate.New("file://db/migrations", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing migrations: %v\n", err)
		os.Exit(1)
	}
	defer func() { _, _ = m.Close() }()

	if up {
		fmt.Println("Running migrations up...")
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("No new migrations to apply.")
				return
			}
			fmt.Fprintf(os.Stderr, "Error running migrations up: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully.")
	}

	if down {
		fmt.Println("Rolling back last migration...")
		if err := m.Steps(-1); err != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back migration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration rolled back successfully.")
	}
}

func runSetup(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to the database.
	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error pinging database: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Prompt for admin details.
	fmt.Print("Admin name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Admin password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Team name [Default Team]: ")
	teamName, _ := reader.ReadString('\n')
	teamName = strings.TrimSpace(teamName)
	if teamName == "" {
		teamName = "Default Team"
	}

	// Hash the password.
	bcryptCost := cfg.Auth.BcryptCost
	if bcryptCost == 0 {
		bcryptCost = 12
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Create user, team, and team_member in a transaction.
	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting transaction: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()
	now := time.Now()
	slug := strings.ToLower(strings.ReplaceAll(teamName, " ", "-"))

	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, $5, $5)`,
		userID, email, string(hash), name, now,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		os.Exit(1)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO teams (id, name, slug, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $4)`,
		teamID, teamName, slug, now,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating team: %v\n", err)
		os.Exit(1)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO team_members (id, team_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, 'owner', $4)`,
		memberID, teamID, userID, now,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating team member: %v\n", err)
		os.Exit(1)
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error committing transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Admin user created successfully!")
	fmt.Printf("  User ID: %s\n", userID)
	fmt.Printf("  Email:   %s\n", email)
	fmt.Printf("  Team:    %s (ID: %s)\n", teamName, teamID)
	fmt.Println()

	// Generate DKIM keys.
	keyBits := cfg.DKIM.KeyBits
	if keyBits == 0 {
		keyBits = 2048
	}
	selector := cfg.DKIM.Selector
	if selector == "" {
		selector = "mailit"
	}

	fmt.Printf("Generating %d-bit DKIM key pair (selector: %s)...\n", keyBits, selector)

	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating DKIM key: %v\n", err)
		os.Exit(1)
	}

	// Encode private key to PEM.
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode public key to DER for DNS record.
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding public key: %v\n", err)
		os.Exit(1)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	// Build the base64 public key value (strip PEM headers/footers for DNS).
	pubLines := strings.Split(string(pubPEM), "\n")
	var pubBase64 string
	for _, line := range pubLines {
		if strings.HasPrefix(line, "-----") || line == "" {
			continue
		}
		pubBase64 += line
	}

	fmt.Println()
	fmt.Println("=== DKIM DNS Record ===")
	fmt.Printf("Add a TXT record for: %s._domainkey.<your-domain>\n", selector)
	fmt.Printf("Value: v=DKIM1; k=rsa; p=%s\n", pubBase64)
	fmt.Println()
	fmt.Println("=== DKIM Private Key (store securely) ===")
	fmt.Println(string(privPEM))
	fmt.Println()
	fmt.Println("Setup complete! You can now start the server with: mailit serve")
}

// setupLogger creates a slog.Logger based on the logging config.
func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// dsnToURL converts the DatabaseConfig into a postgres:// connection URL
// suitable for golang-migrate.
func dsnToURL(db config.DatabaseConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.DBName, db.SSLMode,
	)
}

// asynqLogger adapts slog to the asynq Logger interface.
type asynqLogger struct {
	logger *slog.Logger
}

func newAsynqLogger(logger *slog.Logger) *asynqLogger {
	return &asynqLogger{logger: logger.With("component", "asynq")}
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
	os.Exit(1)
}
