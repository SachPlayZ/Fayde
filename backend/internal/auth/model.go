// Package auth implements user authentication: registration, login, JWT issuance,
// and request-level middleware.
package auth

import "time"

// User represents an application user as stored in the database.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// PublicUser is the subset of User safe to include in API responses.
type PublicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// authResponse is the response body for signup and login endpoints.
type authResponse struct {
	Token string     `json:"token"`
	User  PublicUser `json:"user"`
}
