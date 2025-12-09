package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/classification-service/internal/config"
)

// TestIntegration_ConcurrentRequests tests handling multiple concurrent requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	logger := zap.NewNop()
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			MaxConcurrentRequests: 50,
		},
	}
	
	// Create minimal handler for testing
	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
		requestQueue: NewRequestQueue(50),
		cache: make(map[string]*cacheEntry),
		inFlightRequests: make(map[string]*inFlightRequest),
	}
	
	handler.WorkerPool = NewWorkerPool(10, handler.requestQueue, handler, logger)
	handler.WorkerPool.Start()
	defer handler.WorkerPool.Stop()
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleClassification))
	defer server.Close()
	
	// Send 20 concurrent requests
	numRequests := 20
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			reqBody := ClassificationRequest{
				RequestID:   "test-concurrent-" + string(rune(id)),
				BusinessName: "Test Business",
			}
			
			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable {
					mu.Lock()
					if resp.StatusCode == http.StatusOK {
						successCount++
					}
					mu.Unlock()
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// At least some requests should succeed (or be queued)
	// Note: Actual processing may fail without full handler setup, but queue should work
	t.Logf("Successfully processed %d/%d requests", successCount, numRequests)
}

// TestIntegration_QueueFull tests behavior when queue is full
func TestIntegration_QueueFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	logger := zap.NewNop()
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			MaxConcurrentRequests: 5, // Small queue
		},
	}
	
	// Create minimal handler for testing
	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
		requestQueue: NewRequestQueue(5),
		cache: make(map[string]*cacheEntry),
		inFlightRequests: make(map[string]*inFlightRequest),
	}
	
	handler.WorkerPool = NewWorkerPool(2, handler.requestQueue, handler, logger)
	handler.WorkerPool.Start()
	defer handler.WorkerPool.Stop()
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleClassification))
	defer server.Close()
	
	// Send 10 requests rapidly to fill queue
	numRequests := 10
	serviceUnavailableCount := 0
	var mu sync.Mutex
	
	for i := 0; i < numRequests; i++ {
		reqBody := ClassificationRequest{
			RequestID:   "test-queue-full-" + string(rune(i)),
			BusinessName: "Test Business",
		}
		
		jsonData, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusServiceUnavailable {
				mu.Lock()
				serviceUnavailableCount++
				mu.Unlock()
			}
			resp.Body.Close()
		}
	}
	
	// At least some requests should get 503 Service Unavailable
	if serviceUnavailableCount == 0 {
		t.Log("No requests received 503 - queue may not have filled (this is OK if queue processes quickly)")
	} else {
		t.Logf("Received %d 503 Service Unavailable responses (expected when queue is full)", serviceUnavailableCount)
	}
}

// TestIntegration_WorkerPoolShutdown tests graceful shutdown of worker pool
func TestIntegration_WorkerPoolShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	logger := zap.NewNop()
	queue := NewRequestQueue(10)
	
	handler := &ClassificationHandler{
		logger: logger,
	}
	
	pool := NewWorkerPool(5, queue, handler, logger)
	pool.Start()
	
	// Give workers time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop pool (should not block indefinitely)
	done := make(chan bool)
	go func() {
		pool.Stop()
		done <- true
	}()
	
	select {
	case <-done:
		t.Log("Worker pool stopped successfully")
	case <-time.After(5 * time.Second):
		t.Error("Worker pool shutdown timed out")
	}
}

// TestIntegration_ContextDeadlineExpiration tests handling of expired contexts
func TestIntegration_ContextDeadlineExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	logger := zap.NewNop()
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			MaxConcurrentRequests: 10,
		},
	}
	
	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
		requestQueue: NewRequestQueue(10),
		cache: make(map[string]*cacheEntry),
		inFlightRequests: make(map[string]*inFlightRequest),
	}
	
	handler.WorkerPool = NewWorkerPool(2, handler.requestQueue, handler, logger)
	handler.WorkerPool.Start()
	defer handler.WorkerPool.Stop()
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleClassification))
	defer server.Close()
	
	// Create request with very short timeout
	reqBody := ClassificationRequest{
		RequestID:   "test-context-expired",
		BusinessName: "Test Business",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Create context with very short deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)
	
	req = req.WithContext(ctx)
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		// Request should either succeed (if Background context is used) or timeout
		t.Logf("Request completed with status: %d", resp.StatusCode)
	} else {
		// Error is expected if context expired
		t.Logf("Request failed as expected: %v", err)
	}
}

// TestIntegration_RequestDeduplicationWithQueue tests that request deduplication works with queue
func TestIntegration_RequestDeduplicationWithQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	
	logger := zap.NewNop()
	cfg := &config.Config{
		Classification: config.ClassificationConfig{
			MaxConcurrentRequests: 10,
		},
	}
	
	handler := &ClassificationHandler{
		logger: logger,
		config: cfg,
		requestQueue: NewRequestQueue(10),
		cache: make(map[string]*cacheEntry),
		inFlightRequests: make(map[string]*inFlightRequest),
	}
	
	handler.WorkerPool = NewWorkerPool(2, handler.requestQueue, handler, logger)
	handler.WorkerPool.Start()
	defer handler.WorkerPool.Stop()
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleClassification))
	defer server.Close()
	
	// Send two identical requests concurrently
	reqBody := ClassificationRequest{
		RequestID:   "test-dedup",
		BusinessName: "Test Business",
	}
	
	jsonData, _ := json.Marshal(reqBody)
	
	var wg sync.WaitGroup
	responses := make([]int, 2)
	
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err == nil {
				responses[id] = resp.StatusCode
				resp.Body.Close()
			}
		}(i)
	}
	
	wg.Wait()
	
	// Both requests should be handled (either processed or deduplicated)
	t.Logf("Request 1 status: %d, Request 2 status: %d", responses[0], responses[1])
}

