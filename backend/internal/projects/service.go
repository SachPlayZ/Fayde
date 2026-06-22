package projects

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Project, error) {
	p, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("projects.service.create: %w", err)
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*Project, error) {
	projects, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("projects.service.list: %w", err)
	}
	if projects == nil {
		projects = []*Project{}
	}
	return projects, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Project, error) {
	p, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("projects.service.get: %w", err)
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Project, error) {
	p, err := s.repo.Update(ctx, id, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("projects.service.update: %w", err)
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return fmt.Errorf("projects.service.delete: %w", err)
	}
	return nil
}
