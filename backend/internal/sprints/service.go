package sprints

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Sprint, error) {
	sp, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("sprints.service.create: %w", err)
	}
	return sp, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*Sprint, error) {
	sprints, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("sprints.service.list: %w", err)
	}
	if sprints == nil {
		sprints = []*Sprint{}
	}
	return sprints, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Sprint, error) {
	sp, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("sprints.service.get: %w", err)
	}
	return sp, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Sprint, error) {
	sp, err := s.repo.Update(ctx, id, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("sprints.service.update: %w", err)
	}
	return sp, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return fmt.Errorf("sprints.service.delete: %w", err)
	}
	return nil
}

func (s *Service) AddTask(ctx context.Context, sprintID, taskID string) error {
	if err := s.repo.AddTask(ctx, sprintID, taskID); err != nil {
		return fmt.Errorf("sprints.service.add_task: %w", err)
	}
	return nil
}

func (s *Service) RemoveTask(ctx context.Context, sprintID, taskID string) error {
	if err := s.repo.RemoveTask(ctx, sprintID, taskID); err != nil {
		return fmt.Errorf("sprints.service.remove_task: %w", err)
	}
	return nil
}

func (s *Service) ListTaskIDs(ctx context.Context, sprintID string) ([]string, error) {
	ids, err := s.repo.ListTaskIDs(ctx, sprintID)
	if err != nil {
		return nil, fmt.Errorf("sprints.service.list_task_ids: %w", err)
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}
