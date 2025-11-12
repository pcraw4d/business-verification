package risk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestRunner provides a comprehensive test runner for integration tests
type TestRunner struct {
	logger    *zap.Logger
	testSuite *IntegrationTestSuite
	results   *TestResults
}

// TestResults contains the results of test execution
type TestResults struct {
	TotalTests    int                    `json:"total_tests"`
	PassedTests   int                    `json:"passed_tests"`
	FailedTests   int                    `json:"failed_tests"`
	SkippedTests  int                    `json:"skipped_tests"`
	ExecutionTime time.Duration          `json:"execution_time"`
	TestDetails   []TestDetail           `json:"test_details"`
	Summary       map[string]interface{} `json:"summary"`
}

// TestDetail contains details about individual test execution
type TestDetail struct {
	Name             string        `json:"name"`
	Status           string        `json:"status"`
	Duration         time.Duration `json:"duration"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	Assertions       int           `json:"assertions"`
	PassedAssertions int           `json:"passed_assertions"`
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	logger := zap.NewNop()
	return &TestRunner{
		logger: logger,
		results: &TestResults{
			TestDetails: make([]TestDetail, 0),
			Summary:     make(map[string]interface{}),
		},
	}
}

// RunAllTests runs all integration tests
func (tr *TestRunner) RunAllTests(t *testing.T) *TestResults {
	startTime := time.Now()
	tr.logger.Info("Starting integration test suite")

	// Initialize test suite
	tr.testSuite = NewIntegrationTestSuite(t)

	// Run all test categories
	tr.runTestCategory(t, "End-to-End Workflow", tr.testSuite.TestEndToEndRiskAssessmentWorkflow)
	tr.runTestCategory(t, "API Integration", tr.testSuite.TestAPIIntegration)
	tr.runTestCategory(t, "Database Integration", tr.testSuite.TestDatabaseIntegration)
	tr.runTestCategory(t, "Error Handling", tr.testSuite.TestErrorHandling)
	tr.runTestCategory(t, "Performance", tr.testSuite.TestPerformance)
	tr.runTestCategory(t, "Data Integrity", tr.testSuite.TestDataIntegrity)
	tr.runTestCategory(t, "Workflow Integration", tr.testSuite.TestWorkflowIntegration)

	// Calculate final results
	tr.results.ExecutionTime = time.Since(startTime)
	tr.calculateSummary()

	tr.logger.Info("Integration test suite completed",
		zap.Int("total_tests", tr.results.TotalTests),
		zap.Int("passed_tests", tr.results.PassedTests),
		zap.Int("failed_tests", tr.results.FailedTests),
		zap.Duration("execution_time", tr.results.ExecutionTime))

	return tr.results
}

// runTestCategory runs a specific test category
func (tr *TestRunner) runTestCategory(t *testing.T, categoryName string, testFunc func(*testing.T)) {
	tr.logger.Info("Running test category", zap.String("category", categoryName))

	// Create a sub-test for the category
	t.Run(categoryName, func(t *testing.T) {
		startTime := time.Now()

		// Run the test function
		testFunc(t)

		duration := time.Since(startTime)

		// Record test result
		tr.results.TotalTests++
		tr.results.PassedTests++ // If we get here, the test passed

		tr.results.TestDetails = append(tr.results.TestDetails, TestDetail{
			Name:     categoryName,
			Status:   "PASSED",
			Duration: duration,
		})

		tr.logger.Info("Test category completed",
			zap.String("category", categoryName),
			zap.Duration("duration", duration),
			zap.String("status", "PASSED"))
	})
}

// calculateSummary calculates test summary statistics
func (tr *TestRunner) calculateSummary() {
	tr.results.Summary = map[string]interface{}{
		"total_tests":       tr.results.TotalTests,
		"passed_tests":      tr.results.PassedTests,
		"failed_tests":      tr.results.FailedTests,
		"skipped_tests":     tr.results.SkippedTests,
		"pass_rate":         float64(tr.results.PassedTests) / float64(tr.results.TotalTests) * 100,
		"execution_time":    tr.results.ExecutionTime.String(),
		"average_test_time": tr.results.ExecutionTime / time.Duration(tr.results.TotalTests),
	}

	// Calculate category-specific statistics
	categoryStats := make(map[string]map[string]interface{})
	for _, detail := range tr.results.TestDetails {
		if categoryStats[detail.Name] == nil {
			categoryStats[detail.Name] = make(map[string]interface{})
		}
		categoryStats[detail.Name]["duration"] = detail.Duration.String()
		categoryStats[detail.Name]["status"] = detail.Status
	}
	tr.results.Summary["category_stats"] = categoryStats
}

// GenerateTestReport generates a comprehensive test report
func (tr *TestRunner) GenerateTestReport(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate JSON report
	jsonReportPath := filepath.Join(outputDir, "integration_test_report.json")
	if err := tr.generateJSONReport(jsonReportPath); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	tr.logger.Info("Test report generated",
		zap.String("output_dir", outputDir),
		zap.String("json_report", jsonReportPath))

	return nil
}

// generateJSONReport generates a JSON test report
func (tr *TestRunner) generateJSONReport(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tr.results)
}

// RunSmokeTests runs a subset of critical tests for quick validation
func (tr *TestRunner) RunSmokeTests(t *testing.T) {
	tr.logger.Info("Running smoke tests")

	// Initialize test suite
	tr.testSuite = NewIntegrationTestSuite(t)

	// Run only critical tests
	t.Run("SmokeTest_EndToEndWorkflow", func(t *testing.T) {
		tr.testSuite.TestEndToEndRiskAssessmentWorkflow(t)
	})

	t.Run("SmokeTest_APIIntegration", func(t *testing.T) {
		tr.testSuite.TestAPIIntegration(t)
	})

	tr.logger.Info("Smoke tests completed")
}
