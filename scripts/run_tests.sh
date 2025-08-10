#!/bin/bash

# KYB Tool - Test Runner Script
# This script provides a comprehensive test execution framework

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_CONFIG_FILE="$PROJECT_ROOT/test/test_config.yaml"
COVERAGE_DIR="$PROJECT_ROOT/test/coverage"
REPORTS_DIR="$PROJECT_ROOT/test/reports"
TEST_DATA_DIR="$PROJECT_ROOT/test/testdata"

# Default values
TEST_TYPE="unit"
VERBOSE=false
COVERAGE=false
PARALLEL=false
TIMEOUT="30s"
OUTPUT_FORMAT="text"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date +'%Y-%m-%d %H:%M:%S')] ${message}${NC}"
}

# Function to print usage
print_usage() {
    echo "KYB Tool - Test Runner"
    echo "====================="
    echo ""
    echo "Usage: $0 [OPTIONS] [TEST_PATHS...]"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE        Test type: unit, integration, performance, e2e (default: unit)"
    echo "  -v, --verbose          Enable verbose output"
    echo "  -c, --coverage         Generate coverage report"
    echo "  -p, --parallel         Run tests in parallel"
    echo "  --timeout DURATION     Test timeout (default: 30s)"
    echo "  -f, --format FORMAT    Output format: text, json, junit (default: text)"
    echo "  -h, --help             Show this help message"
    echo ""
    echo "Test Types:"
    echo "  unit                   Unit tests (fast, no external dependencies)"
    echo "  integration            Integration tests (requires database)"
    echo "  performance            Performance and load tests"
    echo "  e2e                    End-to-end tests (requires full stack)"
    echo "  all                    Run all test types"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run unit tests"
    echo "  $0 -t integration                     # Run integration tests"
    echo "  $0 -c -v ./internal/config           # Run config tests with coverage and verbose output"
    echo "  $0 -t all -c -f json                 # Run all tests with coverage in JSON format"
    echo ""
}

# Function to check prerequisites
check_prerequisites() {
    print_status $BLUE "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status $RED "Error: Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.22"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_status $RED "Error: Go version $GO_VERSION is less than required version $REQUIRED_VERSION"
        exit 1
    fi
    
    print_status $GREEN "✓ Go $GO_VERSION is installed"
    
    # Check if we're in the project root
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        print_status $RED "Error: Not in project root directory"
        exit 1
    fi
    
    print_status $GREEN "✓ In project root directory"
    
    # Create necessary directories
    mkdir -p "$COVERAGE_DIR"
    mkdir -p "$REPORTS_DIR"
    mkdir -p "$TEST_DATA_DIR"
    
    print_status $GREEN "✓ Test directories created"
}

# Function to setup test environment
setup_test_environment() {
    print_status $BLUE "Setting up test environment..."
    
    # Set test environment variables
    export TEST_ENV="true"
    export TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
    export TEST_DB_PORT="${TEST_DB_PORT:-5432}"
    export TEST_DB_USER="${TEST_DB_USER:-test_user}"
    export TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-test_password}"
    export TEST_DB_NAME="${TEST_DB_NAME:-kyb_test}"
    
    # Set test timeouts
    export TEST_TIMEOUT="$TIMEOUT"
    export TEST_VERBOSE="$VERBOSE"
    
    print_status $GREEN "✓ Test environment configured"
}

# Function to run unit tests
run_unit_tests() {
    print_status $BLUE "Running unit tests..."
    
    local test_args=()
    
    if [ "$VERBOSE" = true ]; then
        test_args+=("-v")
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args+=("-coverprofile=$COVERAGE_DIR/unit_coverage.out")
    fi
    
    if [ "$PARALLEL" = true ]; then
        test_args+=("-parallel=4")
    fi
    
    test_args+=("-timeout=$TIMEOUT")
    test_args+=("./internal/...")
    
    # Run all tests except integration and e2e
    test_args+=("-run" "^Test[^IE].*")
    
    if ! go test "${test_args[@]}"; then
        print_status $RED "✗ Unit tests failed"
        return 1
    fi
    
    print_status $GREEN "✓ Unit tests passed"
}

# Function to run integration tests
run_integration_tests() {
    print_status $BLUE "Running integration tests..."
    
    # Check if database is available
    if ! check_database_connection; then
        print_status $YELLOW "⚠ Database not available, skipping integration tests"
        return 0
    fi
    
    local test_args=()
    
    if [ "$VERBOSE" = true ]; then
        test_args+=("-v")
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args+=("-coverprofile=$COVERAGE_DIR/integration_coverage.out")
    fi
    
    test_args+=("-timeout=$TIMEOUT")
    test_args+=("-tags=integration")
    test_args+=("./internal/...")
    
    # Only run integration tests
    test_args+=("-run" "^TestIntegration")
    
    if ! go test "${test_args[@]}"; then
        print_status $RED "✗ Integration tests failed"
        return 1
    fi
    
    print_status $GREEN "✓ Integration tests passed"
}

# Function to run performance tests
run_performance_tests() {
    print_status $BLUE "Running performance tests..."
    
    local test_args=()
    
    if [ "$VERBOSE" = true ]; then
        test_args+=("-v")
    fi
    
    test_args+=("-timeout=$TIMEOUT")
    test_args+=("-tags=performance")
    test_args+=("./internal/...")
    
    # Only run performance tests
    test_args+=("-run" "^TestPerformance")
    
    if ! go test "${test_args[@]}"; then
        print_status $RED "✗ Performance tests failed"
        return 1
    fi
    
    print_status $GREEN "✓ Performance tests passed"
}

# Function to run e2e tests
run_e2e_tests() {
    print_status $BLUE "Running end-to-end tests..."
    
    # Check if full stack is available
    if ! check_full_stack_availability; then
        print_status $YELLOW "⚠ Full stack not available, skipping e2e tests"
        return 0
    fi
    
    local test_args=()
    
    if [ "$VERBOSE" = true ]; then
        test_args+=("-v")
    fi
    
    test_args+=("-timeout=$TIMEOUT")
    test_args+=("-tags=e2e")
    test_args+=("./internal/...")
    
    # Only run e2e tests
    test_args+=("-run" "^TestE2E")
    
    if ! go test "${test_args[@]}"; then
        print_status $RED "✗ E2E tests failed"
        return 1
    fi
    
    print_status $GREEN "✓ E2E tests passed"
}

# Function to check database connection
check_database_connection() {
    # Simple check - can be enhanced based on your database setup
    if command -v pg_isready &> /dev/null; then
        pg_isready -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" > /dev/null 2>&1
        return $?
    fi
    
    # Fallback: try to connect with Go
    go run -c "
    package main
    import (
        \"database/sql\"
        \"os\"
        _ \"github.com/lib/pq\"
    )
    func main() {
        dsn := os.Getenv(\"TEST_DB_HOST\") + \":\" + os.Getenv(\"TEST_DB_PORT\") + \"/\" + os.Getenv(\"TEST_DB_NAME\") + \"?user=\" + os.Getenv(\"TEST_DB_USER\") + \"&password=\" + os.Getenv(\"TEST_DB_PASSWORD\") + \"&sslmode=disable\"
        db, err := sql.Open(\"postgres\", dsn)
        if err != nil {
            os.Exit(1)
        }
        defer db.Close()
        if err := db.Ping(); err != nil {
            os.Exit(1)
        }
    }" > /dev/null 2>&1
}

# Function to check full stack availability
check_full_stack_availability() {
    # Check if API server is running
    if curl -s "http://localhost:8080/health" > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Function to generate coverage report
generate_coverage_report() {
    if [ "$COVERAGE" != true ]; then
        return 0
    fi
    
    print_status $BLUE "Generating coverage report..."
    
    # Merge coverage files if they exist
    local coverage_files=()
    if [ -f "$COVERAGE_DIR/unit_coverage.out" ]; then
        coverage_files+=("$COVERAGE_DIR/unit_coverage.out")
    fi
    if [ -f "$COVERAGE_DIR/integration_coverage.out" ]; then
        coverage_files+=("$COVERAGE_DIR/integration_coverage.out")
    fi
    
    if [ ${#coverage_files[@]} -eq 0 ]; then
        print_status $YELLOW "⚠ No coverage files found"
        return 0
    fi
    
    # Merge coverage files
    if [ ${#coverage_files[@]} -gt 1 ]; then
        go tool cover -func="${coverage_files[0]}" > "$COVERAGE_DIR/coverage.txt"
        for file in "${coverage_files[@]:1}"; do
            go tool cover -func="$file" >> "$COVERAGE_DIR/coverage.txt"
        done
    else
        go tool cover -func="${coverage_files[0]}" > "$COVERAGE_DIR/coverage.txt"
    fi
    
    # Generate HTML report
    go tool cover -html="${coverage_files[0]}" -o "$COVERAGE_DIR/coverage.html"
    
    print_status $GREEN "✓ Coverage report generated: $COVERAGE_DIR/coverage.html"
}

# Function to cleanup test environment
cleanup_test_environment() {
    print_status $BLUE "Cleaning up test environment..."
    
    # Remove temporary files
    find "$PROJECT_ROOT" -name "*.test" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "test.db" -delete 2>/dev/null || true
    
    print_status $GREEN "✓ Test environment cleaned up"
}

# Function to parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type)
                TEST_TYPE="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL=true
                shift
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -f|--format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            -*)
                print_status $RED "Error: Unknown option $1"
                print_usage
                exit 1
                ;;
            *)
                break
                ;;
        esac
    done
}

# Main function
main() {
    print_status $BLUE "Starting KYB Tool test runner..."
    
    # Parse arguments
    parse_arguments "$@"
    
    # Check prerequisites
    check_prerequisites
    
    # Setup test environment
    setup_test_environment
    
    # Run tests based on type
    local exit_code=0
    
    case $TEST_TYPE in
        unit)
            run_unit_tests || exit_code=1
            ;;
        integration)
            run_integration_tests || exit_code=1
            ;;
        performance)
            run_performance_tests || exit_code=1
            ;;
        e2e)
            run_e2e_tests || exit_code=1
            ;;
        all)
            run_unit_tests || exit_code=1
            run_integration_tests || exit_code=1
            run_performance_tests || exit_code=1
            run_e2e_tests || exit_code=1
            ;;
        *)
            print_status $RED "Error: Unknown test type '$TEST_TYPE'"
            print_usage
            exit 1
            ;;
    esac
    
    # Generate coverage report
    generate_coverage_report
    
    # Cleanup
    cleanup_test_environment
    
    # Final status
    if [ $exit_code -eq 0 ]; then
        print_status $GREEN "✓ All tests completed successfully"
    else
        print_status $RED "✗ Some tests failed"
    fi
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"
