// Package service contains the business logic of the application. It sits
// between the HTTP handlers and the repository: handlers never talk to the
// repository directly, and the repository never knows about HTTP.
package service

import (
	"context"
	"errors"
	"fmt"
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

const maxCategoryLength = 60

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
}

// TaskService defines the use cases exposed to the handler layer.
type TaskService interface {
	CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error)
	GetTask(ctx context.Context, id uint64) (*models.Task, error)
	ListTasks(ctx context.Context, query TaskListQuery) ([]*models.Task, error)
	UpdateTask(ctx context.Context, id uint64, req models.UpdateTaskRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, id uint64) error
	DuplicateTask(ctx context.Context, id uint64) (*models.Task, error)
	SetFavorite(ctx context.Context, id uint64, favorite bool) (*models.Task, error)
	SetArchived(ctx context.Context, id uint64, archived bool) (*models.Task, error)
	BulkDelete(ctx context.Context, ids []uint64) (int64, error)
	BulkComplete(ctx context.Context, ids []uint64) (int64, error)
	Stats(ctx context.Context) (*models.Stats, error)
	Categories(ctx context.Context) ([]string, error)
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService builds a TaskService backed by the given repository.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error) {
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
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		DueDate:     dueDate,
		Status:      models.Status(req.Status),
		Priority:    priority,
		Category:    category,
		Color:       req.NormalizedColor(),
	}

	return s.repo.Create(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, id uint64) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context, query TaskListQuery) ([]*models.Task, error) {
	filter := repository.TaskFilter{
		Search:          strings.TrimSpace(query.Search),
		SortBy:          query.SortBy,
		Category:        strings.TrimSpace(query.Category),
		FavoriteOnly:    query.FavoriteOnly,
		IncludeArchived: query.IncludeArchived,
	}

	if query.Status != "" {
		status := models.Status(query.Status)
		if !status.IsValid() {
			return nil, wrapValidation(errors.New("status filter must be one of: pending, in_progress, completed"))
		}
		filter.Status = status
	}

	if query.Priority != "" {
		priority := models.Priority(query.Priority)
		if !priority.IsValid() {
			return nil, wrapValidation(errors.New("priority filter must be one of: low, medium, high"))
		}
		filter.Priority = priority
	}

	return s.repo.List(ctx, filter)
}

func (s *taskService) UpdateTask(ctx context.Context, id uint64, req models.UpdateTaskRequest) (*models.Task, error) {
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

func (s *taskService) DeleteTask(ctx context.Context, id uint64) error {
	if id == 0 {
		return wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.Delete(ctx, id)
}

// DuplicateTask copies an existing task (fresh id, timestamps, and
// favorite/archived reset) so the user gets an independent working copy.
func (s *taskService) DuplicateTask(ctx context.Context, id uint64) (*models.Task, error) {
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	copy := &models.Task{
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

func (s *taskService) SetFavorite(ctx context.Context, id uint64, favorite bool) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.SetFavorite(ctx, id, favorite)
}

func (s *taskService) SetArchived(ctx context.Context, id uint64, archived bool) (*models.Task, error) {
	if id == 0 {
		return nil, wrapValidation(errors.New("id must be a positive integer"))
	}
	return s.repo.SetArchived(ctx, id, archived)
}

func (s *taskService) BulkDelete(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, wrapValidation(errors.New("ids must not be empty"))
	}
	return s.repo.BulkDelete(ctx, ids)
}

func (s *taskService) BulkComplete(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, wrapValidation(errors.New("ids must not be empty"))
	}
	return s.repo.BulkSetStatus(ctx, ids, models.StatusCompleted)
}

func (s *taskService) Stats(ctx context.Context) (*models.Stats, error) {
	return s.repo.Stats(ctx)
}

func (s *taskService) Categories(ctx context.Context) ([]string, error) {
	return s.repo.Categories(ctx)
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
