#!/bin/bash
# Performance Testing Script for Classification Service
# Tests website scraping performance, caching, and overall response times

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "⚡ Classification Service Performance Test"
echo "==========================================="
echo ""

# Configuration
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
PYTHON_ML_SERVICE_URL="${PYTHON_ML_SERVICE_URL:-http://localhost:8000}"
NUM_REQUESTS="${NUM_REQUESTS:-10}"
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-3}"

echo "📊 Test Configuration:"
echo "   Classification Service: $CLASSIFICATION_SERVICE_URL"
echo "   Python ML Service: $PYTHON_ML_SERVICE_URL"
echo "   Number of Requests: $NUM_REQUESTS"
echo "   Concurrent Requests: $CONCURRENT_REQUESTS"
echo ""

# Test data samples
declare -a TEST_BUSINESSES=(
    '{"business_name":"Acme Corporation","description":"Technology consulting and software development","website_url":"https://www.acme.com"}'
    '{"business_name":"Joe'\''s Pizza","description":"Family-owned pizza restaurant serving authentic Italian cuisine","website_url":"https://www.joespizza.com"}'
    '{"business_name":"City Medical Center","description":"Full-service hospital providing emergency and specialized care","website_url":"https://www.citymedical.com"}'
    '{"business_name":"Green Energy Solutions","description":"Renewable energy consulting and solar panel installation","website_url":"https://www.greenenergy.com"}'
    '{"business_name":"Main Street Bank","description":"Community bank offering checking, savings, and loan services","website_url":"https://www.mainstreetbank.com"}'
)

# Function to make a classification request
make_request() {
    local url="$1"
    local data="$2"
    local request_num="$3"
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "${url}/v1/classify" 2>/dev/null || echo "ERROR")
    
    end_time=$(date +%s.%N)
    
    if [[ "$response" == *"ERROR"* ]]; then
        echo "❌ Request $request_num failed"
        return 1
    fi
    
    http_code=$(echo "$response" | tail -n2 | head -n1)
    time_total=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d' | sed '$d')
    
    # Calculate latency
    latency=$(echo "$end_time - $start_time" | bc 2>/dev/null || echo "$time_total")
    
    echo "$request_num|$http_code|$latency|$time_total"
}

# Function to run performance test
run_performance_test() {
    local url="$1"
    local num_requests="$2"
    local concurrent="$3"
    
    echo "🚀 Starting performance test..."
    echo ""
    
    local results_file=$(mktemp)
    local pids=()
    local request_count=0
    
    # Run requests
    for ((i=1; i<=num_requests; i++)); do
        # Select test data (cycle through available samples)
        test_index=$(( (i-1) % ${#TEST_BUSINESSES[@]} ))
        test_data="${TEST_BUSINESSES[$test_index]}"
        
        # Make request in background
        (make_request "$url" "$test_data" "$i" >> "$results_file" 2>&1) &
        pids+=($!)
        
        # Limit concurrent requests
        if (( i % concurrent == 0 )); then
            wait "${pids[@]}"
            pids=()
        fi
    done
    
    # Wait for remaining requests
    wait "${pids[@]}"
    
    # Analyze results
    echo "📊 Performance Results:"
    echo ""
    
    if [ ! -s "$results_file" ]; then
        echo -e "${RED}❌ No results collected${NC}"
        rm "$results_file"
        return 1
    fi
    
    # Parse results
    local success_count=0
    local error_count=0
    local total_latency=0
    local latencies=()
    
    while IFS='|' read -r req_num http_code latency time_total; do
        if [[ "$http_code" == "200" ]]; then
            ((success_count++))
            total_latency=$(echo "$total_latency + $latency" | bc 2>/dev/null || echo "$total_latency")
            latencies+=("$latency")
        else
            ((error_count++))
        fi
    done < "$results_file"
    
    # Calculate statistics
    if [ $success_count -gt 0 ]; then
        avg_latency=$(echo "scale=3; $total_latency / $success_count" | bc 2>/dev/null || echo "N/A")
        
        # Sort latencies for percentile calculation
        IFS=$'\n' sorted_latencies=($(sort -n <<<"${latencies[*]}"))
        unset IFS
        
        p95_index=$(( success_count * 95 / 100 ))
        p99_index=$(( success_count * 99 / 100 ))
        
        p95_latency="${sorted_latencies[$p95_index]:-N/A}"
        p99_latency="${sorted_latencies[$p99_index]:-N/A}"
        
        echo "   Total Requests: $num_requests"
        echo -e "   ${GREEN}Successful: $success_count${NC}"
        if [ $error_count -gt 0 ]; then
            echo -e "   ${RED}Failed: $error_count${NC}"
        fi
        echo ""
        echo "   Latency Statistics:"
        echo "   - Average: ${avg_latency}s"
        echo "   - P95: ${p95_latency}s (target: < 8s)"
        echo "   - P99: ${p99_latency}s (target: < 12s)"
        echo ""
        
        # Check against targets
        if command -v bc &> /dev/null; then
            if (( $(echo "$avg_latency < 5" | bc -l) )); then
                echo -e "   ${GREEN}✅ Average latency meets target (< 5s)${NC}"
            else
                echo -e "   ${YELLOW}⚠️  Average latency exceeds target (>= 5s)${NC}"
            fi
            
            if [ "$p95_latency" != "N/A" ] && (( $(echo "$p95_latency < 8" | bc -l) )); then
                echo -e "   ${GREEN}✅ P95 latency meets target (< 8s)${NC}"
            elif [ "$p95_latency" != "N/A" ]; then
                echo -e "   ${YELLOW}⚠️  P95 latency exceeds target (>= 8s)${NC}"
            fi
        fi
    else
        echo -e "${RED}❌ All requests failed${NC}"
    fi
    
    rm "$results_file"
}

# Check if services are available
echo "🔍 Checking service availability..."
echo ""

# Check classification service
if curl -f -s "${CLASSIFICATION_SERVICE_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Classification service is available${NC}"
else
    echo -e "${RED}❌ Classification service is not available at $CLASSIFICATION_SERVICE_URL${NC}"
    echo "   Make sure the service is running"
    exit 1
fi

# Check Python ML service (optional)
if [ -n "$PYTHON_ML_SERVICE_URL" ]; then
    if curl -f -s "${PYTHON_ML_SERVICE_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Python ML service is available${NC}"
    else
        echo -e "${YELLOW}⚠️  Python ML service is not available (tests will use fallback)${NC}"
    fi
fi

echo ""

# Run performance test
run_performance_test "$CLASSIFICATION_SERVICE_URL" "$NUM_REQUESTS" "$CONCURRENT_REQUESTS"

echo ""
echo "💡 Usage Tips:"
echo "   - Set CLASSIFICATION_SERVICE_URL to test different service"
echo "   - Set NUM_REQUESTS to change number of test requests"
echo "   - Set CONCURRENT_REQUESTS to change concurrency level"
echo "   - Install 'bc' for better latency calculations: brew install bc"
echo ""

