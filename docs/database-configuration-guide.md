# Database Configuration Guide

## Overview

This guide provides comprehensive documentation for the optimized PostgreSQL database configuration implemented for the KYB Platform's classification and risk assessment workloads.

## Table of Contents

1. [Configuration Overview](#configuration-overview)
2. [Memory Optimization](#memory-optimization)
3. [Connection Pool Management](#connection-pool-management)
4. [Query Optimization](#query-optimization)
5. [Performance Monitoring](#performance-monitoring)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)
8. [Maintenance](#maintenance)

## Configuration Overview

### Performance Targets

- **Query Performance**: 50% faster database query performance
- **API Response Times**: <200ms average response time
- **Cost Optimization**: 30% reduction in database costs
- **System Uptime**: 99.9% availability target
- **Classification Accuracy**: 95%+ accuracy
- **Risk Detection**: 90%+ accuracy

### Key Optimizations

1. **Memory Management**: Optimized shared buffers, work memory, and cache settings
2. **Connection Pooling**: Intelligent connection management for concurrent operations
3. **Query Optimization**: SSD-optimized settings and parallel query execution
4. **Index Strategy**: Optimized indexes for classification and risk assessment queries
5. **Monitoring**: Comprehensive performance monitoring and alerting

## Memory Optimization

### Core Memory Settings

#### Shared Buffers
```sql
-- 25% of system RAM for shared buffers
ALTER SYSTEM SET shared_buffers = '256MB';
```

**Purpose**: Caches frequently accessed data pages in memory
**Optimization**: Set to 25% of available RAM for read-heavy workloads
**Impact**: Reduces disk I/O for frequently accessed data

#### Effective Cache Size
```sql
-- 75% of system RAM for effective cache size
ALTER SYSTEM SET effective_cache_size = '1GB';
```

**Purpose**: Helps query planner make better decisions about index vs sequential scans
**Optimization**: Set to 75% of total system RAM
**Impact**: Improves query planning accuracy

#### Work Memory
```sql
-- Memory for sorting, hash joins, and other operations
ALTER SYSTEM SET work_mem = '16MB';
```

**Purpose**: Memory allocated for sorting, hash joins, and other operations
**Optimization**: Increased for complex classification and risk assessment queries
**Impact**: Reduces temporary file usage and improves query performance

#### Maintenance Work Memory
```sql
-- Memory for maintenance operations
ALTER SYSTEM SET maintenance_work_mem = '256MB';
```

**Purpose**: Memory for maintenance operations like VACUUM, CREATE INDEX
**Optimization**: Increased for better performance during maintenance
**Impact**: Faster index creation and maintenance operations

### Memory Configuration by Workload

#### Classification Workload
```sql
-- Optimized for classification queries
ALTER SYSTEM SET work_mem = '32MB';
ALTER SYSTEM SET default_statistics_target = 200;
```

#### Risk Assessment Workload
```sql
-- Optimized for JSONB operations
ALTER SYSTEM SET work_mem = '24MB';
ALTER SYSTEM SET default_statistics_target = 150;
```

#### Mixed Workload
```sql
-- Balanced settings for mixed workload
ALTER SYSTEM SET work_mem = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

## Connection Pool Management

### Optimized Connection Pool Settings

```go
type OptimizedPoolConfig struct {
    MaxOpenConns    int           // 200 connections
    MaxIdleConns    int           // 40 connections (20% of max)
    ConnMaxLifetime time.Duration // 9 minutes
    ConnMaxIdleTime time.Duration // 3 minutes
}
```

### Connection Pool Calculation

#### Maximum Open Connections
```go
// Base calculation: 2 * CPU cores for I/O bound operations
baseConns := numCPU * 2

// Add connections for concurrent operations:
// - Classification processing: +10
// - Risk assessment: +10
// - API requests: +20
// - Background tasks: +10
additionalConns := 50

totalConns := baseConns + additionalConns
```

#### Maximum Idle Connections
```go
// Keep 20% of max connections as idle for quick response
idleConns := maxOpenConns / 5
```

### Connection Pool Monitoring

```go
// Monitor connection pool performance
func (cpo *ConnectionPoolOptimizer) MonitorConnectionPool(ctx context.Context, db *sql.DB, interval time.Duration) {
    // Log performance metrics every 30 seconds
    // Check for performance issues
    // Alert on high utilization
}
```

## Query Optimization

### SSD Optimization Settings

```sql
-- Optimize for SSD storage
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
```

**Purpose**: Optimize query planning for SSD storage
**Impact**: Encourages index usage and improves query performance

### Parallel Query Execution

```sql
-- Enable parallel query execution
ALTER SYSTEM SET max_parallel_workers_per_gather = 4;
ALTER SYSTEM SET max_parallel_workers = 8;
ALTER SYSTEM SET parallel_tuple_cost = 0.1;
ALTER SYSTEM SET parallel_setup_cost = 1000.0;
```

**Purpose**: Enable parallel execution for large queries
**Impact**: Faster execution of complex classification and risk assessment queries

### WAL and Checkpoint Optimization

```sql
-- Optimize WAL and checkpoint settings
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET max_wal_size = '1GB';
ALTER SYSTEM SET min_wal_size = '256MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.7;
```

**Purpose**: Optimize write performance and checkpoint behavior
**Impact**: Better write performance and reduced I/O spikes

## Performance Monitoring

### Key Performance Metrics

#### Connection Pool Metrics
- **Open Connections**: Current number of open connections
- **Idle Connections**: Number of idle connections
- **In-Use Connections**: Number of active connections
- **Wait Count**: Number of connection wait events
- **Wait Duration**: Total time spent waiting for connections

#### Memory Usage Metrics
- **Shared Buffer Hit Ratio**: Percentage of buffer cache hits
- **Work Memory Usage**: Current work memory utilization
- **Maintenance Memory Usage**: Current maintenance memory utilization

#### Query Performance Metrics
- **Average Query Time**: Average execution time for queries
- **Slow Query Count**: Number of queries exceeding threshold
- **Index Usage**: Percentage of queries using indexes

### Monitoring Views

#### Classification Performance Monitor
```sql
CREATE OR REPLACE VIEW classification_performance_monitor AS
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation,
    most_common_vals,
    most_common_freqs
FROM pg_stats 
WHERE schemaname = 'public' 
AND tablename IN (
    'merchants', 'business_classifications', 'risk_assessments', 
    'risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments'
);
```

#### Connection Monitor
```sql
CREATE OR REPLACE VIEW connection_monitor AS
SELECT 
    state,
    COUNT(*) as connection_count,
    AVG(EXTRACT(EPOCH FROM (now() - state_change))) as avg_duration_seconds
FROM pg_stat_activity 
WHERE state IS NOT NULL
GROUP BY state;
```

#### Query Performance Monitor
```sql
CREATE OR REPLACE VIEW query_performance_monitor AS
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query LIKE '%merchants%' 
   OR query LIKE '%classification%' 
   OR query LIKE '%risk%'
ORDER BY mean_time DESC;
```

### Performance Testing

#### Configuration Validation
```sql
-- Validate configuration settings
SELECT * FROM validate_postgresql_configuration();
```

#### Performance Tests
```sql
-- Run performance tests
SELECT * FROM run_performance_tests();
```

## Best Practices

### Configuration Management

1. **Environment-Specific Settings**: Adjust settings based on environment (dev/staging/prod)
2. **Gradual Changes**: Apply configuration changes gradually and monitor impact
3. **Backup Before Changes**: Always backup configuration before making changes
4. **Documentation**: Document all configuration changes and their rationale

### Monitoring Best Practices

1. **Regular Monitoring**: Monitor key metrics continuously
2. **Alert Thresholds**: Set appropriate alert thresholds for key metrics
3. **Performance Baselines**: Establish performance baselines for comparison
4. **Trend Analysis**: Analyze performance trends over time

### Maintenance Best Practices

1. **Regular VACUUM**: Schedule regular VACUUM operations
2. **Index Maintenance**: Monitor and maintain indexes regularly
3. **Statistics Updates**: Keep table statistics up to date
4. **Configuration Reviews**: Review configuration settings periodically

## Troubleshooting

### Common Issues

#### High Connection Utilization
**Symptoms**: High wait counts, slow response times
**Solutions**:
- Increase max_open_conns
- Optimize connection usage in application
- Check for connection leaks

#### Low Buffer Hit Ratio
**Symptoms**: High disk I/O, slow queries
**Solutions**:
- Increase shared_buffers
- Optimize queries to use indexes
- Check for missing indexes

#### Slow Query Performance
**Symptoms**: Queries taking longer than expected
**Solutions**:
- Check query execution plans
- Optimize indexes
- Increase work_mem for complex queries
- Enable parallel query execution

#### Memory Pressure
**Symptoms**: Out of memory errors, system slowdown
**Solutions**:
- Reduce work_mem
- Increase shared_buffers
- Optimize query complexity
- Check for memory leaks

### Diagnostic Queries

#### Check Current Configuration
```sql
SELECT name, setting, unit, context 
FROM pg_settings 
WHERE name IN (
    'shared_buffers', 'work_mem', 'maintenance_work_mem', 
    'random_page_cost', 'effective_io_concurrency',
    'max_connections', 'statement_timeout'
)
ORDER BY name;
```

#### Check Connection Pool Status
```sql
SELECT 
    state,
    COUNT(*) as connection_count,
    AVG(EXTRACT(EPOCH FROM (now() - state_change))) as avg_duration_seconds
FROM pg_stat_activity 
WHERE state IS NOT NULL
GROUP BY state;
```

#### Check Memory Usage
```sql
SELECT 
    SUM(blks_hit) as shared_buffers_hit,
    SUM(blks_read) as shared_buffers_read,
    100.0 * SUM(blks_hit) / nullif(SUM(blks_hit) + SUM(blks_read), 0) AS hit_percent
FROM pg_stat_database;
```

## Maintenance

### Regular Maintenance Tasks

#### Daily
- Monitor connection pool metrics
- Check for slow queries
- Review error logs

#### Weekly
- Analyze query performance trends
- Review index usage
- Check for configuration drift

#### Monthly
- Review and update configuration settings
- Analyze performance baselines
- Plan capacity upgrades

### Configuration Updates

#### Applying Configuration Changes
1. **Test in Development**: Test changes in development environment first
2. **Staged Rollout**: Apply changes in staging environment
3. **Monitor Impact**: Monitor performance impact after changes
4. **Rollback Plan**: Have rollback plan ready

#### Configuration Validation
```bash
# Run configuration validation
go run scripts/database/test-configuration-changes.go
```

### Performance Optimization

#### Continuous Optimization
1. **Monitor Trends**: Monitor performance trends over time
2. **Identify Bottlenecks**: Identify and address performance bottlenecks
3. **Optimize Queries**: Continuously optimize slow queries
4. **Update Indexes**: Add or modify indexes as needed

#### Capacity Planning
1. **Growth Projections**: Plan for expected growth
2. **Resource Scaling**: Scale resources as needed
3. **Performance Targets**: Maintain performance targets
4. **Cost Optimization**: Optimize costs while maintaining performance

## Conclusion

This database configuration guide provides comprehensive optimization for the KYB Platform's classification and risk assessment workloads. The configuration is designed to achieve:

- **50% faster query performance**
- **<200ms API response times**
- **30% cost reduction**
- **99.9% system uptime**

Regular monitoring and maintenance are essential to maintain optimal performance. Use the provided monitoring tools and diagnostic queries to ensure the configuration continues to meet performance targets.

For questions or issues, refer to the troubleshooting section or contact the development team.

---

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Next Review**: Monthly during implementation
