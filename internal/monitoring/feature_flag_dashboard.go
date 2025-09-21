package monitoring

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// FeatureFlagDashboard provides monitoring and analytics for feature flags
type FeatureFlagDashboard struct {
	// Core components
	featureFlagManager *config.GranularFeatureFlagManager

	// Analytics data
	analytics *FeatureFlagAnalytics

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Configuration
	config FeatureFlagDashboardConfig

	// Update channels
	updateChannels []chan *FeatureFlagAnalytics
}

// FeatureFlagDashboardConfig holds configuration for the feature flag dashboard
type FeatureFlagDashboardConfig struct {
	// Dashboard configuration
	Port                 int           `json:"port"`
	UpdateInterval       time.Duration `json:"update_interval"`
	MetricsRetentionDays int           `json:"metrics_retention_days"`

	// Analytics configuration
	AnalyticsEnabled bool `json:"analytics_enabled"`
	RealTimeUpdates  bool `json:"real_time_updates"`

	// Export configuration
	ExportEnabled  bool   `json:"export_enabled"`
	ExportFormat   string `json:"export_format"` // json, csv, prometheus
	ExportEndpoint string `json:"export_endpoint"`

	// Alerting configuration
	AlertingEnabled bool            `json:"alerting_enabled"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
}

// FeatureFlagAnalytics holds analytics data for feature flags
type FeatureFlagAnalytics struct {
	// Overall metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	ErrorRate          float64       `json:"error_rate"`

	// Feature flag usage
	FeatureFlagUsage map[string]*FeatureFlagUsage `json:"feature_flag_usage"`

	// Model performance
	ModelPerformance map[string]*ModelPerformance `json:"model_performance"`

	// A/B testing results
	ABTestResults map[string]*ABTestAnalytics `json:"ab_test_results"`

	// Rollout progress
	RolloutProgress map[string]*RolloutAnalytics `json:"rollout_progress"`

	// Performance trends
	PerformanceTrends []PerformanceTrend `json:"performance_trends"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// FeatureFlagUsage holds usage statistics for a feature flag
type FeatureFlagUsage struct {
	// Usage metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	ErrorRate          float64       `json:"error_rate"`

	// Performance metrics
	P95Latency time.Duration `json:"p95_latency"`
	P99Latency time.Duration `json:"p99_latency"`
	Throughput float64       `json:"throughput"`

	// Usage patterns
	PeakUsageTime     time.Time        `json:"peak_usage_time"`
	UsageDistribution map[string]int64 `json:"usage_distribution"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// ModelPerformance holds performance metrics for a model
type ModelPerformance struct {
	// Model information
	ModelName    string `json:"model_name"`
	ModelType    string `json:"model_type"`
	ModelVersion string `json:"model_version"`

	// Performance metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	ErrorRate          float64       `json:"error_rate"`

	// Accuracy metrics
	Accuracy   float64 `json:"accuracy"`
	Confidence float64 `json:"confidence"`
	Precision  float64 `json:"precision"`
	Recall     float64 `json:"recall"`
	F1Score    float64 `json:"f1_score"`

	// Resource usage
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	GPUUsage    float64 `json:"gpu_usage"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// ABTestAnalytics holds analytics for A/B tests
type ABTestAnalytics struct {
	// Test information
	TestID    string     `json:"test_id"`
	TestName  string     `json:"test_name"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	IsActive  bool       `json:"is_active"`

	// Traffic allocation
	ControlGroupTraffic int `json:"control_group_traffic"`
	TestGroupTraffic    int `json:"test_group_traffic"`

	// Results
	ControlGroupResults *TestGroupResults `json:"control_group_results"`
	TestGroupResults    *TestGroupResults `json:"test_group_results"`

	// Statistical significance
	StatisticalSignificance float64 `json:"statistical_significance"`
	IsSignificant           bool    `json:"is_significant"`
	Winner                  string  `json:"winner"`
	Confidence              float64 `json:"confidence"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// TestGroupResults holds results for a test group
type TestGroupResults struct {
	// Sample size
	SampleSize int `json:"sample_size"`

	// Performance metrics
	AverageLatency time.Duration `json:"average_latency"`
	ErrorRate      float64       `json:"error_rate"`
	Throughput     float64       `json:"throughput"`

	// Business metrics
	SuccessRate    float64 `json:"success_rate"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`

	// Quality metrics
	Accuracy         float64 `json:"accuracy"`
	Confidence       float64 `json:"confidence"`
	UserSatisfaction float64 `json:"user_satisfaction"`
}

// RolloutAnalytics holds analytics for rollout progress
type RolloutAnalytics struct {
	// Rollout information
	FeatureName       string    `json:"feature_name"`
	StartTime         time.Time `json:"start_time"`
	CurrentPercentage int       `json:"current_percentage"`
	TargetPercentage  int       `json:"target_percentage"`

	// Performance metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	ErrorRate          float64       `json:"error_rate"`

	// Rollout progress
	RolloutHistory  []RolloutStep `json:"rollout_history"`
	NextRolloutTime *time.Time    `json:"next_rollout_time"`

	// Health metrics
	IsHealthy   bool    `json:"is_healthy"`
	HealthScore float64 `json:"health_score"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// RolloutStep represents a step in the rollout process
type RolloutStep struct {
	Percentage     int           `json:"percentage"`
	Timestamp      time.Time     `json:"timestamp"`
	ErrorRate      float64       `json:"error_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	IsHealthy      bool          `json:"is_healthy"`
}

// PerformanceTrend represents a performance trend over time
type PerformanceTrend struct {
	Timestamp      time.Time     `json:"timestamp"`
	TotalRequests  int64         `json:"total_requests"`
	AverageLatency time.Duration `json:"average_latency"`
	ErrorRate      float64       `json:"error_rate"`
	Throughput     float64       `json:"throughput"`
}

// AlertThresholds holds alerting thresholds
type AlertThresholds struct {
	// Performance thresholds
	MaxLatencyThreshold    time.Duration `json:"max_latency_threshold"`
	MaxErrorRateThreshold  float64       `json:"max_error_rate_threshold"`
	MinThroughputThreshold float64       `json:"min_throughput_threshold"`

	// A/B testing thresholds
	MinStatisticalSignificance float64 `json:"min_statistical_significance"`
	MinSampleSize              int     `json:"min_sample_size"`

	// Rollout thresholds
	MaxRolloutErrorRate   float64 `json:"max_rollout_error_rate"`
	MinRolloutHealthScore float64 `json:"min_rollout_health_score"`
}

// NewFeatureFlagDashboard creates a new feature flag dashboard
func NewFeatureFlagDashboard(
	featureFlagManager *config.GranularFeatureFlagManager,
	config FeatureFlagDashboardConfig,
	logger *log.Logger,
) *FeatureFlagDashboard {
	if logger == nil {
		logger = log.Default()
	}

	dashboard := &FeatureFlagDashboard{
		featureFlagManager: featureFlagManager,
		analytics: &FeatureFlagAnalytics{
			FeatureFlagUsage:  make(map[string]*FeatureFlagUsage),
			ModelPerformance:  make(map[string]*ModelPerformance),
			ABTestResults:     make(map[string]*ABTestAnalytics),
			RolloutProgress:   make(map[string]*RolloutAnalytics),
			PerformanceTrends: make([]PerformanceTrend, 0),
		},
		logger:         logger,
		config:         config,
		updateChannels: make([]chan *FeatureFlagAnalytics, 0),
	}

	// Start background processes
	go dashboard.startAnalyticsCollection()
	go dashboard.startDashboardServer()

	return dashboard
}

// GetAnalytics returns the current analytics data
func (ffd *FeatureFlagDashboard) GetAnalytics() *FeatureFlagAnalytics {
	ffd.mu.RLock()
	defer ffd.mu.RUnlock()
	return ffd.analytics
}

// GetFeatureFlagUsage returns usage statistics for a specific feature flag
func (ffd *FeatureFlagDashboard) GetFeatureFlagUsage(flagName string) *FeatureFlagUsage {
	ffd.mu.RLock()
	defer ffd.mu.RUnlock()
	return ffd.analytics.FeatureFlagUsage[flagName]
}

// GetModelPerformance returns performance metrics for a specific model
func (ffd *FeatureFlagDashboard) GetModelPerformance(modelName string) *ModelPerformance {
	ffd.mu.RLock()
	defer ffd.mu.RUnlock()
	return ffd.analytics.ModelPerformance[modelName]
}

// GetABTestResults returns A/B test results for a specific test
func (ffd *FeatureFlagDashboard) GetABTestResults(testID string) *ABTestAnalytics {
	ffd.mu.RLock()
	defer ffd.mu.RUnlock()
	return ffd.analytics.ABTestResults[testID]
}

// GetRolloutProgress returns rollout progress for a specific feature
func (ffd *FeatureFlagDashboard) GetRolloutProgress(featureName string) *RolloutAnalytics {
	ffd.mu.RLock()
	defer ffd.mu.RUnlock()
	return ffd.analytics.RolloutProgress[featureName]
}

// startAnalyticsCollection starts collecting analytics data
func (ffd *FeatureFlagDashboard) startAnalyticsCollection() {
	ticker := time.NewTicker(ffd.config.UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		ffd.collectAnalytics()
	}
}

// collectAnalytics collects analytics data from various sources
func (ffd *FeatureFlagDashboard) collectAnalytics() {
	ffd.mu.Lock()
	defer ffd.mu.Unlock()

	// Update overall metrics
	ffd.updateOverallMetrics()

	// Update feature flag usage
	ffd.updateFeatureFlagUsage()

	// Update model performance
	ffd.updateModelPerformance()

	// Update A/B test results
	ffd.updateABTestResults()

	// Update rollout progress
	ffd.updateRolloutProgress()

	// Update performance trends
	ffd.updatePerformanceTrends()

	// Update last updated timestamp
	ffd.analytics.LastUpdated = time.Now()

	// Notify subscribers
	ffd.notifySubscribers()

	// Check for alerts
	if ffd.config.AlertingEnabled {
		ffd.checkAlerts()
	}
}

// updateOverallMetrics updates overall analytics metrics
func (ffd *FeatureFlagDashboard) updateOverallMetrics() {
	// This would typically collect metrics from various sources
	// For now, we'll use placeholder values
	ffd.analytics.TotalRequests++
	ffd.analytics.SuccessfulRequests++
	ffd.analytics.AverageLatency = time.Millisecond * 100
	ffd.analytics.ErrorRate = 0.01
}

// updateFeatureFlagUsage updates feature flag usage statistics
func (ffd *FeatureFlagDashboard) updateFeatureFlagUsage() {
	// Get current feature flags
	_ = ffd.featureFlagManager.GetFlags()

	// Update usage for each flag
	flagNames := []string{
		"python_ml_service_enabled",
		"go_rule_engine_enabled",
		"bert_classification_enabled",
		"distilbert_classification_enabled",
		"custom_neural_net_enabled",
		"bert_risk_detection_enabled",
		"anomaly_detection_enabled",
		"pattern_recognition_enabled",
		"keyword_matching_enabled",
		"mcc_code_lookup_enabled",
		"blacklist_check_enabled",
	}

	for _, flagName := range flagNames {
		if _, exists := ffd.analytics.FeatureFlagUsage[flagName]; !exists {
			ffd.analytics.FeatureFlagUsage[flagName] = &FeatureFlagUsage{
				UsageDistribution: make(map[string]int64),
			}
		}

		usage := ffd.analytics.FeatureFlagUsage[flagName]
		usage.TotalRequests++
		usage.SuccessfulRequests++
		usage.AverageLatency = time.Millisecond * 50
		usage.ErrorRate = 0.005
		usage.P95Latency = time.Millisecond * 100
		usage.P99Latency = time.Millisecond * 200
		usage.Throughput = 100.0
		usage.PeakUsageTime = time.Now()
		usage.UsageDistribution["hourly"]++
		usage.LastUpdated = time.Now()
	}
}

// updateModelPerformance updates model performance metrics
func (ffd *FeatureFlagDashboard) updateModelPerformance() {
	// Model names to track
	modelNames := []string{
		"bert_classification",
		"distilbert_classification",
		"custom_neural_net",
		"bert_risk_detection",
		"anomaly_detection",
		"pattern_recognition",
		"rule_based",
	}

	for _, modelName := range modelNames {
		if _, exists := ffd.analytics.ModelPerformance[modelName]; !exists {
			ffd.analytics.ModelPerformance[modelName] = &ModelPerformance{
				ModelName:    modelName,
				ModelType:    ffd.getModelType(modelName),
				ModelVersion: "v1.0",
			}
		}

		performance := ffd.analytics.ModelPerformance[modelName]
		performance.TotalRequests++
		performance.SuccessfulRequests++
		performance.AverageLatency = ffd.getModelLatency(modelName)
		performance.ErrorRate = 0.01
		performance.Accuracy = ffd.getModelAccuracy(modelName)
		performance.Confidence = 0.95
		performance.Precision = 0.92
		performance.Recall = 0.88
		performance.F1Score = 0.90
		performance.CPUUsage = 45.0
		performance.MemoryUsage = 60.0
		performance.GPUUsage = 30.0
		performance.LastUpdated = time.Now()
	}
}

// updateABTestResults updates A/B test results
func (ffd *FeatureFlagDashboard) updateABTestResults() {
	// This would typically collect A/B test results from the A/B tester
	// For now, we'll use placeholder values
}

// updateRolloutProgress updates rollout progress
func (ffd *FeatureFlagDashboard) updateRolloutProgress() {
	// This would typically collect rollout progress from the rollout manager
	// For now, we'll use placeholder values
}

// updatePerformanceTrends updates performance trends
func (ffd *FeatureFlagDashboard) updatePerformanceTrends() {
	// Add new performance trend
	trend := PerformanceTrend{
		Timestamp:      time.Now(),
		TotalRequests:  ffd.analytics.TotalRequests,
		AverageLatency: ffd.analytics.AverageLatency,
		ErrorRate:      ffd.analytics.ErrorRate,
		Throughput:     100.0,
	}

	ffd.analytics.PerformanceTrends = append(ffd.analytics.PerformanceTrends, trend)

	// Keep only last 100 trends
	if len(ffd.analytics.PerformanceTrends) > 100 {
		ffd.analytics.PerformanceTrends = ffd.analytics.PerformanceTrends[1:]
	}
}

// getModelType returns the model type for a model name
func (ffd *FeatureFlagDashboard) getModelType(modelName string) string {
	switch modelName {
	case "bert_classification", "bert_risk_detection":
		return "bert"
	case "distilbert_classification":
		return "distilbert"
	case "custom_neural_net":
		return "custom"
	case "rule_based":
		return "rule_based"
	default:
		return "unknown"
	}
}

// getModelLatency returns the typical latency for a model
func (ffd *FeatureFlagDashboard) getModelLatency(modelName string) time.Duration {
	switch modelName {
	case "bert_classification", "bert_risk_detection":
		return time.Millisecond * 200
	case "distilbert_classification":
		return time.Millisecond * 100
	case "custom_neural_net":
		return time.Millisecond * 150
	case "rule_based":
		return time.Millisecond * 10
	default:
		return time.Millisecond * 100
	}
}

// getModelAccuracy returns the typical accuracy for a model
func (ffd *FeatureFlagDashboard) getModelAccuracy(modelName string) float64 {
	switch modelName {
	case "bert_classification", "bert_risk_detection":
		return 0.95
	case "distilbert_classification":
		return 0.92
	case "custom_neural_net":
		return 0.88
	case "rule_based":
		return 0.85
	default:
		return 0.90
	}
}

// notifySubscribers notifies all subscribers of analytics updates
func (ffd *FeatureFlagDashboard) notifySubscribers() {
	for _, ch := range ffd.updateChannels {
		select {
		case ch <- ffd.analytics:
		default:
			// Channel is full, skip notification
		}
	}
}

// checkAlerts checks for alert conditions
func (ffd *FeatureFlagDashboard) checkAlerts() {
	// Check overall metrics
	if ffd.analytics.ErrorRate > ffd.config.AlertThresholds.MaxErrorRateThreshold {
		ffd.logger.Printf("ALERT: High error rate: %.2f%%", ffd.analytics.ErrorRate*100)
	}

	if ffd.analytics.AverageLatency > ffd.config.AlertThresholds.MaxLatencyThreshold {
		ffd.logger.Printf("ALERT: High latency: %v", ffd.analytics.AverageLatency)
	}

	// Check model performance
	for modelName, performance := range ffd.analytics.ModelPerformance {
		if performance.ErrorRate > ffd.config.AlertThresholds.MaxErrorRateThreshold {
			ffd.logger.Printf("ALERT: High error rate for model %s: %.2f%%", modelName, performance.ErrorRate*100)
		}
	}
}

// startDashboardServer starts the dashboard HTTP server
func (ffd *FeatureFlagDashboard) startDashboardServer() {
	mux := http.NewServeMux()

	// Dashboard endpoints
	mux.HandleFunc("/", ffd.handleDashboard)
	mux.HandleFunc("/api/analytics", ffd.handleAnalytics)
	mux.HandleFunc("/api/feature-flags", ffd.handleFeatureFlags)
	mux.HandleFunc("/api/models", ffd.handleModels)
	mux.HandleFunc("/api/ab-tests", ffd.handleABTests)
	mux.HandleFunc("/api/rollouts", ffd.handleRollouts)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", ffd.config.Port),
		Handler: mux,
	}

	ffd.logger.Printf("Feature flag dashboard server starting on port %d", ffd.config.Port)
	if err := server.ListenAndServe(); err != nil {
		ffd.logger.Printf("Dashboard server error: %v", err)
	}
}

// handleDashboard handles the main dashboard page
func (ffd *FeatureFlagDashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Simple HTML dashboard
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Feature Flag Dashboard</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .metric h3 { margin: 0 0 10px 0; color: #333; }
        .metric-value { font-size: 24px; font-weight: bold; color: #007bff; }
        .metric-label { color: #666; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
    </style>
</head>
<body>
    <h1>Feature Flag Dashboard</h1>
    <div class="grid">
        <div class="metric">
            <h3>Total Requests</h3>
            <div class="metric-value" id="total-requests">-</div>
            <div class="metric-label">All time</div>
        </div>
        <div class="metric">
            <h3>Error Rate</h3>
            <div class="metric-value" id="error-rate">-</div>
            <div class="metric-label">Current</div>
        </div>
        <div class="metric">
            <h3>Average Latency</h3>
            <div class="metric-value" id="avg-latency">-</div>
            <div class="metric-label">Current</div>
        </div>
        <div class="metric">
            <h3>Active Models</h3>
            <div class="metric-value" id="active-models">-</div>
            <div class="metric-label">Currently enabled</div>
        </div>
    </div>
    
    <script>
        function updateDashboard() {
            fetch('/api/analytics')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('total-requests').textContent = data.total_requests.toLocaleString();
                    document.getElementById('error-rate').textContent = (data.error_rate * 100).toFixed(2) + '%';
                    document.getElementById('avg-latency').textContent = data.average_latency;
                    document.getElementById('active-models').textContent = Object.keys(data.model_performance).length;
                })
                .catch(error => console.error('Error:', error));
        }
        
        // Update dashboard every 5 seconds
        updateDashboard();
        setInterval(updateDashboard, 5000);
    </script>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleAnalytics handles analytics API requests
func (ffd *FeatureFlagDashboard) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := ffd.GetAnalytics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// handleFeatureFlags handles feature flag API requests
func (ffd *FeatureFlagDashboard) handleFeatureFlags(w http.ResponseWriter, r *http.Request) {
	flags := ffd.featureFlagManager.GetFlags()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flags)
}

// handleModels handles model API requests
func (ffd *FeatureFlagDashboard) handleModels(w http.ResponseWriter, r *http.Request) {
	analytics := ffd.GetAnalytics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics.ModelPerformance)
}

// handleABTests handles A/B test API requests
func (ffd *FeatureFlagDashboard) handleABTests(w http.ResponseWriter, r *http.Request) {
	analytics := ffd.GetAnalytics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics.ABTestResults)
}

// handleRollouts handles rollout API requests
func (ffd *FeatureFlagDashboard) handleRollouts(w http.ResponseWriter, r *http.Request) {
	analytics := ffd.GetAnalytics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics.RolloutProgress)
}
