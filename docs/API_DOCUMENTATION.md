# KYB Platform API Documentation

**Version**: 1.0  
**Base URL**: `https://api-gateway-production.up.railway.app`  
**API Version**: v1  
**Last Updated**: 2025-01-27

---

## Table of Contents

1. [Authentication](#authentication)
2. [Error Handling](#error-handling)
3. [API Endpoints](#api-endpoints)
   - [Health Checks](#health-checks)
   - [Classification](#classification)
   - [Merchants](#merchants)
   - [Risk Assessment](#risk-assessment)
   - [Business Intelligence](#business-intelligence)
   - [Authentication](#authentication-endpoints)
4. [Rate Limiting](#rate-limiting)
5. [CORS](#cors)

---

## Authentication

### JWT Token Authentication

Most endpoints require authentication via JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Public Endpoints

The following endpoints do not require authentication:
- `GET /health`
- `GET /`
- `POST /api/v1/classify` (public classification endpoint)

---

## Error Handling

### Standard Error Response Format

All errors follow a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details (optional)",
    "field": "field_name (optional, for validation errors)",
    "validation": [
      {
        "field": "field_name",
        "message": "Validation error message",
        "code": "VALIDATION_ERROR"
      }
    ]
  },
  "request_id": "request-id-if-available",
  "timestamp": "2025-01-27T12:00:00Z",
  "path": "/api/v1/endpoint",
  "method": "POST"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request format or parameters |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Authentication required or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate) |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |
| `NOT_IMPLEMENTED` | 501 | Feature not yet implemented |

### Example Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "validation": [
      {
        "field": "business_name",
        "message": "business_name is required",
        "code": "VALIDATION_ERROR"
      }
    ]
  },
  "request_id": "req_1234567890",
  "timestamp": "2025-01-27T12:00:00Z",
  "path": "/api/v1/classify",
  "method": "POST"
}
```

---

## API Endpoints

### Health Checks

#### GET /health

Check API Gateway health status.

**Authentication**: Not required

**Query Parameters**:
- `detailed` (boolean, optional): Include detailed health information
- `counts` (boolean, optional): Include table counts (requires `detailed=true`)

**Response**: `200 OK`

```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "timestamp": "2025-01-27T12:00:00Z",
  "environment": "production",
  "features": {
    "supabase_integration": true,
    "authentication": true,
    "rate_limiting": true,
    "cors_enabled": true
  },
  "response_time_ms": 45
}
```

**Detailed Response** (with `?detailed=true`):
```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "timestamp": "2025-01-27T12:00:00Z",
  "environment": "production",
  "supabase_status": {
    "connected": true,
    "url": "https://your-project.supabase.co"
  },
  "table_counts": {
    "classifications": 1234,
    "merchants": 567,
    "risk_keywords": 890,
    "business_risk_assessments": 234
  },
  "response_time_ms": 120
}
```

#### GET /api/v1/classification/health

Check Classification Service health.

**Authentication**: Not required

**Response**: `200 OK`

#### GET /api/v1/merchant/health

Check Merchant Service health.

**Authentication**: Not required

**Response**: `200 OK`

#### GET /api/v1/risk/health

Check Risk Assessment Service health.

**Authentication**: Not required

**Response**: `200 OK`

---

### Classification

#### POST /api/v1/classify

Classify a business and determine industry codes.

**Authentication**: Not required (public endpoint)

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "description": "Technology solutions provider",
  "website_url": "https://acme.com",
  "request_id": "req_1234567890"
}
```

**Request Fields**:
- `business_name` (string, required): Business name
- `description` (string, optional): Business description
- `website_url` (string, optional): Business website URL
- `request_id` (string, optional): Custom request ID

**Response**: `200 OK`

```json
{
  "request_id": "req_1234567890",
  "business_name": "Acme Corporation",
  "description": "Technology solutions provider",
  "classification": {
    "industry": "Technology",
    "mcc_codes": ["5734", "7372"],
    "sic_codes": ["7372", "7379"],
    "naics_codes": ["541511", "541512"],
    "website_content": {
      "scraped": true,
      "content_length": 50000,
      "keywords_found": 25
    }
  },
  "risk_assessment": {
    "risk_level": "low",
    "risk_score": 0.15,
    "risk_factors": []
  },
  "verification_status": {
    "status": "verified",
    "processing_time": "2.5s"
  },
  "confidence_score": 0.95,
  "data_source": "smart_crawling_classification_service",
  "status": "success",
  "success": true,
  "timestamp": "2025-01-27T12:00:00Z",
  "processing_time": "2.5s"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields
- `500 Internal Server Error`: Classification service error

**Caching**: Responses are cached for 5 minutes. Subsequent requests with same parameters return cached results.

**Headers**:
- `X-Cache`: `HIT` or `MISS`
- `Cache-Control`: `public, max-age=300`

---

### Merchants

#### GET /api/v1/merchants

List merchants with pagination, filtering, and sorting.

**Authentication**: Required

**Query Parameters**:
- `page` (integer, optional, default: 1): Page number
- `page_size` (integer, optional, default: 20, max: 100): Items per page
- `portfolio_type` (string, optional): Filter by portfolio type
- `risk_level` (string, optional): Filter by risk level
- `status` (string, optional): Filter by status
- `search` (string, optional): Search in name and legal_name
- `sort_by` (string, optional, default: "created_at"): Sort field (name, legal_name, created_at, updated_at, portfolio_type, risk_level, status)
- `sort_order` (string, optional, default: "desc"): Sort order (asc, desc)

**Response**: `200 OK`

```json
{
  "merchants": [
    {
      "id": "merchant_1234567890",
      "name": "Acme Corporation",
      "legal_name": "Acme Corporation Inc.",
      "registration_number": "REG123456",
      "tax_id": "TAX123456",
      "industry": "Technology",
      "industry_code": "7372",
      "business_type": "Corporation",
      "founded_date": "2020-01-15T00:00:00Z",
      "employee_count": 150,
      "annual_revenue": 5000000.00,
      "address": {
        "street": "123 Main St",
        "city": "San Francisco",
        "state": "CA",
        "zip": "94105",
        "country": "USA"
      },
      "contact_info": {
        "email": "contact@acme.com",
        "phone": "+1-555-123-4567",
        "website": "https://acme.com"
      },
      "portfolio_type": "enterprise",
      "risk_level": "low",
      "compliance_status": "compliant",
      "status": "active",
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-20T14:30:00Z",
      "created_by": "user_123"
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "total_pages": 8,
  "has_next": true,
  "has_previous": false
}
```

**Error Responses**:
- `401 Unauthorized`: Missing or invalid authentication token
- `400 Bad Request`: Invalid query parameters

#### POST /api/v1/merchants

Create a new merchant.

**Authentication**: Required

**Request Body**:
```json
{
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation Inc.",
  "registration_number": "REG123456",
  "tax_id": "TAX123456",
  "industry": "Technology",
  "industry_code": "7372",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 150,
  "annual_revenue": 5000000.00,
  "address": {
    "street": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "zip": "94105",
    "country": "USA"
  },
  "contact_info": {
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "website": "https://acme.com"
  },
  "portfolio_type": "enterprise",
  "risk_level": "low",
  "compliance_status": "compliant",
  "status": "active"
}
```

**Request Fields**:
- `name` (string, required): Merchant name
- `legal_name` (string, required): Legal business name
- `registration_number` (string, optional): Business registration number
- `tax_id` (string, optional): Tax identification number
- `industry` (string, optional): Industry
- `industry_code` (string, optional): Industry code
- `business_type` (string, optional): Business type
- `founded_date` (string, optional): ISO 8601 date
- `employee_count` (integer, optional): Number of employees
- `annual_revenue` (number, optional): Annual revenue
- `address` (object, optional): Address object
- `contact_info` (object, optional): Contact information
- `portfolio_type` (string, optional): Portfolio type
- `risk_level` (string, optional): Risk level
- `compliance_status` (string, optional): Compliance status
- `status` (string, optional): Status

**Response**: `201 Created`

```json
{
  "id": "merchant_1234567890",
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation Inc.",
  "created_at": "2025-01-27T12:00:00Z",
  "updated_at": "2025-01-27T12:00:00Z",
  "created_by": "user_123"
}
```

**Error Responses**:
- `400 Bad Request`: Missing required fields or validation error
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Server error

#### GET /api/v1/merchants/{id}

Get a specific merchant by ID.

**Authentication**: Required

**Path Parameters**:
- `id` (string, required): Merchant ID

**Response**: `200 OK`

```json
{
  "id": "merchant_1234567890",
  "name": "Acme Corporation",
  "legal_name": "Acme Corporation Inc.",
  "registration_number": "REG123456",
  "tax_id": "TAX123456",
  "industry": "Technology",
  "industry_code": "7372",
  "business_type": "Corporation",
  "founded_date": "2020-01-15T00:00:00Z",
  "employee_count": 150,
  "annual_revenue": 5000000.00,
  "address": {
    "street": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "zip": "94105",
    "country": "USA"
  },
  "contact_info": {
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "website": "https://acme.com"
  },
  "portfolio_type": "enterprise",
  "risk_level": "low",
  "compliance_status": "compliant",
  "status": "active",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-20T14:30:00Z",
  "created_by": "user_123"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid merchant ID
- `401 Unauthorized`: Missing or invalid authentication token
- `404 Not Found`: Merchant not found
- `503 Service Unavailable`: Database unavailable

#### POST /api/v1/merchants/search

Search merchants with advanced criteria.

**Authentication**: Required

**Request Body**:
```json
{
  "query": "Acme",
  "filters": {
    "portfolio_type": "enterprise",
    "risk_level": "low",
    "status": "active"
  },
  "page": 1,
  "page_size": 20
}
```

**Response**: `200 OK` (same format as GET /api/v1/merchants)

#### GET /api/v1/merchants/analytics

Get merchant analytics and statistics.

**Authentication**: Required

**Response**: `200 OK`

```json
{
  "total_merchants": 150,
  "by_portfolio_type": {
    "enterprise": 50,
    "small_business": 75,
    "startup": 25
  },
  "by_risk_level": {
    "low": 100,
    "medium": 40,
    "high": 10
  },
  "by_status": {
    "active": 120,
    "inactive": 30
  }
}
```

---

### Risk Assessment

#### POST /api/v1/risk/assess

Perform a risk assessment for a business.

**Authentication**: Required

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "business_address": "123 Main St, San Francisco, CA 94105",
  "industry": "Technology",
  "country": "USA",
  "phone": "+1-555-123-4567",
  "email": "contact@acme.com",
  "website": "https://acme.com"
}
```

**Request Fields**:
- `business_name` (string, required): Business name
- `business_address` (string, required): Business address
- `industry` (string, required): Industry
- `country` (string, required): Country code
- `phone` (string, optional): Phone number
- `email` (string, optional): Email address
- `website` (string, optional): Website URL

**Response**: `200 OK`

```json
{
  "id": "assess_1234567890",
  "business_name": "Acme Corporation",
  "risk_score": 0.15,
  "risk_level": "low",
  "risk_factors": [],
  "recommendations": [
    "Continue monitoring",
    "Standard due diligence"
  ],
  "confidence": 0.95,
  "created_at": "2025-01-27T12:00:00Z",
  "processing_time": "3.5s"
}
```

**Error Responses**:
- `400 Bad Request`: Validation error
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Assessment processing error

#### GET /api/v1/risk/assess/{id}

Get a specific risk assessment by ID.

**Authentication**: Required

**Path Parameters**:
- `id` (string, required): Assessment ID

**Response**: `200 OK` (same format as POST response)

#### POST /api/v1/risk/assess/batch

Perform batch risk assessments.

**Authentication**: Required

**Request Body**:
```json
{
  "assessments": [
    {
      "business_name": "Acme Corporation",
      "business_address": "123 Main St, San Francisco, CA 94105",
      "industry": "Technology",
      "country": "USA"
    },
    {
      "business_name": "Beta Inc",
      "business_address": "456 Oak Ave, New York, NY 10001",
      "industry": "Retail",
      "country": "USA"
    }
  ]
}
```

**Response**: `200 OK`

```json
{
  "results": [
    {
      "id": "assess_1234567890",
      "business_name": "Acme Corporation",
      "risk_score": 0.15,
      "risk_level": "low",
      "status": "completed"
    },
    {
      "id": "assess_1234567891",
      "business_name": "Beta Inc",
      "risk_score": 0.45,
      "risk_level": "medium",
      "status": "completed"
    }
  ],
  "total": 2,
  "completed": 2,
  "failed": 0
}
```

#### GET /api/v1/risk/benchmarks

Get risk benchmarks for an industry.

**Authentication**: Required

**Query Parameters**:
- `mcc` (string, optional): MCC code
- `naics` (string, optional): NAICS code
- `sic` (string, optional): SIC code

**Note**: At least one of `mcc`, `naics`, or `sic` is required.

**Response**: `200 OK`

```json
{
  "industry": "Technology",
  "mcc": "7372",
  "benchmarks": {
    "average_risk_score": 0.25,
    "median_risk_score": 0.20,
    "p25_risk_score": 0.15,
    "p75_risk_score": 0.35,
    "sample_size": 1000
  },
  "updated_at": "2025-01-27T12:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Missing required parameters
- `503 Service Unavailable`: Feature not available in production (unless enabled)

#### GET /api/v1/risk/predictions/{merchant_id}

Get risk predictions for a merchant.

**Authentication**: Required

**Path Parameters**:
- `merchant_id` (string, required): Merchant ID

**Response**: `200 OK`

```json
{
  "merchant_id": "merchant_1234567890",
  "predictions": [
    {
      "scenario": "baseline",
      "predicted_risk_score": 0.15,
      "confidence": 0.90,
      "time_horizon": "30_days"
    },
    {
      "scenario": "optimistic",
      "predicted_risk_score": 0.10,
      "confidence": 0.85,
      "time_horizon": "30_days"
    },
    {
      "scenario": "pessimistic",
      "predicted_risk_score": 0.25,
      "confidence": 0.88,
      "time_horizon": "30_days"
    }
  ],
  "generated_at": "2025-01-27T12:00:00Z"
}
```

---

### Business Intelligence

#### POST /api/v1/bi/analyze

Analyze business intelligence data.

**Authentication**: Required

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com"
}
```

**Response**: `200 OK`

```json
{
  "business_name": "Acme Corporation",
  "analysis": {
    "industry": "Technology",
    "competitors": [],
    "market_share": 0.05,
    "growth_rate": 0.15
  },
  "generated_at": "2025-01-27T12:00:00Z"
}
```

---

### Authentication Endpoints

#### POST /api/v1/auth/register

Register a new user.

**Authentication**: Not required

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "user_metadata": {
    "name": "John Doe"
  }
}
```

**Response**: `201 Created`

```json
{
  "user": {
    "id": "user_1234567890",
    "email": "user@example.com"
  },
  "access_token": "jwt-token-here",
  "refresh_token": "refresh-token-here"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid email or password
- `409 Conflict`: Email already registered

---

## Rate Limiting

### Limits

- **Default**: 1000 requests per hour per IP
- **Burst**: 2000 requests
- **Window**: 3600 seconds (1 hour)

### Rate Limit Headers

When rate limit is approached or exceeded, the following headers are included:

- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets

### Rate Limit Exceeded Response

**Status**: `429 Too Many Requests`

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later."
  },
  "request_id": "req_1234567890",
  "timestamp": "2025-01-27T12:00:00Z",
  "path": "/api/v1/merchants",
  "method": "GET"
}
```

---

## CORS

### Configuration

- **Allowed Origins**: Configurable (default: `*` for development)
- **Allowed Methods**: `GET, POST, PUT, DELETE, OPTIONS`
- **Allowed Headers**: `Content-Type, Authorization`
- **Allow Credentials**: `true`

### Preflight Requests

All endpoints support OPTIONS preflight requests for CORS.

---

## Response Times

### Expected Response Times

- Health check: < 100ms
- Classification (first request): < 5 seconds
- Classification (cached): < 100ms
- Merchant list: < 2 seconds
- Risk assessment: < 10 seconds
- Business intelligence: < 5 seconds

---

## Best Practices

### Request Headers

Always include:
- `Content-Type: application/json` for POST/PUT requests
- `Authorization: Bearer <token>` for protected endpoints
- `X-Request-ID: <unique-id>` (optional, for request tracking)

### Error Handling

- Always check HTTP status codes
- Parse error response for detailed error information
- Use `request_id` for support requests
- Implement retry logic for 5xx errors with exponential backoff

### Pagination

- Use appropriate `page_size` (default: 20, max: 100)
- Check `has_next` and `has_previous` for navigation
- Use `total` and `total_pages` for UI display

### Caching

- Classification results are cached for 5 minutes
- Check `X-Cache` header to see if response was cached
- Don't cache sensitive merchant data

---

## Support

For API support:
- **Documentation**: This document
- **Issue Tracker**: [Link to issue tracker]
- **Email**: api-support@kyb-platform.com

---

**Last Updated**: 2025-01-27  
**API Version**: 1.0

