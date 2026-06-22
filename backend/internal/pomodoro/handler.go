package pomodoro

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for pomodoro endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new pomodoro Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Start handles POST /pomodoro/start.
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	session, err := h.svc.Start(r.Context(), userID, req)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to start session"})
		return
	}

	httputil.JSON(w, 201, session)
}

// Complete handles POST /pomodoro/{id}/complete.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")

	session, err := h.svc.Complete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "session not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to complete session"})
		return
	}

	httputil.JSON(w, 200, session)
}

// Abandon handles POST /pomodoro/{id}/abandon.
func (h *Handler) Abandon(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	id := chi.URLParam(r, "id")

	session, err := h.svc.Abandon(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "session not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to abandon session"})
		return
	}

	httputil.JSON(w, 200, session)
}

// History handles GET /pomodoro/history.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	sessions, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to list sessions"})
		return
	}

	if sessions == nil {
		sessions = []*Session{}
	}
	httputil.JSON(w, 200, sessions)
}

// Active handles GET /pomodoro/active.
func (h *Handler) Active(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	session, err := h.svc.ActiveSession(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "no active session"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to get active session"})
		return
	}

	httputil.JSON(w, 200, session)
}
