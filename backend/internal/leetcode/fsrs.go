package leetcode

import (
	"time"

	gofsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// ScheduleReview applies the FSRS-5 algorithm to a card for the given rating.
// Pure function: no I/O, no DB access. card may be nil for a problem being
// reviewed for the first time, in which case a fresh card is used.
func ScheduleReview(card *CardState, rating int, now time.Time) (*CardState, *ReviewLog, error) {
	if rating < RatingAgain || rating > RatingEasy {
		return nil, nil, ErrInvalidRating
	}

	fc := toFSRSCard(card)
	params := gofsrs.DefaultParam()
	scheduler := params.NewBasicScheduler(fc, now)
	info := scheduler.Review(gofsrs.Rating(rating))

	updated := fromFSRSCard(info.Card)
	if card != nil {
		updated.ID = card.ID
		updated.ProblemID = card.ProblemID
		updated.UserID = card.UserID
	}

	log := &ReviewLog{
		Rating:        rating,
		ReviewedAt:    now,
		ScheduledDays: int(info.ReviewLog.ScheduledDays),
		ElapsedDays:   int(info.ReviewLog.ElapsedDays),
		Stability:     updated.Stability,
		Difficulty:    updated.Difficulty,
		State:         updated.State,
	}

	return updated, log, nil
}

func toFSRSCard(card *CardState) gofsrs.Card {
	if card == nil {
		return gofsrs.NewCard()
	}
	fc := gofsrs.Card{
		Due:           card.DueDate,
		Stability:     card.Stability,
		Difficulty:    card.Difficulty,
		ElapsedDays:   uint64(max(0, card.ElapsedDays)),
		ScheduledDays: uint64(max(0, card.ScheduledDays)),
		Reps:          uint64(max(0, card.Reps)),
		Lapses:        uint64(max(0, card.Lapses)),
		State:         stateToFSRS(card.State),
	}
	if card.LastReview != nil {
		fc.LastReview = *card.LastReview
	}
	return fc
}

func fromFSRSCard(fc gofsrs.Card) *CardState {
	lastReview := fc.LastReview
	return &CardState{
		Stability:     fc.Stability,
		Difficulty:    fc.Difficulty,
		ElapsedDays:   int(fc.ElapsedDays),
		ScheduledDays: int(fc.ScheduledDays),
		Reps:          int(fc.Reps),
		Lapses:        int(fc.Lapses),
		State:         stateFromFSRS(fc.State),
		DueDate:       fc.Due,
		LastReview:    &lastReview,
	}
}

func stateToFSRS(s string) gofsrs.State {
	switch s {
	case StateLearning:
		return gofsrs.Learning
	case StateReview:
		return gofsrs.Review
	case StateRelearning:
		return gofsrs.Relearning
	default:
		return gofsrs.New
	}
}

func stateFromFSRS(s gofsrs.State) string {
	switch s {
	case gofsrs.Learning:
		return StateLearning
	case gofsrs.Review:
		return StateReview
	case gofsrs.Relearning:
		return StateRelearning
	default:
		return StateNew
	}
}
