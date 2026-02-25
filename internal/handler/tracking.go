package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mailit-dev/mailit/internal/pkg"
	"github.com/mailit-dev/mailit/internal/service"
)

// 1x1 transparent GIF.
var transparentGIF, _ = base64.StdEncoding.DecodeString(
	"R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
)

type TrackingHandler struct {
	svc service.TrackingService
}

func NewTrackingHandler(svc service.TrackingService) *TrackingHandler {
	return &TrackingHandler{svc: svc}
}

// TrackOpen handles GET /track/open/{id} — returns a 1x1 transparent GIF
// and records an open event.
func (h *TrackingHandler) TrackOpen(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		// Return the pixel anyway to avoid broken images.
		h.servePixel(w)
		return
	}

	// Record the open event (fire-and-forget; don't let errors block the pixel).
	_ = h.svc.HandleOpen(r.Context(), id)

	h.servePixel(w)
}

// TrackClick handles GET /track/click/{id} — records a click event and
// redirects to the original URL.
func (h *TrackingHandler) TrackClick(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid tracking ID")
		return
	}

	originalURL, err := h.svc.HandleClick(r.Context(), id)
	if err != nil {
		pkg.Error(w, http.StatusNotFound, "tracking link not found")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// Unsubscribe handles POST /unsubscribe?token={id} — marks the contact as
// unsubscribed and returns a confirmation page.
func (h *TrackingHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		pkg.Error(w, http.StatusBadRequest, "missing token parameter")
		return
	}

	id, err := uuid.Parse(tokenStr)
	if err != nil {
		pkg.Error(w, http.StatusBadRequest, "invalid token")
		return
	}

	if err := h.svc.HandleUnsubscribe(r.Context(), id); err != nil {
		pkg.Error(w, http.StatusNotFound, "unsubscribe link not found or expired")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Unsubscribed</title></head>
<body style="font-family:sans-serif;text-align:center;padding:60px">
<h1>You have been unsubscribed</h1>
<p>You will no longer receive emails from this sender.</p>
</body></html>`))
}

func (h *TrackingHandler) servePixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(transparentGIF)
}
