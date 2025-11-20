# API Documentation

**Date:** 2025-01-27  
**Version:** 1.0.0  
**Purpose:** Comprehensive documentation of all API endpoints, including portfolio-level endpoints, comparison endpoints, request/response schemas, error responses, route mappings, and path transformations.

---

## Table of Contents

1. [Overview](#overview)
2. [Base URLs](#base-urls)
3. [Authentication](#authentication)
4. [Portfolio-Level Endpoints](#portfolio-level-endpoints)
5. [Comparison Endpoints](#comparison-endpoints)
6. [Merchant Endpoints](#merchant-endpoints)
7. [Risk Assessment Endpoints](#risk-assessment-endpoints)
8. [Analytics Endpoints](#analytics-endpoints)
9. [Error Responses](#error-responses)
10. [Route Mappings](#route-mappings)
11. [Path Transformations](#path-transformations)

---

## Overview

This API provides comprehensive business verification, risk assessment, and portfolio analytics capabilities. The API is organized into several categories:

- **Portfolio-Level Endpoints:** Aggregate data across all merchants
- **Comparison Endpoints:** Compare merchant data against portfolio or industry benchmarks
- **Merchant Endpoints:** Individual merchant operations
- **Risk Assessment Endpoints:** Risk scoring and assessment
- **Analytics Endpoints:** Portfolio-level analytics and insights

---

## Base URLs

### Production
```
https://api-gateway-production.up.railway.app
```

### Development
```
http://localhost:8080
```

### API Versions
- **v1:** Current stable version (recommended)
- **v3:** Enhanced endpoints (dashboard metrics)

---

## Authentication

### API Key Authentication

All endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

### Public Endpoints

The following endpoints do not require authentication:
- `/health`
- `/api/v1/classify`
- `/api/v1/classification/health`
- `/api/v1/merchant/health`
- `/api/v1/risk/health`

---

## Portfolio-Level Endpoints

### Get Portfolio Analytics

**Endpoint:** `GET /api/v1/merchants/analytics`

**Description:** Returns portfolio-wide analytics aggregated across all merchants.

**Query Parameters:** None

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/analytics" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface PortfolioAnalytics {
  totalMerchants: number;
  averageClassificationConfidence: number; // 0-1
  averageSecurityTrustScore: number; // 0-1
  averageDataQuality: number; // 0-1
  industryDistribution: Record<string, number>; // Industry name -> count
  countryDistribution: Record<string, number>; // Country code -> count
  timestamp: string; // ISO 8601
}
```

**Response Example:**
```json
{
  "totalMerchants": 1250,
  "averageClassificationConfidence": 0.92,
  "averageSecurityTrustScore": 0.85,
  "averageDataQuality": 0.88,
  "industryDistribution": {
    "Technology": 450,
    "Retail": 320,
    "Finance": 280,
    "Healthcare": 200
  },
  "countryDistribution": {
    "US": 800,
    "GB": 250,
    "CA": 200
  },
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Get Portfolio Statistics

**Endpoint:** `GET /api/v1/merchants/statistics`

**Description:** Returns portfolio-wide statistics including risk distribution and industry breakdown.

**Query Parameters:** None

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/statistics" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface PortfolioStatistics {
  totalMerchants: number;
  averageRiskScore: number; // 0-1
  riskDistribution: {
    low: number;
    medium: number;
    high: number;
    critical: number;
  };
  industryBreakdown: Array<{
    industry: string;
    count: number;
    averageRiskScore: number;
  }>;
  countryBreakdown: Array<{
    country: string;
    count: number;
    averageRiskScore: number;
  }>;
  timestamp: string; // ISO 8601
}
```

**Response Example:**
```json
{
  "totalMerchants": 1250,
  "averageRiskScore": 0.45,
  "riskDistribution": {
    "low": 500,
    "medium": 400,
    "high": 300,
    "critical": 50
  },
  "industryBreakdown": [
    {
      "industry": "Technology",
      "count": 450,
      "averageRiskScore": 0.35
    },
    {
      "industry": "Retail",
      "count": 320,
      "averageRiskScore": 0.55
    }
  ],
  "countryBreakdown": [
    {
      "country": "US",
      "count": 800,
      "averageRiskScore": 0.42
    }
  ],
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

## Comparison Endpoints

### Get Merchant Risk Score

**Endpoint:** `GET /api/v1/merchants/{id}/risk-score`

**Description:** Returns the risk score for a specific merchant, including confidence and assessment date.

**Path Parameters:**
- `id` (string, required) - Merchant ID

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/merchant-123/risk-score" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface MerchantRiskScore {
  merchantId: string;
  score: number; // 0-1
  level: 'low' | 'medium' | 'high' | 'critical';
  confidence: number; // 0-1
  assessmentDate: string; // ISO 8601
  factors: Array<{
    name: string;
    score: number;
    weight: number;
  }>;
}
```

**Response Example:**
```json
{
  "merchantId": "merchant-123",
  "score": 0.65,
  "level": "medium",
  "confidence": 0.85,
  "assessmentDate": "2025-01-27T10:00:00Z",
  "factors": [
    {
      "name": "transaction_volume",
      "score": 0.8,
      "weight": 0.3
    },
    {
      "name": "business_age",
      "score": 0.5,
      "weight": 0.2
    }
  ]
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Merchant not found
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Get Risk Benchmarks

**Endpoint:** `GET /api/v1/risk/benchmarks`

**Description:** Returns industry risk benchmarks for comparison against merchant risk scores.

**Query Parameters:**
- `mcc` (string, optional) - Merchant Category Code
- `naics` (string, optional) - NAICS code
- `sic` (string, optional) - SIC code

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/risk/benchmarks?mcc=5734" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskBenchmarks {
  industry: string;
  average_risk_score: number; // 0-1
  percentile_25: number; // 0-1
  percentile_50: number; // 0-1 (median)
  percentile_75: number; // 0-1
  sample_size: number;
  timestamp: string; // ISO 8601
}
```

**Response Example:**
```json
{
  "industry": "Technology",
  "average_risk_score": 0.45,
  "percentile_25": 0.30,
  "percentile_50": 0.45,
  "percentile_75": 0.60,
  "sample_size": 450,
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid query parameters
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

## Analytics Endpoints

### Get Risk Trends

**Endpoint:** `GET /api/v1/analytics/trends`

**Description:** Returns portfolio risk trends over time with predictions and confidence bands.

**Query Parameters:**
- `timeframe` (string, optional) - Time range: `7d`, `30d`, `90d`, `6m`, `1y` (default: `30d`)
- `industry` (string, optional) - Filter by industry
- `country` (string, optional) - Filter by country code
- `limit` (number, optional) - Maximum number of data points (default: 100)

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/analytics/trends?timeframe=6m&industry=Technology" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskTrends {
  trends: Array<{
    date: string; // ISO 8601
    industry?: string;
    country?: string;
    average_risk_score: number; // 0-1
    trend_direction: 'increasing' | 'decreasing' | 'stable';
    change_percentage: number; // Percentage change
  }>;
  summary: {
    average_risk_score: number; // 0-1
    overall_trend: 'increasing' | 'decreasing' | 'stable';
    total_change_percentage: number;
  };
  timestamp: string; // ISO 8601
}
```

**Response Example:**
```json
{
  "trends": [
    {
      "date": "2025-01-01T00:00:00Z",
      "industry": "Technology",
      "average_risk_score": 0.40,
      "trend_direction": "decreasing",
      "change_percentage": -5.0
    },
    {
      "date": "2025-01-15T00:00:00Z",
      "industry": "Technology",
      "average_risk_score": 0.38,
      "trend_direction": "decreasing",
      "change_percentage": -2.0
    }
  ],
  "summary": {
    "average_risk_score": 0.39,
    "overall_trend": "decreasing",
    "total_change_percentage": -7.0
  },
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid query parameters
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Get Risk Insights

**Endpoint:** `GET /api/v1/analytics/insights`

**Description:** Returns portfolio risk insights with key findings and recommendations.

**Query Parameters:**
- `industry` (string, optional) - Filter by industry
- `country` (string, optional) - Filter by country code
- `risk_level` (string, optional) - Filter by risk level: `low`, `medium`, `high`, `critical`

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/analytics/insights?risk_level=high" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskInsights {
  insights: Array<{
    title: string;
    description: string;
    impact: 'low' | 'medium' | 'high';
    category: string;
  }>;
  recommendations: Array<{
    title: string;
    description: string;
    priority: 'low' | 'medium' | 'high';
    actions: string[];
  }>;
  timestamp: string; // ISO 8601
}
```

**Response Example:**
```json
{
  "insights": [
    {
      "title": "High Risk Merchant Concentration",
      "description": "15% of merchants are classified as high risk, concentrated in Retail industry",
      "impact": "high",
      "category": "risk_distribution"
    }
  ],
  "recommendations": [
    {
      "title": "Increase Monitoring Frequency",
      "description": "Consider increasing monitoring frequency for high-risk merchants",
      "priority": "high",
      "actions": [
        "Configure daily risk assessments",
        "Set up automated alerts"
      ]
    }
  ],
  "timestamp": "2025-01-27T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid query parameters
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

## Merchant Endpoints

### Get Merchant by ID

**Endpoint:** `GET /api/v1/merchants/{id}`

**Description:** Returns detailed information for a specific merchant.

**Path Parameters:**
- `id` (string, required) - Merchant ID

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/merchant-123" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface Merchant {
  id: string;
  businessName: string;
  industry: string;
  status: 'active' | 'inactive' | 'pending';
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  // ... other merchant fields
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Merchant not found
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Get Merchant Analytics

**Endpoint:** `GET /api/v1/merchants/{id}/analytics`

**Description:** Returns analytics data for a specific merchant.

**Path Parameters:**
- `id` (string, required) - Merchant ID

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/merchant-123/analytics" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface AnalyticsData {
  merchantId: string;
  classificationConfidence: number; // 0-1
  securityTrustScore: number; // 0-1
  dataQualityScore: number; // 0-1
  // ... other analytics fields
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Merchant not found
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

## Risk Assessment Endpoints

### Get Risk Indicators

**Endpoint:** `GET /api/v1/risk/indicators/{id}`

**Description:** Returns active risk indicators (alerts) for a merchant.

**Path Parameters:**
- `id` (string, required) - Merchant ID

**Query Parameters:**
- `status` (string, optional) - Filter by status: `active`, `resolved` (default: `active`)
- `severity` (string, optional) - Filter by severity: `low`, `medium`, `high`, `critical`

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/risk/indicators/merchant-123?status=active&severity=high" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskIndicatorsData {
  alerts: Array<{
    id: string;
    type: string;
    severity: 'low' | 'medium' | 'high' | 'critical';
    description: string;
    createdAt: string; // ISO 8601
    status: 'active' | 'resolved';
  }>;
  timestamp: string; // ISO 8601
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Merchant not found
- `400 Bad Request` - Invalid query parameters
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Explain Risk Assessment

**Endpoint:** `GET /api/v1/risk/explain/{assessmentId}`

**Description:** Returns SHAP values and feature importance for a risk assessment.

**Path Parameters:**
- `assessmentId` (string, required) - Risk assessment ID

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/risk/explain/assessment-123" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskExplanationResponse {
  assessmentId: string;
  shapValues: Array<{
    feature: string;
    value: number; // SHAP value
  }>;
  featureImportance: Array<{
    feature: string;
    importance: number; // 0-1
  }>;
  riskFactors: Array<{
    name: string;
    score: number; // 0-1
    weight: number; // 0-1
    impact: number; // Calculated: score * weight
  }>;
  timestamp: string; // ISO 8601
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Assessment not found
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

### Get Risk Recommendations

**Endpoint:** `GET /api/v1/merchants/{id}/risk-recommendations`

**Description:** Returns actionable risk recommendations for a merchant.

**Path Parameters:**
- `id` (string, required) - Merchant ID

**Request Example:**
```bash
curl -X GET "https://api-gateway-production.up.railway.app/api/v1/merchants/merchant-123/risk-recommendations" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response Schema:**
```typescript
interface RiskRecommendationsResponse {
  recommendations: Array<{
    id: string;
    title: string;
    description: string;
    priority: 'low' | 'medium' | 'high';
    actions: string[];
  }>;
  timestamp: string; // ISO 8601
}
```

**Status Codes:**
- `200 OK` - Success
- `404 Not Found` - Merchant not found
- `401 Unauthorized` - Missing or invalid authentication
- `500 Internal Server Error` - Server error

---

## Error Responses

### Standard Error Format

All error responses follow this format:

```typescript
interface ErrorResponse {
  error: string;
  message: string;
  statusCode: number;
  timestamp: string; // ISO 8601
  path?: string; // Request path
}
```

### Error Response Example

```json
{
  "error": "Not Found",
  "message": "Merchant not found: merchant-123",
  "statusCode": 404,
  "timestamp": "2025-01-27T10:30:00Z",
  "path": "/api/v1/merchants/merchant-123"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Backend service unavailable
- `503 Service Unavailable` - Service temporarily unavailable

---

## Route Mappings

### API Gateway to Backend Service Mappings

| API Gateway Route | Backend Service | Backend Route |
|-------------------|----------------|---------------|
| `/api/v1/merchants/*` | Merchant Service | `/api/v1/merchants/*` |
| `/api/v1/merchants/analytics` | Merchant Service | `/api/v1/merchants/analytics` |
| `/api/v1/merchants/statistics` | Merchant Service | `/api/v1/merchants/statistics` |
| `/api/v1/analytics/trends` | Risk Assessment Service | `/api/v1/analytics/trends` |
| `/api/v1/analytics/insights` | Risk Assessment Service | `/api/v1/analytics/insights` |
| `/api/v1/risk/assess` | Risk Assessment Service | `/api/v1/assessments` |
| `/api/v1/risk/benchmarks` | Risk Assessment Service | `/api/v1/benchmarks` |
| `/api/v1/risk/indicators/{id}` | Risk Assessment Service | `/api/v1/indicators/{id}` |
| `/api/v1/risk/explain/{id}` | Risk Assessment Service | `/api/v1/explain/{id}` |
| `/api/v3/dashboard/metrics` | BI Service | `/dashboard/metrics` |

---

## Path Transformations

### Risk Assessment Service Transformations

The following routes require path transformation when proxying to the Risk Assessment Service:

| API Gateway Route | Transformed Backend Route |
|-------------------|--------------------------|
| `/api/v1/risk/assess` | `/api/v1/assessments` |
| `/api/v1/risk/benchmarks` | `/api/v1/benchmarks` |
| `/api/v1/risk/predictions/{merchant_id}` | `/api/v1/predictions/{merchant_id}` |
| `/api/v1/risk/indicators/{id}` | `/api/v1/indicators/{id}` |

### No Transformation Routes

The following routes are passed as-is to backend services:

- `/api/v1/merchants/*` → Merchant Service (no transformation)
- `/api/v1/analytics/*` → Risk Assessment Service (no transformation)
- `/api/v3/dashboard/metrics` → BI Service (no transformation)

---

## Rate Limiting

### Default Limits

- **Rate Limit:** 100 requests per 60 seconds per IP
- **Burst Size:** 200 requests
- **Window Size:** 60 seconds

### Rate Limit Headers

When rate limit is exceeded, the following headers are included:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1643299200
```

### Rate Limit Response

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests",
  "statusCode": 429,
  "timestamp": "2025-01-27T10:30:00Z"
}
```

---

## Caching

### Cache Headers

Responses may include cache headers:

```
Cache-Control: public, max-age=300
ETag: "abc123"
Last-Modified: Mon, 27 Jan 2025 10:30:00 GMT
```

### Cache TTL

- **Portfolio Analytics:** 7 minutes (420 seconds)
- **Portfolio Statistics:** 5 minutes (300 seconds)
- **Risk Trends:** 5 minutes (300 seconds)
- **Risk Insights:** 5 minutes (300 seconds)
- **Risk Benchmarks:** 10 minutes (600 seconds)
- **Merchant Risk Score:** 2 minutes (120 seconds)

---

## Pagination

### Paginated Endpoints

Some endpoints support pagination:

**Query Parameters:**
- `limit` (number, optional) - Number of results per page (default: 20, max: 100)
- `offset` (number, optional) - Number of results to skip (default: 0)

**Response Headers:**
```
X-Total-Count: 1250
X-Page-Size: 20
X-Page-Number: 1
```

---

## Versioning

### API Versioning Strategy

- **v1:** Current stable version (recommended)
- **v3:** Enhanced endpoints (dashboard metrics)

### Version in URL

All endpoints include version in the URL path:
- `/api/v1/merchants/*`
- `/api/v3/dashboard/metrics`

### Deprecation Policy

- Deprecated endpoints will be marked with `X-API-Deprecated` header
- Minimum 6 months notice before removal
- Migration guides provided for deprecated endpoints

---

## Conclusion

This documentation covers all portfolio-level endpoints, comparison endpoints, request/response schemas, error responses, route mappings, and path transformations.

**Last Updated:** 2025-01-27  
**Version:** 1.0.0  
**Status:** ✅ Complete
