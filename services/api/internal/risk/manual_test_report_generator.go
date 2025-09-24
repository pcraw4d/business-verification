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

// GenerateJSONReport generates a JSON report for manual testing
func (mtrg *ManualTestReportGenerator) GenerateJSONReport(results *ManualTestResults) error {
	mtrg.logger.Info("Generating JSON manual test report")

	reportPath := filepath.Join(mtrg.config.ReportOutputPath, "manual-test-report.json")

	// Create the report data
	reportData := map[string]interface{}{
		"test_session_id":   results.TestSessionID,
		"start_time":        results.StartTime,
		"end_time":          results.EndTime,
		"total_duration":    results.TotalDuration,
		"tester_name":       results.TesterName,
		"test_environment":  results.TestEnvironment,
		"total_scenarios":   results.TotalScenarios,
		"passed_scenarios":  results.PassedScenarios,
		"failed_scenarios":  results.FailedScenarios,
		"skipped_scenarios": results.SkippedScenarios,
		"pass_rate":         results.PassRate,
		"scenario_results":  results.ScenarioResults,
		"issues_found":      results.IssuesFound,
		"recommendations":   results.Recommendations,
		"summary":           results.Summary,
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

	mtrg.logger.Info("JSON manual test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateHTMLReport generates an HTML report for manual testing
func (mtrg *ManualTestReportGenerator) GenerateHTMLReport(results *ManualTestResults) error {
	mtrg.logger.Info("Generating HTML manual test report")

	reportPath := filepath.Join(mtrg.config.ReportOutputPath, "manual-test-report.html")

	// Parse the template
	tmpl, err := template.New("html-report").Parse(ManualTestHTMLTemplate)
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

	mtrg.logger.Info("HTML manual test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateMarkdownReport generates a Markdown report for manual testing
func (mtrg *ManualTestReportGenerator) GenerateMarkdownReport(results *ManualTestResults) error {
	mtrg.logger.Info("Generating Markdown manual test report")

	reportPath := filepath.Join(mtrg.config.ReportOutputPath, "manual-test-report.md")

	// Generate Markdown content
	markdown := mtrg.generateMarkdownContent(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	mtrg.logger.Info("Markdown manual test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateWorkflowReport generates a report for a specific workflow test
func (mtrg *ManualTestReportGenerator) GenerateWorkflowReport(results *ManualTestResults) error {
	mtrg.logger.Info("Generating workflow test report")

	reportPath := filepath.Join(mtrg.config.ReportOutputPath, fmt.Sprintf("workflow-report-%s.json", results.TestSessionID))

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write workflow report: %w", err)
	}

	mtrg.logger.Info("Workflow test report generated", zap.String("path", reportPath))
	return nil
}

// GenerateScenarioReport generates a report for a specific test scenario
func (mtrg *ManualTestReportGenerator) GenerateScenarioReport(result *ScenarioResult) error {
	mtrg.logger.Info("Generating scenario test report", zap.String("scenario_id", result.ScenarioID))

	reportPath := filepath.Join(mtrg.config.ReportOutputPath, fmt.Sprintf("scenario-report-%s.json", result.ScenarioID))

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenario report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write scenario report: %w", err)
	}

	mtrg.logger.Info("Scenario test report generated", zap.String("path", reportPath))
	return nil
}

// generateMarkdownContent generates the Markdown content for the report
func (mtrg *ManualTestReportGenerator) generateMarkdownContent(results *ManualTestResults) string {
	markdown := fmt.Sprintf(`# Manual Test Report

## Test Session Information

**Session ID:** %s  
**Tester:** %s  
**Environment:** %s  
**Start Time:** %s  
**End Time:** %s  
**Total Duration:** %s

## Summary

| Metric | Value |
|--------|-------|
| Total Scenarios | %d |
| Passed Scenarios | %d |
| Failed Scenarios | %d |
| Skipped Scenarios | %d |
| Pass Rate | %.2f%% |
| Total Issues | %d |

## Scenario Results

`,
		results.TestSessionID,
		results.TesterName,
		results.TestEnvironment,
		results.StartTime.Format("2006-01-02 15:04:05"),
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.TotalDuration,
		results.TotalScenarios,
		results.PassedScenarios,
		results.FailedScenarios,
		results.SkippedScenarios,
		results.PassRate,
		len(results.IssuesFound))

	// Add scenario results
	for scenarioID, result := range results.ScenarioResults {
		status := "✅ PASSED"
		if result.Status == "Failed" {
			status = "❌ FAILED"
		} else if result.Status == "Skipped" {
			status = "⏭️ SKIPPED"
		}

		markdown += fmt.Sprintf(`### %s %s

**Scenario:** %s  
**Duration:** %s  
**Steps Executed:** %d (Passed: %d, Failed: %d)  
**Issues Found:** %d

`,
			scenarioID,
			status,
			result.ScenarioName,
			result.Duration,
			result.StepsExecuted,
			result.StepsPassed,
			result.StepsFailed,
			len(result.IssuesFound))

		if len(result.IssuesFound) > 0 {
			markdown += "**Issues:**\n"
			for _, issue := range result.IssuesFound {
				markdown += fmt.Sprintf("- **%s:** %s (Severity: %s)\n",
					issue.Title, issue.Description, issue.Severity)
			}
			markdown += "\n"
		}
	}

	// Add issues summary
	if len(results.IssuesFound) > 0 {
		markdown += `## Issues Summary

`
		for _, issue := range results.IssuesFound {
			markdown += fmt.Sprintf(`### %s

**Severity:** %s  
**Priority:** %s  
**Category:** %s  
**Scenario:** %s  
**Step:** %d  
**Description:** %s  
**Expected Result:** %s  
**Actual Result:** %s  
**Reported By:** %s  
**Reported At:** %s  
**Status:** %s

`,
				issue.Title,
				issue.Severity,
				issue.Priority,
				issue.Category,
				issue.ScenarioID,
				issue.StepNumber,
				issue.Description,
				issue.ExpectedResult,
				issue.ActualResult,
				issue.ReportedBy,
				issue.ReportedAt.Format("2006-01-02 15:04:05"),
				issue.Status)
		}
	}

	// Add recommendations
	if len(results.Recommendations) > 0 {
		markdown += `## Recommendations

`
		for i, recommendation := range results.Recommendations {
			markdown += fmt.Sprintf("%d. %s\n", i+1, recommendation)
		}
		markdown += "\n"
	}

	markdown += fmt.Sprintf(`## Detailed Summary

`)
	for key, value := range results.Summary {
		markdown += fmt.Sprintf("- **%s:** %v\n", key, value)
	}

	markdown += fmt.Sprintf(`

---
*Report generated by Manual Test Suite on %s*
`,
		time.Now().Format("2006-01-02 15:04:05"))

	return markdown
}

// ManualTestHTMLTemplate is the HTML template for manual test reports
const ManualTestHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Manual Test Report - {{.TestSessionID}}</title>
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
        .scenarios { margin-bottom: 30px; }
        .scenario { background-color: #f8f9fa; margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .scenario.passed { border-left-color: #28a745; }
        .scenario.failed { border-left-color: #dc3545; }
        .scenario.skipped { border-left-color: #ffc107; }
        .scenario h3 { margin: 0 0 15px 0; color: #333; }
        .scenario .meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 15px; }
        .scenario .meta-item { background-color: white; padding: 10px; border-radius: 4px; }
        .scenario .meta-item strong { color: #666; }
        .issues { background-color: #fff3cd; border: 1px solid #ffeaa7; border-radius: 8px; padding: 20px; margin-bottom: 30px; }
        .issues h3 { margin: 0 0 15px 0; color: #856404; }
        .issue { background-color: white; padding: 15px; margin-bottom: 10px; border-radius: 4px; border-left: 4px solid #dc3545; }
        .issue.critical { border-left-color: #dc3545; }
        .issue.high { border-left-color: #fd7e14; }
        .issue.medium { border-left-color: #ffc107; }
        .issue.low { border-left-color: #6c757d; }
        .recommendations { background-color: #d1ecf1; border: 1px solid #bee5eb; border-radius: 8px; padding: 20px; margin-bottom: 30px; }
        .recommendations h3 { margin: 0 0 15px 0; color: #0c5460; }
        .recommendations ul { margin: 0; padding-left: 20px; }
        .recommendations li { margin-bottom: 5px; color: #0c5460; }
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
            <h1>Manual Test Report</h1>
            <h2>Session: {{.TestSessionID}}</h2>
            <p>Generated on {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
            <p><strong>Tester:</strong> {{.TesterName}} | <strong>Environment:</strong> {{.TestEnvironment}}</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <h3>Total Scenarios</h3>
                <div class="number">{{.TotalScenarios}}</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="number">{{.PassedScenarios}}</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="number">{{.FailedScenarios}}</div>
            </div>
            <div class="summary-card skipped">
                <h3>Skipped</h3>
                <div class="number">{{.SkippedScenarios}}</div>
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

        {{if .IssuesFound}}
        <div class="issues">
            <h3>Issues Found ({{len .IssuesFound}})</h3>
            {{range .IssuesFound}}
            <div class="issue {{.Severity | lower}}">
                <h4>{{.Title}} <span class="status-badge status-{{.Severity | lower}}">{{.Severity}}</span></h4>
                <p><strong>Scenario:</strong> {{.ScenarioID}} | <strong>Step:</strong> {{.StepNumber}}</p>
                <p><strong>Description:</strong> {{.Description}}</p>
                <p><strong>Expected:</strong> {{.ExpectedResult}}</p>
                <p><strong>Actual:</strong> {{.ActualResult}}</p>
                <p><strong>Reported:</strong> {{.ReportedAt.Format "2006-01-02 15:04:05"}} by {{.ReportedBy}}</p>
            </div>
            {{end}}
        </div>
        {{end}}

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

        <div class="scenarios">
            <h2>Scenario Results</h2>
            {{range $id, $result := .ScenarioResults}}
            <div class="scenario {{if eq $result.Status "Passed"}}passed{{else if eq $result.Status "Failed"}}failed{{else}}skipped{{end}}">
                <h3>{{$result.ScenarioName}} <span class="status-badge status-{{if eq $result.Status "Passed"}}passed{{else if eq $result.Status "Failed"}}failed{{else}}skipped{{end}}">{{$result.Status}}</span></h3>
                <div class="meta">
                    <div class="meta-item">
                        <strong>Scenario ID:</strong> {{$result.ScenarioID}}
                    </div>
                    <div class="meta-item">
                        <strong>Duration:</strong> {{$result.Duration}}
                    </div>
                    <div class="meta-item">
                        <strong>Steps Executed:</strong> {{$result.StepsExecuted}}
                    </div>
                    <div class="meta-item">
                        <strong>Steps Passed:</strong> {{$result.StepsPassed}}
                    </div>
                    <div class="meta-item">
                        <strong>Steps Failed:</strong> {{$result.StepsFailed}}
                    </div>
                    <div class="meta-item">
                        <strong>Issues:</strong> {{len $result.IssuesFound}}
                    </div>
                </div>
                {{if $result.Notes}}
                <div style="background-color: white; padding: 10px; border-radius: 4px; margin-top: 10px;">
                    <strong>Notes:</strong> {{$result.Notes}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p>Report generated by Manual Test Suite</p>
            <p>Execution time: {{.StartTime.Format "2006-01-02 15:04:05"}} - {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>
`
