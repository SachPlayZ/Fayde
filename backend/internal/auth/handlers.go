package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for authentication endpoints.
type Handler struct {
	svc      *Service
	validate *validator.Validate
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

type signupRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Signup handles POST /auth/signup.
// On success returns 201 with a message — no token until email is verified.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		httputil.ValidationError(w, validationFields(errs.(validator.ValidationErrors)))
		return
	}

	if err := h.svc.Signup(r.Context(), strings.ToLower(req.Email), req.Password); err != nil {
		if isDuplicateEmailError(err) {
			httputil.Error(w, http.StatusConflict, "email already registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "signup failed")
		return
	}

	httputil.JSON(w, http.StatusCreated, map[string]string{"message": "verification email sent"})
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// unverifiedError is the JSON body returned when a login attempt is made by an unverified user.
type unverifiedError struct {
	Code    string `json:"code"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := h.validate.Struct(req); errs != nil {
		httputil.ValidationError(w, validationFields(errs.(validator.ValidationErrors)))
		return
	}

	result, err := h.svc.Login(r.Context(), strings.ToLower(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailNotVerified) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(unverifiedError{
				Code:    "unverified",
				Email:   req.Email,
				Message: "please verify your email before logging in",
			})
			return
		}
		if errors.Is(err, ErrOAuthAccount) {
			httputil.Error(w, http.StatusBadRequest, "this account uses social login — please sign in with Google or GitHub")
			return
		}
		httputil.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	httputil.JSON(w, http.StatusOK, authResponse{
		Token: result.Token,
		User:  PublicUser{ID: result.User.ID, Email: result.User.Email, Role: result.User.Role},
	})
}

type verifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmail handles POST /auth/verify-email.
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			httputil.Error(w, http.StatusGone, "verification link has expired")
			return
		}
		httputil.Error(w, http.StatusBadRequest, "invalid or already used verification link")
		return
	}

	httputil.JSON(w, http.StatusOK, authResponse{
		Token: result.Token,
		User:  PublicUser{ID: result.User.ID, Email: result.User.Email, Role: result.User.Role},
	})
}

type resendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendVerification handles POST /auth/resend-verification.
func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always 204 — don't leak whether email exists.
	_ = h.svc.ResendVerification(r.Context(), strings.ToLower(req.Email))
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /auth/me — returns the authenticated user's profile.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "user not found")
		return
	}

	httputil.JSON(w, http.StatusOK, PublicUser{
		ID: user.ID, Email: user.Email, Role: user.Role,
		Theme: user.Theme, DigestEnabled: user.DigestEnabled,
		NotifPrefs: user.NotifPrefs, ChatURL: user.ChatURL, ChatKind: user.ChatKind,
		InboxToken: user.InboxToken,
	})
}

// UpdatePreferences handles PATCH /auth/me/preferences.
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Theme         *string          `json:"theme"`
		DigestEnabled *bool            `json:"digest_enabled"`
		NotifPrefs    *json.RawMessage `json:"notif_prefs"`
		ChatURL       *string          `json:"notif_chat_url"`
		ChatKind      *string          `json:"notif_chat_kind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	prefs := Preferences{
		Theme:         body.Theme,
		DigestEnabled: body.DigestEnabled,
		NotifPrefs:    body.NotifPrefs,
		ChatURL:       body.ChatURL,
		ChatKind:      body.ChatKind,
	}
	if err := h.svc.UpdatePreferences(r.Context(), userID, prefs); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validationFields converts validator.ValidationErrors into a string map.
func validationFields(errs validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(errs))
	for _, e := range errs {
		fields[strings.ToLower(e.Field())] = e.Tag()
	}
	return fields
}

// isDuplicateEmailError reports whether the error originates from a unique constraint
// violation on the email column (Postgres error code 23505).
func isDuplicateEmailError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
