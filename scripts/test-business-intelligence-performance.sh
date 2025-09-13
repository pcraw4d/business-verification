#!/bin/bash

# Business Intelligence Performance Testing Script
# Tests the performance of business intelligence endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
BASE_URL="http://localhost:8080"
CONCURRENT_REQUESTS=10
TOTAL_REQUESTS=100
TIMEOUT=30

# Test results directory
mkdir -p "$TEST_DIR/test-results"

# Function to print colored output
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to check if server is running
check_server() {
    print_status "Checking if server is running..."
    
    if curl -s --connect-timeout 5 "$BASE_URL/health" > /dev/null 2>&1; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not running or not accessible"
        return 1
    fi
}

# Function to start server if not running
start_server() {
    print_status "Starting server..."
    
    cd "$TEST_DIR"
    
    # Check if server is already running
    if check_server; then
        print_success "Server is already running"
        return 0
    fi
    
    # Start server in background
    print_status "Starting server in background..."
    go run cmd/api-enhanced/main-enhanced-with-database-classification.go &
    SERVER_PID=$!
    
    # Wait for server to start
    print_status "Waiting for server to start..."
    sleep 5
    
    # Check if server started successfully
    if check_server; then
        print_success "Server started successfully (PID: $SERVER_PID)"
        return 0
    else
        print_error "Failed to start server"
        return 1
    fi
}

# Function to stop server
stop_server() {
    if [ ! -z "$SERVER_PID" ]; then
        print_status "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        print_success "Server stopped"
    fi
}

# Function to test endpoint performance
test_endpoint_performance() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local test_name="$4"
    
    print_status "Testing performance for $test_name..."
    
    # Create test data file
    local test_data_file="/tmp/bi_perf_test_${test_name}.json"
    echo "$data" > "$test_data_file"
    
    # Run performance test
    local start_time=$(date +%s.%N)
    
    # Use curl with timing
    local response=$(curl -s -w "\n%{time_total}\n%{time_connect}\n%{time_starttransfer}\n%{http_code}" \
        -X "$method" \
        -H "Content-Type: application/json" \
        -d @"$test_data_file" \
        --connect-timeout $TIMEOUT \
        --max-time $TIMEOUT \
        "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc)
    
    # Parse response
    local http_code=$(echo "$response" | tail -n1)
    local time_starttransfer=$(echo "$response" | tail -n2 | head -n1)
    local time_connect=$(echo "$response" | tail -n3 | head -n1)
    local time_total=$(echo "$response" | tail -n4 | head -n1)
    local response_body=$(echo "$response" | head -n -4)
    
    # Clean up test data file
    rm -f "$test_data_file"
    
    # Record results
    echo "$test_name,$endpoint,$method,$http_code,$time_total,$time_connect,$time_starttransfer,$total_time" >> "$TEST_DIR/test-results/bi-performance-results.csv"
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        print_success "$test_name: HTTP $http_code, Total: ${time_total}s, Connect: ${time_connect}s, Transfer: ${time_starttransfer}s"
    else
        print_warning "$test_name: HTTP $http_code, Total: ${time_total}s, Connect: ${time_connect}s, Transfer: ${time_starttransfer}s"
    fi
}

# Function to test concurrent requests
test_concurrent_requests() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local test_name="$4"
    local concurrent="$5"
    
    print_status "Testing concurrent requests for $test_name ($concurrent concurrent)..."
    
    # Create test data file
    local test_data_file="/tmp/bi_concurrent_test_${test_name}.json"
    echo "$data" > "$test_data_file"
    
    local start_time=$(date +%s.%N)
    
    # Run concurrent requests
    local pids=()
    for i in $(seq 1 $concurrent); do
        (
            curl -s -w "\n%{time_total}\n%{http_code}" \
                -X "$method" \
                -H "Content-Type: application/json" \
                -d @"$test_data_file" \
                --connect-timeout $TIMEOUT \
                --max-time $TIMEOUT \
                "$BASE_URL$endpoint" > "/tmp/bi_concurrent_response_${i}.txt" 2>/dev/null || echo "ERROR" > "/tmp/bi_concurrent_response_${i}.txt"
        ) &
        pids+=($!)
    done
    
    # Wait for all requests to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc)
    
    # Analyze results
    local success_count=0
    local total_response_time=0
    local min_time=999999
    local max_time=0
    
    for i in $(seq 1 $concurrent); do
        local response_file="/tmp/bi_concurrent_response_${i}.txt"
        if [ -f "$response_file" ]; then
            local http_code=$(tail -n1 "$response_file")
            local response_time=$(tail -n2 "$response_file" | head -n1)
            
            if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
                ((success_count++))
            fi
            
            if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                total_response_time=$(echo "$total_response_time + $response_time" | bc)
                if (( $(echo "$response_time < $min_time" | bc -l) )); then
                    min_time=$response_time
                fi
                if (( $(echo "$response_time > $max_time" | bc -l) )); then
                    max_time=$response_time
                fi
            fi
            
            rm -f "$response_file"
        fi
    done
    
    local avg_response_time=$(echo "scale=3; $total_response_time / $concurrent" | bc)
    local success_rate=$(echo "scale=2; $success_count * 100 / $concurrent" | bc)
    
    # Record results
    echo "$test_name,$endpoint,$method,$concurrent,$success_count,$success_rate,$avg_response_time,$min_time,$max_time,$total_time" >> "$TEST_DIR/test-results/bi-concurrent-results.csv"
    
    print_success "$test_name: $success_count/$concurrent successful ($success_rate%), Avg: ${avg_response_time}s, Min: ${min_time}s, Max: ${max_time}s, Total: ${total_time}s"
    
    # Clean up test data file
    rm -f "$test_data_file"
}

# Function to run load test
run_load_test() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local test_name="$4"
    
    print_status "Running load test for $test_name ($TOTAL_REQUESTS requests)..."
    
    # Create test data file
    local test_data_file="/tmp/bi_load_test_${test_name}.json"
    echo "$data" > "$test_data_file"
    
    local start_time=$(date +%s.%N)
    
    # Run load test
    local success_count=0
    local total_response_time=0
    local min_time=999999
    local max_time=0
    
    for i in $(seq 1 $TOTAL_REQUESTS); do
        local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -d @"$test_data_file" \
            --connect-timeout $TIMEOUT \
            --max-time $TIMEOUT \
            "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
        
        local http_code=$(echo "$response" | tail -n1)
        local response_time=$(echo "$response" | tail -n2 | head -n1)
        
        if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
            ((success_count++))
        fi
        
        if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
            total_response_time=$(echo "$total_response_time + $response_time" | bc)
            if (( $(echo "$response_time < $min_time" | bc -l) )); then
                min_time=$response_time
            fi
            if (( $(echo "$response_time > $max_time" | bc -l) )); then
                max_time=$response_time
            fi
        fi
        
        # Progress indicator
        if [ $((i % 10)) -eq 0 ]; then
            echo -n "."
        fi
    done
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc)
    local avg_response_time=$(echo "scale=3; $total_response_time / $TOTAL_REQUESTS" | bc)
    local success_rate=$(echo "scale=2; $success_count * 100 / $TOTAL_REQUESTS" | bc)
    local requests_per_second=$(echo "scale=2; $TOTAL_REQUESTS / $total_time" | bc)
    
    # Record results
    echo "$test_name,$endpoint,$method,$TOTAL_REQUESTS,$success_count,$success_rate,$avg_response_time,$min_time,$max_time,$total_time,$requests_per_second" >> "$TEST_DIR/test-results/bi-load-results.csv"
    
    echo ""
    print_success "$test_name: $success_count/$TOTAL_REQUESTS successful ($success_rate%), Avg: ${avg_response_time}s, Min: ${min_time}s, Max: ${max_time}s, Total: ${total_time}s, RPS: ${requests_per_second}"
    
    # Clean up test data file
    rm -f "$test_data_file"
}

# Function to generate performance report
generate_performance_report() {
    local report_file="$TEST_DIR/test-results/business-intelligence-performance-report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating performance report: $report_file"
    
    cat > "$report_file" << EOF
Business Intelligence Performance Testing Report
===============================================
Generated: $(date)
Test Suite: Performance Testing
Version: 1.0.0

Test Configuration:
- Base URL: $BASE_URL
- Concurrent Requests: $CONCURRENT_REQUESTS
- Total Requests: $TOTAL_REQUESTS
- Timeout: ${TIMEOUT}s

Test Results Summary:
EOF

    # Add performance results
    if [ -f "$TEST_DIR/test-results/bi-performance-results.csv" ]; then
        echo "" >> "$report_file"
        echo "Individual Endpoint Performance:" >> "$report_file"
        echo "================================" >> "$report_file"
        cat "$TEST_DIR/test-results/bi-performance-results.csv" >> "$report_file"
    fi
    
    # Add concurrent results
    if [ -f "$TEST_DIR/test-results/bi-concurrent-results.csv" ]; then
        echo "" >> "$report_file"
        echo "Concurrent Request Performance:" >> "$report_file"
        echo "===============================" >> "$report_file"
        cat "$TEST_DIR/test-results/bi-concurrent-results.csv" >> "$report_file"
    fi
    
    # Add load test results
    if [ -f "$TEST_DIR/test-results/bi-load-results.csv" ]; then
        echo "" >> "$report_file"
        echo "Load Test Performance:" >> "$report_file"
        echo "=====================" >> "$report_file"
        cat "$TEST_DIR/test-results/bi-load-results.csv" >> "$report_file"
    fi
    
    echo "" >> "$report_file"
    echo "Performance Analysis:" >> "$report_file"
    echo "====================" >> "$report_file"
    echo "- Response times under 1 second are considered good" >> "$report_file"
    echo "- Success rates above 95% are considered acceptable" >> "$report_file"
    echo "- Requests per second above 10 are considered good" >> "$report_file"
    
    print_success "Performance report generated: $report_file"
}

# Main execution
main() {
    print_status "🚀 Business Intelligence Performance Testing"
    print_status "============================================="
    
    # Initialize CSV files
    echo "Test Name,Endpoint,Method,HTTP Code,Total Time,Connect Time,Transfer Time,Wall Time" > "$TEST_DIR/test-results/bi-performance-results.csv"
    echo "Test Name,Endpoint,Method,Concurrent,Success Count,Success Rate,Avg Response Time,Min Time,Max Time,Total Time" > "$TEST_DIR/test-results/bi-concurrent-results.csv"
    echo "Test Name,Endpoint,Method,Total Requests,Success Count,Success Rate,Avg Response Time,Min Time,Max Time,Total Time,Requests Per Second" > "$TEST_DIR/test-results/bi-load-results.csv"
    
    # Start server
    if ! start_server; then
        print_error "Failed to start server. Exiting."
        exit 1
    fi
    
    # Test data
    local market_analysis_data='{
        "business_id": "test-business-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        },
        "options": {
            "include_competitors": true,
            "include_trends": true,
            "include_forecasts": true
        }
    }'
    
    local competitive_analysis_data='{
        "business_id": "test-business-123",
        "competitors": ["competitor1", "competitor2"],
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        },
        "options": {
            "include_market_share": true,
            "include_pricing": true,
            "include_features": true
        }
    }'
    
    local growth_analytics_data='{
        "business_id": "test-business-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        },
        "options": {
            "include_revenue": true,
            "include_customers": true,
            "include_metrics": true
        }
    }'
    
    # Run performance tests
    print_status "📊 Running Performance Tests"
    print_status "============================"
    
    # Individual endpoint performance tests
    test_endpoint_performance "/v2/business-intelligence/market-analysis" "POST" "$market_analysis_data" "Market Analysis"
    test_endpoint_performance "/v2/business-intelligence/competitive-analysis" "POST" "$competitive_analysis_data" "Competitive Analysis"
    test_endpoint_performance "/v2/business-intelligence/growth-analytics" "POST" "$growth_analytics_data" "Growth Analytics"
    
    # Concurrent request tests
    print_status "🔄 Running Concurrent Request Tests"
    print_status "===================================="
    
    test_concurrent_requests "/v2/business-intelligence/market-analysis" "POST" "$market_analysis_data" "Market Analysis" 5
    test_concurrent_requests "/v2/business-intelligence/competitive-analysis" "POST" "$competitive_analysis_data" "Competitive Analysis" 5
    test_concurrent_requests "/v2/business-intelligence/growth-analytics" "POST" "$growth_analytics_data" "Growth Analytics" 5
    
    # Load tests
    print_status "⚡ Running Load Tests"
    print_status "===================="
    
    run_load_test "/v2/business-intelligence/market-analysis" "POST" "$market_analysis_data" "Market Analysis"
    run_load_test "/v2/business-intelligence/competitive-analysis" "POST" "$competitive_analysis_data" "Competitive Analysis"
    run_load_test "/v2/business-intelligence/growth-analytics" "POST" "$growth_analytics_data" "Growth Analytics"
    
    # Generate report
    generate_performance_report
    
    # Stop server
    stop_server
    
    print_status "📋 Final Test Summary"
    print_status "===================="
    print_success "Performance testing completed successfully!"
    print_status "Check the test-results directory for detailed reports."
}

# Trap to ensure server is stopped on exit
trap stop_server EXIT

# Run main function
main "$@"
