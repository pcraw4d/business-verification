# Data Catalog API Documentation

## Overview

The Data Catalog API provides comprehensive data catalog capabilities for the KYB Platform. This API enables organizations to discover, organize, and manage their data assets across the entire data ecosystem, providing metadata management, asset discovery, data lineage tracking, and governance capabilities.

### Key Features

- **Comprehensive Asset Management**: Support for databases, tables, views, APIs, files, streams, models, reports, dashboards, and metrics
- **Advanced Metadata Management**: Rich metadata support with schemas, lineage, quality metrics, and governance information  
- **Asset Discovery**: Automated asset discovery with configurable scanning and cataloging
- **Data Quality Monitoring**: Comprehensive data quality assessment and monitoring
- **Usage Analytics**: Detailed usage patterns, performance metrics, and user analytics
- **Governance and Compliance**: Policy management, access controls, and compliance tracking
- **Background Processing**: Asynchronous catalog creation and processing with progress tracking

### Supported Catalog Types

- **database**: Database catalogs for relational and NoSQL databases
- **table**: Table-level catalogs for structured data
- **view**: View catalogs for database views and virtual tables
- **api**: API catalogs for REST, GraphQL, and other API endpoints
- **file**: File catalogs for data files and documents
- **stream**: Stream catalogs for real-time data streams
- **model**: Model catalogs for machine learning and analytical models
- **report**: Report catalogs for business reports and analytics
- **dashboard**: Dashboard catalogs for BI dashboards and visualizations
- **metric**: Metric catalogs for KPIs and business metrics

### Supported Asset Types

- **dataset**: Data collections and datasets
- **schema**: Schema definitions and structures
- **column**: Individual data columns and fields
- **metric**: Business metrics and KPIs
- **dimension**: Data dimensions for analysis
- **kpi**: Key Performance Indicators
- **report**: Business reports and analytics
- **dashboard**: Interactive dashboards
- **visualization**: Data visualizations and charts
- **model**: Analytical and ML models

### Supported Catalog Statuses

- **active**: Catalog is active and available
- **inactive**: Catalog is inactive but preserved
- **deprecated**: Catalog is deprecated and scheduled for removal
- **draft**: Catalog is in draft mode and not yet published

## Authentication

All API endpoints require authentication using API keys. Include your API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All API responses are returned in JSON format with the following structure:

```json
{
  "id": "catalog_1234567890",
  "name": "Enterprise Data Catalog",
  "type": "database",
  "status": "active",
  "category": "enterprise",
  "assets": [...],
  "collections": [...],
  "schemas": [...],
  "connections": [...],
  "summary": {...},
  "statistics": {...},
  "health": {...},
  "tags": [...],
  "owners": [...],
  "stewards": [...],
  "domains": [...],
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

## API Endpoints

### 1. Create Catalog

**POST** `/catalog`

Creates and processes a data catalog immediately.

#### Request Body

```json
{
  "name": "Enterprise Data Catalog",
  "description": "Comprehensive enterprise data catalog",
  "type": "database",
  "category": "enterprise",
  "assets": [
    {
      "id": "asset_1",
      "name": "Customer Database",
      "type": "dataset",
      "description": "Main customer database",
      "location": "postgres://localhost:5432/customers",
      "format": "postgresql",
      "size": 1024000000,
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
          },
          {
            "name": "email",
            "type": "varchar",
            "description": "Customer email",
            "nullable": true,
            "length": 255,
            "unique": true
          }
        ],
        "constraints": [
          {
            "name": "pk_customer",
            "type": "primary_key",
            "columns": ["customer_id"],
            "enabled": true
          }
        ],
        "indexes": [
          {
            "name": "idx_customer_email",
            "type": "btree",
            "columns": ["email"],
            "unique": true
          }
        ]
      },
      "connection": {
        "id": "conn_1",
        "name": "Customer DB Connection",
        "type": "postgresql",
        "protocol": "postgresql",
        "host": "localhost",
        "port": 5432,
        "database": "customers",
        "schema": "public"
      },
      "classification": "confidential",
      "sensitivity": "high",
      "tags": ["customer", "pii", "production"]
    }
  ],
  "collections": [
    {
      "id": "collection_1",
      "name": "Customer Data Collection",
      "description": "Collection of customer-related datasets",
      "type": "business",
      "assets": ["asset_1"],
      "owner": "data_team",
      "tags": ["customer", "core"]
    }
  ],
  "schemas": [
    {
      "id": "schema_1",
      "name": "Customer Schema",
      "version": "1.0",
      "description": "Customer data schema definition",
      "type": "json_schema",
      "content": {
        "type": "object",
        "properties": {
          "customer_id": {
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "email": {
            "type": "string",
            "format": "email"
          }
        }
      },
      "assets": ["asset_1"]
    }
  ],
  "connections": [
    {
      "id": "conn_1",
      "name": "Customer DB Connection",
      "type": "postgresql",
      "description": "Connection to customer database",
      "protocol": "postgresql",
      "host": "localhost",
      "port": 5432,
      "database": "customers",
      "schema": "public",
      "credentials": {
        "username": "app_user",
        "auth_type": "password"
      },
      "properties": {
        "pool_size": 10,
        "timeout": 30
      },
      "status": "active"
    }
  ],
  "tags": ["enterprise", "production", "customer"],
  "owners": ["data_team", "engineering"],
  "stewards": ["data_steward_1", "data_steward_2"],
  "domains": ["customer", "finance", "operations"],
  "options": {
    "auto_discovery": true,
    "include_metadata": true,
    "include_schema": true,
    "include_lineage": true,
    "include_usage": true,
    "include_quality": true,
    "scan_frequency": "daily",
    "notify_changes": true
  },
  "metadata": {
    "priority": "high",
    "environment": "production",
    "region": "us-east-1"
  }
}
```

#### Response

```json
{
  "id": "catalog_1234567890",
  "name": "Enterprise Data Catalog",
  "type": "database",
  "status": "active",
  "category": "enterprise",
  "assets": [
    {
      "id": "asset_1",
      "name": "Customer Database",
      "type": "dataset",
      "description": "Main customer database",
      "location": "postgres://localhost:5432/customers",
      "format": "postgresql",
      "size": 1024000000,
      "schema": {...},
      "connection": {...},
      "classification": "confidential",
      "sensitivity": "high",
      "quality": {
        "score": 0.85,
        "completeness": 0.90,
        "accuracy": 0.85,
        "consistency": 0.88,
        "validity": 0.92,
        "timeliness": 0.75,
        "uniqueness": 0.95,
        "integrity": 0.88,
        "issues": [],
        "last_assessment": "2024-12-19T10:30:00Z",
        "next_assessment": "2024-12-20T10:30:00Z"
      },
      "lineage": {
        "upstream": [],
        "downstream": [],
        "jobs": [],
        "processes": [],
        "last_update": "2024-12-19T10:30:00Z"
      },
      "usage": {
        "access_count": 150,
        "query_count": 450,
        "download_count": 25,
        "users": [],
        "applications": [],
        "patterns": [],
        "performance": {
          "avg_response_time": 125.5,
          "max_response_time": 500.0,
          "min_response_time": 50.0,
          "throughput_qps": 25.5,
          "error_rate": 0.02,
          "availability_pct": 99.5
        },
        "last_accessed": "2024-12-19T08:30:00Z",
        "popularity_rank": 5
      },
      "governance": {
        "owner": "data_team",
        "steward": "data_steward_1",
        "custodian": "database_admin",
        "domain": "customer",
        "policies": [],
        "compliance": [],
        "retention": {
          "policy": "customer_data_retention",
          "period": "7_years",
          "action": "archive",
          "schedule": "annual"
        },
        "access": {
          "level": "restricted",
          "groups": ["data_team", "analytics"],
          "users": [],
          "roles": ["data_analyst", "data_scientist"],
          "permissions": ["read", "query"]
        }
      },
      "tags": ["customer", "pii", "production"],
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "collections": [...],
  "schemas": [...],
  "connections": [...],
  "summary": {
    "total_assets": 1,
    "total_collections": 1,
    "total_schemas": 1,
    "total_connections": 1,
    "asset_types": {
      "dataset": 1
    },
    "asset_statuses": {
      "active": 1
    },
    "data_volume": "2.5TB",
    "last_update": "2024-12-19T10:30:00Z",
    "coverage": 0.85,
    "completeness": 0.90,
    "metrics": {
      "discovery_rate": 0.75,
      "cataloged_percentage": 0.85,
      "quality_score": 0.88
    }
  },
  "statistics": {
    "access_stats": {
      "total_access": 15000,
      "unique_users": 85,
      "popular_assets": ["customer_data", "sales_metrics", "product_catalog"],
      "access_patterns": ["batch", "streaming", "api"],
      "peak_hours": [9, 10, 14, 15]
    },
    "quality_stats": {
      "overall_score": 0.85,
      "passing_assets": 8,
      "failing_assets": 2,
      "issue_types": {
        "completeness": 3,
        "accuracy": 1,
        "consistency": 2
      },
      "trend_direction": "improving"
    },
    "lineage_stats": {
      "tracked_assets": 10,
      "lineage_paths": 25,
      "orphan_assets": 3,
      "complexity_score": 0.65
    },
    "governance_stats": {
      "managed_assets": 10,
      "policy_violations": 2,
      "compliance_score": 0.92,
      "approvals_pending": 1
    },
    "performance_stats": {
      "avg_response_time": 125.5,
      "total_queries": 45000,
      "error_rate": 0.02,
      "availability": 99.8
    },
    "trends": [
      {
        "metric": "usage",
        "period": "daily",
        "values": [100, 120, 110, 140, 135, 155, 150],
        "timestamps": [...],
        "direction": "increasing",
        "change": 0.25,
        "significance": "moderate"
      }
    ]
  },
  "health": {
    "overall_status": "healthy",
    "component_health": [
      {
        "component": "metadata",
        "status": "healthy",
        "score": 0.92,
        "issues": [],
        "last_check": "2024-12-19T10:30:00Z"
      },
      {
        "component": "discovery",
        "status": "healthy",
        "score": 0.88,
        "issues": [],
        "last_check": "2024-12-19T10:30:00Z"
      },
      {
        "component": "lineage",
        "status": "warning",
        "score": 0.75,
        "issues": ["Some assets missing lineage information"],
        "last_check": "2024-12-19T10:30:00Z"
      }
    ],
    "issues": [
      {
        "id": "issue_1",
        "type": "metadata",
        "severity": "low",
        "component": "lineage",
        "description": "Some assets missing upstream lineage information",
        "impact": "Reduced lineage visibility",
        "resolution": "Configure lineage discovery for missing assets",
        "detected_at": "2024-12-19T08:30:00Z",
        "status": "open"
      }
    ],
    "recommendations": [
      "Enable auto-discovery for new data sources",
      "Review and update metadata for orphaned assets",
      "Configure quality monitoring for critical assets"
    ],
    "last_check": "2024-12-19T10:30:00Z",
    "next_check": "2024-12-19T14:30:00Z"
  },
  "tags": ["enterprise", "production", "customer"],
  "owners": ["data_team", "engineering"],
  "stewards": ["data_steward_1", "data_steward_2"],
  "domains": ["customer", "finance", "operations"],
  "metadata": {
    "priority": "high",
    "environment": "production",
    "region": "us-east-1"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 2. Get Catalog

**GET** `/catalog?id={id}`

Retrieves details of a specific data catalog.

#### Parameters

- `id` (required): The catalog ID

#### Response

```json
{
  "id": "catalog_1234567890",
  "name": "Enterprise Data Catalog",
  "type": "database",
  "status": "active",
  "category": "enterprise",
  "assets": [...],
  "collections": [...],
  "schemas": [...],
  "connections": [...],
  "summary": {...},
  "statistics": {...},
  "health": {...},
  "tags": [...],
  "owners": [...],
  "stewards": [...],
  "domains": [...],
  "metadata": {...},
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

### 3. List Catalogs

**GET** `/catalog`

Lists all data catalogs.

#### Response

```json
{
  "catalogs": [
    {
      "id": "catalog_1234567890",
      "name": "Enterprise Data Catalog",
      "type": "database",
      "status": "active",
      "category": "enterprise",
      "summary": {...},
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### 4. Create Catalog Job

**POST** `/catalog/jobs`

Creates a background catalog processing job.

#### Request Body

Same as Create Catalog endpoint.

#### Response

```json
{
  "id": "catalog_job_1234567890",
  "request_id": "req_1234567890",
  "type": "catalog_creation",
  "status": "pending",
  "progress": 0,
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "metadata": {
    "priority": "high",
    "environment": "production"
  }
}
```

### 5. Get Catalog Job

**GET** `/catalog/jobs?id={id}`

Retrieves the status of a background catalog job.

#### Parameters

- `id` (required): The job ID

#### Response

```json
{
  "id": "catalog_job_1234567890",
  "request_id": "req_1234567890",
  "type": "catalog_creation",
  "status": "completed",
  "progress": 100,
  "result": {
    "id": "catalog_1234567890",
    "name": "Enterprise Data Catalog",
    "type": "database",
    "status": "active",
    "category": "enterprise",
    "assets": [...],
    "collections": [...],
    "schemas": [...],
    "connections": [...],
    "summary": {...},
    "statistics": {...},
    "health": {...},
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:35:00Z"
  },
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:35:00Z",
  "completed_at": "2024-12-19T10:35:00Z",
  "metadata": {
    "priority": "high",
    "environment": "production"
  }
}
```

### 6. List Catalog Jobs

**GET** `/catalog/jobs`

Lists all background catalog jobs.

#### Response

```json
{
  "jobs": [
    {
      "id": "catalog_job_1234567890",
      "request_id": "req_1234567890",
      "type": "catalog_creation",
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
  "error": "Catalog not found"
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

class DataCatalogClient {
  constructor(apiKey, baseURL = 'https://api.kyb-platform.com') {
    this.client = axios.create({
      baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async createCatalog(catalogRequest) {
    try {
      const response = await this.client.post('/catalog', catalogRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create catalog: ${error.response?.data?.error || error.message}`);
    }
  }

  async getCatalog(id) {
    try {
      const response = await this.client.get(`/catalog?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get catalog: ${error.response?.data?.error || error.message}`);
    }
  }

  async listCatalogs() {
    try {
      const response = await this.client.get('/catalog');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to list catalogs: ${error.response?.data?.error || error.message}`);
    }
  }

  async createCatalogJob(catalogRequest) {
    try {
      const response = await this.client.post('/catalog/jobs', catalogRequest);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to create catalog job: ${error.response?.data?.error || error.message}`);
    }
  }

  async getCatalogJob(id) {
    try {
      const response = await this.client.get(`/catalog/jobs?id=${id}`);
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get catalog job: ${error.response?.data?.error || error.message}`);
    }
  }

  async listCatalogJobs() {
    try {
      const response = await this.client.get('/catalog/jobs');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to list catalog jobs: ${error.response?.data?.error || error.message}`);
    }
  }
}

// Usage example
const client = new DataCatalogClient('your-api-key');

const catalogRequest = {
  name: 'Enterprise Data Catalog',
  description: 'Comprehensive enterprise data catalog',
  type: 'database',
  category: 'enterprise',
  assets: [
    {
      id: 'asset_1',
      name: 'Customer Database',
      type: 'dataset',
      description: 'Main customer database',
      location: 'postgres://localhost:5432/customers',
      format: 'postgresql',
      size: 1024000000,
      schema: {
        type: 'relational',
        version: '1.0',
        columns: [
          {
            name: 'customer_id',
            type: 'integer',
            description: 'Unique customer identifier',
            nullable: false,
            primary_key: true
          }
        ]
      },
      connection: {
        id: 'conn_1',
        name: 'Customer DB Connection',
        type: 'postgresql',
        protocol: 'postgresql',
        host: 'localhost',
        port: 5432,
        database: 'customers',
        schema: 'public'
      },
      tags: ['customer', 'pii', 'production']
    }
  ],
  collections: [],
  schemas: [],
  connections: [],
  tags: ['enterprise', 'production'],
  owners: ['data_team'],
  stewards: ['data_steward'],
  domains: ['customer'],
  options: {
    auto_discovery: true,
    include_metadata: true,
    include_schema: true,
    include_lineage: true,
    include_usage: true,
    include_quality: true,
    scan_frequency: 'daily',
    notify_changes: true
  },
  metadata: {
    priority: 'high',
    environment: 'production'
  }
};

async function example() {
  try {
    // Create catalog immediately
    const catalog = await client.createCatalog(catalogRequest);
    console.log('Created catalog:', catalog.id);

    // Create background job
    const job = await client.createCatalogJob(catalogRequest);
    console.log('Created job:', job.id);

    // Poll for job completion
    let jobStatus = await client.getCatalogJob(job.id);
    while (jobStatus.status !== 'completed' && jobStatus.status !== 'failed') {
      await new Promise(resolve => setTimeout(resolve, 2000));
      jobStatus = await client.getCatalogJob(job.id);
      console.log(`Job progress: ${jobStatus.progress}%`);
    }

    if (jobStatus.status === 'completed') {
      console.log('Job completed:', jobStatus.result.id);
    } else {
      console.log('Job failed:', jobStatus.error);
    }

    // List all catalogs
    const catalogs = await client.listCatalogs();
    console.log(`Total catalogs: ${catalogs.total}`);

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
import time

class DataCatalogClient:
    def __init__(self, api_key, base_url='https://api.kyb-platform.com'):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        }

    def create_catalog(self, catalog_request):
        """Create a catalog immediately."""
        try:
            response = requests.post(
                f'{self.base_url}/catalog',
                headers=self.headers,
                json=catalog_request
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to create catalog: {e}')

    def get_catalog(self, catalog_id):
        """Get catalog details by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/catalog?id={catalog_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get catalog: {e}')

    def list_catalogs(self):
        """List all catalogs."""
        try:
            response = requests.get(
                f'{self.base_url}/catalog',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to list catalogs: {e}')

    def create_catalog_job(self, catalog_request):
        """Create a background catalog job."""
        try:
            response = requests.post(
                f'{self.base_url}/catalog/jobs',
                headers=self.headers,
                json=catalog_request
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to create catalog job: {e}')

    def get_catalog_job(self, job_id):
        """Get catalog job status by ID."""
        try:
            response = requests.get(
                f'{self.base_url}/catalog/jobs?id={job_id}',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to get catalog job: {e}')

    def list_catalog_jobs(self):
        """List all catalog jobs."""
        try:
            response = requests.get(
                f'{self.base_url}/catalog/jobs',
                headers=self.headers
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'Failed to list catalog jobs: {e}')

# Usage example
def main():
    client = DataCatalogClient('your-api-key')

    catalog_request = {
        'name': 'Enterprise Data Catalog',
        'description': 'Comprehensive enterprise data catalog',
        'type': 'database',
        'category': 'enterprise',
        'assets': [
            {
                'id': 'asset_1',
                'name': 'Customer Database',
                'type': 'dataset',
                'description': 'Main customer database',
                'location': 'postgres://localhost:5432/customers',
                'format': 'postgresql',
                'size': 1024000000,
                'schema': {
                    'type': 'relational',
                    'version': '1.0',
                    'columns': [
                        {
                            'name': 'customer_id',
                            'type': 'integer',
                            'description': 'Unique customer identifier',
                            'nullable': False,
                            'primary_key': True
                        }
                    ]
                },
                'connection': {
                    'id': 'conn_1',
                    'name': 'Customer DB Connection',
                    'type': 'postgresql',
                    'protocol': 'postgresql',
                    'host': 'localhost',
                    'port': 5432,
                    'database': 'customers',
                    'schema': 'public'
                },
                'tags': ['customer', 'pii', 'production']
            }
        ],
        'collections': [],
        'schemas': [],
        'connections': [],
        'tags': ['enterprise', 'production'],
        'owners': ['data_team'],
        'stewards': ['data_steward'],
        'domains': ['customer'],
        'options': {
            'auto_discovery': True,
            'include_metadata': True,
            'include_schema': True,
            'include_lineage': True,
            'include_usage': True,
            'include_quality': True,
            'scan_frequency': 'daily',
            'notify_changes': True
        },
        'metadata': {
            'priority': 'high',
            'environment': 'production'
        }
    }

    try:
        # Create catalog immediately
        catalog = client.create_catalog(catalog_request)
        print(f'Created catalog: {catalog["id"]}')

        # Create background job
        job = client.create_catalog_job(catalog_request)
        print(f'Created job: {job["id"]}')

        # Poll for job completion
        job_status = client.get_catalog_job(job['id'])
        while job_status['status'] not in ['completed', 'failed']:
            time.sleep(2)
            job_status = client.get_catalog_job(job['id'])
            print(f'Job progress: {job_status["progress"]}%')

        if job_status['status'] == 'completed':
            print(f'Job completed: {job_status["result"]["id"]}')
        else:
            print(f'Job failed: {job_status["error"]}')

        # List all catalogs
        catalogs = client.list_catalogs()
        print(f'Total catalogs: {catalogs["total"]}')

    except Exception as e:
        print(f'Error: {e}')

if __name__ == '__main__':
    main()
```

### React/TypeScript

```typescript
interface CatalogAsset {
  id: string;
  name: string;
  type: string;
  description: string;
  location: string;
  format: string;
  size: number;
  schema: {
    type: string;
    version: string;
    columns: {
      name: string;
      type: string;
      description: string;
      nullable: boolean;
      primary_key?: boolean;
      length?: number;
      unique?: boolean;
    }[];
  };
  connection: {
    id: string;
    name: string;
    type: string;
    protocol: string;
    host: string;
    port: number;
    database: string;
    schema: string;
  };
  tags: string[];
}

interface CatalogRequest {
  name: string;
  description: string;
  type: string;
  category: string;
  assets: CatalogAsset[];
  collections: any[];
  schemas: any[];
  connections: any[];
  tags: string[];
  owners: string[];
  stewards: string[];
  domains: string[];
  options: {
    auto_discovery: boolean;
    include_metadata: boolean;
    include_schema: boolean;
    include_lineage: boolean;
    include_usage: boolean;
    include_quality: boolean;
    scan_frequency: string;
    notify_changes: boolean;
  };
  metadata: Record<string, any>;
}

interface CatalogResponse {
  id: string;
  name: string;
  type: string;
  status: string;
  category: string;
  assets: CatalogAsset[];
  collections: any[];
  schemas: any[];
  connections: any[];
  summary: {
    total_assets: number;
    total_collections: number;
    total_schemas: number;
    total_connections: number;
    asset_types: Record<string, number>;
    asset_statuses: Record<string, number>;
    data_volume: string;
    coverage: number;
    completeness: number;
  };
  statistics: any;
  health: any;
  tags: string[];
  owners: string[];
  stewards: string[];
  domains: string[];
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

interface CatalogJob {
  id: string;
  request_id: string;
  type: string;
  status: string;
  progress: number;
  result?: CatalogResponse;
  error?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  metadata: Record<string, any>;
}

class DataCatalogClient {
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

  async createCatalog(catalogRequest: CatalogRequest): Promise<CatalogResponse> {
    return this.request<CatalogResponse>('/catalog', {
      method: 'POST',
      body: JSON.stringify(catalogRequest),
    });
  }

  async getCatalog(id: string): Promise<CatalogResponse> {
    return this.request<CatalogResponse>(`/catalog?id=${id}`);
  }

  async listCatalogs(): Promise<{ catalogs: CatalogResponse[]; total: number }> {
    return this.request<{ catalogs: CatalogResponse[]; total: number }>('/catalog');
  }

  async createCatalogJob(catalogRequest: CatalogRequest): Promise<CatalogJob> {
    return this.request<CatalogJob>('/catalog/jobs', {
      method: 'POST',
      body: JSON.stringify(catalogRequest),
    });
  }

  async getCatalogJob(id: string): Promise<CatalogJob> {
    return this.request<CatalogJob>(`/catalog/jobs?id=${id}`);
  }

  async listCatalogJobs(): Promise<{ jobs: CatalogJob[]; total: number }> {
    return this.request<{ jobs: CatalogJob[]; total: number }>('/catalog/jobs');
  }
}

// React component example
import React, { useState, useEffect } from 'react';

interface DataCatalogProps {
  apiKey: string;
}

const DataCatalog: React.FC<DataCatalogProps> = ({ apiKey }) => {
  const [client] = useState(() => new DataCatalogClient(apiKey));
  const [catalogs, setCatalogs] = useState<CatalogResponse[]>([]);
  const [jobs, setJobs] = useState<CatalogJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCatalogs();
    loadJobs();
  }, []);

  const loadCatalogs = async () => {
    try {
      setLoading(true);
      const response = await client.listCatalogs();
      setCatalogs(response.catalogs);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load catalogs');
    } finally {
      setLoading(false);
    }
  };

  const loadJobs = async () => {
    try {
      const response = await client.listCatalogJobs();
      setJobs(response.jobs);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load jobs');
    }
  };

  const createCatalog = async (catalogRequest: CatalogRequest) => {
    try {
      setLoading(true);
      const catalog = await client.createCatalog(catalogRequest);
      setCatalogs(prev => [...prev, catalog]);
      return catalog;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create catalog');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const createCatalogJob = async (catalogRequest: CatalogRequest) => {
    try {
      const job = await client.createCatalogJob(catalogRequest);
      setJobs(prev => [...prev, job]);
      return job;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create catalog job');
      throw err;
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>Data Catalog</h2>
      
      <h3>Catalogs ({catalogs.length})</h3>
      <div>
        {catalogs.map(catalog => (
          <div key={catalog.id}>
            <h4>{catalog.name}</h4>
            <p>Type: {catalog.type}</p>
            <p>Status: {catalog.status}</p>
            <p>Category: {catalog.category}</p>
            <p>Assets: {catalog.summary.total_assets}</p>
            <p>Collections: {catalog.summary.total_collections}</p>
            <p>Data Volume: {catalog.summary.data_volume}</p>
            <p>Coverage: {(catalog.summary.coverage * 100).toFixed(1)}%</p>
          </div>
        ))}
      </div>

      <h3>Jobs ({jobs.length})</h3>
      <div>
        {jobs.map(job => (
          <div key={job.id}>
            <h4>Job {job.id}</h4>
            <p>Type: {job.type}</p>
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

export default DataCatalog;
```

## Best Practices

### Catalog Design

1. **Define Clear Asset Categories**: Organize assets into logical categories and domains
2. **Implement Proper Tagging**: Use consistent tagging strategies for discovery
3. **Document Metadata**: Provide comprehensive metadata for all assets
4. **Maintain Schema Information**: Keep schema definitions up-to-date
5. **Track Data Lineage**: Document data flows and dependencies

### Asset Management

1. **Asset Discovery**: Implement automated asset discovery where possible
2. **Quality Monitoring**: Regularly assess and monitor data quality
3. **Usage Tracking**: Monitor asset usage patterns and performance
4. **Governance Controls**: Implement proper access controls and policies
5. **Lifecycle Management**: Manage asset lifecycle from creation to retirement

### Performance Optimization

1. **Background Processing**: Use background jobs for large catalog operations
2. **Incremental Updates**: Implement incremental catalog updates
3. **Caching Strategy**: Cache frequently accessed catalog data
4. **Pagination**: Use pagination for large catalog listings
5. **Index Optimization**: Optimize indexes for search and discovery

### Security and Governance

1. **Access Controls**: Implement role-based access controls
2. **Data Classification**: Classify data by sensitivity and compliance requirements
3. **Audit Logging**: Log all catalog operations for audit purposes
4. **Policy Enforcement**: Implement and enforce data governance policies
5. **Compliance Monitoring**: Monitor compliance with regulatory requirements

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Requests per minute**: 100 requests per minute per API key
- **Concurrent jobs**: Maximum 10 concurrent background jobs per API key
- **Request size**: Maximum 50MB per request

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
   - Check enum values for type and status fields

2. **Authentication Errors**
   - Verify API key is valid and not expired
   - Check Authorization header format
   - Ensure API key has required permissions

3. **Job Timeouts**
   - Increase timeout settings for large catalogs
   - Use background jobs for complex operations
   - Monitor job progress and implement retry logic

4. **Memory Issues**
   - Reduce catalog size for complex catalogs
   - Use pagination for large asset listings
   - Implement incremental catalog updates

### Support

For additional support:

- **Documentation**: [https://docs.kyb-platform.com/api/catalog](https://docs.kyb-platform.com/api/catalog)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)

---

**API Version**: 3.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
