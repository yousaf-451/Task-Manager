// Package service contains the business logic of the application. It sits
// between the HTTP handlers and the repository: handlers never talk to the
// repository directly, and the repository never knows about HTTP.
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/granet/task-manager/internal/models"
	"github.com/granet/task-manager/internal/repository"
)

// ErrValidation is wrapped around any input validation failure so handlers
// can distinguish "bad request" from "internal error" using errors.Is.
var ErrValidation = errors.New("validation error")

// ErrNotFound is re-exported so handlers only need to import the service
// package, not the repository package, to check for a 404 condition.
var ErrNotFound = repository.ErrNotFound
// ErrConflict is wrapped around errors that mean "this request can't
// succeed because it conflicts with existing data" (e.g. signing up with
// an email that's already registered), so handlers can map it to HTTP 409.
var ErrConflict = errors.New("conflict")

const maxCategoryLength = 60

// Pagination bounds, re-exported here so the handler layer doesn't need to
// import the repository package just to clamp a query parameter.
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// TaskListQuery captures the optional query-string parameters the
// dashboard can send when listing tasks.
type TaskListQuery struct {
	Search          string
	Status          string
	Priority        string
	Category        string
	FavoriteOnly    bool
	IncludeArchived bool
	SortBy          string

	// Page is 1-indexed; PageSize <= 0 falls back to DefaultPageSize and
	// is clamped to MaxPageSize by the repository.
	Page     int
	PageSize int
}

// TaskService defines the use cases exposed to the handler layer. Every
// method takes a userID (the caller's identity, resolved by the handler
// from the request) and scopes the operation to that user's own tasks.
type TaskService interface {
	CreateTask(ctx context.Context, userID uint64, req models.CreateTaskRequest) (*models.Task, error)
	GetTask(ctx context.Context, userID, id uint64) (*models.Task, error)
	ListTasks(ctx context.Context, userID uint64, query TaskListQuery) ([]*models.Task, *models.Pagination, error)
	UpdateTask(ctx context.Context, userID, id uint64, req models.UpdateTaskRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, userID, id uint64) error
	DuplicateTask(ctx context.Context, userID, id uint64) (*models.Task, error)
	SetFavorite(ctx context.Context, userID, id uint64, favorite bool) (*models.Task, error)
	SetArchived(ctx context.Context, userID, id uint64, archived bool) (*models.Task, error)
	BulkDelete(ctx context.Context, userID uint64, ids []uint64) (int64, error)
	BulkComplete(ctx context.Context, userID uint64, ids []uint64) (int64, error)
	Stats(ctx context.Context, userID uint64) (*models.Stats, error)
	Categories(ctx context.Context, userID uint64) ([]string, error)
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService builds a TaskService backed by the given repository.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, userID uint64, req models.CreateTaskRequest) (*models.Task, error) {
	log.Printf("[SERVICE] CreateTask - validating request for user %d, title=%q", userID, req.Title)

	dueDate, err := req.Validate()
	if err != nil {
		return nil, wrapValidation(err)
	}
	priority, err := req.NormalizedPriority()
	if err != nil {
		return nil, wrapValidation(err)
	}
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, wrapValidation(err)
	}

	task := &models.Task{
		UserID:      userID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		DueDate:     dueDate,
		Status:      models.Status(req.Status),
		Priority:    priority,
		Category:    category,
		Color:       req.NormalizedColor(),
	}

	log.Printf("[SERVICE] CreateTask - validation passed, calling repository.Create()")
	created, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}
	log.Printf("[SERVICE] CreateTask - repository returned task with ID %d", created.ID)
	return created, nil
}

func (s *taskService) GetTask(ctx context.Context, userID, id uint64) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *taskService) ListTasks(ctx context.Context, userID uint64, query TaskListQuery) ([]*models.Task, *models.Pagination, error) {
	log.Printf("[SERVICE] ListTasks - building filter for user %d, page=%d pageSize=%d", userID, query.Page, query.PageSize)

	filter := repository.TaskFilter{
		UserID:          userID,
		Search:          strings.TrimSpace(query.Search),
		SortBy:          query.SortBy,
		Category:        strings.TrimSpace(query.Category),
		FavoriteOnly:    query.FavoriteOnly,
		IncludeArchived: query.IncludeArchived,
		Page:            query.Page,
		PageSize:        query.PageSize,
	}

	if query.Status != "" {
		status := models.Status(query.Status)
		if !status.IsValid() {
			return nil, nil, wrapValidation(errors.New("status filter must be one of: pending, in_progress, completed"))
		}
		filter.Status = status
	}

	if query.Priority != "" {
		priority := models.Priority(query.Priority)
		if !priority.IsValid() {
			return nil, nil, wrapValidation(errors.New("priority filter must be one of: low, medium, high"))
		}
		filter.Priority = priority
	}

	// Normalize page/pageSize the same way the repository will, so the
	// pagination metadata we return always matches what was actually
	// queried (rather than echoing back whatever the client asked for).
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	filter.Page = page
	filter.PageSize = pageSize

	log.Printf("[SERVICE] ListTasks - calling repository.List() with filter: %+v", filter)
	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("[SERVICE] ListTasks - repository returned %d tasks", len(tasks))

	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &models.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	return tasks, pagination, nil
}

func (s *taskService) UpdateTask(ctx context.Context, userID, id uint64, req models.UpdateTaskRequest) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}

	dueDate, err := req.Validate()
	if err != nil {
		return nil, wrapValidation(err)
	}
	priority, err := req.NormalizedPriority()
	if err != nil {
		return nil, wrapValidation(err)
	}
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, wrapValidation(err)
	}

	task := &models.Task{
		ID:          id,
		UserID:      userID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		DueDate:     dueDate,
		Status:      models.Status(req.Status),
		Priority:    priority,
		Category:    category,
		Color:       req.NormalizedColor(),
	}

	return s.repo.Update(ctx, task)
}

func (s *taskService) DeleteTask(ctx context.Context, userID, id uint64) error {
	if id == 0 {
		return wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.Delete(ctx, userID, id)
}

// DuplicateTask copies an existing task (fresh id, timestamps, and
// favorite/archived reset) so the user gets an independent working copy.
func (s *taskService) DuplicateTask(ctx context.Context, userID, id uint64) (*models.Task, error) {
	original, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	copy := &models.Task{
		UserID:      userID,
		Title:       original.Title + " (copy)",
		Description: original.Description,
		DueDate:     original.DueDate,
		Status:      models.StatusPending,
		Priority:    original.Priority,
		Category:    original.Category,
		Color:       original.Color,
	}
	return s.repo.Create(ctx, copy)
}

func (s *taskService) SetFavorite(ctx context.Context, userID, id uint64, favorite bool) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.SetFavorite(ctx, userID, id, favorite)
}

func (s *taskService) SetArchived(ctx context.Context, userID, id uint64, archived bool) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.SetArchived(ctx, userID, id, archived)
}

func (s *taskService) BulkDelete(ctx context.Context, userID uint64, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, wrapValidation(errors.New("ids must not be empty"))
	}
	return s.repo.BulkDelete(ctx, userID, ids)
}

func (s *taskService) BulkComplete(ctx context.Context, userID uint64, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, wrapValidation(errors.New("ids must not be empty"))
	}
	return s.repo.BulkSetStatus(ctx, userID, ids, models.StatusCompleted)
}

func (s *taskService) Stats(ctx context.Context, userID uint64) (*models.Stats, error) {
	return s.repo.Stats(ctx, userID)
}

func (s *taskService) Categories(ctx context.Context, userID uint64) ([]string, error) {
	return s.repo.Categories(ctx, userID)
}

func normalizeCategory(raw string) (string, error) {
	category := strings.TrimSpace(raw)
	if len(category) > maxCategoryLength {
		return "", fmt.Errorf("category must be %d characters or fewer", maxCategoryLength)
	}
	return category, nil
}

// wrapValidation joins a specific validation message with the sentinel
// ErrValidation so callers can do: errors.Is(err, service.ErrValidation).
func wrapValidation(err error) error {
	return errors.Join(ErrValidation, err)
}
