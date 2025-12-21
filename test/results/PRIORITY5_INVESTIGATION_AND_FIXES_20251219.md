# Priority 5: Investigation and Additional Fixes

## December 19, 2025

---

## Overview

This document details the investigation and fixes implemented to address the remaining issues:

1. **Entertainment fix investigation** - Added comprehensive logging to debug why fix isn't working
2. **Coca-Cola classification fix** - Added "beverage manufacturing" pattern and fallback search

---

## Fix 1: Entertainment Investigation and Enhanced Logging

### Problem

- Entertainment fix not taking effect
- Netflix and Disney still falling back to "General Business"
- Need to understand why Entertainment industry isn't being found

### Solution

- Added comprehensive logging throughout Entertainment fix logic:
  - Log when Entertainment keywords are detected
  - Log when Entertainment keywords are NOT detected
  - Log which Entertainment keywords matched
  - Log when checking industry matches
  - Log when searching all industries
  - Log when Entertainment industry is found/not found
  - Log score comparisons

### Code Changes

```go
// Log Entertainment keyword detection
if hasEntertainmentKeywords {
    r.logger.Printf("🎬 [Priority 5.3] Entertainment keywords detected: %v (from input keywords: %v)", matchedEntertainmentKeywords, keywords)
} else {
    r.logger.Printf("🎬 [Priority 5.3] No Entertainment keywords found in input keywords: %v", keywords)
}

// Log industry search
r.logger.Printf("🎬 [Priority 5.3] Checking for Entertainment industry in %d matched industries", len(industryMatches))

// Log when Entertainment industry found
r.logger.Printf("🎬 [Priority 5.3] ✅ Found Entertainment industry via full search: %s (ID: %d)", industry.Name, industry.ID)

// Log when Entertainment industry not found
r.logger.Printf("🎬 [Priority 5.3] ❌ Entertainment industry not found in any of %d industries", len(allIndustries))
```

### Expected Impact

- Will help identify why Entertainment fix isn't working
- Will show if keywords are being extracted
- Will show if Entertainment industry exists in database
- Will show if fallback search is working

---

## Fix 2: Coca-Cola Classification Fix

### Problem

- Coca-Cola ("Beverage manufacturing") → "General Business"
- "Beverage manufacturing" may not match Food & Beverage keywords
- May not be classified as Manufacturing, so fix doesn't apply

### Solution

1. **Added "beverage manufacturing" keyword patterns**:

   - Added to `service.go` keyword patterns
   - Added to `enhanced_website_scraper.go` regex patterns
   - Added to `supabase_repository.go` Food & Beverage keywords list

2. **Enhanced keyword detection**:

   - Check if both "beverage" and "manufacturing" are in keywords
   - If both present, treat as Food & Beverage keywords

3. **Added fallback search for Food & Beverage**:

   - If falling back to General Business but have Food & Beverage keywords, search all industries
   - Similar to Entertainment fix

4. **Added comprehensive logging**:
   - Log when Food & Beverage keywords are detected
   - Log when "beverage manufacturing" pattern is detected
   - Log when searching for Food & Beverage industry
   - Log score comparisons

### Code Changes

**service.go**:

```go
"food_beverage": {
    // ... existing keywords ...
    "beverage manufacturing", "soft drink", "soda", "juice", "bottled beverage", "carbonated", "cola"
}
```

**enhanced_website_scraper.go**:

```go
// Added to regex pattern:
`\b(...|beverage|beverages|beverage manufacturing|soft drink|soda|juice|bottled beverage|carbonated|cola|...)\b`
```

**supabase_repository.go**:

```go
// Enhanced keyword detection
if !hasFoodBeverageKeywords && len(keywords) > 0 {
    // Check if both "beverage" and "manufacturing" are in keywords
    hasBeverage := false
    hasManufacturing := false
    for _, k := range keywords {
        kl := strings.ToLower(k)
        if strings.Contains(kl, "beverage") {
            hasBeverage = true
        }
        if strings.Contains(kl, "manufacturing") {
            hasManufacturing = true
        }
    }
    if hasBeverage && hasManufacturing {
        hasFoodBeverageKeywords = true
        matchedFoodBeverageKeywords = append(matchedFoodBeverageKeywords, "beverage manufacturing")
    }
}

// Added fallback search
if hasFoodBeverageKeywords && bestIndustryID == 26 {
    // Search all industries for Food & Beverage
    allIndustries, err := r.GetAllIndustries(ctx)
    // ... search logic ...
}
```

### Expected Impact

- Food & Beverage accuracy: 67% → 100%
- Coca-Cola should now classify correctly as Food & Beverage

---

## Files Modified

1. **internal/classification/repository/supabase_repository.go**

   - Added comprehensive logging for Entertainment fix
   - Enhanced Food & Beverage keyword detection
   - Added "beverage manufacturing" pattern detection
   - Added fallback search for Food & Beverage

2. **internal/classification/service.go**

   - Added "beverage manufacturing" and related keywords to Food & Beverage patterns

3. **internal/classification/enhanced_website_scraper.go**
   - Added "beverage manufacturing" and related keywords to regex pattern

---

## Testing Plan

### Test Cases

1. **Entertainment** (with logging):

   - Netflix: "Streaming entertainment services"
   - Disney: "Entertainment and media"
   - Check logs to see:
     - Are Entertainment keywords extracted?
     - Is Entertainment industry found?
     - Is fallback search working?

2. **Food & Beverage** (Coca-Cola):
   - Coca-Cola: "Beverage manufacturing"
   - Expected: Should classify as Food & Beverage
   - Check logs to see:
     - Is "beverage manufacturing" pattern detected?
     - Is Food & Beverage industry found?

### Verification Steps

1. Deploy fixes to Railway
2. Run accuracy tests: `./test/scripts/analyze_classification_accuracy.sh`
3. Check Railway logs for Entertainment and Food & Beverage logging
4. Verify Entertainment accuracy: Should improve (or logs will show why not)
5. Verify Food & Beverage accuracy: Should be 100% (Coca-Cola should work)

---

## Expected Results

### Before Fixes

| Industry        | Accuracy  | Issues                             |
| --------------- | --------- | ---------------------------------- |
| Entertainment   | 0% (0/2)  | Netflix, Disney → General Business |
| Food & Beverage | 67% (2/3) | Coca-Cola → General Business       |

### After Fixes (Expected)

| Industry        | Accuracy      | Improvement                        |
| --------------- | ------------- | ---------------------------------- |
| Entertainment   | 80-100% (2/2) | +80-100% (or logs will show issue) |
| Food & Beverage | 100% (3/3)    | +33%                               |

### Overall Accuracy Impact

- **Before**: 75% (15/20)
- **After (Expected)**: 85-90% (17-18/20)
- **Improvement**: +10-15%

---

## Next Steps

1. ✅ **Code Changes**: All fixes implemented
2. ⏳ **Testing**: Deploy and test fixes
3. ⏳ **Log Analysis**: Review Railway logs for Entertainment and Food & Beverage
4. ⏳ **Verification**: Run accuracy tests and verify improvements

---

## Status

**Status**: ✅ **FIXES IMPLEMENTED** - Ready for testing

**Next Action**: Deploy to Railway, run tests, and analyze logs

---

**Date**: December 19, 2025  
**Priority**: High  
**Impact**: Expected +10-15% overall accuracy improvement
