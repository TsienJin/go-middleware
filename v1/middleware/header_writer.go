package middleware

import "net/http"

type HdrField struct {
	Header string
	Value  string
}

func HeaderField(k string, v string) *HdrField {
	return &HdrField{
		Header: k,
		Value:  v,
	}
}

func WithResponseHeaders(headers ...*HdrField) Layer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, h := range headers {
				w.Header().Set(h.Header, h.Value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
