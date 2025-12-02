# Phase 1 Testing Guide

## Overview

Phase 1 implementation is complete and deployed. This guide helps you test and validate the improvements.

## Prerequisites

✅ Playwright service deployed and working  
✅ `PLAYWRIGHT_SERVICE_URL` set in classification service  
✅ Classification service has latest code deployed

## Testing the Playwright Service

### 1. Health Check
```bash
curl https://playwright-service-production-b21a.up.railway.app/health
```
**Expected:** `{"status":"ok","service":"playwright-scraper"}`

### 2. Test Scraping
```bash
curl -X POST https://playwright-service-production-b21a.up.railway.app/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```
**Expected:** JSON response with `html` field containing full HTML

## Testing the Classification Service

### 1. Test with Simple Website (Should use SimpleHTTP strategy)
```bash
curl -X POST https://your-classification-service.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Restaurant",
    "website_url": "https://example.com"
  }'
```

**Check logs for:**
- `"strategy": "simple_http"` or `"strategy": "browser_headers"`
- `"quality_score": 0.7` or higher
- `"word_count": 200` or higher

### 2. Test with JavaScript-Heavy Website (Should use Playwright strategy)
```bash
curl -X POST https://your-classification-service.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Modern Web App",
    "website_url": "https://react.dev"
  }'
```

**Check logs for:**
- `"strategy": "playwright"` (if SimpleHTTP and BrowserHeaders fail)
- `"quality_score": 0.7` or higher
- Structured content fields populated

### 3. Test with Various Website Types

Test with diverse websites to verify strategy fallback:
- Static HTML sites (should use SimpleHTTP)
- Sites with bot detection (should use BrowserHeaders)
- JavaScript-heavy SPAs (should use Playwright)

## Key Metrics to Monitor

### Scrape Success Rate
**Target:** ≥95%

**How to measure:**
- Count successful scrapes vs total attempts
- Check logs for "Strategy succeeded" vs "all scraping strategies failed"

### Content Quality Scores
**Target:** ≥0.7 for 90%+ of successful scrapes

**How to measure:**
- Check `quality_score` in logs
- Average should be ≥0.7
- 90%+ of scrapes should have score ≥0.7

### Word Count
**Target:** Average ≥200 words

**How to measure:**
- Check `word_count` in logs
- Average across all successful scrapes

### Strategy Distribution
**Expected:**
- SimpleHTTP: ~60% of requests
- BrowserHeaders: ~20-30% of requests
- Playwright: ~10-20% of requests

**How to measure:**
- Count log entries by strategy name
- Should see automatic fallback when earlier strategies fail

## Log Analysis

### Key Log Messages to Look For

**Successful Scrape:**
```
INFO: Starting scrape with structured content extraction url=...
INFO: Attempting scrape strategy strategy=simple_http attempt=1
INFO: Strategy succeeded strategy=simple_http quality_score=0.85
INFO: Scrape completed word_count=342 quality_score=0.85
```

**Strategy Fallback:**
```
INFO: Attempting scrape strategy strategy=simple_http attempt=1
WARN: Strategy failed, trying next strategy=simple_http
INFO: Attempting scrape strategy strategy=browser_headers attempt=2
INFO: Strategy succeeded strategy=browser_headers quality_score=0.82
```

**Playwright Usage:**
```
INFO: Attempting scrape strategy strategy=simple_http attempt=1
WARN: Strategy failed, trying next strategy=simple_http
INFO: Attempting scrape strategy strategy=browser_headers attempt=2
WARN: Strategy failed, trying next strategy=browser_headers
INFO: Attempting scrape strategy strategy=playwright attempt=3
INFO: Strategy succeeded strategy=playwright quality_score=0.88
```

## Expected Improvements

### Before Phase 1
- Scrape success rate: ~50%
- Content quality: Poor
- "No output" errors: Common
- Classification accuracy: <5%

### After Phase 1 (Expected)
- Scrape success rate: ≥95% ✅
- Content quality score: ≥0.7 (90%+ of scrapes) ✅
- Average word count: ≥200 ✅
- "No output" errors: <2% ✅
- **Classification accuracy: 50-60%** (to be validated)

## Troubleshooting

### Issue: Playwright strategy not being used
**Check:**
- `PLAYWRIGHT_SERVICE_URL` is set correctly
- Playwright service is accessible from classification service
- Check logs for connection errors

### Issue: Low quality scores
**Check:**
- Are websites returning proper HTML?
- Check if structured content is being extracted (title, headings, about)
- Verify content validation thresholds

### Issue: High Playwright usage (>30%)
**Possible causes:**
- SimpleHTTP and BrowserHeaders strategies failing too often
- Check for bot detection issues
- Verify HTTP client configuration

## Success Criteria Checklist

- [ ] Playwright service health check passes
- [ ] Playwright service can scrape test URLs
- [ ] Classification service logs show strategy usage
- [ ] Scrape success rate ≥95%
- [ ] Quality scores ≥0.7 for 90%+ of scrapes
- [ ] Average word count ≥200
- [ ] "No output" errors <2%
- [ ] Strategy fallback working correctly
- [ ] Classification accuracy improved (validate with test set)

## Next Phase

Once Phase 1 testing confirms success metrics:
- Move to **Phase 2**: Enhance Layer 1 (return top 3 codes, improve confidence)
- Expected accuracy after Phase 2: 80-85%

