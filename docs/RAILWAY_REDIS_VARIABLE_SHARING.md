# Railway Redis Variable Sharing Guide

**Date**: November 13, 2025  
**Status**: ✅ **REQUIRED SETUP STEP**

---

## 🔍 Issue

When you add Railway's Redis plugin, the environment variables (`REDISHOST`, `REDISPORT`, `REDISPASSWORD`, `REDIS_URL`) are **only available in the Redis service itself**, not automatically shared with other services.

---

## ✅ Solution Options

You have **two options** to make Redis variables available to other services:

### Option 1: Use Railway Variable Interpolation (Recommended) ⭐

**Best for**: Automatic synchronization, no manual updates needed

**Steps**:
1. Go to **Project Settings** → **Variables**
2. For each service that needs Redis, add these variables:

```
REDISHOST=${{Redis.REDISHOST}}
REDISPORT=${{Redis.REDISPORT}}
REDISPASSWORD=${{Redis.REDISPASSWORD}}
REDIS_URL=${{Redis.REDIS_URL}}
```

**Note**: Replace `Redis` with the actual name of your Redis service if different.

**Benefits**:
- ✅ Automatically stays in sync with Redis service
- ✅ No manual updates needed if Redis variables change
- ✅ Railway handles the reference

---

### Option 2: Manually Copy as Shared Variables

**Best for**: Simple setup, but requires manual updates

**Steps**:
1. Go to **Redis service** → **Variables** tab
2. Copy the values of:
   - `REDISHOST`
   - `REDISPORT`
   - `REDISPASSWORD`
   - `REDIS_URL`
3. Go to **Project Settings** → **Variables**
4. Add each variable as a **shared variable** with the copied values

**Benefits**:
- ✅ Simple to set up
- ❌ Requires manual updates if Redis variables change

---

## 📋 Step-by-Step: Option 1 (Recommended)

### For Risk Assessment Service

1. **Open Railway Dashboard**
   - Navigate to your project
   - Go to **Project Settings** → **Variables**

2. **Add Redis Variable References**
   - Click **"New Variable"** or **"Add Variable"**
   - Add each of these (replace `Redis` with your Redis service name if different):

   ```
   Variable Name: REDISHOST
   Value: ${{Redis.REDISHOST}}
   Scope: Shared (or specific to risk-assessment-service)
   ```

   ```
   Variable Name: REDISPORT
   Value: ${{Redis.REDISPORT}}
   Scope: Shared (or specific to risk-assessment-service)
   ```

   ```
   Variable Name: REDISPASSWORD
   Value: ${{Redis.REDISPASSWORD}}
   Scope: Shared (or specific to risk-assessment-service)
   ```

   ```
   Variable Name: REDIS_URL
   Value: ${{Redis.REDIS_URL}}
   Scope: Shared (or specific to risk-assessment-service)
   ```

3. **Verify Service Name**
   - Check the exact name of your Redis service in Railway dashboard
   - If it's not `Redis`, replace `Redis` in the interpolation syntax
   - Example: If service is named `redis-cache`, use `${{redis-cache.REDISHOST}}`

4. **Redeploy Services**
   - Railway will automatically redeploy services with the new variables
   - Or manually trigger a redeploy if needed

---

## 📋 Step-by-Step: Option 2 (Alternative)

### Copy Values Manually

1. **Get Redis Variables**
   - Go to **Redis service** → **Variables** tab
   - Note the values of:
     - `REDISHOST` (e.g., `redis.railway.internal` or similar)
     - `REDISPORT` (usually `6379`)
     - `REDISPASSWORD` (long random string)
     - `REDIS_URL` (full connection string)

2. **Add as Shared Variables**
   - Go to **Project Settings** → **Variables**
   - Add each variable:
     - Name: `REDISHOST`, Value: (copied value)
     - Name: `REDISPORT`, Value: (copied value)
     - Name: `REDISPASSWORD`, Value: (copied value)
     - Name: `REDIS_URL`, Value: (copied value)
   - Mark as **Shared** (applies to all services)

3. **Redeploy Services**
   - Services will pick up the new variables

---

## 🔍 Verification

### Check Variables are Available

**In Railway Dashboard**:
1. Go to **Risk Assessment Service** → **Variables** tab
2. Verify these variables are present:
   - `REDISHOST`
   - `REDISPORT`
   - `REDISPASSWORD`
   - `REDIS_URL`

**In Service Logs** (after redeploy):
```
🔧 Initializing Redis cache using Railway Redis plugin redis_host: "..." redis_port: "6379" has_password: true
✅ Risk Assessment Service Redis cache initialized successfully (Railway plugin)
```

---

## ⚠️ Important Notes

### Service Name in Interpolation

The service name in `${{Redis.REDISHOST}}` must match the **exact name** of your Redis service in Railway.

**To find the service name**:
1. Go to Railway dashboard
2. Look at the Redis service name
3. Use that exact name in the interpolation

**Examples**:
- If service is named `Redis`: `${{Redis.REDISHOST}}`
- If service is named `redis`: `${{redis.REDISHOST}}`
- If service is named `redis-cache`: `${{redis-cache.REDISHOST}}`

### Variable Scope

- **Shared**: Available to all services (recommended)
- **Service-specific**: Only available to that service

For Redis, **shared** is recommended since multiple services may need it.

---

## 🎯 Quick Reference

### Variable Interpolation Syntax

```
${{ServiceName.VARIABLE_NAME}}
```

**Examples**:
```
REDISHOST=${{Redis.REDISHOST}}
REDISPORT=${{Redis.REDISPORT}}
REDISPASSWORD=${{Redis.REDISPASSWORD}}
REDIS_URL=${{Redis.REDIS_URL}}
```

### Manual Copy Alternative

If interpolation doesn't work, manually copy values from Redis service variables to project shared variables.

---

## 📝 Checklist

- [ ] Redis plugin added and deployed
- [ ] Redis service name identified
- [ ] Variables added using interpolation OR manually copied
- [ ] Variables verified in service's Variables tab
- [ ] Services redeployed
- [ ] Redis connection verified in service logs

---

**Last Updated**: November 13, 2025

