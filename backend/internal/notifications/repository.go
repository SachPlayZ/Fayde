package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID, nType string, taskID *string, message string) (*Notification, error)
	ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	UnreadCount(ctx context.Context, userID string) (int, error)
	ExistsRecent(ctx context.Context, taskID, nType string, since time.Time) (bool, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Create(ctx context.Context, userID, nType string, taskID *string, message string) (*Notification, error) {
	const q = `INSERT INTO notifications (user_id, type, task_id, message)
		VALUES ($1,$2,$3,$4) RETURNING id, user_id, type, task_id, message, read, created_at`
	n := &Notification{}
	err := r.pool.QueryRow(ctx, q, userID, nType, taskID, message).
		Scan(&n.ID, &n.UserID, &n.Type, &n.TaskID, &n.Message, &n.Read, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("notifications: create: %w", err)
	}
	return n, nil
}

func (r *pgRepository) ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]*Notification, error) {
	q := `SELECT id, user_id, type, task_id, message, read, created_at
		FROM notifications WHERE user_id=$1`
	if unreadOnly {
		q += ` AND read=false`
	}
	q += ` ORDER BY created_at DESC LIMIT 50`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("notifications: list: %w", err)
	}
	defer rows.Close()
	var out []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.TaskID, &n.Message, &n.Read, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("notifications: scan: %w", err)
		}
		out = append(out, n)
	}
	if out == nil {
		out = []*Notification{}
	}
	return out, rows.Err()
}

func (r *pgRepository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (r *pgRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET read=true WHERE user_id=$1`, userID)
	return err
}

func (r *pgRepository) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read=false`, userID).Scan(&count)
	return count, err
}

func (r *pgRepository) ExistsRecent(ctx context.Context, taskID, nType string, since time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM notifications WHERE task_id=$1 AND type=$2 AND created_at > $3)`,
		taskID, nType, since).Scan(&exists)
	return exists, err
}
