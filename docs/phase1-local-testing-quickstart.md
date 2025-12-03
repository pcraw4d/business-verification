# Phase 1 Local Testing - Quick Start

## Quick Setup

### 1. Set Environment Variables

```bash
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your_anon_key"
export SUPABASE_SERVICE_ROLE_KEY="your_service_role_key"
export PORT="8081"
export LOG_LEVEL="debug"

# Optional: For Playwright strategy
export PLAYWRIGHT_SERVICE_URL="https://playwright-service-production-b21a.up.railway.app"
```

### 2. Start Service

```bash
cd services/classification-service
go run cmd/main.go
```

### 3. Run Tests

In another terminal:

```bash
./scripts/test-phase1-local.sh
```

Or test manually:

```bash
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "website_url": "https://example.com"
  }'
```

## What to Check

Watch the service logs for:

1. **Startup:**
   ```
   ✅ Phase 1 enhanced website scraper initialized for keyword extraction
   ✅ Classification repository initialized with Phase 1 enhanced scraper
   ```

2. **During Request:**
   ```
   🌐 [Phase1] [KeywordExtraction] Starting enhanced website scraping
   🔍 [Phase1] Attempting scrape strategy: simple_http
   ✅ [Phase1] Strategy succeeded: simple_http
   📊 [Phase1] Content quality score: 0.85
   ✅ [Phase1] [KeywordExtraction] Successfully extracted N keywords
   ```

## Success Indicators

- ✅ Service starts without errors
- ✅ Phase 1 initialization logs appear
- ✅ Requests with `website_url` trigger Phase 1 scraper
- ✅ Quality scores ≥0.7
- ✅ Word counts ≥200
- ✅ Strategy fallback works (simple_http → browser_headers → playwright)

