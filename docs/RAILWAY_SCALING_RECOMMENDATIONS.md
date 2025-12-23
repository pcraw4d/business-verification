# Railway Scaling Recommendations - Fix 5.7

## Overview

This document provides recommendations for scaling the classification service on Railway to handle increased load and improve performance.

## Current Configuration

### Service Limits
- **Max Concurrent Requests**: 20 (configurable via `MAX_CONCURRENT_REQUESTS`)
- **Request Timeout**: 120 seconds
- **Memory Limit**: 768 MiB (default, configurable via `GOMEMLIMIT_MB`)
- **Worker Pool Size**: 30% of max concurrent requests (default: 6 workers, max: 20)

### Rate Limiting
- **Global Rate Limit**: 200 requests/minute (configurable via `GLOBAL_RATE_LIMIT`)
- **Per-IP Rate Limit**: 100 requests/minute (configurable via `PER_IP_RATE_LIMIT`)
- **Burst Size**: 20 requests (configurable via `RATE_LIMIT_BURST`)

## Scaling Recommendations

### 1. Increase Memory Allocation (IMMEDIATE)

**Current**: Default Railway limits (typically 512MB-1GB)

**Recommendation**: Increase to 2GB-4GB

**Steps**:
1. In Railway dashboard, go to your service settings
2. Navigate to "Resources" or "Scaling" section
3. Increase memory allocation to 2GB (2048 MB) minimum
4. For production with high load, consider 4GB (4096 MB)

**Expected Impact**:
- Reduces OOM kills
- Allows more concurrent requests
- Improves cache hit rates
- Better handling of memory-intensive operations

**Configuration**:
```bash
# Set GOMEMLIMIT to 80% of allocated memory
GOMEMLIMIT_MB=1536  # For 2GB allocation (80% = 1.6GB)
GOMEMLIMIT_MB=3072  # For 4GB allocation (80% = 3.2GB)
```

### 2. Increase CPU Allocation (MEDIUM PRIORITY)

**Current**: Default Railway CPU limits

**Recommendation**: Increase to 2-4 vCPUs

**Steps**:
1. In Railway dashboard, go to service settings
2. Increase CPU allocation to 2 vCPUs minimum
3. For production, consider 4 vCPUs

**Expected Impact**:
- Faster request processing
- Better parallel processing
- Reduced latency
- Improved worker pool efficiency

### 3. Horizontal Scaling (HIGH PRIORITY)

**Current**: Single service instance

**Recommendation**: Deploy 2-3 instances with load balancing

**Steps**:
1. In Railway dashboard, enable horizontal scaling
2. Set minimum instances: 2
3. Set maximum instances: 5 (for auto-scaling)
4. Configure load balancer (Railway handles this automatically)

**Expected Impact**:
- Increased throughput (2-3x)
- Better fault tolerance
- Improved availability
- Automatic scaling under load

**Configuration**:
```bash
# Each instance should have:
MAX_CONCURRENT_REQUESTS=20  # Per instance
GLOBAL_RATE_LIMIT=200       # Per instance (total = instances * 200)
```

### 4. Enable Auto-Scaling (RECOMMENDED)

**Configuration**:
- **Min Instances**: 2
- **Max Instances**: 5
- **Scale Up Trigger**: CPU > 70% OR Memory > 80% OR Queue Depth > 15
- **Scale Down Trigger**: CPU < 30% AND Memory < 50% AND Queue Depth < 5 (for 5 minutes)

**Expected Impact**:
- Automatic scaling based on load
- Cost optimization (scale down during low traffic)
- Better handling of traffic spikes

### 5. Database Connection Pooling (IMPORTANT)

**Current**: Default Supabase connection limits

**Recommendation**: Optimize connection pool settings

**Configuration**:
```go
// In Supabase client configuration
MaxOpenConns: 25
MaxIdleConns: 5
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 1 * time.Minute
```

**Expected Impact**:
- Better database performance
- Reduced connection overhead
- Improved query response times

### 6. Redis Cache Scaling (OPTIONAL BUT RECOMMENDED)

**Current**: In-memory cache only (or optional Redis)

**Recommendation**: Enable Redis for distributed caching

**Benefits**:
- Shared cache across instances
- Better cache hit rates
- Reduced memory pressure per instance
- Improved performance

**Configuration**:
```bash
REDIS_ENABLED=true
REDIS_URL=<your-redis-url>
CACHE_TTL=10m
```

### 7. Resource Monitoring and Alerts

**Recommendation**: Set up monitoring for:
- Memory usage (alert if > 80%)
- CPU usage (alert if > 85%)
- Request queue depth (alert if > 15)
- Error rate (alert if > 5%)
- Response latency (alert if P95 > 10s)

**Railway Dashboard**:
- Enable metrics collection
- Set up alerts for resource thresholds
- Monitor service health

## Implementation Priority

### Phase 1: Immediate (Week 1)
1. ✅ Increase memory to 2GB
2. ✅ Set GOMEMLIMIT_MB=1536
3. ✅ Enable Redis cache (if available)

### Phase 2: Short-term (Week 2-3)
1. ✅ Increase CPU to 2 vCPUs
2. ✅ Deploy 2 instances (horizontal scaling)
3. ✅ Set up basic monitoring

### Phase 3: Long-term (Month 1)
1. ✅ Enable auto-scaling (2-5 instances)
2. ✅ Optimize database connection pooling
3. ✅ Set up comprehensive monitoring and alerts

## Cost Considerations

### Current (Single Instance, 1GB RAM, 1 vCPU)
- Estimated: $5-10/month (Railway pricing)

### Recommended (2 Instances, 2GB RAM each, 2 vCPU each)
- Estimated: $20-40/month

### Production (3 Instances, 4GB RAM each, 4 vCPU each, Auto-scaling)
- Estimated: $60-120/month

## Performance Targets

After implementing scaling recommendations:

| Metric | Current | Target | Method |
|--------|---------|--------|--------|
| Max Concurrent Requests | 20 | 60-100 | Horizontal scaling (3 instances × 20-30) |
| Request Throughput | ~100 req/min | 300-500 req/min | Multiple instances + load balancing |
| Average Latency | 43.7s | <10s | More resources + optimizations |
| Error Rate | 67.1% | <5% | Better resource allocation + rate limiting |
| Memory Usage | High | <70% | Increased allocation + LRU cache |

## Monitoring

### Key Metrics to Track
1. **Request Queue Depth**: Should stay < 10 per instance
2. **Memory Usage**: Should stay < 70%
3. **CPU Usage**: Should stay < 80%
4. **Error Rate**: Should stay < 5%
5. **Response Latency**: P95 should be < 10s
6. **Cache Hit Rate**: Should be > 60%

### Health Endpoint
Monitor `/health` endpoint for:
- Memory usage percentage
- Queue size
- Active workers
- Cache size
- Service status

## Rollout Plan

1. **Week 1**: Increase memory allocation, set GOMEMLIMIT
2. **Week 2**: Deploy second instance, test load balancing
3. **Week 3**: Enable auto-scaling, optimize configurations
4. **Week 4**: Monitor and tune based on metrics

## Notes

- All scaling should be done incrementally
- Monitor closely after each change
- Have rollback plan ready
- Test thoroughly before production deployment
- Consider cost vs. performance trade-offs

---

**Last Updated**: December 23, 2025  
**Status**: Recommendations ready for implementation

