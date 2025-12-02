# Phase 1 Logging Investigation

## Issue
Phase 1 logs (`[Phase1]`) are not appearing in Railway logs, even though:
- ✅ Phase 1 code is integrated into `EnhancedWebsiteScraper`
- ✅ `PLAYWRIGHT_SERVICE_URL` is set
- ✅ Code has been deployed

## Findings

### 1. Multiple Scraping Paths
The classification service uses **multiple different scrapers**:

1. **`EnhancedWebsiteScraper`** (with Phase 1 features)
   - Used by: ML method (`ml_method.go`)
   - Only called when: `pythonMLService != nil && websiteURL != ""`
   - Path: `MLClassificationMethod.extractWebsiteContent()` → `EnhancedWebsiteScraper.ScrapeWebsite()`

2. **Keyword Extraction Scraper** (different code path)
   - Used by: `SupabaseKeywordRepository.extractKeywordsFromWebsite()`
   - Logs show: `[KeywordExtraction] [SinglePage]` and `[KeywordExtraction] [MultiPage]`
   - This is a **different scraper** that doesn't use Phase 1 features

3. **Smart Crawler** (another path)
   - Used by: `WebsiteContentService` when `smartCrawler` is available
   - Different code path entirely

### 2. Current Logs Show
From Railway logs, we see:
```
[KeywordExtraction] [SinglePage] Starting single-page website scraping
[KeywordExtraction] [MultiPage] Starting multi-page website analysis
```

These are **NOT** using the Phase 1 enhanced scraper.

### 3. Why Phase 1 Isn't Being Used
The Phase 1 enhanced scraper is only used when:
- ML method is active
- Python ML service is configured
- Website URL is provided
- ML method calls `extractWebsiteContent()`

But the **keyword extraction** path (which is what we're seeing in logs) uses a different scraper.

## Solution

We need to ensure Phase 1 scraper is used in **all** scraping paths, not just the ML method path.

### Option 1: Update Keyword Extraction to Use Phase 1 Scraper
- Modify `SupabaseKeywordRepository.extractKeywordsFromWebsite()` to use `EnhancedWebsiteScraper`
- This would require passing the scraper instance to the repository

### Option 2: Update WebsiteContentService to Use Phase 1 Scraper
- The `WebsiteContentService` already uses `external.WebsiteScraper` directly
- But it might not be using the Phase 1 enhanced version with strategies

### Option 3: Test with ML Method Path
- Make a classification request that triggers the ML method
- Ensure `PYTHON_ML_SERVICE_URL` is set
- This should use `EnhancedWebsiteScraper` with Phase 1 features

## Next Steps

1. **Verify ML Method Path:**
   - Check if `PYTHON_ML_SERVICE_URL` is set in Railway
   - Make a test request that would trigger ML method
   - Look for `[Phase1]` logs in that path

2. **Update All Scraping Paths:**
   - Ensure all scrapers use the Phase 1 enhanced scraper
   - Or at least log which scraper is being used

3. **Add Debug Logging:**
   - Log when `EnhancedWebsiteScraper` is initialized
   - Log when Phase 1 scraper is actually called
   - Log which code path is being used

## Test Command

To test the ML method path specifically:
```bash
curl -X POST https://your-classification-service.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A test company",
    "website_url": "https://example.com"
  }'
```

This should trigger the ML method path which uses `EnhancedWebsiteScraper` with Phase 1 features.

