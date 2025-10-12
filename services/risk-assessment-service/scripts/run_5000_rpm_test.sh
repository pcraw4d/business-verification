#!/bin/bash

# Enhanced Risk Assessment Service - 5000 RPM Performance Test Script
# This script runs comprehensive performance tests to achieve 5000+ requests per minute

set -e

# Configuration
SERVICE_URL="http://localhost:8080"
TEST_DURATION="10m"
CONCURRENT_USERS=200
TARGET_RPS=83.33  # 5000 RPM / 60 seconds
TARGET_RPM=5000
MAX_LATENCY="2s"
MAX_ERROR_RATE=0.01  # 1%

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

# Print banner
print_banner() {
    echo "================================================================"
    echo "🚀 Enhanced Risk Assessment Service - 5000 RPM Performance Test"
    echo "================================================================"
    echo "Target: 5000+ requests per minute"
    echo "Duration: $TEST_DURATION"
    echo "Concurrent Users: $CONCURRENT_USERS"
    echo "Max Latency: $MAX_LATENCY"
    echo "Max Error Rate: $(echo "$MAX_ERROR_RATE * 100" | bc)%"
    echo "================================================================"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if service is running
    if ! curl -s "$SERVICE_URL/health" > /dev/null; then
        log_error "Service is not running at $SERVICE_URL"
        log_info "Please start the service first: go run cmd/main.go"
        exit 1
    fi
    
    # Check if performance test binary exists
    if [ ! -f "./performance_test" ]; then
        log_info "Building performance test binary..."
        go build -o performance_test cmd/performance_test.go
    fi
    
    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        log_warning "bc command not found. Installing..."
        if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y bc
        elif command -v yum &> /dev/null; then
            sudo yum install -y bc
        elif command -v brew &> /dev/null; then
            brew install bc
        else
            log_error "Cannot install bc. Please install it manually."
            exit 1
        fi
    fi
    
    log_success "Prerequisites check completed"
}

# Run baseline test
run_baseline_test() {
    log_info "Running baseline performance test..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="2m" \
        --users=50 \
        --rps=50 \
        --pattern="constant" \
        --max-latency="$MAX_LATENCY" \
        --max-error-rate="$MAX_ERROR_RATE" \
        --verbose \
        --output="baseline_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Baseline test completed successfully"
    else
        log_error "Baseline test failed"
        exit 1
    fi
}

# Run constant load test
run_constant_test() {
    log_info "Running constant load test (5000 RPM target)..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="$TEST_DURATION" \
        --users="$CONCURRENT_USERS" \
        --rps="$TARGET_RPS" \
        --rpm="$TARGET_RPM" \
        --pattern="constant" \
        --max-latency="$MAX_LATENCY" \
        --max-error-rate="$MAX_ERROR_RATE" \
        --verbose \
        --output="constant_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Constant load test completed successfully"
    else
        log_warning "Constant load test did not meet all targets"
    fi
}

# Run ramp test
run_ramp_test() {
    log_info "Running ramp load test..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="$TEST_DURATION" \
        --users="$CONCURRENT_USERS" \
        --rps="$TARGET_RPS" \
        --rpm="$TARGET_RPM" \
        --pattern="ramp" \
        --ramp-up="2m" \
        --steady-state="6m" \
        --ramp-down="2m" \
        --max-latency="$MAX_LATENCY" \
        --max-error-rate="$MAX_ERROR_RATE" \
        --verbose \
        --output="ramp_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Ramp load test completed successfully"
    else
        log_warning "Ramp load test did not meet all targets"
    fi
}

# Run spike test
run_spike_test() {
    log_info "Running spike load test..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="$TEST_DURATION" \
        --users="$CONCURRENT_USERS" \
        --rps="$TARGET_RPS" \
        --rpm="$TARGET_RPM" \
        --pattern="spike" \
        --spike-multiplier="3.0" \
        --max-latency="$MAX_LATENCY" \
        --max-error-rate="$MAX_ERROR_RATE" \
        --verbose \
        --output="spike_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Spike load test completed successfully"
    else
        log_warning "Spike load test did not meet all targets"
    fi
}

# Run sine wave test
run_sine_test() {
    log_info "Running sine wave load test..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="$TEST_DURATION" \
        --users="$CONCURRENT_USERS" \
        --rps="$TARGET_RPS" \
        --rpm="$TARGET_RPM" \
        --pattern="sine" \
        --sine-amplitude="0.3" \
        --sine-period="3m" \
        --max-latency="$MAX_LATENCY" \
        --max-error-rate="$MAX_ERROR_RATE" \
        --verbose \
        --output="sine_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Sine wave load test completed successfully"
    else
        log_warning "Sine wave load test did not meet all targets"
    fi
}

# Run stress test
run_stress_test() {
    log_info "Running stress test to find breaking point..."
    
    ./performance_test \
        --url="$SERVICE_URL" \
        --duration="5m" \
        --users=500 \
        --rps=200 \
        --rpm=12000 \
        --pattern="constant" \
        --max-latency="5s" \
        --max-error-rate="0.05" \
        --verbose \
        --output="stress_results.json"
    
    if [ $? -eq 0 ]; then
        log_success "Stress test completed successfully"
    else
        log_warning "Stress test reached system limits (expected)"
    fi
}

# Generate performance report
generate_report() {
    log_info "Generating performance report..."
    
    echo "================================================================"
    echo "📊 PERFORMANCE TEST SUMMARY REPORT"
    echo "================================================================"
    echo "Test Date: $(date)"
    echo "Service URL: $SERVICE_URL"
    echo "Target: 5000+ requests per minute"
    echo ""
    
    # Check if result files exist and extract key metrics
    if [ -f "constant_results.json" ]; then
        echo "✅ Constant Load Test: COMPLETED"
        # In a real implementation, you would parse the JSON and extract metrics
        echo "   - Results saved to: constant_results.json"
    fi
    
    if [ -f "ramp_results.json" ]; then
        echo "✅ Ramp Load Test: COMPLETED"
        echo "   - Results saved to: ramp_results.json"
    fi
    
    if [ -f "spike_results.json" ]; then
        echo "✅ Spike Load Test: COMPLETED"
        echo "   - Results saved to: spike_results.json"
    fi
    
    if [ -f "sine_results.json" ]; then
        echo "✅ Sine Wave Load Test: COMPLETED"
        echo "   - Results saved to: sine_results.json"
    fi
    
    if [ -f "stress_results.json" ]; then
        echo "✅ Stress Test: COMPLETED"
        echo "   - Results saved to: stress_results.json"
    fi
    
    echo ""
    echo "📈 Key Metrics:"
    echo "   - Target RPM: $TARGET_RPM"
    echo "   - Concurrent Users: $CONCURRENT_USERS"
    echo "   - Max Latency: $MAX_LATENCY"
    echo "   - Max Error Rate: $(echo "$MAX_ERROR_RATE * 100" | bc)%"
    echo ""
    echo "📁 All results saved to JSON files for detailed analysis"
    echo "================================================================"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Remove temporary files if needed
    # rm -f temp_*.json
    log_success "Cleanup completed"
}

# Main execution
main() {
    print_banner
    
    # Set up signal handling
    trap cleanup EXIT
    
    # Check prerequisites
    check_prerequisites
    
    # Run tests
    log_info "Starting performance test suite..."
    
    # Run baseline test first
    run_baseline_test
    
    # Run main performance tests
    run_constant_test
    run_ramp_test
    run_spike_test
    run_sine_test
    
    # Run stress test
    run_stress_test
    
    # Generate report
    generate_report
    
    log_success "Performance test suite completed!"
    log_info "Check the JSON result files for detailed metrics"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --url)
            SERVICE_URL="$2"
            shift 2
            ;;
        --duration)
            TEST_DURATION="$2"
            shift 2
            ;;
        --users)
            CONCURRENT_USERS="$2"
            shift 2
            ;;
        --target-rpm)
            TARGET_RPM="$2"
            TARGET_RPS=$(echo "scale=2; $TARGET_RPM / 60" | bc)
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --url URL              Service URL (default: http://localhost:8080)"
            echo "  --duration DURATION    Test duration (default: 10m)"
            echo "  --users USERS          Concurrent users (default: 200)"
            echo "  --target-rpm RPM       Target requests per minute (default: 5000)"
            echo "  --help                 Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Run with defaults"
            echo "  $0 --target-rpm 10000                # Target 10,000 RPM"
            echo "  $0 --users 500 --duration 15m        # 500 users for 15 minutes"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main
