# v1 Dashboard Metrics Deprecation Analysis

**Date**: 2025-11-17  
**Status**: 📋 **RECOMMENDATION**

## Current State

### v3 Dashboard Metrics (`/api/v3/dashboard/metrics`)
- ✅ **Status**: Working (200 OK)
- ✅ **Data**: Comprehensive metrics from BI Service
- ✅ **Format**: Enhanced structure with `overview`, `performance`, `business`, `customer_kpis`, `financial_kpis`, etc.
- ✅ **Frontend**: Primary endpoint, tried first

### v1 Dashboard Metrics (`/api/v1/dashboard/metrics`)
- ❌ **Status**: Broken (500 error - "Failed to get dashboard")
- ❌ **Backend**: Risk Assessment Service endpoint `/api/v1/reporting/dashboards/metrics` appears to not exist or is broken
- ⚠️ **Frontend**: Only used as fallback if v3 returns 404
- ⚠️ **Usage**: No active consumers (frontend only uses if v3 fails)

## Analysis

### Frontend Usage Pattern
```typescript
// Try v3 endpoint first, fallback to v1
let response = await fetch(ApiEndpoints.dashboard.metrics('v3'), {
  method: 'GET',
  headers,
});

if (!response.ok && response.status === 404) {
  // Fallback to v1 endpoint if v3 doesn't exist
  response = await fetch(ApiEndpoints.dashboard.metrics('v1'), {
    method: 'GET',
    headers,
  });
}
```

**Key Points**:
- Frontend only falls back to v1 if v3 returns **404** (not found)
- Since v3 is working, the fallback is never triggered
- v1 is currently broken (500 error), so even if fallback triggered, it wouldn't work

### Backend Status
- v3 routes to: BI Service `/dashboard/kpis` ✅ Working
- v1 routes to: Risk Assessment Service `/api/v1/reporting/dashboards/metrics` ❌ Broken

## Recommendation

### Option 1: Remove v1 (Recommended)
**Rationale**:
- v3 is working and provides comprehensive data
- v1 is broken and not being used
- No external consumers identified
- Simplifies codebase
- Frontend already handles v3-only gracefully

**Actions**:
1. Remove `/api/v1/dashboard/metrics` route from API Gateway
2. Remove `ProxyToDashboardMetricsV1` handler
3. Update frontend to remove v1 fallback logic
4. Update tests to remove v1 references

### Option 2: Keep v1 for Backward Compatibility
**Rationale**:
- API versioning best practice (deprecation period)
- Potential future consumers
- Safety net during transition

**Actions**:
1. Fix v1 backend endpoint (Risk Assessment Service)
2. Keep route and handler
3. Document deprecation timeline
4. Add deprecation headers to v1 responses

## Decision

**Recommendation: Remove v1** ✅

**Reasons**:
1. v3 is production-ready and working
2. v1 is broken and unused
3. No external consumers
4. Frontend handles v3-only gracefully
5. Reduces maintenance burden
6. Cleaner API surface

**Migration Path**:
- Remove v1 route immediately (it's not working anyway)
- Frontend already uses v3 as primary
- No breaking changes for existing consumers (v3 works)

