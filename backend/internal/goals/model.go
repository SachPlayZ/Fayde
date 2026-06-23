package goals

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// Goal is a top-level objective.
type Goal struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	TargetDate  *string      `json:"target_date"`
	ParentID    *string      `json:"parent_id"`
	Position    float64      `json:"position"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	KeyResults  []*KeyResult `json:"key_results"`
	Progress    float64      `json:"progress"` // 0..100, computed
}

// KeyResult is a measurable outcome under a goal.
type KeyResult struct {
	ID         string  `json:"id"`
	GoalID     string  `json:"goal_id"`
	Title      string  `json:"title"`
	MetricType string  `json:"metric_type"`
	CurrentVal float64 `json:"current_val"`
	TargetVal  float64 `json:"target_val"`
	Position   float64 `json:"position"`
}

type CreateGoalRequest struct {
	Title       string  `json:"title" validate:"required,min=1"`
	Description string  `json:"description"`
	TargetDate  *string `json:"target_date"`
	ParentID    *string `json:"parent_id"`
}

type UpdateGoalRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	TargetDate  **string `json:"target_date"`
	Position    *float64 `json:"position"`
}

type KRRequest struct {
	Title      string   `json:"title" validate:"required,min=1"`
	MetricType *string  `json:"metric_type"`
	CurrentVal *float64 `json:"current_val"`
	TargetVal  *float64 `json:"target_val"`
}
