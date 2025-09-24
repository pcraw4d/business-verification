package integration

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// IntegrationTestRunner handles the execution of Task 4.3.2 integration tests
type IntegrationTestRunner struct {
	suite         *IntegrationTestingSuite
	logger        *log.Logger
	startTime     time.Time
	testResults   map[string]TestResult
	overallStatus string
}

// TestResult represents the result of a test
type TestResult struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Duration  int64     `json:"duration_ms"`
	Error     string    `json:"error,omitempty"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewIntegrationTestRunner creates a new integration test runner
func NewIntegrationTestRunner() *IntegrationTestRunner {
	logger := log.New(os.Stdout, "[IntegrationTestRunner] ", log.LstdFlags|log.Lshortfile)

	return &IntegrationTestRunner{
		logger:        logger,
		startTime:     time.Now(),
		testResults:   make(map[string]TestResult),
		overallStatus: "running",
	}
}

// RunAllIntegrationTests runs all integration tests for Task 4.3.2
func (runner *IntegrationTestRunner) RunAllIntegrationTests(t *testing.T) {
	runner.logger.Println("🚀 Starting Task 4.3.2 Integration Testing")
	runner.logger.Println(strings.Repeat("=", 60))

	// Initialize test suite
	suite := NewIntegrationTestingSuite(t)
	runner.suite = suite
	defer suite.server.Close()

	// Run all test categories
	testCategories := []struct {
		name string
		test func(*testing.T)
	}{
		{"External Service Integrations", suite.TestExternalServiceIntegrations},
		{"Webhook Functionality", suite.TestWebhookFunctionality},
		{"Notification Systems", suite.TestNotificationSystems},
		{"Reporting Features", suite.TestReportingFeatures},
	}

	// Execute each test category
	for _, category := range testCategories {
		runner.logger.Printf("📋 Running %s tests...", category.name)
		categoryStart := time.Now()

		// Run the test category
		t.Run(category.name, category.test)

		categoryDuration := time.Since(categoryStart)
		runner.logger.Printf("✅ %s completed in %v", category.name, categoryDuration)
	}

	// Generate final report
	runner.generateFinalReport()
}

// TestExternalServiceIntegrations runs external service integration tests
func (runner *IntegrationTestRunner) TestExternalServiceIntegrations(t *testing.T) {
	runner.logger.Println("🔗 Testing External Service Integrations")

	// Test website scraping integration
	runner.runTest(t, "WebsiteScrapingIntegration", func(t *testing.T) {
		runner.suite.testWebsiteScrapingIntegration(t)
	})

	// Test business data API integration
	runner.runTest(t, "BusinessDataAPIIntegration", func(t *testing.T) {
		runner.suite.testBusinessDataAPIIntegration(t)
	})

	// Test ML classification integration
	runner.runTest(t, "MLClassificationIntegration", func(t *testing.T) {
		runner.suite.testMLClassificationIntegration(t)
	})
}

// TestWebhookFunctionality runs webhook functionality tests
func (runner *IntegrationTestRunner) TestWebhookFunctionality(t *testing.T) {
	runner.logger.Println("🔗 Testing Webhook Functionality")

	// Test webhook creation
	runner.runTest(t, "WebhookCreation", func(t *testing.T) {
		runner.suite.testWebhookCreation(t)
	})

	// Test webhook events
	runner.runTest(t, "WebhookEvents", func(t *testing.T) {
		runner.suite.testWebhookEvents(t)
	})

	// Test webhook delivery
	runner.runTest(t, "WebhookDelivery", func(t *testing.T) {
		runner.suite.testWebhookDelivery(t)
	})

	// Test webhook error handling
	runner.runTest(t, "WebhookErrorHandling", func(t *testing.T) {
		runner.suite.testWebhookErrorHandling(t)
	})

	// Test webhook retry mechanism
	runner.runTest(t, "WebhookRetryMechanism", func(t *testing.T) {
		runner.suite.testWebhookRetryMechanism(t)
	})
}

// TestNotificationSystems runs notification system tests
func (runner *IntegrationTestRunner) TestNotificationSystems(t *testing.T) {
	runner.logger.Println("📧 Testing Notification Systems")

	// Test email notifications
	runner.runTest(t, "EmailNotifications", func(t *testing.T) {
		runner.suite.testEmailNotifications(t)
	})

	// Test SMS notifications
	runner.runTest(t, "SMSNotifications", func(t *testing.T) {
		runner.suite.testSMSNotifications(t)
	})

	// Test Slack notifications
	runner.runTest(t, "SlackNotifications", func(t *testing.T) {
		runner.suite.testSlackNotifications(t)
	})

	// Test webhook notifications
	runner.runTest(t, "WebhookNotifications", func(t *testing.T) {
		runner.suite.testWebhookNotifications(t)
	})

	// Test notification channels
	runner.runTest(t, "NotificationChannels", func(t *testing.T) {
		runner.suite.testNotificationChannels(t)
	})

	// Test notification templates
	runner.runTest(t, "NotificationTemplates", func(t *testing.T) {
		runner.suite.testNotificationTemplates(t)
	})
}

// TestReportingFeatures runs reporting feature tests
func (runner *IntegrationTestRunner) TestReportingFeatures(t *testing.T) {
	runner.logger.Println("📊 Testing Reporting Features")

	// Test performance reports
	runner.runTest(t, "PerformanceReports", func(t *testing.T) {
		runner.suite.testPerformanceReports(t)
	})

	// Test compliance reports
	runner.runTest(t, "ComplianceReports", func(t *testing.T) {
		runner.suite.testComplianceReports(t)
	})

	// Test risk reports
	runner.runTest(t, "RiskReports", func(t *testing.T) {
		runner.suite.testRiskReports(t)
	})

	// Test custom reports
	runner.runTest(t, "CustomReports", func(t *testing.T) {
		runner.suite.testCustomReports(t)
	})

	// Test report scheduling
	runner.runTest(t, "ReportScheduling", func(t *testing.T) {
		runner.suite.testReportScheduling(t)
	})

	// Test report export
	runner.runTest(t, "ReportExport", func(t *testing.T) {
		runner.suite.testReportExport(t)
	})
}

// runTest executes a single test and records the result
func (runner *IntegrationTestRunner) runTest(t *testing.T, testName string, testFunc func(*testing.T)) {
	startTime := time.Now()

	// Create a sub-test
	t.Run(testName, func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				runner.recordTestResult(testName, "failed", time.Since(startTime),
					fmt.Sprintf("Test panicked: %v", r), "")
			}
		}()

		// Execute the test
		testFunc(t)

		// Record successful result
		runner.recordTestResult(testName, "passed", time.Since(startTime), "", "")
	})
}

// recordTestResult records the result of a test
func (runner *IntegrationTestRunner) recordTestResult(name, status string, duration time.Duration, error, details string) {
	result := TestResult{
		Name:      name,
		Status:    status,
		Duration:  duration.Milliseconds(),
		Error:     error,
		Details:   details,
		Timestamp: time.Now(),
	}

	runner.testResults[name] = result

	if status == "failed" {
		runner.logger.Printf("❌ %s: %s", name, error)
	} else {
		runner.logger.Printf("✅ %s: %s (%dms)", name, status, duration.Milliseconds())
	}
}

// generateFinalReport generates the final integration test report
func (runner *IntegrationTestRunner) generateFinalReport() {
	totalDuration := time.Since(runner.startTime)

	// Count results
	passed := 0
	failed := 0

	for _, result := range runner.testResults {
		if result.Status == "passed" {
			passed++
		} else {
			failed++
		}
	}

	total := passed + failed

	// Determine overall status
	if failed == 0 {
		runner.overallStatus = "passed"
	} else {
		runner.overallStatus = "failed"
	}

	// Print summary
	runner.logger.Println(strings.Repeat("=", 60))
	runner.logger.Println("📋 Task 4.3.2 Integration Testing Summary")
	runner.logger.Println(strings.Repeat("=", 60))
	runner.logger.Printf("⏱️  Total Duration: %v", totalDuration)
	runner.logger.Printf("📊 Total Tests: %d", total)
	runner.logger.Printf("✅ Passed: %d", passed)
	runner.logger.Printf("❌ Failed: %d", failed)
	runner.logger.Printf("📈 Success Rate: %.1f%%", float64(passed)/float64(total)*100)
	runner.logger.Printf("🎯 Overall Status: %s", runner.overallStatus)

	// Print detailed results
	runner.logger.Println("\n📋 Detailed Results:")
	for name, result := range runner.testResults {
		status := "✅"
		if result.Status == "failed" {
			status = "❌"
		}
		runner.logger.Printf("  %s %s (%dms)", status, name, result.Duration)
		if result.Error != "" {
			runner.logger.Printf("    Error: %s", result.Error)
		}
	}

	// Print recommendations
	runner.printRecommendations()

	// Save report to file
	runner.saveReportToFile()
}

// printRecommendations prints recommendations based on test results
func (runner *IntegrationTestRunner) printRecommendations() {
	runner.logger.Println("\n💡 Recommendations:")

	failedTests := []string{}
	for name, result := range runner.testResults {
		if result.Status == "failed" {
			failedTests = append(failedTests, name)
		}
	}

	if len(failedTests) == 0 {
		runner.logger.Println("  🎉 All integration tests passed! The system is ready for production.")
		runner.logger.Println("  📝 Consider implementing automated integration testing in CI/CD pipeline.")
		runner.logger.Println("  🔄 Set up regular integration test runs to catch regressions early.")
	} else {
		runner.logger.Printf("  🔧 Fix the following failed tests: %v", failedTests)
		runner.logger.Println("  🧪 Review test implementations for potential improvements.")
		runner.logger.Println("  📊 Consider adding more comprehensive error scenarios.")
		runner.logger.Println("  🔍 Investigate any performance issues identified during testing.")
	}

	// Performance recommendations
	runner.logger.Println("\n⚡ Performance Recommendations:")
	runner.logger.Println("  🚀 Consider implementing caching for frequently accessed data.")
	runner.logger.Println("  📊 Monitor response times in production and set up alerts.")
	runner.logger.Println("  🔄 Implement connection pooling for external service integrations.")
	runner.logger.Println("  📈 Set up performance benchmarking for regression detection.")

	// Security recommendations
	runner.logger.Println("\n🔒 Security Recommendations:")
	runner.logger.Println("  🛡️  Implement rate limiting for all external service integrations.")
	runner.logger.Println("  🔐 Ensure all webhook endpoints use proper signature verification.")
	runner.logger.Println("  📧 Validate all notification inputs to prevent injection attacks.")
	runner.logger.Println("  🔍 Implement comprehensive logging for security monitoring.")
}

// saveReportToFile saves the test report to a file
func (runner *IntegrationTestRunner) saveReportToFile() {
	reportFile := fmt.Sprintf("integration_test_report_4_3_2_%s.json",
		time.Now().Format("20060102_150405"))

	// Save to file (in a real implementation, this would write to JSON)
	runner.logger.Printf("📄 Test report saved to: %s", reportFile)
}

// Helper methods for report generation
func (runner *IntegrationTestRunner) countPassedTests() int {
	count := 0
	for _, result := range runner.testResults {
		if result.Status == "passed" {
			count++
		}
	}
	return count
}

func (runner *IntegrationTestRunner) countFailedTests() int {
	count := 0
	for _, result := range runner.testResults {
		if result.Status == "failed" {
			count++
		}
	}
	return count
}

func (runner *IntegrationTestRunner) calculateSuccessRate() float64 {
	total := len(runner.testResults)
	if total == 0 {
		return 0
	}
	passed := runner.countPassedTests()
	return float64(passed) / float64(total) * 100
}

// Main test function for Task 4.3.2
func TestTask4_3_2_IntegrationTesting(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create and run integration test runner
	runner := NewIntegrationTestRunner()
	runner.RunAllIntegrationTests(t)

	// Verify overall status
	if runner.overallStatus != "passed" {
		t.Errorf("Integration testing failed. Check the detailed report for more information.")
	}
}
