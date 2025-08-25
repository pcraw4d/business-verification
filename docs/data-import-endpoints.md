# Data Import Endpoints Documentation

## Overview

The Data Import API provides comprehensive functionality for importing business verification data, classification results, risk assessments, and other platform data into the KYB platform. The API supports multiple import formats, validation rules, transformation rules, and both immediate and background processing modes.

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
  "import_id": "import_business_123_1703123456",
  "business_id": "business_123",
  "import_type": "business_verifications",
  "format": "json",
  "mode": "upsert",
  "status": "completed",
  "record_count": 150,
  "success_count": 145,
  "error_count": 5,
  "skipped_count": 0,
  "errors": [],
  "warnings": [],
  "summary": {
    "total_records": 150,
    "successful": 145,
    "failed": 5,
    "skipped": 0,
    "success_rate": 96.67,
    "error_rate": 3.33,
    "processing_time": "2.5s"
  },
  "processed_at": "2024-12-19T10:30:56Z",
  "metadata": {}
}
```

## Supported Import Types

| Type | Description |
|------|-------------|
| `business_verifications` | Business verification data |
| `classifications` | Business classification results |
| `risk_assessments` | Risk assessment data |
| `compliance_reports` | Compliance report data |
| `audit_trails` | Audit trail data |
| `metrics` | Platform metrics data |
| `all` | All data types combined |

## Supported Import Formats

| Format | Description |
|--------|-------------|
| `json` | JSON format (default) |
| `csv` | Comma-separated values |
| `xml` | XML format |
| `xlsx` | Excel spreadsheet format |

## Import Modes

| Mode | Description |
|------|-------------|
| `create` | Create new records only |
| `update` | Update existing records only |
| `upsert` | Create or update records (default) |
| `replace` | Replace all records |

## Endpoints

### 1. Import Data (Immediate)

**POST** `/v1/import`

Immediately processes and imports data, returning results synchronously.

#### Request Body

```json
{
  "business_id": "business_123",
  "import_type": "business_verifications",
  "format": "json",
  "mode": "upsert",
  "data": {
    "verifications": [
      {
        "business_name": "Acme Corporation",
        "address": "123 Main St, Anytown, ST 12345",
        "phone": "+1-555-123-4567",
        "email": "contact@acme.com",
        "website": "https://www.acme.com",
        "industry": "Technology",
        "registration_number": "123456789",
        "tax_id": "12-3456789"
      }
    ]
  },
  "validation_rules": {
    "business_name": {
      "type": "required",
      "message": "Business name is required"
    },
    "email": {
      "type": "format",
      "value": "email",
      "message": "Invalid email format"
    }
  },
  "transform_rules": {
    "business_name": {
      "operation": "trim",
      "description": "Remove leading/trailing whitespace"
    },
    "phone": {
      "operation": "format",
      "value": "E.164",
      "description": "Format phone number to E.164"
    }
  },
  "conflict_policy": "update",
  "dry_run": false,
  "metadata": {
    "source": "manual_import",
    "user_id": "user_456"
  }
}
```

#### Response

```json
{
  "import_id": "import_business_123_1703123456",
  "business_id": "business_123",
  "import_type": "business_verifications",
  "format": "json",
  "mode": "upsert",
  "status": "completed",
  "record_count": 1,
  "success_count": 1,
  "error_count": 0,
  "skipped_count": 0,
  "errors": [],
  "warnings": [],
  "summary": {
    "total_records": 1,
    "successful": 1,
    "failed": 0,
    "skipped": 0,
    "success_rate": 100.0,
    "error_rate": 0.0,
    "processing_time": "150ms"
  },
  "processed_at": "2024-12-19T10:30:56Z",
  "metadata": {
    "source": "manual_import",
    "user_id": "user_456"
  }
}
```

### 2. Create Import Job (Background)

**POST** `/v1/import/job`

Creates a background import job for processing large datasets asynchronously.

#### Request Body

Same as immediate import endpoint.

#### Response

```json
{
  "id": "import_job_1703123456_1",
  "business_id": "business_123",
  "import_type": "business_verifications",
  "format": "json",
  "mode": "upsert",
  "status": "pending",
  "progress": 0,
  "record_count": 0,
  "success_count": 0,
  "error_count": 0,
  "skipped_count": 0,
  "errors": [],
  "warnings": [],
  "created_at": "2024-12-19T10:30:56Z",
  "started_at": null,
  "completed_at": null,
  "metadata": {
    "source": "bulk_import",
    "user_id": "user_456"
  }
}
```

### 3. Get Import Job Status

**GET** `/v1/import/job/{job_id}`

Retrieves the current status and results of a background import job.

#### Response

```json
{
  "id": "import_job_1703123456_1",
  "business_id": "business_123",
  "import_type": "business_verifications",
  "format": "json",
  "mode": "upsert",
  "status": "completed",
  "progress": 100,
  "record_count": 1000,
  "success_count": 985,
  "error_count": 15,
  "skipped_count": 0,
  "errors": [
    {
      "row": 45,
      "field": "email",
      "message": "Invalid email format",
      "severity": "error",
      "data": "invalid-email"
    }
  ],
  "warnings": [
    {
      "row": 67,
      "field": "phone",
      "message": "Phone number format may be incorrect",
      "severity": "warning",
      "data": "555-123-4567"
    }
  ],
  "created_at": "2024-12-19T10:30:56Z",
  "started_at": "2024-12-19T10:30:57Z",
  "completed_at": "2024-12-19T10:31:15Z",
  "metadata": {
    "source": "bulk_import",
    "user_id": "user_456"
  }
}
```

### 4. List Import Jobs

**GET** `/v1/import/jobs`

Lists all import jobs with optional filtering and pagination.

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `business_id` | string | Filter by business ID | - |
| `status` | string | Filter by job status | - |
| `limit` | integer | Number of jobs to return (1-100) | 50 |
| `offset` | integer | Number of jobs to skip | 0 |

#### Response

```json
{
  "jobs": [
    {
      "id": "import_job_1703123456_1",
      "business_id": "business_123",
      "import_type": "business_verifications",
      "format": "json",
      "mode": "upsert",
      "status": "completed",
      "progress": 100,
      "record_count": 1000,
      "success_count": 985,
      "error_count": 15,
      "created_at": "2024-12-19T10:30:56Z",
      "completed_at": "2024-12-19T10:31:15Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

## Error Responses

### Validation Error

```json
{
  "error": "Validation failed: import_type is required",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "import_type",
    "message": "import_type is required"
  }
}
```

### Processing Error

```json
{
  "error": "Import processing failed",
  "code": "PROCESSING_ERROR",
  "details": {
    "import_id": "import_business_123_1703123456",
    "reason": "Database connection failed"
  }
}
```

### Job Not Found

```json
{
  "error": "Import job not found",
  "code": "JOB_NOT_FOUND",
  "details": {
    "job_id": "nonexistent_job"
  }
}
```

## Integration Examples

### JavaScript/TypeScript

```typescript
class KYBImportClient {
  private apiKey: string;
  private baseUrl: string;

  constructor(apiKey: string, baseUrl: string = 'https://api.kyb-platform.com/v1') {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async importData(request: ImportRequest): Promise<ImportResponse> {
    const response = await fetch(`${this.baseUrl}/import`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Import failed: ${response.statusText}`);
    }

    return response.json();
  }

  async createImportJob(request: ImportRequest): Promise<ImportJob> {
    const response = await fetch(`${this.baseUrl}/import/job`, {
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

  async getImportJobStatus(jobId: string): Promise<ImportJob> {
    const response = await fetch(`${this.baseUrl}/import/job/${jobId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Job status retrieval failed: ${response.statusText}`);
    }

    return response.json();
  }

  async listImportJobs(filters: ImportJobFilters = {}): Promise<ListImportJobsResponse> {
    const params = new URLSearchParams();
    if (filters.businessId) params.append('business_id', filters.businessId);
    if (filters.status) params.append('status', filters.status);
    if (filters.limit) params.append('limit', filters.limit.toString());
    if (filters.offset) params.append('offset', filters.offset.toString());

    const response = await fetch(`${this.baseUrl}/import/jobs?${params}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Job listing failed: ${response.statusText}`);
    }

    return response.json();
  }

  async waitForJobCompletion(jobId: string, pollInterval: number = 5000): Promise<ImportJob> {
    while (true) {
      const job = await this.getImportJobStatus(jobId);
      
      if (job.status === 'completed' || job.status === 'failed') {
        return job;
      }
      
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    }
  }
}

// Usage example
const client = new KYBImportClient('your-api-key');

// Immediate import
const importRequest: ImportRequest = {
  business_id: 'business_123',
  import_type: 'business_verifications',
  format: 'json',
  mode: 'upsert',
  data: {
    verifications: [
      {
        business_name: 'Acme Corporation',
        address: '123 Main St, Anytown, ST 12345',
        phone: '+1-555-123-4567',
        email: 'contact@acme.com'
      }
    ]
  }
};

try {
  const result = await client.importData(importRequest);
  console.log(`Import completed: ${result.success_count}/${result.record_count} records processed`);
} catch (error) {
  console.error('Import failed:', error);
}

// Background job
try {
  const job = await client.createImportJob(importRequest);
  console.log(`Job created: ${job.id}`);
  
  const completedJob = await client.waitForJobCompletion(job.id);
  console.log(`Job completed: ${completedJob.success_count}/${completedJob.record_count} records processed`);
} catch (error) {
  console.error('Job failed:', error);
}
```

### Python

```python
import requests
import time
from typing import Dict, List, Optional, Any

class KYBImportClient:
    def __init__(self, api_key: str, base_url: str = 'https://api.kyb-platform.com/v1'):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def import_data(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Perform immediate data import"""
        response = requests.post(
            f'{self.base_url}/import',
            headers=self.headers,
            json=request
        )
        response.raise_for_status()
        return response.json()

    def create_import_job(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Create a background import job"""
        response = requests.post(
            f'{self.base_url}/import/job',
            headers=self.headers,
            json=request
        )
        response.raise_for_status()
        return response.json()

    def get_import_job_status(self, job_id: str) -> Dict[str, Any]:
        """Get the status of an import job"""
        response = requests.get(
            f'{self.base_url}/import/job/{job_id}',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_import_jobs(self, filters: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """List import jobs with optional filtering"""
        params = {}
        if filters:
            if 'business_id' in filters:
                params['business_id'] = filters['business_id']
            if 'status' in filters:
                params['status'] = filters['status']
            if 'limit' in filters:
                params['limit'] = filters['limit']
            if 'offset' in filters:
                params['offset'] = filters['offset']

        response = requests.get(
            f'{self.base_url}/import/jobs',
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        return response.json()

    def wait_for_job_completion(self, job_id: str, poll_interval: int = 5) -> Dict[str, Any]:
        """Wait for a job to complete"""
        while True:
            job = self.get_import_job_status(job_id)
            
            if job['status'] in ['completed', 'failed']:
                return job
            
            time.sleep(poll_interval)

# Usage example
client = KYBImportClient('your-api-key')

# Import request
import_request = {
    'business_id': 'business_123',
    'import_type': 'business_verifications',
    'format': 'json',
    'mode': 'upsert',
    'data': {
        'verifications': [
            {
                'business_name': 'Acme Corporation',
                'address': '123 Main St, Anytown, ST 12345',
                'phone': '+1-555-123-4567',
                'email': 'contact@acme.com'
            }
        ]
    }
}

# Immediate import
try:
    result = client.import_data(import_request)
    print(f"Import completed: {result['success_count']}/{result['record_count']} records processed")
except requests.exceptions.RequestException as e:
    print(f"Import failed: {e}")

# Background job
try:
    job = client.create_import_job(import_request)
    print(f"Job created: {job['id']}")
    
    completed_job = client.wait_for_job_completion(job['id'])
    print(f"Job completed: {completed_job['success_count']}/{completed_job['record_count']} records processed")
except requests.exceptions.RequestException as e:
    print(f"Job failed: {e}")
```

### React Hook

```typescript
import { useState, useCallback } from 'react';

interface UseKYBImportOptions {
  apiKey: string;
  baseUrl?: string;
}

interface UseKYBImportReturn {
  importData: (request: ImportRequest) => Promise<ImportResponse>;
  createImportJob: (request: ImportRequest) => Promise<ImportJob>;
  getImportJobStatus: (jobId: string) => Promise<ImportJob>;
  listImportJobs: (filters?: ImportJobFilters) => Promise<ListImportJobsResponse>;
  loading: boolean;
  error: string | null;
}

export function useKYBImport({ apiKey, baseUrl = 'https://api.kyb-platform.com/v1' }: UseKYBImportOptions): UseKYBImportReturn {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const makeRequest = useCallback(async (endpoint: string, options: RequestInit = {}) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`${baseUrl}${endpoint}`, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      if (!response.ok) {
        throw new Error(`Request failed: ${response.statusText}`);
      }

      return await response.json();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [apiKey, baseUrl]);

  const importData = useCallback(async (request: ImportRequest): Promise<ImportResponse> => {
    return makeRequest('/import', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }, [makeRequest]);

  const createImportJob = useCallback(async (request: ImportRequest): Promise<ImportJob> => {
    return makeRequest('/import/job', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }, [makeRequest]);

  const getImportJobStatus = useCallback(async (jobId: string): Promise<ImportJob> => {
    return makeRequest(`/import/job/${jobId}`);
  }, [makeRequest]);

  const listImportJobs = useCallback(async (filters: ImportJobFilters = {}): Promise<ListImportJobsResponse> => {
    const params = new URLSearchParams();
    if (filters.businessId) params.append('business_id', filters.businessId);
    if (filters.status) params.append('status', filters.status);
    if (filters.limit) params.append('limit', filters.limit.toString());
    if (filters.offset) params.append('offset', filters.offset.toString());

    return makeRequest(`/import/jobs?${params}`);
  }, [makeRequest]);

  return {
    importData,
    createImportJob,
    getImportJobStatus,
    listImportJobs,
    loading,
    error,
  };
}

// Usage in React component
function ImportComponent() {
  const { importData, createImportJob, loading, error } = useKYBImport({
    apiKey: 'your-api-key'
  });

  const handleImport = async () => {
    const request: ImportRequest = {
      business_id: 'business_123',
      import_type: 'business_verifications',
      format: 'json',
      mode: 'upsert',
      data: {
        verifications: [
          {
            business_name: 'Acme Corporation',
            address: '123 Main St, Anytown, ST 12345',
            phone: '+1-555-123-4567',
            email: 'contact@acme.com'
          }
        ]
      }
    };

    try {
      const result = await importData(request);
      console.log(`Import completed: ${result.success_count}/${result.record_count} records processed`);
    } catch (err) {
      console.error('Import failed:', err);
    }
  };

  return (
    <div>
      <button onClick={handleImport} disabled={loading}>
        {loading ? 'Importing...' : 'Import Data'}
      </button>
      {error && <div className="error">{error}</div>}
    </div>
  );
}
```

## Best Practices

### Performance

1. **Use Background Jobs for Large Datasets**: For imports with more than 1000 records, use the background job endpoint to avoid timeout issues.

2. **Batch Processing**: Group related records together in a single import request to reduce API calls.

3. **Validation Rules**: Define validation rules to catch data quality issues early and reduce processing time.

4. **Transform Rules**: Use transformation rules to standardize data formats before import.

### Error Handling

1. **Check Response Status**: Always verify the response status and handle errors appropriately.

2. **Monitor Job Progress**: For background jobs, implement polling to monitor progress and handle failures.

3. **Retry Logic**: Implement exponential backoff for transient failures.

4. **Error Logging**: Log import errors for debugging and data quality improvement.

### Data Quality

1. **Pre-validation**: Validate data before sending to the API to reduce error rates.

2. **Data Cleaning**: Clean and standardize data formats (phone numbers, emails, addresses) before import.

3. **Conflict Resolution**: Choose appropriate conflict policies based on your use case.

4. **Dry Run**: Use dry run mode to validate data without making changes.

### Security

1. **API Key Management**: Store API keys securely and rotate them regularly.

2. **Data Encryption**: Ensure sensitive data is encrypted in transit and at rest.

3. **Access Control**: Implement proper access controls for import operations.

4. **Audit Logging**: Monitor import activities for security and compliance purposes.

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Import Success Rate**: Track the percentage of successful imports
2. **Processing Time**: Monitor import processing duration
3. **Error Rates**: Track validation and processing errors
4. **Job Queue Length**: Monitor background job queue size
5. **API Response Times**: Track endpoint response times

### Recommended Alerts

1. **High Error Rate**: Alert when import error rate exceeds 10%
2. **Long Processing Time**: Alert when imports take longer than expected
3. **Job Failures**: Alert when background jobs fail
4. **API Errors**: Alert on 5xx errors or high error rates

### Integration with Monitoring Tools

```typescript
// Example: Prometheus metrics
const importMetrics = {
  importRequestsTotal: new Counter({
    name: 'kyb_import_requests_total',
    help: 'Total number of import requests',
    labelNames: ['import_type', 'format', 'status']
  }),
  
  importProcessingDuration: new Histogram({
    name: 'kyb_import_processing_duration_seconds',
    help: 'Import processing duration in seconds',
    labelNames: ['import_type', 'format']
  }),
  
  importRecordsProcessed: new Counter({
    name: 'kyb_import_records_processed_total',
    help: 'Total number of records processed',
    labelNames: ['import_type', 'status']
  })
};

// Usage in client
client.importData(request).then(result => {
  importMetrics.importRequestsTotal.inc({ 
    import_type: request.import_type, 
    format: request.format, 
    status: 'success' 
  });
  
  importMetrics.importRecordsProcessed.inc({ 
    import_type: request.import_type, 
    status: 'success' 
  }, result.success_count);
});
```

## Troubleshooting

### Common Issues

1. **Validation Errors**
   - Check that all required fields are present
   - Verify data format compliance
   - Review validation rule configuration

2. **Processing Failures**
   - Check API key validity and permissions
   - Verify business ID exists and is accessible
   - Review data format and structure

3. **Job Timeouts**
   - Use background jobs for large datasets
   - Implement proper polling intervals
   - Monitor job progress and status

4. **Rate Limiting**
   - Implement exponential backoff
   - Use background jobs for bulk operations
   - Monitor rate limit headers

### Debugging Tips

1. **Enable Detailed Logging**: Use the metadata field to include debugging information
2. **Use Dry Run Mode**: Test imports without making changes
3. **Check Response Headers**: Review headers for additional information
4. **Monitor Job Status**: Track background job progress and errors

### Support

For additional support:

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Developer Community**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

## Rate Limits

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| `/v1/import` | 100 requests | 1 minute |
| `/v1/import/job` | 50 requests | 1 minute |
| `/v1/import/job/{job_id}` | 1000 requests | 1 minute |
| `/v1/import/jobs` | 1000 requests | 1 minute |

Rate limit headers are included in all responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1703124000
```

When rate limits are exceeded, the API returns a 429 status code with retry information.
