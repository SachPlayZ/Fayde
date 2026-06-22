package templates

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

// GET /templates
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	templates, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list templates")
		return
	}
	httputil.JSON(w, http.StatusOK, templates)
}

// POST /templates
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
	t, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create template")
		return
	}
	httputil.JSON(w, http.StatusCreated, t)
}

// GET /templates/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	t, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "template not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get template")
		return
	}
	httputil.JSON(w, http.StatusOK, t)
}

// DELETE /templates/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.Error(w, http.StatusNotFound, "template not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete template")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
