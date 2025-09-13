package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/company/kyb-platform/internal/config"
)

// MonitoringIntegrationService integrates all monitoring components
type MonitoringIntegrationService struct {
	logger             *zap.Logger
	config             *config.AlertingConfig
	healthCheckService *HealthCheckService
	alertingService    *EnhancedAlertingService
	metricsCollector   MetricsCollector
	performanceMonitor PerformanceMonitor
	mu                 sync.RWMutex
	started            bool
	stopCh             chan struct{}
	metricsTicker      *time.Ticker
	healthCheckTicker  *time.Ticker
	alertingTicker     *time.Ticker
}

// NewMonitoringIntegrationService creates a new monitoring integration service
func NewMonitoringIntegrationService(
	logger *zap.Logger,
	config *config.AlertingConfig,
	healthCheckService *HealthCheckService,
	alertingService *EnhancedAlertingService,
	metricsCollector MetricsCollector,
	performanceMonitor PerformanceMonitor,
) *MonitoringIntegrationService {
	return &MonitoringIntegrationService{
		logger:             logger,
		config:             config,
		healthCheckService: healthCheckService,
		alertingService:    alertingService,
		metricsCollector:   metricsCollector,
		performanceMonitor: performanceMonitor,
		stopCh:             make(chan struct{}),
	}
}

// Start starts the monitoring integration service
func (m *MonitoringIntegrationService) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("monitoring integration service already started")
	}

	m.logger.Info("Starting monitoring integration service")

	// Start metrics collection
	if m.metricsCollector != nil {
		m.metricsTicker = time.NewTicker(30 * time.Second)
		go m.metricsCollectionLoop()
	}

	// Start health checks
	if m.healthCheckService != nil {
		m.healthCheckTicker = time.NewTicker(30 * time.Second)
		go m.healthCheckLoop()
	}

	// Start alerting
	if m.alertingService != nil {
		m.alertingTicker = time.NewTicker(30 * time.Second)
		go m.alertingLoop()
	}

	m.started = true
	m.logger.Info("Monitoring integration service started successfully")
	return nil
}

// Stop stops the monitoring integration service
func (m *MonitoringIntegrationService) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.logger.Info("Stopping monitoring integration service")

	// Stop tickers
	if m.metricsTicker != nil {
		m.metricsTicker.Stop()
	}
	if m.healthCheckTicker != nil {
		m.healthCheckTicker.Stop()
	}
	if m.alertingTicker != nil {
		m.alertingTicker.Stop()
	}

	// Signal stop
	close(m.stopCh)

	m.started = false
	m.logger.Info("Monitoring integration service stopped")
	return nil
}

// GetSystemStatus returns the overall system status
func (m *MonitoringIntegrationService) GetSystemStatus() SystemStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := SystemStatus{
		Timestamp: time.Now(),
		Overall:   "healthy",
		Services:  make(map[string]ServiceStatus),
	}

	// Check health of all services
	if m.healthCheckService != nil {
		healthChecks := m.healthCheckService.GetAllHealthChecks()
		hasCritical := false
		hasWarning := false

		for _, check := range healthChecks {
			serviceStatus := ServiceStatus{
				Name:      check.Name,
				Status:    string(check.Status),
				Message:   check.Message,
				LastCheck: check.LastChecked,
				Duration:  check.Duration,
			}

			status.Services[check.Name] = serviceStatus

			switch check.Status {
			case HealthStatusCritical:
				hasCritical = true
			case HealthStatusWarning:
				hasWarning = true
			}
		}

		if hasCritical {
			status.Overall = "critical"
		} else if hasWarning {
			status.Overall = "warning"
		}
	}

	// Add metrics summary
	if m.metricsCollector != nil {
		status.Metrics = MetricsSummary{
			RequestRate:  m.metricsCollector.GetRequestRate(),
			ResponseTime: m.metricsCollector.GetResponseTime(),
			ErrorRate:    m.metricsCollector.GetErrorRate(),
			ActiveUsers:  m.metricsCollector.GetActiveUsers(),
			MemoryUsage:  m.metricsCollector.GetMemoryUsage(),
			CPUUsage:     m.metricsCollector.GetCPUUsage(),
		}
	}

	// Add active alerts
	if m.alertingService != nil {
		activeAlerts := m.alertingService.GetActiveAlerts()
		status.ActiveAlerts = len(activeAlerts)
		status.CriticalAlerts = 0
		status.WarningAlerts = 0

		for _, alert := range activeAlerts {
			switch alert.Severity {
			case "critical":
				status.CriticalAlerts++
			case "warning":
				status.WarningAlerts++
			}
		}
	}

	return status
}

// GetMetricsSummary returns a summary of current metrics
func (m *MonitoringIntegrationService) GetMetricsSummary() MetricsSummary {
	if m.metricsCollector == nil {
		return MetricsSummary{}
	}

	return MetricsSummary{
		RequestRate:  m.metricsCollector.GetRequestRate(),
		ResponseTime: m.metricsCollector.GetResponseTime(),
		ErrorRate:    m.metricsCollector.GetErrorRate(),
		ActiveUsers:  m.metricsCollector.GetActiveUsers(),
		MemoryUsage:  m.metricsCollector.GetMemoryUsage(),
		CPUUsage:     m.metricsCollector.GetCPUUsage(),
	}
}

// GetHealthSummary returns a summary of health checks
func (m *MonitoringIntegrationService) GetHealthSummary() HealthSummary {
	if m.healthCheckService == nil {
		return HealthSummary{}
	}

	checks := m.healthCheckService.GetAllHealthChecks()
	summary := HealthSummary{
		Total:    len(checks),
		Healthy:  0,
		Warning:  0,
		Critical: 0,
		Checks:   make([]HealthCheckSummary, len(checks)),
	}

	for i, check := range checks {
		summary.Checks[i] = HealthCheckSummary{
			Name:      check.Name,
			Status:    string(check.Status),
			Message:   check.Message,
			LastCheck: check.LastChecked,
			Duration:  check.Duration,
		}

		switch check.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusWarning:
			summary.Warning++
		case HealthStatusCritical:
			summary.Critical++
		}
	}

	return summary
}

// GetAlertsSummary returns a summary of alerts
func (m *MonitoringIntegrationService) GetAlertsSummary() AlertsSummary {
	if m.alertingService == nil {
		return AlertsSummary{}
	}

	activeAlerts := m.alertingService.GetActiveAlerts()
	summary := AlertsSummary{
		Total:    len(activeAlerts),
		Critical: 0,
		Warning:  0,
		Info:     0,
		Alerts:   make([]AlertSummary, len(activeAlerts)),
	}

	for i, alert := range activeAlerts {
		summary.Alerts[i] = AlertSummary{
			ID:          alert.ID,
			Title:       alert.Title,
			Description: alert.Description,
			Severity:    alert.Severity,
			Source:      alert.Source,
			Timestamp:   alert.Timestamp,
		}

		switch alert.Severity {
		case "critical":
			summary.Critical++
		case "warning":
			summary.Warning++
		case "info":
			summary.Info++
		}
	}

	return summary
}

// metricsCollectionLoop runs the metrics collection loop
func (m *MonitoringIntegrationService) metricsCollectionLoop() {
	for {
		select {
		case <-m.metricsTicker.C:
			// In a real implementation, you would collect and process metrics here
			m.logger.Debug("Metrics collection tick")
		case <-m.stopCh:
			m.logger.Info("Metrics collection loop stopped")
			return
		}
	}
}

// healthCheckLoop runs the health check loop
func (m *MonitoringIntegrationService) healthCheckLoop() {
	for {
		select {
		case <-m.healthCheckTicker.C:
			// Perform health checks
			checks := m.healthCheckService.GetAllHealthChecks()

			// Check for critical issues and create alerts
			for _, check := range checks {
				if check.Status == HealthStatusCritical {
					alert := &Alert{
						Title:       fmt.Sprintf("Health Check Failed: %s", check.Name),
						Description: check.Message,
						Severity:    "critical",
						Source:      "health_check_service",
						Labels: map[string]string{
							"service": "health_check",
							"check":   check.Name,
						},
					}

					if m.alertingService != nil {
						m.alertingService.CreateAlert(alert)
					}
				}
			}

			m.logger.Debug("Health check tick", zap.Int("checks", len(checks)))
		case <-m.stopCh:
			m.logger.Info("Health check loop stopped")
			return
		}
	}
}

// alertingLoop runs the alerting loop
func (m *MonitoringIntegrationService) alertingLoop() {
	for {
		select {
		case <-m.alertingTicker.C:
			// Collect current metrics
			metrics := make(map[string]float64)
			if m.metricsCollector != nil {
				metrics["high_error_rate"] = m.metricsCollector.GetErrorRate()
				metrics["high_response_time"] = m.metricsCollector.GetResponseTime()
				metrics["high_memory_usage"] = m.metricsCollector.GetMemoryUsage()
				metrics["high_cpu_usage"] = m.metricsCollector.GetCPUUsage()
				metrics["high_active_users"] = float64(m.metricsCollector.GetActiveUsers())
				metrics["high_request_rate"] = m.metricsCollector.GetRequestRate()
			}

			// Check alert rules
			if m.alertingService != nil {
				m.alertingService.CheckAlertRules(context.Background(), metrics)
			}

			m.logger.Debug("Alerting tick")
		case <-m.stopCh:
			m.logger.Info("Alerting loop stopped")
			return
		}
	}
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	Timestamp      time.Time                `json:"timestamp"`
	Overall        string                   `json:"overall"`
	Services       map[string]ServiceStatus `json:"services"`
	Metrics        MetricsSummary           `json:"metrics"`
	ActiveAlerts   int                      `json:"active_alerts"`
	CriticalAlerts int                      `json:"critical_alerts"`
	WarningAlerts  int                      `json:"warning_alerts"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	LastCheck time.Time     `json:"last_check"`
	Duration  time.Duration `json:"duration"`
}

// MetricsSummary represents a summary of metrics
type MetricsSummary struct {
	RequestRate  float64 `json:"request_rate"`
	ResponseTime float64 `json:"response_time"`
	ErrorRate    float64 `json:"error_rate"`
	ActiveUsers  int64   `json:"active_users"`
	MemoryUsage  float64 `json:"memory_usage"`
	CPUUsage     float64 `json:"cpu_usage"`
}

// HealthSummary represents a summary of health checks
type HealthSummary struct {
	Total    int                  `json:"total"`
	Healthy  int                  `json:"healthy"`
	Warning  int                  `json:"warning"`
	Critical int                  `json:"critical"`
	Checks   []HealthCheckSummary `json:"checks"`
}

// HealthCheckSummary represents a summary of a health check
type HealthCheckSummary struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	LastCheck time.Time     `json:"last_check"`
	Duration  time.Duration `json:"duration"`
}

// AlertsSummary represents a summary of alerts
type AlertsSummary struct {
	Total    int            `json:"total"`
	Critical int            `json:"critical"`
	Warning  int            `json:"warning"`
	Info     int            `json:"info"`
	Alerts   []AlertSummary `json:"alerts"`
}

// AlertSummary represents a summary of an alert
type AlertSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
}
