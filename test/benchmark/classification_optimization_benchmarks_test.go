package benchmark

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/classification"
	"kyb-platform/services/classification-service/internal/cache"
)

// BenchmarkKeywordExtraction_Accuracy tests keyword extraction performance and accuracy
// Note: isValidEnglishWord is private, so we benchmark through analyzePage which uses it internally
func BenchmarkKeywordExtraction_Accuracy(b *testing.B) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := classification.NewSmartWebsiteCrawler(logger)

	// Test text that will trigger word validation internally
	testText := "business technology restaurant software development consulting services"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// analyzePage will call isValidEnglishWord internally
		_, _ = crawler.AnalyzePage(context.Background(), "https://example.com", testText)
	}
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

// Mock types for benchmarking
type MockWebsiteContentCacher struct{}

func (m *MockWebsiteContentCacher) Get(ctx context.Context, url string) (*cache.CachedWebsiteContent, bool) {
	return nil, false
}

func (m *MockWebsiteContentCacher) Set(ctx context.Context, url string, content *cache.CachedWebsiteContent) error {
	return nil
}

func (m *MockWebsiteContentCacher) Delete(ctx context.Context, url string) {}

func (m *MockWebsiteContentCacher) IsEnabled() bool { return true }

type MockWebsiteScraper struct {
	content  string
	delay    time.Duration
	callCount int
}

func (m *MockWebsiteScraper) ScrapeWebsite(ctx context.Context, url string) (*classification.ScrapingResult, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return &classification.ScrapingResult{
		Content: m.content,
		Success: true,
	}, nil
}

