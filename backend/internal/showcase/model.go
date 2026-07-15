// Package showcase lets a user publish hackathon/portfolio project entries,
// either on a public page (by username or ID) or via a dedicated read-only
// bearer-token API for embedding on external sites and agents.
package showcase

import "time"

// TechItem is one entry in a showcase entry's tech-stack list.
type TechItem struct {
	Name      string `json:"name"`
	IsSponsor bool   `json:"is_sponsor"`
}

// Entry is a single showcase project entry.
type Entry struct {
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

// CreateRequest is the body for POST /showcase.
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

// UpdateRequest is the body for PATCH /showcase/{id}. Pointer fields are only
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

// PublicEntries is the response for GET /p/{slug}.
type PublicEntries struct {
	OwnerDisplayName string   `json:"owner_display_name"`
	Entries          []*Entry `json:"entries"`
}

// ShowcaseToken is a bearer token scoped only to showcase read endpoints.
type ShowcaseToken struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateTokenResult is returned once on creation — includes the raw token.
type CreateTokenResult struct {
	Token string `json:"token"`
	ShowcaseToken
}

// CreateTokenRequest is the body for POST /settings/showcase-tokens.
type CreateTokenRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

// TokenLookupResult is used internally for token validation.
type TokenLookupResult struct {
	ID     string
	UserID string
}
