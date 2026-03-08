package middleware

import (
	"context"
	"net/http"
	"regexp"
)

type contextKey string

const tenantIDKey contextKey = "tenant_id"

var uuidRegex = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// TenantIDFromContext returns the tenant ID stored in the request context.
func TenantIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(tenantIDKey).(string)
	return v
}

// Tenant extracts X-Tenant-ID from the request header, validates its UUID
// format, and stores it in the request context. Returns 400 if the header
// is missing or invalid.
func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip tenant check for health endpoint.
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			http.Error(w, `{"error":"missing X-Tenant-ID header"}`, http.StatusBadRequest)
			return
		}

		if !uuidRegex.MatchString(tenantID) {
			http.Error(w, `{"error":"invalid X-Tenant-ID format"}`, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
