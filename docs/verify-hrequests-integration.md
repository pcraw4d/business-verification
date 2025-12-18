# hrequests Integration Verification Guide

## Integration Flow

```
Classification Service Startup
    ↓
NewEnhancedWebsiteScraper()
    ↓
external.NewWebsiteScraper()  ← Reads HREQUESTS_SERVICE_URL
    ↓
NewWebsiteScraperWithStrategies()
    ↓
Checks HREQUESTS_SERVICE_URL env var
    ↓
If set: Creates HrequestsScraper (Strategy 0)
    ↓
Strategy order: hrequests → SimpleHTTP → BrowserHeaders → Playwright
```

## Verification Steps

### Step 1: Check Environment Variable

**In Railway Dashboard:**
1. Go to classification-service settings
2. Check Environment Variables
3. Verify `HREQUESTS_SERVICE_URL` is set to: `https://hrequestsservice-production.up.railway.app/`

**Expected Value:**
```
HREQUESTS_SERVICE_URL=https://hrequestsservice-production.up.railway.app/
```

---

### Step 2: Check Startup Logs

**Look for these log messages in classification-service startup logs:**

**✅ If hrequests is enabled:**
```
✅ [Scraper] hrequests strategy enabled
   service_url=https://hrequestsservice-production.up.railway.app/
```

**⚠️ If hrequests is disabled:**
```
ℹ️ [Scraper] hrequests strategy disabled (HREQUESTS_SERVICE_URL not set)
```

**Location:** Check Railway logs for classification-service startup

---

### Step 3: Test Classification Request

**Make a test classification request:**

```bash
curl -X POST https://your-classification-service.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "website_url": "https://example.com"
  }'
```

**Check logs for scraping activity:**

**Expected log sequence:**

1. **hrequests attempt:**
```
🔍 [Phase1] [Hrequests] Starting scrape attempt
   url=https://example.com
```

2. **Success (if hrequests works):**
```
✅ [Phase1] [Hrequests] Scrape succeeded
   quality_score=0.85
   word_count=500
```

3. **Early exit (if quality is high):**
```
✅ [EarlyExit] High-quality content found, skipping remaining strategies
   strategy=hrequests
   quality_score=0.85
   word_count=500
```

4. **Fallback (if hrequests fails):**
```
⚠️ [Phase1] [Hrequests] Scrape failed
   url=https://example.com
🔍 [Phase1] [SimpleHTTP] Starting scrape attempt
```

---

### Step 4: Verify Strategy Usage

**Monitor logs to see which strategies are being used:**

**Log patterns to look for:**
- `[Phase1] [Hrequests]` - hrequests strategy
- `[Phase1] [SimpleHTTP]` - SimpleHTTP strategy  
- `[Phase1] [BrowserHeaders]` - BrowserHeaders strategy
- `[Phase1] [Playwright]` - Playwright fallback

**Expected distribution (after warm-up):**
- hrequests: 60-70% of requests
- SimpleHTTP/BrowserHeaders: 10-20% of requests
- Playwright: 20-30% of requests

---

### Step 5: Test Parallel Smart Crawling

**The parallel smart crawling should also be active:**

**Look for logs:**
```
✅ [ParallelScrape] Parallel execution completed: method=single_page
   quality=0.85
   words=500
```

Or:
```
✅ [ParallelScrape] Parallel execution completed: method=merged
   quality=0.90
   words=1200
```

---

## Troubleshooting

### Issue: hrequests not being used

**Check:**
1. ✅ `HREQUESTS_SERVICE_URL` is set correctly
2. ✅ hrequests-scraper service is running and healthy
3. ✅ Network connectivity between services
4. ✅ Check logs for initialization messages

**Common issues:**
- URL has trailing slash (should be: `https://hrequestsservice-production.up.railway.app/`)
- Service name mismatch (use Railway service name for internal routing)
- Network policy blocking inter-service communication

### Issue: All requests using Playwright

**Possible causes:**
1. hrequests service is down or unreachable
2. hrequests failing for all sites (check error logs)
3. Environment variable not set correctly

**Check logs for:**
```
⚠️ [Hrequests] HTTP request failed
⚠️ [Hrequests] Service returned error status
```

---

## Success Indicators

✅ **Integration Successful When:**

1. Startup logs show: `✅ [Scraper] hrequests strategy enabled`
2. Classification requests show hrequests being attempted first
3. Logs show `[Phase1] [Hrequests]` for 60-70% of requests
4. Early exit logs appear for high-quality content
5. Overall success rate remains ≥95%
6. Average latency improves to ~1.5s

---

## Test Script

Save this as `test-hrequests-integration.sh`:

```bash
#!/bin/bash

CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://your-classification-service.railway.app}"
TEST_URL="https://example.com"

echo "=== Testing hrequests Integration ==="
echo ""

echo "1. Testing hrequests service health..."
curl -k -s https://hrequestsservice-production.up.railway.app/health | jq .
echo ""

echo "2. Testing hrequests scraping directly..."
curl -k -s -X POST https://hrequestsservice-production.up.railway.app/scrape \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"$TEST_URL\"}" | jq '{success, method, latency_ms, content: {title, word_count, quality_score}}'
echo ""

echo "3. Testing classification service..."
RESPONSE=$(curl -k -s -X POST "$CLASSIFICATION_SERVICE_URL/api/v1/classify" \
  -H "Content-Type: application/json" \
  -d "{
    \"business_name\": \"Test Company\",
    \"website_url\": \"$TEST_URL\"
  }")

echo "$RESPONSE" | jq '{success, primary_industry, confidence_score, processing_time}' || echo "$RESPONSE"
echo ""

echo "4. Check Railway logs for:"
echo "   - ✅ [Scraper] hrequests strategy enabled"
echo "   - 🔍 [Phase1] [Hrequests] Starting scrape attempt"
echo "   - ✅ [Phase1] [Hrequests] Scrape succeeded"
echo ""
```

---

## Next Steps

After verification:

1. Monitor for 24-48 hours
2. Track strategy usage distribution
3. Measure latency improvements
4. Verify success rates remain high
5. Adjust thresholds if needed

