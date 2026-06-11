// Package attachments handles file upload/download for tasks via S3.
package attachments

import "time"

// Attachment represents a file attached to a task.
type Attachment struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	UserID      string    `json:"user_id"`
	S3Key       string    `json:"s3_key,omitempty"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	// URL is a transient pre-signed S3 URL — not stored in the database.
	URL string `json:"url"`
}
