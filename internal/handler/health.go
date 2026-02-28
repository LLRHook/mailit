package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Pinger is satisfied by *pgxpool.Pool directly. For *redis.Client,
// use PingFunc as an adapter.
type Pinger interface {
	Ping(ctx context.Context) error
}

// PingFunc adapts a function to the Pinger interface.
type PingFunc func(ctx context.Context) error

// Ping calls the underlying function.
func (f PingFunc) Ping(ctx context.Context) error { return f(ctx) }

// HealthHandler provides health and readiness endpoints.
type HealthHandler struct {
	pgPinger    Pinger
	redisPinger Pinger
	ready       atomic.Bool
}

// NewHealthHandler creates a HealthHandler that pings the given dependencies.
func NewHealthHandler(pg Pinger, redisPinger Pinger) *HealthHandler {
	h := &HealthHandler{
		pgPinger:    pg,
		redisPinger: redisPinger,
	}
	h.ready.Store(true)
	return h
}

// SetReady sets the readiness flag. Call with false at the start of graceful
// shutdown so /readyz returns 503 while in-flight requests drain.
func (h *HealthHandler) SetReady(v bool) {
	h.ready.Store(v)
}

// Healthz pings Postgres and Redis concurrently and returns 200 if both are
// healthy, or 503 with details about which dependency is degraded.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	type depResult struct {
		name   string
		status string
	}

	var wg sync.WaitGroup
	results := make(chan depResult, 2)

	check := func(name string, p Pinger) {
		defer wg.Done()
		status := "ok"
		if err := p.Ping(ctx); err != nil {
			status = "unavailable"
		}
		results <- depResult{name: name, status: status}
	}

	wg.Add(2)
	go check("postgres", h.pgPinger)
	go check("redis", h.redisPinger)
	wg.Wait()
	close(results)

	deps := make(map[string]string, 2)
	allOK := true
	for res := range results {
		deps[res.name] = res.status
		if res.status != "ok" {
			allOK = false
		}
	}

	status := "ok"
	httpCode := http.StatusOK
	if !allOK {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       status,
		"dependencies": deps,
	})
}

// Readyz returns 503 when the server is shutting down, otherwise delegates to
// Healthz. Load balancers should use this endpoint to decide whether to route
// traffic to this instance.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "shutting_down",
		})
		return
	}
	h.Healthz(w, r)
}
