// Package inbox turns inbound emails (via Resend) into tasks.
package inbox

import (
	"context"
	"encoding/json"
	"fmt"
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
	apiKey   string
}

func NewHandler(repo *Repository, tasksSvc *tasks.Service, apiKey string) *Handler {
	return &Handler{repo: repo, tasksSvc: tasksSvc, apiKey: apiKey}
}

// inboundPayload is a permissive view of Resend's inbound email event.
type inboundPayload struct {
	To      json.RawMessage `json:"to"`
	Subject string          `json:"subject"`
	Text    string          `json:"text"`
	Data    *struct {
		EmailID string          `json:"email_id"`
		To      json.RawMessage `json:"to"`
		Subject string          `json:"subject"`
		Text    string          `json:"text"`
	} `json:"data"`
}

func fetchEmailBody(ctx context.Context, apiKey, emailID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.resend.com/emails/receiving/"+emailID, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("resend returned status %d", resp.StatusCode)
	}

	var res struct {
		Text string `json:"text"`
		Html string `json:"html"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.Text != "" {
		return res.Text, nil
	}
	return res.Html, nil
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
	emailID := ""
	if p.Data != nil {
		to = p.Data.To
		if p.Data.Subject != "" {
			subject = p.Data.Subject
		}
		if p.Data.Text != "" {
			text = p.Data.Text
		}
		emailID = p.Data.EmailID
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

	// Fetch body if we have emailID and apiKey
	if emailID != "" && h.apiKey != "" {
		if fetchedBody, err := fetchEmailBody(r.Context(), h.apiKey, emailID); err == nil && fetchedBody != "" {
			text = fetchedBody
		} else if err != nil {
			log.Printf("inbox: failed to fetch email body: %v", err)
		}
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
