package risk

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ManualTestRunner provides execution and management of manual test scenarios
type ManualTestRunner struct {
	logger          *zap.Logger
	guide           *ManualTestingGuide
	config          *ManualTestConfig
	results         *ManualTestResults
	reportGenerator *ManualTestReportGenerator
}

// ManualTestConfig contains configuration for manual testing
type ManualTestConfig struct {
	TestEnvironment      string            `json:"test_environment"`
	TestTimeout          time.Duration     `json:"test_timeout"`
	ReportOutputPath     string            `json:"report_output_path"`
	LogLevel             string            `json:"log_level"`
	EnableScreenshots    bool              `json:"enable_screenshots"`
	ScreenshotPath       string            `json:"screenshot_path"`
	TestDataPath         string            `json:"test_data_path"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	BrowserConfig        *BrowserConfig    `json:"browser_config"`
	APIConfig            *APIConfig        `json:"api_config"`
}

// BrowserConfig contains browser configuration for manual testing
type BrowserConfig struct {
	BrowserType   string            `json:"browser_type"` // Chrome, Firefox, Safari, Edge
	Headless      bool              `json:"headless"`
	WindowSize    string            `json:"window_size"` // e.g., "1920x1080"
	UserAgent     string            `json:"user_agent"`
	ProxySettings map[string]string `json:"proxy_settings"`
	Extensions    []string          `json:"extensions"`
	Cookies       []Cookie          `json:"cookies"`
}

// Cookie represents a browser cookie
type Cookie struct {
	Name     string     `json:"name"`
	Value    string     `json:"value"`
	Domain   string     `json:"domain"`
	Path     string     `json:"path"`
	Secure   bool       `json:"secure"`
	HttpOnly bool       `json:"http_only"`
	Expires  *time.Time `json:"expires,omitempty"`
}

// ManualTestReportGenerator generates reports for manual testing
type ManualTestReportGenerator struct {
	logger *zap.Logger
	config *ManualTestConfig
}

// NewManualTestRunner creates a new manual test runner
func NewManualTestRunner(config *ManualTestConfig) *ManualTestRunner {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	guide := NewManualTestingGuide(logger)

	// Load test scenarios
	scenarios := CreateKYBTestScenarios()
	for _, scenario := range scenarios {
		guide.AddTestScenario(scenario)
	}

	// Load workflow tests
	workflowTests := CreateKYBWorkflowTests()
	for _, workflowTest := range workflowTests {
		guide.AddWorkflowTest(workflowTest)
	}

	// Load validation rules
	validationRules := CreateKYBValidationRules()
	for _, rule := range validationRules {
		guide.AddValidationRule(rule)
	}

	return &ManualTestRunner{
		logger:          logger,
		guide:           guide,
		config:          config,
		results:         &ManualTestResults{},
		reportGenerator: NewManualTestReportGenerator(logger, config),
	}
}

// NewManualTestReportGenerator creates a new manual test report generator
func NewManualTestReportGenerator(logger *zap.Logger, config *ManualTestConfig) *ManualTestReportGenerator {
	return &ManualTestReportGenerator{
		logger: logger,
		config: config,
	}
}

// RunManualTestSuite runs the complete manual test suite
func (mtr *ManualTestRunner) RunManualTestSuite(ctx context.Context, testerName string) (*ManualTestResults, error) {
	mtr.logger.Info("Starting manual test suite execution",
		zap.String("tester", testerName),
		zap.String("environment", mtr.config.TestEnvironment))

	// Initialize results
	mtr.results = &ManualTestResults{
		TestSessionID:   fmt.Sprintf("manual_session_%d", time.Now().Unix()),
		StartTime:       time.Now(),
		TesterName:      testerName,
		TestEnvironment: mtr.config.TestEnvironment,
		ScenarioResults: make(map[string]ScenarioResult),
		IssuesFound:     make([]Issue, 0),
		Recommendations: make([]string, 0),
		Summary:         make(map[string]interface{}),
	}

	// Get all workflow tests
	workflowTests := mtr.guide.GetWorkflowTests()
	mtr.results.TotalScenarios = len(workflowTests)

	// Execute each workflow test
	for _, workflowTest := range workflowTests {
		select {
		case <-ctx.Done():
			mtr.results.EndTime = time.Now()
			mtr.results.TotalDuration = mtr.results.EndTime.Sub(mtr.results.StartTime)
			return mtr.results, ctx.Err()
		default:
		}

		mtr.logger.Info("Executing workflow test", zap.String("id", workflowTest.ID), zap.String("name", workflowTest.Name))

		workflowResults, err := mtr.guide.ExecuteWorkflowTest(ctx, workflowTest.ID, testerName)
		if err != nil {
			mtr.logger.Error("Failed to execute workflow test", zap.String("id", workflowTest.ID), zap.Error(err))
			continue
		}

		// Aggregate results
		mtr.results.PassedScenarios += workflowResults.PassedScenarios
		mtr.results.FailedScenarios += workflowResults.FailedScenarios
		mtr.results.SkippedScenarios += workflowResults.SkippedScenarios

		// Merge scenario results
		for scenarioID, scenarioResult := range workflowResults.ScenarioResults {
			mtr.results.ScenarioResults[scenarioID] = scenarioResult
		}

		// Collect issues
		mtr.results.IssuesFound = append(mtr.results.IssuesFound, workflowResults.IssuesFound...)
	}

	// Calculate pass rate
	if mtr.results.TotalScenarios > 0 {
		mtr.results.PassRate = float64(mtr.results.PassedScenarios) / float64(mtr.results.TotalScenarios) * 100
	}

	mtr.results.EndTime = time.Now()
	mtr.results.TotalDuration = mtr.results.EndTime.Sub(mtr.results.StartTime)

	// Generate summary
	mtr.generateSummary()

	// Generate recommendations
	mtr.generateRecommendations()

	// Generate reports
	if err := mtr.generateReports(); err != nil {
		mtr.logger.Error("Failed to generate reports", zap.Error(err))
	}

	mtr.logger.Info("Manual test suite execution completed",
		zap.Int("total_scenarios", mtr.results.TotalScenarios),
		zap.Int("passed_scenarios", mtr.results.PassedScenarios),
		zap.Int("failed_scenarios", mtr.results.FailedScenarios),
		zap.Float64("pass_rate", mtr.results.PassRate))

	return mtr.results, nil
}

// RunSpecificWorkflow runs a specific workflow test
func (mtr *ManualTestRunner) RunSpecificWorkflow(ctx context.Context, workflowID string, testerName string) (*ManualTestResults, error) {
	mtr.logger.Info("Running specific workflow test", zap.String("workflow_id", workflowID), zap.String("tester", testerName))

	results, err := mtr.guide.ExecuteWorkflowTest(ctx, workflowID, testerName)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow test: %w", err)
	}

	// Generate reports for specific workflow
	if err := mtr.reportGenerator.GenerateWorkflowReport(results); err != nil {
		mtr.logger.Error("Failed to generate workflow report", zap.Error(err))
	}

	return results, nil
}

// RunSpecificScenario runs a specific test scenario
func (mtr *ManualTestRunner) RunSpecificScenario(ctx context.Context, scenarioID string, testerName string) (*ScenarioResult, error) {
	mtr.logger.Info("Running specific test scenario", zap.String("scenario_id", scenarioID), zap.String("tester", testerName))

	result, err := mtr.guide.ExecuteTestScenario(ctx, scenarioID, testerName)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test scenario: %w", err)
	}

	// Generate report for specific scenario
	if err := mtr.reportGenerator.GenerateScenarioReport(result); err != nil {
		mtr.logger.Error("Failed to generate scenario report", zap.Error(err))
	}

	return result, nil
}

// generateSummary generates a summary of the manual test results
func (mtr *ManualTestRunner) generateSummary() {
	mtr.results.Summary = map[string]interface{}{
		"test_session_id":   mtr.results.TestSessionID,
		"tester_name":       mtr.results.TesterName,
		"test_environment":  mtr.results.TestEnvironment,
		"start_time":        mtr.results.StartTime.Format(time.RFC3339),
		"end_time":          mtr.results.EndTime.Format(time.RFC3339),
		"total_duration":    mtr.results.TotalDuration.String(),
		"total_scenarios":   mtr.results.TotalScenarios,
		"passed_scenarios":  mtr.results.PassedScenarios,
		"failed_scenarios":  mtr.results.FailedScenarios,
		"skipped_scenarios": mtr.results.SkippedScenarios,
		"pass_rate":         mtr.results.PassRate,
		"total_issues":      len(mtr.results.IssuesFound),
		"critical_issues":   mtr.countIssuesBySeverity("Critical"),
		"high_issues":       mtr.countIssuesBySeverity("High"),
		"medium_issues":     mtr.countIssuesBySeverity("Medium"),
		"low_issues":        mtr.countIssuesBySeverity("Low"),
	}
}

// countIssuesBySeverity counts issues by severity level
func (mtr *ManualTestRunner) countIssuesBySeverity(severity string) int {
	count := 0
	for _, issue := range mtr.results.IssuesFound {
		if issue.Severity == severity {
			count++
		}
	}
	return count
}

// generateRecommendations generates recommendations based on test results
func (mtr *ManualTestRunner) generateRecommendations() {
	recommendations := make([]string, 0)

	// Low pass rate recommendation
	if mtr.results.PassRate < 90 {
		recommendations = append(recommendations, "Low pass rate detected. Review failed scenarios and fix underlying issues.")
	}

	// High failure rate recommendation
	if mtr.results.FailedScenarios > 0 {
		recommendations = append(recommendations, "Test failures detected. Review failed scenarios and implement fixes.")
	}

	// Critical issues recommendation
	criticalIssues := mtr.countIssuesBySeverity("Critical")
	if criticalIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d critical issues found. Address these issues immediately.", criticalIssues))
	}

	// High priority issues recommendation
	highIssues := mtr.countIssuesBySeverity("High")
	if highIssues > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d high priority issues found. Address these issues soon.", highIssues))
	}

	// Long execution time recommendation
	if mtr.results.TotalDuration > 2*time.Hour {
		recommendations = append(recommendations, "Long execution time detected. Consider optimizing test scenarios and reducing complexity.")
	}

	mtr.results.Recommendations = recommendations
}

// generateReports generates comprehensive reports for manual testing
func (mtr *ManualTestRunner) generateReports() error {
	mtr.logger.Info("Generating manual test reports")

	// Create report output directory if it doesn't exist
	if err := os.MkdirAll(mtr.config.ReportOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create report output directory: %w", err)
	}

	// Generate JSON report
	if err := mtr.reportGenerator.GenerateJSONReport(mtr.results); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate HTML report
	if err := mtr.reportGenerator.GenerateHTMLReport(mtr.results); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Generate Markdown report
	if err := mtr.reportGenerator.GenerateMarkdownReport(mtr.results); err != nil {
		return fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	mtr.logger.Info("Manual test reports generated successfully")
	return nil
}

// GetResults returns the manual test results
func (mtr *ManualTestRunner) GetResults() *ManualTestResults {
	return mtr.results
}

// GetGuide returns the manual testing guide
func (mtr *ManualTestRunner) GetGuide() *ManualTestingGuide {
	return mtr.guide
}

// PrintSummary prints a summary of the manual test results
func (mtr *ManualTestRunner) PrintSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("MANUAL TEST SUITE SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Test Session ID: %s\n", mtr.results.TestSessionID)
	fmt.Printf("Tester: %s\n", mtr.results.TesterName)
	fmt.Printf("Environment: %s\n", mtr.results.TestEnvironment)
	fmt.Printf("Start Time: %s\n", mtr.results.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n", mtr.results.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Duration: %s\n", mtr.results.TotalDuration)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Scenarios: %d\n", mtr.results.TotalScenarios)
	fmt.Printf("Passed Scenarios: %d\n", mtr.results.PassedScenarios)
	fmt.Printf("Failed Scenarios: %d\n", mtr.results.FailedScenarios)
	fmt.Printf("Skipped Scenarios: %d\n", mtr.results.SkippedScenarios)
	fmt.Printf("Pass Rate: %.2f%%\n", mtr.results.PassRate)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Issues: %d\n", len(mtr.results.IssuesFound))
	fmt.Printf("Critical Issues: %d\n", mtr.countIssuesBySeverity("Critical"))
	fmt.Printf("High Issues: %d\n", mtr.countIssuesBySeverity("High"))
	fmt.Printf("Medium Issues: %d\n", mtr.countIssuesBySeverity("Medium"))
	fmt.Printf("Low Issues: %d\n", mtr.countIssuesBySeverity("Low"))
	fmt.Println(strings.Repeat("-", 80))

	// Print scenario results
	fmt.Println("SCENARIO RESULTS:")
	for scenarioID, result := range mtr.results.ScenarioResults {
		status := "✅ PASSED"
		if result.Status == "Failed" {
			status = "❌ FAILED"
		} else if result.Status == "Skipped" {
			status = "⏭️ SKIPPED"
		}
		fmt.Printf("  %s: %s (%d steps, %s)\n",
			scenarioID, status, result.StepsExecuted, result.Duration)
	}

	// Print recommendations
	if len(mtr.results.Recommendations) > 0 {
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println("RECOMMENDATIONS:")
		for i, recommendation := range mtr.results.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, recommendation)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Reports generated in: %s\n", mtr.config.ReportOutputPath)
	fmt.Println(strings.Repeat("=", 80))
}
