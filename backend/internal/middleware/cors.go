// Package middleware contains cross-cutting HTTP concerns (CORS, logging)
// that wrap the router without polluting handlers or business logic.
package middleware

import (
	"net/http"
)

// CORS returns middleware that allows the given list of origins to call
// the API from a browser, including preflight OPTIONS requests.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowed[origin] || allowed["*"]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "600")
				// The session cookie (see internal/middleware/auth.go) only
				// gets sent by the browser on cross-origin requests if both
				// sides opt in: the frontend's fetch() calls pass
				// `credentials: "include"`, and the server must echo this
				// header back. Note this is only safe together with an
				// explicit origin above (never "*") - browsers reject the
				// combination of a wildcard origin with credentials.
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
