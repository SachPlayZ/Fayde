package attachments

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for attachments.
type Repository interface {
	Insert(ctx context.Context, taskID, userID, s3Key, filename, contentType string, sizeBytes int64) (*Attachment, error)
	ListByTask(ctx context.Context, taskID, userID string) ([]*Attachment, error)
	GetByID(ctx context.Context, id, taskID, userID string) (*Attachment, error)
	Delete(ctx context.Context, id, taskID, userID string) (s3Key string, err error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Insert creates a new attachment record.
func (r *pgRepository) Insert(ctx context.Context, taskID, userID, s3Key, filename, contentType string, sizeBytes int64) (*Attachment, error) {
	const q = `
		INSERT INTO task_attachments (task_id, user_id, s3_key, filename, content_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, user_id, s3_key, filename, content_type, size_bytes, created_at`

	a := &Attachment{}
	err := r.pool.QueryRow(ctx, q, taskID, userID, s3Key, filename, contentType, sizeBytes).
		Scan(&a.ID, &a.TaskID, &a.UserID, &a.S3Key, &a.Filename, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("attachments: insert: %w", err)
	}
	return a, nil
}

// ListByTask returns all attachments for a task owned by the given user.
func (r *pgRepository) ListByTask(ctx context.Context, taskID, userID string) ([]*Attachment, error) {
	const q = `
		SELECT id, task_id, user_id, s3_key, filename, content_type, size_bytes, created_at
		FROM task_attachments
		WHERE task_id = $1 AND user_id = $2
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("attachments: list by task: %w", err)
	}
	defer rows.Close()

	var list []*Attachment
	for rows.Next() {
		a := &Attachment{}
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.S3Key, &a.Filename, &a.ContentType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("attachments: scan: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attachments: rows: %w", err)
	}

	if list == nil {
		list = []*Attachment{}
	}
	return list, nil
}

// GetByID fetches a single attachment by id, taskID, and userID.
func (r *pgRepository) GetByID(ctx context.Context, id, taskID, userID string) (*Attachment, error) {
	const q = `
		SELECT id, task_id, user_id, s3_key, filename, content_type, size_bytes, created_at
		FROM task_attachments
		WHERE id = $1 AND task_id = $2 AND user_id = $3`

	a := &Attachment{}
	err := r.pool.QueryRow(ctx, q, id, taskID, userID).
		Scan(&a.ID, &a.TaskID, &a.UserID, &a.S3Key, &a.Filename, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("attachments: get by id: %w", err)
	}
	return a, nil
}

// Delete removes an attachment and returns its S3 key for subsequent S3 deletion.
func (r *pgRepository) Delete(ctx context.Context, id, taskID, userID string) (string, error) {
	const q = `
		DELETE FROM task_attachments
		WHERE id = $1 AND task_id = $2 AND user_id = $3
		RETURNING s3_key`

	var s3Key string
	err := r.pool.QueryRow(ctx, q, id, taskID, userID).Scan(&s3Key)
	if err != nil {
		return "", fmt.Errorf("attachments: delete: %w", err)
	}
	return s3Key, nil
}
