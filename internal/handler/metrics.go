package handler

import (
	"net/http"

	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type MetricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(s service.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: s}
}

// Get handles GET /metrics?period=7d|30d|24h.
func (h *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}

	resp, err := h.service.Get(r.Context(), auth.TeamID, period)
	if err != nil {
		pkg.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
