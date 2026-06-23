// Package auth implements user authentication: registration, login, JWT issuance,
// and request-level middleware.
package auth

import (
	"encoding/json"
	"time"
)
// User represents an application user as stored in the database.
type User struct {
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	passwordHash  *string         // nil for OAuth-only users
	Role          string          `json:"role"`
	Theme         string          `json:"theme"`
	DigestEnabled bool            `json:"digest_enabled"`
	NotifPrefs    json.RawMessage `json:"notif_prefs"`
	ChatURL       *string         `json:"notif_chat_url"`
	ChatKind      *string         `json:"notif_chat_kind"`
	InboxToken    *string         `json:"inbox_token"`
	EmailVerified bool            `json:"email_verified"`
	Provider      string          `json:"provider"`
	ProviderID    *string         `json:"-"`
	CreatedAt     time.Time       `json:"created_at"`
	DisplayName   *string         `json:"display_name"`
	AvatarURL     *string         `json:"avatar_url"`
}

// PublicUser is the subset of User safe to include in API responses.
type PublicUser struct {
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	Role          string          `json:"role"`
	Theme         string          `json:"theme"`
	DigestEnabled bool            `json:"digest_enabled"`
	NotifPrefs    json.RawMessage `json:"notif_prefs"`
	ChatURL       *string         `json:"notif_chat_url"`
	ChatKind      *string         `json:"notif_chat_kind"`
	InboxToken    *string         `json:"inbox_token"`
	DisplayName   *string         `json:"display_name"`
	AvatarURL     *string         `json:"avatar_url"`
}

// authResponse is the response body for signup and login endpoints.
type authResponse struct {
	Token string     `json:"token"`
	User  PublicUser `json:"user"`
}
