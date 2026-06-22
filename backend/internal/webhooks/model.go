package webhooks

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type OutboundWebhook struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"secret,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRequest struct {
	Name   string   `json:"name" validate:"required,min=1"`
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
	Secret string   `json:"secret"`
}

type UpdateRequest struct {
	Name    *string  `json:"name"`
	URL     *string  `json:"url"`
	Events  []string `json:"events"`
	Secret  *string  `json:"secret"`
	Enabled *bool    `json:"enabled"`
}
