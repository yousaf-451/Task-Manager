package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/granet/task-manager/internal/service"
)

// SessionCookieName is the cookie both the login/signup handlers (which set
// it) and RequireAuth (which reads it) agree on. It lives here, rather than
// in the handler package, so both sides share a single source of truth.
const SessionCookieName = "session_token"

type ctxKey int

const userIDCtxKey ctxKey = iota

// ContextWithUserID returns a copy of ctx carrying the authenticated
// user's id, for handlers to read back out with UserIDFromContext.
func ContextWithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

// UserIDFromContext returns the user id attached by RequireAuth, if any.
// ok is false for a request that never went through RequireAuth (which
// should not happen for any route that needs it - see routes.go).
func UserIDFromContext(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(userIDCtxKey).(uint64)
	return id, ok
}

// RequireAuth returns a wrapper that only calls next if the request carries
// a valid, unexpired session cookie; otherwise it responds 401 and never
// calls next. On success, the resolved user id is attached to the request
// context (see ContextWithUserID) so downstream handlers know who is
// making the call without re-deriving it themselves.
//
// This replaces the old X-User-Id header (see the User model's earlier doc
// comment in migration 0004) with a real, server-verified identity.
func RequireAuth(authService service.AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				writeUnauthorized(w)
				return
			}

			user, err := authService.CurrentUser(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, service.ErrInvalidCredentials) {
					writeUnauthorized(w)
					return
				}
				writeServerError(w)
				return
			}

			ctx := ContextWithUserID(r.Context(), user.ID)
			next(w, r.WithContext(ctx))
		}
	}
}

// writeUnauthorized and writeServerError intentionally duplicate the tiny
// envelope shape from internal/handler/response.go rather than importing
// the handler package from here, keeping middleware -> service as the only
// cross-package dependency this file needs.
type authEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func writeUnauthorized(w http.ResponseWriter) {
	writeAuthJSON(w, http.StatusUnauthorized, "authentication required")
}

func writeServerError(w http.ResponseWriter) {
	writeAuthJSON(w, http.StatusInternalServerError, "internal server error")
}

func writeAuthJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(authEnvelope{Success: false, Error: message})
}
