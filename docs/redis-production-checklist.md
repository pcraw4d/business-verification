# Redis Production Checklist for Railway

## Quick Checklist

Use this checklist to ensure Redis is properly configured in production on Railway.

---

## Pre-Deployment Checklist

### ✅ Railway Redis Service

- [ ] Redis service added to Railway project
- [ ] Redis service is running and healthy
- [ ] Redis connection URL is available in Railway dashboard

### ✅ Environment Variables

- [ ] `REDIS_ENABLED=true` is set
- [ ] `REDIS_URL` is set (usually auto-provided by Railway)
- [ ] `ENABLE_WEBSITE_CONTENT_CACHE=true` is set
- [ ] `WEBSITE_CONTENT_CACHE_TTL=24h` is set (optional, default: 24h)

### ✅ Code Verification

- [ ] Code supports Redis (✅ Already implemented)
- [ ] Redis client initialization is in place (✅ Already implemented)
- [ ] Error handling for Redis failures (✅ Already implemented)

---

## Post-Deployment Verification

### ✅ Log Verification

After deployment, check logs for:

- [ ] `✅ Website content cache initialized` message appears
- [ ] `Redis cache initialized for classification service` message appears
- [ ] No Redis connection errors

### ✅ Functionality Verification

- [ ] Make a classification request
- [ ] Check response headers for `X-Cache: HIT` or `X-Cache: MISS`
- [ ] Make identical request - should see `X-Cache: HIT` on second request
- [ ] Response time should be faster on cached requests

### ✅ Redis Monitoring

- [ ] Check Redis service metrics in Railway dashboard
- [ ] Monitor Redis memory usage
- [ ] Check connection count
- [ ] Verify cache operations are working

---

## Troubleshooting Checklist

If Redis is not working:

- [ ] Verify `REDIS_ENABLED=true` is set
- [ ] Verify `REDIS_URL` is correct and accessible
- [ ] Check Redis service is running
- [ ] Verify network connectivity between services
- [ ] Check logs for connection errors
- [ ] Verify Redis URL format is correct

---

## Environment Variables to Set in Railway

### Required

```bash
REDIS_ENABLED=true
REDIS_URL=<railway-redis-url>
ENABLE_WEBSITE_CONTENT_CACHE=true
```

### Optional

```bash
WEBSITE_CONTENT_CACHE_TTL=24h
CACHE_ENABLED=true
CACHE_TTL=5m
```

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

If you see failure messages, Redis caching is disabled but service continues with in-memory cache.

---

## Performance Expectations

Once Redis is enabled:

- **Cache Hit Rate**: Should be >60% for website content
- **Response Time**: Cached requests should be 50-90% faster
- **External Requests**: Should decrease significantly
- **Memory Usage**: Monitor Redis memory usage

---

## Files

- **Setup Guide**: `docs/redis-production-setup-railway.md`
- **Checklist**: `docs/redis-production-checklist.md` (this document)

