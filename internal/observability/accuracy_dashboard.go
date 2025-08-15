package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DashboardMetric represents a metric displayed on the accuracy dashboard
type DashboardMetric struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Trend       string                 `json:"trend"` // "up", "down", "stable"
	TrendValue  float64                `json:"trend_value"`
	Status      string                 `json:"status"` // "good", "warning", "critical"
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DashboardWidget represents a widget on the accuracy dashboard
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "metric", "chart", "table", "alert"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    map[string]int         `json:"position"` // x, y, width, height
	Config      map[string]interface{} `json:"config"`
	Data        interface{}            `json:"data"`
	LastUpdated time.Time              `json:"last_updated"`
}

// AccuracyDashboard represents the accuracy tracking dashboard
type AccuracyDashboard struct {
	logger  *Logger
	metrics *Metrics

	// Dashboard components
	widgets      map[string]*DashboardWidget
	widgetsMutex sync.RWMutex

	// Performance monitoring
	performanceMetrics map[string]*DashboardMetric
	metricsMutex       sync.RWMutex

	// Alerting rules
	alertingRules map[string]*AlertingRule
	alertingMutex sync.RWMutex

	// Reporting
	reports      map[string]*AccuracyReport
	reportsMutex sync.RWMutex

	// Configuration
	enableDashboard     bool
	refreshInterval     time.Duration
	metricRetentionDays int
	maxWidgets          int
}

// AlertingRule represents an accuracy-based alerting rule
type AlertingRule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	MetricType    string                 `json:"metric_type"`
	Condition     string                 `json:"condition"` // "above", "below", "equals"
	Threshold     float64                `json:"threshold"`
	Severity      string                 `json:"severity"` // "low", "medium", "high", "critical"
	Enabled       bool                   `json:"enabled"`
	Actions       []string               `json:"actions"` // "email", "slack", "webhook"
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AccuracyReport represents an accuracy report for the dashboard
type AccuracyReport struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	ReportType      string                 `json:"report_type"` // "daily", "weekly", "monthly"
	TimeRange       time.Duration          `json:"time_range"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Data            map[string]interface{} `json:"data"`
	Summary         string                 `json:"summary"`
	Recommendations []string               `json:"recommendations"`
}

// NewAccuracyDashboard creates a new accuracy dashboard
func NewAccuracyDashboard(logger *Logger, metrics *Metrics) *AccuracyDashboard {
	dashboard := &AccuracyDashboard{
		logger:  logger,
		metrics: metrics,

		// Initialize storage
		widgets:            make(map[string]*DashboardWidget),
		performanceMetrics: make(map[string]*DashboardMetric),
		alertingRules:      make(map[string]*AlertingRule),
		reports:            make(map[string]*AccuracyReport),

		// Configuration
		enableDashboard:     true,
		refreshInterval:     time.Minute * 5,
		metricRetentionDays: 30,
		maxWidgets:          50,
	}

	// Initialize default widgets
	dashboard.initializeDefaultWidgets()

	// Initialize default alerting rules
	dashboard.initializeDefaultAlertingRules()

	// Start dashboard refresh loop
	go dashboard.startRefreshLoop()

	return dashboard
}

// GetDashboardData returns the complete dashboard data
func (ad *AccuracyDashboard) GetDashboardData(ctx context.Context) (map[string]interface{}, error) {
	ad.widgetsMutex.RLock()
	defer ad.widgetsMutex.RUnlock()
	ad.metricsMutex.RLock()
	defer ad.metricsMutex.RUnlock()

	// Log dashboard access
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "dashboard_data_requested", "", map[string]interface{}{
			"widgets_count": len(ad.widgets),
			"metrics_count": len(ad.performanceMetrics),
		})
	}

	dashboardData := map[string]interface{}{
		"widgets":          ad.getWidgetsData(),
		"metrics":          ad.getMetricsData(),
		"last_updated":     time.Now(),
		"refresh_interval": ad.refreshInterval.String(),
		"total_widgets":    len(ad.widgets),
		"total_metrics":    len(ad.performanceMetrics),
		"total_alerts":     len(ad.alertingRules),
	}

	return dashboardData, nil
}

// GetWidget retrieves a specific widget by ID
func (ad *AccuracyDashboard) GetWidget(ctx context.Context, widgetID string) (*DashboardWidget, error) {
	ad.widgetsMutex.RLock()
	defer ad.widgetsMutex.RUnlock()

	widget, exists := ad.widgets[widgetID]
	if !exists {
		return nil, fmt.Errorf("widget not found: %s", widgetID)
	}

	// Log widget access
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "widget_accessed", widgetID, map[string]interface{}{
			"widget_type":  widget.Type,
			"widget_title": widget.Title,
		})
	}

	return widget, nil
}

// AddWidget adds a new widget to the dashboard
func (ad *AccuracyDashboard) AddWidget(ctx context.Context, widget *DashboardWidget) error {
	ad.widgetsMutex.Lock()
	defer ad.widgetsMutex.Unlock()

	// Check widget limit
	if len(ad.widgets) >= ad.maxWidgets {
		return fmt.Errorf("maximum number of widgets reached: %d", ad.maxWidgets)
	}

	// Set widget metadata
	widget.LastUpdated = time.Now()
	if widget.Config == nil {
		widget.Config = make(map[string]interface{})
	}

	// Store widget
	ad.widgets[widget.ID] = widget

	// Log widget addition
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "widget_added", widget.ID, map[string]interface{}{
			"widget_type":  widget.Type,
			"widget_title": widget.Title,
		})
	}

	return nil
}

// UpdateWidget updates an existing widget
func (ad *AccuracyDashboard) UpdateWidget(ctx context.Context, widgetID string, updates map[string]interface{}) error {
	ad.widgetsMutex.Lock()
	defer ad.widgetsMutex.Unlock()

	widget, exists := ad.widgets[widgetID]
	if !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		widget.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		widget.Description = description
	}
	if position, ok := updates["position"].(map[string]int); ok {
		widget.Position = position
	}
	if config, ok := updates["config"].(map[string]interface{}); ok {
		widget.Config = config
	}
	if data, ok := updates["data"]; ok {
		widget.Data = data
	}

	// Update timestamp
	widget.LastUpdated = time.Now()

	// Log widget update
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "widget_updated", widgetID, map[string]interface{}{
			"updates_applied": len(updates),
		})
	}

	return nil
}

// RemoveWidget removes a widget from the dashboard
func (ad *AccuracyDashboard) RemoveWidget(ctx context.Context, widgetID string) error {
	ad.widgetsMutex.Lock()
	defer ad.widgetsMutex.Unlock()

	if _, exists := ad.widgets[widgetID]; !exists {
		return fmt.Errorf("widget not found: %s", widgetID)
	}

	delete(ad.widgets, widgetID)

	// Log widget removal
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "widget_removed", widgetID, nil)
	}

	return nil
}

// GetPerformanceMetrics returns performance metrics for the dashboard
func (ad *AccuracyDashboard) GetPerformanceMetrics(ctx context.Context) (map[string]*DashboardMetric, error) {
	ad.metricsMutex.RLock()
	defer ad.metricsMutex.RUnlock()

	// Log metrics access
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "performance_metrics_requested", "", map[string]interface{}{
			"metrics_count": len(ad.performanceMetrics),
		})
	}

	metrics := make(map[string]*DashboardMetric)
	for k, v := range ad.performanceMetrics {
		metrics[k] = v
	}

	return metrics, nil
}

// UpdatePerformanceMetric updates a performance metric
func (ad *AccuracyDashboard) UpdatePerformanceMetric(ctx context.Context, metricID string, value float64, trend string, status string) error {
	ad.metricsMutex.Lock()
	defer ad.metricsMutex.Unlock()

	metric, exists := ad.performanceMetrics[metricID]
	if !exists {
		metric = &DashboardMetric{
			ID:          metricID,
			LastUpdated: time.Now(),
			Metadata:    make(map[string]interface{}),
		}
		ad.performanceMetrics[metricID] = metric
	}

	// Update metric
	metric.Value = value
	metric.Trend = trend
	metric.Status = status
	metric.LastUpdated = time.Now()

	// Log metric update
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "performance_metric_updated", metricID, map[string]interface{}{
			"value":  value,
			"trend":  trend,
			"status": status,
		})
	}

	return nil
}

// GetAlertingRules returns alerting rules for the dashboard
func (ad *AccuracyDashboard) GetAlertingRules(ctx context.Context) (map[string]*AlertingRule, error) {
	ad.alertingMutex.RLock()
	defer ad.alertingMutex.RUnlock()

	rules := make(map[string]*AlertingRule)
	for k, v := range ad.alertingRules {
		rules[k] = v
	}

	return rules, nil
}

// AddAlertingRule adds a new alerting rule
func (ad *AccuracyDashboard) AddAlertingRule(ctx context.Context, rule *AlertingRule) error {
	ad.alertingMutex.Lock()
	defer ad.alertingMutex.Unlock()

	// Set rule metadata
	if rule.Metadata == nil {
		rule.Metadata = make(map[string]interface{})
	}

	// Store rule
	ad.alertingRules[rule.ID] = rule

	// Log rule addition
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "alerting_rule_added", rule.ID, map[string]interface{}{
			"rule_name": rule.Name,
			"severity":  rule.Severity,
		})
	}

	return nil
}

// UpdateAlertingRule updates an existing alerting rule
func (ad *AccuracyDashboard) UpdateAlertingRule(ctx context.Context, ruleID string, updates map[string]interface{}) error {
	ad.alertingMutex.Lock()
	defer ad.alertingMutex.Unlock()

	rule, exists := ad.alertingRules[ruleID]
	if !exists {
		return fmt.Errorf("alerting rule not found: %s", ruleID)
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		rule.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		rule.Description = description
	}
	if threshold, ok := updates["threshold"].(float64); ok {
		rule.Threshold = threshold
	}
	if severity, ok := updates["severity"].(string); ok {
		rule.Severity = severity
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if actions, ok := updates["actions"].([]string); ok {
		rule.Actions = actions
	}

	// Log rule update
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "alerting_rule_updated", ruleID, map[string]interface{}{
			"updates_applied": len(updates),
		})
	}

	return nil
}

// GenerateAccuracyReport generates an accuracy report
func (ad *AccuracyDashboard) GenerateAccuracyReport(ctx context.Context, reportType string, timeRange time.Duration) (*AccuracyReport, error) {
	start := time.Now()

	// Log report generation start
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "accuracy_report_generation_started", "", map[string]interface{}{
			"report_type": reportType,
			"time_range":  timeRange.String(),
		})
	}

	// Generate report data
	reportData := ad.generateReportData(reportType, timeRange)

	// Create report
	report := &AccuracyReport{
		ID:              ad.generateReportID(),
		Title:           fmt.Sprintf("%s Accuracy Report", reportType),
		Description:     fmt.Sprintf("Accuracy report for the last %s", timeRange.String()),
		ReportType:      reportType,
		TimeRange:       timeRange,
		GeneratedAt:     time.Now(),
		Data:            reportData,
		Summary:         ad.generateReportSummary(reportData),
		Recommendations: ad.generateRecommendations(reportData),
	}

	// Store report
	ad.reportsMutex.Lock()
	ad.reports[report.ID] = report
	ad.reportsMutex.Unlock()

	// Log report generation completion
	if ad.logger != nil {
		ad.logger.WithComponent("accuracy_dashboard").LogBusinessEvent(ctx, "accuracy_report_generated", report.ID, map[string]interface{}{
			"report_type":        reportType,
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	return report, nil
}

// GetAccuracyReports returns accuracy reports
func (ad *AccuracyDashboard) GetAccuracyReports(ctx context.Context, filters map[string]interface{}) ([]*AccuracyReport, error) {
	ad.reportsMutex.RLock()
	defer ad.reportsMutex.RUnlock()

	var reports []*AccuracyReport
	for _, report := range ad.reports {
		if ad.matchesReportFilters(report, filters) {
			reports = append(reports, report)
		}
	}

	return reports, nil
}

// GetDashboardStats returns statistics about the dashboard
func (ad *AccuracyDashboard) GetDashboardStats() map[string]interface{} {
	ad.widgetsMutex.RLock()
	defer ad.widgetsMutex.RUnlock()
	ad.metricsMutex.RLock()
	defer ad.metricsMutex.RUnlock()
	ad.alertingMutex.RLock()
	defer ad.alertingMutex.RUnlock()
	ad.reportsMutex.RLock()
	defer ad.reportsMutex.RUnlock()

	stats := map[string]interface{}{
		"total_widgets":           len(ad.widgets),
		"total_metrics":           len(ad.performanceMetrics),
		"total_alerting_rules":    len(ad.alertingRules),
		"total_reports":           len(ad.reports),
		"dashboard_enabled":       ad.enableDashboard,
		"refresh_interval":        ad.refreshInterval.String(),
		"metric_retention_days":   ad.metricRetentionDays,
		"max_widgets":             ad.maxWidgets,
		"widget_types":            make(map[string]int),
		"metric_status_breakdown": make(map[string]int),
		"rule_severity_breakdown": make(map[string]int),
	}

	// Calculate breakdowns
	for _, widget := range ad.widgets {
		stats["widget_types"].(map[string]int)[widget.Type]++
	}

	for _, metric := range ad.performanceMetrics {
		stats["metric_status_breakdown"].(map[string]int)[metric.Status]++
	}

	for _, rule := range ad.alertingRules {
		stats["rule_severity_breakdown"].(map[string]int)[rule.Severity]++
	}

	return stats
}

// Helper methods

// getWidgetsData returns widgets data for the dashboard
func (ad *AccuracyDashboard) getWidgetsData() []*DashboardWidget {
	var widgets []*DashboardWidget
	for _, widget := range ad.widgets {
		widgets = append(widgets, widget)
	}
	return widgets
}

// getMetricsData returns metrics data for the dashboard
func (ad *AccuracyDashboard) getMetricsData() map[string]*DashboardMetric {
	metrics := make(map[string]*DashboardMetric)
	for k, v := range ad.performanceMetrics {
		metrics[k] = v
	}
	return metrics
}

// generateReportData generates data for accuracy reports
func (ad *AccuracyDashboard) generateReportData(reportType string, timeRange time.Duration) map[string]interface{} {
	// This would typically fetch data from the accuracy validator
	// For now, return sample data
	return map[string]interface{}{
		"overall_accuracy": 0.85,
		"industry_breakdown": map[string]float64{
			"technology": 0.92,
			"retail":     0.78,
			"finance":    0.88,
		},
		"confidence_breakdown": map[string]float64{
			"high":   0.95,
			"medium": 0.82,
			"low":    0.65,
		},
		"total_classifications": 1250,
		"total_feedback":        89,
		"alerts_generated":      3,
	}
}

// generateReportSummary generates a summary for accuracy reports
func (ad *AccuracyDashboard) generateReportSummary(data map[string]interface{}) string {
	overallAccuracy, _ := data["overall_accuracy"].(float64)
	totalClassifications, _ := data["total_classifications"].(int)
	totalFeedback, _ := data["total_feedback"].(int)

	return fmt.Sprintf("Overall accuracy is %.1f%% with %d classifications and %d feedback submissions.",
		overallAccuracy*100, totalClassifications, totalFeedback)
}

// generateRecommendations generates recommendations based on report data
func (ad *AccuracyDashboard) generateRecommendations(data map[string]interface{}) []string {
	var recommendations []string

	overallAccuracy, _ := data["overall_accuracy"].(float64)
	if overallAccuracy < 0.8 {
		recommendations = append(recommendations, "Overall accuracy is below target. Consider reviewing classification algorithms.")
	}

	industryBreakdown, _ := data["industry_breakdown"].(map[string]float64)
	for industry, accuracy := range industryBreakdown {
		if accuracy < 0.7 {
			recommendations = append(recommendations, fmt.Sprintf("Accuracy for %s industry is below 70%%. Consider industry-specific improvements.", industry))
		}
	}

	confidenceBreakdown, _ := data["confidence_breakdown"].(map[string]float64)
	if lowAccuracy, exists := confidenceBreakdown["low"]; exists && lowAccuracy < 0.6 {
		recommendations = append(recommendations, "Low confidence classifications have poor accuracy. Consider improving confidence scoring.")
	}

	return recommendations
}

// matchesReportFilters checks if a report matches the specified filters
func (ad *AccuracyDashboard) matchesReportFilters(report *AccuracyReport, filters map[string]interface{}) bool {
	if filters == nil {
		return true
	}

	for key, value := range filters {
		switch key {
		case "report_type":
			if reportType, ok := value.(string); ok && report.ReportType != reportType {
				return false
			}
		case "time_range":
			if timeRange, ok := value.(time.Duration); ok && report.TimeRange != timeRange {
				return false
			}
		}
	}

	return true
}

// generateReportID generates a unique report ID
func (ad *AccuracyDashboard) generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

// startRefreshLoop starts the dashboard refresh loop
func (ad *AccuracyDashboard) startRefreshLoop() {
	ticker := time.NewTicker(ad.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Refresh dashboard data
		ad.refreshDashboardData()

		// Clean up old metrics
		ad.cleanupOldMetrics()
	}
}

// refreshDashboardData refreshes dashboard data
func (ad *AccuracyDashboard) refreshDashboardData() {
	// This would typically fetch fresh data from various sources
	// For now, just update timestamps
	ad.widgetsMutex.Lock()
	for _, widget := range ad.widgets {
		widget.LastUpdated = time.Now()
	}
	ad.widgetsMutex.Unlock()

	ad.metricsMutex.Lock()
	for _, metric := range ad.performanceMetrics {
		metric.LastUpdated = time.Now()
	}
	ad.metricsMutex.Unlock()
}

// cleanupOldMetrics removes metrics older than retention period
func (ad *AccuracyDashboard) cleanupOldMetrics() {
	ad.metricsMutex.Lock()
	defer ad.metricsMutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -ad.metricRetentionDays)
	for id, metric := range ad.performanceMetrics {
		if metric.LastUpdated.Before(cutoff) {
			delete(ad.performanceMetrics, id)
		}
	}
}

// initializeDefaultWidgets initializes default dashboard widgets
func (ad *AccuracyDashboard) initializeDefaultWidgets() {
	defaultWidgets := []*DashboardWidget{
		{
			ID:          "overall_accuracy",
			Type:        "metric",
			Title:       "Overall Accuracy",
			Description: "Overall classification accuracy",
			Position:    map[string]int{"x": 0, "y": 0, "width": 6, "height": 4},
			Config:      map[string]interface{}{"unit": "percentage", "decimals": 1},
			LastUpdated: time.Now(),
		},
		{
			ID:          "industry_accuracy",
			Type:        "chart",
			Title:       "Accuracy by Industry",
			Description: "Classification accuracy broken down by industry",
			Position:    map[string]int{"x": 6, "y": 0, "width": 6, "height": 4},
			Config:      map[string]interface{}{"chart_type": "bar"},
			LastUpdated: time.Now(),
		},
		{
			ID:          "confidence_accuracy",
			Type:        "chart",
			Title:       "Accuracy by Confidence",
			Description: "Classification accuracy by confidence level",
			Position:    map[string]int{"x": 0, "y": 4, "width": 6, "height": 4},
			Config:      map[string]interface{}{"chart_type": "line"},
			LastUpdated: time.Now(),
		},
		{
			ID:          "active_alerts",
			Type:        "table",
			Title:       "Active Alerts",
			Description: "Currently active accuracy alerts",
			Position:    map[string]int{"x": 6, "y": 4, "width": 6, "height": 4},
			Config:      map[string]interface{}{"columns": []string{"Severity", "Message", "Created"}},
			LastUpdated: time.Now(),
		},
	}

	for _, widget := range defaultWidgets {
		ad.widgets[widget.ID] = widget
	}
}

// initializeDefaultAlertingRules initializes default alerting rules
func (ad *AccuracyDashboard) initializeDefaultAlertingRules() {
	defaultRules := []*AlertingRule{
		{
			ID:          "overall_accuracy_low",
			Name:        "Overall Accuracy Below 80%",
			Description: "Alert when overall accuracy drops below 80%",
			MetricType:  "overall_accuracy",
			Condition:   "below",
			Threshold:   0.8,
			Severity:    "high",
			Enabled:     true,
			Actions:     []string{"email", "slack"},
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:          "industry_accuracy_low",
			Name:        "Industry Accuracy Below 70%",
			Description: "Alert when any industry accuracy drops below 70%",
			MetricType:  "industry_accuracy",
			Condition:   "below",
			Threshold:   0.7,
			Severity:    "medium",
			Enabled:     true,
			Actions:     []string{"slack"},
			Metadata:    make(map[string]interface{}),
		},
		{
			ID:          "confidence_accuracy_low",
			Name:        "Low Confidence Accuracy Below 60%",
			Description: "Alert when low confidence accuracy drops below 60%",
			MetricType:  "confidence_accuracy",
			Condition:   "below",
			Threshold:   0.6,
			Severity:    "low",
			Enabled:     true,
			Actions:     []string{"slack"},
			Metadata:    make(map[string]interface{}),
		},
	}

	for _, rule := range defaultRules {
		ad.alertingRules[rule.ID] = rule
	}
}
