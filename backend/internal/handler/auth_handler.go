package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/granet/task-manager/internal/middleware"
	"github.com/granet/task-manager/internal/models"
	"github.com/granet/task-manager/internal/service"
)

// AuthHandler wires HTTP requests to the AuthService: signup, login,
// logout, "who am I", and account deletion. It is the only place that
// touches the session cookie directly (setting it on login/signup,
// clearing it on logout/delete) - internal/middleware/auth.go is the only
// other place that reads it.
type AuthHandler struct {
	service service.AuthService
	cookieSecure bool
}

// NewAuthHandler builds an AuthHandler backed by the given service.
// cookieSecure should be true in any deployment served over HTTPS (see
// config.Config.CookieSecure).
func NewAuthHandler(s service.AuthService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{service: s, cookieSecure: cookieSecure}
}

// Signup handles POST /api/auth/signup. Body: {"name","email","password"}.
// On success it also logs the new user in immediately (sets the session
// cookie), so signup feels like one step instead of "create an account,
// then separately log in with it".
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if _, err := h.service.Signup(r.Context(), req); err != nil {
		handleServiceError(w, err)
		return
	}

	// Signup succeeded; log them in with the same credentials to obtain a
	// session (rather than duplicating session-creation logic here).
	user, token, expiresAt, err := h.service.Login(r.Context(), models.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.setSessionCookie(w, token, expiresAt)
	writeSuccess(w, http.StatusCreated, toUserResponse(user))
}

// Login handles POST /api/auth/login. Body: {"email","password"}.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, token, expiresAt, err := h.service.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	h.setSessionCookie(w, token, expiresAt)
	writeSuccess(w, http.StatusOK, toUserResponse(user))
}

// Logout handles POST /api/auth/logout. Requires a valid session (see
// middleware.RequireAuth in routes.go), which it then invalidates.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			handleServiceError(w, err)
			return
		}
	}
	h.clearSessionCookie(w)
	writeSuccess(w, http.StatusOK, map[string]bool{"loggedOut": true})
}

// Me handles GET /api/auth/me - returns the currently signed-in user.
// Requires a valid session (see middleware.RequireAuth in routes.go).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.service.CurrentUser(r.Context(), cookie.Value)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if user.ID != userID {
		// Defensive only - RequireAuth already resolved this same cookie
		// to this same user id, so this branch should be unreachable.
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	writeSuccess(w, http.StatusOK, toUserResponse(user))
}

// DeleteAccount handles DELETE /api/auth/me - permanently deletes the
// signed-in user's own account (and, via ON DELETE CASCADE, all of their
// tasks and sessions). Requires a valid session.
func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := h.service.DeleteAccount(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	h.clearSessionCookie(w)
	writeSuccess(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		// Lax is enough here: SameSite is scoped to the registrable domain
		// ("site"), not scheme+host+port, so a request from
		// localhost:5173 to localhost:8080 (or the equivalent same-domain
		// setup in production) is still same-site and the cookie is sent.
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
