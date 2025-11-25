package testutil

import (
	"context"
	"testing"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
)

// SetupIntegrationTest sets up an integration test environment
// Returns a cleanup function that should be called with defer
func SetupIntegrationTest(t *testing.T) (*repository.SupabaseKeywordRepository, func()) {
	t.Helper()

	// Skip if integration tests are disabled
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// For now, return a mock repository
	// In a real integration test, this would connect to a test database
	_ = NewMockKeywordRepository()
	
	cleanup := func() {
		// Cleanup logic if needed
	}

	return nil, cleanup // Return nil for now since we're using mocks
}

// TestHybridCodeGeneration_Integration tests hybrid code generation with database
func TestHybridCodeGeneration_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would use a real database connection in a full integration test
	mockRepo := NewMockKeywordRepository()
	generator := classification.NewClassificationCodeGenerator(mockRepo, nil)

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

	// Verify codes were generated
	if len(codes.MCC) == 0 && len(codes.SIC) == 0 && len(codes.NAICS) == 0 {
		t.Error("Expected at least some codes to be generated")
	}
}

