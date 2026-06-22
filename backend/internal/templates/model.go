package templates

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type TaskTemplate struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	EffortPoints *int      `json:"effort_points"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name         string `json:"name" validate:"required,min=1"`
	Title        string `json:"title" validate:"required,min=1"`
	Description  string `json:"description"`
	Status       string `json:"status" validate:"omitempty,oneof=todo in_progress done failed"`
	Priority     string `json:"priority" validate:"omitempty,oneof=low medium high"`
	EffortPoints *int   `json:"effort_points"`
}
