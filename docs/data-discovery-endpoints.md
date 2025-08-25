# Data Discovery API Endpoints

## Overview

The Data Discovery API provides comprehensive data discovery capabilities for identifying, profiling, and analyzing data assets across various sources. This API supports automated discovery, manual discovery, scheduled discovery, and incremental discovery with advanced profiling and pattern detection.

## Authentication

All endpoints require authentication using API keys:

```http
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All responses are returned in JSON format with the following structure:

```json
{
  "id": "discovery_1234567890",
  "name": "Discovery Name",
  "type": "auto",
  "status": "completed",
  "results": { ... },
  "created_at": "2024-12-19T10:30:00Z"
}
```

## Supported Discovery Types

- `auto` - Automated discovery with default settings
- `manual` - Manual discovery with custom configuration
- `scheduled` - Scheduled discovery with cron expressions
- `incremental` - Incremental discovery for new/changed data
- `full` - Full discovery of all data sources

## Supported Discovery Statuses

- `pending` - Discovery is queued for processing
- `running` - Discovery is currently executing
- `completed` - Discovery has finished successfully
- `failed` - Discovery encountered an error
- `cancelled` - Discovery was cancelled

## Supported Profile Types

- `statistical` - Statistical analysis of data
- `quality` - Data quality assessment
- `pattern` - Pattern detection and analysis
- `anomaly` - Anomaly detection
- `comprehensive` - Complete profiling including all types

## Supported Pattern Types

- `temporal` - Time-based patterns
- `sequential` - Sequential patterns
- `correlation` - Correlation patterns
- `outlier` - Outlier detection
- `trend` - Trend analysis
- `seasonal` - Seasonal patterns
- `cyclic` - Cyclic patterns
- `custom` - Custom pattern detection

## API Endpoints

### 1. Create Discovery

**POST** `/discovery`

Creates and executes a data discovery process immediately.

#### Request Body

```json
{
  "name": "Customer Data Discovery",
  "description": "Discover and profile customer data across all sources",
  "type": "auto",
  "sources": [
    {
      "id": "source_1",
      "name": "Customer Database",
      "type": "database",
      "location": "postgres://localhost:5432/customers",
      "connection": {
        "id": "conn_1",
        "name": "Customer DB Connection",
        "type": "postgresql",
        "protocol": "postgres",
        "host": "localhost",
        "port": 5432,
        "database": "customers"
      },
      "enabled": true
    }
  ],
  "rules": [
    {
      "id": "rule_1",
      "name": "PII Detection",
      "type": "validation",
      "description": "Detect personally identifiable information",
      "condition": "column_name LIKE '%email%' OR column_name LIKE '%phone%'",
      "action": "flag_as_pii",
      "priority": 1,
      "enabled": true
    }
  ],
  "profiles": [
    {
      "id": "profile_1",
      "name": "Statistical Profile",
      "type": "statistical",
      "description": "Generate statistical analysis",
      "config": {
        "sample_size": 1000,
        "confidence": 0.95,
        "thresholds": {
          "completeness": 0.8,
          "accuracy": 0.9
        }
      },
      "enabled": true
    }
  ],
  "patterns": [
    {
      "id": "pattern_1",
      "name": "Temporal Pattern",
      "type": "temporal",
      "description": "Detect temporal patterns",
      "algorithm": "seasonal_decomposition",
      "config": {
        "window_size": 30,
        "sensitivity": 0.8,
        "min_occurrence": 5
      },
      "enabled": true
    }
  ],
  "filters": {
    "types": ["database", "table"],
    "sources": ["source_1"],
    "date_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-12-19T23:59:59Z"
    },
    "size_range": {
      "min": 1000,
      "max": 1000000000
    },
    "tags": ["customer", "pii"]
  },
  "options": {
    "parallel": true,
    "max_workers": 4,
    "timeout": 3600,
    "retry_count": 3,
    "batch_size": 100,
    "include_stats": true,
    "include_profiles": true,
    "include_patterns": true
  },
  "schedule": {
    "type": "daily",
    "cron": "0 2 * * *",
    "start_at": "2024-12-20T02:00:00Z",
    "enabled": true
  }
}
```

#### Response

```json
{
  "id": "discovery_1234567890",
  "name": "Customer Data Discovery",
  "type": "auto",
  "status": "completed",
  "sources": [...],
  "rules": [...],
  "profiles": [...],
  "patterns": [...],
  "results": {
    "assets": [
      {
        "id": "asset_1",
        "name": "Customer Database",
        "type": "database",
        "location": "postgres://localhost:5432/customers",
        "size": 1024000000,
        "format": "postgresql",
        "schema": {
          "type": "relational",
          "version": "1.0",
          "columns": [
            {
              "name": "customer_id",
              "type": "integer",
              "description": "Unique customer identifier",
              "nullable": false,
              "primary_key": true
            },
            {
              "name": "name",
              "type": "varchar",
              "description": "Customer name",
              "nullable": false,
              "length": 255
            }
          ]
        },
        "quality": {
          "score": 0.85,
          "completeness": 0.90,
          "accuracy": 0.85,
          "consistency": 0.88,
          "validity": 0.92,
          "timeliness": 0.75,
          "uniqueness": 0.95,
          "integrity": 0.88,
          "issues": []
        },
        "profile": {
          "statistics": {
            "count": 10000,
            "mean": 0.0,
            "median": 0.0,
            "std_dev": 0.0
          }
        },
        "patterns": [
          {
            "id": "pattern_1",
            "type": "temporal",
            "column": "created_at",
            "pattern": "daily_cycle",
            "confidence": 0.85,
            "occurrences": 365,
            "frequency": 1.0
          }
        ],
        "anomalies": [
          {
            "id": "anomaly_1",
            "type": "outlier",
            "column": "customer_id",
            "value": "999999",
            "score": 0.95,
            "severity": "high",
            "description": "Unusual customer ID value"
          }
        ],
        "tags": ["customer", "pii", "production"],
        "discovered_at": "2024-12-19T10:30:00Z"
      }
    ],
    "profiles": [...],
    "patterns": [...],
    "anomalies": [...],
    "recommendations": [
      {
        "id": "rec_1",
        "type": "quality",
        "title": "Improve Data Quality",
        "description": "Address data quality issues in customer database",
        "priority": "high",
        "impact": "high",
        "effort": "medium",
        "actions": [
          "Review data validation rules",
          "Implement data quality monitoring"
        ]
      }
    ]
  },
  "summary": {
    "total_assets": 1,
    "total_profiles": 1,
    "total_patterns": 1,
    "total_anomalies": 1,
    "asset_types": {
      "database": 1
    },
    "quality_scores": {
      "overall": 0.85
    },
    "pattern_types": {
      "temporal": 1
    },
    "anomaly_types": {
      "outlier": 1
    },
    "data_volume": "1GB",
    "coverage": 0.85,
    "completeness": 0.90
  },
  "statistics": {
    "performance_stats": {
      "total_time": 120.5,
      "avg_time_per_asset": 120.5,
      "throughput": 0.008,
      "success_rate": 1.0,
      "error_rate": 0.0
    },
    "quality_stats": {
      "overall_score": 0.85,
      "high_quality": 0,
      "medium_quality": 1,
      "low_quality": 0,
      "trend_direction": "stable"
    },
    "pattern_stats": {
      "total_patterns": 1,
      "pattern_types": {
        "temporal": 1
      },
      "avg_confidence": 0.85,
      "high_confidence": 1
    },
    "anomaly_stats": {
      "total_anomalies": 1,
      "anomaly_types": {
        "outlier": 1
      },
      "severity_levels": {
        "high": 1
      },
      "avg_score": 0.95,
      "high_severity": 1
    },
    "trends": [
      {
        "metric": "assets_discovered",
        "period": "daily",
        "values": [1, 2, 1, 3, 2, 4, 3],
        "timestamps": [...],
        "direction": "increasing",
        "change": 0.5,
        "significance": "moderate"
      }
    ]
  },
  "insights": [
    {
      "id": "insight_1",
      "type": "quality",
      "title": "Data Quality Issues Detected",
      "description": "Several data quality issues were found in the discovered assets",
      "severity": "medium",
      "confidence": 0.85,
      "impact": "medium",
      "actions": [
        "Review data validation rules",
        "Implement quality monitoring"
      ]
    },
    {
      "id": "insight_2",
      "type": "pattern",
      "title": "Temporal Patterns Identified",
      "description": "Strong temporal patterns were detected in customer data",
      "severity": "low",
      "confidence": 0.90,
      "impact": "low",
      "actions": [
        "Consider time-based analytics",
        "Implement temporal monitoring"
      ]
    }
  ],
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "completed_at": "2024-12-19T10:32:00Z"
}
```

### 2. Get Discovery

**GET** `/discovery?id={id}`

Retrieves details of a specific discovery by ID.

#### Parameters

- `id` (required) - The discovery ID

#### Response

Returns the same structure as the Create Discovery response.

### 3. List Discoveries

**GET** `/discovery`

Lists all discoveries with pagination support.

#### Response

```json
{
  "discoveries": [
    {
      "id": "discovery_1234567890",
      "name": "Customer Data Discovery",
      "type": "auto",
      "status": "completed",
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 4. Create Discovery Job

**POST** `/discovery/jobs`

Creates a background discovery job for asynchronous processing.

#### Request Body

Same as Create Discovery endpoint.

#### Response

```json
{
  "id": "discovery_job_1234567890",
  "request_id": "req_1234567890",
  "type": "discovery_creation",
  "status": "pending",
  "progress": 0,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 5. Get Discovery Job

**GET** `/discovery/jobs?id={id}`

Retrieves the status of a specific discovery job.

#### Parameters

- `id` (required) - The job ID

#### Response

```json
{
  "id": "discovery_job_1234567890",
  "request_id": "req_1234567890",
  "type": "discovery_creation",
  "status": "completed",
  "progress": 100,
  "result": {
    "id": "discovery_1234567890",
    "name": "Customer Data Discovery",
    "type": "auto",
    "status": "completed",
    "results": {...}
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:32:00Z",
  "completed_at": "2024-12-19T10:32:00Z"
}
```

### 6. List Discovery Jobs

**GET** `/discovery/jobs`

Lists all discovery jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "discovery_job_1234567890",
      "request_id": "req_1234567890",
      "type": "discovery_creation",
      "status": "completed",
      "progress": 100,
      "created_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "name is required"
}
```

### 404 Not Found

```json
{
  "error": "Discovery not found"
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

// Create discovery
async function createDiscovery() {
  try {
    const response = await axios.post('https://api.example.com/discovery', {
      name: 'Customer Data Discovery',
      type: 'auto',
      sources: [{
        id: 'source_1',
        name: 'Customer Database',
        type: 'database',
        location: 'postgres://localhost:5432/customers',
        enabled: true
      }]
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });
    
    console.log('Discovery created:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error creating discovery:', error.response.data);
  }
}

// Get discovery status
async function getDiscovery(id) {
  try {
    const response = await axios.get(`https://api.example.com/discovery?id=${id}`, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY'
      }
    });
    
    console.log('Discovery details:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error getting discovery:', error.response.data);
  }
}

// Create background job
async function createDiscoveryJob() {
  try {
    const response = await axios.post('https://api.example.com/discovery/jobs', {
      name: 'Scheduled Discovery',
      type: 'scheduled',
      sources: [{
        id: 'source_1',
        name: 'Customer Database',
        type: 'database',
        location: 'postgres://localhost:5432/customers',
        enabled: true
      }],
      schedule: {
        type: 'daily',
        cron: '0 2 * * *',
        enabled: true
      }
    }, {
      headers: {
        'Authorization': 'Bearer YOUR_API_KEY',
        'Content-Type': 'application/json'
      }
    });
    
    console.log('Discovery job created:', response.data);
    return response.data;
  } catch (error) {
    console.error('Error creating discovery job:', error.response.data);
  }
}
```

### Python

```python
import requests
import json

class DataDiscoveryAPI:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }
    
    def create_discovery(self, discovery_data):
        """Create a new discovery"""
        response = requests.post(
            f'{self.base_url}/discovery',
            headers=self.headers,
            json=discovery_data
        )
        response.raise_for_status()
        return response.json()
    
    def get_discovery(self, discovery_id):
        """Get discovery details"""
        response = requests.get(
            f'{self.base_url}/discovery',
            headers=self.headers,
            params={'id': discovery_id}
        )
        response.raise_for_status()
        return response.json()
    
    def list_discoveries(self):
        """List all discoveries"""
        response = requests.get(
            f'{self.base_url}/discovery',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
    
    def create_discovery_job(self, discovery_data):
        """Create a background discovery job"""
        response = requests.post(
            f'{self.base_url}/discovery/jobs',
            headers=self.headers,
            json=discovery_data
        )
        response.raise_for_status()
        return response.json()
    
    def get_discovery_job(self, job_id):
        """Get job status"""
        response = requests.get(
            f'{self.base_url}/discovery/jobs',
            headers=self.headers,
            params={'id': job_id}
        )
        response.raise_for_status()
        return response.json()

# Usage example
api = DataDiscoveryAPI('https://api.example.com', 'YOUR_API_KEY')

# Create discovery
discovery_data = {
    'name': 'Customer Data Discovery',
    'type': 'auto',
    'sources': [{
        'id': 'source_1',
        'name': 'Customer Database',
        'type': 'database',
        'location': 'postgres://localhost:5432/customers',
        'enabled': True
    }]
}

try:
    discovery = api.create_discovery(discovery_data)
    print(f"Discovery created: {discovery['id']}")
    
    # Get discovery details
    details = api.get_discovery(discovery['id'])
    print(f"Discovery status: {details['status']}")
    
except requests.exceptions.RequestException as e:
    print(f"API error: {e}")
```

### React/TypeScript

```typescript
interface DiscoveryRequest {
  name: string;
  type: string;
  sources: DiscoverySource[];
  rules?: DiscoveryRule[];
  profiles?: DiscoveryProfile[];
  patterns?: DiscoveryPattern[];
  filters?: DiscoveryFilters;
  options?: DiscoveryOptions;
  schedule?: DiscoverySchedule;
}

interface DiscoveryResponse {
  id: string;
  name: string;
  type: string;
  status: string;
  results: DiscoveryResults;
  summary: DiscoverySummary;
  statistics: DiscoveryStatistics;
  insights: DiscoveryInsight[];
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

class DataDiscoveryService {
  private baseUrl: string;
  private apiKey: string;

  constructor(baseUrl: string, apiKey: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`);
    }

    return response.json();
  }

  async createDiscovery(data: DiscoveryRequest): Promise<DiscoveryResponse> {
    return this.request<DiscoveryResponse>('/discovery', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getDiscovery(id: string): Promise<DiscoveryResponse> {
    return this.request<DiscoveryResponse>(`/discovery?id=${id}`);
  }

  async listDiscoveries(): Promise<{ discoveries: DiscoveryResponse[]; total: number }> {
    return this.request<{ discoveries: DiscoveryResponse[]; total: number }>('/discovery');
  }

  async createDiscoveryJob(data: DiscoveryRequest): Promise<any> {
    return this.request('/discovery/jobs', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getDiscoveryJob(id: string): Promise<any> {
    return this.request(`/discovery/jobs?id=${id}`);
  }
}

// React component example
import React, { useState, useEffect } from 'react';

const DiscoveryComponent: React.FC = () => {
  const [discoveries, setDiscoveries] = useState<DiscoveryResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const discoveryService = new DataDiscoveryService(
    'https://api.example.com',
    'YOUR_API_KEY'
  );

  useEffect(() => {
    loadDiscoveries();
  }, []);

  const loadDiscoveries = async () => {
    setLoading(true);
    try {
      const response = await discoveryService.listDiscoveries();
      setDiscoveries(response.discoveries);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const createDiscovery = async () => {
    const discoveryData: DiscoveryRequest = {
      name: 'New Discovery',
      type: 'auto',
      sources: [{
        id: 'source_1',
        name: 'Test Source',
        type: 'database',
        location: 'postgres://localhost:5432/test',
        enabled: true,
      }],
    };

    try {
      const discovery = await discoveryService.createDiscovery(discoveryData);
      console.log('Discovery created:', discovery);
      loadDiscoveries(); // Refresh list
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div>
      <h1>Data Discoveries</h1>
      <button onClick={createDiscovery}>Create New Discovery</button>
      
      <div>
        {discoveries.map(discovery => (
          <div key={discovery.id}>
            <h3>{discovery.name}</h3>
            <p>Status: {discovery.status}</p>
            <p>Type: {discovery.type}</p>
            <p>Created: {new Date(discovery.created_at).toLocaleDateString()}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DiscoveryComponent;
```

## Best Practices

### Discovery Design

1. **Source Configuration**: Configure data sources with proper connection details and credentials
2. **Rule Definition**: Define discovery rules to identify specific data patterns and issues
3. **Profile Selection**: Choose appropriate profile types based on your analysis needs
4. **Pattern Detection**: Configure pattern detection algorithms for your data characteristics
5. **Filtering**: Use filters to focus discovery on relevant data subsets

### Performance Optimization

1. **Parallel Processing**: Enable parallel processing for large datasets
2. **Batch Processing**: Use appropriate batch sizes for optimal performance
3. **Timeout Configuration**: Set reasonable timeouts based on data volume
4. **Resource Limits**: Configure worker limits to prevent resource exhaustion
5. **Incremental Discovery**: Use incremental discovery for regular updates

### Error Handling

1. **Validation**: Validate all input parameters before submission
2. **Retry Logic**: Implement retry logic for transient failures
3. **Monitoring**: Monitor job progress and handle timeouts gracefully
4. **Error Recovery**: Implement error recovery mechanisms for failed discoveries
5. **Logging**: Log all discovery activities for debugging and audit

### Security

1. **Authentication**: Use secure API keys for authentication
2. **Authorization**: Implement proper access controls for discovery resources
3. **Data Protection**: Ensure sensitive data is properly protected during discovery
4. **Audit Logging**: Log all discovery activities for security audit
5. **Encryption**: Use encryption for data in transit and at rest

### Monitoring and Alerting

1. **Job Monitoring**: Monitor discovery job progress and completion
2. **Performance Metrics**: Track discovery performance and resource usage
3. **Quality Metrics**: Monitor data quality scores and trends
4. **Anomaly Detection**: Set up alerts for unusual discovery patterns
5. **Capacity Planning**: Monitor resource usage for capacity planning

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Requests per minute**: 100 requests per minute per API key
- **Concurrent jobs**: Maximum 10 concurrent discovery jobs per account
- **Job duration**: Maximum 24 hours per discovery job

## Troubleshooting

### Common Issues

1. **Authentication Errors**: Verify API key is valid and has proper permissions
2. **Validation Errors**: Check request body for required fields and valid values
3. **Timeout Errors**: Increase timeout values for large datasets
4. **Resource Errors**: Reduce parallel workers or batch sizes
5. **Connection Errors**: Verify data source connectivity and credentials

### Debug Information

Enable debug logging by including the `X-Debug` header:

```http
X-Debug: true
```

### Support

For additional support:

- **Documentation**: https://docs.example.com/discovery-api
- **API Status**: https://status.example.com
- **Support Email**: api-support@example.com
- **Community Forum**: https://community.example.com

## Future Enhancements

1. **Advanced Algorithms**: Support for machine learning-based discovery
2. **Real-time Discovery**: Real-time data discovery capabilities
3. **Collaborative Discovery**: Multi-user discovery workflows
4. **Integration APIs**: Integration with external data platforms
5. **Advanced Analytics**: Enhanced analytics and reporting features
