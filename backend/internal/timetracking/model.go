package timetracking

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type TimeEntry struct {
	ID              string     `json:"id"`
	TaskID          string     `json:"task_id"`
	UserID          string     `json:"user_id"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationSeconds *int       `json:"duration_seconds"`
	Note            string     `json:"note"`
	CreatedAt       time.Time  `json:"created_at"`
}

type StartRequest struct {
	Note string `json:"note"`
}

type StopRequest struct {
	Note string `json:"note"`
}
