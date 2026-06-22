package apitokens

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
	tokens, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list api tokens")
		return
	}
	httputil.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
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
	result, err := h.svc.Generate(r.Context(), userID, req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to generate api token")
		return
	}
	httputil.JSON(w, http.StatusCreated, result)
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
			httputil.Error(w, http.StatusNotFound, "api token not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete api token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
