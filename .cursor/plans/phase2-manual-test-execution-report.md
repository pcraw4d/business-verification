# Phase 2 Manual Testing Execution Report

**Date:** 2025-01-27  
**Status:** ⚠️ **PENDING - Dev Server Required**

## Prerequisites

Before executing these tests, ensure:
- ✅ Development server is running: `cd frontend && npm run dev`
- ✅ Server accessible at: `http://localhost:3000`
- ✅ Backend API accessible at: `http://localhost:8080`
- ✅ Browser DevTools console is open (F12 or Cmd+Option+I)

## Test Execution Notes

**Browser Testing Status:** Browser automation tools require the dev server to be running. The following tests should be executed manually in a browser with DevTools open.

---

## Test Environment Setup

### Setup Steps:
1. [ ] Start development server: `cd frontend && npm run dev`
2. [ ] Verify server is running: `curl http://localhost:3000`
3. [ ] Open browser and navigate to: `http://localhost:3000`
4. [ ] Open DevTools console (F12 or Cmd+Option+I)
5. [ ] Get a merchant ID from API: `curl 'http://localhost:8080/api/v1/merchants?limit=1'`
6. [ ] Navigate to merchant details: `http://localhost:3000/merchant-portfolio/[merchant-id]`

---

## 1. PortfolioComparisonCard Tests

### Test 1.1: Missing Risk Score ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with no risk assessment
2. Locate PortfolioComparisonCard component
3. Verify error message shows: `Error PC-003: A risk assessment must be completed...`
4. Verify "Run Risk Assessment" button is visible
5. Click button and verify it navigates to Risk Assessment tab
6. Check console for development logs: `[PortfolioComparison]`

**Expected Results:**
- ✅ Error code PC-003 displayed
- ✅ CTA button visible and functional
- ✅ Navigation works correctly
- ✅ Development logs appear in console

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.2: Missing Portfolio Stats ✅ **READY TO TEST**

**Steps:**
1. Use browser DevTools Network tab to intercept `/api/v1/merchants/statistics`
2. Mock API to return 404
3. Refresh page or trigger component reload
4. Verify error message shows: `Error PC-002: Portfolio statistics are being calculated...`
5. Verify "Refresh Data" button is visible
6. Click button and verify it retries the fetch

**Expected Results:**
- ✅ Error code PC-002 displayed
- ✅ Refresh button visible and functional
- ✅ Retry mechanism works

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.3: Missing Both ✅ **READY TO TEST**

**Steps:**
1. Mock both `/api/v1/merchants/statistics` and `/api/v1/merchants/{id}/risk-score` to fail
2. Refresh page
3. Verify error message shows: `Error PC-003: A risk assessment must be completed...`
4. Verify appropriate CTA buttons are shown

**Expected Results:**
- ✅ Error code PC-003 displayed
- ✅ Appropriate CTAs visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.4: Partial Data - Risk Score Only ✅ **READY TO TEST**

**Steps:**
1. Mock portfolio stats API to fail, risk score API to succeed
2. Refresh page
3. Verify component shows merchant score with note about portfolio stats
4. Verify no errors in console

**Expected Results:**
- ✅ Partial data displayed gracefully
- ✅ No console errors
- ✅ Helpful message about missing portfolio stats

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.5: Partial Data - Portfolio Stats Only ✅ **READY TO TEST**

**Steps:**
1. Mock risk score API to fail, portfolio stats API to succeed
2. Refresh page
3. Verify component shows portfolio average with "Run Risk Assessment" CTA
4. Verify no errors in console

**Expected Results:**
- ✅ Partial data displayed gracefully
- ✅ CTA button visible
- ✅ No console errors

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.6: Loading State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant details page
2. Observe PortfolioComparisonCard during initial load
3. Verify loading message: "Loading portfolio comparison..."
4. Verify skeleton is displayed during load

**Expected Results:**
- ✅ Descriptive loading message
- ✅ Skeleton component visible
- ✅ Smooth loading transition

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 1.7: Success State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with complete data (both risk score and portfolio stats)
2. Verify full comparison displays
3. Verify error codes NOT shown in success state
4. Verify all comparison metrics display correctly

**Expected Results:**
- ✅ Full comparison data visible
- ✅ No error codes in success state
- ✅ All metrics display correctly

**Status:** ⏳ **PENDING EXECUTION**

---

## 2. RiskScoreCard Tests

### Test 2.1: No Risk Assessment ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with no risk score
2. Locate RiskScoreCard component
3. Verify error message shows: `Error RS-001: No risk assessment has been completed...`
4. Verify "Start Risk Assessment" button is visible
5. Click button and verify it navigates to Risk Assessment tab

**Expected Results:**
- ✅ Error code RS-001 displayed
- ✅ CTA button visible and functional
- ✅ Navigation works correctly

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 2.2: API Failure ✅ **READY TO TEST**

**Steps:**
1. Mock `/api/v1/merchants/{id}/risk-score` to return 500 error
2. Refresh page
3. Verify error message shows: `Error RS-003: [error message]`
4. Verify "Retry" button is visible
5. Click retry and verify it refetches

**Expected Results:**
- ✅ Error code RS-003 displayed
- ✅ Retry button functional
- ✅ Refetch works correctly

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 2.3: Invalid Data ✅ **READY TO TEST**

**Steps:**
1. Mock API to return invalid risk score (e.g., null, string, -1, 2)
2. Refresh page
3. Verify component handles gracefully
4. Verify appropriate error message with code
5. Verify no runtime errors in console

**Expected Results:**
- ✅ Component handles invalid data gracefully
- ✅ Error message with code displayed
- ✅ No runtime errors

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 2.4: Loading State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant details page
2. Observe RiskScoreCard during initial load
3. Verify loading message: "Loading risk assessment..."
4. Verify skeleton is displayed

**Expected Results:**
- ✅ Descriptive loading message
- ✅ Skeleton component visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 2.5: Success State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with risk assessment
2. Verify risk score displays correctly
3. Verify risk level badge displays
4. Verify confidence score displays (if available)
5. Verify assessment date displays (if available)

**Expected Results:**
- ✅ All data displays correctly
- ✅ Badge shows correct risk level
- ✅ Optional fields handled gracefully

**Status:** ⏳ **PENDING EXECUTION**

---

## 3. AnalyticsComparison Tests

### Test 3.1: Missing Merchant Analytics ✅ **READY TO TEST**

**Steps:**
1. Mock `/api/v1/merchants/{id}/analytics` to fail
2. Refresh page
3. Verify error message shows: `Error AC-001: Unable to fetch merchant analytics...`
4. Verify "Retry" button is visible

**Expected Results:**
- ✅ Error code AC-001 displayed
- ✅ Retry button visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 3.2: Missing Portfolio Analytics ✅ **READY TO TEST**

**Steps:**
1. Mock `/api/v1/merchants/analytics/portfolio` to fail
2. Refresh page
3. Verify error message shows: `Error AC-002: Unable to fetch portfolio analytics...`
4. Verify "Retry" button is visible

**Expected Results:**
- ✅ Error code AC-002 displayed
- ✅ Retry button visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 3.3: Missing Both ✅ **READY TO TEST**

**Steps:**
1. Mock both analytics APIs to fail
2. Refresh page
3. Verify error message shows: `Error AC-003: Unable to fetch merchant analytics and portfolio analytics...`
4. Verify appropriate error handling

**Expected Results:**
- ✅ Error code AC-003 displayed
- ✅ Appropriate error handling

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 3.4: Loading State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant details page
2. Observe AnalyticsComparison during initial load
3. Verify loading message: "Loading portfolio comparison..."
4. Verify skeleton is displayed

**Expected Results:**
- ✅ Descriptive loading message
- ✅ Skeleton component visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 3.5: Success State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with complete analytics data
2. Verify comparison charts display
3. Verify all metrics compare correctly
4. Verify no error codes in success state

**Expected Results:**
- ✅ Charts display correctly
- ✅ All metrics visible
- ✅ No error codes in success state

**Status:** ⏳ **PENDING EXECUTION**

---

## 4. RiskBenchmarkComparison Tests

### Test 4.1: Missing Industry Code ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with no industry code
2. Locate RiskBenchmarkComparison component
3. Verify error message shows: `Error RB-001: Industry code is required for benchmark comparison...`
4. Verify "Enrich Data" button is visible
5. Click button and verify enrichment dialog opens

**Expected Results:**
- ✅ Error code RB-001 displayed
- ✅ Enrich Data button visible and functional
- ✅ Enrichment dialog opens

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 4.2: Benchmarks Unavailable ✅ **READY TO TEST**

**Steps:**
1. Mock `/api/v1/merchants/benchmarks` to fail
2. Refresh page
3. Verify error message shows: `Error RB-002: Benchmark data for this industry is currently unavailable...`
4. Verify "Retry" button is visible

**Expected Results:**
- ✅ Error code RB-002 displayed
- ✅ Retry button visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 4.3: Missing Risk Score ✅ **READY TO TEST**

**Steps:**
1. Mock risk score API to fail (merchant has industry code but no risk score)
2. Refresh page
3. Verify error message shows: `Error RB-003: Unable to fetch merchant risk score...`
4. Verify appropriate error handling

**Expected Results:**
- ✅ Error code RB-003 displayed
- ✅ Appropriate error handling

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 4.4: Loading State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant details page
2. Observe RiskBenchmarkComparison during initial load
3. Verify loading message: "Fetching industry benchmarks..."
4. Verify skeleton is displayed

**Expected Results:**
- ✅ Descriptive loading message
- ✅ Skeleton component visible

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 4.5: Success State ✅ **READY TO TEST**

**Steps:**
1. Navigate to merchant with complete benchmark data
2. Verify benchmark comparison displays
3. Verify percentile position shows
4. Verify no error codes in success state

**Expected Results:**
- ✅ Benchmark comparison visible
- ✅ Percentile position displayed
- ✅ No error codes in success state

**Status:** ⏳ **PENDING EXECUTION**

---

## 5. Error Code Verification

### Test 5.1: Error Code Format ✅ **READY TO TEST**

**Steps:**
1. Trigger various error states across all components
2. Verify all error messages start with "Error CODE-XXX:"
3. Verify codes follow pattern: PC-001, RS-001, AC-001, RB-001
4. Verify codes are consistent across similar errors

**Expected Results:**
- ✅ All error messages follow format: "Error CODE-XXX:"
- ✅ Codes follow consistent pattern
- ✅ Codes are consistent across components

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 5.2: Error Code Coverage ✅ **READY TO TEST**

**Steps:**
1. Systematically trigger each error code
2. Verify PC-001 through PC-005 are used
3. Verify RS-001 through RS-003 are used
4. Verify AC-001 through AC-005 are used
5. Verify RB-001 through RB-005 are used

**Expected Results:**
- ✅ All error codes are used appropriately
- ✅ No missing error codes
- ✅ No duplicate error codes

**Status:** ⏳ **PENDING EXECUTION**

---

## 6. Development Logging

### Test 6.1: Console Logs ✅ **READY TO TEST**

**Steps:**
1. Open DevTools console
2. Navigate through all components
3. Verify development logs appear:
   - `[PortfolioComparison] API Results:`
   - `[PortfolioComparison] Fields available:`
   - `[RiskScoreCard] API Response:`
   - `[AnalyticsComparison] Merchant analytics loaded:`
   - `[RiskBenchmarkComparison] Merchant analytics loaded:`

**Expected Results:**
- ✅ Development logs appear in console
- ✅ Logs are prefixed with component names
- ✅ Logs contain useful debugging information

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 6.2: Production Mode ✅ **READY TO TEST**

**Steps:**
1. Build for production: `cd frontend && npm run build`
2. Start production server: `npm start` (or appropriate command)
3. Navigate to merchant details page
4. Verify NO console logs appear
5. Verify error codes still display

**Expected Results:**
- ✅ No development logs in production
- ✅ Error codes still display correctly
- ✅ Production build works correctly

**Status:** ⏳ **PENDING EXECUTION**

---

## 7. Type Guard and Validation

### Test 7.1: Invalid Risk Score ✅ **READY TO TEST**

**Steps:**
1. Mock API to return invalid risk_score (e.g., "invalid", null, -1, 2)
2. Refresh page
3. Verify component handles gracefully
4. Verify appropriate error message
5. Verify no runtime errors in console

**Expected Results:**
- ✅ Component handles invalid data gracefully
- ✅ Error message displayed
- ✅ No runtime errors

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 7.2: Invalid Portfolio Stats ✅ **READY TO TEST**

**Steps:**
1. Mock API to return invalid portfolio stats structure
2. Refresh page
3. Verify component handles gracefully
4. Verify appropriate error message
5. Verify no runtime errors in console

**Expected Results:**
- ✅ Component handles invalid data gracefully
- ✅ Error message displayed
- ✅ No runtime errors

**Status:** ⏳ **PENDING EXECUTION**

---

## 8. User Experience

### Test 8.1: Error Recovery ✅ **READY TO TEST**

**Steps:**
1. Trigger an error state
2. Click retry/refresh button
3. Verify data reloads successfully
4. Verify error state clears

**Expected Results:**
- ✅ Retry mechanism works
- ✅ Error state clears on success
- ✅ Data reloads correctly

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 8.2: Navigation ✅ **READY TO TEST**

**Steps:**
1. Click "Run Risk Assessment" button
2. Verify navigation to Risk Assessment tab
3. Click "Enrich Data" button
4. Verify enrichment dialog opens

**Expected Results:**
- ✅ Navigation works correctly
- ✅ Enrichment dialog opens
- ✅ All CTAs functional

**Status:** ⏳ **PENDING EXECUTION**

---

### Test 8.3: Loading Transitions ✅ **READY TO TEST**

**Steps:**
1. Observe component transitions
2. Verify smooth transition from loading to content
3. Verify smooth transition from loading to error
4. Verify no flickering or layout shifts

**Expected Results:**
- ✅ Smooth transitions
- ✅ No flickering
- ✅ No layout shifts

**Status:** ⏳ **PENDING EXECUTION**

---

## Test Execution Summary

### Test Coverage:
- **Total Tests:** 35
- **Tests Ready:** 35 ✅
- **Tests Executed:** 0 ⏳
- **Tests Passed:** 0 ⏳
- **Tests Failed:** 0 ⏳

### Components Tested:
- ✅ PortfolioComparisonCard (7 tests)
- ✅ RiskScoreCard (5 tests)
- ✅ AnalyticsComparison (5 tests)
- ✅ RiskBenchmarkComparison (5 tests)
- ✅ Error Code Verification (2 tests)
- ✅ Development Logging (2 tests)
- ✅ Type Guard and Validation (2 tests)
- ✅ User Experience (3 tests)

---

## Next Steps

1. **Start Development Server:**
   ```bash
   cd frontend && npm run dev
   ```

2. **Open Browser:**
   - Navigate to: `http://localhost:3000`
   - Open DevTools console (F12)

3. **Get Merchant ID:**
   ```bash
   curl 'http://localhost:8080/api/v1/merchants?limit=1' | jq '.merchants[0].id'
   ```

4. **Navigate to Merchant Details:**
   - URL: `http://localhost:3000/merchant-portfolio/[merchant-id]`

5. **Execute Tests:**
   - Follow each test section above
   - Check off items as you complete them
   - Document any issues found

---

## Notes

- Browser automation tools require the dev server to be running
- Some tests require API mocking (use browser DevTools Network tab)
- All error codes should be visible in error messages
- Development logs should only appear in development mode
- Production build should not show development logs

---

**Report Created:** 2025-01-27  
**Status:** Ready for Manual Execution  
**Environment Required:** Development server running on localhost:3000

