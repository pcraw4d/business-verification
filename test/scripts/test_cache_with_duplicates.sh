#!/bin/bash

# Targeted Cache Test with Duplicate Requests
# Tests if cache is working by making the same request twice

set -e

CLASSIFICATION_API_URL="https://classification-service-production.up.railway.app"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Targeted Cache Test - Duplicate Requests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${CYAN}$CLASSIFICATION_API_URL${NC}"
echo ""

# Test request data
TEST_DATA='{
  "business_name": "Cache Test Company",
  "description": "Testing cache functionality with duplicate requests",
  "website_url": "https://cachetest.example.com"
}'

echo -e "${BLUE}Test Request Data:${NC}"
echo "$TEST_DATA" | jq '.' 2>/dev/null || echo "$TEST_DATA"
echo ""

# First request
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Request #1: Initial Request (should be cache miss)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

FIRST_START=$(date +%s.%N)
FIRST_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" \
  --max-time 60 \
  -w "\n%{http_code}" 2>&1)

FIRST_END=$(date +%s.%N)
FIRST_DURATION=$(echo "$FIRST_END - $FIRST_START" | bc)

# Extract HTTP status code (last line)
FIRST_HTTP_CODE=$(echo "$FIRST_RESPONSE" | tail -n1)
FIRST_BODY=$(echo "$FIRST_RESPONSE" | sed '$d')

echo -e "HTTP Status: ${CYAN}$FIRST_HTTP_CODE${NC}"
echo -e "Processing Time: ${CYAN}${FIRST_DURATION}s${NC}"
echo ""

# Parse first response
if command -v jq &> /dev/null; then
    FIRST_FROM_CACHE=$(echo "$FIRST_BODY" | jq -r '.from_cache // false' 2>/dev/null || echo "false")
    FIRST_REQUEST_ID=$(echo "$FIRST_BODY" | jq -r '.request_id // "N/A"' 2>/dev/null || echo "N/A")
    FIRST_INDUSTRY=$(echo "$FIRST_BODY" | jq -r '.primary_industry // "N/A"' 2>/dev/null || echo "N/A")
    FIRST_SUCCESS=$(echo "$FIRST_BODY" | jq -r '.success // false' 2>/dev/null || echo "false")
    
    echo -e "Request ID: ${CYAN}$FIRST_REQUEST_ID${NC}"
    echo -e "From Cache: ${CYAN}$FIRST_FROM_CACHE${NC}"
    echo -e "Success: ${CYAN}$FIRST_SUCCESS${NC}"
    echo -e "Primary Industry: ${CYAN}$FIRST_INDUSTRY${NC}"
    
    if [ "$FIRST_FROM_CACHE" = "true" ]; then
        echo -e "${YELLOW}⚠️  First request hit cache (unexpected)${NC}"
    else
        echo -e "${GREEN}✅ First request is cache miss (expected)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  jq not available, showing raw response:${NC}"
    echo "$FIRST_BODY" | head -20
fi

echo ""
echo -e "${YELLOW}Waiting 2 seconds before second request...${NC}"
sleep 2
echo ""

# Second request (should hit cache)
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Request #2: Duplicate Request (should be cache hit)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

SECOND_START=$(date +%s.%N)
SECOND_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" \
  --max-time 60 \
  -w "\n%{http_code}" 2>&1)

SECOND_END=$(date +%s.%N)
SECOND_DURATION=$(echo "$SECOND_END - $SECOND_START" | bc)

# Extract HTTP status code (last line)
SECOND_HTTP_CODE=$(echo "$SECOND_RESPONSE" | tail -n1)
SECOND_BODY=$(echo "$SECOND_RESPONSE" | sed '$d')

echo -e "HTTP Status: ${CYAN}$SECOND_HTTP_CODE${NC}"
echo -e "Processing Time: ${CYAN}${SECOND_DURATION}s${NC}"
echo ""

# Parse second response
if command -v jq &> /dev/null; then
    SECOND_FROM_CACHE=$(echo "$SECOND_BODY" | jq -r '.from_cache // false' 2>/dev/null || echo "false")
    SECOND_REQUEST_ID=$(echo "$SECOND_BODY" | jq -r '.request_id // "N/A"' 2>/dev/null || echo "N/A")
    SECOND_INDUSTRY=$(echo "$SECOND_BODY" | jq -r '.primary_industry // "N/A"' 2>/dev/null || echo "N/A")
    SECOND_SUCCESS=$(echo "$SECOND_BODY" | jq -r '.success // false' 2>/dev/null || echo "false")
    
    echo -e "Request ID: ${CYAN}$SECOND_REQUEST_ID${NC}"
    echo -e "From Cache: ${CYAN}$SECOND_FROM_CACHE${NC}"
    echo -e "Success: ${CYAN}$SECOND_SUCCESS${NC}"
    echo -e "Primary Industry: ${CYAN}$SECOND_INDUSTRY${NC}"
    
    if [ "$SECOND_FROM_CACHE" = "true" ]; then
        echo -e "${GREEN}✅ Second request hit cache (expected)${NC}"
        CACHE_WORKING=true
    else
        echo -e "${RED}❌ Second request did NOT hit cache (unexpected)${NC}"
        CACHE_WORKING=false
    fi
    
    # Compare processing times
    SPEEDUP=$(echo "scale=2; $FIRST_DURATION / $SECOND_DURATION" | bc 2>/dev/null || echo "N/A")
    if [ "$CACHE_WORKING" = "true" ]; then
        echo -e "Speed Improvement: ${GREEN}${SPEEDUP}x faster${NC}"
    fi
    
    # Check if results match
    if [ "$FIRST_INDUSTRY" = "$SECOND_INDUSTRY" ] && [ "$FIRST_INDUSTRY" != "N/A" ]; then
        echo -e "${GREEN}✅ Results match between requests${NC}"
    else
        echo -e "${YELLOW}⚠️  Results differ between requests${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  jq not available, showing raw response:${NC}"
    echo "$SECOND_BODY" | head -20
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ "$CACHE_WORKING" = "true" ]; then
    echo -e "${GREEN}✅ CACHE IS WORKING${NC}"
    echo -e "  - First request: Cache miss (${FIRST_DURATION}s)"
    echo -e "  - Second request: Cache hit (${SECOND_DURATION}s)"
    echo -e "  - Speed improvement: ${SPEEDUP}x"
else
    echo -e "${RED}❌ CACHE IS NOT WORKING${NC}"
    echo -e "  - First request: Cache miss (${FIRST_DURATION}s)"
    echo -e "  - Second request: Cache miss (${SECOND_DURATION}s)"
    echo -e "  - Possible causes:"
    echo -e "    1. Redis connection issues"
    echo -e "    2. Cache disabled in configuration"
    echo -e "    3. Cache keys not matching"
    echo -e "    4. Cache TTL expired"
fi

echo ""
echo -e "${BLUE}========================================${NC}"

