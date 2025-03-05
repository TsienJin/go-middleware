package middleware

import "net/http"

// Apply is the middleware applicator function, which applies designated layers to the http.Handler
func Apply(target http.Handler, layers ...Layer) http.Handler {
	for _, l := range layers {
		target = l(target)
	}
	return target
}
