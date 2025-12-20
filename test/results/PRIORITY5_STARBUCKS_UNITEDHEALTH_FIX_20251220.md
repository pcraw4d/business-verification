# Priority 5: Starbucks and UnitedHealth Group Fix
## December 20, 2025

---

## Issues Identified

### Test Results Analysis
- **Overall Accuracy**: 85% (17/20)
- **Food & Beverage**: 33.3% (1/3) - 2 failures
- **Healthcare**: 66.7% (2/3) - 1 failure

### Specific Failures

#### 1. Starbucks (Test 4)
- **Input**: "Coffee retail and food service"
- **Expected**: Food & Beverage
- **Predicted**: Retail
- **Confidence**: 0.95
- **Processing Path**: layer1 (fast path)
- **Issue**: "retail" keyword matches Retail industry before "coffee" can be checked

#### 2. UnitedHealth Group (Test 17)
- **Input**: "Healthcare insurance"
- **Expected**: Healthcare
- **Predicted**: Insurance
- **Confidence**: 0.95
- **Processing Path**: layer1 (fast path)
- **Issue**: "insurance" keyword matches Insurance industry before "healthcare" can be checked

---

## Root Cause

The fast path in `tryFastPath()` checks obvious keywords in order and matches the **first** high-confidence industry found. This causes:

1. **Starbucks**: "retail" matches Retail industry (0.70+ confidence) → Fast path succeeds with "Retail"
2. **UnitedHealth**: "insurance" matches Insurance industry (0.70+ confidence) → Fast path succeeds with "Insurance"

The problem is that higher-priority keywords (coffee, healthcare) are not checked **before** lower-priority keywords (retail, insurance).

---

## Solution

### Added Keyword Prioritization Logic

**File**: `internal/classification/multi_strategy_classifier.go:tryFastPath()`

**Changes**:
1. **Check for Food & Beverage keywords FIRST** (before processing any industry matches):
   - Keywords: coffee, restaurant, cafe, food, beverage, dining, kitchen, bakery, bar, pub, brewery, winery, pizzeria, diner, bistro, eatery
   - If Food & Beverage keywords are present, **skip** Retail/Manufacturing matches

2. **Check for Healthcare keywords FIRST** (before processing any industry matches):
   - Keywords: healthcare, medical, health, clinic, hospital, physician, doctor, pharmacy, pharmaceutical
   - If Healthcare keywords are present, **skip** Insurance matches

**Logic Flow**:
```
1. Extract obvious keywords
2. Check for Food & Beverage keywords → Set hasFoodBeverageKeywords flag
3. Check for Healthcare keywords → Set hasHealthcareKeywords flag
4. For each keyword:
   a. Get industry matches
   b. If hasFoodBeverageKeywords AND matched Retail/Manufacturing → Skip (continue to next keyword)
   c. If hasHealthcareKeywords AND matched Insurance → Skip (continue to next keyword)
   d. Otherwise → Use this match
```

---

## Expected Impact

### Before Fix

| Business | Description | Predicted | Status |
|----------|-------------|-----------|--------|
| Starbucks | Coffee retail and food service | Retail | ❌ |
| UnitedHealth Group | Healthcare insurance | Insurance | ❌ |

### After Fix

| Business | Description | Predicted | Status |
|----------|-------------|-----------|--------|
| Starbucks | Coffee retail and food service | Food & Beverage | ✅ |
| UnitedHealth Group | Healthcare insurance | Healthcare | ✅ |

---

## Testing Plan

1. **Deploy Fix**: Deploy to Railway
2. **Run Accuracy Tests**: Verify improvements
3. **Expected Results**:
   - Overall accuracy: 85% → **90%** (18/20)
   - Food & Beverage: 33.3% → **66.7%** (2/3)
   - Healthcare: 66.7% → **100%** (3/3)

---

## Files Modified

1. **internal/classification/multi_strategy_classifier.go**
   - Added Food & Beverage keyword detection
   - Added Healthcare keyword detection
   - Added prioritization logic to skip Retail/Insurance matches when higher-priority keywords are present

---

## Related Fixes

- ✅ **Coca-Cola Fix**: Added "beverage" to obviousKeywordMap (commit `9c7fea9fb`)
- ✅ **Starbucks Fix**: Prioritize Food & Beverage keywords over Retail (this commit)
- ✅ **UnitedHealth Fix**: Prioritize Healthcare keywords over Insurance (this commit)

---

**Status**: ✅ **FIX IMPLEMENTED** - Ready for deployment and testing  
**Date**: December 20, 2025

