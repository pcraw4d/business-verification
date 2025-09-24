package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthDashboard provides comprehensive health monitoring dashboard functionality
type HealthDashboard struct {
	logger        *Logger
	healthChecker *HealthChecker
	config        *HealthDashboardConfig
	healthData    map[string]*HealthData
	exporters     []HealthDashboardExporter
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	started       bool
}

// HealthDashboardConfig holds configuration for health dashboard
type HealthDashboardConfig struct {
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

// HealthData represents health dashboard data
type HealthData struct {
	Timestamp      time.Time                   `json:"timestamp"`
	OverallStatus  HealthStatus                `json:"overall_status"`
	TotalChecks    int                         `json:"total_checks"`
	HealthyCount   int                         `json:"healthy_count"`
	UnhealthyCount int                         `json:"unhealthy_count"`
	DegradedCount  int                         `json:"degraded_count"`
	UnknownCount   int                         `json:"unknown_count"`
	CriticalFailed int                         `json:"critical_failed"`
	Checks         map[string]*HealthCheckData `json:"checks"`
	Metrics        map[string]interface{}      `json:"metrics"`
	Metadata       map[string]interface{}      `json:"metadata"`
}

// HealthCheckData represents individual health check data
type HealthCheckData struct {
	Name       string                 `json:"name"`
	Status     HealthStatus           `json:"status"`
	Message    string                 `json:"message"`
	LastCheck  time.Time              `json:"last_check"`
	Duration   time.Duration          `json:"duration"`
	Tags       map[string]string      `json:"tags"`
	Critical   bool                   `json:"critical"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// HealthDashboardExporter interface for exporting health dashboard data
type HealthDashboardExporter interface {
	Export(data *HealthData) error
	Name() string
	Type() string
}

// JSONHealthDashboardExporter exports health dashboard data as JSON
type JSONHealthDashboardExporter struct {
	logger *Logger
}

// NewJSONHealthDashboardExporter creates a new JSON health dashboard exporter
func NewJSONHealthDashboardExporter(logger *Logger) *JSONHealthDashboardExporter {
	return &JSONHealthDashboardExporter{
		logger: logger,
	}
}

// Export exports health dashboard data as JSON
func (jhde *JSONHealthDashboardExporter) Export(data *HealthData) error {
	jhde.logger.Debug("Health dashboard data exported as JSON", map[string]interface{}{
		"overall_status": data.OverallStatus,
		"total_checks":   data.TotalChecks,
		"healthy_count":  data.HealthyCount,
		"timestamp":      data.Timestamp,
	})

	return nil
}

// Name returns the exporter name
func (jhde *JSONHealthDashboardExporter) Name() string {
	return "json"
}

// Type returns the exporter type
func (jhde *JSONHealthDashboardExporter) Type() string {
	return "json"
}

// PrometheusHealthDashboardExporter exports health dashboard data to Prometheus
type PrometheusHealthDashboardExporter struct {
	logger *Logger
}

// NewPrometheusHealthDashboardExporter creates a new Prometheus health dashboard exporter
func NewPrometheusHealthDashboardExporter(logger *Logger) *PrometheusHealthDashboardExporter {
	return &PrometheusHealthDashboardExporter{
		logger: logger,
	}
}

// Export exports health dashboard data to Prometheus
func (phde *PrometheusHealthDashboardExporter) Export(data *HealthData) error {
	phde.logger.Debug("Health dashboard data exported to Prometheus", map[string]interface{}{
		"overall_status": data.OverallStatus,
		"total_checks":   data.TotalChecks,
		"timestamp":      data.Timestamp,
	})

	// In a real implementation, this would export metrics to Prometheus
	return nil
}

// Name returns the exporter name
func (phde *PrometheusHealthDashboardExporter) Name() string {
	return "prometheus"
}

// Type returns the exporter type
func (phde *PrometheusHealthDashboardExporter) Type() string {
	return "prometheus"
}

// NewHealthDashboard creates a new health dashboard
func NewHealthDashboard(
	logger *Logger,
	healthChecker *HealthChecker,
	config *HealthDashboardConfig,
) *HealthDashboard {
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthDashboard{
		logger:        logger,
		healthChecker: healthChecker,
		config:        config,
		healthData:    make(map[string]*HealthData),
		exporters:     make([]HealthDashboardExporter, 0),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the health dashboard
func (hd *HealthDashboard) Start() error {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	if hd.started {
		return fmt.Errorf("health dashboard already started")
	}

	hd.logger.Info("Starting health dashboard", map[string]interface{}{
		"service_name": hd.config.ServiceName,
		"version":      hd.config.Version,
		"environment":  hd.config.Environment,
	})

	// Start data collection
	if hd.config.Enabled {
		go hd.startDataCollection()
	}

	// Start data export
	if hd.config.ExportEnabled {
		go hd.startDataExport()
	}

	hd.started = true
	hd.logger.Info("Health dashboard started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the health dashboard
func (hd *HealthDashboard) Stop() error {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	if !hd.started {
		return fmt.Errorf("health dashboard not started")
	}

	hd.logger.Info("Stopping health dashboard", map[string]interface{}{})

	hd.cancel()
	hd.started = false

	hd.logger.Info("Health dashboard stopped successfully", map[string]interface{}{})
	return nil
}

// GetHealthData returns current health data
func (hd *HealthDashboard) GetHealthData() (*HealthData, error) {
	if hd.healthChecker == nil {
		return nil, fmt.Errorf("health checker not available")
	}

	status := hd.healthChecker.GetStatus()
	checks := status["checks"].(map[string]interface{})

	healthData := &HealthData{
		Timestamp:      time.Now(),
		OverallStatus:  HealthStatus(status["overall_status"].(string)),
		TotalChecks:    status["total_checks"].(int),
		HealthyCount:   status["healthy"].(int),
		UnhealthyCount: status["unhealthy"].(int),
		DegradedCount:  status["degraded"].(int),
		UnknownCount:   status["unknown"].(int),
		CriticalFailed: status["critical_failed"].(int),
		Checks:         make(map[string]*HealthCheckData),
		Metrics: map[string]interface{}{
			"uptime":                hd.calculateUptime(),
			"average_response_time": hd.calculateAverageResponseTime(),
			"error_rate":            hd.calculateErrorRate(),
		},
		Metadata: map[string]interface{}{
			"service_name": hd.config.ServiceName,
			"version":      hd.config.Version,
			"environment":  hd.config.Environment,
		},
	}

	// Convert checks data
	for name, checkData := range checks {
		check := checkData.(map[string]interface{})
		healthData.Checks[name] = &HealthCheckData{
			Name:       name,
			Status:     HealthStatus(check["status"].(string)),
			Message:    check["message"].(string),
			LastCheck:  check["last_check"].(time.Time),
			Duration:   check["duration"].(time.Duration),
			Tags:       check["tags"].(map[string]string),
			Critical:   check["critical"].(bool),
			RetryCount: check["retry_count"].(int),
			MaxRetries: check["max_retries"].(int),
			Metadata:   make(map[string]interface{}),
		}
	}

	return healthData, nil
}

// GetHealthHistory returns historical health data
func (hd *HealthDashboard) GetHealthHistory(duration time.Duration) ([]*HealthData, error) {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	var history []*HealthData
	cutoff := time.Now().Add(-duration)

	for _, data := range hd.healthData {
		if data.Timestamp.After(cutoff) {
			history = append(history, &HealthData{
				Timestamp:      data.Timestamp,
				OverallStatus:  data.OverallStatus,
				TotalChecks:    data.TotalChecks,
				HealthyCount:   data.HealthyCount,
				UnhealthyCount: data.UnhealthyCount,
				DegradedCount:  data.DegradedCount,
				UnknownCount:   data.UnknownCount,
				CriticalFailed: data.CriticalFailed,
				Checks:         data.Checks,
				Metrics:        data.Metrics,
				Metadata:       data.Metadata,
			})
		}
	}

	return history, nil
}

// GetHealthTrends returns health trends over time
func (hd *HealthDashboard) GetHealthTrends(duration time.Duration) (map[string]interface{}, error) {
	history, err := hd.GetHealthHistory(duration)
	if err != nil {
		return nil, fmt.Errorf("failed to get health history: %w", err)
	}

	if len(history) == 0 {
		return map[string]interface{}{
			"trends": "no_data",
		}, nil
	}

	trends := map[string]interface{}{
		"overall_status_trend": hd.calculateStatusTrend(history),
		"availability_trend":   hd.calculateAvailabilityTrend(history),
		"response_time_trend":  hd.calculateResponseTimeTrend(history),
		"error_rate_trend":     hd.calculateErrorRateTrend(history),
		"critical_failures":    hd.calculateCriticalFailures(history),
		"recovery_time":        hd.calculateRecoveryTime(history),
	}

	return trends, nil
}

// GetHealthSummary returns a health summary
func (hd *HealthDashboard) GetHealthSummary() (map[string]interface{}, error) {
	healthData, err := hd.GetHealthData()
	if err != nil {
		return nil, fmt.Errorf("failed to get health data: %w", err)
	}

	trends, err := hd.GetHealthTrends(1 * time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to get health trends: %w", err)
	}

	summary := map[string]interface{}{
		"current_status":  healthData.OverallStatus,
		"total_checks":    healthData.TotalChecks,
		"healthy_count":   healthData.HealthyCount,
		"unhealthy_count": healthData.UnhealthyCount,
		"degraded_count":  healthData.DegradedCount,
		"critical_failed": healthData.CriticalFailed,
		"availability":    hd.calculateAvailability(healthData),
		"uptime":          healthData.Metrics["uptime"],
		"trends":          trends,
		"last_updated":    healthData.Timestamp,
		"metadata":        healthData.Metadata,
	}

	return summary, nil
}

// AddExporter adds a health dashboard exporter
func (hd *HealthDashboard) AddExporter(exporter HealthDashboardExporter) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.exporters = append(hd.exporters, exporter)

	hd.logger.Info("Health dashboard exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
		"type":     exporter.Type(),
	})
}

// startDataCollection starts the data collection process
func (hd *HealthDashboard) startDataCollection() {
	ticker := time.NewTicker(hd.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hd.ctx.Done():
			hd.logger.Info("Health data collection stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			hd.collectHealthData()
		}
	}
}

// collectHealthData collects current health data
func (hd *HealthDashboard) collectHealthData() {
	healthData, err := hd.GetHealthData()
	if err != nil {
		hd.logger.Error("Failed to collect health data", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Store the data
	hd.mu.Lock()
	key := healthData.Timestamp.Format("2006-01-02T15:04:05")
	hd.healthData[key] = healthData

	// Clean up old data
	hd.cleanupOldData()

	hd.mu.Unlock()

	hd.logger.Debug("Health data collected", map[string]interface{}{
		"overall_status": healthData.OverallStatus,
		"total_checks":   healthData.TotalChecks,
		"healthy_count":  healthData.HealthyCount,
	})
}

// cleanupOldData removes old health data
func (hd *HealthDashboard) cleanupOldData() {
	cutoff := time.Now().Add(-hd.config.DataRetentionPeriod)

	for key, data := range hd.healthData {
		if data.Timestamp.Before(cutoff) {
			delete(hd.healthData, key)
		}
	}

	// Limit the number of data points
	if len(hd.healthData) > hd.config.MaxDataPoints {
		// Remove oldest entries
		count := 0
		for key := range hd.healthData {
			if count >= len(hd.healthData)-hd.config.MaxDataPoints {
				break
			}
			delete(hd.healthData, key)
			count++
		}
	}
}

// startDataExport starts the data export process
func (hd *HealthDashboard) startDataExport() {
	ticker := time.NewTicker(hd.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hd.ctx.Done():
			hd.logger.Info("Health data export stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			hd.exportHealthData()
		}
	}
}

// exportHealthData exports current health data
func (hd *HealthDashboard) exportHealthData() {
	healthData, err := hd.GetHealthData()
	if err != nil {
		hd.logger.Error("Failed to get health data for export", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for _, exporter := range hd.exporters {
		if err := exporter.Export(healthData); err != nil {
			hd.logger.Error("Failed to export health data", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
		}
	}
}

// calculateUptime calculates system uptime
func (hd *HealthDashboard) calculateUptime() time.Duration {
	// In a real implementation, this would calculate actual uptime
	// For now, return a mock value
	return 24 * time.Hour
}

// calculateAverageResponseTime calculates average response time
func (hd *HealthDashboard) calculateAverageResponseTime() time.Duration {
	// In a real implementation, this would calculate from metrics
	// For now, return a mock value
	return 150 * time.Millisecond
}

// calculateErrorRate calculates error rate
func (hd *HealthDashboard) calculateErrorRate() float64 {
	// In a real implementation, this would calculate from metrics
	// For now, return a mock value
	return 0.5
}

// calculateAvailability calculates system availability
func (hd *HealthDashboard) calculateAvailability(healthData *HealthData) float64 {
	if healthData.TotalChecks == 0 {
		return 0.0
	}

	healthyCount := healthData.HealthyCount + healthData.DegradedCount
	return float64(healthyCount) / float64(healthData.TotalChecks) * 100.0
}

// calculateStatusTrend calculates status trend over time
func (hd *HealthDashboard) calculateStatusTrend(history []*HealthData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	if recent.OverallStatus == older.OverallStatus {
		return "stable"
	}

	if recent.OverallStatus == HealthStatusHealthy && older.OverallStatus != HealthStatusHealthy {
		return "improving"
	}

	if recent.OverallStatus != HealthStatusHealthy && older.OverallStatus == HealthStatusHealthy {
		return "degrading"
	}

	return "fluctuating"
}

// calculateAvailabilityTrend calculates availability trend
func (hd *HealthDashboard) calculateAvailabilityTrend(history []*HealthData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentAvailability := hd.calculateAvailability(recent)
	olderAvailability := hd.calculateAvailability(older)

	diff := recentAvailability - olderAvailability

	if diff > 5 {
		return "improving"
	} else if diff < -5 {
		return "degrading"
	}

	return "stable"
}

// calculateResponseTimeTrend calculates response time trend
func (hd *HealthDashboard) calculateResponseTimeTrend(history []*HealthData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentResponseTime := recent.Metrics["average_response_time"].(time.Duration)
	olderResponseTime := older.Metrics["average_response_time"].(time.Duration)

	diff := recentResponseTime - olderResponseTime

	if diff > 50*time.Millisecond {
		return "slower"
	} else if diff < -50*time.Millisecond {
		return "faster"
	}

	return "stable"
}

// calculateErrorRateTrend calculates error rate trend
func (hd *HealthDashboard) calculateErrorRateTrend(history []*HealthData) string {
	if len(history) < 2 {
		return "stable"
	}

	recent := history[len(history)-1]
	older := history[0]

	recentErrorRate := recent.Metrics["error_rate"].(float64)
	olderErrorRate := older.Metrics["error_rate"].(float64)

	diff := recentErrorRate - olderErrorRate

	if diff > 1.0 {
		return "increasing"
	} else if diff < -1.0 {
		return "decreasing"
	}

	return "stable"
}

// calculateCriticalFailures calculates critical failures over time
func (hd *HealthDashboard) calculateCriticalFailures(history []*HealthData) int {
	totalFailures := 0
	for _, data := range history {
		totalFailures += data.CriticalFailed
	}
	return totalFailures
}

// calculateRecoveryTime calculates average recovery time
func (hd *HealthDashboard) calculateRecoveryTime(history []*HealthData) time.Duration {
	// In a real implementation, this would calculate actual recovery times
	// For now, return a mock value
	return 2 * time.Minute
}
