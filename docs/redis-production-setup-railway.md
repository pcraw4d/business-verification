# Redis Production Setup for Railway - Classification Service

## Overview

This guide explains how to enable Redis caching in production on Railway for the classification service. Since you already have Redis in Railway, we just need to configure the classification service to use it.

---

## Current Status

✅ **Redis Service**: Already exists in Railway  
✅ **Code Support**: Already implemented in classification service  
⏳ **Configuration**: Need to set environment variables

---

## Quick Setup Steps

### Step 1: Identify Your Redis Service Name

1. Go to Railway Dashboard → Your Project
2. Look at your services list
3. Find your Redis service and note its **exact name** (e.g., `Redis`, `redis-cache`, `redis-service`)
4. This name is needed for variable interpolation

**Common Railway Redis Variable Names** (in the Redis service):
- `REDIS_URL` - Full connection URL
- `REDISHOST` - Redis hostname  
- `REDISPORT` - Redis port (usually 6379)
- `REDISPASSWORD` - Redis password

**Important**: These variables exist in the Redis service, but need to be shared with Classification Service.

---

### Step 2: Configure Classification Service

Go to **Classification Service** → **Variables** tab and add:

#### Option A: Use Railway Variable Interpolation (Recommended) ⭐

This automatically syncs with Redis service variables:

```bash
# Enable Redis
REDIS_ENABLED=true

# Use Railway's variable interpolation to reference Redis service
# Replace 'Redis' with your actual Redis service name
REDIS_URL=${{Redis.REDIS_URL}}

# Enable website content cache
ENABLE_WEBSITE_CONTENT_CACHE=true
```

**How to find your Redis service name**:
1. Go to Railway Dashboard → Your Project
2. Look at services list
3. Find Redis service (might be named `Redis`, `redis`, `redis-cache`, etc.)
4. Use that exact name in the interpolation: `${{YourRedisServiceName.REDIS_URL}}`

**Example**: If your Redis service is named `redis-cache`:
```bash
REDIS_URL=${{redis-cache.REDIS_URL}}
```

#### Option B: Manually Set Redis URL

If interpolation doesn't work, manually copy the Redis URL:

```bash
# Enable Redis
REDIS_ENABLED=true

# Set Redis URL (copy from Redis service → Variables tab)
REDIS_URL=redis://default:password@redis.railway.internal:6379

# Enable website content cache
ENABLE_WEBSITE_CONTENT_CACHE=true
```

---

### Step 3: Optional Configuration

Add these for fine-tuning:

```bash
# Website content cache TTL (default: 24h)
WEBSITE_CONTENT_CACHE_TTL=24h

# Enable classification result cache (uses Redis if enabled)
CACHE_ENABLED=true
CACHE_TTL=5m
```

---

## Required Environment Variables Summary

### Minimum Required

```bash
REDIS_ENABLED=true
REDIS_URL=${{Redis.REDIS_URL}}  # or manual URL
ENABLE_WEBSITE_CONTENT_CACHE=true
```

### Recommended

```bash
REDIS_ENABLED=true
REDIS_URL=${{Redis.REDIS_URL}}
ENABLE_WEBSITE_CONTENT_CACHE=true
WEBSITE_CONTENT_CACHE_TTL=24h
CACHE_ENABLED=true
CACHE_TTL=5m
```

---

## How Railway Variable Interpolation Works

Railway allows services to reference variables from other services using:

```
${{ServiceName.VARIABLE_NAME}}
```

**Example**:
- Redis service name: `Redis`
- Variable in Redis service: `REDIS_URL`
- Reference in Classification Service: `${{Redis.REDIS_URL}}`

**Benefits**:
- ✅ Automatically stays in sync
- ✅ No manual updates needed
- ✅ Railway handles the reference

---

## Step-by-Step Railway Dashboard Instructions

### Method 1: Using Variable Interpolation (Recommended)

1. **Open Railway Dashboard**
   - Navigate to your project
   - Click on **Classification Service**

2. **Go to Variables Tab**
   - Click **"Variables"** tab
   - Click **"+ New Variable"** or **"Add Variable"**

3. **Add Redis Variables**
   
   **Variable 1**:
   - **Name**: `REDIS_ENABLED`
   - **Value**: `true`
   - **Scope**: Service (or Shared if you want all services to use Redis)
   
   **Variable 2**:
   - **Name**: `REDIS_URL`
   - **Value**: `${{Redis.REDIS_URL}}` (replace `Redis` with your Redis service name)
   - **Scope**: Service
   
   **Variable 3**:
   - **Name**: `ENABLE_WEBSITE_CONTENT_CACHE`
   - **Value**: `true`
   - **Scope**: Service

4. **Deploy**
   - Click **"Deploy"** or wait for auto-deploy
   - Check logs for Redis connection success

### Method 2: Manual Copy (If Interpolation Doesn't Work)

1. **Get Redis URL**
   - Go to **Redis Service** → **Variables** tab
   - Copy the value of `REDIS_URL` (or construct from `REDISHOST`, `REDISPORT`, `REDISPASSWORD`)

2. **Set in Classification Service**
   - Go to **Classification Service** → **Variables** tab
   - Add `REDIS_URL` with the copied value
   - Add `REDIS_ENABLED=true`
   - Add `ENABLE_WEBSITE_CONTENT_CACHE=true`

3. **Deploy**
   - Deploy the service
   - Check logs for connection

---

## Verification

### After Deployment

Check service logs for:

**Success Messages**:
```
✅ Website content cache initialized
Redis cache initialized for classification service
```

**Failure Messages** (Non-critical, service continues):
```
⚠️ Failed to connect to Redis for website content cache, caching disabled
Using in-memory cache only (Redis not enabled or URL not provided)
```

### Test Cache Functionality

1. **Make a Classification Request**
   ```bash
   curl -X POST https://your-service.railway.app/classify \
     -H "Content-Type: application/json" \
     -d '{"business_name": "Test", "description": "Test"}'
   ```

2. **Check Response Headers**
   - Look for `X-Cache: HIT` or `X-Cache: MISS`
   - Second identical request should show `X-Cache: HIT`

3. **Monitor Performance**
   - Cached requests should be faster
   - Check Redis metrics in Railway dashboard

---

## Troubleshooting

### Redis Not Connecting

**Check**:
1. ✅ `REDIS_ENABLED=true` is set
2. ✅ `REDIS_URL` is correct (check Redis service → Variables)
3. ✅ Redis service is running
4. ✅ Services are in same Railway project
5. ✅ Variable interpolation syntax is correct: `${{ServiceName.VARIABLE}}`

**Common Issues**:

**Issue**: `Failed to parse Redis URL`
- **Solution**: Check `REDIS_URL` format is correct: `redis://[password@]host:port[/db]`

**Issue**: `Connection refused`
- **Solution**: Verify Redis service name in interpolation matches actual service name

**Issue**: Variable interpolation not working
- **Solution**: Use manual copy method instead

### Finding Your Redis Service Name

1. Go to Railway Dashboard → Your Project
2. Look at the services list
3. Find your Redis service
4. The name shown is what you use in interpolation: `${{YourRedisServiceName.REDIS_URL}}`

---

## What Gets Cached

### Website Content Cache (When Enabled)

- Scraped website content
- Page text, titles, keywords
- Structured data
- TTL: 24 hours (configurable)

**Benefits**:
- Reduces redundant HTTP requests
- Faster response times
- Lower external API costs

### Classification Result Cache (When Enabled)

- Complete classification results
- Industry, codes, confidence scores
- TTL: 5 minutes (configurable)

**Benefits**:
- Instant responses for identical requests
- Reduced processing load
- Better user experience

---

## Performance Expectations

Once Redis is enabled:

- **Cache Hit Rate**: Should be >60% for website content
- **Response Time**: Cached requests 50-90% faster
- **External Requests**: Significant reduction
- **Memory Usage**: Monitor Redis memory in Railway dashboard

---

## Production Checklist

- [ ] Redis service exists and is running in Railway
- [ ] `REDIS_ENABLED=true` is set in Classification Service
- [ ] `REDIS_URL` is set (via interpolation or manual)
- [ ] `ENABLE_WEBSITE_CONTENT_CACHE=true` is set
- [ ] Service deployed and logs show Redis connection success
- [ ] Cache operations verified (check `X-Cache` headers)
- [ ] Redis metrics monitored in Railway dashboard

---

## Next Steps After Setup

1. **Monitor Cache Hit Rates**: Check logs and metrics
2. **Tune TTL Values**: Adjust based on usage patterns
3. **Monitor Redis Memory**: Ensure adequate capacity
4. **Set Up Alerts**: Monitor Redis health and memory usage
5. **Performance Testing**: Measure improvement in response times

---

## Files

- **Detailed Setup**: `docs/redis-production-setup-railway.md` (this document)
- **Quick Checklist**: `docs/redis-production-checklist.md`
- **Environment Variables**: `docs/railway-redis-env-variables.md`
- **Code Implementation**: `services/classification-service/cmd/main.go:80-106`

---

## Quick Reference

### Railway Dashboard Path

1. **Project** → **Classification Service** → **Variables**
2. Add:
   - `REDIS_ENABLED=true`
   - `REDIS_URL=${{Redis.REDIS_URL}}` (or manual URL)
   - `ENABLE_WEBSITE_CONTENT_CACHE=true`
3. **Deploy** → Check logs

### Expected Log Output

```
✅ Website content cache initialized
Redis cache initialized for classification service
```

---

## Support

If issues persist:
1. Check Railway Redis service logs
2. Verify variable names match exactly
3. Check service connectivity in Railway dashboard
4. Review classification service logs for specific errors
