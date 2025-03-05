package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestWithTimeout_Success verifies that the next handler’s response is passed through
// when the request finishes before the timeout.
func TestWithTimeout_Success(t *testing.T) {
	handlerOutput := "OK"

	// finalHandler that writes "OK" immediately.
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(handlerOutput))
	})
	// Use a relatively long timeout so that the finalHandler finishes in time.
	timeoutMiddleware := WithTimeout(100 * time.Millisecond)
	handler := timeoutMiddleware(finalHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading body: %v", err)
	}
	if string(body) != handlerOutput {
		t.Errorf("Expected body %q, got %q", handlerOutput, string(body))
	}
}

// TestWithTimeout_Timeout forces the timeout path. The finalHandler blocks indefinitely,
// and we use a very short timeout to trigger the context deadline.
func TestWithTimeout_Timeout(t *testing.T) {
	// finalHandler that blocks indefinitely.
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block forever.
		select {}
	})
	// Use a very short timeout so that the context expires immediately.
	timeoutMiddleware := WithTimeout(1 * time.Nanosecond)
	handler := timeoutMiddleware(finalHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusGatewayTimeout, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading body: %v", err)
	}
	if !strings.Contains(string(body), "Request timed out") {
		t.Errorf("Expected body to contain %q, got %q", "Request timed out", string(body))
	}
}

func TestWithTimeout_TimeoutRace(t *testing.T) {
	// finalHandler that blocks indefinitely.
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	// Use a very short timeout so that the context expires immediately.
	timeoutMiddleware := WithTimeout(1 * time.Nanosecond)
	handler := timeoutMiddleware(finalHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusGatewayTimeout, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading body: %v", err)
	}
	if !strings.Contains(string(body), "Request timed out") {
		t.Errorf("Expected body to contain %q, got %q", "Request timed out", string(body))
	}
}
