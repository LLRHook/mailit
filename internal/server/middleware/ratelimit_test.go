package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, rdb
}

func TestRateLimit_Disabled(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    false,
		DefaultRPS: 5,
	}

	handler := RateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimit_NoAuth(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 5,
	}

	handler := RateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should pass through when there's no auth context
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimit_UnderLimit(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 10,
		Window:     time.Second,
	}

	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}

	handler := RateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimit_ExceedLimit(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 2,
		Window:     time.Second,
	}

	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}

	handler := RateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Send requests until we exceed the limit
	var lastCode int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), AuthContextKey, authCtx)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		lastCode = rec.Code
	}

	// After exceeding, we should get 429
	assert.Equal(t, http.StatusTooManyRequests, lastCode)
}

func TestRateLimit_HeadersPresent(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 100,
		Window:     time.Second,
	}

	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}

	handler := RateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "100", rec.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
}

func TestSendRateLimit(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 100,
		SendRPS:    10,
		Window:     time.Second,
	}

	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}

	handler := SendRateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/emails", nil)
	ctx := context.WithValue(req.Context(), AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Send rate limit should use SendRPS, not DefaultRPS
	assert.Equal(t, "10", rec.Header().Get("X-RateLimit-Limit"))
}

func TestBatchRateLimit(t *testing.T) {
	_, rdb := setupMiniredis(t)

	cfg := RateLimitConfig{
		Enabled:    true,
		DefaultRPS: 100,
		BatchRPS:   5,
		Window:     time.Second,
	}

	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}

	handler := BatchRateLimit(rdb, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/emails/batch", nil)
	ctx := context.WithValue(req.Context(), AuthContextKey, authCtx)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "5", rec.Header().Get("X-RateLimit-Limit"))
}
