package dependencies

import (
	"encoding/json"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	list, err := h.svc.GetDependencies(r.Context(), taskID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get dependencies")
		return
	}
	httputil.JSON(w, http.StatusOK, list)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	_ = userID
	taskID := chi.URLParam(r, "id")
	var body struct {
		DependsOnID string `json:"depends_on_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DependsOnID == "" {
		httputil.Error(w, http.StatusBadRequest, "depends_on_id required")
		return
	}
	if err := h.svc.Add(r.Context(), taskID, body.DependsOnID); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	depID := chi.URLParam(r, "depId")
	if err := h.svc.Remove(r.Context(), taskID, depID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to remove dependency")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
