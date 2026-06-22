package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler handles HTTP requests for GitHub link endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new GitHub Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /tasks/{id}/github.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")
	links, err := h.svc.List(r.Context(), taskID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to list links"})
		return
	}

	if links == nil {
		links = []*GitHubLink{}
	}
	httputil.JSON(w, 200, links)
}

// Link handles POST /tasks/{id}/github.
func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")

	var req LinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}
	if err := validate.Struct(req); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "repo is required"})
		return
	}

	link, err := h.svc.Link(r.Context(), taskID, req)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to create link"})
		return
	}

	httputil.JSON(w, 201, link)
}

// Unlink handles DELETE /tasks/{id}/github/{linkId}.
func (h *Handler) Unlink(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")
	linkID := chi.URLParam(r, "linkId")

	if err := h.svc.Unlink(r.Context(), linkID, taskID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "link not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to unlink"})
		return
	}

	httputil.JSON(w, 204, nil)
}

// Webhook handles POST /webhooks/github (public, no auth).
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "failed to read body"})
		return
	}

	sigHeader := r.Header.Get("X-Hub-Signature-256")
	if !verifyGitHubSignature([]byte(h.svc.webhookSecret), body, sigHeader) {
		httputil.JSON(w, 401, map[string]string{"error": "invalid signature"})
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event != "issues" && event != "pull_request" {
		// Not an event we handle; acknowledge and ignore
		httputil.JSON(w, 200, map[string]string{"message": "ignored"})
		return
	}

	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), payload); err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to handle webhook"})
		return
	}

	httputil.JSON(w, 200, map[string]string{"message": "ok"})
}

func verifyGitHubSignature(secret, body []byte, sigHeader string) bool {
	if len(secret) == 0 {
		return true
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sigHeader))
}
