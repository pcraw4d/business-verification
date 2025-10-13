package performance

import (
	"context"
	"database/sql"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceMonitor monitors system performance metrics
type PerformanceMonitor struct {
	db         *sql.DB
	cache      CacheMonitor
	pool       PoolMonitor
	query      QueryMonitor
	logger     *zap.Logger
	metrics    *SystemMetrics
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	collectors []MetricsCollector
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	Timestamp       time.Time       `json:"timestamp"`
	CPUUsage        float64         `json:"cpu_usage"`
	MemoryUsage     MemoryMetrics   `json:"memory_usage"`
	DatabaseMetrics DatabaseMetrics `json:"database_metrics"`
	CacheMetrics    CacheMetrics    `json:"cache_metrics"`
	PoolMetrics     PoolMetrics     `json:"pool_metrics"`
	QueryMetrics    QueryMetrics    `json:"query_metrics"`
	RequestMetrics  RequestMetrics  `json:"request_metrics"`
	ErrorMetrics    ErrorMetrics    `json:"error_metrics"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	AllocatedMB   float64   `json:"allocated_mb"`
	TotalAllocMB  float64   `json:"total_alloc_mb"`
	SysMB         float64   `json:"sys_mb"`
	NumGC         uint32    `json:"num_gc"`
	GCPauseMS     float64   `json:"gc_pause_ms"`
	HeapObjects   uint64    `json:"heap_objects"`
	StackInUseMB  float64   `json:"stack_in_use_mb"`
	StackSysMB    float64   `json:"stack_sys_mb"`
	MSpanInUseMB  float64   `json:"mspan_in_use_mb"`
	MSpanSysMB    float64   `json:"mspan_sys_mb"`
	MCacheInUseMB float64   `json:"mcache_in_use_mb"`
	MCacheSysMB   float64   `json:"mcache_sys_mb"`
	BuckHashSysMB float64   `json:"buck_hash_sys_mb"`
	GCSysMB       float64   `json:"gc_sys_mb"`
	OtherSysMB    float64   `json:"other_sys_mb"`
	NextGC        uint64    `json:"next_gc"`
	LastGC        time.Time `json:"last_gc"`
	PauseTotalNS  uint64    `json:"pause_total_ns"`
	NumForcedGC   uint32    `json:"num_forced_gc"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
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

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Sets           int64         `json:"sets"`
	Deletes        int64         `json:"deletes"`
	Errors         int64         `json:"errors"`
	TotalRequests  int64         `json:"total_requests"`
	HitRate        float64       `json:"hit_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUpdated    time.Time     `json:"last_updated"`
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

// QueryMetrics represents query performance metrics
type QueryMetrics struct {
	QueryCount     int64         `json:"query_count"`
	CacheHits      int64         `json:"cache_hits"`
	CacheMisses    int64         `json:"cache_misses"`
	AverageLatency time.Duration `json:"average_latency"`
	SlowQueries    int64         `json:"slow_queries"`
	ErrorCount     int64         `json:"error_count"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// RequestMetrics represents HTTP request metrics
type RequestMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`
	MaxLatency         time.Duration `json:"max_latency"`
	MinLatency         time.Duration `json:"min_latency"`
	RequestsPerSecond  float64       `json:"requests_per_second"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// ErrorMetrics represents error metrics
type ErrorMetrics struct {
	TotalErrors      int64            `json:"total_errors"`
	ErrorRate        float64          `json:"error_rate"`
	ErrorsByType     map[string]int64 `json:"errors_by_type"`
	ErrorsByEndpoint map[string]int64 `json:"errors_by_endpoint"`
	LastError        time.Time        `json:"last_error"`
	LastUpdated      time.Time        `json:"last_updated"`
}

// Monitor interfaces
type CacheMonitor interface {
	GetMetrics() *CacheMetrics
}

type PoolMonitor interface {
	GetMetrics() *PoolMetrics
}

type QueryMonitor interface {
	GetMetrics() *QueryMetrics
}

// MetricsCollector interface for collecting metrics
type MetricsCollector interface {
	Collect(ctx context.Context) (interface{}, error)
	GetName() string
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(db *sql.DB, cache CacheMonitor, pool PoolMonitor, query QueryMonitor, logger *zap.Logger) *PerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &PerformanceMonitor{
		db:         db,
		cache:      cache,
		pool:       pool,
		query:      query,
		logger:     logger,
		metrics:    &SystemMetrics{},
		ctx:        ctx,
		cancel:     cancel,
		collectors: []MetricsCollector{},
	}

	// Add default collectors
	monitor.AddCollector(&MemoryCollector{})
	monitor.AddCollector(&DatabaseCollector{db: db})
	monitor.AddCollector(&RuntimeCollector{})

	return monitor
}

// AddCollector adds a metrics collector
func (pm *PerformanceMonitor) AddCollector(collector MetricsCollector) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.collectors = append(pm.collectors, collector)
}

// Start starts the performance monitoring
func (pm *PerformanceMonitor) Start(interval time.Duration) {
	go pm.collectMetrics(interval)
	pm.logger.Info("Performance monitoring started", zap.Duration("interval", interval))
}

// Stop stops the performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.cancel()
	pm.logger.Info("Performance monitoring stopped")
}

// GetMetrics returns current system metrics
func (pm *PerformanceMonitor) GetMetrics() *SystemMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.metrics
}

// GetHealthStatus returns the health status of the system
func (pm *PerformanceMonitor) GetHealthStatus() *HealthStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	status := &HealthStatus{
		Overall: "healthy",
		Checks:  make(map[string]CheckStatus),
	}

	// Check database health
	if pm.metrics.DatabaseMetrics.HealthCheckFailures > 10 {
		status.Checks["database"] = CheckStatus{
			Status:  "unhealthy",
			Message: "High number of health check failures",
		}
		status.Overall = "degraded"
	} else {
		status.Checks["database"] = CheckStatus{
			Status:  "healthy",
			Message: "Database is responding normally",
		}
	}

	// Check cache health
	if pm.metrics.CacheMetrics.HitRate < 0.5 {
		status.Checks["cache"] = CheckStatus{
			Status:  "degraded",
			Message: "Low cache hit rate",
		}
		if status.Overall == "healthy" {
			status.Overall = "degraded"
		}
	} else {
		status.Checks["cache"] = CheckStatus{
			Status:  "healthy",
			Message: "Cache is performing well",
		}
	}

	// Check memory health
	if pm.metrics.MemoryUsage.AllocatedMB > 1000 {
		status.Checks["memory"] = CheckStatus{
			Status:  "degraded",
			Message: "High memory usage",
		}
		if status.Overall == "healthy" {
			status.Overall = "degraded"
		}
	} else {
		status.Checks["memory"] = CheckStatus{
			Status:  "healthy",
			Message: "Memory usage is normal",
		}
	}

	// Check error rate
	if pm.metrics.ErrorMetrics.ErrorRate > 0.01 {
		status.Checks["errors"] = CheckStatus{
			Status:  "unhealthy",
			Message: "High error rate",
		}
		status.Overall = "unhealthy"
	} else {
		status.Checks["errors"] = CheckStatus{
			Status:  "healthy",
			Message: "Error rate is acceptable",
		}
	}

	return status
}

// ResetMetrics resets all performance metrics
func (pm *PerformanceMonitor) ResetMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Reset all metrics to zero values
	pm.metrics = &SystemMetrics{
		Timestamp:       time.Now(),
		CPUUsage:        0,
		MemoryUsage:     MemoryMetrics{},
		DatabaseMetrics: DatabaseMetrics{},
		CacheMetrics:    CacheMetrics{},
		PoolMetrics:     PoolMetrics{},
		QueryMetrics:    QueryMetrics{},
		RequestMetrics:  RequestMetrics{},
		ErrorMetrics:    ErrorMetrics{},
	}

	pm.logger.Info("Performance metrics reset")
}

// collectMetrics collects metrics from all collectors
func (pm *PerformanceMonitor) collectMetrics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.collectAllMetrics()
		}
	}
}

// collectAllMetrics collects metrics from all sources
func (pm *PerformanceMonitor) collectAllMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update timestamp
	pm.metrics.Timestamp = time.Now()

	// Collect memory metrics
	pm.collectMemoryMetrics()

	// Collect database metrics
	if pm.pool != nil {
		poolMetrics := pm.pool.GetMetrics()
		pm.metrics.DatabaseMetrics = DatabaseMetrics{
			ActiveConnections:   poolMetrics.ActiveConnections,
			IdleConnections:     poolMetrics.IdleConnections,
			TotalConnections:    poolMetrics.TotalConnections,
			WaitCount:           poolMetrics.WaitCount,
			WaitDuration:        poolMetrics.WaitDuration,
			MaxIdleClosed:       poolMetrics.MaxIdleClosed,
			MaxIdleTimeClosed:   poolMetrics.MaxIdleTimeClosed,
			MaxLifetimeClosed:   poolMetrics.MaxLifetimeClosed,
			ConnectionsCreated:  poolMetrics.ConnectionsCreated,
			ConnectionsClosed:   poolMetrics.ConnectionsClosed,
			LastHealthCheck:     poolMetrics.LastHealthCheck,
			HealthCheckFailures: poolMetrics.HealthCheckFailures,
		}
	}

	// Collect cache metrics
	if pm.cache != nil {
		pm.metrics.CacheMetrics = *pm.cache.GetMetrics()
	}

	// Collect query metrics
	if pm.query != nil {
		pm.metrics.QueryMetrics = *pm.query.GetMetrics()
	}

	// Collect from custom collectors
	for _, collector := range pm.collectors {
		if data, err := collector.Collect(pm.ctx); err != nil {
			pm.logger.Error("Failed to collect metrics",
				zap.String("collector", collector.GetName()),
				zap.Error(err))
		} else {
			pm.logger.Debug("Collected metrics",
				zap.String("collector", collector.GetName()),
				zap.Any("data", data))
		}
	}

	// Log performance summary
	pm.logPerformanceSummary()
}

// collectMemoryMetrics collects memory usage metrics
func (pm *PerformanceMonitor) collectMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pm.metrics.MemoryUsage = MemoryMetrics{
		AllocatedMB:   float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB:  float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:         float64(m.Sys) / 1024 / 1024,
		NumGC:         m.NumGC,
		GCPauseMS:     float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000,
		HeapObjects:   m.HeapObjects,
		StackInUseMB:  float64(m.StackInuse) / 1024 / 1024,
		StackSysMB:    float64(m.StackSys) / 1024 / 1024,
		MSpanInUseMB:  float64(m.MSpanInuse) / 1024 / 1024,
		MSpanSysMB:    float64(m.MSpanSys) / 1024 / 1024,
		MCacheInUseMB: float64(m.MCacheInuse) / 1024 / 1024,
		MCacheSysMB:   float64(m.MCacheSys) / 1024 / 1024,
		BuckHashSysMB: float64(m.BuckHashSys) / 1024 / 1024,
		GCSysMB:       float64(m.GCSys) / 1024 / 1024,
		OtherSysMB:    float64(m.OtherSys) / 1024 / 1024,
		NextGC:        m.NextGC,
		LastGC:        time.Unix(0, int64(m.LastGC)),
		PauseTotalNS:  m.PauseTotalNs,
		NumForcedGC:   m.NumForcedGC,
	}
}

// logPerformanceSummary logs a summary of current performance
func (pm *PerformanceMonitor) logPerformanceSummary() {
	pm.logger.Info("Performance summary",
		zap.Float64("memory_mb", pm.metrics.MemoryUsage.AllocatedMB),
		zap.Int("active_connections", pm.metrics.DatabaseMetrics.ActiveConnections),
		zap.Float64("cache_hit_rate", pm.metrics.CacheMetrics.HitRate),
		zap.Duration("avg_query_latency", pm.metrics.QueryMetrics.AverageLatency),
		zap.Float64("error_rate", pm.metrics.ErrorMetrics.ErrorRate))
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Overall string                 `json:"overall"`
	Checks  map[string]CheckStatus `json:"checks"`
}

// CheckStatus represents the status of a health check
type CheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MemoryCollector collects memory-related metrics
type MemoryCollector struct{}

func (mc *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m, nil
}

func (mc *MemoryCollector) GetName() string {
	return "memory"
}

// DatabaseCollector collects database-related metrics
type DatabaseCollector struct {
	db *sql.DB
}

func (dc *DatabaseCollector) Collect(ctx context.Context) (interface{}, error) {
	stats := dc.db.Stats()
	return stats, nil
}

func (dc *DatabaseCollector) GetName() string {
	return "database"
}

// RuntimeCollector collects runtime-related metrics
type RuntimeCollector struct{}

func (rc *RuntimeCollector) Collect(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"cpus":       runtime.NumCPU(),
		"version":    runtime.Version(),
	}, nil
}

func (rc *RuntimeCollector) GetName() string {
	return "runtime"
}
