package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestValidationMiddleware_Basic(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	config := &ValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB
	}
	validator := NewValidator(config, logger)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test valid JSON request
	req, _ := http.NewRequest("POST", "/v1/risk/assess", bytes.NewBufferString(`{"business_id": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	validator.Middleware(handler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestValidationMiddleware_InvalidJSON(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	config := &ValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB
		Enabled:     true,
	}
	validator := NewValidator(config, logger)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test invalid JSON request with a path that triggers validation
	req, _ := http.NewRequest("POST", "/v1/risk/assess", bytes.NewBufferString(`{"business_id": "test"`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	validator.Middleware(handler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestValidationMiddleware_GETRequest(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	config := &ValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB
	}
	validator := NewValidator(config, logger)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test GET request (should not validate JSON)
	req, _ := http.NewRequest("GET", "/v1/risk/categories", nil)
	rr := httptest.NewRecorder()

	validator.Middleware(handler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestValidationMiddleware_Performance(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})

	config := &ValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB
	}
	validator := NewValidator(config, logger)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test performance with multiple concurrent requests
	concurrency := 10
	requestsPerGoroutine := 100

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				req, _ := http.NewRequest("POST", "/v1/risk/assess", bytes.NewBufferString(`{"business_id": "test"}`))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				validator.Middleware(handler).ServeHTTP(rr, req)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	totalRequests := concurrency * requestsPerGoroutine

	t.Logf("Processed %d requests in %v (%.2f requests/second)",
		totalRequests, duration, float64(totalRequests)/duration.Seconds())
}
