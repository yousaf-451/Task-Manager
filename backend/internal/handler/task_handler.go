// Package handler is the HTTP layer. Handlers are intentionally thin: they
// decode requests, delegate to the service layer, and translate the result
// (or error) into an HTTP response. No SQL and no business rules live here.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/granet/task-manager/internal/middleware"
	"github.com/granet/task-manager/internal/models"
	"github.com/granet/task-manager/internal/service"
)

// currentUserID resolves the calling user's id from the request context.
// It is only ever set there by middleware.RequireAuth, which has already
// validated the session cookie before any of these handlers run (see
// routes.go) - so by the time a handler calls this, the id is trustworthy,
// unlike the old self-reported X-User-Id header this replaced.
func currentUserID(r *http.Request) uint64 {
	id, _ := middleware.UserIDFromContext(r.Context())
	return id
}

// TaskHandler wires HTTP requests to the TaskService.
type TaskHandler struct {
	service service.TaskService
}

// NewTaskHandler builds a TaskHandler backed by the given service.
func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] CreateTask - request received: %s %s", r.Method, r.URL.Path)

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	log.Printf("[HANDLER] CreateTask - decoded request body: %+v", req)

	log.Printf("[HANDLER] CreateTask - calling service.CreateTask()")
	task, err := h.service.CreateTask(r.Context(), currentUserID(r), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	log.Printf("[HANDLER] CreateTask - success! new task ID: %d", task.ID)

	writeSuccess(w, http.StatusCreated, toTaskResponse(task))
}

// ListTasks handles GET /api/tasks
// Supports optional query parameters: search, status, priority, category,
// favorite, includeArchived, sortBy, page, pageSize.
//
// The response is paginated: only `pageSize` rows are ever fetched from
// the database for a single call (see repository.TaskFilter), regardless
// of how many total tasks the user has. The response envelope's `data`
// field is `{ tasks: [...], pagination: {...} }` so the client always
// knows the current page, page size, total row count, and total pages.
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] ListTasks - request received: %s %s", r.Method, r.URL.String())

	q := r.URL.Query()

	query := service.TaskListQuery{
		Search:          q.Get("search"),
		Status:          q.Get("status"),
		Priority:        q.Get("priority"),
		Category:        q.Get("category"),
		SortBy:          q.Get("sortBy"),
		FavoriteOnly:    q.Get("favorite") == "true",
		IncludeArchived: q.Get("includeArchived") == "true",
		Page:            parsePositiveIntOrDefault(q.Get("page"), 1),
		PageSize:        parsePositiveIntOrDefault(q.Get("pageSize"), service.DefaultPageSize),
	}
	log.Printf("[HANDLER] ListTasks - parsed query params: %+v", query)

	log.Printf("[HANDLER] ListTasks - calling service.ListTasks()")
	tasks, pagination, err := h.service.ListTasks(r.Context(), currentUserID(r), query)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	log.Printf("[HANDLER] ListTasks - success! returned %d tasks (page %d)", len(tasks), pagination.Page)

	writeSuccess(w, http.StatusOK, toPaginatedTasksResponse(tasks, pagination))
}

// GetTask handles GET /api/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.GetTask(r.Context(), currentUserID(r), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, toTaskResponse(task))
}

// UpdateTask handles PUT /api/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := h.service.UpdateTask(r.Context(), currentUserID(r), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, toTaskResponse(task))
}

// DeleteTask handles DELETE /api/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.DeleteTask(r.Context(), currentUserID(r), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]uint64{"id": id})
}

// HealthCheck handles GET /health - used by uptime checks / docker healthcheck.
func (h *TaskHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ToggleFavorite handles PATCH /api/tasks/{id}/favorite
// Body: {"favorite": true|false}
func (h *TaskHandler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := h.service.SetFavorite(r.Context(), currentUserID(r), id, body.Favorite)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, toTaskResponse(task))
}

// ToggleArchive handles PATCH /api/tasks/{id}/archive
// Body: {"archived": true|false}
func (h *TaskHandler) ToggleArchive(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Archived bool `json:"archived"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := h.service.SetArchived(r.Context(), currentUserID(r), id, body.Archived)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, toTaskResponse(task))
}

// DuplicateTask handles POST /api/tasks/{id}/duplicate
func (h *TaskHandler) DuplicateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.service.DuplicateTask(r.Context(), currentUserID(r), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, toTaskResponse(task))
}

// bulkIDsRequest is the shared body shape for the bulk endpoints.
type bulkIDsRequest struct {
	IDs []uint64 `json:"ids"`
}

// BulkDelete handles POST /api/tasks/bulk/delete
// Body: {"ids": [1,2,3]}
func (h *TaskHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	var body bulkIDsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	count, err := h.service.BulkDelete(r.Context(), currentUserID(r), body.IDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]int64{"deleted": count})
}

// BulkComplete handles POST /api/tasks/bulk/complete
// Body: {"ids": [1,2,3]}
func (h *TaskHandler) BulkComplete(w http.ResponseWriter, r *http.Request) {
	var body bulkIDsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	count, err := h.service.BulkComplete(r.Context(), currentUserID(r), body.IDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]int64{"completed": count})
}

// Stats handles GET /api/tasks/stats - aggregate counts for the dashboard.
func (h *TaskHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats(r.Context(), currentUserID(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, stats)
}

// Categories handles GET /api/categories - distinct category values in use,
// used to populate the category filter dropdown on the frontend.
func (h *TaskHandler) Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.Categories(r.Context(), currentUserID(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, categories)
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

// parseID extracts and validates the {id} path parameter using the Go 1.22
// enhanced ServeMux path-value API (see routes/routes.go).
func parseID(r *http.Request) (uint64, error) {
	raw := r.PathValue("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

// parsePositiveIntOrDefault parses a query-string integer, falling back to
// fallback when it's missing, non-numeric, or not positive. Used for page
// and pageSize so a malformed value degrades gracefully instead of erroring.
func parsePositiveIntOrDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// handleServiceError maps errors returned by the service layer to the
// appropriate HTTP status code.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		log.Printf("handler: internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
