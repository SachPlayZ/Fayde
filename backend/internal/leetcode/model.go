package leetcode

import (
	"errors"
	"time"
)

const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"

	StateNew        = "new"
	StateLearning   = "learning"
	StateReview     = "review"
	StateRelearning = "relearning"

	RatingAgain = 1
	RatingHard  = 2
	RatingGood  = 3
	RatingEasy  = 4
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidRating = errors.New("rating must be between 1 and 4")
)

// Problem is a tracked LeetCode problem, with its FSRS card state
// populated at query time (not persisted on this struct).
type Problem struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	LCNumber   *int       `json:"lc_number"`
	Slug       *string    `json:"slug"`
	Title      string     `json:"title"`
	URL        *string    `json:"url"`
	Difficulty string     `json:"difficulty"`
	Topics     []string   `json:"topics"`
	Notes      string     `json:"notes"`
	SolvedAt   *time.Time `json:"solved_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	Card *CardState `json:"card,omitempty"`
}

// CardState is the FSRS scheduling state for a single problem's review card.
type CardState struct {
	ID            string     `json:"id"`
	ProblemID     string     `json:"problem_id"`
	UserID        string     `json:"user_id"`
	Stability     float64    `json:"stability"`
	Difficulty    float64    `json:"difficulty"`
	ElapsedDays   int        `json:"elapsed_days"`
	ScheduledDays int        `json:"scheduled_days"`
	Reps          int        `json:"reps"`
	Lapses        int        `json:"lapses"`
	State         string     `json:"card_state"`
	DueDate       time.Time  `json:"due_date"`
	LastReview    *time.Time `json:"last_review"`
}

// ReviewLog is a single review event, recording FSRS state after the review.
type ReviewLog struct {
	ID            string    `json:"id"`
	CardID        string    `json:"card_id"`
	UserID        string    `json:"user_id"`
	ReviewedAt    time.Time `json:"reviewed_at"`
	Rating        int       `json:"rating"`
	ScheduledDays int       `json:"scheduled_days"`
	ElapsedDays   int       `json:"elapsed_days"`
	Stability     float64   `json:"stability"`
	Difficulty    float64   `json:"difficulty"`
	State         string    `json:"card_state"`
}

// ProblemDetail is a problem with its full review history.
type ProblemDetail struct {
	*Problem
	Reviews []*ReviewLog `json:"reviews"`
}

type CreateRequest struct {
	LCNumber   *int     `json:"lc_number"`
	Slug       *string  `json:"slug"`
	Title      string   `json:"title" validate:"required,min=1"`
	URL        *string  `json:"url"`
	Difficulty string   `json:"difficulty"`
	Topics     []string `json:"topics"`
	Notes      string   `json:"notes"`
}

type UpdateRequest struct {
	LCNumber   *int      `json:"lc_number"`
	Slug       *string   `json:"slug"`
	Title      *string   `json:"title"`
	URL        *string   `json:"url"`
	Difficulty *string   `json:"difficulty"`
	Topics     *[]string `json:"topics"`
	Notes      *string   `json:"notes"`
}

type ReviewRequest struct {
	Rating int `json:"rating"`
}

// ListFilter narrows the problem list. Status accepts "due", "overdue", or "" (all).
type ListFilter struct {
	Topic      string
	Difficulty string
	Status     string
}

type Stats struct {
	TotalProblems int     `json:"total_problems"`
	TotalSolved   int     `json:"total_solved"`
	DueTodayCount int     `json:"due_today_count"`
	OverdueCount  int     `json:"overdue_count"`
	TotalReviews  int     `json:"total_reviews"`
	ReviewStreak  int     `json:"review_streak"`
	RetentionRate float64 `json:"retention_rate"`
}
