package customfields

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateDef(ctx context.Context, userID string, req CreateDefRequest) (*FieldDefinition, error) {
	return s.repo.CreateDef(ctx, userID, req)
}

func (s *Service) ListDefs(ctx context.Context, userID string) ([]*FieldDefinition, error) {
	return s.repo.ListDefs(ctx, userID)
}

func (s *Service) GetDef(ctx context.Context, id, userID string) (*FieldDefinition, error) {
	d, err := s.repo.GetDef(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return d, nil
}

func (s *Service) DeleteDef(ctx context.Context, id, userID string) error {
	return s.repo.DeleteDef(ctx, id, userID)
}

func (s *Service) SetValue(ctx context.Context, taskID, fieldID, value string) (*FieldValue, error) {
	return s.repo.SetValue(ctx, taskID, fieldID, value)
}

func (s *Service) ListValues(ctx context.Context, taskID string) ([]*FieldValue, error) {
	return s.repo.ListValues(ctx, taskID)
}
