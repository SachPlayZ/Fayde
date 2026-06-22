package totp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler handles HTTP requests for TOTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new TOTP Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type codeRequest struct {
	Code string `json:"code" validate:"required"`
}

// Setup handles POST /auth/totp/setup.
func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	// We need the user email for the TOTP issuer; fetch from context claims or use userID as fallback.
	// The auth package stores email in JWT claims; use userID as account name if email unavailable.
	userEmail := userID // fallback

	resp, err := h.svc.GenerateSecret(r.Context(), userID, userEmail)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to generate TOTP secret"})
		return
	}

	httputil.JSON(w, 201, resp)
}

// Enable handles POST /auth/totp/enable.
func (h *Handler) Enable(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}
	if err := validate.Struct(req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "code is required"})
		return
	}

	if err := h.svc.Enable(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "TOTP not set up"})
			return
		}
		if errors.Is(err, ErrInvalidCode) {
			httputil.JSON(w, 400, map[string]string{"error": "invalid TOTP code"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to enable TOTP"})
		return
	}

	httputil.JSON(w, 200, map[string]string{"message": "TOTP enabled"})
}

// Disable handles POST /auth/totp/disable.
func (h *Handler) Disable(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}
	if err := validate.Struct(req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "code is required"})
		return
	}

	if err := h.svc.Disable(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "TOTP not set up"})
			return
		}
		if errors.Is(err, ErrInvalidCode) {
			httputil.JSON(w, 400, map[string]string{"error": "invalid TOTP code"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to disable TOTP"})
		return
	}

	httputil.JSON(w, 200, map[string]string{"message": "TOTP disabled"})
}

// Status handles GET /auth/totp/status.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	enabled, err := h.svc.IsTOTPEnabled(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to get TOTP status"})
		return
	}

	httputil.JSON(w, 200, map[string]bool{"enabled": enabled})
}
