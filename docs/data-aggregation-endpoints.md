# Data Aggregation Endpoints

## Overview

The Data Aggregation API provides comprehensive data aggregation capabilities for the KYB platform, allowing users to perform various aggregation operations on business data including metrics, risk assessments, compliance reports, and performance analytics.

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
  "aggregation_id": "agg_1234567890_1",
  "business_id": "business_123",
  "aggregation_type": "business_metrics",
  "status": "success",
  "is_successful": true,
  "aggregated_data": { ... },
  "summary": { ... },
  "aggregated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

## Supported Aggregation Types

- `business_metrics` - Business performance metrics and KPIs
- `risk_assessments` - Risk assessment data aggregation
- `compliance_reports` - Compliance and regulatory reporting
- `performance_analytics` - Performance and efficiency analytics
- `trend_analysis` - Trend and pattern analysis
- `custom` - Custom aggregation rules
- `all` - All aggregation types

## Supported Aggregation Operations

- `count` - Count occurrences
- `sum` - Sum numeric values
- `average` - Calculate average
- `min` - Find minimum value
- `max` - Find maximum value
- `median` - Calculate median
- `percentile` - Calculate percentile (requires parameter)
- `group_by` - Group data by field
- `pivot` - Create pivot tables
- `custom` - Custom aggregation logic

## Endpoints

### 1. Aggregate Data

**POST** `/v1/aggregate`

Performs immediate data aggregation with the provided rules or schema.

#### Request Body

```json
{
  "business_id": "business_123",
  "aggregation_type": "business_metrics",
  "data": {
    "verifications": [
      {"status": "passed", "score": 0.95, "created_at": "2024-12-01T10:00:00Z"},
      {"status": "failed", "score": 0.30, "created_at": "2024-12-02T11:00:00Z"},
      {"status": "passed", "score": 0.88, "created_at": "2024-12-03T12:00:00Z"}
    ]
  },
  "rules": [
    {
      "field": "verification_count",
      "operation": "count",
      "parameters": {},
      "description": "Count total verifications",
      "enabled": true,
      "order": 1
    },
    {
      "field": "success_rate",
      "operation": "average",
      "parameters": {},
      "description": "Calculate average success rate",
      "enabled": true,
      "order": 2
    }
  ],
  "schema_id": "business_metrics_default",
  "group_by": ["status"],
  "filters": {
    "date_range": "last_30_days"
  },
  "time_range": {
    "start": "2024-12-01T00:00:00Z",
    "end": "2024-12-19T23:59:59Z"
  },
  "include_metadata": true,
  "metadata": {
    "request_id": "req_123",
    "user_id": "user_456"
  }
}
```

#### Response

```json
{
  "aggregation_id": "agg_1234567890_1",
  "business_id": "business_123",
  "aggregation_type": "business_metrics",
  "status": "success",
  "is_successful": true,
  "original_data": { ... },
  "aggregated_data": {
    "verification_count": 100,
    "success_rate": 0.85,
    "average_score": 0.78
  },
  "applied_rules": [
    {
      "field": "verification_count",
      "operation": "count",
      "description": "Count total verifications",
      "enabled": true,
      "order": 1
    }
  ],
  "summary": {
    "total_rules": 2,
    "applied_rules": 2,
    "skipped_rules": 0,
    "failed_rules": 0,
    "success_rate": 1.0,
    "data_count": 100,
    "processing_time": "150ms"
  },
  "grouped_results": {
    "passed": {
      "count": 85,
      "average_score": 0.88
    },
    "failed": {
      "count": 15,
      "average_score": 0.45
    }
  },
  "time_range": {
    "start": "2024-12-01T00:00:00Z",
    "end": "2024-12-19T23:59:59Z"
  },
  "aggregated_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms",
  "metadata": {
    "request_id": "req_123",
    "user_id": "user_456"
  }
}
```

### 2. Create Aggregation Job

**POST** `/v1/aggregate/job`

Creates a background aggregation job for processing large datasets.

#### Request Body

Same as the immediate aggregation endpoint.

#### Response

```json
{
  "job_id": "agg_job_1702998000_1",
  "status": "pending",
  "created_at": "2024-12-19T10:30:00Z",
  "message": "Aggregation job created successfully"
}
```

### 3. Get Aggregation Job

**GET** `/v1/aggregate/job/{job_id}`

Retrieves the status and results of an aggregation job.

#### Response

```json
{
  "id": "agg_job_1702998000_1",
  "business_id": "business_123",
  "aggregation_type": "business_metrics",
  "status": "completed",
  "progress": 100,
  "created_at": "2024-12-19T10:30:00Z",
  "started_at": "2024-12-19T10:30:02Z",
  "completed_at": "2024-12-19T10:30:05Z",
  "result": {
    "aggregation_id": "agg_1234567890_1",
    "business_id": "business_123",
    "aggregation_type": "business_metrics",
    "status": "success",
    "is_successful": true,
    "aggregated_data": { ... },
    "summary": { ... },
    "aggregated_at": "2024-12-19T10:30:05Z",
    "processing_time": "3s"
  },
  "metadata": {
    "request_id": "req_123",
    "user_id": "user_456"
  }
}
```

### 4. List Aggregation Jobs

**GET** `/v1/aggregate/jobs`

Lists aggregation jobs with optional filtering and pagination.

#### Query Parameters

- `business_id` - Filter by business ID
- `status` - Filter by job status (pending, processing, completed, failed)
- `aggregation_type` - Filter by aggregation type
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

#### Response

```json
{
  "jobs": [
    {
      "id": "agg_job_1702998000_1",
      "business_id": "business_123",
      "aggregation_type": "business_metrics",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

### 5. Get Aggregation Schema

**GET** `/v1/aggregate/schema/{schema_id}`

Retrieves a specific aggregation schema.

#### Response

```json
{
  "id": "business_metrics_default",
  "name": "Default Business Metrics",
  "description": "Default schema for business metrics aggregation",
  "type": "business_metrics",
  "rules": [
    {
      "field": "verification_count",
      "operation": "count",
      "parameters": {},
      "description": "Count total verifications",
      "enabled": true,
      "order": 1
    },
    {
      "field": "success_rate",
      "operation": "average",
      "parameters": {},
      "description": "Calculate average success rate",
      "enabled": true,
      "order": 2
    }
  ],
  "version": "1.0.0",
  "created_at": "2024-12-19T10:00:00Z",
  "updated_at": "2024-12-19T10:00:00Z"
}
```

### 6. List Aggregation Schemas

**GET** `/v1/aggregate/schemas`

Lists available aggregation schemas.

#### Query Parameters

- `type` - Filter by aggregation type

#### Response

```json
{
  "schemas": [
    {
      "id": "business_metrics_default",
      "name": "Default Business Metrics",
      "description": "Default schema for business metrics aggregation",
      "type": "business_metrics",
      "version": "1.0.0",
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:00:00Z"
    }
  ],
  "total": 1
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid request: aggregation_type is required",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "aggregation_type",
    "message": "aggregation_type is required"
  }
}
```

### 401 Unauthorized

```json
{
  "error": "Invalid or missing API key",
  "code": "AUTHENTICATION_ERROR"
}
```

### 404 Not Found

```json
{
  "error": "Job not found",
  "code": "NOT_FOUND"
}
```

### 429 Too Many Requests

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

### 500 Internal Server Error

```json
{
  "error": "Aggregation failed: internal error",
  "code": "INTERNAL_ERROR"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
// Immediate aggregation
const response = await fetch('https://api.kyb-platform.com/v1/aggregate', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json',
    'X-Business-ID': 'business_123'
  },
  body: JSON.stringify({
    aggregation_type: 'business_metrics',
    data: {
      verifications: [
        { status: 'passed', score: 0.95 },
        { status: 'failed', score: 0.30 }
      ]
    },
    rules: [
      {
        field: 'verification_count',
        operation: 'count',
        description: 'Count total verifications',
        enabled: true,
        order: 1
      }
    ]
  })
});

const result = await response.json();
console.log('Aggregation ID:', result.aggregation_id);
console.log('Aggregated Data:', result.aggregated_data);

// Background job
const jobResponse = await fetch('https://api.kyb-platform.com/v1/aggregate/job', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    aggregation_type: 'business_metrics',
    data: { /* large dataset */ },
    schema_id: 'business_metrics_default'
  })
});

const job = await jobResponse.json();
console.log('Job ID:', job.job_id);

// Poll for job completion
const checkJob = async (jobId) => {
  const statusResponse = await fetch(`https://api.kyb-platform.com/v1/aggregate/job/${jobId}`, {
    headers: {
      'Authorization': 'Bearer YOUR_API_KEY'
    }
  });
  
  const jobStatus = await statusResponse.json();
  
  if (jobStatus.status === 'completed') {
    console.log('Job completed:', jobStatus.result);
    return jobStatus.result;
  } else if (jobStatus.status === 'failed') {
    throw new Error(`Job failed: ${jobStatus.error}`);
  } else {
    // Wait and retry
    await new Promise(resolve => setTimeout(resolve, 2000));
    return checkJob(jobId);
  }
};

const finalResult = await checkJob(job.job_id);
```

### Python

```python
import requests
import time

# Immediate aggregation
response = requests.post(
    'https://api.kyb-platform.com/v1/aggregate',
    headers={
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json',
        'X-Business-ID': 'business_123'
    },
    json={
        'aggregation_type': 'business_metrics',
        'data': {
            'verifications': [
                {'status': 'passed', 'score': 0.95},
                {'status': 'failed', 'score': 0.30}
            ]
        },
        'rules': [
            {
                'field': 'verification_count',
                'operation': 'count',
                'description': 'Count total verifications',
                'enabled': True,
                'order': 1
            }
        ]
    }
)

result = response.json()
print(f"Aggregation ID: {result['aggregation_id']}")
print(f"Aggregated Data: {result['aggregated_data']}")

# Background job
job_response = requests.post(
    'https://api.kyb-platform.com/v1/aggregate/job',
    headers={
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
    },
    json={
        'aggregation_type': 'business_metrics',
        'data': {},  # large dataset
        'schema_id': 'business_metrics_default'
    }
)

job = job_response.json()
print(f"Job ID: {job['job_id']}")

# Poll for job completion
def check_job(job_id):
    while True:
        status_response = requests.get(
            f'https://api.kyb-platform.com/v1/aggregate/job/{job_id}',
            headers={'Authorization': 'Bearer YOUR_API_KEY'}
        )
        
        job_status = status_response.json()
        
        if job_status['status'] == 'completed':
            print(f"Job completed: {job_status['result']}")
            return job_status['result']
        elif job_status['status'] == 'failed':
            raise Exception(f"Job failed: {job_status['error']}")
        else:
            time.sleep(2)

final_result = check_job(job['job_id'])
```

### React

```jsx
import React, { useState, useEffect } from 'react';

const DataAggregation = () => {
  const [aggregationResult, setAggregationResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const performAggregation = async (data) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('https://api.kyb-platform.com/v1/aggregate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${process.env.REACT_APP_API_KEY}`,
          'Content-Type': 'application/json',
          'X-Business-ID': 'business_123'
        },
        body: JSON.stringify({
          aggregation_type: 'business_metrics',
          data: data,
          rules: [
            {
              field: 'verification_count',
              operation: 'count',
              description: 'Count total verifications',
              enabled: true,
              order: 1
            }
          ]
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      setAggregationResult(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Data Aggregation</h2>
      
      <button 
        onClick={() => performAggregation({
          verifications: [
            { status: 'passed', score: 0.95 },
            { status: 'failed', score: 0.30 }
          ]
        })}
        disabled={loading}
      >
        {loading ? 'Aggregating...' : 'Perform Aggregation'}
      </button>

      {error && (
        <div style={{ color: 'red' }}>
          Error: {error}
        </div>
      )}

      {aggregationResult && (
        <div>
          <h3>Aggregation Results</h3>
          <p>ID: {aggregationResult.aggregation_id}</p>
          <p>Status: {aggregationResult.status}</p>
          <p>Processing Time: {aggregationResult.processing_time}</p>
          <pre>{JSON.stringify(aggregationResult.aggregated_data, null, 2)}</pre>
        </div>
      )}
    </div>
  );
};

export default DataAggregation;
```

## Best Practices

### Performance

1. **Use Background Jobs**: For large datasets (>1000 records), use background jobs instead of immediate aggregation
2. **Optimize Rules**: Keep aggregation rules simple and efficient
3. **Use Schemas**: Leverage pre-configured schemas for common aggregation patterns
4. **Filter Data**: Use filters to reduce the dataset size before aggregation
5. **Cache Results**: Cache aggregation results for frequently requested data

### Error Handling

1. **Validate Input**: Always validate aggregation rules and data before processing
2. **Handle Timeouts**: Implement proper timeout handling for long-running aggregations
3. **Retry Logic**: Implement retry logic for failed jobs
4. **Monitor Progress**: Track job progress for better user experience
5. **Graceful Degradation**: Handle partial failures gracefully

### Data Quality

1. **Data Validation**: Validate input data before aggregation
2. **Handle Missing Data**: Implement proper handling for missing or null values
3. **Data Types**: Ensure proper data types for aggregation operations
4. **Outlier Detection**: Consider outlier detection for statistical aggregations
5. **Data Sampling**: Use sampling for very large datasets

### Security

1. **Input Sanitization**: Sanitize all input data to prevent injection attacks
2. **Access Control**: Implement proper access control for aggregation schemas
3. **Rate Limiting**: Respect rate limits to prevent abuse
4. **Audit Logging**: Log all aggregation requests for audit purposes
5. **Data Encryption**: Ensure sensitive data is encrypted in transit and at rest

## Monitoring and Alerting

### Key Metrics

- **Aggregation Success Rate**: Monitor the percentage of successful aggregations
- **Processing Time**: Track average and 95th percentile processing times
- **Job Queue Length**: Monitor the number of pending aggregation jobs
- **Error Rate**: Track aggregation error rates by type
- **Resource Usage**: Monitor CPU and memory usage during aggregation

### Alerts

- **High Error Rate**: Alert when aggregation error rate exceeds 5%
- **Long Processing Times**: Alert when average processing time exceeds 30 seconds
- **Job Queue Backlog**: Alert when job queue length exceeds 100
- **Resource Exhaustion**: Alert when CPU or memory usage exceeds 80%
- **Schema Errors**: Alert when schema validation errors occur

### Logging

```json
{
  "level": "info",
  "message": "Data aggregation completed",
  "aggregation_id": "agg_1234567890_1",
  "business_id": "business_123",
  "type": "business_metrics",
  "success": true,
  "processing_time": "150ms",
  "rules_applied": 2,
  "data_count": 100,
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Troubleshooting

### Common Issues

1. **Invalid Aggregation Type**
   - **Cause**: Using an unsupported aggregation type
   - **Solution**: Check the list of supported aggregation types

2. **Missing Required Fields**
   - **Cause**: Missing required fields in aggregation rules
   - **Solution**: Ensure all required fields are provided

3. **Job Timeout**
   - **Cause**: Large datasets taking too long to process
   - **Solution**: Use background jobs for large datasets

4. **Memory Issues**
   - **Cause**: Processing very large datasets in memory
   - **Solution**: Implement data streaming or chunking

5. **Schema Not Found**
   - **Cause**: Referencing a non-existent schema
   - **Solution**: Check available schemas or create custom rules

### Debugging

1. **Enable Debug Logging**: Set log level to debug for detailed information
2. **Check Job Status**: Monitor job status and progress
3. **Validate Rules**: Test aggregation rules with sample data
4. **Profile Performance**: Use performance profiling tools
5. **Monitor Resources**: Track system resource usage

### Support

For additional support:

- **Documentation**: Check the API documentation for detailed information
- **Logs**: Review application logs for error details
- **Metrics**: Monitor system metrics for performance issues
- **Support Team**: Contact the support team with specific error details

## Rate Limits

- **Immediate Aggregation**: 100 requests per minute per API key
- **Background Jobs**: 50 job creations per minute per API key
- **Job Status Checks**: 200 requests per minute per API key
- **Schema Retrieval**: 300 requests per minute per API key

When rate limits are exceeded, the API returns a 429 status code with details about when the limit resets.
