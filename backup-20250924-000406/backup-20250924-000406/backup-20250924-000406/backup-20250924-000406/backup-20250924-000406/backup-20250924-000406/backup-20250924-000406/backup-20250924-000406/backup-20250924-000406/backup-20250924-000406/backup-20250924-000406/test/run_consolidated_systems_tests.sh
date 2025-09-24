#!/bin/bash

# =====================================================
# Consolidated Systems Test Runner
# Task 2.3.3: Test Consolidated Systems
# =====================================================
# This script runs comprehensive tests for the consolidated
# audit and compliance systems to ensure they work correctly
# and maintain data integrity while improving performance.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT="10m"
VERBOSE_FLAG="-v"
COVERAGE_FLAG="-cover"

echo -e "${BLUE}=====================================================${NC}"
echo -e "${BLUE}  Consolidated Systems Test Suite${NC}"
echo -e "${BLUE}  Task 2.3.3: Test Consolidated Systems${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo ""

# Function to run tests with proper error handling
run_test_suite() {
    local test_name="$1"
    local test_path="$2"
    local description="$3"
    
    echo -e "${YELLOW}Running $test_name...${NC}"
    echo -e "${BLUE}Description: $description${NC}"
    echo ""
    
    if go test -timeout $TEST_TIMEOUT $VERBOSE_FLAG $COVERAGE_FLAG "$test_path"; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        echo ""
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    local test_name="$1"
    local test_path="$2"
    local description="$3"
    
    echo -e "${YELLOW}Running $test_name...${NC}"
    echo -e "${BLUE}Description: $description${NC}"
    echo ""
    
    if go test -timeout $TEST_TIMEOUT $VERBOSE_FLAG -run "TestConsolidatedSystemsPerformance" "$test_path"; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        echo ""
        return 1
    fi
}

# Track test results
total_tests=0
passed_tests=0
failed_tests=0

# Test 1: Integration Tests
echo -e "${BLUE}1. Testing Integration${NC}"
echo "====================================================="
if run_test_suite "Integration Tests" "./test/integration" "Integration tests for consolidated systems including validation and data integrity"; then
    ((passed_tests++))
else
    ((failed_tests++))
fi
((total_tests++))

# Test 2: Performance Tests
echo -e "${BLUE}2. Testing Performance${NC}"
echo "====================================================="
if run_performance_tests "Performance Tests" "./test/performance" "Performance benchmarks for consolidated systems including creation, validation, and concurrent operations"; then
    ((passed_tests++))
else
    ((failed_tests++))
fi
((total_tests++))

# Test 3: Existing Compliance Tests (to ensure no regression)
echo -e "${BLUE}3. Testing Existing Compliance System${NC}"
echo "====================================================="
if run_test_suite "Existing Compliance Tests" "./test/compliance" "Regression tests for existing compliance functionality to ensure no breaking changes"; then
    ((passed_tests++))
else
    ((failed_tests++))
fi
((total_tests++))

# Generate test summary
echo -e "${BLUE}=====================================================${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}=====================================================${NC}"
echo ""

if [ $failed_tests -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
    echo -e "${GREEN}Total Tests: $total_tests${NC}"
    echo -e "${GREEN}Passed: $passed_tests${NC}"
    echo -e "${GREEN}Failed: $failed_tests${NC}"
    echo ""
    echo -e "${GREEN}✅ Consolidated Systems are working correctly!${NC}"
    echo -e "${GREEN}✅ Data integrity is maintained!${NC}"
    echo -e "${GREEN}✅ Performance meets requirements!${NC}"
    echo -e "${GREEN}✅ Integration is successful!${NC}"
    echo ""
    echo -e "${GREEN}Subtask 2.3.3: Test Consolidated Systems - COMPLETED${NC}"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    echo -e "${RED}Total Tests: $total_tests${NC}"
    echo -e "${GREEN}Passed: $passed_tests${NC}"
    echo -e "${RED}Failed: $failed_tests${NC}"
    echo ""
    echo -e "${RED}❌ Consolidated Systems need attention!${NC}"
    echo -e "${RED}❌ Please review failed tests and fix issues!${NC}"
    echo ""
    echo -e "${RED}Subtask 2.3.3: Test Consolidated Systems - FAILED${NC}"
    exit 1
fi