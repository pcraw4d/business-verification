#!/bin/bash

# Performance Testing Script for API Gateway
# Measures API response times, tests concurrent requests, and identifies slow queries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}API Gateway Performance Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Configuration
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8080}"
TEST_MERCHANT_ID="${TEST_MERCHANT_ID:-merchant-123}"
ITERATIONS="${ITERATIONS:-50}"
CONCURRENT="${CONCURRENT:-10}"
TIMEOUT="${TIMEOUT:-30}"

echo -e "${YELLOW}Configuration:${NC}"
echo "  API Gateway URL: $API_GATEWAY_URL"
echo "  Test Merchant ID: $TEST_MERCHANT_ID"
echo "  Iterations: $ITERATIONS"
echo "  Concurrent Requests: $CONCURRENT"
echo "  Timeout: ${TIMEOUT}s"
echo ""

# Check if API Gateway is running
if ! curl -s --max-time 5 "${API_GATEWAY_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}Error: API Gateway is not running at ${API_GATEWAY_URL}${NC}"
    echo "Please start the API Gateway and try again."
    exit 1
fi

echo -e "${GREEN}✓ API Gateway is running${NC}"
echo ""

# Function to measure response time
measure_response_time() {
    local method=$1
    local path=$2
    local query_params=$3
    
    local url="${API_GATEWAY_URL}${path}"
    if [ -n "$query_params" ]; then
        url="${url}?${query_params}"
    fi
    
    local start_time=$(date +%s%N)
    local response=$(curl -s -w "\nHTTP_STATUS:%{http_code}\nTIME_TOTAL:%{time_total}\n" \
        -X "$method" \
        -H "Content-Type: application/json" \
        --max-time $TIMEOUT \
        "$url" 2>&1)
    local end_time=$(date +%s%N)
    
    local status_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
    local time_total=$(echo "$response" | grep "TIME_TOTAL" | cut -d: -f2)
    
    # Convert to milliseconds
    local time_ms=$(echo "$time_total * 1000" | bc)
    
    echo "$time_ms $status_code"
}

# Function to run concurrent requests
run_concurrent_requests() {
    local method=$1
    local path=$2
    local query_params=$3
    local count=$4
    local concurrent=$5
    local test_name=$6
    
    echo -e "${BLUE}Testing: ${test_name}${NC}"
    echo "  Path: ${path}"
    echo "  Requests: ${count} (${concurrent} concurrent)"
    
    local temp_file=$(mktemp)
    local pids=()
    local completed=0
    
    # Function to make a single request
    make_request() {
        local result=$(measure_response_time "$method" "$path" "$query_params")
        echo "$result" >> "$temp_file"
    }
    
    # Make requests with concurrency limit
    for ((i=1; i<=count; i++)); do
        # Wait if we've reached concurrency limit
        while [ ${#pids[@]} -ge $concurrent ]; do
            for pid in "${pids[@]}"; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    # Remove finished PID
                    pids=("${pids[@]/$pid}")
                    completed=$((completed + 1))
                fi
            done
            sleep 0.1
        done
        
        # Start new request
        make_request &
        pids+=($!)
    done
    
    # Wait for all requests to complete
    for pid in "${pids[@]}"; do
        wait "$pid" 2>/dev/null || true
    done
    
    # Calculate statistics
    local times=($(awk '{print $1}' "$temp_file" | sort -n))
    local status_codes=($(awk '{print $2}' "$temp_file"))
    
    local count=${#times[@]}
    if [ $count -eq 0 ]; then
        echo -e "${RED}  ✗ No successful requests${NC}"
        rm -f "$temp_file"
        return 1
    fi
    
    # Calculate percentiles
    local min=${times[0]}
    local max=${times[count-1]}
    local p50=${times[$((count * 50 / 100))]}
    local p95=${times[$((count * 95 / 100))]}
    local p99=${times[$((count * 99 / 100))]}
    
    # Calculate mean
    local sum=0
    for time in "${times[@]}"; do
        sum=$(echo "$sum + $time" | bc)
    done
    local mean=$(echo "scale=2; $sum / $count" | bc)
    
    # Count successes
    local successes=0
    for status in "${status_codes[@]}"; do
        if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
            successes=$((successes + 1))
        fi
    done
    
    local success_rate=$(echo "scale=2; $successes * 100 / $count" | bc)
    
    # Display results
    echo -e "${GREEN}  Results:${NC}"
    echo "    Min: ${min}ms"
    echo "    Max: ${max}ms"
    echo "    Mean: ${mean}ms"
    echo "    P50: ${p50}ms"
    echo "    P95: ${p95}ms"
    echo "    P99: ${p99}ms"
    echo "    Success Rate: ${success_rate}%"
    
    # Check if p95 meets requirement (< 500ms)
    if (( $(echo "$p95 < 500" | bc -l) )); then
        echo -e "${GREEN}  ✓ P95 ($p95 ms) is within acceptable range (< 500ms)${NC}"
    else
        echo -e "${RED}  ✗ P95 ($p95 ms) exceeds acceptable range (< 500ms)${NC}"
    fi
    
    # Check success rate
    if (( $(echo "$success_rate >= 95" | bc -l) )); then
        echo -e "${GREEN}  ✓ Success rate ($success_rate%) is acceptable (>= 95%)${NC}"
    else
        echo -e "${RED}  ✗ Success rate ($success_rate%) is below acceptable (>= 95%)${NC}"
    fi
    
    rm -f "$temp_file"
    echo ""
}

# Test endpoints
echo -e "${BLUE}Running Performance Tests...${NC}"
echo ""

# Health check (should be very fast)
run_concurrent_requests "GET" "/health" "" 100 10 "Health Check"

# Merchant endpoints
run_concurrent_requests "GET" "/api/v1/merchants" "" $ITERATIONS $CONCURRENT "Get All Merchants"
run_concurrent_requests "GET" "/api/v1/merchants/${TEST_MERCHANT_ID}" "" $ITERATIONS $CONCURRENT "Get Merchant by ID"
run_concurrent_requests "GET" "/api/v1/merchants/analytics" "" 30 3 "Get Portfolio Analytics"
run_concurrent_requests "GET" "/api/v1/merchants/statistics" "" 30 3 "Get Portfolio Statistics"

# Analytics endpoints
run_concurrent_requests "GET" "/api/v1/analytics/trends" "timeframe=30d" 30 3 "Get Risk Trends"
run_concurrent_requests "GET" "/api/v1/analytics/insights" "timeframe=30d" 30 3 "Get Risk Insights"

# Risk Assessment endpoints
run_concurrent_requests "GET" "/api/v1/risk/benchmarks" "industry=Technology" 30 3 "Get Risk Benchmarks"
run_concurrent_requests "GET" "/api/v1/risk/metrics" "" 30 3 "Get Risk Metrics"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Performance Testing Complete${NC}"
echo -e "${BLUE}========================================${NC}"

