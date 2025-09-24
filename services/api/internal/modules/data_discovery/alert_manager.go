package data_discovery

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NewAlertManager creates a new alert manager
func NewAlertManager(config *ExtractionMonitorConfig, logger *zap.Logger) *AlertManager {
	return &AlertManager{
		config: config,
		logger: logger,
		alerts: make([]Alert, 0),
	}
}

// CreateAlert creates a new alert
func (am *AlertManager) CreateAlert(alertType, severity, message string, metrics interface{}) {
	// Check cooldown period
	if time.Since(am.lastAlert) < am.config.AlertSettings.AlertCooldownPeriod {
		am.logger.Debug("Alert suppressed due to cooldown period",
			zap.String("alert_type", alertType),
			zap.String("severity", severity))
		return
	}

	alert := Alert{
		ID:           am.generateAlertID(),
		Type:         alertType,
		Severity:     severity,
		Message:      message,
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
		Metrics:      am.extractMetrics(metrics),
	}

	am.mu.Lock()
	am.alerts = append(am.alerts, alert)
	am.mu.Unlock()

	am.lastAlert = time.Now()

	// Log the alert
	am.logger.Warn("Extraction monitoring alert created",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", alertType),
		zap.String("severity", severity),
		zap.String("message", message))

	// Send alert through configured channels
	am.sendAlert(alert)
}

// GetActiveAlerts returns all active (non-resolved) alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var activeAlerts []Alert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetAlertsByType returns alerts filtered by type
func (am *AlertManager) GetAlertsByType(alertType string) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var filteredAlerts []Alert
	for _, alert := range am.alerts {
		if alert.Type == alertType {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// GetAlertsBySeverity returns alerts filtered by severity
func (am *AlertManager) GetAlertsBySeverity(severity string) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var filteredAlerts []Alert
	for _, alert := range am.alerts {
		if alert.Severity == severity {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// AcknowledgeAlert acknowledges an alert by ID
func (am *AlertManager) AcknowledgeAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i := range am.alerts {
		if am.alerts[i].ID == alertID {
			am.alerts[i].Acknowledged = true
			am.logger.Info("Alert acknowledged",
				zap.String("alert_id", alertID))
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// ResolveAlert resolves an alert by ID
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i := range am.alerts {
		if am.alerts[i].ID == alertID {
			am.alerts[i].Resolved = true
			am.logger.Info("Alert resolved",
				zap.String("alert_id", alertID))
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// GetAlertSummary returns a summary of alert statistics
func (am *AlertManager) GetAlertSummary() *AlertSummary {
	am.mu.RLock()
	defer am.mu.RUnlock()

	summary := &AlertSummary{
		TotalAlerts:        len(am.alerts),
		ActiveAlerts:       0,
		AcknowledgedAlerts: 0,
		ResolvedAlerts:     0,
		AlertsByType:       make(map[string]int),
		AlertsBySeverity:   make(map[string]int),
		RecentAlerts:       make([]Alert, 0),
	}

	// Calculate statistics
	for _, alert := range am.alerts {
		// Count by status
		if !alert.Resolved {
			summary.ActiveAlerts++
		} else {
			summary.ResolvedAlerts++
		}

		if alert.Acknowledged {
			summary.AcknowledgedAlerts++
		}

		// Count by type
		summary.AlertsByType[alert.Type]++

		// Count by severity
		summary.AlertsBySeverity[alert.Severity]++

		// Get recent alerts (last 24 hours)
		if time.Since(alert.Timestamp) < 24*time.Hour {
			summary.RecentAlerts = append(summary.RecentAlerts, alert)
		}
	}

	return summary
}

// CleanupOldAlerts removes alerts older than the specified duration
func (am *AlertManager) CleanupOldAlerts(maxAge time.Duration) int {
	am.mu.Lock()
	defer am.mu.Unlock()

	var keptAlerts []Alert
	removedCount := 0

	for _, alert := range am.alerts {
		if time.Since(alert.Timestamp) < maxAge {
			keptAlerts = append(keptAlerts, alert)
		} else {
			removedCount++
		}
	}

	am.alerts = keptAlerts

	am.logger.Info("Cleaned up old alerts",
		zap.Int("removed_count", removedCount),
		zap.Int("remaining_count", len(am.alerts)))

	return removedCount
}

// GetAlertTrends returns alert trends over time
func (am *AlertManager) GetAlertTrends(duration time.Duration) *AlertTrends {
	am.mu.RLock()
	defer am.mu.RUnlock()

	trends := &AlertTrends{
		Period:           duration,
		TotalAlerts:      0,
		AlertsByHour:     make(map[int]int),
		AlertsByType:     make(map[string]int),
		AlertsBySeverity: make(map[string]int),
		PeakHour:         0,
		PeakCount:        0,
	}

	cutoffTime := time.Now().Add(-duration)

	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoffTime) {
			trends.TotalAlerts++

			// Count by hour
			hour := alert.Timestamp.Hour()
			trends.AlertsByHour[hour]++
			if trends.AlertsByHour[hour] > trends.PeakCount {
				trends.PeakCount = trends.AlertsByHour[hour]
				trends.PeakHour = hour
			}

			// Count by type
			trends.AlertsByType[alert.Type]++

			// Count by severity
			trends.AlertsBySeverity[alert.Severity]++
		}
	}

	return trends
}

// AlertSummary represents a summary of alert statistics
type AlertSummary struct {
	TotalAlerts        int            `json:"total_alerts"`
	ActiveAlerts       int            `json:"active_alerts"`
	AcknowledgedAlerts int            `json:"acknowledged_alerts"`
	ResolvedAlerts     int            `json:"resolved_alerts"`
	AlertsByType       map[string]int `json:"alerts_by_type"`
	AlertsBySeverity   map[string]int `json:"alerts_by_severity"`
	RecentAlerts       []Alert        `json:"recent_alerts"`
}

// AlertTrends represents alert trends over time
type AlertTrends struct {
	Period           time.Duration  `json:"period"`
	TotalAlerts      int            `json:"total_alerts"`
	AlertsByHour     map[int]int    `json:"alerts_by_hour"`
	AlertsByType     map[string]int `json:"alerts_by_type"`
	AlertsBySeverity map[string]int `json:"alerts_by_severity"`
	PeakHour         int            `json:"peak_hour"`
	PeakCount        int            `json:"peak_count"`
}

// Helper methods
func (am *AlertManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func (am *AlertManager) extractMetrics(metrics interface{}) map[string]interface{} {
	// Convert metrics to map for alert storage
	metricsMap := make(map[string]interface{})

	switch m := metrics.(type) {
	case *ExtractionMetrics:
		metricsMap["total_requests"] = m.TotalRequests
		metricsMap["successful_requests"] = m.SuccessfulRequests
		metricsMap["failed_requests"] = m.FailedRequests
		metricsMap["average_processing_time"] = m.AverageProcessingTime.String()
		metricsMap["average_quality_score"] = m.AverageQualityScore
		metricsMap["fields_discovered_per_request"] = m.FieldsDiscoveredPerRequest
		metricsMap["memory_usage_mb"] = m.MemoryUsage
		metricsMap["cpu_usage_percent"] = m.CPUUsage
		metricsMap["concurrent_requests"] = m.ConcurrentRequests
	case map[string]interface{}:
		metricsMap = m
	default:
		metricsMap["raw_metrics"] = fmt.Sprintf("%v", metrics)
	}

	return metricsMap
}

func (am *AlertManager) sendAlert(alert Alert) {
	if !am.config.AlertSettings.Enabled {
		return
	}

	// Send through configured channels
	for _, channel := range am.config.AlertSettings.AlertChannels {
		switch channel {
		case "log":
			am.sendLogAlert(alert)
		case "metrics":
			am.sendMetricsAlert(alert)
		case "email":
			am.sendEmailAlert(alert)
		case "webhook":
			am.sendWebhookAlert(alert)
		default:
			am.logger.Warn("Unknown alert channel",
				zap.String("channel", channel))
		}
	}
}

func (am *AlertManager) sendLogAlert(alert Alert) {
	// Log alert with appropriate level based on severity
	switch alert.Severity {
	case "critical":
		am.logger.Error("CRITICAL ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.Type),
			zap.String("message", alert.Message),
			zap.Any("metrics", alert.Metrics))
	case "warning":
		am.logger.Warn("WARNING ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.Type),
			zap.String("message", alert.Message),
			zap.Any("metrics", alert.Metrics))
	default:
		am.logger.Info("INFO ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.Type),
			zap.String("message", alert.Message),
			zap.Any("metrics", alert.Metrics))
	}
}

func (am *AlertManager) sendMetricsAlert(alert Alert) {
	// This would integrate with metrics systems like Prometheus, DataDog, etc.
	am.logger.Debug("Sending metrics alert",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", alert.Type))
}

func (am *AlertManager) sendEmailAlert(alert Alert) {
	// This would send email notifications
	am.logger.Debug("Sending email alert",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", alert.Type))
}

func (am *AlertManager) sendWebhookAlert(alert Alert) {
	// This would send webhook notifications
	am.logger.Debug("Sending webhook alert",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", alert.Type))
}

// GetAlertHistory returns alert history with pagination
func (am *AlertManager) GetAlertHistory(limit, offset int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Sort alerts by timestamp (newest first)
	alerts := make([]Alert, len(am.alerts))
	copy(alerts, am.alerts)

	// Sort by timestamp descending
	for i := 0; i < len(alerts)-1; i++ {
		for j := i + 1; j < len(alerts); j++ {
			if alerts[i].Timestamp.Before(alerts[j].Timestamp) {
				alerts[i], alerts[j] = alerts[j], alerts[i]
			}
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(alerts) {
		return []Alert{}
	}
	if end > len(alerts) {
		end = len(alerts)
	}

	return alerts[start:end]
}

// GetAlertByID returns a specific alert by ID
func (am *AlertManager) GetAlertByID(alertID string) (*Alert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, alert := range am.alerts {
		if alert.ID == alertID {
			return &alert, nil
		}
	}

	return nil, fmt.Errorf("alert not found: %s", alertID)
}

// UpdateAlert updates an alert's properties
func (am *AlertManager) UpdateAlert(alertID string, updates map[string]interface{}) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i := range am.alerts {
		if am.alerts[i].ID == alertID {
			// Apply updates
			if acknowledged, exists := updates["acknowledged"]; exists {
				if ack, ok := acknowledged.(bool); ok {
					am.alerts[i].Acknowledged = ack
				}
			}
			if resolved, exists := updates["resolved"]; exists {
				if res, ok := resolved.(bool); ok {
					am.alerts[i].Resolved = res
				}
			}
			if message, exists := updates["message"]; exists {
				if msg, ok := message.(string); ok {
					am.alerts[i].Message = msg
				}
			}

			am.logger.Info("Alert updated",
				zap.String("alert_id", alertID))
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// GetAlertStatistics returns detailed alert statistics
func (am *AlertManager) GetAlertStatistics() *AlertStatistics {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := &AlertStatistics{
		TotalAlerts:      len(am.alerts),
		AlertsByType:     make(map[string]AlertTypeStats),
		AlertsBySeverity: make(map[string]AlertSeverityStats),
		TimeDistribution: make(map[string]int),
		ResolutionTime:   make(map[string]time.Duration),
	}

	// Calculate statistics
	for _, alert := range am.alerts {
		// Type statistics
		typeStats := stats.AlertsByType[alert.Type]
		typeStats.Count++
		if alert.Resolved {
			typeStats.Resolved++
		}
		if alert.Acknowledged {
			typeStats.Acknowledged++
		}
		stats.AlertsByType[alert.Type] = typeStats

		// Severity statistics
		severityStats := stats.AlertsBySeverity[alert.Severity]
		severityStats.Count++
		if alert.Resolved {
			severityStats.Resolved++
		}
		if alert.Acknowledged {
			severityStats.Acknowledged++
		}
		stats.AlertsBySeverity[alert.Severity] = severityStats

		// Time distribution (by hour)
		hour := alert.Timestamp.Format("15")
		stats.TimeDistribution[hour]++

		// Resolution time (if resolved)
		if alert.Resolved {
			// This would need to track resolution time
			// For now, we'll use a placeholder
			stats.ResolutionTime[alert.Type] = time.Hour
		}
	}

	return stats
}

// AlertStatistics represents detailed alert statistics
type AlertStatistics struct {
	TotalAlerts      int                           `json:"total_alerts"`
	AlertsByType     map[string]AlertTypeStats     `json:"alerts_by_type"`
	AlertsBySeverity map[string]AlertSeverityStats `json:"alerts_by_severity"`
	TimeDistribution map[string]int                `json:"time_distribution"`
	ResolutionTime   map[string]time.Duration      `json:"resolution_time"`
}

// AlertTypeStats represents statistics for a specific alert type
type AlertTypeStats struct {
	Count        int `json:"count"`
	Resolved     int `json:"resolved"`
	Acknowledged int `json:"acknowledged"`
}

// AlertSeverityStats represents statistics for a specific alert severity
type AlertSeverityStats struct {
	Count        int `json:"count"`
	Resolved     int `json:"resolved"`
	Acknowledged int `json:"acknowledged"`
}
