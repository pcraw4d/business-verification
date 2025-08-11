#!/bin/bash

# KYB Platform Alerting Rules Validation Script
# This script validates the alerting rules configuration and tests alert delivery

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
ALERTMANAGER_URL="${ALERTMANAGER_URL:-http://localhost:9093}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log "Checking dependencies..."
    
    if ! command -v curl &> /dev/null; then
        error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v promtool &> /dev/null; then
        warning "promtool not found - skipping Prometheus rule validation"
    fi
    
    success "Dependencies check completed"
}

# Validate Prometheus configuration
validate_prometheus_config() {
    log "Validating Prometheus configuration..."
    
    if [ -f "deployments/prometheus/prometheus.yml" ]; then
        if command -v promtool &> /dev/null; then
            if promtool check config deployments/prometheus/prometheus.yml; then
                success "Prometheus configuration is valid"
            else
                error "Prometheus configuration is invalid"
                return 1
            fi
        else
            warning "Skipping Prometheus config validation (promtool not available)"
        fi
    else
        error "Prometheus configuration file not found"
        return 1
    fi
}

# Validate Prometheus alerting rules
validate_alerting_rules() {
    log "Validating Prometheus alerting rules..."
    
    if [ -f "deployments/prometheus/alerts.yml" ]; then
        if command -v promtool &> /dev/null; then
            if promtool check rules deployments/prometheus/alerts.yml; then
                success "Alerting rules are valid"
            else
                error "Alerting rules are invalid"
                return 1
            fi
        else
            warning "Skipping alerting rules validation (promtool not available)"
        fi
    else
        error "Alerting rules file not found"
        return 1
    fi
}

# Validate AlertManager configuration
validate_alertmanager_config() {
    log "Validating AlertManager configuration..."
    
    if [ -f "deployments/alertmanager/alertmanager.yml" ]; then
        if command -v amtool &> /dev/null; then
            if amtool check-config deployments/alertmanager/alertmanager.yml; then
                success "AlertManager configuration is valid"
            else
                error "AlertManager configuration is invalid"
                return 1
            fi
        else
            warning "Skipping AlertManager config validation (amtool not available)"
        fi
    else
        error "AlertManager configuration file not found"
        return 1
    fi
}

# Check Prometheus connectivity
check_prometheus_connectivity() {
    log "Checking Prometheus connectivity..."
    
    if curl -s --max-time 10 "${PROMETHEUS_URL}/-/healthy" > /dev/null; then
        success "Prometheus is accessible"
    else
        error "Prometheus is not accessible at ${PROMETHEUS_URL}"
        return 1
    fi
}

# Check AlertManager connectivity
check_alertmanager_connectivity() {
    log "Checking AlertManager connectivity..."
    
    if curl -s --max-time 10 "${ALERTMANAGER_URL}/-/healthy" > /dev/null; then
        success "AlertManager is accessible"
    else
        error "AlertManager is not accessible at ${ALERTMANAGER_URL}"
        return 1
    fi
}

# Check Grafana connectivity
check_grafana_connectivity() {
    log "Checking Grafana connectivity..."
    
    if curl -s --max-time 10 "${GRAFANA_URL}/api/health" > /dev/null; then
        success "Grafana is accessible"
    else
        error "Grafana is not accessible at ${GRAFANA_URL}"
        return 1
    fi
}

# Get current alerting rules from Prometheus
get_current_alerts() {
    log "Fetching current alerting rules from Prometheus..."
    
    local response
    response=$(curl -s --max-time 10 "${PROMETHEUS_URL}/api/v1/rules")
    
    if [ $? -eq 0 ]; then
        local alert_count
        alert_count=$(echo "$response" | jq '.data.groups[] | select(.name == "kyb-platform-alerts") | .rules | length' 2>/dev/null || echo "0")
        
        if [ "$alert_count" -gt 0 ]; then
            success "Found $alert_count alerting rules in Prometheus"
            echo "$response" | jq '.data.groups[] | select(.name == "kyb-platform-alerts") | .rules[].name' 2>/dev/null || warning "Could not parse alert names"
        else
            warning "No alerting rules found in Prometheus"
        fi
    else
        error "Failed to fetch alerting rules from Prometheus"
        return 1
    fi
}

# Get current firing alerts
get_firing_alerts() {
    log "Checking for currently firing alerts..."
    
    local response
    response=$(curl -s --max-time 10 "${PROMETHEUS_URL}/api/v1/alerts")
    
    if [ $? -eq 0 ]; then
        local firing_count
        firing_count=$(echo "$response" | jq '.data[] | select(.state == "firing") | length' 2>/dev/null || echo "0")
        
        if [ "$firing_count" -gt 0 ]; then
            warning "Found $firing_count currently firing alerts"
            echo "$response" | jq '.data[] | select(.state == "firing") | .labels.alertname' 2>/dev/null || warning "Could not parse firing alert names"
        else
            success "No currently firing alerts"
        fi
    else
        error "Failed to fetch alerts from Prometheus"
        return 1
    fi
}

# Get AlertManager alerts
get_alertmanager_alerts() {
    log "Checking AlertManager alerts..."
    
    local response
    response=$(curl -s --max-time 10 "${ALERTMANAGER_URL}/api/v1/alerts")
    
    if [ $? -eq 0 ]; then
        local alert_count
        alert_count=$(echo "$response" | jq 'length' 2>/dev/null || echo "0")
        
        if [ "$alert_count" -gt 0 ]; then
            success "Found $alert_count alerts in AlertManager"
            echo "$response" | jq '.[].labels.alertname' 2>/dev/null || warning "Could not parse AlertManager alert names"
        else
            success "No alerts in AlertManager"
        fi
    else
        error "Failed to fetch alerts from AlertManager"
        return 1
    fi
}

# Test alert notification delivery
test_alert_notification() {
    log "Testing alert notification delivery..."
    
    # Create a test alert
    local test_alert='{
        "status": "firing",
        "labels": {
            "alertname": "TestAlert",
            "severity": "warning",
            "team": "platform",
            "category": "test"
        },
        "annotations": {
            "summary": "Test alert for validation",
            "description": "This is a test alert to validate notification delivery"
        },
        "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%S.000Z)'",
        "endsAt": "'$(date -u -d '+1 hour' +%Y-%m-%dT%H:%M:%S.000Z)'"
    }'
    
    # Send test alert to AlertManager
    local response
    response=$(curl -s --max-time 10 -X POST \
        -H "Content-Type: application/json" \
        -d "$test_alert" \
        "${ALERTMANAGER_URL}/api/v1/alerts")
    
    if [ $? -eq 0 ]; then
        success "Test alert sent to AlertManager"
        
        # Wait a moment for processing
        sleep 2
        
        # Check if test alert is in AlertManager
        local test_alert_response
        test_alert_response=$(curl -s --max-time 10 "${ALERTMANAGER_URL}/api/v1/alerts")
        
        if echo "$test_alert_response" | jq '.[] | select(.labels.alertname == "TestAlert")' > /dev/null 2>&1; then
            success "Test alert received by AlertManager"
        else
            warning "Test alert not found in AlertManager"
        fi
    else
        error "Failed to send test alert to AlertManager"
        return 1
    fi
}

# Validate alerting rules syntax
validate_rules_syntax() {
    log "Validating alerting rules syntax..."
    
    local alerts_file="deployments/prometheus/alerts.yml"
    local errors=0
    
    if [ ! -f "$alerts_file" ]; then
        error "Alerting rules file not found: $alerts_file"
        return 1
    fi
    
    # Check for basic YAML syntax
    if command -v python3 &> /dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('$alerts_file'))" 2>/dev/null; then
            success "Alerting rules YAML syntax is valid"
        else
            error "Alerting rules YAML syntax is invalid"
            errors=$((errors + 1))
        fi
    else
        warning "Skipping YAML syntax validation (python3 not available)"
    fi
    
    # Check for required fields in alert rules
    local required_fields=("alert" "expr" "for" "labels" "annotations")
    local line_num=0
    
    while IFS= read -r line; do
        line_num=$((line_num + 1))
        
        # Check for alert rule start
        if [[ "$line" =~ ^[[:space:]]*-[[:space:]]*alert:[[:space:]]* ]]; then
            local alert_name
            alert_name=$(echo "$line" | sed 's/.*alert:[[:space:]]*//')
            
            # Check if alert name is provided
            if [ -z "$alert_name" ]; then
                error "Line $line_num: Alert name is missing"
                errors=$((errors + 1))
            fi
        fi
        
        # Check for expression
        if [[ "$line" =~ ^[[:space:]]*expr:[[:space:]]* ]]; then
            local expr
            expr=$(echo "$line" | sed 's/.*expr:[[:space:]]*//')
            
            if [ -z "$expr" ]; then
                error "Line $line_num: Expression is missing"
                errors=$((errors + 1))
            fi
        fi
        
        # Check for severity label
        if [[ "$line" =~ ^[[:space:]]*severity:[[:space:]]* ]]; then
            local severity
            severity=$(echo "$line" | sed 's/.*severity:[[:space:]]*//')
            
            if [[ ! "$severity" =~ ^(critical|warning|info)$ ]]; then
                error "Line $line_num: Invalid severity level: $severity"
                errors=$((errors + 1))
            fi
        fi
    done < "$alerts_file"
    
    if [ $errors -eq 0 ]; then
        success "Alerting rules syntax validation passed"
    else
        error "Alerting rules syntax validation failed with $errors errors"
        return 1
    fi
}

# Generate alerting rules summary
generate_summary() {
    log "Generating alerting rules summary..."
    
    local alerts_file="deployments/prometheus/alerts.yml"
    local total_alerts=0
    local critical_alerts=0
    local warning_alerts=0
    local info_alerts=0
    
    if [ -f "$alerts_file" ]; then
        # Count alerts by severity
        critical_alerts=$(grep -c "severity: critical" "$alerts_file" 2>/dev/null || echo "0")
        warning_alerts=$(grep -c "severity: warning" "$alerts_file" 2>/dev/null || echo "0")
        info_alerts=$(grep -c "severity: info" "$alerts_file" 2>/dev/null || echo "0")
        
        # Ensure we have valid numbers for arithmetic
        critical_alerts=${critical_alerts:-0}
        warning_alerts=${warning_alerts:-0}
        info_alerts=${info_alerts:-0}
        
        # Use expr for arithmetic to avoid shell issues
        total_alerts=$(expr "$critical_alerts" + "$warning_alerts" + "$info_alerts" 2>/dev/null || echo "0")
        
        echo ""
        echo "=== Alerting Rules Summary ==="
        echo "Total Alerts: $total_alerts"
        echo "Critical Alerts: $critical_alerts"
        echo "Warning Alerts: $warning_alerts"
        echo "Info Alerts: $info_alerts"
        echo ""
        
        # List alert categories
        echo "=== Alert Categories ==="
        grep "category:" "$alerts_file" | sort | uniq -c | sed 's/.*category: //' | sort
        echo ""
        
        success "Alerting rules summary generated"
    else
        error "Alerting rules file not found"
        return 1
    fi
}

# Main validation function
main() {
    echo "KYB Platform Alerting Rules Validation"
    echo "======================================"
    echo ""
    
    local exit_code=0
    
    # Run all validation checks
    check_dependencies || exit_code=1
    echo ""
    
    validate_prometheus_config || exit_code=1
    echo ""
    
    validate_alerting_rules || exit_code=1
    echo ""
    
    validate_alertmanager_config || exit_code=1
    echo ""
    
    validate_rules_syntax || exit_code=1
    echo ""
    
    # Connectivity checks (only if services are running)
    if [ "$1" != "--skip-connectivity" ]; then
        check_prometheus_connectivity || exit_code=1
        echo ""
        
        check_alertmanager_connectivity || exit_code=1
        echo ""
        
        check_grafana_connectivity || exit_code=1
        echo ""
        
        get_current_alerts || exit_code=1
        echo ""
        
        get_firing_alerts || exit_code=1
        echo ""
        
        get_alertmanager_alerts || exit_code=1
        echo ""
        
        test_alert_notification || exit_code=1
        echo ""
    fi
    
    generate_summary || exit_code=1
    echo ""
    
    if [ $exit_code -eq 0 ]; then
        success "Alerting rules validation completed successfully"
    else
        error "Alerting rules validation completed with errors"
    fi
    
    exit $exit_code
}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-connectivity    Skip connectivity checks (useful for offline validation)"
    echo "  --help                 Show this help message"
    echo ""
    echo "This script validates the KYB Platform alerting rules configuration."
}

# Parse command line arguments
case "${1:-}" in
    --help)
        show_help
        exit 0
        ;;
    --skip-connectivity)
        main --skip-connectivity
        ;;
    "")
        main
        ;;
    *)
        error "Unknown option: $1"
        show_help
        exit 1
        ;;
esac
