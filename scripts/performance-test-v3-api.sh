#!/bin/bash

# V3 API Performance Testing Script
# This script performs load testing and performance validation

set -e

# Configuration
API_BASE_URL="http://localhost:8080/api/v3"
API_KEY="test-api-key-123"
LOG_FILE="v3-api-performance.log"
RESULTS_FILE="performance-results.json"

# Test parameters
CONCURRENT_USERS=50
DURATION=60  # seconds
RAMP_UP_TIME=10  # seconds

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

# Performance metrics
declare -A metrics
declare -A response_times
declare -A status_codes

# Initialize metrics
init_metrics() {
    metrics["total_requests"]=0
    metrics["successful_requests"]=0
    metrics["failed_requests"]=0
    metrics["total_response_time"]=0
    metrics["min_response_time"]=999999
    metrics["max_response_time"]=0
    metrics["start_time"]=$(date +%s)
}

# Update metrics
update_metrics() {
    local status_code=$1
    local response_time=$2
    
    metrics["total_requests"]=$((metrics["total_requests"] + 1))
    metrics["total_response_time"]=$((metrics["total_response_time"] + response_time))
    
    if [ "$status_code" -eq 200 ]; then
        metrics["successful_requests"]=$((metrics["successful_requests"] + 1))
    else
        metrics["failed_requests"]=$((metrics["failed_requests"] + 1))
    fi
    
    if [ "$response_time" -lt "${metrics["min_response_time"]}" ]; then
        metrics["min_response_time"]=$response_time
    fi
    
    if [ "$response_time" -gt "${metrics["max_response_time"]}" ]; then
        metrics["max_response_time"]=$response_time
    fi
}

# Calculate average response time
calculate_avg_response_time() {
    if [ "${metrics["total_requests"]}" -gt 0 ]; then
        echo $((metrics["total_response_time"] / metrics["total_requests"]))
    else
        echo 0
    fi
}

# Calculate success rate
calculate_success_rate() {
    if [ "${metrics["total_requests"]}" -gt 0 ]; then
        echo "scale=2; ${metrics["successful_requests"]} * 100 / ${metrics["total_requests"]}" | bc
    else
        echo 0
    fi
}

# Calculate requests per second
calculate_rps() {
    local end_time=$(date +%s)
    local duration=$((end_time - metrics["start_time"]))
    if [ "$duration" -gt 0 ]; then
        echo "scale=2; ${metrics["total_requests"]} / $duration" | bc
    else
        echo 0
    fi
}

# Single request test
test_single_request() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-}
    
    local start_time=$(date +%s%N)
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $API_KEY" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
            -X "$method" \
            -H "Authorization: Bearer $API_KEY" \
            "$API_BASE_URL$endpoint")
    fi
    
    local end_time=$(date +%s%N)
    
    # Parse response
    local response_body=$(echo "$response" | head -n -2)
    local status_code=$(echo "$response" | tail -n 2 | head -n 1)
    local curl_time=$(echo "$response" | tail -n 1)
    
    # Calculate response time in milliseconds
    local response_time_ms=$(echo "scale=0; $curl_time * 1000 / 1" | bc)
    
    # Update metrics
    update_metrics "$status_code" "$response_time_ms"
    
    echo "$status_code:$response_time_ms"
}

# Load test function
load_test() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-}
    local test_name=$4
    local duration=${5:-30}
    
    log "Starting load test: $test_name"
    log "Endpoint: $method $endpoint"
    log "Duration: ${duration}s"
    log "Concurrent users: $CONCURRENT_USERS"
    
    local start_time=$(date +%s)
    local end_time=$((start_time + duration))
    local current_time=$start_time
    
    # Start background processes
    local pids=()
    
    while [ "$current_time" -lt "$end_time" ]; do
        # Start concurrent requests
        for ((i=1; i<=CONCURRENT_USERS; i++)); do
            (
                result=$(test_single_request "$endpoint" "$method" "$data")
                echo "$result" >> "/tmp/load_test_$$_$i"
            ) &
            pids+=($!)
        done
        
        # Wait for all requests to complete
        for pid in "${pids[@]}"; do
            wait "$pid" 2>/dev/null || true
        done
        
        current_time=$(date +%s)
        sleep 1
    done
    
    # Collect results
    local total_requests=0
    local successful_requests=0
    local total_response_time=0
    local min_response_time=999999
    local max_response_time=0
    
    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        if [ -f "/tmp/load_test_$$_$i" ]; then
            while IFS= read -r line; do
                if [ -n "$line" ]; then
                    total_requests=$((total_requests + 1))
                    status_code=$(echo "$line" | cut -d: -f1)
                    response_time=$(echo "$line" | cut -d: -f2)
                    
                    if [ "$status_code" -eq 200 ]; then
                        successful_requests=$((successful_requests + 1))
                    fi
                    
                    total_response_time=$((total_response_time + response_time))
                    
                    if [ "$response_time" -lt "$min_response_time" ]; then
                        min_response_time=$response_time
                    fi
                    
                    if [ "$response_time" -gt "$max_response_time" ]; then
                        max_response_time=$response_time
                    fi
                fi
            done < "/tmp/load_test_$$_$i"
            rm -f "/tmp/load_test_$$_$i"
        fi
    done
    
    # Calculate metrics
    local avg_response_time=0
    if [ "$total_requests" -gt 0 ]; then
        avg_response_time=$((total_response_time / total_requests))
    fi
    
    local success_rate=0
    if [ "$total_requests" -gt 0 ]; then
        success_rate=$(echo "scale=2; $successful_requests * 100 / $total_requests" | bc)
    fi
    
    local rps=0
    if [ "$duration" -gt 0 ]; then
        rps=$(echo "scale=2; $total_requests / $duration" | bc)
    fi
    
    # Log results
    log "Load test completed: $test_name"
    log "Total requests: $total_requests"
    log "Successful requests: $successful_requests"
    log "Success rate: ${success_rate}%"
    log "Average response time: ${avg_response_time}ms"
    log "Min response time: ${min_response_time}ms"
    log "Max response time: ${max_response_time}ms"
    log "Requests per second: $rps"
    
    # Save results to JSON
    cat >> "$RESULTS_FILE" << EOF
{
  "test_name": "$test_name",
  "endpoint": "$method $endpoint",
  "duration": $duration,
  "concurrent_users": $CONCURRENT_USERS,
  "total_requests": $total_requests,
  "successful_requests": $successful_requests,
  "success_rate": $success_rate,
  "avg_response_time_ms": $avg_response_time,
  "min_response_time_ms": $min_response_time,
  "max_response_time_ms": $max_response_time,
  "requests_per_second": $rps,
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    success "Load test completed: $test_name"
}

# Performance benchmarks
run_performance_benchmarks() {
    log "=== Running Performance Benchmarks ==="
    
    # Initialize results file
    echo "[" > "$RESULTS_FILE"
    local first_test=true
    
    # Test 1: Dashboard endpoint (lightweight)
    if [ "$first_test" = true ]; then
        first_test=false
    else
        echo "," >> "$RESULTS_FILE"
    fi
    
    load_test "/dashboard" "GET" "" "Dashboard Overview" 30
    
    # Test 2: Performance metrics (medium load)
    echo "," >> "$RESULTS_FILE"
    load_test "/performance/metrics" "GET" "" "Performance Metrics" 30
    
    # Test 3: Business analytics (heavier load)
    echo "," >> "$RESULTS_FILE"
    load_test "/analytics/business/metrics" "GET" "" "Business Analytics" 30
    
    # Test 4: Error tracking (POST request)
    echo "," >> "$RESULTS_FILE"
    local error_data='{
        "error_type": "performance_test",
        "error_message": "Test error for performance testing",
        "severity": "info",
        "category": "testing",
        "component": "performance_test",
        "endpoint": "/api/v3/test",
        "user_id": "perf_test_user",
        "request_id": "perf_test_req_123",
        "context": {"test_type": "load_test"},
        "tags": {"environment": "performance_test"}
    }'
    
    load_test "/errors" "POST" "$error_data" "Error Tracking POST" 30
    
    # Test 5: Alert creation (POST request)
    echo "," >> "$RESULTS_FILE"
    local alert_data='{
        "name": "Performance Test Alert",
        "description": "Test alert for performance testing",
        "severity": "info",
        "category": "testing",
        "condition": "test_condition > 0",
        "threshold": 1,
        "duration": "1m",
        "operator": ">",
        "labels": {"test": "performance"},
        "notifications": ["test"]
    }'
    
    load_test "/alerts" "POST" "$alert_data" "Alert Creation POST" 30
    
    echo "]" >> "$RESULTS_FILE"
    
    success "All performance benchmarks completed"
}

# Stress test
run_stress_test() {
    log "=== Running Stress Test ==="
    
    local stress_duration=120  # 2 minutes
    local stress_users=100     # 100 concurrent users
    
    log "Stress test parameters:"
    log "Duration: ${stress_duration}s"
    log "Concurrent users: $stress_users"
    
    # Run stress test on dashboard endpoint
    load_test "/dashboard" "GET" "" "Stress Test Dashboard" "$stress_duration"
    
    success "Stress test completed"
}

# Latency test
run_latency_test() {
    log "=== Running Latency Test ==="
    
    local latency_requests=1000
    local endpoint="/dashboard"
    
    log "Latency test: $latency_requests requests to $endpoint"
    
    local start_time=$(date +%s%N)
    
    for ((i=1; i<=latency_requests; i++)); do
        result=$(test_single_request "$endpoint")
        status_code=$(echo "$result" | cut -d: -f1)
        response_time=$(echo "$result" | cut -d: -f2)
        
        if [ "$status_code" -eq 200 ]; then
            echo "$response_time" >> "/tmp/latency_test_$$"
        fi
    done
    
    local end_time=$(date +%s%N)
    local total_time=$(((end_time - start_time) / 1000000))  # Convert to milliseconds
    
    # Calculate latency statistics
    if [ -f "/tmp/latency_test_$$" ]; then
        local avg_latency=$(awk '{ sum += $1 } END { print int(sum/NR) }' "/tmp/latency_test_$$")
        local p50_latency=$(sort -n "/tmp/latency_test_$$" | awk 'NR == int(NR/2) + 1')
        local p95_latency=$(sort -n "/tmp/latency_test_$$" | awk 'NR == int(NR*0.95)')
        local p99_latency=$(sort -n "/tmp/latency_test_$$" | awk 'NR == int(NR*0.99)')
        
        log "Latency test results:"
        log "Total time: ${total_time}ms"
        log "Average latency: ${avg_latency}ms"
        log "P50 latency: ${p50_latency}ms"
        log "P95 latency: ${p95_latency}ms"
        log "P99 latency: ${p99_latency}ms"
        
        rm -f "/tmp/latency_test_$$"
    fi
    
    success "Latency test completed"
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

# Generate performance report
generate_report() {
    log "=== Generating Performance Report ==="
    
    if [ -f "$RESULTS_FILE" ]; then
        log "Performance results saved to: $RESULTS_FILE"
        
        # Calculate overall metrics
        local total_tests=$(jq length "$RESULTS_FILE")
        local avg_success_rate=$(jq -r '[.[].success_rate] | add / length' "$RESULTS_FILE")
        local avg_response_time=$(jq -r '[.[].avg_response_time_ms] | add / length' "$RESULTS_FILE")
        local avg_rps=$(jq -r '[.[].requests_per_second] | add / length' "$RESULTS_FILE")
        
        log "Overall Performance Summary:"
        log "Total tests: $total_tests"
        log "Average success rate: ${avg_success_rate}%"
        log "Average response time: ${avg_response_time}ms"
        log "Average requests per second: $avg_rps"
        
        # Check performance thresholds
        if (( $(echo "$avg_response_time < 500" | bc -l) )); then
            success "✅ Response time meets target (< 500ms)"
        else
            warning "⚠️  Response time exceeds target (${avg_response_time}ms)"
        fi
        
        if (( $(echo "$avg_success_rate > 95" | bc -l) )); then
            success "✅ Success rate meets target (> 95%)"
        else
            warning "⚠️  Success rate below target (${avg_success_rate}%)"
        fi
        
        if (( $(echo "$avg_rps > 10" | bc -l) )); then
            success "✅ Throughput meets target (> 10 RPS)"
        else
            warning "⚠️  Throughput below target ($avg_rps RPS)"
        fi
    else
        error "No performance results found"
    fi
}

# Main execution
main() {
    log "Starting V3 API Performance Tests"
    log "API Base URL: $API_BASE_URL"
    log "Log File: $LOG_FILE"
    log "Results File: $RESULTS_FILE"
    
    # Clear log file
    > "$LOG_FILE"
    
    # Check server
    check_server
    
    # Initialize metrics
    init_metrics
    
    # Run performance tests
    run_performance_benchmarks
    run_stress_test
    run_latency_test
    
    # Generate report
    generate_report
    
    log "=== Performance Test Summary ==="
    success "All V3 API performance tests completed!"
    log "Check $LOG_FILE for detailed logs"
    log "Check $RESULTS_FILE for performance results"
}

# Run main function
main "$@"
