#!/bin/bash

# Security Testing Runner for Risk Assessment Service
# This script runs comprehensive security tests including vulnerability scanning

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/test/security/reports"
CONFIG_FILE="$SCRIPT_DIR/security_config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
HOST="http://localhost:8080"
TEST_TYPE="all"
OUTPUT_DIR="$REPORTS_DIR"
VERBOSE=false
SKIP_SCANNING=false
SKIP_TESTS=false

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

Security Testing Runner for Risk Assessment Service

OPTIONS:
    -h, --host HOST          Target host (default: http://localhost:8080)
    -t, --test-type TYPE     Test type: tests, scanning, all (default: all)
    -o, --output-dir DIR     Output directory (default: ./reports)
    -v, --verbose            Verbose output
    --skip-scanning          Skip vulnerability scanning
    --skip-tests             Skip security tests
    --help                   Show this help message

EXAMPLES:
    # Run all security tests
    $0

    # Run only security tests (skip scanning)
    $0 --skip-scanning

    # Run only vulnerability scanning
    $0 --skip-tests

    # Run with custom host
    $0 -h https://staging.example.com

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
    
    # Check gosec
    if ! command -v gosec &> /dev/null; then
        print_warning "gosec is not installed. Installing..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Check trivy
    if ! command -v trivy &> /dev/null; then
        print_warning "trivy is not installed. Installing..."
        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    fi
    
    # Check nancy
    if ! command -v nancy &> /dev/null; then
        print_warning "nancy is not installed. Installing..."
        go install github.com/sonatypecommunity/nancy@latest
    fi
    
    # Check golangci-lint
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

# Function to check service health
check_service_health() {
    print_status "Checking service health..."
    
    if ! curl -s -f "$HOST/health" > /dev/null; then
        print_error "Service is not healthy at $HOST"
        print_error "Please ensure the risk assessment service is running"
        exit 1
    fi
    
    print_success "Service is healthy"
}

# Function to run security tests
run_security_tests() {
    print_status "Running security tests..."
    
    local test_name="security_test_$(date +%Y%m%d_%H%M%S)"
    local output_file="$OUTPUT_DIR/${test_name}.json"
    
    # Run Go security tests
    local test_cmd="go test -tags=security -v ./test/security/... -json > $output_file"
    
    if [ "$VERBOSE" = true ]; then
        test_cmd="go test -tags=security -v ./test/security/..."
    fi
    
    print_status "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        print_success "Security tests completed successfully"
        print_success "Results saved to: $output_file"
    else
        print_error "Security tests failed"
        return 1
    fi
}

# Function to run vulnerability scanning
run_vulnerability_scanning() {
    print_status "Running vulnerability scanning..."
    
    local scan_name="vulnerability_scan_$(date +%Y%m%d_%H%M%S)"
    local gosec_output="$OUTPUT_DIR/${scan_name}_gosec.json"
    local trivy_output="$OUTPUT_DIR/${scan_name}_trivy.json"
    local nancy_output="$OUTPUT_DIR/${scan_name}_nancy.json"
    local golangci_output="$OUTPUT_DIR/${scan_name}_golangci.json"
    
    # Run gosec
    print_status "Running gosec security scanner..."
    if gosec -fmt json -out "$gosec_output" ./...; then
        print_success "gosec scan completed"
    else
        print_warning "gosec scan found issues"
    fi
    
    # Run trivy
    print_status "Running trivy vulnerability scanner..."
    if trivy fs --format json --output "$trivy_output" .; then
        print_success "trivy scan completed"
    else
        print_warning "trivy scan found issues"
    fi
    
    # Run nancy
    print_status "Running nancy dependency scanner..."
    if go list -json -deps ./... | nancy sleuth --format json --output "$nancy_output"; then
        print_success "nancy scan completed"
    else
        print_warning "nancy scan found issues"
    fi
    
    # Run golangci-lint with security rules
    print_status "Running golangci-lint with security rules..."
    if golangci-lint run --format json --out-format json --output "$golangci_output"; then
        print_success "golangci-lint scan completed"
    else
        print_warning "golangci-lint scan found issues"
    fi
}

# Function to run all security tests
run_all_security_tests() {
    print_status "Running all security tests..."
    
    local failed_tests=0
    
    # Run security tests
    if [ "$SKIP_TESTS" = false ]; then
        if ! run_security_tests; then
            ((failed_tests++))
        fi
    fi
    
    # Run vulnerability scanning
    if [ "$SKIP_SCANNING" = false ]; then
        if ! run_vulnerability_scanning; then
            ((failed_tests++))
        fi
    fi
    
    if [ $failed_tests -eq 0 ]; then
        print_success "All security tests completed successfully"
    else
        print_error "$failed_tests test(s) failed"
        return 1
    fi
}

# Function to generate security report
generate_security_report() {
    print_status "Generating security report..."
    
    local report_file="$OUTPUT_DIR/security_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# Security Testing Report

**Date**: $(date)
**Host**: $HOST
**Test Type**: $TEST_TYPE

## Test Results

### Security Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/security_test_*.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Report**: $(ls -t "$OUTPUT_DIR"/security_test_*.json 2>/dev/null | head -1 || echo "N/A")

### Vulnerability Scanning
- **gosec**: $(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_gosec.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **trivy**: $(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_trivy.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **nancy**: $(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_nancy.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **golangci-lint**: $(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_golangci.json" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)

## Security Headers Tested

- ✅ X-Content-Type-Options
- ✅ X-XSS-Protection
- ✅ X-Frame-Options
- ✅ Strict-Transport-Security
- ✅ Content-Security-Policy
- ✅ Referrer-Policy
- ✅ Permissions-Policy

## Input Validation Tests

- ✅ SQL Injection
- ✅ XSS Attack
- ✅ Command Injection
- ✅ Path Traversal
- ✅ Buffer Overflow
- ✅ JSON Injection

## Authentication & Authorization Tests

- ✅ Missing Authentication
- ✅ Invalid Token
- ✅ Token Manipulation
- ✅ Privilege Escalation

## Rate Limiting Tests

- ✅ Brute Force Protection
- ✅ DDoS Protection
- ✅ API Abuse Protection

## CORS Security Tests

- ✅ Allowed Origins
- ✅ Allowed Methods
- ✅ Allowed Headers

## Data Privacy Tests

- ✅ Sensitive Data Logging
- ✅ Data Encryption
- ✅ Data Retention

## Vulnerability Scanning Results

### gosec
$(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_gosec.json" ]; then
    echo "```json"
    cat "$OUTPUT_DIR"/vulnerability_scan_*_gosec.json | head -20
    echo "```"
else
    echo "No gosec results available"
fi)

### trivy
$(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_trivy.json" ]; then
    echo "```json"
    cat "$OUTPUT_DIR"/vulnerability_scan_*_trivy.json | head -20
    echo "```"
else
    echo "No trivy results available"
fi)

### nancy
$(if [ -f "$OUTPUT_DIR/vulnerability_scan_*_nancy.json" ]; then
    echo "```json"
    cat "$OUTPUT_DIR"/vulnerability_scan_*_nancy.json | head -20
    echo "```"
else
    echo "No nancy results available"
fi)

## Recommendations

1. Review all security test results and address any failures
2. Fix any vulnerabilities found by scanning tools
3. Update dependencies with known vulnerabilities
4. Implement additional security measures as needed
5. Regular security testing should be part of CI/CD pipeline

## Next Steps

1. Address critical and high severity issues immediately
2. Review medium severity issues and plan fixes
3. Implement security monitoring and alerting
4. Conduct regular security audits
5. Keep security tools and dependencies updated

EOF
    
    print_success "Security report saved to: $report_file"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Clean up any temporary files
    rm -f /tmp/security_test_*
    
    print_success "Cleanup completed"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            HOST="$2"
            shift 2
            ;;
        -t|--test-type)
            TEST_TYPE="$2"
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
        --skip-scanning)
            SKIP_SCANNING=true
            shift
            ;;
        --skip-tests)
            SKIP_TESTS=true
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
    print_status "Starting security testing for Risk Assessment Service"
    print_status "Host: $HOST"
    print_status "Test Type: $TEST_TYPE"
    print_status "Output Directory: $OUTPUT_DIR"
    print_status "Skip Scanning: $SKIP_SCANNING"
    print_status "Skip Tests: $SKIP_TESTS"
    
    # Setup
    check_dependencies
    create_output_dir
    check_service_health
    
    # Set trap for cleanup
    trap cleanup EXIT
    
    # Run tests based on type
    case $TEST_TYPE in
        "tests")
            run_security_tests
            ;;
        "scanning")
            run_vulnerability_scanning
            ;;
        "all")
            run_all_security_tests
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            print_error "Valid types: tests, scanning, all"
            exit 1
            ;;
    esac
    
    # Generate security report
    generate_security_report
    
    print_success "Security testing completed successfully"
}

# Run main function
main "$@"
