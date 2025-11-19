# Postman Collection Import Instructions

**Date**: 2025-11-18  
**Collection File**: `postman/KYB_Route_Testing_Collection.json`

---

## Quick Import Steps

1. **Open Postman**
   - Desktop app or web version (https://web.postman.co)

2. **Click Import**
   - Top left corner of Postman interface
   - Or use keyboard shortcut: `Ctrl+I` (Windows/Linux) or `Cmd+I` (Mac)

3. **Select File**
   - Click **File** tab
   - Click **Upload Files** or drag and drop
   - Navigate to: `postman/KYB_Route_Testing_Collection.json`
   - Select the file

4. **Import**
   - Click **Import** button
   - Collection will appear in your Postman sidebar

---

## What You'll Get

After import, you'll have a complete collection with:

### Collection Structure
- **Name**: KYB Route Testing Collection
- **Base URL Variable**: `{{base_url}}` = `https://api-gateway-service-production-21fd.up.railway.app`

### Folders and Requests

#### Phase 3.1: Authentication Routes (6 requests)
- ✅ Register - Valid
- ✅ Register - Missing Fields
- ✅ Register - Invalid Email
- ✅ Login - Valid
- ✅ Login - Invalid Credentials
- ✅ Login - Missing Fields

#### Phase 3.2: UUID Validation (3 requests)
- ✅ Risk Indicators - Invalid UUID
- ✅ Risk Indicators - Edge Case (indicators)
- ✅ Risk Indicators - Valid UUID

#### Phase 3.3: CORS Configuration (1 request)
- ✅ CORS Preflight - Auth Register

#### Phase 6: Error Handling (1 request)
- ✅ 404 Handler - Invalid Route

#### Health Checks (1 request)
- ✅ API Gateway Health

**Total: 12 requests** with automated tests

---

## Each Request Includes

✅ **HTTP Method** (GET, POST, OPTIONS)  
✅ **Complete URL** (using base_url variable)  
✅ **Headers** (Content-Type, Origin, etc.)  
✅ **Request Body** (for POST requests)  
✅ **Automated Test Scripts** (validates status codes, response structure, error messages)

---

## Running Tests

### Run Individual Request
1. Select a request from the collection
2. Click **Send**
3. Check **Test Results** tab for automated test results

### Run Entire Collection
1. Right-click collection name
2. Select **Run collection**
3. Review test results in the **Test Results** tab

### Run Specific Folder
1. Right-click folder name
2. Select **Run folder**
3. All requests in that folder will execute

---

## Test Scripts

Each request includes automated tests that verify:

- **Status Codes**: Correct HTTP status (200, 201, 400, 401, 404)
- **Response Structure**: Expected JSON structure
- **Error Messages**: Appropriate error messages
- **CORS Headers**: Correct CORS configuration
- **UUID Validation**: Proper UUID format validation

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

## Troubleshooting

### Import Fails
- **Check file path**: Ensure `postman/KYB_Route_Testing_Collection.json` exists
- **Check JSON format**: File should be valid JSON
- **Try again**: Sometimes Postman needs a refresh

### Tests Failing
- **Check Service Status**: Verify API Gateway is healthy
- **Check Environment Variables**: Ensure Railway config is correct
- **Check Response**: Review actual response vs. expected
- **Check Railway Logs**: Look for errors in service logs

### CORS Tests Failing
- **Verify Origin**: Check `CORS_ALLOWED_ORIGINS` in Railway
- **Check Headers**: Verify CORS headers in response
- **Test from Browser**: Use browser DevTools to verify

### Authentication Tests Failing
- **Check Supabase**: Verify Supabase is configured
- **Check Credentials**: Ensure test credentials are valid
- **Check Response**: Review error messages

---

## Next Steps After Import

1. ✅ **Verify Collection**: Check that all 12 requests are present
2. ✅ **Run Health Check**: Test API Gateway health endpoint first
3. ✅ **Run Authentication Tests**: Test register and login endpoints
4. ✅ **Run UUID Validation Tests**: Verify UUID validation works
5. ✅ **Run CORS Tests**: Verify CORS configuration
6. ✅ **Run Error Handling Tests**: Verify 404 handler
7. ✅ **Document Results**: Record pass/fail status for each test
8. ✅ **Fix Issues**: Address any failures
9. ✅ **Retest**: Run tests again after fixes

---

## Collection File Location

**Path**: `postman/KYB_Route_Testing_Collection.json`  
**Format**: Postman Collection v2.1.0  
**Size**: ~15 KB  
**Last Updated**: 2025-11-18

---

**Ready to Import!** 🚀

