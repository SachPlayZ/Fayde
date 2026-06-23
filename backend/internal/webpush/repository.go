package webpush

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists web push subscriptions.
type Repository interface {
	Subscribe(ctx context.Context, userID, endpoint, p256dh, auth string) (*Subscription, error)
	Unsubscribe(ctx context.Context, userID, endpoint string) error
	ListByUser(ctx context.Context, userID string) ([]*Subscription, error)
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}

type pgRepository struct{ pool *pgxpool.Pool }

// NewRepository builds a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Subscribe(ctx context.Context, userID, endpoint, p256dh, auth string) (*Subscription, error) {
	const q = `INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, endpoint) DO UPDATE SET p256dh=EXCLUDED.p256dh, auth=EXCLUDED.auth
		RETURNING id, user_id, endpoint, p256dh, auth, created_at`
	s := &Subscription{}
	err := r.pool.QueryRow(ctx, q, userID, endpoint, p256dh, auth).
		Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("webpush: subscribe: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Unsubscribe(ctx context.Context, userID, endpoint string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE user_id=$1 AND endpoint=$2`, userID, endpoint)
	return err
}

func (r *pgRepository) ListByUser(ctx context.Context, userID string) ([]*Subscription, error) {
	const q = `SELECT id, user_id, endpoint, p256dh, auth, created_at
		FROM push_subscriptions WHERE user_id=$1`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("webpush: list: %w", err)
	}
	defer rows.Close()
	var out []*Subscription
	for rows.Next() {
		s := &Subscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("webpush: scan: %w", err)
		}
		out = append(out, s)
	}
	if out == nil {
		out = []*Subscription{}
	}
	return out, rows.Err()
}

// DeleteByEndpoint removes a dead subscription (e.g. 404/410 from push service).
func (r *pgRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint=$1`, endpoint)
	return err
}
