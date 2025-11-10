package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"kyb-platform/internal/monitoring"

	"github.com/google/uuid"
)

// UnifiedDatabaseMonitor handles database monitoring using the unified monitoring system
type UnifiedDatabaseMonitor struct {
	db                *sql.DB
	config            *DatabaseConfig
	monitoringAdapter *monitoring.MonitoringAdapter
	mu                sync.RWMutex
	queryStats        map[string]*QueryStats
}

// NewUnifiedDatabaseMonitor creates a new unified database monitor
func NewUnifiedDatabaseMonitor(db *sql.DB, config *DatabaseConfig, logger *log.Logger) *UnifiedDatabaseMonitor {
	return &UnifiedDatabaseMonitor{
		db:                db,
		config:            config,
		monitoringAdapter: monitoring.NewMonitoringAdapter(db, logger),
		queryStats:        make(map[string]*QueryStats),
	}
}

// CollectMetrics collects current database metrics and stores them in the unified monitoring system
func (udm *UnifiedDatabaseMonitor) CollectMetrics(ctx context.Context) (*DatabaseMetrics, error) {
	metrics := &DatabaseMetrics{
		Timestamp:  time.Now(),
		TableSizes: make(map[string]int64),
		IndexSizes: make(map[string]int64),
		Metadata:   make(map[string]interface{}),
	}

	// Get connection pool stats
	stats := udm.db.Stats()
	metrics.ConnectionCount = stats.OpenConnections
	metrics.MaxConnections = stats.MaxOpenConnections
	metrics.ActiveConnections = stats.InUse
	metrics.IdleConnections = stats.Idle

	// Get database size
	if size, err := udm.getDatabaseSize(ctx); err == nil {
		metrics.DatabaseSize = size
	}

	// Get table sizes
	if tableSizes, err := udm.getTableSizes(ctx); err == nil {
		metrics.TableSizes = tableSizes
	}

	// Get index sizes
	if indexSizes, err := udm.getIndexSizes(ctx); err == nil {
		metrics.IndexSizes = indexSizes
	}

	// Get cache hit ratio
	if hitRatio, err := udm.getCacheHitRatio(ctx); err == nil {
		metrics.CacheHitRatio = hitRatio
	}

	// Get lock information
	if lockCount, err := udm.getLockCount(ctx); err == nil {
		metrics.LockCount = lockCount
	}

	// Get uptime
	if uptime, err := udm.getUptime(ctx); err == nil {
		metrics.Uptime = uptime
	}

	// Calculate query statistics
	udm.calculateQueryStats(metrics)

	// Store metrics in unified monitoring system
	// Convert database.DatabaseMetrics to monitoring.DatabaseMetrics
	monitoringMetrics := &monitoring.DatabaseMetrics{
		Timestamp:         metrics.Timestamp,
		ConnectionCount:   metrics.ConnectionCount,
		ActiveConnections: metrics.ActiveConnections,
		IdleConnections:   metrics.IdleConnections,
		MaxConnections:    metrics.MaxConnections,
		QueryCount:        metrics.QueryCount,
		SlowQueryCount:    metrics.SlowQueryCount,
		ErrorCount:        metrics.ErrorCount,
		AvgQueryTime:      metrics.AvgQueryTime,
		MaxQueryTime:      metrics.MaxQueryTime,
		DatabaseSize:      metrics.DatabaseSize,
		TableSizes:        metrics.TableSizes,
		IndexSizes:        metrics.IndexSizes,
		LockCount:         metrics.LockCount,
		DeadlockCount:     metrics.DeadlockCount,
		CacheHitRatio:     metrics.CacheHitRatio,
		Uptime:            metrics.Uptime,
		LastBackup:        metrics.LastBackup,
		BackupSize:        metrics.BackupSize,
		Metadata:          metrics.Metadata,
	}
	if err := udm.monitoringAdapter.RecordDatabaseMetrics(ctx, monitoringMetrics); err != nil {
		log.Printf("Warning: failed to record database metrics in unified system: %v", err)
	}

	return metrics, nil
}

// GetMetrics returns collected metrics from the unified monitoring system
func (udm *UnifiedDatabaseMonitor) GetMetrics(ctx context.Context, limit int) ([]*monitoring.UnifiedMetric, error) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Last 24 hours

	return udm.monitoringAdapter.GetPerformanceMetrics(ctx, "database", "database_monitor", startTime, endTime, limit)
}

// GetLatestMetrics returns the most recent metrics from the unified monitoring system
func (udm *UnifiedDatabaseMonitor) GetLatestMetrics(ctx context.Context) (*monitoring.UnifiedMetric, error) {
	metrics, err := udm.GetMetrics(ctx, 1)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metrics found")
	}

	return metrics[0], nil
}

// GetMetricsSummary returns a summary of metrics from the unified monitoring system
func (udm *UnifiedDatabaseMonitor) GetMetricsSummary(ctx context.Context) (map[string]interface{}, error) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Last 24 hours

	return udm.monitoringAdapter.GetDatabaseMetricsSummary(ctx, startTime, endTime)
}

// RecordQuery records a query execution for monitoring
func (udm *UnifiedDatabaseMonitor) RecordQuery(ctx context.Context, query string, duration time.Duration, err error) {
	udm.mu.Lock()
	defer udm.mu.Unlock()

	// Create a simple hash of the query for grouping
	queryHash := fmt.Sprintf("%d", len(query)) // Simplified for demo

	stats, exists := udm.queryStats[queryHash]
	if !exists {
		stats = &QueryStats{
			MinTime: duration,
		}
		udm.queryStats[queryHash] = stats
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

	// Record query performance metric in unified system
	// Use LegacyPerformanceMetric which has Tags and Metadata fields
	queryMetric := &monitoring.LegacyPerformanceMetric{
		Name:      "query_execution_time",
		Value:     float64(duration.Milliseconds()),
		Unit:      "ms",
		Timestamp: time.Now(),
		Tags: map[string]interface{}{
			"query_hash": queryHash,
			"error":      err != nil,
		},
		Metadata: map[string]interface{}{
			"query_length": len(query),
			"error_count":  stats.ErrorCount,
		},
	}

	if recordErr := udm.monitoringAdapter.RecordPerformanceMetric(ctx, "database", "query_monitor", queryMetric); recordErr != nil {
		log.Printf("Warning: failed to record query metric in unified system: %v", recordErr)
	}
}

// GetQueryStats returns query statistics
func (udm *UnifiedDatabaseMonitor) GetQueryStats() map[string]*QueryStats {
	udm.mu.RLock()
	defer udm.mu.RUnlock()

	result := make(map[string]*QueryStats)
	for k, v := range udm.queryStats {
		result[k] = v
	}
	return result
}

// GetActiveAlerts returns active database alerts from the unified monitoring system
func (udm *UnifiedDatabaseMonitor) GetActiveAlerts(ctx context.Context) ([]*monitoring.UnifiedAlert, error) {
	return udm.monitoringAdapter.GetActiveAlerts(ctx, "database")
}

// CreateAlert creates a new alert in the unified monitoring system
func (udm *UnifiedDatabaseMonitor) CreateAlert(ctx context.Context, alertName, description string, severity monitoring.AlertSeverity, condition map[string]interface{}) error {
	return udm.monitoringAdapter.RecordAlert(ctx, "database", "database_monitor", alertName, description, severity, condition)
}

// UpdateAlertStatus updates the status of an alert in the unified monitoring system
func (udm *UnifiedDatabaseMonitor) UpdateAlertStatus(ctx context.Context, alertID string, status monitoring.AlertStatus, userID *string) error {
	// Convert string IDs to UUIDs
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		return fmt.Errorf("invalid alert ID: %w", err)
	}

	var userUUID *uuid.UUID
	if userID != nil {
		parsedUserID, err := uuid.Parse(*userID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %w", err)
		}
		userUUID = &parsedUserID
	}

	return udm.monitoringAdapter.UpdateAlertStatus(ctx, alertUUID, status, userUUID)
}

// Helper methods for database information gathering

func (udm *UnifiedDatabaseMonitor) getDatabaseSize(ctx context.Context) (int64, error) {
	query := `
		SELECT pg_database_size(current_database())
	`
	var size int64
	err := udm.db.QueryRowContext(ctx, query).Scan(&size)
	return size, err
}

func (udm *UnifiedDatabaseMonitor) getTableSizes(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_total_relation_size(schemaname||'.'||tablename) as size
		FROM pg_tables 
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY size DESC
		LIMIT 20
	`

	rows, err := udm.db.QueryContext(ctx, query)
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
		sizes[schema+"."+table] = size
	}

	return sizes, nil
}

func (udm *UnifiedDatabaseMonitor) getIndexSizes(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT 
			schemaname,
			indexname,
			pg_relation_size(schemaname||'.'||indexname) as size
		FROM pg_indexes 
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY size DESC
		LIMIT 20
	`

	rows, err := udm.db.QueryContext(ctx, query)
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
		sizes[schema+"."+index] = size
	}

	return sizes, nil
}

func (udm *UnifiedDatabaseMonitor) getCacheHitRatio(ctx context.Context) (float64, error) {
	query := `
		SELECT 
			round(
				(sum(blks_hit) * 100.0 / (sum(blks_hit) + sum(blks_read)))::numeric, 2
			) as cache_hit_ratio
		FROM pg_stat_database 
		WHERE datname = current_database()
	`

	var ratio float64
	err := udm.db.QueryRowContext(ctx, query).Scan(&ratio)
	return ratio, err
}

func (udm *UnifiedDatabaseMonitor) getLockCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM pg_locks 
		WHERE NOT granted
	`

	var count int
	err := udm.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (udm *UnifiedDatabaseMonitor) getUptime(ctx context.Context) (time.Duration, error) {
	query := `
		SELECT 
			EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))::int as uptime_seconds
	`

	var uptimeSeconds int
	err := udm.db.QueryRowContext(ctx, query).Scan(&uptimeSeconds)
	if err != nil {
		return 0, err
	}

	return time.Duration(uptimeSeconds) * time.Second, nil
}

func (udm *UnifiedDatabaseMonitor) calculateQueryStats(metrics *DatabaseMetrics) {
	udm.mu.RLock()
	defer udm.mu.RUnlock()

	var totalQueries, totalErrors, slowQueries int64
	var totalTime time.Duration

	for _, stats := range udm.queryStats {
		totalQueries += stats.Count
		totalErrors += stats.ErrorCount
		totalTime += stats.TotalTime

		// Use a default slow query threshold if not configured
		slowQueryThreshold := 1 * time.Second // Default threshold
		if udm.config != nil {
			// Check if config has SlowQueryThreshold field (may not exist in all configs)
			// For now, use default
		}
		if stats.AvgTime > slowQueryThreshold {
			slowQueries += stats.Count
		}
	}

	metrics.QueryCount = totalQueries
	metrics.ErrorCount = totalErrors
	metrics.SlowQueryCount = slowQueries

	if totalQueries > 0 {
		metrics.AvgQueryTime = float64(totalTime.Milliseconds()) / float64(totalQueries)
	}

	// Find max query time
	for _, stats := range udm.queryStats {
		if float64(stats.MaxTime.Milliseconds()) > metrics.MaxQueryTime {
			metrics.MaxQueryTime = float64(stats.MaxTime.Milliseconds())
		}
	}
}
