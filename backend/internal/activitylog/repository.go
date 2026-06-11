package activitylog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for activity logs.
type Repository interface {
	Insert(ctx context.Context, taskID, userID, action string, changes json.RawMessage) error
	ListByTask(ctx context.Context, taskID string) ([]*ActivityLog, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Insert records a new activity log entry.
func (r *pgRepository) Insert(ctx context.Context, taskID, userID, action string, changes json.RawMessage) error {
	const q = `
		INSERT INTO activity_logs (task_id, user_id, action, changes)
		VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, q, taskID, userID, action, changes)
	if err != nil {
		return fmt.Errorf("activitylog: insert: %w", err)
	}
	return nil
}

// ListByTask returns all activity logs for a given task, ordered by creation time ascending.
func (r *pgRepository) ListByTask(ctx context.Context, taskID string) ([]*ActivityLog, error) {
	const q = `
		SELECT id, task_id, user_id, action, changes, created_at
		FROM activity_logs
		WHERE task_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("activitylog: list by task: %w", err)
	}
	defer rows.Close()

	var logs []*ActivityLog
	for rows.Next() {
		l := &ActivityLog{}
		if err := rows.Scan(&l.ID, &l.TaskID, &l.UserID, &l.Action, &l.Changes, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("activitylog: scan: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("activitylog: rows: %w", err)
	}

	if logs == nil {
		logs = []*ActivityLog{}
	}
	return logs, nil
}
