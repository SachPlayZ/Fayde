package tasks

import "time"

// Tag is a lightweight tag summary embedded in Task responses.
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Task represents a task record as stored in the database.
type Task struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	DueDate          *time.Time `json:"due_date"`
	Recurrence       *string    `json:"recurrence"`
	RecurrenceEnd    *time.Time `json:"recurrence_end"`
	ParentTaskID     *string    `json:"parent_task_id"`
	AssigneeID       *string    `json:"assignee_id"`
	AssigneeEmail    *string    `json:"assignee_email"`
	ExternalEventID  *string    `json:"external_event_id"`
	SortOrder        float64    `json:"sort_order"`
	EffortPoints     *int       `json:"effort_points"`
	ProjectID        *string    `json:"project_id"`
	ProjectName      *string    `json:"project_name"`
	TotalTimeSeconds int        `json:"total_time_seconds"`
	Tags             []Tag      `json:"tags"`
	SubtaskCount     int        `json:"subtask_count"`
	SubtasksDone     int        `json:"subtasks_done"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateRequest contains the fields for creating a new task.
type CreateRequest struct {
	Title         string     `json:"title"          validate:"required,min=1"`
	Description   string     `json:"description"`
	Status        string     `json:"status"         validate:"omitempty,oneof=todo in_progress done failed"`
	Priority      string     `json:"priority"       validate:"omitempty,oneof=low medium high"`
	DueDate       *time.Time `json:"due_date"`
	Recurrence    *string    `json:"recurrence"     validate:"omitempty,oneof=daily weekly monthly"`
	RecurrenceEnd *time.Time `json:"recurrence_end"`
	AssigneeID    *string    `json:"assignee_id"`
	EffortPoints  *int       `json:"effort_points"`
	ProjectID     *string    `json:"project_id"`
}

// UpdateRequest contains the fields for a partial task update.
// Pointer fields are only updated when non-nil.
type UpdateRequest struct {
	Title         *string    `json:"title"          validate:"omitempty,min=1"`
	Description   *string    `json:"description"`
	Status        *string    `json:"status"         validate:"omitempty,oneof=todo in_progress done failed"`
	Priority      *string    `json:"priority"       validate:"omitempty,oneof=low medium high"`
	DueDate       *time.Time `json:"due_date"`
	Recurrence    *string    `json:"recurrence"     validate:"omitempty,oneof=daily weekly monthly"`
	RecurrenceEnd *time.Time `json:"recurrence_end"`
	AssigneeID    *string    `json:"assignee_id"`
	SortOrder     *float64   `json:"sort_order"`
	EffortPoints  *int       `json:"effort_points"`
	ProjectID     *string    `json:"project_id"`
}

// ListParams describes filters, sorting, and pagination for listing tasks.
type ListParams struct {
	Status    string
	Search    string
	Sort      string
	Order     string
	Page      int
	Limit     int
	ProjectID string
}

// BulkUpdateRequest updates status/priority for multiple tasks.
type BulkUpdateRequest struct {
	IDs      []string `json:"ids"`
	Status   *string  `json:"status"   validate:"omitempty,oneof=todo in_progress done failed"`
	Priority *string  `json:"priority" validate:"omitempty,oneof=low medium high"`
}

// BulkDeleteRequest deletes multiple tasks.
type BulkDeleteRequest struct {
	IDs []string `json:"ids"`
}

// ReorderItem is a single task + new sort_order pair.
type ReorderItem struct {
	ID        string  `json:"id"`
	SortOrder float64 `json:"sort_order"`
}

// ListResult is the paginated response envelope for task lists.
type ListResult struct {
	Data  []*Task `json:"data"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
	Total int     `json:"total"`
}
