package notes

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// Note is a foldered rich document.
type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ParentID  *string   `json:"parent_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"` // BlockNote JSON
	Plain     string    `json:"-"`    // derived plaintext, not exposed
	IsFolder  bool      `json:"is_folder"`
	Icon      *string   `json:"icon"`
	Position  float64   `json:"position"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteRef is a lightweight reference used for backlinks and link pickers.
type NoteRef struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Icon  *string `json:"icon"`
}

type CreateRequest struct {
	ParentID *string `json:"parent_id"`
	Title    *string `json:"title"`
	IsFolder bool    `json:"is_folder"`
	Icon     *string `json:"icon"`
}

type UpdateRequest struct {
	ParentID **string `json:"parent_id"` // double ptr: distinguish "set null" from "absent"
	Title    *string  `json:"title"`
	Body     *string  `json:"body"`
	Plain    *string  `json:"plain"`
	Icon     *string  `json:"icon"`
	Position *float64 `json:"position"`
	Archived *bool    `json:"archived"`
}
