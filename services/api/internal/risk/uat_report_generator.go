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

// GenerateReports generates comprehensive UAT reports
func (uatrg *UATReportGenerator) GenerateReports(results *UATResults) error {
	uatrg.logger.Info("Generating UAT reports")

	// Create report output directory if it doesn't exist
	if err := os.MkdirAll(uatrg.config.ReportOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create report output directory: %w", err)
	}

	// Generate JSON report
	if err := uatrg.GenerateJSONReport(results); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate HTML report
	if err := uatrg.GenerateHTMLReport(results); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Generate Markdown report
	if err := uatrg.GenerateMarkdownReport(results); err != nil {
		return fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	uatrg.logger.Info("UAT reports generated successfully")
	return nil
}

// GenerateJSONReport generates a JSON report for UAT results
func (uatrg *UATReportGenerator) GenerateJSONReport(results *UATResults) error {
	uatrg.logger.Info("Generating JSON UAT report")

	reportPath := filepath.Join(uatrg.config.ReportOutputPath, "uat-report.json")

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	uatrg.logger.Info("JSON UAT report generated", zap.String("path", reportPath))
	return nil
}

// GenerateHTMLReport generates an HTML report for UAT results
func (uatrg *UATReportGenerator) GenerateHTMLReport(results *UATResults) error {
	uatrg.logger.Info("Generating HTML UAT report")

	reportPath := filepath.Join(uatrg.config.ReportOutputPath, "uat-report.html")

	// Parse the template
	tmpl, err := template.New("html-report").Parse(UATHTMLTemplate)
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

	uatrg.logger.Info("HTML UAT report generated", zap.String("path", reportPath))
	return nil
}

// GenerateMarkdownReport generates a Markdown report for UAT results
func (uatrg *UATReportGenerator) GenerateMarkdownReport(results *UATResults) error {
	uatrg.logger.Info("Generating Markdown UAT report")

	reportPath := filepath.Join(uatrg.config.ReportOutputPath, "uat-report.md")

	// Generate Markdown content
	markdown := uatrg.generateMarkdownContent(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	uatrg.logger.Info("Markdown UAT report generated", zap.String("path", reportPath))
	return nil
}

// generateMarkdownContent generates the Markdown content for the report
func (uatrg *UATReportGenerator) generateMarkdownContent(results *UATResults) string {
	markdown := fmt.Sprintf(`# User Acceptance Testing Report

## Test Session Information

**Session ID:** %s  
**Environment:** %s  
**Start Time:** %s  
**End Time:** %s  
**Total Duration:** %s

## Summary

| Metric | Value |
|--------|-------|
| Total Test Cases | %d |
| Passed Test Cases | %d |
| Failed Test Cases | %d |
| Skipped Test Cases | %d |
| Pass Rate | %.2f%% |

## User Experience Metrics

### Overall Performance
- **Overall Pass Rate:** %.2f%%
- **Overall User Satisfaction:** %.2f/10
- **Overall Usability Score:** %.2f/10
- **Average Completion Time:** %s
- **Average Error Rate:** %.2f%%
- **Recommendation Rate:** %.2f%%

## Test Case Results

`,
		results.SessionID,
		results.Environment,
		results.StartTime.Format("2006-01-02 15:04:05"),
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.TotalDuration,
		results.TotalTestCases,
		results.PassedTestCases,
		results.FailedTestCases,
		results.SkippedTestCases,
		results.PassRate,
		results.Summary.OverallPassRate,
		results.Summary.OverallUserSatisfaction,
		results.Summary.OverallUsabilityScore,
		results.Summary.AverageCompletionTime,
		results.Summary.AverageErrorRate,
		results.Summary.RecommendationRate)

	// Add test case results
	for testCaseID, testCaseResults := range results.TestCaseResults {
		if len(testCaseResults) == 0 {
			continue
		}

		// Calculate averages for this test case
		successCount := 0
		totalSatisfaction := 0.0
		totalUsability := 0.0
		totalCompletionTime := time.Duration(0)
		totalErrorRate := 0.0
		recommendationCount := 0

		for _, result := range testCaseResults {
			if result.Success {
				successCount++
			}
			if result.UserSatisfaction != nil {
				totalSatisfaction += result.UserSatisfaction.OverallRating
				if result.UserSatisfaction.WouldRecommend {
					recommendationCount++
				}
			}
			if result.UsabilityMetrics != nil {
				totalUsability += result.UsabilityMetrics.SatisfactionScore
				totalCompletionTime += result.UsabilityMetrics.TimeToComplete
				totalErrorRate += result.UsabilityMetrics.ErrorRate
			}
		}

		passRate := float64(successCount) / float64(len(testCaseResults)) * 100
		avgSatisfaction := totalSatisfaction / float64(len(testCaseResults))
		avgUsability := totalUsability / float64(len(testCaseResults))
		avgCompletionTime := totalCompletionTime / time.Duration(len(testCaseResults))
		avgErrorRate := totalErrorRate / float64(len(testCaseResults))
		recommendationRate := float64(recommendationCount) / float64(len(testCaseResults)) * 100

		status := "✅ PASSED"
		if successCount == 0 {
			status = "❌ FAILED"
		}

		markdown += fmt.Sprintf(`### %s: %s (%d users)

**Results:** %d users tested  
**Pass Rate:** %.2f%%  
**User Satisfaction:** %.2f/10  
**Usability Score:** %.2f/10  
**Average Completion Time:** %s  
**Error Rate:** %.2f%%  
**Recommendation Rate:** %.2f%%

`,
			testCaseID, status, len(testCaseResults),
			len(testCaseResults),
			passRate,
			avgSatisfaction,
			avgUsability,
			avgCompletionTime,
			avgErrorRate,
			recommendationRate)
	}

	// Add category metrics
	if len(results.Summary.CategoryMetrics) > 0 {
		markdown += `## Category Performance

`
		for category, metrics := range results.Summary.CategoryMetrics {
			markdown += fmt.Sprintf(`### %s

- **Test Cases:** %d
- **Pass Rate:** %.2f%%
- **User Satisfaction:** %.2f/10
- **Usability Score:** %.2f/10
- **Average Completion Time:** %s
- **Error Rate:** %.2f%%
- **Recommendation Rate:** %.2f%%

`,
				category,
				metrics.TestCaseCount,
				metrics.PassRate,
				metrics.UserSatisfaction,
				metrics.UsabilityScore,
				metrics.AverageCompletionTime,
				metrics.ErrorRate,
				metrics.RecommendationRate)
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
**Test Case:** %s  
**Description:** %s  
**User Impact:** %s  
**Business Impact:** %s  
**Recommendation:** %s  
**Detected:** %s

`,
				issue.Title,
				issue.Severity,
				issue.Category,
				issue.TestCaseID,
				issue.Description,
				issue.UserImpact,
				issue.BusinessImpact,
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
*Report generated by User Acceptance Testing Suite on %s*
`,
		time.Now().Format("2006-01-02 15:04:05"))

	return markdown
}

// UATHTMLTemplate is the HTML template for UAT reports
const UATHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Acceptance Testing Report - {{.SessionID}}</title>
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
        .test-cases { margin-bottom: 30px; }
        .test-case { background-color: #f8f9fa; margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .test-case h3 { margin: 0 0 15px 0; color: #333; }
        .test-case .meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 15px; }
        .test-case .meta-item { background-color: white; padding: 10px; border-radius: 4px; }
        .test-case .meta-item strong { color: #666; }
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
            <h1>User Acceptance Testing Report</h1>
            <h2>Session: {{.SessionID}}</h2>
            <p>Generated on {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
            <p><strong>Environment:</strong> {{.Environment}}</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <h3>Total Test Cases</h3>
                <div class="number">{{.TotalTestCases}}</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="number">{{.PassedTestCases}}</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="number">{{.FailedTestCases}}</div>
            </div>
            <div class="summary-card skipped">
                <h3>Skipped</h3>
                <div class="number">{{.SkippedTestCases}}</div>
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
                <h4>User Satisfaction</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallUserSatisfaction}}/10</div>
            </div>
            <div class="metric-card">
                <h4>Usability Score</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallUsabilityScore}}/10</div>
            </div>
            <div class="metric-card">
                <h4>Average Completion Time</h4>
                <div class="metric-value">{{.Summary.AverageCompletionTime}}</div>
            </div>
            <div class="metric-card">
                <h4>Error Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.AverageErrorRate}}%</div>
            </div>
            <div class="metric-card">
                <h4>Recommendation Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.RecommendationRate}}%</div>
            </div>
        </div>

        {{if .Issues}}
        <div class="issues">
            <h3>Issues Found ({{len .Issues}})</h3>
            {{range .Issues}}
            <div class="issue {{.Severity | lower}}">
                <h4>{{.Title}} <span class="status-badge status-{{.Severity | lower}}">{{.Severity}}</span></h4>
                <p><strong>Test Case:</strong> {{.TestCaseID}} | <strong>Category:</strong> {{.Category}}</p>
                <p><strong>Description:</strong> {{.Description}}</p>
                <p><strong>User Impact:</strong> {{.UserImpact}}</p>
                <p><strong>Business Impact:</strong> {{.BusinessImpact}}</p>
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

        <div class="test-cases">
            <h2>Test Case Results</h2>
            {{range $id, $results := .TestCaseResults}}
            <div class="test-case">
                <h3>{{$id}}</h3>
                <div class="meta">
                    <div class="meta-item">
                        <strong>Users Tested:</strong> {{len $results}}
                    </div>
                    {{if $results}}
                    <div class="meta-item">
                        <strong>User Satisfaction:</strong> {{printf "%.2f" (index $results 0).UserSatisfaction.OverallRating}}/10
                    </div>
                    <div class="meta-item">
                        <strong>Completion Time:</strong> {{(index $results 0).UsabilityMetrics.TimeToComplete}}
                    </div>
                    <div class="meta-item">
                        <strong>Would Recommend:</strong> {{if (index $results 0).UserSatisfaction.WouldRecommend}}Yes{{else}}No{{end}}
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p>Report generated by User Acceptance Testing Suite</p>
            <p>Execution time: {{.StartTime.Format "2006-01-02 15:04:05"}} - {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>
`
