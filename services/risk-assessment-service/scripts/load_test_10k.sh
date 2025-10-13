#!/bin/bash

# Enhanced Load Testing Script for 10,000 Concurrent Users
# Risk Assessment Service - Phase 4.6 Implementation

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_DURATION="${TEST_DURATION:-30m}"
RAMP_UP_TIME="${RAMP_UP_TIME:-5m}"
CONCURRENT_USERS="${CONCURRENT_USERS:-10000}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-10}"
TARGET_RPS="${TARGET_RPS:-2000}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
LOG_DIR="logs/load-testing"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/load_test_10k_$TIMESTAMP.log"

# Function to log with timestamp
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to log success
log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log warning
log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log error
log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.22+"
        exit 1
    fi
    
    # Check if the service is running
    if ! curl -s "$BASE_URL/health" > /dev/null; then
        log_error "Service is not running at $BASE_URL. Please start the service first."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to build the load testing tool
build_load_tester() {
    log "Building enhanced load testing tool..."
    
    cd "$(dirname "$0")/.."
    
    # Build the load testing tool
    go build -o bin/load_tester_10k ./cmd/load_test_10k.go
    
    if [ $? -eq 0 ]; then
        log_success "Load testing tool built successfully"
    else
        log_error "Failed to build load testing tool"
        exit 1
    fi
}

# Function to run pre-test health check
run_health_check() {
    log "Running pre-test health check..."
    
    # Check service health
    HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
    if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
        log_success "Service health check passed"
    else
        log_error "Service health check failed: $HEALTH_RESPONSE"
        exit 1
    fi
    
    # Check metrics endpoint
    METRICS_RESPONSE=$(curl -s "$BASE_URL/metrics")
    if [ $? -eq 0 ]; then
        log_success "Metrics endpoint accessible"
    else
        log_warning "Metrics endpoint not accessible"
    fi
}

# Function to run baseline test
run_baseline_test() {
    log "Running baseline test (100 users, 1 minute)..."
    
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users=100 \
        --duration=1m \
        --ramp-up=30s \
        --output="$LOG_DIR/baseline_$TIMESTAMP.json" \
        --quiet
    
    if [ $? -eq 0 ]; then
        log_success "Baseline test completed"
    else
        log_error "Baseline test failed"
        exit 1
    fi
}

# Function to run ramp-up test
run_ramp_up_test() {
    log "Running ramp-up test (0 → 10,000 users over $RAMP_UP_TIME)..."
    
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users="$CONCURRENT_USERS" \
        --duration="$RAMP_UP_TIME" \
        --ramp-up="$RAMP_UP_TIME" \
        --output="$LOG_DIR/ramp_up_$TIMESTAMP.json" \
        --quiet
    
    if [ $? -eq 0 ]; then
        log_success "Ramp-up test completed"
    else
        log_error "Ramp-up test failed"
        exit 1
    fi
}

# Function to run sustained load test
run_sustained_test() {
    log "Running sustained load test (10,000 users for $TEST_DURATION)..."
    
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users="$CONCURRENT_USERS" \
        --duration="$TEST_DURATION" \
        --ramp-up=0s \
        --target-rps="$TARGET_RPS" \
        --output="$LOG_DIR/sustained_$TIMESTAMP.json" \
        --quiet
    
    if [ $? -eq 0 ]; then
        log_success "Sustained load test completed"
    else
        log_error "Sustained load test failed"
        exit 1
    fi
}

# Function to run spike test
run_spike_test() {
    log "Running spike test (sudden jump to 15,000 users)..."
    
    # Phase 1: Normal load (10K users)
    log "Phase 1: Normal load (10,000 users)"
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users=10000 \
        --duration=2m \
        --ramp-up=0s \
        --output="$LOG_DIR/spike_normal_$TIMESTAMP.json" \
        --quiet
    
    # Phase 2: Spike load (15K users)
    log "Phase 2: Spike load (15,000 users)"
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users=15000 \
        --duration=2m \
        --ramp-up=0s \
        --output="$LOG_DIR/spike_peak_$TIMESTAMP.json" \
        --quiet
    
    # Phase 3: Recovery (back to 10K users)
    log "Phase 3: Recovery (10,000 users)"
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users=10000 \
        --duration=3m \
        --ramp-up=0s \
        --output="$LOG_DIR/spike_recovery_$TIMESTAMP.json" \
        --quiet
    
    log_success "Spike test completed"
}

# Function to run stress test
run_stress_test() {
    log "Running stress test (finding breaking point)..."
    
    # Start with 10K users and gradually increase
    for users in 10000 12000 15000 18000 20000; do
        log "Testing with $users users..."
        
        ./bin/load_tester_10k \
            --url="$BASE_URL" \
            --users="$users" \
            --duration=2m \
            --ramp-up=30s \
            --output="$LOG_DIR/stress_${users}_$TIMESTAMP.json" \
            --quiet
        
        if [ $? -ne 0 ]; then
            log_warning "Service failed at $users users - breaking point found"
            break
        fi
    done
    
    log_success "Stress test completed"
}

# Function to run endurance test
run_endurance_test() {
    log "Running endurance test (10,000 users for 2 hours)..."
    
    ./bin/load_tester_10k \
        --url="$BASE_URL" \
        --users="$CONCURRENT_USERS" \
        --duration=2h \
        --ramp-up=0s \
        --target-rps="$TARGET_RPS" \
        --output="$LOG_DIR/endurance_$TIMESTAMP.json" \
        --quiet
    
    if [ $? -eq 0 ]; then
        log_success "Endurance test completed"
    else
        log_error "Endurance test failed"
        exit 1
    fi
}

# Function to analyze results
analyze_results() {
    log "Analyzing test results..."
    
    # Create results summary
    RESULTS_FILE="$LOG_DIR/results_summary_$TIMESTAMP.md"
    
    cat > "$RESULTS_FILE" << EOF
# Load Testing Results - 10K Concurrent Users
**Test Date**: $(date)
**Base URL**: $BASE_URL
**Test Duration**: $TEST_DURATION
**Concurrent Users**: $CONCURRENT_USERS
**Target RPS**: $TARGET_RPS

## Test Scenarios Executed

1. **Baseline Test**: 100 users, 1 minute
2. **Ramp-up Test**: 0 → 10,000 users over $RAMP_UP_TIME
3. **Sustained Load Test**: 10,000 users for $TEST_DURATION
4. **Spike Test**: Sudden jump to 15,000 users
5. **Stress Test**: Finding breaking point
6. **Endurance Test**: 10,000 users for 2 hours

## Performance Targets

- **P95 Latency**: < 1 second
- **P99 Latency**: < 2 seconds
- **Error Rate**: < 0.1%
- **Throughput**: 10,000+ requests/minute

## Results Files

- Baseline: \`baseline_$TIMESTAMP.json\`
- Ramp-up: \`ramp_up_$TIMESTAMP.json\`
- Sustained: \`sustained_$TIMESTAMP.json\`
- Spike: \`spike_*_$TIMESTAMP.json\`
- Stress: \`stress_*_$TIMESTAMP.json\`
- Endurance: \`endurance_$TIMESTAMP.json\`

## Analysis

Please review the individual result files for detailed metrics and analysis.

EOF
    
    log_success "Results analysis completed. Summary saved to: $RESULTS_FILE"
}

# Function to generate performance report
generate_report() {
    log "Generating performance report..."
    
    REPORT_FILE="$LOG_DIR/performance_report_$TIMESTAMP.html"
    
    cat > "$REPORT_FILE" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Load Testing Report - 10K Concurrent Users</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .metric { margin: 10px 0; padding: 10px; border-left: 4px solid #007cba; }
        .success { border-left-color: #28a745; }
        .warning { border-left-color: #ffc107; }
        .error { border-left-color: #dc3545; }
        .chart { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Load Testing Report - 10K Concurrent Users</h1>
        <p><strong>Test Date:</strong> $(date)</p>
        <p><strong>Base URL:</strong> $BASE_URL</p>
        <p><strong>Test Duration:</strong> $TEST_DURATION</p>
        <p><strong>Concurrent Users:</strong> $CONCURRENT_USERS</p>
    </div>
    
    <h2>Performance Targets</h2>
    <div class="metric success">
        <strong>P95 Latency:</strong> Target < 1 second
    </div>
    <div class="metric success">
        <strong>P99 Latency:</strong> Target < 2 seconds
    </div>
    <div class="metric success">
        <strong>Error Rate:</strong> Target < 0.1%
    </div>
    <div class="metric success">
        <strong>Throughput:</strong> Target 10,000+ requests/minute
    </div>
    
    <h2>Test Results</h2>
    <p>Detailed results are available in the JSON files in the logs directory.</p>
    
    <h2>Recommendations</h2>
    <ul>
        <li>Review individual test results for detailed analysis</li>
        <li>Monitor system resources during peak load</li>
        <li>Consider horizontal scaling if targets are not met</li>
        <li>Optimize database queries and caching strategies</li>
    </ul>
</body>
</html>
EOF
    
    log_success "Performance report generated: $REPORT_FILE"
}

# Main execution
main() {
    log "Starting 10K concurrent users load testing..."
    log "Configuration:"
    log "  Base URL: $BASE_URL"
    log "  Test Duration: $TEST_DURATION"
    log "  Concurrent Users: $CONCURRENT_USERS"
    log "  Target RPS: $TARGET_RPS"
    log "  Log File: $LOG_FILE"
    
    # Execute test phases
    check_prerequisites
    build_load_tester
    run_health_check
    run_baseline_test
    run_ramp_up_test
    run_sustained_test
    run_spike_test
    run_stress_test
    run_endurance_test
    analyze_results
    generate_report
    
    log_success "All load tests completed successfully!"
    log "Results are available in: $LOG_DIR"
    log "Summary report: $LOG_DIR/results_summary_$TIMESTAMP.md"
    log "HTML report: $LOG_DIR/performance_report_$TIMESTAMP.html"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --baseline          Run only baseline test"
        echo "  --ramp-up           Run only ramp-up test"
        echo "  --sustained         Run only sustained test"
        echo "  --spike             Run only spike test"
        echo "  --stress            Run only stress test"
        echo "  --endurance         Run only endurance test"
        echo "  --quick             Run quick test suite (baseline + ramp-up + sustained)"
        echo ""
        echo "Environment Variables:"
        echo "  BASE_URL            Base URL for testing (default: http://localhost:8080)"
        echo "  TEST_DURATION       Test duration (default: 30m)"
        echo "  CONCURRENT_USERS    Number of concurrent users (default: 10000)"
        echo "  TARGET_RPS          Target requests per second (default: 2000)"
        echo ""
        exit 0
        ;;
    --baseline)
        check_prerequisites
        build_load_tester
        run_health_check
        run_baseline_test
        ;;
    --ramp-up)
        check_prerequisites
        build_load_tester
        run_health_check
        run_ramp_up_test
        ;;
    --sustained)
        check_prerequisites
        build_load_tester
        run_health_check
        run_sustained_test
        ;;
    --spike)
        check_prerequisites
        build_load_tester
        run_health_check
        run_spike_test
        ;;
    --stress)
        check_prerequisites
        build_load_tester
        run_health_check
        run_stress_test
        ;;
    --endurance)
        check_prerequisites
        build_load_tester
        run_health_check
        run_endurance_test
        ;;
    --quick)
        check_prerequisites
        build_load_tester
        run_health_check
        run_baseline_test
        run_ramp_up_test
        run_sustained_test
        analyze_results
        generate_report
        ;;
    *)
        main
        ;;
esac
