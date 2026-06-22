package savedfilters

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type SavedFilter struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	Params    json.RawMessage `json:"params"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateRequest struct {
	Name   string          `json:"name" validate:"required,min=1"`
	Params json.RawMessage `json:"params"`
}
