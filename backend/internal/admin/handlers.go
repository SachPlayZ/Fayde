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

type analyticsResponse struct {
	ByStatus         map[string]int       `json:"by_status"`
	ByPriority       map[string]int       `json:"by_priority"`
	CompletionRate7d []dailyStat          `json:"completion_rate_7d"`
	OverdueByUser    []overdueUserStat    `json:"overdue_by_user"`
}

type dailyStat struct {
	Date    string `json:"date"`
	Done    int    `json:"done"`
	Created int    `json:"created"`
}

type overdueUserStat struct {
	UserEmail string `json:"user_email"`
	Count     int    `json:"count"`
	OldestDue string `json:"oldest_due"`
}

// Analytics handles GET /admin/analytics.
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := analyticsResponse{
		ByStatus:         map[string]int{"todo": 0, "in_progress": 0, "done": 0, "failed": 0},
		ByPriority:       map[string]int{"low": 0, "medium": 0, "high": 0},
		CompletionRate7d: []dailyStat{},
		OverdueByUser:    []overdueUserStat{},
	}

	// By status
	rows, err := h.pool.Query(ctx, `SELECT status::text, COUNT(*) FROM tasks GROUP BY status`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s string
			var c int
			if rows.Scan(&s, &c) == nil {
				resp.ByStatus[s] = c
			}
		}
	}

	// By priority
	rows2, err := h.pool.Query(ctx, `SELECT priority::text, COUNT(*) FROM tasks GROUP BY priority`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var p string
			var c int
			if rows2.Scan(&p, &c) == nil {
				resp.ByPriority[p] = c
			}
		}
	}

	// 7-day completion rate
	rows3, err := h.pool.Query(ctx, `
		WITH days AS (
			SELECT generate_series(
				(CURRENT_DATE - INTERVAL '6 days')::date,
				CURRENT_DATE::date,
				'1 day'::interval
			)::date AS d
		)
		SELECT days.d::text,
			COUNT(CASE WHEN t.status='done' AND t.updated_at::date=days.d THEN 1 END) AS done,
			COUNT(CASE WHEN t.created_at::date=days.d THEN 1 END) AS created
		FROM days LEFT JOIN tasks t ON true
		GROUP BY days.d ORDER BY days.d`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var ds dailyStat
			if rows3.Scan(&ds.Date, &ds.Done, &ds.Created) == nil {
				resp.CompletionRate7d = append(resp.CompletionRate7d, ds)
			}
		}
	}

	// Overdue by user
	rows4, err := h.pool.Query(ctx, `
		SELECT u.email, COUNT(t.id), MIN(t.due_date)::text
		FROM tasks t JOIN users u ON u.id=t.user_id
		WHERE t.due_date < now() AND t.status NOT IN ('done','failed')
		GROUP BY u.email ORDER BY COUNT(t.id) DESC`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var s overdueUserStat
			if rows4.Scan(&s.UserEmail, &s.Count, &s.OldestDue) == nil {
				resp.OverdueByUser = append(resp.OverdueByUser, s)
			}
		}
	}

	httputil.JSON(w, http.StatusOK, resp)
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
