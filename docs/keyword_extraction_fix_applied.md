# Keyword Extraction Fix Applied

## Issue

Cloud Services Inc with description "Cloud computing services" was not extracting keywords from the description.

## Root Cause

1. **Description extraction only triggered if no keywords from business name:**

   - If `extractKeywords` returned any keywords (even generic "services"), description fallback was skipped
   - Original code: `if len(keywords) == 0 && description != ""`

2. **Keyword filtering might remove relevant keywords:**

   - `filterRelevantKeywords` may filter out keywords that don't match business name context
   - "cloud" and "computing" from description might be filtered if business name is "Cloud Services Inc"

3. **Keywords not prioritized correctly:**
   - Description keywords should take priority over generic business name keywords
   - "cloud computing" is more specific than "services"

## Solution Applied

### Enhanced Description Keyword Extraction

**File:** `internal/classification/multi_strategy_classifier.go`

**Changes:**

1. **Always check description:** Changed from conditional to always checking description if it exists
2. **Prioritize description keywords:** Description keywords are added first in merged list
3. **Better logging:** Added detailed logging to track keyword extraction and merging
4. **Merge logic:** Properly merges description keywords with existing keywords, avoiding duplicates

**Code Changes:**

```go
// Phase 2: Always check description for keywords, even if we have some from business name/website
if description != "" {
    descriptionKeywords := msc.extractObviousKeywords("", description, "")
    if len(descriptionKeywords) > 0 {
        if len(keywords) == 0 {
            keywords = descriptionKeywords
        } else {
            // Merge with description keywords first (higher priority)
            mergedKeywords := []string{}
            keywordMap := make(map[string]bool)
            // Add description keywords first
            for _, kw := range descriptionKeywords {
                lowerKw := strings.ToLower(kw)
                if !keywordMap[lowerKw] {
                    mergedKeywords = append(mergedKeywords, kw)
                    keywordMap[lowerKw] = true
                }
            }
            // Add existing keywords that aren't duplicates
            for _, kw := range keywords {
                lowerKw := strings.ToLower(kw)
                if !keywordMap[lowerKw] {
                    mergedKeywords = append(mergedKeywords, kw)
                    keywordMap[lowerKw] = true
                }
            }
            keywords = mergedKeywords
        }
    }
}
```

## Expected Results

After this fix:

- ✅ Description keywords should always be extracted if description exists
- ✅ Description keywords should take priority over generic business name keywords
- ✅ "Cloud computing services" should extract "cloud" and "computing"
- ✅ Keywords should appear in `key_terms_found` in API response

## Testing

1. Test Cloud Services Inc with description "Cloud computing services"
2. Verify keywords appear in response: `["cloud", "computing"]`
3. Verify industry classification improves (should be Technology, not General Business)
4. Re-run full test suite to verify pass rate improves
