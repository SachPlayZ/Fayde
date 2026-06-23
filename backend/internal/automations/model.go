package automations

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// Trigger describes what fires a rule.
type Trigger struct {
	Event string `json:"event"` // created | status_changed | updated
	To    string `json:"to"`    // optional status filter for status_changed
}

// Condition is a single AND-ed predicate on the task.
type Condition struct {
	Field string `json:"field"` // priority | status
	Op    string `json:"op"`    // eq | neq
	Value string `json:"value"`
}

// Action is something the rule does when it matches.
type Action struct {
	Type  string `json:"type"`  // set_priority | set_status | notify | webhook
	Value string `json:"value"` // new value / message / webhook url
	Kind  string `json:"kind"`  // for webhook: slack | discord
}

// Automation is a complete if-this-then-that rule.
type Automation struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	Name       string      `json:"name"`
	Enabled    bool        `json:"enabled"`
	Trigger    Trigger     `json:"trigger"`
	Conditions []Condition `json:"conditions"`
	Actions    []Action    `json:"actions"`
	CreatedAt  time.Time   `json:"created_at"`
}

type CreateRequest struct {
	Name       string      `json:"name" validate:"required,min=1"`
	Trigger    Trigger     `json:"trigger"`
	Conditions []Condition `json:"conditions"`
	Actions    []Action    `json:"actions"`
}

type UpdateRequest struct {
	Name       *string      `json:"name"`
	Enabled    *bool        `json:"enabled"`
	Trigger    *Trigger     `json:"trigger"`
	Conditions *[]Condition `json:"conditions"`
	Actions    *[]Action    `json:"actions"`
}
