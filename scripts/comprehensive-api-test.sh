#!/bin/bash

# Comprehensive API Testing Script
# Tests all API endpoints and validates responses are not using placeholder/mock data

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_BASE_URL="${API_BASE_URL:-https://api-gateway-service-production-21fd.up.railway.app}"
TEST_MERCHANT_ID="${TEST_MERCHANT_ID:-biz_thegreen_1762487805256}"
OUTPUT_DIR="${OUTPUT_DIR:-./test-results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${OUTPUT_DIR}/api-test-report-${TIMESTAMP}.json"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Placeholder/mock data patterns to detect
PLACEHOLDER_PATTERNS=(
    "Sample Merchant"
    "Mock"
    "mock"
    "placeholder"
    "TODO"
    "test-"
    "dummy"
    "fake"
    "example"
    "sample"
)

# Test results
declare -a PASSED_TESTS
declare -a FAILED_TESTS
declare -a WARNINGS
TOTAL_TESTS=0
PASSED_COUNT=0
FAILED_COUNT=0
WARNING_COUNT=0

# Initialize JSON report
echo "{" > "$REPORT_FILE"
echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"," >> "$REPORT_FILE"
echo "  \"api_base_url\": \"$API_BASE_URL\"," >> "$REPORT_FILE"
echo "  \"tests\": [" >> "$REPORT_FILE"

# Function to check for placeholder data
check_placeholder_data() {
    local response="$1"
    local endpoint="$2"
    local found_placeholders=()
    
    for pattern in "${PLACEHOLDER_PATTERNS[@]}"; do
        if echo "$response" | grep -qi "$pattern"; then
            found_placeholders+=("$pattern")
        fi
    done
    
    if [ ${#found_placeholders[@]} -gt 0 ]; then
        echo "${found_placeholders[@]}"
    fi
}

# Function to run a test
run_test() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="${5:-200}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    echo "  Endpoint: $method $endpoint"
    
    local curl_cmd="curl -s -w \"\n%{http_code}\" -X $method"
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H \"Content-Type: application/json\" -d '$data'"
    fi
    
    curl_cmd="$curl_cmd \"$API_BASE_URL$endpoint\""
    
    local response=$(eval $curl_cmd)
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    local test_result=""
    local status=""
    local issues=()
    
    # Check HTTP status
    if [ "$http_code" -eq "$expected_status" ]; then
        status="passed"
    else
        status="failed"
        issues+=("Expected status $expected_status, got $http_code")
    fi
    
    # Check if response is valid JSON
    if ! echo "$body" | jq . >/dev/null 2>&1; then
        status="failed"
        issues+=("Response is not valid JSON")
    fi
    
    # Check for placeholder data
    local placeholders=$(check_placeholder_data "$body" "$endpoint")
    if [ -n "$placeholders" ]; then
        if [ "$status" = "passed" ]; then
            status="warning"
            WARNING_COUNT=$((WARNING_COUNT + 1))
        fi
        issues+=("Placeholder data detected: $placeholders")
    fi
    
    # Check if response is empty
    if [ -z "$body" ] || [ "$body" = "null" ] || [ "$body" = "{}" ] || [ "$body" = "[]" ]; then
        if [ "$status" = "passed" ]; then
            status="warning"
            WARNING_COUNT=$((WARNING_COUNT + 1))
        fi
        issues+=("Response is empty or null")
    fi
    
    # Record result
    if [ "$status" = "passed" ]; then
        PASSED_COUNT=$((PASSED_COUNT + 1))
        PASSED_TESTS+=("$test_name")
        echo -e "${GREEN}  ✅ PASSED${NC}"
    elif [ "$status" = "warning" ]; then
        WARNINGS+=("$test_name: ${issues[*]}")
        echo -e "${YELLOW}  ⚠️  WARNING: ${issues[*]}${NC}"
    else
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_TESTS+=("$test_name: ${issues[*]}")
        echo -e "${RED}  ❌ FAILED: ${issues[*]}${NC}"
    fi
    
    # Add to JSON report
    if [ $TOTAL_TESTS -gt 1 ]; then
        echo "," >> "$REPORT_FILE"
    fi
    echo "    {" >> "$REPORT_FILE"
    echo "      \"name\": \"$test_name\"," >> "$REPORT_FILE"
    echo "      \"endpoint\": \"$endpoint\"," >> "$REPORT_FILE"
    echo "      \"method\": \"$method\"," >> "$REPORT_FILE"
    echo "      \"expected_status\": $expected_status," >> "$REPORT_FILE"
    echo "      \"actual_status\": $http_code," >> "$REPORT_FILE"
    echo "      \"status\": \"$status\"," >> "$REPORT_FILE"
    echo "      \"issues\": $(echo "${issues[@]}" | jq -R -s -c 'split(" ")')," >> "$REPORT_FILE"
    echo "      \"response_preview\": $(echo "$body" | head -c 500 | jq -R -s .)" >> "$REPORT_FILE"
    echo "    }" >> "$REPORT_FILE"
    
    echo ""
}

# Start testing
echo -e "${YELLOW}🧪 Comprehensive API Testing${NC}"
echo "API Base URL: $API_BASE_URL"
echo "Test Merchant ID: $TEST_MERCHANT_ID"
echo "Report File: $REPORT_FILE"
echo ""

# Health Checks
echo -e "${BLUE}=== Health Checks ===${NC}"
run_test "API Gateway Health" "GET" "/health" "" 200
run_test "Merchant Service Health" "GET" "/api/v1/merchant/health" "" 200
run_test "Risk Assessment Health" "GET" "/api/v1/risk/health" "" 200
run_test "Classification Health" "GET" "/api/v1/classification/health" "" 200

# Merchant Endpoints
echo -e "${BLUE}=== Merchant Endpoints ===${NC}"
run_test "Get Merchant by ID" "GET" "/api/v1/merchants/$TEST_MERCHANT_ID" "" 200
run_test "List Merchants" "GET" "/api/v1/merchants?page=1&page_size=10" "" 200
run_test "Merchant Analytics" "GET" "/api/v1/merchants/analytics" "" 200

# Risk Assessment Endpoints
echo -e "${BLUE}=== Risk Assessment Endpoints ===${NC}"
run_test "Risk Assessment" "POST" "/api/v1/risk/assess" "{\"merchantId\":\"$TEST_MERCHANT_ID\",\"includeTrendAnalysis\":true}" 200
run_test "Risk Benchmarks (MCC)" "GET" "/api/v1/risk/benchmarks?mcc=5411" "" 200
run_test "Risk Benchmarks (NAICS)" "GET" "/api/v1/risk/benchmarks?naics=541110" "" 200
run_test "Risk Predictions" "GET" "/api/v1/risk/predictions/$TEST_MERCHANT_ID?horizons=3,6,12" "" 200

# Business Intelligence Endpoints
echo -e "${BLUE}=== Business Intelligence Endpoints ===${NC}"
run_test "BI Analysis" "POST" "/api/v1/bi/analyze" "{\"business_name\":\"Test Business\",\"description\":\"Test description\",\"website_url\":\"https://example.com\"}" 200

# Classification Endpoints
echo -e "${BLUE}=== Classification Endpoints ===${NC}"
run_test "Classify Business" "POST" "/api/v1/classify" "{\"business_name\":\"Test Business\",\"description\":\"Test description\",\"website_url\":\"https://example.com\"}" 200

# Close JSON report
echo "  ]," >> "$REPORT_FILE"
echo "  \"summary\": {" >> "$REPORT_FILE"
echo "    \"total_tests\": $TOTAL_TESTS," >> "$REPORT_FILE"
echo "    \"passed\": $PASSED_COUNT," >> "$REPORT_FILE"
echo "    \"failed\": $FAILED_COUNT," >> "$REPORT_FILE"
echo "    \"warnings\": $WARNING_COUNT" >> "$REPORT_FILE"
echo "  }" >> "$REPORT_FILE"
echo "}" >> "$REPORT_FILE"

# Print summary
echo ""
echo -e "${YELLOW}📊 Test Summary${NC}"
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_COUNT${NC}"
echo -e "${RED}Failed: $FAILED_COUNT${NC}"
echo -e "${YELLOW}Warnings: $WARNING_COUNT${NC}"
echo ""

if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${RED}Failed Tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo ""
fi

if [ $WARNING_COUNT -gt 0 ]; then
    echo -e "${YELLOW}Warnings:${NC}"
    for warning in "${WARNINGS[@]}"; do
        echo "  - $warning"
    done
    echo ""
fi

echo "Full report saved to: $REPORT_FILE"
echo ""

# Exit with appropriate code
if [ $FAILED_COUNT -gt 0 ]; then
    exit 1
elif [ $WARNING_COUNT -gt 0 ]; then
    exit 2
else
    exit 0
fi

