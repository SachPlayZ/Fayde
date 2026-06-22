package customfields

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateDef(ctx context.Context, userID string, req CreateDefRequest) (*FieldDefinition, error)
	ListDefs(ctx context.Context, userID string) ([]*FieldDefinition, error)
	GetDef(ctx context.Context, id, userID string) (*FieldDefinition, error)
	DeleteDef(ctx context.Context, id, userID string) error
	SetValue(ctx context.Context, taskID, fieldID, value string) (*FieldValue, error)
	ListValues(ctx context.Context, taskID string) ([]*FieldValue, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

func (r *pgRepository) CreateDef(ctx context.Context, userID string, req CreateDefRequest) (*FieldDefinition, error) {
	const q = `INSERT INTO custom_field_definitions (id, user_id, name, field_type, options)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, user_id, name, field_type, options, created_at`
	id := uuid.New().String()
	d := &FieldDefinition{}
	err := r.pool.QueryRow(ctx, q, id, userID, req.Name, req.FieldType, req.Options).
		Scan(&d.ID, &d.UserID, &d.Name, &d.FieldType, &d.Options, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("customfields.CreateDef: %w", err)
	}
	return d, nil
}

func (r *pgRepository) ListDefs(ctx context.Context, userID string) ([]*FieldDefinition, error) {
	const q = `SELECT id, user_id, name, field_type, options, created_at
		FROM custom_field_definitions WHERE user_id=$1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("customfields.ListDefs: %w", err)
	}
	defer rows.Close()
	var out []*FieldDefinition
	for rows.Next() {
		d := &FieldDefinition{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.FieldType, &d.Options, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("customfields.ListDefs scan: %w", err)
		}
		out = append(out, d)
	}
	if out == nil {
		out = []*FieldDefinition{}
	}
	return out, rows.Err()
}

func (r *pgRepository) GetDef(ctx context.Context, id, userID string) (*FieldDefinition, error) {
	const q = `SELECT id, user_id, name, field_type, options, created_at
		FROM custom_field_definitions WHERE id=$1 AND user_id=$2`
	d := &FieldDefinition{}
	err := r.pool.QueryRow(ctx, q, id, userID).
		Scan(&d.ID, &d.UserID, &d.Name, &d.FieldType, &d.Options, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("customfields.GetDef: %w", err)
	}
	return d, nil
}

func (r *pgRepository) DeleteDef(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM custom_field_definitions WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("customfields.DeleteDef: %w", err)
	}
	return nil
}

func (r *pgRepository) SetValue(ctx context.Context, taskID, fieldID, value string) (*FieldValue, error) {
	const q = `INSERT INTO custom_field_values (id, task_id, field_id, value)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (task_id, field_id) DO UPDATE SET value=$4
		RETURNING id, task_id, field_id, value`
	id := uuid.New().String()
	fv := &FieldValue{}
	err := r.pool.QueryRow(ctx, q, id, taskID, fieldID, value).
		Scan(&fv.ID, &fv.TaskID, &fv.FieldID, &fv.Value)
	if err != nil {
		return nil, fmt.Errorf("customfields.SetValue: %w", err)
	}
	// Fetch name from definition
	r.pool.QueryRow(ctx, `SELECT name FROM custom_field_definitions WHERE id=$1`, fieldID).Scan(&fv.Name) //nolint
	return fv, nil
}

func (r *pgRepository) ListValues(ctx context.Context, taskID string) ([]*FieldValue, error) {
	const q = `SELECT v.id, v.task_id, v.field_id, d.name, v.value
		FROM custom_field_values v
		JOIN custom_field_definitions d ON d.id=v.field_id
		WHERE v.task_id=$1`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("customfields.ListValues: %w", err)
	}
	defer rows.Close()
	var out []*FieldValue
	for rows.Next() {
		fv := &FieldValue{}
		if err := rows.Scan(&fv.ID, &fv.TaskID, &fv.FieldID, &fv.Name, &fv.Value); err != nil {
			return nil, fmt.Errorf("customfields.ListValues scan: %w", err)
		}
		out = append(out, fv)
	}
	if out == nil {
		out = []*FieldValue{}
	}
	return out, rows.Err()
}
