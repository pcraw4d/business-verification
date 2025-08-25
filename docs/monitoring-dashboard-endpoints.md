# Monitoring Dashboard Endpoints Documentation

## Overview

The Monitoring Dashboard Endpoints provide a comprehensive API for accessing real-time monitoring data, system health metrics, performance indicators, business metrics, security information, and alerts. These endpoints serve as the backend for the KYB platform's monitoring dashboard, enabling real-time visibility into system operations and business performance.

## Base URL

```
https://api.kyb-platform.com/v3/dashboard
```

## Authentication

All dashboard endpoints require authentication using API keys or JWT tokens:

```bash
# Using API Key
Authorization: Bearer YOUR_API_KEY

# Using JWT Token
Authorization: Bearer YOUR_JWT_TOKEN
```

## Rate Limiting

- **Standard Rate Limit**: 100 requests per minute per API key
- **Real-time Updates**: 10 requests per minute per client
- **Export Endpoints**: 5 requests per minute per API key

## Endpoints

### 1. Get Complete Dashboard Data

Retrieves all dashboard data in a single request, including overview, system health, performance, business metrics, security metrics, and alerts.

**Endpoint**: `GET /dashboard/data`

**Response**: Complete dashboard data with all metrics

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/data" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json"
```

**Example Response**:
```json
{
  "overview": {
    "total_requests": 125000,
    "active_users": 45,
    "success_rate": 98.5,
    "average_response_time": 45.2,
    "uptime": 99.99,
    "system_status": "healthy"
  },
  "system_health": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.4,
    "network_latency": 12.5,
    "database_status": "healthy",
    "cache_status": "healthy",
    "external_apis": {
      "government_data": "healthy",
      "credit_bureau": "healthy",
      "risk_assessment": "healthy"
    }
  },
  "performance": {
    "request_rate": 1250.5,
    "error_rate": 0.15,
    "response_time_p50": 45.2,
    "response_time_p95": 120.8,
    "response_time_p99": 250.3,
    "throughput": 1250.5
  },
  "business": {
    "verifications_today": 1250,
    "verifications_this_week": 8750,
    "success_rate": 98.5,
    "average_processing_time": 2.3,
    "top_industries": [
      {
        "industry": "Technology",
        "count": 450,
        "success_rate": 99.2
      },
      {
        "industry": "Finance",
        "count": 320,
        "success_rate": 97.8
      },
      {
        "industry": "Healthcare",
        "count": 280,
        "success_rate": 98.9
      }
    ],
    "risk_distribution": {
      "low": 850,
      "medium": 320,
      "high": 80
    }
  },
  "security": {
    "failed_logins": 12,
    "blocked_requests": 45,
    "rate_limit_hits": 23,
    "security_alerts": 3,
    "last_security_scan": "2024-12-19T10:30:00Z"
  },
  "alerts": [
    {
      "id": "alert-001",
      "type": "performance",
      "severity": "warning",
      "message": "Response time exceeded threshold",
      "timestamp": "2024-12-19T10:00:00Z",
      "acknowledged": false
    },
    {
      "id": "alert-002",
      "type": "security",
      "severity": "info",
      "message": "Multiple failed login attempts detected",
      "timestamp": "2024-12-19T09:30:00Z",
      "acknowledged": true
    }
  ],
  "last_updated": "2024-12-19T10:35:00Z"
}
```

### 2. Get Dashboard Overview

Retrieves high-level system overview metrics.

**Endpoint**: `GET /dashboard/overview`

**Response**: Overview metrics only

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/overview" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "total_requests": 125000,
  "active_users": 45,
  "success_rate": 98.5,
  "average_response_time": 45.2,
  "uptime": 99.99,
  "system_status": "healthy"
}
```

### 3. Get System Health

Retrieves detailed system health metrics including CPU, memory, disk usage, and external API status.

**Endpoint**: `GET /dashboard/system-health`

**Response**: System health metrics

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/system-health" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "cpu_usage": 45.2,
  "memory_usage": 67.8,
  "disk_usage": 23.4,
  "network_latency": 12.5,
  "database_status": "healthy",
  "cache_status": "healthy",
  "external_apis": {
    "government_data": "healthy",
    "credit_bureau": "healthy",
    "risk_assessment": "healthy"
  }
}
```

### 4. Get Performance Metrics

Retrieves performance metrics including request rates, error rates, and response time percentiles.

**Endpoint**: `GET /dashboard/performance`

**Response**: Performance metrics

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/performance" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "request_rate": 1250.5,
  "error_rate": 0.15,
  "response_time_p50": 45.2,
  "response_time_p95": 120.8,
  "response_time_p99": 250.3,
  "throughput": 1250.5
}
```

### 5. Get Business Metrics

Retrieves business-specific metrics including verification counts, success rates, and industry breakdowns.

**Endpoint**: `GET /dashboard/business`

**Response**: Business metrics

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/business" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "verifications_today": 1250,
  "verifications_this_week": 8750,
  "success_rate": 98.5,
  "average_processing_time": 2.3,
  "top_industries": [
    {
      "industry": "Technology",
      "count": 450,
      "success_rate": 99.2
    },
    {
      "industry": "Finance",
      "count": 320,
      "success_rate": 97.8
    },
    {
      "industry": "Healthcare",
      "count": 280,
      "success_rate": 98.9
    }
  ],
  "risk_distribution": {
    "low": 850,
    "medium": 320,
    "high": 80
  }
}
```

### 6. Get Security Metrics

Retrieves security-related metrics including failed logins, blocked requests, and security alerts.

**Endpoint**: `GET /dashboard/security`

**Response**: Security metrics

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/security" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "failed_logins": 12,
  "blocked_requests": 45,
  "rate_limit_hits": 23,
  "security_alerts": 3,
  "last_security_scan": "2024-12-19T10:30:00Z"
}
```

### 7. Get Alerts

Retrieves current system alerts with severity levels and acknowledgment status.

**Endpoint**: `GET /dashboard/alerts`

**Response**: Array of alerts

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/alerts" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
[
  {
    "id": "alert-001",
    "type": "performance",
    "severity": "warning",
    "message": "Response time exceeded threshold",
    "timestamp": "2024-12-19T10:00:00Z",
    "acknowledged": false
  },
  {
    "id": "alert-002",
    "type": "security",
    "severity": "info",
    "message": "Multiple failed login attempts detected",
    "timestamp": "2024-12-19T09:30:00Z",
    "acknowledged": true
  }
]
```

### 8. Get Dashboard Configuration

Retrieves current dashboard configuration settings.

**Endpoint**: `GET /dashboard/config`

**Response**: Dashboard configuration

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/config" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "refresh_interval": 30,
  "theme": "light",
  "timezone": "UTC",
  "language": "en"
}
```

### 9. Update Dashboard Configuration

Updates dashboard configuration settings.

**Endpoint**: `PUT /dashboard/config`

**Request Body**: Dashboard configuration object

**Example Request**:
```bash
curl -X PUT "https://api.kyb-platform.com/v3/dashboard/config" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_interval": 60,
    "theme": "dark",
    "timezone": "America/New_York",
    "language": "es"
  }'
```

**Example Response**:
```json
{
  "status": "success",
  "message": "Configuration updated"
}
```

### 10. Get Real-time Updates

WebSocket endpoint for real-time dashboard updates (placeholder for future implementation).

**Endpoint**: `GET /dashboard/realtime`

**Response**: WebSocket upgrade response

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/dashboard/realtime" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response**:
```json
{
  "message": "WebSocket endpoint - upgrade logic to be implemented",
  "timestamp": "2024-12-19T10:35:00Z"
}
```

### 11. Export Dashboard Data

Exports dashboard data in various formats for external analysis.

**Endpoint**: `GET /dashboard/export`

**Query Parameters**:
- `format` (optional): Export format (`json`, `csv`). Default: `json`

**Example Request**:
```bash
# Export as JSON
curl -X GET "https://api.kyb-platform.com/v3/dashboard/export?format=json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  --output dashboard-data.json

# Export as CSV (not yet implemented)
curl -X GET "https://api.kyb-platform.com/v3/dashboard/export?format=csv" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Example Response** (JSON):
```json
{
  "overview": { ... },
  "system_health": { ... },
  "performance": { ... },
  "business": { ... },
  "security": { ... },
  "alerts": [ ... ],
  "last_updated": "2024-12-19T10:35:00Z"
}
```

## Error Responses

### Standard Error Format

All endpoints return errors in the following format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error details"
    }
  },
  "timestamp": "2024-12-19T10:35:00Z"
}
```

### Common Error Codes

| Status Code | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 429 | `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| 500 | `INTERNAL_ERROR` | Internal server error |

### Example Error Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later.",
    "details": {
      "limit": 100,
      "window": "1 minute",
      "retry_after": 30
    }
  },
  "timestamp": "2024-12-19T10:35:00Z"
}
```

## Integration Examples

### JavaScript/TypeScript Integration

```javascript
class DashboardAPI {
  constructor(apiKey, baseURL = 'https://api.kyb-platform.com/v3') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  async getDashboardData() {
    const response = await fetch(`${this.baseURL}/dashboard/data`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Dashboard API error: ${response.status}`);
    }

    return await response.json();
  }

  async getSystemHealth() {
    const response = await fetch(`${this.baseURL}/dashboard/system-health`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`System health API error: ${response.status}`);
    }

    return await response.json();
  }

  async updateConfig(config) {
    const response = await fetch(`${this.baseURL}/dashboard/config`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(config)
    });

    if (!response.ok) {
      throw new Error(`Config update error: ${response.status}`);
    }

    return await response.json();
  }
}

// Usage example
const dashboard = new DashboardAPI('your-api-key');

// Get complete dashboard data
dashboard.getDashboardData()
  .then(data => {
    console.log('Dashboard data:', data);
    updateDashboardUI(data);
  })
  .catch(error => {
    console.error('Error fetching dashboard data:', error);
  });

// Update configuration
dashboard.updateConfig({
  refresh_interval: 60,
  theme: 'dark',
  timezone: 'America/New_York'
})
  .then(result => {
    console.log('Config updated:', result);
  })
  .catch(error => {
    console.error('Error updating config:', error);
  });
```

### Python Integration

```python
import requests
import json
from typing import Dict, Any

class DashboardAPI:
    def __init__(self, api_key: str, base_url: str = 'https://api.kyb-platform.com/v3'):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def get_dashboard_data(self) -> Dict[str, Any]:
        """Get complete dashboard data"""
        response = requests.get(
            f'{self.base_url}/dashboard/data',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def get_system_health(self) -> Dict[str, Any]:
        """Get system health metrics"""
        response = requests.get(
            f'{self.base_url}/dashboard/system-health',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def get_business_metrics(self) -> Dict[str, Any]:
        """Get business metrics"""
        response = requests.get(
            f'{self.base_url}/dashboard/business',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def update_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Update dashboard configuration"""
        response = requests.put(
            f'{self.base_url}/dashboard/config',
            headers=self.headers,
            json=config
        )
        response.raise_for_status()
        return response.json()

    def export_data(self, format: str = 'json') -> bytes:
        """Export dashboard data"""
        response = requests.get(
            f'{self.base_url}/dashboard/export',
            headers=self.headers,
            params={'format': format}
        )
        response.raise_for_status()
        return response.content

# Usage example
dashboard = DashboardAPI('your-api-key')

try:
    # Get complete dashboard data
    data = dashboard.get_dashboard_data()
    print(f"Total requests: {data['overview']['total_requests']}")
    print(f"Success rate: {data['overview']['success_rate']}%")
    
    # Get system health
    health = dashboard.get_system_health()
    print(f"CPU usage: {health['cpu_usage']}%")
    print(f"Memory usage: {health['memory_usage']}%")
    
    # Get business metrics
    business = dashboard.get_business_metrics()
    print(f"Verifications today: {business['verifications_today']}")
    
    # Update configuration
    result = dashboard.update_config({
        'refresh_interval': 60,
        'theme': 'dark'
    })
    print(f"Config updated: {result['message']}")
    
    # Export data
    export_data = dashboard.export_data('json')
    with open('dashboard-export.json', 'wb') as f:
        f.write(export_data)
    print("Dashboard data exported successfully")

except requests.exceptions.RequestException as e:
    print(f"API request failed: {e}")
```

### React Integration

```jsx
import React, { useState, useEffect } from 'react';

const DashboardComponent = ({ apiKey }) => {
  const [dashboardData, setDashboardData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const response = await fetch('https://api.kyb-platform.com/v3/dashboard/data', {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      setDashboardData(data);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
    
    // Set up auto-refresh every 30 seconds
    const interval = setInterval(fetchDashboardData, 30000);
    
    return () => clearInterval(interval);
  }, [apiKey]);

  if (loading) {
    return <div>Loading dashboard data...</div>;
  }

  if (error) {
    return <div>Error loading dashboard: {error}</div>;
  }

  if (!dashboardData) {
    return <div>No dashboard data available</div>;
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>KYB Platform Dashboard</h1>
        <div className="system-status">
          Status: {dashboardData.overview.system_status}
        </div>
      </div>

      <div className="dashboard-grid">
        {/* Overview Cards */}
        <div className="card">
          <h3>Overview</h3>
          <div className="metric">
            <span>Total Requests:</span>
            <span>{dashboardData.overview.total_requests.toLocaleString()}</span>
          </div>
          <div className="metric">
            <span>Success Rate:</span>
            <span>{dashboardData.overview.success_rate}%</span>
          </div>
          <div className="metric">
            <span>Active Users:</span>
            <span>{dashboardData.overview.active_users}</span>
          </div>
        </div>

        {/* System Health */}
        <div className="card">
          <h3>System Health</h3>
          <div className="metric">
            <span>CPU Usage:</span>
            <span>{dashboardData.system_health.cpu_usage}%</span>
          </div>
          <div className="metric">
            <span>Memory Usage:</span>
            <span>{dashboardData.system_health.memory_usage}%</span>
          </div>
          <div className="metric">
            <span>Database:</span>
            <span className={`status-${dashboardData.system_health.database_status}`}>
              {dashboardData.system_health.database_status}
            </span>
          </div>
        </div>

        {/* Business Metrics */}
        <div className="card">
          <h3>Business Metrics</h3>
          <div className="metric">
            <span>Verifications Today:</span>
            <span>{dashboardData.business.verifications_today.toLocaleString()}</span>
          </div>
          <div className="metric">
            <span>Success Rate:</span>
            <span>{dashboardData.business.success_rate}%</span>
          </div>
        </div>

        {/* Alerts */}
        <div className="card">
          <h3>Alerts</h3>
          {dashboardData.alerts.length === 0 ? (
            <p>No active alerts</p>
          ) : (
            <ul className="alerts-list">
              {dashboardData.alerts.map(alert => (
                <li key={alert.id} className={`alert alert-${alert.severity}`}>
                  <strong>{alert.type}:</strong> {alert.message}
                  <small>{new Date(alert.timestamp).toLocaleString()}</small>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      <div className="dashboard-footer">
        <small>Last updated: {new Date(dashboardData.last_updated).toLocaleString()}</small>
      </div>
    </div>
  );
};

export default DashboardComponent;
```

## Monitoring and Alerting

### Dashboard Health Monitoring

Monitor the dashboard endpoints themselves to ensure they're functioning correctly:

```yaml
# Prometheus alerting rules
groups:
  - name: dashboard_endpoints
    rules:
      - alert: DashboardEndpointDown
        expr: up{job="kyb-api", endpoint=~"/dashboard/.*"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Dashboard endpoint {{ $labels.endpoint }} is down"
          description: "Dashboard endpoint {{ $labels.endpoint }} has been down for more than 1 minute"

      - alert: DashboardHighResponseTime
        expr: http_request_duration_seconds{job="kyb-api", endpoint=~"/dashboard/.*"} > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Dashboard endpoint {{ $labels.endpoint }} is slow"
          description: "Dashboard endpoint {{ $labels.endpoint }} is taking more than 2 seconds to respond"

      - alert: DashboardHighErrorRate
        expr: rate(http_requests_total{job="kyb-api", endpoint=~"/dashboard/.*", status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate on dashboard endpoints"
          description: "Dashboard endpoints are returning 5xx errors at a rate of {{ $value }} errors per second"
```

### Grafana Dashboard

Create a Grafana dashboard to monitor dashboard endpoint performance:

```json
{
  "dashboard": {
    "title": "KYB Dashboard Endpoints Monitoring",
    "panels": [
      {
        "title": "Dashboard Endpoints Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "http_request_duration_seconds{job=\"kyb-api\", endpoint=~\"/dashboard/.*\"}",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Dashboard Endpoints Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"kyb-api\", endpoint=~\"/dashboard/.*\"}[5m])",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Dashboard Endpoints Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"kyb-api\", endpoint=~\"/dashboard/.*\", status=~\"5..\"}[5m])",
            "legendFormat": "{{endpoint}}"
          }
        ]
      }
    ]
  }
}
```

## Best Practices

### 1. Caching Strategy

- **Client-side Caching**: Cache dashboard data for 30 seconds to reduce API calls
- **Server-side Caching**: The API implements 30-second TTL caching for dashboard data
- **Conditional Requests**: Use ETags and Last-Modified headers for efficient caching

```javascript
// Client-side caching example
class CachedDashboardAPI {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.cache = new Map();
    this.cacheTTL = 30000; // 30 seconds
  }

  async getDashboardData() {
    const now = Date.now();
    const cached = this.cache.get('dashboard_data');
    
    if (cached && (now - cached.timestamp) < this.cacheTTL) {
      return cached.data;
    }

    const data = await this.fetchDashboardData();
    this.cache.set('dashboard_data', {
      data,
      timestamp: now
    });

    return data;
  }
}
```

### 2. Error Handling

Implement robust error handling for dashboard API calls:

```javascript
class DashboardAPIWithRetry {
  constructor(apiKey, maxRetries = 3) {
    this.apiKey = apiKey;
    this.maxRetries = maxRetries;
  }

  async getDashboardData() {
    let lastError;
    
    for (let attempt = 1; attempt <= this.maxRetries; attempt++) {
      try {
        return await this.fetchDashboardData();
      } catch (error) {
        lastError = error;
        
        if (error.status === 429) {
          // Rate limited - wait and retry
          const retryAfter = parseInt(error.headers.get('Retry-After')) || 30;
          await this.sleep(retryAfter * 1000);
        } else if (error.status >= 500) {
          // Server error - retry with exponential backoff
          await this.sleep(Math.pow(2, attempt) * 1000);
        } else {
          // Client error - don't retry
          throw error;
        }
      }
    }
    
    throw lastError;
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### 3. Real-time Updates

For real-time dashboard updates, implement WebSocket connections:

```javascript
class RealTimeDashboard {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    this.ws = new WebSocket(`wss://api.kyb-platform.com/v3/dashboard/realtime`);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      
      // Send authentication
      this.ws.send(JSON.stringify({
        type: 'auth',
        apiKey: this.apiKey
      }));
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleUpdate(data);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.scheduleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  scheduleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.pow(2, this.reconnectAttempts) * 1000;
      
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, delay);
    }
  }

  handleUpdate(data) {
    switch (data.type) {
      case 'dashboard_update':
        this.updateDashboard(data.payload);
        break;
      case 'alert':
        this.showAlert(data.payload);
        break;
      case 'system_status':
        this.updateSystemStatus(data.payload);
        break;
    }
  }
}
```

### 4. Performance Optimization

- **Lazy Loading**: Load dashboard sections on demand
- **Data Compression**: Use gzip compression for API responses
- **Connection Pooling**: Reuse HTTP connections
- **Request Batching**: Combine multiple requests when possible

```javascript
// Lazy loading example
class LazyDashboard {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.loadedSections = new Set();
  }

  async loadSection(section) {
    if (this.loadedSections.has(section)) {
      return this.cache[section];
    }

    const data = await this.fetchSection(section);
    this.loadedSections.add(section);
    this.cache[section] = data;
    
    return data;
  }

  async fetchSection(section) {
    const endpoints = {
      overview: '/dashboard/overview',
      health: '/dashboard/system-health',
      performance: '/dashboard/performance',
      business: '/dashboard/business',
      security: '/dashboard/security',
      alerts: '/dashboard/alerts'
    };

    const response = await fetch(`https://api.kyb-platform.com/v3${endpoints[section]}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json'
      }
    });

    return response.json();
  }
}
```

## Troubleshooting

### Common Issues

1. **High Response Times**
   - Check server resource usage (CPU, memory, disk)
   - Verify database connection pool settings
   - Review external API response times
   - Consider implementing caching

2. **Authentication Errors**
   - Verify API key format and validity
   - Check API key permissions
   - Ensure proper Authorization header format
   - Verify API key hasn't expired

3. **Rate Limiting**
   - Implement exponential backoff
   - Use caching to reduce API calls
   - Consider upgrading API plan for higher limits
   - Implement request queuing

4. **Data Inconsistencies**
   - Check cache TTL settings
   - Verify data source synchronization
   - Review data collection intervals
   - Check for timezone issues

### Debugging Tools

```javascript
// Debug logging for dashboard API
class DebugDashboardAPI {
  constructor(apiKey, debug = false) {
    this.apiKey = apiKey;
    this.debug = debug;
  }

  async request(endpoint, options = {}) {
    const startTime = Date.now();
    
    if (this.debug) {
      console.log(`[Dashboard API] Requesting: ${endpoint}`);
    }

    try {
      const response = await fetch(`https://api.kyb-platform.com/v3${endpoint}`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
          ...options.headers
        },
        ...options
      });

      const duration = Date.now() - startTime;
      
      if (this.debug) {
        console.log(`[Dashboard API] Response: ${response.status} (${duration}ms)`);
      }

      if (!response.ok) {
        const errorText = await response.text();
        console.error(`[Dashboard API] Error: ${response.status} - ${errorText}`);
        throw new Error(`HTTP ${response.status}: ${errorText}`);
      }

      const data = await response.json();
      
      if (this.debug) {
        console.log(`[Dashboard API] Data size: ${JSON.stringify(data).length} bytes`);
      }

      return data;
    } catch (error) {
      if (this.debug) {
        console.error(`[Dashboard API] Request failed:`, error);
      }
      throw error;
    }
  }
}
```

## Migration Guide

### From Legacy Dashboard API

If migrating from a legacy dashboard API, follow these steps:

1. **Update Endpoint URLs**
   ```javascript
   // Old
   const oldEndpoint = 'https://api.kyb-platform.com/v2/dashboard';
   
   // New
   const newEndpoint = 'https://api.kyb-platform.com/v3/dashboard';
   ```

2. **Update Response Structure**
   ```javascript
   // Old response structure
   const oldData = {
     metrics: { ... },
     alerts: [ ... ]
   };
   
   // New response structure
   const newData = {
     overview: { ... },
     system_health: { ... },
     performance: { ... },
     business: { ... },
     security: { ... },
     alerts: [ ... ]
   };
   ```

3. **Update Error Handling**
   ```javascript
   // Old error format
   const oldError = {
     error: "Error message"
   };
   
   // New error format
   const newError = {
     error: {
       code: "ERROR_CODE",
       message: "Error message",
       details: { ... }
     }
   };
   ```

4. **Update Authentication**
   ```javascript
   // Old authentication
   headers: {
     'X-API-Key': apiKey
   }
   
   // New authentication
   headers: {
     'Authorization': `Bearer ${apiKey}`
   }
   ```

## Support

For technical support and questions about the Dashboard API:

- **Documentation**: [https://docs.kyb-platform.com/api/dashboard](https://docs.kyb-platform.com/api/dashboard)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Developer Community**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

## Changelog

### Version 3.0.0 (2024-12-19)
- Initial release of comprehensive dashboard endpoints
- Added unified dashboard data endpoint
- Implemented system health, performance, business, and security metrics
- Added dashboard configuration management
- Implemented data export functionality
- Added comprehensive error handling and rate limiting
- Added caching for improved performance
- Added comprehensive documentation and examples
