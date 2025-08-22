#!/bin/bash

# V3 API Integration Testing Script
# This script tests all the new v3 API endpoints

set -e

# Configuration
API_BASE_URL="http://localhost:8080/api/v3"
API_KEY="test-api-key-123"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdC11c2VyLTEiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlIjoidXNlciIsImNsaWVudF9pZCI6InRlc3QtY2xpZW50IiwiZXhwIjoxNzM2OTY4MDAwLCJpYXQiOjE3MzY5NjQ0MDAsIm5iZiI6MTczNjk2NDQwMH0.test-signature"
LOG_FILE="v3-api-test.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}❌ $1${NC}" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# Test function
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    
    log "Testing $test_name: $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: ApiKey $API_KEY" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" \
            -X "$method" \
            -H "Authorization: ApiKey $API_KEY" \
            "$API_BASE_URL$endpoint")
    fi
    
    # Extract status code (last line)
    status_code=$(echo "$response" | tail -n1)
    # Extract response body (all lines except last)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" -eq "$expected_status" ]; then
        success "$test_name passed (Status: $status_code)"
        echo "Response: $response_body" | tee -a "$LOG_FILE"
    else
        error "$test_name failed (Expected: $expected_status, Got: $status_code)"
        echo "Response: $response_body" | tee -a "$LOG_FILE"
        return 1
    fi
    
    echo "" | tee -a "$LOG_FILE"
}

# Check if server is running
check_server() {
    log "Checking if server is running..."
    if curl -s "$API_BASE_URL/dashboard" > /dev/null 2>&1; then
        success "Server is running"
    else
        error "Server is not running. Please start the server first."
        exit 1
    fi
}

# Dashboard Tests
test_dashboard_endpoints() {
    log "=== Testing Dashboard Endpoints ==="
    
    test_endpoint "GET" "/dashboard" "" 200 "Get Dashboard Overview"
    test_endpoint "GET" "/dashboard/metrics" "" 200 "Get Dashboard Metrics"
    test_endpoint "GET" "/dashboard/system" "" 200 "Get System Dashboard"
    test_endpoint "GET" "/dashboard/performance" "" 200 "Get Performance Dashboard"
    test_endpoint "GET" "/dashboard/business" "" 200 "Get Business Dashboard"
}

# Alert Management Tests
test_alert_endpoints() {
    log "=== Testing Alert Management Endpoints ==="
    
    # Get all alerts
    test_endpoint "GET" "/alerts" "" 200 "Get All Alerts"
    
    # Create a test alert
    alert_data='{
        "name": "Test High Response Time",
        "description": "Test alert for response time",
        "severity": "warning",
        "category": "performance",
        "condition": "response_time > 500",
        "threshold": 500,
        "duration": "1m",
        "operator": ">",
        "labels": {"environment": "test"},
        "notifications": ["email"]
    }'
    
    test_endpoint "POST" "/alerts" "$alert_data" 201 "Create Alert"
    
    # Get alert history
    test_endpoint "GET" "/alerts/history" "" 200 "Get Alert History"
}

# Escalation Management Tests
test_escalation_endpoints() {
    log "=== Testing Escalation Management Endpoints ==="
    
    # Get escalation policies
    test_endpoint "GET" "/escalation/policies" "" 200 "Get Escalation Policies"
    
    # Create a test escalation policy
    escalation_data='{
        "name": "Test Critical Alert Escalation",
        "description": "Test escalation policy for critical alerts",
        "levels": [
            {
                "level": 1,
                "delay": "5m",
                "notifications": ["email"],
                "recipients": ["test@company.com"]
            }
        ]
    }'
    
    test_endpoint "POST" "/escalation/policies" "$escalation_data" 201 "Create Escalation Policy"
    
    # Get escalation history
    test_endpoint "GET" "/escalation/history" "" 200 "Get Escalation History"
}

# Performance Monitoring Tests
test_performance_endpoints() {
    log "=== Testing Performance Monitoring Endpoints ==="
    
    test_endpoint "GET" "/performance/metrics" "" 200 "Get Performance Metrics"
    test_endpoint "GET" "/performance/metrics/detailed" "" 200 "Get Detailed Performance Metrics"
    test_endpoint "GET" "/performance/alerts" "" 200 "Get Performance Alerts"
    test_endpoint "GET" "/performance/trends" "" 200 "Get Performance Trends"
    test_endpoint "GET" "/performance/optimization/history" "" 200 "Get Optimization History"
    test_endpoint "GET" "/performance/benchmarks" "" 200 "Get Performance Benchmarks"
    
    # Trigger performance optimization
    optimization_data='{
        "target_metrics": ["response_time", "throughput"],
        "constraints": {
            "max_cpu": 80,
            "max_memory": 85
        },
        "strategy": "conservative",
        "dry_run": true
    }'
    
    test_endpoint "POST" "/performance/optimize" "$optimization_data" 200 "Trigger Performance Optimization"
}

# Error Tracking Tests
test_error_endpoints() {
    log "=== Testing Error Tracking Endpoints ==="
    
    # Get all errors
    test_endpoint "GET" "/errors" "" 200 "Get All Errors"
    
    # Create a test error
    error_data='{
        "error_type": "test_validation_error",
        "error_message": "Test error for validation",
        "severity": "warning",
        "category": "test_validation",
        "component": "test_api",
        "endpoint": "/api/v3/test",
        "user_id": "test_user_123",
        "request_id": "test_req_456",
        "context": {"test_input": "invalid_data"},
        "tags": {"environment": "test"}
    }'
    
    test_endpoint "POST" "/errors" "$error_data" 201 "Create Error"
    
    # Get errors by severity
    test_endpoint "GET" "/errors/severity/warning" "" 200 "Get Errors by Severity"
    
    # Get errors by category
    test_endpoint "GET" "/errors/category/test_validation" "" 200 "Get Errors by Category"
    
    # Get error patterns
    test_endpoint "GET" "/errors/patterns" "" 200 "Get Error Patterns"
}

# Business Intelligence Tests
test_business_intelligence_endpoints() {
    log "=== Testing Business Intelligence Endpoints ==="
    
    test_endpoint "GET" "/analytics/business/metrics" "" 200 "Get Business Metrics"
    test_endpoint "GET" "/analytics/performance" "" 200 "Get Performance Analytics"
    test_endpoint "GET" "/analytics/system" "" 200 "Get System Analytics"
    test_endpoint "GET" "/analytics/trends" "" 200 "Get Trend Analysis"
    test_endpoint "GET" "/analytics/report" "" 200 "Get Analytics Report"
    
    # Get custom analytics
    custom_analytics_data='{
        "time_range": "24h",
        "metrics": ["response_time", "throughput", "error_rate"],
        "dimensions": ["endpoint", "user_type"],
        "filters": {"environment": "test"},
        "granularity": "1h"
    }'
    
    test_endpoint "POST" "/analytics/custom" "$custom_analytics_data" 200 "Get Custom Analytics"
}

# Enterprise Integration Tests
test_enterprise_integration_endpoints() {
    log "=== Testing Enterprise Integration Endpoints ==="
    
    test_endpoint "GET" "/integrations/status" "" 200 "Get Integration Status"
    test_endpoint "GET" "/integrations/api-metrics" "" 200 "Get API Metrics"
    test_endpoint "GET" "/integrations/logs" "" 200 "Get Integration Logs"
    
    # Configure integration
    integration_data='{
        "integration_type": "webhook",
        "config": {
            "endpoint": "https://api.example.com/webhook",
            "timeout": 30,
            "retry_count": 3
        },
        "credentials": {
            "api_key": "test_api_key"
        },
        "webhook_url": "https://test-app.com/webhook",
        "filters": {
            "event_types": ["alert", "error"]
        }
    }'
    
    test_endpoint "POST" "/integrations/configure" "$integration_data" 201 "Configure Integration"
    
    # Test integration
    test_data='{
        "integration_type": "webhook",
        "config": {
            "endpoint": "https://api.example.com/webhook"
        }
    }'
    
    test_endpoint "POST" "/integrations/test" "$test_data" 200 "Test Integration"
    
    # Sync data
    sync_data='{
        "integration_type": "database",
        "config": {
            "sync_mode": "incremental"
        }
    }'
    
    test_endpoint "POST" "/integrations/sync" "$sync_data" 200 "Sync Data"
    
    # Handle webhook
    webhook_data='{
        "event_type": "test_alert_fired",
        "event_data": {
            "alert_id": "test_alert_123",
            "severity": "critical"
        },
        "timestamp": "2024-01-15T10:30:00Z",
        "source": "test_monitoring_system",
        "correlation_id": "test_corr_456"
    }'
    
    test_endpoint "POST" "/integrations/webhook" "$webhook_data" 200 "Handle Webhook"
}

# Performance Load Testing
test_performance_load() {
    log "=== Testing Performance Under Load ==="
    
    # Test concurrent requests
    log "Testing 10 concurrent requests to dashboard endpoint..."
    
    for i in {1..10}; do
        (
            response=$(curl -s -w "\n%{http_code}" \
                -H "Authorization: ApiKey $API_KEY" \
                "$API_BASE_URL/dashboard")
            status_code=$(echo "$response" | tail -n1)
            if [ "$status_code" -eq 200 ]; then
                echo "Request $i: ✅ Success"
            else
                echo "Request $i: ❌ Failed (Status: $status_code)"
            fi
        ) &
    done
    
    wait
    
    success "Concurrent load test completed"
}

# Main test execution
main() {
    log "Starting V3 API Integration Tests"
    log "API Base URL: $API_BASE_URL"
    log "Log File: $LOG_FILE"
    
    # Clear log file
    > "$LOG_FILE"
    
    # Check server
    check_server
    
    # Run all test suites
    test_dashboard_endpoints
    test_alert_endpoints
    test_escalation_endpoints
    test_performance_endpoints
    test_error_endpoints
    test_business_intelligence_endpoints
    test_enterprise_integration_endpoints
    
    # Performance testing
    test_performance_load
    
    log "=== Test Summary ==="
    success "All V3 API integration tests completed!"
    log "Check $LOG_FILE for detailed results"
}

# Run main function
main "$@"
