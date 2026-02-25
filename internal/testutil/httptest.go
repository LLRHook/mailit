package testutil

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/server/middleware"
)

// SetupRouter creates a chi router with a route registration function.
// Used in handler tests to mount specific handler methods.
func SetupRouter(register func(r chi.Router)) *chi.Mux {
	r := chi.NewRouter()
	register(r)
	return r
}

// AuthenticatedRequest injects an AuthContext into the request for handler testing.
func AuthenticatedRequest(r *http.Request, teamID uuid.UUID, userID uuid.UUID) *http.Request {
	authCtx := &middleware.AuthContext{
		TeamID:     teamID,
		UserID:     &userID,
		Permission: "full",
		AuthMethod: "jwt",
	}
	ctx := context.WithValue(r.Context(), middleware.AuthContextKey, authCtx)
	return r.WithContext(ctx)
}

// APIKeyRequest injects an API key AuthContext into the request.
func APIKeyRequest(r *http.Request, teamID uuid.UUID, permission string) *http.Request {
	authCtx := &middleware.AuthContext{
		TeamID:     teamID,
		Permission: permission,
		AuthMethod: "api_key",
	}
	ctx := context.WithValue(r.Context(), middleware.AuthContextKey, authCtx)
	return r.WithContext(ctx)
}

// WithURLParam adds a chi URL parameter to the request context.
func WithURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// WithURLParams adds multiple chi URL parameters to the request context.
func WithURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
