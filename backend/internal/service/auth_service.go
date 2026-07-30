package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/granet/task-manager/internal/models"
	"github.com/granet/task-manager/internal/repository"
)

// ErrInvalidCredentials is returned by Login when the email doesn't match
// any account, or the password doesn't match the stored hash. The two
// cases are deliberately not distinguished in the returned error (and the
// handler returns the same generic message either way) so a caller can't
// use the API to enumerate which emails have accounts.
var ErrInvalidCredentials = errors.New("invalid email or password")

// SessionDuration is how long a session stays valid after login/signup
// before the user has to log in again.
const SessionDuration = 7 * 24 * time.Hour

// sessionTokenBytes is the amount of randomness (in bytes) behind each
// session token; hex-encoded this becomes a 64-character string, matching
// the `sessions.id CHAR(64)` column (see migrations/up/0005_add_auth.sql).
const sessionTokenBytes = 32

// bcryptCost is deliberately the library default rather than a custom
// value - it is already tuned to be slow enough to resist brute-forcing
// while staying fast enough for interactive login.
const bcryptCost = bcrypt.DefaultCost

// AuthService covers the full account lifecycle: creating an account,
// logging in and out, finding out who the current session belongs to, and
// deleting an account.
type AuthService interface {
	Signup(ctx context.Context, req models.SignupRequest) (*models.User, error)
	Login(ctx context.Context, req models.LoginRequest) (user *models.User, token string, expiresAt time.Time, err error)
	Logout(ctx context.Context, token string) error
	CurrentUser(ctx context.Context, token string) (*models.User, error)
	DeleteAccount(ctx context.Context, userID uint64) error
}

type authService struct {
	users    repository.UserRepository
	sessions repository.SessionRepository
}

// NewAuthService builds an AuthService backed by the given repositories.
func NewAuthService(users repository.UserRepository, sessions repository.SessionRepository) AuthService {
	return &authService{users: users, sessions: sessions}
}

// Signup validates the request, hashes the password (the plaintext
// password is never stored or logged), and creates the account. It does
// not log the new user in - callers that want that should follow up with
// Login using the same credentials (or the handler layer can do so
// transparently for a smoother signup flow).
func (s *authService) Signup(ctx context.Context, req models.SignupRequest) (*models.User, error) {
	name, email, password, err := req.Validate()
	if err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}

	user, err := s.users.Create(ctx, name, email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, errors.Join(ErrConflict, err)
		}
		return nil, err
	}
	return user, nil
}

// Login verifies the given credentials and, if they check out, creates a
// new session and returns its token and expiry so the handler can set the
// session cookie.
func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.User, string, time.Time, error) {
	email, password, err := req.Validate()
	if err != nil {
		return nil, "", time.Time{}, errors.Join(ErrValidation, err)
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Same error as a wrong password - don't reveal whether the
			// email exists.
			return nil, "", time.Time{}, ErrInvalidCredentials
		}
		return nil, "", time.Time{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	token, err := newSessionToken()
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("service: generate session token: %w", err)
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := s.sessions.Create(ctx, token, user.ID, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expiresAt, nil
}

// Logout deletes the session identified by token. A missing/already-gone
// token is treated as a successful logout - the end state the caller wants
// (no valid session) is already true.
func (s *authService) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}

// CurrentUser resolves a session token to the user it belongs to, or
// ErrInvalidCredentials if the token is missing, unknown, or expired. This
// is what internal/middleware/auth.go calls on every authenticated
// request.
func (s *authService) CurrentUser(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, ErrInvalidCredentials
	}

	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		// Best-effort cleanup; an error here shouldn't turn into a 500 for
		// what is otherwise a normal "please log in again" response.
		_ = s.sessions.Delete(ctx, token)
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return user, nil
}

// DeleteAccount permanently removes the given user's account. Their tasks
// and sessions are removed automatically by the database (ON DELETE
// CASCADE - see migrations 0004 and 0005), so there is nothing else to
// clean up here.
func (s *authService) DeleteAccount(ctx context.Context, userID uint64) error {
	if err := s.users.Delete(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return errors.Join(ErrNotFound, err)
		}
		return err
	}
	return nil
}

// newSessionToken generates a cryptographically random, hex-encoded
// session token.
func newSessionToken() (string, error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
