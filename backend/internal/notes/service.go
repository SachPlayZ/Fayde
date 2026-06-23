package notes

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
)

// backlinkRe matches [[<uuid>]] references inside a note body.
var backlinkRe = regexp.MustCompile(`\[\[([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})\]\]`)

type Service struct {
	repo      Repository
	sseBroker *sse.Broker
}

func NewService(repo Repository, sseBroker *sse.Broker) *Service {
	return &Service{repo: repo, sseBroker: sseBroker}
}

func (s *Service) List(ctx context.Context, userID string) ([]*Note, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Note, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Note, error) {
	n, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	s.publish(userID, "note_created", n)
	return n, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Note, error) {
	n, err := s.repo.Update(ctx, id, userID, req)
	if err != nil {
		return nil, err
	}
	// Resolve [[id]] backlinks from the saved body.
	if req.Body != nil {
		dst := extractBacklinks(*req.Body)
		if err := s.repo.SetBacklinks(ctx, id, dst); err != nil {
			return nil, fmt.Errorf("notes: set backlinks: %w", err)
		}
	}
	s.publish(userID, "note_updated", n)
	return n, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return err
	}
	s.publish(userID, "note_deleted", map[string]string{"id": id})
	return nil
}

func (s *Service) Reorder(ctx context.Context, userID string, items []ReorderItem) error {
	return s.repo.Reorder(ctx, userID, items)
}

func (s *Service) LinkTask(ctx context.Context, noteID, taskID, userID string) error {
	return s.repo.LinkTask(ctx, noteID, taskID, userID)
}

func (s *Service) UnlinkTask(ctx context.Context, noteID, taskID, userID string) error {
	return s.repo.UnlinkTask(ctx, noteID, taskID, userID)
}

func (s *Service) ListTaskLinks(ctx context.Context, noteID, userID string) ([]string, error) {
	return s.repo.ListTaskLinks(ctx, noteID, userID)
}

func (s *Service) ListByTask(ctx context.Context, taskID, userID string) ([]*NoteRef, error) {
	return s.repo.ListByTask(ctx, taskID, userID)
}

func (s *Service) Backlinks(ctx context.Context, noteID, userID string) ([]*NoteRef, error) {
	return s.repo.Backlinks(ctx, noteID, userID)
}

func extractBacklinks(body string) []string {
	matches := backlinkRe.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

func (s *Service) publish(userID, event string, payload any) {
	if s.sseBroker != nil {
		s.sseBroker.Publish(userID, sse.Event{Type: event, Payload: payload})
	}
}
