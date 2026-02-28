package engine

import (
	"sync"
	"time"
)

const (
	circuitStateClosed   = "closed"
	circuitStateOpen     = "open"
	circuitStateHalfOpen = "half-open"

	defaultFailureThreshold = 5
	defaultResetTimeout     = 5 * time.Minute
)

// CircuitBreaker implements a per-MX-host circuit breaker that prevents
// repeated delivery attempts to hosts that are consistently failing.
// States: closed -> open (after threshold consecutive failures) -> half-open (after timeout).
type CircuitBreaker struct {
	mu               sync.Mutex
	hosts            map[string]*hostState
	failureThreshold int
	resetTimeout     time.Duration
	nowFunc          func() time.Time
}

// hostState tracks the circuit breaker state for a single MX host.
type hostState struct {
	state               string
	consecutiveFailures int
	lastFailureTime     time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker. Zero values for
// failureThreshold or resetTimeout are replaced with sensible defaults
// (5 failures, 5 minutes).
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = defaultFailureThreshold
	}
	if resetTimeout <= 0 {
		resetTimeout = defaultResetTimeout
	}

	return &CircuitBreaker{
		hosts:            make(map[string]*hostState),
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		nowFunc:          time.Now,
	}
}

// Allow returns true if the circuit for the given host permits a delivery
// attempt. It returns true for closed and half-open states, false for open.
// If the host has no state yet, it is treated as closed (allowed).
func (cb *CircuitBreaker) Allow(host string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	hs, exists := cb.hosts[host]
	if !exists {
		return true
	}

	switch hs.state {
	case circuitStateClosed:
		return true
	case circuitStateOpen:
		// Check if the reset timeout has elapsed; if so, transition to half-open.
		if cb.nowFunc().Sub(hs.lastFailureTime) >= cb.resetTimeout {
			hs.state = circuitStateHalfOpen
			return true
		}
		return false
	case circuitStateHalfOpen:
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful delivery to the given host. If the
// circuit is half-open, it transitions back to closed. In any state, the
// consecutive failure count is reset to zero.
func (cb *CircuitBreaker) RecordSuccess(host string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	hs, exists := cb.hosts[host]
	if !exists {
		return
	}

	hs.consecutiveFailures = 0
	hs.state = circuitStateClosed
}

// RecordFailure records a failed delivery to the given host. If the
// consecutive failure count reaches the threshold, the circuit opens.
// If the circuit is half-open, a single failure re-opens it.
func (cb *CircuitBreaker) RecordFailure(host string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	hs, exists := cb.hosts[host]
	if !exists {
		hs = &hostState{state: circuitStateClosed}
		cb.hosts[host] = hs
	}

	hs.consecutiveFailures++
	hs.lastFailureTime = cb.nowFunc()

	switch hs.state {
	case circuitStateClosed:
		if hs.consecutiveFailures >= cb.failureThreshold {
			hs.state = circuitStateOpen
		}
	case circuitStateHalfOpen:
		// Any failure from half-open re-opens the circuit.
		hs.state = circuitStateOpen
	}
}
