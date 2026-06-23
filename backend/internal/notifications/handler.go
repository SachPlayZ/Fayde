package notifications

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	unread := r.URL.Query().Get("unread") == "true"
	items, err := h.svc.ListByUser(r.Context(), userID, unread)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.MarkRead(r.Context(), id, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to mark read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if err := h.svc.MarkAllRead(r.Context(), userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to mark all read")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Snooze(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		Until time.Time `json:"until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Until.IsZero() {
		httputil.Error(w, http.StatusBadRequest, "until required")
		return
	}
	if err := h.svc.Snooze(r.Context(), id, userID, body.Until); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to snooze")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	count, err := h.svc.UnreadCount(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]int{"count": count})
}
