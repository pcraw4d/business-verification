#!/bin/bash
# Phase 5 Performance Test Script
# Tests cache performance, layer distribution, and overall system performance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
ENDPOINT="${API_URL}/v1/classify"
NUM_REQUESTS="${NUM_REQUESTS:-100}"
CONCURRENT="${CONCURRENT:-10}"

echo "⚡ Phase 5 Performance Test - Cache & Layer Distribution"
echo "========================================================"
echo ""
echo "📊 Configuration:"
echo "   API URL: $ENDPOINT"
echo "   Total Requests: $NUM_REQUESTS"
echo "   Concurrent: $CONCURRENT"
echo ""

# Test data - diverse businesses for layer distribution testing
declare -a TEST_BUSINESSES=(
    '{"business_name":"McDonalds","website_url":"https://www.mcdonalds.com"}'
    '{"business_name":"Starbucks Coffee","website_url":"https://www.starbucks.com"}'
    '{"business_name":"Amazon","website_url":"https://www.amazon.com"}'
    '{"business_name":"Microsoft","website_url":"https://www.microsoft.com"}'
    '{"business_name":"Bank of America","website_url":"https://www.bankofamerica.com"}'
    '{"business_name":"Walmart","website_url":"https://www.walmart.com"}'
    '{"business_name":"Apple","website_url":"https://www.apple.com"}'
    '{"business_name":"Tesla","website_url":"https://www.tesla.com"}'
    '{"business_name":"Nike","website_url":"https://www.nike.com"}'
    '{"business_name":"Coca-Cola","website_url":"https://www.coca-cola.com"}'
)

# Counters
total_requests=0
cache_hits=0
cache_misses=0
layer1_count=0
layer2_count=0
layer3_count=0
total_time_ms=0
errors=0

# Temporary files for results
RESULTS_FILE=$(mktemp)
TIMING_FILE=$(mktemp)
LAYER_FILE=$(mktemp)
CACHE_FILE=$(mktemp)

cleanup() {
    rm -f "$RESULTS_FILE" "$TIMING_FILE" "$LAYER_FILE" "$CACHE_FILE"
}
trap cleanup EXIT

# Function to make a single classification request
make_request() {
    local data="$1"
    local request_num="$2"
    
    local start_time=$(date +%s%N)
    local response=$(curl -s --max-time 60 -X POST "$ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$data" 2>&1)
    local end_time=$(date +%s%N)
    local duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    # Parse response
    local from_cache=$(echo "$response" | jq -r '.from_cache // false' 2>/dev/null)
    local processing_path=$(echo "$response" | jq -r '.processing_path // .classification.explanation.layer_used // "unknown"' 2>/dev/null)
    local status_code=$(echo "$response" | jq -r '.status_code // 200' 2>/dev/null)
    
    # Write results to files
    echo "$duration_ms" >> "$TIMING_FILE"
    echo "$processing_path" >> "$LAYER_FILE"
    echo "$from_cache" >> "$CACHE_FILE"
    
    # Check for errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        echo "ERROR" >> "$RESULTS_FILE"
        return 1
    fi
    
    echo "SUCCESS" >> "$RESULTS_FILE"
    return 0
}

echo "🧪 Test 1: Cache Performance Test"
echo "-----------------------------------"
echo ""

# First request (cache miss)
echo "Making first request (cache miss expected)..."
FIRST_DATA='{"business_name":"McDonalds","website_url":"https://www.mcdonalds.com"}'
FIRST_START=$(date +%s%N)
FIRST_RESPONSE=$(curl -s --max-time 60 -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$FIRST_DATA")
FIRST_END=$(date +%s%N)
FIRST_TIME_MS=$(( (FIRST_END - FIRST_START) / 1000000 ))
FIRST_CACHE=$(echo "$FIRST_RESPONSE" | jq -r '.from_cache // false' 2>/dev/null)

echo "   First request time: ${FIRST_TIME_MS}ms"
echo "   From cache: $FIRST_CACHE"
echo ""

# Second request (cache hit expected)
echo "Making second request (cache hit expected)..."
SECOND_START=$(date +%s%N)
SECOND_RESPONSE=$(curl -s --max-time 60 -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$FIRST_DATA")
SECOND_END=$(date +%s%N)
SECOND_TIME_MS=$(( (SECOND_END - SECOND_START) / 1000000 ))
SECOND_CACHE=$(echo "$SECOND_RESPONSE" | jq -r '.from_cache // false' 2>/dev/null)

echo "   Second request time: ${SECOND_TIME_MS}ms"
echo "   From cache: $SECOND_CACHE"
echo ""

if [ "$SECOND_CACHE" = "true" ]; then
    echo -e "${GREEN}✅ Cache working correctly${NC}"
    SPEEDUP=$(echo "scale=2; $FIRST_TIME_MS / $SECOND_TIME_MS" | bc)
    echo "   Speedup: ${SPEEDUP}x faster"
else
    echo -e "${YELLOW}⚠️  Cache not working (may need time to populate)${NC}"
fi
echo ""

# Test 2: Layer Distribution
echo "🧪 Test 2: Layer Distribution Test ($NUM_REQUESTS requests)"
echo "-----------------------------------------------------------"
echo ""

# Clear result files
> "$RESULTS_FILE"
> "$TIMING_FILE"
> "$LAYER_FILE"
> "$CACHE_FILE"

# Make requests in parallel batches
BATCH_SIZE=$CONCURRENT
BATCHES=$(( (NUM_REQUESTS + BATCH_SIZE - 1) / BATCH_SIZE ))

for batch in $(seq 1 $BATCHES); do
    echo "Processing batch $batch/$BATCHES..."
    
    # Calculate how many requests in this batch
    REMAINING=$((NUM_REQUESTS - total_requests))
    CURRENT_BATCH=$((REMAINING < BATCH_SIZE ? REMAINING : BATCH_SIZE))
    
    # Launch concurrent requests
    for i in $(seq 1 $CURRENT_BATCH); do
        # Rotate through test businesses
        BUSINESS_INDEX=$(( (total_requests + i) % ${#TEST_BUSINESSES[@]} ))
        DATA="${TEST_BUSINESSES[$BUSINESS_INDEX]}"
        
        make_request "$DATA" $((total_requests + i)) &
    done
    
    # Wait for batch to complete
    wait
    
    total_requests=$((total_requests + CURRENT_BATCH))
done

echo ""
echo "📊 Analyzing results..."
echo ""

# Count cache hits/misses
cache_hits=$(grep -c "true" "$CACHE_FILE" 2>/dev/null || echo "0")
cache_misses=$(grep -c "false" "$CACHE_FILE" 2>/dev/null || echo "0")

# Count layer distribution
layer1_count=$(grep -c "layer1" "$LAYER_FILE" 2>/dev/null || echo "0")
layer2_count=$(grep -c "layer2" "$LAYER_FILE" 2>/dev/null || echo "0")
layer3_count=$(grep -c "layer3" "$LAYER_FILE" 2>/dev/null || echo "0")

# Calculate timing statistics
if [ -s "$TIMING_FILE" ]; then
    TIMES=($(sort -n "$TIMING_FILE"))
    COUNT=${#TIMES[@]}
    
    # Calculate average
    SUM=0
    for time in "${TIMES[@]}"; do
        SUM=$((SUM + time))
    done
    AVG_MS=$((SUM / COUNT))
    
    # Calculate percentiles
    P50_INDEX=$((COUNT * 50 / 100))
    P95_INDEX=$((COUNT * 95 / 100))
    P99_INDEX=$((COUNT * 99 / 100))
    
    P50_MS=${TIMES[$P50_INDEX]}
    P95_MS=${TIMES[$P95_INDEX]}
    P99_MS=${TIMES[$P99_INDEX]}
    
    MIN_MS=${TIMES[0]}
    MAX_MS=${TIMES[$((COUNT - 1))]}
else
    AVG_MS=0
    P50_MS=0
    P95_MS=0
    P99_MS=0
    MIN_MS=0
    MAX_MS=0
fi

# Count errors
errors=$(grep -c "ERROR" "$RESULTS_FILE" 2>/dev/null || echo "0")
success=$((total_requests - errors))

# Calculate cache hit rate
if [ $total_requests -gt 0 ]; then
    cache_hit_rate=$(echo "scale=2; $cache_hits * 100 / $total_requests" | bc)
else
    cache_hit_rate=0
fi

# Calculate layer percentages
if [ $total_requests -gt 0 ]; then
    layer1_pct=$(echo "scale=1; $layer1_count * 100 / $total_requests" | bc)
    layer2_pct=$(echo "scale=1; $layer2_count * 100 / $total_requests" | bc)
    layer3_pct=$(echo "scale=1; $layer3_count * 100 / $total_requests" | bc)
else
    layer1_pct=0
    layer2_pct=0
    layer3_pct=0
fi

# Display results
echo "=========================================="
echo "📈 Performance Test Results"
echo "=========================================="
echo ""
echo "Request Statistics:"
echo "   Total Requests: $total_requests"
echo "   Successful: $success"
echo "   Errors: $errors"
echo ""

echo "Cache Performance:"
echo "   Cache Hits: $cache_hits"
echo "   Cache Misses: $cache_misses"
echo "   Cache Hit Rate: ${cache_hit_rate}%"
echo ""

echo "Layer Distribution:"
echo "   Layer 1 (Keyword): $layer1_count (${layer1_pct}%)"
echo "   Layer 2 (Embedding): $layer2_count (${layer2_pct}%)"
echo "   Layer 3 (LLM): $layer3_count (${layer3_pct}%)"
echo ""

echo "Response Time Statistics (ms):"
echo "   Average: ${AVG_MS}ms"
echo "   P50 (median): ${P50_MS}ms"
echo "   P95: ${P95_MS}ms"
echo "   P99: ${P99_MS}ms"
echo "   Min: ${MIN_MS}ms"
echo "   Max: ${MAX_MS}ms"
echo ""

# Performance thresholds
echo "Performance Thresholds:"
THRESHOLD_P95=2000
THRESHOLD_AVG=1000
THRESHOLD_CACHE_HIT=20

if [ "$P95_MS" -lt "$THRESHOLD_P95" ]; then
    echo -e "   P95 latency: ${GREEN}✅ PASS${NC} (< ${THRESHOLD_P95}ms)"
else
    echo -e "   P95 latency: ${RED}❌ FAIL${NC} (>= ${THRESHOLD_P95}ms)"
fi

if [ "$AVG_MS" -lt "$THRESHOLD_AVG" ]; then
    echo -e "   Average latency: ${GREEN}✅ PASS${NC} (< ${THRESHOLD_AVG}ms)"
else
    echo -e "   Average latency: ${RED}❌ FAIL${NC} (>= ${THRESHOLD_AVG}ms)"
fi

if [ "$(echo "$cache_hit_rate >= $THRESHOLD_CACHE_HIT" | bc -l)" -eq 1 ]; then
    echo -e "   Cache hit rate: ${GREEN}✅ PASS${NC} (>= ${THRESHOLD_CACHE_HIT}%)"
else
    echo -e "   Cache hit rate: ${YELLOW}⚠️  LOW${NC} (< ${THRESHOLD_CACHE_HIT}%)"
fi

echo ""
echo "=========================================="
echo "✅ Performance test completed"
echo "=========================================="

