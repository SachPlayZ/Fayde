package friends

import (
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("request already exists")
	ErrForbidden      = errors.New("forbidden")
	ErrCannotSelfAdd  = errors.New("cannot send request to yourself")
)

// Friendship represents a friendship request/relationship between two users.
type Friendship struct {
	ID          string     `json:"id"`
	RequesterID string     `json:"requester_id"`
	AddresseeID string     `json:"addressee_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// FriendUser is a user record suitable for friend-list responses.
type FriendUser struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// FriendRequest is a pending friendship request enriched with the other party's info.
type FriendRequest struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Direction   string     `json:"direction"` // "incoming" or "outgoing"
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	User        FriendUser `json:"user"` // the other party
}

// SendRequestInput holds validated input for sending a friend request.
type SendRequestInput struct {
	AddresseeEmail string `json:"addressee_email" validate:"required,email"`
}
