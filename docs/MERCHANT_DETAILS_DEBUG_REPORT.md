# Merchant Details Page Debug Report

**Date:** 2025-01-20  
**Page URL:** https://frontend-service-production-b225.up.railway.app/merchant-details/merchant_1763614602674531538

## Console Errors

### React Error #418 (Hydration Mismatch)
- **Status:** ❌ Still occurring
- **Error:** "Minified React error #418; visit https://react.dev/errors/418?args[]=HTML&args[]="
- **Location:** `06525dfb60487280.js:1`
- **Cause:** Server-rendered HTML doesn't match client-rendered HTML
- **Impact:** Page still loads but with hydration warnings

## UI Issues - Missing Data

### Overview Tab
- **Issue:** Shows "Missing required data for comparison" with Retry button
- **Component:** `PortfolioComparisonCard`
- **Root Cause:** Either `merchantScore` (from risk-score API) or `portfolioAvg` (from statistics API) is null/undefined
- **API Calls:**
  - ✅ `/api/v1/merchants/merchant_1763614602674531538/risk-score` - 200 OK
  - ✅ `/api/v1/merchants/statistics` - 200 OK
- **Code Location:** `frontend/components/merchant/PortfolioComparisonCard.tsx:54-58`
  ```typescript
  if (merchantScore == null || portfolioAvg == null) {
    setError('Missing required data for comparison.');
    setLoading(false);
    return;
  }
  ```

### Business Analytics Tab
- **Status:** ✅ Working - Charts are rendering
- **API Calls:**
  - ✅ `/api/v1/merchants/merchant_1763614602674531538/analytics` - 200 OK
  - ✅ `/api/v1/merchants/analytics` (portfolio) - 200 OK
  - ✅ `/api/v1/merchants/merchant_1763614602674531538/website-analysis` - 200 OK

### Risk Assessment Tab
- **Issue 1:** "No industry code available for this merchant. Cannot perform benchmark comparison."
  - **Component:** `RiskBenchmarkComparison`
  - **Root Cause:** Merchant missing industry code (MCC/NAICS/SIC)
  
- **Issue 2:** "No risk assessment found. Please run a risk assessment first."
  - **Component:** `RiskAssessmentTab`
  - **Root Cause:** No risk assessment exists for this merchant
  
- **Issue 3:** "API Error 404" for risk recommendations
  - **Component:** `RiskRecommendationsSection`
  - **API Call:** `/api/v1/merchants/merchant_1763614602674531538/risk-recommendations` - 404
  - **Status:** ✅ Expected - endpoint not implemented (handled gracefully)

- **API Calls:**
  - ✅ `/api/v1/merchants/merchant_1763614602674531538/risk-score` - 200 OK
  - ✅ `/api/v1/risk/predictions/merchant_1763614602674531538` - 200 OK
  - ❌ `/api/v1/risk/history/merchant_1763614602674531538?limit=10` - 404 (expected)
  - ❌ `/api/v1/merchants/merchant_1763614602674531538/risk-recommendations` - 404 (expected)
  - ❌ WebSocket connection to `/api/v1/risk/ws` - Failed (expected)

## Network Requests Summary

### Successful API Calls (200 OK)
1. ✅ `/api/v1/merchants/merchant_1763614602674531538` - Merchant data
2. ✅ `/api/v1/merchants/statistics` - Portfolio statistics
3. ✅ `/api/v1/merchants/merchant_1763614602674531538/risk-score` - Risk score
4. ✅ `/api/v1/merchants/merchant_1763614602674531538/analytics` - Merchant analytics
5. ✅ `/api/v1/merchants/analytics` - Portfolio analytics
6. ✅ `/api/v1/merchants/merchant_1763614602674531538/website-analysis` - Website analysis
7. ✅ `/api/v1/risk/predictions/merchant_1763614602674531538` - Risk predictions

### Expected 404s (Optional Endpoints)
1. ❌ `/api/v1/risk/history/merchant_1763614602674531538?limit=10` - Not implemented
2. ❌ `/api/v1/merchants/merchant_1763614602674531538/risk-recommendations` - Not implemented
3. ❌ WebSocket `/api/v1/risk/ws` - Not implemented

## Root Causes

### 1. React Error #418 (Hydration Mismatch)
- **Status:** Still occurring despite fixes
- **Possible Causes:**
  - Date formatting still causing issues (need to verify all date formatting is client-side only)
  - Dynamic imports with `ssr: false` may still cause hydration issues
  - Browser extensions modifying HTML before React loads
  - Invalid HTML tag nesting

### 2. Missing Data in PortfolioComparisonCard
- **Issue:** `merchantScore` or `portfolioAvg` is null/undefined
- **Investigation Needed:**
  - Check API response structure for `/api/v1/merchants/merchant_1763614602674531538/risk-score`
  - Check API response structure for `/api/v1/merchants/statistics`
  - Verify that `risk_score` field exists in risk-score response
  - Verify that `averageRiskScore` field exists in statistics response

### 3. Missing Industry Code
- **Issue:** Merchant has no industry code (MCC/NAICS/SIC)
- **Impact:** Cannot perform benchmark comparison
- **Solution:** Merchant needs to be enriched with industry codes

### 4. Missing Risk Assessment
- **Issue:** No risk assessment exists for this merchant
- **Impact:** Cannot display risk assessment details
- **Solution:** User needs to run a risk assessment first

## Recommendations

### Immediate Fixes

1. **Investigate PortfolioComparisonCard Data Issue**
   - Add logging to see what values are being returned from APIs
   - Check if API responses match expected TypeScript interfaces
   - Add fallback handling for missing data

2. **Fix React Error #418**
   - Verify all date formatting is client-side only
   - Check for any server-side rendering of dynamic content
   - Consider using `suppressHydrationWarning` more broadly if needed

3. **Improve Error Messages**
   - Make error messages more specific (e.g., "Missing merchant risk score" vs "Missing required data")
   - Add helpful actions (e.g., "Run risk assessment" button)

### Long-term Improvements

1. **Data Validation**
   - Add validation for required fields before attempting comparisons
   - Show clear messages about what data is missing and how to fix it

2. **API Response Handling**
   - Add better error handling for partial API responses
   - Implement graceful degradation when optional data is missing

3. **User Experience**
   - Add loading states that are more informative
   - Provide clear CTAs when data is missing (e.g., "Run Risk Assessment" button)

