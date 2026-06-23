package notes

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, userID string) ([]*Note, error)
	Get(ctx context.Context, id, userID string) (*Note, error)
	Create(ctx context.Context, userID string, req CreateRequest) (*Note, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*Note, error)
	Delete(ctx context.Context, id, userID string) error
	Reorder(ctx context.Context, userID string, items []ReorderItem) error

	LinkTask(ctx context.Context, noteID, taskID, userID string) error
	UnlinkTask(ctx context.Context, noteID, taskID, userID string) error
	ListTaskLinks(ctx context.Context, noteID, userID string) ([]string, error)
	ListByTask(ctx context.Context, taskID, userID string) ([]*NoteRef, error)

	SetBacklinks(ctx context.Context, srcID string, dstIDs []string) error
	Backlinks(ctx context.Context, noteID, userID string) ([]*NoteRef, error)
}

type ReorderItem struct {
	ID       string  `json:"id"`
	Position float64 `json:"position"`
	ParentID *string `json:"parent_id"`
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

const noteCols = `id, user_id, parent_id, title, body, plain, is_folder, icon, position, archived, created_at, updated_at`

func scanNote(row pgx.Row) (*Note, error) {
	n := &Note{}
	err := row.Scan(&n.ID, &n.UserID, &n.ParentID, &n.Title, &n.Body, &n.Plain,
		&n.IsFolder, &n.Icon, &n.Position, &n.Archived, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Note, error) {
	q := `SELECT ` + noteCols + ` FROM notes WHERE user_id=$1 AND archived=false
		ORDER BY position ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("notes: list: %w", err)
	}
	defer rows.Close()
	out := []*Note{}
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*Note, error) {
	q := `SELECT ` + noteCols + ` FROM notes WHERE id=$1 AND user_id=$2`
	n, err := scanNote(r.pool.QueryRow(ctx, q, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("notes: get: %w", err)
	}
	return n, nil
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*Note, error) {
	title := "Untitled"
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	}
	q := `INSERT INTO notes (user_id, parent_id, title, is_folder, icon,
		position)
		VALUES ($1,$2,$3,$4,$5, COALESCE((SELECT MAX(position)+1 FROM notes WHERE user_id=$1),0))
		RETURNING ` + noteCols
	n, err := scanNote(r.pool.QueryRow(ctx, q, userID, req.ParentID, title, req.IsFolder, req.Icon))
	if err != nil {
		return nil, fmt.Errorf("notes: create: %w", err)
	}
	return n, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Note, error) {
	sets := []string{"updated_at=now()"}
	args := []any{}
	idx := 1
	add := func(frag string, val any) {
		sets = append(sets, fmt.Sprintf(frag, idx))
		args = append(args, val)
		idx++
	}
	if req.Title != nil {
		add("title=$%d", *req.Title)
	}
	if req.Body != nil {
		add("body=$%d", *req.Body)
	}
	if req.Plain != nil {
		add("plain=$%d", *req.Plain)
	}
	if req.Icon != nil {
		add("icon=$%d", *req.Icon)
	}
	if req.Position != nil {
		add("position=$%d", *req.Position)
	}
	if req.Archived != nil {
		add("archived=$%d", *req.Archived)
	}
	if req.ParentID != nil {
		add("parent_id=$%d", *req.ParentID) // *req.ParentID is *string (may be nil → SQL NULL)
	}

	q := fmt.Sprintf(`UPDATE notes SET %s WHERE id=$%d AND user_id=$%d RETURNING `+noteCols,
		join(sets), idx, idx+1)
	args = append(args, id, userID)
	n, err := scanNote(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("notes: update: %w", err)
	}
	return n, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM notes WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("notes: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) Reorder(ctx context.Context, userID string, items []ReorderItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, it := range items {
		_, err := tx.Exec(ctx,
			`UPDATE notes SET position=$1, parent_id=$2, updated_at=now() WHERE id=$3 AND user_id=$4`,
			it.Position, it.ParentID, it.ID, userID)
		if err != nil {
			return fmt.Errorf("notes: reorder: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ── task links ──

func (r *pgRepository) LinkTask(ctx context.Context, noteID, taskID, userID string) error {
	const q = `INSERT INTO note_links (note_id, task_id)
		SELECT $1,$2 WHERE EXISTS(SELECT 1 FROM notes WHERE id=$1 AND user_id=$3)
		AND EXISTS(SELECT 1 FROM tasks WHERE id=$2 AND user_id=$3)
		ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, noteID, taskID, userID)
	return err
}

func (r *pgRepository) UnlinkTask(ctx context.Context, noteID, taskID, userID string) error {
	const q = `DELETE FROM note_links WHERE note_id=$1 AND task_id=$2
		AND EXISTS(SELECT 1 FROM notes WHERE id=$1 AND user_id=$3)`
	_, err := r.pool.Exec(ctx, q, noteID, taskID, userID)
	return err
}

func (r *pgRepository) ListTaskLinks(ctx context.Context, noteID, userID string) ([]string, error) {
	const q = `SELECT l.task_id FROM note_links l
		JOIN notes n ON n.id=l.note_id WHERE l.note_id=$1 AND n.user_id=$2`
	rows, err := r.pool.Query(ctx, q, noteID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *pgRepository) ListByTask(ctx context.Context, taskID, userID string) ([]*NoteRef, error) {
	const q = `SELECT n.id, n.title, n.icon FROM note_links l
		JOIN notes n ON n.id=l.note_id
		WHERE l.task_id=$1 AND n.user_id=$2 ORDER BY n.title`
	return r.scanRefs(ctx, q, taskID, userID)
}

// ── backlinks ──

func (r *pgRepository) SetBacklinks(ctx context.Context, srcID string, dstIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM note_backlinks WHERE src_id=$1`, srcID); err != nil {
		return err
	}
	for _, dst := range dstIDs {
		if dst == srcID {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO note_backlinks (src_id, dst_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			srcID, dst); err != nil {
			// ignore FK violations for non-existent targets
			continue
		}
	}
	return tx.Commit(ctx)
}

func (r *pgRepository) Backlinks(ctx context.Context, noteID, userID string) ([]*NoteRef, error) {
	const q = `SELECT n.id, n.title, n.icon FROM note_backlinks b
		JOIN notes n ON n.id=b.src_id
		WHERE b.dst_id=$1 AND n.user_id=$2 ORDER BY n.title`
	return r.scanRefs(ctx, q, noteID, userID)
}

func (r *pgRepository) scanRefs(ctx context.Context, q string, args ...any) ([]*NoteRef, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*NoteRef{}
	for rows.Next() {
		ref := &NoteRef{}
		if err := rows.Scan(&ref.ID, &ref.Title, &ref.Icon); err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}

func join(sets []string) string {
	out := ""
	for i, s := range sets {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
