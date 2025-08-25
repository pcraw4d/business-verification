# Data Validation API Documentation

## Overview

The Data Validation API provides comprehensive data validation capabilities for the KYB Platform. This API enables organizations to define, execute, and monitor data validation rules across their datasets, ensuring data integrity, accuracy, and compliance with business rules and schemas.

### Key Features

- **Multiple Validation Types**: Support for schema, rule, custom, format, business, compliance, cross-field, and reference validations
- **Advanced Schema Validation**: JSON Schema-based validation with custom properties, patterns, formats, and ranges
- **Custom Validators**: Support for custom validation logic in multiple programming languages
- **Validation Scoring**: Comprehensive validation scoring with weighted severity levels and overall quality metrics
- **Background Processing**: Asynchronous validation execution with progress tracking and status monitoring
- **Validation Reporting**: Detailed validation reports with trends, recommendations, and actionable insights

### Supported Validation Types

| Type | Description | Use Case |
|------|-------------|----------|
| `schema` | JSON Schema validation | Data structure and type validation |
| `rule` | Business rule validation | Custom business logic validation |
| `custom` | Custom validator execution | Complex validation logic |
| `format` | Format validation | Email, phone, date format validation |
| `business` | Business rule validation | Domain-specific business rules |
| `compliance` | Compliance validation | Regulatory and policy compliance |
| `cross_field` | Cross-field validation | Relationships between fields |
| `reference` | Reference validation | Foreign key and reference integrity |

### Supported Validation Severities

| Severity | Weight | Description |
|----------|--------|-------------|
| `critical` | 4.0 | Critical validation failures |
| `high` | 3.0 | High priority validation issues |
| `medium` | 2.0 | Medium priority validation warnings |
| `low` | 1.0 | Low priority validation suggestions |

## Authentication

All API endpoints require authentication using API keys. Include your API key in the request headers:

```http
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All API responses follow a consistent JSON format:

### Success Response
```json
{
  "id": "validation_1234567890",
  "name": "Customer Data Validation",
  "status": "completed",
  "overall_score": 0.95,
  "validations": [...],
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### Error Response
```json
{
  "error": "Validation failed",
  "message": "Detailed error message",
  "code": "VALIDATION_ERROR",
  "details": {...}
}
```

## API Endpoints

### 1. Create Validation

**POST** `/validation`

Creates and executes a data validation immediately.

#### Request Body

```json
{
  "name": "Customer Data Validation",
  "description": "Validate customer data for completeness and accuracy",
  "dataset": "customer_data",
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30,
    "phone": "+1-555-123-4567"
  },
  "schema": {
    "type": "object",
    "version": "1.0",
    "properties": {
      "name": {
        "type": "string",
        "description": "Customer name",
        "required": true,
        "min_length": 2,
        "max_length": 100
      },
      "email": {
        "type": "string",
        "description": "Customer email",
        "required": true,
        "format": "email"
      },
      "age": {
        "type": "integer",
        "description": "Customer age",
        "required": true,
        "min_value": 18,
        "max_value": 120
      },
      "phone": {
        "type": "string",
        "description": "Customer phone",
        "required": false,
        "format": "phone"
      }
    },
    "required": ["name", "email", "age"],
    "patterns": {
      "email": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
      "phone": "^\\+[1-9]\\d{1,14}$"
    },
    "formats": {
      "email": "email",
      "phone": "phone"
    }
  },
  "rules": [
    {
      "name": "email_format_rule",
      "type": "format",
      "description": "Validate email format",
      "severity": "high",
      "expression": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
      "parameters": {
        "field": "email"
      },
      "conditions": [
        {
          "name": "email_present",
          "description": "Check if email is present",
          "operator": "not_empty",
          "value": true,
          "field": "email",
          "function": "exists"
        }
      ],
      "actions": [
        {
          "name": "log_error",
          "type": "log",
          "description": "Log validation error",
          "parameters": {
            "level": "error"
          },
          "condition": "validation_failed",
          "priority": 1
        }
      ],
      "enabled": true
    },
    {
      "name": "age_business_rule",
      "type": "business",
      "description": "Validate age is 18 or older",
      "severity": "critical",
      "expression": "age >= 18",
      "parameters": {
        "field": "age"
      },
      "enabled": true
    },
    {
      "name": "name_completeness_rule",
      "type": "rule",
      "description": "Validate name completeness",
      "severity": "medium",
      "expression": "name.length >= 2",
      "parameters": {
        "field": "name"
      },
      "enabled": true
    }
  ],
  "validators": [
    {
      "name": "custom_email_validator",
      "description": "Custom email validation logic",
      "type": "javascript",
      "code": "function validate(data) { return data.email.includes('@') && data.email.includes('.'); }",
      "language": "javascript",
      "parameters": {},
      "timeout": "5s",
      "enabled": true
    },
    {
      "name": "business_logic_validator",
      "description": "Complex business logic validation",
      "type": "python",
      "code": "def validate(data):\n    return data.get('age', 0) >= 18 and '@' in data.get('email', '')",
      "language": "python",
      "parameters": {},
      "timeout": "10s",
      "enabled": true
    }
  ],
  "options": {
    "stop_on_first_error": false,
    "continue_on_error": true,
    "max_errors": 100,
    "timeout": "30s",
    "parallel": true,
    "batch_size": 1000,
    "cache_results": true,
    "log_level": "info",
    "custom": {
      "validation_mode": "strict",
      "enable_caching": true
    }
  },
  "metadata": {
    "department": "data_team",
    "priority": "high",
    "owner": "data_analyst"
  }
}
```

#### Response

```json
{
  "id": "validation_1234567890",
  "name": "Customer Data Validation",
  "status": "completed",
  "overall_score": 0.95,
  "validations": [
    {
      "name": "schema_validation",
      "type": "schema",
      "status": "passed",
      "severity": "high",
      "score": 0.95,
      "errors": [],
      "warnings": [
        {
          "id": "schema_warning_123",
          "type": "format_warning",
          "message": "Email format could be improved",
          "severity": "medium",
          "field": "email",
          "value": "john@example",
          "suggestion": "Use a valid email format like 'john@example.com'",
          "path": "data.email",
          "context": {
            "format": "email"
          },
          "timestamp": "2024-12-19T10:30:00Z"
        }
      ],
      "metrics": {
        "total_fields": 4,
        "validated_fields": 4,
        "invalid_fields": 0,
        "validation_rate": 1.0
      },
      "execution_time": "50ms",
      "timestamp": "2024-12-19T10:30:00Z"
    },
    {
      "name": "email_format_rule",
      "type": "format",
      "status": "passed",
      "severity": "high",
      "score": 0.90,
      "errors": [],
      "warnings": [],
      "metrics": {
        "rule_type": "format",
        "rule_severity": "high",
        "execution_time": "30ms",
        "success_rate": 0.90
      },
      "execution_time": "30ms",
      "timestamp": "2024-12-19T10:30:00Z"
    },
    {
      "name": "age_business_rule",
      "type": "business",
      "status": "passed",
      "severity": "critical",
      "score": 1.0,
      "errors": [],
      "warnings": [],
      "metrics": {
        "rule_type": "business",
        "rule_severity": "critical",
        "execution_time": "25ms",
        "success_rate": 1.0
      },
      "execution_time": "25ms",
      "timestamp": "2024-12-19T10:30:00Z"
    },
    {
      "name": "custom_email_validator",
      "type": "custom",
      "status": "passed",
      "severity": "medium",
      "score": 0.88,
      "errors": [],
      "warnings": [],
      "metrics": {
        "validator_name": "custom_email_validator",
        "validator_type": "javascript",
        "execution_time": "150ms",
        "success_rate": 0.88
      },
      "execution_time": "150ms",
      "timestamp": "2024-12-19T10:30:00Z"
    }
  ],
  "summary": {
    "total_validations": 4,
    "passed_validations": 4,
    "failed_validations": 0,
    "warning_validations": 0,
    "error_validations": 0,
    "pass_rate": 1.0,
    "fail_rate": 0.0,
    "warning_rate": 0.0,
    "error_rate": 0.0,
    "total_errors": 0,
    "total_warnings": 1,
    "critical_errors": 0,
    "high_errors": 0,
    "medium_errors": 0,
    "low_errors": 0,
    "metrics": {
      "overall_execution_time": "255ms",
      "average_score": 0.93,
      "validation_efficiency": 0.95
    }
  },
  "metadata": {
    "department": "data_team",
    "priority": "high",
    "owner": "data_analyst"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 2. Get Validation

**GET** `/validation?id={id}`

Retrieves details of a specific validation.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Validation ID |

#### Response

```json
{
  "id": "validation_1234567890",
  "name": "Customer Data Validation",
  "status": "completed",
  "overall_score": 0.95,
  "validations": [...],
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 3. List Validations

**GET** `/validation`

Lists all validations.

#### Response

```json
{
  "validations": [
    {
      "id": "validation_1234567890",
      "name": "Customer Data Validation",
      "status": "completed",
      "overall_score": 0.95,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 4. Create Validation Job

**POST** `/validation/jobs`

Creates a background validation job for processing large datasets.

#### Request Body

Same as Create Validation endpoint.

#### Response

```json
{
  "id": "validation_job_1234567890",
  "request_id": "req_1234567890",
  "status": "pending",
  "progress": 0,
  "result": null,
  "error": null,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "completed_at": null,
  "metadata": {
    "department": "data_team",
    "priority": "high"
  }
}
```

### 5. Get Validation Job

**GET** `/validation/jobs?id={id}`

Retrieves the status of a background validation job.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Job ID |

#### Response

```json
{
  "id": "validation_job_1234567890",
  "request_id": "req_1234567890",
  "status": "completed",
  "progress": 100,
  "result": {
    "id": "validation_1234567890",
    "name": "Customer Data Validation",
    "status": "completed",
    "overall_score": 0.95,
    "validations": [...],
    "summary": {...},
    "metadata": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  },
  "error": null,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:35:00Z",
  "completed_at": "2024-12-19T10:35:00Z",
  "metadata": {
    "department": "data_team",
    "priority": "high"
  }
}
```

### 6. List Validation Jobs

**GET** `/validation/jobs`

Lists all background validation jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "validation_job_1234567890",
      "request_id": "req_1234567890",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:35:00Z"
    }
  ],
  "total": 1
}
```

## Error Responses

### Validation Error (400)

```json
{
  "error": "Validation failed",
  "message": "name is required",
  "code": "VALIDATION_ERROR"
}
```

### Not Found Error (404)

```json
{
  "error": "Validation not found",
  "message": "Validation with ID 'invalid_id' not found",
  "code": "NOT_FOUND"
}
```

### Internal Server Error (500)

```json
{
  "error": "Internal server error",
  "message": "An unexpected error occurred",
  "code": "INTERNAL_ERROR"
}
```

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class DataValidationClient {
  constructor(apiKey, baseURL = 'https://api.kyb-platform.com/v3') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  async createValidation(validationRequest) {
    try {
      const response = await axios.post(`${this.baseURL}/validation`, validationRequest, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      return response.data;
    } catch (error) {
      throw new Error(`Validation failed: ${error.response?.data?.message || error.message}`);
    }
  }

  async getValidation(id) {
    try {
      const response = await axios.get(`${this.baseURL}/validation?id=${id}`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get validation: ${error.response?.data?.message || error.message}`);
    }
  }

  async createValidationJob(validationRequest) {
    try {
      const response = await axios.post(`${this.baseURL}/validation/jobs`, validationRequest, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      return response.data;
    } catch (error) {
      throw new Error(`Job creation failed: ${error.response?.data?.message || error.message}`);
    }
  }

  async getValidationJob(id) {
    try {
      const response = await axios.get(`${this.baseURL}/validation/jobs?id=${id}`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`
        }
      });
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get job: ${error.response?.data?.message || error.message}`);
    }
  }

  async waitForJobCompletion(jobId, pollInterval = 5000) {
    return new Promise((resolve, reject) => {
      const checkStatus = async () => {
        try {
          const job = await this.getValidationJob(jobId);
          
          if (job.status === 'completed') {
            resolve(job);
          } else if (job.status === 'failed') {
            reject(new Error(`Job failed: ${job.error}`));
          } else {
            setTimeout(checkStatus, pollInterval);
          }
        } catch (error) {
          reject(error);
        }
      };
      
      checkStatus();
    });
  }
}

// Usage example
async function validateCustomerData() {
  const client = new DataValidationClient('your-api-key');
  
  const validationRequest = {
    name: 'Customer Data Validation',
    description: 'Validate customer data for completeness and accuracy',
    dataset: 'customer_data',
    data: {
      name: 'John Doe',
      email: 'john@example.com',
      age: 30
    },
    rules: [
      {
        name: 'email_format_rule',
        type: 'format',
        description: 'Validate email format',
        severity: 'high',
        expression: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
        parameters: { field: 'email' },
        enabled: true
      }
    ],
    options: {
      stop_on_first_error: false,
      continue_on_error: true,
      max_errors: 100,
      timeout: '30s',
      parallel: true
    }
  };

  try {
    // Immediate validation
    const result = await client.createValidation(validationRequest);
    console.log('Validation completed:', result.overall_score);
    
    // Background job validation
    const job = await client.createValidationJob(validationRequest);
    console.log('Job created:', job.id);
    
    const completedJob = await client.waitForJobCompletion(job.id);
    console.log('Job completed:', completedJob.result.overall_score);
  } catch (error) {
    console.error('Validation error:', error.message);
  }
}
```

### Python

```python
import requests
import time
from typing import Dict, Any, Optional

class DataValidationClient:
    def __init__(self, api_key: str, base_url: str = 'https://api.kyb-platform.com/v3'):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }
    
    def create_validation(self, validation_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create and execute a validation immediately."""
        try:
            response = requests.post(
                f'{self.base_url}/validation',
                json=validation_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Validation failed: {e}')
    
    def get_validation(self, validation_id: str) -> Dict[str, Any]:
        """Get validation details by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/validation?id={validation_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get validation: {e}')
    
    def create_validation_job(self, validation_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create a background validation job."""
        try:
            response = requests.post(
                f'{self.base_url}/validation/jobs',
                json=validation_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Job creation failed: {e}')
    
    def get_validation_job(self, job_id: str) -> Dict[str, Any]:
        """Get job status by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/validation/jobs?id={job_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get job: {e}')
    
    def wait_for_job_completion(self, job_id: str, poll_interval: int = 5) -> Dict[str, Any]:
        """Wait for job completion with polling."""
        while True:
            job = self.get_validation_job(job_id)
            
            if job['status'] == 'completed':
                return job
            elif job['status'] == 'failed':
                raise Exception(f'Job failed: {job.get("error", "Unknown error")}')
            
            time.sleep(poll_interval)

# Usage example
def validate_customer_data():
    client = DataValidationClient('your-api-key')
    
    validation_request = {
        'name': 'Customer Data Validation',
        'description': 'Validate customer data for completeness and accuracy',
        'dataset': 'customer_data',
        'data': {
            'name': 'John Doe',
            'email': 'john@example.com',
            'age': 30
        },
        'schema': {
            'type': 'object',
            'version': '1.0',
            'properties': {
                'name': {
                    'type': 'string',
                    'description': 'Customer name',
                    'required': True,
                    'min_length': 2,
                    'max_length': 100
                },
                'email': {
                    'type': 'string',
                    'description': 'Customer email',
                    'required': True,
                    'format': 'email'
                },
                'age': {
                    'type': 'integer',
                    'description': 'Customer age',
                    'required': True,
                    'min_value': 18,
                    'max_value': 120
                }
            },
            'required': ['name', 'email', 'age']
        },
        'rules': [
            {
                'name': 'email_format_rule',
                'type': 'format',
                'description': 'Validate email format',
                'severity': 'high',
                'expression': '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
                'parameters': {'field': 'email'},
                'enabled': True
            },
            {
                'name': 'age_business_rule',
                'type': 'business',
                'description': 'Validate age is 18 or older',
                'severity': 'critical',
                'expression': 'age >= 18',
                'parameters': {'field': 'age'},
                'enabled': True
            }
        ],
        'options': {
            'stop_on_first_error': False,
            'continue_on_error': True,
            'max_errors': 100,
            'timeout': '30s',
            'parallel': True
        }
    }
    
    try:
        # Immediate validation
        result = client.create_validation(validation_request)
        print(f'Validation completed: {result["overall_score"]}')
        
        # Background job validation
        job = client.create_validation_job(validation_request)
        print(f'Job created: {job["id"]}')
        
        completed_job = client.wait_for_job_completion(job['id'])
        print(f'Job completed: {completed_job["result"]["overall_score"]}')
        
    except Exception as e:
        print(f'Validation error: {e}')

if __name__ == '__main__':
    validate_customer_data()
```

### React/TypeScript

```typescript
interface ValidationRequest {
  name: string;
  description: string;
  dataset: string;
  data: any;
  schema?: ValidationSchema;
  rules: ValidationRule[];
  validators?: CustomValidator[];
  options: ValidationOptions;
  metadata?: Record<string, any>;
}

interface ValidationResponse {
  id: string;
  name: string;
  status: string;
  overall_score: number;
  validations: ValidationResult[];
  summary: ValidationSummary;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

interface ValidationJob {
  id: string;
  request_id: string;
  status: string;
  progress: number;
  result?: ValidationResponse;
  error?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  metadata?: Record<string, any>;
}

class DataValidationService {
  private apiKey: string;
  private baseURL: string;

  constructor(apiKey: string, baseURL: string = 'https://api.kyb-platform.com/v3') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  private async makeRequest<T>(
    endpoint: string,
    method: 'GET' | 'POST' = 'GET',
    body?: any
  ): Promise<T> {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Request failed');
    }

    return response.json();
  }

  async createValidation(request: ValidationRequest): Promise<ValidationResponse> {
    return this.makeRequest<ValidationResponse>('/validation', 'POST', request);
  }

  async getValidation(id: string): Promise<ValidationResponse> {
    return this.makeRequest<ValidationResponse>(`/validation?id=${id}`);
  }

  async createValidationJob(request: ValidationRequest): Promise<ValidationJob> {
    return this.makeRequest<ValidationJob>('/validation/jobs', 'POST', request);
  }

  async getValidationJob(id: string): Promise<ValidationJob> {
    return this.makeRequest<ValidationJob>(`/validation/jobs?id=${id}`);
  }

  async waitForJobCompletion(jobId: string, pollInterval: number = 5000): Promise<ValidationJob> {
    return new Promise((resolve, reject) => {
      const checkStatus = async () => {
        try {
          const job = await this.getValidationJob(jobId);
          
          if (job.status === 'completed') {
            resolve(job);
          } else if (job.status === 'failed') {
            reject(new Error(`Job failed: ${job.error}`));
          } else {
            setTimeout(checkStatus, pollInterval);
          }
        } catch (error) {
          reject(error);
        }
      };
      
      checkStatus();
    });
  }
}

// React Hook for validation
import { useState, useCallback } from 'react';

export function useDataValidation(apiKey: string) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ValidationResponse | null>(null);

  const service = new DataValidationService(apiKey);

  const validateData = useCallback(async (request: ValidationRequest) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await service.createValidation(request);
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Validation failed');
    } finally {
      setLoading(false);
    }
  }, [service]);

  const validateDataAsync = useCallback(async (request: ValidationRequest) => {
    setLoading(true);
    setError(null);
    
    try {
      const job = await service.createValidationJob(request);
      const completedJob = await service.waitForJobCompletion(job.id);
      setResult(completedJob.result!);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Validation failed');
    } finally {
      setLoading(false);
    }
  }, [service]);

  return {
    loading,
    error,
    result,
    validateData,
    validateDataAsync,
  };
}

// React Component Example
import React, { useState } from 'react';

interface ValidationFormProps {
  apiKey: string;
}

export const ValidationForm: React.FC<ValidationFormProps> = ({ apiKey }) => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    age: '',
  });

  const { loading, error, result, validateData } = useDataValidation(apiKey);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const validationRequest: ValidationRequest = {
      name: 'Customer Data Validation',
      description: 'Validate customer form data',
      dataset: 'customer_form',
      data: formData,
      rules: [
        {
          name: 'email_format_rule',
          type: 'format',
          description: 'Validate email format',
          severity: 'high',
          expression: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
          parameters: { field: 'email' },
          enabled: true,
        },
        {
          name: 'age_business_rule',
          type: 'business',
          description: 'Validate age is 18 or older',
          severity: 'critical',
          expression: 'age >= 18',
          parameters: { field: 'age' },
          enabled: true,
        },
      ],
      options: {
        stop_on_first_error: false,
        continue_on_error: true,
        max_errors: 100,
        timeout: '30s',
        parallel: true,
      },
    };

    await validateData(validationRequest);
  };

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <div>
          <label>Name:</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />
        </div>
        <div>
          <label>Email:</label>
          <input
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
          />
        </div>
        <div>
          <label>Age:</label>
          <input
            type="number"
            value={formData.age}
            onChange={(e) => setFormData({ ...formData, age: e.target.value })}
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? 'Validating...' : 'Validate Data'}
        </button>
      </form>

      {error && (
        <div style={{ color: 'red' }}>
          Error: {error}
        </div>
      )}

      {result && (
        <div>
          <h3>Validation Results</h3>
          <p>Overall Score: {result.overall_score}</p>
          <p>Status: {result.status}</p>
          <div>
            <h4>Validations:</h4>
            {result.validations.map((validation, index) => (
              <div key={index}>
                <p>{validation.name}: {validation.status} (Score: {validation.score})</p>
                {validation.errors.length > 0 && (
                  <ul>
                    {validation.errors.map((error, errorIndex) => (
                      <li key={errorIndex} style={{ color: 'red' }}>
                        {error.message}
                      </li>
                    ))}
                  </ul>
                )}
                {validation.warnings.length > 0 && (
                  <ul>
                    {validation.warnings.map((warning, warningIndex) => (
                      <li key={warningIndex} style={{ color: 'orange' }}>
                        {warning.message}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
```

## Best Practices

### Validation Design

1. **Start with Schema Validation**: Always define a schema for your data structure
2. **Use Appropriate Severity Levels**: Critical for business-critical rules, High for important validations
3. **Implement Progressive Validation**: Start with basic format validation, then business rules
4. **Design Reusable Rules**: Create rules that can be applied across multiple datasets
5. **Consider Performance**: Use parallel processing for large datasets

### Performance Optimization

1. **Use Background Jobs**: For large datasets, use background job processing
2. **Enable Caching**: Cache validation results for repeated validations
3. **Optimize Rule Expressions**: Use efficient expressions and avoid complex computations
4. **Batch Processing**: Process data in batches for better performance
5. **Monitor Execution Times**: Track validation performance and optimize slow rules

### Error Handling

1. **Graceful Degradation**: Continue validation even if some rules fail
2. **Detailed Error Messages**: Provide clear, actionable error messages
3. **Error Categorization**: Categorize errors by severity and type
4. **Retry Logic**: Implement retry logic for transient failures
5. **Logging**: Log all validation activities for debugging

### Security

1. **Input Sanitization**: Sanitize all input data before validation
2. **Code Execution**: Be careful with custom validators that execute code
3. **Access Control**: Implement proper access control for validation resources
4. **Rate Limiting**: Implement rate limiting to prevent abuse
5. **Audit Logging**: Log all validation activities for security auditing

### Monitoring and Alerting

1. **Validation Metrics**: Track validation success rates and performance
2. **Error Monitoring**: Monitor validation errors and failures
3. **Performance Alerts**: Set up alerts for slow validations
4. **Quality Thresholds**: Set up alerts when validation scores drop below thresholds
5. **Trend Analysis**: Monitor validation trends over time

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Immediate Validations**: 100 requests per minute per API key
- **Background Jobs**: 50 job creations per minute per API key
- **Read Operations**: 1000 requests per minute per API key

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

## Monitoring

### Health Checks

Monitor the validation service health:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.kyb-platform.com/v3/health
```

### Metrics

Key metrics to monitor:

- **Validation Success Rate**: Percentage of successful validations
- **Average Validation Score**: Overall validation quality
- **Validation Duration**: Time taken for validations
- **Error Rate**: Percentage of validation errors
- **Job Completion Rate**: Percentage of completed background jobs

### Alerts

Set up alerts for:

- High error rates (>5%)
- Low validation scores (<0.8)
- Slow validations (>30 seconds)
- Failed background jobs
- Rate limit violations

## Troubleshooting

### Common Issues

1. **Validation Timeout**
   - **Cause**: Complex validation rules or large datasets
   - **Solution**: Use background jobs for large datasets, optimize rule expressions

2. **High Error Rates**
   - **Cause**: Invalid data or poorly designed validation rules
   - **Solution**: Review validation rules, check data quality, adjust severity levels

3. **Slow Performance**
   - **Cause**: Inefficient validation rules or large datasets
   - **Solution**: Optimize rules, use parallel processing, enable caching

4. **Rate Limit Exceeded**
   - **Cause**: Too many requests in a short time
   - **Solution**: Implement request throttling, use background jobs

### Debug Information

Enable debug logging by setting the log level to "debug" in validation options:

```json
{
  "options": {
    "log_level": "debug"
  }
}
```

### Support

For technical support:

- **Documentation**: [https://docs.kyb-platform.com/validation](https://docs.kyb-platform.com/validation)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: validation-support@kyb-platform.com
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

## Future Enhancements

### Planned Features

1. **Machine Learning Validation**: AI-powered validation rules
2. **Real-time Validation**: Stream validation for real-time data
3. **Validation Templates**: Pre-built validation templates for common use cases
4. **Advanced Analytics**: Deep insights into validation patterns and trends
5. **Integration Ecosystem**: Connectors for popular data platforms

### API Versioning

The API follows semantic versioning. Breaking changes will be introduced in new major versions with migration guides provided.

### Migration Guide

When upgrading to new API versions:

1. Review the changelog for breaking changes
2. Update your integration code accordingly
3. Test thoroughly in a staging environment
4. Monitor validation results after migration
5. Update documentation and training materials

---

**API Version**: 3.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
