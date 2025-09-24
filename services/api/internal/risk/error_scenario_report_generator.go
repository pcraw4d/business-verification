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

// GenerateReports generates comprehensive error scenario reports
func (esrg *ErrorScenarioReportGenerator) GenerateReports(results *ErrorScenarioResults) error {
	esrg.logger.Info("Generating error scenario reports")

	// Create report output directory if it doesn't exist
	if err := os.MkdirAll(esrg.config.ReportOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create report output directory: %w", err)
	}

	// Generate JSON report
	if err := esrg.GenerateJSONReport(results); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate HTML report
	if err := esrg.GenerateHTMLReport(results); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Generate Markdown report
	if err := esrg.GenerateMarkdownReport(results); err != nil {
		return fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	esrg.logger.Info("Error scenario reports generated successfully")
	return nil
}

// GenerateJSONReport generates a JSON report for error scenario results
func (esrg *ErrorScenarioReportGenerator) GenerateJSONReport(results *ErrorScenarioResults) error {
	esrg.logger.Info("Generating JSON error scenario report")

	reportPath := filepath.Join(esrg.config.ReportOutputPath, "error-scenario-report.json")

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	esrg.logger.Info("JSON error scenario report generated", zap.String("path", reportPath))
	return nil
}

// GenerateHTMLReport generates an HTML report for error scenario results
func (esrg *ErrorScenarioReportGenerator) GenerateHTMLReport(results *ErrorScenarioResults) error {
	esrg.logger.Info("Generating HTML error scenario report")

	reportPath := filepath.Join(esrg.config.ReportOutputPath, "error-scenario-report.html")

	// Parse the template
	tmpl, err := template.New("html-report").Parse(ErrorScenarioHTMLTemplate)
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

	esrg.logger.Info("HTML error scenario report generated", zap.String("path", reportPath))
	return nil
}

// GenerateMarkdownReport generates a Markdown report for error scenario results
func (esrg *ErrorScenarioReportGenerator) GenerateMarkdownReport(results *ErrorScenarioResults) error {
	esrg.logger.Info("Generating Markdown error scenario report")

	reportPath := filepath.Join(esrg.config.ReportOutputPath, "error-scenario-report.md")

	// Generate Markdown content
	markdown := esrg.generateMarkdownContent(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	esrg.logger.Info("Markdown error scenario report generated", zap.String("path", reportPath))
	return nil
}

// generateMarkdownContent generates the Markdown content for the report
func (esrg *ErrorScenarioReportGenerator) generateMarkdownContent(results *ErrorScenarioResults) string {
	markdown := fmt.Sprintf(`# Error Scenario Test Report

## Test Session Information

**Session ID:** %s  
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

## Error Scenario Results

### Overall Performance
- **Overall Pass Rate:** %.2f%%
- **Critical Failures:** %d
- **High Severity Failures:** %d
- **Medium Severity Failures:** %d
- **Low Severity Failures:** %d
- **Recovery Success Rate:** %.2f%%
- **Average Recovery Time:** %s
- **Data Loss Incidents:** %d
- **Service Downtime:** %s

## Scenario Results

`,
		results.SessionID,
		results.Environment,
		results.StartTime.Format("2006-01-02 15:04:05"),
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.TotalDuration,
		results.TotalScenarios,
		results.PassedScenarios,
		results.FailedScenarios,
		results.SkippedScenarios,
		results.PassRate,
		results.Summary.OverallPassRate,
		results.Summary.CriticalFailures,
		results.Summary.HighSeverityFailures,
		results.Summary.MediumSeverityFailures,
		results.Summary.LowSeverityFailures,
		results.Summary.RecoverySuccessRate,
		results.Summary.AverageRecoveryTime,
		results.Summary.DataLossIncidents,
		results.Summary.ServiceDowntime)

	// Add scenario results
	for scenarioID, scenarioResults := range results.ScenarioResults {
		if len(scenarioResults) == 0 {
			continue
		}

		// Calculate averages for this scenario
		successCount := 0
		recoveryCount := 0
		totalRecoveryTime := time.Duration(0)
		dataLossCount := 0

		for _, result := range scenarioResults {
			if result.Success {
				successCount++
			}
			if result.RecoveryAttempted && result.RecoverySuccess {
				recoveryCount++
				totalRecoveryTime += result.RecoveryTime
			}
			if result.Impact != nil && result.Impact.DataLoss {
				dataLossCount++
			}
		}

		passRate := float64(successCount) / float64(len(scenarioResults)) * 100
		avgRecoveryTime := time.Duration(0)
		if recoveryCount > 0 {
			avgRecoveryTime = totalRecoveryTime / time.Duration(recoveryCount)
		}

		status := "✅ PASSED"
		if successCount == 0 {
			status = "❌ FAILED"
		}

		markdown += fmt.Sprintf(`### %s: %s (%d iterations)

**Results:** %d iterations  
**Pass Rate:** %.2f%%  
**Average Recovery Time:** %s  
**Data Loss Incidents:** %d

`,
			scenarioID, status, len(scenarioResults),
			len(scenarioResults),
			passRate,
			avgRecoveryTime,
			dataLossCount)
	}

	// Add category metrics
	if len(results.Summary.CategoryMetrics) > 0 {
		markdown += `## Category Performance

`
		for category, metrics := range results.Summary.CategoryMetrics {
			markdown += fmt.Sprintf(`### %s

- **Scenarios:** %d
- **Pass Rate:** %.2f%%
- **Failure Rate:** %.2f%%
- **Recovery Success Rate:** %.2f%%
- **Average Recovery Time:** %s
- **Critical Failures:** %d
- **Data Loss Incidents:** %d

`,
				category,
				metrics.ScenarioCount,
				metrics.PassRate,
				metrics.FailureRate,
				metrics.RecoverySuccessRate,
				metrics.AverageRecoveryTime,
				metrics.CriticalFailures,
				metrics.DataLossIncidents)
		}
	}

	// Add issues
	if len(results.Issues) > 0 {
		markdown += `## Issues Found

`
		for _, issue := range results.Issues {
			markdown += fmt.Sprintf(`### %s

**Severity:** %s  
**Category:** %s  
**Scenario:** %s  
**Description:** %s  
**Impact:** %s  
**Recommendation:** %s  
**Detected:** %s

`,
				issue.Title,
				issue.Severity,
				issue.Category,
				issue.ScenarioID,
				issue.Description,
				issue.Impact.BusinessImpact,
				issue.Recommendation,
				issue.DetectedAt.Format("2006-01-02 15:04:05"))
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

	markdown += fmt.Sprintf(`---
*Report generated by Error Scenario Testing Suite on %s*
`,
		time.Now().Format("2006-01-02 15:04:05"))

	return markdown
}

// ErrorScenarioHTMLTemplate is the HTML template for error scenario reports
const ErrorScenarioHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error Scenario Test Report - {{.SessionID}}</title>
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
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .metric-card { background-color: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #17a2b8; }
        .metric-card h4 { margin: 0 0 15px 0; color: #333; }
        .metric-value { font-size: 1.5em; font-weight: bold; color: #17a2b8; }
        .scenarios { margin-bottom: 30px; }
        .scenario { background-color: #f8f9fa; margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
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
            <h1>Error Scenario Test Report</h1>
            <h2>Session: {{.SessionID}}</h2>
            <p>Generated on {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
            <p><strong>Environment:</strong> {{.Environment}}</p>
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

        <div class="metrics">
            <div class="metric-card">
                <h4>Overall Pass Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallPassRate}}%</div>
            </div>
            <div class="metric-card">
                <h4>Critical Failures</h4>
                <div class="metric-value">{{.Summary.CriticalFailures}}</div>
            </div>
            <div class="metric-card">
                <h4>Recovery Success Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.RecoverySuccessRate}}%</div>
            </div>
            <div class="metric-card">
                <h4>Average Recovery Time</h4>
                <div class="metric-value">{{.Summary.AverageRecoveryTime}}</div>
            </div>
            <div class="metric-card">
                <h4>Data Loss Incidents</h4>
                <div class="metric-value">{{.Summary.DataLossIncidents}}</div>
            </div>
            <div class="metric-card">
                <h4>Service Downtime</h4>
                <div class="metric-value">{{.Summary.ServiceDowntime}}</div>
            </div>
        </div>

        {{if .Issues}}
        <div class="issues">
            <h3>Issues Found ({{len .Issues}})</h3>
            {{range .Issues}}
            <div class="issue {{.Severity | lower}}">
                <h4>{{.Title}} <span class="status-badge status-{{.Severity | lower}}">{{.Severity}}</span></h4>
                <p><strong>Scenario:</strong> {{.ScenarioID}} | <strong>Category:</strong> {{.Category}}</p>
                <p><strong>Description:</strong> {{.Description}}</p>
                <p><strong>Impact:</strong> {{.Impact.BusinessImpact}}</p>
                <p><strong>Recommendation:</strong> {{.Recommendation}}</p>
                <p><strong>Detected:</strong> {{.DetectedAt.Format "2006-01-02 15:04:05"}}</p>
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
            {{range $id, $results := .ScenarioResults}}
            <div class="scenario">
                <h3>{{$id}}</h3>
                <div class="meta">
                    <div class="meta-item">
                        <strong>Iterations:</strong> {{len $results}}
                    </div>
                    {{if $results}}
                    <div class="meta-item">
                        <strong>Success Rate:</strong> {{printf "%.2f" (index $results 0).Success}}%
                    </div>
                    <div class="meta-item">
                        <strong>Recovery Time:</strong> {{(index $results 0).RecoveryTime}}
                    </div>
                    <div class="meta-item">
                        <strong>Data Loss:</strong> {{if (index $results 0).Impact.DataLoss}}Yes{{else}}No{{end}}
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p>Report generated by Error Scenario Testing Suite</p>
            <p>Execution time: {{.StartTime.Format "2006-01-02 15:04:05"}} - {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>
`
