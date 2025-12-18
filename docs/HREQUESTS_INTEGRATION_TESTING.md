# hrequests Integration Testing Guide

## Setup Verification

✅ **Completed:**

- hrequests-scraper service built and deployed
- `HREQUESTS_SERVICE_URL` environment variable set in classification-service
- Code changes committed and deployed

## Testing Steps

### 1. Verify hrequests Service Health

Test the hrequests-scraper service health endpoint:

```bash
# Get the hrequests service URL from Railway
curl https://your-hrequests-service.railway.app/health
```

**Expected Response:**

```json
{
  "status": "healthy",
  "service": "hrequests-scraper",
  "timestamp": 1234567890.0
}
```

### 2. Test hrequests Scraping Directly

Test the scraping endpoint:

```bash
curl -X POST https://your-hrequests-service.railway.app/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

**Expected Response:**

```json
{
  "success": true,
  "content": {
    "title": "Example Domain",
    "word_count": 500,
    "quality_score": 0.85,
    ...
  },
  "method": "hrequests",
  "latency_ms": 1234
}
```

### 3. Verify Classification Service Integration

Check classification-service logs for hrequests initialization:

**Look for these log messages:**

```
✅ [Scraper] hrequests strategy enabled
   service_url=http://hrequests-scraper:8080
```

**If hrequests is disabled (URL not set):**

```
ℹ️ [Scraper] hrequests strategy disabled (HREQUESTS_SERVICE_URL not set)
```

### 4. Test Classification with Hybrid Scraping

Make a classification request and check logs:

```bash
curl -X POST https://your-classification-service.railway.app/api/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "website": "https://example.com"
  }'
```

**Look for these log patterns:**

**Successful hrequests scrape:**

```
🔍 [Phase1] [Hrequests] Starting scrape attempt
✅ [Phase1] [Hrequests] Scrape succeeded
   quality_score=0.85
   word_count=500
```

**Early exit (high quality):**

```
✅ [EarlyExit] High-quality content found, skipping remaining strategies
   strategy=hrequests
   quality_score=0.85
   word_count=500
```

**Fallback to Playwright:**

```
⚠️ [Phase1] [Hrequests] Scrape failed
🔍 [Phase1] [Playwright] Starting scrape attempt
✅ [Phase1] [Playwright] Scrape succeeded
```

### 5. Monitor Strategy Usage

Track which strategies are being used:

**Expected Distribution (after warm-up):**

- hrequests: 60-70% of requests
- Playwright fallback: 30-40% of requests
- Overall success rate: ≥95%

**Log indicators:**

- `[Phase1] [Hrequests]` - hrequests strategy used
- `[Phase1] [SimpleHTTP]` - SimpleHTTP strategy used
- `[Phase1] [BrowserHeaders]` - BrowserHeaders strategy used
- `[Phase1] [Playwright]` - Playwright fallback used

### 6. Performance Monitoring

**Key Metrics to Track:**

1. **Latency:**

   - hrequests: ~1s average
   - Playwright: ~2.5s average
   - Overall: ~1.5s average (with hybrid)

2. **Early Exit Rate:**

   - Look for `[EarlyExit]` log messages
   - Target: 20-30% of successful scrapes

3. **Success Rates:**
   - hrequests success: 80-90%
   - Overall success: ≥95%

### 7. Troubleshooting

**If hrequests is not being used:**

1. Check environment variable:

   ```bash
   # In Railway, verify HREQUESTS_SERVICE_URL is set
   # Should be: http://hrequests-scraper:8080 (internal)
   # Or: https://your-hrequests-service.railway.app (external)
   ```

2. Check service connectivity:

   - Verify hrequests service is running
   - Check network connectivity between services
   - Verify service URL format (use internal Railway service name)

3. Check logs for errors:
   ```
   ⚠️ [Hrequests] HTTP request failed
   ⚠️ [Hrequests] Service returned error status
   ```

**If builds are failing:**

1. Verify Dockerfile path in Railway settings
2. Check root directory is set to `services/hrequests-scraper/`
3. Verify `railway.json` exists in service directory

## Success Criteria

✅ **Integration Successful When:**

- hrequests service responds to health checks
- Classification service logs show hrequests strategy enabled
- Classification requests complete successfully
- Logs show hrequests being used for 60-70% of requests
- Overall success rate remains ≥95%
- Average latency improves to ~1.5s

## Next Steps

1. Monitor for 24-48 hours
2. Collect metrics on strategy distribution
3. Adjust thresholds if needed (early exit, quality scores)
4. Optimize based on real-world performance data
