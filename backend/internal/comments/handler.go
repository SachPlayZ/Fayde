package comments

import (
	"encoding/json"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	items, err := h.svc.List(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list comments")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.svc.Create(r.Context(), taskID, userID, body.Body)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create comment")
		return
	}
	httputil.JSON(w, http.StatusCreated, c)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	cID := chi.URLParam(r, "cId")
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.svc.Update(r.Context(), cID, userID, body.Body)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update comment")
		return
	}
	httputil.JSON(w, http.StatusOK, c)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	cID := chi.URLParam(r, "cId")
	if err := h.svc.Delete(r.Context(), cID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
