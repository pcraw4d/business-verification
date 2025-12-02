#!/bin/bash

# Redis Cache Functionality Test Script
# Tests the classification service cache functionality

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
# Get service URL from environment or prompt user
if [ -z "$CLASSIFICATION_SERVICE_URL" ]; then
    echo "⚠️  CLASSIFICATION_SERVICE_URL not set."
    echo ""
    echo "Please provide your Railway Classification Service URL:"
    echo "  Example: https://classification-service-production.up.railway.app"
    echo ""
    read -p "Enter service URL (or press Enter to use localhost:8081): " user_url
    if [ -z "$user_url" ]; then
        SERVICE_URL="http://localhost:8081"
        echo "Using default: $SERVICE_URL"
    else
        SERVICE_URL="$user_url"
        echo "Using: $SERVICE_URL"
    fi
    echo ""
else
    SERVICE_URL="$CLASSIFICATION_SERVICE_URL"
fi

ENDPOINT="${SERVICE_URL}/classify"

echo "=========================================="
echo "Redis Cache Functionality Test"
echo "=========================================="
echo ""
echo "Service URL: ${SERVICE_URL}"
echo "Endpoint: ${ENDPOINT}"
echo ""

# Test data - Use consistent data for cache testing
TEST_BUSINESS_NAME="Cache Test Company"
TEST_DESCRIPTION="A test business for cache verification - $(date +%s)"
TEST_WEBSITE="https://example.com"

# Function to make request and extract cache header
make_request() {
    local request_num=$1
    local description=$2
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Request #${request_num}: ${description}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Create temp file for response
    local temp_file=$(mktemp)
    local header_file=$(mktemp)
    
    # Make request and capture response with headers
    http_code=$(curl -s -o "$temp_file" -w "%{http_code}" \
        -D "$header_file" \
        -X POST "${ENDPOINT}" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"${TEST_BUSINESS_NAME}\",
            \"description\": \"${TEST_DESCRIPTION}\",
            \"website_url\": \"${TEST_WEBSITE}\"
        }" \
        -w "\n%{time_total}" 2>&1 | tail -1)
    
    # Extract time from response
    time_total=$(tail -1 "$temp_file" 2>/dev/null || echo "0")
    body=$(head -n -1 "$temp_file" 2>/dev/null || cat "$temp_file")
    
    # Extract cache header
    cache_header=$(grep -i "x-cache" "$header_file" 2>/dev/null | tr -d '\r' || echo "X-Cache: (not present)")
    
    # Extract HTTP code (from curl output)
    http_code=$(echo "$http_code" | head -1)
    
    echo "HTTP Status: ${http_code}"
    echo "Response Time: ${time_total}s"
    echo "Cache Header: ${cache_header}"
    
    # Check if response is valid JSON
    if command -v jq &> /dev/null; then
        if echo "$body" | jq . > /dev/null 2>&1; then
            echo "✅ Valid JSON response"
            
            # Extract key fields if available
            if echo "$body" | jq -e '.classification' > /dev/null 2>&1; then
                industry=$(echo "$body" | jq -r '.classification.industry // .classification.primary_industry // "N/A"')
                echo "Industry: ${industry}"
            fi
        else
            echo "⚠️  Response is not valid JSON"
            echo "Response body: ${body:0:200}..."
        fi
    else
        echo "ℹ️  jq not available - skipping JSON validation"
    fi
    
    # Cleanup
    rm -f "$temp_file" "$header_file"
    
    echo ""
    
    # Return time for comparison
    echo "$time_total"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "⚠️  Warning: jq not found. JSON parsing will be limited."
    echo "   Install with: brew install jq (macOS) or apt-get install jq (Linux)"
    echo ""
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo "${RED}Error: curl is required but not found${NC}"
    exit 1
fi

# Test 1: First request (should be cache MISS)
echo "${YELLOW}Test 1: First Request (Expected: Cache MISS)${NC}"
time1=$(make_request 1 "First request - should populate cache")
echo ""

# Wait a moment for cache to be written
echo "Waiting 2 seconds for cache to be written..."
sleep 2
echo ""

# Test 2: Second request (should be cache HIT)
echo "${YELLOW}Test 2: Second Request (Expected: Cache HIT)${NC}"
time2=$(make_request 2 "Second request - should hit cache")
echo ""

# Test 3: Third request (should also be cache HIT)
echo "${YELLOW}Test 3: Third Request (Expected: Cache HIT)${NC}"
time3=$(make_request 3 "Third request - should also hit cache")
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "Request 1 (MISS): ${time1}s"
echo "Request 2 (HIT):  ${time2}s"
echo "Request 3 (HIT):  ${time3}s"
echo ""

# Calculate improvement
if [ -n "$time1" ] && [ -n "$time2" ]; then
    improvement=$(echo "scale=2; (($time1 - $time2) / $time1) * 100" | bc 2>/dev/null || echo "N/A")
    if [ "$improvement" != "N/A" ]; then
        if (( $(echo "$time2 < $time1" | bc -l) )); then
            echo "${GREEN}✅ Cache is working!${NC}"
            echo "   Performance improvement: ${improvement}%"
            echo "   Request 2 was faster than Request 1"
        else
            echo "${YELLOW}⚠️  Cache may not be working optimally${NC}"
            echo "   Request 2 was not faster than Request 1"
        fi
    fi
fi

echo ""
echo "=========================================="
echo "Next Steps:"
echo "=========================================="
echo "1. Check Railway logs for cache operations"
echo "2. Monitor Redis metrics in Railway dashboard"
echo "3. Verify cache hit rate improves over time"
echo ""

