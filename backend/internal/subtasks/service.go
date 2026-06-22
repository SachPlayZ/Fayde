package subtasks

import (
	"context"
	"fmt"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, taskID string) ([]*Subtask, error) {
	return s.repo.List(ctx, taskID)
}

func (s *Service) Create(ctx context.Context, taskID, title string) (*Subtask, error) {
	if title == "" {
		return nil, fmt.Errorf("subtasks: title required")
	}
	return s.repo.Create(ctx, taskID, title)
}

func (s *Service) Update(ctx context.Context, id string, done bool, title string) (*Subtask, error) {
	return s.repo.Update(ctx, id, done, title)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Reorder(ctx context.Context, taskID string, ids []string) error {
	return s.repo.Reorder(ctx, taskID, ids)
}
