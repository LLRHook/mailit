//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuppressionRepository_Create(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewSuppressionRepository(testPool)
	entry := newTestSuppressionEntry()

	err := repo.Create(ctx, entry)
	require.NoError(t, err)
	assert.Equal(t, testTeamID, entry.TeamID)
	assert.Equal(t, "suppressed@example.com", entry.Email)
	assert.Equal(t, "bounce", entry.Reason)
}

func TestSuppressionRepository_GetByTeamAndEmail(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewSuppressionRepository(testPool)
	entry := newTestSuppressionEntry()

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// Should find by team + email
	got, err := repo.GetByTeamAndEmail(ctx, testTeamID, "suppressed@example.com")
	require.NoError(t, err)
	assert.Equal(t, entry.ID, got.ID)
	assert.Equal(t, entry.Email, got.Email)
	assert.Equal(t, entry.Reason, got.Reason)

	// Should not find with wrong team
	wrongTeamID := uuid.New()
	_, err = repo.GetByTeamAndEmail(ctx, wrongTeamID, "suppressed@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))

	// Should not find with wrong email
	_, err = repo.GetByTeamAndEmail(ctx, testTeamID, "nonexistent@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestSuppressionRepository_UniqueConstraint(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewSuppressionRepository(testPool)

	entry1 := newTestSuppressionEntry()
	err := repo.Create(ctx, entry1)
	require.NoError(t, err)

	// Same team + same email should violate unique constraint
	entry2 := newTestSuppressionEntry()
	entry2.ID = uuid.New()
	err = repo.Create(ctx, entry2)
	require.Error(t, err, "expected unique constraint violation for duplicate team+email")
}

func TestSuppressionRepository_Delete(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewSuppressionRepository(testPool)
	entry := newTestSuppressionEntry()

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, entry.ID)
	require.NoError(t, err)

	// Verify it is gone
	_, err = repo.GetByTeamAndEmail(ctx, testTeamID, entry.Email)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))

	// Deleting again should return not found
	err = repo.Delete(ctx, entry.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}
