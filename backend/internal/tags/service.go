package tags

import "context"

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*Tag, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID, name, color string) (*Tag, error) {
	if color == "" {
		color = "#6366f1"
	}
	return s.repo.Create(ctx, userID, name, color)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) AddToTask(ctx context.Context, taskID, tagID string) error {
	return s.repo.AddToTask(ctx, taskID, tagID)
}

func (s *Service) RemoveFromTask(ctx context.Context, taskID, tagID string) error {
	return s.repo.RemoveFromTask(ctx, taskID, tagID)
}

func (s *Service) ListByTask(ctx context.Context, taskID string) ([]*Tag, error) {
	return s.repo.ListByTask(ctx, taskID)
}
