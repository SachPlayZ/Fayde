package sharing

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for share tokens.
type Repository interface {
	Create(ctx context.Context, taskID, token string) (*ShareToken, error)
	GetByToken(ctx context.Context, token string) (*ShareToken, error)
	GetByTaskID(ctx context.Context, taskID string) (*ShareToken, error)
	DeleteByTaskID(ctx context.Context, taskID string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, taskID, token string) (*ShareToken, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO task_share_tokens (id, task_id, token)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, token, created_at
	`, id, taskID, token)

	var s ShareToken
	if err := row.Scan(&s.ID, &s.TaskID, &s.Token, &s.CreatedAt); err != nil {
		return nil, fmt.Errorf("sharing.Create: %w", err)
	}
	return &s, nil
}

func (r *pgRepository) GetByToken(ctx context.Context, token string) (*ShareToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, task_id, token, created_at
		FROM task_share_tokens WHERE token = $1
	`, token)

	var s ShareToken
	if err := row.Scan(&s.ID, &s.TaskID, &s.Token, &s.CreatedAt); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("sharing.GetByToken: %w", err)
	}
	return &s, nil
}

func (r *pgRepository) GetByTaskID(ctx context.Context, taskID string) (*ShareToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, task_id, token, created_at
		FROM task_share_tokens WHERE task_id = $1
	`, taskID)

	var s ShareToken
	if err := row.Scan(&s.ID, &s.TaskID, &s.Token, &s.CreatedAt); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("sharing.GetByTaskID: %w", err)
	}
	return &s, nil
}

func (r *pgRepository) DeleteByTaskID(ctx context.Context, taskID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM task_share_tokens WHERE task_id = $1`, taskID)
	if err != nil {
		return fmt.Errorf("sharing.DeleteByTaskID: %w", err)
	}
	return nil
}
