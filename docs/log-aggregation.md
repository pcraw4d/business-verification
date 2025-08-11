# KYB Platform - Log Aggregation

## Overview

The KYB Platform implements comprehensive log aggregation using structured logging, Elasticsearch, and centralized log management. This system provides real-time log collection, correlation, analysis, and search capabilities across all application components.

## Log Aggregation Architecture

### Log Aggregation Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   KYB Platform  │    │   Log Shipper   │    │   Elasticsearch │
│      API        │───▶│   (Buffered)    │───▶│   (Log Storage) │
│                 │    │                 │    │                 │
│ • Structured    │    │ • Batch         │    │ • Indexed       │
│   Logging       │    │ • Retry Logic   │    │ • Searchable    │
│ • Correlation   │    │ • Error Handling│    │ • Analytics     │
│   IDs           │    │ • Compression   │    │ • Visualization │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Log Files     │    │   Kibana        │    │   Log Analysis  │
│                 │    │                 │    │                 │
│ • Local Storage │    │ • Dashboards    │    │ • Search        │
│ • Rotation      │    │ • Visualizations│    │ • Filtering     │
│ • Compression   │    │ • Alerts        │    │ • Aggregation   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Log Flow

1. **Application Logging**: KYB Platform generates structured logs with correlation IDs
2. **Log Collection**: Logs are collected and buffered locally
3. **Log Shipping**: Batched logs are shipped to Elasticsearch
4. **Log Storage**: Logs are indexed and stored in Elasticsearch
5. **Log Analysis**: Logs are analyzed and visualized in Kibana
6. **Log Search**: Real-time search and filtering capabilities

## Log Aggregation Components

### 1. Log Aggregation System

**Location**: `internal/observability/log_aggregation.go`

**Features**:
- Structured logging with correlation IDs
- Multiple output destinations (console, file, Elasticsearch)
- Log shipping with batching and retry logic
- Log search and analysis capabilities
- Log statistics and aggregation
- Business event logging
- Security event logging
- Performance event logging
- Database event logging
- External API event logging

**Key Components**:
```go
// Log Aggregation System
type LogAggregationSystem struct {
    logger        *zap.Logger
    elasticClient *elasticsearch.Client
    config        *LogAggregationConfig
    shutdownChan  chan struct{}
}

// Log Entry Structure
type LogEntry struct {
    Timestamp   time.Time              `json:"timestamp"`
    Level       string                 `json:"level"`
    Message     string                 `json:"message"`
    Logger      string                 `json:"logger"`
    Environment string                 `json:"environment"`
    Application string                 `json:"application"`
    Version     string                 `json:"version"`
    TraceID     string                 `json:"trace_id,omitempty"`
    SpanID      string                 `json:"span_id,omitempty"`
    UserID      string                 `json:"user_id,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    Endpoint    string                 `json:"endpoint,omitempty"`
    Method      string                 `json:"method,omitempty"`
    StatusCode  int                    `json:"status_code,omitempty"`
    Duration    float64                `json:"duration,omitempty"`
    IPAddress   string                 `json:"ip_address,omitempty"`
    UserAgent   string                 `json:"user_agent,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Stack       string                 `json:"stack,omitempty"`
    Fields      map[string]interface{} `json:"fields,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Log Shipper
type LogShipper struct {
    elasticClient *elasticsearch.Client
    config        *LogAggregationConfig
    logChan       chan LogEntry
    shutdownChan  chan struct{}
    logger        *zap.Logger
}
```

### 2. Log Aggregation Middleware

**Location**: `internal/api/middleware/log_aggregation.go`

**Features**:
- Automatic HTTP request logging
- Correlation ID injection
- Business event logging
- Security event logging
- Performance event logging
- Database event logging
- External API event logging

**Integration**:
```go
// Initialize log aggregation middleware
logAggregationMiddleware := middleware.NewLogAggregationMiddleware(
    logAggregationSystem,
    logger,
    environment,
)

// Apply to HTTP handlers
router.Use(logAggregationMiddleware.LogHTTPRequests)
router.Use(logAggregationMiddleware.LogBusinessEvents)
router.Use(logAggregationMiddleware.LogSecurityEvents)
router.Use(logAggregationMiddleware.LogPerformanceEvents)
```

### 3. Elasticsearch Configuration

**Location**: `deployments/elasticsearch/elasticsearch.yml`

**Key Features**:
- Multi-node cluster configuration
- Security and authentication
- Index lifecycle management
- Snapshot and restore
- Monitoring and alerting
- Performance optimization

## Log Categories

### 1. HTTP Request Logs

**Purpose**: Monitor API requests and responses

**Log Fields**:
- Request method, path, query parameters
- Response status code and duration
- User agent and IP address
- Correlation IDs (trace_id, span_id, request_id)
- Request headers and content type
- Response headers and content type

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "HTTP request completed",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "span_id": "xyz789",
  "request_id": "req-20240115-103000-001",
  "endpoint": "/v1/classify",
  "method": "POST",
  "status_code": 200,
  "duration": 0.245,
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "fields": {
    "query": "business_name=Acme Corp",
    "content_type": "application/json",
    "accept": "application/json"
  }
}
```

### 2. Business Event Logs

**Purpose**: Track business-specific events and activities

**Log Categories**:
- Classification requests and results
- Risk assessment events
- Compliance check events
- Authentication events
- User registration events
- API key usage events

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "Business event",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "user_id": "user-123",
  "fields": {
    "event_type": "classification",
    "event_name": "business_classification_request",
    "endpoint": "/v1/classify",
    "method": "POST",
    "business_name": "Acme Corporation",
    "confidence_score": 0.95,
    "naics_code": "541511"
  }
}
```

### 3. Security Event Logs

**Purpose**: Monitor security-related activities and threats

**Log Categories**:
- Authentication attempts (success/failure)
- API key usage and abuse
- Rate limit violations
- Suspicious activities
- Access control events
- Security incidents

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "WARN",
  "message": "Security event",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "ip_address": "192.168.1.100",
  "fields": {
    "event_type": "authentication",
    "event_name": "login_attempt",
    "severity": "info",
    "category": "security",
    "endpoint": "/v1/auth/login",
    "method": "POST",
    "user_agent": "Mozilla/5.0...",
    "success": true
  }
}
```

### 4. Performance Event Logs

**Purpose**: Monitor application performance and bottlenecks

**Log Categories**:
- Slow HTTP requests
- Database query performance
- External API call performance
- Memory usage events
- CPU usage events
- Goroutine count events

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "Performance event",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "fields": {
    "operation": "http_request",
    "duration": 1.5,
    "category": "performance",
    "endpoint": "/v1/classify",
    "method": "POST",
    "duration_ms": 1500,
    "threshold_ms": 1000
  }
}
```

### 5. Database Event Logs

**Purpose**: Monitor database operations and performance

**Log Categories**:
- Query execution and duration
- Connection pool usage
- Transaction events
- Error events
- Performance issues
- Schema changes

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "Database event",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "fields": {
    "operation": "SELECT",
    "query": "SELECT * FROM businesses WHERE name = $1",
    "duration": 0.025,
    "category": "database",
    "table": "businesses",
    "rows_affected": 1
  }
}
```

### 6. External API Event Logs

**Purpose**: Monitor external service dependencies

**Log Categories**:
- API call requests and responses
- Response times and status codes
- Error events and retries
- Rate limiting events
- Circuit breaker events
- Dependency health

**Example Log Entry**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "External API event",
  "logger": "kyb-platform",
  "environment": "production",
  "application": "kyb-platform",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "fields": {
    "provider": "business_registry",
    "endpoint": "/api/v1/search",
    "method": "GET",
    "status_code": 200,
    "duration": 0.5,
    "category": "external_api",
    "success": true
  }
}
```

## Log Configuration

### Log Aggregation Configuration

```go
type LogAggregationConfig struct {
    // Elasticsearch configuration
    ElasticsearchURL      string
    ElasticsearchUsername string
    ElasticsearchPassword string
    ElasticsearchIndex    string
    
    // Log shipping configuration
    BatchSize     int
    BatchTimeout  time.Duration
    RetryAttempts int
    RetryDelay    time.Duration
    
    // Log level configuration
    MinLevel      zapcore.Level
    Environment   string
    Application   string
    Version       string
    
    // Output configuration
    EnableConsole bool
    EnableFile    bool
    EnableElastic bool
    LogFilePath   string
    
    // Buffer configuration
    BufferSize    int
    FlushInterval time.Duration
}
```

### Environment Configuration

**Development Environment**:
```yaml
log_aggregation:
  elasticsearch_url: "http://localhost:9200"
  elasticsearch_index: "kyb-platform-dev"
  batch_size: 100
  batch_timeout: "5s"
  retry_attempts: 3
  retry_delay: "1s"
  min_level: "DEBUG"
  environment: "development"
  enable_console: true
  enable_file: true
  enable_elastic: false
  log_file_path: "/tmp/kyb-platform.log"
  buffer_size: 1000
  flush_interval: "10s"
```

**Production Environment**:
```yaml
log_aggregation:
  elasticsearch_url: "https://elasticsearch.kybplatform.com:9200"
  elasticsearch_index: "kyb-platform-prod"
  batch_size: 500
  batch_timeout: "10s"
  retry_attempts: 5
  retry_delay: "2s"
  min_level: "INFO"
  environment: "production"
  enable_console: false
  enable_file: true
  enable_elastic: true
  log_file_path: "/var/log/kyb-platform/application.log"
  buffer_size: 5000
  flush_interval: "30s"
```

## Log Shipping

### Batch Processing

**Batch Configuration**:
- **Batch Size**: Number of log entries per batch (default: 100-500)
- **Batch Timeout**: Maximum time to wait before shipping (default: 5-10s)
- **Retry Attempts**: Number of retry attempts on failure (default: 3-5)
- **Retry Delay**: Delay between retry attempts (default: 1-2s)

**Batch Processing Flow**:
1. Log entries are buffered in memory
2. Batch is shipped when size limit is reached
3. Batch is shipped when timeout is reached
4. Failed batches are retried with exponential backoff
5. Persistent failures are logged locally

### Error Handling

**Error Scenarios**:
- Network connectivity issues
- Elasticsearch cluster unavailability
- Authentication/authorization failures
- Index creation failures
- Bulk operation failures

**Error Handling Strategy**:
- Retry with exponential backoff
- Circuit breaker pattern for persistent failures
- Local fallback logging
- Alert generation for critical failures
- Graceful degradation

## Log Search and Analysis

### Search Capabilities

**Search Queries**:
```json
// Search by time range
{
  "query": {
    "range": {
      "timestamp": {
        "gte": "2024-01-15T00:00:00Z",
        "lte": "2024-01-15T23:59:59Z"
      }
    }
  }
}

// Search by log level
{
  "query": {
    "term": {
      "level": "ERROR"
    }
  }
}

// Search by endpoint
{
  "query": {
    "term": {
      "endpoint": "/v1/classify"
    }
  }
}

// Search by trace ID
{
  "query": {
    "term": {
      "trace_id": "abc123def456"
    }
  }
}

// Search by user ID
{
  "query": {
    "term": {
      "user_id": "user-123"
    }
  }
}

// Search by error message
{
  "query": {
    "match": {
      "message": "database connection failed"
    }
  }
}
```

### Aggregation Queries

**Log Level Distribution**:
```json
{
  "aggs": {
    "log_levels": {
      "terms": {
        "field": "level"
      }
    }
  }
}
```

**Endpoint Usage**:
```json
{
  "aggs": {
    "endpoints": {
      "terms": {
        "field": "endpoint",
        "size": 10
      }
    }
  }
}
```

**Error Rate Over Time**:
```json
{
  "aggs": {
    "error_rate": {
      "filter": {
        "term": {
          "level": "ERROR"
        }
      }
    }
  }
}
```

**Response Time Percentiles**:
```json
{
  "aggs": {
    "response_time": {
      "percentiles": {
        "field": "duration",
        "percents": [50, 95, 99]
      }
    }
  }
}
```

## Log Visualization

### Kibana Dashboards

**Dashboard Categories**:
1. **Overview Dashboard**
   - Log volume over time
   - Log level distribution
   - Error rate trends
   - Active users

2. **Performance Dashboard**
   - Response time percentiles
   - Slow request analysis
   - Database performance
   - External API performance

3. **Security Dashboard**
   - Authentication events
   - API key usage
   - Rate limit violations
   - Security incidents

4. **Business Dashboard**
   - Classification requests
   - Risk assessment events
   - Compliance check events
   - User activity

5. **Infrastructure Dashboard**
   - System resource usage
   - Database connections
   - External API health
   - Application health

### Dashboard Metrics

**Key Performance Indicators**:
- Log volume (logs per minute)
- Error rate (errors per minute)
- Response time (p50, p95, p99)
- User activity (active users)
- API usage (requests per minute)

**Business Metrics**:
- Classification accuracy
- Risk assessment volume
- Compliance check status
- Authentication success rate
- API key usage patterns

**System Metrics**:
- Memory usage
- CPU utilization
- Database performance
- External API performance
- Network traffic

## Log Retention and Archival

### Retention Policy

**Log Retention Periods**:
- **Hot Storage**: Last 7 days (frequent access)
- **Warm Storage**: 7 days to 30 days (moderate access)
- **Cold Storage**: 30 days to 1 year (infrequent access)
- **Archive Storage**: 1+ years (long-term retention)

**Index Lifecycle Management**:
```json
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0ms",
        "actions": {
          "rollover": {
            "max_size": "50GB",
            "max_age": "1d"
          }
        }
      },
      "warm": {
        "min_age": "1d",
        "actions": {
          "shrink": {
            "number_of_shards": 1
          },
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "cold": {
        "min_age": "30d",
        "actions": {
          "freeze": {}
        }
      },
      "delete": {
        "min_age": "365d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

### Backup and Recovery

**Backup Strategy**:
- **Snapshot Frequency**: Daily snapshots
- **Snapshot Retention**: 30 days
- **Backup Location**: S3-compatible storage
- **Recovery Testing**: Monthly recovery tests

**Backup Configuration**:
```json
{
  "type": "s3",
  "settings": {
    "bucket": "kyb-platform-logs-backup",
    "region": "us-east-1",
    "base_path": "elasticsearch/snapshots"
  }
}
```

## Log Security

### Access Control

**Authentication**:
- Elasticsearch security enabled
- Username/password authentication
- API key authentication
- Service account authentication

**Authorization**:
- Role-based access control (RBAC)
- Index-level permissions
- Field-level security
- Document-level security

**Audit Logging**:
- Authentication events
- Authorization events
- Access denied events
- System access events

### Data Protection

**Encryption**:
- Transport layer security (TLS)
- Encryption at rest
- Field-level encryption
- API key encryption

**Data Privacy**:
- PII data masking
- Sensitive data redaction
- Data retention compliance
- GDPR compliance

## Monitoring and Alerting

### Log Monitoring

**Monitoring Metrics**:
- Log ingestion rate
- Log processing latency
- Error rates
- Storage usage
- Search performance

**Health Checks**:
- Elasticsearch cluster health
- Index health
- Shard allocation
- Node status

### Alerting Rules

**Critical Alerts**:
- Log ingestion failures
- Elasticsearch cluster down
- High error rates
- Storage capacity issues

**Warning Alerts**:
- High log volume
- Slow search performance
- Index health issues
- Backup failures

**Business Alerts**:
- Unusual error patterns
- Security incidents
- Performance degradation
- User activity anomalies

## Troubleshooting

### Common Issues

1. **Log Ingestion Failures**
   - Check network connectivity
   - Verify Elasticsearch cluster health
   - Review authentication credentials
   - Check index permissions

2. **High Log Volume**
   - Review log level configuration
   - Implement log filtering
   - Optimize log message size
   - Consider log sampling

3. **Search Performance Issues**
   - Review index mapping
   - Optimize search queries
   - Add search indices
   - Increase cluster resources

4. **Storage Issues**
   - Review retention policies
   - Implement data archival
   - Optimize index settings
   - Scale storage capacity

### Debugging Commands

```bash
# Check Elasticsearch cluster health
curl -X GET "localhost:9200/_cluster/health?pretty"

# Check index status
curl -X GET "localhost:9200/_cat/indices?v"

# Search logs
curl -X GET "localhost:9200/kyb-platform-prod/_search" \
  -H "Content-Type: application/json" \
  -d '{"query":{"match_all":{}}}'

# Check log shipping status
curl -X GET "localhost:9200/_cat/tasks?v"

# Monitor log ingestion
curl -X GET "localhost:9200/_stats/indexing?pretty"
```

## Best Practices

### Log Design

**Structured Logging**:
- Use consistent log formats
- Include correlation IDs
- Add relevant context
- Avoid sensitive data

**Log Levels**:
- DEBUG: Detailed debugging information
- INFO: General information
- WARN: Warning conditions
- ERROR: Error conditions
- FATAL: Critical errors

**Performance Considerations**:
- Minimize log message size
- Use efficient serialization
- Implement log sampling
- Optimize log shipping

### Operational Practices

**Monitoring**:
- Monitor log ingestion rates
- Track error rates
- Monitor storage usage
- Alert on failures

**Maintenance**:
- Regular index optimization
- Periodic backup testing
- Log retention cleanup
- Performance tuning

**Security**:
- Secure log transmission
- Encrypt sensitive data
- Implement access controls
- Regular security audits

---

This documentation provides a comprehensive overview of the KYB Platform's log aggregation system. For specific implementation details, refer to the log aggregation code and configuration files referenced throughout this document.
