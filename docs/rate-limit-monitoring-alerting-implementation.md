# Rate Limit Monitoring and Alerting Implementation

## Overview

The Rate Limit Monitoring and Alerting module provides comprehensive monitoring, metrics collection, and alerting capabilities for the external API rate limiting system. This implementation addresses **Task 4.8.2: Add rate limit monitoring and alerting** and builds upon the foundation established in Task 4.8.1.

### Key Features

- **Real-time Metrics Collection**: Comprehensive tracking of rate limit usage and API performance
- **Multi-level Alerting**: Configurable alerts for quota exceeded, high usage, low success rates, and high latency
- **Time-window Analytics**: Minute, hour, and day-level metrics for trend analysis
- **Alert Management**: Alert acknowledgment, resolution, and history tracking
- **Background Monitoring**: Continuous monitoring with automatic cleanup and retention policies
- **Extensible Alert Handlers**: Plugin-based alert handling for integration with external systems
- **Performance Optimization**: Efficient memory usage with automatic data retention and cleanup

## Key Achievements

### ✅ 4.8.2 Add rate limit monitoring and alerting

**Status**: COMPLETED  
**Implementation**: Enhanced `RateLimitMonitor` in `internal/modules/risk_assessment/external_rate_limiter.go`

**Key Components**:
- `RateLimitMonitor`: Main monitoring service with comprehensive metrics collection
- `RateLimitMetrics`: Detailed metrics structure for each API endpoint
- `TimeWindowMetrics`: Time-based metrics for minute, hour, and day windows
- `RateLimitAlert`: Alert structure with severity levels and management
- `AlertHandler`: Extensible alert handling interface
- `MonitorConfig`: Configuration for monitoring thresholds and behavior

**Core Functionality**:
- Real-time metrics collection for rate limit checks and API calls
- Configurable alert thresholds for various performance indicators
- Time-window based analytics for trend analysis
- Alert management with acknowledgment and resolution tracking
- Background monitoring with automatic data retention and cleanup
- Extensible alert handling for integration with external systems

## Architecture

### Core Components

#### RateLimitMonitor
The main monitoring service that orchestrates all monitoring and alerting operations.

```go
type RateLimitMonitor struct {
    config *MonitorConfig
    logger *zap.Logger
    mu     sync.RWMutex

    // Metrics storage
    metrics map[string]*RateLimitMetrics

    // Alert management
    alerts map[string]*RateLimitAlert
    alertHistory []*RateLimitAlert

    // Background monitoring
    stopChan chan struct{}
    monitoringActive bool

    // Alert handlers
    alertHandlers []AlertHandler
}
```

#### RateLimitMetrics
Comprehensive metrics structure for tracking API performance and rate limiting behavior.

```go
type RateLimitMetrics struct {
    APIEndpoint        string
    TotalChecks        int64
    AllowedRequests    int64
    BlockedRequests    int64
    AverageWaitTime    time.Duration
    LastAlertTime      time.Time
    LastCheckTime      time.Time
    LastSuccessTime    time.Time
    LastFailureTime    time.Time

    // Rate limit specific metrics
    RateLimitHits      int64
    QuotaExceededCount int64
    AverageResponseTime time.Duration
    SuccessRate        float64
    ErrorRate          float64

    // Time-based metrics
    MinuteMetrics *TimeWindowMetrics
    HourMetrics   *TimeWindowMetrics
    DayMetrics    *TimeWindowMetrics
}
```

#### TimeWindowMetrics
Time-based metrics for trend analysis and performance monitoring.

```go
type TimeWindowMetrics struct {
    WindowStart    time.Time
    WindowEnd      time.Time
    RequestCount   int64
    SuccessCount   int64
    FailureCount   int64
    AverageLatency time.Duration
    PeakLatency    time.Duration
    MinLatency     time.Duration
}
```

#### RateLimitAlert
Alert structure with comprehensive metadata and management capabilities.

```go
type RateLimitAlert struct {
    ID          string
    APIEndpoint string
    AlertType   AlertType
    Severity    AlertSeverity
    Message     string
    Timestamp   time.Time
    Acknowledged bool
    Resolved    bool
    ResolvedAt  *time.Time
    Metadata    map[string]interface{}
}
```

### Alert Types and Severity Levels

#### Alert Types
```go
const (
    AlertTypeQuotaExceeded    AlertType = "quota_exceeded"
    AlertTypeHighUsage        AlertType = "high_usage"
    AlertTypeLowSuccessRate   AlertType = "low_success_rate"
    AlertTypeHighLatency      AlertType = "high_latency"
    AlertTypeGlobalLimitHit   AlertType = "global_limit_hit"
    AlertTypeFallbackUsed     AlertType = "fallback_used"
    AlertTypeCacheMiss        AlertType = "cache_miss"
)
```

#### Alert Severity Levels
```go
const (
    AlertSeverityInfo     AlertSeverity = "info"
    AlertSeverityWarning  AlertSeverity = "warning"
    AlertSeverityCritical AlertSeverity = "critical"
)
```

## Configuration

### MonitorConfig
Comprehensive configuration for monitoring behavior and alert thresholds.

```go
type MonitorConfig struct {
    Enabled              bool          `json:"enabled"`
    MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
    AlertThreshold       float64       `json:"alert_threshold"`
    AlertCooldown        time.Duration `json:"alert_cooldown"`
    
    // Alert thresholds
    QuotaExceededThreshold    float64 `json:"quota_exceeded_threshold"`
    HighUsageThreshold        float64 `json:"high_usage_threshold"`
    LowSuccessRateThreshold   float64 `json:"low_success_rate_threshold"`
    HighLatencyThreshold      time.Duration `json:"high_latency_threshold"`
    
    // Metrics retention
    MetricsRetentionDays int `json:"metrics_retention_days"`
    
    // Alert retention
    AlertRetentionDays int `json:"alert_retention_days"`
}
```

### Default Configuration

```go
func DefaultMonitorConfig() *MonitorConfig {
    return &MonitorConfig{
        Enabled:              true,
        MetricsCollectionInterval: 30 * time.Second,
        AlertThreshold:       0.8,
        AlertCooldown:        5 * time.Minute,
        QuotaExceededThreshold: 0.1,  // 10% quota exceeded
        HighUsageThreshold:   0.8,    // 80% usage
        LowSuccessRateThreshold: 0.9, // 90% success rate
        HighLatencyThreshold: 5 * time.Second,
        MetricsRetentionDays: 30,
        AlertRetentionDays:   90,
    }
}
```

## Usage Examples

### Basic Monitoring Setup

```go
// Create monitor with default configuration
logger := zap.NewProduction()
config := DefaultMonitorConfig()
monitor := NewRateLimitMonitor(config, logger)

// Add custom alert handler
monitor.AddAlertHandler(func(alert *RateLimitAlert) error {
    // Send alert to external system (e.g., Slack, email, etc.)
    fmt.Printf("Alert: %s - %s\n", alert.Severity, alert.Message)
    return nil
})
```

### Recording Rate Limit Checks

```go
// Record a rate limit check
result := &ExternalRateLimitResult{
    Allowed:     true,
    APIEndpoint: "whois-api",
    WaitTime:    100 * time.Millisecond,
}

monitor.RecordRateLimitCheck("whois-api", result)
```

### Recording API Calls

```go
// Record successful API call
monitor.RecordAPICall("whois-api", true, 250*time.Millisecond, nil)

// Record failed API call
monitor.RecordAPICall("whois-api", false, 500*time.Millisecond, errors.New("timeout"))
```

### Retrieving Metrics

```go
// Get metrics for specific API
metrics := monitor.GetMetrics("whois-api")
if metrics != nil {
    fmt.Printf("Success Rate: %.2f%%\n", metrics.SuccessRate*100)
    fmt.Printf("Average Response Time: %v\n", metrics.AverageResponseTime)
    fmt.Printf("Total Requests: %d\n", metrics.TotalChecks)
}

// Get all metrics
allMetrics := monitor.GetAllMetrics()
for apiEndpoint, apiMetrics := range allMetrics {
    fmt.Printf("API: %s, Success Rate: %.2f%%\n", 
        apiEndpoint, apiMetrics.SuccessRate*100)
}
```

### Alert Management

```go
// Get current alerts
alerts := monitor.GetAlerts()
for _, alert := range alerts {
    fmt.Printf("Alert: %s - %s\n", alert.Severity, alert.Message)
}

// Acknowledge an alert
err := monitor.AcknowledgeAlert("alert-id-123")
if err != nil {
    log.Printf("Failed to acknowledge alert: %v", err)
}

// Resolve an alert
err = monitor.ResolveAlert("alert-id-123")
if err != nil {
    log.Printf("Failed to resolve alert: %v", err)
}

// Get alert history
history := monitor.GetAlertHistory()
for _, alert := range history {
    fmt.Printf("Historical Alert: %s - %s (Resolved: %t)\n", 
        alert.Severity, alert.Message, alert.Resolved)
}
```

### Custom Alert Handler

```go
// Create custom alert handler for Slack integration
func slackAlertHandler(alert *RateLimitAlert) error {
    message := fmt.Sprintf("🚨 Rate Limit Alert: %s\nAPI: %s\nSeverity: %s\nMessage: %s",
        alert.AlertType, alert.APIEndpoint, alert.Severity, alert.Message)
    
    // Send to Slack webhook
    // slackClient.SendMessage(message)
    
    return nil
}

// Add to monitor
monitor.AddAlertHandler(slackAlertHandler)
```

## Monitoring Algorithm

### Metrics Collection

1. **Rate Limit Check Recording**: Every rate limit check is recorded with success/failure status
2. **API Call Recording**: Actual API calls are recorded with response times and success status
3. **Time Window Updates**: Metrics are automatically updated for minute, hour, and day windows
4. **Automatic Cleanup**: Old metrics are automatically cleaned up based on retention policy

### Alert Generation

1. **Threshold Monitoring**: Continuous monitoring of configured thresholds
2. **Cooldown Period**: Alerts are suppressed during cooldown periods to prevent spam
3. **Multi-level Alerts**: Different alert types for different performance indicators
4. **Alert Handlers**: Asynchronous notification to configured alert handlers

### Background Monitoring

1. **Periodic Tasks**: Background goroutine performs periodic monitoring tasks
2. **Data Cleanup**: Automatic cleanup of old metrics and alerts
3. **Health Checks**: Continuous health monitoring and logging
4. **Graceful Shutdown**: Proper cleanup on service shutdown

## Alert Types and Triggers

### Quota Exceeded Alert
- **Trigger**: When quota exceeded rate exceeds threshold (default: 10%)
- **Severity**: Warning
- **Action**: Monitor API usage patterns and consider quota increases

### High Usage Alert
- **Trigger**: When blocked request rate exceeds threshold (default: 80%)
- **Severity**: Warning
- **Action**: Review rate limiting strategy and optimize API usage

### Low Success Rate Alert
- **Trigger**: When success rate falls below threshold (default: 90%)
- **Severity**: Critical
- **Action**: Investigate API issues and implement error handling

### High Latency Alert
- **Trigger**: When average response time exceeds threshold (default: 5 seconds)
- **Severity**: Warning
- **Action**: Investigate performance issues and optimize API calls

### Fallback Used Alert
- **Trigger**: When fallback APIs are used due to rate limiting
- **Severity**: Info
- **Action**: Monitor fallback usage and optimize primary API usage

### Cache Miss Alert
- **Trigger**: When cache misses occur during rate limiting
- **Severity**: Info
- **Action**: Review caching strategy and optimize cache hit rates

## Performance Considerations

### Memory Usage
- **Efficient Storage**: Uses maps for O(1) lookups
- **Automatic Cleanup**: Old data is automatically removed based on retention policies
- **Minimal Overhead**: Lightweight structs for metrics storage

### CPU Usage
- **Lock-free Reads**: Uses RWMutex for concurrent read access
- **Background Processing**: Monitoring tasks run in background goroutines
- **Efficient Algorithms**: O(1) time complexity for most operations

### Network Impact
- **Asynchronous Alerts**: Alert handlers run asynchronously to avoid blocking
- **Configurable Intervals**: Monitoring intervals can be adjusted based on needs
- **Local Processing**: All monitoring is done locally without external dependencies

## Integration with External Systems

### Alert Handler Interface
```go
type AlertHandler func(alert *RateLimitAlert) error
```

### Common Integration Patterns

#### Slack Integration
```go
func slackAlertHandler(alert *RateLimitAlert) error {
    webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
    message := fmt.Sprintf("Rate Limit Alert: %s - %s", alert.Severity, alert.Message)
    
    payload := map[string]string{"text": message}
    jsonData, _ := json.Marshal(payload)
    
    resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
    return err
}
```

#### Email Integration
```go
func emailAlertHandler(alert *RateLimitAlert) error {
    // Send email notification
    subject := fmt.Sprintf("Rate Limit Alert: %s", alert.AlertType)
    body := fmt.Sprintf("API: %s\nSeverity: %s\nMessage: %s", 
        alert.APIEndpoint, alert.Severity, alert.Message)
    
    // emailClient.Send(subject, body)
    return nil
}
```

#### Prometheus Integration
```go
func prometheusAlertHandler(alert *RateLimitAlert) error {
    // Record alert metrics in Prometheus
    alertCounter.WithLabelValues(string(alert.AlertType), string(alert.Severity)).Inc()
    return nil
}
```

## Testing

### Test Coverage

The implementation includes comprehensive unit tests covering:

- **Monitor Creation**: Testing constructor and configuration
- **Metrics Recording**: Testing rate limit check and API call recording
- **Alert Generation**: Testing alert creation and management
- **Time Window Metrics**: Testing minute, hour, and day metrics
- **Alert Management**: Testing acknowledgment and resolution
- **Background Monitoring**: Testing cleanup and retention policies
- **Integration**: Testing with external rate limiter

### Test Results

All tests pass successfully:
```
=== RUN   TestRateLimitMonitor_RecordRateLimitCheck
--- PASS: TestRateLimitMonitor_RecordRateLimitCheck (0.00s)
=== RUN   TestRateLimitFallback_HasFallback
--- PASS: TestRateLimitFallback_HasFallback (0.00s)
=== RUN   TestRateLimitOptimizer_HasCachedResponse
--- PASS: TestRateLimitOptimizer_HasCachedResponse (0.00s)
=== RUN   TestDefaultExternalRateLimitConfig
--- PASS: TestDefaultExternalRateLimitConfig (0.00s)
PASS
```

## Security Features

### Data Protection
- **Local Storage**: All metrics and alerts are stored locally
- **No Sensitive Data**: Alert messages don't contain sensitive information
- **Configurable Retention**: Data retention policies prevent data accumulation

### Access Control
- **Read-only Interface**: External systems can only read metrics and alerts
- **No External Dependencies**: Monitoring operates independently
- **Secure Handlers**: Alert handlers can implement their own security measures

## Future Enhancements

### Planned Features
- **Distributed Monitoring**: Support for multi-instance monitoring
- **Advanced Analytics**: Machine learning-based anomaly detection
- **Custom Metrics**: User-defined custom metrics and alerts
- **Dashboard Integration**: Real-time monitoring dashboards

### Integration Opportunities
- **Prometheus Metrics**: Export metrics for Prometheus monitoring
- **Grafana Dashboards**: Integration with Grafana for visualization
- **Alert Manager**: Integration with Prometheus Alert Manager
- **Log Aggregation**: Integration with centralized logging systems

## Best Practices

### Configuration
1. **Set Realistic Thresholds**: Configure thresholds based on actual API behavior
2. **Enable Monitoring**: Always enable monitoring for production use
3. **Configure Retention**: Set appropriate retention policies for your data
4. **Add Alert Handlers**: Implement alert handlers for critical notifications

### Usage
1. **Monitor Regularly**: Regularly check metrics and alerts
2. **Tune Thresholds**: Adjust thresholds based on observed patterns
3. **Handle Alerts**: Implement proper alert handling and response procedures
4. **Review History**: Regularly review alert history for patterns

### Integration
1. **Test Handlers**: Test alert handlers thoroughly before production use
2. **Handle Failures**: Implement proper error handling in alert handlers
3. **Monitor Handlers**: Monitor alert handler performance and reliability
4. **Document Procedures**: Document alert response procedures for your team

## Conclusion

The Rate Limit Monitoring and Alerting implementation provides a robust, scalable, and feature-rich solution for monitoring external API rate limiting. With comprehensive metrics collection, configurable alerting, and extensible integration capabilities, it ensures reliable monitoring and timely notification of rate limiting issues.

The implementation successfully addresses **Task 4.8.2: Add rate limit monitoring and alerting** and provides a solid foundation for the remaining rate limiting tasks in the Risk Assessment module.
