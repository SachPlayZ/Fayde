package subtasks

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, taskID string) ([]*Subtask, error)
	Create(ctx context.Context, taskID, title string) (*Subtask, error)
	Update(ctx context.Context, id string, done bool, title string) (*Subtask, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, taskID string, ids []string) error
	CountByTask(ctx context.Context, taskID string) (total, done int, err error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) List(ctx context.Context, taskID string) ([]*Subtask, error) {
	const q = `SELECT id, task_id, title, done, position, created_at
		FROM subtasks WHERE task_id = $1 ORDER BY position ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("subtasks: list: %w", err)
	}
	defer rows.Close()
	var out []*Subtask
	for rows.Next() {
		s := &Subtask{}
		if err := rows.Scan(&s.ID, &s.TaskID, &s.Title, &s.Done, &s.Position, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("subtasks: scan: %w", err)
		}
		out = append(out, s)
	}
	if out == nil {
		out = []*Subtask{}
	}
	return out, rows.Err()
}

func (r *pgRepository) Create(ctx context.Context, taskID, title string) (*Subtask, error) {
	const q = `INSERT INTO subtasks (task_id, title, position)
		VALUES ($1, $2, COALESCE((SELECT MAX(position)+1 FROM subtasks WHERE task_id=$1), 0))
		RETURNING id, task_id, title, done, position, created_at`
	s := &Subtask{}
	err := r.pool.QueryRow(ctx, q, taskID, title).
		Scan(&s.ID, &s.TaskID, &s.Title, &s.Done, &s.Position, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("subtasks: create: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Update(ctx context.Context, id string, done bool, title string) (*Subtask, error) {
	const q = `UPDATE subtasks SET done=$2, title=$3 WHERE id=$1
		RETURNING id, task_id, title, done, position, created_at`
	s := &Subtask{}
	err := r.pool.QueryRow(ctx, q, id, done, title).
		Scan(&s.ID, &s.TaskID, &s.Title, &s.Done, &s.Position, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("subtasks: update: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM subtasks WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("subtasks: delete: %w", err)
	}
	return nil
}

func (r *pgRepository) Reorder(ctx context.Context, taskID string, ids []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("subtasks: reorder begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	for i, id := range ids {
		if _, err := tx.Exec(ctx, `UPDATE subtasks SET position=$1 WHERE id=$2 AND task_id=$3`, i, id, taskID); err != nil {
			return fmt.Errorf("subtasks: reorder update: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *pgRepository) CountByTask(ctx context.Context, taskID string) (total, done int, err error) {
	const q = `SELECT COUNT(*), COUNT(*) FILTER (WHERE done) FROM subtasks WHERE task_id=$1`
	err = r.pool.QueryRow(ctx, q, taskID).Scan(&total, &done)
	return
}
