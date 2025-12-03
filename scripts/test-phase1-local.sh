#!/bin/bash

# Phase 1 Local Testing Script
# Tests the Phase 1 enhanced scraper locally

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Phase 1 Local Testing ===${NC}\n"

# Check if service is running
SERVICE_URL="${SERVICE_URL:-http://localhost:8081}"
HEALTH_URL="${SERVICE_URL}/health"

echo -e "${YELLOW}Checking service health at ${HEALTH_URL}...${NC}"
if curl -s -f "${HEALTH_URL}" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Service is running${NC}\n"
else
    echo -e "${RED}❌ Service is not running at ${SERVICE_URL}${NC}"
    echo -e "${YELLOW}Please start the classification service first:${NC}"
    echo -e "  ${BLUE}Option 1:${NC} Run in a separate terminal:"
    echo -e "    ./scripts/start-classification-service.sh"
    echo -e ""
    echo -e "  ${BLUE}Option 2:${NC} Manual start:"
    echo -e "    cd services/classification-service"
    echo -e "    export \$(cat ../../.env | grep -v '^#' | xargs)"
    echo -e "    export SUPABASE_ANON_KEY=\"\${SUPABASE_API_KEY}\""
    echo -e "    export PORT=\"8081\""
    echo -e "    export LOG_LEVEL=\"debug\""
    echo -e "    go run cmd/main.go"
    echo ""
    exit 1
fi

# Test websites
declare -a TEST_URLS=(
    "https://example.com"
    "https://www.apple.com"
    "https://www.microsoft.com"
    "https://react.dev"
    "https://www.starbucks.com"
)

echo -e "${BLUE}Testing Phase 1 Enhanced Scraper${NC}\n"
echo -e "${YELLOW}Looking for Phase 1 logs in service output...${NC}\n"

for url in "${TEST_URLS[@]}"; do
    echo -e "${BLUE}Testing: ${url}${NC}"
    
    response=$(curl -s -X POST "${SERVICE_URL}/v1/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"Test Company\",
            \"website_url\": \"${url}\"
        }")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Request successful${NC}"
        
        # Extract key fields
        success=$(echo "$response" | grep -o '"success":[^,]*' | cut -d: -f2)
        industry=$(echo "$response" | grep -o '"primary_industry":"[^"]*"' | cut -d'"' -f4)
        
        echo -e "  Success: ${success}"
        echo -e "  Industry: ${industry}"
    else
        echo -e "${RED}❌ Request failed${NC}"
    fi
    
    echo ""
    sleep 1
done

echo -e "${BLUE}=== Test Complete ===${NC}\n"
echo -e "${YELLOW}Check the service logs for Phase 1 markers:${NC}"
echo -e "  - [Phase1] [KeywordExtraction]"
echo -e "  - Strategy usage (simple_http, browser_headers, playwright)"
echo -e "  - Quality scores"
echo -e "  - Word counts"

