# Data Lineage API Documentation

## Overview

The Data Lineage API provides comprehensive data lineage tracking and impact analysis capabilities for the KYB Platform. This API enables organizations to trace data flows, understand dependencies, analyze impact, and visualize data lineage across their entire data ecosystem.

### Key Features

- **Multiple Lineage Types**: Support for data flow, transformation, dependency, impact, source, target, process, and system lineage types
- **Advanced Lineage Tracking**: Comprehensive tracking of data sources, targets, processes, and transformations
- **Impact Analysis**: Detailed impact analysis with risk assessment and recommendations
- **Visualization Support**: Graph-based lineage visualization with node positioning and edge relationships
- **Background Processing**: Asynchronous lineage analysis with progress tracking and status monitoring
- **Lineage Reporting**: Detailed lineage reports with trends, recommendations, and actionable insights

### Supported Lineage Types

- **data_flow**: Data flow between sources, processes, and targets
- **transformation**: Data transformation and processing lineage
- **dependency**: Data dependency relationships
- **impact**: Impact analysis and risk assessment
- **source**: Source system lineage
- **target**: Target system lineage
- **process**: Process and ETL lineage
- **system**: System-level lineage

### Supported Lineage Directions

- **upstream**: Track data lineage upstream to sources
- **downstream**: Track data lineage downstream to targets
- **bidirectional**: Track lineage in both directions

### Supported Lineage Statuses

- **active**: Lineage is currently active and being tracked
- **inactive**: Lineage is inactive but preserved
- **deprecated**: Lineage is deprecated and no longer maintained
- **error**: Lineage encountered an error

## Authentication

All API endpoints require authentication using API keys. Include your API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All API responses are returned in JSON format with the following structure:

```json
{
  "id": "lineage_1234567890",
  "name": "Customer Data Lineage",
  "type": "data_flow",
  "status": "active",
  "dataset": "customer_data",
  "direction": "downstream",
  "depth": 3,
  "nodes": [...],
  "edges": [...],
  "paths": [...],
  "impact": {...},
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

## API Endpoints

### 1. Create Lineage

**POST** `/lineage`

Creates and executes a data lineage analysis immediately.

#### Request Body

```json
{
  "name": "Customer Data Lineage",
  "description": "Track customer data flow from source to analytics",
  "dataset": "customer_data",
  "type": "data_flow",
  "direction": "downstream",
  "depth": 3,
  "sources": [
    {
      "id": "source_1",
      "name": "Customer Database",
      "type": "database",
      "location": "postgres://localhost:5432/customers",
      "format": "postgresql",
      "schema": {
        "table": "customers"
      },
      "connection": {
        "id": "conn_1",
        "name": "Customer DB Connection",
        "type": "postgresql",
        "protocol": "postgresql",
        "host": "localhost",
        "port": 5432,
        "database": "customers",
        "schema": "public",
        "table": "customers"
      },
      "properties": {
        "refresh_rate": "daily"
      },
      "metadata": {
        "owner": "data_team"
      }
    }
  ],
  "targets": [
    {
      "id": "target_1",
      "name": "Analytics Warehouse",
      "type": "warehouse",
      "location": "bigquery://project/dataset",
      "format": "bigquery",
      "schema": {
        "dataset": "analytics",
        "table": "customer_analytics"
      },
      "connection": {
        "id": "conn_2",
        "name": "BigQuery Connection",
        "type": "bigquery",
        "protocol": "https",
        "host": "bigquery.googleapis.com",
        "port": 443,
        "database": "project",
        "schema": "dataset",
        "table": "customer_analytics"
      },
      "properties": {
        "refresh_rate": "hourly"
      },
      "metadata": {
        "owner": "analytics_team"
      }
    }
  ],
  "processes": [
    {
      "id": "process_1",
      "name": "Data Transformation",
      "type": "etl",
      "description": "Transform customer data for analytics",
      "inputs": ["source_1"],
      "outputs": ["target_1"],
      "logic": "SELECT * FROM customers WHERE active = true",
      "parameters": {
        "filter": "active_customers"
      },
      "schedule": "0 2 * * *",
      "status": "active",
      "metadata": {
        "owner": "etl_team"
      }
    }
  ],
  "transformations": [
    {
      "id": "transform_1",
      "name": "Customer Filter",
      "type": "filter",
      "description": "Filter active customers only",
      "input_fields": ["customer_id", "name", "email", "active"],
      "output_fields": ["customer_id", "name", "email"],
      "logic": "active = true",
      "rules": [
        {
          "id": "rule_1",
          "name": "Active Customer Rule",
          "type": "filter",
          "description": "Only include active customers",
          "expression": "active = true",
          "parameters": {},
          "priority": 1,
          "enabled": true
        }
      ],
      "conditions": [
        {
          "id": "condition_1",
          "name": "Active Status Check",
          "type": "validation",
          "description": "Check if customer is active",
          "expression": "active",
          "parameters": {},
          "operator": "equals",
          "value": true
        }
      ],
      "metadata": {
        "owner": "data_team"
      }
    }
  ],
  "filters": {
    "types": ["data_flow", "transformation"],
    "statuses": ["active"],
    "date_range": {
      "start": "2024-11-19T00:00:00Z",
      "end": "2024-12-19T00:00:00Z"
    },
    "tags": ["customer", "analytics"],
    "owners": ["data_team"],
    "custom": {
      "priority": "high"
    }
  },
  "options": {
    "include_metadata": true,
    "include_schema": true,
    "include_stats": true,
    "max_depth": 5,
    "max_nodes": 100,
    "format": "json",
    "direction": "downstream",
    "custom": {
      "visualization": "graph"
    }
  },
  "metadata": {
    "priority": "high",
    "tags": ["customer", "analytics"]
  }
}
```

#### Response

```json
{
  "id": "lineage_1234567890",
  "name": "Customer Data Lineage",
  "type": "data_flow",
  "status": "active",
  "dataset": "customer_data",
  "direction": "downstream",
  "depth": 3,
  "nodes": [
    {
      "id": "source_1",
      "name": "Customer Database",
      "type": "database",
      "category": "source",
      "location": "postgres://localhost:5432/customers",
      "status": "active",
      "properties": {
        "format": "postgresql",
        "connection": {...}
      },
      "schema": {
        "table": "customers"
      },
      "stats": {
        "row_count": 1000000,
        "column_count": 50,
        "size_bytes": 1024000000,
        "last_updated": "2024-12-19T10:30:00Z",
        "refresh_rate": "daily",
        "quality": 0.95
      },
      "metadata": {
        "owner": "data_team"
      },
      "position": {
        "x": 0,
        "y": 0
      }
    }
  ],
  "edges": [
    {
      "id": "edge_source_1_process_1",
      "source": "source_1",
      "target": "process_1",
      "type": "data_flow",
      "direction": "downstream",
      "properties": {
        "flow_type": "extract",
        "frequency": "daily"
      },
      "transformations": [],
      "metadata": {}
    }
  ],
  "paths": [
    {
      "id": "path_source_1_target_1",
      "name": "Path from Customer Database to Analytics Warehouse",
      "nodes": ["source_1", "target_1"],
      "edges": ["edge_source_1_target_1"],
      "length": 2,
      "type": "data_flow",
      "properties": {
        "path_type": "direct",
        "complexity": "low"
      },
      "metadata": {}
    }
  ],
  "impact": {
    "affected_nodes": ["source_1", "target_1"],
    "affected_edges": [],
    "affected_paths": [],
    "impact_score": 0.75,
    "risk_level": "medium",
    "recommendations": [
      "Monitor data quality metrics",
      "Implement data validation checks",
      "Set up automated alerts"
    ],
    "analysis": {
      "critical_paths": 2,
      "bottlenecks": 1,
      "dependencies": 5
    },
    "metadata": {}
  },
  "summary": {
    "total_nodes": 3,
    "total_edges": 2,
    "total_paths": 1,
    "node_types": {
      "source": 1,
      "process": 1,
      "target": 1
    },
    "edge_types": {
      "data_flow": 2
    },
    "path_types": {
      "data_flow": 1
    },
    "max_depth": 3,
    "avg_path_length": 2.0,
    "complexity": "medium",
    "metrics": {
      "data_volume": "1.5TB",
      "refresh_frequency": "hourly",
      "data_quality": 0.95
    }
  },
  "metadata": {
    "priority": "high",
    "tags": ["customer", "analytics"]
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 2. Get Lineage

**GET** `/lineage?id={id}`

Retrieves details of a specific lineage analysis.

#### Parameters

- `id` (required): The lineage ID

#### Response

```json
{
  "id": "lineage_1234567890",
  "name": "Customer Data Lineage",
  "type": "data_flow",
  "status": "active",
  "dataset": "customer_data",
  "direction": "downstream",
  "depth": 3,
  "nodes": [...],
  "edges": [...],
  "paths": [...],
  "impact": {...},
  "summary": {...},
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 3. List Lineages

**GET** `/lineage`

Lists all lineage analyses.

#### Response

```json
{
  "lineages": [
    {
      "id": "lineage_1234567890",
      "name": "Customer Data Lineage",
      "type": "data_flow",
      "status": "active",
      "dataset": "customer_data",
      "direction": "downstream",
      "depth": 3,
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 4. Create Lineage Job

**POST** `/lineage/jobs`

Creates a background lineage analysis job.

#### Request Body

Same as Create Lineage endpoint.

#### Response

```json
{
  "id": "lineage_job_1234567890",
  "request_id": "req_1234567890",
  "status": "pending",
  "progress": 0,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "priority": "high",
    "tags": ["customer", "analytics"]
  }
}
```

### 5. Get Lineage Job

**GET** `/lineage/jobs?id={id}`

Retrieves the status of a background lineage job.

#### Parameters

- `id` (required): The job ID

#### Response

```json
{
  "id": "lineage_job_1234567890",
  "request_id": "req_1234567890",
  "status": "completed",
  "progress": 100,
  "result": {
    "id": "lineage_1234567890",
    "name": "Customer Data Lineage",
    "type": "data_flow",
    "status": "active",
    "dataset": "customer_data",
    "direction": "downstream",
    "depth": 3,
    "nodes": [...],
    "edges": [...],
    "paths": [...],
    "impact": {...},
    "summary": {...},
    "metadata": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:35:00Z",
  "completed_at": "2024-12-19T10:35:00Z",
  "metadata": {
    "priority": "high",
    "tags": ["customer", "analytics"]
  }
}
```

### 6. List Lineage Jobs

**GET** `/lineage/jobs`

Lists all background lineage jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "lineage_job_1234567890",
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

### 400 Bad Request

```json
{
  "error": "name is required"
}
```

### 404 Not Found

```json
{
  "error": "Lineage not found"
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

class DataLineageClient {
  constructor(apiKey, baseURL = 'https://api.kyb-platform.com') {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createLineage(lineageRequest) {
    try {
      const response = await this.client.post('/lineage', lineageRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create lineage: ${error.response?.data?.error || error.message}`);
    }
  }

  async getLineage(id) {
    try {
      const response = await this.client.get(`/lineage?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get lineage: ${error.response?.data?.error || error.message}`);
    }
  }

  async listLineages() {
    try {
      const response = await this.client.get('/lineage');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to list lineages: ${error.response?.data?.error || error.message}`);
    }
  }

  async createLineageJob(lineageRequest) {
    try {
      const response = await this.client.post('/lineage/jobs', lineageRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create lineage job: ${error.response?.data?.error || error.message}`);
    }
  }

  async getLineageJob(id) {
    try {
      const response = await this.client.get(`/lineage/jobs?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get lineage job: ${error.response?.data?.error || error.message}`);
    }
  }

  async listLineageJobs() {
    try {
      const response = await this.client.get('/lineage/jobs');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to list lineage jobs: ${error.response?.data?.error || error.message}`);
    }
  }
}

// Usage example
const client = new DataLineageClient('your-api-key');

const lineageRequest = {
  name: 'Customer Data Lineage',
  description: 'Track customer data flow from source to analytics',
  dataset: 'customer_data',
  type: 'data_flow',
  direction: 'downstream',
  depth: 3,
  sources: [
    {
      id: 'source_1',
      name: 'Customer Database',
      type: 'database',
      location: 'postgres://localhost:5432/customers',
      format: 'postgresql',
      schema: { table: 'customers' },
      connection: {
        id: 'conn_1',
        name: 'Customer DB Connection',
        type: 'postgresql',
        protocol: 'postgresql',
        host: 'localhost',
        port: 5432,
        database: 'customers',
        schema: 'public',
        table: 'customers'
      },
      properties: { refresh_rate: 'daily' },
      metadata: { owner: 'data_team' }
    }
  ],
  targets: [
    {
      id: 'target_1',
      name: 'Analytics Warehouse',
      type: 'warehouse',
      location: 'bigquery://project/dataset',
      format: 'bigquery',
      schema: { dataset: 'analytics', table: 'customer_analytics' },
      connection: {
        id: 'conn_2',
        name: 'BigQuery Connection',
        type: 'bigquery',
        protocol: 'https',
        host: 'bigquery.googleapis.com',
        port: 443,
        database: 'project',
        schema: 'dataset',
        table: 'customer_analytics'
      },
      properties: { refresh_rate: 'hourly' },
      metadata: { owner: 'analytics_team' }
    }
  ],
  processes: [
    {
      id: 'process_1',
      name: 'Data Transformation',
      type: 'etl',
      description: 'Transform customer data for analytics',
      inputs: ['source_1'],
      outputs: ['target_1'],
      logic: 'SELECT * FROM customers WHERE active = true',
      parameters: { filter: 'active_customers' },
      schedule: '0 2 * * *',
      status: 'active',
      metadata: { owner: 'etl_team' }
    }
  ],
  transformations: [],
  filters: {
    types: ['data_flow'],
    statuses: ['active'],
    date_range: {
      start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
      end: new Date().toISOString()
    },
    tags: ['customer', 'analytics'],
    owners: ['data_team'],
    custom: { priority: 'high' }
  },
  options: {
    include_metadata: true,
    include_schema: true,
    include_stats: true,
    max_depth: 5,
    max_nodes: 100,
    format: 'json',
    direction: 'downstream',
    custom: { visualization: 'graph' }
  },
  metadata: {
    priority: 'high',
    tags: ['customer', 'analytics']
  }
};

async function example() {
  try {
    // Create lineage immediately
    const lineage = await client.createLineage(lineageRequest);
    console.log('Created lineage:', lineage.id);

    // Create background job
    const job = await client.createLineageJob(lineageRequest);
    console.log('Created job:', job.id);

    // Poll for job completion
    let jobStatus = await client.getLineageJob(job.id);
    while (jobStatus.status !== 'completed' && jobStatus.status !== 'failed') {
      await new Promise(resolve => setTimeout(resolve, 2000));
      jobStatus = await client.getLineageJob(job.id);
      console.log(`Job progress: ${jobStatus.progress}%`);
    }

    if (jobStatus.status === 'completed') {
      console.log('Job completed:', jobStatus.result.id);
    } else {
      console.log('Job failed:', jobStatus.error);
    }

    // List all lineages
    const lineages = await client.listLineages();
    console.log(`Total lineages: ${lineages.total}`);

  } catch (error) {
    console.error('Error:', error.message);
  }
}

example();
```

### Python

```python
import requests
import json
from datetime import datetime, timedelta
import time

class DataLineageClient:
    def __init__(self, api_key, base_url='https://api.kyb-platform.com'):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def create_lineage(self, lineage_request):
        """Create a lineage analysis immediately."""
        try:
            response = requests.post(
                f'{self.base_url}/lineage',
                headers=self.headers,
                json=lineage_request
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to create lineage: {e}')

    def get_lineage(self, lineage_id):
        """Get lineage details by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/lineage?id={lineage_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get lineage: {e}')

    def list_lineages(self):
        """List all lineages."""
        try:
            response = requests.get(
                f'{self.base_url}/lineage',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to list lineages: {e}')

    def create_lineage_job(self, lineage_request):
        """Create a background lineage job."""
        try:
            response = requests.post(
                f'{self.base_url}/lineage/jobs',
                headers=self.headers,
                json=lineage_request
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to create lineage job: {e}')

    def get_lineage_job(self, job_id):
        """Get lineage job status by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/lineage/jobs?id={job_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get lineage job: {e}')

    def list_lineage_jobs(self):
        """List all lineage jobs."""
        try:
            response = requests.get(
                f'{self.base_url}/lineage/jobs',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to list lineage jobs: {e}')

# Usage example
def main():
    client = DataLineageClient('your-api-key')

    lineage_request = {
        'name': 'Customer Data Lineage',
        'description': 'Track customer data flow from source to analytics',
        'dataset': 'customer_data',
        'type': 'data_flow',
        'direction': 'downstream',
        'depth': 3,
        'sources': [
            {
                'id': 'source_1',
                'name': 'Customer Database',
                'type': 'database',
                'location': 'postgres://localhost:5432/customers',
                'format': 'postgresql',
                'schema': {'table': 'customers'},
                'connection': {
                    'id': 'conn_1',
                    'name': 'Customer DB Connection',
                    'type': 'postgresql',
                    'protocol': 'postgresql',
                    'host': 'localhost',
                    'port': 5432,
                    'database': 'customers',
                    'schema': 'public',
                    'table': 'customers'
                },
                'properties': {'refresh_rate': 'daily'},
                'metadata': {'owner': 'data_team'}
            }
        ],
        'targets': [
            {
                'id': 'target_1',
                'name': 'Analytics Warehouse',
                'type': 'warehouse',
                'location': 'bigquery://project/dataset',
                'format': 'bigquery',
                'schema': {'dataset': 'analytics', 'table': 'customer_analytics'},
                'connection': {
                    'id': 'conn_2',
                    'name': 'BigQuery Connection',
                    'type': 'bigquery',
                    'protocol': 'https',
                    'host': 'bigquery.googleapis.com',
                    'port': 443,
                    'database': 'project',
                    'schema': 'dataset',
                    'table': 'customer_analytics'
                },
                'properties': {'refresh_rate': 'hourly'},
                'metadata': {'owner': 'analytics_team'}
            }
        ],
        'processes': [
            {
                'id': 'process_1',
                'name': 'Data Transformation',
                'type': 'etl',
                'description': 'Transform customer data for analytics',
                'inputs': ['source_1'],
                'outputs': ['target_1'],
                'logic': 'SELECT * FROM customers WHERE active = true',
                'parameters': {'filter': 'active_customers'},
                'schedule': '0 2 * * *',
                'status': 'active',
                'metadata': {'owner': 'etl_team'}
            }
        ],
        'transformations': [],
        'filters': {
            'types': ['data_flow'],
            'statuses': ['active'],
            'date_range': {
                'start': (datetime.now() - timedelta(days=30)).isoformat(),
                'end': datetime.now().isoformat()
            },
            'tags': ['customer', 'analytics'],
            'owners': ['data_team'],
            'custom': {'priority': 'high'}
        },
        'options': {
            'include_metadata': True,
            'include_schema': True,
            'include_stats': True,
            'max_depth': 5,
            'max_nodes': 100,
            'format': 'json',
            'direction': 'downstream',
            'custom': {'visualization': 'graph'}
        },
        'metadata': {
            'priority': 'high',
            'tags': ['customer', 'analytics']
        }
    }

    try:
        # Create lineage immediately
        lineage = client.create_lineage(lineage_request)
        print(f'Created lineage: {lineage["id"]}')

        # Create background job
        job = client.create_lineage_job(lineage_request)
        print(f'Created job: {job["id"]}')

        # Poll for job completion
        job_status = client.get_lineage_job(job['id'])
        while job_status['status'] not in ['completed', 'failed']:
            time.sleep(2)
            job_status = client.get_lineage_job(job['id'])
            print(f'Job progress: {job_status["progress"]}%')

        if job_status['status'] == 'completed':
            print(f'Job completed: {job_status["result"]["id"]}')
        else:
            print(f'Job failed: {job_status["error"]}')

        # List all lineages
        lineages = client.list_lineages()
        print(f'Total lineages: {lineages["total"]}')

    except Exception as e:
        print(f'Error: {e}')

if __name__ == '__main__':
    main()
```

### React/TypeScript

```typescript
interface LineageSource {
  id: string;
  name: string;
  type: string;
  location: string;
  format: string;
  schema: Record<string, any>;
  connection: {
    id: string;
    name: string;
    type: string;
    protocol: string;
    host: string;
    port: number;
    database: string;
    schema: string;
    table: string;
  };
  properties: Record<string, any>;
  metadata: Record<string, any>;
}

interface LineageTarget {
  id: string;
  name: string;
  type: string;
  location: string;
  format: string;
  schema: Record<string, any>;
  connection: {
    id: string;
    name: string;
    type: string;
    protocol: string;
    host: string;
    port: number;
    database: string;
    schema: string;
    table: string;
  };
  properties: Record<string, any>;
  metadata: Record<string, any>;
}

interface LineageProcess {
  id: string;
  name: string;
  type: string;
  description: string;
  inputs: string[];
  outputs: string[];
  logic: string;
  parameters: Record<string, any>;
  schedule: string;
  status: string;
  metadata: Record<string, any>;
}

interface LineageRequest {
  name: string;
  description: string;
  dataset: string;
  type: 'data_flow' | 'transformation' | 'dependency' | 'impact' | 'source' | 'target' | 'process' | 'system';
  direction: 'upstream' | 'downstream' | 'bidirectional';
  depth: number;
  sources: LineageSource[];
  targets: LineageTarget[];
  processes: LineageProcess[];
  transformations: any[];
  filters: {
    types: string[];
    statuses: string[];
    date_range: {
      start: string;
      end: string;
    };
    tags: string[];
    owners: string[];
    custom: Record<string, any>;
  };
  options: {
    include_metadata: boolean;
    include_schema: boolean;
    include_stats: boolean;
    max_depth: number;
    max_nodes: number;
    format: string;
    direction: 'upstream' | 'downstream' | 'bidirectional';
    custom: Record<string, any>;
  };
  metadata: Record<string, any>;
}

interface LineageResponse {
  id: string;
  name: string;
  type: string;
  status: string;
  dataset: string;
  direction: string;
  depth: number;
  nodes: any[];
  edges: any[];
  paths: any[];
  impact: {
    affected_nodes: string[];
    affected_edges: string[];
    affected_paths: string[];
    impact_score: number;
    risk_level: string;
    recommendations: string[];
    analysis: Record<string, any>;
    metadata: Record<string, any>;
  };
  summary: {
    total_nodes: number;
    total_edges: number;
    total_paths: number;
    node_types: Record<string, number>;
    edge_types: Record<string, number>;
    path_types: Record<string, number>;
    max_depth: number;
    avg_path_length: number;
    complexity: string;
    metrics: Record<string, any>;
  };
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

interface LineageJob {
  id: string;
  request_id: string;
  status: string;
  progress: number;
  result?: LineageResponse;
  error?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  metadata: Record<string, any>;
}

class DataLineageClient {
  private apiKey: string;
  private baseURL: string;

  constructor(apiKey: string, baseURL: string = 'https://api.kyb-platform.com') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
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
      throw new Error(`API request failed: ${error.error || response.statusText}`);
    }

    return response.json();
  }

  async createLineage(lineageRequest: LineageRequest): Promise<LineageResponse> {
    return this.request<LineageResponse>('/lineage', {
      method: 'POST',
      body: JSON.stringify(lineageRequest),
    });
  }

  async getLineage(id: string): Promise<LineageResponse> {
    return this.request<LineageResponse>(`/lineage?id=${id}`);
  }

  async listLineages(): Promise<{ lineages: LineageResponse[]; total: number }> {
    return this.request<{ lineages: LineageResponse[]; total: number }>('/lineage');
  }

  async createLineageJob(lineageRequest: LineageRequest): Promise<LineageJob> {
    return this.request<LineageJob>('/lineage/jobs', {
      method: 'POST',
      body: JSON.stringify(lineageRequest),
    });
  }

  async getLineageJob(id: string): Promise<LineageJob> {
    return this.request<LineageJob>(`/lineage/jobs?id=${id}`);
  }

  async listLineageJobs(): Promise<{ jobs: LineageJob[]; total: number }> {
    return this.request<{ jobs: LineageJob[]; total: number }>('/lineage/jobs');
  }
}

// React component example
import React, { useState, useEffect } from 'react';

interface DataLineageProps {
  apiKey: string;
}

const DataLineage: React.FC<DataLineageProps> = ({ apiKey }) => {
  const [client] = useState(() => new DataLineageClient(apiKey));
  const [lineages, setLineages] = useState<LineageResponse[]>([]);
  const [jobs, setJobs] = useState<LineageJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadLineages();
    loadJobs();
  }, []);

  const loadLineages = async () => {
    try {
      setLoading(true);
      const response = await client.listLineages();
      setLineages(response.lineages);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load lineages');
    } finally {
      setLoading(false);
    }
  };

  const loadJobs = async () => {
    try {
      const response = await client.listLineageJobs();
      setJobs(response.jobs);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load jobs');
    }
  };

  const createLineage = async (lineageRequest: LineageRequest) => {
    try {
      setLoading(true);
      const lineage = await client.createLineage(lineageRequest);
      setLineages(prev => [...prev, lineage]);
      return lineage;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create lineage');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const createLineageJob = async (lineageRequest: LineageRequest) => {
    try {
      const job = await client.createLineageJob(lineageRequest);
      setJobs(prev => [...prev, job]);
      return job;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create lineage job');
      throw err;
    }
  };

  const pollJobStatus = async (jobId: string) => {
    const poll = async (): Promise<LineageJob> => {
      const job = await client.getLineageJob(jobId);
      if (job.status === 'completed' || job.status === 'failed') {
        return job;
      }
      await new Promise(resolve => setTimeout(resolve, 2000));
      return poll();
    };
    return poll();
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>Data Lineage</h2>
      
      <h3>Lineages ({lineages.length})</h3>
      <div>
        {lineages.map(lineage => (
          <div key={lineage.id}>
            <h4>{lineage.name}</h4>
            <p>Type: {lineage.type}</p>
            <p>Status: {lineage.status}</p>
            <p>Dataset: {lineage.dataset}</p>
            <p>Nodes: {lineage.summary.total_nodes}</p>
            <p>Edges: {lineage.summary.total_edges}</p>
            <p>Paths: {lineage.summary.total_paths}</p>
          </div>
        ))}
      </div>

      <h3>Jobs ({jobs.length})</h3>
      <div>
        {jobs.map(job => (
          <div key={job.id}>
            <h4>Job {job.id}</h4>
            <p>Status: {job.status}</p>
            <p>Progress: {job.progress}%</p>
            <p>Created: {new Date(job.created_at).toLocaleString()}</p>
            {job.completed_at && (
              <p>Completed: {new Date(job.completed_at).toLocaleString()}</p>
            )}
            {job.error && <p>Error: {job.error}</p>}
          </div>
        ))}
      </div>
    </div>
  );
};

export default DataLineage;
```

## Best Practices

### Lineage Design

1. **Define Clear Data Sources**: Clearly identify and document all data sources with proper connection details
2. **Map Data Flows**: Create comprehensive mappings of data flows from sources to targets
3. **Document Transformations**: Document all data transformations with clear logic and rules
4. **Set Appropriate Depth**: Choose appropriate lineage depth based on complexity and requirements
5. **Use Meaningful Names**: Use descriptive names for lineages, sources, targets, and processes

### Performance Optimization

1. **Limit Depth**: Use reasonable depth limits to avoid excessive processing
2. **Filter Results**: Use filters to focus on relevant lineage components
3. **Background Jobs**: Use background jobs for complex lineage analysis
4. **Caching**: Implement caching for frequently accessed lineage data
5. **Pagination**: Use pagination for large lineage datasets

### Error Handling

1. **Validate Inputs**: Always validate lineage request parameters
2. **Handle Timeouts**: Implement proper timeout handling for long-running operations
3. **Retry Logic**: Implement retry logic for transient failures
4. **Error Logging**: Log errors with sufficient context for debugging
5. **Graceful Degradation**: Handle partial failures gracefully

### Security

1. **API Key Management**: Securely manage and rotate API keys
2. **Input Validation**: Validate all input parameters to prevent injection attacks
3. **Access Control**: Implement proper access control for lineage data
4. **Audit Logging**: Log all lineage operations for audit purposes
5. **Data Encryption**: Encrypt sensitive lineage data in transit and at rest

### Monitoring and Alerting

1. **Job Monitoring**: Monitor background job status and progress
2. **Performance Metrics**: Track lineage analysis performance metrics
3. **Error Alerts**: Set up alerts for lineage analysis failures
4. **Usage Monitoring**: Monitor API usage and rate limits
5. **Health Checks**: Implement health checks for lineage services

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Requests per minute**: 100 requests per minute per API key
- **Concurrent jobs**: Maximum 10 concurrent background jobs per API key
- **Request size**: Maximum 10MB per request

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

## Troubleshooting

### Common Issues

1. **Invalid Request Format**
   - Ensure all required fields are provided
   - Validate JSON format and data types
   - Check enum values for type, direction, and status fields

2. **Authentication Errors**
   - Verify API key is valid and not expired
   - Check Authorization header format
   - Ensure API key has required permissions

3. **Job Timeouts**
   - Increase timeout settings for complex lineage analysis
   - Use background jobs for long-running operations
   - Monitor job progress and implement retry logic

4. **Memory Issues**
   - Reduce lineage depth for complex data flows
   - Use filters to limit scope of analysis
   - Implement pagination for large datasets

### Debug Information

Enable debug logging by including the `X-Debug` header:

```
X-Debug: true
```

Debug responses include additional information:

```json
{
  "id": "lineage_1234567890",
  "name": "Customer Data Lineage",
  "type": "data_flow",
  "status": "active",
  "debug": {
    "processing_time_ms": 1250,
    "nodes_processed": 15,
    "edges_processed": 28,
    "paths_found": 12,
    "memory_usage_mb": 45.2
  }
}
```

### Support

For additional support:

- **Documentation**: [https://docs.kyb-platform.com/api/lineage](https://docs.kyb-platform.com/api/lineage)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

## Future Enhancements

### Planned Features

1. **Real-time Lineage**: Real-time lineage tracking and updates
2. **Advanced Visualization**: Enhanced graph visualization with interactive features
3. **Lineage Templates**: Pre-built lineage templates for common patterns
4. **Integration APIs**: Direct integration with popular data platforms
5. **Machine Learning**: ML-powered lineage discovery and impact prediction

### API Versioning

The API follows semantic versioning. Current version: `v3.0.0`

- **v3.0.0**: Current stable version
- **v2.x.x**: Deprecated, will be removed in 2025
- **v1.x.x**: Deprecated, will be removed in 2024

### Migration Guide

For users migrating from v2.x.x to v3.0.0:

1. Update request/response models to match v3.0.0 format
2. Replace deprecated endpoints with new equivalents
3. Update authentication to use Bearer token format
4. Review and update error handling for new error codes
5. Test thoroughly in staging environment before production deployment

---

**API Version**: 3.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
