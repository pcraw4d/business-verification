package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleSecurityTestRunner provides a simplified security testing framework
type SimpleSecurityTestRunner struct {
	testSuite *SimpleSecurityTestSuite
	results   []SecurityTestResult
	reportDir string
}

// NewSimpleSecurityTestRunner creates a new simplified security test runner
func NewSimpleSecurityTestRunner(t *testing.T) *SimpleSecurityTestRunner {
	testSuite := NewSimpleSecurityTestSuite(t)

	// Create report directory
	reportDir := "test/reports/security"
	err := os.MkdirAll(reportDir, 0755)
	require.NoError(t, err)

	return &SimpleSecurityTestRunner{
		testSuite: testSuite,
		results:   []SecurityTestResult{},
		reportDir: reportDir,
	}
}

// RunComprehensiveSecurityTests runs all security tests and generates reports
func (str *SimpleSecurityTestRunner) RunComprehensiveSecurityTests(t *testing.T) {
	t.Helper()

	t.Log("Starting comprehensive security testing...")

	// Run all security tests
	str.results = str.testSuite.RunAllSecurityTests(t)

	// Generate reports
	str.generateJSONReport(t)
	str.generateMarkdownReport(t)
	str.generateSummaryReport(t)

	// Validate results
	str.validateSecurityResults(t)

	t.Log("Security testing completed successfully")
}

// generateJSONReport generates a JSON report of security test results
func (str *SimpleSecurityTestRunner) generateJSONReport(t *testing.T) {
	t.Helper()

	report := SecurityTestReport{
		GeneratedAt:     time.Now(),
		TestSuite:       "KYB Platform Security Tests (Simplified)",
		Version:         "1.0.0",
		TotalTests:      len(str.results),
		Results:         str.results,
		Summary:         str.calculateSummary(),
		Recommendations: str.generateRecommendations(),
	}

	// Write JSON report
	jsonData, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)

	jsonPath := filepath.Join(str.reportDir, "security_test_results.json")
	err = os.WriteFile(jsonPath, jsonData, 0644)
	require.NoError(t, err)

	t.Logf("JSON report generated: %s", jsonPath)
}

// generateMarkdownReport generates a markdown report of security test results
func (str *SimpleSecurityTestRunner) generateMarkdownReport(t *testing.T) {
	t.Helper()

	markdownReport := str.testSuite.GenerateSecurityReport(str.results)

	markdownPath := filepath.Join(str.reportDir, "security_test_report.md")
	err := os.WriteFile(markdownPath, []byte(markdownReport), 0644)
	require.NoError(t, err)

	t.Logf("Markdown report generated: %s", markdownPath)
}

// generateSummaryReport generates a summary report for quick review
func (str *SimpleSecurityTestRunner) generateSummaryReport(t *testing.T) {
	t.Helper()

	summary := str.calculateSummary()

	summaryReport := fmt.Sprintf(`# Security Test Summary

Generated: %s

## Quick Overview
- Total Tests: %d
- Passed: %d (%.1f%%)
- Failed: %d (%.1f%%)
- Warnings: %d (%.1f%%)

## Critical Issues
%s

## Security Score: %d/100

## Next Steps
%s
`,
		time.Now().Format("2006-01-02 15:04:05"),
		summary.TotalTests,
		summary.Passed,
		float64(summary.Passed)/float64(summary.TotalTests)*100,
		summary.Failed,
		float64(summary.Failed)/float64(summary.TotalTests)*100,
		summary.Warnings,
		float64(summary.Warnings)/float64(summary.TotalTests)*100,
		str.formatCriticalIssues(),
		str.calculateSecurityScore(),
		str.formatNextSteps(),
	)

	summaryPath := filepath.Join(str.reportDir, "security_summary.md")
	err := os.WriteFile(summaryPath, []byte(summaryReport), 0644)
	require.NoError(t, err)

	t.Logf("Summary report generated: %s", summaryPath)
}

// validateSecurityResults validates that security tests meet minimum requirements
func (str *SimpleSecurityTestRunner) validateSecurityResults(t *testing.T) {
	t.Helper()

	summary := str.calculateSummary()

	// Critical security requirements
	t.Run("Security Test Validation", func(t *testing.T) {
		// No critical failures allowed
		assert.Equal(t, 0, summary.CriticalFailures, "Critical security failures must be zero")

		// Minimum 80% pass rate
		passRate := float64(summary.Passed) / float64(summary.TotalTests) * 100
		assert.GreaterOrEqual(t, passRate, 80.0, "Security test pass rate must be at least 80%%")

		// Authentication tests must pass
		authTests := str.getTestsByCategory("AUTHENTICATION")
		authPassed := 0
		for _, test := range authTests {
			if test.Status == "PASS" {
				authPassed++
			}
		}
		assert.Greater(t, authPassed, 0, "At least one authentication test must pass")

		// Authorization tests must pass
		authzTests := str.getTestsByCategory("AUTHORIZATION")
		authzPassed := 0
		for _, test := range authzTests {
			if test.Status == "PASS" {
				authzPassed++
			}
		}
		assert.Greater(t, authzPassed, 0, "At least one authorization test must pass")

		// Input validation tests must pass
		inputTests := str.getTestsByCategory("INPUT_VALIDATION")
		inputPassed := 0
		for _, test := range inputTests {
			if test.Status == "PASS" {
				inputPassed++
			}
		}
		assert.Greater(t, inputPassed, 0, "At least one input validation test must pass")
	})
}

// calculateSummary calculates summary statistics for the test results
func (str *SimpleSecurityTestRunner) calculateSummary() TestSummary {
	summary := TestSummary{
		TotalTests: len(str.results),
	}

	for _, result := range str.results {
		switch result.Status {
		case "PASS":
			summary.Passed++
		case "FAIL":
			summary.Failed++
			// Check if it's a critical failure
			if str.isCriticalFailure(result) {
				summary.CriticalFailures++
			}
		case "WARN":
			summary.Warnings++
		}
	}

	return summary
}

// isCriticalFailure determines if a test failure is critical
func (str *SimpleSecurityTestRunner) isCriticalFailure(result SecurityTestResult) bool {
	criticalCategories := []string{"AUTHENTICATION", "AUTHORIZATION", "INPUT_VALIDATION"}
	criticalTestNames := []string{
		"Invalid Token Rejection",
		"Missing Authentication Header",
		"User Role Access Denied",
		"SQL Injection Prevention",
		"XSS Prevention",
	}

	// Check category
	for _, category := range criticalCategories {
		if result.Category == category && result.Status == "FAIL" {
			return true
		}
	}

	// Check specific test names
	for _, testName := range criticalTestNames {
		if result.TestName == testName && result.Status == "FAIL" {
			return true
		}
	}

	return false
}

// getTestsByCategory returns tests filtered by category
func (str *SimpleSecurityTestRunner) getTestsByCategory(category string) []SecurityTestResult {
	var filtered []SecurityTestResult
	for _, result := range str.results {
		if result.Category == category {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// calculateSecurityScore calculates an overall security score (0-100)
func (str *SimpleSecurityTestRunner) calculateSecurityScore() int {
	if len(str.results) == 0 {
		return 0
	}

	score := 0
	for _, result := range str.results {
		switch result.Status {
		case "PASS":
			score += 10
		case "WARN":
			score += 5
		case "FAIL":
			if str.isCriticalFailure(result) {
				score -= 20 // Critical failures heavily penalize score
			} else {
				score -= 5 // Regular failures have moderate penalty
			}
		}
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateRecommendations generates security recommendations based on test results
func (str *SimpleSecurityTestRunner) generateRecommendations() []string {
	var recommendations []string

	summary := str.calculateSummary()

	// Critical failures
	if summary.CriticalFailures > 0 {
		recommendations = append(recommendations,
			"🚨 CRITICAL: Address all critical security failures immediately")
	}

	// Authentication issues
	authTests := str.getTestsByCategory("AUTHENTICATION")
	for _, test := range authTests {
		if test.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("🔐 Authentication: Fix %s", test.TestName))
		}
	}

	// Authorization issues
	authzTests := str.getTestsByCategory("AUTHORIZATION")
	for _, test := range authzTests {
		if test.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("🛡️ Authorization: Fix %s", test.TestName))
		}
	}

	// Input validation issues
	inputTests := str.getTestsByCategory("INPUT_VALIDATION")
	for _, test := range inputTests {
		if test.Status == "FAIL" {
			recommendations = append(recommendations,
				fmt.Sprintf("🔍 Input Validation: Fix %s", test.TestName))
		}
	}

	// General recommendations
	if summary.Passed == summary.TotalTests {
		recommendations = append(recommendations,
			"✅ Excellent: All security tests passed. Continue regular security testing.")
	} else {
		recommendations = append(recommendations,
			"📋 General: Implement regular security testing and monitoring")
		recommendations = append(recommendations,
			"🔍 Monitoring: Set up continuous security monitoring and alerting")
	}

	return recommendations
}

// formatCriticalIssues formats critical issues for the summary report
func (str *SimpleSecurityTestRunner) formatCriticalIssues() string {
	criticalIssues := []string{}

	for _, result := range str.results {
		if str.isCriticalFailure(result) {
			criticalIssues = append(criticalIssues,
				fmt.Sprintf("- %s: %s", result.Category, result.TestName))
		}
	}

	if len(criticalIssues) == 0 {
		return "None - All critical security tests passed ✅"
	}

	return fmt.Sprintf("%d critical issues found:\n%s", len(criticalIssues),
		fmt.Sprintf("%s", criticalIssues))
}

// formatNextSteps formats next steps for the summary report
func (str *SimpleSecurityTestRunner) formatNextSteps() string {
	summary := str.calculateSummary()

	if summary.CriticalFailures > 0 {
		return "1. 🚨 Address critical security failures immediately\n2. 🔍 Review authentication and authorization implementations\n3. 🛡️ Implement missing security controls\n4. 📋 Schedule follow-up security testing"
	}

	if summary.Failed > 0 {
		return "1. 🔧 Fix failed security tests\n2. ⚠️ Address security warnings\n3. 📋 Schedule follow-up security testing\n4. 🔍 Implement continuous security monitoring"
	}

	if summary.Warnings > 0 {
		return "1. ⚠️ Address security warnings\n2. 📋 Schedule follow-up security testing\n3. 🔍 Implement continuous security monitoring\n4. 📚 Consider security training for team"
	}

	return "1. ✅ Continue regular security testing\n2. 🔍 Implement continuous security monitoring\n3. 📚 Provide security training for team\n4. 🚀 Consider advanced security features"
}

// Close cleans up the test runner
func (str *SimpleSecurityTestRunner) Close() {
	if str.testSuite != nil {
		str.testSuite.Close()
	}
}

// TestSimpleSecurityTestRunner tests the simplified security test runner functionality
func TestSimpleSecurityTestRunner(t *testing.T) {
	runner := NewSimpleSecurityTestRunner(t)
	defer runner.Close()

	// Run comprehensive security tests
	runner.RunComprehensiveSecurityTests(t)

	// Verify reports were generated
	jsonPath := filepath.Join(runner.reportDir, "security_test_results.json")
	markdownPath := filepath.Join(runner.reportDir, "security_test_report.md")
	summaryPath := filepath.Join(runner.reportDir, "security_summary.md")

	assert.FileExists(t, jsonPath, "JSON report should be generated")
	assert.FileExists(t, markdownPath, "Markdown report should be generated")
	assert.FileExists(t, summaryPath, "Summary report should be generated")

	// Verify report content
	jsonData, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	var report SecurityTestReport
	err = json.Unmarshal(jsonData, &report)
	require.NoError(t, err)

	assert.Greater(t, report.TotalTests, 0, "Should have test results")
	assert.NotEmpty(t, report.Results, "Should have test results")
	assert.NotEmpty(t, report.Recommendations, "Should have recommendations")
}
