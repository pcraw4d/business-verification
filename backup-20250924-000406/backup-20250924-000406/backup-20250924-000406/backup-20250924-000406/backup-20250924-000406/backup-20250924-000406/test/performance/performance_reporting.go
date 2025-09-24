package performance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PerformanceReport represents a comprehensive performance test report
type PerformanceReport struct {
	TestSuite       string                 `json:"test_suite"`
	TestDate        time.Time              `json:"test_date"`
	Configuration   *PerformanceTestConfig `json:"configuration"`
	OverallMetrics  *PerformanceMetrics    `json:"overall_metrics"`
	TestResults     []TestResult           `json:"test_results"`
	Recommendations []string               `json:"recommendations"`
	PassedTests     int                    `json:"passed_tests"`
	FailedTests     int                    `json:"failed_tests"`
	TotalTests      int                    `json:"total_tests"`
	SuccessRate     float64                `json:"success_rate"`
}

// TestResult represents the result of a single performance test
type TestResult struct {
	TestName        string              `json:"test_name"`
	Description     string              `json:"description"`
	Status          string              `json:"status"` // "passed", "failed", "warning"
	Metrics         *PerformanceMetrics `json:"metrics"`
	Duration        time.Duration       `json:"duration"`
	ErrorMessage    string              `json:"error_message,omitempty"`
	Recommendations []string            `json:"recommendations,omitempty"`
}

// PerformanceReporter manages performance test reporting
type PerformanceReporter struct {
	reports   []*PerformanceReport
	outputDir string
	config    *PerformanceTestConfig
}

// NewPerformanceReporter creates a new performance reporter
func NewPerformanceReporter(outputDir string, config *PerformanceTestConfig) *PerformanceReporter {
	return &PerformanceReporter{
		reports:   make([]*PerformanceReport, 0),
		outputDir: outputDir,
		config:    config,
	}
}

// AddTestResult adds a test result to the current report
func (pr *PerformanceReporter) AddTestResult(testName, description string, metrics *PerformanceMetrics, duration time.Duration, err error) {
	// Find or create current report
	var currentReport *PerformanceReport
	if len(pr.reports) == 0 {
		currentReport = pr.createNewReport()
		pr.reports = append(pr.reports, currentReport)
	} else {
		currentReport = pr.reports[len(pr.reports)-1]
	}

	// Create test result
	testResult := TestResult{
		TestName:    testName,
		Description: description,
		Metrics:     metrics,
		Duration:    duration,
	}

	// Determine status and recommendations
	if err != nil {
		testResult.Status = "failed"
		testResult.ErrorMessage = err.Error()
		testResult.Recommendations = pr.generateErrorRecommendations(err)
		currentReport.FailedTests++
	} else {
		testResult.Status = pr.evaluateTestStatus(metrics)
		testResult.Recommendations = pr.generatePerformanceRecommendations(metrics)
		if testResult.Status == "passed" {
			currentReport.PassedTests++
		} else {
			currentReport.FailedTests++
		}
	}

	currentReport.TestResults = append(currentReport.TestResults, testResult)
	currentReport.TotalTests++
}

// createNewReport creates a new performance report
func (pr *PerformanceReporter) createNewReport() *PerformanceReport {
	return &PerformanceReport{
		TestSuite:       "Merchant Portfolio Performance Tests",
		TestDate:        time.Now(),
		Configuration:   pr.config,
		TestResults:     make([]TestResult, 0),
		Recommendations: make([]string, 0),
	}
}

// evaluateTestStatus evaluates the status of a performance test
func (pr *PerformanceReporter) evaluateTestStatus(metrics *PerformanceMetrics) string {
	// Check response time requirements
	if metrics.AverageResponseTime > pr.config.ResponseTimeLimit {
		return "failed"
	}

	// Check error rate requirements
	if metrics.ErrorRate > 5.0 {
		return "failed"
	}

	// Check throughput requirements
	if metrics.RequestsPerSecond < 1.0 {
		return "warning"
	}

	return "passed"
}

// generatePerformanceRecommendations generates recommendations based on performance metrics
func (pr *PerformanceReporter) generatePerformanceRecommendations(metrics *PerformanceMetrics) []string {
	recommendations := make([]string, 0)

	// Response time recommendations
	if metrics.AverageResponseTime > pr.config.ResponseTimeLimit/2 {
		recommendations = append(recommendations,
			"Consider optimizing database queries or implementing caching to improve response times")
	}

	if metrics.MaxResponseTime > pr.config.ResponseTimeLimit*2 {
		recommendations = append(recommendations,
			"Investigate and optimize slow requests that exceed acceptable response time limits")
	}

	// Throughput recommendations
	if metrics.RequestsPerSecond < 10.0 {
		recommendations = append(recommendations,
			"Consider implementing connection pooling or horizontal scaling to improve throughput")
	}

	// Error rate recommendations
	if metrics.ErrorRate > 1.0 {
		recommendations = append(recommendations,
			"Investigate and fix errors that are causing elevated error rates")
	}

	// Memory and resource recommendations
	if metrics.TotalRequests > 1000 && metrics.AverageResponseTime > 1*time.Second {
		recommendations = append(recommendations,
			"Consider implementing pagination or data streaming for large datasets")
	}

	return recommendations
}

// generateErrorRecommendations generates recommendations for failed tests
func (pr *PerformanceReporter) generateErrorRecommendations(err error) []string {
	recommendations := make([]string, 0)
	errorMsg := strings.ToLower(err.Error())

	if strings.Contains(errorMsg, "timeout") {
		recommendations = append(recommendations,
			"Increase timeout values or optimize slow operations")
	}

	if strings.Contains(errorMsg, "connection") {
		recommendations = append(recommendations,
			"Check database connections and implement connection pooling")
	}

	if strings.Contains(errorMsg, "memory") {
		recommendations = append(recommendations,
			"Optimize memory usage or increase available memory")
	}

	if strings.Contains(errorMsg, "concurrent") {
		recommendations = append(recommendations,
			"Review concurrent access patterns and implement proper locking")
	}

	return recommendations
}

// GenerateReport generates a comprehensive performance report
func (pr *PerformanceReporter) GenerateReport() (*PerformanceReport, error) {
	if len(pr.reports) == 0 {
		return nil, fmt.Errorf("no test results to generate report")
	}

	report := pr.reports[len(pr.reports)-1]

	// Calculate overall metrics
	report.OverallMetrics = pr.calculateOverallMetrics(report.TestResults)

	// Calculate success rate
	if report.TotalTests > 0 {
		report.SuccessRate = float64(report.PassedTests) / float64(report.TotalTests) * 100
	}

	// Generate overall recommendations
	report.Recommendations = pr.generateOverallRecommendations(report)

	return report, nil
}

// calculateOverallMetrics calculates overall performance metrics
func (pr *PerformanceReporter) calculateOverallMetrics(testResults []TestResult) *PerformanceMetrics {
	if len(testResults) == 0 {
		return &PerformanceMetrics{}
	}

	overall := &PerformanceMetrics{
		MinResponseTime: time.Hour,
	}

	for _, result := range testResults {
		if result.Metrics == nil {
			continue
		}

		metrics := result.Metrics
		overall.TotalRequests += metrics.TotalRequests
		overall.SuccessfulRequests += metrics.SuccessfulRequests
		overall.FailedRequests += metrics.FailedRequests

		if metrics.MaxResponseTime > overall.MaxResponseTime {
			overall.MaxResponseTime = metrics.MaxResponseTime
		}

		if metrics.MinResponseTime < overall.MinResponseTime {
			overall.MinResponseTime = metrics.MinResponseTime
		}
	}

	// Calculate averages
	if len(testResults) > 0 {
		totalAvgTime := time.Duration(0)
		totalRPS := 0.0
		validResults := 0

		for _, result := range testResults {
			if result.Metrics != nil && result.Metrics.TotalRequests > 0 {
				totalAvgTime += result.Metrics.AverageResponseTime
				totalRPS += result.Metrics.RequestsPerSecond
				validResults++
			}
		}

		if validResults > 0 {
			overall.AverageResponseTime = totalAvgTime / time.Duration(validResults)
			overall.RequestsPerSecond = totalRPS / float64(validResults)
		}
	}

	// Calculate error rate
	if overall.TotalRequests > 0 {
		overall.ErrorRate = float64(overall.FailedRequests) / float64(overall.TotalRequests) * 100
	}

	return overall
}

// generateOverallRecommendations generates overall recommendations
func (pr *PerformanceReporter) generateOverallRecommendations(report *PerformanceReport) []string {
	recommendations := make([]string, 0)

	// Success rate recommendations
	if report.SuccessRate < 90.0 {
		recommendations = append(recommendations,
			"Overall test success rate is below 90%. Review failed tests and implement fixes.")
	}

	// Performance recommendations
	if report.OverallMetrics.AverageResponseTime > pr.config.ResponseTimeLimit {
		recommendations = append(recommendations,
			"Average response time exceeds limits. Consider performance optimization.")
	}

	if report.OverallMetrics.ErrorRate > 2.0 {
		recommendations = append(recommendations,
			"Error rate is elevated. Investigate and fix underlying issues.")
	}

	// Scalability recommendations
	if report.OverallMetrics.RequestsPerSecond < float64(pr.config.ConcurrentUsers) {
		recommendations = append(recommendations,
			"Throughput may not support target concurrent users. Consider scaling improvements.")
	}

	// Add specific test recommendations
	for _, result := range report.TestResults {
		if result.Status == "failed" || result.Status == "warning" {
			recommendations = append(recommendations, result.Recommendations...)
		}
	}

	// Remove duplicates
	uniqueRecommendations := make([]string, 0)
	seen := make(map[string]bool)
	for _, rec := range recommendations {
		if !seen[rec] {
			uniqueRecommendations = append(uniqueRecommendations, rec)
			seen[rec] = true
		}
	}

	return uniqueRecommendations
}

// SaveReport saves the performance report to files
func (pr *PerformanceReporter) SaveReport(report *PerformanceReport) error {
	if err := os.MkdirAll(pr.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save JSON report
	jsonFile := filepath.Join(pr.outputDir, fmt.Sprintf("performance_report_%s.json",
		report.TestDate.Format("2006-01-02_15-04-05")))

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	// Save summary report
	summaryFile := filepath.Join(pr.outputDir, "performance_summary.txt")
	summaryContent := pr.generateSummaryReport(report)
	if err := os.WriteFile(summaryFile, []byte(summaryContent), 0644); err != nil {
		return fmt.Errorf("failed to write summary report: %w", err)
	}

	return nil
}

// generateSummaryReport generates a text summary report
func (pr *PerformanceReporter) generateSummaryReport(report *PerformanceReport) string {
	summary := fmt.Sprintf(`Performance Test Summary
========================

Test Suite: %s
Test Date: %s
Configuration: %d concurrent users, %d max merchants

Test Results:
- Total Tests: %d
- Passed: %d
- Failed: %d
- Success Rate: %.1f%%

Overall Performance Metrics:
- Total Requests: %d
- Successful Requests: %d
- Failed Requests: %d
- Error Rate: %.2f%%
- Average Response Time: %v
- Max Response Time: %v
- Min Response Time: %v
- Requests Per Second: %.2f

Test Details:
`,
		report.TestSuite,
		report.TestDate.Format("2006-01-02 15:04:05"),
		report.Configuration.ConcurrentUsers,
		report.Configuration.MaxMerchants,
		report.TotalTests,
		report.PassedTests,
		report.FailedTests,
		report.SuccessRate,
		report.OverallMetrics.TotalRequests,
		report.OverallMetrics.SuccessfulRequests,
		report.OverallMetrics.FailedRequests,
		report.OverallMetrics.ErrorRate,
		report.OverallMetrics.AverageResponseTime,
		report.OverallMetrics.MaxResponseTime,
		report.OverallMetrics.MinResponseTime,
		report.OverallMetrics.RequestsPerSecond)

	// Add test details
	for _, result := range report.TestResults {
		summary += fmt.Sprintf(`
%s - %s
Status: %s
Duration: %v
`,
			result.TestName,
			result.Description,
			strings.ToUpper(result.Status),
			result.Duration)

		if result.ErrorMessage != "" {
			summary += fmt.Sprintf("Error: %s\n", result.ErrorMessage)
		}

		if result.Metrics != nil {
			summary += fmt.Sprintf(`Metrics:
  - Total Requests: %d
  - Average Response Time: %v
  - Max Response Time: %v
  - Requests Per Second: %.2f
  - Error Rate: %.2f%%
`,
				result.Metrics.TotalRequests,
				result.Metrics.AverageResponseTime,
				result.Metrics.MaxResponseTime,
				result.Metrics.RequestsPerSecond,
				result.Metrics.ErrorRate)
		}

		if len(result.Recommendations) > 0 {
			summary += "Recommendations:\n"
			for _, rec := range result.Recommendations {
				summary += fmt.Sprintf("  - %s\n", rec)
			}
		}
	}

	// Add overall recommendations
	if len(report.Recommendations) > 0 {
		summary += "\nOverall Recommendations:\n"
		for _, rec := range report.Recommendations {
			summary += fmt.Sprintf("- %s\n", rec)
		}
	}

	return summary
}
