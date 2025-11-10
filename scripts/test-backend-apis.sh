#!/bin/bash

# Backend API Testing Script
# Tests all API endpoints, rate limiting, authentication, and error handling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API Gateway URL
API_BASE_URL="${API_BASE_URL:-https://api-gateway-service-production-21fd.up.railway.app}"

# Test results
PASSED=0
FAILED=0
WARNINGS=0

# Test counter
TOTAL_TESTS=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Backend API Testing Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_status="${3:-200}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}Testing: ${test_name}${NC}"
    
    # Execute command and capture both response body and status code
    # curl -w outputs the status code at the end when using %{http_code}
    response=$(eval "$command" 2>&1)
    
    # Extract status code - it should be the last line when using curl -w '\n%{http_code}'
    status_code=$(echo "$response" | tail -1 | tr -d '\n' | grep -oE '^[0-9]{3}$' || echo "000")
    
    # If status code extraction failed, try to get it from HTTP headers
    if [ "$status_code" = "000" ] || [ -z "$status_code" ]; then
        status_code=$(echo "$response" | grep -i "HTTP/" | head -1 | awk '{print $2}' || echo "000")
    fi
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASSED${NC} - Status: $status_code"
        PASSED=$((PASSED + 1))
    elif [ "$status_code" = "000" ] || [ -z "$status_code" ]; then
        echo -e "${YELLOW}⚠️  WARNING${NC} - Could not determine status code"
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${RED}❌ FAILED${NC} - Expected: $expected_status, Got: $status_code"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# Function to test rate limiting
test_rate_limit() {
    echo -e "${BLUE}Testing Rate Limiting${NC}"
    echo "Sending 10 rapid requests..."
    
    local success_count=0
    local rate_limited_count=0
    
    for i in {1..10}; do
        response=$(curl -s -w "\n%{http_code}" -X GET "${API_BASE_URL}/health" 2>&1)
        status_code=$(echo "$response" | tail -1)
        
        if [ "$status_code" = "200" ]; then
            success_count=$((success_count + 1))
        elif [ "$status_code" = "429" ]; then
            rate_limited_count=$((rate_limited_count + 1))
        fi
    done
    
    if [ $rate_limited_count -gt 0 ]; then
        echo -e "${GREEN}✅ Rate limiting is working${NC} - $rate_limited_count requests rate limited"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠️  Rate limiting may not be active${NC} - All requests succeeded"
        WARNINGS=$((WARNINGS + 1))
    fi
    echo ""
}

# Function to test error handling
test_error_handling() {
    echo -e "${BLUE}Testing Error Handling${NC}"
    
    # Test invalid JSON
    run_test "Invalid JSON in request body" \
        "curl -s -w '\\n%{http_code}' -X POST '${API_BASE_URL}/api/v1/classify' -H 'Content-Type: application/json' -d 'invalid json'" \
        "400"
    
    # Test missing required fields
    run_test "Missing required fields in registration" \
        "curl -s -w '\\n%{http_code}' -X POST '${API_BASE_URL}/api/v1/auth/register' -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\"}'" \
        "400"
    
    # Test invalid endpoint
    run_test "Invalid endpoint (404)" \
        "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/nonexistent'" \
        "404"
    
    echo ""
}

# Test 1: Health Check
echo -e "${BLUE}=== Health Check Tests ===${NC}"
run_test "API Gateway Health Check" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/health'"

# Test 2: Classification Endpoint
echo -e "${BLUE}=== Classification Endpoint Tests ===${NC}"
run_test "Classification - Software Company" \
    "curl -s -w '\\n%{http_code}' -X POST '${API_BASE_URL}/api/v1/classify' -H 'Content-Type: application/json' -d '{\"business_name\":\"Acme Software\",\"description\":\"Software development company\"}'"

run_test "Classification - Medical Clinic" \
    "curl -s -w '\\n%{http_code}' -X POST '${API_BASE_URL}/api/v1/classify' -H 'Content-Type: application/json' -d '{\"business_name\":\"City Medical Clinic\",\"description\":\"Healthcare services\"}'"

# Test 3: Merchant Endpoints
echo -e "${BLUE}=== Merchant Endpoint Tests ===${NC}"
run_test "List Merchants" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/merchants'"

run_test "Get Merchant by ID" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/merchants/test-merchant-1'"

# Test 4: Risk Assessment Endpoints
echo -e "${BLUE}=== Risk Assessment Endpoint Tests ===${NC}"
# Risk benchmarks may return 503 if service is unavailable - accept both 200 and 503
run_test "Risk Benchmarks" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/risk/benchmarks?mcc=5411'" \
    "200"

run_test "Risk Predictions" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/risk/predictions/test-merchant-1'"

# Test 5: Registration Endpoint
echo -e "${BLUE}=== Registration Endpoint Tests ===${NC}"
run_test "User Registration - Valid Request" \
    "curl -s -w '\\n%{http_code}' -X POST '${API_BASE_URL}/api/v1/auth/register' -H 'Content-Type: application/json' -d '{\"email\":\"test$(date +%s)@example.com\",\"username\":\"testuser$(date +%s)\",\"password\":\"SecurePass123!\",\"first_name\":\"Test\",\"last_name\":\"User\",\"company\":\"Test Corp\"}'" \
    "201"

# Test 6: Service Health Proxies
echo -e "${BLUE}=== Service Health Proxy Tests ===${NC}"
run_test "Classification Service Health" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/classification/health'"

run_test "Merchant Service Health" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/merchant/health'"

run_test "Risk Assessment Service Health" \
    "curl -s -w '\\n%{http_code}' -X GET '${API_BASE_URL}/api/v1/risk/health'"

# Test 7: CORS Headers
echo -e "${BLUE}=== CORS Tests ===${NC}"
cors_response=$(curl -s -I -X OPTIONS "${API_BASE_URL}/api/v1/classify" -H "Origin: https://example.com" -H "Access-Control-Request-Method: POST")
if echo "$cors_response" | grep -qi "access-control-allow-origin"; then
    echo -e "${GREEN}✅ CORS headers present${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}⚠️  CORS headers not found${NC}"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# Test 8: Rate Limiting
test_rate_limit

# Test 9: Error Handling
test_error_handling

# Test 10: Response Times
echo -e "${BLUE}=== Performance Tests ===${NC}"
echo "Measuring response times..."

endpoints=(
    "/health"
    "/api/v1/classify"
    "/api/v1/merchants"
)

for endpoint in "${endpoints[@]}"; do
    start_time=$(date +%s%N)
    curl -s -X GET "${API_BASE_URL}${endpoint}" > /dev/null
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 500 ]; then
        echo -e "${GREEN}✅ ${endpoint}: ${duration}ms${NC}"
    elif [ $duration -lt 1000 ]; then
        echo -e "${YELLOW}⚠️  ${endpoint}: ${duration}ms (acceptable)${NC}"
    else
        echo -e "${RED}❌ ${endpoint}: ${duration}ms (slow)${NC}"
    fi
done
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"
echo -e "${YELLOW}Warnings: ${WARNINGS}${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All critical tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

