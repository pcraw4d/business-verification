#!/bin/bash

# Comprehensive Streaming Response Test
# Tests streaming endpoint with multiple scenarios

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration - try multiple possible URLs
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-}"
if [ -z "$CLASSIFICATION_SERVICE_URL" ]; then
    # Try to detect the service URL
    if curl -s -f "http://localhost:8081/health" > /dev/null 2>&1; then
        CLASSIFICATION_SERVICE_URL="http://localhost:8081"
        echo -e "${CYAN}Detected local service on port 8081${NC}"
    elif curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
        CLASSIFICATION_SERVICE_URL="http://localhost:8080"
        echo -e "${CYAN}Detected local service on port 8080${NC}"
    elif curl -s -f "https://classification-service-production.up.railway.app/health" > /dev/null 2>&1; then
        CLASSIFICATION_SERVICE_URL="https://classification-service-production.up.railway.app"
        echo -e "${CYAN}Using Railway production service${NC}"
    else
        echo -e "${YELLOW}Could not detect service. Using default: http://localhost:8081${NC}"
        echo -e "${YELLOW}Set CLASSIFICATION_SERVICE_URL environment variable to override${NC}"
        CLASSIFICATION_SERVICE_URL="http://localhost:8081"
    fi
fi

ENDPOINT="${CLASSIFICATION_SERVICE_URL}/v1/classify"
HEALTH_ENDPOINT="${CLASSIFICATION_SERVICE_URL}/health"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Comprehensive Streaming Response Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Service URL: ${CYAN}${CLASSIFICATION_SERVICE_URL}${NC}"
echo -e "Endpoint: ${CYAN}${ENDPOINT}?stream=true${NC}"
echo ""

# Test 0: Health check
echo -e "${YELLOW}Test 0: Health Check${NC}"
if curl -s -f "$HEALTH_ENDPOINT" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Service is healthy${NC}"
else
    echo -e "${RED}✗ Service health check failed${NC}"
    echo -e "${YELLOW}Continuing anyway...${NC}"
fi
echo ""

# Test 1: Basic streaming request
echo -e "${YELLOW}Test 1: Basic Streaming Request${NC}"
echo ""

REQUEST_BODY='{
  "business_name": "Microsoft Corporation",
  "description": "Software development and cloud computing services",
  "website_url": "https://microsoft.com"
}'

echo "Request:"
echo "$REQUEST_BODY" | jq '.'
echo ""
echo -e "${GREEN}Streaming Response (NDJSON):${NC}"
echo ""

START_TIME=$(date +%s%N)
LINE_COUNT=0
FIRST_BYTE_TIME=0

curl -s -X POST "${ENDPOINT}?stream=true" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" 2>&1 | while IFS= read -r line; do
    if [ -n "$line" ]; then
        LINE_COUNT=$((LINE_COUNT + 1))
        
        # Calculate time to first byte
        if [ $FIRST_BYTE_TIME -eq 0 ]; then
            CURRENT_TIME=$(date +%s%N)
            FIRST_BYTE_TIME=$(( (CURRENT_TIME - START_TIME) / 1000000 ))
            echo -e "${CYAN}Time to first byte: ${FIRST_BYTE_TIME}ms${NC}"
        fi
        
        # Parse and display
        TYPE=$(echo "$line" | jq -r '.type // empty' 2>/dev/null || echo "")
        STATUS=$(echo "$line" | jq -r '.status // empty' 2>/dev/null || echo "")
        MESSAGE=$(echo "$line" | jq -r '.message // empty' 2>/dev/null || echo "")
        
        if [ "$TYPE" = "progress" ]; then
            STEP=$(echo "$line" | jq -r '.step // empty' 2>/dev/null || echo "")
            echo -e "  ${BLUE}[${STATUS}]${NC} ${MESSAGE}"
            if [ -n "$STEP" ]; then
                echo -e "    ${CYAN}Step: ${STEP}${NC}"
            fi
            
            # Show additional data if available
            if echo "$line" | jq -e '.primary_industry' > /dev/null 2>&1; then
                INDUSTRY=$(echo "$line" | jq -r '.primary_industry')
                CONFIDENCE=$(echo "$line" | jq -r '.confidence')
                echo -e "    ${GREEN}Industry: ${INDUSTRY} (${CONFIDENCE})${NC}"
            fi
        elif [ "$TYPE" = "complete" ]; then
            END_TIME=$(date +%s%N)
            TOTAL_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
            echo -e "  ${GREEN}✓ Complete!${NC}"
            echo -e "    ${CYAN}Total time: ${TOTAL_TIME}ms${NC}"
            echo -e "    ${CYAN}Total lines: ${LINE_COUNT}${NC}"
        elif [ "$TYPE" = "error" ]; then
            ERROR_MSG=$(echo "$line" | jq -r '.message // empty' 2>/dev/null || echo "")
            echo -e "  ${RED}✗ Error: ${ERROR_MSG}${NC}"
        else
            # Raw JSON output for unknown types
            echo "$line" | jq -c '.' 2>/dev/null || echo "$line"
        fi
        echo ""
    fi
  done

echo ""
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 2: Compare streaming vs non-streaming
echo -e "${YELLOW}Test 2: Compare Streaming vs Non-Streaming${NC}"
echo ""

echo "Non-streaming request (traditional):"
START_TIME=$(date +%s%N)
curl -s -X POST "${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" > /tmp/non_streaming_response.json 2>&1
END_TIME=$(date +%s%N)
NON_STREAMING_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
echo -e "  ${CYAN}Total time: ${NON_STREAMING_TIME}ms${NC}"

if [ -f /tmp/non_streaming_response.json ]; then
    if jq -e '.success' /tmp/non_streaming_response.json > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓ Response received${NC}"
    else
        echo -e "  ${RED}✗ Invalid response${NC}"
        cat /tmp/non_streaming_response.json | head -5
    fi
fi
echo ""

echo "Streaming request:"
START_TIME=$(date +%s%N)
FIRST_BYTE_TIME=0
curl -s -X POST "${ENDPOINT}?stream=true" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" 2>&1 | while IFS= read -r line; do
    if [ -n "$line" ] && [ $FIRST_BYTE_TIME -eq 0 ]; then
        END_TIME=$(date +%s%N)
        FIRST_BYTE_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
        echo -e "  ${CYAN}Time to first byte: ${FIRST_BYTE_TIME}ms${NC}"
        if [ $NON_STREAMING_TIME -gt 0 ]; then
            IMPROVEMENT=$(( NON_STREAMING_TIME - FIRST_BYTE_TIME ))
            if [ $IMPROVEMENT -gt 0 ]; then
                PERCENTAGE=$(( (IMPROVEMENT * 100) / NON_STREAMING_TIME ))
                echo -e "  ${GREEN}Improvement: ${IMPROVEMENT}ms faster (${PERCENTAGE}% better perceived latency)${NC}"
            fi
        fi
        break
    fi
  done

echo ""
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 3: Test with different business types
echo -e "${YELLOW}Test 3: Test with Different Business Types${NC}"
echo ""

declare -a TEST_BUSINESSES=(
    '{"business_name": "Starbucks", "description": "Coffee shops and beverages", "website_url": "https://starbucks.com"}'
    '{"business_name": "JPMorgan Chase", "description": "Banking and financial services", "website_url": "https://jpmorganchase.com"}'
    '{"business_name": "Mayo Clinic", "description": "Medical center and hospital services", "website_url": "https://mayoclinic.org"}'
    '{"business_name": "Amazon", "description": "E-commerce and retail services", "website_url": "https://amazon.com"}'
    '{"business_name": "Apple Inc", "description": "Consumer electronics and software", "website_url": "https://apple.com"}'
)

for business in "${TEST_BUSINESSES[@]}"; do
    BUSINESS_NAME=$(echo "$business" | jq -r '.business_name')
    echo -e "${BLUE}Testing: ${BUSINESS_NAME}${NC}"
    
    START_TIME=$(date +%s%N)
    curl -s -X POST "${ENDPOINT}?stream=true" \
      -H "Content-Type: application/json" \
      -d "$business" 2>&1 | while IFS= read -r line; do
        if [ -n "$line" ]; then
            TYPE=$(echo "$line" | jq -r '.type // empty' 2>/dev/null || echo "")
            if [ "$TYPE" = "progress" ]; then
                STATUS=$(echo "$line" | jq -r '.status // empty' 2>/dev/null || echo "")
                echo "  → ${STATUS}"
            elif [ "$TYPE" = "complete" ]; then
                END_TIME=$(date +%s%N)
                PROCESSING_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
                INDUSTRY=$(echo "$line" | jq -r '.data.primary_industry // empty' 2>/dev/null || echo "")
                CONFIDENCE=$(echo "$line" | jq -r '.data.confidence_score // empty' 2>/dev/null || echo "")
                echo -e "  ${GREEN}✓ Industry: ${INDUSTRY} (confidence: ${CONFIDENCE})${NC}"
                echo -e "  ${CYAN}Processing time: ${PROCESSING_TIME}ms${NC}"
            fi
        fi
      done
    echo ""
done

echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}All streaming tests complete!${NC}"
echo ""
echo -e "${CYAN}Summary:${NC}"
echo "  - Streaming endpoint: ${ENDPOINT}?stream=true"
echo "  - Format: NDJSON (newline-delimited JSON)"
echo "  - Content-Type: application/x-ndjson"
echo "  - Progress updates sent as steps complete"
echo ""

