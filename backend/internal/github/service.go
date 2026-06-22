package github

import (
	"context"
	"fmt"
)

// TasksService is the interface used to update task status when GitHub events fire.
type TasksService interface {
	UpdateTaskStatus(ctx context.Context, taskID, userID, status string) error
}

// Service handles GitHub link business logic.
type Service struct {
	repo          Repository
	tasksSvc      TasksService
	webhookSecret string
}

// NewService creates a new GitHub Service.
func NewService(repo Repository, tasksSvc TasksService, webhookSecret string) *Service {
	return &Service{repo: repo, tasksSvc: tasksSvc, webhookSecret: webhookSecret}
}

// Link creates a GitHub link for a task.
func (s *Service) Link(ctx context.Context, taskID string, req LinkRequest) (*GitHubLink, error) {
	link, err := s.repo.Link(ctx, taskID, req)
	if err != nil {
		return nil, fmt.Errorf("github.service.Link: %w", err)
	}
	return link, nil
}

// Unlink removes a GitHub link.
func (s *Service) Unlink(ctx context.Context, id, taskID string) error {
	return s.repo.Unlink(ctx, id, taskID)
}

// List returns all GitHub links for a task.
func (s *Service) List(ctx context.Context, taskID string) ([]*GitHubLink, error) {
	return s.repo.List(ctx, taskID)
}

// HandleWebhook processes a GitHub webhook payload and auto-closes tasks.
func (s *Service) HandleWebhook(ctx context.Context, payload GitHubWebhookPayload) error {
	repo := payload.Repository.FullName

	// Issue closed
	if payload.Issue != nil && payload.Action == "closed" {
		links, _ := s.repo.FindByIssue(ctx, repo, payload.Issue.Number)
		for _, link := range links {
			_ = s.tasksSvc.UpdateTaskStatus(ctx, link.TaskID, "", "done")
		}
	}

	// PR merged
	if payload.PullRequest != nil && payload.Action == "closed" && payload.PullRequest.Merged {
		links, _ := s.repo.FindByPR(ctx, repo, payload.PullRequest.Number)
		for _, link := range links {
			_ = s.tasksSvc.UpdateTaskStatus(ctx, link.TaskID, "", "done")
		}
	}

	return nil
}
