package boards

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
)

// NotificationsService is the interface used by boards.Service to send notifications.
type NotificationsService interface {
	Create(ctx context.Context, userID, nType string, taskID *string, message string)
}

// Service handles board business logic.
type Service struct {
	repo      Repository
	notifSvc  NotificationsService
	sseBroker *sse.Broker
}

// NewService creates a new boards Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetNotificationsService wires in the notifications dependency post-construction.
func (s *Service) SetNotificationsService(notifSvc NotificationsService) {
	s.notifSvc = notifSvc
}

// SetSSEBroker wires in the SSE broker post-construction.
func (s *Service) SetSSEBroker(broker *sse.Broker) {
	s.sseBroker = broker
}

// CreateBoard creates a board and adds the owner as a member with role "owner".
func (s *Service) CreateBoard(ctx context.Context, ownerID string, input CreateBoardInput) (*Board, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("boards.CreateBoard: name is required")
	}
	b, err := s.repo.CreateBoard(ctx, ownerID, input.Name, input.Description)
	if err != nil {
		return nil, fmt.Errorf("boards.CreateBoard: %w", err)
	}
	if err := s.repo.AddMember(ctx, b.ID, ownerID, "owner"); err != nil {
		return nil, fmt.Errorf("boards.CreateBoard: add owner: %w", err)
	}
	return b, nil
}

// GetBoard returns board detail if the caller is a member.
func (s *Service) GetBoard(ctx context.Context, callerID, boardID string) (*BoardDetail, error) {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotMember
	}

	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, boardID)
	if err != nil {
		return nil, err
	}
	tasks, err := s.repo.ListTasks(ctx, boardID)
	if err != nil {
		return nil, err
	}
	completions, err := s.repo.ListTodayCompletions(ctx, boardID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if members == nil {
		members = []*BoardMember{}
	}
	if tasks == nil {
		tasks = []*BoardTask{}
	}
	if completions == nil {
		completions = []*Completion{}
	}

	inv, err := s.repo.GetInviteByBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}
	var shareToken *string
	if inv != nil {
		shareToken = &inv.Token
	}

	board.MemberCount = len(members)

	return &BoardDetail{
		Board:       *board,
		Members:     members,
		Tasks:       tasks,
		Completions: completions,
		ShareToken:  shareToken,
	}, nil
}

// UpdateBoard updates board name/description (owner only).
func (s *Service) UpdateBoard(ctx context.Context, callerID, boardID string, input UpdateBoardInput) error {
	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}
	if board.OwnerID != callerID {
		return ErrForbidden
	}

	name := board.Name
	description := board.Description
	if input.Name != nil {
		name = *input.Name
	}
	if input.Description != nil {
		description = *input.Description
	}

	return s.repo.UpdateBoard(ctx, boardID, name, description)
}

// DeleteBoard deletes a board (owner only).
func (s *Service) DeleteBoard(ctx context.Context, callerID, boardID string) error {
	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}
	if board.OwnerID != callerID {
		return ErrForbidden
	}
	return s.repo.DeleteBoard(ctx, boardID)
}

// ListBoards returns boards the caller is a member of.
func (s *Service) ListBoards(ctx context.Context, callerID string) ([]*Board, error) {
	return s.repo.ListBoardsByUser(ctx, callerID)
}

// AddTask adds a task to a board (any member).
func (s *Service) AddTask(ctx context.Context, callerID, boardID string, input AddTaskInput) (*BoardTask, error) {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotMember
	}
	if input.Title == "" {
		return nil, fmt.Errorf("boards.AddTask: title is required")
	}
	return s.repo.AddTask(ctx, boardID, input.Title, callerID)
}

// DeleteTask removes a board task (owner or creator only).
func (s *Service) DeleteTask(ctx context.Context, callerID, boardID, taskID string) error {
	board, err := s.repo.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.BoardID != boardID {
		return ErrNotFound
	}
	if board.OwnerID != callerID && task.CreatedBy != callerID {
		return ErrForbidden
	}
	return s.repo.DeleteTask(ctx, taskID)
}

// Complete marks a board task as done for today for the caller. Notifies other members.
func (s *Service) Complete(ctx context.Context, callerID, boardID, taskID string) (*Completion, error) {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotMember
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.BoardID != boardID {
		return nil, ErrNotFound
	}

	today := time.Now().UTC()
	c, err := s.repo.Complete(ctx, taskID, callerID, today)
	if err != nil {
		return nil, fmt.Errorf("boards.Complete: %w", err)
	}

	// Notify other members via SSE + notifications (async, non-blocking).
	go s.notifyCompletion(context.Background(), callerID, boardID, taskID, task.Title, today.Format("2006-01-02"))

	return c, nil
}

func (s *Service) notifyCompletion(ctx context.Context, callerID, boardID, taskID, taskTitle, date string) {
	displayName, _ := s.repo.GetUserDisplay(ctx, callerID)
	if displayName == "" {
		displayName = "A member"
	}

	members, err := s.repo.ListMembers(ctx, boardID)
	if err != nil {
		return
	}

	payload := SSEBoardTaskCompleted{
		BoardID:     boardID,
		BoardTaskID: taskID,
		UserID:      callerID,
		DisplayName: displayName,
		TaskTitle:   taskTitle,
		Date:        date,
	}

	msg := fmt.Sprintf("%s completed '%s'", displayName, taskTitle)

	for _, m := range members {
		if m.UserID == callerID {
			continue
		}
		if s.sseBroker != nil {
			s.sseBroker.Publish(m.UserID, sse.Event{Type: "board_task_completed", Payload: payload})
		}
		if s.notifSvc != nil {
			s.notifSvc.Create(ctx, m.UserID, "board_task_completed", nil, msg)
		}
	}
}

// Uncomplete removes today's completion for a board task (for the caller only).
func (s *Service) Uncomplete(ctx context.Context, callerID, boardID, taskID string) error {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task.BoardID != boardID {
		return ErrNotFound
	}

	return s.repo.Uncomplete(ctx, taskID, callerID, time.Now().UTC())
}

// InviteFriend directly adds a friend to a board. Caller must be a member; target must be a friend of caller.
func (s *Service) InviteFriend(ctx context.Context, callerID, boardID, friendID string) error {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}

	friends, err := s.repo.AreFriends(ctx, callerID, friendID)
	if err != nil {
		return err
	}
	if !friends {
		return ErrNotFriends
	}

	already, err := s.repo.IsMember(ctx, boardID, friendID)
	if err != nil {
		return err
	}
	if already {
		return ErrAlreadyMember
	}

	return s.repo.AddMember(ctx, boardID, friendID, "member")
}

// CreateShareToken generates (or rotates) the board's invite token.
func (s *Service) CreateShareToken(ctx context.Context, callerID, boardID string) (*BoardInvite, error) {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotMember
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("boards.CreateShareToken: generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	return s.repo.UpsertInvite(ctx, boardID, token, callerID)
}

// RevokeShareToken deletes the board's invite token (caller must be member).
func (s *Service) RevokeShareToken(ctx context.Context, callerID, boardID string) error {
	ok, err := s.repo.IsMember(ctx, boardID, callerID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotMember
	}
	return s.repo.DeleteInvite(ctx, boardID)
}

// JoinPreview returns public info about a board via its invite token (no auth required).
func (s *Service) JoinPreview(ctx context.Context, token string) (*JoinPreview, error) {
	inv, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	board, err := s.repo.GetBoard(ctx, inv.BoardID)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, inv.BoardID)
	if err != nil {
		return nil, err
	}

	return &JoinPreview{
		BoardName:   board.Name,
		MemberCount: len(members),
	}, nil
}

// JoinViaToken adds an authenticated user to the board referenced by the token.
func (s *Service) JoinViaToken(ctx context.Context, callerID, token string) (*Board, error) {
	inv, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	already, err := s.repo.IsMember(ctx, inv.BoardID, callerID)
	if err != nil {
		return nil, err
	}
	if already {
		return s.repo.GetBoard(ctx, inv.BoardID)
	}

	if err := s.repo.AddMember(ctx, inv.BoardID, callerID, "member"); err != nil {
		return nil, fmt.Errorf("boards.JoinViaToken: add member: %w", err)
	}

	return s.repo.GetBoard(ctx, inv.BoardID)
}
