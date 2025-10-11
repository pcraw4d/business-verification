#!/bin/bash

# Deployment Validation Script for Risk Assessment Service
# This script validates that the deployed service meets all performance targets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_URL=${1:-""}
VALIDATION_TIMEOUT=300  # 5 minutes
CONCURRENT_REQUESTS=10
TOTAL_REQUESTS=100

# Performance Targets
TARGET_LATENCY_P95=200  # milliseconds
TARGET_LATENCY_P99=300  # milliseconds
TARGET_ERROR_RATE=0.05  # 5%
TARGET_MEMORY_USAGE=1536  # MB (1.5GB)
TARGET_ACCURACY_LSTM=0.85  # 85%
TARGET_ACCURACY_XGBOOST=0.88  # 88%

# Test Results
declare -A test_results
total_tests=0
passed_tests=0

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

# Test result tracking
record_test() {
    local test_name="$1"
    local result="$2"
    local details="$3"
    
    test_results["$test_name"]="$result"
    total_tests=$((total_tests + 1))
    
    if [ "$result" = "PASS" ]; then
        passed_tests=$((passed_tests + 1))
        log_success "$test_name: PASS - $details"
    else
        log_error "$test_name: FAIL - $details"
    fi
}

# Get service URL
get_service_url() {
    if [ -n "$SERVICE_URL" ]; then
        echo "$SERVICE_URL"
        return
    fi
    
    # Try to get URL from Railway
    if command -v railway &> /dev/null; then
        local url=$(railway domain 2>/dev/null || echo "")
        if [ -n "$url" ]; then
            echo "https://$url"
            return
        fi
    fi
    
    # Try to get URL from environment
    if [ -n "$RAILWAY_PUBLIC_DOMAIN" ]; then
        echo "https://$RAILWAY_PUBLIC_DOMAIN"
        return
    fi
    
    log_error "Service URL not provided and could not be determined"
    exit 1
}

# Wait for service to be ready
wait_for_service() {
    local url="$1"
    local max_attempts=30
    local attempt=1
    
    log_info "Waiting for service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$url/health" > /dev/null 2>&1; then
            log_success "Service is ready"
            return 0
        fi
        
        log_info "Attempt $attempt/$max_attempts - service not ready yet..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    log_error "Service failed to become ready within timeout"
    return 1
}

# Test health endpoint
test_health_endpoint() {
    local url="$1"
    
    log_info "Testing health endpoint..."
    
    local response=$(curl -s -w "%{http_code}" "$url/health" -o /tmp/health_response.json)
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        local health_status=$(jq -r '.status' /tmp/health_response.json 2>/dev/null || echo "unknown")
        record_test "health_endpoint" "PASS" "HTTP 200, status: $health_status"
    else
        record_test "health_endpoint" "FAIL" "HTTP $http_code"
    fi
}

# Test metrics endpoint
test_metrics_endpoint() {
    local url="$1"
    
    log_info "Testing metrics endpoint..."
    
    local response=$(curl -s -w "%{http_code}" "$url/metrics" -o /tmp/metrics_response.json)
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        local total_requests=$(jq -r '.overall_metrics.total_requests' /tmp/metrics_response.json 2>/dev/null || echo "0")
        record_test "metrics_endpoint" "PASS" "HTTP 200, total requests: $total_requests"
    else
        record_test "metrics_endpoint" "FAIL" "HTTP $http_code"
    fi
}

# Test risk assessment endpoint
test_risk_assessment() {
    local url="$1"
    
    log_info "Testing risk assessment endpoint..."
    
    local request_data='{
        "business_name": "Test Company",
        "business_address": "123 Test St, Test City, TC 12345",
        "industry": "technology",
        "country": "US",
        "prediction_horizon": 3
    }'
    
    local response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$request_data" \
        "$url/api/v1/assess" \
        -o /tmp/assessment_response.json)
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        local assessment_id=$(jq -r '.id' /tmp/assessment_response.json 2>/dev/null || echo "unknown")
        local risk_score=$(jq -r '.risk_score' /tmp/assessment_response.json 2>/dev/null || echo "unknown")
        record_test "risk_assessment" "PASS" "HTTP 200, ID: $assessment_id, Score: $risk_score"
    else
        record_test "risk_assessment" "FAIL" "HTTP $http_code"
    fi
}

# Test advanced prediction endpoint
test_advanced_prediction() {
    local url="$1"
    
    log_info "Testing advanced prediction endpoint..."
    
    local request_data='{
        "business": {
            "business_name": "Test Company",
            "business_address": "123 Test St, Test City, TC 12345",
            "industry": "technology",
            "country": "US"
        },
        "prediction_horizons": [3, 6, 12],
        "model_preference": "auto",
        "include_temporal_analysis": true,
        "include_scenario_analysis": true,
        "include_model_comparison": true
    }'
    
    local response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$request_data" \
        "$url/api/v1/risk/predict-advanced" \
        -o /tmp/prediction_response.json)
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        local predictions_count=$(jq -r '.predictions | length' /tmp/prediction_response.json 2>/dev/null || echo "0")
        record_test "advanced_prediction" "PASS" "HTTP 200, predictions: $predictions_count"
    else
        record_test "advanced_prediction" "FAIL" "HTTP $http_code"
    fi
}

# Test model-specific endpoints
test_model_endpoints() {
    local url="$1"
    
    log_info "Testing model-specific endpoints..."
    
    # Test XGBoost model info
    local xgb_response=$(curl -s -w "%{http_code}" "$url/api/v1/models/xgboost/info" -o /tmp/xgb_info.json)
    local xgb_code="${xgb_response: -3}"
    
    if [ "$xgb_code" = "200" ]; then
        record_test "xgboost_model_info" "PASS" "HTTP 200"
    else
        record_test "xgboost_model_info" "FAIL" "HTTP $xgb_code"
    fi
    
    # Test LSTM model info
    local lstm_response=$(curl -s -w "%{http_code}" "$url/api/v1/models/lstm/info" -o /tmp/lstm_info.json)
    local lstm_code="${lstm_response: -3}"
    
    if [ "$lstm_code" = "200" ]; then
        record_test "lstm_model_info" "PASS" "HTTP 200"
    else
        record_test "lstm_model_info" "FAIL" "HTTP $lstm_code"
    fi
    
    # Test model performance
    local perf_response=$(curl -s -w "%{http_code}" "$url/api/v1/models/performance" -o /tmp/performance.json)
    local perf_code="${perf_response: -3}"
    
    if [ "$perf_code" = "200" ]; then
        record_test "model_performance" "PASS" "HTTP 200"
    else
        record_test "model_performance" "FAIL" "HTTP $perf_code"
    fi
}

# Performance testing
test_performance() {
    local url="$1"
    
    log_info "Running performance tests..."
    
    # Create test data
    local test_data='{
        "business_name": "Performance Test Company",
        "business_address": "456 Performance Ave, Test City, TC 54321",
        "industry": "technology",
        "country": "US",
        "prediction_horizon": 6
    }'
    
    # Run concurrent requests
    local start_time=$(date +%s)
    local error_count=0
    local response_times=()
    
    for i in $(seq 1 $TOTAL_REQUESTS); do
        (
            local req_start=$(date +%s%3N)
            local response=$(curl -s -w "%{http_code}" -X POST \
                -H "Content-Type: application/json" \
                -d "$test_data" \
                "$url/api/v1/assess" \
                -o /dev/null)
            local req_end=$(date +%s%3N)
            local req_time=$((req_end - req_start))
            
            echo "$req_time" >> /tmp/response_times.txt
            
            local http_code="${response: -3}"
            if [ "$http_code" != "200" ]; then
                echo "1" >> /tmp/error_count.txt
            fi
        ) &
        
        # Limit concurrent requests
        if [ $((i % CONCURRENT_REQUESTS)) -eq 0 ]; then
            wait
        fi
    done
    
    wait
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    # Calculate metrics
    local error_count=$(wc -l < /tmp/error_count.txt 2>/dev/null || echo "0")
    local error_rate=$(echo "scale=4; $error_count / $TOTAL_REQUESTS" | bc -l)
    
    # Calculate percentiles
    sort -n /tmp/response_times.txt > /tmp/sorted_times.txt
    local p95_line=$(echo "scale=0; $TOTAL_REQUESTS * 0.95" | bc -l | cut -d. -f1)
    local p99_line=$(echo "scale=0; $TOTAL_REQUESTS * 0.99" | bc -l | cut -d. -f1)
    
    local p95_latency=$(sed -n "${p95_line}p" /tmp/sorted_times.txt)
    local p99_latency=$(sed -n "${p99_line}p" /tmp/sorted_times.txt)
    
    # Test latency targets
    if [ "$p95_latency" -le "$TARGET_LATENCY_P95" ]; then
        record_test "latency_p95" "PASS" "P95: ${p95_latency}ms (target: ${TARGET_LATENCY_P95}ms)"
    else
        record_test "latency_p95" "FAIL" "P95: ${p95_latency}ms (target: ${TARGET_LATENCY_P95}ms)"
    fi
    
    if [ "$p99_latency" -le "$TARGET_LATENCY_P99" ]; then
        record_test "latency_p99" "PASS" "P99: ${p99_latency}ms (target: ${TARGET_LATENCY_P99}ms)"
    else
        record_test "latency_p99" "FAIL" "P99: ${p99_latency}ms (target: ${TARGET_LATENCY_P99}ms)"
    fi
    
    # Test error rate
    if (( $(echo "$error_rate <= $TARGET_ERROR_RATE" | bc -l) )); then
        record_test "error_rate" "PASS" "Error rate: $(echo "scale=2; $error_rate * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ERROR_RATE * 100" | bc -l)%)"
    else
        record_test "error_rate" "FAIL" "Error rate: $(echo "scale=2; $error_rate * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ERROR_RATE * 100" | bc -l)%)"
    fi
    
    # Calculate throughput
    local throughput=$(echo "scale=2; $TOTAL_REQUESTS / $total_time" | bc -l)
    log_info "Throughput: $throughput requests/second"
    
    # Cleanup temp files
    rm -f /tmp/response_times.txt /tmp/error_count.txt /tmp/sorted_times.txt
}

# Test memory usage
test_memory_usage() {
    local url="$1"
    
    log_info "Testing memory usage..."
    
    # Get memory usage from metrics
    local response=$(curl -s "$url/metrics" 2>/dev/null || echo "{}")
    local memory_usage=$(echo "$response" | jq -r '.overall_metrics.total_memory_usage' 2>/dev/null || echo "0")
    local memory_mb=$((memory_usage / 1024 / 1024))
    
    if [ "$memory_mb" -le "$TARGET_MEMORY_USAGE" ]; then
        record_test "memory_usage" "PASS" "Memory: ${memory_mb}MB (target: ${TARGET_MEMORY_USAGE}MB)"
    else
        record_test "memory_usage" "FAIL" "Memory: ${memory_mb}MB (target: ${TARGET_MEMORY_USAGE}MB)"
    fi
}

# Test model accuracy (if available)
test_model_accuracy() {
    local url="$1"
    
    log_info "Testing model accuracy..."
    
    # Get model performance metrics
    local response=$(curl -s "$url/api/v1/models/performance" 2>/dev/null || echo "{}")
    
    # Check XGBoost accuracy
    local xgb_accuracy=$(echo "$response" | jq -r '.models.xgboost.accuracy' 2>/dev/null || echo "0")
    if (( $(echo "$xgb_accuracy >= $TARGET_ACCURACY_XGBOOST" | bc -l) )); then
        record_test "xgboost_accuracy" "PASS" "Accuracy: $(echo "scale=2; $xgb_accuracy * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ACCURACY_XGBOOST * 100" | bc -l)%)"
    else
        record_test "xgboost_accuracy" "FAIL" "Accuracy: $(echo "scale=2; $xgb_accuracy * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ACCURACY_XGBOOST * 100" | bc -l)%)"
    fi
    
    # Check LSTM accuracy
    local lstm_accuracy=$(echo "$response" | jq -r '.models.lstm.accuracy' 2>/dev/null || echo "0")
    if (( $(echo "$lstm_accuracy >= $TARGET_ACCURACY_LSTM" | bc -l) )); then
        record_test "lstm_accuracy" "PASS" "Accuracy: $(echo "scale=2; $lstm_accuracy * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ACCURACY_LSTM * 100" | bc -l)%)"
    else
        record_test "lstm_accuracy" "FAIL" "Accuracy: $(echo "scale=2; $lstm_accuracy * 100" | bc -l)% (target: $(echo "scale=2; $TARGET_ACCURACY_LSTM * 100" | bc -l)%)"
    fi
}

# Generate validation report
generate_report() {
    local url="$1"
    
    log_info "Generating validation report..."
    
    echo ""
    echo "=========================================="
    echo "DEPLOYMENT VALIDATION REPORT"
    echo "=========================================="
    echo "Service URL: $url"
    echo "Validation Time: $(date)"
    echo "Total Tests: $total_tests"
    echo "Passed Tests: $passed_tests"
    echo "Failed Tests: $((total_tests - passed_tests))"
    echo "Success Rate: $(echo "scale=2; $passed_tests * 100 / $total_tests" | bc -l)%"
    echo ""
    
    echo "Test Results:"
    echo "-------------"
    for test_name in "${!test_results[@]}"; do
        local result="${test_results[$test_name]}"
        if [ "$result" = "PASS" ]; then
            echo -e "✅ $test_name: $result"
        else
            echo -e "❌ $test_name: $result"
        fi
    done
    
    echo ""
    echo "Performance Targets:"
    echo "-------------------"
    echo "Latency P95: ≤${TARGET_LATENCY_P95}ms"
    echo "Latency P99: ≤${TARGET_LATENCY_P99}ms"
    echo "Error Rate: ≤$(echo "scale=2; $TARGET_ERROR_RATE * 100" | bc -l)%"
    echo "Memory Usage: ≤${TARGET_MEMORY_USAGE}MB"
    echo "XGBoost Accuracy: ≥$(echo "scale=2; $TARGET_ACCURACY_XGBOOST * 100" | bc -l)%"
    echo "LSTM Accuracy: ≥$(echo "scale=2; $TARGET_ACCURACY_LSTM * 100" | bc -l)%"
    
    echo ""
    if [ $passed_tests -eq $total_tests ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED! Deployment validation successful.${NC}"
        return 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED! Deployment validation unsuccessful.${NC}"
        return 1
    fi
}

# Cleanup function
cleanup() {
    rm -f /tmp/health_response.json
    rm -f /tmp/metrics_response.json
    rm -f /tmp/assessment_response.json
    rm -f /tmp/prediction_response.json
    rm -f /tmp/xgb_info.json
    rm -f /tmp/lstm_info.json
    rm -f /tmp/performance.json
}

# Help function
show_help() {
    echo "Deployment Validation Script for Risk Assessment Service"
    echo ""
    echo "Usage: $0 [SERVICE_URL]"
    echo ""
    echo "Arguments:"
    echo "  SERVICE_URL    Service URL to validate (optional)"
    echo "                 If not provided, will try to get from Railway"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Auto-detect service URL"
    echo "  $0 https://your-service.railway.app  # Use specific URL"
    echo ""
    echo "Environment Variables:"
    echo "  RAILWAY_PUBLIC_DOMAIN    Railway public domain (auto-detected)"
    echo ""
    echo "Performance Targets:"
    echo "  Latency P95: ≤200ms"
    echo "  Latency P99: ≤300ms"
    echo "  Error Rate: ≤5%"
    echo "  Memory Usage: ≤1.5GB"
    echo "  XGBoost Accuracy: ≥88%"
    echo "  LSTM Accuracy: ≥85%"
}

# Check for help flag
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Check prerequisites
if ! command -v curl &> /dev/null; then
    log_error "curl is required but not installed"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    log_error "jq is required but not installed"
    exit 1
fi

if ! command -v bc &> /dev/null; then
    log_error "bc is required but not installed"
    exit 1
fi

# Trap cleanup on exit
trap cleanup EXIT

# Main execution
main() {
    log_info "Starting deployment validation for Risk Assessment Service"
    echo ""
    
    # Get service URL
    local service_url=$(get_service_url)
    log_info "Service URL: $service_url"
    
    # Wait for service to be ready
    if ! wait_for_service "$service_url"; then
        exit 1
    fi
    
    # Run validation tests
    test_health_endpoint "$service_url"
    test_metrics_endpoint "$service_url"
    test_risk_assessment "$service_url"
    test_advanced_prediction "$service_url"
    test_model_endpoints "$service_url"
    test_performance "$service_url"
    test_memory_usage "$service_url"
    test_model_accuracy "$service_url"
    
    # Generate report
    if generate_report "$service_url"; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
