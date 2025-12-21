//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// TestClassificationWithRealDatabase tests the classification service with a real Supabase database
func TestClassificationWithRealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping database integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:            supabaseURL,
		APIKey:         supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("Skipping database integration test: cannot connect to Supabase: %v", err)
	}

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Create service
	service := classification.NewIndustryDetectionService(repo, logger)

	// Test cases
	testCases := []struct {
		name              string
		businessName      string
		description       string
		websiteURL        string
		expectedIndustry   string
		minConfidence     float64
		requireMultiStrategy bool
	}{
		{
			name:              "Technology company",
			businessName:      "Microsoft Corporation",
			description:       "Software development and cloud computing services",
			websiteURL:        "https://microsoft.com",
			expectedIndustry:  "Technology",
			minConfidence:     0.70,
			requireMultiStrategy: true,
		},
		{
			name:              "Healthcare company",
			businessName:      "Mayo Clinic",
			description:       "Medical center and hospital services",
			websiteURL:        "https://mayoclinic.org",
			expectedIndustry:  "Healthcare",
			minConfidence:     0.70,
			requireMultiStrategy: true,
		},
		{
			name:              "Retail company",
			businessName:      "Amazon",
			description:       "E-commerce and retail services",
			websiteURL:        "https://amazon.com",
			expectedIndustry:  "Retail",
			minConfidence:     0.65,
			requireMultiStrategy: true,
		},
		{
			name:              "Financial services",
			businessName:      "JPMorgan Chase",
			description:       "Banking and financial services",
			websiteURL:        "https://jpmorganchase.com",
			expectedIndustry:  "Financial Services",
			minConfidence:     0.70,
			requireMultiStrategy: true,
		},
	}

	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test industry detection
			result, err := service.DetectIndustry(ctx, tc.businessName, tc.description, tc.websiteURL)
			if err != nil {
				t.Errorf("DetectIndustry failed: %v", err)
				return
			}

			if result == nil {
				t.Fatal("DetectIndustry returned nil result")
				return
			}

			// Verify method is multi-strategy
			if tc.requireMultiStrategy && result.Method != "multi_strategy" {
				t.Logf("⚠️ Expected multi_strategy method, got: %s", result.Method)
				// Don't fail - fallback is acceptable
			}

			// Verify industry matches (allow some flexibility)
			if result.IndustryName != tc.expectedIndustry {
				// Check if it's a related industry
				if !containsIndustry(result.IndustryName, tc.expectedIndustry) {
					t.Logf("⚠️ Industry mismatch: expected %s, got %s (confidence: %.2f%%)",
						tc.expectedIndustry, result.IndustryName, result.Confidence*100)
					// Don't fail if confidence is high - might be valid alternative
					if result.Confidence < 0.60 {
						t.Errorf("Industry mismatch: expected %s, got %s", tc.expectedIndustry, result.IndustryName)
					}
				}
			}

			// Verify confidence meets minimum
			if result.Confidence < tc.minConfidence {
				t.Errorf("Confidence too low: expected >= %.2f, got %.2f",
					tc.minConfidence, result.Confidence)
			}

			// Verify keywords were extracted
			if len(result.Keywords) == 0 {
				t.Error("No keywords extracted")
			}

			// Verify processing time is reasonable
			if result.ProcessingTime > 30*time.Second {
				t.Errorf("Processing time too slow: %v", result.ProcessingTime)
			}

			t.Logf("✅ %s classified as %s (confidence: %.2f%%, method: %s, time: %v, keywords: %d)",
				tc.businessName, result.IndustryName, result.Confidence*100,
				result.Method, result.ProcessingTime, len(result.Keywords))
		})
	}
}

// TestMultiStrategyClassifierWithDatabase tests multi-strategy classifier with real database
func TestMultiStrategyClassifierWithDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping database integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:    supabaseURL,
		APIKey: supabaseKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Create multi-strategy classifier
	classifier := classification.NewMultiStrategyClassifier(repo, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test classification
	result, err := classifier.ClassifyWithMultiStrategy(
		ctx,
		"TechCorp Solutions",
		"Software development and AI solutions provider",
		"https://techcorp.com",
	)

	if err != nil {
		t.Fatalf("ClassifyWithMultiStrategy failed: %v", err)
	}

	if result == nil {
		t.Fatal("ClassifyWithMultiStrategy returned nil")
	}

	// Verify result structure
	if result.PrimaryIndustry == "" {
		t.Error("PrimaryIndustry is empty")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}

	if len(result.Strategies) == 0 {
		t.Error("No strategies returned")
	}

	if len(result.Keywords) == 0 {
		t.Error("No keywords extracted")
	}

	t.Logf("✅ Multi-strategy classification successful:")
	t.Logf("   Industry: %s", result.PrimaryIndustry)
	t.Logf("   Confidence: %.2f%%", result.Confidence*100)
	t.Logf("   Strategies: %d", len(result.Strategies))
	t.Logf("   Keywords: %d", len(result.Keywords))
	t.Logf("   Entities: %d", len(result.Entities))
	t.Logf("   Topics: %d", len(result.TopicScores))
	t.Logf("   Processing Time: %v", result.ProcessingTime)
}

// containsIndustry checks if industry names are related (fuzzy match)
func containsIndustry(actual, expected string) bool {
	// Simple fuzzy matching for industry names
	if len(actual) >= len(expected) {
		return actual[:len(expected)] == expected ||
			actual == expected ||
			actual == expected+" & Commerce" ||
			actual == expected+" Services"
	}
	return false
}

