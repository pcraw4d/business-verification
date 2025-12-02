# Phase 1 Log Analysis Guide

## How to Access Railway Logs

1. Go to [Railway Dashboard](https://railway.app)
2. Select your project
3. Click on **Classification Service**
4. Go to **Logs** tab
5. Filter logs by time range (last hour/day)

## Key Log Patterns to Search For

### 1. Scraping Strategy Logs

Search for: `"Starting scrape with structured content extraction"`

**Example log entry:**
```json
{
  "level": "info",
  "msg": "Starting scrape with structured content extraction",
  "url": "https://example.com",
  "timestamp": "2025-12-02T21:00:00Z"
}
```

### 2. Strategy Attempts

Search for: `"Attempting scrape strategy"`

**Example:**
```json
{
  "level": "info",
  "msg": "Attempting scrape strategy",
  "strategy": "simple_http",
  "url": "https://example.com",
  "attempt": 1
}
```

### 3. Strategy Success

Search for: `"Strategy succeeded"`

**Example:**
```json
{
  "level": "info",
  "msg": "Strategy succeeded",
  "strategy": "simple_http",
  "quality_score": 0.85,
  "word_count": 342,
  "strategy_duration_ms": 1234,
  "total_duration_ms": 1234
}
```

### 4. Strategy Failure

Search for: `"Strategy failed, trying next"`

**Example:**
```json
{
  "level": "warn",
  "msg": "Strategy failed, trying next",
  "strategy": "simple_http",
  "error": "HTTP error: 403 Forbidden",
  "quality_score": 0.0,
  "strategy_duration_ms": 500
}
```

### 5. All Strategies Failed

Search for: `"All scraping strategies failed"`

**Example:**
```json
{
  "level": "error",
  "msg": "All scraping strategies failed",
  "url": "https://example.com",
  "error": "all scraping strategies failed: HTTP error: 403",
  "total_duration_ms": 5000
}
```

## Metrics Extraction

### Using Railway Logs UI

1. Use the search/filter box
2. Search for specific patterns
3. Export logs if needed
4. Count occurrences

### Using Command Line (if Railway CLI available)

```bash
railway logs --service classification-service | grep "Strategy succeeded" | wc -l
railway logs --service classification-service | grep "quality_score" | jq '.quality_score'
```

## Calculating Metrics

### Scrape Success Rate

1. Count total scraping attempts:
   - Search: `"Starting scrape with structured content extraction"`
   - Count results

2. Count successful scrapes:
   - Search: `"Strategy succeeded"`
   - Count results

3. Calculate: `(successful / total) * 100`
   - Target: ≥95%

### Quality Score Analysis

1. Extract all quality scores:
   - Search: `"quality_score"`
   - Extract the numeric values

2. Calculate:
   - Average quality score
   - Count how many are ≥0.7
   - Percentage with ≥0.7
   - Target: ≥0.7 for 90%+ of scrapes

### Word Count Analysis

1. Extract all word counts:
   - Search: `"word_count"`
   - Extract the numeric values

2. Calculate:
   - Average word count
   - Target: ≥200

### Strategy Distribution

1. Count each strategy:
   - Search: `"strategy": "simple_http"` → Count
   - Search: `"strategy": "browser_headers"` → Count
   - Search: `"strategy": "playwright"` → Count

2. Calculate percentages:
   - SimpleHTTP: (count / total) * 100
   - BrowserHeaders: (count / total) * 100
   - Playwright: (count / total) * 100

3. Expected:
   - SimpleHTTP: ~60%
   - BrowserHeaders: ~20-30%
   - Playwright: ~10-20%

## Sample Log Queries

### Find all successful scrapes with quality scores
```
"Strategy succeeded" AND "quality_score"
```

### Find Playwright usage
```
"strategy": "playwright"
```

### Find failed scrapes
```
"All scraping strategies failed"
```

### Find high quality scrapes
```
"quality_score" AND (quality_score >= 0.7)
```

## Troubleshooting Common Issues

### Issue: No scraping logs found
**Possible causes:**
- Service not using enhanced scraper
- Logs not being generated
- Wrong service/logs being checked

**Solution:**
- Verify `PLAYWRIGHT_SERVICE_URL` is set
- Check if service is using latest code
- Verify logging configuration

### Issue: All strategies failing
**Possible causes:**
- Network issues
- Bot detection too aggressive
- Playwright service down

**Solution:**
- Check Playwright service health
- Verify network connectivity
- Check for rate limiting

### Issue: Low quality scores
**Possible causes:**
- Poor content extraction
- Websites returning minimal content
- Extraction functions not working

**Solution:**
- Check if structured content is being extracted
- Verify HTML parsing is working
- Test extraction functions individually

## Reporting Results

Create a report with:

1. **Test Date/Time**
2. **Total Tests:** Number of classification requests
3. **Scrape Success Rate:** Percentage
4. **Average Quality Score:** Number
5. **Quality Score Distribution:** How many ≥0.7
6. **Average Word Count:** Number
7. **Strategy Distribution:** Percentages for each
8. **Issues Found:** Any problems
9. **Recommendations:** Next steps

## Example Report Format

```markdown
# Phase 1 Testing Results - [Date]

## Summary
- Total Tests: 50
- Successful Scrapes: 48 (96%)
- Failed Scrapes: 2 (4%)

## Metrics
- Scrape Success Rate: 96% ✅ (Target: ≥95%)
- Average Quality Score: 0.78 ✅ (Target: ≥0.7)
- Quality Scores ≥0.7: 45/48 (94%) ✅ (Target: 90%+)
- Average Word Count: 312 ✅ (Target: ≥200)

## Strategy Distribution
- SimpleHTTP: 30 (60%) ✅
- BrowserHeaders: 12 (24%) ✅
- Playwright: 6 (12%) ✅

## Issues
- 2 timeouts on JavaScript-heavy sites (resolved with Playwright)

## Conclusion
Phase 1 targets met. Ready for Phase 2.
```

