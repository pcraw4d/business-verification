# External API Integration Documentation

## Overview

The Risk Assessment Service integrates with premium external data providers to enhance risk analysis capabilities. This document provides comprehensive information about the external API integrations, their mock implementations, and go-live procedures.

## Supported External APIs

### 1. Thomson Reuters

**Purpose**: Financial data, company profiles, and ESG scoring
**Mock Implementation**: `internal/external/thomson_reuters/thomson_reuters_mock.go`

#### Features
- Company profile information
- Financial data and ratios
- Risk metrics and scoring
- ESG (Environmental, Social, Governance) scores
- Executive information
- Ownership structure analysis

#### API Endpoints (Mock)
```go
// Get company profile
GetCompanyProfile(ctx context.Context, businessName, country string) (*CompanyProfile, error)

// Get financial data
GetFinancialData(ctx context.Context, companyID string) (*FinancialData, error)

// Get financial ratios
GetFinancialRatios(ctx context.Context, companyID string) (*FinancialRatios, error)

// Get risk metrics
GetRiskMetrics(ctx context.Context, companyID string) (*RiskMetrics, error)

// Get ESG score
GetESGScore(ctx context.Context, companyID string) (*ESGScore, error)

// Get executive information
GetExecutiveInfo(ctx context.Context, companyID string) (*ExecutiveInfo, error)

// Get ownership structure
GetOwnershipStructure(ctx context.Context, companyID string) (*OwnershipStructure, error)

// Get comprehensive data (all endpoints)
GetComprehensiveData(ctx context.Context, businessName, country string) (*ThomsonReutersResult, error)
```

#### Risk Factors Generated
- **Revenue Growth Risk**: Based on revenue size and growth patterns
- **Profitability Risk**: Based on net income and financial health
- **Overall Risk Score**: Comprehensive risk assessment from Thomson Reuters
- **ESG Risk**: Environmental, Social, and Governance risk factors

### 2. OFAC (Office of Foreign Assets Control)

**Purpose**: Sanctions screening and compliance verification
**Mock Implementation**: `internal/external/ofac/ofac_mock.go`

#### Features
- Sanctions list searches
- Entity verification
- Compliance status checking
- PEP (Politically Exposed Person) screening

#### API Endpoints (Mock)
```go
// Search sanctions lists
SearchSanctions(ctx context.Context, entityName, entityType string) (*SanctionsSearch, error)

// Verify entity compliance
VerifyEntity(ctx context.Context, entityName, entityType string) (*EntityVerification, error)

// Get compliance status
GetComplianceStatus(ctx context.Context, entityName string) (*ComplianceStatus, error)

// Get comprehensive data
GetComprehensiveData(ctx context.Context, entityName string) (*OFACResult, error)
```

#### Risk Factors Generated
- **Sanctions Risk**: Based on sanctions list matches
- **Compliance Risk**: Overall compliance status assessment
- **PEP Risk**: Politically Exposed Person risk factors

### 3. World-Check

**Purpose**: Enhanced due diligence and adverse media screening
**Mock Implementation**: `internal/external/worldcheck/worldcheck_mock.go`

#### Features
- Entity profiling
- Adverse media monitoring
- PEP status verification
- Sanctions information
- Risk assessment scoring

#### API Endpoints (Mock)
```go
// Search entity profile
SearchProfile(ctx context.Context, entityName string) (*Profile, error)

// Get adverse media
GetAdverseMedia(ctx context.Context, entityName string) (*AdverseMedia, error)

// Get PEP status
GetPEPStatus(ctx context.Context, entityName string) (*PEPStatus, error)

// Get sanctions information
GetSanctionsInfo(ctx context.Context, entityName string) (*SanctionsInfo, error)

// Get risk assessment
GetRiskAssessment(ctx context.Context, entityName string) (*RiskAssessment, error)

// Get comprehensive data
GetComprehensiveData(ctx context.Context, entityName string) (*WorldCheckResult, error)
```

#### Risk Factors Generated
- **Adverse Media Risk**: Based on negative media coverage
- **PEP Risk**: Politically Exposed Person risk assessment
- **Sanctions Risk**: Sanctions-related risk factors
- **Overall Risk Score**: Comprehensive risk assessment from World-Check

## External API Manager

The `ExternalAPIManager` coordinates calls to all external APIs and provides a unified interface for risk assessment.

### Configuration

```go
type ExternalAPIManagerConfig struct {
    ThomsonReuters *ExternalDataConfig `yaml:"thomson_reuters"`
    OFAC          *ExternalDataConfig `yaml:"ofac"`
    WorldCheck    *ExternalDataConfig `yaml:"worldcheck"`
}

type ExternalDataConfig struct {
    APIKey    string        `yaml:"api_key"`
    BaseURL   string        `yaml:"base_url"`
    Timeout   time.Duration `yaml:"timeout"`
    RateLimit int           `yaml:"rate_limit"`
    Enabled   bool          `yaml:"enabled"`
}
```

### Usage

```go
// Initialize the manager
config := &ExternalAPIManagerConfig{
    ThomsonReuters: &ExternalDataConfig{
        APIKey:    "your-api-key",
        BaseURL:   "https://api.thomsonreuters.com",
        Timeout:   30 * time.Second,
        RateLimit: 100,
        Enabled:   true,
    },
    // ... other API configs
}

manager := NewExternalAPIManager(config, logger)

// Get comprehensive data from all APIs
result, err := manager.GetComprehensiveData(ctx, "Acme Corp", "US")
```

## Go-Live Procedures

### 1. API Key Management

#### Production API Keys
- Store API keys in secure environment variables or secret management systems
- Use different keys for different environments (dev, staging, production)
- Implement key rotation policies

#### Environment Variables
```bash
# Thomson Reuters
THOMSON_REUTERS_API_KEY=your_production_key
THOMSON_REUTERS_BASE_URL=https://api.thomsonreuters.com
THOMSON_REUTERS_TIMEOUT=30s
THOMSON_REUTERS_RATE_LIMIT=100

# OFAC
OFAC_API_KEY=your_production_key
OFAC_BASE_URL=https://api.ofac.treasury.gov
OFAC_TIMEOUT=30s
OFAC_RATE_LIMIT=50

# World-Check
WORLDCHECK_API_KEY=your_production_key
WORLDCHECK_BASE_URL=https://api.worldcheck.com
WORLDCHECK_TIMEOUT=30s
WORLDCHECK_RATE_LIMIT=75
```

### 2. Rate Limiting and Timeouts

#### Recommended Settings
- **Thomson Reuters**: 100 requests/minute, 30s timeout
- **OFAC**: 50 requests/minute, 30s timeout  
- **World-Check**: 75 requests/minute, 30s timeout

#### Implementation
```go
// Rate limiting is handled by the mock implementations
// In production, implement proper rate limiting middleware
func (api *ExternalAPI) makeRequest(ctx context.Context, endpoint string) error {
    // Check rate limit
    if !api.rateLimiter.Allow() {
        return errors.New("rate limit exceeded")
    }
    
    // Make request with timeout
    ctx, cancel := context.WithTimeout(ctx, api.timeout)
    defer cancel()
    
    // ... make HTTP request
}
```

### 3. Error Handling and Retry Logic

#### Retry Strategy
```go
func (api *ExternalAPI) makeRequestWithRetry(ctx context.Context, endpoint string) (*Response, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        response, err := api.makeRequest(ctx, endpoint)
        if err == nil {
            return response, nil
        }
        
        // Exponential backoff
        delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
        time.Sleep(delay)
    }
    
    return nil, errors.New("max retries exceeded")
}
```

#### Error Types
- **Rate Limit Exceeded**: 429 status code
- **Authentication Failed**: 401 status code
- **Service Unavailable**: 503 status code
- **Timeout**: Context deadline exceeded

### 4. Data Quality and Validation

#### Data Quality Metrics
```go
type DataQuality struct {
    Completeness float64 `json:"completeness"` // 0.0 - 1.0
    Accuracy     float64 `json:"accuracy"`     // 0.0 - 1.0
    Timeliness   float64 `json:"timeliness"`   // 0.0 - 1.0
    Consistency  float64 `json:"consistency"`  // 0.0 - 1.0
    Overall      float64 `json:"overall"`      // 0.0 - 1.0
}
```

#### Validation Rules
- Verify required fields are present
- Check data format and types
- Validate against known patterns
- Cross-reference with other data sources

### 5. Monitoring and Alerting

#### Key Metrics to Monitor
- API response times
- Error rates by API
- Rate limit utilization
- Data quality scores
- Request success rates

#### Alerting Thresholds
- Response time > 5 seconds
- Error rate > 5%
- Rate limit utilization > 80%
- Data quality score < 0.8

### 6. Testing Strategy

#### Unit Tests
- Test individual API mock implementations
- Test error handling scenarios
- Test rate limiting behavior
- Test data validation

#### Integration Tests
- Test API manager coordination
- Test comprehensive data retrieval
- Test concurrent API calls
- Test timeout scenarios

#### Load Tests
- Test under high request volumes
- Test rate limit handling
- Test memory usage
- Test response time under load

## Security Considerations

### 1. API Key Security
- Never log API keys
- Use secure storage (Azure Key Vault, AWS Secrets Manager)
- Implement key rotation
- Monitor key usage

### 2. Data Privacy
- Encrypt sensitive data in transit and at rest
- Implement data retention policies
- Comply with GDPR and other regulations
- Audit data access

### 3. Network Security
- Use HTTPS for all API communications
- Implement certificate pinning
- Use VPN or private networks where possible
- Monitor network traffic

## Performance Optimization

### 1. Caching Strategy
```go
// Cache API responses to reduce external calls
type APICache struct {
    cache map[string]*CacheEntry
    ttl   time.Duration
}

func (c *APICache) Get(key string) (*Response, bool) {
    entry, exists := c.cache[key]
    if !exists || time.Since(entry.Timestamp) > c.ttl {
        return nil, false
    }
    return entry.Response, true
}
```

### 2. Concurrent Processing
```go
// Process multiple API calls concurrently
func (manager *ExternalAPIManager) GetComprehensiveData(ctx context.Context, businessName, country string) (*ComprehensiveResult, error) {
    results := make(chan result, 3)
    
    // Start all API calls concurrently
    go func() {
        tr, err := manager.thomsonReuters.GetComprehensiveData(ctx, businessName, country)
        results <- result{source: "thomson_reuters", data: tr, err: err}
    }()
    
    go func() {
        ofac, err := manager.ofac.GetComprehensiveData(ctx, businessName)
        results <- result{source: "ofac", data: ofac, err: err}
    }()
    
    go func() {
        wc, err := manager.worldCheck.GetComprehensiveData(ctx, businessName)
        results <- result{source: "worldcheck", data: wc, err: err}
    }()
    
    // Collect results
    // ... process results
}
```

### 3. Connection Pooling
```go
// Use HTTP client with connection pooling
func NewHTTPClient() *http.Client {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Rate Limit Exceeded
**Symptoms**: 429 status codes, requests failing
**Solutions**:
- Implement exponential backoff
- Increase rate limit if possible
- Cache responses to reduce API calls
- Distribute requests across multiple API keys

#### 2. Authentication Failures
**Symptoms**: 401 status codes, invalid API key errors
**Solutions**:
- Verify API key is correct and active
- Check key permissions and scope
- Implement key rotation
- Monitor key usage and limits

#### 3. Timeout Issues
**Symptoms**: Context deadline exceeded, slow responses
**Solutions**:
- Increase timeout values
- Optimize request payloads
- Implement request retries
- Use connection pooling

#### 4. Data Quality Issues
**Symptoms**: Incomplete or inaccurate data
**Solutions**:
- Implement data validation
- Cross-reference multiple sources
- Set up data quality monitoring
- Implement data cleansing

### Debugging Tools

#### Logging
```go
// Enable detailed logging for debugging
logger.Info("API request started",
    zap.String("api", "thomson_reuters"),
    zap.String("endpoint", endpoint),
    zap.String("business_name", businessName))

logger.Error("API request failed",
    zap.String("api", "thomson_reuters"),
    zap.Error(err),
    zap.Duration("duration", time.Since(startTime)))
```

#### Metrics
```go
// Track API performance metrics
type APIMetrics struct {
    RequestCount    int64
    ErrorCount      int64
    AverageLatency  time.Duration
    RateLimitHits   int64
}
```

## Migration from Mock to Production

### 1. Gradual Rollout
- Start with non-critical endpoints
- Use feature flags to control rollout
- Monitor performance and errors
- Roll back if issues occur

### 2. A/B Testing
- Compare mock vs production results
- Measure accuracy and performance
- Validate data quality
- Adjust configurations as needed

### 3. Fallback Strategy
- Keep mock implementations as fallback
- Implement circuit breakers
- Use cached data when APIs are down
- Alert on API failures

## Conclusion

The external API integration provides comprehensive risk assessment capabilities through premium data sources. The mock implementations allow for development and testing, while the go-live procedures ensure smooth production deployment with proper monitoring, security, and performance optimization.

For questions or issues, refer to the individual API documentation or contact the development team.
