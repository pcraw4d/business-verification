package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/webanalysis"
	"go.uber.org/zap"
)

// BetaPerformanceMonitor monitors performance and reliability for beta features
type BetaPerformanceMonitor struct {
	logger           *zap.Logger
	betaFramework    *webanalysis.BetaTestingFramework
	alertManager     *AlertManager
	metricsCollector *MetricsCollector
	benchmarks       *BetaBenchmarks
	impactTracker    *UserImpactTracker
	costTracker      *BetaCostTracker
	mu               sync.RWMutex
}

// BetaBenchmarks defines performance benchmarks for beta features
type BetaBenchmarks struct {
	EnhancedScrapingSuccessRate  float64       `json:"enhanced_scraping_success_rate"`
	EnhancedScrapingResponseTime time.Duration `json:"enhanced_scraping_response_time"`
	EnhancedScrapingAccuracy     float64       `json:"enhanced_scraping_accuracy"`
	EnhancedScrapingDataQuality  float64       `json:"enhanced_scraping_data_quality"`
	UserSatisfactionThreshold    float64       `json:"user_satisfaction_threshold"`
	CostPerRequestThreshold      float64       `json:"cost_per_request_threshold"`
	ConcurrentRequestsThreshold  int           `json:"concurrent_requests_threshold"`
	ErrorRateThreshold           float64       `json:"error_rate_threshold"`
}

// UserImpactMetrics tracks user impact of beta features
type UserImpactMetrics struct {
	UserID       string        `json:"user_id"`
	TestID       string        `json:"test_id"`
	Method       string        `json:"method"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Accuracy     float64       `json:"accuracy"`
	DataQuality  float64       `json:"data_quality"`
	Satisfaction int           `json:"satisfaction"`
	ErrorType    string        `json:"error_type,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ImpactScore  float64       `json:"impact_score"`
	Timestamp    time.Time     `json:"timestamp"`
}

// BetaCostMetrics tracks cost metrics for beta features
type BetaCostMetrics struct {
	TestID         string    `json:"test_id"`
	Method         string    `json:"method"`
	ProxyCost      float64   `json:"proxy_cost"`
	ComputingCost  float64   `json:"computing_cost"`
	StorageCost    float64   `json:"storage_cost"`
	TotalCost      float64   `json:"total_cost"`
	CostPerRequest float64   `json:"cost_per_request"`
	RequestCount   int       `json:"request_count"`
	Timestamp      time.Time `json:"timestamp"`
}

// BetaPerformanceAlert represents a performance alert for beta features
type BetaPerformanceAlert struct {
	AlertID      string     `json:"alert_id"`
	AlertType    string     `json:"alert_type"`
	Severity     string     `json:"severity"`
	Message      string     `json:"message"`
	Metric       string     `json:"metric"`
	CurrentValue float64    `json:"current_value"`
	Threshold    float64    `json:"threshold"`
	TestID       string     `json:"test_id,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
	Resolved     bool       `json:"resolved"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// NewBetaPerformanceMonitor creates a new beta performance monitor
func NewBetaPerformanceMonitor(
	logger *zap.Logger,
	betaFramework *webanalysis.BetaTestingFramework,
	alertManager *AlertManager,
) *BetaPerformanceMonitor {
	return &BetaPerformanceMonitor{
		logger:           logger,
		betaFramework:    betaFramework,
		alertManager:     alertManager,
		metricsCollector: NewMetricsCollector(logger),
		benchmarks:       NewDefaultBetaBenchmarks(),
		impactTracker:    NewUserImpactTracker(logger),
		costTracker:      NewBetaCostTracker(logger),
	}
}

// NewDefaultBetaBenchmarks creates default performance benchmarks
func NewDefaultBetaBenchmarks() *BetaBenchmarks {
	return &BetaBenchmarks{
		EnhancedScrapingSuccessRate:  0.95,            // 95% success rate
		EnhancedScrapingResponseTime: 5 * time.Second, // 5 seconds max
		EnhancedScrapingAccuracy:     0.90,            // 90% accuracy
		EnhancedScrapingDataQuality:  0.85,            // 85% data quality
		UserSatisfactionThreshold:    4.0,             // 4.0/5.0 satisfaction
		CostPerRequestThreshold:      0.10,            // $0.10 per request
		ConcurrentRequestsThreshold:  50,              // 50 concurrent requests
		ErrorRateThreshold:           0.05,            // 5% error rate
	}
}

// MonitorABTestResult monitors an A/B test result and checks against benchmarks
func (bpm *BetaPerformanceMonitor) MonitorABTestResult(ctx context.Context, result *webanalysis.ABTestResult) error {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()

	// Record metrics
	bpm.metricsCollector.RecordABTestResult(result)

	// Check performance against benchmarks
	alerts := bpm.checkPerformanceBenchmarks(result)
	for _, alert := range alerts {
		bpm.alertManager.SendAlert(ctx, alert)
	}

	// Track user impact
	impactMetrics := bpm.createUserImpactMetrics(result)
	bpm.impactTracker.TrackImpact(impactMetrics)

	// Track costs
	costMetrics := bpm.estimateCosts(result)
	bpm.costTracker.TrackCosts(costMetrics)

	bpm.logger.Info("Monitored A/B test result",
		zap.String("test_id", result.TestID),
		zap.String("method", result.Method),
		zap.Bool("success", result.Success),
		zap.Duration("response_time", result.ResponseTime),
		zap.Float64("accuracy", result.Accuracy),
		zap.Float64("data_quality", result.DataQuality),
		zap.Int("alerts_generated", len(alerts)),
	)

	return nil
}

// MonitorUserFeedback monitors user feedback and checks satisfaction
func (bpm *BetaPerformanceMonitor) MonitorUserFeedback(ctx context.Context, feedback *webanalysis.BetaFeedback) error {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()

	// Check satisfaction threshold
	if float64(feedback.Satisfaction) < bpm.benchmarks.UserSatisfactionThreshold {
		alert := &BetaPerformanceAlert{
			AlertID:      bpm.generateAlertID("satisfaction"),
			AlertType:    "user_satisfaction",
			Severity:     "warning",
			Message:      fmt.Sprintf("User satisfaction below threshold: %d/5", feedback.Satisfaction),
			Metric:       "user_satisfaction",
			CurrentValue: float64(feedback.Satisfaction),
			Threshold:    bpm.benchmarks.UserSatisfactionThreshold,
			UserID:       feedback.UserID,
			TestID:       feedback.TestID,
			Timestamp:    time.Now(),
		}
		bpm.alertManager.SendAlert(ctx, alert)
	}

	// Track user impact
	impactMetrics := &UserImpactMetrics{
		UserID:       feedback.UserID,
		TestID:       feedback.TestID,
		Method:       feedback.Method,
		Success:      true, // Feedback submission is successful
		Satisfaction: feedback.Satisfaction,
		ImpactScore:  bpm.calculateImpactScore(feedback),
		Timestamp:    time.Now(),
	}
	bpm.impactTracker.TrackImpact(impactMetrics)

	bpm.logger.Info("Monitored user feedback",
		zap.String("user_id", feedback.UserID),
		zap.String("test_id", feedback.TestID),
		zap.String("method", feedback.Method),
		zap.Int("satisfaction", feedback.Satisfaction),
		zap.Int("accuracy", feedback.Accuracy),
		zap.Int("speed", feedback.Speed),
	)

	return nil
}

// GetPerformanceReport generates a comprehensive performance report
func (bpm *BetaPerformanceMonitor) GetPerformanceReport(ctx context.Context, timeRange time.Duration) (*BetaPerformanceReport, error) {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()

	report := &BetaPerformanceReport{
		Generated:  time.Now(),
		TimeRange:  timeRange,
		Benchmarks: bpm.benchmarks,
	}

	// Get A/B test metrics
	abMetrics, err := bpm.metricsCollector.GetABTestMetrics(ctx, timeRange)
	if err == nil {
		report.ABTestMetrics = abMetrics
	}

	// Get user impact metrics
	impactMetrics, err := bpm.impactTracker.GetImpactMetrics(ctx, timeRange)
	if err == nil {
		report.UserImpactMetrics = impactMetrics
	}

	// Get cost metrics
	costMetrics, err := bpm.costTracker.GetCostMetrics(ctx, timeRange)
	if err == nil {
		report.CostMetrics = costMetrics
	}

	// Get alert summary
	alertSummary, err := bpm.alertManager.GetAlertSummary(ctx, timeRange)
	if err == nil {
		report.AlertSummary = alertSummary
	}

	// Calculate overall performance score
	report.OverallPerformanceScore = bpm.calculateOverallPerformanceScore(report)

	return report, nil
}

// UpdateBenchmarks updates performance benchmarks
func (bpm *BetaPerformanceMonitor) UpdateBenchmarks(benchmarks *BetaBenchmarks) error {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()

	bpm.benchmarks = benchmarks

	bpm.logger.Info("Updated beta performance benchmarks",
		zap.Float64("enhanced_success_rate", benchmarks.EnhancedScrapingSuccessRate),
		zap.Duration("enhanced_response_time", benchmarks.EnhancedScrapingResponseTime),
		zap.Float64("enhanced_accuracy", benchmarks.EnhancedScrapingAccuracy),
		zap.Float64("enhanced_data_quality", benchmarks.EnhancedScrapingDataQuality),
		zap.Float64("user_satisfaction_threshold", benchmarks.UserSatisfactionThreshold),
		zap.Float64("cost_per_request_threshold", benchmarks.CostPerRequestThreshold),
		zap.Int("concurrent_requests_threshold", benchmarks.ConcurrentRequestsThreshold),
		zap.Float64("error_rate_threshold", benchmarks.ErrorRateThreshold),
	)

	return nil
}

// GetBenchmarks returns current performance benchmarks
func (bpm *BetaPerformanceMonitor) GetBenchmarks() *BetaBenchmarks {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()

	return bpm.benchmarks
}

// checkPerformanceBenchmarks checks performance against benchmarks and generates alerts
func (bpm *BetaPerformanceMonitor) checkPerformanceBenchmarks(result *webanalysis.ABTestResult) []*BetaPerformanceAlert {
	var alerts []*BetaPerformanceAlert

	// Only check enhanced scraping results
	if result.Method != "enhanced" {
		return alerts
	}

	// Check success rate
	if !result.Success {
		alert := &BetaPerformanceAlert{
			AlertID:      bpm.generateAlertID("success"),
			AlertType:    "scraping_failure",
			Severity:     "error",
			Message:      "Enhanced scraping failed",
			Metric:       "success_rate",
			CurrentValue: 0.0,
			Threshold:    bpm.benchmarks.EnhancedScrapingSuccessRate,
			TestID:       result.TestID,
			Timestamp:    time.Now(),
		}
		alerts = append(alerts, alert)
	}

	// Check response time
	if result.ResponseTime > bpm.benchmarks.EnhancedScrapingResponseTime {
		alert := &BetaPerformanceAlert{
			AlertID:      bpm.generateAlertID("response_time"),
			AlertType:    "slow_response",
			Severity:     "warning",
			Message:      fmt.Sprintf("Response time exceeded threshold: %v", result.ResponseTime),
			Metric:       "response_time",
			CurrentValue: float64(result.ResponseTime.Milliseconds()) / 1000.0,
			Threshold:    float64(bpm.benchmarks.EnhancedScrapingResponseTime.Milliseconds()) / 1000.0,
			TestID:       result.TestID,
			Timestamp:    time.Now(),
		}
		alerts = append(alerts, alert)
	}

	// Check accuracy
	if result.Accuracy < bpm.benchmarks.EnhancedScrapingAccuracy {
		alert := &BetaPerformanceAlert{
			AlertID:      bpm.generateAlertID("accuracy"),
			AlertType:    "low_accuracy",
			Severity:     "warning",
			Message:      fmt.Sprintf("Accuracy below threshold: %.2f", result.Accuracy),
			Metric:       "accuracy",
			CurrentValue: result.Accuracy,
			Threshold:    bpm.benchmarks.EnhancedScrapingAccuracy,
			TestID:       result.TestID,
			Timestamp:    time.Now(),
		}
		alerts = append(alerts, alert)
	}

	// Check data quality
	if result.DataQuality < bpm.benchmarks.EnhancedScrapingDataQuality {
		alert := &BetaPerformanceAlert{
			AlertID:      bpm.generateAlertID("data_quality"),
			AlertType:    "low_data_quality",
			Severity:     "warning",
			Message:      fmt.Sprintf("Data quality below threshold: %.2f", result.DataQuality),
			Metric:       "data_quality",
			CurrentValue: result.DataQuality,
			Threshold:    bpm.benchmarks.EnhancedScrapingDataQuality,
			TestID:       result.TestID,
			Timestamp:    time.Now(),
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

// createUserImpactMetrics creates user impact metrics from A/B test result
func (bpm *BetaPerformanceMonitor) createUserImpactMetrics(result *webanalysis.ABTestResult) *UserImpactMetrics {
	impactScore := bpm.calculateImpactScoreFromResult(result)

	return &UserImpactMetrics{
		TestID:       result.TestID,
		Method:       result.Method,
		Success:      result.Success,
		ResponseTime: result.ResponseTime,
		Accuracy:     result.Accuracy,
		DataQuality:  result.DataQuality,
		ImpactScore:  impactScore,
		Timestamp:    time.Now(),
	}
}

// estimateCosts estimates costs for an A/B test result
func (bpm *BetaPerformanceMonitor) estimateCosts(result *webanalysis.ABTestResult) *BetaCostMetrics {
	// Simple cost estimation - in a real implementation, this would be more sophisticated
	proxyCost := 0.01     // $0.01 per request for proxy
	computingCost := 0.02 // $0.02 per request for computing
	storageCost := 0.005  // $0.005 per request for storage
	totalCost := proxyCost + computingCost + storageCost

	return &BetaCostMetrics{
		TestID:         result.TestID,
		Method:         result.Method,
		ProxyCost:      proxyCost,
		ComputingCost:  computingCost,
		StorageCost:    storageCost,
		TotalCost:      totalCost,
		CostPerRequest: totalCost,
		RequestCount:   1,
		Timestamp:      time.Now(),
	}
}

// calculateImpactScore calculates impact score from feedback
func (bpm *BetaPerformanceMonitor) calculateImpactScore(feedback *webanalysis.BetaFeedback) float64 {
	// Weighted combination of satisfaction, accuracy, and speed
	satisfactionWeight := 0.4
	accuracyWeight := 0.35
	speedWeight := 0.25

	satisfactionScore := float64(feedback.Satisfaction) / 5.0
	accuracyScore := float64(feedback.Accuracy) / 5.0
	speedScore := float64(feedback.Speed) / 5.0

	impactScore := satisfactionScore*satisfactionWeight +
		accuracyScore*accuracyWeight +
		speedScore*speedWeight

	return impactScore
}

// calculateImpactScoreFromResult calculates impact score from A/B test result
func (bpm *BetaPerformanceMonitor) calculateImpactScoreFromResult(result *webanalysis.ABTestResult) float64 {
	// Weighted combination of success, accuracy, data quality, and response time
	successWeight := 0.4
	accuracyWeight := 0.3
	qualityWeight := 0.2
	speedWeight := 0.1

	successScore := 0.0
	if result.Success {
		successScore = 1.0
	}

	// Normalize response time (faster is better)
	speedScore := 1.0 - (float64(result.ResponseTime.Milliseconds()) / 10000.0) // Normalize to 10 seconds
	if speedScore < 0 {
		speedScore = 0
	}

	impactScore := successScore*successWeight +
		result.Accuracy*accuracyWeight +
		result.DataQuality*qualityWeight +
		speedScore*speedWeight

	return impactScore
}

// calculateOverallPerformanceScore calculates overall performance score
func (bpm *BetaPerformanceMonitor) calculateOverallPerformanceScore(report *BetaPerformanceReport) float64 {
	// Weighted combination of different metrics
	weights := map[string]float64{
		"ab_test_performance": 0.4,
		"user_impact":         0.3,
		"cost_efficiency":     0.2,
		"alert_health":        0.1,
	}

	abTestScore := 0.0
	if report.ABTestMetrics != nil {
		abTestScore = (report.ABTestMetrics.EnhancedSuccessRate +
			report.ABTestMetrics.EnhancedAccuracy +
			report.ABTestMetrics.EnhancedDataQuality) / 3.0
	}

	userImpactScore := 0.0
	if report.UserImpactMetrics != nil && len(report.UserImpactMetrics) > 0 {
		totalImpact := 0.0
		for _, impact := range report.UserImpactMetrics {
			totalImpact += impact.ImpactScore
		}
		userImpactScore = totalImpact / float64(len(report.UserImpactMetrics))
	}

	costEfficiencyScore := 0.0
	if report.CostMetrics != nil {
		// Lower cost is better, so invert the score
		costEfficiencyScore = 1.0 - (report.CostMetrics.AverageCostPerRequest / bpm.benchmarks.CostPerRequestThreshold)
		if costEfficiencyScore < 0 {
			costEfficiencyScore = 0
		}
	}

	alertHealthScore := 0.0
	if report.AlertSummary != nil {
		// Fewer alerts is better
		totalAlerts := report.AlertSummary.TotalAlerts
		if totalAlerts == 0 {
			alertHealthScore = 1.0
		} else {
			alertHealthScore = 1.0 - (float64(totalAlerts) / 100.0) // Normalize to 100 alerts
			if alertHealthScore < 0 {
				alertHealthScore = 0
			}
		}
	}

	overallScore := abTestScore*weights["ab_test_performance"] +
		userImpactScore*weights["user_impact"] +
		costEfficiencyScore*weights["cost_efficiency"] +
		alertHealthScore*weights["alert_health"]

	return overallScore
}

// generateAlertID generates a unique alert ID
func (bpm *BetaPerformanceMonitor) generateAlertID(alertType string) string {
	return fmt.Sprintf("beta_alert_%s_%d", alertType, time.Now().UnixNano())
}

// BetaPerformanceReport represents a comprehensive performance report
type BetaPerformanceReport struct {
	Generated               time.Time            `json:"generated"`
	TimeRange               time.Duration        `json:"time_range"`
	Benchmarks              *BetaBenchmarks      `json:"benchmarks"`
	ABTestMetrics           *ABTestMetrics       `json:"ab_test_metrics,omitempty"`
	UserImpactMetrics       []*UserImpactMetrics `json:"user_impact_metrics,omitempty"`
	CostMetrics             *CostMetrics         `json:"cost_metrics,omitempty"`
	AlertSummary            *AlertSummary        `json:"alert_summary,omitempty"`
	OverallPerformanceScore float64              `json:"overall_performance_score"`
}

// ABTestMetrics represents A/B test performance metrics
type ABTestMetrics struct {
	TotalTests              int           `json:"total_tests"`
	EnhancedSuccessRate     float64       `json:"enhanced_success_rate"`
	EnhancedAccuracy        float64       `json:"enhanced_accuracy"`
	EnhancedDataQuality     float64       `json:"enhanced_data_quality"`
	EnhancedAvgResponseTime time.Duration `json:"enhanced_avg_response_time"`
	BasicSuccessRate        float64       `json:"basic_success_rate"`
	BasicAccuracy           float64       `json:"basic_accuracy"`
	BasicDataQuality        float64       `json:"basic_data_quality"`
	BasicAvgResponseTime    time.Duration `json:"basic_avg_response_time"`
}

// CostMetrics represents cost performance metrics
type CostMetrics struct {
	TotalCost             float64            `json:"total_cost"`
	AverageCostPerRequest float64            `json:"average_cost_per_request"`
	TotalRequests         int                `json:"total_requests"`
	CostByMethod          map[string]float64 `json:"cost_by_method"`
}

// AlertSummary represents alert performance summary
type AlertSummary struct {
	TotalAlerts    int `json:"total_alerts"`
	ErrorAlerts    int `json:"error_alerts"`
	WarningAlerts  int `json:"warning_alerts"`
	InfoAlerts     int `json:"info_alerts"`
	ResolvedAlerts int `json:"resolved_alerts"`
}
