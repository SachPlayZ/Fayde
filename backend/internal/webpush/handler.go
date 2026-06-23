package webpush

import (
	"encoding/json"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
)

// Handler exposes web push subscription endpoints.
type Handler struct{ svc *Service }

// NewHandler builds a webpush Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// PublicKey returns the VAPID public key the browser needs to subscribe.
func (h *Handler) PublicKey(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"public_key": h.svc.PublicKey()})
}

// Subscribe stores the browser's PushSubscription.
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid subscription")
		return
	}
	sub, err := h.svc.Subscribe(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to subscribe")
		return
	}
	httputil.JSON(w, http.StatusCreated, sub)
}

// Unsubscribe removes a subscription. Endpoint comes from the `endpoint` query
// param (DELETE has no body); falls back to a JSON body if present.
func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		var req struct {
			Endpoint string `json:"endpoint"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		endpoint = req.Endpoint
	}
	if endpoint == "" {
		httputil.Error(w, http.StatusBadRequest, "endpoint required")
		return
	}
	if err := h.svc.Unsubscribe(r.Context(), userID, endpoint); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to unsubscribe")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
