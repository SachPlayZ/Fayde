package sharing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Service handles share token business logic.
type Service struct {
	repo Repository
}

// NewService creates a new sharing Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateToken generates a new share token for a task (revokes any existing one).
func (s *Service) CreateToken(ctx context.Context, taskID string) (*ShareToken, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("sharing.CreateToken: generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	_ = s.repo.DeleteByTaskID(ctx, taskID) // ignore error — may not exist

	return s.repo.Create(ctx, taskID, token)
}

// GetByToken returns the share token record for a given token string.
func (s *Service) GetByToken(ctx context.Context, token string) (*ShareToken, error) {
	return s.repo.GetByToken(ctx, token)
}

// GetByTaskID returns the share token for a task, if any.
func (s *Service) GetByTaskID(ctx context.Context, taskID string) (*ShareToken, error) {
	return s.repo.GetByTaskID(ctx, taskID)
}

// RevokeToken deletes the share token for a task.
func (s *Service) RevokeToken(ctx context.Context, taskID string) error {
	return s.repo.DeleteByTaskID(ctx, taskID)
}
