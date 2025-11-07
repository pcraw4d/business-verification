# KYB Platform - Testing and Debugging Guide

## Overview

This guide provides comprehensive instructions for testing and debugging the KYB Platform to ensure all APIs are working correctly, no placeholder data is being used, and all features are properly implemented in the UI.

## Quick Start

Run all tests with a single command:

```bash
./scripts/run-all-tests.sh
```

This will execute:
1. **Comprehensive API Testing** - Tests all endpoints and validates responses
2. **Placeholder Data Detection** - Scans codebase for mock/placeholder data
3. **Unused Features Analysis** - Identifies backend features not used in frontend

## Test Suites

### 1. Comprehensive API Testing

**Script**: `scripts/comprehensive-api-test.sh`

**Purpose**: Validates all API endpoints are working correctly and returning real data (not placeholders).

**Features**:
- Tests all API endpoints (Health, Merchant, Risk, BI, Classification)
- Validates HTTP status codes
- Checks for placeholder/mock data in responses
- Validates JSON response format
- Generates detailed JSON report

**Usage**:
```bash
# Use default configuration
./scripts/comprehensive-api-test.sh

# Custom configuration
API_BASE_URL=https://your-api.com \
TEST_MERCHANT_ID=your-merchant-id \
./scripts/comprehensive-api-test.sh
```

**Output**:
- Console output with color-coded results
- JSON report in `test-results/api-test-report-{timestamp}.json`

**What it checks**:
- ✅ HTTP status codes match expected values
- ✅ Responses are valid JSON
- ✅ No placeholder data (Sample, Mock, placeholder, etc.)
- ✅ Responses are not empty or null

### 2. Placeholder Data Detection

**Script**: `scripts/detect-placeholder-data.sh`

**Purpose**: Scans the codebase for placeholder, mock, or temporary data that should be replaced with real implementations.

**Features**:
- Scans Go and JavaScript files
- Detects common placeholder patterns
- Excludes test files
- Provides file locations and line numbers

**Usage**:
```bash
./scripts/detect-placeholder-data.sh
```

**Patterns Detected**:
- "Sample Merchant"
- "Mock" / "mock"
- "TODO.*return"
- "placeholder"
- "test-"
- "dummy"
- "fake"
- "example"
- "For now"
- "temporary"
- "fallback"

**Output**:
- List of files and line numbers with potential placeholder data
- Summary count of issues found

### 3. Unused Features Analysis

**Script**: `scripts/analyze-unused-features.js`

**Purpose**: Identifies backend API endpoints and frontend components that are defined but not being used.

**Features**:
- Extracts all API endpoints from Go handlers
- Extracts all API calls from JavaScript
- Extracts component definitions and usages
- Identifies unused endpoints and components
- Generates JSON report with recommendations

**Usage**:
```bash
node scripts/analyze-unused-features.js
```

**Output**:
- List of unused API endpoints
- List of potentially unused components
- Recommendations for implementation
- JSON report in `test-results/unused-features-analysis.json`

## Manual Testing

### Testing API Endpoints Directly

#### Using curl

```bash
# Health check
curl https://api-gateway-service-production-21fd.up.railway.app/health

# Get merchant
curl https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants/{merchant-id}

# Risk assessment
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"merchantId":"your-merchant-id"}'
```

#### Using the test script

```bash
# Test specific endpoints
./scripts/test-risk-endpoints.sh
```

### Testing Frontend

1. **Open browser console** (F12)
2. **Check for errors** in the console
3. **Monitor network tab** to see API calls
4. **Verify data population** in UI elements

### Debugging Tips

#### Backend Debugging

1. **Check Railway logs**:
   ```bash
   # View logs in Railway dashboard or CLI
   railway logs
   ```

2. **Check Supabase connection**:
   - Verify environment variables are set
   - Check health endpoint: `/api/v1/merchant/health`
   - Review Supabase dashboard for data

3. **Verify API Gateway routing**:
   - Check API Gateway logs
   - Verify service URLs are correct
   - Test direct service calls vs. through gateway

#### Frontend Debugging

1. **Enable detailed logging**:
   - Check browser console for detailed logs
   - Look for API call logs with `📡` emoji
   - Check for data structure logs with `📦` emoji

2. **Check sessionStorage**:
   ```javascript
   // In browser console
   console.log(JSON.parse(sessionStorage.getItem('merchantData')));
   console.log(JSON.parse(sessionStorage.getItem('merchantApiResults')));
   ```

3. **Verify API responses**:
   - Check Network tab in DevTools
   - Verify response status codes
   - Check response content-type (should be `application/json`)
   - Inspect response body for actual data

## Common Issues and Solutions

### Issue: API returns placeholder data

**Symptoms**: API responses contain "Sample Merchant", "Mock", etc.

**Solution**:
1. Check if Supabase query is working (see logs)
2. Verify merchant exists in Supabase database
3. Check if fallback to mock data is being used
4. Review `getMerchant()` function in merchant service

### Issue: Frontend not displaying data

**Symptoms**: UI shows "-" or empty fields

**Solution**:
1. Check browser console for errors
2. Verify API calls are successful (Network tab)
3. Check if data structure matches expected format
4. Verify `populateMerchantDetails()` is being called
5. Check field name variations (camelCase vs snake_case)

### Issue: CORS errors

**Symptoms**: Browser console shows CORS policy errors

**Solution**:
1. Verify API Gateway CORS middleware is configured
2. Check that only one `Access-Control-Allow-Origin` header is set
3. Verify allowed origins include frontend URL
4. Check preflight (OPTIONS) requests are handled

### Issue: 404 errors for API endpoints

**Symptoms**: API calls return 404

**Solution**:
1. Verify route is registered in API Gateway
2. Check path transformation in proxy handlers
3. Verify service is running and healthy
4. Check service URL configuration

## Continuous Testing

### Pre-commit Testing

Add to your workflow:
```bash
# Before committing
./scripts/run-all-tests.sh
```

### CI/CD Integration

Add to your CI/CD pipeline:
```yaml
# Example GitHub Actions
- name: Run API Tests
  run: ./scripts/comprehensive-api-test.sh

- name: Check for Placeholders
  run: ./scripts/detect-placeholder-data.sh

- name: Analyze Unused Features
  run: node scripts/analyze-unused-features.js
```

## Test Results

All test results are saved in the `test-results/` directory:

- `api-test-report-{timestamp}.json` - Detailed API test results
- `unused-features-analysis.json` - Unused features analysis

## Best Practices

1. **Run tests before deploying** - Ensure no placeholder data goes to production
2. **Review unused features regularly** - Implement UI for backend features
3. **Monitor API responses** - Use the comprehensive test script regularly
4. **Check logs** - Review both backend and frontend logs for issues
5. **Test with real data** - Use actual merchant IDs, not test IDs

## Troubleshooting

### Tests fail but services work

- Check if test merchant ID exists in database
- Verify API Gateway URL is correct
- Check network connectivity
- Review test script configuration

### False positives in placeholder detection

- Some patterns may be legitimate (e.g., "example" in documentation)
- Review each match manually
- Update patterns in script if needed

### Unused features analysis shows too many

- Some features may be used indirectly
- Check for dynamic API calls
- Review component instantiation patterns
- Manually verify before removing

## Support

For issues or questions:
1. Check logs (backend and frontend)
2. Review test reports
3. Check Railway deployment status
4. Verify Supabase connection

