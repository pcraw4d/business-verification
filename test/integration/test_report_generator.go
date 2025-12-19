//go:build comprehensive_test
// +build comprehensive_test

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TestReportGenerator generates comprehensive test reports
type TestReportGenerator struct {
	results []ClassificationTestResult
	metrics *TestMetrics
}

// NewTestReportGenerator creates a new report generator
func NewTestReportGenerator(results []ClassificationTestResult, metrics *TestMetrics) *TestReportGenerator {
	return &TestReportGenerator{
		results: results,
		metrics: metrics,
	}
}

// GenerateHTMLReport generates an HTML report
func (g *TestReportGenerator) GenerateHTMLReport(outputPath string) error {
	html := g.buildHTMLReport()
	return os.WriteFile(outputPath, []byte(html), 0644)
}

// GenerateJSONReport generates a JSON report
func (g *TestReportGenerator) GenerateJSONReport(outputPath string) error {
	report := g.buildJSONReport()
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	return os.WriteFile(outputPath, jsonData, 0644)
}

// GenerateMarkdownReport generates a Markdown report
func (g *TestReportGenerator) GenerateMarkdownReport(outputPath string) error {
	markdown := g.buildMarkdownReport()
	return os.WriteFile(outputPath, []byte(markdown), 0644)
}

// buildJSONReport builds the JSON report structure
func (g *TestReportGenerator) buildJSONReport() map[string]interface{} {
	return map[string]interface{}{
		"test_summary": map[string]interface{}{
			"total_samples":    g.metrics.TotalTests,
			"successful_tests":  g.metrics.SuccessfulTests,
			"failed_tests":      g.metrics.FailedTests,
			"overall_accuracy":  g.metrics.Accuracy,
			"timestamp":         time.Now().Format(time.RFC3339),
		},
		"performance_metrics": map[string]interface{}{
			"average_latency_ms": g.metrics.AverageLatency.Milliseconds(),
			"p50_latency_ms":     g.metrics.P50Latency.Milliseconds(),
			"p95_latency_ms":     g.metrics.P95Latency.Milliseconds(),
			"p99_latency_ms":     g.metrics.P99Latency.Milliseconds(),
			"throughput_rps":     g.metrics.Throughput,
		},
		"accuracy_metrics": map[string]interface{}{
			"overall_accuracy":      g.metrics.Accuracy,
			"accuracy_by_industry":  g.metrics.AccuracyByIndustry,
			"accuracy_by_complexity": g.metrics.AccuracyByComplexity,
		},
		"strategy_distribution": map[string]interface{}{
			"counts":        g.metrics.StrategyDistribution,
			"percentages":   calculatePercentages(g.metrics.StrategyDistribution, g.metrics.TotalTests),
			"success_rates": g.metrics.StrategySuccessRate,
			"latencies_ms":  convertDurationsToMS(g.metrics.StrategyLatency),
		},
		"optimization_metrics": map[string]interface{}{
			"early_exit_count": g.metrics.EarlyExitCount,
			"early_exit_rate":  g.metrics.EarlyExitRate,
			"cache_hit_count":  g.metrics.CacheHitCount,
			"cache_hit_rate":   g.metrics.CacheHitRate,
			"fallback_usage":   g.metrics.FallbackUsage,
		},
		"frontend_compatibility": map[string]interface{}{
			"all_fields_present":  g.metrics.FrontendCompatibility.AllFieldsPresent,
			"data_types_correct":  g.metrics.FrontendCompatibility.DataTypesCorrect,
			"structure_valid":     g.metrics.FrontendCompatibility.StructureValid,
			"industry_present":    g.metrics.FrontendCompatibility.IndustryPresent,
			"codes_present":       g.metrics.FrontendCompatibility.CodesPresent,
			"explanation_present": g.metrics.FrontendCompatibility.ExplanationPresent,
			"top3_codes_present":  g.metrics.FrontendCompatibility.Top3CodesPresent,
		},
		"code_accuracy": map[string]interface{}{
			"mcc_accuracy":   g.metrics.CodeAccuracy.MCCAccuracy,
			"naics_accuracy": g.metrics.CodeAccuracy.NAICSAccuracy,
			"sic_accuracy":   g.metrics.CodeAccuracy.SICAccuracy,
			"top3_match_rate": g.metrics.CodeAccuracy.Top3MatchRate,
		},
		"detailed_results": g.results,
	}
}

// buildHTMLReport builds an HTML report
func (g *TestReportGenerator) buildHTMLReport() string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Comprehensive Classification E2E Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background: #f9f9f9; border-radius: 5px; min-width: 200px; }
        .metric-label { font-weight: bold; color: #666; }
        .metric-value { font-size: 24px; color: #4CAF50; }
        .success { color: #4CAF50; }
        .warning { color: #FF9800; }
        .error { color: #F44336; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #4CAF50; color: white; }
        tr:hover { background-color: #f5f5f5; }
        .summary { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Comprehensive Classification E2E Test Report</h1>
        <div class="summary">
            <p><strong>Generated:</strong> ` + time.Now().Format(time.RFC3339) + `</p>
            <p><strong>Total Tests:</strong> ` + fmt.Sprintf("%d", g.metrics.TotalTests) + `</p>
            <p><strong>Successful:</strong> ` + fmt.Sprintf("%d (%.1f%%)", g.metrics.SuccessfulTests, float64(g.metrics.SuccessfulTests)/float64(g.metrics.TotalTests)*100) + `</p>
            <p><strong>Failed:</strong> ` + fmt.Sprintf("%d (%.1f%%)", g.metrics.FailedTests, float64(g.metrics.FailedTests)/float64(g.metrics.TotalTests)*100) + `</p>
            <p><strong>Overall Accuracy:</strong> ` + fmt.Sprintf("%.2f%%", g.metrics.Accuracy*100) + `</p>
        </div>

        <h2>Performance Metrics</h2>
        <div class="metric">
            <div class="metric-label">Average Latency</div>
            <div class="metric-value">` + fmt.Sprintf("%.0fms", float64(g.metrics.AverageLatency.Milliseconds())) + `</div>
        </div>
        <div class="metric">
            <div class="metric-label">P95 Latency</div>
            <div class="metric-value">` + fmt.Sprintf("%.0fms", float64(g.metrics.P95Latency.Milliseconds())) + `</div>
        </div>
        <div class="metric">
            <div class="metric-label">P99 Latency</div>
            <div class="metric-value">` + fmt.Sprintf("%.0fms", float64(g.metrics.P99Latency.Milliseconds())) + `</div>
        </div>
        <div class="metric">
            <div class="metric-label">Throughput</div>
            <div class="metric-value">` + fmt.Sprintf("%.2f req/s", g.metrics.Throughput) + `</div>
        </div>

        <h2>Strategy Distribution</h2>
        <table>
            <tr>
                <th>Strategy</th>
                <th>Count</th>
                <th>Percentage</th>
                <th>Success Rate</th>
                <th>Avg Latency</th>
            </tr>`

	for strategy, count := range g.metrics.StrategyDistribution {
		percentage := float64(count) / float64(g.metrics.TotalTests) * 100
		successRate := g.metrics.StrategySuccessRate[strategy] * 100
		avgLatency := g.metrics.StrategyLatency[strategy].Milliseconds()
		html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
                <td>%.1f%%</td>
                <td>%.1f%%</td>
                <td>%.0fms</td>
            </tr>`, strategy, count, percentage, successRate, float64(avgLatency))
	}

	html += `
        </table>

        <h2>Optimization Metrics</h2>
        <div class="metric">
            <div class="metric-label">Early Exit Rate</div>
            <div class="metric-value">` + fmt.Sprintf("%.1f%%", g.metrics.EarlyExitRate*100) + `</div>
        </div>
        <div class="metric">
            <div class="metric-label">Cache Hit Rate</div>
            <div class="metric-value">` + fmt.Sprintf("%.1f%%", g.metrics.CacheHitRate*100) + `</div>
        </div>

        <h2>Frontend Compatibility</h2>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
                <th>Status</th>
            </tr>
            <tr>
                <td>All Fields Present</td>
                <td>` + fmt.Sprintf("%.1f%%", g.metrics.FrontendCompatibility.AllFieldsPresent*100) + `</td>
                <td>` + getStatusHTML(g.metrics.FrontendCompatibility.AllFieldsPresent) + `</td>
            </tr>
            <tr>
                <td>Industry Present</td>
                <td>` + fmt.Sprintf("%.1f%%", g.metrics.FrontendCompatibility.IndustryPresent*100) + `</td>
                <td>` + getStatusHTML(g.metrics.FrontendCompatibility.IndustryPresent) + `</td>
            </tr>
            <tr>
                <td>Codes Present</td>
                <td>` + fmt.Sprintf("%.1f%%", g.metrics.FrontendCompatibility.CodesPresent*100) + `</td>
                <td>` + getStatusHTML(g.metrics.FrontendCompatibility.CodesPresent) + `</td>
            </tr>
            <tr>
                <td>Explanation Present</td>
                <td>` + fmt.Sprintf("%.1f%%", g.metrics.FrontendCompatibility.ExplanationPresent*100) + `</td>
                <td>` + getStatusHTML(g.metrics.FrontendCompatibility.ExplanationPresent) + `</td>
            </tr>
            <tr>
                <td>Top 3 Codes Present</td>
                <td>` + fmt.Sprintf("%.1f%%", g.metrics.FrontendCompatibility.Top3CodesPresent*100) + `</td>
                <td>` + getStatusHTML(g.metrics.FrontendCompatibility.Top3CodesPresent) + `</td>
            </tr>
        </table>

        <h2>Accuracy by Industry</h2>
        <table>
            <tr>
                <th>Industry</th>
                <th>Accuracy</th>
            </tr>`

	for industry, accuracy := range g.metrics.AccuracyByIndustry {
		html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%.2f%%</td>
            </tr>`, industry, accuracy*100)
	}

	html += `
        </table>
    </div>
</body>
</html>`

	return html
}

// buildMarkdownReport builds a Markdown report
func (g *TestReportGenerator) buildMarkdownReport() string {
	md := fmt.Sprintf(`# Comprehensive Classification E2E Test Report

**Generated:** %s

## Test Summary

- **Total Tests:** %d
- **Successful:** %d (%.1f%%)
- **Failed:** %d (%.1f%%)
- **Overall Accuracy:** %.2f%%

## Performance Metrics

- **Average Latency:** %.0fms
- **P50 Latency:** %.0fms
- **P95 Latency:** %.0fms
- **P99 Latency:** %.0fms
- **Throughput:** %.2f req/s

## Strategy Distribution

| Strategy | Count | Percentage | Success Rate | Avg Latency |
|----------|-------|-----------|--------------|-------------|
`,
		time.Now().Format(time.RFC3339),
		g.metrics.TotalTests,
		g.metrics.SuccessfulTests,
		float64(g.metrics.SuccessfulTests)/float64(g.metrics.TotalTests)*100,
		g.metrics.FailedTests,
		float64(g.metrics.FailedTests)/float64(g.metrics.TotalTests)*100,
		g.metrics.Accuracy*100,
		float64(g.metrics.AverageLatency.Milliseconds()),
		float64(g.metrics.P50Latency.Milliseconds()),
		float64(g.metrics.P95Latency.Milliseconds()),
		float64(g.metrics.P99Latency.Milliseconds()),
		g.metrics.Throughput,
	)

	for strategy, count := range g.metrics.StrategyDistribution {
		percentage := float64(count) / float64(g.metrics.TotalTests) * 100
		successRate := g.metrics.StrategySuccessRate[strategy] * 100
		avgLatency := g.metrics.StrategyLatency[strategy].Milliseconds()
		md += fmt.Sprintf("| %s | %d | %.1f%% | %.1f%% | %.0fms |\n",
			strategy, count, percentage, successRate, float64(avgLatency))
	}

	md += fmt.Sprintf(`
## Optimization Metrics

- **Early Exit Rate:** %.1f%% (%d occurrences)
- **Cache Hit Rate:** %.1f%% (%d hits)

## Frontend Compatibility

- **All Fields Present:** %.1f%%
- **Industry Present:** %.1f%%
- **Codes Present:** %.1f%%
- **Explanation Present:** %.1f%%
- **Top 3 Codes Present:** %.1f%%

## Accuracy by Industry

`,
		g.metrics.EarlyExitRate*100,
		g.metrics.EarlyExitCount,
		g.metrics.CacheHitRate*100,
		g.metrics.CacheHitCount,
		g.metrics.FrontendCompatibility.AllFieldsPresent*100,
		g.metrics.FrontendCompatibility.IndustryPresent*100,
		g.metrics.FrontendCompatibility.CodesPresent*100,
		g.metrics.FrontendCompatibility.ExplanationPresent*100,
		g.metrics.FrontendCompatibility.Top3CodesPresent*100,
	)

	md += "| Industry | Accuracy |\n"
	md += "|----------|----------|\n"
	for industry, accuracy := range g.metrics.AccuracyByIndustry {
		md += fmt.Sprintf("| %s | %.2f%% |\n", industry, accuracy*100)
	}

	return md
}

func getStatusHTML(value float64) string {
	if value >= 0.95 {
		return `<span class="success">✓ Pass</span>`
	} else if value >= 0.80 {
		return `<span class="warning">⚠ Warning</span>`
	}
	return `<span class="error">✗ Fail</span>`
}

