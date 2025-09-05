package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	var (
		projectRoot    = flag.String("project", ".", "Project root directory")
		outputFile     = flag.String("output", "", "Output file for results (JSON)")
		format         = flag.String("format", "json", "Output format: json, markdown")
		verbose        = flag.Bool("verbose", false, "Enable verbose logging")
		generateReport = flag.Bool("report", false, "Generate detailed report")
		history        = flag.Bool("history", false, "Show metrics history")
		trends         = flag.Bool("trends", false, "Show trends analysis")
		alerts         = flag.Bool("alerts", false, "Show quality alerts")
		period         = flag.String("period", "7d", "Trend period: 1d, 7d, 30d")
		severity       = flag.String("severity", "all", "Alert severity: critical, high, medium, low, all")
	)
	flag.Parse()

	// Setup logging
	var logger *zap.Logger
	var err error
	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Resolve project root
	absProjectRoot, err := filepath.Abs(*projectRoot)
	if err != nil {
		logger.Fatal("Failed to resolve project root", zap.Error(err))
	}

	logger.Info("Starting code quality validation",
		zap.String("project_root", absProjectRoot),
		zap.String("format", *format),
	)

	// Create validator
	validator := observability.NewCodeQualityValidator(observability.NewLogger(logger))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Perform validation
	startTime := time.Now()
	metrics := validator.ValidateCodeQuality(ctx)

	duration := time.Since(startTime)
	logger.Info("Code quality validation completed",
		zap.Duration("duration", duration),
		zap.Float64("complexity", metrics.Complexity),
		zap.Float64("maintainability", metrics.Maintainability),
	)

	// Handle different output modes
	switch {
	case *history:
		showHistory(validator, logger)
	case *trends:
		showTrends(validator, logger, *period)
	case *alerts:
		showAlerts(validator, logger, *severity)
	case *generateReport:
		generateDetailedReport(validator, metrics, logger, *format, *outputFile)
	default:
		showSummary(metrics, validator, logger, *format, *outputFile)
	}
}

func showSummary(metrics *observability.CodeQualityMetrics, validator *observability.CodeQualityValidator, logger *zap.Logger, format, outputFile string) {
	var output []byte
	var err error

	switch format {
	case "markdown", "md":
		report, err := validator.GenerateQualityReport(metrics)
		if err != nil {
			logger.Fatal("Failed to generate quality report", zap.Error(err))
		}
		output, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal report", zap.Error(err))
		}
	case "json":
		fallthrough
	default:
		output, err = json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal metrics to JSON", zap.Error(err))
		}
	}

	// Write to file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, output, 0644)
		if err != nil {
			logger.Fatal("Failed to write output file", zap.Error(err))
		}
		logger.Info("Results written to file", zap.String("file", outputFile))
	} else {
		fmt.Println(string(output))
	}

	// Print summary to stderr
	fmt.Fprintf(os.Stderr, "\n=== Code Quality Summary ===\n")
	fmt.Fprintf(os.Stderr, "Complexity: %.1f/100\n", metrics.Complexity)
	fmt.Fprintf(os.Stderr, "Maintainability: %.1f/100\n", metrics.Maintainability)
	fmt.Fprintf(os.Stderr, "Reliability: %.1f/100\n", metrics.Reliability)
	fmt.Fprintf(os.Stderr, "Security: %.1f/100\n", metrics.Security)
	fmt.Fprintf(os.Stderr, "Test Coverage: %.1f%%\n", metrics.TestCoverage)
}

func showHistory(validator *observability.CodeQualityValidator, logger *zap.Logger) {
	history := validator.GetMetricsHistory()

	if len(history) == 0 {
		fmt.Println("No historical data available.")
		return
	}

	fmt.Println("=== Code Quality History ===")
	for i, metrics := range history {
		fmt.Printf("%d. %s - Quality: %.1f, Maintainability: %.1f, Debt: %.1f%%, Tests: %.1f%%\n",
			i+1,
			metrics.Timestamp.Format("2006-01-02 15:04:05"),
			metrics.Complexity,
			metrics.Maintainability,
			metrics.Reliability*100,
			metrics.TestCoverage,
		)
	}
}

func showTrends(validator *observability.CodeQualityValidator, logger *zap.Logger, period string) {
	history := validator.GetMetricsHistory()

	if len(history) < 2 {
		fmt.Println("Insufficient historical data for trend analysis.")
		return
	}

	// Calculate trends based on period
	var filteredHistory []observability.CodeQualityMetrics
	now := time.Now()

	for _, metric := range history {
		switch period {
		case "1d":
			if now.Sub(metric.Timestamp) <= 24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		case "7d":
			if now.Sub(metric.Timestamp) <= 7*24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		case "30d":
			if now.Sub(metric.Timestamp) <= 30*24*time.Hour {
				filteredHistory = append(filteredHistory, metric)
			}
		default:
			filteredHistory = append(filteredHistory, metric)
		}
	}

	if len(filteredHistory) < 2 {
		fmt.Printf("Insufficient data for period: %s\n", period)
		return
	}

	first := filteredHistory[0]
	last := filteredHistory[len(filteredHistory)-1]

	fmt.Printf("=== Code Quality Trends (%s) ===\n", period)
	fmt.Printf("Quality Score: %.1f → %.1f (%+.1f)\n",
		first.Complexity, last.Complexity,
		last.Complexity-first.Complexity)
	fmt.Printf("Maintainability: %.1f → %.1f (%+.1f)\n",
		first.Maintainability, last.Maintainability,
		last.Maintainability-first.Maintainability)
	fmt.Printf("Test Coverage: %.1f%% → %.1f%% (%+.1f%%)\n",
		first.TestCoverage, last.TestCoverage,
		last.TestCoverage-first.TestCoverage)
	fmt.Printf("Technical Debt: %.1f%% → %.1f%% (%+.1f%%)\n",
		first.Reliability*100, last.Reliability*100,
		(first.Reliability-last.Reliability)*100)
	fmt.Printf("Data Points: %d\n", len(filteredHistory))
}

func showAlerts(validator *observability.CodeQualityValidator, logger *zap.Logger, severity string) {
	ctx := context.Background()
	metrics, err := validator.ValidateCodeQuality(ctx)
	if err != nil {
		logger.Fatal("Failed to validate code quality for alerts", zap.Error(err))
	}

	// Generate alerts based on metrics
	var alerts []map[string]interface{}

	// Critical alerts
	if severity == "all" || severity == "critical" {
		if metrics.Reliability > 0.5 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "critical",
				"title":    "High Technical Debt",
				"message":  "Technical debt ratio is above 50%",
				"value":    metrics.Reliability,
			})
		}

		if metrics.TestCoverage < 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "critical",
				"title":    "Low Test Coverage",
				"message":  "Test coverage is below 50%",
				"value":    metrics.TestCoverage,
			})
		}
	}

	// High severity alerts
	if severity == "all" || severity == "high" {
		if metrics.CyclomaticComplexity > 15 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "high",
				"title":    "High Complexity",
				"message":  "Cyclomatic complexity is above 15",
				"value":    metrics.CyclomaticComplexity,
			})
		}

		if metrics.Complexity < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "high",
				"title":    "Low Code Quality",
				"message":  "Code quality score is below 60",
				"value":    metrics.Complexity,
			})
		}
	}

	// Medium severity alerts
	if severity == "all" || severity == "medium" {
		if metrics.AverageFunctionSize > 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "medium",
				"title":    "Large Functions",
				"message":  "Average function size is above 50 lines",
				"value":    metrics.AverageFunctionSize,
			})
		}

		if metrics.DocumentationCoverage < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "medium",
				"title":    "Low Documentation",
				"message":  "Documentation coverage is below 60%",
				"value":    metrics.DocumentationCoverage,
			})
		}
	}

	// Low severity alerts
	if severity == "all" || severity == "low" {
		if metrics.CodeSmells > 5 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "low",
				"title":    "Code Smells",
				"message":  "Multiple code smells detected",
				"value":    metrics.CodeSmells,
			})
		}

		if metrics.CommentRatio < 10 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "low",
				"title":    "Low Comment Ratio",
				"message":  "Comment ratio is below 10%",
				"value":    metrics.CommentRatio,
			})
		}
	}

	if len(alerts) == 0 {
		fmt.Printf("No %s severity alerts found.\n", severity)
		return
	}

	fmt.Printf("=== Code Quality Alerts (%s severity) ===\n", severity)
	for i, alert := range alerts {
		fmt.Printf("%d. [%s] %s: %s (Value: %.1f)\n",
			i+1,
			alert["severity"],
			alert["title"],
			alert["message"],
			alert["value"],
		)
	}
}

func generateDetailedReport(validator *observability.CodeQualityValidator, metrics *observability.CodeQualityMetrics, logger *zap.Logger, format, outputFile string) {
	var output []byte
	var err error

	switch format {
	case "markdown", "md":
		report, err := validator.GenerateQualityReport(metrics)
		if err != nil {
			logger.Fatal("Failed to generate quality report", zap.Error(err))
		}
		output, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal report", zap.Error(err))
		}
	case "json":
		fallthrough
	default:
		// Create detailed JSON report
		detailedReport := map[string]interface{}{
			"metrics": metrics,
			"report": map[string]interface{}{
				"summary": map[string]interface{}{
					"quality_score":         metrics.Complexity,
					"maintainability_index": metrics.Maintainability,
					"technical_debt_ratio":  metrics.Reliability,
					"test_coverage":         metrics.TestCoverage,
					"improvement_score":     metrics.ImprovementScore,
					"trend_direction":       metrics.TrendDirection,
				},
				"recommendations": generateRecommendations(metrics),
				"trends":          generateTrends(validator),
				"alerts":          generateAlerts(metrics, "all"),
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		output, err = json.MarshalIndent(detailedReport, "", "  ")
		if err != nil {
			logger.Fatal("Failed to marshal detailed report to JSON", zap.Error(err))
		}
	}

	// Write to file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, output, 0644)
		if err != nil {
			logger.Fatal("Failed to write output file", zap.Error(err))
		}
		logger.Info("Detailed report written to file", zap.String("file", outputFile))
	} else {
		fmt.Println(string(output))
	}
}

// Helper functions for report generation
func generateRecommendations(metrics *observability.CodeQualityMetrics) []string {
	var recommendations []string

	if metrics.CyclomaticComplexity > 10 {
		recommendations = append(recommendations, "High cyclomatic complexity detected. Consider refactoring complex functions.")
	}

	if metrics.AverageFunctionSize > 30 {
		recommendations = append(recommendations, "Large average function size. Break down large functions into smaller, focused functions.")
	}

	if metrics.TestCoverage < 80 {
		recommendations = append(recommendations, "Test coverage below 80%. Increase test coverage for better code quality.")
	}

	if metrics.Reliability > 0.3 {
		recommendations = append(recommendations, "High technical debt ratio. Prioritize debt reduction in upcoming sprints.")
	}

	if metrics.CodeSmells > 10 {
		recommendations = append(recommendations, "Multiple code smells detected. Review and refactor problematic code.")
	}

	if metrics.DocumentationCoverage < 70 {
		recommendations = append(recommendations, "Low documentation coverage. Improve code documentation.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Code quality is good. Continue maintaining current standards.")
	}

	return recommendations
}

func generateTrends(validator *observability.CodeQualityValidator) map[string]interface{} {
	history := validator.GetMetricsHistory()

	if len(history) < 2 {
		return map[string]interface{}{
			"status":  "insufficient_data",
			"message": "Insufficient historical data for trend analysis",
		}
	}

	recent := history[len(history)-1]
	previous := history[len(history)-2]

	return map[string]interface{}{
		"status": "available",
		"changes": map[string]interface{}{
			"quality_score": map[string]interface{}{
				"current":  recent.CodeQualityScore,
				"previous": previous.CodeQualityScore,
				"change":   recent.CodeQualityScore - previous.CodeQualityScore,
			},
			"maintainability": map[string]interface{}{
				"current":  recent.MaintainabilityIndex,
				"previous": previous.MaintainabilityIndex,
				"change":   recent.MaintainabilityIndex - previous.MaintainabilityIndex,
			},
			"test_coverage": map[string]interface{}{
				"current":  recent.TestCoverage,
				"previous": previous.TestCoverage,
				"change":   recent.TestCoverage - previous.TestCoverage,
			},
			"technical_debt": map[string]interface{}{
				"current":  recent.TechnicalDebtRatio,
				"previous": previous.TechnicalDebtRatio,
				"change":   previous.TechnicalDebtRatio - recent.TechnicalDebtRatio, // Lower is better
			},
		},
		"trend_direction":   recent.TrendDirection,
		"improvement_score": recent.ImprovementScore,
	}
}

func generateAlerts(metrics *observability.CodeQualityMetrics, severity string) []map[string]interface{} {
	var alerts []map[string]interface{}

	// Critical alerts
	if severity == "all" || severity == "critical" {
		if metrics.Reliability > 0.5 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "critical",
				"title":    "High Technical Debt",
				"message":  "Technical debt ratio is above 50%",
				"value":    metrics.Reliability,
			})
		}

		if metrics.TestCoverage < 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "critical",
				"title":    "Low Test Coverage",
				"message":  "Test coverage is below 50%",
				"value":    metrics.TestCoverage,
			})
		}
	}

	// High severity alerts
	if severity == "all" || severity == "high" {
		if metrics.CyclomaticComplexity > 15 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "high",
				"title":    "High Complexity",
				"message":  "Cyclomatic complexity is above 15",
				"value":    metrics.CyclomaticComplexity,
			})
		}

		if metrics.Complexity < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "high",
				"title":    "Low Code Quality",
				"message":  "Code quality score is below 60",
				"value":    metrics.Complexity,
			})
		}
	}

	// Medium severity alerts
	if severity == "all" || severity == "medium" {
		if metrics.AverageFunctionSize > 50 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "medium",
				"title":    "Large Functions",
				"message":  "Average function size is above 50 lines",
				"value":    metrics.AverageFunctionSize,
			})
		}

		if metrics.DocumentationCoverage < 60 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "medium",
				"title":    "Low Documentation",
				"message":  "Documentation coverage is below 60%",
				"value":    metrics.DocumentationCoverage,
			})
		}
	}

	// Low severity alerts
	if severity == "all" || severity == "low" {
		if metrics.CodeSmells > 5 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "low",
				"title":    "Code Smells",
				"message":  "Multiple code smells detected",
				"value":    metrics.CodeSmells,
			})
		}

		if metrics.CommentRatio < 10 {
			alerts = append(alerts, map[string]interface{}{
				"severity": "low",
				"title":    "Low Comment Ratio",
				"message":  "Comment ratio is below 10%",
				"value":    metrics.CommentRatio,
			})
		}
	}

	return alerts
}
