package performance

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestAPILoadPerformance tests API endpoint performance under load
func TestAPILoadPerformance(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing time
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	concurrentRequests := 100
	requestsPerSecond := 50

	var wg sync.WaitGroup
	errors := make(chan error, concurrentRequests)
	latencies := make(chan time.Duration, concurrentRequests)

	start := time.Now()

	// Send concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Rate limiting
			time.Sleep(time.Duration(i) * time.Second / time.Duration(requestsPerSecond))

			reqStart := time.Now()
			resp, err := http.Get(server.URL)
			latency := time.Since(reqStart)

			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}

			latencies <- latency
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	close(errors)
	close(latencies)

	// Collect results
	errorCount := len(errors)
	var totalLatency time.Duration
	latencyCount := 0

	for latency := range latencies {
		totalLatency += latency
		latencyCount++
	}

	avgLatency := totalLatency / time.Duration(latencyCount)

	// Performance assertions
	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	if avgLatency > 100*time.Millisecond {
		t.Errorf("Average latency too high: %v (expected < 100ms)", avgLatency)
	}

	t.Logf("Performance Results:")
	t.Logf("  Total Requests: %d", concurrentRequests)
	t.Logf("  Successful: %d", latencyCount)
	t.Logf("  Errors: %d", errorCount)
	t.Logf("  Total Time: %v", totalTime)
	t.Logf("  Average Latency: %v", avgLatency)
	t.Logf("  Requests/Second: %.2f", float64(concurrentRequests)/totalTime.Seconds())
}

// TestAPIConcurrentRequests tests API handling of concurrent requests
func TestAPIConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	concurrentRequests := 50
	var wg sync.WaitGroup
	errors := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := http.Get(server.URL)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d: unexpected status code: %d", id, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := len(errors)
	if errorCount > 0 {
		t.Errorf("Expected 0 errors with concurrent requests, got %d", errorCount)
	}
}

// TestAPITimeoutHandling tests API timeout handling
func TestAPITimeoutHandling(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	start := time.Now()
	_, err = client.Do(req)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if duration > 1*time.Second {
		t.Errorf("Timeout took too long: %v (expected < 1s)", duration)
	}
}

