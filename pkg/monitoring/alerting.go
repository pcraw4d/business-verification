package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents an alert that can be sent
type Alert struct {
	Level     AlertLevel             `json:"level"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AlertHandler defines the interface for alert handlers
type AlertHandler interface {
	SendAlert(ctx context.Context, alert Alert) error
	GetName() string
}

// AlertManager manages alerting across the system
type AlertManager struct {
	handlers []AlertHandler
	throttle map[string]time.Time // Throttle alerts to prevent spam
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		handlers: make([]AlertHandler, 0),
		throttle: make(map[string]time.Time),
	}
}

// AddHandler adds an alert handler
func (am *AlertManager) AddHandler(handler AlertHandler) {
	am.handlers = append(am.handlers, handler)
}

// SendAlert sends an alert through all handlers
func (am *AlertManager) SendAlert(ctx context.Context, alert Alert) error {
	// Create throttle key
	throttleKey := fmt.Sprintf("%s:%s", alert.Level, alert.Title)

	// Check if we should throttle this alert
	if lastSent, exists := am.throttle[throttleKey]; exists {
		if time.Since(lastSent) < 5*time.Minute {
			// Skip sending - too recent
			return nil
		}
	}

	// Update throttle
	am.throttle[throttleKey] = time.Now()

	// Send through all handlers
	for _, handler := range am.handlers {
		if err := handler.SendAlert(ctx, alert); err != nil {
			log.Printf("Failed to send alert via %s: %v", handler.GetName(), err)
		}
	}

	return nil
}

// LogAlertHandler logs alerts to the system log
type LogAlertHandler struct {
	name string
}

// NewLogAlertHandler creates a new log alert handler
func NewLogAlertHandler() *LogAlertHandler {
	return &LogAlertHandler{name: "log"}
}

func (lah *LogAlertHandler) GetName() string {
	return lah.name
}

func (lah *LogAlertHandler) SendAlert(ctx context.Context, alert Alert) error {
	log.Printf("🚨 ALERT [%s] %s: %s", alert.Level, alert.Title, alert.Message)
	return nil
}

// ThresholdMonitor monitors metrics against thresholds
type ThresholdMonitor struct {
	alertManager *AlertManager
	serviceName  string
}

// NewThresholdMonitor creates a new threshold monitor
func NewThresholdMonitor(alertManager *AlertManager, serviceName string) *ThresholdMonitor {
	return &ThresholdMonitor{
		alertManager: alertManager,
		serviceName:  serviceName,
	}
}

// CheckResponseTime checks if response time exceeds threshold
func (tm *ThresholdMonitor) CheckResponseTime(ctx context.Context, avgResponseTime time.Duration) {
	if avgResponseTime > 2*time.Second {
		alert := Alert{
			Level:     AlertLevelWarning,
			Title:     "High Response Time",
			Message:   fmt.Sprintf("Average response time is %v, exceeding 2s threshold", avgResponseTime),
			Service:   tm.serviceName,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"avg_response_time": avgResponseTime.String(),
				"threshold":         "2s",
			},
		}
		tm.alertManager.SendAlert(ctx, alert)
	}
}

// CheckErrorRate checks if error rate exceeds threshold
func (tm *ThresholdMonitor) CheckErrorRate(ctx context.Context, errorRate float64) {
	if errorRate > 5.0 { // 5% error rate
		alert := Alert{
			Level:     AlertLevelCritical,
			Title:     "High Error Rate",
			Message:   fmt.Sprintf("Error rate is %.2f%%, exceeding 5%% threshold", errorRate),
			Service:   tm.serviceName,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"error_rate": errorRate,
				"threshold":  "5%",
			},
		}
		tm.alertManager.SendAlert(ctx, alert)
	}
}

// CheckCacheHitRate checks if cache hit rate is below threshold
func (tm *ThresholdMonitor) CheckCacheHitRate(ctx context.Context, hitRate float64) {
	if hitRate < 70.0 { // 70% hit rate
		alert := Alert{
			Level:     AlertLevelWarning,
			Title:     "Low Cache Hit Rate",
			Message:   fmt.Sprintf("Cache hit rate is %.2f%%, below 70%% threshold", hitRate),
			Service:   tm.serviceName,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"hit_rate":  hitRate,
				"threshold": "70%",
			},
		}
		tm.alertManager.SendAlert(ctx, alert)
	}
}

// CheckHealthStatus checks if any health check is failing
func (tm *ThresholdMonitor) CheckHealthStatus(ctx context.Context, healthStatus map[string]interface{}) {
	if status, ok := healthStatus["status"].(string); ok {
		if status == "unhealthy" {
			alert := Alert{
				Level:     AlertLevelCritical,
				Title:     "Service Unhealthy",
				Message:   "One or more health checks are failing",
				Service:   tm.serviceName,
				Timestamp: time.Now(),
				Metadata:  healthStatus,
			}
			tm.alertManager.SendAlert(ctx, alert)
		} else if status == "degraded" {
			alert := Alert{
				Level:     AlertLevelWarning,
				Title:     "Service Degraded",
				Message:   "One or more health checks are degraded",
				Service:   tm.serviceName,
				Timestamp: time.Now(),
				Metadata:  healthStatus,
			}
			tm.alertManager.SendAlert(ctx, alert)
		}
	}
}
