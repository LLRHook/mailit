//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mailit-dev/mailit/internal/model"
)

var testPool *pgxpool.Pool

// Fixed IDs used across all integration tests, matching the testutil constants.
var (
	fixedTime  = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	testTeamID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	testUserID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("mailit_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	mig, err := migrate.New("file://../../../db/migrations", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init migrations: %v\n", err)
		os.Exit(1)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}
	srcErr, dbErr := mig.Close()
	if srcErr != nil || dbErr != nil {
		fmt.Fprintf(os.Stderr, "migration close errors: src=%v db=%v\n", srcErr, dbErr)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

func truncateAll(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"email_tracking_links", "email_events", "emails",
		"domain_dns_records", "domains",
		"suppression_list", "api_keys",
		"email_metrics", "webhook_events", "webhooks",
		"broadcasts", "template_versions", "templates",
		"segments", "contact_properties", "contacts", "audiences", "topics",
		"contact_import_jobs", "inbound_emails", "logs",
		"team_invitations", "team_members", "teams", "users",
	}
	for _, table := range tables {
		_, err := testPool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("truncating %s: %v", table, err)
		}
	}
}

func seedTeam(t *testing.T, ctx context.Context) {
	t.Helper()

	_, err := testPool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		testUserID, "test@example.com",
		"$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012",
		"Test User", true, fixedTime)
	if err != nil {
		t.Fatalf("seeding user: %v", err)
	}

	_, err = testPool.Exec(ctx,
		`INSERT INTO teams (id, name, slug, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $4)`,
		testTeamID, "Test Team", "test-team", fixedTime)
	if err != nil {
		t.Fatalf("seeding team: %v", err)
	}

	memberID := uuid.New()
	_, err = testPool.Exec(ctx,
		`INSERT INTO team_members (id, team_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		memberID, testTeamID, testUserID, model.RoleOwner, fixedTime)
	if err != nil {
		t.Fatalf("seeding team member: %v", err)
	}
}

// newTestEmail creates a test email model for integration tests.
func newTestEmail() *model.Email {
	html := "<p>Hello</p>"
	text := "Hello"
	return &model.Email{
		ID:          uuid.New(),
		TeamID:      testTeamID,
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Test Subject",
		HTMLBody:    &html,
		TextBody:    &text,
		Status:      model.EmailStatusQueued,
		Tags:        []string{},
		Headers:     model.JSONMap{},
		Attachments: model.JSONArray{},
		RetryCount:  0,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	}
}

// newTestDomain creates a test domain model for integration tests.
func newTestDomain() *model.Domain {
	privKey := "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"
	return &model.Domain{
		ID:             uuid.New(),
		TeamID:         testTeamID,
		Name:           "example.com",
		Status:         model.DomainStatusPending,
		DKIMPrivateKey: &privKey,
		DKIMSelector:   "mailit",
		OpenTracking:   false,
		ClickTracking:  false,
		TLSPolicy:      "opportunistic",
		CreatedAt:      fixedTime,
		UpdatedAt:      fixedTime,
	}
}

// newTestAPIKey creates a test API key model for integration tests.
func newTestAPIKey() *model.APIKey {
	return &model.APIKey{
		ID:         uuid.New(),
		TeamID:     testTeamID,
		Name:       "Test Key",
		KeyHash:    "abc123hash",
		KeyPrefix:  "re_1234abcd...",
		Permission: model.PermissionFull,
		CreatedAt:  fixedTime,
	}
}

// newTestSuppressionEntry creates a test suppression entry for integration tests.
func newTestSuppressionEntry() *model.SuppressionEntry {
	return &model.SuppressionEntry{
		ID:        uuid.New(),
		TeamID:    testTeamID,
		Email:     "suppressed@example.com",
		Reason:    model.SuppressionBounce,
		CreatedAt: fixedTime,
	}
}
