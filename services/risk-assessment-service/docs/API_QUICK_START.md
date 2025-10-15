# API Quick Start Guide

## Overview

This guide will get you up and running with the Risk Assessment Service API in under 5 minutes. We'll cover authentication, basic risk assessment, and common use cases.

## Prerequisites

- API key from your KYB Platform account
- HTTP client (curl, Postman, or your preferred tool)
- Basic understanding of REST APIs

## Authentication

All API requests require authentication using your API key. Include it in the `Authorization` header:

```bash
Authorization: Bearer YOUR_API_KEY
```

### Getting Your API Key

1. Log into your KYB Platform dashboard
2. Navigate to **Settings** → **API Keys**
3. Click **Create New API Key**
4. Copy the generated key (you won't see it again!)

## Base URL

```
https://risk-assessment-service-production.up.railway.app/api/v1
```

## Quick Start Examples

### 1. Basic Risk Assessment

Perform a simple risk assessment for a business:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/assess" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "industry": "Technology",
    "country": "US",
    "phone": "+1-555-123-4567",
    "email": "contact@acme.com",
    "website": "https://www.acme.com"
  }'
```

**Response:**
```json
{
  "id": "risk_1703123456789",
  "business_id": "biz_123456789",
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
  "confidence_score": 0.85,
  "status": "completed",
  "created_at": "2023-12-21T10:30:00Z",
  "updated_at": "2023-12-21T10:30:00Z"
}
```

### 2. Get Risk Assessment by ID

Retrieve a previously created risk assessment:

```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/assess/risk_1703123456789" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 3. Risk Prediction

Generate future risk predictions using ML models:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/assess/risk_1703123456789/predict" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "horizon_months": 6,
    "scenarios": ["optimistic", "realistic", "pessimistic"]
  }'
```

**Response:**
```json
{
  "business_id": "biz_123456789",
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
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 4. Compliance Check

Perform compliance screening:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/compliance/check" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "industry": "Technology",
    "country": "US",
    "compliance_types": ["kyc", "aml", "sanctions"]
  }'
```

### 5. Sanctions Screening

Check against sanctions lists:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/sanctions/screen" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "business_address": "123 Main St, Anytown, ST 12345",
    "country": "US"
  }'
```

## Advanced Features

### Multi-Horizon Predictions

Generate predictions for multiple time horizons:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/risk/predict-advanced" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_id": "biz_123456789",
    "horizons": [3, 6, 12],
    "model_types": ["xgboost", "lstm", "ensemble"],
    "include_temporal_analysis": true,
    "include_confidence_intervals": true
  }'
```

### Explainable AI

Get detailed explanations of risk factors:

```bash
curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/explain/risk_1703123456789" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Scenario Analysis

Perform Monte Carlo scenario analysis:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/scenarios/analyze" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "business_id": "biz_123456789",
    "scenarios": [
      {
        "name": "optimistic",
        "probability": 0.2,
        "description": "Best-case scenario with favorable market conditions"
      },
      {
        "name": "realistic",
        "probability": 0.6,
        "description": "Most likely scenario based on current trends"
      },
      {
        "name": "pessimistic",
        "probability": 0.2,
        "description": "Worst-case scenario with economic downturn"
      }
    ],
    "monte_carlo_runs": 10000,
    "time_horizon": 12
  }'
```

## Error Handling

The API uses standard HTTP status codes and returns detailed error information:

### Common Error Responses

**400 Bad Request:**
```json
{
  "error": "Invalid request data",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "business_name",
    "message": "business_name is required"
  }
}
```

**401 Unauthorized:**
```json
{
  "error": "Invalid or missing API key",
  "code": "AUTHENTICATION_ERROR"
}
```

**429 Rate Limit Exceeded:**
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "limit": 100,
    "remaining": 0,
    "reset_time": "2024-01-15T11:00:00Z"
  }
}
```

## Rate Limits

- **Rate Limit**: 100 requests per minute per API key
- **Headers**: Rate limit information is included in response headers:
  - `X-RateLimit-Limit`: Maximum requests per minute
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Time when the rate limit resets

## Best Practices

### 1. Use HTTPS
Always use HTTPS for API requests to ensure data security.

### 2. Handle Errors Gracefully
Implement proper error handling in your application:

```javascript
try {
  const response = await fetch('/api/v1/assess', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(requestData)
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(`API Error: ${error.message}`);
  }
  
  const result = await response.json();
  console.log('Risk Assessment:', result);
} catch (error) {
  console.error('Error:', error.message);
}
```

### 3. Cache Results
Risk assessments are expensive operations. Cache results when appropriate:

```javascript
// Check cache first
const cached = localStorage.getItem(`risk_${businessId}`);
if (cached) {
  return JSON.parse(cached);
}

// Make API call
const result = await assessRisk(businessData);

// Cache for 1 hour
localStorage.setItem(`risk_${businessId}`, JSON.stringify(result));
setTimeout(() => {
  localStorage.removeItem(`risk_${businessId}`);
}, 3600000); // 1 hour
```

### 4. Use Webhooks for Real-time Updates
Set up webhooks to receive real-time notifications:

```bash
curl -X POST "https://risk-assessment-service-production.up.railway.app/api/v1/webhooks" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/risk-assessment",
    "events": ["assessment.completed", "prediction.updated"],
    "secret": "your_webhook_secret"
  }'
```

## SDKs and Libraries

We provide official SDKs for popular programming languages:

### Go SDK
```bash
go get github.com/kyb-platform/go-sdk
```

```go
import "github.com/kyb-platform/go-sdk"

client := kyb.NewClient(&kyb.Config{
    BaseURL: "https://risk-assessment-service-production.up.railway.app/api/v1",
    APIKey:  "YOUR_API_KEY",
})

assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
    BusinessName:    "Acme Corporation",
    BusinessAddress: "123 Main St, Anytown, ST 12345",
    Industry:        "Technology",
    Country:         "US",
})
```

### Python SDK
```bash
pip install kyb-sdk
```

```python
from kyb_sdk import KYBClient

client = KYBClient(api_key="YOUR_API_KEY")

assessment = client.assess_risk(
    business_name="Acme Corporation",
    business_address="123 Main St, Anytown, ST 12345",
    industry="Technology",
    country="US"
)
```

### Node.js SDK
```bash
npm install kyb-sdk
```

```javascript
const { KYBClient } = require('kyb-sdk');

const client = new KYBClient('YOUR_API_KEY');

const assessment = await client.assessRisk({
    businessName: 'Acme Corporation',
    businessAddress: '123 Main St, Anytown, ST 12345',
    industry: 'Technology',
    country: 'US'
});
```

## Next Steps

1. **Explore the Full API**: Check out the complete [API Documentation](API_DOCUMENTATION.md)
2. **Try Advanced Features**: Experiment with [explainable AI](API_DOCUMENTATION.md#explainability) and [scenario analysis](API_DOCUMENTATION.md#scenario-analysis)
3. **Set Up Webhooks**: Configure real-time notifications for your application
4. **Join the Community**: Get help and share experiences in our [developer community](https://community.kyb-platform.com)

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: [api-support@kyb-platform.com](mailto:api-support@kyb-platform.com)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)

## Changelog

### v2.0.0 (2024-01-15)
- **NEW**: Advanced multi-horizon predictions with LSTM models
- **NEW**: Explainable AI with SHAP-like feature contributions
- **NEW**: Scenario analysis with Monte Carlo simulations
- **NEW**: Industry-specific risk models (9 sectors)
- **NEW**: A/B testing framework for model validation
- **NEW**: Premium external API integrations
- **NEW**: Batch processing capabilities
- **NEW**: Webhook notifications
- **ENHANCED**: Performance scaling to 5000+ req/min
- **ENHANCED**: Global coverage for 10+ countries

### v1.0.0 (2024-01-15)
- Initial release
- Basic risk assessment endpoints
- Compliance checking
- Analytics and insights
