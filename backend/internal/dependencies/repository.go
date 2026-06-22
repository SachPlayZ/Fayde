package dependencies

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, taskID, dependsOnID string) error
	Remove(ctx context.Context, taskID, dependsOnID string) error
	ListBlockedBy(ctx context.Context, taskID string) ([]Dependency, error)
	ListBlocking(ctx context.Context, taskID string) ([]Dependency, error)
	GetAllDependencies(ctx context.Context) ([][2]string, error)
	ListDependentsOf(ctx context.Context, doneTaskID string) ([]string, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) Add(ctx context.Context, taskID, dependsOnID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO task_dependencies (task_id, depends_on_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		taskID, dependsOnID)
	if err != nil {
		return fmt.Errorf("deps: add: %w", err)
	}
	return nil
}

func (r *pgRepository) Remove(ctx context.Context, taskID, dependsOnID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM task_dependencies WHERE task_id=$1 AND depends_on_id=$2`,
		taskID, dependsOnID)
	return err
}

// ListBlockedBy returns tasks that this task depends on (blockers).
func (r *pgRepository) ListBlockedBy(ctx context.Context, taskID string) ([]Dependency, error) {
	const q = `SELECT td.task_id, td.depends_on_id, t.title
		FROM task_dependencies td JOIN tasks t ON t.id=td.depends_on_id
		WHERE td.task_id=$1`
	return r.scanDeps(ctx, q, taskID)
}

// ListBlocking returns tasks that depend on this task (blocked tasks).
func (r *pgRepository) ListBlocking(ctx context.Context, taskID string) ([]Dependency, error) {
	const q = `SELECT td.task_id, td.depends_on_id, t.title
		FROM task_dependencies td JOIN tasks t ON t.id=td.task_id
		WHERE td.depends_on_id=$1`
	return r.scanDeps(ctx, q, taskID)
}

func (r *pgRepository) scanDeps(ctx context.Context, q, arg string) ([]Dependency, error) {
	rows, err := r.pool.Query(ctx, q, arg)
	if err != nil {
		return nil, fmt.Errorf("deps: list: %w", err)
	}
	defer rows.Close()
	var out []Dependency
	for rows.Next() {
		var d Dependency
		if err := rows.Scan(&d.TaskID, &d.DependsOnID, &d.Title); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []Dependency{}
	}
	return out, rows.Err()
}

func (r *pgRepository) GetAllDependencies(ctx context.Context) ([][2]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT task_id, depends_on_id FROM task_dependencies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][2]string
	for rows.Next() {
		var pair [2]string
		if err := rows.Scan(&pair[0], &pair[1]); err != nil {
			return nil, err
		}
		out = append(out, pair)
	}
	return out, rows.Err()
}

// ListDependentsOf returns task IDs that directly depend on the given task.
func (r *pgRepository) ListDependentsOf(ctx context.Context, doneTaskID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT task_id FROM task_dependencies WHERE depends_on_id=$1`, doneTaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
