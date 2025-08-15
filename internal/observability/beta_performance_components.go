package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/webanalysis.bak"
	"go.uber.org/zap"
	// "github.com/pcraw4d/business-verification/internal/webanalysis" // Temporarily disabled
)

// MetricsCollector collects and aggregates metrics for beta testing
type MetricsCollector struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	abTestResults map[string][]*webanalysis.ABTestResult // testID -> results
	metrics       map[string]*ABTestMetrics              // testID -> metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:        logger,
		abTestResults: make(map[string][]*webanalysis.ABTestResult),
		metrics:       make(map[string]*ABTestMetrics),
	}
}

// RecordABTestResult records an A/B test result
func (mc *MetricsCollector) RecordABTestResult(result *webanalysis.ABTestResult) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if result == nil {
		return fmt.Errorf("cannot record nil result")
	}

	// Add to results
	mc.abTestResults[result.TestID] = append(mc.abTestResults[result.TestID], result)

	// Invalidate cached metrics
	delete(mc.metrics, result.TestID)

	mc.logger.Debug("Recorded A/B test result",
		zap.String("test_id", result.TestID),
		zap.String("method", result.Method),
		zap.Bool("success", result.Success),
		zap.Duration("response_time", result.ResponseTime),
	)

	return nil
}

// GetABTestMetrics retrieves A/B test metrics for a time range
func (mc *MetricsCollector) GetABTestMetrics(ctx context.Context, timeRange time.Duration) (*ABTestMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := &ABTestMetrics{}

	// Collect all results within the time range
	var allResults []*webanalysis.ABTestResult
	cutoffTime := time.Now().Add(-timeRange)

	for _, results := range mc.abTestResults {
		for _, result := range results {
			if result.Timestamp.After(cutoffTime) {
				allResults = append(allResults, result)
			}
		}
	}

	if len(allResults) == 0 {
		return metrics, nil
	}

	// Separate basic and enhanced results
	var basicResults, enhancedResults []*webanalysis.ABTestResult
	for _, result := range allResults {
		if result.Method == "basic" {
			basicResults = append(basicResults, result)
		} else if result.Method == "enhanced" {
			enhancedResults = append(enhancedResults, result)
		}
	}

	metrics.TotalTests = len(allResults)

	// Calculate enhanced metrics
	if len(enhancedResults) > 0 {
		metrics.EnhancedSuccessRate = mc.calculateSuccessRate(enhancedResults)
		metrics.EnhancedAccuracy = mc.calculateAverageAccuracy(enhancedResults)
		metrics.EnhancedDataQuality = mc.calculateAverageDataQuality(enhancedResults)
		metrics.EnhancedAvgResponseTime = mc.calculateAverageResponseTime(enhancedResults)
	}

	// Calculate basic metrics
	if len(basicResults) > 0 {
		metrics.BasicSuccessRate = mc.calculateSuccessRate(basicResults)
		metrics.BasicAccuracy = mc.calculateAverageAccuracy(basicResults)
		metrics.BasicDataQuality = mc.calculateAverageDataQuality(basicResults)
		metrics.BasicAvgResponseTime = mc.calculateAverageResponseTime(basicResults)
	}

	return metrics, nil
}

// calculateSuccessRate calculates success rate for a set of results
func (mc *MetricsCollector) calculateSuccessRate(results []*webanalysis.ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(results))
}

// calculateAverageAccuracy calculates average accuracy
func (mc *MetricsCollector) calculateAverageAccuracy(results []*webanalysis.ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalAccuracy := 0.0
	for _, result := range results {
		totalAccuracy += result.Accuracy
	}

	return totalAccuracy / float64(len(results))
}

// calculateAverageDataQuality calculates average data quality
func (mc *MetricsCollector) calculateAverageDataQuality(results []*webanalysis.ABTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	for _, result := range results {
		totalQuality += result.DataQuality
	}

	return totalQuality / float64(len(results))
}

// calculateAverageResponseTime calculates average response time
func (mc *MetricsCollector) calculateAverageResponseTime(results []*webanalysis.ABTestResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	totalDuration := time.Duration(0)
	for _, result := range results {
		totalDuration += result.ResponseTime
	}

	return totalDuration / time.Duration(len(results))
}

// UserImpactTracker tracks user impact metrics
type UserImpactTracker struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	impacts map[string][]*UserImpactMetrics // userID -> impacts
}

// NewUserImpactTracker creates a new user impact tracker
func NewUserImpactTracker(logger *zap.Logger) *UserImpactTracker {
	return &UserImpactTracker{
		logger:  logger,
		impacts: make(map[string][]*UserImpactMetrics),
	}
}

// TrackImpact tracks user impact metrics
func (uit *UserImpactTracker) TrackImpact(impact *UserImpactMetrics) error {
	uit.mu.Lock()
	defer uit.mu.Unlock()

	if impact == nil {
		return fmt.Errorf("cannot track nil impact")
	}

	// Add to impacts
	uit.impacts[impact.UserID] = append(uit.impacts[impact.UserID], impact)

	uit.logger.Debug("Tracked user impact",
		zap.String("user_id", impact.UserID),
		zap.String("test_id", impact.TestID),
		zap.String("method", impact.Method),
		zap.Bool("success", impact.Success),
		zap.Float64("impact_score", impact.ImpactScore),
	)

	return nil
}

// GetImpactMetrics retrieves user impact metrics for a time range
func (uit *UserImpactTracker) GetImpactMetrics(ctx context.Context, timeRange time.Duration) ([]*UserImpactMetrics, error) {
	uit.mu.RLock()
	defer uit.mu.RUnlock()

	var allImpacts []*UserImpactMetrics
	cutoffTime := time.Now().Add(-timeRange)

	for _, impacts := range uit.impacts {
		for _, impact := range impacts {
			if impact.Timestamp.After(cutoffTime) {
				allImpacts = append(allImpacts, impact)
			}
		}
	}

	return allImpacts, nil
}

// BetaCostTracker tracks cost metrics for beta features
type BetaCostTracker struct {
	logger *zap.Logger
	mu     sync.RWMutex
	costs  map[string][]*BetaCostMetrics // testID -> costs
}

// NewBetaCostTracker creates a new beta cost tracker
func NewBetaCostTracker(logger *zap.Logger) *BetaCostTracker {
	return &BetaCostTracker{
		logger: logger,
		costs:  make(map[string][]*BetaCostMetrics),
	}
}

// TrackCosts tracks cost metrics
func (bct *BetaCostTracker) TrackCosts(cost *BetaCostMetrics) error {
	bct.mu.Lock()
	defer bct.mu.Unlock()

	if cost == nil {
		return fmt.Errorf("cannot track nil cost")
	}

	// Add to costs
	bct.costs[cost.TestID] = append(bct.costs[cost.TestID], cost)

	bct.logger.Debug("Tracked cost metrics",
		zap.String("test_id", cost.TestID),
		zap.String("method", cost.Method),
		zap.Float64("total_cost", cost.TotalCost),
		zap.Float64("cost_per_request", cost.CostPerRequest),
	)

	return nil
}

// GetCostMetrics retrieves cost metrics for a time range
func (bct *BetaCostTracker) GetCostMetrics(ctx context.Context, timeRange time.Duration) (*CostMetrics, error) {
	bct.mu.RLock()
	defer bct.mu.RUnlock()

	metrics := &CostMetrics{
		CostByMethod: make(map[string]float64),
	}

	// Collect all costs within the time range
	var allCosts []*BetaCostMetrics
	cutoffTime := time.Now().Add(-timeRange)

	for _, costs := range bct.costs {
		for _, cost := range costs {
			if cost.Timestamp.After(cutoffTime) {
				allCosts = append(allCosts, cost)
			}
		}
	}

	if len(allCosts) == 0 {
		return metrics, nil
	}

	// Calculate total metrics
	totalCost := 0.0
	totalRequests := 0
	costByMethod := make(map[string]float64)
	methodRequests := make(map[string]int)

	for _, cost := range allCosts {
		totalCost += cost.TotalCost
		totalRequests += cost.RequestCount
		costByMethod[cost.Method] += cost.TotalCost
		methodRequests[cost.Method] += cost.RequestCount
	}

	metrics.TotalCost = totalCost
	metrics.TotalRequests = totalRequests
	metrics.AverageCostPerRequest = totalCost / float64(totalRequests)
	metrics.CostByMethod = costByMethod

	return metrics, nil
}

// AlertManager manages performance alerts
type AlertManager struct {
	logger *zap.Logger
	mu     sync.RWMutex
	alerts map[string]*BetaPerformanceAlert // alertID -> alert
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		logger: logger,
		alerts: make(map[string]*BetaPerformanceAlert),
	}
}

// SendAlert sends a performance alert
func (am *AlertManager) SendAlert(ctx context.Context, alert *BetaPerformanceAlert) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert == nil {
		return fmt.Errorf("cannot send nil alert")
	}

	// Store alert
	am.alerts[alert.AlertID] = alert

	am.logger.Warn("Performance alert sent",
		zap.String("alert_id", alert.AlertID),
		zap.String("alert_type", alert.AlertType),
		zap.String("severity", alert.Severity),
		zap.String("message", alert.Message),
		zap.String("metric", alert.Metric),
		zap.Float64("current_value", alert.CurrentValue),
		zap.Float64("threshold", alert.Threshold),
		zap.String("test_id", alert.TestID),
		zap.String("user_id", alert.UserID),
	)

	return nil
}

// ResolveAlert resolves a performance alert
func (am *AlertManager) ResolveAlert(ctx context.Context, alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	am.logger.Info("Performance alert resolved",
		zap.String("alert_id", alertID),
		zap.String("alert_type", alert.AlertType),
		zap.String("severity", alert.Severity),
	)

	return nil
}

// GetAlertSummary retrieves alert summary for a time range
func (am *AlertManager) GetAlertSummary(ctx context.Context, timeRange time.Duration) (*AlertSummary, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	summary := &AlertSummary{}
	cutoffTime := time.Now().Add(-timeRange)

	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoffTime) {
			summary.TotalAlerts++

			switch alert.Severity {
			case "error":
				summary.ErrorAlerts++
			case "warning":
				summary.WarningAlerts++
			case "info":
				summary.InfoAlerts++
			}

			if alert.Resolved {
				summary.ResolvedAlerts++
			}
		}
	}

	return summary, nil
}

// GetAlerts retrieves all alerts for a time range
func (am *AlertManager) GetAlerts(ctx context.Context, timeRange time.Duration) ([]*BetaPerformanceAlert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var alerts []*BetaPerformanceAlert
	cutoffTime := time.Now().Add(-timeRange)

	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoffTime) {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// ExportMetrics exports metrics to JSON format
func (mc *MetricsCollector) ExportMetrics(ctx context.Context, timeRange time.Duration) ([]byte, error) {
	metrics, err := mc.GetABTestMetrics(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}

// ExportImpactMetrics exports impact metrics to JSON format
func (uit *UserImpactTracker) ExportImpactMetrics(ctx context.Context, timeRange time.Duration) ([]byte, error) {
	metrics, err := uit.GetImpactMetrics(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}

// ExportCostMetrics exports cost metrics to JSON format
func (bct *BetaCostTracker) ExportCostMetrics(ctx context.Context, timeRange time.Duration) ([]byte, error) {
	metrics, err := bct.GetCostMetrics(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}

// ExportAlerts exports alerts to JSON format
func (am *AlertManager) ExportAlerts(ctx context.Context, timeRange time.Duration) ([]byte, error) {
	alerts, err := am.GetAlerts(ctx, timeRange)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(alerts, "", "  ")
}
