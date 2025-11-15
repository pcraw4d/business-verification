# Restored Endpoints Documentation

## Overview

This document provides comprehensive documentation for all restored risk management endpoints. These endpoints were restored as part of the functionality restoration migration.

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: Configured via environment variables

## Authentication

Most admin endpoints require authentication. Public endpoints are marked below.

## Error Response Format

All endpoints return consistent error responses:

```json
{
  "error": "Error message",
  "request_id": "uuid-or-header-value",
  "timestamp": "2025-01-27T12:00:00Z"
}
```

Error responses include the `X-Request-ID` header for request tracking.

## Risk Threshold Management

### GET /v1/risk/thresholds
Get all risk thresholds.

**Query Parameters:**
- `category` (optional): Filter by risk category (financial, operational, regulatory, etc.)
- `industry_code` (optional): Filter by industry code

**Response:**
```json
{
  "thresholds": [
    {
      "category": "financial",
      "low_max": 25.0,
      "medium_max": 50.0,
      "high_max": 75.0,
      "critical_min": 90.0,
      "updated_at": "2025-01-27T12:00:00Z"
    }
  ],
  "count": 1,
  "category": "financial",
  "industry": "",
  "timestamp": "2025-01:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Success

### POST /v1/admin/risk/thresholds
Create a new risk threshold configuration.

**Request Body:**
```json
{
  "name": "Financial Risk Threshold",
  "category": "financial",
  "risk_levels": {
    "low": 25.0,
    "medium": 50.0,
    "high": 75.0,
    "critical": 90.0
  },
  "is_active": true,
  "priority": 1,
  "description": "Threshold for financial risk assessment"
}
```

**Response:**
```json
{
  "id": "threshold-uuid",
  "config": { /* full threshold config */ },
  "timestamp": "2025-01-27T12:00:00Z"
}
```

**Status Codes:**
- `201 Created`: Success
- `400 Bad Request`: Invalid request data
- `503 Service Unavailable`: Threshold management service unavailable

### PUT /v1/admin/risk/thresholds/{threshold_id}
Update an existing risk threshold.

**Request Body:**
```json
{
  "name": "Updated Threshold Name",
  "description": "Updated description",
  "risk_levels": {
    "low": 30.0,
    "medium": 55.0,
    "high": 80.0,
    "critical": 95.0
  },
  "is_active": false
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Threshold not found
- `503 Service Unavailable`: Threshold management service unavailable

### DELETE /v1/admin/risk/thresholds/{threshold_id}
Delete a risk threshold configuration.

**Status Codes:**
- `200 OK`: Success
- `404 Not Found`: Threshold not found
- `503 Service Unavailable`: Threshold management service unavailable

## Threshold Export/Import

### GET /v1/admin/risk/threshold-export
Export all risk thresholds as JSON.

**Response:**
- Content-Type: `application/json`
- Content-Disposition: `attachment; filename=thresholds_export.json`
- Body: JSON array of threshold configurations

**Status Codes:**
- `200 OK`: Success
- `503 Service Unavailable`: Threshold management service unavailable

### POST /v1/admin/risk/threshold-import
Import risk thresholds from JSON.

**Request Body:**
```json
{
  "thresholds": [
    {
      "id": "threshold-uuid",
      "name": "Imported Threshold",
      "category": "financial",
      "risk_levels": { /* ... */ }
    }
  ]
}
```

**Response:**
```json
{
  "imported_count": 1,
  "errors": [],
  "timestamp": "2025-01-27T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid JSON or validation errors
- `503 Service Unavailable`: Threshold management service unavailable

## Risk Factors and Categories

### GET /v1/risk/factors
Get all risk factors.

**Query Parameters:**
- `category` (optional): Filter by risk category

**Response:**
```json
{
  "factors": [
    {
      "id": "financial_stability",
      "name": "Financial Stability",
      "description": "Measures the financial health and stability",
      "category": "financial",
      "weight": 0.3,
      "thresholds": {
        "low": 25.0,
        "medium": 50.0,
        "high": 75.0,
        "critical": 90.0
      },
      "created_at": "2025-01-27T12:00:00Z",
      "updated_at": "2025-01-27T12:00:00Z"
    }
  ],
  "count": 1,
  "timestamp": "2025-01-27T12:00:00Z"
}
```

### GET /v1/risk/categories
Get all risk categories.

**Response:**
```json
{
  "categories": [
    {
      "category": "financial",
      "name": "Financial Risk",
      "description": "Risks related to financial stability"
    }
  ],
  "count": 1,
  "timestamp": "2025-01-27T12:00:00Z"
}
```

## Recommendation Rules

### POST /v1/admin/risk/recommendation-rules
Create a new recommendation rule.

**Request Body:**
```json
{
  "name": "High Risk Review Rule",
  "category": "financial",
  "conditions": [
    {
      "factor": "risk_score",
      "operator": ">",
      "value": 75
    }
  ],
  "recommendations": [
    {
      "action": "review",
      "priority": "high",
      "message": "High risk detected - manual review required"
    }
  ],
  "enabled": true,
  "priority": 1
}
```

**Status Codes:**
- `201 Created`: Success
- `400 Bad Request`: Invalid request data
- `503 Service Unavailable`: Recommendation engine unavailable

### PUT /v1/admin/risk/recommendation-rules/{rule_id}
Update an existing recommendation rule.

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Rule not found
- `503 Service Unavailable`: Recommendation engine unavailable

### DELETE /v1/admin/risk/recommendation-rules/{rule_id}
Delete a recommendation rule.

**Status Codes:**
- `200 OK`: Success
- `404 Not Found`: Rule not found
- `503 Service Unavailable`: Recommendation engine unavailable

## Notification Channels

### POST /v1/admin/risk/notification-channels
Create a new notification channel.

**Supported Channel Types:**
- `email`: Email notifications
- `sms`: SMS notifications
- `slack`: Slack webhook
- `webhook`: Generic webhook
- `teams`: Microsoft Teams
- `discord`: Discord webhook
- `pagerduty`: PagerDuty integration
- `dashboard`: Dashboard notifications

**Request Body:**
```json
{
  "name": "email-alerts",
  "type": "email",
  "enabled": true,
  "config": {
    "recipients": ["admin@example.com"]
  }
}
```

**Status Codes:**
- `201 Created`: Success
- `400 Bad Request`: Invalid channel type or configuration
- `503 Service Unavailable`: Alert system unavailable

### PUT /v1/admin/risk/notification-channels/{channel_id}
Update an existing notification channel.

**Request Body:**
```json
{
  "enabled": false,
  "config": {
    "recipients": ["admin@example.com", "ops@example.com"]
  }
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid configuration
- `404 Not Found`: Channel not found
- `503 Service Unavailable`: Alert system unavailable

### DELETE /v1/admin/risk/notification-channels/{channel_id}
Delete a notification channel.

**Status Codes:**
- `200 OK`: Success
- `404 Not Found`: Channel not found
- `503 Service Unavailable`: Alert system unavailable

## System Monitoring

### GET /v1/admin/risk/system/health
Get system health status for risk management services.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-27T12:00:00Z",
  "services": {
    "risk_detection": "operational",
    "recommendation_engine": "operational",
    "trend_analysis": "operational",
    "alert_system": "operational"
  },
  "version": "1.0.0"
}
```

### GET /v1/admin/risk/system/metrics
Get system metrics for risk management services.

**Response:**
```json
{
  "timestamp": "2025-01-27T12:00:00Z",
  "assessments": {
    "total": 0,
    "completed": 0,
    "pending": 0,
    "failed": 0
  },
  "alerts": {
    "active": 0,
    "acknowledged": 0,
    "resolved": 0
  },
  "performance": {
    "avg_processing_time_ms": 0,
    "p95_processing_time_ms": 0,
    "p99_processing_time_ms": 0
  }
}
```

### POST /v1/admin/risk/system/cleanup
Cleanup old system data.

**Request Body:**
```json
{
  "older_than_days": 90,
  "data_types": ["alerts", "trends", "assessments"]
}
```

**Response:**
```json
{
  "cleaned": {
    "alerts": 0,
    "trends": 0,
    "assessments": 0
  },
  "older_than_days": 90,
  "data_types": ["alerts", "trends"],
  "message": "Cleanup completed successfully",
  "timestamp": "2025-01-27T12:00:00Z"
}
```

## Request ID Tracking

All endpoints support request ID tracking via the `X-Request-ID` header:

```bash
curl -H "X-Request-ID: my-custom-request-id" \
  http://localhost:8080/v1/risk/thresholds
```

If not provided, a request ID is generated automatically and returned in the response headers.

## Database Persistence

### With Database Configured
- Thresholds persist across server restarts
- Export/Import works with database storage
- All CRUD operations are persisted

### Without Database (In-Memory Fallback)
- Thresholds work in-memory only
- Data is lost on server restart
- Health check reports `postgres: not_configured`
- System gracefully degrades to in-memory mode

## Graceful Degradation

The system is designed to work with or without:
- **Database**: Falls back to in-memory storage
- **Redis**: Works without caching (no performance impact, just no caching)

Health check endpoint (`/health/detailed`) reports the status of all services.

## Rate Limiting

All endpoints are subject to rate limiting:
- Default: 100 requests per minute
- Burst size: 10 requests
- Rate limit information available at `/rate-limits`

## Examples

### Complete Workflow: Create, Export, Import Threshold

```bash
# 1. Create a threshold
THRESHOLD_ID=$(curl -s -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Threshold",
    "category": "financial",
    "risk_levels": {"low": 25, "medium": 50, "high": 75, "critical": 90}
  }' | jq -r '.id')

# 2. Export all thresholds
curl http://localhost:8080/v1/admin/risk/threshold-export > thresholds.json

# 3. Import thresholds (after modification)
curl -X POST http://localhost:8080/v1/admin/risk/threshold-import \
  -H "Content-Type: application/json" \
  -d @thresholds.json

# 4. Verify
curl http://localhost:8080/v1/risk/thresholds | jq '.'
```

### Error Handling Example

```bash
# Invalid request (missing required fields)
curl -X POST http://localhost:8080/v1/admin/risk/thresholds \
  -H "Content-Type: application/json" \
  -d '{}'
# Returns: 400 Bad Request

# Non-existent resource
curl -X GET http://localhost:8080/v1/admin/risk/thresholds/nonexistent-id
# Returns: 404 Not Found
```

## Testing

Comprehensive test scripts are available in the `test/` directory:
- `restoration_tests.sh`: Full test suite
- `test_database_persistence.sh`: Database persistence testing
- `test_graceful_degradation.sh`: Graceful degradation testing

See `test/README.md` for detailed testing instructions.

