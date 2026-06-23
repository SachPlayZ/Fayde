package webpush

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	wp "github.com/SherClockHolmes/webpush-go"
)

// Config holds the VAPID keypair and subject used to authenticate pushes.
type Config struct {
	PublicKey  string
	PrivateKey string
	Subject    string // mailto: or https: identifying the sender
}

// Enabled reports whether VAPID keys are configured.
func (c Config) Enabled() bool { return c.PublicKey != "" && c.PrivateKey != "" }

// Service manages subscriptions and sends web push notifications.
type Service struct {
	repo Repository
	cfg  Config
}

// NewService builds a webpush Service.
func NewService(repo Repository, cfg Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// PublicKey returns the VAPID public key for the browser to subscribe with.
func (s *Service) PublicKey() string { return s.cfg.PublicKey }

// Subscribe stores a browser subscription for a user.
func (s *Service) Subscribe(ctx context.Context, userID string, req SubscribeRequest) (*Subscription, error) {
	return s.repo.Subscribe(ctx, userID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth)
}

// Unsubscribe removes a browser subscription for a user.
func (s *Service) Unsubscribe(ctx context.Context, userID, endpoint string) error {
	return s.repo.Unsubscribe(ctx, userID, endpoint)
}

// SendToUser pushes a payload to every subscription a user owns.
// Dead subscriptions (404/410) are pruned. No-op if VAPID is not configured.
func (s *Service) SendToUser(ctx context.Context, userID string, p Payload) {
	if !s.cfg.Enabled() {
		return
	}
	subs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		log.Printf("webpush: list subs: %v", err)
		return
	}
	body, err := json.Marshal(p)
	if err != nil {
		log.Printf("webpush: marshal: %v", err)
		return
	}
	for _, sub := range subs {
		s.sendOne(ctx, sub, body)
	}
}

func (s *Service) sendOne(ctx context.Context, sub *Subscription, body []byte) {
	resp, err := wp.SendNotification(body, &wp.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     wp.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &wp.Options{
		Subscriber:      s.cfg.Subject,
		VAPIDPublicKey:  s.cfg.PublicKey,
		VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL:             86400,
	})
	if err != nil {
		log.Printf("webpush: send: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		if err := s.repo.DeleteByEndpoint(ctx, sub.Endpoint); err != nil {
			log.Printf("webpush: prune dead sub: %v", err)
		}
		return
	}
	if resp.StatusCode >= 400 {
		log.Printf("webpush: push service returned %d", resp.StatusCode)
	}
}
