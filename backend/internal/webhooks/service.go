package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Service handles outbound webhook business logic.
type Service struct {
	repo Repository
}

// NewService creates a new webhooks Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a new outbound webhook.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*OutboundWebhook, error) {
	hook, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("webhooks.service.Create: %w", err)
	}
	return hook, nil
}

// List returns all webhooks for the user.
func (s *Service) List(ctx context.Context, userID string) ([]*OutboundWebhook, error) {
	return s.repo.List(ctx, userID)
}

// Update updates an outbound webhook.
func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*OutboundWebhook, error) {
	return s.repo.Update(ctx, id, userID, req)
}

// Delete removes an outbound webhook.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

// Fire dispatches an event to all enabled webhooks for the user that subscribe to it.
func (s *Service) Fire(ctx context.Context, userID, event string, payload any) {
	hooks, err := s.repo.ListEnabledForUser(ctx, userID)
	if err != nil {
		return
	}

	for _, hook := range hooks {
		if !containsEvent(hook.Events, event) {
			continue
		}
		go func(h *OutboundWebhook) {
			body, _ := json.Marshal(map[string]any{
				"event":   event,
				"payload": payload,
			})

			req, _ := http.NewRequest("POST", h.URL, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Rivz-Event", event)

			if h.Secret != "" {
				mac := hmac.New(sha256.New, []byte(h.Secret))
				mac.Write(body)
				req.Header.Set("X-Rivz-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("webhook delivery failed to %s: %v", h.URL, err)
				return
			}
			defer resp.Body.Close()
		}(hook)
	}
}

func containsEvent(events []string, event string) bool {
	for _, e := range events {
		if e == event || e == "*" {
			return true
		}
	}
	return false
}
