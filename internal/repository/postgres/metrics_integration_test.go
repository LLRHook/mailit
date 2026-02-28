//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mailit-dev/mailit/internal/model"
)

func TestMetricsRepository_UpsertInsert(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewMetricsRepository(testPool)

	m := &model.EmailMetrics{
		TeamID:      testTeamID,
		PeriodStart: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		PeriodType:  model.PeriodTypeHourly,
		Sent:        5,
		Delivered:   3,
		Bounced:     1,
		Failed:      0,
		Opened:      2,
		Clicked:     1,
		Complained:  0,
	}

	err := repo.Upsert(ctx, m)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, m.ID, "ID should be populated after upsert")

	// Verify by listing
	results, err := repo.ListByTeam(ctx, testTeamID, model.PeriodTypeHourly,
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 5, results[0].Sent)
	assert.Equal(t, 3, results[0].Delivered)
	assert.Equal(t, 1, results[0].Bounced)
}

func TestMetricsRepository_UpsertIncrement(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewMetricsRepository(testPool)

	periodStart := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	// First upsert (insert)
	m1 := &model.EmailMetrics{
		TeamID:      testTeamID,
		PeriodStart: periodStart,
		PeriodType:  model.PeriodTypeHourly,
		Sent:        5,
		Delivered:   3,
		Bounced:     1,
		Failed:      0,
		Opened:      2,
		Clicked:     1,
		Complained:  0,
	}
	err := repo.Upsert(ctx, m1)
	require.NoError(t, err)

	// Second upsert (increment via ON CONFLICT)
	m2 := &model.EmailMetrics{
		TeamID:      testTeamID,
		PeriodStart: periodStart,
		PeriodType:  model.PeriodTypeHourly,
		Sent:        3,
		Delivered:   2,
		Bounced:     0,
		Failed:      1,
		Opened:      0,
		Clicked:     0,
		Complained:  1,
	}
	err = repo.Upsert(ctx, m2)
	require.NoError(t, err)

	// Verify the values were added (not replaced)
	results, err := repo.ListByTeam(ctx, testTeamID, model.PeriodTypeHourly,
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 8, results[0].Sent)      // 5 + 3
	assert.Equal(t, 5, results[0].Delivered)  // 3 + 2
	assert.Equal(t, 1, results[0].Bounced)    // 1 + 0
	assert.Equal(t, 1, results[0].Failed)     // 0 + 1
	assert.Equal(t, 2, results[0].Opened)     // 2 + 0
	assert.Equal(t, 1, results[0].Clicked)    // 1 + 0
	assert.Equal(t, 1, results[0].Complained) // 0 + 1
}

func TestMetricsRepository_ListByTeamWithDateRange(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewMetricsRepository(testPool)

	// Create metrics for 3 different hours
	for i := 0; i < 3; i++ {
		m := &model.EmailMetrics{
			TeamID:      testTeamID,
			PeriodStart: time.Date(2025, 1, 15, 8+i, 0, 0, 0, time.UTC),
			PeriodType:  model.PeriodTypeHourly,
			Sent:        (i + 1) * 10,
		}
		err := repo.Upsert(ctx, m)
		require.NoError(t, err)
	}

	// Query a range that includes only the first 2 hours
	results, err := repo.ListByTeam(ctx, testTeamID, model.PeriodTypeHourly,
		time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Len(t, results, 2)
	// Ordered by period_start ASC
	assert.Equal(t, 10, results[0].Sent)
	assert.Equal(t, 20, results[1].Sent)
}

func TestMetricsRepository_AggregateTotalsEmpty(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	seedTeam(t, ctx)

	repo := NewMetricsRepository(testPool)

	// Query with no data should return zeroes
	totals, err := repo.AggregateTotals(ctx, testTeamID, model.PeriodTypeDaily,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, totals)
	assert.Equal(t, 0, totals.Sent)
	assert.Equal(t, 0, totals.Delivered)
	assert.Equal(t, 0, totals.Bounced)
	assert.Equal(t, 0, totals.Failed)
	assert.Equal(t, 0, totals.Opened)
	assert.Equal(t, 0, totals.Clicked)
	assert.Equal(t, 0, totals.Complained)
}
