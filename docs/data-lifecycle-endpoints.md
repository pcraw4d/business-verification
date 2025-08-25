# Data Lifecycle Management API

## Overview

The Data Lifecycle Management API provides comprehensive lifecycle management capabilities for data throughout its entire lifecycle, from creation to disposal. This API enables organizations to define, execute, and monitor data lifecycle policies with automated stages, retention management, and compliance tracking.

## Authentication

All endpoints require authentication via API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All responses are returned in JSON format with the following structure:

```json
{
  "id": "string",
  "instance": {...},
  "summary": {...},
  "statistics": {...},
  "stages": [...],
  "retention": {...},
  "timeline": {...},
  "created_at": "2024-12-19T10:00:00Z",
  "status": "string"
}
```

## Supported Lifecycle Stage Types

- `creation` - Data creation stage
- `processing` - Data processing stage
- `storage` - Data storage stage
- `archival` - Data archival stage
- `retrieval` - Data retrieval stage
- `disposal` - Data disposal stage

## Supported Lifecycle Statuses

- `active` - Stage is currently active
- `inactive` - Stage is inactive
- `suspended` - Stage is suspended
- `completed` - Stage is completed
- `failed` - Stage has failed

## Supported Retention Policy Types

- `time_based` - Time-based retention
- `event_based` - Event-based retention
- `legal_hold` - Legal hold retention
- `regulatory` - Regulatory retention
- `business` - Business retention

## Supported Data Classification Levels

- `public` - Public data
- `internal` - Internal data
- `confidential` - Confidential data
- `restricted` - Restricted data
- `secret` - Secret data

## Endpoints

### 1. Create Lifecycle Instance

**POST** `/lifecycle`

Creates and executes a data lifecycle instance immediately.

#### Request Body

```json
{
  "policy_id": "policy-1",
  "data_id": "data-1",
  "stages": [
    {
      "id": "stage-1",
      "name": "Creation",
      "type": "creation",
      "description": "Data creation stage",
      "order": 1,
      "duration": "5m",
      "conditions": [
        {
          "id": "condition-1",
          "name": "Data Validation",
          "type": "validation",
          "expression": "data.quality > 0.8",
          "parameters": {
            "threshold": 0.8
          },
          "enabled": true,
          "priority": 1
        }
      ],
      "actions": [
        {
          "id": "action-1",
          "name": "Data Validation",
          "type": "validation",
          "description": "Validate data quality",
          "parameters": {
            "quality_threshold": 0.8
          },
          "enabled": true,
          "retry_policy": {
            "max_attempts": 3,
            "initial_delay": "5s",
            "max_delay": "1m",
            "backoff_multiplier": 2.0,
            "retryable_errors": ["timeout", "network_error"]
          },
          "timeout": "1m"
        }
      ],
      "triggers": [
        {
          "id": "trigger-1",
          "name": "Scheduled Trigger",
          "type": "schedule",
          "schedule": "0 0 * * *",
          "conditions": {
            "time_window": "business_hours"
          },
          "enabled": true,
          "last_triggered": "2024-12-19T10:00:00Z",
          "next_trigger": "2024-12-20T00:00:00Z"
        }
      ],
      "status": "active",
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:00:00Z"
    }
  ],
  "retention_policies": [
    {
      "id": "retention-1",
      "name": "Data Retention",
      "description": "Data retention policy",
      "type": "time_based",
      "duration": "8760h",
      "conditions": [
        {
          "id": "condition-1",
          "name": "Business Value",
          "type": "business_value",
          "expression": "data.business_value > 0",
          "parameters": {
            "min_value": 0
          },
          "enabled": true
        }
      ],
      "actions": [
        {
          "id": "action-1",
          "name": "Archive Data",
          "type": "archive",
          "description": "Archive data to long-term storage",
          "parameters": {
            "storage_tier": "cold_storage"
          },
          "enabled": true,
          "order": 1
        }
      ],
      "exceptions": [
        {
          "id": "exception-1",
          "reason": "Legal Hold",
          "description": "Data under legal hold",
          "start_date": "2024-12-19T10:00:00Z",
          "end_date": "2025-12-19T10:00:00Z",
          "approved_by": "legal_team",
          "status": "active"
        }
      ],
      "status": "active",
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:00:00Z"
    }
  ],
  "options": {
    "auto_execute": true,
    "parallel_stages": false,
    "retry_failed": true,
    "notifications": true,
    "audit_trail": true,
    "monitoring": true,
    "validation": true
  }
}
```

#### Response

```json
{
  "id": "lifecycle_1234567890",
  "instance": {
    "id": "instance_1234567890",
    "policy_id": "policy-1",
    "data_id": "data-1",
    "status": "active",
    "current_stage": "Creation",
    "stages": [
      {
        "stage_id": "stage-1",
        "stage_name": "Creation",
        "status": "completed",
        "started_at": "2024-12-19T10:00:00Z",
        "completed_at": "2024-12-19T10:05:00Z",
        "duration": "5m",
        "actions": [
          {
            "action_id": "action-1",
            "action_name": "Data Validation",
            "status": "completed",
            "started_at": "2024-12-19T10:00:00Z",
            "completed_at": "2024-12-19T10:00:30Z",
            "duration": "30s",
            "attempts": 1,
            "error": "",
            "result": "success"
          }
        ],
        "errors": [],
        "metadata": {}
      }
    ],
    "retention": {
      "policy_id": "retention-1",
      "status": "active",
      "start_date": "2024-12-19T10:00:00Z",
      "expiry_date": "2025-12-19T10:00:00Z",
      "last_review": "2024-12-19T10:00:00Z",
      "next_review": "2025-01-19T10:00:00Z",
      "actions": [],
      "exceptions": []
    },
    "created_at": "2024-12-19T10:00:00Z",
    "updated_at": "2024-12-19T10:00:00Z",
    "completed_at": null,
    "metadata": {}
  },
  "summary": {
    "total_stages": 1,
    "completed_stages": 1,
    "active_stages": 0,
    "failed_stages": 0,
    "total_actions": 1,
    "completed_actions": 1,
    "failed_actions": 0,
    "progress": 1.0,
    "estimated_completion": "2024-12-19T11:00:00Z",
    "last_activity": "2024-12-19T10:05:00Z"
  },
  "statistics": {
    "stage_distribution": {
      "Creation": 1
    },
    "action_distribution": {
      "Data Validation": 1
    },
    "duration_stats": {
      "Data Validation": 30000
    },
    "error_stats": {},
    "performance_metrics": {
      "avg_stage_duration": 60.0,
      "success_rate": 0.95
    },
    "timeline_events": []
  },
  "stages": [
    {
      "id": "stage-1",
      "name": "Creation",
      "type": "stage",
      "status": "completed",
      "progress": 1.0,
      "started_at": "2024-12-19T10:00:00Z",
      "completed_at": "2024-12-19T10:05:00Z",
      "duration": 300000.0,
      "actions": [
        {
          "id": "action-1",
          "name": "Data Validation",
          "type": "action",
          "status": "completed",
          "started_at": "2024-12-19T10:00:00Z",
          "completed_at": "2024-12-19T10:00:30Z",
          "duration": 30000.0,
          "attempts": 1,
          "error": ""
        }
      ],
      "errors": []
    }
  ],
  "retention": {
    "policy_id": "retention-1",
    "status": "active",
    "start_date": "2024-12-19T10:00:00Z",
    "expiry_date": "2025-12-19T10:00:00Z",
    "days_remaining": 365,
    "last_review": "2024-12-19T10:00:00Z",
    "next_review": "2025-01-19T10:00:00Z",
    "actions": [],
    "exceptions": []
  },
  "timeline": {
    "start_date": "2024-12-19T10:00:00Z",
    "end_date": "2024-12-19T11:00:00Z",
    "duration": 3600.0,
    "milestones": [
      {
        "id": "milestone-1",
        "name": "Lifecycle Started",
        "description": "Data lifecycle process initiated",
        "date": "2024-12-19T10:00:00Z",
        "status": "completed",
        "type": "start"
      },
      {
        "id": "milestone-2",
        "name": "Processing Complete",
        "description": "Data processing stage completed",
        "date": "2024-12-19T10:05:00Z",
        "status": "completed",
        "type": "processing"
      }
    ],
    "events": [
      {
        "id": "event-1",
        "type": "stage_started",
        "stage": "creation",
        "action": "data_creation",
        "status": "completed",
        "timestamp": "2024-12-19T10:00:00Z",
        "duration": 60.0,
        "description": "Data creation stage started"
      }
    ],
    "projections": [
      {
        "type": "completion",
        "date": "2024-12-19T11:00:00Z",
        "confidence": 0.95,
        "description": "Expected completion time"
      }
    ]
  },
  "created_at": "2024-12-19T10:00:00Z",
  "status": "completed"
}
```

### 2. Get Lifecycle Instance

**GET** `/lifecycle?id={id}`

Retrieves a specific lifecycle instance.

#### Response

Same structure as Create Lifecycle Instance response.

### 3. List Lifecycle Instances

**GET** `/lifecycle`

Lists all lifecycle instances.

#### Response

```json
{
  "instances": [
    {
      "id": "instance-1",
      "policy_id": "policy-1",
      "data_id": "data-1",
      "status": "active",
      "current_stage": "Processing"
    }
  ],
  "total": 1,
  "timestamp": "2024-12-19T10:00:00Z"
}
```

### 4. Create Lifecycle Job

**POST** `/lifecycle/jobs`

Creates a background lifecycle execution job.

#### Request Body

Same as Create Lifecycle Instance.

#### Response

```json
{
  "job_id": "job_1234567890",
  "status": "created",
  "created_at": "2024-12-19T10:00:00Z"
}
```

### 5. Get Lifecycle Job

**GET** `/lifecycle/jobs?id={id}`

Retrieves job status and results.

#### Response

```json
{
  "id": "job_1234567890",
  "type": "lifecycle_execution",
  "status": "completed",
  "progress": 1.0,
  "created_at": "2024-12-19T10:00:00Z",
  "started_at": "2024-12-19T10:00:01Z",
  "completed_at": "2024-12-19T10:00:06Z",
  "result": {
    "instance_id": "instance_1234567890",
    "summary": {...},
    "stages": [...],
    "retention": {...},
    "timeline": {...},
    "statistics": {...},
    "generated_at": "2024-12-19T10:00:06Z"
  }
}
```

### 6. List Lifecycle Jobs

**GET** `/lifecycle/jobs`

Lists all lifecycle jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "job_1234567890",
      "type": "lifecycle_execution",
      "status": "completed",
      "progress": 1.0,
      "created_at": "2024-12-19T10:00:00Z"
    }
  ],
  "total": 1,
  "timestamp": "2024-12-19T10:00:00Z"
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Validation error: policy ID is required"
}
```

### 401 Unauthorized

```json
{
  "error": "Invalid API key"
}
```

### 404 Not Found

```json
{
  "error": "Instance not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal server error"
}
```

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

// Create lifecycle instance
async function createLifecycleInstance() {
  try {
    const response = await axios.post('https://api.example.com/lifecycle', {
      policy_id: 'policy-1',
      data_id: 'data-1',
      stages: [{
        id: 'stage-1',
        name: 'Creation',
        type: 'creation',
        description: 'Data creation stage',
        order: 1,
        duration: '5m',
        conditions: [],
        actions: [{
          id: 'action-1',
          name: 'Data Validation',
          type: 'validation',
          description: 'Validate data quality',
          parameters: {
            quality_threshold: 0.8
          },
          enabled: true,
          retry_policy: {
            max_attempts: 3,
            initial_delay: '5s',
            max_delay: '1m',
            backoff_multiplier: 2.0,
            retryable_errors: ['timeout', 'network_error']
          },
          timeout: '1m'
        }],
        triggers: [],
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }],
      retention_policies: [{
        id: 'retention-1',
        name: 'Data Retention',
        description: 'Data retention policy',
        type: 'time_based',
        duration: '8760h',
        conditions: [],
        actions: [],
        exceptions: [],
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }],
      options: {
        auto_execute: true,
        parallel_stages: false,
        retry_failed: true,
        notifications: true,
        audit_trail: true,
        monitoring: true,
        validation: true
      }
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });

    console.log('Lifecycle instance created:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error creating lifecycle instance:', error.response?.data || error.message);
    throw error;
  }
}

// Get lifecycle instance
async function getLifecycleInstance(instanceId) {
  try {
    const response = await axios.get(`https://api.example.com/lifecycle?id=${instanceId}`, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY'
      }
    });

    console.log('Lifecycle instance retrieved:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error retrieving lifecycle instance:', error.response?.data || error.message);
    throw error;
  }
}

// Create background job
async function createLifecycleJob() {
  try {
    const response = await axios.post('https://api.example.com/lifecycle/jobs', {
      policy_id: 'policy-1',
      data_id: 'data-1',
      stages: [/* ... */],
      retention_policies: [/* ... */],
      options: {/* ... */}
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });

    console.log('Lifecycle job created:', response.data);
    return response.data.job_id;
  } catch (error) {
    console.error('Error creating lifecycle job:', error.response?.data || error.message);
    throw error;
  }
}

// Monitor job progress
async function monitorJobProgress(jobId) {
  try {
    const response = await axios.get(`https://api.example.com/lifecycle/jobs?id=${jobId}`, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY'
      }
    });

    const job = response.data;
    console.log(`Job ${jobId} status: ${job.status}, progress: ${job.progress * 100}%`);

    if (job.status === 'completed') {
      console.log('Job completed successfully!');
      console.log('Results:', job.result);
    } else if (job.status === 'failed') {
      console.error('Job failed:', job.error);
    }

    return job;
  } catch (error) {
    console.error('Error monitoring job:', error.response?.data || error.message);
    throw error;
  }
}
```

### Python

```python
import requests
import json
from datetime import datetime, timedelta

class DataLifecycleClient:
    def __init__(self, api_key, base_url="https://api.example.com"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def create_lifecycle_instance(self):
        """Create a lifecycle instance"""
        url = f"{self.base_url}/lifecycle"
        
        payload = {
            "policy_id": "policy-1",
            "data_id": "data-1",
            "stages": [{
                "id": "stage-1",
                "name": "Creation",
                "type": "creation",
                "description": "Data creation stage",
                "order": 1,
                "duration": "5m",
                "conditions": [],
                "actions": [{
                    "id": "action-1",
                    "name": "Data Validation",
                    "type": "validation",
                    "description": "Validate data quality",
                    "parameters": {
                        "quality_threshold": 0.8
                    },
                    "enabled": True,
                    "retry_policy": {
                        "max_attempts": 3,
                        "initial_delay": "5s",
                        "max_delay": "1m",
                        "backoff_multiplier": 2.0,
                        "retryable_errors": ["timeout", "network_error"]
                    },
                    "timeout": "1m"
                }],
                "triggers": [],
                "status": "active",
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat()
            }],
            "retention_policies": [{
                "id": "retention-1",
                "name": "Data Retention",
                "description": "Data retention policy",
                "type": "time_based",
                "duration": "8760h",
                "conditions": [],
                "actions": [],
                "exceptions": [],
                "status": "active",
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat()
            }],
            "options": {
                "auto_execute": True,
                "parallel_stages": False,
                "retry_failed": True,
                "notifications": True,
                "audit_trail": True,
                "monitoring": True,
                "validation": True
            }
        }

        response = requests.post(url, headers=self.headers, json=payload)
        response.raise_for_status()
        return response.json()

    def get_lifecycle_instance(self, instance_id):
        """Get a lifecycle instance"""
        url = f"{self.base_url}/lifecycle?id={instance_id}"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def list_lifecycle_instances(self):
        """List all lifecycle instances"""
        url = f"{self.base_url}/lifecycle"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def create_lifecycle_job(self, lifecycle_request):
        """Create a background lifecycle job"""
        url = f"{self.base_url}/lifecycle/jobs"
        response = requests.post(url, headers=self.headers, json=lifecycle_request)
        response.raise_for_status()
        return response.json()

    def get_lifecycle_job(self, job_id):
        """Get job status and results"""
        url = f"{self.base_url}/lifecycle/jobs?id={job_id}"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def list_lifecycle_jobs(self):
        """List all lifecycle jobs"""
        url = f"{self.base_url}/lifecycle/jobs"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

# Usage example
def main():
    client = DataLifecycleClient("YOUR_API_KEY")
    
    try:
        # Create lifecycle instance
        instance = client.create_lifecycle_instance()
        print(f"Created instance: {instance['id']}")
        
        # Create background job
        job_response = client.create_lifecycle_job(instance)
        job_id = job_response['job_id']
        print(f"Created job: {job_id}")
        
        # Monitor job progress
        import time
        while True:
            job = client.get_lifecycle_job(job_id)
            print(f"Job status: {job['status']}, progress: {job['progress'] * 100:.1f}%")
            
            if job['status'] in ['completed', 'failed']:
                if job['status'] == 'completed':
                    print("Job completed successfully!")
                    print(f"Results: {job['result']}")
                else:
                    print(f"Job failed: {job['error']}")
                break
            
            time.sleep(2)
            
    except requests.exceptions.RequestException as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
```

### React/TypeScript

```typescript
import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface LifecycleInstance {
  id: string;
  policy_id: string;
  data_id: string;
  status: string;
  current_stage: string;
}

interface LifecycleJob {
  id: string;
  type: string;
  status: string;
  progress: number;
  created_at: string;
  result?: any;
  error?: string;
}

interface LifecycleClientProps {
  apiKey: string;
  baseUrl?: string;
}

const LifecycleClient: React.FC<LifecycleClientProps> = ({ 
  apiKey, 
  baseUrl = 'https://api.example.com' 
}) => {
  const [instances, setInstances] = useState<LifecycleInstance[]>([]);
  const [jobs, setJobs] = useState<LifecycleJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const headers = {
    'Authorization': `Bearer ${apiKey}`,
    'Content-Type': 'application/json'
  };

  const createLifecycleInstance = async () => {
    setLoading(true);
    setError(null);

    try {
      const payload = {
        policy_id: 'policy-1',
        data_id: 'data-1',
        stages: [{
          id: 'stage-1',
          name: 'Creation',
          type: 'creation',
          description: 'Data creation stage',
          order: 1,
          duration: '5m',
          conditions: [],
          actions: [{
            id: 'action-1',
            name: 'Data Validation',
            type: 'validation',
            description: 'Validate data quality',
            parameters: {
              quality_threshold: 0.8
            },
            enabled: true,
            retry_policy: {
              max_attempts: 3,
              initial_delay: '5s',
              max_delay: '1m',
              backoff_multiplier: 2.0,
              retryable_errors: ['timeout', 'network_error']
            },
            timeout: '1m'
          }],
          triggers: [],
          status: 'active',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }],
        retention_policies: [{
          id: 'retention-1',
          name: 'Data Retention',
          description: 'Data retention policy',
          type: 'time_based',
          duration: '8760h',
          conditions: [],
          actions: [],
          exceptions: [],
          status: 'active',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }],
        options: {
          auto_execute: true,
          parallel_stages: false,
          retry_failed: true,
          notifications: true,
          audit_trail: true,
          monitoring: true,
          validation: true
        }
      };

      const response = await axios.post(`${baseUrl}/lifecycle`, payload, { headers });
      console.log('Lifecycle instance created:', response.data);
      
      // Refresh instances list
      loadInstances();
      
      return response.data;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error creating lifecycle instance:', errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loadInstances = async () => {
    try {
      const response = await axios.get(`${baseUrl}/lifecycle`, { headers });
      setInstances(response.data.instances);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error loading instances:', errorMessage);
    }
  };

  const createLifecycleJob = async () => {
    setLoading(true);
    setError(null);

    try {
      const payload = {
        policy_id: 'policy-1',
        data_id: 'data-1',
        stages: [/* ... */],
        retention_policies: [/* ... */],
        options: {/* ... */}
      };

      const response = await axios.post(`${baseUrl}/lifecycle/jobs`, payload, { headers });
      console.log('Lifecycle job created:', response.data);
      
      // Refresh jobs list
      loadJobs();
      
      return response.data.job_id;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error creating lifecycle job:', errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loadJobs = async () => {
    try {
      const response = await axios.get(`${baseUrl}/lifecycle/jobs`, { headers });
      setJobs(response.data.jobs);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error loading jobs:', errorMessage);
    }
  };

  const monitorJobProgress = async (jobId: string) => {
    try {
      const response = await axios.get(`${baseUrl}/lifecycle/jobs?id=${jobId}`, { headers });
      const job = response.data;
      
      console.log(`Job ${jobId} status: ${job.status}, progress: ${job.progress * 100}%`);
      
      if (job.status === 'completed') {
        console.log('Job completed successfully!');
        console.log('Results:', job.result);
      } else if (job.status === 'failed') {
        console.error('Job failed:', job.error);
      }
      
      return job;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      console.error('Error monitoring job:', errorMessage);
      throw err;
    }
  };

  useEffect(() => {
    loadInstances();
    loadJobs();
  }, []);

  return (
    <div className="lifecycle-client">
      <h2>Data Lifecycle Management</h2>
      
      {error && (
        <div className="error">
          Error: {error}
        </div>
      )}
      
      <div className="actions">
        <button 
          onClick={createLifecycleInstance} 
          disabled={loading}
        >
          {loading ? 'Creating...' : 'Create Instance'}
        </button>
        
        <button 
          onClick={createLifecycleJob} 
          disabled={loading}
        >
          {loading ? 'Creating...' : 'Create Job'}
        </button>
      </div>
      
      <div className="instances">
        <h3>Lifecycle Instances ({instances.length})</h3>
        <div className="instance-list">
          {instances.map(instance => (
            <div key={instance.id} className="instance-item">
              <h4>Instance {instance.id}</h4>
              <p>Policy: {instance.policy_id}</p>
              <p>Data: {instance.data_id}</p>
              <p>Status: {instance.status}</p>
              <p>Current Stage: {instance.current_stage}</p>
            </div>
          ))}
        </div>
      </div>
      
      <div className="jobs">
        <h3>Lifecycle Jobs ({jobs.length})</h3>
        <div className="job-list">
          {jobs.map(job => (
            <div key={job.id} className="job-item">
              <h4>Job {job.id}</h4>
              <p>Type: {job.type}</p>
              <p>Status: {job.status}</p>
              <p>Progress: {(job.progress * 100).toFixed(1)}%</p>
              <p>Created: {new Date(job.created_at).toLocaleString()}</p>
              {job.error && <p className="error">Error: {job.error}</p>}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default LifecycleClient;
```

## Best Practices

### Lifecycle Design
- **Define Clear Stages**: Create well-defined stages with clear entry and exit criteria
- **Stage Dependencies**: Establish proper dependencies between stages
- **Error Handling**: Implement comprehensive error handling and recovery mechanisms
- **Monitoring**: Set up monitoring and alerting for each stage

### Retention Management
- **Policy Design**: Design retention policies based on business and regulatory requirements
- **Exception Handling**: Implement proper exception handling for legal holds and special cases
- **Review Process**: Establish regular review processes for retention policies
- **Compliance Tracking**: Track compliance with retention requirements

### Performance Optimization
- **Parallel Processing**: Use parallel stages where possible to improve performance
- **Resource Management**: Optimize resource usage during lifecycle execution
- **Caching**: Implement caching for frequently accessed data
- **Batch Processing**: Use batch processing for large datasets

### Security and Compliance
- **Data Classification**: Properly classify data based on sensitivity
- **Access Control**: Implement proper access controls for lifecycle operations
- **Audit Trail**: Maintain comprehensive audit trails for all operations
- **Compliance Monitoring**: Monitor compliance with data lifecycle requirements

## Rate Limiting

- **Requests per minute**: 100
- **Requests per hour**: 1000
- **Concurrent jobs**: 10

## Monitoring

### Key Metrics
- Lifecycle instance creation rate
- Stage completion time
- Job success rate
- Retention policy compliance
- Error rates by stage

### Alerts
- Stage failures
- Job timeouts
- Retention policy violations
- Compliance breaches

## Troubleshooting

### Common Issues

**Stage Failures**
- Check stage conditions and parameters
- Verify action configurations
- Review error logs for specific failure reasons

**Job Timeouts**
- Increase timeout values for long-running actions
- Optimize action performance
- Consider breaking large actions into smaller ones

**Retention Policy Issues**
- Verify retention policy configurations
- Check exception handling
- Review compliance requirements

### Debug Information

Enable debug logging by setting the `X-Debug` header:

```
X-Debug: true
```

### Support

For technical support:
- Email: support@example.com
- Documentation: https://docs.example.com
- Status page: https://status.example.com

## Future Enhancements

- **Advanced Analytics**: Enhanced lifecycle analytics and reporting
- **Integration APIs**: Integration with third-party lifecycle tools
- **Automated Optimization**: AI-powered lifecycle optimization
- **Real-time Monitoring**: Real-time lifecycle monitoring
- **Advanced Scheduling**: Sophisticated scheduling capabilities
