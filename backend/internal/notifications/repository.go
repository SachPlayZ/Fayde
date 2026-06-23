package notifications

import (
	"context"
	"encoding/json"
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
	DeliveryTargets(ctx context.Context, userID string) (*DeliveryPrefs, error)
	Snooze(ctx context.Context, id, userID string, until time.Time) error
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
		FROM notifications WHERE user_id=$1
		AND (snoozed_until IS NULL OR snoozed_until <= now())`
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
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read=false
		 AND (snoozed_until IS NULL OR snoozed_until <= now())`, userID).Scan(&count)
	return count, err
}

func (r *pgRepository) Snooze(ctx context.Context, id, userID string, until time.Time) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE notifications SET snoozed_until=$1, read=false WHERE id=$2 AND user_id=$3`,
		until, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("notifications: snooze: not found")
	}
	return nil
}

func (r *pgRepository) ExistsRecent(ctx context.Context, taskID, nType string, since time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM notifications WHERE task_id=$1 AND type=$2 AND created_at > $3)`,
		taskID, nType, since).Scan(&exists)
	return exists, err
}

// DeliveryTargets returns a user's per-channel prefs plus the email and chat target.
func (r *pgRepository) DeliveryTargets(ctx context.Context, userID string) (*DeliveryPrefs, error) {
	const q = `SELECT email, notif_prefs, notif_chat_url, notif_chat_kind FROM users WHERE id=$1`
	var (
		email    string
		rawPrefs []byte
		chatURL  *string
		chatKind *string
	)
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&email, &rawPrefs, &chatURL, &chatKind); err != nil {
		return nil, fmt.Errorf("notifications: delivery targets: %w", err)
	}
	prefs := &DeliveryPrefs{UserMail: email}
	if len(rawPrefs) > 0 {
		if err := json.Unmarshal(rawPrefs, prefs); err != nil {
			return nil, fmt.Errorf("notifications: parse prefs: %w", err)
		}
	}
	if chatURL != nil {
		prefs.ChatURL = *chatURL
	}
	if chatKind != nil {
		prefs.ChatKind = *chatKind
	}
	return prefs, nil
}
