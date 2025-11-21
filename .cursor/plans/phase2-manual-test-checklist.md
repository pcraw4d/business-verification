# Phase 2 Manual Testing Checklist

## Test Environment Setup

- [x] Start development server: `cd frontend && npm run dev` ✅ **Server is running on http://localhost:3000**
- [ ] Open browser DevTools console (F12 or Cmd+Option+I)
- [ ] Navigate to a merchant details page: `http://localhost:3000/merchant-portfolio/[merchant-id]`

## 1. PortfolioComparisonCard Tests

### Test 1.1: Missing Risk Score

- [ ] Navigate to merchant with no risk assessment
- [ ] Verify error message shows: "Error PC-003: A risk assessment must be completed..."
- [ ] Verify "Run Risk Assessment" button is visible
- [ ] Click button and verify it navigates to Risk Assessment tab
- [ ] Check console for development logs: `[PortfolioComparison]`

### Test 1.2: Missing Portfolio Stats

- [ ] Mock API to return 404 for `/api/v1/merchants/statistics`
- [ ] Verify error message shows: "Error PC-002: Portfolio statistics are being calculated..."
- [ ] Verify "Refresh Data" button is visible
- [ ] Click button and verify it retries the fetch

### Test 1.3: Missing Both

- [ ] Mock both APIs to fail
- [ ] Verify error message shows: "Error PC-003: A risk assessment must be completed..."
- [ ] Verify appropriate CTA buttons are shown

### Test 1.4: Partial Data - Risk Score Only

- [ ] Mock portfolio stats to fail, risk score to succeed
- [ ] Verify component shows merchant score with note about portfolio stats
- [ ] Verify no errors in console

### Test 1.5: Partial Data - Portfolio Stats Only

- [ ] Mock risk score to fail, portfolio stats to succeed
- [ ] Verify component shows portfolio average with "Run Risk Assessment" CTA
- [ ] Verify no errors in console

### Test 1.6: Loading State

- [ ] Verify loading message: "Loading portfolio comparison..."
- [ ] Verify skeleton is displayed during load

### Test 1.7: Success State

- [ ] Verify full comparison displays when both data available
- [ ] Verify error codes NOT shown in success state
- [ ] Verify all comparison metrics display correctly

## 2. RiskScoreCard Tests

### Test 2.1: No Risk Assessment

- [ ] Navigate to merchant with no risk score
- [ ] Verify error message shows: "Error RS-001: No risk assessment has been completed..."
- [ ] Verify "Start Risk Assessment" button is visible
- [ ] Click button and verify it navigates to Risk Assessment tab

### Test 2.2: API Failure

- [ ] Mock API to return 500 error
- [ ] Verify error message shows: "Error RS-003: [error message]"
- [ ] Verify "Retry" button is visible
- [ ] Click retry and verify it refetches

### Test 2.3: Invalid Data

- [ ] Mock API to return invalid risk score (e.g., null, string)
- [ ] Verify component handles gracefully
- [ ] Verify appropriate error message with code

### Test 2.4: Loading State

- [ ] Verify loading message: "Loading risk assessment..."
- [ ] Verify skeleton is displayed

### Test 2.5: Success State

- [ ] Verify risk score displays correctly
- [ ] Verify risk level badge displays
- [ ] Verify confidence score displays
- [ ] Verify assessment date displays (if available)

## 3. AnalyticsComparison Tests

### Test 3.1: Missing Merchant Analytics

- [ ] Mock merchant analytics API to fail
- [ ] Verify error message shows: "Error AC-001: Unable to fetch merchant analytics..."
- [ ] Verify "Retry" button is visible

### Test 3.2: Missing Portfolio Analytics

- [ ] Mock portfolio analytics API to fail
- [ ] Verify error message shows: "Error AC-002: Unable to fetch portfolio analytics..."
- [ ] Verify "Retry" button is visible

### Test 3.3: Missing Both

- [ ] Mock both APIs to fail
- [ ] Verify error message shows: "Error AC-003: Unable to fetch merchant analytics and portfolio analytics..."
- [ ] Verify appropriate error handling

### Test 3.4: Loading State

- [ ] Verify loading message: "Loading portfolio comparison..."
- [ ] Verify skeleton is displayed

### Test 3.5: Success State

- [ ] Verify comparison charts display
- [ ] Verify all metrics compare correctly
- [ ] Verify no error codes in success state

## 4. RiskBenchmarkComparison Tests

### Test 4.1: Missing Industry Code

- [ ] Navigate to merchant with no industry code
- [ ] Verify error message shows: "Error RB-001: Industry code is required for benchmark comparison..."
- [ ] Verify "Enrich Data" button is visible
- [ ] Click button and verify enrichment dialog opens

### Test 4.2: Benchmarks Unavailable

- [ ] Mock benchmarks API to fail
- [ ] Verify error message shows: "Error RB-002: Benchmark data for this industry is currently unavailable..."
- [ ] Verify "Retry" button is visible

### Test 4.3: Missing Risk Score

- [ ] Mock risk score API to fail
- [ ] Verify error message shows: "Error RB-003: Unable to fetch merchant risk score..."
- [ ] Verify appropriate error handling

### Test 4.4: Loading State

- [ ] Verify loading message: "Fetching industry benchmarks..."
- [ ] Verify skeleton is displayed

### Test 4.5: Success State

- [ ] Verify benchmark comparison displays
- [ ] Verify percentile position shows
- [ ] Verify no error codes in success state

## 5. Error Code Verification

### Test 5.1: Error Code Format

- [ ] Verify all error messages start with "Error CODE-XXX:"
- [ ] Verify codes follow pattern: PC-001, RS-001, AC-001, RB-001
- [ ] Verify codes are consistent across similar errors

### Test 5.2: Error Code Coverage

- [ ] Verify PC-001 through PC-005 are used
- [ ] Verify RS-001 through RS-003 are used
- [ ] Verify AC-001 through AC-005 are used
- [ ] Verify RB-001 through RB-005 are used

## 6. Development Logging

### Test 6.1: Console Logs

- [ ] Open DevTools console
- [ ] Navigate through all components
- [ ] Verify development logs appear:
  - `[PortfolioComparison] API Results:`
  - `[PortfolioComparison] Fields available:`
  - `[RiskScoreCard] API Response:`
  - `[AnalyticsComparison] Merchant analytics loaded:`
  - `[RiskBenchmarkComparison] Merchant analytics loaded:`

### Test 6.2: Production Mode

- [ ] Build for production: `npm run build`
- [ ] Start production server
- [ ] Verify NO console logs appear
- [ ] Verify error codes still display

## 7. Type Guard and Validation

### Test 7.1: Invalid Risk Score

- [ ] Mock API to return invalid risk_score (e.g., "invalid", null, -1, 2)
- [ ] Verify component handles gracefully
- [ ] Verify appropriate error message
- [ ] Verify no runtime errors

### Test 7.2: Invalid Portfolio Stats

- [ ] Mock API to return invalid portfolio stats structure
- [ ] Verify component handles gracefully
- [ ] Verify appropriate error message
- [ ] Verify no runtime errors

## 8. User Experience

### Test 8.1: Error Recovery

- [ ] Trigger an error state
- [ ] Click retry/refresh button
- [ ] Verify data reloads successfully
- [ ] Verify error state clears

### Test 8.2: Navigation

- [ ] Click "Run Risk Assessment" button
- [ ] Verify navigation to Risk Assessment tab
- [ ] Click "Enrich Data" button
- [ ] Verify enrichment dialog opens

### Test 8.3: Loading Transitions

- [ ] Verify smooth transition from loading to content
- [ ] Verify smooth transition from loading to error
- [ ] Verify no flickering or layout shifts

## Success Criteria

- ✅ All error messages include error codes
- ✅ All error states have actionable CTAs
- ✅ Loading states are descriptive
- ✅ Partial data scenarios handled gracefully
- ✅ No console errors
- ✅ All buttons functional
- ✅ Development logs appear in dev mode only
- ✅ Type guards prevent runtime errors
