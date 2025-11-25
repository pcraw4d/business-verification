package repository

import (
	"context"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestKeywordExtractionPerformance_Phase10_3 tests keyword extraction performance (Phase 10.3)
func TestKeywordExtractionPerformance_Phase10_3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	// Test 1: Keyword extraction time for multiple pages (simulated)
	t.Run("keyword_extraction_time_multiple_pages", func(t *testing.T) {
		// Simulate processing multiple pages of content
		pageContents := []string{
			"<html><head><title>Wine Shop - Premium Wines</title></head><body><h1>Welcome</h1><p>We sell fine wines and beverages.</p></body></html>",
			"<html><head><title>About Us</title></head><body><h1>About</h1><p>We are a retail store specializing in wine and spirits.</p></body></html>",
			"<html><head><title>Products</title></head><body><h1>Products</h1><p>Browse our selection of wines, spirits, and beverages.</p></body></html>",
		}

		start := time.Now()
		totalKeywords := 0
		for i, content := range pageContents {
			keywords := repo.extractBusinessKeywords(repo.extractTextFromHTML(content))
			totalKeywords += len(keywords)
			if i == 0 {
				// First page should be fast (cached regex)
				firstPageTime := time.Since(start)
				if firstPageTime > 100*time.Millisecond {
					t.Errorf("First page extraction took too long: %v (expected < 100ms)", firstPageTime)
				}
			}
		}
		totalTime := time.Since(start)

		// Should process 3 pages in reasonable time
		if totalTime > 500*time.Millisecond {
			t.Errorf("Processing 3 pages took too long: %v (expected < 500ms)", totalTime)
		}

		// Should extract keywords from all pages
		if totalKeywords < 5 {
			t.Errorf("Expected at least 5 keywords from 3 pages, got %d", totalKeywords)
		}

		t.Logf("Performance: Processed %d pages in %v, extracted %d keywords (avg: %.2fms/page)", len(pageContents), totalTime, totalKeywords, float64(totalTime.Nanoseconds())/float64(len(pageContents))/1e6)
	})

	// Test 2: DNS resolution time with retries
	t.Run("dns_resolution_time_with_retries", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resolver := &net.Resolver{
			PreferGo: true,
		}

		hosts := []string{"google.com", "github.com", "example.com"}

		start := time.Now()
		for _, host := range hosts {
			// First call (cache miss)
			_, err := repo.getCachedDNSResolution(host, resolver, ctx)
			if err != nil {
				t.Logf("DNS resolution failed for %s (may be network issue): %v", host, err)
				continue
			}

			// Second call (cache hit - should be much faster)
			cacheStart := time.Now()
			_, err = repo.getCachedDNSResolution(host, resolver, ctx)
			cacheTime := time.Since(cacheStart)
			if err != nil {
				t.Logf("Cached DNS resolution failed for %s: %v", host, err)
				continue
			}

			// Cached lookup should be very fast (< 10ms)
			if cacheTime > 10*time.Millisecond {
				t.Errorf("Cached DNS lookup for %s took too long: %v (expected < 10ms)", host, cacheTime)
			}
		}
		totalTime := time.Since(start)

		t.Logf("DNS resolution: Resolved %d hosts in %v (with caching)", len(hosts), totalTime)
	})

	// Test 3: Memory usage during keyword extraction
	t.Run("memory_usage_during_extraction", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Process large content
		largeContent := strings.Repeat("wine shop retail store beverage spirits alcohol ", 1000)
		keywords := repo.extractBusinessKeywords(largeContent)

		runtime.GC()
		runtime.ReadMemStats(&m2)

		// Memory should not grow excessively
		memUsed := int64(m2.Alloc - m1.Alloc)
		maxMemBytes := int64(10 * 1024 * 1024) // 10MB max

		if memUsed > maxMemBytes {
			t.Errorf("Memory usage too high: %d bytes (expected < %d bytes)", memUsed, maxMemBytes)
		}

		// Should still extract keywords
		if len(keywords) == 0 {
			t.Error("Expected keywords from large content, got none")
		}

		t.Logf("Memory usage: %d bytes for processing %d characters, extracted %d keywords", memUsed, len(largeContent), len(keywords))
	})

	// Test 4: Concurrent request handling
	t.Run("concurrent_request_handling", func(t *testing.T) {
		domains := []string{"example1.com", "example2.com", "example3.com", "example4.com", "example5.com"}

		start := time.Now()
		done := make(chan bool, len(domains))

		// Apply rate limiting concurrently
		for _, domain := range domains {
			go func(d string) {
				repo.applyRateLimit(d)
				done <- true
			}(domain)
		}

		// Wait for all goroutines
		for i := 0; i < len(domains); i++ {
			<-done
		}
		concurrentTime := time.Since(start)

		// Sequential rate limiting (for comparison)
		start = time.Now()
		for _, domain := range domains {
			repo.applyRateLimit(domain)
		}
		sequentialTime := time.Since(start)

		// Concurrent should not be significantly slower (rate limiting may serialize)
		if concurrentTime > sequentialTime*2 {
			t.Errorf("Concurrent rate limiting took too long: %v vs sequential %v", concurrentTime, sequentialTime)
		}

		t.Logf("Concurrent handling: %d domains in %v (sequential: %v)", len(domains), concurrentTime, sequentialTime)
	})

	// Test 5: Regex caching performance
	t.Run("regex_caching_performance", func(t *testing.T) {
		patterns := []string{
			`\b(wine|shop)\b`,
			`\b(retail|store)\b`,
			`\b(technology|software)\b`,
			`\b(healthcare|medical)\b`,
		}

		// First pass (compile and cache)
		start := time.Now()
		for _, pattern := range patterns {
			_ = repo.getCachedRegex(pattern)
		}
		firstPassTime := time.Since(start)

		// Second pass (use cache)
		start = time.Now()
		for _, pattern := range patterns {
			_ = repo.getCachedRegex(pattern)
		}
		secondPassTime := time.Since(start)

		// Cached pass should be significantly faster
		if secondPassTime >= firstPassTime {
			t.Errorf("Cached regex lookup not faster: first=%v, cached=%v", firstPassTime, secondPassTime)
		}

		// Cached pass should be very fast (< 1ms total)
		if secondPassTime > 1*time.Millisecond {
			t.Errorf("Cached regex lookup took too long: %v (expected < 1ms)", secondPassTime)
		}

		t.Logf("Regex caching: First pass=%v, Cached pass=%v (speedup: %.2fx)", firstPassTime, secondPassTime, float64(firstPassTime)/float64(secondPassTime))
	})
}

// BenchmarkKeywordExtraction benchmarks keyword extraction performance
func BenchmarkKeywordExtraction(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	content := "We are a wine shop and retail store selling fine wines, spirits, and beverages. Visit our wine bar for tastings."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.extractBusinessKeywords(content)
	}
}

// BenchmarkRegexCaching benchmarks regex caching performance
func BenchmarkRegexCaching(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	pattern := `\b(wine|shop|retail|store|beverage)\b`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.getCachedRegex(pattern)
	}
}

// BenchmarkHTMLTextExtraction benchmarks HTML text extraction
func BenchmarkHTMLTextExtraction(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	html := `<html><head><title>Wine Shop</title></head><body><h1>Welcome</h1><p>We sell fine wines and beverages.</p></body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.extractTextFromHTML(html)
	}
}

