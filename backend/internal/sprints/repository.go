package sprints

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*Sprint, error)
	List(ctx context.Context, userID string) ([]*Sprint, error)
	Get(ctx context.Context, id, userID string) (*Sprint, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*Sprint, error)
	Delete(ctx context.Context, id, userID string) error
	AddTask(ctx context.Context, sprintID, taskID string) error
	RemoveTask(ctx context.Context, sprintID, taskID string) error
	ListTaskIDs(ctx context.Context, sprintID string) ([]string, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func scanSprint(row interface{ Scan(dest ...any) error }) (*Sprint, error) {
	s := &Sprint{}
	err := row.Scan(&s.ID, &s.UserID, &s.Name, &s.StartDate, &s.EndDate, &s.Goal, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*Sprint, error) {
	const q = `INSERT INTO sprints (user_id, name, start_date, end_date, goal)
		VALUES ($1, $2, $3::date, $4::date, $5)
		RETURNING id, user_id, name, start_date::text, end_date::text, goal, created_at`
	s, err := scanSprint(r.pool.QueryRow(ctx, q, userID, req.Name, req.StartDate, req.EndDate, req.Goal))
	if err != nil {
		return nil, fmt.Errorf("sprints.create: %w", err)
	}
	return s, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Sprint, error) {
	const q = `SELECT id, user_id, name, start_date::text, end_date::text, goal, created_at
		FROM sprints WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("sprints.list: %w", err)
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		s, err := scanSprint(rows)
		if err != nil {
			return nil, fmt.Errorf("sprints.list scan: %w", err)
		}
		sprints = append(sprints, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sprints.list rows: %w", err)
	}
	return sprints, nil
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*Sprint, error) {
	const q = `SELECT id, user_id, name, start_date::text, end_date::text, goal, created_at
		FROM sprints WHERE id=$1 AND user_id=$2`
	s, err := scanSprint(r.pool.QueryRow(ctx, q, id, userID))
	if err != nil {
		return nil, fmt.Errorf("sprints.get: %w", err)
	}
	return s, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Sprint, error) {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.StartDate != nil {
		sets = append(sets, fmt.Sprintf("start_date = $%d::date", argIdx))
		args = append(args, *req.StartDate)
		argIdx++
	}
	if req.EndDate != nil {
		sets = append(sets, fmt.Sprintf("end_date = $%d::date", argIdx))
		args = append(args, *req.EndDate)
		argIdx++
	}
	if req.Goal != nil {
		sets = append(sets, fmt.Sprintf("goal = $%d", argIdx))
		args = append(args, *req.Goal)
		argIdx++
	}
	if len(sets) == 0 {
		return r.Get(ctx, id, userID)
	}

	args = append(args, id, userID)
	q := fmt.Sprintf(`UPDATE sprints SET %s WHERE id=$%d AND user_id=$%d`,
		strings.Join(sets, ", "), argIdx, argIdx+1)

	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("sprints.update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("sprints.update: %w", ErrNotFound)
	}
	return r.Get(ctx, id, userID)
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sprints WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("sprints.delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("sprints.delete: %w", ErrNotFound)
	}
	return nil
}

func (r *pgRepository) AddTask(ctx context.Context, sprintID, taskID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sprint_tasks (sprint_id, task_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		sprintID, taskID)
	if err != nil {
		return fmt.Errorf("sprints.add_task: %w", err)
	}
	return nil
}

func (r *pgRepository) RemoveTask(ctx context.Context, sprintID, taskID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM sprint_tasks WHERE sprint_id=$1 AND task_id=$2`,
		sprintID, taskID)
	if err != nil {
		return fmt.Errorf("sprints.remove_task: %w", err)
	}
	return nil
}

func (r *pgRepository) ListTaskIDs(ctx context.Context, sprintID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT task_id FROM sprint_tasks WHERE sprint_id=$1`, sprintID)
	if err != nil {
		return nil, fmt.Errorf("sprints.list_task_ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("sprints.list_task_ids scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sprints.list_task_ids rows: %w", err)
	}
	return ids, nil
}
