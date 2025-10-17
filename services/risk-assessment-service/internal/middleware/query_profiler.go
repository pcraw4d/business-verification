package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QueryProfilerMiddleware provides database query profiling middleware
type QueryProfilerMiddleware struct {
	profiler *QueryProfiler
	logger   *zap.Logger
}

// QueryProfiler tracks database query execution times
type QueryProfiler struct {
	queries   map[string]*QueryMetrics
	mu        sync.RWMutex
	threshold time.Duration
	logger    *zap.Logger
}

// QueryMetrics represents metrics for a database query
type QueryMetrics struct {
	Query        string        `json:"query"`
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	MinTime      time.Duration `json:"min_time"`
	MaxTime      time.Duration `json:"max_time"`
	LastExecuted time.Time     `json:"last_executed"`
	Caller       string        `json:"caller"`
	ErrorCount   int64         `json:"error_count"`
	RowsAffected int64         `json:"rows_affected"`
}

// NewQueryProfilerMiddleware creates a new query profiler middleware
func NewQueryProfilerMiddleware(profiler *QueryProfiler, logger *zap.Logger) *QueryProfilerMiddleware {
	return &QueryProfilerMiddleware{
		profiler: profiler,
		logger:   logger,
	}
}

// NewQueryProfiler creates a new query profiler
func NewQueryProfiler(threshold time.Duration, logger *zap.Logger) *QueryProfiler {
	return &QueryProfiler{
		queries:   make(map[string]*QueryMetrics),
		threshold: threshold,
		logger:    logger,
	}
}

// Middleware returns the HTTP middleware function
func (qpm *QueryProfilerMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add query profiler to request context
			ctx := context.WithValue(r.Context(), "query_profiler", qpm.profiler)

			// Wrap response writer to track response time
			start := time.Now()
			ww := &queryProfilerResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r.WithContext(ctx))

			// Log request completion with query metrics
			duration := time.Since(start)
			qpm.logRequestMetrics(r, duration, ww.statusCode)
		})
	}
}

// ProfileQuery profiles a database query execution
func (qp *QueryProfiler) ProfileQuery(ctx context.Context, query string, fn func() (sql.Result, error)) (sql.Result, error) {
	start := time.Now()
	caller := qp.getCaller()

	result, err := fn()

	duration := time.Since(start)

	// Record query metrics
	qp.recordQuery(query, duration, err, caller, result)

	return result, err
}

// ProfileQueryRow profiles a database query that returns rows
func (qp *QueryProfiler) ProfileQueryRow(ctx context.Context, query string, fn func() (*sql.Rows, error)) (*sql.Rows, error) {
	start := time.Now()
	caller := qp.getCaller()

	rows, err := fn()

	duration := time.Since(start)

	// Record query metrics
	qp.recordQuery(query, duration, err, caller, nil)

	return rows, err
}

// ProfileQueryRowSingle profiles a single row query
func (qp *QueryProfiler) ProfileQueryRowSingle(ctx context.Context, query string, fn func() *sql.Row) *sql.Row {
	start := time.Now()
	caller := qp.getCaller()

	row := fn()

	duration := time.Since(start)

	// Record query metrics
	qp.recordQuery(query, duration, nil, caller, nil)

	return row
}

// GetQueryMetrics returns metrics for all tracked queries
func (qp *QueryProfiler) GetQueryMetrics() map[string]*QueryMetrics {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	metrics := make(map[string]*QueryMetrics)
	for key, metric := range qp.queries {
		metrics[key] = metric
	}

	return metrics
}

// GetSlowQueries returns queries that exceed the threshold
func (qp *QueryProfiler) GetSlowQueries() []*QueryMetrics {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	var slowQueries []*QueryMetrics
	for _, metric := range qp.queries {
		if metric.AverageTime > qp.threshold {
			slowQueries = append(slowQueries, metric)
		}
	}

	return slowQueries
}

// GetTopQueries returns the top N queries by execution time
func (qp *QueryProfiler) GetTopQueries(limit int) []*QueryMetrics {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	var queries []*QueryMetrics
	for _, metric := range qp.queries {
		queries = append(queries, metric)
	}

	// Sort by total time (descending)
	for i := 0; i < len(queries)-1; i++ {
		for j := 0; j < len(queries)-i-1; j++ {
			if queries[j].TotalTime < queries[j+1].TotalTime {
				queries[j], queries[j+1] = queries[j+1], queries[j]
			}
		}
	}

	if len(queries) > limit {
		queries = queries[:limit]
	}

	return queries
}

// ClearMetrics clears all query metrics
func (qp *QueryProfiler) ClearMetrics() {
	qp.mu.Lock()
	defer qp.mu.Unlock()

	qp.queries = make(map[string]*QueryMetrics)

	qp.logger.Info("Query metrics cleared")
}

// Helper methods

func (qp *QueryProfiler) recordQuery(query string, duration time.Duration, err error, caller string, result sql.Result) {
	qp.mu.Lock()
	defer qp.mu.Unlock()

	// Normalize query for grouping
	normalizedQuery := qp.normalizeQuery(query)

	// Get or create metrics
	metrics, exists := qp.queries[normalizedQuery]
	if !exists {
		metrics = &QueryMetrics{
			Query:        query,
			Caller:       caller,
			MinTime:      duration,
			LastExecuted: time.Now(),
		}
		qp.queries[normalizedQuery] = metrics
	}

	// Update metrics
	metrics.Count++
	metrics.TotalTime += duration
	metrics.AverageTime = metrics.TotalTime / time.Duration(metrics.Count)
	metrics.LastExecuted = time.Now()

	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}

	if err != nil {
		metrics.ErrorCount++
	}

	if result != nil {
		if rowsAffected, err := result.RowsAffected(); err == nil {
			metrics.RowsAffected += rowsAffected
		}
	}

	// Log slow queries
	if duration > qp.threshold {
		qp.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("execution_time", duration),
			zap.Duration("threshold", qp.threshold),
			zap.String("caller", caller),
			zap.Error(err))
	}
}

func (qp *QueryProfiler) normalizeQuery(query string) string {
	// Simple normalization - remove specific values and normalize whitespace
	normalized := strings.TrimSpace(query)

	// Convert to lowercase for grouping
	normalized = strings.ToLower(normalized)

	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Remove specific values (simplified)
	normalized = qp.removeSpecificValues(normalized)

	return normalized
}

func (qp *QueryProfiler) removeSpecificValues(query string) string {
	// Remove quoted strings
	normalized := query
	for {
		start := strings.Index(normalized, "'")
		if start == -1 {
			break
		}
		end := strings.Index(normalized[start+1:], "'")
		if end == -1 {
			break
		}
		normalized = normalized[:start] + "'?'" + normalized[start+end+2:]
	}

	// Remove numeric literals (simplified)
	// This is a basic implementation - a full version would use regex

	return normalized
}

func (qp *QueryProfiler) getCaller() string {
	// Get the caller function name
	pc := make([]uintptr, 1)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return "unknown"
	}

	frame, _ := runtime.CallersFrames(pc).Next()
	return fmt.Sprintf("%s:%d", frame.Function, frame.Line)
}

func (qpm *QueryProfilerMiddleware) logRequestMetrics(r *http.Request, duration time.Duration, statusCode int) {
	// Get query metrics from context
	profiler, ok := r.Context().Value("query_profiler").(*QueryProfiler)
	if !ok {
		return
	}

	// Get slow queries for this request
	slowQueries := profiler.GetSlowQueries()

	if len(slowQueries) > 0 {
		qpm.logger.Info("Request completed with slow queries",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status_code", statusCode),
			zap.Int("slow_queries", len(slowQueries)))
	}
}

// queryProfilerResponseWriter wraps http.ResponseWriter to capture status code
type queryProfilerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *queryProfilerResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Database wrapper functions for easy integration

// WrapDB wraps a database connection with query profiling
func WrapDB(db *sql.DB, profiler *QueryProfiler) *ProfiledDB {
	return &ProfiledDB{
		DB:       db,
		profiler: profiler,
	}
}

// ProfiledDB wraps sql.DB with query profiling
type ProfiledDB struct {
	*sql.DB
	profiler *QueryProfiler
}

// ExecContext executes a query with profiling
func (pdb *ProfiledDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return pdb.profiler.ProfileQuery(ctx, query, func() (sql.Result, error) {
		return pdb.DB.ExecContext(ctx, query, args...)
	})
}

// QueryContext executes a query that returns rows with profiling
func (pdb *ProfiledDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return pdb.profiler.ProfileQueryRow(ctx, query, func() (*sql.Rows, error) {
		return pdb.DB.QueryContext(ctx, query, args...)
	})
}

// QueryRowContext executes a query that returns a single row with profiling
func (pdb *ProfiledDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return pdb.profiler.ProfileQueryRowSingle(ctx, query, func() *sql.Row {
		return pdb.DB.QueryRowContext(ctx, query, args...)
	})
}
