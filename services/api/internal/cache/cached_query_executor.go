// Package cache provides cached query execution for the KYB Platform
// This module integrates the query cache manager with existing database operations
// to provide transparent caching for frequently executed queries.

package cache

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// CachedQueryExecutor provides cached execution of database queries
type CachedQueryExecutor struct {
	cacheManager *QueryCacheManager
	db           *sql.DB
	config       *ExecutorConfig
}

// ExecutorConfig defines configuration for the cached query executor
type ExecutorConfig struct {
	EnableCaching      bool          `json:"enable_caching"`
	DefaultCacheTTL    time.Duration `json:"default_cache_ttl"`
	CacheOnError       bool          `json:"cache_on_error"`
	MaxCacheSize       int           `json:"max_cache_size"`
	EnableQueryLogging bool          `json:"enable_query_logging"`
}

// QueryResult represents a cached query result
type QueryResult struct {
	Rows      []map[string]interface{} `json:"rows"`
	RowCount  int                      `json:"row_count"`
	Timestamp time.Time                `json:"timestamp"`
	QueryHash string                   `json:"query_hash"`
}

// CachedQuery represents a query that can be cached
type CachedQuery struct {
	Query    string        `json:"query"`
	Args     []interface{} `json:"args"`
	CacheKey *CacheKey     `json:"cache_key"`
	TTL      time.Duration `json:"ttl"`
	UseCache bool          `json:"use_cache"`
}

// NewCachedQueryExecutor creates a new cached query executor
func NewCachedQueryExecutor(cacheManager *QueryCacheManager, db *sql.DB, config *ExecutorConfig) *CachedQueryExecutor {
	if config == nil {
		config = getDefaultExecutorConfig()
	}

	return &CachedQueryExecutor{
		cacheManager: cacheManager,
		db:           db,
		config:       config,
	}
}

// ExecuteQuery executes a query with caching support
func (cqe *CachedQueryExecutor) ExecuteQuery(ctx context.Context, query *CachedQuery) (*QueryResult, error) {
	if !cqe.config.EnableCaching || !query.UseCache {
		return cqe.executeQueryDirectly(ctx, query)
	}

	// Try to get from cache first
	if cachedResult, found, err := cqe.cacheManager.Get(ctx, query.CacheKey); err == nil && found {
		if result, ok := cachedResult.(*QueryResult); ok {
			if cqe.config.EnableQueryLogging {
				log.Printf("Cache HIT for query: %s", query.Query)
			}
			return result, nil
		}
	}

	// Execute query directly
	result, err := cqe.executeQueryDirectly(ctx, query)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := cqe.cacheManager.Set(ctx, query.CacheKey, result); err != nil {
		log.Printf("Warning: Failed to cache query result: %v", err)
	}

	if cqe.config.EnableQueryLogging {
		log.Printf("Cache MISS for query: %s", query.Query)
	}

	return result, nil
}

// ExecuteQueryWithCache executes a query with automatic cache key generation
func (cqe *CachedQueryExecutor) ExecuteQueryWithCache(ctx context.Context, query string, args []interface{}, queryType string, ttl time.Duration) (*QueryResult, error) {
	// Generate cache key
	cacheKey := cqe.generateCacheKey(query, args, queryType, ttl)

	cachedQuery := &CachedQuery{
		Query:    query,
		Args:     args,
		CacheKey: cacheKey,
		TTL:      ttl,
		UseCache: true,
	}

	return cqe.ExecuteQuery(ctx, cachedQuery)
}

// ExecuteQueryDirectly executes a query without caching
func (cqe *CachedQueryExecutor) ExecuteQueryDirectly(ctx context.Context, query string, args []interface{}) (*QueryResult, error) {
	cachedQuery := &CachedQuery{
		Query:    query,
		Args:     args,
		UseCache: false,
	}

	return cqe.executeQueryDirectly(ctx, cachedQuery)
}

// executeQueryDirectly executes a query directly against the database
func (cqe *CachedQueryExecutor) executeQueryDirectly(ctx context.Context, query *CachedQuery) (*QueryResult, error) {
	start := time.Now()

	rows, err := cqe.db.QueryContext(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan rows
	var resultRows []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for the row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		resultRows = append(resultRows, row)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	executionTime := time.Since(start)

	if cqe.config.EnableQueryLogging {
		log.Printf("Query executed in %v: %s", executionTime, query.Query)
	}

	return &QueryResult{
		Rows:      resultRows,
		RowCount:  len(resultRows),
		Timestamp: time.Now(),
		QueryHash: cqe.generateQueryHash(query.Query, query.Args),
	}, nil
}

// generateCacheKey generates a cache key for a query
func (cqe *CachedQueryExecutor) generateCacheKey(query string, args []interface{}, queryType string, ttl time.Duration) *CacheKey {
	// Create parameters map
	params := make(map[string]interface{})
	params["query"] = query
	params["args"] = args

	return &CacheKey{
		QueryType: queryType,
		Params:    params,
		Version:   "v1",
		TTL:       ttl,
	}
}

// generateQueryHash generates a hash for a query
func (cqe *CachedQueryExecutor) generateQueryHash(query string, args []interface{}) string {
	data := fmt.Sprintf("%s:%v", query, args)
	return fmt.Sprintf("%x", data)
}

// InvalidateCacheByQueryType invalidates cache entries for a specific query type
func (cqe *CachedQueryExecutor) InvalidateCacheByQueryType(ctx context.Context, queryType string) error {
	pattern := fmt.Sprintf("kyb:cache:%s:*", queryType)
	return cqe.cacheManager.InvalidateByPattern(ctx, pattern)
}

// InvalidateCacheByBusinessID invalidates cache entries for a specific business
func (cqe *CachedQueryExecutor) InvalidateCacheByBusinessID(ctx context.Context, businessID string) error {
	patterns := []string{
		fmt.Sprintf("kyb:cache:classification:*business_id*%s*", businessID),
		fmt.Sprintf("kyb:cache:risk_assessment:*business_id*%s*", businessID),
		fmt.Sprintf("kyb:cache:business_data:*business_id*%s*", businessID),
	}

	for _, pattern := range patterns {
		if err := cqe.cacheManager.InvalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Warning: Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}

	return nil
}

// InvalidateCacheByUserID invalidates cache entries for a specific user
func (cqe *CachedQueryExecutor) InvalidateCacheByUserID(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("kyb:cache:user_data:*user_id*%s*", userID),
		fmt.Sprintf("kyb:cache:classification:*user_id*%s*", userID),
	}

	for _, pattern := range patterns {
		if err := cqe.cacheManager.InvalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Warning: Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}

	return nil
}

// GetCacheMetrics returns cache performance metrics
func (cqe *CachedQueryExecutor) GetCacheMetrics() *CacheMetrics {
	return cqe.cacheManager.GetMetrics()
}

// Predefined query builders for common operations

// NewClassificationQuery creates a cached query for business classification
func NewClassificationQuery(businessID, websiteURL string) *CachedQuery {
	query := `
		SELECT 
			bc.id,
			bc.business_id,
			bc.primary_industry,
			bc.secondary_industries,
			bc.confidence_score,
			bc.classification_metadata,
			bc.created_at,
			bc.updated_at
		FROM business_classifications bc
		WHERE bc.business_id = $1
		ORDER BY bc.created_at DESC
		LIMIT 1
	`

	cacheKey := NewClassificationCacheKey(businessID, websiteURL).Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{businessID},
		CacheKey: cacheKey,
		TTL:      30 * time.Minute,
		UseCache: true,
	}
}

// NewRiskAssessmentQuery creates a cached query for risk assessment
func NewRiskAssessmentQuery(businessID string, riskLevels []string) *CachedQuery {
	query := `
		SELECT 
			ra.id,
			ra.business_id,
			ra.risk_score,
			ra.risk_level,
			ra.detected_keywords,
			ra.assessment_method,
			ra.assessment_date
		FROM business_risk_assessments ra
		WHERE ra.business_id = $1
			AND ra.risk_level = ANY($2)
		ORDER BY ra.assessment_date DESC
	`

	cacheKey := NewRiskAssessmentCacheKey(businessID, fmt.Sprintf("%v", riskLevels)).Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{businessID, riskLevels},
		CacheKey: cacheKey,
		TTL:      1 * time.Hour,
		UseCache: true,
	}
}

// NewUserDataQuery creates a cached query for user data
func NewUserDataQuery(userID string) *CachedQuery {
	query := `
		SELECT 
			u.id,
			u.email,
			u.name,
			u.role,
			u.is_active,
			u.created_at,
			u.updated_at,
			u.metadata
		FROM users u
		WHERE u.id = $1
	`

	cacheKey := NewUserDataCacheKey(userID).Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{userID},
		CacheKey: cacheKey,
		TTL:      2 * time.Hour,
		UseCache: true,
	}
}

// NewBusinessDataQuery creates a cached query for business data
func NewBusinessDataQuery(businessID string) *CachedQuery {
	query := `
		SELECT 
			b.id,
			b.user_id,
			b.name,
			b.website_url,
			b.industry,
			b.created_at,
			b.updated_at,
			b.metadata
		FROM businesses b
		WHERE b.id = $1
	`

	cacheKey := NewBusinessDataCacheKey(businessID).Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{businessID},
		CacheKey: cacheKey,
		TTL:      1 * time.Hour,
		UseCache: true,
	}
}

// NewTimeBasedClassificationQuery creates a cached query for time-based classifications
func NewTimeBasedClassificationQuery(startTime, endTime time.Time, limit int) *CachedQuery {
	query := `
		SELECT 
			c.id,
			c.business_name,
			c.actual_classification,
			c.confidence_score,
			c.classification_method,
			c.created_at
		FROM classifications c
		WHERE c.created_at BETWEEN $1 AND $2
		ORDER BY c.created_at DESC
		LIMIT $3
	`

	cacheKey := NewCacheKeyBuilder("time_based_classification").
		AddParam("start_time", startTime).
		AddParam("end_time", endTime).
		AddParam("limit", limit).
		SetTTL(15 * time.Minute).
		Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{startTime, endTime, limit},
		CacheKey: cacheKey,
		TTL:      15 * time.Minute,
		UseCache: true,
	}
}

// NewIndustryBasedClassificationQuery creates a cached query for industry-based classifications
func NewIndustryBasedClassificationQuery(startTime, endTime time.Time, industry string, limit int) *CachedQuery {
	query := `
		SELECT 
			c.id,
			c.business_name,
			c.actual_classification,
			c.confidence_score,
			c.classification_method,
			c.created_at
		FROM classifications c
		WHERE c.created_at BETWEEN $1 AND $2
			AND ($3 IS NULL OR c.actual_classification = $3)
		ORDER BY c.actual_classification, c.created_at DESC
		LIMIT $4
	`

	cacheKey := NewCacheKeyBuilder("industry_based_classification").
		AddParam("start_time", startTime).
		AddParam("end_time", endTime).
		AddParam("industry", industry).
		AddParam("limit", limit).
		SetTTL(15 * time.Minute).
		Build()

	return &CachedQuery{
		Query:    query,
		Args:     []interface{}{startTime, endTime, industry, limit},
		CacheKey: cacheKey,
		TTL:      15 * time.Minute,
		UseCache: true,
	}
}

// getDefaultExecutorConfig returns default executor configuration
func getDefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		EnableCaching:      true,
		DefaultCacheTTL:    15 * time.Minute,
		CacheOnError:       false,
		MaxCacheSize:       10000,
		EnableQueryLogging: true,
	}
}

// CacheIntegration provides integration with existing database operations
type CacheIntegration struct {
	executor *CachedQueryExecutor
}

// NewCacheIntegration creates a new cache integration
func NewCacheIntegration(executor *CachedQueryExecutor) *CacheIntegration {
	return &CacheIntegration{
		executor: executor,
	}
}

// ExecuteWithCache executes a query with caching using the integration layer
func (ci *CacheIntegration) ExecuteWithCache(ctx context.Context, query string, args []interface{}, queryType string, ttl time.Duration) (*QueryResult, error) {
	return ci.executor.ExecuteQueryWithCache(ctx, query, args, queryType, ttl)
}

// ExecuteWithoutCache executes a query without caching
func (ci *CacheIntegration) ExecuteWithoutCache(ctx context.Context, query string, args []interface{}) (*QueryResult, error) {
	return ci.executor.ExecuteQueryDirectly(ctx, query, args)
}

// InvalidateBusinessCache invalidates all cache entries for a business
func (ci *CacheIntegration) InvalidateBusinessCache(ctx context.Context, businessID string) error {
	return ci.executor.InvalidateCacheByBusinessID(ctx, businessID)
}

// InvalidateUserCache invalidates all cache entries for a user
func (ci *CacheIntegration) InvalidateUserCache(ctx context.Context, userID string) error {
	return ci.executor.InvalidateCacheByUserID(ctx, userID)
}

// GetPerformanceMetrics returns cache performance metrics
func (ci *CacheIntegration) GetPerformanceMetrics() *CacheMetrics {
	return ci.executor.GetCacheMetrics()
}
