package subtasks

import (
	"encoding/json"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc      *Service
	tasksSvc *tasks.Service
}

func NewHandler(svc *Service, tasksSvc *tasks.Service) *Handler {
	return &Handler{svc: svc, tasksSvc: tasksSvc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusNotFound, "task not found")
		return
	}
	items, err := h.svc.List(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list subtasks")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusNotFound, "task not found")
		return
	}
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title required")
		return
	}
	s, err := h.svc.Create(r.Context(), taskID, body.Title)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create subtask")
		return
	}
	httputil.JSON(w, http.StatusCreated, s)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	subID := chi.URLParam(r, "subId")
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusNotFound, "task not found")
		return
	}
	var body struct {
		Done  bool   `json:"done"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	s, err := h.svc.Update(r.Context(), subID, body.Done, body.Title)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update subtask")
		return
	}
	httputil.JSON(w, http.StatusOK, s)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	subID := chi.URLParam(r, "subId")
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusNotFound, "task not found")
		return
	}
	if err := h.svc.Delete(r.Context(), subID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete subtask")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		httputil.Error(w, http.StatusNotFound, "task not found")
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.Reorder(r.Context(), taskID, body.IDs); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to reorder subtasks")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
