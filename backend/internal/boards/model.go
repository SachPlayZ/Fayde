package boards

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrNotMember     = errors.New("not a member")
	ErrAlreadyMember = errors.New("already a member")
	ErrNotFriends    = errors.New("users are not friends")
	ErrInvalidToken  = errors.New("invalid or expired invite token")
)

// Board is the top-level board record.
type Board struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	MemberCount int       `json:"member_count"`
}

// BoardMember represents a member of a board.
type BoardMember struct {
	BoardID  string    `json:"board_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	// enriched
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// BoardTask is a task definition on a board (shared by all members).
type BoardTask struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Completion tracks one member's completion of a board task for a specific date.
type Completion struct {
	ID             string    `json:"id"`
	BoardTaskID    string    `json:"board_task_id"`
	UserID         string    `json:"user_id"`
	CompletionDate string    `json:"completion_date"` // YYYY-MM-DD
	CompletedAt    time.Time `json:"completed_at"`
}

// BoardDetail is the rich response for GET /boards/{id}.
type BoardDetail struct {
	Board
	Members     []*BoardMember     `json:"members"`
	Tasks       []*BoardTask       `json:"tasks"`
	Completions []*Completion      `json:"completions"` // today's completions for all members
	SharedTasks []*BoardSharedTask `json:"shared_tasks"`
	ShareToken  *string            `json:"share_token"`
}

// BoardSharedTask is a personal task a member has published to a board.
// The task stays owned/editable only by the sharer; other members see this
// read-only projection.
type BoardSharedTask struct {
	BoardID  string    `json:"board_id"`
	TaskID   string    `json:"task_id"`
	SharedBy string    `json:"shared_by"`
	SharedAt time.Time `json:"shared_at"`

	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`

	OwnerEmail       string  `json:"owner_email"`
	OwnerDisplayName *string `json:"owner_display_name"`
	OwnerAvatarURL   *string `json:"owner_avatar_url"`
}

// TaskBoardEntry is one board a given task has been shared to, from the
// task's point of view (used by the task-detail "Board Sharing" control).
type TaskBoardEntry struct {
	BoardID   string    `json:"board_id"`
	BoardName string    `json:"board_name"`
	TaskID    string    `json:"task_id"`
	SharedBy  string    `json:"shared_by"`
	SharedAt  time.Time `json:"shared_at"`
}

// ShareTaskInput holds the task to publish to a board.
type ShareTaskInput struct {
	TaskID string `json:"task_id" validate:"required"`
}

// SSESharedTaskUpdated is sent to board members when a shared task's fields change.
type SSESharedTaskUpdated struct {
	BoardID string `json:"board_id"`
	TaskID  string `json:"task_id"`
}

// BoardInvite is the invite-link record.
type BoardInvite struct {
	ID        string     `json:"id"`
	BoardID   string     `json:"board_id"`
	Token     string     `json:"token"`
	CreatedBy string     `json:"created_by"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// JoinPreview is returned by the public GET /boards/join/{token} route.
type JoinPreview struct {
	BoardName   string `json:"board_name"`
	MemberCount int    `json:"member_count"`
}

// CreateBoardInput holds input for creating a board.
type CreateBoardInput struct {
	Name        string `json:"name"        validate:"required,min=1,max=200"`
	Description string `json:"description"`
}

// UpdateBoardInput holds input for updating a board.
type UpdateBoardInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// AddTaskInput holds input for adding a task to a board.
type AddTaskInput struct {
	Title string `json:"title" validate:"required,min=1,max=500"`
}

// InviteFriendInput holds a friend's user ID to invite directly.
type InviteFriendInput struct {
	FriendUserID string `json:"friend_user_id" validate:"required"`
}

// SSEBoardTaskCompleted is the SSE payload sent when someone completes a board task.
type SSEBoardTaskCompleted struct {
	BoardID     string `json:"board_id"`
	BoardTaskID string `json:"board_task_id"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	TaskTitle   string `json:"task_title"`
	Date        string `json:"date"`
}
