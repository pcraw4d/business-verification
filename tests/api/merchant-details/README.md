# Merchant Details API Testing Environment

This directory contains API testing resources for the merchant-details page endpoints.

## Contents

- `postman-collection.json` - Postman collection with all merchant-details endpoints
- `insomnia-collection.json` - Insomnia collection (alternative to Postman)
- `.env.example` - Environment variables template
- `README.md` - This file

## Setup

### 1. Environment Variables

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with your actual values:
- `BASE_URL` - Your API base URL (default: http://localhost:8080)
- `AUTH_TOKEN` - Your Bearer token for authentication
- `TEST_MERCHANT_ID` - A test merchant ID to use for testing

### 2. Postman Setup

1. Open Postman
2. Click **Import** button
3. Select `postman-collection.json`
4. Go to **Environments** and create a new environment
5. Add variables:
   - `baseUrl` - Your API base URL
   - `merchantId` - Test merchant ID
   - `authToken` - Your Bearer token
6. Select your environment from the dropdown

### 3. Insomnia Setup

1. Open Insomnia
2. Go to **Application** > **Preferences** > **Data** > **Import Data**
3. Select `insomnia-collection.json`
4. The workspace and environment will be imported automatically
5. Update environment variables in **Manage Environments**

## Getting an Authentication Token

### Method 1: From Browser Session Storage

1. Open the application in your browser
2. Open DevTools (F12)
3. Go to **Application** tab > **Session Storage**
4. Find `authToken` and copy its value

### Method 2: From Login Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "your@email.com", "password": "yourpassword"}'
```

The response will contain a token in the `token` field.

## Available Endpoints

### Business Analytics

- **GET** `/api/v1/merchants/{merchantId}/analytics`
  - Retrieve comprehensive analytics data for a merchant
  - Returns: Classification, security, quality, and intelligence metrics

- **GET** `/api/v1/merchants/{merchantId}/website-analysis`
  - Retrieve website analysis data
  - Returns: SSL status, security headers, performance metrics, accessibility score

### Risk Assessment

- **POST** `/api/v1/risk/assess`
  - Trigger risk assessment for a merchant
  - Returns: 202 Accepted with assessment ID
  - Body: `{"merchantId": "string", "options": {...}}`

- **GET** `/api/v1/risk/assess/{assessmentId}`
  - Get assessment status
  - Returns: Assessment status and results (when complete)

- **GET** `/api/v1/merchants/{merchantId}/risk-score`
  - Get current risk score for a merchant
  - Returns: Risk score, level, factors, last updated

- **GET** `/api/v1/merchants/{merchantId}/website-risk`
  - Get website risk assessment
  - Returns: Website risk score, indicators, last analyzed

## Testing Workflow

### 1. Test Merchant Analytics

```bash
# Get merchant analytics
curl -X GET "http://localhost:8080/api/v1/merchants/test-merchant-123/analytics" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get website analysis
curl -X GET "http://localhost:8080/api/v1/merchants/test-merchant-123/website-analysis" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. Test Risk Assessment (Async Flow)

```bash
# Step 1: Start assessment
curl -X POST "http://localhost:8080/api/v1/risk/assess" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": "test-merchant-123",
    "options": {
      "includeHistory": true,
      "includePredictions": true
    }
  }'

# Response: {"assessmentId": "assess-123", "status": "pending"}

# Step 2: Check status (poll until complete)
curl -X GET "http://localhost:8080/api/v1/risk/assess/assess-123" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Test Scripts

### Using curl

See examples in the Testing Workflow section above.

### Using httpie

```bash
# Install httpie: pip install httpie

# Get analytics
http GET localhost:8080/api/v1/merchants/test-merchant-123/analytics \
  Authorization:"Bearer YOUR_TOKEN"

# Start risk assessment
http POST localhost:8080/api/v1/risk/assess \
  Authorization:"Bearer YOUR_TOKEN" \
  merchantId=test-merchant-123 \
  options:='{"includeHistory":true}'
```

## Troubleshooting

### 401 Unauthorized

- Check that your `AUTH_TOKEN` is valid and not expired
- Ensure the token is prefixed with "Bearer " in the Authorization header
- Verify the token has the required permissions

### 404 Not Found

- Verify the `merchantId` exists in the database
- Check that the endpoint URL is correct
- Ensure the API version (`v1`) is correct

### 500 Internal Server Error

- Check server logs for detailed error messages
- Verify database connectivity
- Ensure all required services are running

## Notes

- All endpoints require authentication via Bearer token
- Rate limiting may apply - check response headers for rate limit information
- Some endpoints return async results - use polling for status checks
- Test data should be cleaned up after testing to avoid polluting the database

