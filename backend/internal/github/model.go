package github

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type GitHubLink struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	Repo        string    `json:"repo"`
	IssueNumber *int      `json:"issue_number"`
	PRNumber    *int      `json:"pr_number"`
	IssueURL    *string   `json:"issue_url"`
	PRURL       *string   `json:"pr_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type LinkRequest struct {
	Repo        string  `json:"repo" validate:"required"`
	IssueNumber *int    `json:"issue_number"`
	PRNumber    *int    `json:"pr_number"`
	IssueURL    *string `json:"issue_url"`
	PRURL       *string `json:"pr_url"`
}

// GitHubWebhookPayload represents incoming GitHub webhook for issues/PRs.
type GitHubWebhookPayload struct {
	Action string `json:"action"`
	Issue  *struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	} `json:"issue"`
	PullRequest *struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		Merged  bool   `json:"merged"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}
