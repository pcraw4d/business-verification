# Redis Monitoring Guide for Production

## Overview

This guide explains how to monitor Redis cache performance and metrics in Railway after successful configuration.

---

## Step 1: Monitor Redis Metrics in Railway

### Accessing Redis Metrics

1. **Go to Railway Dashboard**
   - Navigate to your project
   - Click on **Redis Service**

2. **View Metrics Tab**
   - Click **"Metrics"** tab
   - You'll see real-time metrics

### Key Metrics to Monitor

#### Memory Usage

**What to Look For**:
- Memory usage should increase as cache fills
- Should stabilize after initial warm-up period
- Monitor for memory limits

**Expected Behavior**:
- Starts low (empty cache)
- Increases as requests are made
- Levels off as cache reaches steady state
- Typical usage: 10-100MB depending on cache size

**Alert Thresholds**:
- ⚠️ Warning: >80% of allocated memory
- 🚨 Critical: >95% of allocated memory

#### Active Connections

**What to Look For**:
- Should show connections from Classification Service
- Typically 1-5 connections (connection pooling)
- Should remain stable

**Expected Behavior**:
- Connections appear when Classification Service starts
- Connections remain active during service operation
- Connections close when service stops

**Normal Range**: 1-10 connections

#### Commands (Operations)

**What to Look For**:
- `GET` operations (cache reads)
- `SET` operations (cache writes)
- `DEL` operations (cache deletions)

**Expected Behavior**:
- `SET` operations on first requests (cache writes)
- `GET` operations on subsequent requests (cache reads)
- `GET` operations should outnumber `SET` operations over time (cache hits)

**Calculate Cache Hit Rate**:
```
Cache Hit Rate = (GET operations - SET operations) / GET operations * 100
```

**Target**: >60% hit rate for website content cache

---

## Step 2: Check Service Logs for Cache Operations

### In Railway Dashboard

1. Go to **Classification Service** → **Logs** tab
2. Look for cache-related log messages

### Expected Log Messages

#### Cache Initialization (Startup)

```
✅ Website content cache initialized
Redis cache initialized for classification service
```

#### Cache Hits

```
Cache hit from Redis
Classification served from cache
Stored in Redis cache
```

#### Cache Misses

```
Cache miss - fetching fresh content
Stored in Redis cache
```

#### Cache Operations

```
Cache hit from Redis key: classification:abc123
Stored in Redis cache key: classification:abc123 ttl: 5m
```

### Debug Logging

To see more detailed cache operations, enable debug mode:

```bash
# In Railway → Classification Service → Variables
LOG_LEVEL=debug
```

This will show:
- Cache key generation
- Cache lookup results
- Cache storage operations
- Redis connection status

---

## Step 3: Calculate Cache Hit Rate

### Method 1: From Redis Metrics

1. Go to **Redis Service** → **Metrics**
2. Note the command counts:
   - Total GET operations
   - Total SET operations
3. Calculate:
   ```
   Hit Rate = (GET - SET) / GET * 100
   ```

### Method 2: From Service Logs

Count cache-related log messages:
- Cache hits: "Cache hit from Redis"
- Cache misses: "Cache miss" or "Stored in Redis cache" (first time)

### Method 3: From Response Headers

Monitor `X-Cache` headers in API responses:
- `X-Cache: HIT` = Cache hit
- `X-Cache: MISS` = Cache miss

**Target Hit Rates**:
- **Website Content Cache**: >60%
- **Classification Result Cache**: >40%

---

## Step 4: Monitor Performance Improvements

### Response Time Tracking

**Before Redis**:
- Average: 6-8 seconds
- P95: 10-12 seconds
- P99: 15+ seconds

**After Redis (Expected)**:
- Cached requests: 0.1-0.2 seconds
- Uncached requests: 6-8 seconds
- Average (with cache): 1-3 seconds (depending on hit rate)

### Track Over Time

Monitor:
- Average response time
- P95 response time
- Cache hit rate
- Redis memory usage

### Performance Dashboard

Create a simple monitoring dashboard tracking:
1. **Response Times**: Average, P95, P99
2. **Cache Hit Rate**: Percentage
3. **Redis Memory**: Usage and trends
4. **Request Volume**: Total requests per hour/day

---

## Step 5: Set Up Alerts

### Recommended Alerts

#### Redis Memory Alert

**Condition**: Memory usage >80% of limit  
**Action**: Review cache TTL, consider increasing Redis memory

#### Cache Hit Rate Alert

**Condition**: Hit rate <40% for 1 hour  
**Action**: Review cache key generation, check TTL settings

#### Redis Connection Alert

**Condition**: Connection failures >5 in 5 minutes  
**Action**: Check Redis service health, verify network connectivity

#### Response Time Alert

**Condition**: Average response time >5 seconds for 10 minutes  
**Action**: Check cache hit rate, review service performance

---

## Step 6: Optimize Cache Configuration

### Based on Metrics

#### If Cache Hit Rate is Low (<40%)

**Possible Causes**:
- Cache TTL too short
- Requests not identical (different cache keys)
- Cache being cleared too frequently

**Solutions**:
1. Increase cache TTL
2. Review cache key generation
3. Check for cache invalidation logic

#### If Memory Usage is High (>80%)

**Possible Causes**:
- Cache TTL too long
- Too much data being cached
- Memory limit too low

**Solutions**:
1. Reduce cache TTL
2. Implement cache eviction policies
3. Increase Redis memory limit
4. Review what's being cached

#### If Response Times Still High

**Possible Causes**:
- Low cache hit rate
- Redis connection issues
- Network latency

**Solutions**:
1. Improve cache hit rate
2. Check Redis connection pool
3. Monitor network latency
4. Review cache key strategy

---

## Monitoring Checklist

### Daily Checks

- [ ] Redis memory usage
- [ ] Cache hit rate
- [ ] Average response times
- [ ] Error rates

### Weekly Reviews

- [ ] Cache hit rate trends
- [ ] Memory usage trends
- [ ] Performance improvements
- [ ] Cost savings (reduced external API calls)

### Monthly Analysis

- [ ] Overall performance improvements
- [ ] Cache effectiveness
- [ ] Optimization opportunities
- [ ] Capacity planning

---

## Tools and Resources

### Railway Dashboard

- **Redis Metrics**: Real-time metrics
- **Service Logs**: Cache operation logs
- **Service Health**: Overall service status

### External Monitoring (Optional)

- **Prometheus**: For detailed metrics
- **Grafana**: For visualization
- **Datadog/New Relic**: For APM

---

## Expected Results After 24 Hours

### Cache Hit Rate

- **Website Content**: 60-80%
- **Classification Results**: 40-60%

### Performance

- **Average Response Time**: 1-3 seconds (down from 6-8s)
- **Cached Requests**: 0.1-0.2 seconds
- **Performance Improvement**: 50-90% faster

### Redis Usage

- **Memory**: 20-100MB (depending on traffic)
- **Connections**: 1-5 active
- **Operations**: GET > SET (more hits than misses)

---

## Troubleshooting

### Low Cache Hit Rate

1. Check cache key generation
2. Verify requests are identical
3. Review cache TTL settings
4. Check for cache invalidation

### High Memory Usage

1. Review cache TTL
2. Check what's being cached
3. Implement eviction policies
4. Consider increasing Redis memory

### Connection Issues

1. Check Redis service health
2. Verify network connectivity
3. Review connection pool settings
4. Check for connection leaks

---

## Files

- **Monitoring Guide**: `docs/redis-monitoring-guide.md` (this document)
- **Test Results**: `docs/redis-cache-test-results.md`
- **Testing Guide**: `docs/redis-cache-testing-guide.md`

---

## Quick Reference

### Check Metrics

1. Railway Dashboard → Redis Service → Metrics
2. Check: Memory, Connections, Commands

### Check Logs

1. Railway Dashboard → Classification Service → Logs
2. Look for: "Cache hit", "Stored in Redis cache"

### Calculate Hit Rate

```
Hit Rate = (GET - SET) / GET * 100
```

Target: >60%

---

## Success Criteria

✅ **Cache Hit Rate**: >60%  
✅ **Response Times**: <3s average  
✅ **Memory Usage**: Stable and within limits  
✅ **No Connection Issues**: Stable connections  
✅ **Performance Improvement**: 50-90% faster  

