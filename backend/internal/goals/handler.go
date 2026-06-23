package goals

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	gs, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list goals")
		return
	}
	httputil.JSON(w, http.StatusOK, gs)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	g, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get goal")
		return
	}
	httputil.JSON(w, http.StatusOK, g)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title required")
		return
	}
	g, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create goal")
		return
	}
	httputil.JSON(w, http.StatusCreated, g)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	g, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update goal")
		return
	}
	httputil.JSON(w, http.StatusOK, g)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete goal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddKR(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req KRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title required")
		return
	}
	kr, err := h.svc.AddKR(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "goal not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to add key result")
		return
	}
	httputil.JSON(w, http.StatusCreated, kr)
}

func (h *Handler) UpdateKR(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req KRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	kr, err := h.svc.UpdateKR(r.Context(), chi.URLParam(r, "krId"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "key result not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update key result")
		return
	}
	httputil.JSON(w, http.StatusOK, kr)
}

func (h *Handler) DeleteKR(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.DeleteKR(r.Context(), chi.URLParam(r, "krId"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "key result not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete key result")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	tasks, err := h.svc.ListTasks(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	httputil.JSON(w, http.StatusOK, tasks)
}

func (h *Handler) LinkTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.LinkTask(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "taskId"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "goal or task not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to link task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnlinkTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if err := h.svc.UnlinkTask(r.Context(), chi.URLParam(r, "taskId"), userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to unlink task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
