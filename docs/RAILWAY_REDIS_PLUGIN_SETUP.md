# Railway Redis Plugin Setup Guide

**Date**: November 13, 2025  
**Status**: ✅ **MIGRATING TO RAILWAY REDIS PLUGIN**

---

## 🎯 Overview

We're migrating from a custom Redis service to Railway's managed Redis plugin. Railway's Redis plugin:
- Automatically provides connection details via environment variables
- Handles service discovery automatically
- Provides better reliability and management
- No need for custom Dockerfile or service configuration

---

## 📋 Setup Steps

### Step 1: Add Redis Plugin in Railway Dashboard

1. **Open Railway Dashboard**
   - Navigate to your Railway project
   - Click the **`+ New`** button
   - Select **`Database`**
   - Choose **`Redis`** from the list

2. **Wait for Deployment**
   - Railway will automatically deploy the Redis service
   - This may take a few minutes

3. **Verify Redis Service**
   - Check that the Redis service appears in your services list
   - Status should be "Active" or "Running"

---

### Step 2: Railway Provides Environment Variables (But They're Not Shared)

Railway's Redis plugin automatically provides these environment variables **in the Redis service**:

- **`REDISHOST`** - Redis hostname
- **`REDISPORT`** - Redis port (usually 6379)
- **`REDISPASSWORD`** - Redis password
- **`REDIS_URL`** - Complete Redis connection URL

**⚠️ Important**: These variables are **NOT automatically shared** with other services. You need to manually share them (see Step 3).

---

### Step 3: Share Redis Variables with Other Services

**⚠️ Required**: Redis variables are only in the Redis service. You must share them with other services.

**Option A: Use Variable Interpolation (Recommended)** ⭐

1. Go to **Project Settings** → **Variables**
2. Add these variables using Railway's interpolation syntax:

   ```
   REDISHOST=${{Redis.REDISHOST}}
   REDISPORT=${{Redis.REDISPORT}}
   REDISPASSWORD=${{Redis.REDISPASSWORD}}
   REDIS_URL=${{Redis.REDIS_URL}}
   ```

   **Note**: Replace `Redis` with your actual Redis service name if different.

3. Mark as **Shared** (applies to all services)

**Option B: Manually Copy Values**

1. Go to **Redis service** → **Variables** tab
2. Copy the values of `REDISHOST`, `REDISPORT`, `REDISPASSWORD`, `REDIS_URL`
3. Go to **Project Settings** → **Variables**
4. Add each as a shared variable with the copied values

**See**: `docs/RAILWAY_REDIS_VARIABLE_SHARING.md` for detailed instructions.

---

### Step 4: Remove Custom Redis Service

**In Railway Dashboard**:
1. Find the `redis-cache` service (custom service)
2. Delete or stop the service
3. This prevents conflicts with Railway's Redis plugin

**In Code**:
- The `redis-cache` service has been removed from `railway.json`
- Redis is now configured in the `databases` section

---

## 🔧 Code Changes

### Updated Configuration

**`railway.json`**:
```json
{
  "databases": [
    {
      "name": "postgres",
      "type": "postgresql",
      "version": "15"
    },
    {
      "name": "redis",
      "type": "redis"
    }
  ]
}
```

**Removed**: Custom `redis-cache` service from `services` array

---

### Updated Application Code

**`services/risk-assessment-service/cmd/main.go`**:
- Now checks for Railway's Redis plugin environment variables first:
  - `REDISHOST` and `REDISPORT` (preferred)
  - Falls back to `REDIS_URL` if plugin variables not available
- Automatically uses Railway-provided connection details

**Code Flow**:
1. Check for `REDISHOST` and `REDISPORT` (Railway plugin)
2. If available, use those to construct connection
3. Otherwise, fall back to `REDIS_URL`
4. Initialize Redis cache with appropriate configuration

---

## ✅ Expected Behavior

### After Setup

**Service Logs Should Show**:
```
🔧 Initializing Redis cache using Railway Redis plugin redis_host: "..." redis_port: "6379" has_password: true
✅ Risk Assessment Service Redis cache initialized successfully (Railway plugin) redis_host: "..." redis_port: "6379" pool_size: 50
```

**If Redis Plugin Not Available**:
```
⚠️  Redis not configured - running without Redis cache
```

---

## 🔍 Verification

### 1. Check Environment Variables

**In Railway Dashboard**:
- Go to any service (e.g., Risk Assessment Service)
- Check **Variables** tab
- Verify `REDISHOST`, `REDISPORT`, `REDISPASSWORD`, and `REDIS_URL` are present

### 2. Check Service Logs

**After Deployment**:
- Review Risk Assessment Service logs
- Look for Redis initialization messages
- Verify successful connection

### 3. Test Redis Connection

**If you have Railway CLI access**:
```bash
# Connect to a service container
railway shell --service risk-assessment-service

# Test Redis connection (if redis-cli is available)
redis-cli -h $REDISHOST -p $REDISPORT -a $REDISPASSWORD ping
# Should return: PONG
```

---

## 🚨 Troubleshooting

### Issue: Environment Variables Not Available

**Symptoms**: Service logs show "Redis not configured"

**Solution**:
1. Verify Redis plugin is added in Railway dashboard
2. Check that Redis service is running
3. Wait a few minutes for environment variables to propagate
4. Redeploy services if needed

---

### Issue: Connection Still Fails

**Symptoms**: "Failed to initialize Redis cache" error

**Solution**:
1. Verify `REDISHOST` and `REDISPORT` are set correctly
2. Check `REDISPASSWORD` is not empty
3. Verify Redis service is running in Railway dashboard
4. Check service logs for specific error messages

---

### Issue: Services Can't Find Redis

**Symptoms**: DNS lookup errors

**Solution**:
- Railway's Redis plugin handles service discovery automatically
- No need to configure service names
- Use the provided environment variables directly

---

## 📊 Benefits of Railway Redis Plugin

✅ **Automatic Service Discovery**: No need to configure service names  
✅ **Managed Service**: Railway handles updates and maintenance  
✅ **Automatic Environment Variables**: Connection details provided automatically  
✅ **Better Reliability**: Railway manages the Redis service  
✅ **Simplified Configuration**: No custom Dockerfile needed  

---

## 🔄 Migration Checklist

- [ ] Add Redis plugin in Railway dashboard
- [ ] Verify Redis service is running
- [ ] Check environment variables are available
- [ ] Remove/stop custom `redis-cache` service
- [ ] Update `railway.json` (already done)
- [ ] Deploy updated code
- [ ] Verify Redis connection in service logs
- [ ] Test Redis functionality

---

## 📝 Notes

- **No Manual Configuration Needed**: Railway handles everything automatically
- **Environment Variables**: Automatically shared with all services
- **Service Discovery**: Handled by Railway, no DNS configuration needed
- **Backward Compatible**: Code falls back to `REDIS_URL` if plugin variables not available

---

**Last Updated**: November 13, 2025

