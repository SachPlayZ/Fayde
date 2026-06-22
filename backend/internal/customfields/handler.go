package customfields

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) ListDefs(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	defs, err := h.svc.ListDefs(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list custom fields")
		return
	}
	httputil.JSON(w, http.StatusOK, defs)
}

func (h *Handler) CreateDef(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateDefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}
	def, err := h.svc.CreateDef(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create custom field")
		return
	}
	httputil.JSON(w, http.StatusCreated, def)
}

func (h *Handler) DeleteDef(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteDef(r.Context(), id, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete custom field")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListValues(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	values, err := h.svc.ListValues(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list custom field values")
		return
	}
	httputil.JSON(w, http.StatusOK, values)
}

func (h *Handler) SetValue(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID := chi.URLParam(r, "id")
	fieldID := chi.URLParam(r, "fieldId")

	// Verify field belongs to this user
	if _, err := h.svc.GetDef(r.Context(), fieldID, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "custom field not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to verify custom field")
		return
	}

	var req SetValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := validate.Struct(req); err != nil {
		fields := map[string]string{}
		for _, e := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
		httputil.ValidationError(w, fields)
		return
	}
	fv, err := h.svc.SetValue(r.Context(), taskID, fieldID, req.Value)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to set custom field value")
		return
	}
	httputil.JSON(w, http.StatusOK, fv)
}
