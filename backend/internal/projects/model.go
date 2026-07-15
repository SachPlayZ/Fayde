// Package projects lets a user publish hackathon/portfolio project entries,
// either on a public page (by username or ID) or via a public embed API for
// external sites and agents (authenticated with the same account API tokens
// used everywhere else).
package projects

import "time"

// TechItem is one entry in a project's tech-stack list.
type TechItem struct {
	Name      string `json:"name"`
	IsSponsor bool   `json:"is_sponsor"`
}

// Project is a single published project entry.
type Project struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Title     string     `json:"title"`
	Tagline   string     `json:"tagline"`
	Problem   string     `json:"problem"`
	Solution  string     `json:"solution"`
	TechStack []TechItem `json:"tech_stack"`
	DemoURL   *string    `json:"demo_url"`
	RepoURL   *string    `json:"repo_url"`
	LiveURL   *string    `json:"live_url"`
	LogoURL   *string    `json:"logo_url"`
	BannerURL *string    `json:"banner_url"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// logoS3Key/bannerS3Key are internal only — never serialized, used by the
	// service to build LogoURL/BannerURL and by the redirect handlers.
	logoS3Key   *string
	bannerS3Key *string
}

// CreateRequest is the body for POST /projects.
type CreateRequest struct {
	Title     string     `json:"title" validate:"required,min=1"`
	Tagline   string     `json:"tagline"`
	Problem   string     `json:"problem"`
	Solution  string     `json:"solution"`
	TechStack []TechItem `json:"tech_stack"`
	DemoURL   *string    `json:"demo_url"`
	RepoURL   *string    `json:"repo_url"`
	LiveURL   *string    `json:"live_url"`
}

// UpdateRequest is the body for PATCH /projects/{id}. Pointer fields are only
// applied when non-nil, so partial updates work.
type UpdateRequest struct {
	Title     *string     `json:"title"`
	Tagline   *string     `json:"tagline"`
	Problem   *string     `json:"problem"`
	Solution  *string     `json:"solution"`
	TechStack *[]TechItem `json:"tech_stack"`
	DemoURL   *string     `json:"demo_url"`
	RepoURL   *string     `json:"repo_url"`
	LiveURL   *string     `json:"live_url"`
}

// PublicProjects is the response for GET /p/{slug}.
type PublicProjects struct {
	OwnerDisplayName string     `json:"owner_display_name"`
	Projects         []*Project `json:"projects"`
}
