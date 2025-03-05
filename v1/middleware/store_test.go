package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWithStore_Success verifies that WithStore injects the expected value into the context
// and that GetStoreFromContext retrieves it correctly.
func TestWithStore_Success(t *testing.T) {

	key := ContextKey("testKey")
	expectedVal := 2359

	// Create middleware for storing val
	middleware := WithStore(key, expectedVal)

	// Create dummy endpoint handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, ok := GetStoreFromContext[int](r.Context(), key)
		if !ok {
			t.Errorf("Expected value for %s, but none was found", key)
			return
		}

		if val != expectedVal {
			t.Errorf("Expected %d, but received %d", expectedVal, val)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// create actual handler wrapped with middleware
	handler := middleware(dummyHandler)

	// Create test HTTP request and recorder
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Execute the handler
	handler.ServeHTTP(rec, req)

	if status := rec.Result().StatusCode; status != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, status)
		return
	}
}

// TestWithStore_MissingKey checks that GetStoreFromContext returns the zero value and false when a key
// is not present in the context.
func TestWithStore_MissingKey(t *testing.T) {
	key := ContextKey("missingKey")
	ctx := context.Background()

	val, ok := GetStoreFromContext[string](ctx, key)
	if ok {
		t.Errorf("Did not expect to find a value for key %s", key)
	}
	if val != "" {
		t.Errorf("Expected nothing for string, but got %s", val)
	}
}

// TestGetStore_TypeMismatch ensures that GetStoreFromContext returns false when the stored value's type
// does not match the expected type.
func TestGetStore_TypeMismatch(t *testing.T) {
	key := ContextKey("mismatchKey")
	// Store a string value.
	ctx := context.WithValue(context.Background(), key, "hello")

	_, ok := GetStoreFromContext[int](ctx, key)
	if ok {
		t.Errorf("Expected type assertion to fail for key %q when retrieving as int", key)
	}
}
