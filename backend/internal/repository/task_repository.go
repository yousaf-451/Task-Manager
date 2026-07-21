// Package repository is the data-access layer. It is the only layer that
// knows about SQL; every other layer talks to it through the TaskRepository
// interface, which keeps the service layer testable and the database
// swappable.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/granet/task-manager/internal/models"
)

// ErrNotFound is returned when a task lookup does not match any row.
var ErrNotFound = errors.New("task not found")

// TaskFilter narrows down a ListTasks call. Empty fields are ignored.
type TaskFilter struct {
	Search          string          // matches against title or description (LIKE)
	Status          models.Status   // exact match on status when non-empty
	Priority        models.Priority // exact match on priority when non-empty
	Category        string          // exact match on category when non-empty
	FavoriteOnly    bool            // when true, only favorited tasks
	IncludeArchived bool            // when false (default), archived tasks are excluded
	SortBy          string          // "due_date_asc" | "due_date_desc" | "priority" | "created_at_desc" (default)
}

const taskColumns = `id, title, description, due_date, status, priority, category, color, favorite, archived, created_at, updated_at`

// TaskRepository defines the persistence operations available for tasks.
// Depending on an interface (rather than the concrete MySQL type) lets the
// service layer be unit tested with an in-memory fake if desired.
type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) (*models.Task, error)
	GetByID(ctx context.Context, id uint64) (*models.Task, error)
	List(ctx context.Context, filter TaskFilter) ([]*models.Task, error)
	Update(ctx context.Context, t *models.Task) (*models.Task, error)
	Delete(ctx context.Context, id uint64) error
	SetFavorite(ctx context.Context, id uint64, favorite bool) (*models.Task, error)
	SetArchived(ctx context.Context, id uint64, archived bool) (*models.Task, error)
	BulkDelete(ctx context.Context, ids []uint64) (int64, error)
	BulkSetStatus(ctx context.Context, ids []uint64, status models.Status) (int64, error)
	Stats(ctx context.Context) (*models.Stats, error)
	Categories(ctx context.Context) ([]string, error)
}

// mysqlTaskRepository is the MySQL-backed implementation of TaskRepository.
type mysqlTaskRepository struct {
	db *sql.DB
}

// NewMySQLTaskRepository builds a TaskRepository backed by the given
// *sql.DB connection pool.
func NewMySQLTaskRepository(db *sql.DB) TaskRepository {
	return &mysqlTaskRepository{db: db}
}

func (r *mysqlTaskRepository) Create(ctx context.Context, t *models.Task) (*models.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, due_date, status, priority, category, color)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.DueDate, t.Status, t.Priority, t.Category, t.Color)
	if err != nil {
		return nil, fmt.Errorf("repository: create task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: read last insert id: %w", err)
	}

	return r.GetByID(ctx, uint64(id))
}

func (r *mysqlTaskRepository) GetByID(ctx context.Context, id uint64) (*models.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get task by id: %w", err)
	}

	return t, nil
}

func (r *mysqlTaskRepository) List(ctx context.Context, filter TaskFilter) ([]*models.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE 1 = 1`
	args := make([]any, 0, 6)

	if !filter.IncludeArchived {
		query += " AND archived = FALSE"
	}

	if filter.Search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		like := "%" + filter.Search + "%"
		args = append(args, like, like)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Priority != "" {
		query += " AND priority = ?"
		args = append(args, filter.Priority)
	}

	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}

	if filter.FavoriteOnly {
		query += " AND favorite = TRUE"
	}

	switch filter.SortBy {
	case "due_date_asc":
		query += " ORDER BY due_date ASC, id ASC"
	case "due_date_desc":
		query += " ORDER BY due_date DESC, id ASC"
	case "priority":
		// FIELD() orders explicitly rather than relying on enum declaration
		// order, so this stays correct even if the enum values are reordered.
		query += " ORDER BY FIELD(priority, 'high', 'medium', 'low'), due_date ASC"
	default:
		query += " ORDER BY created_at DESC, id DESC"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repository: list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*models.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("repository: scan task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate task rows: %w", err)
	}

	return tasks, nil
}

func (r *mysqlTaskRepository) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	const query = `
		UPDATE tasks
		SET title = ?, description = ?, due_date = ?, status = ?, priority = ?, category = ?, color = ?
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.DueDate, t.Status, t.Priority, t.Category, t.Color, t.ID)
	if err != nil {
		return nil, fmt.Errorf("repository: update task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("repository: read rows affected: %w", err)
	}
	if rows == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, t.ID)
}

func (r *mysqlTaskRepository) Delete(ctx context.Context, id uint64) error {
	const query = `DELETE FROM tasks WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: read rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *mysqlTaskRepository) SetFavorite(ctx context.Context, id uint64, favorite bool) (*models.Task, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE tasks SET favorite = ? WHERE id = ?`, favorite, id)
	if err != nil {
		return nil, fmt.Errorf("repository: set favorite: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *mysqlTaskRepository) SetArchived(ctx context.Context, id uint64, archived bool) (*models.Task, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE tasks SET archived = ? WHERE id = ?`, archived, id)
	if err != nil {
		return nil, fmt.Errorf("repository: set archived: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *mysqlTaskRepository) BulkDelete(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args := inClauseQuery(`DELETE FROM tasks WHERE id IN (%s)`, ids)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("repository: bulk delete: %w", err)
	}
	return res.RowsAffected()
}

func (r *mysqlTaskRepository) BulkSetStatus(ctx context.Context, ids []uint64, status models.Status) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args := inClauseQuery(`UPDATE tasks SET status = ? WHERE id IN (%s)`, ids)
	args = append([]any{status}, args...)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("repository: bulk set status: %w", err)
	}
	return res.RowsAffected()
}

func (r *mysqlTaskRepository) Stats(ctx context.Context) (*models.Stats, error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(status = 'completed'), 0) AS completed,
			COALESCE(SUM(status = 'pending'), 0) AS pending,
			COALESCE(SUM(status = 'in_progress'), 0) AS in_progress,
			COALESCE(SUM(priority = 'high' AND status <> 'completed'), 0) AS high_priority_count,
			COALESCE(SUM(due_date < CURDATE() AND status <> 'completed'), 0) AS overdue,
			COALESCE(SUM(due_date BETWEEN CURDATE() AND CURDATE() + INTERVAL 7 DAY AND status <> 'completed'), 0) AS upcoming_week,
			COALESCE(SUM(favorite = TRUE), 0) AS favorites,
			COALESCE(SUM(archived = TRUE), 0) AS archived
		FROM tasks
		WHERE archived = FALSE
	`
	// Archived tasks are excluded from every count except the archived
	// count itself, which is why that one is computed separately below.
	var s models.Stats
	row := r.db.QueryRowContext(ctx, query)
	if err := row.Scan(
		&s.Total, &s.Completed, &s.Pending, &s.InProgress,
		&s.HighPriority, &s.Overdue, &s.UpcomingWeek, &s.Favorites, &s.Archived,
	); err != nil {
		return nil, fmt.Errorf("repository: stats: %w", err)
	}

	// The main query filters WHERE archived = FALSE, so s.Archived comes
	// back as 0. Fetch the true archived count with a second, separate
	// COUNT so the dashboard can still surface it.
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE archived = TRUE`).Scan(&s.Archived); err != nil {
		return nil, fmt.Errorf("repository: archived count: %w", err)
	}

	return &s, nil
}

func (r *mysqlTaskRepository) Categories(ctx context.Context) ([]string, error) {
	const query = `
		SELECT DISTINCT category FROM tasks
		WHERE category <> '' AND archived = FALSE
		ORDER BY category ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("repository: scan category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate categories: %w", err)
	}
	return categories, nil
}

// inClauseQuery builds a `WHERE id IN (?, ?, ...)` fragment safely (each id
// is still passed as a bound parameter, never string-concatenated).
func inClauseQuery(template string, ids []uint64) (string, []any) {
	placeholders := ""
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args[i] = id
	}
	return fmt.Sprintf(template, placeholders), args
}

// rowScanner abstracts over *sql.Row and *sql.Rows so scanTask can be
// shared by both GetByID (single row) and List (multiple rows).
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (*models.Task, error) {
	var t models.Task
	err := row.Scan(
		&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status,
		&t.Priority, &t.Category, &t.Color, &t.Favorite, &t.Archived,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
