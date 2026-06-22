package totp

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for TOTP secrets.
type Repository interface {
	Create(ctx context.Context, userID, secret string) (*TOTPSecret, error)
	Get(ctx context.Context, userID string) (*TOTPSecret, error)
	Enable(ctx context.Context, userID string) error
	Disable(ctx context.Context, userID string) error
	GetUserTOTPEnabled(ctx context.Context, userID string) (bool, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, userID, secret string) (*TOTPSecret, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO totp_secrets (user_id, secret, enabled)
		VALUES ($1, $2, false)
		ON CONFLICT (user_id) DO UPDATE SET secret = $2, enabled = false
		RETURNING user_id, secret, enabled, created_at
	`, userID, secret)

	var ts TOTPSecret
	if err := row.Scan(&ts.UserID, &ts.Secret, &ts.Enabled, &ts.CreatedAt); err != nil {
		return nil, fmt.Errorf("totp.Create: %w", err)
	}
	return &ts, nil
}

func (r *pgRepository) Get(ctx context.Context, userID string) (*TOTPSecret, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, secret, enabled, created_at
		FROM totp_secrets WHERE user_id = $1
	`, userID)

	var ts TOTPSecret
	if err := row.Scan(&ts.UserID, &ts.Secret, &ts.Enabled, &ts.CreatedAt); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("totp.Get: %w", err)
	}
	return &ts, nil
}

func (r *pgRepository) Enable(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE totp_secrets SET enabled = true WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("totp.Enable (secret): %w", err)
	}
	_, err = r.pool.Exec(ctx, `UPDATE users SET totp_enabled = true WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("totp.Enable (user): %w", err)
	}
	return nil
}

func (r *pgRepository) Disable(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE totp_secrets SET enabled = false WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("totp.Disable (secret): %w", err)
	}
	_, err = r.pool.Exec(ctx, `UPDATE users SET totp_enabled = false WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("totp.Disable (user): %w", err)
	}
	return nil
}

func (r *pgRepository) GetUserTOTPEnabled(ctx context.Context, userID string) (bool, error) {
	row := r.pool.QueryRow(ctx, `SELECT totp_enabled FROM users WHERE id = $1`, userID)
	var enabled bool
	if err := row.Scan(&enabled); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("totp.GetUserTOTPEnabled: %w", err)
	}
	return enabled, nil
}
