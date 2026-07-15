package showcase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateEntry(ctx context.Context, userID string, req CreateRequest) (*Entry, error)
	ListEntries(ctx context.Context, userID string) ([]*Entry, error)
	GetEntry(ctx context.Context, id, userID string) (*Entry, error)
	UpdateEntry(ctx context.Context, id, userID string, req UpdateRequest) (*Entry, error)
	DeleteEntry(ctx context.Context, id, userID string) error
	SetLogoKey(ctx context.Context, id, userID, key string) error
	SetBannerKey(ctx context.Context, id, userID, key string) error
	ClearLogoKey(ctx context.Context, id, userID string) (oldKey string, err error)
	ClearBannerKey(ctx context.Context, id, userID string) (oldKey string, err error)
	GetEntryKeys(ctx context.Context, id string) (logoKey, bannerKey *string, err error)

	ListEntriesByUsername(ctx context.Context, slug string) ([]*Entry, string, error)

	CreateToken(ctx context.Context, userID, name, hash, prefix string) (*ShowcaseToken, error)
	ListTokens(ctx context.Context, userID string) ([]*ShowcaseToken, error)
	DeleteToken(ctx context.Context, id, userID string) error
	FindTokenByHash(ctx context.Context, hash string) (*TokenLookupResult, error)
	UpdateTokenLastUsed(ctx context.Context, id string) error
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

const entryCols = `id, user_id, title, tagline, problem, solution, tech_stack,
	demo_url, repo_url, live_url, logo_s3_key, banner_s3_key, sort_order, created_at, updated_at`

func scanEntry(row interface {
	Scan(dest ...any) error
}) (*Entry, error) {
	e := &Entry{}
	var techStack []byte
	if err := row.Scan(&e.ID, &e.UserID, &e.Title, &e.Tagline, &e.Problem, &e.Solution, &techStack,
		&e.DemoURL, &e.RepoURL, &e.LiveURL, &e.logoS3Key, &e.bannerS3Key, &e.SortOrder, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	if len(techStack) > 0 {
		if err := json.Unmarshal(techStack, &e.TechStack); err != nil {
			return nil, fmt.Errorf("showcase: unmarshal tech_stack: %w", err)
		}
	}
	if e.TechStack == nil {
		e.TechStack = []TechItem{}
	}
	return e, nil
}

func (r *pgRepository) CreateEntry(ctx context.Context, userID string, req CreateRequest) (*Entry, error) {
	techStack := req.TechStack
	if techStack == nil {
		techStack = []TechItem{}
	}
	tsJSON, err := json.Marshal(techStack)
	if err != nil {
		return nil, fmt.Errorf("showcase: marshal tech_stack: %w", err)
	}

	const q = `INSERT INTO showcase_entries
		(user_id, title, tagline, problem, solution, tech_stack, demo_url, repo_url, live_url,
		 sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,
			COALESCE((SELECT MAX(sort_order)+1 FROM showcase_entries WHERE user_id=$1), 0))
		RETURNING ` + entryCols

	e, err := scanEntry(r.pool.QueryRow(ctx, q,
		userID, req.Title, req.Tagline, req.Problem, req.Solution, tsJSON, req.DemoURL, req.RepoURL, req.LiveURL))
	if err != nil {
		return nil, fmt.Errorf("showcase.CreateEntry: %w", err)
	}
	return e, nil
}

func (r *pgRepository) ListEntries(ctx context.Context, userID string) ([]*Entry, error) {
	const q = `SELECT ` + entryCols + ` FROM showcase_entries WHERE user_id=$1 ORDER BY sort_order ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("showcase.ListEntries: %w", err)
	}
	defer rows.Close()
	var out []*Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("showcase.ListEntries scan: %w", err)
		}
		out = append(out, e)
	}
	if out == nil {
		out = []*Entry{}
	}
	return out, rows.Err()
}

func (r *pgRepository) GetEntry(ctx context.Context, id, userID string) (*Entry, error) {
	const q = `SELECT ` + entryCols + ` FROM showcase_entries WHERE id=$1 AND user_id=$2`
	e, err := scanEntry(r.pool.QueryRow(ctx, q, id, userID))
	if err != nil {
		return nil, fmt.Errorf("showcase.GetEntry: %w", err)
	}
	return e, nil
}

func (r *pgRepository) UpdateEntry(ctx context.Context, id, userID string, req UpdateRequest) (*Entry, error) {
	var techJSON []byte
	var err error
	if req.TechStack != nil {
		techJSON, err = json.Marshal(*req.TechStack)
		if err != nil {
			return nil, fmt.Errorf("showcase: marshal tech_stack: %w", err)
		}
	}

	const q = `UPDATE showcase_entries SET
		title = COALESCE($3, title),
		tagline = COALESCE($4, tagline),
		problem = COALESCE($5, problem),
		solution = COALESCE($6, solution),
		tech_stack = COALESCE($7, tech_stack),
		demo_url = COALESCE($8, demo_url),
		repo_url = COALESCE($9, repo_url),
		live_url = COALESCE($10, live_url),
		updated_at = now()
		WHERE id=$1 AND user_id=$2
		RETURNING ` + entryCols

	e, err := scanEntry(r.pool.QueryRow(ctx, q,
		id, userID, req.Title, req.Tagline, req.Problem, req.Solution, nullableJSON(techJSON), req.DemoURL, req.RepoURL, req.LiveURL))
	if err != nil {
		return nil, fmt.Errorf("showcase.UpdateEntry: %w", err)
	}
	return e, nil
}

// nullableJSON returns nil (SQL NULL) for an empty byte slice so COALESCE
// leaves the column untouched when the field wasn't part of the update.
func nullableJSON(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return b
}

func (r *pgRepository) DeleteEntry(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM showcase_entries WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("showcase.DeleteEntry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) SetLogoKey(ctx context.Context, id, userID, key string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE showcase_entries SET logo_s3_key=$3, updated_at=now() WHERE id=$1 AND user_id=$2`, id, userID, key)
	if err != nil {
		return fmt.Errorf("showcase.SetLogoKey: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) SetBannerKey(ctx context.Context, id, userID, key string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE showcase_entries SET banner_s3_key=$3, updated_at=now() WHERE id=$1 AND user_id=$2`, id, userID, key)
	if err != nil {
		return fmt.Errorf("showcase.SetBannerKey: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearLogoKey reads the current logo key then clears it (two-step rather
// than RETURNING a subquery of the same row, whose same-statement visibility
// is unreliable).
func (r *pgRepository) ClearLogoKey(ctx context.Context, id, userID string) (string, error) {
	var key *string
	if scanErr := r.pool.QueryRow(ctx,
		`SELECT logo_s3_key FROM showcase_entries WHERE id=$1 AND user_id=$2`, id, userID).Scan(&key); scanErr != nil {
		return "", fmt.Errorf("showcase.ClearLogoKey: %w", scanErr)
	}
	tag, updErr := r.pool.Exec(ctx,
		`UPDATE showcase_entries SET logo_s3_key=NULL, updated_at=now() WHERE id=$1 AND user_id=$2`, id, userID)
	if updErr != nil {
		return "", fmt.Errorf("showcase.ClearLogoKey: %w", updErr)
	}
	if tag.RowsAffected() == 0 {
		return "", ErrNotFound
	}
	if key == nil {
		return "", nil
	}
	return *key, nil
}

func (r *pgRepository) ClearBannerKey(ctx context.Context, id, userID string) (string, error) {
	var key *string
	if scanErr := r.pool.QueryRow(ctx,
		`SELECT banner_s3_key FROM showcase_entries WHERE id=$1 AND user_id=$2`, id, userID).Scan(&key); scanErr != nil {
		return "", fmt.Errorf("showcase.ClearBannerKey: %w", scanErr)
	}
	tag, updErr := r.pool.Exec(ctx,
		`UPDATE showcase_entries SET banner_s3_key=NULL, updated_at=now() WHERE id=$1 AND user_id=$2`, id, userID)
	if updErr != nil {
		return "", fmt.Errorf("showcase.ClearBannerKey: %w", updErr)
	}
	if tag.RowsAffected() == 0 {
		return "", ErrNotFound
	}
	if key == nil {
		return "", nil
	}
	return *key, nil
}

func (r *pgRepository) GetEntryKeys(ctx context.Context, id string) (*string, *string, error) {
	var logoKey, bannerKey *string
	err := r.pool.QueryRow(ctx,
		`SELECT logo_s3_key, banner_s3_key FROM showcase_entries WHERE id=$1`, id).Scan(&logoKey, &bannerKey)
	if err != nil {
		return nil, nil, fmt.Errorf("showcase.GetEntryKeys: %w", err)
	}
	return logoKey, bannerKey, nil
}

// ListEntriesByUsername resolves a public slug: tries users.username first,
// falls back to matching users.id (UUID) so entries are shareable before a
// username is set. Returns the entries and the owner's display name.
func (r *pgRepository) ListEntriesByUsername(ctx context.Context, slug string) ([]*Entry, string, error) {
	const userQ = `SELECT id, COALESCE(display_name, email) FROM users WHERE username=$1 OR id::text=$1`
	var userID, displayName string
	if err := r.pool.QueryRow(ctx, userQ, slug).Scan(&userID, &displayName); err != nil {
		return nil, "", ErrNotFound
	}

	entries, err := r.ListEntries(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	return entries, displayName, nil
}

func (r *pgRepository) CreateToken(ctx context.Context, userID, name, hash, prefix string) (*ShowcaseToken, error) {
	const q = `INSERT INTO showcase_tokens (user_id, name, token_hash, token_prefix)
		VALUES ($1,$2,$3,$4)
		RETURNING id, user_id, name, token_prefix, last_used_at, created_at`
	t := &ShowcaseToken{}
	err := r.pool.QueryRow(ctx, q, userID, name, hash, prefix).
		Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.LastUsedAt, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("showcase.CreateToken: %w", err)
	}
	return t, nil
}

func (r *pgRepository) ListTokens(ctx context.Context, userID string) ([]*ShowcaseToken, error) {
	const q = `SELECT id, user_id, name, token_prefix, last_used_at, created_at
		FROM showcase_tokens WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("showcase.ListTokens: %w", err)
	}
	defer rows.Close()
	var out []*ShowcaseToken
	for rows.Next() {
		t := &ShowcaseToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.TokenPrefix, &t.LastUsedAt, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("showcase.ListTokens scan: %w", err)
		}
		out = append(out, t)
	}
	if out == nil {
		out = []*ShowcaseToken{}
	}
	return out, rows.Err()
}

func (r *pgRepository) DeleteToken(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM showcase_tokens WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("showcase.DeleteToken: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) FindTokenByHash(ctx context.Context, hash string) (*TokenLookupResult, error) {
	lr := &TokenLookupResult{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id FROM showcase_tokens WHERE token_hash=$1`, hash).
		Scan(&lr.ID, &lr.UserID)
	if err != nil {
		return nil, fmt.Errorf("showcase.FindTokenByHash: %w", err)
	}
	return lr, nil
}

func (r *pgRepository) UpdateTokenLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE showcase_tokens SET last_used_at=now() WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("showcase.UpdateTokenLastUsed: %w", err)
	}
	return nil
}
