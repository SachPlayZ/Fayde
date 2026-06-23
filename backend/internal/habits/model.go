package habits

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// Habit is a recurring practice tracked by daily/weekly completion.
type Habit struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Cadence         string    `json:"cadence"`
	TargetPerPeriod int       `json:"target_per_period"`
	Color           *string   `json:"color"`
	Position        float64   `json:"position"`
	Archived        bool      `json:"archived"`
	CreatedAt       time.Time `json:"created_at"`

	// Computed fields (populated on list/get).
	CurrentStreak int  `json:"current_streak"`
	LongestStreak int  `json:"longest_streak"`
	DoneToday     bool `json:"done_today"`
}

// Log is a single day's completion entry.
type Log struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int    `json:"count"`
}

type CreateRequest struct {
	Name            string  `json:"name"            validate:"required,min=1"`
	Cadence         string  `json:"cadence"`
	TargetPerPeriod int     `json:"target_per_period"`
	Color           *string `json:"color"`
}

type UpdateRequest struct {
	Name            *string  `json:"name"`
	Cadence         *string  `json:"cadence"`
	TargetPerPeriod *int     `json:"target_per_period"`
	Color           *string  `json:"color"`
	Position        *float64 `json:"position"`
	Archived        *bool    `json:"archived"`
}
