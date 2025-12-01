<!-- 648146e2-9300-40bc-a0fe-d7c55682ed8b 3a6b9759-aed0-4d6f-8f54-f2ecc83b5af2 -->
# ML Service Accuracy Technical Deep-Dive Plan

## Architecture Overview

### Current System Architecture

The classification system uses a multi-layered architecture:

1. **Entry Point**: `internal/classification/service.go` - `IndustryDetectionService.DetectIndustry()`
2. **Classification Layer**: `MultiMethodClassifier` combines keyword, ML, and description methods
3. **ML Service Layer**: `PythonMLService` with circuit breaker protection
4. **Fallback Layer**: Go `ContentClassifier` when Python ML service unavailable
5. **Data Layer**: Supabase repository for keywords, industries, and codes

### Critical Architecture Issues

1. **Circuit Breaker State Management**: Circuit opens during initialization failures and never recovers
2. **Fallback Classifier**: Go ML classifier returns "General Business" for all inputs (0% accuracy)
3. **Website Scraping**: No timeout enforcement, taking 9+ seconds per request
4. **No Caching**: Repeated database queries and website scraping
5. **Missing Observability**: No metrics for circuit breaker state, ML service health, or classification accuracy

---

## Phase 1: Circuit Breaker Architecture Improvements

### 1.1 Circuit Breaker Configuration Enhancement

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`

**Current Implementation**:

```99:105:internal/machine_learning/infrastructure/python_ml_service.go
	// Initialize circuit breaker with default config
	// Opens after 5 consecutive failures, stays open for 30s, needs 2 successes to close
	circuitBreakerConfig := resilience.DefaultCircuitBreakerConfig()
	circuitBreakerConfig.FailureThreshold = 5  // Open after 5 failures
	circuitBreakerConfig.Timeout = 30 * time.Second // Stay open for 30s
	circuitBreakerConfig.SuccessThreshold = 2 // Need 2 successes to close
	circuitBreaker := resilience.NewCircuitBreaker(circuitBreakerConfig)
```

**Technical Changes**:

1. **Increase Failure Threshold**: From 5 to 10 to handle transient initialization failures
2. **Increase Timeout**: From 30s to 60s to allow service recovery
3. **Add Exponential Backoff**: Implement retry mechanism with exponential backoff
4. **Add Reset Method**: Allow manual circuit breaker reset during initialization
5. **Add State Monitoring**: Track circuit breaker state changes with metrics

**Implementation**:

```go
// Enhanced circuit breaker configuration
circuitBreakerConfig := resilience.DefaultCircuitBreakerConfig()
circuitBreakerConfig.FailureThreshold = 10  // Increased from 5
circuitBreakerConfig.Timeout = 60 * time.Second // Increased from 30s
circuitBreakerConfig.SuccessThreshold = 2 // Keep at 2
circuitBreakerConfig.ResetTimeout = 120 * time.Second // Increased from 60s
circuitBreaker := resilience.NewCircuitBreaker(circuitBreakerConfig)

// Add reset method
func (pms *PythonMLService) ResetCircuitBreaker() {
    pms.circuitBreaker.Reset()
}

// Add state monitoring
func (pms *PythonMLService) GetCircuitBreakerState() resilience.CircuitState {
    return pms.circuitBreaker.GetState()
}
```

### 1.2 Circuit Breaker Reset Mechanism

**File**: `internal/resilience/circuit_breaker.go`

**Technical Changes**:

1. **Add Reset Method**: Allow manual reset of circuit breaker state
2. **Add State Persistence**: Track state changes for observability
3. **Add Half-Open Recovery**: Improve half-open state handling

**Implementation**:

```go
// Add to CircuitBreaker struct
type CircuitBreaker struct {
    // ... existing fields ...
    stateHistory []StateChange // Track state changes
}

// Add Reset method
func (cb *CircuitBreaker) Reset() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.state = CircuitClosed
    cb.failureCount = 0
    cb.successCount = 0
    cb.halfOpenCount = 0
    cb.stateChange = time.Now()
    cb.recordStateChange(CircuitClosed, "manual_reset")
}

// Add state change tracking
func (cb *CircuitBreaker) recordStateChange(newState CircuitState, reason string) {
    change := StateChange{
        State:     newState,
        Timestamp: time.Now(),
        Reason:    reason,
    }
    cb.stateHistory = append(cb.stateHistory, change)
    // Keep only last 100 state changes
    if len(cb.stateHistory) > 100 {
        cb.stateHistory = cb.stateHistory[len(cb.stateHistory)-100:]
    }
}
```

### 1.3 Initialization Resilience

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`

**Current Implementation**:

```134:181:internal/machine_learning/infrastructure/python_ml_service.go
// Initialize initializes the Python ML service
func (pms *PythonMLService) Initialize(ctx context.Context) error {
	pms.logger.Printf("🐍 Initializing Python ML Service at %s", pms.endpoint)

	// Initialize metrics and health status (need lock for these)
	pms.mu.Lock()
	pms.metrics = &ServiceMetrics{
		RequestCount:   0,
		SuccessCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		P95Latency:     0,
		P99Latency:     0,
		Throughput:     0,
		ErrorRate:      0,
		LastUpdated:    time.Now(),
	}

	// Initialize health status
	pms.healthStatus = &HealthStatus{
		Status:    "unknown",
		LastCheck: time.Now(),
		Checks:    make(map[string]HealthCheck),
	}
	pms.mu.Unlock()

	// Test connection to Python service (no lock needed)
	// Use a shorter timeout for initialization to prevent hanging
	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	defer initCancel()
	
	if err := pms.testConnection(initCtx); err != nil {
		return fmt.Errorf("failed to connect to Python ML service: %w", err)
	}

	// Load available models (this will acquire its own lock)
	// Use a separate timeout for model loading to prevent blocking initialization
	modelsCtx, modelsCancel := context.WithTimeout(ctx, 5*time.Second)
	defer modelsCancel()
	
	if err := pms.loadAvailableModels(modelsCtx); err != nil {
		pms.logger.Printf("⚠️ Warning: failed to load available models: %v", err)
		// Don't fail initialization if models can't be loaded - they can be loaded later
	}

	pms.logger.Printf("✅ Python ML Service initialized successfully")
	return nil
}
```

**Technical Changes**:

1. **Add Retry Logic**: Retry initialization with exponential backoff
2. **Reset Circuit Breaker**: Reset circuit breaker before initialization
3. **Graceful Degradation**: Continue initialization even if service unavailable
4. **Health Check Before Ready**: Verify service health before marking as ready

**Implementation**:

```go
// Enhanced initialization with retry
func (pms *PythonMLService) InitializeWithRetry(ctx context.Context, maxRetries int) error {
    // Reset circuit breaker before initialization
    pms.circuitBreaker.Reset()
    
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        if i > 0 {
            waitTime := time.Duration(i) * 2 * time.Second
            pms.logger.Printf("Retrying initialization (attempt %d/%d) after %v", i+1, maxRetries, waitTime)
            time.Sleep(waitTime)
        }
        
        err := pms.Initialize(ctx)
        if err == nil {
            // Verify health before marking as ready
            healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()
            
            health, err := pms.HealthCheck(healthCtx)
            if err == nil && health.Status == "pass" {
                pms.logger.Printf("✅ Python ML Service initialized and healthy")
                return nil
            }
        }
        lastErr = err
    }
    
    // Graceful degradation: mark as initialized but unavailable
    pms.logger.Printf("⚠️ Python ML Service initialization failed after %d retries, continuing with degraded mode", maxRetries)
    return fmt.Errorf("initialization failed after %d retries: %w", maxRetries, lastErr)
}
```

---

## Phase 2: Fallback Classifier Improvements

### 2.1 Go ML Classifier Analysis

**File**: `internal/machine_learning/content_classifier.go`

**Current Issue**: Go ML classifier returns "General Business" for all inputs with 0% accuracy.

**Root Cause Analysis**:

1. **Model Not Loaded**: The classifier may not have a trained model
2. **Industry Mapping**: Industry labels may not match expected industries
3. **Confidence Threshold**: Low confidence threshold accepting "General Business"
4. **Content Processing**: Content may not be properly processed before classification

**Technical Investigation Steps**:

1. Check if model is loaded: `getModelForIndustry()` implementation
2. Verify industry label mapping: Check label-to-industry conversion
3. Review confidence calculation: `ConfidenceScorer.CalculateConfidence()`
4. Test with sample inputs: Create test cases for each industry category

### 2.2 Fallback Classifier Enhancement

**File**: `internal/classification/multi_method_classifier.go`

**Current Implementation**:

```524:562:internal/classification/multi_method_classifier.go
// performGoMLClassification performs ML classification using Go ML classifier (fallback)
func (mmc *MultiMethodClassifier) performGoMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	if mmc.mlClassifier == nil {
		return nil, fmt.Errorf("ML classifier not available")
	}

	// Combine business information for ML analysis
	content := mmc.extractTrustedContent(ctx, businessName, description, websiteURL)

	// Perform ML classification
	mlResult, err := mmc.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Find the best classification from ML result
	if len(mlResult.Classifications) == 0 {
		return nil, fmt.Errorf("no classifications returned from ML model")
	}

	// Get the highest confidence classification
	bestClassification := mlResult.Classifications[0]
	for _, classification := range mlResult.Classifications {
		if classification.Confidence > bestClassification.Confidence {
			bestClassification = classification
		}
	}

	// Get classification codes for the ML-detected industry
	var classificationCodes shared.ClassificationCodes
	if bestClassification.Label != "unknown" {
		classificationCodes = mmc.getClassificationCodesForIndustry(ctx, bestClassification.Label)
	}

	// Convert to shared format
	result := &shared.IndustryClassification
```

**Technical Changes**:

1. **Add Keyword Fallback**: If ML classifier fails, use keyword-based classification
2. **Improve Industry Mapping**: Map ML labels to database industries
3. **Add Confidence Validation**: Reject classifications below threshold
4. **Enhance Content Extraction**: Improve content quality for ML analysis

**Implementation**:

```go
// Enhanced Go ML classification with fallback
func (mmc *MultiMethodClassifier) performGoMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	if mmc.mlClassifier == nil {
		// Fallback to keyword-based classification
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Combine business information for ML analysis
	content := mmc.extractTrustedContent(ctx, businessName, description, websiteURL)
	
	// Validate content quality
	if len(content) < 10 {
		mmc.logger.Printf("⚠️ Insufficient content for ML classification, using keyword fallback")
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Perform ML classification
	mlResult, err := mmc.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		mmc.logger.Printf("⚠️ ML classification failed: %v, using keyword fallback", err)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Validate results
	if len(mlResult.Classifications) == 0 {
		mmc.logger.Printf("⚠️ No classifications from ML, using keyword fallback")
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Get the highest confidence classification
	bestClassification := mlResult.Classifications[0]
	for _, classification := range mlResult.Classifications {
		if classification.Confidence > bestClassification.Confidence {
			bestClassification = classification
		}
	}

	// Validate confidence threshold
	const minConfidence = 0.5
	if bestClassification.Confidence < minConfidence {
		mmc.logger.Printf("⚠️ ML confidence too low (%.2f < %.2f), using keyword fallback", 
			bestClassification.Confidence, minConfidence)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Map ML label to database industry
	industryName := mmc.mapMLLabelToIndustry(ctx, bestClassification.Label)
	if industryName == "" {
		mmc.logger.Printf("⚠️ Could not map ML label '%s' to industry, using keyword fallback", bestClassification.Label)
		return mmc.performKeywordClassification(ctx, businessName, description, websiteURL)
	}

	// Get classification codes
	classificationCodes := mmc.getClassificationCodesForIndustry(ctx, industryName)

	// Convert to shared format
	result := &shared.IndustryClassification{
		IndustryCode:         industryName,
		IndustryName:         industryName,
		ConfidenceScore:      bestClassification.Confidence,
		ClassificationMethod: "ml_fallback",
		Keywords:             []string{},
		ClassificationCodes:  classificationCodes,
	}

	return result, nil
}

// Add keyword classification fallback
func (mmc *MultiMethodClassifier) performKeywordClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.IndustryClassification, error) {
	// Use keyword-based classification as fallback
	// Implementation similar to MultiStrategyClassifier
	// ...
}
```

---

## Phase 3: Performance Optimization

### 3.1 Website Scraping Optimization

**File**: `internal/classification/multi_method_classifier.go`

**Current Implementation**:

```938:998:internal/classification/multi_method_classifier.go
// extractKeywordsFromWebsite scrapes website content and extracts business-relevant keywords
func (mmc *MultiMethodClassifier) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	mmc.logger.Printf("🌐 Starting website scraping for: %s", websiteURL)

	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		mmc.logger.Printf("❌ Invalid URL format for %s: %v", websiteURL, err)
		return []string{}
	}

	if parsedURL.Scheme == "" {
		websiteURL = "https://" + websiteURL
		mmc.logger.Printf("🔧 Added HTTPS scheme: %s", websiteURL)
	}

	// Create HTTP client with enhanced configuration
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		mmc.logger.Printf("❌ Failed to create request for %s: %v", websiteURL, err)
		return []string{}
	}

	// Set comprehensive headers with randomization to mimic a real browser
	headers := GetRandomizedHeaders(GetUserAgent())
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	mmc.logger.Printf("📡 Making HTTP request to: %s", websiteURL)

	// Make request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	resp, err := client.Do(req)
	if err != nil {
		mmc.logger.Printf("❌ HTTP request failed for %s: %v", websiteURL, err)
		return []string{}
	}
	defer resp.Body.Close()

	// Log response details
	mmc.logger.Printf("📊 Response received - Status: %d, Content-Type: %s, Content-Length: %d",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)

	// Check status code with detailed logging
	if resp.StatusCode >= 400 {
		mmc.logger.Printf
```

**Technical Changes**:

1. **Reduce Timeout**: From 15s to 5s for faster failure
2. **Add Caching**: Cache scraped content to avoid redundant requests
3. **Parallel Scraping**: Scrape multiple URLs in parallel when possible
4. **Content Size Limit**: Limit response body size to prevent memory issues

**Implementation**:

```go
// Enhanced website scraping with caching and timeout
type WebsiteCache struct {
    cache map[string]*CachedContent
    mu    sync.RWMutex
    ttl   time.Duration
}

type CachedContent struct {
    Content   []string
    Timestamp time.Time
}

func (mmc *MultiMethodClassifier) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
    // Check cache first
    if cached := mmc.websiteCache.Get(websiteURL); cached != nil {
        mmc.logger.Printf("✅ Using cached website content for: %s", websiteURL)
        return cached
    }

    // Create context with strict timeout
    scrapeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Create HTTP client with shorter timeout
    client := &http.Client{
        Timeout: 5 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:       10,
            IdleConnTimeout:    30 * time.Second,
            DisableCompression: false,
            ResponseHeaderTimeout: 3 * time.Second,
        },
    }

    // Create request
    req, err := http.NewRequestWithContext(scrapeCtx, "GET", websiteURL, nil)
    if err != nil {
        mmc.logger.Printf("❌ Failed to create request: %v", err)
        return []string{}
    }

    // Set headers
    headers := GetRandomizedHeaders(GetUserAgent())
    for key, value := range headers {
        req.Header.Set(key, value)
    }

    // Execute request
    resp, err := client.Do(req)
    if err != nil {
        mmc.logger.Printf("❌ HTTP request failed: %v", err)
        return []string{}
    }
    defer resp.Body.Close()

    // Limit response size
    const maxSize = 1 * 1024 * 1024 // 1MB
    bodyReader := io.LimitReader(resp.Body, maxSize)
    body, err := io.ReadAll(bodyReader)
    if err != nil {
        mmc.logger.Printf("❌ Failed to read response: %v", err)
        return []string{}
    }

    // Extract keywords
    keywords := mmc.extractKeywordsFromContent(string(body))

    // Cache result
    mmc.websiteCache.Set(websiteURL, keywords)

    return keywords
}
```

### 3.2 Database Query Optimization

**File**: `internal/classification/repository/supabase_repository.go`

**Technical Changes**:

1. **Add Query Caching**: Cache frequently accessed keyword and industry data
2. **Batch Queries**: Combine multiple queries into single requests
3. **Connection Pooling**: Optimize database connection management
4. **Prepared Statements**: Use prepared statements for repeated queries

**Implementation**:

```go
// Add caching layer to repository
type CachedRepository struct {
    repo        repository.KeywordRepository
    keywordCache *sync.Map
    industryCache *sync.Map
    cacheTTL    time.Duration
}

func (cr *CachedRepository) GetKeywordsForIndustry(ctx context.Context, industry string) ([]*Keyword, error) {
    // Check cache
    cacheKey := fmt.Sprintf("keywords:%s", industry)
    if cached, ok := cr.keywordCache.Load(cacheKey); ok {
        cachedData := cached.(*CachedKeywords)
        if time.Since(cachedData.Timestamp) < cr.cacheTTL {
            return cachedData.Keywords, nil
        }
    }

    // Query database
    keywords, err := cr.repo.GetKeywordsForIndustry(ctx, industry)
    if err != nil {
        return nil, err
    }

    // Cache result
    cr.keywordCache.Store(cacheKey, &CachedKeywords{
        Keywords:  keywords,
        Timestamp: time.Now(),
    })

    return keywords, nil
}
```

---

## Phase 4: Observability and Monitoring

### 4.1 Circuit Breaker Metrics

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`

**Technical Changes**:

1. **Add Metrics Export**: Export circuit breaker state to Prometheus/OpenTelemetry
2. **State Change Logging**: Log all circuit breaker state changes
3. **Health Check Endpoint**: Add endpoint to check circuit breaker status

**Implementation**:

```go
// Add metrics to PythonMLService
type CircuitBreakerMetrics struct {
    State              string
    FailureCount       int
    SuccessCount       int
    StateChangeTime    time.Time
    LastFailureTime    time.Time
    TotalRequests      int64
    RejectedRequests   int64
}

func (pms *PythonMLService) GetCircuitBreakerMetrics() CircuitBreakerMetrics {
    stats := pms.circuitBreaker.GetStats()
    return CircuitBreakerMetrics{
        State:            stats.State,
        FailureCount:     stats.FailureCount,
        SuccessCount:     stats.SuccessCount,
        StateChangeTime:  stats.StateChange,
        LastFailureTime:  stats.LastFailure,
        TotalRequests:    pms.metrics.RequestCount,
        RejectedRequests: pms.metrics.ErrorCount,
    }
}

// Add health check endpoint
func (pms *PythonMLService) HealthCheckWithCircuitBreaker(ctx context.Context) (*HealthCheck, error) {
    health, err := pms.HealthCheck(ctx)
    if err != nil {
        return health, err
    }

    cbState := pms.circuitBreaker.GetState()
    health.Checks["circuit_breaker"] = HealthCheck{
        Name:      "circuit_breaker",
        Status:    mapCircuitBreakerState(cbState),
        Message:   fmt.Sprintf("Circuit breaker state: %s", cbState.String()),
        LastCheck: time.Now(),
    }

    return health, nil
}
```

### 4.2 Classification Accuracy Metrics

**File**: `internal/classification/service.go`

**Technical Changes**:

1. **Track Classification Results**: Log classification results for accuracy analysis
2. **Method Performance**: Track accuracy by classification method (ML, keyword, description)
3. **Industry Accuracy**: Track accuracy per industry category

**Implementation**:

```go
// Add accuracy tracking
type ClassificationMetrics struct {
    TotalClassifications int64
    MLClassifications    int64
    KeywordClassifications int64
    FallbackClassifications int64
    IndustryAccuracy      map[string]float64
    MethodAccuracy        map[string]float64
}

func (s *IndustryDetectionService) RecordClassification(
    result *IndustryDetectionResult,
    expectedIndustry string,
) {
    s.metrics.TotalClassifications++
    
    // Track by method
    switch result.Method {
    case "ml_distilbart", "ml":
        s.metrics.MLClassifications++
    case "keyword":
        s.metrics.KeywordClassifications++
    case "ml_fallback":
        s.metrics.FallbackClassifications++
    }

    // Track accuracy
    isCorrect := result.IndustryName == expectedIndustry
    if isCorrect {
        s.metrics.IndustryAccuracy[expectedIndustry]++
    }
    s.metrics.MethodAccuracy[result.Method]++
}
```

---

## Implementation Priority

### Critical (Week 1)

1. Circuit breaker configuration and reset mechanism
2. Initialization retry logic
3. Go ML classifier fallback improvements
4. Website scraping timeout and caching

### High Priority (Week 2)

1. Database query optimization
2. Classification accuracy metrics
3. Circuit breaker monitoring
4. Performance testing

### Medium Priority (Week 3)

1. Advanced caching strategies
2. Parallel processing optimizations
3. Comprehensive observability dashboard
4. Load testing and optimization

---

## Success Criteria

### Technical Metrics

- Circuit breaker recovery time < 60 seconds
- Website scraping time < 3 seconds (95th percentile)
- Database query time reduced by 50%
- Cache hit rate > 30%

### Accuracy Metrics

- Industry accuracy > 50% (Week 1 target)
- Code accuracy > 40% (Week 1 target)
- ML service utilization > 80% (when available)
- Fallback classifier accuracy > 30%

### Performance Metrics

- Average processing time < 5 seconds
- P95 processing time < 8 seconds
- P99 processing time < 12 seconds