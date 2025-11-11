# KYB Platform Testing Guide

**Version**: 1.0  
**Last Updated**: 2025-01-27

---

## Overview

This guide provides comprehensive testing instructions for the KYB Platform, including automated test scripts, manual testing procedures, and test checklists.

---

## Table of Contents

1. [Automated Test Scripts](#automated-test-scripts)
2. [Manual Testing Procedures](#manual-testing-procedures)
3. [Test Checklists](#test-checklists)
4. [Test Scenarios](#test-scenarios)
5. [Troubleshooting](#troubleshooting)

---

## Automated Test Scripts

### Prerequisites

- Bash shell (macOS/Linux)
- `curl` command-line tool
- `bc` for calculations (optional, for load testing)
- JWT token for protected endpoints (optional)

### Setup

```bash
# Navigate to project root
cd /path/to/kyb-platform

# Make scripts executable (already done, but verify)
chmod +x scripts/test-api-endpoints.sh
chmod +x scripts/test-integration.sh
chmod +x scripts/test-load.sh

# Set environment variables
export API_BASE_URL="https://api-gateway-service-production-21fd.up.railway.app"
export JWT_TOKEN="your-jwt-token-here"  # Optional, for protected endpoints
export TEST_RESULTS_DIR="./test-results"
```

### 1. API Endpoint Testing

**Script**: `scripts/test-api-endpoints.sh`

**Purpose**: Tests all API endpoints with various scenarios including:
- Health checks
- Classification endpoint
- Merchant endpoints (with authentication)
- Risk assessment endpoints
- Error handling
- CORS configuration
- Security headers
- Authentication
- Rate limiting

**Usage**:
```bash
./scripts/test-api-endpoints.sh
```

**Output**:
- Test results printed to console
- JSON responses saved to `test-results/`
- Test report generated

**Example Output**:
```
==========================================
Testing: Health Check Endpoints
==========================================
  Testing api_gateway_health... ✓ PASSED (HTTP 200)
  Testing api_gateway_health_detailed... ✓ PASSED (HTTP 200)
  ...
```

### 2. Integration Testing

**Script**: `scripts/test-integration.sh`

**Purpose**: Tests end-to-end flows including:
- Complete merchant verification flow
- Data consistency across services
- Error scenarios
- Cross-service communication
- Response time validation

**Usage**:
```bash
./scripts/test-integration.sh
```

**Features**:
- Tests complete merchant verification flow (classify → create → assess)
- Verifies data persistence
- Tests cross-service communication
- Measures response times

### 3. Load Testing

**Script**: `scripts/test-load.sh`

**Purpose**: Tests API endpoints under load to:
- Identify bottlenecks
- Test rate limiting
- Monitor resource usage
- Measure performance under load

**Usage**:
```bash
# Default: 10 concurrent users, 60 seconds
./scripts/test-load.sh

# Custom configuration
CONCURRENT_USERS=20 TEST_DURATION=120 ./scripts/test-load.sh
```

**Configuration**:
- `CONCURRENT_USERS`: Number of concurrent users (default: 10)
- `REQUESTS_PER_USER`: Requests per user (default: 10)
- `TEST_DURATION`: Test duration in seconds (default: 60)

---

## Manual Testing Procedures

### 1. UI Flow Testing

#### Test Environment Setup

1. Open browser (Chrome, Firefox, or Safari - latest versions)
2. Navigate to: `https://frontend-service-production-b225.up.railway.app`
3. Open browser developer tools (F12)
4. Enable network throttling (optional, for testing slow connections)

#### Test Scenarios

**Scenario 1: Add Merchant Flow**

1. Navigate to "Add Merchant" page
2. Fill in form:
   - Business Name: "Test Company Inc"
   - Legal Name: "Test Company Incorporated"
   - Address: "123 Test Street, Test City, TS 12345"
   - Industry: "Retail"
   - Website: "https://testcompany.com"
3. Submit form
4. **Verify**:
   - ✅ Form submits without errors
   - ✅ Redirect to merchant details page
   - ✅ All three analyses appear (Business Intelligence, Risk Assessment, Risk Indicators)
   - ✅ No console errors
   - ✅ Data persists after page refresh

**Scenario 2: Merchant List with Filtering**

1. Navigate to merchant list page
2. Test pagination:
   - Click "Next" button
   - Click "Previous" button
   - Verify page numbers update correctly
3. Test filtering:
   - Filter by Portfolio Type
   - Filter by Risk Level
   - Filter by Status
   - Use search query
4. Test sorting:
   - Sort by Name (ascending/descending)
   - Sort by Created Date
   - Sort by Risk Level
5. **Verify**:
   - ✅ All filters work correctly
   - ✅ Sorting works for all fields
   - ✅ Combined filters work together
   - ✅ Results update without page reload

**Scenario 3: Error Handling**

1. Submit form with missing required fields
2. **Verify**:
   - ✅ Error messages are displayed
   - ✅ Error messages are clear and helpful
   - ✅ Form doesn't submit
   - ✅ No console errors

2. Test with network throttling (slow 3G)
3. **Verify**:
   - ✅ Loading indicators appear
   - ✅ Timeout handling works
   - ✅ Error messages are shown if timeout occurs

### 2. Browser Compatibility Testing

**Test Browsers**:
- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

**Test on Each Browser**:
- [ ] Add merchant flow works
- [ ] Merchant list displays correctly
- [ ] Filtering and sorting work
- [ ] No console errors
- [ ] Responsive design works
- [ ] Forms submit correctly

### 3. Responsive Design Testing

**Test Screen Sizes**:
- Desktop (1920x1080)
- Laptop (1366x768)
- Tablet (768x1024)
- Mobile (375x667)

**Verify on Each Size**:
- [ ] Layout is readable
- [ ] Forms are usable
- [ ] Navigation works
- [ ] No horizontal scrolling
- [ ] Touch targets are adequate (mobile)

---

## Test Checklists

### API Endpoint Testing Checklist

#### Health Checks
- [ ] `GET /health` returns 200
- [ ] `GET /health?detailed=true` returns detailed info
- [ ] `GET /api/v1/classification/health` returns 200
- [ ] `GET /api/v1/merchant/health` returns 200
- [ ] `GET /api/v1/risk/health` returns 200

#### Classification
- [ ] `POST /api/v1/classify` with valid data returns 200
- [ ] `POST /api/v1/classify` with missing name returns 400
- [ ] `POST /api/v1/classify` with invalid JSON returns 400
- [ ] Cached responses return faster (< 100ms)

#### Merchants (Requires Authentication)
- [ ] `GET /api/v1/merchants` returns 200
- [ ] `GET /api/v1/merchants?page=1&page_size=10` works
- [ ] `GET /api/v1/merchants?portfolio_type=enterprise` filters correctly
- [ ] `GET /api/v1/merchants?sort_by=name&sort_order=asc` sorts correctly
- [ ] `POST /api/v1/merchants` creates merchant (201)
- [ ] `POST /api/v1/merchants` with missing fields returns 400
- [ ] `GET /api/v1/merchants/{id}` returns merchant (200)
- [ ] `GET /api/v1/merchants/{id}` with invalid ID returns 404

#### Risk Assessment (Requires Authentication)
- [ ] `POST /api/v1/risk/assess` returns 200
- [ ] `POST /api/v1/risk/assess` with missing fields returns 400
- [ ] `GET /api/v1/risk/benchmarks?mcc=5411` returns 200
- [ ] `GET /api/v1/risk/benchmarks` without params returns 400

#### Error Handling
- [ ] Invalid endpoint returns 404
- [ ] Wrong HTTP method returns 405
- [ ] Invalid JSON returns 400
- [ ] Missing required fields returns 400
- [ ] Error responses have consistent format

#### Security
- [ ] Security headers are present
- [ ] CORS preflight requests work
- [ ] Protected endpoints require authentication
- [ ] Invalid tokens are rejected (401)
- [ ] Rate limiting works (429 after limit)

### Integration Testing Checklist

- [ ] Complete merchant verification flow works end-to-end
- [ ] Data persists across page refreshes
- [ ] Cross-service communication works
- [ ] Error scenarios are handled gracefully
- [ ] Response times are acceptable
- [ ] No data corruption on errors

### Load Testing Checklist

- [ ] System handles 10 concurrent users
- [ ] System handles 50 concurrent users
- [ ] Response times remain acceptable under load
- [ ] Rate limiting prevents abuse
- [ ] No memory leaks during extended load
- [ ] System recovers after load test

---

## Test Scenarios

### Scenario 1: Happy Path - Complete Merchant Verification

**Steps**:
1. Classify business → Get classification result
2. Create merchant → Get merchant ID
3. Perform risk assessment → Get risk score
4. Retrieve merchant → Verify data persisted

**Expected Results**:
- ✅ All steps complete successfully
- ✅ Data is consistent across services
- ✅ Response times are acceptable
- ✅ No errors in logs

### Scenario 2: Error Recovery

**Steps**:
1. Submit invalid data → Get error
2. Submit valid data → Success
3. Verify system recovered

**Expected Results**:
- ✅ Errors are handled gracefully
- ✅ System continues to work after error
- ✅ No data corruption

### Scenario 3: Concurrent Users

**Steps**:
1. Simulate 10 concurrent users
2. Monitor response times
3. Check for errors

**Expected Results**:
- ✅ All requests complete
- ✅ Response times remain acceptable
- ✅ No rate limit issues (within limits)
- ✅ No errors

---

## Troubleshooting

### Common Issues

#### Test Scripts Fail

**Issue**: Scripts return errors or fail to execute

**Solutions**:
- Verify `curl` is installed: `which curl`
- Check script permissions: `ls -l scripts/test-*.sh`
- Verify API_BASE_URL is correct
- Check network connectivity

#### Authentication Errors

**Issue**: Protected endpoints return 401

**Solutions**:
- Set JWT_TOKEN environment variable
- Verify token is valid and not expired
- Check token format: `Bearer <token>`

#### Rate Limiting

**Issue**: Tests fail with 429 errors

**Solutions**:
- Wait for rate limit window to reset
- Reduce number of concurrent requests
- Increase rate limit in configuration (if needed for testing)

#### Slow Response Times

**Issue**: Response times exceed targets

**Solutions**:
- Check network connectivity
- Verify services are running
- Check service logs for errors
- Review database query performance

---

## Test Results

### Viewing Test Results

Test results are saved to `test-results/` directory:

```bash
# List all test results
ls -la test-results/

# View a specific test result
cat test-results/test_report_*.txt

# View JSON response
cat test-results/classification_valid_*.json | jq .
```

### Interpreting Results

- **Green (✓)**: Test passed
- **Red (✗)**: Test failed
- **Yellow (⚠)**: Test skipped or warning

### Success Criteria

- **API Tests**: > 90% pass rate
- **Integration Tests**: All critical flows pass
- **Load Tests**: Response times within SLA, no errors

---

## Continuous Testing

### Running Tests Regularly

```bash
# Daily API tests
0 9 * * * /path/to/scripts/test-api-endpoints.sh

# Weekly integration tests
0 10 * * 1 /path/to/scripts/test-integration.sh

# Monthly load tests
0 11 1 * * /path/to/scripts/test-load.sh
```

### Test Automation

Consider integrating these scripts into CI/CD pipeline:
- Run API tests on every deployment
- Run integration tests before production release
- Run load tests weekly

---

## Support

For testing issues:
- **Documentation**: See this guide and API documentation
- **Test Scripts**: See `scripts/` directory
- **Test Results**: See `test-results/` directory

---

**Last Updated**: 2025-01-27  
**Version**: 1.0

