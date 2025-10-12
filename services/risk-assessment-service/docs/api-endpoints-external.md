# External API Integration Endpoints

## Overview

This document describes the API endpoints for external data integration in the Risk Assessment Service. These endpoints provide access to premium external data sources including Thomson Reuters, OFAC, and World-Check.

## Base URL

```
https://api.kyb-platform.com/v3/external
```

## Authentication

All endpoints require authentication using API keys:

```bash
Authorization: Bearer <your-api-key>
```

## Endpoints

### 1. Get Comprehensive External Data

Retrieve comprehensive data from all enabled external APIs.

**Endpoint**: `POST /external/comprehensive`

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "country": "US",
  "entity_type": "corporation",
  "include_apis": ["thomson_reuters", "ofac", "worldcheck"]
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "thomson_reuters": {
      "company_profile": {
        "company_id": "TR_ACME_1234567890",
        "name": "Acme Corporation",
        "industry": "technology",
        "country": "US",
        "registration_number": "12345678",
        "founded_year": 2010,
        "employee_count": 500,
        "revenue": 50000000
      },
      "financial_data": {
        "revenue": 50000000,
        "net_income": 5000000,
        "total_assets": 100000000,
        "total_liabilities": 40000000,
        "cash_flow": 8000000
      },
      "risk_metrics": {
        "overall_risk_score": 0.3,
        "financial_risk": 0.2,
        "operational_risk": 0.4,
        "market_risk": 0.3
      },
      "esg_score": {
        "overall_esg_score": 85.0,
        "environmental_score": 80.0,
        "social_score": 90.0,
        "governance_score": 85.0
      }
    },
    "ofac": {
      "sanctions_search": {
        "entity_name": "Acme Corporation",
        "matches_found": 0,
        "search_date": "2025-01-12T14:37:19Z"
      },
      "compliance_status": {
        "status": "compliant",
        "sanctions_matches": 0,
        "last_checked": "2025-01-12T14:37:19Z"
      },
      "entity_verification": {
        "verified": true,
        "verification_date": "2025-01-12T14:37:19Z",
        "confidence_score": 0.95
      }
    },
    "worldcheck": {
      "profile": {
        "entity_name": "Acme Corporation",
        "entity_type": "corporation",
        "country": "US",
        "risk_level": "low"
      },
      "adverse_media": {
        "articles_found": 0,
        "risk_score": 0.1,
        "last_updated": "2025-01-12T14:37:19Z"
      },
      "pep_status": {
        "is_pep": false,
        "pep_type": "",
        "confidence": 0.9
      },
      "risk_assessment": {
        "overall_risk_score": 0.2,
        "risk_level": "low",
        "risk_factors": ["low_media_risk", "no_pep_status"]
      }
    }
  },
  "processing_time": "1.2s",
  "data_quality": {
    "thomson_reuters": 0.95,
    "ofac": 0.98,
    "worldcheck": 0.92
  }
}
```

### 2. Get Thomson Reuters Data

Retrieve financial and company data from Thomson Reuters.

**Endpoint**: `POST /external/thomson-reuters`

**Request Body**:
```json
{
  "business_name": "Acme Corporation",
  "country": "US",
  "data_types": ["profile", "financial", "ratios", "risk_metrics", "esg", "executives", "ownership"]
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "company_profile": {
      "company_id": "TR_ACME_1234567890",
      "name": "Acme Corporation",
      "industry": "technology",
      "country": "US",
      "registration_number": "12345678",
      "founded_year": 2010,
      "employee_count": 500,
      "revenue": 50000000,
      "website": "https://www.acme.com",
      "description": "Leading technology company"
    },
    "financial_data": {
      "revenue": 50000000,
      "net_income": 5000000,
      "total_assets": 100000000,
      "total_liabilities": 40000000,
      "cash_flow": 8000000,
      "revenue_growth": 0.15,
      "profit_margin": 0.10
    },
    "financial_ratios": {
      "current_ratio": 2.5,
      "debt_to_equity": 0.4,
      "return_on_equity": 0.12,
      "gross_margin": 0.35,
      "operating_margin": 0.20
    },
    "risk_metrics": {
      "overall_risk_score": 0.3,
      "financial_risk": 0.2,
      "operational_risk": 0.4,
      "market_risk": 0.3,
      "credit_risk": 0.25
    },
    "esg_score": {
      "overall_esg_score": 85.0,
      "environmental_score": 80.0,
      "social_score": 90.0,
      "governance_score": 85.0,
      "esg_risk_level": "low"
    },
    "executive_info": {
      "ceo": {
        "name": "John Smith",
        "title": "Chief Executive Officer",
        "tenure_years": 5,
        "compensation": 2000000
      },
      "cfo": {
        "name": "Jane Doe",
        "title": "Chief Financial Officer",
        "tenure_years": 3,
        "compensation": 1500000
      }
    },
    "ownership_structure": {
      "company_id": "TR_ACME_1234567890",
      "ownership_data": [
        {
          "owner_name": "Institutional Investor A",
          "owner_type": "institution",
          "ownership_percentage": 15.5,
          "shares": 1500000
        }
      ],
      "last_updated": "2025-01-12T14:37:19Z"
    }
  },
  "processing_time": "0.8s",
  "data_quality": 0.95
}
```

### 3. Get OFAC Data

Retrieve sanctions and compliance data from OFAC.

**Endpoint**: `POST /external/ofac`

**Request Body**:
```json
{
  "entity_name": "Acme Corporation",
  "entity_type": "corporation"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "sanctions_search": {
      "entity_name": "Acme Corporation",
      "entity_type": "corporation",
      "matches_found": 0,
      "search_date": "2025-01-12T14:37:19Z",
      "lists_searched": ["SDN", "SSI", "FSE", "NS_MBS"]
    },
    "compliance_status": {
      "status": "compliant",
      "sanctions_matches": 0,
      "last_checked": "2025-01-12T14:37:19Z",
      "compliance_score": 0.98
    },
    "entity_verification": {
      "verified": true,
      "verification_date": "2025-01-12T14:37:19Z",
      "confidence_score": 0.95,
      "verification_method": "automated"
    }
  },
  "processing_time": "0.3s",
  "data_quality": 0.98
}
```

### 4. Get World-Check Data

Retrieve due diligence and adverse media data from World-Check.

**Endpoint**: `POST /external/worldcheck`

**Request Body**:
```json
{
  "entity_name": "Acme Corporation"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "profile": {
      "entity_name": "Acme Corporation",
      "entity_type": "corporation",
      "country": "US",
      "risk_level": "low",
      "profile_id": "WC_ACME_1234567890"
    },
    "adverse_media": {
      "articles_found": 0,
      "risk_score": 0.1,
      "last_updated": "2025-01-12T14:37:19Z",
      "media_sources": ["news", "regulatory", "legal"]
    },
    "pep_status": {
      "is_pep": false,
      "pep_type": "",
      "confidence": 0.9,
      "last_checked": "2025-01-12T14:37:19Z"
    },
    "sanctions_info": {
      "sanctions_matches": 0,
      "risk_level": "low",
      "last_updated": "2025-01-12T14:37:19Z"
    },
    "risk_assessment": {
      "overall_risk_score": 0.2,
      "risk_level": "low",
      "risk_factors": ["low_media_risk", "no_pep_status", "no_sanctions"],
      "assessment_date": "2025-01-12T14:37:19Z"
    }
  },
  "processing_time": "0.5s",
  "data_quality": 0.92
}
```

### 5. Get API Status

Check the status of all external APIs.

**Endpoint**: `GET /external/status`

**Response**:
```json
{
  "success": true,
  "data": {
    "thomson_reuters": {
      "status": "healthy",
      "response_time": "0.8s",
      "last_checked": "2025-01-12T14:37:19Z",
      "rate_limit": {
        "current": 45,
        "limit": 100,
        "reset_time": "2025-01-12T15:00:00Z"
      }
    },
    "ofac": {
      "status": "healthy",
      "response_time": "0.3s",
      "last_checked": "2025-01-12T14:37:19Z",
      "rate_limit": {
        "current": 20,
        "limit": 50,
        "reset_time": "2025-01-12T15:00:00Z"
      }
    },
    "worldcheck": {
      "status": "healthy",
      "response_time": "0.5s",
      "last_checked": "2025-01-12T14:37:19Z",
      "rate_limit": {
        "current": 30,
        "limit": 75,
        "reset_time": "2025-01-12T15:00:00Z"
      }
    }
  }
}
```

### 6. Get Supported APIs

Get list of supported external APIs and their capabilities.

**Endpoint**: `GET /external/supported`

**Response**:
```json
{
  "success": true,
  "data": {
    "apis": [
      {
        "name": "thomson_reuters",
        "display_name": "Thomson Reuters",
        "enabled": true,
        "capabilities": [
          "company_profiles",
          "financial_data",
          "risk_metrics",
          "esg_scoring",
          "executive_info",
          "ownership_structure"
        ],
        "rate_limit": 100,
        "timeout": "30s"
      },
      {
        "name": "ofac",
        "display_name": "OFAC",
        "enabled": true,
        "capabilities": [
          "sanctions_screening",
          "compliance_verification",
          "entity_verification"
        ],
        "rate_limit": 50,
        "timeout": "30s"
      },
      {
        "name": "worldcheck",
        "display_name": "World-Check",
        "enabled": true,
        "capabilities": [
          "entity_profiling",
          "adverse_media",
          "pep_screening",
          "risk_assessment"
        ],
        "rate_limit": 75,
        "timeout": "30s"
      }
    ]
  }
}
```

### 7. Health Check

Perform a comprehensive health check of all external APIs.

**Endpoint**: `GET /external/health`

**Response**:
```json
{
  "success": true,
  "data": {
    "overall_status": "healthy",
    "timestamp": "2025-01-12T14:37:19Z",
    "apis": [
      {
        "name": "thomson_reuters",
        "status": "healthy",
        "response_time": "0.8s",
        "error_rate": 0.0,
        "last_error": null
      },
      {
        "name": "ofac",
        "status": "healthy",
        "response_time": "0.3s",
        "error_rate": 0.0,
        "last_error": null
      },
      {
        "name": "worldcheck",
        "status": "healthy",
        "response_time": "0.5s",
        "error_rate": 0.0,
        "last_error": null
      }
    ],
    "summary": {
      "total_apis": 3,
      "healthy_apis": 3,
      "unhealthy_apis": 0,
      "average_response_time": "0.53s"
    }
  }
}
```

## Error Responses

### Standard Error Format

```json
{
  "success": false,
  "error": {
    "code": "API_ERROR",
    "message": "External API request failed",
    "details": "Rate limit exceeded for Thomson Reuters API",
    "timestamp": "2025-01-12T14:37:19Z"
  }
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_REQUEST` | Invalid request parameters | 400 |
| `AUTHENTICATION_FAILED` | Invalid or missing API key | 401 |
| `RATE_LIMIT_EXCEEDED` | API rate limit exceeded | 429 |
| `API_UNAVAILABLE` | External API is unavailable | 503 |
| `TIMEOUT` | Request timeout | 504 |
| `INTERNAL_ERROR` | Internal server error | 500 |

### Rate Limit Exceeded

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded for Thomson Reuters API",
    "details": "Current usage: 100/100 requests per minute",
    "retry_after": 60,
    "timestamp": "2025-01-12T14:37:19Z"
  }
}
```

### API Unavailable

```json
{
  "success": false,
  "error": {
    "code": "API_UNAVAILABLE",
    "message": "Thomson Reuters API is currently unavailable",
    "details": "Service is experiencing high load",
    "timestamp": "2025-01-12T14:37:19Z"
  }
}
```

## Rate Limiting

### Limits by API

| API | Rate Limit | Time Window |
|-----|------------|-------------|
| Thomson Reuters | 100 requests | per minute |
| OFAC | 50 requests | per minute |
| World-Check | 75 requests | per minute |

### Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1642008000
```

## Request/Response Examples

### cURL Examples

#### Get Comprehensive Data
```bash
curl -X POST "https://api.kyb-platform.com/v3/external/comprehensive" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "country": "US",
    "entity_type": "corporation",
    "include_apis": ["thomson_reuters", "ofac", "worldcheck"]
  }'
```

#### Get Thomson Reuters Data
```bash
curl -X POST "https://api.kyb-platform.com/v3/external/thomson-reuters" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "country": "US",
    "data_types": ["profile", "financial", "risk_metrics"]
  }'
```

#### Check API Status
```bash
curl -X GET "https://api.kyb-platform.com/v3/external/status" \
  -H "Authorization: Bearer your-api-key"
```

### Python Examples

```python
import requests

# Get comprehensive data
response = requests.post(
    "https://api.kyb-platform.com/v3/external/comprehensive",
    headers={"Authorization": "Bearer your-api-key"},
    json={
        "business_name": "Acme Corporation",
        "country": "US",
        "entity_type": "corporation",
        "include_apis": ["thomson_reuters", "ofac", "worldcheck"]
    }
)

data = response.json()
print(f"Thomson Reuters risk score: {data['data']['thomson_reuters']['risk_metrics']['overall_risk_score']}")
```

### JavaScript Examples

```javascript
// Get comprehensive data
const response = await fetch('https://api.kyb-platform.com/v3/external/comprehensive', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer your-api-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    business_name: 'Acme Corporation',
    country: 'US',
    entity_type: 'corporation',
    include_apis: ['thomson_reuters', 'ofac', 'worldcheck']
  })
});

const data = await response.json();
console.log('OFAC compliance status:', data.data.ofac.compliance_status.status);
```

## Best Practices

### 1. Request Optimization
- Use specific data types to reduce response size
- Cache responses when possible
- Batch requests when appropriate

### 2. Error Handling
- Implement exponential backoff for retries
- Handle rate limit errors gracefully
- Monitor API health status

### 3. Security
- Store API keys securely
- Use HTTPS for all requests
- Implement request logging and monitoring

### 4. Performance
- Use concurrent requests when possible
- Implement connection pooling
- Monitor response times and error rates

## Support

For technical support or questions about the external API integration:

- **Documentation**: [External APIs Documentation](./external-apis.md)
- **API Status**: [API Status Page](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Developer Portal**: [Developer Portal](https://developers.kyb-platform.com)
