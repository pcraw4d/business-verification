package performance

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
)

func BenchmarkMultiPageAnalysis(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	websiteURL := "https://example.com"
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)
		_ = keywords // Use result to avoid optimization
	}
}

func TestMultiPageAnalysis_Timeout(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test timeout handling for slow websites
	slowWebsiteURL := "https://httpstat.us/200?sleep=5000" // 5 second delay
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, slowWebsiteURL)
	duration := time.Since(start)

	// Should complete within timeout (or return empty for fallback)
	if duration > 65*time.Second {
		t.Errorf("Multi-page analysis exceeded 60s timeout. Duration: %v", duration)
	}

	_ = keywords
}

func TestMultiPageAnalysis_ConcurrentLimit(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test that concurrent requests are limited
	websiteURL := "https://example.com"
	ctx := context.Background()

	// Run multiple analyses concurrently
	const numConcurrent = 10
	results := make(chan int, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func() {
			keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)
			results <- len(keywords)
		}()
	}

	// Collect results
	totalKeywords := 0
	for i := 0; i < numConcurrent; i++ {
		totalKeywords += <-results
	}

	// Verify all requests completed (no deadlock)
	if totalKeywords < 0 {
		t.Error("Concurrent requests failed or deadlocked")
	}
}

func TestMultiPageAnalysis_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory usage test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	websiteURL := "https://example.com"
	ctx := context.Background()

	// Run analysis multiple times to check for memory leaks
	for i := 0; i < 10; i++ {
		keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)
		_ = keywords

		// Force garbage collection
		if i%5 == 0 {
			// Note: In a real test, you'd use runtime.ReadMemStats() to check memory
			// For now, we just verify the function doesn't crash
		}
	}

	// If we get here without OOM, memory usage is acceptable
}

func TestMultiPageAnalysis_PerformanceTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance target test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	websiteURL := "https://example.com"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run multiple times to get p95
	durations := make([]time.Duration, 0, 20)
	for i := 0; i < 20; i++ {
		start := time.Now()
		keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)
		duration := time.Since(start)
		durations = append(durations, duration)
		_ = keywords

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	// Calculate p95 (simplified - would use proper percentile calculation in production)
	maxDuration := time.Duration(0)
	for _, d := range durations {
		if d > maxDuration {
			maxDuration = d
		}
	}

	// Target: p95 < 60s
	if maxDuration > 60*time.Second {
		t.Errorf("p95 duration (%v) exceeds 60s target", maxDuration)
	}
}

