# API Mocking & Test Merchants Solution

## Current State

### ✅ What's Already Set Up
1. **MSW (Mock Service Worker) v2** - Installed and configured for unit tests
2. **MSW Handlers** - Basic success scenarios exist in `frontend/__tests__/mocks/handlers.ts`
3. **Test Merchant SQL** - Script exists in `test/sql/test_merchant_data.sql`
4. **Test Data Seeders** - Go seeder exists in `test/integration/test_data_seeder.go`

### ❌ What's Blocking
1. **MSW Browser Setup** - No `public/mockServiceWorker.js` worker file
2. **Error Scenarios** - Handlers only have success cases (no 404, 500, missing data)
3. **Browser Integration** - MSW not initialized in browser (only in test environment)
4. **Test Merchants** - SQL script not executed against database
5. **Toggle Mechanism** - No way to enable/disable MSW in browser

## Solution Overview

### Option 1: MSW Browser Setup (Recommended)
**Pros:**
- ✅ Intercepts at fetch level (works with all API calls)
- ✅ No external dependencies
- ✅ Can toggle on/off easily
- ✅ Already installed and configured

**Cons:**
- ⚠️ Requires browser worker setup
- ⚠️ Need to extend handlers for error scenarios

### Option 2: Postman Mock Server (Alternative)
**Pros:**
- ✅ External HTTP mock server (no code changes)
- ✅ Can be shared across team
- ✅ Postman MCP can create/manage it

**Cons:**
- ❌ Requires external service
- ❌ Network latency
- ❌ CORS configuration needed
- ❌ Less flexible for dynamic scenarios

## Recommended Approach: MSW Browser Setup

### Step 1: Generate MSW Browser Worker
```bash
cd frontend
npx msw init public/ --save
```

### Step 2: Extend Handlers with Error Scenarios
Add handlers for:
- 404 errors (merchant not found)
- 500 errors (server errors)
- Missing data scenarios (no risk assessment, no analytics, etc.)
- Network errors (timeout, CORS)

### Step 3: Browser Integration
Create `frontend/lib/msw-browser.ts` to initialize MSW in browser when enabled.

### Step 4: Toggle Mechanism
Use environment variable or localStorage to enable/disable MSW:
- `NEXT_PUBLIC_MSW_ENABLED=true` (development only)
- Or localStorage: `localStorage.setItem('msw-enabled', 'true')`

### Step 5: Seed Test Merchants
Create script to run SQL against Supabase database.

## Implementation Plan

1. ✅ Generate MSW browser worker
2. ✅ Extend handlers with error scenarios
3. ✅ Create browser integration
4. ✅ Add toggle mechanism
5. ✅ Create test merchant seeder script
6. ✅ Optional: Create Postman collection for external mocking

