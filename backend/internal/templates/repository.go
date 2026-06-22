package templates

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*TaskTemplate, error)
	List(ctx context.Context, userID string) ([]*TaskTemplate, error)
	Get(ctx context.Context, id, userID string) (*TaskTemplate, error)
	Delete(ctx context.Context, id, userID string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func scanTemplate(row interface{ Scan(dest ...any) error }) (*TaskTemplate, error) {
	t := &TaskTemplate{}
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Title, &t.Description, &t.Status, &t.Priority, &t.EffortPoints, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*TaskTemplate, error) {
	status := req.Status
	if status == "" {
		status = "todo"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	const q = `INSERT INTO task_templates (user_id, name, title, description, status, priority, effort_points)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, title, description, status, priority, effort_points, created_at`
	t, err := scanTemplate(r.pool.QueryRow(ctx, q, userID, req.Name, req.Title, req.Description, status, priority, req.EffortPoints))
	if err != nil {
		return nil, fmt.Errorf("templates.create: %w", err)
	}
	return t, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*TaskTemplate, error) {
	const q = `SELECT id, user_id, name, title, description, status, priority, effort_points, created_at
		FROM task_templates WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("templates.list: %w", err)
	}
	defer rows.Close()

	var templates []*TaskTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, fmt.Errorf("templates.list scan: %w", err)
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("templates.list rows: %w", err)
	}
	return templates, nil
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*TaskTemplate, error) {
	const q = `SELECT id, user_id, name, title, description, status, priority, effort_points, created_at
		FROM task_templates WHERE id=$1 AND user_id=$2`
	t, err := scanTemplate(r.pool.QueryRow(ctx, q, id, userID))
	if err != nil {
		return nil, fmt.Errorf("templates.get: %w", err)
	}
	return t, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM task_templates WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("templates.delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("templates.delete: %w", ErrNotFound)
	}
	return nil
}
