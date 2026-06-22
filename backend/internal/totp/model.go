package totp

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidCode = errors.New("invalid TOTP code")
var ErrAlreadyEnabled = errors.New("TOTP already enabled")

type TOTPSecret struct {
	UserID    string    `json:"user_id"`
	Secret    string    `json:"-"` // never expose in JSON
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type SetupResponse struct {
	Secret string `json:"secret"`
	QRURL  string `json:"qr_url"`
}
