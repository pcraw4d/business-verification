#!/bin/bash

# KYB Platform Monitoring Setup Script
# This script sets up comprehensive performance monitoring, alerting, and dashboards

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/configs"
INTERNAL_DIR="$PROJECT_ROOT/internal"
LOGS_DIR="$PROJECT_ROOT/logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check for required commands
    for cmd in go docker docker-compose curl jq; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again."
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

# Create necessary directories
create_directories() {
    log_info "Creating necessary directories..."
    
    local dirs=(
        "$LOGS_DIR"
        "$LOGS_DIR/monitoring"
        "$LOGS_DIR/alerts"
        "$PROJECT_ROOT/monitoring"
        "$PROJECT_ROOT/monitoring/grafana"
        "$PROJECT_ROOT/monitoring/prometheus"
        "$PROJECT_ROOT/monitoring/alertmanager"
    )
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_success "Created directory: $dir"
        fi
    done
}

# Setup Prometheus configuration
setup_prometheus() {
    log_info "Setting up Prometheus configuration..."
    
    cat > "$PROJECT_ROOT/monitoring/prometheus/prometheus.yml" << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "kyb-alert-rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    scrape_timeout: 10s

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

    # Create alert rules
    cat > "$PROJECT_ROOT/monitoring/prometheus/kyb-alert-rules.yml" << 'EOF'
groups:
  - name: kyb-platform
    rules:
      - alert: HighMemoryUsage
        expr: kyb_memory_usage_bytes / 1024 / 1024 > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"
          description: "Memory usage is {{ $value }}MB, which is above the 80MB threshold"

      - alert: CriticalMemoryUsage
        expr: kyb_memory_usage_bytes / 1024 / 1024 > 95
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Critical memory usage detected"
          description: "Memory usage is {{ $value }}MB, which is above the 95MB threshold"

      - alert: HighCPUUsage
        expr: kyb_cpu_usage_percent > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is {{ $value }}%, which is above the 80% threshold"

      - alert: CriticalCPUUsage
        expr: kyb_cpu_usage_percent > 95
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Critical CPU usage detected"
          description: "CPU usage is {{ $value }}%, which is above the 95% threshold"

      - alert: HighDatabaseConnections
        expr: kyb_database_connections_active > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High database connection usage"
          description: "Database connections are {{ $value }}, which is above the 80 threshold"

      - alert: LowClassificationAccuracy
        expr: kyb_classification_accuracy_percent < 90
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low classification accuracy"
          description: "Classification accuracy is {{ $value }}%, which is below the 90% threshold"

      - alert: HighErrorRate
        expr: rate(kyb_errors_total[5m]) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/second, which is above the 0.05 threshold"
EOF

    log_success "Prometheus configuration created"
}

# Setup AlertManager configuration
setup_alertmanager() {
    log_info "Setting up AlertManager configuration..."
    
    cat > "$PROJECT_ROOT/monitoring/alertmanager/alertmanager.yml" << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@kyb-platform.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/alerts'
        send_resolved: true

  - name: 'critical-alerts'
    email_configs:
      - to: 'admin@kyb-platform.com'
        subject: '[CRITICAL] KYB Platform Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          Severity: {{ .Labels.severity }}
          Time: {{ .StartsAt }}
          {{ end }}
    webhook_configs:
      - url: 'http://localhost:5001/alerts'
        send_resolved: true

  - name: 'warning-alerts'
    email_configs:
      - to: 'admin@kyb-platform.com'
        subject: '[WARNING] KYB Platform Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          Severity: {{ .Labels.severity }}
          Time: {{ .StartsAt }}
          {{ end }}

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

    log_success "AlertManager configuration created"
}

# Setup Grafana configuration
setup_grafana() {
    log_info "Setting up Grafana configuration..."
    
    # Create Grafana datasource configuration
    cat > "$PROJECT_ROOT/monitoring/grafana/datasources.yml" << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

    # Create Grafana dashboard provisioning
    cat > "$PROJECT_ROOT/monitoring/grafana/dashboards.yml" << 'EOF'
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF

    log_success "Grafana configuration created"
}

# Create Docker Compose file for monitoring stack
create_docker_compose() {
    log_info "Creating Docker Compose configuration for monitoring stack..."
    
    cat > "$PROJECT_ROOT/monitoring/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: kyb-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/kyb-alert-rules.yml:/etc/prometheus/kyb-alert-rules.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: kyb-alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: kyb-grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml
      - ./grafana/dashboards.yml:/etc/grafana/provisioning/dashboards/dashboards.yml
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    restart: unless-stopped

volumes:
  prometheus_data:
  alertmanager_data:
  grafana_data:
EOF

    log_success "Docker Compose configuration created"
}

# Build the monitoring Go application
build_monitoring_app() {
    log_info "Building monitoring application..."
    
    cd "$PROJECT_ROOT"
    
    # Build the main application with monitoring enabled
    go build -o bin/kyb-platform-monitoring \
        -ldflags "-X main.monitoringEnabled=true" \
        ./cmd/server
    
    if [ $? -eq 0 ]; then
        log_success "Monitoring application built successfully"
    else
        log_error "Failed to build monitoring application"
        exit 1
    fi
}

# Create monitoring startup script
create_startup_script() {
    log_info "Creating monitoring startup script..."
    
    cat > "$PROJECT_ROOT/scripts/start_monitoring.sh" << 'EOF'
#!/bin/bash

# KYB Platform Monitoring Startup Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MONITORING_DIR="$PROJECT_ROOT/monitoring"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Start monitoring stack
start_monitoring_stack() {
    log_info "Starting monitoring stack..."
    
    cd "$MONITORING_DIR"
    
    # Start Docker Compose services
    docker-compose up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # Check service health
    check_service_health "Prometheus" "http://localhost:9090/-/healthy"
    check_service_health "AlertManager" "http://localhost:9093/-/healthy"
    check_service_health "Grafana" "http://localhost:3000/api/health"
    
    log_success "Monitoring stack started successfully"
}

# Check service health
check_service_health() {
    local service_name="$1"
    local health_url="$2"
    
    log_info "Checking $service_name health..."
    
    for i in {1..30}; do
        if curl -s "$health_url" > /dev/null 2>&1; then
            log_success "$service_name is healthy"
            return 0
        fi
        sleep 2
    done
    
    log_error "$service_name is not responding"
    return 1
}

# Start the main application
start_main_app() {
    log_info "Starting main KYB Platform application..."
    
    cd "$PROJECT_ROOT"
    
    # Start the application with monitoring enabled
    ./bin/kyb-platform-monitoring &
    local app_pid=$!
    
    echo "$app_pid" > "$PROJECT_ROOT/kyb-platform.pid"
    
    log_success "Main application started with PID: $app_pid"
}

# Main execution
main() {
    log_info "Starting KYB Platform monitoring..."
    
    start_monitoring_stack
    start_main_app
    
    log_success "KYB Platform monitoring is now running!"
    log_info "Access points:"
    log_info "  - Grafana: http://localhost:3000 (admin/admin123)"
    log_info "  - Prometheus: http://localhost:9090"
    log_info "  - AlertManager: http://localhost:9093"
    log_info "  - KYB Platform: http://localhost:8080"
}

main "$@"
EOF

    chmod +x "$PROJECT_ROOT/scripts/start_monitoring.sh"
    log_success "Monitoring startup script created"
}

# Create monitoring test script
create_test_script() {
    log_info "Creating monitoring test script..."
    
    cat > "$PROJECT_ROOT/scripts/test_monitoring.sh" << 'EOF'
#!/bin/bash

# KYB Platform Monitoring Test Script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test monitoring endpoints
test_monitoring_endpoints() {
    log_info "Testing monitoring endpoints..."
    
    local endpoints=(
        "http://localhost:9090/-/healthy:Prometheus"
        "http://localhost:9093/-/healthy:AlertManager"
        "http://localhost:3000/api/health:Grafana"
        "http://localhost:8080/health:KYB Platform"
        "http://localhost:8080/metrics:KYB Platform Metrics"
    )
    
    for endpoint in "${endpoints[@]}"; do
        local url="${endpoint%%:*}"
        local name="${endpoint##*:}"
        
        log_info "Testing $name at $url"
        
        if curl -s "$url" > /dev/null 2>&1; then
            log_success "$name is responding"
        else
            log_error "$name is not responding"
            return 1
        fi
    done
    
    return 0
}

# Test alerting system
test_alerting_system() {
    log_info "Testing alerting system..."
    
    cd "$PROJECT_ROOT"
    
    # Run the alerting system tests
    go test -v ./internal/observability -run TestAlertingSystem
    
    if [ $? -eq 0 ]; then
        log_success "Alerting system tests passed"
    else
        log_error "Alerting system tests failed"
        return 1
    fi
}

# Test performance monitoring
test_performance_monitoring() {
    log_info "Testing performance monitoring..."
    
    cd "$PROJECT_ROOT"
    
    # Run the performance monitoring tests
    go test -v ./internal/observability -run TestPerformanceMonitor
    
    if [ $? -eq 0 ]; then
        log_success "Performance monitoring tests passed"
    else
        log_error "Performance monitoring tests failed"
        return 1
    fi
}

# Generate test load
generate_test_load() {
    log_info "Generating test load..."
    
    # Generate some test requests to create metrics
    for i in {1..100}; do
        curl -s "http://localhost:8080/api/health" > /dev/null 2>&1 &
    done
    
    wait
    
    log_success "Test load generated"
}

# Main execution
main() {
    log_info "Starting monitoring tests..."
    
    test_monitoring_endpoints
    test_alerting_system
    test_performance_monitoring
    generate_test_load
    
    log_success "All monitoring tests passed!"
}

main "$@"
EOF

    chmod +x "$PROJECT_ROOT/scripts/test_monitoring.sh"
    log_success "Monitoring test script created"
}

# Create monitoring documentation
create_documentation() {
    log_info "Creating monitoring documentation..."
    
    cat > "$PROJECT_ROOT/docs/monitoring.md" << 'EOF'
# KYB Platform Monitoring Documentation

## Overview

The KYB Platform includes comprehensive monitoring capabilities for performance tracking, alerting, and visualization.

## Components

### 1. Performance Monitoring
- **Memory Usage**: Tracks application memory consumption
- **CPU Usage**: Monitors CPU utilization
- **Database Connections**: Tracks database connection pool usage
- **Goroutine Count**: Monitors goroutine usage
- **GC Duration**: Tracks garbage collection performance

### 2. Business Metrics
- **Classification Accuracy**: Tracks classification model accuracy
- **Risk Detection Latency**: Monitors risk detection response times
- **API Response Time**: Tracks API endpoint performance
- **Error Rate**: Monitors application error rates
- **Request Rate**: Tracks request throughput

### 3. Alerting System
- **Threshold-based Alerts**: Configurable thresholds for all metrics
- **Multiple Severity Levels**: Info, Warning, Critical
- **Notification Channels**: Email, Slack, Webhook
- **Alert Suppression**: Maintenance windows and deployment suppression
- **Escalation Policies**: Automatic escalation for critical alerts

### 4. Dashboards
- **Performance Overview**: System and application metrics
- **Database Performance**: Database-specific metrics
- **Business Metrics**: Business-specific performance indicators
- **Alerts Dashboard**: Current alerts and notification status

## Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.22+
- curl and jq

### Installation
1. Run the setup script:
   ```bash
   ./scripts/setup_monitoring.sh
   ```

2. Start the monitoring stack:
   ```bash
   ./scripts/start_monitoring.sh
   ```

3. Test the monitoring system:
   ```bash
   ./scripts/test_monitoring.sh
   ```

## Access Points

- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093
- **KYB Platform**: http://localhost:8080
- **Metrics Endpoint**: http://localhost:8080/metrics

## Configuration

### Performance Monitoring
Configuration is located in `configs/performance_alerting.yml`:
- Collection intervals
- Alert thresholds
- Notification channels
- Suppression rules

### Dashboards
Dashboard configuration is in `configs/monitoring_dashboards.yml`:
- Panel definitions
- Query configurations
- Visualization settings

## Alerting

### Alert Types
- **System Alerts**: Memory, CPU, database connections
- **Application Alerts**: Error rates, response times
- **Business Alerts**: Classification accuracy, risk detection latency

### Notification Channels
- **Email**: SMTP-based email notifications
- **Slack**: Webhook-based Slack notifications
- **Webhook**: Custom webhook notifications

### Alert Suppression
- **Maintenance Windows**: Suppress alerts during scheduled maintenance
- **Deployment Suppression**: Suppress alerts during deployments
- **Business Hours**: Suppress non-critical alerts during business hours

## Troubleshooting

### Common Issues
1. **Services not starting**: Check Docker and Docker Compose installation
2. **Metrics not appearing**: Verify the application is running and accessible
3. **Alerts not firing**: Check alert thresholds and notification channel configuration

### Logs
- Application logs: `logs/`
- Monitoring logs: `logs/monitoring/`
- Alert logs: `logs/alerts/`

## Maintenance

### Regular Tasks
- Monitor alert thresholds and adjust as needed
- Review and update dashboard configurations
- Test alerting system regularly
- Clean up old metrics and logs

### Updates
- Update monitoring configurations as needed
- Add new metrics and alerts as the system evolves
- Review and optimize dashboard performance
EOF

    log_success "Monitoring documentation created"
}

# Main execution
main() {
    log_info "Starting KYB Platform monitoring setup..."
    
    check_dependencies
    create_directories
    setup_prometheus
    setup_alertmanager
    setup_grafana
    create_docker_compose
    build_monitoring_app
    create_startup_script
    create_test_script
    create_documentation
    
    log_success "KYB Platform monitoring setup completed successfully!"
    log_info "Next steps:"
    log_info "1. Run './scripts/start_monitoring.sh' to start the monitoring stack"
    log_info "2. Run './scripts/test_monitoring.sh' to test the monitoring system"
    log_info "3. Access Grafana at http://localhost:3000 (admin/admin123)"
    log_info "4. Review the monitoring documentation in docs/monitoring.md"
}

main "$@"
