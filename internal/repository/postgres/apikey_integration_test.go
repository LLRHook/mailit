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
)

func TestAPIKeyRepository_Create(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewAPIKeyRepository(testPool)
	key := newTestAPIKey()

	err := repo.Create(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, testTeamID, key.TeamID)
	assert.Equal(t, "Test Key", key.Name)
	assert.Equal(t, "abc123hash", key.KeyHash)
	assert.Equal(t, "re_1234abcd...", key.KeyPrefix)
}

func TestAPIKeyRepository_GetByHash(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewAPIKeyRepository(testPool)
	key := newTestAPIKey()

	err := repo.Create(ctx, key)
	require.NoError(t, err)

	// Should find by hash
	got, err := repo.GetByHash(ctx, "abc123hash")
	require.NoError(t, err)
	assert.Equal(t, key.ID, got.ID)
	assert.Equal(t, key.TeamID, got.TeamID)
	assert.Equal(t, key.Name, got.Name)
	assert.Equal(t, key.KeyHash, got.KeyHash)

	// Non-existent hash
	_, err = repo.GetByHash(ctx, "nonexistent_hash")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestAPIKeyRepository_ListByTeamID(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewAPIKeyRepository(testPool)

	// Create 3 keys
	for i := 0; i < 3; i++ {
		key := newTestAPIKey()
		key.ID = uuid.New()
		key.KeyHash = "hash_" + string(rune('a'+i))
		key.Name = "Key " + string(rune('A'+i))
		err := repo.Create(ctx, key)
		require.NoError(t, err)
	}

	keys, err := repo.ListByTeamID(ctx, testTeamID)
	require.NoError(t, err)
	assert.Len(t, keys, 3)

	// Should not find keys for a different team
	otherTeamID := uuid.New()
	keys, err = repo.ListByTeamID(ctx, otherTeamID)
	require.NoError(t, err)
	assert.Len(t, keys, 0)
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewAPIKeyRepository(testPool)
	key := newTestAPIKey()

	err := repo.Create(ctx, key)
	require.NoError(t, err)
	assert.Nil(t, key.LastUsedAt)

	// Update last used
	now := time.Now().UTC().Truncate(time.Microsecond)
	err = repo.UpdateLastUsed(ctx, key.KeyHash, now)
	require.NoError(t, err)

	// Verify the update persisted
	got, err := repo.GetByHash(ctx, key.KeyHash)
	require.NoError(t, err)
	require.NotNil(t, got.LastUsedAt)
	assert.WithinDuration(t, now, *got.LastUsedAt, time.Second)

	// Update non-existent hash should return not found
	err = repo.UpdateLastUsed(ctx, "nonexistent_hash", now)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}
