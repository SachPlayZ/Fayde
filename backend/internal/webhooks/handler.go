package webhooks

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler handles HTTP requests for outbound webhook endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new webhooks Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /settings/webhooks.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	hooks, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to list webhooks"})
		return
	}

	if hooks == nil {
		hooks = []*OutboundWebhook{}
	}
	httputil.JSON(w, 200, hooks)
}

// Create handles POST /settings/webhooks.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}
	if err := validate.Struct(req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "validation failed"})
		return
	}

	hook, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to create webhook"})
		return
	}

	httputil.JSON(w, 201, hook)
}

// Update handles PATCH /settings/webhooks/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	hook, err := h.svc.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "webhook not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to update webhook"})
		return
	}

	httputil.JSON(w, 200, hook)
}

// Delete handles DELETE /settings/webhooks/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "webhook not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to delete webhook"})
		return
	}

	httputil.JSON(w, 204, nil)
}
