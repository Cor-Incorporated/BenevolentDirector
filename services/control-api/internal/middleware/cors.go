package middleware

import (
	"net/http"
	"os"
	"strings"
)

const (
	corsEnvKey        = "CORS_ALLOWED_ORIGINS"
	corsDefaultOrigin = "http://localhost:5173"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// Allowed origins are read from the CORS_ALLOWED_ORIGINS env var
// (comma-separated). If unset, only http://localhost:5173 is allowed.
// The /health path is excluded from CORS processing.
func CORS() Middleware {
	origins := parseCORSOrigins()
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Data-Classification")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// parseCORSOrigins reads and splits the CORS_ALLOWED_ORIGINS env var.
func parseCORSOrigins() []string {
	raw := os.Getenv(corsEnvKey)
	if raw == "" {
		return []string{corsDefaultOrigin}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{corsDefaultOrigin}
	}
	return origins
}
