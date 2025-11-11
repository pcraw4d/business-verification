#!/bin/bash

# Load Testing Script
# Tests API endpoints under load to identify bottlenecks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-https://api-gateway-service-production-21fd.up.railway.app}"
JWT_TOKEN="${JWT_TOKEN:-}"
CONCURRENT_USERS="${CONCURRENT_USERS:-10}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-10}"
TEST_DURATION="${TEST_DURATION:-60}"  # seconds
TEST_RESULTS_DIR="${TEST_RESULTS_DIR:-./test-results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create test results directory
mkdir -p "$TEST_RESULTS_DIR"

# Statistics
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0
TOTAL_RESPONSE_TIME=0
MIN_RESPONSE_TIME=999999
MAX_RESPONSE_TIME=0
RATE_LIMIT_HITS=0

# Function to make a request and measure time
make_request() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="${3:-}"
    
    local start_time=$(date +%s%N)
    
    # Build curl command
    local curl_cmd="curl -s -w '\n%{http_code}\n%{time_total}' -X $method"
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"
    
    if [ -n "$JWT_TOKEN" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $JWT_TOKEN'"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$API_BASE_URL$endpoint'"
    
    local response=$(eval $curl_cmd 2>&1)
    local end_time=$(date +%s%N)
    
    # Parse response
    local http_code=$(echo "$response" | tail -n2 | head -n1)
    local time_total=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d' | sed '$d')
    
    # Convert time_total to milliseconds
    local time_ms=$(echo "$time_total * 1000" | bc | cut -d'.' -f1)
    
    # Update statistics
    TOTAL_REQUESTS=$((TOTAL_REQUESTS + 1))
    
    if [ "$http_code" == "200" ] || [ "$http_code" == "201" ]; then
        SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
    elif [ "$http_code" == "429" ]; then
        RATE_LIMIT_HITS=$((RATE_LIMIT_HITS + 1))
        FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
    else
        FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
    fi
    
    # Update timing statistics
    TOTAL_RESPONSE_TIME=$((TOTAL_RESPONSE_TIME + time_ms))
    if [ $time_ms -lt $MIN_RESPONSE_TIME ]; then
        MIN_RESPONSE_TIME=$time_ms
    fi
    if [ $time_ms -gt $MAX_RESPONSE_TIME ]; then
        MAX_RESPONSE_TIME=$time_ms
    fi
    
    echo "$http_code|$time_ms"
}

# Function to run load test for an endpoint
run_load_test() {
    local test_name="$1"
    local endpoint="$2"
    local method="${3:-GET}"
    local data="${4:-}"
    local duration="${5:-$TEST_DURATION}"
    
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}Load Test: $test_name${NC}"
    echo -e "${BLUE}Endpoint: $method $endpoint${NC}"
    echo -e "${BLUE}Concurrent Users: $CONCURRENT_USERS${NC}"
    echo -e "${BLUE}Duration: ${duration}s${NC}"
    echo -e "${BLUE}==========================================${NC}"
    
    # Reset statistics for this test
    local test_total=0
    local test_successful=0
    local test_failed=0
    local test_total_time=0
    local test_min_time=999999
    local test_max_time=0
    local test_rate_limits=0
    
    # Create temporary file for results
    local results_file="$TEST_RESULTS_DIR/${test_name}_load_${TIMESTAMP}.txt"
    
    # Function to run requests in background
    run_user_requests() {
        local user_id=$1
        local end_time=$(($(date +%s) + duration))
        local user_requests=0
        local user_successful=0
        local user_failed=0
        
        while [ $(date +%s) -lt $end_time ]; do
            local result=$(make_request "$endpoint" "$method" "$data")
            local http_code=$(echo "$result" | cut -d'|' -f1)
            local time_ms=$(echo "$result" | cut -d'|' -f2)
            
            user_requests=$((user_requests + 1))
            
            if [ "$http_code" == "200" ] || [ "$http_code" == "201" ]; then
                user_successful=$((user_successful + 1))
            elif [ "$http_code" == "429" ]; then
                test_rate_limits=$((test_rate_limits + 1))
                user_failed=$((user_failed + 1))
            else
                user_failed=$((user_failed + 1))
            fi
            
            # Update test statistics
            test_total=$((test_total + 1))
            test_total_time=$((test_total_time + time_ms))
            if [ $time_ms -lt $test_min_time ]; then
                test_min_time=$time_ms
            fi
            if [ $time_ms -gt $test_max_time ]; then
                test_max_time=$time_ms
            fi
            
            # Small delay to avoid overwhelming
            sleep 0.1
        done
        
        echo "User $user_id: $user_requests requests, $user_successful successful, $user_failed failed" >> "$results_file"
    }
    
    # Start concurrent users
    local start_time=$(date +%s)
    echo "Starting load test at $(date)..."
    
    for i in $(seq 1 $CONCURRENT_USERS); do
        run_user_requests $i &
    done
    
    # Wait for all background jobs
    wait
    
    local end_time=$(date +%s)
    local actual_duration=$((end_time - start_time))
    
    # Calculate statistics
    local test_successful=$((test_total - test_failed - test_rate_limits))
    local avg_time=0
    if [ $test_total -gt 0 ]; then
        avg_time=$((test_total_time / test_total))
    fi
    
    # Print results
    echo ""
    echo "Load Test Results:"
    echo "  Duration: ${actual_duration}s"
    echo "  Total Requests: $test_total"
    echo "  Successful: $test_successful"
    echo "  Failed: $test_failed"
    echo "  Rate Limited: $test_rate_limits"
    echo "  Success Rate: $(( test_successful * 100 / test_total ))%"
    echo "  Average Response Time: ${avg_time}ms"
    echo "  Min Response Time: ${test_min_time}ms"
    echo "  Max Response Time: ${test_max_time}ms"
    echo "  Requests/Second: $(( test_total / actual_duration ))"
    
    # Save detailed results
    {
        echo "Load Test Results: $test_name"
        echo "Endpoint: $method $endpoint"
        echo "Duration: ${actual_duration}s"
        echo "Concurrent Users: $CONCURRENT_USERS"
        echo ""
        echo "Statistics:"
        echo "  Total Requests: $test_total"
        echo "  Successful: $test_successful"
        echo "  Failed: $test_failed"
        echo "  Rate Limited: $test_rate_limits"
        echo "  Success Rate: $(( test_successful * 100 / test_total ))%"
        echo "  Average Response Time: ${avg_time}ms"
        echo "  Min Response Time: ${test_min_time}ms"
        echo "  Max Response Time: ${test_max_time}ms"
        echo "  Requests/Second: $(( test_total / actual_duration ))"
    } > "$results_file"
}

# Function to test rate limiting
test_rate_limiting() {
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}Rate Limiting Test${NC}"
    echo -e "${BLUE}==========================================${NC}"
    
    echo "Sending rapid requests to test rate limiting..."
    
    local rate_limit_hit=0
    local requests_sent=0
    
    for i in {1..100}; do
        local result=$(make_request "/health" "GET")
        local http_code=$(echo "$result" | cut -d'|' -f1)
        requests_sent=$((requests_sent + 1))
        
        if [ "$http_code" == "429" ]; then
            rate_limit_hit=1
            echo -e "${GREEN}Rate limit hit at request $i${NC}"
            break
        fi
        
        sleep 0.05
    done
    
    if [ $rate_limit_hit -eq 1 ]; then
        echo -e "${GREEN}✓ Rate limiting is working${NC}"
    else
        echo -e "${YELLOW}⚠ Rate limit not hit (may need more requests or higher limits)${NC}"
    fi
}

# Function to test resource usage
test_resource_usage() {
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}Resource Usage Test${NC}"
    echo -e "${BLUE}==========================================${NC}"
    
    echo "Monitoring response times under load..."
    
    # Run load test on health endpoint
    run_load_test "health_check" "/health" "GET" "" "30"
    
    # Run load test on classification endpoint
    run_load_test "classification" "/api/v1/classify" "POST" '{"business_name":"Load Test Company"}' "30"
}

# Function to identify bottlenecks
identify_bottlenecks() {
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}Bottleneck Analysis${NC}"
    echo -e "${BLUE}==========================================${NC}"
    
    echo "Testing different endpoints to identify slowest..."
    
    # Test health check
    echo "Testing health check..."
    local health_times=()
    for i in {1..10}; do
        local result=$(make_request "/health" "GET")
        local time_ms=$(echo "$result" | cut -d'|' -f2)
        health_times+=($time_ms)
    done
    
    # Calculate average
    local health_sum=0
    for time in "${health_times[@]}"; do
        health_sum=$((health_sum + time))
    done
    local health_avg=$((health_sum / ${#health_times[@]}))
    
    echo "  Health Check Average: ${health_avg}ms"
    
    # Test classification
    echo "Testing classification..."
    local classify_times=()
    for i in {1..5}; do
        local result=$(make_request "/api/v1/classify" "POST" '{"business_name":"Bottleneck Test"}')
        local time_ms=$(echo "$result" | cut -d'|' -f2)
        classify_times+=($time_ms)
        sleep 1
    done
    
    # Calculate average
    local classify_sum=0
    for time in "${classify_times[@]}"; do
        classify_sum=$((classify_sum + time))
    done
    local classify_avg=$((classify_sum / ${#classify_times[@]}))
    
    echo "  Classification Average: ${classify_avg}ms"
    
    # Identify bottleneck
    if [ $classify_avg -gt 5000 ]; then
        echo -e "${YELLOW}⚠ Classification endpoint may be a bottleneck (${classify_avg}ms)${NC}"
    fi
    
    if [ $health_avg -gt 100 ]; then
        echo -e "${YELLOW}⚠ Health check may be slow (${health_avg}ms)${NC}"
    fi
}

# Function to generate load test report
generate_report() {
    local report_file="$TEST_RESULTS_DIR/load_test_report_${TIMESTAMP}.txt"
    
    {
        echo "Load Test Report"
        echo "================"
        echo "Date: $(date)"
        echo "API Base URL: $API_BASE_URL"
        echo "Concurrent Users: $CONCURRENT_USERS"
        echo "Test Duration: ${TEST_DURATION}s"
        echo ""
        echo "Overall Statistics:"
        echo "  Total Requests: $TOTAL_REQUESTS"
        echo "  Successful: $SUCCESSFUL_REQUESTS"
        echo "  Failed: $FAILED_REQUESTS"
        echo "  Rate Limited: $RATE_LIMIT_HITS"
        echo "  Success Rate: $(( SUCCESSFUL_REQUESTS * 100 / TOTAL_REQUESTS ))%"
        if [ $TOTAL_REQUESTS -gt 0 ]; then
            echo "  Average Response Time: $(( TOTAL_RESPONSE_TIME / TOTAL_REQUESTS ))ms"
        fi
        echo "  Min Response Time: ${MIN_RESPONSE_TIME}ms"
        echo "  Max Response Time: ${MAX_RESPONSE_TIME}ms"
        echo ""
        echo "Test Results Directory: $TEST_RESULTS_DIR"
    } > "$report_file"
    
    echo ""
    echo "=========================================="
    echo "Load Test Report Generated"
    echo "=========================================="
    echo "Report: $report_file"
    echo "Total Requests: $TOTAL_REQUESTS"
    echo -e "Successful: ${GREEN}$SUCCESSFUL_REQUESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_REQUESTS${NC}"
    echo -e "Rate Limited: ${YELLOW}$RATE_LIMIT_HITS${NC}"
}

# Main execution
main() {
    echo "=========================================="
    echo "KYB Platform Load Testing"
    echo "=========================================="
    echo "API Base URL: $API_BASE_URL"
    echo "Concurrent Users: $CONCURRENT_USERS"
    echo "Test Duration: ${TEST_DURATION}s"
    echo "Test Results Directory: $TEST_RESULTS_DIR"
    echo ""
    
    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        echo -e "${YELLOW}Warning: 'bc' not found. Some calculations may be inaccurate.${NC}"
        echo "Install with: brew install bc (macOS) or apt-get install bc (Linux)"
        echo ""
    fi
    
    # Run load tests
    test_rate_limiting
    test_resource_usage
    identify_bottlenecks
    
    # Generate report
    generate_report
}

# Run main function
main

