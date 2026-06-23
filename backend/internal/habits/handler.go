package habits

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
		httputil.Error(w, http.StatusInternalServerError, "failed to list habits")
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
	hb, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create habit")
		return
	}
	httputil.JSON(w, http.StatusCreated, hb)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	hb, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update habit")
		return
	}
	httputil.JSON(w, http.StatusOK, hb)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete habit")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body struct {
		Date string `json:"date"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	done, err := h.svc.Toggle(r.Context(), chi.URLParam(r, "id"), userID, body.Date)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to toggle habit")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]bool{"done": done})
}

func (h *Handler) Logs(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	if from == "" {
		from = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}
	logs, err := h.svc.Logs(r.Context(), chi.URLParam(r, "id"), userID, from, to)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get logs")
		return
	}
	httputil.JSON(w, http.StatusOK, logs)
}
