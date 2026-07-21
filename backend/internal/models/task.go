// Package models defines the domain entities shared across layers.
package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Status represents the lifecycle state of a task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// IsValid reports whether the status is one of the allowed enum values.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

// Priority represents how urgent a task is. Added in migration 0003;
// existing rows default to "medium" so the column is backward compatible
// with data written before this field existed.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// IsValid reports whether the priority is one of the allowed enum values.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

const dateLayout = "2006-01-02"
const defaultTaskColor = "#0e6b5c"

// Task is the core domain entity persisted in MySQL.
type Task struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"dueDate"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Category    string    `json:"category"`
	Color       string    `json:"color"`
	Favorite    bool      `json:"favorite"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Stats aggregates counts used by the dashboard. It is computed by the
// repository with a handful of cheap COUNT() queries rather than by
// fetching every row and counting in Go, so it stays fast as the table
// grows.
type Stats struct {
	Total             int `json:"total"`
	Completed         int `json:"completed"`
	Pending           int `json:"pending"`
	InProgress        int `json:"inProgress"`
	HighPriority      int `json:"highPriority"`
	Overdue           int `json:"overdue"`
	UpcomingWeek      int `json:"upcomingWeek"`
	Favorites         int `json:"favorites"`
	Archived          int `json:"archived"`
}

// MarshalDueDate formats the due date as YYYY-MM-DD for JSON responses.
// Used by handlers when building API responses (see handler/response.go).
func (t *Task) DueDateString() string {
	return t.DueDate.Format(dateLayout)
}

// ---------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------

// CreateTaskRequest is the payload accepted by POST /api/tasks.
// Priority, Category and Color are optional: omitting them (or sending an
// older client payload that predates these fields) falls back to sane
// defaults in validateCommon, so the endpoint stays backward compatible.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"` // expected format: YYYY-MM-DD
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Color       string `json:"color"`
}

// UpdateTaskRequest is the payload accepted by PUT /api/tasks/{id}.
// All fields are required; this is a full replace, which keeps the API
// predictable (clients always send the complete task).
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Color       string `json:"color"`
}

// Field length limits, mirrored from the column definitions in schema.sql /
// migrations/up/0001_init_schema.sql. Validating here means an overlong
// value fails with a clean 400 instead of surfacing as a raw MySQL
// "Data too long for column" error.
const (
	maxTitleLength       = 150
	maxDescriptionLength = 2000
	maxCategoryLength    = 60
)

// Validate checks a CreateTaskRequest and returns the first validation
// error encountered, or nil if the request is well formed. It also
// returns the parsed due date so callers don't need to re-parse it.
func (r *CreateTaskRequest) Validate() (time.Time, error) {
	return validateCommon(r.Title, r.Description, r.DueDate, r.Status)
}

// Validate checks an UpdateTaskRequest using the same rules as create.
func (r *UpdateTaskRequest) Validate() (time.Time, error) {
	return validateCommon(r.Title, r.Description, r.DueDate, r.Status)
}

// NormalizedPriority returns the request's priority, defaulting to
// "medium" when omitted, which is what pre-existing API clients do
// implicitly since they never send this field.
func (r *CreateTaskRequest) NormalizedPriority() (Priority, error) {
	return normalizePriority(r.Priority)
}

func (r *UpdateTaskRequest) NormalizedPriority() (Priority, error) {
	return normalizePriority(r.Priority)
}

func normalizePriority(raw string) (Priority, error) {
	if raw == "" {
		return PriorityMedium, nil
	}
	p := Priority(raw)
	if !p.IsValid() {
		return "", errors.New("priority must be one of: low, medium, high")
	}
	return p, nil
}

// NormalizedColor returns the request's color, falling back to the
// default accent when omitted or blank.
func (r *CreateTaskRequest) NormalizedColor() string { return normalizeColor(r.Color) }
func (r *UpdateTaskRequest) NormalizedColor() string { return normalizeColor(r.Color) }

func normalizeColor(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultTaskColor
	}
	return raw
}

func validateCommon(title, description, dueDate, status string) (time.Time, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return time.Time{}, errors.New("title is required")
	}
	if len(title) > maxTitleLength {
		return time.Time{}, fmt.Errorf("title must be %d characters or fewer", maxTitleLength)
	}

	if len(strings.TrimSpace(description)) > maxDescriptionLength {
		return time.Time{}, fmt.Errorf("description must be %d characters or fewer", maxDescriptionLength)
	}

	if dueDate == "" {
		return time.Time{}, errors.New("dueDate is required (format: YYYY-MM-DD)")
	}
	parsed, err := time.Parse(dateLayout, dueDate)
	if err != nil {
		return time.Time{}, errors.New("dueDate must be a valid date in YYYY-MM-DD format")
	}

	if status == "" {
		return time.Time{}, errors.New("status is required")
	}
	if !Status(status).IsValid() {
		return time.Time{}, errors.New("status must be one of: pending, in_progress, completed")
	}

	return parsed, nil
}
