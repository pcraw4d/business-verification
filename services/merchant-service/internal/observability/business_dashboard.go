package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BusinessDashboard provides comprehensive business metrics dashboard functionality
type BusinessDashboard struct {
	logger           *Logger
	metricsCollector *MetricsCollector
	config           *BusinessDashboardConfig
	businessData     map[string]*BusinessData
	exporters        []BusinessDashboardExporter
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	started          bool
}

// BusinessDashboardConfig holds configuration for business dashboard
type BusinessDashboardConfig struct {
	Enabled             bool
	RefreshInterval     time.Duration
	DataRetentionPeriod time.Duration
	MaxDataPoints       int
	ExportEnabled       bool
	ExportInterval      time.Duration
	Environment         string
	ServiceName         string
	Version             string
}

// BusinessData represents business dashboard data
type BusinessData struct {
	Timestamp             time.Time              `json:"timestamp"`
	ClassificationMetrics *ClassificationMetrics `json:"classification_metrics"`
	RiskAssessmentMetrics *RiskAssessmentMetrics `json:"risk_assessment_metrics"`
	ComplianceMetrics     *ComplianceMetrics     `json:"compliance_metrics"`
	UserMetrics           *UserMetrics           `json:"user_metrics"`
	APIMetrics            *APIMetrics            `json:"api_metrics"`
	PerformanceMetrics    *PerformanceMetrics    `json:"performance_metrics"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// ClassificationMetrics represents classification-related metrics
type ClassificationMetrics struct {
	TotalRequests          int64              `json:"total_requests"`
	SuccessfulRequests     int64              `json:"successful_requests"`
	FailedRequests         int64              `json:"failed_requests"`
	AverageAccuracy        float64            `json:"average_accuracy"`
	AverageConfidence      float64            `json:"average_confidence"`
	ProcessingTime         time.Duration      `json:"processing_time"`
	RequestsByMethod       map[string]int64   `json:"requests_by_method"`
	RequestsByIndustry     map[string]int64   `json:"requests_by_industry"`
	ConfidenceDistribution map[string]int64   `json:"confidence_distribution"`
	AccuracyByMethod       map[string]float64 `json:"accuracy_by_method"`
	ErrorRate              float64            `json:"error_rate"`
	Throughput             float64            `json:"throughput"`
}

// RiskAssessmentMetrics represents risk assessment metrics
type RiskAssessmentMetrics struct {
	TotalAssessments  int64            `json:"total_assessments"`
	HighRiskCount     int64            `json:"high_risk_count"`
	MediumRiskCount   int64            `json:"medium_risk_count"`
	LowRiskCount      int64            `json:"low_risk_count"`
	AverageRiskScore  float64          `json:"average_risk_score"`
	ProcessingTime    time.Duration    `json:"processing_time"`
	RiskDistribution  map[string]int64 `json:"risk_distribution"`
	AssessmentsByType map[string]int64 `json:"assessments_by_type"`
	FalsePositiveRate float64          `json:"false_positive_rate"`
	FalseNegativeRate float64          `json:"false_negative_rate"`
	Throughput        float64          `json:"throughput"`
}

// ComplianceMetrics represents compliance-related metrics
type ComplianceMetrics struct {
	TotalChecks       int64            `json:"total_checks"`
	PassedChecks      int64            `json:"passed_checks"`
	FailedChecks      int64            `json:"failed_checks"`
	ComplianceRate    float64          `json:"compliance_rate"`
	ChecksByFramework map[string]int64 `json:"checks_by_framework"`
	ChecksByStatus    map[string]int64 `json:"checks_by_status"`
	AverageCheckTime  time.Duration    `json:"average_check_time"`
	CriticalFailures  int64            `json:"critical_failures"`
	WarningCount      int64            `json:"warning_count"`
	Throughput        float64          `json:"throughput"`
}

// UserMetrics represents user-related metrics
type UserMetrics struct {
	TotalUsers      int64            `json:"total_users"`
	ActiveUsers     int64            `json:"active_users"`
	NewUsers        int64            `json:"new_users"`
	UserActivity    map[string]int64 `json:"user_activity"`
	SessionDuration time.Duration    `json:"session_duration"`
	PageViews       int64            `json:"page_views"`
	UniqueVisitors  int64            `json:"unique_visitors"`
	BounceRate      float64          `json:"bounce_rate"`
	ConversionRate  float64          `json:"conversion_rate"`
	UserRetention   float64          `json:"user_retention"`
}

// APIMetrics represents API-related metrics
type APIMetrics struct {
	TotalRequests          int64            `json:"total_requests"`
	SuccessfulRequests     int64            `json:"successful_requests"`
	FailedRequests         int64            `json:"failed_requests"`
	AverageResponseTime    time.Duration    `json:"average_response_time"`
	RequestsByEndpoint     map[string]int64 `json:"requests_by_endpoint"`
	RequestsByMethod       map[string]int64 `json:"requests_by_method"`
	RequestsByStatus       map[string]int64 `json:"requests_by_status"`
	ErrorRate              float64          `json:"error_rate"`
	Throughput             float64          `json:"throughput"`
	RateLimitHits          int64            `json:"rate_limit_hits"`
	AuthenticationFailures int64            `json:"authentication_failures"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"`
	ErrorRate           float64       `json:"error_rate"`
	Availability        float64       `json:"availability"`
	Uptime              time.Duration `json:"uptime"`
	CPUUsage            float64       `json:"cpu_usage"`
	MemoryUsage         float64       `json:"memory_usage"`
	DiskUsage           float64       `json:"disk_usage"`
	NetworkLatency      time.Duration `json:"network_latency"`
	DatabaseLatency     time.Duration `json:"database_latency"`
}

// BusinessDashboardExporter interface for exporting business dashboard data
type BusinessDashboardExporter interface {
	Export(data *BusinessData) error
	Name() string
	Type() string
}

// JSONBusinessDashboardExporter exports business dashboard data as JSON
type JSONBusinessDashboardExporter struct {
	logger *Logger
}

// NewJSONBusinessDashboardExporter creates a new JSON business dashboard exporter
func NewJSONBusinessDashboardExporter(logger *Logger) *JSONBusinessDashboardExporter {
	return &JSONBusinessDashboardExporter{
		logger: logger,
	}
}

// Export exports business dashboard data as JSON
func (jbde *JSONBusinessDashboardExporter) Export(data *BusinessData) error {
	jbde.logger.Debug("Business dashboard data exported as JSON", map[string]interface{}{
		"timestamp":               data.Timestamp,
		"classification_requests": data.ClassificationMetrics.TotalRequests,
		"risk_assessments":        data.RiskAssessmentMetrics.TotalAssessments,
		"compliance_checks":       data.ComplianceMetrics.TotalChecks,
		"active_users":            data.UserMetrics.ActiveUsers,
	})

	return nil
}

// Name returns the exporter name
func (jbde *JSONBusinessDashboardExporter) Name() string {
	return "json"
}

// Type returns the exporter type
func (jbde *JSONBusinessDashboardExporter) Type() string {
	return "json"
}

// PrometheusBusinessDashboardExporter exports business dashboard data to Prometheus
type PrometheusBusinessDashboardExporter struct {
	logger *Logger
}

// NewPrometheusBusinessDashboardExporter creates a new Prometheus business dashboard exporter
func NewPrometheusBusinessDashboardExporter(logger *Logger) *PrometheusBusinessDashboardExporter {
	return &PrometheusBusinessDashboardExporter{
		logger: logger,
	}
}

// Export exports business dashboard data to Prometheus
func (pbde *PrometheusBusinessDashboardExporter) Export(data *BusinessData) error {
	pbde.logger.Debug("Business dashboard data exported to Prometheus", map[string]interface{}{
		"timestamp":               data.Timestamp,
		"classification_requests": data.ClassificationMetrics.TotalRequests,
		"risk_assessments":        data.RiskAssessmentMetrics.TotalAssessments,
		"compliance_checks":       data.ComplianceMetrics.TotalChecks,
	})

	// In a real implementation, this would export metrics to Prometheus
	return nil
}

// Name returns the exporter name
func (pbde *PrometheusBusinessDashboardExporter) Name() string {
	return "prometheus"
}

// Type returns the exporter type
func (pbde *PrometheusBusinessDashboardExporter) Type() string {
	return "prometheus"
}

// NewBusinessDashboard creates a new business dashboard
func NewBusinessDashboard(
	logger *Logger,
	metricsCollector *MetricsCollector,
	config *BusinessDashboardConfig,
) *BusinessDashboard {
	ctx, cancel := context.WithCancel(context.Background())

	return &BusinessDashboard{
		logger:           logger,
		metricsCollector: metricsCollector,
		config:           config,
		businessData:     make(map[string]*BusinessData),
		exporters:        make([]BusinessDashboardExporter, 0),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start starts the business dashboard
func (bd *BusinessDashboard) Start() error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	if bd.started {
		return fmt.Errorf("business dashboard already started")
	}

	bd.logger.Info("Starting business dashboard", map[string]interface{}{
		"service_name": bd.config.ServiceName,
		"version":      bd.config.Version,
		"environment":  bd.config.Environment,
	})

	// Start data collection
	if bd.config.Enabled {
		go bd.startDataCollection()
	}

	// Start data export
	if bd.config.ExportEnabled {
		go bd.startDataExport()
	}

	bd.started = true
	bd.logger.Info("Business dashboard started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the business dashboard
func (bd *BusinessDashboard) Stop() error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	if !bd.started {
		return fmt.Errorf("business dashboard not started")
	}

	bd.logger.Info("Stopping business dashboard", map[string]interface{}{})

	bd.cancel()
	bd.started = false

	bd.logger.Info("Business dashboard stopped successfully", map[string]interface{}{})
	return nil
}

// GetBusinessData returns current business data
func (bd *BusinessDashboard) GetBusinessData() (*BusinessData, error) {
	businessData := &BusinessData{
		Timestamp:             time.Now(),
		ClassificationMetrics: bd.collectClassificationMetrics(),
		RiskAssessmentMetrics: bd.collectRiskAssessmentMetrics(),
		ComplianceMetrics:     bd.collectComplianceMetrics(),
		UserMetrics:           bd.collectUserMetrics(),
		APIMetrics:            bd.collectAPIMetrics(),
		PerformanceMetrics:    bd.collectPerformanceMetrics(),
		Metadata: map[string]interface{}{
			"service_name": bd.config.ServiceName,
			"version":      bd.config.Version,
			"environment":  bd.config.Environment,
		},
	}

	return businessData, nil
}

// GetBusinessHistory returns historical business data
func (bd *BusinessDashboard) GetBusinessHistory(duration time.Duration) ([]*BusinessData, error) {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	var history []*BusinessData
	cutoff := time.Now().Add(-duration)

	for _, data := range bd.businessData {
		if data.Timestamp.After(cutoff) {
			history = append(history, &BusinessData{
				Timestamp:             data.Timestamp,
				ClassificationMetrics: data.ClassificationMetrics,
				RiskAssessmentMetrics: data.RiskAssessmentMetrics,
				ComplianceMetrics:     data.ComplianceMetrics,
				UserMetrics:           data.UserMetrics,
				APIMetrics:            data.APIMetrics,
				PerformanceMetrics:    data.PerformanceMetrics,
				Metadata:              data.Metadata,
			})
		}
	}

	return history, nil
}

// GetBusinessTrends returns business trends over time
func (bd *BusinessDashboard) GetBusinessTrends(duration time.Duration) (map[string]interface{}, error) {
	history, err := bd.GetBusinessHistory(duration)
	if err != nil {
		return nil, fmt.Errorf("failed to get business history: %w", err)
	}

	if len(history) == 0 {
		return map[string]interface{}{
			"trends": "no_data",
		}, nil
	}

	trends := map[string]interface{}{
		"classification_trend":  bd.calculateClassificationTrend(history),
		"risk_assessment_trend": bd.calculateRiskAssessmentTrend(history),
		"compliance_trend":      bd.calculateComplianceTrend(history),
		"user_activity_trend":   bd.calculateUserActivityTrend(history),
		"api_usage_trend":       bd.calculateAPIUsageTrend(history),
		"performance_trend":     bd.calculatePerformanceTrend(history),
	}

	return trends, nil
}

// GetBusinessSummary returns a business summary
func (bd *BusinessDashboard) GetBusinessSummary() (map[string]interface{}, error) {
	businessData, err := bd.GetBusinessData()
	if err != nil {
		return nil, fmt.Errorf("failed to get business data: %w", err)
	}

	trends, err := bd.GetBusinessTrends(1 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to get business trends: %w", err)
	}

	summary := map[string]interface{}{
		"classification": map[string]interface{}{
			"total_requests":   businessData.ClassificationMetrics.TotalRequests,
			"success_rate":     bd.calculateSuccessRate(businessData.ClassificationMetrics),
			"average_accuracy": businessData.ClassificationMetrics.AverageAccuracy,
			"throughput":       businessData.ClassificationMetrics.Throughput,
		},
		"risk_assessment": map[string]interface{}{
			"total_assessments":  businessData.RiskAssessmentMetrics.TotalAssessments,
			"high_risk_count":    businessData.RiskAssessmentMetrics.HighRiskCount,
			"average_risk_score": businessData.RiskAssessmentMetrics.AverageRiskScore,
			"throughput":         businessData.RiskAssessmentMetrics.Throughput,
		},
		"compliance": map[string]interface{}{
			"total_checks":      businessData.ComplianceMetrics.TotalChecks,
			"compliance_rate":   businessData.ComplianceMetrics.ComplianceRate,
			"critical_failures": businessData.ComplianceMetrics.CriticalFailures,
			"throughput":        businessData.ComplianceMetrics.Throughput,
		},
		"users": map[string]interface{}{
			"total_users":     businessData.UserMetrics.TotalUsers,
			"active_users":    businessData.UserMetrics.ActiveUsers,
			"new_users":       businessData.UserMetrics.NewUsers,
			"conversion_rate": businessData.UserMetrics.ConversionRate,
		},
		"api": map[string]interface{}{
			"total_requests":        businessData.APIMetrics.TotalRequests,
			"success_rate":          bd.calculateAPISuccessRate(businessData.APIMetrics),
			"average_response_time": businessData.APIMetrics.AverageResponseTime,
			"error_rate":            businessData.APIMetrics.ErrorRate,
		},
		"performance": map[string]interface{}{
			"average_response_time": businessData.PerformanceMetrics.AverageResponseTime,
			"availability":          businessData.PerformanceMetrics.Availability,
			"throughput":            businessData.PerformanceMetrics.Throughput,
			"error_rate":            businessData.PerformanceMetrics.ErrorRate,
		},
		"trends":       trends,
		"last_updated": businessData.Timestamp,
		"metadata":     businessData.Metadata,
	}

	return summary, nil
}

// AddExporter adds a business dashboard exporter
func (bd *BusinessDashboard) AddExporter(exporter BusinessDashboardExporter) {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	bd.exporters = append(bd.exporters, exporter)

	bd.logger.Info("Business dashboard exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
		"type":     exporter.Type(),
	})
}

// collectClassificationMetrics collects classification metrics
func (bd *BusinessDashboard) collectClassificationMetrics() *ClassificationMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &ClassificationMetrics{
		TotalRequests:      1000,
		SuccessfulRequests: 950,
		FailedRequests:     50,
		AverageAccuracy:    0.95,
		AverageConfidence:  0.87,
		ProcessingTime:     150 * time.Millisecond,
		RequestsByMethod: map[string]int64{
			"keyword":    600,
			"similarity": 300,
			"pattern":    100,
		},
		RequestsByIndustry: map[string]int64{
			"technology": 400,
			"finance":    300,
			"healthcare": 200,
			"retail":     100,
		},
		ConfidenceDistribution: map[string]int64{
			"high":   700,
			"medium": 200,
			"low":    100,
		},
		AccuracyByMethod: map[string]float64{
			"keyword":    0.96,
			"similarity": 0.94,
			"pattern":    0.92,
		},
		ErrorRate:  0.05,
		Throughput: 16.67, // requests per second
	}
}

// collectRiskAssessmentMetrics collects risk assessment metrics
func (bd *BusinessDashboard) collectRiskAssessmentMetrics() *RiskAssessmentMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &RiskAssessmentMetrics{
		TotalAssessments: 500,
		HighRiskCount:    50,
		MediumRiskCount:  200,
		LowRiskCount:     250,
		AverageRiskScore: 0.35,
		ProcessingTime:   200 * time.Millisecond,
		RiskDistribution: map[string]int64{
			"high":   50,
			"medium": 200,
			"low":    250,
		},
		AssessmentsByType: map[string]int64{
			"aml":       300,
			"kyc":       150,
			"sanctions": 50,
		},
		FalsePositiveRate: 0.02,
		FalseNegativeRate: 0.01,
		Throughput:        8.33, // assessments per second
	}
}

// collectComplianceMetrics collects compliance metrics
func (bd *BusinessDashboard) collectComplianceMetrics() *ComplianceMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &ComplianceMetrics{
		TotalChecks:    800,
		PassedChecks:   760,
		FailedChecks:   40,
		ComplianceRate: 0.95,
		ChecksByFramework: map[string]int64{
			"fatf":    400,
			"pci_dss": 200,
			"sox":     100,
			"gdpr":    100,
		},
		ChecksByStatus: map[string]int64{
			"passed":  760,
			"failed":  30,
			"warning": 10,
		},
		AverageCheckTime: 100 * time.Millisecond,
		CriticalFailures: 5,
		WarningCount:     10,
		Throughput:       13.33, // checks per second
	}
}

// collectUserMetrics collects user metrics
func (bd *BusinessDashboard) collectUserMetrics() *UserMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &UserMetrics{
		TotalUsers:  1000,
		ActiveUsers: 750,
		NewUsers:    50,
		UserActivity: map[string]int64{
			"login":     500,
			"logout":    450,
			"api_call":  2000,
			"page_view": 5000,
		},
		SessionDuration: 25 * time.Minute,
		PageViews:       5000,
		UniqueVisitors:  800,
		BounceRate:      0.15,
		ConversionRate:  0.08,
		UserRetention:   0.85,
	}
}

// collectAPIMetrics collects API metrics
func (bd *BusinessDashboard) collectAPIMetrics() *APIMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &APIMetrics{
		TotalRequests:       5000,
		SuccessfulRequests:  4800,
		FailedRequests:      200,
		AverageResponseTime: 120 * time.Millisecond,
		RequestsByEndpoint: map[string]int64{
			"/api/v3/classify":   2000,
			"/api/v3/risk":       1500,
			"/api/v3/compliance": 1000,
			"/api/v3/health":     500,
		},
		RequestsByMethod: map[string]int64{
			"GET":    3000,
			"POST":   1800,
			"PUT":    150,
			"DELETE": 50,
		},
		RequestsByStatus: map[string]int64{
			"200": 4500,
			"201": 300,
			"400": 100,
			"401": 50,
			"500": 50,
		},
		ErrorRate:              0.04,
		Throughput:             83.33, // requests per second
		RateLimitHits:          25,
		AuthenticationFailures: 50,
	}
}

// collectPerformanceMetrics collects performance metrics
func (bd *BusinessDashboard) collectPerformanceMetrics() *PerformanceMetrics {
	// In a real implementation, this would collect actual metrics
	// For now, return mock data
	return &PerformanceMetrics{
		AverageResponseTime: 120 * time.Millisecond,
		P95ResponseTime:     250 * time.Millisecond,
		P99ResponseTime:     500 * time.Millisecond,
		Throughput:          83.33, // requests per second
		ErrorRate:           0.04,
		Availability:        99.9,
		Uptime:              24 * time.Hour,
		CPUUsage:            45.0,
		MemoryUsage:         60.0,
		DiskUsage:           30.0,
		NetworkLatency:      5 * time.Millisecond,
		DatabaseLatency:     10 * time.Millisecond,
	}
}

// calculateSuccessRate calculates success rate for classification metrics
func (bd *BusinessDashboard) calculateSuccessRate(metrics *ClassificationMetrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0.0
	}
	return float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100.0
}

// calculateAPISuccessRate calculates success rate for API metrics
func (bd *BusinessDashboard) calculateAPISuccessRate(metrics *APIMetrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0.0
	}
	return float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100.0
}

// calculateClassificationTrend calculates classification trend
func (bd *BusinessDashboard) calculateClassificationTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentThroughput := recent.ClassificationMetrics.Throughput
	olderThroughput := older.ClassificationMetrics.Throughput

	diff := recentThroughput - olderThroughput

	if diff > 5 {
		return "increasing"
	} else if diff < -5 {
		return "decreasing"
	}

	return "stable"
}

// calculateRiskAssessmentTrend calculates risk assessment trend
func (bd *BusinessDashboard) calculateRiskAssessmentTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentThroughput := recent.RiskAssessmentMetrics.Throughput
	olderThroughput := older.RiskAssessmentMetrics.Throughput

	diff := recentThroughput - olderThroughput

	if diff > 2 {
		return "increasing"
	} else if diff < -2 {
		return "decreasing"
	}

	return "stable"
}

// calculateComplianceTrend calculates compliance trend
func (bd *BusinessDashboard) calculateComplianceTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentRate := recent.ComplianceMetrics.ComplianceRate
	olderRate := older.ComplianceMetrics.ComplianceRate

	diff := recentRate - olderRate

	if diff > 0.05 {
		return "improving"
	} else if diff < -0.05 {
		return "degrading"
	}

	return "stable"
}

// calculateUserActivityTrend calculates user activity trend
func (bd *BusinessDashboard) calculateUserActivityTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentActive := recent.UserMetrics.ActiveUsers
	olderActive := older.UserMetrics.ActiveUsers

	diff := recentActive - olderActive

	if diff > 50 {
		return "increasing"
	} else if diff < -50 {
		return "decreasing"
	}

	return "stable"
}

// calculateAPIUsageTrend calculates API usage trend
func (bd *BusinessDashboard) calculateAPIUsageTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentThroughput := recent.APIMetrics.Throughput
	olderThroughput := older.APIMetrics.Throughput

	diff := recentThroughput - olderThroughput

	if diff > 10 {
		return "increasing"
	} else if diff < -10 {
		return "decreasing"
	}

	return "stable"
}

// calculatePerformanceTrend calculates performance trend
func (bd *BusinessDashboard) calculatePerformanceTrend(history []*BusinessData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentResponseTime := recent.PerformanceMetrics.AverageResponseTime
	olderResponseTime := older.PerformanceMetrics.AverageResponseTime

	diff := recentResponseTime - olderResponseTime

	if diff > 50*time.Millisecond {
		return "slower"
	} else if diff < -50*time.Millisecond {
		return "faster"
	}

	return "stable"
}

// startDataCollection starts the data collection process
func (bd *BusinessDashboard) startDataCollection() {
	ticker := time.NewTicker(bd.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bd.ctx.Done():
			bd.logger.Info("Business data collection stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			bd.collectBusinessData()
		}
	}
}

// collectBusinessData collects current business data
func (bd *BusinessDashboard) collectBusinessData() {
	businessData, err := bd.GetBusinessData()
	if err != nil {
		bd.logger.Error("Failed to collect business data", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Store the data
	bd.mu.Lock()
	key := businessData.Timestamp.Format("2006-01-02T15:04:05")
	bd.businessData[key] = businessData

	// Clean up old data
	bd.cleanupOldData()

	bd.mu.Unlock()

	bd.logger.Debug("Business data collected", map[string]interface{}{
		"classification_requests": businessData.ClassificationMetrics.TotalRequests,
		"risk_assessments":        businessData.RiskAssessmentMetrics.TotalAssessments,
		"compliance_checks":       businessData.ComplianceMetrics.TotalChecks,
	})
}

// cleanupOldData removes old business data
func (bd *BusinessDashboard) cleanupOldData() {
	cutoff := time.Now().Add(-bd.config.DataRetentionPeriod)

	for key, data := range bd.businessData {
		if data.Timestamp.Before(cutoff) {
			delete(bd.businessData, key)
		}
	}

	// Limit the number of data points
	if len(bd.businessData) > bd.config.MaxDataPoints {
		// Remove oldest entries
		count := 0
		for key := range bd.businessData {
			if count >= len(bd.businessData)-bd.config.MaxDataPoints {
				break
			}
			delete(bd.businessData, key)
			count++
		}
	}
}

// startDataExport starts the data export process
func (bd *BusinessDashboard) startDataExport() {
	ticker := time.NewTicker(bd.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bd.ctx.Done():
			bd.logger.Info("Business data export stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			bd.exportBusinessData()
		}
	}
}

// exportBusinessData exports current business data
func (bd *BusinessDashboard) exportBusinessData() {
	businessData, err := bd.GetBusinessData()
	if err != nil {
		bd.logger.Error("Failed to get business data for export", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, exporter := range bd.exporters {
		if err := exporter.Export(businessData); err != nil {
			bd.logger.Error("Failed to export business data", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
		}
	}
}
