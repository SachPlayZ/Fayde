// Package admin implements admin-only endpoints.
package admin

import (
	"context"
	"encoding/json"
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

// trackRequest is the body for POST /track.
type trackRequest struct {
	Path      string  `json:"path"`
	SessionID string  `json:"session_id"`
	UserID    *string `json:"user_id,omitempty"`
}

// TrackPageView handles POST /track.
// Public endpoint — no auth required.
func (h *Handler) TrackPageView(w http.ResponseWriter, r *http.Request) {
	var req trackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" || req.SessionID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	_, _ = h.pool.Exec(r.Context(),
		`INSERT INTO page_views (path, user_id, session_id) VALUES ($1, $2::uuid, $3)`,
		req.Path, req.UserID, req.SessionID,
	)
	w.WriteHeader(http.StatusNoContent)
}

type siteMetricsResponse struct {
	TotalUsers     int          `json:"total_users"`
	NewUsers       []dailyCount `json:"new_users"`
	TotalViews     int          `json:"total_views"`
	PageViews      []dailyViews `json:"page_views"`
	UniqueVisitors int          `json:"unique_visitors"`
	TopPages       []pageCount  `json:"top_pages"`
	ActiveUsers7d  int          `json:"active_users_7d"`
}

type dailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type dailyViews struct {
	Date   string `json:"date"`
	Views  int    `json:"views"`
	Unique int    `json:"unique"`
}

type pageCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// SiteMetrics handles GET /admin/site-metrics?range=7d|30d|90d.
func (h *Handler) SiteMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rangeDays := 30
	switch r.URL.Query().Get("range") {
	case "7d":
		rangeDays = 7
	case "90d":
		rangeDays = 90
	}

	resp := siteMetricsResponse{
		NewUsers:  []dailyCount{},
		PageViews: []dailyViews{},
		TopPages:  []pageCount{},
	}

	// Total users.
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&resp.TotalUsers)

	// New users per day in range.
	newUsersRows, err := h.pool.Query(ctx, `
		WITH days AS (
			SELECT generate_series(
				(CURRENT_DATE - ($1::int - 1) * INTERVAL '1 day')::date,
				CURRENT_DATE::date,
				'1 day'::interval
			)::date AS d
		)
		SELECT days.d::text, COUNT(u.id)
		FROM days
		LEFT JOIN users u ON u.created_at::date = days.d
		GROUP BY days.d ORDER BY days.d`, rangeDays)
	if err == nil {
		defer newUsersRows.Close()
		for newUsersRows.Next() {
			var dc dailyCount
			if newUsersRows.Scan(&dc.Date, &dc.Count) == nil {
				resp.NewUsers = append(resp.NewUsers, dc)
			}
		}
	}

	// Total page views in range.
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM page_views WHERE created_at >= now() - $1::int * INTERVAL '1 day'`, rangeDays).Scan(&resp.TotalViews)

	// Unique visitors (distinct session_ids) in range.
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT session_id) FROM page_views WHERE created_at >= now() - $1::int * INTERVAL '1 day'`, rangeDays).Scan(&resp.UniqueVisitors)

	// Page views + unique sessions per day in range.
	pvRows, err := h.pool.Query(ctx, `
		WITH days AS (
			SELECT generate_series(
				(CURRENT_DATE - ($1::int - 1) * INTERVAL '1 day')::date,
				CURRENT_DATE::date,
				'1 day'::interval
			)::date AS d
		)
		SELECT days.d::text,
			COUNT(pv.id) AS views,
			COUNT(DISTINCT pv.session_id) AS unique_sessions
		FROM days
		LEFT JOIN page_views pv ON pv.created_at::date = days.d
		GROUP BY days.d ORDER BY days.d`, rangeDays)
	if err == nil {
		defer pvRows.Close()
		for pvRows.Next() {
			var dv dailyViews
			if pvRows.Scan(&dv.Date, &dv.Views, &dv.Unique) == nil {
				resp.PageViews = append(resp.PageViews, dv)
			}
		}
	}

	// Top 8 pages in range.
	topRows, err := h.pool.Query(ctx, `
		SELECT path, COUNT(*) AS cnt
		FROM page_views
		WHERE created_at >= now() - $1::int * INTERVAL '1 day'
		GROUP BY path ORDER BY cnt DESC LIMIT 8`, rangeDays)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var pc pageCount
			if topRows.Scan(&pc.Path, &pc.Count) == nil {
				resp.TopPages = append(resp.TopPages, pc)
			}
		}
	}

	// Active users in last 7 days (always 7d regardless of range).
	_ = h.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM page_views
		WHERE user_id IS NOT NULL AND created_at >= now() - INTERVAL '7 days'`).Scan(&resp.ActiveUsers7d)

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
