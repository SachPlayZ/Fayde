package sharing

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type ShareToken struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// PublicTask is the read-only view returned for shared tasks.
type PublicTask struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
}
