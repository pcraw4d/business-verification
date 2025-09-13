#!/bin/bash

# Business Intelligence Components Validation Script
# This script validates the business intelligence components without running the full server

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

echo -e "${BLUE}🧪 Business Intelligence Components Validation${NC}"
echo -e "${BLUE}=============================================${NC}"
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

# Function to check if file exists
check_file_exists() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✅ File exists: $1${NC}"
        return 0
    else
        echo -e "${RED}❌ File not found: $1${NC}"
        return 1
    fi
}

# Function to validate Go code compilation
validate_go_compilation() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating Go compilation for $test_name..."
    
    # Try to compile the file
    if go build -o /dev/null "$file" 2>/dev/null; then
        echo -e "${GREEN}  ✅ Go code compiles successfully${NC}"
        return 0
    else
        echo -e "${RED}  ❌ Go code compilation failed${NC}"
        return 1
    fi
}

# Function to validate business intelligence handler
validate_business_intelligence_handler() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating business intelligence handler for $test_name..."
    
    # Check for required types
    if grep -q "type BusinessIntelligenceHandler struct" "$file"; then
        echo -e "${GREEN}  ✅ BusinessIntelligenceHandler struct found${NC}"
    else
        echo -e "${RED}  ❌ BusinessIntelligenceHandler struct not found${NC}"
        return 1
    fi
    
    # Check for required types
    if grep -q "type BusinessIntelligenceType string" "$file"; then
        echo -e "${GREEN}  ✅ BusinessIntelligenceType found${NC}"
    else
        echo -e "${RED}  ❌ BusinessIntelligenceType not found${NC}"
        return 1
    fi
    
    # Check for required statuses
    if grep -q "type BusinessIntelligenceStatus string" "$file"; then
        echo -e "${GREEN}  ✅ BusinessIntelligenceStatus found${NC}"
    else
        echo -e "${RED}  ❌ BusinessIntelligenceStatus not found${NC}"
        return 1
    fi
    
    # Check for market analysis request
    if grep -q "type MarketAnalysisRequest struct" "$file"; then
        echo -e "${GREEN}  ✅ MarketAnalysisRequest struct found${NC}"
    else
        echo -e "${RED}  ❌ MarketAnalysisRequest struct not found${NC}"
        return 1
    fi
    
    # Check for competitive analysis request
    if grep -q "type CompetitiveAnalysisRequest struct" "$file"; then
        echo -e "${GREEN}  ✅ CompetitiveAnalysisRequest struct found${NC}"
    else
        echo -e "${RED}  ❌ CompetitiveAnalysisRequest struct not found${NC}"
        return 1
    fi
    
    # Check for growth analytics request
    if grep -q "type GrowthAnalyticsRequest struct" "$file"; then
        echo -e "${GREEN}  ✅ GrowthAnalyticsRequest struct found${NC}"
    else
        echo -e "${RED}  ❌ GrowthAnalyticsRequest struct not found${NC}"
        return 1
    fi
    
    # Check for analysis options
    if grep -q "type AnalysisOptions struct" "$file"; then
        echo -e "${GREEN}  ✅ AnalysisOptions struct found${NC}"
    else
        echo -e "${RED}  ❌ AnalysisOptions struct not found${NC}"
        return 1
    fi
    
    # Check for time range
    if grep -q "type BITimeRange struct" "$file"; then
        echo -e "${GREEN}  ✅ BITimeRange struct found${NC}"
    else
        echo -e "${RED}  ❌ BITimeRange struct not found${NC}"
        return 1
    fi
    
    # Check for competitor data
    if grep -q "type CompetitorData struct" "$file"; then
        echo -e "${GREEN}  ✅ CompetitorData struct found${NC}"
    else
        echo -e "${RED}  ❌ CompetitorData struct not found${NC}"
        return 1
    fi
    
    # Check for market position data
    if grep -q "type MarketPositionData struct" "$file"; then
        echo -e "${GREEN}  ✅ MarketPositionData struct found${NC}"
    else
        echo -e "${RED}  ❌ MarketPositionData struct not found${NC}"
        return 1
    fi
    
    # Check for business intelligence response types
    if grep -q "type MarketAnalysisResponse struct" "$file"; then
        echo -e "${GREEN}  ✅ MarketAnalysisResponse struct found${NC}"
    else
        echo -e "${RED}  ❌ MarketAnalysisResponse struct not found${NC}"
        return 1
    fi
    
    if grep -q "type CompetitiveAnalysisResponse struct" "$file"; then
        echo -e "${GREEN}  ✅ CompetitiveAnalysisResponse struct found${NC}"
    else
        echo -e "${RED}  ❌ CompetitiveAnalysisResponse struct not found${NC}"
        return 1
    fi
    
    if grep -q "type GrowthAnalyticsResponse struct" "$file"; then
        echo -e "${GREEN}  ✅ GrowthAnalyticsResponse struct found${NC}"
    else
        echo -e "${RED}  ❌ GrowthAnalyticsResponse struct not found${NC}"
        return 1
    fi
    
    # Check for business intelligence job
    if grep -q "type BusinessIntelligenceJob struct" "$file"; then
        echo -e "${GREEN}  ✅ BusinessIntelligenceJob struct found${NC}"
    else
        echo -e "${RED}  ❌ BusinessIntelligenceJob struct not found${NC}"
        return 1
    fi
    
    # Check for business intelligence result
    if grep -q "type BusinessIntelligenceResult struct" "$file"; then
        echo -e "${GREEN}  ✅ BusinessIntelligenceResult struct found${NC}"
    else
        echo -e "${RED}  ❌ BusinessIntelligenceResult struct not found${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate business intelligence constants
validate_business_intelligence_constants() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating business intelligence constants for $test_name..."
    
    # Check for intelligence types
    local intelligence_types=(
        "IntelligenceTypeMarketAnalysis"
        "IntelligenceTypeCompetitiveAnalysis"
        "IntelligenceTypeGrowthAnalytics"
        "IntelligenceTypeIndustryBenchmark"
        "IntelligenceTypeRiskAssessment"
        "IntelligenceTypeComplianceCheck"
    )
    
    for type in "${intelligence_types[@]}"; do
        if grep -q "$type" "$file"; then
            echo -e "${GREEN}  ✅ $type found${NC}"
        else
            echo -e "${RED}  ❌ $type not found${NC}"
            return 1
        fi
    done
    
    # Check for status constants
    local status_constants=(
        "BIStatusPending"
        "BIStatusRunning"
        "BIStatusCompleted"
        "BIStatusFailed"
        "BIStatusCancelled"
    )
    
    for status in "${status_constants[@]}"; do
        if grep -q "$status" "$file"; then
            echo -e "${GREEN}  ✅ $status found${NC}"
        else
            echo -e "${RED}  ❌ $status not found${NC}"
            return 1
        fi
    done
    
    return 0
}

# Function to validate business intelligence methods
validate_business_intelligence_methods() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating business intelligence methods for $test_name..."
    
    # Check for constructor
    if grep -q "func NewBusinessIntelligenceHandler" "$file"; then
        echo -e "${GREEN}  ✅ NewBusinessIntelligenceHandler function found${NC}"
    else
        echo -e "${RED}  ❌ NewBusinessIntelligenceHandler function not found${NC}"
        return 1
    fi
    
    # Check for market analysis methods
    local market_methods=(
        "CreateMarketAnalysis"
        "GetMarketAnalysis"
        "ListMarketAnalyses"
        "CreateMarketAnalysisJob"
        "GetMarketAnalysisJob"
        "ListMarketAnalysisJobs"
    )
    
    for method in "${market_methods[@]}"; do
        if grep -q "$method" "$file"; then
            echo -e "${GREEN}  ✅ $method found${NC}"
        else
            echo -e "${RED}  ❌ $method not found${NC}"
            return 1
        fi
    done
    
    # Check for competitive analysis methods
    local competitive_methods=(
        "CreateCompetitiveAnalysis"
        "GetCompetitiveAnalysis"
        "ListCompetitiveAnalyses"
        "CreateCompetitiveAnalysisJob"
        "GetCompetitiveAnalysisJob"
        "ListCompetitiveAnalysisJobs"
    )
    
    for method in "${competitive_methods[@]}"; do
        if grep -q "$method" "$file"; then
            echo -e "${GREEN}  ✅ $method found${NC}"
        else
            echo -e "${RED}  ❌ $method not found${NC}"
            return 1
        fi
    done
    
    # Check for growth analytics methods
    local growth_methods=(
        "CreateGrowthAnalytics"
        "GetGrowthAnalytics"
        "ListGrowthAnalytics"
        "CreateGrowthAnalyticsJob"
        "GetGrowthAnalyticsJob"
        "ListGrowthAnalyticsJobs"
    )
    
    for method in "${growth_methods[@]}"; do
        if grep -q "$method" "$file"; then
            echo -e "${GREEN}  ✅ $method found${NC}"
        else
            echo -e "${RED}  ❌ $method not found${NC}"
            return 1
        fi
    done
    
    # Check for aggregation methods
    local aggregation_methods=(
        "CreateBusinessIntelligenceAggregation"
        "GetBusinessIntelligenceAggregation"
        "ListBusinessIntelligenceAggregations"
        "CreateBusinessIntelligenceAggregationJob"
        "GetBusinessIntelligenceAggregationJob"
        "ListBusinessIntelligenceAggregationJobs"
    )
    
    for method in "${aggregation_methods[@]}"; do
        if grep -q "$method" "$file"; then
            echo -e "${GREEN}  ✅ $method found${NC}"
        else
            echo -e "${RED}  ❌ $method not found${NC}"
            return 1
        fi
    done
    
    return 0
}

# Function to validate routes
validate_routes() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating routes for $test_name..."
    
    # Check for business intelligence routes
    local routes=(
        "/v2/business-intelligence/market-analysis"
        "/v2/business-intelligence/competitive-analysis"
        "/v2/business-intelligence/growth-analytics"
        "/v2/business-intelligence/aggregation"
    )
    
    for route in "${routes[@]}"; do
        if grep -q "$route" "$file"; then
            echo -e "${GREEN}  ✅ Route $route found${NC}"
        else
            echo -e "${RED}  ❌ Route $route not found${NC}"
            return 1
        fi
    done
    
    return 0
}

# Function to generate test report
generate_test_report() {
    local report_file="$REPORT_DIR/business-intelligence-components-validation-report-$TIMESTAMP.txt"
    
    echo "Generating test report: $report_file"
    
    cat > "$report_file" << EOF
Business Intelligence Components Validation - Test Report
=======================================================
Generated: $(date)
Test Suite: Component Validation
Version: 1.0.0

Test Results Summary:
- Total Tests Run: $TOTAL_TESTS
- Tests Passed: $TESTS_PASSED
- Tests Failed: $TESTS_FAILED
- Success Rate: $(( TESTS_PASSED * 100 / TOTAL_TESTS ))%

Test Categories:
1. File Existence Tests
2. Go Compilation Tests
3. Business Intelligence Handler Validation
4. Business Intelligence Constants Validation
5. Business Intelligence Methods Validation
6. Routes Validation

Detailed Results:
EOF
    
    echo "Test report generated: $report_file"
}

# Main test execution
main() {
    local total_tests=0
    local tests_passed=0
    local tests_failed=0
    
    print_test_header "File Existence Tests"
    
    # Test 1: Check if business intelligence handler file exists
    if check_file_exists "$TEST_DIR/internal/api/handlers/business_intelligence_handler.go"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    # Test 2: Check if routes file exists
    if check_file_exists "$TEST_DIR/internal/api/routes/routes.go"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Go Compilation Tests"
    
    # Test 3: Validate business intelligence handler compilation
    if validate_go_compilation "$TEST_DIR/internal/api/handlers/business_intelligence_handler.go" "Business Intelligence Handler"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    # Test 4: Validate routes compilation
    if validate_go_compilation "$TEST_DIR/internal/api/routes/routes.go" "Routes"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Business Intelligence Handler Validation"
    
    # Test 5: Validate business intelligence handler structure
    if validate_business_intelligence_handler "$TEST_DIR/internal/api/handlers/business_intelligence_handler.go" "Business Intelligence Handler"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Business Intelligence Constants Validation"
    
    # Test 6: Validate business intelligence constants
    if validate_business_intelligence_constants "$TEST_DIR/internal/api/handlers/business_intelligence_handler.go" "Business Intelligence Constants"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Business Intelligence Methods Validation"
    
    # Test 7: Validate business intelligence methods
    if validate_business_intelligence_methods "$TEST_DIR/internal/api/handlers/business_intelligence_handler.go" "Business Intelligence Methods"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Routes Validation"
    
    # Test 8: Validate routes
    if validate_routes "$TEST_DIR/internal/api/routes/routes.go" "Routes"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
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
        echo -e "${GREEN}🎉 All tests passed! Business Intelligence components are properly structured.${NC}"
        exit 0
    else
        echo -e "${RED}⚠️  Some tests failed. Please review the issues above.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
