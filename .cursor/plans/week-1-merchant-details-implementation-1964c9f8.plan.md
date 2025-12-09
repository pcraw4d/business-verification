---
name: Week 1 Merchant Details Implementation Plan
overview: ""
todos:
  - id: 4136c5fe-91ac-4b00-be9a-ed2bd6943610
    content: Set up API testing environment with Postman/Insomnia collections, environment variables, and test scripts
    status: pending
  - id: d1d0c5a2-0e9c-4428-ba8e-8f6dd0ca9cb6
    content: Create OpenAPI 3.0 specification for all merchant-details endpoints with request/response schemas and examples
    status: pending
  - id: fb9fb3d4-d7e4-40cd-9c82-dbc8974ef148
    content: Create merchant analytics repository with methods to fetch classification, security, and quality data from database
    status: pending
  - id: 78facea2-b53d-4797-9c61-3fd4429fc746
    content: Implement MerchantAnalyticsService to aggregate data from multiple sources and return AnalyticsData
    status: pending
  - id: 74a04688-bf4f-4298-8344-5d0988d33d26
    content: Create MerchantAnalyticsHandler with GetMerchantAnalytics and GetWebsiteAnalysis endpoints
    status: pending
  - id: eaf7c328-c9f5-4650-a124-0efc144f647d
    content: Register analytics routes in merchant_routes.go with proper middleware (auth, rate limiting)
    status: pending
  - id: 80fb45c9-3e6a-45a8-88f6-6d5c67015a34
    content: Create risk assessment repository for storing and retrieving assessment records
    status: pending
  - id: 9e612f13-641b-4071-b154-984cc269cc7e
    content: Implement background job system for processing risk assessments asynchronously
    status: pending
  - id: 24395c13-1b37-4959-b235-56a1ba2089ad
    content: Enhance risk assessment service with StartAssessment, GetAssessmentStatus, and ProcessAssessment methods
    status: pending
  - id: a9a62f4f-b493-4192-8661-c0c2b8391c5c
    content: Enhance risk assessment handler to implement async pattern (202 Accepted) and status checking endpoint
    status: pending
  - id: 130d58aa-d8e8-46d4-900e-1924cbae8150
    content: Create test data fixtures (JSON, SQL) for merchants, analytics, and risk assessments with various scenarios
    status: pending
  - id: 2dd6fadd-ae08-4679-83a7-63775e84de9d
    content: Create E2E tests for merchant-details page navigation, tab switching, data loading, and error handling
    status: pending
  - id: bf5f12ce-7aaa-4d5b-86d6-53a7ba84c030
    content: Create API integration tests for analytics and risk assessment endpoints with various test scenarios
    status: pending
  - id: 8261f440-9566-463d-a3ee-a8dc03e15bcc
    content: Enhance CI/CD pipeline to run merchant-details tests on PRs and report results
    status: pending
---

# Week 1 Merchant Details Implementation Plan

## Overview

Implement all technical tasks from Plan 1 Week 1, focusing on backend API development, API testing infrastructure, documentation, and testing enhancements for the merchant-details page.

## Task 2: Backend API Development

### 2.1 Set Up API Testing Environment

**Files to create:**

- `tests/api/merchant-details/postman-collection.json` - Postman collection for merchant-details endpoints
- `tests/api/merchant-details/insomnia-collection.json` - Insomnia collection (alternative)
- `tests/api/merchant-details/README.md` - Testing environment documentation
- `tests/api/merchant-details/.env.example` - Environment variables template

**Implementation:**

- Create Postman collection with all merchant-details endpoints organized by category
- Configure authentication (Bearer token) in collection
- Set up environment variables for dev/staging/prod
- Document token retrieval process
- Create test scripts for manual API testing

### 2.2 Document API Response Structures

**Files to create:**

- `api/openapi/merchant-details-api-spec.yaml` - OpenAPI 3.0 specification
- `docs/api/merchant-details-api-reference.md` - Human-readable API reference

**Endpoints to document:**

1. `GET /api/v1/merchants/{merchantId}/analytics`
2. `GET /api/v1/merchants/{merchantId}/website-analysis`
3. `POST /api/v1/risk/assess`
4. `GET /api/v1/merchants/{merchantId}/risk-score` (existing, needs documentation)
5. `GET /api/v1/merchants/{merchantId}/website-risk` (existing, needs documentation)

**Implementation:**

- Use OpenAPI 3.0 format with complete request/response schemas
- Include example requests and responses
- Document error response formats
- Add authentication requirements
- Include query parameters and path variables

### 2.3 Implement High-Priority Endpoints

#### 2.3.1 Business Analytics Endpoints

**File: `internal/api/handlers/merchant_analytics_handler.go` (new)**

- Create new handler for merchant-specific analytics
- Implement `GetMerchantAnalytics` handler for `GET /api/v1/merchants/{merchantId}/analytics`
- Implement `GetWebsiteAnalysis` handler for `GET /api/v1/merchants/{merchantId}/website-analysis`

**File: `internal/services/merchant_analytics_service.go` (new)**

- Create `MerchantAnalyticsService` interface
- Implement service with methods:
- `GetMerchantAnalytics(ctx, merchantId) (*AnalyticsData, error)`
- `GetWebsiteAnalysis(ctx, merchantId) (*WebsiteAnalysisData, error)`
- Aggregate data from multiple sources:
- Classification data from `business_classifications` table
- Security data (SSL, headers) - may need new table or external service
- Quality metrics from merchant data
- Intelligence data (if available)

**File: `internal/database/merchant_analytics_repository.go` (new)**

- Create repository for analytics data access
- Methods:
- `GetClassificationByMerchantID(ctx, merchantId) (*ClassificationData, error)`
- `GetSecurityDataByMerchantID(ctx, merchantId) (*SecurityData, error)`
- `GetQualityMetricsByMerchantID(ctx, merchantId) (*QualityData, error)`
- Use existing `MerchantPortfolioRepository` patterns

**File: `internal/api/routes/merchant_routes.go` (modify)**

- Add routes:
- `GET /api/v1/merchants/{merchantId}/analytics` → `MerchantAnalyticsHandler.GetMerchantAnalytics`
- `GET /api/v1/merchants/{merchantId}/website-analysis` → `MerchantAnalyticsHandler.GetWebsiteAnalysis`
- Ensure routes use existing middleware (auth, rate limiting)

**Data Models (new files or add to existing):**

- `internal/models/merchant_analytics.go` - AnalyticsData, ClassificationData, SecurityData, QualityData structs
- Match structure from plan document with proper JSON tags

**Enhancement to existing:**

- Modify `internal/api/handlers/merchant_portfolio_handler.go` - Update `GetMerchantAnalytics` to use new service (currently returns mock data)

#### 2.3.2 Risk Assessment Endpoints

**File: `internal/api/handlers/risk_assessment_handler.go` (new or enhance existing)**

- Enhance `POST /api/v1/risk/assess` endpoint
- Implement async assessment pattern:
- Accept request → Create assessment record → Queue background job → Return 202 Accepted with assessment ID
- Add handler for checking assessment status: `GET /api/v1/risk/assess/{assessmentId}`

**File: `internal/services/risk_assessment_service.go` (new or enhance)**

- Create/enhance service for risk assessments
- Methods:
- `StartAssessment(ctx, merchantId, options) (assessmentId, error)`
- `GetAssessmentStatus(ctx, assessmentId) (*AssessmentStatus, error)`
- `ProcessAssessment(ctx, assessmentId)` - Background processing

**File: `internal/jobs/risk_assessment_job.go` (new)**

- Create background job processor for risk assessments
- Use Go channels or a simple job queue
- Process assessments asynchronously
- Update assessment status in database

**File: `internal/database/risk_assessment_repository.go` (new)**

- Create repository for risk assessment data
- Methods:
- `CreateAssessment(ctx, assessment) error`
- `GetAssessmentByID(ctx, assessmentId) (*RiskAssessment, error)`
- `UpdateAssessmentStatus(ctx, assessmentId, status) error`

**Database migration (if needed):**

- Create `risk_assessments` table if it doesn't exist
- Columns: id, merchant_id, status, options, result, created_at, updated_at, completed_at

**File: `internal/api/routes/risk_routes.go` (modify)**

- Ensure `POST /api/v1/risk/assess` route is registered
- Add `GET /api/v1/risk/assess/{assessmentId}` route

## Task 3: Testing Infrastructure Setup

### 3.1 Enhance Automated Testing Framework

**File: `test/e2e/merchant_details_e2e_test.go` (new)**

- Create comprehensive E2E tests for merchant-details page
- Test scenarios:
- Navigation from add-merchant to merchant-details
- Tab switching and content loading
- API data loading and display
- Error handling and fallbacks
- Export functionality

**File: `test/e2e/merchant_analytics_api_test.go` (new)**

- Create API integration tests for analytics endpoints
- Test:
- GET /api/v1/merchants/{merchantId}/analytics
- GET /api/v1/merchants/{merchantId}/website-analysis
- Error cases (invalid merchant ID, missing data)
- Response structure validation

**File: `test/integration/risk_assessment_integration_test.go` (new or enhance)**

- Test risk assessment endpoints
- Test async assessment flow:
- POST /api/v1/risk/assess → 202 Accepted
- Poll GET /api/v1/risk/assess/{assessmentId} until complete
- Verify assessment results

**File: `test/fixtures/merchant_test_data.go` (enhance)**

- Add test merchant records with complete data
- Add test merchants with partial data
- Add test merchants with risk assessment data
- Add test merchants with analytics data

**File: `test/sql/test_merchant_data.sql` (new)**

- SQL scripts to seed test database
- Insert test merchants with various data completeness levels
- Insert classification data
- Insert risk assessment data

### 3.2 Enhance CI/CD Pipeline

**File: `.github/workflows/e2e-tests.yml` (modify)**

- Add merchant-details specific test job or enhance existing e2e-tests job
- Ensure merchant-details tests run on PRs
- Add test result reporting for merchant-details tests

**File: `.github/workflows/merchant-details-tests.yml` (new, optional)**

- Dedicated workflow for merchant-details tests if needed
- Run on changes to merchant-details related files
- Path filters:
- `cmd/frontend-service/static/merchant-details.html`
- `internal/api/handlers/merchant_analytics_handler.go`
- `internal/api/handlers/risk_assessment_handler.go`
- `test/e2e/merchant_details_e2e_test.go`

### 3.3 Prepare Test Data

**File: `test/testdata/merchants.json` (new or enhance)**

- JSON fixtures for test merchants
- Include complete, partial, and minimal merchant data

**File: `test/testdata/analytics_responses.json` (new)**

- Mock analytics API responses for testing
- Various scenarios: complete data, partial data, errors

**File: `test/testdata/risk_assessments.json` (new)**

- Mock risk assessment responses
- Various risk levels and scenarios

**Enhancement:**

- Update existing test data files if they exist
- Ensure test data covers all edge cases mentioned in plan

## Implementation Order

1. **API Testing Environment** (2.1) - Quick setup, enables manual testing
2. **API Documentation** (2.2) - Define contracts before implementation
3. **Database/Repository Layer** (2.3.1, 2.3.2) - Foundation for services
4. **Service Layer** (2.3.1, 2.3.2) - Business logic
5. **Handler Layer** (2.3.1, 2.3.2) - HTTP handlers
6. **Route Registration** (2.3.1, 2.3.2) - Wire everything together
7. **Test Data Preparation** (3.3) - Enable testing
8. **E2E Tests** (3.1) - Validate end-to-end flow
9. **CI/CD Enhancement** (3.2) - Automate testing

## Success Criteria

- [ ] Postman/Insomnia collection created and working
- [ ] OpenAPI specification complete with all endpoints
- [ ] GET /api/v1/merchants/{merchantId}/analytics returns real data (not mock)
- [ ] GET /api/v1/merchants/{merchantId}/website-analysis implemented
- [ ] POST /api/v1/risk/assess implements async pattern with 202 response
- [ ] Background job system processes risk assessments
- [ ] E2E tests for merchant-details page created and passing
- [ ] API integration tests for new endpoints passing
- [ ] CI/CD pipeline runs merchant-details tests on PRs
- [ ] Test data fixtures prepared and documented