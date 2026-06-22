package sprints

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Sprint struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Goal      string    `json:"goal"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
	Goal      string `json:"goal"`
}

type UpdateRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	Goal      *string `json:"goal"`
}

type AddTaskRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}
