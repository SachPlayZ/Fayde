package pomodoro

import (
	"context"
	"fmt"
)

// Service handles pomodoro session business logic.
type Service struct {
	repo Repository
}

// NewService creates a new pomodoro Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Start begins a new pomodoro session. Defaults duration to 25 if 0.
func (s *Service) Start(ctx context.Context, userID string, req StartRequest) (*Session, error) {
	dur := req.DurationMinutes
	if dur <= 0 {
		dur = 25
	}
	session, err := s.repo.Start(ctx, userID, req.TaskID, dur)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.service.Start: %w", err)
	}
	return session, nil
}

// Complete marks a session as completed.
func (s *Service) Complete(ctx context.Context, id, userID string) (*Session, error) {
	return s.repo.Complete(ctx, id, userID)
}

// Abandon ends a session without completing it.
func (s *Service) Abandon(ctx context.Context, id, userID string) (*Session, error) {
	return s.repo.Abandon(ctx, id, userID)
}

// List returns the last 50 sessions for the user.
func (s *Service) List(ctx context.Context, userID string) ([]*Session, error) {
	return s.repo.List(ctx, userID)
}

// ActiveSession returns the current active session, if any.
func (s *Service) ActiveSession(ctx context.Context, userID string) (*Session, error) {
	return s.repo.ActiveSession(ctx, userID)
}
