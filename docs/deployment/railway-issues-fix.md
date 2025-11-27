# Railway Issues Fix - Python ML Service & Keyword Extraction

**Date**: November 27, 2025  
**Status**: 🔧 **Issues Identified - Fix Required**

---

## Issues Identified

### Issue 1: Python ML Service Not Initialized ❌

**Symptoms:**
- No initialization messages in logs
- No "🐍 Initializing Python ML Service" log
- No "✅ Python ML Service initialized successfully" log
- Classification using standard keyword-based method instead of enhanced DistilBART

**Root Cause:**
- `PYTHON_ML_SERVICE_URL` environment variable is **not set** in Railway for classification-service

**Impact:**
- Enhanced classification with DistilBART is not available
- No explanation or content summary in responses
- Lower classification accuracy

**Fix:**
```bash
# Set the environment variable in Railway
railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app" --service classification-service
```

Or use the automated script:
```bash
./scripts/configure-classification-service-railway.sh
```

### Issue 2: Corrupted Keyword Extraction ❌

**Symptoms:**
- Keywords appear as garbled text: `technology:ai axjy orux orux oemf yduc xyipy`
- Keywords should be normal words like: `wine`, `retail`, `beverage`, `store`
- This affects classification accuracy

**Root Cause:**
- Text extraction from website content is producing corrupted/encoded text
- Possible encoding issues in HTML parsing
- Website content might be compressed or encoded incorrectly

**Impact:**
- Classification accuracy is reduced
- Industry detection is inaccurate
- Code generation is affected

**Fix Required:**
1. Check website content extraction encoding
2. Verify HTML parsing handles different encodings correctly
3. Add validation to filter out corrupted keywords

---

## Immediate Actions

### Step 1: Set Python ML Service URL

```bash
# Option 1: Using Railway CLI
railway link --service classification-service
railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app" --service classification-service

# Option 2: Using automated script
./scripts/configure-classification-service-railway.sh

# Option 3: Railway Dashboard
# Go to Railway Dashboard → classification-service → Variables
# Add: PYTHON_ML_SERVICE_URL = https://python-ml-service-production.up.railway.app
```

### Step 2: Verify Python ML Service is Running

```bash
# Check Python ML service health
curl https://python-ml-service-production.up.railway.app/health

# Expected response:
# {"status":"healthy","distilbart_status":"loaded",...}
```

### Step 3: Verify Classification Service Initialization

After setting the environment variable, check logs for:
```
🐍 Initializing Python ML Service at https://python-ml-service-production.up.railway.app
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

### Step 4: Fix Keyword Extraction

The keyword extraction issue needs code fixes. See the next section.

---

## Keyword Extraction Fix

The corrupted keywords suggest an encoding issue. We need to:

1. **Add encoding detection** in HTML parsing
2. **Validate keywords** before adding them to the list
3. **Filter out corrupted keywords** (non-printable characters, invalid UTF-8)

### Proposed Fix

Add validation to keyword extraction:

```go
// Validate keyword before adding
func isValidKeyword(keyword string) bool {
    // Check length
    if len(keyword) < 2 || len(keyword) > 50 {
        return false
    }
    
    // Check for valid UTF-8
    if !utf8.ValidString(keyword) {
        return false
    }
    
    // Check for non-printable characters (except spaces)
    for _, r := range keyword {
        if !unicode.IsPrint(r) && r != ' ' {
            return false
        }
    }
    
    // Check for common word patterns (at least 50% letters)
    letterCount := 0
    for _, r := range keyword {
        if unicode.IsLetter(r) {
            letterCount++
        }
    }
    if float64(letterCount)/float64(len(keyword)) < 0.5 {
        return false
    }
    
    return true
}
```

---

## Verification Checklist

After fixes:

- [ ] `PYTHON_ML_SERVICE_URL` is set in Railway
- [ ] Python ML service is running and healthy
- [ ] Classification service logs show Python ML initialization
- [ ] Classification requests with website URL use enhanced classification
- [ ] Keywords are normal words (not corrupted)
- [ ] Classification accuracy improves
- [ ] Explanation and content summary appear in responses

---

## Expected Behavior After Fix

### Before Fix:
```
Starting industry detection
- Top keywords: [technology:ai axjy orux orux oemf yduc xyipy...]
Enhanced classification result
  primary_industry: Technology
  confidence: 0.508
```

### After Fix:
```
🐍 Initializing Python ML Service at https://python-ml-service-production.up.railway.app
✅ Python ML Service initialized successfully

Using Python ML service for enhanced classification
Python ML service enhanced classification successful
  industry: Food & Beverage
  confidence: 0.92
  quantization_enabled: true

Enhanced classification result
  primary_industry: Food & Beverage
  confidence: 0.92
  explanation: "The business was classified as Food & Beverage based on..."
  content_summary: "This business is a wine and spirits retailer..."
```

---

## Next Steps

1. ✅ **Set `PYTHON_ML_SERVICE_URL`** in Railway (immediate)
2. ⏳ **Fix keyword extraction validation** (code change required)
3. ⏳ **Test enhanced classification** with website URL
4. ⏳ **Verify keywords are normal words**
5. ⏳ **Monitor classification accuracy**

