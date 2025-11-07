#!/bin/bash

# Master Test Runner
# Runs all testing and analysis scripts

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo -e "${CYAN}🧪 KYB Platform - Comprehensive Testing Suite${NC}\n"
echo "This script will run:"
echo "  1. API Endpoint Testing"
echo "  2. Placeholder Data Detection"
echo "  3. Unused Features Analysis"
echo ""

# Create test results directory
mkdir -p test-results

# Track results
declare -a FAILED_TESTS
TOTAL_TESTS=0
PASSED_TESTS=0

# Test 1: API Testing
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 1: Comprehensive API Testing${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
if bash scripts/comprehensive-api-test.sh; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}✅ API Testing: PASSED${NC}\n"
else
    FAILED_TESTS+=("API Testing")
    echo -e "${RED}❌ API Testing: FAILED${NC}\n"
fi

# Test 2: Placeholder Detection
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 2: Placeholder Data Detection${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
if bash scripts/detect-placeholder-data.sh; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}✅ Placeholder Detection: PASSED${NC}\n"
else
    # This test exits with 1 if issues found, but that's expected
    echo -e "${YELLOW}⚠️  Placeholder Detection: Issues found (review above)${NC}\n"
fi

# Test 3: Unused Features Analysis
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 3: Unused Features Analysis${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
if node scripts/analyze-unused-features.js; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}✅ Unused Features Analysis: PASSED${NC}\n"
else
    # This test exits with 1 if unused features found, but that's expected
    echo -e "${YELLOW}⚠️  Unused Features Analysis: Unused features found (review above)${NC}\n"
fi

# Summary
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📊 Test Summary${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo -e "${RED}Failed: ${#FAILED_TESTS[@]}${NC}"
    echo "Failed tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
else
    echo -e "${GREEN}Failed: 0${NC}"
fi

echo ""
echo "Test results saved in: test-results/"
echo ""

# List generated reports
echo -e "${BLUE}Generated Reports:${NC}"
ls -lh test-results/ 2>/dev/null | tail -n +2 | awk '{print "  - " $9 " (" $5 ")"}' || echo "  No reports generated"

echo ""

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All critical tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please review the results above.${NC}"
    exit 1
fi

