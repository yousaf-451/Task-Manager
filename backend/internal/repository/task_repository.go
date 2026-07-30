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
	"log"

	"github.com/granet/task-manager/internal/models"
)

// ErrNotFound is returned when a task lookup does not match any row, OR
// when it matches a row that belongs to a different user. The two cases
// are indistinguishable on purpose: a user should get a 404, not a 403,
// for tasks that exist but aren't theirs (this avoids leaking whether a
// given task id exists at all to someone who doesn't own it).
var ErrNotFound = errors.New("task not found")

// defaultPageSize / maxPageSize bound how many rows a single List call can
// return, so a client can never accidentally (or deliberately) ask for
// the entire table in one request.
const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// TaskFilter narrows down a ListTasks call. Empty fields are ignored.
type TaskFilter struct {
	UserID          uint64          // required: every list is scoped to one user
	Search          string          // matches against title or description (LIKE)
	Status          models.Status   // exact match on status when non-empty
	Priority        models.Priority // exact match on priority when non-empty
	Category        string          // exact match on category when non-empty
	FavoriteOnly    bool            // when true, only favorited tasks
	IncludeArchived bool            // when false (default), archived tasks are excluded
	SortBy          string          // "due_date_asc" | "due_date_desc" | "priority" | "created_at_desc" (default)

	// Pagination. Page is 1-indexed; PageSize <= 0 falls back to
	// defaultPageSize and is clamped to maxPageSize.
	Page     int
	PageSize int
}

// normalizedPageSize returns the effective page size after applying
// defaults/clamping, without mutating the filter.
func (f TaskFilter) normalizedPageSize() int {
	if f.PageSize <= 0 {
		return defaultPageSize
	}
	if f.PageSize > maxPageSize {
		return maxPageSize
	}
	return f.PageSize
}

// normalizedPage returns the effective 1-indexed page number.
func (f TaskFilter) normalizedPage() int {
	if f.Page <= 0 {
		return 1
	}
	return f.Page
}

// offset returns the SQL OFFSET for this filter's page/pageSize.
func (f TaskFilter) offset() int {
	return (f.normalizedPage() - 1) * f.normalizedPageSize()
}

const taskColumns = `id, user_id, title, description, due_date, status, priority, category, color, favorite, archived, created_at, updated_at`

// TaskRepository defines the persistence operations available for tasks.
// Every method that reads/writes a specific task takes a userID and scopes
// the query to `WHERE user_id = ?` (in addition to any id filter), so one
// user can never see, edit, or delete another user's tasks - this is
// enforced at the data-access layer, not just in the UI.
//
// Depending on an interface (rather than the concrete MySQL type) lets the
// service layer be unit tested with an in-memory fake if desired.
type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) (*models.Task, error)
	GetByID(ctx context.Context, userID, id uint64) (*models.Task, error)
	List(ctx context.Context, filter TaskFilter) ([]*models.Task, error)
	Count(ctx context.Context, filter TaskFilter) (int, error)
	Update(ctx context.Context, t *models.Task) (*models.Task, error)
	Delete(ctx context.Context, userID, id uint64) error
	SetFavorite(ctx context.Context, userID, id uint64, favorite bool) (*models.Task, error)
	SetArchived(ctx context.Context, userID, id uint64, archived bool) (*models.Task, error)
	BulkDelete(ctx context.Context, userID uint64, ids []uint64) (int64, error)
	BulkSetStatus(ctx context.Context, userID uint64, ids []uint64, status models.Status) (int64, error)
	Stats(ctx context.Context, userID uint64) (*models.Stats, error)
	Categories(ctx context.Context, userID uint64) ([]string, error)
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
		INSERT INTO tasks (user_id, title, description, due_date, status, priority, category, color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	log.Printf("[REPOSITORY] Create - running SQL INSERT for title=%q", t.Title)
	res, err := r.db.ExecContext(ctx, query, t.UserID, t.Title, t.Description, t.DueDate, t.Status, t.Priority, t.Category, t.Color)
	if err != nil {
		return nil, fmt.Errorf("repository: create task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: read last insert id: %w", err)
	}
	log.Printf("[REPOSITORY] Create - insert successful, new row ID: %d", id)

	return r.GetByID(ctx, t.UserID, uint64(id))
}

func (r *mysqlTaskRepository) GetByID(ctx context.Context, userID, id uint64) (*models.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE id = ? AND user_id = ?`
	row := r.db.QueryRowContext(ctx, query, id, userID)

	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: get task by id: %w", err)
	}

	return t, nil
}

// whereClause builds the shared WHERE fragment (and args) used by both
// List and Count, so the two queries can never drift out of sync with
// each other - that would make the reported totalPages wrong.
func (f TaskFilter) whereClause() (string, []any) {
	query := "WHERE user_id = ?"
	args := make([]any, 0, 6)
	args = append(args, f.UserID)

	if !f.IncludeArchived {
		query += " AND archived = FALSE"
	}

	if f.Search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}

	if f.Status != "" {
		query += " AND status = ?"
		args = append(args, f.Status)
	}

	if f.Priority != "" {
		query += " AND priority = ?"
		args = append(args, f.Priority)
	}

	if f.Category != "" {
		query += " AND category = ?"
		args = append(args, f.Category)
	}

	if f.FavoriteOnly {
		query += " AND favorite = TRUE"
	}

	return query, args
}

func (r *mysqlTaskRepository) List(ctx context.Context, filter TaskFilter) ([]*models.Task, error) {
	where, args := filter.whereClause()
	query := `SELECT ` + taskColumns + ` FROM tasks ` + where

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

	// Pagination: only ever fetch one page's worth of rows from the
	// database, never the whole table. LIMIT/OFFSET values are bound as
	// parameters like everything else, not string-concatenated.
	query += " LIMIT ? OFFSET ?"
	args = append(args, filter.normalizedPageSize(), filter.offset())

	log.Printf("[REPOSITORY] List - running SQL SELECT query: %s | args=%v", query, args)
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
	log.Printf("[REPOSITORY] List - query returned %d rows from database", len(tasks))

	return tasks, nil
}

// Count returns how many rows match the filter (ignoring Page/PageSize),
// which the service layer uses to compute totalPages for the client.
func (r *mysqlTaskRepository) Count(ctx context.Context, filter TaskFilter) (int, error) {
	where, args := filter.whereClause()
	query := `SELECT COUNT(*) FROM tasks ` + where

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("repository: count tasks: %w", err)
	}
	return count, nil
}

func (r *mysqlTaskRepository) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	const query = `
		UPDATE tasks
		SET title = ?, description = ?, due_date = ?, status = ?, priority = ?, category = ?, color = ?
		WHERE id = ? AND user_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.DueDate, t.Status, t.Priority, t.Category, t.Color, t.ID, t.UserID)
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

	return r.GetByID(ctx, t.UserID, t.ID)
}

func (r *mysqlTaskRepository) Delete(ctx context.Context, userID, id uint64) error {
	const query = `DELETE FROM tasks WHERE id = ? AND user_id = ?`

	res, err := r.db.ExecContext(ctx, query, id, userID)
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

func (r *mysqlTaskRepository) SetFavorite(ctx context.Context, userID, id uint64, favorite bool) (*models.Task, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE tasks SET favorite = ? WHERE id = ? AND user_id = ?`, favorite, id, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: set favorite: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *mysqlTaskRepository) SetArchived(ctx context.Context, userID, id uint64, archived bool) (*models.Task, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE tasks SET archived = ? WHERE id = ? AND user_id = ?`, archived, id, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: set archived: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *mysqlTaskRepository) BulkDelete(ctx context.Context, userID uint64, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args := inClauseQuery(`DELETE FROM tasks WHERE user_id = ? AND id IN (%s)`, ids)
	args = append([]any{userID}, args...)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("repository: bulk delete: %w", err)
	}
	return res.RowsAffected()
}

func (r *mysqlTaskRepository) BulkSetStatus(ctx context.Context, userID uint64, ids []uint64, status models.Status) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, idArgs := inClauseQuery(`UPDATE tasks SET status = ? WHERE user_id = ? AND id IN (%s)`, ids)
	args := append([]any{status, userID}, idArgs...)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("repository: bulk set status: %w", err)
	}
	return res.RowsAffected()
}

func (r *mysqlTaskRepository) Stats(ctx context.Context, userID uint64) (*models.Stats, error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(status = 'completed'), 0) AS completed,
			COALESCE(SUM(status = 'pending'), 0) AS pending,
			COALESCE(SUM(status = 'in_progress'), 0) AS in_progress,
			COALESCE(SUM(priority = 'high' AND status <> 'completed'), 0) AS high_priority,
			COALESCE(SUM(due_date < CURDATE() AND status <> 'completed'), 0) AS overdue,
			COALESCE(SUM(due_date BETWEEN CURDATE() AND CURDATE() + INTERVAL 7 DAY AND status <> 'completed'), 0) AS upcoming_week,
			COALESCE(SUM(favorite = TRUE), 0) AS favorites,
			COALESCE(SUM(archived = TRUE), 0) AS archived
		FROM tasks
		WHERE user_id = ? AND archived = FALSE
	`
	// Archived tasks are excluded from every count except the archived
	// count itself, which is why that one is computed separately below.
	var s models.Stats
	row := r.db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(
		&s.Total, &s.Completed, &s.Pending, &s.InProgress,
		&s.HighPriority, &s.Overdue, &s.UpcomingWeek, &s.Favorites, &s.Archived,
	); err != nil {
		return nil, fmt.Errorf("repository: stats: %w", err)
	}

	// The main query filters WHERE archived = FALSE, so s.Archived comes
	// back as 0. Fetch the true archived count with a second, separate
	// COUNT so the dashboard can still surface it.
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE user_id = ? AND archived = TRUE`, userID).Scan(&s.Archived); err != nil {
		return nil, fmt.Errorf("repository: archived count: %w", err)
	}

	return &s, nil
}

func (r *mysqlTaskRepository) Categories(ctx context.Context, userID uint64) ([]string, error) {
	const query = `
		SELECT DISTINCT category FROM tasks
		WHERE user_id = ? AND category <> '' AND archived = FALSE
		ORDER BY category ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
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

// inClauseQuery builds a `... IN (?, ?, ...)` fragment safely (each id
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
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.DueDate, &t.Status,
		&t.Priority, &t.Category, &t.Color, &t.Favorite, &t.Archived,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
