package observability

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
)

// ExampleAdvancedMonitoringDashboard demonstrates how to use the advanced monitoring dashboard
func ExampleAdvancedMonitoringDashboard() {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// Create dashboard configuration
	config := &AdvancedDashboardConfig{
		DashboardEnabled:               true,
		UpdateInterval:                 30 * time.Second,
		HealthCheckInterval:            1 * time.Minute,
		AlertSummaryInterval:           5 * time.Minute,
		RealTimeUpdateInterval:         5 * time.Second,
		MaxAlertsDisplayed:             50,
		MaxMetricsHistory:              1000,
		ShowDetailedMetrics:            true,
		ShowTrendAnalysis:              true,
		ShowMLModelMetrics:             true,
		ShowEnsembleMetrics:            true,
		ShowUncertaintyMetrics:         true,
		ShowSecurityMetrics:            true,
		IntegrateMLMonitoring:          true,
		IntegrateEnsembleMonitoring:    true,
		IntegrateUncertaintyMonitoring: true,
		IntegrateSecurityMonitoring:    true,
		IntegratePerformanceMonitoring: true,
	}

	// Create advanced monitoring dashboard
	dashboard := NewAdvancedMonitoringDashboard(
		config,
		nil, // metricsCollector - would be injected in real usage
		nil, // performanceMonitor - would be injected in real usage
		nil, // alertManager - would be injected in real usage
		nil, // healthChecker - would be injected in real usage
		logger,
	)

	// Get dashboard data
	ctx := context.Background()
	data, err := dashboard.GetDashboardData(ctx)
	if err != nil {
		logger.Error("Failed to get dashboard data", zap.Error(err))
		return
	}

	// Display dashboard information
	fmt.Printf("=== Advanced Monitoring Dashboard ===\n")
	fmt.Printf("Timestamp: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Printf("Overall Health: %s\n", data.OverallHealth)
	fmt.Printf("Health Score: %.2f/100\n", data.HealthScore)
	fmt.Printf("Last Updated: %s\n", dashboard.GetLastUpdateTime().Format(time.RFC3339))

	// Display ML model metrics
	if len(data.MLModelMetrics) > 0 {
		fmt.Printf("\n=== ML Model Metrics ===\n")
		fmt.Printf("ML Model Health: %s\n", data.MLModelHealth)
		for modelID, metrics := range data.MLModelMetrics {
			fmt.Printf("Model %s: Accuracy=%.2f, Drift=%.2f, Status=%s\n",
				modelID, metrics.Accuracy, metrics.DriftScore, metrics.DriftStatus)
		}
	}

	// Display ensemble metrics
	if len(data.EnsembleMetrics) > 0 {
		fmt.Printf("\n=== Ensemble Metrics ===\n")
		fmt.Printf("Ensemble Health: %s\n", data.EnsembleHealth)
		for methodID, metrics := range data.EnsembleMetrics {
			fmt.Printf("Method %s: Weight=%.2f, Contribution=%.2f, Accuracy=%.2f\n",
				methodID, metrics.Weight, metrics.Contribution, metrics.Accuracy)
		}
	}

	// Display uncertainty metrics
	if data.UncertaintyMetrics != nil {
		fmt.Printf("\n=== Uncertainty Metrics ===\n")
		fmt.Printf("Uncertainty Health: %s\n", data.UncertaintyHealth)
		fmt.Printf("Overall Uncertainty: %.2f\n", data.UncertaintyMetrics.OverallUncertainty)
		fmt.Printf("Calibration Score: %.2f\n", data.UncertaintyMetrics.CalibrationScore)
		fmt.Printf("Reliability Score: %.2f\n", data.UncertaintyMetrics.ReliabilityScore)
	}

	// Display security metrics
	if data.SecurityMetrics != nil {
		fmt.Printf("\n=== Security Metrics ===\n")
		fmt.Printf("Security Health: %s\n", data.SecurityHealth)
		fmt.Printf("Overall Compliance: %.2f\n", data.SecurityMetrics.OverallCompliance)
		fmt.Printf("Data Source Trust Rate: %.2f\n", data.SecurityMetrics.DataSourceTrustRate)
		fmt.Printf("Website Verification Rate: %.2f\n", data.SecurityMetrics.WebsiteVerificationRate)
		fmt.Printf("Security Violation Rate: %.2f\n", data.SecurityMetrics.SecurityViolationRate)
	}

	// Display performance metrics
	if data.PerformanceMetrics != nil {
		fmt.Printf("\n=== Performance Metrics ===\n")
		fmt.Printf("Performance Health: %s\n", data.PerformanceHealth)
		fmt.Printf("Overall Accuracy: %.2f\n", data.PerformanceMetrics.OverallAccuracy)
		fmt.Printf("Average Latency: %v\n", data.PerformanceMetrics.AverageLatency)
		fmt.Printf("Throughput: %.2f\n", data.PerformanceMetrics.Throughput)
		fmt.Printf("Error Rate: %.2f\n", data.PerformanceMetrics.ErrorRate)
		fmt.Printf("CPU Usage: %.2f%%\n", data.PerformanceMetrics.CPUUsage)
		fmt.Printf("Memory Usage: %.2f%%\n", data.PerformanceMetrics.MemoryUsage)
		fmt.Printf("Cache Hit Rate: %.2f%%\n", data.PerformanceMetrics.CacheHitRate)
	}

	// Display alerts summary
	if data.AlertsSummary != nil {
		fmt.Printf("\n=== Alerts Summary ===\n")
		fmt.Printf("Total Alerts: %d\n", data.AlertsSummary.TotalAlerts)
		fmt.Printf("Critical Alerts: %d\n", data.AlertsSummary.CriticalAlerts)
		fmt.Printf("Warning Alerts: %d\n", data.AlertsSummary.WarningAlerts)
		fmt.Printf("Unacknowledged Alerts: %d\n", data.AlertsSummary.UnacknowledgedAlerts)
		fmt.Printf("Resolved Alerts: %d\n", data.AlertsSummary.ResolvedAlerts)
	}

	// Display recommendations
	if len(data.Recommendations) > 0 {
		fmt.Printf("\n=== Recommendations ===\n")
		for i, recommendation := range data.Recommendations {
			fmt.Printf("%d. %s\n", i+1, recommendation)
		}
	} else {
		fmt.Printf("\n=== Recommendations ===\n")
		fmt.Println("No recommendations at this time.")
	}

	fmt.Printf("\n=== Dashboard Export ===\n")

	// Export dashboard data as JSON
	jsonData, err := dashboard.ExportDashboardData(ctx, "json")
	if err != nil {
		logger.Error("Failed to export JSON data", zap.Error(err))
	} else {
		fmt.Printf("JSON export size: %d bytes\n", len(jsonData))
	}

	// Export dashboard data as YAML
	yamlData, err := dashboard.ExportDashboardData(ctx, "yaml")
	if err != nil {
		logger.Error("Failed to export YAML data", zap.Error(err))
	} else {
		fmt.Printf("YAML export size: %d bytes\n", len(yamlData))
	}

	fmt.Println("\n=== Advanced Monitoring Dashboard Example Complete ===")
}

// ExampleAdvancedMonitoringDashboardIntegration demonstrates integration with existing monitoring components
func ExampleAdvancedMonitoringDashboardIntegration() {
	// This example shows how to integrate the advanced monitoring dashboard
	// with existing monitoring components in a real application

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// In a real application, you would inject these components:
	// - metricsCollector: *MetricsCollector
	// - performanceMonitor: *PerformanceMonitor
	// - alertManager: *AlertManager
	// - healthChecker: *HealthChecker

	// For this example, we'll use nil values
	var (
		metricsCollector   *MetricsCollector
		performanceMonitor *PerformanceMonitor
		alertManager       *AlertManager
		healthChecker      *HealthChecker
	)

	// Create dashboard with default configuration
	dashboard := NewAdvancedMonitoringDashboard(
		DefaultAdvancedDashboardConfig(),
		metricsCollector,
		performanceMonitor,
		alertManager,
		healthChecker,
		logger,
	)

	// Simulate getting dashboard data periodically
	ctx := context.Background()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	fmt.Println("Starting advanced monitoring dashboard...")
	fmt.Println("Press Ctrl+C to stop")

	for {
		select {
		case <-ticker.C:
			data, err := dashboard.GetDashboardData(ctx)
			if err != nil {
				logger.Error("Failed to get dashboard data", zap.Error(err))
				continue
			}

			// Log dashboard status
			logger.Info("Dashboard status update",
				zap.String("health", data.OverallHealth),
				zap.Float64("score", data.HealthScore),
				zap.Int("total_alerts", data.AlertsSummary.TotalAlerts),
				zap.Int("critical_alerts", data.AlertsSummary.CriticalAlerts),
				zap.Int("recommendations", len(data.Recommendations)),
			)

			// In a real application, you would:
			// 1. Send data to a monitoring service (Prometheus, Grafana, etc.)
			// 2. Trigger alerts based on health status
			// 3. Update a web dashboard
			// 4. Store metrics in a time-series database

		case <-ctx.Done():
			fmt.Println("Stopping advanced monitoring dashboard...")
			return
		}
	}
}

