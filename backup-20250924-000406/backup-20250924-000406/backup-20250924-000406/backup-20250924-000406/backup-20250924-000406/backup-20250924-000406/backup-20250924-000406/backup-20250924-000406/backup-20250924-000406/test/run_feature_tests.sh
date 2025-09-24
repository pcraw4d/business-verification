#!/bin/bash

# Feature Functionality Testing Script
# This script executes comprehensive feature functionality tests for the KYB Platform

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$SCRIPT_DIR"
CONFIG_FILE="$TEST_DIR/test_config.yaml"
REPORT_DIR="$TEST_DIR/test-reports"
LOG_DIR="$TEST_DIR/test-logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -c, --config FILE       Use custom config file (default: test_config.yaml)"
    echo "  -r, --report-dir DIR    Set report directory (default: test-reports)"
    echo "  -l, --log-dir DIR       Set log directory (default: test-logs)"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -p, --parallel          Enable parallel test execution"
    echo "  -t, --timeout DURATION  Set test timeout (default: 30m)"
    echo "  --business-classification  Run only business classification tests"
    echo "  --risk-assessment       Run only risk assessment tests"
    echo "  --compliance-checking   Run only compliance checking tests"
    echo "  --merchant-management   Run only merchant management tests"
    echo "  --benchmark             Run benchmark tests"
    echo "  --load-test             Run load tests"
    echo "  --clean                 Clean up test artifacts before running"
    echo "  --no-cleanup            Skip cleanup after tests"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all tests"
    echo "  $0 --business-classification         # Run only business classification tests"
    echo "  $0 --benchmark                       # Run benchmark tests"
    echo "  $0 --load-test --timeout 10m         # Run load tests with 10 minute timeout"
    echo "  $0 --clean --verbose                 # Clean and run with verbose output"
}

# Default values
CONFIG_FILE="$TEST_DIR/test_config.yaml"
REPORT_DIR="$TEST_DIR/test-reports"
LOG_DIR="$TEST_DIR/test-logs"
VERBOSE=false
PARALLEL=false
TIMEOUT="30m"
CLEAN=false
NO_CLEANUP=false
BENCHMARK=false
LOAD_TEST=false
TEST_CATEGORY=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -r|--report-dir)
            REPORT_DIR="$2"
            shift 2
            ;;
        -l|--log-dir)
            LOG_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --business-classification)
            TEST_CATEGORY="business_classification"
            shift
            ;;
        --risk-assessment)
            TEST_CATEGORY="risk_assessment"
            shift
            ;;
        --compliance-checking)
            TEST_CATEGORY="compliance_checking"
            shift
            ;;
        --merchant-management)
            TEST_CATEGORY="merchant_management"
            shift
            ;;
        --benchmark)
            BENCHMARK=true
            shift
            ;;
        --load-test)
            LOAD_TEST=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --no-cleanup)
            NO_CLEANUP=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Function to check prerequisites
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
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "go.mod not found. Please run this script from the project root or test directory."
        exit 1
    fi
    
    # Check if config file exists
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Config file not found: $CONFIG_FILE"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to setup test environment
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create directories
    mkdir -p "$REPORT_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$TEST_DIR/testdata"
    
    # Set environment variables
    export TEST_CONFIG_FILE="$CONFIG_FILE"
    export TEST_REPORT_DIR="$REPORT_DIR"
    export TEST_LOG_DIR="$LOG_DIR"
    export TEST_VERBOSE="$VERBOSE"
    export TEST_PARALLEL="$PARALLEL"
    export TEST_TIMEOUT="$TIMEOUT"
    
    log_success "Test environment setup completed"
}

# Function to clean test artifacts
clean_test_artifacts() {
    log_info "Cleaning test artifacts..."
    
    # Clean test reports
    if [[ -d "$REPORT_DIR" ]]; then
        rm -rf "$REPORT_DIR"/*
        log_info "Cleaned test reports directory"
    fi
    
    # Clean test logs
    if [[ -d "$LOG_DIR" ]]; then
        rm -rf "$LOG_DIR"/*
        log_info "Cleaned test logs directory"
    fi
    
    # Clean test data
    if [[ -d "$TEST_DIR/testdata" ]]; then
        rm -rf "$TEST_DIR/testdata"/*
        log_info "Cleaned test data directory"
    fi
    
    # Clean Go test cache
    go clean -testcache
    log_info "Cleaned Go test cache"
    
    log_success "Test artifacts cleaned"
}

# Function to run tests
run_tests() {
    log_info "Running feature functionality tests..."
    
    # Build test flags
    TEST_FLAGS="-v"
    
    if [[ "$VERBOSE" == "true" ]]; then
        TEST_FLAGS="$TEST_FLAGS -test.v"
    fi
    
    if [[ "$PARALLEL" == "true" ]]; then
        TEST_FLAGS="$TEST_FLAGS -test.parallel"
    fi
    
    # Set timeout
    TEST_FLAGS="$TEST_FLAGS -test.timeout=$TIMEOUT"
    
    # Change to test directory
    cd "$TEST_DIR"
    
    # Run specific test category or all tests
    if [[ -n "$TEST_CATEGORY" ]]; then
        log_info "Running $TEST_CATEGORY tests..."
        go test $TEST_FLAGS -run "TestFeatureFunctionality/$TEST_CATEGORY" ./...
    else
        log_info "Running all feature functionality tests..."
        go test $TEST_FLAGS -run "TestFeatureFunctionality" ./...
    fi
    
    # Check test result
    if [[ $? -eq 0 ]]; then
        log_success "Tests completed successfully"
    else
        log_error "Tests failed"
        exit 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    log_info "Running benchmark tests..."
    
    # Change to test directory
    cd "$TEST_DIR"
    
    # Run benchmark tests
    go test -bench=. -benchmem -run=^$ ./...
    
    # Check benchmark result
    if [[ $? -eq 0 ]]; then
        log_success "Benchmark tests completed successfully"
    else
        log_error "Benchmark tests failed"
        exit 1
    fi
}

# Function to run load tests
run_load_tests() {
    log_info "Running load tests..."
    
    # Change to test directory
    cd "$TEST_DIR"
    
    # Run load tests with timeout
    timeout "$TIMEOUT" go test -run "TestLoad" -test.timeout="$TIMEOUT" ./...
    
    # Check load test result
    if [[ $? -eq 0 ]]; then
        log_success "Load tests completed successfully"
    else
        log_error "Load tests failed or timed out"
        exit 1
    fi
}

# Function to generate test report
generate_test_report() {
    log_info "Generating test report..."
    
    # Check if report directory exists and has files
    if [[ -d "$REPORT_DIR" && "$(ls -A "$REPORT_DIR")" ]]; then
        log_success "Test report generated in $REPORT_DIR"
        
        # List report files
        log_info "Report files:"
        ls -la "$REPORT_DIR"
    else
        log_warning "No test report generated"
    fi
}

# Function to cleanup after tests
cleanup_after_tests() {
    if [[ "$NO_CLEANUP" == "false" ]]; then
        log_info "Cleaning up after tests..."
        
        # Clean temporary files
        find "$TEST_DIR" -name "*.tmp" -delete 2>/dev/null || true
        
        # Clean test cache if requested
        if [[ "$CLEAN" == "true" ]]; then
            go clean -testcache
        fi
        
        log_success "Cleanup completed"
    else
        log_info "Skipping cleanup (--no-cleanup specified)"
    fi
}

# Function to show test summary
show_test_summary() {
    log_info "Test Summary:"
    echo "  Config File: $CONFIG_FILE"
    echo "  Report Directory: $REPORT_DIR"
    echo "  Log Directory: $LOG_DIR"
    echo "  Verbose: $VERBOSE"
    echo "  Parallel: $PARALLEL"
    echo "  Timeout: $TIMEOUT"
    echo "  Test Category: ${TEST_CATEGORY:-all}"
    echo "  Benchmark: $BENCHMARK"
    echo "  Load Test: $LOAD_TEST"
    echo "  Clean: $CLEAN"
    echo "  No Cleanup: $NO_CLEANUP"
}

# Main execution
main() {
    log_info "Starting Feature Functionality Testing"
    log_info "======================================"
    
    # Show test summary
    show_test_summary
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    # Setup test environment
    setup_test_environment
    
    # Clean test artifacts if requested
    if [[ "$CLEAN" == "true" ]]; then
        clean_test_artifacts
    fi
    
    # Run tests based on options
    if [[ "$BENCHMARK" == "true" ]]; then
        run_benchmark_tests
    elif [[ "$LOAD_TEST" == "true" ]]; then
        run_load_tests
    else
        run_tests
    fi
    
    # Generate test report
    generate_test_report
    
    # Cleanup after tests
    cleanup_after_tests
    
    log_success "Feature Functionality Testing completed successfully!"
    log_info "Check the report directory for detailed results: $REPORT_DIR"
}

# Run main function
main "$@"
