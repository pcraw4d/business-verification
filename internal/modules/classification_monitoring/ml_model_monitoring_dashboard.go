package classification_monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MLModelMonitoringDashboard provides a comprehensive dashboard for ML model monitoring
type MLModelMonitoringDashboard struct {
	config *MLModelDashboardConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Monitoring components
	bertMonitor              *BERTModelMonitor
	ensembleWeightMonitor    *EnsembleWeightMonitor
	uncertaintyMonitor       *UncertaintyQuantificationMonitor
	securityProcessingMonitor *SecurityProcessingTimeMonitor
	accuracyTracker          *AdvancedAccuracyTracker

	// Dashboard state
	lastUpdateTime time.Time
	healthStatus   string
	alertsSummary  *AlertsSummary
}

// MLModelDashboardConfig holds configuration for the ML model monitoring dashboard
type MLModelDashboardConfig struct {
	// Dashboard settings
	DashboardEnabled        bool          `json:"dashboard_enabled"`
	UpdateInterval          time.Duration `json:"update_interval"`
	HealthCheckInterval     time.Duration `json:"health_check_interval"`
	AlertSummaryInterval    time.Duration `json:"alert_summary_interval"`

	// Display settings
	MaxAlertsDisplayed      int           `json:"max_alerts_displayed"`
	MaxMetricsHistory       int           `json:"max_metrics_history"`
	ShowDetailedMetrics     bool          `json:"show_detailed_metrics"`
	ShowTrendAnalysis       bool          `json:"show_trend_analysis"`

	// Integration settings
	IntegrateBERTMonitoring      bool `json:"integrate_bert_monitoring"`
	IntegrateEnsembleMonitoring  bool `json:"integrate_ensemble_monitoring"`
	IntegrateUncertaintyMonitoring bool `json:"integrate_uncertainty_monitoring"`
	IntegrateSecurityMonitoring  bool `json:"integrate_security_monitoring"`
	IntegrateAccuracyTracking    bool `json:"integrate_accuracy_tracking"`
}

// AlertsSummary represents a summary of all alerts across monitoring components
type AlertsSummary struct {
	Timestamp           time.Time `json:"timestamp"`
	TotalAlerts         int       `json:"total_alerts"`
	CriticalAlerts      int       `json:"critical_alerts"`
	WarningAlerts       int       `json:"warning_alerts"`
	UnacknowledgedAlerts int      `json:"unacknowledged_alerts"`
	ResolvedAlerts      int       `json:"resolved_alerts"`
	AlertBreakdown      map[string]int `json:"alert_breakdown"`
}

// MLModelDashboardMetrics represents comprehensive metrics for the dashboard
type MLModelDashboardMetrics struct {
	Timestamp           time.Time                           `json:"timestamp"`
	OverallHealth       string                              `json:"overall_health"`
	HealthScore         float64                             `json:"health_score"`
	
	// BERT Model Metrics
	BERTMetrics         map[string]*BERTModelMetrics        `json:"bert_metrics,omitempty"`
	BERTHealth          string                              `json:"bert_health,omitempty"`
	
	// Ensemble Weight Metrics
	EnsembleWeights     map[string]float64                  `json:"ensemble_weights,omitempty"`
	WeightDistributionHealth string                         `json:"weight_distribution_health,omitempty"`
	
	// Uncertainty Quantification Metrics
	UncertaintyMetrics  *UncertaintyQuantificationMetrics   `json:"uncertainty_metrics,omitempty"`
	UncertaintyHealth   string                              `json:"uncertainty_health,omitempty"`
	
	// Security Processing Metrics
	SecurityMetrics     *SecurityProcessingMetrics          `json:"security_metrics,omitempty"`
	SecurityHealth      string                              `json:"security_health,omitempty"`
	
	// Accuracy Tracking Metrics
	AccuracyMetrics     *RealTimeMetrics                    `json:"accuracy_metrics,omitempty"`
	AccuracyHealth      string                              `json:"accuracy_health,omitempty"`
	
	// Alerts Summary
	AlertsSummary       *AlertsSummary                      `json:"alerts_summary"`
	
	// Performance Summary
	PerformanceSummary  *PerformanceSummary                 `json:"performance_summary"`
	
	// Recommendations
	Recommendations     []string                            `json:"recommendations"`
}

// PerformanceSummary represents overall performance summary
type PerformanceSummary struct {
	OverallAccuracy     float64   `json:"overall_accuracy"`
	AverageLatency      time.Duration `json:"average_latency"`
	Throughput          float64   `json:"throughput"`
	ErrorRate           float64   `json:"error_rate"`
	Uptime              float64   `json:"uptime"`
	LastUpdated         time.Time `json:"last_updated"`
}

// NewMLModelMonitoringDashboard creates a new ML model monitoring dashboard
func NewMLModelMonitoringDashboard(
	config *MLModelDashboardConfig,
	bertMonitor *BERTModelMonitor,
	ensembleWeightMonitor *EnsembleWeightMonitor,
	uncertaintyMonitor *UncertaintyQuantificationMonitor,
	securityProcessingMonitor *SecurityProcessingTimeMonitor,
	accuracyTracker *AdvancedAccuracyTracker,
	logger *zap.Logger,
) *MLModelMonitoringDashboard {
	if config == nil {
		config = DefaultMLModelDashboardConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &MLModelMonitoringDashboard{
		config:                   config,
		logger:                   logger,
		bertMonitor:              bertMonitor,
		ensembleWeightMonitor:    ensembleWeightMonitor,
		uncertaintyMonitor:       uncertaintyMonitor,
		securityProcessingMonitor: securityProcessingMonitor,
		accuracyTracker:          accuracyTracker,
		lastUpdateTime:           time.Now(),
		healthStatus:             "unknown",
		alertsSummary:            &AlertsSummary{},
	}
}

// DefaultMLModelDashboardConfig returns default configuration
func DefaultMLModelDashboardConfig() *MLModelDashboardConfig {
	return &MLModelDashboardConfig{
		DashboardEnabled:              true,
		UpdateInterval:                30 * time.Second,
		HealthCheckInterval:           1 * time.Minute,
		AlertSummaryInterval:          5 * time.Minute,
		MaxAlertsDisplayed:            50,
		MaxMetricsHistory:             1000,
		ShowDetailedMetrics:           true,
		ShowTrendAnalysis:             true,
		IntegrateBERTMonitoring:       true,
		IntegrateEnsembleMonitoring:   true,
		IntegrateUncertaintyMonitoring: true,
		IntegrateSecurityMonitoring:   true,
		IntegrateAccuracyTracking:     true,
	}
}

// GetDashboardMetrics returns comprehensive dashboard metrics
func (dashboard *MLModelMonitoringDashboard) GetDashboardMetrics(ctx context.Context) (*MLModelDashboardMetrics, error) {
	dashboard.mu.Lock()
	defer dashboard.mu.Unlock()

	metrics := &MLModelDashboardMetrics{
		Timestamp:      time.Now(),
		OverallHealth:  "unknown",
		HealthScore:    0.0,
		AlertsSummary:  &AlertsSummary{},
		PerformanceSummary: &PerformanceSummary{},
		Recommendations: make([]string, 0),
	}

	// Collect BERT model metrics
	if dashboard.config.IntegrateBERTMonitoring && dashboard.bertMonitor != nil {
		bertMetrics := dashboard.bertMonitor.GetAllBERTModelMetrics()
		metrics.BERTMetrics = bertMetrics
		metrics.BERTHealth = dashboard.assessBERTHealth(bertMetrics)
	}

	// Collect ensemble weight metrics
	if dashboard.config.IntegrateEnsembleMonitoring && dashboard.ensembleWeightMonitor != nil {
		weightMetrics := dashboard.ensembleWeightMonitor.GetWeightDistributionMetrics()
		metrics.EnsembleWeights = weightMetrics.CurrentWeights
		metrics.WeightDistributionHealth = weightMetrics.DistributionHealth
	}

	// Collect uncertainty quantification metrics
	if dashboard.config.IntegrateUncertaintyMonitoring && dashboard.uncertaintyMonitor != nil {
		uncertaintyMetrics := dashboard.uncertaintyMonitor.GetUncertaintyMetrics()
		metrics.UncertaintyMetrics = uncertaintyMetrics
		metrics.UncertaintyHealth = uncertaintyMetrics.HealthStatus
	}

	// Collect security processing metrics
	if dashboard.config.IntegrateSecurityMonitoring && dashboard.securityProcessingMonitor != nil {
		securityMetrics := dashboard.securityProcessingMonitor.GetSecurityProcessingMetrics()
		metrics.SecurityMetrics = securityMetrics
		metrics.SecurityHealth = securityMetrics.OverallHealth
	}

	// Collect accuracy tracking metrics
	if dashboard.config.IntegrateAccuracyTracking && dashboard.accuracyTracker != nil {
		accuracyMetrics := dashboard.accuracyTracker.GetRealTimeMetrics()
		metrics.AccuracyMetrics = accuracyMetrics
		metrics.AccuracyHealth = dashboard.assessAccuracyHealth(accuracyMetrics)
	}

	// Collect alerts summary
	metrics.AlertsSummary = dashboard.collectAlertsSummary()

	// Collect performance summary
	metrics.PerformanceSummary = dashboard.collectPerformanceSummary()

	// Determine overall health
	metrics.OverallHealth = dashboard.determineOverallHealth(metrics)
	metrics.HealthScore = dashboard.calculateHealthScore(metrics)

	// Generate recommendations
	metrics.Recommendations = dashboard.generateRecommendations(metrics)

	dashboard.lastUpdateTime = time.Now()
	dashboard.healthStatus = metrics.OverallHealth

	return metrics, nil
}

// assessBERTHealth assesses BERT model health
func (dashboard *MLModelMonitoringDashboard) assessBERTHealth(bertMetrics map[string]*BERTModelMetrics) string {
	if len(bertMetrics) == 0 {
		return "unknown"
	}

	criticalCount := 0
	warningCount := 0

	for _, metrics := range bertMetrics {
		if metrics.DriftStatus == "critical" {
			criticalCount++
		} else if metrics.DriftStatus == "warning" {
			warningCount++
		}
	}

	if criticalCount > 0 {
		return "critical"
	} else if warningCount > 0 {
		return "warning"
	}

	return "healthy"
}

// assessAccuracyHealth assesses accuracy tracking health
func (dashboard *MLModelMonitoringDashboard) assessAccuracyHealth(accuracyMetrics *RealTimeMetrics) string {
	if accuracyMetrics == nil {
		return "unknown"
	}

	// Check overall accuracy
	if accuracyMetrics.CurrentAccuracy < 0.8 { // 80% accuracy threshold
		return "warning"
	}

	// Check if accuracy is improving
	if accuracyMetrics.CurrentAccuracy < 0.7 { // 70% accuracy threshold
		return "critical"
	}

	return "healthy"
}

// collectAlertsSummary collects alerts from all monitoring components
func (dashboard *MLModelMonitoringDashboard) collectAlertsSummary() *AlertsSummary {
	summary := &AlertsSummary{
		Timestamp:           time.Now(),
		TotalAlerts:         0,
		CriticalAlerts:      0,
		WarningAlerts:       0,
		UnacknowledgedAlerts: 0,
		ResolvedAlerts:      0,
		AlertBreakdown:      make(map[string]int),
	}

	// Collect BERT model alerts
	if dashboard.config.IntegrateBERTMonitoring && dashboard.bertMonitor != nil {
		bertAlerts := dashboard.bertMonitor.GetDriftAlerts()
		for _, alert := range bertAlerts {
			summary.TotalAlerts++
			if alert.Severity == "critical" {
				summary.CriticalAlerts++
			} else if alert.Severity == "warning" {
				summary.WarningAlerts++
			}
			if !alert.Acknowledged {
				summary.UnacknowledgedAlerts++
			}
			if alert.Resolved {
				summary.ResolvedAlerts++
			}
			summary.AlertBreakdown["bert_drift"]++
		}
	}

	// Collect ensemble weight alerts
	if dashboard.config.IntegrateEnsembleMonitoring && dashboard.ensembleWeightMonitor != nil {
		weightAlerts := dashboard.ensembleWeightMonitor.GetWeightAlerts()
		for _, alert := range weightAlerts {
			summary.TotalAlerts++
			if alert.Severity == "critical" {
				summary.CriticalAlerts++
			} else if alert.Severity == "warning" {
				summary.WarningAlerts++
			}
			if !alert.Acknowledged {
				summary.UnacknowledgedAlerts++
			}
			if alert.Resolved {
				summary.ResolvedAlerts++
			}
			summary.AlertBreakdown["weight_distribution"]++
		}
	}

	// Collect uncertainty alerts
	if dashboard.config.IntegrateUncertaintyMonitoring && dashboard.uncertaintyMonitor != nil {
		uncertaintyAlerts := dashboard.uncertaintyMonitor.GetUncertaintyAlerts()
		for _, alert := range uncertaintyAlerts {
			summary.TotalAlerts++
			if alert.Severity == "critical" {
				summary.CriticalAlerts++
			} else if alert.Severity == "warning" {
				summary.WarningAlerts++
			}
			if !alert.Acknowledged {
				summary.UnacknowledgedAlerts++
			}
			if alert.Resolved {
				summary.ResolvedAlerts++
			}
			summary.AlertBreakdown["uncertainty_quantification"]++
		}
	}

	// Collect security processing alerts
	if dashboard.config.IntegrateSecurityMonitoring && dashboard.securityProcessingMonitor != nil {
		securityAlerts := dashboard.securityProcessingMonitor.GetSecurityProcessingAlerts()
		for _, alert := range securityAlerts {
			summary.TotalAlerts++
			if alert.Severity == "critical" {
				summary.CriticalAlerts++
			} else if alert.Severity == "warning" {
				summary.WarningAlerts++
			}
			if !alert.Acknowledged {
				summary.UnacknowledgedAlerts++
			}
			if alert.Resolved {
				summary.ResolvedAlerts++
			}
			summary.AlertBreakdown["security_processing"]++
		}
	}

	return summary
}

// collectPerformanceSummary collects overall performance summary
func (dashboard *MLModelMonitoringDashboard) collectPerformanceSummary() *PerformanceSummary {
	summary := &PerformanceSummary{
		LastUpdated: time.Now(),
	}

	// Collect accuracy metrics
	if dashboard.config.IntegrateAccuracyTracking && dashboard.accuracyTracker != nil {
		summary.OverallAccuracy = dashboard.accuracyTracker.GetOverallAccuracy()
	}

	// Collect other performance metrics from various components
	// This would be implemented based on specific performance tracking needs

	return summary
}

// determineOverallHealth determines overall system health
func (dashboard *MLModelMonitoringDashboard) determineOverallHealth(metrics *MLModelDashboardMetrics) string {
	// Check for critical health status in any component
	if metrics.BERTHealth == "critical" || 
	   metrics.WeightDistributionHealth == "critical" ||
	   metrics.UncertaintyHealth == "critical" ||
	   metrics.SecurityHealth == "critical" ||
	   metrics.AccuracyHealth == "critical" {
		return "critical"
	}

	// Check for warning health status in any component
	if metrics.BERTHealth == "warning" || 
	   metrics.WeightDistributionHealth == "warning" ||
	   metrics.UncertaintyHealth == "warning" ||
	   metrics.SecurityHealth == "warning" ||
	   metrics.AccuracyHealth == "warning" {
		return "warning"
	}

	// Check for critical alerts
	if metrics.AlertsSummary.CriticalAlerts > 0 {
		return "critical"
	}

	// Check for warning alerts
	if metrics.AlertsSummary.WarningAlerts > 0 {
		return "warning"
	}

	return "healthy"
}

// calculateHealthScore calculates overall health score (0-100)
func (dashboard *MLModelMonitoringDashboard) calculateHealthScore(metrics *MLModelDashboardMetrics) float64 {
	score := 100.0

	// Deduct points for critical health status
	if metrics.BERTHealth == "critical" {
		score -= 20
	} else if metrics.BERTHealth == "warning" {
		score -= 10
	}

	if metrics.WeightDistributionHealth == "critical" {
		score -= 20
	} else if metrics.WeightDistributionHealth == "warning" {
		score -= 10
	}

	if metrics.UncertaintyHealth == "critical" {
		score -= 20
	} else if metrics.UncertaintyHealth == "warning" {
		score -= 10
	}

	if metrics.SecurityHealth == "critical" {
		score -= 20
	} else if metrics.SecurityHealth == "warning" {
		score -= 10
	}

	if metrics.AccuracyHealth == "critical" {
		score -= 20
	} else if metrics.AccuracyHealth == "warning" {
		score -= 10
	}

	// Deduct points for alerts
	score -= float64(metrics.AlertsSummary.CriticalAlerts) * 5
	score -= float64(metrics.AlertsSummary.WarningAlerts) * 2

	// Ensure score is not negative
	if score < 0 {
		score = 0
	}

	return score
}

// generateRecommendations generates recommendations based on current metrics
func (dashboard *MLModelMonitoringDashboard) generateRecommendations(metrics *MLModelDashboardMetrics) []string {
	recommendations := make([]string, 0)

	// BERT model recommendations
	if metrics.BERTHealth == "critical" {
		recommendations = append(recommendations, "BERT model health is critical. Investigate model drift and consider retraining.")
	} else if metrics.BERTHealth == "warning" {
		recommendations = append(recommendations, "BERT model health is warning. Monitor model performance closely.")
	}

	// Ensemble weight recommendations
	if metrics.WeightDistributionHealth == "critical" {
		recommendations = append(recommendations, "Weight distribution is critical. Check for method imbalances and adjust weights.")
	} else if metrics.WeightDistributionHealth == "warning" {
		recommendations = append(recommendations, "Weight distribution is warning. Monitor weight stability.")
	}

	// Uncertainty quantification recommendations
	if metrics.UncertaintyHealth == "critical" {
		recommendations = append(recommendations, "Uncertainty quantification is critical. Check calibration and reliability.")
	} else if metrics.UncertaintyHealth == "warning" {
		recommendations = append(recommendations, "Uncertainty quantification is warning. Monitor calibration accuracy.")
	}

	// Security processing recommendations
	if metrics.SecurityHealth == "critical" {
		recommendations = append(recommendations, "Security processing is critical. Check processing times and error rates.")
	} else if metrics.SecurityHealth == "warning" {
		recommendations = append(recommendations, "Security processing is warning. Monitor processing performance.")
	}

	// Accuracy recommendations
	if metrics.AccuracyHealth == "critical" {
		recommendations = append(recommendations, "Accuracy is critical. Investigate classification performance.")
	} else if metrics.AccuracyHealth == "warning" {
		recommendations = append(recommendations, "Accuracy is warning. Monitor classification trends.")
	}

	// Alert recommendations
	if metrics.AlertsSummary.CriticalAlerts > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Address %d critical alerts immediately.", metrics.AlertsSummary.CriticalAlerts))
	}

	if metrics.AlertsSummary.WarningAlerts > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d warning alerts.", metrics.AlertsSummary.WarningAlerts))
	}

	if metrics.AlertsSummary.UnacknowledgedAlerts > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Acknowledge %d unacknowledged alerts.", metrics.AlertsSummary.UnacknowledgedAlerts))
	}

	// Performance recommendations
	if metrics.PerformanceSummary.OverallAccuracy < 0.8 {
		recommendations = append(recommendations, "Overall accuracy is below 80%. Consider model improvements.")
	}

	if metrics.PerformanceSummary.ErrorRate > 0.1 {
		recommendations = append(recommendations, "Error rate is above 10%. Investigate error causes.")
	}

	return recommendations
}

// GetHealthStatus returns current health status
func (dashboard *MLModelMonitoringDashboard) GetHealthStatus() string {
	dashboard.mu.RLock()
	defer dashboard.mu.RUnlock()

	return dashboard.healthStatus
}

// GetLastUpdateTime returns last update time
func (dashboard *MLModelMonitoringDashboard) GetLastUpdateTime() time.Time {
	dashboard.mu.RLock()
	defer dashboard.mu.RUnlock()

	return dashboard.lastUpdateTime
}

// GetAlertsSummary returns current alerts summary
func (dashboard *MLModelMonitoringDashboard) GetAlertsSummary() *AlertsSummary {
	dashboard.mu.RLock()
	defer dashboard.mu.RUnlock()

	return dashboard.alertsSummary
}

// AcknowledgeAllAlerts acknowledges all unacknowledged alerts
func (dashboard *MLModelMonitoringDashboard) AcknowledgeAllAlerts(ctx context.Context) error {
	dashboard.mu.Lock()
	defer dashboard.mu.Unlock()

	// Acknowledge BERT model alerts
	if dashboard.config.IntegrateBERTMonitoring && dashboard.bertMonitor != nil {
		bertAlerts := dashboard.bertMonitor.GetDriftAlerts()
		for _, alert := range bertAlerts {
			if !alert.Acknowledged {
				dashboard.bertMonitor.AcknowledgeAlert(alert.ID)
			}
		}
	}

	// Acknowledge ensemble weight alerts
	if dashboard.config.IntegrateEnsembleMonitoring && dashboard.ensembleWeightMonitor != nil {
		weightAlerts := dashboard.ensembleWeightMonitor.GetWeightAlerts()
		for _, alert := range weightAlerts {
			if !alert.Acknowledged {
				dashboard.ensembleWeightMonitor.AcknowledgeWeightAlert(alert.ID)
			}
		}
	}

	// Acknowledge uncertainty alerts
	if dashboard.config.IntegrateUncertaintyMonitoring && dashboard.uncertaintyMonitor != nil {
		uncertaintyAlerts := dashboard.uncertaintyMonitor.GetUncertaintyAlerts()
		for _, alert := range uncertaintyAlerts {
			if !alert.Acknowledged {
				dashboard.uncertaintyMonitor.AcknowledgeUncertaintyAlert(alert.ID)
			}
		}
	}

	// Acknowledge security processing alerts
	if dashboard.config.IntegrateSecurityMonitoring && dashboard.securityProcessingMonitor != nil {
		securityAlerts := dashboard.securityProcessingMonitor.GetSecurityProcessingAlerts()
		for _, alert := range securityAlerts {
			if !alert.Acknowledged {
				dashboard.securityProcessingMonitor.AcknowledgeSecurityProcessingAlert(alert.ID)
			}
		}
	}

	dashboard.logger.Info("All alerts acknowledged")
	return nil
}

// GetDetailedReport returns a detailed monitoring report
func (dashboard *MLModelMonitoringDashboard) GetDetailedReport(ctx context.Context) (*MLModelDetailedReport, error) {
	metrics, err := dashboard.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}

	report := &MLModelDetailedReport{
		Timestamp:           time.Now(),
		DashboardMetrics:    metrics,
		ComponentReports:    make(map[string]interface{}),
		HealthAnalysis:      &HealthAnalysis{},
		PerformanceAnalysis: &MLPerformanceAnalysis{},
		Recommendations:     metrics.Recommendations,
	}

	// Generate component-specific reports
	if dashboard.config.IntegrateBERTMonitoring && dashboard.bertMonitor != nil {
		report.ComponentReports["bert_models"] = dashboard.bertMonitor.GetAllBERTModelMetrics()
	}

	if dashboard.config.IntegrateEnsembleMonitoring && dashboard.ensembleWeightMonitor != nil {
		report.ComponentReports["ensemble_weights"] = dashboard.ensembleWeightMonitor.GetWeightDistributionMetrics()
	}

	if dashboard.config.IntegrateUncertaintyMonitoring && dashboard.uncertaintyMonitor != nil {
		report.ComponentReports["uncertainty_quantification"] = dashboard.uncertaintyMonitor.GetUncertaintyMetrics()
	}

	if dashboard.config.IntegrateSecurityMonitoring && dashboard.securityProcessingMonitor != nil {
		report.ComponentReports["security_processing"] = dashboard.securityProcessingMonitor.GetSecurityProcessingMetrics()
	}

	// Generate health analysis
	report.HealthAnalysis = dashboard.generateHealthAnalysis(metrics)

	// Generate performance analysis
	report.PerformanceAnalysis = dashboard.generatePerformanceAnalysis(metrics)

	return report, nil
}

// MLModelDetailedReport represents a detailed monitoring report
type MLModelDetailedReport struct {
	Timestamp           time.Time                 `json:"timestamp"`
	DashboardMetrics    *MLModelDashboardMetrics  `json:"dashboard_metrics"`
	ComponentReports    map[string]interface{}    `json:"component_reports"`
	HealthAnalysis      *HealthAnalysis           `json:"health_analysis"`
	PerformanceAnalysis *MLPerformanceAnalysis    `json:"performance_analysis"`
	Recommendations     []string                  `json:"recommendations"`
}

// HealthAnalysis represents health analysis results
type HealthAnalysis struct {
	OverallHealth       string   `json:"overall_health"`
	HealthScore         float64  `json:"health_score"`
	CriticalIssues      []string `json:"critical_issues"`
	WarningIssues       []string `json:"warning_issues"`
	HealthyComponents   []string `json:"healthy_components"`
}

// MLPerformanceAnalysis represents performance analysis results
type MLPerformanceAnalysis struct {
	OverallPerformance  string   `json:"overall_performance"`
	PerformanceScore    float64  `json:"performance_score"`
	Bottlenecks         []string `json:"bottlenecks"`
	OptimizationOpportunities []string `json:"optimization_opportunities"`
}

// generateHealthAnalysis generates health analysis
func (dashboard *MLModelMonitoringDashboard) generateHealthAnalysis(metrics *MLModelDashboardMetrics) *HealthAnalysis {
	analysis := &HealthAnalysis{
		OverallHealth:     metrics.OverallHealth,
		HealthScore:       metrics.HealthScore,
		CriticalIssues:    make([]string, 0),
		WarningIssues:     make([]string, 0),
		HealthyComponents: make([]string, 0),
	}

	// Analyze each component
	if metrics.BERTHealth == "critical" {
		analysis.CriticalIssues = append(analysis.CriticalIssues, "BERT model health is critical")
	} else if metrics.BERTHealth == "warning" {
		analysis.WarningIssues = append(analysis.WarningIssues, "BERT model health is warning")
	} else if metrics.BERTHealth == "healthy" {
		analysis.HealthyComponents = append(analysis.HealthyComponents, "BERT models")
	}

	if metrics.WeightDistributionHealth == "critical" {
		analysis.CriticalIssues = append(analysis.CriticalIssues, "Weight distribution is critical")
	} else if metrics.WeightDistributionHealth == "warning" {
		analysis.WarningIssues = append(analysis.WarningIssues, "Weight distribution is warning")
	} else if metrics.WeightDistributionHealth == "healthy" {
		analysis.HealthyComponents = append(analysis.HealthyComponents, "Ensemble weights")
	}

	if metrics.UncertaintyHealth == "critical" {
		analysis.CriticalIssues = append(analysis.CriticalIssues, "Uncertainty quantification is critical")
	} else if metrics.UncertaintyHealth == "warning" {
		analysis.WarningIssues = append(analysis.WarningIssues, "Uncertainty quantification is warning")
	} else if metrics.UncertaintyHealth == "healthy" {
		analysis.HealthyComponents = append(analysis.HealthyComponents, "Uncertainty quantification")
	}

	if metrics.SecurityHealth == "critical" {
		analysis.CriticalIssues = append(analysis.CriticalIssues, "Security processing is critical")
	} else if metrics.SecurityHealth == "warning" {
		analysis.WarningIssues = append(analysis.WarningIssues, "Security processing is warning")
	} else if metrics.SecurityHealth == "healthy" {
		analysis.HealthyComponents = append(analysis.HealthyComponents, "Security processing")
	}

	if metrics.AccuracyHealth == "critical" {
		analysis.CriticalIssues = append(analysis.CriticalIssues, "Accuracy tracking is critical")
	} else if metrics.AccuracyHealth == "warning" {
		analysis.WarningIssues = append(analysis.WarningIssues, "Accuracy tracking is warning")
	} else if metrics.AccuracyHealth == "healthy" {
		analysis.HealthyComponents = append(analysis.HealthyComponents, "Accuracy tracking")
	}

	return analysis
}

// generatePerformanceAnalysis generates performance analysis
func (dashboard *MLModelMonitoringDashboard) generatePerformanceAnalysis(metrics *MLModelDashboardMetrics) *MLPerformanceAnalysis {
	analysis := &MLPerformanceAnalysis{
		OverallPerformance:        "unknown",
		PerformanceScore:         0.0,
		Bottlenecks:              make([]string, 0),
		OptimizationOpportunities: make([]string, 0),
	}

	// Calculate performance score based on accuracy and other metrics
	score := 0.0
	if metrics.PerformanceSummary.OverallAccuracy > 0 {
		score += metrics.PerformanceSummary.OverallAccuracy * 40 // 40% weight for accuracy
	}

	// Add other performance factors
	if metrics.PerformanceSummary.ErrorRate < 0.1 {
		score += 20 // 20% weight for low error rate
	}

	if metrics.PerformanceSummary.AverageLatency < 500*time.Millisecond {
		score += 20 // 20% weight for low latency
	}

	if metrics.PerformanceSummary.Throughput > 10 {
		score += 20 // 20% weight for high throughput
	}

	analysis.PerformanceScore = score

	// Determine overall performance
	if score >= 80 {
		analysis.OverallPerformance = "excellent"
	} else if score >= 60 {
		analysis.OverallPerformance = "good"
	} else if score >= 40 {
		analysis.OverallPerformance = "fair"
	} else {
		analysis.OverallPerformance = "poor"
	}

	// Identify bottlenecks
	if metrics.PerformanceSummary.AverageLatency > 1*time.Second {
		analysis.Bottlenecks = append(analysis.Bottlenecks, "High latency detected")
	}

	if metrics.PerformanceSummary.ErrorRate > 0.2 {
		analysis.Bottlenecks = append(analysis.Bottlenecks, "High error rate detected")
	}

	if metrics.PerformanceSummary.Throughput < 5 {
		analysis.Bottlenecks = append(analysis.Bottlenecks, "Low throughput detected")
	}

	// Identify optimization opportunities
	if metrics.PerformanceSummary.OverallAccuracy < 0.9 {
		analysis.OptimizationOpportunities = append(analysis.OptimizationOpportunities, "Improve model accuracy")
	}

	if metrics.PerformanceSummary.AverageLatency > 200*time.Millisecond {
		analysis.OptimizationOpportunities = append(analysis.OptimizationOpportunities, "Optimize processing latency")
	}

	return analysis
}
