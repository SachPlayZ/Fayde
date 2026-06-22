package tags

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
	userID := auth.UserIDFromContext(r.Context())
	tags, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list tags")
		return
	}
	httputil.JSON(w, http.StatusOK, tags)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "name required")
		return
	}
	tag, err := h.svc.Create(r.Context(), userID, body.Name, body.Color)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create tag")
		return
	}
	httputil.JSON(w, http.StatusCreated, tag)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete tag")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddToTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	var body struct {
		TagID string `json:"tag_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TagID == "" {
		httputil.Error(w, http.StatusBadRequest, "tag_id required")
		return
	}
	if err := h.svc.AddToTask(r.Context(), taskID, body.TagID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to add tag to task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveFromTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	tagID := chi.URLParam(r, "tagId")
	if err := h.svc.RemoveFromTask(r.Context(), taskID, tagID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to remove tag from task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
