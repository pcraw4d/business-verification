# Python ML Service Port Configuration Issue

**Date**: November 27, 2025  
**Issue**: Service running on port 8080, but Railway routing expects different port  
**Status**: 🔧 **Port Mismatch Identified**

---

## Issue Summary

The Python ML service logs show:
```
INFO:     Uvicorn running on http://0.0.0.0:8080 (Press CTRL+C to quit)
```

But Railway shows the service is configured for port 8000. External requests return:
```json
{"status":"error","code":502,"message":"Application failed to respond"}
```

This is a **Railway proxy error**, not an application error, indicating Railway's router cannot reach the application.

---

## Root Cause

**Port Mismatch:**
- Service is running on port **8080** (from logs)
- Railway expects service on port **8000** (from configuration)
- Railway's proxy/router cannot connect to the application

**Why this happens:**
- Railway automatically sets the `PORT` environment variable
- The service should use `$PORT` from environment
- If Railway sets `PORT=8080` but the proxy expects `8000`, there's a mismatch
- Or the service is ignoring the `PORT` environment variable

---

## Solution

### Option 1: Force Port 8000 (Recommended)

Railway should automatically set `PORT`, but we can ensure the service uses the correct port:

**Update `python_ml_service/start.sh`:**
```bash
#!/bin/sh
# Start script for Railway deployment
# Railway sets PORT environment variable automatically

# Ensure we use Railway's PORT, default to 8000
PORT=${PORT:-8000}
logger.info(f"🌐 Starting server on port {PORT}")

exec python -m uvicorn app:app --host 0.0.0.0 --port "$PORT" --workers 1
```

**Update `python_ml_service/app.py` (if using `if __name__ == "__main__"`):**
```python
# Get port from environment (Railway sets this automatically)
port = int(os.getenv("PORT", "8000"))
logger.info(f"🌐 Starting server on port {port}")

uvicorn.run(
    app,
    host="0.0.0.0",
    port=port,  # Use environment PORT
    log_level="info"
)
```

### Option 2: Check Railway PORT Environment Variable

1. Go to Railway Dashboard → `python-ml-service` → Variables
2. Check if `PORT` is set
3. If it's set to `8080`, change it to `8000` (or vice versa)
4. If not set, add `PORT=8000`

### Option 3: Verify Service Configuration

**In Railway Dashboard:**
1. Go to `python-ml-service` → Settings
2. Check "Service Settings"
3. Verify:
   - Port is set correctly
   - Health check path is `/health`
   - Start command is `sh start.sh`

---

## Verification

After fixing the port:

1. **Check service logs** - Should show:
   ```
   🌐 Starting server on port 8000
   INFO:     Uvicorn running on http://0.0.0.0:8000
   ```

2. **Test endpoints:**
   ```bash
   curl https://python-ml-service-production-a6b8.up.railway.app/ping
   # Expected: {"status":"ok","message":"Python ML Service is running"}
   
   curl https://python-ml-service-production-a6b8.up.railway.app/health
   # Expected: {"status":"healthy",...}
   ```

3. **Check classification service logs** - Should show:
   ```
   ✅ Python ML Service initialized successfully
   ```

---

## Current Status

- ✅ Service is running (logs show it started)
- ✅ Health endpoint works internally (200 OK in logs)
- ❌ External requests return 502 (Railway proxy can't reach app)
- ❌ Port mismatch: Service on 8080, Railway expects 8000

---

## Immediate Fix

The service should use Railway's `PORT` environment variable. Since Railway is setting `PORT=8080`, but the service might be defaulting to 8000, or vice versa, we need to ensure consistency.

**Check Railway Variables:**
- Go to Railway Dashboard → `python-ml-service` → Variables
- Verify `PORT` is set and matches what the service is using
- If `PORT` is not set or is `8080`, set it to `8000` and redeploy

**Or ensure service uses Railway's PORT:**
- The `start.sh` script already uses `$PORT`
- The `app.py` also reads `PORT` from environment
- But Railway might be setting it to 8080

**Best solution:** Let Railway set `PORT` automatically, and ensure the service uses it. Railway should handle this, but there might be a configuration issue.

