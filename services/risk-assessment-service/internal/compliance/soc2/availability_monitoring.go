package soc2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AvailabilityMonitor implements SOC 2 availability monitoring requirements
type AvailabilityMonitor struct {
	logger    *zap.Logger
	config    *AvailabilityConfig
	metrics   *AvailabilityMetrics
	mu        sync.RWMutex
	startTime time.Time
	uptime    time.Duration
	downtime  time.Duration
	incidents []*AvailabilityIncident
}

// AvailabilityConfig represents availability monitoring configuration
type AvailabilityConfig struct {
	TargetUptime         float64       `json:"target_uptime"`         // Target uptime percentage (e.g., 99.9)
	MonitoringInterval   time.Duration `json:"monitoring_interval"`   // How often to check availability
	HealthCheckTimeout   time.Duration `json:"health_check_timeout"`  // Timeout for health checks
	EnableNotifications  bool          `json:"enable_notifications"`  // Enable downtime notifications
	NotificationChannels []string      `json:"notification_channels"` // Channels to send notifications
	EnableAutoRecovery   bool          `json:"enable_auto_recovery"`  // Enable automatic recovery attempts
	MaxDowntimeMinutes   int           `json:"max_downtime_minutes"`  // Maximum allowed downtime in minutes
}

// AvailabilityMetrics represents current availability metrics
type AvailabilityMetrics struct {
	CurrentUptime       float64       `json:"current_uptime"`
	TotalUptime         time.Duration `json:"total_uptime"`
	TotalDowntime       time.Duration `json:"total_downtime"`
	TotalIncidents      int           `json:"total_incidents"`
	LastIncident        *time.Time    `json:"last_incident"`
	CurrentStatus       ServiceStatus `json:"current_status"`
	HealthScore         float64       `json:"health_score"`
	ResponseTime        time.Duration `json:"response_time"`
	ErrorRate           float64       `json:"error_rate"`
	LastHealthCheck     time.Time     `json:"last_health_check"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
}

// AvailabilityIncident represents a service availability incident
type AvailabilityIncident struct {
	ID          string                 `json:"id"`
	Type        IncidentType           `json:"type"`
	Severity    IncidentSeverity       `json:"severity"`
	Status      IncidentStatus         `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Description string                 `json:"description"`
	RootCause   string                 `json:"root_cause"`
	Impact      string                 `json:"impact"`
	Resolution  string                 `json:"resolution"`
	DetectedBy  string                 `json:"detected_by"`
	ResolvedBy  string                 `json:"resolved_by"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ServiceStatus represents the current status of the service
type ServiceStatus string

const (
	ServiceStatusHealthy   ServiceStatus = "healthy"
	ServiceStatusDegraded  ServiceStatus = "degraded"
	ServiceStatusUnhealthy ServiceStatus = "unhealthy"
	ServiceStatusUnknown   ServiceStatus = "unknown"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	ID           string                 `json:"id"`
	Service      string                 `json:"service"`
	Status       ServiceStatus          `json:"status"`
	ResponseTime time.Duration          `json:"response_time"`
	Error        string                 `json:"error,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewAvailabilityMonitor creates a new availability monitor
func NewAvailabilityMonitor(config *AvailabilityConfig, logger *zap.Logger) *AvailabilityMonitor {
	return &AvailabilityMonitor{
		logger: logger,
		config: config,
		metrics: &AvailabilityMetrics{
			CurrentStatus: ServiceStatusUnknown,
			HealthScore:   100.0,
		},
		startTime: time.Now(),
		incidents: make([]*AvailabilityIncident, 0),
	}
}

// Start begins availability monitoring
func (am *AvailabilityMonitor) Start(ctx context.Context) error {
	am.logger.Info("Starting availability monitoring",
		zap.Float64("target_uptime", am.config.TargetUptime),
		zap.Duration("monitoring_interval", am.config.MonitoringInterval))

	// Start monitoring loop
	go am.monitoringLoop(ctx)

	return nil
}

// Stop stops availability monitoring
func (am *AvailabilityMonitor) Stop() error {
	am.logger.Info("Stopping availability monitoring")
	return nil
}

// GetMetrics returns current availability metrics
func (am *AvailabilityMonitor) GetMetrics() *AvailabilityMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Calculate current uptime
	totalTime := time.Since(am.startTime)
	currentUptime := float64(am.uptime) / float64(totalTime) * 100

	// Create a copy of metrics
	metrics := *am.metrics
	metrics.CurrentUptime = currentUptime
	metrics.TotalUptime = am.uptime
	metrics.TotalDowntime = am.downtime

	return &metrics
}

// RecordHealthCheck records a health check result
func (am *AvailabilityMonitor) RecordHealthCheck(ctx context.Context, check *HealthCheck) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	am.metrics.LastHealthCheck = now
	am.metrics.ResponseTime = check.ResponseTime

	// Update status based on health check
	previousStatus := am.metrics.CurrentStatus
	am.metrics.CurrentStatus = check.Status

	// Handle status changes
	if previousStatus != check.Status {
		am.logger.Info("Service status changed",
			zap.String("previous_status", string(previousStatus)),
			zap.String("current_status", string(check.Status)),
			zap.String("service", check.Service))

		// Handle status transitions
		if check.Status == ServiceStatusUnhealthy || check.Status == ServiceStatusDegraded {
			am.handleServiceDowntime(ctx, check)
		} else if previousStatus == ServiceStatusUnhealthy || previousStatus == ServiceStatusDegraded {
			am.handleServiceRecovery(ctx, check)
		}
	}

	// Update consecutive failures
	if check.Status == ServiceStatusUnhealthy {
		am.metrics.ConsecutiveFailures++
	} else {
		am.metrics.ConsecutiveFailures = 0
	}

	// Update health score
	am.updateHealthScore(check)

	return nil
}

// RecordIncident records an availability incident
func (am *AvailabilityMonitor) RecordIncident(ctx context.Context, incident *AvailabilityIncident) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Generate incident ID if not provided
	if incident.ID == "" {
		incident.ID = generateIncidentID()
	}

	// Set timestamps
	now := time.Now()
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = now
	}
	incident.UpdatedAt = now

	// Add to incidents list
	am.incidents = append(am.incidents, incident)
	am.metrics.TotalIncidents++
	am.metrics.LastIncident = &now

	am.logger.Info("Availability incident recorded",
		zap.String("incident_id", incident.ID),
		zap.String("type", string(incident.Type)),
		zap.String("severity", string(incident.Severity)),
		zap.String("description", incident.Description))

	// Send notifications if enabled
	if am.config.EnableNotifications {
		go am.sendIncidentNotification(incident)
	}

	return nil
}

// GetIncidents returns availability incidents
func (am *AvailabilityMonitor) GetIncidents(ctx context.Context, filters map[string]interface{}) ([]*AvailabilityIncident, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Apply filters if provided
	filteredIncidents := make([]*AvailabilityIncident, 0)
	for _, incident := range am.incidents {
		if am.matchesFilters(incident, filters) {
			filteredIncidents = append(filteredIncidents, incident)
		}
	}

	return filteredIncidents, nil
}

// GetUptimePercentage returns the current uptime percentage
func (am *AvailabilityMonitor) GetUptimePercentage() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()

	totalTime := time.Since(am.startTime)
	if totalTime == 0 {
		return 100.0
	}

	return float64(am.uptime) / float64(totalTime) * 100
}

// IsTargetUptimeMet checks if the target uptime is being met
func (am *AvailabilityMonitor) IsTargetUptimeMet() bool {
	currentUptime := am.GetUptimePercentage()
	return currentUptime >= am.config.TargetUptime
}

// GetHealthScore returns the current health score
func (am *AvailabilityMonitor) GetHealthScore() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.metrics.HealthScore
}

// monitoringLoop runs the continuous monitoring loop
func (am *AvailabilityMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(am.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			am.logger.Info("Monitoring loop stopped")
			return
		case <-ticker.C:
			am.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck performs a health check
func (am *AvailabilityMonitor) performHealthCheck(ctx context.Context) {
	start := time.Now()

	// Perform health check (in a real implementation, this would check actual services)
	status := am.checkServiceHealth(ctx)
	responseTime := time.Since(start)

	// Record the health check
	check := &HealthCheck{
		ID:           generateHealthCheckID(),
		Service:      "risk-assessment-service",
		Status:       status,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"check_type": "automated",
		},
	}

	if err := am.RecordHealthCheck(ctx, check); err != nil {
		am.logger.Error("Failed to record health check", zap.Error(err))
	}
}

// checkServiceHealth checks the health of the service
func (am *AvailabilityMonitor) checkServiceHealth(ctx context.Context) ServiceStatus {
	// In a real implementation, this would check:
	// - Database connectivity
	// - External API availability
	// - Memory and CPU usage
	// - Disk space
	// - Network connectivity

	// For now, we'll simulate a healthy service
	// In production, this would perform actual health checks
	return ServiceStatusHealthy
}

// handleServiceDowntime handles when the service goes down
func (am *AvailabilityMonitor) handleServiceDowntime(ctx context.Context, check *HealthCheck) {
	// Create incident if this is a new downtime
	if am.metrics.CurrentStatus == ServiceStatusUnhealthy {
		incident := &AvailabilityIncident{
			Type:        IncidentTypeSystemCompromise,
			Severity:    IncidentSeverityHigh,
			Status:      IncidentStatusOpen,
			StartTime:   time.Now(),
			Description: fmt.Sprintf("Service health check failed: %s", check.Error),
			DetectedBy:  "availability-monitor",
			Metadata: map[string]interface{}{
				"health_check_id": check.ID,
				"response_time":   check.ResponseTime,
			},
		}

		if err := am.RecordIncident(ctx, incident); err != nil {
			am.logger.Error("Failed to record downtime incident", zap.Error(err))
		}
	}

	// Update downtime tracking
	am.downtime += am.config.MonitoringInterval
}

// handleServiceRecovery handles when the service recovers
func (am *AvailabilityMonitor) handleServiceRecovery(ctx context.Context, check *HealthCheck) {
	// Update uptime tracking
	am.uptime += am.config.MonitoringInterval

	// Close any open incidents
	for _, incident := range am.incidents {
		if incident.Status == IncidentStatusOpen && incident.EndTime == nil {
			now := time.Now()
			incident.EndTime = &now
			incident.Duration = now.Sub(incident.StartTime)
			incident.Status = IncidentStatusResolved
			incident.Resolution = "Service recovered automatically"
			incident.ResolvedBy = "availability-monitor"
			incident.UpdatedAt = now

			am.logger.Info("Availability incident resolved",
				zap.String("incident_id", incident.ID),
				zap.Duration("duration", incident.Duration))
		}
	}
}

// updateHealthScore updates the health score based on various factors
func (am *AvailabilityMonitor) updateHealthScore(check *HealthCheck) {
	score := 100.0

	// Deduct points for consecutive failures
	if am.metrics.ConsecutiveFailures > 0 {
		score -= float64(am.metrics.ConsecutiveFailures) * 5.0
	}

	// Deduct points for high response time
	if check.ResponseTime > 1*time.Second {
		score -= 10.0
	} else if check.ResponseTime > 500*time.Millisecond {
		score -= 5.0
	}

	// Deduct points for degraded status
	if check.Status == ServiceStatusDegraded {
		score -= 20.0
	} else if check.Status == ServiceStatusUnhealthy {
		score -= 50.0
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	am.metrics.HealthScore = score
}

// sendIncidentNotification sends incident notifications
func (am *AvailabilityMonitor) sendIncidentNotification(incident *AvailabilityIncident) {
	am.logger.Info("Sending incident notification",
		zap.String("incident_id", incident.ID),
		zap.Strings("channels", am.config.NotificationChannels))

	// In a real implementation, this would send notifications via:
	// - Email
	// - Slack
	// - PagerDuty
	// - SMS
	// - Webhooks
}

// matchesFilters checks if an incident matches the given filters
func (am *AvailabilityMonitor) matchesFilters(incident *AvailabilityIncident, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "type":
			if incident.Type != value {
				return false
			}
		case "severity":
			if incident.Severity != value {
				return false
			}
		case "status":
			if incident.Status != value {
				return false
			}
		case "start_time_after":
			if incident.StartTime.Before(value.(time.Time)) {
				return false
			}
		case "start_time_before":
			if incident.StartTime.After(value.(time.Time)) {
				return false
			}
		}
	}
	return true
}

// generateHealthCheckID generates a unique health check ID
func generateHealthCheckID() string {
	return fmt.Sprintf("hc_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
