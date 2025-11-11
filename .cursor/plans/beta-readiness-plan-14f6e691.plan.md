<!-- 14f6e691-2493-4887-8f22-8f0442d86ea6 cb51f6e3-082b-463b-bec4-d855ad2679fd -->
# Beta Readiness Implementation Plan

## Executive Summary

This plan addresses all identified issues and improvements from the Beta readiness analysis to achieve 100% beta testing readiness. The plan is organized by priority with specific tasks, file locations, and acceptance criteria. All tasks must be completed pre-beta - no post-beta improvements.

## Phase 1: Critical Blockers (Must Complete Before Beta)

### Task 1.1: Fix Add-Merchant Redirect Issue

**Priority**: CRITICAL
**Type**: Manual Testing + Code Fix
**Location**: `cmd/frontend-service/static/js/add-merchant.js`, `services/frontend/public/js/add-merchant.js`

**Actions**:

1. Manual browser testing in incognito mode
2. Check browser console for JavaScript errors
3. Verify sessionStorage is being set correctly
4. Test with network throttling to simulate slow API calls
5. Verify merchant-details page loads data from sessionStorage
6. Fix any identified issues (timeout handling, error handling, redirect logic)

**Acceptance Criteria**:

- Form submission successfully redirects to merchant-details page
- Classification data persists in sessionStorage
- No JavaScript console errors
- Works with slow network connections

### Task 1.2: Fix Service Discovery URL Configuration

**Priority**: HIGH
**Type**: Code Fix
**Location**: `cmd/service-discovery/main.go` (lines 604, 616, 628, 664, etc.)

**Actions**:

1. Update all hardcoded service URLs to match current Railway production URLs
2. Replace old URLs with:

- `api-gateway-service-production-21fd.up.railway.app`
- `classification-service-production.up.railway.app`
- `merchant-service-production.up.railway.app`
- `frontend-service-production-b225.up.railway.app`

3. Consider using environment variables for service URLs
4. Test service discovery dashboard after changes

**Acceptance Criteria**:

- All services show as healthy in service discovery dashboard
- No 404 errors when checking service health
- Service URLs match production deployment

### Task 1.3: Complete UI Flow Testing

**Priority**: CRITICAL
**Type**: Manual Testing

**Actions**:

1. Test all critical user journeys:

- User registration flow
- Merchant addition flow
- Classification flow
- Risk assessment flow
- Merchant details view
- Dashboard navigation

2. Verify form submissions work correctly
3. Test navigation flows between pages
4. Verify data persistence across page loads
5. Test responsive design on mobile/tablet
6. Test browser compatibility (Chrome, Firefox, Safari, Edge)
7. Check accessibility (keyboard navigation, screen readers)

**Acceptance Criteria**:

- All critical user journeys work end-to-end
- No broken navigation links
- Forms submit and persist data correctly
- Responsive design works on all screen sizes
- No accessibility blockers

## Phase 2: High Priority Backend Improvements

### Task 2.1: Standardize Error Handling Across Services

**Priority**: HIGH
**Type**: Code Implementation
**Location**:

- `pkg/errors/response.go` (already exists)
- `services/classification-service/internal/handlers/`
- `services/merchant-service/internal/handlers/`
- `services/risk-assessment-service/internal/handlers/`

**Actions**:

1. Adopt error helper from API Gateway (`pkg/errors/response.go`) in all services
2. Replace inconsistent error responses with standardized helpers:

- `WriteError()`, `WriteBadRequest()`, `WriteUnauthorized()`, etc.

3. Ensure all error responses include:

- Request ID
- Timestamp
- Path and method
- Standardized error codes

4. Update all handlers in Classification, Merchant, and Risk Assessment services

**Acceptance Criteria**:

- All services use standardized error response format
- Consistent error structure across all endpoints
- Request IDs included in all error responses
- All error scenarios tested

### Task 2.2: Complete Backend API Testing

**Priority**: HIGH
**Type**: Testing
**Location**: `scripts/test-backend-apis.sh` (already exists)

**Actions**:

1. Run comprehensive backend API test script
2. Test all API endpoints:

- Classification endpoints
- Merchant endpoints
- Risk assessment endpoints
- Authentication endpoints

3. Test error scenarios:

- Invalid input validation
- Missing required fields
- Invalid authentication
- Service unavailable scenarios

4. Test rate limiting with appropriate thresholds
5. Verify CORS configuration
6. Document test results

**Acceptance Criteria**:

- All API endpoints tested and passing
- Error scenarios return appropriate status codes
- Rate limiting works correctly
- CORS configured properly

### Task 2.3: Implement JWT Token Decoding in Merchant Service

**Priority**: HIGH
**Type**: Code Implementation
**Location**: `services/merchant-service/internal/handlers/merchant.go:923`

**Actions**:

1. Extract JWT token from Authorization header
2. Decode JWT to extract user ID
3. Use user ID instead of fallback logic
4. Add proper error handling for invalid tokens
5. Test with valid and invalid tokens

**Acceptance Criteria**:

- User ID extracted from JWT token
- Fallback logic removed or improved
- Proper error handling for invalid tokens
- Tests pass with JWT authentication

### Task 2.4: Security Vulnerability Fixes

**Priority**: HIGH
**Type**: Code Review + Fixes

**Actions**:

1. **SQL Injection Prevention**:

- Audit all SQL queries in all services
- Verify all queries use prepared statements
- Remove any string concatenation in SQL
- Test for SQL injection vulnerabilities
- Files to review:
- `services/merchant-service/internal/repository/`
- `services/risk-assessment-service/internal/repository/`
- `services/classification-service/internal/repository/`

2. **Input Validation**:

- Verify all inputs are validated at API boundaries
- Verify all inputs are sanitized before processing
- Test for validation bypasses
- Review validation logic in all handlers

3. **Authentication Enforcement**:

- Verify authentication is enforced on all protected endpoints
- Test for authentication bypasses
- Review authentication middleware

4. **Security Headers**:

- Add security headers to all services (currently only Risk Assessment has them)
- Review CORS configuration
- Test security headers

**Acceptance Criteria**:

- All SQL queries use prepared statements
- All inputs validated and sanitized
- Authentication enforced on protected endpoints
- Security headers present in all services
- Security audit completed and documented

## Phase 3: Medium Priority Improvements

### Task 3.1: Performance Optimizations

**Priority**: MEDIUM
**Type**: Code Optimization

**Actions**:

1. **Health Check Performance**:

- Optimize API Gateway health check endpoint (currently 430ms, target < 100ms)
- Location: `services/api-gateway/internal/handlers/health.go`

2. **Database Query Optimization**:

- Review N+1 query problems in Merchant Service
- Add missing database indexes
- Optimize slow queries in Risk Assessment Service
- Use connection pooling effectively
- Files to review:
- `services/merchant-service/internal/repository/`
- `services/risk-assessment-service/internal/repository/`

3. **Caching Implementation**:

- Add Redis cache to API Gateway
- Add in-memory cache to Classification Service
- Configure browser caching headers
- Review cache invalidation strategies

4. **Frontend Optimization**:

- Minify JavaScript and CSS files
- Implement code splitting
- Bundle assets
- Compress assets
- Reduce redundant API calls (199 API calls found across 69 files)

**Acceptance Criteria**:

- Health check response time < 100ms
- Database queries optimized with proper indexes
- Caching implemented where appropriate
- Frontend assets minified and optimized
- Performance metrics documented

### Task 3.2: Monitoring and Observability Improvements

**Priority**: MEDIUM
**Type**: Code Implementation
**Location**: `services/risk-assessment-service/cmd/main.go:776, 780`

**Actions**:

1. **Enable Prometheus Metrics Server**:

- Configure Prometheus metrics endpoint
- Add metrics collection for key operations
- Test metrics endpoint

2. **Configure Grafana Dashboards**:

- Set up Grafana dashboard configuration
- Create dashboards for:
- Service health
- API performance
- Error rates
- Database performance

3. **Query Optimizer Metrics**:

- Add GetMetrics to QueryOptimizer if needed
- Location: `services/risk-assessment-service/internal/performance/adapters.go:102`

**Acceptance Criteria**:

- Prometheus metrics endpoint accessible
- Grafana dashboards configured
- Key metrics being collected
- Monitoring documentation updated

### Task 3.3: Code Quality Improvements

**Priority**: MEDIUM
**Type**: Code Refactoring

**Actions**:

1. **Update PostgREST Client Versions**:

- Update indirect dependencies
- Run `go mod tidy` in all services
- Verify builds pass

2. **Reduce Code Duplication**:

- Identify and consolidate ~650 lines of duplication
- Standardize configuration code
- Standardize handler patterns

3. **Address Remaining TODO Items**:

- Review and prioritize TODO comments
- Complete high-priority TODOs
- Document deferred TODOs
- Remove outdated TODOs

**Acceptance Criteria**:

- Dependencies updated and standardized
- Code duplication reduced
- High-priority TODOs completed
- Code quality metrics improved

### Task 3.4: Merchant Service Enhancements

**Priority**: MEDIUM
**Type**: Feature Enhancement
**Location**: `services/merchant-service/internal/handlers/merchant.go:755, 756`

**Actions**:

1. **Enhance Pagination Support**:

- Improve Supabase query with better pagination
- Add cursor-based pagination if needed
- Test with large datasets

2. **Add Filtering and Sorting**:

- Implement filtering capabilities
- Implement sorting capabilities
- Add query parameters for filters and sort
- Test filtering and sorting

**Acceptance Criteria**:

- Pagination works efficiently with large datasets
- Filtering works correctly
- Sorting works correctly
- API documentation updated

## Phase 4: Integration and End-to-End Testing

### Task 4.1: End-to-End Integration Testing

**Priority**: HIGH
**Type**: Testing

**Actions**:

1. Test complete merchant verification flow:

- User registration → Merchant addition → Classification → Risk assessment → Dashboard view

2. Verify data consistency across services
3. Test error scenarios:

- Service failures
- Network timeouts
- Invalid data

4. Test cross-service communication
5. Verify retry logic works correctly
6. Test data persistence

**Acceptance Criteria**:

- Complete end-to-end flow works correctly
- Data consistent across services
- Error handling works correctly
- Retry logic functions properly

### Task 4.2: Load Testing

**Priority**: MEDIUM
**Type**: Testing

**Actions**:

1. Set up load testing tools (e.g., k6, Apache Bench)
2. Test API endpoints under load:

- Classification endpoint
- Merchant endpoints
- Risk assessment endpoints

3. Identify bottlenecks
4. Test rate limiting under load
5. Monitor resource usage

**Acceptance Criteria**:

- Services handle expected load
- No memory leaks under load
- Rate limiting works under load
- Performance metrics documented

## Phase 5: Documentation and Final Preparation

### Task 5.1: Complete API Documentation

**Priority**: MEDIUM
**Type**: Documentation

**Actions**:

1. Document all API endpoints
2. Include request/response examples
3. Document error codes and messages
4. Update OpenAPI/Swagger specifications
5. Document authentication requirements

**Acceptance Criteria**:

- All API endpoints documented
- Examples provided for all endpoints
- Error codes documented
- API documentation accessible

### Task 5.2: Update Deployment Guides

**Priority**: MEDIUM
**Type**: Documentation

**Actions**:

1. Update deployment documentation
2. Document service URLs
3. Document environment variables
4. Document database schema
5. Document monitoring setup

**Acceptance Criteria**:

- Deployment guides complete and accurate
- Service URLs documented
- Environment variables documented
- Database schema documented

### Task 5.3: Create Beta Testing Guide

**Priority**: HIGH
**Type**: Documentation

**Actions**:

1. Create beta testing guide for testers
2. Document known issues and limitations
3. Create test scenarios
4. Document feedback collection process
5. Create rollback plan

**Acceptance Criteria**:

- Beta testing guide complete
- Test scenarios documented
- Feedback process documented
- Rollback plan documented

## Phase 6: Technical Debt (Post-Beta)

### Task 6.1: Consolidate Frontend Services

**Priority**: LOW (Post-Beta)
**Type**: Code Refactoring

**Actions**:

1. Consolidate `services/frontend/` and `cmd/frontend-service/`
2. Create single source of truth
3. Update deployment process
4. Remove duplicate code

### Task 6.2: Remove Legacy Service References

**Priority**: LOW (Post-Beta)
**Type**: Code Cleanup

**Actions**:

1. Remove legacy service URLs from service discovery
2. Clean up deprecated code
3. Update documentation

### Task 6.3: External Integration Completion

**Priority**: LOW (Post-Beta)
**Type**: Feature Implementation

**Actions**:

1. Complete Thomson Reuters client integration (currently mock)
2. Complete World-Check client integration (currently mock)
3. Document integration requirements

## Testing Checklist

### Automated Testing

- [ ] Service health checks (all 9 services)
- [ ] Backend API endpoint testing
- [ ] Load testing
- [ ] Security scanning
- [ ] SQL injection testing
- [ ] Input validation testing

### Manual Testing

- [ ] Browser-based UI flow testing
- [ ] JavaScript console error checking
- [ ] SessionStorage verification
- [ ] Form validation testing
- [ ] Responsive design testing
- [ ] Browser compatibility testing
- [ ] Accessibility testing
- [ ] End-to-end integration testing

## Success Criteria for Beta Readiness

1. ✅ All critical services healthy and operational
2. ✅ All critical UI flows working end-to-end
3. ✅ No critical bugs or security issues
4. ✅ Performance acceptable for beta users
5. ✅ Documentation complete and accurate
6. ✅ Monitoring and alerting in place
7. ✅ Rollback plan documented
8. ✅ Error handling standardized across services
9. ✅ Security vulnerabilities addressed
10. ✅ Beta testing guide created

## Timeline Estimate

- **Phase 1 (Critical Blockers)**: 2-3 days
- **Phase 2 (High Priority Backend)**: 3-4 days
- **Phase 3 (Medium Priority)**: 4-5 days
- **Phase 4 (Integration Testing)**: 2-3 days
- **Phase 5 (Documentation)**: 1-2 days

**Total Estimated Time**: 15-20 days (all pre-beta)

## Risk Mitigation

1. **Add-Merchant Redirect Issue**: If manual testing reveals complex issues, allocate additional time
2. **Security Vulnerabilities**: If critical vulnerabilities found, prioritize fixes immediately
3. **Performance Issues**: If performance targets not met, implement caching and optimization
4. **Integration Issues**: If cross-service communication fails, review service discovery and configuration

## Notes

- All changes should be tested in development before production deployment
- Monitor Railway deployment logs for any issues
- Keep documentation updated as changes are made
- Regular progress reviews should be conducted
- Beta testing should not begin until all Phase 1 and Phase 2 tasks are complete

### To-dos

- [x] Fix add-merchant redirect issue - manual browser testing, check console errors, verify sessionStorage, test with network throttling, fix timeout/error handling
- [x] Update service discovery URLs in cmd/service-discovery/main.go to match current Railway production URLs
- [ ] Complete manual UI flow testing - all critical user journeys, form submissions, navigation, data persistence, responsive design, browser compatibility
- [x] Adopt error helper (pkg/errors/response.go) in Classification, Merchant, and Risk Assessment services - replace inconsistent error responses
- [x] Run comprehensive backend API tests - all endpoints, error scenarios, rate limiting, CORS verification
- [x] Implement JWT token decoding in Merchant Service (merchant.go:923) - extract user ID from JWT, remove fallback logic
- [x] Security: Audit all SQL queries, ensure prepared statements, remove string concatenation, test for SQL injection vulnerabilities
- [x] Security: Verify all inputs validated and sanitized, test for validation bypasses, review validation logic in all handlers
- [x] Security: Verify authentication enforced on protected endpoints, test for bypasses, review authentication middleware
- [x] Add security headers to all services (currently only Risk Assessment has them), review CORS configuration
- [x] Optimize API Gateway health check endpoint (currently 430ms, target < 100ms)
- [x] Review N+1 queries, add missing indexes, optimize slow queries in Risk Assessment Service, use connection pooling
- [x] Add Redis cache to API Gateway, in-memory cache to Classification Service, configure browser caching headers
- [ ] Minify JavaScript/CSS, implement code splitting, bundle assets, reduce redundant API calls (199 calls found)
- [x] Enable Prometheus metrics server in Risk Assessment Service (main.go:776), configure metrics collection
- [x] Configure Grafana dashboards in Risk Assessment Service (main.go:780) - service health, API performance, error rates
- [x] Update PostgREST client versions, run go mod tidy in all services, verify builds
- [x] Identify and consolidate ~650 lines of duplication, standardize configuration and handler patterns
- [x] Enhance pagination support in Merchant Service (merchant.go:755) - improve Supabase query, add cursor-based pagination
- [x] Add filtering and sorting capabilities to Merchant Service (merchant.go:756) - implement query parameters
- [x] Complete end-to-end integration testing - merchant verification flow, data consistency, error scenarios, cross-service communication
- [x] Set up load testing - test API endpoints under load, identify bottlenecks, test rate limiting, monitor resource usage
- [x] Document all API endpoints, include request/response examples, document error codes, update OpenAPI specifications
- [x] Update deployment documentation - service URLs, environment variables, database schema, monitoring setup
- [x] Create beta testing guide - test scenarios, known issues, feedback collection process, rollback plan