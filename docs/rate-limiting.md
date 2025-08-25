# Rate Limiting System

## Overview

The Enhanced Business Intelligence System implements a comprehensive rate limiting solution to protect against abuse, ensure fair resource usage, and maintain system stability. The rate limiting system consists of two main components:

1. **General API Rate Limiting** - Protects all API endpoints
2. **Authentication Rate Limiting** - Specifically protects authentication endpoints with lockout mechanisms

## Architecture

### Components

#### 1. APIRateLimiter
- **Purpose**: General rate limiting for all API endpoints
- **Strategy**: Token bucket algorithm with configurable burst sizes
- **Storage**: In-memory or Redis-based (distributed)
- **Key Generation**: Based on client IP, user ID, API key, and endpoint path

#### 2. AuthRateLimiter
- **Purpose**: Specialized rate limiting for authentication endpoints
- **Strategy**: Sliding window with lockout mechanisms
- **Features**: Failed attempt tracking, progressive lockouts, account protection
- **Endpoints**: Login, registration, password reset, verification

#### 3. Rate Limit Stores
- **MemoryRateLimitStore**: In-memory storage for single-instance deployments
- **RedisRateLimitStore**: Redis-based storage for distributed deployments
- **MemoryAuthRateLimitStore**: In-memory storage for authentication rate limiting
- **RedisAuthRateLimitStore**: Redis-based storage for authentication rate limiting

### Key Features

- **Multiple Rate Limiting Strategies**: Token bucket, sliding window, fixed window
- **Distributed Rate Limiting**: Redis support for multi-instance deployments
- **Flexible Key Generation**: IP-based, user-based, API key-based, or composite keys
- **Progressive Lockouts**: Escalating lockout durations for repeated violations
- **Real-time Statistics**: Request counts, blocked requests, active lockouts
- **Automatic Cleanup**: Periodic cleanup of expired rate limit entries
- **Configurable Limits**: Different limits for different endpoints and user types

## Configuration

### General Rate Limiting Configuration

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100
  burst_size: 20
  window_size: 60s
  strategy: "token_bucket"  # token_bucket, sliding_window, fixed_window
  distributed: false
  redis_url: "redis://localhost:6379"
  redis_key_prefix: "rate_limit"
  cleanup_interval: 5m
  max_keys: 10000
```

### Authentication Rate Limiting Configuration

```yaml
auth_rate_limit:
  enabled: true
  login_attempts_per: 5
  register_attempts_per: 3
  password_reset_attempts_per: 3
  window_size: 60s
  lockout_duration: 15m
  max_lockouts: 3
  permanent_lockout_duration: 24h
  distributed: false
  redis_url: "redis://localhost:6379"
  redis_key_prefix: "auth_rate_limit"
```

### Environment Variables

```bash
# General Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST_SIZE=20
RATE_LIMIT_WINDOW_SIZE=60s
RATE_LIMIT_STRATEGY=token_bucket
RATE_LIMIT_DISTRIBUTED=false
RATE_LIMIT_REDIS_URL=redis://localhost:6379
RATE_LIMIT_REDIS_KEY_PREFIX=rate_limit

# Authentication Rate Limiting
AUTH_RATE_LIMIT_ENABLED=true
AUTH_RATE_LIMIT_LOGIN_ATTEMPTS_PER=5
AUTH_RATE_LIMIT_REGISTER_ATTEMPTS_PER=3
AUTH_RATE_LIMIT_PASSWORD_RESET_ATTEMPTS_PER=3
AUTH_RATE_LIMIT_WINDOW_SIZE=60s
AUTH_RATE_LIMIT_LOCKOUT_DURATION=15m
AUTH_RATE_LIMIT_MAX_LOCKOUTS=3
AUTH_RATE_LIMIT_PERMANENT_LOCKOUT_DURATION=24h
AUTH_RATE_LIMIT_DISTRIBUTED=false
AUTH_RATE_LIMIT_REDIS_URL=redis://localhost:6379
AUTH_RATE_LIMIT_REDIS_KEY_PREFIX=auth_rate_limit
```

## Usage

### Basic Setup

```go
package main

import (
    "github.com/pcraw4d/business-verification/internal/api/middleware"
    "go.uber.org/zap"
)

func main() {
    logger := zap.NewProduction()
    
    // Configure general rate limiting
    rateLimitConfig := &middleware.RateLimitConfig{
        Enabled:           true,
        RequestsPerMinute: 100,
        BurstSize:         20,
        WindowSize:        time.Minute,
        Strategy:          "token_bucket",
        Distributed:       false,
    }
    
    // Create rate limiter
    rateLimiter := middleware.NewAPIRateLimiter(rateLimitConfig, logger)
    
    // Configure authentication rate limiting
    authRateLimitConfig := &middleware.AuthRateLimitConfig{
        Enabled:                  true,
        LoginAttemptsPer:         5,
        RegisterAttemptsPer:      3,
        PasswordResetAttemptsPer: 3,
        WindowSize:               60 * time.Second,
        LockoutDuration:          15 * time.Minute,
        MaxLockouts:              3,
        PermanentLockoutDuration: 24 * time.Hour,
    }
    
    // Create auth rate limiter
    authRateLimiter := middleware.NewAuthRateLimiter(authRateLimitConfig, logger)
    
    // Apply middleware to your HTTP server
    mux := http.NewServeMux()
    
    // Apply auth rate limiting first (more restrictive)
    handler := authRateLimiter.Middleware(mux)
    
    // Apply general rate limiting
    handler = rateLimiter.Middleware(handler)
    
    // Start server
    http.ListenAndServe(":8080", handler)
}
```

### Advanced Configuration

```go
// Distributed rate limiting with Redis
rateLimitConfig := &middleware.RateLimitConfig{
    Enabled:        true,
    Distributed:    true,
    RedisURL:       "redis://localhost:6379",
    RedisKeyPrefix: "rate_limit",
    Strategy:       "sliding_window",
    MaxKeys:        50000,
}

// Custom key generation
type CustomRateLimiter struct {
    *middleware.APIRateLimiter
}

func (crl *CustomRateLimiter) generateKey(r *http.Request) string {
    // Custom key generation logic
    clientIP := crl.getClientIP(r)
    userAgent := r.Header.Get("User-Agent")
    return fmt.Sprintf("%s:%s", clientIP, userAgent)
}

// Dynamic configuration updates
func updateRateLimitingConfig(rateLimiter *middleware.APIRateLimiter) {
    newConfig := &middleware.RateLimitConfig{
        Enabled:           true,
        RequestsPerMinute: 200, // Increased limit
        BurstSize:         50,
        WindowSize:        time.Minute,
    }
    
    rateLimiter.UpdateConfig(newConfig)
}
```

### Integration with Authentication

```go
// Record failed authentication attempts
func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Attempt authentication
    if !authenticateUser(r) {
        // Record failed attempt
        authRateLimiter.RecordFailedAttempt(r, middleware.LoginAttempt)
        
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Successful authentication
    // Reset rate limit for this user
    key := authRateLimiter.generateKey(r)
    authRateLimiter.ResetKey(key)
    
    // Continue with login process
    w.WriteHeader(http.StatusOK)
}
```

## Rate Limiting Strategies

### 1. Token Bucket Algorithm

**Description**: A bucket is filled with tokens at a constant rate. Each request consumes one token. If the bucket is empty, the request is rejected.

**Advantages**:
- Allows for burst traffic
- Simple to understand and implement
- Predictable behavior

**Configuration**:
```yaml
strategy: "token_bucket"
requests_per_minute: 100
burst_size: 20
```

### 2. Sliding Window Algorithm

**Description**: Maintains a sliding window of requests. New requests are allowed if the count within the window is below the limit.

**Advantages**:
- More accurate rate limiting
- Better for distributed systems
- Prevents edge case issues

**Configuration**:
```yaml
strategy: "sliding_window"
requests_per_minute: 100
window_size: 60s
```

### 3. Fixed Window Algorithm

**Description**: Divides time into fixed windows. Each window has a separate counter. When a window expires, the counter resets.

**Advantages**:
- Simple implementation
- Low memory usage
- Fast performance

**Configuration**:
```yaml
strategy: "fixed_window"
requests_per_minute: 100
window_size: 60s
```

## Key Generation Strategies

### 1. IP-Based Keys

```go
// Basic IP-based key generation
key := clientIP
```

**Use Cases**:
- General API protection
- DDoS protection
- Geographic rate limiting

### 2. User-Based Keys

```go
// User ID-based key generation
key := fmt.Sprintf("%s:%s", clientIP, userID)
```

**Use Cases**:
- Per-user rate limiting
- Premium user tiers
- User-specific quotas

### 3. API Key-Based Keys

```go
// API key-based key generation
apiKeyHash := sha256.Sum256([]byte(apiKey))
key := fmt.Sprintf("%s:%s", clientIP, hex.EncodeToString(apiKeyHash[:16]))
```

**Use Cases**:
- API key rate limiting
- Tiered API access
- Partner integrations

### 4. Composite Keys

```go
// Composite key with multiple factors
keyParts := []string{clientIP}
if userID != "" {
    keyParts = append(keyParts, userID)
}
if path != "" {
    pathHash := sha256.Sum256([]byte(path))
    keyParts = append(keyParts, hex.EncodeToString(pathHash[:8]))
}
key := strings.Join(keyParts, ":")
```

**Use Cases**:
- Endpoint-specific limits
- Complex rate limiting rules
- Multi-factor identification

## HTTP Headers

### Rate Limit Headers

The system sets the following headers on all responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Authentication Rate Limit Headers

For authentication endpoints:

```
X-AuthRateLimit-Limit: 5
X-AuthRateLimit-Remaining: 4
X-AuthRateLimit-Reset: 1640995200
```

### Retry-After Header

When rate limits are exceeded:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
```

## Monitoring and Statistics

### Rate Limiting Statistics

```go
// Get current statistics
stats := rateLimiter.GetStats()
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Blocked Requests: %d\n", stats.BlockedRequests)
fmt.Printf("Active Keys: %d\n", stats.ActiveKeys)
fmt.Printf("Last Reset: %s\n", stats.LastReset)
```

### Authentication Rate Limiting Statistics

```go
// Get auth rate limiting statistics
authStats := authRateLimiter.GetStats()
fmt.Printf("Total Attempts: %d\n", authStats.TotalAttempts)
fmt.Printf("Failed Attempts: %d\n", authStats.FailedAttempts)
fmt.Printf("Lockouts: %d\n", authStats.Lockouts)
fmt.Printf("Active Lockouts: %d\n", authStats.ActiveLockouts)
```

### Prometheus Metrics

```go
// Example Prometheus metrics
var (
    rateLimitRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rate_limit_requests_total",
            Help: "Total number of rate limit requests",
        },
        []string{"key", "allowed"},
    )
    
    rateLimitRemaining = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "rate_limit_remaining",
            Help: "Remaining requests for rate limit key",
        },
        []string{"key"},
    )
)
```

## Best Practices

### 1. Configuration Guidelines

- **Start Conservative**: Begin with lower limits and increase based on usage patterns
- **Monitor Performance**: Track rate limiting impact on legitimate users
- **Use Different Limits**: Apply different limits for different endpoints and user types
- **Plan for Bursts**: Configure burst sizes to handle legitimate traffic spikes

### 2. Key Generation Best Practices

- **Use Stable Keys**: Ensure keys remain consistent for the same client
- **Avoid Collisions**: Use sufficient entropy in key generation
- **Consider Privacy**: Don't include sensitive information in keys
- **Balance Performance**: Keep keys reasonably short for performance

### 3. Distributed Rate Limiting

- **Use Redis**: For multi-instance deployments, use Redis for shared state
- **Configure Redis Properly**: Set appropriate TTL and memory limits
- **Monitor Redis Performance**: Track Redis latency and memory usage
- **Plan for Failures**: Have fallback mechanisms for Redis outages

### 4. Security Considerations

- **Protect Against Bypass**: Ensure rate limiting cannot be easily bypassed
- **Monitor for Abuse**: Track unusual patterns and adjust limits accordingly
- **Use Progressive Penalties**: Implement escalating lockouts for repeated violations
- **Log Security Events**: Log all rate limiting violations for security analysis

### 5. Performance Optimization

- **Use Efficient Storage**: Choose appropriate storage backends for your scale
- **Optimize Key Generation**: Use fast hashing algorithms and efficient string operations
- **Implement Cleanup**: Regularly clean up expired entries to prevent memory leaks
- **Cache Frequently**: Cache rate limit results for frequently accessed keys

## Troubleshooting

### Common Issues

#### 1. Rate Limiting Too Aggressive

**Symptoms**: Legitimate users getting rate limited frequently

**Solutions**:
- Increase rate limits for affected endpoints
- Adjust burst sizes to allow for traffic spikes
- Review key generation to ensure proper client identification
- Check for proxy or load balancer configurations affecting client IP detection

#### 2. Memory Usage High

**Symptoms**: High memory usage in rate limiting components

**Solutions**:
- Reduce `max_keys` configuration
- Decrease `cleanup_interval` for more frequent cleanup
- Use Redis for distributed storage
- Monitor and optimize key generation

#### 3. Redis Performance Issues

**Symptoms**: High latency in distributed rate limiting

**Solutions**:
- Optimize Redis configuration
- Use Redis clustering for high availability
- Implement Redis connection pooling
- Monitor Redis memory usage and implement eviction policies

#### 4. False Positives

**Symptoms**: Legitimate requests being blocked incorrectly

**Solutions**:
- Review key generation logic
- Check for shared IP addresses (corporate networks, NAT)
- Implement whitelist mechanisms for trusted clients
- Use user-based keys instead of IP-based keys where appropriate

### Debugging Commands

```bash
# Check rate limiting statistics
curl -X GET "http://localhost:8080/api/v1/admin/rate-limit/stats" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Reset rate limit for specific key
curl -X POST "http://localhost:8080/api/v1/admin/rate-limit/reset" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key": "192.168.1.1"}'

# Check Redis rate limiting data
redis-cli KEYS "rate_limit:*"
redis-cli TTL "rate_limit:192.168.1.1"
```

## Future Enhancements

### Planned Features

1. **Machine Learning-Based Rate Limiting**
   - Adaptive limits based on user behavior patterns
   - Anomaly detection for automated attacks
   - Dynamic threshold adjustment

2. **Geographic Rate Limiting**
   - Country-based rate limits
   - Regional traffic management
   - Compliance with local regulations

3. **Advanced Analytics**
   - Real-time rate limiting dashboards
   - Predictive analytics for capacity planning
   - Automated optimization recommendations

4. **Enhanced Security**
   - CAPTCHA integration for repeated violations
   - Device fingerprinting for better identification
   - Integration with threat intelligence feeds

5. **API Gateway Integration**
   - Native integration with popular API gateways
   - Standard rate limiting protocols (RFC 6585)
   - Webhook notifications for rate limit events

### Contributing

To contribute to the rate limiting system:

1. Follow the existing code style and patterns
2. Add comprehensive tests for new features
3. Update documentation for configuration changes
4. Include performance benchmarks for optimizations
5. Consider backward compatibility for existing deployments

## Conclusion

The rate limiting system provides comprehensive protection against abuse while maintaining flexibility for legitimate use cases. By implementing multiple strategies, supporting distributed deployments, and providing extensive monitoring capabilities, the system can scale with your application while maintaining security and performance.

For additional support or questions, please refer to the API documentation or contact the development team.
