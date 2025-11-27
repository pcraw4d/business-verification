# Python ML Service 502 Error Fix

**Date**: November 27, 2025  
**Issue**: Python ML Service returns 502 Bad Gateway  
**Status**: 🔧 **Troubleshooting Required**

---

## Issue Summary

The classification service is trying to connect to the Python ML service but getting a 502 Bad Gateway error:

```
⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
error: failed to connect to Python ML service: ping returned status 502
```

**URL being used:**
```
https://python-ml-service-production-a6b8.up.railway.app
```

---

## Root Cause Analysis

A 502 Bad Gateway error typically means:

1. **Service is not running** - The Python ML service container crashed or failed to start
2. **Service is still starting** - Models are loading (can take 60-90 seconds)
3. **Service crashed during startup** - Error in initialization code
4. **Port mismatch** - Unlikely on Railway (handles automatically), but possible
5. **Network/routing issue** - Railway routing problem

---

## Troubleshooting Steps

### Step 1: Check Python ML Service Logs

**In Railway Dashboard:**
1. Go to Railway Dashboard → Your Project
2. Click on `python-ml-service`
3. Go to "Logs" tab
4. Look for:
   - Service startup messages
   - Model loading messages
   - Any error messages
   - Crash/panic messages

**Expected logs (successful startup):**
```
🚀 Starting Python ML Service...
📱 Device: cpu
📚 Models will be loaded lazily on first request
🌐 Starting server on port 8000
```

**Expected logs (when models load):**
```
🚀 Starting lazy model loading...
📥 Loading DistilBART classifier (this may take 60-90 seconds)...
✅ DistilBART classifier initialized with quantization: True
✅ All models loaded successfully
```

**Error indicators:**
- `ImportError` - Missing dependencies
- `ModuleNotFoundError` - Missing Python modules
- `NameError` - Code error
- `ConnectionError` - Network issue
- `MemoryError` - Out of memory
- `TimeoutError` - Operation timed out

### Step 2: Check Service Status

**In Railway Dashboard:**
1. Go to `python-ml-service` → "Settings"
2. Check "Service Status"
3. Verify:
   - Service is "Running" (not "Stopped" or "Crashed")
   - Last deployment was successful
   - Health check is passing

### Step 3: Test Health Endpoint Directly

```bash
# Test the health endpoint
curl https://python-ml-service-production-a6b8.up.railway.app/health

# Test the ping endpoint (used by classification service)
curl https://python-ml-service-production-a6b8.up.railway.app/ping

# Test root endpoint
curl https://python-ml-service-production-a6b8.up.railway.app/
```

**Expected responses:**
- `/health`: `{"status":"healthy","timestamp":"...","service":"Python ML Service","version":"2.0.0"}`
- `/ping`: `{"status":"ok","message":"Python ML Service is running"}`
- `/`: `{"status":"ok","service":"Python ML Service","version":"2.0.0"}`

**If all return 502:**
- Service is definitely not running
- Check Railway logs for crash/error

**If some work but `/ping` doesn't:**
- Check if `/ping` endpoint exists (it should)
- Check for routing issues

### Step 4: Check Port Configuration

**Railway automatically handles ports**, but verify:

1. **Python ML Service:**
   - Railway sets `PORT` environment variable automatically
   - Service should use `PORT` from environment (defaults to 8000)
   - Check `python_ml_service/start.sh` uses `$PORT`

2. **Classification Service:**
   - Uses the public URL (no port needed)
   - Railway routes automatically

**Verify in Railway:**
- Go to `python-ml-service` → "Settings" → "Variables"
- Check if `PORT` is set (Railway sets this automatically)
- Should be `8000` or Railway's assigned port

### Step 5: Check Service Deployment

**Verify deployment:**
1. Go to Railway Dashboard → `python-ml-service`
2. Check "Deployments" tab
3. Verify:
   - Latest deployment is "Active"
   - Build was successful
   - No deployment errors

**If deployment failed:**
- Check build logs
- Verify Dockerfile is correct
- Check for dependency issues

---

## Common Fixes

### Fix 1: Service Crashed During Startup

**Symptoms:**
- Service shows as "Crashed" in Railway
- Logs show error/panic before service starts

**Solution:**
1. Check logs for the specific error
2. Fix the error (common issues below)
3. Redeploy the service

**Common startup errors:**
- Missing dependencies in `requirements.txt`
- Import errors (missing modules)
- Code errors (syntax, NameError, etc.)
- Memory issues (out of RAM)
- Port conflicts (unlikely on Railway)

### Fix 2: Service Still Starting (Models Loading)

**Symptoms:**
- Service is "Running" but returns 502
- Logs show "Starting lazy model loading..."
- No error messages

**Solution:**
- **Wait 60-90 seconds** for models to load
- The service should respond to `/health` and `/ping` immediately (even while models load)
- If `/ping` returns 502, the service itself isn't running

### Fix 3: Port Mismatch

**Symptoms:**
- Service is running but not accessible
- Different port in logs vs. Railway configuration

**Solution:**
1. Verify `start.sh` uses `$PORT`:
   ```bash
   PORT=${PORT:-8000}
   exec python -m uvicorn app:app --host 0.0.0.0 --port "$PORT" --workers 1
   ```

2. Verify Railway sets `PORT` automatically (it should)

3. Check service logs for actual port:
   ```
   🌐 Starting server on port 8000
   ```

### Fix 4: Network/Routing Issue

**Symptoms:**
- Service is running (logs show it)
- Health checks pass internally
- External requests return 502

**Solution:**
1. Check Railway service settings
2. Verify public URL is correct
3. Check Railway status page for outages
4. Try redeploying the service

---

## Immediate Actions

### 1. Check Python ML Service Logs

**Most Important:** Check the Python ML service logs in Railway to see what's happening.

**Look for:**
- ✅ Service started successfully
- ❌ Service crashed/errored
- ⏳ Service still starting (models loading)

### 2. Test Endpoints Directly

```bash
# Test all endpoints
curl -v https://python-ml-service-production-a6b8.up.railway.app/ping
curl -v https://python-ml-service-production-a6b8.up.railway.app/health
curl -v https://python-ml-service-production-a6b8.up.railway.app/
```

The `-v` flag shows detailed connection info, which helps identify the issue.

### 3. Check Service Status in Railway

- Is the service "Running"?
- Is the health check passing?
- When was the last successful deployment?

---

## Expected Behavior

### When Service is Running:

**Logs:**
```
🚀 Starting Python ML Service...
📱 Device: cpu
📚 Models will be loaded lazily on first request
🌐 Starting server on port 8000
INFO:     Started server process [1]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:8000
```

**Endpoints:**
- `/ping` → `200 OK` → `{"status":"ok","message":"Python ML Service is running"}`
- `/health` → `200 OK` → `{"status":"healthy",...}`
- `/` → `200 OK` → `{"status":"ok","service":"Python ML Service",...}`

### When Service is Not Running:

**Endpoints:**
- All endpoints → `502 Bad Gateway`
- Connection refused/timeout

---

## Next Steps

1. ✅ **Check Python ML Service Logs** - Most important step
2. ⏳ **Test endpoints directly** - Verify service accessibility
3. ⏳ **Check service status** - Verify it's running in Railway
4. ⏳ **Fix any errors found** - Based on log analysis
5. ⏳ **Redeploy if needed** - After fixes

---

## Port Configuration Verification

**Railway automatically handles ports**, so port issues are unlikely. However, verify:

1. **Python ML Service:**
   - Uses `PORT` environment variable (set by Railway)
   - Defaults to 8000 if not set
   - `start.sh` correctly uses `$PORT`

2. **Classification Service:**
   - Uses public URL (no port needed)
   - Railway routes automatically

**The port is NOT the issue** - Railway handles this automatically. The 502 error indicates the service itself is not running or not accessible.

---

## Summary

The 502 error means the Python ML service is **not running or not accessible**. The most likely causes are:

1. **Service crashed during startup** (check logs)
2. **Service failed to start** (check logs)
3. **Network/routing issue** (less likely)

**Action Required:** Check the Python ML service logs in Railway to identify the specific error.

