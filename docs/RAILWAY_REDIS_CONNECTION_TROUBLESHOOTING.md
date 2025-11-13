# Railway Redis Connection Troubleshooting

**Date**: November 13, 2025  
**Status**: 🔧 **TROUBLESHOOTING**

---

## ❌ Current Issue

**Error**: `lookup redis-cache on [fd12::10]:53: no such host`

**Service**: Risk Assessment Service  
**Redis URL**: `redis://redis-cache:6379`

---

## 🔍 Root Cause Analysis

The DNS lookup is failing, which means Railway cannot resolve the service name `redis-cache`. This could be due to:

1. **Redis service not deployed**: The `redis-cache` service might not be running
2. **Service name mismatch**: The actual service name in Railway might differ
3. **IPv6 connectivity issue**: Railway uses IPv6 for private networking
4. **Service discovery not working**: Railway's internal DNS might not be configured

---

## ✅ Verification Steps

### 1. Verify Redis Service is Deployed

**In Railway Dashboard**:
1. Navigate to your Railway project
2. Check the services list
3. Verify `redis-cache` service exists and is **running**
4. Check the service logs for any errors

**Expected**: Redis service should show as "Active" or "Running"

---

### 2. Verify Service Name Matches

**Check `railway.json`**:
```json
{
  "name": "redis-cache",
  ...
}
```

**In Railway Dashboard**:
- The service name in the dashboard should exactly match `redis-cache`
- If it's different (e.g., `redis-cache-production`), update `REDIS_URL` accordingly

---

### 3. Check Environment Variables

**In Railway Dashboard**:
1. Go to **Project Settings** → **Variables**
2. Verify `REDIS_URL=redis://redis-cache:6379` is set
3. Ensure it's marked as **shared** (applies to all services)
4. Check if there's a service-specific override

**Alternative**: Check if Railway provides a `REDIS_URL` automatically for the Redis service

---

### 4. Verify Services are in Same Project

**Critical**: Both services must be in the **same Railway project** for service discovery to work.

- Risk Assessment Service: Check project name
- Redis Service: Check project name
- They must match exactly

---

## 🔧 Potential Solutions

### Solution 1: Use Railway Private Network URL

If Railway provides a private network URL, use that instead:

```bash
# Check Railway dashboard for Redis service's private URL
# It might be something like:
REDIS_URL=redis://redis-cache.railway.internal:6379
```

**Action**: Check Railway dashboard for Redis service's internal URL

---

### Solution 2: Use Environment Variable from Railway

Railway might automatically provide a `REDIS_URL` environment variable for the Redis service. Check:

1. Go to Redis service in Railway dashboard
2. Check **Variables** tab
3. Look for automatically generated connection strings
4. Use that URL format for other services

---

### Solution 3: Verify Redis Service is Actually Running

**Check Redis Service Logs**:
```bash
# In Railway dashboard, check redis-cache service logs
# Should see Redis starting messages
```

**Expected Logs**:
```
Redis server started
Ready to accept connections
```

---

### Solution 4: Use Full Service URL (If Exposed)

If Redis service is exposed (not recommended for security), use the full URL:

```bash
REDIS_URL=redis://redis-cache-production.up.railway.app:6379
```

**Note**: This is **not recommended** for production as it exposes Redis publicly.

---

### Solution 5: Check IPv6 Support

Railway uses IPv6 for private networking. The error shows IPv6 DNS server `[fd12::10]:53`.

**Verify**:
- Go client should support IPv6 (it does by default)
- Network connectivity should work over IPv6

**If IPv6 is the issue**, Railway might need to be configured differently.

---

## 🎯 Immediate Actions

### Step 1: Verify Redis Service Status

1. Open Railway dashboard
2. Navigate to your project
3. Find `redis-cache` service
4. Check status (should be "Active" or "Running")
5. Check logs for any errors

**If service is not running**:
- Check deployment status
- Review build logs
- Verify Dockerfile is correct

---

### Step 2: Check Service Name

1. In Railway dashboard, check the exact service name
2. Compare with `railway.json` (should be `redis-cache`)
3. If different, either:
   - Update `railway.json` to match dashboard name
   - Or update `REDIS_URL` to use the dashboard name

---

### Step 3: Test Redis Connection Manually

If you have access to Railway CLI:

```bash
# Connect to Risk Assessment Service container
railway shell --service risk-assessment-service

# Test Redis connection
redis-cli -h redis-cache ping
# Should return: PONG
```

---

### Step 4: Check Railway Service Discovery

Railway might use a different service discovery mechanism:

1. Check Railway documentation for service discovery
2. Verify if services need to be in the same environment
3. Check if there are any Railway-specific networking requirements

---

## 📋 Checklist

- [ ] Redis service is deployed and running
- [ ] Service name matches exactly (`redis-cache`)
- [ ] `REDIS_URL` environment variable is set correctly
- [ ] Both services are in the same Railway project
- [ ] Both services are in the same environment (production/staging)
- [ ] Redis service logs show it's accepting connections
- [ ] No firewall or network restrictions blocking connections

---

## 🔄 Alternative: Use Railway's Redis Plugin

If Railway provides a Redis plugin/addon:

1. Check Railway's marketplace for Redis
2. Use Railway's managed Redis service
3. Railway will automatically provide connection details
4. Use the provided `REDIS_URL` environment variable

---

## 📝 Next Steps

1. **Verify Redis service is running** in Railway dashboard
2. **Check exact service name** in Railway dashboard
3. **Verify environment variables** are set correctly
4. **Test connection** manually if possible
5. **Check Railway documentation** for service discovery specifics

---

## 🆘 If Still Failing

If Redis connection still fails after all checks:

1. **Contact Railway Support**: They can verify service discovery configuration
2. **Check Railway Status**: Ensure Railway's internal networking is operational
3. **Review Railway Logs**: Check for any platform-level issues
4. **Consider Alternative**: Use Railway's managed Redis if available

---

**Last Updated**: November 13, 2025

