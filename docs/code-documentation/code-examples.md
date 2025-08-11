# KYB Platform - Code Examples

This document provides comprehensive code examples for common patterns and implementations in the KYB Platform. Each example includes complete, runnable code with explanations and best practices.

## Table of Contents

1. [API Handler Examples](#api-handler-examples)
2. [Service Layer Examples](#service-layer-examples)
3. [Database Operations](#database-operations)
4. [Middleware Examples](#middleware-examples)
5. [Testing Examples](#testing-examples)
6. [Configuration Examples](#configuration-examples)
7. [Error Handling Patterns](#error-handling-patterns)
8. [Common Utilities](#common-utilities)

## API Handler Examples

### Basic API Handler

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/classification"
    "github.com/pcraw4d/business-verification/internal/observability"
)

// ClassificationRequest represents the request body for business classification
type ClassificationRequest struct {
    BusinessName string            `json:"business_name" validate:"required,min=1,max=200"`
    Description  string            `json:"description,omitempty"`
    Industry     string            `json:"industry,omitempty"`
    Location     string            `json:"location,omitempty"`
    Metadata     map[string]string `json:"metadata,omitempty"`
}

// ClassificationResponse represents the response from classification
type ClassificationResponse struct {
    ID           string                    `json:"id"`
    BusinessName string                    `json:"business_name"`
    Primary      classification.Result     `json:"primary"`
    Alternatives []classification.Result   `json:"alternatives"`
    Confidence   float64                   `json:"confidence"`
    Method       string                    `json:"method"`
    CreatedAt    time.Time                 `json:"created_at"`
    ProcessingTime int64                   `json:"processing_time_ms"`
}

// ClassificationHandler handles business classification requests
type ClassificationHandler struct {
    service    *classification.Service
    logger     *observability.Logger
    validator  *validator.Validate
}

// NewClassificationHandler creates a new classification handler
func NewClassificationHandler(service *classification.Service, logger *observability.Logger) *ClassificationHandler {
    return &ClassificationHandler{
        service:   service,
        logger:    logger,
        validator: validator.New(),
    }
}

// Classify handles POST /v1/classify requests
func (h *ClassificationHandler) Classify(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Parse and validate request
    var req ClassificationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Failed to decode request", "error", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if err := h.validator.Struct(req); err != nil {
        h.logger.Error("Request validation failed", "error", err)
        http.Error(w, "Validation failed", http.StatusBadRequest)
        return
    }
    
    // Perform classification
    result, err := h.service.Classify(r.Context(), classification.Request{
        BusinessName: req.BusinessName,
        Description:  req.Description,
        Industry:     req.Industry,
        Location:     req.Location,
        Metadata:     req.Metadata,
    })
    if err != nil {
        h.logger.Error("Classification failed", "error", err, "business_name", req.BusinessName)
        http.Error(w, "Classification failed", http.StatusInternalServerError)
        return
    }
    
    // Build response
    response := ClassificationResponse{
        ID:           result.ID,
        BusinessName: req.BusinessName,
        Primary:      result.Primary,
        Alternatives: result.Alternatives,
        Confidence:   result.Confidence,
        Method:       result.Method,
        CreatedAt:    result.CreatedAt,
        ProcessingTime: time.Since(start).Milliseconds(),
    }
    
    // Send response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
    
    h.logger.Info("Classification completed", 
        "business_name", req.BusinessName,
        "confidence", result.Confidence,
        "processing_time_ms", response.ProcessingTime)
}
```

### Batch Processing Handler

```go
// BatchClassificationRequest represents a batch classification request
type BatchClassificationRequest struct {
    Businesses []ClassificationRequest `json:"businesses" validate:"required,min=1,max=1000"`
    Options    BatchOptions            `json:"options,omitempty"`
}

type BatchOptions struct {
    ParallelWorkers int  `json:"parallel_workers,omitempty"`
    ReturnAll       bool `json:"return_all,omitempty"`
    MinConfidence   float64 `json:"min_confidence,omitempty"`
}

// BatchClassificationResponse represents batch classification results
type BatchClassificationResponse struct {
    ID           string                    `json:"id"`
    Total        int                       `json:"total"`
    Successful   int                       `json:"successful"`
    Failed       int                       `json:"failed"`
    Results      []ClassificationResponse  `json:"results"`
    Errors       []BatchError              `json:"errors,omitempty"`
    CreatedAt    time.Time                 `json:"created_at"`
    ProcessingTime int64                   `json:"processing_time_ms"`
}

type BatchError struct {
    Index   int    `json:"index"`
    Message string `json:"message"`
}

// BatchClassify handles POST /v1/classify/batch requests
func (h *ClassificationHandler) BatchClassify(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Parse request
    var req BatchClassificationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("Failed to decode batch request", "error", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := h.validator.Struct(req); err != nil {
        h.logger.Error("Batch request validation failed", "error", err)
        http.Error(w, "Validation failed", http.StatusBadRequest)
        return
    }
    
    // Set default options
    if req.Options.ParallelWorkers == 0 {
        req.Options.ParallelWorkers = 10
    }
    if req.Options.MinConfidence == 0 {
        req.Options.MinConfidence = 0.5
    }
    
    // Convert to service requests
    serviceRequests := make([]classification.Request, len(req.Businesses))
    for i, business := range req.Businesses {
        serviceRequests[i] = classification.Request{
            BusinessName: business.BusinessName,
            Description:  business.Description,
            Industry:     business.Industry,
            Location:     business.Location,
            Metadata:     business.Metadata,
        }
    }
    
    // Perform batch classification
    results, err := h.service.BatchClassify(r.Context(), serviceRequests, classification.BatchOptions{
        ParallelWorkers: req.Options.ParallelWorkers,
        ReturnAll:       req.Options.ReturnAll,
        MinConfidence:   req.Options.MinConfidence,
    })
    if err != nil {
        h.logger.Error("Batch classification failed", "error", err)
        http.Error(w, "Batch classification failed", http.StatusInternalServerError)
        return
    }
    
    // Build response
    response := BatchClassificationResponse{
        ID:            results.ID,
        Total:         len(req.Businesses),
        Successful:    len(results.Results),
        Failed:        len(results.Errors),
        CreatedAt:     time.Now(),
        ProcessingTime: time.Since(start).Milliseconds(),
    }
    
    // Convert results
    for _, result := range results.Results {
        response.Results = append(response.Results, ClassificationResponse{
            ID:           result.ID,
            BusinessName: result.BusinessName,
            Primary:      result.Primary,
            Alternatives: result.Alternatives,
            Confidence:   result.Confidence,
            Method:       result.Method,
            CreatedAt:    result.CreatedAt,
        })
    }
    
    // Convert errors
    for _, err := range results.Errors {
        response.Errors = append(response.Errors, BatchError{
            Index:   err.Index,
            Message: err.Message,
        })
    }
    
    // Send response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
    
    h.logger.Info("Batch classification completed",
        "total", response.Total,
        "successful", response.Successful,
        "failed", response.Failed,
        "processing_time_ms", response.ProcessingTime)
}
```

## Service Layer Examples

### Classification Service Implementation

```go
package classification

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/observability"
    "github.com/pcraw4d/business-verification/internal/database"
)

// Service provides business classification functionality
type Service struct {
    classifier    Classifier
    repository    Repository
    cache         Cache
    logger        *observability.Logger
    metrics       *observability.Metrics
}

// NewService creates a new classification service
func NewService(classifier Classifier, repository Repository, cache Cache, logger *observability.Logger, metrics *observability.Metrics) *Service {
    return &Service{
        classifier: classifier,
        repository: repository,
        cache:      cache,
        logger:     logger,
        metrics:    metrics,
    }
}

// Classify performs business classification
func (s *Service) Classify(ctx context.Context, req Request) (*Result, error) {
    start := time.Now()
    
    // Check cache first
    cacheKey := s.generateCacheKey(req)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
        s.metrics.IncCounter("classification_cache_hits")
        return cached, nil
    }
    
    // Perform classification
    result, err := s.classifier.Classify(ctx, req)
    if err != nil {
        s.metrics.IncCounter("classification_errors")
        s.logger.Error("Classification failed", "error", err, "business_name", req.BusinessName)
        return nil, fmt.Errorf("classification failed: %w", err)
    }
    
    // Store in cache
    if err := s.cache.Set(ctx, cacheKey, result, time.Hour); err != nil {
        s.logger.Warn("Failed to cache classification result", "error", err)
    }
    
    // Store in database
    if err := s.repository.StoreClassification(ctx, result); err != nil {
        s.logger.Warn("Failed to store classification", "error", err)
    }
    
    // Record metrics
    s.metrics.IncCounter("classification_requests")
    s.metrics.RecordHistogram("classification_duration_ms", time.Since(start).Milliseconds())
    s.metrics.RecordHistogram("classification_confidence", result.Confidence)
    
    return result, nil
}

// BatchClassify performs batch classification
func (s *Service) BatchClassify(ctx context.Context, requests []Request, options BatchOptions) (*BatchResult, error) {
    start := time.Now()
    
    if len(requests) == 0 {
        return &BatchResult{ID: generateID()}, nil
    }
    
    // Limit batch size
    if len(requests) > 1000 {
        return nil, fmt.Errorf("batch size too large: %d (max 1000)", len(requests))
    }
    
    // Set default options
    if options.ParallelWorkers == 0 {
        options.ParallelWorkers = 10
    }
    if options.MinConfidence == 0 {
        options.MinConfidence = 0.5
    }
    
    // Create result channels
    resultChan := make(chan *Result, len(requests))
    errorChan := make(chan *BatchError, len(requests))
    
    // Create worker pool
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, options.ParallelWorkers)
    
    // Submit work
    for i, req := range requests {
        wg.Add(1)
        go func(index int, request Request) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore
            
            result, err := s.Classify(ctx, request)
            if err != nil {
                errorChan <- &BatchError{
                    Index:   index,
                    Message: err.Error(),
                }
            } else {
                resultChan <- result
            }
        }(i, req)
    }
    
    // Wait for all workers to complete
    go func() {
        wg.Wait()
        close(resultChan)
        close(errorChan)
    }()
    
    // Collect results
    var results []*Result
    var errors []*BatchError
    
    for result := range resultChan {
        if result.Confidence >= options.MinConfidence || options.ReturnAll {
            results = append(results, result)
        }
    }
    
    for err := range errorChan {
        errors = append(errors, err)
    }
    
    batchResult := &BatchResult{
        ID:        generateID(),
        Results:   results,
        Errors:    errors,
        CreatedAt: time.Now(),
    }
    
    // Record metrics
    s.metrics.IncCounter("batch_classification_requests")
    s.metrics.RecordHistogram("batch_classification_duration_ms", time.Since(start).Milliseconds())
    s.metrics.RecordHistogram("batch_classification_size", len(requests))
    
    return batchResult, nil
}

// generateCacheKey creates a cache key for the request
func (s *Service) generateCacheKey(req Request) string {
    // Implementation depends on cache strategy
    return fmt.Sprintf("classification:%s:%s:%s", 
        req.BusinessName, req.Industry, req.Location)
}
```

## Database Operations

### Repository Pattern Implementation

```go
package database

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/classification"
    "github.com/pcraw4d/business-verification/internal/observability"
)

// Repository provides data access for classifications
type Repository struct {
    db      *sql.DB
    logger  *observability.Logger
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB, logger *observability.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}

// StoreClassification stores a classification result
func (r *Repository) StoreClassification(ctx context.Context, result *classification.Result) error {
    query := `
        INSERT INTO classifications (
            id, business_name, primary_code, primary_description,
            confidence, method, created_at, metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        result.ID,
        result.BusinessName,
        result.Primary.Code,
        result.Primary.Description,
        result.Confidence,
        result.Method,
        result.CreatedAt,
        result.Metadata,
    )
    
    if err != nil {
        r.logger.Error("Failed to store classification", "error", err, "id", result.ID)
        return err
    }
    
    // Store alternatives
    for _, alt := range result.Alternatives {
        if err := r.storeAlternative(ctx, result.ID, alt); err != nil {
            r.logger.Error("Failed to store alternative", "error", err, "id", result.ID)
            return err
        }
    }
    
    return nil
}

// GetClassification retrieves a classification by ID
func (r *Repository) GetClassification(ctx context.Context, id string) (*classification.Result, error) {
    query := `
        SELECT id, business_name, primary_code, primary_description,
               confidence, method, created_at, metadata
        FROM classifications
        WHERE id = ?
    `
    
    var result classification.Result
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &result.ID,
        &result.BusinessName,
        &result.Primary.Code,
        &result.Primary.Description,
        &result.Confidence,
        &result.Method,
        &result.CreatedAt,
        &result.Metadata,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, classification.ErrNotFound
        }
        r.logger.Error("Failed to get classification", "error", err, "id", id)
        return nil, err
    }
    
    // Load alternatives
    alternatives, err := r.getAlternatives(ctx, id)
    if err != nil {
        r.logger.Error("Failed to get alternatives", "error", err, "id", id)
        return nil, err
    }
    result.Alternatives = alternatives
    
    return &result, nil
}

// GetClassificationsByBusinessName retrieves classifications by business name
func (r *Repository) GetClassificationsByBusinessName(ctx context.Context, businessName string, limit int) ([]*classification.Result, error) {
    query := `
        SELECT id, business_name, primary_code, primary_description,
               confidence, method, created_at, metadata
        FROM classifications
        WHERE business_name = ?
        ORDER BY created_at DESC
        LIMIT ?
    `
    
    rows, err := r.db.QueryContext(ctx, query, businessName, limit)
    if err != nil {
        r.logger.Error("Failed to query classifications", "error", err, "business_name", businessName)
        return nil, err
    }
    defer rows.Close()
    
    var results []*classification.Result
    for rows.Next() {
        var result classification.Result
        err := rows.Scan(
            &result.ID,
            &result.BusinessName,
            &result.Primary.Code,
            &result.Primary.Description,
            &result.Confidence,
            &result.Method,
            &result.CreatedAt,
            &result.Metadata,
        )
        if err != nil {
            r.logger.Error("Failed to scan classification", "error", err)
            return nil, err
        }
        
        // Load alternatives
        alternatives, err := r.getAlternatives(ctx, result.ID)
        if err != nil {
            r.logger.Error("Failed to get alternatives", "error", err, "id", result.ID)
            return nil, err
        }
        result.Alternatives = alternatives
        
        results = append(results, &result)
    }
    
    return results, nil
}

// storeAlternative stores an alternative classification result
func (r *Repository) storeAlternative(ctx context.Context, classificationID string, alt classification.Result) error {
    query := `
        INSERT INTO classification_alternatives (
            classification_id, code, description, confidence
        ) VALUES (?, ?, ?, ?)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        classificationID,
        alt.Code,
        alt.Description,
        alt.Confidence,
    )
    
    return err
}

// getAlternatives retrieves alternatives for a classification
func (r *Repository) getAlternatives(ctx context.Context, classificationID string) ([]classification.Result, error) {
    query := `
        SELECT code, description, confidence
        FROM classification_alternatives
        WHERE classification_id = ?
        ORDER BY confidence DESC
    `
    
    rows, err := r.db.QueryContext(ctx, query, classificationID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var alternatives []classification.Result
    for rows.Next() {
        var alt classification.Result
        err := rows.Scan(&alt.Code, &alt.Description, &alt.Confidence)
        if err != nil {
            return nil, err
        }
        alternatives = append(alternatives, alt)
    }
    
    return alternatives, nil
}
```

## Middleware Examples

### Authentication Middleware

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/pcraw4d/business-verification/internal/auth"
    "github.com/pcraw4d/business-verification/internal/observability"
)

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
    authService *auth.Service
    logger      *observability.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *auth.Service, logger *observability.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        authService: authService,
        logger:      logger,
    }
}

// Authenticate validates JWT tokens and sets user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            m.logger.Warn("Missing authorization header", "path", r.URL.Path)
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        // Parse Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            m.logger.Warn("Invalid authorization header format", "path", r.URL.Path)
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }
        
        token := parts[1]
        
        // Validate token
        claims, err := m.authService.ValidateToken(r.Context(), token)
        if err != nil {
            m.logger.Warn("Invalid token", "error", err, "path", r.URL.Path)
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Set user context
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// RequireRole middleware ensures the user has the required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user")
            if user == nil {
                m.logger.Warn("No user in context", "path", r.URL.Path)
                http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
            }
            
            claims, ok := user.(*auth.Claims)
            if !ok {
                m.logger.Error("Invalid user context type", "path", r.URL.Path)
                http.Error(w, "Internal server error", http.StatusInternalServerError)
                return
            }
            
            // Check if user has required role
            if !m.authService.HasRole(claims, role) {
                m.logger.Warn("Insufficient permissions", 
                    "user_id", claims.UserID, 
                    "required_role", role, 
                    "path", r.URL.Path)
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Rate Limiting Middleware

```go
// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
    limiter  *rate.Limiter
    logger   *observability.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(requestsPerSecond int, logger *observability.Logger) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
        logger:  logger,
    }
}

// RateLimit applies rate limiting to requests
func (m *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get client identifier (IP address or user ID)
        clientID := m.getClientID(r)
        
        // Check rate limit
        if !m.limiter.Allow() {
            m.logger.Warn("Rate limit exceeded", "client_id", clientID, "path", r.URL.Path)
            
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// getClientID extracts client identifier from request
func (m *RateLimitMiddleware) getClientID(r *http.Request) string {
    // Try to get user ID from context first
    if user := r.Context().Value("user"); user != nil {
        if claims, ok := user.(*auth.Claims); ok {
            return claims.UserID
        }
    }
    
    // Fall back to IP address
    return r.RemoteAddr
}
```

## Testing Examples

### Unit Test Example

```go
package handlers_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/api/handlers"
    "github.com/pcraw4d/business-verification/internal/classification"
    "github.com/pcraw4d/business-verification/internal/observability"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockClassificationService is a mock implementation of the classification service
type MockClassificationService struct {
    mock.Mock
}

func (m *MockClassificationService) Classify(ctx context.Context, req classification.Request) (*classification.Result, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*classification.Result), args.Error(1)
}

func (m *MockClassificationService) BatchClassify(ctx context.Context, requests []classification.Request, options classification.BatchOptions) (*classification.BatchResult, error) {
    args := m.Called(ctx, requests, options)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*classification.BatchResult), args.Error(1)
}

func TestClassificationHandler_Classify(t *testing.T) {
    // Setup
    mockService := new(MockClassificationService)
    logger := observability.NewLogger("test")
    handler := handlers.NewClassificationHandler(mockService, logger)
    
    // Test cases
    tests := []struct {
        name           string
        requestBody    handlers.ClassificationRequest
        mockResult     *classification.Result
        mockError      error
        expectedStatus int
        expectedBody   map[string]interface{}
    }{
        {
            name: "successful classification",
            requestBody: handlers.ClassificationRequest{
                BusinessName: "Test Business Inc",
                Description:  "Software development company",
            },
            mockResult: &classification.Result{
                ID:           "test-id",
                BusinessName: "Test Business Inc",
                Primary: classification.Result{
                    Code:        "541511",
                    Description: "Custom Computer Programming Services",
                },
                Confidence:   0.95,
                Method:       "hybrid",
                CreatedAt:    time.Now(),
            },
            mockError:      nil,
            expectedStatus: http.StatusOK,
            expectedBody: map[string]interface{}{
                "business_name": "Test Business Inc",
                "confidence":    0.95,
                "method":        "hybrid",
            },
        },
        {
            name: "validation error",
            requestBody: handlers.ClassificationRequest{
                BusinessName: "", // Invalid: empty business name
            },
            mockResult:     nil,
            mockError:      nil,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   nil,
        },
        {
            name: "service error",
            requestBody: handlers.ClassificationRequest{
                BusinessName: "Test Business Inc",
            },
            mockResult:     nil,
            mockError:      assert.AnError,
            expectedStatus: http.StatusInternalServerError,
            expectedBody:   nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock expectations
            if tt.mockResult != nil {
                mockService.On("Classify", mock.Anything, mock.Anything).Return(tt.mockResult, tt.mockError)
            }
            
            // Create request
            body, _ := json.Marshal(tt.requestBody)
            req := httptest.NewRequest("POST", "/v1/classify", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            
            // Create response recorder
            w := httptest.NewRecorder()
            
            // Execute handler
            handler.Classify(w, req)
            
            // Assertions
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedBody != nil {
                var response map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &response)
                assert.NoError(t, err)
                
                for key, expectedValue := range tt.expectedBody {
                    assert.Equal(t, expectedValue, response[key])
                }
            }
            
            // Verify mock expectations
            mockService.AssertExpectations(t)
        })
    }
}

func TestClassificationHandler_BatchClassify(t *testing.T) {
    // Setup
    mockService := new(MockClassificationService)
    logger := observability.NewLogger("test")
    handler := handlers.NewClassificationHandler(mockService, logger)
    
    // Test data
    requestBody := handlers.BatchClassificationRequest{
        Businesses: []handlers.ClassificationRequest{
            {BusinessName: "Business 1"},
            {BusinessName: "Business 2"},
        },
        Options: handlers.BatchOptions{
            ParallelWorkers: 5,
            MinConfidence:   0.8,
        },
    }
    
    mockResult := &classification.BatchResult{
        ID:        "batch-id",
        Results:   []*classification.Result{},
        Errors:    []*classification.BatchError{},
        CreatedAt: time.Now(),
    }
    
    // Setup mock expectations
    mockService.On("BatchClassify", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)
    
    // Create request
    body, _ := json.Marshal(requestBody)
    req := httptest.NewRequest("POST", "/v1/classify/batch", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Create response recorder
    w := httptest.NewRecorder()
    
    // Execute handler
    handler.BatchClassify(w, req)
    
    // Assertions
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response handlers.BatchClassificationResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    
    assert.Equal(t, "batch-id", response.ID)
    assert.Equal(t, 2, response.Total)
    assert.Equal(t, 0, response.Successful)
    assert.Equal(t, 0, response.Failed)
    
    // Verify mock expectations
    mockService.AssertExpectations(t)
}
```

## Configuration Examples

### Environment Configuration

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
    
    "github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
    Logging  LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
    Port         int
    Host         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
    SSLMode  string
    MaxConns int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
    JWTSecret     string
    JWTExpiration time.Duration
    RefreshSecret string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
    Level  string
    Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("failed to load .env file: %w", err)
    }
    
    config := &Config{}
    
    // Server configuration
    config.Server.Port = getEnvInt("SERVER_PORT", 8080)
    config.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
    config.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second)
    config.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)
    config.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)
    
    // Database configuration
    config.Database.Host = getEnv("DB_HOST", "localhost")
    config.Database.Port = getEnvInt("DB_PORT", 5432)
    config.Database.User = getEnv("DB_USER", "postgres")
    config.Database.Password = getEnv("DB_PASSWORD", "")
    config.Database.Name = getEnv("DB_NAME", "kyb_platform")
    config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
    config.Database.MaxConns = getEnvInt("DB_MAX_CONNS", 10)
    
    // Redis configuration
    config.Redis.Host = getEnv("REDIS_HOST", "localhost")
    config.Redis.Port = getEnvInt("REDIS_PORT", 6379)
    config.Redis.Password = getEnv("REDIS_PASSWORD", "")
    config.Redis.DB = getEnvInt("REDIS_DB", 0)
    
    // Auth configuration
    config.Auth.JWTSecret = getEnv("JWT_SECRET", "")
    config.Auth.JWTExpiration = getEnvDuration("JWT_EXPIRATION", 24*time.Hour)
    config.Auth.RefreshSecret = getEnv("REFRESH_SECRET", "")
    
    // Logging configuration
    config.Logging.Level = getEnv("LOG_LEVEL", "info")
    config.Logging.Format = getEnv("LOG_FORMAT", "json")
    
    // Validate required configuration
    if err := config.validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return config, nil
}

// validate validates the configuration
func (c *Config) validate() error {
    if c.Auth.JWTSecret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    
    if c.Auth.RefreshSecret == "" {
        return fmt.Errorf("REFRESH_SECRET is required")
    }
    
    if c.Database.Password == "" {
        return fmt.Errorf("DB_PASSWORD is required")
    }
    
    return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

// getEnvDuration gets a duration environment variable with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

## Error Handling Patterns

### Custom Error Types

```go
package errors

import (
    "fmt"
    "net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
    ErrorTypeNotFound     ErrorType = "NOT_FOUND"
    ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
    ErrorTypeForbidden    ErrorType = "FORBIDDEN"
    ErrorTypeConflict     ErrorType = "CONFLICT"
    ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
    ErrorTypeExternal     ErrorType = "EXTERNAL_ERROR"
)

// AppError represents an application error
type AppError struct {
    Type       ErrorType `json:"type"`
    Message    string    `json:"message"`
    Code       string    `json:"code,omitempty"`
    Details    string    `json:"details,omitempty"`
    HTTPStatus int       `json:"-"`
    Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
    return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details string) *AppError {
    return &AppError{
        Type:       ErrorTypeValidation,
        Message:    message,
        Details:    details,
        HTTPStatus: http.StatusBadRequest,
    }
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Type:       ErrorTypeNotFound,
        Message:    fmt.Sprintf("%s not found", resource),
        HTTPStatus: http.StatusNotFound,
    }
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Type:       ErrorTypeUnauthorized,
        Message:    message,
        HTTPStatus: http.StatusUnauthorized,
    }
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *AppError {
    return &AppError{
        Type:       ErrorTypeForbidden,
        Message:    message,
        HTTPStatus: http.StatusForbidden,
    }
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
    return &AppError{
        Type:       ErrorTypeConflict,
        Message:    message,
        HTTPStatus: http.StatusConflict,
    }
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *AppError {
    return &AppError{
        Type:       ErrorTypeInternal,
        Message:    message,
        HTTPStatus: http.StatusInternalServerError,
        Cause:      cause,
    }
}

// NewExternalError creates a new external service error
func NewExternalError(service string, message string, cause error) *AppError {
    return &AppError{
        Type:       ErrorTypeExternal,
        Message:    fmt.Sprintf("%s service error: %s", service, message),
        HTTPStatus: http.StatusBadGateway,
        Cause:      cause,
    }
}

// ErrorResponse represents an error response
type ErrorResponse struct {
    Error   *AppError `json:"error"`
    RequestID string  `json:"request_id,omitempty"`
    Timestamp string  `json:"timestamp"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *AppError, requestID string) *ErrorResponse {
    return &ErrorResponse{
        Error:     err,
        RequestID: requestID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}
```

## Common Utilities

### HTTP Response Utilities

```go
package utils

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/pcraw4d/business-verification/internal/errors"
)

// Response represents a standard API response
type Response struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *errors.AppError `json:"error,omitempty"`
    RequestID string      `json:"request_id,omitempty"`
    Timestamp string      `json:"timestamp"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}, requestID string) *Response {
    return &Response{
        Success:   true,
        Data:      data,
        RequestID: requestID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *errors.AppError, requestID string) *Response {
    return &Response{
        Success:   false,
        Error:     err,
        RequestID: requestID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    if data != nil {
        json.NewEncoder(w).Encode(data)
    }
}

// WriteSuccessResponse writes a success response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}, requestID string) {
    response := NewSuccessResponse(data, requestID)
    WriteJSON(w, http.StatusOK, response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, err *errors.AppError, requestID string) {
    response := NewErrorResponse(err, requestID)
    WriteJSON(w, err.HTTPStatus, response)
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, message string, details string, requestID string) {
    err := errors.NewValidationError(message, details)
    WriteErrorResponse(w, err, requestID)
}

// WriteNotFoundError writes a not found error response
func WriteNotFoundError(w http.ResponseWriter, resource string, requestID string) {
    err := errors.NewNotFoundError(resource)
    WriteErrorResponse(w, err, requestID)
}

// WriteUnauthorizedError writes an unauthorized error response
func WriteUnauthorizedError(w http.ResponseWriter, message string, requestID string) {
    err := errors.NewUnauthorizedError(message)
    WriteErrorResponse(w, err, requestID)
}

// WriteForbiddenError writes a forbidden error response
func WriteForbiddenError(w http.ResponseWriter, message string, requestID string) {
    err := errors.NewForbiddenError(message)
    WriteErrorResponse(w, err, requestID)
}

// WriteInternalError writes an internal error response
func WriteInternalError(w http.ResponseWriter, message string, cause error, requestID string) {
    err := errors.NewInternalError(message, cause)
    WriteErrorResponse(w, err, requestID)
}
```

---

*Last updated: January 2024*
