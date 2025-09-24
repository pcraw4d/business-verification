package data_discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ExtractionMonitor provides comprehensive monitoring and optimization for data point extraction
type ExtractionMonitor struct {
	config    *ExtractionMonitorConfig
	logger    *zap.Logger
	metrics   *ExtractionMetrics
	optimizer *ExtractionOptimizer
	alerts    *AlertManager
	mu        sync.RWMutex
	startTime time.Time
	stopChan  chan struct{}
}

// ExtractionMonitorConfig defines configuration for the extraction monitoring system
type ExtractionMonitorConfig struct {
	// Monitoring settings
	MetricsCollectionInterval time.Duration         `json:"metrics_collection_interval"`
	PerformanceThresholds     PerformanceThresholds `json:"performance_thresholds"`
	AlertSettings             AlertSettings         `json:"alert_settings"`

	// Optimization settings
	OptimizationEnabled      bool                   `json:"optimization_enabled"`
	AutoOptimizationInterval time.Duration          `json:"auto_optimization_interval"`
	OptimizationThresholds   OptimizationThresholds `json:"optimization_thresholds"`

	// Storage settings
	MetricsRetentionPeriod time.Duration `json:"metrics_retention_period"`
	MaxMetricsHistory      int           `json:"max_metrics_history"`
}

// PerformanceThresholds defines performance thresholds for monitoring
type PerformanceThresholds struct {
	MaxProcessingTime        time.Duration `json:"max_processing_time"`
	MinSuccessRate           float64       `json:"min_success_rate"`
	MaxErrorRate             float64       `json:"max_error_rate"`
	MinDataPointsPerBusiness int           `json:"min_data_points_per_business"`
	MaxMemoryUsage           int64         `json:"max_memory_usage_mb"`
	MinQualityScore          float64       `json:"min_quality_score"`
}

// AlertSettings defines alert configuration
type AlertSettings struct {
	Enabled             bool          `json:"enabled"`
	AlertChannels       []string      `json:"alert_channels"`
	CriticalThreshold   float64       `json:"critical_threshold"`
	WarningThreshold    float64       `json:"warning_threshold"`
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
}

// OptimizationThresholds defines thresholds for triggering optimization
type OptimizationThresholds struct {
	PerformanceDegradationThreshold float64 `json:"performance_degradation_threshold"`
	QualityImprovementThreshold     float64 `json:"quality_improvement_threshold"`
	ResourceUtilizationThreshold    float64 `json:"resource_utilization_threshold"`
	SuccessRateImprovementThreshold float64 `json:"success_rate_improvement_threshold"`
}

// ExtractionMetrics tracks comprehensive metrics for data extraction
type ExtractionMetrics struct {
	mu sync.RWMutex

	// Performance metrics
	TotalRequests         int64         `json:"total_requests"`
	SuccessfulRequests    int64         `json:"successful_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	MaxProcessingTime     time.Duration `json:"max_processing_time"`
	MinProcessingTime     time.Duration `json:"min_processing_time"`

	// Quality metrics
	AverageQualityScore        float64        `json:"average_quality_score"`
	QualityScoreDistribution   map[string]int `json:"quality_score_distribution"`
	FieldsDiscoveredPerRequest float64        `json:"fields_discovered_per_request"`

	// Resource metrics
	MemoryUsage        int64   `json:"memory_usage_mb"`
	CPUUsage           float64 `json:"cpu_usage_percent"`
	ConcurrentRequests int     `json:"concurrent_requests"`

	// Field-specific metrics
	FieldDiscoveryRates  map[string]float64       `json:"field_discovery_rates"`
	FieldQualityScores   map[string]float64       `json:"field_quality_scores"`
	FieldProcessingTimes map[string]time.Duration `json:"field_processing_times"`

	// Error metrics
	ErrorTypes map[string]int64   `json:"error_types"`
	ErrorRates map[string]float64 `json:"error_rates"`

	// Historical data
	History     []MetricsSnapshot `json:"history"`
	LastUpdated time.Time         `json:"last_updated"`
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	Timestamp             time.Time     `json:"timestamp"`
	TotalRequests         int64         `json:"total_requests"`
	SuccessRate           float64       `json:"success_rate"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	AverageQualityScore   float64       `json:"average_quality_score"`
	MemoryUsage           int64         `json:"memory_usage_mb"`
	CPUUsage              float64       `json:"cpu_usage_percent"`
	FieldsDiscovered      float64       `json:"fields_discovered"`
}

// ExtractionOptimizer provides optimization capabilities for data extraction
type ExtractionOptimizer struct {
	config     *ExtractionMonitorConfig
	logger     *zap.Logger
	metrics    *ExtractionMetrics
	strategies []OptimizationStrategy
	mu         sync.RWMutex
}

// OptimizationStrategy defines an optimization strategy
type OptimizationStrategy struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Enabled       bool                   `json:"enabled"`
	Priority      int                    `json:"priority"`
	Parameters    map[string]interface{} `json:"parameters"`
	LastApplied   time.Time              `json:"last_applied"`
	Effectiveness float64                `json:"effectiveness"`
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	config    *ExtractionMonitorConfig
	logger    *zap.Logger
	alerts    []Alert
	lastAlert time.Time
	mu        sync.RWMutex
}

// Alert represents a monitoring alert
type Alert struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`     // "performance", "quality", "error", "resource"
	Severity     string                 `json:"severity"` // "critical", "warning", "info"
	Message      string                 `json:"message"`
	Timestamp    time.Time              `json:"timestamp"`
	Acknowledged bool                   `json:"acknowledged"`
	Resolved     bool                   `json:"resolved"`
	Metrics      map[string]interface{} `json:"metrics"`
}

// DefaultExtractionMonitorConfig returns default configuration for extraction monitoring
func DefaultExtractionMonitorConfig() *ExtractionMonitorConfig {
	return &ExtractionMonitorConfig{
		MetricsCollectionInterval: 30 * time.Second,
		PerformanceThresholds: PerformanceThresholds{
			MaxProcessingTime:        5 * time.Second,
			MinSuccessRate:           0.95,
			MaxErrorRate:             0.05,
			MinDataPointsPerBusiness: 8,
			MaxMemoryUsage:           512, // MB
			MinQualityScore:          0.7,
		},
		AlertSettings: AlertSettings{
			Enabled:             true,
			AlertChannels:       []string{"log", "metrics"},
			CriticalThreshold:   0.8,
			WarningThreshold:    0.9,
			AlertCooldownPeriod: 5 * time.Minute,
		},
		OptimizationEnabled:      true,
		AutoOptimizationInterval: 10 * time.Minute,
		OptimizationThresholds: OptimizationThresholds{
			PerformanceDegradationThreshold: 0.1,
			QualityImprovementThreshold:     0.05,
			ResourceUtilizationThreshold:    0.8,
			SuccessRateImprovementThreshold: 0.02,
		},
		MetricsRetentionPeriod: 24 * time.Hour,
		MaxMetricsHistory:      1000,
	}
}

// NewExtractionMonitor creates a new extraction monitoring system
func NewExtractionMonitor(config *ExtractionMonitorConfig, logger *zap.Logger) *ExtractionMonitor {
	if config == nil {
		config = DefaultExtractionMonitorConfig()
	}

	monitor := &ExtractionMonitor{
		config:    config,
		logger:    logger,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
	}

	// Initialize metrics
	monitor.metrics = &ExtractionMetrics{
		QualityScoreDistribution: make(map[string]int),
		FieldDiscoveryRates:      make(map[string]float64),
		FieldQualityScores:       make(map[string]float64),
		FieldProcessingTimes:     make(map[string]time.Duration),
		ErrorTypes:               make(map[string]int64),
		ErrorRates:               make(map[string]float64),
		History:                  make([]MetricsSnapshot, 0),
	}

	// Initialize optimizer
	monitor.optimizer = NewExtractionOptimizer(config, logger, monitor.metrics)

	// Initialize alert manager
	monitor.alerts = NewAlertManager(config, logger)

	// Start background monitoring
	go monitor.startBackgroundMonitoring()

	return monitor
}

// RecordExtractionResult records metrics for a completed extraction
func (em *ExtractionMonitor) RecordExtractionResult(ctx context.Context, result *DataDiscoveryResult, processingTime time.Duration, err error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Update basic metrics
	em.metrics.TotalRequests++
	if err != nil {
		em.metrics.FailedRequests++
		em.recordError(err)
	} else {
		em.metrics.SuccessfulRequests++
	}

	// Update processing time metrics
	em.updateProcessingTimeMetrics(processingTime)

	// Update quality metrics if result is available
	if result != nil {
		em.updateQualityMetrics(result)
		em.updateFieldMetrics(result)
	}

	// Update resource metrics
	em.updateResourceMetrics()

	// Create snapshot
	em.createMetricsSnapshot()

	// Trigger optimization if enabled
	if em.config.OptimizationEnabled {
		em.triggerOptimization()
	}

	em.logger.Debug("Recorded extraction result",
		zap.Duration("processing_time", processingTime),
		zap.Int("fields_discovered", func() int {
			if result != nil {
				return len(result.DiscoveredFields)
			}
			return 0
		}()),
		zap.Float64("confidence_score", func() float64 {
			if result != nil {
				return result.ConfidenceScore
			}
			return 0.0
		}()),
		zap.Error(err))
}

// GetMetrics returns current metrics
func (em *ExtractionMonitor) GetMetrics() *ExtractionMetrics {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &ExtractionMetrics{}
	*metrics = *em.metrics

	// Deep copy maps
	metrics.QualityScoreDistribution = make(map[string]int)
	for k, v := range em.metrics.QualityScoreDistribution {
		metrics.QualityScoreDistribution[k] = v
	}

	metrics.FieldDiscoveryRates = make(map[string]float64)
	for k, v := range em.metrics.FieldDiscoveryRates {
		metrics.FieldDiscoveryRates[k] = v
	}

	metrics.FieldQualityScores = make(map[string]float64)
	for k, v := range em.metrics.FieldQualityScores {
		metrics.FieldQualityScores[k] = v
	}

	metrics.FieldProcessingTimes = make(map[string]time.Duration)
	for k, v := range em.metrics.FieldProcessingTimes {
		metrics.FieldProcessingTimes[k] = v
	}

	metrics.ErrorTypes = make(map[string]int64)
	for k, v := range em.metrics.ErrorTypes {
		metrics.ErrorTypes[k] = v
	}

	metrics.ErrorRates = make(map[string]float64)
	for k, v := range em.metrics.ErrorRates {
		metrics.ErrorRates[k] = v
	}

	return metrics
}

// GetPerformanceReport returns a comprehensive performance report
func (em *ExtractionMonitor) GetPerformanceReport() *PerformanceReport {
	metrics := em.GetMetrics()

	report := &PerformanceReport{
		Timestamp:                   time.Now(),
		Uptime:                      time.Since(em.startTime),
		TotalRequests:               metrics.TotalRequests,
		SuccessRate:                 em.calculateSuccessRate(metrics),
		AverageProcessingTime:       metrics.AverageProcessingTime,
		AverageQualityScore:         metrics.AverageQualityScore,
		FieldsDiscoveredPerRequest:  metrics.FieldsDiscoveredPerRequest,
		MemoryUsage:                 metrics.MemoryUsage,
		CPUUsage:                    metrics.CPUUsage,
		ConcurrentRequests:          metrics.ConcurrentRequests,
		QualityDistribution:         metrics.QualityScoreDistribution,
		TopPerformingFields:         em.getTopPerformingFields(metrics),
		ProblematicFields:           em.getProblematicFields(metrics),
		ErrorAnalysis:               em.getErrorAnalysis(metrics),
		OptimizationRecommendations: em.getOptimizationRecommendations(),
		Alerts:                      em.alerts.GetActiveAlerts(),
	}

	return report
}

// PerformanceReport represents a comprehensive performance report
type PerformanceReport struct {
	Timestamp                   time.Time                    `json:"timestamp"`
	Uptime                      time.Duration                `json:"uptime"`
	TotalRequests               int64                        `json:"total_requests"`
	SuccessRate                 float64                      `json:"success_rate"`
	AverageProcessingTime       time.Duration                `json:"average_processing_time"`
	AverageQualityScore         float64                      `json:"average_quality_score"`
	FieldsDiscoveredPerRequest  float64                      `json:"fields_discovered_per_request"`
	MemoryUsage                 int64                        `json:"memory_usage_mb"`
	CPUUsage                    float64                      `json:"cpu_usage_percent"`
	ConcurrentRequests          int                          `json:"concurrent_requests"`
	QualityDistribution         map[string]int               `json:"quality_distribution"`
	TopPerformingFields         []FieldPerformance           `json:"top_performing_fields"`
	ProblematicFields           []FieldPerformance           `json:"problematic_fields"`
	ErrorAnalysis               ErrorAnalysis                `json:"error_analysis"`
	OptimizationRecommendations []OptimizationRecommendation `json:"optimization_recommendations"`
	Alerts                      []Alert                      `json:"alerts"`
}

// FieldPerformance represents performance metrics for a specific field
type FieldPerformance struct {
	FieldName             string        `json:"field_name"`
	DiscoveryRate         float64       `json:"discovery_rate"`
	AverageQualityScore   float64       `json:"average_quality_score"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	SuccessRate           float64       `json:"success_rate"`
	ErrorRate             float64       `json:"error_rate"`
	BusinessImpact        string        `json:"business_impact"`
}

// ErrorAnalysis provides detailed error analysis
type ErrorAnalysis struct {
	TotalErrors       int64              `json:"total_errors"`
	ErrorRate         float64            `json:"error_rate"`
	MostCommonErrors  []ErrorOccurrence  `json:"most_common_errors"`
	ErrorTrends       map[string]float64 `json:"error_trends"`
	RootCauseAnalysis []RootCause        `json:"root_cause_analysis"`
}

// ErrorOccurrence represents an error occurrence
type ErrorOccurrence struct {
	ErrorType    string    `json:"error_type"`
	Count        int64     `json:"count"`
	Percentage   float64   `json:"percentage"`
	Description  string    `json:"description"`
	LastOccurred time.Time `json:"last_occurred"`
}

// RootCause represents a root cause analysis
type RootCause struct {
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	Confidence     float64 `json:"confidence"`
	Impact         string  `json:"impact"`
	Recommendation string  `json:"recommendation"`
}

// OptimizationRecommendation represents an optimization recommendation
type OptimizationRecommendation struct {
	Type               string   `json:"type"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	ExpectedImpact     float64  `json:"expected_impact"`
	ImplementationCost string   `json:"implementation_cost"`
	Actions            []string `json:"actions"`
}

// Helper methods for metrics updates
func (em *ExtractionMonitor) updateProcessingTimeMetrics(processingTime time.Duration) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	// Update average processing time
	if em.metrics.TotalRequests == 1 {
		em.metrics.AverageProcessingTime = processingTime
		em.metrics.MaxProcessingTime = processingTime
		em.metrics.MinProcessingTime = processingTime
	} else {
		// Calculate running average
		total := em.metrics.AverageProcessingTime * time.Duration(em.metrics.TotalRequests-1)
		em.metrics.AverageProcessingTime = (total + processingTime) / time.Duration(em.metrics.TotalRequests)

		// Update min/max
		if processingTime > em.metrics.MaxProcessingTime {
			em.metrics.MaxProcessingTime = processingTime
		}
		if processingTime < em.metrics.MinProcessingTime {
			em.metrics.MinProcessingTime = processingTime
		}
	}
}

func (em *ExtractionMonitor) updateQualityMetrics(result *DataDiscoveryResult) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	// Update fields discovered per request
	fieldsCount := float64(len(result.DiscoveredFields))
	if em.metrics.TotalRequests == 1 {
		em.metrics.FieldsDiscoveredPerRequest = fieldsCount
	} else {
		total := em.metrics.FieldsDiscoveredPerRequest * float64(em.metrics.TotalRequests-1)
		em.metrics.FieldsDiscoveredPerRequest = (total + fieldsCount) / float64(em.metrics.TotalRequests)
	}

	// Update quality score distribution
	if len(result.QualityAssessments) > 0 {
		for _, assessment := range result.QualityAssessments {
			category := assessment.QualityCategory
			em.metrics.QualityScoreDistribution[category]++
		}

		// Calculate average quality score
		totalScore := 0.0
		for _, assessment := range result.QualityAssessments {
			totalScore += assessment.QualityScore.OverallScore
		}
		avgScore := totalScore / float64(len(result.QualityAssessments))

		if em.metrics.TotalRequests == 1 {
			em.metrics.AverageQualityScore = avgScore
		} else {
			total := em.metrics.AverageQualityScore * float64(em.metrics.TotalRequests-1)
			em.metrics.AverageQualityScore = (total + avgScore) / float64(em.metrics.TotalRequests)
		}
	}
}

func (em *ExtractionMonitor) updateFieldMetrics(result *DataDiscoveryResult) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	// Update field-specific metrics
	for _, field := range result.DiscoveredFields {
		// Update discovery rates
		if _, exists := em.metrics.FieldDiscoveryRates[field.FieldType]; !exists {
			em.metrics.FieldDiscoveryRates[field.FieldType] = 0
		}
		em.metrics.FieldDiscoveryRates[field.FieldType]++

		// Update quality scores
		for _, assessment := range result.QualityAssessments {
			if assessment.FieldName == field.FieldName {
				if _, exists := em.metrics.FieldQualityScores[field.FieldType]; !exists {
					em.metrics.FieldQualityScores[field.FieldType] = 0
				}
				em.metrics.FieldQualityScores[field.FieldType] = (em.metrics.FieldQualityScores[field.FieldType] + assessment.QualityScore.OverallScore) / 2
				break
			}
		}
	}
}

func (em *ExtractionMonitor) updateResourceMetrics() {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	// This would typically integrate with system monitoring
	// For now, we'll use placeholder values
	em.metrics.MemoryUsage = 128 // MB
	em.metrics.CPUUsage = 15.5   // Percent
	em.metrics.ConcurrentRequests = 5
}

func (em *ExtractionMonitor) recordError(err error) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	errorType := "unknown"
	if err != nil {
		errorType = fmt.Sprintf("%T", err)
	}

	em.metrics.ErrorTypes[errorType]++
}

func (em *ExtractionMonitor) createMetricsSnapshot() {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	snapshot := MetricsSnapshot{
		Timestamp:             time.Now(),
		TotalRequests:         em.metrics.TotalRequests,
		SuccessRate:           em.calculateSuccessRate(em.metrics),
		AverageProcessingTime: em.metrics.AverageProcessingTime,
		AverageQualityScore:   em.metrics.AverageQualityScore,
		MemoryUsage:           em.metrics.MemoryUsage,
		CPUUsage:              em.metrics.CPUUsage,
		FieldsDiscovered:      em.metrics.FieldsDiscoveredPerRequest,
	}

	em.metrics.History = append(em.metrics.History, snapshot)

	// Trim history if it exceeds max size
	if len(em.metrics.History) > em.config.MaxMetricsHistory {
		em.metrics.History = em.metrics.History[1:]
	}

	em.metrics.LastUpdated = time.Now()
}

func (em *ExtractionMonitor) calculateSuccessRate(metrics *ExtractionMetrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0.0
	}
	return float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
}

func (em *ExtractionMonitor) startBackgroundMonitoring() {
	ticker := time.NewTicker(em.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Use a goroutine to avoid blocking
			go em.collectBackgroundMetrics()
		case <-em.stopChan:
			return
		}
	}
}

func (em *ExtractionMonitor) collectBackgroundMetrics() {
	// Collect system metrics, check for alerts, etc.
	em.checkAlerts()

	if em.config.OptimizationEnabled {
		em.triggerOptimization()
	}
}

func (em *ExtractionMonitor) checkAlerts() {
	em.mu.RLock()
	metrics := em.metrics
	successRate := em.calculateSuccessRate(metrics)
	em.mu.RUnlock()

	// Check performance thresholds
	if successRate < em.config.PerformanceThresholds.MinSuccessRate {
		em.alerts.CreateAlert("performance", "critical",
			"Success rate below threshold", metrics)
	}

	if metrics.AverageProcessingTime > em.config.PerformanceThresholds.MaxProcessingTime {
		em.alerts.CreateAlert("performance", "warning",
			"Processing time above threshold", metrics)
	}

	if metrics.AverageQualityScore < em.config.PerformanceThresholds.MinQualityScore {
		em.alerts.CreateAlert("quality", "warning",
			"Quality score below threshold", metrics)
	}
}

func (em *ExtractionMonitor) triggerOptimization() {
	// This would trigger optimization strategies
	em.optimizer.RunOptimization()
}

func (em *ExtractionMonitor) getTopPerformingFields(metrics *ExtractionMetrics) []FieldPerformance {
	// Implementation would analyze field performance and return top performers
	return []FieldPerformance{}
}

func (em *ExtractionMonitor) getProblematicFields(metrics *ExtractionMetrics) []FieldPerformance {
	// Implementation would identify problematic fields
	return []FieldPerformance{}
}

func (em *ExtractionMonitor) getErrorAnalysis(metrics *ExtractionMetrics) ErrorAnalysis {
	// Implementation would provide detailed error analysis
	return ErrorAnalysis{}
}

func (em *ExtractionMonitor) getOptimizationRecommendations() []OptimizationRecommendation {
	// Implementation would generate optimization recommendations
	return []OptimizationRecommendation{}
}

// Stop gracefully shuts down the extraction monitor
func (em *ExtractionMonitor) Stop() {
	close(em.stopChan)
}
