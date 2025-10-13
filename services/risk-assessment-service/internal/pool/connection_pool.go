package pool

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// PoolConfig represents connection pool configuration
type PoolConfig struct {
	MaxConnections     int           `json:"max_connections"`
	MinConnections     int           `json:"min_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`
	MaxLifetime        time.Duration `json:"max_lifetime"`
	HealthCheckPeriod  time.Duration `json:"health_check_period"`
	RetryAttempts      int           `json:"retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay"`
}

// PoolMetrics represents connection pool metrics
type PoolMetrics struct {
	ActiveConnections   int           `json:"active_connections"`
	IdleConnections     int           `json:"idle_connections"`
	TotalConnections    int           `json:"total_connections"`
	WaitCount           int64         `json:"wait_count"`
	WaitDuration        time.Duration `json:"wait_duration"`
	MaxIdleClosed       int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed   int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed   int64         `json:"max_lifetime_closed"`
	ConnectionsCreated  int64         `json:"connections_created"`
	ConnectionsClosed   int64         `json:"connections_closed"`
	LastHealthCheck     time.Time     `json:"last_health_check"`
	HealthCheckFailures int64         `json:"health_check_failures"`
}

// ConnectionPool manages database connections with pooling
type ConnectionPool struct {
	db      *sql.DB
	config  *PoolConfig
	logger  *zap.Logger
	metrics *PoolMetrics
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(dsn string, config *PoolConfig, logger *zap.Logger) (*ConnectionPool, error) {
	if config == nil {
		config = &PoolConfig{}
	}

	// Set default values
	if config.MaxConnections == 0 {
		config.MaxConnections = 25
	}
	if config.MinConnections == 0 {
		config.MinConnections = 5
	}
	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = 5
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 30 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 1 * time.Minute
	}
	if config.MaxLifetime == 0 {
		config.MaxLifetime = 5 * time.Minute
	}
	if config.HealthCheckPeriod == 0 {
		config.HealthCheckPeriod = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.MaxLifetime)
	db.SetConnMaxIdleTime(config.IdleTimeout)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pool := &ConnectionPool{
		db:      db,
		config:  config,
		logger:  logger,
		metrics: &PoolMetrics{},
		ctx:     context.Background(),
	}

	// Start health check routine
	go pool.healthCheckRoutine()

	logger.Info("Connection pool initialized successfully",
		zap.Int("max_connections", config.MaxConnections),
		zap.Int("max_idle_connections", config.MaxIdleConnections),
		zap.Duration("max_lifetime", config.MaxLifetime))

	return pool, nil
}

// GetConnection returns a database connection from the pool
func (p *ConnectionPool) GetConnection(ctx context.Context) (*sql.Conn, error) {
	start := time.Now()
	defer func() {
		p.mu.Lock()
		p.metrics.WaitCount++
		p.metrics.WaitDuration += time.Since(start)
		p.mu.Unlock()
	}()

	conn, err := p.db.Conn(ctx)
	if err != nil {
		p.mu.Lock()
		p.metrics.ConnectionsClosed++
		p.mu.Unlock()
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	p.mu.Lock()
	p.metrics.ConnectionsCreated++
	p.mu.Unlock()

	return conn, nil
}

// GetDB returns the underlying database instance
func (p *ConnectionPool) GetDB() *sql.DB {
	return p.db
}

// Exec executes a query without returning rows
func (p *ConnectionPool) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (p *ConnectionPool) Query(ctx context.Context, query string, args ...interface{}) (*pooledRows, error) {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Wrap rows to close connection when rows are closed
	return &pooledRows{Rows: rows, conn: conn}, nil
}

// QueryRow executes a query that returns a single row
func (p *ConnectionPool) QueryRow(ctx context.Context, query string, args ...interface{}) *pooledRow {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return &pooledRow{Row: &sql.Row{}, conn: nil}
	}

	row := conn.QueryRowContext(ctx, query, args...)
	return &pooledRow{Row: row, conn: conn}
}

// Begin starts a transaction
func (p *ConnectionPool) Begin(ctx context.Context) (*pooledTx, error) {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Wrap transaction to close connection when transaction is done
	return &pooledTx{Tx: tx, conn: conn}, nil
}

// Prepare creates a prepared statement
func (p *ConnectionPool) Prepare(ctx context.Context, query string) (*pooledStmt, error) {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	stmt, err := conn.PrepareContext(ctx, query)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Wrap statement to close connection when statement is closed
	return &pooledStmt{Stmt: stmt, conn: conn}, nil
}

// GetMetrics returns current pool metrics
func (p *ConnectionPool) GetMetrics() *PoolMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Get current stats from database
	stats := p.db.Stats()

	metrics := &PoolMetrics{
		ActiveConnections:   stats.OpenConnections - stats.Idle,
		IdleConnections:     stats.Idle,
		TotalConnections:    stats.OpenConnections,
		WaitCount:           p.metrics.WaitCount,
		WaitDuration:        p.metrics.WaitDuration,
		MaxIdleClosed:       stats.MaxIdleClosed,
		MaxIdleTimeClosed:   stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:   stats.MaxLifetimeClosed,
		ConnectionsCreated:  p.metrics.ConnectionsCreated,
		ConnectionsClosed:   p.metrics.ConnectionsClosed,
		LastHealthCheck:     p.metrics.LastHealthCheck,
		HealthCheckFailures: p.metrics.HealthCheckFailures,
	}

	return metrics
}

// ResetMetrics resets pool metrics
func (p *ConnectionPool) ResetMetrics() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics = &PoolMetrics{}
}

// Health checks pool health
func (p *ConnectionPool) Health(ctx context.Context) error {
	conn, err := p.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.PingContext(ctx)
}

// Close closes the connection pool
func (p *ConnectionPool) Close() error {
	p.cancel()
	return p.db.Close()
}

// healthCheckRoutine periodically checks pool health
func (p *ConnectionPool) healthCheckRoutine() {
	ticker := time.NewTicker(p.config.HealthCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check on the pool
func (p *ConnectionPool) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.Health(ctx); err != nil {
		p.mu.Lock()
		p.metrics.HealthCheckFailures++
		p.mu.Unlock()

		p.logger.Error("Connection pool health check failed",
			zap.Error(err),
			zap.Int64("failures", p.metrics.HealthCheckFailures))
	} else {
		p.mu.Lock()
		p.metrics.LastHealthCheck = time.Now()
		p.mu.Unlock()
	}
}

// pooledRows wraps sql.Rows to close connection when rows are closed
type pooledRows struct {
	*sql.Rows
	conn *sql.Conn
}

func (pr *pooledRows) Close() error {
	err := pr.Rows.Close()
	pr.conn.Close()
	return err
}

// pooledRow wraps sql.Row to close connection when row is scanned
type pooledRow struct {
	*sql.Row
	conn *sql.Conn
}

func (pr *pooledRow) Scan(dest ...interface{}) error {
	err := pr.Row.Scan(dest...)
	pr.conn.Close()
	return err
}

// pooledTx wraps sql.Tx to close connection when transaction is done
type pooledTx struct {
	*sql.Tx
	conn *sql.Conn
}

func (pt *pooledTx) Commit() error {
	err := pt.Tx.Commit()
	pt.conn.Close()
	return err
}

func (pt *pooledTx) Rollback() error {
	err := pt.Tx.Rollback()
	pt.conn.Close()
	return err
}

// pooledStmt wraps sql.Stmt to close connection when statement is closed
type pooledStmt struct {
	*sql.Stmt
	conn *sql.Conn
}

func (ps *pooledStmt) Close() error {
	err := ps.Stmt.Close()
	ps.conn.Close()
	return err
}
