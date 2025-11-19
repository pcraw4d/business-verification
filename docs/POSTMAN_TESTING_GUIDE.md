# Postman Testing Guide

**Date**: 2025-11-18  
**Status**: Ready for Execution

---

## Postman Collection Created

A comprehensive Postman collection has been created at:
**`postman/KYB_Route_Testing_Collection.json`**

This collection includes all critical route tests from the testing plan.

---

## Importing the Collection

### Recommended: Import via Postman UI

1. Open Postman (desktop app or web)
2. Click **Import** button (top left)
3. Select **File** tab
4. Navigate to and select: `postman/KYB_Route_Testing_Collection.json`
5. Click **Import**

The collection will be imported with:
- ✅ All 11+ requests fully configured
- ✅ HTTP methods, URLs, headers, and request bodies
- ✅ Automated test scripts for each request
- ✅ Organized into folders by testing phase
- ✅ Collection variable for base URL

**Note**: A basic collection structure has already been created in your Postman account via MCP server (Collection ID: `efbf4531-5a02-4d09-8364-141b8622ca08`). You can either:
- Import the JSON file to create a new complete collection, OR
- Delete the existing basic collection and import the JSON file

---

## Collection Structure

The collection is organized into folders:

1. **Phase 3.1: Authentication Routes**
   - Register - Valid
   - Register - Missing Fields
   - Register - Invalid Email
   - Login - Valid
   - Login - Invalid Credentials
   - Login - Missing Fields

2. **Phase 3.2: UUID Validation**
   - Risk Indicators - Invalid UUID
   - Risk Indicators - Edge Case (indicators)
   - Risk Indicators - Valid UUID

3. **Phase 3.3: CORS Configuration**
   - CORS Preflight - Auth Register

4. **Phase 6: Error Handling**
   - 404 Handler - Invalid Route

5. **Health Checks**
   - API Gateway Health

---

## Running Tests

### Run Individual Tests

1. Select a request from the collection
2. Click **Send**
3. Review response in the **Response** tab
4. Check **Test Results** tab for automated test results

### Run Collection

1. Right-click on collection name
2. Select **Run collection**
3. Review test results in the **Test Results** tab

### Run with Newman (CLI)

```bash
# Install Newman
npm install -g newman

# Run collection
newman run postman/KYB_Route_Testing_Collection.json

# Run with HTML report
newman run postman/KYB_Route_Testing_Collection.json -r html --reporter-html-export report.html
```

---

## Automated Tests

Each request includes automated tests that verify:

- **Status Codes**: Correct HTTP status codes
- **Response Structure**: Expected JSON structure
- **Error Messages**: Appropriate error messages
- **CORS Headers**: Correct CORS configuration

Test results appear in the **Test Results** tab after sending requests.

---

## Expected Results

### Authentication Routes

- **Register - Valid**: 200/201 with user info
- **Register - Missing Fields**: 400 with error message
- **Register - Invalid Email**: 400 with error message
- **Login - Valid**: 200 with token and user info
- **Login - Invalid Credentials**: 401 Unauthorized
- **Login - Missing Fields**: 400 with error message

### UUID Validation

- **Invalid UUID**: 400 with "Invalid merchant ID format" error
- **Edge Case (indicators)**: 400 Bad Request
- **Valid UUID**: 200 OK (or appropriate response)

### CORS Configuration

- **Preflight Request**: 200 OK with CORS headers
- **Allow-Origin**: Specific origin (not wildcard)
- **Allow-Credentials**: true

### Error Handling

- **404 Handler**: 404 with helpful error structure
- **Error Code**: "NOT_FOUND"
- **Suggestions**: Available endpoints listed

---

## Documenting Results

After running tests:

1. **Export Results**: 
   - Right-click collection → **Export**
   - Save results for documentation

2. **Update Test Results Document**:
   - Open `docs/MANUAL_TEST_RESULTS.md`
   - Record pass/fail status for each test
   - Note any issues or observations

3. **Screenshot Responses**:
   - Take screenshots of failed tests
   - Include in documentation

---

## Troubleshooting

### Tests Failing

1. **Check Service Status**: Verify API Gateway is healthy
2. **Check Environment Variables**: Ensure Railway config is correct
3. **Check Response**: Review actual response vs. expected
4. **Check Railway Logs**: Look for errors in service logs

### CORS Tests Failing

1. **Verify Origin**: Check `CORS_ALLOWED_ORIGINS` in Railway
2. **Check Headers**: Verify CORS headers in response
3. **Test from Browser**: Use browser DevTools to verify

### Authentication Tests Failing

1. **Check Supabase**: Verify Supabase is configured
2. **Check Credentials**: Ensure test credentials are valid
3. **Check Response**: Review error messages

---

## Next Steps

1. **Import Collection**: Import `postman/KYB_Route_Testing_Collection.json`
2. **Run Tests**: Execute all tests in the collection
3. **Document Results**: Update `docs/MANUAL_TEST_RESULTS.md`
4. **Fix Issues**: Address any failures
5. **Retest**: Run tests again after fixes

---

**Collection Location**: `postman/KYB_Route_Testing_Collection.json`  
**Last Updated**: 2025-11-18

