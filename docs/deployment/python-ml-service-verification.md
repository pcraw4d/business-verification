# Python ML Service Verification Guide

**Date**: November 27, 2025  
**Status**: ✅ **Environment Variable Set**

---

## Current Configuration

The `PYTHON_ML_SERVICE_URL` is now set in Railway:
```
https://python-ml-service-production-a6b8.up.railway.app/
```

**Note**: There's a trailing slash in the URL. This is generally fine as most HTTP servers handle it, but URLs will become:
- `https://python-ml-service-production-a6b8.up.railway.app//health` (double slash)
- `https://python-ml-service-production-a6b8.up.railway.app//classify-enhanced` (double slash)

Most servers normalize this automatically, but for cleaner URLs, you can remove the trailing slash.

---

## Verification Steps

### Step 1: Check Classification Service Logs

After the environment variable is set, the classification service will automatically redeploy. Check the logs for:

**✅ Success:**
```
🐍 Initializing Python ML Service at https://python-ml-service-production-a6b8.up.railway.app/
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

**⚠️ If initialization fails:**
```
🐍 Initializing Python ML Service at https://python-ml-service-production-a6b8.up.railway.app/
⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
  error: [error details]
```

### Step 2: Verify Python ML Service is Running

```bash
# Test health endpoint
curl https://python-ml-service-production-a6b8.up.railway.app/health

# Expected response:
# {"status":"healthy","distilbart_status":"loaded",...}
```

### Step 3: Test Enhanced Classification

Make a classification request with a website URL:

```bash
curl -X POST https://classification-service-production.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Business",
    "description": "Software development services",
    "website_url": "https://example.com"
  }'
```

**Expected logs:**
```
Using Python ML service for enhanced classification
  request_id: req_xxx
  website_url: https://example.com
Python ML service enhanced classification successful
  industry: Technology
  confidence: 0.92
  quantization_enabled: true
```

**Expected response includes:**
- `explanation`: Human-readable explanation
- `contentSummary`: Website content summary
- `quantizationEnabled`: true
- `modelVersion`: distilbart-v1.0

### Step 4: Check UI

After making a classification request, check the UI for:

1. ✅ **Primary Industry** with confidence level
2. ✅ **Top 3 codes** (MCC/SIC/NAICS) with confidence
3. ✅ **Code distribution** chart
4. ✅ **Explanation** section (from DistilBART)
5. ✅ **Content Summary** (from website analysis)
6. ✅ **Quantization indicator** (if enabled)

---

## Troubleshooting

### Issue: Python ML Service Not Initializing

**Check:**
1. Is Python ML service running?
   ```bash
   curl https://python-ml-service-production-a6b8.up.railway.app/health
   ```

2. Is the URL correct in Railway?
   - Go to Railway Dashboard → classification-service → Variables
   - Verify `PYTHON_ML_SERVICE_URL` value

3. Check classification service logs for initialization errors

**Fix:**
- If Python ML service is down, check its logs
- If URL is wrong, update it in Railway
- If connection timeout, check network/firewall settings

### Issue: Enhanced Classification Not Being Used

**Check:**
1. Does the request include `website_url`?
   - Enhanced classification only works when website URL is provided

2. Check logs for:
   ```
   Using Python ML service for enhanced classification
   ```
   If this doesn't appear, the service might not be initialized

**Fix:**
- Ensure `PYTHON_ML_SERVICE_URL` is set correctly
- Verify Python ML service is running
- Check that requests include `website_url`

### Issue: Trailing Slash in URL

**Current URL:**
```
https://python-ml-service-production-a6b8.up.railway.app/
```

**Recommended (remove trailing slash):**
```
https://python-ml-service-production-a6b8.up.railway.app
```

**Why:**
- Cleaner URLs (no double slashes)
- More consistent with HTTP standards
- Still works with trailing slash, but cleaner without

**To fix:**
1. Go to Railway Dashboard → classification-service → Variables
2. Edit `PYTHON_ML_SERVICE_URL`
3. Remove the trailing slash
4. Save (service will auto-redeploy)

---

## Expected Behavior

### Before (Without Python ML Service):
```
Starting industry detection
- Top keywords: [technology:ai axjy orux...] (corrupted)
Enhanced classification result
  primary_industry: Technology
  confidence: 0.508
```

### After (With Python ML Service):
```
🐍 Initializing Python ML Service at https://python-ml-service-production-a6b8.up.railway.app/
✅ Python ML Service initialized successfully

Using Python ML service for enhanced classification
Python ML service enhanced classification successful
  industry: Technology
  confidence: 0.92
  quantization_enabled: true

Enhanced classification result
  primary_industry: Technology
  confidence: 0.92
  explanation: "The business was classified as Technology based on..."
  content_summary: "This business provides software development services..."
```

---

## Next Steps

1. ✅ **Environment variable is set** - Done!
2. ⏳ **Wait for service redeploy** - Classification service will auto-redeploy
3. ⏳ **Check logs** - Verify initialization success
4. ⏳ **Test enhanced classification** - Make a request with website URL
5. ⏳ **Verify UI** - Check that all enhanced outputs appear
6. ⏳ **Optional: Remove trailing slash** - For cleaner URLs

---

## Monitoring

After deployment, monitor:

- **Classification Service Logs**: Look for Python ML service initialization
- **Python ML Service Logs**: Look for model loading and quantization status
- **Classification Accuracy**: Should improve with DistilBART
- **Response Times**: Should be reasonable (100-300ms for enhanced classification)

---

## Success Criteria

✅ Python ML service initializes successfully  
✅ Enhanced classification is used when website URL is provided  
✅ Response includes explanation and content summary  
✅ UI displays all required outputs  
✅ Classification accuracy improves  
✅ Keywords are normal words (not corrupted) - *Separate fix needed*

