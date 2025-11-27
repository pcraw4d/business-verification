# Classification Service - Python ML Service Integration Setup

**Date**: November 27, 2025  
**Status**: ✅ **Ready for Configuration**

---

## Overview

The Classification Service can use the Python ML Service for enhanced classification with DistilBART when a website URL is provided. This document explains how to configure the integration and verify the fallback logic.

---

## Required Configuration

### Railway Environment Variable

**For Classification Service:**

The `PYTHON_ML_SERVICE_URL` environment variable must be set in Railway for the classification service to connect to the Python ML service.

```bash
PYTHON_ML_SERVICE_URL=https://python-ml-service-production.up.railway.app
```

### How to Set in Railway

#### Option 1: Using Railway CLI (Recommended)

```bash
# Run the configuration script
./scripts/configure-classification-service-railway.sh
```

This script will:
1. Detect the Python ML service URL automatically (if available)
2. Set the `PYTHON_ML_SERVICE_URL` environment variable
3. Verify the configuration

#### Option 2: Manual Setup via Railway Dashboard

1. Go to Railway Dashboard
2. Select your project
3. Click on `classification-service`
4. Go to "Variables" tab
5. Click "New Variable"
6. Add:
   - **Name**: `PYTHON_ML_SERVICE_URL`
   - **Value**: Your Python ML service URL (e.g., `https://python-ml-service-production-xxx.up.railway.app`)
7. Save

#### Option 3: Using Railway CLI Directly

```bash
# Link to classification-service
railway link --service classification-service

# Set the environment variable
railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app" --service classification-service
```

---

## Fallback Logic Verification

### How It Works

The classification service has **robust fallback logic** that ensures it always works, even if the Python ML service is unavailable:

```
Classification Request
│
├─ Check: Is PYTHON_ML_SERVICE_URL set?
│  │
│  ├─ NO → Use standard keyword-based classification ✅
│  │
│  └─ YES → Check: Is Python ML service initialized?
│     │
│     ├─ NO → Use standard keyword-based classification ✅
│     │
│     └─ YES → Check: Does request have website URL?
│        │
│        ├─ NO → Use standard keyword-based classification ✅
│        │
│        └─ YES → Call Python ML Service
│           │
│           ├─ SUCCESS → Return enhanced classification ✅
│           │
│           └─ FAILURE → Log warning, fallback to standard classification ✅
```

### Fallback Scenarios

1. **Environment Variable Not Set**
   - **Log**: `ℹ️ Python ML Service URL not configured, enhanced classification will not be available`
   - **Action**: Uses standard keyword-based classification
   - **Status**: ✅ Service continues normally

2. **Python ML Service Initialization Fails**
   - **Log**: `⚠️ Failed to initialize Python ML Service, continuing without enhanced classification`
   - **Action**: Sets `pythonMLService = nil`, uses standard classification
   - **Status**: ✅ Service continues normally

3. **Python ML Service Call Fails**
   - **Log**: `Python ML service enhanced classification failed, falling back to standard classification`
   - **Action**: Continues with standard keyword-based classification
   - **Status**: ✅ Service continues normally

4. **No Website URL in Request**
   - **Action**: Skips Python ML service, uses standard classification
   - **Status**: ✅ Service continues normally

### Code Verification

The fallback logic is implemented in:

1. **Initialization** (`services/classification-service/cmd/main.go:82-102`):
   ```go
   pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
   if pythonMLServiceURL != "" {
       // Initialize service
       if err := pythonMLService.Initialize(initCtx); err != nil {
           pythonMLService = nil // Safe fallback
       }
   }
   ```

2. **Request Handling** (`services/classification-service/internal/handlers/classification.go:1003-1066`):
   ```go
   if h.pythonMLService != nil && req.WebsiteURL != "" {
       pms, ok := h.pythonMLService.(*infrastructure.PythonMLService)
       if ok && pms != nil {
           // Try enhanced classification
           if err != nil {
               // Fallback to standard classification
           }
       }
   }
   // Standard classification continues here
   ```

---

## Verification Steps

### Step 1: Verify Environment Variable

```bash
# Check if variable is set in Railway
railway variables --service classification-service | grep PYTHON_ML_SERVICE_URL
```

Or check Railway Dashboard → classification-service → Variables

### Step 2: Check Service Logs

After setting the variable, check classification service logs for:

**✅ Success:**
```
🐍 Initializing Python ML Service at https://python-ml-service-production.up.railway.app
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

**⚠️ Fallback (if initialization fails):**
```
🐍 Initializing Python ML Service at https://python-ml-service-production.up.railway.app
⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
✅ Classification services initialized
  python_ml_service: false
```

**ℹ️ Not Configured:**
```
ℹ️ Python ML Service URL not configured, enhanced classification will not be available
✅ Classification services initialized
  python_ml_service: false
```

### Step 3: Test Enhanced Classification

Make a classification request with a website URL:

```bash
curl -X POST https://classification-service-production.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "TechCorp Solutions",
    "description": "Software development services",
    "website_url": "https://techcorp.com"
  }'
```

**Expected Logs (if Python ML service is available):**
```
Using Python ML service for enhanced classification
  request_id: req_xxx
  website_url: https://techcorp.com
Python ML service enhanced classification successful
  industry: Technology
  confidence: 0.92
  quantization_enabled: true
```

**Expected Logs (if Python ML service is unavailable):**
```
Starting industry detection
  business_name: TechCorp Solutions
Industry detection successful
  industry: Technology
  confidence: 0.85
```

### Step 4: Test Fallback

Test that the service gracefully handles Python ML service failures:

1. **Temporarily set wrong URL:**
   ```bash
   railway variables set PYTHON_ML_SERVICE_URL="https://invalid-url.com" --service classification-service
   ```

2. **Make classification request** (should still work with standard classification)

3. **Check logs** for fallback message:
   ```
   ⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
   ```

4. **Restore correct URL:**
   ```bash
   railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app" --service classification-service
   ```

---

## Troubleshooting

### Issue: Python ML Service Not Being Called

**Symptoms:**
- Logs show "Starting industry detection" instead of "Using Python ML service"
- No explanation or contentSummary in response

**Check:**
1. Is `PYTHON_ML_SERVICE_URL` set?
   ```bash
   railway variables --service classification-service | grep PYTHON_ML_SERVICE_URL
   ```

2. Is Python ML service initialized?
   - Check logs for: "✅ Python ML Service initialized successfully"
   - Check logs for: "python_ml_service: true"

3. Does request include website URL?
   - Enhanced classification only works when `website_url` is provided

### Issue: Panic on Classification Request

**Symptoms:**
- Service crashes with "nil pointer dereference"
- Error: `runtime error: invalid memory address or nil pointer dereference`

**Cause:**
- This was fixed in commit `746ae19e9` - ensure you have the latest code

**Verify:**
- Check that the code has the nil check: `if ok && pms != nil`

### Issue: Python ML Service Returns 502

**Symptoms:**
- Logs show: "Python ML service enhanced classification failed"
- Falls back to standard classification

**Check:**
1. Is Python ML service running?
   ```bash
   curl https://python-ml-service-production.up.railway.app/health
   ```

2. Are models loaded?
   - Check Python ML service logs for: "✅ DistilBART classifier initialized"

3. Is service URL correct?
   - Verify the URL in Railway dashboard matches the environment variable

---

## Expected Behavior Summary

| Scenario | Python ML Service | Website URL | Result |
|----------|-------------------|-------------|--------|
| ✅ Configured & Available | ✅ Available | ✅ Provided | Enhanced classification with DistilBART |
| ✅ Configured & Available | ✅ Available | ❌ Not provided | Standard keyword-based classification |
| ✅ Configured | ❌ Initialization failed | ✅ Provided | Standard keyword-based classification (fallback) |
| ✅ Configured | ❌ Call failed | ✅ Provided | Standard keyword-based classification (fallback) |
| ❌ Not configured | N/A | ✅ Provided | Standard keyword-based classification |
| ❌ Not configured | N/A | ❌ Not provided | Standard keyword-based classification |

**Key Point**: The service **always** works, regardless of Python ML service availability. Enhanced classification is a **nice-to-have** feature that gracefully falls back when unavailable.

---

## Next Steps

1. ✅ **Set Environment Variable**: Use the script or Railway dashboard
2. ✅ **Verify Initialization**: Check logs for successful initialization
3. ✅ **Test Enhanced Classification**: Make a request with website URL
4. ✅ **Verify Fallback**: Test that service works when Python ML service is unavailable
5. ✅ **Monitor Logs**: Watch for any errors or warnings

---

## Related Documentation

- [Python ML Service Integration Summary](../deployment/python-ml-service-integration-summary.md)
- [Python ML Service Railway Setup](../deployment/python-ml-service-railway-setup.md)
- [Classification Service Deployment Guide](../classification-service-deployment-guide.md)

