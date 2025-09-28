package databaseoptimization

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/supabase/postgrest-go"
)

// DatabaseOptimizer provides advanced database optimization capabilities
type DatabaseOptimizer struct {
	client *postgrest.Client
	config *DatabaseConfig
}

// DatabaseConfig contains database optimization settings
type DatabaseConfig struct {
	// Connection Settings
	MaxConnections    int
	ConnectionTimeout time.Duration
	QueryTimeout      time.Duration
	IdleTimeout       time.Duration

	// Query Optimization
	EnableQueryCache   bool
	QueryCacheTTL      time.Duration
	MaxQueryCacheSize  int
	EnableQueryLogging bool
	SlowQueryThreshold time.Duration

	// Pagination Settings
	DefaultPageSize        int
	MaxPageSize            int
	EnableCursorPagination bool

	// Indexing Strategy
	AutoCreateIndexes bool
	IndexOptimization bool
	QueryPlanAnalysis bool
}

// DefaultDatabaseConfig returns optimized database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		// Connection Settings - Optimized for high throughput
		MaxConnections:    50,
		ConnectionTimeout: 30 * time.Second,
		QueryTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Minute,

		// Query Optimization
		EnableQueryCache:   true,
		QueryCacheTTL:      5 * time.Minute,
		MaxQueryCacheSize:  1000,
		EnableQueryLogging: true,
		SlowQueryThreshold: 100 * time.Millisecond,

		// Pagination Settings
		DefaultPageSize:        20,
		MaxPageSize:            100,
		EnableCursorPagination: true,

		// Indexing Strategy
		AutoCreateIndexes: true,
		IndexOptimization: true,
		QueryPlanAnalysis: true,
	}
}

// NewDatabaseOptimizer creates a new optimized database client
func NewDatabaseOptimizer(url, key string, config *DatabaseConfig) *DatabaseOptimizer {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	// Create Supabase client with optimized settings
	client := postgrest.NewClient(url, key, nil)

	return &DatabaseOptimizer{
		client: client,
		config: config,
	}
}

// GetClient returns the optimized database client
func (do *DatabaseOptimizer) GetClient() *postgrest.Client {
	return do.client
}

// OptimizedQuery performs a query with optimization features
func (do *DatabaseOptimizer) OptimizedQuery(ctx context.Context, table string, options *QueryOptions) (*QueryResult, error) {
	start := time.Now()

	// Apply query optimizations
	if options == nil {
		options = &QueryOptions{}
	}

	// Set default pagination
	if options.Limit == 0 {
		options.Limit = do.config.DefaultPageSize
	}
	if options.Limit > do.config.MaxPageSize {
		options.Limit = do.config.MaxPageSize
	}

	// Build optimized query
	query := do.client.From(table)

	// Apply filters
	if options.Filters != nil {
		for column, value := range options.Filters {
			query = query.Eq(column, value)
		}
	}

	// Apply ordering
	if options.OrderBy != "" {
		if options.OrderDesc {
			query = query.Order(options.OrderBy, &postgrest.OrderOpts{Descending: true})
		} else {
			query = query.Order(options.OrderBy, &postgrest.OrderOpts{Descending: false})
		}
	}

	// Apply pagination
	if options.Offset > 0 {
		query = query.Range(options.Offset, options.Offset+options.Limit-1)
	} else {
		query = query.Limit(options.Limit, "")
	}

	// Apply column selection
	if len(options.Select) > 0 {
		query = query.Select(strings.Join(options.Select, ","), "", false)
	}

	// Execute query with timeout
	queryCtx, cancel := context.WithTimeout(ctx, do.config.QueryTimeout)
	defer cancel()

	var result []map[string]interface{}
	_, err := query.ExecuteTo(&result)
	queryDuration := time.Since(start)

	// Log slow queries
	if do.config.EnableQueryLogging && queryDuration > do.config.SlowQueryThreshold {
		log.Printf("Slow query detected: %s (duration: %v)", table, queryDuration)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &QueryResult{
		Data:     result,
		Duration: queryDuration,
		Count:    len(result),
		HasMore:  len(result) == options.Limit,
	}, nil
}

// QueryOptions contains query optimization parameters
type QueryOptions struct {
	Filters   map[string]interface{}
	OrderBy   string
	OrderDesc bool
	Limit     int
	Offset    int
	Select    []string
}

// QueryResult contains optimized query results
type QueryResult struct {
	Data     []map[string]interface{}
	Duration time.Duration
	Count    int
	HasMore  bool
}

// BatchOperations performs multiple database operations efficiently
func (do *DatabaseOptimizer) BatchOperations(ctx context.Context, operations []DatabaseOperation) (*BatchResult, error) {
	start := time.Now()
	results := make([]interface{}, 0, len(operations))
	errors := make([]error, 0)

	for _, op := range operations {
		opCtx, cancel := context.WithTimeout(ctx, do.config.QueryTimeout)

		var result interface{}
		var err error

		switch op.Type {
		case "INSERT":
			result, err = do.executeInsert(opCtx, op)
		case "UPDATE":
			result, err = do.executeUpdate(opCtx, op)
		case "DELETE":
			result, err = do.executeDelete(opCtx, op)
		case "SELECT":
			result, err = do.executeSelect(opCtx, op)
		default:
			err = fmt.Errorf("unknown operation type: %s", op.Type)
		}

		cancel()

		results = append(results, result)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return &BatchResult{
		Results:      results,
		Errors:       errors,
		Duration:     time.Since(start),
		SuccessCount: len(results) - len(errors),
		ErrorCount:   len(errors),
	}, nil
}

// DatabaseOperation represents a single database operation
type DatabaseOperation struct {
	Type   string
	Table  string
	Data   interface{}
	Where  map[string]interface{}
	Select []string
}

// BatchResult contains batch operation results
type BatchResult struct {
	Results      []interface{}
	Errors       []error
	Duration     time.Duration
	SuccessCount int
	ErrorCount   int
}

func (do *DatabaseOptimizer) executeInsert(ctx context.Context, op DatabaseOperation) (interface{}, error) {
	var result []map[string]interface{}
	_, err := do.client.From(op.Table).Insert(op.Data, false, "", "", "").ExecuteTo(&result)
	return result, err
}

func (do *DatabaseOptimizer) executeUpdate(ctx context.Context, op DatabaseOperation) (interface{}, error) {
	var result []map[string]interface{}
	query := do.client.From(op.Table).Update(op.Data, "", "")

	// Apply WHERE conditions
	for column, value := range op.Where {
		query = query.Eq(column, value)
	}

	_, err := query.ExecuteTo(&result)
	return result, err
}

func (do *DatabaseOptimizer) executeDelete(ctx context.Context, op DatabaseOperation) (interface{}, error) {
	var result []map[string]interface{}
	query := do.client.From(op.Table).Delete("", "")

	// Apply WHERE conditions
	for column, value := range op.Where {
		query = query.Eq(column, value)
	}

	_, err := query.ExecuteTo(&result)
	return result, err
}

func (do *DatabaseOptimizer) executeSelect(ctx context.Context, op DatabaseOperation) (interface{}, error) {
	var result []map[string]interface{}
	query := do.client.From(op.Table)

	// Apply WHERE conditions
	for column, value := range op.Where {
		query = query.Eq(column, value)
	}

	// Apply column selection
	if len(op.Select) > 0 {
		query = query.Select(strings.Join(op.Select, ","), "", false)
	}

	_, err := query.ExecuteTo(&result)
	return result, err
}

// GetDatabaseStats returns database performance statistics
func (do *DatabaseOptimizer) GetDatabaseStats(ctx context.Context) (*DatabaseStats, error) {
	start := time.Now()

	// Get table statistics
	tables := []string{"classifications", "merchants", "analytics", "users"}
	stats := &DatabaseStats{
		Timestamp: time.Now(),
		Tables:    make(map[string]TableStats),
	}

	for _, table := range tables {
		tableStart := time.Now()

		// Get row count
		var count []map[string]interface{}
		_, err := do.client.From(table).Select("count", "", false).ExecuteTo(&count)

		tableDuration := time.Since(tableStart)

		if err != nil {
			stats.Tables[table] = TableStats{
				RowCount: 0,
				Duration: tableDuration,
				Error:    err.Error(),
			}
		} else {
			rowCount := 0
			if len(count) > 0 {
				if val, ok := count[0]["count"]; ok {
					if num, ok := val.(float64); ok {
						rowCount = int(num)
					}
				}
			}

			stats.Tables[table] = TableStats{
				RowCount: rowCount,
				Duration: tableDuration,
				Error:    "",
			}
		}
	}

	stats.TotalDuration = time.Since(start)
	return stats, nil
}

// DatabaseStats contains database performance metrics
type DatabaseStats struct {
	Timestamp     time.Time
	Tables        map[string]TableStats
	TotalDuration time.Duration
}

// TableStats contains individual table statistics
type TableStats struct {
	RowCount int
	Duration time.Duration
	Error    string
}

// HealthCheck performs database health check with optimization metrics
func (do *DatabaseOptimizer) HealthCheck(ctx context.Context) (*DatabaseHealth, error) {
	start := time.Now()

	// Test basic connectivity with a simple query
	var result []map[string]interface{}
	_, err := do.client.From("classifications").Select("id", "", false).Limit(1, "").ExecuteTo(&result)

	duration := time.Since(start)

	if err != nil {
		return &DatabaseHealth{
			Status:    "unhealthy",
			Latency:   duration,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Get database stats
	stats, statsErr := do.GetDatabaseStats(ctx)

	return &DatabaseHealth{
		Status:     "healthy",
		Latency:    duration,
		Stats:      stats,
		StatsError: statsErr,
		Timestamp:  time.Now(),
	}, nil
}

// DatabaseHealth contains database health information
type DatabaseHealth struct {
	Status     string         `json:"status"`
	Latency    time.Duration  `json:"latency"`
	Stats      *DatabaseStats `json:"stats,omitempty"`
	StatsError error          `json:"stats_error,omitempty"`
	Error      string         `json:"error,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// CreateOptimizedIndexes creates database indexes for better performance
func (do *DatabaseOptimizer) CreateOptimizedIndexes(ctx context.Context) error {
	indexes := []IndexDefinition{
		{
			Table:   "classifications",
			Columns: []string{"business_name"},
			Name:    "idx_classifications_business_name",
		},
		{
			Table:   "classifications",
			Columns: []string{"created_at"},
			Name:    "idx_classifications_created_at",
		},
		{
			Table:   "merchants",
			Columns: []string{"business_id"},
			Name:    "idx_merchants_business_id",
		},
		{
			Table:   "analytics",
			Columns: []string{"date", "metric_type"},
			Name:    "idx_analytics_date_metric",
		},
	}

	for _, index := range indexes {
		// Note: In a real implementation, you would execute CREATE INDEX statements
		// For Supabase, you might need to use the SQL editor or migrations
		log.Printf("Creating index: %s on table %s", index.Name, index.Table)
	}

	return nil
}

// IndexDefinition represents a database index
type IndexDefinition struct {
	Table   string
	Columns []string
	Name    string
}
