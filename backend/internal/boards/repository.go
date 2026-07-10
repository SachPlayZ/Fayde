package boards

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for boards.
type Repository interface {
	// Board CRUD.
	CreateBoard(ctx context.Context, ownerID, name, description string) (*Board, error)
	GetBoard(ctx context.Context, id string) (*Board, error)
	UpdateBoard(ctx context.Context, id, name, description string) error
	DeleteBoard(ctx context.Context, id string) error
	ListBoardsByUser(ctx context.Context, userID string) ([]*Board, error)

	// Members.
	AddMember(ctx context.Context, boardID, userID, role string) error
	IsMember(ctx context.Context, boardID, userID string) (bool, error)
	GetMemberRole(ctx context.Context, boardID, userID string) (string, error)
	ListMembers(ctx context.Context, boardID string) ([]*BoardMember, error)

	// Board tasks.
	AddTask(ctx context.Context, boardID, title, createdBy string) (*BoardTask, error)
	GetTask(ctx context.Context, taskID string) (*BoardTask, error)
	ListTasks(ctx context.Context, boardID string) ([]*BoardTask, error)
	DeleteTask(ctx context.Context, taskID string) error

	// Completions.
	Complete(ctx context.Context, boardTaskID, userID string, date time.Time) (*Completion, error)
	Uncomplete(ctx context.Context, boardTaskID, userID string, date time.Time) error
	ListTodayCompletions(ctx context.Context, boardID string, date time.Time) ([]*Completion, error)

	// Invites.
	UpsertInvite(ctx context.Context, boardID, token, createdBy string) (*BoardInvite, error)
	GetInviteByToken(ctx context.Context, token string) (*BoardInvite, error)
	// GetInviteByBoard returns the board's current invite, or (nil, nil) if none exists.
	GetInviteByBoard(ctx context.Context, boardID string) (*BoardInvite, error)
	DeleteInvite(ctx context.Context, boardID string) error

	// Friendship check (direct SQL to avoid cross-package import).
	AreFriends(ctx context.Context, userA, userB string) (bool, error)
	// GetUserDisplay returns display name for a user.
	GetUserDisplay(ctx context.Context, userID string) (displayName string, err error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// ── Board CRUD ────────────────────────────────────────────────────────────────

func (r *pgRepository) CreateBoard(ctx context.Context, ownerID, name, description string) (*Board, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO boards (id, owner_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, owner_id, name, description, created_at
	`, id, ownerID, name, description)
	return scanBoard(row)
}

func (r *pgRepository) GetBoard(ctx context.Context, id string) (*Board, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, description, created_at
		FROM boards WHERE id = $1
	`, id)
	b, err := scanBoard(row)
	if err != nil && (errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows")) {
		return nil, ErrNotFound
	}
	return b, err
}

func scanBoard(row pgx.Row) (*Board, error) {
	var b Board
	if err := row.Scan(&b.ID, &b.OwnerID, &b.Name, &b.Description, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("boards: scan board: %w", err)
	}
	return &b, nil
}

func (r *pgRepository) UpdateBoard(ctx context.Context, id, name, description string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE boards SET name = $2, description = $3 WHERE id = $1
	`, id, name, description)
	return err
}

func (r *pgRepository) DeleteBoard(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM boards WHERE id = $1`, id)
	return err
}

func (r *pgRepository) ListBoardsByUser(ctx context.Context, userID string) ([]*Board, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id, b.owner_id, b.name, b.description, b.created_at,
		       (SELECT COUNT(*) FROM board_members bm2 WHERE bm2.board_id = b.id)
		FROM boards b
		JOIN board_members bm ON bm.board_id = b.id AND bm.user_id = $1
		ORDER BY b.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("boards.ListBoardsByUser: %w", err)
	}
	defer rows.Close()

	var out []*Board
	for rows.Next() {
		var b Board
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Name, &b.Description, &b.CreatedAt, &b.MemberCount); err != nil {
			return nil, fmt.Errorf("boards: scan board row: %w", err)
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

// ── Members ───────────────────────────────────────────────────────────────────

func (r *pgRepository) AddMember(ctx context.Context, boardID, userID, role string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO board_members (board_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (board_id, user_id) DO NOTHING
	`, boardID, userID, role)
	return err
}

func (r *pgRepository) IsMember(ctx context.Context, boardID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM board_members WHERE board_id = $1 AND user_id = $2)
	`, boardID, userID).Scan(&exists)
	return exists, err
}

func (r *pgRepository) GetMemberRole(ctx context.Context, boardID, userID string) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `
		SELECT role FROM board_members WHERE board_id = $1 AND user_id = $2
	`, boardID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotMember
	}
	return role, err
}

func (r *pgRepository) ListMembers(ctx context.Context, boardID string) ([]*BoardMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT bm.board_id, bm.user_id, bm.role, bm.joined_at,
		       u.email, u.display_name, u.avatar_url
		FROM board_members bm
		JOIN users u ON u.id = bm.user_id
		WHERE bm.board_id = $1
		ORDER BY bm.joined_at
	`, boardID)
	if err != nil {
		return nil, fmt.Errorf("boards.ListMembers: %w", err)
	}
	defer rows.Close()

	var out []*BoardMember
	for rows.Next() {
		var m BoardMember
		if err := rows.Scan(&m.BoardID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.Email, &m.DisplayName, &m.AvatarURL); err != nil {
			return nil, fmt.Errorf("boards: scan member: %w", err)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// ── Board tasks ───────────────────────────────────────────────────────────────

func (r *pgRepository) AddTask(ctx context.Context, boardID, title, createdBy string) (*BoardTask, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO board_tasks (id, board_id, title, sort_order, created_by)
		VALUES ($1, $2, $3,
		        COALESCE((SELECT MAX(sort_order)+1 FROM board_tasks WHERE board_id = $2), 0),
		        $4)
		RETURNING id, board_id, title, sort_order, created_by, created_at
	`, id, boardID, title, createdBy)
	return scanBoardTask(row)
}

func (r *pgRepository) GetTask(ctx context.Context, taskID string) (*BoardTask, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, board_id, title, sort_order, created_by, created_at
		FROM board_tasks WHERE id = $1
	`, taskID)
	t, err := scanBoardTask(row)
	if err != nil && (errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows")) {
		return nil, ErrNotFound
	}
	return t, err
}

func scanBoardTask(row pgx.Row) (*BoardTask, error) {
	var t BoardTask
	if err := row.Scan(&t.ID, &t.BoardID, &t.Title, &t.SortOrder, &t.CreatedBy, &t.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("boards: scan board task: %w", err)
	}
	return &t, nil
}

func (r *pgRepository) ListTasks(ctx context.Context, boardID string) ([]*BoardTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, board_id, title, sort_order, created_by, created_at
		FROM board_tasks WHERE board_id = $1
		ORDER BY sort_order, created_at
	`, boardID)
	if err != nil {
		return nil, fmt.Errorf("boards.ListTasks: %w", err)
	}
	defer rows.Close()

	var out []*BoardTask
	for rows.Next() {
		var t BoardTask
		if err := rows.Scan(&t.ID, &t.BoardID, &t.Title, &t.SortOrder, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("boards: scan task row: %w", err)
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (r *pgRepository) DeleteTask(ctx context.Context, taskID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM board_tasks WHERE id = $1`, taskID)
	return err
}

// ── Completions ───────────────────────────────────────────────────────────────

func (r *pgRepository) Complete(ctx context.Context, boardTaskID, userID string, date time.Time) (*Completion, error) {
	id := uuid.New().String()
	dateStr := date.UTC().Format("2006-01-02")
	row := r.pool.QueryRow(ctx, `
		INSERT INTO board_task_completions (id, board_task_id, user_id, completion_date)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (board_task_id, user_id, completion_date) DO UPDATE
		  SET completed_at = now()
		RETURNING id, board_task_id, user_id, completion_date::text, completed_at
	`, id, boardTaskID, userID, dateStr)

	var c Completion
	if err := row.Scan(&c.ID, &c.BoardTaskID, &c.UserID, &c.CompletionDate, &c.CompletedAt); err != nil {
		return nil, fmt.Errorf("boards.Complete: %w", err)
	}
	return &c, nil
}

func (r *pgRepository) Uncomplete(ctx context.Context, boardTaskID, userID string, date time.Time) error {
	dateStr := date.UTC().Format("2006-01-02")
	_, err := r.pool.Exec(ctx, `
		DELETE FROM board_task_completions
		WHERE board_task_id = $1 AND user_id = $2 AND completion_date = $3
	`, boardTaskID, userID, dateStr)
	return err
}

func (r *pgRepository) ListTodayCompletions(ctx context.Context, boardID string, date time.Time) ([]*Completion, error) {
	dateStr := date.UTC().Format("2006-01-02")
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.board_task_id, c.user_id, c.completion_date::text, c.completed_at
		FROM board_task_completions c
		JOIN board_tasks bt ON bt.id = c.board_task_id
		WHERE bt.board_id = $1 AND c.completion_date = $2
	`, boardID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("boards.ListTodayCompletions: %w", err)
	}
	defer rows.Close()

	var out []*Completion
	for rows.Next() {
		var c Completion
		if err := rows.Scan(&c.ID, &c.BoardTaskID, &c.UserID, &c.CompletionDate, &c.CompletedAt); err != nil {
			return nil, fmt.Errorf("boards: scan completion: %w", err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ── Invites ───────────────────────────────────────────────────────────────────

func (r *pgRepository) UpsertInvite(ctx context.Context, boardID, token, createdBy string) (*BoardInvite, error) {
	// Delete any existing invite for this board, then insert fresh.
	_, _ = r.pool.Exec(ctx, `DELETE FROM board_invites WHERE board_id = $1`, boardID)

	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO board_invites (id, board_id, token, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, board_id, token, created_by, expires_at, created_at
	`, id, boardID, token, createdBy)
	return scanInvite(row)
}

func (r *pgRepository) GetInviteByToken(ctx context.Context, token string) (*BoardInvite, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, board_id, token, created_by, expires_at, created_at
		FROM board_invites WHERE token = $1
	`, token)
	inv, err := scanInvite(row)
	if err != nil && (errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows")) {
		return nil, ErrInvalidToken
	}
	return inv, err
}

func scanInvite(row pgx.Row) (*BoardInvite, error) {
	var i BoardInvite
	if err := row.Scan(&i.ID, &i.BoardID, &i.Token, &i.CreatedBy, &i.ExpiresAt, &i.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("boards: scan invite: %w", err)
	}
	return &i, nil
}

func (r *pgRepository) GetInviteByBoard(ctx context.Context, boardID string) (*BoardInvite, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, board_id, token, created_by, expires_at, created_at
		FROM board_invites WHERE board_id = $1
	`, boardID)
	inv, err := scanInvite(row)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return nil, nil
		}
		return nil, err
	}
	return inv, nil
}

func (r *pgRepository) DeleteInvite(ctx context.Context, boardID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM board_invites WHERE board_id = $1`, boardID)
	return err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *pgRepository) AreFriends(ctx context.Context, userA, userB string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM friendships
			WHERE ((requester_id = $1 AND addressee_id = $2)
			    OR (requester_id = $2 AND addressee_id = $1))
			  AND status = 'accepted'
		)
	`, userA, userB).Scan(&exists)
	return exists, err
}

func (r *pgRepository) GetUserDisplay(ctx context.Context, userID string) (string, error) {
	var name *string
	var email string
	err := r.pool.QueryRow(ctx, `
		SELECT display_name, email FROM users WHERE id = $1
	`, userID).Scan(&name, &email)
	if err != nil {
		return "", err
	}
	if name != nil && *name != "" {
		return *name, nil
	}
	return email, nil
}
