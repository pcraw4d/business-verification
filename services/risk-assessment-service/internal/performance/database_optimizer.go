package performance

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/pool"
)

// DatabaseOptimizer provides database-specific optimizations
type DatabaseOptimizer struct {
	pool   *pool.ConnectionPool
	logger *zap.Logger

	// Optimization settings
	config *DatabaseOptimizationConfig
	stats  *DatabaseStats
}

// DatabaseOptimizationConfig represents database optimization configuration
type DatabaseOptimizationConfig struct {
	// Connection pool settings
	MaxConnections     int           `json:"max_connections"`
	MinConnections     int           `json:"min_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`
	MaxLifetime        time.Duration `json:"max_lifetime"`

	// Query optimization
	QueryTimeout       time.Duration `json:"query_timeout"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
	EnableQueryCache   bool          `json:"enable_query_cache"`
	QueryCacheSize     int           `json:"query_cache_size"`

	// Index optimization
	EnableIndexOptimization bool          `json:"enable_index_optimization"`
	IndexAnalysisInterval   time.Duration `json:"index_analysis_interval"`

	// Statistics
	EnableStatsCollection   bool          `json:"enable_stats_collection"`
	StatsCollectionInterval time.Duration `json:"stats_collection_interval"`
}

// DatabaseStats represents database performance statistics
type DatabaseStats struct {
	// Connection stats
	ActiveConnections int           `json:"active_connections"`
	IdleConnections   int           `json:"idle_connections"`
	TotalConnections  int           `json:"total_connections"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`

	// Query stats
	TotalQueries     int64         `json:"total_queries"`
	SlowQueries      int64         `json:"slow_queries"`
	AverageQueryTime time.Duration `json:"average_query_time"`
	MaxQueryTime     time.Duration `json:"max_query_time"`

	// Cache stats
	QueryCacheHits    int64   `json:"query_cache_hits"`
	QueryCacheMisses  int64   `json:"query_cache_misses"`
	QueryCacheHitRate float64 `json:"query_cache_hit_rate"`

	// Index stats
	IndexUsage     map[string]int64 `json:"index_usage"`
	UnusedIndexes  []string         `json:"unused_indexes"`
	MissingIndexes []string         `json:"missing_indexes"`

	// Performance indicators
	IsOptimized       bool      `json:"is_optimized"`
	OptimizationScore float64   `json:"optimization_score"`
	LastOptimized     time.Time `json:"last_optimized"`
	LastUpdated       time.Time `json:"last_updated"`
}

// NewDatabaseOptimizer creates a new database optimizer
func NewDatabaseOptimizer(pool *pool.ConnectionPool, logger *zap.Logger) *DatabaseOptimizer {
	config := &DatabaseOptimizationConfig{
		MaxConnections:     100,
		MinConnections:     10,
		MaxIdleConnections: 20,
		ConnectionTimeout:  30 * time.Second,
		IdleTimeout:        5 * time.Minute,
		MaxLifetime:        1 * time.Hour,

		QueryTimeout:       10 * time.Second,
		SlowQueryThreshold: 1 * time.Second,
		EnableQueryCache:   true,
		QueryCacheSize:     1000,

		EnableIndexOptimization: true,
		IndexAnalysisInterval:   24 * time.Hour,

		EnableStatsCollection:   true,
		StatsCollectionInterval: 5 * time.Minute,
	}

	return &DatabaseOptimizer{
		pool:   pool,
		logger: logger,
		config: config,
		stats:  &DatabaseStats{},
	}
}

// Optimize performs database optimization
func (do *DatabaseOptimizer) Optimize(ctx context.Context) error {
	do.logger.Info("Starting database optimization")

	// Update statistics
	if err := do.updateStats(ctx); err != nil {
		do.logger.Error("Failed to update database stats", zap.Error(err))
	}

	// Optimize connection pool
	if err := do.optimizeConnectionPool(ctx); err != nil {
		do.logger.Error("Failed to optimize connection pool", zap.Error(err))
	}

	// Optimize indexes
	if do.config.EnableIndexOptimization {
		if err := do.optimizeIndexes(ctx); err != nil {
			do.logger.Error("Failed to optimize indexes", zap.Error(err))
		}
	}

	// Analyze slow queries
	if err := do.analyzeSlowQueries(ctx); err != nil {
		do.logger.Error("Failed to analyze slow queries", zap.Error(err))
	}

	// Update optimization score
	do.calculateOptimizationScore()

	do.stats.LastOptimized = time.Now()
	do.logger.Info("Database optimization completed",
		zap.Float64("optimization_score", do.stats.OptimizationScore),
		zap.Bool("is_optimized", do.stats.IsOptimized))

	return nil
}

// GetStats returns current database statistics
func (do *DatabaseOptimizer) GetStats() *DatabaseStats {
	return do.stats
}

// updateStats updates database statistics
func (do *DatabaseOptimizer) updateStats(ctx context.Context) error {
	// Get connection pool metrics
	poolMetrics := do.pool.GetMetrics()
	do.stats.ActiveConnections = poolMetrics.ActiveConnections
	do.stats.IdleConnections = poolMetrics.IdleConnections
	do.stats.TotalConnections = poolMetrics.TotalConnections
	do.stats.WaitCount = poolMetrics.WaitCount
	do.stats.WaitDuration = poolMetrics.WaitDuration

	// Get query statistics
	if err := do.getQueryStats(ctx); err != nil {
		return fmt.Errorf("failed to get query stats: %w", err)
	}

	// Get index statistics
	if do.config.EnableIndexOptimization {
		if err := do.getIndexStats(ctx); err != nil {
			return fmt.Errorf("failed to get index stats: %w", err)
		}
	}

	do.stats.LastUpdated = time.Now()
	return nil
}

// getQueryStats retrieves query performance statistics
func (do *DatabaseOptimizer) getQueryStats(ctx context.Context) error {
	// Query to get query statistics from pg_stat_statements
	query := `
		SELECT 
			COUNT(*) as total_queries,
			COUNT(*) FILTER (WHERE mean_time > $1) as slow_queries,
			AVG(mean_time) as avg_query_time,
			MAX(mean_time) as max_query_time
		FROM pg_stat_statements
		WHERE calls > 0
	`

	row := do.pool.QueryRow(ctx, query, do.config.SlowQueryThreshold.Milliseconds())

	var totalQueries, slowQueries int64
	var avgQueryTime, maxQueryTime float64

	err := row.Scan(&totalQueries, &slowQueries, &avgQueryTime, &maxQueryTime)
	if err != nil {
		// If pg_stat_statements is not available, use default values
		do.logger.Warn("pg_stat_statements not available, using default query stats")
		do.stats.TotalQueries = 0
		do.stats.SlowQueries = 0
		do.stats.AverageQueryTime = 0
		do.stats.MaxQueryTime = 0
		return nil
	}

	do.stats.TotalQueries = totalQueries
	do.stats.SlowQueries = slowQueries
	do.stats.AverageQueryTime = time.Duration(avgQueryTime) * time.Millisecond
	do.stats.MaxQueryTime = time.Duration(maxQueryTime) * time.Millisecond

	return nil
}

// getIndexStats retrieves index usage statistics
func (do *DatabaseOptimizer) getIndexStats(ctx context.Context) error {
	// Query to get index usage statistics
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
		ORDER BY idx_scan DESC
	`

	rows, err := do.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query index stats: %w", err)
	}
	defer rows.Close()

	do.stats.IndexUsage = make(map[string]int64)
	do.stats.UnusedIndexes = []string{}

	for rows.Next() {
		var schemaName, tableName, indexName string
		var idxScan, idxTupRead, idxTupFetch int64

		if err := rows.Scan(&schemaName, &tableName, &indexName, &idxScan, &idxTupRead, &idxTupFetch); err != nil {
			continue
		}

		fullIndexName := fmt.Sprintf("%s.%s.%s", schemaName, tableName, indexName)
		do.stats.IndexUsage[fullIndexName] = idxScan

		// Identify unused indexes (no scans in the last period)
		if idxScan == 0 {
			do.stats.UnusedIndexes = append(do.stats.UnusedIndexes, fullIndexName)
		}
	}

	return nil
}

// optimizeConnectionPool optimizes the database connection pool
func (do *DatabaseOptimizer) optimizeConnectionPool(ctx context.Context) error {
	do.logger.Info("Optimizing database connection pool")

	// Check if we need to adjust connection pool settings
	currentMetrics := do.pool.GetMetrics()

	// If we have too many active connections, we might need to increase the pool
	if currentMetrics.ActiveConnections > int(float64(do.config.MaxConnections)*0.8) {
		do.logger.Warn("High connection usage detected",
			zap.Int("active_connections", currentMetrics.ActiveConnections),
			zap.Int("max_connections", do.config.MaxConnections))
	}

	// If we have too many idle connections, we might need to reduce the pool
	if currentMetrics.IdleConnections > int(float64(do.config.MaxIdleConnections)*0.8) {
		do.logger.Warn("High idle connection usage detected",
			zap.Int("idle_connections", currentMetrics.IdleConnections),
			zap.Int("max_idle_connections", do.config.MaxIdleConnections))
	}

	// Check wait duration
	if currentMetrics.WaitDuration > 1*time.Second {
		do.logger.Warn("High connection wait time detected",
			zap.Duration("wait_duration", currentMetrics.WaitDuration))
	}

	return nil
}

// optimizeIndexes optimizes database indexes
func (do *DatabaseOptimizer) optimizeIndexes(ctx context.Context) error {
	do.logger.Info("Optimizing database indexes")

	// Analyze tables to update statistics
	analyzeQuery := "ANALYZE"
	if _, err := do.pool.Exec(ctx, analyzeQuery); err != nil {
		do.logger.Warn("Failed to analyze tables", zap.Error(err))
	}

	// Check for missing indexes on frequently queried columns
	if err := do.checkMissingIndexes(ctx); err != nil {
		do.logger.Warn("Failed to check for missing indexes", zap.Error(err))
	}

	// Suggest index optimizations
	if err := do.suggestIndexOptimizations(ctx); err != nil {
		do.logger.Warn("Failed to suggest index optimizations", zap.Error(err))
	}

	return nil
}

// checkMissingIndexes checks for missing indexes
func (do *DatabaseOptimizer) checkMissingIndexes(ctx context.Context) error {
	// Query to find tables that might benefit from indexes
	query := `
		SELECT 
			schemaname,
			tablename,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch
		FROM pg_stat_user_tables
		WHERE schemaname = 'public'
		AND seq_scan > idx_scan * 2
		AND seq_tup_read > 1000
		ORDER BY seq_tup_read DESC
	`

	rows, err := do.pool.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	do.stats.MissingIndexes = []string{}

	for rows.Next() {
		var schemaName, tableName string
		var seqScan, seqTupRead, idxScan, idxTupFetch int64

		if err := rows.Scan(&schemaName, &tableName, &seqScan, &seqTupRead, &idxScan, &idxTupFetch); err != nil {
			continue
		}

		tableName = fmt.Sprintf("%s.%s", schemaName, tableName)
		do.stats.MissingIndexes = append(do.stats.MissingIndexes, tableName)

		do.logger.Info("Table might benefit from additional indexes",
			zap.String("table", tableName),
			zap.Int64("seq_scan", seqScan),
			zap.Int64("seq_tup_read", seqTupRead),
			zap.Int64("idx_scan", idxScan))
	}

	return nil
}

// suggestIndexOptimizations suggests index optimizations
func (do *DatabaseOptimizer) suggestIndexOptimizations(ctx context.Context) error {
	// This would typically involve more sophisticated analysis
	// For now, we'll log suggestions based on the statistics we've collected

	if len(do.stats.UnusedIndexes) > 0 {
		do.logger.Info("Unused indexes detected",
			zap.Strings("unused_indexes", do.stats.UnusedIndexes))
	}

	if len(do.stats.MissingIndexes) > 0 {
		do.logger.Info("Tables that might benefit from indexes",
			zap.Strings("missing_indexes", do.stats.MissingIndexes))
	}

	return nil
}

// analyzeSlowQueries analyzes slow queries
func (do *DatabaseOptimizer) analyzeSlowQueries(ctx context.Context) error {
	do.logger.Info("Analyzing slow queries")

	// Query to get slow queries from pg_stat_statements
	query := `
		SELECT 
			query,
			calls,
			total_time,
			mean_time,
			stddev_time,
			rows
		FROM pg_stat_statements
		WHERE mean_time > $1
		ORDER BY mean_time DESC
		LIMIT 10
	`

	rows, err := do.pool.Query(ctx, query, do.config.SlowQueryThreshold.Milliseconds())
	if err != nil {
		// If pg_stat_statements is not available, skip analysis
		do.logger.Warn("pg_stat_statements not available, skipping slow query analysis")
		return nil
	}
	defer rows.Close()

	slowQueries := []string{}

	for rows.Next() {
		var queryText string
		var calls int64
		var totalTime, meanTime, stddevTime float64
		var rowCount int64

		if err := rows.Scan(&queryText, &calls, &totalTime, &meanTime, &stddevTime, &rowCount); err != nil {
			continue
		}

		// Truncate long queries for logging
		if len(queryText) > 200 {
			queryText = queryText[:200] + "..."
		}

		slowQueries = append(slowQueries, queryText)

		do.logger.Warn("Slow query detected",
			zap.String("query", queryText),
			zap.Int64("calls", calls),
			zap.Float64("mean_time_ms", meanTime),
			zap.Float64("total_time_ms", totalTime))
	}

	if len(slowQueries) > 0 {
		do.logger.Info("Slow query analysis completed",
			zap.Int("slow_queries_found", len(slowQueries)))
	}

	return nil
}

// calculateOptimizationScore calculates the database optimization score
func (do *DatabaseOptimizer) calculateOptimizationScore() {
	score := 100.0

	// Connection pool score (30% weight)
	if do.stats.ActiveConnections > int(float64(do.config.MaxConnections)*0.8) {
		score -= 15.0
	}
	if do.stats.WaitDuration > 1*time.Second {
		score -= 15.0
	}

	// Query performance score (40% weight)
	if do.stats.SlowQueries > 10 {
		score -= 20.0
	}
	if do.stats.AverageQueryTime > 100*time.Millisecond {
		score -= 20.0
	}

	// Index optimization score (30% weight)
	if len(do.stats.UnusedIndexes) > 5 {
		score -= 15.0
	}
	if len(do.stats.MissingIndexes) > 3 {
		score -= 15.0
	}

	do.stats.OptimizationScore = score
	do.stats.IsOptimized = score >= 80.0
}

// GetOptimizationRecommendations returns optimization recommendations
func (do *DatabaseOptimizer) GetOptimizationRecommendations() []string {
	recommendations := []string{}

	// Connection pool recommendations
	if do.stats.ActiveConnections > int(float64(do.config.MaxConnections)*0.8) {
		recommendations = append(recommendations, "Consider increasing max_connections")
	}

	if do.stats.WaitDuration > 1*time.Second {
		recommendations = append(recommendations, "High connection wait time - check connection pool settings")
	}

	// Query performance recommendations
	if do.stats.SlowQueries > 10 {
		recommendations = append(recommendations, "Multiple slow queries detected - consider query optimization")
	}

	if do.stats.AverageQueryTime > 100*time.Millisecond {
		recommendations = append(recommendations, "Average query time is high - consider indexing")
	}

	// Index recommendations
	if len(do.stats.UnusedIndexes) > 5 {
		recommendations = append(recommendations, "Multiple unused indexes detected - consider dropping them")
	}

	if len(do.stats.MissingIndexes) > 3 {
		recommendations = append(recommendations, "Tables might benefit from additional indexes")
	}

	return recommendations
}
