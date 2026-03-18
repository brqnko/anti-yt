package v1

import (
	"encoding/json"
	"net/http"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Documentation</title>
</head>
<body>
  <div id="redoc-container"></div>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  <script>
    Redoc.init("/api/v1/swagger/openapi.json", {}, document.getElementById("redoc-container"));
  </script>
</body>
</html>`

// SwaggerMiddleware returns a middleware that serves Swagger UI at /api/v1/swagger/
// and the OpenAPI spec at /api/v1/swagger/openapi.json.
func SwaggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v1/swagger", "/api/v1/swagger/":
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(swaggerUIHTML))
			case "/api/v1/swagger/openapi.json":
				swagger, err := GetSwagger()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				swagger.Servers = nil
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(swagger)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
