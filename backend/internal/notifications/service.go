package notifications

import (
	"context"
	"log"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
)

type Service struct {
	repo      Repository
	sseBroker *sse.Broker
}

func NewService(repo Repository, sseBroker *sse.Broker) *Service {
	return &Service{repo: repo, sseBroker: sseBroker}
}

func (s *Service) Create(ctx context.Context, userID, nType string, taskID *string, message string) {
	n, err := s.repo.Create(ctx, userID, nType, taskID, message)
	if err != nil {
		log.Printf("notifications: create: %v", err)
		return
	}
	s.sseBroker.Publish(userID, sse.Event{Type: "notification", Payload: n})
}

func (s *Service) ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
	return s.repo.ListByUser(ctx, userID, unreadOnly)
}

func (s *Service) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *Service) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *Service) ExistsRecent(ctx context.Context, taskID, nType string, since time.Time) (bool, error) {
	return s.repo.ExistsRecent(ctx, taskID, nType, since)
}
