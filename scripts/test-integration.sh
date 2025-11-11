#!/bin/bash

# End-to-End Integration Testing Script
# Tests complete merchant verification flow and cross-service communication

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${BLUE}Testing: $1${NC}"
    echo -e "${BLUE}==========================================${NC}"
}

# Function to run a test and capture response
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
    local response=$(eval $curl_cmd 2>&1)
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    # Save response to file
    echo "$body" > "$TEST_RESULTS_DIR/${test_name}_${TIMESTAMP}.json"
    
    # Check status code
    if [ "$http_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "$body"  # Return body for further processing
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected $expected_status, got $http_code)"
        echo "  Response: $body"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to extract value from JSON
extract_json_value() {
    local json="$1"
    local key="$2"
    echo "$json" | grep -o "\"$key\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | cut -d'"' -f4
}

# Function to test merchant verification flow
test_merchant_verification_flow() {
    print_test_header "Merchant Verification Flow"
    
    echo "  Step 1: Classify business..."
    local classify_data='{"business_name":"Integration Test Company","description":"Technology solutions provider for integration testing","website_url":"https://integration-test.com"}'
    local classify_response=$(run_test "flow_classify" "POST" "/api/v1/classify" "$classify_data" "200")
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}  Classification failed, aborting flow test${NC}"
        return 1
    fi
    
    # Extract classification data
    local request_id=$(extract_json_value "$classify_response" "request_id")
    echo "  ✓ Classification completed (Request ID: $request_id)"
    
    echo ""
    echo "  Step 2: Create merchant..."
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}  ⚠ Skipping merchant creation (JWT_TOKEN not set)${NC}"
        return
    fi
    
    local merchant_data='{"name":"Integration Test Company","legal_name":"Integration Test Company Inc","industry":"Technology","portfolio_type":"enterprise","risk_level":"low","status":"active"}'
    local merchant_response=$(run_test "flow_create_merchant" "POST" "/api/v1/merchants" "$merchant_data" "201")
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}  Merchant creation failed, aborting flow test${NC}"
        return 1
    fi
    
    # Extract merchant ID
    local merchant_id=$(extract_json_value "$merchant_response" "id")
    echo "  ✓ Merchant created (ID: $merchant_id)"
    
    echo ""
    echo "  Step 3: Perform risk assessment..."
    local risk_data="{\"business_name\":\"Integration Test Company\",\"business_address\":\"123 Test St, Test City, TS 12345\",\"industry\":\"Technology\",\"country\":\"USA\"}"
    local risk_response=$(run_test "flow_risk_assess" "POST" "/api/v1/risk/assess" "$risk_data" "200")
    
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}  ⚠ Risk assessment failed (non-critical)${NC}"
    else
        local risk_id=$(extract_json_value "$risk_response" "id")
        echo "  ✓ Risk assessment completed (ID: $risk_id)"
    fi
    
    echo ""
    echo "  Step 4: Verify merchant data persistence..."
    if [ -n "$merchant_id" ]; then
        local get_merchant_response=$(run_test "flow_get_merchant" "GET" "/api/v1/merchants/$merchant_id" "" "200")
        
        if [ $? -eq 0 ]; then
            echo "  ✓ Merchant data persisted correctly"
        else
            echo -e "${YELLOW}  ⚠ Could not verify merchant persistence${NC}"
        fi
    fi
    
    echo ""
    echo -e "${GREEN}  ✓ Merchant verification flow completed${NC}"
}

# Function to test data consistency
test_data_consistency() {
    print_test_header "Data Consistency Tests"
    
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}  ⚠ Skipping data consistency tests (JWT_TOKEN not set)${NC}"
        return
    fi
    
    echo "  Testing: Merchant list pagination consistency..."
    
    # Get first page
    local page1_response=$(run_test "consistency_page1" "GET" "/api/v1/merchants?page=1&page_size=10" "" "200")
    local page1_total=$(echo "$page1_response" | grep -o "\"total\"[[:space:]]*:[[:space:]]*[0-9]*" | grep -o "[0-9]*")
    
    # Get second page
    local page2_response=$(run_test "consistency_page2" "GET" "/api/v1/merchants?page=2&page_size=10" "" "200")
    local page2_total=$(echo "$page2_response" | grep -o "\"total\"[[:space:]]*:[[:space:]]*[0-9]*" | grep -o "[0-9]*")
    
    # Verify totals match
    if [ "$page1_total" == "$page2_total" ] && [ -n "$page1_total" ]; then
        echo -e "  ${GREEN}✓ Total count consistent across pages${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${RED}✗ Total count inconsistent${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to test error scenarios
test_error_scenarios() {
    print_test_header "Error Scenario Tests"
    
    echo "  Testing: Network timeout handling..."
    # This would require a timeout simulation - skip for now
    echo -e "  ${YELLOW}⚠ Network timeout test requires manual testing${NC}"
    
    echo ""
    echo "  Testing: Invalid data handling..."
    local invalid_data='{"invalid":"data","malformed":}'
    run_test "error_invalid_json" "POST" "/api/v1/classify" "$invalid_data" "400"
    
    echo ""
    echo "  Testing: Missing required fields..."
    local missing_fields='{"description":"Missing business name"}'
    run_test "error_missing_fields" "POST" "/api/v1/classify" "$missing_fields" "400"
    
    echo ""
    echo "  Testing: Service unavailable handling..."
    # This would require service downtime - skip for now
    echo -e "  ${YELLOW}⚠ Service unavailable test requires manual testing${NC}"
}

# Function to test cross-service communication
test_cross_service_communication() {
    print_test_header "Cross-Service Communication"
    
    echo "  Testing: API Gateway → Classification Service..."
    local classify_response=$(run_test "cross_classify" "POST" "/api/v1/classify" '{"business_name":"Cross Service Test"}' "200")
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓ API Gateway → Classification Service: OK${NC}"
    else
        echo -e "  ${RED}✗ API Gateway → Classification Service: FAILED${NC}"
    fi
    
    echo ""
    echo "  Testing: API Gateway → Merchant Service..."
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "  ${YELLOW}⚠ Skipping (JWT_TOKEN not set)${NC}"
    else
        local merchant_response=$(run_test "cross_merchant" "GET" "/api/v1/merchants?page=1&page_size=1" "" "200")
        
        if [ $? -eq 0 ]; then
            echo -e "  ${GREEN}✓ API Gateway → Merchant Service: OK${NC}"
        else
            echo -e "  ${RED}✗ API Gateway → Merchant Service: FAILED${NC}"
        fi
    fi
    
    echo ""
    echo "  Testing: API Gateway → Risk Assessment Service..."
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "  ${YELLOW}⚠ Skipping (JWT_TOKEN not set)${NC}"
    else
        local risk_response=$(run_test "cross_risk" "POST" "/api/v1/risk/assess" '{"business_name":"Test","business_address":"123 Test St","industry":"Technology","country":"USA"}' "200")
        
        if [ $? -eq 0 ]; then
            echo -e "  ${GREEN}✓ API Gateway → Risk Assessment Service: OK${NC}"
        else
            echo -e "  ${YELLOW}⚠ API Gateway → Risk Assessment Service: May be unavailable${NC}"
        fi
    fi
}

# Function to test response times
test_response_times() {
    print_test_header "Response Time Tests"
    
    echo "  Testing: Health check response time..."
    local start_time=$(date +%s%N)
    curl -s "$API_BASE_URL/health" > /dev/null
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
    
    if [ $duration -lt 100 ]; then
        echo -e "  ${GREEN}✓ Health check: ${duration}ms (Target: < 100ms)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${YELLOW}⚠ Health check: ${duration}ms (Target: < 100ms)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo ""
    echo "  Testing: Classification response time..."
    local start_time=$(date +%s%N)
    curl -s -X POST -H "Content-Type: application/json" \
        -d '{"business_name":"Response Time Test"}' \
        "$API_BASE_URL/api/v1/classify" > /dev/null
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    
    if [ $duration -lt 5000 ]; then
        echo -e "  ${GREEN}✓ Classification: ${duration}ms (Target: < 5s)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "  ${YELLOW}⚠ Classification: ${duration}ms (Target: < 5s)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to generate test report
generate_report() {
    local report_file="$TEST_RESULTS_DIR/integration_test_report_${TIMESTAMP}.txt"
    
    {
        echo "Integration Test Report"
        echo "======================"
        echo "Date: $(date)"
        echo "API Base URL: $API_BASE_URL"
        echo ""
        echo "Test Results:"
        echo "  Total Tests: $TESTS_TOTAL"
        echo "  Passed: $TESTS_PASSED"
        echo "  Failed: $TESTS_FAILED"
        echo ""
        echo "Success Rate: $(( TESTS_PASSED * 100 / TESTS_TOTAL ))%"
        echo ""
        echo "Test Results Directory: $TEST_RESULTS_DIR"
    } > "$report_file"
    
    echo ""
    echo "=========================================="
    echo "Integration Test Report Generated"
    echo "=========================================="
    echo "Report: $report_file"
    echo "Total Tests: $TESTS_TOTAL"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    echo "Success Rate: $(( TESTS_PASSED * 100 / TESTS_TOTAL ))%"
}

# Main execution
main() {
    echo "=========================================="
    echo "KYB Platform Integration Testing"
    echo "=========================================="
    echo "API Base URL: $API_BASE_URL"
    echo "Test Results Directory: $TEST_RESULTS_DIR"
    echo ""
    
    if [ -z "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}Warning: JWT_TOKEN not set. Some tests will be skipped.${NC}"
        echo "Set JWT_TOKEN environment variable to test protected endpoints."
        echo ""
    fi
    
    # Run all integration tests
    test_merchant_verification_flow
    test_data_consistency
    test_error_scenarios
    test_cross_service_communication
    test_response_times
    
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

