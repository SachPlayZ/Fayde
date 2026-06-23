// Package inbox turns inbound emails (via Resend) into tasks.
package inbox

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// tokenRe pulls the inbox token out of an address local-part: u+<token>@... or <token>@...
var tokenRe = regexp.MustCompile(`(?:u\+)?([0-9a-f]{12})@`)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) UserIDByToken(ctx context.Context, token string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE inbox_token=$1`, token).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return id, err
}

// Handler receives Resend inbound webhooks.
type Handler struct {
	repo     *Repository
	tasksSvc *tasks.Service
}

func NewHandler(repo *Repository, tasksSvc *tasks.Service) *Handler {
	return &Handler{repo: repo, tasksSvc: tasksSvc}
}

// inboundPayload is a permissive view of Resend's inbound email event.
type inboundPayload struct {
	To      json.RawMessage `json:"to"`
	Subject string          `json:"subject"`
	Text    string          `json:"text"`
	Data    *struct {
		To      json.RawMessage `json:"to"`
		Subject string          `json:"subject"`
		Text    string          `json:"text"`
	} `json:"data"`
}

// Webhook handles POST /webhooks/email (public). Always returns 200 so the
// provider does not retry on unmatched mail.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	var p inboundPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	to, subject, text := p.To, p.Subject, p.Text
	if p.Data != nil {
		to, subject, text = p.Data.To, p.Data.Subject, p.Data.Text
	}

	token := extractToken(to)
	if token == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, err := h.repo.UserIDByToken(r.Context(), token)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	title := strings.TrimSpace(subject)
	if title == "" {
		title = "Untitled from email"
	}
	if _, err := h.tasksSvc.CreateTask(r.Context(), userID, tasks.CreateRequest{
		Title:       title,
		Description: strings.TrimSpace(text),
	}); err != nil {
		log.Printf("inbox: create task: %v", err)
	}
	w.WriteHeader(http.StatusOK)
}

// extractToken scans the `to` field (string or array) for an inbox token.
func extractToken(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var addrs []string
	if err := json.Unmarshal(raw, &addrs); err != nil {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			addrs = []string{single}
		}
	}
	for _, a := range addrs {
		if m := tokenRe.FindStringSubmatch(strings.ToLower(a)); m != nil {
			return m[1]
		}
	}
	return ""
}
