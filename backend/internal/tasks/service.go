package tasks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/activitylog"
	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Service handles business logic for task operations.
type Service struct {
	repo        Repository
	activitySvc *activitylog.Service
	sseBroker   *sse.Broker
}

// NewService creates a new tasks Service.
func NewService(repo Repository, activitySvc *activitylog.Service, sseBroker *sse.Broker) *Service {
	return &Service{repo: repo, activitySvc: activitySvc, sseBroker: sseBroker}
}

// CreateTask creates a new task for the given user.
func (s *Service) CreateTask(ctx context.Context, userID string, req CreateRequest) (*Task, error) {
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("service: validate: %w", err)
	}
	task, err := s.repo.CreateTask(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("service: create task: %w", err)
	}

	if logErr := s.activitySvc.Log(ctx, task.ID, userID, "created", nil); logErr != nil {
		log.Printf("activitylog: create task %s: %v", task.ID, logErr)
	}

	s.sseBroker.Publish(userID, sse.Event{Type: "task.created", Payload: task})

	return task, nil
}

// ListTasks returns a paginated, filtered list of tasks for the given user.
func (s *Service) ListTasks(ctx context.Context, userID string, p ListParams) (*ListResult, error) {
	tasks, total, err := s.repo.ListTasks(ctx, userID, p)
	if err != nil {
		return nil, fmt.Errorf("service: list tasks: %w", err)
	}
	if tasks == nil {
		tasks = []*Task{}
	}
	return &ListResult{
		Data:  tasks,
		Page:  p.Page,
		Limit: p.Limit,
		Total: total,
	}, nil
}

// GetTask returns a task by ID, scoped to the given user.
func (s *Service) GetTask(ctx context.Context, id, userID string) (*Task, error) {
	task, err := s.repo.GetTask(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get task: %w", err)
	}
	return task, nil
}

// UpdateTask applies a partial update to a task and logs changed fields.
func (s *Service) UpdateTask(ctx context.Context, id, userID string, req UpdateRequest) (*Task, error) {
	// Fetch old task to compute diff.
	old, err := s.repo.GetTask(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("service: get task for update: %w", err)
	}

	task, err := s.repo.UpdateTask(ctx, id, userID, req)
	if err != nil {
		return nil, fmt.Errorf("service: update task: %w", err)
	}

	changes := buildChanges(old, req)
	if len(changes) > 0 {
		if logErr := s.activitySvc.Log(ctx, task.ID, userID, "updated", changes); logErr != nil {
			log.Printf("activitylog: update task %s: %v", task.ID, logErr)
		}
	}

	s.sseBroker.Publish(userID, sse.Event{Type: "task.updated", Payload: task})

	return task, nil
}

// DeleteTask removes a task owned by the given user.
func (s *Service) DeleteTask(ctx context.Context, id, userID string) error {
	if err := s.repo.DeleteTask(ctx, id, userID); err != nil {
		return fmt.Errorf("service: delete task: %w", err)
	}

	if logErr := s.activitySvc.Log(ctx, id, userID, "deleted", nil); logErr != nil {
		log.Printf("activitylog: delete task %s: %v", id, logErr)
	}

	s.sseBroker.Publish(userID, sse.Event{Type: "task.deleted", Payload: map[string]string{"id": id}})

	return nil
}

// buildChanges computes a map of changed fields from old task and update request.
// Each entry is [oldValue, newValue].
func buildChanges(old *Task, req UpdateRequest) map[string][2]interface{} {
	changes := make(map[string][2]interface{})

	if req.Title != nil && *req.Title != old.Title {
		changes["title"] = [2]interface{}{old.Title, *req.Title}
	}
	if req.Description != nil && *req.Description != old.Description {
		changes["description"] = [2]interface{}{old.Description, *req.Description}
	}
	if req.Status != nil && *req.Status != old.Status {
		changes["status"] = [2]interface{}{old.Status, *req.Status}
	}
	if req.Priority != nil && *req.Priority != old.Priority {
		changes["priority"] = [2]interface{}{old.Priority, *req.Priority}
	}
	if req.DueDate != nil {
		oldDue := formatDueDate(old.DueDate)
		newDue := formatDueDate(req.DueDate)
		if oldDue != newDue {
			changes["due_date"] = [2]interface{}{oldDue, newDue}
		}
	}

	return changes
}

func formatDueDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
