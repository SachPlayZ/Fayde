package leetcode

import (
	"context"
	"time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, userID string, f ListFilter) ([]*Problem, error) {
	return s.repo.List(ctx, userID, f)
}

func (s *Service) Get(ctx context.Context, id, userID string) (*ProblemDetail, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Problem, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Problem, error) {
	return s.repo.Update(ctx, id, userID, req)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

// Review applies the FSRS scheduling algorithm for a rating and persists the result.
func (s *Service) Review(ctx context.Context, problemID, userID string, req ReviewRequest) (*CardState, error) {
	if req.Rating < RatingAgain || req.Rating > RatingEasy {
		return nil, ErrInvalidRating
	}
	card, isSolved, err := s.repo.GetCard(ctx, problemID, userID)
	if err != nil {
		return nil, err
	}
	updated, log, err := ScheduleReview(card, req.Rating, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	log.UserID = userID
	markSolved := !isSolved && req.Rating >= RatingGood
	if err := s.repo.SubmitReview(ctx, updated, log, markSolved); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) Queue(ctx context.Context, userID string) ([]*Problem, error) {
	return s.repo.Queue(ctx, userID)
}

func (s *Service) DueTodayCount(ctx context.Context, userID string) (int, error) {
	return s.repo.DueTodayCount(ctx, userID)
}

// Stats assembles review stats, computing the current review streak in Go
// (same consecutive-day-walk approach as habits.streaks).
func (s *Service) Stats(ctx context.Context, userID string) (*Stats, error) {
	stats, err := s.repo.RawStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	dates, err := s.repo.ReviewDates(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats.ReviewStreak = reviewStreak(dates, time.Now().UTC().Truncate(24*time.Hour))
	return stats, nil
}

// reviewStreak returns the count of consecutive days (including today or
// yesterday) that had at least one review.
func reviewStreak(dates []time.Time, today time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	days := make([]time.Time, 0, len(dates))
	seen := map[string]bool{}
	for _, d := range dates {
		d = d.UTC().Truncate(24 * time.Hour)
		k := d.Format("2006-01-02")
		if !seen[k] {
			seen[k] = true
			days = append(days, d)
		}
	}
	yesterday := today.Add(-24 * time.Hour)
	if !days[0].Equal(today) && !days[0].Equal(yesterday) {
		return 0
	}
	streak := 1
	for i := 1; i < len(days); i++ {
		if days[i-1].Sub(days[i]) == 24*time.Hour {
			streak++
		} else {
			break
		}
	}
	return streak
}
