// Package docs embeds the OpenAPI specification and exposes two small,
// dependency-free HTTP handlers so API consumers can browse interactive
// documentation without installing anything.
package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

// swaggerUIPage is a minimal HTML shell that loads Swagger UI from a CDN
// and points it at our embedded spec. No build step, no npm dependency.
const swaggerUIPage = `<!doctype html>
<html>
<head>
  <title>Task Manager API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`

// SpecHandler serves the raw OpenAPI YAML at GET /openapi.yaml.
func SpecHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(openAPISpec)
}

// UIHandler serves a Swagger UI page at GET /docs for interactive browsing.
func UIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIPage))
}
