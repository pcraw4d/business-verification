# Data Quality API Documentation

## Overview

The Data Quality API provides comprehensive data quality management capabilities for the KYB Platform. This API enables organizations to define, execute, and monitor data quality checks across their datasets, ensuring data integrity, accuracy, and compliance with business rules.

### Key Features

- **Multiple Quality Check Types**: Support for completeness, accuracy, consistency, validity, timeliness, uniqueness, integrity, and custom quality checks
- **Advanced Quality Rules**: Configurable quality rules with expressions, parameters, and tolerance settings
- **Quality Scoring**: Comprehensive quality scoring with weighted severity levels and overall quality metrics
- **Background Processing**: Asynchronous quality check execution with progress tracking and status monitoring
- **Quality Monitoring**: Real-time quality monitoring with thresholds, alerts, and notifications
- **Quality Reporting**: Detailed quality reports with trends, recommendations, and actionable insights
- **Quality Actions**: Automated quality actions based on check results and severity levels

### Quality Check Types

| Type | Description | Use Case |
|------|-------------|----------|
| `completeness` | Checks for missing required fields | Ensure all required data is present |
| `accuracy` | Validates data accuracy and correctness | Verify data matches expected values |
| `consistency` | Ensures data consistency across records | Maintain data coherence |
| `validity` | Validates data format and structure | Ensure data meets format requirements |
| `timeliness` | Checks data freshness and update frequency | Monitor data currency |
| `uniqueness` | Validates unique constraints | Prevent duplicate records |
| `integrity` | Ensures referential integrity | Maintain data relationships |
| `custom` | Custom quality checks with user-defined logic | Implement business-specific rules |

### Quality Severity Levels

| Level | Weight | Description | Action |
|-------|--------|-------------|--------|
| `critical` | 4.0 | Critical quality issues | Immediate action required |
| `high` | 3.0 | High priority issues | Prompt attention needed |
| `medium` | 2.0 | Medium priority issues | Monitor and address |
| `low` | 1.0 | Low priority issues | Optional improvements |

## Authentication

All API endpoints require authentication using API keys or JWT tokens.

```http
Authorization: Bearer <your-api-key>
```

## Response Format

All API responses follow a consistent JSON format:

```json
{
  "id": "quality_1234567890",
  "name": "Customer Data Quality Check",
  "status": "completed",
  "overall_score": 0.92,
  "checks": [...],
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

## API Endpoints

### 1. Create Quality Check

**POST** `/quality`

Creates and executes a data quality check immediately.

#### Request Body

```json
{
  "name": "Customer Data Quality Check",
  "description": "Comprehensive quality check for customer data",
  "dataset": "customer_data",
  "checks": [
    {
      "name": "email_completeness",
      "type": "completeness",
      "description": "Check for missing email addresses",
      "severity": "high",
      "parameters": {
        "required_fields": ["id", "name", "email"]
      },
      "rules": [
        {
          "name": "email_not_null",
          "description": "Email field must not be null",
          "expression": "email IS NOT NULL",
          "parameters": {
            "field": "email"
          },
          "expected": "non_null",
          "tolerance": 0.0
        }
      ],
      "conditions": [
        {
          "name": "email_format",
          "description": "Check email format",
          "operator": "regex_match",
          "value": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
          "field": "email",
          "function": "regex"
        }
      ],
      "actions": [
        {
          "name": "log_issue",
          "type": "log",
          "description": "Log quality issues",
          "parameters": {
            "level": "warning"
          },
          "condition": "score < 0.9",
          "priority": 1
        }
      ]
    }
  ],
  "thresholds": {
    "overall_score": 0.9,
    "critical_checks": 1.0,
    "high_checks": 0.95,
    "medium_checks": 0.9,
    "low_checks": 0.8,
    "pass_rate": 0.95,
    "fail_rate": 0.05,
    "warning_rate": 0.1
  },
  "notifications": {
    "email": ["admin@company.com"],
    "slack": ["#data-quality"],
    "conditions": {
      "critical": ["email", "slack"],
      "high": ["email"]
    },
    "template": "quality_alert_template"
  },
  "metadata": {
    "department": "data_team",
    "priority": "high"
  }
}
```

#### Response

```json
{
  "id": "quality_1234567890",
  "name": "Customer Data Quality Check",
  "status": "completed",
  "overall_score": 0.92,
  "checks": [
    {
      "name": "email_completeness",
      "type": "completeness",
      "status": "passed",
      "score": 0.95,
      "severity": "high",
      "issues": [],
      "metrics": {
        "total_records": 1000,
        "complete_records": 950,
        "missing_records": 50,
        "completeness_rate": 0.95
      },
      "execution_time": "150ms",
      "timestamp": "2024-12-19T10:30:00Z"
    }
  ],
  "summary": {
    "total_checks": 1,
    "passed_checks": 1,
    "failed_checks": 0,
    "warning_checks": 0,
    "error_checks": 0,
    "pass_rate": 1.0,
    "fail_rate": 0.0,
    "warning_rate": 0.0,
    "error_rate": 0.0,
    "total_issues": 0,
    "critical_issues": 0,
    "high_issues": 0,
    "medium_issues": 0,
    "low_issues": 0,
    "metrics": {}
  },
  "metadata": {
    "department": "data_team",
    "priority": "high"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 2. Get Quality Check

**GET** `/quality?id={id}`

Retrieves details of a specific quality check.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Quality check ID |

#### Response

```json
{
  "id": "quality_1234567890",
  "name": "Customer Data Quality Check",
  "status": "completed",
  "overall_score": 0.92,
  "checks": [...],
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 3. List Quality Checks

**GET** `/quality`

Lists all quality checks with pagination support.

#### Response

```json
{
  "quality_checks": [
    {
      "id": "quality_1234567890",
      "name": "Customer Data Quality Check",
      "status": "completed",
      "overall_score": 0.92,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 4. Create Quality Job

**POST** `/quality/jobs`

Creates a background quality check job for asynchronous processing.

#### Request Body

```json
{
  "name": "Large Dataset Quality Check",
  "description": "Background quality check for large customer dataset",
  "dataset": "customer_data_large",
  "checks": [
    {
      "name": "completeness_check",
      "type": "completeness",
      "description": "Check for missing required fields",
      "severity": "high",
      "parameters": {
        "required_fields": ["id", "name", "email"]
      }
    }
  ],
  "thresholds": {
    "overall_score": 0.9,
    "critical_checks": 1.0,
    "high_checks": 0.95,
    "medium_checks": 0.9,
    "low_checks": 0.8,
    "pass_rate": 0.95,
    "fail_rate": 0.05,
    "warning_rate": 0.1
  },
  "notifications": {
    "email": ["admin@company.com"],
    "conditions": {
      "critical": ["email"]
    },
    "template": "quality_job_template"
  },
  "metadata": {
    "priority": "high"
  }
}
```

#### Response

```json
{
  "id": "quality_job_1234567890",
  "request_id": "req_1234567890",
  "status": "pending",
  "progress": 0,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "priority": "high"
  }
}
```

### 5. Get Quality Job Status

**GET** `/quality/jobs?id={id}`

Retrieves the status of a background quality job.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Job ID |

#### Response

```json
{
  "id": "quality_job_1234567890",
  "request_id": "req_1234567890",
  "status": "completed",
  "progress": 100,
  "result": {
    "id": "quality_job_1234567890",
    "name": "Large Dataset Quality Check",
    "status": "completed",
    "overall_score": 0.92,
    "checks": [...],
    "summary": {...},
    "metadata": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:35:00Z",
  "completed_at": "2024-12-19T10:35:00Z",
  "metadata": {
    "priority": "high"
  }
}
```

### 6. List Quality Jobs

**GET** `/quality/jobs`

Lists all quality jobs with pagination support.

#### Response

```json
{
  "jobs": [
    {
      "id": "quality_job_1234567890",
      "request_id": "req_1234567890",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:35:00Z",
      "completed_at": "2024-12-19T10:35:00Z"
    }
  ],
  "total": 1
}
```

## Error Responses

### Validation Errors

```json
{
  "error": "name is required"
}
```

### Not Found Errors

```json
{
  "error": "Quality check not found"
}
```

### Server Errors

```json
{
  "error": "Internal server error"
}
```

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class DataQualityClient {
  constructor(baseURL, apiKey) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createQualityCheck(qualityRequest) {
    try {
      const response = await this.client.post('/quality', qualityRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Quality check creation failed: ${error.response?.data?.error || error.message}`);
    }
  }

  async getQualityCheck(id) {
    try {
      const response = await this.client.get(`/quality?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Quality check retrieval failed: ${error.response?.data?.error || error.message}`);
    }
  }

  async createQualityJob(qualityRequest) {
    try {
      const response = await this.client.post('/quality/jobs', qualityRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Quality job creation failed: ${error.response?.data?.error || error.message}`);
    }
  }

  async getQualityJobStatus(id) {
    try {
      const response = await this.client.get(`/quality/jobs?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Quality job status retrieval failed: ${error.response?.data?.error || error.message}`);
    }
  }

  async waitForJobCompletion(jobId, maxWaitTime = 300000) {
    const startTime = Date.now();
    
    while (Date.now() - startTime < maxWaitTime) {
      const job = await this.getQualityJobStatus(jobId);
      
      if (job.status === 'completed') {
        return job.result;
      } else if (job.status === 'failed') {
        throw new Error(`Job failed: ${job.error}`);
      }
      
      await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds
    }
    
    throw new Error('Job timeout exceeded');
  }
}

// Usage example
async function runQualityCheck() {
  const client = new DataQualityClient('https://api.kyb-platform.com', 'your-api-key');
  
  const qualityRequest = {
    name: 'Customer Data Quality Check',
    description: 'Check customer data quality',
    dataset: 'customer_data',
    checks: [
      {
        name: 'email_completeness',
        type: 'completeness',
        description: 'Check for missing email addresses',
        severity: 'high',
        parameters: {
          required_fields: ['id', 'name', 'email']
        }
      }
    ],
    thresholds: {
      overall_score: 0.9,
      critical_checks: 1.0,
      high_checks: 0.95,
      medium_checks: 0.9,
      low_checks: 0.8,
      pass_rate: 0.95,
      fail_rate: 0.05,
      warning_rate: 0.1
    },
    notifications: {
      email: ['admin@company.com'],
      conditions: {
        critical: ['email']
      },
      template: 'quality_alert_template'
    }
  };

  try {
    // For immediate processing
    const result = await client.createQualityCheck(qualityRequest);
    console.log('Quality check completed:', result.overall_score);
    
    // For background processing
    const job = await client.createQualityJob(qualityRequest);
    console.log('Job created:', job.id);
    
    const jobResult = await client.waitForJobCompletion(job.id);
    console.log('Job completed:', jobResult.overall_score);
  } catch (error) {
    console.error('Quality check failed:', error.message);
  }
}
```

### Python

```python
import requests
import time
from typing import Dict, Any, Optional

class DataQualityClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }
    
    def create_quality_check(self, quality_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create and execute a quality check immediately."""
        try:
            response = requests.post(
                f'{self.base_url}/quality',
                json=quality_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f"Quality check creation failed: {e}")
    
    def get_quality_check(self, check_id: str) -> Dict[str, Any]:
        """Retrieve a quality check by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/quality?id={check_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f"Quality check retrieval failed: {e}")
    
    def create_quality_job(self, quality_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create a background quality job."""
        try:
            response = requests.post(
                f'{self.base_url}/quality/jobs',
                json=quality_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f"Quality job creation failed: {e}")
    
    def get_quality_job_status(self, job_id: str) -> Dict[str, Any]:
        """Get the status of a quality job."""
        try:
            response = requests.get(
                f'{self.base_url}/quality/jobs?id={job_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f"Quality job status retrieval failed: {e}")
    
    def wait_for_job_completion(self, job_id: str, max_wait_time: int = 300) -> Dict[str, Any]:
        """Wait for a job to complete and return the result."""
        start_time = time.time()
        
        while time.time() - start_time < max_wait_time:
            job = self.get_quality_job_status(job_id)
            
            if job['status'] == 'completed':
                return job['result']
            elif job['status'] == 'failed':
                raise Exception(f"Job failed: {job.get('error', 'Unknown error')}")
            
            time.sleep(5)  # Wait 5 seconds
        
        raise Exception("Job timeout exceeded")

# Usage example
def run_quality_check():
    client = DataQualityClient('https://api.kyb-platform.com', 'your-api-key')
    
    quality_request = {
        "name": "Customer Data Quality Check",
        "description": "Check customer data quality",
        "dataset": "customer_data",
        "checks": [
            {
                "name": "email_completeness",
                "type": "completeness",
                "description": "Check for missing email addresses",
                "severity": "high",
                "parameters": {
                    "required_fields": ["id", "name", "email"]
                }
            }
        ],
        "thresholds": {
            "overall_score": 0.9,
            "critical_checks": 1.0,
            "high_checks": 0.95,
            "medium_checks": 0.9,
            "low_checks": 0.8,
            "pass_rate": 0.95,
            "fail_rate": 0.05,
            "warning_rate": 0.1
        },
        "notifications": {
            "email": ["admin@company.com"],
            "conditions": {
                "critical": ["email"]
            },
            "template": "quality_alert_template"
        }
    }
    
    try:
        # For immediate processing
        result = client.create_quality_check(quality_request)
        print(f"Quality check completed: {result['overall_score']}")
        
        # For background processing
        job = client.create_quality_job(quality_request)
        print(f"Job created: {job['id']}")
        
        job_result = client.wait_for_job_completion(job['id'])
        print(f"Job completed: {job_result['overall_score']}")
        
    except Exception as e:
        print(f"Quality check failed: {e}")

if __name__ == "__main__":
    run_quality_check()
```

### React/TypeScript

```typescript
interface QualityCheck {
  name: string;
  type: 'completeness' | 'accuracy' | 'consistency' | 'validity' | 'timeliness' | 'uniqueness' | 'integrity' | 'custom';
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  parameters?: Record<string, any>;
  rules?: QualityRule[];
  conditions?: QualityCondition[];
  actions?: QualityAction[];
}

interface QualityRequest {
  name: string;
  description: string;
  dataset: string;
  checks: QualityCheck[];
  thresholds: QualityThresholds;
  notifications: QualityNotifications;
  metadata?: Record<string, any>;
}

interface QualityResponse {
  id: string;
  name: string;
  status: string;
  overall_score: number;
  checks: QualityCheckResult[];
  summary: QualitySummary;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

interface QualityJob {
  id: string;
  request_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  result?: QualityResponse;
  error?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  metadata: Record<string, any>;
}

class DataQualityService {
  private baseURL: string;
  private apiKey: string;

  constructor(baseURL: string, apiKey: string) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  async createQualityCheck(qualityRequest: QualityRequest): Promise<QualityResponse> {
    return this.request<QualityResponse>('/quality', {
      method: 'POST',
      body: JSON.stringify(qualityRequest),
    });
  }

  async getQualityCheck(id: string): Promise<QualityResponse> {
    return this.request<QualityResponse>(`/quality?id=${id}`);
  }

  async listQualityChecks(): Promise<{ quality_checks: QualityResponse[]; total: number }> {
    return this.request<{ quality_checks: QualityResponse[]; total: number }>('/quality');
  }

  async createQualityJob(qualityRequest: QualityRequest): Promise<QualityJob> {
    return this.request<QualityJob>('/quality/jobs', {
      method: 'POST',
      body: JSON.stringify(qualityRequest),
    });
  }

  async getQualityJobStatus(id: string): Promise<QualityJob> {
    return this.request<QualityJob>(`/quality/jobs?id=${id}`);
  }

  async listQualityJobs(): Promise<{ jobs: QualityJob[]; total: number }> {
    return this.request<{ jobs: QualityJob[]; total: number }>('/quality/jobs');
  }

  async waitForJobCompletion(jobId: string, maxWaitTime: number = 300000): Promise<QualityResponse> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < maxWaitTime) {
      const job = await this.getQualityJobStatus(jobId);
      
      if (job.status === 'completed') {
        return job.result!;
      } else if (job.status === 'failed') {
        throw new Error(`Job failed: ${job.error || 'Unknown error'}`);
      }
      
      await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds
    }
    
    throw new Error('Job timeout exceeded');
  }
}

// React Hook for Quality Checks
import { useState, useEffect } from 'react';

export function useQualityCheck(qualityService: DataQualityService) {
  const [qualityChecks, setQualityChecks] = useState<QualityResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchQualityChecks = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await qualityService.listQualityChecks();
      setQualityChecks(response.quality_checks);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch quality checks');
    } finally {
      setLoading(false);
    }
  };

  const createQualityCheck = async (qualityRequest: QualityRequest) => {
    try {
      setLoading(true);
      setError(null);
      const result = await qualityService.createQualityCheck(qualityRequest);
      setQualityChecks(prev => [result, ...prev]);
      return result;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create quality check');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchQualityChecks();
  }, []);

  return {
    qualityChecks,
    loading,
    error,
    createQualityCheck,
    refresh: fetchQualityChecks,
  };
}

// React Component Example
import React, { useState } from 'react';

interface QualityCheckFormProps {
  onSubmit: (qualityRequest: QualityRequest) => Promise<void>;
  loading: boolean;
}

export function QualityCheckForm({ onSubmit, loading }: QualityCheckFormProps) {
  const [formData, setFormData] = useState<QualityRequest>({
    name: '',
    description: '',
    dataset: '',
    checks: [],
    thresholds: {
      overall_score: 0.9,
      critical_checks: 1.0,
      high_checks: 0.95,
      medium_checks: 0.9,
      low_checks: 0.8,
      pass_rate: 0.95,
      fail_rate: 0.05,
      warning_rate: 0.1,
    },
    notifications: {
      email: [],
      conditions: {},
      template: '',
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="quality-check-form">
      <div className="form-group">
        <label htmlFor="name">Quality Check Name</label>
        <input
          type="text"
          id="name"
          value={formData.name}
          onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
          required
        />
      </div>

      <div className="form-group">
        <label htmlFor="description">Description</label>
        <textarea
          id="description"
          value={formData.description}
          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
          required
        />
      </div>

      <div className="form-group">
        <label htmlFor="dataset">Dataset</label>
        <input
          type="text"
          id="dataset"
          value={formData.dataset}
          onChange={(e) => setFormData(prev => ({ ...prev, dataset: e.target.value }))}
          required
        />
      </div>

      <button type="submit" disabled={loading}>
        {loading ? 'Creating...' : 'Create Quality Check'}
      </button>
    </form>
  );
}
```

## Best Practices

### 1. Quality Check Design

- **Start Simple**: Begin with basic completeness and validity checks before adding complex rules
- **Use Appropriate Severity Levels**: Assign severity based on business impact, not technical complexity
- **Define Clear Thresholds**: Set realistic thresholds that balance quality requirements with operational constraints
- **Document Rules**: Provide clear descriptions for all quality rules and their business justification

### 2. Performance Optimization

- **Batch Processing**: Use background jobs for large datasets or complex quality checks
- **Parallel Execution**: Run independent quality checks in parallel when possible
- **Caching**: Cache quality check results for frequently accessed datasets
- **Monitoring**: Monitor quality check execution times and optimize slow checks

### 3. Error Handling

- **Graceful Degradation**: Handle quality check failures without affecting data processing
- **Retry Logic**: Implement retry mechanisms for transient failures
- **Fallback Strategies**: Provide alternative quality checks when primary checks fail
- **Error Reporting**: Ensure quality check errors are properly logged and reported

### 4. Security Considerations

- **Input Validation**: Validate all quality check parameters and expressions
- **Access Control**: Implement proper access controls for quality check creation and execution
- **Data Privacy**: Ensure quality checks don't expose sensitive data in logs or reports
- **Audit Trail**: Maintain audit trails for all quality check operations

### 5. Monitoring and Alerting

- **Quality Metrics**: Track key quality metrics over time to identify trends
- **Alert Thresholds**: Set up alerts for quality degradation below acceptable levels
- **Dashboard**: Create dashboards to visualize quality trends and issues
- **Escalation**: Implement escalation procedures for critical quality issues

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Immediate Quality Checks**: 100 requests per minute per API key
- **Background Jobs**: 50 job creations per minute per API key
- **Status Queries**: 200 requests per minute per API key

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Monitoring and Metrics

### Key Metrics

- **Quality Check Success Rate**: Percentage of quality checks that complete successfully
- **Average Execution Time**: Mean time to complete quality checks
- **Quality Score Trends**: Changes in overall quality scores over time
- **Issue Distribution**: Distribution of quality issues by severity and type
- **Job Completion Rate**: Percentage of background jobs that complete successfully

### Health Checks

Monitor the following endpoints for system health:

- **GET** `/health` - Overall system health
- **GET** `/metrics` - System metrics and performance indicators

## Troubleshooting

### Common Issues

1. **Quality Check Timeout**
   - **Cause**: Large datasets or complex quality rules
   - **Solution**: Use background jobs for large datasets, optimize quality rules

2. **High False Positive Rate**
   - **Cause**: Overly strict quality rules or incorrect thresholds
   - **Solution**: Review and adjust quality rules, validate threshold settings

3. **Performance Degradation**
   - **Cause**: Inefficient quality rules or resource constraints
   - **Solution**: Optimize quality rules, increase system resources

4. **Missing Quality Issues**
   - **Cause**: Incorrect quality rule logic or data format issues
   - **Solution**: Validate quality rule expressions, check data format compliance

### Debug Information

Enable debug logging by setting the `X-Debug` header:

```http
X-Debug: true
```

Debug responses include additional information:

```json
{
  "id": "quality_1234567890",
  "name": "Customer Data Quality Check",
  "status": "completed",
  "overall_score": 0.92,
  "checks": [...],
  "summary": {...},
  "debug": {
    "execution_time": "1.5s",
    "memory_usage": "256MB",
    "rule_evaluations": 150,
    "data_records_processed": 10000
  }
}
```

## Support

For technical support and questions:

- **Documentation**: [https://docs.kyb-platform.com/api/quality](https://docs.kyb-platform.com/api/quality)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

## Version History

### v1.0.0 (Current)
- Initial release of Data Quality API
- Support for 8 quality check types
- Background job processing
- Quality scoring and reporting
- Comprehensive monitoring and alerting

### Upcoming Features
- Real-time quality streaming
- Machine learning-powered quality recommendations
- Advanced quality rule templates
- Integration with external quality tools
- Quality data lineage tracking
