package middleware

import "net/http"

// Auth is a stub authentication middleware. It currently passes all requests
// through without verification. Replace with real auth logic when ready.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
