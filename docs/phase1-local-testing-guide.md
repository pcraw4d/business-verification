# Phase 1 Local Testing Guide

## Overview

This guide helps you test Phase 1 enhanced scraper functionality locally before deploying to Railway.

## Prerequisites

1. **Go 1.22+** installed
2. **Supabase credentials** configured
3. **Playwright service URL** (optional, for Playwright strategy testing)

## Setup

### 1. Set Environment Variables

Create a `.env` file in `services/classification-service/` or export variables:

```bash
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your_anon_key"
export SUPABASE_SERVICE_ROLE_KEY="your_service_role_key"
export PORT="8081"
export LOG_LEVEL="debug"  # For detailed Phase 1 logs

# Optional: For Playwright strategy testing
export PLAYWRIGHT_SERVICE_URL="https://playwright-service-production-b21a.up.railway.app"

# Optional: For ML service testing
export PYTHON_ML_SERVICE_URL="https://your-ml-service.railway.app"
```

### 2. Start the Service

```bash
cd services/classification-service
go run cmd/main.go
```

**Expected startup logs:**
```
🚀 Starting Classification Service
✅ Configuration loaded successfully
✅ Classification adapters initialized
✅ Phase 1 enhanced website scraper initialized for keyword extraction
✅ Classification repository initialized with Phase 1 enhanced scraper
🚀 Classification Service listening
```

## Testing

### Quick Test Script

Run the automated test script:

```bash
./scripts/test-phase1-local.sh
```

This will:
- Check service health
- Test multiple websites
- Show request results
- Guide you to check logs for Phase 1 markers

### Manual Testing

#### 1. Test Simple Website (SimpleHTTP Strategy)

```bash
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Restaurant",
    "website_url": "https://example.com"
  }'
```

**Check logs for:**
- `[Phase1] [KeywordExtraction] Starting enhanced website scraping`
- `[Phase1] Attempting scrape strategy: simple_http`
- `[Phase1] Strategy succeeded: simple_http`
- Quality score and word count

#### 2. Test JavaScript-Heavy Website (Playwright Strategy)

```bash
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Modern Web App",
    "website_url": "https://react.dev"
  }'
```

**Check logs for:**
- Strategy fallback: `simple_http` → `browser_headers` → `playwright`
- `[Phase1] Strategy succeeded: playwright`
- Higher quality scores from structured content

#### 3. Test Various Website Types

Test different website types to verify strategy fallback:

```bash
# Static HTML site
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Static Site", "website_url": "https://example.com"}'

# Site with bot detection
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Protected Site", "website_url": "https://www.apple.com"}'

# JavaScript SPA
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "SPA", "website_url": "https://react.dev"}'
```

## What to Look For in Logs

### Phase 1 Initialization

```
✅ Phase 1 enhanced website scraper initialized for keyword extraction
✅ Classification repository initialized with Phase 1 enhanced scraper
```

### Keyword Extraction Logs

```
🌐 [Phase1] [KeywordExtraction] Starting enhanced website scraping for: https://example.com
🔍 [Phase1] Attempting scrape strategy: simple_http
✅ [Phase1] Strategy succeeded: simple_http
✅ [Phase1] [KeywordExtraction] Successfully extracted N keywords in Xms
```

### Strategy Fallback

```
🔍 [Phase1] Attempting scrape strategy: simple_http
⚠️ [Phase1] Strategy failed: simple_http (reason: ...)
🔍 [Phase1] Attempting scrape strategy: browser_headers
✅ [Phase1] Strategy succeeded: browser_headers
```

### Quality Metrics

```
📊 [Phase1] Content quality score: 0.85
📝 [Phase1] Word count: 342
✅ [Phase1] Content validation passed
```

## Success Criteria

### Scrape Success Rate
- **Target:** ≥95%
- **Check:** Count successful scrapes vs failures in logs

### Quality Scores
- **Target:** ≥0.7 for 90%+ of successful scrapes
- **Check:** Look for `quality_score` in logs

### Word Count
- **Target:** Average ≥200 words
- **Check:** Look for `word_count` in logs

### Strategy Distribution
- **Expected:**
  - SimpleHTTP: ~60% of requests
  - BrowserHeaders: ~20-30% of requests
  - Playwright: ~10-20% of requests

## Troubleshooting

### Issue: Service won't start

**Check:**
- Environment variables are set correctly
- Supabase credentials are valid
- Port 8081 is not in use

### Issue: No Phase 1 logs appearing

**Check:**
- Service is using latest code (check git commit)
- `LOG_LEVEL=debug` is set
- Requests include `website_url` parameter
- Service is using keyword extraction path (not ML path)

### Issue: Playwright strategy not working

**Check:**
- `PLAYWRIGHT_SERVICE_URL` is set
- Playwright service is accessible
- Check for connection errors in logs

### Issue: Low quality scores

**Check:**
- Websites are returning proper HTML
- Structured content extraction is working (title, headings, about)
- Content validation thresholds are appropriate

## Next Steps

Once local testing confirms Phase 1 is working:

1. ✅ Verify all success criteria are met
2. ✅ Test with diverse website types
3. ✅ Monitor strategy distribution
4. ✅ Check quality scores and word counts
5. ✅ Proceed with Railway deployment

## Test URLs

Here are some good test URLs for different scenarios:

- **Simple static site:** `https://example.com`
- **Bot protection:** `https://www.apple.com`, `https://www.microsoft.com`
- **JavaScript SPA:** `https://react.dev`, `https://vuejs.org`
- **E-commerce:** `https://www.starbucks.com`
- **Content-heavy:** `https://www.wikipedia.org`

