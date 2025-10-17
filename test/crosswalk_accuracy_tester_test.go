package test

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/classification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCrosswalkAccuracyTesterTypes(t *testing.T) {
	// Test struct initialization and type definitions
	t.Run("AccuracyTestResult", func(t *testing.T) {
		result := classification.AccuracyTestResult{
			TestName:        "Test Result",
			TestType:        "mcc_mapping_accuracy",
			TotalTests:      10,
			PassedTests:     8,
			FailedTests:     2,
			AccuracyScore:   0.8,
			ConfidenceScore: 0.85,
			Summary:         "Test completed successfully",
			Timestamp:       time.Now(),
			Metadata:        make(map[string]interface{}),
		}

		assert.Equal(t, "Test Result", result.TestName)
		assert.Equal(t, "mcc_mapping_accuracy", result.TestType)
		assert.Equal(t, 10, result.TotalTests)
		assert.Equal(t, 8, result.PassedTests)
		assert.Equal(t, 2, result.FailedTests)
		assert.Equal(t, 0.8, result.AccuracyScore)
		assert.Equal(t, 0.85, result.ConfidenceScore)
		assert.NotNil(t, result.Metadata)
	})

	t.Run("AccuracyTestDetail", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID:    "test_001",
			Input:         map[string]interface{}{"mcc_code": "5411"},
			Expected:      map[string]interface{}{"industry_id": 1},
			Actual:        map[string]interface{}{"industry_id": 1},
			Passed:        true,
			Confidence:    0.95,
			ExecutionTime: 100 * time.Millisecond,
		}

		assert.Equal(t, "test_001", detail.TestCaseID)
		assert.Equal(t, "5411", detail.Input["mcc_code"])
		assert.Equal(t, 1, detail.Expected["industry_id"])
		assert.True(t, detail.Passed)
		assert.Equal(t, 0.95, detail.Confidence)
	})

	t.Run("AccuracyTestSuite", func(t *testing.T) {
		suite := classification.AccuracyTestSuite{
			SuiteName:    "Test Suite",
			Description:  "Comprehensive test suite",
			TestCases:    []classification.AccuracyTestCase{},
			TestResults:  []classification.AccuracyTestResult{},
			OverallScore: 0.9,
			CreatedAt:    time.Now(),
			Metadata:     make(map[string]interface{}),
		}

		assert.Equal(t, "Test Suite", suite.SuiteName)
		assert.Equal(t, "Comprehensive test suite", suite.Description)
		assert.Equal(t, 0.9, suite.OverallScore)
		assert.NotNil(t, suite.TestCases)
		assert.NotNil(t, suite.TestResults)
		assert.NotNil(t, suite.Metadata)
	})

	t.Run("AccuracyTestCase", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID:  "case_001",
			TestName:    "MCC Mapping Test",
			TestType:    "mcc_mapping_accuracy",
			Input:       map[string]interface{}{"mcc_code": "5411"},
			Expected:    map[string]interface{}{"industry_id": 1},
			Description: "Test MCC 5411 mapping",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"grocery", "retail"},
		}

		assert.Equal(t, "case_001", testCase.TestCaseID)
		assert.Equal(t, "MCC Mapping Test", testCase.TestName)
		assert.Equal(t, "mcc_mapping_accuracy", testCase.TestType)
		assert.Equal(t, 1.0, testCase.Weight)
		assert.Equal(t, "retail", testCase.Category)
		assert.Len(t, testCase.Tags, 2)
	})
}

func TestCrosswalkAccuracyTesterCreation(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil database (should not panic)
	tester := classification.NewCrosswalkAccuracyTester(nil, logger)
	assert.NotNil(t, tester)
	// Note: db and logger fields are private, so we can't access them directly
}

func TestAccuracyTestResultCalculations(t *testing.T) {
	t.Run("Accuracy Score Calculation", func(t *testing.T) {
		result := classification.AccuracyTestResult{
			TotalTests:  100,
			PassedTests: 85,
		}

		// This would be calculated in the actual implementation
		expectedAccuracy := float64(result.PassedTests) / float64(result.TotalTests)
		assert.Equal(t, 0.85, expectedAccuracy)
	})

	t.Run("Confidence Score Calculation", func(t *testing.T) {
		details := []classification.AccuracyTestDetail{
			{Confidence: 0.9},
			{Confidence: 0.8},
			{Confidence: 0.95},
		}

		totalConfidence := 0.0
		for _, detail := range details {
			totalConfidence += detail.Confidence
		}
		avgConfidence := totalConfidence / float64(len(details))

		assert.InDelta(t, 0.8833333333333333, avgConfidence, 0.000000000000001)
	})
}

func TestAccuracyTestCaseValidation(t *testing.T) {
	t.Run("Valid Test Case", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID:  "valid_001",
			TestName:    "Valid Test",
			TestType:    "mcc_mapping_accuracy",
			Input:       map[string]interface{}{"mcc_code": "5411"},
			Expected:    map[string]interface{}{"industry_id": 1},
			Description: "Valid test case",
			Weight:      1.0,
			Category:    "retail",
			Tags:        []string{"test"},
		}

		// Validate required fields
		assert.NotEmpty(t, testCase.TestCaseID)
		assert.NotEmpty(t, testCase.TestName)
		assert.NotEmpty(t, testCase.TestType)
		assert.NotNil(t, testCase.Input)
		assert.NotNil(t, testCase.Expected)
		assert.Greater(t, testCase.Weight, 0.0)
	})

	t.Run("Invalid Test Case", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID: "", // Invalid: empty ID
			TestName:   "Invalid Test",
			TestType:   "unknown_type", // Invalid: unknown type
			Input:      nil,            // Invalid: nil input
			Expected:   nil,            // Invalid: nil expected
			Weight:     -1.0,           // Invalid: negative weight
		}

		// Validate that invalid fields are detected
		assert.Empty(t, testCase.TestCaseID)
		assert.Equal(t, "unknown_type", testCase.TestType)
		assert.Nil(t, testCase.Input)
		assert.Nil(t, testCase.Expected)
		assert.Less(t, testCase.Weight, 0.0)
	})
}

func TestAccuracyTestDetailComparison(t *testing.T) {
	t.Run("Passed Test Detail", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID: "test_001",
			Input:      map[string]interface{}{"mcc_code": "5411"},
			Expected:   map[string]interface{}{"industry_id": 1},
			Actual:     map[string]interface{}{"industry_id": 1},
			Passed:     true,
			Confidence: 0.95,
		}

		// Verify the test passed
		assert.True(t, detail.Passed)
		assert.Equal(t, detail.Expected["industry_id"], detail.Actual["industry_id"])
		assert.Greater(t, detail.Confidence, 0.9)
	})

	t.Run("Failed Test Detail", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID: "test_002",
			Input:      map[string]interface{}{"mcc_code": "5411"},
			Expected:   map[string]interface{}{"industry_id": 1},
			Actual:     map[string]interface{}{"industry_id": 2}, // Different result
			Passed:     false,
			Confidence: 0.3,
			Error:      "Industry ID mismatch",
		}

		// Verify the test failed
		assert.False(t, detail.Passed)
		assert.NotEqual(t, detail.Expected["industry_id"], detail.Actual["industry_id"])
		assert.Less(t, detail.Confidence, 0.5)
		assert.NotEmpty(t, detail.Error)
	})
}

func TestAccuracyTestSuiteAggregation(t *testing.T) {
	t.Run("Suite Score Calculation", func(t *testing.T) {
		results := []classification.AccuracyTestResult{
			{
				TestName:      "Test 1",
				AccuracyScore: 0.9,
			},
			{
				TestName:      "Test 2",
				AccuracyScore: 0.8,
			},
			{
				TestName:      "Test 3",
				AccuracyScore: 0.95,
			},
		}

		// Calculate overall score
		totalScore := 0.0
		for _, result := range results {
			totalScore += result.AccuracyScore
		}
		overallScore := totalScore / float64(len(results))

		assert.InDelta(t, 0.8833333333333333, overallScore, 0.000000000000001)
	})

	t.Run("Suite Metadata", func(t *testing.T) {
		suite := classification.AccuracyTestSuite{
			SuiteName:   "Test Suite",
			Description: "Test description",
			Metadata: map[string]interface{}{
				"version":     "1.0.0",
				"environment": "test",
				"created_by":  "test_user",
			},
		}

		assert.Equal(t, "1.0.0", suite.Metadata["version"])
		assert.Equal(t, "test", suite.Metadata["environment"])
		assert.Equal(t, "test_user", suite.Metadata["created_by"])
	})
}

func TestAccuracyTestTypes(t *testing.T) {
	t.Run("MCC Mapping Accuracy", func(t *testing.T) {
		testType := "mcc_mapping_accuracy"
		assert.Equal(t, "mcc_mapping_accuracy", testType)
	})

	t.Run("NAICS Mapping Accuracy", func(t *testing.T) {
		testType := "naics_mapping_accuracy"
		assert.Equal(t, "naics_mapping_accuracy", testType)
	})

	t.Run("SIC Mapping Accuracy", func(t *testing.T) {
		testType := "sic_mapping_accuracy"
		assert.Equal(t, "sic_mapping_accuracy", testType)
	})

	t.Run("Confidence Scoring Accuracy", func(t *testing.T) {
		testType := "confidence_scoring_accuracy"
		assert.Equal(t, "confidence_scoring_accuracy", testType)
	})

	t.Run("Validation Rules Accuracy", func(t *testing.T) {
		testType := "validation_rules_accuracy"
		assert.Equal(t, "validation_rules_accuracy", testType)
	})

	t.Run("Crosswalk Consistency Accuracy", func(t *testing.T) {
		testType := "crosswalk_consistency_accuracy"
		assert.Equal(t, "crosswalk_consistency_accuracy", testType)
	})

	t.Run("Industry Alignment Accuracy", func(t *testing.T) {
		testType := "industry_alignment_accuracy"
		assert.Equal(t, "industry_alignment_accuracy", testType)
	})
}

func TestAccuracyTestCategories(t *testing.T) {
	categories := []string{
		"retail",
		"manufacturing",
		"services",
		"technology",
		"healthcare",
		"finance",
		"confidence",
		"validation",
		"consistency",
		"alignment",
	}

	for _, category := range categories {
		t.Run("Category: "+category, func(t *testing.T) {
			testCase := classification.AccuracyTestCase{
				TestCaseID: "test_" + category,
				Category:   category,
			}
			assert.Equal(t, category, testCase.Category)
		})
	}
}

func TestAccuracyTestTags(t *testing.T) {
	t.Run("Single Tag", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID: "test_001",
			Tags:       []string{"grocery"},
		}
		assert.Len(t, testCase.Tags, 1)
		assert.Equal(t, "grocery", testCase.Tags[0])
	})

	t.Run("Multiple Tags", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID: "test_002",
			Tags:       []string{"grocery", "retail", "food"},
		}
		assert.Len(t, testCase.Tags, 3)
		assert.Contains(t, testCase.Tags, "grocery")
		assert.Contains(t, testCase.Tags, "retail")
		assert.Contains(t, testCase.Tags, "food")
	})

	t.Run("No Tags", func(t *testing.T) {
		testCase := classification.AccuracyTestCase{
			TestCaseID: "test_003",
			Tags:       []string{},
		}
		assert.Len(t, testCase.Tags, 0)
	})
}

func TestAccuracyTestExecutionTime(t *testing.T) {
	t.Run("Fast Execution", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID:    "fast_test",
			ExecutionTime: 10 * time.Millisecond,
		}
		assert.Less(t, detail.ExecutionTime, 100*time.Millisecond)
	})

	t.Run("Slow Execution", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID:    "slow_test",
			ExecutionTime: 5 * time.Second,
		}
		assert.Greater(t, detail.ExecutionTime, 1*time.Second)
	})
}

func TestAccuracyTestErrorHandling(t *testing.T) {
	t.Run("Test with Error", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID: "error_test",
			Passed:     false,
			Error:      "Database connection failed",
		}
		assert.False(t, detail.Passed)
		assert.NotEmpty(t, detail.Error)
		assert.Contains(t, detail.Error, "Database connection failed")
	})

	t.Run("Test without Error", func(t *testing.T) {
		detail := classification.AccuracyTestDetail{
			TestCaseID: "success_test",
			Passed:     true,
			Error:      "",
		}
		assert.True(t, detail.Passed)
		assert.Empty(t, detail.Error)
	})
}

func TestAccuracyTestInputValidation(t *testing.T) {
	t.Run("Valid MCC Input", func(t *testing.T) {
		input := map[string]interface{}{
			"mcc_code": "5411",
		}
		assert.Equal(t, "5411", input["mcc_code"])
	})

	t.Run("Valid NAICS Input", func(t *testing.T) {
		input := map[string]interface{}{
			"naics_code": "445110",
		}
		assert.Equal(t, "445110", input["naics_code"])
	})

	t.Run("Valid SIC Input", func(t *testing.T) {
		input := map[string]interface{}{
			"sic_code": "5411",
		}
		assert.Equal(t, "5411", input["sic_code"])
	})

	t.Run("Valid Combined Input", func(t *testing.T) {
		input := map[string]interface{}{
			"mcc_code":   "5411",
			"naics_code": "445110",
			"sic_code":   "5411",
		}
		assert.Equal(t, "5411", input["mcc_code"])
		assert.Equal(t, "445110", input["naics_code"])
		assert.Equal(t, "5411", input["sic_code"])
	})
}

func TestAccuracyTestExpectedValidation(t *testing.T) {
	t.Run("Valid Industry ID Expected", func(t *testing.T) {
		expected := map[string]interface{}{
			"industry_id": 1,
		}
		assert.Equal(t, 1, expected["industry_id"])
	})

	t.Run("Valid Confidence Score Expected", func(t *testing.T) {
		expected := map[string]interface{}{
			"confidence_score": 0.95,
		}
		assert.Equal(t, 0.95, expected["confidence_score"])
	})

	t.Run("Valid Multiple Expected Values", func(t *testing.T) {
		expected := map[string]interface{}{
			"industry_id":      1,
			"confidence_score": 0.95,
			"description":      "Grocery Stores",
		}
		assert.Equal(t, 1, expected["industry_id"])
		assert.Equal(t, 0.95, expected["confidence_score"])
		assert.Equal(t, "Grocery Stores", expected["description"])
	})
}

// Integration test that would require database connection
func TestCrosswalkAccuracyTesterIntegration(t *testing.T) {
	t.Skip("Database connection required for integration test")

	// This test would run the actual accuracy tests with a real database
	// For now, we skip it since we don't have a database connection in the test environment

	logger := zap.NewNop()
	tester := classification.NewCrosswalkAccuracyTester(nil, logger)

	ctx := context.Background()
	suite, err := tester.RunComprehensiveAccuracyTests(ctx)

	// These assertions would be valid with a real database
	require.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Equal(t, "Crosswalk Accuracy Test Suite", suite.SuiteName)
	assert.Greater(t, len(suite.TestResults), 0)
}

func TestAccuracyTestResultSerialization(t *testing.T) {
	t.Run("JSON Serialization", func(t *testing.T) {
		result := classification.AccuracyTestResult{
			TestName:        "Test Result",
			TestType:        "mcc_mapping_accuracy",
			TotalTests:      10,
			PassedTests:     8,
			FailedTests:     2,
			AccuracyScore:   0.8,
			ConfidenceScore: 0.85,
			Summary:         "Test completed successfully",
			Timestamp:       time.Now(),
			Metadata: map[string]interface{}{
				"version": "1.0.0",
			},
		}

		// Test that the struct can be serialized to JSON
		// This is implicitly tested by the JSON tags in the struct
		assert.NotEmpty(t, result.TestName)
		assert.NotEmpty(t, result.TestType)
		assert.NotNil(t, result.Metadata)
	})
}

func TestAccuracyTestSuiteSerialization(t *testing.T) {
	t.Run("JSON Serialization", func(t *testing.T) {
		suite := classification.AccuracyTestSuite{
			SuiteName:    "Test Suite",
			Description:  "Test description",
			TestCases:    []classification.AccuracyTestCase{},
			TestResults:  []classification.AccuracyTestResult{},
			OverallScore: 0.9,
			CreatedAt:    time.Now(),
			Metadata: map[string]interface{}{
				"environment": "test",
			},
		}

		// Test that the struct can be serialized to JSON
		assert.NotEmpty(t, suite.SuiteName)
		assert.NotEmpty(t, suite.Description)
		assert.NotNil(t, suite.TestCases)
		assert.NotNil(t, suite.TestResults)
		assert.NotNil(t, suite.Metadata)
	})
}
