# Comprehensive Testing Documentation

**Date**: 2025-01-XX  
**Version**: 1.0.0  
**Status**: Active

## Overview

This document provides comprehensive testing guidelines, procedures, and checklists for the KYB Platform frontend application. It covers manual testing, automated testing, API verification, and deployment procedures.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Pre-Deployment Testing](#pre-deployment-testing)
3. [Automated Testing Scripts](#automated-testing-scripts)
4. [Manual Testing Checklist](#manual-testing-checklist)
5. [API Endpoint Verification](#api-endpoint-verification)
6. [Page-by-Page Testing Guide](#page-by-page-testing-guide)
7. [Component Testing](#component-testing)
8. [Error Handling Testing](#error-handling-testing)
9. [Performance Testing](#performance-testing)
10. [Deployment Checklist](#deployment-checklist)
11. [Troubleshooting Guide](#troubleshooting-guide)

## Testing Strategy

### Testing Pyramid

```
              /\
             /  \
            / E2E \
           /______\
          /        \
         /  Manual  \
        /____________\
       /              \
      /   Automated   \
     /_________________ \
```

### Test Types

1. **Unit Tests** - Component and utility function tests
2. **Integration Tests** - API integration and data flow tests
3. **E2E Tests** - Full user journey tests (Playwright)
4. **Visual Regression Tests** - UI consistency tests
5. **Manual Tests** - Human verification of functionality
6. **Automated Page Tests** - HTTP-based page accessibility tests

## Pre-Deployment Testing

### 1. Build Verification

**Script**: `npm run verify-env`

**Checks**:
- ✅ `NEXT_PUBLIC_API_BASE_URL` is set
- ✅ Not set to localhost in production
- ✅ Valid URL format
- ✅ Uses HTTPS in production

**Usage**:
```bash
cd frontend
npm run verify-env
```

**Expected Output**:
```
✅ NEXT_PUBLIC_API_BASE_URL is set: https://api-gateway-service-production-21fd.up.railway.app
✅ All required checks passed!
```

### 2. TypeScript Compilation

**Command**: `npm run build`

**Checks**:
- TypeScript compilation
- Build verification script
- Next.js build process

### 3. Linting

**Command**: `npm run lint`

**Checks**:
- ESLint rules compliance
- TypeScript strict mode
- Code quality standards

## Automated Testing Scripts

### Page Testing Script

**Script**: `npm run test:pages`

**Purpose**: Tests all pages for HTTP accessibility and basic functionality

**Features**:
- Tests 33+ pages for 200 status codes
- Checks for localhost API references
- Verifies API endpoint accessibility
- Provides detailed test report

**Usage**:
```bash
# Basic test
npm run test:pages

# Verbose output
npm run test:pages:verbose

# Custom base URL
npm run test:pages -- --base-url http://localhost:3000

# Custom API URL
npm run test:pages -- --api-url https://api-gateway-service-production-21fd.up.railway.app
```

**Environment Variables**:
- `BASE_URL` - Frontend base URL (default: http://localhost:3000)
- `API_URL` - API Gateway URL (default: Railway production URL)

**Output**:
```
🚀 Starting Automated Page Testing

Base URL: http://localhost:3000
API URL: https://api-gateway-service-production-21fd.up.railway.app
Testing 33 pages...

✅ Passed: 30
❌ Failed: 2
⚠️  Warnings: 1

🌐 API Endpoints: 5/5 accessible
```

### Unit Tests

**Command**: `npm test`

**Coverage**: Component tests, utility functions, API client tests

### E2E Tests

**Command**: `npm run test:e2e`

**Coverage**: Full user journeys, form submissions, navigation

### Visual Regression Tests

**Command**: `npm run test:visual`

**Coverage**: UI consistency across browsers and viewports

## Manual Testing Checklist

### Critical Paths (Must Test Before Deployment)

#### 1. Add Merchant Flow
- [ ] Navigate to `/add-merchant`
- [ ] Fill out merchant form
- [ ] Submit form
- [ ] Verify success message
- [ ] Check API call goes to Railway (not localhost)
- [ ] Verify merchant appears in portfolio
- [ ] Check browser console for errors

#### 2. Merchant Portfolio
- [ ] Navigate to `/merchant-portfolio`
- [ ] Verify merchant list loads
- [ ] Test search functionality
- [ ] Test filtering
- [ ] Test sorting
- [ ] Click on merchant to view details
- [ ] Verify all tabs work (Overview, Analytics, Risk, etc.)

#### 3. Risk Assessment
- [ ] Navigate to `/risk-dashboard`
- [ ] Verify risk metrics load
- [ ] Check charts render correctly
- [ ] Navigate to `/risk-indicators`
- [ ] Verify indicators display
- [ ] Test filtering options

#### 4. Compliance Pages
- [ ] Navigate to `/compliance`
- [ ] Verify compliance status loads
- [ ] Navigate to `/compliance/gap-analysis`
- [ ] Verify page loads (no 404)
- [ ] Navigate to `/compliance/progress-tracking`
- [ ] Verify page loads (no 404)

### Platform Pages

- [ ] `/` - Home page auto-redirects after 3 seconds
- [ ] `/dashboard-hub` - All dashboard links work
- [ ] `/dashboard` - Business Intelligence metrics load

### Merchant Management

- [ ] `/merchant-hub` - Hub page loads
- [ ] `/merchant-hub/integration` - Integration page loads
- [ ] `/merchant-portfolio` - Portfolio loads with data
- [ ] `/merchant/bulk-operations` - Bulk operations page loads
- [ ] `/merchant/comparison` - Comparison page loads
- [ ] `/merchant-details/[id]` - Details page loads for valid IDs
- [ ] `/risk-assessment/portfolio` - Risk portfolio loads

### Administration

- [ ] `/admin` - Admin dashboard loads
- [ ] `/admin/models` - Models page loads
- [ ] `/admin/queue` - Queue page loads
- [ ] `/sessions` - Sessions page loads and fetches data

### Additional Pages

- [ ] `/register` - Registration form works
- [ ] `/monitoring` - Monitoring dashboard loads
- [ ] `/analytics-insights` - Analytics page loads
- [ ] `/business-intelligence` - BI page loads
- [ ] `/market-analysis` - Market analysis loads
- [ ] `/competitive-analysis` - Competitive analysis loads

## API Endpoint Verification

### Critical Endpoints

Test these endpoints directly via API Gateway:

1. **Merchant Endpoints**
   - `GET /api/v1/merchants` - List merchants
   - `POST /api/v1/merchants` - Create merchant
   - `GET /api/v1/merchants/{id}` - Get merchant
   - `GET /api/v1/merchants/{id}/analytics` - Get analytics

2. **Dashboard Endpoints**
   - `GET /api/v3/dashboard/metrics` - Dashboard metrics (v3)
   - `GET /api/v1/dashboard/metrics` - Dashboard metrics (v1 fallback)

3. **Risk Endpoints**
   - `GET /api/v1/risk/metrics` - Risk metrics
   - `GET /api/v1/risk/assess` - Start assessment
   - `GET /api/v1/risk/indicators/{merchantId}` - Risk indicators

4. **Compliance Endpoints**
   - `GET /api/v1/compliance/status` - Compliance status

5. **Session Endpoints**
   - `GET /api/v1/sessions` - List sessions

### Testing API Endpoints

**Using curl**:
```bash
# Test merchant list
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants

# Test dashboard metrics
curl https://api-gateway-service-production-21fd.up.railway.app/api/v3/dashboard/metrics
```

**Using Browser DevTools**:
1. Open Network tab
2. Navigate to page
3. Check API calls:
   - ✅ URL should be Railway API Gateway (not localhost)
   - ✅ Status should be 200 (or appropriate status)
   - ✅ Response should contain expected data
   - ✅ No CORS errors

## Page-by-Page Testing Guide

### Home Page (`/`)

**Expected Behavior**:
- Page loads immediately
- Auto-redirects to `/merchant-portfolio` after 3 seconds
- Shows welcome message and feature cards
- "Enter Merchant Portfolio" button works

**API Calls**:
- None (static page)

**Console Checks**:
- No errors
- No warnings about localhost

### Add Merchant Page (`/add-merchant`)

**Expected Behavior**:
- Form loads with all fields
- Validation works (required fields, email format, etc.)
- Submit button works
- Success message appears after submission
- Redirects to merchant details or portfolio

**API Calls**:
- `POST /api/v1/merchants` - Should go to Railway API Gateway

**Console Checks**:
- No errors
- API call shows correct URL (not localhost)
- Success/error messages display correctly

### Merchant Portfolio (`/merchant-portfolio`)

**Expected Behavior**:
- Merchant list loads
- Search works
- Filters work
- Sorting works
- Pagination works
- Click merchant opens details page

**API Calls**:
- `GET /api/v1/merchants` - With query parameters for search/filter

**Console Checks**:
- No errors
- Data loads correctly
- Empty states display when no data

### Compliance Pages

#### Gap Analysis (`/compliance/gap-analysis`)

**Expected Behavior**:
- Page loads (no 404)
- Content displays
- No RSC errors in console

**API Calls**:
- May fetch compliance data

**Console Checks**:
- No 404 errors for `?_rsc=*` requests
- No React Server Component errors

#### Progress Tracking (`/compliance/progress-tracking`)

**Expected Behavior**:
- Page loads (no 404)
- Progress indicators display
- No RSC errors

**API Calls**:
- May fetch compliance progress data

**Console Checks**:
- No 404 errors for `?_rsc=*` requests
- No React Server Component errors

### Risk Assessment Portfolio (`/risk-assessment/portfolio`)

**Expected Behavior**:
- Page loads (no 404)
- Risk assessment table displays
- No RSC errors

**API Calls**:
- May fetch risk assessment data

**Console Checks**:
- No 404 errors for `?_rsc=*` requests
- No React Server Component errors

## Component Testing

### Core Components

#### MerchantForm
- [ ] All fields render
- [ ] Validation works
- [ ] Submit handler called
- [ ] Error messages display
- [ ] Success state works

#### DataEnrichment
- [ ] Sources load
- [ ] Trigger enrichment works
- [ ] Loading states display
- [ ] Error handling works

#### ExportButton
- [ ] Export button renders
- [ ] Click triggers export
- [ ] File downloads correctly
- [ ] Different formats work (CSV, PDF, JSON, Excel)

#### BulkOperationsManager
- [ ] Selection works
- [ ] Bulk operations execute
- [ ] Progress tracking works
- [ ] Error handling works

#### DataTable
- [ ] Data displays
- [ ] Sorting works
- [ ] Filtering works
- [ ] Pagination works
- [ ] Search works

## Error Handling Testing

### Network Errors

1. **Disconnect Network**
   - Navigate to page
   - Try to submit form
   - Verify error message displays
   - Verify retry mechanism works

2. **API Timeout**
   - Simulate slow network
   - Verify timeout handling
   - Verify error message

3. **API Error Response**
   - Test with invalid data
   - Verify error messages
   - Verify form doesn't break

### Validation Errors

- [ ] Required field validation
- [ ] Email format validation
- [ ] Phone format validation
- [ ] URL format validation
- [ ] Number range validation

## Performance Testing

### Lighthouse Audit

**Command**: `npm run lighthouse`

**Targets**:
- Performance: ≥90
- Accessibility: 100
- Best Practices: ≥90
- SEO: ≥90

### Bundle Analysis

**Command**: `npm run analyze-bundle`

**Checks**:
- Bundle size
- Code splitting
- Duplicate dependencies
- Unused code

## Deployment Checklist

### Pre-Deployment

- [ ] All tests pass (`npm test`)
- [ ] Build verification passes (`npm run verify-env`)
- [ ] TypeScript compilation succeeds (`npm run build`)
- [ ] Linting passes (`npm run lint`)
- [ ] No console errors in development
- [ ] All critical paths tested manually

### Railway Configuration

- [ ] `NEXT_PUBLIC_API_BASE_URL` set to Railway API Gateway URL
- [ ] `NODE_ENV` set to `production`
- [ ] `USE_NEW_UI` set (if applicable)
- [ ] `NEXT_PUBLIC_USE_NEW_UI` set (if applicable)
- [ ] All required environment variables set

### Deployment Steps

1. **Set Environment Variables in Railway**
   ```
   NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
   NODE_ENV=production
   ```

2. **Trigger Rebuild**
   - Railway will automatically rebuild when env vars change
   - Or manually trigger rebuild from Railway dashboard

3. **Monitor Build Logs**
   - Check build verification script output
   - Verify environment variables are detected
   - Check for build errors

4. **Post-Deployment Verification**
   - [ ] Frontend service is accessible
   - [ ] Home page loads
   - [ ] No console errors
   - [ ] API calls go to Railway (not localhost)
   - [ ] No CORS errors
   - [ ] All critical pages load

### Post-Deployment Testing

1. **Run Automated Tests**
   ```bash
   npm run test:pages -- --base-url https://frontend-service-production-b225.up.railway.app
   ```

2. **Manual Browser Testing**
   - Test all critical paths
   - Check browser console
   - Verify network requests
   - Test on multiple browsers

3. **Monitor for 24 Hours**
   - Check error logs
   - Monitor API error rates
   - Verify no localhost API calls
   - Check CORS errors

## Troubleshooting Guide

### Issue: API calls go to localhost

**Symptoms**:
- Browser console shows `localhost:8080` in network requests
- CORS errors
- API calls fail

**Solution**:
1. Verify `NEXT_PUBLIC_API_BASE_URL` is set in Railway
2. **IMPORTANT**: Rebuild the service (env vars are embedded at build time)
3. Check build logs for verification script output
4. Verify the variable is available during build

### Issue: 404 errors on RSC requests

**Symptoms**:
- Console shows `?_rsc=*` requests returning 404
- Pages don't load correctly

**Solution**:
1. Verify pages have metadata exports
2. Check Next.js App Router configuration
3. Verify pages are server components (not client components)
4. Check for routing conflicts

### Issue: Build fails with env var error

**Symptoms**:
- Build verification script fails
- Build stops with error

**Solution**:
1. Set `NEXT_PUBLIC_API_BASE_URL` in Railway
2. Verify URL format is correct
3. Check for typos in variable name
4. Ensure variable is set before build starts

### Issue: CORS errors

**Symptoms**:
- Browser console shows CORS errors
- API calls fail

**Solution**:
1. Verify API Gateway CORS configuration
2. Check `CORS_ALLOWED_ORIGINS` includes frontend URL
3. Verify frontend origin matches allowed origins
4. Check API Gateway logs

### Issue: Pages return 500 errors

**Symptoms**:
- Pages don't load
- Server errors in console

**Solution**:
1. Check Railway service logs
2. Verify API Gateway is accessible
3. Check for missing environment variables
4. Verify database connections (if applicable)

## Test Results Template

```markdown
## Test Results - [Date]

### Environment
- Frontend URL: [URL]
- API Gateway URL: [URL]
- Browser: [Browser/Version]

### Automated Tests
- Page Tests: [X/Y passed]
- API Endpoint Tests: [X/Y accessible]
- Unit Tests: [X/Y passed]
- E2E Tests: [X/Y passed]

### Manual Tests
- Critical Paths: [X/Y passed]
- Platform Pages: [X/Y passed]
- Merchant Pages: [X/Y passed]
- Compliance Pages: [X/Y passed]
- Admin Pages: [X/Y passed]

### Issues Found
1. [Issue description]
   - Severity: [Critical/High/Medium/Low]
   - Status: [Open/Fixed/Deferred]

### Notes
[Additional notes or observations]
```

## Continuous Testing

### Daily Checks
- Run automated page tests
- Check error logs
- Monitor API error rates

### Weekly Checks
- Full manual testing of critical paths
- Lighthouse audit
- Bundle analysis
- Performance review

### Pre-Release Checks
- Complete testing checklist
- All automated tests pass
- Manual testing of all features
- Performance benchmarks met
- Documentation updated

---

**Last Updated**: 2025-01-XX  
**Next Review**: After deployment

