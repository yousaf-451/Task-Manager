package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is a minimal record of a logged-in browser session: which user
// it belongs to and when it stops being valid. The session token itself
// (the primary key) is generated and hex-encoded by the service layer
// (see AuthService) and doubles as the value stored in the session cookie.
type Session struct {
	Token     string
	UserID    uint64
	ExpiresAt time.Time
}

// SessionRepository persists sessions created at login and consulted on
// every authenticated request afterwards (see internal/middleware/auth.go).
type SessionRepository interface {
	Create(ctx context.Context, token string, userID uint64, expiresAt time.Time) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
}

type mysqlSessionRepository struct {
	db *sql.DB
}

// NewMySQLSessionRepository builds a SessionRepository backed by the given
// *sql.DB connection pool.
func NewMySQLSessionRepository(db *sql.DB) SessionRepository {
	return &mysqlSessionRepository{db: db}
}

// Create inserts a new session row. Called once, right after a successful
// login or signup.
func (r *mysqlSessionRepository) Create(ctx context.Context, token string, userID uint64, expiresAt time.Time) error {
	const insert = `INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`
	if _, err := r.db.ExecContext(ctx, insert, token, userID, expiresAt); err != nil {
		return fmt.Errorf("repository: create session: %w", err)
	}
	return nil
}

// GetByToken looks up a session by its token (the primary key, so this is
// a fast lookup even with a large sessions table). Returns ErrNotFound if
// no such session exists; callers are responsible for checking ExpiresAt
// themselves (an expired-but-present row is still "found" here, so the
// caller can decide whether to also clean it up).
func (r *mysqlSessionRepository) GetByToken(ctx context.Context, token string) (*Session, error) {
	const query = `SELECT id, user_id, expires_at FROM sessions WHERE id = ?`

	var s Session
	err := r.db.QueryRowContext(ctx, query, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get session: %w", err)
	}
	return &s, nil
}

// Delete removes a session (logout, or an expired session being cleaned up
// opportunistically). Deleting an already-missing token is not an error -
// the end state (no such session) is what the caller wants either way.
func (r *mysqlSessionRepository) Delete(ctx context.Context, token string) error {
	const query = `DELETE FROM sessions WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("repository: delete session: %w", err)
	}
	return nil
}
