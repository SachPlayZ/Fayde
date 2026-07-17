package leetcode

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
	q := r.URL.Query()
	f := ListFilter{
		Topic:      q.Get("topic"),
		Difficulty: q.Get("difficulty"),
		Status:     q.Get("status"),
	}
	items, err := h.svc.List(r.Context(), userID, f)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list problems")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	p, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "problem not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get problem")
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		httputil.Error(w, http.StatusBadRequest, "title required")
		return
	}
	p, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create problem")
		return
	}
	httputil.JSON(w, http.StatusCreated, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	p, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "problem not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update problem")
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "problem not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete problem")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	card, err := h.svc.Review(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrInvalidRating) {
		httputil.Error(w, http.StatusBadRequest, "rating must be between 1 and 4")
		return
	}
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "problem not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to submit review")
		return
	}
	httputil.JSON(w, http.StatusOK, card)
}

func (h *Handler) Queue(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	items, err := h.svc.Queue(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get review queue")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	stats, err := h.svc.Stats(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	httputil.JSON(w, http.StatusOK, stats)
}
