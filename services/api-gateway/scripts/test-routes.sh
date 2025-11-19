#!/bin/bash

# Route Testing Script for API Gateway
# Tests all routes through the API Gateway to verify routing, path transformations, and error handling

# Don't exit on error - we want to test all routes even if some fail
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_GATEWAY_URL="${API_GATEWAY_URL:-http://localhost:8080}"
TEST_MERCHANT_ID="${TEST_MERCHANT_ID:-merchant-123}"
TIMEOUT=10

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Test result tracking
declare -a FAILED_TESTS=()

# Function to print test result
print_result() {
    local status=$1
    local test_name=$2
    local message=$3
    
    if [ "$status" == "PASS" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((PASSED++))
    elif [ "$status" == "FAIL" ]; then
        echo -e "${RED}✗${NC} $test_name: $message"
        ((FAILED++))
        FAILED_TESTS+=("$test_name: $message")
    elif [ "$status" == "SKIP" ]; then
        echo -e "${YELLOW}⊘${NC} $test_name: $message"
        ((SKIPPED++))
    fi
}

# Function to test a route
test_route() {
    local method=$1
    local path=$2
    local expected_status=$3
    local test_name=$4
    local query_params=$5
    local body=$6
    
    # Build URL
    local url="${API_GATEWAY_URL}${path}"
    if [ -n "$query_params" ]; then
        url="${url}?${query_params}"
    fi
    
    # Make request
    local response
    local curl_exit_code=0
    if [ -n "$body" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$body" \
            --max-time $TIMEOUT \
            "$url" 2>&1)
        curl_exit_code=$?
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            --max-time $TIMEOUT \
            "$url" 2>&1)
        curl_exit_code=$?
    fi
    
    # Check if curl failed (service not running)
    if echo "$response" | grep -q "Connection refused\|Failed to connect"; then
        print_result "SKIP" "$test_name" "API Gateway not running at $API_GATEWAY_URL"
        return
    fi
    
    # Extract status code (last line)
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    # Check status code
    if [ "$status_code" == "$expected_status" ]; then
        print_result "PASS" "$test_name"
    else
        print_result "FAIL" "$test_name" "Expected $expected_status, got $status_code"
    fi
}

# Function to test CORS headers
test_cors() {
    local path=$1
    local test_name=$2
    
    local url="${API_GATEWAY_URL}${path}"
    
    # Test OPTIONS request (preflight)
    local headers=$(curl -s -I -X OPTIONS \
        -H "Origin: https://frontend-service-production-b225.up.railway.app" \
        -H "Access-Control-Request-Method: GET" \
        --max-time $TIMEOUT \
        "$url" 2>&1) || true
    
    if echo "$headers" | grep -q "Connection refused\|Failed to connect"; then
        print_result "SKIP" "$test_name" "API Gateway not running"
        return
    fi
    
    if echo "$headers" | grep -q "Access-Control-Allow-Origin"; then
        print_result "PASS" "$test_name"
    else
        print_result "FAIL" "$test_name" "CORS headers missing"
    fi
}

echo "=========================================="
echo "API Gateway Route Testing"
echo "=========================================="
echo "API Gateway URL: $API_GATEWAY_URL"
echo "Test Merchant ID: $TEST_MERCHANT_ID"
echo ""

# Test health check
echo "--- Health Check Routes ---"
test_route "GET" "/health" "200" "Health Check"
test_route "GET" "/health?detailed=true" "200" "Health Check (Detailed)"
test_route "GET" "/" "200" "Root Endpoint"
test_route "GET" "/metrics" "200" "Metrics Endpoint"
echo ""

# Test merchant routes
echo "--- Merchant Routes ---"
test_route "GET" "/api/v1/merchants" "200" "Get All Merchants"
test_route "GET" "/api/v1/merchants/$TEST_MERCHANT_ID" "200" "Get Merchant by ID"
test_route "GET" "/api/v1/merchants/$TEST_MERCHANT_ID/analytics" "200" "Get Merchant Analytics"
test_route "GET" "/api/v1/merchants/$TEST_MERCHANT_ID/risk-score" "200" "Get Merchant Risk Score"
test_route "GET" "/api/v1/merchants/$TEST_MERCHANT_ID/website-analysis" "200" "Get Merchant Website Analysis"
test_route "GET" "/api/v1/merchants/analytics" "200" "Get Portfolio Analytics"
test_route "GET" "/api/v1/merchants/statistics" "200" "Get Portfolio Statistics"
test_route "POST" "/api/v1/merchants/search" "200" "Search Merchants" "" '{"query":"test"}'
echo ""

# Test analytics routes
echo "--- Analytics Routes ---"
test_route "GET" "/api/v1/analytics/trends" "200" "Get Risk Trends"
test_route "GET" "/api/v1/analytics/trends?timeframe=30d&limit=10" "200" "Get Risk Trends (with params)"
test_route "GET" "/api/v1/analytics/insights" "200" "Get Risk Insights"
test_route "GET" "/api/v1/analytics/insights?timeframe=90d&limit=5" "200" "Get Risk Insights (with params)"
echo ""

# Test risk assessment routes
echo "--- Risk Assessment Routes ---"
test_route "GET" "/api/v1/risk/benchmarks?industry=Technology" "200" "Get Risk Benchmarks"
test_route "GET" "/api/v1/risk/indicators/$TEST_MERCHANT_ID?status=active" "200" "Get Risk Indicators"
test_route "GET" "/api/v1/risk/predictions/$TEST_MERCHANT_ID" "200" "Get Risk Predictions"
test_route "GET" "/api/v1/risk/metrics" "200" "Get Risk Metrics"
test_route "POST" "/api/v1/risk/assess" "200" "Assess Risk" "" "{\"merchant_id\":\"$TEST_MERCHANT_ID\"}"
echo ""

# Test service health routes
echo "--- Service Health Routes ---"
test_route "GET" "/api/v1/classification/health" "200" "Classification Health"
test_route "GET" "/api/v1/merchant/health" "200" "Merchant Health"
test_route "GET" "/api/v1/risk/health" "200" "Risk Health"
echo ""

# Test V3 routes
echo "--- V3 Dashboard Routes ---"
test_route "GET" "/api/v3/dashboard/metrics" "200" "Dashboard Metrics V3"
echo ""

# Test error cases
echo "--- Error Cases ---"
test_route "GET" "/api/v1/merchants/invalid-id-123" "404" "Get Merchant (Invalid ID)"
test_route "GET" "/api/v1/merchants/invalid-id-123/analytics" "404" "Get Merchant Analytics (Invalid ID)"
test_route "GET" "/api/v1/risk/indicators/invalid-id-123" "404" "Get Risk Indicators (Invalid ID)"
test_route "GET" "/api/v1/nonexistent/route" "404" "Non-existent Route"
echo ""

# Test CORS
echo "--- CORS Headers ---"
test_cors "/api/v1/merchants" "CORS Headers (Merchants)"
test_cors "/api/v1/analytics/trends" "CORS Headers (Analytics)"
test_cors "/api/v1/risk/metrics" "CORS Headers (Risk)"
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "${YELLOW}Skipped: $SKIPPED${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo "Failed Tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} $test"
    done
    echo ""
    exit 1
fi

if [ $PASSED -gt 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}No tests ran (API Gateway may not be running)${NC}"
    exit 0
fi

