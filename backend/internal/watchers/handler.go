package watchers

import (
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	watchers, err := h.svc.List(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list watchers")
		return
	}
	httputil.JSON(w, http.StatusOK, watchers)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	if err := h.svc.Add(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to add watcher")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	if err := h.svc.Remove(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to remove watcher")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	watching, err := h.svc.IsWatching(r.Context(), taskID, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to check watcher status")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"watching": watching})
}
