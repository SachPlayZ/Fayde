package webhooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for outbound webhooks.
type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*OutboundWebhook, error)
	List(ctx context.Context, userID string) ([]*OutboundWebhook, error)
	ListEnabledForUser(ctx context.Context, userID string) ([]*OutboundWebhook, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*OutboundWebhook, error)
	Delete(ctx context.Context, id, userID string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*OutboundWebhook, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO outbound_webhooks (id, user_id, name, url, events, secret, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, true)
		RETURNING id, user_id, name, url, events, secret, enabled, created_at
	`, id, userID, req.Name, req.URL, req.Events, req.Secret)

	return scanWebhook(row)
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*OutboundWebhook, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, url, events, secret, enabled, created_at
		FROM outbound_webhooks WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("webhooks.List: %w", err)
	}
	defer rows.Close()

	var hooks []*OutboundWebhook
	for rows.Next() {
		h, err := scanWebhookRow(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (r *pgRepository) ListEnabledForUser(ctx context.Context, userID string) ([]*OutboundWebhook, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, url, events, secret, enabled, created_at
		FROM outbound_webhooks WHERE user_id = $1 AND enabled = true
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("webhooks.ListEnabledForUser: %w", err)
	}
	defer rows.Close()

	var hooks []*OutboundWebhook
	for rows.Next() {
		h, err := scanWebhookRow(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*OutboundWebhook, error) {
	// Build dynamic SET clause
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.URL != nil {
		setClauses = append(setClauses, fmt.Sprintf("url = $%d", argIdx))
		args = append(args, *req.URL)
		argIdx++
	}
	if req.Events != nil {
		setClauses = append(setClauses, fmt.Sprintf("events = $%d", argIdx))
		args = append(args, req.Events)
		argIdx++
	}
	if req.Secret != nil {
		setClauses = append(setClauses, fmt.Sprintf("secret = $%d", argIdx))
		args = append(args, *req.Secret)
		argIdx++
	}
	if req.Enabled != nil {
		setClauses = append(setClauses, fmt.Sprintf("enabled = $%d", argIdx))
		args = append(args, *req.Enabled)
		argIdx++
	}

	if len(setClauses) == 0 {
		// Nothing to update — fetch current
		row := r.pool.QueryRow(ctx, `
			SELECT id, user_id, name, url, events, secret, enabled, created_at
			FROM outbound_webhooks WHERE id = $1 AND user_id = $2
		`, id, userID)
		return scanWebhook(row)
	}

	args = append(args, id, userID)
	query := fmt.Sprintf(`
		UPDATE outbound_webhooks SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, name, url, events, secret, enabled, created_at
	`, strings.Join(setClauses, ", "), argIdx, argIdx+1)

	row := r.pool.QueryRow(ctx, query, args...)
	h, err := scanWebhook(row)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("webhooks.Update: %w", err)
	}
	return h, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM outbound_webhooks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("webhooks.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWebhook(row scanner) (*OutboundWebhook, error) {
	var h OutboundWebhook
	if err := row.Scan(&h.ID, &h.UserID, &h.Name, &h.URL, &h.Events, &h.Secret, &h.Enabled, &h.CreatedAt); err != nil {
		return nil, fmt.Errorf("webhooks scan: %w", err)
	}
	return &h, nil
}

func scanWebhookRow(rows interface{ Scan(dest ...any) error }) (*OutboundWebhook, error) {
	return scanWebhook(rows)
}
