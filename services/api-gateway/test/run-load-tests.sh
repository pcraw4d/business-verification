#!/bin/bash

# Load Testing Script for API Gateway
# Tests API Gateway under various load conditions to identify bottlenecks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}API Gateway Load Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Configuration
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8080}"
CONCURRENT_USERS="${CONCURRENT_USERS:-50}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-20}"
DURATION="${DURATION:-60}" # seconds
TIMEOUT="${TIMEOUT:-30}"

echo -e "${YELLOW}Configuration:${NC}"
echo "  API Gateway URL: $API_GATEWAY_URL"
echo "  Concurrent Users: $CONCURRENT_USERS"
echo "  Requests per User: $REQUESTS_PER_USER"
echo "  Test Duration: ${DURATION}s"
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

# Function to run load test
run_load_test() {
    local endpoint=$1
    local method=$2
    local concurrent=$3
    local requests=$4
    local test_name=$5
    
    echo -e "${BLUE}Running Load Test: ${test_name}${NC}"
    echo "  Endpoint: ${endpoint}"
    echo "  Concurrent Users: ${concurrent}"
    echo "  Requests per User: ${requests}"
    echo "  Total Requests: $((concurrent * requests))"
    
    local temp_file=$(mktemp)
    local pids=()
    local completed=0
    local start_time=$(date +%s)
    
    # Function to make a single request
    make_request() {
        local url="${API_GATEWAY_URL}${endpoint}"
        local request_start=$(date +%s%N)
        
        local response=$(curl -s -w "\nHTTP_STATUS:%{http_code}\nTIME_TOTAL:%{time_total}\n" \
            -X "$method" \
            -H "Content-Type: application/json" \
            --max-time $TIMEOUT \
            "$url" 2>&1)
        
        local request_end=$(date +%s%N)
        local request_time=$((request_end - request_start))
        local request_time_ms=$((request_time / 1000000))
        
        local status_code=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
        local time_total=$(echo "$response" | grep "TIME_TOTAL" | cut -d: -f2)
        
        echo "${request_time_ms} ${status_code}" >> "$temp_file"
    }
    
    # Function for a single user
    user_worker() {
        for ((i=0; i<requests; i++)); do
            make_request
        done
    }
    
    # Launch concurrent users
    for ((i=0; i<concurrent; i++)); do
        user_worker &
        pids+=($!)
    done
    
    # Wait for all users to complete
    for pid in "${pids[@]}"; do
        wait "$pid" 2>/dev/null || true
    done
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
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
        sum=$((sum + time))
    done
    local mean=$((sum / count))
    
    # Count successes and failures
    local successes=0
    local failures=0
    for status in "${status_codes[@]}"; do
        if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
            successes=$((successes + 1))
        else
            failures=$((failures + 1))
        fi
    done
    
    local success_rate=$(echo "scale=2; $successes * 100 / $count" | bc)
    local error_rate=$(echo "scale=2; $failures * 100 / $count" | bc)
    local throughput=$(echo "scale=2; $count / $total_duration" | bc)
    
    # Display results
    echo -e "${GREEN}  Results:${NC}"
    echo "    Total Requests: ${count}"
    echo "    Successful: ${successes}"
    echo "    Failed: ${failures}"
    echo "    Success Rate: ${success_rate}%"
    echo "    Error Rate: ${error_rate}%"
    echo "    Throughput: ${throughput} req/s"
    echo "    Duration: ${total_duration}s"
    echo "    Min Response Time: ${min}ms"
    echo "    Max Response Time: ${max}ms"
    echo "    Mean Response Time: ${mean}ms"
    echo "    P50 Response Time: ${p50}ms"
    echo "    P95 Response Time: ${p95}ms"
    echo "    P99 Response Time: ${p99}ms"
    
    # Check thresholds
    if (( $(echo "$error_rate > 1.0" | bc -l) )); then
        echo -e "${RED}  ✗ Error rate (${error_rate}%) exceeds 1%${NC}"
    else
        echo -e "${GREEN}  ✓ Error rate (${error_rate}%) is acceptable${NC}"
    fi
    
    if (( $(echo "$p95 > 2000" | bc -l) )); then
        echo -e "${YELLOW}  ⚠️  P95 response time (${p95}ms) exceeds 2s - potential bottleneck${NC}"
    else
        echo -e "${GREEN}  ✓ P95 response time (${p95}ms) is acceptable${NC}"
    fi
    
    rm -f "$temp_file"
    echo ""
}

# Test scenarios
echo -e "${BLUE}Running Load Tests...${NC}"
echo ""

# Light load
run_load_test "/health" "GET" 10 10 "Light Load - Health Check"

# Medium load
run_load_test "/api/v1/merchants" "GET" 50 20 "Medium Load - Get Merchants"

# Heavy load
run_load_test "/api/v1/merchants" "GET" 100 50 "Heavy Load - Get Merchants"

# Stress test
run_load_test "/api/v1/merchants" "GET" 200 100 "Stress Test - Get Merchants"

# Database-heavy endpoints
run_load_test "/api/v1/merchants/statistics" "GET" 50 20 "Database Load - Statistics"
run_load_test "/api/v1/merchants/analytics" "GET" 50 20 "Database Load - Analytics"
run_load_test "/api/v1/analytics/trends?timeframe=30d" "GET" 50 20 "Database Load - Risk Trends"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Load Testing Complete${NC}"
echo -e "${BLUE}========================================${NC}"

