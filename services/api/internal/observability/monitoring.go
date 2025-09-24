package observability

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ApplicationMonitoringService provides comprehensive application monitoring
type ApplicationMonitoringService struct {
	logger             *Logger
	metricsCollector   *MetricsCollector
	errorTracker       *ErrorTracker
	userAnalytics      *UserAnalytics
	healthChecker      *HealthChecker
	performanceMonitor *PerformanceMonitor
	config             *MonitoringConfig
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
	started            bool
}

// MonitoringConfig holds configuration for application monitoring
type MonitoringConfig struct {
	Enabled                      bool
	MetricsCollectionInterval    time.Duration
	ErrorTrackingEnabled         bool
	UserAnalyticsEnabled         bool
	HealthCheckInterval          time.Duration
	PerformanceMonitoringEnabled bool
	Environment                  string
	ServiceName                  string
	Version                      string
	Tags                         map[string]string
}

// NewApplicationMonitoringService creates a new application monitoring service
func NewApplicationMonitoringService(logger *Logger, config *MonitoringConfig) *ApplicationMonitoringService {
	ctx, cancel := context.WithCancel(context.Background())

	return &ApplicationMonitoringService{
		logger: logger,
		config: config,
		ctx:    ctx,
		cancel: cancel,
		metricsCollector: &MetricsCollector{
			logger:    logger,
			metrics:   make(map[string]*Metric),
			exporters: make([]MetricsExporter, 0),
		},
		errorTracker: &ErrorTracker{
			logger:    logger,
			errors:    make(map[string]*ErrorEvent),
			exporters: make([]ErrorExporter, 0),
			alertConfig: &ErrorAlertConfig{
				Enabled:           true,
				CriticalThreshold: 10,
				HighThreshold:     50,
				MediumThreshold:   100,
				TimeWindow:        5 * time.Minute,
				AlertChannels:     []string{"email", "slack"},
			},
		},
		userAnalytics: &UserAnalytics{
			logger:    logger,
			events:    make(map[string]*UserEvent),
			exporters: make([]UserAnalyticsExporter, 0),
			config: &UserAnalyticsConfig{
				Enabled:              true,
				TrackPageViews:       true,
				TrackClicks:          true,
				TrackFormSubmissions: true,
				TrackAPIUsage:        true,
				AnonymizeIP:          true,
				RetentionDays:        30,
				BatchSize:            100,
				FlushInterval:        30 * time.Second,
			},
		},
		healthChecker: &HealthChecker{
			logger:    logger,
			checks:    make(map[string]*HealthCheck),
			exporters: make([]HealthExporter, 0),
			config: &HealthCheckConfig{
				Enabled:        true,
				CheckInterval:  30 * time.Second,
				Timeout:        10 * time.Second,
				RetryCount:     3,
				RetryInterval:  5 * time.Second,
				AlertOnFailure: true,
				AlertChannels:  []string{"email", "slack"},
			},
		},
		performanceMonitor: &PerformanceMonitor{
			logger:    logger,
			metrics:   make(map[string]*PerformanceMetric),
			exporters: make([]PerformanceExporter, 0),
			config: &PerformanceConfig{
				Enabled:              true,
				CollectionInterval:   30 * time.Second,
				TrackHTTPRequests:    true,
				TrackDatabaseQueries: true,
				TrackExternalAPIs:    true,
				TrackMemoryUsage:     true,
				TrackCPUUsage:        true,
				TrackGoroutines:      true,
				TrackGC:              true,
				Percentiles:          []float64{0.5, 0.9, 0.95, 0.99},
			},
		},
	}
}

// Start starts the monitoring service
func (ams *ApplicationMonitoringService) Start() error {
	ams.mu.Lock()
	defer ams.mu.Unlock()

	if ams.started {
		return fmt.Errorf("monitoring service already started")
	}

	ams.logger.Info("Starting application monitoring service", map[string]interface{}{
		"service_name": ams.config.ServiceName,
		"version":      ams.config.Version,
		"environment":  ams.config.Environment,
	})

	// Start metrics collection
	if ams.config.Enabled {
		go ams.startMetricsCollection()
	}

	// Start error tracking
	if ams.config.ErrorTrackingEnabled {
		go ams.startErrorTracking()
	}

	// Start user analytics
	if ams.config.UserAnalyticsEnabled {
		go ams.startUserAnalytics()
	}

	// Start health checks
	go ams.startHealthChecks()

	// Start performance monitoring
	if ams.config.PerformanceMonitoringEnabled {
		go ams.startPerformanceMonitoring()
	}

	ams.started = true
	ams.logger.Info("Application monitoring service started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the monitoring service
func (ams *ApplicationMonitoringService) Stop() error {
	ams.mu.Lock()
	defer ams.mu.Unlock()

	if !ams.started {
		return fmt.Errorf("monitoring service not started")
	}

	ams.logger.Info("Stopping application monitoring service", map[string]interface{}{})
	ams.cancel()
	ams.started = false
	ams.logger.Info("Application monitoring service stopped", map[string]interface{}{})
	return nil
}

// RecordMetric records a metric
func (ams *ApplicationMonitoringService) RecordMetric(name string, value float64, metricType MetricType, labels map[string]string) {
	if !ams.config.Enabled {
		return
	}

	ams.metricsCollector.RecordMetric(name, value, metricType, labels)
}

// IncrementCounter increments a counter metric
func (ams *ApplicationMonitoringService) IncrementCounter(name string, labels map[string]string) {
	ams.RecordMetric(name, 1, MetricTypeCounter, labels)
}

// SetGauge sets a gauge metric
func (ams *ApplicationMonitoringService) SetGauge(name string, value float64, labels map[string]string) {
	ams.RecordMetric(name, value, MetricTypeGauge, labels)
}

// RecordHistogram records a histogram metric
func (ams *ApplicationMonitoringService) RecordHistogram(name string, value float64, labels map[string]string) {
	ams.RecordMetric(name, value, MetricTypeHistogram, labels)
}

// TrackError tracks an error
func (ams *ApplicationMonitoringService) TrackError(err error, severity ErrorSeverity, context map[string]interface{}, tags map[string]string) {
	if !ams.config.ErrorTrackingEnabled {
		return
	}

	ams.errorTracker.TrackError(err, severity, context, tags)
}

// TrackUserEvent tracks a user event
func (ams *ApplicationMonitoringService) TrackUserEvent(userID, sessionID, eventType string, eventData map[string]interface{}, tags map[string]string) {
	if !ams.config.UserAnalyticsEnabled {
		return
	}

	ams.userAnalytics.TrackEvent(userID, sessionID, eventType, eventData, tags)
}

// AddHealthCheck adds a health check
func (ams *ApplicationMonitoringService) AddHealthCheck(name string, checkFunc func() error, interval time.Duration, critical bool) {
	ams.healthChecker.AddCheck(name, checkFunc, interval, critical)
}

// GetHealthStatus returns the overall health status
func (ams *ApplicationMonitoringService) GetHealthStatus() map[string]interface{} {
	return ams.healthChecker.GetStatus()
}

// GetMetrics returns current metrics
func (ams *ApplicationMonitoringService) GetMetrics() map[string]*Metric {
	return ams.metricsCollector.GetMetrics()
}

// GetPerformanceMetrics returns current performance metrics
func (ams *ApplicationMonitoringService) GetPerformanceMetrics() map[string]*PerformanceMetric {
	return ams.performanceMonitor.GetMetrics()
}

// GetErrorSummary returns error summary
func (ams *ApplicationMonitoringService) GetErrorSummary() map[string]interface{} {
	return ams.errorTracker.GetSummary()
}

// GetUserAnalyticsSummary returns user analytics summary
func (ams *ApplicationMonitoringService) GetUserAnalyticsSummary() map[string]interface{} {
	return ams.userAnalytics.GetSummary()
}

// startMetricsCollection starts the metrics collection routine
func (ams *ApplicationMonitoringService) startMetricsCollection() {
	ticker := time.NewTicker(ams.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ams.ctx.Done():
			return
		case <-ticker.C:
			ams.collectSystemMetrics()
		}
	}
}

// startErrorTracking starts the error tracking routine
func (ams *ApplicationMonitoringService) startErrorTracking() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ams.ctx.Done():
			return
		case <-ticker.C:
			ams.errorTracker.ProcessAlerts()
		}
	}
}

// startUserAnalytics starts the user analytics routine
func (ams *ApplicationMonitoringService) startUserAnalytics() {
	ticker := time.NewTicker(ams.userAnalytics.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ams.ctx.Done():
			return
		case <-ticker.C:
			ams.userAnalytics.FlushEvents()
		}
	}
}

// startHealthChecks starts the health check routine
func (ams *ApplicationMonitoringService) startHealthChecks() {
	ticker := time.NewTicker(ams.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ams.ctx.Done():
			return
		case <-ticker.C:
			ams.healthChecker.RunChecks()
		}
	}
}

// startPerformanceMonitoring starts the performance monitoring routine
func (ams *ApplicationMonitoringService) startPerformanceMonitoring() {
	ticker := time.NewTicker(ams.performanceMonitor.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ams.ctx.Done():
			return
		case <-ticker.C:
			ams.collectPerformanceMetrics()
		}
	}
}

// collectSystemMetrics collects system-level metrics
func (ams *ApplicationMonitoringService) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	ams.SetGauge("system_memory_alloc_bytes", float64(m.Alloc), map[string]string{"type": "alloc"})
	ams.SetGauge("system_memory_total_alloc_bytes", float64(m.TotalAlloc), map[string]string{"type": "total_alloc"})
	ams.SetGauge("system_memory_sys_bytes", float64(m.Sys), map[string]string{"type": "sys"})
	ams.SetGauge("system_memory_heap_alloc_bytes", float64(m.HeapAlloc), map[string]string{"type": "heap_alloc"})
	ams.SetGauge("system_memory_heap_sys_bytes", float64(m.HeapSys), map[string]string{"type": "heap_sys"})
	ams.SetGauge("system_memory_heap_idle_bytes", float64(m.HeapIdle), map[string]string{"type": "heap_idle"})
	ams.SetGauge("system_memory_heap_inuse_bytes", float64(m.HeapInuse), map[string]string{"type": "heap_inuse"})

	// GC metrics
	ams.SetGauge("system_gc_runs_total", float64(m.NumGC), map[string]string{})
	ams.SetGauge("system_gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), map[string]string{})

	// Goroutine metrics
	ams.SetGauge("system_goroutines", float64(runtime.NumGoroutine()), map[string]string{})

	// CPU metrics (simplified)
	ams.SetGauge("system_cpu_usage_percent", 0.0, map[string]string{}) // Would need actual CPU monitoring

	ams.logger.Debug("System metrics collected", map[string]interface{}{
		"memory_alloc": m.Alloc,
		"goroutines":   runtime.NumGoroutine(),
		"gc_runs":      m.NumGC,
	})
}

// collectPerformanceMetrics collects performance metrics
func (ams *ApplicationMonitoringService) collectPerformanceMetrics() {
	if !ams.config.PerformanceMonitoringEnabled {
		return
	}

	// Collect HTTP request metrics
	if ams.performanceMonitor.config.TrackHTTPRequests {
		ams.collectHTTPMetrics()
	}

	// Collect database metrics
	if ams.performanceMonitor.config.TrackDatabaseQueries {
		ams.collectDatabaseMetrics()
	}

	// Collect external API metrics
	if ams.performanceMonitor.config.TrackExternalAPIs {
		ams.collectExternalAPIMetrics()
	}
}

// collectHTTPMetrics collects HTTP-related performance metrics
func (ams *ApplicationMonitoringService) collectHTTPMetrics() {
	// This would integrate with HTTP middleware to collect actual metrics
	ams.logger.Debug("Collecting HTTP performance metrics", map[string]interface{}{})
}

// collectDatabaseMetrics collects database-related performance metrics
func (ams *ApplicationMonitoringService) collectDatabaseMetrics() {
	// This would integrate with database layer to collect actual metrics
	ams.logger.Debug("Collecting database performance metrics", map[string]interface{}{})
}

// collectExternalAPIMetrics collects external API performance metrics
func (ams *ApplicationMonitoringService) collectExternalAPIMetrics() {
	// This would integrate with external API clients to collect actual metrics
	ams.logger.Debug("Collecting external API performance metrics", map[string]interface{}{})
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:                      true,
		MetricsCollectionInterval:    30 * time.Second,
		ErrorTrackingEnabled:         true,
		UserAnalyticsEnabled:         true,
		HealthCheckInterval:          30 * time.Second,
		PerformanceMonitoringEnabled: true,
		Environment:                  "development",
		ServiceName:                  "kyb-platform",
		Version:                      "1.0.0",
		Tags: map[string]string{
			"service": "kyb-platform",
			"version": "1.0.0",
		},
	}
}
