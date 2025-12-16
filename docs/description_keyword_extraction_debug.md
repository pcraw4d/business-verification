# Description Keyword Extraction Debug

## Issue

Cloud Services Inc with description "Cloud computing services" is not extracting keywords from the description.

## Investigation

### Current Flow

1. **Keyword Extraction (Goroutine):**

   - Calls `extractKeywords(ctx, businessName, websiteURL)` - does NOT use description
   - `extractKeywords` calls `keywordRepo.ClassifyBusiness(ctx, businessName, websiteURL)`
   - This may return keywords like "services" from "Cloud Services Inc"
   - If keywords are found, description fallback doesn't trigger

2. **Description Fallback:**

   - Only triggers if `len(keywords) == 0`
   - If `extractKeywords` returns any keywords (even irrelevant ones), fallback is skipped

3. **Enhanced Fallback (Added):**
   - Now also checks description even if keywords exist
   - Merges description keywords with existing keywords
   - But keywords need to make it into `MultiStrategyResult.Keywords`

### Root Cause

The `extractKeywords` function uses `ClassifyBusiness` which only uses business name and website, not description. For "Cloud Services Inc", it might return "services" as a keyword, which prevents the description fallback from triggering.

### Solution Applied

1. ✅ Added logging to see what keywords are extracted
2. ✅ Enhanced fallback to check description even when keywords exist
3. ✅ Merge description keywords with existing keywords
4. ⚠️ Need to verify keywords make it into `MultiStrategyResult.Keywords`

### Next Steps

1. Check logs to see what keywords `extractKeywords` returns
2. Verify description keywords are being extracted
3. Ensure merged keywords make it into final result
4. Test with Cloud Services Inc again
