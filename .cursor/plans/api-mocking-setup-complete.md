# API Mocking Setup - Complete ✅

## What Was Blocking

1. ❌ **MSW Browser Worker** - Not generated
2. ❌ **Error Scenarios** - Handlers only had success cases
3. ❌ **Browser Integration** - MSW not initialized in browser
4. ❌ **Test Merchants** - SQL script not executed

## What's Now Fixed

### ✅ 1. MSW Browser Worker Generated
- **File:** `frontend/public/mockServiceWorker.js`
- **Status:** ✅ Generated via `npx msw init public/`
- **Usage:** Automatically loaded by MSW in browser

### ✅ 2. Error Scenario Handlers Created
- **File:** `frontend/__tests__/mocks/handlers-error-scenarios.ts`
- **Scenarios:**
  - ✅ 404 errors (merchant not found)
  - ✅ 500 errors (server errors)
  - ✅ Missing risk assessment
  - ✅ Missing portfolio statistics
  - ✅ Missing merchant analytics
  - ✅ Missing industry code
  - ✅ Network timeout simulation
  - ✅ Test merchant handlers (merchant-no-risk, merchant-complete-123, etc.)

### ✅ 3. Browser Integration
- **File:** `frontend/lib/msw-browser.ts`
- **Features:**
  - ✅ Auto-initializes in browser (development only)
  - ✅ Toggle via `NEXT_PUBLIC_MSW_ENABLED=true` or `localStorage.setItem('msw-enabled', 'true')`
  - ✅ Combines all handlers (success + error scenarios)
  - ✅ Exposes worker globally for debugging: `window.__MSW_WORKER__`

### ✅ 4. App Integration
- **File:** `frontend/app/layout.tsx`
- **Status:** ✅ MSW browser module imported (development only)

### ✅ 5. Test Merchant Seeder
- **File:** `scripts/seed-test-merchants.sh`
- **Status:** ✅ Created (requires manual execution via Supabase Dashboard)

## How to Use

### Enable MSW in Browser

**Option 1: Environment Variable**
```bash
# In .env.local or railway.env
NEXT_PUBLIC_MSW_ENABLED=true
```

**Option 2: Browser Console**
```javascript
localStorage.setItem('msw-enabled', 'true');
// Reload page
```

**Option 3: Disable**
```javascript
localStorage.setItem('msw-enabled', 'false');
// Or remove NEXT_PUBLIC_MSW_ENABLED
```

### Test Error Scenarios

Use these merchant IDs to trigger specific error scenarios:

- `merchant-404` - 404 Not Found
- `merchant-500` - 500 Server Error
- `merchant-no-risk` - No risk assessment
- `merchant-no-analytics` - No analytics data
- `merchant-no-industry-code` - No industry code (for benchmarks)
- `merchant-timeout` - Network timeout
- `merchant-complete-123` - Complete data (success scenario)

### Seed Test Merchants

**Option 1: Supabase Dashboard**
1. Open Supabase Dashboard
2. Go to SQL Editor
3. Copy contents of `test/sql/test_merchant_data.sql`
4. Execute

**Option 2: Supabase CLI**
```bash
supabase db execute --file test/sql/test_merchant_data.sql
```

**Option 3: Script (provides instructions)**
```bash
./scripts/seed-test-merchants.sh
```

## Postman MCP Alternative

While MSW is recommended for browser testing, Postman MCP can create external mock servers:

**Pros:**
- ✅ External HTTP mock server
- ✅ Can be shared across team
- ✅ No code changes needed

**Cons:**
- ❌ Requires external service
- ❌ Network latency
- ❌ CORS configuration needed
- ❌ Less flexible for dynamic scenarios

**To Create Postman Mock Server:**
1. Use Postman MCP to create a collection
2. Add requests for all API endpoints
3. Create mock server from collection
4. Update `NEXT_PUBLIC_API_BASE_URL` to point to mock server

## Next Steps

1. ✅ **Enable MSW** - Set `NEXT_PUBLIC_MSW_ENABLED=true`
2. ✅ **Test Error Scenarios** - Use test merchant IDs
3. ✅ **Seed Test Merchants** - Run SQL script in Supabase
4. ✅ **Complete Phase 2 Tests** - Use MSW to test all error states

## Files Created/Modified

- ✅ `frontend/public/mockServiceWorker.js` (generated)
- ✅ `frontend/__tests__/mocks/handlers-error-scenarios.ts` (new)
- ✅ `frontend/lib/msw-browser.ts` (new)
- ✅ `frontend/app/layout.tsx` (modified - added MSW import)
- ✅ `frontend/__tests__/mocks/handlers.ts` (modified - added portfolio endpoints)
- ✅ `scripts/seed-test-merchants.sh` (new)
- ✅ `.cursor/plans/api-mocking-solution.md` (new)
- ✅ `.cursor/plans/api-mocking-setup-complete.md` (this file)

## Status: ✅ Ready to Use

All blocking issues resolved! MSW is now ready for browser testing of Phase 2 error scenarios.

