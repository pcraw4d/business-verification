#!/bin/bash

# KYB Platform Performance Testing Script
# This script provides a comprehensive interface for running performance tests

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_BASE_URL="http://localhost:8080"
DEFAULT_REPORT_PATH="./performance-reports"
DEFAULT_ENVIRONMENT="development"

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

# Help function
show_help() {
    cat << EOF
KYB Platform Performance Testing Script

Usage: $0 [OPTIONS] [TEST_TYPE]

OPTIONS:
    -u, --base-url URL        Base URL for KYB platform API (default: $DEFAULT_BASE_URL)
    -r, --report-path PATH    Path to save performance reports (default: $DEFAULT_REPORT_PATH)
    -e, --environment ENV     Environment: development, staging, production (default: $DEFAULT_ENVIRONMENT)
    -t, --timeout SECONDS     Timeout for tests in seconds (default: 1800)
    -v, --verbose             Enable verbose output
    -h, --help                Show this help message

TEST_TYPE:
    load                     Run load test only
    stress                   Run stress test only
    memory                   Run memory test only
    response-time            Run response time test only
    end-to-end               Run end-to-end test only
    comprehensive            Run all performance tests (default)
    quick                    Run quick validation test
    custom                   Run custom test configuration

EXAMPLES:
    $0                                    # Run comprehensive tests
    $0 load                              # Run load test only
    $0 -u https://api.kyb-platform.com   # Run against production API
    $0 -e staging -t 3600                # Run against staging with 1 hour timeout
    $0 -v comprehensive                  # Run comprehensive tests with verbose output

PERFORMANCE TARGETS:
    - API response times: <200ms average
    - ML model inference: <100ms for classification, <50ms for risk detection
    - Database query performance: 50% improvement
    - System uptime: 99.9%
    - Error rate: <1%
    - Throughput: >100 req/s

EOF
}

# Parse command line arguments
parse_arguments() {
    BASE_URL="$DEFAULT_BASE_URL"
    REPORT_PATH="$DEFAULT_REPORT_PATH"
    ENVIRONMENT="$DEFAULT_ENVIRONMENT"
    TIMEOUT=1800
    VERBOSE=false
    TEST_TYPE="comprehensive"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--base-url)
                BASE_URL="$2"
                shift 2
                ;;
            -r|--report-path)
                REPORT_PATH="$2"
                shift 2
                ;;
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            load|stress|memory|response-time|end-to-end|comprehensive|quick|custom)
                TEST_TYPE="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate environment
validate_environment() {
    log_info "Validating environment..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the project root
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Please run this script from the project root."
        exit 1
    fi
    
    # Check if performance testing framework exists
    if [ ! -d "internal/testing/performance" ]; then
        log_error "Performance testing framework not found"
        exit 1
    fi
    
    # Check if configuration file exists
    if [ ! -f "configs/performance-test-config.json" ]; then
        log_warning "Performance test configuration file not found"
    fi
    
    log_success "Environment validation passed"
}

# Setup test environment
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create report directory
    mkdir -p "$REPORT_PATH"
    
    # Create bin directory
    mkdir -p "./bin"
    
    # Download dependencies
    log_info "Downloading Go dependencies..."
    go mod download
    go mod tidy
    
    log_success "Test environment setup completed"
}

# Build performance test binary
build_performance_test_binary() {
    log_info "Building performance test binary..."
    
    if [ -f "./bin/performance-test" ] && [ "./bin/performance-test" -nt "./cmd/performance-test/main.go" ]; then
        log_info "Performance test binary is up to date"
        return
    fi
    
    go build -o "./bin/performance-test" ./cmd/performance-test
    
    if [ $? -eq 0 ]; then
        log_success "Performance test binary built successfully"
    else
        log_error "Failed to build performance test binary"
        exit 1
    fi
}

# Run performance test
run_performance_test() {
    local test_type="$1"
    local base_url="$2"
    local report_path="$3"
    local timeout="$4"
    local verbose="$5"
    
    log_info "Running $test_type performance test..."
    log_info "Base URL: $base_url"
    log_info "Report Path: $report_path"
    log_info "Timeout: ${timeout}s"
    
    # Build command
    local cmd="./bin/performance-test -base-url $base_url -report-path $report_path -test-type $test_type"
    
    if [ "$verbose" = true ]; then
        cmd="$cmd -v"
    fi
    
    # Run test with timeout
    if timeout "$timeout" $cmd; then
        log_success "$test_type test completed successfully"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            log_error "$test_type test timed out after ${timeout}s"
        else
            log_error "$test_type test failed with exit code $exit_code"
        fi
        return $exit_code
    fi
}

# Generate test summary
generate_test_summary() {
    local report_path="$1"
    
    log_info "Generating test summary..."
    
    if [ ! -d "$report_path" ]; then
        log_warning "No report directory found: $report_path"
        return
    fi
    
    # Count test reports
    local json_reports=$(find "$report_path" -name "*.json" | wc -l)
    local markdown_reports=$(find "$report_path" -name "*.md" | wc -l)
    
    log_info "Test reports generated:"
    log_info "  JSON reports: $json_reports"
    log_info "  Markdown reports: $markdown_reports"
    
    # Show latest report
    local latest_report=$(find "$report_path" -name "*.json" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-)
    if [ -n "$latest_report" ]; then
        log_info "Latest report: $latest_report"
        
        # Show summary if jq is available
        if command -v jq &> /dev/null; then
            log_info "Report summary:"
            jq -r '.summary // empty' "$latest_report" 2>/dev/null || log_warning "Could not parse report summary"
        else
            log_warning "Install jq for pretty JSON output and report parsing"
        fi
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Add any cleanup tasks here
}

# Main function
main() {
    # Set up signal handling
    trap cleanup EXIT
    
    # Parse arguments
    parse_arguments "$@"
    
    # Show configuration
    log_info "Performance Test Configuration:"
    log_info "  Base URL: $BASE_URL"
    log_info "  Report Path: $REPORT_PATH"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Test Type: $TEST_TYPE"
    log_info "  Timeout: ${TIMEOUT}s"
    log_info "  Verbose: $VERBOSE"
    
    # Validate environment
    validate_environment
    
    # Setup test environment
    setup_test_environment
    
    # Build performance test binary
    build_performance_test_binary
    
    # Run the specified test
    case "$TEST_TYPE" in
        load|stress|memory|response-time|end-to-end|comprehensive)
            run_performance_test "$TEST_TYPE" "$BASE_URL" "$REPORT_PATH" "$TIMEOUT" "$VERBOSE"
            ;;
        quick)
            run_performance_test "response-time" "$BASE_URL" "$REPORT_PATH" 300 "$VERBOSE"
            ;;
        custom)
            log_info "Custom test configuration not implemented yet"
            log_info "Please use one of the predefined test types"
            exit 1
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            show_help
            exit 1
            ;;
    esac
    
    # Generate test summary
    generate_test_summary "$REPORT_PATH"
    
    log_success "Performance testing completed!"
}

# Run main function
main "$@"