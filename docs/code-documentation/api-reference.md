# Enhanced Business Intelligence System - API Reference

## Overview

This document provides comprehensive API reference documentation for the Enhanced Business Intelligence System. It covers all endpoints, request/response formats, authentication, error handling, and usage examples.

## Table of Contents

1. [Authentication](#authentication)
2. [Base URL and Versioning](#base-url-and-versioning)
3. [Common Response Formats](#common-response-formats)
4. [Error Handling](#error-handling)
5. [Classification Endpoints](#classification-endpoints)
6. [Risk Assessment Endpoints](#risk-assessment-endpoints)
7. [Data Discovery Endpoints](#data-discovery-endpoints)
8. [Caching Endpoints](#caching-endpoints)
9. [Monitoring Endpoints](#monitoring-endpoints)
10. [Health and Status Endpoints](#health-and-status-endpoints)

## Authentication

### API Key Authentication

The API supports API key authentication for secure access control.

**Header**: `Authorization: Bearer YOUR_API_KEY`

**Example**:
```bash
curl -H "Authorization: Bearer sk_live_1234567890abcdef" \
     -H "Content-Type: application/json" \
     -X POST https://api.kyb-platform.com/v1/classify \
     -d '{"business_name": "Acme Corporation"}'
```

### JWT Token Authentication

For advanced use cases, JWT token authentication is supported.

**Header**: `Authorization: Bearer JWT_TOKEN`

**Token Format**:
```json
{
  "sub": "user_1234567890",
  "iss": "kyb-platform",
  "aud": "kyb-api",
  "iat": 1640995200,
  "exp": 1641081600,
  "permissions": ["classify", "risk_assess", "data_discover"]
}
```

## Base URL and Versioning

### Base URL
- **Production**: `https://api.kyb-platform.com`
- **Staging**: `https://staging-api.kyb-platform.com`
- **Development**: `http://localhost:8080`

### API Versioning
The API uses URL-based versioning:
- **Current Version**: `/v1`
- **Future Versions**: `/v2`, `/v3`, etc.

### Rate Limiting
- **Standard Plan**: 1,000 requests per hour
- **Professional Plan**: 10,000 requests per hour
- **Enterprise Plan**: 100,000 requests per hour

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Common Response Formats

### Success Response
All successful API responses follow this format:
```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "metadata": {
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "processing_time": "1.2s",
    "cache_hit": false
  }
}
```

### Error Response
All error responses follow this format:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid business name provided",
    "details": {
      "field": "business_name",
      "reason": "Business name cannot be empty"
    },
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

## Error Handling

### HTTP Status Codes
- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Error Codes
- `VALIDATION_ERROR`: Input validation failed
- `CLASSIFICATION_ERROR`: Classification processing failed
- `RISK_ASSESSMENT_ERROR`: Risk assessment processing failed
- `DATA_DISCOVERY_ERROR`: Data discovery processing failed
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded
- `AUTHENTICATION_ERROR`: Authentication failed
- `AUTHORIZATION_ERROR`: Authorization failed
- `INTERNAL_ERROR`: Internal server error
- `SERVICE_UNAVAILABLE`: Service temporarily unavailable

## Classification Endpoints

### POST /v1/classify

Classifies a business using industry codes (NAICS, SIC, MCC).

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "description": "Technology consulting services",
  "website": "https://acme.com",
  "industry": "Technology",
  "keywords": ["consulting", "technology", "services"],
  "options": {
    "include_alternatives": true,
    "max_results": 3,
    "confidence_threshold": 0.7,
    "strategies": ["keyword", "ml", "similarity"]
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95,
        "reasoning": "Strong keyword matches: 'consulting', 'technology', 'services'"
      },
      "alternatives": [
        {
          "type": "SIC",
          "code": "7371",
          "description": "Computer Programming Services",
          "confidence": 0.92,
          "reasoning": "High similarity to primary classification"
        },
        {
          "type": "MCC",
          "code": "7392",
          "description": "Management Consulting Services",
          "confidence": 0.88,
          "reasoning": "Keyword match: 'consulting'"
        }
      ]
    },
    "strategies_used": ["keyword", "ml", "similarity"],
    "processing_time": "1.2s"
  },
  "metadata": {
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z",
    "processing_time": "1.2s",
    "cache_hit": false
  }
}
```

**cURL Example**:
```bash
curl -X POST https://api.kyb-platform.com/v1/classify \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "description": "Technology consulting services",
    "website": "https://acme.com"
  }'
```

### GET /v1/classify/{classification_id}

Retrieves a specific classification result.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95
      },
      "alternatives": [...]
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/classify/history/{business_id}

Retrieves classification history for a business.

**Query Parameters**:
- `limit` (optional): Number of results to return (default: 10, max: 100)
- `offset` (optional): Number of results to skip (default: 0)
- `start_date` (optional): Start date for filtering (ISO 8601 format)
- `end_date` (optional): End date for filtering (ISO 8601 format)

**Response**:
```json
{
  "success": true,
  "data": {
    "business_id": "business_1234567890",
    "classifications": [
      {
        "id": "class_1234567890",
        "primary_code": {
          "type": "NAICS",
          "code": "541511",
          "description": "Custom Computer Programming Services",
          "confidence": 0.95
        },
        "created_at": "2024-12-19T10:30:00Z"
      }
    ],
    "total_count": 1,
    "has_more": false
  }
}
```

## Risk Assessment Endpoints

### POST /v1/risk/assess

Assesses business risk factors and provides comprehensive risk scoring.

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com",
  "industry": "Technology",
  "options": {
    "include_security_analysis": true,
    "include_financial_analysis": true,
    "include_compliance_analysis": true,
    "include_reputation_analysis": true
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "risk_1234567890",
    "business_name": "Acme Corporation",
    "overall_risk": "LOW",
    "risk_score": 0.15,
    "risk_factors": {
      "security_risk": {
        "level": "LOW",
        "score": 0.1,
        "factors": [
          {
            "factor": "SSL Certificate",
            "status": "VALID",
            "details": "Valid SSL certificate until 2025-12-19"
          },
          {
            "factor": "Security Headers",
            "status": "GOOD",
            "details": "HSTS, CSP, and other security headers present"
          }
        ]
      },
      "financial_risk": {
        "level": "MEDIUM",
        "score": 0.3,
        "factors": [
          {
            "factor": "Company Size",
            "status": "UNKNOWN",
            "details": "Company size information not available"
          }
        ]
      },
      "compliance_risk": {
        "level": "LOW",
        "score": 0.05,
        "factors": [
          {
            "factor": "Data Protection",
            "status": "COMPLIANT",
            "details": "Privacy policy and data protection measures in place"
          }
        ]
      },
      "reputation_risk": {
        "level": "LOW",
        "score": 0.1,
        "factors": [
          {
            "factor": "Online Presence",
            "status": "GOOD",
            "details": "Professional website and social media presence"
          }
        ]
      }
    },
    "recommendations": [
      "Consider obtaining company size information for better financial risk assessment",
      "Monitor security headers regularly for compliance"
    ],
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/risk/assess/{risk_assessment_id}

Retrieves a specific risk assessment result.

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "risk_1234567890",
    "business_name": "Acme Corporation",
    "overall_risk": "LOW",
    "risk_score": 0.15,
    "risk_factors": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/risk/history/{business_id}

Retrieves risk assessment history for a business.

**Query Parameters**:
- `limit` (optional): Number of results to return (default: 10, max: 100)
- `offset` (optional): Number of results to skip (default: 0)
- `start_date` (optional): Start date for filtering (ISO 8601 format)
- `end_date` (optional): End date for filtering (ISO 8601 format)

## Data Discovery Endpoints

### POST /v1/discover

Discovers and extracts comprehensive business information from multiple sources.

**Request**:
```json
{
  "business_name": "Acme Corporation",
  "website": "https://acme.com",
  "options": {
    "include_website_analysis": true,
    "include_web_search": true,
    "include_social_media": true,
    "max_results": 10
  }
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "discover_1234567890",
    "business_name": "Acme Corporation",
    "website": "https://acme.com",
    "discovered_data": {
      "company_info": {
        "name": "Acme Corporation",
        "description": "Leading technology consulting firm",
        "founded": "2010",
        "headquarters": "San Francisco, CA",
        "contact": {
          "email": "contact@acme.com",
          "phone": "+1-555-123-4567",
          "address": "123 Tech Street, San Francisco, CA 94105"
        }
      },
      "team_info": {
        "size": "50-100 employees",
        "leadership": [
          {
            "name": "John Doe",
            "title": "CEO",
            "linkedin": "https://linkedin.com/in/johndoe"
          }
        ]
      },
      "products_services": [
        "Technology Consulting",
        "Software Development",
        "Digital Transformation",
        "Cloud Solutions"
      ],
      "business_model": "B2B",
      "technology_stack": [
        "React",
        "Node.js",
        "AWS",
        "Docker"
      ],
      "market_presence": {
        "regions": ["North America", "Europe"],
        "industries": ["Technology", "Finance", "Healthcare"],
        "competitors": ["TechCorp", "InnovateTech"]
      }
    },
    "data_quality": {
      "completeness": 0.85,
      "accuracy": 0.92,
      "consistency": 0.88,
      "freshness": 0.95,
      "overall": 0.90
    },
    "sources": [
      {
        "type": "website",
        "url": "https://acme.com",
        "confidence": 0.95
      },
      {
        "type": "web_search",
        "query": "Acme Corporation technology consulting",
        "confidence": 0.88
      }
    ],
    "created_at": "2024-12-19T10:30:00Z"
  }
}
```

### GET /v1/discover/{discovery_id}

Retrieves a specific data discovery result.

### GET /v1/discover/history/{business_id}

Retrieves data discovery history for a business.

## Caching Endpoints

### GET /v1/cache/{key}

Retrieves a cached value.

**Response**:
```json
{
  "success": true,
  "data": {
    "key": "classification_acme_corp",
    "value": {
      "classification": {...},
      "risk_assessment": {...}
    },
    "metadata": {
      "ttl": 3600,
      "created_at": "2024-12-19T10:30:00Z",
      "expires_at": "2024-12-19T11:30:00Z",
      "access_count": 5,
      "last_accessed": "2024-12-19T10:25:00Z"
    }
  }
}
```

### PUT /v1/cache/{key}

Stores a value in the cache.

**Request**:
```json
{
  "value": {
    "classification": {...},
    "risk_assessment": {...}
  },
  "ttl": 3600,
  "tags": ["classification", "acme_corp"],
  "priority": "high"
}
```

### DELETE /v1/cache/{key}

Removes a cached value.

### GET /v1/cache/stats

Retrieves cache statistics.

**Response**:
```json
{
  "success": true,
  "data": {
    "total_entries": 1000,
    "total_size": "50MB",
    "hit_rate": 0.85,
    "miss_rate": 0.15,
    "eviction_rate": 0.05,
    "average_ttl": 1800,
    "memory_usage": "25MB",
    "disk_usage": "25MB"
  }
}
```

### POST /v1/cache/optimize

Triggers cache optimization.

**Request**:
```json
{
  "strategy": "size_adjustment",
  "parameters": {
    "target_size": "100MB",
    "eviction_policy": "lru"
  }
}
```

## Monitoring Endpoints

### GET /v1/monitoring/metrics

Retrieves system metrics.

**Query Parameters**:
- `type` (optional): Metric type (performance, quality, errors)
- `timeframe` (optional): Timeframe for metrics (1h, 24h, 7d, 30d)
- `granularity` (optional): Metric granularity (1m, 5m, 1h, 1d)

**Response**:
```json
{
  "success": true,
  "data": {
    "performance": {
      "response_time": {
        "p50": 150,
        "p95": 300,
        "p99": 500,
        "average": 180
      },
      "throughput": {
        "requests_per_second": 100,
        "concurrent_users": 50
      }
    },
    "quality": {
      "accuracy_rate": 0.95,
      "confidence_average": 0.88,
      "misclassification_rate": 0.05
    },
    "errors": {
      "error_rate": 0.02,
      "error_types": {
        "validation_error": 0.01,
        "classification_error": 0.005,
        "internal_error": 0.005
      }
    },
    "resources": {
      "cpu_usage": 0.45,
      "memory_usage": 0.60,
      "disk_usage": 0.30
    }
  }
}
```

### GET /v1/monitoring/alerts

Retrieves active alerts.

**Response**:
```json
{
  "success": true,
  "data": {
    "alerts": [
      {
        "id": "alert_1234567890",
        "type": "performance",
        "severity": "warning",
        "message": "Response time exceeded threshold",
        "details": {
          "metric": "response_time_p95",
          "value": 350,
          "threshold": 300
        },
        "created_at": "2024-12-19T10:30:00Z",
        "status": "active"
      }
    ],
    "total_count": 1
  }
}
```

### POST /v1/monitoring/alerts/{alert_id}/acknowledge

Acknowledges an alert.

### GET /v1/monitoring/patterns

Retrieves pattern analysis results.

**Response**:
```json
{
  "success": true,
  "data": {
    "patterns": [
      {
        "id": "pattern_1234567890",
        "type": "misclassification",
        "severity": "medium",
        "description": "High misclassification rate for technology companies",
        "affected_businesses": 25,
        "confidence": 0.85,
        "recommendations": [
          "Update keyword database for technology sector",
          "Retrain ML model with more technology examples"
        ],
        "created_at": "2024-12-19T10:30:00Z"
      }
    ]
  }
}
```

## Health and Status Endpoints

### GET /v1/health

Retrieves system health status.

**Response**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "7d 12h 30m 15s",
    "timestamp": "2024-12-19T10:30:00Z",
    "components": {
      "database": {
        "status": "healthy",
        "response_time": "5ms"
      },
      "cache": {
        "status": "healthy",
        "hit_rate": 0.85
      },
      "external_apis": {
        "status": "healthy",
        "response_time": "150ms"
      }
    }
  }
}
```

### GET /v1/status

Retrieves detailed system status.

**Response**:
```json
{
  "success": true,
  "data": {
    "system": {
      "version": "1.0.0",
      "environment": "production",
      "region": "us-west-2",
      "instance_id": "i-1234567890abcdef0"
    },
    "performance": {
      "cpu_usage": 0.45,
      "memory_usage": 0.60,
      "disk_usage": 0.30,
      "network_io": "10MB/s"
    },
    "requests": {
      "total_requests": 1000000,
      "requests_per_second": 100,
      "error_rate": 0.02,
      "average_response_time": 180
    },
    "cache": {
      "total_entries": 1000,
      "hit_rate": 0.85,
      "memory_usage": "25MB",
      "disk_usage": "25MB"
    },
    "database": {
      "connections": 10,
      "active_queries": 5,
      "slow_queries": 0
    }
  }
}
```

### GET /v1/version

Retrieves API version information.

**Response**:
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "build_date": "2024-12-19T10:30:00Z",
    "git_commit": "abc123def456",
    "features": [
      "classification",
      "risk_assessment",
      "data_discovery",
      "caching",
      "monitoring"
    ],
    "deprecated_features": [],
    "upcoming_features": [
      "advanced_analytics",
      "machine_learning_enhancements"
    ]
  }
}
```

## SDK Examples

### Python SDK

```python
import kyb_client

# Initialize client
client = kyb_client.Client(api_key="YOUR_API_KEY")

# Classify business
result = client.classify(
    business_name="Acme Corporation",
    description="Technology consulting services",
    website="https://acme.com"
)

print(f"Primary classification: {result.primary_code.code}")
print(f"Confidence: {result.primary_code.confidence}")

# Assess risk
risk = client.assess_risk(
    business_name="Acme Corporation",
    website="https://acme.com"
)

print(f"Overall risk: {risk.overall_risk}")
print(f"Risk score: {risk.risk_score}")

# Discover data
discovery = client.discover_data(
    business_name="Acme Corporation",
    website="https://acme.com"
)

print(f"Company size: {discovery.team_info.size}")
print(f"Data quality: {discovery.data_quality.overall}")
```

### JavaScript SDK

```javascript
const { KYBClient } = require('kyb-client');

// Initialize client
const client = new KYBClient('YOUR_API_KEY');

// Classify business
const result = await client.classify({
    business_name: 'Acme Corporation',
    description: 'Technology consulting services',
    website: 'https://acme.com'
});

console.log(`Primary classification: ${result.primary_code.code}`);
console.log(`Confidence: ${result.primary_code.confidence}`);

// Assess risk
const risk = await client.assessRisk({
    business_name: 'Acme Corporation',
    website: 'https://acme.com'
});

console.log(`Overall risk: ${risk.overall_risk}`);
console.log(`Risk score: ${risk.risk_score}`);

// Discover data
const discovery = await client.discoverData({
    business_name: 'Acme Corporation',
    website: 'https://acme.com'
});

console.log(`Company size: ${discovery.team_info.size}`);
console.log(`Data quality: ${discovery.data_quality.overall}`);
```

## Webhook Integration

### Webhook Configuration

Configure webhooks to receive real-time notifications:

```json
{
  "url": "https://your-app.com/webhooks/kyb",
  "events": ["classification.completed", "risk_assessment.completed"],
  "secret": "webhook_secret_1234567890"
}
```

### Webhook Payload

```json
{
  "event": "classification.completed",
  "timestamp": "2024-12-19T10:30:00Z",
  "data": {
    "id": "class_1234567890",
    "business_name": "Acme Corporation",
    "classification": {
      "primary_code": {
        "type": "NAICS",
        "code": "541511",
        "description": "Custom Computer Programming Services",
        "confidence": 0.95
      }
    }
  }
}
```

## Rate Limiting

### Rate Limit Headers

All API responses include rate limit headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Reset-Time: 2024-12-19T11:30:00Z
```

### Rate Limit Exceeded Response

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later.",
    "details": {
      "limit": 1000,
      "reset_time": "2024-12-19T11:30:00Z"
    },
    "request_id": "req_1234567890",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

## Best Practices

### Request Optimization

1. **Use appropriate timeouts**: Set reasonable timeouts for your requests
2. **Implement retry logic**: Retry failed requests with exponential backoff
3. **Cache responses**: Cache responses to reduce API calls
4. **Batch requests**: Use batch endpoints when available

### Error Handling

1. **Check status codes**: Always check HTTP status codes
2. **Handle rate limits**: Implement rate limit handling
3. **Log errors**: Log errors for debugging and monitoring
4. **Provide fallbacks**: Implement fallback mechanisms

### Security

1. **Secure API keys**: Keep API keys secure and rotate regularly
2. **Use HTTPS**: Always use HTTPS for API calls
3. **Validate responses**: Validate API responses before processing
4. **Monitor usage**: Monitor API usage for anomalies

## Conclusion

This API reference provides comprehensive documentation for all endpoints in the Enhanced Business Intelligence System. The API is designed to be:

- **RESTful**: Follows REST principles for consistency
- **Secure**: Multiple authentication methods and security features
- **Scalable**: Rate limiting and performance optimization
- **Reliable**: Comprehensive error handling and monitoring
- **Extensible**: Versioned API with backward compatibility

For additional support, please refer to the SDK documentation or contact our support team.
