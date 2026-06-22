package dependencies

import (
	"context"
	"fmt"

	"github.com/SachPlayZ/rivz-asn/backend/internal/notifications"
)

type Service struct {
	repo          Repository
	notificationsSvc *notifications.Service
}

func NewService(repo Repository, notifSvc *notifications.Service) *Service {
	return &Service{repo: repo, notificationsSvc: notifSvc}
}

func (s *Service) Add(ctx context.Context, taskID, dependsOnID string) error {
	if taskID == dependsOnID {
		return fmt.Errorf("deps: task cannot depend on itself")
	}
	if err := s.detectCycle(ctx, taskID, dependsOnID); err != nil {
		return err
	}
	return s.repo.Add(ctx, taskID, dependsOnID)
}

func (s *Service) Remove(ctx context.Context, taskID, dependsOnID string) error {
	return s.repo.Remove(ctx, taskID, dependsOnID)
}

func (s *Service) GetDependencies(ctx context.Context, taskID string) (*DependencyList, error) {
	blockedBy, err := s.repo.ListBlockedBy(ctx, taskID)
	if err != nil {
		return nil, err
	}
	blocking, err := s.repo.ListBlocking(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &DependencyList{BlockedBy: blockedBy, Blocking: blocking}, nil
}

// NotifyUnblocked checks if completing doneTaskID unblocks any dependent tasks
// and sends notifications for each.
func (s *Service) NotifyUnblocked(ctx context.Context, doneTaskID, ownerUserID string) {
	dependents, err := s.repo.ListDependentsOf(ctx, doneTaskID)
	if err != nil {
		return
	}
	for _, depID := range dependents {
		// Check if this dependent task is now fully unblocked (all its blockers done).
		blockers, err := s.repo.ListBlockedBy(ctx, depID)
		if err != nil {
			continue
		}
		allDone := true
		for _, b := range blockers {
			if b.DependsOnID != doneTaskID {
				allDone = false
				break
			}
		}
		if allDone {
			msg := fmt.Sprintf("Task is now unblocked")
			s.notificationsSvc.Create(ctx, ownerUserID, "dependency_unblocked", &depID, msg)
		}
	}
}

// detectCycle runs a BFS from dependsOnID and checks if taskID is reachable —
// which would mean adding this edge creates a cycle.
func (s *Service) detectCycle(ctx context.Context, taskID, dependsOnID string) error {
	all, err := s.repo.GetAllDependencies(ctx)
	if err != nil {
		return err
	}
	// Build adjacency: task_id → depends_on_id
	adj := make(map[string][]string)
	for _, pair := range all {
		adj[pair[0]] = append(adj[pair[0]], pair[1])
	}
	// BFS from dependsOnID following existing edges; if we reach taskID it's a cycle.
	visited := map[string]bool{}
	queue := []string{dependsOnID}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr == taskID {
			return fmt.Errorf("deps: would create cycle")
		}
		if visited[curr] {
			continue
		}
		visited[curr] = true
		queue = append(queue, adj[curr]...)
	}
	return nil
}
