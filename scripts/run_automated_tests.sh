#!/bin/bash

# KYB Platform - Automated Test Runner
# This script runs comprehensive tests including unit, integration, performance, and E2E tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_TIMEOUT="30m"
COVERAGE_THRESHOLD=90
PARALLEL_JOBS=4

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test reports
TEST_REPORT_DIR="$PROJECT_ROOT/test/reports"
COVERAGE_REPORT_DIR="$PROJECT_ROOT/test/coverage"
PERFORMANCE_REPORT_DIR="$PROJECT_ROOT/test/performance"

# Create report directories
mkdir -p "$TEST_REPORT_DIR" "$COVERAGE_REPORT_DIR" "$PERFORMANCE_REPORT_DIR"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test result tracking
record_test_result() {
    local result=$1
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    case $result in
        "PASS")
            PASSED_TESTS=$((PASSED_TESTS + 1))
            ;;
        "FAIL")
            FAILED_TESTS=$((FAILED_TESTS + 1))
            ;;
        "SKIP")
            SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
            ;;
    esac
}

# Print test summary
print_test_summary() {
    echo
    echo "=========================================="
    echo "           TEST EXECUTION SUMMARY"
    echo "=========================================="
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Skipped: $SKIPPED_TESTS"
    echo "Success Rate: $((PASSED_TESTS * 100 / TOTAL_TESTS))%"
    echo "=========================================="
    echo
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Check if required tools are installed
    local required_tools=("docker" "docker-compose")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_warning "$tool is not installed - some tests may be skipped"
        fi
    done
    
    log_success "Prerequisites check completed"
}

# Setup test environment
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Set test environment variables
    export TEST_ENV="automated"
    export TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
    export TEST_DB_PORT="${TEST_DB_PORT:-5432}"
    export TEST_DB_USER="${TEST_DB_USER:-test_user}"
    export TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-test_password}"
    export TEST_DB_NAME="${TEST_DB_NAME:-kyb_platform_test}"
    export TEST_REDIS_HOST="${TEST_REDIS_HOST:-localhost}"
    export TEST_REDIS_PORT="${TEST_REDIS_PORT:-6379}"
    
    # Create test configuration
    cat > "$PROJECT_ROOT/test_config.yaml" << EOF
test:
  database:
    host: $TEST_DB_HOST
    port: $TEST_DB_PORT
    username: $TEST_DB_USER
    password: $TEST_DB_PASSWORD
    database: $TEST_DB_NAME
  redis:
    host: $TEST_REDIS_HOST
    port: $TEST_REDIS_PORT
  coverage:
    thresholds:
      statements: $COVERAGE_THRESHOLD
      branches: $((COVERAGE_THRESHOLD - 10))
      functions: $COVERAGE_THRESHOLD
      lines: $COVERAGE_THRESHOLD
EOF
    
    log_success "Test environment setup completed"
}

# Start test services
start_test_services() {
    log_info "Starting test services..."
    
    if command -v docker-compose &> /dev/null; then
        cd "$PROJECT_ROOT"
        docker-compose -f docker-compose.test.yml up -d postgres-test redis-test
        
        # Wait for services to be ready
        log_info "Waiting for test services to be ready..."
        sleep 30
        
        # Verify services are running
        if docker-compose -f docker-compose.test.yml ps | grep -q "Up"; then
            log_success "Test services started successfully"
        else
            log_error "Failed to start test services"
            return 1
        fi
    else
        log_warning "docker-compose not available - skipping service startup"
    fi
}

# Stop test services
stop_test_services() {
    log_info "Stopping test services..."
    
    if command -v docker-compose &> /dev/null; then
        cd "$PROJECT_ROOT"
        docker-compose -f docker-compose.test.yml down
        log_success "Test services stopped"
    fi
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."
    
    local start_time=$(date +%s)
    local test_output="$TEST_REPORT_DIR/unit_tests.log"
    local coverage_output="$COVERAGE_REPORT_DIR/unit_coverage.out"
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests with coverage
    if go test -v -race -coverprofile="$coverage_output" -covermode=atomic \
        -timeout="$TEST_TIMEOUT" \
        -coverpkg=./... \
        ./internal/... ./pkg/... 2>&1 | tee "$test_output"; then
        
        # Generate coverage report
        go tool cover -html="$coverage_output" -o "$COVERAGE_REPORT_DIR/unit_coverage.html"
        go tool cover -func="$coverage_output" > "$COVERAGE_REPORT_DIR/unit_coverage.txt"
        
        # Check coverage threshold
        local coverage=$(go tool cover -func="$coverage_output" | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
            log_success "Unit tests passed with coverage: ${coverage}%"
            record_test_result "PASS"
        else
            log_error "Unit tests coverage ${coverage}% is below threshold ${COVERAGE_THRESHOLD}%"
            record_test_result "FAIL"
            return 1
        fi
    else
        log_error "Unit tests failed"
        record_test_result "FAIL"
        return 1
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "Unit tests completed in ${duration}s"
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    local start_time=$(date +%s)
    local test_output="$TEST_REPORT_DIR/integration_tests.log"
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests
    if go test -v -tags=integration \
        -timeout="$TEST_TIMEOUT" \
        ./test/integration/... 2>&1 | tee "$test_output"; then
        
        log_success "Integration tests passed"
        record_test_result "PASS"
    else
        log_error "Integration tests failed"
        record_test_result "FAIL"
        return 1
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "Integration tests completed in ${duration}s"
}

# Run performance tests
run_performance_tests() {
    log_info "Running performance tests..."
    
    local start_time=$(date +%s)
    local test_output="$PERFORMANCE_REPORT_DIR/performance_tests.log"
    local benchmark_output="$PERFORMANCE_REPORT_DIR/benchmarks.log"
    
    cd "$PROJECT_ROOT"
    
    # Run benchmarks
    if go test -v -bench=. -benchmem \
        -timeout="$TEST_TIMEOUT" \
        ./internal/... ./pkg/... 2>&1 | tee "$benchmark_output"; then
        
        log_success "Performance benchmarks completed"
        record_test_result "PASS"
    else
        log_warning "Performance benchmarks failed"
        record_test_result "SKIP"
    fi
    
    # Run performance tests
    if go test -v -tags=performance \
        -timeout="$TEST_TIMEOUT" \
        ./test/performance/... 2>&1 | tee "$test_output"; then
        
        log_success "Performance tests passed"
        record_test_result "PASS"
    else
        log_warning "Performance tests failed"
        record_test_result "SKIP"
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "Performance tests completed in ${duration}s"
}

# Run end-to-end tests
run_e2e_tests() {
    log_info "Running end-to-end tests..."
    
    local start_time=$(date +%s)
    local test_output="$TEST_REPORT_DIR/e2e_tests.log"
    
    cd "$PROJECT_ROOT"
    
    # Check if E2E tests exist
    if [ ! -d "./test/e2e" ]; then
        log_warning "E2E test directory not found - skipping"
        record_test_result "SKIP"
        return 0
    fi
    
    # Run E2E tests
    if go test -v -tags=e2e \
        -timeout="$TEST_TIMEOUT" \
        ./test/e2e/... 2>&1 | tee "$test_output"; then
        
        log_success "E2E tests passed"
        record_test_result "PASS"
    else
        log_error "E2E tests failed"
        record_test_result "FAIL"
        return 1
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "E2E tests completed in ${duration}s"
}

# Run security tests
run_security_tests() {
    log_info "Running security tests..."
    
    local start_time=$(date +%s)
    local test_output="$TEST_REPORT_DIR/security_tests.log"
    
    cd "$PROJECT_ROOT"
    
    # Run security tests
    if go test -v -tags=security \
        -timeout="$TEST_TIMEOUT" \
        ./test/security/... 2>&1 | tee "$test_output"; then
        
        log_success "Security tests passed"
        record_test_result "PASS"
    else
        log_warning "Security tests failed"
        record_test_result "SKIP"
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "Security tests completed in ${duration}s"
}

# Generate test report
generate_test_report() {
    log_info "Generating test report..."
    
    local report_file="$TEST_REPORT_DIR/test_report.html"
    local coverage_file="$COVERAGE_REPORT_DIR/unit_coverage.txt"
    
    # Create HTML report
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>KYB Platform Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .test-section { margin: 20px 0; padding: 10px; border: 1px solid #ddd; border-radius: 5px; }
        .pass { color: green; }
        .fail { color: red; }
        .skip { color: orange; }
        .coverage { background-color: #f9f9f9; padding: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>KYB Platform Test Report</h1>
        <p>Generated: $(date)</p>
    </div>
    
    <div class="summary">
        <h2>Test Summary</h2>
        <p>Total Tests: $TOTAL_TESTS</p>
        <p class="pass">Passed: $PASSED_TESTS</p>
        <p class="fail">Failed: $FAILED_TESTS</p>
        <p class="skip">Skipped: $SKIPPED_TESTS</p>
        <p>Success Rate: $((PASSED_TESTS * 100 / TOTAL_TESTS))%</p>
    </div>
    
    <div class="test-section">
        <h3>Coverage Report</h3>
        <div class="coverage">
EOF
    
    if [ -f "$coverage_file" ]; then
        cat "$coverage_file" >> "$report_file"
    else
        echo "Coverage report not available" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF
        </div>
    </div>
    
    <div class="test-section">
        <h3>Test Logs</h3>
        <ul>
            <li><a href="unit_tests.log">Unit Tests Log</a></li>
            <li><a href="integration_tests.log">Integration Tests Log</a></li>
            <li><a href="e2e_tests.log">E2E Tests Log</a></li>
            <li><a href="security_tests.log">Security Tests Log</a></li>
        </ul>
    </div>
    
    <div class="test-section">
        <h3>Coverage Reports</h3>
        <ul>
            <li><a href="../coverage/unit_coverage.html">Unit Tests Coverage</a></li>
        </ul>
    </div>
    
    <div class="test-section">
        <h3>Performance Reports</h3>
        <ul>
            <li><a href="../performance/benchmarks.log">Benchmarks</a></li>
            <li><a href="../performance/performance_tests.log">Performance Tests</a></li>
        </ul>
    </div>
</body>
</html>
EOF
    
    log_success "Test report generated: $report_file"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    stop_test_services
    
    # Remove temporary files
    rm -f "$PROJECT_ROOT/test_config.yaml"
    
    log_success "Cleanup completed"
}

# Main execution
main() {
    local test_types=("unit" "integration" "performance" "e2e" "security")
    local run_all=true
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit)
                test_types=("unit")
                run_all=false
                shift
                ;;
            --integration)
                test_types=("integration")
                run_all=false
                shift
                ;;
            --performance)
                test_types=("performance")
                run_all=false
                shift
                ;;
            --e2e)
                test_types=("e2e")
                run_all=false
                shift
                ;;
            --security)
                test_types=("security")
                run_all=false
                shift
                ;;
            --coverage-threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --unit                 Run only unit tests"
                echo "  --integration          Run only integration tests"
                echo "  --performance          Run only performance tests"
                echo "  --e2e                  Run only E2E tests"
                echo "  --security             Run only security tests"
                echo "  --coverage-threshold   Set coverage threshold (default: 90)"
                echo "  --timeout              Set test timeout (default: 30m)"
                echo "  --help                 Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    echo "=========================================="
    echo "      KYB Platform Automated Testing"
    echo "=========================================="
    echo "Test Types: ${test_types[*]}"
    echo "Coverage Threshold: ${COVERAGE_THRESHOLD}%"
    echo "Test Timeout: $TEST_TIMEOUT"
    echo "=========================================="
    echo
    
    # Set up trap for cleanup
    trap cleanup EXIT
    
    # Run tests
    check_prerequisites
    setup_test_environment
    start_test_services
    
    for test_type in "${test_types[@]}"; do
        case $test_type in
            "unit")
                run_unit_tests
                ;;
            "integration")
                run_integration_tests
                ;;
            "performance")
                run_performance_tests
                ;;
            "e2e")
                run_e2e_tests
                ;;
            "security")
                run_security_tests
                ;;
        esac
    done
    
    generate_test_report
    print_test_summary
    
    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "All tests passed!"
        exit 0
    else
        log_error "Some tests failed!"
        exit 1
    fi
}

# Run main function
main "$@"
