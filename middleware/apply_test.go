package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testMiddleware is a helper that returns a middleware layer which writes a label
// before calling the next handler. This lets us verify the order of middleware execution.
func testMiddleware(label string) Layer {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write the label indicating this middleware was executed.
			// In a real scenario, you might modify the request or response rather than raw writes.
			_, _ = w.Write([]byte(label))
			// Call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}

// TestApply_NoLayers verifies that when no layers are provided,
// the original handler is executed with no modifications.
func TestApply_NoLayers(t *testing.T) {
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a simple response.
		_, _ = w.Write([]byte("final"))
	})
	// Apply no middleware layers.
	handler := Apply(finalHandler)

	// Create a test request and recorder.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	// Verify the final response.
	result := rec.Body.String()
	if result != "final" {
		t.Errorf("Expected response %q, got %q", "final", result)
	}
}

// TestApply_MultipleLayers verifies that multiple layers are applied in sequence.
// In this example, the first middleware writes "L1" and the second middleware writes "L2".
// Because of the way Apply works, the middleware are applied in order,
// so the outer (last applied) middleware executes first.
func TestApply_MultipleLayers(t *testing.T) {
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("final"))
	})

	// Create two test layers.
	layer1 := testMiddleware("L1")
	layer2 := testMiddleware("L2")

	// Apply layers; note the order. The first layer wraps finalHandler,
	// and the second wraps the result of layer1.
	handler := Apply(finalHandler, layer1, layer2)

	// Create a test request and recorder.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	// Expected order:
	// The outer (layer2) writes "L2", then inner layer (layer1) writes "L1", and then finalHandler writes "final".
	expectedOutput := "L2L1final"
	result := strings.TrimSpace(rec.Body.String())
	if result != expectedOutput {
		t.Errorf("Expected response %q, got %q", expectedOutput, result)
	}
}
