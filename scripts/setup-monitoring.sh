#!/bin/bash

# KYB Platform Monitoring Setup Script
# This script sets up the complete monitoring stack for the KYB Platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="kyb-platform"
MONITORING_NAMESPACE="monitoring"
GRAFANA_ADMIN_PASSWORD="admin123"
PROMETHEUS_RETENTION="15d"

# Functions
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

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

create_directories() {
    log_info "Creating monitoring directories..."
    
    mkdir -p monitoring/data/prometheus
    mkdir -p monitoring/data/grafana
    mkdir -p monitoring/data/alertmanager
    mkdir -p monitoring/logs
    
    # Set proper permissions
    chmod 755 monitoring/data/prometheus
    chmod 755 monitoring/data/grafana
    chmod 755 monitoring/data/alertmanager
    chmod 755 monitoring/logs
    
    log_success "Monitoring directories created"
}

setup_prometheus() {
    log_info "Setting up Prometheus..."
    
    # Create Prometheus configuration if it doesn't exist
    if [ ! -f "monitoring/prometheus.yml" ]; then
        log_warning "Prometheus configuration not found. Using default configuration."
    fi
    
    # Create alert rules if they don't exist
    if [ ! -f "monitoring/alert_rules.yml" ]; then
        log_warning "Alert rules not found. Using default rules."
    fi
    
    log_success "Prometheus configuration ready"
}

setup_grafana() {
    log_info "Setting up Grafana..."
    
    # Create Grafana dashboard if it doesn't exist
    if [ ! -f "monitoring/grafana-dashboard.json" ]; then
        log_warning "Grafana dashboard not found. Using default dashboard."
    fi
    
    # Set Grafana admin password
    export GF_SECURITY_ADMIN_PASSWORD="$GRAFANA_ADMIN_PASSWORD"
    
    log_success "Grafana configuration ready"
}

setup_alertmanager() {
    log_info "Setting up AlertManager..."
    
    # Create AlertManager configuration if it doesn't exist
    if [ ! -f "monitoring/alertmanager.yml" ]; then
        log_warning "AlertManager configuration not found. Using default configuration."
    fi
    
    log_success "AlertManager configuration ready"
}

start_monitoring_stack() {
    log_info "Starting monitoring stack..."
    
    # Start the monitoring stack using Docker Compose
    docker-compose -f docker-compose.monitoring.yml up -d
    
    log_success "Monitoring stack started"
}

wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    # Wait for Prometheus
    log_info "Waiting for Prometheus..."
    until curl -s http://localhost:9090/-/healthy > /dev/null; do
        sleep 2
    done
    log_success "Prometheus is ready"
    
    # Wait for Grafana
    log_info "Waiting for Grafana..."
    until curl -s http://localhost:3000/api/health > /dev/null; do
        sleep 2
    done
    log_success "Grafana is ready"
    
    # Wait for AlertManager
    log_info "Waiting for AlertManager..."
    until curl -s http://localhost:9093/-/healthy > /dev/null; do
        sleep 2
    done
    log_success "AlertManager is ready"
}

import_grafana_dashboard() {
    log_info "Importing Grafana dashboard..."
    
    # Wait a bit more for Grafana to be fully ready
    sleep 10
    
    # Import dashboard
    if [ -f "monitoring/grafana-dashboard.json" ]; then
        curl -X POST \
            -H "Content-Type: application/json" \
            -d @monitoring/grafana-dashboard.json \
            http://admin:$GRAFANA_ADMIN_PASSWORD@localhost:3000/api/dashboards/db
        log_success "Grafana dashboard imported"
    else
        log_warning "Grafana dashboard file not found. Skipping import."
    fi
}

setup_kyb_application() {
    log_info "Setting up KYB application monitoring..."
    
    # Check if the KYB application is running
    if curl -s http://localhost:8080/health > /dev/null; then
        log_success "KYB application is running and accessible"
    else
        log_warning "KYB application is not running. Please start the application first."
    fi
    
    # Test monitoring endpoints
    log_info "Testing monitoring endpoints..."
    
    # Test health endpoint
    if curl -s http://localhost:8080/health > /dev/null; then
        log_success "Health endpoint is accessible"
    else
        log_warning "Health endpoint is not accessible"
    fi
    
    # Test metrics endpoint
    if curl -s http://localhost:8080/metrics > /dev/null; then
        log_success "Metrics endpoint is accessible"
    else
        log_warning "Metrics endpoint is not accessible"
    fi
}

display_summary() {
    log_success "Monitoring setup completed successfully!"
    echo ""
    echo "=== Monitoring Stack Summary ==="
    echo "Prometheus:     http://localhost:9090"
    echo "Grafana:        http://localhost:3000 (admin/$GRAFANA_ADMIN_PASSWORD)"
    echo "AlertManager:   http://localhost:9093"
    echo "Node Exporter:  http://localhost:9100"
    echo "Blackbox:       http://localhost:9115"
    echo ""
    echo "=== KYB Application Endpoints ==="
    echo "Health Check:   http://localhost:8080/health"
    echo "Metrics:        http://localhost:8080/metrics"
    echo "Monitoring API: http://localhost:8080/api/v3/monitoring"
    echo ""
    echo "=== Useful Commands ==="
    echo "View logs:      docker-compose -f docker-compose.monitoring.yml logs -f"
    echo "Stop stack:     docker-compose -f docker-compose.monitoring.yml down"
    echo "Restart stack:  docker-compose -f docker-compose.monitoring.yml restart"
    echo ""
    echo "=== Next Steps ==="
    echo "1. Access Grafana and explore the KYB Platform dashboard"
    echo "2. Configure alert notifications in AlertManager"
    echo "3. Set up additional monitoring rules as needed"
    echo "4. Monitor the application performance and health"
}

cleanup() {
    log_info "Cleaning up..."
    
    # Stop the monitoring stack
    docker-compose -f docker-compose.monitoring.yml down
    
    log_success "Cleanup completed"
}

# Main execution
main() {
    log_info "Starting KYB Platform monitoring setup..."
    
    # Parse command line arguments
    case "${1:-setup}" in
        "setup")
            check_dependencies
            create_directories
            setup_prometheus
            setup_grafana
            setup_alertmanager
            start_monitoring_stack
            wait_for_services
            import_grafana_dashboard
            setup_kyb_application
            display_summary
            ;;
        "start")
            start_monitoring_stack
            wait_for_services
            display_summary
            ;;
        "stop")
            cleanup
            ;;
        "restart")
            cleanup
            sleep 5
            start_monitoring_stack
            wait_for_services
            display_summary
            ;;
        "status")
            docker-compose -f docker-compose.monitoring.yml ps
            ;;
        "logs")
            docker-compose -f docker-compose.monitoring.yml logs -f
            ;;
        *)
            echo "Usage: $0 {setup|start|stop|restart|status|logs}"
            echo ""
            echo "Commands:"
            echo "  setup   - Complete monitoring stack setup (default)"
            echo "  start   - Start the monitoring stack"
            echo "  stop    - Stop the monitoring stack"
            echo "  restart - Restart the monitoring stack"
            echo "  status  - Show status of monitoring services"
            echo "  logs    - Show logs from monitoring services"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
