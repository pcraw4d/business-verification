package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// PerformanceMonitor provides real-time database performance monitoring
type PerformanceMonitor struct {
	db        *sql.DB
	logger    *log.Logger
	config    *MonitoringConfig
	metrics   *PerformanceMetrics
	mu        sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
}

// MonitoringConfig contains configuration for performance monitoring
type MonitoringConfig struct {
	// Monitoring intervals
	MetricsInterval    time.Duration
	SlowQueryThreshold time.Duration

	// Alert thresholds
	MaxQueryTime       time.Duration
	MaxConnectionCount int
	MinCacheHitRate    float64

	// Data retention
	MetricsRetention  time.Duration
	MaxMetricsHistory int

	// Monitoring features
	MonitorQueries     bool
	MonitorConnections bool
	MonitorCache       bool
	MonitorLocks       bool
}

// PerformanceMetrics contains collected performance metrics
type PerformanceMetrics struct {
	Timestamp         time.Time
	QueryMetrics      *QueryMetrics
	ConnectionMetrics *ConnectionMetrics
	CacheMetrics      *CacheMetrics
	LockMetrics       *LockMetrics
	SystemMetrics     *SystemMetrics
}

// QueryMetrics contains query-related performance metrics
type QueryMetrics struct {
	TotalQueries     int64
	SlowQueries      int64
	AverageQueryTime time.Duration
	MaxQueryTime     time.Duration
	QueriesPerSecond float64
	ErrorRate        float64
	TopSlowQueries   []SlowQuery
}

// ConnectionMetrics contains connection-related metrics
type ConnectionMetrics struct {
	TotalConnections      int
	ActiveConnections     int
	IdleConnections       int
	MaxConnections        int
	ConnectionUtilization float64
}

// CacheMetrics contains cache-related metrics
type CacheMetrics struct {
	CacheHitRate   float64
	CacheSize      int64
	CacheEvictions int64
	CacheMisses    int64
}

// LockMetrics contains lock-related metrics
type LockMetrics struct {
	LockWaits       int64
	Deadlocks       int64
	LockTimeouts    int64
	AverageWaitTime time.Duration
}

// SystemMetrics contains system-level metrics
type SystemMetrics struct {
	DatabaseSize int64
	TableCount   int
	IndexCount   int
	Uptime       time.Duration
}

// SlowQuery represents a slow query with its details
type SlowQuery struct {
	Query        string
	AverageTime  time.Duration
	CallCount    int64
	TotalTime    time.Duration
	LastExecuted time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(db *sql.DB, config *MonitoringConfig) *PerformanceMonitor {
	if config == nil {
		config = &MonitoringConfig{
			MetricsInterval:    30 * time.Second,
			SlowQueryThreshold: 1 * time.Second,
			MaxQueryTime:       5 * time.Second,
			MaxConnectionCount: 100,
			MinCacheHitRate:    0.90,
			MetricsRetention:   24 * time.Hour,
			MaxMetricsHistory:  1000,
			MonitorQueries:     true,
			MonitorConnections: true,
			MonitorCache:       true,
			MonitorLocks:       true,
		}
	}

	return &PerformanceMonitor{
		db:       db,
		logger:   log.New(log.Writer(), "[PERF_MONITOR] ", log.LstdFlags),
		config:   config,
		metrics:  &PerformanceMetrics{},
		stopChan: make(chan struct{}),
	}
}

// Start begins performance monitoring
func (pm *PerformanceMonitor) Start(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		return fmt.Errorf("performance monitor is already running")
	}

	pm.isRunning = true
	pm.logger.Println("Starting performance monitoring...")

	// Start monitoring goroutine
	go pm.monitoringLoop(ctx)

	return nil
}

// Stop stops performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return
	}

	pm.logger.Println("Stopping performance monitoring...")
	close(pm.stopChan)
	pm.isRunning = false
}

// GetCurrentMetrics returns the current performance metrics
func (pm *PerformanceMonitor) GetCurrentMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy of the metrics
	metricsCopy := *pm.metrics
	return &metricsCopy
}

// GetMetricsHistory returns historical performance metrics
func (pm *PerformanceMonitor) GetMetricsHistory(limit int) ([]*PerformanceMetrics, error) {
	// This would typically query a metrics storage system
	// For now, return empty slice as placeholder
	return []*PerformanceMetrics{}, nil
}

// monitoringLoop runs the main monitoring loop
func (pm *PerformanceMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(pm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Println("Monitoring stopped due to context cancellation")
			return
		case <-pm.stopChan:
			pm.logger.Println("Monitoring stopped")
			return
		case <-ticker.C:
			if err := pm.collectMetrics(ctx); err != nil {
				pm.logger.Printf("Failed to collect metrics: %v", err)
			}
		}
	}
}

// collectMetrics collects all performance metrics
func (pm *PerformanceMonitor) collectMetrics(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics.Timestamp = time.Now()

	// Collect query metrics
	if pm.config.MonitorQueries {
		if err := pm.collectQueryMetrics(ctx); err != nil {
			pm.logger.Printf("Failed to collect query metrics: %v", err)
		}
	}

	// Collect connection metrics
	if pm.config.MonitorConnections {
		if err := pm.collectConnectionMetrics(ctx); err != nil {
			pm.logger.Printf("Failed to collect connection metrics: %v", err)
		}
	}

	// Collect cache metrics
	if pm.config.MonitorCache {
		if err := pm.collectCacheMetrics(ctx); err != nil {
			pm.logger.Printf("Failed to collect cache metrics: %v", err)
		}
	}

	// Collect lock metrics
	if pm.config.MonitorLocks {
		if err := pm.collectLockMetrics(ctx); err != nil {
			pm.logger.Printf("Failed to collect lock metrics: %v", err)
		}
	}

	// Collect system metrics
	if err := pm.collectSystemMetrics(ctx); err != nil {
		pm.logger.Printf("Failed to collect system metrics: %v", err)
	}

	// Check for alerts
	pm.checkAlerts()

	return nil
}

// collectQueryMetrics collects query-related metrics
func (pm *PerformanceMonitor) collectQueryMetrics(ctx context.Context) error {
	pm.metrics.QueryMetrics = &QueryMetrics{}

	// Get query statistics from pg_stat_statements
	query := `
		SELECT 
			query,
			calls,
			total_time,
			mean_time,
			rows
		FROM pg_stat_statements 
		WHERE mean_time > $1
		ORDER BY mean_time DESC
		LIMIT 10
	`

	rows, err := pm.db.QueryContext(ctx, query, pm.config.SlowQueryThreshold.Milliseconds())
	if err != nil {
		return fmt.Errorf("failed to query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	var totalQueries int64
	var totalTime float64
	var slowQueries int64

	for rows.Next() {
		var queryText string
		var calls, totalTimeMs, meanTimeMs float64
		var rowsCount int64

		if err := rows.Scan(&queryText, &calls, &totalTimeMs, &meanTimeMs, &rowsCount); err != nil {
			continue
		}

		totalQueries += int64(calls)
		totalTime += totalTimeMs

		if meanTimeMs > pm.config.SlowQueryThreshold.Milliseconds() {
			slowQueries += int64(calls)

			// Add to slow queries list
			pm.metrics.QueryMetrics.TopSlowQueries = append(pm.metrics.QueryMetrics.TopSlowQueries, SlowQuery{
				Query:        truncateQuery(queryText, 100),
				AverageTime:  time.Duration(meanTimeMs) * time.Millisecond,
				CallCount:    int64(calls),
				TotalTime:    time.Duration(totalTimeMs) * time.Millisecond,
				LastExecuted: time.Now(), // This would ideally come from the query
			})
		}
	}

	// Calculate metrics
	pm.metrics.QueryMetrics.TotalQueries = totalQueries
	pm.metrics.QueryMetrics.SlowQueries = slowQueries

	if totalQueries > 0 {
		pm.metrics.QueryMetrics.AverageQueryTime = time.Duration(totalTime/float64(totalQueries)) * time.Millisecond
		pm.metrics.QueryMetrics.QueriesPerSecond = float64(totalQueries) / pm.config.MetricsInterval.Seconds()
		pm.metrics.QueryMetrics.ErrorRate = float64(slowQueries) / float64(totalQueries) * 100
	}

	return nil
}

// collectConnectionMetrics collects connection-related metrics
func (pm *PerformanceMonitor) collectConnectionMetrics(ctx context.Context) error {
	pm.metrics.ConnectionMetrics = &ConnectionMetrics{}

	// Get connection statistics
	query := `
		SELECT 
			count(*) as total_connections,
			count(*) FILTER (WHERE state = 'active') as active_connections,
			count(*) FILTER (WHERE state = 'idle') as idle_connections
		FROM pg_stat_activity
	`

	var total, active, idle int
	err := pm.db.QueryRowContext(ctx, query).Scan(&total, &active, &idle)
	if err != nil {
		return fmt.Errorf("failed to get connection metrics: %w", err)
	}

	pm.metrics.ConnectionMetrics.TotalConnections = total
	pm.metrics.ConnectionMetrics.ActiveConnections = active
	pm.metrics.ConnectionMetrics.IdleConnections = idle

	// Get max connections setting
	var maxConnections int
	err = pm.db.QueryRowContext(ctx, "SHOW max_connections").Scan(&maxConnections)
	if err != nil {
		pm.logger.Printf("Failed to get max_connections setting: %v", err)
		maxConnections = 100 // Default fallback
	}

	pm.metrics.ConnectionMetrics.MaxConnections = maxConnections
	pm.metrics.ConnectionMetrics.ConnectionUtilization = float64(total) / float64(maxConnections) * 100

	return nil
}

// collectCacheMetrics collects cache-related metrics
func (pm *PerformanceMonitor) collectCacheMetrics(ctx context.Context) error {
	pm.metrics.CacheMetrics = &CacheMetrics{}

	// Get buffer cache hit ratio
	query := `
		SELECT 
			round(
				(sum(blks_hit) * 100.0 / (sum(blks_hit) + sum(blks_read))), 2
			) as cache_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	var hitRatio float64
	err := pm.db.QueryRowContext(ctx, query).Scan(&hitRatio)
	if err != nil {
		return fmt.Errorf("failed to get cache hit ratio: %w", err)
	}

	pm.metrics.CacheMetrics.CacheHitRate = hitRatio / 100.0

	// Get shared buffer size
	var sharedBuffers int64
	err = pm.db.QueryRowContext(ctx, "SHOW shared_buffers").Scan(&sharedBuffers)
	if err != nil {
		pm.logger.Printf("Failed to get shared_buffers setting: %v", err)
	} else {
		pm.metrics.CacheMetrics.CacheSize = sharedBuffers
	}

	return nil
}

// collectLockMetrics collects lock-related metrics
func (pm *PerformanceMonitor) collectLockMetrics(ctx context.Context) error {
	pm.metrics.LockMetrics = &LockMetrics{}

	// Get lock statistics
	query := `
		SELECT 
			count(*) FILTER (WHERE wait_event_type = 'Lock') as lock_waits,
			count(*) FILTER (WHERE wait_event = 'deadlock_detection') as deadlocks
		FROM pg_stat_activity
	`

	var lockWaits, deadlocks int64
	err := pm.db.QueryRowContext(ctx, query).Scan(&lockWaits, &deadlocks)
	if err != nil {
		return fmt.Errorf("failed to get lock metrics: %w", err)
	}

	pm.metrics.LockMetrics.LockWaits = lockWaits
	pm.metrics.LockMetrics.Deadlocks = deadlocks

	return nil
}

// collectSystemMetrics collects system-level metrics
func (pm *PerformanceMonitor) collectSystemMetrics(ctx context.Context) error {
	pm.metrics.SystemMetrics = &SystemMetrics{}

	// Get database size
	var dbSize int64
	err := pm.db.QueryRowContext(ctx, "SELECT pg_database_size(current_database())").Scan(&dbSize)
	if err != nil {
		return fmt.Errorf("failed to get database size: %w", err)
	}
	pm.metrics.SystemMetrics.DatabaseSize = dbSize

	// Get table count
	var tableCount int
	err = pm.db.QueryRowContext(ctx, `
		SELECT count(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to get table count: %w", err)
	}
	pm.metrics.SystemMetrics.TableCount = tableCount

	// Get index count
	var indexCount int
	err = pm.db.QueryRowContext(ctx, `
		SELECT count(*) 
		FROM pg_indexes 
		WHERE schemaname = 'public'
	`).Scan(&indexCount)
	if err != nil {
		return fmt.Errorf("failed to get index count: %w", err)
	}
	pm.metrics.SystemMetrics.IndexCount = indexCount

	// Get database uptime
	var uptimeSeconds int64
	err = pm.db.QueryRowContext(ctx, `
		SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))
	`).Scan(&uptimeSeconds)
	if err != nil {
		return fmt.Errorf("failed to get uptime: %w", err)
	}
	pm.metrics.SystemMetrics.Uptime = time.Duration(uptimeSeconds) * time.Second

	return nil
}

// checkAlerts checks for performance alerts
func (pm *PerformanceMonitor) checkAlerts() {
	alerts := []string{}

	// Check query performance
	if pm.metrics.QueryMetrics != nil {
		if pm.metrics.QueryMetrics.AverageQueryTime > pm.config.MaxQueryTime {
			alerts = append(alerts, fmt.Sprintf("Average query time (%.2fms) exceeds threshold (%.2fms)",
				float64(pm.metrics.QueryMetrics.AverageQueryTime.Nanoseconds())/1e6,
				float64(pm.config.MaxQueryTime.Nanoseconds())/1e6))
		}

		if pm.metrics.QueryMetrics.ErrorRate > 10.0 {
			alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%%", pm.metrics.QueryMetrics.ErrorRate))
		}
	}

	// Check connection usage
	if pm.metrics.ConnectionMetrics != nil {
		if pm.metrics.ConnectionMetrics.TotalConnections > pm.config.MaxConnectionCount {
			alerts = append(alerts, fmt.Sprintf("High connection count: %d (max: %d)",
				pm.metrics.ConnectionMetrics.TotalConnections, pm.config.MaxConnectionCount))
		}
	}

	// Check cache performance
	if pm.metrics.CacheMetrics != nil {
		if pm.metrics.CacheMetrics.CacheHitRate < pm.config.MinCacheHitRate {
			alerts = append(alerts, fmt.Sprintf("Low cache hit rate: %.2f%% (min: %.2f%%)",
				pm.metrics.CacheMetrics.CacheHitRate*100, pm.config.MinCacheHitRate*100))
		}
	}

	// Log alerts
	for _, alert := range alerts {
		pm.logger.Printf("ALERT: %s", alert)
	}
}

// GeneratePerformanceReport generates a comprehensive performance report
func (pm *PerformanceMonitor) GeneratePerformanceReport() ([]byte, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	report := map[string]interface{}{
		"timestamp": pm.metrics.Timestamp,
		"metrics":   pm.metrics,
		"config":    pm.config,
		"status":    "running",
	}

	return json.MarshalIndent(report, "", "  ")
}

// GetSlowQueries returns the current slow queries
func (pm *PerformanceMonitor) GetSlowQueries() []SlowQuery {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.metrics.QueryMetrics == nil {
		return []SlowQuery{}
	}

	return pm.metrics.QueryMetrics.TopSlowQueries
}

// GetPerformanceSummary returns a summary of current performance
func (pm *PerformanceMonitor) GetPerformanceSummary() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := map[string]interface{}{
		"timestamp": pm.metrics.Timestamp,
		"status":    "healthy",
	}

	if pm.metrics.QueryMetrics != nil {
		summary["queries_per_second"] = pm.metrics.QueryMetrics.QueriesPerSecond
		summary["average_query_time_ms"] = float64(pm.metrics.QueryMetrics.AverageQueryTime.Nanoseconds()) / 1e6
		summary["slow_queries"] = pm.metrics.QueryMetrics.SlowQueries
	}

	if pm.metrics.ConnectionMetrics != nil {
		summary["active_connections"] = pm.metrics.ConnectionMetrics.ActiveConnections
		summary["connection_utilization"] = pm.metrics.ConnectionMetrics.ConnectionUtilization
	}

	if pm.metrics.CacheMetrics != nil {
		summary["cache_hit_rate"] = pm.metrics.CacheMetrics.CacheHitRate
	}

	return summary
}

// Helper function to truncate query text
func truncateQuery(query string, maxLength int) string {
	if len(query) <= maxLength {
		return query
	}
	return query[:maxLength] + "..."
}
