// Package dashboard aggregates a personal home-screen summary for a user.
package dashboard

import (
	"context"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/habits"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskBrief is a compact task for dashboard lists.
type TaskBrief struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Priority string  `json:"priority"`
	Status   string  `json:"status"`
	DueDate  *string `json:"due_date"`
}

// Summary is the full dashboard payload.
type Summary struct {
	DueToday          []*TaskBrief    `json:"due_today"`
	Overdue           []*TaskBrief    `json:"overdue"`
	Upcoming          []*TaskBrief    `json:"upcoming"`
	CompletedThisWeek int             `json:"completed_this_week"`
	CreatedThisWeek   int             `json:"created_this_week"`
	TimeThisWeekMin   int             `json:"time_this_week_minutes"`
	PomodorosToday    int             `json:"pomodoros_today"`
	Habits            []*habits.Habit `json:"habits"`
}

type Service struct {
	pool      *pgxpool.Pool
	habitsSvc *habits.Service
}

func NewService(pool *pgxpool.Pool, habitsSvc *habits.Service) *Service {
	return &Service{pool: pool, habitsSvc: habitsSvc}
}

func (s *Service) Summary(ctx context.Context, userID string) (*Summary, error) {
	sum := &Summary{
		DueToday: []*TaskBrief{}, Overdue: []*TaskBrief{}, Upcoming: []*TaskBrief{},
		Habits: []*habits.Habit{},
	}

	const openFilter = `status NOT IN ('done','failed')`

	dueToday, err := s.tasks(ctx,
		`WHERE user_id=$1 AND due_date::date = (now() at time zone 'utc')::date AND `+openFilter+
			` ORDER BY priority DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	sum.DueToday = dueToday

	overdue, err := s.tasks(ctx,
		`WHERE user_id=$1 AND due_date < now() AND due_date::date < (now() at time zone 'utc')::date AND `+openFilter+
			` ORDER BY due_date ASC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	sum.Overdue = overdue

	upcoming, err := s.tasks(ctx,
		`WHERE user_id=$1 AND due_date::date > (now() at time zone 'utc')::date AND `+openFilter+
			` ORDER BY due_date ASC LIMIT 10`, userID)
	if err != nil {
		return nil, err
	}
	sum.Upcoming = upcoming

	_ = s.pool.QueryRow(ctx,
		`SELECT count(*) FROM tasks WHERE user_id=$1 AND status='done' AND updated_at >= now()-interval '7 days'`, userID).
		Scan(&sum.CompletedThisWeek)
	_ = s.pool.QueryRow(ctx,
		`SELECT count(*) FROM tasks WHERE user_id=$1 AND created_at >= now()-interval '7 days'`, userID).
		Scan(&sum.CreatedThisWeek)

	var secs int
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(duration_seconds),0) FROM time_entries
		 WHERE user_id=$1 AND started_at >= now()-interval '7 days'`, userID).Scan(&secs)
	sum.TimeThisWeekMin = secs / 60

	_ = s.pool.QueryRow(ctx,
		`SELECT count(*) FROM pomodoro_sessions WHERE user_id=$1 AND completed=true
		 AND started_at::date = (now() at time zone 'utc')::date`, userID).Scan(&sum.PomodorosToday)

	if hs, err := s.habitsSvc.List(ctx, userID); err == nil {
		sum.Habits = hs
	}

	return sum, nil
}

func (s *Service) tasks(ctx context.Context, where string, args ...any) ([]*TaskBrief, error) {
	q := `SELECT id, title, priority, status, to_char(due_date,'YYYY-MM-DD"T"HH24:MI:SSZ') FROM tasks ` + where
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*TaskBrief{}
	for rows.Next() {
		t := &TaskBrief{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Priority, &t.Status, &t.DueDate); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Handler exposes GET /dashboard.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	sum, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}
	httputil.JSON(w, http.StatusOK, sum)
}
