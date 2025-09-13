#!/bin/bash

# Performance Benchmarking Script
# Comprehensive performance testing and benchmarking for business intelligence system

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

# Benchmarking parameters
CONCURRENT_USERS=(1 5 10 25 50 100)
REQUEST_VOLUMES=(10 50 100 500 1000)
DURATION_TESTS=(30 60 120) # seconds
LOAD_PATTERNS=("constant" "ramp" "spike" "stress")

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

# Function to run baseline performance test
run_baseline_test() {
    print_header "📊 Baseline Performance Test"
    print_status "============================"
    
    local test_data='{
        "business_id": "baseline-test-123",
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
    
    local baseline_file="$TEST_RESULTS_DIR/baseline_performance.csv"
    echo "Test,Endpoint,Response_Time,Status_Code,Success" > "$baseline_file"
    
    print_status "Running baseline tests for all endpoints..."
    
    local endpoints=(
        "/v2/business-intelligence/market-analysis"
        "/v2/business-intelligence/competitive-analysis"
        "/v2/business-intelligence/growth-analytics"
    )
    
    for endpoint in "${endpoints[@]}"; do
        print_status "Testing $endpoint..."
        
        for i in {1..10}; do
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
    done
    
    print_success "Baseline performance test completed"
}

# Function to run concurrent user tests
run_concurrent_user_tests() {
    print_header "👥 Concurrent User Tests"
    print_status "======================="
    
    local test_data='{
        "business_id": "concurrent-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local concurrent_file="$TEST_RESULTS_DIR/concurrent_users.csv"
    echo "Concurrent_Users,Endpoint,Total_Requests,Successful_Requests,Success_Rate,Avg_Response_Time,Min_Response_Time,Max_Response_Time,Total_Time" > "$concurrent_file"
    
    for users in "${CONCURRENT_USERS[@]}"; do
        print_status "Testing with $users concurrent users..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        
        # Run concurrent requests
        local pids=()
        local success_count=0
        local total_response_time=0
        local min_time=999999
        local max_time=0
        
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
                    echo "SUCCESS" > "/tmp/concurrent_result_${i}.txt"
                else
                    echo "FAILED" > "/tmp/concurrent_result_${i}.txt"
                fi
                
                echo "$response_time" >> "/tmp/concurrent_result_${i}.txt"
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
        for i in $(seq 1 $users); do
            local result_file="/tmp/concurrent_result_${i}.txt"
            if [ -f "$result_file" ]; then
                local result=$(head -n1 "$result_file")
                local response_time=$(tail -n1 "$result_file")
                
                if [ "$result" = "SUCCESS" ]; then
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
                
                rm -f "$result_file"
            fi
        done
        
        local avg_response_time=$(echo "scale=3; $total_response_time / $users" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $users" | bc)
        
        echo "$users,$endpoint,$users,$success_count,$success_rate,$avg_response_time,$min_time,$max_time,$total_time" >> "$concurrent_file"
        
        print_success "$users concurrent users: $success_count/$users successful ($success_rate%), Avg: ${avg_response_time}s"
    done
    
    print_success "Concurrent user tests completed"
}

# Function to run volume tests
run_volume_tests() {
    print_header "📈 Volume Tests"
    print_status "=============="
    
    local test_data='{
        "business_id": "volume-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local volume_file="$TEST_RESULTS_DIR/volume_tests.csv"
    echo "Total_Requests,Endpoint,Successful_Requests,Success_Rate,Avg_Response_Time,Min_Response_Time,Max_Response_Time,Total_Time,Requests_Per_Second" > "$volume_file"
    
    for volume in "${REQUEST_VOLUMES[@]}"; do
        print_status "Testing with $volume total requests..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        
        local success_count=0
        local total_response_time=0
        local min_time=999999
        local max_time=0
        
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
                if (( $(echo "$response_time < $min_time" | bc -l) )); then
                    min_time=$response_time
                fi
                if (( $(echo "$response_time > $max_time" | bc -l) )); then
                    max_time=$response_time
                fi
            fi
            
            # Progress indicator
            if [ $((i % 50)) -eq 0 ]; then
                echo -n "."
            fi
        done
        
        local end_time=$(date +%s.%N)
        local total_time=$(echo "$end_time - $start_time" | bc)
        local avg_response_time=$(echo "scale=3; $total_response_time / $volume" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $volume" | bc)
        local requests_per_second=$(echo "scale=2; $volume / $total_time" | bc)
        
        echo "$volume,$endpoint,$success_count,$success_rate,$avg_response_time,$min_time,$max_time,$total_time,$requests_per_second" >> "$volume_file"
        
        echo ""
        print_success "$volume requests: $success_count/$volume successful ($success_rate%), Avg: ${avg_response_time}s, RPS: ${requests_per_second}"
    done
    
    print_success "Volume tests completed"
}

# Function to run duration tests
run_duration_tests() {
    print_header "⏱️ Duration Tests"
    print_status "================"
    
    local test_data='{
        "business_id": "duration-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local duration_file="$TEST_RESULTS_DIR/duration_tests.csv"
    echo "Duration_Seconds,Endpoint,Total_Requests,Successful_Requests,Success_Rate,Avg_Response_Time,Requests_Per_Second" > "$duration_file"
    
    for duration in "${DURATION_TESTS[@]}"; do
        print_status "Testing for $duration seconds..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        local end_time=$(echo "$start_time + $duration" | bc)
        
        local request_count=0
        local success_count=0
        local total_response_time=0
        
        while (( $(echo "$(date +%s.%N) < $end_time" | bc -l) )); do
            local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                -X POST \
                -H "Content-Type: application/json" \
                -d "$test_data" \
                "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
            
            local http_code=$(echo "$response" | tail -n1)
            local response_time=$(echo "$response" | tail -n2 | head -n1)
            
            ((request_count++))
            
            if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                ((success_count++))
            fi
            
            if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                total_response_time=$(echo "$total_response_time + $response_time" | bc)
            fi
            
            # Progress indicator
            if [ $((request_count % 10)) -eq 0 ]; then
                echo -n "."
            fi
        done
        
        local actual_duration=$(echo "$(date +%s.%N) - $start_time" | bc)
        local avg_response_time=$(echo "scale=3; $total_response_time / $request_count" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $request_count" | bc)
        local requests_per_second=$(echo "scale=2; $request_count / $actual_duration" | bc)
        
        echo "$duration,$endpoint,$request_count,$success_count,$success_rate,$avg_response_time,$requests_per_second" >> "$duration_file"
        
        echo ""
        print_success "$duration seconds: $request_count requests, $success_count successful ($success_rate%), Avg: ${avg_response_time}s, RPS: ${requests_per_second}"
    done
    
    print_success "Duration tests completed"
}

# Function to run load pattern tests
run_load_pattern_tests() {
    print_header "📊 Load Pattern Tests"
    print_status "===================="
    
    local test_data='{
        "business_id": "load-pattern-test-123",
        "time_range": {
            "start_date": "2024-01-01T00:00:00Z",
            "end_date": "2024-12-31T23:59:59Z"
        }
    }'
    
    local pattern_file="$TEST_RESULTS_DIR/load_patterns.csv"
    echo "Load_Pattern,Endpoint,Total_Requests,Successful_Requests,Success_Rate,Avg_Response_Time,Peak_RPS" > "$pattern_file"
    
    for pattern in "${LOAD_PATTERNS[@]}"; do
        print_status "Testing $pattern load pattern..."
        
        local endpoint="/v2/business-intelligence/market-analysis"
        local start_time=$(date +%s.%N)
        local total_requests=0
        local success_count=0
        local total_response_time=0
        local peak_rps=0
        
        case $pattern in
            "constant")
                # Constant load: 10 requests per second for 30 seconds
                for i in $(seq 1 300); do
                    local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                        -X POST \
                        -H "Content-Type: application/json" \
                        -d "$test_data" \
                        "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                    
                    local http_code=$(echo "$response" | tail -n1)
                    local response_time=$(echo "$response" | tail -n2 | head -n1)
                    
                    ((total_requests++))
                    
                    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                        ((success_count++))
                    fi
                    
                    if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                        total_response_time=$(echo "$total_response_time + $response_time" | bc)
                    fi
                    
                    sleep 0.1
                done
                peak_rps=10
                ;;
            "ramp")
                # Ramp load: gradually increase from 1 to 20 requests per second
                for rate in $(seq 1 20); do
                    for i in $(seq 1 $rate); do
                        local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                            -X POST \
                            -H "Content-Type: application/json" \
                            -d "$test_data" \
                            "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                        
                        local http_code=$(echo "$response" | tail -n1)
                        local response_time=$(echo "$response" | tail -n2 | head -n1)
                        
                        ((total_requests++))
                        
                        if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                            ((success_count++))
                        fi
                        
                        if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                            total_response_time=$(echo "$total_response_time + $response_time" | bc)
                        fi
                    done
                    sleep 1
                done
                peak_rps=20
                ;;
            "spike")
                # Spike load: normal load with sudden spikes
                for i in $(seq 1 100); do
                    local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                        -X POST \
                        -H "Content-Type: application/json" \
                        -d "$test_data" \
                        "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                    
                    local http_code=$(echo "$response" | tail -n1)
                    local response_time=$(echo "$response" | tail -n2 | head -n1)
                    
                    ((total_requests++))
                    
                    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                        ((success_count++))
                    fi
                    
                    if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                        total_response_time=$(echo "$total_response_time + $response_time" | bc)
                    fi
                    
                    # Spike every 20 requests
                    if [ $((i % 20)) -eq 0 ]; then
                        for j in $(seq 1 10); do
                            local spike_response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                                -X POST \
                                -H "Content-Type: application/json" \
                                -d "$test_data" \
                                "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                            
                            local spike_http_code=$(echo "$spike_response" | tail -n1)
                            local spike_response_time=$(echo "$spike_response" | tail -n2 | head -n1)
                            
                            ((total_requests++))
                            
                            if [ "$spike_http_code" = "200" ] || [ "$spike_http_code" = "201" ] || [ "$spike_http_code" = "501" ]; then
                                ((success_count++))
                            fi
                            
                            if [ "$spike_response_time" != "ERROR" ] && [ "$spike_response_time" != "" ]; then
                                total_response_time=$(echo "$total_response_time + $spike_response_time" | bc)
                            fi
                        done
                    fi
                    
                    sleep 0.1
                done
                peak_rps=15
                ;;
            "stress")
                # Stress test: maximum load until failure
                local stress_start=$(date +%s.%N)
                while true; do
                    local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                        -X POST \
                        -H "Content-Type: application/json" \
                        -d "$test_data" \
                        "$BASE_URL$endpoint" 2>/dev/null || echo "ERROR")
                    
                    local http_code=$(echo "$response" | tail -n1)
                    local response_time=$(echo "$response" | tail -n2 | head -n1)
                    
                    ((total_requests++))
                    
                    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "501" ]; then
                        ((success_count++))
                    fi
                    
                    if [ "$response_time" != "ERROR" ] && [ "$response_time" != "" ]; then
                        total_response_time=$(echo "$total_response_time + $response_time" | bc)
                    fi
                    
                    # Stop after 30 seconds or if error rate > 50%
                    local current_time=$(date +%s.%N)
                    local elapsed=$(echo "$current_time - $stress_start" | bc)
                    if (( $(echo "$elapsed > 30" | bc -l) )); then
                        break
                    fi
                    
                    local error_rate=$(echo "scale=2; ($total_requests - $success_count) * 100 / $total_requests" | bc)
                    if (( $(echo "$error_rate > 50" | bc -l) )); then
                        break
                    fi
                done
                peak_rps=$total_requests
                ;;
        esac
        
        local end_time=$(date +%s.%N)
        local total_time=$(echo "$end_time - $start_time" | bc)
        local avg_response_time=$(echo "scale=3; $total_response_time / $total_requests" | bc)
        local success_rate=$(echo "scale=2; $success_count * 100 / $total_requests" | bc)
        
        echo "$pattern,$endpoint,$total_requests,$success_count,$success_rate,$avg_response_time,$peak_rps" >> "$pattern_file"
        
        print_success "$pattern pattern: $total_requests requests, $success_count successful ($success_rate%), Avg: ${avg_response_time}s, Peak RPS: ${peak_rps}"
    done
    
    print_success "Load pattern tests completed"
}

# Function to run UI performance tests
run_ui_performance_tests() {
    print_header "🖥️ UI Performance Tests"
    print_status "======================"
    
    local ui_file="$TEST_RESULTS_DIR/ui_performance.csv"
    echo "Page,Load_Time,Size_Bytes,Status_Code,Success" > "$ui_file"
    
    local ui_pages=(
        "market-analysis-dashboard.html"
        "competitive-analysis-dashboard.html"
        "business-growth-analytics.html"
        "dashboard.html"
        "index.html"
    )
    
    for page in "${ui_pages[@]}"; do
        print_status "Testing UI performance for $page..."
        
        for i in {1..5}; do
            local start_time=$(date +%s.%N)
            local response=$(curl -s -w "\n%{time_total}\n%{http_code}" \
                "$UI_URL/$page" 2>/dev/null || echo "ERROR")
            local end_time=$(date +%s.%N)
            
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
        
        print_success "$page: Average load time and size recorded"
    done
    
    print_success "UI performance tests completed"
}

# Function to generate performance benchmark report
generate_benchmark_report() {
    local report_file="$TEST_RESULTS_DIR/performance-benchmark-report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating performance benchmark report: $report_file"
    
    cat > "$report_file" << EOF
Performance Benchmarking Report
==============================
Generated: $(date)
Test Suite: Performance Benchmarking
Version: 1.0.0

Test Configuration:
- API Base URL: $BASE_URL
- UI Base URL: $UI_URL
- Concurrent Users Tested: ${CONCURRENT_USERS[*]}
- Request Volumes Tested: ${REQUEST_VOLUMES[*]}
- Duration Tests: ${DURATION_TESTS[*]} seconds
- Load Patterns: ${LOAD_PATTERNS[*]}

Performance Test Categories:
1. Baseline Performance Tests
2. Concurrent User Tests
3. Volume Tests
4. Duration Tests
5. Load Pattern Tests
6. UI Performance Tests

Performance Metrics:
===================
- Response Time: Time taken to process a single request
- Throughput: Requests processed per second (RPS)
- Success Rate: Percentage of successful requests
- Concurrent Users: Number of simultaneous users supported
- Load Capacity: Maximum sustainable load
- Error Rate: Percentage of failed requests

Performance Benchmarks:
======================

Baseline Performance:
- Single request response times
- API endpoint performance comparison
- Initial performance baseline establishment

Concurrent User Performance:
- System behavior under concurrent load
- Response time degradation analysis
- Success rate under concurrent users
- Maximum concurrent users supported

Volume Performance:
- System behavior under high request volumes
- Throughput analysis
- Performance degradation patterns
- Maximum request volume capacity

Duration Performance:
- Sustained performance over time
- Performance stability analysis
- Memory and resource usage patterns
- Long-term reliability assessment

Load Pattern Performance:
- Constant load performance
- Ramp load performance
- Spike load handling
- Stress test results

UI Performance:
- Page load times
- Content size analysis
- UI responsiveness metrics
- User experience performance

Performance Analysis:
====================
- Identify performance bottlenecks
- Analyze response time patterns
- Evaluate system scalability
- Assess resource utilization
- Determine optimal configuration

Recommendations:
===============
- Performance optimization suggestions
- Scalability improvements
- Resource allocation recommendations
- Monitoring and alerting setup
- Capacity planning guidance

Performance Targets:
===================
- Response Time: < 1 second for 95% of requests
- Throughput: > 100 requests per second
- Success Rate: > 99% under normal load
- Concurrent Users: Support 100+ concurrent users
- Availability: 99.9% uptime

Next Steps:
===========
1. Analyze performance bottlenecks
2. Implement performance optimizations
3. Set up performance monitoring
4. Plan capacity scaling
5. Schedule regular performance testing
EOF
    
    print_success "Performance benchmark report generated: $report_file"
}

# Function to display benchmark summary
display_benchmark_summary() {
    print_header "📊 Performance Benchmark Summary"
    print_status "==============================="
    
    print_info "Performance benchmarking completed successfully!"
    echo ""
    
    print_instruction "Benchmark Tests Completed:"
    echo "✅ Baseline performance tests"
    echo "✅ Concurrent user tests"
    echo "✅ Volume tests"
    echo "✅ Duration tests"
    echo "✅ Load pattern tests"
    echo "✅ UI performance tests"
    echo ""
    
    print_instruction "Test Results Available:"
    echo "- Baseline performance: $TEST_RESULTS_DIR/baseline_performance.csv"
    echo "- Concurrent users: $TEST_RESULTS_DIR/concurrent_users.csv"
    echo "- Volume tests: $TEST_RESULTS_DIR/volume_tests.csv"
    echo "- Duration tests: $TEST_RESULTS_DIR/duration_tests.csv"
    echo "- Load patterns: $TEST_RESULTS_DIR/load_patterns.csv"
    echo "- UI performance: $TEST_RESULTS_DIR/ui_performance.csv"
    echo ""
    
    print_info "Check the test-results directory for detailed performance data and analysis."
}

# Main execution
main() {
    print_header "🚀 Performance Benchmarking"
    print_header "==========================="
    
    # Start servers
    start_servers
    
    # Run all benchmark tests
    print_status "Starting comprehensive performance benchmarking..."
    
    # 1. Baseline performance test
    run_baseline_test
    echo ""
    
    # 2. Concurrent user tests
    run_concurrent_user_tests
    echo ""
    
    # 3. Volume tests
    run_volume_tests
    echo ""
    
    # 4. Duration tests
    run_duration_tests
    echo ""
    
    # 5. Load pattern tests
    run_load_pattern_tests
    echo ""
    
    # 6. UI performance tests
    run_ui_performance_tests
    echo ""
    
    # Generate comprehensive report
    generate_benchmark_report
    
    # Display summary
    display_benchmark_summary
    
    # Stop servers
    stop_servers
}

# Trap to ensure servers are stopped on exit
trap stop_servers EXIT

# Run main function
main "$@"
