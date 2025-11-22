# Test Fixes - Collapsible Sections Progress

**Date:** 2025-01-27  
**Focus:** Fixing collapsible section handling in RiskAlertsSection and RiskRecommendationsSection

---

## Progress Summary

### RiskAlertsSection
- **Before:** 4/19 passing (21%)
- **After:** 5/19 passing (26%)
- **Improvements:**
  - Fixed button selectors to handle buttons with badge counts
  - Added userEvent for expanding collapsible sections
  - Fixed toast notification expectations
  - Fixed empty state text matching

### RiskRecommendationsSection
- **Before:** 0/19 passing (0% - syntax errors)
- **After:** 10/19 passing (53%)
- **Improvements:**
  - Fixed syntax errors (async/await in waitFor)
  - Fixed API endpoint pattern (`/api/v1/risk/recommendations/:merchantId`)
  - Added userEvent for expanding collapsible sections
  - Fixed test expectations for collapsible content

### Overall Test Suite
- **Before:** 551/717 passing (76.8%)
- **After:** 557/717 passing (77.7%)
- **Improvement:** +6 tests fixed

---

## Key Fixes Applied

### 1. Collapsible Section Handling
- **Issue:** Alerts/recommendations are in collapsible sections that are closed by default
- **Solution:** 
  - Wait for component to load
  - Find and click collapsible trigger buttons (severity/priority labels)
  - Wait for content to appear after expansion
  - Then assert on the content

### 2. Button Selector Fixes
- **Issue:** Button selectors like `/^high$/i` were too strict
- **Solution:** Use more flexible selectors that account for badge counts and additional text
  ```typescript
  const highButtons = screen.getAllByRole('button');
  const highButton = highButtons.find(btn => 
    btn.textContent?.includes('High') && btn.textContent?.includes('1')
  ) || screen.getByRole('button', { name: /high/i });
  ```

### 3. API Endpoint Pattern Fixes
- **Issue:** MSW handlers used incorrect endpoint patterns
- **Solution:** Aligned with actual API endpoint formats:
  - RiskAlertsSection: `*/api/v1/risk/indicators/:merchantId`
  - RiskRecommendationsSection: `*/api/v1/risk/recommendations/:merchantId`

### 4. Toast Notification Expectations
- **Issue:** Toast expectations didn't match actual implementation
- **Solution:** Updated to match actual toast messages:
  - Critical: `'1 Critical Alert'` with description `'Immediate attention required'`
  - High: `'1 High Priority Alert'` with description `'Review recommended'`

---

## Remaining Issues

### RiskAlertsSection (14 failures)
- Timeout issues in several tests (auto-refresh, collapsible sections, dismiss functionality)
- Need to increase timeouts or fix async handling
- Some tests may need to account for component state changes

### RiskRecommendationsSection (9 failures)
- Similar collapsible section issues
- Filter tests need fixing
- Action item display tests need collapsible expansion

---

## Next Steps

1. **Fix Timeout Issues**
   - Increase test timeouts where needed
   - Fix async/await handling in tests with fake timers
   - Ensure proper cleanup of timers

2. **Complete Collapsible Section Fixes**
   - Fix remaining tests that need section expansion
   - Add helper functions for expanding sections consistently

3. **Continue with Other Test Files**
   - Move to next failing test files
   - Apply similar patterns for collapsible components

---

**Status:** ✅ **Good Progress - 77.7% Pass Rate**  
**Phase 6 Tests:** ✅ **126/126 Passing (100%)**

