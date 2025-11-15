package performance

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockDataFetcher simulates data fetching
type MockDataFetcher struct {
	latency time.Duration
}

func NewMockDataFetcher(latency time.Duration) *MockDataFetcher {
	return &MockDataFetcher{latency: latency}
}

func (f *MockDataFetcher) Fetch(ctx context.Context, key string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(f.latency):
		return "data-" + key, nil
	}
}

// TestParallelFetchPerformance tests parallel fetching performance
func TestParallelFetchPerformance(t *testing.T) {
	fetcher := NewMockDataFetcher(50 * time.Millisecond)
	keys := []string{"key1", "key2", "key3", "key4", "key5"}

	// Sequential fetch
	start := time.Now()
	for _, key := range keys {
		_, _ = fetcher.Fetch(context.Background(), key)
	}
	sequentialTime := time.Since(start)

	// Parallel fetch
	start = time.Now()
	var wg sync.WaitGroup
	results := make([]string, len(keys))
	for i, key := range keys {
		wg.Add(1)
		go func(idx int, k string) {
			defer wg.Done()
			result, _ := fetcher.Fetch(context.Background(), k)
			results[idx] = result
		}(i, key)
	}
	wg.Wait()
	parallelTime := time.Since(start)

	// Parallel should be faster
	if parallelTime >= sequentialTime {
		t.Errorf("Parallel fetch not faster: sequential=%v, parallel=%v", sequentialTime, parallelTime)
	}

	improvement := float64(sequentialTime-parallelTime) / float64(sequentialTime) * 100
	t.Logf("Parallel Fetch Performance:")
	t.Logf("  Sequential Time: %v", sequentialTime)
	t.Logf("  Parallel Time: %v", parallelTime)
	t.Logf("  Improvement: %.2f%%", improvement)
}

// TestParallelFetchWithContext tests parallel fetching with context cancellation
func TestParallelFetchWithContext(t *testing.T) {
	fetcher := NewMockDataFetcher(100 * time.Millisecond)
	keys := []string{"key1", "key2", "key3"}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	errors := 0

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			_, err := fetcher.Fetch(ctx, k)
			if err != nil {
				errors++
			}
		}(key)
	}

	wg.Wait()

	// Some requests should timeout
	if errors == 0 {
		t.Error("Expected some timeouts, got 0")
	}

	t.Logf("Context Cancellation Test:")
	t.Logf("  Total Requests: %d", len(keys))
	t.Logf("  Timeouts: %d", errors)
}

// TestParallelFetchConcurrencyLimit tests parallel fetching with concurrency limit
func TestParallelFetchConcurrencyLimit(t *testing.T) {
	fetcher := NewMockDataFetcher(10 * time.Millisecond)
	keys := make([]string, 100)
	for i := range keys {
		keys[i] = "key" + string(rune(i))
	}

	concurrencyLimit := 10
	semaphore := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	start := time.Now()
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			_, _ = fetcher.Fetch(context.Background(), k)
		}(key)
	}
	wg.Wait()
	duration := time.Since(start)

	t.Logf("Concurrency Limited Fetch:")
	t.Logf("  Total Requests: %d", len(keys))
	t.Logf("  Concurrency Limit: %d", concurrencyLimit)
	t.Logf("  Total Time: %v", duration)
	t.Logf("  Requests/Second: %.2f", float64(len(keys))/duration.Seconds())
}

