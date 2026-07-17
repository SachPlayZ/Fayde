package leetcode

import (
	"testing"
	"time"
)

func TestScheduleReview_NewCard(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		rating    int
		wantState string
	}{
		{RatingAgain, StateLearning},
		{RatingHard, StateLearning},
		{RatingGood, StateLearning},
		{RatingEasy, StateReview},
	}

	for _, c := range cases {
		updated, log, err := ScheduleReview(nil, c.rating, now)
		if err != nil {
			t.Fatalf("rating=%d: unexpected error: %v", c.rating, err)
		}
		if updated.State != c.wantState {
			t.Errorf("rating=%d: state = %q, want %q", c.rating, updated.State, c.wantState)
		}
		if !updated.DueDate.After(now) {
			t.Errorf("rating=%d: due date %v not after now %v", c.rating, updated.DueDate, now)
		}
		if log.Rating != c.rating {
			t.Errorf("rating=%d: log.Rating = %d", c.rating, log.Rating)
		}
		if updated.Stability <= 0 {
			t.Errorf("rating=%d: stability = %v, want > 0", c.rating, updated.Stability)
		}
	}
}

func TestScheduleReview_InvalidRating(t *testing.T) {
	if _, _, err := ScheduleReview(nil, 0, time.Now()); err != ErrInvalidRating {
		t.Errorf("rating=0: err = %v, want ErrInvalidRating", err)
	}
	if _, _, err := ScheduleReview(nil, 5, time.Now()); err != ErrInvalidRating {
		t.Errorf("rating=5: err = %v, want ErrInvalidRating", err)
	}
}

func TestScheduleReview_ReviewStateAgainIncrementsLapses(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	// Simulate a card already in Review state with some stability.
	card := &CardState{
		ID: "card-1", ProblemID: "p-1", UserID: "u-1",
		Stability: 10, Difficulty: 5, State: StateReview,
		DueDate: now, Lapses: 0,
	}
	lastReview := now.Add(-10 * 24 * time.Hour)
	card.LastReview = &lastReview

	updated, _, err := ScheduleReview(card, RatingAgain, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Lapses != 1 {
		t.Errorf("lapses = %d, want 1", updated.Lapses)
	}
	if updated.State != StateRelearning {
		t.Errorf("state = %q, want %q", updated.State, StateRelearning)
	}
}
