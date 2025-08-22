# Technical Debt Monitoring System

## Overview

The Technical Debt Monitoring System provides comprehensive monitoring and metrics collection for tracking technical debt across the KYB Platform. This system helps identify, measure, and track technical debt reduction efforts through automated scanning, metrics collection, and reporting.

## Features

### Core Capabilities

- **Automated Code Scanning**: Periodic scanning of the codebase to identify technical debt indicators
- **Comprehensive Metrics Collection**: 25+ metrics covering code quality, maintainability, and technical debt
- **Prometheus Integration**: Real-time metrics export for monitoring and alerting
- **Historical Tracking**: Maintains history of metrics for trend analysis
- **Automated Alerts**: Threshold-based alerts for critical technical debt issues
- **API Endpoints**: RESTful API for accessing metrics and reports
- **Trend Analysis**: Calculates trends over time to measure improvement

### Metrics Collected

#### Code Quality Metrics
- **Total Lines of Code**: Overall codebase size
- **Deprecated Code Lines**: Lines marked as deprecated
- **Legacy Code Lines**: Lines containing legacy patterns
- **Dead Code Lines**: Unused or unreachable code
- **Code Smells**: Various code quality issues
- **Code Complexity Score**: Overall complexity measurement
- **Cyclomatic Complexity**: Average complexity per function

#### Test and Build Metrics
- **Test Coverage Percentage**: Percentage of code covered by tests
- **Test Pass Rate**: Percentage of tests passing
- **Build Success Rate**: Percentage of successful builds
- **Security Vulnerabilities**: Number of security issues

#### Maintainability Metrics
- **Code Quality Score**: Overall code quality (0-100)
- **Maintainability Index**: Maintainability measurement (0-100)
- **Code Duplication Percentage**: Percentage of duplicated code
- **Migration Progress Percentage**: Progress of legacy code migration

#### Technical Debt Metrics
- **Technical Debt Ratio**: Ratio of technical debt to total code
- **Technical Debt Cost**: Estimated cost in development hours
- **Refactoring Opportunities**: Number of identified refactoring opportunities
- **Priority Issues**: Issues categorized by severity
- **Technical Debt Trend**: Direction of technical debt (increasing/decreasing/stable)

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Technical Debt Monitor                   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Code Scanner  │  │  Metrics Store  │  │  Prometheus  │ │
│  │                 │  │                 │  │   Exporter   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   API Handler   │  │  Alert Manager  │  │  Trend Calc  │ │
│  │                 │  │                 │  │              │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Periodic Scanning**: The monitor scans the codebase every hour
2. **Metrics Collection**: Various metrics are collected during scanning
3. **Data Storage**: Metrics are stored in memory with history
4. **Prometheus Export**: Metrics are exported to Prometheus
5. **API Access**: Metrics are available via REST API
6. **Alerting**: Threshold-based alerts are generated

## API Endpoints

### Technical Debt Metrics

#### GET `/api/v3/technical-debt/metrics`
Returns current technical debt metrics.

**Response:**
```json
{
  "metrics": {
    "timestamp": "2024-01-15T10:30:00Z",
    "total_lines_of_code": 15000,
    "deprecated_code_lines": 500,
    "legacy_code_lines": 200,
    "technical_debt_ratio": 0.047,
    "test_coverage_percentage": 85.5,
    "code_quality_score": 82.3,
    "maintainability_index": 78.9,
    "technical_debt_cost": 100.0,
    "refactoring_opportunities": 25,
    "priority_issues": 5
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/metrics",
  "duration_ms": 45
}
```

#### GET `/api/v3/technical-debt/report`
Returns a comprehensive technical debt report with recommendations.

**Response:**
```json
{
  "report": {
    "timestamp": "2024-01-15T10:30:00Z",
    "total_lines_of_code": 15000,
    "deprecated_code_lines": 500,
    "legacy_code_lines": 200,
    "technical_debt_ratio": 0.047,
    "test_coverage_percentage": 85.5,
    "code_quality_score": 82.3,
    "maintainability_index": 78.9,
    "technical_debt_cost": 100.0,
    "refactoring_opportunities": 25,
    "priority_issues": 5,
    "technical_debt_trend": "decreasing",
    "next_refactoring_target": "internal/classification/service.go",
    "recommendations": [
      "Test coverage below 80%. Increase test coverage for better code quality."
    ]
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/report",
  "duration_ms": 67
}
```

#### GET `/api/v3/technical-debt/history?limit=50`
Returns historical technical debt metrics.

**Query Parameters:**
- `limit` (optional): Number of historical records to return (default: 50)

**Response:**
```json
{
  "history": [
    {
      "timestamp": "2024-01-15T09:30:00Z",
      "total_lines_of_code": 15000,
      "technical_debt_ratio": 0.048,
      "test_coverage_percentage": 85.2
    }
  ],
  "count": 1,
  "limit": 50,
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/history",
  "duration_ms": 23
}
```

#### GET `/api/v3/technical-debt/trends?days=30`
Returns technical debt trends over time.

**Query Parameters:**
- `days` (optional): Number of days to analyze (default: 30)

**Response:**
```json
{
  "trends": {
    "technical_debt_ratio": {
      "start": 0.050,
      "end": 0.047,
      "change": -0.003,
      "trend": "decreasing"
    },
    "test_coverage": {
      "start": 84.5,
      "end": 85.5,
      "change": 1.0,
      "trend": "increasing"
    }
  },
  "days": 30,
  "count": 30,
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/trends",
  "duration_ms": 89
}
```

#### GET `/api/v3/technical-debt/alerts`
Returns technical debt alerts based on thresholds.

**Response:**
```json
{
  "alerts": [
    {
      "severity": "medium",
      "metric": "test_coverage",
      "value": 75.5,
      "threshold": 80,
      "message": "Test coverage is below 80%. Increase test coverage for better code quality.",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1,
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/alerts",
  "duration_ms": 34
}
```

#### POST `/api/v3/technical-debt/scan`
Triggers an immediate technical debt scan.

**Response:**
```json
{
  "message": "Technical debt scan triggered successfully",
  "timestamp": "2024-01-15T10:30:00Z",
  "endpoint": "/api/v3/technical-debt/scan",
  "duration_ms": 12
}
```

## Prometheus Metrics

The system exports the following Prometheus metrics:

### Gauge Metrics
- `kyb_technical_debt_ratio{module,environment}` - Technical debt ratio
- `kyb_deprecated_code_lines{module,environment}` - Lines of deprecated code
- `kyb_legacy_code_lines{module,environment}` - Lines of legacy code
- `kyb_test_coverage_percentage{module,environment}` - Test coverage percentage
- `kyb_code_complexity_score{module,environment}` - Code complexity score
- `kyb_build_success_rate{environment}` - Build success rate
- `kyb_test_pass_rate{environment}` - Test pass rate
- `kyb_security_vulnerabilities{severity,environment}` - Security vulnerabilities
- `kyb_code_duplication_percentage{module,environment}` - Code duplication percentage
- `kyb_migration_progress_percentage{module,environment}` - Migration progress
- `kyb_module_health_score{module,environment}` - Module health score
- `kyb_code_quality_score{module,environment}` - Code quality score
- `kyb_maintainability_index{module,environment}` - Maintainability index
- `kyb_cyclomatic_complexity{module,environment}` - Cyclomatic complexity
- `kyb_technical_debt_cost{module,environment}` - Technical debt cost
- `kyb_refactoring_opportunities{priority,environment}` - Refactoring opportunities
- `kyb_dead_code_lines{module,environment}` - Dead code lines
- `kyb_code_smells{type,environment}` - Code smells
- `kyb_priority_issues{severity,environment}` - Priority issues

### Counter Metrics
- `kyb_deprecated_api_calls_total{api_endpoint,environment}` - Deprecated API calls

## Configuration

### Environment Variables
- `ENVIRONMENT`: Environment name (default: "development")
- `TECHNICAL_DEBT_SCAN_INTERVAL`: Scan interval in hours (default: 1)
- `TECHNICAL_DEBT_MAX_HISTORY`: Maximum number of historical records (default: 100)

### Thresholds
The system uses the following default thresholds for alerts:

- **Technical Debt Ratio**: > 30% (high severity)
- **Test Coverage**: < 80% (medium severity)
- **Code Quality Score**: < 70% (medium severity)
- **Maintainability Index**: < 60% (high severity)
- **Refactoring Opportunities**: > 20 (medium severity)
- **Priority Issues**: > 10 (high severity)

## Usage Examples

### Starting the Monitor

```go
package main

import (
    "context"
    "log"
    
    "github.com/pcraw4d/business-verification/internal/observability"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Create technical debt monitor
    monitor := observability.NewTechnicalDebtMonitor(logger, ".")
    
    // Start monitoring
    ctx := context.Background()
    monitor.Start(ctx)
}
```

### Using the API Handler

```go
package main

import (
    "net/http"
    
    "github.com/pcraw4d/business-verification/internal/api/handlers"
    "github.com/pcraw4d/business-verification/internal/observability"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Create monitor and handler
    monitor := observability.NewTechnicalDebtMonitor(logger, ".")
    handler := handlers.NewTechnicalDebtMonitorHandler(monitor, logger)
    
    // Set up routes
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v3/technical-debt/metrics", handler.GetTechnicalDebtMetrics)
    mux.HandleFunc("/api/v3/technical-debt/report", handler.GetTechnicalDebtReport)
    mux.HandleFunc("/api/v3/technical-debt/history", handler.GetTechnicalDebtHistory)
    mux.HandleFunc("/api/v3/technical-debt/trends", handler.GetTechnicalDebtTrends)
    mux.HandleFunc("/api/v3/technical-debt/alerts", handler.GetTechnicalDebtAlerts)
    mux.HandleFunc("/api/v3/technical-debt/scan", handler.TriggerTechnicalDebtScan)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

### Recording Deprecated API Calls

```go
// In your API handlers
func (h *SomeHandler) DeprecatedEndpoint(w http.ResponseWriter, r *http.Request) {
    // Record the deprecated API call
    h.technicalDebtMonitor.RecordDeprecatedAPICall("/api/v1/deprecated")
    
    // Handle the request
    // ...
}
```

## Monitoring and Alerting

### Grafana Dashboard

Create a Grafana dashboard with the following panels:

1. **Technical Debt Overview**
   - Technical debt ratio over time
   - Test coverage trend
   - Code quality score

2. **Code Quality Metrics**
   - Deprecated code lines
   - Legacy code lines
   - Code smells by type

3. **Build and Test Metrics**
   - Build success rate
   - Test pass rate
   - Security vulnerabilities

4. **Alerts Panel**
   - Current alerts by severity
   - Alert history

### Alert Rules

Configure Prometheus alert rules:

```yaml
groups:
  - name: technical-debt
    rules:
      - alert: HighTechnicalDebtRatio
        expr: kyb_technical_debt_ratio > 0.3
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "High technical debt ratio detected"
          description: "Technical debt ratio is {{ $value }} (threshold: 0.3)"
      
      - alert: LowTestCoverage
        expr: kyb_test_coverage_percentage < 80
        for: 5m
        labels:
          severity: medium
        annotations:
          summary: "Low test coverage detected"
          description: "Test coverage is {{ $value }}% (threshold: 80%)"
      
      - alert: HighPriorityIssues
        expr: kyb_priority_issues{severity="high"} > 10
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "High number of priority issues"
          description: "{{ $value }} high priority issues detected"
```

## Best Practices

### Implementation Guidelines

1. **Regular Monitoring**: Run the monitor continuously to track trends
2. **Threshold Tuning**: Adjust thresholds based on your team's standards
3. **Actionable Alerts**: Ensure alerts lead to actionable items
4. **Historical Analysis**: Use trend data to measure improvement
5. **Integration**: Integrate with CI/CD pipelines for automated checks

### Maintenance

1. **Regular Reviews**: Review metrics weekly with the development team
2. **Threshold Updates**: Update thresholds as code quality improves
3. **Pattern Updates**: Add new patterns for deprecated/legacy code detection
4. **Performance Monitoring**: Monitor the monitor's performance impact

### Integration with Development Workflow

1. **Pre-commit Hooks**: Check technical debt metrics before commits
2. **Pull Request Gates**: Require technical debt improvements in PRs
3. **Sprint Planning**: Include technical debt reduction in sprint planning
4. **Retrospectives**: Review technical debt trends in team retrospectives

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Reduce `maxHistory` or increase scan interval
2. **Slow Scans**: Exclude more directories or reduce scan frequency
3. **Missing Metrics**: Check file permissions and exclude patterns
4. **API Timeouts**: Increase timeout values for large codebases

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
logger, _ := zap.NewDevelopment()
monitor := observability.NewTechnicalDebtMonitor(logger, ".")
```

## Future Enhancements

### Planned Features

1. **Advanced Code Analysis**: Integration with static analysis tools
2. **Machine Learning**: ML-based code quality prediction
3. **Team Metrics**: Per-developer technical debt metrics
4. **Integration APIs**: Webhook support for external tools
5. **Custom Rules**: User-defined technical debt rules
6. **Visualization**: Built-in charts and graphs
7. **Export Options**: CSV, JSON, PDF report exports

### Roadmap

- **Q1 2024**: Basic monitoring and alerting
- **Q2 2024**: Advanced code analysis integration
- **Q3 2024**: Team metrics and custom rules
- **Q4 2024**: ML-based predictions and visualizations

## Conclusion

The Technical Debt Monitoring System provides comprehensive visibility into code quality and technical debt across the KYB Platform. By tracking metrics over time and providing actionable alerts, it helps teams maintain high code quality and reduce technical debt systematically.

For questions or support, please refer to the development team or create an issue in the project repository.
