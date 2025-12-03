# Local Testing Quick Start

## Step 1: Start the Service

Open a terminal and run:

```bash
./scripts/start-classification-service.sh
```

**Expected output:**
```
✅ Loading environment from .env file
✅ Mapped SUPABASE_API_KEY to SUPABASE_ANON_KEY
✅ Environment configured
🚀 Starting Classification Service
✅ Phase 1 enhanced website scraper initialized for keyword extraction
✅ Classification repository initialized with Phase 1 enhanced scraper
🚀 Classification Service listening on :8081
```

**Keep this terminal open** - the service runs in the foreground.

## Step 2: Run Tests

Open **another terminal** and run:

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

## Step 3: Watch the Logs

In the **first terminal** (where the service is running), you should see:

```
🌐 [Phase1] [KeywordExtraction] Starting enhanced website scraping for: https://example.com
🔍 [Phase1] Attempting scrape strategy: simple_http
✅ [Phase1] Strategy succeeded: simple_http
📊 [Phase1] Content quality score: 0.85
✅ [Phase1] [KeywordExtraction] Successfully extracted 15 keywords in 234ms
```

## Troubleshooting

### Service won't start
- Check that port 8081 is available: `lsof -ti:8081`
- Verify .env file exists and has Supabase credentials
- Check Go is installed: `go version`

### No Phase 1 logs
- Ensure `LOG_LEVEL=debug` is set
- Make sure requests include `website_url` parameter
- Check that service is using latest code

### Port conflicts
- Change PORT in .env: `export PORT=8082`
- Update SERVICE_URL in test script: `export SERVICE_URL=http://localhost:8082`

