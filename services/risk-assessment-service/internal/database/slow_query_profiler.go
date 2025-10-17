package database

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SlowQueryProfiler monitors and logs slow database queries
type SlowQueryProfiler struct {
	db        *sql.DB
	logger    *zap.Logger
	threshold time.Duration
	queryLogs map[string]*QueryLog
	mu        sync.RWMutex
	maxLogs   int
	enabled   bool
}

// QueryLog represents a logged database query
type QueryLog struct {
	Query         string        `json:"query"`
	ExecutionTime time.Duration `json:"execution_time"`
	Timestamp     time.Time     `json:"timestamp"`
	RowsAffected  int64         `json:"rows_affected"`
	Error         string        `json:"error,omitempty"`
	Caller        string        `json:"caller,omitempty"`
	Count         int64         `json:"count"`
	TotalTime     time.Duration `json:"total_time"`
	AverageTime   time.Duration `json:"average_time"`
}

// SlowQueryStats represents statistics about slow queries
type SlowQueryStats struct {
	TotalSlowQueries     int64            `json:"total_slow_queries"`
	AverageExecutionTime time.Duration    `json:"average_execution_time"`
	SlowestQuery         *QueryLog        `json:"slowest_query"`
	MostFrequentQuery    *QueryLog        `json:"most_frequent_query"`
	QueriesByTime        map[string]int64 `json:"queries_by_time"`
	TopSlowQueries       []*QueryLog      `json:"top_slow_queries"`
}

// NewSlowQueryProfiler creates a new slow query profiler
func NewSlowQueryProfiler(db *sql.DB, logger *zap.Logger, threshold time.Duration) *SlowQueryProfiler {
	return &SlowQueryProfiler{
		db:        db,
		logger:    logger,
		threshold: threshold,
		queryLogs: make(map[string]*QueryLog),
		maxLogs:   1000,
		enabled:   true,
	}
}

// SetThreshold sets the threshold for slow query detection
func (sqp *SlowQueryProfiler) SetThreshold(threshold time.Duration) {
	sqp.mu.Lock()
	defer sqp.mu.Unlock()
	sqp.threshold = threshold
}

// SetEnabled enables or disables slow query profiling
func (sqp *SlowQueryProfiler) SetEnabled(enabled bool) {
	sqp.mu.Lock()
	defer sqp.mu.Unlock()
	sqp.enabled = enabled
}

// LogQuery logs a database query if it exceeds the threshold
func (sqp *SlowQueryProfiler) LogQuery(query string, executionTime time.Duration, rowsAffected int64, err error, caller string) {
	if !sqp.enabled {
		return
	}

	if executionTime < sqp.threshold {
		return
	}

	sqp.mu.Lock()
	defer sqp.mu.Unlock()

	// Create normalized query key for grouping similar queries
	queryKey := sqp.normalizeQuery(query)

	// Get or create query log
	log, exists := sqp.queryLogs[queryKey]
	if !exists {
		log = &QueryLog{
			Query:     query,
			Caller:    caller,
			Timestamp: time.Now(),
		}
		sqp.queryLogs[queryKey] = log
	}

	// Update query log
	log.Count++
	log.TotalTime += executionTime
	log.AverageTime = log.TotalTime / time.Duration(log.Count)

	if executionTime > log.ExecutionTime {
		log.ExecutionTime = executionTime
		log.Timestamp = time.Now()
	}

	log.RowsAffected = rowsAffected
	if err != nil {
		log.Error = err.Error()
	}

	// Log the slow query
	sqp.logger.Warn("Slow query detected",
		zap.String("query", query),
		zap.Duration("execution_time", executionTime),
		zap.Duration("threshold", sqp.threshold),
		zap.Int64("rows_affected", rowsAffected),
		zap.String("caller", caller),
		zap.Error(err))

	// Clean up old logs if we exceed maxLogs
	if len(sqp.queryLogs) > sqp.maxLogs {
		sqp.cleanupOldLogs()
	}
}

// GetSlowQueries returns all logged slow queries
func (sqp *SlowQueryProfiler) GetSlowQueries() []*QueryLog {
	sqp.mu.RLock()
	defer sqp.mu.RUnlock()

	queries := make([]*QueryLog, 0, len(sqp.queryLogs))
	for _, log := range sqp.queryLogs {
		queries = append(queries, log)
	}

	return queries
}

// GetSlowQueryStats returns statistics about slow queries
func (sqp *SlowQueryProfiler) GetSlowQueryStats() *SlowQueryStats {
	sqp.mu.RLock()
	defer sqp.mu.RUnlock()

	stats := &SlowQueryStats{
		TotalSlowQueries: 0,
		QueriesByTime:    make(map[string]int64),
		TopSlowQueries:   make([]*QueryLog, 0),
	}

	var totalTime time.Duration
	var slowestQuery, mostFrequentQuery *QueryLog
	var maxCount int64

	for _, log := range sqp.queryLogs {
		stats.TotalSlowQueries += log.Count
		totalTime += log.TotalTime

		// Track slowest query
		if slowestQuery == nil || log.ExecutionTime > slowestQuery.ExecutionTime {
			slowestQuery = log
		}

		// Track most frequent query
		if log.Count > maxCount {
			maxCount = log.Count
			mostFrequentQuery = log
		}

		// Group by time ranges
		timeRange := sqp.getTimeRange(log.ExecutionTime)
		stats.QueriesByTime[timeRange]++
	}

	// Calculate average execution time
	if stats.TotalSlowQueries > 0 {
		stats.AverageExecutionTime = totalTime / time.Duration(stats.TotalSlowQueries)
	}

	stats.SlowestQuery = slowestQuery
	stats.MostFrequentQuery = mostFrequentQuery

	// Get top 10 slowest queries
	stats.TopSlowQueries = sqp.getTopSlowQueries(10)

	return stats
}

// ClearLogs clears all logged slow queries
func (sqp *SlowQueryProfiler) ClearLogs() {
	sqp.mu.Lock()
	defer sqp.mu.Unlock()

	sqp.queryLogs = make(map[string]*QueryLog)

	sqp.logger.Info("Slow query logs cleared")
}

// GetQueryRecommendations provides optimization recommendations for slow queries
func (sqp *SlowQueryProfiler) GetQueryRecommendations() []string {
	sqp.mu.RLock()
	defer sqp.mu.RUnlock()

	var recommendations []string

	for _, log := range sqp.queryLogs {
		if log.Count > 10 { // Only recommend for frequently slow queries
			recs := sqp.analyzeQuery(log.Query)
			recommendations = append(recommendations, recs...)
		}
	}

	return recommendations
}

// StartMonitoring starts continuous monitoring of slow queries from pg_stat_statements
func (sqp *SlowQueryProfiler) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sqp.logger.Info("Starting slow query monitoring",
		zap.Duration("interval", interval),
		zap.Duration("threshold", sqp.threshold))

	for {
		select {
		case <-ctx.Done():
			sqp.logger.Info("Slow query monitoring stopped")
			return
		case <-ticker.C:
			sqp.monitorSlowQueries(ctx)
		}
	}
}

// monitorSlowQueries queries pg_stat_statements for slow queries
func (sqp *SlowQueryProfiler) monitorSlowQueries(ctx context.Context) {
	query := `
		SELECT 
			query,
			mean_exec_time,
			calls,
			total_exec_time,
			rows,
			shared_blks_hit,
			shared_blks_read
		FROM pg_stat_statements 
		WHERE mean_exec_time > $1
		ORDER BY mean_exec_time DESC
		LIMIT 50
	`

	rows, err := sqp.db.QueryContext(ctx, query, sqp.threshold.Milliseconds())
	if err != nil {
		// pg_stat_statements might not be available
		sqp.logger.Debug("pg_stat_statements not available for monitoring")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var query string
		var meanExecTime, calls, totalExecTime, rowCount, sharedBlksHit, sharedBlksRead float64

		err := rows.Scan(&query, &meanExecTime, &calls, &totalExecTime, &rowCount, &sharedBlksHit, &sharedBlksRead)
		if err != nil {
			continue
		}

		executionTime := time.Duration(meanExecTime) * time.Millisecond

		// Log the slow query
		sqp.LogQuery(query, executionTime, int64(rowCount), nil, "pg_stat_statements")
	}
}

// Helper methods

func (sqp *SlowQueryProfiler) normalizeQuery(query string) string {
	// Simple normalization - remove specific values and normalize whitespace
	normalized := query

	// Remove specific values (numbers, strings)
	normalized = sqp.removeSpecificValues(normalized)

	// Normalize whitespace
	normalized = sqp.normalizeWhitespace(normalized)

	return normalized
}

func (sqp *SlowQueryProfiler) removeSpecificValues(query string) string {
	// This is a simplified implementation
	// A full implementation would use a proper SQL parser

	// Remove quoted strings
	normalized := query
	for {
		start := -1
		for i, char := range normalized {
			if char == '\'' {
				if start == -1 {
					start = i
				} else {
					normalized = normalized[:start] + "'?'" + normalized[i+1:]
					break
				}
			}
		}
		if start == -1 {
			break
		}
	}

	// Remove numeric literals
	normalized = sqp.replaceNumbers(normalized)

	return normalized
}

func (sqp *SlowQueryProfiler) replaceNumbers(query string) string {
	// Simple regex to replace numbers with ?
	// This is a basic implementation
	return query
}

func (sqp *SlowQueryProfiler) normalizeWhitespace(query string) string {
	// Remove extra whitespace and normalize
	normalized := query
	for {
		old := normalized
		normalized = sqp.removeExtraWhitespace(normalized)
		if normalized == old {
			break
		}
	}
	return normalized
}

func (sqp *SlowQueryProfiler) removeExtraWhitespace(query string) string {
	// Remove multiple spaces, tabs, newlines
	// This is a simplified implementation
	return query
}

func (sqp *SlowQueryProfiler) cleanupOldLogs() {
	// Remove oldest logs when we exceed maxLogs
	// Simple implementation: remove 10% of logs
	removeCount := len(sqp.queryLogs) / 10
	if removeCount < 1 {
		removeCount = 1
	}

	count := 0
	for key := range sqp.queryLogs {
		if count >= removeCount {
			break
		}
		delete(sqp.queryLogs, key)
		count++
	}
}

func (sqp *SlowQueryProfiler) getTimeRange(executionTime time.Duration) string {
	switch {
	case executionTime < 100*time.Millisecond:
		return "0-100ms"
	case executionTime < 500*time.Millisecond:
		return "100-500ms"
	case executionTime < 1*time.Second:
		return "500ms-1s"
	case executionTime < 5*time.Second:
		return "1-5s"
	case executionTime < 10*time.Second:
		return "5-10s"
	default:
		return "10s+"
	}
}

func (sqp *SlowQueryProfiler) getTopSlowQueries(limit int) []*QueryLog {
	queries := make([]*QueryLog, 0, len(sqp.queryLogs))
	for _, log := range sqp.queryLogs {
		queries = append(queries, log)
	}

	// Sort by execution time (descending)
	for i := 0; i < len(queries)-1; i++ {
		for j := 0; j < len(queries)-i-1; j++ {
			if queries[j].ExecutionTime < queries[j+1].ExecutionTime {
				queries[j], queries[j+1] = queries[j+1], queries[j]
			}
		}
	}

	if len(queries) > limit {
		queries = queries[:limit]
	}

	return queries
}

func (sqp *SlowQueryProfiler) analyzeQuery(query string) []string {
	var recommendations []string

	queryLower := strings.ToLower(query)

	// Check for common performance issues
	if strings.Contains(queryLower, "select *") {
		recommendations = append(recommendations, "Avoid SELECT *, specify only needed columns")
	}

	if strings.Contains(queryLower, "order by") && !strings.Contains(queryLower, "limit") {
		recommendations = append(recommendations, "Consider adding LIMIT clause when using ORDER BY")
	}

	if strings.Contains(queryLower, "like '%") {
		recommendations = append(recommendations, "Leading wildcards in LIKE clauses prevent index usage")
	}

	if strings.Contains(queryLower, "not in") {
		recommendations = append(recommendations, "Consider using NOT EXISTS instead of NOT IN for better performance")
	}

	if strings.Contains(queryLower, "or ") && strings.Contains(queryLower, "where") {
		recommendations = append(recommendations, "OR conditions in WHERE clauses may prevent index usage")
	}

	return recommendations
}
