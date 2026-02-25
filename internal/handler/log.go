package handler

import (
	"net/http"

	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/server/middleware"
	"github.com/mailit-dev/mailit/internal/service"
)

type LogHandler struct {
	service service.LogService
}

func NewLogHandler(s service.LogService) *LogHandler {
	return &LogHandler{service: s}
}

// List handles GET /logs.
func (h *LogHandler) List(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuth(r.Context())
	if auth == nil {
		pkg.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	params := parsePagination(r)
	level := r.URL.Query().Get("level")

	resp, err := h.service.List(r.Context(), auth.TeamID, level, &params)
	if err != nil {
		pkg.HandleError(w, err)
		return
	}
	pkg.JSON(w, http.StatusOK, resp)
}
