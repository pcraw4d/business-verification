package performance

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DBOptimizer provides database performance optimization
type DBOptimizer struct {
	logger   *zap.Logger
	db       *sql.DB
	profiler *Profiler
	config   *DBConfig
	stats    *DBStats
	mu       sync.RWMutex
}

// DBConfig contains database optimization configuration
type DBConfig struct {
	MaxOpenConns        int           `json:"max_open_conns"`
	MaxIdleConns        int           `json:"max_idle_conns"`
	ConnMaxLifetime     time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime     time.Duration `json:"conn_max_idle_time"`
	QueryTimeout        time.Duration `json:"query_timeout"`
	SlowQueryThreshold  time.Duration `json:"slow_query_threshold"`
	EnableQueryLogging  bool          `json:"enable_query_logging"`
	EnableSlowQueryLog  bool          `json:"enable_slow_query_log"`
	EnableConnectionLog bool          `json:"enable_connection_log"`
}

// DBStats contains database performance statistics
type DBStats struct {
	OpenConnections   int           `json:"open_connections"`
	InUseConnections  int           `json:"in_use_connections"`
	IdleConnections   int           `json:"idle_connections"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	TotalQueries      int64         `json:"total_queries"`
	SlowQueries       int64         `json:"slow_queries"`
	AverageQueryTime  time.Duration `json:"average_query_time"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// QueryStats tracks individual query performance
type QueryStats struct {
	Query        string        `json:"query"`
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	MinTime      time.Duration `json:"min_time"`
	MaxTime      time.Duration `json:"max_time"`
	LastTime     time.Duration `json:"last_time"`
	LastExecuted time.Time     `json:"last_executed"`
}

// NewDBOptimizer creates a new database optimizer
func NewDBOptimizer(logger *zap.Logger, db *sql.DB, profiler *Profiler, config *DBConfig) *DBOptimizer {
	optimizer := &DBOptimizer{
		logger:   logger,
		db:       db,
		profiler: profiler,
		config:   config,
		stats:    &DBStats{},
	}

	// Apply configuration
	optimizer.applyConfig()

	// Start monitoring
	go optimizer.monitor()

	return optimizer
}

// applyConfig applies database configuration
func (dbo *DBOptimizer) applyConfig() {
	dbo.db.SetMaxOpenConns(dbo.config.MaxOpenConns)
	dbo.db.SetMaxIdleConns(dbo.config.MaxIdleConns)
	dbo.db.SetConnMaxLifetime(dbo.config.ConnMaxLifetime)
	dbo.db.SetConnMaxIdleTime(dbo.config.ConnMaxIdleTime)

	dbo.logger.Info("Database configuration applied",
		zap.Int("max_open_conns", dbo.config.MaxOpenConns),
		zap.Int("max_idle_conns", dbo.config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", dbo.config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", dbo.config.ConnMaxIdleTime),
		zap.Duration("query_timeout", dbo.config.QueryTimeout))
}

// ExecuteQuery executes a query with performance monitoring
func (dbo *DBOptimizer) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()

	// Add timeout to context
	if dbo.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dbo.config.QueryTimeout)
		defer cancel()
	}

	// Execute query
	rows, err := dbo.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics
	dbo.updateQueryStats(query, duration, err)

	// Log slow queries
	if dbo.config.EnableSlowQueryLog && duration > dbo.config.SlowQueryThreshold {
		dbo.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Duration("threshold", dbo.config.SlowQueryThreshold),
			zap.Error(err))
	}

	// Log all queries if enabled
	if dbo.config.EnableQueryLogging {
		dbo.logger.Debug("Query executed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err))
	}

	return rows, err
}

// ExecuteQueryRow executes a single row query with performance monitoring
func (dbo *DBOptimizer) ExecuteQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()

	// Add timeout to context
	if dbo.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dbo.config.QueryTimeout)
		defer cancel()
	}

	// Execute query
	row := dbo.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics
	dbo.updateQueryStats(query, duration, nil)

	// Log slow queries
	if dbo.config.EnableSlowQueryLog && duration > dbo.config.SlowQueryThreshold {
		dbo.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Duration("threshold", dbo.config.SlowQueryThreshold))
	}

	// Log all queries if enabled
	if dbo.config.EnableQueryLogging {
		dbo.logger.Debug("Query executed",
			zap.String("query", query),
			zap.Duration("duration", duration))
	}

	return row
}

// ExecuteExec executes an INSERT/UPDATE/DELETE query with performance monitoring
func (dbo *DBOptimizer) ExecuteExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()

	// Add timeout to context
	if dbo.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, dbo.config.QueryTimeout)
		defer cancel()
	}

	// Execute query
	result, err := dbo.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Update statistics
	dbo.updateQueryStats(query, duration, err)

	// Log slow queries
	if dbo.config.EnableSlowQueryLog && duration > dbo.config.SlowQueryThreshold {
		dbo.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Duration("threshold", dbo.config.SlowQueryThreshold),
			zap.Error(err))
	}

	// Log all queries if enabled
	if dbo.config.EnableQueryLogging {
		dbo.logger.Debug("Query executed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err))
	}

	return result, err
}

// updateQueryStats updates query performance statistics
func (dbo *DBOptimizer) updateQueryStats(query string, duration time.Duration, err error) {
	dbo.mu.Lock()
	defer dbo.mu.Unlock()

	dbo.stats.TotalQueries++
	if duration > dbo.config.SlowQueryThreshold {
		dbo.stats.SlowQueries++
	}

	// Update average query time
	if dbo.stats.TotalQueries == 1 {
		dbo.stats.AverageQueryTime = duration
	} else {
		// Calculate running average
		totalTime := dbo.stats.AverageQueryTime * time.Duration(dbo.stats.TotalQueries-1)
		dbo.stats.AverageQueryTime = (totalTime + duration) / time.Duration(dbo.stats.TotalQueries)
	}

	dbo.stats.LastUpdated = time.Now()

	// Record in profiler
	if dbo.profiler != nil {
		dbo.profiler.RecordMetric("db_query", duration)
	}
}

// GetStats returns current database statistics
func (dbo *DBOptimizer) GetStats() *DBStats {
	dbo.mu.RLock()
	defer dbo.mu.RUnlock()

	// Get current connection stats
	stats := dbo.db.Stats()

	// Create a copy of our stats
	result := &DBStats{
		OpenConnections:   stats.OpenConnections,
		InUseConnections:  stats.InUse,
		IdleConnections:   stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		TotalQueries:      dbo.stats.TotalQueries,
		SlowQueries:       dbo.stats.SlowQueries,
		AverageQueryTime:  dbo.stats.AverageQueryTime,
		LastUpdated:       dbo.stats.LastUpdated,
	}

	return result
}

// monitor monitors database performance
func (dbo *DBOptimizer) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := dbo.GetStats()

		// Log connection pool status
		if dbo.config.EnableConnectionLog {
			dbo.logger.Info("Database connection pool status",
				zap.Int("open_connections", stats.OpenConnections),
				zap.Int("in_use_connections", stats.InUseConnections),
				zap.Int("idle_connections", stats.IdleConnections),
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("wait_duration", stats.WaitDuration))
		}

		// Check for connection pool issues
		if stats.WaitCount > 0 {
			dbo.logger.Warn("Database connection pool under pressure",
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("wait_duration", stats.WaitDuration),
				zap.Int("open_connections", stats.OpenConnections))
		}

		// Check for slow queries
		if stats.SlowQueries > 0 {
			slowQueryRatio := float64(stats.SlowQueries) / float64(stats.TotalQueries)
			if slowQueryRatio > 0.1 { // More than 10% slow queries
				dbo.logger.Warn("High percentage of slow queries",
					zap.Int64("slow_queries", stats.SlowQueries),
					zap.Int64("total_queries", stats.TotalQueries),
					zap.Float64("slow_query_ratio", slowQueryRatio))
			}
		}
	}
}

// OptimizeQuery optimizes a query for better performance
func (dbo *DBOptimizer) OptimizeQuery(query string) string {
	// Basic query optimization
	optimized := query

	// Remove unnecessary whitespace
	optimized = fmt.Sprintf("%s", optimized)

	// Add query hints if needed
	// This is a placeholder for more sophisticated optimization

	return optimized
}

// GetSlowQueries returns information about slow queries
func (dbo *DBOptimizer) GetSlowQueries() map[string]*QueryStats {
	// This would typically query the database for slow query logs
	// For now, return empty map
	return make(map[string]*QueryStats)
}

// AnalyzeQueryPlan analyzes the execution plan for a query
func (dbo *DBOptimizer) AnalyzeQueryPlan(ctx context.Context, query string) (string, error) {
	// This would typically use EXPLAIN or similar database-specific commands
	// For now, return a placeholder
	return "Query plan analysis not implemented", nil
}

// GetIndexRecommendations returns index recommendations for better performance
func (dbo *DBOptimizer) GetIndexRecommendations() []string {
	// This would typically analyze query patterns and suggest indexes
	// For now, return some common recommendations
	return []string{
		"Consider adding indexes on frequently queried columns",
		"Review foreign key constraints for performance impact",
		"Consider partitioning large tables",
		"Optimize query patterns to reduce full table scans",
	}
}

// ResetStats resets database statistics
func (dbo *DBOptimizer) ResetStats() {
	dbo.mu.Lock()
	defer dbo.mu.Unlock()

	dbo.stats = &DBStats{
		LastUpdated: time.Now(),
	}

	dbo.logger.Info("Database statistics reset")
}

// SetConfig updates database configuration
func (dbo *DBOptimizer) SetConfig(config *DBConfig) {
	dbo.mu.Lock()
	defer dbo.mu.Unlock()

	dbo.config = config
	dbo.applyConfig()

	dbo.logger.Info("Database configuration updated")
}

// GetPerformanceReport generates a database performance report
func (dbo *DBOptimizer) GetPerformanceReport() string {
	stats := dbo.GetStats()

	report := fmt.Sprintf("=== DATABASE PERFORMANCE REPORT ===\n")
	report += fmt.Sprintf("Generated: %s\n", stats.LastUpdated.Format(time.RFC3339))
	report += fmt.Sprintf("Open Connections: %d\n", stats.OpenConnections)
	report += fmt.Sprintf("In Use Connections: %d\n", stats.InUseConnections)
	report += fmt.Sprintf("Idle Connections: %d\n", stats.IdleConnections)
	report += fmt.Sprintf("Wait Count: %d\n", stats.WaitCount)
	report += fmt.Sprintf("Wait Duration: %v\n", stats.WaitDuration)
	report += fmt.Sprintf("Total Queries: %d\n", stats.TotalQueries)
	report += fmt.Sprintf("Slow Queries: %d\n", stats.SlowQueries)
	report += fmt.Sprintf("Average Query Time: %v\n", stats.AverageQueryTime)

	if stats.TotalQueries > 0 {
		slowQueryRatio := float64(stats.SlowQueries) / float64(stats.TotalQueries) * 100
		report += fmt.Sprintf("Slow Query Ratio: %.2f%%\n", slowQueryRatio)
	}

	report += fmt.Sprintf("\n=== CONFIGURATION ===\n")
	report += fmt.Sprintf("Max Open Connections: %d\n", dbo.config.MaxOpenConns)
	report += fmt.Sprintf("Max Idle Connections: %d\n", dbo.config.MaxIdleConns)
	report += fmt.Sprintf("Connection Max Lifetime: %v\n", dbo.config.ConnMaxLifetime)
	report += fmt.Sprintf("Connection Max Idle Time: %v\n", dbo.config.ConnMaxIdleTime)
	report += fmt.Sprintf("Query Timeout: %v\n", dbo.config.QueryTimeout)
	report += fmt.Sprintf("Slow Query Threshold: %v\n", dbo.config.SlowQueryThreshold)

	return report
}

// DefaultDBConfig returns a default database configuration
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		MaxOpenConns:        25,
		MaxIdleConns:        5,
		ConnMaxLifetime:     5 * time.Minute,
		ConnMaxIdleTime:     1 * time.Minute,
		QueryTimeout:        30 * time.Second,
		SlowQueryThreshold:  1 * time.Second,
		EnableQueryLogging:  false,
		EnableSlowQueryLog:  true,
		EnableConnectionLog: true,
	}
}
