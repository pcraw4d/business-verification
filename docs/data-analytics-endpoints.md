# Data Analytics Endpoints

## Overview

The Data Analytics API provides comprehensive analytics capabilities for the KYB platform, allowing users to perform real-time analytics, metrics calculation, data aggregation, custom queries, trend analysis, and predictive analytics. This API supports both immediate analytics processing and background job processing for complex analytics operations.

## Base URL

```
https://api.kyb-platform.com/v1
```

## Authentication

All endpoints require API key authentication via the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All responses are returned in JSON format with the following structure:

```json
{
  "analytics_id": "analytics_1234567890_1",
  "business_id": "business_123",
  "type": "verification_trends",
  "status": "success",
  "is_successful": true,
  "results": [
    {
      "operation": "count",
      "field": "verifications",
      "value": 1500,
      "group_by": {
        "status": "completed"
      },
      "confidence": 0.98
    }
  ],
  "insights": [
    {
      "type": "trend",
      "title": "Increasing Verification Success Rate",
      "description": "Success rate has increased by 5% over the last 30 days",
      "severity": "low",
      "confidence": 0.85,
      "recommendation": "Continue monitoring the trend and investigate contributing factors"
    }
  ],
  "predictions": [
    {
      "field": "monthly_verifications",
      "predicted_value": 1800,
      "confidence": 0.92,
      "time_horizon": "30_days",
      "factors": ["seasonal_trends", "market_growth"],
      "range": {
        "min": 1700,
        "max": 1900,
        "percentile_25": 1750,
        "percentile_75": 1850
      }
    }
  ],
  "trends": [
    {
      "field": "verification_volume",
      "direction": "increasing",
      "slope": 0.15,
      "strength": 0.78,
      "time_range": {
        "start": "2024-11-19T10:30:00Z",
        "end": "2024-12-19T10:30:00Z"
      },
      "data_points": [
        {
          "timestamp": "2024-11-19T10:30:00Z",
          "value": 1200
        },
        {
          "timestamp": "2024-12-19T10:30:00Z",
          "value": 1500
        }
      ]
    }
  ],
  "correlations": [
    {
      "field1": "verification_volume",
      "field2": "success_rate",
      "coefficient": 0.65,
      "strength": "moderate",
      "significance": 0.01
    }
  ],
  "summary": {
    "total_records": 1500,
    "time_range": {
      "start": "2024-11-19T10:30:00Z",
      "end": "2024-12-19T10:30:00Z"
    },
    "key_metrics": {
      "total_verifications": 1500,
      "success_rate": 0.95,
      "average_processing_time": "2.5s"
    },
    "top_insights": [
      "Success rate is trending upward",
      "Verification volume is increasing"
    ],
    "recommendations": [
      "Monitor success rate trends",
      "Consider scaling resources for increased volume"
    ]
  },
  "metadata": { ... },
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "300ms"
}
```

## Supported Analytics Types

- `verification_trends` - Verification trends and patterns analysis
- `success_rates` - Success rate analysis and metrics
- `risk_distribution` - Risk distribution and assessment analysis
- `industry_analysis` - Industry-specific analytics and insights
- `geographic_analysis` - Geographic distribution and patterns
- `performance_metrics` - Performance and efficiency metrics
- `compliance_metrics` - Compliance and regulatory metrics
- `custom_query` - Custom analytics queries
- `predictive_analysis` - Predictive analytics and forecasting

## Supported Analytics Operations

- `count` - Count records or occurrences
- `sum` - Sum of numeric values
- `average` - Average of numeric values
- `median` - Median of numeric values
- `min` - Minimum value
- `max` - Maximum value
- `percentage` - Percentage calculations
- `trend` - Trend analysis and patterns
- `correlation` - Correlation analysis between fields
- `prediction` - Predictive analytics
- `anomaly_detection` - Anomaly detection and analysis

## Endpoints

### 1. Analyze Data

**POST** `/v1/analytics`

Performs immediate data analytics with the provided configuration.

#### Request Body

```json
{
  "business_id": "business_123",
  "analytics_type": "verification_trends",
  "operations": ["count", "trend"],
  "filters": {
    "status": "completed",
    "date_range": "last_30_days"
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "group_by": ["date", "industry"],
  "order_by": ["date DESC"],
  "limit": 100,
  "offset": 0,
  "custom_query": "SELECT COUNT(*) FROM verifications WHERE status = 'completed'",
  "parameters": {
    "confidence_level": 0.95,
    "trend_window": "30_days"
  },
  "include_insights": true,
  "include_predictions": true,
  "include_trends": true,
  "include_correlations": true,
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

#### Response

```json
{
  "analytics_id": "analytics_1234567890_1",
  "business_id": "business_123",
  "type": "verification_trends",
  "status": "success",
  "is_successful": true,
  "results": [
    {
      "operation": "count",
      "field": "verifications",
      "value": 1500,
      "group_by": {
        "status": "completed"
      },
      "confidence": 0.98
    },
    {
      "operation": "trend",
      "field": "verification_volume",
      "value": 0.15,
      "group_by": {
        "time_period": "daily"
      },
      "confidence": 0.85
    }
  ],
  "insights": [
    {
      "type": "trend",
      "title": "Increasing Verification Success Rate",
      "description": "Success rate has increased by 5% over the last 30 days",
      "severity": "low",
      "confidence": 0.85,
      "recommendation": "Continue monitoring the trend and investigate contributing factors"
    },
    {
      "type": "anomaly",
      "title": "Unusual Verification Volume Spike",
      "description": "Verification volume increased by 25% on December 15th",
      "severity": "medium",
      "confidence": 0.92,
      "recommendation": "Investigate the cause of the volume spike"
    }
  ],
  "predictions": [
    {
      "field": "monthly_verifications",
      "predicted_value": 1800,
      "confidence": 0.92,
      "time_horizon": "30_days",
      "factors": ["seasonal_trends", "market_growth"],
      "range": {
        "min": 1700,
        "max": 1900,
        "percentile_25": 1750,
        "percentile_75": 1850
      }
    }
  ],
  "trends": [
    {
      "field": "verification_volume",
      "direction": "increasing",
      "slope": 0.15,
      "strength": 0.78,
      "time_range": {
        "start": "2024-11-19T10:30:00Z",
        "end": "2024-12-19T10:30:00Z"
      },
      "data_points": [
        {
          "timestamp": "2024-11-19T10:30:00Z",
          "value": 1200
        },
        {
          "timestamp": "2024-12-19T10:30:00Z",
          "value": 1500
        }
      ]
    }
  ],
  "correlations": [
    {
      "field1": "verification_volume",
      "field2": "success_rate",
      "coefficient": 0.65,
      "strength": "moderate",
      "significance": 0.01,
      "data_points": [
        {
          "value1": 1000,
          "value2": 0.92
        },
        {
          "value1": 1500,
          "value2": 0.95
        }
      ]
    }
  ],
  "summary": {
    "total_records": 1500,
    "time_range": {
      "start": "2024-11-19T10:30:00Z",
      "end": "2024-12-19T10:30:00Z"
    },
    "key_metrics": {
      "total_verifications": 1500,
      "success_rate": 0.95,
      "average_processing_time": "2.5s"
    },
    "top_insights": [
      "Success rate is trending upward",
      "Verification volume is increasing"
    ],
    "recommendations": [
      "Monitor success rate trends",
      "Consider scaling resources for increased volume"
    ]
  },
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  },
  "generated_at": "2024-12-19T10:30:00Z",
  "processing_time": "300ms"
}
```

### 2. Create Analytics Job

**POST** `/v1/analytics/jobs`

Creates a background job for performing complex analytics operations.

#### Request Body

Same as the immediate analytics request.

#### Response

```json
{
  "job_id": "analytics_job_1234567890_1",
  "business_id": "business_123",
  "type": "verification_trends",
  "status": "pending",
  "progress": 0.0,
  "total_steps": 6,
  "current_step": 0,
  "step_description": "Initializing analytics job",
  "created_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 3. Get Analytics Job

**GET** `/v1/analytics/jobs?job_id={job_id}`

Retrieves the status and results of a background analytics job.

#### Response

```json
{
  "job_id": "analytics_job_1234567890_1",
  "business_id": "business_123",
  "type": "verification_trends",
  "status": "completed",
  "progress": 1.0,
  "total_steps": 6,
  "current_step": 6,
  "step_description": "Analytics completed successfully",
  "result": {
    "analytics_id": "analytics_1234567890_1",
    "business_id": "business_123",
    "type": "verification_trends",
    "status": "success",
    "is_successful": true,
    "results": [
      {
        "operation": "count",
        "field": "verifications",
        "value": 1500,
        "group_by": {
          "status": "completed"
        }
      }
    ],
    "insights": [
      {
        "type": "trend",
        "title": "Increasing Verification Success Rate",
        "description": "Success rate has increased by 5% over the last 30 days",
        "severity": "low",
        "confidence": 0.85
      }
    ],
    "predictions": [
      {
        "field": "monthly_verifications",
        "predicted_value": 1800,
        "confidence": 0.92,
        "time_horizon": "30_days"
      }
    ],
    "trends": [
      {
        "field": "verification_volume",
        "direction": "increasing",
        "slope": 0.15,
        "strength": 0.78
      }
    ],
    "correlations": [
      {
        "field1": "verification_volume",
        "field2": "success_rate",
        "coefficient": 0.65,
        "strength": "moderate"
      }
    ],
    "summary": {
      "total_records": 1500,
      "key_metrics": {
        "total_verifications": 1500,
        "success_rate": 0.95
      }
    },
    "generated_at": "2024-12-19T10:30:00Z",
    "processing_time": "500ms"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "started_at": "2024-12-19T10:30:01Z",
  "completed_at": "2024-12-19T10:30:05Z",
  "metadata": {
    "source": "verification_data",
    "version": "1.0"
  }
}
```

### 4. List Analytics Jobs

**GET** `/v1/analytics/jobs`

Lists all analytics jobs with optional filtering and pagination.

#### Query Parameters

- `status` (optional): Filter by job status (pending, processing, completed, failed, cancelled)
- `business_id` (optional): Filter by business ID
- `analytics_type` (optional): Filter by analytics type
- `limit` (optional): Number of jobs to return (default: 50, max: 100)
- `offset` (optional): Number of jobs to skip (default: 0)

#### Response

```json
{
  "jobs": [
    {
      "job_id": "analytics_job_1234567890_1",
      "business_id": "business_123",
      "type": "verification_trends",
      "status": "completed",
      "progress": 1.0,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

### 5. Get Analytics Schema

**GET** `/v1/analytics/schemas?schema_id={schema_id}`

Retrieves a pre-configured analytics schema.

#### Response

```json
{
  "id": "verification_trends_schema",
  "name": "Verification Trends Analysis",
  "description": "Analyze verification trends over time",
  "type": "verification_trends",
  "operations": ["count", "trend"],
  "default_filters": {
    "status": "completed"
  },
  "default_group_by": ["date"],
  "default_order_by": ["date"],
  "parameters": {
    "confidence_level": 0.95,
    "trend_window": "30_days"
  },
  "include_insights": true,
  "include_trends": true,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 6. List Analytics Schemas

**GET** `/v1/analytics/schemas`

Lists all available analytics schemas with optional filtering and pagination.

#### Query Parameters

- `analytics_type` (optional): Filter by analytics type
- `limit` (optional): Number of schemas to return (default: 50, max: 100)
- `offset` (optional): Number of schemas to skip (default: 0)

#### Response

```json
{
  "schemas": [
    {
      "id": "verification_trends_schema",
      "name": "Verification Trends Analysis",
      "description": "Analyze verification trends over time",
      "type": "verification_trends",
      "operations": ["count", "trend"],
      "include_insights": true,
      "include_trends": true,
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

## Configuration Options

### DataAnalyticsRequest

```json
{
  "business_id": "business_123",
  "analytics_type": "verification_trends",
  "operations": ["count", "trend", "correlation"],
  "filters": {
    "status": "completed",
    "date_range": "last_30_days",
    "industry": "technology"
  },
  "time_range": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-12-31T23:59:59Z"
  },
  "group_by": ["date", "industry", "region"],
  "order_by": ["date DESC", "count DESC"],
  "limit": 100,
  "offset": 0,
  "custom_query": "SELECT COUNT(*) FROM verifications WHERE status = 'completed'",
  "parameters": {
    "confidence_level": 0.95,
    "trend_window": "30_days",
    "correlation_threshold": 0.5,
    "prediction_horizon": "90_days"
  },
  "include_insights": true,
  "include_predictions": true,
  "include_trends": true,
  "include_correlations": true,
  "metadata": {
    "source": "verification_data",
    "version": "1.0",
    "generated_by": "user_123"
  }
}
```

### Analytics Types

#### Verification Trends
- **Purpose**: Analyze verification trends and patterns over time
- **Operations**: count, trend, correlation
- **Insights**: Trend identification, pattern recognition
- **Use Cases**: Performance monitoring, capacity planning

#### Success Rates
- **Purpose**: Analyze verification success rates and factors
- **Operations**: average, percentage, correlation
- **Insights**: Success factor analysis, improvement opportunities
- **Use Cases**: Quality assurance, process optimization

#### Risk Distribution
- **Purpose**: Analyze risk distribution across verifications
- **Operations**: count, average, correlation
- **Insights**: Risk pattern identification, risk factor analysis
- **Use Cases**: Risk management, compliance monitoring

#### Industry Analysis
- **Purpose**: Industry-specific analytics and insights
- **Operations**: count, average, trend, correlation
- **Insights**: Industry trends, benchmarking
- **Use Cases**: Market analysis, competitive intelligence

#### Geographic Analysis
- **Purpose**: Geographic distribution and patterns
- **Operations**: count, trend, correlation
- **Insights**: Geographic trends, regional patterns
- **Use Cases**: Market expansion, regional optimization

#### Performance Metrics
- **Purpose**: Performance and efficiency metrics
- **Operations**: average, min, max, trend
- **Insights**: Performance trends, optimization opportunities
- **Use Cases**: Performance monitoring, optimization

#### Compliance Metrics
- **Purpose**: Compliance and regulatory metrics
- **Operations**: count, percentage, trend
- **Insights**: Compliance trends, risk assessment
- **Use Cases**: Regulatory reporting, compliance monitoring

#### Custom Query
- **Purpose**: Custom analytics queries
- **Operations**: Any supported operation
- **Insights**: Custom insights based on query
- **Use Cases**: Custom analysis, specific requirements

#### Predictive Analysis
- **Purpose**: Predictive analytics and forecasting
- **Operations**: prediction, trend, correlation
- **Insights**: Future trends, predictions
- **Use Cases**: Forecasting, planning

## Analytics Operations

### Count
- **Purpose**: Count records or occurrences
- **Parameters**: field, group_by, filters
- **Output**: Numeric count with optional grouping

### Sum
- **Purpose**: Sum of numeric values
- **Parameters**: field, group_by, filters
- **Output**: Numeric sum with optional grouping

### Average
- **Purpose**: Average of numeric values
- **Parameters**: field, group_by, filters
- **Output**: Numeric average with optional grouping

### Median
- **Purpose**: Median of numeric values
- **Parameters**: field, group_by, filters
- **Output**: Numeric median with optional grouping

### Min/Max
- **Purpose**: Minimum or maximum values
- **Parameters**: field, group_by, filters
- **Output**: Numeric min/max with optional grouping

### Percentage
- **Purpose**: Percentage calculations
- **Parameters**: field, total_field, group_by, filters
- **Output**: Percentage values with optional grouping

### Trend
- **Purpose**: Trend analysis and patterns
- **Parameters**: field, time_field, window, group_by
- **Output**: Trend direction, slope, strength

### Correlation
- **Purpose**: Correlation analysis between fields
- **Parameters**: field1, field2, group_by, filters
- **Output**: Correlation coefficient, strength, significance

### Prediction
- **Purpose**: Predictive analytics
- **Parameters**: field, time_horizon, factors, confidence_level
- **Output**: Predicted values, confidence intervals, factors

### Anomaly Detection
- **Purpose**: Anomaly detection and analysis
- **Parameters**: field, threshold, window, group_by
- **Output**: Anomaly scores, detected anomalies

## Error Responses

### Validation Error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "business_id is required"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Job Not Found

```json
{
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Analytics job not found"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### Processing Error

```json
{
  "error": {
    "code": "ANALYTICS_ERROR",
    "message": "Failed to perform analytics"
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
// Perform immediate analytics
async function performAnalytics() {
  const response = await fetch('https://api.kyb-platform.com/v1/analytics', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      analytics_type: 'verification_trends',
      operations: ['count', 'trend'],
      include_insights: true,
      include_trends: true,
      filters: {
        status: 'completed'
      }
    })
  });

  const analyticsResult = await response.json();
  return analyticsResult;
}

// Create a background analytics job
async function createAnalyticsJob() {
  const response = await fetch('https://api.kyb-platform.com/v1/analytics/jobs', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      business_id: 'business_123',
      analytics_type: 'predictive_analysis',
      operations: ['prediction', 'correlation'],
      include_predictions: true,
      include_correlations: true,
      parameters: {
        prediction_horizon: '90_days',
        confidence_level: 0.95
      }
    })
  });

  const job = await response.json();
  return job;
}

// Poll job status
async function pollJobStatus(jobId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/analytics/jobs?job_id=${jobId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const job = await response.json();
  return job;
}

// Get analytics schema
async function getAnalyticsSchema(schemaId) {
  const response = await fetch(`https://api.kyb-platform.com/v1/analytics/schemas?schema_id=${schemaId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const schema = await response.json();
  return schema;
}

// List analytics jobs
async function listAnalyticsJobs(status, businessId, analyticsType) {
  const params = new URLSearchParams();
  if (status) params.append('status', status);
  if (businessId) params.append('business_id', businessId);
  if (analyticsType) params.append('analytics_type', analyticsType);

  const response = await fetch(`https://api.kyb-platform.com/v1/analytics/jobs?${params}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });

  const jobs = await response.json();
  return jobs;
}
```

### Python

```python
import requests
import json

# Perform immediate analytics
def perform_analytics():
    url = 'https://api.kyb-platform.com/v1/analytics'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'analytics_type': 'verification_trends',
        'operations': ['count', 'trend'],
        'include_insights': True,
        'include_trends': True,
        'filters': {
            'status': 'completed'
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Create a background analytics job
def create_analytics_job():
    url = 'https://api.kyb-platform.com/v1/analytics/jobs'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    }
    
    data = {
        'business_id': 'business_123',
        'analytics_type': 'predictive_analysis',
        'operations': ['prediction', 'correlation'],
        'include_predictions': True,
        'include_correlations': True,
        'parameters': {
            'prediction_horizon': '90_days',
            'confidence_level': 0.95
        }
    }
    
    response = requests.post(url, headers=headers, json=data)
    return response.json()

# Poll job status
def poll_job_status(job_id):
    url = f'https://api.kyb-platform.com/v1/analytics/jobs?job_id={job_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# Get analytics schema
def get_analytics_schema(schema_id):
    url = f'https://api.kyb-platform.com/v1/analytics/schemas?schema_id={schema_id}'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    response = requests.get(url, headers=headers)
    return response.json()

# List analytics jobs
def list_analytics_jobs(status=None, business_id=None, analytics_type=None):
    url = 'https://api.kyb-platform.com/v1/analytics/jobs'
    headers = {
        'Authorization': 'Bearer YOUR_API_KEY'
    }
    
    params = {}
    if status:
        params['status'] = status
    if business_id:
        params['business_id'] = business_id
    if analytics_type:
        params['analytics_type'] = analytics_type
    
    response = requests.get(url, headers=headers, params=params)
    return response.json()
```

### React Component

```jsx
import React, { useState, useEffect } from 'react';

const AnalyticsComponent = () => {
  const [analyticsData, setAnalyticsData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [jobStatus, setJobStatus] = useState(null);

  const performAnalytics = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/analytics', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          analytics_type: 'verification_trends',
          operations: ['count', 'trend'],
          include_insights: true,
          include_trends: true,
          filters: {
            status: 'completed'
          }
        })
      });

      const result = await response.json();
      setAnalyticsData(result);
    } catch (error) {
      console.error('Error performing analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  const createAnalyticsJob = async () => {
    setLoading(true);
    try {
      const response = await fetch('https://api.kyb-platform.com/v1/analytics/jobs', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer YOUR_API_KEY',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          business_id: 'business_123',
          analytics_type: 'predictive_analysis',
          operations: ['prediction', 'correlation'],
          include_predictions: true,
          include_correlations: true
        })
      });

      const job = await response.json();
      setJobStatus(job);
      
      // Start polling for job status
      pollJobStatus(job.job_id);
    } catch (error) {
      console.error('Error creating analytics job:', error);
    } finally {
      setLoading(false);
    }
  };

  const pollJobStatus = async (jobId) => {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`https://api.kyb-platform.com/v1/analytics/jobs?job_id=${jobId}`, {
          headers: {
            'Authorization': 'Bearer YOUR_API_KEY'
          }
        });

        const job = await response.json();
        setJobStatus(job);

        if (job.status === 'completed' || job.status === 'failed') {
          clearInterval(interval);
        }
      } catch (error) {
        console.error('Error polling job status:', error);
        clearInterval(interval);
      }
    }, 2000); // Poll every 2 seconds
  };

  return (
    <div>
      <h2>Data Analytics</h2>
      
      <button onClick={performAnalytics} disabled={loading}>
        {loading ? 'Analyzing...' : 'Perform Analytics'}
      </button>
      
      <button onClick={createAnalyticsJob} disabled={loading}>
        {loading ? 'Creating Job...' : 'Create Analytics Job'}
      </button>

      {analyticsData && (
        <div>
          <h3>Analytics Result</h3>
          <p>Analytics ID: {analyticsData.analytics_id}</p>
          <p>Type: {analyticsData.type}</p>
          <p>Status: {analyticsData.status}</p>
          <p>Processing Time: {analyticsData.processing_time}</p>
          
          <h4>Results</h4>
          {analyticsData.results.map((result, index) => (
            <div key={index}>
              <p>Operation: {result.operation}</p>
              <p>Field: {result.field}</p>
              <p>Value: {result.value}</p>
            </div>
          ))}
          
          {analyticsData.insights && analyticsData.insights.length > 0 && (
            <div>
              <h4>Insights</h4>
              {analyticsData.insights.map((insight, index) => (
                <div key={index}>
                  <p>Title: {insight.title}</p>
                  <p>Description: {insight.description}</p>
                  <p>Severity: {insight.severity}</p>
                  <p>Confidence: {insight.confidence}</p>
                </div>
              ))}
            </div>
          )}
          
          {analyticsData.predictions && analyticsData.predictions.length > 0 && (
            <div>
              <h4>Predictions</h4>
              {analyticsData.predictions.map((prediction, index) => (
                <div key={index}>
                  <p>Field: {prediction.field}</p>
                  <p>Predicted Value: {prediction.predicted_value}</p>
                  <p>Confidence: {prediction.confidence}</p>
                  <p>Time Horizon: {prediction.time_horizon}</p>
                </div>
              ))}
            </div>
          )}
          
          {analyticsData.trends && analyticsData.trends.length > 0 && (
            <div>
              <h4>Trends</h4>
              {analyticsData.trends.map((trend, index) => (
                <div key={index}>
                  <p>Field: {trend.field}</p>
                  <p>Direction: {trend.direction}</p>
                  <p>Slope: {trend.slope}</p>
                  <p>Strength: {trend.strength}</p>
                </div>
              ))}
            </div>
          )}
          
          {analyticsData.correlations && analyticsData.correlations.length > 0 && (
            <div>
              <h4>Correlations</h4>
              {analyticsData.correlations.map((correlation, index) => (
                <div key={index}>
                  <p>Field 1: {correlation.field1}</p>
                  <p>Field 2: {correlation.field2}</p>
                  <p>Coefficient: {correlation.coefficient}</p>
                  <p>Strength: {correlation.strength}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {jobStatus && (
        <div>
          <h3>Analytics Job Status</h3>
          <p>Job ID: {jobStatus.job_id}</p>
          <p>Status: {jobStatus.status}</p>
          <p>Progress: {(jobStatus.progress * 100).toFixed(1)}%</p>
          <p>Step: {jobStatus.step_description}</p>
          
          {jobStatus.result && (
            <div>
              <h4>Job Completed</h4>
              <p>Analytics ID: {jobStatus.result.analytics_id}</p>
              <p>Status: {jobStatus.result.status}</p>
              <p>Processing Time: {jobStatus.result.processing_time}</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default AnalyticsComponent;
```

## Best Practices

### 1. Analytics Design

- Choose appropriate analytics types for your use case
- Use multiple operations to get comprehensive insights
- Include relevant filters to focus on specific data
- Consider time ranges for trend analysis

### 2. Performance

- Use background jobs for complex analytics
- Implement proper polling mechanisms for job status
- Consider data size and complexity when choosing operations
- Use appropriate grouping and filtering

### 3. Insights and Predictions

- Enable insights for automatic pattern recognition
- Use predictions for forecasting and planning
- Monitor confidence levels for predictions
- Validate insights with domain knowledge

### 4. Error Handling

- Implement proper error handling for all API calls
- Handle job failures gracefully
- Monitor job progress and timeouts
- Log errors for debugging and monitoring

### 5. Security

- Validate all input data and parameters
- Implement proper access controls
- Use secure API keys and authentication
- Monitor for abuse and rate limiting

### 6. Monitoring

- Track analytics success rates
- Monitor job completion times
- Alert on failed analytics operations
- Monitor prediction accuracy

## Rate Limiting

- **Standard Analytics**: 20 requests per minute per API key
- **Background Jobs**: 5 job creations per minute per API key
- **Schema Retrieval**: 50 requests per minute per API key
- **Job Status Queries**: 100 requests per minute per API key

## Monitoring and Observability

### Key Metrics

- **Analytics Request Rate**: Number of analytics requests per minute
- **Success Rate**: Percentage of successful analytics operations
- **Processing Time**: Average time to complete analytics
- **Job Completion Rate**: Percentage of completed background jobs
- **Prediction Accuracy**: Accuracy of predictive analytics
- **Insight Quality**: Quality and relevance of generated insights

### Health Checks

Monitor the following endpoints for system health:

```bash
# Check analytics service health
curl -X GET "https://api.kyb-platform.com/v1/health/analytics" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Check background job processing
curl -X GET "https://api.kyb-platform.com/v1/health/jobs" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Logging

All analytics operations are logged with the following information:

- Request ID for correlation
- Business ID for tracking
- Analytics type and operations
- Processing time and performance metrics
- Error details and stack traces
- Insight and prediction quality metrics

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Ensure all required fields are provided
   - Check analytics type and operation compatibility
   - Verify custom query syntax for custom queries

2. **Job Failures**
   - Check job status and error messages
   - Verify data size and complexity
   - Monitor system resources

3. **Performance Issues**
   - Use background jobs for large datasets
   - Optimize filters and grouping
   - Consider data sampling for initial analysis

4. **Prediction Accuracy**
   - Monitor prediction confidence levels
   - Validate predictions with historical data
   - Adjust prediction parameters as needed

5. **Authentication Errors**
   - Verify API key is valid and active
   - Check API key permissions
   - Ensure proper header format

### Debug Information

Enable debug logging by including the `X-Debug` header:

```bash
curl -X POST "https://api.kyb-platform.com/v1/analytics" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-Debug: true" \
  -H "Content-Type: application/json" \
  -d '{"business_id": "business_123", "analytics_type": "verification_trends", "operations": ["count"], "include_insights": true}'
```

### Support

For additional support and troubleshooting:

- Check the API documentation for detailed endpoint information
- Review error logs and monitoring dashboards
- Contact support with request IDs and error details
- Provide reproducible examples for complex issues

## Migration Guide

### From Previous Versions

If migrating from previous analytics APIs:

1. **Update Endpoint URLs**: Use the new `/v1/analytics` endpoints
2. **Update Request Format**: Follow the new request structure
3. **Update Response Handling**: Handle the new response format
4. **Test Thoroughly**: Verify all analytics work correctly
5. **Update Documentation**: Update client documentation and examples

### Breaking Changes

- New authentication requirements
- Updated request/response formats
- New error codes and messages
- Enhanced validation rules
- Improved performance characteristics

## Future Enhancements

### Planned Features

1. **Real-time Analytics**: Streaming analytics for live data
2. **Advanced ML Models**: Machine learning-powered analytics
3. **Custom Algorithms**: User-defined analytics algorithms
4. **Analytics Dashboards**: Interactive analytics dashboards
5. **Collaborative Analytics**: Shared and collaborative analytics
6. **Analytics Notifications**: Email and webhook notifications for completed analytics
7. **Advanced Visualizations**: Interactive charts and graphs
8. **Analytics Versioning**: Version control for analytics configurations

### API Versioning

The analytics API follows semantic versioning:

- **v1**: Current stable version
- **v2**: Planned major version with new features
- **Beta**: Experimental features and endpoints

Check the API documentation for the latest version information and migration guides.
