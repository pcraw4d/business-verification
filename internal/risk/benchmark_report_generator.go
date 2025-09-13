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

// GenerateReports generates comprehensive benchmark reports
func (brg *BenchmarkReportGenerator) GenerateReports(results *BenchmarkResults) error {
	brg.logger.Info("Generating benchmark reports")

	// Create report output directory if it doesn't exist
	if err := os.MkdirAll(brg.config.ReportOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create report output directory: %w", err)
	}

	// Generate JSON report
	if err := brg.GenerateJSONReport(results); err != nil {
		return fmt.Errorf("failed to generate JSON report: %w", err)
	}

	// Generate HTML report
	if err := brg.GenerateHTMLReport(results); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Generate Markdown report
	if err := brg.GenerateMarkdownReport(results); err != nil {
		return fmt.Errorf("failed to generate Markdown report: %w", err)
	}

	brg.logger.Info("Benchmark reports generated successfully")
	return nil
}

// GenerateJSONReport generates a JSON report for benchmark results
func (brg *BenchmarkReportGenerator) GenerateJSONReport(results *BenchmarkResults) error {
	brg.logger.Info("Generating JSON benchmark report")

	reportPath := filepath.Join(brg.config.ReportOutputPath, "benchmark-report.json")

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(reportPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	brg.logger.Info("JSON benchmark report generated", zap.String("path", reportPath))
	return nil
}

// GenerateHTMLReport generates an HTML report for benchmark results
func (brg *BenchmarkReportGenerator) GenerateHTMLReport(results *BenchmarkResults) error {
	brg.logger.Info("Generating HTML benchmark report")

	reportPath := filepath.Join(brg.config.ReportOutputPath, "benchmark-report.html")

	// Parse the template
	tmpl, err := template.New("html-report").Parse(BenchmarkHTMLTemplate)
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

	brg.logger.Info("HTML benchmark report generated", zap.String("path", reportPath))
	return nil
}

// GenerateMarkdownReport generates a Markdown report for benchmark results
func (brg *BenchmarkReportGenerator) GenerateMarkdownReport(results *BenchmarkResults) error {
	brg.logger.Info("Generating Markdown benchmark report")

	reportPath := filepath.Join(brg.config.ReportOutputPath, "benchmark-report.md")

	// Generate Markdown content
	markdown := brg.generateMarkdownContent(results)

	// Write to file
	if err := os.WriteFile(reportPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report: %w", err)
	}

	brg.logger.Info("Markdown benchmark report generated", zap.String("path", reportPath))
	return nil
}

// generateMarkdownContent generates the Markdown content for the report
func (brg *BenchmarkReportGenerator) generateMarkdownContent(results *BenchmarkResults) string {
	markdown := fmt.Sprintf(`# Performance Benchmark Report

## Benchmark Session Information

**Session ID:** %s  
**Environment:** %s  
**Start Time:** %s  
**End Time:** %s  
**Total Duration:** %s

## Summary

| Metric | Value |
|--------|-------|
| Total Benchmarks | %d |
| Passed Benchmarks | %d |
| Failed Benchmarks | %d |
| Skipped Benchmarks | %d |
| Pass Rate | %.2f%% |

## Performance Metrics

### Overall Performance
- **Throughput:** %.2f operations/second
- **Average Latency:** %.2f ms
- **P95 Latency:** %.2f ms
- **P99 Latency:** %.2f ms
- **Error Rate:** %.2f%%
- **Success Rate:** %.2f%%

### Resource Usage
- **Memory Usage:** %d MB
- **CPU Usage:** %.2f%%
- **Goroutines:** %d

## Benchmark Results

`,
		results.SessionID,
		results.Environment,
		results.StartTime.Format("2006-01-02 15:04:05"),
		results.EndTime.Format("2006-01-02 15:04:05"),
		results.TotalDuration,
		results.TotalBenchmarks,
		results.PassedBenchmarks,
		results.FailedBenchmarks,
		results.SkippedBenchmarks,
		results.PassRate,
		results.Summary.OverallThroughput,
		results.Summary.OverallLatency,
		results.Summary.OverallP95Latency,
		results.Summary.OverallP99Latency,
		results.Summary.OverallErrorRate,
		results.Summary.OverallSuccessRate,
		results.Summary.OverallMemoryUsage/(1024*1024),
		results.Summary.OverallCPUUsage,
		results.Summary.OverallGoroutines)

	// Add benchmark results
	for benchmarkID, benchmarkResults := range results.BenchmarkResults {
		if len(benchmarkResults) == 0 {
			continue
		}

		// Calculate averages for this benchmark
		avgThroughput := 0.0
		avgLatency := 0.0
		avgP95Latency := 0.0
		avgP99Latency := 0.0
		avgErrorRate := 0.0
		avgSuccessRate := 0.0
		successCount := 0

		for _, result := range benchmarkResults {
			if result.Success && result.Metrics != nil {
				avgThroughput += result.Metrics.Throughput
				avgLatency += result.Metrics.Latency
				avgP95Latency += result.Metrics.P95Latency
				avgP99Latency += result.Metrics.P99Latency
				avgErrorRate += result.Metrics.ErrorRate
				avgSuccessRate += result.Metrics.SuccessRate
				successCount++
			}
		}

		if successCount > 0 {
			avgThroughput /= float64(successCount)
			avgLatency /= float64(successCount)
			avgP95Latency /= float64(successCount)
			avgP99Latency /= float64(successCount)
			avgErrorRate /= float64(successCount)
			avgSuccessRate /= float64(successCount)
		}

		markdown += fmt.Sprintf(`### %s

**Results:** %d iterations  
**Average Throughput:** %.2f operations/second  
**Average Latency:** %.2f ms  
**P95 Latency:** %.2f ms  
**P99 Latency:** %.2f ms  
**Error Rate:** %.2f%%  
**Success Rate:** %.2f%%

`,
			benchmarkID,
			len(benchmarkResults),
			avgThroughput,
			avgLatency,
			avgP95Latency,
			avgP99Latency,
			avgErrorRate,
			avgSuccessRate)
	}

	// Add category metrics
	if len(results.Summary.CategoryMetrics) > 0 {
		markdown += `## Category Performance

`
		for category, metrics := range results.Summary.CategoryMetrics {
			markdown += fmt.Sprintf(`### %s

- **Benchmarks:** %d
- **Average Throughput:** %.2f operations/second
- **Average Latency:** %.2f ms
- **P95 Latency:** %.2f ms
- **P99 Latency:** %.2f ms
- **Error Rate:** %.2f%%
- **Success Rate:** %.2f%%
- **Pass Rate:** %.2f%%

`,
				category,
				metrics.BenchmarkCount,
				metrics.AvgThroughput,
				metrics.AvgLatency,
				metrics.AvgP95Latency,
				metrics.AvgP99Latency,
				metrics.AvgErrorRate,
				metrics.AvgSuccessRate,
				metrics.PassRate)
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
**Benchmark:** %s  
**Description:** %s  
**Expected:** %v  
**Actual:** %v  
**Impact:** %s  
**Recommendation:** %s  
**Detected:** %s

`,
				issue.Title,
				issue.Severity,
				issue.Category,
				issue.BenchmarkID,
				issue.Description,
				issue.ExpectedValue,
				issue.ActualValue,
				issue.Impact,
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
*Report generated by Performance Benchmarking Suite on %s*
`,
		time.Now().Format("2006-01-02 15:04:05"))

	return markdown
}

// BenchmarkHTMLTemplate is the HTML template for benchmark reports
const BenchmarkHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Performance Benchmark Report - {{.SessionID}}</title>
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
        .benchmarks { margin-bottom: 30px; }
        .benchmark { background-color: #f8f9fa; margin-bottom: 20px; padding: 20px; border-radius: 8px; border-left: 4px solid #007bff; }
        .benchmark h3 { margin: 0 0 15px 0; color: #333; }
        .benchmark .meta { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 15px; }
        .benchmark .meta-item { background-color: white; padding: 10px; border-radius: 4px; }
        .benchmark .meta-item strong { color: #666; }
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
            <h1>Performance Benchmark Report</h1>
            <h2>Session: {{.SessionID}}</h2>
            <p>Generated on {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
            <p><strong>Environment:</strong> {{.Environment}}</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <h3>Total Benchmarks</h3>
                <div class="number">{{.TotalBenchmarks}}</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="number">{{.PassedBenchmarks}}</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="number">{{.FailedBenchmarks}}</div>
            </div>
            <div class="summary-card skipped">
                <h3>Skipped</h3>
                <div class="number">{{.SkippedBenchmarks}}</div>
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
                <h4>Overall Throughput</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallThroughput}} ops/s</div>
            </div>
            <div class="metric-card">
                <h4>Average Latency</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallLatency}} ms</div>
            </div>
            <div class="metric-card">
                <h4>P95 Latency</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallP95Latency}} ms</div>
            </div>
            <div class="metric-card">
                <h4>P99 Latency</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallP99Latency}} ms</div>
            </div>
            <div class="metric-card">
                <h4>Error Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallErrorRate}}%</div>
            </div>
            <div class="metric-card">
                <h4>Success Rate</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallSuccessRate}}%</div>
            </div>
            <div class="metric-card">
                <h4>Memory Usage</h4>
                <div class="metric-value">{{printf "%.0f" (div .Summary.OverallMemoryUsage 1048576)}} MB</div>
            </div>
            <div class="metric-card">
                <h4>CPU Usage</h4>
                <div class="metric-value">{{printf "%.2f" .Summary.OverallCPUUsage}}%</div>
            </div>
        </div>

        {{if .Issues}}
        <div class="issues">
            <h3>Issues Found ({{len .Issues}})</h3>
            {{range .Issues}}
            <div class="issue {{.Severity | lower}}">
                <h4>{{.Title}} <span class="status-badge status-{{.Severity | lower}}">{{.Severity}}</span></h4>
                <p><strong>Benchmark:</strong> {{.BenchmarkID}} | <strong>Category:</strong> {{.Category}}</p>
                <p><strong>Description:</strong> {{.Description}}</p>
                <p><strong>Expected:</strong> {{.ExpectedValue}} | <strong>Actual:</strong> {{.ActualValue}}</p>
                <p><strong>Impact:</strong> {{.Impact}}</p>
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

        <div class="benchmarks">
            <h2>Benchmark Results</h2>
            {{range $id, $results := .BenchmarkResults}}
            <div class="benchmark">
                <h3>{{$id}}</h3>
                <div class="meta">
                    <div class="meta-item">
                        <strong>Iterations:</strong> {{len $results}}
                    </div>
                    {{if $results}}
                    <div class="meta-item">
                        <strong>Avg Throughput:</strong> {{printf "%.2f" (index $results 0).Metrics.Throughput}} ops/s
                    </div>
                    <div class="meta-item">
                        <strong>Avg Latency:</strong> {{printf "%.2f" (index $results 0).Metrics.Latency}} ms
                    </div>
                    <div class="meta-item">
                        <strong>P95 Latency:</strong> {{printf "%.2f" (index $results 0).Metrics.P95Latency}} ms
                    </div>
                    <div class="meta-item">
                        <strong>P99 Latency:</strong> {{printf "%.2f" (index $results 0).Metrics.P99Latency}} ms
                    </div>
                    <div class="meta-item">
                        <strong>Error Rate:</strong> {{printf "%.2f" (index $results 0).Metrics.ErrorRate}}%
                    </div>
                    <div class="meta-item">
                        <strong>Success Rate:</strong> {{printf "%.2f" (index $results 0).Metrics.SuccessRate}}%
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>

        <div class="footer">
            <p>Report generated by Performance Benchmarking Suite</p>
            <p>Execution time: {{.StartTime.Format "2006-01-02 15:04:05"}} - {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
        </div>
    </div>
</body>
</html>
`
