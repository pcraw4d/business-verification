# Redis Production Setup - Action Items

## Immediate Actions Required

### ✅ Step 1: Verify Redis Service in Railway

1. Go to Railway Dashboard
2. Navigate to your project
3. Verify Redis service exists and is running
4. Note the **exact service name** (e.g., `Redis`, `redis-cache`)

**Status**: ⏳ **Action Required**

---

### ✅ Step 2: Set Environment Variables in Classification Service

Go to **Classification Service** → **Variables** tab and add:

#### Required Variables

```bash
REDIS_ENABLED=true
REDIS_URL=${{Redis.REDIS_URL}}  # Replace 'Redis' with your Redis service name
ENABLE_WEBSITE_CONTENT_CACHE=true
```

#### Optional Variables (Recommended)

```bash
WEBSITE_CONTENT_CACHE_TTL=24h
CACHE_ENABLED=true
CACHE_TTL=5m
```

**Status**: ⏳ **Action Required**

---

### ✅ Step 3: Deploy and Verify

1. Deploy the Classification Service
2. Check logs for:
   ```
   ✅ Website content cache initialized
   Redis cache initialized for classification service
   ```
3. If you see warnings, check troubleshooting section

**Status**: ⏳ **Action Required**

---

## How to Set Variables in Railway

### Method 1: Railway Dashboard (Easiest)

1. **Open Railway Dashboard**
   - Go to your project
   - Click **Classification Service**
   - Click **Variables** tab

2. **Add Variables**
   - Click **"+ New Variable"** or **"Add Variable"**
   - Add each variable:
     - Name: `REDIS_ENABLED`
     - Value: `true`
     - Click **"Add"**
   - Repeat for other variables

3. **For REDIS_URL**:
   - **Option A** (Recommended): Use interpolation
     - Name: `REDIS_URL`
     - Value: `${{Redis.REDIS_URL}}` (replace `Redis` with your Redis service name)
   - **Option B**: Manual copy
     - Go to Redis service → Variables tab
     - Copy `REDIS_URL` value
     - Paste into Classification Service variables

### Method 2: Railway CLI

```bash
# Set variables via CLI
railway variables set REDIS_ENABLED=true --service classification-service
railway variables set REDIS_URL='${{Redis.REDIS_URL}}' --service classification-service
railway variables set ENABLE_WEBSITE_CONTENT_CACHE=true --service classification-service
```

---

## Verification Checklist

After deployment, verify:

- [ ] Logs show: `✅ Website content cache initialized`
- [ ] Logs show: `Redis cache initialized for classification service`
- [ ] No Redis connection errors in logs
- [ ] Test request shows `X-Cache: HIT` or `X-Cache: MISS` header
- [ ] Second identical request shows `X-Cache: HIT`
- [ ] Redis metrics visible in Railway dashboard

---

## Troubleshooting

### If Redis Connection Fails

1. **Check Variable Names**
   - Verify `REDIS_ENABLED=true` (exact spelling)
   - Verify `REDIS_URL` is set correctly

2. **Check Redis Service Name**
   - Verify Redis service name matches in interpolation
   - Example: If service is `redis-cache`, use `${{redis-cache.REDIS_URL}}`

3. **Check Redis Service Status**
   - Verify Redis service is running
   - Check Redis service logs for errors

4. **Try Manual URL**
   - If interpolation fails, manually copy `REDIS_URL` from Redis service
   - Set it directly in Classification Service

---

## Expected Results

Once configured correctly:

- ✅ **Faster Response Times**: Cached requests 50-90% faster
- ✅ **Reduced Load**: Fewer external HTTP requests
- ✅ **Better Scalability**: Distributed caching across instances
- ✅ **Cost Savings**: Reduced external API calls

---

## Files

- **Detailed Guide**: `docs/redis-production-setup-railway.md`
- **Quick Checklist**: `docs/redis-production-checklist.md`
- **Action Items**: `docs/redis-production-action-items.md` (this document)

