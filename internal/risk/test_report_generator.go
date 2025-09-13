package risk

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// TestReportGenerator generates comprehensive test reports in multiple formats
type TestReportGenerator struct {
	logger *zap.Logger
	config *IntegrationTestConfig
}

// HTMLReportTemplate is the HTML template for test reports
const HTMLReportTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Integration Test Report - {{.SuiteName}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 2px solid #e0e0e0; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .summary-card { background-color: #f8f9fa; padding: 20px; border-radius: 8px; text-align: center; border-left: 4px solid #007bff; }
        .summary-card.passed { border-left-color: #28a745; }
        .summary-card.failed { border-left-color: #dc3545; }
        .summary-card.skipped { border-left-color: #ffc107; }
        .summary-card h3 { margin: 0 0 10px 0; color: #333; }
        .summary-card .number { font-size: 2em; font-weight: bold; margin: 10px 0; }
        .passed .number { color: #28a745; }
        .failed .number { color: #dc3545; }
        .skipped .number { color: #ffc107; }
        .test-suites { margin-bottom: 30px; }
        .test-suite { background-color: #f8f9fa; margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .test-suite.passed { border-left-color: #28a745; }
        .test-suite.failed { border-left-color: #dc3545; }
        .test-suite h3 { margin: 0 0 15px 0; color: #333; }
        .test-suite .meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 15px; }
        .test-suite .meta-item { background-color: white; padding: 10px; border-radius: 4px; }
        .test-suite .meta-item strong { color: #666; }
        .recommendations { background-color: #fff3cd; border: 1px solid #ffeaa7; border-radius: 8px; padding: 20px; margin-bottom: 30px; }
        .recommendations h3 { margin: 0 0 15px 0; color: #856404; }
        .recommendations ul { margin: 0; padding-left: 20px; }
        .recommendations li { margin-bottom: 5px; color: #856404; }
        .footer { text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; color: #666; }
        .status-badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 0.8em; font-weight: bold; text-transform: uppercase; }
        .status-passed { background-color: #d4edda; color: #155724; }
        .status-failed { background-color: #f8d7da; color: #721c24; }
        .status-skipped { background-color: #fff3cd; color: #856404; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Integration Test Report</h1>
            <h2>{{.SuiteName}}</h2>
            <p>Generated on {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <h3>Total Tests</h3>
                <div class="number">{{.TotalTests}}</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="number">{{.PassedTests}}</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="number">{{.FailedTests}}</div>
            </div>
            <div class="summary-card skipped">
                <h3>Skipped</h3>
                <div class="number">{{.SkippedTests}}</div>
            </div>
            <div class="summary-card">
                <h3>Pass Rate</h3>
                <div class="number">{{printf "%.1f" .PassRate}}%</div>
            </div>
            <div class="summary-card">
                <h3>Duration</h3>
                <div class="number">{{.TotalDuration}}</div>
            </div>
        </div>

        {{if .Recommendations}}
        <div class="recommendations">
            <h3>Recommendations</h3>
            <ul>
                {{range .Recommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}

        <div class="test-suites">
            <h2>Test Suite Results</h2>
            {{range $name, $result := .TestSuiteResults}}
            <div class="test-suite {{if eq $result.Status "PASSED"}}passed{{else}}failed{{end}}">
                <h3>{{$result.SuiteName}} <span class="status-badge status-{{if eq $result.Status "PASSED"}}passed{{else}}failed{{end}}">{{$result.Status}}</span></h3>
                <div class="meta">
                    <div class="meta-item">
                        <strong>Category:</strong> {{$result.Category}}
                    </div>
                    <div class="meta-item">
                        <strong>Duration:</strong> {{$result.Duration}}
                    </div>
                    <div class="meta-item">
                        <strong>Tests:</strong> {{$result.TotalTests}}
                    </div>
                    <div class="meta-item">
                        <strong>Passed:</strong> {{$result.PassedTests}}
                    </div>
                    <div class="meta-item">
                        <strong>Failed:</strong> {{$result.FailedTests}}
                    </div>
                    <div class="meta-item">
                        <strong>Pass Rate:</strong> {{printf "%.1f" $result.PassRate}}%
                    </div>
                </div>
                {{if $result.ErrorMessage}}
                <div style="background-color: #f8d7da; color: #721c24; padding: 10px; border-radius: 4px; margin-top: 10px;">
                    <strong>Error:</strong> {{$result.ErrorMessage}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p>Report generated by Automated Integration Test Suite</p>
            <p>Execution time: {{.StartTime.Format "2006-01-02 15:04:05"}} - {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>
`

// GenerateJSONReport generates a JSON test report
func (trg *TestReportGenerator) GenerateJSONReport(results *IntegrationTestResults) error {
	trg.logger.Info("Generating JSON test report")

	reportPath := filepath.Join(trg.config.ReportOutputPath, "test-report.json")

	// Create the report data
	reportData := map[string]interface{}{
		"suite_name":         results.SuiteName,
		"start_time":         results.StartTime,
		"end_time":           results.EndTime,
		"total_duration":     results.TotalDuration,
		"total_tests":        results.TotalTests,
		"passed_tests":       results.PassedTests,
		"failed_tests":       results.FailedTests,
		"skipped_tests":      results.SkippedTests,
		"pass_rate":          results.PassRate,
		"test_suite_results": results.TestSuiteResults,
		"overall_summary":    results.OverallSummary,
		"recommendations":    results.Recommendations,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	trg.logger.Info("JSON test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateHTMLReport generates an HTML test report
func (trg *TestReportGenerator) GenerateHTMLReport(results *IntegrationTestResults) error {
	trg.logger.Info("Generating HTML test report")

	reportPath := filepath.Join(trg.config.ReportOutputPath, "test-report.html")

	// Parse the template
	tmpl, err := template.New("html-report").Parse(HTMLReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Create the report file
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML report file: %w", err)
	}
	defer file.Close()

	// Execute the template
	if err := tmpl.Execute(file, results); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	trg.logger.Info("HTML test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateMarkdownReport generates a Markdown test report
func (trg *TestReportGenerator) GenerateMarkdownReport(results *IntegrationTestResults) error {
	trg.logger.Info("Generating Markdown test report")

	reportPath := filepath.Join(trg.config.ReportOutputPath, "test-report.md")

	// Generate Markdown content
	markdown := trg.generateMarkdownContent(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	trg.logger.Info("Markdown test report generated", zap.String("path", reportPath))
	return nil
}

// generateMarkdownContent generates the Markdown content for the report
func (trg *TestReportGenerator) generateMarkdownContent(results *IntegrationTestResults) string {
	markdown := fmt.Sprintf(`# Integration Test Report

## %s

**Generated on:** %s  
**Execution Time:** %s - %s  
**Total Duration:** %s

## Summary

| Metric | Value |
|--------|-------|
| Total Tests | %d |
| Passed Tests | %d |
| Failed Tests | %d |
| Skipped Tests | %d |
| Pass Rate | %.2f%% |

## Test Suite Results

`,
		results.SuiteName,
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.StartTime.Format("2006-01-02 15:04:05"),
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.TotalDuration,
		results.TotalTests,
		results.PassedTests,
		results.FailedTests,
		results.SkippedTests,
		results.PassRate)

	// Add test suite results
	for name, result := range results.TestSuiteResults {
		status := "✅ PASSED"
		if result.Status == "FAILED" {
			status = "❌ FAILED"
		}

		markdown += fmt.Sprintf(`### %s %s

**Category:** %s  
**Duration:** %s  
**Tests:** %d (Passed: %d, Failed: %d, Skipped: %d)  
**Pass Rate:** %.2f%%

`,
			name,
			status,
			result.Category,
			result.Duration,
			result.TotalTests,
			result.PassedTests,
			result.FailedTests,
			result.SkippedTests,
			result.PassRate)

		if result.ErrorMessage != "" {
			markdown += fmt.Sprintf(`**Error:** %s

`,
				result.ErrorMessage)
		}
	}

	// Add recommendations
	if len(results.Recommendations) > 0 {
		markdown += `## Recommendations

`
		for _, recommendation := range results.Recommendations {
			markdown += fmt.Sprintf("- %s\n", recommendation)
		}
		markdown += "\n"
	}

	// Add overall summary
	markdown += `## Overall Summary

`
	for key, value := range results.OverallSummary {
		markdown += fmt.Sprintf("- **%s:** %v\n", key, value)
	}

	markdown += fmt.Sprintf(`

---
*Report generated by Automated Integration Test Suite on %s*
`,
		time.Now().Format("2006-01-02 15:04:05"))

	return markdown
}

// GenerateSummaryReport generates a summary report for CI/CD integration
func (trg *TestReportGenerator) GenerateSummaryReport(results *IntegrationTestResults) error {
	trg.logger.Info("Generating summary test report")

	reportPath := filepath.Join(trg.config.ReportOutputPath, "test-summary.json")

	// Create a simplified summary for CI/CD
	summary := map[string]interface{}{
		"suite_name":    results.SuiteName,
		"status":        trg.determineOverallStatus(results),
		"total_tests":   results.TotalTests,
		"passed_tests":  results.PassedTests,
		"failed_tests":  results.FailedTests,
		"skipped_tests": results.SkippedTests,
		"pass_rate":     results.PassRate,
		"duration":      results.TotalDuration.String(),
		"timestamp":     results.EndTime,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write summary report: %w", err)
	}

	trg.logger.Info("Summary test report generated", zap.String("path", reportPath))
	return nil
}

// determineOverallStatus determines the overall status of the test suite
func (trg *TestReportGenerator) determineOverallStatus(results *IntegrationTestResults) string {
	if results.FailedTests > 0 {
		return "FAILED"
	}
	if results.SkippedTests > 0 && results.PassedTests == 0 {
		return "SKIPPED"
	}
	return "PASSED"
}

// GenerateJUnitReport generates a JUnit XML report for CI/CD integration
func (trg *TestReportGenerator) GenerateJUnitReport(results *IntegrationTestResults) error {
	trg.logger.Info("Generating JUnit XML test report")

	reportPath := filepath.Join(trg.config.ReportOutputPath, "test-results.xml")

	// Generate JUnit XML content
	xml := trg.generateJUnitXML(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(xml), 0644); err != nil {
		return fmt.Errorf("failed to write JUnit report: %w", err)
	}

	trg.logger.Info("JUnit XML test report generated", zap.String("path", reportPath))
	return nil
}

// generateJUnitXML generates the JUnit XML content
func (trg *TestReportGenerator) generateJUnitXML(results *IntegrationTestResults) string {
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="%s" tests="%d" failures="%d" skipped="%d" time="%.3f" timestamp="%s">
`,
		results.SuiteName,
		results.TotalTests,
		results.FailedTests,
		results.SkippedTests,
		results.TotalDuration.Seconds(),
		results.StartTime.Format("2006-01-02T15:04:05"))

	// Add test cases for each test suite
	for name, result := range results.TestSuiteResults {
		testCaseName := fmt.Sprintf("%s.%s", result.Category, name)

		if result.Status == "FAILED" {
			xml += fmt.Sprintf(`  <testcase classname="%s" name="%s" time="%.3f">
    <failure message="Test suite failed">%s</failure>
  </testcase>
`,
				result.Category,
				testCaseName,
				result.Duration.Seconds(),
				result.ErrorMessage)
		} else if result.Status == "SKIPPED" {
			xml += fmt.Sprintf(`  <testcase classname="%s" name="%s" time="%.3f">
    <skipped/>
  </testcase>
`,
				result.Category,
				testCaseName,
				result.Duration.Seconds())
		} else {
			xml += fmt.Sprintf(`  <testcase classname="%s" name="%s" time="%.3f"/>
`,
				result.Category,
				testCaseName,
				result.Duration.Seconds())
		}
	}

	xml += `</testsuite>`
	return xml
}
