# Local Testing Guide for Phase 1

This guide explains how to test Phase 1 features locally using Docker Compose, without relying on Railway deployments.

## Prerequisites

1. **Docker** and **Docker Compose** installed
   - Docker Desktop: https://docs.docker.com/get-docker/
   - Verify: `docker --version` and `docker compose version`

2. **Environment Variables**
   - Create `.env` file in project root with Supabase credentials:
   ```bash
   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_ANON_KEY=your_anon_key
   SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
   DATABASE_URL=postgresql://...
   ```

3. **Network Access**
   - Services need internet access to scrape websites
   - Supabase connection required for classification data

## Quick Start

### Option 1: Start All Services (Recommended)

```bash
# Start Playwright and Classification services
./scripts/start-local-services.sh

# Run tests
./scripts/test-phase1-local.sh
```

### Option 2: Start Services Individually

```bash
# Start only required services
docker compose -f docker-compose.local.yml up -d redis-cache playwright-scraper classification-service

# Check service health
curl http://localhost:3000/health  # Playwright
curl http://localhost:8081/health  # Classification
```

## Service Architecture

The local setup includes:

1. **Redis Cache** (`redis-cache:6379`)
   - Used by classification service for caching

2. **Playwright Scraper** (`playwright-scraper:3000`)
   - Node.js service for JavaScript-heavy website scraping
   - Accessible at `http://localhost:3000`
   - Health endpoint: `GET /health`
   - Scrape endpoint: `POST /scrape` with `{"url": "..."}`

3. **Classification Service** (`classification-service:8081`)
   - Main Go service with Phase 1 enhanced scraping
   - Accessible at `http://localhost:8081`
   - Health endpoint: `GET /health`
   - Classify endpoint: `POST /v1/classify`

## Testing Phase 1 Features

### 1. Test Classification Endpoint

```bash
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Business",
    "website_url": "https://example.com"
  }'
```

### 2. Monitor Logs for Phase 1 Markers

```bash
# View classification service logs
docker compose -f docker-compose.local.yml logs -f classification-service

# Look for Phase 1 markers:
# - [Phase1] Starting scraping
# - [Phase1] Strategy: simple_http|browser_headers|playwright
# - [Phase1] Content quality score: X.XX
# - [Phase1] KeywordExtraction: ...
```

### 3. Test Multiple Websites

Use the test script:

```bash
./scripts/test-phase1-local.sh
```

Or manually test different websites:

```bash
# Simple static site
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Example", "website_url": "https://example.com"}'

# JavaScript-heavy site (should use Playwright)
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Microsoft", "website_url": "https://www.microsoft.com"}'
```

## Expected Behavior

### Strategy Fallback

1. **SimpleHTTPScraper** (fastest, ~60% success)
   - Tries first for simple static sites
   - Logs: `[Phase1] Strategy: simple_http`

2. **BrowserHeadersScraper** (realistic headers, ~80% success)
   - Tries if SimpleHTTP fails
   - Logs: `[Phase1] Strategy: browser_headers`

3. **PlaywrightScraper** (JavaScript rendering, ~95% success)
   - Tries if both above fail
   - Logs: `[Phase1] Strategy: playwright`

### Content Quality Metrics

Look for logs showing:
- `Content quality score: X.XX` (should be ≥0.5, ideally ≥0.7)
- `Word count: XXX` (should be ≥50, ideally ≥200)
- `Has title: true/false`
- `Has meta description: true/false`

### Success Criteria

- ✅ Scrape success rate ≥95%
- ✅ Content quality score ≥0.7 for 90%+ of scrapes
- ✅ Average word count ≥200
- ✅ Phase 1 logs appear in classification service
- ✅ Strategy fallback working (logs show which strategy succeeded)

## Troubleshooting

### Services Won't Start

```bash
# Check Docker is running
docker ps

# Check logs
docker compose -f docker-compose.local.yml logs

# Restart services
docker compose -f docker-compose.local.yml restart
```

### Playwright Service Issues

```bash
# Check Playwright service logs
docker compose -f docker-compose.local.yml logs playwright-scraper

# Test Playwright directly
curl -X POST http://localhost:3000/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Classification Service Issues

```bash
# Check classification service logs
docker compose -f docker-compose.local.yml logs classification-service

# Verify environment variables
docker compose -f docker-compose.local.yml exec classification-service env | grep -E "SUPABASE|PLAYWRIGHT"

# Test health endpoint
curl http://localhost:8081/health
```

### No Phase 1 Logs Appearing

1. **Check PLAYWRIGHT_SERVICE_URL is set:**
   ```bash
   docker compose -f docker-compose.local.yml exec classification-service env | grep PLAYWRIGHT
   ```
   Should show: `PLAYWRIGHT_SERVICE_URL=http://playwright-scraper:3000`

2. **Check logs for errors:**
   ```bash
   docker compose -f docker-compose.local.yml logs classification-service | grep -i error
   ```

3. **Verify keyword extraction path:**
   - Ensure `SupabaseKeywordRepository` is using the Phase 1 enhanced scraper
   - Check that `NewSupabaseKeywordRepositoryWithScraper` is being called

## Stopping Services

```bash
# Stop all services
docker compose -f docker-compose.local.yml down

# Stop and remove volumes (clean slate)
docker compose -f docker-compose.local.yml down -v
```

## Next Steps

After verifying Phase 1 works locally:

1. Test with your actual test dataset
2. Measure success metrics (scrape rate, quality scores)
3. Compare with pre-Phase 1 baseline
4. Proceed to Railway deployment once local testing passes

## Additional Resources

- [Phase 1 Testing Guide](./phase1-testing-guide.md)
- [Phase 1 Implementation Summary](./phase1-implementation-summary.md)
- [Phase 1 Log Analysis Guide](./phase1-log-analysis-guide.md)

