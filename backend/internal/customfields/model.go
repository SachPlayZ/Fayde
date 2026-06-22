package customfields

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type FieldDefinition struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	FieldType string          `json:"field_type"`
	Options   json.RawMessage `json:"options,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type FieldValue struct {
	ID      string `json:"id"`
	TaskID  string `json:"task_id"`
	FieldID string `json:"field_id"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type CreateDefRequest struct {
	Name      string          `json:"name" validate:"required,min=1"`
	FieldType string          `json:"field_type" validate:"required,oneof=text number date select"`
	Options   json.RawMessage `json:"options"`
}

type SetValueRequest struct {
	Value string `json:"value" validate:"required"`
}
