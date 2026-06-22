package sharing

import (
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler handles HTTP requests for task sharing endpoints.
type Handler struct {
	svc  *Service
	pool *pgxpool.Pool
}

// NewHandler creates a new sharing Handler.
func NewHandler(svc *Service, pool *pgxpool.Pool) *Handler {
	return &Handler{svc: svc, pool: pool}
}

// CreateToken handles POST /tasks/{id}/share.
func (h *Handler) CreateToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")

	shareToken, err := h.svc.CreateToken(r.Context(), taskID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to create share token"})
		return
	}

	httputil.JSON(w, 201, map[string]string{
		"token": shareToken.Token,
		"url":   "/share/" + shareToken.Token,
	})
}

// RevokeToken handles DELETE /tasks/{id}/share.
func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")

	if err := h.svc.RevokeToken(r.Context(), taskID); err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to revoke share token"})
		return
	}

	httputil.JSON(w, 204, nil)
}

// GetToken handles GET /tasks/{id}/share.
func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	taskID := chi.URLParam(r, "id")

	shareToken, err := h.svc.GetByTaskID(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "no share token found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to get share token"})
		return
	}

	httputil.JSON(w, 200, map[string]string{"token": shareToken.Token})
}

// PublicView handles GET /share/{token} (no auth required).
func (h *Handler) PublicView(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	shareToken, err := h.svc.GetByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.JSON(w, 404, map[string]string{"error": "share token not found"})
			return
		}
		httputil.JSON(w, 500, map[string]string{"error": "failed to get share token"})
		return
	}

	// Fetch task directly via pool to avoid circular package dependency
	row := h.pool.QueryRow(r.Context(), `
		SELECT id, title, COALESCE(description, ''), status, priority, due_date
		FROM tasks WHERE id = $1
	`, shareToken.TaskID)

	var pt PublicTask
	if err := row.Scan(&pt.ID, &pt.Title, &pt.Description, &pt.Status, &pt.Priority, &pt.DueDate); err != nil {
		httputil.JSON(w, 404, map[string]string{"error": "task not found"})
		return
	}

	httputil.JSON(w, 200, pt)
}
