# Railway Redis Environment Variables

## Quick Reference

Copy these environment variables to your Railway Classification Service → Variables tab.

---

## Required Variables

```bash
# Enable Redis caching
REDIS_ENABLED=true

# Redis URL (Railway provides this automatically when Redis service is linked)
# Check Railway Redis service → Connect tab for exact URL
REDIS_URL=redis://default:password@redis.railway.internal:6379

# Enable website content cache (recommended)
ENABLE_WEBSITE_CONTENT_CACHE=true
```

---

## Optional Variables

```bash
# Website content cache TTL (default: 24h)
WEBSITE_CONTENT_CACHE_TTL=24h

# Enable classification result cache
CACHE_ENABLED=true

# Classification result cache TTL (default: 5m)
CACHE_TTL=5m
```

---

## How to Set in Railway

1. Go to Railway Dashboard
2. Select your **Classification Service**
3. Click **"Variables"** tab
4. Click **"+ New Variable"**
5. Add each variable:
   - **Name**: `REDIS_ENABLED`
   - **Value**: `true`
6. Repeat for other variables
7. **Deploy** the service

---

## Railway Auto-Provided Variables

When you link a Redis service to your Classification Service, Railway automatically provides:

- `REDIS_URL` (or similar, check Redis service → Connect tab)
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_PASSWORD` (if required)

**Note**: Variable names may vary. Check your Railway Redis service dashboard for exact names.

---

## Verification

After setting variables and deploying, check logs for:

```
✅ Website content cache initialized
Redis cache initialized for classification service
```

If you see warnings, Redis is not connected but service continues with in-memory cache.

---

## Files

- **Detailed Setup**: `docs/redis-production-setup-railway.md`
- **Checklist**: `docs/redis-production-checklist.md`
- **Quick Reference**: `docs/railway-redis-env-variables.md` (this document)

