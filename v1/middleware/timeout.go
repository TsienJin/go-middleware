package middleware

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"
)

// timeoutResponseWriter buffers the output of the underlying handler.
// It also protects against concurrent writes.
type timeoutResponseWriter struct {
	mu       sync.Mutex
	w        http.ResponseWriter
	buf      bytes.Buffer
	wroteHdr bool
	status   int
	timedOut bool
}

// Header forwards the header access to the underlying ResponseWriter.
func (trw *timeoutResponseWriter) Header() http.Header {
	return trw.w.Header()
}

// WriteHeader records the status code.
func (trw *timeoutResponseWriter) WriteHeader(statusCode int) {
	trw.mu.Lock()
	defer trw.mu.Unlock()
	if trw.wroteHdr {
		return
	}
	trw.wroteHdr = true
	trw.status = statusCode
}

// Write buffers the output. If the writer has been marked as timed out, further writes are dropped.
func (trw *timeoutResponseWriter) Write(b []byte) (int, error) {
	trw.mu.Lock()
	defer trw.mu.Unlock()
	if trw.timedOut {
		// Discard writes that come in after timeout.
		return 0, nil
	}
	if !trw.wroteHdr {
		trw.wroteHdr = true
		trw.status = http.StatusOK
	}
	return trw.buf.Write(b)
}

// Flush writes the buffered status code and bytes to the underlying ResponseWriter.
func (trw *timeoutResponseWriter) Flush() {
	trw.mu.Lock()
	defer trw.mu.Unlock()
	// If a timeout was signaled, avoid writing.
	if trw.timedOut {
		return
	}
	trw.w.WriteHeader(trw.status)
	_, err := trw.w.Write(trw.buf.Bytes())
	if err != nil {
		trw.w.WriteHeader(http.StatusInternalServerError)
	}
}

// WithTimeout returns a middleware layer that times out after a given duration.
// Instead of writing directly to w, the next handler writes to a buffered writer.
// If the timeout expires before the handler completes, the timeout response is written and
// further writes from next are ignored.
func WithTimeout(duration time.Duration) Layer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			done := make(chan struct{})
			trw := &timeoutResponseWriter{w: w}

			// Call the next handler in a separate goroutine writing to our buffered writer.
			go func() {
				next.ServeHTTP(trw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// If the handler finishes first, flush buffered output to the real ResponseWriter.
				trw.Flush()
			case <-ctx.Done():
				// Signal timeout: mark the writer so that any later writes are ignored.
				trw.mu.Lock()
				trw.timedOut = true
				trw.mu.Unlock()
				// Write the timeout error.
				http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			}
		})
	}
}
