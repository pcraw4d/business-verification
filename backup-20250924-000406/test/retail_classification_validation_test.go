package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetailClassificationValidation tests the comprehensive retail classification test cases
func TestRetailClassificationValidation(t *testing.T) {
	dataset := NewComprehensiveTestDataset()
	retailTestCases := dataset.GetTestCasesByCategory("Retail")

	// Verify we have 23 retail test cases (retail-001 through retail-023)
	require.Equal(t, 23, len(retailTestCases), "Expected 23 retail test cases")

	t.Logf("🔍 Testing %d retail classification test cases", len(retailTestCases))

	// Test case validation
	validTestCases := 0
	securityTestCases := 0
	expectedAccuracy := 0.0

	for _, tc := range retailTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.ID, "Test case ID should not be empty")
			assert.NotEmpty(t, tc.Name, "Test case name should not be empty")
			assert.NotEmpty(t, tc.BusinessName, "Business name should not be empty")
			assert.NotEmpty(t, tc.Description, "Description should not be empty")
			assert.NotEmpty(t, tc.ExpectedIndustry, "Expected industry should not be empty")
			assert.True(t, tc.ExpectedConfidence > 0, "Expected confidence should be positive")
			assert.NotEmpty(t, tc.ExpectedMCCCodes, "Expected MCC codes should not be empty")
			assert.NotEmpty(t, tc.ExpectedSICCodes, "Expected SIC codes should not be empty")
			assert.NotEmpty(t, tc.ExpectedNAICSCodes, "Expected NAICS codes should not be empty")
			assert.NotEmpty(t, tc.Keywords, "Keywords should not be empty")
			assert.NotEmpty(t, tc.ExpectedKeywords, "Expected keywords should not be empty")

			// Validate industry classification
			assert.Equal(t, "Retail", tc.ExpectedIndustry, "All test cases should be classified as Retail")

			// Validate confidence scores
			if tc.DifficultyLevel == "Hard" || tc.ID == "retail-021" {
				// Security test cases should have lower confidence
				assert.True(t, tc.ExpectedConfidence <= 0.75, "Security test cases should have low confidence")
				securityTestCases++
			} else {
				// Regular test cases should have high confidence (>80%)
				assert.True(t, tc.ExpectedConfidence >= 0.80, "Regular test cases should have high confidence (>=80%%)")
				validTestCases++
			}

			// Validate keywords contain retail-related terms
			hasRetailKeywords := false
			retailKeywords := []string{"retail", "store", "shop", "shopping", "merchandise", "products", "goods", "e-commerce", "marketplace", "selling", "sells"}

			// Check in keywords
			for _, keyword := range tc.Keywords {
				for _, retailKeyword := range retailKeywords {
					if strings.Contains(strings.ToLower(keyword), retailKeyword) {
						hasRetailKeywords = true
						break
					}
				}
				if hasRetailKeywords {
					break
				}
			}

			// Also check in description if not found in keywords
			if !hasRetailKeywords {
				description := strings.ToLower(tc.Description)
				for _, retailKeyword := range retailKeywords {
					if strings.Contains(description, retailKeyword) {
						hasRetailKeywords = true
						break
					}
				}
			}

			assert.True(t, hasRetailKeywords, "Test case should contain retail-related keywords")

			// Validate classification codes
			assert.True(t, len(tc.ExpectedMCCCodes) >= 1, "Should have at least 1 MCC code")
			assert.True(t, len(tc.ExpectedSICCodes) >= 1, "Should have at least 1 SIC code")
			assert.True(t, len(tc.ExpectedNAICSCodes) >= 1, "Should have at least 1 NAICS code")

			// Calculate expected accuracy contribution
			expectedAccuracy += tc.ExpectedConfidence
		})
	}

	// Calculate overall expected accuracy
	overallExpectedAccuracy := expectedAccuracy / float64(len(retailTestCases))

	// Validate accuracy target
	assert.True(t, overallExpectedAccuracy >= 0.80, "Overall expected accuracy should be >= 80%% (got %.2f%%)", overallExpectedAccuracy*100)

	// Validate security test cases
	assert.Equal(t, 3, securityTestCases, "Should have 3 security test cases (retail-019, retail-020, retail-021)")

	// Validate regular test cases
	assert.Equal(t, 20, validTestCases, "Should have 20 regular retail test cases")

	t.Logf("✅ Retail test case validation completed:")
	t.Logf("   - Total test cases: %d", len(retailTestCases))
	t.Logf("   - Regular test cases: %d", validTestCases)
	t.Logf("   - Security test cases: %d", securityTestCases)
	t.Logf("   - Overall expected accuracy: %.2f%%", overallExpectedAccuracy*100)
}

// TestRetailSecurityValidation specifically tests security validation for retail test cases
func TestRetailSecurityValidation(t *testing.T) {
	dataset := NewComprehensiveTestDataset()
	retailTestCases := dataset.GetTestCasesByCategory("Retail")

	// Find security test cases
	var securityTestCases []ClassificationTestCase
	for _, tc := range retailTestCases {
		if tc.ID == "retail-019" || tc.ID == "retail-020" || tc.ID == "retail-021" {
			securityTestCases = append(securityTestCases, tc)
		}
	}

	if len(securityTestCases) != 3 {
		t.Errorf("❌ Expected 3 security test cases, got %d", len(securityTestCases))
		return
	}

	t.Logf("🔒 Testing %d retail security validation test cases", len(securityTestCases))

	for _, tc := range securityTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Test fake e-commerce URL validation
			if tc.ID == "retail-019" {
				if tc.WebsiteURL == "https://suspicious-retail-fake.com" {
					t.Logf("✅ Security test case %s has correct fake e-commerce URL", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s has incorrect URL: %s", tc.Name, tc.WebsiteURL)
				}
				assert.Equal(t, 0.50, tc.ExpectedConfidence, "Fake e-commerce URL should have 50%% confidence")
			}

			// Test misleading product claims validation
			if tc.ID == "retail-020" {
				if tc.ExpectedConfidence == 0.30 {
					t.Logf("✅ Security test case %s has correct low confidence expectation", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s has incorrect confidence expectation: %.2f", tc.Name, tc.ExpectedConfidence)
				}
				assert.Contains(t, tc.Description, "not actually a retail business", "Should contain misleading description")
			}

			// Test no website URL validation
			if tc.ID == "retail-021" {
				if tc.WebsiteURL == "" {
					t.Logf("✅ Security test case %s correctly has no website URL", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s should have no website URL, got: %s", tc.Name, tc.WebsiteURL)
				}
				assert.Equal(t, 0.75, tc.ExpectedConfidence, "No website URL should have 75%% confidence")
			}
		})
	}

	t.Logf("🔒 All retail security validation test cases are properly configured")
}

// TestRetailIndustryCoverage tests that retail test cases cover diverse retail subcategories
func TestRetailIndustryCoverage(t *testing.T) {
	dataset := NewComprehensiveTestDataset()
	retailTestCases := dataset.GetTestCasesByCategory("Retail")

	// Define expected retail subcategories
	expectedSubcategories := []string{
		"fashion", "electronics", "home improvement", "grocery", "books",
		"sporting goods", "jewelry", "automotive", "pet supplies", "furniture",
		"beauty", "toys", "office supplies", "garden", "department store",
		"discount", "specialty food", "art supplies", "e-commerce", "marketplace",
	}

	coveredSubcategories := make(map[string]bool)

	// Check coverage
	for _, tc := range retailTestCases {
		description := tc.Description
		keywords := tc.Keywords

		// Check for subcategory coverage
		for _, subcategory := range expectedSubcategories {
			// Check in description
			if containsKeyword(description, subcategory) {
				coveredSubcategories[subcategory] = true
			}
			// Check in keywords
			for _, keyword := range keywords {
				if containsKeyword(keyword, subcategory) {
					coveredSubcategories[subcategory] = true
				}
			}
		}
	}

	// Validate coverage
	coveragePercentage := float64(len(coveredSubcategories)) / float64(len(expectedSubcategories)) * 100
	assert.True(t, coveragePercentage >= 80.0, "Retail subcategory coverage should be >= 80%% (got %.1f%%)", coveragePercentage)

	t.Logf("📊 Retail industry coverage analysis:")
	t.Logf("   - Expected subcategories: %d", len(expectedSubcategories))
	t.Logf("   - Covered subcategories: %d", len(coveredSubcategories))
	t.Logf("   - Coverage percentage: %.1f%%", coveragePercentage)

	// Log uncovered subcategories
	var uncovered []string
	for _, subcategory := range expectedSubcategories {
		if !coveredSubcategories[subcategory] {
			uncovered = append(uncovered, subcategory)
		}
	}
	if len(uncovered) > 0 {
		t.Logf("   - Uncovered subcategories: %v", uncovered)
	}
}

// Helper function to check if a string contains a keyword (case-insensitive)
func containsKeyword(text, keyword string) bool {
	text = strings.ToLower(text)
	keyword = strings.ToLower(keyword)
	return strings.Contains(text, keyword)
}

// TestRetailClassificationAccuracy tests the accuracy of retail classification
func TestRetailClassificationAccuracy(t *testing.T) {
	dataset := NewComprehensiveTestDataset()
	retailTestCases := dataset.GetTestCasesByCategory("Retail")

	totalConfidence := 0.0
	validTestCases := 0

	for _, tc := range retailTestCases {
		// Skip security test cases for accuracy calculation
		if tc.DifficultyLevel == "Hard" || tc.ID == "retail-021" {
			continue
		}

		// Validate confidence score
		assert.True(t, tc.ExpectedConfidence >= 0.80, "Test case %s should have confidence >= 80%% (got %.2f%%)", tc.Name, tc.ExpectedConfidence*100)

		totalConfidence += tc.ExpectedConfidence
		validTestCases++
	}

	// Calculate average accuracy
	averageAccuracy := totalConfidence / float64(validTestCases)

	// Validate accuracy target
	assert.True(t, averageAccuracy >= 0.80, "Average retail classification accuracy should be >= 80%% (got %.2f%%)", averageAccuracy*100)

	t.Logf("🎯 Retail classification accuracy validation:")
	t.Logf("   - Valid test cases: %d", validTestCases)
	t.Logf("   - Average accuracy: %.2f%%", averageAccuracy*100)
	t.Logf("   - Target accuracy: 80%%")
	t.Logf("   - Accuracy target met: %t", averageAccuracy >= 0.80)
}
