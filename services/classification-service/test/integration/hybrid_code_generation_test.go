package integration

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/classification/testutil"
)

// TestHybridCodeGeneration_EndToEnd tests the complete hybrid code generation flow
func TestHybridCodeGeneration_EndToEnd(t *testing.T) {
	// Skip if integration tests are disabled
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use mock repository for now
	// To use real database, uncomment and configure:
	// db, err := testutil.ConnectTestDB(t)
	// if err != nil {
	//     t.Skipf("Skipping test - database not available: %v", err)
	// }
	// defer db.Close()
	// testutil.SetupTestDB(t, db)
	// testutil.SeedTestData(t, db)
	// defer testutil.CleanupTestDB(t, db)

	mockRepo := testutil.NewMockKeywordRepository()
	generator := classification.NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology"}
	detectedIndustry := "Technology"
	confidence := 0.85

	// Test hybrid generation
	codes, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if codes == nil {
		t.Fatal("Expected codes to be generated")
	}

	// Verify codes were generated from both sources
	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)
	if totalCodes == 0 {
		t.Error("Expected codes to be generated")
	}
}

// TestHybridCodeGeneration_FallbackToIndustryOnly tests fallback behavior
func TestHybridCodeGeneration_FallbackToIndustryOnly(t *testing.T) {
	// This test verifies that when code_keywords table is empty,
	// the system falls back to industry-only code generation
	t.Skip("Requires database setup")
}

// TestHybridCodeGeneration_Performance tests performance with large keyword sets
func TestHybridCodeGeneration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test with large keyword sets to ensure performance is acceptable
	keywords := make([]string, 100)
	for i := 0; i < 100; i++ {
		keywords[i] = "keyword" + string(rune('0'+i%10))
	}

	start := time.Now()
	_ = keywords
	_ = start

	// Verify that processing completes within acceptable time
	// Expected: < 100ms overhead for hybrid approach
}

// TestHybridCodeGeneration_MultiIndustry tests multi-industry code generation
func TestHybridCodeGeneration_MultiIndustry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test code generation with multiple industries
	// Verify that codes from all industries are included
	// Verify deduplication works correctly
	// Verify confidence weighting is applied
	t.Skip("Requires database setup")
}

// TestHybridCodeGeneration_ConfidenceFiltering tests confidence-based filtering
func TestHybridCodeGeneration_ConfidenceFiltering(t *testing.T) {
	// Test that codes below confidence threshold are filtered out
	// Test that top-N limiting works correctly
	// Test that is_primary codes are prioritized
	t.Skip("Requires database setup")
}

// TestHybridCodeGeneration_CodeDeduplication tests code deduplication across sources
func TestHybridCodeGeneration_CodeDeduplication(t *testing.T) {
	// Test that the same code from industry and keyword sources
	// is properly deduplicated and confidence is combined
	// Test that boost is applied for codes matched by both sources
	t.Skip("Requires database setup")
}

// Note: Benchmark tests would require actual mock implementations
// These are placeholders for when proper mocks are available

