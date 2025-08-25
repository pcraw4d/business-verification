# Security Monitoring Documentation

## Overview

The Security Monitoring system provides comprehensive monitoring, alerting, and analytics for security events across the KYB Platform. This system tracks security-related activities, generates alerts based on configurable thresholds, and provides detailed metrics for security analysis and incident response.

## Architecture

### Core Components

1. **SecurityMonitor**: Main monitoring component that processes events and generates alerts
2. **SecurityEvent**: Represents individual security events with detailed metadata
3. **SecurityAlert**: Represents alerts that can be sent to external systems
4. **SecurityMetrics**: Provides aggregated metrics and analytics
5. **EventFilters/AlertFilters**: Enable filtering and querying of events and alerts

### Event Types

The system tracks various types of security events:

#### Authentication Events
- `login_attempt`: User login attempts
- `login_success`: Successful logins
- `login_failure`: Failed login attempts
- `logout`: User logout events
- `password_change`: Password change events
- `password_reset`: Password reset events
- `account_lockout`: Account lockout events
- `account_unlock`: Account unlock events

#### Authorization Events
- `access_denied`: Access denied events
- `permission_denied`: Permission denied events
- `role_change`: Role change events
- `permission_change`: Permission change events

#### Input Validation Events
- `invalid_input`: Invalid input detection
- `sql_injection_attempt`: SQL injection attempts
- `xss_attempt`: Cross-site scripting attempts
- `path_traversal_attempt`: Path traversal attempts

#### Rate Limiting Events
- `rate_limit_exceeded`: General rate limit exceeded
- `auth_rate_limit_exceeded`: Authentication rate limit exceeded

#### API Security Events
- `invalid_api_key`: Invalid API key usage
- `expired_token`: Expired token usage
- `invalid_token`: Invalid token usage
- `token_refresh`: Token refresh events

#### System Security Events
- `security_header_violation`: Security header violations
- `csp_violation`: Content Security Policy violations
- `hsts_violation`: HTTP Strict Transport Security violations

#### Data Security Events
- `data_access`: Data access events
- `data_modification`: Data modification events
- `data_export`: Data export events
- `data_deletion`: Data deletion events

#### System Events
- `system_startup`: System startup events
- `system_shutdown`: System shutdown events
- `configuration_change`: Configuration change events
- `backup_completed`: Backup completion events
- `backup_failed`: Backup failure events

### Severity Levels

Events are categorized by severity:

- **Info**: Informational events (e.g., successful logins)
- **Low**: Minor security events (e.g., failed login attempts)
- **Medium**: Moderate security events (e.g., rate limit exceeded)
- **High**: High-priority security events (e.g., multiple failed logins)
- **Critical**: Critical security events (e.g., SQL injection attempts)

## Configuration

### SecurityMonitorConfig Structure

```go
type SecurityMonitorConfig struct {
    // Event Storage
    MaxEvents         int           `json:"max_events" yaml:"max_events"`
    EventRetention    time.Duration `json:"event_retention" yaml:"event_retention"`
    
    // Alerting
    AlertThresholds   map[SecurityEventSeverity]int `json:"alert_thresholds" yaml:"alert_thresholds"`
    AlertCooldown     time.Duration                 `json:"alert_cooldown" yaml:"alert_cooldown"`
    
    // Metrics
    MetricsInterval   time.Duration `json:"metrics_interval" yaml:"metrics_interval"`
    
    // External Integrations
    WebhookURL        string        `json:"webhook_url" yaml:"webhook_url"`
    WebhookTimeout    time.Duration `json:"webhook_timeout" yaml:"webhook_timeout"`
    
    // Filtering
    ExcludeSources    []string      `json:"exclude_sources" yaml:"exclude_sources"`
    ExcludeEventTypes []SecurityEventType `json:"exclude_event_types" yaml:"exclude_event_types"`
}
```

### Default Configuration

```go
config := &SecurityMonitorConfig{
    MaxEvents:       10000,
    EventRetention:  30 * 24 * time.Hour, // 30 days
    AlertThresholds: map[SecurityEventSeverity]int{
        SeverityCritical: 1,
        SeverityHigh:     5,
        SeverityMedium:   10,
        SeverityLow:      50,
    },
    AlertCooldown:   5 * time.Minute,
    MetricsInterval: 1 * time.Minute,
    WebhookTimeout:  10 * time.Second,
}
```

## Usage

### Basic Usage

```go
package main

import (
    "time"
    "go.uber.org/zap"
    "github.com/your-org/kyb-platform/internal/security/monitoring"
)

func main() {
    logger := zap.NewProduction()
    
    // Create security monitor with default configuration
    monitor := monitoring.NewSecurityMonitor(nil, logger)
    defer monitor.Stop()
    
    // Record a security event
    event := &monitoring.SecurityEvent{
        Type:     monitoring.EventTypeLoginFailure,
        Severity: monitoring.SeverityMedium,
        Source:   "auth",
        UserID:   "user123",
        IPAddress: "192.168.1.1",
        UserAgent: "Mozilla/5.0...",
        Endpoint: "/api/auth/login",
        Method:   "POST",
        Details: map[string]interface{}{
            "reason": "invalid_password",
            "attempt_count": 3,
        },
    }
    
    err := monitor.RecordEvent(event)
    if err != nil {
        logger.Error("failed to record security event", zap.Error(err))
    }
}
```

### Custom Configuration

```go
config := &monitoring.SecurityMonitorConfig{
    MaxEvents:       5000,
    EventRetention:  7 * 24 * time.Hour, // 7 days
    AlertThresholds: map[monitoring.SecurityEventSeverity]int{
        monitoring.SeverityCritical: 1,
        monitoring.SeverityHigh:     3,
        monitoring.SeverityMedium:   5,
    },
    AlertCooldown:   2 * time.Minute,
    MetricsInterval: 30 * time.Second,
    WebhookURL:      "https://alerts.example.com/webhook",
    WebhookTimeout:  5 * time.Second,
    ExcludeSources:  []string{"health_check", "metrics"},
    ExcludeEventTypes: []monitoring.SecurityEventType{
        monitoring.EventTypeSystemStartup,
        monitoring.EventTypeSystemShutdown,
    },
}

monitor := monitoring.NewSecurityMonitor(config, logger)
```

### Event Callbacks

```go
// Set up event callback
monitor.SetEventCallback(func(event *monitoring.SecurityEvent) {
    // Custom event processing
    logger.Info("security event received",
        zap.String("type", string(event.Type)),
        zap.String("severity", string(event.Severity)),
        zap.String("source", event.Source))
    
    // Send to external systems
    sendToSIEM(event)
})

// Set up alert callback
monitor.SetAlertCallback(func(alert *monitoring.SecurityAlert) {
    // Custom alert processing
    logger.Warn("security alert generated",
        zap.String("title", alert.Title),
        zap.String("severity", string(alert.Severity)))
    
    // Send to notification system
    sendNotification(alert)
})
```

## Event Management

### Recording Events

```go
// Authentication event
authEvent := &monitoring.SecurityEvent{
    Type:     monitoring.EventTypeLoginFailure,
    Severity: monitoring.SeverityMedium,
    Source:   "auth",
    UserID:   "user123",
    IPAddress: "192.168.1.1",
    UserAgent: r.UserAgent(),
    Endpoint: r.URL.Path,
    Method:   r.Method,
    Details: map[string]interface{}{
        "reason": "invalid_password",
        "attempt_count": 3,
        "lockout_threshold": 5,
    },
}

err := monitor.RecordEvent(authEvent)
```

### Querying Events

```go
// Get all events
events, err := monitor.GetEvents(monitoring.EventFilters{})
if err != nil {
    logger.Error("failed to get events", zap.Error(err))
}

// Filter events by type
events, err := monitor.GetEvents(monitoring.EventFilters{
    Types: []monitoring.SecurityEventType{
        monitoring.EventTypeLoginFailure,
        monitoring.EventTypeSQLInjectionAttempt,
    },
})

// Filter events by severity
events, err := monitor.GetEvents(monitoring.EventFilters{
    Severities: []monitoring.SecurityEventSeverity{
        monitoring.SeverityHigh,
        monitoring.SeverityCritical,
    },
})

// Filter events by time range
startTime := time.Now().Add(-24 * time.Hour)
endTime := time.Now()
events, err := monitor.GetEvents(monitoring.EventFilters{
    StartTime: &startTime,
    EndTime:   &endTime,
})

// Filter events by IP address
events, err := monitor.GetEvents(monitoring.EventFilters{
    IPAddresses: []string{"192.168.1.1", "192.168.1.2"},
})

// Filter events by user ID
events, err := monitor.GetEvents(monitoring.EventFilters{
    UserIDs: []string{"user123", "user456"},
})

// Filter resolved events
resolved := true
events, err := monitor.GetEvents(monitoring.EventFilters{
    Resolved: &resolved,
})
```

### Resolving Events

```go
// Resolve a security event
err := monitor.ResolveEvent(
    "evt_1234567890",
    "admin",
    "Investigated and determined to be false positive",
)
if err != nil {
    logger.Error("failed to resolve event", zap.Error(err))
}
```

## Alert Management

### Alert Generation

Alerts are automatically generated based on configurable thresholds:

```go
config := &monitoring.SecurityMonitorConfig{
    AlertThresholds: map[monitoring.SecurityEventSeverity]int{
        monitoring.SeverityCritical: 1,  // Alert on first critical event
        monitoring.SeverityHigh:     3,  // Alert after 3 high events
        monitoring.SeverityMedium:   10, // Alert after 10 medium events
        monitoring.SeverityLow:      50, // Alert after 50 low events
    },
    AlertCooldown: 5 * time.Minute, // Prevent alert spam
}
```

### Querying Alerts

```go
// Get all alerts
alerts, err := monitor.GetAlerts(monitoring.AlertFilters{})
if err != nil {
    logger.Error("failed to get alerts", zap.Error(err))
}

// Filter alerts by severity
alerts, err := monitor.GetAlerts(monitoring.AlertFilters{
    Severities: []monitoring.SecurityEventSeverity{
        monitoring.SeverityCritical,
        monitoring.SeverityHigh,
    },
})

// Filter unacknowledged alerts
acknowledged := false
alerts, err := monitor.GetAlerts(monitoring.AlertFilters{
    Acknowledged: &acknowledged,
})

// Filter alerts by time range
startTime := time.Now().Add(-1 * time.Hour)
endTime := time.Now()
alerts, err := monitor.GetAlerts(monitoring.AlertFilters{
    StartTime: &startTime,
    EndTime:   &endTime,
})
```

### Acknowledging Alerts

```go
// Acknowledge an alert
err := monitor.AcknowledgeAlert("alt_1234567890", "admin")
if err != nil {
    logger.Error("failed to acknowledge alert", zap.Error(err))
}
```

## Metrics and Analytics

### Getting Metrics

```go
// Get current security metrics
metrics := monitor.GetMetrics()

fmt.Printf("Total Events: %d\n", metrics.TotalEvents)
fmt.Printf("Active Alerts: %d\n", metrics.ActiveAlerts)
fmt.Printf("Resolved Events: %d\n", metrics.ResolvedEvents)
fmt.Printf("Average Resolution Time: %v\n", metrics.AverageResolutionTime)

// Events by type
for eventType, count := range metrics.EventsByType {
    fmt.Printf("%s: %d\n", eventType, count)
}

// Events by severity
for severity, count := range metrics.EventsBySeverity {
    fmt.Printf("%s: %d\n", severity, count)
}

// Top IP addresses
for _, ipCount := range metrics.TopIPAddresses {
    fmt.Printf("%s: %d events\n", ipCount.IPAddress, ipCount.Count)
}

// Top endpoints
for _, endpointCount := range metrics.TopEndpoints {
    fmt.Printf("%s: %d events\n", endpointCount.Endpoint, endpointCount.Count)
}
```

### Custom Metrics Analysis

```go
// Analyze events over time
func analyzeSecurityTrends(monitor *monitoring.SecurityMonitor) {
    // Get events from last 24 hours
    startTime := time.Now().Add(-24 * time.Hour)
    events, err := monitor.GetEvents(monitoring.EventFilters{
        StartTime: &startTime,
    })
    if err != nil {
        return
    }
    
    // Analyze by hour
    hourlyStats := make(map[int]int)
    for _, event := range events {
        hour := event.Timestamp.Hour()
        hourlyStats[hour]++
    }
    
    // Find peak hours
    var peakHour int
    maxEvents := 0
    for hour, count := range hourlyStats {
        if count > maxEvents {
            maxEvents = count
            peakHour = hour
        }
    }
    
    fmt.Printf("Peak security activity hour: %d:00 (%d events)\n", peakHour, maxEvents)
}
```

## Integration Examples

### Authentication Integration

```go
// In your authentication middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... authentication logic ...
        
        if authFailed {
            // Record failed login attempt
            event := &monitoring.SecurityEvent{
                Type:     monitoring.EventTypeLoginFailure,
                Severity: monitoring.SeverityMedium,
                Source:   "auth",
                UserID:   username,
                IPAddress: getClientIP(r),
                UserAgent: r.UserAgent(),
                Endpoint: r.URL.Path,
                Method:   r.Method,
                Details: map[string]interface{}{
                    "reason": "invalid_password",
                    "attempt_count": getFailedAttempts(username),
                },
            }
            
            monitor.RecordEvent(event)
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Input Validation Integration

```go
// In your input validation middleware
func validationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... validation logic ...
        
        if sqlInjectionDetected {
            // Record SQL injection attempt
            event := &monitoring.SecurityEvent{
                Type:     monitoring.EventTypeSQLInjectionAttempt,
                Severity: monitoring.SeverityCritical,
                Source:   "validation",
                IPAddress: getClientIP(r),
                UserAgent: r.UserAgent(),
                Endpoint: r.URL.Path,
                Method:   r.Method,
                Details: map[string]interface{}{
                    "payload": sanitizedPayload,
                    "pattern": detectedPattern,
                },
            }
            
            monitor.RecordEvent(event)
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiting Integration

```go
// In your rate limiting middleware
func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ... rate limiting logic ...
        
        if rateLimitExceeded {
            // Record rate limit exceeded
            event := &monitoring.SecurityEvent{
                Type:     monitoring.EventTypeRateLimitExceeded,
                Severity: monitoring.SeverityHigh,
                Source:   "rate_limiting",
                IPAddress: getClientIP(r),
                UserAgent: r.UserAgent(),
                Endpoint: r.URL.Path,
                Method:   r.Method,
                Details: map[string]interface{}{
                    "limit": rateLimit,
                    "window": rateLimitWindow,
                    "requests": requestCount,
                },
            }
            
            monitor.RecordEvent(event)
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Webhook Integration

### External Alerting

```go
config := &monitoring.SecurityMonitorConfig{
    WebhookURL:     "https://alerts.example.com/webhook",
    WebhookTimeout: 10 * time.Second,
}

// The monitor will automatically send alerts to the webhook
monitor := monitoring.NewSecurityMonitor(config, logger)
```

### Webhook Payload Format

```json
{
  "id": "alt_1234567890",
  "event_id": "evt_1234567890",
  "type": "immediate",
  "severity": "critical",
  "title": "SQL injection attempt detected",
  "message": "Security event of type sql_injection_attempt with severity critical detected from source validation",
  "source": "validation",
  "details": {
    "payload": "1' OR '1'='1",
    "pattern": "sql_injection"
  },
  "timestamp": "2024-12-19T10:30:00Z",
  "acknowledged": false
}
```

## Performance Considerations

### Memory Usage

- Events are stored in memory with configurable limits
- Default max events: 10,000
- Events are automatically cleaned up based on retention period
- Use appropriate retention periods for your use case

### Processing Performance

- Events are processed asynchronously using channels
- Background goroutines handle event processing, alerting, and metrics
- Non-blocking event recording with channel buffering
- Configurable metrics update intervals

### Scalability

- Thread-safe operations with RWMutex
- Concurrent event recording supported
- Efficient filtering and querying
- Minimal memory allocations

## Best Practices

### 1. Event Classification

- Use appropriate event types and severity levels
- Include relevant context in event details
- Be consistent with event naming and categorization

### 2. Alert Thresholds

- Start with conservative thresholds
- Monitor alert volume and adjust as needed
- Use different thresholds for different environments

### 3. Event Retention

- Balance storage requirements with compliance needs
- Consider data privacy regulations
- Implement appropriate retention policies

### 4. Monitoring Integration

- Integrate with existing monitoring systems
- Set up appropriate alerting channels
- Monitor the security monitoring system itself

### 5. Performance Tuning

- Adjust buffer sizes based on event volume
- Monitor memory usage and adjust limits
- Use appropriate metrics update intervals

## Troubleshooting

### Common Issues

#### 1. High Memory Usage

**Problem**: Monitor using too much memory
**Solution**: Reduce MaxEvents or EventRetention

```go
config := &SecurityMonitorConfig{
    MaxEvents:      5000,  // Reduce from 10000
    EventRetention: 7 * 24 * time.Hour, // Reduce from 30 days
}
```

#### 2. Too Many Alerts

**Problem**: Receiving too many alerts
**Solution**: Increase alert thresholds

```go
config := &SecurityMonitorConfig{
    AlertThresholds: map[SecurityEventSeverity]int{
        SeverityCritical: 2,  // Increase from 1
        SeverityHigh:     10, // Increase from 5
        SeverityMedium:   20, // Increase from 10
    },
    AlertCooldown: 10 * time.Minute, // Increase from 5 minutes
}
```

#### 3. Missing Events

**Problem**: Events not being recorded
**Solution**: Check excluded sources and event types

```go
config := &SecurityMonitorConfig{
    ExcludeSources:    []string{}, // Remove exclusions
    ExcludeEventTypes: []SecurityEventType{}, // Remove exclusions
}
```

#### 4. Channel Full Errors

**Problem**: Event channel full errors
**Solution**: Increase channel buffer size or process events faster

```go
// In the monitor implementation, increase buffer sizes
eventChan: make(chan *SecurityEvent, 2000), // Increase from 1000
alertChan: make(chan *SecurityAlert, 200),  // Increase from 100
```

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
logger := zap.NewDevelopment()
monitor := monitoring.NewSecurityMonitor(config, logger)
```

## Future Enhancements

### Planned Features

1. **Database Storage**: Persistent event storage with database backends
2. **Advanced Analytics**: Machine learning-based anomaly detection
3. **Real-time Dashboards**: Web-based monitoring dashboards
4. **Integration APIs**: REST APIs for external system integration
5. **Advanced Filtering**: Complex query language for event filtering
6. **Automated Response**: Automated incident response actions
7. **Compliance Reporting**: Built-in compliance and audit reporting

### Extension Points

The monitoring system is designed for easy extension:
- Custom event types and severity levels
- Custom alert generation logic
- Custom metrics calculations
- Custom webhook integrations
- Custom event processing pipelines

## Conclusion

The Security Monitoring system provides comprehensive security event tracking, alerting, and analytics for the KYB Platform. By following the best practices outlined in this documentation, you can effectively monitor and respond to security events while maintaining system performance and scalability.
