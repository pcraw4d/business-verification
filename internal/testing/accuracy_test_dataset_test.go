package testing

import (
	"context"
	"testing"
	"time"
)

// TestAccuracyTestDataset_LoadAllTestCases tests loading all test cases
func TestAccuracyTestDataset_LoadAllTestCases(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This would require a test database setup
	// For now, we'll test the validation and structure
	t.Run("validate test case structure", func(t *testing.T) {
		tc := &TestCase{
			BusinessName:              "Test Business",
			BusinessDescription:       "Test Description",
			WebsiteURL:                "https://test.com",
			ExpectedPrimaryIndustry:   "Technology",
			ExpectedIndustryConfidence: 0.95,
			ExpectedMCCCodes:          []string{"5734", "5045"},
			ExpectedNAICSCodes:        []string{"541511"},
			ExpectedSICCodes:          []string{"7372"},
			TestCategory:              "Technology",
			TestSubcategory:           "Software Development",
			IsEdgeCase:                false,
			IsHighConfidence:          true,
			ExpectedConfidenceMin:     0.80,
			BusinessSize:              "large",
			BusinessType:              "corporation",
			LocationCountry:           "US",
			IsActive:                  true,
		}

		// Create a mock dataset manager (without DB)
		atd := &AccuracyTestDataset{
			logger: nil,
		}

		err := atd.ValidateTestCase(tc)
		if err != nil {
			t.Errorf("expected no validation error, got: %v", err)
		}
	})
}

// TestAccuracyTestDataset_ValidateTestCase tests test case validation
func TestAccuracyTestDataset_ValidateTestCase(t *testing.T) {
	atd := &AccuracyTestDataset{logger: nil}

	tests := []struct {
		name    string
		testCase *TestCase
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid test case",
			testCase: &TestCase{
				BusinessName:              "Valid Business",
				TestCategory:              "Technology",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: 0.95,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: false,
		},
		{
			name: "missing business name",
			testCase: &TestCase{
				TestCategory:              "Technology",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: 0.95,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: true,
			errMsg:  "business_name is required",
		},
		{
			name: "missing test category",
			testCase: &TestCase{
				BusinessName:              "Test Business",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: 0.95,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: true,
			errMsg:  "test_category is required",
		},
		{
			name: "missing expected industry",
			testCase: &TestCase{
				BusinessName:              "Test Business",
				TestCategory:              "Technology",
				ExpectedIndustryConfidence: 0.95,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: true,
			errMsg:  "expected_primary_industry is required",
		},
		{
			name: "missing all expected codes",
			testCase: &TestCase{
				BusinessName:              "Test Business",
				TestCategory:              "Technology",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: 0.95,
			},
			wantErr: true,
			errMsg:  "at least one expected code type",
		},
		{
			name: "invalid confidence too high",
			testCase: &TestCase{
				BusinessName:              "Test Business",
				TestCategory:              "Technology",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: 1.5,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: true,
			errMsg:  "expected_industry_confidence must be between",
		},
		{
			name: "invalid confidence too low",
			testCase: &TestCase{
				BusinessName:              "Test Business",
				TestCategory:              "Technology",
				ExpectedPrimaryIndustry:   "Technology",
				ExpectedIndustryConfidence: -0.1,
				ExpectedMCCCodes:          []string{"5734"},
			},
			wantErr: true,
			errMsg:  "expected_industry_confidence must be between",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := atd.ValidateTestCase(tt.testCase)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestAccuracyTestDataset_ExportTestCases tests exporting test cases to JSON
func TestAccuracyTestDataset_ExportTestCases(t *testing.T) {
	atd := &AccuracyTestDataset{logger: nil}

	testCases := []*TestCase{
		{
			ID:                        1,
			BusinessName:              "Test Business 1",
			ExpectedPrimaryIndustry:   "Technology",
			ExpectedIndustryConfidence: 0.95,
			ExpectedMCCCodes:          []string{"5734"},
			TestCategory:              "Technology",
			IsActive:                  true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		},
		{
			ID:                        2,
			BusinessName:              "Test Business 2",
			ExpectedPrimaryIndustry:   "Healthcare",
			ExpectedIndustryConfidence: 0.90,
			ExpectedNAICSCodes:        []string{"622110"},
			TestCategory:              "Healthcare",
			IsActive:                  true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		},
	}

	ctx := context.Background()
	jsonData, err := atd.ExportTestCases(ctx, testCases)
	if err != nil {
		t.Fatalf("failed to export test cases: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON data")
	}

	// Verify JSON is valid by checking it contains expected fields
	jsonStr := string(jsonData)
	if !containsSubstring(jsonStr, "Test Business 1") {
		t.Error("expected JSON to contain test case data")
	}
	if !containsSubstring(jsonStr, "total_cases") {
		t.Error("expected JSON to contain total_cases field")
	}
	if !containsSubstring(jsonStr, "exported_at") {
		t.Error("expected JSON to contain exported_at field")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestDatasetStatistics_Structure tests the DatasetStatistics structure
func TestDatasetStatistics_Structure(t *testing.T) {
	stats := &DatasetStatistics{
		TotalTestCases:      1000,
		CategoryCounts:      map[string]int{"Technology": 200, "Healthcare": 150},
		IndustryCounts:      map[string]int{"Technology": 200, "Healthcare": 150},
		EdgeCaseCount:       50,
		HighConfidenceCount: 800,
		VerifiedCount:       900,
	}

	if stats.TotalTestCases != 1000 {
		t.Errorf("expected TotalTestCases to be 1000, got: %d", stats.TotalTestCases)
	}
	if len(stats.CategoryCounts) != 2 {
		t.Errorf("expected 2 categories, got: %d", len(stats.CategoryCounts))
	}
	if stats.EdgeCaseCount != 50 {
		t.Errorf("expected EdgeCaseCount to be 50, got: %d", stats.EdgeCaseCount)
	}
}

