# Priority 5: Classification Accuracy Fixes
## December 19, 2025

---

## Overview

This document details the fixes implemented to address three critical classification accuracy issues identified in post-deployment testing:

1. **Entertainment keyword matching** - Entertainment businesses falling back to "General Business"
2. **Food & Beverage vs Retail** - Food & Beverage businesses misclassified as Retail
3. **Healthcare vs Insurance** - Healthcare businesses misclassified as Insurance

---

## Issue Analysis

### 1. Entertainment Keyword Matching

**Problem**: 
- Netflix ("Streaming entertainment services") → "General Business"
- Disney ("Entertainment and media") → "General Business"
- Both cases had Entertainment keywords but fell back to default

**Root Cause**:
- Entertainment keywords are extracted but may not match in database keyword index
- Entertainment industry may not have sufficient keyword weights in database
- Early exit or low confidence causing fallback to "General Business"

**Test Cases**:
- Test 7: Netflix → Expected "Entertainment", Got "General Business" (confidence: 0.60)
- Test 13: Disney → Expected "Entertainment", Got "General Business" (confidence: 0.60)

### 2. Food & Beverage vs Retail

**Problem**:
- Starbucks ("Coffee retail and food service") → "Retail"
- "Retail" keyword matching Retail industry, overriding Food & Beverage keywords

**Root Cause**:
- Both "retail" and "food service" keywords present
- Retail industry scoring higher than Food & Beverage
- No prioritization logic for Food & Beverage when both keywords present

**Test Cases**:
- Test 4: Starbucks → Expected "Food & Beverage", Got "Retail" (confidence: 0.95)

### 3. Healthcare vs Insurance

**Problem**:
- One Healthcare test case → "Insurance" (confidence: 0.95)
- Healthcare businesses mentioning "insurance" (e.g., "health insurance") being misclassified

**Root Cause**:
- "Insurance" keyword in financial keywords list
- Healthcare businesses may mention "insurance" in descriptions
- No distinction logic between Healthcare and Insurance industries

**Test Cases**:
- Test 17: Healthcare → Expected "Healthcare", Got "Insurance" (confidence: 0.95)

---

## Solution Implementation

### Fix 1: Entertainment Keyword Prioritization

**Location**: `internal/classification/repository/supabase_repository.go` (line ~2876)

**Implementation**:
- Added post-processing logic to detect Entertainment keywords in input
- Boost Entertainment industry score by 50% when Entertainment keywords are present
- Check all matched industries for Entertainment industry and boost if found

**Code Logic**:
```go
// Entertainment keywords list
entertainmentKeywords := []string{"entertainment", "streaming", "media", "video", ...}

// Detect Entertainment keywords in input
hasEntertainmentKeywords := false
for _, kw := range keywords {
    // Check if keyword matches Entertainment patterns
    if matchesEntertainment(kw) {
        hasEntertainmentKeywords = true
        break
    }
}

// Boost Entertainment industry if keywords present
if hasEntertainmentKeywords {
    for industryID, matched := range industryMatches {
        if industry.Name contains "entertainment" {
            entertainmentScore *= 1.5 // 50% boost
            if entertainmentScore > bestScore {
                bestIndustryID = industryID // Switch to Entertainment
            }
        }
    }
}
```

**Expected Impact**:
- Entertainment accuracy: 0% → 80-100%
- Netflix and Disney should now classify correctly

### Fix 2: Food & Beverage vs Retail Prioritization

**Location**: `internal/classification/repository/supabase_repository.go` (line ~2912)

**Implementation**:
- Added post-processing logic to detect Food & Beverage keywords
- When Retail is winning but Food & Beverage keywords are present, boost Food & Beverage
- Allow Food & Beverage to win even if score is slightly lower (90% threshold)

**Code Logic**:
```go
// Food & Beverage keywords list
foodBeverageKeywords := []string{"restaurant", "cafe", "coffee", "food", ...}

// Detect Food & Beverage keywords
hasFoodBeverageKeywords := false
for _, kw := range keywords {
    if matchesFoodBeverage(kw) {
        hasFoodBeverageKeywords = true
        break
    }
}

// Boost Food & Beverage if Retail is winning
if hasFoodBeverageKeywords && bestIndustry.Name contains "retail" {
    for industryID, matched := range industryMatches {
        if industry.Name contains "food" or "beverage" or "restaurant" {
            foodBeverageScore *= 1.4 // 40% boost
            if foodBeverageScore > bestScore * 0.9 {
                bestIndustryID = industryID // Switch to Food & Beverage
            }
        }
    }
}
```

**Expected Impact**:
- Food & Beverage accuracy: 33% → 80-100%
- Starbucks should now classify as Food & Beverage

### Fix 3: Healthcare vs Insurance Distinction

**Location**: `internal/classification/repository/supabase_repository.go` (line ~2950)

**Implementation**:
- Added post-processing logic to detect Healthcare keywords
- When Insurance is winning but Healthcare keywords are present, boost Healthcare
- Allow Healthcare to win even if score is slightly lower (85% threshold)

**Code Logic**:
```go
// Healthcare keywords list
healthcareKeywords := []string{"healthcare", "health", "medical", "hospital", ...}

// Detect Healthcare keywords
hasHealthcareKeywords := false
for _, kw := range keywords {
    if matchesHealthcare(kw) {
        hasHealthcareKeywords = true
        break
    }
}

// Boost Healthcare if Insurance is winning
if hasHealthcareKeywords && bestIndustry.Name contains "insurance" {
    for industryID, matched := range industryMatches {
        if industry.Name contains "healthcare" or "health" or "medical" {
            healthcareScore *= 1.5 // 50% boost
            if healthcareScore > bestScore * 0.85 {
                bestIndustryID = industryID // Switch to Healthcare
            }
        }
    }
}
```

**Expected Impact**:
- Healthcare accuracy: 66.7% → 100%
- Healthcare businesses mentioning "insurance" should classify correctly

---

## Implementation Details

### Keyword Lists

**Entertainment Keywords** (37 keywords):
- Core: entertainment, media, streaming, video, audio, podcast
- Content: music, film, movie, cinema, television, tv, broadcasting
- Creative: content, creative, art, gaming, game, esports
- Events: sports, events, concert, festival, theater, theatre, performance, show
- Production: production, studio, record, label, artist, actor, director, producer

**Food & Beverage Keywords** (38 keywords):
- Establishments: restaurant, restaurants, cafe, cafes, coffee, coffee shop
- Service: food, dining, kitchen, catering, bakery, bar, pub
- Production: brewery, winery, wine, beer, cocktail, beverage, drink, alcohol, spirits, liquor
- Dining Types: menu, chef, cook, cuisine, delivery, takeout, fast food, casual dining, fine dining
- Venues: bistro, eatery, diner, tavern, gastropub, food truck

**Healthcare Keywords** (27 keywords):
- Core: healthcare, health, medical, hospital, clinic, doctor, physician
- Services: patient, care, treatment, pharmacy, diagnostic, therapy
- Specialties: wellness, dental, vision, mental, psychology, counseling, rehabilitation
- Emergency: emergency, ambulance, laboratory, radiology, imaging, surgery, nursing

### Boost Factors

| Fix | Boost Factor | Threshold | Rationale |
|-----|--------------|-----------|-----------|
| Entertainment | 1.5x (50%) | None | Strong keyword match should always win |
| Food & Beverage | 1.4x (40%) | 90% of Retail score | Food & Beverage more specific than Retail |
| Healthcare | 1.5x (50%) | 85% of Insurance score | Healthcare more specific than Insurance |

---

## Expected Results

### Before Fixes

| Industry | Accuracy | Issues |
|----------|----------|--------|
| Entertainment | 0% (0/2) | Netflix, Disney → General Business |
| Food & Beverage | 33% (1/3) | Starbucks → Retail, Coca-Cola → General Business |
| Healthcare | 66.7% (2/3) | 1 case → Insurance |

### After Fixes (Expected)

| Industry | Accuracy | Improvement |
|----------|----------|-------------|
| Entertainment | 80-100% (2/2) | +80-100% |
| Food & Beverage | 80-100% (3/3) | +47-67% |
| Healthcare | 100% (3/3) | +33.3% |

### Overall Accuracy Impact

- **Before**: 65% (13/20)
- **After (Expected)**: 85-90% (17-18/20)
- **Improvement**: +20-25%

---

## Testing Plan

### Test Cases

1. **Entertainment**:
   - Netflix: "Streaming entertainment services"
   - Disney: "Entertainment and media"
   - Expected: Both should classify as "Entertainment"

2. **Food & Beverage**:
   - Starbucks: "Coffee retail and food service"
   - Coca-Cola: "Beverage manufacturing"
   - McDonald's: "Fast food restaurant chain"
   - Expected: All should classify as "Food & Beverage"

3. **Healthcare**:
   - Healthcare business mentioning "insurance"
   - Expected: Should classify as "Healthcare", not "Insurance"

### Verification Steps

1. Deploy fixes to Railway
2. Run accuracy analysis script: `./test/scripts/analyze_classification_accuracy.sh`
3. Verify Entertainment accuracy: 80-100%
4. Verify Food & Beverage accuracy: 80-100%
5. Verify Healthcare accuracy: 100%
6. Verify overall accuracy: 85-90%

---

## Files Modified

1. **internal/classification/repository/supabase_repository.go**
   - Added Entertainment keyword prioritization (line ~2876)
   - Added Food & Beverage vs Retail prioritization (line ~2912)
   - Added Healthcare vs Insurance distinction (line ~2950)

---

## Next Steps

1. ✅ **Code Changes**: Implemented all three fixes
2. ⏳ **Testing**: Deploy and test fixes
3. ⏳ **Verification**: Run accuracy analysis script
4. ⏳ **Optimization**: Fine-tune boost factors if needed

---

## Status

**Status**: ✅ **FIXES IMPLEMENTED** - Ready for testing

**Next Action**: Deploy to Railway and run accuracy tests

---

**Date**: December 19, 2025  
**Priority**: High  
**Impact**: Expected +20-25% overall accuracy improvement

