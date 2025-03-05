package middleware

import "net/http"

type Layer func(handler http.Handler) http.Handler

// ContextKey 's should be unique and immutable.
type ContextKey string
