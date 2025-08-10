# KYB Platform API Usage Examples

This document provides comprehensive examples for using the KYB Platform API endpoints. All examples use `curl` commands and include proper authentication, error handling, and response parsing.

## Table of Contents

1. [Authentication](#authentication)
2. [Business Classification](#business-classification)
3. [Risk Assessment](#risk-assessment)
4. [Compliance Checking](#compliance-checking)
5. [User Management](#user-management)
6. [Health and Monitoring](#health-and-monitoring)

---

## Authentication

### Register a New User

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!",
    "first_name": "John",
    "last_name": "Doe",
    "company": "Acme Corp"
  }'
```

**Response:**
```json
{
  "user_id": "user_123456789",
  "email": "user@example.com",
  "status": "pending_verification",
  "message": "Registration successful. Please check your email for verification."
}
```

### Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "user_id": "user_123456789",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "company": "Acme Corp",
    "role": "user"
  }
}
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Logout

```bash
curl -X POST http://localhost:8080/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

---

## Business Classification

### Single Business Classification

```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Software Solutions",
    "business_type": "technology",
    "industry": "software development",
    "description": "Enterprise software solutions for business automation",
    "keywords": ["software", "automation", "enterprise", "SaaS"]
  }'
```

**Response:**
```json
{
  "classification_id": "class_123456789",
  "business_name": "Acme Software Solutions",
  "primary_classification": {
    "code": "541511",
    "name": "Custom Computer Programming Services",
    "type": "NAICS",
    "confidence": 0.95
  },
  "secondary_classifications": [
    {
      "code": "541512",
      "name": "Computer Systems Design Services",
      "type": "NAICS",
      "confidence": 0.87
    }
  ],
  "crosswalk_codes": {
    "mcc": ["5734"],
    "sic": ["7371", "7372"]
  },
  "confidence_score": 0.91,
  "classification_methods": ["keyword_based", "business_type_based", "fuzzy_matching"],
  "processing_time_ms": 245
}
```

### Batch Classification

```bash
curl -X POST http://localhost:8080/v1/classify/batch \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "businesses": [
      {
        "business_name": "XYZ Corp",
        "business_type": "manufacturing"
      },
      {
        "business_name": "Law Office Associates",
        "industry": "legal services",
        "description": "Corporate law and litigation services"
      },
      {
        "business_name": "Tech Startup Inc",
        "business_type": "technology",
        "keywords": ["AI", "machine learning", "startup"]
      }
    ]
  }'
```

**Response:**
```json
{
  "batch_id": "batch_123456789",
  "total_businesses": 3,
  "processed_count": 3,
  "success_count": 3,
  "failed_count": 0,
  "results": [
    {
      "business_name": "XYZ Corp",
      "primary_classification": {
        "code": "332996",
        "name": "Fabricated Pipe and Pipe Fitting Manufacturing",
        "type": "NAICS",
        "confidence": 0.78
      }
    },
    {
      "business_name": "Law Office Associates",
      "primary_classification": {
        "code": "541110",
        "name": "Offices of Lawyers",
        "type": "NAICS",
        "confidence": 0.92
      }
    },
    {
      "business_name": "Tech Startup Inc",
      "primary_classification": {
        "code": "541715",
        "name": "Research and Development in the Physical, Engineering, and Life Sciences",
        "type": "NAICS",
        "confidence": 0.85
      }
    }
  ],
  "processing_time_ms": 892
}
```

### Get Classification History

```bash
curl -X GET "http://localhost:8080/v1/classify/history/XYZ%20Corp?limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "business_name": "XYZ Corp",
  "total_classifications": 5,
  "classifications": [
    {
      "classification_id": "class_123456789",
      "timestamp": "2024-01-15T10:30:00Z",
      "primary_classification": {
        "code": "332996",
        "name": "Fabricated Pipe and Pipe Fitting Manufacturing",
        "type": "NAICS",
        "confidence": 0.78
      }
    }
  ]
}
```

### Generate Confidence Report

```bash
curl -X POST http://localhost:8080/v1/classify/confidence-report \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "businesses": [
      {"business_name": "Acme Software Solutions"},
      {"business_name": "XYZ Corp"},
      {"business_name": "Law Office Associates"}
    ]
  }'
```

**Response:**
```json
{
  "report_id": "conf_report_123456789",
  "total_businesses": 3,
  "average_confidence": 0.85,
  "confidence_distribution": {
    "high": 2,
    "medium": 1,
    "low": 0
  },
  "classification_methods_summary": {
    "keyword_based": 3,
    "business_type_based": 3,
    "fuzzy_matching": 2
  }
}
```

---

## Risk Assessment

### Single Business Risk Assessment

```bash
curl -X POST http://localhost:8080/v1/risk/assess \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "business_id": "business_123456789",
    "business_name": "Acme Software Solutions",
    "categories": ["financial", "operational", "regulatory"],
    "industry": "technology",
    "business_type": "corporation",
    "annual_revenue": 5000000,
    "employee_count": 50,
    "years_in_business": 8
  }'
```

**Response:**
```json
{
  "assessment_id": "assessment_123456789",
  "business_id": "business_123456789",
  "business_name": "Acme Software Solutions",
  "overall_risk_score": 0.23,
  "risk_level": "low",
  "assessment_date": "2024-01-15T10:30:00Z",
  "categories": {
    "financial": {
      "score": 0.15,
      "level": "low",
      "factors": [
        {
          "factor": "revenue_stability",
          "score": 0.12,
          "description": "Stable revenue growth over 3 years"
        }
      ]
    },
    "operational": {
      "score": 0.28,
      "level": "low",
      "factors": [
        {
          "factor": "business_continuity",
          "score": 0.25,
          "description": "Established business with 8 years of operation"
        }
      ]
    },
    "regulatory": {
      "score": 0.31,
      "level": "medium",
      "factors": [
        {
          "factor": "compliance_status",
          "score": 0.31,
          "description": "Standard regulatory requirements for technology sector"
        }
      ]
    }
  },
  "recommendations": [
    {
      "category": "regulatory",
      "priority": "medium",
      "description": "Consider implementing SOC 2 compliance framework",
      "impact": "Reduce regulatory risk score by 15%"
    }
  ]
}
```

### Get Risk Categories

```bash
curl -X GET http://localhost:8080/v1/risk/categories \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "categories": [
    {
      "id": "financial",
      "name": "Financial Risk",
      "description": "Risks related to financial stability and performance",
      "factors": ["revenue_stability", "profitability", "cash_flow", "debt_levels"]
    },
    {
      "id": "operational",
      "name": "Operational Risk",
      "description": "Risks related to business operations and processes",
      "factors": ["business_continuity", "process_efficiency", "technology_reliability"]
    },
    {
      "id": "regulatory",
      "name": "Regulatory Risk",
      "description": "Risks related to compliance and regulatory requirements",
      "factors": ["compliance_status", "licensing", "regulatory_changes"]
    }
  ],
  "total": 5
}
```

### Get Risk Factors

```bash
curl -X GET "http://localhost:8080/v1/risk/factors?category=financial" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "factors": [
    {
      "id": "revenue_stability",
      "name": "Revenue Stability",
      "category": "financial",
      "description": "Assessment of revenue consistency and growth patterns",
      "weight": 0.25,
      "calculation_method": "statistical_analysis"
    },
    {
      "id": "profitability",
      "name": "Profitability",
      "category": "financial",
      "description": "Analysis of profit margins and financial performance",
      "weight": 0.20,
      "calculation_method": "ratio_analysis"
    }
  ],
  "total": 4,
  "category": "financial"
}
```

---

## Compliance Checking

### Check Compliance Status

```bash
curl -X POST http://localhost:8080/v1/compliance/check \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "business_id": "business_123456789",
    "frameworks": ["SOC2", "PCI_DSS", "GDPR"],
    "include_details": true
  }'
```

**Response:**
```json
{
  "check_id": "check_123456789",
  "business_id": "business_123456789",
  "overall_status": "compliant",
  "compliance_score": 0.87,
  "frameworks": {
    "SOC2": {
      "status": "compliant",
      "score": 0.92,
      "requirements_met": 18,
      "total_requirements": 20,
      "last_assessment": "2024-01-10T15:30:00Z"
    },
    "PCI_DSS": {
      "status": "partially_compliant",
      "score": 0.78,
      "requirements_met": 7,
      "total_requirements": 9,
      "last_assessment": "2024-01-12T09:15:00Z"
    },
    "GDPR": {
      "status": "compliant",
      "score": 0.91,
      "requirements_met": 15,
      "total_requirements": 16,
      "last_assessment": "2024-01-08T14:20:00Z"
    }
  },
  "recommendations": [
    {
      "framework": "PCI_DSS",
      "priority": "high",
      "description": "Implement additional data encryption measures",
      "impact": "Improve PCI DSS compliance score by 15%"
    }
  ]
}
```

### Get Compliance Status

```bash
curl -X GET "http://localhost:8080/v1/compliance/status/business_123456789" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Generate Compliance Report

```bash
curl -X POST "http://localhost:8080/v1/compliance/status/business_123456789/report" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "detailed",
    "frameworks": ["SOC2", "PCI_DSS"],
    "include_recommendations": true
  }'
```

---

## User Management

### Get User Profile

```bash
curl -X GET http://localhost:8080/v1/users/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "user_id": "user_123456789",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "company": "Acme Corp",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z",
  "last_login": "2024-01-15T10:30:00Z",
  "status": "active"
}
```

### Update User Profile

```bash
curl -X PUT http://localhost:8080/v1/users/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Smith",
    "company": "Acme Corporation"
  }'
```

### Change Password

```bash
curl -X POST http://localhost:8080/v1/users/change-password \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePassword123!",
    "new_password": "NewSecurePassword456!"
  }'
```

---

## Health and Monitoring

### Health Check

```bash
curl -X GET http://localhost:8080/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "database": {
      "status": "healthy",
      "response_time_ms": 12
    },
    "classification_service": {
      "status": "healthy",
      "response_time_ms": 45
    },
    "risk_service": {
      "status": "healthy",
      "response_time_ms": 78
    }
  }
}
```

### Get Metrics

```bash
curl -X GET http://localhost:8080/v1/metrics \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "api_requests_total": 15420,
  "api_requests_duration_seconds": 0.245,
  "classification_requests_total": 8920,
  "classification_accuracy": 0.96,
  "risk_assessments_total": 3420,
  "compliance_checks_total": 1280,
  "active_users": 156,
  "system_uptime_seconds": 86400
}
```

### Get Data Source Health

```bash
curl -X GET http://localhost:8080/v1/datasources/health \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "overall_status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "sources": {
    "business_database": {
      "status": "healthy",
      "response_time_ms": 25,
      "last_check": "2024-01-15T10:29:55Z"
    },
    "financial_data_provider": {
      "status": "healthy",
      "response_time_ms": 180,
      "last_check": "2024-01-15T10:29:50Z"
    },
    "regulatory_database": {
      "status": "degraded",
      "response_time_ms": 1200,
      "last_check": "2024-01-15T10:29:45Z",
      "error": "High latency detected"
    }
  }
}
```

---

## Error Handling Examples

### Invalid Authentication

```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Authorization: Bearer invalid_token" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company"}'
```

**Response:**
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired authentication token",
  "status_code": 401,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Validation Error

```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Response:**
```json
{
  "error": "validation_error",
  "message": "Business name is required",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "business_name",
    "constraint": "required"
  }
}
```

### Rate Limiting

```bash
# After exceeding rate limit
curl -X POST http://localhost:8080/v1/classify \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company"}'
```

**Response:**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Please try again later.",
  "status_code": 429,
  "timestamp": "2024-01-15T10:30:00Z",
  "retry_after": 60
}
```

---

## SDK Examples

### Python SDK

```python
from kyb_client import KYBClient

# Initialize client
client = KYBClient(
    base_url="http://localhost:8080",
    api_key="your_api_key"
)

# Classify a business
result = client.classify_business(
    business_name="Acme Software Solutions",
    business_type="technology"
)

print(f"Classification: {result.primary_classification.name}")
print(f"Confidence: {result.confidence_score}")

# Assess risk
risk = client.assess_risk(
    business_id="business_123",
    categories=["financial", "operational"]
)

print(f"Risk Level: {risk.risk_level}")
print(f"Risk Score: {risk.overall_risk_score}")
```

### JavaScript SDK

```javascript
const { KYBClient } = require('kyb-client');

// Initialize client
const client = new KYBClient({
  baseUrl: 'http://localhost:8080',
  apiKey: 'your_api_key'
});

// Classify a business
const result = await client.classifyBusiness({
  businessName: 'Acme Software Solutions',
  businessType: 'technology'
});

console.log(`Classification: ${result.primaryClassification.name}`);
console.log(`Confidence: ${result.confidenceScore}`);

// Assess risk
const risk = await client.assessRisk({
  businessId: 'business_123',
  categories: ['financial', 'operational']
});

console.log(`Risk Level: ${risk.riskLevel}`);
console.log(`Risk Score: ${risk.overallRiskScore}`);
```

---

## Best Practices

1. **Authentication**: Always include valid authentication tokens in requests
2. **Rate Limiting**: Implement exponential backoff for rate-limited requests
3. **Error Handling**: Check response status codes and handle errors gracefully
4. **Caching**: Cache classification results when possible to improve performance
5. **Batch Processing**: Use batch endpoints for processing multiple businesses
6. **Monitoring**: Monitor API response times and error rates
7. **Security**: Never log or store authentication tokens
8. **Validation**: Validate input data before sending to the API

---

## Support

For additional support and documentation:

- **API Documentation**: `/docs` endpoint for interactive documentation
- **Status Page**: Check system status at `/health`
- **Contact**: support@kybplatform.com
- **Documentation**: https://docs.kybplatform.com
