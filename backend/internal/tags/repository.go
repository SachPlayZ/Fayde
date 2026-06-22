package tags

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListByUser(ctx context.Context, userID string) ([]*Tag, error)
	Create(ctx context.Context, userID, name, color string) (*Tag, error)
	Delete(ctx context.Context, id, userID string) error
	AddToTask(ctx context.Context, taskID, tagID string) error
	RemoveFromTask(ctx context.Context, taskID, tagID string) error
	ListByTask(ctx context.Context, taskID string) ([]*Tag, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) ListByUser(ctx context.Context, userID string) ([]*Tag, error) {
	const q = `SELECT id, user_id, name, color FROM tags WHERE user_id=$1 ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("tags: list: %w", err)
	}
	defer rows.Close()
	var out []*Tag
	for rows.Next() {
		t := &Tag{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("tags: scan: %w", err)
		}
		out = append(out, t)
	}
	if out == nil {
		out = []*Tag{}
	}
	return out, rows.Err()
}

func (r *pgRepository) Create(ctx context.Context, userID, name, color string) (*Tag, error) {
	const q = `INSERT INTO tags (user_id, name, color) VALUES ($1,$2,$3)
		ON CONFLICT (user_id, name) DO UPDATE SET color=EXCLUDED.color
		RETURNING id, user_id, name, color`
	t := &Tag{}
	err := r.pool.QueryRow(ctx, q, userID, name, color).Scan(&t.ID, &t.UserID, &t.Name, &t.Color)
	if err != nil {
		return nil, fmt.Errorf("tags: create: %w", err)
	}
	return t, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (r *pgRepository) AddToTask(ctx context.Context, taskID, tagID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO task_tags (task_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		taskID, tagID)
	return err
}

func (r *pgRepository) RemoveFromTask(ctx context.Context, taskID, tagID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM task_tags WHERE task_id=$1 AND tag_id=$2`, taskID, tagID)
	return err
}

func (r *pgRepository) ListByTask(ctx context.Context, taskID string) ([]*Tag, error) {
	const q = `SELECT t.id, t.user_id, t.name, t.color
		FROM tags t JOIN task_tags tt ON tt.tag_id=t.id WHERE tt.task_id=$1 ORDER BY t.name ASC`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("tags: list by task: %w", err)
	}
	defer rows.Close()
	var out []*Tag
	for rows.Next() {
		t := &Tag{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("tags: scan: %w", err)
		}
		out = append(out, t)
	}
	if out == nil {
		out = []*Tag{}
	}
	return out, rows.Err()
}
