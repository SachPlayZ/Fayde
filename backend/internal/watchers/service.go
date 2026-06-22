package watchers

import (
	"context"
)

type NotificationsService interface {
	Create(ctx context.Context, userID, notifType string, taskID *string, message string)
}

type Service struct {
	repo     Repository
	notifSvc NotificationsService
}

func NewService(repo Repository, notifSvc NotificationsService) *Service {
	return &Service{repo: repo, notifSvc: notifSvc}
}

func (s *Service) Add(ctx context.Context, taskID, userID string) error {
	return s.repo.Add(ctx, taskID, userID)
}

func (s *Service) Remove(ctx context.Context, taskID, userID string) error {
	return s.repo.Remove(ctx, taskID, userID)
}

func (s *Service) List(ctx context.Context, taskID string) ([]*Watcher, error) {
	return s.repo.List(ctx, taskID)
}

func (s *Service) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	return s.repo.IsWatching(ctx, taskID, userID)
}

func (s *Service) NotifyWatchers(ctx context.Context, taskID, updaterUserID, taskTitle string) {
	watchers, err := s.repo.List(ctx, taskID)
	if err != nil {
		return
	}
	msg := "Task '" + taskTitle + "' was updated"
	for _, w := range watchers {
		if w.UserID == updaterUserID {
			continue
		}
		tid := taskID
		s.notifSvc.Create(ctx, w.UserID, "task_updated", &tid, msg)
	}
}
