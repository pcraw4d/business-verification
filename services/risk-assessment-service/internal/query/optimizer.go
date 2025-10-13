package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// QueryOptimizer provides query optimization and caching
type QueryOptimizer struct {
	db     *sql.DB
	cache  QueryCache
	logger *zap.Logger
}

// QueryCache interface for caching query results
type QueryCache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// QueryMetrics represents query performance metrics
type QueryMetrics struct {
	QueryCount      int64         `json:"query_count"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	AverageLatency  time.Duration `json:"average_latency"`
	SlowQueries     int64         `json:"slow_queries"`
	ErrorCount      int64         `json:"error_count"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// PreparedStatement represents a prepared statement with caching
type PreparedStatement struct {
	stmt   *sql.Stmt
	query  string
	cache  QueryCache
	logger *zap.Logger
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB, cache QueryCache, logger *zap.Logger) *QueryOptimizer {
	return &QueryOptimizer{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// Prepare creates a prepared statement with caching
func (qo *QueryOptimizer) Prepare(ctx context.Context, query string) (*PreparedStatement, error) {
	// Normalize query for caching
	normalizedQuery := qo.normalizeQuery(query)
	
	stmt, err := qo.db.PrepareContext(ctx, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return &PreparedStatement{
		stmt:   stmt,
		query:  normalizedQuery,
		cache:  qo.cache,
		logger: qo.logger,
	}, nil
}

// ExecuteWithCache executes a query with result caching
func (qo *QueryOptimizer) ExecuteWithCache(ctx context.Context, query string, args []interface{}, dest interface{}, ttl time.Duration) error {
	// Generate cache key
	cacheKey := qo.generateCacheKey(query, args)
	
	// Try to get from cache first
	if qo.cache != nil {
		if err := qo.cache.Get(ctx, cacheKey, dest); err == nil {
			return nil
		}
	}

	// Execute query
	start := time.Now()
	rows, err := qo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Process results
	if err := qo.scanRows(rows, dest); err != nil {
		return fmt.Errorf("failed to scan rows: %w", err)
	}

	// Cache result if cache is available
	if qo.cache != nil && ttl > 0 {
		if err := qo.cache.Set(ctx, cacheKey, dest, ttl); err != nil {
			qo.logger.Warn("Failed to cache query result",
				zap.String("query", query),
				zap.Error(err))
		}
	}

	latency := time.Since(start)
	if latency > 1*time.Second {
		qo.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("latency", latency))
	}

	return nil
}

// BatchExecute executes multiple queries in a batch
func (qo *QueryOptimizer) BatchExecute(ctx context.Context, queries []string, args [][]interface{}) error {
	if len(queries) != len(args) {
		return fmt.Errorf("queries and args length mismatch")
	}

	tx, err := qo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, query := range queries {
		if _, err := tx.ExecContext(ctx, query, args[i]...); err != nil {
			return fmt.Errorf("batch query %d failed: %w", i, err)
		}
	}

	return tx.Commit()
}

// ExecuteWithRetry executes a query with retry logic
func (qo *QueryOptimizer) ExecuteWithRetry(ctx context.Context, query string, args []interface{}, maxRetries int) (*sql.Rows, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		rows, err := qo.db.QueryContext(ctx, query, args...)
		if err == nil {
			return rows, nil
		}

		lastErr = err
		
		// Check if error is retryable
		if !qo.isRetryableError(err) {
			return nil, err
		}

		if attempt < maxRetries {
			// Exponential backoff
			delay := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return nil, fmt.Errorf("query failed after %d retries: %w", maxRetries, lastErr)
}

// AnalyzeQuery analyzes query performance
func (qo *QueryOptimizer) AnalyzeQuery(ctx context.Context, query string) (*QueryAnalysis, error) {
	// Add EXPLAIN ANALYZE to the query
	explainQuery := fmt.Sprintf("EXPLAIN ANALYZE %s", query)
	
	rows, err := qo.db.QueryContext(ctx, explainQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}
	defer rows.Close()

	var analysis QueryAnalysis
	var plan strings.Builder

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, fmt.Errorf("failed to scan analysis result: %w", err)
		}
		plan.WriteString(line)
		plan.WriteString("\n")
	}

	analysis.Query = query
	analysis.ExecutionPlan = plan.String()
	analysis.AnalyzedAt = time.Now()

	return &analysis, nil
}

// GetIndexSuggestions suggests indexes for a query
func (qo *QueryOptimizer) GetIndexSuggestions(ctx context.Context, query string) ([]IndexSuggestion, error) {
	// This is a simplified implementation
	// In a real system, you would use pg_stat_statements or similar tools
	
	suggestions := []IndexSuggestion{}
	
	// Analyze WHERE clauses for potential indexes
	whereClauses := qo.extractWhereClauses(query)
	for _, clause := range whereClauses {
		if qo.shouldIndexColumn(clause) {
			suggestions = append(suggestions, IndexSuggestion{
				Table:  clause.Table,
				Column: clause.Column,
				Type:   "btree",
				Reason: "Frequently used in WHERE clause",
			})
		}
	}

	return suggestions, nil
}

// normalizeQuery normalizes a query for consistent caching
func (qo *QueryOptimizer) normalizeQuery(query string) string {
	// Remove extra whitespace
	query = strings.TrimSpace(query)
	
	// Convert to lowercase for consistency
	query = strings.ToLower(query)
	
	// Remove comments
	lines := strings.Split(query, "\n")
	var cleanLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

// generateCacheKey generates a cache key for a query and its arguments
func (qo *QueryOptimizer) generateCacheKey(query string, args []interface{}) string {
	key := fmt.Sprintf("query:%s", qo.normalizeQuery(query))
	
	if len(args) > 0 {
		key += fmt.Sprintf(":args:%v", args)
	}
	
	return key
}

// scanRows scans rows into the destination
func (qo *QueryOptimizer) scanRows(rows *sql.Rows, dest interface{}) error {
	// This is a simplified implementation
	// In a real system, you would use reflection or code generation
	
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// For now, we'll just iterate through rows
	// In a real implementation, you would use reflection to populate dest
	for rows.Next() {
		// Scan logic would go here
		_ = columns
	}

	return rows.Err()
}

// isRetryableError checks if an error is retryable
func (qo *QueryOptimizer) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryableErrors := []string{
		"connection reset by peer",
		"broken pipe",
		"connection refused",
		"timeout",
		"temporary failure",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// extractWhereClauses extracts WHERE clauses from a query
func (qo *QueryOptimizer) extractWhereClauses(query string) []WhereClause {
	// This is a simplified implementation
	// In a real system, you would use a SQL parser
	
	clauses := []WhereClause{}
	
	// Look for WHERE keyword
	whereIndex := strings.Index(strings.ToLower(query), "where")
	if whereIndex == -1 {
		return clauses
	}

	// Extract the WHERE clause
	whereClause := query[whereIndex+5:]
	
	// Simple parsing - look for column references
	// This is very basic and would need proper SQL parsing in production
	words := strings.Fields(whereClause)
	for i, word := range words {
		if strings.Contains(word, ".") {
			parts := strings.Split(word, ".")
			if len(parts) == 2 {
				clauses = append(clauses, WhereClause{
					Table:  parts[0],
					Column: parts[1],
				})
			}
		}
		_ = i // Avoid unused variable warning
	}

	return clauses
}

// shouldIndexColumn determines if a column should be indexed
func (qo *QueryOptimizer) shouldIndexColumn(clause WhereClause) bool {
	// Simple heuristic - in production, you would use query statistics
	return true
}

// QueryAnalysis represents the analysis of a query
type QueryAnalysis struct {
	Query         string    `json:"query"`
	ExecutionPlan string    `json:"execution_plan"`
	AnalyzedAt    time.Time `json:"analyzed_at"`
}

// IndexSuggestion represents a suggested database index
type IndexSuggestion struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// WhereClause represents a WHERE clause condition
type WhereClause struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

// Execute executes a prepared statement
func (ps *PreparedStatement) Execute(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return ps.stmt.QueryContext(ctx, args...)
}

// ExecuteRow executes a prepared statement and returns a single row
func (ps *PreparedStatement) ExecuteRow(ctx context.Context, args ...interface{}) *sql.Row {
	return ps.stmt.QueryRowContext(ctx, args...)
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return ps.stmt.Close()
}
