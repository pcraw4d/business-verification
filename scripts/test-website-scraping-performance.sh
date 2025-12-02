#!/bin/bash

# Website Scraping Performance Test Script
# Tests the website scraping optimizations against target metrics

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "=========================================="
echo "Website Scraping Performance Test"
echo "=========================================="
echo ""
echo "Service URL: ${SERVICE_URL}"
echo ""
echo "Target Metrics (from optimization plan):"
echo "  - Fast-path scraping: 2-4s (down from 30-60s)"
echo "  - Regular scraping: 8-12s (down from 60-90s)"
echo "  - Request success rate: >80% (up from ~0%)"
echo "  - ML service utilization: >80% (up from 0%)"
echo ""

# Test websites (various types for comprehensive testing)
TEST_WEBSITES=(
    "https://example.com"
    "https://www.wikipedia.org"
    "https://github.com"
)

# Function to test a single website
test_website() {
    local website=$1
    local test_num=$2
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Test ${test_num}: ${website}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Create unique business name for each test
    local business_name="Test Business $(date +%s)-${test_num}"
    
    # Measure time
    start_time=$(date +%s.%N)
    
    # Make request
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
        -X POST "${SERVICE_URL}/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"${business_name}\",
            \"description\": \"A test business for website scraping performance testing\",
            \"website_url\": \"${website}\"
        }" 2>&1)
    
    end_time=$(date +%s.%N)
    
    # Extract HTTP code and time
    http_code=$(echo "$response" | tail -2 | head -1)
    curl_time=$(echo "$response" | tail -1)
    body=$(echo "$response" | head -n -2)
    
    # Calculate actual time
    actual_time=$(echo "$end_time - $start_time" | bc 2>/dev/null || echo "$curl_time")
    
    echo "HTTP Status: ${http_code}"
    echo "Response Time: ${curl_time}s"
    
    # Check response
    if [ "$http_code" = "200" ]; then
        echo "${GREEN}✅ Request successful${NC}"
        
        # Try to extract processing time and other metrics
        if command -v jq &> /dev/null; then
            processing_time=$(echo "$body" | jq -r '.processing_time // "N/A"' 2>/dev/null || echo "N/A")
            if [ "$processing_time" != "N/A" ] && [ "$processing_time" != "null" ]; then
                # Convert nanoseconds to seconds if needed
                if [[ "$processing_time" == *"ns"* ]]; then
                    processing_time=$(echo "$processing_time" | sed 's/ns//' | awk '{print $1/1000000000}')
                fi
                echo "Processing Time: ${processing_time}s"
            fi
            
            # Check for website analysis info
            website_analyzed=$(echo "$body" | jq -r '.classification.website_content.scraped // false' 2>/dev/null || echo "false")
            if [ "$website_analyzed" = "true" ]; then
                echo "${GREEN}✅ Website was analyzed${NC}"
            fi
        fi
        
        # Check if response time meets targets
        if (( $(echo "$curl_time < 5" | bc -l 2>/dev/null || echo "0") )); then
            echo "${GREEN}✅ Meets fast-path target (<5s)${NC}"
        elif (( $(echo "$curl_time < 12" | bc -l 2>/dev/null || echo "0") )); then
            echo "${YELLOW}⚠️  Meets regular target (<12s) but not fast-path${NC}"
        else
            echo "${RED}❌ Exceeds target (>12s)${NC}"
        fi
        
        return 0
    else
        echo "${RED}❌ Request failed with status ${http_code}${NC}"
        if [ -n "$body" ]; then
            echo "Error: ${body:0:200}"
        fi
        return 1
    fi
}

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo "${RED}Error: curl is required but not found${NC}"
    exit 1
fi

# Check if bc is available
if ! command -v bc &> /dev/null; then
    echo "${YELLOW}Warning: bc not found. Time comparisons may be limited.${NC}"
fi

echo "${BLUE}Starting performance tests...${NC}"
echo ""

# Run tests
success_count=0
total_count=0
total_time=0
fast_path_count=0
regular_count=0
slow_count=0

for i in "${!TEST_WEBSITES[@]}"; do
    website="${TEST_WEBSITES[$i]}"
    test_num=$((i + 1))
    
    if test_website "$website" "$test_num"; then
        success_count=$((success_count + 1))
        
        # Extract time (simplified - would need better parsing)
        time_val=$(echo "$curl_time" | grep -oE '[0-9]+\.[0-9]+' | head -1 || echo "0")
        if [ -n "$time_val" ]; then
            total_time=$(echo "$total_time + $time_val" | bc 2>/dev/null || echo "$total_time")
            
            if (( $(echo "$time_val < 5" | bc -l 2>/dev/null || echo "0") )); then
                fast_path_count=$((fast_path_count + 1))
            elif (( $(echo "$time_val < 12" | bc -l 2>/dev/null || echo "0") )); then
                regular_count=$((regular_count + 1))
            else
                slow_count=$((slow_count + 1))
            fi
        fi
    fi
    
    total_count=$((total_count + 1))
    
    # Wait between tests
    if [ $i -lt $((${#TEST_WEBSITES[@]} - 1)) ]; then
        echo "Waiting 3 seconds before next test..."
        sleep 3
        echo ""
    fi
done

# Summary
echo "=========================================="
echo "Performance Test Summary"
echo "=========================================="
echo ""

if [ $total_count -gt 0 ]; then
    success_rate=$(echo "scale=2; ($success_count / $total_count) * 100" | bc 2>/dev/null || echo "N/A")
    echo "Total Tests: ${total_count}"
    echo "Successful: ${success_count}"
    echo "Failed: $((total_count - success_count))"
    echo "Success Rate: ${success_rate}%"
    
    if [ "$success_rate" != "N/A" ]; then
        if (( $(echo "$success_rate >= 80" | bc -l 2>/dev/null || echo "0") )); then
            echo "${GREEN}✅ Success rate meets target (>80%)${NC}"
        else
            echo "${RED}❌ Success rate below target (<80%)${NC}"
        fi
    fi
    
    echo ""
    
    if [ $total_count -gt 0 ]; then
        avg_time=$(echo "scale=2; $total_time / $total_count" | bc 2>/dev/null || echo "N/A")
        echo "Average Response Time: ${avg_time}s"
        
        if [ "$avg_time" != "N/A" ]; then
            if (( $(echo "$avg_time < 5" | bc -l 2>/dev/null || echo "0") )); then
                echo "${GREEN}✅ Average time meets fast-path target (<5s)${NC}"
            elif (( $(echo "$avg_time < 12" | bc -l 2>/dev/null || echo "0") )); then
                echo "${YELLOW}⚠️  Average time meets regular target (<12s)${NC}"
            else
                echo "${RED}❌ Average time exceeds target (>12s)${NC}"
            fi
        fi
    fi
    
    echo ""
    echo "Performance Distribution:"
    echo "  Fast-path (<5s): ${fast_path_count} requests"
    echo "  Regular (<12s): ${regular_count} requests"
    echo "  Slow (>12s): ${slow_count} requests"
fi

echo ""
echo "=========================================="
echo "Recommendations"
echo "=========================================="
echo ""

if [ $success_count -lt $total_count ]; then
    echo "${YELLOW}⚠️  Some requests failed. Check:${NC}"
    echo "  1. Service logs for errors"
    echo "  2. Network connectivity"
    echo "  3. Service health status"
    echo ""
fi

if [ $slow_count -gt 0 ]; then
    echo "${YELLOW}⚠️  Some requests are slow. Check:${NC}"
    echo "  1. Fast-path mode is enabled"
    echo "  2. Parallel processing is working"
    echo "  3. Cache is being utilized"
    echo "  4. Website scraping timeouts"
    echo ""
fi

echo "For detailed analysis:"
echo "1. Check Railway logs: Classification Service → Logs"
echo "2. Look for [FAST-PATH] and [PARALLEL] markers"
echo "3. Monitor cache hit rates"
echo "4. Review timeout patterns"
echo ""

