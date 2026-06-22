package watchers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, taskID, userID string) error
	Remove(ctx context.Context, taskID, userID string) error
	List(ctx context.Context, taskID string) ([]*Watcher, error)
	IsWatching(ctx context.Context, taskID, userID string) (bool, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Add(ctx context.Context, taskID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO task_watchers (task_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		taskID, userID)
	if err != nil {
		return fmt.Errorf("watchers.Add: %w", err)
	}
	return nil
}

func (r *pgRepository) Remove(ctx context.Context, taskID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM task_watchers WHERE task_id=$1 AND user_id=$2`, taskID, userID)
	if err != nil {
		return fmt.Errorf("watchers.Remove: %w", err)
	}
	return nil
}

func (r *pgRepository) List(ctx context.Context, taskID string) ([]*Watcher, error) {
	const q = `SELECT w.task_id, w.user_id, u.email
		FROM task_watchers w JOIN users u ON u.id=w.user_id
		WHERE w.task_id=$1`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("watchers.List: %w", err)
	}
	defer rows.Close()
	var out []*Watcher
	for rows.Next() {
		w := &Watcher{}
		if err := rows.Scan(&w.TaskID, &w.UserID, &w.UserEmail); err != nil {
			return nil, fmt.Errorf("watchers.List scan: %w", err)
		}
		out = append(out, w)
	}
	if out == nil {
		out = []*Watcher{}
	}
	return out, rows.Err()
}

func (r *pgRepository) IsWatching(ctx context.Context, taskID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM task_watchers WHERE task_id=$1 AND user_id=$2)`,
		taskID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("watchers.IsWatching: %w", err)
	}
	return exists, nil
}
