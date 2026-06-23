package habits

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, userID string) ([]*Habit, error)
	Create(ctx context.Context, userID string, req CreateRequest) (*Habit, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*Habit, error)
	Delete(ctx context.Context, id, userID string) error
	Toggle(ctx context.Context, id, userID, date string) (bool, error)
	Logs(ctx context.Context, id, userID, from, to string) ([]*Log, error)
	logDates(ctx context.Context, id string) ([]time.Time, error)
	owns(ctx context.Context, id, userID string) (bool, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

const habitCols = `id, user_id, name, cadence, target_per_period, color, position, archived, created_at`

func scanHabit(row pgx.Row) (*Habit, error) {
	h := &Habit{}
	err := row.Scan(&h.ID, &h.UserID, &h.Name, &h.Cadence, &h.TargetPerPeriod,
		&h.Color, &h.Position, &h.Archived, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Habit, error) {
	q := `SELECT ` + habitCols + ` FROM habits WHERE user_id=$1 AND archived=false
		ORDER BY position ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("habits: list: %w", err)
	}
	defer rows.Close()
	out := []*Habit{}
	for rows.Next() {
		h, err := scanHabit(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*Habit, error) {
	cadence := req.Cadence
	if cadence == "" {
		cadence = "daily"
	}
	target := req.TargetPerPeriod
	if target <= 0 {
		target = 1
	}
	q := `INSERT INTO habits (user_id, name, cadence, target_per_period, color, position)
		VALUES ($1,$2,$3,$4,$5, COALESCE((SELECT MAX(position)+1 FROM habits WHERE user_id=$1),0))
		RETURNING ` + habitCols
	return scanHabit(r.pool.QueryRow(ctx, q, userID, req.Name, cadence, target, req.Color))
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Habit, error) {
	sets := []string{}
	args := []any{}
	idx := 1
	add := func(frag string, v any) {
		sets = append(sets, fmt.Sprintf(frag, idx))
		args = append(args, v)
		idx++
	}
	if req.Name != nil {
		add("name=$%d", *req.Name)
	}
	if req.Cadence != nil {
		add("cadence=$%d", *req.Cadence)
	}
	if req.TargetPerPeriod != nil {
		add("target_per_period=$%d", *req.TargetPerPeriod)
	}
	if req.Color != nil {
		add("color=$%d", *req.Color)
	}
	if req.Position != nil {
		add("position=$%d", *req.Position)
	}
	if req.Archived != nil {
		add("archived=$%d", *req.Archived)
	}
	if len(sets) == 0 {
		return r.get(ctx, id, userID)
	}
	q := fmt.Sprintf(`UPDATE habits SET %s WHERE id=$%d AND user_id=$%d RETURNING `+habitCols,
		joinSets(sets), idx, idx+1)
	args = append(args, id, userID)
	h, err := scanHabit(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return h, err
}

func (r *pgRepository) get(ctx context.Context, id, userID string) (*Habit, error) {
	h, err := scanHabit(r.pool.QueryRow(ctx, `SELECT `+habitCols+` FROM habits WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return h, err
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM habits WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Toggle flips a habit's completion for a date. Returns the new done state.
func (r *pgRepository) Toggle(ctx context.Context, id, userID, date string) (bool, error) {
	owns, err := r.owns(ctx, id, userID)
	if err != nil || !owns {
		return false, ErrNotFound
	}
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM habit_logs WHERE habit_id=$1 AND log_date=$2)`, id, date).
		Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		_, err = r.pool.Exec(ctx, `DELETE FROM habit_logs WHERE habit_id=$1 AND log_date=$2`, id, date)
		return false, err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO habit_logs (habit_id, log_date, count) VALUES ($1,$2,1)
		 ON CONFLICT (habit_id, log_date) DO NOTHING`, id, date)
	return true, err
}

func (r *pgRepository) Logs(ctx context.Context, id, userID, from, to string) ([]*Log, error) {
	owns, err := r.owns(ctx, id, userID)
	if err != nil || !owns {
		return nil, ErrNotFound
	}
	rows, err := r.pool.Query(ctx,
		`SELECT to_char(log_date,'YYYY-MM-DD'), count FROM habit_logs
		 WHERE habit_id=$1 AND log_date BETWEEN $2 AND $3 ORDER BY log_date`, id, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Log{}
	for rows.Next() {
		l := &Log{}
		if err := rows.Scan(&l.Date, &l.Count); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *pgRepository) logDates(ctx context.Context, id string) ([]time.Time, error) {
	rows, err := r.pool.Query(ctx, `SELECT log_date FROM habit_logs WHERE habit_id=$1 ORDER BY log_date DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *pgRepository) owns(ctx context.Context, id, userID string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM habits WHERE id=$1 AND user_id=$2)`, id, userID).Scan(&ok)
	return ok, err
}

func joinSets(sets []string) string {
	out := ""
	for i, s := range sets {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
