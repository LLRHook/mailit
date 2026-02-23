package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/pkg"
)

type AuthContext struct {
	TeamID     uuid.UUID
	UserID     *uuid.UUID
	Permission string
	AuthMethod string // "api_key" or "jwt"
}

const AuthContextKey contextKey = "auth"

// APIKeyLookup is the function signature for looking up an API key by hash.
type APIKeyLookup func(ctx context.Context, keyHash string) (*AuthContext, error)

// APIKeyLastUsedUpdate is the function signature for updating last_used_at on an API key.
type APIKeyLastUsedUpdate func(ctx context.Context, keyHash string, usedAt time.Time)

// Auth creates middleware that supports both API key and JWT authentication.
func Auth(jwtSecret string, apiKeyPrefix string, lookupKey APIKeyLookup, updateLastUsed APIKeyLastUsedUpdate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				pkg.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			var authCtx *AuthContext
			var err error

			if strings.HasPrefix(authHeader, "Bearer "+apiKeyPrefix) {
				// API Key auth
				apiKey := strings.TrimPrefix(authHeader, "Bearer ")
				authCtx, err = authenticateAPIKey(r.Context(), apiKey, lookupKey, updateLastUsed)
			} else if strings.HasPrefix(authHeader, "Bearer ") {
				// JWT auth
				token := strings.TrimPrefix(authHeader, "Bearer ")
				authCtx, err = authenticateJWT(token, jwtSecret)
			} else {
				pkg.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			if err != nil {
				pkg.Error(w, http.StatusUnauthorized, "invalid credentials")
				return
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authenticateAPIKey(ctx context.Context, key string, lookup APIKeyLookup, updateLastUsed APIKeyLastUsedUpdate) (*AuthContext, error) {
	h := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(h[:])

	authCtx, err := lookup(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	if updateLastUsed != nil {
		go updateLastUsed(context.Background(), keyHash, time.Now())
	}

	return authCtx, nil
}

func authenticateJWT(tokenStr string, secret string) (*AuthContext, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	teamID, err := uuid.Parse(claims["team_id"].(string))
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, err
	}

	return &AuthContext{
		TeamID:     teamID,
		UserID:     &userID,
		Permission: "full",
		AuthMethod: "jwt",
	}, nil
}

// GetAuth extracts the auth context from the request context.
func GetAuth(ctx context.Context) *AuthContext {
	if auth, ok := ctx.Value(AuthContextKey).(*AuthContext); ok {
		return auth
	}
	return nil
}

// GenerateJWT creates a new JWT token for a user.
func GenerateJWT(secret string, userID, teamID uuid.UUID, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID.String(),
		"team_id": teamID.String(),
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
