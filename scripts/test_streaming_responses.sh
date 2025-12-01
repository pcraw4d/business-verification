#!/bin/bash

# Test Streaming Responses for Classification Service
# OPTIMIZATION #17: Streaming Responses

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
ENDPOINT="${CLASSIFICATION_SERVICE_URL}/v1/classify"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Streaming Responses${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 1: Basic streaming request
echo -e "${YELLOW}Test 1: Basic Streaming Request${NC}"
echo "Endpoint: ${ENDPOINT}?stream=true"
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

# Make streaming request and process NDJSON lines
curl -s -X POST "${ENDPOINT}?stream=true" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" | while IFS= read -r line; do
    if [ -n "$line" ]; then
      echo "$line" | jq -c '.'
      
      # Extract type for status updates
      TYPE=$(echo "$line" | jq -r '.type // empty')
      STATUS=$(echo "$line" | jq -r '.status // empty')
      
      if [ "$TYPE" = "progress" ]; then
        MESSAGE=$(echo "$line" | jq -r '.message // empty')
        STEP=$(echo "$line" | jq -r '.step // empty')
        echo -e "  ${BLUE}→ Progress: ${STATUS}${NC} - ${MESSAGE} (${STEP})"
      elif [ "$TYPE" = "complete" ]; then
        echo -e "  ${GREEN}✓ Complete!${NC}"
        PROCESSING_TIME=$(echo "$line" | jq -r '.processing_time_ms // empty')
        echo -e "  ${GREEN}Processing time: ${PROCESSING_TIME}ms${NC}"
      elif [ "$TYPE" = "error" ]; then
        ERROR_MSG=$(echo "$line" | jq -r '.message // empty')
        echo -e "  ${RED}✗ Error: ${ERROR_MSG}${NC}"
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
  -d "$REQUEST_BODY" > /dev/null
END_TIME=$(date +%s%N)
NON_STREAMING_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
echo "  Time to first byte (full response): ${NON_STREAMING_TIME}ms"
echo ""

echo "Streaming request:"
START_TIME=$(date +%s%N)
FIRST_BYTE_TIME=0
curl -s -X POST "${ENDPOINT}?stream=true" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" | while IFS= read -r line; do
    if [ -n "$line" ] && [ $FIRST_BYTE_TIME -eq 0 ]; then
      END_TIME=$(date +%s%N)
      FIRST_BYTE_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
      echo "  Time to first byte (streaming): ${FIRST_BYTE_TIME}ms"
      echo -e "  ${GREEN}Improvement: $(( NON_STREAMING_TIME - FIRST_BYTE_TIME ))ms faster${NC}"
      break
    fi
  done

echo ""
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 3: Test with different business types
echo -e "${YELLOW}Test 3: Test with Different Business Types${NC}"
echo ""

TEST_BUSINESSES=(
  '{"business_name": "Starbucks", "description": "Coffee shops and beverages", "website_url": "https://starbucks.com"}'
  '{"business_name": "JPMorgan Chase", "description": "Banking and financial services", "website_url": "https://jpmorganchase.com"}'
  '{"business_name": "Mayo Clinic", "description": "Medical center and hospital services", "website_url": "https://mayoclinic.org"}'
)

for business in "${TEST_BUSINESSES[@]}"; do
  BUSINESS_NAME=$(echo "$business" | jq -r '.business_name')
  echo -e "${BLUE}Testing: ${BUSINESS_NAME}${NC}"
  
  curl -s -X POST "${ENDPOINT}?stream=true" \
    -H "Content-Type: application/json" \
    -d "$business" | while IFS= read -r line; do
      if [ -n "$line" ]; then
        TYPE=$(echo "$line" | jq -r '.type // empty')
        if [ "$TYPE" = "progress" ]; then
          STATUS=$(echo "$line" | jq -r '.status // empty')
          echo "  → ${STATUS}"
        elif [ "$TYPE" = "complete" ]; then
          INDUSTRY=$(echo "$line" | jq -r '.data.primary_industry // empty')
          CONFIDENCE=$(echo "$line" | jq -r '.data.confidence_score // empty')
          echo -e "  ${GREEN}✓ Industry: ${INDUSTRY} (confidence: ${CONFIDENCE})${NC}"
        fi
      fi
    done
  echo ""
done

echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Streaming tests complete!${NC}"
echo ""

