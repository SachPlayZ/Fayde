package savedfilters

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	filters, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list saved filters")
		return
	}
	httputil.JSON(w, http.StatusOK, filters)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateRequest
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
	sf, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create saved filter")
		return
	}
	httputil.JSON(w, http.StatusCreated, sf)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "saved filter not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete saved filter")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
