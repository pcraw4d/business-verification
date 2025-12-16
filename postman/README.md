# KYB Platform API - Postman Collection

## Overview

This Postman collection provides comprehensive API testing for the KYB Platform, including all endpoints for merchant management, classification, risk assessment, business intelligence, and monitoring.

## Collection Details

- **Collection Name:** KYB Platform API
- **Collection ID:** `e3dfe3e8-2b90-46d0-837e-220cb93b3009`
- **Base URL:** `https://api-gateway-service-production-21fd.up.railway.app`

## Quick Start

### 1. Import Collection

1. Open Postman
2. Click **Import**
3. Select `kyb-platform-api-collection.json`
4. The collection will be imported with all endpoints configured

### 2. Configure Variables

The collection includes these variables:

| Variable       | Description              | Default Value                                                |
| -------------- | ------------------------ | ------------------------------------------------------------ |
| `baseUrl`      | API Gateway base URL     | `https://api-gateway-service-production-21fd.up.railway.app` |
| `authToken`    | JWT authentication token | (empty - set after login)                                    |
| `merchantId`   | Merchant ID for testing  | (empty)                                                      |
| `assessmentId` | Risk assessment ID       | (empty)                                                      |

### 3. Authenticate

1. Use **Authentication → Login** request
2. Enter your credentials:
   ```json
   {
     "email": "your-email@example.com",
     "password": "your-password"
   }
   ```
3. Copy the `token` from the response
4. Set it in collection variables:
   - Click collection → **Variables** tab
   - Paste token into `authToken` value

### 4. Start Testing

Begin with health checks (no authentication required):

- **Health & Status → API Gateway Health Check**

Then test authenticated endpoints:

- **Merchants → List Merchants**
- **Classification → Classify Business**

## Collection Structure

### Health & Status (4 endpoints)

- API Gateway Health Check
- Classification Service Health
- Merchant Service Health
- Risk Assessment Service Health

### Authentication (2 endpoints)

- Register User
- Login

### Classification (1 endpoint)

- Classify Business

### Merchants (11 endpoints)

- List Merchants (with pagination)
- Get Merchant by ID
- Create Merchant
- Update Merchant
- Delete Merchant
- Search Merchants
- Get Merchant Analytics
- Get Merchant Website Analysis
- Get Merchant Risk Score
- Get All Merchants Analytics
- Get Merchants Statistics

### Risk Assessment (5 endpoints)

- Assess Risk
- Get Risk Assessment Status
- Get Risk Benchmarks
- Get Risk Predictions
- Get Risk Indicators

### Business Intelligence (1 endpoint)

- Analyze Business Intelligence

### Monitoring & Analytics (5 endpoints)

- Get Analytics Trends
- Get Analytics Insights
- Get Monitoring Metrics
- Get Monitoring Health
- Get Monitoring Alerts

### Compliance (1 endpoint)

- Get Compliance Status

### Sessions (4 endpoints)

- Get Current Session
- Get Session Metrics
- Get Session Activity
- Get Session Status

### Dashboard (v3) (1 endpoint)

- Get Dashboard Metrics (v3)

## Authentication

All endpoints (except health checks) require Bearer token authentication. The collection is configured with:

- **Type:** Bearer Token
- **Token:** `{{authToken}}` (collection variable)

The token is automatically included in all requests once set in the collection variables.

## Example Workflows

### 1. Create and Analyze a Merchant

1. **Create Merchant** → Copy the returned `merchantId`
2. Set `merchantId` in collection variables
3. **Get Merchant by ID** → Verify merchant was created
4. **Get Merchant Analytics** → View analytics
5. **Assess Risk** → Start risk assessment
6. Copy `assessmentId` from response
7. **Get Risk Assessment Status** → Check assessment progress

### 2. Classify a Business

1. **Classify Business** → Submit business information
2. Review returned industry codes (NAICS, SIC, MCC)
3. Use classification data to create a merchant

### 3. Monitor System Health

1. **API Gateway Health Check** → Verify gateway is running
2. **Classification Service Health** → Check classification service
3. **Merchant Service Health** → Check merchant service
4. **Risk Assessment Service Health** → Check risk service
5. **Get Monitoring Metrics** → View system metrics

## Troubleshooting

### 401 Unauthorized

- Ensure `authToken` is set in collection variables
- Token may have expired - login again to get a new token

### 404 Not Found

- Verify the endpoint URL is correct
- Check that `baseUrl` variable is set correctly
- Ensure the service is running

### 500 Internal Server Error

- Check service health endpoints
- Review server logs
- Verify request body format matches API expectations

## Environment Variables

For different environments, you can create Postman environments:

### Production

```
baseUrl: https://api-gateway-service-production-21fd.up.railway.app
```

### Staging (if available)

```
baseUrl: https://api-gateway-service-staging.up.railway.app
```

### Local Development

```
baseUrl: http://localhost:8080
```

## API Documentation

For detailed API documentation, see:

- `/docs/API_DOCUMENTATION.md`
- `/docs/developer-guides/api-development.md`
- `/docs/PRODUCTION_URLS_REFERENCE.md`

## Support

For issues or questions:

1. Check the API documentation
2. Review service health endpoints
3. Check Railway deployment status
4. Review server logs

## Collection Maintenance

To update the collection:

1. Export current collection from Postman
2. Make changes to the JSON file
3. Re-import into Postman

Or use the Postman API to programmatically update the collection.
