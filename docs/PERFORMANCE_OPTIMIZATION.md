# Performance Optimization Guide

## Overview

This document provides comprehensive guidance on achieving and maintaining sub-1-second API response times (95th percentile) for the Risk Assessment Service. The performance optimization system includes profiling, monitoring, caching, database optimization, and automated performance testing.

## Performance Targets

### Primary Targets
- **P95 Response Time**: < 1 second
- **P99 Response Time**: < 2 seconds
- **Average Response Time**: < 500ms
- **Throughput**: 1000 requests/minute
- **Error Rate**: < 1%
- **Availability**: 99.9%

### Secondary Targets
- **Memory Usage**: < 500MB
- **CPU Usage**: < 80%
- **Goroutine Count**: < 1000
- **Cache Hit Rate**: > 80%

## Architecture Components

### 1. Performance Profiler (`internal/performance/profiler.go`)

The profiler tracks operation performance with detailed metrics:

```go
// Start timing an operation
timer := profiler.StartTimer("operation_name")
defer timer()

// Record custom metrics
profiler.RecordMetric("custom_operation", duration)
```

**Key Features:**
- Automatic timing for operations
- Percentile calculations (P50, P95, P99)
- Memory usage tracking
- Goroutine monitoring
- Slow operation detection

### 2. Database Optimizer (`internal/performance/db_optimizer.go`)

Optimizes database performance with connection pooling and query monitoring:

```go
// Execute optimized query
rows, err := dbOptimizer.ExecuteQuery(ctx, query, args...)

// Get database statistics
stats := dbOptimizer.GetStats()
```

**Key Features:**
- Connection pool optimization
- Query timeout management
- Slow query detection
- Connection monitoring
- Automatic query optimization

### 3. Cache Optimizer (`internal/performance/cache_optimizer.go`)

Provides intelligent caching with TTL and LRU eviction:

```go
// Get from cache
value, err := cacheOptimizer.Get(ctx, "cache_name", "key")

// Set in cache
err := cacheOptimizer.Set(ctx, "cache_name", "key", value, ttl)
```

**Key Features:**
- In-memory caching with TTL
- LRU eviction policy
- Cache hit/miss tracking
- Automatic cleanup
- Multiple cache namespaces

### 4. Response Monitor (`internal/performance/response_monitor.go`)

Monitors API response times and generates alerts:

```go
// Record response time
responseMonitor.RecordResponse(endpoint, method, duration, success)

// Check system health
healthy := responseMonitor.IsHealthy()
```

**Key Features:**
- Real-time response time tracking
- Performance alerting
- Health score calculation
- Endpoint-specific metrics
- Automatic threshold monitoring

### 5. Performance Middleware (`internal/performance/middleware.go`)

HTTP middleware for automatic performance monitoring:

```go
// Apply performance middleware
router.Use(performanceMiddleware.Middleware())
```

**Key Features:**
- Automatic request timing
- Cache integration
- Response time logging
- Performance metrics collection
- Configurable skip paths

## Usage Examples

### 1. Basic Performance Monitoring

```go
// Initialize performance optimizer
config := performance.DefaultOptimizerConfig()
config.TargetP95 = 1 * time.Second
config.TargetP99 = 2 * time.Second

optimizer := performance.NewOptimizer(logger, db, config)

// Get performance statistics
stats := optimizer.GetPerformanceStats()

// Run optimization
report, err := optimizer.Optimize()
```

### 2. Performance Testing

```bash
# Run standard performance test
make performance-test

# Run quick test (5 users, 50 requests each, 2 minutes)
make performance-test-quick

# Run stress test (50 users, 200 requests each, 10 minutes)
make performance-test-stress

# Run spike test (100 users, 100 requests each, 5 minutes)
make performance-test-spike
```

### 3. Performance Optimization

```bash
# Get optimization report
make optimize-performance

# Check performance statistics
make performance-stats

# Check performance health
make performance-health

# Get full performance report
make performance-report
```

### 4. ML Model Validation

```bash
# Run standard ML validation
make validate-ml

# Run quick validation (3-fold, 100 samples, 30 days)
make validate-ml-quick

# Run comprehensive validation (10-fold, 5000 samples, 2 years)
make validate-ml-comprehensive
```

## API Endpoints

### Performance Optimization Endpoints

- `GET /api/v1/optimization/report` - Get optimization report
- `GET /api/v1/optimization/stats` - Get performance statistics
- `GET /api/v1/optimization/health` - Check performance health
- `GET /api/v1/optimization/report/full` - Get full performance report

### Legacy Performance Endpoints

- `GET /api/v1/performance/stats` - Get legacy performance stats
- `GET /api/v1/performance/alerts` - Get performance alerts
- `GET /api/v1/performance/health` - Check legacy performance health
- `POST /api/v1/performance/reset` - Reset performance metrics
- `POST /api/v1/performance/targets` - Update performance targets
- `POST /api/v1/performance/alerts/clear` - Clear performance alerts

## Configuration

### Performance Optimizer Configuration

```go
type OptimizerConfig struct {
    EnableProfiling        bool          `json:"enable_profiling"`
    EnableDBOptimization   bool          `json:"enable_db_optimization"`
    EnableCaching          bool          `json:"enable_caching"`
    EnableResponseMonitoring bool        `json:"enable_response_monitoring"`
    PerformanceThreshold   time.Duration `json:"performance_threshold"`
    OptimizationInterval   time.Duration `json:"optimization_interval"`
    EnableAutoOptimization bool          `json:"enable_auto_optimization"`
    TargetP95              time.Duration `json:"target_p95"`
    TargetP99              time.Duration `json:"target_p99"`
    TargetThroughput       int           `json:"target_throughput"`
}
```

### Database Configuration

```go
type DBConfig struct {
    MaxOpenConns        int           `json:"max_open_conns"`
    MaxIdleConns        int           `json:"max_idle_conns"`
    ConnMaxLifetime     time.Duration `json:"conn_max_lifetime"`
    ConnMaxIdleTime     time.Duration `json:"conn_max_idle_time"`
    QueryTimeout        time.Duration `json:"query_timeout"`
    SlowQueryThreshold  time.Duration `json:"slow_query_threshold"`
    EnableQueryLogging  bool          `json:"enable_query_logging"`
    EnableSlowQueryLog  bool          `json:"enable_slow_query_log"`
    EnableConnectionLog bool          `json:"enable_connection_log"`
}
```

### Cache Configuration

```go
type CacheConfig struct {
    DefaultTTL        time.Duration `json:"default_ttl"`
    MaxSize           int           `json:"max_size"`
    CleanupInterval   time.Duration `json:"cleanup_interval"`
    EnableStats       bool          `json:"enable_stats"`
    EnableProfiling   bool          `json:"enable_profiling"`
    LRUEnabled        bool          `json:"lru_enabled"`
    CompressionEnabled bool         `json:"compression_enabled"`
}
```

## Performance Optimization Strategies

### 1. Database Optimization

- **Connection Pooling**: Configure optimal connection pool settings
- **Query Optimization**: Use prepared statements and proper indexing
- **Query Timeout**: Set appropriate timeouts to prevent hanging queries
- **Slow Query Monitoring**: Track and optimize slow queries
- **Connection Monitoring**: Monitor connection pool health

### 2. Caching Strategy

- **Response Caching**: Cache API responses for frequently accessed data
- **Query Result Caching**: Cache database query results
- **Session Caching**: Cache user sessions and authentication data
- **Static Content Caching**: Cache static assets and configuration
- **Cache Warming**: Pre-populate cache with frequently accessed data

### 3. Code Optimization

- **Profiling**: Use the built-in profiler to identify bottlenecks
- **Memory Management**: Optimize memory allocation and garbage collection
- **Concurrency**: Use goroutines and channels effectively
- **Algorithm Optimization**: Choose efficient algorithms and data structures
- **Resource Pooling**: Reuse expensive resources like connections

### 4. Infrastructure Optimization

- **Load Balancing**: Distribute load across multiple instances
- **Auto-scaling**: Scale resources based on demand
- **CDN**: Use Content Delivery Network for static content
- **Compression**: Enable gzip compression for responses
- **Keep-Alive**: Use HTTP keep-alive connections

## Monitoring and Alerting

### Performance Metrics

- **Response Time Percentiles**: P50, P95, P99
- **Throughput**: Requests per second
- **Error Rate**: Percentage of failed requests
- **Resource Usage**: CPU, memory, disk, network
- **Cache Performance**: Hit rate, miss rate, eviction rate
- **Database Performance**: Query time, connection pool status

### Alerting Rules

- **High Response Time**: P95 > 1 second
- **Very High Response Time**: P99 > 2 seconds
- **High Error Rate**: Error rate > 1%
- **Low Throughput**: Throughput < 1000 req/min
- **High Memory Usage**: Memory > 500MB
- **High CPU Usage**: CPU > 80%
- **Database Issues**: Connection pool exhaustion or slow queries

### Health Checks

- **Liveness Probe**: Basic service availability
- **Readiness Probe**: Service ready to accept requests
- **Performance Health**: Overall performance score
- **Resource Health**: Resource usage within limits
- **Dependency Health**: External service availability

## Troubleshooting

### Common Performance Issues

1. **Slow Response Times**
   - Check database query performance
   - Review cache hit rates
   - Analyze profiler data
   - Check for memory leaks

2. **High Memory Usage**
   - Review memory allocation patterns
   - Check for goroutine leaks
   - Optimize data structures
   - Increase garbage collection frequency

3. **Database Performance Issues**
   - Review slow query logs
   - Check connection pool status
   - Optimize database indexes
   - Review query patterns

4. **Cache Performance Issues**
   - Check cache hit rates
   - Review TTL settings
   - Monitor cache evictions
   - Optimize cache key generation

### Performance Debugging

```bash
# Get detailed performance report
curl -X GET http://localhost:8080/api/v1/optimization/report/full

# Check specific performance metrics
curl -X GET http://localhost:8080/api/v1/optimization/stats | jq .

# Monitor performance in real-time
watch -n 5 'curl -s http://localhost:8080/api/v1/optimization/health | jq .'

# Run performance test
make performance-test-quick
```

## Best Practices

### 1. Development

- **Profile Early**: Use profiling during development
- **Test Performance**: Include performance tests in CI/CD
- **Monitor Continuously**: Set up continuous monitoring
- **Optimize Incrementally**: Make small, measurable improvements
- **Document Changes**: Document performance-related changes

### 2. Production

- **Set Alerts**: Configure appropriate alerting thresholds
- **Monitor Trends**: Track performance trends over time
- **Plan Capacity**: Plan for traffic growth
- **Regular Optimization**: Run optimization regularly
- **Backup Plans**: Have fallback strategies for performance issues

### 3. Maintenance

- **Regular Reviews**: Review performance metrics regularly
- **Update Targets**: Adjust performance targets as needed
- **Clean Up**: Remove unused code and resources
- **Update Dependencies**: Keep dependencies up to date
- **Document Lessons**: Document performance lessons learned

## Performance Testing

### Test Types

1. **Load Testing**: Normal expected load
2. **Stress Testing**: Beyond normal capacity
3. **Spike Testing**: Sudden traffic increases
4. **Volume Testing**: Large amounts of data
5. **Endurance Testing**: Extended periods of load

### Test Scenarios

```bash
# Standard load test
make performance-test

# Stress test with high concurrency
make performance-test-stress

# Spike test with sudden load increase
make performance-test-spike

# Quick validation test
make performance-test-quick
```

### Test Validation

- **Response Time**: All percentiles within targets
- **Throughput**: Meets or exceeds target throughput
- **Error Rate**: Below 1% error rate
- **Resource Usage**: Within acceptable limits
- **Stability**: No memory leaks or crashes

## Conclusion

The performance optimization system provides comprehensive tools for achieving and maintaining sub-1-second API response times. By following the guidelines in this document and using the provided tools, you can ensure optimal performance for the Risk Assessment Service.

For additional support or questions, refer to the API documentation or contact the development team.
