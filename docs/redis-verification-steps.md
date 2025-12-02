# Redis Verification Steps - After Railway Configuration

## Overview

After configuring Redis in Railway, follow these steps to verify everything is working correctly.

---

## Step 1: Verify Environment Variables

### In Railway Dashboard

1. Go to **Classification Service** → **Variables** tab
2. Verify these variables are set:
   - ✅ `REDIS_ENABLED=true`
   - ✅ `REDIS_URL` is set (either via interpolation or manual)
   - ✅ `ENABLE_WEBSITE_CONTENT_CACHE=true`

### Check Variable Values

If using interpolation:
- `REDIS_URL=${{Redis.REDIS_URL}}` (or your Redis service name)

If manually set:
- `REDIS_URL=redis://default:password@host:6379` (actual connection string)

---

## Step 2: Check Service Logs

### After Deployment

1. Go to **Classification Service** → **Logs** tab in Railway
2. Look for these **success messages**:

```
✅ Website content cache initialized
Redis cache initialized for classification service
```

### If You See Warnings

**Non-critical warnings** (service continues with in-memory cache):
```
⚠️ Failed to connect to Redis for website content cache, caching disabled
Using in-memory cache only (Redis not enabled or URL not provided)
```

**If you see warnings, check**:
1. `REDIS_ENABLED=true` is set correctly
2. `REDIS_URL` is correct and accessible
3. Redis service is running in Railway
4. Network connectivity between services

---

## Step 3: Test Cache Functionality

### Make a Test Classification Request

```bash
# Replace with your actual service URL
curl -X POST https://your-classification-service.railway.app/classify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_name": "Test Company",
    "description": "A test business for verification",
    "website": "https://example.com"
  }'
```

### Check Response Headers

Look for cache-related headers in the response:
- `X-Cache: HIT` - Content was served from cache
- `X-Cache: MISS` - Content was fetched fresh

### Test Cache Hit

1. **First Request**: Should show `X-Cache: MISS` (or no header)
2. **Second Identical Request**: Should show `X-Cache: HIT`
3. **Response Time**: Second request should be faster

---

## Step 4: Monitor Redis Metrics

### In Railway Dashboard

1. Go to **Redis Service** → **Metrics** tab
2. Monitor:
   - **Memory Usage**: Should increase as cache fills
   - **Connections**: Should show active connections
   - **Commands**: Should show cache operations (GET, SET)

### Expected Behavior

- Memory usage increases gradually as content is cached
- Connection count shows active connections from Classification Service
- Command count shows cache read/write operations

---

## Step 5: Verify Performance Improvements

### Before Redis (Baseline)

- Response time: ~6-8 seconds for new classifications
- External requests: Multiple HTTP requests per classification
- Cache hit rate: 0% (no caching)

### After Redis (Expected)

- **Cached requests**: 0.5-2 seconds (50-90% faster)
- **External requests**: Reduced significantly
- **Cache hit rate**: Should be >60% for website content

### Measure Performance

1. Make several classification requests
2. Note response times
3. Check cache hit rate in logs
4. Monitor Redis memory usage

---

## Troubleshooting

### Issue: Redis Not Connecting

**Symptoms**:
- Logs show: `⚠️ Failed to connect to Redis`
- No cache operations in Redis metrics

**Solutions**:
1. Verify `REDIS_ENABLED=true` is set
2. Check `REDIS_URL` format is correct
3. Verify Redis service is running
4. Check variable interpolation syntax: `${{ServiceName.VARIABLE}}`
5. Try manual `REDIS_URL` if interpolation fails

### Issue: Cache Not Working

**Symptoms**:
- No `X-Cache` headers in responses
- Every request processes fully
- No cache operations in Redis metrics

**Solutions**:
1. Verify `ENABLE_WEBSITE_CONTENT_CACHE=true`
2. Check Redis connection in logs
3. Verify `REDIS_URL` is accessible
4. Check service logs for initialization messages

### Issue: Variable Interpolation Not Working

**Symptoms**:
- `REDIS_URL` shows literal `${{Redis.REDIS_URL}}` instead of actual URL

**Solutions**:
1. Verify Redis service name matches exactly
2. Check interpolation syntax: `${{ServiceName.VARIABLE}}`
3. Try manual copy method instead
4. Go to Redis service → Variables → Copy `REDIS_URL` value manually

---

## Verification Checklist

After configuration, verify:

- [ ] Environment variables set correctly in Railway
- [ ] Service logs show Redis connection success
- [ ] Test request completes successfully
- [ ] Response headers show cache status
- [ ] Second identical request shows cache hit
- [ ] Redis metrics show active connections
- [ ] Redis metrics show memory usage
- [ ] Performance improvements observed

---

## Next Steps

Once verified:

1. **Monitor Cache Hit Rates**: Track over time
2. **Tune TTL Values**: Adjust based on usage patterns
3. **Set Up Alerts**: Monitor Redis health
4. **Performance Testing**: Measure improvements
5. **Document Results**: Record performance gains

---

## Expected Log Messages

### Success

```
✅ Website content cache initialized
Redis cache initialized for classification service
```

### Failure (Non-Critical)

```
⚠️ Failed to connect to Redis for website content cache, caching disabled
Using in-memory cache only (Redis not enabled or URL not provided)
```

**Note**: If you see failure messages, the service continues with in-memory caching, but Redis benefits are not available.

---

## Files

- **Setup Guide**: `docs/redis-production-setup-railway.md`
- **Verification Steps**: `docs/redis-verification-steps.md` (this document)
- **Troubleshooting**: See troubleshooting sections in setup guide

