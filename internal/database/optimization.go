package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseOptimizer handles database performance optimization
type DatabaseOptimizer struct {
	db     *sql.DB
	config *OptimizationConfig
}

// OptimizationConfig contains database optimization settings
type OptimizationConfig struct {
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	QueryTimeout    time.Duration `yaml:"query_timeout"`
	EnableIndexing  bool          `yaml:"enable_indexing"`
	EnableProfiling bool          `yaml:"enable_profiling"`
}

// NewDatabaseOptimizer creates a new database optimizer
func NewDatabaseOptimizer(db *sql.DB, config *OptimizationConfig) *DatabaseOptimizer {
	return &DatabaseOptimizer{
		db:     db,
		config: config,
	}
}

// OptimizeConnectionPool optimizes database connection pool settings
func (do *DatabaseOptimizer) OptimizeConnectionPool() error {
	log.Println("Optimizing database connection pool...")

	// Set connection pool parameters
	do.db.SetMaxOpenConns(do.config.MaxOpenConns)
	do.db.SetMaxIdleConns(do.config.MaxIdleConns)
	do.db.SetConnMaxLifetime(do.config.ConnMaxLifetime)
	do.db.SetConnMaxIdleTime(do.config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := do.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connection pool optimized: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
		do.config.MaxOpenConns, do.config.MaxIdleConns, do.config.ConnMaxLifetime, do.config.ConnMaxIdleTime)

	return nil
}

// CreatePerformanceIndexes creates critical performance indexes
func (do *DatabaseOptimizer) CreatePerformanceIndexes() error {
	if !do.config.EnableIndexing {
		log.Println("Indexing disabled, skipping index creation")
		return nil
	}

	log.Println("Creating performance indexes...")

	indexes := []struct {
		name  string
		query string
	}{
		{
			name:  "idx_business_verifications_status",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_verifications_status ON business_verifications(status)",
		},
		{
			name:  "idx_business_verifications_created_at",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_verifications_created_at ON business_verifications(created_at)",
		},
		{
			name:  "idx_business_verifications_status_created",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_verifications_status_created ON business_verifications(status, created_at)",
		},
		{
			name:  "idx_classifications_business_id",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_business_id ON classifications(business_id)",
		},
		{
			name:  "idx_classifications_business_confidence",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_classifications_business_confidence ON classifications(business_id, confidence)",
		},
		{
			name:  "idx_merchants_verification_id",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_merchants_verification_id ON merchants(verification_id)",
		},
		{
			name:  "idx_monitoring_metrics_timestamp",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_monitoring_metrics_timestamp ON monitoring_metrics(timestamp)",
		},
		{
			name:  "idx_pipeline_jobs_status",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pipeline_jobs_status ON pipeline_jobs(status)",
		},
		{
			name:  "idx_pipeline_jobs_created_at",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pipeline_jobs_created_at ON pipeline_jobs(created_at)",
		},
		{
			name:  "idx_pipeline_jobs_status_created",
			query: "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pipeline_jobs_status_created ON pipeline_jobs(status, created_at)",
		},
	}

	for _, index := range indexes {
		log.Printf("Creating index: %s", index.name)

		ctx, cancel := context.WithTimeout(context.Background(), do.config.QueryTimeout)
		_, err := do.db.ExecContext(ctx, index.query)
		cancel()

		if err != nil {
			log.Printf("Warning: Failed to create index %s: %v", index.name, err)
			// Continue with other indexes even if one fails
		} else {
			log.Printf("Successfully created index: %s", index.name)
		}
	}

	log.Println("Performance indexes creation completed")
	return nil
}

// AnalyzeTablePerformance analyzes table performance and provides recommendations
func (do *DatabaseOptimizer) AnalyzeTablePerformance() (*PerformanceAnalysis, error) {
	log.Println("Analyzing table performance...")

	analysis := &PerformanceAnalysis{
		Timestamp: time.Now(),
		Tables:    make(map[string]TableStats),
	}

	// Get table statistics
	tables := []string{
		"business_verifications",
		"classifications",
		"merchants",
		"monitoring_metrics",
		"pipeline_jobs",
	}

	for _, table := range tables {
		stats, err := do.getTableStats(table)
		if err != nil {
			log.Printf("Warning: Failed to get stats for table %s: %v", table, err)
			continue
		}
		analysis.Tables[table] = stats
	}

	// Generate recommendations
	analysis.Recommendations = do.generateRecommendations(analysis)

	log.Println("Table performance analysis completed")
	return analysis, nil
}

// TableStats contains statistics for a database table
type TableStats struct {
	RowCount        int64      `json:"row_count"`
	TableSize       string     `json:"table_size"`
	IndexSize       string     `json:"index_size"`
	TotalSize       string     `json:"total_size"`
	IndexCount      int        `json:"index_count"`
	LastAnalyzed    *time.Time `json:"last_analyzed"`
	SequentialScans int64      `json:"sequential_scans"`
	IndexScans      int64      `json:"index_scans"`
}

// PerformanceAnalysis contains overall performance analysis
type PerformanceAnalysis struct {
	Timestamp       time.Time             `json:"timestamp"`
	Tables          map[string]TableStats `json:"tables"`
	Recommendations []string              `json:"recommendations"`
}

func (do *DatabaseOptimizer) getTableStats(tableName string) (TableStats, error) {
	var stats TableStats

	// Get basic table statistics
	query := `
		SELECT 
			n_tup_ins + n_tup_upd + n_tup_del as row_count,
			pg_size_pretty(pg_total_relation_size($1)) as total_size,
			pg_size_pretty(pg_relation_size($1)) as table_size,
			pg_size_pretty(pg_indexes_size($1)) as index_size,
			(SELECT count(*) FROM pg_indexes WHERE tablename = $1) as index_count,
			last_analyze,
			seq_scan as sequential_scans,
			idx_scan as index_scans
		FROM pg_stat_user_tables 
		WHERE relname = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), do.config.QueryTimeout)
	defer cancel()

	err := do.db.QueryRowContext(ctx, query, tableName).Scan(
		&stats.RowCount,
		&stats.TotalSize,
		&stats.TableSize,
		&stats.IndexSize,
		&stats.IndexCount,
		&stats.LastAnalyzed,
		&stats.SequentialScans,
		&stats.IndexScans,
	)

	if err != nil {
		return stats, fmt.Errorf("failed to get stats for table %s: %w", tableName, err)
	}

	return stats, nil
}

func (do *DatabaseOptimizer) generateRecommendations(analysis *PerformanceAnalysis) []string {
	var recommendations []string

	for tableName, stats := range analysis.Tables {
		// Check for missing indexes
		if stats.IndexCount == 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Table %s has no indexes - consider adding indexes for frequently queried columns", tableName))
		}

		// Check for excessive sequential scans
		if stats.SequentialScans > stats.IndexScans && stats.RowCount > 1000 {
			recommendations = append(recommendations,
				fmt.Sprintf("Table %s has more sequential scans than index scans - consider adding indexes", tableName))
		}

		// Check for large table size
		if stats.RowCount > 100000 {
			recommendations = append(recommendations,
				fmt.Sprintf("Table %s is large (%d rows) - consider partitioning or archiving old data", tableName, stats.RowCount))
		}

		// Check for outdated statistics
		if stats.LastAnalyzed != nil && time.Since(*stats.LastAnalyzed) > 7*24*time.Hour {
			recommendations = append(recommendations,
				fmt.Sprintf("Table %s statistics are outdated (last analyzed: %v) - consider running ANALYZE", tableName, stats.LastAnalyzed))
		}
	}

	return recommendations
}

// OptimizeQueries provides query optimization utilities
type QueryOptimizer struct {
	db *sql.DB
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// PrepareOptimizedStatements prepares optimized SQL statements
func (qo *QueryOptimizer) PrepareOptimizedStatements() (*OptimizedStatements, error) {
	log.Println("Preparing optimized SQL statements...")

	statements := &OptimizedStatements{}

	// Prepare frequently used statements
	var err error

	statements.GetBusinessVerification, err = qo.db.Prepare(`
		SELECT id, business_name, status, created_at, updated_at 
		FROM business_verifications 
		WHERE id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetBusinessVerification: %w", err)
	}

	statements.GetClassifications, err = qo.db.Prepare(`
		SELECT business_id, code, description, confidence, created_at 
		FROM classifications 
		WHERE business_id = $1 
		ORDER BY confidence DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetClassifications: %w", err)
	}

	statements.GetMerchantData, err = qo.db.Prepare(`
		SELECT id, verification_id, name, status, created_at 
		FROM merchants 
		WHERE verification_id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetMerchantData: %w", err)
	}

	statements.GetMonitoringMetrics, err = qo.db.Prepare(`
		SELECT metric_name, metric_value, timestamp 
		FROM monitoring_metrics 
		WHERE timestamp >= $1 AND timestamp <= $2 
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetMonitoringMetrics: %w", err)
	}

	log.Println("Optimized SQL statements prepared successfully")
	return statements, nil
}

// OptimizedStatements contains prepared SQL statements
type OptimizedStatements struct {
	GetBusinessVerification *sql.Stmt
	GetClassifications      *sql.Stmt
	GetMerchantData         *sql.Stmt
	GetMonitoringMetrics    *sql.Stmt
}

// Close closes all prepared statements
func (os *OptimizedStatements) Close() error {
	var errs []error

	if os.GetBusinessVerification != nil {
		if err := os.GetBusinessVerification.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if os.GetClassifications != nil {
		if err := os.GetClassifications.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if os.GetMerchantData != nil {
		if err := os.GetMerchantData.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if os.GetMonitoringMetrics != nil {
		if err := os.GetMonitoringMetrics.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing statements: %v", errs)
	}

	return nil
}

// BatchInsertClassifications performs batch insert for better performance
func (qo *QueryOptimizer) BatchInsertClassifications(classifications []Classification) error {
	if len(classifications) == 0 {
		return nil
	}

	log.Printf("Batch inserting %d classifications...", len(classifications))

	tx, err := qo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO classifications (business_id, code, description, confidence, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch insert statement: %w", err)
	}
	defer stmt.Close()

	for _, classification := range classifications {
		_, err := stmt.Exec(
			classification.BusinessID,
			classification.Code,
			classification.Description,
			classification.Confidence,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully batch inserted %d classifications", len(classifications))
	return nil
}

// Classification represents a business classification
type Classification struct {
	BusinessID  string  `json:"business_id"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}
