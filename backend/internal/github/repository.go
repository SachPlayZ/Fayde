package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for GitHub links.
type Repository interface {
	Link(ctx context.Context, taskID string, req LinkRequest) (*GitHubLink, error)
	Unlink(ctx context.Context, id, taskID string) error
	List(ctx context.Context, taskID string) ([]*GitHubLink, error)
	FindByIssue(ctx context.Context, repo string, issueNumber int) ([]*GitHubLink, error)
	FindByPR(ctx context.Context, repo string, prNumber int) ([]*GitHubLink, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a Postgres-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Link(ctx context.Context, taskID string, req LinkRequest) (*GitHubLink, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO github_links (id, task_id, repo, issue_number, pr_number, issue_url, pr_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, task_id, repo, issue_number, pr_number, issue_url, pr_url, created_at
	`, id, taskID, req.Repo, req.IssueNumber, req.PRNumber, req.IssueURL, req.PRURL)

	var l GitHubLink
	if err := row.Scan(&l.ID, &l.TaskID, &l.Repo, &l.IssueNumber, &l.PRNumber, &l.IssueURL, &l.PRURL, &l.CreatedAt); err != nil {
		return nil, fmt.Errorf("github.Link: %w", err)
	}
	return &l, nil
}

func (r *pgRepository) Unlink(ctx context.Context, id, taskID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM github_links WHERE id = $1 AND task_id = $2`, id, taskID)
	if err != nil {
		return fmt.Errorf("github.Unlink: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) List(ctx context.Context, taskID string) ([]*GitHubLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, repo, issue_number, pr_number, issue_url, pr_url, created_at
		FROM github_links WHERE task_id = $1 ORDER BY created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("github.List: %w", err)
	}
	defer rows.Close()

	var links []*GitHubLink
	for rows.Next() {
		var l GitHubLink
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Repo, &l.IssueNumber, &l.PRNumber, &l.IssueURL, &l.PRURL, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("github.List scan: %w", err)
		}
		links = append(links, &l)
	}
	return links, nil
}

func (r *pgRepository) FindByIssue(ctx context.Context, repo string, issueNumber int) ([]*GitHubLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, repo, issue_number, pr_number, issue_url, pr_url, created_at
		FROM github_links WHERE repo = $1 AND issue_number = $2
	`, repo, issueNumber)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, fmt.Errorf("github.FindByIssue: %w", err)
	}
	defer rows.Close()

	var links []*GitHubLink
	for rows.Next() {
		var l GitHubLink
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Repo, &l.IssueNumber, &l.PRNumber, &l.IssueURL, &l.PRURL, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("github.FindByIssue scan: %w", err)
		}
		links = append(links, &l)
	}
	return links, nil
}

func (r *pgRepository) FindByPR(ctx context.Context, repo string, prNumber int) ([]*GitHubLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, repo, issue_number, pr_number, issue_url, pr_url, created_at
		FROM github_links WHERE repo = $1 AND pr_number = $2
	`, repo, prNumber)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, fmt.Errorf("github.FindByPR: %w", err)
	}
	defer rows.Close()

	var links []*GitHubLink
	for rows.Next() {
		var l GitHubLink
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Repo, &l.IssueNumber, &l.PRNumber, &l.IssueURL, &l.PRURL, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("github.FindByPR scan: %w", err)
		}
		links = append(links, &l)
	}
	return links, nil
}
