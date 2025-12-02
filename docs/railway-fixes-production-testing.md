# Railway Logs Fixes - Production Testing Guide

**Date**: December 2, 2025  
**Status**: Testing in Progress  
**Deployment**: Production

---

## Overview

This document tracks the testing of Railway logs fixes deployed to production:
- **Phase 1**: ML Service Timeout Fix (3s → 5s)
- **Phase 2 & 3**: Content Quality Thresholds (Fast-path mode)

---

## Testing Approach

### 1. Monitor Railway Logs

Check Railway logs for the following indicators:

#### Phase 1 Success Indicators:
- ✅ No "context deadline exceeded" errors for lightweight model
- ✅ Lightweight model success messages
- ✅ Reduced fallback to full model
- ✅ Request completion times < 10s for fast-path

#### Phase 2 & 3 Success Indicators:
- ✅ More pages passing quality checks
- ✅ Reduced "Insufficient keywords" warnings
- ✅ Reduced "Insufficient relevance" warnings
- ✅ Fast-path mode using lenient thresholds (5 keywords, 0.5 relevance)

### 2. Test Commands

#### Check Service Health
```bash
curl https://api-gateway-service-production-21fd.up.railway.app/health
curl https://classification-service-production.up.railway.app/health
```

#### Make Test Classification Request
```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A technology company",
    "website_url": "https://example.com"
  }'
```

#### Monitor Railway Logs
```bash
# View classification service logs
railway logs --service classification-service

# Filter for timeout errors
railway logs --service classification-service | grep -i "timeout\|deadline"

# Filter for fast-path indicators
railway logs --service classification-service | grep -i "fast-path\|lightweight"

# Filter for content quality checks
railway logs --service classification-service | grep -i "contentcheck\|insufficient"
```

---

## Success Criteria

### Phase 1: ML Service Timeout Fix

**Target Metrics:**
- ✅ Lightweight model success rate > 90%
- ✅ No "context deadline exceeded" errors
- ✅ Request completion times improved
- ✅ Reduced fallback to full model

**Log Patterns to Look For:**
```
✅ Fast classification completed in <5s
✅ Lightweight model success
❌ context deadline exceeded (should NOT appear)
```

### Phase 2 & 3: Content Quality Thresholds

**Target Metrics:**
- ✅ 50%+ improvement in pages passing quality checks
- ✅ Classification accuracy maintained
- ✅ Reduced "Insufficient" warnings

**Log Patterns to Look For:**
```
✅ [FAST-PATH] Sufficient content: keywords=5+ (was 10+)
✅ [FAST-PATH] Sufficient content: relevance=0.5+ (was 0.7+)
✅ [ContentCheck] [FAST-PATH] Sufficient content
```

---

## Monitoring Dashboard Queries

### Check for Timeout Errors
```sql
-- Check Railway logs for timeout errors
SELECT COUNT(*) as timeout_errors
FROM logs
WHERE message LIKE '%context deadline exceeded%'
  AND timestamp > NOW() - INTERVAL '1 hour';
```

### Check Fast-Path Success Rate
```sql
-- Check for fast-path success indicators
SELECT COUNT(*) as fast_path_success
FROM logs
WHERE message LIKE '%Fast classification completed%'
  AND timestamp > NOW() - INTERVAL '1 hour';
```

### Check Content Quality Pass Rate
```sql
-- Check content quality check results
SELECT 
  COUNT(*) FILTER (WHERE message LIKE '%Sufficient content%') as passed,
  COUNT(*) FILTER (WHERE message LIKE '%Insufficient%') as failed
FROM logs
WHERE message LIKE '%ContentCheck%'
  AND timestamp > NOW() - INTERVAL '1 hour';
```

---

## Manual Testing Checklist

### Phase 1 Testing
- [ ] Make classification request with website URL
- [ ] Check logs for "Fast classification completed" message
- [ ] Verify no "context deadline exceeded" errors
- [ ] Check response time is < 10s
- [ ] Verify lightweight model is being used

### Phase 2 & 3 Testing
- [ ] Make classification request with minimal content
- [ ] Check logs for fast-path mode indicators
- [ ] Verify content quality checks use lenient thresholds
- [ ] Check for "Sufficient content" messages with lower thresholds
- [ ] Verify classification accuracy is maintained

---

## Troubleshooting

### If Timeout Errors Persist
1. Check ML service health
2. Verify timeout configuration is deployed
3. Check network latency
4. Review ML service logs

### If Content Quality Checks Still Failing
1. Verify fast-path mode is being triggered
2. Check threshold values in logs
3. Verify useFastPath parameter is being passed correctly
4. Review content quality check logic

---

## Next Steps

1. **Monitor for 24 hours** after deployment
2. **Collect metrics** on:
   - Timeout error rate
   - Fast-path success rate
   - Content quality pass rate
   - Request completion times
3. **Compare with baseline** from before fixes
4. **Adjust thresholds** if needed based on results

---

## Test Results

### Initial Testing (Date: ___________)

**Phase 1 Results:**
- Timeout Errors: ___
- Fast-Path Success: ___/___
- Average Response Time: ___s

**Phase 2 & 3 Results:**
- Content Quality Pass Rate: ___%
- "Insufficient" Warnings: ___
- Classification Accuracy: ___%

**Overall Status:** ⬜ Pass ⬜ Partial ⬜ Fail

---

**Last Updated**: December 2, 2025  
**Next Review**: After 24 hours of monitoring

