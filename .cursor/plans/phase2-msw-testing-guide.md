# Phase 2 MSW Testing Guide

## ✅ Setup Complete

### 1. MSW Enabled
- **File:** `frontend/.env.local`
- **Setting:** `NEXT_PUBLIC_MSW_ENABLED=true`
- **Status:** ✅ Enabled

### 2. Test Merchants Seeded
- **Database:** Supabase
- **Merchants Created:**
  - ✅ `merchant-complete-123` - Complete data (success scenario)
  - ✅ `merchant-404` - Triggers 404 error
  - ✅ `merchant-500` - Triggers 500 error
  - ✅ `merchant-no-risk` - No risk assessment
  - ✅ `merchant-no-analytics` - No analytics data
  - ✅ `merchant-no-industry-code` - No industry code

### 3. Error Handlers Ready
- **File:** `frontend/__tests__/mocks/handlers-error-scenarios.ts`
- **Scenarios:** All error scenarios configured

## Testing Error Scenarios

### Test 1: 404 Error (Merchant Not Found)
**URL:** `http://localhost:3000/merchant-details/merchant-404`

**Expected:**
- Error message with code (e.g., "Error PC-005: Merchant not found")
- CTA button visible (Retry/Refresh)

**MSW Handler:** Returns 404 status

### Test 2: 500 Server Error
**URL:** `http://localhost:3000/merchant-details/merchant-500`

**Expected:**
- Error message with code (e.g., "Error RS-003: Internal server error")
- Retry button visible

**MSW Handler:** Returns 500 status

### Test 3: Missing Risk Assessment
**URL:** `http://localhost:3000/merchant-details/merchant-no-risk`

**Expected:**
- Error message: "Error RS-001: No risk assessment found"
- "Start Risk Assessment" button visible
- PortfolioComparisonCard shows "Run Risk Assessment" CTA

**MSW Handler:** Returns empty/null risk score

### Test 4: Missing Analytics
**URL:** `http://localhost:3000/merchant-details/merchant-no-analytics`

**Expected:**
- Error message: "Error AC-001: Merchant analytics not found"
- Retry button visible

**MSW Handler:** Returns 404 for analytics endpoint

### Test 5: Missing Industry Code
**URL:** `http://localhost:3000/merchant-details/merchant-no-industry-code`

**Expected:**
- RiskBenchmarkComparison shows: "Error RB-001: Industry code is required"
- "Enrich Data" button visible

**MSW Handler:** Returns analytics without industry codes

### Test 6: Complete Data (Success)
**URL:** `http://localhost:3000/merchant-details/merchant-complete-123`

**Expected:**
- All components load successfully
- No error messages
- All data displays correctly

## MSW Toggle

### Enable MSW
```bash
# In .env.local
NEXT_PUBLIC_MSW_ENABLED=true

# Or in browser console
localStorage.setItem('msw-enabled', 'true')
```

### Disable MSW
```bash
# In .env.local
NEXT_PUBLIC_MSW_ENABLED=false

# Or in browser console
localStorage.setItem('msw-enabled', 'false')
```

### Check MSW Status
```javascript
// In browser console
console.log('MSW Enabled:', localStorage.getItem('msw-enabled') === 'true');
console.log('MSW Worker:', window.__MSW_WORKER__);
```

## Phase 2 Test Checklist with MSW

### PortfolioComparisonCard Tests
- [ ] Test 1.1: Missing Risk Score → Use `merchant-no-risk`
- [ ] Test 1.2: Missing Portfolio Stats → MSW returns 404 for `/merchants/statistics`
- [ ] Test 1.3: Missing Both → Use `merchant-no-risk` + mock portfolio stats 404
- [ ] Test 1.4: Partial Data - Risk Score Only → MSW returns risk score but no portfolio stats
- [ ] Test 1.5: Partial Data - Portfolio Stats Only → MSW returns portfolio stats but no risk score
- [ ] Test 1.6: Loading State → Verify skeleton
- [ ] Test 1.7: Success State → Use `merchant-complete-123`

### RiskScoreCard Tests
- [ ] Test 2.1: No Risk Assessment → Use `merchant-no-risk`
- [ ] Test 2.2: API Failure → Use `merchant-500`
- [ ] Test 2.3: Invalid Data → MSW returns invalid risk score format
- [ ] Test 2.4: Loading State → Verify skeleton
- [ ] Test 2.5: Success State → Use `merchant-complete-123`

### AnalyticsComparison Tests
- [ ] Test 3.1: Missing Merchant Analytics → Use `merchant-no-analytics`
- [ ] Test 3.2: Missing Portfolio Analytics → MSW returns 404 for portfolio analytics
- [ ] Test 3.3: Missing Both → Use `merchant-no-analytics` + mock portfolio analytics 404
- [ ] Test 3.4: Loading State → Verify skeleton
- [ ] Test 3.5: Success State → Use `merchant-complete-123`

### RiskBenchmarkComparison Tests
- [ ] Test 4.1: Missing Industry Code → Use `merchant-no-industry-code`
- [ ] Test 4.2: Benchmarks Unavailable → MSW returns 404 for benchmarks
- [ ] Test 4.3: Missing Risk Score → Use `merchant-no-risk`
- [ ] Test 4.4: Loading State → Verify skeleton
- [ ] Test 4.5: Success State → Use `merchant-complete-123`

## Next Steps

1. ✅ **Restart Dev Server** - MSW changes require restart
   ```bash
   cd frontend
   npm run dev
   ```

2. ✅ **Verify MSW is Active** - Check browser console for `[MSW] ✅ Mock Service Worker started`

3. ✅ **Test Error Scenarios** - Navigate to test merchant URLs above

4. ✅ **Complete Phase 2 Tests** - Use MSW to test all error states systematically

## Troubleshooting

### MSW Not Starting
- Check `.env.local` has `NEXT_PUBLIC_MSW_ENABLED=true`
- Check browser console for errors
- Verify `public/mockServiceWorker.js` exists
- Restart dev server

### Handlers Not Matching
- Check browser console for `[MSW] Unhandled request` warnings
- Verify API base URL matches in handlers
- Check handler order (more specific handlers first)

### Error Scenarios Not Triggering
- Verify merchant ID matches handler conditions
- Check MSW worker is active: `window.__MSW_WORKER__`
- Review handler logic in `handlers-error-scenarios.ts`

