package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHeaderField verifies that HeaderField returns a HdrField with the expected key and value.
func TestHeaderField(t *testing.T) {
	key, value := "Content-Type", "application/json"
	h := HeaderField(key, value)

	if h.Header != key {
		t.Errorf("Expected header key %q, got %q", key, h.Header)
	}
	if h.Value != value {
		t.Errorf("Expected header value %q, got %q", value, h.Value)
	}
}

// TestWithResponseHeaders verifies that WithResponseHeaders sets the specified headers on the response.
func TestWithResponseHeaders(t *testing.T) {
	// Create a couple of header fields.
	hf1 := HeaderField("X-Test-Header", "Foo")
	hf2 := HeaderField("Content-Type", "text/plain")

	// Create the middleware using WithResponseHeaders.
	middleware := WithResponseHeaders(hf1, hf2)

	// Prepare a final handler that writes a simple response.
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap the final handler with the middleware.
	handler := middleware(finalHandler)

	// Create a test HTTP request and ResponseRecorder.
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Get the resulting response.
	resp := rec.Result()

	// Read the body to complete the response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	if !strings.Contains(string(body), "OK") {
		t.Errorf("Expected response body to contain %q, got %q", "OK", string(body))
	}

	// Verify headers were set correctly.
	if got := resp.Header.Get("X-Test-Header"); got != "Foo" {
		t.Errorf("Expected X-Test-Header to be %q, got %q", "Foo", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/plain" {
		t.Errorf("Expected Content-Type to be %q, got %q", "text/plain", got)
	}
}
