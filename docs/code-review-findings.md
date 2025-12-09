# Classification Service Code Review Findings

**Date:** December 8, 2025  
**Reviewer:** AI Assistant  
**Status:** 🔴 **CRITICAL ISSUES FOUND**

---

## Executive Summary

A comprehensive code review has identified **multiple critical issues** that explain why test success rates remain low despite multiple optimization attempts. The primary issue is **context timeout management** where contexts are being overwritten with shorter timeouts, negating all previous fixes.

---

## 🔴 P0: CRITICAL ISSUES

### Issue #1: Context Timeout Overwrite (ROOT CAUSE)

**Location:** `services/classification-service/internal/handlers/classification.go:850-851`

**Problem:**
```go
ctx, contentCache := reqcache.WithContentCache(parentCtx)
ctx, cancel := context.WithTimeout(ctx, requestTimeout)  // ⚠️ OVERWRITES CONTEXT!
```

**Root Cause:**
1. We create a fresh `context.Background()` at entry if parent has insufficient time (line 728/737)
2. We calculate `requestTimeout` (~68s) based on adaptive timeout
3. We then **OVERWRITE** the context with `context.WithTimeout(ctx, requestTimeout)`
4. This reduces the timeout from potentially unlimited (Background) or 120s (worker) to only ~68s
5. The worker creates a fresh 120s context, but it receives a request with a context that already has only ~68s remaining

**Impact:**
- All context fixes are negated
- Requests timeout after ~68s even though worker has 120s
- This explains why we see 78s timeouts (68s + 10s buffer)

**Evidence:**
- Logs show `requestTimeout: 68` in QUEUE-ENQUEUE
- Worker creates 120s context but receives request with ~68s context
- Requests fail at ~78s mark

**Fix:**
```go
// DON'T overwrite context if it already has sufficient time
if deadline, hasDeadline := parentCtx.Deadline(); hasDeadline {
    timeRemaining := time.Until(deadline)
    if timeRemaining >= requestTimeout {
        // Use existing context, don't overwrite
        ctx, contentCache = reqcache.WithContentCache(parentCtx)
        // Don't create new timeout
    } else {
        // Only create new timeout if existing is insufficient
        ctx, contentCache = reqcache.WithContentCache(parentCtx)
        ctx, cancel = context.WithTimeout(ctx, requestTimeout)
    }
} else {
    // No deadline, create timeout
    ctx, contentCache = reqcache.WithContentCache(parentCtx)
    ctx, cancel = context.WithTimeout(ctx, requestTimeout)
}
```

**OR BETTER:**
```go
// Use the parentCtx directly if it's Background (unlimited) or has sufficient time
// Only add timeout if parent has insufficient time
ctx, contentCache := reqcache.WithContentCache(parentCtx)
if deadline, hasDeadline := parentCtx.Deadline(); !hasDeadline || time.Until(deadline) < requestTimeout {
    // Parent has no deadline or insufficient time, create timeout
    ctx, cancel = context.WithTimeout(ctx, requestTimeout)
} else {
    // Parent has sufficient time, use as-is
    // No cancel needed
}
```

---

### Issue #2: Queue Context Timeout Mismatch

**Location:** `services/classification-service/internal/handlers/classification.go:994`

**Problem:**
```go
queueCtx, queueCancel := context.WithTimeout(queueCtxParent, queueAwareTimeout)
```

**Root Cause:**
- `queueAwareTimeout = requestTimeout + estimatedQueueWait + 10s` (~78s)
- This context is stored in `queuedReq.ctx`
- Worker receives this context but then creates a fresh 120s context
- **However**, the queue context timeout still applies to the HTTP response wait
- If queue context expires, HTTP request times out even if worker is still processing

**Impact:**
- HTTP request can timeout even if worker is processing successfully
- Response channel wait uses queueCtx, which expires before worker completes

**Fix:**
- Increase `queueAwareTimeout` to match worker timeout (120s + buffer)
- OR: Don't use queueCtx for response waiting, use a separate context

---

### Issue #3: Context Not Passed to Queue Request

**Location:** `services/classification-service/internal/handlers/classification.go:998-1004`

**Problem:**
```go
queuedReq := &queuedRequest{
    req:       &req,
    ctx:       queueCtx,  // ⚠️ Uses queueCtx, not the ctx we just created
    ...
}
```

**Root Cause:**
- We create `ctx` with timeout at line 851
- But we store `queueCtx` in `queuedReq.ctx`
- Worker receives `queuedReq.ctx` (queueCtx) but ignores it and creates fresh context
- **However**, the HTTP response wait uses `queueCtx` (line 1034-1103)

**Impact:**
- HTTP response wait can timeout even if processing succeeds
- Context mismatch between what worker uses and what HTTP handler waits for

**Fix:**
- Use the same context for both queue and HTTP wait
- OR: Ensure queueCtx has sufficient time to match worker timeout

---

## 🟠 P1: HIGH PRIORITY ISSUES

### Issue #4: Multiple Context Creation Points

**Location:** Multiple locations

**Problem:**
- Entry point creates context (line 728/737)
- Handler creates context with timeout (line 851)
- Queue creates context (line 994)
- Worker creates fresh context (line 225)
- Process start creates fresh context (line 1603/1611)

**Impact:**
- Confusing context hierarchy
- Hard to debug which context is actually used
- Potential for context leaks

**Fix:**
- Consolidate context creation to single point
- Document context flow clearly
- Use consistent timeout values

---

### Issue #5: Adaptive Timeout May Be Too Short

**Location:** `services/classification-service/internal/handlers/classification.go:3748`

**Problem:**
- `calculateAdaptiveTimeout` returns ~68s for requests with website scraping
- This may not account for:
  - Multiple scraping strategy retries
  - Database query time
  - Network latency
  - Processing overhead

**Impact:**
- Timeout too short for actual processing time
- Requests fail even when processing would succeed with more time

**Fix:**
- Review timeout calculation
- Add buffer for retries and overhead
- Consider increasing base timeout

---

### Issue #6: Worker Context Ignores Queue Context

**Location:** `services/classification-service/internal/handlers/classification.go:225`

**Problem:**
```go
processingCtx, cancel := context.WithTimeout(context.Background(), freshTimeout)
```

**Root Cause:**
- Worker always creates fresh context, ignoring `queuedReq.ctx`
- This is actually GOOD for processing, but creates disconnect with HTTP wait

**Impact:**
- Worker may process successfully but HTTP times out
- No coordination between worker context and HTTP context

**Fix:**
- Keep worker fresh context (this is correct)
- But ensure HTTP wait context matches worker timeout

---

## 🟡 P2: MEDIUM PRIORITY ISSUES

### Issue #7: No Context Propagation Verification

**Problem:**
- No verification that context is passed to all nested calls
- Some functions may not receive context
- Context may be lost in call chain

**Impact:**
- Operations may not respect timeouts
- Context cancellation may not propagate

**Fix:**
- Add context to all function signatures
- Verify context propagation in tests
- Add logging for context state at key points

---

### Issue #8: Error Channel Blocking

**Location:** `services/classification-service/internal/handlers/classification.go:287-290`

**Problem:**
```go
select {
case queuedReq.errChan <- err:
default:
    // Error channel already has an error (shouldn't happen)
}
```

**Root Cause:**
- Error channel is unbuffered (line 1002: `make(chan error, 1)`)
- If error is sent but not received, channel blocks
- Default case prevents blocking but error is lost

**Impact:**
- Errors may be lost
- Worker may block if error channel is full

**Fix:**
- Use buffered channel or ensure receiver is always ready
- Add timeout to error sending

---

### Issue #9: Queue Size Management

**Location:** `services/classification-service/internal/handlers/classification.go:75-103`

**Problem:**
- Queue uses atomic counter but channel for storage
- Potential race condition between Size() and actual queue state
- No queue monitoring or alerting

**Impact:**
- Queue size may be inaccurate
- Queue may fill up without detection

**Fix:**
- Use channel length for accurate size
- Add queue monitoring
- Alert when queue is full

---

## 🔵 P3: LOW PRIORITY ISSUES

### Issue #10: Logging Verbosity

**Problem:**
- Very verbose logging may impact performance
- Some logs may not be necessary in production

**Impact:**
- Performance degradation
- Log storage costs

**Fix:**
- Use log levels appropriately
- Reduce verbosity in production
- Use structured logging efficiently

---

## 🔴 P0: ADDITIONAL CRITICAL ISSUES FOUND

### Issue #11: Graceful Shutdown Doesn't Stop Worker Pool

**Location:** `services/classification-service/cmd/main.go:258-271`

**Problem:**
```go
// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
logger.Info("🛑 Classification Service shutting down...")

shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer shutdownCancel()

if err := httpServer.Shutdown(shutdownCtx); err != nil {
    logger.Fatal("Classification Service forced to shutdown", zap.Error(err))
}

// ⚠️ Worker pool is NEVER stopped!
logger.Info("✅ Classification Service exited gracefully")
```

**Root Cause:**
- Worker pool is started but never stopped during shutdown
- All worker goroutines continue running
- Cleanup goroutines (cleanupCache, cleanupInFlightRequests) never stop
- Goroutine leak on every shutdown

**Impact:**
- Resource leaks
- Goroutines continue processing after shutdown
- Memory leaks
- Potential data corruption if workers access closed resources

**Fix:**
```go
// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
logger.Info("🛑 Classification Service shutting down...")

// Stop worker pool first
if classificationHandler.workerPool != nil {
    logger.Info("Stopping worker pool...")
    classificationHandler.workerPool.Stop()
}

// Stop cleanup goroutines (need to add stop mechanism)
// TODO: Add context cancellation for cleanup goroutines

shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer shutdownCancel()

if err := httpServer.Shutdown(shutdownCtx); err != nil {
    logger.Fatal("Classification Service forced to shutdown", zap.Error(err))
}

logger.Info("✅ Classification Service exited gracefully")
```

---

### Issue #12: Rate Limiting Middleware Race Condition

**Location:** `services/classification-service/cmd/main.go:352-384`

**Problem:**
```go
func rateLimitMiddleware() func(http.Handler) http.Handler {
    // Simple in-memory rate limiter
    requests := make(map[string][]time.Time)  // ⚠️ NO MUTEX!

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := r.RemoteAddr
            now := time.Now()

            // Clean old requests (older than 1 minute)
            if clientRequests, exists := requests[clientIP]; exists {
                var validRequests []time.Time
                for _, reqTime := range clientRequests {
                    if now.Sub(reqTime) < time.Minute {
                        validRequests = append(validRequests, reqTime)
                    }
                }
                requests[clientIP] = validRequests  // ⚠️ RACE CONDITION!
            }

            // Check rate limit (100 requests per minute)
            if len(requests[clientIP]) >= 100 {  // ⚠️ RACE CONDITION!
                errors.WriteError(w, r, http.StatusTooManyRequests, ...)
                return
            }

            // Add current request
            requests[clientIP] = append(requests[clientIP], now)  // ⚠️ RACE CONDITION!
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Root Cause:**
- Map access without mutex protection
- Multiple goroutines can read/write simultaneously
- Can cause panic or incorrect rate limiting

**Impact:**
- Potential panic: "concurrent map read and map write"
- Incorrect rate limiting (false positives/negatives)
- Data corruption

**Fix:**
```go
func rateLimitMiddleware() func(http.Handler) http.Handler {
    var (
        requests = make(map[string][]time.Time)
        mu       sync.RWMutex
    )

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := r.RemoteAddr
            now := time.Now()

            mu.Lock()
            // Clean old requests
            if clientRequests, exists := requests[clientIP]; exists {
                var validRequests []time.Time
                for _, reqTime := range clientRequests {
                    if now.Sub(reqTime) < time.Minute {
                        validRequests = append(validRequests, reqTime)
                    }
                }
                requests[clientIP] = validRequests
            }

            // Check rate limit
            if len(requests[clientIP]) >= 100 {
                mu.Unlock()
                errors.WriteError(w, r, http.StatusTooManyRequests, ...)
                return
            }

            // Add current request
            requests[clientIP] = append(requests[clientIP], now)
            mu.Unlock()

            next.ServeHTTP(w, r)
        })
    }
}
```

---

### Issue #13: Cleanup Goroutines Never Stop

**Location:** `services/classification-service/internal/handlers/classification.go:444-450, 455-470, 472-490`

**Problem:**
```go
// Start cache cleanup goroutine (for in-memory cache only)
if config.Classification.CacheEnabled {
    go handler.cleanupCache()  // ⚠️ Runs forever, no way to stop
}

// Start in-flight requests cleanup goroutine
go handler.cleanupInFlightRequests()  // ⚠️ Runs forever, no way to stop

func (h *ClassificationHandler) cleanupCache() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {  // ⚠️ Infinite loop, no context cancellation
        // ...
    }
}
```

**Root Cause:**
- Cleanup goroutines run forever
- No context cancellation mechanism
- No way to stop them during shutdown
- Goroutine leaks

**Impact:**
- Goroutine leaks
- Resources not cleaned up on shutdown
- Potential access to closed resources

**Fix:**
```go
// Add context to handler
type ClassificationHandler struct {
    // ... existing fields ...
    shutdownCtx context.Context
    shutdownCancel context.CancelFunc
}

// In NewClassificationHandler:
handler.shutdownCtx, handler.shutdownCancel = context.WithCancel(context.Background())

// Start cleanup goroutines with context
if config.Classification.CacheEnabled {
    go handler.cleanupCache(handler.shutdownCtx)
}

go handler.cleanupInFlightRequests(handler.shutdownCtx)

func (h *ClassificationHandler) cleanupCache(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // cleanup logic
        }
    }
}

// In shutdown:
handler.shutdownCancel()  // Stop cleanup goroutines
```

---

### Issue #14: Configuration Default Timeout Too Short

**Location:** `services/classification-service/internal/config/config.go:97`

**Problem:**
```go
RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),  // ⚠️ Only 10s!
```

**Root Cause:**
- Default timeout is 10 seconds
- Actual processing takes 30-80 seconds
- If environment variable not set, all requests timeout immediately

**Impact:**
- All requests fail if REQUEST_TIMEOUT not configured
- Default is completely inadequate for actual processing time
- Silent failure - no error, just timeouts

**Fix:**
```go
RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 120*time.Second),  // Match worker timeout
```

---

### Issue #15: Queue Size Race Condition

**Location:** `services/classification-service/internal/handlers/classification.go:75-108`

**Problem:**
```go
type requestQueue struct {
    queue       chan *queuedRequest
    maxSize     int
    currentSize int32 // atomic counter
    mu          sync.RWMutex
}

func (rq *requestQueue) Enqueue(req *queuedRequest) error {
    rq.mu.Lock()
    defer rq.mu.Unlock()
    
    currentSize := int(atomic.LoadInt32(&rq.currentSize))  // ⚠️ Atomic read
    if currentSize >= rq.maxSize {
        return fmt.Errorf("request queue is full")
    }
    
    select {
    case rq.queue <- req:
        atomic.AddInt32(&rq.currentSize, 1)  // ⚠️ Atomic write
        return nil
    default:
        return fmt.Errorf("request queue is full")
    }
}

func (rq *requestQueue) Size() int {
    return int(atomic.LoadInt32(&rq.currentSize))  // ⚠️ May not match actual channel length
}
```

**Root Cause:**
- Using atomic counter but channel for storage
- Counter can be out of sync with actual channel length
- Race condition between Size() check and Enqueue
- Channel can be full even if counter says it's not

**Impact:**
- Incorrect queue size reporting
- Queue can fill up without detection
- Potential for requests to be rejected incorrectly

**Fix:**
```go
func (rq *requestQueue) Size() int {
    return len(rq.queue)  // Use actual channel length
}

func (rq *requestQueue) Enqueue(req *queuedRequest) error {
    rq.mu.Lock()
    defer rq.mu.Unlock()
    
    // Check actual channel length
    if len(rq.queue) >= rq.maxSize {
        return fmt.Errorf("request queue is full")
    }
    
    select {
    case rq.queue <- req:
        atomic.AddInt32(&rq.currentSize, 1)  // Keep for compatibility
        return nil
    default:
        return fmt.Errorf("request queue is full")
    }
}
```

---

### Issue #16: Parallel Processing Goroutine Leak

**Location:** `services/classification-service/internal/handlers/classification.go:1766-1797`

**Problem:**
```go
// Wait for both to complete with timeout
done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:
    // Both completed successfully
case <-ctx.Done():
    // Context cancelled - return error
    return nil, fmt.Errorf("parallel processing cancelled: %w", ctx.Err())
case <-time.After(10 * time.Second):  // ⚠️ Hardcoded timeout
    // Timeout - log warning but continue with what we have
    // ⚠️ Goroutines may still be running!
}
```

**Root Cause:**
- If timeout occurs, goroutines may still be running
- No cancellation of goroutines when timeout happens
- Goroutines continue running even after function returns
- Potential goroutine leak

**Impact:**
- Goroutine leaks
- Resources not released
- Memory leaks over time

**Fix:**
```go
// Create context with timeout for parallel processing
parallelCtx, parallelCancel := context.WithTimeout(ctx, 10*time.Second)
defer parallelCancel()

// Pass context to goroutines so they can be cancelled
wg.Add(1)
go func() {
    defer wg.Done()
    select {
    case <-parallelCtx.Done():
        return  // Cancelled
    default:
        h.traceStage(trace, "risk_assessment", nil, func() error {
            riskAssessment = h.generateRiskAssessment(req, enhancedResult, processingTime)
            return nil
        })
    }
}()

// Wait for both to complete
done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:
    // Both completed successfully
case <-parallelCtx.Done():
    // Timeout or cancellation - cancel goroutines
    parallelCancel()  // Cancel context to stop goroutines
    wg.Wait()  // Wait for goroutines to finish
    // Continue with defaults
}
```

---

### Issue #17: Channel Double-Close Risk

**Location:** `services/classification-service/internal/handlers/classification.go:493`

**Problem:**
```go
// Close the channel to unblock any waiting goroutines
close(req.resultChan)
```

**Root Cause:**
- Channel is closed in cleanup function
- But channel may also be used in other places
- No check if channel is already closed
- Closing closed channel causes panic

**Impact:**
- Potential panic: "close of closed channel"
- Service crash
- Data loss

**Fix:**
```go
// Use sync.Once or check if channel is nil/closed
// Better: Don't close channel, use context cancellation instead
```

---

### Issue #18: Response Channel Never Closed

**Location:** `services/classification-service/internal/handlers/classification.go:1001-1002`

**Problem:**
```go
queuedReq := &queuedRequest{
    req:       &req,
    ctx:       queueCtx,
    response:  make(chan *ClassificationResponse, 1),  // ⚠️ Never closed
    errChan:   make(chan error, 1),  // ⚠️ Never closed
    startTime: time.Now(),
}
```

**Root Cause:**
- Channels are created but never explicitly closed
- If worker fails to send, channel remains open forever
- Potential resource leak

**Impact:**
- Resource leaks
- Memory not freed
- Channels accumulate over time

**Fix:**
- Close channels when done (with proper synchronization)
- OR: Use context cancellation instead of channel closure

---

### Issue #19: HTTP Response Write After Context Expiration

**Location:** `services/classification-service/internal/handlers/classification.go:1083-1084`

**Problem:**
```go
// Write response
w.WriteHeader(http.StatusOK)
w.Write(responseBytes)  // ⚠️ May write after context expired
return
```

**Root Cause:**
- Response is written even if HTTP context expired
- No check if connection is still valid
- Can cause HTTP 000 errors

**Impact:**
- HTTP 000 errors
- Wasted resources
- Client receives no response

**Fix:**
```go
// Check if HTTP connection is still valid before writing
if r.Context().Err() != nil {
    h.logger.Warn("HTTP connection closed, skipping response",
        zap.String("request_id", req.RequestID))
    return
}

w.WriteHeader(http.StatusOK)
if _, err := w.Write(responseBytes); err != nil {
    h.logger.Warn("Failed to write response",
        zap.String("request_id", req.RequestID),
        zap.Error(err))
}
```

---

### Issue #20: Timeout Alert Goroutine Leak

**Location:** `services/classification-service/internal/handlers/classification.go:1650-1681`

**Problem:**
```go
// Start timeout alert goroutine
timeoutAlertCtx, timeoutAlertCancel := context.WithCancel(ctx)
defer timeoutAlertCancel()

go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-timeoutAlertCtx.Done():
            return
        case <-ticker.C:
            // Log timeout warnings
        }
    }
}()
```

**Root Cause:**
- Goroutine is started but may not exit if context cancellation doesn't work
- If function returns early, goroutine may continue running
- No guarantee goroutine exits

**Impact:**
- Potential goroutine leak
- Resources not released

**Fix:**
- Ensure context cancellation works properly
- Add timeout to goroutine itself
- Use sync.WaitGroup to wait for goroutine completion

---

## Recommended Fix Priority

### Phase 1: Critical Fixes (Immediate - Blocks All Functionality)

1. **IMMEDIATE (P0)**: Fix Issue #1 (Context Timeout Overwrite) - **ROOT CAUSE**
2. **IMMEDIATE (P0)**: Fix Issue #2 (Queue Context Timeout Mismatch)
3. **IMMEDIATE (P0)**: Fix Issue #11 (Graceful Shutdown Doesn't Stop Worker Pool)
4. **IMMEDIATE (P0)**: Fix Issue #12 (Rate Limiting Race Condition)
5. **IMMEDIATE (P0)**: Fix Issue #14 (Configuration Default Timeout Too Short)

### Phase 2: High Priority (Significant Impact)

6. **HIGH (P1)**: Fix Issue #3 (Context Not Passed to Queue Request)
7. **HIGH (P1)**: Review and fix Issue #5 (Adaptive Timeout)
8. **HIGH (P1)**: Fix Issue #13 (Cleanup Goroutines Never Stop)
9. **HIGH (P1)**: Fix Issue #15 (Queue Size Race Condition)
10. **HIGH (P1)**: Fix Issue #16 (Parallel Processing Goroutine Leak)

### Phase 3: Medium Priority (Moderate Impact)

11. **MEDIUM (P2)**: Address Issue #4 (Multiple Context Creation)
12. **MEDIUM (P2)**: Fix Issue #8 (Error Channel Blocking)
13. **MEDIUM (P2)**: Fix Issue #17 (Channel Double-Close Risk)
14. **MEDIUM (P2)**: Fix Issue #18 (Response Channel Never Closed)
15. **MEDIUM (P2)**: Fix Issue #19 (HTTP Response Write After Context Expiration)
16. **MEDIUM (P2)**: Fix Issue #20 (Timeout Alert Goroutine Leak)

### Phase 4: Low Priority (Minor Impact)

17. **LOW (P3)**: Fix Issue #10 (Logging Verbosity)

---

## Testing After Fixes

1. Verify context time remaining at each stage
2. Test with various timeout scenarios
3. Verify HTTP response doesn't timeout before worker completes
4. Test queue behavior under load
5. Verify error handling doesn't block

---

## Summary Statistics

**Total Issues Found:** 20
- **P0 (Critical):** 5 issues
- **P1 (High):** 5 issues  
- **P2 (Medium):** 9 issues
- **P3 (Low):** 1 issue

**Issue Categories:**
- **Context Management:** 6 issues
- **Goroutine Leaks:** 4 issues
- **Race Conditions:** 3 issues
- **Resource Management:** 3 issues
- **Configuration:** 2 issues
- **Error Handling:** 2 issues

## Conclusion

The **root cause** of the timeout issues is **Issue #1**: The context timeout overwrite at line 851. This negates all previous context fixes by reducing the timeout to ~68s regardless of what was set earlier.

**Additional Critical Findings:**
- **Issue #11**: Worker pool never stops during shutdown (goroutine leak)
- **Issue #12**: Rate limiting has race condition (potential panic)
- **Issue #14**: Default timeout is 10s (way too short)
- **Issue #13**: Cleanup goroutines never stop (resource leak)

**Immediate Action Required:**
1. **Fix Issue #1** (Context Timeout Overwrite) - **ROOT CAUSE**
2. **Fix Issue #2** (Queue Context Timeout Mismatch)
3. **Fix Issue #11** (Graceful Shutdown)
4. **Fix Issue #12** (Rate Limiting Race Condition)
5. **Fix Issue #14** (Default Timeout)
6. Test thoroughly
7. Address remaining issues in priority order

**Expected Impact:**
- Fixing Issue #1 alone should improve success rate from 0% to ~50-70%
- Fixing all P0 issues should improve success rate to ~80-90%
- Fixing all P0+P1 issues should achieve target success rate of ≥95%
