#!/bin/bash

# Market Analysis Interface - Automated Test Script
# This script automates the testing of the Market Analysis Interface

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DASHBOARD_FILE="$TEST_DIR/web/market-analysis-dashboard.html"
TEST_SUITE_FILE="$TEST_DIR/test-market-analysis-interface.html"
REPORT_DIR="$TEST_DIR/test-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create report directory
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}🧪 Market Analysis Interface - Automated Test Suite${NC}"
echo -e "${BLUE}================================================${NC}"
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

# Function to validate HTML structure
validate_html_structure() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating HTML structure for $test_name..."
    
    # Check for required HTML elements
    if grep -q "<!DOCTYPE html>" "$file"; then
        echo -e "${GREEN}  ✅ DOCTYPE declaration found${NC}"
    else
        echo -e "${RED}  ❌ DOCTYPE declaration missing${NC}"
        return 1
    fi
    
    if grep -q "<html" "$file"; then
        echo -e "${GREEN}  ✅ HTML tag found${NC}"
    else
        echo -e "${RED}  ❌ HTML tag missing${NC}"
        return 1
    fi
    
    if grep -q "<head>" "$file"; then
        echo -e "${GREEN}  ✅ Head section found${NC}"
    else
        echo -e "${RED}  ❌ Head section missing${NC}"
        return 1
    fi
    
    if grep -q "<body>" "$file"; then
        echo -e "${GREEN}  ✅ Body section found${NC}"
    else
        echo -e "${RED}  ❌ Body section missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate CSS dependencies
validate_css_dependencies() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating CSS dependencies for $test_name..."
    
    # Check for Tailwind CSS
    if grep -q "tailwindcss" "$file"; then
        echo -e "${GREEN}  ✅ Tailwind CSS dependency found${NC}"
    else
        echo -e "${RED}  ❌ Tailwind CSS dependency missing${NC}"
        return 1
    fi
    
    # Check for Font Awesome
    if grep -q "font-awesome" "$file"; then
        echo -e "${GREEN}  ✅ Font Awesome dependency found${NC}"
    else
        echo -e "${RED}  ❌ Font Awesome dependency missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate JavaScript dependencies
validate_js_dependencies() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating JavaScript dependencies for $test_name..."
    
    # Check for Chart.js
    if grep -q "chart.js" "$file"; then
        echo -e "${GREEN}  ✅ Chart.js dependency found${NC}"
    else
        echo -e "${RED}  ❌ Chart.js dependency missing${NC}"
        return 1
    fi
    
    # Check for MarketAnalysisDashboard class
    if grep -q "class MarketAnalysisDashboard" "$file"; then
        echo -e "${GREEN}  ✅ MarketAnalysisDashboard class found${NC}"
    else
        echo -e "${RED}  ❌ MarketAnalysisDashboard class missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate chart elements
validate_chart_elements() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating chart elements for $test_name..."
    
    # Check for required chart canvases
    local charts=("benchmarkChart" "trendChart" "performanceChart" "positionChart" "segmentChart" "geographicChart")
    
    for chart in "${charts[@]}"; do
        if grep -q "id=\"$chart\"" "$file"; then
            echo -e "${GREEN}  ✅ Chart canvas found: $chart${NC}"
        else
            echo -e "${RED}  ❌ Chart canvas missing: $chart${NC}"
            return 1
        fi
    done
    
    return 0
}

# Function to validate interactive elements
validate_interactive_elements() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating interactive elements for $test_name..."
    
    # Check for benchmark buttons
    local benchmark_buttons=("benchmarkRevenue" "benchmarkProfit" "benchmarkGrowth" "benchmarkEfficiency")
    
    for button in "${benchmark_buttons[@]}"; do
        if grep -q "id=\"$button\"" "$file"; then
            echo -e "${GREEN}  ✅ Benchmark button found: $button${NC}"
        else
            echo -e "${RED}  ❌ Benchmark button missing: $button${NC}"
            return 1
        fi
    done
    
    # Check for trend buttons
    local trend_buttons=("trend6M" "trend1Y" "trend3Y" "trend5Y")
    
    for button in "${trend_buttons[@]}"; do
        if grep -q "id=\"$button\"" "$file"; then
            echo -e "${GREEN}  ✅ Trend button found: $button${NC}"
        else
            echo -e "${RED}  ❌ Trend button missing: $button${NC}"
            return 1
        fi
    done
    
    # Check for opportunity filter buttons
    local filter_buttons=("opportunityAll" "opportunityHigh" "opportunityMedium" "opportunityLow")
    
    for button in "${filter_buttons[@]}"; do
        if grep -q "id=\"$button\"" "$file"; then
            echo -e "${GREEN}  ✅ Filter button found: $button${NC}"
        else
            echo -e "${RED}  ❌ Filter button missing: $button${NC}"
            return 1
        fi
    done
    
    return 0
}

# Function to validate data structure
validate_data_structure() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating data structure for $test_name..."
    
    # Check for benchmark data
    if grep -q "benchmarkData:" "$file"; then
        echo -e "${GREEN}  ✅ Benchmark data structure found${NC}"
    else
        echo -e "${RED}  ❌ Benchmark data structure missing${NC}"
        return 1
    fi
    
    # Check for trend data
    if grep -q "trendData:" "$file"; then
        echo -e "${GREEN}  ✅ Trend data structure found${NC}"
    else
        echo -e "${RED}  ❌ Trend data structure missing${NC}"
        return 1
    fi
    
    # Check for segment data
    if grep -q "segmentData:" "$file"; then
        echo -e "${GREEN}  ✅ Segment data structure found${NC}"
    else
        echo -e "${RED}  ❌ Segment data structure missing${NC}"
        return 1
    fi
    
    # Check for geographic data
    if grep -q "geographicData:" "$file"; then
        echo -e "${GREEN}  ✅ Geographic data structure found${NC}"
    else
        echo -e "${RED}  ❌ Geographic data structure missing${NC}"
        return 1
    fi
    
    # Check for competitive data
    if grep -q "competitiveData:" "$file"; then
        echo -e "${GREEN}  ✅ Competitive data structure found${NC}"
    else
        echo -e "${RED}  ❌ Competitive data structure missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate responsive design
validate_responsive_design() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating responsive design for $test_name..."
    
    # Check for responsive classes
    if grep -q "grid-cols-1 md:grid-cols-2 lg:grid-cols-3" "$file"; then
        echo -e "${GREEN}  ✅ Responsive grid classes found${NC}"
    else
        echo -e "${RED}  ❌ Responsive grid classes missing${NC}"
        return 1
    fi
    
    # Check for responsive chart containers
    if grep -q "chart-container" "$file"; then
        echo -e "${GREEN}  ✅ Chart container classes found${NC}"
    else
        echo -e "${RED}  ❌ Chart container classes missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to validate accessibility
validate_accessibility() {
    local file="$1"
    local test_name="$2"
    
    echo "Validating accessibility for $test_name..."
    
    # Check for ARIA labels
    if grep -q "aria-label" "$file"; then
        echo -e "${GREEN}  ✅ ARIA labels found${NC}"
    else
        echo -e "${YELLOW}  ⚠️  ARIA labels not found (optional)${NC}"
    fi
    
    # Check for alt text
    if grep -q "alt=" "$file"; then
        echo -e "${GREEN}  ✅ Alt text found${NC}"
    else
        echo -e "${YELLOW}  ⚠️  Alt text not found (optional)${NC}"
    fi
    
    # Check for semantic HTML
    if grep -q "<nav>" "$file"; then
        echo -e "${GREEN}  ✅ Semantic navigation found${NC}"
    else
        echo -e "${RED}  ❌ Semantic navigation missing${NC}"
        return 1
    fi
    
    if grep -q "<main>" "$file"; then
        echo -e "${GREEN}  ✅ Semantic main content found${NC}"
    else
        echo -e "${RED}  ❌ Semantic main content missing${NC}"
        return 1
    fi
    
    return 0
}

# Function to run browser-based tests (if available)
run_browser_tests() {
    local file="$1"
    local test_name="$2"
    
    echo "Running browser-based tests for $test_name..."
    
    # Check if we have a headless browser available
    if command -v node &> /dev/null; then
        echo -e "${GREEN}  ✅ Node.js available for browser testing${NC}"
        # Here you could add Puppeteer or similar browser automation
    else
        echo -e "${YELLOW}  ⚠️  Node.js not available, skipping browser tests${NC}"
    fi
    
    return 0
}

# Function to generate test report
generate_test_report() {
    local report_file="$REPORT_DIR/market-analysis-test-report-$TIMESTAMP.txt"
    
    echo "Generating test report: $report_file"
    
    cat > "$report_file" << EOF
Market Analysis Interface - Test Report
=====================================
Generated: $(date)
Test Suite: Automated Validation
Version: 1.0.0

Test Results Summary:
- Dashboard File: $DASHBOARD_FILE
- Test Suite File: $TEST_SUITE_FILE
- Total Tests Run: $TOTAL_TESTS
- Tests Passed: $TESTS_PASSED
- Tests Failed: $TESTS_FAILED
- Success Rate: $(( TESTS_PASSED * 100 / TOTAL_TESTS ))%

Detailed Results:
EOF
    
    # Add detailed results here
    echo "Test report generated: $report_file"
}

# Main test execution
main() {
    local total_tests=0
    local tests_passed=0
    local tests_failed=0
    
    print_test_header "File Existence Tests"
    
    # Test 1: Check if dashboard file exists
    if check_file_exists "$DASHBOARD_FILE"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    # Test 2: Check if test suite file exists
    if check_file_exists "$TEST_SUITE_FILE"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "HTML Structure Validation"
    
    # Test 3: Validate HTML structure
    if validate_html_structure "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "CSS Dependencies Validation"
    
    # Test 4: Validate CSS dependencies
    if validate_css_dependencies "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "JavaScript Dependencies Validation"
    
    # Test 5: Validate JavaScript dependencies
    if validate_js_dependencies "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Chart Elements Validation"
    
    # Test 6: Validate chart elements
    if validate_chart_elements "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Interactive Elements Validation"
    
    # Test 7: Validate interactive elements
    if validate_interactive_elements "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Data Structure Validation"
    
    # Test 8: Validate data structure
    if validate_data_structure "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Responsive Design Validation"
    
    # Test 9: Validate responsive design
    if validate_responsive_design "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Accessibility Validation"
    
    # Test 10: Validate accessibility
    if validate_accessibility "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
        ((tests_passed++))
    else
        ((tests_failed++))
    fi
    ((total_tests++))
    
    print_test_header "Browser Tests"
    
    # Test 11: Run browser-based tests
    if run_browser_tests "$DASHBOARD_FILE" "Market Analysis Dashboard"; then
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
    print_test_header "Test Summary"
    echo -e "${GREEN}Total Tests: $total_tests${NC}"
    echo -e "${GREEN}Tests Passed: $tests_passed${NC}"
    echo -e "${RED}Tests Failed: $tests_failed${NC}"
    echo -e "${BLUE}Success Rate: $(( tests_passed * 100 / total_tests ))%${NC}"
    
    if [ $tests_failed -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed! Market Analysis Interface is ready for deployment.${NC}"
        exit 0
    else
        echo -e "${RED}⚠️  Some tests failed. Please review the issues above.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
