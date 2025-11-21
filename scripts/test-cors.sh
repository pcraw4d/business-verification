#!/bin/bash

# Test CORS Configuration
# This script verifies that CORS headers are properly configured
# Usage: ./scripts/test-cors.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BACKEND_URL="http://localhost:8080"
FRONTEND_ORIGIN="http://localhost:3000"
TEST_ENDPOINTS=(
    "/health"
    "/api/v1/merchants/merchant_1763614602674531538"
    "/api/v3/merchants/merchant_1763614602674531538"
)

echo -e "${BLUE}🧪 Testing CORS Configuration...${NC}"
echo ""
echo -e "${YELLOW}Backend URL: ${BACKEND_URL}${NC}"
echo -e "${YELLOW}Frontend Origin: ${FRONTEND_ORIGIN}${NC}"
echo ""

# Check if backend is running
if ! curl -s "${BACKEND_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running on ${BACKEND_URL}${NC}"
    echo -e "${YELLOW}   Please start the backend first: ./scripts/restart-backend.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Backend is running${NC}"
echo ""

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Function to test CORS headers
test_cors_headers() {
    local endpoint=$1
    local url="${BACKEND_URL}${endpoint}"
    local test_name="CORS Test: ${endpoint}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Testing: ${test_name}${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Test OPTIONS preflight request
    echo -e "${YELLOW}1. Testing OPTIONS (preflight) request...${NC}"
    local preflight_response=$(curl -s -i -X OPTIONS \
        -H "Origin: ${FRONTEND_ORIGIN}" \
        -H "Access-Control-Request-Method: GET" \
        -H "Access-Control-Request-Headers: Content-Type,Authorization" \
        "${url}" 2>&1)
    
    # Check for CORS headers in preflight
    local allow_origin=$(echo "$preflight_response" | grep -i "access-control-allow-origin" | head -1 || true)
    local allow_methods=$(echo "$preflight_response" | grep -i "access-control-allow-methods" | head -1 || true)
    local allow_headers=$(echo "$preflight_response" | grep -i "access-control-allow-headers" | head -1 || true)
    local allow_credentials=$(echo "$preflight_response" | grep -i "access-control-allow-credentials" | head -1 || true)
    
    # Count Access-Control-Allow-Origin headers (should be exactly 1)
    local origin_count=$(echo "$preflight_response" | grep -ic "access-control-allow-origin" || true)
    
    echo "   Preflight Response Headers:"
    if [ -n "$allow_origin" ]; then
        echo -e "   ${GREEN}✓ Access-Control-Allow-Origin: ${allow_origin#*: }${NC}"
    else
        echo -e "   ${RED}✗ Access-Control-Allow-Origin: MISSING${NC}"
    fi
    
    if [ -n "$allow_methods" ]; then
        echo -e "   ${GREEN}✓ Access-Control-Allow-Methods: ${allow_methods#*: }${NC}"
    else
        echo -e "   ${YELLOW}⚠ Access-Control-Allow-Methods: MISSING${NC}"
    fi
    
    if [ -n "$allow_headers" ]; then
        echo -e "   ${GREEN}✓ Access-Control-Allow-Headers: ${allow_headers#*: }${NC}"
    else
        echo -e "   ${YELLOW}⚠ Access-Control-Allow-Headers: MISSING${NC}"
    fi
    
    if [ -n "$allow_credentials" ]; then
        echo -e "   ${GREEN}✓ Access-Control-Allow-Credentials: ${allow_credentials#*: }${NC}"
    fi
    
    # Check for duplicate headers
    if [ "$origin_count" -gt 1 ]; then
        echo -e "   ${RED}✗ DUPLICATE HEADERS DETECTED: Found ${origin_count} Access-Control-Allow-Origin headers${NC}"
        echo -e "   ${RED}   This will cause CORS errors in the browser!${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    elif [ "$origin_count" -eq 0 ]; then
        echo -e "   ${RED}✗ NO CORS HEADER: Access-Control-Allow-Origin header is missing${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    elif [ "$origin_count" -eq 1 ]; then
        echo -e "   ${GREEN}✓ Single Access-Control-Allow-Origin header (correct)${NC}"
    fi
    
    echo ""
    
    # Test GET request with Origin header
    echo -e "${YELLOW}2. Testing GET request with Origin header...${NC}"
    local get_response=$(curl -s -i -X GET \
        -H "Origin: ${FRONTEND_ORIGIN}" \
        -H "Content-Type: application/json" \
        "${url}" 2>&1)
    
    # Check for CORS headers in GET response
    local get_allow_origin=$(echo "$get_response" | grep -i "access-control-allow-origin" | head -1 || true)
    local get_origin_count=$(echo "$get_response" | grep -ic "access-control-allow-origin" || true)
    
    echo "   GET Response Headers:"
    if [ -n "$get_allow_origin" ]; then
        echo -e "   ${GREEN}✓ Access-Control-Allow-Origin: ${get_allow_origin#*: }${NC}"
    else
        echo -e "   ${RED}✗ Access-Control-Allow-Origin: MISSING${NC}"
    fi
    
    # Check for duplicate headers in GET response
    if [ "$get_origin_count" -gt 1 ]; then
        echo -e "   ${RED}✗ DUPLICATE HEADERS DETECTED: Found ${get_origin_count} Access-Control-Allow-Origin headers${NC}"
        echo -e "   ${RED}   This will cause CORS errors in the browser!${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    elif [ "$get_origin_count" -eq 0 ]; then
        echo -e "   ${RED}✗ NO CORS HEADER: Access-Control-Allow-Origin header is missing${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    elif [ "$get_origin_count" -eq 1 ]; then
        echo -e "   ${GREEN}✓ Single Access-Control-Allow-Origin header (correct)${NC}"
    fi
    
    # Verify the origin value
    if [ -n "$get_allow_origin" ]; then
        local origin_value=$(echo "$get_allow_origin" | cut -d: -f2- | xargs)
        if [[ "$origin_value" == "$FRONTEND_ORIGIN" ]] || [[ "$origin_value" == "*" ]]; then
            echo -e "   ${GREEN}✓ Origin value is correct: ${origin_value}${NC}"
        else
            echo -e "   ${YELLOW}⚠ Origin value is: ${origin_value} (expected: ${FRONTEND_ORIGIN} or *)${NC}"
        fi
    fi
    
    echo ""
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

# Run tests for each endpoint
for endpoint in "${TEST_ENDPOINTS[@]}"; do
    if test_cors_headers "$endpoint"; then
        echo -e "${GREEN}✅ Test passed for ${endpoint}${NC}"
    else
        echo -e "${RED}❌ Test failed for ${endpoint}${NC}"
    fi
    echo ""
done

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 CORS Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All CORS tests passed!${NC}"
    echo -e "${GREEN}   CORS configuration is correct.${NC}"
    echo -e "${GREEN}   No duplicate headers detected.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some CORS tests failed!${NC}"
    echo -e "${YELLOW}   Please check the backend CORS configuration.${NC}"
    exit 1
fi

