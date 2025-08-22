# API v3 Endpoints Documentation

## Overview

The v3 API provides enhanced business intelligence, observability, and enterprise integration capabilities. This documentation covers all the new endpoints implemented in the v3 API.

## Base URL

```
https://api.business-verification.com/api/v3
```

## Authentication

All endpoints require authentication via API key in the header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All endpoints return responses in the following format:

```json
{
  "success": true,
  "data": {},
  "meta": {
    "response_time": "150ms",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Dashboard Endpoints

### Get Dashboard Overview
**GET** `/dashboard`

Returns the main dashboard overview with key metrics.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

**Response:**
```json
{
  "success": true,
  "data": {
    "total_requests": 5000,
    "success_rate": 99.5,
    "avg_response_time": 120,
    "active_alerts": 3
  }
}
```

### Get Dashboard Metrics
**GET** `/dashboard/metrics`

Returns detailed dashboard metrics.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

**Response:**
```json
{
  "success": true,
  "data": {
    "performance": {
      "response_time": 120,
      "throughput": 200,
      "error_rate": 0.5
    },
    "business": {
      "active_users": 150,
      "total_verifications": 2500
    }
  }
}
```

### Get System Dashboard
**GET** `/dashboard/system`

Returns system-specific dashboard data.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

### Get Performance Dashboard
**GET** `/dashboard/performance`

Returns performance-specific dashboard data.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

### Get Business Dashboard
**GET** `/dashboard/business`

Returns business-specific dashboard data.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

## Alert Management Endpoints

### Get All Alerts
**GET** `/alerts`

Returns all configured alert rules.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "alert_20240115103000_123",
      "name": "High Error Rate",
      "description": "Alert when error rate exceeds 5%",
      "severity": "critical",
      "category": "errors",
      "enabled": true,
      "condition": "error_rate > 0.05",
      "threshold": 0.05,
      "duration": "5m"
    }
  ]
}
```

### Get Specific Alert
**GET** `/alerts/{id}`

Returns details for a specific alert rule.

**Path Parameters:**
- `id`: Alert rule ID

### Create Alert
**POST** `/alerts`

Creates a new alert rule.

**Request Body:**
```json
{
  "name": "High Response Time",
  "description": "Alert when response time exceeds 500ms",
  "severity": "warning",
  "category": "performance",
  "condition": "response_time > 500",
  "threshold": 500,
  "duration": "1m",
  "operator": ">",
  "labels": {
    "environment": "production"
  },
  "notifications": ["email", "slack"]
}
```

### Update Alert
**PUT** `/alerts/{id}`

Updates an existing alert rule.

**Path Parameters:**
- `id`: Alert rule ID

**Request Body:**
```json
{
  "enabled": false,
  "threshold": 600,
  "severity": "critical"
}
```

### Delete Alert
**DELETE** `/alerts/{id}`

Deletes an alert rule.

**Path Parameters:**
- `id`: Alert rule ID

### Get Alert History
**GET** `/alerts/history`

Returns alert history.

**Query Parameters:**
- `limit` (optional): Number of history entries (default: 100)

## Escalation Management Endpoints

### Get Escalation Policies
**GET** `/escalation/policies`

Returns all escalation policies.

### Get Specific Escalation Policy
**GET** `/escalation/policies/{id}`

Returns details for a specific escalation policy.

**Path Parameters:**
- `id`: Policy ID

### Create Escalation Policy
**POST** `/escalation/policies`

Creates a new escalation policy.

**Request Body:**
```json
{
  "name": "Critical Alert Escalation",
  "description": "Escalation policy for critical alerts",
  "levels": [
    {
      "level": 1,
      "delay": "5m",
      "notifications": ["email"],
      "recipients": ["oncall@company.com"]
    },
    {
      "level": 2,
      "delay": "15m",
      "notifications": ["email", "slack"],
      "recipients": ["manager@company.com"]
    }
  ]
}
```

### Update Escalation Policy
**PUT** `/escalation/policies/{id}`

Updates an existing escalation policy.

**Path Parameters:**
- `id`: Policy ID

### Delete Escalation Policy
**DELETE** `/escalation/policies/{id}`

Deletes an escalation policy.

**Path Parameters:**
- `id`: Policy ID

### Get Escalation History
**GET** `/escalation/history`

Returns escalation history.

### Trigger Escalation
**POST** `/escalation/trigger`

Manually triggers an escalation.

**Request Body:**
```json
{
  "alert_id": "alert_123",
  "policy_id": "policy_456",
  "reason": "Manual escalation triggered"
}
```

## Performance Monitoring Endpoints

### Get Performance Metrics
**GET** `/performance/metrics`

Returns performance metrics.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "1h")

### Get Detailed Performance Metrics
**GET** `/performance/metrics/detailed`

Returns detailed performance metrics.

### Get Performance Alerts
**GET** `/performance/alerts`

Returns performance-related alerts.

### Get Performance Trends
**GET** `/performance/trends`

Returns performance trends analysis.

### Trigger Performance Optimization
**POST** `/performance/optimize`

Triggers performance optimization.

**Request Body:**
```json
{
  "target_metrics": ["response_time", "throughput"],
  "constraints": {
    "max_cpu": 80,
    "max_memory": 85
  },
  "strategy": "aggressive",
  "dry_run": false
}
```

### Get Optimization History
**GET** `/performance/optimization/history`

Returns optimization history.

### Get Performance Benchmarks
**GET** `/performance/benchmarks`

Returns performance benchmarks.

## Error Tracking Endpoints

### Get All Errors
**GET** `/errors`

Returns all tracked errors.

### Get Specific Error
**GET** `/errors/{id}`

Returns details for a specific error.

**Path Parameters:**
- `id`: Error ID

### Create Error
**POST** `/errors`

Creates a new error entry.

**Request Body:**
```json
{
  "error_type": "validation_error",
  "error_message": "Invalid input format",
  "severity": "warning",
  "category": "input_validation",
  "component": "api_gateway",
  "endpoint": "/api/v3/classify",
  "user_id": "user_123",
  "request_id": "req_456",
  "context": {
    "input": "invalid_data"
  },
  "tags": {
    "environment": "production"
  }
}
```

### Get Errors by Severity
**GET** `/errors/severity/{severity}`

Returns errors filtered by severity.

**Path Parameters:**
- `severity`: Error severity level

### Get Errors by Category
**GET** `/errors/category/{category}`

Returns errors filtered by category.

**Path Parameters:**
- `category`: Error category

### Get Error Patterns
**GET** `/errors/patterns`

Returns error pattern analysis.

### Update Error Status
**PUT** `/errors/{id}/status`

Updates the status of an error.

**Path Parameters:**
- `id`: Error ID

**Request Body:**
```json
{
  "status": "resolved",
  "assigned_to": "developer@company.com",
  "resolution_note": "Fixed input validation logic"
}
```

## Business Intelligence Endpoints

### Get Business Metrics
**GET** `/analytics/business/metrics`

Returns business intelligence metrics.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

### Get Performance Analytics
**GET** `/analytics/performance`

Returns performance analytics.

**Query Parameters:**
- `time_range` (optional): Time range for analytics (default: "24h")

### Get System Analytics
**GET** `/analytics/system`

Returns system analytics.

**Query Parameters:**
- `time_range` (optional): Time range for analytics (default: "24h")

### Get Trend Analysis
**GET** `/analytics/trends`

Returns trend analysis.

**Query Parameters:**
- `time_range` (optional): Time range for trends (default: "7d")

### Get Custom Analytics
**POST** `/analytics/custom`

Returns custom analytics based on request parameters.

**Request Body:**
```json
{
  "time_range": "24h",
  "metrics": ["response_time", "throughput", "error_rate"],
  "dimensions": ["endpoint", "user_type"],
  "filters": {
    "environment": "production"
  },
  "granularity": "1h"
}
```

### Get Analytics Report
**GET** `/analytics/report`

Returns analytics report.

**Query Parameters:**
- `type` (optional): Report type (default: "summary")
- `time_range` (optional): Time range for report (default: "24h")

## Enterprise Integration Endpoints

### Get Integration Status
**GET** `/integrations/status`

Returns integration status.

**Query Parameters:**
- `type` (optional): Integration type (default: "all")

### Configure Integration
**POST** `/integrations/configure`

Configures an integration.

**Request Body:**
```json
{
  "integration_type": "webhook",
  "config": {
    "endpoint": "https://api.example.com/webhook",
    "timeout": 30,
    "retry_count": 3
  },
  "credentials": {
    "api_key": "your_api_key"
  },
  "webhook_url": "https://your-app.com/webhook",
  "filters": {
    "event_types": ["alert", "error"]
  }
}
```

### Test Integration
**POST** `/integrations/test`

Tests an integration configuration.

**Request Body:**
```json
{
  "integration_type": "webhook",
  "config": {
    "endpoint": "https://api.example.com/webhook"
  }
}
```

### Sync Data
**POST** `/integrations/sync`

Triggers data synchronization.

**Request Body:**
```json
{
  "integration_type": "database",
  "config": {
    "sync_mode": "incremental"
  }
}
```

### Handle Webhook
**POST** `/integrations/webhook`

Handles incoming webhook events.

**Request Body:**
```json
{
  "event_type": "alert_fired",
  "event_data": {
    "alert_id": "alert_123",
    "severity": "critical"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "source": "monitoring_system",
  "correlation_id": "corr_456"
}
```

### Get API Metrics
**GET** `/integrations/api-metrics`

Returns API usage metrics.

**Query Parameters:**
- `time_range` (optional): Time range for metrics (default: "24h")

### Get Integration Logs
**GET** `/integrations/logs`

Returns integration logs.

**Query Parameters:**
- `type` (optional): Integration type
- `limit` (optional): Number of log entries (default: 100)

## Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": "Error description",
  "meta": {
    "response_time": "50ms",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `422 Unprocessable Entity`: Validation error
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Rate Limiting

API requests are rate limited to:
- 1000 requests per minute per API key
- 10000 requests per hour per API key

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset time

## Pagination

For endpoints that return lists, pagination is supported:

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 100, max: 1000)

**Response Headers:**
- `X-Total-Count`: Total number of items
- `X-Page-Count`: Total number of pages

## Versioning

The v3 API is versioned and backward compatibility is maintained within major versions. The API version is included in the URL path.

## Support

For API support and questions:
- Email: api-support@business-verification.com
- Documentation: https://docs.business-verification.com/api/v3
- Status Page: https://status.business-verification.com
