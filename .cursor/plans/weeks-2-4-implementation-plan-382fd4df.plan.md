---
name: Weeks 2-4 Implementation Plan
overview: ""
todos: []
---

# Weeks 2-4 Implementation Plan

## Overview

This plan implements the short-term actions for Weeks 2-4, building on Week 1 foundation. The plan covers backend API completion, performance optimizations, UX enhancements, and comprehensive QA testing.

**Timeline:** 15 working days

**Priority:** High

**Status:** Ready for Implementation

---

## Week 2: Complete Backend Integration + Frontend Migration Setup

### Task 2.0: React/Next.js Migration with shadcn/ui (NEW - HIGH PRIORITY)

**Duration:** 24-32 hours

**Priority:** High

**Owner:** Frontend Developer

**Timeline Impact:** Extends Week 2-3 timeline

#### 2.0.1 Set Up Next.js Project Structure

**Files to Create:**

- `frontend/package.json` - Next.js project configuration
- `frontend/next.config.js` - Next.js configuration
- `frontend/tsconfig.json` - TypeScript configuration
- `frontend/tailwind.config.js` - Tailwind CSS configuration (for shadcn/ui)
- `frontend/components.json` - shadcn/ui configuration
- `frontend/.env.local` - Environment variables

**Implementation Steps:**

1. **Initialize Next.js Project**

   - Create `frontend/` directory at project root
   - Run `npx create-next-app@latest frontend --typescript --tailwind --app`
   - Configure for static export (if needed) or API routes
   - Set up TypeScript strict mode

2. **Install and Configure shadcn/ui**

   - Install shadcn/ui CLI: `npx shadcn-ui@latest init`
   - Configure `components.json` with project paths
   - Set up Tailwind CSS with shadcn/ui theme
   - Install required dependencies (Radix UI, class-variance-authority, etc.)

3. **Set Up Project Structure**
   ```
   frontend/
   ├── app/                    # Next.js App Router
   │   ├── layout.tsx          # Root layout
   │   ├── page.tsx            # Home page
   │   └── merchant-details/   # Merchant details routes
   ├── components/             # React components
   │   ├── ui/                 # shadcn/ui components
   │   ├── merchant/           # Merchant-specific components
   │   └── risk/               # Risk assessment components
   ├── lib/                    # Utilities
   │   ├── utils.ts            # Utility functions
   │   └── api.ts              # API client
   ├── hooks/                  # Custom React hooks
   └── types/                  # TypeScript types
   ```

4. **Configure Build Process**

   - Set up Next.js build script
   - Configure output directory for Go service
   - Set up environment variables
   - Configure API proxy for development

**Deliverables:**

- Next.js project initialized
- shadcn/ui installed and configured
- Project structure created
- Build process configured

#### 2.0.2 Migrate Core Components to React

**Files to Create:**

- `frontend/components/ui/` - shadcn/ui components (Button, Card, Dialog, etc.)
- `frontend/components/merchant/merchant-details.tsx` - Main merchant details component
- `frontend/components/merchant/merchant-tabs.tsx` - Tab navigation component
- `frontend/app/merchant-details/[id]/page.tsx` - Merchant details page route

**Implementation Steps:**

1. **Install Required shadcn/ui Components**

   - Button, Card, Dialog, Tabs, Badge, Skeleton, Toast, Alert
   - Run: `npx shadcn-ui@latest add button card dialog tabs badge skeleton toast alert`
   - Customize theme colors to match existing design

2. **Create Merchant Details Layout**

   - Create `MerchantDetailsLayout` component
   - Implement tab navigation using shadcn/ui Tabs
   - Add responsive design
   - Integrate with existing API endpoints

3. **Migrate Tab Components**

   - Overview Tab → `MerchantOverviewTab.tsx`
   - Business Analytics Tab → `BusinessAnalyticsTab.tsx`
   - Risk Assessment Tab → `RiskAssessmentTab.tsx`
   - Risk Indicators Tab → `RiskIndicatorsTab.tsx`
   - Use shadcn/ui Card components for data display

4. **Create API Client**

   - Create `lib/api.ts` with typed API client
   - Use fetch with proper error handling
   - Implement request caching (from Task 3.1)
   - Add TypeScript types for all API responses

**Deliverables:**

- Core components migrated to React
- shadcn/ui components integrated
- Tab navigation working
- API client created

#### 2.0.3 Migrate JavaScript Components to React

**Files to Migrate:**

- `cmd/frontend-service/static/js/components/merchant-context.js` → `frontend/hooks/useMerchantContext.ts`
- `cmd/frontend-service/static/js/components/session-manager.js` → `frontend/hooks/useSessionManager.ts`
- `cmd/frontend-service/static/js/components/risk-score-panel.js` → `frontend/components/risk/RiskScorePanel.tsx`
- `cmd/frontend-service/static/js/components/data-enrichment.js` → `frontend/components/merchant/DataEnrichment.tsx`
- All other components in `js/components/` directory

**Implementation Steps:**

1. **Convert Class-Based Components to React Hooks**

   - Convert JavaScript classes to React functional components
   - Use `useState`, `useEffect`, `useContext` hooks
   - Maintain existing functionality

2. **Migrate Risk Components**

   - Risk Score Panel → React component with shadcn/ui Card
   - Risk Visualization → React component with Chart.js integration
   - Risk Indicators → React component with shadcn/ui Badge and Alert
   - Risk History → React component with shadcn/ui Table

3. **Migrate Utility Components**

   - Navigation → React component
   - Export Button → React component with shadcn/ui Button
   - Loading States → shadcn/ui Skeleton components
   - Toast Notifications → shadcn/ui Toast (from Task 3.5)

4. **Update Styling**

   - Replace custom CSS with Tailwind CSS classes
   - Use shadcn/ui theme variables
   - Ensure responsive design maintained

**Deliverables:**

- All JavaScript components migrated to React
- Functionality preserved
- shadcn/ui styling applied
- Responsive design maintained

#### 2.0.4 Update Go Service to Serve Next.js Build

**Files to Modify:**

- `cmd/frontend-service/main.go` - Update to serve Next.js build output
- `cmd/frontend-service/Dockerfile` - Add Next.js build step
- `cmd/frontend-service/build-frontend.sh` - Update build script

**Implementation Steps:**

1. **Update Build Process**

   - Add Next.js build step before Go build
   - Configure Next.js to output to `cmd/frontend-service/static/` or separate directory
   - Update Dockerfile to build Next.js app
   - Ensure static files are served correctly

2. **Update Go Service**

   - Modify `main.go` to serve Next.js build output
   - Handle Next.js routing (catch-all for client-side routing)
   - Maintain API proxy functionality
   - Update health check endpoint

3. **Development Setup**

   - Configure Next.js dev server to proxy API calls
   - Set up hot reload for development
   - Ensure environment variables are passed correctly

**Deliverables:**

- Go service updated to serve Next.js build
- Build process working
- Development setup configured
- Production deployment ready

#### 2.0.5 Testing and Validation

**Files to Create:**

- `frontend/__tests__/` - React component tests
- `frontend/.eslintrc.json` - ESLint configuration
- `frontend/jest.config.js` - Jest configuration (if using Jest)

**Implementation Steps:**

1. **Component Testing**

   - Write tests for migrated components
   - Test shadcn/ui component integration
   - Test API integration
   - Test responsive design

2. **Integration Testing**

   - Test full merchant details page flow
   - Test tab navigation
   - Test API calls and error handling
   - Test loading states

3. **Visual Regression Testing**

   - Compare new design with existing design
   - Ensure all features work correctly
   - Test cross-browser compatibility

**Deliverables:**

- Component tests written
- Integration tests passing
- Visual regression tests passing
- All functionality validated

### Task 2.1: Enhance Business Analytics Endpoints

**Files to Modify:**

- `internal/services/merchant_analytics_service.go` - Add parallel fetching and caching
- `internal/api/handlers/merchant_analytics_handler.go` - Already exists, verify integration
- `internal/database/merchant_analytics_repository.go` - Verify repository methods exist

**Implementation Steps:**

1. **Enhance GetMerchantAnalytics with Parallel Fetching**

   - Modify `GetMerchantAnalytics` in `merchant_analytics_service.go` to use goroutines for parallel data fetching
   - Fetch classification, security, quality, intelligence, and verification data concurrently
   - Use sync.WaitGroup and mutex for thread-safe aggregation
   - Add timeout context (30 seconds)

2. **Add Caching Layer**

   - Integrate with existing cache infrastructure (check `pkg/cache/`)
   - Add cache check before database queries
   - Cache analytics results with 5-minute TTL
   - Use cache key format: `analytics:{merchantId}`

3. **Enhance Error Handling**

   - Validate merchant exists before fetching analytics
   - Check merchant status (active/inactive)
   - Return partial data if some sources fail (non-critical errors)
   - Return error only if critical data (classification) fails

4. **Complete Website Analysis Endpoint**

   - Verify `GetWebsiteAnalysis` implementation in `merchant_analytics_service.go`
   - Ensure it triggers new analysis if data is older than 24 hours
   - Add background job for website analysis if needed

**Testing:**

- Unit tests for parallel fetching logic
- Integration tests for caching behavior
- Performance tests for concurrent requests
- Error handling tests for partial failures

---

### Task 2.2: Complete Risk Assessment Endpoints

**Files to Modify:**

- `internal/services/risk_assessment_service.go` - Add missing methods
- `internal/api/handlers/async_risk_assessment_handler.go` - Add new handlers
- `internal/jobs/risk_assessment_job.go` - Complete background job processing
- `internal/api/routes/risk_routes.go` - Register new routes

**Implementation Steps:**

1. **Complete Background Job Processing**

   - Enhance `Process` method in `risk_assessment_job.go`
   - Update assessment status throughout processing lifecycle
   - Save assessment results to repository
   - Handle errors and update status accordingly

2. **Implement GET /api/v1/risk/history/{merchantId}**

   - Add `GetRiskHistory` method to `RiskAssessmentService`
   - Create handler method in `AsyncRiskAssessmentHandler`
   - Support pagination with limit/offset query parameters
   - Register route in `risk_routes.go`

3. **Implement GET /api/v1/risk/predictions/{merchantId}**

   - Add `GetPredictions` method to service
   - Create handler method
   - Support optional query parameters: `horizons`, `includeScenarios`, `includeConfidence`
   - Register route

4. **Implement GET /api/v1/risk/explain/{assessmentId}**

   - Add `ExplainAssessment` method to service
   - Create handler method
   - Return explainability data for risk assessment
   - Register route

5. **Implement GET /api/v1/merchants/{merchantId}/risk-recommendations**

   - Add `GetRecommendations` method to service
   - Create handler method
   - Return actionable risk mitigation recommendations
   - Register route in merchant routes

**Testing:**

- Unit tests for all new service methods
- Integration tests for background job processing
- API endpoint tests for all new routes
- Error handling tests

---

### Task 2.3: Implement Risk Indicators Endpoints

**Files to Create:**

- `internal/services/risk_indicators_service.go` - New service
- `internal/api/handlers/risk_indicators_handler.go` - New handler
- `internal/models/risk_indicators.go` - Data models
- `internal/database/risk_indicators_repository.go` - Repository

**Implementation Steps:**

1. **Create Risk Indicators Service**

   - Define `RiskIndicatorsService` interface
   - Implement `GetRiskIndicators` method
   - Calculate overall risk score from indicators
   - Filter indicators by severity and status

2. **Create Risk Indicators Handler**

   - Implement `GetRiskIndicators` handler
   - Implement `GetRiskAlerts` handler
   - Support query parameters: `severity`, `status`
   - Register routes in `risk_routes.go`

3. **Create Data Models**

   - Define `RiskIndicatorsData` struct
   - Define `RiskIndicator` struct with fields: ID, Type, Name, Severity, Status, Description, DetectedAt, Score

4. **Create Repository**

   - Implement database queries for risk indicators
   - Support filtering by merchant ID, severity, status
   - Add indexes for performance

**Testing:**

- Unit tests for service logic
- Integration tests for repository
- API endpoint tests
- Filtering and sorting tests

---

### Task 2.4: Complete Data Enrichment Integration

**Files to Modify/Create:**

- `internal/services/data_enrichment_service.go` - Complete service implementation
- `internal/api/handlers/data_enrichment_handler.go` - Create handler
- `cmd/frontend-service/static/js/components/data-enrichment.js` - Enhance frontend component

**Implementation Steps:**

1. **Backend: Complete Enrichment Service**

   - Implement `TriggerEnrichment` method
   - Validate enrichment source
   - Create enrichment job
   - Queue job for background processing
   - Implement `GetEnrichmentSources` method

2. **Backend: Create Handler**

   - Implement POST `/api/v1/merchants/{merchantId}/enrichment/trigger`
   - Implement GET `/api/v1/merchants/{merchantId}/enrichment/sources`
   - Register routes in merchant routes

3. **Frontend: Enhance Data Enrichment Component**

   - Verify `data-enrichment.js` exists
   - Add `triggerEnrichment` method
   - Add `getEnrichmentSources` method
   - Add error handling and loading states
   - Integrate with merchant details page

**Testing:**

- Backend service tests
- API endpoint tests
- Frontend integration tests
- Error handling tests

---

### Task 2.5: Complete External Data Sources Integration

**Files to Modify/Create:**

- `internal/services/external_data_sources_service.go` - Create service
- `internal/api/handlers/external_data_sources_handler.go` - Create handler
- `cmd/frontend-service/static/js/components/external-data-sources.js` - Verify/enhance component

**Implementation Steps:**

Similar structure to Data Enrichment:

1. Create service with methods to fetch external data
2. Create handler with API endpoints
3. Enhance frontend component
4. Register routes

**Testing:**

- Service tests
- API endpoint tests
- Frontend integration tests

---

### Task 2.6: Implement Consistent Error Handling

**Files to Create/Modify:**

- `pkg/errors/api_errors.go` - Standardized error types
- `internal/api/middleware/error_handling.go` - Enhance existing middleware
- `cmd/frontend-service/static/js/utils/error-handler.js` - Frontend error utility

**Implementation Steps:**

1. **Backend: Standardized Error Types**

   - Create `APIError` struct with Code, Message, Details
   - Implement `WriteError` helper function
   - Create error code constants
   - Add retry logic with exponential backoff

2. **Backend: Enhance Error Middleware**

   - Review existing `error_handling.go`
   - Ensure consistent error response format
   - Add error logging
   - Map internal errors to HTTP status codes

3. **Frontend: Error Handling Utility**

   - Create `ErrorHandler` class
   - Implement `handleAPIError` method
   - Add `showErrorNotification` for user feedback
   - Add `logError` for debugging
   - Integrate with existing error handling

**Testing:**

- Error response format tests
- Retry logic tests
- Frontend error display tests
- Error logging verification

---

## Week 3: Performance Optimization

### Task 3.1: API Request Optimization

**Files to Create:**

- `cmd/frontend-service/static/js/utils/api-batcher.js` - Request batching
- `cmd/frontend-service/static/js/utils/api-cache.js` - Response caching
- `cmd/frontend-service/static/js/utils/request-deduplicator.js` - Request deduplication

**Implementation Steps:**

1. **Implement Request Batching**

   - Create `APIBatcher` class
   - Batch requests within 100ms window
   - Deduplicate pending requests
   - Integrate with existing API calls

2. **Implement Response Caching**

   - Create `APICache` class with TTL support
   - Use Map for in-memory cache
   - Add sessionStorage persistence
   - Default TTL: 5 minutes
   - Cache key format: `{url}-{options}`

3. **Implement Request Deduplication**

   - Create `RequestDeduplicator` class
   - Track pending requests by key
   - Return existing promise for duplicate requests
   - Clean up after request completes

4. **Integrate with Merchant Details Page**

   - Update `merchant-details.html` to use new utilities
   - Replace direct fetch calls with cached/batched versions
   - Measure performance improvements

**Testing:**

- Unit tests for each utility class
- Integration tests with real API calls
- Performance benchmarks
- Cache hit/miss ratio tests

---

### Task 3.2: Lazy Loading Enhancements

**Files to Create/Modify:**

- `cmd/frontend-service/static/js/utils/lazy-loader.js` - Lazy loading utility
- `cmd/frontend-service/static/merchant-details.html` - Integrate lazy loading

**Implementation Steps:**

1. **Implement Lazy Loader**

   - Create `LazyLoader` class using IntersectionObserver
   - Observe expandable sections
   - Load data when section becomes visible
   - Track loaded sections to prevent duplicate loads

2. **Defer Non-Critical API Calls**

   - Identify non-critical API calls (analytics, recommendations, external sources)
   - Use `requestIdleCallback` or `setTimeout` fallback
   - Load after page load completes
   - Add loading indicators

3. **Integrate with Merchant Details**

   - Apply lazy loading to expandable sections
   - Defer non-critical data loading
   - Measure performance improvements

**Testing:**

- IntersectionObserver behavior tests
- Lazy loading trigger tests
- Performance impact measurements
- User experience validation

---

### Task 3.3: Loading State Improvements

**Files to Create:**

- `cmd/frontend-service/static/js/components/skeleton-loader.js` - Skeleton loaders
- `cmd/frontend-service/static/js/components/progress-indicator.js` - Progress indicators
- `cmd/frontend-service/static/css/skeleton-loaders.css` - Skeleton styles

**Implementation Steps:**

1. **Implement Skeleton Loaders**

   - Create `SkeletonLoader` class
   - Add methods: `createCardSkeleton`, `show`, `hide`
   - Add CSS animations for pulse effect
   - Support different skeleton types (card, list, table)

2. **Implement Progress Indicators**

   - Create `ProgressIndicator` class
   - Show/hide progress bar
   - Update progress percentage
   - Estimate time remaining
   - Support multiple progress bars

3. **Integrate with Merchant Details**

   - Replace loading spinners with skeleton loaders
   - Add progress indicators for long operations
   - Improve perceived performance

**Testing:**

- Skeleton loader display tests
- Progress indicator accuracy tests
- Animation performance tests
- User experience validation

---

### Task 3.4: Empty State Design

**Files to Create:**

- `cmd/frontend-service/static/js/components/empty-state.js` - Empty state component
- `cmd/frontend-service/static/css/empty-states.css` - Empty state styles

**Implementation Steps:**

1. **Create Empty State Component**

   - Create `EmptyState` class
   - Support types: `noData`, `error`, `noResults`
   - Include icon, title, message, and optional action button
   - Make it reusable across tabs

2. **Apply to All Tabs**

   - Add empty states to Overview tab
   - Add empty states to Analytics tab
   - Add empty states to Risk tab
   - Add empty states to other tabs as needed

3. **Add Helpful Guidance**

   - Provide context-specific messages
   - Include actionable CTAs where appropriate
   - Add retry functionality for errors

**Testing:**

- Empty state display tests
- Error state handling tests
- CTA functionality tests
- User experience validation

---

### Task 3.5: Success Feedback Implementation

**Files to Create:**

- `cmd/frontend-service/static/js/components/toast-notification.js` - Toast notifications
- `cmd/frontend-service/static/css/toast-notifications.css` - Toast styles

**Implementation Steps:**

1. **Implement Toast Notifications**

   - Create `ToastNotification` class
   - Support types: `success`, `error`, `info`, `warning`
   - Auto-dismiss after duration (default 3s, errors 5s)
   - Animate in/out
   - Support multiple toasts

2. **Add Success Animations**

   - Add success animations for completed actions
   - Add confirmation dialogs for critical actions
   - Improve user feedback

3. **Integrate Throughout Application**

   - Use toasts for API success/error responses
   - Use toasts for form submissions
   - Use toasts for data updates

**Testing:**

- Toast display tests
- Animation tests
- Multiple toast handling tests
- User experience validation

---

## Week 4: Quality Assurance

### Task 4.1: Execute Comprehensive Test Suites

**Files to Reference:**

- `cmd/frontend-service/static/docs/navigation-testing-guide.md` - Navigation tests
- `cmd/frontend-service/static/docs/export-functionality-testing-guide.md` - Export tests
- `cmd/frontend-service/static/docs/cross-browser-testing-guide.md` - Cross-browser tests

**Implementation Steps:**

1. **Execute Navigation Testing**

   - Follow navigation testing guide
   - Execute all 50+ test cases
   - Document results
   - Report issues

2. **Execute Export Testing**

   - Follow export testing guide
   - Test all formats (CSV, PDF, JSON, Excel)
   - Test all tabs
   - Document results

3. **Execute Cross-Browser Testing**

   - Test in Chrome, Firefox, Safari, Edge
   - Execute all 80+ test cases
   - Document browser-specific issues
   - Verify fixes

**Deliverables:**

- Test execution reports
- Issue tracking
- Test coverage metrics

---

### Task 4.2: Fix Identified Issues

**Process:**

1. Prioritize issues (Critical, High, Medium, Low)
2. Assign to developers
3. Fix issues
4. Verify fixes
5. Update tests if needed

**Deliverables:**

- All critical issues fixed
- High-priority issues fixed
- Medium-priority issues addressed
- Issue tracking updated

---

### Task 4.3: Prepare for Beta Release

**Files to Create:**

- `docs/release-notes/beta-release-notes.md` - Release notes
- `docs/beta-tester-guide.md` - Beta tester guide
- `docs/technical-documentation/beta-deployment-guide.md` - Deployment guide

**Implementation Steps:**

1. **Create Beta Release Checklist**

   - Verify all critical bugs fixed
   - Verify all high-priority features implemented
   - Check test coverage (target: 80%+ unit, 100% critical paths)
   - Verify performance metrics
   - Verify cross-browser compatibility
   - Verify accessibility compliance (WCAG 2.1 AA)

2. **Prepare Release Documentation**

   - Create release notes with new features, bug fixes, known issues
   - Create beta tester guide with access instructions, testing focus, issue reporting
   - Update technical documentation (API docs, deployment guide, rollback procedure)

3. **Prepare Beta Tester Communication**

   - Create beta access instructions
   - Prepare feedback collection process
   - Set up issue tracking system

**Deliverables:**

- Beta release checklist complete
- Release documentation prepared
- Beta tester communication ready
- Ready for beta release

---

## Success Criteria

### Completion Checklist

- [ ] All high-priority API endpoints implemented
- [ ] All endpoints integrated with frontend
- [ ] Error handling standardized
- [ ] Performance optimizations implemented
- [ ] Loading states improved
- [ ] Empty states created
- [ ] Success feedback implemented
- [ ] All test suites executed
- [ ] All critical issues fixed
- [ ] Beta release prepared

---

## Detailed Implementation Todos

### Week 2: Complete Backend Integration + Frontend Migration Setup

#### Task 2.0: React/Next.js Migration with shadcn/ui

**2.0.1 Set Up Next.js Project Structure**

- [ ] **2.0.1.1** Create `frontend/` directory at project root
- [ ] **2.0.1.2** Run `npx create-next-app@latest frontend --typescript --tailwind --app --no-src-dir`
- [ ] **2.0.1.3** Verify `frontend/package.json` created with Next.js dependencies
- [ ] **2.0.1.4** Configure TypeScript strict mode in `frontend/tsconfig.json`
- [ ] **2.0.1.5** Create `frontend/next.config.js` with static export configuration (if needed)
- [ ] **2.0.1.6** Set up `frontend/.env.local` with API base URL and environment variables
- [ ] **2.0.1.7** Install shadcn/ui CLI: `npx shadcn-ui@latest init` in frontend directory
- [ ] **2.0.1.8** Configure `frontend/components.json` with project paths (components, utils, styles)
- [ ] **2.0.1.9** Verify `frontend/tailwind.config.js` includes shadcn/ui theme configuration
- [ ] **2.0.1.10** Install required dependencies: `npm install class-variance-authority clsx tailwind-merge lucide-react`
- [ ] **2.0.1.11** Create `frontend/lib/utils.ts` with `cn()` utility function
- [ ] **2.0.1.12** Create directory structure: `frontend/app/`, `frontend/components/ui/`, `frontend/components/merchant/`, `frontend/components/risk/`, `frontend/lib/`, `frontend/hooks/`, `frontend/types/`
- [ ] **2.0.1.13** Create `frontend/app/layout.tsx` with root layout including metadata
- [ ] **2.0.1.14** Create `frontend/app/page.tsx` as home page placeholder
- [ ] **2.0.1.15** Configure Next.js build output directory (if serving from Go service)
- [ ] **2.0.1.16** Set up API proxy in `frontend/next.config.js` for development
- [ ] **2.0.1.17** Test Next.js dev server starts: `npm run dev`
- [ ] **2.0.1.18** Verify shadcn/ui components can be installed: `npx shadcn-ui@latest add button`

**2.0.2 Migrate Core Components to React**

- [ ] **2.0.2.1** Install shadcn/ui Button component: `npx shadcn-ui@latest add button`
- [ ] **2.0.2.2** Install shadcn/ui Card component: `npx shadcn-ui@latest add card`
- [ ] **2.0.2.3** Install shadcn/ui Dialog component: `npx shadcn-ui@latest add dialog`
- [ ] **2.0.2.4** Install shadcn/ui Tabs component: `npx shadcn-ui@latest add tabs`
- [ ] **2.0.2.5** Install shadcn/ui Badge component: `npx shadcn-ui@latest add badge`
- [ ] **2.0.2.6** Install shadcn/ui Skeleton component: `npx shadcn-ui@latest add skeleton`
- [ ] **2.0.2.7** Install shadcn/ui Toast component: `npx shadcn-ui@latest add toast`
- [ ] **2.0.2.8** Install shadcn/ui Alert component: `npx shadcn-ui@latest add alert`
- [ ] **2.0.2.9** Customize shadcn/ui theme colors in `frontend/tailwind.config.js` to match existing design
- [ ] **2.0.2.10** Create `frontend/types/merchant.ts` with TypeScript interfaces for Merchant, AnalyticsData, RiskAssessment
- [ ] **2.0.2.11** Create `frontend/lib/api.ts` with typed API client functions
- [ ] **2.0.2.12** Implement `getMerchant(merchantId: string)` function in `frontend/lib/api.ts`
- [ ] **2.0.2.13** Implement `getMerchantAnalytics(merchantId: string)` function in `frontend/lib/api.ts`
- [ ] **2.0.2.14** Implement `getRiskAssessment(merchantId: string)` function in `frontend/lib/api.ts`
- [ ] **2.0.2.15** Add error handling and retry logic to API client functions
- [ ] **2.0.2.16** Create `frontend/components/merchant/MerchantDetailsLayout.tsx` component
- [ ] **2.0.2.17** Implement tab navigation in `MerchantDetailsLayout.tsx` using shadcn/ui Tabs
- [ ] **2.0.2.18** Add responsive design classes to `MerchantDetailsLayout.tsx`
- [ ] **2.0.2.19** Create `frontend/components/merchant/MerchantOverviewTab.tsx` component
- [ ] **2.0.2.20** Create `frontend/components/merchant/BusinessAnalyticsTab.tsx` component
- [ ] **2.0.2.21** Create `frontend/components/merchant/RiskAssessmentTab.tsx` component
- [ ] **2.0.2.22** Create `frontend/components/merchant/RiskIndicatorsTab.tsx` component
- [ ] **2.0.2.23** Use shadcn/ui Card components for data display in each tab
- [ ] **2.0.2.24** Create `frontend/app/merchant-details/[id]/page.tsx` route
- [ ] **2.0.2.25** Integrate `MerchantDetailsLayout` in merchant details page route
- [ ] **2.0.2.26** Add loading state using shadcn/ui Skeleton in page route
- [ ] **2.0.2.27** Add error handling with shadcn/ui Alert in page route
- [ ] **2.0.2.28** Test merchant details page loads and displays data correctly

**2.0.3 Migrate JavaScript Components to React**

- [ ] **2.0.3.1** Read `cmd/frontend-service/static/js/components/merchant-context.js` to understand functionality
- [ ] **2.0.3.2** Create `frontend/hooks/useMerchantContext.ts` hook
- [ ] **2.0.3.3** Implement `useState` for merchant data in `useMerchantContext.ts`
- [ ] **2.0.3.4** Implement `useEffect` for loading merchant data in `useMerchantContext.ts`
- [ ] **2.0.3.5** Create `frontend/contexts/MerchantContext.tsx` with React Context
- [ ] **2.0.3.6** Create `frontend/providers/MerchantProvider.tsx` component
- [ ] **2.0.3.7** Read `cmd/frontend-service/static/js/components/session-manager.js` to understand functionality
- [ ] **2.0.3.8** Create `frontend/hooks/useSessionManager.ts` hook
- [ ] **2.0.3.9** Implement sessionStorage management in `useSessionManager.ts`
- [ ] **2.0.3.10** Read `cmd/frontend-service/static/js/components/risk-score-panel.js` to understand functionality
- [ ] **2.0.3.11** Create `frontend/components/risk/RiskScorePanel.tsx` component
- [ ] **2.0.3.12** Use shadcn/ui Card and Badge components in `RiskScorePanel.tsx`
- [ ] **2.0.3.13** Read `cmd/frontend-service/static/js/components/risk-visualization.js` to understand functionality
- [ ] **2.0.3.14** Create `frontend/components/risk/RiskVisualization.tsx` component
- [ ] **2.0.3.15** Integrate Chart.js in `RiskVisualization.tsx` (install `react-chartjs-2`)
- [ ] **2.0.3.16** Read `cmd/frontend-service/static/js/components/risk-indicators-ui-template.js` to understand functionality
- [ ] **2.0.3.17** Create `frontend/components/risk/RiskIndicators.tsx` component
- [ ] **2.0.3.18** Use shadcn/ui Badge and Alert components in `RiskIndicators.tsx`
- [ ] **2.0.3.19** Read `cmd/frontend-service/static/js/components/risk-history.js` to understand functionality
- [ ] **2.0.3.20** Create `frontend/components/risk/RiskHistory.tsx` component
- [ ] **2.0.3.21** Install shadcn/ui Table component: `npx shadcn-ui@latest add table`
- [ ] **2.0.3.22** Use shadcn/ui Table in `RiskHistory.tsx`
- [ ] **2.0.3.23** Read `cmd/frontend-service/static/js/components/data-enrichment.js` to understand functionality
- [ ] **2.0.3.24** Create `frontend/components/merchant/DataEnrichment.tsx` component
- [ ] **2.0.3.25** Use shadcn/ui Button and Dialog in `DataEnrichment.tsx`
- [ ] **2.0.3.26** Read `cmd/frontend-service/static/components/navigation.js` to understand functionality
- [ ] **2.0.3.27** Create `frontend/components/layout/Navigation.tsx` component
- [ ] **2.0.3.28** Read `cmd/frontend-service/static/js/components/export-button.js` to understand functionality
- [ ] **2.0.3.29** Create `frontend/components/common/ExportButton.tsx` component
- [ ] **2.0.3.30** Use shadcn/ui Button in `ExportButton.tsx`
- [ ] **2.0.3.31** Replace custom CSS classes with Tailwind CSS classes in all migrated components
- [ ] **2.0.3.32** Update responsive design using Tailwind responsive utilities
- [ ] **2.0.3.33** Test all migrated components render correctly
- [ ] **2.0.3.34** Verify all functionality preserved from original JavaScript components

**2.0.4 Update Go Service to Serve Next.js Build**

- [ ] **2.0.4.1** Read `cmd/frontend-service/main.go` to understand current static file serving
- [ ] **2.0.4.2** Create `cmd/frontend-service/build-frontend.sh` script (or update existing)
- [ ] **2.0.4.3** Add `cd frontend && npm install` step to build script
- [ ] **2.0.4.4** Add `cd frontend && npm run build` step to build script
- [ ] **2.0.4.5** Configure Next.js output to `cmd/frontend-service/static/` or separate `frontend-dist/` directory
- [ ] **2.0.4.6** Read `cmd/frontend-service/Dockerfile` to understand current build process
- [ ] **2.0.4.7** Add Node.js installation step to Dockerfile (if not present)
- [ ] **2.0.4.8** Add `WORKDIR /app/frontend` step to Dockerfile
- [ ] **2.0.4.9** Add `COPY frontend/package*.json ./` step to Dockerfile
- [ ] **2.0.4.10** Add `RUN npm install` step to Dockerfile
- [ ] **2.0.4.11** Add `COPY frontend/ ./` step to Dockerfile
- [ ] **2.0.4.12** Add `RUN npm run build` step to Dockerfile
- [ ] **2.0.4.13** Update Dockerfile to copy Next.js build output to static directory
- [ ] **2.0.4.14** Modify `cmd/frontend-service/main.go` to serve Next.js build output
- [ ] **2.0.4.15** Add catch-all route handler for Next.js client-side routing in `main.go`
- [ ] **2.0.4.16** Ensure API proxy functionality still works in `main.go`
- [ ] **2.0.4.17** Update health check endpoint if needed
- [ ] **2.0.4.18** Test Go service serves Next.js build correctly locally
- [ ] **2.0.4.19** Configure Next.js dev server proxy in `frontend/next.config.js` for development
- [ ] **2.0.4.20** Test development workflow: Next.js dev server with API proxy
- [ ] **2.0.4.21** Verify environment variables are passed correctly to Next.js build

**2.0.5 Testing and Validation**

- [ ] **2.0.5.1** Install testing dependencies: `npm install --save-dev @testing-library/react @testing-library/jest-dom jest jest-environment-jsdom`
- [ ] **2.0.5.2** Create `frontend/jest.config.js` configuration file
- [ ] **2.0.5.3** Create `frontend/.eslintrc.json` configuration file
- [ ] **2.0.5.4** Create `frontend/__tests__/components/merchant/MerchantDetailsLayout.test.tsx`
- [ ] **2.0.5.5** Write test for `MerchantDetailsLayout` component rendering
- [ ] **2.0.5.6** Write test for tab navigation in `MerchantDetailsLayout`
- [ ] **2.0.5.7** Create `frontend/__tests__/hooks/useMerchantContext.test.tsx`
- [ ] **2.0.5.8** Write test for `useMerchantContext` hook
- [ ] **2.0.5.9** Create `frontend/__tests__/components/risk/RiskScorePanel.test.tsx`
- [ ] **2.0.5.10** Write test for `RiskScorePanel` component
- [ ] **2.0.5.11** Create `frontend/__tests__/lib/api.test.ts`
- [ ] **2.0.5.12** Write tests for API client functions with mocked fetch
- [ ] **2.0.5.13** Run all tests: `npm test`
- [ ] **2.0.5.14** Fix any failing tests
- [ ] **2.0.5.15** Test full merchant details page flow manually
- [ ] **2.0.5.16** Test tab navigation works correctly
- [ ] **2.0.5.17** Test API calls and error handling
- [ ] **2.0.5.18** Test loading states display correctly
- [ ] **2.0.5.19** Compare new design with existing design visually
- [ ] **2.0.5.20** Test cross-browser compatibility (Chrome, Firefox, Safari, Edge)
- [ ] **2.0.5.21** Test responsive design on mobile devices
- [ ] **2.0.5.22** Verify all features work correctly in new React implementation
- [ ] **2.0.5.23** Document any differences or improvements in design

#### Task 2.1: Enhance Business Analytics Endpoints

**2.1.1 Enhance GetMerchantAnalytics with Parallel Fetching**

- [ ] **2.1.1.1** Read `internal/services/merchant_analytics_service.go` to understand current implementation
- [ ] **2.1.1.2** Add `sync` package import to `merchant_analytics_service.go`
- [ ] **2.1.1.3** Modify `GetMerchantAnalytics` method signature to accept timeout context
- [ ] **2.1.1.4** Add `context.WithTimeout(ctx, 30*time.Second)` at start of `GetMerchantAnalytics`
- [ ] **2.1.1.5** Create `sync.WaitGroup` variable for goroutine synchronization
- [ ] **2.1.1.6** Create `sync.Mutex` variable for thread-safe data access
- [ ] **2.1.1.7** Create variables for classification, security, quality, intelligence, verification data
- [ ] **2.1.1.8** Create `[]error` slice for collecting errors
- [ ] **2.1.1.9** Implement goroutine for fetching classification data with error handling
- [ ] **2.1.1.10** Implement goroutine for fetching security data with error handling
- [ ] **2.1.1.11** Implement goroutine for fetching quality metrics with error handling
- [ ] **2.1.1.12** Implement goroutine for fetching intelligence data with error handling
- [ ] **2.1.1.13** Implement goroutine for fetching verification data with error handling
- [ ] **2.1.1.14** Add `wg.Wait()` to wait for all goroutines to complete
- [ ] **2.1.1.15** Add error handling: return error if classification data is nil (critical)
- [ ] **2.1.1.16** Return partial data if some sources fail (non-critical errors)
- [ ] **2.1.1.17** Assemble `AnalyticsData` struct with all fetched data
- [ ] **2.1.1.18** Test parallel fetching with unit tests
- [ ] **2.1.1.19** Verify timeout works correctly (test with long-running operations)
- [ ] **2.1.1.20** Measure performance improvement vs sequential fetching

**2.1.2 Add Caching Layer**

- [ ] **2.1.2.1** Check if cache infrastructure exists in `pkg/cache/` directory
- [ ] **2.1.2.2** Read existing cache implementation to understand interface
- [ ] **2.1.2.3** Add cache field to `merchantAnalyticsService` struct
- [ ] **2.1.2.4** Update `NewMerchantAnalyticsService` constructor to accept cache parameter
- [ ] **2.1.2.5** Add cache check at start of `GetMerchantAnalytics` method
- [ ] **2.1.2.6** Generate cache key: `fmt.Sprintf("analytics:%s", merchantId)`
- [ ] **2.1.2.7** Check cache using `s.cache.Get(ctx, cacheKey)`
- [ ] **2.1.2.8** Unmarshal cached data if found and return early
- [ ] **2.1.2.9** Move existing data fetching logic to `fetchAnalyticsData` helper method
- [ ] **2.1.2.10** Call `fetchAnalyticsData` if cache miss
- [ ] **2.1.2.11** Marshal fetched data to JSON
- [ ] **2.1.2.12** Cache result with 5-minute TTL: `s.cache.Set(ctx, cacheKey, dataJSON, 5*time.Minute)`
- [ ] **2.1.2.13** Handle cache errors gracefully (log but don't fail request)
- [ ] **2.1.2.14** Write unit tests for cache hit scenario
- [ ] **2.1.2.15** Write unit tests for cache miss scenario
- [ ] **2.1.2.16** Write unit tests for cache expiration
- [ ] **2.1.2.17** Test cache performance improvement

**2.1.3 Enhance Error Handling**

- [ ] **2.1.3.1** Add merchant validation at start of `GetMerchantAnalytics`
- [ ] **2.1.3.2** Call `s.merchantRepo.GetByID(ctx, merchantId)` to verify merchant exists
- [ ] **2.1.3.3** Return `ErrMerchantNotFound` if merchant not found
- [ ] **2.1.3.4** Check `merchant.Status` field
- [ ] **2.1.3.5** Return `ErrMerchantInactive` if status is not "active"
- [ ] **2.1.3.6** Continue with data fetching if merchant is valid
- [ ] **2.1.3.7** Update handler to return appropriate HTTP status codes for errors
- [ ] **2.1.3.8** Write unit tests for merchant not found scenario
- [ ] **2.1.3.9** Write unit tests for inactive merchant scenario
- [ ] **2.1.3.10** Write unit tests for partial data return (some sources fail)

**2.1.4 Complete Website Analysis Endpoint**

- [ ] **2.1.4.1** Read `GetWebsiteAnalysis` implementation in `merchant_analytics_service.go`
- [ ] **2.1.4.2** Verify merchant website URL retrieval logic
- [ ] **2.1.4.3** Check if website analysis data exists in repository
- [ ] **2.1.4.4** Add logic to check if analysis data is older than 24 hours
- [ ] **2.1.4.5** Create `triggerAnalysis` method if analysis is stale or missing
- [ ] **2.1.4.6** Implement background job for website analysis (if needed)
- [ ] **2.1.4.7** Return cached analysis if fresh (< 24 hours)
- [ ] **2.1.4.8** Return new analysis if triggered
- [ ] **2.1.4.9** Write unit tests for website analysis endpoint
- [ ] **2.1.4.10** Test analysis triggering logic

#### Task 2.2: Complete Risk Assessment Endpoints

**2.2.1 Complete Background Job Processing**

- [ ] **2.2.1.1** Read `internal/jobs/risk_assessment_job.go` to understand current implementation
- [ ] **2.2.1.2** Enhance `Process` method in `risk_assessment_job.go`
- [ ] **2.2.1.3** Add merchant data retrieval: `j.merchantRepo.GetByID(ctx, j.MerchantID)`
- [ ] **2.2.1.4** Add error handling for merchant retrieval
- [ ] **2.2.1.5** Update assessment status to "processing": `j.assessmentRepo.UpdateStatus(ctx, j.AssessmentID, "processing")`
- [ ] **2.2.1.6** Call risk engine: `j.riskEngine.Assess(ctx, merchant)`
- [ ] **2.2.1.7** Handle risk engine errors
- [ ] **2.2.1.8** Update status to "failed" if error occurs: `j.assessmentRepo.UpdateStatus(ctx, j.AssessmentID, "failed")`
- [ ] **2.2.1.9** Set assessment ID on result: `assessment.ID = j.AssessmentID`
- [ ] **2.2.1.10** Save assessment results: `j.assessmentRepo.SaveResults(ctx, assessment)`
- [ ] **2.2.1.11** Update status to "completed": `j.assessmentRepo.UpdateStatus(ctx, j.AssessmentID, "completed")`
- [ ] **2.2.1.12** Add logging for each step
- [ ] **2.2.1.13** Write unit tests for job processing
- [ ] **2.2.1.14** Test error handling scenarios

**2.2.2 Implement GET /api/v1/risk/history/{merchantId}**

- [ ] **2.2.2.1** Add `GetRiskHistory` method to `RiskAssessmentService` interface
- [ ] **2.2.2.2** Implement `GetRiskHistory(ctx, merchantId, limit, offset)` in service
- [ ] **2.2.2.3** Add repository method `GetHistoryByMerchantID(ctx, merchantId, limit, offset)`
- [ ] **2.2.2.4** Create handler method `GetRiskHistory` in `AsyncRiskAssessmentHandler`
- [ ] **2.2.2.5** Extract `merchantId` from path using `extractMerchantID` helper
- [ ] **2.2.2.6** Parse `limit` query parameter (default: 10)
- [ ] **2.2.2.7** Parse `offset` query parameter (default: 0)
- [ ] **2.2.2.8** Call service method `GetRiskHistory`
- [ ] **2.2.2.9** Handle errors and return appropriate HTTP status
- [ ] **2.2.2.10** Encode response as JSON
- [ ] **2.2.2.11** Register route in `internal/api/routes/risk_routes.go`: `GET /api/v1/risk/history/{merchantId}`
- [ ] **2.2.2.12** Add middleware (auth, rate limiting) to route
- [ ] **2.2.2.13** Write unit tests for handler
- [ ] **2.2.2.14** Write integration tests for endpoint
- [ ] **2.2.2.15** Test pagination with different limit/offset values

**2.2.3 Implement GET /api/v1/risk/predictions/{merchantId}**

- [ ] **2.2.3.1** Add `GetPredictions` method to `RiskAssessmentService` interface
- [ ] **2.2.3.2** Implement `GetPredictions(ctx, merchantId, options)` in service
- [ ] **2.2.3.3** Create handler method `GetRiskPredictions` in handler
- [ ] **2.2.3.4** Extract `merchantId` from path
- [ ] **2.2.3.5** Parse query parameters: `horizons`, `includeScenarios`, `includeConfidence`
- [ ] **2.2.3.6** Create options struct from query parameters
- [ ] **2.2.3.7** Call service method with options
- [ ] **2.2.3.8** Handle errors appropriately
- [ ] **2.2.3.9** Encode response as JSON
- [ ] **2.2.3.10** Register route: `GET /api/v1/risk/predictions/{merchantId}`
- [ ] **2.2.3.11** Add middleware to route
- [ ] **2.2.3.12** Write unit tests
- [ ] **2.2.3.13** Write integration tests
- [ ] **2.2.3.14** Test with different query parameter combinations

**2.2.4 Implement GET /api/v1/risk/explain/{assessmentId}**

- [ ] **2.2.4.1** Add `ExplainAssessment` method to service interface
- [ ] **2.2.4.2** Implement `ExplainAssessment(ctx, assessmentId)` in service
- [ ] **2.2.4.3** Create handler method `ExplainRiskAssessment` in handler
- [ ] **2.2.4.4** Extract `assessmentId` from path
- [ ] **2.2.4.5** Call service method
- [ ] **2.2.4.6** Handle errors (assessment not found, etc.)
- [ ] **2.2.4.7** Encode explanation data as JSON
- [ ] **2.2.4.8** Register route: `GET /api/v1/risk/explain/{assessmentId}`
- [ ] **2.2.4.9** Add middleware
- [ ] **2.2.4.10** Write unit tests
- [ ] **2.2.4.11** Write integration tests

**2.2.5 Implement GET /api/v1/merchants/{merchantId}/risk-recommendations**

- [ ] **2.2.5.1** Add `GetRecommendations` method to service interface
- [ ] **2.2.5.2** Implement `GetRecommendations(ctx, merchantId)` in service
- [ ] **2.2.5.3** Create handler method `GetRiskRecommendations` in handler
- [ ] **2.2.5.4** Extract `merchantId` from path
- [ ] **2.2.5.5** Call service method
- [ ] **2.2.5.6** Handle errors
- [ ] **2.2.5.7** Encode recommendations as JSON
- [ ] **2.2.5.8** Register route in `internal/api/routes/merchant_routes.go`: `GET /api/v1/merchants/{merchantId}/risk-recommendations`
- [ ] **2.2.5.9** Add middleware
- [ ] **2.2.5.10** Write unit tests
- [ ] **2.2.5.11** Write integration tests

#### Task 2.3: Implement Risk Indicators Endpoints

**2.3.1 Create Risk Indicators Service**

- [ ] **2.3.1.1** Create `internal/services/risk_indicators_service.go` file
- [ ] **2.3.1.2** Define `RiskIndicatorsService` interface with `GetRiskIndicators` method
- [ ] **2.3.1.3** Create `riskIndicatorsService` struct with repository field
- [ ] **2.3.1.4** Implement `NewRiskIndicatorsService` constructor
- [ ] **2.3.1.5** Implement `GetRiskIndicators(ctx, merchantId)` method
- [ ] **2.3.1.6** Call repository to get indicators: `s.indicatorsRepo.GetByMerchantID(ctx, merchantId)`
- [ ] **2.3.1.7** Implement `calculateOverallScore` helper method
- [ ] **2.3.1.8** Calculate overall score from indicators
- [ ] **2.3.1.9** Assemble `RiskIndicatorsData` struct
- [ ] **2.3.1.10** Return data with timestamp
- [ ] **2.3.1.11** Write unit tests for service

**2.3.2 Create Risk Indicators Handler**

- [ ] **2.3.2.1** Create `internal/api/handlers/risk_indicators_handler.go` file
- [ ] **2.3.2.2** Create `RiskIndicatorsHandler` struct with service field
- [ ] **2.3.2.3** Implement `NewRiskIndicatorsHandler` constructor
- [ ] **2.3.2.4** Implement `GetRiskIndicators` handler method
- [ ] **2.3.2.5** Extract `merchantId` from path
- [ ] **2.3.2.6** Call service method
- [ ] **2.3.2.7** Handle errors
- [ ] **2.3.2.8** Encode response as JSON
- [ ] **2.3.2.9** Implement `GetRiskAlerts` handler method
- [ ] **2.3.2.10** Parse `severity` query parameter
- [ ] **2.3.2.11** Parse `status` query parameter
- [ ] **2.3.2.12** Call service method with filters
- [ ] **2.3.2.13** Register routes in `internal/api/routes/risk_routes.go`
- [ ] **2.3.2.14** Add middleware to routes
- [ ] **2.3.2.15** Write unit tests for handlers

**2.3.3 Create Data Models**

- [ ] **2.3.3.1** Create `internal/models/risk_indicators.go` file
- [ ] **2.3.3.2** Define `RiskIndicatorsData` struct with JSON tags
- [ ] **2.3.3.3** Define `RiskIndicator` struct with all fields (ID, Type, Name, Severity, Status, Description, DetectedAt, Score)
- [ ] **2.3.3.4** Add proper JSON tags to all fields
- [ ] **2.3.3.5** Add validation tags if needed

**2.3.4 Create Repository**

- [ ] **2.3.4.1** Create `internal/database/risk_indicators_repository.go` file
- [ ] **2.3.4.2** Create `RiskIndicatorsRepository` struct with database connection
- [ ] **2.3.4.3** Implement `NewRiskIndicatorsRepository` constructor
- [ ] **2.3.4.4** Implement `GetByMerchantID(ctx, merchantId)` method
- [ ] **2.3.4.5** Write SQL query to fetch indicators from database
- [ ] **2.3.4.6** Add support for filtering by severity
- [ ] **2.3.4.7** Add support for filtering by status
- [ ] **2.3.4.8** Scan results into `RiskIndicator` structs
- [ ] **2.3.4.9** Handle database errors
- [ ] **2.3.4.10** Add database indexes for performance (if needed)
- [ ] **2.3.4.11** Write unit tests for repository

#### Task 2.4: Complete Data Enrichment Integration

**2.4.1 Backend: Complete Enrichment Service**

- [ ] **2.4.1.1** Read `internal/services/data_enrichment_service.go` if exists, or create new file
- [ ] **2.4.1.2** Define `DataEnrichmentService` interface
- [ ] **2.4.1.3** Implement `TriggerEnrichment(ctx, merchantId, source)` method
- [ ] **2.4.1.4** Implement `isValidSource` helper method
- [ ] **2.4.1.5** Validate source parameter
- [ ] **2.4.1.6** Return `ErrInvalidSource` if invalid
- [ ] **2.4.1.7** Generate job ID using `generateJobID()` helper
- [ ] **2.4.1.8** Create `EnrichmentJob` struct with all fields
- [ ] **2.4.1.9** Save job to repository: `s.jobRepo.Create(ctx, job)`
- [ ] **2.4.1.10** Queue job: `s.jobQueue.Enqueue(ctx, job)`
- [ ] **2.4.1.11** Implement `GetEnrichmentSources(ctx)` method
- [ ] **2.4.1.12** Call repository: `s.sourcesRepo.GetAll(ctx)`
- [ ] **2.4.1.13** Write unit tests for service

**2.4.2 Backend: Create Handler**

- [ ] **2.4.2.1** Create `internal/api/handlers/data_enrichment_handler.go` file
- [ ] **2.4.2.2** Create handler struct with service field
- [ ] **2.4.2.3** Implement `TriggerEnrichment` handler method
- [ ] **2.4.2.4** Parse request body for `source` parameter
- [ ] **2.4.2.5** Extract `merchantId` from path
- [ ] **2.4.2.6** Call service method
- [ ] **2.4.2.7** Handle errors
- [ ] **2.4.2.8** Encode job response as JSON
- [ ] **2.4.2.9** Implement `GetEnrichmentSources` handler method
- [ ] **2.4.2.10** Call service method
- [ ] **2.4.2.11** Encode sources as JSON
- [ ] **2.4.2.12** Register routes in `internal/api/routes/merchant_routes.go`
- [ ] **2.4.2.13** Add middleware
- [ ] **2.4.2.14** Write unit tests

**2.4.3 Frontend: Enhance Data Enrichment Component**

- [ ] **2.4.3.1** Read `cmd/frontend-service/static/js/components/data-enrichment.js` to understand current implementation
- [ ] **2.4.3.2** If migrating to React, create `frontend/components/merchant/DataEnrichment.tsx` (already done in 2.0.3.24)
- [ ] **2.4.3.3** Implement `triggerEnrichment` method in component
- [ ] **2.4.3.4** Make POST request to `/api/v1/merchants/${merchantId}/enrichment/trigger`
- [ ] **2.4.3.5** Include `source` in request body
- [ ] **2.4.3.6** Add error handling
- [ ] **2.4.3.7** Implement `getEnrichmentSources` method
- [ ] **2.4.3.8** Make GET request to `/api/v1/merchants/${merchantId}/enrichment/sources`
- [ ] **2.4.3.9** Add loading states using shadcn/ui Skeleton
- [ ] **2.4.3.10** Add success/error feedback using shadcn/ui Toast
- [ ] **2.4.3.11** Integrate component into merchant details page
- [ ] **2.4.3.12** Test component functionality

#### Task 2.5: Complete External Data Sources Integration

- [ ] **2.5.1** Create `internal/services/external_data_sources_service.go` (similar structure to data enrichment)
- [ ] **2.5.2** Create `internal/api/handlers/external_data_sources_handler.go`
- [ ] **2.5.3** Create `frontend/components/merchant/ExternalDataSources.tsx` (or enhance existing)
- [ ] **2.5.4** Register routes
- [ ] **2.5.5** Write tests
- [ ] **2.5.6** Integrate with merchant details page

#### Task 2.6: Implement Consistent Error Handling

**2.6.1 Backend: Standardized Error Types**

- [ ] **2.6.1.1** Create `pkg/errors/api_errors.go` file
- [ ] **2.6.1.2** Define `APIError` struct with Code, Message, Details fields
- [ ] **2.6.1.3** Implement `Error()` method for `APIError`
- [ ] **2.6.1.4** Create `WriteError` helper function
- [ ] **2.6.1.5** Implement `getErrorCode` helper function to map errors to codes
- [ ] **2.6.1.6** Create error code constants (e.g., `ErrCodeNotFound`, `ErrCodeValidation`, etc.)
- [ ] **2.6.1.7** Implement `RetryWithBackoff` function with exponential backoff
- [ ] **2.6.1.8** Add context cancellation support to retry function
- [ ] **2.6.1.9** Write unit tests for error handling utilities

**2.6.2 Backend: Enhance Error Middleware**

- [ ] **2.6.2.1** Read `internal/api/middleware/error_handling.go` to understand current implementation
- [ ] **2.6.2.2** Update error middleware to use `APIError` struct
- [ ] **2.6.2.3** Map internal errors to HTTP status codes
- [ ] **2.6.2.4** Add error logging with context
- [ ] **2.6.2.5** Ensure consistent error response format
- [ ] **2.6.2.6** Test error middleware with various error types

**2.6.3 Frontend: Error Handling Utility**

- [ ] **2.6.3.1** Create `frontend/lib/error-handler.ts` file (or `cmd/frontend-service/static/js/utils/error-handler.js` if not migrating)
- [ ] **2.6.3.2** Create `ErrorHandler` class
- [ ] **2.6.3.3** Implement `handleAPIError` static method
- [ ] **2.6.3.4** Parse error response structure
- [ ] **2.6.3.5** Extract error message and code
- [ ] **2.6.3.6** Implement `showErrorNotification` method
- [ ] **2.6.3.7** Create toast notification for errors (use shadcn/ui Toast if React)
- [ ] **2.6.3.8** Implement `logError` method
- [ ] **2.6.3.9** Send errors to logging service (if available)
- [ ] **2.6.3.10** Integrate error handler into API client
- [ ] **2.6.3.11** Test error handling with various error scenarios

### Week 3: Performance Optimization

#### Task 3.1: API Request Optimization

**3.1.1 Implement Request Batching**

- [ ] **3.1.1.1** Create `frontend/lib/api-batcher.ts` (or `cmd/frontend-service/static/js/utils/api-batcher.js`)
- [ ] **3.1.1.2** Create `APIBatcher` class
- [ ] **3.1.1.3** Initialize `pendingRequests` Map in constructor
- [ ] **3.1.1.4** Set `batchTimeout` to 100ms
- [ ] **3.1.1.5** Implement `batchRequest(key, requestFn)` method
- [ ] **3.1.1.6** Check if request already pending, return existing promise
- [ ] **3.1.1.7** Create new promise and store in Map
- [ ] **3.1.1.8** Clean up promise from Map when done
- [ ] **3.1.1.9** Implement `executeRequest` method with debounce
- [ ] **3.1.1.10** Integrate batcher into API client
- [ ] **3.1.1.11** Write unit tests for batcher
- [ ] **3.1.1.12** Measure performance improvement

**3.1.2 Implement Response Caching**

- [ ] **3.1.2.1** Create `frontend/lib/api-cache.ts` (or `cmd/frontend-service/static/js/utils/api-cache.js`)
- [ ] **3.1.2.2** Create `APICache` class
- [ ] **3.1.2.3** Initialize cache Map in constructor
- [ ] **3.1.2.4** Set default TTL to 5 minutes
- [ ] **3.1.2.5** Implement `get(key)` method
- [ ] **3.1.2.6** Check if cached data exists
- [ ] **3.1.2.7** Check if cached data is expired
- [ ] **3.1.2.8** Delete expired entries
- [ ] **3.1.2.9** Implement `set(key, data, ttl)` method
- [ ] **3.1.2.10** Store data with expiration timestamp
- [ ] **3.1.2.11** Implement `clear()` method
- [ ] **3.1.2.12** Implement `persist(key)` method for sessionStorage
- [ ] **3.1.2.13** Implement `restore(key)` method from sessionStorage
- [ ] **3.1.2.14** Create `cachedFetch` wrapper function
- [ ] **3.1.2.15** Generate cache key from URL and options
- [ ] **3.1.2.16** Check cache before making request
- [ ] **3.1.2.17** Cache response after fetch
- [ ] **3.1.2.18** Integrate into API client
- [ ] **3.1.2.19** Write unit tests
- [ ] **3.1.2.20** Test cache hit/miss scenarios

**3.1.3 Implement Request Deduplication**

- [ ] **3.1.3.1** Create `frontend/lib/request-deduplicator.ts` (or `cmd/frontend-service/static/js/utils/request-deduplicator.js`)
- [ ] **3.1.3.2** Create `RequestDeduplicator` class
- [ ] **3.1.3.3** Initialize `pendingRequests` Map
- [ ] **3.1.3.4** Implement `deduplicate(key, requestFn)` method
- [ ] **3.1.3.5** Check if request already pending, return existing promise
- [ ] **3.1.3.6** Create new request promise
- [ ] **3.1.3.7** Remove from Map when done
- [ ] **3.1.3.8** Integrate into API client
- [ ] **3.1.3.9** Write unit tests
- [ ] **3.1.3.10** Test deduplication with concurrent requests

#### Task 3.2: Lazy Loading Enhancements

**3.2.1 Implement Lazy Loader**

- [ ] **3.2.1.1** Create `frontend/lib/lazy-loader.ts` (or `cmd/frontend-service/static/js/utils/lazy-loader.js`)
- [ ] **3.2.1.2** Create `LazyLoader` class
- [ ] **3.2.1.3** Initialize IntersectionObserver with 50px rootMargin
- [ ] **3.2.1.4** Create `loadedSections` Set to track loaded sections
- [ ] **3.2.1.5** Implement `observe(sectionElement, loadFn)` method
- [ ] **3.2.1.6** Check if section already loaded
- [ ] **3.2.1.7** Observe element with IntersectionObserver
- [ ] **3.2.1.8** Store load function in element dataset
- [ ] **3.2.1.9** Implement `handleIntersection` callback
- [ ] **3.2.1.10** Execute load function when element intersects
- [ ] **3.2.1.11** Mark section as loaded
- [ ] **3.2.1.12** Unobserve element after loading
- [ ] **3.2.1.13** Integrate lazy loader into merchant details page
- [ ] **3.2.1.14** Apply to expandable sections
- [ ] **3.2.1.15** Write unit tests
- [ ] **3.2.1.16** Test lazy loading behavior

**3.2.2 Defer Non-Critical API Calls**

- [ ] **3.2.2.1** Identify non-critical API calls (analytics, recommendations, external sources)
- [ ] **3.2.2.2** Create `loadNonCriticalData` function
- [ ] **3.2.2.3** Add window load event listener
- [ ] **3.2.2.4** Check for `requestIdleCallback` support
- [ ] **3.2.2.5** Use `requestIdleCallback` if available
- [ ] **3.2.2.6** Use `setTimeout` fallback (2000ms delay)
- [ ] **3.2.2.7** Load analytics data in `loadNonCriticalData`
- [ ] **3.2.2.8** Load recommendations in `loadNonCriticalData`
- [ ] **3.2.2.9** Load external sources in `loadNonCriticalData`
- [ ] **3.2.2.10** Test deferred loading
- [ ] **3.2.2.11** Measure performance improvement

#### Task 3.3: Loading State Improvements

**3.3.1 Implement Skeleton Loaders**

- [ ] **3.3.1.1** Create `frontend/components/ui/skeleton-loader.tsx` (or use shadcn/ui Skeleton if React)
- [ ] **3.3.1.2** Create `SkeletonLoader` class/component
- [ ] **3.3.1.3** Implement `createCardSkeleton` method
- [ ] **3.3.1.4** Create skeleton HTML structure with lines
- [ ] **3.3.1.5** Implement `show(element)` method
- [ ] **3.3.1.6** Implement `hide(element)` method
- [ ] **3.3.1.7** Create `frontend/styles/skeleton-loaders.css` (or add to existing CSS)
- [ ] **3.3.1.8** Add skeleton loading animation CSS
- [ ] **3.3.1.9** Add skeleton line styles
- [ ] **3.3.1.10** Add pulse animation keyframes
- [ ] **3.3.1.11** Integrate skeleton loaders into merchant details page
- [ ] **3.3.1.12** Replace loading spinners with skeletons
- [ ] **3.3.1.13** Test skeleton display

**3.3.2 Implement Progress Indicators**

- [ ] **3.3.2.1** Create `frontend/components/ui/progress-indicator.tsx` (or `cmd/frontend-service/static/js/components/progress-indicator.js`)
- [ ] **3.3.2.2** Create `ProgressIndicator` class/component
- [ ] **3.3.2.3** Implement `show()` method
- [ ] **3.3.2.4** Implement `hide()` method
- [ ] **3.3.2.5** Implement `update(percentage)` method
- [ ] **3.3.2.6** Update progress bar width
- [ ] **3.3.2.7** Update progress text
- [ ] **3.3.2.8** Implement `estimateTimeRemaining(completed, total)` method
- [ ] **3.3.2.9** Calculate remaining time based on rate
- [ ] **3.3.2.10** Integrate progress indicators into long operations
- [ ] **3.3.2.11** Test progress updates

#### Task 3.4: Empty State Design

**3.4.1 Create Empty State Component**

- [ ] **3.4.1.1** Create `frontend/components/ui/empty-state.tsx` (or `cmd/frontend-service/static/js/components/empty-state.js`)
- [ ] **3.4.1.2** Create `EmptyState` class/component
- [ ] **3.4.1.3** Implement `create(type, options)` method
- [ ] **3.4.1.4** Define templates for `noData`, `error`, `noResults` types
- [ ] **3.4.1.5** Include icon, title, message, and optional action in templates
- [ ] **3.4.1.6** Implement `show(element, type, options)` method
- [ ] **3.4.1.7** Create `frontend/styles/empty-states.css` (or add to existing)
- [ ] **3.4.1.8** Style empty state container
- [ ] **3.4.1.9** Style empty state icon
- [ ] **3.4.1.10** Style empty state title and message
- [ ] **3.4.1.11** Style action button
- [ ] **3.4.1.12** Apply empty states to Overview tab
- [ ] **3.4.1.13** Apply empty states to Analytics tab
- [ ] **3.4.1.14** Apply empty states to Risk tab
- [ ] **3.4.1.15** Add retry functionality for error states
- [ ] **3.4.1.16** Test empty state display

#### Task 3.5: Success Feedback Implementation

**3.5.1 Implement Toast Notifications**

- [ ] **3.5.1.1** If React: shadcn/ui Toast already installed (from 2.0.2.7)
- [ ] **3.5.1.2** If vanilla JS: Create `cmd/frontend-service/static/js/components/toast-notification.js`
- [ ] **3.5.1.3** Create `ToastNotification` class
- [ ] **3.5.1.4** Implement `show(message, type, duration)` method
- [ ] **3.5.1.5** Create toast element with appropriate classes
- [ ] **3.5.1.6** Add animation classes for show/hide
- [ ] **3.5.1.7** Auto-dismiss after duration
- [ ] **3.5.1.8** Implement `success(message)` static method
- [ ] **3.5.1.9** Implement `error(message)` static method (5s duration)
- [ ] **3.5.1.10** Implement `info(message)` static method
- [ ] **3.5.1.11** Create `frontend/styles/toast-notifications.css` (or add to existing)
- [ ] **3.5.1.12** Style toast container
- [ ] **3.5.1.13** Add animations for slide in/out
- [ ] **3.5.1.14** Style different toast types (success, error, info)
- [ ] **3.5.1.15** Integrate toasts into API client for success/error responses
- [ ] **3.5.1.16** Integrate toasts into form submissions
- [ ] **3.5.1.17** Test toast display and dismissal

### Week 4: Quality Assurance

#### Task 4.1: Execute Comprehensive Test Suites

**4.1.1 Execute Navigation Testing**

- [ ] **4.1.1.1** Read `cmd/frontend-service/static/docs/navigation-testing-guide.md`
- [ ] **4.1.1.2** Execute all 50+ test cases from guide
- [ ] **4.1.1.3** Test form submission flow
- [ ] **4.1.1.4** Test data persistence
- [ ] **4.1.1.5** Test tab navigation
- [ ] **4.1.1.6** Test error handling
- [ ] **4.1.1.7** Test performance
- [ ] **4.1.1.8** Test accessibility
- [ ] **4.1.1.9** Document test results
- [ ] **4.1.1.10** Report issues found

**4.1.2 Execute Export Testing**

- [ ] **4.1.2.1** Read `cmd/frontend-service/static/docs/export-functionality-testing-guide.md`
- [ ] **4.1.2.2** Execute all 60+ test cases from guide
- [ ] **4.1.2.3** Test CSV export format
- [ ] **4.1.2.4** Test PDF export format
- [ ] **4.1.2.5** Test JSON export format
- [ ] **4.1.2.6** Test Excel export format
- [ ] **4.1.2.7** Test export from all tabs
- [ ] **4.1.2.8** Document test results
- [ ] **4.1.2.9** Report issues found

**4.1.3 Execute Cross-Browser Testing**

- [ ] **4.1.3.1** Read `cmd/frontend-service/static/docs/cross-browser-testing-guide.md`
- [ ] **4.1.3.2** Execute all 80+ test cases from guide
- [ ] **4.1.3.3** Test in Chrome browser
- [ ] **4.1.3.4** Test in Firefox browser
- [ ] **4.1.3.5** Test in Safari browser
- [ ] **4.1.3.6** Test in Edge browser
- [ ] **4.1.3.7** Document browser-specific issues
- [ ] **4.1.3.8** Verify fixes for browser issues
- [ ] **4.1.3.9** Create test report

#### Task 4.2: Fix Identified Issues

**4.2.1 Prioritize Issues**

- [ ] **4.2.1.1** Review all issues from testing
- [ ] **4.2.1.2** Categorize issues: Critical, High, Medium, Low
- [ ] **4.2.1.3** Create issue tracking document/spreadsheet
- [ ] **4.2.1.4** Assign issues to developers
- [ ] **4.2.1.5** Set deadlines for each priority level

**4.2.2 Fix Critical Issues**

- [ ] **4.2.2.1** Fix all critical issues
- [ ] **4.2.2.2** Verify fixes with testing
- [ ] **4.2.2.3** Update issue tracking

**4.2.3 Fix High-Priority Issues**

- [ ] **4.2.3.1** Fix all high-priority issues
- [ ] **4.2.3.2** Verify fixes with testing
- [ ] **4.2.3.3** Update issue tracking

**4.2.4 Address Medium-Priority Issues**

- [ ] **4.2.4.1** Fix medium-priority issues (time permitting)
- [ ] **4.2.4.2** Verify fixes
- [ ] **4.2.4.3** Update issue tracking

#### Task 4.3: Prepare for Beta Release

**4.3.1 Create Beta Release Checklist**

- [ ] **4.3.1.1** Verify all critical bugs fixed
- [ ] **4.3.1.2** Verify all high-priority features implemented
- [ ] **4.3.1.3** Check test coverage meets targets (80%+ unit, 100% critical paths)
- [ ] **4.3.1.4** Verify performance metrics meet targets
- [ ] **4.3.1.5** Verify cross-browser compatibility
- [ ] **4.3.1.6** Verify accessibility compliance (WCAG 2.1 AA)
- [ ] **4.3.1.7** Verify documentation complete
- [ ] **4.3.1.8** Verify release notes prepared
- [ ] **4.3.1.9** Verify beta tester communication prepared

**4.3.2 Prepare Release Documentation**

- [ ] **4.3.2.1** Create `docs/release-notes/beta-release-notes.md`
- [ ] **4.3.2.2** List all new features
- [ ] **4.3.2.3** List all bug fixes
- [ ] **4.3.2.4** List known issues
- [ ] **4.3.2.5** Add upgrade instructions
- [ ] **4.3.2.6** Create `docs/beta-tester-guide.md`
- [ ] **4.3.2.7** Add beta access instructions
- [ ] **4.3.2.8** Add what to test section
- [ ] **4.3.2.9** Add issue reporting process
- [ ] **4.3.2.10** Add feedback collection process
- [ ] **4.3.2.11** Update `docs/technical-documentation/beta-deployment-guide.md`
- [ ] **4.3.2.12** Update API documentation
- [ ] **4.3.2.13** Add deployment guide
- [ ] **4.3.2.14** Add rollback procedure

**4.3.3 Prepare Beta Tester Communication**

- [ ] **4.3.3.1** Create beta access instructions document
- [ ] **4.3.3.2** Prepare feedback collection process
- [ ] **4.3.3.3** Set up issue tracking system (if not already)
- [ ] **4.3.3.4** Create beta tester onboarding email template
- [ ] **4.3.3.5** Prepare beta testing timeline

---

## Dependencies

### External Dependencies

- Backend team availability
- Database access
- Testing tools access
- CI/CD platform

### Internal Dependencies

- Week 1 tasks completed
- API documentation available
- Test data prepared

---

## Risks and Mitigations

### Risk 1: API Endpoint Delays

**Mitigation:**

- Prioritize critical endpoints
- Use mock data for frontend development
- Adjust timeline if needed

### Risk 2: Performance Issues

**Mitigation:**

- Profile early and often
- Implement optimizations incrementally
- Test with realistic data volumes

### Risk 3: Testing Coverage Gaps

**Mitigation:**

- Follow comprehensive testing guides
- Use automated testing where possible
- Conduct peer reviews

---

## Next Steps

After completing Weeks 2-4, proceed to:

- **Plan 3: Long-Term Actions (Months 2-3)** - Implement advanced features and continuous improvement