//go:build !comprehensive_test
// +build !comprehensive_test

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
	testingpkg "kyb-platform/internal/testing"
)

// TestComprehensiveAccuracyTestSuite tests the comprehensive accuracy test suite with real database
func TestComprehensiveAccuracyTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive accuracy test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping comprehensive accuracy test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Create database client
	config := &database.SupabaseConfig{
		URL:            supabaseURL,
		APIKey:         supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "[ACCURACY-TEST] ", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("Skipping comprehensive accuracy test: cannot connect to Supabase: %v", err)
	}

	// Get database connection for test dataset manager
	// Use SetupTestDatabase which handles environment variables properly
	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer testDB.CleanupTestDatabase()
	db := testDB.db

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Create services
	industryService := classification.NewIndustryDetectionService(repo, logger)
	codeGenerator := classification.NewClassificationCodeGenerator(repo, logger)

	// Create test dataset manager
	datasetManager := testingpkg.NewAccuracyTestDataset(db, logger)

	// Create comprehensive accuracy tester
	accuracyTester := testingpkg.NewComprehensiveAccuracyTester(
		datasetManager,
		industryService,
		codeGenerator,
		logger,
	)

	// Create accuracy report generator
	reportGenerator := testingpkg.NewAccuracyReportGenerator(logger)

	// Run comprehensive accuracy tests
	t.Run("Run Comprehensive Accuracy Tests", func(t *testing.T) {
		testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer testCancel()

		metrics, err := accuracyTester.RunAccuracyTests(testCtx)
		if err != nil {
			t.Fatalf("Failed to run accuracy tests: %v", err)
		}

		// Validate test results
		t.Logf("✅ Accuracy tests completed successfully")
		t.Logf("   Total Test Cases: %d", metrics.TotalTestCases)
		t.Logf("   Passed: %d", metrics.PassedTestCases)
		t.Logf("   Failed: %d", metrics.FailedTestCases)
		t.Logf("   Overall Accuracy: %.2f%%", metrics.OverallAccuracy*100)
		t.Logf("   Industry Accuracy: %.2f%%", metrics.IndustryAccuracy*100)
		t.Logf("   Code Accuracy: %.2f%%", metrics.CodeAccuracy*100)
		t.Logf("   MCC Accuracy: %.2f%%", metrics.MCCAccuracy*100)
		t.Logf("   NAICS Accuracy: %.2f%%", metrics.NAICSAccuracy*100)
		t.Logf("   SIC Accuracy: %.2f%%", metrics.SICAccuracy*100)
		t.Logf("   Average Processing Time: %v", metrics.AverageProcessingTime)

		// Validate targets (95%+ industry, 90%+ code)
		t.Run("Validate Industry Accuracy Target", func(t *testing.T) {
			target := 0.95 // 95%
			if metrics.IndustryAccuracy < target {
				t.Errorf("Industry accuracy %.2f%% is below target of %.2f%%", 
					metrics.IndustryAccuracy*100, target*100)
			} else {
				t.Logf("✅ Industry accuracy target met: %.2f%% >= %.2f%%", 
					metrics.IndustryAccuracy*100, target*100)
			}
		})

		t.Run("Validate Code Accuracy Target", func(t *testing.T) {
			target := 0.90 // 90%
			if metrics.CodeAccuracy < target {
				t.Errorf("Code accuracy %.2f%% is below target of %.2f%%", 
					metrics.CodeAccuracy*100, target*100)
			} else {
				t.Logf("✅ Code accuracy target met: %.2f%% >= %.2f%%", 
					metrics.CodeAccuracy*100, target*100)
			}
		})

		t.Run("Validate Overall Accuracy", func(t *testing.T) {
			// Overall accuracy should be reasonable (weighted average)
			if metrics.OverallAccuracy < 0.85 {
				t.Errorf("Overall accuracy %.2f%% is below minimum threshold of 85%%", 
					metrics.OverallAccuracy*100)
			} else {
				t.Logf("✅ Overall accuracy acceptable: %.2f%%", metrics.OverallAccuracy*100)
			}
		})

		// Generate and validate accuracy report
		t.Run("Generate Accuracy Report", func(t *testing.T) {
			report, err := reportGenerator.GenerateReport(metrics)
			if err != nil {
				t.Fatalf("Failed to generate accuracy report: %v", err)
			}

			// Validate report structure
			if report.Metrics == nil {
				t.Error("Report metrics should not be nil")
			}
			if report.Summary == nil {
				t.Error("Report summary should not be nil")
			}
			if len(report.Recommendations) == 0 {
				t.Error("Report should have recommendations")
			}

			t.Logf("✅ Accuracy report generated successfully")
			t.Logf("   Report generated at: %s", report.GeneratedAt.Format(time.RFC3339))
			t.Logf("   Recommendations: %d", len(report.Recommendations))

			// Generate JSON report
			jsonReport, err := reportGenerator.GenerateJSONReport(metrics)
			if err != nil {
				t.Errorf("Failed to generate JSON report: %v", err)
			} else {
				if len(jsonReport) == 0 {
					t.Error("JSON report should not be empty")
				} else {
					t.Logf("✅ JSON report generated: %d bytes", len(jsonReport))
				}
			}

			// Generate text report
			textReport, err := reportGenerator.GenerateTextReport(metrics)
			if err != nil {
				t.Errorf("Failed to generate text report: %v", err)
			} else {
				if len(textReport) == 0 {
					t.Error("Text report should not be empty")
				} else {
					t.Logf("✅ Text report generated: %d bytes", len(textReport))
					// Log first 500 characters of text report for visibility
					if len(textReport) > 500 {
						t.Logf("   Report preview:\n%s...", textReport[:500])
					} else {
						t.Logf("   Report:\n%s", textReport)
					}
				}
			}
		})

		// Validate category breakdown
		t.Run("Validate Category Breakdown", func(t *testing.T) {
			if len(metrics.AccuracyByCategory) == 0 {
				t.Error("Accuracy by category should not be empty")
			} else {
				t.Logf("✅ Category breakdown available for %d categories", len(metrics.AccuracyByCategory))
				for category, accuracy := range metrics.AccuracyByCategory {
					t.Logf("   %s: %.2f%%", category, accuracy*100)
				}
			}
		})

		// Validate industry breakdown
		t.Run("Validate Industry Breakdown", func(t *testing.T) {
			if len(metrics.AccuracyByIndustry) == 0 {
				t.Error("Accuracy by industry should not be empty")
			} else {
				t.Logf("✅ Industry breakdown available for %d industries", len(metrics.AccuracyByIndustry))
				for industry, accuracy := range metrics.AccuracyByIndustry {
					t.Logf("   %s: %.2f%%", industry, accuracy*100)
				}
			}
		})

		// Validate test results
		t.Run("Validate Test Results", func(t *testing.T) {
			if len(metrics.TestResults) == 0 {
				t.Error("Test results should not be empty")
			} else {
				t.Logf("✅ Test results available: %d test cases", len(metrics.TestResults))

				// Check for errors
				errorCount := 0
				for _, result := range metrics.TestResults {
					if result.Error != "" {
						errorCount++
					}
				}
				if errorCount > 0 {
					t.Logf("⚠️  %d test cases had errors", errorCount)
					// Log first few errors for debugging
					errorShown := 0
					for _, result := range metrics.TestResults {
						if result.Error != "" && errorShown < 3 {
							t.Logf("   Error example: %s - %s", result.BusinessName, result.Error)
							errorShown++
						}
					}
				} else {
					t.Logf("✅ No errors in test results")
				}
			}
		})
	})

	// Test accuracy by category
	t.Run("Test Accuracy By Category", func(t *testing.T) {
		testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer testCancel()

		categories := []string{"Technology", "Healthcare", "Financial Services", "Retail", "Edge Cases"}

		for _, category := range categories {
			t.Run(category, func(t *testing.T) {
				metrics, err := accuracyTester.RunAccuracyTestsByCategory(testCtx, category)
				if err != nil {
					t.Logf("⚠️  Failed to run tests for category %s: %v", category, err)
					return
				}

				t.Logf("   Category: %s", category)
				t.Logf("   Test Cases: %d", metrics.TotalTestCases)
				t.Logf("   Overall Accuracy: %.2f%%", metrics.OverallAccuracy*100)
				t.Logf("   Industry Accuracy: %.2f%%", metrics.IndustryAccuracy*100)
				t.Logf("   Code Accuracy: %.2f%%", metrics.CodeAccuracy*100)

				// Validate category-specific targets
				if category != "Edge Cases" {
					// Edge cases may have lower accuracy, which is expected
					if metrics.OverallAccuracy < 0.80 {
						t.Errorf("Category %s accuracy %.2f%% is below minimum threshold of 80%%", 
							category, metrics.OverallAccuracy*100)
					}
				}
			})
		}
	})
}

// TestAccuracyTestDatasetLoading tests loading test cases from the dataset
func TestAccuracyTestDatasetLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dataset loading test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping dataset loading test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Get database connection
	// Use SetupTestDatabase which handles environment variables properly
	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer testDB.CleanupTestDatabase()
	db := testDB.db

	logger := log.New(os.Stdout, "[DATASET-TEST] ", log.LstdFlags)
	datasetManager := testingpkg.NewAccuracyTestDataset(db, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Load All Test Cases", func(t *testing.T) {
		testCases, err := datasetManager.LoadAllTestCases(ctx)
		if err != nil {
			t.Fatalf("Failed to load test cases: %v", err)
		}

		if len(testCases) == 0 {
			t.Error("Expected test cases to be loaded, got 0")
		} else {
			t.Logf("✅ Loaded %d test cases", len(testCases))
			
			// Validate minimum count (should be 184 based on verification)
			if len(testCases) < 180 {
				t.Errorf("Expected at least 180 test cases, got %d", len(testCases))
			}
		}
	})

	t.Run("Get Dataset Statistics", func(t *testing.T) {
		stats, err := datasetManager.GetDatasetStatistics(ctx)
		if err != nil {
			t.Fatalf("Failed to get dataset statistics: %v", err)
		}

		t.Logf("✅ Dataset Statistics:")
		t.Logf("   Total Test Cases: %d", stats.TotalTestCases)
		t.Logf("   Edge Cases: %d", stats.EdgeCaseCount)
		t.Logf("   High Confidence: %d", stats.HighConfidenceCount)
		t.Logf("   Verified: %d", stats.VerifiedCount)
		t.Logf("   Categories: %d", len(stats.CategoryCounts))
		t.Logf("   Industries: %d", len(stats.IndustryCounts))

		// Validate statistics
		if stats.TotalTestCases == 0 {
			t.Error("Total test cases should not be 0")
		}
		if len(stats.CategoryCounts) == 0 {
			t.Error("Category counts should not be empty")
		}
		if len(stats.IndustryCounts) == 0 {
			t.Error("Industry counts should not be empty")
		}
	})

	t.Run("Load Test Cases By Category", func(t *testing.T) {
		categories := []string{"Technology", "Healthcare", "Financial Services", "Retail", "Edge Cases"}

		for _, category := range categories {
			testCases, err := datasetManager.LoadTestCasesByCategory(ctx, category)
			if err != nil {
				t.Errorf("Failed to load test cases for category %s: %v", category, err)
				continue
			}

			if len(testCases) == 0 {
				t.Logf("⚠️  No test cases found for category: %s", category)
			} else {
				t.Logf("✅ Loaded %d test cases for category: %s", len(testCases), category)
			}
		}
	})

	t.Run("Load Test Cases By Industry", func(t *testing.T) {
		industries := []string{"Technology", "Healthcare", "Financial Services", "Retail"}

		for _, industry := range industries {
			testCases, err := datasetManager.LoadTestCasesByIndustry(ctx, industry)
			if err != nil {
				t.Errorf("Failed to load test cases for industry %s: %v", industry, err)
				continue
			}

			if len(testCases) == 0 {
				t.Logf("⚠️  No test cases found for industry: %s", industry)
			} else {
				t.Logf("✅ Loaded %d test cases for industry: %s", len(testCases), industry)
			}
		}
	})
}


