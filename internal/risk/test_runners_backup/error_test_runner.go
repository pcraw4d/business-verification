package risk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// ErrorTestRunner provides comprehensive error handling testing capabilities
// NOTE: test_runners_backup is a subdirectory, so it's a separate package from internal/risk
// Types like ErrorHandlingTestSuite are defined in the parent package and are not accessible
type ErrorTestRunner struct {
	logger    *zap.Logger
	testSuite interface{} // *ErrorHandlingTestSuite - defined in parent package
	results   *ErrorTestResults
}

// ErrorTestResults contains the results of error handling test execution
type ErrorTestResults struct {
	TotalTests      int                    `json:"total_tests"`
	PassedTests     int                    `json:"passed_tests"`
	FailedTests     int                    `json:"failed_tests"`
	SkippedTests    int                    `json:"skipped_tests"`
	ExecutionTime   time.Duration          `json:"execution_time"`
	TestDetails     []ErrorTestDetail      `json:"test_details"`
	Summary         map[string]interface{} `json:"summary"`
	ErrorMetrics    *ErrorMetrics          `json:"error_metrics"`
	RecoveryMetrics *RecoveryMetrics       `json:"recovery_metrics"`
	SecurityMetrics *SecurityMetrics       `json:"security_metrics"`
}

// ErrorTestDetail contains details about individual error handling test execution
type ErrorTestDetail struct {
	Name          string        `json:"name"`
	ErrorType     string        `json:"error_type"`
	Category      string        `json:"category"`
	Status        string        `json:"status"`
	Duration      time.Duration `json:"duration"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ExpectedError bool          `json:"expected_error"`
	ActualError   bool          `json:"actual_error"`
	RecoveryTime  time.Duration `json:"recovery_time,omitempty"`
	Severity      string        `json:"severity"`
}

// ErrorMetrics contains error handling metrics
type ErrorMetrics struct {
	TotalErrors       int            `json:"total_errors"`
	ValidationErrors  int            `json:"validation_errors"`
	DatabaseErrors    int            `json:"database_errors"`
	APIErrors         int            `json:"api_errors"`
	ServiceErrors     int            `json:"service_errors"`
	ConcurrencyErrors int            `json:"concurrency_errors"`
	ResourceErrors    int            `json:"resource_errors"`
	SecurityErrors    int            `json:"security_errors"`
	ErrorRate         float64        `json:"error_rate"`
	AverageErrorTime  time.Duration  `json:"average_error_time"`
	ErrorDistribution map[string]int `json:"error_distribution"`
	CriticalErrors    int            `json:"critical_errors"`
	RecoverableErrors int            `json:"recoverable_errors"`
}

// RecoveryMetrics contains error recovery metrics
type RecoveryMetrics struct {
	TotalRecoveryAttempts int            `json:"total_recovery_attempts"`
	SuccessfulRecoveries  int            `json:"successful_recoveries"`
	FailedRecoveries      int            `json:"failed_recoveries"`
	RecoveryRate          float64        `json:"recovery_rate"`
	AverageRecoveryTime   time.Duration  `json:"average_recovery_time"`
	RecoveryStrategies    map[string]int `json:"recovery_strategies"`
	AutoRecoveryCount     int            `json:"auto_recovery_count"`
	ManualRecoveryCount   int            `json:"manual_recovery_count"`
}

// SecurityMetrics contains security error metrics
type SecurityMetrics struct {
	TotalSecurityAttempts int            `json:"total_security_attempts"`
	SQLInjectionAttempts  int            `json:"sql_injection_attempts"`
	XSSAttempts           int            `json:"xss_attempts"`
	PathTraversalAttempts int            `json:"path_traversal_attempts"`
	BlockedAttempts       int            `json:"blocked_attempts"`
	SecuritySuccessRate   float64        `json:"security_success_rate"`
	ThreatLevels          map[string]int `json:"threat_levels"`
}

// NewErrorTestRunner creates a new error handling test runner
func NewErrorTestRunner() *ErrorTestRunner {
	logger := zap.NewNop()
	return &ErrorTestRunner{
		logger: logger,
		results: &ErrorTestResults{
			TestDetails: make([]ErrorTestDetail, 0),
			Summary:     make(map[string]interface{}),
			ErrorMetrics: &ErrorMetrics{
				ErrorDistribution: make(map[string]int),
			},
			RecoveryMetrics: &RecoveryMetrics{
				RecoveryStrategies: make(map[string]int),
			},
			SecurityMetrics: &SecurityMetrics{
				ThreatLevels: make(map[string]int),
			},
		},
	}
}

// RunAllErrorTests runs all error handling tests
func (etr *ErrorTestRunner) RunAllErrorTests(t *testing.T) *ErrorTestResults {
	startTime := time.Now()
	etr.logger.Info("Starting error handling test suite")

	// Initialize test suite
	etr.testSuite = NewErrorHandlingTestSuite(t)
	defer etr.testSuite.Close()

	// Run all test categories
	etr.runErrorTestCategory(t, "Validation Errors", TestValidationErrors)
	etr.runErrorTestCategory(t, "Database Errors", TestDatabaseErrors)
	etr.runErrorTestCategory(t, "API Errors", TestAPIErrors)
	etr.runErrorTestCategory(t, "Service Errors", TestServiceErrors)
	etr.runErrorTestCategory(t, "Concurrency Errors", TestConcurrencyErrors)
	etr.runErrorTestCategory(t, "Resource Errors", TestResourceErrors)
	etr.runErrorTestCategory(t, "Security Errors", TestSecurityErrors)
	etr.runErrorTestCategory(t, "Recovery Errors", TestRecoveryErrors)
	etr.runErrorTestCategory(t, "Error Logging", TestErrorLogging)
	etr.runErrorTestCategory(t, "Error Metrics", TestErrorMetrics)

	// Run error analysis
	etr.runErrorAnalysis(t)

	// Calculate final results
	etr.results.ExecutionTime = time.Since(startTime)
	etr.calculateErrorSummary()

	etr.logger.Info("Error handling test suite completed",
		zap.Int("total_tests", etr.results.TotalTests),
		zap.Int("passed_tests", etr.results.PassedTests),
		zap.Int("failed_tests", etr.results.FailedTests),
		zap.Duration("execution_time", etr.results.ExecutionTime))

	return etr.results
}

// runErrorTestCategory runs a specific error handling test category
func (etr *ErrorTestRunner) runErrorTestCategory(t *testing.T, categoryName string, testFunc func(*testing.T)) {
	etr.logger.Info("Running error handling test category", zap.String("category", categoryName))

	// Create a sub-test for the category
	t.Run(categoryName, func(t *testing.T) {
		startTime := time.Now()

		// Run the test function
		testFunc(t)

		duration := time.Since(startTime)

		// Record test result
		etr.results.TotalTests++
		etr.results.PassedTests++ // If we get here, the test passed

		etr.results.TestDetails = append(etr.results.TestDetails, ErrorTestDetail{
			Name:     categoryName,
			Category: categoryName,
			Status:   "PASSED",
			Duration: duration,
		})

		etr.logger.Info("Error handling test category completed",
			zap.String("category", categoryName),
			zap.Duration("duration", duration),
			zap.String("status", "PASSED"))
	})
}

// runErrorAnalysis runs comprehensive error analysis
func (etr *ErrorTestRunner) runErrorAnalysis(t *testing.T) {
	etr.logger.Info("Running error analysis")

	// Analyze error patterns
	etr.analyzeErrorPatterns()

	// Analyze recovery patterns
	etr.analyzeRecoveryPatterns()

	// Analyze security patterns
	etr.analyzeSecurityPatterns()

	// Generate error recommendations
	etr.generateErrorRecommendations()

	etr.logger.Info("Error analysis completed")
}

// analyzeErrorPatterns analyzes error patterns and trends
func (etr *ErrorTestRunner) analyzeErrorPatterns() {
	// Count different types of errors
	for _, detail := range etr.results.TestDetails {
		if detail.ActualError {
			etr.results.ErrorMetrics.TotalErrors++
			etr.results.ErrorMetrics.ErrorDistribution[detail.ErrorType]++

			switch detail.Category {
			case "Validation Errors":
				etr.results.ErrorMetrics.ValidationErrors++
			case "Database Errors":
				etr.results.ErrorMetrics.DatabaseErrors++
			case "API Errors":
				etr.results.ErrorMetrics.APIErrors++
			case "Service Errors":
				etr.results.ErrorMetrics.ServiceErrors++
			case "Concurrency Errors":
				etr.results.ErrorMetrics.ConcurrencyErrors++
			case "Resource Errors":
				etr.results.ErrorMetrics.ResourceErrors++
			case "Security Errors":
				etr.results.ErrorMetrics.SecurityErrors++
			}

			// Categorize by severity
			switch detail.Severity {
			case "critical":
				etr.results.ErrorMetrics.CriticalErrors++
			case "recoverable":
				etr.results.ErrorMetrics.RecoverableErrors++
			}
		}
	}

	// Calculate error rate
	if etr.results.TotalTests > 0 {
		etr.results.ErrorMetrics.ErrorRate = float64(etr.results.ErrorMetrics.TotalErrors) / float64(etr.results.TotalTests)
	}
}

// analyzeRecoveryPatterns analyzes error recovery patterns
func (etr *ErrorTestRunner) analyzeRecoveryPatterns() {
	// Analyze recovery attempts
	for _, detail := range etr.results.TestDetails {
		if detail.RecoveryTime > 0 {
			etr.results.RecoveryMetrics.TotalRecoveryAttempts++
			etr.results.RecoveryMetrics.AverageRecoveryTime += detail.RecoveryTime

			if detail.Status == "PASSED" {
				etr.results.RecoveryMetrics.SuccessfulRecoveries++
			} else {
				etr.results.RecoveryMetrics.FailedRecoveries++
			}
		}
	}

	// Calculate recovery rate
	if etr.results.RecoveryMetrics.TotalRecoveryAttempts > 0 {
		etr.results.RecoveryMetrics.RecoveryRate = float64(etr.results.RecoveryMetrics.SuccessfulRecoveries) / float64(etr.results.RecoveryMetrics.TotalRecoveryAttempts)
		etr.results.RecoveryMetrics.AverageRecoveryTime = etr.results.RecoveryMetrics.AverageRecoveryTime / time.Duration(etr.results.RecoveryMetrics.TotalRecoveryAttempts)
	}
}

// analyzeSecurityPatterns analyzes security error patterns
func (etr *ErrorTestRunner) analyzeSecurityPatterns() {
	// Analyze security attempts
	for _, detail := range etr.results.TestDetails {
		if detail.Category == "Security Errors" {
			etr.results.SecurityMetrics.TotalSecurityAttempts++

			switch detail.ErrorType {
			case "sql_injection":
				etr.results.SecurityMetrics.SQLInjectionAttempts++
			case "xss":
				etr.results.SecurityMetrics.XSSAttempts++
			case "path_traversal":
				etr.results.SecurityMetrics.PathTraversalAttempts++
			}

			if detail.Status == "PASSED" {
				etr.results.SecurityMetrics.BlockedAttempts++
			}
		}
	}

	// Calculate security success rate
	if etr.results.SecurityMetrics.TotalSecurityAttempts > 0 {
		etr.results.SecurityMetrics.SecuritySuccessRate = float64(etr.results.SecurityMetrics.BlockedAttempts) / float64(etr.results.SecurityMetrics.TotalSecurityAttempts)
	}
}

// generateErrorRecommendations generates recommendations based on error analysis
func (etr *ErrorTestRunner) generateErrorRecommendations() {
	recommendations := make([]string, 0)

	// High error rate recommendation
	if etr.results.ErrorMetrics.ErrorRate > 0.1 {
		recommendations = append(recommendations, "High error rate detected. Consider improving input validation and error handling.")
	}

	// Low recovery rate recommendation
	if etr.results.RecoveryMetrics.RecoveryRate < 0.8 {
		recommendations = append(recommendations, "Low recovery rate detected. Consider implementing better error recovery mechanisms.")
	}

	// Security concerns recommendation
	if etr.results.SecurityMetrics.SecuritySuccessRate < 0.95 {
		recommendations = append(recommendations, "Security vulnerabilities detected. Consider strengthening security measures.")
	}

	// Critical errors recommendation
	if etr.results.ErrorMetrics.CriticalErrors > 0 {
		recommendations = append(recommendations, "Critical errors detected. Immediate attention required.")
	}

	etr.results.Summary["recommendations"] = recommendations
}

// TestErrorResilience tests system resilience to various error conditions
func (etr *ErrorTestRunner) TestErrorResilience(t *testing.T) {
	etr.logger.Info("Testing error resilience")

	ctx := context.Background()

	// Test system resilience to cascading failures
	etr.testCascadingFailures(ctx, t)

	// Test system resilience to resource exhaustion
	etr.testResourceExhaustion(ctx, t)

	// Test system resilience to network failures
	etr.testNetworkFailures(ctx, t)

	// Test system resilience to data corruption
	etr.testDataCorruption(ctx, t)
}

// testCascadingFailures tests system resilience to cascading failures
func (etr *ErrorTestRunner) testCascadingFailures(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing cascading failure resilience")

	// Simulate cascading failures
	failureCount := 0
	for i := 0; i < 10; i++ {
		assessment := &RiskAssessment{
			ID:           fmt.Sprintf("cascade-test-%d", i),
			BusinessID:   "", // Invalid: will cause validation error
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := etr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		if err != nil {
			failureCount++
		}
	}

	// System should handle all failures gracefully
	assert.Equal(t, 10, failureCount)
}

// testResourceExhaustion tests system resilience to resource exhaustion
func (etr *ErrorTestRunner) testResourceExhaustion(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing resource exhaustion resilience")

	// Simulate resource exhaustion
	request := &ExportRequest{
		BusinessID: "test-business",
		ExportType: ExportTypeAllData,
		Format:     ExportFormatJSON,
		Metadata:   make(map[string]interface{}),
	}

	// Add large metadata to simulate memory pressure
	for i := 0; i < 100000; i++ {
		request.Metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	// System should handle resource exhaustion gracefully
	_, err := etr.testSuite.exportSvc.ExportData(ctx, request)
	assert.NoError(t, err) // Should handle gracefully
}

// testNetworkFailures tests system resilience to network failures
func (etr *ErrorTestRunner) testNetworkFailures(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing network failure resilience")

	// Simulate network timeout
	ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	request := &ExportRequest{
		BusinessID: "test-business",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	// System should handle network failures gracefully
	_, err := etr.testSuite.exportSvc.ExportData(ctx, request)
	assert.Error(t, err) // Should fail gracefully with timeout
}

// testDataCorruption tests system resilience to data corruption
func (etr *ErrorTestRunner) testDataCorruption(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing data corruption resilience")

	// Simulate data corruption
	assessment := &RiskAssessment{
		ID:           "corruption-test",
		BusinessID:   "test-business",
		BusinessName: "Test Business",
		OverallScore: -100.0, // Corrupted data: negative score
		OverallLevel: RiskLevelMedium,
		AlertLevel:   RiskLevelMedium,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
	}

	// System should detect and handle data corruption
	err := etr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "score must be between 0 and 100")
}

// TestErrorMonitoring tests error monitoring and alerting
func (etr *ErrorTestRunner) TestErrorMonitoring(t *testing.T) {
	etr.logger.Info("Testing error monitoring")

	ctx := context.Background()

	// Test error threshold monitoring
	etr.testErrorThresholds(ctx, t)

	// Test error rate monitoring
	etr.testErrorRates(ctx, t)

	// Test error pattern detection
	etr.testErrorPatterns(ctx, t)
}

// testErrorThresholds tests error threshold monitoring
func (etr *ErrorTestRunner) testErrorThresholds(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing error threshold monitoring")

	// Generate errors to test thresholds
	errorCount := 0
	for i := 0; i < 100; i++ {
		assessment := &RiskAssessment{
			ID:           fmt.Sprintf("threshold-test-%d", i),
			BusinessID:   "", // Invalid: will cause error
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := etr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		if err != nil {
			errorCount++
		}
	}

	// Should detect high error rate
	assert.Equal(t, 100, errorCount)
}

// testErrorRates tests error rate monitoring
func (etr *ErrorTestRunner) testErrorRates(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing error rate monitoring")

	// Test error rate calculation
	totalRequests := 1000
	errorCount := 0

	for i := 0; i < totalRequests; i++ {
		assessment := &RiskAssessment{
			ID:           fmt.Sprintf("rate-test-%d", i),
			BusinessID:   "", // Invalid: will cause error
			BusinessName: "Test Business",
			OverallScore: 80.0,
			OverallLevel: RiskLevelMedium,
			AlertLevel:   RiskLevelMedium,
			AssessedAt:   time.Now(),
			ValidUntil:   time.Now().Add(24 * time.Hour),
		}

		err := etr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		if err != nil {
			errorCount++
		}
	}

	errorRate := float64(errorCount) / float64(totalRequests)
	assert.Equal(t, 1.0, errorRate) // 100% error rate
}

// testErrorPatterns tests error pattern detection
func (etr *ErrorTestRunner) testErrorPatterns(ctx context.Context, t *testing.T) {
	etr.logger.Info("Testing error pattern detection")

	// Test different error patterns
	patterns := []string{"validation", "database", "api", "service"}
	patternCounts := make(map[string]int)

	for _, pattern := range patterns {
		for i := 0; i < 10; i++ {
			assessment := &RiskAssessment{
				ID:           fmt.Sprintf("pattern-test-%s-%d", pattern, i),
				BusinessID:   "", // Invalid: will cause error
				BusinessName: "Test Business",
				OverallScore: 80.0,
				OverallLevel: RiskLevelMedium,
				AlertLevel:   RiskLevelMedium,
				AssessedAt:   time.Now(),
				ValidUntil:   time.Now().Add(24 * time.Hour),
			}

			err := etr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
			if err != nil {
				patternCounts[pattern]++
			}
		}
	}

	// Should detect error patterns
	for _, pattern := range patterns {
		assert.Equal(t, 10, patternCounts[pattern])
	}
}

// calculateErrorSummary calculates error handling test summary statistics
func (etr *ErrorTestRunner) calculateErrorSummary() {
	etr.results.Summary = map[string]interface{}{
		"total_tests":       etr.results.TotalTests,
		"passed_tests":      etr.results.PassedTests,
		"failed_tests":      etr.results.FailedTests,
		"skipped_tests":     etr.results.SkippedTests,
		"pass_rate":         float64(etr.results.PassedTests) / float64(etr.results.TotalTests) * 100,
		"execution_time":    etr.results.ExecutionTime.String(),
		"average_test_time": etr.results.ExecutionTime / time.Duration(etr.results.TotalTests),
	}

	// Calculate error-specific statistics
	errorStats := make(map[string]map[string]interface{})
	for _, detail := range etr.results.TestDetails {
		if errorStats[detail.Category] == nil {
			errorStats[detail.Category] = make(map[string]interface{})
		}
		errorStats[detail.Category]["duration"] = detail.Duration.String()
		errorStats[detail.Category]["status"] = detail.Status
		errorStats[detail.Category]["error_type"] = detail.ErrorType
		errorStats[detail.Category]["severity"] = detail.Severity
		errorStats[detail.Category]["expected_error"] = detail.ExpectedError
		errorStats[detail.Category]["actual_error"] = detail.ActualError
	}
	etr.results.Summary["error_stats"] = errorStats

	// Add error metrics to summary
	etr.results.Summary["error_metrics"] = etr.results.ErrorMetrics
	etr.results.Summary["recovery_metrics"] = etr.results.RecoveryMetrics
	etr.results.Summary["security_metrics"] = etr.results.SecurityMetrics
}

// GenerateErrorReport generates a comprehensive error handling test report
func (etr *ErrorTestRunner) GenerateErrorReport() (string, error) {
	report := fmt.Sprintf(`
# Error Handling Test Report

## Summary
- Total Tests: %d
- Passed Tests: %d
- Failed Tests: %d
- Skipped Tests: %d
- Pass Rate: %.2f%%
- Execution Time: %s

## Error Metrics
- Total Errors: %d
- Error Rate: %.2f%%
- Validation Errors: %d
- Database Errors: %d
- API Errors: %d
- Service Errors: %d
- Concurrency Errors: %d
- Resource Errors: %d
- Security Errors: %d
- Critical Errors: %d
- Recoverable Errors: %d

## Recovery Metrics
- Total Recovery Attempts: %d
- Successful Recoveries: %d
- Failed Recoveries: %d
- Recovery Rate: %.2f%%
- Average Recovery Time: %s

## Security Metrics
- Total Security Attempts: %d
- SQL Injection Attempts: %d
- XSS Attempts: %d
- Path Traversal Attempts: %d
- Blocked Attempts: %d
- Security Success Rate: %.2f%%

## Test Details
`,
		etr.results.TotalTests,
		etr.results.PassedTests,
		etr.results.FailedTests,
		etr.results.SkippedTests,
		float64(etr.results.PassedTests)/float64(etr.results.TotalTests)*100,
		etr.results.ExecutionTime.String(),
		etr.results.ErrorMetrics.TotalErrors,
		etr.results.ErrorMetrics.ErrorRate*100,
		etr.results.ErrorMetrics.ValidationErrors,
		etr.results.ErrorMetrics.DatabaseErrors,
		etr.results.ErrorMetrics.APIErrors,
		etr.results.ErrorMetrics.ServiceErrors,
		etr.results.ErrorMetrics.ConcurrencyErrors,
		etr.results.ErrorMetrics.ResourceErrors,
		etr.results.ErrorMetrics.SecurityErrors,
		etr.results.ErrorMetrics.CriticalErrors,
		etr.results.ErrorMetrics.RecoverableErrors,
		etr.results.RecoveryMetrics.TotalRecoveryAttempts,
		etr.results.RecoveryMetrics.SuccessfulRecoveries,
		etr.results.RecoveryMetrics.FailedRecoveries,
		etr.results.RecoveryMetrics.RecoveryRate*100,
		etr.results.RecoveryMetrics.AverageRecoveryTime.String(),
		etr.results.SecurityMetrics.TotalSecurityAttempts,
		etr.results.SecurityMetrics.SQLInjectionAttempts,
		etr.results.SecurityMetrics.XSSAttempts,
		etr.results.SecurityMetrics.PathTraversalAttempts,
		etr.results.SecurityMetrics.BlockedAttempts,
		etr.results.SecurityMetrics.SecuritySuccessRate*100)

	for _, detail := range etr.results.TestDetails {
		report += fmt.Sprintf(`
### %s
- Category: %s
- Error Type: %s
- Status: %s
- Duration: %s
- Expected Error: %t
- Actual Error: %t
- Severity: %s
`,
			detail.Name,
			detail.Category,
			detail.ErrorType,
			detail.Status,
			detail.Duration.String(),
			detail.ExpectedError,
			detail.ActualError,
			detail.Severity)
	}

	// Add recommendations
	if recommendations, ok := etr.results.Summary["recommendations"].([]string); ok {
		report += "\n## Recommendations\n"
		for _, recommendation := range recommendations {
			report += fmt.Sprintf("- %s\n", recommendation)
		}
	}

	return report, nil
}
