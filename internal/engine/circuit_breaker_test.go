package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_DefaultStateIsClosed(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)

	// A host with no history should be allowed (closed state).
	assert.True(t, cb.Allow("mx1.example.com"))
}

func TestCircuitBreaker_OpensAfterThresholdFailures(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)
	host := "mx1.example.com"

	// Record exactly threshold failures.
	for i := 0; i < 5; i++ {
		cb.RecordFailure(host)
	}

	// Circuit should now be open.
	assert.False(t, cb.Allow(host))
}

func TestCircuitBreaker_DeniesRequestsWhenOpen(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreaker(3, 5*time.Minute)
	cb.nowFunc = func() time.Time { return now }
	host := "mx1.example.com"

	// Trip the circuit.
	for i := 0; i < 3; i++ {
		cb.RecordFailure(host)
	}

	// Advance time, but not past the reset timeout.
	now = now.Add(2 * time.Minute)
	assert.False(t, cb.Allow(host), "should deny when open and timeout has not elapsed")

	// Repeated checks should still deny.
	assert.False(t, cb.Allow(host))
	assert.False(t, cb.Allow(host))
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreaker(3, 5*time.Minute)
	cb.nowFunc = func() time.Time { return now }
	host := "mx1.example.com"

	// Trip the circuit.
	for i := 0; i < 3; i++ {
		cb.RecordFailure(host)
	}
	require.False(t, cb.Allow(host))

	// Advance time past the reset timeout.
	now = now.Add(6 * time.Minute)

	// Should now be allowed (half-open).
	assert.True(t, cb.Allow(host))

	// Verify state is actually half-open.
	cb.mu.Lock()
	hs := cb.hosts[host]
	assert.Equal(t, circuitStateHalfOpen, hs.state)
	cb.mu.Unlock()
}

func TestCircuitBreaker_ClosesOnSuccessFromHalfOpen(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreaker(3, 5*time.Minute)
	cb.nowFunc = func() time.Time { return now }
	host := "mx1.example.com"

	// Trip the circuit.
	for i := 0; i < 3; i++ {
		cb.RecordFailure(host)
	}

	// Advance past timeout to go half-open.
	now = now.Add(6 * time.Minute)
	require.True(t, cb.Allow(host)) // transitions to half-open

	// Record a success.
	cb.RecordSuccess(host)

	// Should be closed now — allow should return true.
	assert.True(t, cb.Allow(host))

	// Verify internal state.
	cb.mu.Lock()
	hs := cb.hosts[host]
	assert.Equal(t, circuitStateClosed, hs.state)
	assert.Equal(t, 0, hs.consecutiveFailures)
	cb.mu.Unlock()
}

func TestCircuitBreaker_ReOpensOnFailureFromHalfOpen(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreaker(3, 5*time.Minute)
	cb.nowFunc = func() time.Time { return now }
	host := "mx1.example.com"

	// Trip the circuit.
	for i := 0; i < 3; i++ {
		cb.RecordFailure(host)
	}

	// Advance past timeout to go half-open.
	now = now.Add(6 * time.Minute)
	require.True(t, cb.Allow(host)) // transitions to half-open

	// Record a failure — should re-open.
	cb.RecordFailure(host)

	// Should be denied again.
	assert.False(t, cb.Allow(host))

	// Verify internal state.
	cb.mu.Lock()
	hs := cb.hosts[host]
	assert.Equal(t, circuitStateOpen, hs.state)
	cb.mu.Unlock()
}

func TestCircuitBreaker_IndependentTrackingPerHost(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Minute)
	hostA := "mx1.example.com"
	hostB := "mx2.other.com"

	// Trip circuit for hostA only.
	for i := 0; i < 3; i++ {
		cb.RecordFailure(hostA)
	}

	// hostA should be denied, hostB should still be allowed.
	assert.False(t, cb.Allow(hostA))
	assert.True(t, cb.Allow(hostB))

	// Record some failures on hostB but not enough to trip.
	cb.RecordFailure(hostB)
	cb.RecordFailure(hostB)
	assert.True(t, cb.Allow(hostB), "hostB should still be allowed under threshold")
}

func TestCircuitBreaker_SuccessResetsConsecutiveFailureCount(t *testing.T) {
	cb := NewCircuitBreaker(5, 5*time.Minute)
	host := "mx1.example.com"

	// Record 4 failures (one short of threshold).
	for i := 0; i < 4; i++ {
		cb.RecordFailure(host)
	}
	require.True(t, cb.Allow(host), "should still be closed before threshold")

	// Record a success — resets the counter.
	cb.RecordSuccess(host)

	// Record 4 more failures — should not trip because counter was reset.
	for i := 0; i < 4; i++ {
		cb.RecordFailure(host)
	}
	assert.True(t, cb.Allow(host), "should still be closed after reset + 4 failures")

	// One more failure should now trip it (total 5 consecutive since last success).
	cb.RecordFailure(host)
	assert.False(t, cb.Allow(host), "should be open after 5 consecutive failures")
}

func TestCircuitBreaker_DefaultConfigValues(t *testing.T) {
	// Zero values should use defaults.
	cb := NewCircuitBreaker(0, 0)

	assert.Equal(t, defaultFailureThreshold, cb.failureThreshold)
	assert.Equal(t, defaultResetTimeout, cb.resetTimeout)
	assert.NotNil(t, cb.hosts)
	assert.NotNil(t, cb.nowFunc)

	// Negative values should also use defaults.
	cb2 := NewCircuitBreaker(-1, -1*time.Second)
	assert.Equal(t, defaultFailureThreshold, cb2.failureThreshold)
	assert.Equal(t, defaultResetTimeout, cb2.resetTimeout)
}
