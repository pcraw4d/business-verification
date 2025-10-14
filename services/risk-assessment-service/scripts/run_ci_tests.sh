#!/bin/bash
set -e

# CI Test Runner for Risk Assessment Service
# This script runs all tests in the correct order for CI/CD pipeline

echo "🧪 Starting CI Test Suite for Risk Assessment Service..."

# Configuration
COVERAGE_THRESHOLD=95
TIMEOUT=10m
VERBOSE=false
SKIP_INTEGRATION=false
SKIP_PERFORMANCE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --coverage-threshold)
      COVERAGE_THRESHOLD="$2"
      shift 2
      ;;
    --timeout)
      TIMEOUT="$2"
      shift 2
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    --skip-integration)
      SKIP_INTEGRATION=true
      shift
      ;;
    --skip-performance)
      SKIP_PERFORMANCE=true
      shift
      ;;
    --help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  --coverage-threshold N    Set coverage threshold (default: 95)"
      echo "  --timeout DURATION        Set test timeout (default: 10m)"
      echo "  --verbose                 Enable verbose output"
      echo "  --skip-integration        Skip integration tests"
      echo "  --skip-performance        Skip performance tests"
      echo "  --help                    Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Set verbose flag for Go tests
if [ "$VERBOSE" = true ]; then
  GO_TEST_FLAGS="-v"
else
  GO_TEST_FLAGS=""
fi

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

# Function to check if required environment variables are set
check_environment() {
    print_status "Checking environment variables..."
    
    local required_vars=(
        "DATABASE_URL"
        "REDIS_URL"
        "SUPABASE_URL"
        "SUPABASE_ANON_KEY"
        "SUPABASE_SERVICE_ROLE_KEY"
    )
    
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -ne 0 ]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        exit 1
    fi
    
    print_success "All required environment variables are set"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    local start_time=$(date +%s)
    
    # Run unit tests with coverage
    go test $GO_TEST_FLAGS \
        -race \
        -coverprofile=coverage.out \
        -covermode=atomic \
        -timeout=$TIMEOUT \
        ./...
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Unit tests passed in ${duration}s"
    else
        print_error "Unit tests failed"
        exit $exit_code
    fi
}

# Function to check test coverage
check_coverage() {
    print_status "Checking test coverage..."
    
    if [ ! -f "coverage.out" ]; then
        print_error "Coverage file not found. Run unit tests first."
        exit 1
    fi
    
    # Generate coverage report
    go tool cover -func=coverage.out > coverage_report.txt
    go tool cover -html=coverage.out -o coverage.html
    
    # Extract total coverage percentage
    local total_coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    
    print_status "Total coverage: ${total_coverage}%"
    print_status "Coverage threshold: ${COVERAGE_THRESHOLD}%"
    
    # Check if coverage meets threshold
    if (( $(echo "$total_coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
        print_success "Coverage meets threshold"
    else
        print_error "Coverage ${total_coverage}% is below threshold ${COVERAGE_THRESHOLD}%"
        exit 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    if [ "$SKIP_INTEGRATION" = true ]; then
        print_warning "Skipping integration tests"
        return 0
    fi
    
    print_status "Running integration tests..."
    
    local start_time=$(date +%s)
    
    # Run integration tests
    go test $GO_TEST_FLAGS \
        -tags=integration \
        -timeout=$TIMEOUT \
        ./...
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Integration tests passed in ${duration}s"
    else
        print_error "Integration tests failed"
        exit $exit_code
    fi
}

# Function to run performance tests
run_performance_tests() {
    if [ "$SKIP_PERFORMANCE" = true ]; then
        print_warning "Skipping performance tests"
        return 0
    fi
    
    print_status "Running performance tests..."
    
    local start_time=$(date +%s)
    
    # Run performance tests
    go test $GO_TEST_FLAGS \
        -tags=performance \
        -bench=. \
        -benchmem \
        -timeout=$TIMEOUT \
        ./...
    
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        print_success "Performance tests passed in ${duration}s"
    else
        print_error "Performance tests failed"
        exit $exit_code
    fi
}

# Function to run security tests
run_security_tests() {
    print_status "Running security tests..."
    
    # Install gosec if not present
    if ! command -v gosec &> /dev/null; then
        print_status "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Run security scan
    gosec -fmt json -out gosec-report.json ./...
    
    # Check for high severity issues
    if [ -f gosec-report.json ]; then
        local high_issues=$(jq '.Stats.High' gosec-report.json 2>/dev/null || echo "0")
        if [ "$high_issues" -gt 0 ]; then
            print_error "High severity security issues found: $high_issues"
            cat gosec-report.json
            exit 1
        fi
    fi
    
    print_success "Security tests passed"
}

# Function to run linting
run_linting() {
    print_status "Running linters..."
    
    # Install golangci-lint if not present
    if ! command -v golangci-lint &> /dev/null; then
        print_status "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    fi
    
    # Run linters
    golangci-lint run --timeout=5m --verbose
    
    print_success "Linting passed"
}

# Function to generate test report
generate_test_report() {
    print_status "Generating test report..."
    
    local report_file="test-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "service": "risk-assessment-service",
  "environment": "${ENVIRONMENT:-unknown}",
  "go_version": "$(go version | awk '{print $3}')",
  "test_results": {
    "unit_tests": {
      "status": "passed",
      "coverage": "$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')%"
    },
    "integration_tests": {
      "status": "$([ "$SKIP_INTEGRATION" = true ] && echo "skipped" || echo "passed")"
    },
    "performance_tests": {
      "status": "$([ "$SKIP_PERFORMANCE" = true ] && echo "skipped" || echo "passed")"
    },
    "security_tests": {
      "status": "passed"
    },
    "linting": {
      "status": "passed"
    }
  },
  "artifacts": [
    "coverage.out",
    "coverage.html",
    "coverage_report.txt",
    "gosec-report.json",
    "$report_file"
  ]
}
EOF
    
    print_success "Test report generated: $report_file"
}

# Main execution
main() {
    print_status "Starting CI test suite with the following configuration:"
    echo "  Coverage threshold: ${COVERAGE_THRESHOLD}%"
    echo "  Timeout: $TIMEOUT"
    echo "  Verbose: $VERBOSE"
    echo "  Skip integration: $SKIP_INTEGRATION"
    echo "  Skip performance: $SKIP_PERFORMANCE"
    echo ""
    
    # Change to service directory
    cd "$(dirname "$0")/.."
    
    # Check environment
    check_environment
    
    # Run tests in order
    run_linting
    run_security_tests
    run_unit_tests
    check_coverage
    run_integration_tests
    run_performance_tests
    
    # Generate report
    generate_test_report
    
    print_success "All tests passed! 🎉"
    print_status "Test artifacts:"
    echo "  - coverage.out (coverage data)"
    echo "  - coverage.html (coverage report)"
    echo "  - coverage_report.txt (coverage summary)"
    echo "  - gosec-report.json (security scan results)"
    echo "  - test-report-*.json (test summary)"
}

# Run main function
main "$@"
