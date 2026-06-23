package calendarsync

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("connection not found")

// Repository manages database interaction for Google Calendar connections.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetConnection fetches a user's Google Calendar connection.
func (r *Repository) GetConnection(ctx context.Context, userID string) (*CalendarConnection, error) {
	conn := &CalendarConnection{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, email, access_token, refresh_token, expiry, created_at, updated_at
		 FROM calendar_connections WHERE user_id=$1 AND provider='google'`, userID).
		Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.Email, &conn.AccessToken, &conn.RefreshToken, &conn.Expiry, &conn.CreatedAt, &conn.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return conn, err
}

// SaveConnection stores or updates a user's Google Calendar connection.
func (r *Repository) SaveConnection(ctx context.Context, userID, email, accessToken, refreshToken string, expiry time.Time) (*CalendarConnection, error) {
	conn := &CalendarConnection{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO calendar_connections (user_id, provider, email, access_token, refresh_token, expiry)
		 VALUES ($1, 'google', $2, $3, $4, $5)
		 ON CONFLICT (user_id, provider) DO UPDATE SET
		   email=EXCLUDED.email,
		   access_token=EXCLUDED.access_token,
		   refresh_token=EXCLUDED.refresh_token,
		   expiry=EXCLUDED.expiry,
		   updated_at=now()
		 RETURNING id, user_id, provider, email, access_token, refresh_token, expiry, created_at, updated_at`,
		userID, email, accessToken, refreshToken, expiry).
		Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.Email, &conn.AccessToken, &conn.RefreshToken, &conn.Expiry, &conn.CreatedAt, &conn.UpdatedAt)
	return conn, err
}

// UpdateAccessToken updates the Google access token and expiry after a refresh.
func (r *Repository) UpdateAccessToken(ctx context.Context, userID, accessToken string, expiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE calendar_connections SET access_token=$1, expiry=$2, updated_at=now()
		 WHERE user_id=$3 AND provider='google'`, accessToken, expiry, userID)
	return err
}

// DeleteConnection deletes a user's Google Calendar connection.
func (r *Repository) DeleteConnection(ctx context.Context, userID string) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM calendar_connections WHERE user_id=$1 AND provider='google'`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateTaskExternalEventID writes the Google Calendar event ID back to the task.
func (r *Repository) UpdateTaskExternalEventID(ctx context.Context, taskID string, eventID *string) error {
	_, err := r.pool.Exec(ctx, `UPDATE tasks SET external_event_id=$1, updated_at=now() WHERE id=$2`, eventID, taskID)
	return err
}
