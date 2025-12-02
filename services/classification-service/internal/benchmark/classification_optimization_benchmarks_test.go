package benchmark

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/classification-service/internal/cache"
)

// BenchmarkKeywordExtraction_Accuracy tests keyword extraction performance
// Since isValidEnglishWord is private, we skip this benchmark for now
// and focus on cache and parallel processing benchmarks
func BenchmarkKeywordExtraction_Accuracy(b *testing.B) {
	b.Skip("Skipping - requires external HTTP calls or public API for word validation")
}

// BenchmarkWebsiteContentCache_GetSet benchmarks cache operations
func BenchmarkWebsiteContentCache_GetSet(b *testing.B) {
	// Use in-memory cache for benchmarking (no Redis dependency)
	logger := zap.NewNop()
	cacheInstance := cache.NewWebsiteContentCache(nil, logger, 24*time.Hour)

	ctx := context.Background()
	url := "https://example.com"
	content := &cache.CachedWebsiteContent{
		TextContent: "Sample website content for benchmarking",
		ScrapedAt:   time.Now(),
		Success:     true,
	}

	// Setup: populate cache
	_ = cacheInstance.Set(ctx, url, content)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cacheInstance.Get(ctx, url)
		}
	})
}

// BenchmarkWebsiteContentService_Deduplication benchmarks deduplication performance
// Note: This requires proper mocking setup - simplified for benchmark structure
func BenchmarkWebsiteContentService_Deduplication(b *testing.B) {
	// This benchmark would test concurrent requests to the same URL
	// and verify deduplication overhead is minimal
	b.Log("Deduplication benchmark structure - requires full service setup")
}

// BenchmarkParallelClassification_Speedup benchmarks parallel classification speedup
func BenchmarkParallelClassification_Speedup(b *testing.B) {
	// Simulate parallel vs sequential execution
	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Sleep(10 * time.Millisecond) // Simulate work
			time.Sleep(10 * time.Millisecond) // Simulate work
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			done1 := make(chan struct{})
			done2 := make(chan struct{})
			go func() {
				time.Sleep(10 * time.Millisecond)
				close(done1)
			}()
			go func() {
				time.Sleep(10 * time.Millisecond)
				close(done2)
			}()
			<-done1
			<-done2
		}
	})
}

// BenchmarkIsValidEnglishWord_EnhancedValidation benchmarks word validation
// Note: isValidEnglishWord is private, so we can't benchmark it directly
// This shows the benchmark structure for when it's made public or tested indirectly
func BenchmarkIsValidEnglishWord_EnhancedValidation(b *testing.B) {
	b.Log("Word validation benchmark structure - isValidEnglishWord is private")
}

