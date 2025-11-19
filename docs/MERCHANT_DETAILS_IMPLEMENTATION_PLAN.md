# Merchant Details & Portfolio Comparison Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to implement portfolio comparison features in the merchant details page and verify/improve portfolio-level dashboard implementations. The plan addresses architecture clarity, API Gateway routing, handler verification, and frontend integration with robust testing and risk mitigation strategies.

**Related Documents:**
- [Merchant Details Backend vs Frontend Comparison](./MERCHANT_DETAILS_BACKEND_FRONTEND_COMPARISON.md)
- [Merchant vs Portfolio Level Features](./MERCHANT_VS_PORTFOLIO_LEVEL_FEATURES.md)

**Last Updated:** 2025-01-27  
**Status:** Planning Phase  
**Estimated Timeline:** 4-6 weeks

---

## Goals and Objectives

### Primary Goals

1. **Verify Dashboard Pages Integration**
   - Ensure Business Intelligence, Risk, and Risk Indicators dashboards correctly use portfolio-level endpoints
   - Validate API Gateway routing for all portfolio endpoints
   - Confirm data accuracy and performance

2. **Implement Portfolio Comparison Features**
   - Add merchant-to-portfolio comparison capabilities to merchant details page
   - Display merchant data in context of portfolio averages and industry benchmarks
   - Provide percentile rankings and position indicators

3. **Enhance Merchant Details Page**
   - Integrate missing high-priority features (risk alerts, benchmarks, explainability, recommendations)
   - Add enrichment UI components
   - Improve data visualization with comparison charts

4. **Ensure System Reliability**
   - Verify API Gateway routing correctness
   - Validate handler implementations
   - Implement comprehensive error handling
   - Add monitoring and observability

### Success Metrics

#### Technical Metrics
- **API Gateway Routing Accuracy:** 100% of endpoints correctly routed
- **Handler Verification:** All handlers tested and verified
- **Test Coverage:** Minimum 80% coverage for new code
- **Performance:** API response times < 500ms (p95)
- **Error Rate:** < 0.1% for all endpoints

#### Feature Completion Metrics
- **Dashboard Verification:** 100% of dashboard pages verified and documented
- **Portfolio Comparison Features:** 5/5 comparison scenarios implemented
- **High-Priority Features:** 6/6 high-priority features implemented
- **UI Components:** All comparison components implemented and tested

#### User Experience Metrics
- **Page Load Time:** < 2 seconds for merchant details page
- **Comparison Data Availability:** 100% of comparison data loads successfully
- **Visual Clarity:** All comparisons clearly displayed with appropriate indicators

---

## Risk Mitigation Strategies

### Risk 1: Incorrect API Gateway Routing

**Risk Description:** Endpoints may be routed to wrong services or handlers, causing 404/502 errors.

**Mitigation:**
- **Pre-Implementation:** Comprehensive audit of all API Gateway routes
- **During Implementation:** Route-by-route verification with test requests
- **Post-Implementation:** Automated routing tests in CI/CD pipeline
- **Monitoring:** Alert on 404/502 errors for new endpoints

**Verification Tasks:**
- [ ] Audit all route registrations in `services/api-gateway/cmd/main.go`
- [ ] Verify route order (specific routes before PathPrefix)
- [ ] Test each endpoint with curl/Postman
- [ ] Document route mappings in API Gateway documentation

### Risk 2: Handler Implementation Issues

**Risk Description:** Handlers may have incorrect path transformations, validation, or error handling.

**Mitigation:**
- **Code Review:** Thorough review of handler implementations
- **Unit Testing:** Comprehensive unit tests for all handlers
- **Integration Testing:** End-to-end tests for critical paths
- **Error Handling:** Consistent error handling patterns

**Verification Tasks:**
- [ ] Review all handler functions in `services/api-gateway/internal/handlers/gateway.go`
- [ ] Verify path transformations are correct
- [ ] Test error handling scenarios
- [ ] Validate input sanitization and SQL injection prevention

### Risk 3: Data Inconsistency Between Services

**Risk Description:** Portfolio data from different services may be inconsistent or out of sync.

**Mitigation:**
- **Data Validation:** Validate data structure and types
- **Caching Strategy:** Implement appropriate caching with TTL
- **Error Handling:** Graceful degradation when portfolio data unavailable
- **Monitoring:** Alert on data inconsistencies

**Verification Tasks:**
- [ ] Verify data schemas match between services
- [ ] Test with empty/null data scenarios
- [ ] Validate date/time formats are consistent
- [ ] Test edge cases (single merchant, empty portfolio)

### Risk 4: Performance Degradation

**Risk Description:** Adding portfolio comparison queries may slow down merchant details page.

**Mitigation:**
- **Parallel Fetching:** Fetch merchant and portfolio data in parallel
- **Caching:** Cache portfolio statistics (5-10 minute TTL)
- **Lazy Loading:** Load comparison data only when tab is viewed
- **Request Deduplication:** Prevent duplicate API calls
- **Performance Monitoring:** Track response times and optimize slow queries

**Verification Tasks:**
- [ ] Implement request deduplication
- [ ] Add caching for portfolio statistics
- [ ] Monitor API response times
- [ ] Load test with concurrent users

### Risk 5: Frontend State Management Issues

**Risk Description:** Complex state management for merchant + portfolio data may cause bugs.

**Current State (Based on Codebase Review):**
- **State Management Pattern:** Codebase uses `useState` + `useEffect` pattern (NOT React Query)
- **Existing Patterns:** Request deduplication and caching already implemented via `RequestDeduplicator` and `APICache`
- **Components:** Dashboard pages use `useState` for loading/error states

**Mitigation:**
- **State Management:** Follow existing patterns (useState + useEffect) - **DO NOT introduce React Query** (not used in codebase)
- **Request Optimization:** Leverage existing `RequestDeduplicator` and `APICache` utilities
- **Error Boundaries:** Implement error boundaries for graceful failures
- **Loading States:** Clear loading indicators for all async operations (already implemented in dashboard pages)
- **Type Safety:** Comprehensive TypeScript types for all data structures

**Verification Tasks:**
- [ ] Review existing state management patterns in dashboard pages
- [ ] Ensure new components follow same patterns (useState + useEffect)
- [ ] Leverage existing `RequestDeduplicator` and `APICache` utilities
- [ ] Test error scenarios (network failures, invalid data)
- [ ] Verify loading states are displayed correctly
- [ ] Test concurrent tab switching

### Risk 6: Type Mismatches and Schema Changes

**Risk Description:** Backend schema changes may break frontend types.

**Mitigation:**
- **Type Generation:** Generate TypeScript types from OpenAPI spec
- **Type Validation:** Runtime validation with Zod or similar
- **Versioning:** API versioning for breaking changes
- **Documentation:** Keep API documentation up to date

**Verification Tasks:**
- [ ] Generate types from backend schemas
- [ ] Add runtime validation
- [ ] Document all type changes
- [ ] Test with different data shapes

---

## Phase 0: Critical Blockers Resolution (Before Phase 1)

### Goal
Resolve critical blockers that prevent implementation from proceeding.

### Tasks

#### Task 0.1: Backend Analytics Endpoints Implementation
**Objective:** Implement backend handlers for analytics endpoints that currently return "not yet implemented".

**Priority:** **CRITICAL** - Blocks Phase 2 and Phase 3

**Subtasks:**
1. **Implement Risk Trends Handler**
   - [ ] Review `services/risk-assessment-service/internal/handlers/risk_assessment.go` (line 813)
   - [ ] Implement `HandleRiskTrends` function
   - [ ] Query database for portfolio risk trends
   - [ ] Return trend data (6-month history, predictions, confidence bands)
   - [ ] Add error handling
   - [ ] Test endpoint directly

2. **Implement Risk Insights Handler**
   - [ ] Review `services/risk-assessment-service/internal/handlers/risk_assessment.go` (line 819)
   - [ ] Implement `HandleRiskInsights` function
   - [ ] Query database for portfolio risk insights
   - [ ] Return insights data (key findings, recommendations, patterns)
   - [ ] Add error handling
   - [ ] Test endpoint directly

**Deliverables:**
- Implemented backend handlers
- Test results
- Response schema documentation

**Testing Tollgate:**
- Endpoints return actual data (not "not yet implemented")
- Response schemas match expected structure
- Error handling works correctly

---

#### Task 0.2: API Gateway Route Registration
**Objective:** Add missing route registrations for analytics endpoints.

**Priority:** **CRITICAL** - Required for endpoints to be accessible

**Subtasks:**
1. **Add Analytics Routes**
   - [ ] Add `/analytics/trends` route registration in `services/api-gateway/cmd/main.go`
   - [ ] Add `/analytics/insights` route registration in `services/api-gateway/cmd/main.go`
   - [ ] Determine correct handler: `ProxyToRiskAssessment` (analytics routes are in risk-assessment service)
   - [ ] Register before PathPrefix to ensure correct routing
   - [ ] Add explicit registration for `/merchants/statistics` (currently relies on PathPrefix)

2. **Verify Route Order**
   - [ ] Ensure analytics routes registered before `/risk` PathPrefix
   - [ ] Ensure merchant routes registered in correct order
   - [ ] Test route matching

**Deliverables:**
- Updated route registrations
- Route test results
- Documentation of route order

**Testing Tollgate:**
- All routes accessible through API Gateway
- Routes match correct handlers
- No route conflicts

---

## Phase 1: Investigation & Verification (Week 1)

### Goal
Verify current state of dashboard pages and API Gateway routing before making changes.

### Tasks

#### Task 1.1: Dashboard Pages Audit
**Objective:** Verify dashboard pages are using correct portfolio-level endpoints.

**Current State (Based on Codebase Review):**
- **Business Intelligence Dashboard** (`frontend/app/dashboard/page.tsx`): Currently calls `getDashboardMetrics()` (v3 endpoint `/api/v3/dashboard/metrics`) - **NOT using portfolio endpoints**
- **Risk Dashboard** (`frontend/app/risk-dashboard/page.tsx`): Currently calls `getRiskMetrics()` (`/api/v1/risk/metrics`) - **NOT using analytics/trends/insights**
- **Risk Indicators Dashboard** (`frontend/app/risk-indicators/page.tsx`): Currently calls `getRiskMetrics()` - **NOT using aggregate indicators**

**Subtasks:**
1. **Business Intelligence Dashboard Review**
   - [x] Located: `frontend/app/dashboard/page.tsx`
   - [x] Current API call: `getDashboardMetrics()` → `/api/v3/dashboard/metrics`
   - [ ] **ISSUE:** Should call `GET /api/v1/merchants/analytics` (portfolio-wide)
   - [ ] **ISSUE:** Should call `GET /api/v1/merchants/statistics` (portfolio-wide)
   - [ ] Document discrepancy: Currently uses v3 dashboard metrics instead of portfolio analytics/statistics
   - [ ] Test current implementation to understand data structure
   - [ ] Verify if v3 endpoint provides portfolio data or if it's different
   - [ ] Determine if v3 endpoint should be replaced or used alongside portfolio endpoints

2. **Risk Dashboard Review**
   - [x] Located: `frontend/app/risk-dashboard/page.tsx`
   - [x] Current API call: `getRiskMetrics()` → `/api/v1/risk/metrics`
   - [ ] **ISSUE:** Should call `GET /api/v1/analytics/trends` (portfolio trends)
   - [ ] **ISSUE:** Should call `GET /api/v1/analytics/insights` (portfolio insights)
   - [ ] **NOTE:** Backend endpoints `/api/v1/analytics/trends` and `/api/v1/analytics/insights` return "not yet implemented" - **CRITICAL BLOCKER**
   - [ ] Document that analytics endpoints need backend implementation first
   - [ ] Verify if `getRiskMetrics()` provides sufficient portfolio data
   - [ ] Test current implementation

3. **Risk Indicators Dashboard Review**
   - [x] Located: `frontend/app/risk-indicators/page.tsx`
   - [x] Current API call: `getRiskMetrics()` → `/api/v1/risk/metrics`
   - [ ] **ISSUE:** Should call aggregate risk indicators endpoint (needs identification)
   - [ ] Identify correct aggregate risk indicators endpoint
   - [ ] Document current vs expected endpoints
   - [ ] Test current implementation

**Deliverables:**
- Dashboard audit report documenting current state
- List of discrepancies between expected and actual endpoints
- Recommendations for fixes (if needed)

**Testing Tollgate:**
- All dashboard pages load without errors
- All expected endpoints are called
- Data displayed matches expected structure

---

#### Task 1.2: API Gateway Route Audit
**Objective:** Verify all API Gateway routes are correctly configured.

**Current State (Based on Codebase Review):**
- **Merchant Routes:** `/merchants/analytics` explicitly registered (line 131), `/merchants/statistics` **NOT explicitly registered** (relies on PathPrefix at line 139)
- **Risk Routes:** `/risk/benchmarks` explicitly registered (line 168), `/risk/metrics` explicitly registered but maps to `/metrics` in handler
- **Analytics Routes:** `/analytics/trends` and `/analytics/insights` **NOT registered** - **CRITICAL GAP**

**Subtasks:**
1. **Route Registration Review**
   - [x] Reviewed: `services/api-gateway/cmd/main.go` route registrations
   - [x] Route order verified: Specific routes before PathPrefix (correct)
   - [x] **ISSUE FOUND:** `/merchants/statistics` not explicitly registered - relies on PathPrefix catch-all
   - [x] **ISSUE FOUND:** `/analytics/trends` and `/analytics/insights` **NOT registered at all**
   - [ ] Document all merchant-related routes (explicit + PathPrefix)
   - [ ] Document all risk-related routes
   - [ ] **CRITICAL:** Identify where analytics routes should be registered (risk-assessment service)
   - [ ] Verify CORS middleware is applied correctly
   - [ ] Check for route conflicts or shadowing
   - [ ] **ACTION REQUIRED:** Add explicit route registration for `/merchants/statistics`
   - [ ] **ACTION REQUIRED:** Add explicit route registration for `/analytics/trends` and `/analytics/insights`

2. **Handler Implementation Review**
   - [x] Reviewed: `services/api-gateway/internal/handlers/gateway.go`
   - [x] **FOUND:** `ProxyToMerchants` (line 147) - routes all merchant paths directly (no transformation)
   - [x] **FOUND:** `ProxyToRiskAssessment` (line 529) - has path transformations:
     - `/api/v1/risk/assess` → `/api/v1/assess` (line 574)
     - `/api/v1/risk/metrics` → `/api/v1/metrics` (line 576)
     - Other `/risk/*` paths kept as-is (line 578-581)
   - [ ] **ISSUE:** Analytics routes (`/analytics/trends`, `/analytics/insights`) are NOT handled - need handler registration
   - [ ] Verify path transformations are correct
   - [x] **FOUND:** SQL injection prevention in risk indicators handler (line 555)
   - [ ] Verify error handling is comprehensive
   - [ ] Document all path transformations
   - [ ] **ACTION REQUIRED:** Determine if analytics routes should use `ProxyToRiskAssessment` or new handler

3. **Route Testing**
   - [ ] Create test script to verify each route
   - [ ] Test merchant endpoints: `/api/v1/merchants/{id}`, `/api/v1/merchants/analytics`, `/api/v1/merchants/statistics`
   - [ ] **CRITICAL:** Test `/api/v1/merchants/statistics` - relies on PathPrefix catch-all (line 139 in main.go)
   - [ ] **CRITICAL:** Test `/api/v1/analytics/trends` - **NOT explicitly registered, may not route correctly**
   - [ ] **CRITICAL:** Test `/api/v1/analytics/insights` - **NOT explicitly registered, may not route correctly**
   - [ ] Test risk endpoints: `/api/v1/risk/benchmarks`, `/api/v1/risk/metrics` (maps to `/api/v1/metrics`)
   - [ ] Test with valid and invalid IDs
   - [ ] Test with query parameters
   - [ ] Verify responses match expected service responses
   - [ ] **Document routing gaps:** Analytics routes may need explicit registration

**Deliverables:**
- Route audit report
- Handler verification report
- Test results for all routes
- Documentation of route mappings

**Testing Tollgate:**
- All routes tested and verified
- No route conflicts identified
- All handlers correctly transform paths
- No security vulnerabilities found

---

#### Task 1.3: Backend Service Endpoint Verification
**Objective:** Verify backend services expose all required endpoints.

**Subtasks:**
1. **Merchant Service Verification**
   - [ ] Review `services/merchant-service/cmd/main.go` route registrations
   - [ ] Verify `GET /api/v1/merchants/analytics` endpoint exists
   - [ ] Verify `GET /api/v1/merchants/statistics` endpoint exists
   - [ ] Verify `GET /api/v1/merchants/{id}/risk-score` endpoint exists
   - [ ] Test each endpoint directly (bypassing API Gateway)
   - [ ] Verify response schemas match expected structure
   - [ ] Document any missing endpoints

2. **Risk Assessment Service Verification**
   - [ ] Review `services/risk-assessment-service/cmd/main.go` route registrations
   - [ ] Verify `GET /api/v1/risk/benchmarks` endpoint exists (✅ Registered at line 688)
   - [ ] **CRITICAL:** Verify `GET /api/v1/analytics/trends` endpoint exists (⚠️ Registered at line 785 but returns "not yet implemented")
   - [ ] **CRITICAL:** Verify `GET /api/v1/analytics/insights` endpoint exists (⚠️ Registered at line 786 but returns "not yet implemented")
   - [ ] **NOTE:** `GET /api/v1/risk/metrics` maps to `/api/v1/metrics` in API Gateway (line 576 in gateway.go)
   - [ ] Test each endpoint directly (bypassing API Gateway)
   - [ ] **BLOCKER:** Analytics endpoints return "not yet implemented" - backend implementation required
   - [ ] Verify response schemas match expected structure
   - [ ] Document implementation status for each endpoint

3. **Handler Implementation Verification**
   - [ ] Review merchant service handlers (`services/merchant-service/internal/handlers/merchant.go`)
   - [ ] Review risk assessment service handlers (`services/risk-assessment-service/internal/handlers/risk_assessment.go`)
   - [ ] Verify handlers return correct data structures
   - [ ] Test error handling scenarios
   - [ ] Verify database queries are optimized

**Deliverables:**
- Backend service endpoint verification report
- List of missing endpoints (if any)
- Response schema documentation
- Performance baseline measurements

**Testing Tollgate:**
- All required endpoints exist and are accessible
- Response schemas match expected structure
- Error handling is comprehensive
- Performance meets requirements (< 500ms p95)

---

#### Task 1.4: Frontend API Client Audit
**Objective:** Verify frontend API client functions exist and are correctly implemented.

**Subtasks:**
1. **API Client Review**
   - [x] Reviewed: `frontend/lib/api.ts`
   - [x] **FOUND:** `getRiskMetrics()` exists (line 806)
   - [x] **FOUND:** `getDashboardMetrics()` exists (line 705) - uses v3 endpoint
   - [ ] **MISSING:** `getPortfolioAnalytics()` - **NOT in api.ts**
   - [ ] **MISSING:** `getPortfolioStatistics()` - **NOT in api.ts**
   - [ ] **MISSING:** `getRiskTrends()` - **NOT in api.ts**
   - [ ] **MISSING:** `getRiskInsights()` - **NOT in api.ts**
   - [ ] **MISSING:** `getRiskBenchmarks()` - **NOT in api.ts**
   - [ ] **MISSING:** `getMerchantRiskScore()` - **NOT in api.ts**
   - [x] **FOUND:** Request deduplication exists (`RequestDeduplicator` class)
   - [x] **FOUND:** Retry logic exists (`retryWithBackoff` function)
   - [x] **FOUND:** Caching exists (`APICache` class)
   - [ ] Verify error handling in API functions
   - [ ] Document all existing API functions

2. **Type Definitions Review**
   - [ ] Review `frontend/types/merchant.ts` and related type files
   - [ ] Verify types match backend response schemas
   - [ ] Identify missing types for portfolio data
   - [ ] Document type gaps

**Deliverables:**
- API client audit report
- List of missing API functions
- Type definition gaps document

**Testing Tollgate:**
- All existing API functions work correctly
- Error handling is comprehensive
- Types are accurate

---

## Phase 2: Dashboard Pages Fixes (Week 2)

### Goal
Fix any issues found in dashboard pages and ensure they use correct portfolio-level endpoints.

### Prerequisites
**CRITICAL BLOCKERS TO RESOLVE FIRST:**
1. **Backend Implementation:** `/api/v1/analytics/trends` and `/api/v1/analytics/insights` return "not yet implemented" - must implement backend handlers first
2. **API Gateway Routing:** Analytics routes not registered - must add route registrations
3. **API Functions:** Missing frontend API functions for portfolio endpoints

### Tasks

#### Task 2.1: Fix Business Intelligence Dashboard
**Objective:** Ensure BI Dashboard uses correct portfolio endpoints.

**Current State:**
- Dashboard currently uses `getDashboardMetrics()` → `/api/v3/dashboard/metrics`
- Should use portfolio endpoints: `/api/v1/merchants/analytics` and `/api/v1/merchants/statistics`

**Subtasks:**
1. **API Gateway Route Verification (Prerequisite)**
   - [ ] Verify `/api/v1/merchants/statistics` routes correctly (currently relies on PathPrefix)
   - [ ] **RECOMMENDATION:** Add explicit route registration for `/merchants/statistics` in `services/api-gateway/cmd/main.go` before PathPrefix
   - [ ] Test route accessibility

2. **Create Missing API Functions**
   - [ ] Add `getPortfolioAnalytics()` function to `frontend/lib/api.ts`
     - Use `ApiEndpoints.merchants.analytics()` (portfolio-level, no ID)
     - Follow existing pattern: caching, deduplication, retry logic
     - Add to `ApiEndpoints.merchants` object if needed
   - [ ] Add `getPortfolioStatistics()` function to `frontend/lib/api.ts`
     - Use `ApiEndpoints.merchants.statistics()` (already exists at line 115 in api-config.ts)
     - Follow existing pattern: caching (5-10 min TTL), deduplication, retry logic

3. **Update Dashboard Component**
   - [ ] Update `frontend/app/dashboard/page.tsx` to call portfolio endpoints
   - [ ] Replace or supplement `getDashboardMetrics()` with `getPortfolioAnalytics()` and `getPortfolioStatistics()`
   - [ ] Determine if v3 dashboard metrics should be kept alongside portfolio data
   - [ ] Update charts to use portfolio data instead of mock data
   - [ ] Add error handling for portfolio data fetching

4. **Update Type Definitions**
   - [ ] Add `PortfolioAnalytics` type to `frontend/types/merchant.ts`
   - [ ] Add `PortfolioStatistics` type to `frontend/types/merchant.ts`
   - [ ] Ensure types match backend response schemas

5. **Update UI Components**
   - [ ] Verify dashboard displays portfolio-wide data correctly
   - [ ] Update charts/graphs to use portfolio data
   - [ ] Add loading states for portfolio data
   - [ ] Add error states for failed portfolio data fetches

**Deliverables:**
- Updated BI Dashboard component
- New API functions
- Updated type definitions
- Test results

**Testing Tollgate:**
- Dashboard loads without errors
- Portfolio analytics data displays correctly
- Portfolio statistics data displays correctly
- Error handling works correctly
- Loading states display correctly

---

#### Task 2.2: Fix Risk Dashboard
**Objective:** Ensure Risk Dashboard uses correct portfolio endpoints.

**Current State:**
- Dashboard currently uses `getRiskMetrics()` → `/api/v1/risk/metrics` (maps to `/api/v1/metrics`)
- Should use: `/api/v1/analytics/trends` and `/api/v1/analytics/insights` (but these return "not yet implemented")

**CRITICAL BLOCKER:**
- Backend endpoints `/api/v1/analytics/trends` and `/api/v1/analytics/insights` return "not yet implemented"
- Must implement backend handlers before frontend can use them

**Subtasks:**
1. **Backend Implementation (Prerequisite - HIGH PRIORITY)**
   - [ ] Implement `HandleRiskTrends` handler in `services/risk-assessment-service/internal/handlers/risk_assessment.go`
     - Currently returns "not yet implemented" (line 815)
     - Implement actual risk trends logic
   - [ ] Implement `HandleRiskInsights` handler in `services/risk-assessment-service/internal/handlers/risk_assessment.go`
     - Currently returns "not yet implemented" (line 821)
     - Implement actual risk insights logic
   - [ ] Test backend endpoints directly (bypassing API Gateway)

2. **API Gateway Route Registration (Prerequisite)**
   - [ ] Add explicit route registration for `/analytics/trends` in `services/api-gateway/cmd/main.go`
   - [ ] Add explicit route registration for `/analytics/insights` in `services/api-gateway/cmd/main.go`
   - [ ] Determine correct handler: `ProxyToRiskAssessment` (analytics routes are in risk-assessment service)
   - [ ] Test routes through API Gateway

3. **Create Missing API Functions**
   - [x] **FOUND:** `getRiskMetrics()` already exists (line 806 in api.ts)
   - [ ] Add `getRiskTrends()` function to `frontend/lib/api.ts`
     - Use `ApiEndpoints` builder (add to api-config.ts if needed)
     - Follow existing pattern: caching, deduplication, retry logic
   - [ ] Add `getRiskInsights()` function to `frontend/lib/api.ts`
     - Use `ApiEndpoints` builder (add to api-config.ts if needed)
     - Follow existing pattern: caching, deduplication, retry logic

4. **Update Dashboard Component**
   - [ ] Update `frontend/app/risk-dashboard/page.tsx` to call analytics endpoints
   - [ ] Keep `getRiskMetrics()` for metrics cards
   - [ ] Add `getRiskTrends()` for trend chart
   - [ ] Add `getRiskInsights()` for insights section
   - [ ] Replace mock chart data with real API data
   - [ ] Add error handling for portfolio data fetching

5. **Update Type Definitions**
   - [ ] Add `RiskTrends` type
   - [ ] Add `RiskInsights` type
   - [ ] Add `RiskMetrics` type (verify if already exists)
   - [ ] Ensure types match backend response schemas

6. **Update UI Components**
   - [ ] Verify dashboard displays portfolio risk data correctly
   - [ ] Update charts/graphs to use portfolio data
   - [ ] Add loading states for portfolio data
   - [ ] Add error states for failed portfolio data fetches

**Deliverables:**
- Updated Risk Dashboard component
- New API functions
- Updated type definitions
- Test results

**Testing Tollgate:**
- Dashboard loads without errors
- Portfolio risk trends data displays correctly
- Portfolio risk insights data displays correctly
- Portfolio risk metrics data displays correctly
- Error handling works correctly
- Loading states display correctly

---

#### Task 2.3: Fix Risk Indicators Dashboard
**Objective:** Ensure Risk Indicators Dashboard uses correct aggregate endpoints.

**Current State:**
- Dashboard currently uses `getRiskMetrics()` → `/api/v1/risk/metrics`
- Needs aggregate risk indicators endpoint (to be identified)

**Subtasks:**
1. **Identify Aggregate Endpoint**
   - [ ] Search backend services for aggregate risk indicators endpoint
   - [ ] Check if endpoint exists: `/api/v1/risk/indicators` (without merchant ID) or similar
   - [ ] Verify endpoint returns portfolio-wide aggregate data
   - [ ] Document endpoint path and response schema

2. **API Gateway Route Verification**
   - [ ] Verify aggregate endpoint routes correctly through API Gateway
   - [ ] Test endpoint accessibility
   - [ ] Add route registration if needed

3. **Create API Function**
   - [ ] Add aggregate risk indicators API function to `frontend/lib/api.ts`
     - Use `ApiEndpoints` builder (add to api-config.ts if needed)
     - Follow existing pattern: caching, deduplication, retry logic

4. **Update Dashboard Component**
   - [ ] Update `frontend/app/risk-indicators/page.tsx` to use aggregate endpoint
   - [ ] Keep `getRiskMetrics()` if still needed for overall metrics
   - [ ] Replace mock risk counts with real aggregate data
   - [ ] Replace mock trend data with real aggregate data
   - [ ] Add error handling for aggregate data fetching

5. **Update Type Definitions**
   - [ ] Add aggregate risk indicators type
   - [ ] Ensure types match backend response schemas

6. **Update UI Components**
   - [ ] Verify dashboard displays aggregate risk indicators correctly
   - [ ] Update charts/graphs to use aggregate data
   - [ ] Add loading states for aggregate data
   - [ ] Add error states for failed aggregate data fetches

**Deliverables:**
- Updated Risk Indicators Dashboard component
- New API functions (if needed)
- Updated type definitions
- Test results

**Testing Tollgate:**
- Dashboard loads without errors
- Aggregate risk indicators data displays correctly
- Error handling works correctly
- Loading states display correctly

---

## Phase 3: Merchant Details Page - Portfolio Comparison Features (Week 3-4)

### Goal
Add portfolio comparison features to merchant details page.

### Tasks

#### Task 3.1: Add Portfolio Statistics Comparison
**Objective:** Add merchant vs portfolio statistics comparison to MerchantOverviewTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/merchants/statistics` endpoint is accessible via API Gateway
   - [ ] Test endpoint with sample requests
   - [ ] Verify response schema matches expected structure

2. **Frontend API Function**
   - [ ] Add `getPortfolioStatistics()` function to `frontend/lib/api.ts` if not exists
   - [ ] Add error handling and retry logic
   - [ ] Add request deduplication
   - [ ] Add caching (5-10 minute TTL)

3. **Type Definitions**
   - [ ] Add `PortfolioStatistics` type to `frontend/types/merchant.ts`
   - [ ] Ensure type matches backend response schema
   - [ ] Add comparison result types

4. **UI Component**
   - [ ] Create `PortfolioComparisonCard` component
   - [ ] Add to `MerchantOverviewTab`
   - [ ] Display portfolio statistics summary
   - [ ] Display merchant vs portfolio comparison
   - [ ] Add visual indicators (above/below average badges)
   - [ ] Add loading state
   - [ ] Add error state with retry button

5. **Comparison Logic**
   - [ ] Calculate merchant risk score vs portfolio average
   - [ ] Calculate merchant position in portfolio distribution
   - [ ] Calculate percentile ranking
   - [ ] Display comparison metrics

**Deliverables:**
- `PortfolioComparisonCard` component
- Updated `MerchantOverviewTab` component
- API function implementation
- Type definitions
- Comparison logic implementation

**Testing Tollgate:**
- Portfolio statistics fetch successfully
- Comparison calculations are correct
- UI displays comparison data correctly
- Error handling works correctly
- Loading states display correctly
- Performance meets requirements (< 2s page load)

---

#### Task 3.2: Add Risk Benchmark Comparison
**Objective:** Add merchant vs industry benchmarks comparison to RiskAssessmentTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/risk/benchmarks?mcc={code}` endpoint is accessible via API Gateway
   - [ ] Test endpoint with various industry codes (MCC, NAICS, SIC)
   - [ ] Verify response schema matches expected structure
   - [ ] Test fallback behavior when industry code not found

2. **Frontend API Function**
   - [ ] Add `getRiskBenchmarks()` function to `frontend/lib/api.ts`
   - [ ] Accept `mcc`, `naics`, or `sic` parameters
   - [ ] Add error handling and retry logic
   - [ ] Add request deduplication
   - [ ] Add caching (10-15 minute TTL for benchmarks)

3. **Type Definitions**
   - [ ] Add `RiskBenchmarks` type
   - [ ] Add `BenchmarkComparison` type
   - [ ] Ensure types match backend response schema

4. **UI Component**
   - [ ] Create `RiskBenchmarkComparison` component
   - [ ] Add to `RiskAssessmentTab`
   - [ ] Extract industry code from merchant data (MCC, NAICS, or SIC)
   - [ ] Fetch benchmarks for merchant's industry
   - [ ] Display industry benchmark chart
   - [ ] Display merchant score vs benchmarks (average, median, 75th, 90th percentile)
   - [ ] Display percentile indicator
   - [ ] Add comparison table
   - [ ] Add loading state
   - [ ] Add error state with retry button
   - [ ] Handle case when industry code not available

5. **Comparison Logic**
   - [ ] Get merchant risk score (from assessment or risk-score endpoint)
   - [ ] Compare merchant score vs industry benchmarks
   - [ ] Calculate percentile position
   - [ ] Determine if merchant is above/below industry average

**Deliverables:**
- `RiskBenchmarkComparison` component
- Updated `RiskAssessmentTab` component
- API function implementation
- Type definitions
- Comparison logic implementation

**Testing Tollgate:**
- Benchmarks fetch successfully for valid industry codes
- Comparison calculations are correct
- UI displays benchmark comparison correctly
- Error handling works correctly (including missing industry code)
- Loading states display correctly
- Performance meets requirements

---

#### Task 3.3: Add Portfolio Analytics Comparison
**Objective:** Add merchant vs portfolio analytics comparison to BusinessAnalyticsTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/merchants/analytics` endpoint is accessible via API Gateway
   - [ ] Test endpoint to verify it returns portfolio-wide analytics
   - [ ] Verify response schema matches expected structure

2. **Frontend API Function**
   - [ ] Verify `getPortfolioAnalytics()` function exists in `frontend/lib/api.ts`
   - [ ] Add if missing with error handling and retry logic
   - [ ] Add request deduplication
   - [ ] Add caching (5-10 minute TTL)

3. **Type Definitions**
   - [ ] Add `PortfolioAnalytics` type if not exists
   - [ ] Add `AnalyticsComparison` type
   - [ ] Ensure types match backend response schema

4. **UI Component**
   - [ ] Create `AnalyticsComparison` component
   - [ ] Add to `BusinessAnalyticsTab`
   - [ ] Fetch portfolio analytics
   - [ ] Compare merchant analytics vs portfolio averages:
     - Classification confidence vs portfolio average
     - Security trust score vs portfolio average
     - Data quality vs portfolio average
   - [ ] Display side-by-side comparison
   - [ ] Display comparison charts (overlay or grouped)
   - [ ] Add difference indicators (positive/negative)
   - [ ] Add loading state
   - [ ] Add error state with retry button

5. **Comparison Logic**
   - [ ] Calculate differences between merchant and portfolio metrics
   - [ ] Determine if merchant is above/below portfolio average
   - [ ] Calculate percentage differences

**Deliverables:**
- `AnalyticsComparison` component
- Updated `BusinessAnalyticsTab` component
- API function implementation (if needed)
- Type definitions
- Comparison logic implementation

**Testing Tollgate:**
- Portfolio analytics fetch successfully
- Comparison calculations are correct
- UI displays comparison data correctly
- Charts render correctly
- Error handling works correctly
- Loading states display correctly
- Performance meets requirements

---

#### Task 3.4: Add Merchant Risk Score Integration
**Objective:** Add merchant risk score to MerchantOverviewTab for comparison.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/merchants/{id}/risk-score` endpoint is accessible via API Gateway
   - [ ] Test endpoint with sample merchant ID
   - [ ] Verify response schema matches expected structure

2. **Frontend API Function**
   - [ ] Add `getMerchantRiskScore()` function to `frontend/lib/api.ts`
   - [ ] Add error handling and retry logic
   - [ ] Add request deduplication
   - [ ] Add caching (2-5 minute TTL for risk scores)

3. **Type Definitions**
   - [ ] Add `MerchantRiskScore` type
   - [ ] Ensure type matches backend response schema

4. **UI Component**
   - [ ] Create `RiskScoreCard` component
   - [ ] Add to `MerchantOverviewTab`
   - [ ] Display merchant risk score
   - [ ] Display risk level (low/medium/high)
   - [ ] Add visual indicator (color-coded badge)
   - [ ] Add loading state
   - [ ] Add error state

5. **Integration with Portfolio Comparison**
   - [ ] Use risk score in portfolio comparison calculations
   - [ ] Display risk score vs portfolio average in comparison card

**Deliverables:**
- `RiskScoreCard` component
- Updated `MerchantOverviewTab` component
- API function implementation
- Type definitions

**Testing Tollgate:**
- Risk score fetches successfully
- UI displays risk score correctly
- Risk score is used in portfolio comparison
- Error handling works correctly
- Loading states display correctly

---

#### Task 3.5: Add Portfolio Context Badges
**Objective:** Add visual indicators showing merchant's position in portfolio.

**Subtasks:**
1. **Component Creation**
   - [ ] Create `PortfolioContextBadge` component
   - [ ] Support different badge types:
     - "Above Average" / "Below Average"
     - "Top 10%" / "Top 25%" / "Bottom 10%" / "Bottom 25%"
     - Industry ranking (e.g., "Top 5 in Industry")
   - [ ] Add color coding (green for good, red for concerning)
   - [ ] Add tooltips with detailed information

2. **Integration**
   - [ ] Add badges to merchant details page header
   - [ ] Add badges to relevant tabs (Overview, Risk Assessment, Analytics)
   - [ ] Calculate badge values from portfolio statistics and benchmarks

3. **Logic Implementation**
   - [ ] Calculate percentile from portfolio statistics
   - [ ] Determine badge text based on percentile
   - [ ] Calculate industry ranking from benchmarks

**Deliverables:**
- `PortfolioContextBadge` component
- Integration in merchant details page
- Badge calculation logic

**Testing Tollgate:**
- Badges display correctly
- Badge calculations are accurate
- Badges update when data changes
- Tooltips provide useful information

---

## Phase 4: High-Priority Features Implementation (Week 4-5)

### Goal
Implement high-priority features that enhance merchant details page functionality.

### Tasks

#### Task 4.1: Add Risk Alerts
**Objective:** Add risk alerts to RiskIndicatorsTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/risk/alerts/{merchantId}` endpoint exists
   - [ ] Test endpoint with sample merchant ID
   - [ ] Verify response schema

2. **Frontend API Function**
   - [ ] Add `getRiskAlerts()` function to `frontend/lib/api.ts`
   - [ ] Add error handling and retry logic

3. **Type Definitions**
   - [ ] Add `RiskAlert` type
   - [ ] Add `RiskAlerts` type

4. **UI Component**
   - [ ] Create `RiskAlertsSection` component
   - [ ] Add to `RiskIndicatorsTab`
   - [ ] Display active alerts prominently
   - [ ] Group alerts by severity
   - [ ] Add alert notifications (optional: toast notifications)
   - [ ] Add loading state
   - [ ] Add error state

**Deliverables:**
- `RiskAlertsSection` component
- Updated `RiskIndicatorsTab` component
- API function implementation
- Type definitions

**Testing Tollgate:**
- Alerts fetch successfully
- Alerts display correctly
- Alerts are grouped by severity
- Error handling works correctly

---

#### Task 4.2: Add Risk Explainability
**Objective:** Add risk explainability section to RiskAssessmentTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/risk/explain/{assessmentId}` endpoint exists
   - [ ] Test endpoint with sample assessment ID
   - [ ] Verify response schema (SHAP values, feature importance)

2. **Frontend API Function**
   - [ ] Verify `explainRiskAssessment()` function exists in `frontend/lib/api.ts`
   - [ ] Update if needed with error handling

3. **Type Definitions**
   - [ ] Add `RiskExplanation` type
   - [ ] Add types for SHAP values and feature importance

4. **UI Component**
   - [ ] Create `RiskExplainabilitySection` component
   - [ ] Add to `RiskAssessmentTab`
   - [ ] Display SHAP values
   - [ ] Display feature importance chart
   - [ ] Add explanations for risk factors
   - [ ] Add loading state
   - [ ] Add error state

**Deliverables:**
- `RiskExplainabilitySection` component
- Updated `RiskAssessmentTab` component
- Type definitions

**Testing Tollgate:**
- Explainability data fetches successfully
- SHAP values display correctly
- Feature importance chart renders correctly
- Error handling works correctly

---

#### Task 4.3: Add Risk Recommendations
**Objective:** Add risk recommendations section to RiskAssessmentTab.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/risk/recommendations/{merchantId}` endpoint exists
   - [ ] Test endpoint with sample merchant ID
   - [ ] Verify response schema

2. **Frontend API Function**
   - [ ] Verify `getRiskRecommendations()` function exists in `frontend/lib/api.ts`
   - [ ] Update if needed with error handling

3. **Type Definitions**
   - [ ] Add `RiskRecommendation` type
   - [ ] Add `RiskRecommendations` type

4. **UI Component**
   - [ ] Create `RiskRecommendationsSection` component
   - [ ] Add to `RiskAssessmentTab`
   - [ ] Display actionable recommendations
   - [ ] Group recommendations by priority
   - [ ] Add visual indicators (high/medium/low priority)
   - [ ] Add loading state
   - [ ] Add error state

**Deliverables:**
- `RiskRecommendationsSection` component
- Updated `RiskAssessmentTab` component
- Type definitions

**Testing Tollgate:**
- Recommendations fetch successfully
- Recommendations display correctly
- Recommendations are prioritized correctly
- Error handling works correctly

---

#### Task 4.4: Add Enrichment UI
**Objective:** Add enrichment button and status display to merchant details page.

**Subtasks:**
1. **Backend Integration**
   - [ ] Verify `GET /api/v1/merchants/{id}/enrichment-sources` endpoint exists
   - [ ] Verify `POST /api/v1/merchants/{id}/enrichment` endpoint exists
   - [ ] Test both endpoints

2. **Frontend API Function**
   - [ ] Verify `getEnrichmentSources()` function exists in `frontend/lib/api.ts`
   - [ ] Verify `triggerEnrichment()` function exists in `frontend/lib/api.ts`
   - [ ] Update if needed with error handling

3. **Type Definitions**
   - [ ] Add `EnrichmentSource` type
   - [ ] Add `EnrichmentStatus` type

4. **UI Component**
   - [ ] Create `EnrichmentButton` component
   - [ ] Add to merchant details page header or MerchantOverviewTab
   - [ ] Display enrichment sources
   - [ ] Add "Enrich Data" button
   - [ ] Show enrichment status (pending, in progress, completed, failed)
   - [ ] Add loading state during enrichment
   - [ ] Add error handling

**Deliverables:**
- `EnrichmentButton` component
- Integration in merchant details page
- Type definitions

**Testing Tollgate:**
- Enrichment sources fetch successfully
- Enrichment trigger works correctly
- Enrichment status displays correctly
- Error handling works correctly

---

## Phase 5: Testing & Quality Assurance (Week 5-6)

### Goal
Comprehensive testing of all implemented features.

### Tasks

#### Task 5.1: Unit Testing
**Objective:** Write unit tests for all new components and functions.

**Subtasks:**
1. **Frontend Unit Tests**
   - [ ] Write tests for all new API functions
   - [ ] Write tests for all new components
   - [ ] Write tests for comparison logic
   - [ ] Achieve minimum 80% code coverage
   - [ ] Test error handling scenarios
   - [ ] Test edge cases (empty data, null values, etc.)

2. **Backend Unit Tests (if handlers modified)**
   - [ ] Write tests for API Gateway handlers (if modified)
   - [ ] Write tests for service handlers (if modified)
   - [ ] Test path transformations
   - [ ] Test error handling

**Deliverables:**
- Unit test suite
- Test coverage report
- Test documentation

**Testing Tollgate:**
- All unit tests pass
- Minimum 80% code coverage achieved
- All error scenarios tested

---

#### Task 5.2: Integration Testing
**Objective:** Test end-to-end flows for all features.

**Subtasks:**
1. **Dashboard Integration Tests**
   - [ ] Test Business Intelligence Dashboard loads portfolio data
   - [ ] Test Risk Dashboard loads portfolio data
   - [ ] Test Risk Indicators Dashboard loads aggregate data
   - [ ] Test error handling when services are down
   - [ ] Test loading states

2. **Merchant Details Integration Tests**
   - [ ] Test merchant details page loads merchant data
   - [ ] Test portfolio comparison features load correctly
   - [ ] Test risk benchmark comparison
   - [ ] Test analytics comparison
   - [ ] Test risk alerts display
   - [ ] Test risk explainability
   - [ ] Test risk recommendations
   - [ ] Test enrichment flow
   - [ ] Test error handling for all features
   - [ ] Test concurrent tab switching

3. **API Gateway Integration Tests**
   - [ ] Test all routes through API Gateway
   - [ ] Test path transformations
   - [ ] Test error responses
   - [ ] Test CORS headers
   - [ ] Test authentication/authorization

**Deliverables:**
- Integration test suite
- Test results report
- Bug reports (if any)

**Testing Tollgate:**
- All integration tests pass
- All features work end-to-end
- Error handling works correctly
- Performance meets requirements

---

#### Task 5.3: Performance Testing
**Objective:** Verify performance meets requirements.

**Subtasks:**
1. **API Performance Tests**
   - [ ] Measure API response times for all endpoints
   - [ ] Verify p95 response time < 500ms
   - [ ] Test with concurrent requests
   - [ ] Test caching effectiveness
   - [ ] Identify and optimize slow queries

2. **Frontend Performance Tests**
   - [ ] Measure page load times
   - [ ] Verify merchant details page loads < 2 seconds
   - [ ] Test with slow network (throttled)
   - [ ] Test with large datasets
   - [ ] Optimize bundle size if needed

3. **Load Testing**
   - [ ] Test with multiple concurrent users
   - [ ] Test API Gateway under load
   - [ ] Test database queries under load
   - [ ] Identify bottlenecks

**Deliverables:**
- Performance test results
- Optimization recommendations
- Performance baseline documentation

**Testing Tollgate:**
- All performance requirements met
- No performance regressions
- Caching is effective

---

#### Task 5.4: Security Testing
**Objective:** Verify security of all implementations.

**Subtasks:**
1. **Input Validation Testing**
   - [ ] Test SQL injection prevention
   - [ ] Test XSS prevention
   - [ ] Test input sanitization
   - [ ] Test ID validation (UUID and custom formats)

2. **Authentication/Authorization Testing**
   - [ ] Test authentication requirements
   - [ ] Test authorization checks
   - [ ] Test token validation
   - [ ] Test unauthorized access attempts

3. **Error Handling Security**
   - [ ] Test error messages don't leak sensitive information
   - [ ] Test error responses are consistent
   - [ ] Test rate limiting

**Deliverables:**
- Security test results
- Security audit report
- Fixes for any vulnerabilities found

**Testing Tollgate:**
- No security vulnerabilities found
- All input validation works correctly
- Authentication/authorization works correctly

---

## Phase 6: Documentation & Deployment (Week 6)

### Goal
Document all changes and prepare for deployment.

### Tasks

#### Task 6.1: API Documentation
**Objective:** Document all API endpoints and changes.

**Subtasks:**
1. **Endpoint Documentation**
   - [ ] Document all portfolio-level endpoints
   - [ ] Document all comparison endpoints
   - [ ] Document request/response schemas
   - [ ] Document error responses
   - [ ] Add examples for each endpoint

2. **API Gateway Documentation**
   - [ ] Document route mappings
   - [ ] Document path transformations
   - [ ] Document authentication requirements
   - [ ] Document rate limiting

**Deliverables:**
- API documentation
- API Gateway documentation
- OpenAPI spec updates (if applicable)

---

#### Task 6.2: Frontend Documentation
**Objective:** Document frontend components and usage.

**Subtasks:**
1. **Component Documentation**
   - [ ] Document all new components
   - [ ] Document component props
   - [ ] Document usage examples
   - [ ] Document state management

2. **Architecture Documentation**
   - [ ] Document data flow for comparison features
   - [ ] Document caching strategy
   - [ ] Document error handling patterns

**Deliverables:**
- Component documentation
- Architecture documentation
- Usage examples

---

#### Task 6.3: Deployment Preparation
**Objective:** Prepare for production deployment.

**Subtasks:**
1. **Deployment Checklist**
   - [ ] Verify all tests pass
   - [ ] Verify performance meets requirements
   - [ ] Verify security checks pass
   - [ ] Prepare rollback plan
   - [ ] Prepare monitoring alerts

2. **Monitoring Setup**
   - [ ] Set up alerts for new endpoints
   - [ ] Set up performance monitoring
   - [ ] Set up error rate monitoring
   - [ ] Set up dashboard for new metrics

**Deliverables:**
- Deployment checklist
- Rollback plan
- Monitoring configuration

---

## Testing Tollgates Summary

### Phase 0 Tollgates
- ✅ Backend analytics endpoints implemented and return data
- ✅ API Gateway routes registered and accessible
- ✅ All routes tested and verified

### Phase 1 Tollgates
- ✅ All dashboard pages load without errors
- ✅ All expected endpoints are called
- ✅ All routes tested and verified
- ✅ All handlers correctly transform paths
- ✅ All required endpoints exist and are accessible
- ✅ All existing API functions work correctly

### Phase 2 Tollgates
- ✅ All dashboard pages load without errors
- ✅ Portfolio data displays correctly
- ✅ Error handling works correctly
- ✅ Loading states display correctly

### Phase 3 Tollgates
- ✅ Portfolio statistics fetch successfully
- ✅ Comparison calculations are correct
- ✅ UI displays comparison data correctly
- ✅ Error handling works correctly
- ✅ Performance meets requirements (< 2s page load)

### Phase 4 Tollgates
- ✅ All high-priority features implemented
- ✅ All features work correctly
- ✅ Error handling works correctly

### Phase 5 Tollgates
- ✅ All unit tests pass
- ✅ Minimum 80% code coverage achieved
- ✅ All integration tests pass
- ✅ All performance requirements met
- ✅ No security vulnerabilities found

### Phase 6 Tollgates
- ✅ All documentation complete
- ✅ Deployment checklist complete
- ✅ Monitoring configured

---

## Optimization Opportunities

### Performance Optimizations
1. **Caching Strategy**
   - **EXISTING:** `APICache` class already implemented in `frontend/lib/api-cache.ts`
   - **EXISTING:** Caching already used in API functions (5 min default TTL)
   - Leverage existing cache for portfolio statistics (5-10 min TTL)
   - Cache risk benchmarks (10-15 min TTL)
   - Cache merchant risk scores (2-5 min TTL)
   - **NOTE:** Redis caching would be backend-side, not frontend

2. **Request Optimization**
   - **EXISTING:** `RequestDeduplicator` class already implemented in `frontend/lib/request-deduplicator.ts`
   - **EXISTING:** Request deduplication already used in API functions
   - Leverage existing deduplication for parallel requests
   - Batch multiple portfolio requests if possible
   - **NOTE:** Do NOT use React Query - codebase uses useState + useEffect pattern

3. **Data Fetching Optimization**
   - Fetch portfolio data in parallel with merchant data
   - Lazy load comparison data only when tab is viewed
   - Use Suspense boundaries for better loading UX

### Code Quality Optimizations
1. **Type Safety**
   - Generate TypeScript types from OpenAPI spec
   - Add runtime validation with Zod
   - Ensure type consistency across frontend/backend

2. **Error Handling**
   - Implement consistent error handling patterns
   - Add error boundaries for component isolation
   - Provide user-friendly error messages

3. **Code Reusability**
   - Create reusable comparison components
   - Extract common comparison logic into utilities
   - Create shared types for comparison data

### User Experience Optimizations
1. **Loading States**
   - Implement skeleton loaders for better perceived performance
   - Show partial data while loading comparison data
   - Add progress indicators for long-running operations

2. **Visual Enhancements**
   - Add animations for data updates
   - Use color coding for above/below average indicators
   - Add tooltips with detailed information

3. **Accessibility**
   - Ensure all components are keyboard accessible
   - Add ARIA labels for screen readers
   - Ensure color contrast meets WCAG standards

---

## Investigation Items

### Required Investigations
1. **Dashboard Page Implementation** ✅ **COMPLETED**
   - [x] Investigated: All dashboard pages located and reviewed
   - [x] **FOUND:** BI Dashboard uses v3 endpoint, not portfolio endpoints
   - [x] **FOUND:** Risk Dashboard uses risk metrics, not analytics/trends/insights
   - [x] **FOUND:** Risk Indicators Dashboard uses risk metrics, not aggregate indicators
   - [x] Discrepancies documented in Task 1.1

2. **API Gateway Route Conflicts** ✅ **PARTIALLY COMPLETED**
   - [x] Investigated: Route order is correct (specific before PathPrefix)
   - [x] **FOUND:** `/merchants/statistics` relies on PathPrefix (not explicit)
   - [x] **FOUND:** `/analytics/trends` and `/analytics/insights` NOT registered
   - [ ] Test edge cases (similar route patterns)
   - [ ] Verify PathPrefix catch-all works correctly for statistics

3. **Backend Service Response Schemas** ⚠️ **IN PROGRESS**
   - [x] Investigated: Merchant service endpoints exist
   - [x] **FOUND:** Analytics/trends/insights return "not yet implemented"
   - [ ] Verify actual response schemas from working endpoints
   - [ ] Document schema differences
   - [ ] Create TypeScript types matching actual schemas

4. **Performance Baseline**
   - [ ] Measure current performance of merchant details page
   - [ ] Measure current performance of dashboard pages
   - [ ] Establish performance baseline for comparison

5. **Caching Infrastructure** ✅ **COMPLETED**
   - [x] **FOUND:** Frontend caching via `APICache` class (in-memory, 5 min default TTL)
   - [x] **FOUND:** Request deduplication via `RequestDeduplicator` class
   - [ ] Determine if Redis is available for backend-side caching
   - [x] Caching best practices identified: Use existing `APICache` and `RequestDeduplicator`

### Optional Investigations
1. **API Versioning Strategy** ⚠️ **NEEDS CLARIFICATION**
   - [x] **FOUND:** v3 endpoint exists: `/api/v3/dashboard/metrics` (line 117 in main.go)
   - [x] **FOUND:** v1 dashboard endpoint deprecated (line 147 commented out)
   - [ ] Determine relationship between v3 dashboard metrics and portfolio analytics/statistics
   - [ ] Clarify if v3 should be used alongside or instead of portfolio endpoints
   - [ ] Plan for future API changes

2. **Monitoring and Observability**
   - [ ] Investigate existing monitoring setup
   - [ ] Identify gaps in observability
   - [ ] Plan for enhanced monitoring

3. **Testing Infrastructure**
   - [x] **FOUND:** Test files exist: `frontend/__tests__/lib/api.test.ts`
   - [ ] Investigate existing test infrastructure
   - [ ] Identify testing best practices
   - [ ] Plan for test automation

---

## Success Criteria

### Technical Success
- ✅ All API Gateway routes correctly configured
- ✅ All handlers verified and tested
- ✅ All endpoints accessible and returning correct data
- ✅ All tests passing with 80%+ coverage
- ✅ Performance requirements met (< 500ms API, < 2s page load)
- ✅ No security vulnerabilities
- ✅ Error handling comprehensive

### Feature Success
- ✅ All dashboard pages verified and working correctly
- ✅ All portfolio comparison features implemented
- ✅ All high-priority features implemented
- ✅ All UI components implemented and tested
- ✅ All comparison calculations accurate

### User Experience Success
- ✅ Merchant details page loads quickly
- ✅ Comparison data displays clearly
- ✅ Error states are user-friendly
- ✅ Loading states provide good feedback
- ✅ All features are accessible

---

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|----------------|
| Phase 0: Critical Blockers Resolution | Week 0-1 | Backend analytics handlers, API Gateway routes |
| Phase 1: Investigation & Verification | Week 1 | Audit reports, verification results |
| Phase 2: Dashboard Pages Fixes | Week 2 | Fixed dashboard pages |
| Phase 3: Portfolio Comparison Features | Week 3-4 | Comparison components, API functions |
| Phase 4: High-Priority Features | Week 4-5 | Risk alerts, explainability, recommendations, enrichment |
| Phase 5: Testing & QA | Week 5-6 | Test suites, test results, performance reports |
| Phase 6: Documentation & Deployment | Week 6 | Documentation, deployment preparation |

**Total Estimated Duration:** 5-7 weeks (includes Phase 0 for critical blockers)

---

## Dependencies

### External Dependencies
- Backend services must be available and stable
- API Gateway must be deployed and accessible
- Database must be accessible
- Redis (if used for caching) must be available

### Internal Dependencies
- TypeScript types must be kept in sync with backend
- API client functions must be maintained
- Component library (shadcn UI) must be available
- Testing infrastructure must be set up

---

## Rollback Plan

### If Issues Found in Phase 1-2
- Revert dashboard page changes
- Document issues found
- Create new plan to address issues

### If Issues Found in Phase 3-4
- Disable comparison features via feature flag
- Revert to previous version of merchant details page
- Fix issues in separate branch
- Re-deploy after fixes

### If Issues Found in Phase 5-6
- Fix issues before deployment
- Re-run all tests
- Verify fixes don't introduce new issues

---

## Monitoring and Alerts

### Key Metrics to Monitor
1. **API Response Times**
   - Alert if p95 > 500ms
   - Alert if p99 > 1000ms

2. **Error Rates**
   - Alert if error rate > 0.1%
   - Alert on 404/502 errors

3. **Page Load Times**
   - Alert if merchant details page load time > 2s
   - Alert if dashboard page load time > 3s

4. **API Gateway Health**
   - Alert on routing errors
   - Alert on handler failures

5. **Service Health**
   - Alert if backend services are down
   - Alert on database connection issues

### Dashboards to Create
1. **API Gateway Dashboard**
   - Request rates by endpoint
   - Error rates by endpoint
   - Response times by endpoint

2. **Frontend Dashboard**
   - Page load times
   - Error rates by page
   - User engagement metrics

3. **Comparison Features Dashboard**
   - Comparison feature usage
   - Comparison data fetch success rates
   - Comparison calculation performance

---

## Conclusion

This comprehensive plan provides a structured approach to implementing portfolio comparison features and verifying dashboard implementations. By following this plan with careful attention to testing tollgates, risk mitigation, and quality assurance, we can ensure a successful implementation that enhances the merchant details page with valuable portfolio context while maintaining system reliability and performance.

**Next Steps:**
1. Review and approve this plan
2. **CRITICAL:** Begin Phase 0: Critical Blockers Resolution (backend analytics implementation)
3. Set up project tracking (Jira, GitHub Issues, etc.)
4. Schedule regular progress reviews

**Known Issues & Blockers:**
1. **Backend Analytics Endpoints:** `/api/v1/analytics/trends` and `/api/v1/analytics/insights` return "not yet implemented" - **MUST FIX FIRST**
2. **API Gateway Routing:** Analytics routes not registered - **MUST FIX FIRST**
3. **Missing API Functions:** Portfolio analytics/statistics functions don't exist in frontend
4. **Dashboard Endpoints:** Dashboards use wrong endpoints (v3 metrics instead of portfolio endpoints)
5. **State Management:** Plan mentions React Query but codebase uses useState + useEffect - **ALIGNED**

**Improvements Made:**
- Added Phase 0 to address critical blockers before main implementation
- Updated plan to reflect actual codebase state (useState pattern, existing utilities)
- Identified missing route registrations
- Documented backend implementation requirements
- Aligned with existing caching and deduplication utilities

---

**Document Version:** 1.1  
**Last Updated:** 2025-01-27  
**Status:** Reviewed and Updated - Aligned with Codebase  
**Owner:** Development Team

---

## Plan Review Summary

### Key Improvements Made

1. **Added Phase 0: Critical Blockers Resolution**
   - Identified backend analytics endpoints that return "not yet implemented"
   - Added tasks to implement backend handlers before frontend work
   - Added API Gateway route registration tasks

2. **Aligned with Actual Codebase State**
   - Updated state management section: Removed React Query reference (codebase uses useState + useEffect)
   - Documented existing utilities: `APICache`, `RequestDeduplicator`, `retryWithBackoff`
   - Identified actual dashboard implementations and their current endpoints

3. **Identified Critical Gaps**
   - Analytics routes not registered in API Gateway
   - `/merchants/statistics` relies on PathPrefix (not explicit)
   - Backend analytics handlers return "not yet implemented"
   - Missing frontend API functions for portfolio endpoints

4. **Enhanced Task Details**
   - Added "Current State" sections based on codebase review
   - Marked completed investigation items
   - Added specific file paths and line numbers
   - Added prerequisites and blockers clearly marked

5. **Improved Risk Mitigation**
   - Updated to reflect actual codebase patterns
   - Removed references to technologies not in use
   - Added specific verification tasks based on actual code structure

### Critical Blockers Identified

1. **Backend Implementation Required:**
   - `HandleRiskTrends` returns "not yet implemented" (line 815 in risk_assessment.go)
   - `HandleRiskInsights` returns "not yet implemented" (line 821 in risk_assessment.go)

2. **API Gateway Routing Gaps:**
   - `/analytics/trends` and `/analytics/insights` not registered
   - `/merchants/statistics` relies on PathPrefix catch-all

3. **Frontend API Functions Missing:**
   - `getPortfolioAnalytics()`
   - `getPortfolioStatistics()`
   - `getRiskTrends()`
   - `getRiskInsights()`
   - `getRiskBenchmarks()`
   - `getMerchantRiskScore()`

### Next Actions

1. **Immediate (Phase 0):**
   - Implement backend analytics handlers
   - Register analytics routes in API Gateway
   - Test route accessibility

2. **Short-term (Phase 1):**
   - Complete dashboard audit
   - Verify all routes work correctly
   - Document discrepancies

3. **Medium-term (Phase 2-3):**
   - Fix dashboard pages
   - Implement portfolio comparison features

### Notes on Linting Warnings

The plan document has markdown formatting warnings (blank lines around lists/headings). These are non-critical formatting issues and can be fixed in a separate formatting pass. The content and structure are correct and aligned with the codebase.

---

**Review Completed:** 2025-01-27  
**Reviewed By:** AI Assistant  
**Codebase Alignment:** ✅ Verified

