package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// ResourceMonitor provides comprehensive database resource monitoring
type ResourceMonitor struct {
	db        *sql.DB
	logger    *log.Logger
	config    *ResourceMonitorConfig
	metrics   *ResourceMetrics
	mu        sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
	history   []*ResourceMetrics
	historyMu sync.RWMutex
}

// ResourceMonitorConfig contains configuration for resource monitoring
type ResourceMonitorConfig struct {
	// Monitoring intervals
	CollectionInterval time.Duration
	HistoryRetention   time.Duration
	MaxHistoryEntries  int

	// Alert thresholds
	CPUThreshold        float64
	MemoryThreshold     float64
	ConnectionThreshold int
	DiskSpaceThreshold  float64
	QueryTimeThreshold  time.Duration

	// Monitoring features
	MonitorCPU         bool
	MonitorMemory      bool
	MonitorConnections bool
	MonitorDiskSpace   bool
	MonitorLocks       bool
	MonitorQueries     bool
	MonitorIndexes     bool
	MonitorCache       bool
}

// ResourceMetrics contains comprehensive resource usage metrics
type ResourceMetrics struct {
	Timestamp         time.Time
	CPUMetrics        *CPUMetrics
	MemoryMetrics     *MemoryMetrics
	ConnectionMetrics *ResourceConnectionMetrics
	DiskSpaceMetrics  *DiskSpaceMetrics
	LockMetrics       *ResourceLockMetrics
	QueryMetrics      *ResourceQueryMetrics
	IndexMetrics      *IndexMetrics
	CacheMetrics      *ResourceCacheMetrics
	SystemMetrics     *ResourceSystemMetrics
}

// CPUMetrics contains CPU-related metrics
type CPUMetrics struct {
	UsagePercent     float64
	LoadAverage1Min  float64
	LoadAverage5Min  float64
	LoadAverage15Min float64
	ProcessCount     int
	ActiveProcesses  int
}

// MemoryMetrics contains memory-related metrics
type MemoryMetrics struct {
	TotalMemory        int64
	UsedMemory         int64
	FreeMemory         int64
	SharedBuffers      int64
	WorkMem            int64
	MaintenanceWorkMem int64
	EffectiveCacheSize int64
	MemoryUtilization  float64
}

// ResourceConnectionMetrics contains connection-related metrics
// (renamed to avoid conflict with performance_monitor.go)
type ResourceConnectionMetrics struct {
	TotalConnections      int
	ActiveConnections     int
	IdleConnections       int
	MaxConnections        int
	ConnectionUtilization float64
	LongRunningQueries    int
	BlockedQueries        int
}

// DiskSpaceMetrics contains disk space-related metrics
type DiskSpaceMetrics struct {
	DatabaseSize    int64
	TableSizes      map[string]int64
	IndexSizes      map[string]int64
	TotalDiskSpace  int64
	FreeDiskSpace   int64
	DiskUtilization float64
	GrowthRate      float64 // MB per hour
}

// ResourceLockMetrics contains lock-related metrics
// (renamed to avoid conflict with performance_monitor.go)
type ResourceLockMetrics struct {
	LockWaits       int
	Deadlocks       int
	LockTimeouts    int
	AverageWaitTime time.Duration
	MaxWaitTime     time.Duration
	LockedTables    int
	BlockingQueries int
}

// ResourceQueryMetrics contains query-related metrics
// (renamed to avoid conflict with performance_monitor.go)
type ResourceQueryMetrics struct {
	TotalQueries     int64
	SlowQueries      int64
	AverageQueryTime time.Duration
	MaxQueryTime     time.Duration
	QueriesPerSecond float64
	ErrorRate        float64
	CacheHitRate     float64
	IndexHitRate     float64
}

// IndexMetrics contains index-related metrics
type IndexMetrics struct {
	TotalIndexes     int
	UnusedIndexes    int
	DuplicateIndexes int
	IndexBloat       float64
	IndexScanCount   int64
	IndexSize        int64
	IndexUtilization float64
}

// ResourceCacheMetrics contains cache-related metrics
// (renamed to avoid conflict with performance_monitor.go and redis_cache.go)
type ResourceCacheMetrics struct {
	SharedBuffers     int64
	SharedBuffersUsed int64
	SharedBuffersHit  int64
	SharedBuffersRead int64
	CacheHitRate      float64
	BufferHitRate     float64
	CacheSize         int64
	CacheUtilization  float64
}

// ResourceSystemMetrics contains system-level metrics
// (renamed to avoid conflict with performance_monitor.go)
type ResourceSystemMetrics struct {
	Uptime        time.Duration
	Version       string
	Configuration map[string]string
	Extensions    []string
	DatabaseCount int
	TableCount    int
	IndexCount    int
	FunctionCount int
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(db *sql.DB, config *ResourceMonitorConfig) *ResourceMonitor {
	if config == nil {
		config = &ResourceMonitorConfig{
			CollectionInterval:  30 * time.Second,
			HistoryRetention:    24 * time.Hour,
			MaxHistoryEntries:   1000,
			CPUThreshold:        80.0,
			MemoryThreshold:     85.0,
			ConnectionThreshold: 80,
			DiskSpaceThreshold:  90.0,
			QueryTimeThreshold:  1 * time.Second,
			MonitorCPU:          true,
			MonitorMemory:       true,
			MonitorConnections:  true,
			MonitorDiskSpace:    true,
			MonitorLocks:        true,
			MonitorQueries:      true,
			MonitorIndexes:      true,
			MonitorCache:        true,
		}
	}

	return &ResourceMonitor{
		db:       db,
		logger:   log.New(log.Writer(), "[RESOURCE_MONITOR] ", log.LstdFlags),
		config:   config,
		metrics:  &ResourceMetrics{},
		stopChan: make(chan struct{}),
		history:  make([]*ResourceMetrics, 0),
	}
}

// Start begins resource monitoring
func (rm *ResourceMonitor) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.isRunning {
		return fmt.Errorf("resource monitor is already running")
	}

	rm.isRunning = true
	rm.logger.Println("Starting resource monitoring...")

	// Start monitoring goroutine
	go rm.monitoringLoop(ctx)

	return nil
}

// Stop stops resource monitoring
func (rm *ResourceMonitor) Stop() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.isRunning {
		return
	}

	rm.logger.Println("Stopping resource monitoring...")
	close(rm.stopChan)
	rm.isRunning = false
}

// GetCurrentMetrics returns the current resource metrics
func (rm *ResourceMonitor) GetCurrentMetrics() *ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy of the metrics
	metricsCopy := *rm.metrics
	return &metricsCopy
}

// GetMetricsHistory returns historical resource metrics
func (rm *ResourceMonitor) GetMetricsHistory(limit int) []*ResourceMetrics {
	rm.historyMu.RLock()
	defer rm.historyMu.RUnlock()

	if limit <= 0 || limit > len(rm.history) {
		limit = len(rm.history)
	}

	// Return a copy of the history
	historyCopy := make([]*ResourceMetrics, limit)
	copy(historyCopy, rm.history[len(rm.history)-limit:])

	return historyCopy
}

// GetResourceTrends returns resource usage trends
func (rm *ResourceMonitor) GetResourceTrends(duration time.Duration) (*ResourceTrends, error) {
	rm.historyMu.RLock()
	defer rm.historyMu.RUnlock()

	if len(rm.history) < 2 {
		return nil, fmt.Errorf("insufficient history data for trend analysis")
	}

	cutoff := time.Now().Add(-duration)
	var relevantMetrics []*ResourceMetrics

	for _, metrics := range rm.history {
		if metrics.Timestamp.After(cutoff) {
			relevantMetrics = append(relevantMetrics, metrics)
		}
	}

	if len(relevantMetrics) < 2 {
		return nil, fmt.Errorf("insufficient data points for trend analysis")
	}

	return rm.calculateTrends(relevantMetrics), nil
}

// ResourceTrends contains trend analysis results
type ResourceTrends struct {
	Duration          time.Duration
	DataPoints        int
	CPUTrend          *TrendData
	MemoryTrend       *TrendData
	ConnectionTrend   *TrendData
	DiskSpaceTrend    *TrendData
	QueryTimeTrend    *TrendData
	CacheHitRateTrend *TrendData
}

// TrendData contains trend information for a specific metric
type TrendData struct {
	StartValue    float64
	EndValue      float64
	Change        float64
	ChangePercent float64
	Trend         string // "increasing", "decreasing", "stable"
	Volatility    float64
	PeakValue     float64
	MinValue      float64
}

// monitoringLoop runs the main monitoring loop
func (rm *ResourceMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rm.logger.Println("Resource monitoring stopped due to context cancellation")
			return
		case <-rm.stopChan:
			rm.logger.Println("Resource monitoring stopped")
			return
		case <-ticker.C:
			if err := rm.collectMetrics(ctx); err != nil {
				rm.logger.Printf("Failed to collect resource metrics: %v", err)
			}
		}
	}
}

// collectMetrics collects all resource metrics
func (rm *ResourceMonitor) collectMetrics(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.metrics.Timestamp = time.Now()

	// Collect CPU metrics
	if rm.config.MonitorCPU {
		if err := rm.collectCPUMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect CPU metrics: %v", err)
		}
	}

	// Collect memory metrics
	if rm.config.MonitorMemory {
		if err := rm.collectMemoryMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect memory metrics: %v", err)
		}
	}

	// Collect connection metrics
	if rm.config.MonitorConnections {
		if err := rm.collectConnectionMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect connection metrics: %v", err)
		}
	}

	// Collect disk space metrics
	if rm.config.MonitorDiskSpace {
		if err := rm.collectDiskSpaceMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect disk space metrics: %v", err)
		}
	}

	// Collect lock metrics
	if rm.config.MonitorLocks {
		if err := rm.collectLockMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect lock metrics: %v", err)
		}
	}

	// Collect query metrics
	if rm.config.MonitorQueries {
		if err := rm.collectQueryMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect query metrics: %v", err)
		}
	}

	// Collect index metrics
	if rm.config.MonitorIndexes {
		if err := rm.collectIndexMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect index metrics: %v", err)
		}
	}

	// Collect cache metrics
	if rm.config.MonitorCache {
		if err := rm.collectCacheMetrics(ctx); err != nil {
			rm.logger.Printf("Failed to collect cache metrics: %v", err)
		}
	}

	// Collect system metrics
	if err := rm.collectSystemMetrics(ctx); err != nil {
		rm.logger.Printf("Failed to collect system metrics: %v", err)
	}

	// Store metrics in history
	rm.storeMetricsInHistory()

	// Check for alerts
	rm.checkResourceAlerts()

	return nil
}

// collectCPUMetrics collects CPU-related metrics
func (rm *ResourceMonitor) collectCPUMetrics(ctx context.Context) error {
	rm.metrics.CPUMetrics = &CPUMetrics{}

	// Get load average (PostgreSQL doesn't directly provide this, so we'll use system info)
	// For now, we'll use a placeholder approach
	query := `
		SELECT 
			count(*) as process_count,
			count(*) FILTER (WHERE state = 'active') as active_processes
		FROM pg_stat_activity
	`

	var processCount, activeProcesses int
	err := rm.db.QueryRowContext(ctx, query).Scan(&processCount, &activeProcesses)
	if err != nil {
		return fmt.Errorf("failed to get process metrics: %w", err)
	}

	rm.metrics.CPUMetrics.ProcessCount = processCount
	rm.metrics.CPUMetrics.ActiveProcesses = activeProcesses

	// Calculate CPU usage based on active processes (simplified)
	if processCount > 0 {
		rm.metrics.CPUMetrics.UsagePercent = float64(activeProcesses) / float64(processCount) * 100
	}

	// Load averages would typically come from system monitoring
	// For now, we'll use placeholder values
	rm.metrics.CPUMetrics.LoadAverage1Min = 0.5
	rm.metrics.CPUMetrics.LoadAverage5Min = 0.6
	rm.metrics.CPUMetrics.LoadAverage15Min = 0.7

	return nil
}

// collectMemoryMetrics collects memory-related metrics
func (rm *ResourceMonitor) collectMemoryMetrics(ctx context.Context) error {
	rm.metrics.MemoryMetrics = &MemoryMetrics{}

	// Get shared buffers information
	query := `
		SELECT 
			setting::bigint as shared_buffers,
			(SELECT setting::bigint FROM pg_settings WHERE name = 'work_mem') as work_mem,
			(SELECT setting::bigint FROM pg_settings WHERE name = 'maintenance_work_mem') as maintenance_work_mem,
			(SELECT setting::bigint FROM pg_settings WHERE name = 'effective_cache_size') as effective_cache_size
		FROM pg_settings 
		WHERE name = 'shared_buffers'
	`

	var sharedBuffers, workMem, maintenanceWorkMem, effectiveCacheSize int64
	err := rm.db.QueryRowContext(ctx, query).Scan(&sharedBuffers, &workMem, &maintenanceWorkMem, &effectiveCacheSize)
	if err != nil {
		return fmt.Errorf("failed to get memory settings: %w", err)
	}

	rm.metrics.MemoryMetrics.SharedBuffers = sharedBuffers
	rm.metrics.MemoryMetrics.WorkMem = workMem
	rm.metrics.MemoryMetrics.MaintenanceWorkMem = maintenanceWorkMem
	rm.metrics.MemoryMetrics.EffectiveCacheSize = effectiveCacheSize

	// Get buffer cache statistics
	bufferQuery := `
		SELECT 
			round(
				(sum(blks_hit) * 100.0 / (sum(blks_hit) + sum(blks_read))), 2
			) as buffer_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	var bufferHitRatio float64
	err = rm.db.QueryRowContext(ctx, bufferQuery).Scan(&bufferHitRatio)
	if err != nil {
		rm.logger.Printf("Failed to get buffer hit ratio: %v", err)
	} else {
		// Estimate used memory based on hit ratio
		rm.metrics.MemoryMetrics.UsedMemory = int64(float64(sharedBuffers) * (bufferHitRatio / 100.0))
		rm.metrics.MemoryMetrics.FreeMemory = sharedBuffers - rm.metrics.MemoryMetrics.UsedMemory
		rm.metrics.MemoryMetrics.MemoryUtilization = bufferHitRatio
	}

	// Total memory would typically come from system monitoring
	rm.metrics.MemoryMetrics.TotalMemory = sharedBuffers * 2 // Simplified estimate

	return nil
}

// collectConnectionMetrics collects connection-related metrics
func (rm *ResourceMonitor) collectConnectionMetrics(ctx context.Context) error {
	rm.metrics.ConnectionMetrics = &ResourceConnectionMetrics{}

	// Get connection statistics
	query := `
		SELECT 
			count(*) as total_connections,
			count(*) FILTER (WHERE state = 'active') as active_connections,
			count(*) FILTER (WHERE state = 'idle') as idle_connections,
			count(*) FILTER (WHERE state = 'active' AND query_start < now() - interval '5 minutes') as long_running_queries,
			count(*) FILTER (WHERE wait_event_type = 'Lock') as blocked_queries
		FROM pg_stat_activity
	`

	var total, active, idle, longRunning, blocked int
	err := rm.db.QueryRowContext(ctx, query).Scan(&total, &active, &idle, &longRunning, &blocked)
	if err != nil {
		return fmt.Errorf("failed to get connection metrics: %w", err)
	}

	rm.metrics.ConnectionMetrics.TotalConnections = total
	rm.metrics.ConnectionMetrics.ActiveConnections = active
	rm.metrics.ConnectionMetrics.IdleConnections = idle
	rm.metrics.ConnectionMetrics.LongRunningQueries = longRunning
	rm.metrics.ConnectionMetrics.BlockedQueries = blocked

	// Get max connections setting
	var maxConnections int
	err = rm.db.QueryRowContext(ctx, "SHOW max_connections").Scan(&maxConnections)
	if err != nil {
		rm.logger.Printf("Failed to get max_connections setting: %v", err)
		maxConnections = 100 // Default fallback
	}

	rm.metrics.ConnectionMetrics.MaxConnections = maxConnections
	rm.metrics.ConnectionMetrics.ConnectionUtilization = float64(total) / float64(maxConnections) * 100

	return nil
}

// collectDiskSpaceMetrics collects disk space-related metrics
func (rm *ResourceMonitor) collectDiskSpaceMetrics(ctx context.Context) error {
	rm.metrics.DiskSpaceMetrics = &DiskSpaceMetrics{
		TableSizes: make(map[string]int64),
		IndexSizes: make(map[string]int64),
	}

	// Get database size
	var dbSize int64
	err := rm.db.QueryRowContext(ctx, "SELECT pg_database_size(current_database())").Scan(&dbSize)
	if err != nil {
		return fmt.Errorf("failed to get database size: %w", err)
	}
	rm.metrics.DiskSpaceMetrics.DatabaseSize = dbSize

	// Get table sizes
	tableQuery := `
		SELECT 
			schemaname,
			tablename,
			pg_total_relation_size(schemaname||'.'||tablename) as size
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
		LIMIT 20
	`

	rows, err := rm.db.QueryContext(ctx, tableQuery)
	if err != nil {
		rm.logger.Printf("Failed to get table sizes: %v", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var schema, table string
			var size int64
			if err := rows.Scan(&schema, &table, &size); err == nil {
				rm.metrics.DiskSpaceMetrics.TableSizes[table] = size
			}
		}
	}

	// Get index sizes
	indexQuery := `
		SELECT 
			indexname,
			pg_relation_size(indexname::regclass) as size
		FROM pg_indexes 
		WHERE schemaname = 'public'
		ORDER BY pg_relation_size(indexname::regclass) DESC
		LIMIT 20
	`

	rows, err = rm.db.QueryContext(ctx, indexQuery)
	if err != nil {
		rm.logger.Printf("Failed to get index sizes: %v", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var index string
			var size int64
			if err := rows.Scan(&index, &size); err == nil {
				rm.metrics.DiskSpaceMetrics.IndexSizes[index] = size
			}
		}
	}

	// Calculate total disk space (simplified)
	rm.metrics.DiskSpaceMetrics.TotalDiskSpace = dbSize * 2 // Simplified estimate
	rm.metrics.DiskSpaceMetrics.FreeDiskSpace = rm.metrics.DiskSpaceMetrics.TotalDiskSpace - dbSize
	rm.metrics.DiskSpaceMetrics.DiskUtilization = float64(dbSize) / float64(rm.metrics.DiskSpaceMetrics.TotalDiskSpace) * 100

	return nil
}

// collectLockMetrics collects lock-related metrics
func (rm *ResourceMonitor) collectLockMetrics(ctx context.Context) error {
	rm.metrics.LockMetrics = &ResourceLockMetrics{}

	// Get lock statistics
	query := `
		SELECT 
			count(*) FILTER (WHERE wait_event_type = 'Lock') as lock_waits,
			count(*) FILTER (WHERE wait_event = 'deadlock_detection') as deadlocks,
			count(*) FILTER (WHERE wait_event_type = 'Lock' AND wait_event = 'lock_timeout') as lock_timeouts,
			count(DISTINCT relation) FILTER (WHERE wait_event_type = 'Lock') as locked_tables,
			count(*) FILTER (WHERE wait_event_type = 'Lock' AND state = 'active') as blocking_queries
		FROM pg_stat_activity
	`

	var lockWaits, deadlocks, lockTimeouts, lockedTables, blockingQueries int
	err := rm.db.QueryRowContext(ctx, query).Scan(&lockWaits, &deadlocks, &lockTimeouts, &lockedTables, &blockingQueries)
	if err != nil {
		return fmt.Errorf("failed to get lock metrics: %w", err)
	}

	rm.metrics.LockMetrics.LockWaits = lockWaits
	rm.metrics.LockMetrics.Deadlocks = deadlocks
	rm.metrics.LockMetrics.LockTimeouts = lockTimeouts
	rm.metrics.LockMetrics.LockedTables = lockedTables
	rm.metrics.LockMetrics.BlockingQueries = blockingQueries

	// Get average wait time (simplified)
	if lockWaits > 0 {
		rm.metrics.LockMetrics.AverageWaitTime = time.Duration(lockWaits) * 100 * time.Millisecond // Simplified
		rm.metrics.LockMetrics.MaxWaitTime = time.Duration(lockWaits) * 500 * time.Millisecond     // Simplified
	}

	return nil
}

// collectQueryMetrics collects query-related metrics
func (rm *ResourceMonitor) collectQueryMetrics(ctx context.Context) error {
	rm.metrics.QueryMetrics = &ResourceQueryMetrics{}

	// Get query statistics from pg_stat_statements
	query := `
		SELECT 
			sum(calls) as total_queries,
			sum(calls) FILTER (WHERE mean_time > $1) as slow_queries,
			avg(mean_time) as avg_query_time,
			max(mean_time) as max_query_time,
			sum(calls) / extract(epoch from (now() - stats_reset)) as queries_per_second
		FROM pg_stat_statements
	`

	var totalQueries, slowQueries int64
	var avgQueryTime, maxQueryTime, queriesPerSecond float64
	err := rm.db.QueryRowContext(ctx, query, rm.config.QueryTimeThreshold.Milliseconds()).Scan(
		&totalQueries, &slowQueries, &avgQueryTime, &maxQueryTime, &queriesPerSecond)
	if err != nil {
		rm.logger.Printf("Failed to get query metrics: %v", err)
		// Set default values
		rm.metrics.QueryMetrics.TotalQueries = 0
		rm.metrics.QueryMetrics.SlowQueries = 0
		rm.metrics.QueryMetrics.AverageQueryTime = 0
		rm.metrics.QueryMetrics.MaxQueryTime = 0
		rm.metrics.QueryMetrics.QueriesPerSecond = 0
		rm.metrics.QueryMetrics.ErrorRate = 0
		return nil
	}

	rm.metrics.QueryMetrics.TotalQueries = totalQueries
	rm.metrics.QueryMetrics.SlowQueries = slowQueries
	rm.metrics.QueryMetrics.AverageQueryTime = time.Duration(avgQueryTime) * time.Millisecond
	rm.metrics.QueryMetrics.MaxQueryTime = time.Duration(maxQueryTime) * time.Millisecond
	rm.metrics.QueryMetrics.QueriesPerSecond = queriesPerSecond

	// Calculate error rate (simplified)
	if totalQueries > 0 {
		rm.metrics.QueryMetrics.ErrorRate = float64(slowQueries) / float64(totalQueries) * 100
	}

	// Get cache hit rate
	cacheQuery := `
		SELECT 
			round(
				(sum(blks_hit) * 100.0 / (sum(blks_hit) + sum(blks_read))), 2
			) as cache_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	var cacheHitRatio float64
	err = rm.db.QueryRowContext(ctx, cacheQuery).Scan(&cacheHitRatio)
	if err != nil {
		rm.logger.Printf("Failed to get cache hit ratio: %v", err)
	} else {
		rm.metrics.QueryMetrics.CacheHitRate = cacheHitRatio
	}

	// Get index hit rate (simplified)
	rm.metrics.QueryMetrics.IndexHitRate = cacheHitRatio * 0.95 // Simplified estimate

	return nil
}

// collectIndexMetrics collects index-related metrics
func (rm *ResourceMonitor) collectIndexMetrics(ctx context.Context) error {
	rm.metrics.IndexMetrics = &IndexMetrics{}

	// Get index statistics
	query := `
		SELECT 
			count(*) as total_indexes,
			count(*) FILTER (WHERE idx_tup_read = 0) as unused_indexes,
			sum(pg_relation_size(indexrelid)) as index_size
		FROM pg_stat_user_indexes
	`

	var totalIndexes, unusedIndexes int
	var indexSize int64
	err := rm.db.QueryRowContext(ctx, query).Scan(&totalIndexes, &unusedIndexes, &indexSize)
	if err != nil {
		return fmt.Errorf("failed to get index metrics: %w", err)
	}

	rm.metrics.IndexMetrics.TotalIndexes = totalIndexes
	rm.metrics.IndexMetrics.UnusedIndexes = unusedIndexes
	rm.metrics.IndexMetrics.IndexSize = indexSize

	// Calculate index utilization
	if totalIndexes > 0 {
		rm.metrics.IndexMetrics.IndexUtilization = float64(totalIndexes-unusedIndexes) / float64(totalIndexes) * 100
	}

	// Get index scan count
	var indexScanCount int64
	err = rm.db.QueryRowContext(ctx, "SELECT sum(idx_tup_read) FROM pg_stat_user_indexes").Scan(&indexScanCount)
	if err != nil {
		rm.logger.Printf("Failed to get index scan count: %v", err)
	} else {
		rm.metrics.IndexMetrics.IndexScanCount = indexScanCount
	}

	// Calculate index bloat (simplified)
	rm.metrics.IndexMetrics.IndexBloat = float64(unusedIndexes) / float64(totalIndexes) * 100

	return nil
}

// collectCacheMetrics collects cache-related metrics
func (rm *ResourceMonitor) collectCacheMetrics(ctx context.Context) error {
	rm.metrics.CacheMetrics = &ResourceCacheMetrics{}

	// Get shared buffer statistics
	query := `
		SELECT 
			setting::bigint as shared_buffers,
			(SELECT sum(blks_hit) FROM pg_stat_database) as shared_buffers_hit,
			(SELECT sum(blks_read) FROM pg_stat_database) as shared_buffers_read
		FROM pg_settings 
		WHERE name = 'shared_buffers'
	`

	var sharedBuffers, sharedBuffersHit, sharedBuffersRead int64
	err := rm.db.QueryRowContext(ctx, query).Scan(&sharedBuffers, &sharedBuffersHit, &sharedBuffersRead)
	if err != nil {
		return fmt.Errorf("failed to get cache metrics: %w", err)
	}

	rm.metrics.CacheMetrics.SharedBuffers = sharedBuffers
	rm.metrics.CacheMetrics.SharedBuffersHit = sharedBuffersHit
	rm.metrics.CacheMetrics.SharedBuffersRead = sharedBuffersRead

	// Calculate cache hit rate
	if sharedBuffersHit+sharedBuffersRead > 0 {
		rm.metrics.CacheMetrics.CacheHitRate = float64(sharedBuffersHit) / float64(sharedBuffersHit+sharedBuffersRead) * 100
		rm.metrics.CacheMetrics.BufferHitRate = rm.metrics.CacheMetrics.CacheHitRate
	}

	// Estimate cache utilization
	rm.metrics.CacheMetrics.CacheSize = sharedBuffers
	rm.metrics.CacheMetrics.CacheUtilization = rm.metrics.CacheMetrics.CacheHitRate / 100.0

	return nil
}

// collectSystemMetrics collects system-level metrics
func (rm *ResourceMonitor) collectSystemMetrics(ctx context.Context) error {
	rm.metrics.SystemMetrics = &ResourceSystemMetrics{
		Configuration: make(map[string]string),
	}

	// Get database version
	var version string
	err := rm.db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		rm.logger.Printf("Failed to get database version: %v", err)
	} else {
		rm.metrics.SystemMetrics.Version = version
	}

	// Get uptime
	var uptimeSeconds int64
	err = rm.db.QueryRowContext(ctx, "SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))").Scan(&uptimeSeconds)
	if err != nil {
		rm.logger.Printf("Failed to get uptime: %v", err)
	} else {
		rm.metrics.SystemMetrics.Uptime = time.Duration(uptimeSeconds) * time.Second
	}

	// Get database count
	var dbCount int
	err = rm.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_database").Scan(&dbCount)
	if err != nil {
		rm.logger.Printf("Failed to get database count: %v", err)
	} else {
		rm.metrics.SystemMetrics.DatabaseCount = dbCount
	}

	// Get table count
	var tableCount int
	err = rm.db.QueryRowContext(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		rm.logger.Printf("Failed to get table count: %v", err)
	} else {
		rm.metrics.SystemMetrics.TableCount = tableCount
	}

	// Get index count
	var indexCount int
	err = rm.db.QueryRowContext(ctx, "SELECT count(*) FROM pg_indexes WHERE schemaname = 'public'").Scan(&indexCount)
	if err != nil {
		rm.logger.Printf("Failed to get index count: %v", err)
	} else {
		rm.metrics.SystemMetrics.IndexCount = indexCount
	}

	// Get function count
	var functionCount int
	err = rm.db.QueryRowContext(ctx, "SELECT count(*) FROM information_schema.routines WHERE routine_schema = 'public'").Scan(&functionCount)
	if err != nil {
		rm.logger.Printf("Failed to get function count: %v", err)
	} else {
		rm.metrics.SystemMetrics.FunctionCount = functionCount
	}

	return nil
}

// storeMetricsInHistory stores metrics in history
func (rm *ResourceMonitor) storeMetricsInHistory() {
	rm.historyMu.Lock()
	defer rm.historyMu.Unlock()

	// Create a copy of the current metrics
	metricsCopy := *rm.metrics

	// Add to history
	rm.history = append(rm.history, &metricsCopy)

	// Trim history if it exceeds max entries
	if len(rm.history) > rm.config.MaxHistoryEntries {
		rm.history = rm.history[len(rm.history)-rm.config.MaxHistoryEntries:]
	}

	// Remove old entries based on retention period
	cutoff := time.Now().Add(-rm.config.HistoryRetention)
	var filteredHistory []*ResourceMetrics
	for _, metrics := range rm.history {
		if metrics.Timestamp.After(cutoff) {
			filteredHistory = append(filteredHistory, metrics)
		}
	}
	rm.history = filteredHistory
}

// checkResourceAlerts checks for resource usage alerts
func (rm *ResourceMonitor) checkResourceAlerts() {
	alerts := []string{}

	// Check CPU usage
	if rm.metrics.CPUMetrics != nil && rm.metrics.CPUMetrics.UsagePercent > rm.config.CPUThreshold {
		alerts = append(alerts, fmt.Sprintf("High CPU usage: %.2f%% (threshold: %.2f%%)",
			rm.metrics.CPUMetrics.UsagePercent, rm.config.CPUThreshold))
	}

	// Check memory usage
	if rm.metrics.MemoryMetrics != nil && rm.metrics.MemoryMetrics.MemoryUtilization > rm.config.MemoryThreshold {
		alerts = append(alerts, fmt.Sprintf("High memory usage: %.2f%% (threshold: %.2f%%)",
			rm.metrics.MemoryMetrics.MemoryUtilization, rm.config.MemoryThreshold))
	}

	// Check connection usage
	if rm.metrics.ConnectionMetrics != nil && rm.metrics.ConnectionMetrics.TotalConnections > rm.config.ConnectionThreshold {
		alerts = append(alerts, fmt.Sprintf("High connection count: %d (threshold: %d)",
			rm.metrics.ConnectionMetrics.TotalConnections, rm.config.ConnectionThreshold))
	}

	// Check disk space usage
	if rm.metrics.DiskSpaceMetrics != nil && rm.metrics.DiskSpaceMetrics.DiskUtilization > rm.config.DiskSpaceThreshold {
		alerts = append(alerts, fmt.Sprintf("High disk usage: %.2f%% (threshold: %.2f%%)",
			rm.metrics.DiskSpaceMetrics.DiskUtilization, rm.config.DiskSpaceThreshold))
	}

	// Check query performance
	if rm.metrics.QueryMetrics != nil && rm.metrics.QueryMetrics.AverageQueryTime > rm.config.QueryTimeThreshold {
		alerts = append(alerts, fmt.Sprintf("Slow average query time: %.2fms (threshold: %.2fms)",
			float64(rm.metrics.QueryMetrics.AverageQueryTime.Nanoseconds())/1e6,
			float64(rm.config.QueryTimeThreshold.Nanoseconds())/1e6))
	}

	// Log alerts
	for _, alert := range alerts {
		rm.logger.Printf("ALERT: %s", alert)
	}
}

// calculateTrends calculates resource usage trends
func (rm *ResourceMonitor) calculateTrends(metrics []*ResourceMetrics) *ResourceTrends {
	if len(metrics) < 2 {
		return nil
	}

	trends := &ResourceTrends{
		Duration:   metrics[len(metrics)-1].Timestamp.Sub(metrics[0].Timestamp),
		DataPoints: len(metrics),
	}

	// Calculate CPU trend
	if metrics[0].CPUMetrics != nil && metrics[len(metrics)-1].CPUMetrics != nil {
		trends.CPUTrend = rm.calculateTrendData(
			metrics[0].CPUMetrics.UsagePercent,
			metrics[len(metrics)-1].CPUMetrics.UsagePercent,
			extractCPUMetrics(metrics),
		)
	}

	// Calculate memory trend
	if metrics[0].MemoryMetrics != nil && metrics[len(metrics)-1].MemoryMetrics != nil {
		trends.MemoryTrend = rm.calculateTrendData(
			metrics[0].MemoryMetrics.MemoryUtilization,
			metrics[len(metrics)-1].MemoryMetrics.MemoryUtilization,
			extractMemoryMetrics(metrics),
		)
	}

	// Calculate connection trend
	if metrics[0].ConnectionMetrics != nil && metrics[len(metrics)-1].ConnectionMetrics != nil {
		trends.ConnectionTrend = rm.calculateTrendData(
			float64(metrics[0].ConnectionMetrics.TotalConnections),
			float64(metrics[len(metrics)-1].ConnectionMetrics.TotalConnections),
			extractConnectionMetrics(metrics),
		)
	}

	// Calculate disk space trend
	if metrics[0].DiskSpaceMetrics != nil && metrics[len(metrics)-1].DiskSpaceMetrics != nil {
		trends.DiskSpaceTrend = rm.calculateTrendData(
			metrics[0].DiskSpaceMetrics.DiskUtilization,
			metrics[len(metrics)-1].DiskSpaceMetrics.DiskUtilization,
			extractDiskSpaceMetrics(metrics),
		)
	}

	// Calculate query time trend
	if metrics[0].QueryMetrics != nil && metrics[len(metrics)-1].QueryMetrics != nil {
		trends.QueryTimeTrend = rm.calculateTrendData(
			float64(metrics[0].QueryMetrics.AverageQueryTime.Nanoseconds())/1e6,
			float64(metrics[len(metrics)-1].QueryMetrics.AverageQueryTime.Nanoseconds())/1e6,
			extractQueryTimeMetrics(metrics),
		)
	}

	// Calculate cache hit rate trend
	if metrics[0].CacheMetrics != nil && metrics[len(metrics)-1].CacheMetrics != nil {
		trends.CacheHitRateTrend = rm.calculateTrendData(
			metrics[0].CacheMetrics.CacheHitRate,
			metrics[len(metrics)-1].CacheMetrics.CacheHitRate,
			extractCacheHitRateMetrics(metrics),
		)
	}

	return trends
}

// calculateTrendData calculates trend data for a specific metric
func (rm *ResourceMonitor) calculateTrendData(startValue, endValue float64, values []float64) *TrendData {
	if len(values) == 0 {
		return nil
	}

	change := endValue - startValue
	changePercent := 0.0
	if startValue != 0 {
		changePercent = (change / startValue) * 100
	}

	// Determine trend
	trend := "stable"
	if changePercent > 5.0 {
		trend = "increasing"
	} else if changePercent < -5.0 {
		trend = "decreasing"
	}

	// Calculate volatility (standard deviation)
	volatility := rm.calculateVolatility(values)

	// Find peak and min values
	peak := values[0]
	min := values[0]
	for _, v := range values {
		if v > peak {
			peak = v
		}
		if v < min {
			min = v
		}
	}

	return &TrendData{
		StartValue:    startValue,
		EndValue:      endValue,
		Change:        change,
		ChangePercent: changePercent,
		Trend:         trend,
		Volatility:    volatility,
		PeakValue:     peak,
		MinValue:      min,
	}
}

// calculateVolatility calculates the volatility (standard deviation) of values
func (rm *ResourceMonitor) calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values) - 1)

	// Return standard deviation
	return variance
}

// Helper functions to extract metrics from history
func extractCPUMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.CPUMetrics != nil {
			values = append(values, m.CPUMetrics.UsagePercent)
		}
	}
	return values
}

func extractMemoryMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.MemoryMetrics != nil {
			values = append(values, m.MemoryMetrics.MemoryUtilization)
		}
	}
	return values
}

func extractConnectionMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.ConnectionMetrics != nil {
			values = append(values, float64(m.ConnectionMetrics.TotalConnections))
		}
	}
	return values
}

func extractDiskSpaceMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.DiskSpaceMetrics != nil {
			values = append(values, m.DiskSpaceMetrics.DiskUtilization)
		}
	}
	return values
}

func extractQueryTimeMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.QueryMetrics != nil {
			values = append(values, float64(m.QueryMetrics.AverageQueryTime.Nanoseconds())/1e6)
		}
	}
	return values
}

func extractCacheHitRateMetrics(metrics []*ResourceMetrics) []float64 {
	var values []float64
	for _, m := range metrics {
		if m.CacheMetrics != nil {
			values = append(values, m.CacheMetrics.CacheHitRate)
		}
	}
	return values
}
