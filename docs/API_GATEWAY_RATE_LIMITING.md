# API Gateway Rate Limiting Configuration

**Date**: 2025-11-10  
**Service**: API Gateway

---

## Overview

The API Gateway implements in-memory rate limiting to protect backend services from excessive requests.

---

## Configuration

### Default Settings

Rate limiting is **enabled by default** with the following thresholds:

- **Enabled**: `true`
- **Requests Per Window**: `1000`
- **Window Size**: `3600` seconds (1 hour)
- **Burst Size**: `2000`

### What This Means

- **1000 requests per hour** per client IP address
- **Burst allowance**: Up to 2000 requests in a short period
- **Window**: Rolling 1-hour window

---

## Environment Variables

Rate limiting can be configured via environment variables:

```bash
# Enable/disable rate limiting
RATE_LIMIT_ENABLED=true

# Number of requests allowed per window
RATE_LIMIT_REQUESTS_PER=1000

# Window size in seconds
RATE_LIMIT_WINDOW_SIZE=3600

# Burst size (maximum requests in short period)
RATE_LIMIT_BURST_SIZE=2000
```

---

## How It Works

1. **Client Identification**: Uses client IP address (from `X-Forwarded-For` or `X-Real-IP` headers, or `RemoteAddr`)

2. **Request Tracking**: Maintains a map of client IPs to request timestamps

3. **Window Calculation**: Removes requests older than the window size

4. **Limit Check**: If requests within the window exceed the limit, returns HTTP 429 (Too Many Requests)

5. **Cleanup**: Runs a cleanup goroutine every 5 minutes to remove old entries

---

## Response When Rate Limited

When rate limit is exceeded, the API Gateway returns:

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests"
}
```

**HTTP Status**: `429 Too Many Requests`

---

## Testing Rate Limiting

### Current Thresholds

With default settings (1000 requests/hour), you would need to send **more than 1000 requests in an hour** to trigger rate limiting.

### Testing with Lower Thresholds

To test rate limiting, temporarily lower the thresholds:

```bash
# Set in Railway environment variables
RATE_LIMIT_REQUESTS_PER=10
RATE_LIMIT_WINDOW_SIZE=60
```

Then send 11 rapid requests to trigger rate limiting.

### Example Test

```bash
# Send 11 rapid requests
for i in {1..11}; do
  curl -s -w "\n%{http_code}\n" "https://api-gateway-service-production-21fd.up.railway.app/health"
done
```

The 11th request should return `429 Too Many Requests`.

---

## Production Recommendations

### Current Settings (1000/hour)

**Pros**:
- Very permissive, unlikely to affect legitimate users
- Good for high-traffic scenarios
- Prevents abuse without blocking normal usage

**Cons**:
- May not catch sophisticated attacks
- High threshold means rate limiting rarely triggers

### Recommended Settings for Production

For a production API, consider:

```bash
RATE_LIMIT_REQUESTS_PER=100    # 100 requests per minute
RATE_LIMIT_WINDOW_SIZE=60      # 1 minute window
RATE_LIMIT_BURST_SIZE=150     # Allow bursts up to 150
```

This provides:
- **100 requests per minute** per IP
- Better protection against abuse
- Still allows for normal usage patterns

---

## Monitoring

Rate limiting events are logged with:

```go
logger.Warn("Rate limit exceeded",
    zap.String("key", clientIP),
    zap.Int("requests", len(validRequests)),
    zap.Int("limit", rl.config.RequestsPer))
```

Monitor these logs to:
- Identify abusive clients
- Adjust thresholds if needed
- Track rate limiting effectiveness

---

## Implementation Details

### Location
- **Middleware**: `services/api-gateway/internal/middleware/rate_limit.go`
- **Configuration**: `services/api-gateway/internal/config/config.go`

### Algorithm
- In-memory storage (not distributed)
- Sliding window approach
- Automatic cleanup of old entries

### Limitations
- **In-memory only**: Rate limits are per-instance, not shared across multiple gateway instances
- **No persistence**: Restarting the service resets rate limit counters
- **IP-based**: Uses client IP, which may not be accurate behind proxies

### Future Improvements
- Redis-based distributed rate limiting
- Per-endpoint rate limits
- Per-user rate limits (with authentication)
- Rate limit headers in responses (X-RateLimit-*)

---

## Rate Limit Headers (Future Enhancement)

Consider adding rate limit headers to responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1636560000
```

This helps clients understand their rate limit status.

---

**Last Updated**: 2025-11-10

