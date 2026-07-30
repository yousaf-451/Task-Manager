package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/granet/task-manager/internal/models"
)

// envelope is the consistent JSON shape returned by every endpoint.
// Exactly one of Data / Error is populated.
type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// taskResponse mirrors models.Task but serializes DueDate as a plain
// YYYY-MM-DD string instead of a full RFC3339 timestamp, which is what the
// frontend's <input type="date"> expects.
type taskResponse struct {
	ID          uint64 `json:"id"`
	UserID      uint64 `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Color       string `json:"color"`
	Favorite    bool   `json:"favorite"`
	Archived    bool   `json:"archived"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toTaskResponse(t *models.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDateString(),
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		Category:    t.Category,
		Color:       t.Color,
		Favorite:    t.Favorite,
		Archived:    t.Archived,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toTaskResponseList(tasks []*models.Task) []taskResponse {
	out := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toTaskResponse(t))
	}
	return out
}

// paginatedTasksResponse is the `data` payload for GET /api/tasks. Keeping
// the task list and pagination metadata together in one object (rather
// than, say, a `X-Total-Count` header) means the frontend never has to
// guess whether more pages exist.
type paginatedTasksResponse struct {
	Tasks      []taskResponse    `json:"tasks"`
	Pagination models.Pagination `json:"pagination"`
}

func toPaginatedTasksResponse(tasks []*models.Task, pagination *models.Pagination) paginatedTasksResponse {
	return paginatedTasksResponse{
		Tasks:      toTaskResponseList(tasks),
		Pagination: *pagination,
	}
}

// userResponse mirrors models.User for JSON responses.
type userResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

func toUserResponse(u *models.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toUserResponseList(users []*models.User) []userResponse {
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, payload envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("handler: failed to encode JSON response: %v", err)
	}
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, envelope{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, envelope{Success: false, Error: message})
}
