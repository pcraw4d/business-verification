package error_monitoring

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// checkAlertConditions checks if alert conditions are met
func (erm *ErrorRateMonitor) checkAlertConditions(ctx context.Context, stats *ProcessErrorStats) {
	if !erm.config.EnableRealTimeAlerts || erm.alertManager == nil {
		return
	}

	now := time.Now()
	_, exists := erm.config.ProcessMonitoring[stats.ProcessName]
	if !exists {
		// Use default config values
	}

	// Check for critical error rate
	if stats.ErrorRate >= erm.config.CriticalErrorRate {
		if !stats.AlertStatus.Active || stats.AlertStatus.Level != "critical" {
			alert := Alert{
				ID:          fmt.Sprintf("critical_%s_%d", stats.ProcessName, now.Unix()),
				Type:        "error_rate",
				Level:       "critical",
				ProcessName: stats.ProcessName,
				Message:     fmt.Sprintf("Critical error rate detected for %s: %.2f%% (threshold: %.2f%%)", stats.ProcessName, stats.ErrorRate*100, erm.config.CriticalErrorRate*100),
				Timestamp:   now,
				Context: map[string]interface{}{
					"error_rate":     stats.ErrorRate,
					"threshold":      erm.config.CriticalErrorRate,
					"total_requests": stats.TotalRequests,
					"total_errors":   stats.TotalErrors,
					"trend":          stats.ErrorTrend,
				},
			}

			if err := erm.alertManager.SendAlert(ctx, alert); err != nil {
				erm.logger.Error("Failed to send critical alert", zap.Error(err))
			}

			stats.AlertStatus = AlertStatus{
				Level:                 "critical",
				Active:                true,
				LastTriggered:         now,
				Message:               alert.Message,
				ConsecutiveViolations: stats.AlertStatus.ConsecutiveViolations + 1,
			}
		}
	} else if stats.ErrorRate >= erm.config.WarningErrorRate {
		if !stats.AlertStatus.Active || stats.AlertStatus.Level != "warning" {
			alert := Alert{
				ID:          fmt.Sprintf("warning_%s_%d", stats.ProcessName, now.Unix()),
				Type:        "error_rate",
				Level:       "warning",
				ProcessName: stats.ProcessName,
				Message:     fmt.Sprintf("Warning error rate detected for %s: %.2f%% (threshold: %.2f%%)", stats.ProcessName, stats.ErrorRate*100, erm.config.WarningErrorRate*100),
				Timestamp:   now,
				Context: map[string]interface{}{
					"error_rate":     stats.ErrorRate,
					"threshold":      erm.config.WarningErrorRate,
					"total_requests": stats.TotalRequests,
					"total_errors":   stats.TotalErrors,
					"trend":          stats.ErrorTrend,
				},
			}

			if err := erm.alertManager.SendAlert(ctx, alert); err != nil {
				erm.logger.Error("Failed to send warning alert", zap.Error(err))
			}

			stats.AlertStatus = AlertStatus{
				Level:                 "warning",
				Active:                true,
				LastTriggered:         now,
				Message:               alert.Message,
				ConsecutiveViolations: stats.AlertStatus.ConsecutiveViolations + 1,
			}
		}
	}
}

// checkAlertClearance checks if alerts should be cleared
func (erm *ErrorRateMonitor) checkAlertClearance(ctx context.Context, stats *ProcessErrorStats) {
	if !stats.AlertStatus.Active || erm.alertManager == nil {
		return
	}

	now := time.Now()

	// Check if error rate is back to acceptable levels
	if stats.ErrorRate < erm.config.MaxErrorRate {
		// Clear alert
		alertID := fmt.Sprintf("%s_%s", stats.AlertStatus.Level, stats.ProcessName)
		if err := erm.alertManager.ClearAlert(ctx, alertID); err != nil {
			erm.logger.Error("Failed to clear alert", zap.Error(err))
		}

		stats.AlertStatus = AlertStatus{
			Level:                 "none",
			Active:                false,
			LastCleared:           now,
			ConsecutiveViolations: 0,
		}
	}
}

// updateGlobalStats updates global error statistics
func (erm *ErrorRateMonitor) updateGlobalStats() {
	var totalRequests int64
	var totalErrors int64

	erm.globalStats.ErrorRateByProcess = make(map[string]float64)
	erm.globalStats.CriticalProcesses = make([]string, 0)

	for processName, stats := range erm.processStats {
		totalRequests += stats.TotalRequests
		totalErrors += stats.TotalErrors
		erm.globalStats.ErrorRateByProcess[processName] = stats.ErrorRate

		// Identify critical processes
		if stats.ErrorRate > erm.config.CriticalErrorRate {
			erm.globalStats.CriticalProcesses = append(erm.globalStats.CriticalProcesses, processName)
		}

		// Update top error categories
		for category, count := range stats.ErrorsByCategory {
			erm.globalStats.TopErrorCategories[category] += count
		}
	}

	if totalRequests > 0 {
		erm.globalStats.OverallErrorRate = float64(totalErrors) / float64(totalRequests)
	}

	erm.globalStats.TotalRequests = totalRequests
	erm.globalStats.TotalErrors = totalErrors
	erm.globalStats.LastUpdated = time.Now()

	// Determine health status
	if erm.globalStats.OverallErrorRate > erm.config.CriticalErrorRate {
		erm.globalStats.HealthStatus = "critical"
	} else if erm.globalStats.OverallErrorRate > erm.config.WarningErrorRate {
		erm.globalStats.HealthStatus = "warning"
	} else if erm.globalStats.OverallErrorRate > erm.config.MaxErrorRate {
		erm.globalStats.HealthStatus = "degraded"
	} else {
		erm.globalStats.HealthStatus = "healthy"
	}
}

// generateTrendAnalysis generates trend analysis
func (erm *ErrorRateMonitor) generateTrendAnalysis() *ErrorRateTrendAnalysis {
	analysis := &ErrorRateTrendAnalysis{
		ProcessTrends: make(map[string]string),
		Seasonality:   make(map[string]float64),
	}

	// Calculate global trend
	if len(erm.processStats) > 0 {
		increasing := 0
		decreasing := 0
		stable := 0

		for _, stats := range erm.processStats {
			switch stats.ErrorTrend {
			case "increasing":
				increasing++
			case "decreasing":
				decreasing++
			default:
				stable++
			}
			analysis.ProcessTrends[stats.ProcessName] = stats.ErrorTrend
		}

		if increasing > decreasing && increasing > stable {
			analysis.GlobalTrend = "increasing"
		} else if decreasing > increasing && decreasing > stable {
			analysis.GlobalTrend = "decreasing"
		} else {
			analysis.GlobalTrend = "stable"
		}
	}

	// Simple prediction based on current trend
	if analysis.GlobalTrend == "increasing" {
		analysis.PredictedErrorRate = erm.globalStats.OverallErrorRate * 1.2
		analysis.TrendConfidence = 0.7
	} else if analysis.GlobalTrend == "decreasing" {
		analysis.PredictedErrorRate = erm.globalStats.OverallErrorRate * 0.8
		analysis.TrendConfidence = 0.7
	} else {
		analysis.PredictedErrorRate = erm.globalStats.OverallErrorRate
		analysis.TrendConfidence = 0.9
	}

	return analysis
}

// generateRecommendations generates recommendations based on error patterns
func (erm *ErrorRateMonitor) generateRecommendations() []string {
	recommendations := make([]string, 0)

	// Check global error rate
	if erm.globalStats.OverallErrorRate > erm.config.MaxErrorRate {
		recommendations = append(recommendations, "Global error rate exceeds target of 5%. Immediate investigation required.")
	}

	// Check individual processes
	for processName, stats := range erm.processStats {
		processConfig, exists := erm.config.ProcessMonitoring[processName]
		if !exists {
			processConfig = ProcessMonitoringConfig{MaxErrorRate: erm.config.MaxErrorRate}
		}

		if stats.ErrorRate > processConfig.MaxErrorRate {
			recommendations = append(recommendations, fmt.Sprintf("Process '%s' error rate (%.2f%%) exceeds threshold (%.2f%%). Review and optimize.", processName, stats.ErrorRate*100, processConfig.MaxErrorRate*100))
		}

		if stats.ErrorTrend == "increasing" {
			recommendations = append(recommendations, fmt.Sprintf("Process '%s' shows increasing error trend. Monitor closely and investigate root cause.", processName))
		}
	}

	// Check top error categories
	for category, count := range erm.globalStats.TopErrorCategories {
		if count > 10 { // Threshold for recommendations
			switch category {
			case "connectivity":
				recommendations = append(recommendations, "High number of connectivity errors detected. Check network stability and external service availability.")
			case "performance":
				recommendations = append(recommendations, "Performance-related errors detected. Consider scaling resources or optimizing slow operations.")
			case "data_quality":
				recommendations = append(recommendations, "Data quality issues detected. Improve input validation and data sanitization.")
			case "security":
				recommendations = append(recommendations, "Security-related errors detected. Review authentication and authorization mechanisms.")
			case "system":
				recommendations = append(recommendations, "System errors detected. Check application health and infrastructure stability.")
			case "external":
				recommendations = append(recommendations, "External service errors detected. Review third-party integrations and implement fallback mechanisms.")
			}
		}
	}

	return recommendations
}

// checkCompliance checks compliance with error rate targets
func (erm *ErrorRateMonitor) checkCompliance() ComplianceStatus {
	compliance := ComplianceStatus{
		IsCompliant:        true,
		ComplianceScore:    1.0,
		ViolatingProcesses: make([]string, 0),
	}

	// Check global compliance
	if erm.globalStats.OverallErrorRate > erm.config.MaxErrorRate {
		compliance.IsCompliant = false
		compliance.LastViolation = time.Now()
	}

	// Check process compliance
	violationCount := 0
	totalProcesses := len(erm.processStats)

	for processName, stats := range erm.processStats {
		processConfig, exists := erm.config.ProcessMonitoring[processName]
		if !exists {
			processConfig = ProcessMonitoringConfig{MaxErrorRate: erm.config.MaxErrorRate}
		}

		if stats.ErrorRate > processConfig.MaxErrorRate {
			compliance.IsCompliant = false
			compliance.ViolatingProcesses = append(compliance.ViolatingProcesses, processName)
			violationCount++

			if stats.LastErrorTime.After(compliance.LastViolation) {
				compliance.LastViolation = stats.LastErrorTime
			}
		}
	}

	// Calculate compliance score
	if totalProcesses > 0 {
		compliance.ComplianceScore = float64(totalProcesses-violationCount) / float64(totalProcesses)
	}

	return compliance
}
