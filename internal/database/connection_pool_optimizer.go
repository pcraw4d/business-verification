package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"time"
)

// ConnectionPoolOptimizer provides optimized database connection pool configuration
// for the KYB Platform's classification and risk assessment workloads
type ConnectionPoolOptimizer struct {
	logger *log.Logger
}

// OptimizedPoolConfig represents optimized connection pool settings
type OptimizedPoolConfig struct {
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	MaxConnAttempts int           `json:"max_conn_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// ConnectionPoolMetrics tracks connection pool performance
type ConnectionPoolMetrics struct {
	OpenConnections   int           `json:"open_connections"`
	IdleConnections   int           `json:"idle_connections"`
	InUseConnections  int           `json:"in_use_connections"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// NewConnectionPoolOptimizer creates a new connection pool optimizer
func NewConnectionPoolOptimizer(logger *log.Logger) *ConnectionPoolOptimizer {
	return &ConnectionPoolOptimizer{
		logger: logger,
	}
}

// GetOptimizedPoolConfig returns optimized connection pool configuration
// based on system resources and workload characteristics
func (cpo *ConnectionPoolOptimizer) GetOptimizedPoolConfig() *OptimizedPoolConfig {
	// Get system information
	numCPU := runtime.NumCPU()

	// Calculate optimal settings based on system resources and workload
	// For classification and risk assessment workloads, we need:
	// - Higher connection limits for concurrent processing
	// - Longer connection lifetimes for complex queries
	// - Appropriate idle connection management

	maxOpenConns := cpo.calculateMaxOpenConns(numCPU)
	maxIdleConns := cpo.calculateMaxIdleConns(maxOpenConns)
	connMaxLifetime := cpo.calculateConnMaxLifetime()
	connMaxIdleTime := cpo.calculateConnMaxIdleTime()

	return &OptimizedPoolConfig{
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		MaxConnAttempts: 3,
		RetryDelay:      5 * time.Second,
	}
}

// calculateMaxOpenConns calculates optimal maximum open connections
func (cpo *ConnectionPoolOptimizer) calculateMaxOpenConns(numCPU int) int {
	// Base calculation: 2 * CPU cores for I/O bound operations
	// Add extra connections for concurrent classification and risk assessment
	baseConns := numCPU * 2

	// Add connections for concurrent operations:
	// - Classification processing: +10
	// - Risk assessment: +10
	// - API requests: +20
	// - Background tasks: +10
	additionalConns := 50

	totalConns := baseConns + additionalConns

	// Cap at reasonable maximum for Supabase
	if totalConns > 200 {
		totalConns = 200
	}

	// Ensure minimum for small systems
	if totalConns < 25 {
		totalConns = 25
	}

	return totalConns
}

// calculateMaxIdleConns calculates optimal maximum idle connections
func (cpo *ConnectionPoolOptimizer) calculateMaxIdleConns(maxOpenConns int) int {
	// Keep 20% of max connections as idle for quick response
	idleConns := maxOpenConns / 5

	// Ensure minimum idle connections
	if idleConns < 5 {
		idleConns = 5
	}

	// Cap at reasonable maximum
	if idleConns > 50 {
		idleConns = 50
	}

	return idleConns
}

// calculateConnMaxLifetime calculates optimal connection maximum lifetime
func (cpo *ConnectionPoolOptimizer) calculateConnMaxLifetime() time.Duration {
	// For classification and risk assessment workloads:
	// - Longer lifetime for complex queries
	// - Balance between performance and resource usage
	// - Account for potential long-running ML operations

	// Base lifetime: 5 minutes
	baseLifetime := 5 * time.Minute

	// Add extra time for complex operations
	// - ML model inference: +2 minutes
	// - Risk assessment processing: +1 minute
	// - Classification processing: +1 minute
	additionalTime := 4 * time.Minute

	totalLifetime := baseLifetime + additionalTime

	// Cap at reasonable maximum
	if totalLifetime > 15*time.Minute {
		totalLifetime = 15 * time.Minute
	}

	return totalLifetime
}

// calculateConnMaxIdleTime calculates optimal connection maximum idle time
func (cpo *ConnectionPoolOptimizer) calculateConnMaxIdleTime() time.Duration {
	// For classification workloads:
	// - Shorter idle time to free resources quickly
	// - Balance between connection reuse and resource usage

	// Base idle time: 1 minute
	baseIdleTime := 1 * time.Minute

	// Add extra time for burst handling
	// - Handle classification bursts: +1 minute
	// - Handle risk assessment bursts: +1 minute
	additionalTime := 2 * time.Minute

	totalIdleTime := baseIdleTime + additionalTime

	// Cap at reasonable maximum
	if totalIdleTime > 5*time.Minute {
		totalIdleTime = 5 * time.Minute
	}

	return totalIdleTime
}

// OptimizeConnectionPool applies optimized settings to a database connection
func (cpo *ConnectionPoolOptimizer) OptimizeConnectionPool(ctx context.Context, db *sql.DB) error {
	config := cpo.GetOptimizedPoolConfig()

	cpo.logger.Printf("Optimizing connection pool with settings: %+v", config)

	// Apply optimized settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database after optimization: %w", err)
	}

	cpo.logger.Printf("Connection pool optimized successfully")
	return nil
}

// GetPerformanceMetrics retrieves current connection pool performance metrics
func (cpo *ConnectionPoolOptimizer) GetPerformanceMetrics(db *sql.DB) (*ConnectionPoolMetrics, error) {
	stats := db.Stats()

	return &ConnectionPoolMetrics{
		OpenConnections:   stats.OpenConnections,
		IdleConnections:   stats.Idle,
		InUseConnections:  stats.InUse,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		LastUpdated:       time.Now(),
	}, nil
}

// MonitorConnectionPool continuously monitors connection pool performance
func (cpo *ConnectionPoolOptimizer) MonitorConnectionPool(ctx context.Context, db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := cpo.GetPerformanceMetrics(db)
			if err != nil {
				cpo.logger.Printf("Failed to get connection pool metrics: %v", err)
				continue
			}

			// Log performance metrics
			cpo.logger.Printf("Connection Pool Metrics: Open=%d, Idle=%d, InUse=%d, WaitCount=%d, WaitDuration=%v",
				metrics.OpenConnections, metrics.IdleConnections, metrics.InUseConnections,
				metrics.WaitCount, metrics.WaitDuration)

			// Check for performance issues
			if metrics.WaitCount > 100 {
				cpo.logger.Printf("WARNING: High wait count detected: %d", metrics.WaitCount)
			}

			if metrics.WaitDuration > 1*time.Second {
				cpo.logger.Printf("WARNING: High wait duration detected: %v", metrics.WaitDuration)
			}

			// Check connection utilization
			utilization := float64(metrics.InUseConnections) / float64(metrics.OpenConnections) * 100
			if utilization > 80 {
				cpo.logger.Printf("WARNING: High connection utilization: %.2f%%", utilization)
			}
		}
	}
}

// ValidateConnectionPool validates that the connection pool is properly configured
func (cpo *ConnectionPoolOptimizer) ValidateConnectionPool(db *sql.DB) error {
	stats := db.Stats()
	config := cpo.GetOptimizedPoolConfig()

	// Check if settings are applied correctly
	if stats.MaxOpenConnections != config.MaxOpenConns {
		return fmt.Errorf("max open connections mismatch: expected %d, got %d",
			config.MaxOpenConns, stats.MaxOpenConnections)
	}

	// Check for connection leaks
	if stats.OpenConnections > config.MaxOpenConns {
		return fmt.Errorf("connection leak detected: %d connections open (max: %d)",
			stats.OpenConnections, config.MaxOpenConns)
	}

	// Check for high wait times
	if stats.WaitCount > 1000 {
		return fmt.Errorf("excessive wait count detected: %d", stats.WaitCount)
	}

	return nil
}

// GetConnectionPoolRecommendations provides recommendations for connection pool optimization
func (cpo *ConnectionPoolOptimizer) GetConnectionPoolRecommendations(db *sql.DB) []string {
	var recommendations []string
	stats := db.Stats()
	config := cpo.GetOptimizedPoolConfig()

	// Check connection utilization
	utilization := float64(stats.InUse) / float64(stats.OpenConnections) * 100
	if utilization > 90 {
		recommendations = append(recommendations,
			"Consider increasing max open connections - utilization is very high")
	} else if utilization < 20 {
		recommendations = append(recommendations,
			"Consider decreasing max open connections - utilization is low")
	}

	// Check wait times
	if stats.WaitCount > 100 {
		recommendations = append(recommendations,
			"High wait count detected - consider increasing connection pool size")
	}

	// Check connection lifetime
	if stats.MaxLifetimeClosed > 100 {
		recommendations = append(recommendations,
			"Many connections closed due to max lifetime - consider increasing ConnMaxLifetime")
	}

	// Check idle connections
	if stats.MaxIdleClosed > 100 {
		recommendations = append(recommendations,
			"Many idle connections closed - consider increasing MaxIdleConns or ConnMaxIdleTime")
	}

	// Check for connection leaks
	if stats.OpenConnections > config.MaxOpenConns {
		recommendations = append(recommendations,
			"Connection leak detected - investigate application code for unclosed connections")
	}

	return recommendations
}

// CreateOptimizedPostgresDB creates a new PostgreSQL database instance with optimized connection pool
func CreateOptimizedPostgresDB(cfg *DatabaseConfig, logger *log.Logger) (*PostgresDB, error) {
	// Create database instance
	db := NewPostgresDB(cfg)

	// Create connection pool optimizer
	optimizer := NewConnectionPoolOptimizer(logger)

	// Connect to database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Optimize connection pool
	if err := optimizer.OptimizeConnectionPool(ctx, db.GetDB()); err != nil {
		return nil, fmt.Errorf("failed to optimize connection pool: %w", err)
	}

	// Validate connection pool
	if err := optimizer.ValidateConnectionPool(db.GetDB()); err != nil {
		logger.Printf("Connection pool validation warning: %v", err)
	}

	// Start monitoring (in background)
	go optimizer.MonitorConnectionPool(ctx, db.GetDB(), 30*time.Second)

	logger.Printf("Optimized PostgreSQL database instance created successfully")
	return db, nil
}

// GetConnectionPoolHealth returns the health status of the connection pool
func (cpo *ConnectionPoolOptimizer) GetConnectionPoolHealth(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	config := cpo.GetOptimizedPoolConfig()

	utilization := float64(stats.InUse) / float64(stats.OpenConnections) * 100

	health := map[string]interface{}{
		"status": "healthy",
		"metrics": map[string]interface{}{
			"open_connections":    stats.OpenConnections,
			"idle_connections":    stats.Idle,
			"in_use_connections":  stats.InUse,
			"utilization_percent": utilization,
			"wait_count":          stats.WaitCount,
			"wait_duration":       stats.WaitDuration.String(),
		},
		"configuration": map[string]interface{}{
			"max_open_conns":     config.MaxOpenConns,
			"max_idle_conns":     config.MaxIdleConns,
			"conn_max_lifetime":  config.ConnMaxLifetime.String(),
			"conn_max_idle_time": config.ConnMaxIdleTime.String(),
		},
		"recommendations": cpo.GetConnectionPoolRecommendations(db),
	}

	// Determine health status
	if stats.WaitCount > 1000 || utilization > 95 {
		health["status"] = "critical"
	} else if stats.WaitCount > 100 || utilization > 80 {
		health["status"] = "warning"
	}

	return health
}
