# API Gateway Integration Guide
## Risk Assessment Service Integration

### Overview

This document provides comprehensive guidance on integrating the Risk Assessment Service with the existing API Gateway in the KYB Platform. It covers service discovery, routing configuration, authentication, monitoring, and troubleshooting.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Service Discovery](#service-discovery)
3. [Routing Configuration](#routing-configuration)
4. [Authentication Flow](#authentication-flow)
5. [Request/Response Handling](#requestresponse-handling)
6. [Error Handling](#error-handling)
7. [Monitoring and Observability](#monitoring-and-observability)
8. [Performance Optimization](#performance-optimization)
9. [Security Considerations](#security-considerations)
10. [Troubleshooting](#troubleshooting)
11. [Testing](#testing)

---

## Architecture Overview

### System Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   Client Apps   │    │   API Gateway    │    │ Risk Assessment     │
│                 │    │                  │    │ Service             │
│ • Web App       │───▶│ • Routing        │───▶│ • Risk Analysis     │
│ • Mobile App    │    │ • Auth           │    │ • ML Predictions    │
│ • Third-party   │    │ • Rate Limiting  │    │ • External APIs     │
│   Integrations  │    │ • Monitoring     │    │ • Data Storage      │
└─────────────────┘    └──────────────────┘    └─────────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   Backend        │
                       │   Services       │
                       │                  │
                       │ • Classification │
                       │ • Merchant       │
                       │ • BI Service     │
                       └──────────────────┘
```

### Integration Points

1. **API Gateway**: Routes requests to Risk Assessment Service
2. **Service Discovery**: Locates Risk Assessment Service instances
3. **Authentication**: Validates and forwards authentication tokens
4. **Rate Limiting**: Applies rate limits to Risk Assessment endpoints
5. **Monitoring**: Collects metrics and logs from Risk Assessment Service
6. **Error Handling**: Manages errors and fallbacks

---

## Service Discovery

### Static Configuration

The API Gateway uses static URL configuration for service discovery:

```go
// services/api-gateway/internal/config/config.go
type ServicesConfig struct {
    ClassificationURL     string
    MerchantURL           string
    FrontendURL           string
    BIServiceURL          string
    RiskAssessmentURL     string  // New field for Risk Assessment Service
}
```

### Environment Variables

```bash
# Production
RISK_ASSESSMENT_SERVICE_URL=https://risk-assessment-service-production.up.railway.app

# Staging
RISK_ASSESSMENT_SERVICE_URL=https://risk-assessment-service-staging.up.railway.app

# Development
RISK_ASSESSMENT_SERVICE_URL=https://risk-assessment-service-dev.up.railway.app
```

### Service Health Checks

The API Gateway performs health checks on the Risk Assessment Service:

```go
// Health check endpoint
GET /api/v1/risk/health

// Expected response
{
    "service": "risk-assessment-service",
    "version": "1.0.0",
    "status": "healthy",
    "timestamp": "2024-12-01T10:30:00Z",
    "uptime": "24h30m15s",
    "dependencies": {
        "database": "healthy",
        "redis": "healthy",
        "external_apis": "healthy"
    }
}
```

---

## Routing Configuration

### Route Definitions

The API Gateway routes Risk Assessment requests using the following patterns:

```go
// services/api-gateway/cmd/main.go

// Health check route
api.HandleFunc("/risk/health", gatewayHandler.ProxyToRiskAssessmentHealth).Methods("GET")

// Main risk assessment endpoint
api.HandleFunc("/risk/assess", func(w http.ResponseWriter, r *http.Request) {
    // Handle OPTIONS requests for CORS preflight
    if r.Method == "OPTIONS" {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
        w.WriteHeader(http.StatusOK)
        return
    }
    gatewayHandler.ProxyToRiskAssessment(w, r)
}).Methods("POST", "OPTIONS")

// All other risk assessment routes
api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)
```

### Route Patterns

| Pattern | Method | Description | Example |
|---------|--------|-------------|---------|
| `/api/v1/risk/health` | GET | Health check | `GET /api/v1/risk/health` |
| `/api/v1/risk/assess` | POST | Risk assessment | `POST /api/v1/risk/assess` |
| `/api/v1/risk/predict` | POST | Risk prediction | `POST /api/v1/risk/predict` |
| `/api/v1/risk/batch` | POST | Batch processing | `POST /api/v1/risk/batch` |
| `/api/v1/risk/webhooks` | POST | Webhook management | `POST /api/v1/risk/webhooks` |
| `/api/v1/risk/reports` | GET | Report generation | `GET /api/v1/risk/reports` |

### Path Transformation

The API Gateway transforms paths when forwarding requests:

```go
// Original request: /api/v1/risk/assess
// Forwarded to: https://risk-assessment-service-production.up.railway.app/assess

// Original request: /api/v1/risk/predict
// Forwarded to: https://risk-assessment-service-production.up.railway.app/predict
```

---

## Authentication Flow

### Authentication Methods

The API Gateway supports multiple authentication methods for Risk Assessment requests:

1. **JWT Tokens**: Standard JWT authentication
2. **API Keys**: Service-to-service authentication
3. **Service Tokens**: Internal service authentication

### Authentication Flow

```
┌─────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│   Client    │    │   API Gateway   │    │ Risk Assessment      │
│             │    │                 │    │ Service              │
└─────────────┘    └─────────────────┘    └──────────────────────┘
       │                     │                       │
       │ 1. Request + Token  │                       │
       ├────────────────────▶│                       │
       │                     │                       │
       │                     │ 2. Validate Token     │
       │                     ├──────────────────────▶│
       │                     │                       │
       │                     │ 3. Token Valid        │
       │                     │◀──────────────────────┤
       │                     │                       │
       │                     │ 4. Forward Request    │
       │                     ├──────────────────────▶│
       │                     │                       │
       │                     │ 5. Response           │
       │                     │◀──────────────────────┤
       │                     │                       │
       │ 6. Response         │                       │
       │◀────────────────────┤                       │
```

### Token Validation

```go
// API Gateway validates tokens before forwarding requests
func (h *GatewayHandler) ProxyToRiskAssessment(w http.ResponseWriter, r *http.Request) {
    // Extract and validate authentication token
    token := r.Header.Get("Authorization")
    if token == "" {
        http.Error(w, "Authorization header required", http.StatusUnauthorized)
        return
    }
    
    // Validate token with Risk Assessment Service
    if !h.validateToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    // Forward request with validated token
    h.proxyRequest(w, r, h.config.Services.RiskAssessmentURL, path)
}
```

### Service-to-Service Authentication

For internal service communication:

```go
// Add service token to requests
r.Header.Set("X-Service-Token", h.config.ServiceToken)
r.Header.Set("X-Service-ID", "api-gateway")
```

---

## Request/Response Handling

### Request Processing

The API Gateway processes Risk Assessment requests through the following steps:

1. **Request Validation**: Validate request format and required fields
2. **Authentication**: Verify authentication token
3. **Rate Limiting**: Apply rate limits based on client/token
4. **Path Transformation**: Transform API Gateway path to service path
5. **Request Forwarding**: Forward request to Risk Assessment Service
6. **Response Processing**: Process and return response

### Request Headers

The API Gateway adds/modifies the following headers:

```go
// Add correlation ID for tracing
correlationID := r.Header.Get("X-Request-ID")
if correlationID == "" {
    correlationID = fmt.Sprintf("req-%d", time.Now().UnixNano())
}
r.Header.Set("X-Request-ID", correlationID)

// Add service identification
r.Header.Set("X-Service-ID", "api-gateway")
r.Header.Set("X-Gateway-Version", "1.0.8")

// Preserve original headers
r.Header.Set("X-Original-Host", r.Host)
r.Header.Set("X-Original-Path", r.URL.Path)
```

### Response Headers

The API Gateway adds the following response headers:

```go
// Add CORS headers
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

// Add gateway information
w.Header().Set("X-Gateway-Service", "api-gateway")
w.Header().Set("X-Gateway-Version", "1.0.8")
w.Header().Set("X-Response-Time", fmt.Sprintf("%dms", responseTime))
```

### Request/Response Logging

```go
// Log request details
h.logger.Info("Risk Assessment request",
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.String("correlation_id", correlationID),
    zap.String("user_id", userID),
    zap.String("tenant_id", tenantID))

// Log response details
h.logger.Info("Risk Assessment response",
    zap.String("correlation_id", correlationID),
    zap.Int("status_code", response.StatusCode),
    zap.Duration("response_time", responseTime))
```

---

## Error Handling

### Error Categories

The API Gateway handles different types of errors:

1. **Authentication Errors**: Invalid or missing tokens
2. **Rate Limiting Errors**: Too many requests
3. **Service Errors**: Risk Assessment Service errors
4. **Network Errors**: Connection timeouts, DNS failures
5. **Validation Errors**: Invalid request format

### Error Response Format

```json
{
    "error": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Rate limit exceeded. Please try again later.",
        "details": {
            "limit": 1000,
            "remaining": 0,
            "reset_time": "2024-12-01T11:00:00Z"
        },
        "correlation_id": "req-1234567890",
        "timestamp": "2024-12-01T10:30:00Z"
    }
}
```

### Error Handling Implementation

```go
func (h *GatewayHandler) handleRiskAssessmentError(w http.ResponseWriter, r *http.Request, err error) {
    correlationID := r.Header.Get("X-Request-ID")
    
    switch {
    case errors.Is(err, ErrRateLimitExceeded):
        h.logger.Warn("Rate limit exceeded",
            zap.String("correlation_id", correlationID),
            zap.Error(err))
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        
    case errors.Is(err, ErrAuthenticationFailed):
        h.logger.Warn("Authentication failed",
            zap.String("correlation_id", correlationID),
            zap.Error(err))
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        
    case errors.Is(err, ErrServiceUnavailable):
        h.logger.Error("Risk Assessment Service unavailable",
            zap.String("correlation_id", correlationID),
            zap.Error(err))
        http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
        
    default:
        h.logger.Error("Unexpected error",
            zap.String("correlation_id", correlationID),
            zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}
```

### Circuit Breaker Pattern

```go
// Implement circuit breaker for Risk Assessment Service
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       CircuitState
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitBreakerOpen
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = Open
        }
        return err
    }
    
    cb.failures = 0
    cb.state = Closed
    return nil
}
```

---

## Monitoring and Observability

### Metrics Collection

The API Gateway collects the following metrics for Risk Assessment requests:

```go
// Request metrics
var (
    riskAssessmentRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "risk_assessment_requests_total",
            Help: "Total number of risk assessment requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
    
    riskAssessmentRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "risk_assessment_request_duration_seconds",
            Help: "Duration of risk assessment requests",
        },
        []string{"method", "endpoint"},
    )
    
    riskAssessmentActiveConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "risk_assessment_active_connections",
            Help: "Number of active connections to risk assessment service",
        },
    )
)
```

### Logging

Structured logging for Risk Assessment requests:

```go
// Request logging
h.logger.Info("Risk Assessment request started",
    zap.String("correlation_id", correlationID),
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.String("user_id", userID),
    zap.String("tenant_id", tenantID),
    zap.String("remote_addr", r.RemoteAddr))

// Response logging
h.logger.Info("Risk Assessment request completed",
    zap.String("correlation_id", correlationID),
    zap.Int("status_code", response.StatusCode),
    zap.Duration("duration", responseTime),
    zap.Int64("response_size", responseSize))
```

### Health Checks

```go
// Health check for Risk Assessment Service integration
func (h *GatewayHandler) CheckRiskAssessmentHealth() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    req, err := http.NewRequestWithContext(ctx, "GET", 
        h.config.Services.RiskAssessmentURL+"/health", nil)
    if err != nil {
        return err
    }
    
    resp, err := h.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed with status %d", resp.StatusCode)
    }
    
    return nil
}
```

---

## Performance Optimization

### Connection Pooling

```go
// Configure HTTP client with connection pooling
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableKeepAlives:   false,
    },
    Timeout: 30 * time.Second,
}
```

### Request Timeout Configuration

```go
// Set appropriate timeouts for different operations
var timeouts = map[string]time.Duration{
    "assess":    10 * time.Second,  // Risk assessment
    "predict":   15 * time.Second,  // Risk prediction
    "batch":     60 * time.Second,  // Batch processing
    "health":    5 * time.Second,   // Health check
}
```

### Caching Strategy

```go
// Cache risk assessment results
type RiskAssessmentCache struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *RiskAssessmentCache) Get(key string) (*RiskAssessmentResult, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.cache[key]
    if !exists || time.Since(entry.Timestamp) > c.ttl {
        return nil, false
    }
    
    return entry.Result, true
}
```

---

## Security Considerations

### Input Validation

```go
// Validate risk assessment requests
func validateRiskAssessmentRequest(req *RiskAssessmentRequest) error {
    if req.BusinessName == "" {
        return errors.New("business name is required")
    }
    
    if req.BusinessAddress == "" {
        return errors.New("business address is required")
    }
    
    if len(req.BusinessName) > 255 {
        return errors.New("business name too long")
    }
    
    return nil
}
```

### Rate Limiting

```go
// Apply rate limiting to Risk Assessment endpoints
func (h *GatewayHandler) applyRateLimit(w http.ResponseWriter, r *http.Request) bool {
    clientID := getClientID(r)
    
    if !h.rateLimiter.Allow(clientID) {
        w.Header().Set("X-RateLimit-Limit", "1000")
        w.Header().Set("X-RateLimit-Remaining", "0")
        w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Hour).Format(time.RFC3339))
        
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return false
    }
    
    return true
}
```

### CORS Configuration

```go
// Configure CORS for Risk Assessment endpoints
func (h *GatewayHandler) setupCORS(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    
    if h.isAllowedOrigin(origin) {
        w.Header().Set("Access-Control-Allow-Origin", origin)
    }
    
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
    w.Header().Set("Access-Control-Max-Age", "86400")
}
```

---

## Troubleshooting

### Common Issues

#### 1. Service Discovery Failures

**Problem**: API Gateway cannot locate Risk Assessment Service
**Symptoms**: 502 Bad Gateway errors
**Solutions**:
```bash
# Check service URL configuration
echo $RISK_ASSESSMENT_SERVICE_URL

# Test service connectivity
curl -f $RISK_ASSESSMENT_SERVICE_URL/health

# Check DNS resolution
nslookup risk-assessment-service-production.up.railway.app
```

#### 2. Authentication Failures

**Problem**: Requests fail with 401 Unauthorized
**Symptoms**: Authentication errors in logs
**Solutions**:
```bash
# Verify JWT token
jwt decode <token>

# Check token expiration
jwt decode <token> | jq '.exp'

# Validate service token
curl -H "X-Service-Token: <token>" $RISK_ASSESSMENT_SERVICE_URL/health
```

#### 3. Rate Limiting Issues

**Problem**: Requests fail with 429 Too Many Requests
**Symptoms**: Rate limit exceeded errors
**Solutions**:
```bash
# Check rate limit configuration
grep -r "rate_limit" configs/

# Monitor rate limit metrics
curl $PROMETHEUS_URL/api/v1/query?query=rate_limit_requests_total

# Adjust rate limits if needed
railway variables set RATE_LIMIT_REQUESTS_PER_MINUTE=2000
```

#### 4. Performance Issues

**Problem**: Slow response times
**Symptoms**: High latency metrics
**Solutions**:
```bash
# Check service performance
curl -w "@curl-format.txt" $RISK_ASSESSMENT_SERVICE_URL/health

# Monitor connection pool
curl $PROMETHEUS_URL/api/v1/query?query=risk_assessment_active_connections

# Check for bottlenecks
railway logs --service risk-assessment-service --tail 100
```

### Debug Commands

```bash
# Check API Gateway health
curl -f https://api-gateway.kyb-platform.com/health

# Test Risk Assessment endpoint
curl -X POST https://api-gateway.kyb-platform.com/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"business_name": "Test Company", "business_address": "123 Test St"}'

# Check service discovery
curl https://api-gateway.kyb-platform.com/api/v1/risk/health

# Monitor metrics
curl $PROMETHEUS_URL/api/v1/query?query=risk_assessment_requests_total
```

---

## Testing

### Unit Tests

```go
func TestRiskAssessmentProxy(t *testing.T) {
    // Setup test server
    testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "success"}`))
    }))
    defer testServer.Close()
    
    // Create gateway handler
    handler := &GatewayHandler{
        config: &Config{
            Services: ServicesConfig{
                RiskAssessmentURL: testServer.URL,
            },
        },
    }
    
    // Test request
    req := httptest.NewRequest("POST", "/api/v1/risk/assess", nil)
    w := httptest.NewRecorder()
    
    handler.ProxyToRiskAssessment(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Tests

```go
func TestRiskAssessmentIntegration(t *testing.T) {
    // Test end-to-end flow
    client := &http.Client{}
    
    // 1. Test health check
    resp, err := client.Get("https://api-gateway.kyb-platform.com/api/v1/risk/health")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 2. Test risk assessment
    reqBody := `{"business_name": "Test Company", "business_address": "123 Test St"}`
    req, err := http.NewRequest("POST", 
        "https://api-gateway.kyb-platform.com/api/v1/risk/assess", 
        strings.NewReader(reqBody))
    assert.NoError(t, err)
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    
    resp, err = client.Do(req)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Load Tests

```bash
# Run load test against API Gateway
go run ./cmd/load_test.go \
  -url="https://api-gateway.kyb-platform.com" \
  -endpoint="/api/v1/risk/assess" \
  -duration=5m \
  -users=100 \
  -type=load
```

---

## Best Practices

### 1. Configuration Management
- Use environment variables for service URLs
- Implement configuration validation
- Use different configurations for each environment
- Document all configuration options

### 2. Error Handling
- Implement comprehensive error handling
- Use structured error responses
- Log errors with correlation IDs
- Implement circuit breaker patterns

### 3. Monitoring
- Collect comprehensive metrics
- Implement health checks
- Use structured logging
- Set up alerting for critical issues

### 4. Security
- Validate all inputs
- Implement proper authentication
- Use HTTPS for all communications
- Apply rate limiting

### 5. Performance
- Use connection pooling
- Implement caching where appropriate
- Set appropriate timeouts
- Monitor performance metrics

---

## Support and Maintenance

### Regular Tasks
- Monitor service health and performance
- Review and update rate limits
- Rotate authentication tokens
- Update service configurations

### Emergency Procedures
- Service failure: Check health endpoints and logs
- Performance issues: Monitor metrics and adjust limits
- Security incidents: Review logs and update configurations
- Data issues: Validate request/response formats

### Contact Information
- **DevOps Team**: devops@kyb-platform.com
- **Platform Team**: platform@kyb-platform.com
- **Emergency**: +1-XXX-XXX-XXXX

---

**Document Version**: 1.0.0  
**Last Updated**: December 2024  
**Next Review**: March 2025
