#!/bin/bash

# Load testing script for the risk assessment service
# This script runs comprehensive load tests to validate the 1000 req/min target

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_URL="${SERVICE_URL:-http://localhost:8080}"
TEST_DURATION="${TEST_DURATION:-5m}"
CONCURRENT_USERS="${CONCURRENT_USERS:-20}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-50}"
TARGET_RPS="${TARGET_RPS:-16.67}" # 1000 req/min = 16.67 RPS
OUTPUT_DIR="load_test_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}🚀 Starting Load Testing Suite${NC}"
echo "=================================================="
echo -e "Service URL: ${YELLOW}$SERVICE_URL${NC}"
echo -e "Test Duration: ${YELLOW}$TEST_DURATION${NC}"
echo -e "Concurrent Users: ${YELLOW}$CONCURRENT_USERS${NC}"
echo -e "Target RPS: ${YELLOW}$TARGET_RPS${NC}"
echo "=================================================="

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to run a load test
run_load_test() {
    local test_name="$1"
    local test_type="$2"
    local extra_args="$3"
    
    echo -e "${BLUE}📊 Running $test_name${NC}"
    
    local output_file="$OUTPUT_DIR/${test_name}_${TIMESTAMP}.json"
    
    go run ./cmd/load_test.go \
        -url="$SERVICE_URL" \
        -duration="$TEST_DURATION" \
        -users="$CONCURRENT_USERS" \
        -requests="$REQUESTS_PER_USER" \
        -rps="$TARGET_RPS" \
        -type="$test_type" \
        -output="$output_file" \
        -verbose \
        $extra_args
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $test_name completed successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name failed${NC}"
        return 1
    fi
}

# Function to check service health
check_service_health() {
    echo -e "${YELLOW}🔍 Checking service health...${NC}"
    
    local health_url="$SERVICE_URL/health"
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$health_url" || echo "000")
    
    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✅ Service is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Service is not healthy (HTTP $response)${NC}"
        return 1
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    echo -e "${YELLOW}⏳ Waiting for service to be ready...${NC}"
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if check_service_health; then
            return 0
        fi
        
        echo -e "${YELLOW}Attempt $attempt/$max_attempts - waiting 10 seconds...${NC}"
        sleep 10
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ Service failed to become ready after $max_attempts attempts${NC}"
    return 1
}

# Function to analyze results
analyze_results() {
    echo -e "${BLUE}📈 Analyzing Load Test Results${NC}"
    echo "=================================================="
    
    local results_dir="$OUTPUT_DIR"
    local total_tests=0
    local passed_tests=0
    
    for result_file in "$results_dir"/*_${TIMESTAMP}.json; do
        if [ -f "$result_file" ]; then
            total_tests=$((total_tests + 1))
            
            local test_name=$(basename "$result_file" "_${TIMESTAMP}.json")
            echo -e "${YELLOW}📊 $test_name Results:${NC}"
            
            # Extract key metrics using jq if available
            if command -v jq &> /dev/null; then
                local total_requests=$(jq -r '.total_requests' "$result_file")
                local successful_requests=$(jq -r '.successful_requests' "$result_file")
                local failed_requests=$(jq -r '.failed_requests' "$result_file")
                local error_rate=$(jq -r '.error_rate' "$result_file")
                local requests_per_minute=$(jq -r '.requests_per_minute' "$result_file")
                local avg_response_time=$(jq -r '.average_response_time' "$result_file")
                
                echo -e "  Total Requests: $total_requests"
                echo -e "  Successful: $successful_requests"
                echo -e "  Failed: $failed_requests"
                echo -e "  Error Rate: $(echo "$error_rate * 100" | bc -l)%"
                echo -e "  Requests/Min: $requests_per_minute"
                echo -e "  Avg Response Time: $avg_response_time"
                
                # Check if test passed
                local error_rate_percent=$(echo "$error_rate * 100" | bc -l)
                if (( $(echo "$error_rate_percent < 5" | bc -l) )) && (( $(echo "$requests_per_minute >= 800" | bc -l) )); then
                    echo -e "  ${GREEN}✅ PASSED${NC}"
                    passed_tests=$((passed_tests + 1))
                else
                    echo -e "  ${RED}❌ FAILED${NC}"
                fi
            else
                echo -e "  ${YELLOW}⚠️  jq not available, showing raw results${NC}"
                echo -e "  $(head -20 "$result_file")"
            fi
            
            echo ""
        fi
    done
    
    echo -e "${BLUE}📊 Overall Results:${NC}"
    echo -e "  Tests Run: $total_tests"
    echo -e "  Tests Passed: $passed_tests"
    echo -e "  Tests Failed: $((total_tests - passed_tests))"
    
    if [ $passed_tests -eq $total_tests ] && [ $total_tests -gt 0 ]; then
        echo -e "${GREEN}🎉 All load tests passed!${NC}"
        return 0
    else
        echo -e "${RED}❌ Some load tests failed${NC}"
        return 1
    fi
}

# Function to generate report
generate_report() {
    echo -e "${BLUE}📝 Generating Load Test Report${NC}"
    
    local report_file="$OUTPUT_DIR/load_test_report_${TIMESTAMP}.md"
    
    cat > "$report_file" << EOF
# Load Test Report

**Generated:** $(date)
**Service URL:** $SERVICE_URL
**Test Duration:** $TEST_DURATION
**Concurrent Users:** $CONCURRENT_USERS
**Target RPS:** $TARGET_RPS

## Test Results

EOF

    for result_file in "$OUTPUT_DIR"/*_${TIMESTAMP}.json; do
        if [ -f "$result_file" ]; then
            local test_name=$(basename "$result_file" "_${TIMESTAMP}.json")
            echo "### $test_name" >> "$report_file"
            echo "" >> "$report_file"
            
            if command -v jq &> /dev/null; then
                jq -r '
                    "| Metric | Value |\n|--------|-------|\n" +
                    "| Total Requests | " + (.total_requests | tostring) + " |\n" +
                    "| Successful Requests | " + (.successful_requests | tostring) + " |\n" +
                    "| Failed Requests | " + (.failed_requests | tostring) + " |\n" +
                    "| Error Rate | " + ((.error_rate * 100) | tostring) + "% |\n" +
                    "| Requests/Minute | " + (.requests_per_minute | tostring) + " |\n" +
                    "| Average Response Time | " + .average_response_time + " |\n" +
                    "| Min Response Time | " + .min_response_time + " |\n" +
                    "| Max Response Time | " + .max_response_time + " |\n"
                ' "$result_file" >> "$report_file"
            else
                echo "Raw JSON results:" >> "$report_file"
                echo '```json' >> "$report_file"
                cat "$result_file" >> "$report_file"
                echo '```' >> "$report_file"
            fi
            
            echo "" >> "$report_file"
        fi
    done
    
    echo -e "${GREEN}📄 Report generated: $report_file${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}🎯 Load Testing Target: 1000 requests/minute${NC}"
    echo ""
    
    # Check if service is running
    if ! wait_for_service; then
        echo -e "${RED}❌ Cannot proceed with load tests - service is not available${NC}"
        exit 1
    fi
    
    # Run different types of load tests
    local test_results=()
    
    echo -e "${BLUE}🧪 Running Load Test Suite${NC}"
    echo "=================================================="
    
    # 1. Standard Load Test
    if run_load_test "standard_load" "load"; then
        test_results+=("standard_load:PASS")
    else
        test_results+=("standard_load:FAIL")
    fi
    
    # 2. Stress Test
    if run_load_test "stress_test" "stress"; then
        test_results+=("stress_test:PASS")
    else
        test_results+=("stress_test:FAIL")
    fi
    
    # 3. Spike Test
    if run_load_test "spike_test" "spike"; then
        test_results+=("spike_test:PASS")
    else
        test_results+=("spike_test:FAIL")
    fi
    
    # 4. High Load Test (targeting 1000 req/min)
    if run_load_test "high_load" "load" "-users=50 -requests=20 -rps=16.67"; then
        test_results+=("high_load:PASS")
    else
        test_results+=("high_load:FAIL")
    fi
    
    # 5. Sustained Load Test (longer duration)
    if run_load_test "sustained_load" "load" "-duration=10m -users=30 -requests=30"; then
        test_results+=("sustained_load:PASS")
    else
        test_results+=("sustained_load:FAIL")
    fi
    
    echo ""
    echo -e "${BLUE}📊 Load Test Suite Completed${NC}"
    echo "=================================================="
    
    # Analyze results
    analyze_results
    
    # Generate report
    generate_report
    
    # Final summary
    echo ""
    echo -e "${BLUE}🎯 Load Testing Summary${NC}"
    echo "=================================================="
    echo -e "Target: ${YELLOW}1000 requests/minute${NC}"
    echo -e "Results Directory: ${YELLOW}$OUTPUT_DIR${NC}"
    echo -e "Timestamp: ${YELLOW}$TIMESTAMP${NC}"
    
    # Check if we met our target
    local overall_success=true
    for result in "${test_results[@]}"; do
        if [[ "$result" == *":FAIL" ]]; then
            overall_success=false
            break
        fi
    done
    
    if [ "$overall_success" = true ]; then
        echo -e "${GREEN}🎉 All load tests passed! Service meets 1000 req/min target.${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some load tests failed. Service may not meet 1000 req/min target.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
