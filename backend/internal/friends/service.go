package friends

import (
	"context"
	"errors"
	"fmt"
)

// Service implements friendship business logic.
type Service struct {
	repo Repository
}

// NewService creates a new friends Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SendRequest creates a pending friend request from requester to the user with addresseeEmail.
func (s *Service) SendRequest(ctx context.Context, requesterID, addresseeEmail string) (*Friendship, error) {
	addressee, err := s.repo.UserByEmail(ctx, addresseeEmail)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("friends.SendRequest: lookup addressee: %w", err)
	}

	if addressee.ID == requesterID {
		return nil, ErrCannotSelfAdd
	}

	// Check for an existing relationship in either direction.
	existing, err := s.repo.GetBetween(ctx, requesterID, addressee.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("friends.SendRequest: check existing: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyExists
	}

	return s.repo.Create(ctx, requesterID, addressee.ID)
}

// AcceptRequest accepts a pending friend request addressed to callerID.
func (s *Service) AcceptRequest(ctx context.Context, callerID, requestID string) error {
	f, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if f.AddresseeID != callerID {
		return ErrForbidden
	}
	if f.Status != "pending" {
		return fmt.Errorf("friends.AcceptRequest: request is not pending (status=%s)", f.Status)
	}
	return s.repo.UpdateStatus(ctx, requestID, "accepted")
}

// DeclineRequest declines a pending friend request addressed to callerID.
func (s *Service) DeclineRequest(ctx context.Context, callerID, requestID string) error {
	f, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if f.AddresseeID != callerID {
		return ErrForbidden
	}
	if f.Status != "pending" {
		return fmt.Errorf("friends.DeclineRequest: request is not pending (status=%s)", f.Status)
	}
	return s.repo.UpdateStatus(ctx, requestID, "declined")
}

// Remove deletes a friendship or cancels/declines a request. Caller must be a party to the friendship.
func (s *Service) Remove(ctx context.Context, callerID, friendshipID string) error {
	f, err := s.repo.GetByID(ctx, friendshipID)
	if err != nil {
		return err
	}
	if f.RequesterID != callerID && f.AddresseeID != callerID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, friendshipID)
}

// ListFriends returns accepted friends for userID.
func (s *Service) ListFriends(ctx context.Context, userID string) ([]*FriendRequest, error) {
	return s.repo.ListFriends(ctx, userID)
}

// ListRequests returns pending incoming and outgoing requests for userID.
func (s *Service) ListRequests(ctx context.Context, userID string) ([]*FriendRequest, error) {
	return s.repo.ListRequests(ctx, userID)
}

// SearchUsers finds users matching the email query prefix.
func (s *Service) SearchUsers(ctx context.Context, q string) ([]*FriendUser, error) {
	if q == "" {
		return []*FriendUser{}, nil
	}
	return s.repo.SearchByEmail(ctx, q)
}
