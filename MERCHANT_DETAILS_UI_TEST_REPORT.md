# Merchant Details UI Test Report

**Date**: 2025-01-11  
**Environment**: Railway Production Deployment (`https://frontend-service-production-b225.up.railway.app/`)  
**Test Scenario**: Complete UI functionality test of merchant details page after add merchant flow

## Test Summary

### ✅ **Working Features**

1. **Page Rendering**: The merchant details page renders correctly with all 8 tabs visible
2. **MutationObserver Fix**: The infinite loop fix is working - observer disconnects after button setup
3. **Risk Assessment Tab**: Fully functional with visualizations (risk gauge, trend chart, factor chart)
4. **Tab Navigation**: All 8 tabs are clickable and respond to clicks
5. **Merchant Data Population**: Merchant name, website, and other basic data are populated correctly
6. **Page Title**: Correctly displays "UI Test Business Corp - Merchant Details - KYB Platform"

### ❌ **Issues Found**

#### **Issue 1: Tab Content Not Switching Correctly**
**Severity**: HIGH  
**Description**: Multiple tabs are displaying the same content (Business Analytics content) instead of their own unique content.

**Affected Tabs**:
- **Merchant Detail tab**: Shows Business Analytics content (should show merchant detail form)
- **Business Analytics tab**: Shows Business Analytics content (correct)
- **Overview tab**: Shows Business Analytics content (should show overview information)
- **Financial tab**: Shows Business Analytics content (should show financial information)
- **Compliance tab**: Shows Business Analytics content (should show compliance information)

**Expected Behavior**:
- Each tab should display its own unique content
- Tab switching should hide/show the appropriate tab content containers

**Root Cause Hypothesis**:
- Tab content containers may not have proper `active` class management
- Tab switching logic may not be correctly hiding/showing tab content
- Multiple tab content containers may be visible at the same time

#### **Issue 2: Risk Indicators Tab Error**
**Severity**: HIGH  
**Description**: The Risk Indicators tab displays an error message: "Failed to Load Risk Indicator" with "There was an error loading the risk data. Please try again." and a "Retry" button.

**Evidence from Console**:
```
⚠️ Failed to load merchant data, using fallback data: Error: Unable to load merchant details
⚠️ Failed to load analytics data, using fallback data: TypeError: Failed to fetch
⚠️ Failed to load risk assessment, using fallback data: Error: HTTP error! status: 404
```

**API Errors**:
- `GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants/biz_uitestbu_1762889812022` - CORS error: "The 'Access-Control-Allow-Origin' header contains multiple values 'https://frontend-service-production-b225.up.railway.app, *', but only one is allowed."
- `GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk-assessment/...` - 404 Not Found

**Root Cause Hypothesis**:
- API Gateway CORS configuration issue (duplicate Access-Control-Allow-Origin headers)
- Risk assessment API endpoint may not exist or merchant ID not found
- Fallback data mechanism may not be working correctly

#### **Issue 3: Contact Tab Minimal Content**
**Severity**: MEDIUM  
**Description**: The Contact tab only displays "Website: http://www.uitestcorp.com" with minimal content.

**Expected Behavior**:
- Should display full contact information (phone, email, address, etc.)
- Should match the data entered in the add merchant form

**Root Cause Hypothesis**:
- Contact tab content may not be properly populated from merchant data
- Contact tab template may be incomplete or missing fields

#### **Issue 4: API Configuration Issues**
**Severity**: MEDIUM  
**Description**: Multiple API calls are returning HTML instead of JSON, indicating routing or endpoint issues.

**Evidence from Console**:
```
⚠️ API returned non-JSON response for features, using empty array
⚠️ API returned non-JSON response for supported sources, using empty array
⚠️ Error fetching data source info, using default mock data: Non-JSON response received
```

**Root Cause Hypothesis**:
- API endpoints may be returning HTML error pages instead of JSON
- API Gateway routing may be incorrect
- Endpoints may not exist or are misconfigured

#### **Issue 5: Export Button Module Loading Error**
**Severity**: LOW  
**Description**: Export button module fails to load with MIME type error.

**Evidence from Console**:
```
Failed to load module script: Expected a JavaScript-or-Wasm module script but the server responded with a MIME type of "text/html".
```

**Root Cause Hypothesis**:
- Export button module file may not exist or path is incorrect
- Server may be returning HTML instead of JavaScript for the module

## Console Errors Summary

### Critical Errors:
1. **CORS Error**: Multiple Access-Control-Allow-Origin headers in API Gateway response
2. **404 Errors**: Risk assessment API endpoint not found
3. **Non-JSON Responses**: Multiple API endpoints returning HTML instead of JSON

### Warnings:
1. Export button module loading failure (fallback to direct API calls works)
2. Mock data warnings (expected in MVP stage)
3. API configuration warnings (non-critical)

## Visual Issues

### Layout Issues:
- All tabs appear to be rendering correctly in terms of layout
- Tab buttons are properly styled and positioned
- Content areas are properly sized

### Content Issues:
- Multiple tabs showing duplicate content (Business Analytics)
- Risk Indicators tab showing error message instead of content
- Contact tab showing minimal content

## Next Steps

1. **Investigate Tab Content Switching Logic**
   - Review tab switching JavaScript code
   - Verify tab content container visibility management
   - Check for proper `active` class toggling

2. **Fix Risk Indicators Tab API Issues**
   - Investigate CORS configuration in API Gateway
   - Verify risk assessment API endpoint exists
   - Check fallback data mechanism

3. **Fix Contact Tab Content Population**
   - Review contact tab template
   - Verify data mapping from merchant data to contact fields
   - Ensure all contact fields are displayed

4. **Fix API Endpoint Issues**
   - Verify API Gateway routing configuration
   - Check endpoint existence and responses
   - Ensure proper JSON responses

5. **Fix Export Button Module Loading**
   - Verify export button module file exists
   - Check file path and MIME type configuration

