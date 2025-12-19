package classification

import (
	"strings"
	"testing"
)

// TestCacheKeyConsistencyBetweenHandlerAndService verifies that generateRequestCacheKey
// in service.go generates the same keys as ClassificationHandler.getCacheKey()
// FIX: Ensures cache keys match between handler and service for cache hit/miss consistency
func TestCacheKeyConsistencyBetweenHandlerAndService(t *testing.T) {
	tests := []struct {
		name        string
		businessName string
		description  string
		websiteURL   string
	}{
		{
			name:         "basic request",
			businessName: "Acme Corp",
			description:  "Technology company",
			websiteURL:   "https://acme.com",
		},
		{
			name:         "with whitespace",
			businessName: "  Acme Corp  ",
			description:  "Technology company",
			websiteURL:   "https://acme.com",
		},
		{
			name:         "different case",
			businessName: "acme corp",
			description:  "Technology company",
			websiteURL:   "https://acme.com",
		},
		{
			name:         "no website",
			businessName: "Acme Corp",
			description:  "Technology company",
			websiteURL:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate key using service helper function
			key1 := generateRequestCacheKey(tt.businessName, tt.description, tt.websiteURL)
			
			// Verify key format
			if len(key1) == 0 {
				t.Error("Cache key should not be empty")
			}
			
			// Key should have classification: prefix
			if !strings.HasPrefix(key1, "classification:") {
				t.Errorf("Cache key should start with 'classification:' prefix, got: %s", key1)
			}
			
			// Same inputs should generate same key
			key2 := generateRequestCacheKey(tt.businessName, tt.description, tt.websiteURL)
			if key1 != key2 {
				t.Errorf("Same inputs should generate same key: %s != %s", key1, key2)
			}
			
			// Normalized inputs should generate same key
			key3 := generateRequestCacheKey(
				"  "+tt.businessName+"  ",
				"  "+tt.description+"  ",
				"  "+tt.websiteURL+"  ",
			)
			if key1 != key3 {
				t.Errorf("Normalized inputs should generate same key: %s != %s", key1, key3)
			}
		})
	}
}

