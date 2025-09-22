package automation

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"kyb-platform/internal/machine_learning/infrastructure"
)

// PerformanceMonitor handles performance monitoring and data drift detection
type PerformanceMonitor struct {
	// Core components
	mlService  *infrastructure.PythonMLService
	ruleEngine *infrastructure.GoRuleEngine

	// Monitoring configuration
	config *PerformanceMonitoringConfig

	// Performance tracking
	performanceMetrics map[string]*PerformanceMetric
	driftDetectors     map[string]*DriftDetector
	alertManager       *AlertManager

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger interface{}

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// PerformanceMonitoringConfig holds configuration for performance monitoring
type PerformanceMonitoringConfig struct {
	// Monitoring configuration
	Enabled               bool          `json:"enabled"`
	MonitoringInterval    time.Duration `json:"monitoring_interval"`
	MetricsRetentionDays  int           `json:"metrics_retention_days"`
	DriftDetectionEnabled bool          `json:"drift_detection_enabled"`
	AlertingEnabled       bool          `json:"alerting_enabled"`

	// Performance thresholds
	AccuracyThreshold   float64       `json:"accuracy_threshold"`
	LatencyThreshold    time.Duration `json:"latency_threshold"`
	ErrorRateThreshold  float64       `json:"error_rate_threshold"`
	ThroughputThreshold float64       `json:"throughput_threshold"`

	// Drift detection thresholds
	DriftThreshold          float64 `json:"drift_threshold"`
	StatisticalSignificance float64 `json:"statistical_significance"`
	MinimumSampleSize       int     `json:"minimum_sample_size"`

	// Alerting configuration
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
	AlertRecipients     []string      `json:"alert_recipients"`
	AlertChannels       []string      `json:"alert_channels"`

	// Data sources
	ReferenceDataSources []string `json:"reference_data_sources"`
	CurrentDataSources   []string `json:"current_data_sources"`
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	ModelID         string                 `json:"model_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Accuracy        float64                `json:"accuracy"`
	Precision       float64                `json:"precision"`
	Recall          float64                `json:"recall"`
	F1Score         float64                `json:"f1_score"`
	Latency         time.Duration          `json:"latency"`
	Throughput      float64                `json:"throughput"`
	ErrorRate       float64                `json:"error_rate"`
	ConfidenceScore float64                `json:"confidence_score"`
	ResourceUsage   *ResourceUsage         `json:"resource_usage"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics"`
}

// ResourceUsage represents resource usage metrics
type ResourceUsage struct {
	CPUUsage           float64 `json:"cpu_usage"`
	MemoryUsage        float64 `json:"memory_usage"`
	DiskUsage          float64 `json:"disk_usage"`
	NetworkUsage       float64 `json:"network_usage"`
	GPUUsage           float64 `json:"gpu_usage"`
	ConcurrentRequests int     `json:"concurrent_requests"`
}

// DriftDetector handles data drift detection
type DriftDetector struct {
	ModelID          string                `json:"model_id"`
	DriftType        string                `json:"drift_type"` // data, concept, prediction
	ReferenceData    []DataSample          `json:"reference_data"`
	CurrentData      []DataSample          `json:"current_data"`
	DriftScore       float64               `json:"drift_score"`
	DriftDetected    bool                  `json:"drift_detected"`
	LastDetection    time.Time             `json:"last_detection"`
	DetectionHistory []DriftDetection      `json:"detection_history"`
	Config           *DriftDetectionConfig `json:"config"`
}

// DataSample represents a data sample for drift detection
type DataSample struct {
	ID        string                 `json:"id"`
	Features  map[string]interface{} `json:"features"`
	Label     interface{}            `json:"label"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DriftDetection represents a drift detection result
type DriftDetection struct {
	Timestamp       time.Time `json:"timestamp"`
	DriftScore      float64   `json:"drift_score"`
	DriftDetected   bool      `json:"drift_detected"`
	DriftType       string    `json:"drift_type"`
	StatisticalTest string    `json:"statistical_test"`
	PValue          float64   `json:"p_value"`
	Confidence      float64   `json:"confidence"`
	Recommendation  string    `json:"recommendation"`
}

// DriftDetectionConfig holds configuration for drift detection
type DriftDetectionConfig struct {
	Enabled                 bool          `json:"enabled"`
	DriftThreshold          float64       `json:"drift_threshold"`
	StatisticalSignificance float64       `json:"statistical_significance"`
	MinimumSampleSize       int           `json:"minimum_sample_size"`
	WindowSize              int           `json:"window_size"`
	UpdateInterval          time.Duration `json:"update_interval"`
}

// AlertManager handles alerting for performance issues
type AlertManager struct {
	config       *AlertingConfig
	activeAlerts map[string]*Alert
	alertHistory []*Alert
	mu           sync.RWMutex
}

// AlertingConfig holds configuration for alerting
type AlertingConfig struct {
	Enabled            bool               `json:"enabled"`
	CooldownPeriod     time.Duration      `json:"cooldown_period"`
	Recipients         []string           `json:"recipients"`
	Channels           []string           `json:"channels"`
	SeverityThresholds map[string]float64 `json:"severity_thresholds"`
}

// Alert represents an alert
type Alert struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	ModelID    string                 `json:"model_id"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(
	mlService *infrastructure.PythonMLService,
	ruleEngine *infrastructure.GoRuleEngine,
	config *PerformanceMonitoringConfig,
	logger interface{},
) *PerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &PerformanceMonitor{
		mlService:          mlService,
		ruleEngine:         ruleEngine,
		config:             config,
		performanceMetrics: make(map[string]*PerformanceMetric),
		driftDetectors:     make(map[string]*DriftDetector),
		alertManager: NewAlertManager(&AlertingConfig{
			Enabled:        config.AlertingEnabled,
			CooldownPeriod: config.AlertCooldownPeriod,
			Recipients:     config.AlertRecipients,
			Channels:       config.AlertChannels,
		}),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start monitoring
	if config.Enabled {
		go monitor.startMonitoring()
	}

	return monitor
}

// startMonitoring starts the performance monitoring
func (pm *PerformanceMonitor) startMonitoring() {
	ticker := time.NewTicker(pm.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.collectPerformanceMetrics()
			if pm.config.DriftDetectionEnabled {
				pm.detectDrift()
			}
		}
	}
}

// collectPerformanceMetrics collects performance metrics for all models
func (pm *PerformanceMonitor) collectPerformanceMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Collect metrics for ML models
	if pm.mlService != nil {
		mlMetrics := pm.collectMLServiceMetrics()
		for modelID, metric := range mlMetrics {
			pm.performanceMetrics[modelID] = metric
		}
	}

	// Collect metrics for rule engine
	if pm.ruleEngine != nil {
		ruleMetrics := pm.collectRuleEngineMetrics()
		pm.performanceMetrics["rule_engine"] = ruleMetrics
	}

	// Check for performance issues and generate alerts
	pm.checkPerformanceThresholds()
}

// collectMLServiceMetrics collects metrics from the ML service
func (pm *PerformanceMonitor) collectMLServiceMetrics() map[string]*PerformanceMetric {
	metrics := make(map[string]*PerformanceMetric)

	// This would collect actual metrics from the ML service
	// For now, return placeholder metrics
	metrics["bert_classification"] = &PerformanceMetric{
		ModelID:         "bert_classification",
		Timestamp:       time.Now(),
		Accuracy:        0.95,
		Precision:       0.94,
		Recall:          0.96,
		F1Score:         0.95,
		Latency:         50 * time.Millisecond,
		Throughput:      100.0,
		ErrorRate:       0.02,
		ConfidenceScore: 0.93,
		ResourceUsage: &ResourceUsage{
			CPUUsage:           45.0,
			MemoryUsage:        512.0,
			DiskUsage:          1024.0,
			NetworkUsage:       10.0,
			ConcurrentRequests: 25,
		},
	}

	return metrics
}

// collectRuleEngineMetrics collects metrics from the rule engine
func (pm *PerformanceMonitor) collectRuleEngineMetrics() *PerformanceMetric {
	// This would collect actual metrics from the rule engine
	// For now, return placeholder metrics
	return &PerformanceMetric{
		ModelID:         "rule_engine",
		Timestamp:       time.Now(),
		Accuracy:        0.90,
		Precision:       0.89,
		Recall:          0.91,
		F1Score:         0.90,
		Latency:         5 * time.Millisecond,
		Throughput:      1000.0,
		ErrorRate:       0.01,
		ConfidenceScore: 0.88,
		ResourceUsage: &ResourceUsage{
			CPUUsage:           10.0,
			MemoryUsage:        128.0,
			DiskUsage:          256.0,
			NetworkUsage:       2.0,
			ConcurrentRequests: 100,
		},
	}
}

// checkPerformanceThresholds checks if performance metrics exceed thresholds
func (pm *PerformanceMonitor) checkPerformanceThresholds() {
	for modelID, metric := range pm.performanceMetrics {
		// Check accuracy threshold
		if metric.Accuracy < pm.config.AccuracyThreshold {
			pm.alertManager.CreateAlert(&Alert{
				ID:        fmt.Sprintf("accuracy_%s_%d", modelID, time.Now().Unix()),
				Type:      "performance",
				Severity:  "warning",
				ModelID:   modelID,
				Message:   fmt.Sprintf("Accuracy %.3f below threshold %.3f", metric.Accuracy, pm.config.AccuracyThreshold),
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"metric":    "accuracy",
					"value":     metric.Accuracy,
					"threshold": pm.config.AccuracyThreshold,
				},
			})
		}

		// Check latency threshold
		if metric.Latency > pm.config.LatencyThreshold {
			pm.alertManager.CreateAlert(&Alert{
				ID:        fmt.Sprintf("latency_%s_%d", modelID, time.Now().Unix()),
				Type:      "performance",
				Severity:  "warning",
				ModelID:   modelID,
				Message:   fmt.Sprintf("Latency %v above threshold %v", metric.Latency, pm.config.LatencyThreshold),
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"metric":    "latency",
					"value":     metric.Latency,
					"threshold": pm.config.LatencyThreshold,
				},
			})
		}

		// Check error rate threshold
		if metric.ErrorRate > pm.config.ErrorRateThreshold {
			pm.alertManager.CreateAlert(&Alert{
				ID:        fmt.Sprintf("error_rate_%s_%d", modelID, time.Now().Unix()),
				Type:      "critical",
				Severity:  "critical",
				ModelID:   modelID,
				Message:   fmt.Sprintf("Error rate %.3f above threshold %.3f", metric.ErrorRate, pm.config.ErrorRateThreshold),
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"metric":    "error_rate",
					"value":     metric.ErrorRate,
					"threshold": pm.config.ErrorRateThreshold,
				},
			})
		}

		// Check throughput threshold
		if metric.Throughput < pm.config.ThroughputThreshold {
			pm.alertManager.CreateAlert(&Alert{
				ID:        fmt.Sprintf("throughput_%s_%d", modelID, time.Now().Unix()),
				Type:      "performance",
				Severity:  "warning",
				ModelID:   modelID,
				Message:   fmt.Sprintf("Throughput %.1f below threshold %.1f", metric.Throughput, pm.config.ThroughputThreshold),
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"metric":    "throughput",
					"value":     metric.Throughput,
					"threshold": pm.config.ThroughputThreshold,
				},
			})
		}
	}
}

// detectDrift detects data drift for all models
func (pm *PerformanceMonitor) detectDrift() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for modelID, detector := range pm.driftDetectors {
		if detector.Config.Enabled {
			driftDetection := pm.performDriftDetection(detector)
			if driftDetection.DriftDetected {
				// Update detector
				detector.DriftScore = driftDetection.DriftScore
				detector.DriftDetected = true
				detector.LastDetection = driftDetection.Timestamp
				detector.DetectionHistory = append(detector.DetectionHistory, driftDetection)

				// Create alert
				pm.alertManager.CreateAlert(&Alert{
					ID:        fmt.Sprintf("drift_%s_%d", modelID, time.Now().Unix()),
					Type:      "drift",
					Severity:  "warning",
					ModelID:   modelID,
					Message:   fmt.Sprintf("Data drift detected: score %.3f", driftDetection.DriftScore),
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"drift_score": driftDetection.DriftScore,
						"drift_type":  driftDetection.DriftType,
						"confidence":  driftDetection.Confidence,
					},
				})
			}
		}
	}
}

// performDriftDetection performs drift detection for a specific detector
func (pm *PerformanceMonitor) performDriftDetection(detector *DriftDetector) DriftDetection {
	// This would implement actual drift detection algorithms
	// For now, return a placeholder implementation

	// Simulate drift detection
	driftScore := pm.calculateDriftScore(detector.ReferenceData, detector.CurrentData)
	driftDetected := driftScore > detector.Config.DriftThreshold

	return DriftDetection{
		Timestamp:       time.Now(),
		DriftScore:      driftScore,
		DriftDetected:   driftDetected,
		DriftType:       detector.DriftType,
		StatisticalTest: "ks_test", // Kolmogorov-Smirnov test
		PValue:          0.05,
		Confidence:      0.95,
		Recommendation:  pm.generateDriftRecommendation(driftScore, driftDetected),
	}
}

// calculateDriftScore calculates drift score between reference and current data
func (pm *PerformanceMonitor) calculateDriftScore(reference, current []DataSample) float64 {
	if len(reference) == 0 || len(current) == 0 {
		return 0.0
	}

	// Use multiple drift detection methods and combine results
	ksScore := pm.calculateKolmogorovSmirnovScore(reference, current)
	psiScore := pm.calculatePopulationStabilityIndex(reference, current)
	jsScore := pm.calculateJensenShannonDivergence(reference, current)

	// Weighted combination of different drift scores
	combinedScore := (0.4 * ksScore) + (0.3 * psiScore) + (0.3 * jsScore)
	return math.Min(combinedScore, 1.0) // Cap at 1.0
}

// calculateMean calculates mean value for data samples
func (pm *PerformanceMonitor) calculateMean(samples []DataSample) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	sum := 0.0
	count := 0

	for _, sample := range samples {
		// This is a simplified calculation
		// In reality, you'd extract numerical features properly
		if val, ok := sample.Features["value"].(float64); ok {
			sum += val
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// calculateKolmogorovSmirnovScore calculates KS test score for drift detection
func (pm *PerformanceMonitor) calculateKolmogorovSmirnovScore(reference, current []DataSample) float64 {
	if len(reference) == 0 || len(current) == 0 {
		return 0.0
	}

	// Extract numerical features for KS test
	refValues := pm.extractNumericalFeatures(reference)
	curValues := pm.extractNumericalFeatures(current)

	if len(refValues) == 0 || len(curValues) == 0 {
		return 0.0
	}

	// Sort values for KS test
	refValues = pm.sortFloat64Slice(refValues)
	curValues = pm.sortFloat64Slice(curValues)

	// Calculate empirical distribution functions
	n := len(refValues)
	m := len(curValues)

	maxDiff := 0.0
	i, j := 0, 0

	for i < n && j < m {
		if refValues[i] <= curValues[j] {
			i++
		} else {
			j++
		}

		// Calculate difference in empirical CDFs
		refCDF := float64(i) / float64(n)
		curCDF := float64(j) / float64(m)
		diff := math.Abs(refCDF - curCDF)

		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// KS statistic
	ksStat := maxDiff

	// Convert to drift score (0-1 scale)
	// Higher KS statistic indicates more drift
	return math.Min(ksStat*2, 1.0) // Scale and cap at 1.0
}

// calculatePopulationStabilityIndex calculates PSI for drift detection
func (pm *PerformanceMonitor) calculatePopulationStabilityIndex(reference, current []DataSample) float64 {
	if len(reference) == 0 || len(current) == 0 {
		return 0.0
	}

	// Extract numerical features
	refValues := pm.extractNumericalFeatures(reference)
	curValues := pm.extractNumericalFeatures(current)

	if len(refValues) == 0 || len(curValues) == 0 {
		return 0.0
	}

	// Create bins for PSI calculation
	numBins := 10
	refBins := pm.createBins(refValues, numBins)
	curBins := pm.createBins(curValues, numBins)

	// Calculate PSI
	psi := 0.0
	for i := 0; i < numBins; i++ {
		refPct := float64(refBins[i]) / float64(len(refValues))
		curPct := float64(curBins[i]) / float64(len(curValues))

		// Avoid division by zero
		if refPct > 0 && curPct > 0 {
			psi += (curPct - refPct) * math.Log(curPct/refPct)
		}
	}

	// Convert PSI to drift score (0-1 scale)
	// PSI > 0.2 indicates significant drift
	if psi > 0.2 {
		return math.Min(psi/0.5, 1.0) // Scale and cap at 1.0
	}
	return psi / 0.2 // Scale to 0-1 range
}

// calculateJensenShannonDivergence calculates JS divergence for drift detection
func (pm *PerformanceMonitor) calculateJensenShannonDivergence(reference, current []DataSample) float64 {
	if len(reference) == 0 || len(current) == 0 {
		return 0.0
	}

	// Extract numerical features
	refValues := pm.extractNumericalFeatures(reference)
	curValues := pm.extractNumericalFeatures(current)

	if len(refValues) == 0 || len(curValues) == 0 {
		return 0.0
	}

	// Create probability distributions
	refDist := pm.createProbabilityDistribution(refValues)
	curDist := pm.createProbabilityDistribution(curValues)

	// Calculate JS divergence
	jsDiv := pm.jensenShannonDivergence(refDist, curDist)

	// Convert to drift score (0-1 scale)
	// JS divergence is bounded by log(2), so normalize
	return math.Min(jsDiv/math.Log(2), 1.0)
}

// extractNumericalFeatures extracts numerical features from data samples
func (pm *PerformanceMonitor) extractNumericalFeatures(samples []DataSample) []float64 {
	values := make([]float64, 0, len(samples))

	for _, sample := range samples {
		// Extract numerical features (simplified)
		if val, ok := sample.Features["value"].(float64); ok {
			values = append(values, val)
		}
		// Add more feature extraction logic as needed
	}

	return values
}

// sortFloat64Slice sorts a slice of float64 values
func (pm *PerformanceMonitor) sortFloat64Slice(values []float64) []float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple bubble sort (in production, use sort.Float64s)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// createBins creates histogram bins for PSI calculation
func (pm *PerformanceMonitor) createBins(values []float64, numBins int) []int {
	if len(values) == 0 {
		return make([]int, numBins)
	}

	// Find min and max values
	minVal := values[0]
	maxVal := values[0]
	for _, val := range values {
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	// Create bins
	bins := make([]int, numBins)
	binWidth := (maxVal - minVal) / float64(numBins)

	if binWidth == 0 {
		// All values are the same
		bins[0] = len(values)
		return bins
	}

	for _, val := range values {
		binIndex := int((val - minVal) / binWidth)
		if binIndex >= numBins {
			binIndex = numBins - 1
		}
		bins[binIndex]++
	}

	return bins
}

// createProbabilityDistribution creates a probability distribution from values
func (pm *PerformanceMonitor) createProbabilityDistribution(values []float64) map[float64]float64 {
	distribution := make(map[float64]float64)
	total := float64(len(values))

	if total == 0 {
		return distribution
	}

	// Count occurrences
	counts := make(map[float64]int)
	for _, val := range values {
		counts[val]++
	}

	// Convert to probabilities
	for val, count := range counts {
		distribution[val] = float64(count) / total
	}

	return distribution
}

// jensenShannonDivergence calculates JS divergence between two distributions
func (pm *PerformanceMonitor) jensenShannonDivergence(p, q map[float64]float64) float64 {
	// Create combined distribution
	combined := make(map[float64]float64)

	// Add all keys from both distributions
	for key := range p {
		combined[key] = 0.0
	}
	for key := range q {
		combined[key] = 0.0
	}

	// Calculate average distribution
	for key := range combined {
		pVal := p[key]
		qVal := q[key]
		combined[key] = (pVal + qVal) / 2.0
	}

	// Calculate KL divergences
	klPQ := pm.kullbackLeiblerDivergence(p, combined)
	klQP := pm.kullbackLeiblerDivergence(q, combined)

	// JS divergence is the average of the two KL divergences
	return (klPQ + klQP) / 2.0
}

// kullbackLeiblerDivergence calculates KL divergence between two distributions
func (pm *PerformanceMonitor) kullbackLeiblerDivergence(p, q map[float64]float64) float64 {
	kl := 0.0

	for key, pVal := range p {
		qVal := q[key]
		if pVal > 0 && qVal > 0 {
			kl += pVal * math.Log(pVal/qVal)
		}
	}

	return kl
}

// generateDriftRecommendation generates recommendation based on drift detection
func (pm *PerformanceMonitor) generateDriftRecommendation(driftScore float64, driftDetected bool) string {
	if !driftDetected {
		return "No action required - no significant drift detected"
	}

	if driftScore > 0.5 {
		return "High drift detected - consider retraining model"
	} else if driftScore > 0.2 {
		return "Medium drift detected - monitor closely and consider model update"
	} else {
		return "Low drift detected - continue monitoring"
	}
}

// AddDriftDetector adds a drift detector for a model
func (pm *PerformanceMonitor) AddDriftDetector(modelID string, detector *DriftDetector) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.driftDetectors[modelID] = detector
}

// GetPerformanceMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetPerformanceMetrics() map[string]*PerformanceMetric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := make(map[string]*PerformanceMetric)
	for k, v := range pm.performanceMetrics {
		metrics[k] = v
	}
	return metrics
}

// GetDriftDetectors returns all drift detectors
func (pm *PerformanceMonitor) GetDriftDetectors() map[string]*DriftDetector {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to avoid race conditions
	detectors := make(map[string]*DriftDetector)
	for k, v := range pm.driftDetectors {
		detectors[k] = v
	}
	return detectors
}

// Stop stops the performance monitor
func (pm *PerformanceMonitor) Stop() {
	pm.cancel()
}

// NewAlertManager creates a new alert manager
func NewAlertManager(config *AlertingConfig) *AlertManager {
	return &AlertManager{
		config:       config,
		activeAlerts: make(map[string]*Alert),
		alertHistory: make([]*Alert, 0),
	}
}

// CreateAlert creates a new alert
func (am *AlertManager) CreateAlert(alert *Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if alert already exists (cooldown)
	if existingAlert, exists := am.activeAlerts[alert.ID]; exists {
		if time.Since(existingAlert.Timestamp) < am.config.CooldownPeriod {
			return // Skip duplicate alert
		}
	}

	am.activeAlerts[alert.ID] = alert
	am.alertHistory = append(am.alertHistory, alert)

	// Send alert (this would integrate with actual alerting systems)
	am.sendAlert(alert)
}

// sendAlert sends an alert through configured channels
func (am *AlertManager) sendAlert(alert *Alert) {
	// This would integrate with actual alerting systems (email, Slack, PagerDuty, etc.)
	fmt.Printf("ALERT [%s] %s: %s\n", alert.Severity, alert.ModelID, alert.Message)
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() map[string]*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Return a copy to avoid race conditions
	alerts := make(map[string]*Alert)
	for k, v := range am.activeAlerts {
		alerts[k] = v
	}
	return alerts
}

// ResolveAlert resolves an alert
func (am *AlertManager) ResolveAlert(alertID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.activeAlerts[alertID]; exists {
		alert.Resolved = true
		now := time.Now()
		alert.ResolvedAt = &now
		delete(am.activeAlerts, alertID)
	}
}

// Enhanced Performance Monitoring Methods

// RecordPrediction records a prediction for performance tracking
func (pm *PerformanceMonitor) RecordPrediction(modelID string, prediction interface{}, actual interface{}, latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update performance metrics
	if metric, exists := pm.performanceMetrics[modelID]; exists {
		// Update latency (rolling average)
		metric.Latency = time.Duration((int64(metric.Latency) + int64(latency)) / 2)

		// Update throughput (requests per second)
		metric.Throughput = pm.calculateThroughput(modelID)

		// Update accuracy if actual value is provided
		if actual != nil {
			correct := pm.isPredictionCorrect(prediction, actual)
			metric.Accuracy = pm.updateRollingAverage(metric.Accuracy, correct, 0.1) // 10% weight for new samples
		}

		// Update confidence score
		metric.ConfidenceScore = pm.calculateConfidenceScore(prediction)

		// Update resource usage
		metric.ResourceUsage = pm.getCurrentResourceUsage()
	}
}

// calculateThroughput calculates current throughput for a model
func (pm *PerformanceMonitor) calculateThroughput(modelID string) float64 {
	// This would track actual request counts and timing
	// For now, return a placeholder value
	return 100.0 // requests per second
}

// isPredictionCorrect checks if a prediction is correct
func (pm *PerformanceMonitor) isPredictionCorrect(prediction, actual interface{}) float64 {
	// This would implement actual prediction correctness checking
	// For now, return a placeholder value
	return 1.0 // 1.0 for correct, 0.0 for incorrect
}

// updateRollingAverage updates a rolling average with new value
func (pm *PerformanceMonitor) updateRollingAverage(current, newValue, weight float64) float64 {
	return current*(1-weight) + newValue*weight
}

// calculateConfidenceScore calculates confidence score for a prediction
func (pm *PerformanceMonitor) calculateConfidenceScore(prediction interface{}) float64 {
	// This would implement actual confidence calculation
	// For now, return a placeholder value
	return 0.95
}

// getCurrentResourceUsage gets current resource usage
func (pm *PerformanceMonitor) getCurrentResourceUsage() *ResourceUsage {
	// This would get actual resource usage from system
	// For now, return placeholder values
	return &ResourceUsage{
		CPUUsage:           45.0,
		MemoryUsage:        512.0,
		DiskUsage:          1024.0,
		NetworkUsage:       10.0,
		ConcurrentRequests: 25,
	}
}

// GetPerformanceSummary returns a summary of performance metrics
func (pm *PerformanceMonitor) GetPerformanceSummary() *PerformanceSummary {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := &PerformanceSummary{
		Timestamp:      time.Now(),
		TotalModels:    len(pm.performanceMetrics),
		ActiveAlerts:   len(pm.alertManager.GetActiveAlerts()),
		DriftDetectors: len(pm.driftDetectors),
		OverallHealth:  pm.calculateOverallHealth(),
		ModelSummaries: make(map[string]*ModelPerformanceSummary),
	}

	// Create model summaries
	for modelID, metric := range pm.performanceMetrics {
		summary.ModelSummaries[modelID] = &ModelPerformanceSummary{
			ModelID:         modelID,
			Accuracy:        metric.Accuracy,
			Latency:         metric.Latency,
			Throughput:      metric.Throughput,
			ErrorRate:       metric.ErrorRate,
			ConfidenceScore: metric.ConfidenceScore,
			HealthStatus:    pm.calculateModelHealth(metric),
			LastUpdated:     metric.Timestamp,
		}
	}

	return summary
}

// calculateOverallHealth calculates overall system health
func (pm *PerformanceMonitor) calculateOverallHealth() string {
	healthyModels := 0
	totalModels := len(pm.performanceMetrics)

	for _, metric := range pm.performanceMetrics {
		if pm.calculateModelHealth(metric) == "healthy" {
			healthyModels++
		}
	}

	if totalModels == 0 {
		return "unknown"
	}

	healthRatio := float64(healthyModels) / float64(totalModels)

	if healthRatio >= 0.9 {
		return "healthy"
	} else if healthRatio >= 0.7 {
		return "warning"
	} else {
		return "critical"
	}
}

// calculateModelHealth calculates health status for a model
func (pm *PerformanceMonitor) calculateModelHealth(metric *PerformanceMetric) string {
	// Check various health indicators
	if metric.Accuracy < pm.config.AccuracyThreshold {
		return "critical"
	}
	if metric.Latency > pm.config.LatencyThreshold {
		return "warning"
	}
	if metric.ErrorRate > pm.config.ErrorRateThreshold {
		return "critical"
	}
	if metric.Throughput < pm.config.ThroughputThreshold {
		return "warning"
	}

	return "healthy"
}

// PerformanceSummary represents a summary of performance metrics
type PerformanceSummary struct {
	Timestamp      time.Time                           `json:"timestamp"`
	TotalModels    int                                 `json:"total_models"`
	ActiveAlerts   int                                 `json:"active_alerts"`
	DriftDetectors int                                 `json:"drift_detectors"`
	OverallHealth  string                              `json:"overall_health"`
	ModelSummaries map[string]*ModelPerformanceSummary `json:"model_summaries"`
}

// ModelPerformanceSummary represents performance summary for a single model
type ModelPerformanceSummary struct {
	ModelID         string        `json:"model_id"`
	Accuracy        float64       `json:"accuracy"`
	Latency         time.Duration `json:"latency"`
	Throughput      float64       `json:"throughput"`
	ErrorRate       float64       `json:"error_rate"`
	ConfidenceScore float64       `json:"confidence_score"`
	HealthStatus    string        `json:"health_status"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// GetDriftSummary returns a summary of drift detection results
func (pm *PerformanceMonitor) GetDriftSummary() *DriftSummary {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := &DriftSummary{
		Timestamp:         time.Now(),
		TotalDetectors:    len(pm.driftDetectors),
		DriftDetected:     0,
		HighDriftModels:   make([]string, 0),
		MediumDriftModels: make([]string, 0),
		LowDriftModels:    make([]string, 0),
		NoDriftModels:     make([]string, 0),
	}

	for modelID, detector := range pm.driftDetectors {
		if detector.DriftDetected {
			summary.DriftDetected++

			if detector.DriftScore > 0.5 {
				summary.HighDriftModels = append(summary.HighDriftModels, modelID)
			} else if detector.DriftScore > 0.2 {
				summary.MediumDriftModels = append(summary.MediumDriftModels, modelID)
			} else {
				summary.LowDriftModels = append(summary.LowDriftModels, modelID)
			}
		} else {
			summary.NoDriftModels = append(summary.NoDriftModels, modelID)
		}
	}

	return summary
}

// DriftSummary represents a summary of drift detection results
type DriftSummary struct {
	Timestamp         time.Time `json:"timestamp"`
	TotalDetectors    int       `json:"total_detectors"`
	DriftDetected     int       `json:"drift_detected"`
	HighDriftModels   []string  `json:"high_drift_models"`
	MediumDriftModels []string  `json:"medium_drift_models"`
	LowDriftModels    []string  `json:"low_drift_models"`
	NoDriftModels     []string  `json:"no_drift_models"`
}
