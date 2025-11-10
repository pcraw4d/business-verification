package performance

import (
	"time"

	"kyb-platform/services/risk-assessment-service/internal/cache"
	"kyb-platform/services/risk-assessment-service/internal/pool"
	"kyb-platform/services/risk-assessment-service/internal/query"
)

// CacheAdapter adapts cache.Cache to CacheMonitor interface
type CacheAdapter struct {
	cache cache.Cache
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(c cache.Cache) *CacheAdapter {
	return &CacheAdapter{cache: c}
}

// GetMetrics returns cache metrics converted to performance.CacheMetrics
func (a *CacheAdapter) GetMetrics() *CacheMetrics {
	if a.cache == nil {
		return &CacheMetrics{}
	}

	cacheMetrics := a.cache.GetMetrics()
	if cacheMetrics == nil {
		return &CacheMetrics{}
	}

	// Convert cache.CacheMetrics to performance.CacheMetrics
	return &CacheMetrics{
		Hits:           cacheMetrics.Hits,
		Misses:         cacheMetrics.Misses,
		Sets:           cacheMetrics.Sets,
		Deletes:        cacheMetrics.Deletes,
		Errors:         cacheMetrics.Errors,
		TotalRequests:  cacheMetrics.TotalRequests,
		HitRate:        cacheMetrics.HitRate,
		AverageLatency: cacheMetrics.AverageLatency,
		LastUpdated:    cacheMetrics.LastUpdated,
	}
}

// PoolAdapter adapts pool.ConnectionPool to PoolMonitor interface
type PoolAdapter struct {
	pool *pool.ConnectionPool
}

// NewPoolAdapter creates a new pool adapter
func NewPoolAdapter(p *pool.ConnectionPool) *PoolAdapter {
	return &PoolAdapter{pool: p}
}

// GetMetrics returns pool metrics converted to performance.PoolMetrics
func (a *PoolAdapter) GetMetrics() *PoolMetrics {
	if a.pool == nil {
		return &PoolMetrics{}
	}

	poolMetrics := a.pool.GetMetrics()
	if poolMetrics == nil {
		return &PoolMetrics{}
	}

	// Convert pool.PoolMetrics to performance.PoolMetrics
	return &PoolMetrics{
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

// QueryAdapter adapts query.QueryOptimizer to QueryMonitor interface
type QueryAdapter struct {
	optimizer *query.QueryOptimizer
}

// NewQueryAdapter creates a new query adapter
func NewQueryAdapter(q *query.QueryOptimizer) *QueryAdapter {
	return &QueryAdapter{optimizer: q}
}

// GetMetrics returns query metrics converted to performance.QueryMetrics
func (a *QueryAdapter) GetMetrics() *QueryMetrics {
	if a.optimizer == nil {
		return &QueryMetrics{}
	}

	// QueryOptimizer doesn't have GetMetrics method yet
	// Return empty metrics for now - this is acceptable as the performance monitor can still function
	// TODO: Add GetMetrics to QueryOptimizer if query metrics are needed
	return &QueryMetrics{
		LastUpdated: time.Now(),
	}
}

