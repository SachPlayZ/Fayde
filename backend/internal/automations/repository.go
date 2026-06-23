package automations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func scan(row pgx.Row) (*Automation, error) {
	a := &Automation{}
	var trig, conds, acts []byte
	if err := row.Scan(&a.ID, &a.UserID, &a.Name, &a.Enabled, &trig, &conds, &acts, &a.CreatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(trig, &a.Trigger)
	_ = json.Unmarshal(conds, &a.Conditions)
	_ = json.Unmarshal(acts, &a.Actions)
	if a.Conditions == nil {
		a.Conditions = []Condition{}
	}
	if a.Actions == nil {
		a.Actions = []Action{}
	}
	return a, nil
}

const cols = `id, user_id, name, enabled, trigger, conditions, actions, created_at`

func (r *Repository) List(ctx context.Context, userID string) ([]*Automation, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM automations WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Automation{}
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListEnabled returns enabled rules for a user — used by the engine.
func (r *Repository) ListEnabled(ctx context.Context, userID string) ([]*Automation, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM automations WHERE user_id=$1 AND enabled=true`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Automation{}
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) Create(ctx context.Context, userID string, req CreateRequest) (*Automation, error) {
	trig, _ := json.Marshal(req.Trigger)
	conds, _ := json.Marshal(req.Conditions)
	acts, _ := json.Marshal(req.Actions)
	return scan(r.pool.QueryRow(ctx,
		`INSERT INTO automations (user_id, name, trigger, conditions, actions)
		 VALUES ($1,$2,$3,$4,$5) RETURNING `+cols,
		userID, req.Name, trig, conds, acts))
}

func (r *Repository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Automation, error) {
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
	if req.Enabled != nil {
		add("enabled=$%d", *req.Enabled)
	}
	if req.Trigger != nil {
		b, _ := json.Marshal(*req.Trigger)
		add("trigger=$%d", b)
	}
	if req.Conditions != nil {
		b, _ := json.Marshal(*req.Conditions)
		add("conditions=$%d", b)
	}
	if req.Actions != nil {
		b, _ := json.Marshal(*req.Actions)
		add("actions=$%d", b)
	}
	if len(sets) == 0 {
		return nil, ErrNotFound
	}
	q := `UPDATE automations SET `
	for i, s := range sets {
		if i > 0 {
			q += ", "
		}
		q += s
	}
	q += fmt.Sprintf(` WHERE id=$%d AND user_id=$%d RETURNING `+cols, idx, idx+1)
	args = append(args, id, userID)
	a, err := scan(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM automations WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetTaskStatus / SetTaskPriority apply actions directly (bypassing tasks.Service
// to avoid re-triggering automations).
func (r *Repository) SetTaskStatus(ctx context.Context, taskID, userID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE tasks SET status=$1, updated_at=now() WHERE id=$2 AND user_id=$3`, status, taskID, userID)
	return err
}

func (r *Repository) SetTaskPriority(ctx context.Context, taskID, userID, priority string) error {
	_, err := r.pool.Exec(ctx, `UPDATE tasks SET priority=$1, updated_at=now() WHERE id=$2 AND user_id=$3`, priority, taskID, userID)
	return err
}
