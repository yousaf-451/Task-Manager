// Package routes wires HTTP paths to handlers and applies middleware. It is
// the outermost layer of the request flow:
//
//	React -> [routes] -> handler -> service -> repository -> MySQL
package routes

import (
	"net/http"

	"github.com/granet/task-manager/internal/docs"
	"github.com/granet/task-manager/internal/handler"
	"github.com/granet/task-manager/internal/middleware"
	"github.com/granet/task-manager/internal/service"
)

// NewRouter builds the fully configured HTTP handler for the application,
// including middleware (logging, CORS, auth) and all API routes.
//
// It uses Go 1.22's enhanced http.ServeMux, which supports method-specific
// patterns (e.g. "GET /api/tasks") and path parameters (e.g. "{id}") natively,
// so no external routing library is required.
func NewRouter(
	taskHandler *handler.TaskHandler,
	authHandler *handler.AuthHandler,
	authService service.AuthService,
	allowedOrigins []string,
) http.Handler {
	mux := http.NewServeMux()

	// requireAuth wraps a handler so it only runs for requests carrying a
	// valid, unexpired session cookie (see middleware/auth.go); anything
	// touching a specific user's data goes through this.
	requireAuth := middleware.RequireAuth(authService)

	// Health check - useful for load balancers / container orchestrators.
	mux.HandleFunc("GET /health", taskHandler.HealthCheck)

	// API documentation - interactive Swagger UI + raw OpenAPI spec.
	mux.HandleFunc("GET /docs", docs.UIHandler)
	mux.HandleFunc("GET /openapi.yaml", docs.SpecHandler)

	// Auth - signup/login are public (that's the whole point); logout, "who
	// am I", and account deletion require an existing session.
	mux.HandleFunc("POST /api/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", requireAuth(authHandler.Logout))
	mux.HandleFunc("GET /api/auth/me", requireAuth(authHandler.Me))
	mux.HandleFunc("DELETE /api/auth/me", requireAuth(authHandler.DeleteAccount))

	// Task CRUD endpoints - every one is scoped to the signed-in user (see
	// handler.currentUserID, resolved from the context RequireAuth sets).
	mux.HandleFunc("POST /api/tasks", requireAuth(taskHandler.CreateTask))
	mux.HandleFunc("GET /api/tasks", requireAuth(taskHandler.ListTasks))
	mux.HandleFunc("GET /api/tasks/{id}", requireAuth(taskHandler.GetTask))
	mux.HandleFunc("PUT /api/tasks/{id}", requireAuth(taskHandler.UpdateTask))
	mux.HandleFunc("DELETE /api/tasks/{id}", requireAuth(taskHandler.DeleteTask))

	// Dashboard / productivity endpoints. Registered before the bulk routes
	// below so "stats" is never mistaken for a path parameter by the mux.
	mux.HandleFunc("GET /api/tasks/stats", requireAuth(taskHandler.Stats))
	mux.HandleFunc("GET /api/categories", requireAuth(taskHandler.Categories))

	mux.HandleFunc("PATCH /api/tasks/{id}/favorite", requireAuth(taskHandler.ToggleFavorite))
	mux.HandleFunc("PATCH /api/tasks/{id}/archive", requireAuth(taskHandler.ToggleArchive))
	mux.HandleFunc("POST /api/tasks/{id}/duplicate", requireAuth(taskHandler.DuplicateTask))

	mux.HandleFunc("POST /api/tasks/bulk/delete", requireAuth(taskHandler.BulkDelete))
	mux.HandleFunc("POST /api/tasks/bulk/complete", requireAuth(taskHandler.BulkComplete))

	// Apply middleware, outermost first: every request is logged, then
	// CORS headers are applied before reaching the mux.
	var h http.Handler = mux
	h = middleware.CORS(allowedOrigins)(h)
	h = middleware.Logger(h)

	return h
}
