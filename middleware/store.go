package middleware

import (
	"context"
	"net/http"
)

// WithStore creates and returns a middleware layer that attaches a value of generic type T
// to the request context under the specified key. This allows subsequent handlers in the chain
// to retrieve the stored value from the context.
//
// Parameters:
//
//	key   - The context key under which the value will be stored (of type ContextKey).
//	value - The value of generic type T to be stored.
//
// Returns:
//
//	A middleware Layer, which is a function that wraps an http.Handler.
//	When executing, the middleware injects the key/value pair into the request context
//	before calling the next handler.
func WithStore[T any](key ContextKey, value T) Layer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetStoreFromContext attempts to retrieve a value of generic type T from the provided context
// using the specified key. It performs a type assertion to convert the stored value to type T.
//
// Parameters:
//
//	ctx - The context from which the value will be retrieved.
//	key - The context key corresponding to the value (of type ContextKey).
//
// Returns:
//
//	The retrieved value of type T and a boolean indicating whether the value was found
//	and successfully asserted to type T. If the key is not present or the value cannot
//	be converted to type T, the function returns the zero value of T and false.
func GetStoreFromContext[T any](ctx context.Context, key ContextKey) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}
