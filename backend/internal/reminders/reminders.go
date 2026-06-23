// Package reminders implements custom per-task reminders delivered by the scheduler.
package reminders

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Reminder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TaskID    *string   `json:"task_id"`
	RemindAt  time.Time `json:"remind_at"`
	Note      string    `json:"note"`
	Sent      bool      `json:"sent"`
	CreatedAt time.Time `json:"created_at"`
}

// Due is a reminder ready to fire, joined with its task title.
type Due struct {
	ID        string
	UserID    string
	TaskID    *string
	Note      string
	TaskTitle string
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) Create(ctx context.Context, userID string, taskID *string, remindAt time.Time, note string) (*Reminder, error) {
	rm := &Reminder{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO reminders (user_id, task_id, remind_at, note) VALUES ($1,$2,$3,$4)
		 RETURNING id, user_id, task_id, remind_at, note, sent, created_at`,
		userID, taskID, remindAt, note).
		Scan(&rm.ID, &rm.UserID, &rm.TaskID, &rm.RemindAt, &rm.Note, &rm.Sent, &rm.CreatedAt)
	return rm, err
}

func (r *Repository) ListByTask(ctx context.Context, taskID, userID string) ([]*Reminder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, task_id, remind_at, note, sent, created_at FROM reminders
		 WHERE task_id=$1 AND user_id=$2 ORDER BY remind_at`, taskID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Reminder{}
	for rows.Next() {
		rm := &Reminder{}
		if err := rows.Scan(&rm.ID, &rm.UserID, &rm.TaskID, &rm.RemindAt, &rm.Note, &rm.Sent, &rm.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM reminders WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DuePending returns unsent reminders whose time has arrived and marks them sent.
func (r *Repository) DuePending(ctx context.Context, now time.Time) ([]*Due, error) {
	rows, err := r.pool.Query(ctx,
		`UPDATE reminders SET sent=true
		 WHERE id IN (SELECT id FROM reminders WHERE NOT sent AND remind_at <= $1 LIMIT 200)
		 RETURNING id, user_id, task_id, note`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Due
	for rows.Next() {
		d := &Due{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.TaskID, &d.Note); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Resolve task titles.
	for _, d := range out {
		if d.TaskID != nil {
			_ = r.pool.QueryRow(ctx, `SELECT title FROM tasks WHERE id=$1`, *d.TaskID).Scan(&d.TaskTitle)
		}
	}
	return out, nil
}

// Handler exposes reminder endpoints.
type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) ListByTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	items, err := h.repo.ListByTask(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list reminders")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	taskID := chi.URLParam(r, "id")
	var body struct {
		RemindAt time.Time `json:"remind_at"`
		Note     string    `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RemindAt.IsZero() {
		httputil.Error(w, http.StatusBadRequest, "remind_at required")
		return
	}
	rm, err := h.repo.Create(r.Context(), userID, &taskID, body.RemindAt, body.Note)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create reminder")
		return
	}
	httputil.JSON(w, http.StatusCreated, rm)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.repo.Delete(r.Context(), chi.URLParam(r, "reminderId"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "reminder not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete reminder")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
