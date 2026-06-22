package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for users.
type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdatePreferences(ctx context.Context, id string, theme *string, digestEnabled *bool) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const userSelect = `id, email, password_hash, role, COALESCE(theme,'system'), COALESCE(digest_enabled,true), created_at`

// CreateUser inserts a new user and returns the created record.
func (r *pgRepository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2)
		RETURNING ` + userSelect
	u := &User{}
	err := r.pool.QueryRow(ctx, q, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Theme, &u.DigestEnabled, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail fetches a user by email address.
func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	q := `SELECT ` + userSelect + ` FROM users WHERE email = $1`
	u := &User{}
	err := r.pool.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Theme, &u.DigestEnabled, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("auth: get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID fetches a user by primary key.
func (r *pgRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	q := `SELECT ` + userSelect + ` FROM users WHERE id = $1`
	u := &User{}
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Theme, &u.DigestEnabled, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("auth: get user by id: %w", err)
	}
	return u, nil
}

// UpdatePreferences updates theme and/or digest_enabled for a user.
func (r *pgRepository) UpdatePreferences(ctx context.Context, id string, theme *string, digestEnabled *bool) error {
	if theme == nil && digestEnabled == nil {
		return nil
	}
	sets := []string{}
	args := []any{}
	idx := 1
	if theme != nil {
		sets = append(sets, fmt.Sprintf("theme=$%d", idx))
		args = append(args, *theme)
		idx++
	}
	if digestEnabled != nil {
		sets = append(sets, fmt.Sprintf("digest_enabled=$%d", idx))
		args = append(args, *digestEnabled)
		idx++
	}
	args = append(args, id)
	q := fmt.Sprintf(`UPDATE users SET %s WHERE id=$%d`,
		joinSets(sets), idx)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func joinSets(sets []string) string {
	result := ""
	for i, s := range sets {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
