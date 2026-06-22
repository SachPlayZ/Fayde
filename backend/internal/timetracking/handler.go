package timetracking

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// POST /tasks/{id}/time/start
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// note is optional; ignore decode error
		req = StartRequest{}
	}
	e, err := h.svc.Start(r.Context(), taskID, userID, req.Note)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to start timer")
		return
	}
	httputil.JSON(w, http.StatusCreated, e)
}

// POST /tasks/{id}/time/stop/{entryId}
func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	entryID := chi.URLParam(r, "entryId")
	var req StopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = StopRequest{}
	}
	e, err := h.svc.Stop(r.Context(), entryID, userID, req.Note)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "time entry not found or already stopped")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to stop timer")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

// GET /tasks/{id}/time
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	entries, err := h.svc.List(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list time entries")
		return
	}
	httputil.JSON(w, http.StatusOK, entries)
}

// DELETE /tasks/{id}/time/{entryId}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	entryID := chi.URLParam(r, "entryId")
	if err := h.svc.Delete(r.Context(), entryID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "time entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete time entry")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /tasks/{id}/time/active
func (h *Handler) Active(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	e, err := h.svc.ActiveEntry(r.Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "no active time entry")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get active entry")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}
