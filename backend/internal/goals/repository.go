package goals

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, userID string) ([]*Goal, error)
	Get(ctx context.Context, id, userID string) (*Goal, error)
	Create(ctx context.Context, userID string, req CreateGoalRequest) (*Goal, error)
	Update(ctx context.Context, id, userID string, req UpdateGoalRequest) (*Goal, error)
	Delete(ctx context.Context, id, userID string) error

	keyResults(ctx context.Context, goalID string) ([]*KeyResult, error)
	AddKR(ctx context.Context, goalID, userID string, req KRRequest) (*KeyResult, error)
	UpdateKR(ctx context.Context, krID, userID string, req KRRequest) (*KeyResult, error)
	DeleteKR(ctx context.Context, krID, userID string) error

	// taskCompletion returns done/total task counts for a goal.
	taskCompletion(ctx context.Context, goalID string) (done, total int, err error)
	owns(ctx context.Context, id, userID string) (bool, error)

	LinkTask(ctx context.Context, goalID, taskID, userID string) error
	UnlinkTask(ctx context.Context, taskID, userID string) error
	ListTasks(ctx context.Context, goalID, userID string) ([]*LinkedTask, error)
}

// LinkedTask is a compact task linked to a goal.
type LinkedTask struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

const goalCols = `id, user_id, title, description, status,
	to_char(target_date,'YYYY-MM-DD'), parent_id, position, created_at, updated_at`

func scanGoal(row pgx.Row) (*Goal, error) {
	g := &Goal{}
	err := row.Scan(&g.ID, &g.UserID, &g.Title, &g.Description, &g.Status,
		&g.TargetDate, &g.ParentID, &g.Position, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Goal, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+goalCols+` FROM goals WHERE user_id=$1
		ORDER BY position ASC, created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("goals: list: %w", err)
	}
	defer rows.Close()
	out := []*Goal{}
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*Goal, error) {
	g, err := scanGoal(r.pool.QueryRow(ctx, `SELECT `+goalCols+` FROM goals WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateGoalRequest) (*Goal, error) {
	q := `INSERT INTO goals (user_id, title, description, target_date, parent_id, position)
		VALUES ($1,$2,$3,$4,$5, COALESCE((SELECT MAX(position)+1 FROM goals WHERE user_id=$1),0))
		RETURNING ` + goalCols
	return scanGoal(r.pool.QueryRow(ctx, q, userID, req.Title, req.Description, req.TargetDate, req.ParentID))
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateGoalRequest) (*Goal, error) {
	sets := []string{"updated_at=now()"}
	args := []any{}
	idx := 1
	add := func(frag string, v any) {
		sets = append(sets, fmt.Sprintf(frag, idx))
		args = append(args, v)
		idx++
	}
	if req.Title != nil {
		add("title=$%d", *req.Title)
	}
	if req.Description != nil {
		add("description=$%d", *req.Description)
	}
	if req.Status != nil {
		add("status=$%d", *req.Status)
	}
	if req.TargetDate != nil {
		add("target_date=$%d", *req.TargetDate)
	}
	if req.Position != nil {
		add("position=$%d", *req.Position)
	}
	q := fmt.Sprintf(`UPDATE goals SET %s WHERE id=$%d AND user_id=$%d RETURNING `+goalCols,
		joinSets(sets), idx, idx+1)
	args = append(args, id, userID)
	g, err := scanGoal(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM goals WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) keyResults(ctx context.Context, goalID string) ([]*KeyResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, goal_id, title, metric_type, current_val, target_val, position
		 FROM key_results WHERE goal_id=$1 ORDER BY position ASC`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*KeyResult{}
	for rows.Next() {
		k := &KeyResult{}
		if err := rows.Scan(&k.ID, &k.GoalID, &k.Title, &k.MetricType, &k.CurrentVal, &k.TargetVal, &k.Position); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *pgRepository) AddKR(ctx context.Context, goalID, userID string, req KRRequest) (*KeyResult, error) {
	owns, err := r.owns(ctx, goalID, userID)
	if err != nil || !owns {
		return nil, ErrNotFound
	}
	mt := "percent"
	if req.MetricType != nil {
		mt = *req.MetricType
	}
	target := 100.0
	if req.TargetVal != nil {
		target = *req.TargetVal
	}
	cur := 0.0
	if req.CurrentVal != nil {
		cur = *req.CurrentVal
	}
	k := &KeyResult{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO key_results (goal_id, title, metric_type, current_val, target_val, position)
		 VALUES ($1,$2,$3,$4,$5, COALESCE((SELECT MAX(position)+1 FROM key_results WHERE goal_id=$1),0))
		 RETURNING id, goal_id, title, metric_type, current_val, target_val, position`,
		goalID, req.Title, mt, cur, target).
		Scan(&k.ID, &k.GoalID, &k.Title, &k.MetricType, &k.CurrentVal, &k.TargetVal, &k.Position)
	return k, err
}

func (r *pgRepository) UpdateKR(ctx context.Context, krID, userID string, req KRRequest) (*KeyResult, error) {
	// Ensure ownership via the parent goal.
	var ok bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM key_results k JOIN goals g ON g.id=k.goal_id
		 WHERE k.id=$1 AND g.user_id=$2)`, krID, userID).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	sets := []string{}
	args := []any{}
	idx := 1
	add := func(frag string, v any) {
		sets = append(sets, fmt.Sprintf(frag, idx))
		args = append(args, v)
		idx++
	}
	if req.Title != "" {
		add("title=$%d", req.Title)
	}
	if req.MetricType != nil {
		add("metric_type=$%d", *req.MetricType)
	}
	if req.CurrentVal != nil {
		add("current_val=$%d", *req.CurrentVal)
	}
	if req.TargetVal != nil {
		add("target_val=$%d", *req.TargetVal)
	}
	if len(sets) == 0 {
		return nil, ErrNotFound
	}
	q := fmt.Sprintf(`UPDATE key_results SET %s WHERE id=$%d
		RETURNING id, goal_id, title, metric_type, current_val, target_val, position`, joinSets(sets), idx)
	args = append(args, krID)
	k := &KeyResult{}
	err := r.pool.QueryRow(ctx, q, args...).
		Scan(&k.ID, &k.GoalID, &k.Title, &k.MetricType, &k.CurrentVal, &k.TargetVal, &k.Position)
	return k, err
}

func (r *pgRepository) DeleteKR(ctx context.Context, krID, userID string) error {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM key_results k USING goals g
		 WHERE k.id=$1 AND k.goal_id=g.id AND g.user_id=$2`, krID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) taskCompletion(ctx context.Context, goalID string) (done, total int, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE status='done'), COUNT(*) FROM tasks WHERE goal_id=$1`, goalID).
		Scan(&done, &total)
	return
}

func (r *pgRepository) LinkTask(ctx context.Context, goalID, taskID, userID string) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE tasks SET goal_id=$1 WHERE id=$2 AND user_id=$3
		 AND EXISTS(SELECT 1 FROM goals WHERE id=$1 AND user_id=$3)`,
		goalID, taskID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) UnlinkTask(ctx context.Context, taskID, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE tasks SET goal_id=NULL WHERE id=$1 AND user_id=$2`, taskID, userID)
	return err
}

func (r *pgRepository) ListTasks(ctx context.Context, goalID, userID string) ([]*LinkedTask, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.title, t.status FROM tasks t
		 JOIN goals g ON g.id=t.goal_id
		 WHERE t.goal_id=$1 AND g.user_id=$2 ORDER BY t.created_at`, goalID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*LinkedTask{}
	for rows.Next() {
		t := &LinkedTask{}
		if err := rows.Scan(&t.ID, &t.Title, &t.Status); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *pgRepository) owns(ctx context.Context, id, userID string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM goals WHERE id=$1 AND user_id=$2)`, id, userID).Scan(&ok)
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
