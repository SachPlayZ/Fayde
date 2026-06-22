package savedfilters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*SavedFilter, error)
	List(ctx context.Context, userID string) ([]*SavedFilter, error)
	Delete(ctx context.Context, id, userID string) error
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*SavedFilter, error) {
	const q = `INSERT INTO saved_filters (id, user_id, name, params)
		VALUES ($1,$2,$3,$4)
		RETURNING id, user_id, name, params, created_at`
	id := uuid.New().String()
	sf := &SavedFilter{}
	err := r.pool.QueryRow(ctx, q, id, userID, req.Name, req.Params).
		Scan(&sf.ID, &sf.UserID, &sf.Name, &sf.Params, &sf.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("savedfilters.Create: %w", err)
	}
	return sf, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*SavedFilter, error) {
	const q = `SELECT id, user_id, name, params, created_at
		FROM saved_filters WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("savedfilters.List: %w", err)
	}
	defer rows.Close()
	var out []*SavedFilter
	for rows.Next() {
		sf := &SavedFilter{}
		if err := rows.Scan(&sf.ID, &sf.UserID, &sf.Name, &sf.Params, &sf.CreatedAt); err != nil {
			return nil, fmt.Errorf("savedfilters.List scan: %w", err)
		}
		out = append(out, sf)
	}
	if out == nil {
		out = []*SavedFilter{}
	}
	return out, rows.Err()
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM saved_filters WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("savedfilters.Delete: %w", err)
	}
	return nil
}
