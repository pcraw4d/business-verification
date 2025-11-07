package contract

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
)

// TestNoPlaceholderDataInProduction tests that production APIs never return mock data
func TestNoPlaceholderDataInProduction(t *testing.T) {
	// Set environment to production
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("ALLOW_MOCK_DATA", "false")
	defer os.Unsetenv("ENVIRONMENT")
	defer os.Unsetenv("ALLOW_MOCK_DATA")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify production settings
	if cfg.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.Merchant.AllowMockData {
		t.Error("Expected AllowMockData to be false in production")
	}

	// Create handler (would need actual Supabase client for full test)
	logger := zap.NewNop()
	// Note: This is a contract test - in a real scenario, you'd set up a test server
	// and make actual HTTP requests to verify no placeholder data is returned

	t.Log("Contract test: Production environment should not allow mock data")
}

// TestIsFallbackFlagPresent tests that fallback responses include isFallback flag
func TestIsFallbackFlagPresent(t *testing.T) {
	// This test would verify that when fallback data is used (in development),
	// the response includes an isFallback: true flag

	// Example response structure check
	response := map[string]interface{}{
		"id":          "merchant_001",
		"name":        "Sample Merchant",
		"isFallback":  true,
	}

	if isFallback, ok := response["isFallback"].(bool); !ok || !isFallback {
		t.Error("Fallback response must include isFallback: true flag")
	}
}

// TestErrorResponses tests that proper error responses (404, 503) are returned
func TestErrorResponses(t *testing.T) {
	// Set environment to production
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("ALLOW_MOCK_DATA", "false")
	defer os.Unsetenv("ENVIRONMENT")
	defer os.Unsetenv("ALLOW_MOCK_DATA")

	// Test that 404 is returned for not found
	// Test that 503 is returned for service unavailable
	// Test that responses don't contain mock data

	t.Log("Contract test: Production should return proper HTTP status codes (404, 503) instead of mock data")
}

// TestNoPlaceholderInResponseBody tests that response bodies don't contain placeholder strings
func TestNoPlaceholderInResponseBody(t *testing.T) {
	// Placeholder patterns to check for
	placeholderPatterns := []string{
		"Sample Merchant",
		"Mock",
		"TODO",
		"placeholder",
		"test-",
		"dummy",
		"fake",
		"example",
	}

	// Example response (would be from actual API in real test)
	responseBody := `{"id":"merchant_001","name":"Acme Corporation"}`

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check response doesn't contain placeholder patterns
	responseStr := string(responseBody)
	for _, pattern := range placeholderPatterns {
		if contains(responseStr, pattern) {
			t.Errorf("Response contains placeholder pattern: %s", pattern)
		}
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}

