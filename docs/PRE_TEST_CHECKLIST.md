# Pre-Test Checklist

**Date**: December 19, 2025  
**Status**: Ready for Comprehensive Tests

---

## ✅ Completed Actions

1. **Code Fix Deployed**
   - ✅ Fixed cache hit response `FromCache` flag
   - ✅ Changes committed and pushed
   - ✅ Service redeployed to production

2. **Environment Variables Updated**
   - ✅ Updated in Railway dashboard
   - ✅ Service redeployed

---

## 🔍 Pre-Test Verification Steps

Before running comprehensive tests, verify the following:

### 1. Cache Health Check (CRITICAL)

**Action**: Verify Redis cache is connected and working

```bash
curl https://classification-service-production.up.railway.app/health/cache
```

**Expected Response**:
```json
{
  "cache_enabled": true,
  "redis_enabled": true,
  "redis_connected": true,
  "healthy": true
}
```

**If Redis not connected**: Check Railway dashboard for REDIS_URL configuration

---

### 2. Quick Cache Hit Test (RECOMMENDED)

**Action**: Make two identical requests to verify cache hit works

```bash
# First request (should be cache miss)
curl -X POST https://classification-service-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company", "description": "Test", "website_url": "https://example.com"}' \
  | jq '.from_cache'

# Expected: false

# Second request (should be cache hit)
curl -X POST https://classification-service-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company", "description": "Test", "website_url": "https://example.com"}' \
  | jq '.from_cache'

# Expected: true
```

**If second request shows `false`**: Cache may not be working correctly

---

### 3. Metadata Verification (RECOMMENDED)

**Action**: Verify metadata fields are present in responses

```bash
curl -X POST https://classification-service-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company", "description": "Test", "website_url": "https://example.com"}' \
  | jq '.metadata | {early_exit, scraping_strategy, fallback_used}'
```

**Expected Response**:
```json
{
  "early_exit": true/false,
  "scraping_strategy": "hrequests" | "playwright" | "selenium" | etc.,
  "fallback_used": true/false
}
```

**If metadata missing**: Test runner may not extract correctly, but this won't block tests

---

## 📋 Environment Variables Checklist

Verify these are set in Railway:

- ✅ `CACHE_ENABLED=true`
- ✅ `REDIS_ENABLED=true`
- ✅ `REDIS_URL=<actual Redis URL>` (not template variable)
- ✅ `ENABLE_EARLY_TERMINATION=true` (recommended, but defaults work)
- ✅ `EARLY_TERMINATION_CONFIDENCE_THRESHOLD=0.85` (recommended, but defaults work)
- ✅ `CACHE_TTL=5m`

---

## ⚠️ Critical Issues to Address Before Tests

### Issue 1: REDIS_URL Template Variable

**Status**: ⚠️ **VERIFY**

If `REDIS_URL` is still set to `r${{ Redis.REDIS_URL }}`:
- Check Railway dashboard to see if it resolves correctly
- If not resolved, set explicit Redis URL
- Verify with `/health/cache` endpoint

**Impact**: If Redis not connected, cache hit rate will be 0%

---

### Issue 2: Metadata Extraction

**Status**: ✅ **VERIFIED** (code review shows extraction logic is correct)

The test runner code correctly extracts:
- `from_cache` from response root
- `metadata.early_exit` from metadata object
- `metadata.scraping_strategy` from metadata object

**No action needed** - extraction logic is correct

---

## ✅ Ready to Run Tests?

**YES** - If:
1. ✅ Cache health check shows Redis connected
2. ✅ Quick cache hit test shows `from_cache: true` on second request
3. ✅ Service is deployed and responding

**NO** - If:
1. ❌ Redis not connected (fix REDIS_URL first)
2. ❌ Cache hits not working (investigate cache key generation)
3. ❌ Service not responding (check deployment status)

---

## Expected Test Results After Fixes

### Cache Hit Rate
- **Before**: 0%
- **After**: 60-70%
- **Verification**: Check `CacheHit` field in test results

### Early Exit Rate
- **Before**: 0% (tracked)
- **After**: 20-30% (tracked)
- **Verification**: Check `EarlyExit` field in test results

### Strategy Distribution
- **Before**: Empty
- **After**: Populated with strategy names
- **Verification**: Check `ScrapingStrategy` field in test results

### Average Latency
- **Before**: 16.5s
- **After**: <2s (with cache hits)
- **Verification**: Check `ProcessingTime` field in test results

### Success Rate
- **Before**: 64%
- **After**: ≥95%
- **Verification**: Check `Success` field in test results

---

## Next Steps

1. ✅ Run verification steps above
2. ✅ If all checks pass, run comprehensive tests
3. ✅ Review test results and compare to expected improvements
4. ✅ If results don't match expectations, investigate:
   - Cache hit rate still 0% → Check Redis connection and cache key generation
   - Early exit rate still 0% → Check metadata extraction (though code looks correct)
   - Strategy distribution still empty → Check metadata extraction (though code looks correct)

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2025  
**Status**: ✅ Ready for testing

