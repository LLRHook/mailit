package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuth_MissingAuthorizationHeader(t *testing.T) {
	handler := Auth("secret", "re_", nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidFormat(t *testing.T) {
	handler := Auth("secret", "re_", nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ValidJWT(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New()
	teamID := uuid.New()

	token, err := GenerateJWT(secret, userID, teamID, 1*time.Hour)
	assert.NoError(t, err)

	var capturedAuth *AuthContext
	handler := Auth(secret, "re_", nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = GetAuth(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, capturedAuth)
	assert.Equal(t, teamID, capturedAuth.TeamID)
	assert.Equal(t, &userID, capturedAuth.UserID)
	assert.Equal(t, "jwt", capturedAuth.AuthMethod)
	assert.Equal(t, "full", capturedAuth.Permission)
}

func TestAuth_ExpiredJWT(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New()
	teamID := uuid.New()

	// Generate a token that's already expired
	token, err := GenerateJWT(secret, userID, teamID, -1*time.Hour)
	assert.NoError(t, err)

	handler := Auth(secret, "re_", nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidJWTSignature(t *testing.T) {
	userID := uuid.New()
	teamID := uuid.New()

	// Generate with one secret, verify with another
	token, err := GenerateJWT("wrong-secret", userID, teamID, 1*time.Hour)
	assert.NoError(t, err)

	handler := Auth("correct-secret", "re_", nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ValidAPIKey(t *testing.T) {
	apiKeyPrefix := "re_"
	apiKey := "re_test_key_12345"
	teamID := uuid.New()

	lookup := func(ctx context.Context, keyHash string) (*AuthContext, error) {
		return &AuthContext{
			TeamID:     teamID,
			Permission: "full",
			AuthMethod: "api_key",
		}, nil
	}

	var lastUsedCalled atomic.Bool
	updateLastUsed := func(ctx context.Context, keyHash string, usedAt time.Time) {
		lastUsedCalled.Store(true)
	}

	var capturedAuth *AuthContext
	handler := Auth("jwt-secret", apiKeyPrefix, lookup, updateLastUsed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = GetAuth(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, capturedAuth)
	assert.Equal(t, teamID, capturedAuth.TeamID)
	assert.Equal(t, "api_key", capturedAuth.AuthMethod)

	// Wait briefly for the goroutine to execute
	time.Sleep(50 * time.Millisecond)
	assert.True(t, lastUsedCalled.Load())
}

func TestAuth_InvalidAPIKey(t *testing.T) {
	apiKeyPrefix := "re_"
	apiKey := "re_bad_key"

	lookup := func(ctx context.Context, keyHash string) (*AuthContext, error) {
		return nil, assert.AnError
	}

	handler := Auth("jwt-secret", apiKeyPrefix, lookup, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGenerateJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	teamID := uuid.New()

	token, err := GenerateJWT(secret, userID, teamID, 1*time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGetAuth_NoContext(t *testing.T) {
	ctx := context.Background()
	auth := GetAuth(ctx)
	assert.Nil(t, auth)
}

func TestGetAuth_WithContext(t *testing.T) {
	teamID := uuid.New()
	authCtx := &AuthContext{
		TeamID:     teamID,
		Permission: "full",
		AuthMethod: "jwt",
	}
	ctx := context.WithValue(context.Background(), AuthContextKey, authCtx)

	auth := GetAuth(ctx)
	assert.NotNil(t, auth)
	assert.Equal(t, teamID, auth.TeamID)
}
