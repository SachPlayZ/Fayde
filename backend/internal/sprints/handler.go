package sprints

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

func validationFields(errs validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(errs))
	for _, e := range errs {
		fields[strings.ToLower(e.Field())] = e.Tag()
	}
	return fields
}

// GET /sprints
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	sprints, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list sprints")
		return
	}
	httputil.JSON(w, http.StatusOK, sprints)
}

// POST /sprints
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := h.validate.Struct(req); errs != nil {
		httputil.ValidationError(w, validationFields(errs.(validator.ValidationErrors)))
		return
	}
	sp, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create sprint")
		return
	}
	httputil.JSON(w, http.StatusCreated, sp)
}

// PATCH /sprints/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := h.validate.Struct(req); errs != nil {
		httputil.ValidationError(w, validationFields(errs.(validator.ValidationErrors)))
		return
	}
	sp, err := h.svc.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "sprint not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update sprint")
		return
	}
	httputil.JSON(w, http.StatusOK, sp)
}

// DELETE /sprints/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "sprint not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete sprint")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /sprints/{id}/tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ids, err := h.svc.ListTaskIDs(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list sprint tasks")
		return
	}
	httputil.JSON(w, http.StatusOK, ids)
}

// POST /sprints/{id}/tasks
func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req AddTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if errs := h.validate.Struct(req); errs != nil {
		httputil.ValidationError(w, validationFields(errs.(validator.ValidationErrors)))
		return
	}
	if err := h.svc.AddTask(r.Context(), id, req.TaskID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to add task to sprint")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /sprints/{id}/tasks/{taskId}
func (h *Handler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")
	if err := h.svc.RemoveTask(r.Context(), id, taskID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to remove task from sprint")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
