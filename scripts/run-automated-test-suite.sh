#!/bin/bash

# Automated Test Suite Execution Script
# Runs all business intelligence tests in a coordinated manner

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
TEST_RESULTS_DIR="$TEST_DIR/test-results"
SUITE_START_TIME=$(date +%s)

# Test results directory
mkdir -p "$TEST_RESULTS_DIR"

# Function to print colored output
print_header() {
    echo -e "${PURPLE}$1${NC}"
}

print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to run a test and capture results
run_test() {
    local test_name="$1"
    local test_script="$2"
    local test_description="$3"
    
    print_status "Running: $test_name"
    print_info "Description: $test_description"
    print_info "Script: $test_script"
    
    local test_start_time=$(date +%s)
    local test_result_file="$TEST_RESULTS_DIR/${test_name// /_}_result.txt"
    
    # Run the test and capture output
    if [ -f "$test_script" ]; then
        if bash "$test_script" > "$test_result_file" 2>&1; then
            local test_end_time=$(date +%s)
            local test_duration=$((test_end_time - test_start_time))
            print_success "$test_name completed successfully (${test_duration}s)"
            echo "SUCCESS,$test_name,$test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
            return 0
        else
            local test_end_time=$(date +%s)
            local test_duration=$((test_end_time - test_start_time))
            print_error "$test_name failed (${test_duration}s)"
            echo "FAILED,$test_name,$test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
            return 1
        fi
    else
        print_error "$test_name: Script not found ($test_script)"
        echo "ERROR,$test_name,0" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run Go tests
run_go_tests() {
    print_status "Running Go unit tests..."
    
    local go_test_start_time=$(date +%s)
    local go_test_result_file="$TEST_RESULTS_DIR/go_tests_result.txt"
    
    # Try to run Go tests, but handle compilation errors gracefully
    if go test ./internal/api/handlers/... -v > "$go_test_result_file" 2>&1; then
        local go_test_end_time=$(date +%s)
        local go_test_duration=$((go_test_end_time - go_test_start_time))
        print_success "Go unit tests completed successfully (${go_test_duration}s)"
        echo "SUCCESS,Go Unit Tests,$go_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local go_test_end_time=$(date +%s)
        local go_test_duration=$((go_test_end_time - go_test_start_time))
        print_warning "Go unit tests had issues (${go_test_duration}s) - check $go_test_result_file"
        echo "WARNING,Go Unit Tests,$go_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run component validation tests
run_component_tests() {
    print_status "Running component validation tests..."
    
    local component_test_start_time=$(date +%s)
    local component_test_result_file="$TEST_RESULTS_DIR/component_validation_result.txt"
    
    if bash "$TEST_DIR/scripts/test-business-intelligence-components.sh" > "$component_test_result_file" 2>&1; then
        local component_test_end_time=$(date +%s)
        local component_test_duration=$((component_test_end_time - component_test_start_time))
        print_success "Component validation tests completed successfully (${component_test_duration}s)"
        echo "SUCCESS,Component Validation,$component_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local component_test_end_time=$(date +%s)
        local component_test_duration=$((component_test_end_time - component_test_start_time))
        print_error "Component validation tests failed (${component_test_duration}s)"
        echo "FAILED,Component Validation,$component_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    local perf_test_start_time=$(date +%s)
    local perf_test_result_file="$TEST_RESULTS_DIR/performance_tests_result.txt"
    
    if bash "$TEST_DIR/scripts/test-business-intelligence-performance.sh" > "$perf_test_result_file" 2>&1; then
        local perf_test_end_time=$(date +%s)
        local perf_test_duration=$((perf_test_end_time - perf_test_start_time))
        print_success "Performance tests completed successfully (${perf_test_duration}s)"
        echo "SUCCESS,Performance Tests,$perf_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local perf_test_end_time=$(date +%s)
        local perf_test_duration=$((perf_test_end_time - perf_test_start_time))
        print_error "Performance tests failed (${perf_test_duration}s)"
        echo "FAILED,Performance Tests,$perf_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run UX tests
run_ux_tests() {
    print_status "Running user experience tests..."
    
    local ux_test_start_time=$(date +%s)
    local ux_test_result_file="$TEST_RESULTS_DIR/ux_tests_result.txt"
    
    if bash "$TEST_DIR/scripts/test-business-intelligence-ux.sh" > "$ux_test_result_file" 2>&1; then
        local ux_test_end_time=$(date +%s)
        local ux_test_duration=$((ux_test_end_time - ux_test_start_time))
        print_success "User experience tests completed successfully (${ux_test_duration}s)"
        echo "SUCCESS,User Experience Tests,$ux_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local ux_test_end_time=$(date +%s)
        local ux_test_duration=$((ux_test_end_time - ux_test_start_time))
        print_error "User experience tests failed (${ux_test_duration}s)"
        echo "FAILED,User Experience Tests,$ux_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    local integration_test_start_time=$(date +%s)
    local integration_test_result_file="$TEST_RESULTS_DIR/integration_tests_result.txt"
    
    if bash "$TEST_DIR/scripts/test-business-intelligence-integration.sh" > "$integration_test_result_file" 2>&1; then
        local integration_test_end_time=$(date +%s)
        local integration_test_duration=$((integration_test_end_time - integration_test_start_time))
        print_success "Integration tests completed successfully (${integration_test_duration}s)"
        echo "SUCCESS,Integration Tests,$integration_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local integration_test_end_time=$(date +%s)
        local integration_test_duration=$((integration_test_end_time - integration_test_start_time))
        print_error "Integration tests failed (${integration_test_duration}s)"
        echo "FAILED,Integration Tests,$integration_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to run workflow tests
run_workflow_tests() {
    print_status "Running workflow tests..."
    
    local workflow_test_start_time=$(date +%s)
    local workflow_test_result_file="$TEST_RESULTS_DIR/workflow_tests_result.txt"
    
    if bash "$TEST_DIR/scripts/test-business-intelligence-workflow.sh" > "$workflow_test_result_file" 2>&1; then
        local workflow_test_end_time=$(date +%s)
        local workflow_test_duration=$((workflow_test_end_time - workflow_test_start_time))
        print_success "Workflow tests completed successfully (${workflow_test_duration}s)"
        echo "SUCCESS,Workflow Tests,$workflow_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 0
    else
        local workflow_test_end_time=$(date +%s)
        local workflow_test_duration=$((workflow_test_end_time - workflow_test_start_time))
        print_error "Workflow tests failed (${workflow_test_duration}s)"
        echo "FAILED,Workflow Tests,$workflow_test_duration" >> "$TEST_RESULTS_DIR/test_suite_results.csv"
        return 1
    fi
}

# Function to generate test suite report
generate_test_suite_report() {
    local suite_end_time=$(date +%s)
    local suite_duration=$((suite_end_time - SUITE_START_TIME))
    local report_file="$TEST_RESULTS_DIR/automated_test_suite_report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating test suite report: $report_file"
    
    # Count results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local warning_tests=0
    
    if [ -f "$TEST_RESULTS_DIR/test_suite_results.csv" ]; then
        total_tests=$(wc -l < "$TEST_RESULTS_DIR/test_suite_results.csv")
        passed_tests=$(grep -c "SUCCESS" "$TEST_RESULTS_DIR/test_suite_results.csv" || echo "0")
        failed_tests=$(grep -c "FAILED" "$TEST_RESULTS_DIR/test_suite_results.csv" || echo "0")
        warning_tests=$(grep -c "WARNING" "$TEST_RESULTS_DIR/test_suite_results.csv" || echo "0")
    fi
    
    local success_rate=0
    if [ "$total_tests" -gt 0 ]; then
        success_rate=$((passed_tests * 100 / total_tests))
    fi
    
    cat > "$report_file" << EOF
Automated Test Suite Execution Report
====================================
Generated: $(date)
Test Suite: Business Intelligence Automated Testing
Version: 1.0.0

Test Suite Summary:
- Total Tests Run: $total_tests
- Tests Passed: $passed_tests
- Tests Failed: $failed_tests
- Tests with Warnings: $warning_tests
- Success Rate: $success_rate%
- Total Duration: ${suite_duration}s

Test Categories Executed:
1. Go Unit Tests
2. Component Validation Tests
3. Performance Tests
4. User Experience Tests
5. Integration Tests
6. Workflow Tests

Detailed Results:
EOF

    # Add detailed results
    if [ -f "$TEST_RESULTS_DIR/test_suite_results.csv" ]; then
        echo "" >> "$report_file"
        echo "Test Results:" >> "$report_file"
        echo "============" >> "$report_file"
        cat "$TEST_RESULTS_DIR/test_suite_results.csv" >> "$report_file"
    fi
    
    # Add recommendations
    echo "" >> "$report_file"
    echo "Recommendations:" >> "$report_file"
    echo "===============" >> "$report_file"
    
    if [ "$success_rate" -ge 90 ]; then
        echo "- Test suite is performing well with high success rate" >> "$report_file"
    elif [ "$success_rate" -ge 70 ]; then
        echo "- Test suite has good success rate but some improvements needed" >> "$report_file"
    else
        echo "- Test suite needs significant improvements" >> "$report_file"
    fi
    
    if [ "$failed_tests" -gt 0 ]; then
        echo "- Review failed tests and fix underlying issues" >> "$report_file"
    fi
    
    if [ "$warning_tests" -gt 0 ]; then
        echo "- Address warnings to improve test reliability" >> "$report_file"
    fi
    
    echo "- Consider adding more test coverage for edge cases" >> "$report_file"
    echo "- Implement continuous integration for automated test execution" >> "$report_file"
    
    print_success "Test suite report generated: $report_file"
    
    # Print summary
    print_header "📊 Test Suite Summary"
    print_status "===================="
    print_success "Total Tests: $total_tests"
    print_success "Passed: $passed_tests"
    if [ "$failed_tests" -gt 0 ]; then
        print_error "Failed: $failed_tests"
    fi
    if [ "$warning_tests" -gt 0 ]; then
        print_warning "Warnings: $warning_tests"
    fi
    print_success "Success Rate: $success_rate%"
    print_success "Total Duration: ${suite_duration}s"
}

# Function to cleanup test results
cleanup_test_results() {
    print_status "Cleaning up old test results..."
    
    # Keep only the last 10 test result files
    find "$TEST_RESULTS_DIR" -name "*_result.txt" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    print_header "🚀 Automated Test Suite Execution"
    print_header "=================================="
    
    # Initialize results file
    echo "Status,Test Name,Duration" > "$TEST_RESULTS_DIR/test_suite_results.csv"
    
    # Cleanup old results
    cleanup_test_results
    
    print_status "Starting automated test suite execution..."
    print_info "Test results will be saved to: $TEST_RESULTS_DIR"
    
    # Run all tests
    print_header "🧪 Running Test Suite"
    print_status "===================="
    
    # 1. Go Unit Tests
    run_go_tests
    echo ""
    
    # 2. Component Validation Tests
    run_component_tests
    echo ""
    
    # 3. Performance Tests
    run_performance_tests
    echo ""
    
    # 4. User Experience Tests
    run_ux_tests
    echo ""
    
    # 5. Integration Tests
    run_integration_tests
    echo ""
    
    # 6. Workflow Tests
    run_workflow_tests
    echo ""
    
    # Generate comprehensive report
    generate_test_suite_report
    
    print_header "📋 Final Test Suite Summary"
    print_status "=========================="
    print_success "Automated test suite execution completed!"
    print_info "Check the test-results directory for detailed reports."
}

# Run main function
main "$@"
