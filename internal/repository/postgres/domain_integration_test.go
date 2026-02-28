//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/model"
)

func TestDomainRepository_Create(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewDomainRepository(testPool)
	domain := newTestDomain()

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, domain.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ID, got.ID)
	assert.Equal(t, domain.TeamID, got.TeamID)
	assert.Equal(t, domain.Name, got.Name)
	assert.Equal(t, model.DomainStatusPending, got.Status)
	assert.Equal(t, domain.DKIMSelector, got.DKIMSelector)
	assert.Equal(t, domain.TLSPolicy, got.TLSPolicy)
}

func TestDomainRepository_GetVerifiedByName(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewDomainRepository(testPool)

	// Create a pending domain
	pending := newTestDomain()
	pending.Name = "pending.example.com"
	pending.Status = model.DomainStatusPending
	err := repo.Create(ctx, pending)
	require.NoError(t, err)

	// Create a verified domain
	verified := newTestDomain()
	verified.ID = uuid.New()
	verified.Name = "verified.example.com"
	verified.Status = model.DomainStatusVerified
	err = repo.Create(ctx, verified)
	require.NoError(t, err)

	// Verified domain should be found
	got, err := repo.GetVerifiedByName(ctx, "verified.example.com")
	require.NoError(t, err)
	assert.Equal(t, verified.ID, got.ID)
	assert.Equal(t, model.DomainStatusVerified, got.Status)

	// Pending domain should not be found via GetVerifiedByName
	_, err = repo.GetVerifiedByName(ctx, "pending.example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound for pending domain, got: %v", err)
}

func TestDomainRepository_UniqueConstraint(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewDomainRepository(testPool)

	domain1 := newTestDomain()
	domain1.Name = "unique-test.example.com"
	err := repo.Create(ctx, domain1)
	require.NoError(t, err)

	// Same team + same domain name should violate unique constraint
	domain2 := newTestDomain()
	domain2.ID = uuid.New()
	domain2.Name = "unique-test.example.com"
	err = repo.Create(ctx, domain2)
	require.Error(t, err, "expected unique constraint violation")
}

func TestDomainRepository_Delete(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewDomainRepository(testPool)
	domain := newTestDomain()

	err := repo.Create(ctx, domain)
	require.NoError(t, err)

	// Verify the domain exists
	_, err = repo.GetByID(ctx, domain.ID)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, domain.ID)
	require.NoError(t, err)

	// Verify it is gone
	_, err = repo.GetByID(ctx, domain.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))

	// Deleting again should return not found
	err = repo.Delete(ctx, domain.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}
