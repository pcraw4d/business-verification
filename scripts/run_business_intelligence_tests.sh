#!/bin/bash

# Business Intelligence API Test Runner
# This script runs all tests for the Business Intelligence API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to run tests with coverage
run_tests_with_coverage() {
    local test_path=$1
    local test_name=$2
    
    print_status "Running $test_name tests..."
    
    if go test -v -coverprofile=coverage.out -covermode=atomic "$test_path"; then
        print_success "$test_name tests passed"
        
        # Generate coverage report
        if command -v go tool cover &> /dev/null; then
            coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            print_status "Coverage for $test_name: $coverage"
        fi
        
        return 0
    else
        print_error "$test_name tests failed"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local benchmark_path=$1
    local benchmark_name=$2
    
    print_status "Running $benchmark_name benchmarks..."
    
    if go test -bench=. -benchmem "$benchmark_path"; then
        print_success "$benchmark_name benchmarks completed"
        return 0
    else
        print_error "$benchmark_name benchmarks failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    local integration_path=$1
    local integration_name=$2
    
    print_status "Running $integration_name integration tests..."
    
    if go test -v -tags=integration "$integration_path"; then
        print_success "$integration_name integration tests passed"
        return 0
    else
        print_error "$integration_name integration tests failed"
        return 1
    fi
}

# Function to run error handling tests
run_error_handling_tests() {
    local error_path=$1
    local error_name=$2
    
    print_status "Running $error_name error handling tests..."
    
    if go test -v "$error_path"; then
        print_success "$error_name error handling tests passed"
        return 0
    else
        print_error "$error_name error handling tests failed"
        return 1
    fi
}

# Main execution
main() {
    print_status "Starting Business Intelligence API Test Suite"
    print_status "=============================================="
    
    # Change to project root directory
    cd "$(dirname "$0")/.."
    
    # Initialize test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Test paths
    local unit_test_path="./internal/api/handlers"
    local integration_test_path="./test/integration"
    local performance_test_path="./test/performance"
    local error_handling_test_path="./test/error_handling"
    
    # Run unit tests
    print_status "Phase 1: Unit Tests"
    print_status "-------------------"
    
    if run_tests_with_coverage "$unit_test_path" "Business Intelligence Handler Unit"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    ((total_tests++))
    
    # Run integration tests
    print_status "Phase 2: Integration Tests"
    print_status "--------------------------"
    
    if run_integration_tests "$integration_test_path" "Business Intelligence Integration"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    ((total_tests++))
    
    # Run performance tests
    print_status "Phase 3: Performance Tests"
    print_status "--------------------------"
    
    if run_benchmarks "$performance_test_path" "Business Intelligence Performance"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    ((total_tests++))
    
    # Run error handling tests
    print_status "Phase 4: Error Handling Tests"
    print_status "------------------------------"
    
    if run_error_handling_tests "$error_handling_test_path" "Business Intelligence Error Handling"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    ((total_tests++))
    
    # Run comprehensive test suite
    print_status "Phase 5: Comprehensive Test Suite"
    print_status "----------------------------------"
    
    if go test -v -race -coverprofile=coverage.out -covermode=atomic ./...; then
        print_success "Comprehensive test suite passed"
        ((passed_tests++))
    else
        print_error "Comprehensive test suite failed"
        ((failed_tests++))
    fi
    ((total_tests++))
    
    # Generate overall coverage report
    print_status "Generating overall coverage report..."
    if command -v go tool cover &> /dev/null; then
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_status "Overall test coverage: $coverage"
        
        # Generate HTML coverage report
        if go tool cover -html=coverage.out -o coverage.html; then
            print_success "HTML coverage report generated: coverage.html"
        fi
    fi
    
    # Print test summary
    print_status "Test Summary"
    print_status "============"
    print_status "Total test phases: $total_tests"
    print_success "Passed: $passed_tests"
    if [ $failed_tests -gt 0 ]; then
        print_error "Failed: $failed_tests"
    else
        print_success "Failed: $failed_tests"
    fi
    
    # Calculate success rate
    local success_rate=$((passed_tests * 100 / total_tests))
    print_status "Success rate: $success_rate%"
    
    # Exit with appropriate code
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests passed! 🎉"
        exit 0
    else
        print_error "Some tests failed. Please review the output above."
        exit 1
    fi
}

# Function to show help
show_help() {
    echo "Business Intelligence API Test Runner"
    echo "====================================="
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -u, --unit     Run only unit tests"
    echo "  -i, --integration  Run only integration tests"
    echo "  -p, --performance  Run only performance tests"
    echo "  -e, --error    Run only error handling tests"
    echo "  -c, --coverage Generate coverage report"
    echo "  -v, --verbose  Verbose output"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 --unit             # Run only unit tests"
    echo "  $0 --coverage         # Run all tests with coverage"
    echo "  $0 --performance      # Run only performance tests"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--unit)
            print_status "Running only unit tests..."
            run_tests_with_coverage "./internal/api/handlers" "Business Intelligence Handler Unit"
            exit $?
            ;;
        -i|--integration)
            print_status "Running only integration tests..."
            run_integration_tests "./test/integration" "Business Intelligence Integration"
            exit $?
            ;;
        -p|--performance)
            print_status "Running only performance tests..."
            run_benchmarks "./test/performance" "Business Intelligence Performance"
            exit $?
            ;;
        -e|--error)
            print_status "Running only error handling tests..."
            run_error_handling_tests "./test/error_handling" "Business Intelligence Error Handling"
            exit $?
            ;;
        -c|--coverage)
            print_status "Running tests with coverage..."
            go test -v -coverprofile=coverage.out -covermode=atomic ./...
            if command -v go tool cover &> /dev/null; then
                coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
                print_status "Test coverage: $coverage"
                go tool cover -html=coverage.out -o coverage.html
                print_success "HTML coverage report generated: coverage.html"
            fi
            exit $?
            ;;
        -v|--verbose)
            set -x
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function if no specific options provided
main
