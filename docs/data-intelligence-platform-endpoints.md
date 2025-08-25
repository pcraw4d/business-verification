# Data Intelligence Platform Endpoints

## Overview

The Data Intelligence Platform provides advanced analytics, insights, and intelligence capabilities for the Enhanced Business Intelligence System. This platform enables sophisticated data analysis, pattern recognition, anomaly detection, predictive modeling, and actionable recommendations.

## API Endpoints

### Base URL
```
https://api.kyb-platform.com/v3/intelligence
```

### Authentication
All endpoints require authentication using API keys or JWT tokens in the Authorization header:
```
Authorization: Bearer <your-api-token>
```

---

## 1. Create Intelligence Analysis

**POST** `/intelligence`

Creates and executes an intelligence analysis immediately, providing real-time insights, predictions, and recommendations.

### Request Body

```json
{
  "platform_id": "platform-123",
  "analysis_id": "analysis-456",
  "type": "trend",
  "parameters": {
    "data_source": "business_metrics",
    "time_range": "3_months",
    "confidence_threshold": 0.85
  },
  "data_range": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-03-31T23:59:59Z",
    "time_zone": "UTC"
  },
  "options": {
    "real_time": false,
    "batch_mode": true,
    "parallel": false,
    "notifications": true,
    "audit_trail": true,
    "monitoring": true,
    "validation": true
  }
}
```

### Response

```json
{
  "id": "intelligence-1703123456789",
  "analysis": {
    "id": "analysis-456",
    "name": "trend Analysis",
    "type": "trend",
    "description": "Intelligence analysis of type trend",
    "status": "completed",
    "started_at": "2024-12-19T10:30:45Z",
    "completed_at": "2024-12-19T10:30:46Z",
    "duration": 1000000000,
    "parameters": {
      "data_source": "business_metrics",
      "time_range": "3_months",
      "confidence_threshold": 0.85
    },
    "results": {
      "trend_direction": "upward",
      "trend_strength": 0.85,
      "trend_confidence": 0.92,
      "data_points": 1250
    },
    "errors": [],
    "metadata": {}
  },
  "insights": [
    {
      "id": "insight-1",
      "title": "Strong Upward Trend Detected",
      "description": "Analysis reveals a consistent upward trend in business performance metrics",
      "type": "trend",
      "category": "performance",
      "confidence": 0.92,
      "impact": "high",
      "data": {
        "trend_strength": 0.85,
        "duration": "3 months",
        "growth_rate": "15%"
      },
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "insight-2",
      "title": "Seasonal Pattern Identified",
      "description": "Clear seasonal patterns detected in customer engagement metrics",
      "type": "pattern",
      "category": "behavior",
      "confidence": 0.88,
      "impact": "medium",
      "data": {
        "pattern_type": "seasonal",
        "period": "monthly",
        "strength": 0.78
      },
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "insight-3",
      "title": "Anomaly Detection Alert",
      "description": "Three significant anomalies detected in recent data points",
      "type": "anomaly",
      "category": "alert",
      "confidence": 0.95,
      "impact": "high",
      "data": {
        "anomaly_count": 3,
        "severity": "medium",
        "dates": ["2024-01-15", "2024-02-03", "2024-02-18"]
      },
      "created_at": "2024-12-19T10:30:46Z"
    }
  ],
  "predictions": [
    {
      "id": "prediction-1",
      "title": "Revenue Forecast",
      "description": "Predicted revenue growth for the next 30 days",
      "type": "revenue",
      "value": 1250.5,
      "confidence": 0.87,
      "horizon": 2592000000000000,
      "factors": ["historical_trend", "seasonality", "market_conditions"],
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "prediction-2",
      "title": "Customer Growth",
      "description": "Expected customer acquisition rate",
      "type": "customers",
      "value": 150,
      "confidence": 0.82,
      "horizon": 2592000000000000,
      "factors": ["acquisition_rate", "retention_rate", "market_expansion"],
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "prediction-3",
      "title": "Risk Assessment",
      "description": "Predicted risk level for compliance violations",
      "type": "risk",
      "value": "low",
      "confidence": 0.91,
      "horizon": 604800000000000,
      "factors": ["compliance_history", "regulatory_changes", "internal_controls"],
      "created_at": "2024-12-19T10:30:46Z"
    }
  ],
  "recommendations": [
    {
      "id": "rec-1",
      "title": "Optimize Marketing Strategy",
      "description": "Leverage seasonal patterns to optimize marketing campaigns",
      "type": "strategy",
      "priority": "high",
      "impact": "high",
      "effort": "medium",
      "actions": [
        "adjust_campaign_timing",
        "increase_budget_during_peaks",
        "target_seasonal_customers"
      ],
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "rec-2",
      "title": "Investigate Anomalies",
      "description": "Investigate the three detected anomalies to understand root causes",
      "type": "investigation",
      "priority": "high",
      "impact": "medium",
      "effort": "high",
      "actions": [
        "review_system_logs",
        "analyze_user_behavior",
        "check_external_factors"
      ],
      "created_at": "2024-12-19T10:30:46Z"
    },
    {
      "id": "rec-3",
      "title": "Enhance Monitoring",
      "description": "Implement enhanced monitoring for early anomaly detection",
      "type": "monitoring",
      "priority": "medium",
      "impact": "medium",
      "effort": "low",
      "actions": [
        "set_up_alerts",
        "configure_dashboards",
        "establish_baselines"
      ],
      "created_at": "2024-12-19T10:30:46Z"
    }
  ],
  "statistics": {
    "total_analyses": 15,
    "completed_analyses": 12,
    "failed_analyses": 2,
    "active_analyses": 1,
    "total_insights": 45,
    "total_predictions": 28,
    "total_recommendations": 32,
    "performance_metrics": {
      "avg_processing_time": 2.5,
      "success_rate": 0.93,
      "accuracy": 0.89
    },
    "accuracy_metrics": {
      "prediction_accuracy": 0.87,
      "insight_relevance": 0.92,
      "recommendation_quality": 0.85
    },
    "timeline_events": [
      {
        "id": "event-1",
        "type": "analysis_started",
        "analysis": "analysis-456",
        "action": "intelligence_analysis",
        "status": "completed",
        "timestamp": "2024-12-19T10:30:45Z",
        "duration": 1000.0,
        "description": "Intelligence analysis completed successfully"
      }
    ]
  },
  "timeline": {
    "start_date": "2024-12-19T10:30:45Z",
    "end_date": "2024-12-19T10:30:46Z",
    "duration": 1000.0,
    "milestones": [
      {
        "id": "milestone-1",
        "name": "Analysis Started",
        "description": "Intelligence analysis process initiated",
        "date": "2024-12-19T10:30:45Z",
        "status": "completed",
        "type": "start"
      },
      {
        "id": "milestone-2",
        "name": "Data Processing",
        "description": "Data processing and analysis completed",
        "date": "2024-12-19T10:30:45Z",
        "status": "completed",
        "type": "processing"
      },
      {
        "id": "milestone-3",
        "name": "Analysis Complete",
        "description": "Intelligence analysis completed successfully",
        "date": "2024-12-19T10:30:46Z",
        "status": "completed",
        "type": "completion"
      }
    ],
    "events": [
      {
        "id": "event-1",
        "type": "analysis_started",
        "analysis": "analysis-456",
        "action": "intelligence_analysis",
        "status": "completed",
        "timestamp": "2024-12-19T10:30:45Z",
        "duration": 1000.0,
        "description": "Intelligence analysis started"
      }
    ],
    "projections": [
      {
        "type": "performance",
        "date": "2025-01-19T10:30:46Z",
        "confidence": 0.85,
        "description": "Expected performance improvement based on insights"
      }
    ]
  },
  "created_at": "2024-12-19T10:30:46Z",
  "status": "completed"
}
```

### Analysis Types

| Type | Description | Use Case |
|------|-------------|----------|
| `trend` | Identifies trends and patterns in data over time | Business performance analysis, growth tracking |
| `pattern` | Detects recurring patterns and seasonality | Customer behavior analysis, operational patterns |
| `anomaly` | Identifies unusual data points and outliers | Fraud detection, system monitoring |
| `prediction` | Forecasts future values based on historical data | Revenue forecasting, demand planning |
| `correlation` | Analyzes relationships between variables | Risk assessment, feature analysis |
| `clustering` | Groups similar data points into clusters | Customer segmentation, market analysis |

---

## 2. Get Intelligence Analysis

**GET** `/intelligence?id={analysis_id}`

Retrieves details of a specific intelligence analysis.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | The ID of the analysis to retrieve |

### Response

Returns the same structure as the Create Intelligence Analysis response.

---

## 3. List Intelligence Analyses

**GET** `/intelligence`

Lists all intelligence analyses in the system.

### Response

```json
{
  "analyses": [
    {
      "id": "analysis-1",
      "name": "Sample Intelligence Analysis",
      "type": "trend",
      "description": "Sample intelligence analysis for demonstration",
      "status": "completed",
      "started_at": "2024-12-18T10:30:45Z",
      "completed_at": "2024-12-18T10:35:45Z",
      "duration": 300000000000,
      "parameters": {},
      "results": {
        "trend_direction": "upward",
        "trend_strength": 0.85,
        "confidence": 0.92
      },
      "errors": [],
      "metadata": {}
    }
  ],
  "total": 1,
  "timestamp": "2024-12-19T10:30:46Z"
}
```

---

## 4. Create Intelligence Job

**POST** `/intelligence/jobs`

Creates a background intelligence analysis job for long-running analyses.

### Request Body

Same as Create Intelligence Analysis.

### Response

```json
{
  "job_id": "intelligence-1703123456789",
  "status": "created",
  "created_at": "2024-12-19T10:30:46Z"
}
```

---

## 5. Get Intelligence Job

**GET** `/intelligence/jobs?id={job_id}`

Retrieves the status and results of a background intelligence job.

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | The ID of the job to retrieve |

### Response

```json
{
  "id": "intelligence-1703123456789",
  "type": "intelligence_analysis",
  "status": "completed",
  "progress": 1.0,
  "created_at": "2024-12-19T10:30:46Z",
  "started_at": "2024-12-19T10:30:47Z",
  "completed_at": "2024-12-19T10:30:52Z",
  "result": {
    "analysis_id": "analysis-456",
    "insights": [...],
    "predictions": [...],
    "recommendations": [...],
    "statistics": {...},
    "timeline": {...},
    "generated_at": "2024-12-19T10:30:52Z"
  },
  "error": null
}
```

---

## 6. List Intelligence Jobs

**GET** `/intelligence/jobs`

Lists all intelligence jobs in the system.

### Response

```json
{
  "jobs": [
    {
      "id": "intelligence-1703123456789",
      "type": "intelligence_analysis",
      "status": "completed",
      "progress": 1.0,
      "created_at": "2024-12-19T10:30:46Z",
      "started_at": "2024-12-19T10:30:47Z",
      "completed_at": "2024-12-19T10:30:52Z",
      "result": {...},
      "error": null
    }
  ],
  "total": 1,
  "timestamp": "2024-12-19T10:30:52Z"
}
```

---

## Data Models

### Intelligence Analysis Types

```typescript
enum IntelligenceAnalysisType {
  TREND = "trend",
  PATTERN = "pattern",
  ANOMALY = "anomaly",
  PREDICTION = "prediction",
  CORRELATION = "correlation",
  CLUSTERING = "clustering"
}
```

### Intelligence Status

```typescript
enum IntelligenceStatus {
  PENDING = "pending",
  RUNNING = "running",
  COMPLETED = "completed",
  FAILED = "failed",
  CANCELLED = "cancelled"
}
```

### Data Source Types

```typescript
enum DataSourceType {
  INTERNAL = "internal",
  EXTERNAL = "external",
  API = "api",
  DATABASE = "database",
  FILE = "file",
  STREAM = "stream"
}
```

### Intelligence Model Types

```typescript
enum IntelligenceModelType {
  MACHINE_LEARNING = "machine_learning",
  STATISTICAL = "statistical",
  RULE_BASED = "rule_based",
  HYBRID = "hybrid",
  CUSTOM = "custom"
}
```

---

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class IntelligencePlatformClient {
  constructor(baseURL, apiKey) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createAnalysis(analysisRequest) {
    try {
      const response = await this.client.post('/intelligence', analysisRequest);
      return response.data;
    } catch (error) {
      console.error('Error creating intelligence analysis:', error.response?.data);
      throw error;
    }
  }

  async getAnalysis(analysisId) {
    try {
      const response = await this.client.get(`/intelligence?id=${analysisId}`);
      return response.data;
    } catch (error) {
      console.error('Error retrieving intelligence analysis:', error.response?.data);
      throw error;
    }
  }

  async createJob(analysisRequest) {
    try {
      const response = await this.client.post('/intelligence/jobs', analysisRequest);
      return response.data;
    } catch (error) {
      console.error('Error creating intelligence job:', error.response?.data);
      throw error;
    }
  }

  async getJob(jobId) {
    try {
      const response = await this.client.get(`/intelligence/jobs?id=${jobId}`);
      return response.data;
    } catch (error) {
      console.error('Error retrieving intelligence job:', error.response?.data);
      throw error;
    }
  }

  async pollJobCompletion(jobId, interval = 5000, timeout = 300000) {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      const job = await this.getJob(jobId);
      
      if (job.status === 'completed') {
        return job.result;
      } else if (job.status === 'failed') {
        throw new Error(`Job failed: ${job.error}`);
      }
      
      await new Promise(resolve => setTimeout(resolve, interval));
    }
    
    throw new Error('Job polling timeout');
  }
}

// Usage example
const client = new IntelligencePlatformClient('https://api.kyb-platform.com/v3', 'your-api-key');

const analysisRequest = {
  platform_id: "platform-123",
  analysis_id: "analysis-456",
  type: "trend",
  parameters: {
    data_source: "business_metrics",
    time_range: "3_months"
  },
  data_range: {
    start_date: "2024-01-01T00:00:00Z",
    end_date: "2024-03-31T23:59:59Z",
    time_zone: "UTC"
  },
  options: {
    real_time: false,
    batch_mode: true,
    notifications: true,
    audit_trail: true,
    monitoring: true,
    validation: true
  }
};

// Create analysis immediately
client.createAnalysis(analysisRequest)
  .then(result => {
    console.log('Analysis completed:', result.insights);
    console.log('Predictions:', result.predictions);
    console.log('Recommendations:', result.recommendations);
  })
  .catch(error => {
    console.error('Analysis failed:', error);
  });

// Create background job
client.createJob(analysisRequest)
  .then(job => {
    console.log('Job created:', job.job_id);
    return client.pollJobCompletion(job.job_id);
  })
  .then(result => {
    console.log('Job completed:', result);
  })
  .catch(error => {
    console.error('Job failed:', error);
  });
```

### Python

```python
import requests
import time
from typing import Dict, Any, Optional

class IntelligencePlatformClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }
    
    def create_analysis(self, analysis_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create and execute an intelligence analysis immediately."""
        try:
            response = requests.post(
                f'{self.base_url}/intelligence',
                json=analysis_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f'Error creating intelligence analysis: {e}')
            raise
    
    def get_analysis(self, analysis_id: str) -> Dict[str, Any]:
        """Retrieve a specific intelligence analysis."""
        try:
            response = requests.get(
                f'{self.base_url}/intelligence',
                params={'id': analysis_id},
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f'Error retrieving intelligence analysis: {e}')
            raise
    
    def create_job(self, analysis_request: Dict[str, Any]) -> Dict[str, Any]:
        """Create a background intelligence analysis job."""
        try:
            response = requests.post(
                f'{self.base_url}/intelligence/jobs',
                json=analysis_request,
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f'Error creating intelligence job: {e}')
            raise
    
    def get_job(self, job_id: str) -> Dict[str, Any]:
        """Retrieve the status of a background intelligence job."""
        try:
            response = requests.get(
                f'{self.base_url}/intelligence/jobs',
                params={'id': job_id},
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f'Error retrieving intelligence job: {e}')
            raise
    
    def poll_job_completion(self, job_id: str, interval: int = 5, timeout: int = 300) -> Dict[str, Any]:
        """Poll for job completion with timeout."""
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            job = self.get_job(job_id)
            
            if job['status'] == 'completed':
                return job['result']
            elif job['status'] == 'failed':
                raise Exception(f"Job failed: {job.get('error', 'Unknown error')}")
            
            time.sleep(interval)
        
        raise Exception('Job polling timeout')

# Usage example
client = IntelligencePlatformClient('https://api.kyb-platform.com/v3', 'your-api-key')

analysis_request = {
    'platform_id': 'platform-123',
    'analysis_id': 'analysis-456',
    'type': 'trend',
    'parameters': {
        'data_source': 'business_metrics',
        'time_range': '3_months'
    },
    'data_range': {
        'start_date': '2024-01-01T00:00:00Z',
        'end_date': '2024-03-31T23:59:59Z',
        'time_zone': 'UTC'
    },
    'options': {
        'real_time': False,
        'batch_mode': True,
        'notifications': True,
        'audit_trail': True,
        'monitoring': True,
        'validation': True
    }
}

# Create analysis immediately
try:
    result = client.create_analysis(analysis_request)
    print('Analysis completed:')
    print(f'Insights: {len(result["insights"])}')
    print(f'Predictions: {len(result["predictions"])}')
    print(f'Recommendations: {len(result["recommendations"])}')
except Exception as e:
    print(f'Analysis failed: {e}')

# Create background job
try:
    job = client.create_job(analysis_request)
    print(f'Job created: {job["job_id"]}')
    
    result = client.poll_job_completion(job['job_id'])
    print('Job completed:', result)
except Exception as e:
    print(f'Job failed: {e}')
```

### React/TypeScript

```typescript
import React, { useState, useEffect } from 'react';
import axios, { AxiosInstance } from 'axios';

interface IntelligenceAnalysisRequest {
  platform_id: string;
  analysis_id: string;
  type: 'trend' | 'pattern' | 'anomaly' | 'prediction' | 'correlation' | 'clustering';
  parameters: Record<string, any>;
  data_range: {
    start_date: string;
    end_date: string;
    time_zone: string;
  };
  options: {
    real_time: boolean;
    batch_mode: boolean;
    parallel: boolean;
    notifications: boolean;
    audit_trail: boolean;
    monitoring: boolean;
    validation: boolean;
  };
}

interface IntelligenceAnalysisResponse {
  id: string;
  analysis: {
    id: string;
    name: string;
    type: string;
    status: string;
    started_at: string;
    completed_at: string;
    duration: number;
    results: Record<string, any>;
  };
  insights: Array<{
    id: string;
    title: string;
    description: string;
    type: string;
    category: string;
    confidence: number;
    impact: string;
    data: Record<string, any>;
    created_at: string;
  }>;
  predictions: Array<{
    id: string;
    title: string;
    description: string;
    type: string;
    value: any;
    confidence: number;
    horizon: number;
    factors: string[];
    created_at: string;
  }>;
  recommendations: Array<{
    id: string;
    title: string;
    description: string;
    type: string;
    priority: string;
    impact: string;
    effort: string;
    actions: string[];
    created_at: string;
  }>;
  statistics: {
    total_analyses: number;
    completed_analyses: number;
    failed_analyses: number;
    active_analyses: number;
    total_insights: number;
    total_predictions: number;
    total_recommendations: number;
    performance_metrics: Record<string, number>;
    accuracy_metrics: Record<string, number>;
  };
  timeline: {
    start_date: string;
    end_date: string;
    duration: number;
    milestones: Array<{
      id: string;
      name: string;
      description: string;
      date: string;
      status: string;
      type: string;
    }>;
    events: Array<{
      id: string;
      type: string;
      analysis: string;
      action: string;
      status: string;
      timestamp: string;
      duration: number;
      description: string;
    }>;
    projections: Array<{
      type: string;
      date: string;
      confidence: number;
      description: string;
    }>;
  };
  created_at: string;
  status: string;
}

class IntelligencePlatformClient {
  private client: AxiosInstance;

  constructor(baseURL: string, apiKey: string) {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
      },
    });
  }

  async createAnalysis(request: IntelligenceAnalysisRequest): Promise<IntelligenceAnalysisResponse> {
    try {
      const response = await this.client.post('/intelligence', request);
      return response.data;
    } catch (error) {
      console.error('Error creating intelligence analysis:', error);
      throw error;
    }
  }

  async getAnalysis(analysisId: string): Promise<IntelligenceAnalysisResponse> {
    try {
      const response = await this.client.get(`/intelligence?id=${analysisId}`);
      return response.data;
    } catch (error) {
      console.error('Error retrieving intelligence analysis:', error);
      throw error;
    }
  }

  async createJob(request: IntelligenceAnalysisRequest): Promise<{ job_id: string; status: string; created_at: string }> {
    try {
      const response = await this.client.post('/intelligence/jobs', request);
      return response.data;
    } catch (error) {
      console.error('Error creating intelligence job:', error);
      throw error;
    }
  }

  async getJob(jobId: string): Promise<any> {
    try {
      const response = await this.client.get(`/intelligence/jobs?id=${jobId}`);
      return response.data;
    } catch (error) {
      console.error('Error retrieving intelligence job:', error);
      throw error;
    }
  }

  async pollJobCompletion(jobId: string, interval: number = 5000, timeout: number = 300000): Promise<any> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      const job = await this.getJob(jobId);
      
      if (job.status === 'completed') {
        return job.result;
      } else if (job.status === 'failed') {
        throw new Error(`Job failed: ${job.error}`);
      }
      
      await new Promise(resolve => setTimeout(resolve, interval));
    }
    
    throw new Error('Job polling timeout');
  }
}

// React Component Example
const IntelligenceAnalysisComponent: React.FC = () => {
  const [analysis, setAnalysis] = useState<IntelligenceAnalysisResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const client = new IntelligencePlatformClient('https://api.kyb-platform.com/v3', 'your-api-key');

  const runAnalysis = async () => {
    setLoading(true);
    setError(null);

    const analysisRequest: IntelligenceAnalysisRequest = {
      platform_id: 'platform-123',
      analysis_id: 'analysis-456',
      type: 'trend',
      parameters: {
        data_source: 'business_metrics',
        time_range: '3_months',
      },
      data_range: {
        start_date: '2024-01-01T00:00:00Z',
        end_date: '2024-03-31T23:59:59Z',
        time_zone: 'UTC',
      },
      options: {
        real_time: false,
        batch_mode: true,
        parallel: false,
        notifications: true,
        audit_trail: true,
        monitoring: true,
        validation: true,
      },
    };

    try {
      const result = await client.createAnalysis(analysisRequest);
      setAnalysis(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="intelligence-analysis">
      <h2>Intelligence Analysis</h2>
      
      <button onClick={runAnalysis} disabled={loading}>
        {loading ? 'Running Analysis...' : 'Run Analysis'}
      </button>

      {error && (
        <div className="error">
          <h3>Error</h3>
          <p>{error}</p>
        </div>
      )}

      {analysis && (
        <div className="results">
          <h3>Analysis Results</h3>
          
          <div className="insights">
            <h4>Insights ({analysis.insights.length})</h4>
            {analysis.insights.map(insight => (
              <div key={insight.id} className="insight">
                <h5>{insight.title}</h5>
                <p>{insight.description}</p>
                <div className="metadata">
                  <span>Type: {insight.type}</span>
                  <span>Category: {insight.category}</span>
                  <span>Confidence: {(insight.confidence * 100).toFixed(1)}%</span>
                  <span>Impact: {insight.impact}</span>
                </div>
              </div>
            ))}
          </div>

          <div className="predictions">
            <h4>Predictions ({analysis.predictions.length})</h4>
            {analysis.predictions.map(prediction => (
              <div key={prediction.id} className="prediction">
                <h5>{prediction.title}</h5>
                <p>{prediction.description}</p>
                <div className="metadata">
                  <span>Value: {prediction.value}</span>
                  <span>Confidence: {(prediction.confidence * 100).toFixed(1)}%</span>
                  <span>Horizon: {Math.round(prediction.horizon / (1000 * 60 * 60 * 24))} days</span>
                </div>
              </div>
            ))}
          </div>

          <div className="recommendations">
            <h4>Recommendations ({analysis.recommendations.length})</h4>
            {analysis.recommendations.map(recommendation => (
              <div key={recommendation.id} className="recommendation">
                <h5>{recommendation.title}</h5>
                <p>{recommendation.description}</p>
                <div className="metadata">
                  <span>Priority: {recommendation.priority}</span>
                  <span>Impact: {recommendation.impact}</span>
                  <span>Effort: {recommendation.effort}</span>
                </div>
                <div className="actions">
                  <h6>Actions:</h6>
                  <ul>
                    {recommendation.actions.map((action, index) => (
                      <li key={index}>{action}</li>
                    ))}
                  </ul>
                </div>
              </div>
            ))}
          </div>

          <div className="statistics">
            <h4>Statistics</h4>
            <div className="stats-grid">
              <div>Total Analyses: {analysis.statistics.total_analyses}</div>
              <div>Completed: {analysis.statistics.completed_analyses}</div>
              <div>Failed: {analysis.statistics.failed_analyses}</div>
              <div>Active: {analysis.statistics.active_analyses}</div>
              <div>Total Insights: {analysis.statistics.total_insights}</div>
              <div>Total Predictions: {analysis.statistics.total_predictions}</div>
              <div>Total Recommendations: {analysis.statistics.total_recommendations}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default IntelligenceAnalysisComponent;
```

---

## Best Practices

### 1. Analysis Design

- **Choose the Right Analysis Type**: Select the analysis type that best fits your use case
- **Set Appropriate Parameters**: Configure parameters based on your data characteristics
- **Use Real-time vs Batch Mode**: Use real-time for immediate insights, batch mode for comprehensive analysis
- **Validate Data Quality**: Ensure data quality before running analysis

### 2. Performance Optimization

- **Use Background Jobs**: For long-running analyses, use background jobs instead of immediate execution
- **Monitor Job Progress**: Implement progress tracking for better user experience
- **Handle Timeouts**: Set appropriate timeouts for job polling
- **Cache Results**: Cache analysis results for frequently requested data

### 3. Error Handling

- **Validate Inputs**: Always validate request parameters before submission
- **Handle Network Errors**: Implement retry logic for network failures
- **Check Job Status**: Monitor job status and handle failures gracefully
- **Log Errors**: Log errors for debugging and monitoring

### 4. Security Considerations

- **API Key Management**: Securely store and rotate API keys
- **Input Validation**: Validate all input parameters to prevent injection attacks
- **Rate Limiting**: Implement rate limiting to prevent abuse
- **Data Privacy**: Ensure sensitive data is properly handled and encrypted

---

## Troubleshooting

### Common Issues

1. **Analysis Fails with Validation Error**
   - Check that all required fields are provided
   - Verify data types and formats
   - Ensure analysis type is valid

2. **Job Stuck in Pending Status**
   - Check system resources and capacity
   - Verify data source connectivity
   - Review job parameters and configuration

3. **Low Confidence Scores**
   - Ensure sufficient data points are available
   - Check data quality and completeness
   - Adjust analysis parameters as needed

4. **Timeout Errors**
   - Increase timeout values for large datasets
   - Use background jobs for long-running analyses
   - Optimize data processing parameters

### Debug Information

Enable debug logging to get detailed information about analysis execution:

```json
{
  "options": {
    "audit_trail": true,
    "monitoring": true,
    "validation": true
  }
}
```

### Support Resources

- **API Documentation**: Complete endpoint documentation and examples
- **Integration Guides**: Step-by-step integration tutorials
- **Community Forum**: User community for questions and support
- **Technical Support**: Direct support for enterprise customers

---

## Rate Limits

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| Create Analysis | 10 requests | 1 minute |
| Get Analysis | 100 requests | 1 minute |
| List Analyses | 50 requests | 1 minute |
| Create Job | 5 requests | 1 minute |
| Get Job | 100 requests | 1 minute |
| List Jobs | 50 requests | 1 minute |

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 3.0.0 | 2024-12-19 | Initial release with comprehensive intelligence platform |
| 3.0.1 | 2024-12-19 | Added clustering analysis type |
| 3.0.2 | 2024-12-19 | Enhanced prediction capabilities |

---

## Changelog

### v3.0.2 (2024-12-19)
- Enhanced prediction capabilities with confidence intervals
- Added support for multiple prediction horizons
- Improved anomaly detection algorithms
- Added correlation analysis type

### v3.0.1 (2024-12-19)
- Added clustering analysis type for customer segmentation
- Enhanced pattern recognition algorithms
- Improved recommendation quality scoring
- Added support for custom analysis parameters

### v3.0.0 (2024-12-19)
- Initial release of the Data Intelligence Platform
- Support for 6 analysis types (trend, pattern, anomaly, prediction, correlation, clustering)
- Comprehensive insights, predictions, and recommendations
- Background job processing for long-running analyses
- Real-time and batch analysis modes
- Advanced statistics and timeline tracking
