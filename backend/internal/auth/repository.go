package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for users.
type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdatePreferences(ctx context.Context, id string, prefs Preferences) error

	UpsertOAuthUser(ctx context.Context, email, provider, providerID string) (*User, error)

	CreateVerificationToken(ctx context.Context, userID string) (string, error)
	ConsumeVerificationToken(ctx context.Context, token string) (string, error)
	DeleteVerificationTokensForUser(ctx context.Context, userID string) error
	MarkEmailVerified(ctx context.Context, userID string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const userSelect = `id, email, password_hash, role,
	COALESCE(theme,'system'), COALESCE(digest_enabled,true),
	COALESCE(notif_prefs,'{}'), notif_chat_url, notif_chat_kind, inbox_token,
	COALESCE(email_verified,false), COALESCE(provider,'local'), provider_id,
	created_at`

func scanUser(row pgx.Row) (*User, error) {
	u := &User{}
	err := row.Scan(
		&u.ID, &u.Email, &u.passwordHash, &u.Role,
		&u.Theme, &u.DigestEnabled,
		&u.NotifPrefs, &u.ChatURL, &u.ChatKind, &u.InboxToken,
		&u.EmailVerified, &u.Provider, &u.ProviderID,
		&u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser inserts a new local user and returns the created record.
func (r *pgRepository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	const q = `INSERT INTO users (email, password_hash, email_verified, provider)
		VALUES ($1, $2, false, 'local')
		RETURNING ` + userSelect
	u, err := scanUser(r.pool.QueryRow(ctx, q, email, passwordHash))
	if err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail fetches a user by email address.
func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	q := `SELECT ` + userSelect + ` FROM users WHERE email = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, q, email))
	if err != nil {
		return nil, fmt.Errorf("auth: get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID fetches a user by primary key.
func (r *pgRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	q := `SELECT ` + userSelect + ` FROM users WHERE id = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("auth: get user by id: %w", err)
	}
	return u, nil
}

// UpdatePreferences updates any provided user preference fields.
func (r *pgRepository) UpdatePreferences(ctx context.Context, id string, prefs Preferences) error {
	sets := []string{}
	args := []any{}
	idx := 1
	if prefs.Theme != nil {
		sets = append(sets, fmt.Sprintf("theme=$%d", idx))
		args = append(args, *prefs.Theme)
		idx++
	}
	if prefs.DigestEnabled != nil {
		sets = append(sets, fmt.Sprintf("digest_enabled=$%d", idx))
		args = append(args, *prefs.DigestEnabled)
		idx++
	}
	if prefs.NotifPrefs != nil {
		sets = append(sets, fmt.Sprintf("notif_prefs=$%d", idx))
		args = append(args, []byte(*prefs.NotifPrefs))
		idx++
	}
	if prefs.ChatURL != nil {
		sets = append(sets, fmt.Sprintf("notif_chat_url=$%d", idx))
		args = append(args, *prefs.ChatURL)
		idx++
	}
	if prefs.ChatKind != nil {
		sets = append(sets, fmt.Sprintf("notif_chat_kind=$%d", idx))
		args = append(args, *prefs.ChatKind)
		idx++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	q := fmt.Sprintf(`UPDATE users SET %s WHERE id=$%d`, joinSets(sets), idx)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

// UpsertOAuthUser inserts or updates an OAuth user by email.
// If a user with that email already exists, their provider info is updated (merge).
func (r *pgRepository) UpsertOAuthUser(ctx context.Context, email, provider, providerID string) (*User, error) {
	const q = `INSERT INTO users (email, provider, provider_id, email_verified)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (email) DO UPDATE
		  SET provider    = EXCLUDED.provider,
		      provider_id = EXCLUDED.provider_id,
		      email_verified = true
		RETURNING ` + userSelect
	u, err := scanUser(r.pool.QueryRow(ctx, q, email, provider, providerID))
	if err != nil {
		return nil, fmt.Errorf("auth: upsert oauth user: %w", err)
	}
	return u, nil
}

// CreateVerificationToken generates a secure random token, stores it, and returns it.
func (r *pgRepository) CreateVerificationToken(ctx context.Context, userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	const q = `INSERT INTO email_verifications (user_id, token) VALUES ($1, $2)`
	if _, err := r.pool.Exec(ctx, q, userID, token); err != nil {
		return "", fmt.Errorf("auth: create verification token: %w", err)
	}
	return token, nil
}

// ConsumeVerificationToken validates the token and returns the owning userID.
// Returns an error if the token is expired, already used, or not found.
func (r *pgRepository) ConsumeVerificationToken(ctx context.Context, token string) (string, error) {
	var userID string
	var expiresAt time.Time
	var usedAt *time.Time
	const q = `SELECT user_id, expires_at, used_at FROM email_verifications WHERE token = $1`
	err := r.pool.QueryRow(ctx, q, token).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("auth: consume token: %w", err)
	}
	if usedAt != nil {
		return "", ErrInvalidToken
	}
	if time.Now().After(expiresAt) {
		return "", ErrTokenExpired
	}
	if _, err := r.pool.Exec(ctx, `UPDATE email_verifications SET used_at=now() WHERE token=$1`, token); err != nil {
		return "", fmt.Errorf("auth: mark token used: %w", err)
	}
	return userID, nil
}

// DeleteVerificationTokensForUser removes all pending tokens for a user (used before resend).
func (r *pgRepository) DeleteVerificationTokensForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_verifications WHERE user_id=$1`, userID)
	return err
}

// MarkEmailVerified sets email_verified=true for the given user.
func (r *pgRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET email_verified=true WHERE id=$1`, userID)
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
