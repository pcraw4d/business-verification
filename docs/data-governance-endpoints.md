# Data Governance Framework API

## Overview

The Data Governance Framework API provides comprehensive governance capabilities for managing data policies, controls, compliance, and risk assessment. This API enables organizations to establish, monitor, and maintain robust data governance frameworks.

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
  "framework": {...},
  "summary": {...},
  "statistics": {...},
  "compliance": {...},
  "risk_assessment": {...},
  "controls": [...],
  "policies": [...],
  "created_at": "2024-12-19T10:00:00Z",
  "status": "string"
}
```

## Supported Framework Types

- `data_quality` - Data quality governance
- `data_privacy` - Data privacy governance  
- `data_security` - Data security governance
- `data_compliance` - Data compliance governance
- `data_retention` - Data retention governance
- `data_lineage` - Data lineage governance

## Supported Framework Statuses

- `draft` - Framework in draft state
- `active` - Framework is active
- `suspended` - Framework is suspended
- `deprecated` - Framework is deprecated
- `archived` - Framework is archived

## Supported Control Types

- `preventive` - Preventive controls
- `detective` - Detective controls
- `corrective` - Corrective controls
- `compensating` - Compensating controls
- `directive` - Directive controls

## Supported Compliance Standards

- `gdpr` - General Data Protection Regulation
- `ccpa` - California Consumer Privacy Act
- `sox` - Sarbanes-Oxley Act
- `hipaa` - Health Insurance Portability and Accountability Act
- `pci` - Payment Card Industry Data Security Standard
- `iso27001` - ISO 27001 Information Security

## Supported Risk Levels

- `low` - Low risk
- `medium` - Medium risk
- `high` - High risk
- `critical` - Critical risk

## Endpoints

### 1. Create Governance Framework

**POST** `/governance`

Creates and executes a governance framework immediately.

#### Request Body

```json
{
  "framework_type": "data_quality",
  "policies": [
    {
      "id": "policy-1",
      "name": "Data Quality Policy",
      "description": "Ensures high data quality standards",
      "category": "Quality",
      "version": "1.0",
      "status": "active",
      "owner": "Data Team",
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:00:00Z",
      "rules": [
        {
          "id": "rule-1",
          "name": "Data Validation Rule",
          "description": "Validates data completeness",
          "type": "validation",
          "condition": "data.completeness > 0.95",
          "action": "flag_incomplete",
          "priority": 1,
          "enabled": true,
          "parameters": {
            "threshold": 0.95
          }
        }
      ],
      "compliance": ["gdpr"],
      "risk_level": "low",
      "tags": ["quality", "compliance"],
      "metadata": {}
    }
  ],
  "controls": [
    {
      "id": "control-1",
      "name": "Data Validation Control",
      "description": "Validates data quality",
      "type": "preventive",
      "category": "Quality",
      "status": "active",
      "priority": 1,
      "effectiveness": 0.90,
      "implementation": {
        "status": "implemented",
        "start_date": "2024-12-19T10:00:00Z",
        "end_date": "2024-12-19T10:00:00Z",
        "owner": "Data Team",
        "resources": ["Data Engineers"],
        "cost": 50000.0,
        "timeline": "2 months",
        "milestones": []
      },
      "monitoring": {
        "enabled": true,
        "frequency": "daily",
        "metrics": ["validation_rate", "error_rate"],
        "thresholds": {
          "error_rate": 0.05
        },
        "alerts": [],
        "reports": ["daily", "weekly"]
      },
      "testing": {
        "enabled": true,
        "frequency": "weekly",
        "method": "automated",
        "scope": "all data",
        "test_cases": [],
        "results": []
      },
      "documentation": "Data validation control documentation",
      "owner": "Data Team",
      "created_at": "2024-12-19T10:00:00Z",
      "updated_at": "2024-12-19T10:00:00Z"
    }
  ],
  "compliance": [
    {
      "id": "comp-1",
      "standard": "gdpr",
      "requirement": "Data Protection",
      "description": "Protect personal data",
      "category": "Privacy",
      "priority": 1,
      "status": "compliant",
      "controls": ["control-1"],
      "evidence": [],
      "due_date": "2024-12-19T10:00:00Z",
      "owner": "Compliance Team"
    }
  ],
  "risk_profile": {
    "overall_risk": "medium",
    "categories": [
      {
        "name": "Data Privacy",
        "description": "Privacy-related risks",
        "risk_level": "low",
        "probability": 0.2,
        "impact": 0.3,
        "score": 0.06
      }
    ],
    "mitigations": [],
    "assessments": [],
    "updated_at": "2024-12-19T10:00:00Z"
  },
  "scope": {
    "data_domains": ["customer", "product"],
    "business_units": ["sales", "marketing"],
    "systems": ["crm", "erp"],
    "processes": ["data_ingestion", "data_processing"],
    "geographies": ["US", "EU"],
    "timeframe": "ongoing",
    "exceptions": []
  },
  "options": {
    "auto_assessment": true,
    "risk_scoring": true,
    "compliance_check": true,
    "control_testing": true,
    "reporting": true,
    "notifications": true,
    "audit_trail": true,
    "version_control": true
  }
}
```

#### Response

```json
{
  "id": "gov_1234567890",
  "framework": {
    "id": "framework_1234567890",
    "name": "Data Quality Governance Framework",
    "description": "Comprehensive governance framework for data management",
    "type": "data_quality",
    "status": "active",
    "version": "1.0.0",
    "owner": "Data Governance Team",
    "created_at": "2024-12-19T10:00:00Z",
    "updated_at": "2024-12-19T10:00:00Z",
    "policies": [...],
    "controls": [...],
    "compliance": [...],
    "risk_profile": {...},
    "scope": {...},
    "metadata": {}
  },
  "summary": {
    "total_policies": 1,
    "active_policies": 1,
    "total_controls": 1,
    "effective_controls": 1,
    "compliance_score": 0.85,
    "risk_score": 0.25,
    "coverage": 0.90,
    "last_assessment": "2024-12-19T10:00:00Z"
  },
  "statistics": {
    "policy_distribution": {
      "data_quality": 1
    },
    "control_effectiveness": {
      "preventive": 0.90
    },
    "compliance_trends": {
      "gdpr": 0.95
    },
    "risk_distribution": {
      "low": 1
    },
    "assessment_history": [
      {
        "date": "2024-11-19T10:00:00Z",
        "score": 0.82,
        "risk_level": "medium",
        "assessor": "Governance Team",
        "notes": "Monthly assessment"
      }
    ]
  },
  "compliance": {
    "overall_score": 0.85,
    "standards": {
      "gdpr": 0.95
    },
    "requirements": [
      {
        "id": "req-1",
        "standard": "GDPR",
        "requirement": "Data Protection",
        "status": "compliant",
        "score": 0.95,
        "last_check": "2024-12-19T10:00:00Z",
        "next_check": "2025-01-19T10:00:00Z"
      }
    ],
    "violations": [],
    "last_audit": "2024-11-19T10:00:00Z",
    "next_audit": "2025-01-19T10:00:00Z"
  },
  "risk_assessment": {
    "overall_risk": "medium",
    "risk_score": 0.25,
    "categories": [
      {
        "name": "Data Privacy",
        "description": "Privacy-related risks",
        "risk_level": "low",
        "probability": 0.2,
        "impact": 0.3,
        "score": 0.06
      }
    ],
    "top_risks": [
      {
        "id": "risk-1",
        "name": "Data Breach",
        "category": "Security",
        "risk_level": "medium",
        "probability": 0.3,
        "impact": 0.7,
        "score": 0.21,
        "status": "mitigated"
      }
    ],
    "mitigations": [],
    "trends": [],
    "last_updated": "2024-12-19T10:00:00Z"
  },
  "controls": [
    {
      "id": "control-1",
      "name": "Data Validation Control",
      "type": "preventive",
      "status": "active",
      "effectiveness": 0.90,
      "last_tested": "2024-11-19T10:00:00Z",
      "next_test": "2025-01-19T10:00:00Z",
      "issues": []
    }
  ],
  "policies": [
    {
      "id": "policy-1",
      "name": "Data Quality Policy",
      "status": "active",
      "compliance": 0.90,
      "last_review": "2024-11-19T10:00:00Z",
      "next_review": "2025-01-19T10:00:00Z",
      "violations": 0,
      "exceptions": 0
    }
  ],
  "created_at": "2024-12-19T10:00:00Z",
  "status": "completed"
}
```

### 2. Get Governance Framework

**GET** `/governance?id={id}`

Retrieves a specific governance framework.

#### Response

Same structure as Create Governance Framework response.

### 3. List Governance Frameworks

**GET** `/governance`

Lists all governance frameworks.

#### Response

```json
{
  "frameworks": [
    {
      "id": "framework-1",
      "name": "Data Quality Framework",
      "type": "data_quality",
      "status": "active",
      "version": "1.0.0"
    }
  ],
  "total": 1,
  "timestamp": "2024-12-19T10:00:00Z"
}
```

### 4. Create Governance Job

**POST** `/governance/jobs`

Creates a background governance assessment job.

#### Request Body

Same as Create Governance Framework.

#### Response

```json
{
  "job_id": "job_1234567890",
  "status": "created",
  "created_at": "2024-12-19T10:00:00Z"
}
```

### 5. Get Governance Job

**GET** `/governance/jobs?id={id}`

Retrieves job status and results.

#### Response

```json
{
  "id": "job_1234567890",
  "type": "governance_assessment",
  "status": "completed",
  "progress": 1.0,
  "created_at": "2024-12-19T10:00:00Z",
  "started_at": "2024-12-19T10:00:01Z",
  "completed_at": "2024-12-19T10:00:06Z",
  "result": {
    "framework_id": "framework_1234567890",
    "summary": {...},
    "compliance": {...},
    "risk_assessment": {...},
    "controls": [...],
    "policies": [...],
    "statistics": {...},
    "generated_at": "2024-12-19T10:00:06Z"
  }
}
```

### 6. List Governance Jobs

**GET** `/governance/jobs`

Lists all governance jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "job_1234567890",
      "type": "governance_assessment",
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
  "error": "Validation error: framework type is required"
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
  "error": "Framework not found"
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

// Create governance framework
async function createGovernanceFramework() {
  try {
    const response = await axios.post('https://api.example.com/governance', {
      framework_type: 'data_quality',
      policies: [{
        id: 'policy-1',
        name: 'Data Quality Policy',
        description: 'Ensures data quality',
        category: 'Quality',
        version: '1.0',
        status: 'active',
        owner: 'Data Team',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        rules: [],
        compliance: ['gdpr'],
        risk_level: 'low',
        tags: ['quality'],
        metadata: {}
      }],
      controls: [{
        id: 'control-1',
        name: 'Data Validation',
        description: 'Validates data',
        type: 'preventive',
        category: 'Quality',
        status: 'active',
        priority: 1,
        effectiveness: 0.90,
        implementation: {
          status: 'implemented',
          start_date: new Date().toISOString(),
          end_date: new Date().toISOString(),
          owner: 'Data Team',
          resources: ['Engineers'],
          cost: 50000.0,
          timeline: '2 months',
          milestones: []
        },
        monitoring: {
          enabled: true,
          frequency: 'daily',
          metrics: ['validation_rate'],
          thresholds: { error_rate: 0.05 },
          alerts: [],
          reports: ['daily']
        },
        testing: {
          enabled: true,
          frequency: 'weekly',
          method: 'automated',
          scope: 'all data',
          test_cases: [],
          results: []
        },
        documentation: 'Control documentation',
        owner: 'Data Team',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }],
      compliance: [{
        id: 'comp-1',
        standard: 'gdpr',
        requirement: 'Data Protection',
        description: 'Protect data',
        category: 'Privacy',
        priority: 1,
        status: 'compliant',
        controls: ['control-1'],
        evidence: [],
        due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
        owner: 'Compliance Team'
      }],
      risk_profile: {
        overall_risk: 'medium',
        categories: [{
          name: 'Data Privacy',
          description: 'Privacy risks',
          risk_level: 'low',
          probability: 0.2,
          impact: 0.3,
          score: 0.06
        }],
        mitigations: [],
        assessments: [],
        updated_at: new Date().toISOString()
      },
      scope: {
        data_domains: ['customer'],
        business_units: ['sales'],
        systems: ['crm'],
        processes: ['data_ingestion'],
        geographies: ['US'],
        timeframe: 'ongoing',
        exceptions: []
      },
      options: {
        auto_assessment: true,
        risk_scoring: true,
        compliance_check: true,
        control_testing: true,
        reporting: true,
        notifications: true,
        audit_trail: true,
        version_control: true
      }
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });

    console.log('Governance framework created:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error creating governance framework:', error.response?.data || error.message);
    throw error;
  }
}

// Get governance framework
async function getGovernanceFramework(frameworkId) {
  try {
    const response = await axios.get(`https://api.example.com/governance?id=${frameworkId}`, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY'
      }
    });

    console.log('Governance framework retrieved:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error retrieving governance framework:', error.response?.data || error.message);
    throw error;
  }
}

// Create background job
async function createGovernanceJob() {
  try {
    const response = await axios.post('https://api.example.com/governance/jobs', {
      framework_type: 'data_quality',
      policies: [/* ... */],
      controls: [/* ... */],
      compliance: [/* ... */],
      risk_profile: {/* ... */},
      scope: {/* ... */},
      options: {/* ... */}
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });

    console.log('Governance job created:', response.data);
    return response.data.job_id;
  } catch (error) {
    console.error('Error creating governance job:', error.response?.data || error.message);
    throw error;
  }
}

// Monitor job progress
async function monitorJobProgress(jobId) {
  try {
    const response = await axios.get(`https://api.example.com/governance/jobs?id=${jobId}`, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY'
      }
    });

    const job = response.data;
    console.log(`Job ${jobId} status: ${job.status}, progress: ${job.progress * 100}%`);

    if (job.status === 'completed') {
      console.log('Job completed with results:', job.result);
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

class DataGovernanceClient:
    def __init__(self, api_key, base_url="https://api.example.com"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def create_governance_framework(self):
        """Create a governance framework"""
        url = f"{self.base_url}/governance"
        
        payload = {
            "framework_type": "data_quality",
            "policies": [{
                "id": "policy-1",
                "name": "Data Quality Policy",
                "description": "Ensures data quality",
                "category": "Quality",
                "version": "1.0",
                "status": "active",
                "owner": "Data Team",
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat(),
                "rules": [],
                "compliance": ["gdpr"],
                "risk_level": "low",
                "tags": ["quality"],
                "metadata": {}
            }],
            "controls": [{
                "id": "control-1",
                "name": "Data Validation",
                "description": "Validates data",
                "type": "preventive",
                "category": "Quality",
                "status": "active",
                "priority": 1,
                "effectiveness": 0.90,
                "implementation": {
                    "status": "implemented",
                    "start_date": datetime.now().isoformat(),
                    "end_date": datetime.now().isoformat(),
                    "owner": "Data Team",
                    "resources": ["Engineers"],
                    "cost": 50000.0,
                    "timeline": "2 months",
                    "milestones": []
                },
                "monitoring": {
                    "enabled": True,
                    "frequency": "daily",
                    "metrics": ["validation_rate"],
                    "thresholds": {"error_rate": 0.05},
                    "alerts": [],
                    "reports": ["daily"]
                },
                "testing": {
                    "enabled": True,
                    "frequency": "weekly",
                    "method": "automated",
                    "scope": "all data",
                    "test_cases": [],
                    "results": []
                },
                "documentation": "Control documentation",
                "owner": "Data Team",
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat()
            }],
            "compliance": [{
                "id": "comp-1",
                "standard": "gdpr",
                "requirement": "Data Protection",
                "description": "Protect data",
                "category": "Privacy",
                "priority": 1,
                "status": "compliant",
                "controls": ["control-1"],
                "evidence": [],
                "due_date": (datetime.now() + timedelta(days=30)).isoformat(),
                "owner": "Compliance Team"
            }],
            "risk_profile": {
                "overall_risk": "medium",
                "categories": [{
                    "name": "Data Privacy",
                    "description": "Privacy risks",
                    "risk_level": "low",
                    "probability": 0.2,
                    "impact": 0.3,
                    "score": 0.06
                }],
                "mitigations": [],
                "assessments": [],
                "updated_at": datetime.now().isoformat()
            },
            "scope": {
                "data_domains": ["customer"],
                "business_units": ["sales"],
                "systems": ["crm"],
                "processes": ["data_ingestion"],
                "geographies": ["US"],
                "timeframe": "ongoing",
                "exceptions": []
            },
            "options": {
                "auto_assessment": True,
                "risk_scoring": True,
                "compliance_check": True,
                "control_testing": True,
                "reporting": True,
                "notifications": True,
                "audit_trail": True,
                "version_control": True
            }
        }

        response = requests.post(url, headers=self.headers, json=payload)
        response.raise_for_status()
        return response.json()

    def get_governance_framework(self, framework_id):
        """Get a governance framework"""
        url = f"{self.base_url}/governance?id={framework_id}"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def list_governance_frameworks(self):
        """List all governance frameworks"""
        url = f"{self.base_url}/governance"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def create_governance_job(self, framework_request):
        """Create a background governance job"""
        url = f"{self.base_url}/governance/jobs"
        response = requests.post(url, headers=self.headers, json=framework_request)
        response.raise_for_status()
        return response.json()

    def get_governance_job(self, job_id):
        """Get job status and results"""
        url = f"{self.base_url}/governance/jobs?id={job_id}"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def list_governance_jobs(self):
        """List all governance jobs"""
        url = f"{self.base_url}/governance/jobs"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

# Usage example
def main():
    client = DataGovernanceClient("YOUR_API_KEY")
    
    try:
        # Create governance framework
        framework = client.create_governance_framework()
        print(f"Created framework: {framework['id']}")
        
        # Create background job
        job_response = client.create_governance_job(framework)
        job_id = job_response['job_id']
        print(f"Created job: {job_id}")
        
        # Monitor job progress
        import time
        while True:
            job = client.get_governance_job(job_id)
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

interface GovernanceFramework {
  id: string;
  name: string;
  type: string;
  status: string;
  version: string;
}

interface GovernanceJob {
  id: string;
  type: string;
  status: string;
  progress: number;
  created_at: string;
  result?: any;
  error?: string;
}

interface GovernanceClientProps {
  apiKey: string;
  baseUrl?: string;
}

const GovernanceClient: React.FC<GovernanceClientProps> = ({ 
  apiKey, 
  baseUrl = 'https://api.example.com' 
}) => {
  const [frameworks, setFrameworks] = useState<GovernanceFramework[]>([]);
  const [jobs, setJobs] = useState<GovernanceJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const headers = {
    'Authorization': `Bearer ${apiKey}`,
    'Content-Type': 'application/json'
  };

  const createGovernanceFramework = async () => {
    setLoading(true);
    setError(null);

    try {
      const payload = {
        framework_type: 'data_quality',
        policies: [{
          id: 'policy-1',
          name: 'Data Quality Policy',
          description: 'Ensures data quality',
          category: 'Quality',
          version: '1.0',
          status: 'active',
          owner: 'Data Team',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          rules: [],
          compliance: ['gdpr'],
          risk_level: 'low',
          tags: ['quality'],
          metadata: {}
        }],
        controls: [{
          id: 'control-1',
          name: 'Data Validation',
          description: 'Validates data',
          type: 'preventive',
          category: 'Quality',
          status: 'active',
          priority: 1,
          effectiveness: 0.90,
          implementation: {
            status: 'implemented',
            start_date: new Date().toISOString(),
            end_date: new Date().toISOString(),
            owner: 'Data Team',
            resources: ['Engineers'],
            cost: 50000.0,
            timeline: '2 months',
            milestones: []
          },
          monitoring: {
            enabled: true,
            frequency: 'daily',
            metrics: ['validation_rate'],
            thresholds: { error_rate: 0.05 },
            alerts: [],
            reports: ['daily']
          },
          testing: {
            enabled: true,
            frequency: 'weekly',
            method: 'automated',
            scope: 'all data',
            test_cases: [],
            results: []
          },
          documentation: 'Control documentation',
          owner: 'Data Team',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }],
        compliance: [{
          id: 'comp-1',
          standard: 'gdpr',
          requirement: 'Data Protection',
          description: 'Protect data',
          category: 'Privacy',
          priority: 1,
          status: 'compliant',
          controls: ['control-1'],
          evidence: [],
          due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
          owner: 'Compliance Team'
        }],
        risk_profile: {
          overall_risk: 'medium',
          categories: [{
            name: 'Data Privacy',
            description: 'Privacy risks',
            risk_level: 'low',
            probability: 0.2,
            impact: 0.3,
            score: 0.06
          }],
          mitigations: [],
          assessments: [],
          updated_at: new Date().toISOString()
        },
        scope: {
          data_domains: ['customer'],
          business_units: ['sales'],
          systems: ['crm'],
          processes: ['data_ingestion'],
          geographies: ['US'],
          timeframe: 'ongoing',
          exceptions: []
        },
        options: {
          auto_assessment: true,
          risk_scoring: true,
          compliance_check: true,
          control_testing: true,
          reporting: true,
          notifications: true,
          audit_trail: true,
          version_control: true
        }
      };

      const response = await axios.post(`${baseUrl}/governance`, payload, { headers });
      console.log('Governance framework created:', response.data);
      
      // Refresh frameworks list
      loadFrameworks();
      
      return response.data;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error creating governance framework:', errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loadFrameworks = async () => {
    try {
      const response = await axios.get(`${baseUrl}/governance`, { headers });
      setFrameworks(response.data.frameworks);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error loading frameworks:', errorMessage);
    }
  };

  const createGovernanceJob = async () => {
    setLoading(true);
    setError(null);

    try {
      const payload = {
        framework_type: 'data_quality',
        policies: [/* ... */],
        controls: [/* ... */],
        compliance: [/* ... */],
        risk_profile: {/* ... */},
        scope: {/* ... */},
        options: {/* ... */}
      };

      const response = await axios.post(`${baseUrl}/governance/jobs`, payload, { headers });
      console.log('Governance job created:', response.data);
      
      // Refresh jobs list
      loadJobs();
      
      return response.data.job_id;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error creating governance job:', errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const loadJobs = async () => {
    try {
      const response = await axios.get(`${baseUrl}/governance/jobs`, { headers });
      setJobs(response.data.jobs);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message;
      setError(errorMessage);
      console.error('Error loading jobs:', errorMessage);
    }
  };

  const monitorJobProgress = async (jobId: string) => {
    try {
      const response = await axios.get(`${baseUrl}/governance/jobs?id=${jobId}`, { headers });
      const job = response.data;
      
      console.log(`Job ${jobId} status: ${job.status}, progress: ${job.progress * 100}%`);
      
      if (job.status === 'completed') {
        console.log('Job completed with results:', job.result);
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
    loadFrameworks();
    loadJobs();
  }, []);

  return (
    <div className="governance-client">
      <h2>Data Governance Framework</h2>
      
      {error && (
        <div className="error">
          Error: {error}
        </div>
      )}
      
      <div className="actions">
        <button 
          onClick={createGovernanceFramework} 
          disabled={loading}
        >
          {loading ? 'Creating...' : 'Create Framework'}
        </button>
        
        <button 
          onClick={createGovernanceJob} 
          disabled={loading}
        >
          {loading ? 'Creating...' : 'Create Job'}
        </button>
      </div>
      
      <div className="frameworks">
        <h3>Governance Frameworks ({frameworks.length})</h3>
        <div className="framework-list">
          {frameworks.map(framework => (
            <div key={framework.id} className="framework-item">
              <h4>{framework.name}</h4>
              <p>Type: {framework.type}</p>
              <p>Status: {framework.status}</p>
              <p>Version: {framework.version}</p>
            </div>
          ))}
        </div>
      </div>
      
      <div className="jobs">
        <h3>Governance Jobs ({jobs.length})</h3>
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

export default GovernanceClient;
```

## Best Practices

### Framework Design
- **Define Clear Scope**: Specify data domains, business units, and systems covered
- **Risk-Based Approach**: Prioritize controls based on risk assessment
- **Compliance Mapping**: Map controls to specific compliance requirements
- **Regular Reviews**: Schedule periodic framework reviews and updates

### Policy Management
- **Clear Ownership**: Assign clear ownership for each policy
- **Version Control**: Maintain version history for policy changes
- **Compliance Tracking**: Monitor compliance with regulatory requirements
- **Regular Updates**: Update policies based on changing requirements

### Control Implementation
- **Effectiveness Monitoring**: Track control effectiveness over time
- **Testing Schedule**: Establish regular testing schedules
- **Documentation**: Maintain comprehensive control documentation
- **Continuous Improvement**: Regularly assess and improve controls

### Risk Management
- **Risk Assessment**: Conduct regular risk assessments
- **Mitigation Strategies**: Develop and implement risk mitigation strategies
- **Monitoring**: Continuously monitor risk levels
- **Reporting**: Provide regular risk reports to stakeholders

### Compliance Management
- **Standards Mapping**: Map controls to compliance standards
- **Evidence Collection**: Maintain evidence of compliance
- **Audit Preparation**: Prepare for regulatory audits
- **Gap Analysis**: Identify and address compliance gaps

## Rate Limiting

- **Requests per minute**: 100
- **Requests per hour**: 1000
- **Concurrent jobs**: 10

## Monitoring

### Key Metrics
- Framework creation rate
- Job completion time
- Compliance scores
- Risk assessment accuracy
- Control effectiveness

### Alerts
- High-risk assessments
- Compliance violations
- Control failures
- Job failures

## Troubleshooting

### Common Issues

**Validation Errors**
- Ensure all required fields are provided
- Check data types and formats
- Verify enum values are correct

**Job Failures**
- Check job logs for detailed error messages
- Verify input data validity
- Ensure sufficient system resources

**Performance Issues**
- Monitor job queue length
- Check system resource usage
- Consider job prioritization

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

- **Advanced Analytics**: Enhanced reporting and analytics capabilities
- **Integration APIs**: Integration with third-party governance tools
- **Automated Assessments**: AI-powered automated assessments
- **Real-time Monitoring**: Real-time governance monitoring
- **Advanced Risk Modeling**: Sophisticated risk modeling capabilities
