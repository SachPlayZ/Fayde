package apitokens

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID, name, tokenHash, tokenPrefix string) (*APIToken, error)
	List(ctx context.Context, userID string) ([]*APIToken, error)
	Delete(ctx context.Context, id, userID string) error
	FindByHash(ctx context.Context, hash string) (*LookupResult, error)
	UpdateLastUsed(ctx context.Context, id string) error
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Create(ctx context.Context, userID, name, tokenHash, tokenPrefix string) (*APIToken, error) {
	const q = `INSERT INTO api_tokens (user_id, name, token_hash, token_prefix)
		VALUES ($1,$2,$3,$4)
		RETURNING id, user_id, name, token_prefix, last_used_at, created_at`
	t := &APIToken{}
	err := r.pool.QueryRow(ctx, q, userID, name, tokenHash, tokenPrefix).
		Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.LastUsedAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("apitokens.Create: %w", err)
	}
	return t, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*APIToken, error) {
	const q = `SELECT id, user_id, name, token_prefix, last_used_at, created_at
		FROM api_tokens WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("apitokens.List: %w", err)
	}
	defer rows.Close()
	var out []*APIToken
	for rows.Next() {
		t := &APIToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.LastUsedAt, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("apitokens.List scan: %w", err)
		}
		out = append(out, t)
	}
	if out == nil {
		out = []*APIToken{}
	}
	return out, rows.Err()
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM api_tokens WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("apitokens.Delete: %w", err)
	}
	return nil
}

func (r *pgRepository) FindByHash(ctx context.Context, hash string) (*LookupResult, error) {
	lr := &LookupResult{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id FROM api_tokens WHERE token_hash=$1`, hash).
		Scan(&lr.ID, &lr.UserID)
	if err != nil {
		return nil, fmt.Errorf("apitokens.FindByHash: %w", err)
	}
	return lr, nil
}

func (r *pgRepository) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_tokens SET last_used_at=now() WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("apitokens.UpdateLastUsed: %w", err)
	}
	return nil
}
