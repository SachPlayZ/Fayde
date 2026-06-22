package projects

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projects, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list projects")
		return
	}
	httputil.JSON(w, http.StatusOK, projects)
}

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
	p, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	httputil.JSON(w, http.StatusCreated, p)
}

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
	p, err := h.svc.Update(r.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "project not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update project")
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "project not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
