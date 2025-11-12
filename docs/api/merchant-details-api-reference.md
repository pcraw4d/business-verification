# Merchant Details API Reference

**Version:** 1.0.0  
**Last Updated:** December 19, 2024  
**Base URL:** `/api/v1`

## Overview

The Merchant Details API provides endpoints for retrieving comprehensive analytics, risk assessment, and website analysis data for merchants. All endpoints require Bearer token authentication.

## Authentication

All endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <your_token>
```

To obtain a token:
1. Login to the application via the login endpoint
2. Check browser DevTools > Application > Session Storage > `authToken`
3. Or use the authentication API to obtain a token

## Base Merchant Data

### GET /merchants/{merchantId}

Retrieve base merchant information.

**Parameters:**
- `merchantId` (path, required): Unique identifier for the merchant

**Response:**
```json
{
  "id": "merchant-123",
  "businessName": "Acme Corporation",
  "industry": "Technology",
  "status": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

## Business Analytics Endpoints

### GET /merchants/{merchantId}/analytics

Retrieve comprehensive analytics data for a merchant including classification, security, quality, and intelligence metrics.

**Parameters:**
- `merchantId` (path, required): Unique identifier for the merchant

**Response:**
```json
{
  "merchantId": "merchant-123",
  "classification": {
    "primaryIndustry": "Technology",
    "confidenceScore": 0.95,
    "riskLevel": "low",
    "mccCodes": [
      {
        "code": "5734",
        "description": "Computer Software Stores",
        "confidence": 0.92
      }
    ],
    "sicCodes": [
      {
        "code": "7372",
        "description": "Prepackaged Software",
        "confidence": 0.88
      }
    ],
    "naicsCodes": [
      {
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.90
      }
    ]
  },
  "security": {
    "trustScore": 0.85,
    "sslValid": true,
    "sslExpiryDate": "2025-12-31T00:00:00Z",
    "securityHeaders": [
      {
        "header": "X-Frame-Options",
        "present": true,
        "value": "DENY"
      }
    ]
  },
  "quality": {
    "completenessScore": 0.92,
    "dataPoints": 45,
    "missingFields": []
  },
  "intelligence": {
    "businessAge": 5,
    "employeeCount": 150,
    "annualRevenue": 5000000
  },
  "timestamp": "2024-12-19T10:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Analytics data retrieved successfully
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Merchant not found
- `500 Internal Server Error`: Server error

### GET /merchants/{merchantId}/website-analysis

Retrieve website analysis data including SSL certificate status, security headers, performance metrics, and accessibility score.

**Parameters:**
- `merchantId` (path, required): Unique identifier for the merchant

**Response:**
```json
{
  "merchantId": "merchant-123",
  "websiteUrl": "https://www.example.com",
  "ssl": {
    "valid": true,
    "expiryDate": "2025-12-31T00:00:00Z",
    "issuer": "Let's Encrypt",
    "grade": "A"
  },
  "securityHeaders": [
    {
      "name": "X-Frame-Options",
      "present": true,
      "value": "DENY"
    },
    {
      "name": "X-Content-Type-Options",
      "present": true,
      "value": "nosniff"
    },
    {
      "name": "Strict-Transport-Security",
      "present": true,
      "value": "max-age=31536000"
    }
  ],
  "performance": {
    "loadTime": 1.2,
    "pageSize": 1024000,
    "requests": 45,
    "score": 85
  },
  "accessibility": {
    "score": 0.92,
    "issues": []
  },
  "lastAnalyzed": "2024-12-19T10:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Website analysis data retrieved successfully
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Merchant not found
- `500 Internal Server Error`: Server error

## Risk Assessment Endpoints

### POST /risk/assess

Trigger a risk assessment for a merchant. This endpoint returns immediately with a 202 Accepted status and an assessment ID. The assessment is processed asynchronously.

**Request Body:**
```json
{
  "merchantId": "merchant-123",
  "options": {
    "includeHistory": true,
    "includePredictions": true
  }
}
```

**Parameters:**
- `merchantId` (body, required): Unique identifier for the merchant
- `options.includeHistory` (body, optional): Include historical risk data (default: false)
- `options.includePredictions` (body, optional): Include risk predictions (default: false)

**Response (202 Accepted):**
```json
{
  "assessmentId": "assess-456",
  "status": "pending",
  "estimatedCompletion": "2024-12-19T10:05:00Z"
}
```

**Status Codes:**
- `202 Accepted`: Assessment started successfully
- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Server error

**Usage:**
After receiving the assessment ID, poll the GET endpoint to check status until completion.

### GET /risk/assess/{assessmentId}

Get the status and results of a risk assessment by assessment ID. Poll this endpoint until status is "completed" to get the final results.

**Parameters:**
- `assessmentId` (path, required): Unique identifier for the assessment

**Response (Pending/Processing):**
```json
{
  "assessmentId": "assess-456",
  "merchantId": "merchant-123",
  "status": "processing",
  "progress": 65,
  "estimatedCompletion": "2024-12-19T10:05:00Z"
}
```

**Response (Completed):**
```json
{
  "assessmentId": "assess-456",
  "merchantId": "merchant-123",
  "status": "completed",
  "progress": 100,
  "result": {
    "overallScore": 0.75,
    "riskLevel": "medium",
    "factors": [
      {
        "name": "Financial Stability",
        "score": 0.8,
        "weight": 0.3
      },
      {
        "name": "Business History",
        "score": 0.7,
        "weight": 0.25
      }
    ]
  },
  "completedAt": "2024-12-19T10:04:30Z"
}
```

**Status Values:**
- `pending`: Assessment queued but not started
- `processing`: Assessment in progress
- `completed`: Assessment completed successfully
- `failed`: Assessment failed

**Status Codes:**
- `200 OK`: Assessment status retrieved successfully
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Assessment not found
- `500 Internal Server Error`: Server error

### GET /merchants/{merchantId}/risk-score

Retrieve the current risk score for a merchant.

**Parameters:**
- `merchantId` (path, required): Unique identifier for the merchant

**Response:**
```json
{
  "merchantId": "merchant-123",
  "overallScore": 0.75,
  "riskLevel": "medium",
  "factors": [
    {
      "name": "Financial Stability",
      "score": 0.8,
      "weight": 0.3
    },
    {
      "name": "Business History",
      "score": 0.7,
      "weight": 0.25
    }
  ],
  "lastUpdated": "2024-12-19T10:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Risk score retrieved successfully
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Merchant not found
- `500 Internal Server Error`: Server error

### GET /merchants/{merchantId}/website-risk

Retrieve website risk assessment data for a merchant.

**Parameters:**
- `merchantId` (path, required): Unique identifier for the merchant

**Response:**
```json
{
  "merchantId": "merchant-123",
  "websiteUrl": "https://www.example.com",
  "riskScore": 0.65,
  "indicators": [
    {
      "type": "ssl",
      "status": "valid",
      "score": 0.9
    },
    {
      "type": "security_headers",
      "status": "good",
      "score": 0.85
    }
  ],
  "lastAnalyzed": "2024-12-19T10:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Website risk data retrieved successfully
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Merchant not found
- `500 Internal Server Error`: Server error

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "bad_request",
  "message": "Invalid request body",
  "details": {
    "field": "merchantId",
    "reason": "merchantId is required"
  }
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Authentication required"
}
```

### 404 Not Found
```json
{
  "error": "not_found",
  "message": "Merchant not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal_error",
  "message": "An internal error occurred"
}
```

## Rate Limiting

API requests are subject to rate limiting. Rate limit information is provided in response headers:

- `X-RateLimit-Limit`: Maximum number of requests allowed
- `X-RateLimit-Remaining`: Number of requests remaining in current window
- `X-RateLimit-Reset`: Timestamp when rate limit resets

When rate limit is exceeded, a `429 Too Many Requests` response is returned.

## Examples

### Example: Get Merchant Analytics

```bash
curl -X GET "https://api.kyb-platform.com/api/v1/merchants/merchant-123/analytics" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### Example: Start Risk Assessment

```bash
curl -X POST "https://api.kyb-platform.com/api/v1/risk/assess" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": "merchant-123",
    "options": {
      "includeHistory": true,
      "includePredictions": true
    }
  }'
```

### Example: Poll Assessment Status

```bash
# After receiving assessmentId from POST /risk/assess
curl -X GET "https://api.kyb-platform.com/api/v1/risk/assess/assess-456" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

## OpenAPI Specification

A complete OpenAPI 3.0 specification is available at:
- File: `api/openapi/merchant-details-api-spec.yaml`
- Can be imported into API testing tools like Postman or Insomnia

## Support

For API support, contact:
- Email: api-support@kyb-platform.com
- Documentation: See `tests/api/merchant-details/README.md` for testing setup

