#!/bin/bash

# Classification Service Performance Monitoring Script
# Monitors production logs for performance metrics

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "=========================================="
echo "Classification Service Performance Monitor"
echo "=========================================="
echo ""
echo "Service URL: ${SERVICE_URL}"
echo ""
echo "This script helps you monitor performance by:"
echo "1. Making test requests to the service"
echo "2. Measuring response times"
echo "3. Checking for cache hits/misses"
echo "4. Identifying performance patterns"
echo ""
echo "Note: For detailed log analysis, check Railway Dashboard → Classification Service → Logs"
echo ""

# Function to make a test request and measure performance
test_request() {
    local request_num=$1
    local description=$2
    local use_cache=$3  # "true" to use same data for cache test
    
    local business_name
    local description_text
    
    if [ "$use_cache" = "true" ]; then
        business_name="Performance Test Company"
        description_text="A test business for performance monitoring"
    else
        business_name="Performance Test Company $(date +%s)"
        description_text="A test business for performance monitoring - $(date +%s)"
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Request #${request_num}: ${description}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Measure time
    start_time=$(date +%s.%N)
    
    # Make request
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
        -X POST "${SERVICE_URL}/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"${business_name}\",
            \"description\": \"${description_text}\",
            \"website_url\": \"https://example.com\"
        }" 2>&1)
    
    end_time=$(date +%s.%N)
    
    # Extract HTTP code and time (handle macOS head command limitations)
    # Response format: body\nhttp_code\ntime_total
    http_code=$(echo "$response" | grep -E '^[0-9]{3}$' | tail -1)
    curl_time=$(echo "$response" | grep -E '^[0-9]+\.[0-9]+$' | tail -1)
    # Body is everything except the last 2 lines (http_code and time)
    body=$(echo "$response" | sed '$d' | sed '$d')
    
    # Calculate actual time
    actual_time=$(echo "$end_time - $start_time" | bc 2>/dev/null || echo "$curl_time")
    
    # Check for cache header
    cache_header=$(curl -s -I -X POST "${SERVICE_URL}/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"${business_name}\",
            \"description\": \"${description_text}\",
            \"website_url\": \"https://example.com\"
        }" 2>&1 | grep -i "x-cache" | tr -d '\r' || echo "X-Cache: (not present)")
    
    echo "HTTP Status: ${http_code}"
    echo "Response Time: ${curl_time}s"
    echo "Cache Header: ${cache_header}"
    
    # Check if response is valid
    if [ "$http_code" = "200" ]; then
        echo "${GREEN}✅ Request successful${NC}"
        
        # Try to extract processing time from response if available
        if command -v jq &> /dev/null; then
            processing_time=$(echo "$body" | jq -r '.processing_time // .metadata.processing_time // "N/A"' 2>/dev/null || echo "N/A")
            if [ "$processing_time" != "N/A" ] && [ "$processing_time" != "null" ]; then
                echo "Processing Time: ${processing_time}"
            fi
        fi
    else
        echo "${RED}❌ Request failed with status ${http_code}${NC}"
    fi
    
    echo ""
    
    # Return time for analysis
    echo "$curl_time"
}

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo "${RED}Error: curl is required but not found${NC}"
    exit 1
fi

# Check if bc is available (for time calculations)
if ! command -v bc &> /dev/null; then
    echo "${YELLOW}Warning: bc not found. Time calculations may be limited.${NC}"
    echo "Install with: brew install bc (macOS) or apt-get install bc (Linux)"
    echo ""
fi

echo "${BLUE}Starting performance monitoring tests...${NC}"
echo ""

# Test 1: First request (should be slower, cache MISS)
echo "${YELLOW}Test 1: First Request (Expected: Cache MISS, slower)${NC}"
time1=$(test_request 1 "First request - baseline performance" "false")
echo ""

# Wait a moment
sleep 2

# Test 2: Second request with same data (should be faster, cache HIT)
echo "${YELLOW}Test 2: Second Request - Same Data (Expected: Cache HIT, faster)${NC}"
time2=$(test_request 2 "Second request - should hit cache" "true")
echo ""

# Test 3: Third request with different data (should be slower, cache MISS)
echo "${YELLOW}Test 3: Third Request - Different Data (Expected: Cache MISS, slower)${NC}"
time3=$(test_request 3 "Third request - different data, no cache" "false")
echo ""

# Summary
echo "=========================================="
echo "Performance Summary"
echo "=========================================="
echo ""
echo "Request 1 (MISS): ${time1}s"
echo "Request 2 (HIT):  ${time2}s"
echo "Request 3 (MISS): ${time3}s"
echo ""

# Analyze performance
if [ -n "$time1" ] && [ -n "$time2" ]; then
    # Check if time2 is faster (cache working)
    if (( $(echo "$time2 < $time1" | bc -l 2>/dev/null || echo "0") )); then
        improvement=$(echo "scale=2; (($time1 - $time2) / $time1) * 100" | bc 2>/dev/null || echo "N/A")
        echo "${GREEN}✅ Cache Performance:${NC}"
        echo "   Request 2 was faster than Request 1"
        if [ "$improvement" != "N/A" ]; then
            echo "   Performance improvement: ${improvement}%"
        fi
    else
        echo "${YELLOW}⚠️  Cache Performance:${NC}"
        echo "   Request 2 was not faster than Request 1"
        echo "   Cache may not be working optimally"
    fi
fi

echo ""
echo "=========================================="
echo "Performance Targets (from plan)"
echo "=========================================="
echo ""
echo "Fast-path scraping: 2-4s (down from 30-60s)"
echo "Regular scraping: 8-12s (down from 60-90s)"
echo "Cached requests: 0.1-0.2s"
echo ""

# Check if times meet targets
if [ -n "$time1" ]; then
    if (( $(echo "$time1 < 5" | bc -l 2>/dev/null || echo "0") )); then
        echo "${GREEN}✅ Request 1 time ($time1 s) meets fast-path target (<5s)${NC}"
    elif (( $(echo "$time1 < 12" | bc -l 2>/dev/null || echo "0") )); then
        echo "${YELLOW}⚠️  Request 1 time ($time1 s) meets regular target (<12s) but not fast-path${NC}"
    else
        echo "${RED}❌ Request 1 time ($time1 s) exceeds target (>12s)${NC}"
    fi
fi

if [ -n "$time2" ]; then
    if (( $(echo "$time2 < 0.5" | bc -l 2>/dev/null || echo "0") )); then
        echo "${GREEN}✅ Request 2 time ($time2 s) meets cached request target (<0.5s)${NC}"
    elif (( $(echo "$time2 < 2" | bc -l 2>/dev/null || echo "0") )); then
        echo "${YELLOW}⚠️  Request 2 time ($time2 s) is fast but could be faster for cache hit${NC}"
    else
        echo "${RED}❌ Request 2 time ($time2 s) is slow for a cached request${NC}"
    fi
fi

echo ""
echo "=========================================="
echo "Next Steps"
echo "=========================================="
echo ""
echo "1. Check Railway logs for detailed metrics:"
echo "   Railway Dashboard → Classification Service → Logs"
echo ""
echo "2. Look for these log patterns:"
echo "   - [FAST-PATH] - Fast-path mode usage"
echo "   - [PARALLEL] - Parallel processing"
echo "   - Sufficient content - Early exit"
echo "   - Cache hit/miss messages"
echo ""
echo "3. Monitor these metrics:"
echo "   - Website scraping times"
echo "   - Success rates"
echo "   - Timeout rates"
echo "   - Cache hit rates"
echo ""

