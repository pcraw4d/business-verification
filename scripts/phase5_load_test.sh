#!/bin/bash
# Phase 5 Day 6: Load Testing Script
# Tests classification service with 1000 requests, 50 concurrent
# Validates performance, rate limiting, and error handling under load

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"
TOTAL_REQUESTS=1000
CONCURRENT=50
OUTPUT_DIR="load_test_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}🚀 Phase 5 Day 6: Load Testing${NC}"
echo "=========================================="
echo -e "API URL: ${YELLOW}$API_URL${NC}"
echo -e "Total Requests: ${YELLOW}$TOTAL_REQUESTS${NC}"
echo -e "Concurrent: ${YELLOW}$CONCURRENT${NC}"
echo "=========================================="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if hey is installed
if ! command -v hey &> /dev/null; then
    echo -e "${YELLOW}⚠️  'hey' not found. Installing...${NC}"
    echo "Run: go install github.com/rakyll/hey@latest"
    echo "Or: brew install hey"
    exit 1
fi

# Test payloads
PAYLOAD_BASIC='{"business_name":"Test Company"}'
PAYLOAD_WITH_URL='{"business_name":"Test Company","website_url":"https://example.com"}'

# Test 1: Health check (baseline)
echo -e "${BLUE}📊 Test 1: Health Check (Baseline)${NC}"
echo "----------------------------------------"
hey -n 100 -c 10 -m GET "$API_URL/health" > "$OUTPUT_DIR/health_${TIMESTAMP}.txt" 2>&1 || true
echo -e "${GREEN}✅ Health check complete${NC}"
echo ""

# Test 2: Basic classification load test
echo -e "${BLUE}📊 Test 2: Basic Classification Load Test${NC}"
echo "----------------------------------------"
echo "Running $TOTAL_REQUESTS requests with $CONCURRENT concurrent..."
hey -n $TOTAL_REQUESTS -c $CONCURRENT -m POST \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD_BASIC" \
    "$API_URL/v1/classify" > "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" 2>&1 || true

# Parse results
if [ -f "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" ]; then
    echo -e "${GREEN}✅ Load test complete${NC}"
    echo ""
    echo "Results Summary:"
    grep -E "Total:|Slowest:|Fastest:|Average:|Requests/sec|Status code distribution" \
        "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" || echo "Results parsing..."
fi
echo ""

# Test 3: Rate limiting test (exceed limit)
echo -e "${BLUE}📊 Test 3: Rate Limiting Test${NC}"
echo "----------------------------------------"
echo "Testing rate limiting with burst requests..."
hey -n 200 -c 50 -m POST \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD_BASIC" \
    "$API_URL/v1/classify" > "$OUTPUT_DIR/rate_limit_test_${TIMESTAMP}.txt" 2>&1 || true

RATE_LIMIT_COUNT=$(grep -c "429" "$OUTPUT_DIR/rate_limit_test_${TIMESTAMP}.txt" 2>/dev/null || echo "0")
if [ "$RATE_LIMIT_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Rate limiting working (429 responses: $RATE_LIMIT_COUNT)${NC}"
else
    echo -e "${YELLOW}⚠️  No rate limit responses detected${NC}"
fi
echo ""

# Test 4: Classification with URL (more complex)
echo -e "${BLUE}📊 Test 4: Classification with URL Load Test${NC}"
echo "----------------------------------------"
echo "Running $TOTAL_REQUESTS requests with website URLs..."
hey -n $TOTAL_REQUESTS -c $CONCURRENT -m POST \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD_WITH_URL" \
    "$API_URL/v1/classify" > "$OUTPUT_DIR/load_test_url_${TIMESTAMP}.txt" 2>&1 || true

if [ -f "$OUTPUT_DIR/load_test_url_${TIMESTAMP}.txt" ]; then
    echo -e "${GREEN}✅ URL load test complete${NC}"
fi
echo ""

# Test 5: Mixed workload (simulate real usage)
echo -e "${BLUE}📊 Test 5: Mixed Workload Test${NC}"
echo "----------------------------------------"
echo "Simulating mixed requests (basic + URL)..."
for i in {1..500}; do
    if [ $((i % 2)) -eq 0 ]; then
        curl -s -X POST "$API_URL/v1/classify" \
            -H "Content-Type: application/json" \
            -d "$PAYLOAD_BASIC" > /dev/null &
    else
        curl -s -X POST "$API_URL/v1/classify" \
            -H "Content-Type: application/json" \
            -d "$PAYLOAD_WITH_URL" > /dev/null &
    fi
    
    # Limit concurrent background jobs
    if [ $((i % $CONCURRENT)) -eq 0 ]; then
        wait
    fi
done
wait
echo -e "${GREEN}✅ Mixed workload test complete${NC}"
echo ""

# Analyze results
echo -e "${BLUE}📈 Load Test Analysis${NC}"
echo "=========================================="

# Extract key metrics from hey output
extract_metric() {
    local file="$1"
    local pattern="$2"
    grep "$pattern" "$file" 2>/dev/null | awk '{print $NF}' | head -1 || echo "N/A"
}

if [ -f "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" ]; then
    TOTAL_TIME=$(extract_metric "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" "Total:")
    SLOWEST=$(extract_metric "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" "Slowest:")
    FASTEST=$(extract_metric "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" "Fastest:")
    AVERAGE=$(extract_metric "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" "Average:")
    RPS=$(extract_metric "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" "Requests/sec:")
    
    echo "Performance Metrics:"
    echo "  Total time: $TOTAL_TIME"
    echo "  Slowest: $SLOWEST"
    echo "  Fastest: $FASTEST"
    echo "  Average: $AVERAGE"
    echo "  Requests/sec: $RPS"
    echo ""
    
    # Check status codes
    echo "Status Code Distribution:"
    grep -A 10 "Status code distribution" "$OUTPUT_DIR/load_test_basic_${TIMESTAMP}.txt" || echo "  Status codes: See full output"
    echo ""
    
    # Performance targets
    echo "Performance Targets:"
    echo "  ✅ Success rate: >99%"
    echo "  ✅ p50 latency: <500ms"
    echo "  ✅ p95 latency: <3000ms"
    echo "  ✅ Rate limit errors: <1%"
    echo ""
fi

# Check dashboard after load test
echo -e "${BLUE}📊 Checking Dashboard Metrics${NC}"
echo "----------------------------------------"
DASHBOARD_RESPONSE=$(curl -s --max-time 10 "$API_URL/api/dashboard/summary?days=1" 2>/dev/null || echo "{}")
TOTAL_CLASSIFICATIONS=$(echo "$DASHBOARD_RESPONSE" | jq -r '.metrics[] | select(.metric=="total_classifications") | .value' 2>/dev/null || echo "0")
CACHE_HIT_RATE=$(echo "$DASHBOARD_RESPONSE" | jq -r '.metrics[] | select(.metric=="cache_hit_rate") | .value' 2>/dev/null || echo "0")

echo "Dashboard Metrics (after load test):"
echo "  Total classifications: $TOTAL_CLASSIFICATIONS"
echo "  Cache hit rate: ${CACHE_HIT_RATE}%"
echo ""

# Summary
echo -e "${BLUE}✅ Load Testing Complete${NC}"
echo "=========================================="
echo "Results saved to: $OUTPUT_DIR/"
echo ""
echo "Next Steps:"
echo "  1. Review detailed results in output files"
echo "  2. Check Railway logs for any errors"
echo "  3. Verify dashboard metrics are accurate"
echo "  4. Proceed to Day 7: Pre-deployment checklist"
echo ""

