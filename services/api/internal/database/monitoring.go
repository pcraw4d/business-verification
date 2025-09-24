package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	Timestamp         time.Time              `json:"timestamp"`
	ConnectionCount   int                    `json:"connection_count"`
	ActiveConnections int                    `json:"active_connections"`
	IdleConnections   int                    `json:"idle_connections"`
	MaxConnections    int                    `json:"max_connections"`
	QueryCount        int64                  `json:"query_count"`
	SlowQueryCount    int64                  `json:"slow_query_count"`
	ErrorCount        int64                  `json:"error_count"`
	AvgQueryTime      float64                `json:"avg_query_time_ms"`
	MaxQueryTime      float64                `json:"max_query_time_ms"`
	DatabaseSize      int64                  `json:"database_size_bytes"`
	TableSizes        map[string]int64       `json:"table_sizes"`
	IndexSizes        map[string]int64       `json:"index_sizes"`
	LockCount         int                    `json:"lock_count"`
	DeadlockCount     int                    `json:"deadlock_count"`
	CacheHitRatio     float64                `json:"cache_hit_ratio"`
	Uptime            time.Duration          `json:"uptime"`
	LastBackup        *time.Time             `json:"last_backup,omitempty"`
	BackupSize        int64                  `json:"backup_size_bytes"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// DatabaseMonitor handles database monitoring
type DatabaseMonitor struct {
	db                 *sql.DB
	config             *DatabaseConfig
	metrics            []*DatabaseMetrics
	maxMetrics         int
	slowQueryThreshold time.Duration
	mu                 sync.RWMutex
	queryStats         map[string]*QueryStats
}

// QueryStats tracks statistics for individual queries
type QueryStats struct {
	Count      int64         `json:"count"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MaxTime    time.Duration `json:"max_time"`
	MinTime    time.Duration `json:"min_time"`
	ErrorCount int64         `json:"error_count"`
	LastSeen   time.Time     `json:"last_seen"`
}

// NewDatabaseMonitor creates a new database monitor
func NewDatabaseMonitor(db *sql.DB, config *DatabaseConfig) *DatabaseMonitor {
	return &DatabaseMonitor{
		db:                 db,
		config:             config,
		metrics:            make([]*DatabaseMetrics, 0),
		maxMetrics:         1000, // Keep last 1000 metrics
		slowQueryThreshold: 100 * time.Millisecond,
		queryStats:         make(map[string]*QueryStats),
	}
}

// CollectMetrics collects current database metrics
func (m *DatabaseMonitor) CollectMetrics(ctx context.Context) (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{
		Timestamp:  time.Now(),
		TableSizes: make(map[string]int64),
		IndexSizes: make(map[string]int64),
		Metadata:   make(map[string]interface{}),
	}

	// Get connection pool stats
	stats := m.db.Stats()
	metrics.ConnectionCount = stats.OpenConnections
	metrics.MaxConnections = stats.MaxOpenConnections
	metrics.ActiveConnections = stats.InUse
	metrics.IdleConnections = stats.Idle

	// Get database size
	if size, err := m.getDatabaseSize(ctx); err == nil {
		metrics.DatabaseSize = size
	}

	// Get table sizes
	if tableSizes, err := m.getTableSizes(ctx); err == nil {
		metrics.TableSizes = tableSizes
	}

	// Get index sizes
	if indexSizes, err := m.getIndexSizes(ctx); err == nil {
		metrics.IndexSizes = indexSizes
	}

	// Get cache hit ratio
	if hitRatio, err := m.getCacheHitRatio(ctx); err == nil {
		metrics.CacheHitRatio = hitRatio
	}

	// Get lock information
	if lockCount, err := m.getLockCount(ctx); err == nil {
		metrics.LockCount = lockCount
	}

	// Get uptime
	if uptime, err := m.getUptime(ctx); err == nil {
		metrics.Uptime = uptime
	}

	// Calculate query statistics
	m.calculateQueryStats(metrics)

	// Store metrics
	m.storeMetrics(metrics)

	return metrics, nil
}

// GetMetrics returns collected metrics
func (m *DatabaseMonitor) GetMetrics(limit int) []*DatabaseMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.metrics) {
		limit = len(m.metrics)
	}

	start := len(m.metrics) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*DatabaseMetrics, limit)
	copy(result, m.metrics[start:])
	return result
}

// GetLatestMetrics returns the most recent metrics
func (m *DatabaseMonitor) GetLatestMetrics() *DatabaseMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.metrics) == 0 {
		return nil
	}

	return m.metrics[len(m.metrics)-1]
}

// GetMetricsSummary returns a summary of metrics
func (m *DatabaseMonitor) GetMetricsSummary() map[string]interface{} {
	metrics := m.GetMetrics(100) // Last 100 metrics
	if len(metrics) == 0 {
		return map[string]interface{}{
			"message": "No metrics available",
		}
	}

	// Calculate averages
	var totalConnections, totalQueries, totalErrors int64
	var totalQueryTime float64
	var slowQueries int64

	for _, metric := range metrics {
		totalConnections += int64(metric.ConnectionCount)
		totalQueries += metric.QueryCount
		totalErrors += metric.ErrorCount
		totalQueryTime += metric.AvgQueryTime
		slowQueries += metric.SlowQueryCount
	}

	count := int64(len(metrics))
	avgConnections := float64(totalConnections) / float64(count)
	avgQueries := float64(totalQueries) / float64(count)
	avgErrors := float64(totalErrors) / float64(count)
	avgQueryTime := totalQueryTime / float64(count)

	// Get latest metrics for current values
	latest := metrics[len(metrics)-1]

	summary := map[string]interface{}{
		"period": map[string]interface{}{
			"start":    metrics[0].Timestamp,
			"end":      latest.Timestamp,
			"duration": latest.Timestamp.Sub(metrics[0].Timestamp),
		},
		"averages": map[string]interface{}{
			"connections":        avgConnections,
			"queries_per_second": avgQueries,
			"errors_per_second":  avgErrors,
			"avg_query_time_ms":  avgQueryTime,
		},
		"current": map[string]interface{}{
			"connections":        latest.ConnectionCount,
			"active_connections": latest.ActiveConnections,
			"idle_connections":   latest.IdleConnections,
			"max_connections":    latest.MaxConnections,
			"database_size_mb":   float64(latest.DatabaseSize) / (1024 * 1024),
			"cache_hit_ratio":    latest.CacheHitRatio,
			"uptime":             latest.Uptime,
		},
		"totals": map[string]interface{}{
			"queries":      totalQueries,
			"errors":       totalErrors,
			"slow_queries": slowQueries,
		},
		"performance": map[string]interface{}{
			"slow_query_threshold_ms": m.slowQueryThreshold.Milliseconds(),
			"max_metrics_stored":      m.maxMetrics,
		},
	}

	return summary
}

// RecordQuery records a query execution for monitoring
func (m *DatabaseMonitor) RecordQuery(query string, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a simple hash of the query for grouping
	queryHash := fmt.Sprintf("%d", len(query)) // Simplified for demo

	stats, exists := m.queryStats[queryHash]
	if !exists {
		stats = &QueryStats{
			MinTime: duration,
		}
		m.queryStats[queryHash] = stats
	}

	stats.Count++
	stats.TotalTime += duration
	stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)
	stats.LastSeen = time.Now()

	if duration > stats.MaxTime {
		stats.MaxTime = duration
	}
	if duration < stats.MinTime {
		stats.MinTime = duration
	}

	if err != nil {
		stats.ErrorCount++
	}
}

// GetQueryStats returns query statistics
func (m *DatabaseMonitor) GetQueryStats() map[string]*QueryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*QueryStats)
	for k, v := range m.queryStats {
		result[k] = v
	}
	return result
}

// SetSlowQueryThreshold sets the threshold for slow queries
func (m *DatabaseMonitor) SetSlowQueryThreshold(threshold time.Duration) {
	m.slowQueryThreshold = threshold
}

// GetSlowQueries returns queries that exceed the slow query threshold
func (m *DatabaseMonitor) GetSlowQueries() []*QueryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var slowQueries []*QueryStats
	for _, stats := range m.queryStats {
		if stats.AvgTime > m.slowQueryThreshold {
			slowQueries = append(slowQueries, stats)
		}
	}

	return slowQueries
}

// ClearMetrics clears all stored metrics
func (m *DatabaseMonitor) ClearMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = make([]*DatabaseMetrics, 0)
	m.queryStats = make(map[string]*QueryStats)
}

// storeMetrics stores metrics with size limit
func (m *DatabaseMonitor) storeMetrics(metrics *DatabaseMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = append(m.metrics, metrics)

	// Remove old metrics if we exceed the limit
	if len(m.metrics) > m.maxMetrics {
		m.metrics = m.metrics[len(m.metrics)-m.maxMetrics:]
	}
}

// calculateQueryStats calculates query statistics for metrics
func (m *DatabaseMonitor) calculateQueryStats(metrics *DatabaseMetrics) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalQueries, totalErrors, slowQueries int64
	var totalTime time.Duration

	for _, stats := range m.queryStats {
		totalQueries += stats.Count
		totalErrors += stats.ErrorCount
		totalTime += stats.TotalTime

		if stats.AvgTime > m.slowQueryThreshold {
			slowQueries += stats.Count
		}
	}

	metrics.QueryCount = totalQueries
	metrics.ErrorCount = totalErrors
	metrics.SlowQueryCount = slowQueries

	if totalQueries > 0 {
		metrics.AvgQueryTime = float64(totalTime.Milliseconds()) / float64(totalQueries)
	}
}

// getDatabaseSize gets the database size in bytes
func (m *DatabaseMonitor) getDatabaseSize(ctx context.Context) (int64, error) {
	query := `
		SELECT pg_database_size(current_database())
	`

	var size int64
	err := m.db.QueryRowContext(ctx, query).Scan(&size)
	return size, err
}

// getTableSizes gets the size of all tables
func (m *DatabaseMonitor) getTableSizes(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_total_relation_size(schemaname||'.'||tablename) as size
		FROM pg_tables 
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY size DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sizes := make(map[string]int64)
	for rows.Next() {
		var schema, table string
		var size int64
		if err := rows.Scan(&schema, &table, &size); err != nil {
			continue
		}
		tableName := fmt.Sprintf("%s.%s", schema, table)
		sizes[tableName] = size
	}

	return sizes, nil
}

// getIndexSizes gets the size of all indexes
func (m *DatabaseMonitor) getIndexSizes(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT 
			schemaname,
			indexname,
			pg_relation_size(schemaname||'.'||indexname) as size
		FROM pg_indexes 
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY size DESC
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sizes := make(map[string]int64)
	for rows.Next() {
		var schema, index string
		var size int64
		if err := rows.Scan(&schema, &index, &size); err != nil {
			continue
		}
		indexName := fmt.Sprintf("%s.%s", schema, index)
		sizes[indexName] = size
	}

	return sizes, nil
}

// getCacheHitRatio gets the cache hit ratio
func (m *DatabaseMonitor) getCacheHitRatio(ctx context.Context) (float64, error) {
	query := `
		SELECT 
			ROUND(100.0 * sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)), 2) as cache_hit_ratio
		FROM pg_statio_user_tables
	`

	var hitRatio float64
	err := m.db.QueryRowContext(ctx, query).Scan(&hitRatio)
	return hitRatio, err
}

// getLockCount gets the number of active locks
func (m *DatabaseMonitor) getLockCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM pg_locks 
		WHERE NOT granted
	`

	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// getUptime gets the database uptime
func (m *DatabaseMonitor) getUptime(ctx context.Context) (time.Duration, error) {
	query := `
		SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time())) * interval '1 second'
	`

	var uptime time.Duration
	err := m.db.QueryRowContext(ctx, query).Scan(&uptime)
	return uptime, err
}

// StartMonitoring starts continuous monitoring
func (m *DatabaseMonitor) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting database monitoring with %v interval", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping database monitoring")
			return
		case <-ticker.C:
			if _, err := m.CollectMetrics(ctx); err != nil {
				log.Printf("Failed to collect metrics: %v", err)
			}
		}
	}
}

// GetHealthStatus returns the overall health status of the database
func (m *DatabaseMonitor) GetHealthStatus() map[string]interface{} {
	latest := m.GetLatestMetrics()
	if latest == nil {
		return map[string]interface{}{
			"status":  "unknown",
			"message": "No metrics available",
		}
	}

	// Determine health status based on metrics
	status := "healthy"
	issues := []string{}

	// Check connection pool
	if latest.ConnectionCount >= latest.MaxConnections {
		status = "critical"
		issues = append(issues, "Connection pool at maximum capacity")
	} else if latest.ConnectionCount >= latest.MaxConnections*8/10 {
		status = "warning"
		issues = append(issues, "Connection pool nearly full")
	}

	// Check error rate
	if latest.ErrorCount > 0 {
		errorRate := float64(latest.ErrorCount) / float64(latest.QueryCount) * 100
		if errorRate > 5 {
			status = "critical"
			issues = append(issues, fmt.Sprintf("High error rate: %.2f%%", errorRate))
		} else if errorRate > 1 {
			status = "warning"
			issues = append(issues, fmt.Sprintf("Elevated error rate: %.2f%%", errorRate))
		}
	}

	// Check slow queries
	if latest.SlowQueryCount > 0 {
		slowQueryRate := float64(latest.SlowQueryCount) / float64(latest.QueryCount) * 100
		if slowQueryRate > 10 {
			status = "warning"
			issues = append(issues, fmt.Sprintf("High slow query rate: %.2f%%", slowQueryRate))
		}
	}

	// Check cache hit ratio
	if latest.CacheHitRatio < 80 {
		status = "warning"
		issues = append(issues, fmt.Sprintf("Low cache hit ratio: %.2f%%", latest.CacheHitRatio))
	}

	// Check locks
	if latest.LockCount > 10 {
		status = "warning"
		issues = append(issues, fmt.Sprintf("High lock count: %d", latest.LockCount))
	}

	health := map[string]interface{}{
		"status":          status,
		"timestamp":       latest.Timestamp,
		"issues":          issues,
		"metrics":         latest,
		"recommendations": m.getRecommendations(latest),
	}

	return health
}

// getRecommendations returns recommendations based on metrics
func (m *DatabaseMonitor) getRecommendations(metrics *DatabaseMetrics) []string {
	var recommendations []string

	if metrics.ConnectionCount >= metrics.MaxConnections*8/10 {
		recommendations = append(recommendations, "Consider increasing max_connections or optimizing connection usage")
	}

	if metrics.CacheHitRatio < 80 {
		recommendations = append(recommendations, "Consider increasing shared_buffers or optimizing queries")
	}

	if metrics.SlowQueryCount > 0 {
		recommendations = append(recommendations, "Review and optimize slow queries")
	}

	if metrics.ErrorCount > 0 {
		recommendations = append(recommendations, "Investigate and fix database errors")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Database performance is good")
	}

	return recommendations
}
