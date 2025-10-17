#!/bin/bash

# Test Automation Runner for Risk Assessment Service
# This script runs the comprehensive test automation framework

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONFIG_FILE="$SCRIPT_DIR/automation_config.yaml"
REPORTS_DIR="$PROJECT_ROOT/test/automation/reports"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TEST_TYPE="all"
CONFIG_FILE_PATH="$CONFIG_FILE"
OUTPUT_DIR="$REPORTS_DIR"
VERBOSE=false
PARALLEL=false
SKIP_CLEANUP=false
SKIP_MONITORING=false

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

Test Automation Runner for Risk Assessment Service

OPTIONS:
    -t, --test-type TYPE     Test type: unit, integration, performance, security, e2e, ml, all (default: all)
    -c, --config FILE        Configuration file path (default: automation_config.yaml)
    -o, --output-dir DIR     Output directory (default: ./reports)
    -v, --verbose            Verbose output
    -p, --parallel           Run tests in parallel
    --skip-cleanup           Skip cleanup operations
    --skip-monitoring        Skip monitoring setup
    --help                   Show this help message

EXAMPLES:
    # Run all tests
    $0

    # Run only unit tests
    $0 -t unit

    # Run tests in parallel
    $0 -p

    # Run with custom config
    $0 -c custom_config.yaml

    # Run with verbose output
    $0 -v

EOF
}

# Function to check dependencies
check_dependencies() {
    print_status "Checking dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    
    # Check required Go version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    required_version="1.21"
    if [ "$(printf '%s\n' "$required_version" "$go_version" | sort -V | head -n1)" != "$required_version" ]; then
        print_error "Go version $required_version or higher is required. Current version: $go_version"
        exit 1
    fi
    
    # Check Python (for Locust)
    if ! command -v python3 &> /dev/null; then
        print_warning "Python 3 is not installed. Performance tests will be skipped."
    fi
    
    # Check Locust
    if ! command -v locust &> /dev/null; then
        print_warning "Locust is not installed. Installing..."
        pip3 install locust
    fi
    
    # Check security tools
    if ! command -v gosec &> /dev/null; then
        print_warning "gosec is not installed. Installing..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    if ! command -v trivy &> /dev/null; then
        print_warning "trivy is not installed. Installing..."
        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    fi
    
    if ! command -v nancy &> /dev/null; then
        print_warning "nancy is not installed. Installing..."
        go install github.com/sonatypecommunity/nancy@latest
    fi
    
    if ! command -v golangci-lint &> /dev/null; then
        print_warning "golangci-lint is not installed. Installing..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin
    fi
    
    print_success "All dependencies are available"
}

# Function to create output directory
create_output_dir() {
    if [ ! -d "$OUTPUT_DIR" ]; then
        print_status "Creating output directory: $OUTPUT_DIR"
        mkdir -p "$OUTPUT_DIR"
    fi
}

# Function to setup test environment
setup_test_environment() {
    print_status "Setting up test environment..."
    
    # Set environment variables
    export ENV=test
    export LOG_LEVEL=error
    export SUPABASE_URL=https://test.supabase.co
    export SUPABASE_API_KEY=test-api-key
    export REDIS_ADDRS=localhost:6379
    export REDIS_DB=1
    export REDIS_KEY_PREFIX=test:
    
    # Start test services if needed
    if command -v docker-compose &> /dev/null; then
        if [ -f "$PROJECT_ROOT/docker-compose.test.yml" ]; then
            print_status "Starting test services with Docker Compose"
            docker-compose -f "$PROJECT_ROOT/docker-compose.test.yml" up -d
        fi
    fi
    
    print_success "Test environment setup completed"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    local test_cmd="go test -v -race -coverprofile=coverage.out -covermode=atomic ./..."
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    if [ "$PARALLEL" = true ]; then
        test_cmd="$test_cmd -parallel 4"
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "Unit tests completed successfully"
        
        # Generate coverage report
        if [ -f "coverage.out" ]; then
            go tool cover -html=coverage.out -o "$OUTPUT_DIR/coverage.html"
            go tool cover -func=coverage.out > "$OUTPUT_DIR/coverage.txt"
            print_success "Coverage report generated"
        fi
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    local test_cmd="go test -tags=integration -v ./test/integration/..."
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "Integration tests completed successfully"
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running performance tests..."
    
    if ! command -v locust &> /dev/null; then
        print_warning "Locust not available, skipping performance tests"
        return 0
    fi
    
    local test_cmd="locust -f test/performance/locustfile.py --host=http://localhost:8080 --headless --users=100 --spawn-rate=10 --run-time=5m --html=$OUTPUT_DIR/performance_report.html"
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "Performance tests completed successfully"
    else
        print_error "Performance tests failed"
        return 1
    fi
}

# Function to run security tests
run_security_tests() {
    print_status "Running security tests..."
    
    # Run Go security tests
    local test_cmd="go test -tags=security -v ./test/security/..."
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "Security tests completed successfully"
    else
        print_error "Security tests failed"
        return 1
    fi
    
    # Run vulnerability scanning
    run_vulnerability_scanning
}

# Function to run vulnerability scanning
run_vulnerability_scanning() {
    print_status "Running vulnerability scanning..."
    
    # Run gosec
    if command -v gosec &> /dev/null; then
        print_status "Running gosec security scanner..."
        gosec -fmt json -out "$OUTPUT_DIR/gosec_report.json" ./...
    fi
    
    # Run trivy
    if command -v trivy &> /dev/null; then
        print_status "Running trivy vulnerability scanner..."
        trivy fs --format json --output "$OUTPUT_DIR/trivy_report.json" .
    fi
    
    # Run nancy
    if command -v nancy &> /dev/null; then
        print_status "Running nancy dependency scanner..."
        go list -json -deps ./... | nancy sleuth --format json --output "$OUTPUT_DIR/nancy_report.json"
    fi
    
    # Run golangci-lint
    if command -v golangci-lint &> /dev/null; then
        print_status "Running golangci-lint with security rules..."
        golangci-lint run --format json --out-format json --output "$OUTPUT_DIR/golangci_report.json"
    fi
    
    print_success "Vulnerability scanning completed"
}

# Function to run end-to-end tests
run_e2e_tests() {
    print_status "Running end-to-end tests..."
    
    local test_cmd="go test -tags=e2e -v ./test/e2e/..."
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "End-to-end tests completed successfully"
    else
        print_error "End-to-end tests failed"
        return 1
    fi
}

# Function to run ML tests
run_ml_tests() {
    print_status "Running ML model tests..."
    
    local test_cmd="go test -tags=ml -v ./test/ml/..."
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "ML model tests completed successfully"
    else
        print_error "ML model tests failed"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests..."
    
    local failed_tests=0
    
    # Run tests based on configuration
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "unit" ]; then
        if ! run_unit_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "integration" ]; then
        if ! run_integration_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "performance" ]; then
        if ! run_performance_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "security" ]; then
        if ! run_security_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "e2e" ]; then
        if ! run_e2e_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "ml" ]; then
        if ! run_ml_tests; then
            ((failed_tests++))
        fi
    fi
    
    if [ $failed_tests -eq 0 ]; then
        print_success "All tests completed successfully"
    else
        print_error "$failed_tests test(s) failed"
        return 1
    fi
}

# Function to generate test report
generate_test_report() {
    print_status "Generating test report..."
    
    local report_file="$OUTPUT_DIR/test_automation_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# Test Automation Report

**Date**: $(date)
**Test Type**: $TEST_TYPE
**Configuration**: $CONFIG_FILE_PATH
**Output Directory**: $OUTPUT_DIR

## Test Results

### Unit Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/coverage.html" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Coverage Report**: $(ls -t "$OUTPUT_DIR"/coverage.html 2>/dev/null | head -1 || echo "N/A")

### Integration Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/integration_results.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)

### Performance Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/performance_report.html" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Report**: $(ls -t "$OUTPUT_DIR"/performance_report.html 2>/dev/null | head -1 || echo "N/A")

### Security Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/gosec_report.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **gosec Report**: $(ls -t "$OUTPUT_DIR"/gosec_report.json 2>/dev/null | head -1 || echo "N/A")
- **trivy Report**: $(ls -t "$OUTPUT_DIR"/trivy_report.json 2>/dev/null | head -1 || echo "N/A")
- **nancy Report**: $(ls -t "$OUTPUT_DIR"/nancy_report.json 2>/dev/null | head -1 || echo "N/A")

### End-to-End Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/e2e_results.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)

### ML Model Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/ml_results.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)

## Test Coverage

$(if [ -f "$OUTPUT_DIR/coverage.txt" ]; then
    echo "### Coverage Summary"
    echo "\`\`\`"
    tail -1 "$OUTPUT_DIR/coverage.txt"
    echo "\`\`\`"
else
    echo "No coverage information available"
fi)

## Recommendations

1. Review all test results and address any failures
2. Fix any security vulnerabilities found
3. Improve test coverage if below 95%
4. Optimize performance if thresholds are not met
5. Regular test automation should be part of CI/CD pipeline

## Next Steps

1. Address critical and high severity issues immediately
2. Review medium severity issues and plan fixes
3. Implement continuous testing in CI/CD pipeline
4. Set up test monitoring and alerting
5. Regular test automation execution

EOF
    
    print_success "Test report saved to: $report_file"
}

# Function to cleanup
cleanup() {
    if [ "$SKIP_CLEANUP" = true ]; then
        print_status "Skipping cleanup operations"
        return 0
    fi
    
    print_status "Performing cleanup operations..."
    
    # Clean up temporary files
    rm -f coverage.out
    rm -f test.log
    rm -f performance_results.json
    rm -f security_results.json
    rm -f *.tmp
    rm -f *.log
    
    # Clean up test data
    rm -rf ./test_data
    rm -rf ./temp
    rm -rf ./logs
    
    # Stop test services if needed
    if command -v docker-compose &> /dev/null; then
        if [ -f "$PROJECT_ROOT/docker-compose.test.yml" ]; then
            print_status "Stopping test services"
            docker-compose -f "$PROJECT_ROOT/docker-compose.test.yml" down
        fi
    fi
    
    print_success "Cleanup completed"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--test-type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE_PATH="$2"
            shift 2
            ;;
        -o|--output-dir)
            OUTPUT_DIR="$2"
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
        --skip-cleanup)
            SKIP_CLEANUP=true
            shift
            ;;
        --skip-monitoring)
            SKIP_MONITORING=true
            shift
            ;;
        --help)
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

# Main execution
main() {
    print_status "Starting test automation for Risk Assessment Service"
    print_status "Test Type: $TEST_TYPE"
    print_status "Configuration: $CONFIG_FILE_PATH"
    print_status "Output Directory: $OUTPUT_DIR"
    print_status "Verbose: $VERBOSE"
    print_status "Parallel: $PARALLEL"
    print_status "Skip Cleanup: $SKIP_CLEANUP"
    print_status "Skip Monitoring: $SKIP_MONITORING"
    
    # Setup
    check_dependencies
    create_output_dir
    setup_test_environment
    
    # Set trap for cleanup
    trap cleanup EXIT
    
    # Run tests
    case $TEST_TYPE in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "security")
            run_security_tests
            ;;
        "e2e")
            run_e2e_tests
            ;;
        "ml")
            run_ml_tests
            ;;
        "all")
            run_all_tests
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            print_error "Valid types: unit, integration, performance, security, e2e, ml, all"
            exit 1
            ;;
    esac
    
    # Generate test report
    generate_test_report
    
    print_success "Test automation completed successfully"
}

# Run main function
main "$@"
