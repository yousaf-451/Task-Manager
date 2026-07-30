package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// User is an account record. Authentication is real (see
// internal/service/auth_service.go): PasswordHash is a bcrypt hash, never
// the plaintext password, and is never serialized to JSON (see the `json:"-"`
// tag below) so it can never leak out through an API response.
//
// "Who is the current user" for a request is resolved from a session
// cookie (see internal/middleware/auth.go), which is set on successful
// login/signup and validated against the `sessions` table on every
// subsequent request.
type User struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

const (
	userNameMaxLength     = 100
	userEmailMaxLength    = 150
	userPasswordMinLength = 8
	userPasswordMaxLength = 72 // bcrypt silently ignores anything past 72 bytes
)

// emailPattern is an intentionally loose sanity check (not a full RFC 5322
// validator) - it exists to reject obviously malformed input like "abc",
// not to police every edge case of what counts as a valid email address.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// SignupRequest is the JSON body accepted by POST /api/auth/signup.
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate trims and checks the request, returning the cleaned
// name/email pair plus the (still plaintext, not-yet-hashed) password the
// service layer should hash and persist.
func (r SignupRequest) Validate() (name string, email string, password string, err error) {
	name = strings.TrimSpace(r.Name)
	email = strings.TrimSpace(strings.ToLower(r.Email))
	password = r.Password

	if name == "" {
		return "", "", "", errors.New("name is required")
	}
	if len(name) > userNameMaxLength {
		return "", "", "", errors.New("name must be 100 characters or fewer")
	}
	if email == "" {
		return "", "", "", errors.New("email is required")
	}
	if len(email) > userEmailMaxLength {
		return "", "", "", errors.New("email must be 150 characters or fewer")
	}
	if !emailPattern.MatchString(email) {
		return "", "", "", errors.New("email must be a valid email address")
	}
	if len(password) < userPasswordMinLength {
		return "", "", "", errors.New("password must be at least 8 characters")
	}
	if len(password) > userPasswordMaxLength {
		return "", "", "", errors.New("password must be 72 characters or fewer")
	}

	return name, email, password, nil
}

// LoginRequest is the JSON body accepted by POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate does minimal shape-checking only; the real check (does this
// password match this email) happens in the service layer against the
// stored bcrypt hash, not here.
func (r LoginRequest) Validate() (email string, password string, err error) {
	email = strings.TrimSpace(strings.ToLower(r.Email))
	password = r.Password

	if email == "" {
		return "", "", errors.New("email is required")
	}
	if password == "" {
		return "", "", errors.New("password is required")
	}

	return email, password, nil
}
