# Postman Test Scripts Guide

## Overview

The KYB Platform API Postman collection now includes test scripts for automated testing and validation. This guide explains how the tests work and how to use them.

## Test Scripts Added

### Collection-Level Tests

- **Pre-request Script**: Validates that `baseUrl` is set
- **Test Script**: Basic response time monitoring and error logging

### Request-Level Tests

#### Health Check Endpoints

- ✅ Status code validation (200)
- ✅ Response time check (< 2000ms)
- ✅ JSON validation
- ✅ Status field validation

#### Authentication Endpoints

**Login Request:**

- ✅ Status code validation (200 or 201)
- ✅ Response time check (< 3000ms)
- ✅ JSON validation
- ✅ **Auto-extracts token** and saves to `authToken` variable
- ✅ Token validation (length, type)

**Register Request:**

- Basic validation tests (add as needed)

#### Classification Endpoints

- ✅ Status code validation (200)
- ✅ Response time check (< 5000ms)
- ✅ JSON validation
- ✅ Response structure validation

#### Merchant Endpoints

**List Merchants:**

- ✅ Status code validation (200)
- ✅ Response time check (< 3000ms)
- ✅ JSON validation
- ✅ Merchants array validation
- ✅ **Auto-extracts first merchant ID** to `merchantId` variable

**Create Merchant:**

- ✅ Status code validation (200 or 201)
- ✅ Response time check (< 3000ms)
- ✅ JSON validation
- ✅ **Auto-extracts merchant ID** to `merchantId` variable

**Other Merchant Endpoints:**

- Basic validation (can be enhanced)

#### Risk Assessment Endpoints

**Assess Risk:**

- ✅ Status code validation (200, 201, or 202)
- ✅ Response time check (< 5000ms)
- ✅ JSON validation
- ✅ **Auto-extracts assessment ID** to `assessmentId` variable

## How to Use

### 1. Re-import the Collection

The updated collection JSON file now includes test scripts. Re-import it:

1. In Postman, click **Import**
2. Select `kyb-platform-api-collection.json`
3. Choose **Replace** when prompted (to update existing collection)
4. All test scripts will be added

### 2. Run Individual Tests

1. Select any request
2. Click **Send**
3. Go to the **Test Results** tab
4. You'll see test results like:
   - ✅ Status code is 200
   - ✅ Response time is less than 3000ms
   - ✅ Response has valid JSON

### 3. Run Collection Tests

1. Click on the collection name
2. Click **Run** (play button)
3. Select which requests to run
4. Click **Run KYB Platform API**
5. View test results for all requests

### 4. Automatic Variable Extraction

The collection automatically extracts and sets variables:

- **Login** → Extracts `token` → Sets `authToken`
- **List Merchants** → Extracts first merchant `id` → Sets `merchantId`
- **Create Merchant** → Extracts merchant `id` → Sets `merchantId`
- **Assess Risk** → Extracts `assessmentId` → Sets `assessmentId`

These variables are automatically used in subsequent requests!

## Test Script Examples

### Basic Test Template

```javascript
pm.test("Status code is 200", function () {
  pm.response.to.have.status(200);
});

pm.test("Response time is less than 3000ms", function () {
  pm.expect(pm.response.responseTime).to.be.below(3000);
});

pm.test("Response has valid JSON", function () {
  pm.response.to.be.json;
});
```

### Token Extraction Example

```javascript
if (pm.response.code === 200 || pm.response.code === 201) {
  const jsonData = pm.response.json();
  let token = jsonData.token || jsonData.access_token || jsonData.data?.token;

  if (token) {
    pm.collectionVariables.set("authToken", token);
    console.log("Token extracted and saved");
  }
}
```

### ID Extraction Example

```javascript
if (pm.response.code === 200 || pm.response.code === 201) {
  const jsonData = pm.response.json();
  const id = jsonData.id || jsonData.data?.id;

  if (id) {
    pm.collectionVariables.set("merchantId", id);
    console.log("ID extracted:", id);
  }
}
```

## Adding More Tests

To add tests to other requests:

1. Select the request in Postman
2. Go to the **Tests** tab
3. Add test scripts using the templates above
4. Save the request

Or update the JSON file and re-import.

## Test Results Interpretation

### ✅ Passing Tests

- Green checkmark = Test passed
- All assertions met

### ❌ Failing Tests

- Red X = Test failed
- Check the error message
- Verify API response format matches expectations

### ⚠️ Warnings

- Yellow warning = Non-critical issue
- Example: Slow response time (> 10 seconds)

## Troubleshooting

### "No tests found"

- **Solution**: Re-import the updated collection JSON file
- The test scripts are in the JSON file

### Tests failing

- Check if API response format matches expected structure
- Verify status codes match expectations
- Check response time thresholds

### Variables not being set

- Check console logs for extraction messages
- Verify response contains the expected field names
- Token/ID field names may vary - check actual response

## Next Steps

1. **Re-import the collection** with test scripts
2. **Run the Login request** to auto-extract token
3. **Run other requests** to see test results
4. **Customize tests** for your specific needs

## Customization

You can customize tests by:

- Adjusting response time thresholds
- Adding more validation checks
- Extracting additional data
- Adding conditional logic based on response

The test scripts are in the `event` array of each request in the JSON file.
