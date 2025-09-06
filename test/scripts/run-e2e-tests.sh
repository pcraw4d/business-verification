#!/bin/bash

# E2E Test Runner Script
# This script sets up and runs end-to-end tests with Docker containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$TEST_DIR")"
DOCKER_COMPOSE_FILE="$TEST_DIR/docker-compose.e2e.yml"

# Default values
CLEANUP=true
VERBOSE=false
TIMEOUT=300
TEST_PATTERN=""

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

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

E2E Test Runner for KYB Platform

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -c, --no-cleanup        Don't cleanup containers after tests
    -t, --timeout SECONDS   Test timeout in seconds (default: 300)
    -p, --pattern PATTERN   Test pattern to run (default: all)
    --skip-setup            Skip Docker setup (assume containers are running)
    --skip-teardown         Skip Docker teardown

EXAMPLES:
    $0                      # Run all E2E tests
    $0 -v                   # Run with verbose output
    $0 -p "TestClassification"  # Run only classification tests
    $0 --no-cleanup         # Keep containers running after tests

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--no-cleanup)
            CLEANUP=false
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -p|--pattern)
            TEST_PATTERN="$2"
            shift 2
            ;;
        --skip-setup)
            SKIP_SETUP=true
            shift
            ;;
        --skip-teardown)
            SKIP_TEARDOWN=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Function to cleanup containers
cleanup_containers() {
    if [[ "$CLEANUP" == "true" && "$SKIP_TEARDOWN" != "true" ]]; then
        print_status "Cleaning up Docker containers..."
        cd "$TEST_DIR"
        docker-compose -f docker-compose.e2e.yml down -v --remove-orphans || true
        print_success "Containers cleaned up"
    fi
}

# Function to setup test environment
setup_test_environment() {
    if [[ "$SKIP_SETUP" == "true" ]]; then
        print_status "Skipping Docker setup (containers assumed to be running)"
        return
    fi

    print_status "Setting up E2E test environment..."

    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi

    # Check if docker-compose is available
    if ! command -v docker-compose >/dev/null 2>&1; then
        print_error "docker-compose is not installed. Please install docker-compose and try again."
        exit 1
    fi

    # Navigate to test directory
    cd "$TEST_DIR"

    # Stop any existing containers
    print_status "Stopping existing containers..."
    docker-compose -f docker-compose.e2e.yml down -v --remove-orphans || true

    # Build and start containers
    print_status "Building and starting test containers..."
    docker-compose -f docker-compose.e2e.yml up -d --build

    # Wait for services to be healthy
    print_status "Waiting for services to be healthy..."
    timeout $TIMEOUT bash -c 'until docker-compose -f docker-compose.e2e.yml ps | grep -q "healthy"; do sleep 2; done' || {
        print_error "Services failed to become healthy within $TIMEOUT seconds"
        docker-compose -f docker-compose.e2e.yml logs
        cleanup_containers
        exit 1
    }

    print_success "Test environment setup complete"
}

# Function to run E2E tests
run_e2e_tests() {
    print_status "Running E2E tests..."

    # Set environment variables for tests
    export E2E_TESTS=true
    export TEST_MODE=true
    export LOG_LEVEL=debug

    # Navigate to project root
    cd "$PROJECT_ROOT"

    # Build test command
    TEST_CMD="go test -v -tags=e2e -timeout=${TIMEOUT}s"
    
    if [[ -n "$TEST_PATTERN" ]]; then
        TEST_CMD="$TEST_CMD -run $TEST_PATTERN"
    fi
    
    TEST_CMD="$TEST_CMD ./test/e2e/..."

    if [[ "$VERBOSE" == "true" ]]; then
        TEST_CMD="$TEST_CMD -test.v"
    fi

    print_status "Executing: $TEST_CMD"

    # Run tests
    if eval "$TEST_CMD"; then
        print_success "E2E tests completed successfully"
        return 0
    else
        print_error "E2E tests failed"
        return 1
    fi
}

# Function to show test results
show_test_results() {
    print_status "Test execution completed"
    
    if [[ -f "$TEST_DIR/reports/e2e-results.json" ]]; then
        print_status "Test results available in $TEST_DIR/reports/e2e-results.json"
    fi
    
    if [[ -f "$TEST_DIR/reports/e2e-coverage.html" ]]; then
        print_status "Coverage report available in $TEST_DIR/reports/e2e-coverage.html"
    fi
}

# Main execution
main() {
    print_status "Starting E2E test execution..."
    print_status "Project root: $PROJECT_ROOT"
    print_status "Test directory: $TEST_DIR"
    print_status "Docker compose file: $DOCKER_COMPOSE_FILE"
    print_status "Timeout: $TIMEOUT seconds"
    print_status "Cleanup: $CLEANUP"
    print_status "Verbose: $VERBOSE"
    
    if [[ -n "$TEST_PATTERN" ]]; then
        print_status "Test pattern: $TEST_PATTERN"
    fi

    # Setup trap for cleanup
    trap cleanup_containers EXIT

    # Setup test environment
    setup_test_environment

    # Run E2E tests
    if run_e2e_tests; then
        print_success "All E2E tests passed!"
        show_test_results
        exit 0
    else
        print_error "E2E tests failed!"
        show_test_results
        exit 1
    fi
}

# Run main function
main "$@"
