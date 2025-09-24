package test

import (
	"context"
	"log"
	"testing"
)

// TestClassificationAccuracyComprehensive runs comprehensive classification accuracy tests
func TestClassificationAccuracyComprehensive(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run all tests
	testRunner.RunAllTests(t)
}

// TestClassificationAccuracyBasic tests basic classification accuracy
func TestClassificationAccuracyBasic(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run basic accuracy test
	testRunner.RunBasicAccuracyTest(t)
}

// TestClassificationAccuracyByIndustry tests classification accuracy by industry
func TestClassificationAccuracyByIndustry(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run industry-specific test
	testRunner.RunIndustrySpecificTest(t)
}

// TestClassificationAccuracyByDifficulty tests classification accuracy by difficulty
func TestClassificationAccuracyByDifficulty(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run difficulty-based test
	testRunner.RunDifficultyBasedTest(t)
}

// TestClassificationAccuracyEdgeCases tests edge case handling
func TestClassificationAccuracyEdgeCases(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run edge case test
	testRunner.RunEdgeCaseTest(t)
}

// TestClassificationAccuracyPerformance tests performance and response times
func TestClassificationAccuracyPerformance(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run performance test
	testRunner.RunPerformanceTest(t)
}

// TestClassificationAccuracyConfidence tests confidence score validation
func TestClassificationAccuracyConfidence(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run confidence validation test
	testRunner.RunConfidenceValidationTest(t)
}

// TestClassificationAccuracyCodeMapping tests industry code mapping accuracy
func TestClassificationAccuracyCodeMapping(t *testing.T) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Run code mapping test
	testRunner.RunCodeMappingTest(t)
}

// TestClassificationAccuracyDatasetValidation tests the test dataset itself
func TestClassificationAccuracyDatasetValidation(t *testing.T) {
	// Create test dataset
	dataset := NewComprehensiveTestDataset()

	// Validate dataset statistics
	stats := dataset.GetStatistics()

	// Check basic statistics
	if stats["total_test_cases"].(int) < 20 {
		t.Errorf("❌ Dataset should have at least 20 test cases, got %d", stats["total_test_cases"])
	}

	// Check category distribution
	categories := stats["categories"].(map[string]int)
	expectedCategories := []string{"Technology", "Healthcare", "Finance", "Retail", "Manufacturing", "Professional Services", "Real Estate", "Education", "Energy", "Edge Cases"}

	for _, category := range expectedCategories {
		if categories[category] == 0 {
			t.Errorf("❌ Dataset missing test cases for category: %s", category)
		}
	}

	// Check difficulty distribution
	difficulties := stats["difficulties"].(map[string]int)
	expectedDifficulties := []string{"Easy", "Medium", "Hard"}

	for _, difficulty := range expectedDifficulties {
		if difficulties[difficulty] == 0 {
			t.Errorf("❌ Dataset missing test cases for difficulty: %s", difficulty)
		}
	}

	// Check industry distribution
	industries := stats["industries"].(map[string]int)
	if len(industries) < 8 {
		t.Errorf("❌ Dataset should cover at least 8 different industries, got %d", len(industries))
	}

	// Check average confidence
	avgConfidence := stats["average_confidence"].(float64)
	if avgConfidence < 0.5 || avgConfidence > 1.0 {
		t.Errorf("❌ Dataset average confidence should be between 0.5 and 1.0, got %.2f", avgConfidence)
	}

	t.Logf("✅ Dataset validation passed:")
	t.Logf("   Total test cases: %d", stats["total_test_cases"])
	t.Logf("   Categories: %d", len(categories))
	t.Logf("   Industries: %d", len(industries))
	t.Logf("   Average confidence: %.2f", avgConfidence)
}

// TestClassificationAccuracyWithRealRepository tests with real repository (if available)
func TestClassificationAccuracyWithRealRepository(t *testing.T) {
	// Skip if no real repository is available
	t.Skip("Skipping real repository test - requires database connection")

	// TODO: Implement real repository test when database is available
	// This would test against the actual Supabase database
}

// BenchmarkClassificationAccuracy benchmarks classification accuracy performance
func BenchmarkClassificationAccuracy(b *testing.B) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Create test dataset
	dataset := NewComprehensiveTestDataset()

	// Use first test case for benchmarking
	testCase := dataset.TestCases[0]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := testRunner.GetClassifier().GenerateClassificationCodes(
			context.Background(),
			testCase.Keywords,
			testCase.BusinessName,
			testCase.ExpectedConfidence,
		)

		if err != nil {
			b.Errorf("Classification failed: %v", err)
		}
	}
}

// BenchmarkClassificationAccuracyByIndustry benchmarks performance by industry
func BenchmarkClassificationAccuracyByIndustry(b *testing.B) {
	// Create test runner with mock repository
	testRunner := NewClassificationAccuracyTestRunnerWithMock(log.Default())

	// Create test dataset
	dataset := NewComprehensiveTestDataset()

	// Test different industries
	industries := []string{"Technology", "Healthcare", "Finance", "Retail", "Manufacturing"}

	for _, industry := range industries {
		testCases := dataset.GetTestCasesByIndustry(industry)
		if len(testCases) == 0 {
			continue
		}

		testCase := testCases[0]

		b.Run(industry, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := testRunner.GetClassifier().GenerateClassificationCodes(
					context.Background(),
					testCase.Keywords,
					testCase.BusinessName,
					testCase.ExpectedConfidence,
				)

				if err != nil {
					b.Errorf("Classification failed for %s: %v", industry, err)
				}
			}
		})
	}
}
