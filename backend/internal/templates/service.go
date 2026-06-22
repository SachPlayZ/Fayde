package templates

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*TaskTemplate, error) {
	t, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("templates.service.create: %w", err)
	}
	return t, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*TaskTemplate, error) {
	templates, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("templates.service.list: %w", err)
	}
	if templates == nil {
		templates = []*TaskTemplate{}
	}
	return templates, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*TaskTemplate, error) {
	t, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("templates.service.get: %w", err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return fmt.Errorf("templates.service.delete: %w", err)
	}
	return nil
}
