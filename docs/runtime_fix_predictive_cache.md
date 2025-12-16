# Runtime Fix: Predictive Cache Goroutine Issue

## Problem

The service was crashing with goroutine-related errors. The stack trace showed:

```
kyb-platform/internal/classification/cache.(*PredictiveCache).PreloadCache
```

The issue was in the predictive cache preloading functionality that spawns background goroutines.

## Root Causes

1. **No Panic Recovery**: Goroutines could panic and crash the entire service
2. **Unlimited Concurrency**: No limit on concurrent preload operations, causing resource exhaustion
3. **No Timeout**: Preload operations could hang indefinitely
4. **Context Issues**: Using `context.Background()` without timeout could lead to goroutine leaks

## Fixes Applied

### 1. Added Panic Recovery

**File:** `internal/classification/cache/predictive_cache.go`

Added panic recovery in the `PreloadCache` goroutine to prevent service crashes:

```go
go func() {
    // Recover from any panics to prevent service crash
    defer func() {
        if r := recover(); r != nil {
            pc.logger.Printf("⚠️ Panic in PreloadCache recovered: %v", r)
        }
    }()
    // ... rest of preload logic
}()
```

### 2. Added Concurrency Limiting

**File:** `internal/classification/cache/predictive_cache.go`

Added a semaphore to limit concurrent preload operations to 3:

```go
type PredictiveCache struct {
    // ... existing fields
    preloadSem chan struct{} // Semaphore to limit concurrent preload operations
}

// In NewPredictiveCache:
preloadSem: make(chan struct{}, 3), // Limit to 3 concurrent preload operations
```

### 3. Added Timeout Context

**File:** `internal/classification/cache/predictive_cache.go`

Added a 10-second timeout for preload operations:

```go
// Create a timeout context for preload operations (10 seconds max)
preloadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

### 4. Added Semaphore Acquisition

**File:** `internal/classification/cache/predictive_cache.go`

Each preload operation now acquires the semaphore before proceeding:

```go
// Acquire semaphore to limit concurrent operations
select {
case pc.preloadSem <- struct{}{}:
    // Got semaphore, proceed with preload
    func() {
        defer func() { <-pc.preloadSem }() // Release semaphore
        // ... preload logic
    }()
case <-preloadCtx.Done():
    // Context cancelled, stop preloading
    return
default:
    // Semaphore full, skip this variation (best effort)
    pc.logger.Printf("⚠️ Preload semaphore full, skipping: %s", variation)
}
```

### 5. Enhanced Error Handling

**File:** `internal/classification/cache/predictive_cache.go`

Preload operations now handle errors gracefully without failing:

```go
result, err := pc.classifyAndCache(preloadCtx, variation, description, websiteURL)
if err != nil {
    // Log error but don't fail - preload is best effort
    pc.logger.Printf("⚠️ Pre-cache failed for %s: %v", variation, err)
}
```

### 6. Added Panic Recovery in Caller

**File:** `internal/classification/multi_strategy_classifier.go`

Added additional panic recovery when calling PreloadCache:

```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            msc.logger.Printf("⚠️ Panic in predictive preload recovered: %v", r)
        }
    }()
    // Use background context with timeout to prevent goroutine leaks
    preloadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    msc.predictiveCache.PreloadCache(preloadCtx, businessName, description, websiteURL)
}()
```

## Benefits

1. **Service Stability**: Panic recovery prevents service crashes
2. **Resource Management**: Semaphore limits prevent resource exhaustion
3. **Timeout Protection**: Operations can't hang indefinitely
4. **Graceful Degradation**: Preload failures don't affect main requests
5. **Best Effort**: Preload is non-blocking and best-effort, so failures are acceptable

## Testing

After applying these fixes:

- ✅ Service starts successfully
- ✅ Service handles multiple requests without crashing
- ✅ Service remains healthy after load
- ✅ No panic errors in logs

## Files Modified

- `internal/classification/cache/predictive_cache.go`
- `internal/classification/multi_strategy_classifier.go`
