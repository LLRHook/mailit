package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/dto"
	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/repository/postgres"
	"github.com/mailit-dev/mailit/internal/testutil"
	tmock "github.com/mailit-dev/mailit/internal/testutil/mock"
)

func newEmailTestDeps(t *testing.T) (*tmock.MockEmailRepository, *tmock.MockSuppressionRepository, *asynq.Client, *redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: mr.Addr()})
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return new(tmock.MockEmailRepository), new(tmock.MockSuppressionRepository), asynqClient, redisClient, mr
}

func TestEmailService_Send_HappyPath(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	// No suppression entry for recipient.
	suppressionRepo.On("GetByTeamAndEmail", ctx, teamID, "recipient@example.com").Return(nil, postgres.ErrNotFound)
	emailRepo.On("Create", ctx, mock.AnythingOfType("*model.Email")).Return(nil)

	req := &dto.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello",
		HTML:    testutil.StringPtr("<p>Hello</p>"),
	}

	resp, err := svc.Send(ctx, teamID, req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	emailRepo.AssertExpectations(t)
	suppressionRepo.AssertExpectations(t)
}

func TestEmailService_Send_IdempotencyKey_DuplicateReturnsSameID(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	// First send.
	suppressionRepo.On("GetByTeamAndEmail", ctx, teamID, "recipient@example.com").Return(nil, postgres.ErrNotFound)
	emailRepo.On("Create", ctx, mock.AnythingOfType("*model.Email")).Return(nil)

	idempotencyKey := "unique-key-123"
	req := &dto.SendEmailRequest{
		From:           "sender@example.com",
		To:             []string{"recipient@example.com"},
		Subject:        "Hello",
		HTML:           testutil.StringPtr("<p>Hello</p>"),
		IdempotencyKey: &idempotencyKey,
	}

	resp1, err := svc.Send(ctx, teamID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp1.ID)

	// Second send with same idempotency key should return same ID.
	resp2, err := svc.Send(ctx, teamID, req)
	require.NoError(t, err)
	assert.Equal(t, resp1.ID, resp2.ID)
}

func TestEmailService_Send_SuppressedRecipient(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	suppressed := testutil.NewTestSuppressionEntry()
	suppressed.Email = "blocked@example.com"
	suppressionRepo.On("GetByTeamAndEmail", ctx, teamID, "blocked@example.com").Return(suppressed, nil)

	req := &dto.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"blocked@example.com"},
		Subject: "Hello",
		HTML:    testutil.StringPtr("<p>Hello</p>"),
	}

	resp, err := svc.Send(ctx, teamID, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suppression list")

	suppressionRepo.AssertExpectations(t)
}

func TestEmailService_List_Paginated(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email1 := *testutil.NewTestEmail()
	email2 := *testutil.NewTestEmail()
	emailRepo.On("List", ctx, teamID, 20, 0).Return([]model.Email{email1, email2}, 2, nil)

	params := &dto.PaginationParams{Page: 1, PerPage: 20}
	resp, err := svc.List(ctx, teamID, params)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 1, resp.Page)

	emailRepo.AssertExpectations(t)
}

func TestEmailService_Get_HappyPath(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email := testutil.NewTestEmail()
	emailRepo.On("GetByID", ctx, email.ID).Return(email, nil)

	resp, err := svc.Get(ctx, teamID, email.ID)

	require.NoError(t, err)
	assert.Equal(t, email.ID.String(), resp.ID)
	assert.Equal(t, email.Subject, resp.Subject)

	emailRepo.AssertExpectations(t)
}

func TestEmailService_Get_WrongTeam(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	wrongTeamID := uuid.New()

	email := testutil.NewTestEmail()
	emailRepo.On("GetByID", ctx, email.ID).Return(email, nil)

	resp, err := svc.Get(ctx, wrongTeamID, email.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	emailRepo.AssertExpectations(t)
}

func TestEmailService_Cancel_HappyPath(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email := testutil.NewTestEmail()
	email.Status = model.EmailStatusQueued
	emailRepo.On("GetByID", ctx, email.ID).Return(email, nil)
	emailRepo.On("Update", ctx, mock.AnythingOfType("*model.Email")).Return(nil)

	resp, err := svc.Cancel(ctx, teamID, email.ID)

	require.NoError(t, err)
	assert.Equal(t, model.EmailStatusCancelled, resp.Status)

	emailRepo.AssertExpectations(t)
}

func TestEmailService_Cancel_WrongStatus(t *testing.T) {
	emailRepo, suppressionRepo, asynqClient, redisClient, _ := newEmailTestDeps(t)
	svc := NewEmailService(emailRepo, suppressionRepo, asynqClient, redisClient)
	ctx := context.Background()
	teamID := testutil.TestTeamID

	email := testutil.NewTestEmail()
	email.Status = model.EmailStatusSent
	emailRepo.On("GetByID", ctx, email.ID).Return(email, nil)

	resp, err := svc.Cancel(ctx, teamID, email.ID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only queued or scheduled")

	emailRepo.AssertExpectations(t)
}
