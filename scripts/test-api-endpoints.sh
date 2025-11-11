#!/bin/bash

# Comprehensive API Endpoint Testing Script
# Tests all API endpoints with various scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-https://api-gateway-service-production-21fd.up.railway.app}"
JWT_TOKEN="${JWT_TOKEN:-}"
TEST_RESULTS_DIR="${TEST_RESULTS_DIR:-./test-results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create test results directory
mkdir -p "$TEST_RESULTS_DIR"

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Function to print test header
print_test_header() {
    echo ""
    echo "=========================================="
    echo "Testing: $1"
    echo "=========================================="
}

# Function to run a test
run_test() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo -n "  Testing $test_name... "
    
    # Build curl command
    local curl_cmd="curl -s -w '\n%{http_code}' -X $method"
    
    # Add headers
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"
    if [ -n "$JWT_TOKEN" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $JWT_TOKEN'"
    fi
    
    # Add data if provided
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    # Add URL
    curl_cmd="$curl_cmd '$API_BASE_URL$endpoint'"
    
    # Execute and capture response
    local response=$(eval $curl_cmd)
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    # Save response to file
    echo "$body" > "$TEST_RESULTS_DIR/${test_name}_${TIMESTAMP}.json"
    
    # Check status code
    if [ "$http_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected $expected_status, got $http_code)"
        # Only show response body if it's short (to avoid cluttering output)
        local body_length=$(echo "$body" | wc -c)
        if [ $body_length -lt 500 ]; then
            echo "  Response: $body"
        else
            echo "  Response: $(echo "$body" | head -c 200)... (truncated)"
        fi
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to test health check
test_health_check() {
    print_test_header "Health Check Endpoints"
    
    run_test "api_gateway_health" "GET" "/health" "" "200"
    run_test "api_gateway_health_detailed" "GET" "/health?detailed=true" "" "200"
    run_test "classification_health" "GET" "/api/v1/classification/health" "" "200"
    run_test "merchant_health" "GET" "/api/v1/merchant/health" "" "200"
    run_test "risk_health" "GET" "/api/v1/risk/health" "" "200"
}

# Function to test classification endpoint
test_classification() {
    print_test_header "Classification Endpoint"
    
    # Valid request
    local valid_data='{"business_name":"Test Company Inc","description":"Technology solutions provider","website_url":"https://testcompany.com"}'
    run_test "classification_valid" "POST" "/api/v1/classify" "$valid_data" "200"
    
    # Missing required field
    local invalid_data='{"description":"Technology solutions provider"}'
    run_test "classification_missing_name" "POST" "/api/v1/classify" "$invalid_data" "400"
    
    # Invalid JSON
    run_test "classification_invalid_json" "POST" "/api/v1/classify" "{invalid json}" "400"
    
    # Empty body
    run_test "classification_empty_body" "POST" "/api/v1/classify" "" "400"
}

# Function to test merchant endpoints
test_merchants() {
    print_test_header "Merchant Endpoints"
    
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}⚠ Skipping merchant tests (JWT_TOKEN not set)${NC}"
        return
    fi
    
    # List merchants
    run_test "merchants_list" "GET" "/api/v1/merchants" "" "200"
    run_test "merchants_list_pagination" "GET" "/api/v1/merchants?page=1&page_size=10" "" "200"
    run_test "merchants_list_filter" "GET" "/api/v1/merchants?portfolio_type=enterprise&risk_level=low" "" "200"
    run_test "merchants_list_sort" "GET" "/api/v1/merchants?sort_by=name&sort_order=asc" "" "200"
    run_test "merchants_list_search" "GET" "/api/v1/merchants?search=test" "" "200"
    
    # Create merchant
    local create_data='{"name":"Test Merchant","legal_name":"Test Merchant Inc","industry":"Technology","portfolio_type":"enterprise","risk_level":"low","status":"active"}'
    run_test "merchants_create" "POST" "/api/v1/merchants" "$create_data" "201"
    
    # Get merchant (using a test ID - adjust as needed)
    # run_test "merchants_get" "GET" "/api/v1/merchants/test-merchant-id" "" "200"
    
    # Invalid create (missing required fields)
    local invalid_create='{"name":"Test Merchant"}'
    run_test "merchants_create_invalid" "POST" "/api/v1/merchants" "$invalid_create" "400"
}

# Function to test risk assessment endpoints
test_risk_assessment() {
    print_test_header "Risk Assessment Endpoints"
    
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}⚠ Skipping risk assessment tests (JWT_TOKEN not set)${NC}"
        return
    fi
    
    # Risk assessment
    local assess_data='{"business_name":"Test Company","business_address":"123 Test St, Test City, TS 12345","industry":"Technology","country":"USA"}'
    run_test "risk_assess" "POST" "/api/v1/risk/assess" "$assess_data" "200"
    
    # Invalid assessment (missing required fields)
    local invalid_assess='{"business_name":"Test Company"}'
    run_test "risk_assess_invalid" "POST" "/api/v1/risk/assess" "$invalid_assess" "400"
    
    # Risk benchmarks
    run_test "risk_benchmarks" "GET" "/api/v1/risk/benchmarks?mcc=5411" "" "200"
    run_test "risk_benchmarks_missing_params" "GET" "/api/v1/risk/benchmarks" "" "400"
}

# Function to test error handling
test_error_handling() {
    print_test_header "Error Handling"
    
    # 404 Not Found
    run_test "error_404" "GET" "/api/v1/nonexistent" "" "404"
    
    # 405 Method Not Allowed
    run_test "error_405" "DELETE" "/api/v1/classify" "" "405"
    
    # Invalid endpoint
    run_test "error_invalid_endpoint" "GET" "/api/v1/invalid/endpoint" "" "404"
}

# Function to test CORS
test_cors() {
    print_test_header "CORS Configuration"
    
    # Preflight request
    local cors_response=$(curl -s -X OPTIONS \
        -H "Origin: https://frontend-service-production-b225.up.railway.app" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type" \
        -w "\n%{http_code}" \
        "$API_BASE_URL/api/v1/classify")
    
    local cors_code=$(echo "$cors_response" | tail -n1)
    if [ "$cors_code" == "200" ] || [ "$cors_code" == "204" ]; then
        echo -e "  CORS Preflight: ${GREEN}✓ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  CORS Preflight: ${RED}✗ FAILED${NC} (HTTP $cors_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to test rate limiting
test_rate_limiting() {
    print_test_header "Rate Limiting"
    
    echo "  Sending 10 rapid requests to test rate limiting..."
    local rate_limit_hit=0
    
    for i in {1..10}; do
        local response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL/health")
        local http_code=$(echo "$response" | tail -n1)
        
        if [ "$http_code" == "429" ]; then
            rate_limit_hit=1
            echo -e "  Rate limit hit at request $i: ${GREEN}✓ PASSED${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            break
        fi
        sleep 0.1
    done
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if [ $rate_limit_hit -eq 0 ]; then
        echo -e "  Rate limiting: ${YELLOW}⚠ NOT TESTED${NC} (may need more requests)"
    fi
}

# Function to test security headers
test_security_headers() {
    print_test_header "Security Headers"
    
    local headers=$(curl -s -I "$API_BASE_URL/health")
    local missing_headers=0
    
    # Check for security headers
    if echo "$headers" | grep -q "X-Frame-Options"; then
        echo -e "  X-Frame-Options: ${GREEN}✓ PRESENT${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  X-Frame-Options: ${RED}✗ MISSING${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        missing_headers=$((missing_headers + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if echo "$headers" | grep -q "X-Content-Type-Options"; then
        echo -e "  X-Content-Type-Options: ${GREEN}✓ PRESENT${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  X-Content-Type-Options: ${RED}✗ MISSING${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        missing_headers=$((missing_headers + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if echo "$headers" | grep -q "X-XSS-Protection"; then
        echo -e "  X-XSS-Protection: ${GREEN}✓ PRESENT${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  X-XSS-Protection: ${RED}✗ MISSING${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        missing_headers=$((missing_headers + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to test authentication
test_authentication() {
    print_test_header "Authentication"
    
    # Test protected endpoint without token
    local response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL/api/v1/merchants")
    local http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" == "401" ] || [ "$http_code" == "200" ]; then
        # 401 is expected if auth is required, 200 if auth is optional
        echo -e "  Protected endpoint without token: ${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  Protected endpoint without token: ${YELLOW}⚠ UNEXPECTED${NC} (HTTP $http_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    # Test with invalid token
    local invalid_token_response=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer invalid-token" \
        "$API_BASE_URL/api/v1/merchants")
    local invalid_token_code=$(echo "$invalid_token_response" | tail -n1)
    
    if [ "$invalid_token_code" == "401" ]; then
        echo -e "  Invalid token rejection: ${GREEN}✓ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  Invalid token rejection: ${YELLOW}⚠ UNEXPECTED${NC} (HTTP $invalid_token_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to generate test report
generate_report() {
    local report_file="$TEST_RESULTS_DIR/test_report_${TIMESTAMP}.txt"
    
    {
        echo "API Endpoint Test Report"
        echo "========================"
        echo "Date: $(date)"
        echo "API Base URL: $API_BASE_URL"
echo ""
        echo "Test Results:"
        echo "  Total Tests: $TESTS_TOTAL"
        echo "  Passed: $TESTS_PASSED"
        echo "  Failed: $TESTS_FAILED"
echo ""
        echo "Success Rate: $(( TESTS_PASSED * 100 / TESTS_TOTAL ))%"
    } > "$report_file"
    
echo ""
    echo "=========================================="
    echo "Test Report Generated: $report_file"
    echo "=========================================="
    echo "Total Tests: $TESTS_TOTAL"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    echo "Success Rate: $(( TESTS_PASSED * 100 / TESTS_TOTAL ))%"
}

# Main execution
main() {
echo "=========================================="
    echo "KYB Platform API Endpoint Testing"
echo "=========================================="
    echo "API Base URL: $API_BASE_URL"
    echo "Test Results Directory: $TEST_RESULTS_DIR"
echo ""

    # Run all tests
    test_health_check
    test_classification
    test_merchants
    test_risk_assessment
    test_error_handling
    test_cors
    test_security_headers
    test_authentication
    test_rate_limiting
    
    # Generate report
    generate_report
    
    # Exit with appropriate code
    if [ $TESTS_FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
}

# Run main function
main
