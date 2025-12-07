# Root Cause Analysis: Timeout and Context Propagation Issues

**Date**: 2025-12-05  
**Status**: Investigation Complete - Root Causes Identified

---

## Executive Summary

The classification service is experiencing widespread timeout failures (88.63% failure rate) with "context deadline exceeded" errors. Investigation has identified **5 root causes** that need to be addressed systematically.

---

## Context Flow Analysis

### Current Call Chain

```
HandleClassification (creates 60s context)
  ↓
processClassification (receives 60s context)
  ↓
generateEnhancedClassification (receives 60s context)
  ↓
ClassifyBusiness(ctx, ...) (receives 60s context)
  ↓
extractKeywords(businessName, websiteURL) ❌ NO CONTEXT PARAMETER!
  ↓ (creates NEW 20s context from Background)
extractKeywordsFromWebsite(ctx, websiteURL) (receives 20s context)
  ↓
Phase 1 Scraper (needs 15s, but context may have <15s remaining)
```

### Problem 1: extractKeywords Doesn't Receive Context

**Location**: `internal/classification/repository/supabase_repository.go:2624`

**Issue**: 
- `extractKeywords()` function signature: `func extractKeywords(businessName, websiteURL string) []ContextualKeyword`
- **No context parameter**, so it cannot respect the parent context timeout
- Creates its own 20s context from `context.Background()`, making it independent of the request lifecycle

**Impact**:
- Parent context (60s) may expire while `extractKeywords` is running with its own 20s context
- No coordination between parent and child contexts
- If parent context expires, the entire request is cancelled, even if `extractKeywords` still has time

**Evidence from Logs**:
```
"Parent context deadline: 19.999855665s from now"
"Phase 1 context deadline: 19.999684071s from now"
```

---

## Root Cause #1: Context Not Propagated to extractKeywords

**Severity**: 🔴 **CRITICAL**

**Description**: 
The `extractKeywords()` function doesn't receive a context parameter, so it creates its own independent context. This breaks context propagation and prevents proper timeout management.

**Code Location**: 
- `internal/classification/repository/supabase_repository.go:2624`
- `internal/classification/repository/supabase_repository.go:1853` (caller)

**Fix Required**:
1. Add `ctx context.Context` parameter to `extractKeywords()`
2. Pass context from `ClassifyBusiness()` to `extractKeywords()`
3. Use parent context instead of creating new Background context
4. Only create new context if parent has insufficient time

---

## Root Cause #2: Adaptive Timeout Calculation Not Used

**Severity**: 🟠 **HIGH**

**Description**:
`calculateAdaptiveTimeout()` calculates 35s for scraping requests, but the actual timeout used is 60s (from `OverallTimeout`). The adaptive timeout is calculated but then ignored because `OverallTimeout` (60s) > `requiredTimeout` (35s).

**Code Location**:
- `services/classification-service/internal/handlers/classification.go:3044-3055`

**Evidence from Logs**:
```
"Adaptive timeout: website scraping detected", "calculated_timeout": 35
"Created context with timeout", "request_timeout": 60
```

**Issue**:
```go
// Line 3044-3055
if requiredTimeout > baseTimeout {
    return requiredTimeout
}
return baseTimeout  // Returns 60s instead of 35s!
```

**Impact**:
- Context has 60s, but by the time it reaches Phase 1, only ~20s remains
- Time is consumed by other operations (Go classification, database queries, etc.)
- Phase 1 gets insufficient time even though adaptive timeout calculated 35s

**Fix Required**:
- Use `requiredTimeout` (35s) instead of `baseTimeout` (60s) when it's calculated
- Or ensure `baseTimeout` is set to the calculated value before creating context

---

## Root Cause #3: HTTP Client Timeout > Context Deadline

**Severity**: 🟠 **HIGH**

**Description**:
HTTP clients in the scraper have 30s timeouts, but when Phase 1 starts, the context deadline is only ~20s. While `http.Client.Do()` should respect context cancellation, the client's own timeout (30s) can interfere with proper context handling.

**Code Locations**:
- `internal/external/website_scraper.go:40` - Default config: 30s
- `internal/external/website_scraper.go:127` - Playwright client: 30s
- `internal/external/website_scraper.go:94-97` - Main client: config.Timeout (30s)

**Evidence from Logs**:
```
"HTTP client timeout configuration", "client_timeout": 20, "context_deadline": 19.97813716
```

**Issue**:
- HTTP client timeout should be ≤ context deadline to ensure context cancellation works properly
- When client timeout (30s) > context deadline (20s), the client may wait for its own timeout instead of respecting context

**Fix Required**:
- Dynamically set HTTP client timeout based on context deadline
- Ensure `client.Timeout ≤ context deadline` when making requests
- Or remove client timeout and rely solely on context cancellation

---

## Root Cause #4: Time Consumed Before Phase 1

**Severity**: 🟡 **MEDIUM**

**Description**:
By the time the context reaches Phase 1 scraper, only ~20s remains out of the original 60s. This suggests ~40s is being consumed by operations before Phase 1:
- Database queries (ClassifyBusiness, keyword lookups)
- Go classification operations
- Context propagation overhead
- Other processing

**Evidence from Logs**:
```
"Created context with timeout", "time_remaining": 59.999977903
...
"Parent context deadline: 19.999855665s from now" (when Phase 1 starts)
```

**Time Budget Analysis**:
- Total context: 60s
- Time remaining when Phase 1 starts: ~20s
- Time consumed before Phase 1: ~40s
- Phase 1 needs: 15s
- Available: ~20s (should work, but timing out)

**Possible Causes**:
1. Database queries taking longer than expected
2. Go classification operations consuming time
3. Multiple sequential operations instead of parallel
4. Context creation/checking overhead

**Fix Required**:
- Profile operations before Phase 1 to identify time consumers
- Optimize database queries
- Make operations parallel where possible
- Reduce time allocated to non-critical operations

---

## Root Cause #5: Playwright Service Timeout Mismatch

**Severity**: 🟡 **MEDIUM**

**Description**:
Playwright service client has a 30s timeout, but the context deadline is only ~20s when it reaches Playwright. The Playwright request times out after ~20s (context deadline), not 30s.

**Code Location**:
- `internal/external/website_scraper.go:127` - Playwright client: 30s timeout

**Evidence from Logs**:
```
"Strategy returned", "strategy": "playwright", "error": "context deadline exceeded", "duration": 19.490941827
```

**Issue**:
- Playwright client timeout (30s) > context deadline (20s)
- Request respects context and times out at 20s
- Playwright service may need more than 20s to complete

**Fix Required**:
- Ensure Playwright client timeout ≤ context deadline
- Or increase context timeout to accommodate Playwright (but this conflicts with Root Cause #2)

---

## Additional Findings

### extractKeywords Timeout Too Short

**Location**: `internal/classification/repository/supabase_repository.go:2660`

**Current**: 20s timeout (recently increased from 5s)  
**Issue**: Still may be insufficient if Phase 1 needs 15s + overhead

**Recommendation**: Increase to 25s or make it dynamic based on parent context

---

### Sequential Multi-Level Extraction

**Location**: `internal/classification/repository/supabase_repository.go:2655-2800+`

**Current**: Level 2 → Level 1 → Level 3 → Level 4 (sequential)  
**Status**: ✅ Already optimized (Level 2 first, early termination added)

---

## Recommended Fix Priority

1. **🔴 CRITICAL**: Fix Root Cause #1 - Add context parameter to `extractKeywords()`
2. **🟠 HIGH**: Fix Root Cause #2 - Use adaptive timeout calculation result
3. **🟠 HIGH**: Fix Root Cause #3 - Make HTTP client timeout respect context deadline
4. **🟡 MEDIUM**: Fix Root Cause #4 - Profile and optimize pre-Phase 1 operations
5. **🟡 MEDIUM**: Fix Root Cause #5 - Align Playwright timeout with context

---

## Testing Strategy

After fixes are applied:

1. **Unit Tests**: Verify context propagation through call chain
2. **Integration Tests**: Test with various timeout scenarios
3. **Performance Tests**: Measure time consumption at each stage
4. **Comprehensive Tests**: Run full test suite to verify success rate improvement

---

## Expected Impact

**Current State**:
- Success Rate: 11.36% (5/44)
- Most requests timeout with "context deadline exceeded"

**After Fixes**:
- Success Rate: Target ≥95%
- Proper context propagation ensures timeouts are respected
- Adaptive timeout ensures sufficient time for operations
- HTTP clients respect context deadlines

---

## Next Steps

1. Implement fixes in priority order
2. Add comprehensive logging for context state at each stage
3. Profile operations to identify time consumers
4. Test fixes incrementally
5. Run comprehensive test suite to validate improvements

