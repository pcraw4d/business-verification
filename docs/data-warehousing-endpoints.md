# Data Warehousing API Documentation

## Overview

The Data Warehousing API provides comprehensive functionality for managing data warehouses, ETL processes, and data pipelines. This API enables organizations to create, configure, and manage enterprise-grade data warehousing solutions with advanced features for data processing, transformation, and analytics.

## Base URL

```
https://api.kyb-platform.com/v1
```

## Authentication

All API requests require authentication using an API key in the Authorization header:

```
Authorization: Bearer YOUR_API_KEY
```

## Response Format

All responses are returned in JSON format with the following structure:

```json
{
  "id": "resource_id",
  "name": "Resource Name",
  "status": "status",
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z"
}
```

## Supported Warehouse Types

- `oltp` - Online Transaction Processing
- `olap` - Online Analytical Processing
- `data_lake` - Data Lake
- `data_mart` - Data Mart
- `hybrid` - Hybrid Warehouse

## Supported ETL Process Types

- `extract` - Extract only
- `transform` - Transform only
- `load` - Load only
- `full` - Full ETL process
- `incremental` - Incremental ETL process

## Supported Pipeline Statuses

- `pending` - Pipeline is pending execution
- `running` - Pipeline is currently running
- `completed` - Pipeline has completed successfully
- `failed` - Pipeline has failed
- `cancelled` - Pipeline has been cancelled

## API Endpoints

### 1. Create Data Warehouse

Creates a new data warehouse with specified configuration.

**Endpoint:** `POST /warehouses`

**Request Body:**
```json
{
  "name": "Analytics Warehouse",
  "type": "olap",
  "description": "Enterprise analytics data warehouse",
  "storage_config": {
    "storage_type": "postgresql",
    "capacity": "10TB",
    "compression": "gzip",
    "partitioning": {
      "strategy": "date_range",
      "columns": ["created_date"],
      "partitions": 12,
      "auto_partition": true
    },
    "indexing": {
      "primary_index": "id",
      "secondary_indexes": ["business_id", "created_date"],
      "auto_indexing": true
    },
    "retention_policy": {
      "retention_period": "7 years",
      "archive_strategy": "cold_storage",
      "cleanup_schedule": "monthly"
    }
  },
  "security_config": {
    "encryption": {
      "algorithm": "AES-256",
      "key_management": "AWS KMS",
      "at_rest": true,
      "in_transit": true
    },
    "access_control": {
      "authentication": "oauth2",
      "authorization": "rbac",
      "roles": ["admin", "analyst", "viewer"],
      "permissions": ["read", "write", "delete"]
    },
    "audit_logging": {
      "enabled": true,
      "log_level": "info",
      "retention_days": 365,
      "destinations": ["cloudwatch", "splunk"]
    },
    "data_masking": {
      "enabled": true,
      "masking_rules": ["pii", "financial"],
      "sensitive_data": ["ssn", "credit_card"]
    }
  },
  "performance_config": {
    "query_optimization": {
      "query_planner": "cost_based",
      "statistics": true,
      "auto_optimization": true
    },
    "caching": {
      "cache_type": "redis",
      "cache_size": "2GB",
      "ttl": "1 hour",
      "eviction_policy": "lru"
    },
    "concurrency": {
      "max_connections": 100,
      "max_queries": 50,
      "connection_pool": 20
    },
    "resource_limits": {
      "cpu_limit": "8 cores",
      "memory_limit": "32GB",
      "disk_limit": "1TB",
      "network_limit": "1Gbps"
    }
  },
  "backup_config": {
    "backup_type": "incremental",
    "schedule": "daily",
    "retention": "30 days",
    "compression": true,
    "encryption": true
  },
  "monitoring_config": {
    "metrics": ["cpu", "memory", "disk", "network"],
    "alerts": ["high_cpu", "low_disk", "connection_failure"],
    "dashboard": "warehouse_monitoring",
    "health_checks": ["connectivity", "performance", "security"]
  }
}
```

**Response:**
```json
{
  "id": "warehouse_1734618600000000000",
  "name": "Analytics Warehouse",
  "type": "olap",
  "status": "creating",
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:30:00Z",
  "configuration": {},
  "metrics": {
    "storage_used": "0 GB",
    "storage_total": "10TB",
    "query_count": 0,
    "avg_query_time": 0,
    "active_connections": 0,
    "cpu_usage": 0,
    "memory_usage": 0
  },
  "health": {
    "status": "healthy",
    "last_check": "2024-12-19T10:30:00Z",
    "issues": [],
    "recommendations": []
  }
}
```

### 2. Get Data Warehouse

Retrieves information about a specific data warehouse.

**Endpoint:** `GET /warehouses?id={warehouse_id}`

**Response:**
```json
{
  "id": "warehouse_1734618600000000000",
  "name": "Analytics Warehouse",
  "type": "olap",
  "status": "active",
  "created_at": "2024-12-19T10:30:00Z",
  "updated_at": "2024-12-19T10:32:00Z",
  "configuration": {},
  "metrics": {
    "storage_used": "2.5TB",
    "storage_total": "10TB",
    "query_count": 15420,
    "avg_query_time": 0.85,
    "active_connections": 15,
    "cpu_usage": 45.2,
    "memory_usage": 67.8
  },
  "health": {
    "status": "healthy",
    "last_check": "2024-12-19T10:32:00Z",
    "issues": [],
    "recommendations": ["Consider increasing cache size for better performance"]
  }
}
```

### 3. List Data Warehouses

Retrieves a list of all data warehouses.

**Endpoint:** `GET /warehouses`

**Response:**
```json
{
  "warehouses": [
    {
      "id": "warehouse_1734618600000000000",
      "name": "Analytics Warehouse",
      "type": "olap",
      "status": "active",
      "created_at": "2024-12-19T10:30:00Z",
      "updated_at": "2024-12-19T10:32:00Z"
    },
    {
      "id": "warehouse_1734618700000000000",
      "name": "Operational Data Store",
      "type": "oltp",
      "status": "active",
      "created_at": "2024-12-19T10:35:00Z",
      "updated_at": "2024-12-19T10:37:00Z"
    }
  ],
  "count": 2
}
```

### 4. Create ETL Process

Creates a new ETL (Extract, Transform, Load) process.

**Endpoint:** `POST /etl`

**Request Body:**
```json
{
  "name": "Customer Data ETL",
  "type": "full",
  "description": "ETL process for customer data integration",
  "source_config": {
    "source_type": "postgresql",
    "connection_string": "postgres://user:pass@localhost:5432/source_db",
    "query": "SELECT * FROM customers WHERE updated_at > $1",
    "filters": {
      "status": "active",
      "region": "US"
    },
    "incremental_key": "updated_at",
    "batch_size": 1000
  },
  "transform_config": {
    "transformations": [
      {
        "name": "clean_customer_data",
        "type": "filter",
        "expression": "email IS NOT NULL AND email != ''",
        "parameters": {},
        "description": "Remove records with invalid email addresses"
      },
      {
        "name": "standardize_phone",
        "type": "transform",
        "expression": "REGEXP_REPLACE(phone, '[^0-9]', '', 'g')",
        "parameters": {},
        "description": "Standardize phone number format"
      }
    ],
    "data_quality": {
      "validation_rules": [
        {
          "name": "email_format",
          "type": "regex",
          "expression": "^[^@]+@[^@]+\\.[^@]+$",
          "severity": "error",
          "description": "Validate email format"
        }
      ],
      "cleaning_rules": [
        {
          "name": "remove_duplicates",
          "type": "deduplication",
          "expression": "DISTINCT ON (email)",
          "description": "Remove duplicate email addresses"
        }
      ],
      "profiling": true
    },
    "aggregations": [
      {
        "name": "customer_count_by_region",
        "function": "count",
        "group_by": ["region"],
        "having": "count > 100",
        "description": "Count customers by region"
      }
    ],
    "joins": [
      {
        "name": "customer_orders",
        "type": "left",
        "left_table": "customers",
        "right_table": "orders",
        "condition": "customers.id = orders.customer_id",
        "description": "Join customers with their orders"
      }
    ]
  },
  "target_config": {
    "target_type": "postgresql",
    "connection_string": "postgres://user:pass@localhost:5432/warehouse",
    "table_name": "customers",
    "schema": "public",
    "load_strategy": "upsert",
    "partitioning": {
      "strategy": "hash",
      "columns": ["region"],
      "partitions": 4,
      "auto_partition": false
    }
  },
  "schedule": {
    "schedule_type": "cron",
    "cron_expression": "0 2 * * *",
    "start_time": "2024-12-19T02:00:00Z",
    "end_time": "2024-12-19T06:00:00Z",
    "timezone": "UTC",
    "retry_policy": {
      "max_retries": 3,
      "retry_interval": "5 minutes",
      "backoff_strategy": "exponential"
    }
  },
  "validation": {
    "pre_validation": [
      {
        "name": "source_data_check",
        "type": "count",
        "expression": "COUNT(*) > 0",
        "severity": "error",
        "description": "Ensure source data exists"
      }
    ],
    "post_validation": [
      {
        "name": "target_data_check",
        "type": "count",
        "expression": "COUNT(*) > 0",
        "severity": "error",
        "description": "Ensure target data was loaded"
      }
    ],
    "data_profiling": true,
    "quality_metrics": ["completeness", "accuracy", "consistency"]
  },
  "error_handling": {
    "error_action": "stop",
    "error_threshold": 10,
    "error_logging": true,
    "error_notification": true
  }
}
```

**Response:**
```json
{
  "id": "etl_1734618800000000000",
  "name": "Customer Data ETL",
  "type": "full",
  "status": "pending",
  "created_at": "2024-12-19T10:40:00Z",
  "updated_at": "2024-12-19T10:40:00Z",
  "configuration": {
    "source": {...},
    "transform": {...},
    "target": {...},
    "schedule": {...}
  },
  "statistics": {
    "total_runs": 0,
    "successful_runs": 0,
    "failed_runs": 0,
    "avg_duration": 0,
    "records_processed": 0,
    "data_volume": "0 MB"
  },
  "errors": []
}
```

### 5. Get ETL Process

Retrieves information about a specific ETL process.

**Endpoint:** `GET /etl?id={etl_id}`

**Response:**
```json
{
  "id": "etl_1734618800000000000",
  "name": "Customer Data ETL",
  "type": "full",
  "status": "completed",
  "created_at": "2024-12-19T10:40:00Z",
  "updated_at": "2024-12-19T10:45:00Z",
  "last_run": "2024-12-19T10:42:00Z",
  "next_run": "2024-12-20T02:00:00Z",
  "configuration": {...},
  "statistics": {
    "total_runs": 1,
    "successful_runs": 1,
    "failed_runs": 0,
    "avg_duration": 180.5,
    "records_processed": 15420,
    "data_volume": "2.5 MB"
  },
  "errors": []
}
```

### 6. List ETL Processes

Retrieves a list of all ETL processes.

**Endpoint:** `GET /etl`

**Response:**
```json
{
  "etl_processes": [
    {
      "id": "etl_1734618800000000000",
      "name": "Customer Data ETL",
      "type": "full",
      "status": "completed",
      "created_at": "2024-12-19T10:40:00Z",
      "updated_at": "2024-12-19T10:45:00Z"
    },
    {
      "id": "etl_1734618900000000000",
      "name": "Order Data ETL",
      "type": "incremental",
      "status": "running",
      "created_at": "2024-12-19T10:50:00Z",
      "updated_at": "2024-12-19T10:52:00Z"
    }
  ],
  "count": 2
}
```

### 7. Create Data Pipeline

Creates a new data pipeline with multiple stages.

**Endpoint:** `POST /pipelines`

**Request Body:**
```json
{
  "name": "Customer Analytics Pipeline",
  "description": "End-to-end pipeline for customer analytics",
  "stages": [
    {
      "name": "data_extraction",
      "type": "extract",
      "order": 1,
      "configuration": {
        "source": "customer_database",
        "query": "SELECT * FROM customers",
        "batch_size": 1000
      },
      "dependencies": [],
      "timeout": "30 minutes",
      "retry_policy": {
        "max_retries": 3,
        "retry_interval": "5 minutes",
        "backoff_strategy": "exponential"
      }
    },
    {
      "name": "data_cleaning",
      "type": "transform",
      "order": 2,
      "configuration": {
        "operations": ["remove_duplicates", "validate_emails", "standardize_phones"]
      },
      "dependencies": ["data_extraction"],
      "timeout": "15 minutes",
      "retry_policy": {
        "max_retries": 2,
        "retry_interval": "3 minutes",
        "backoff_strategy": "linear"
      }
    },
    {
      "name": "data_enrichment",
      "type": "transform",
      "order": 3,
      "configuration": {
        "enrichment_sources": ["geolocation", "demographics", "behavioral"]
      },
      "dependencies": ["data_cleaning"],
      "timeout": "20 minutes",
      "retry_policy": {
        "max_retries": 3,
        "retry_interval": "5 minutes",
        "backoff_strategy": "exponential"
      }
    },
    {
      "name": "data_loading",
      "type": "load",
      "order": 4,
      "configuration": {
        "target": "analytics_warehouse",
        "table": "customers_analytics",
        "strategy": "upsert"
      },
      "dependencies": ["data_enrichment"],
      "timeout": "10 minutes",
      "retry_policy": {
        "max_retries": 2,
        "retry_interval": "2 minutes",
        "backoff_strategy": "linear"
      }
    }
  ],
  "triggers": [
    {
      "name": "daily_schedule",
      "type": "schedule",
      "condition": "time_based",
      "schedule": "0 2 * * *",
      "configuration": {
        "timezone": "UTC"
      }
    },
    {
      "name": "data_availability",
      "type": "event",
      "condition": "file_ready",
      "configuration": {
        "file_pattern": "customers_*.csv",
        "location": "s3://data-bucket/incoming/"
      }
    }
  ],
  "monitoring": {
    "metrics": ["execution_time", "records_processed", "error_rate"],
    "logging": true,
    "tracing": true,
    "health_checks": ["stage_completion", "data_quality", "performance"]
  },
  "alerting": {
    "alerts": [
      {
        "name": "pipeline_failure",
        "condition": "status == 'failed'",
        "severity": "critical",
        "threshold": "1",
        "duration": "5 minutes"
      },
      {
        "name": "slow_execution",
        "condition": "execution_time > 2 hours",
        "severity": "warning",
        "threshold": "1",
        "duration": "10 minutes"
      }
    ],
    "notification_channels": ["email", "slack", "pagerduty"],
    "escalation_policy": "immediate"
  },
  "versioning": {
    "version_control": true,
    "branching": true,
    "tagging": true,
    "rollback": true
  }
}
```

**Response:**
```json
{
  "id": "pipeline_1734619000000000000",
  "name": "Customer Analytics Pipeline",
  "status": "pending",
  "created_at": "2024-12-19T11:00:00Z",
  "updated_at": "2024-12-19T11:00:00Z",
  "stages": [
    {
      "name": "data_extraction",
      "status": "pending",
      "records_processed": 0,
      "errors": []
    },
    {
      "name": "data_cleaning",
      "status": "pending",
      "records_processed": 0,
      "errors": []
    },
    {
      "name": "data_enrichment",
      "status": "pending",
      "records_processed": 0,
      "errors": []
    },
    {
      "name": "data_loading",
      "status": "pending",
      "records_processed": 0,
      "errors": []
    }
  ],
  "statistics": {
    "total_runs": 0,
    "successful_runs": 0,
    "failed_runs": 0,
    "avg_duration": 0,
    "total_records": 0,
    "data_volume": "0 MB"
  },
  "alerts": []
}
```

### 8. Get Data Pipeline

Retrieves information about a specific data pipeline.

**Endpoint:** `GET /pipelines?id={pipeline_id}`

**Response:**
```json
{
  "id": "pipeline_1734619000000000000",
  "name": "Customer Analytics Pipeline",
  "status": "completed",
  "created_at": "2024-12-19T11:00:00Z",
  "updated_at": "2024-12-19T11:30:00Z",
  "last_run": "2024-12-19T11:05:00Z",
  "next_run": "2024-12-20T02:00:00Z",
  "stages": [
    {
      "name": "data_extraction",
      "status": "completed",
      "start_time": "2024-12-19T11:05:00Z",
      "end_time": "2024-12-19T11:08:00Z",
      "duration": "3 minutes",
      "records_processed": 15420,
      "errors": []
    },
    {
      "name": "data_cleaning",
      "status": "completed",
      "start_time": "2024-12-19T11:08:00Z",
      "end_time": "2024-12-19T11:10:00Z",
      "duration": "2 minutes",
      "records_processed": 15380,
      "errors": []
    },
    {
      "name": "data_enrichment",
      "status": "completed",
      "start_time": "2024-12-19T11:10:00Z",
      "end_time": "2024-12-19T11:20:00Z",
      "duration": "10 minutes",
      "records_processed": 15380,
      "errors": []
    },
    {
      "name": "data_loading",
      "status": "completed",
      "start_time": "2024-12-19T11:20:00Z",
      "end_time": "2024-12-19T11:22:00Z",
      "duration": "2 minutes",
      "records_processed": 15380,
      "errors": []
    }
  ],
  "statistics": {
    "total_runs": 1,
    "successful_runs": 1,
    "failed_runs": 0,
    "avg_duration": 17.0,
    "total_records": 15380,
    "data_volume": "5.2 MB"
  },
  "alerts": []
}
```

### 9. List Data Pipelines

Retrieves a list of all data pipelines.

**Endpoint:** `GET /pipelines`

**Response:**
```json
{
  "pipelines": [
    {
      "id": "pipeline_1734619000000000000",
      "name": "Customer Analytics Pipeline",
      "status": "completed",
      "created_at": "2024-12-19T11:00:00Z",
      "updated_at": "2024-12-19T11:30:00Z"
    },
    {
      "id": "pipeline_1734619100000000000",
      "name": "Order Processing Pipeline",
      "status": "running",
      "created_at": "2024-12-19T11:35:00Z",
      "updated_at": "2024-12-19T11:40:00Z"
    }
  ],
  "count": 2
}
```

### 10. Create Warehouse Job

Creates a background job for warehouse operations.

**Endpoint:** `POST /warehouse/jobs`

**Request Body:**
```json
{
  "type": "backup",
  "warehouse_id": "warehouse_1734618600000000000",
  "configuration": {
    "backup_type": "full",
    "compression": true,
    "encryption": true,
    "destination": "s3://backup-bucket/warehouse-backups/"
  }
}
```

**Response:**
```json
{
  "job_id": "job_1734619200000000000",
  "status": "pending",
  "created_at": "2024-12-19T12:00:00Z"
}
```

### 11. Get Warehouse Job

Retrieves the status of a warehouse job.

**Endpoint:** `GET /warehouse/jobs?id={job_id}`

**Response:**
```json
{
  "ID": "job_1734619200000000000",
  "Type": "backup",
  "Status": "completed",
  "Progress": 100,
  "CreatedAt": "2024-12-19T12:00:00Z",
  "UpdatedAt": "2024-12-19T12:05:00Z",
  "Result": {
    "message": "Job completed successfully",
    "timestamp": "2024-12-19T12:05:00Z",
    "backup_size": "2.5GB",
    "backup_location": "s3://backup-bucket/warehouse-backups/backup_20241219_120000.tar.gz"
  },
  "Error": ""
}
```

### 12. List Warehouse Jobs

Retrieves a list of all warehouse jobs.

**Endpoint:** `GET /warehouse/jobs`

**Response:**
```json
{
  "jobs": [
    {
      "ID": "job_1734619200000000000",
      "Type": "backup",
      "Status": "completed",
      "Progress": 100,
      "CreatedAt": "2024-12-19T12:00:00Z",
      "UpdatedAt": "2024-12-19T12:05:00Z"
    },
    {
      "ID": "job_1734619300000000000",
      "Type": "maintenance",
      "Status": "running",
      "Progress": 75,
      "CreatedAt": "2024-12-19T12:10:00Z",
      "UpdatedAt": "2024-12-19T12:12:00Z"
    }
  ],
  "count": 2
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "warehouse name is required"
}
```

### 404 Not Found
```json
{
  "error": "Warehouse not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error occurred"
}
```

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

const API_BASE_URL = 'https://api.kyb-platform.com/v1';
const API_KEY = 'your_api_key';

const headers = {
  'Authorization': `Bearer ${API_KEY}`,
  'Content-Type': 'application/json'
};

// Create a data warehouse
async function createWarehouse() {
  try {
    const response = await axios.post(`${API_BASE_URL}/warehouses`, {
      name: 'Analytics Warehouse',
      type: 'olap',
      storage_config: {
        storage_type: 'postgresql',
        capacity: '10TB'
      }
    }, { headers });
    
    console.log('Warehouse created:', response.data);
    return response.data.id;
  } catch (error) {
    console.error('Error creating warehouse:', error.response.data);
  }
}

// Create an ETL process
async function createETLProcess(warehouseId) {
  try {
    const response = await axios.post(`${API_BASE_URL}/etl`, {
      name: 'Customer Data ETL',
      type: 'full',
      source_config: {
        source_type: 'postgresql',
        connection_string: 'postgres://user:pass@localhost:5432/source_db'
      },
      target_config: {
        target_type: 'postgresql',
        connection_string: 'postgres://user:pass@localhost:5432/warehouse'
      }
    }, { headers });
    
    console.log('ETL process created:', response.data);
    return response.data.id;
  } catch (error) {
    console.error('Error creating ETL process:', error.response.data);
  }
}

// Monitor warehouse job
async function monitorJob(jobId) {
  try {
    const response = await axios.get(`${API_BASE_URL}/warehouse/jobs?id=${jobId}`, { headers });
    console.log('Job status:', response.data);
    return response.data.Status;
  } catch (error) {
    console.error('Error getting job status:', error.response.data);
  }
}
```

### Python

```python
import requests
import json
import time

API_BASE_URL = 'https://api.kyb-platform.com/v1'
API_KEY = 'your_api_key'

headers = {
    'Authorization': f'Bearer {API_KEY}',
    'Content-Type': 'application/json'
}

def create_warehouse():
    """Create a new data warehouse"""
    try:
        payload = {
            'name': 'Analytics Warehouse',
            'type': 'olap',
            'storage_config': {
                'storage_type': 'postgresql',
                'capacity': '10TB'
            }
        }
        
        response = requests.post(
            f'{API_BASE_URL}/warehouses',
            headers=headers,
            json=payload
        )
        response.raise_for_status()
        
        warehouse = response.json()
        print(f"Warehouse created: {warehouse['id']}")
        return warehouse['id']
        
    except requests.exceptions.RequestException as e:
        print(f"Error creating warehouse: {e}")
        return None

def create_pipeline():
    """Create a data pipeline"""
    try:
        payload = {
            'name': 'Customer Analytics Pipeline',
            'stages': [
                {
                    'name': 'data_extraction',
                    'type': 'extract',
                    'order': 1,
                    'configuration': {
                        'source': 'customer_database'
                    }
                },
                {
                    'name': 'data_loading',
                    'type': 'load',
                    'order': 2,
                    'configuration': {
                        'target': 'analytics_warehouse'
                    }
                }
            ]
        }
        
        response = requests.post(
            f'{API_BASE_URL}/pipelines',
            headers=headers,
            json=payload
        )
        response.raise_for_status()
        
        pipeline = response.json()
        print(f"Pipeline created: {pipeline['id']}")
        return pipeline['id']
        
    except requests.exceptions.RequestException as e:
        print(f"Error creating pipeline: {e}")
        return None

def wait_for_job_completion(job_id, max_wait_time=300):
    """Wait for a job to complete"""
    start_time = time.time()
    
    while time.time() - start_time < max_wait_time:
        try:
            response = requests.get(
                f'{API_BASE_URL}/warehouse/jobs?id={job_id}',
                headers=headers
            )
            response.raise_for_status()
            
            job = response.json()
            status = job['Status']
            
            print(f"Job {job_id} status: {status} (Progress: {job['Progress']}%)")
            
            if status in ['completed', 'failed']:
                return status
                
            time.sleep(5)  # Wait 5 seconds before checking again
            
        except requests.exceptions.RequestException as e:
            print(f"Error checking job status: {e}")
            time.sleep(5)
    
    return 'timeout'
```

### React/TypeScript

```typescript
import axios, { AxiosResponse } from 'axios';

interface DataWarehouse {
  id: string;
  name: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface ETLProcess {
  id: string;
  name: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface DataPipeline {
  id: string;
  name: string;
  status: string;
  created_at: string;
  updated_at: string;
  stages: Array<{
    name: string;
    status: string;
    records_processed: number;
  }>;
}

class DataWarehousingAPI {
  private baseURL: string;
  private apiKey: string;

  constructor(baseURL: string, apiKey: string) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
  }

  private getHeaders() {
    return {
      'Authorization': `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json'
    };
  }

  async createWarehouse(warehouseData: any): Promise<DataWarehouse> {
    const response: AxiosResponse<DataWarehouse> = await axios.post(
      `${this.baseURL}/warehouses`,
      warehouseData,
      { headers: this.getHeaders() }
    );
    return response.data;
  }

  async getWarehouse(warehouseId: string): Promise<DataWarehouse> {
    const response: AxiosResponse<DataWarehouse> = await axios.get(
      `${this.baseURL}/warehouses?id=${warehouseId}`,
      { headers: this.getHeaders() }
    );
    return response.data;
  }

  async listWarehouses(): Promise<{ warehouses: DataWarehouse[]; count: number }> {
    const response: AxiosResponse<{ warehouses: DataWarehouse[]; count: number }> = await axios.get(
      `${this.baseURL}/warehouses`,
      { headers: this.getHeaders() }
    );
    return response.data;
  }

  async createETLProcess(etlData: any): Promise<ETLProcess> {
    const response: AxiosResponse<ETLProcess> = await axios.post(
      `${this.baseURL}/etl`,
      etlData,
      { headers: this.getHeaders() }
    );
    return response.data;
  }

  async createPipeline(pipelineData: any): Promise<DataPipeline> {
    const response: AxiosResponse<DataPipeline> = await axios.post(
      `${this.baseURL}/pipelines`,
      pipelineData,
      { headers: this.getHeaders() }
    );
    return response.data;
  }

  async getPipeline(pipelineId: string): Promise<DataPipeline> {
    const response: AxiosResponse<DataPipeline> = await axios.get(
      `${this.baseURL}/pipelines?id=${pipelineId}`,
      { headers: this.getHeaders() }
    );
    return response.data;
  }
}

// Usage example
const api = new DataWarehousingAPI('https://api.kyb-platform.com/v1', 'your_api_key');

// Create a warehouse
const warehouse = await api.createWarehouse({
  name: 'Analytics Warehouse',
  type: 'olap',
  storage_config: {
    storage_type: 'postgresql',
    capacity: '10TB'
  }
});

// Create an ETL process
const etlProcess = await api.createETLProcess({
  name: 'Customer Data ETL',
  type: 'full',
  source_config: {
    source_type: 'postgresql',
    connection_string: 'postgres://user:pass@localhost:5432/source_db'
  },
  target_config: {
    target_type: 'postgresql',
    connection_string: 'postgres://user:pass@localhost:5432/warehouse'
  }
});

// Create a pipeline
const pipeline = await api.createPipeline({
  name: 'Customer Analytics Pipeline',
  stages: [
    {
      name: 'data_extraction',
      type: 'extract',
      order: 1,
      configuration: { source: 'customer_database' }
    },
    {
      name: 'data_loading',
      type: 'load',
      order: 2,
      configuration: { target: 'analytics_warehouse' }
    }
  ]
});
```

## Best Practices

### 1. Warehouse Design
- Choose the appropriate warehouse type based on your use case
- Implement proper partitioning and indexing strategies
- Configure security settings according to your compliance requirements
- Set up monitoring and alerting for proactive issue detection

### 2. ETL Process Design
- Use incremental ETL processes for large datasets
- Implement proper error handling and retry mechanisms
- Validate data quality at each stage
- Monitor ETL performance and optimize bottlenecks

### 3. Pipeline Management
- Design pipelines with clear stage dependencies
- Implement proper timeout and retry policies
- Use event-driven triggers for real-time processing
- Monitor pipeline health and performance

### 4. Job Management
- Use background jobs for long-running operations
- Implement proper job monitoring and status tracking
- Set up alerts for job failures
- Clean up completed jobs periodically

### 5. Security
- Use encrypted connections for all database connections
- Implement proper access control and role-based permissions
- Enable audit logging for compliance
- Use data masking for sensitive information

### 6. Performance
- Optimize query performance with proper indexing
- Use connection pooling for database connections
- Implement caching strategies for frequently accessed data
- Monitor resource usage and scale accordingly

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Standard Plan**: 100 requests per minute
- **Professional Plan**: 1000 requests per minute
- **Enterprise Plan**: 10000 requests per minute

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Monitoring and Alerts

### Key Metrics to Monitor
- Warehouse performance (query response times, throughput)
- ETL process success rates and execution times
- Pipeline stage completion times and error rates
- Resource utilization (CPU, memory, disk, network)
- Data quality metrics and validation failures

### Recommended Alerts
- Warehouse availability and connectivity
- ETL process failures and timeouts
- Pipeline stage failures and delays
- High resource utilization
- Data quality violations

## Troubleshooting

### Common Issues

1. **Warehouse Creation Fails**
   - Check storage capacity and permissions
   - Verify network connectivity
   - Review security configuration

2. **ETL Process Errors**
   - Validate source and target connection strings
   - Check data format and schema compatibility
   - Review transformation rules and expressions

3. **Pipeline Stage Failures**
   - Verify stage dependencies and order
   - Check timeout and retry configurations
   - Review stage-specific configurations

4. **Job Timeouts**
   - Increase timeout values for large operations
   - Optimize query performance
   - Consider breaking large jobs into smaller chunks

### Debug Information

Enable debug logging by setting the log level to "debug" in your configuration. Debug logs include:
- Detailed request/response information
- SQL queries and execution plans
- Performance metrics and timing information
- Error stack traces and context

### Support

For additional support:
- **Documentation**: https://docs.kyb-platform.com/warehousing
- **API Reference**: https://api.kyb-platform.com/docs
- **Support Email**: support@kyb-platform.com
- **Community Forum**: https://community.kyb-platform.com

## Migration Guide

### From Version 2.x to 3.x

1. **Updated Warehouse Types**
   - `warehouse` → `olap`
   - `operational` → `oltp`
   - `lake` → `data_lake`

2. **Enhanced ETL Configuration**
   - New validation and error handling options
   - Improved scheduling and retry policies
   - Enhanced data quality features

3. **Pipeline Improvements**
   - Multi-stage pipeline support
   - Event-driven triggers
   - Advanced monitoring and alerting

### Breaking Changes

- Warehouse type enum values have changed
- ETL process configuration structure updated
- Pipeline API response format modified

## Future Enhancements

### Planned Features
- Real-time streaming pipelines
- Advanced data quality monitoring
- Machine learning integration
- Multi-cloud support
- Advanced security features

### Roadmap
- **Q1 2025**: Real-time streaming and ML integration
- **Q2 2025**: Multi-cloud and advanced security
- **Q3 2025**: Advanced analytics and visualization
- **Q4 2025**: Enterprise features and compliance tools
