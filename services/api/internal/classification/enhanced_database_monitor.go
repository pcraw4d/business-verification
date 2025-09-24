package classification

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EnhancedDatabaseMonitor provides comprehensive database query performance monitoring with optimization recommendations
type EnhancedDatabaseMonitor struct {
	db                *sql.DB
	logger            *zap.Logger
	config            *EnhancedDatabaseConfig
	queryStats        map[string]*EnhancedQueryStats
	performanceAlerts []*DatabasePerformanceAlert
	optimizationCache map[string]*QueryOptimization
	mu                sync.RWMutex

	// Monitoring control
	stopCh           chan struct{}
	monitoringActive bool
}

// EnhancedDatabaseConfig holds configuration for enhanced database monitoring
type EnhancedDatabaseConfig struct {
	Enabled               bool          `json:"enabled"`
	CollectionInterval    time.Duration `json:"collection_interval"`
	SlowQueryThreshold    time.Duration `json:"slow_query_threshold"`
	MaxQueryStats         int           `json:"max_query_stats"`
	OptimizationCacheSize int           `json:"optimization_cache_size"`
	AlertingEnabled       bool          `json:"alerting_enabled"`
	TrackQueryPlans       bool          `json:"track_query_plans"`
	TrackIndexUsage       bool          `json:"track_index_usage"`
	TrackConnectionPool   bool          `json:"track_connection_pool"`
	TrackLockContention   bool          `json:"track_lock_contention"`
	TrackDeadlocks        bool          `json:"track_deadlocks"`
	TrackCacheHitRatio    bool          `json:"track_cache_hit_ratio"`
	TrackTableSizes       bool          `json:"track_table_sizes"`
	TrackIndexSizes       bool          `json:"track_index_sizes"`
	TrackVacuumStats      bool          `json:"track_vacuum_stats"`
	TrackReplicationLag   bool          `json:"track_replication_lag"`
}

// EnhancedQueryStats represents enhanced query performance statistics
type EnhancedQueryStats struct {
	QueryID              string                 `json:"query_id"`
	QueryText            string                 `json:"query_text"`
	QueryHash            string                 `json:"query_hash"`
	ExecutionCount       int64                  `json:"execution_count"`
	TotalExecutionTime   float64                `json:"total_execution_time_ms"`
	AverageExecutionTime float64                `json:"average_execution_time_ms"`
	MinExecutionTime     float64                `json:"min_execution_time_ms"`
	MaxExecutionTime     float64                `json:"max_execution_time_ms"`
	P50ExecutionTime     float64                `json:"p50_execution_time_ms"`
	P95ExecutionTime     float64                `json:"p95_execution_time_ms"`
	P99ExecutionTime     float64                `json:"p99_execution_time_ms"`
	RowsReturned         int64                  `json:"rows_returned"`
	RowsExamined         int64                  `json:"rows_examined"`
	RowsAffected         int64                  `json:"rows_affected"`
	IndexUsageScore      float64                `json:"index_usage_score"`
	CacheHitRatio        float64                `json:"cache_hit_ratio"`
	PerformanceCategory  string                 `json:"performance_category"`
	OptimizationPriority string                 `json:"optimization_priority"`
	OptimizationScore    float64                `json:"optimization_score"`
	LastExecuted         time.Time              `json:"last_executed"`
	FirstExecuted        time.Time              `json:"first_executed"`
	ErrorCount           int64                  `json:"error_count"`
	TimeoutCount         int64                  `json:"timeout_count"`
	LockWaitTime         float64                `json:"lock_wait_time_ms"`
	BufferReads          int64                  `json:"buffer_reads"`
	BufferHits           int64                  `json:"buffer_hits"`
	TempFilesCreated     int64                  `json:"temp_files_created"`
	TempFileSize         int64                  `json:"temp_file_size_bytes"`
	SortOperations       int64                  `json:"sort_operations"`
	HashOperations       int64                  `json:"hash_operations"`
	JoinOperations       int64                  `json:"join_operations"`
	SubqueryOperations   int64                  `json:"subquery_operations"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// DatabasePerformanceAlert represents a database performance alert
type DatabasePerformanceAlert struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	AlertType       string                 `json:"alert_type"`
	Severity        string                 `json:"severity"`
	QueryID         string                 `json:"query_id,omitempty"`
	QueryText       string                 `json:"query_text,omitempty"`
	Threshold       float64                `json:"threshold"`
	ActualValue     float64                `json:"actual_value"`
	Message         string                 `json:"message"`
	Recommendations []string               `json:"recommendations"`
	Resolved        bool                   `json:"resolved"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// QueryOptimization represents query optimization recommendations
type QueryOptimization struct {
	QueryID                 string                        `json:"query_id"`
	QueryText               string                        `json:"query_text"`
	OptimizationScore       float64                       `json:"optimization_score"`
	Priority                string                        `json:"priority"`
	EstimatedImprovement    float64                       `json:"estimated_improvement_percent"`
	Recommendations         []*OptimizationRecommendation `json:"recommendations"`
	IndexSuggestions        []*IndexSuggestion            `json:"index_suggestions"`
	QueryRewriteSuggestions []*QueryRewriteSuggestion     `json:"query_rewrite_suggestions"`
	LastAnalyzed            time.Time                     `json:"last_analyzed"`
	Metadata                map[string]interface{}        `json:"metadata,omitempty"`
}

// OptimizationRecommendation represents a specific optimization recommendation
type OptimizationRecommendation struct {
	Type                 string  `json:"type"`
	Description          string  `json:"description"`
	Impact               string  `json:"impact"` // "low", "medium", "high", "critical"
	Effort               string  `json:"effort"` // "low", "medium", "high"
	EstimatedImprovement float64 `json:"estimated_improvement_percent"`
	Code                 string  `json:"code,omitempty"`
	Example              string  `json:"example,omitempty"`
}

// IndexSuggestion represents an index optimization suggestion
type IndexSuggestion struct {
	TableName            string   `json:"table_name"`
	ColumnNames          []string `json:"column_names"`
	IndexType            string   `json:"index_type"`
	EstimatedImprovement float64  `json:"estimated_improvement_percent"`
	EstimatedSize        int64    `json:"estimated_size_bytes"`
	CreateStatement      string   `json:"create_statement"`
	DropStatement        string   `json:"drop_statement,omitempty"`
	Reason               string   `json:"reason"`
}

// QueryRewriteSuggestion represents a query rewrite suggestion
type QueryRewriteSuggestion struct {
	OriginalQuery        string  `json:"original_query"`
	OptimizedQuery       string  `json:"optimized_query"`
	Description          string  `json:"description"`
	EstimatedImprovement float64 `json:"estimated_improvement_percent"`
	Reason               string  `json:"reason"`
	Example              string  `json:"example,omitempty"`
}

// DatabaseSystemStats represents overall database system statistics
type DatabaseSystemStats struct {
	Timestamp          time.Time              `json:"timestamp"`
	ConnectionCount    int                    `json:"connection_count"`
	ActiveConnections  int                    `json:"active_connections"`
	IdleConnections    int                    `json:"idle_connections"`
	MaxConnections     int                    `json:"max_connections"`
	DatabaseSize       int64                  `json:"database_size_bytes"`
	CacheHitRatio      float64                `json:"cache_hit_ratio"`
	IndexHitRatio      float64                `json:"index_hit_ratio"`
	LockCount          int                    `json:"lock_count"`
	DeadlockCount      int                    `json:"deadlock_count"`
	LongRunningQueries int                    `json:"long_running_queries"`
	SlowQueries        int                    `json:"slow_queries"`
	BlockedQueries     int                    `json:"blocked_queries"`
	TempFilesCreated   int64                  `json:"temp_files_created"`
	TempFileSize       int64                  `json:"temp_file_size_bytes"`
	VacuumOperations   int64                  `json:"vacuum_operations"`
	AnalyzeOperations  int64                  `json:"analyze_operations"`
	ReplicationLag     time.Duration          `json:"replication_lag"`
	Uptime             time.Duration          `json:"uptime"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NewEnhancedDatabaseMonitor creates a new enhanced database monitor
func NewEnhancedDatabaseMonitor(db *sql.DB, logger *zap.Logger, config *EnhancedDatabaseConfig) *EnhancedDatabaseMonitor {
	if config == nil {
		config = DefaultEnhancedDatabaseConfig()
	}

	monitor := &EnhancedDatabaseMonitor{
		db:                db,
		logger:            logger,
		config:            config,
		queryStats:        make(map[string]*EnhancedQueryStats),
		performanceAlerts: make([]*DatabasePerformanceAlert, 0),
		optimizationCache: make(map[string]*QueryOptimization),
		stopCh:            make(chan struct{}),
		monitoringActive:  false,
	}

	// Start monitoring if enabled
	if config.Enabled {
		monitor.Start()
	}

	return monitor
}

// DefaultEnhancedDatabaseConfig returns default configuration
func DefaultEnhancedDatabaseConfig() *EnhancedDatabaseConfig {
	return &EnhancedDatabaseConfig{
		Enabled:               true,
		CollectionInterval:    30 * time.Second,
		SlowQueryThreshold:    100 * time.Millisecond,
		MaxQueryStats:         1000,
		OptimizationCacheSize: 500,
		AlertingEnabled:       true,
		TrackQueryPlans:       true,
		TrackIndexUsage:       true,
		TrackConnectionPool:   true,
		TrackLockContention:   true,
		TrackDeadlocks:        true,
		TrackCacheHitRatio:    true,
		TrackTableSizes:       true,
		TrackIndexSizes:       true,
		TrackVacuumStats:      true,
		TrackReplicationLag:   false,
	}
}

// Start starts the database monitoring
func (edm *EnhancedDatabaseMonitor) Start() {
	edm.mu.Lock()
	defer edm.mu.Unlock()

	if edm.monitoringActive {
		return
	}

	edm.monitoringActive = true

	// Start background monitoring
	go edm.monitoringLoop()

	edm.logger.Info("Enhanced database monitoring started",
		zap.Duration("collection_interval", edm.config.CollectionInterval),
		zap.Duration("slow_query_threshold", edm.config.SlowQueryThreshold),
		zap.Bool("track_query_plans", edm.config.TrackQueryPlans),
		zap.Bool("track_index_usage", edm.config.TrackIndexUsage))
}

// Stop stops the database monitoring
func (edm *EnhancedDatabaseMonitor) Stop() {
	edm.mu.Lock()
	defer edm.mu.Unlock()

	if !edm.monitoringActive {
		return
	}

	edm.monitoringActive = false
	close(edm.stopCh)

	edm.logger.Info("Enhanced database monitoring stopped")
}

// monitoringLoop runs the main monitoring loop
func (edm *EnhancedDatabaseMonitor) monitoringLoop() {
	ticker := time.NewTicker(edm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			edm.collectSystemStats()
			edm.analyzeQueryPerformance()
			edm.generateOptimizationRecommendations()
		case <-edm.stopCh:
			return
		}
	}
}

// RecordQueryExecution records a query execution for monitoring
func (edm *EnhancedDatabaseMonitor) RecordQueryExecution(
	ctx context.Context,
	queryText string,
	executionTime time.Duration,
	rowsReturned, rowsExamined int64,
	errorOccurred bool,
	errorMessage string,
) {
	if !edm.config.Enabled {
		return
	}

	edm.mu.Lock()
	defer edm.mu.Unlock()

	queryHash := edm.generateQueryHash(queryText)
	queryID := edm.generateQueryID(queryText)

	// Get or create query stats
	stats, exists := edm.queryStats[queryHash]
	if !exists {
		stats = &EnhancedQueryStats{
			QueryID:          queryID,
			QueryText:        queryText,
			QueryHash:        queryHash,
			FirstExecuted:    time.Now(),
			MinExecutionTime: float64(executionTime.Milliseconds()),
			Metadata:         make(map[string]interface{}),
		}
		edm.queryStats[queryHash] = stats
	}

	// Update statistics
	stats.ExecutionCount++
	stats.TotalExecutionTime += float64(executionTime.Milliseconds())
	stats.AverageExecutionTime = stats.TotalExecutionTime / float64(stats.ExecutionCount)
	stats.LastExecuted = time.Now()

	if float64(executionTime.Milliseconds()) > stats.MaxExecutionTime {
		stats.MaxExecutionTime = float64(executionTime.Milliseconds())
	}
	if float64(executionTime.Milliseconds()) < stats.MinExecutionTime {
		stats.MinExecutionTime = float64(executionTime.Milliseconds())
	}

	stats.RowsReturned += rowsReturned
	stats.RowsExamined += rowsExamined

	if errorOccurred {
		stats.ErrorCount++
	}

	// Calculate performance metrics
	edm.calculateQueryPerformanceMetrics(stats)

	// Check for alerts
	if edm.config.AlertingEnabled {
		edm.checkQueryAlerts(stats)
	}

	// Clean up old stats if needed
	if len(edm.queryStats) > edm.config.MaxQueryStats {
		edm.cleanupOldQueryStats()
	}
}

// calculateQueryPerformanceMetrics calculates performance metrics for a query
func (edm *EnhancedDatabaseMonitor) calculateQueryPerformanceMetrics(stats *EnhancedQueryStats) {
	// Calculate performance category
	if stats.AverageExecutionTime < float64(edm.config.SlowQueryThreshold.Milliseconds())/2 {
		stats.PerformanceCategory = "excellent"
	} else if stats.AverageExecutionTime < float64(edm.config.SlowQueryThreshold.Milliseconds()) {
		stats.PerformanceCategory = "good"
	} else if stats.AverageExecutionTime < float64(edm.config.SlowQueryThreshold.Milliseconds())*2 {
		stats.PerformanceCategory = "fair"
	} else if stats.AverageExecutionTime < float64(edm.config.SlowQueryThreshold.Milliseconds())*5 {
		stats.PerformanceCategory = "poor"
	} else {
		stats.PerformanceCategory = "critical"
	}

	// Calculate optimization priority
	if stats.PerformanceCategory == "critical" || stats.ExecutionCount > 1000 {
		stats.OptimizationPriority = "high"
	} else if stats.PerformanceCategory == "poor" || stats.ExecutionCount > 100 {
		stats.OptimizationPriority = "medium"
	} else {
		stats.OptimizationPriority = "low"
	}

	// Calculate optimization score (0-100, higher is better)
	optimizationScore := 100.0

	// Deduct points for slow execution
	if stats.AverageExecutionTime > float64(edm.config.SlowQueryThreshold.Milliseconds()) {
		optimizationScore -= (stats.AverageExecutionTime - float64(edm.config.SlowQueryThreshold.Milliseconds())) / 10
	}

	// Deduct points for high error rate
	if stats.ExecutionCount > 0 {
		errorRate := float64(stats.ErrorCount) / float64(stats.ExecutionCount)
		optimizationScore -= errorRate * 50
	}

	// Deduct points for poor row efficiency
	if stats.RowsExamined > 0 {
		efficiency := float64(stats.RowsReturned) / float64(stats.RowsExamined)
		if efficiency < 0.1 {
			optimizationScore -= 20
		} else if efficiency < 0.5 {
			optimizationScore -= 10
		}
	}

	// Ensure score is between 0 and 100
	if optimizationScore < 0 {
		optimizationScore = 0
	}
	if optimizationScore > 100 {
		optimizationScore = 100
	}

	stats.OptimizationScore = optimizationScore
}

// checkQueryAlerts checks for query performance alerts
func (edm *EnhancedDatabaseMonitor) checkQueryAlerts(stats *EnhancedQueryStats) {
	var alerts []*DatabasePerformanceAlert

	// Check for slow queries
	if stats.AverageExecutionTime > float64(edm.config.SlowQueryThreshold.Milliseconds()) {
		alerts = append(alerts, &DatabasePerformanceAlert{
			ID:          fmt.Sprintf("slow_query_%s_%d", stats.QueryHash, time.Now().Unix()),
			Timestamp:   time.Now(),
			AlertType:   "slow_query",
			Severity:    edm.determineAlertSeverity(stats.AverageExecutionTime, float64(edm.config.SlowQueryThreshold.Milliseconds())),
			QueryID:     stats.QueryID,
			QueryText:   stats.QueryText,
			Threshold:   float64(edm.config.SlowQueryThreshold.Milliseconds()),
			ActualValue: stats.AverageExecutionTime,
			Message:     fmt.Sprintf("Query %s has average execution time of %.2fms, exceeding threshold of %.2fms", stats.QueryID, stats.AverageExecutionTime, float64(edm.config.SlowQueryThreshold.Milliseconds())),
			Recommendations: []string{
				"Consider adding indexes to improve query performance",
				"Review query execution plan for optimization opportunities",
				"Consider query rewriting or restructuring",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"execution_count": stats.ExecutionCount,
				"rows_returned":   stats.RowsReturned,
				"rows_examined":   stats.RowsExamined,
			},
		})
	}

	// Check for high error rate
	if stats.ExecutionCount > 10 && float64(stats.ErrorCount)/float64(stats.ExecutionCount) > 0.1 {
		alerts = append(alerts, &DatabasePerformanceAlert{
			ID:          fmt.Sprintf("high_error_rate_%s_%d", stats.QueryHash, time.Now().Unix()),
			Timestamp:   time.Now(),
			AlertType:   "high_error_rate",
			Severity:    "high",
			QueryID:     stats.QueryID,
			QueryText:   stats.QueryText,
			Threshold:   0.1,
			ActualValue: float64(stats.ErrorCount) / float64(stats.ExecutionCount),
			Message:     fmt.Sprintf("Query %s has high error rate of %.2f%%, exceeding threshold of 10%%", stats.QueryID, float64(stats.ErrorCount)/float64(stats.ExecutionCount)*100),
			Recommendations: []string{
				"Review query logic for potential issues",
				"Check for data type mismatches or constraint violations",
				"Consider adding proper error handling",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"execution_count": stats.ExecutionCount,
				"error_count":     stats.ErrorCount,
			},
		})
	}

	// Check for inefficient queries (high rows examined vs returned ratio)
	if stats.RowsExamined > 0 && float64(stats.RowsReturned)/float64(stats.RowsExamined) < 0.01 {
		alerts = append(alerts, &DatabasePerformanceAlert{
			ID:          fmt.Sprintf("inefficient_query_%s_%d", stats.QueryHash, time.Now().Unix()),
			Timestamp:   time.Now(),
			AlertType:   "inefficient_query",
			Severity:    "medium",
			QueryID:     stats.QueryID,
			QueryText:   stats.QueryText,
			Threshold:   0.01,
			ActualValue: float64(stats.RowsReturned) / float64(stats.RowsExamined),
			Message:     fmt.Sprintf("Query %s is inefficient, examining %d rows but returning only %d", stats.QueryID, stats.RowsExamined, stats.RowsReturned),
			Recommendations: []string{
				"Add appropriate indexes to reduce rows examined",
				"Consider adding WHERE clauses to filter data earlier",
				"Review query structure for optimization opportunities",
			},
			Resolved: false,
			Metadata: map[string]interface{}{
				"rows_examined": stats.RowsExamined,
				"rows_returned": stats.RowsReturned,
			},
		})
	}

	// Add alerts to the list
	for _, alert := range alerts {
		edm.performanceAlerts = append(edm.performanceAlerts, alert)

		// Keep only recent alerts
		if len(edm.performanceAlerts) > 1000 {
			edm.performanceAlerts = edm.performanceAlerts[1:]
		}

		edm.logger.Warn("Database performance alert triggered",
			zap.String("alert_id", alert.ID),
			zap.String("alert_type", alert.AlertType),
			zap.String("severity", alert.Severity),
			zap.String("query_id", alert.QueryID),
			zap.String("message", alert.Message))
	}
}

// determineAlertSeverity determines alert severity based on threshold ratio
func (edm *EnhancedDatabaseMonitor) determineAlertSeverity(actual, threshold float64) string {
	ratio := actual / threshold
	if ratio >= 5.0 {
		return "critical"
	} else if ratio >= 3.0 {
		return "high"
	} else if ratio >= 2.0 {
		return "medium"
	}
	return "low"
}

// generateQueryHash generates a hash for the query
func (edm *EnhancedDatabaseMonitor) generateQueryHash(queryText string) string {
	// Simple hash implementation - in production, use a proper hash function
	return fmt.Sprintf("%d", len(queryText))
}

// generateQueryID generates a unique ID for the query
func (edm *EnhancedDatabaseMonitor) generateQueryID(queryText string) string {
	return fmt.Sprintf("query_%d_%d", len(queryText), time.Now().UnixNano())
}

// cleanupOldQueryStats removes old query statistics
func (edm *EnhancedDatabaseMonitor) cleanupOldQueryStats() {
	// Remove oldest 10% of stats
	removeCount := len(edm.queryStats) / 10
	if removeCount == 0 {
		removeCount = 1
	}

	// Simple cleanup - remove oldest by first executed time
	// In production, you might want a more sophisticated cleanup strategy
	count := 0
	for hash, stats := range edm.queryStats {
		if count >= removeCount {
			break
		}
		if time.Since(stats.FirstExecuted) > 24*time.Hour {
			delete(edm.queryStats, hash)
			count++
		}
	}
}

// collectSystemStats collects overall database system statistics
func (edm *EnhancedDatabaseMonitor) collectSystemStats() {
	// This would collect system-wide database statistics
	// Implementation depends on the specific database system
	edm.logger.Debug("Collecting database system statistics")
}

// analyzeQueryPerformance analyzes query performance patterns
func (edm *EnhancedDatabaseMonitor) analyzeQueryPerformance() {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	// Analyze performance patterns and generate insights
	for _, stats := range edm.queryStats {
		if stats.OptimizationScore < 50 && stats.ExecutionCount > 10 {
			edm.logger.Info("Query performance analysis",
				zap.String("query_id", stats.QueryID),
				zap.Float64("optimization_score", stats.OptimizationScore),
				zap.String("performance_category", stats.PerformanceCategory),
				zap.String("optimization_priority", stats.OptimizationPriority),
				zap.Int64("execution_count", stats.ExecutionCount))
		}
	}
}

// generateOptimizationRecommendations generates optimization recommendations
func (edm *EnhancedDatabaseMonitor) generateOptimizationRecommendations() {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	// Generate optimization recommendations for poorly performing queries
	for hash, stats := range edm.queryStats {
		if stats.OptimizationScore < 70 && stats.ExecutionCount > 5 {
			optimization := edm.createQueryOptimization(stats)
			edm.optimizationCache[hash] = optimization
		}
	}

	// Clean up old optimizations
	if len(edm.optimizationCache) > edm.config.OptimizationCacheSize {
		edm.cleanupOldOptimizations()
	}
}

// createQueryOptimization creates optimization recommendations for a query
func (edm *EnhancedDatabaseMonitor) createQueryOptimization(stats *EnhancedQueryStats) *QueryOptimization {
	optimization := &QueryOptimization{
		QueryID:                 stats.QueryID,
		QueryText:               stats.QueryText,
		OptimizationScore:       stats.OptimizationScore,
		Priority:                stats.OptimizationPriority,
		EstimatedImprovement:    100 - stats.OptimizationScore,
		Recommendations:         make([]*OptimizationRecommendation, 0),
		IndexSuggestions:        make([]*IndexSuggestion, 0),
		QueryRewriteSuggestions: make([]*QueryRewriteSuggestion, 0),
		LastAnalyzed:            time.Now(),
		Metadata:                make(map[string]interface{}),
	}

	// Generate recommendations based on performance issues
	if stats.AverageExecutionTime > float64(edm.config.SlowQueryThreshold.Milliseconds()) {
		optimization.Recommendations = append(optimization.Recommendations, &OptimizationRecommendation{
			Type:                 "index_optimization",
			Description:          "Add appropriate indexes to improve query performance",
			Impact:               "high",
			Effort:               "medium",
			EstimatedImprovement: 50.0,
			Code:                 "-- Add indexes based on WHERE and JOIN conditions",
			Example:              "CREATE INDEX idx_table_column ON table_name (column_name);",
		})
	}

	if stats.RowsExamined > stats.RowsReturned*10 {
		optimization.Recommendations = append(optimization.Recommendations, &OptimizationRecommendation{
			Type:                 "query_restructuring",
			Description:          "Restructure query to reduce rows examined",
			Impact:               "high",
			Effort:               "high",
			EstimatedImprovement: 40.0,
			Code:                 "-- Add WHERE clauses to filter data earlier",
			Example:              "SELECT * FROM table WHERE condition1 AND condition2;",
		})
	}

	if stats.ErrorCount > 0 {
		optimization.Recommendations = append(optimization.Recommendations, &OptimizationRecommendation{
			Type:                 "error_handling",
			Description:          "Improve error handling and query logic",
			Impact:               "medium",
			Effort:               "medium",
			EstimatedImprovement: 20.0,
			Code:                 "-- Add proper error handling and validation",
			Example:              "BEGIN ... EXCEPTION WHEN ... END;",
		})
	}

	return optimization
}

// cleanupOldOptimizations removes old optimization recommendations
func (edm *EnhancedDatabaseMonitor) cleanupOldOptimizations() {
	// Remove oldest 20% of optimizations
	removeCount := len(edm.optimizationCache) / 5
	if removeCount == 0 {
		removeCount = 1
	}

	count := 0
	for hash, optimization := range edm.optimizationCache {
		if count >= removeCount {
			break
		}
		if time.Since(optimization.LastAnalyzed) > 7*24*time.Hour {
			delete(edm.optimizationCache, hash)
			count++
		}
	}
}

// GetQueryStats returns query performance statistics
func (edm *EnhancedDatabaseMonitor) GetQueryStats(limit int) map[string]*EnhancedQueryStats {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	result := make(map[string]*EnhancedQueryStats)
	count := 0

	for hash, stats := range edm.queryStats {
		if limit > 0 && count >= limit {
			break
		}
		result[hash] = stats
		count++
	}

	return result
}

// GetPerformanceAlerts returns database performance alerts
func (edm *EnhancedDatabaseMonitor) GetPerformanceAlerts(resolved bool, limit int) []*DatabasePerformanceAlert {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	var alerts []*DatabasePerformanceAlert
	count := 0

	for _, alert := range edm.performanceAlerts {
		if alert.Resolved == resolved {
			if limit > 0 && count >= limit {
				break
			}
			alerts = append(alerts, alert)
			count++
		}
	}

	return alerts
}

// GetOptimizationRecommendations returns query optimization recommendations
func (edm *EnhancedDatabaseMonitor) GetOptimizationRecommendations(priority string, limit int) map[string]*QueryOptimization {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	result := make(map[string]*QueryOptimization)
	count := 0

	for hash, optimization := range edm.optimizationCache {
		if priority == "" || optimization.Priority == priority {
			if limit > 0 && count >= limit {
				break
			}
			result[hash] = optimization
			count++
		}
	}

	return result
}

// GetDatabaseSummary returns a summary of database performance
func (edm *EnhancedDatabaseMonitor) GetDatabaseSummary() map[string]interface{} {
	edm.mu.RLock()
	defer edm.mu.RUnlock()

	// Calculate summary statistics
	totalQueries := 0
	slowQueries := 0
	highErrorQueries := 0
	avgOptimizationScore := 0.0

	for _, stats := range edm.queryStats {
		totalQueries++
		if stats.AverageExecutionTime > float64(edm.config.SlowQueryThreshold.Milliseconds()) {
			slowQueries++
		}
		if stats.ExecutionCount > 0 && float64(stats.ErrorCount)/float64(stats.ExecutionCount) > 0.1 {
			highErrorQueries++
		}
		avgOptimizationScore += stats.OptimizationScore
	}

	if totalQueries > 0 {
		avgOptimizationScore /= float64(totalQueries)
	}

	activeAlerts := 0
	for _, alert := range edm.performanceAlerts {
		if !alert.Resolved {
			activeAlerts++
		}
	}

	summary := map[string]interface{}{
		"timestamp": time.Now(),
		"queries": map[string]interface{}{
			"total_queries":          totalQueries,
			"slow_queries":           slowQueries,
			"high_error_queries":     highErrorQueries,
			"avg_optimization_score": avgOptimizationScore,
		},
		"alerts": map[string]interface{}{
			"total_alerts":    len(edm.performanceAlerts),
			"active_alerts":   activeAlerts,
			"resolved_alerts": len(edm.performanceAlerts) - activeAlerts,
		},
		"optimizations": map[string]interface{}{
			"total_optimizations": len(edm.optimizationCache),
			"high_priority":       edm.countOptimizationsByPriority("high"),
			"medium_priority":     edm.countOptimizationsByPriority("medium"),
			"low_priority":        edm.countOptimizationsByPriority("low"),
		},
		"configuration": map[string]interface{}{
			"slow_query_threshold_ms": edm.config.SlowQueryThreshold.Milliseconds(),
			"collection_interval_sec": edm.config.CollectionInterval.Seconds(),
			"max_query_stats":         edm.config.MaxQueryStats,
			"alerting_enabled":        edm.config.AlertingEnabled,
		},
	}

	return summary
}

// countOptimizationsByPriority counts optimizations by priority
func (edm *EnhancedDatabaseMonitor) countOptimizationsByPriority(priority string) int {
	count := 0
	for _, optimization := range edm.optimizationCache {
		if optimization.Priority == priority {
			count++
		}
	}
	return count
}
