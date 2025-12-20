# Priority 3: Website Scraping Timeouts - Analysis
## December 19, 2025

---

## Problem Statement

**Issue**: 29% of requests with website URLs are timing out (Target: <5%)

**Root Cause Identified**:
1. **Middleware Timeout**: Fixed at 30 seconds
2. **Adaptive Timeout**: Calculates 86 seconds for website scraping requests
3. **Mismatch**: Middleware times out at 30s before adaptive timeout (86s) can be used
4. **Handler Context Fix**: Handler creates fresh context if <90s remaining, but middleware already timed out

---

## Current Timeout Configuration

### 1. Middleware Timeout (`main.go:275`)
```go
router.Use(timeoutMiddleware(30 * time.Second)) // Fixed 30s timeout
```

### 2. Adaptive Timeout Calculation (`classification.go:5098`)
```go
// Website scraping requests: 86s total
requiredTimeout = indexBuildingBudget + phase1ScrapingBudget + multiPageAnalysisBudget + 
                  goClassificationBudget + mlClassificationBudget + generalOverhead + retryBuffer
// = 30 + 18 + 8 + 5 + 10 + 5 + 10 = 86s
```

### 3. Handler Context Fix (`classification.go:948`)
```go
// Creates fresh context if <90s remaining
if timeRemaining < 90*time.Second {
    parentCtx = context.Background()
}
```

### 4. Worker Pool Context (`classification.go:256`)
```go
freshTimeout := 120 * time.Second
processingCtx, cancel := context.WithTimeout(context.Background(), freshTimeout)
```

---

## Timeout Layers

| Layer | Timeout | Purpose | Issue |
|-------|---------|---------|-------|
| Middleware | 30s | Request-level timeout | ❌ Too short for website scraping |
| Handler Context | 90s+ | Handler processing | ✅ Creates fresh context if needed |
| Adaptive Timeout | 86s | Website scraping requests | ✅ Correctly calculated |
| Worker Context | 120s | Worker processing | ✅ Sufficient |

---

## Impact

**Current Behavior**:
- Requests with website URLs: Timeout at 30s (middleware)
- Requests without URLs: Complete successfully (<30s)

**Expected Behavior**:
- Requests with website URLs: Complete in 60-90s (adaptive timeout)
- Requests without URLs: Complete in <30s

---

## Solution Strategy

### Option 1: Increase Middleware Timeout (Recommended)
- **Pros**: Simple, maintains middleware protection
- **Cons**: All requests get longer timeout (even simple ones)
- **Implementation**: Increase to 90-120s

### Option 2: Dynamic Middleware Timeout
- **Pros**: Optimal timeout per request type
- **Cons**: More complex, requires request parsing in middleware
- **Implementation**: Parse request body in middleware, calculate adaptive timeout

### Option 3: Remove Middleware Timeout
- **Pros**: Simplest, relies on handler/worker timeouts
- **Cons**: No request-level timeout protection
- **Implementation**: Remove middleware timeout, rely on handler logic

---

## Recommended Fix

**Option 1 + Option 2 Hybrid**:
1. Increase middleware timeout to 120s (matches worker timeout)
2. Add timeout monitoring and logging
3. Optimize website scraping performance

**Rationale**:
- 120s matches worker pool timeout
- Handler already creates fresh context if <90s
- Adaptive timeout calculation is correct (86s)
- Provides safety margin for network delays

---

## Next Steps

1. ✅ **Analysis Complete** (this document)
2. ⏳ **Increase Middleware Timeout** to 120s
3. ⏳ **Add Timeout Monitoring** and logging
4. ⏳ **Optimize Website Scraping** performance
5. ⏳ **Test** with website URL requests

---

**Status**: Analysis complete, ready for implementation

