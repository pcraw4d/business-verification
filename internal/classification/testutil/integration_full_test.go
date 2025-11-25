package testutil

import (
	"context"
	"os"
	"testing"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestHybridCodeGeneration_WithRealRepository tests hybrid code generation with real Supabase repository
// This test requires a configured Supabase connection
func TestHybridCodeGeneration_WithRealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if Supabase is configured
	// Try SUPABASE_ANON_KEY first, fallback to SUPABASE_API_KEY for compatibility
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_API_KEY") // Fallback for development.env format
	}
	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping test - Supabase not configured (set SUPABASE_URL and SUPABASE_ANON_KEY or SUPABASE_API_KEY)")
	}

	// Create real Supabase client
	cfg := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"), // Optional
	}
	supabaseClient, err := database.NewSupabaseClient(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	
	// Connect to Supabase
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to Supabase: %v", err)
	}
	defer supabaseClient.Close()

	repo := repository.NewSupabaseKeywordRepository(supabaseClient, nil)
	generator := classification.NewClassificationCodeGenerator(repo, nil)

	keywords := []string{"software", "technology", "platform"}
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
	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)
	if totalCodes == 0 {
		t.Error("Expected at least some codes to be generated")
	}

	t.Logf("Generated %d MCC, %d SIC, %d NAICS codes", len(codes.MCC), len(codes.SIC), len(codes.NAICS))
}

// TestHybridCodeGeneration_MultiIndustry_WithRealRepository tests multi-industry with real repository
func TestHybridCodeGeneration_MultiIndustry_WithRealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_API_KEY") // Fallback for development.env format
	}
	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping test - Supabase not configured")
	}

	cfg := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	}
	supabaseClient, err := database.NewSupabaseClient(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to Supabase: %v", err)
	}
	defer supabaseClient.Close()

	repo := repository.NewSupabaseKeywordRepository(supabaseClient, nil)
	generator := classification.NewClassificationCodeGenerator(repo, nil)

	keywords := []string{"software", "technology", "finance", "banking"}
	detectedIndustry := "Technology"
	confidence := 0.85
	additionalIndustries := []classification.IndustryResult{
		{IndustryName: "Software", Confidence: 0.75},
		{IndustryName: "Financial Services", Confidence: 0.70},
	}

	// Test multi-industry generation
	codes, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence, additionalIndustries...)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if codes == nil {
		t.Fatal("Expected codes to be generated")
	}

	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)
	if totalCodes == 0 {
		t.Error("Expected codes to be generated from multiple industries")
	}

	t.Logf("Multi-industry: Generated %d MCC, %d SIC, %d NAICS codes", len(codes.MCC), len(codes.SIC), len(codes.NAICS))
}

// TestKeywordCodeLookup_WithRealRepository tests keyword-based lookup with real repository
func TestKeywordCodeLookup_WithRealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_API_KEY") // Fallback for development.env format
	}
	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping test - Supabase not configured")
	}

	cfg := &database.SupabaseConfig{
		URL:           supabaseURL,
		APIKey:        supabaseKey,
		ServiceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	}
	supabaseClient, err := database.NewSupabaseClient(cfg, nil)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to Supabase: %v", err)
	}
	defer supabaseClient.Close()

	repo := repository.NewSupabaseKeywordRepository(supabaseClient, nil)

	keywords := []string{"software", "technology"}
	codeType := "MCC"
	minRelevance := 0.5

	// Test keyword lookup
	codes, err := repo.GetClassificationCodesByKeywords(ctx, keywords, codeType, minRelevance)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Logf("Found %d codes for keywords: %v", len(codes), keywords)

	// Verify codes have metadata
	for _, code := range codes {
		if code.RelevanceScore < minRelevance {
			t.Errorf("Code %s has relevance score %.2f below threshold %.2f", code.Code, code.RelevanceScore, minRelevance)
		}
		if code.CodeType != codeType {
			t.Errorf("Code %s has wrong type: expected %s, got %s", code.Code, codeType, code.CodeType)
		}
	}
}

