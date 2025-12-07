#!/bin/bash

# Test Phase 1 Implementation Locally
# Tests the classification service with Phase 1 enhanced scraping

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
CLASSIFICATION_URL="${CLASSIFICATION_URL:-http://localhost:8081}"
PLAYWRIGHT_URL="${PLAYWRIGHT_URL:-http://localhost:3000}"

echo -e "${BLUE}=== Phase 1 Local Testing ===${NC}\n"

# Check if services are running
echo -e "${BLUE}Checking services...${NC}"

if ! curl -s -f "${PLAYWRIGHT_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Playwright service is not running at ${PLAYWRIGHT_URL}${NC}"
    echo "Start it with: ./scripts/start-local-services.sh"
    exit 1
fi

if ! curl -s -f "${CLASSIFICATION_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Classification service is not running at ${CLASSIFICATION_URL}${NC}"
    echo "Start it with: ./scripts/start-local-services.sh"
    exit 1
fi

echo -e "${GREEN}✅ Services are running${NC}"
echo ""

# Test websites
TEST_URLS=(
    "https://example.com"
    "https://www.microsoft.com"
    "https://www.apple.com"
    "https://www.amazon.com"
    "https://www.starbucks.com"
)

echo -e "${BLUE}Testing classification with Phase 1 scraping...${NC}"
echo ""

SUCCESS_COUNT=0
FAIL_COUNT=0

for url in "${TEST_URLS[@]}"; do
    echo -e "${YELLOW}Testing: ${url}${NC}"
    
    response=$(curl -s -X POST "${CLASSIFICATION_URL}/v1/classify" \
        -H "Content-Type: application/json" \
        -d "{\"business_name\": \"Test Business\", \"website_url\": \"${url}\"}" \
        -w "\n%{http_code}" 2>&1) || true
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        echo -e "  ${GREEN}✅ Success${NC}"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        
        # Extract key info from response
        echo "$body" | jq -r '.confidence_score, .method, .processing_time' 2>/dev/null | while read line; do
            if [ -n "$line" ]; then
                echo "    $line"
            fi
        done || echo "    Response: $(echo "$body" | head -c 200)"
    else
        echo -e "  ${RED}❌ Failed (HTTP $http_code)${NC}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo "    Response: $(echo "$body" | head -c 200)"
    fi
    
    echo ""
    sleep 1
done

echo -e "${BLUE}=== Test Results ===${NC}"
echo -e "  Success: ${GREEN}${SUCCESS_COUNT}${NC}"
echo -e "  Failed: ${RED}${FAIL_COUNT}${NC}"
echo -e "  Total: $((SUCCESS_COUNT + FAIL_COUNT))"
echo ""

# Check logs for Phase 1 markers
echo -e "${BLUE}Checking logs for Phase 1 markers...${NC}"
if command -v docker &> /dev/null; then
    if docker compose version &> /dev/null; then
        DOCKER_COMPOSE="docker compose"
    else
        DOCKER_COMPOSE="docker-compose"
    fi
    
    echo ""
    echo -e "${YELLOW}Recent Phase 1 logs from Classification service:${NC}"
    $DOCKER_COMPOSE -f docker-compose.local.yml logs --tail=50 classification-service | grep -E "\[Phase1\]|Phase 1|KeywordExtraction|Starting.*scraping" | tail -10 || echo "No Phase 1 logs found"
fi

echo ""
echo -e "${GREEN}✅ Testing complete!${NC}"
echo ""
echo -e "${BLUE}To view full logs:${NC}"
echo -e "  ${YELLOW}docker compose -f docker-compose.local.yml logs -f classification-service${NC}"
