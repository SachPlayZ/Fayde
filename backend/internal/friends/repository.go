package friends

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for friendships.
type Repository interface {
	// UserByEmail looks up a user's ID by email address.
	UserByEmail(ctx context.Context, email string) (*FriendUser, error)
	// UserByID looks up a user by ID.
	UserByID(ctx context.Context, id string) (*FriendUser, error)
	// Create inserts a new pending friendship request.
	Create(ctx context.Context, requesterID, addresseeID string) (*Friendship, error)
	// GetByID returns a friendship by its primary key.
	GetByID(ctx context.Context, id string) (*Friendship, error)
	// GetBetween returns the friendship row for a pair (either direction), if any.
	GetBetween(ctx context.Context, a, b string) (*Friendship, error)
	// UpdateStatus changes a friendship's status.
	UpdateStatus(ctx context.Context, id, status string) error
	// Delete removes a friendship row entirely.
	Delete(ctx context.Context, id string) error
	// ListFriends returns accepted friends for a user.
	ListFriends(ctx context.Context, userID string) ([]*FriendRequest, error)
	// ListRequests returns pending incoming + outgoing requests.
	ListRequests(ctx context.Context, userID string) ([]*FriendRequest, error)
	// SearchByEmail finds users whose email matches the query (ILIKE prefix match).
	SearchByEmail(ctx context.Context, q string) ([]*FriendUser, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) UserByEmail(ctx context.Context, email string) (*FriendUser, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, display_name, avatar_url
		FROM users WHERE lower(email) = lower($1)
	`, email)
	return scanFriendUser(row)
}

func (r *pgRepository) UserByID(ctx context.Context, id string) (*FriendUser, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, display_name, avatar_url
		FROM users WHERE id = $1
	`, id)
	return scanFriendUser(row)
}

func scanFriendUser(row pgx.Row) (*FriendUser, error) {
	var u FriendUser
	if err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("friends: scan user: %w", err)
	}
	return &u, nil
}

func (r *pgRepository) Create(ctx context.Context, requesterID, addresseeID string) (*Friendship, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO friendships (id, requester_id, addressee_id, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, requester_id, addressee_id, status, created_at, responded_at
	`, id, requesterID, addresseeID)
	return scanFriendship(row)
}

func (r *pgRepository) GetByID(ctx context.Context, id string) (*Friendship, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, requester_id, addressee_id, status, created_at, responded_at
		FROM friendships WHERE id = $1
	`, id)
	f, err := scanFriendship(row)
	if err != nil && strings.Contains(err.Error(), "no rows") {
		return nil, ErrNotFound
	}
	return f, err
}

func (r *pgRepository) GetBetween(ctx context.Context, a, b string) (*Friendship, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, requester_id, addressee_id, status, created_at, responded_at
		FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
	`, a, b)
	f, err := scanFriendship(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
	}
	return f, err
}

func scanFriendship(row pgx.Row) (*Friendship, error) {
	var f Friendship
	if err := row.Scan(&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.RespondedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("friends: scan friendship: %w", err)
	}
	return &f, nil
}

func (r *pgRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE friendships SET status = $2, responded_at = now()
		WHERE id = $1
	`, id, status)
	return err
}

func (r *pgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM friendships WHERE id = $1`, id)
	return err
}

func (r *pgRepository) ListFriends(ctx context.Context, userID string) ([]*FriendRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.status, f.created_at, f.responded_at,
		       u.id, u.email, u.display_name, u.avatar_url,
		       f.requester_id
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		  AND f.status = 'accepted'
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("friends.ListFriends: %w", err)
	}
	defer rows.Close()
	return scanRequests(rows, userID)
}

func (r *pgRepository) ListRequests(ctx context.Context, userID string) ([]*FriendRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.status, f.created_at, f.responded_at,
		       u.id, u.email, u.display_name, u.avatar_url,
		       f.requester_id
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		  AND f.status = 'pending'
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("friends.ListRequests: %w", err)
	}
	defer rows.Close()
	return scanRequests(rows, userID)
}

func scanRequests(rows pgx.Rows, callerUserID string) ([]*FriendRequest, error) {
	var out []*FriendRequest
	for rows.Next() {
		var fr FriendRequest
		var requesterID string
		if err := rows.Scan(
			&fr.ID, &fr.Status, &fr.CreatedAt, &fr.RespondedAt,
			&fr.User.ID, &fr.User.Email, &fr.User.DisplayName, &fr.User.AvatarURL,
			&requesterID,
		); err != nil {
			return nil, fmt.Errorf("friends: scan request row: %w", err)
		}
		if requesterID == callerUserID {
			fr.Direction = "outgoing"
		} else {
			fr.Direction = "incoming"
		}
		out = append(out, &fr)
	}
	return out, rows.Err()
}

func (r *pgRepository) SearchByEmail(ctx context.Context, q string) ([]*FriendUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, display_name, avatar_url
		FROM users
		WHERE lower(email) LIKE lower($1)
		LIMIT 20
	`, q+"%")
	if err != nil {
		return nil, fmt.Errorf("friends.SearchByEmail: %w", err)
	}
	defer rows.Close()

	var out []*FriendUser
	for rows.Next() {
		var u FriendUser
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL); err != nil {
			return nil, fmt.Errorf("friends: scan search row: %w", err)
		}
		out = append(out, &u)
	}
	return out, rows.Err()
}
