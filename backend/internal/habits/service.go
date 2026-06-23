package habits

import (
	"context"
	"time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// List returns the user's habits with streaks and today's completion computed.
func (s *Service) List(ctx context.Context, userID string) ([]*Habit, error) {
	hs, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for _, h := range hs {
		dates, err := s.repo.logDates(ctx, h.ID)
		if err != nil {
			continue
		}
		h.CurrentStreak, h.LongestStreak = streaks(dates, today)
		h.DoneToday = containsDay(dates, today)
	}
	return hs, nil
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Habit, error) {
	return s.repo.Create(ctx, userID, req)
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Habit, error) {
	return s.repo.Update(ctx, id, userID, req)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) Toggle(ctx context.Context, id, userID, date string) (bool, error) {
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	return s.repo.Toggle(ctx, id, userID, date)
}

func (s *Service) Logs(ctx context.Context, id, userID, from, to string) ([]*Log, error) {
	return s.repo.Logs(ctx, id, userID, from, to)
}

// streaks computes the current and longest consecutive-day streaks.
// `dates` is sorted descending. A current streak counts back from today or yesterday.
func streaks(dates []time.Time, today time.Time) (current, longest int) {
	if len(dates) == 0 {
		return 0, 0
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
	// days descending. Longest run of consecutive days.
	run := 1
	longest = 1
	for i := 1; i < len(days); i++ {
		if days[i-1].Sub(days[i]) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}
	// Current streak: must include today or yesterday.
	yesterday := today.Add(-24 * time.Hour)
	if !days[0].Equal(today) && !days[0].Equal(yesterday) {
		return 0, longest
	}
	current = 1
	for i := 1; i < len(days); i++ {
		if days[i-1].Sub(days[i]) == 24*time.Hour {
			current++
		} else {
			break
		}
	}
	return current, longest
}

func containsDay(dates []time.Time, day time.Time) bool {
	for _, d := range dates {
		if d.UTC().Truncate(24 * time.Hour).Equal(day) {
			return true
		}
	}
	return false
}
