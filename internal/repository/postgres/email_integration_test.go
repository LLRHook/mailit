//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/model"
)

func TestEmailRepository_CreateAndGetByID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewEmailRepository(testPool)
	email := newTestEmail()

	err := repo.Create(ctx, email)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, email.ID)
	require.NoError(t, err)
	assert.Equal(t, email.ID, got.ID)
	assert.Equal(t, email.TeamID, got.TeamID)
	assert.Equal(t, email.FromAddress, got.FromAddress)
	assert.Equal(t, email.ToAddresses, got.ToAddresses)
	assert.Equal(t, email.Subject, got.Subject)
	assert.Equal(t, email.Status, got.Status)
	assert.Equal(t, *email.HTMLBody, *got.HTMLBody)
	assert.Equal(t, *email.TextBody, *got.TextBody)
}

func TestEmailRepository_GetByTeamAndID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewEmailRepository(testPool)
	email := newTestEmail()

	err := repo.Create(ctx, email)
	require.NoError(t, err)

	// Correct team should find the email
	got, err := repo.GetByTeamAndID(ctx, testTeamID, email.ID)
	require.NoError(t, err)
	assert.Equal(t, email.ID, got.ID)

	// Wrong team should return not found
	wrongTeamID := uuid.New()
	_, err = repo.GetByTeamAndID(ctx, wrongTeamID, email.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound, got: %v", err)
}

func TestEmailRepository_ListWithPagination(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewEmailRepository(testPool)

	// Create 3 emails with slightly different timestamps for ordering
	for i := 0; i < 3; i++ {
		email := newTestEmail()
		email.ID = uuid.New()
		email.Subject = "Email " + string(rune('A'+i))
		email.CreatedAt = fixedTime.Add(time.Duration(i) * time.Minute)
		email.UpdatedAt = email.CreatedAt
		err := repo.Create(ctx, email)
		require.NoError(t, err)
	}

	// List with limit 2
	emails, total, err := repo.List(ctx, testTeamID, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, emails, 2)

	// The list should be ordered by created_at DESC, so the newest first
	assert.Equal(t, "Email C", emails[0].Subject)
	assert.Equal(t, "Email B", emails[1].Subject)

	// List with offset 2 to get the remaining one
	emails, total, err = repo.List(ctx, testTeamID, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, emails, 1)
	assert.Equal(t, "Email A", emails[0].Subject)
}

func TestEmailRepository_UpdateStatus(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewEmailRepository(testPool)
	email := newTestEmail()

	err := repo.Create(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, model.EmailStatusQueued, email.Status)

	// Update the status to sent
	now := time.Now().UTC().Truncate(time.Microsecond)
	email.Status = model.EmailStatusSent
	email.SentAt = &now
	email.UpdatedAt = now

	err = repo.Update(ctx, email)
	require.NoError(t, err)

	// Verify the update persisted
	got, err := repo.GetByID(ctx, email.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EmailStatusSent, got.Status)
	assert.NotNil(t, got.SentAt)
}
