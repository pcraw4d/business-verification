#!/bin/bash

# Railway Deployment Verification Script
# Tests all deployed services for health and functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Railway Deployment Verification ===${NC}"
echo ""

# Production URLs from RAILWAY-SERVICE-URLS.md
API_GATEWAY_URL="https://api-gateway-service-production-21fd.up.railway.app"
CLASSIFICATION_URL="https://classification-service-production.up.railway.app"
MERCHANT_URL="https://merchant-service-production.up.railway.app"
RISK_ASSESSMENT_URL="https://risk-assessment-service-production.up.railway.app"
FRONTEND_URL="https://frontend-service-production-b225.up.railway.app"

# Function to test health endpoint
test_health() {
    local service_name=$1
    local url=$2
    local endpoint="${url}/health"
    
    echo -e "${YELLOW}Testing ${service_name}...${NC}"
    response=$(curl -s -w "\n%{http_code}" "${endpoint}" 2>&1 || echo -e "\n000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ ${service_name} is healthy (HTTP ${http_code})${NC}"
        if [ -n "$body" ]; then
            echo "   Response: $(echo "$body" | head -c 100)..."
        fi
        return 0
    else
        echo -e "${RED}❌ ${service_name} returned HTTP ${http_code}${NC}"
        if [ -n "$body" ]; then
            echo "   Response: $body"
        fi
        return 1
    fi
}

# Function to test API endpoint
test_endpoint() {
    local service_name=$1
    local url=$2
    local method=${3:-GET}
    
    echo -e "${YELLOW}Testing ${service_name} ${method} ${url}...${NC}"
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "${url}" 2>&1 || echo -e "\n000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "${url}" 2>&1 || echo -e "\n000")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ ${service_name} endpoint working (HTTP ${http_code})${NC}"
        return 0
    else
        echo -e "${RED}❌ ${service_name} endpoint returned HTTP ${http_code}${NC}"
        return 1
    fi
}

# Track results
PASSED=0
FAILED=0

echo -e "${BLUE}=== 1. Health Check Verification ===${NC}"
echo ""

# Test all service health endpoints
test_health "API Gateway" "$API_GATEWAY_URL" && ((PASSED++)) || ((FAILED++))
echo ""
test_health "Classification Service" "$CLASSIFICATION_URL" && ((PASSED++)) || ((FAILED++))
echo ""
test_health "Merchant Service" "$MERCHANT_URL" && ((PASSED++)) || ((FAILED++))
echo ""
test_health "Risk Assessment Service" "$RISK_ASSESSMENT_URL" && ((PASSED++)) || ((FAILED++))
echo ""
test_health "Frontend Service" "$FRONTEND_URL" && ((PASSED++)) || ((FAILED++))

echo ""
echo -e "${BLUE}=== 2. API Gateway Routing Tests ===${NC}"
echo ""

# Test API Gateway routing
test_endpoint "API Gateway - Classify" "${API_GATEWAY_URL}/api/v1/classify" "GET" && ((PASSED++)) || ((FAILED++))
echo ""
test_endpoint "API Gateway - Merchants" "${API_GATEWAY_URL}/api/v1/merchants" "GET" && ((PASSED++)) || ((FAILED++))

echo ""
echo -e "${BLUE}=== Summary ===${NC}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check the output above.${NC}"
    exit 1
fi

