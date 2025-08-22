# V2 Architecture Migration Guide

## Overview

This guide helps developers migrate from the legacy observability architecture to the new V2 architecture. The V2 architecture eliminates the adapter pattern and provides direct type usage for better performance and maintainability.

## What Changed

### Before (Legacy Architecture)
```go
// Complex adapter pattern with type conversions
legacyMetrics := performanceMonitor.GetLegacyAdapter().GetLegacyMetrics()
optimizer := NewAutomatedOptimizer(performanceMonitor, config, logger)
optimizer.ForceOptimization() // Used legacy adapter internally
```

### After (V2 Architecture)
```go
// Direct V2 type usage
metrics := performanceMonitor.GetMetricsV2()
optimizer := NewAutomatedOptimizer(performanceMonitor, config, logger)
optimizer.ForceOptimization() // Uses V2 types directly
```

## Migration Steps

### Step 1: Update Import Statements

**Remove adapter imports:**
```go
// OLD
import "github.com/pcraw4d/business-verification/internal/observability/adapters"

// NEW
import "github.com/pcraw4d/business-verification/internal/observability/types"
```

### Step 2: Update Type References

**Replace legacy types with V2 types:**
```go
// OLD
var metrics *PerformanceMetrics
var recommendations []*OptimizationRecommendation

// NEW
var metrics *types.PerformanceMetricsV2
var recommendations []*OptimizationRecommendation
```

### Step 3: Update Method Calls

**Replace legacy method calls:**
```go
// OLD
metrics := performanceMonitor.GetLegacyAdapter().GetLegacyMetrics()

// NEW
metrics := performanceMonitor.GetMetricsV2()
```

### Step 4: Update Field Access

**Update field access patterns:**
```go
// OLD
if metrics.AverageResponseTime > threshold {
    // Handle slow response time
}

// NEW
if metrics.Summary.P50Latency > threshold {
    // Handle slow response time
}
```

## Field Mapping Reference

### Performance Metrics Fields

| Legacy Field | V2 Field Path | Description |
|--------------|---------------|-------------|
| `AverageResponseTime` | `Summary.P50Latency` | Response time metric |
| `RequestsPerSecond` | `Summary.RPS` | Throughput metric |
| `SuccessRate` | `Summary.SuccessRate` | Success rate metric |
| `CPUUsage` | `Summary.CPUUsage` | CPU utilization |
| `MemoryUsage` | `Summary.MemoryUsage` | Memory utilization |
| `ErrorRate` | `Summary.ErrorRate` | Error rate metric |
| `P95ResponseTime` | `Summary.P95Latency` | 95th percentile latency |
| `P99ResponseTime` | `Summary.P99Latency` | 99th percentile latency |

### Detailed Metrics Fields

| Legacy Field | V2 Field Path | Description |
|--------------|---------------|-------------|
| `MinResponseTime` | `Breakdown.Latency.Min` | Minimum response time |
| `MaxResponseTime` | `Breakdown.Latency.Max` | Maximum response time |
| `PeakConcurrency` | `Breakdown.Throughput.Peak` | Peak concurrency |
| `ConcurrentRequests` | `Breakdown.Throughput.Concurrency` | Current concurrency |
| `TimeoutRate` | `Breakdown.Success.TimeoutRate` | Timeout rate |
| `DiskUsage` | `Breakdown.Resources.Disk` | Disk utilization |
| `NetworkIO` | `Breakdown.Resources.Network` | Network I/O |

## Component-Specific Migration

### Performance Monitor

**Before:**
```go
// Get legacy metrics
legacyMetrics := performanceMonitor.GetLegacyAdapter().GetLegacyMetrics()

// Check health
if legacyMetrics.SuccessRate < 0.95 {
    // Handle low success rate
}
```

**After:**
```go
// Get V2 metrics
metrics := performanceMonitor.GetMetricsV2()

// Check health
if metrics.Summary.SuccessRate < 0.95 {
    // Handle low success rate
}
```

### Automated Optimizer

**Before:**
```go
// Create optimizer with legacy adapter
optimizer := NewAutomatedOptimizer(
    performanceMonitor,
    config,
    logger,
)

// Force optimization (used legacy adapter internally)
optimizer.ForceOptimization()
```

**After:**
```go
// Create optimizer with V2 support
optimizer := NewAutomatedOptimizer(
    performanceMonitor,
    config,
    logger,
)

// Force optimization (uses V2 types directly)
optimizer.ForceOptimization()
```

### Performance Optimization System

**Before:**
```go
// Create system with legacy adapter
pos := NewPerformanceOptimizationSystem(
    performanceMonitor,
    regressionDetection,
    benchmarkingSystem,
    predictiveAnalytics,
    config,
    logger,
)

// Generate recommendations (used legacy adapter)
recommendations := pos.GenerateRecommendations()
```

**After:**
```go
// Create system with V2 support
pos := NewPerformanceOptimizationSystem(
    performanceMonitor,
    regressionDetection,
    benchmarkingSystem,
    predictiveAnalytics,
    config,
    logger,
)

// Generate recommendations (uses V2 types)
recommendations := pos.GenerateRecommendations()
```

### Automated Performance Tuning

**Before:**
```go
// Create tuning system with legacy adapter
apts := NewAutomatedPerformanceTuningSystem(
    performanceMonitor,
    optimizationSystem,
    predictiveAnalytics,
    regressionDetection,
    config,
    logger,
)

// Start tuning (used legacy adapter internally)
apts.Start(ctx)
```

**After:**
```go
// Create tuning system with V2 support
apts := NewAutomatedPerformanceTuningSystem(
    performanceMonitor,
    optimizationSystem,
    predictiveAnalytics,
    regressionDetection,
    config,
    logger,
)

// Start tuning (uses V2 types directly)
apts.Start(ctx)
```

## Testing Migration

### Update Test Data

**Before:**
```go
// Create test metrics with legacy structure
testMetrics := &PerformanceMetrics{
    AverageResponseTime: 200 * time.Millisecond,
    RequestsPerSecond:   150.0,
    SuccessRate:         0.98,
    CPUUsage:            0.60,
}
```

**After:**
```go
// Create test metrics with V2 structure
testMetrics := &types.PerformanceMetricsV2{
    Summary: types.MetricsSummary{
        P50Latency:  200 * time.Millisecond,
        RPS:         150.0,
        SuccessRate: 0.98,
        CPUUsage:    0.60,
    },
}
```

### Update Test Assertions

**Before:**
```go
// Test legacy field access
assert.Equal(t, 200*time.Millisecond, metrics.AverageResponseTime)
assert.Equal(t, 150.0, metrics.RequestsPerSecond)
```

**After:**
```go
// Test V2 field access
assert.Equal(t, 200*time.Millisecond, metrics.Summary.P50Latency)
assert.Equal(t, 150.0, metrics.Summary.RPS)
```

## Common Migration Patterns

### 1. Health Check Migration

**Before:**
```go
func checkSystemHealth(metrics *PerformanceMetrics) bool {
    return metrics.SuccessRate >= 0.95 &&
           metrics.AverageResponseTime <= 1*time.Second &&
           metrics.CPUUsage <= 0.8
}
```

**After:**
```go
func checkSystemHealth(metrics *types.PerformanceMetricsV2) bool {
    return metrics.Summary.SuccessRate >= 0.95 &&
           metrics.Summary.P50Latency <= 1*time.Second &&
           metrics.Summary.CPUUsage <= 0.8
}
```

### 2. Alert Condition Migration

**Before:**
```go
func shouldAlert(metrics *PerformanceMetrics) bool {
    return metrics.ErrorRate > 0.05 ||
           metrics.AverageResponseTime > 2*time.Second ||
           metrics.CPUUsage > 0.9
}
```

**After:**
```go
func shouldAlert(metrics *types.PerformanceMetricsV2) bool {
    return metrics.Summary.ErrorRate > 0.05 ||
           metrics.Summary.P50Latency > 2*time.Second ||
           metrics.Summary.CPUUsage > 0.9
}
```

### 3. Optimization Logic Migration

**Before:**
```go
func needsOptimization(metrics *PerformanceMetrics) bool {
    return metrics.AverageResponseTime > threshold ||
           metrics.SuccessRate < 0.95 ||
           metrics.CPUUsage > 0.8
}
```

**After:**
```go
func needsOptimization(metrics *types.PerformanceMetricsV2) bool {
    return metrics.Summary.P50Latency > threshold ||
           metrics.Summary.SuccessRate < 0.95 ||
           metrics.Summary.CPUUsage > 0.8
}
```

## Troubleshooting

### Common Issues

1. **Type Mismatch Errors**
   ```
   cannot use *PerformanceMetrics as *types.PerformanceMetricsV2
   ```
   **Solution**: Update type declarations to use V2 types

2. **Field Not Found Errors**
   ```
   unknown field AverageResponseTime in struct literal
   ```
   **Solution**: Use correct V2 field paths (e.g., `Summary.P50Latency`)

3. **Method Not Found Errors**
   ```
   undefined: GetLegacyAdapter
   ```
   **Solution**: Use `GetMetricsV2()` instead of `GetLegacyAdapter().GetLegacyMetrics()`

### Debugging Tips

1. **Check Type Declarations**
   ```go
   // Ensure you're using V2 types
   var metrics *types.PerformanceMetricsV2
   ```

2. **Verify Field Access**
   ```go
   // Use correct field paths
   fmt.Printf("Success Rate: %f\n", metrics.Summary.SuccessRate)
   fmt.Printf("Response Time: %v\n", metrics.Summary.P50Latency)
   ```

3. **Test Incrementally**
   ```go
   // Test one component at a time
   go build ./internal/observability/performance_monitor.go
   go build ./internal/observability/automated_optimizer.go
   ```

## Migration Checklist

- [ ] Update import statements (remove adapters, add types)
- [ ] Update type declarations to use V2 types
- [ ] Update method calls to use V2 methods
- [ ] Update field access to use V2 field paths
- [ ] Update test data and assertions
- [ ] Verify build success
- [ ] Run tests to ensure functionality
- [ ] Update documentation

## Benefits After Migration

1. **Better Performance**: No more type conversion overhead
2. **Cleaner Code**: Direct type usage without adapters
3. **Type Safety**: Compile-time type checking
4. **Maintainability**: Consistent patterns across codebase
5. **Future-Proof**: Ready for V2 enhancements

## Support

If you encounter issues during migration:

1. Check the [V2 Architecture Documentation](observability-v2-architecture.md)
2. Review the field mapping reference above
3. Use the troubleshooting section for common issues
4. Contact the development team for assistance

## Conclusion

The V2 architecture migration provides significant benefits in terms of performance, maintainability, and type safety. By following this guide, you can successfully migrate your code to use the new V2 types and eliminate the complexity of the adapter pattern.

Remember to migrate incrementally and test thoroughly to ensure a smooth transition.
