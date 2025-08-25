# Data Transformation Endpoints

## Overview

The Data Transformation API provides comprehensive data transformation capabilities for the KYB platform. It allows users to transform business data using various transformation rules and operations, including data cleaning, normalization, enrichment, aggregation, filtering, mapping, and custom transformations.

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
  "transformation_id": "transform_1234567890",
  "business_id": "business_123",
  "transformation_type": "data_cleaning",
  "status": "success",
  "is_successful": true,
  "transformed_data": { ... },
  "applied_rules": [ ... ],
  "summary": { ... },
  "transformed_at": "2024-12-19T10:30:00Z",
  "processing_time": "150ms"
}
```

## Supported Transformation Types

| Type | Description |
|------|-------------|
| `data_cleaning` | Clean and standardize data (trim whitespace, normalize formats) |
| `normalization` | Normalize data formats (phone numbers, addresses, etc.) |
| `enrichment` | Enrich data with additional information |
| `aggregation` | Aggregate and summarize data |
| `filtering` | Filter data based on criteria |
| `mapping` | Map data between different formats |
| `custom` | Custom transformation operations |
| `all` | Apply all transformation types |

## Supported Transformation Operations

| Operation | Description |
|-----------|-------------|
| `trim` | Remove leading and trailing whitespace |
| `to_lower` | Convert text to lowercase |
| `to_upper` | Convert text to uppercase |
| `replace` | Replace text patterns |
| `extract` | Extract data using patterns |
| `format` | Format data (phone, date, etc.) |
| `validate` | Validate data against rules |
| `enrich` | Enrich data with external sources |
| `aggregate` | Aggregate data values |
| `filter` | Filter data based on conditions |
| `map` | Map data between formats |
| `custom` | Custom transformation logic |

## Endpoints

### 1. Transform Data

**POST** `/v1/transform`

Performs immediate data transformation using specified rules or schema.

#### Request Body

```json
{
  "business_id": "business_123",
  "transformation_type": "data_cleaning",
  "data": {
    "business_name": "  Test Company  ",
    "email": "TEST@EXAMPLE.COM",
    "phone": "+1-555-123-4567"
  },
  "rules": [
    {
      "field": "business_name",
      "operation": "trim",
      "parameters": {},
      "description": "Trim whitespace from business name",
      "enabled": true,
      "order": 1
    },
    {
      "field": "email",
      "operation": "to_lower",
      "parameters": {},
      "description": "Convert email to lowercase",
      "enabled": true,
      "order": 2
    }
  ],
  "schema_id": "data_cleaning_default",
  "validate_before": false,
  "validate_after": true,
  "include_metadata": true,
  "metadata": {
    "source": "manual_entry",
    "priority": "high"
  }
}
```

#### Response

```json
{
  "transformation_id": "transform_1234567890",
  "business_id": "business_123",
  "transformation_type": "data_cleaning",
  "status": "success",
  "is_successful": true,
  "original_data": {
    "business_name": "  Test Company  ",
    "email": "TEST@EXAMPLE.COM",
    "phone": "+1-555-123-4567"
  },
  "transformed_data": {
    "business_name": "Test Company",
    "email": "test@example.com",
    "phone": "+1-555-123-4567"
  },
  "applied_rules": [
    {
      "field": "business_name",
      "operation": "trim",
      "description": "Trim whitespace from business name",
      "enabled": true,
      "order": 1
    },
    {
      "field": "email",
      "operation": "to_lower",
      "description": "Convert email to lowercase",
      "enabled": true,
      "order": 2
    }
  ],
  "skipped_rules": [],
  "failed_rules": [],
  "validation_before": null,
  "validation_after": {
    "valid": true,
    "issues": []
  },
  "summary": {
    "total_rules": 2,
    "applied_rules": 2,
    "skipped_rules": 0,
    "failed_rules": 0,
    "success_rate": 1.0,
    "processing_time": "45ms"
  },
  "transformed_at": "2024-12-19T10:30:00Z",
  "processing_time": "45ms",
  "metadata": {
    "source": "manual_entry",
    "priority": "high"
  }
}
```

#### Headers

- `X-Transformation-ID`: Unique transformation identifier
- `X-Processing-Time`: Processing time in milliseconds

### 2. Create Transformation Job

**POST** `/v1/transform/job`

Creates a background transformation job for processing large datasets.

#### Request Body

```json
{
  "business_id": "business_123",
  "transformation_type": "normalization",
  "data": {
    "phone": "+1-555-123-4567"
  },
  "rules": [
    {
      "field": "phone",
      "operation": "format",
      "parameters": {
        "format": "E.164"
      },
      "description": "Format phone number to E.164",
      "enabled": true,
      "order": 1
    }
  ],
  "validate_before": true,
  "validate_after": true,
  "metadata": {
    "batch_id": "batch_123",
    "priority": "normal"
  }
}
```

#### Response

```json
{
  "job_id": "transform_1702998000_1",
  "status": "pending",
  "created_at": "2024-12-19T10:30:00Z",
  "message": "Transformation job created successfully"
}
```

#### Headers

- `X-Job-ID`: Unique job identifier

### 3. Get Transformation Job

**GET** `/v1/transform/job/{job_id}`

Retrieves the status and results of a transformation job.

#### Response

```json
{
  "id": "transform_1702998000_1",
  "business_id": "business_123",
  "transformation_type": "normalization",
  "status": "completed",
  "progress": 100,
  "created_at": "2024-12-19T10:30:00Z",
  "started_at": "2024-12-19T10:30:02Z",
  "completed_at": "2024-12-19T10:30:05Z",
  "result": {
    "transformation_id": "transform_1702998000_1",
    "business_id": "business_123",
    "transformation_type": "normalization",
    "status": "success",
    "is_successful": true,
    "transformed_data": {
      "phone": "+15551234567"
    },
    "applied_rules": [
      {
        "field": "phone",
        "operation": "format",
        "description": "Format phone number to E.164",
        "enabled": true,
        "order": 1
      }
    ],
    "summary": {
      "total_rules": 1,
      "applied_rules": 1,
      "success_rate": 1.0
    },
    "transformed_at": "2024-12-19T10:30:05Z",
    "processing_time": "3s"
  },
  "metadata": {
    "batch_id": "batch_123",
    "priority": "normal"
  }
}
```

### 4. List Transformation Jobs

**GET** `/v1/transform/jobs`

Lists transformation jobs with optional filtering and pagination.

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `page` | integer | Page number for pagination | 1 |
| `limit` | integer | Number of jobs per page (max 100) | 20 |
| `business_id` | string | Filter by business ID | - |
| `status` | string | Filter by job status | - |
| `type` | string | Filter by transformation type | - |

#### Response

```json
{
  "jobs": [
    {
      "id": "transform_1702998000_1",
      "business_id": "business_123",
      "transformation_type": "normalization",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-12-19T10:30:00Z",
      "completed_at": "2024-12-19T10:30:05Z"
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

### 5. Get Transformation Schema

**GET** `/v1/transform/schema/{schema_id}`

Retrieves a specific transformation schema.

#### Response

```json
{
  "id": "data_cleaning_default",
  "name": "Default Data Cleaning",
  "description": "Standard data cleaning transformations",
  "type": "data_cleaning",
  "rules": [
    {
      "field": "business_name",
      "operation": "trim",
      "parameters": {},
      "description": "Trim whitespace from business name",
      "enabled": true,
      "order": 1
    },
    {
      "field": "email",
      "operation": "to_lower",
      "parameters": {},
      "description": "Convert email to lowercase",
      "enabled": true,
      "order": 2
    }
  ],
  "version": "1.0.0",
  "created_at": "2024-12-19T10:00:00Z",
  "updated_at": "2024-12-19T10:00:00Z"
}
```

### 6. List Transformation Schemas

**GET** `/v1/transform/schemas`

Lists available transformation schemas.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | Filter by transformation type |

#### Response

```json
{
  "schemas": [
    {
      "id": "data_cleaning_default",
      "name": "Default Data Cleaning",
      "description": "Standard data cleaning transformations",
      "type": "data_cleaning",
      "version": "1.0.0"
    },
    {
      "id": "normalization_default",
      "name": "Default Normalization",
      "description": "Standard data normalization transformations",
      "type": "normalization",
      "version": "1.0.0"
    }
  ],
  "total": 2
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid request: transformation_type is required",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "transformation_type",
    "message": "transformation_type is required"
  }
}
```

### 401 Unauthorized

```json
{
  "error": "Missing or invalid authorization header",
  "code": "UNAUTHORIZED"
}
```

### 404 Not Found

```json
{
  "error": "Transformation job not found",
  "code": "NOT_FOUND"
}
```

### 500 Internal Server Error

```json
{
  "error": "Transformation failed: internal error",
  "code": "INTERNAL_ERROR"
}
```

## Integration Examples

### JavaScript/TypeScript

```javascript
class DataTransformationClient {
  constructor(apiKey, baseUrl = 'https://api.kyb-platform.com/v1') {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async transformData(request) {
    const response = await fetch(`${this.baseUrl}/transform`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Transformation failed: ${response.statusText}`);
    }

    return response.json();
  }

  async createTransformationJob(request) {
    const response = await fetch(`${this.baseUrl}/transform/job`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Job creation failed: ${response.statusText}`);
    }

    return response.json();
  }

  async getTransformationJob(jobId) {
    const response = await fetch(`${this.baseUrl}/transform/job/${jobId}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Job retrieval failed: ${response.statusText}`);
    }

    return response.json();
  }

  async listTransformationJobs(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const response = await fetch(`${this.baseUrl}/transform/jobs?${queryString}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Job listing failed: ${response.statusText}`);
    }

    return response.json();
  }
}

// Usage example
const client = new DataTransformationClient('your-api-key');

// Immediate transformation
const result = await client.transformData({
  business_id: 'business_123',
  transformation_type: 'data_cleaning',
  data: {
    business_name: '  Test Company  ',
    email: 'TEST@EXAMPLE.COM',
  },
  rules: [
    {
      field: 'business_name',
      operation: 'trim',
      description: 'Trim whitespace',
      enabled: true,
      order: 1,
    },
    {
      field: 'email',
      operation: 'to_lower',
      description: 'Convert to lowercase',
      enabled: true,
      order: 2,
    },
  ],
});

console.log('Transformation result:', result);

// Background job
const job = await client.createTransformationJob({
  business_id: 'business_123',
  transformation_type: 'normalization',
  data: { phone: '+1-555-123-4567' },
  rules: [
    {
      field: 'phone',
      operation: 'format',
      parameters: { format: 'E.164' },
      description: 'Format phone number',
      enabled: true,
      order: 1,
    },
  ],
});

console.log('Job created:', job.job_id);

// Poll for job completion
let jobStatus = await client.getTransformationJob(job.job_id);
while (jobStatus.status === 'pending' || jobStatus.status === 'processing') {
  await new Promise(resolve => setTimeout(resolve, 1000));
  jobStatus = await client.getTransformationJob(job.job_id);
}

console.log('Job completed:', jobStatus);
```

### Python

```python
import requests
import time
from typing import Dict, Any, Optional

class DataTransformationClient:
    def __init__(self, api_key: str, base_url: str = 'https://api.kyb-platform.com/v1'):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json',
        }

    def transform_data(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Perform immediate data transformation."""
        response = requests.post(
            f'{self.base_url}/transform',
            headers=self.headers,
            json=request
        )
        response.raise_for_status()
        return response.json()

    def create_transformation_job(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Create a background transformation job."""
        response = requests.post(
            f'{self.base_url}/transform/job',
            headers=self.headers,
            json=request
        )
        response.raise_for_status()
        return response.json()

    def get_transformation_job(self, job_id: str) -> Dict[str, Any]:
        """Get transformation job status and results."""
        response = requests.get(
            f'{self.base_url}/transform/job/{job_id}',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_transformation_jobs(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """List transformation jobs with optional filtering."""
        response = requests.get(
            f'{self.base_url}/transform/jobs',
            headers=self.headers,
            params=params or {}
        )
        response.raise_for_status()
        return response.json()

    def get_transformation_schema(self, schema_id: str) -> Dict[str, Any]:
        """Get a specific transformation schema."""
        response = requests.get(
            f'{self.base_url}/transform/schema/{schema_id}',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_transformation_schemas(self, transformation_type: Optional[str] = None) -> Dict[str, Any]:
        """List available transformation schemas."""
        params = {'type': transformation_type} if transformation_type else {}
        response = requests.get(
            f'{self.base_url}/transform/schemas',
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        return response.json()

# Usage example
client = DataTransformationClient('your-api-key')

# Immediate transformation
result = client.transform_data({
    'business_id': 'business_123',
    'transformation_type': 'data_cleaning',
    'data': {
        'business_name': '  Test Company  ',
        'email': 'TEST@EXAMPLE.COM',
    },
    'rules': [
        {
            'field': 'business_name',
            'operation': 'trim',
            'description': 'Trim whitespace',
            'enabled': True,
            'order': 1,
        },
        {
            'field': 'email',
            'operation': 'to_lower',
            'description': 'Convert to lowercase',
            'enabled': True,
            'order': 2,
        },
    ],
})

print('Transformation result:', result)

# Background job
job = client.create_transformation_job({
    'business_id': 'business_123',
    'transformation_type': 'normalization',
    'data': {'phone': '+1-555-123-4567'},
    'rules': [
        {
            'field': 'phone',
            'operation': 'format',
            'parameters': {'format': 'E.164'},
            'description': 'Format phone number',
            'enabled': True,
            'order': 1,
        },
    ],
})

print('Job created:', job['job_id'])

# Poll for job completion
job_status = client.get_transformation_job(job['job_id'])
while job_status['status'] in ['pending', 'processing']:
    time.sleep(1)
    job_status = client.get_transformation_job(job['job_id'])

print('Job completed:', job_status)
```

### React

```jsx
import React, { useState, useEffect } from 'react';

const DataTransformationComponent = ({ apiKey, businessId }) => {
  const [transformationResult, setTransformationResult] = useState(null);
  const [jobStatus, setJobStatus] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const transformData = async (data, rules) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('https://api.kyb-platform.com/v1/transform', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          business_id: businessId,
          transformation_type: 'data_cleaning',
          data,
          rules,
          validate_after: true,
        }),
      });

      if (!response.ok) {
        throw new Error(`Transformation failed: ${response.statusText}`);
      }

      const result = await response.json();
      setTransformationResult(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const createTransformationJob = async (data, rules) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('https://api.kyb-platform.com/v1/transform/job', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          business_id: businessId,
          transformation_type: 'normalization',
          data,
          rules,
        }),
      });

      if (!response.ok) {
        throw new Error(`Job creation failed: ${response.statusText}`);
      }

      const job = await response.json();
      
      // Poll for job completion
      const pollJob = async () => {
        const jobResponse = await fetch(`https://api.kyb-platform.com/v1/transform/job/${job.job_id}`, {
          headers: {
            'Authorization': `Bearer ${apiKey}`,
          },
        });

        if (!jobResponse.ok) {
          throw new Error(`Job retrieval failed: ${jobResponse.statusText}`);
        }

        const jobStatus = await jobResponse.json();
        setJobStatus(jobStatus);

        if (jobStatus.status === 'pending' || jobStatus.status === 'processing') {
          setTimeout(pollJob, 1000);
        }
      };

      pollJob();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleImmediateTransform = () => {
    const data = {
      business_name: '  Test Company  ',
      email: 'TEST@EXAMPLE.COM',
    };

    const rules = [
      {
        field: 'business_name',
        operation: 'trim',
        description: 'Trim whitespace',
        enabled: true,
        order: 1,
      },
      {
        field: 'email',
        operation: 'to_lower',
        description: 'Convert to lowercase',
        enabled: true,
        order: 2,
      },
    ];

    transformData(data, rules);
  };

  const handleBackgroundJob = () => {
    const data = { phone: '+1-555-123-4567' };
    const rules = [
      {
        field: 'phone',
        operation: 'format',
        parameters: { format: 'E.164' },
        description: 'Format phone number',
        enabled: true,
        order: 1,
      },
    ];

    createTransformationJob(data, rules);
  };

  return (
    <div className="data-transformation">
      <h2>Data Transformation</h2>
      
      <div className="actions">
        <button onClick={handleImmediateTransform} disabled={loading}>
          Immediate Transformation
        </button>
        <button onClick={handleBackgroundJob} disabled={loading}>
          Background Job
        </button>
      </div>

      {loading && <div className="loading">Processing...</div>}
      
      {error && <div className="error">Error: {error}</div>}

      {transformationResult && (
        <div className="result">
          <h3>Transformation Result</h3>
          <pre>{JSON.stringify(transformationResult, null, 2)}</pre>
        </div>
      )}

      {jobStatus && (
        <div className="job-status">
          <h3>Job Status</h3>
          <p>Status: {jobStatus.status}</p>
          <p>Progress: {jobStatus.progress}%</p>
          {jobStatus.result && (
            <div>
              <h4>Result</h4>
              <pre>{JSON.stringify(jobStatus.result, null, 2)}</pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default DataTransformationComponent;
```

## Best Practices

### Performance

1. **Use Immediate Transformation for Small Datasets**: For datasets with fewer than 100 records, use immediate transformation for faster results.

2. **Use Background Jobs for Large Datasets**: For large datasets, use background jobs to avoid timeout issues and get progress updates.

3. **Optimize Transformation Rules**: Order rules efficiently and disable unnecessary rules to improve performance.

4. **Batch Processing**: When processing multiple datasets, consider batching them into a single transformation job.

### Error Handling

1. **Validate Input Data**: Always validate input data before transformation to catch issues early.

2. **Handle Partial Failures**: Check the `status` field in responses to handle partial transformation failures.

3. **Monitor Job Progress**: For background jobs, implement proper polling with exponential backoff.

4. **Log Transformation Results**: Log transformation results for debugging and audit purposes.

### Data Quality

1. **Pre-Transformation Validation**: Use `validate_before: true` to validate data before transformation.

2. **Post-Transformation Validation**: Use `validate_after: true` to validate transformed data.

3. **Review Failed Rules**: Always review `failed_rules` to understand transformation issues.

4. **Use Appropriate Schemas**: Leverage pre-configured schemas for common transformation patterns.

### Security

1. **Validate Business ID**: Always validate the business ID to ensure data isolation.

2. **Sanitize Input Data**: Sanitize input data to prevent injection attacks.

3. **Use HTTPS**: Always use HTTPS for API communication.

4. **Monitor API Usage**: Monitor API usage for unusual patterns or abuse.

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Transformation Success Rate**: Monitor the success rate of transformations.
2. **Processing Time**: Track average processing time for transformations.
3. **Job Completion Rate**: Monitor background job completion rates.
4. **Error Rates**: Track transformation error rates by type.

### Recommended Alerts

1. **High Error Rate**: Alert when transformation error rate exceeds 5%.
2. **Long Processing Time**: Alert when average processing time exceeds 30 seconds.
3. **Job Failures**: Alert when background job failure rate exceeds 10%.
4. **API Rate Limits**: Alert when approaching API rate limits.

### Logging

```javascript
// Example logging implementation
const logger = {
  info: (message, metadata) => {
    console.log(`[INFO] ${message}`, metadata);
  },
  error: (message, error) => {
    console.error(`[ERROR] ${message}`, error);
  },
  warn: (message, metadata) => {
    console.warn(`[WARN] ${message}`, metadata);
  },
};

// Log transformation events
logger.info('Transformation started', {
  business_id: 'business_123',
  transformation_type: 'data_cleaning',
  rule_count: 5,
});

logger.info('Transformation completed', {
  transformation_id: 'transform_123',
  success_rate: 0.95,
  processing_time: '150ms',
});
```

## Troubleshooting

### Common Issues

1. **Invalid Transformation Type**
   - **Error**: `invalid transformation_type: invalid_type`
   - **Solution**: Use one of the supported transformation types listed above.

2. **Missing Required Fields**
   - **Error**: `transformation_type is required`
   - **Solution**: Ensure all required fields are provided in the request.

3. **Job Not Found**
   - **Error**: `Transformation job not found`
   - **Solution**: Verify the job ID is correct and the job hasn't expired.

4. **Schema Not Found**
   - **Error**: `schema not found: invalid_schema`
   - **Solution**: Use a valid schema ID or provide custom rules.

### Debugging Tips

1. **Check Request Format**: Ensure the request body follows the correct JSON format.

2. **Validate Rules**: Verify that transformation rules have valid operations and parameters.

3. **Monitor Job Status**: For background jobs, check the job status regularly.

4. **Review Error Details**: Check the `failed_rules` array for specific transformation failures.

5. **Test with Small Data**: Test transformations with small datasets before processing large volumes.

### Support

For additional support:

1. **Documentation**: Refer to this documentation for detailed API specifications.
2. **Error Codes**: Check error codes and messages for specific issues.
3. **Logs**: Review application logs for detailed error information.
4. **Contact Support**: Contact the support team with specific error details and request IDs.

## Rate Limits

The Data Transformation API has the following rate limits:

- **Immediate Transformations**: 100 requests per minute per API key
- **Background Jobs**: 50 job creations per minute per API key
- **Job Status Queries**: 200 requests per minute per API key
- **Schema Queries**: 300 requests per minute per API key

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1702998060
```

When rate limits are exceeded, the API returns a 429 status code with details about when the limit resets.
