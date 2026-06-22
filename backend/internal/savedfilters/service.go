package savedfilters

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*SavedFilter, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *Service) List(ctx context.Context, userID string) ([]*SavedFilter, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return err
	}
	return nil
}
