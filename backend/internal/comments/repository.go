package comments

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, taskID string) ([]*Comment, error)
	Create(ctx context.Context, taskID, userID, body string) (*Comment, error)
	Update(ctx context.Context, id, userID, body string) (*Comment, error)
	Delete(ctx context.Context, id, userID string) error
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) List(ctx context.Context, taskID string) ([]*Comment, error) {
	const q = `SELECT c.id, c.task_id, c.user_id, u.email, c.body, c.created_at, c.updated_at
		FROM task_comments c JOIN users u ON u.id=c.user_id
		WHERE c.task_id=$1 ORDER BY c.created_at ASC`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("comments: list: %w", err)
	}
	defer rows.Close()
	var out []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.UserEmail, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("comments: scan: %w", err)
		}
		out = append(out, c)
	}
	if out == nil {
		out = []*Comment{}
	}
	return out, rows.Err()
}

func (r *pgRepository) Create(ctx context.Context, taskID, userID, body string) (*Comment, error) {
	const q = `WITH ins AS (
		INSERT INTO task_comments (task_id, user_id, body) VALUES ($1,$2,$3)
		RETURNING id, task_id, user_id, body, created_at, updated_at
	) SELECT ins.id, ins.task_id, ins.user_id, u.email, ins.body, ins.created_at, ins.updated_at
	  FROM ins JOIN users u ON u.id=ins.user_id`
	c := &Comment{}
	err := r.pool.QueryRow(ctx, q, taskID, userID, body).
		Scan(&c.ID, &c.TaskID, &c.UserID, &c.UserEmail, &c.Body, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("comments: create: %w", err)
	}
	return c, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID, body string) (*Comment, error) {
	const q = `WITH upd AS (
		UPDATE task_comments SET body=$3, updated_at=now() WHERE id=$1 AND user_id=$2
		RETURNING id, task_id, user_id, body, created_at, updated_at
	) SELECT upd.id, upd.task_id, upd.user_id, u.email, upd.body, upd.created_at, upd.updated_at
	  FROM upd JOIN users u ON u.id=upd.user_id`
	c := &Comment{}
	err := r.pool.QueryRow(ctx, q, id, userID, body).
		Scan(&c.ID, &c.TaskID, &c.UserID, &c.UserEmail, &c.Body, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("comments: update: %w", err)
	}
	return c, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM task_comments WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}
