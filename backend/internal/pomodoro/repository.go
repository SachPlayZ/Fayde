package pomodoro

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for pomodoro sessions.
type Repository interface {
	Start(ctx context.Context, userID string, taskID *string, durationMins int) (*Session, error)
	Complete(ctx context.Context, id, userID string) (*Session, error)
	Abandon(ctx context.Context, id, userID string) (*Session, error)
	List(ctx context.Context, userID string) ([]*Session, error)
	ActiveSession(ctx context.Context, userID string) (*Session, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const sessionSelect = `id, task_id, user_id, duration_minutes, completed, started_at, ended_at`

func scanSession(row interface{ Scan(dest ...any) error }) (*Session, error) {
	var s Session
	if err := row.Scan(&s.ID, &s.TaskID, &s.UserID, &s.DurationMinutes, &s.Completed, &s.StartedAt, &s.EndedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgRepository) Start(ctx context.Context, userID string, taskID *string, durationMins int) (*Session, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO pomodoro_sessions (id, user_id, task_id, duration_minutes, completed)
		VALUES ($1, $2, $3, $4, false)
		RETURNING `+sessionSelect,
		id, userID, taskID, durationMins)

	s, err := scanSession(row)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.Start: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Complete(ctx context.Context, id, userID string) (*Session, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pomodoro_sessions
		SET completed = true, ended_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING `+sessionSelect,
		id, userID)

	s, err := scanSession(row)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("pomodoro.Complete: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Abandon(ctx context.Context, id, userID string) (*Session, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pomodoro_sessions
		SET ended_at = NOW()
		WHERE id = $1 AND user_id = $2 AND completed = false
		RETURNING `+sessionSelect,
		id, userID)

	s, err := scanSession(row)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("pomodoro.Abandon: %w", err)
	}
	return s, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+sessionSelect+`
		FROM pomodoro_sessions WHERE user_id = $1
		ORDER BY started_at DESC LIMIT 50
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.List: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("pomodoro.List scan: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *pgRepository) ActiveSession(ctx context.Context, userID string) (*Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+sessionSelect+`
		FROM pomodoro_sessions WHERE user_id = $1 AND ended_at IS NULL LIMIT 1
	`, userID)

	s, err := scanSession(row)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("pomodoro.ActiveSession: %w", err)
	}
	return s, nil
}
