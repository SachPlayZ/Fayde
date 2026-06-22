package projects

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, userID string, req CreateRequest) (*Project, error)
	List(ctx context.Context, userID string) ([]*Project, error)
	Get(ctx context.Context, id, userID string) (*Project, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*Project, error)
	Delete(ctx context.Context, id, userID string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*Project, error) {
	const q = `INSERT INTO projects (user_id, name, description, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, description, color, created_at`
	p := &Project{}
	err := r.pool.QueryRow(ctx, q, userID, req.Name, req.Description, req.Color).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Color, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.create: %w", err)
	}
	return p, nil
}

func (r *pgRepository) List(ctx context.Context, userID string) ([]*Project, error) {
	const q = `SELECT id, user_id, name, description, color, created_at
		FROM projects WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("projects.list: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Color, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("projects.list scan: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projects.list rows: %w", err)
	}
	return projects, nil
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*Project, error) {
	const q = `SELECT id, user_id, name, description, color, created_at
		FROM projects WHERE id=$1 AND user_id=$2`
	p := &Project{}
	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Color, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("projects.get: %w", err)
	}
	return p, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Project, error) {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Color != nil {
		sets = append(sets, fmt.Sprintf("color = $%d", argIdx))
		args = append(args, *req.Color)
		argIdx++
	}
	if len(sets) == 0 {
		return r.Get(ctx, id, userID)
	}

	args = append(args, id, userID)
	q := fmt.Sprintf(`UPDATE projects SET %s WHERE id=$%d AND user_id=$%d`,
		strings.Join(sets, ", "), argIdx, argIdx+1)

	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("projects.update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("projects.update: %w", ErrNotFound)
	}
	return r.Get(ctx, id, userID)
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("projects.delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("projects.delete: %w", ErrNotFound)
	}
	return nil
}
