# Python ML Service Railway Port Configuration

**Date**: November 27, 2025  
**Issue**: Service running on port 8080, but Railway expects 8000  
**Status**: 🔧 **Configuration Issue**

---

## Current Situation

**Service Logs Show:**
```
INFO:     Uvicorn running on http://0.0.0.0:8080
```

**Railway Configuration:**
- `railway.json` sets `PORT=8000`
- `startCommand` is `sh start.sh`
- `start.sh` uses `$PORT` (defaults to 8000)

**External Requests:**
- Return 502: "Application failed to respond"
- Railway proxy cannot reach the service

---

## Root Cause

Railway is **automatically setting `PORT=8080`**, which overrides the `PORT=8000` in `railway.json`. However, Railway's proxy/router might be configured to route to port 8000, causing a mismatch.

**Why this happens:**
1. Railway automatically assigns ports to services
2. Railway sets `PORT` environment variable automatically
3. Railway's `railway.json` variables might be overridden by Railway's automatic port assignment
4. The service uses Railway's `PORT` (8080), but the proxy expects 8000

---

## Solution

### Option 1: Remove PORT from railway.json (Recommended)

Railway automatically sets `PORT`, so we shouldn't override it in `railway.json`. Let Railway handle port assignment automatically.

**Update `python_ml_service/railway.json`:**
```json
{
  "environments": {
    "production": {
      "variables": {
        "ENVIRONMENT": "production",
        "LOG_LEVEL": "info",
        // Remove "PORT": "8000" - let Railway set it automatically
        "USE_QUANTIZATION": "true",
        ...
      }
    }
  }
}
```

### Option 2: Force Port 8000 in Railway Dashboard

1. Go to Railway Dashboard → `python-ml-service` → Variables
2. Check if `PORT` is set to `8080`
3. If so, change it to `8000` and save
4. Service will redeploy and use port 8000

### Option 3: Verify Railway is Using startCommand

Railway's `startCommand` in `railway.json` should override the Dockerfile CMD. Verify:

1. Check Railway service logs for startup command
2. Should see: `🌐 Starting Python ML Service on port XXXX`
3. If you see a different port, Railway might not be using `start.sh`

---

## Verification Steps

### Step 1: Check What Port Railway Assigned

**In Railway Dashboard:**
1. Go to `python-ml-service` → Variables
2. Check the `PORT` variable value
3. Note: Railway might show it as "Auto" or a specific port

### Step 2: Check Service Logs

After redeploy, check logs for:
```
🌐 Starting Python ML Service on port 8000
INFO:     Uvicorn running on http://0.0.0.0:8000
```

Or:
```
🌐 Starting Python ML Service on port 8080
INFO:     Uvicorn running on http://0.0.0.0:8080
```

### Step 3: Test Endpoints

```bash
# Should work if port matches Railway's proxy
curl https://python-ml-service-production-a6b8.up.railway.app/ping
curl https://python-ml-service-production-a6b8.up.railway.app/health
```

---

## Important Note

**Railway automatically handles port routing.** The internal port the service listens on doesn't matter for external access - Railway's proxy routes traffic correctly. However, if Railway sets `PORT=8080` but the proxy expects `8000`, there's a configuration mismatch.

**The real issue:** Railway's proxy cannot reach the service, which suggests:
1. Port mismatch (service on 8080, proxy expects 8000)
2. Service is crashing on requests
3. Service isn't actually listening on the port

---

## Recommended Fix

1. **Remove `PORT` from `railway.json`** - Let Railway set it automatically
2. **Ensure `start.sh` uses `$PORT`** - Already done ✅
3. **Verify Railway is using `startCommand`** - Check logs
4. **Test endpoints after redeploy**

---

## After Fix

The service should:
- Use whatever port Railway assigns
- Log the port on startup
- Respond to external requests correctly
- Initialize successfully in classification service

