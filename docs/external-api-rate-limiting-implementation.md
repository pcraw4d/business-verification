# External API Rate Limiting Implementation

## Overview

The External API Rate Limiting module provides comprehensive rate limiting capabilities for external API calls within the Risk Assessment system. This implementation addresses **Task 4.8.1: Implement per-API rate limiting and quotas** and provides a foundation for the remaining rate limiting tasks.

### Key Features

- **Per-API Rate Limiting**: Individual rate limits for each external API endpoint
- **Global Rate Limiting**: System-wide rate limits to prevent overall API abuse
- **Multi-level Quotas**: Minute, hour, and daily request limits
- **Priority-based Processing**: Higher priority APIs get preferential treatment
- **Monitoring and Metrics**: Comprehensive tracking of rate limit usage
- **Fallback Strategies**: Automatic fallback to alternative APIs when limits are reached
- **Caching and Optimization**: Response caching to reduce API calls
- **Context-aware Waiting**: Intelligent waiting with context cancellation support

## Key Achievements

### ✅ 4.8.1 Implement per-API rate limiting and quotas

**Status**: COMPLETED  
**Implementation**: `internal/modules/risk_assessment/external_rate_limiter.go`

**Key Components**:
- `ExternalAPIRateLimiter`: Main rate limiting service
- `ExternalRateLimitConfig`: Configuration management
- `APIConfig`: Per-API configuration settings
- `ExternalAPILimit`: Rate limit state tracking
- `GlobalRateLimit`: System-wide rate limiting
- `RateLimitMonitor`: Monitoring and metrics collection
- `RateLimitFallback`: Fallback strategy management
- `RateLimitOptimizer`: Caching and optimization

**Core Functionality**:
- Per-API rate limiting with minute, hour, and daily quotas
- Global rate limiting to prevent system-wide abuse
- Priority-based request processing
- Automatic quota reset and counter management
- Comprehensive API call tracking and statistics

## Architecture

### Core Components

#### ExternalAPIRateLimiter
The main rate limiting service that orchestrates all rate limiting operations.

```go
type ExternalAPIRateLimiter struct {
    config *ExternalRateLimitConfig
    logger *zap.Logger
    mu     sync.RWMutex
    
    // Per-API rate limits
    apiLimits map[string]*ExternalAPILimit
    
    // Global rate limiting
    globalLimits *GlobalRateLimit
    
    // Monitoring and alerting
    monitor *RateLimitMonitor
    
    // Fallback strategies
    fallback *RateLimitFallback
    
    // Optimization and caching
    optimizer *RateLimitOptimizer
}
```

#### ExternalRateLimitConfig
Configuration structure for the entire rate limiting system.

```go
type ExternalRateLimitConfig struct {
    // Global settings
    GlobalRequestsPerMinute int
    GlobalRequestsPerHour   int
    GlobalRequestsPerDay    int
    DefaultTimeout          time.Duration
    
    // Per-API configurations
    APIConfigs map[string]*APIConfig
    
    // Monitoring settings
    MonitorConfig *MonitorConfig
    
    // Fallback settings
    FallbackConfig *FallbackConfig
    
    // Optimization settings
    OptimizationConfig *OptimizationConfig
}
```

#### APIConfig
Individual API configuration with rate limits and behavior settings.

```go
type APIConfig struct {
    APIEndpoint       string
    RequestsPerMinute int
    RequestsPerHour   int
    RequestsPerDay    int
    Timeout           time.Duration
    Priority          int           // Higher number = higher priority
    RetryAttempts     int
    BackoffStrategy   string        // linear, exponential, jitter
    QuotaExceeded     bool
    Enabled           bool
}
```

### Data Models

#### ExternalAPILimit
Tracks the current state and statistics for each API endpoint.

```go
type ExternalAPILimit struct {
    Config *APIConfig
    
    // Current usage
    CurrentRequestsPerMinute int
    CurrentRequestsPerHour   int
    CurrentRequestsPerDay    int
    
    // Timestamps
    LastMinuteReset  time.Time
    LastHourReset    time.Time
    LastDayReset     time.Time
    
    // Status
    QuotaExceeded bool
    RetryAfter    time.Time
    LastError     error
    LastSuccess   time.Time
    
    // Statistics
    TotalRequests        int64
    SuccessfulRequests   int64
    FailedRequests       int64
    AverageResponseTime  time.Duration
}
```

#### ExternalRateLimitResult
Result of a rate limit check operation.

```go
type ExternalRateLimitResult struct {
    Allowed           bool
    APIEndpoint       string
    RemainingRequests int
    ResetTime         time.Time
    RetryAfter        time.Time
    QuotaExceeded     bool
    WaitTime          time.Duration
    Priority          int
    FallbackAvailable bool
    CacheHit          bool
}
```

## Configuration

### Default Configuration

The system provides sensible defaults through `DefaultExternalRateLimitConfig()`:

```go
func DefaultExternalRateLimitConfig() *ExternalRateLimitConfig {
    return &ExternalRateLimitConfig{
        GlobalRequestsPerMinute: 100,
        GlobalRequestsPerHour:   5000,
        GlobalRequestsPerDay:    100000,
        DefaultTimeout:          30 * time.Second,
        APIConfigs: map[string]*APIConfig{
            "default": {
                APIEndpoint:       "default",
                RequestsPerMinute: 60,
                RequestsPerHour:   1000,
                RequestsPerDay:    10000,
                Timeout:           30 * time.Second,
                Priority:          1,
                RetryAttempts:     3,
                BackoffStrategy:   "exponential",
                Enabled:           true,
            },
        },
        MonitorConfig: &MonitorConfig{
            Enabled:              true,
            MetricsCollectionInterval: 30 * time.Second,
            AlertThreshold:       0.8,
            AlertCooldown:        5 * time.Minute,
        },
        FallbackConfig: &FallbackConfig{
            Enabled:          true,
            FallbackAPIs:     []string{},
            CacheFallback:    true,
            RetryWithBackoff: true,
            MaxRetryAttempts: 3,
        },
        OptimizationConfig: &OptimizationConfig{
            Enabled:        true,
            CacheEnabled:   true,
            CacheTTL:       5 * time.Minute,
            RequestBatching: false,
            BatchSize:      10,
            BatchTimeout:   1 * time.Second,
        },
    }
}
```

### Custom Configuration

You can create custom configurations for specific APIs:

```go
config := &ExternalRateLimitConfig{
    GlobalRequestsPerMinute: 200,
    GlobalRequestsPerHour:   10000,
    GlobalRequestsPerDay:    200000,
    DefaultTimeout:          60 * time.Second,
    APIConfigs: map[string]*APIConfig{
        "whois-api": {
            APIEndpoint:       "whois-api",
            RequestsPerMinute: 30,
            RequestsPerHour:   500,
            RequestsPerDay:    5000,
            Timeout:           15 * time.Second,
            Priority:          5,
            RetryAttempts:     5,
            BackoffStrategy:   "exponential",
            Enabled:           true,
        },
        "ssl-api": {
            APIEndpoint:       "ssl-api",
            RequestsPerMinute: 100,
            RequestsPerHour:   2000,
            RequestsPerDay:    20000,
            Timeout:           10 * time.Second,
            Priority:          3,
            RetryAttempts:     3,
            BackoffStrategy:   "linear",
            Enabled:           true,
        },
    },
}
```

## Usage Examples

### Basic Rate Limiting

```go
// Create rate limiter
logger := zap.NewProduction()
config := DefaultExternalRateLimitConfig()
limiter := NewExternalAPIRateLimiter(config, logger)

// Check if request is allowed
ctx := context.Background()
result, err := limiter.CheckRateLimit(ctx, "whois-api")
if err != nil {
    log.Fatal(err)
}

if result.Allowed {
    // Make API call
    response, err := makeAPICall("whois-api", data)
    if err != nil {
        limiter.RecordAPICall("whois-api", false, responseTime, err)
    } else {
        limiter.RecordAPICall("whois-api", true, responseTime, nil)
    }
} else {
    // Handle rate limit exceeded
    log.Printf("Rate limit exceeded. Retry after: %v", result.RetryAfter)
}
```

### Waiting for Rate Limit

```go
// Wait until rate limit allows the request
ctx := context.Background()
err := limiter.WaitForRateLimit(ctx, "ssl-api")
if err != nil {
    if err == context.DeadlineExceeded {
        log.Println("Timeout waiting for rate limit")
    } else {
        log.Fatal(err)
    }
}

// Now safe to make API call
response, err := makeAPICall("ssl-api", data)
```

### Dynamic API Configuration

```go
// Add new API configuration
apiConfig := &APIConfig{
    APIEndpoint:       "new-api",
    RequestsPerMinute: 50,
    RequestsPerHour:   1000,
    RequestsPerDay:    10000,
    Timeout:           20 * time.Second,
    Priority:          2,
    RetryAttempts:     3,
    BackoffStrategy:   "exponential",
    Enabled:           true,
}

limiter.AddAPIConfig("new-api", apiConfig)

// Remove API configuration
limiter.RemoveAPIConfig("old-api")
```

### Rate Limit Status Monitoring

```go
// Get status for specific API
status := limiter.GetRateLimitStatus("whois-api")
if status != nil {
    fmt.Printf("API: %s\n", status.Config.APIEndpoint)
    fmt.Printf("Total Requests: %d\n", status.TotalRequests)
    fmt.Printf("Success Rate: %.2f%%\n", 
        float64(status.SuccessfulRequests)/float64(status.TotalRequests)*100)
    fmt.Printf("Average Response Time: %v\n", status.AverageResponseTime)
}

// Get global rate limit status
globalStatus := limiter.GetGlobalRateLimitStatus()
fmt.Printf("Global Requests/Minute: %d/%d\n", 
    globalStatus.CurrentRequestsPerMinute, 
    limiter.config.GlobalRequestsPerMinute)
```

### Rate Limit Reset

```go
// Reset specific API rate limit
limiter.ResetRateLimit("whois-api")

// Reset global rate limit
limiter.ResetGlobalRateLimit()
```

## Rate Limiting Algorithm

### Per-API Rate Limiting

1. **Counter Management**: Each API maintains separate counters for minute, hour, and day limits
2. **Automatic Reset**: Counters automatically reset when their time window expires
3. **Priority Processing**: Higher priority APIs get preferential treatment
4. **Quota Tracking**: Real-time tracking of remaining requests

### Global Rate Limiting

1. **System-wide Limits**: Prevents overall API abuse across all endpoints
2. **Hierarchical Checking**: Global limits checked before per-API limits
3. **Coordinated Reset**: Global counters reset independently of API-specific counters

### Waiting Strategy

1. **Context Awareness**: Respects context cancellation and timeouts
2. **Fallback Detection**: Checks for available fallback APIs
3. **Cache Checking**: Verifies if cached responses are available
4. **Intelligent Waiting**: Calculates optimal wait times based on reset schedules

## Monitoring and Metrics

### Rate Limit Metrics

The system tracks comprehensive metrics for each API:

- **Total Requests**: Total number of requests made
- **Successful Requests**: Number of successful API calls
- **Failed Requests**: Number of failed API calls
- **Average Response Time**: Average response time across all calls
- **Rate Limit Hits**: Number of times rate limits were exceeded
- **Wait Times**: Average time spent waiting for rate limits

### Monitoring Components

#### RateLimitMonitor
- Records rate limit checks and API calls
- Tracks metrics per API endpoint
- Provides alerting capabilities
- Maintains historical data

#### RateLimitMetrics
```go
type RateLimitMetrics struct {
    APIEndpoint        string
    TotalChecks        int64
    AllowedRequests    int64
    BlockedRequests    int64
    AverageWaitTime    time.Duration
    LastAlertTime      time.Time
}
```

## Fallback Strategies

### RateLimitFallback
Provides fallback mechanisms when primary APIs are rate-limited:

- **Alternative APIs**: Automatic switching to backup APIs
- **Cache Fallback**: Using cached responses when available
- **Retry with Backoff**: Exponential backoff for retries
- **Graceful Degradation**: Partial functionality when limits are exceeded

### Fallback Configuration
```go
type FallbackConfig struct {
    Enabled           bool
    FallbackAPIs      []string
    CacheFallback     bool
    RetryWithBackoff  bool
    MaxRetryAttempts  int
}
```

## Optimization and Caching

### RateLimitOptimizer
Provides optimization features to reduce API calls:

- **Response Caching**: Caches API responses to avoid redundant calls
- **Request Batching**: Batches multiple requests when possible
- **TTL Management**: Automatic cache expiration based on TTL
- **Cache Validation**: Validates cached responses before use

### Optimization Configuration
```go
type OptimizationConfig struct {
    Enabled           bool
    CacheEnabled      bool
    CacheTTL          time.Duration
    RequestBatching   bool
    BatchSize         int
    BatchTimeout      time.Duration
}
```

## Testing

### Test Coverage

The implementation includes comprehensive unit tests covering:

- **Rate Limiter Creation**: Testing constructor and configuration
- **Rate Limit Checking**: Testing allowed and blocked scenarios
- **Global Rate Limiting**: Testing system-wide limits
- **Waiting Behavior**: Testing context cancellation and timeouts
- **API Call Recording**: Testing metrics collection
- **Rate Limit Reset**: Testing counter reset functionality
- **Configuration Management**: Testing dynamic API configuration
- **Component Integration**: Testing monitor, fallback, and optimizer components

### Test Results

All tests pass successfully:
```
=== RUN   TestNewExternalAPIRateLimiter
--- PASS: TestNewExternalAPIRateLimiter (0.00s)
=== RUN   TestExternalAPIRateLimiter_CheckRateLimit_Allowed
--- PASS: TestExternalAPIRateLimiter_CheckRateLimit_Allowed (0.00s)
=== RUN   TestExternalAPIRateLimiter_CheckRateLimit_Blocked
--- PASS: TestExternalAPIRateLimiter_CheckRateLimit_Blocked (0.00s)
=== RUN   TestExternalAPIRateLimiter_CheckRateLimit_GlobalLimit
--- PASS: TestExternalAPIRateLimiter_CheckRateLimit_GlobalLimit (0.00s)
=== RUN   TestExternalAPIRateLimiter_WaitForRateLimit
--- PASS: TestExternalAPIRateLimiter_WaitForRateLimit (60.01s)
=== RUN   TestExternalAPIRateLimiter_RecordAPICall
--- PASS: TestExternalAPIRateLimiter_RecordAPICall (0.00s)
=== RUN   TestExternalAPIRateLimiter_ResetRateLimit
--- PASS: TestExternalAPIRateLimiter_ResetRateLimit (0.00s)
=== RUN   TestExternalAPIRateLimiter_ResetGlobalRateLimit
--- PASS: TestExternalAPIRateLimiter_ResetGlobalRateLimit (0.00s)
=== RUN   TestExternalAPIRateLimiter_AddAPIConfig
--- PASS: TestExternalAPIRateLimiter_AddAPIConfig (0.00s)
=== RUN   TestExternalAPIRateLimiter_RemoveAPIConfig
--- PASS: TestExternalAPIRateLimiter_RemoveAPIConfig (0.00s)
=== RUN   TestExternalAPIRateLimiter_GetGlobalRateLimitStatus
--- PASS: TestExternalAPIRateLimiter_GetGlobalRateLimitStatus (0.00s)
=== RUN   TestExternalAPIRateLimiter_ContextCancellation
--- PASS: TestExternalAPIRateLimiter_ContextCancellation (0.10s)
=== RUN   TestRateLimitMonitor_RecordRateLimitCheck
--- PASS: TestRateLimitMonitor_RecordRateLimitCheck (0.00s)
=== RUN   TestRateLimitFallback_HasFallback
--- PASS: TestRateLimitFallback_HasFallback (0.00s)
=== RUN   TestRateLimitOptimizer_HasCachedResponse
--- PASS: TestRateLimitOptimizer_HasCachedResponse (0.00s)
=== RUN   TestDefaultExternalRateLimitConfig
--- PASS: TestDefaultExternalRateLimitConfig (0.00s)
PASS
```

## Performance Considerations

### Memory Usage
- **Efficient Storage**: Uses maps for O(1) lookups
- **Minimal Overhead**: Lightweight structs for rate limit tracking
- **Automatic Cleanup**: Expired cache entries are automatically removed

### CPU Usage
- **Lock-free Reads**: Uses RWMutex for concurrent read access
- **Minimal Locking**: Short critical sections for write operations
- **Efficient Algorithms**: O(1) time complexity for rate limit checks

### Network Impact
- **Reduced API Calls**: Caching and optimization reduce external calls
- **Intelligent Waiting**: Prevents unnecessary polling
- **Fallback Strategies**: Reduces dependency on single API endpoints

## Security Features

### Rate Limit Protection
- **Prevents Abuse**: Protects against API abuse and DoS attacks
- **Quota Enforcement**: Strict enforcement of API quotas
- **Priority Management**: Ensures critical APIs remain available

### Configuration Security
- **Validation**: All configuration values are validated
- **Defaults**: Secure defaults prevent misconfiguration
- **Isolation**: Per-API isolation prevents cross-contamination

## Future Enhancements

### Planned Features
- **Distributed Rate Limiting**: Support for multi-instance deployments
- **Dynamic Configuration**: Runtime configuration updates
- **Advanced Analytics**: Detailed usage analytics and reporting
- **Machine Learning**: Predictive rate limit optimization

### Integration Opportunities
- **Prometheus Metrics**: Export metrics for monitoring systems
- **Alerting Integration**: Integration with alerting systems
- **Dashboard Integration**: Real-time rate limit dashboards
- **API Gateway Integration**: Integration with API gateways

## Best Practices

### Configuration
1. **Set Realistic Limits**: Configure limits based on actual API quotas
2. **Use Priority Wisely**: Assign higher priorities to critical APIs
3. **Enable Monitoring**: Always enable monitoring for production use
4. **Configure Fallbacks**: Set up fallback APIs for critical functionality

### Usage
1. **Check Rate Limits**: Always check rate limits before making API calls
2. **Handle Errors Gracefully**: Implement proper error handling for rate limit exceeded
3. **Use Context**: Always use context for cancellation and timeouts
4. **Monitor Metrics**: Regularly monitor rate limit metrics and adjust configuration

### Testing
1. **Test Rate Limits**: Test rate limit behavior in development
2. **Load Testing**: Perform load testing to validate rate limit effectiveness
3. **Fallback Testing**: Test fallback strategies under various scenarios
4. **Integration Testing**: Test integration with actual API endpoints

## Conclusion

The External API Rate Limiting implementation provides a robust, scalable, and feature-rich solution for managing external API calls. With comprehensive rate limiting, monitoring, fallback strategies, and optimization features, it ensures reliable and efficient API usage while protecting against abuse and quota violations.

The implementation successfully addresses **Task 4.8.1: Implement per-API rate limiting and quotas** and provides a solid foundation for the remaining rate limiting tasks in the Risk Assessment module.
