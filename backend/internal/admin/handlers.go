// Package admin implements admin-only endpoints.
package admin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminTask is a task record enriched with the owner's email.
type AdminTask struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	UserEmail   string     `json:"user_email"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AdminUser is a user record enriched with task count.
type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	TaskCount int       `json:"task_count"`
}

// AdminTaskListResult is the paginated response envelope for admin task lists.
type AdminTaskListResult struct {
	Data  []*AdminTask `json:"data"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
	Total int          `json:"total"`
}

// sortColumns is the whitelist of columns allowed in ORDER BY clauses for tasks.
var sortColumns = map[string]string{
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"due_date":   "t.due_date",
	"priority":   "t.priority",
}

// Handler handles HTTP requests for admin endpoints.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new admin Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// ListTasks handles GET /admin/tasks.
// Returns all tasks across all users with the owner's email.
// Supports the same filter/sort/paginate params as regular tasks.
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	status := q.Get("status")
	search := q.Get("search")
	sort := q.Get("sort")
	order := q.Get("order")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	args := []any{}
	conds := []string{}
	idx := 1

	if status != "" {
		conds = append(conds, fmt.Sprintf("t.status = $%d::task_status", idx))
		args = append(args, status)
		idx++
	}
	if search != "" {
		conds = append(conds, fmt.Sprintf("t.title ILIKE $%d", idx))
		args = append(args, "%"+search+"%")
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	col, ok := sortColumns[sort]
	if !ok {
		col = "t.created_at"
	}
	ord := "DESC"
	if strings.ToUpper(order) == "ASC" {
		ord = "ASC"
	}

	offset := (page - 1) * limit

	qStr := fmt.Sprintf(`
		SELECT t.id, t.user_id, u.email, t.title, t.description,
		       t.status, t.priority, t.due_date, t.created_at, t.updated_at,
		       COUNT(*) OVER() AS total_count
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		%s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d`, where, col, ord, idx, idx+1)

	args = append(args, limit, offset)

	rows, err := h.pool.Query(context.Background(), qStr, args...)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	defer rows.Close()

	var taskList []*AdminTask
	var total int

	for rows.Next() {
		t := &AdminTask{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.UserEmail, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
			&total); err != nil {
			httputil.Error(w, http.StatusInternalServerError, "failed to scan tasks")
			return
		}
		taskList = append(taskList, t)
	}
	if err := rows.Err(); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to iterate tasks")
		return
	}
	if taskList == nil {
		taskList = []*AdminTask{}
	}

	httputil.JSON(w, http.StatusOK, AdminTaskListResult{
		Data:  taskList,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// ListUsers handles GET /admin/users.
// Returns all users with id, email, role, created_at, and task_count.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	const q = `
		SELECT u.id, u.email, u.role, u.created_at,
		       COUNT(t.id) AS task_count
		FROM users u
		LEFT JOIN tasks t ON t.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at ASC`

	rows, err := h.pool.Query(context.Background(), q)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	var users []*AdminUser
	for rows.Next() {
		u := &AdminUser{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.TaskCount); err != nil {
			httputil.Error(w, http.StatusInternalServerError, "failed to scan users")
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to iterate users")
		return
	}
	if users == nil {
		users = []*AdminUser{}
	}

	httputil.JSON(w, http.StatusOK, users)
}
