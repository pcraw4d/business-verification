# Health Check Endpoints Documentation

## Overview

The Enhanced Business Intelligence System provides comprehensive health check endpoints for monitoring system status, readiness, and liveness. These endpoints are designed to work with container orchestration platforms like Kubernetes, load balancers, and monitoring systems.

## Features

### Core Features
- **Comprehensive Health Checks**: System, database, cache, external APIs, ML models, and observability
- **Readiness Probes**: Critical dependency checks for container orchestration
- **Liveness Probes**: Basic system health for process monitoring
- **Detailed Health Information**: Memory, goroutines, disk, and network metrics
- **Module-Specific Health**: Individual module health status
- **Caching**: Performance optimization with configurable TTL
- **Metrics Collection**: Runtime statistics and performance data

### Security Features
- **No Authentication Required**: Health endpoints are publicly accessible
- **Structured Logging**: Comprehensive audit trail
- **Error Handling**: Graceful degradation and error reporting
- **Rate Limiting**: Built-in protection against abuse

### Customization Features
- **Configurable Cache TTL**: Adjustable caching duration
- **Custom Health Checks**: Extensible check framework
- **Environment-Specific**: Version and environment information
- **Response Formatting**: Consistent JSON responses

## Endpoints

### 1. Main Health Check
**Endpoint**: `GET /health`  
**Description**: Comprehensive system health status  
**Authentication**: None required  
**Response**: Full health status with all checks and metrics

**Response Format**:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "uptime": "2h30m15s",
  "ready": true,
  "live": true,
  "checks": {
    "system": {
      "status": "healthy",
      "response_time": "1ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "uptime": "2h30m15s",
        "version": "1.0.0",
        "environment": "production"
      }
    },
    "database": {
      "status": "healthy",
      "response_time": "15ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "connection_pool_size": 10,
        "active_connections": 3,
        "max_connections": 100
      }
    }
  },
  "metrics": {
    "total_checks": 6,
    "healthy_checks": 6,
    "unhealthy_checks": 0,
    "degraded_checks": 0,
    "average_response_time": "12ms",
    "memory_usage": {
      "alloc_bytes": 1048576,
      "sys_bytes": 2097152,
      "heap_alloc_bytes": 1048576,
      "heap_sys_bytes": 2097152,
      "num_gc": 5,
      "memory_usage_percent": 50.0
    },
    "go_runtime": {
      "version": "go1.22.0",
      "num_cpu": 8,
      "num_goroutine": 25,
      "num_cgo_call": 0
    }
  }
}
```

### 2. Readiness Probe
**Endpoint**: `GET /ready`  
**Description**: Critical dependency health for container orchestration  
**Authentication**: None required  
**Response**: Readiness status with critical checks only

**Response Format**:
```json
{
  "ready": true,
  "timestamp": "2024-12-19T10:30:00Z",
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time": "15ms",
      "last_check": "2024-12-19T10:30:00Z"
    },
    "cache": {
      "status": "healthy",
      "response_time": "5ms",
      "last_check": "2024-12-19T10:30:00Z"
    }
  }
}
```

### 3. Liveness Probe
**Endpoint**: `GET /live`  
**Description**: Basic system health for process monitoring  
**Authentication**: None required  
**Response**: Liveness status with basic system checks

**Response Format**:
```json
{
  "live": true,
  "timestamp": "2024-12-19T10:30:00Z",
  "status": "healthy",
  "uptime": "2h30m15s"
}
```

### 4. Detailed Health Check
**Endpoint**: `GET /health/detailed`  
**Description**: Comprehensive health status with additional metrics  
**Authentication**: None required  
**Response**: Full health status with detailed system metrics

**Response Format**:
```json
{
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "uptime": "2h30m15s",
  "ready": true,
  "live": true,
  "checks": {
    "system": { /* ... */ },
    "database": { /* ... */ },
    "cache": { /* ... */ },
    "external_apis": { /* ... */ },
    "ml_models": { /* ... */ },
    "observability": { /* ... */ },
    "memory": {
      "status": "healthy",
      "response_time": "1ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "alloc_bytes": 1048576,
        "sys_bytes": 2097152,
        "heap_alloc_bytes": 1048576,
        "heap_sys_bytes": 2097152,
        "num_gc": 5,
        "memory_usage_percent": 50.0
      }
    },
    "goroutines": {
      "status": "healthy",
      "response_time": "1ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "num_goroutines": 25,
        "num_cpu": 8
      }
    },
    "disk": {
      "status": "healthy",
      "response_time": "5ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "disk_usage_percent": 45.2,
        "available_space_gb": 125.8,
        "total_space_gb": 256.0
      }
    },
    "network": {
      "status": "healthy",
      "response_time": "10ms",
      "last_check": "2024-12-19T10:30:00Z",
      "details": {
        "network_latency_ms": 15.3,
        "packet_loss_percent": 0.01,
        "bandwidth_mbps": 1000.0
      }
    }
  },
  "metrics": { /* ... */ }
}
```

### 5. Module Health Check
**Endpoint**: `GET /health/module?module={module_name}`  
**Description**: Health status of a specific module  
**Authentication**: None required  
**Response**: Module-specific health information

**Response Format**:
```json
{
  "status": "healthy",
  "response_time": "15ms",
  "last_check": "2024-12-19T10:30:00Z",
  "details": {
    "module_specific_info": "value"
  }
}
```

## Health Check Types

### System Health Check
- **Purpose**: Basic system responsiveness
- **Checks**: Uptime, version, environment
- **Response Time**: < 1ms
- **Status**: Always healthy (basic check)

### Database Health Check
- **Purpose**: Database connectivity and performance
- **Checks**: Connection pool, active connections, query performance
- **Response Time**: < 50ms
- **Status**: healthy/degraded/unhealthy

### Cache Health Check
- **Purpose**: Cache connectivity and performance
- **Checks**: Cache size, hit rate, miss rate
- **Response Time**: < 20ms
- **Status**: healthy/degraded/unhealthy

### External APIs Health Check
- **Purpose**: External service connectivity
- **Checks**: API availability, response times
- **Response Time**: < 100ms
- **Status**: healthy/degraded/unhealthy

### ML Models Health Check
- **Purpose**: Machine learning model availability
- **Checks**: Model loading, version, performance
- **Response Time**: < 50ms
- **Status**: healthy/degraded/unhealthy

### Observability Health Check
- **Purpose**: Monitoring system health
- **Checks**: Logging, metrics, tracing
- **Response Time**: < 10ms
- **Status**: healthy/degraded/unhealthy

### Memory Health Check
- **Purpose**: Memory usage monitoring
- **Checks**: Allocated memory, heap usage, garbage collection
- **Response Time**: < 1ms
- **Status**: healthy/degraded/unhealthy

### Goroutines Health Check
- **Purpose**: Goroutine count monitoring
- **Checks**: Number of goroutines, CPU usage
- **Response Time**: < 1ms
- **Status**: healthy/degraded/unhealthy

### Disk Health Check
- **Purpose**: Disk usage monitoring
- **Checks**: Available space, usage percentage
- **Response Time**: < 10ms
- **Status**: healthy/degraded/unhealthy

### Network Health Check
- **Purpose**: Network connectivity monitoring
- **Checks**: Latency, packet loss, bandwidth
- **Response Time**: < 20ms
- **Status**: healthy/degraded/unhealthy

## Configuration

### Health Handler Configuration
```go
type HealthHandler struct {
    logger        *zap.Logger
    healthChecker *health.RailwayHealthChecker
    startTime     time.Time
    version       string
    environment   string
    cacheTTL      time.Duration
}
```

### Configuration Options
- **Cache TTL**: Default 30 seconds, configurable
- **Version**: Application version string
- **Environment**: Environment name (dev, staging, production)
- **Logger**: Structured logging instance

## Usage Examples

### Basic Health Check
```bash
# Check overall system health
curl -X GET http://localhost:8080/health

# Check readiness for container orchestration
curl -X GET http://localhost:8080/ready

# Check liveness for process monitoring
curl -X GET http://localhost:8080/live
```

### Detailed Health Information
```bash
# Get detailed health status with additional metrics
curl -X GET http://localhost:8080/health/detailed

# Check specific module health
curl -X GET "http://localhost:8080/health/module?module=database"
```

### Kubernetes Integration
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform
spec:
  template:
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:latest
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Docker Compose Integration
```yaml
version: '3.8'
services:
  kyb-platform:
    image: kyb-platform:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Integration with Other Middleware

### Security Headers
Health check endpoints are exempt from security headers to ensure compatibility with load balancers and monitoring systems.

### Request Logging
All health check requests are logged with structured logging for monitoring and debugging.

### Rate Limiting
Health check endpoints may have different rate limiting rules to ensure availability.

### CORS Policy
Health check endpoints support CORS for cross-origin monitoring.

## Error Handling

### HTTP Status Codes
- **200 OK**: System is healthy
- **503 Service Unavailable**: System is unhealthy or not ready
- **500 Internal Server Error**: Unexpected error

### Error Response Format
```json
{
  "error": "Health check failed",
  "message": "Database connection timeout",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Performance Considerations

### Caching
- Health check results are cached for 30 seconds by default
- Cache TTL is configurable
- Cache is thread-safe with read-write mutex

### Response Times
- Basic health checks: < 1ms
- Database checks: < 50ms
- External API checks: < 100ms
- Detailed checks: < 200ms

### Resource Usage
- Memory overhead: < 1MB
- CPU overhead: < 1%
- Network overhead: < 1KB per request

## Monitoring and Alerting

### Key Metrics
- **Health Check Response Time**: Average response time across all checks
- **Health Check Success Rate**: Percentage of successful health checks
- **System Uptime**: Total system uptime
- **Memory Usage**: Current memory allocation and usage
- **Goroutine Count**: Number of active goroutines

### Alerting Rules
```yaml
# Prometheus alerting rules
groups:
  - name: health_checks
    rules:
      - alert: HealthCheckFailed
        expr: health_check_status{status="unhealthy"} > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Health check failed"
          description: "System health check is failing"
      
      - alert: HealthCheckDegraded
        expr: health_check_status{status="degraded"} > 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Health check degraded"
          description: "System health check is degraded"
```

## Testing

### Unit Tests
```bash
# Run health handler tests
go test ./internal/api/handlers -run TestHealthHandler

# Run with coverage
go test ./internal/api/handlers -run TestHealthHandler -cover
```

### Integration Tests
```bash
# Test health endpoints
curl -X GET http://localhost:8080/health
curl -X GET http://localhost:8080/ready
curl -X GET http://localhost:8080/live
curl -X GET http://localhost:8080/health/detailed
```

### Load Testing
```bash
# Test health endpoint performance
ab -n 1000 -c 10 http://localhost:8080/health
```

## Best Practices

### 1. Health Check Design
- Keep health checks lightweight and fast
- Avoid external dependencies in basic health checks
- Use appropriate timeouts for external checks
- Implement graceful degradation

### 2. Monitoring Integration
- Use health endpoints for load balancer health checks
- Integrate with monitoring systems (Prometheus, Grafana)
- Set up alerting for health check failures
- Monitor health check response times

### 3. Container Orchestration
- Use readiness probes for traffic routing
- Use liveness probes for process monitoring
- Set appropriate probe intervals and timeouts
- Handle startup delays properly

### 4. Security Considerations
- Health endpoints should be publicly accessible
- Avoid sensitive information in health responses
- Use structured logging for audit trails
- Implement rate limiting if needed

## Troubleshooting

### Common Issues

#### 1. Health Check Timeout
**Symptoms**: Health check requests timeout
**Causes**: External dependencies slow to respond
**Solutions**:
- Increase timeout values
- Implement circuit breakers
- Use caching for slow checks

#### 2. High Response Times
**Symptoms**: Health check response times > 100ms
**Causes**: Database or external API issues
**Solutions**:
- Optimize database queries
- Implement connection pooling
- Use caching strategies

#### 3. Memory Issues
**Symptoms**: High memory usage in health checks
**Causes**: Memory leaks or inefficient checks
**Solutions**:
- Review health check implementations
- Implement memory limits
- Use garbage collection tuning

#### 4. Cache Issues
**Symptoms**: Stale health check data
**Causes**: Cache TTL too long or cache corruption
**Solutions**:
- Adjust cache TTL
- Implement cache invalidation
- Monitor cache hit rates

### Debugging Commands
```bash
# Check health endpoint directly
curl -v http://localhost:8080/health

# Check specific module health
curl -v "http://localhost:8080/health/module?module=database"

# Check logs for health check errors
tail -f logs/application.log | grep health

# Monitor health check metrics
curl http://localhost:8080/metrics | grep health
```

## Migration Guide

### From Basic Health Checks
If migrating from basic health checks to comprehensive health endpoints:

1. **Update Endpoints**: Replace basic `/health` with new endpoints
2. **Update Monitoring**: Update monitoring configuration
3. **Update Load Balancers**: Update health check URLs
4. **Update Documentation**: Update API documentation

### Configuration Changes
```go
// Old configuration
mux.HandleFunc("GET /health", basicHealthHandler)

// New configuration
healthHandler := NewHealthHandler(logger, healthChecker, version, environment)
mux.HandleFunc("GET /health", healthHandler.HandleHealth)
mux.HandleFunc("GET /ready", healthHandler.HandleReadiness)
mux.HandleFunc("GET /live", healthHandler.HandleLiveness)
mux.HandleFunc("GET /health/detailed", healthHandler.HandleDetailedHealth)
mux.HandleFunc("GET /health/module", healthHandler.HandleModuleHealth)
```

## API Reference

### Types

#### HealthStatus
```go
type HealthStatus struct {
    Status      string                 `json:"status"`
    Timestamp   time.Time              `json:"timestamp"`
    Version     string                 `json:"version"`
    Environment string                 `json:"environment"`
    Uptime      time.Duration          `json:"uptime"`
    Checks      map[string]HealthCheck `json:"checks"`
    Metrics     HealthMetrics          `json:"metrics"`
    Ready       bool                   `json:"ready"`
    Live        bool                   `json:"live"`
}
```

#### HealthCheck
```go
type HealthCheck struct {
    Status       string                 `json:"status"`
    ResponseTime time.Duration          `json:"response_time,omitempty"`
    LastCheck    time.Time              `json:"last_check"`
    Error        string                 `json:"error,omitempty"`
    Details      map[string]interface{} `json:"details,omitempty"`
}
```

#### HealthMetrics
```go
type HealthMetrics struct {
    TotalChecks     int           `json:"total_checks"`
    HealthyChecks   int           `json:"healthy_checks"`
    UnhealthyChecks int           `json:"unhealthy_checks"`
    DegradedChecks  int           `json:"degraded_checks"`
    AverageResponse time.Duration `json:"average_response_time"`
    MemoryUsage     MemoryInfo    `json:"memory_usage"`
    GoRuntime       GoRuntimeInfo `json:"go_runtime"`
}
```

### Functions

#### NewHealthHandler
```go
func NewHealthHandler(logger *zap.Logger, healthChecker *health.RailwayHealthChecker, version, environment string) *HealthHandler
```
Creates a new health handler with the specified configuration.

#### HandleHealth
```go
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request)
```
Handles the main health check endpoint.

#### HandleReadiness
```go
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request)
```
Handles the readiness probe endpoint.

#### HandleLiveness
```go
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request)
```
Handles the liveness probe endpoint.

#### HandleDetailedHealth
```go
func (h *HealthHandler) HandleDetailedHealth(w http.ResponseWriter, r *http.Request)
```
Handles the detailed health check endpoint.

#### HandleModuleHealth
```go
func (h *HealthHandler) HandleModuleHealth(w http.ResponseWriter, r *http.Request)
```
Handles module-specific health checks.

## Conclusion

The health check endpoints provide comprehensive monitoring capabilities for the Enhanced Business Intelligence System. They support container orchestration, load balancing, and monitoring integration while maintaining high performance and reliability.

Key benefits:
- **Comprehensive Monitoring**: Full system health visibility
- **Container Ready**: Kubernetes and Docker integration
- **High Performance**: Caching and optimized checks
- **Extensible**: Custom health check support
- **Production Ready**: Structured logging and error handling

For additional support or questions, refer to the system documentation or contact the development team.
