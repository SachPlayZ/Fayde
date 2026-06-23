package automations

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
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list automations")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name required")
		return
	}
	a, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create automation")
		return
	}
	httputil.JSON(w, http.StatusCreated, a)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	a, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "automation not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update automation")
		return
	}
	httputil.JSON(w, http.StatusOK, a)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "automation not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete automation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
