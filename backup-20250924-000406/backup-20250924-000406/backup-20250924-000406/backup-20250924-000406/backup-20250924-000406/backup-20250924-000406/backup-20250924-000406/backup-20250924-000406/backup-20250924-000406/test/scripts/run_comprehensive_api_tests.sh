#!/bin/bash

# Comprehensive API Endpoint Testing Script
# This script runs comprehensive API endpoint tests as specified in subtask 4.2.1

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$PROJECT_ROOT/test"
REPORTS_DIR="$TEST_DIR/reports"
COVERAGE_DIR="$TEST_DIR/coverage"

# Test configuration
INTEGRATION_TESTS=${INTEGRATION_TESTS:-"true"}
PERFORMANCE_TESTS=${PERFORMANCE_TESTS:-"false"}
VERBOSE=${VERBOSE:-"false"}
COVERAGE=${COVERAGE:-"true"}

# Create reports directory
mkdir -p "$REPORTS_DIR"
mkdir -p "$COVERAGE_DIR"

echo -e "${BLUE}🚀 Starting Comprehensive API Endpoint Testing${NC}"
echo -e "${BLUE}================================================${NC}"
echo "Project Root: $PROJECT_ROOT"
echo "Test Directory: $TEST_DIR"
echo "Reports Directory: $REPORTS_DIR"
echo "Integration Tests: $INTEGRATION_TESTS"
echo "Performance Tests: $PERFORMANCE_TESTS"
echo "Coverage: $COVERAGE"
echo ""

# Function to run tests with proper environment
run_tests() {
    local test_type="$1"
    local test_pattern="$2"
    local description="$3"
    
    echo -e "${YELLOW}📋 Running $description${NC}"
    echo "Test Pattern: $test_pattern"
    echo "Test Type: $test_type"
    echo ""
    
    # Set environment variables
    export INTEGRATION_TESTS="$INTEGRATION_TESTS"
    export PERFORMANCE_TESTS="$PERFORMANCE_TESTS"
    export GO_TEST_TIMEOUT="10m"
    
    # Build test command
    local test_cmd="go test -v -timeout=10m"
    
    # Add coverage if requested
    if [ "$COVERAGE" = "true" ]; then
        test_cmd="$test_cmd -coverprofile=$COVERAGE_DIR/${test_type}_coverage.out -covermode=atomic"
    fi
    
    # Add test pattern
    test_cmd="$test_cmd -run $test_pattern"
    
    # Add test directory
    test_cmd="$test_cmd $TEST_DIR/integration/..."
    
    # Run tests and capture output
    local output_file="$REPORTS_DIR/${test_type}_test_results.txt"
    local start_time=$(date +%s)
    
    echo "Executing: $test_cmd"
    echo "Output: $output_file"
    echo ""
    
    if $test_cmd > "$output_file" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        echo -e "${GREEN}✅ $description completed successfully in ${duration}s${NC}"
        
        # Extract test statistics
        local passed=$(grep -c "PASS:" "$output_file" || echo "0")
        local failed=$(grep -c "FAIL:" "$output_file" || echo "0")
        local skipped=$(grep -c "SKIP:" "$output_file" || echo "0")
        
        echo "  📊 Results: $passed passed, $failed failed, $skipped skipped"
        
        # Generate summary
        echo "TEST_TYPE=$test_type" >> "$REPORTS_DIR/test_summary.txt"
        echo "DESCRIPTION=$description" >> "$REPORTS_DIR/test_summary.txt"
        echo "DURATION=${duration}s" >> "$REPORTS_DIR/test_summary.txt"
        echo "PASSED=$passed" >> "$REPORTS_DIR/test_summary.txt"
        echo "FAILED=$failed" >> "$REPORTS_DIR/test_summary.txt"
        echo "SKIPPED=$skipped" >> "$REPORTS_DIR/test_summary.txt"
        echo "STATUS=SUCCESS" >> "$REPORTS_DIR/test_summary.txt"
        echo "---" >> "$REPORTS_DIR/test_summary.txt"
        
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        echo -e "${RED}❌ $description failed after ${duration}s${NC}"
        
        # Extract test statistics
        local passed=$(grep -c "PASS:" "$output_file" || echo "0")
        local failed=$(grep -c "FAIL:" "$output_file" || echo "0")
        local skipped=$(grep -c "SKIP:" "$output_file" || echo "0")
        
        echo "  📊 Results: $passed passed, $failed failed, $skipped skipped"
        
        # Generate summary
        echo "TEST_TYPE=$test_type" >> "$REPORTS_DIR/test_summary.txt"
        echo "DESCRIPTION=$description" >> "$REPORTS_DIR/test_summary.txt"
        echo "DURATION=${duration}s" >> "$REPORTS_DIR/test_summary.txt"
        echo "PASSED=$passed" >> "$REPORTS_DIR/test_summary.txt"
        echo "FAILED=$failed" >> "$REPORTS_DIR/test_summary.txt"
        echo "SKIPPED=$skipped" >> "$REPORTS_DIR/test_summary.txt"
        echo "STATUS=FAILED" >> "$REPORTS_DIR/test_summary.txt"
        echo "---" >> "$REPORTS_DIR/test_summary.txt"
        
        # Show last few lines of output for debugging
        echo -e "${RED}Last 10 lines of output:${NC}"
        tail -10 "$output_file"
        echo ""
    fi
    
    echo ""
}

# Function to generate coverage report
generate_coverage_report() {
    if [ "$COVERAGE" = "true" ]; then
        echo -e "${YELLOW}📊 Generating Coverage Report${NC}"
        
        # Combine coverage files
        echo "mode: atomic" > "$COVERAGE_DIR/combined_coverage.out"
        find "$COVERAGE_DIR" -name "*_coverage.out" -exec grep -h -v "mode: atomic" {} \; >> "$COVERAGE_DIR/combined_coverage.out" 2>/dev/null || true
        
        # Generate HTML coverage report
        if command -v go &> /dev/null; then
            go tool cover -html="$COVERAGE_DIR/combined_coverage.out" -o "$COVERAGE_DIR/coverage.html"
            echo -e "${GREEN}✅ Coverage report generated: $COVERAGE_DIR/coverage.html${NC}"
        fi
        
        # Generate text coverage report
        if [ -f "$COVERAGE_DIR/combined_coverage.out" ]; then
            go tool cover -func="$COVERAGE_DIR/combined_coverage.out" > "$COVERAGE_DIR/coverage.txt"
            echo -e "${GREEN}✅ Coverage summary generated: $COVERAGE_DIR/coverage.txt${NC}"
        fi
        
        echo ""
    fi
}

# Function to generate test report
generate_test_report() {
    echo -e "${YELLOW}📋 Generating Test Report${NC}"
    
    local report_file="$REPORTS_DIR/comprehensive_api_test_report.md"
    
    cat > "$report_file" << EOF
# Comprehensive API Endpoint Test Report

**Generated**: $(date)
**Test Suite**: Subtask 4.2.1 - API Endpoint Testing
**Project**: KYB Platform Supabase Table Improvement Implementation

## Test Summary

EOF

    # Add test results from summary file
    if [ -f "$REPORTS_DIR/test_summary.txt" ]; then
        echo "| Test Type | Description | Duration | Passed | Failed | Skipped | Status |" >> "$report_file"
        echo "|-----------|-------------|----------|--------|--------|---------|--------|" >> "$report_file"
        
        # Parse summary file and add to report
        while IFS= read -r line; do
            if [[ $line == "TEST_TYPE="* ]]; then
                test_type=$(echo "$line" | cut -d'=' -f2)
            elif [[ $line == "DESCRIPTION="* ]]; then
                description=$(echo "$line" | cut -d'=' -f2-)
            elif [[ $line == "DURATION="* ]]; then
                duration=$(echo "$line" | cut -d'=' -f2)
            elif [[ $line == "PASSED="* ]]; then
                passed=$(echo "$line" | cut -d'=' -f2)
            elif [[ $line == "FAILED="* ]]; then
                failed=$(echo "$line" | cut -d'=' -f2)
            elif [[ $line == "SKIPPED="* ]]; then
                skipped=$(echo "$line" | cut -d'=' -f2)
            elif [[ $line == "STATUS="* ]]; then
                status=$(echo "$line" | cut -d'=' -f2)
                echo "| $test_type | $description | $duration | $passed | $failed | $skipped | $status |" >> "$report_file"
            fi
        done < "$REPORTS_DIR/test_summary.txt"
    fi

    cat >> "$report_file" << EOF

## Test Categories

### 1. Business-Related Endpoints
- **Classification Endpoints**: Single and batch business classification
- **Merchant Management**: CRUD operations for merchant entities
- **Risk Assessment**: Business risk evaluation and assessment

### 2. Classification Endpoints
- **Enhanced Classification**: ML-powered classification with v2 endpoints
- **Classification Monitoring**: Accuracy tracking and misclassification detection
- **Performance Analytics**: Classification performance metrics

### 3. User Management Endpoints
- **Authentication**: User registration, login, logout, token refresh
- **Profile Management**: User profile CRUD operations
- **API Key Management**: API key creation, retrieval, and deletion

### 4. Monitoring Endpoints
- **Health Checks**: System health and status monitoring
- **Metrics**: Performance and system metrics collection
- **Compliance**: Compliance checking and reporting
- **Alerting**: Monitoring alerts and notifications

## Test Results

EOF

    # Add individual test results
    for result_file in "$REPORTS_DIR"/*_test_results.txt; do
        if [ -f "$result_file" ]; then
            test_name=$(basename "$result_file" _test_results.txt)
            echo "### $test_name" >> "$report_file"
            echo '```' >> "$report_file"
            head -50 "$result_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo "" >> "$report_file"
        fi
    done

    cat >> "$report_file" << EOF

## Coverage Report

EOF

    if [ -f "$COVERAGE_DIR/coverage.txt" ]; then
        echo '```' >> "$report_file"
        cat "$COVERAGE_DIR/coverage.txt" >> "$report_file"
        echo '```' >> "$report_file"
    fi

    cat >> "$report_file" << EOF

## Recommendations

Based on the test results, the following recommendations are made:

1. **Performance Optimization**: Focus on endpoints with response times > 2 seconds
2. **Error Handling**: Ensure all endpoints have proper error handling and validation
3. **Security**: Verify authentication and authorization on all protected endpoints
4. **Monitoring**: Implement comprehensive monitoring for all API endpoints
5. **Documentation**: Ensure all endpoints are properly documented

## Next Steps

1. Review test results and address any failures
2. Implement performance optimizations for slow endpoints
3. Enhance error handling and validation
4. Update API documentation based on test findings
5. Implement continuous monitoring for API endpoints

EOF

    echo -e "${GREEN}✅ Test report generated: $report_file${NC}"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}🎯 Subtask 4.2.1: API Endpoint Testing${NC}"
    echo "This script tests all API endpoints as specified in the implementation plan."
    echo ""
    
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Run comprehensive API endpoint tests
    run_tests "comprehensive" "TestComprehensiveAPIEndpoints" "Comprehensive API Endpoint Tests"
    
    # Run performance tests if enabled
    if [ "$PERFORMANCE_TESTS" = "true" ]; then
        run_tests "performance" "TestAPIEndpointPerformance" "API Endpoint Performance Tests"
    fi
    
    # Run error handling tests
    run_tests "error_handling" "TestAPIEndpointErrorHandling" "API Endpoint Error Handling Tests"
    
    # Generate coverage report
    generate_coverage_report
    
    # Generate test report
    generate_test_report
    
    # Final summary
    echo -e "${BLUE}📊 Test Execution Summary${NC}"
    echo "=================================="
    
    if [ -f "$REPORTS_DIR/test_summary.txt" ]; then
        local total_tests=0
        local total_passed=0
        local total_failed=0
        local total_skipped=0
        
        while IFS= read -r line; do
            if [[ $line == "PASSED="* ]]; then
                passed=$(echo "$line" | cut -d'=' -f2)
                total_passed=$((total_passed + passed))
                total_tests=$((total_tests + passed))
            elif [[ $line == "FAILED="* ]]; then
                failed=$(echo "$line" | cut -d'=' -f2)
                total_failed=$((total_failed + failed))
                total_tests=$((total_tests + failed))
            elif [[ $line == "SKIPPED="* ]]; then
                skipped=$(echo "$line" | cut -d'=' -f2)
                total_skipped=$((total_skipped + skipped))
            fi
        done < "$REPORTS_DIR/test_summary.txt"
        
        echo "Total Tests: $total_tests"
        echo "Passed: $total_passed"
        echo "Failed: $total_failed"
        echo "Skipped: $total_skipped"
        echo ""
        
        if [ $total_failed -eq 0 ]; then
            echo -e "${GREEN}🎉 All tests passed successfully!${NC}"
            exit 0
        else
            echo -e "${RED}❌ Some tests failed. Please review the results.${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ No test summary found. Tests may have failed to run.${NC}"
        exit 1
    fi
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --integration-only)
            PERFORMANCE_TESTS="false"
            shift
            ;;
        --performance-only)
            INTEGRATION_TESTS="false"
            PERFORMANCE_TESTS="true"
            shift
            ;;
        --no-coverage)
            COVERAGE="false"
            shift
            ;;
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --integration-only    Run only integration tests"
            echo "  --performance-only    Run only performance tests"
            echo "  --no-coverage         Disable coverage reporting"
            echo "  --verbose             Enable verbose output"
            echo "  --help                Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  INTEGRATION_TESTS     Enable integration tests (default: true)"
            echo "  PERFORMANCE_TESTS     Enable performance tests (default: false)"
            echo "  COVERAGE              Enable coverage reporting (default: true)"
            echo "  VERBOSE               Enable verbose output (default: false)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main
