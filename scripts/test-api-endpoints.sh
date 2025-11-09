#!/bin/bash

# API Endpoint Testing Script
# Tests all production API endpoints for functionality

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

API_BASE="https://api-gateway-service-production-21fd.up.railway.app"
FRONTEND_BASE="https://frontend-service-production-b225.up.railway.app"

PASSED=0
FAILED=0

test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=${4:-200}
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" --max-time 10 "$endpoint" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time 30 \
            "$endpoint" 2>&1)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ PASS:${NC} $method $endpoint (HTTP $http_code)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}❌ FAIL:${NC} $method $endpoint (HTTP $http_code, expected $expected_code)"
        echo "   Response: $(echo "$body" | head -3)"
        ((FAILED++))
        return 1
    fi
}

echo "🧪 Testing API Endpoints..."
echo ""

# Health Checks
echo "📋 Health Checks"
echo "----------------"
test_endpoint "GET" "$API_BASE/health"
test_endpoint "GET" "$FRONTEND_BASE/health"
echo ""

# Classification API
echo "📋 Classification API"
echo "---------------------"
test_endpoint "POST" "$API_BASE/api/v1/classify" \
    '{"business_name":"Test Company","description":"Test","website_url":"https://example.com"}' \
    200
echo ""

# Frontend Pages
echo "📋 Frontend Pages"
echo "-----------------"
test_endpoint "GET" "$FRONTEND_BASE/add-merchant"
test_endpoint "GET" "$FRONTEND_BASE/merchant-details"
test_endpoint "GET" "$FRONTEND_BASE/merchant-portfolio"
echo ""

# Summary
echo "=========================================="
echo "📊 API Test Summary"
echo "=========================================="
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi

