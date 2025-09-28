package automation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestRunner provides automated test execution and reporting
type TestRunner struct {
	config     *TestRunnerConfig
	results    *TestResults
	httpClient *http.Client
}

// TestRunnerConfig contains test runner configuration
type TestRunnerConfig struct {
	BaseURL           string
	TestSuites        []TestSuiteConfig
	OutputDir         string
	ReportFormats     []string
	ParallelExecution bool
	MaxConcurrency    int
	Timeout           time.Duration
	RetryCount        int
	Environment       string
}

// TestSuiteConfig contains configuration for a test suite
type TestSuiteConfig struct {
	Name          string
	Package       string
	Type          string // "unit", "integration", "e2e", "performance", "security"
	Priority      string // "critical", "high", "medium", "low"
	Timeout       time.Duration
	RetryCount    int
	SkipOnFailure bool
	Environment   string
}

// TestResults contains aggregated test results
type TestResults struct {
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Duration      time.Duration     `json:"duration"`
	TotalSuites   int               `json:"total_suites"`
	PassedSuites  int               `json:"passed_suites"`
	FailedSuites  int               `json:"failed_suites"`
	SkippedSuites int               `json:"skipped_suites"`
	TotalTests    int               `json:"total_tests"`
	PassedTests   int               `json:"passed_tests"`
	FailedTests   int               `json:"failed_tests"`
	SkippedTests  int               `json:"skipped_tests"`
	Coverage      float64           `json:"coverage"`
	Suites        []TestSuiteResult `json:"suites"`
	Environment   string            `json:"environment"`
	Version       string            `json:"version"`
}

// TestSuiteResult contains results for a test suite
type TestSuiteResult struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Priority     string        `json:"priority"`
	Status       string        `json:"status"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	Coverage     float64       `json:"coverage"`
	Output       string        `json:"output"`
	Error        string        `json:"error,omitempty"`
	RetryCount   int           `json:"retry_count"`
}

// TestReport contains test report data
type TestReport struct {
	Results         *TestResults `json:"results"`
	Summary         string       `json:"summary"`
	Recommendations []string     `json:"recommendations"`
	Artifacts       []string     `json:"artifacts"`
	GeneratedAt     time.Time    `json:"generated_at"`
}

// NewTestRunner creates a new test runner
func NewTestRunner(config *TestRunnerConfig) *TestRunner {
	return &TestRunner{
		config: config,
		results: &TestResults{
			StartTime:   time.Now(),
			Environment: config.Environment,
			Version:     getVersion(),
			Suites:      make([]TestSuiteResult, 0),
		},
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// RunAllTests executes all configured test suites
func (tr *TestRunner) RunAllTests() error {
	log.Printf("Starting test execution for environment: %s", tr.config.Environment)
	log.Printf("Total test suites: %d", len(tr.config.TestSuites))

	// Create output directory
	if err := os.MkdirAll(tr.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Execute test suites
	for _, suiteConfig := range tr.config.TestSuites {
		result := tr.runTestSuite(suiteConfig)
		tr.results.Suites = append(tr.results.Suites, result)

		// Update aggregate results
		tr.results.TotalSuites++
		tr.results.TotalTests += result.TotalTests
		tr.results.PassedTests += result.PassedTests
		tr.results.FailedTests += result.FailedTests
		tr.results.SkippedTests += result.SkippedTests

		switch result.Status {
		case "passed":
			tr.results.PassedSuites++
		case "failed":
			tr.results.FailedSuites++
		case "skipped":
			tr.results.SkippedSuites++
		}

		// Log suite result
		log.Printf("Test suite '%s' completed: %s (%d/%d tests passed)",
			result.Name, result.Status, result.PassedTests, result.TotalTests)
	}

	// Calculate final results
	tr.results.EndTime = time.Now()
	tr.results.Duration = tr.results.EndTime.Sub(tr.results.StartTime)

	if tr.results.TotalTests > 0 {
		tr.results.Coverage = float64(tr.results.PassedTests) / float64(tr.results.TotalTests) * 100
	}

	// Generate reports
	if err := tr.generateReports(); err != nil {
		log.Printf("Warning: Failed to generate reports: %v", err)
	}

	// Log final results
	tr.logFinalResults()

	return nil
}

// runTestSuite executes a single test suite
func (tr *TestRunner) runTestSuite(config TestSuiteConfig) TestSuiteResult {
	result := TestSuiteResult{
		Name:       config.Name,
		Type:       config.Type,
		Priority:   config.Priority,
		StartTime:  time.Now(),
		RetryCount: 0,
	}

	// Determine test command based on suite type
	var cmd *exec.Cmd
	switch config.Type {
	case "unit":
		cmd = tr.buildUnitTestCommand(config)
	case "integration":
		cmd = tr.buildIntegrationTestCommand(config)
	case "e2e":
		cmd = tr.buildE2ETestCommand(config)
	case "performance":
		cmd = tr.buildPerformanceTestCommand(config)
	case "security":
		cmd = tr.buildSecurityTestCommand(config)
	default:
		result.Status = "skipped"
		result.Error = fmt.Sprintf("Unknown test suite type: %s", config.Type)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Execute test with retries
	for attempt := 0; attempt <= config.RetryCount; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying test suite '%s' (attempt %d/%d)", config.Name, attempt+1, config.RetryCount+1)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		output, err := cmd.Output()
		result.Output = string(output)
		result.RetryCount = attempt

		if err != nil {
			result.Error = err.Error()
			if attempt < config.RetryCount {
				continue // Retry
			}
			result.Status = "failed"
		} else {
			result.Status = "passed"
			break // Success
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Parse test results from output
	tr.parseTestOutput(&result)

	return result
}

// buildUnitTestCommand builds command for unit tests
func (tr *TestRunner) buildUnitTestCommand(config TestSuiteConfig) *exec.Cmd {
	args := []string{"test", "-v", "-cover", "-coverprofile=coverage.out"}

	if config.Timeout > 0 {
		args = append(args, "-timeout", config.Timeout.String())
	}

	args = append(args, config.Package)

	return exec.Command("go", args...)
}

// buildIntegrationTestCommand builds command for integration tests
func (tr *TestRunner) buildIntegrationTestCommand(config TestSuiteConfig) *exec.Cmd {
	args := []string{"test", "-v", "-tags=integration", "-cover", "-coverprofile=coverage.out"}

	if config.Timeout > 0 {
		args = append(args, "-timeout", config.Timeout.String())
	}

	args = append(args, config.Package)

	return exec.Command("go", args...)
}

// buildE2ETestCommand builds command for E2E tests
func (tr *TestRunner) buildE2ETestCommand(config TestSuiteConfig) *exec.Cmd {
	args := []string{"test", "-v", "-tags=e2e", "-timeout", "30m"}

	if config.Timeout > 0 {
		args = append(args, "-timeout", config.Timeout.String())
	}

	args = append(args, config.Package)

	return exec.Command("go", args...)
}

// buildPerformanceTestCommand builds command for performance tests
func (tr *TestRunner) buildPerformanceTestCommand(config TestSuiteConfig) *exec.Cmd {
	args := []string{"test", "-v", "-tags=performance", "-bench=.", "-benchmem", "-timeout", "1h"}

	if config.Timeout > 0 {
		args = append(args, "-timeout", config.Timeout.String())
	}

	args = append(args, config.Package)

	return exec.Command("go", args...)
}

// buildSecurityTestCommand builds command for security tests
func (tr *TestRunner) buildSecurityTestCommand(config TestSuiteConfig) *exec.Cmd {
	args := []string{"test", "-v", "-tags=security", "-timeout", "30m"}

	if config.Timeout > 0 {
		args = append(args, "-timeout", config.Timeout.String())
	}

	args = append(args, config.Package)

	return exec.Command("go", args...)
}

// parseTestOutput parses test output to extract test counts
func (tr *TestRunner) parseTestOutput(result *TestSuiteResult) {
	output := result.Output

	// Parse Go test output
	if strings.Contains(output, "PASS") {
		result.Status = "passed"
	}

	// Extract test counts from output
	// This is a simplified parser - in practice, you'd want more robust parsing
	if strings.Contains(output, "ok") {
		// Extract coverage percentage if available
		if coverageStart := strings.Index(output, "coverage: "); coverageStart != -1 {
			coverageEnd := strings.Index(output[coverageStart:], "%")
			if coverageEnd != -1 {
				coverageStr := output[coverageStart+10 : coverageStart+coverageEnd]
				if coverage, err := fmt.Sscanf(coverageStr, "%f", &result.Coverage); err == nil && coverage == 1 {
					// Coverage parsed successfully
				}
			}
		}
	}

	// Count tests (simplified)
	result.TotalTests = strings.Count(output, "=== RUN")
	result.PassedTests = strings.Count(output, "--- PASS:")
	result.FailedTests = strings.Count(output, "--- FAIL:")
	result.SkippedTests = strings.Count(output, "--- SKIP:")
}

// generateReports generates test reports in configured formats
func (tr *TestRunner) generateReports() error {
	report := &TestReport{
		Results:         tr.results,
		Summary:         tr.generateSummary(),
		Recommendations: tr.generateRecommendations(),
		Artifacts:       make([]string, 0),
		GeneratedAt:     time.Now(),
	}

	for _, format := range tr.config.ReportFormats {
		switch format {
		case "json":
			if err := tr.generateJSONReport(report); err != nil {
				log.Printf("Failed to generate JSON report: %v", err)
			}
		case "html":
			if err := tr.generateHTMLReport(report); err != nil {
				log.Printf("Failed to generate HTML report: %v", err)
			}
		case "junit":
			if err := tr.generateJUnitReport(report); err != nil {
				log.Printf("Failed to generate JUnit report: %v", err)
			}
		case "coverage":
			if err := tr.generateCoverageReport(report); err != nil {
				log.Printf("Failed to generate coverage report: %v", err)
			}
		}
	}

	return nil
}

// generateJSONReport generates JSON test report
func (tr *TestRunner) generateJSONReport(report *TestReport) error {
	filename := filepath.Join(tr.config.OutputDir, "test-results.json")

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// generateHTMLReport generates HTML test report
func (tr *TestRunner) generateHTMLReport(report *TestReport) error {
	filename := filepath.Join(tr.config.OutputDir, "test-results.html")

	html := tr.generateHTMLContent(report)

	return os.WriteFile(filename, []byte(html), 0644)
}

// generateJUnitReport generates JUnit XML test report
func (tr *TestRunner) generateJUnitReport(report *TestReport) error {
	filename := filepath.Join(tr.config.OutputDir, "test-results.xml")

	xml := tr.generateJUnitXML(report)

	return os.WriteFile(filename, []byte(xml), 0644)
}

// generateCoverageReport generates coverage report
func (tr *TestRunner) generateCoverageReport(report *TestReport) error {
	// Run go tool cover to generate coverage report
	cmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", filepath.Join(tr.config.OutputDir, "coverage.html"))

	if err := cmd.Run(); err != nil {
		return err
	}

	report.Artifacts = append(report.Artifacts, "coverage.html")

	return nil
}

// generateSummary generates test summary
func (tr *TestRunner) generateSummary() string {
	results := tr.results

	summary := fmt.Sprintf("Test Execution Summary\n")
	summary += fmt.Sprintf("=====================\n")
	summary += fmt.Sprintf("Environment: %s\n", results.Environment)
	summary += fmt.Sprintf("Version: %s\n", results.Version)
	summary += fmt.Sprintf("Duration: %v\n", results.Duration)
	summary += fmt.Sprintf("Total Suites: %d\n", results.TotalSuites)
	summary += fmt.Sprintf("Passed Suites: %d\n", results.PassedSuites)
	summary += fmt.Sprintf("Failed Suites: %d\n", results.FailedSuites)
	summary += fmt.Sprintf("Skipped Suites: %d\n", results.SkippedSuites)
	summary += fmt.Sprintf("Total Tests: %d\n", results.TotalTests)
	summary += fmt.Sprintf("Passed Tests: %d\n", results.PassedTests)
	summary += fmt.Sprintf("Failed Tests: %d\n", results.FailedTests)
	summary += fmt.Sprintf("Skipped Tests: %d\n", results.SkippedTests)
	summary += fmt.Sprintf("Coverage: %.2f%%\n", results.Coverage)

	return summary
}

// generateRecommendations generates test recommendations
func (tr *TestRunner) generateRecommendations() []string {
	recommendations := make([]string, 0)

	if tr.results.FailedTests > 0 {
		recommendations = append(recommendations, "Review and fix failed tests")
	}

	if tr.results.Coverage < 80 {
		recommendations = append(recommendations, "Increase test coverage to at least 80%")
	}

	if tr.results.FailedSuites > 0 {
		recommendations = append(recommendations, "Investigate failed test suites")
	}

	if tr.results.SkippedTests > 0 {
		recommendations = append(recommendations, "Review skipped tests and implement missing tests")
	}

	return recommendations
}

// generateHTMLContent generates HTML content for test report
func (tr *TestRunner) generateHTMLContent(report *TestReport) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Test Results Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .suite { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 5px; }
        .passed { background-color: #d4edda; }
        .failed { background-color: #f8d7da; }
        .skipped { background-color: #fff3cd; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
        .metric { text-align: center; padding: 20px; background-color: #f8f9fa; border-radius: 5px; }
        .metric-value { font-size: 2em; font-weight: bold; }
        .metric-label { color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Test Results Report</h1>
        <p>Generated at: ` + report.GeneratedAt.Format("2006-01-02 15:04:05") + `</p>
        <p>Environment: ` + report.Results.Environment + `</p>
        <p>Version: ` + report.Results.Version + `</p>
    </div>
    
    <div class="metrics">
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Results.TotalTests) + `</div>
            <div class="metric-label">Total Tests</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Results.PassedTests) + `</div>
            <div class="metric-label">Passed</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Results.FailedTests) + `</div>
            <div class="metric-label">Failed</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%.1f%%", report.Results.Coverage) + `</div>
            <div class="metric-label">Coverage</div>
        </div>
    </div>
    
    <div class="summary">
        <h2>Test Suites</h2>`

	for _, suite := range report.Results.Suites {
		statusClass := strings.ToLower(suite.Status)
		html += fmt.Sprintf(`
        <div class="suite %s">
            <h3>%s (%s)</h3>
            <p>Status: %s | Duration: %v | Tests: %d/%d passed</p>
            <p>Coverage: %.2f%%</p>`,
			statusClass, suite.Name, suite.Type, suite.Status, suite.Duration,
			suite.PassedTests, suite.TotalTests, suite.Coverage)

		if suite.Error != "" {
			html += fmt.Sprintf(`<p>Error: %s</p>`, suite.Error)
		}

		html += `</div>`
	}

	html += `
    </div>
    
    <div class="summary">
        <h2>Recommendations</h2>
        <ul>`

	for _, rec := range report.Recommendations {
		html += fmt.Sprintf(`<li>%s</li>`, rec)
	}

	html += `
        </ul>
    </div>
</body>
</html>`

	return html
}

// generateJUnitXML generates JUnit XML content
func (tr *TestRunner) generateJUnitXML(report *TestReport) string {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="KYB Platform Tests" tests="` + fmt.Sprintf("%d", report.Results.TotalTests) + `" failures="` + fmt.Sprintf("%d", report.Results.FailedTests) + `" skipped="` + fmt.Sprintf("%d", report.Results.SkippedTests) + `" time="` + fmt.Sprintf("%.3f", report.Results.Duration.Seconds()) + `">`

	for _, suite := range report.Results.Suites {
		xml += fmt.Sprintf(`
    <testsuite name="%s" tests="%d" failures="%d" skipped="%d" time="%.3f">`,
			suite.Name, suite.TotalTests, suite.FailedTests, suite.SkippedTests, suite.Duration.Seconds())

		if suite.Error != "" {
			xml += fmt.Sprintf(`
        <error message="%s"></error>`, suite.Error)
		}

		xml += `
    </testsuite>`
	}

	xml += `
</testsuites>`

	return xml
}

// logFinalResults logs final test results
func (tr *TestRunner) logFinalResults() {
	results := tr.results

	log.Printf("=== TEST EXECUTION COMPLETED ===")
	log.Printf("Environment: %s", results.Environment)
	log.Printf("Duration: %v", results.Duration)
	log.Printf("Total Suites: %d (Passed: %d, Failed: %d, Skipped: %d)",
		results.TotalSuites, results.PassedSuites, results.FailedSuites, results.SkippedSuites)
	log.Printf("Total Tests: %d (Passed: %d, Failed: %d, Skipped: %d)",
		results.TotalTests, results.PassedTests, results.FailedTests, results.SkippedTests)
	log.Printf("Coverage: %.2f%%", results.Coverage)

	if results.FailedTests > 0 {
		log.Printf("❌ %d tests failed", results.FailedTests)
	} else {
		log.Printf("✅ All tests passed")
	}
}

// getVersion returns the current version
func getVersion() string {
	// In a real implementation, this would read from version file or git
	return "1.0.0"
}

// LoadTestRunnerConfig loads test runner configuration from file
func LoadTestRunnerConfig(filename string) (*TestRunnerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config TestRunnerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveTestRunnerConfig saves test runner configuration to file
func SaveTestRunnerConfig(config *TestRunnerConfig, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
