#!/bin/bash

# Phase 1 Scraping Testing Script
# Tests classification service with various websites and collects metrics

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLASSIFICATION_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"
PLAYWRIGHT_URL="${PLAYWRIGHT_SERVICE_URL:-https://playwright-service-production-b21a.up.railway.app}"

# Test websites - diverse types
declare -a TEST_WEBSITES=(
    "https://example.com|Example Domain|Simple static HTML"
    "https://www.wikipedia.org|Wikipedia|Content-heavy, bot detection"
    "https://react.dev|React|JavaScript-heavy SPA"
    "https://github.com|GitHub|Modern web app"
    "https://stackoverflow.com|Stack Overflow|Content-rich"
    "https://www.apple.com|Apple|Corporate site"
    "https://www.microsoft.com|Microsoft|Enterprise site"
)

# Metrics tracking
TOTAL_TESTS=0
SUCCESSFUL_TESTS=0
FAILED_TESTS=0
QUALITY_SCORES=()
WORD_COUNTS=()
STRATEGIES=()

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Phase 1 Scraping Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check services
echo -e "${YELLOW}Checking services...${NC}"

# Check Playwright service
if curl -s -f -m 5 "$PLAYWRIGHT_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Playwright service: OK${NC}"
else
    echo -e "${RED}❌ Playwright service: FAILED${NC}"
    exit 1
fi

# Check Classification service
if curl -s -f -m 5 "$CLASSIFICATION_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Classification service: OK${NC}"
else
    echo -e "${RED}❌ Classification service: FAILED${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Testing with ${#TEST_WEBSITES[@]} websites...${NC}"
echo ""

# Test each website
for test_case in "${TEST_WEBSITES[@]}"; do
    IFS='|' read -r url name description <<< "$test_case"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}Test $TOTAL_TESTS: $name${NC}"
    echo -e "   URL: $url"
    echo -e "   Type: $description"
    
    # Make classification request
    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 30 \
        -X POST "$CLASSIFICATION_URL/v1/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"$name\",
            \"website_url\": \"$url\"
        }" 2>&1)
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        SUCCESSFUL_TESTS=$((SUCCESSFUL_TESTS + 1))
        echo -e "${GREEN}   ✅ Success (HTTP $HTTP_CODE)${NC}"
        
        # Extract metrics from response (if available in response)
        # Note: Quality scores and strategy info are in logs, not response
        echo "$BODY" | jq -r '.confidence_score // "N/A"' | while read score; do
            if [ "$score" != "N/A" ]; then
                echo -e "   Confidence: $score"
            fi
        done
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}   ❌ Failed (HTTP $HTTP_CODE)${NC}"
        echo "$BODY" | jq -r '.error // .message // "Unknown error"' 2>/dev/null || echo "   Error: $BODY"
    fi
    
    echo ""
    sleep 1  # Rate limiting
done

# Calculate metrics
SUCCESS_RATE=$(awk "BEGIN {printf \"%.1f\", ($SUCCESSFUL_TESTS / $TOTAL_TESTS) * 100}")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Results Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Successful: $SUCCESSFUL_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo -e "Success Rate: ${SUCCESS_RATE}%"
echo ""

# Success criteria
echo -e "${BLUE}Success Criteria:${NC}"
if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then
    echo -e "${GREEN}✅ Scrape success rate: ${SUCCESS_RATE}% (Target: ≥95%)${NC}"
else
    echo -e "${RED}❌ Scrape success rate: ${SUCCESS_RATE}% (Target: ≥95%)${NC}"
fi

echo ""
echo -e "${YELLOW}Note: Quality scores and strategy usage are logged in Railway.${NC}"
echo -e "${YELLOW}Check Railway logs for detailed metrics:${NC}"
echo -e "   - Strategy used (simple_http, browser_headers, playwright)"
echo -e "   - Quality scores (target: ≥0.7)"
echo -e "   - Word counts (target: ≥200)"
echo ""

