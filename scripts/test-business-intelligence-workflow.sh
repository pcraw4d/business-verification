#!/bin/bash

# Business Intelligence Workflow Testing Script
# This script tests the end-to-end business intelligence workflow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="$TEST_DIR/test-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create report directory
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}🧪 Business Intelligence Workflow Testing${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Function to print test header
print_test_header() {
    echo -e "${BLUE}📋 $1${NC}"
    echo -e "${BLUE}$(printf '=%.0s' {1..50})${NC}"
}

# Function to print test result
print_test_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2 - PASSED${NC}"
    else
        echo -e "${RED}❌ $2 - FAILED${NC}"
    fi
}

# Function to test API endpoint
test_api_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    local test_name="$5"
    
    echo "Testing $test_name..."
    
    # Start the server in background
    cd "$TEST_DIR"
    go run cmd/api-enhanced/main-enhanced-with-database-classification.go &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Test the endpoint
    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "http://localhost:8080$endpoint")
    else
        response=$(curl -s -w "%{http_code}" \
            -X "$method" \
            "http://localhost:8080$endpoint")
    fi
    
    # Extract status code (last 3 characters)
    status_code="${response: -3}"
    response_body="${response%???}"
    
    # Stop the server
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}  ✅ Status code: $status_code (expected: $expected_status)${NC}"
        echo -e "${GREEN}  ✅ Response received${NC}"
        return 0
    else
        echo -e "${RED}  ❌ Status code: $status_code (expected: $expected_status)${NC}"
        echo -e "${RED}  ❌ Response: $response_body${NC}"
        return 1
    fi
}

# Function to test market analysis workflow
test_market_analysis_workflow() {
    print_test_header "Market Analysis Workflow Testing"
    
    local tests_passed=0
    local tests_failed=0
    
    # Test 1: Create market analysis
    market_analysis_data='{
        "business_id": "workflow-test-business",
        "industry": "Technology",
        "geographic_area": "North America",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z",
            "time_zone": "UTC"
        },
        "parameters": {
            "market_size_focus": "total"
        },
        "options": {
            "real_time": true,
            "batch_mode": false,
            "parallel": true,
            "notifications": true,
            "audit_trail": true,
            "monitoring": true,
            "validation": true
        }
    }'
    
    if test_api_endpoint "POST" "/v2/business-intelligence/market-analysis" "$market_analysis_data" "200" "Create Market Analysis"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 2: List market analyses
    if test_api_endpoint "GET" "/v2/business-intelligence/market-analyses" "" "200" "List Market Analyses"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 3: Create market analysis job
    if test_api_endpoint "POST" "/v2/business-intelligence/market-analysis/jobs" "$market_analysis_data" "200" "Create Market Analysis Job"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 4: List market analysis jobs
    if test_api_endpoint "GET" "/v2/business-intelligence/market-analysis/jobs/list" "" "200" "List Market Analysis Jobs"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    echo ""
    echo -e "${BLUE}Market Analysis Workflow Results:${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    
    return $tests_failed
}

# Function to test competitive analysis workflow
test_competitive_analysis_workflow() {
    print_test_header "Competitive Analysis Workflow Testing"
    
    local tests_passed=0
    local tests_failed=0
    
    # Test 1: Create competitive analysis
    competitive_analysis_data='{
        "business_id": "workflow-test-business",
        "industry": "Technology",
        "geographic_area": "North America",
        "competitors": ["Competitor A", "Competitor B", "Competitor C"],
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z",
            "time_zone": "UTC"
        },
        "parameters": {
            "analysis_depth": "comprehensive"
        },
        "options": {
            "real_time": true,
            "batch_mode": false,
            "parallel": true,
            "notifications": true,
            "audit_trail": true,
            "monitoring": true,
            "validation": true
        }
    }'
    
    if test_api_endpoint "POST" "/v2/business-intelligence/competitive-analysis" "$competitive_analysis_data" "200" "Create Competitive Analysis"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 2: List competitive analyses
    if test_api_endpoint "GET" "/v2/business-intelligence/competitive-analyses" "" "200" "List Competitive Analyses"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 3: Create competitive analysis job
    if test_api_endpoint "POST" "/v2/business-intelligence/competitive-analysis/jobs" "$competitive_analysis_data" "200" "Create Competitive Analysis Job"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 4: List competitive analysis jobs
    if test_api_endpoint "GET" "/v2/business-intelligence/competitive-analysis/jobs/list" "" "200" "List Competitive Analysis Jobs"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    echo ""
    echo -e "${BLUE}Competitive Analysis Workflow Results:${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    
    return $tests_failed
}

# Function to test growth analytics workflow
test_growth_analytics_workflow() {
    print_test_header "Growth Analytics Workflow Testing"
    
    local tests_passed=0
    local tests_failed=0
    
    # Test 1: Create growth analytics
    growth_analytics_data='{
        "business_id": "workflow-test-business",
        "industry": "Technology",
        "geographic_area": "North America",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z",
            "time_zone": "UTC"
        },
        "parameters": {
            "growth_metrics": ["revenue", "market_share", "customer_base"]
        },
        "options": {
            "real_time": true,
            "batch_mode": false,
            "parallel": true,
            "notifications": true,
            "audit_trail": true,
            "monitoring": true,
            "validation": true
        }
    }'
    
    if test_api_endpoint "POST" "/v2/business-intelligence/growth-analytics" "$growth_analytics_data" "200" "Create Growth Analytics"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 2: List growth analytics
    if test_api_endpoint "GET" "/v2/business-intelligence/growth-analytics/list" "" "200" "List Growth Analytics"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 3: Create growth analytics job
    if test_api_endpoint "POST" "/v2/business-intelligence/growth-analytics/jobs" "$growth_analytics_data" "200" "Create Growth Analytics Job"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 4: List growth analytics jobs
    if test_api_endpoint "GET" "/v2/business-intelligence/growth-analytics/jobs/list" "" "200" "List Growth Analytics Jobs"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    echo ""
    echo -e "${BLUE}Growth Analytics Workflow Results:${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    
    return $tests_failed
}

# Function to test business intelligence aggregation workflow
test_business_intelligence_aggregation_workflow() {
    print_test_header "Business Intelligence Aggregation Workflow Testing"
    
    local tests_passed=0
    local tests_failed=0
    
    # Test 1: Create business intelligence aggregation
    aggregation_data='{
        "business_id": "workflow-test-business",
        "industry": "Technology",
        "geographic_area": "North America",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z",
            "time_zone": "UTC"
        },
        "analysis_types": ["market_analysis", "competitive_analysis", "growth_analytics"],
        "parameters": {
            "aggregation_level": "comprehensive"
        },
        "options": {
            "real_time": true,
            "batch_mode": false,
            "parallel": true,
            "notifications": true,
            "audit_trail": true,
            "monitoring": true,
            "validation": true
        }
    }'
    
    if test_api_endpoint "POST" "/v2/business-intelligence/aggregation" "$aggregation_data" "200" "Create Business Intelligence Aggregation"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 2: List business intelligence aggregations
    if test_api_endpoint "GET" "/v2/business-intelligence/aggregations" "" "200" "List Business Intelligence Aggregations"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 3: Create business intelligence aggregation job
    if test_api_endpoint "POST" "/v2/business-intelligence/aggregation/jobs" "$aggregation_data" "200" "Create Business Intelligence Aggregation Job"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 4: List business intelligence aggregation jobs
    if test_api_endpoint "GET" "/v2/business-intelligence/aggregation/jobs/list" "" "200" "List Business Intelligence Aggregation Jobs"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    echo ""
    echo -e "${BLUE}Business Intelligence Aggregation Workflow Results:${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    
    return $tests_failed
}

# Function to test error handling
test_error_handling() {
    print_test_header "Error Handling Testing"
    
    local tests_passed=0
    local tests_failed=0
    
    # Test 1: Invalid JSON
    if test_api_endpoint "POST" "/v2/business-intelligence/market-analysis" '{"invalid": json}' "400" "Invalid JSON Handling"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 2: Missing required fields
    if test_api_endpoint "POST" "/v2/business-intelligence/competitive-analysis" '{"business_id": "test"}' "400" "Missing Required Fields"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 3: Invalid time range
    if test_api_endpoint "POST" "/v2/business-intelligence/growth-analytics" '{"business_id": "test", "industry": "Technology", "geographic_area": "North America", "time_range": {"start_date": "2024-01-01T00:00:00Z", "end_date": "2023-01-01T00:00:00Z"}}' "400" "Invalid Time Range"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    # Test 4: Non-existent job ID
    if test_api_endpoint "GET" "/v2/business-intelligence/market-analysis/jobs?id=non-existent" "" "404" "Non-existent Job ID"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    
    echo ""
    echo -e "${BLUE}Error Handling Results:${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    
    return $tests_failed
}

# Function to generate test report
generate_test_report() {
    local report_file="$REPORT_DIR/business-intelligence-workflow-test-report-$TIMESTAMP.txt"
    
    echo "Generating test report: $report_file"
    
    cat > "$report_file" << EOF
Business Intelligence Workflow Testing - Test Report
==================================================
Generated: $(date)
Test Suite: End-to-End Workflow Testing
Version: 1.0.0

Test Results Summary:
- Total Tests Run: $TOTAL_TESTS
- Tests Passed: $TESTS_PASSED
- Tests Failed: $TESTS_FAILED
- Success Rate: $(( TESTS_PASSED * 100 / TOTAL_TESTS ))%

Test Categories:
1. Market Analysis Workflow Testing
2. Competitive Analysis Workflow Testing
3. Growth Analytics Workflow Testing
4. Business Intelligence Aggregation Workflow Testing
5. Error Handling Testing

Detailed Results:
EOF
    
    echo "Test report generated: $report_file"
}

# Main test execution
main() {
    local total_tests=0
    local tests_passed=0
    local tests_failed=0
    
    # Test market analysis workflow
    if test_market_analysis_workflow; then
        ((tests_passed += 4))
    else
        ((tests_failed += 4))
    fi
    ((total_tests += 4))
    
    # Test competitive analysis workflow
    if test_competitive_analysis_workflow; then
        ((tests_passed += 4))
    else
        ((tests_failed += 4))
    fi
    ((total_tests += 4))
    
    # Test growth analytics workflow
    if test_growth_analytics_workflow; then
        ((tests_passed += 4))
    else
        ((tests_failed += 4))
    fi
    ((total_tests += 4))
    
    # Test business intelligence aggregation workflow
    if test_business_intelligence_aggregation_workflow; then
        ((tests_passed += 4))
    else
        ((tests_failed += 4))
    fi
    ((total_tests += 4))
    
    # Test error handling
    if test_error_handling; then
        ((tests_passed += 4))
    else
        ((tests_failed += 4))
    fi
    ((total_tests += 4))
    
    # Set global variables for report generation
    TOTAL_TESTS=$total_tests
    TESTS_PASSED=$tests_passed
    TESTS_FAILED=$tests_failed
    
    # Generate test report
    generate_test_report
    
    # Print final summary
    echo ""
    print_test_header "Final Test Summary"
    echo -e "${GREEN}Total Tests: $total_tests${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    echo -e "${BLUE}Success Rate: $(( tests_passed * 100 / total_tests ))%${NC}"
    
    if [ $tests_failed -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed! Business Intelligence workflow is working correctly.${NC}"
        exit 0
    else
        echo -e "${RED}⚠️  Some tests failed. Please review the issues above.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
