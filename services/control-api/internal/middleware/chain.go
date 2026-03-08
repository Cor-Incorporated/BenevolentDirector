package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares so they execute in the order provided.
// Chain(A, B, C)(handler) executes as A -> B -> C -> handler.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
