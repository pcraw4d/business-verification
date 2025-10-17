#!/bin/bash

# Performance Testing Runner for Risk Assessment Service
# This script runs comprehensive performance tests using Locust

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/test/performance/reports"
CONFIG_FILE="$SCRIPT_DIR/performance_config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
HOST="http://localhost:8080"
TEST_TYPE="all"
USERS=100
SPAWN_RATE=10
RUN_TIME="5m"
OUTPUT_DIR="$REPORTS_DIR"
VERBOSE=false
HEADLESS=false

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

Performance Testing Runner for Risk Assessment Service

OPTIONS:
    -h, --host HOST          Target host (default: http://localhost:8080)
    -t, --test-type TYPE     Test type: load, stress, spike, all (default: all)
    -u, --users USERS        Number of users (default: 100)
    -r, --spawn-rate RATE    Spawn rate (default: 10)
    -d, --duration TIME      Test duration (default: 5m)
    -o, --output-dir DIR     Output directory (default: ./reports)
    -v, --verbose            Verbose output
    -H, --headless           Run in headless mode
    --help                   Show this help message

EXAMPLES:
    # Run all performance tests
    $0

    # Run load testing with 200 users for 10 minutes
    $0 -t load -u 200 -d 10m

    # Run stress testing in headless mode
    $0 -t stress -u 500 -H

    # Run spike testing with custom host
    $0 -t spike -h http://staging.example.com -u 1000

EOF
}

# Function to check dependencies
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v locust &> /dev/null; then
        print_error "Locust is not installed. Please install it with: pip install locust"
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        print_error "Python 3 is not installed"
        exit 1
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

# Function to run load tests
run_load_tests() {
    print_status "Running load tests..."
    
    local test_name="load_test_$(date +%Y%m%d_%H%M%S)"
    local output_file="$OUTPUT_DIR/${test_name}.html"
    
    local locust_cmd="locust -f $SCRIPT_DIR/locustfile.py --host=$HOST --users=$USERS --spawn-rate=$SPAWN_RATE --run-time=$RUN_TIME --html=$output_file"
    
    if [ "$HEADLESS" = true ]; then
        locust_cmd="$locust_cmd --headless"
    fi
    
    if [ "$VERBOSE" = true ]; then
        locust_cmd="$locust_cmd --loglevel=DEBUG"
    fi
    
    print_status "Executing: $locust_cmd"
    
    if eval "$locust_cmd"; then
        print_success "Load tests completed successfully"
        print_success "Report saved to: $output_file"
    else
        print_error "Load tests failed"
        return 1
    fi
}

# Function to run stress tests
run_stress_tests() {
    print_status "Running stress tests..."
    
    local test_name="stress_test_$(date +%Y%m%d_%H%M%S)"
    local output_file="$OUTPUT_DIR/${test_name}.html"
    
    # Use higher user count for stress testing
    local stress_users=$((USERS * 5))
    local stress_spawn_rate=$((SPAWN_RATE * 5))
    
    local locust_cmd="locust -f $SCRIPT_DIR/locustfile.py --host=$HOST --users=$stress_users --spawn-rate=$stress_spawn_rate --run-time=$RUN_TIME --html=$output_file --user-class=StressTestUser"
    
    if [ "$HEADLESS" = true ]; then
        locust_cmd="$locust_cmd --headless"
    fi
    
    if [ "$VERBOSE" = true ]; then
        locust_cmd="$locust_cmd --loglevel=DEBUG"
    fi
    
    print_status "Executing: $locust_cmd"
    
    if eval "$locust_cmd"; then
        print_success "Stress tests completed successfully"
        print_success "Report saved to: $output_file"
    else
        print_error "Stress tests failed"
        return 1
    fi
}

# Function to run spike tests
run_spike_tests() {
    print_status "Running spike tests..."
    
    local test_name="spike_test_$(date +%Y%m%d_%H%M%S)"
    local output_file="$OUTPUT_DIR/${test_name}.html"
    
    # Use very high user count for spike testing
    local spike_users=$((USERS * 10))
    local spike_spawn_rate=$((SPAWN_RATE * 10))
    local spike_duration="2m"  # Shorter duration for spike tests
    
    local locust_cmd="locust -f $SCRIPT_DIR/locustfile.py --host=$HOST --users=$spike_users --spawn-rate=$spike_spawn_rate --run-time=$spike_duration --html=$output_file --user-class=StressTestUser"
    
    if [ "$HEADLESS" = true ]; then
        locust_cmd="$locust_cmd --headless"
    fi
    
    if [ "$VERBOSE" = true ]; then
        locust_cmd="$locust_cmd --loglevel=DEBUG"
    fi
    
    print_status "Executing: $locust_cmd"
    
    if eval "$locust_cmd"; then
        print_success "Spike tests completed successfully"
        print_success "Report saved to: $output_file"
    else
        print_error "Spike tests failed"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all performance tests..."
    
    local failed_tests=0
    
    # Run load tests
    if ! run_load_tests; then
        ((failed_tests++))
    fi
    
    # Wait between tests
    print_status "Waiting 30 seconds before next test..."
    sleep 30
    
    # Run stress tests
    if ! run_stress_tests; then
        ((failed_tests++))
    fi
    
    # Wait between tests
    print_status "Waiting 30 seconds before next test..."
    sleep 30
    
    # Run spike tests
    if ! run_spike_tests; then
        ((failed_tests++))
    fi
    
    if [ $failed_tests -eq 0 ]; then
        print_success "All performance tests completed successfully"
    else
        print_error "$failed_tests test(s) failed"
        return 1
    fi
}

# Function to generate summary report
generate_summary_report() {
    print_status "Generating summary report..."
    
    local summary_file="$OUTPUT_DIR/performance_summary_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$summary_file" << EOF
# Performance Testing Summary

**Date**: $(date)
**Host**: $HOST
**Test Type**: $TEST_TYPE
**Users**: $USERS
**Spawn Rate**: $SPAWN_RATE
**Duration**: $RUN_TIME

## Test Results

### Load Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/load_test_*.html" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Report**: $(ls -t "$OUTPUT_DIR"/load_test_*.html 2>/dev/null | head -1 || echo "N/A")

### Stress Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/stress_test_*.html" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Report**: $(ls -t "$OUTPUT_DIR"/stress_test_*.html 2>/dev/null | head -1 || echo "N/A")

### Spike Tests
- **Status**: $(if [ -f "$OUTPUT_DIR/spike_test_*.html" ]; then echo "✅ Completed"; else echo "❌ Failed"; fi)
- **Report**: $(ls -t "$OUTPUT_DIR"/spike_test_*.html 2>/dev/null | head -1 || echo "N/A")

## Performance Thresholds

- **Response Time P95**: < 1000ms
- **Response Time P99**: < 2000ms
- **Error Rate**: < 1%
- **Throughput**: > 1000 req/min

## Recommendations

1. Review individual test reports for detailed analysis
2. Monitor system resources during tests
3. Adjust thresholds based on business requirements
4. Consider scaling infrastructure if thresholds are not met

EOF
    
    print_success "Summary report saved to: $summary_file"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Kill any remaining locust processes
    pkill -f locust 2>/dev/null || true
    
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
        -u|--users)
            USERS="$2"
            shift 2
            ;;
        -r|--spawn-rate)
            SPAWN_RATE="$2"
            shift 2
            ;;
        -d|--duration)
            RUN_TIME="$2"
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
        -H|--headless)
            HEADLESS=true
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
    print_status "Starting performance testing for Risk Assessment Service"
    print_status "Host: $HOST"
    print_status "Test Type: $TEST_TYPE"
    print_status "Users: $USERS"
    print_status "Spawn Rate: $SPAWN_RATE"
    print_status "Duration: $RUN_TIME"
    print_status "Output Directory: $OUTPUT_DIR"
    
    # Setup
    check_dependencies
    create_output_dir
    check_service_health
    
    # Set trap for cleanup
    trap cleanup EXIT
    
    # Run tests based on type
    case $TEST_TYPE in
        "load")
            run_load_tests
            ;;
        "stress")
            run_stress_tests
            ;;
        "spike")
            run_spike_tests
            ;;
        "all")
            run_all_tests
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            print_error "Valid types: load, stress, spike, all"
            exit 1
            ;;
    esac
    
    # Generate summary report
    generate_summary_report
    
    print_success "Performance testing completed successfully"
}

# Run main function
main "$@"
