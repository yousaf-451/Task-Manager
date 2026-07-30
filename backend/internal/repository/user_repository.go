package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/granet/task-manager/internal/models"
)

// ErrDuplicateEmail is returned by Create when the given email already
// belongs to another user (users.email has a UNIQUE constraint - see
// schema.sql). The service layer maps this to HTTP 409 Conflict.
var ErrDuplicateEmail = errors.New("a user with this email already exists")

// mysqlDuplicateEntryErrNum is the MySQL error number for a UNIQUE
// constraint violation (ER_DUP_ENTRY).
const mysqlDuplicateEntryErrNum = 1062

// UserRepository exposes the set of user operations authentication needs:
// looking a user up by id or email, creating a new account, and deleting
// one (which cascades to that user's tasks and sessions - see
// migrations/up/0004 and 0005).
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, name, email, passwordHash string) (*models.User, error)
	Delete(ctx context.Context, id uint64) error
}

type mysqlUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository builds a UserRepository backed by the given
// *sql.DB connection pool.
func NewMySQLUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

const userColumns = `id, name, email, password_hash, created_at`

func (r *mysqlUserRepository) scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: scan user row: %w", err)
	}
	return &u, nil
}

// GetByID looks up a user by primary key. Used to resolve "who am I" once a
// session token has already been validated (see AuthService.CurrentUser).
func (r *mysqlUserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanUser(row)
}

// GetByEmail looks up a user by email (case-insensitive - callers are
// expected to lowercase first, same as SignupRequest/LoginRequest.Validate
// do). Used by login to find the account to check the password against,
// and by signup to give a clean "email already in use" error before even
// attempting the insert.
func (r *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE email = ?`
	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanUser(row)
}

// Create inserts a new user (with an already-hashed password - the
// repository layer never sees or handles plaintext passwords) and returns
// the persisted row. Returns ErrDuplicateEmail if another user already has
// this email (users.email is UNIQUE).
func (r *mysqlUserRepository) Create(ctx context.Context, name, email, passwordHash string) (*models.User, error) {
	const insert = `INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`

	result, err := r.db.ExecContext(ctx, insert, name, email, passwordHash)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntryErrNum {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("repository: create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: create user: get inserted id: %w", err)
	}

	return r.GetByID(ctx, uint64(id))
}

// Delete removes a user's account entirely. Their tasks and sessions are
// removed automatically by the database via ON DELETE CASCADE (see
// migrations/up/0004_add_users_and_ownership.sql and
// 0005_add_auth.sql) - this is exactly the referential-integrity guarantee
// those migrations were added for.
func (r *mysqlUserRepository) Delete(ctx context.Context, id uint64) error {
	const query = `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete user: rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
