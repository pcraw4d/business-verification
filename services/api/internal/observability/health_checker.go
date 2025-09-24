package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthChecker handles health check monitoring
type HealthChecker struct {
	logger    *Logger
	checks    map[string]*HealthCheck
	mu        sync.RWMutex
	config    *HealthCheckConfig
	exporters []HealthExporter
}

// HealthCheck represents a health check
type HealthCheck struct {
	Name       string
	Status     HealthStatus
	Message    string
	LastCheck  time.Time
	Duration   time.Duration
	Tags       map[string]string
	CheckFunc  func() error
	Interval   time.Duration
	Timeout    time.Duration
	Critical   bool
	RetryCount int
	MaxRetries int
}

// HealthStatus represents the status of a health check
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckConfig holds configuration for health checks
type HealthCheckConfig struct {
	Enabled        bool
	CheckInterval  time.Duration
	Timeout        time.Duration
	RetryCount     int
	RetryInterval  time.Duration
	AlertOnFailure bool
	AlertChannels  []string
}

// HealthExporter interface for exporting health data
type HealthExporter interface {
	Export(checks map[string]*HealthCheck) error
	Name() string
}

// LogHealthExporter exports health data to logs
type LogHealthExporter struct {
	logger *Logger
}

// NewLogHealthExporter creates a new log health exporter
func NewLogHealthExporter(logger *Logger) *LogHealthExporter {
	return &LogHealthExporter{
		logger: logger,
	}
}

// Export exports health data to logs
func (lhe *LogHealthExporter) Export(checks map[string]*HealthCheck) error {
	lhe.logger.Info("Health check status", map[string]interface{}{
		"total_checks": len(checks),
		"checks":       checks,
	})
	return nil
}

// Name returns the exporter name
func (lhe *LogHealthExporter) Name() string {
	return "log"
}

// PrometheusHealthExporter exports health data to Prometheus
type PrometheusHealthExporter struct {
	logger *Logger
}

// NewPrometheusHealthExporter creates a new Prometheus health exporter
func NewPrometheusHealthExporter(logger *Logger) *PrometheusHealthExporter {
	return &PrometheusHealthExporter{
		logger: logger,
	}
}

// Export exports health data to Prometheus
func (phe *PrometheusHealthExporter) Export(checks map[string]*HealthCheck) error {
	// In a real implementation, this would export to Prometheus
	phe.logger.Debug("Exporting health data to Prometheus", map[string]interface{}{
		"check_count": len(checks),
	})
	return nil
}

// Name returns the exporter name
func (phe *PrometheusHealthExporter) Name() string {
	return "prometheus"
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *Logger, config *HealthCheckConfig) *HealthChecker {
	return &HealthChecker{
		logger:    logger,
		checks:    make(map[string]*HealthCheck),
		exporters: make([]HealthExporter, 0),
		config:    config,
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(name string, checkFunc func() error, interval time.Duration, critical bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	check := &HealthCheck{
		Name:       name,
		Status:     HealthStatusUnknown,
		Message:    "Not checked yet",
		LastCheck:  time.Time{},
		Duration:   0,
		Tags:       make(map[string]string),
		CheckFunc:  checkFunc,
		Interval:   interval,
		Timeout:    hc.config.Timeout,
		Critical:   critical,
		RetryCount: 0,
		MaxRetries: hc.config.RetryCount,
	}

	hc.checks[name] = check
	hc.logger.Info("Health check added", map[string]interface{}{
		"name":     name,
		"interval": interval.String(),
		"critical": critical,
	})
}

// RemoveCheck removes a health check
func (hc *HealthChecker) RemoveCheck(name string) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if _, exists := hc.checks[name]; !exists {
		return fmt.Errorf("health check %s not found", name)
	}

	delete(hc.checks, name)
	hc.logger.Info("Health check removed", map[string]interface{}{
		"name": name,
	})
	return nil
}

// RunCheck runs a specific health check
func (hc *HealthChecker) RunCheck(name string) error {
	hc.mu.RLock()
	check, exists := hc.checks[name]
	hc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("health check %s not found", name)
	}

	return hc.runSingleCheck(check)
}

// RunChecks runs all health checks
func (hc *HealthChecker) RunChecks() {
	hc.mu.RLock()
	checks := make([]*HealthCheck, 0, len(hc.checks))
	for _, check := range hc.checks {
		checks = append(checks, check)
	}
	hc.mu.RUnlock()

	for _, check := range checks {
		_ = hc.runSingleCheck(check)
	}

	// Export health data
	hc.exportHealthData()
}

// runSingleCheck runs a single health check
func (hc *HealthChecker) runSingleCheck(check *HealthCheck) error {
	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), check.Timeout)
	defer cancel()

	// Run the check function
	err := hc.runCheckWithContext(ctx, check)

	duration := time.Since(start)
	check.Duration = duration
	check.LastCheck = time.Now()

	if err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = err.Error()
		check.RetryCount++

		hc.logger.Warn("Health check failed", map[string]interface{}{
			"name":        check.Name,
			"error":       err.Error(),
			"duration":    duration.String(),
			"retry_count": check.RetryCount,
			"critical":    check.Critical,
		})

		// Send alert if critical and retries exceeded
		if check.Critical && check.RetryCount >= check.MaxRetries {
			hc.sendAlert(check, err)
		}
	} else {
		check.Status = HealthStatusHealthy
		check.Message = "OK"
		check.RetryCount = 0

		hc.logger.Debug("Health check passed", map[string]interface{}{
			"name":     check.Name,
			"duration": duration.String(),
		})
	}
	return nil
}

// runCheckWithContext runs a health check with context
func (hc *HealthChecker) runCheckWithContext(ctx context.Context, check *HealthCheck) error {
	done := make(chan error, 1)

	go func() {
		done <- check.CheckFunc()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("health check timeout after %v", check.Timeout)
	}
}

// GetStatus returns the overall health status
func (hc *HealthChecker) GetStatus() map[string]interface{} {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status := map[string]interface{}{
		"overall_status":  HealthStatusHealthy,
		"total_checks":    len(hc.checks),
		"healthy":         0,
		"unhealthy":       0,
		"degraded":        0,
		"unknown":         0,
		"critical_failed": 0,
		"checks":          make(map[string]interface{}),
	}

	hasUnhealthy := false
	hasCriticalFailed := false

	for name, check := range hc.checks {
		checkStatus := map[string]interface{}{
			"status":      check.Status,
			"message":     check.Message,
			"last_check":  check.LastCheck,
			"duration":    check.Duration.String(),
			"critical":    check.Critical,
			"retry_count": check.RetryCount,
			"tags":        check.Tags,
		}

		status["checks"].(map[string]interface{})[name] = checkStatus

		// Count by status
		switch check.Status {
		case HealthStatusHealthy:
			status["healthy"] = status["healthy"].(int) + 1
		case HealthStatusUnhealthy:
			status["unhealthy"] = status["unhealthy"].(int) + 1
			hasUnhealthy = true
			if check.Critical {
				status["critical_failed"] = status["critical_failed"].(int) + 1
				hasCriticalFailed = true
			}
		case HealthStatusDegraded:
			status["degraded"] = status["degraded"].(int) + 1
			hasUnhealthy = true
		case HealthStatusUnknown:
			status["unknown"] = status["unknown"].(int) + 1
		}
	}

	// Determine overall status
	if hasCriticalFailed {
		status["overall_status"] = HealthStatusUnhealthy
	} else if hasUnhealthy {
		status["overall_status"] = HealthStatusDegraded
	}

	return status
}

// GetCheckStatus returns the status of a specific health check
func (hc *HealthChecker) GetCheckStatus(name string) (*HealthCheck, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	check, exists := hc.checks[name]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &HealthCheck{
		Name:       check.Name,
		Status:     check.Status,
		Message:    check.Message,
		LastCheck:  check.LastCheck,
		Duration:   check.Duration,
		Tags:       check.Tags,
		CheckFunc:  check.CheckFunc,
		Interval:   check.Interval,
		Timeout:    check.Timeout,
		Critical:   check.Critical,
		RetryCount: check.RetryCount,
		MaxRetries: check.MaxRetries,
	}, true
}

// GetUnhealthyChecks returns all unhealthy health checks
func (hc *HealthChecker) GetUnhealthyChecks() []*HealthCheck {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var unhealthy []*HealthCheck
	for _, check := range hc.checks {
		if check.Status == HealthStatusUnhealthy || check.Status == HealthStatusDegraded {
			unhealthy = append(unhealthy, &HealthCheck{
				Name:       check.Name,
				Status:     check.Status,
				Message:    check.Message,
				LastCheck:  check.LastCheck,
				Duration:   check.Duration,
				Tags:       check.Tags,
				CheckFunc:  check.CheckFunc,
				Interval:   check.Interval,
				Timeout:    check.Timeout,
				Critical:   check.Critical,
				RetryCount: check.RetryCount,
				MaxRetries: check.MaxRetries,
			})
		}
	}
	return unhealthy
}

// GetCriticalChecks returns all critical health checks
func (hc *HealthChecker) GetCriticalChecks() []*HealthCheck {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var critical []*HealthCheck
	for _, check := range hc.checks {
		if check.Critical {
			critical = append(critical, &HealthCheck{
				Name:       check.Name,
				Status:     check.Status,
				Message:    check.Message,
				LastCheck:  check.LastCheck,
				Duration:   check.Duration,
				Tags:       check.Tags,
				CheckFunc:  check.CheckFunc,
				Interval:   check.Interval,
				Timeout:    check.Timeout,
				Critical:   check.Critical,
				RetryCount: check.RetryCount,
				MaxRetries: check.MaxRetries,
			})
		}
	}
	return critical
}

// AddExporter adds a health exporter
func (hc *HealthChecker) AddExporter(exporter HealthExporter) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.exporters = append(hc.exporters, exporter)
	hc.logger.Info("Health exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
	})
}

// exportHealthData exports health data using registered exporters
func (hc *HealthChecker) exportHealthData() {
	hc.mu.RLock()
	checks := make(map[string]*HealthCheck)
	for name, check := range hc.checks {
		checks[name] = &HealthCheck{
			Name:       check.Name,
			Status:     check.Status,
			Message:    check.Message,
			LastCheck:  check.LastCheck,
			Duration:   check.Duration,
			Tags:       check.Tags,
			CheckFunc:  check.CheckFunc,
			Interval:   check.Interval,
			Timeout:    check.Timeout,
			Critical:   check.Critical,
			RetryCount: check.RetryCount,
			MaxRetries: check.MaxRetries,
		}
	}
	hc.mu.RUnlock()

	for _, exporter := range hc.exporters {
		if err := exporter.Export(checks); err != nil {
			hc.logger.Error("Failed to export health data", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
		}
	}
}

// sendAlert sends an alert for a failed health check
func (hc *HealthChecker) sendAlert(check *HealthCheck, err error) {
	if !hc.config.AlertOnFailure {
		return
	}

	hc.logger.Error("Health check alert", map[string]interface{}{
		"name":        check.Name,
		"status":      check.Status,
		"message":     check.Message,
		"error":       err.Error(),
		"critical":    check.Critical,
		"retry_count": check.RetryCount,
		"channels":    hc.config.AlertChannels,
	})

	// In a real implementation, this would send alerts via configured channels
}

// StartPeriodicChecks starts periodic health checks
func (hc *HealthChecker) StartPeriodicChecks(ctx context.Context) {
	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hc.logger.Info("Periodic health checks stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			hc.RunChecks()
		}
	}
}

// Default health check functions

// DatabaseHealthCheck creates a database health check
func DatabaseHealthCheck(db interface{}) func() error {
	return func() error {
		// In a real implementation, this would check database connectivity
		// For now, return nil (healthy)
		return nil
	}
}

// RedisHealthCheck creates a Redis health check
func RedisHealthCheck(redis interface{}) func() error {
	return func() error {
		// In a real implementation, this would check Redis connectivity
		// For now, return nil (healthy)
		return nil
	}
}

// ExternalAPIHealthCheck creates an external API health check
func ExternalAPIHealthCheck(apiURL string) func() error {
	return func() error {
		// In a real implementation, this would check external API availability
		// For now, return nil (healthy)
		return nil
	}
}

// DiskSpaceHealthCheck creates a disk space health check
func DiskSpaceHealthCheck(path string, minFreeBytes int64) func() error {
	return func() error {
		// In a real implementation, this would check disk space
		// For now, return nil (healthy)
		return nil
	}
}

// MemoryHealthCheck creates a memory health check
func MemoryHealthCheck(maxUsagePercent float64) func() error {
	return func() error {
		// In a real implementation, this would check memory usage
		// For now, return nil (healthy)
		return nil
	}
}
