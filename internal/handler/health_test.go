package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPinger implements Pinger for testing.
type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestHealthz_BothHealthy(t *testing.T) {
	h := NewHealthHandler(&mockPinger{}, &mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])

	deps := body["dependencies"].(map[string]interface{})
	assert.Equal(t, "ok", deps["postgres"])
	assert.Equal(t, "ok", deps["redis"])
}

func TestHealthz_PostgresDown(t *testing.T) {
	h := NewHealthHandler(
		&mockPinger{err: fmt.Errorf("connection refused")},
		&mockPinger{},
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])

	deps := body["dependencies"].(map[string]interface{})
	assert.Equal(t, "unavailable", deps["postgres"])
	assert.Equal(t, "ok", deps["redis"])
}

func TestHealthz_RedisDown(t *testing.T) {
	h := NewHealthHandler(
		&mockPinger{},
		&mockPinger{err: fmt.Errorf("connection refused")},
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])

	deps := body["dependencies"].(map[string]interface{})
	assert.Equal(t, "ok", deps["postgres"])
	assert.Equal(t, "unavailable", deps["redis"])
}

func TestHealthz_BothDown(t *testing.T) {
	h := NewHealthHandler(
		&mockPinger{err: fmt.Errorf("pg down")},
		&mockPinger{err: fmt.Errorf("redis down")},
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body["status"])

	deps := body["dependencies"].(map[string]interface{})
	assert.Equal(t, "unavailable", deps["postgres"])
	assert.Equal(t, "unavailable", deps["redis"])
}

func TestReadyz_Healthy(t *testing.T) {
	h := NewHealthHandler(&mockPinger{}, &mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	h.Readyz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}

func TestReadyz_ShuttingDown(t *testing.T) {
	h := NewHealthHandler(&mockPinger{}, &mockPinger{})
	h.SetReady(false)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	h.Readyz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "shutting_down", body["status"])
}
