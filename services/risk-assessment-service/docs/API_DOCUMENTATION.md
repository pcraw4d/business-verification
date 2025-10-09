# Risk Assessment Service API Documentation

## Overview

The Risk Assessment Service provides comprehensive business risk assessment capabilities using advanced machine learning models and real-time data analysis. This API enables developers to assess business risks, predict future risk trends, and monitor compliance requirements.

## Base URL

```
https://api.kyb-platform.com/v1
```

## Authentication

All API requests require authentication using an API key. Include your API key in the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

## Rate Limiting

- **Rate Limit**: 100 requests per minute per API key
- **Headers**: Rate limit information is included in response headers:
  - `X-RateLimit-Limit`: Maximum requests per minute
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Time when the rate limit resets

## Error Handling

The API uses standard HTTP status codes and returns detailed error information in JSON format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": "business_name is required",
    "field": "business_name",
    "validation": [
      {
        "field": "business_name",
        "message": "business_name is required",
        "code": "INVALID_BUSINESS_NAME"
      }
    ]
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z",
  "path": "/api/v1/assess",
  "method": "POST"
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_ERROR` | Invalid or missing API key |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource conflict |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable |
| `REQUEST_TIMEOUT` | Request timeout |
| `INTERNAL_ERROR` | Internal server error |

## Endpoints

### 1. Risk Assessment

#### POST /api/v1/assess

Performs a comprehensive risk assessment for a business.

**Request Body:**
```json
{
  "business_name": "Acme Corporation",
  "business_address": "123 Main St, Anytown, ST 12345",
  "industry": "Technology",
  "country": "US",
  "phone": "+1-555-123-4567",
  "email": "contact@acme.com",
  "website": "https://www.acme.com",
  "prediction_horizon": 3,
  "metadata": {
    "annual_revenue": 1000000,
    "employee_count": 50,
    "founded_year": 2020
  }
}
```

**Response:**
```json
{
  "id": "risk_1234567890",
  "business_id": "biz_1234567890",
  "risk_score": 0.75,
  "risk_level": "medium",
  "risk_factors": [
    {
      "category": "financial",
      "name": "Credit Score",
      "score": 0.8,
      "weight": 0.3,
      "description": "Business credit score analysis",
      "source": "internal",
      "confidence": 0.9
    }
  ],
  "prediction_horizon": 3,
  "confidence_score": 0.85,
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "metadata": {
    "annual_revenue": 1000000,
    "employee_count": 50,
    "founded_year": 2020
  }
}
```

**Validation Rules:**
- `business_name`: Required, 1-255 characters
- `business_address`: Required, 10-500 characters
- `industry`: Required, 1-100 characters
- `country`: Required, 2-letter ISO code
- `phone`: Optional, E.164 format
- `email`: Optional, valid email format
- `website`: Optional, valid URL
- `prediction_horizon`: 0-12 months
- `metadata`: Optional, max 50 fields

#### GET /api/v1/assess/{id}

Retrieves a risk assessment by ID.

**Response:**
```json
{
  "id": "risk_1234567890",
  "business_id": "biz_1234567890",
  "risk_score": 0.75,
  "risk_level": "medium",
  "risk_factors": [...],
  "prediction_horizon": 3,
  "confidence_score": 0.85,
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "metadata": {...}
}
```

#### POST /api/v1/assess/{id}/predict

Performs future risk prediction for a business.

**Request Body:**
```json
{
  "horizon_months": 6,
  "scenarios": ["optimistic", "realistic", "pessimistic"]
}
```

**Response:**
```json
{
  "business_id": "biz_1234567890",
  "horizon_months": 6,
  "predicted_score": 0.72,
  "predicted_level": "medium",
  "scenarios": [
    {
      "name": "optimistic",
      "score": 0.65,
      "level": "low",
      "confidence": 0.8
    },
    {
      "name": "realistic",
      "score": 0.72,
      "level": "medium",
      "confidence": 0.85
    },
    {
      "name": "pessimistic",
      "score": 0.85,
      "level": "high",
      "confidence": 0.75
    }
  ],
  "trend_analysis": {
    "direction": "improving",
    "magnitude": 0.05,
    "confidence": 0.8
  },
  "key_factors": [
    {
      "factor": "Market Conditions",
      "impact": 0.3,
      "trend": "improving"
    }
  ],
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### GET /api/v1/assess/{id}/history

Retrieves risk assessment history for a business.

**Response:**
```json
{
  "business_id": "biz_1234567890",
  "assessments": [
    {
      "id": "risk_1234567890",
      "risk_score": 0.75,
      "risk_level": "medium",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "trends": {
    "score_trend": "improving",
    "level_trend": "stable",
    "confidence_trend": "improving"
  }
}
```

### 2. Compliance

#### POST /api/v1/compliance/check

Performs compliance checks for a business.

**Request Body:**
```json
{
  "business_name": "Acme Corporation",
  "business_address": "123 Main St, Anytown, ST 12345",
  "industry": "Technology",
  "country": "US",
  "compliance_types": ["kyc", "aml", "sanctions"]
}
```

**Response:**
```json
{
  "business_id": "biz_1234567890",
  "compliance_status": "compliant",
  "checks": [
    {
      "type": "kyc",
      "status": "passed",
      "score": 0.95,
      "details": "All KYC requirements met"
    }
  ],
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### POST /api/v1/sanctions/screen

Performs sanctions screening for a business.

**Request Body:**
```json
{
  "business_name": "Acme Corporation",
  "business_address": "123 Main St, Anytown, ST 12345",
  "country": "US"
}
```

**Response:**
```json
{
  "business_id": "biz_1234567890",
  "sanctions_status": "clear",
  "matches": [],
  "screening_date": "2024-01-15T10:30:00Z"
}
```

#### POST /api/v1/media/monitor

Sets up adverse media monitoring for a business.

**Request Body:**
```json
{
  "business_name": "Acme Corporation",
  "business_address": "123 Main St, Anytown, ST 12345",
  "monitoring_types": ["news", "social_media", "regulatory"]
}
```

**Response:**
```json
{
  "business_id": "biz_1234567890",
  "monitoring_id": "mon_1234567890",
  "status": "active",
  "alerts": [],
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 3. Analytics

#### GET /api/v1/analytics/trends

Retrieves risk trends and analytics.

**Query Parameters:**
- `industry`: Filter by industry
- `country`: Filter by country
- `timeframe`: Time period (7d, 30d, 90d, 1y)
- `limit`: Number of results (default: 100)

**Response:**
```json
{
  "trends": [
    {
      "industry": "Technology",
      "country": "US",
      "average_risk_score": 0.72,
      "trend_direction": "improving",
      "change_percentage": -5.2,
      "sample_size": 1500
    }
  ],
  "summary": {
    "total_assessments": 10000,
    "average_risk_score": 0.75,
    "high_risk_percentage": 15.2
  }
}
```

#### GET /api/v1/analytics/insights

Retrieves risk insights and recommendations.

**Query Parameters:**
- `industry`: Filter by industry
- `country`: Filter by country
- `risk_level`: Filter by risk level

**Response:**
```json
{
  "insights": [
    {
      "type": "risk_factor",
      "title": "High Credit Risk in Technology Sector",
      "description": "Technology companies show 20% higher credit risk",
      "impact": "high",
      "recommendation": "Increase due diligence for tech companies"
    }
  ],
  "recommendations": [
    {
      "category": "monitoring",
      "action": "Increase monitoring frequency",
      "priority": "medium"
    }
  ]
}
```

## Data Models

### Risk Assessment Request
```json
{
  "business_name": "string",
  "business_address": "string",
  "industry": "string",
  "country": "string",
  "phone": "string",
  "email": "string",
  "website": "string",
  "prediction_horizon": "integer",
  "metadata": "object"
}
```

### Risk Assessment Response
```json
{
  "id": "string",
  "business_id": "string",
  "risk_score": "number",
  "risk_level": "string",
  "risk_factors": "array",
  "prediction_horizon": "integer",
  "confidence_score": "number",
  "status": "string",
  "created_at": "string",
  "updated_at": "string",
  "metadata": "object"
}
```

### Risk Factor
```json
{
  "category": "string",
  "name": "string",
  "score": "number",
  "weight": "number",
  "description": "string",
  "source": "string",
  "confidence": "number"
}
```

### Risk Prediction
```json
{
  "business_id": "string",
  "horizon_months": "integer",
  "predicted_score": "number",
  "predicted_level": "string",
  "scenarios": "array",
  "trend_analysis": "object",
  "key_factors": "array",
  "created_at": "string"
}
```

## Status Codes

| Code | Description |
|------|-------------|
| `pending` | Assessment in progress |
| `completed` | Assessment completed |
| `failed` | Assessment failed |
| `error` | Assessment error |

## Risk Levels

| Level | Score Range | Description |
|-------|-------------|-------------|
| `low` | 0.0 - 0.3 | Low risk |
| `medium` | 0.3 - 0.7 | Medium risk |
| `high` | 0.7 - 1.0 | High risk |

## Risk Categories

| Category | Description |
|----------|-------------|
| `financial` | Financial risk factors |
| `operational` | Operational risk factors |
| `compliance` | Compliance risk factors |
| `reputational` | Reputational risk factors |
| `regulatory` | Regulatory risk factors |

## SDKs

### Go SDK
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client := kyb.NewClient("YOUR_API_KEY")
    
    req := &kyb.RiskAssessmentRequest{
        BusinessName:    "Acme Corporation",
        BusinessAddress: "123 Main St, Anytown, ST 12345",
        Industry:        "Technology",
        Country:         "US",
    }
    
    assessment, err := client.AssessRisk(req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Risk Score: %.2f\n", assessment.RiskScore)
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
}
```

### Python SDK
```python
from kyb_sdk import KYBClient

client = KYBClient(api_key="YOUR_API_KEY")

req = {
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "industry": "Technology",
    "country": "US"
}

assessment = client.assess_risk(req)
print(f"Risk Score: {assessment['risk_score']}")
print(f"Risk Level: {assessment['risk_level']}")
```

### Node.js SDK
```javascript
const KYBClient = require('kyb-sdk');

const client = new KYBClient('YOUR_API_KEY');

const req = {
    business_name: 'Acme Corporation',
    business_address: '123 Main St, Anytown, ST 12345',
    industry: 'Technology',
    country: 'US'
};

client.assessRisk(req)
    .then(assessment => {
        console.log(`Risk Score: ${assessment.risk_score}`);
        console.log(`Risk Level: ${assessment.risk_level}`);
    })
    .catch(error => {
        console.error('Error:', error);
    });
```

## Webhooks

The API supports webhooks for real-time notifications:

### Webhook Events
- `assessment.completed` - Risk assessment completed
- `assessment.failed` - Risk assessment failed
- `prediction.updated` - Risk prediction updated
- `compliance.alert` - Compliance alert triggered
- `sanctions.match` - Sanctions match found
- `media.alert` - Adverse media alert

### Webhook Payload
```json
{
  "event": "assessment.completed",
  "data": {
    "id": "risk_1234567890",
    "business_id": "biz_1234567890",
    "risk_score": 0.75,
    "risk_level": "medium"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Support

For API support and questions:
- **Email**: api-support@kyb-platform.com
- **Documentation**: https://docs.kyb-platform.com
- **Status Page**: https://status.kyb-platform.com

## Changelog

### v1.0.0 (2024-01-15)
- Initial release
- Risk assessment endpoints
- Compliance checking
- Analytics and insights
- Webhook support
