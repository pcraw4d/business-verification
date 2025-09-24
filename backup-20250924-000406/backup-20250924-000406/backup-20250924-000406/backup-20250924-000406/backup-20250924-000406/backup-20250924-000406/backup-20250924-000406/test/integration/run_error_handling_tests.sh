#!/bin/bash

# Error Handling Testing Suite Runner
# This script runs all error handling tests for the Supabase Table Improvement Implementation Plan

set -e

echo "🧪 Starting Error Handling Testing Suite"
echo "========================================"

# Set environment variable for integration tests
export INTEGRATION_TESTS=true

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test and track results
run_test() {
    local test_name="$1"
    local test_file="$2"
    
    echo -e "\n${BLUE}Running: $test_name${NC}"
    echo "File: $test_file"
    echo "----------------------------------------"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if go test -v "$test_file" -timeout 30s; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Function to run all tests in a file
run_test_suite() {
    local suite_name="$1"
    local test_file="$2"
    
    echo -e "\n${YELLOW}🧪 $suite_name${NC}"
    echo "========================================"
    
    run_test "$suite_name" "$test_file"
}

# Main test execution
echo -e "\n${YELLOW}Starting Error Handling Test Suite${NC}"

# Test 1: Enhanced Error Scenarios
run_test_suite "Enhanced Error Scenarios Testing" "./enhanced_error_handling_test.go"

# Test 2: Recovery Procedures
run_test_suite "Recovery Procedures Testing" "./recovery_procedures_test.go"

# Test 3: Logging and Monitoring
run_test_suite "Logging and Monitoring Testing" "./logging_monitoring_test.go"

# Test 4: User Feedback Systems
run_test_suite "User Feedback Systems Testing" "./user_feedback_test.go"

# Test 5: Original Error Handling Tests
run_test_suite "Original Error Handling Tests" "./error_handling_test.go"

# Summary
echo -e "\n${YELLOW}📊 Test Results Summary${NC}"
echo "========================================"
echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All Error Handling Tests Passed!${NC}"
    echo -e "${GREEN}✅ Subtask 4.3.3: Error Handling Testing - COMPLETED${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please review the output above.${NC}"
    exit 1
fi
