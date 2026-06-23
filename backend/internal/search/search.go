// Package search provides unified full-text search across tasks, notes and comments.
package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Result is a single search hit.
type Result struct {
	Type    string  `json:"type"`    // task | note | comment
	ID      string  `json:"id"`      // entity id (task id for comments)
	Title   string  `json:"title"`   // display title
	Snippet string  `json:"snippet"` // highlighted excerpt
	TaskID  *string `json:"task_id,omitempty"`
	Rank    float64 `json:"rank"`
}

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// Search runs a ranked websearch query across the user's tasks, notes and comments.
func (s *Service) Search(ctx context.Context, userID, query string) ([]*Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []*Result{}, nil
	}

	const q = `
WITH q AS (SELECT websearch_to_tsquery('english', $2) AS tsq)
SELECT * FROM (
	SELECT 'task' AS type, t.id::text AS id, t.title AS title,
		ts_headline('english', coalesce(t.description,''), q.tsq,
			'MaxFragments=1,MaxWords=12,MinWords=3') AS snippet,
		NULL::text AS task_id,
		ts_rank(t.search_tsv, q.tsq) AS rank
	FROM tasks t, q WHERE t.user_id=$1 AND t.search_tsv @@ q.tsq
	UNION ALL
	SELECT 'note', n.id::text, n.title,
		ts_headline('english', n.plain, q.tsq, 'MaxFragments=1,MaxWords=12,MinWords=3'),
		NULL::text,
		ts_rank(n.search_tsv, q.tsq)
	FROM notes n, q WHERE n.user_id=$1 AND n.archived=false AND n.search_tsv @@ q.tsq
	UNION ALL
	SELECT 'comment', c.id::text, t2.title,
		ts_headline('english', c.body, q.tsq, 'MaxFragments=1,MaxWords=12,MinWords=3'),
		c.task_id::text,
		ts_rank(c.search_tsv, q.tsq)
	FROM task_comments c
	JOIN tasks t2 ON t2.id=c.task_id
	, q
	WHERE t2.user_id=$1 AND c.search_tsv @@ q.tsq
) results
ORDER BY rank DESC
LIMIT 30`

	rows, err := s.pool.Query(ctx, q, userID, query)
	if err != nil {
		return nil, fmt.Errorf("search: query: %w", err)
	}
	defer rows.Close()
	out := []*Result{}
	for rows.Next() {
		r := &Result{}
		if err := rows.Scan(&r.Type, &r.ID, &r.Title, &r.Snippet, &r.TaskID, &r.Rank); err != nil {
			return nil, fmt.Errorf("search: scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Handler exposes GET /search?q=.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	results, err := h.svc.Search(r.Context(), userID, r.URL.Query().Get("q"))
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "search failed")
		return
	}
	httputil.JSON(w, http.StatusOK, results)
}
