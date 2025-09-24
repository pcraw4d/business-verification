#!/bin/bash

# End-to-End Test Runner for Merchant Portfolio Management
# This script runs comprehensive end-to-end tests for the merchant-centric UI implementation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
E2E_TESTS_DIR="test/e2e"
TEST_TIMEOUT="10m"
COVERAGE_OUTPUT="test/e2e/coverage.out"
TEST_RESULTS_DIR="test/e2e/results"

# Create results directory
mkdir -p "$TEST_RESULTS_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Merchant Portfolio E2E Test Runner  ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print test section header
print_section() {
    echo -e "${YELLOW}=== $1 ===${NC}"
}

# Function to run a specific test
run_test() {
    local test_name="$1"
    local test_file="$2"
    local description="$3"
    
    print_section "$description"
    echo "Running: $test_name"
    echo "File: $test_file"
    echo ""
    
    # Set environment variable for E2E tests
    export E2E_TESTS=true
    
    # Run the test with timeout
    if timeout "$TEST_TIMEOUT" go test -v -run "$test_name" "$test_file" -coverprofile="$TEST_RESULTS_DIR/${test_name}_coverage.out" > "$TEST_RESULTS_DIR/${test_name}_output.log" 2>&1; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        echo "Check $TEST_RESULTS_DIR/${test_name}_output.log for details"
        return 1
    fi
}

# Function to run all tests in a file
run_test_file() {
    local test_file="$1"
    local description="$2"
    
    print_section "$description"
    echo "Running all tests in: $test_file"
    echo ""
    
    # Set environment variable for E2E tests
    export E2E_TESTS=true
    
    # Run all tests in the file
    if timeout "$TEST_TIMEOUT" go test -v "$test_file" -coverprofile="$TEST_RESULTS_DIR/$(basename "$test_file" .go)_coverage.out" > "$TEST_RESULTS_DIR/$(basename "$test_file" .go)_output.log" 2>&1; then
        echo -e "${GREEN}✅ All tests in $(basename "$test_file") PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ Some tests in $(basename "$test_file") FAILED${NC}"
        echo "Check $TEST_RESULTS_DIR/$(basename "$test_file" .go)_output.log for details"
        return 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    print_section "Generating Coverage Report"
    
    # Combine all coverage files
    echo "mode: set" > "$COVERAGE_OUTPUT"
    for coverage_file in "$TEST_RESULTS_DIR"/*_coverage.out; do
        if [ -f "$coverage_file" ]; then
            tail -n +2 "$coverage_file" >> "$COVERAGE_OUTPUT"
        fi
    done
    
    # Generate HTML coverage report
    if command -v go tool cover >/dev/null 2>&1; then
        go tool cover -html="$COVERAGE_OUTPUT" -o "$TEST_RESULTS_DIR/coverage.html"
        echo -e "${GREEN}✅ Coverage report generated: $TEST_RESULTS_DIR/coverage.html${NC}"
    else
        echo -e "${YELLOW}⚠️  go tool cover not available, skipping HTML report${NC}"
    fi
    
    # Show coverage summary
    if [ -f "$COVERAGE_OUTPUT" ]; then
        echo ""
        echo "Coverage Summary:"
        go tool cover -func="$COVERAGE_OUTPUT" | tail -1
    fi
}

# Function to generate test summary
generate_test_summary() {
    print_section "Test Summary"
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Count test results
    for output_file in "$TEST_RESULTS_DIR"/*_output.log; do
        if [ -f "$output_file" ]; then
            total_tests=$((total_tests + 1))
            if grep -q "PASS" "$output_file"; then
                passed_tests=$((passed_tests + 1))
            else
                failed_tests=$((failed_tests + 1))
            fi
        fi
    done
    
    echo "Total Test Files: $total_tests"
    echo -e "Passed: ${GREEN}$passed_tests${NC}"
    echo -e "Failed: ${RED}$failed_tests${NC}"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}🎉 All E2E tests passed!${NC}"
        return 0
    else
        echo -e "${RED}💥 $failed_tests test file(s) failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    local start_time=$(date +%s)
    local exit_code=0
    
    echo "Starting E2E tests at $(date)"
    echo "Test timeout: $TEST_TIMEOUT"
    echo "Results directory: $TEST_RESULTS_DIR"
    echo ""
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}Error: go.mod not found. Please run from project root.${NC}"
        exit 1
    fi
    
    # Check if E2E test files exist
    if [ ! -d "$E2E_TESTS_DIR" ]; then
        echo -e "${RED}Error: E2E test directory not found: $E2E_TESTS_DIR${NC}"
        exit 1
    fi
    
    # Run individual test files
    echo -e "${BLUE}Running E2E Test Files...${NC}"
    echo ""
    
    # Test 1: Merchant Workflow E2E
    if ! run_test_file "$E2E_TESTS_DIR/merchant_workflow_e2e_test.go" "Merchant Workflow E2E Tests"; then
        exit_code=1
    fi
    
    # Test 2: Bulk Operations E2E
    if ! run_test_file "$E2E_TESTS_DIR/bulk_operations_e2e_test.go" "Bulk Operations E2E Tests"; then
        exit_code=1
    fi
    
    # Test 3: Merchant Comparison E2E
    if ! run_test_file "$E2E_TESTS_DIR/merchant_comparison_e2e_test.go" "Merchant Comparison E2E Tests"; then
        exit_code=1
    fi
    
    # Test 4: User Journey E2E
    if ! run_test_file "$E2E_TESTS_DIR/user_journey_e2e_test.go" "User Journey E2E Tests"; then
        exit_code=1
    fi
    
    echo ""
    
    # Generate coverage report
    generate_coverage_report
    
    echo ""
    
    # Generate test summary
    generate_test_summary
    
    # Calculate execution time
    local end_time=$(date +%s)
    local execution_time=$((end_time - start_time))
    echo ""
    echo "Total execution time: ${execution_time}s"
    
    # List result files
    echo ""
    print_section "Generated Files"
    echo "Results directory: $TEST_RESULTS_DIR"
    if [ -d "$TEST_RESULTS_DIR" ]; then
        ls -la "$TEST_RESULTS_DIR"
    fi
    
    echo ""
    echo "E2E tests completed at $(date)"
    
    exit $exit_code
}

# Handle script arguments
case "${1:-}" in
    "merchant-workflow")
        run_test_file "$E2E_TESTS_DIR/merchant_workflow_e2e_test.go" "Merchant Workflow E2E Tests"
        ;;
    "bulk-operations")
        run_test_file "$E2E_TESTS_DIR/bulk_operations_e2e_test.go" "Bulk Operations E2E Tests"
        ;;
    "merchant-comparison")
        run_test_file "$E2E_TESTS_DIR/merchant_comparison_e2e_test.go" "Merchant Comparison E2E Tests"
        ;;
    "user-journey")
        run_test_file "$E2E_TESTS_DIR/user_journey_e2e_test.go" "User Journey E2E Tests"
        ;;
    "coverage")
        generate_coverage_report
        ;;
    "summary")
        generate_test_summary
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [test-type]"
        echo ""
        echo "Test types:"
        echo "  merchant-workflow  Run merchant workflow E2E tests"
        echo "  bulk-operations    Run bulk operations E2E tests"
        echo "  merchant-comparison Run merchant comparison E2E tests"
        echo "  user-journey       Run user journey E2E tests"
        echo "  coverage          Generate coverage report"
        echo "  summary           Generate test summary"
        echo "  (no args)         Run all E2E tests"
        echo ""
        echo "Environment variables:"
        echo "  E2E_TESTS=true    Enable E2E tests (automatically set)"
        echo "  TEST_TIMEOUT      Test timeout (default: 10m)"
        ;;
    *)
        main
        ;;
esac
