<!-- 456c0d6a-f19f-4ad9-afc0-f9d21abf4540 4bad98a4-de2f-473a-95de-28423169615e -->
# Comprehensive Frontend Testing and Debugging Plan

## Phase 1: Fix Critical API Configuration Issues

### 1.1 Verify and Fix Environment Variable Configuration

- **Issue**: `NEXT_PUBLIC_API_BASE_URL` not set during build, causing hardcoded `localhost:8080`
- **Files to check**:
- `frontend/lib/api.ts` (line 26)
- `frontend/app/layout.tsx` (lines 37-38)
- `frontend/next.config.ts` (line 11)
- `frontend/lib/websocket.ts` (line 213)
- `frontend/lib/preload.ts` (line 71)
- `frontend/components/performance/PerformanceOptimizer.tsx` (line 13)
- `frontend/app/register/page.tsx` (line 83)
- `frontend/app/sessions/page.tsx` (line 27)

- **Actions**:

1. Verify Railway environment variable `NEXT_PUBLIC_API_BASE_URL` is set to `https://api-gateway-service-production-21fd.up.railway.app`
2. Create a build verification script to check environment variables before build
3. Update all files using API_BASE_URL to use a centralized configuration
4. Add runtime validation to detect and warn if API_BASE_URL is localhost in production

### 1.2 Fix Next.js RSC 404 Errors

- **Issue**: 404 errors for `?_rsc=1n0h2` requests (React Server Components)
- **Affected routes**:
- `/compliance/progress-tracking?_rsc=1n0h2`
- `/compliance/gap-analysis?_rsc=1n0h2`
- `/risk-assessment/portfolio?_rsc=1n0h2`

- **Actions**:

1. Check if these pages are marked as client components when they should be server components
2. Verify Next.js App Router configuration
3. Check for missing route handlers or incorrect component exports
4. Review `next.config.ts` for RSC-related configuration issues

## Phase 2: Create Centralized API Configuration

### 2.1 Create API Configuration Utility

- **New file**: `frontend/lib/api-config.ts`
- **Purpose**: Single source of truth for API configuration
- **Features**:
- Runtime validation of API URL
- Environment detection (dev vs production)
- Warning system for misconfigured environments
- Type-safe API endpoint builders

### 2.2 Refactor All API Calls

- **Files to update**:
- `frontend/lib/api.ts` - Replace hardcoded API_BASE_URL
- `frontend/app/register/page.tsx` - Use centralized config
- `frontend/app/sessions/page.tsx` - Use centralized config
- `frontend/components/merchant/DataEnrichment.tsx` - Use centralized config (currently uses relative paths)
- `frontend/components/common/ExportButton.tsx` - Verify API calls
- `frontend/components/bulk-operations/BulkOperationsManager.tsx` - Verify API calls

## Phase 3: Comprehensive Page Testing

### 3.1 Create Page Testing Checklist

Test all 33 pages identified in the codebase:

#### Platform Pages

- [ ] `/` (Home) - Auto-redirect functionality
- [ ] `/dashboard-hub` - Dashboard links and navigation

#### Merchant Verification & Risk

- [ ] `/add-merchant` - Form submission, API calls, validation
- [ ] `/dashboard` - Business Intelligence metrics loading
- [ ] `/risk-dashboard` - Risk metrics and charts
- [ ] `/risk-indicators` - Risk indicator data loading

#### Compliance

- [ ] `/compliance` - Compliance status data
- [ ] `/compliance/gap-analysis` - Gap analysis functionality (currently 404 on RSC)
- [ ] `/compliance/progress-tracking` - Progress tracking (currently 404 on RSC)
- [ ] `/compliance/alerts` - Alert system
- [ ] `/compliance/framework-indicators` - Framework indicators
- [ ] `/compliance/summary-reports` - Summary reports

#### Merchant Management

- [ ] `/merchant-hub` - Merchant hub functionality
- [ ] `/merchant-hub/integration` - Integration features
- [ ] `/merchant-portfolio` - Merchant list, search, filters
- [ ] `/merchant-details/[id]` - Merchant detail page, all tabs
- [ ] `/merchant/bulk-operations` - Bulk operations
- [ ] `/merchant/comparison` - Merchant comparison
- [ ] `/risk-assessment/portfolio` - Risk portfolio (currently 404 on RSC)

#### Market Intelligence

- [ ] `/market-analysis` - Market analysis data
- [ ] `/competitive-analysis` - Competitive analysis

#### Administration

- [ ] `/admin` - Admin dashboard metrics
- [ ] `/admin/models` - Model management
- [ ] `/admin/queue` - Queue management
- [ ] `/sessions` - Session management API calls

#### Additional Pages

- [ ] `/register` - Registration form and API
- [ ] `/monitoring` - Monitoring dashboard
- [ ] `/analytics-insights` - Analytics insights
- [ ] `/business-intelligence` - Business intelligence
- [ ] `/business-growth` - Business growth metrics
- [ ] `/api-test` - API testing page
- [ ] `/gap-tracking` - Gap tracking
- [ ] `/gap-analysis/reports` - Gap analysis reports

### 3.2 Create Automated Testing Script

- **New file**: `frontend/scripts/test-all-pages.js`
- **Purpose**: Automated page accessibility and API call verification
- **Features**:
- Check all routes return 200 status
- Verify no console errors on page load
- Check for API calls to correct endpoints (not localhost)
- Verify data loading states
- Check for missing components or broken imports

## Phase 4: API Endpoint Verification

### 4.1 Create API Endpoint Inventory

Document all API endpoints used by frontend:

- Merchant endpoints: `/api/v1/merchants/*`
- Risk endpoints: `/api/v1/risk/*`
- Dashboard endpoints: `/api/v1/dashboard/*`, `/api/v3/dashboard/*`
- Compliance endpoints: `/api/v1/compliance/*`
- Session endpoints: `/api/v1/sessions`
- Auth endpoints: `/v1/auth/*`
- Business Intelligence: `/api/v1/business-intelligence/*`
- Monitoring: `/api/v1/monitoring/*`, `/api/v1/system/*`, `/api/v1/metrics`

### 4.2 Verify API Gateway Routing

- Check API Gateway service is accessible
- Verify CORS configuration allows frontend origin
- Test each endpoint category for proper routing
- Verify fallback endpoints work (e.g., v3 -> v1 fallback in dashboard metrics)

## Phase 5: Component and Feature Testing

### 5.1 Test Core Components

- [ ] `MerchantForm` - Form validation, submission, error handling
- [ ] `DataEnrichment` - Enrichment source loading, trigger functionality
- [ ] `ExportButton` - Export functionality for different formats
- [ ] `BulkOperationsManager` - Bulk update operations
- [ ] All chart components - Data visualization
- [ ] Data tables - Sorting, filtering, pagination
- [ ] Navigation components - Sidebar, breadcrumbs

### 5.2 Test Enhanced Features

- [ ] API caching functionality
- [ ] Request deduplication
- [ ] Error handling and retry logic
- [ ] Loading states and skeletons
- [ ] Toast notifications
- [ ] Modal dialogs
- [ ] Form validation
- [ ] Search and filter functionality

## Phase 6: Build and Deployment Verification

### 6.1 Pre-Build Checks

- Create script to verify all required environment variables
- Check for hardcoded localhost URLs
- Verify all imports are correct
- Check for TypeScript errors

### 6.2 Build Verification

- Verify environment variables are embedded in build
- Check build output for correct API URLs
- Verify no localhost references in production build
- Check bundle sizes and optimization

### 6.3 Post-Deployment Verification

- Health check script for all pages
- API connectivity test
- CORS verification
- Performance check

## Phase 7: Documentation and Monitoring

### 7.1 Create Testing Documentation

- Document all test cases
- Create troubleshooting guide
- Document known issues and workarounds
- Create deployment checklist

### 7.2 Add Monitoring

- Add error tracking for API failures
- Add performance monitoring
- Add usage analytics for features
- Create alerts for critical failures

## Implementation Order

1. **Immediate (Critical)**:

- Fix API base URL configuration
- Create centralized API config
- Fix RSC 404 errors
- Rebuild frontend with correct environment variables

2. **High Priority**:

- Test all pages for basic functionality
- Verify API endpoints are accessible
- Fix any broken API calls

3. **Medium Priority**:

- Comprehensive component testing
- Enhanced feature verification
- Performance optimization

4. **Ongoing**:

- Documentation
- Monitoring setup
- Continuous testing

## Success Criteria (Measurable)

### Critical (Must Pass)

1. **Zero 404 Errors**: All pages return 200 status, no RSC 404 errors
2. **Zero Localhost API Calls**: 100% of API calls go to Railway API Gateway (verified via network logs)
3. **Zero CORS Errors**: No CORS errors in browser console for 24 hours post-deployment
4. **All Forms Functional**: 100% of forms submit successfully (add-merchant, register, etc.)
5. **API Connectivity**: All API endpoints respond with <500ms latency (95th percentile)

### High Priority (Should Pass)

6. **Data Display**: All data loads correctly on all pages (no empty states due to API failures)
7. **Enhanced Features**: All enhanced features work (caching, deduplication, error handling)
8. **No UI Regressions**: Visual regression tests pass (screenshots match baseline)
9. **Performance**: Lighthouse scores maintain or improve:

- Performance: ≥90 (currently 98)
- Accessibility: 100 (maintain)
- Best Practices: ≥90
- SEO: ≥90

### Quality Metrics

10. **Error Rate**: <0.1% API error rate over 24 hours
11. **Test Coverage**: All existing tests pass + new tests for API config
12. **Bundle Size**: No >5% increase in bundle size
13. **Build Time**: Build completes successfully in <10 minutes
14. **Type Safety**: Zero TypeScript errors

### Technical Debt Reduction

15. **Code Consolidation**: All API calls use centralized config (0 direct fetch calls with hardcoded URLs)
16. **Legacy Code**: No references to deprecated legacy UI code
17. **Documentation**: All changes documented with migration guides

### To-dos

- [ ] Create centralized API configuration utility and refactor all API calls to use it
- [ ] Verify NEXT_PUBLIC_API_BASE_URL is set in Railway and create build verification script
- [ ] Fix React Server Component 404 errors for compliance and risk-assessment routes
- [ ] Rebuild frontend service in Railway with correct environment variables
- [ ] Test all platform pages (home, dashboard-hub) for functionality and API calls
- [ ] Test all merchant-related pages including add-merchant, portfolio, details, and bulk operations
- [ ] Test all compliance pages including gap-analysis and progress-tracking (fix 404s first)
- [ ] Test all risk assessment pages including dashboard, indicators, and portfolio
- [ ] Test all administration pages including admin dashboard, models, queue, and sessions
- [ ] Verify all API endpoints are accessible and properly routed through API Gateway
- [ ] Test all core components including forms, data tables, charts, and enhanced features
- [ ] Create automated testing script to verify all pages and API calls
- [ ] Create comprehensive testing documentation and deployment checklist