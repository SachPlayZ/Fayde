package projects

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Project struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name        string `json:"name" validate:"required,min=1"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}
