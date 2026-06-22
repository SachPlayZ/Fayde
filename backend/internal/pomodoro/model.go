package pomodoro

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type Session struct {
	ID              string     `json:"id"`
	TaskID          *string    `json:"task_id"`
	UserID          string     `json:"user_id"`
	DurationMinutes int        `json:"duration_minutes"`
	Completed       bool       `json:"completed"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
}

type StartRequest struct {
	TaskID          *string `json:"task_id"`
	DurationMinutes int     `json:"duration_minutes"`
}
