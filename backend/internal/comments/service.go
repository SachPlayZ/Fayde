package comments

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/notifications"
	"github.com/jackc/pgx/v5/pgxpool"
)

var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)

type Service struct {
	repo          Repository
	notifSvc      *notifications.Service
	pool          *pgxpool.Pool
}

func NewService(repo Repository, notifSvc *notifications.Service, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, notifSvc: notifSvc, pool: pool}
}

func (s *Service) List(ctx context.Context, taskID string) ([]*Comment, error) {
	return s.repo.List(ctx, taskID)
}

func (s *Service) Create(ctx context.Context, taskID, userID, body string) (*Comment, error) {
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("comments: body required")
	}
	c, err := s.repo.Create(ctx, taskID, userID, body)
	if err != nil {
		return nil, err
	}
	// Process @mentions
	for _, m := range mentionRe.FindAllStringSubmatch(body, -1) {
		email := m[1]
		if email == c.UserEmail {
			continue
		}
		mentionedUserID, err := s.lookupUserByEmail(ctx, email)
		if err != nil || mentionedUserID == "" {
			continue
		}
		msg := fmt.Sprintf("%s mentioned you in a comment", c.UserEmail)
		s.notifSvc.Create(ctx, mentionedUserID, "mention", &taskID, msg)
	}
	return c, nil
}

func (s *Service) Update(ctx context.Context, id, userID, body string) (*Comment, error) {
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("comments: body required")
	}
	return s.repo.Update(ctx, id, userID, body)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) lookupUserByEmail(ctx context.Context, email string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM users WHERE email=$1`, email).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}
