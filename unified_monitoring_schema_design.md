# Unified Monitoring Schema Design

## Design Principles

### 1. Single Source of Truth
- **Unified Data Model**: All monitoring data flows through a single, consistent schema
- **Eliminate Redundancy**: Remove duplicate data collection and storage
- **Consistent Metrics**: Standardized metric collection across all components

### 2. Scalability and Performance
- **Horizontal Scaling**: Schema designed to handle growth in data volume
- **Query Optimization**: Optimized for common monitoring queries and dashboards
- **Efficient Storage**: Minimize storage overhead while maintaining data integrity

### 3. Flexibility and Extensibility
- **JSONB Fields**: Flexible metric storage for different component types
- **Tagging System**: Flexible categorization and filtering of metrics
- **Component Agnostic**: Schema works for any system component

### 4. Observability and Debugging
- **Trace Correlation**: Link metrics to specific requests and operations
- **Context Preservation**: Maintain context for debugging and analysis
- **Historical Tracking**: Support for trend analysis and historical reporting

## Core Schema Design

### 1. Unified Performance Metrics Table

```sql
-- Core metrics table - single source of truth for all performance data
CREATE TABLE unified_performance_metrics (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Component identification
    component VARCHAR(100) NOT NULL,           -- 'api', 'classification', 'cache', 'database', etc.
    component_instance VARCHAR(100),           -- Specific instance identifier
    service_name VARCHAR(100) NOT NULL,        -- Service or module name
    
    -- Metric categorization
    metric_type VARCHAR(50) NOT NULL,          -- 'performance', 'resource', 'business', 'security'
    metric_category VARCHAR(50) NOT NULL,      -- 'latency', 'throughput', 'error_rate', 'memory', etc.
    metric_name VARCHAR(100) NOT NULL,         -- Specific metric name
    
    -- Metric data
    metric_value DECIMAL(20,6) NOT NULL,       -- Numeric metric value
    metric_unit VARCHAR(20),                   -- 'ms', 'bytes', 'count', 'percent', etc.
    
    -- Additional context
    tags JSONB,                                -- Flexible key-value metadata
    metadata JSONB,                            -- Additional metric-specific data
    
    -- Request/operation context
    request_id UUID,                           -- Link to specific request
    operation_id UUID,                         -- Link to specific operation
    user_id UUID,                              -- Link to user (if applicable)
    
    -- Data quality
    confidence_score DECIMAL(3,2),             -- Data quality confidence (0.0-1.0)
    data_source VARCHAR(50) NOT NULL,          -- Source of the metric data
    
    -- Indexing and partitioning
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Constraints
    CONSTRAINT valid_metric_value CHECK (metric_value >= 0),
    CONSTRAINT valid_confidence CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0)
);

-- Partitioning by time for performance
CREATE INDEX idx_unified_metrics_timestamp ON unified_performance_metrics (timestamp);
CREATE INDEX idx_unified_metrics_component ON unified_performance_metrics (component, metric_type);
CREATE INDEX idx_unified_metrics_request ON unified_performance_metrics (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_unified_metrics_tags ON unified_performance_metrics USING GIN (tags);
CREATE INDEX idx_unified_metrics_metadata ON unified_performance_metrics USING GIN (metadata);

-- Composite indexes for common queries
CREATE INDEX idx_unified_metrics_component_time ON unified_performance_metrics (component, timestamp);
CREATE INDEX idx_unified_metrics_type_time ON unified_performance_metrics (metric_type, timestamp);
CREATE INDEX idx_unified_metrics_category_time ON unified_performance_metrics (metric_category, timestamp);
```

### 2. Unified Performance Alerts Table

```sql
-- Centralized alerting system
CREATE TABLE unified_performance_alerts (
    -- Primary identification
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Alert identification
    alert_type VARCHAR(50) NOT NULL,           -- 'threshold', 'anomaly', 'trend', 'availability'
    alert_category VARCHAR(50) NOT NULL,       -- 'performance', 'resource', 'business', 'security'
    severity VARCHAR(20) NOT NULL,             -- 'critical', 'warning', 'info'
    
    -- Component context
    component VARCHAR(100) NOT NULL,
    component_instance VARCHAR(100),
    service_name VARCHAR(100) NOT NULL,
    
    -- Alert details
    alert_name VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    condition JSONB NOT NULL,                  -- Alert condition definition
    current_value DECIMAL(20,6),               -- Current metric value
    threshold_value DECIMAL(20,6),             -- Threshold that triggered alert
    
    -- Alert state
    status VARCHAR(20) DEFAULT 'active' NOT NULL, -- 'active', 'acknowledged', 'resolved', 'suppressed'
    acknowledged_by UUID,                       -- User who acknowledged
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    -- Related data
    related_metrics UUID[],                    -- Array of related metric IDs
    related_requests UUID[],                   -- Array of related request IDs
    
    -- Alert metadata
    tags JSONB,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT valid_severity CHECK (severity IN ('critical', 'warning', 'info')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'acknowledged', 'resolved', 'suppressed'))
);

-- Indexes for alert management
CREATE INDEX idx_alerts_status ON unified_performance_alerts (status);
CREATE INDEX idx_alerts_severity ON unified_performance_alerts (severity);
CREATE INDEX idx_alerts_component ON unified_performance_alerts (component);
CREATE INDEX idx_alerts_created ON unified_performance_alerts (created_at);
CREATE INDEX idx_alerts_type ON unified_performance_alerts (alert_type);
CREATE INDEX idx_alerts_tags ON unified_performance_alerts USING GIN (tags);
```

### 3. Performance Health Scores Table

```sql
-- Aggregated health scores for components and services
CREATE TABLE performance_health_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Component identification
    component VARCHAR(100) NOT NULL,
    component_instance VARCHAR(100),
    service_name VARCHAR(100) NOT NULL,
    
    -- Health scores (0.0 - 1.0)
    overall_health DECIMAL(3,2) NOT NULL,      -- Overall component health
    performance_health DECIMAL(3,2) NOT NULL,  -- Performance-specific health
    resource_health DECIMAL(3,2) NOT NULL,     -- Resource utilization health
    availability_health DECIMAL(3,2) NOT NULL, -- Availability health
    security_health DECIMAL(3,2) NOT NULL,     -- Security health
    
    -- Health indicators
    active_alerts INTEGER DEFAULT 0,           -- Number of active alerts
    critical_alerts INTEGER DEFAULT 0,         -- Number of critical alerts
    warning_alerts INTEGER DEFAULT 0,          -- Number of warning alerts
    
    -- Performance indicators
    avg_response_time DECIMAL(10,3),           -- Average response time
    error_rate DECIMAL(5,4),                   -- Error rate percentage
    throughput DECIMAL(10,2),                  -- Requests per second
    
    -- Resource indicators
    cpu_usage DECIMAL(5,2),                    -- CPU usage percentage
    memory_usage DECIMAL(5,2),                 -- Memory usage percentage
    disk_usage DECIMAL(5,2),                   -- Disk usage percentage
    
    -- Metadata
    tags JSONB,
    metadata JSONB,
    
    -- Constraints
    CONSTRAINT valid_health_scores CHECK (
        overall_health >= 0.0 AND overall_health <= 1.0 AND
        performance_health >= 0.0 AND performance_health <= 1.0 AND
        resource_health >= 0.0 AND resource_health <= 1.0 AND
        availability_health >= 0.0 AND availability_health <= 1.0 AND
        security_health >= 0.0 AND security_health <= 1.0
    )
);

-- Indexes for health score queries
CREATE INDEX idx_health_scores_timestamp ON performance_health_scores (timestamp);
CREATE INDEX idx_health_scores_component ON performance_health_scores (component);
CREATE INDEX idx_health_scores_overall ON performance_health_scores (overall_health);
CREATE INDEX idx_health_scores_service ON performance_health_scores (service_name);
```

### 4. Performance Trends Table

```sql
-- Aggregated trend data for dashboards and reporting
CREATE TABLE performance_trends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Time aggregation
    time_bucket TIMESTAMP WITH TIME ZONE NOT NULL, -- Aggregated time bucket
    aggregation_period VARCHAR(20) NOT NULL,       -- 'minute', 'hour', 'day'
    
    -- Component identification
    component VARCHAR(100) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    
    -- Aggregated metrics
    metric_type VARCHAR(50) NOT NULL,
    metric_category VARCHAR(50) NOT NULL,
    
    -- Statistical aggregations
    count_metrics INTEGER NOT NULL,             -- Number of metrics aggregated
    min_value DECIMAL(20,6),                    -- Minimum value
    max_value DECIMAL(20,6),                    -- Maximum value
    avg_value DECIMAL(20,6),                    -- Average value
    median_value DECIMAL(20,6),                 -- Median value
    p95_value DECIMAL(20,6),                    -- 95th percentile
    p99_value DECIMAL(20,6),                    -- 99th percentile
    
    -- Additional aggregations
    sum_value DECIMAL(20,6),                    -- Sum of values
    std_dev DECIMAL(20,6),                      -- Standard deviation
    
    -- Metadata
    tags JSONB,
    
    -- Constraints
    CONSTRAINT valid_aggregation_period CHECK (aggregation_period IN ('minute', 'hour', 'day')),
    CONSTRAINT valid_count CHECK (count_metrics > 0)
);

-- Indexes for trend queries
CREATE INDEX idx_trends_time_bucket ON performance_trends (time_bucket);
CREATE INDEX idx_trends_component ON performance_trends (component);
CREATE INDEX idx_trends_metric_type ON performance_trends (metric_type);
CREATE INDEX idx_trends_period ON performance_trends (aggregation_period);
```

## Go Code Architecture

### 1. Unified Performance Monitor

```go
// Unified monitoring component
type UnifiedPerformanceMonitor struct {
    config     *MonitoringConfig
    db         *sql.DB
    exporters  []MetricExporter
    alerters   []AlertHandler
    collectors map[string]MetricCollector
    logger     *zap.Logger
    metrics    *prometheus.Registry
}

// Configuration for unified monitoring
type MonitoringConfig struct {
    DatabaseURL        string
    CollectionInterval time.Duration
    RetentionPeriod    time.Duration
    AlertThresholds    map[string]AlertThreshold
    Exporters          []ExporterConfig
    Components         []ComponentConfig
}

// Metric collector interface
type MetricCollector interface {
    Collect(ctx context.Context) ([]Metric, error)
    GetType() string
    GetComponent() string
    GetInterval() time.Duration
}

// Metric data structure
type Metric struct {
    ID               string                 `json:"id"`
    Timestamp        time.Time              `json:"timestamp"`
    Component        string                 `json:"component"`
    ComponentInstance string                `json:"component_instance"`
    ServiceName      string                 `json:"service_name"`
    MetricType       string                 `json:"metric_type"`
    MetricCategory   string                 `json:"metric_category"`
    MetricName       string                 `json:"metric_name"`
    MetricValue      float64                `json:"metric_value"`
    MetricUnit       string                 `json:"metric_unit"`
    Tags             map[string]string      `json:"tags"`
    Metadata         map[string]interface{} `json:"metadata"`
    RequestID        *string                `json:"request_id"`
    OperationID      *string                `json:"operation_id"`
    UserID           *string                `json:"user_id"`
    ConfidenceScore  *float64               `json:"confidence_score"`
    DataSource       string                 `json:"data_source"`
}
```

### 2. Specialized Collectors

```go
// HTTP performance collector
type HTTPPerformanceCollector struct {
    serviceName string
    config      *HTTPCollectorConfig
}

func (c *HTTPPerformanceCollector) Collect(ctx context.Context) ([]Metric, error) {
    // Collect HTTP-specific metrics
    return []Metric{
        {
            Component:      "api",
            ServiceName:    c.serviceName,
            MetricType:     "performance",
            MetricCategory: "latency",
            MetricName:     "response_time",
            MetricValue:    responseTime,
            MetricUnit:     "ms",
            Tags: map[string]string{
                "endpoint": endpoint,
                "method":   method,
                "status":   statusCode,
            },
        },
    }, nil
}

// Database performance collector
type DatabasePerformanceCollector struct {
    serviceName string
    config      *DatabaseCollectorConfig
}

func (c *DatabasePerformanceCollector) Collect(ctx context.Context) ([]Metric, error) {
    // Collect database-specific metrics
    return []Metric{
        {
            Component:      "database",
            ServiceName:    c.serviceName,
            MetricType:     "performance",
            MetricCategory: "query_time",
            MetricName:     "query_duration",
            MetricValue:    queryDuration,
            MetricUnit:     "ms",
            Tags: map[string]string{
                "query_type": queryType,
                "table":      tableName,
            },
        },
    }, nil
}

// Classification performance collector
type ClassificationPerformanceCollector struct {
    serviceName string
    config      *ClassificationCollectorConfig
}

func (c *ClassificationPerformanceCollector) Collect(ctx context.Context) ([]Metric, error) {
    // Collect classification-specific metrics
    return []Metric{
        {
            Component:      "classification",
            ServiceName:    c.serviceName,
            MetricType:     "business",
            MetricCategory: "accuracy",
            MetricName:     "classification_accuracy",
            MetricValue:    accuracy,
            MetricUnit:     "percent",
            Tags: map[string]string{
                "model":      modelName,
                "algorithm":  algorithm,
            },
        },
    }, nil
}
```

### 3. Alert Management System

```go
// Alert manager for unified alerting
type UnifiedAlertManager struct {
    db     *sql.DB
    config *AlertConfig
    logger *zap.Logger
}

// Alert configuration
type AlertConfig struct {
    Thresholds map[string]AlertThreshold
    Rules      []AlertRule
    Channels   []AlertChannel
}

// Alert threshold definition
type AlertThreshold struct {
    MetricName    string
    Component     string
    Condition     string // 'gt', 'lt', 'eq', 'ne'
    Value         float64
    Severity      string
    Duration      time.Duration
    Cooldown      time.Duration
}

// Alert rule definition
type AlertRule struct {
    Name        string
    Description string
    Condition   string
    Severity    string
    Actions     []AlertAction
}

// Alert action definition
type AlertAction struct {
    Type    string // 'email', 'webhook', 'slack'
    Config  map[string]interface{}
    Enabled bool
}
```

## Data Flow Architecture

### 1. Metric Collection Flow

```
Component → MetricCollector → UnifiedPerformanceMonitor → Database
    ↓
MetricExporter → External Systems (Prometheus, Grafana, etc.)
```

### 2. Alert Processing Flow

```
Metric Data → AlertManager → Alert Evaluation → Alert Generation
    ↓
Alert Storage → Alert Notification → Alert Resolution
```

### 3. Dashboard Data Flow

```
Dashboard Request → Query Optimization → Data Aggregation → Response
    ↓
Caching Layer → Performance Optimization
```

## Performance Optimization

### 1. Database Optimization

#### Partitioning Strategy
```sql
-- Partition by time for better performance
CREATE TABLE unified_performance_metrics_y2024m01 
PARTITION OF unified_performance_metrics 
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Partition by component for component-specific queries
CREATE TABLE unified_performance_metrics_api 
PARTITION OF unified_performance_metrics 
FOR VALUES IN ('api');
```

#### Indexing Strategy
```sql
-- Time-based indexes for time-series queries
CREATE INDEX idx_metrics_time_component ON unified_performance_metrics (timestamp, component);

-- Component-based indexes for component queries
CREATE INDEX idx_metrics_component_type ON unified_performance_metrics (component, metric_type);

-- Request-based indexes for request tracing
CREATE INDEX idx_metrics_request_time ON unified_performance_metrics (request_id, timestamp) WHERE request_id IS NOT NULL;
```

### 2. Query Optimization

#### Common Query Patterns
```sql
-- Component performance over time
SELECT 
    component,
    metric_category,
    AVG(metric_value) as avg_value,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY metric_value) as p95_value
FROM unified_performance_metrics 
WHERE timestamp >= NOW() - INTERVAL '1 hour'
  AND component = 'api'
GROUP BY component, metric_category;

-- Health score calculation
SELECT 
    component,
    AVG(CASE WHEN metric_category = 'latency' THEN metric_value END) as avg_latency,
    AVG(CASE WHEN metric_category = 'error_rate' THEN metric_value END) as avg_error_rate
FROM unified_performance_metrics 
WHERE timestamp >= NOW() - INTERVAL '5 minutes'
GROUP BY component;
```

### 3. Caching Strategy

#### Redis Caching
```go
// Cache frequently accessed data
type MonitoringCache struct {
    redis *redis.Client
    ttl   time.Duration
}

func (c *MonitoringCache) GetHealthScores(component string) ([]HealthScore, error) {
    // Check cache first
    cached, err := c.redis.Get(ctx, fmt.Sprintf("health:%s", component)).Result()
    if err == nil {
        var scores []HealthScore
        json.Unmarshal([]byte(cached), &scores)
        return scores, nil
    }
    
    // Fetch from database and cache
    scores, err := c.fetchHealthScoresFromDB(component)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 minute
    data, _ := json.Marshal(scores)
    c.redis.Set(ctx, fmt.Sprintf("health:%s", component), data, c.ttl)
    
    return scores, nil
}
```

## Migration Strategy

### 1. Phase 1: Implementation
- Implement unified schema alongside existing systems
- Create unified monitoring components
- Set up data collection pipelines

### 2. Phase 2: Data Migration
- Migrate historical data to unified schema
- Validate data integrity
- Set up monitoring for migration process

### 3. Phase 3: Cutover
- Switch applications to use unified monitoring
- Monitor performance and data quality
- Gradually deprecate old systems

### 4. Phase 4: Cleanup
- Remove redundant tables and code
- Optimize unified system performance
- Document new monitoring system

## Success Metrics

### 1. Performance Metrics
- **Query Performance**: 50% improvement in dashboard query times
- **Storage Efficiency**: 60% reduction in monitoring storage usage
- **Write Performance**: 30% improvement in metric collection performance

### 2. Operational Metrics
- **Maintenance Overhead**: 70% reduction in monitoring maintenance
- **Alert Accuracy**: 90% reduction in duplicate alerts
- **Data Consistency**: 99.9% data consistency across monitoring systems

### 3. Business Metrics
- **System Reliability**: Improved system reliability through better monitoring
- **Development Velocity**: Faster development through simplified monitoring
- **Cost Reduction**: Reduced infrastructure costs through optimized monitoring

## Conclusion

The unified monitoring schema design provides a comprehensive solution for consolidating the current fragmented monitoring infrastructure. The design emphasizes:

1. **Single Source of Truth**: All monitoring data flows through unified tables
2. **Performance Optimization**: Optimized for common monitoring queries and operations
3. **Flexibility**: JSONB fields and tagging system provide flexibility for different metric types
4. **Scalability**: Designed to handle growth in data volume and complexity
5. **Maintainability**: Simplified architecture reduces maintenance overhead

The proposed schema eliminates redundancy while maintaining all existing monitoring capabilities and provides a foundation for future monitoring enhancements.
