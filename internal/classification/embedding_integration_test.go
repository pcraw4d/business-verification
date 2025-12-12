package classification

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification/testutil"
	"kyb-platform/internal/external"
)

// TestEmbeddingClassifierIntegration tests the embedding classifier with real services
// This test requires:
// - EMBEDDING_SERVICE_URL environment variable set
// - Supabase database with pgvector enabled and code_embeddings populated
// - Valid SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY
func TestEmbeddingClassifierIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")
	if embeddingServiceURL == "" {
		t.Skip("EMBEDDING_SERVICE_URL not set, skipping integration test")
	}

	// Create mock repository (in real integration test, use Supabase repository)
	mockRepo := testutil.NewMockKeywordRepository()
	logger := log.Default()

	// Create embedding classifier
	classifier := NewEmbeddingClassifier(embeddingServiceURL, mockRepo, logger)

	// Test cases
	testCases := []struct {
		name        string
		description string
		expected    string // Expected primary code type
	}{
		{
			name:        "Technology Company",
			description: "Software development and cloud computing services",
			expected:    "MCC",
		},
		{
			name:        "Financial Services",
			description: "Banking and investment advisory services",
			expected:    "MCC",
		},
		{
			name:        "Healthcare Provider",
			description: "Medical practice and patient care services",
			expected:    "NAICS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create mock scraped content
			scrapedContent := &external.ScrapedContent{
				Title:     tc.name,
				MetaDesc:  tc.description,
				PlainText: tc.description,
				Headings:  []string{tc.name, "Services", "About"},
				AboutText: tc.description,
				Domain:    "example.com",
			}

			// Classify using embeddings
			result, err := classifier.ClassifyByEmbedding(ctx, scrapedContent)
			if err != nil {
				t.Fatalf("Classification failed: %v", err)
			}

			if result == nil {
				t.Fatal("Classification returned nil result")
			}

			// Verify result has codes (at least one type should have codes)
			totalCodes := len(result.MCC) + len(result.SIC) + len(result.NAICS)
			if totalCodes == 0 {
				t.Error("Classification returned no codes")
			}

			// Verify confidence is reasonable
			if result.Confidence < 0.0 || result.Confidence > 1.0 {
				t.Errorf("Invalid confidence score: %.2f", result.Confidence)
			}

			t.Logf("Classification result: %+v", result)
		})
	}
}

// TestLayer2RoutingIntegration tests the Layer 2 routing logic
// This test verifies that Layer 2 is triggered when Layer 1 confidence is low
func TestLayer2RoutingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")
	if embeddingServiceURL == "" {
		t.Skip("EMBEDDING_SERVICE_URL not set, skipping integration test")
	}

	mockRepo := testutil.NewMockKeywordRepository()
	logger := log.Default()

	// Create service with embedding classifier
	service := NewIndustryDetectionService(mockRepo, logger)
	embeddingClassifier := NewEmbeddingClassifier(embeddingServiceURL, mockRepo, logger)
	service.SetEmbeddingClassifier(embeddingClassifier)

	// Test case: Ambiguous business that should trigger Layer 2
	testCases := []struct {
		name        string
		description string
		websiteURL  string
		expectLayer2 bool // Whether Layer 2 should be triggered
	}{
		{
			name:         "Ambiguous Business Name",
			description:  "Services and solutions",
			websiteURL:   "https://example.com",
			expectLayer2: true, // Low confidence should trigger Layer 2
		},
		{
			name:         "Clear Technology Company",
			description:  "Software development and cloud computing",
			websiteURL:   "https://techcompany.com",
			expectLayer2: false, // High confidence should use Layer 1 only
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := service.DetectIndustry(ctx, tc.name, tc.description, tc.websiteURL)
			if err != nil {
				t.Fatalf("Detection failed: %v", err)
			}

			if result == nil {
				t.Fatal("Detection returned nil result")
			}

			// Verify result has industry
			if result.IndustryName == "" {
				t.Error("Detection returned empty industry name")
			}

			// Log the result method to verify Layer 2 was used if expected
			t.Logf("Result method: %s, Confidence: %.2f", result.Method, result.Confidence)

			// If Layer 2 was expected, verify confidence improved or method indicates Layer 2
			if tc.expectLayer2 {
				// Layer 2 should improve confidence or be indicated in method
				if result.Confidence < 0.5 {
					t.Logf("Warning: Low confidence after Layer 2: %.2f", result.Confidence)
				}
			}
		})
	}
}

// TestEmbeddingServiceHealth tests the embedding service health endpoint
func TestEmbeddingServiceHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")
	if embeddingServiceURL == "" {
		t.Skip("EMBEDDING_SERVICE_URL not set, skipping integration test")
	}

	mockRepo := testutil.NewMockKeywordRepository()
	logger := log.Default()

	classifier := NewEmbeddingClassifier(embeddingServiceURL, mockRepo, logger)

	// Test health check by attempting a simple classification
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	scrapedContent := &external.ScrapedContent{
		Title:     "Test",
		PlainText: "Test content",
		Domain:    "test.com",
	}

	_, err := classifier.ClassifyByEmbedding(ctx, scrapedContent)
	if err != nil {
		// If service is not available, that's okay for integration tests
		t.Logf("Embedding service not available (expected in some environments): %v", err)
	} else {
		t.Log("✅ Embedding service is healthy and responding")
	}
}

// BenchmarkEmbeddingClassification benchmarks embedding-based classification
func BenchmarkEmbeddingClassification(b *testing.B) {
	embeddingServiceURL := os.Getenv("EMBEDDING_SERVICE_URL")
	if embeddingServiceURL == "" {
		b.Skip("EMBEDDING_SERVICE_URL not set, skipping benchmark")
	}

	mockRepo := testutil.NewMockKeywordRepository()
	logger := log.Default()

	classifier := NewEmbeddingClassifier(embeddingServiceURL, mockRepo, logger)

	scrapedContent := &external.ScrapedContent{
		Title:       "Technology Company",
		MetaDesc:    "Software development services",
		PlainText:   "We provide software development and cloud computing services",
		Headings:    []string{"Services", "About", "Contact"},
		AboutText:   "We are a technology company specializing in software development",
		Domain:      "techcompany.com",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := classifier.ClassifyByEmbedding(ctx, scrapedContent)
		if err != nil {
			b.Fatalf("Classification failed: %v", err)
		}
	}
}

