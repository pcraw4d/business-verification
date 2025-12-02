# Redis Cache Testing and Monitoring Guide

## Overview

This guide explains how to test Redis cache functionality and monitor Redis metrics after configuration.

---

## Step 1: Test Cache Functionality

### Option A: Using the Test Script (Recommended)

A test script is available to automatically test cache functionality:

```bash
# Set your service URL
export CLASSIFICATION_SERVICE_URL=https://your-service.railway.app

# Run the test script
./scripts/test-redis-cache.sh
```

**What it tests**:
- ✅ First request (should be cache MISS)
- ✅ Second request (should be cache HIT)
- ✅ Third request (should be cache HIT)
- ✅ Performance improvement comparison
- ✅ Response time differences

**Expected Output**:
```
Request 1 (MISS): 6.5s
Request 2 (HIT):  0.8s
Request 3 (HIT):  0.7s

✅ Cache is working!
   Performance improvement: 87.69%
```

### Option B: Manual Testing with curl

#### Test 1: First Request (Cache MISS)

```bash
curl -X POST https://your-service.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A test business",
    "website": "https://example.com"
  }' \
  -v -w "\nTime: %{time_total}s\n"
```

**Look for**:
- HTTP 200 status
- Response time: ~6-8 seconds (first request)
- No `X-Cache` header or `X-Cache: MISS`

#### Test 2: Second Request (Cache HIT)

```bash
# Make identical request
curl -X POST https://your-service.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A test business",
    "website": "https://example.com"
  }' \
  -v -w "\nTime: %{time_total}s\n"
```

**Look for**:
- HTTP 200 status
- Response time: ~0.5-2 seconds (much faster!)
- `X-Cache: HIT` header (if implemented)

### Option C: Using HTTPie (Alternative)

```bash
# Install HTTPie if needed
# brew install httpie  # macOS
# apt-get install httpie  # Linux

# First request
http POST https://your-service.railway.app/classify \
  business_name="Test Company" \
  description="A test business" \
  website="https://example.com"

# Second request (should be faster)
http POST https://your-service.railway.app/classify \
  business_name="Test Company" \
  description="A test business" \
  website="https://example.com"
```

---

## Step 2: Monitor Redis Metrics in Railway

### Accessing Redis Metrics

1. **Go to Railway Dashboard**
   - Navigate to your project
   - Click on **Redis Service**

2. **View Metrics Tab**
   - Click **"Metrics"** tab
   - You'll see real-time metrics

### Key Metrics to Monitor

#### Memory Usage

**What to look for**:
- Memory usage should increase as cache fills
- Should stabilize after initial warm-up period
- Monitor for memory limits

**Expected behavior**:
- Starts low (empty cache)
- Increases as requests are made
- Levels off as cache reaches steady state

#### Active Connections

**What to look for**:
- Should show connections from Classification Service
- Typically 1-5 connections (connection pooling)
- Should remain stable

**Expected behavior**:
- Connections appear when Classification Service starts
- Connections remain active during service operation
- Connections close when service stops

#### Commands (Operations)

**What to look for**:
- `GET` operations (cache reads)
- `SET` operations (cache writes)
- `DEL` operations (cache deletions)

**Expected behavior**:
- `SET` operations on first requests (cache writes)
- `GET` operations on subsequent requests (cache reads)
- `GET` operations should outnumber `SET` operations over time (cache hits)

### Monitoring Cache Hit Rate

**Calculate cache hit rate**:
```
Cache Hit Rate = (GET operations - SET operations) / GET operations * 100
```

**Target**:
- **Website Content Cache**: >60% hit rate
- **Classification Result Cache**: >40% hit rate

**How to improve**:
- Increase cache TTL if hit rate is low
- Monitor which content is being cached
- Adjust cache keys if needed

---

## Step 3: Check Service Logs for Cache Operations

### In Railway Dashboard

1. Go to **Classification Service** → **Logs** tab
2. Look for cache-related log messages

### Expected Log Messages

#### Cache Hits

```
Cache hit from Redis
Stored in Redis cache
```

#### Cache Misses

```
Cache miss - fetching fresh content
Stored in Redis cache
```

#### Cache Operations

```
Website content cache initialized
Redis cache initialized for classification service
```

### Debug Logging

If you need more detailed logging, you can enable debug mode:

```bash
# In Railway → Classification Service → Variables
LOG_LEVEL=debug
```

This will show more detailed cache operations in logs.

---

## Step 4: Performance Comparison

### Before Redis (Baseline)

- **Response Time**: 6-8 seconds
- **External Requests**: Multiple per classification
- **Cache Hit Rate**: 0%
- **Scalability**: Limited by external API rates

### After Redis (Expected)

- **Cached Requests**: 0.5-2 seconds (50-90% faster)
- **External Requests**: Significantly reduced
- **Cache Hit Rate**: >60% for website content
- **Scalability**: Improved with distributed caching

### Measuring Performance

1. **Make multiple requests** with the same data
2. **Compare response times**:
   - First request: Should be slow (cache MISS)
   - Subsequent requests: Should be fast (cache HIT)
3. **Calculate improvement**:
   ```
   Improvement = (Time1 - Time2) / Time1 * 100
   ```

---

## Troubleshooting

### Issue: Cache Not Working

**Symptoms**:
- No performance improvement
- Response times similar for all requests
- No cache operations in Redis metrics

**Solutions**:
1. Verify `ENABLE_WEBSITE_CONTENT_CACHE=true`
2. Check `REDIS_ENABLED=true`
3. Verify Redis connection in logs
4. Check Redis metrics for operations

### Issue: Low Cache Hit Rate

**Symptoms**:
- Cache hit rate <40%
- Most requests are cache misses

**Solutions**:
1. Check if requests are truly identical
2. Verify cache TTL is appropriate
3. Monitor cache key generation
4. Check if cache is being cleared too frequently

### Issue: High Memory Usage

**Symptoms**:
- Redis memory usage is high
- Memory warnings in Railway

**Solutions**:
1. Reduce cache TTL
2. Implement cache eviction policies
3. Monitor cache size
4. Consider increasing Redis memory limit

---

## Automated Testing

### Continuous Monitoring

Set up monitoring to track:
- Cache hit rate over time
- Average response times
- Redis memory usage
- Error rates

### Alerting

Set up alerts for:
- Cache hit rate drops below threshold
- Redis memory usage exceeds limit
- Redis connection failures
- Response time degradation

---

## Files

- **Test Script**: `scripts/test-redis-cache.sh`
- **Testing Guide**: `docs/redis-cache-testing-guide.md` (this document)
- **Verification Steps**: `docs/redis-verification-steps.md`

---

## Quick Reference

### Test Cache

```bash
# Using test script
export CLASSIFICATION_SERVICE_URL=https://your-service.railway.app
./scripts/test-redis-cache.sh

# Manual test
curl -X POST https://your-service.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test", "description": "Test"}'
```

### Monitor Metrics

1. Railway Dashboard → Redis Service → Metrics
2. Check: Memory, Connections, Commands
3. Calculate cache hit rate

### Check Logs

1. Railway Dashboard → Classification Service → Logs
2. Look for: "Cache hit", "Stored in Redis cache"

---

## Success Criteria

✅ **Cache is working if**:
- Second request is 50-90% faster than first
- Redis metrics show GET/SET operations
- Cache hit rate >60% over time
- Memory usage increases as cache fills
- Service logs show cache operations

---

## Next Steps

After verifying cache functionality:

1. **Monitor Performance**: Track improvements over time
2. **Tune Configuration**: Adjust TTL values based on usage
3. **Set Up Alerts**: Monitor cache health
4. **Document Results**: Record performance improvements
5. **Optimize Further**: Consider additional caching strategies

