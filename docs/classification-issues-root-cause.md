# Classification Issues - Root Cause Analysis

## Critical Finding: Incomplete Request Handler

### Issue 1: Missing Completion Logs - ROOT CAUSE

**Location**: `internal/api/handlers/intelligent_routing_handler.go` line 79

**Problem**:

```go
// Route request through intelligent routing system
response, err := h.router.RouteRequest(ctx, req)
if err != nil {
    h.handleError(w, err, http.StatusInternalServerError, "classification_failed", requestID)
    return
}

// Record metrics
h.recordMetrics(ctx, requestID, true, time.Since(startTime))

// Log successful completion
h.logger.WithComponent("intelligent_routing_handler").Info("classification_request_completed", ...)

// Return response
h.writeResponse(w, response, http.StatusOK)
```

**The Bug**:

- `RouteRequest` returns `(*ServiceEndpoint, error)` - it only routes to an endpoint
- The handler treats the `ServiceEndpoint` as if it's a classification response
- **The handler never actually calls the classification service!**
- The classification service (`DetectIndustry`) is never invoked

**Why logs are missing**:

1. The handler routes to an endpoint but doesn't process the classification
2. The classification service is likely being called from somewhere else (maybe a different code path)
3. But the main API handler isn't completing the request properly
4. This causes the frontend to retry, creating duplicate requests

### Issue 2: Duplicate Requests - ROOT CAUSE

**Problem**:

- Handler doesn't actually process classification
- Returns `ServiceEndpoint` as response (wrong type)
- Frontend receives invalid response
- Frontend retries, creating duplicate requests
- No request deduplication exists

**Evidence**:

- 50+ duplicate "Starting multi-strategy classification" logs
- All show "Cache MISS" (cache key might not be normalized)
- Handler logs "classification_request_completed" but never actually classified

## Solution

### Fix 1: Complete the Request Handler

**File**: `internal/api/handlers/intelligent_routing_handler.go`

**Required Changes**:

1. **Add classification service to handler**:

```go
type IntelligentRoutingHandler struct {
    router          *routing.IntelligentRouter
    detectionService *classification.IndustryDetectionService  // ADD THIS
    logger          *observability.Logger
    metrics         *observability.Metrics
    tracer          trace.Tracer
    requestIDGen    func() string
}
```

2. **Update constructor**:

```go
func NewIntelligentRoutingHandler(
    router *routing.IntelligentRouter,
    detectionService *classification.IndustryDetectionService,  // ADD THIS
    logger *observability.Logger,
    metrics *observability.Metrics,
    tracer trace.Tracer,
) *IntelligentRoutingHandler {
    return &IntelligentRoutingHandler{
        router:          router,
        detectionService: detectionService,  // ADD THIS
        logger:          logger,
        metrics:         metrics,
        tracer:          tracer,
        requestIDGen:    generateRequestID,
    }
}
```

3. **Fix ClassifyBusiness method**:

```go
func (h *IntelligentRoutingHandler) ClassifyBusiness(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    requestID := h.requestIDGen()

    ctx, span := h.tracer.Start(r.Context(), "IntelligentRoutingHandler.ClassifyBusiness")
    defer span.End()

    span.SetAttributes(
        attribute.String("request.id", requestID),
        attribute.String("http.method", r.Method),
        attribute.String("http.url", r.URL.String()),
    )

    // Add request ID to context
    ctx = context.WithValue(ctx, "request_id", requestID)

    // Parse and validate request
    req, err := h.parseClassificationRequest(r)
    if err != nil {
        h.handleError(w, err, http.StatusBadRequest, "invalid_request", requestID)
        return
    }

    // Set request ID
    req.ID = requestID

    // Log request start
    h.logger.WithComponent("intelligent_routing_handler").Info("classification_request_started", map[string]interface{}{
        "request_id":    requestID,
        "business_name": req.BusinessName,
        "website_url":   req.WebsiteURL,
        "user_agent":    r.UserAgent(),
    })

    // FIX: Actually perform classification
    classificationResult, err := h.detectionService.DetectIndustry(
        ctx,
        req.BusinessName,
        req.Description,
        req.WebsiteURL,
    )
    if err != nil {
        h.handleError(w, err, http.StatusInternalServerError, "classification_failed", requestID)
        return
    }

    // Convert to API response format
    response := &shared.BusinessClassificationResponse{
        ID:                    requestID,
        BusinessName:          req.BusinessName,
        DetectedIndustry:      classificationResult.IndustryName,
        Confidence:            classificationResult.Confidence,
        ClassificationMethod:  classificationResult.Method,
        ProcessingTime:        classificationResult.ProcessingTime,
        CreatedAt:             classificationResult.CreatedAt,
        Timestamp:             time.Now(),
        Classifications: []shared.IndustryClassification{
            {
                IndustryName:         classificationResult.IndustryName,
                ConfidenceScore:      classificationResult.Confidence,
                ClassificationMethod: classificationResult.Method,
                Keywords:             classificationResult.Keywords,
            },
        },
        PrimaryClassification: &shared.IndustryClassification{
            IndustryName:         classificationResult.IndustryName,
            ConfidenceScore:      classificationResult.Confidence,
            ClassificationMethod: classificationResult.Method,
            Keywords:             classificationResult.Keywords,
        },
        OverallConfidence:     classificationResult.Confidence,
        ClassificationReasoning: classificationResult.Reasoning,
        Metadata: map[string]interface{}{
            "method": classificationResult.Method,
            "request_id": requestID,
        },
    }

    // Record metrics
    h.recordMetrics(ctx, requestID, true, time.Since(startTime))

    // Log successful completion
    h.logger.WithComponent("intelligent_routing_handler").Info("classification_request_completed", map[string]interface{}{
        "request_id":         requestID,
        "business_name":      req.BusinessName,
        "detected_industry":  classificationResult.IndustryName,
        "confidence":         classificationResult.Confidence,
        "processing_time_ms": time.Since(startTime).Milliseconds(),
        "response_status":    "success",
    })

    // Return response
    h.writeResponse(w, response, http.StatusOK)
}
```

### Fix 2: Add Request Deduplication

**File**: `internal/classification/service.go`

Add in-flight request tracking to prevent duplicate processing:

```go
type IndustryDetectionService struct {
    // ... existing fields ...
    inFlightRequests sync.Map // map[string]*inFlightRequest
}

type inFlightRequest struct {
    resultChan chan *IndustryDetectionResult
    errChan    chan error
    done       bool
    mu         sync.Mutex
}

func (s *IndustryDetectionService) DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
    // Generate cache key for deduplication
    cacheKey := fmt.Sprintf("%s|%s|%s", normalizeString(businessName), normalizeString(description), normalizeString(websiteURL))

    // Check for in-flight request
    if existing, found := s.inFlightRequests.Load(cacheKey); found {
        req := existing.(*inFlightRequest)
        req.mu.Lock()
        if !req.done {
            req.mu.Unlock()
            // Wait for existing request
            select {
            case result := <-req.resultChan:
                s.logger.Printf("♻️ [Deduplication] Reusing in-flight request for: %s", businessName)
                return result, nil
            case err := <-req.errChan:
                return nil, err
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
        req.mu.Unlock()
    }

    // Create new in-flight request
    resultChan := make(chan *IndustryDetectionResult, 1)
    errChan := make(chan error, 1)
    inFlight := &inFlightRequest{
        resultChan: resultChan,
        errChan:    errChan,
        done:       false,
    }
    s.inFlightRequests.Store(cacheKey, inFlight)

    // Perform classification in goroutine
    go func() {
        result, err := s.performClassification(ctx, businessName, description, websiteURL)
        inFlight.mu.Lock()
        inFlight.done = true
        inFlight.mu.Unlock()
        if err != nil {
            errChan <- err
        } else {
            resultChan <- result
        }
        // Clean up after a delay to allow concurrent requests to read
        time.AfterFunc(5*time.Second, func() {
            s.inFlightRequests.Delete(cacheKey)
        })
    }()

    // Wait for result
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (s *IndustryDetectionService) performClassification(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
    // Move existing DetectIndustry logic here
    // ... existing code ...
}

func normalizeString(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}
```

### Fix 3: Update Route Registration

**File**: `internal/api/routes/routes.go`

Update `CreateIntelligentRoutingHandler` to include detection service:

```go
func CreateIntelligentRoutingHandler(
    router *routing.IntelligentRouter,
    detectionService *classification.IndustryDetectionService,  // ADD THIS
    logger *observability.Logger,
    metrics *observability.Metrics,
    tracer trace.Tracer,
) *handlers.IntelligentRoutingHandler {
    return handlers.NewIntelligentRoutingHandler(
        router,
        detectionService,  // ADD THIS
        logger,
        metrics,
        tracer,
    )
}
```

## Summary

**Root Causes**:

1. **Missing completion logs**: Handler doesn't call classification service, so completion never happens
2. **Duplicate requests**: Handler returns wrong response type, causing frontend retries
3. **No deduplication**: Multiple concurrent requests process the same business

**Fixes Required**:

1. Add `IndustryDetectionService` to `IntelligentRoutingHandler`
2. Actually call `DetectIndustry` in the handler
3. Convert result to proper API response format
4. Add request deduplication to prevent duplicate processing
5. Update route registration to pass detection service

**Expected Outcome**:

- Completion logs will appear
- No duplicate requests
- Proper API responses
- Better cache hit rates
