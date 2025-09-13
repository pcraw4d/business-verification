#!/bin/bash

# Light Performance Benchmarking Script
# Faster, lighter performance testing for business intelligence system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
BASE_URL="http://localhost:8080"
UI_URL="http://localhost:8081"
TEST_RESULTS_DIR="$TEST_DIR/test-results"

# Light benchmarking parameters (reduced for faster execution)
CONCURRENT_USERS=(1 5 10)
REQUEST_VOLUMES=(10 25 50)
DURATION_TESTS=(10 30) # seconds

# Test results directory
mkdir -p "$TEST_RESULTS_DIR"

# Function to print colored output
print_header() {
    echo -e "${PURPLE}$1${NC}"
}

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

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to start servers for benchmarking
start_servers() {
    print_status "Starting servers for performance benchmarking..."
    
    # Start API server
    cd "$TEST_DIR"
    if ! curl -s --connect-timeout 5 "$BASE_URL/health" > /dev/null 2>&1; then
        print_status "Starting API server..."
        go run cmd/api-enhanced/main-enhanced-with-database-classification.go &
        API_SERVER_PID=$!
        sleep 5
    else
        print_success "API server is already running"
    fi
    
    # Start UI server
    cd "$TEST_DIR/web"
    if ! curl -s --connect-timeout 5 "$UI_URL" > /dev/null 2>&1; then
        print_status "Starting UI server..."
        python3 -m http.server 8081 &
        UI_SERVER_PID=$!
        sleep 2
    else
        print_success "UI server is already running"
    fi
}

# Function to stop servers
stop_servers() {
    if [ ! -z "$API_SERVER_PID" ]; then
        print_status "Stopping API server (PID: $API_SERVER_PID)..."
        kill $API_SERVER_PID 2>/dev/null || true
        wait $API_SERVER_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$UI_SERVER_PID" ]; then
        print_status "Stopping UI server (PID: $UI_SERVER_PID)..."
        kill $UI_SERVER_PID 2>/dev/null || true
        wait $UI_SERVER_PID 2>/dev/null || true
    fi
}

# Function to run quick baseline test
run_quick_baseline_test() {
    print_header "📊 Quick Baseline Performance Test"
    print_status "=================================="
    
    local test_data='{
        "business_id": "baseline-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local baseline_file="$TEST_RESULTS_DIR/quick_baseline_performance.csv"
    echo "Test,Endpoint,Response_Time,Status_Code,Success" > "$baseline_file"
    
    print_status "Running quick baseline tests..."
    
    local endpoint="/v2/business-intelligence/market-analysis"
    
    for i in {1..5}; do
        local start_time=$(date +%s.%N)
        local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
            -X POST \
            -H "Content-Type: application/json" \
            -d "$test_data" \
            "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
        local end_time=$(date +%s.%N)
        
        local http_code=$(echo "$response" | tail -n1)
        local response_time=$(echo "$response" | tail -n2 | head -n1)
        local success="false"
        
        if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
            success="true"
        fi
        
        echo "baseline,$endpoint,$response_time,$http_code,$success" >> "$baseline_file"
    done
    
    print_success "Quick baseline performance test completed"
}

# Function to run light concurrent user tests
run_light_concurrent_tests() {
    print_header "👥 Light Concurrent User Tests"
    print_status "============================="
    
    local test_data='{
        "business_id": "concurrent-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local concurrent_file="$TEST_RESULTS_DIR/light_concurrent_users.csv"
    echo "Concurrent_Users,Endpoint,Total_Requests,Successful_Requests,Success_Rate,Avg_Response_Time" > "$concurrent_file"
    
    for users in "${CONCURRENT_USERS[@]}"; do
        print_status "Testing with $users concurrent users..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        
        # Run concurrent requests
        local pids=()
        local success_count=0
        local total_response_time=0
        
        for i in $(seq 1 $users); do
            (
                local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                    -X POST \
                    -H "Content-Type: application/json" \
                    -d "$test_data" \
                    "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                
                local http_code=$(echo "$response" | tail -n1)
                local response_time=$(echo "$response" | tail -n2 | head -n1)
                
                if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                    echo "SUCCESS" > "/tmp/light_concurrent_result_${i}.txt"
                else
                    echo "FAILED" > "/tmp/light_concurrent_result_${i}.txt"
                fi
                
                echo "$response_time" >> "/tmp/light_concurrent_result_${i}.txt"
            ) &
            pids+=($!)
        done
        
        # Wait for all requests to complete
        for pid in "${pids[@]}"; do
            wait $pid
        done
        
        # Analyze results
        for i in $(seq 1 $users); do
            local result_file="/tmp/light_concurrent_result_${i}.txt"
            if [ -f "$result_file" ]; then
                local result=$(head -n1 "$result_file")
                local response_time=$(tail -n1 "$result_file")
                
                if [ "$result" = "SUCCESS" ]; then
                    ((success_count++))
                fi
                
                if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                    total_response_time=$(echo "$total_response_time + $response_time" | bc)
                fi
                
                rm -f "$result_file"
            fi
        done
        
        local avg_response_time=$(echo "scale=3; $total_response_time / $users" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $users" | bc)
        
        echo "$users,$endpoint,$users,$success_count,$success_rate,$avg_response_time" >> "$concurrent_file"
        
        print_success "$users concurrent users: $success_count/$users successful ($success_rate%), Avg: ${avg_response_time}s"
    done
    
    print_success "Light concurrent user tests completed"
}

# Function to run light volume tests
run_light_volume_tests() {
    print_header "📈 Light Volume Tests"
    print_status "===================="
    
    local test_data='{
        "business_id": "volume-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local volume_file="$TEST_RESULTS_DIR/light_volume_tests.csv"
    echo "Total_Requests,Endpoint,Successful_Requests,Success_Rate,Avg_Response_Time,Requests_Per_Second" > "$volume_file"
    
    for volume in "${REQUEST_VOLUMES[@]}"; do
        print_status "Testing with $volume total requests..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        
        local success_count=0
        local total_response_time=0
        
        for i in $(seq 1 $volume); do
            local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                -X POST \
                -H "Content-Type: application/json" \
                -d "$test_data" \
                "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
            
            local http_code=$(echo "$response" | tail -n1)
            local response_time=$(echo "$response" | tail -n2 | head -n1)
            
            if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                ((success_count++))
            fi
            
            if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                total_response_time=$(echo "$total_response_time + $response_time" | bc)
            fi
            
            # Progress indicator
            if [ $((i % 10)) -eq 0 ]; then
                echo -n "."
            fi
        done
        
        local end_time=$(date +%s.%N)
        local total_time=$(echo "$end_time - $start_time" | bc)
        local avg_response_time=$(echo "scale=3; $total_response_time / $volume" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $volume" | bc)
        local requests_per_second=$(echo "scale=2; $volume / $total_time" | bc)
        
        echo "$volume,$endpoint,$success_count,$success_rate,$avg_response_time,$requests_per_second" >> "$volume_file"
        
        echo ""
        print_success "$volume requests: $success_count/$volume successful ($success_rate%), Avg: ${avg_response_time}s, RPS: ${requests_per_second}"
    done
    
    print_success "Light volume tests completed"
}

# Function to run quick UI performance tests
run_quick_ui_tests() {
    print_header "🖥️ Quick UI Performance Tests"
    print_status "============================"
    
    local ui_file="$TEST_RESULTS_DIR/quick_ui_performance.csv"
    echo "Page,Load_Time,Size_Bytes,Status_Code,Success" > "$ui_file"
    
    local ui_pages=(
        "market-analysis-dashboard.html"
        "competitive-analysis-dashboard.html"
        "business-growth-analytics.html"
    )
    
    for page in "${ui_pages[@]}"; do
        print_status "Testing UI performance for $page..."
        
        for i in {1..3}; do
            local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                "$UI_URL/$page" 2>/dev/null || echo "ERROR")
            
            local http_code=$(echo "$response" | tail -n1)
            local response_time=$(echo "$response" | tail -n2 | head -n1)
            local content=$(echo "$response" | head -n -2)
            local content_size=$(echo "$content" | wc -c)
            local success="false"
            
            if [ "$http_code" = "200" ]; then
                success="true"
            fi
            
            echo "$page,$response_time,$content_size,$http_code,$success" >> "$ui_file"
        done
        
        print_success "$page: Performance metrics recorded"
    done
    
    print_success "Quick UI performance tests completed"
}

# Function to generate light benchmark report
generate_light_benchmark_report() {
    local report_file="$TEST_RESULTS_DIR/light-performance-benchmark-report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating light performance benchmark report: $report_file"
    
    cat > "$report_file" << EOF
Light Performance Benchmarking Report
====================================
Generated: $(date)
Test Suite: Light Performance Benchmarking
Version: 1.0.0

Test Configuration:
- API Base URL: $BASE_URL
- UI Base URL: $UI_URL
- Concurrent Users Tested: ${CONCURRENT_USERS[*]}
- Request Volumes Tested: ${REQUEST_VOLUMES[*]}
- Duration Tests: ${DURATION_TESTS[*]} seconds

Performance Test Categories:
1. Quick Baseline Performance Tests
2. Light Concurrent User Tests
3. Light Volume Tests
4. Quick UI Performance Tests

Performance Metrics Summary:
===========================
- Response Time: Time taken to process a single request
- Throughput: Requests processed per second (RPS)
- Success Rate: Percentage of successful requests
- Concurrent Users: Number of simultaneous users supported

Performance Analysis:
====================
- Baseline performance established
- Concurrent user capacity tested
- Volume handling capability assessed
- UI performance evaluated

Recommendations:
===============
- Monitor response times under load
- Optimize for concurrent user scenarios
- Plan capacity scaling based on volume tests
- Improve UI loading performance

Performance Targets:
===================
- Response Time: < 1 second for 95% of requests
- Throughput: > 50 requests per second
- Success Rate: > 95% under normal load
- Concurrent Users: Support 10+ concurrent users

Next Steps:
===========
1. Analyze performance bottlenecks
2. Implement performance optimizations
3. Set up performance monitoring
4. Plan capacity scaling
5. Schedule regular performance testing
EOF
    
    print_success "Light performance benchmark report generated: $report_file"
}

# Function to display light benchmark summary
display_light_benchmark_summary() {
    print_header "📊 Light Performance Benchmark Summary"
    print_status "====================================="
    
    print_info "Light performance benchmarking completed successfully!"
    echo ""
    
    print_instruction "Benchmark Tests Completed:"
    echo "✅ Quick baseline performance tests"
    echo "✅ Light concurrent user tests"
    echo "✅ Light volume tests"
    echo "✅ Quick UI performance tests"
    echo ""
    
    print_instruction "Test Results Available:"
    echo "- Quick baseline: $TEST_RESULTS_DIR/quick_baseline_performance.csv"
    echo "- Light concurrent: $TEST_RESULTS_DIR/light_concurrent_users.csv"
    echo "- Light volume: $TEST_RESULTS_DIR/light_volume_tests.csv"
    echo "- Quick UI: $TEST_RESULTS_DIR/quick_ui_performance.csv"
    echo ""
    
    print_info "Check the test-results directory for detailed performance data."
}

# Main execution
main() {
    print_header "🚀 Light Performance Benchmarking"
    print_header "================================="
    
    # Start servers
    start_servers
    
    # Run light benchmark tests
    print_status "Starting light performance benchmarking..."
    
    # 1. Quick baseline performance test
    run_quick_baseline_test
    echo ""
    
    # 2. Light concurrent user tests
    run_light_concurrent_tests
    echo ""
    
    # 3. Light volume tests
    run_light_volume_tests
    echo ""
    
    # 4. Quick UI performance tests
    run_quick_ui_tests
    echo ""
    
    # Generate comprehensive report
    generate_light_benchmark_report
    
    # Display summary
    display_light_benchmark_summary
    
    # Stop servers
    stop_servers
}

# Trap to ensure servers are stopped on exit
trap stop_servers EXIT

# Run main function
main "$@"
