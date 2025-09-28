#!/bin/bash

# Phase E Testing & QA Execution Script
# This script runs comprehensive testing for the KYB platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="$PROJECT_ROOT/test"
REPORTS_DIR="$PROJECT_ROOT/test-reports"
LOG_DIR="$PROJECT_ROOT/logs"

# Environment configuration
ENVIRONMENT=${ENVIRONMENT:-"production"}
BASE_URL=${BASE_URL:-"https://kyb-api-gateway-production.up.railway.app"}
API_KEY=${API_KEY:-"test-api-key"}

# Test configuration
PARALLEL_TESTS=${PARALLEL_TESTS:-"true"}
MAX_CONCURRENCY=${MAX_CONCURRENCY:-"4"}
TEST_TIMEOUT=${TEST_TIMEOUT:-"30m"}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-"80"}

# Test suites to run
RUN_UNIT_TESTS=${RUN_UNIT_TESTS:-"true"}
RUN_INTEGRATION_TESTS=${RUN_INTEGRATION_TESTS:-"true"}
RUN_E2E_TESTS=${RUN_E2E_TESTS:-"true"}
RUN_PERFORMANCE_TESTS=${RUN_PERFORMANCE_TESTS:-"false"}
RUN_SECURITY_TESTS=${RUN_SECURITY_TESTS:-"true"}

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] ${message}${NC}"
}

print_info() {
    print_status "$BLUE" "INFO: $1"
}

print_success() {
    print_status "$GREEN" "SUCCESS: $1"
}

print_warning() {
    print_status "$YELLOW" "WARNING: $1"
}

print_error() {
    print_status "$RED" "ERROR: $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.22 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.22"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is not supported. Please install Go $REQUIRED_VERSION or later."
        exit 1
    fi
    
    # Check if required tools are installed
    local missing_tools=()
    
    if [ "$RUN_PERFORMANCE_TESTS" = "true" ] && ! command -v k6 &> /dev/null; then
        missing_tools+=("k6")
    fi
    
    if [ "$RUN_SECURITY_TESTS" = "true" ] && ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_warning "Missing tools: ${missing_tools[*]}"
        print_info "Some tests may be skipped."
    fi
    
    print_success "Prerequisites check completed"
}

# Function to setup test environment
setup_test_environment() {
    print_info "Setting up test environment..."
    
    # Create directories
    mkdir -p "$REPORTS_DIR"
    mkdir -p "$LOG_DIR"
    
    # Set environment variables
    export TEST_BASE_URL="$BASE_URL"
    export TEST_API_GATEWAY_URL="$BASE_URL"
    export TEST_CLASSIFICATION_URL="https://kyb-classification-service-production.up.railway.app"
    export TEST_MERCHANT_URL="https://kyb-merchant-service-production.up.railway.app"
    export TEST_MONITORING_URL="https://kyb-monitoring-production.up.railway.app"
    export TEST_PIPELINE_URL="https://kyb-pipeline-service-production.up.railway.app"
    export TEST_FRONTEND_URL="https://kyb-frontend-production.up.railway.app"
    export TEST_BI_URL="https://kyb-business-intelligence-gateway-production.up.railway.app"
    
    export SEC_TEST_API_KEY="$API_KEY"
    export SEC_ADMIN_API_KEY="${ADMIN_API_KEY:-admin-api-key}"
    export SEC_USER_API_KEY="${USER_API_KEY:-user-api-key}"
    
    export PERF_TEST_BASE_URL="$BASE_URL"
    export K6_BASE_URL="$BASE_URL"
    export K6_API_KEY="$API_KEY"
    
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Download dependencies
    print_info "Downloading Go dependencies..."
    go mod download
    
    print_success "Test environment setup completed"
}

# Function to run unit tests
run_unit_tests() {
    if [ "$RUN_UNIT_TESTS" != "true" ]; then
        print_info "Skipping unit tests"
        return 0
    fi
    
    print_info "Running unit tests..."
    
    local start_time=$(date +%s)
    
    # Run unit tests with coverage
    go test -v -cover -coverprofile="$REPORTS_DIR/coverage.out" -timeout="$TEST_TIMEOUT" ./... 2>&1 | tee "$LOG_DIR/unit-tests.log"
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Unit tests completed successfully in ${duration}s"
        
        # Generate coverage report
        go tool cover -html="$REPORTS_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"
        
        # Check coverage threshold
        local coverage=$(go tool cover -func="$REPORTS_DIR/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage < $COVERAGE_THRESHOLD" | bc -l) )); then
            print_warning "Coverage $coverage% is below threshold $COVERAGE_THRESHOLD%"
        else
            print_success "Coverage $coverage% meets threshold $COVERAGE_THRESHOLD%"
        fi
    else
        print_error "Unit tests failed"
        return $exit_code
    fi
}

# Function to run integration tests
run_integration_tests() {
    if [ "$RUN_INTEGRATION_TESTS" != "true" ]; then
        print_info "Skipping integration tests"
        return 0
    fi
    
    print_info "Running integration tests..."
    
    local start_time=$(date +%s)
    
    # Run integration tests
    go test -v -tags=integration -timeout="$TEST_TIMEOUT" ./test/integration/... 2>&1 | tee "$LOG_DIR/integration-tests.log"
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Integration tests completed successfully in ${duration}s"
    else
        print_error "Integration tests failed"
        return $exit_code
    fi
}

# Function to run E2E tests
run_e2e_tests() {
    if [ "$RUN_E2E_TESTS" != "true" ]; then
        print_info "Skipping E2E tests"
        return 0
    fi
    
    print_info "Running E2E tests..."
    
    local start_time=$(date +%s)
    
    # Run E2E tests
    go test -v -tags=e2e -timeout="$TEST_TIMEOUT" ./test/e2e/... 2>&1 | tee "$LOG_DIR/e2e-tests.log"
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "E2E tests completed successfully in ${duration}s"
    else
        print_error "E2E tests failed"
        return $exit_code
    fi
}

# Function to run performance tests
run_performance_tests() {
    if [ "$RUN_PERFORMANCE_TESTS" != "true" ]; then
        print_info "Skipping performance tests"
        return 0
    fi
    
    print_info "Running performance tests..."
    
    local start_time=$(date +%s)
    
    # Run performance tests
    go test -v -tags=performance -bench=. -benchmem -timeout=1h ./test/performance/... 2>&1 | tee "$LOG_DIR/performance-tests.log"
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Performance tests completed successfully in ${duration}s"
    else
        print_error "Performance tests failed"
        return $exit_code
    fi
    
    # Run k6 load tests if available
    if command -v k6 &> /dev/null; then
        print_info "Running k6 load tests..."
        k6 run --out json="$REPORTS_DIR/load-test-results.json" ./test/load/load-test.js 2>&1 | tee "$LOG_DIR/load-tests.log"
    else
        print_warning "k6 not available, skipping load tests"
    fi
}

# Function to run security tests
run_security_tests() {
    if [ "$RUN_SECURITY_TESTS" != "true" ]; then
        print_info "Skipping security tests"
        return 0
    fi
    
    print_info "Running security tests..."
    
    local start_time=$(date +%s)
    
    # Run security tests
    go test -v -tags=security -timeout="$TEST_TIMEOUT" ./test/security/... 2>&1 | tee "$LOG_DIR/security-tests.log"
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Security tests completed successfully in ${duration}s"
    else
        print_error "Security tests failed"
        return $exit_code
    fi
}

# Function to generate test report
generate_test_report() {
    print_info "Generating comprehensive test report..."
    
    # Create test results summary
    cat > "$REPORTS_DIR/test-summary.md" << EOF
# Phase E Testing & QA Results

**Date**: $(date)
**Environment**: $ENVIRONMENT
**Base URL**: $BASE_URL

## Test Execution Summary

| Test Suite | Status | Duration | Coverage |
|------------|--------|----------|----------|
| Unit Tests | $([ "$RUN_UNIT_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped") | - | - |
| Integration Tests | $([ "$RUN_INTEGRATION_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped") | - | - |
| E2E Tests | $([ "$RUN_E2E_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped") | - | - |
| Performance Tests | $([ "$RUN_PERFORMANCE_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped") | - | - |
| Security Tests | $([ "$RUN_SECURITY_TESTS" = "true" ] && echo "✅ Completed" || echo "⏭️ Skipped") | - | - |

## Test Artifacts

- Coverage Report: [coverage.html](coverage.html)
- Unit Test Logs: [unit-tests.log](../../logs/unit-tests.log)
- Integration Test Logs: [integration-tests.log](../../logs/integration-tests.log)
- E2E Test Logs: [e2e-tests.log](../../logs/e2e-tests.log)
- Performance Test Logs: [performance-tests.log](../../logs/performance-tests.log)
- Security Test Logs: [security-tests.log](../../logs/security-tests.log)

## Recommendations

1. Review any failed tests and fix issues
2. Ensure test coverage meets the $COVERAGE_THRESHOLD% threshold
3. Address any security vulnerabilities found
4. Optimize performance based on test results

EOF
    
    print_success "Test report generated: $REPORTS_DIR/test-summary.md"
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up test environment..."
    
    # Remove temporary files if any
    find "$PROJECT_ROOT" -name "*.tmp" -delete 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Phase E Testing & QA Execution Script

OPTIONS:
    -e, --environment ENV    Test environment (default: production)
    -u, --base-url URL       Base URL for testing (default: https://kyb-api-gateway-production.up.railway.app)
    -k, --api-key KEY        API key for authentication (default: test-api-key)
    -t, --timeout TIMEOUT    Test timeout (default: 30m)
    -c, --coverage THRESHOLD Coverage threshold percentage (default: 80)
    --skip-unit              Skip unit tests
    --skip-integration       Skip integration tests
    --skip-e2e               Skip E2E tests
    --skip-performance       Skip performance tests
    --skip-security          Skip security tests
    --run-performance        Run performance tests (default: false)
    -h, --help               Show this help message

EXAMPLES:
    $0                                    # Run all tests with defaults
    $0 --environment staging              # Run tests against staging
    $0 --skip-performance --skip-e2e      # Skip performance and E2E tests
    $0 --run-performance                  # Include performance tests
    $0 --coverage 90                      # Set coverage threshold to 90%

ENVIRONMENT VARIABLES:
    ENVIRONMENT              Test environment
    BASE_URL                 Base URL for testing
    API_KEY                  API key for authentication
    ADMIN_API_KEY            Admin API key for security tests
    USER_API_KEY             User API key for security tests
    RUN_UNIT_TESTS           Run unit tests (true/false)
    RUN_INTEGRATION_TESTS    Run integration tests (true/false)
    RUN_E2E_TESTS            Run E2E tests (true/false)
    RUN_PERFORMANCE_TESTS    Run performance tests (true/false)
    RUN_SECURITY_TESTS       Run security tests (true/false)

EOF
}

# Main execution function
main() {
    local start_time=$(date +%s)
    
    print_info "Starting Phase E Testing & QA execution"
    print_info "Environment: $ENVIRONMENT"
    print_info "Base URL: $BASE_URL"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -u|--base-url)
                BASE_URL="$2"
                shift 2
                ;;
            -k|--api-key)
                API_KEY="$2"
                shift 2
                ;;
            -t|--timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            -c|--coverage)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --skip-unit)
                RUN_UNIT_TESTS="false"
                shift
                ;;
            --skip-integration)
                RUN_INTEGRATION_TESTS="false"
                shift
                ;;
            --skip-e2e)
                RUN_E2E_TESTS="false"
                shift
                ;;
            --skip-performance)
                RUN_PERFORMANCE_TESTS="false"
                shift
                ;;
            --skip-security)
                RUN_SECURITY_TESTS="false"
                shift
                ;;
            --run-performance)
                RUN_PERFORMANCE_TESTS="true"
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Execute test phases
    check_prerequisites
    setup_test_environment
    
    local failed_tests=0
    
    # Run test suites
    run_unit_tests || ((failed_tests++))
    run_integration_tests || ((failed_tests++))
    run_e2e_tests || ((failed_tests++))
    run_performance_tests || ((failed_tests++))
    run_security_tests || ((failed_tests++))
    
    # Generate report
    generate_test_report
    
    # Cleanup
    cleanup
    
    # Final results
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests completed successfully in ${total_duration}s"
        print_info "Test reports available in: $REPORTS_DIR"
        exit 0
    else
        print_error "$failed_tests test suite(s) failed"
        print_info "Check logs in: $LOG_DIR"
        print_info "Test reports available in: $REPORTS_DIR"
        exit 1
    fi
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
