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
)

// NewRouter builds the fully configured HTTP handler for the application,
// including middleware (logging, CORS) and all API routes.
//
// It uses Go 1.22's enhanced http.ServeMux, which supports method-specific
// patterns (e.g. "GET /api/tasks") and path parameters (e.g. "{id}") natively,
// so no external routing library is required.
func NewRouter(taskHandler *handler.TaskHandler, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	// Health check - useful for load balancers / container orchestrators.
	mux.HandleFunc("GET /health", taskHandler.HealthCheck)

	// API documentation - interactive Swagger UI + raw OpenAPI spec.
	mux.HandleFunc("GET /docs", docs.UIHandler)
	mux.HandleFunc("GET /openapi.yaml", docs.SpecHandler)

	// Task CRUD endpoints.
	mux.HandleFunc("POST /api/tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /api/tasks", taskHandler.ListTasks)
	mux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.DeleteTask)

	// Dashboard / productivity endpoints. Registered before the bulk routes
	// below so "stats" is never mistaken for a path parameter by the mux.
	mux.HandleFunc("GET /api/tasks/stats", taskHandler.Stats)
	mux.HandleFunc("GET /api/categories", taskHandler.Categories)

	mux.HandleFunc("PATCH /api/tasks/{id}/favorite", taskHandler.ToggleFavorite)
	mux.HandleFunc("PATCH /api/tasks/{id}/archive", taskHandler.ToggleArchive)
	mux.HandleFunc("POST /api/tasks/{id}/duplicate", taskHandler.DuplicateTask)

	mux.HandleFunc("POST /api/tasks/bulk/delete", taskHandler.BulkDelete)
	mux.HandleFunc("POST /api/tasks/bulk/complete", taskHandler.BulkComplete)

	// Apply middleware, outermost first: every request is logged, then
	// CORS headers are applied before reaching the mux.
	var h http.Handler = mux
	h = middleware.CORS(allowedOrigins)(h)
	h = middleware.Logger(h)

	return h
}
