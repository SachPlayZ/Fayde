package apitokens

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type APIToken struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateResult returned once on creation — includes raw token.
type CreateResult struct {
	Token string `json:"token"`
	APIToken
}

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

// LookupResult used internally for token validation.
type LookupResult struct {
	ID     string
	UserID string
}
