#!/bin/bash

# Railway Logs Fixes - Production Testing Script
# Tests Phase 1 (ML Timeout) and Phase 2 & 3 (Content Quality Thresholds)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Production URL (update if needed)
PROD_URL="${PROD_URL:-https://api-gateway-service-production-21fd.up.railway.app}"
CLASSIFICATION_URL="${CLASSIFICATION_URL:-${PROD_URL}/api/v1/classify}"

echo -e "${BLUE}🧪 Railway Logs Fixes - Production Testing${NC}"
echo "=========================================="
echo -e "Testing URL: ${YELLOW}${CLASSIFICATION_URL}${NC}"
echo ""

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
TIMEOUT_ERRORS=0
FAST_PATH_SUCCESS=0
CONTENT_QUALITY_PASS=0

# Test cases for different business types
declare -a TEST_CASES=(
    '{"business_name":"TechCorp Software","description":"Custom software development and cloud solutions","website_url":"https://techcorp.com"}'
    '{"business_name":"Digital Marketing Agency","description":"SEO, social media marketing, and content creation services","website_url":"https://digitalmarketing.com"}'
    '{"business_name":"Healthcare Clinic","description":"Primary care medical services and patient care","website_url":"https://healthcareclinic.com"}'
    '{"business_name":"E-commerce Retail","description":"Online retail store selling consumer products","website_url":"https://ecommerce.com"}'
    '{"business_name":"Consulting Firm","description":"Business strategy and management consulting","website_url":"https://consulting.com"}'
)

# Test Phase 1: ML Service Timeout Fix
echo -e "${BLUE}📊 Phase 1: Testing ML Service Timeout Fix${NC}"
echo "----------------------------------------"

test_ml_timeout() {
    local test_name="$1"
    local request_body="$2"
    local start_time=$(date +%s.%N)
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${YELLOW}Testing: ${test_name}${NC}"
    
    # Make request with timeout tracking
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" \
        -X POST "${CLASSIFICATION_URL}" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        --max-time 30 \
        -d "${request_body}" 2>&1)
    
    http_code=$(echo "$response" | tail -n 2 | head -n 1)
    time_total=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -2)
    
    # Check for timeout errors in response
    if echo "$body" | grep -qi "context deadline exceeded\|timeout\|deadline"; then
        echo -e "${RED}❌ FAIL: Timeout error detected${NC}"
        TIMEOUT_ERRORS=$((TIMEOUT_ERRORS + 1))
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Check HTTP status
    if [ "$http_code" != "200" ]; then
        echo -e "${RED}❌ FAIL: HTTP ${http_code}${NC}"
        echo "Response: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Check response time (should complete within reasonable time)
    time_seconds=$(echo "$time_total" | awk '{print $1}')
    if (( $(echo "$time_seconds > 10" | bc -l) )); then
        echo -e "${YELLOW}⚠️  WARN: Slow response time: ${time_seconds}s${NC}"
    else
        echo -e "${GREEN}✅ Response time: ${time_seconds}s${NC}"
    fi
    
    # Check if fast-path was used (look for indicators in response)
    if echo "$body" | grep -qi "lightweight\|fast\|quantization"; then
        FAST_PATH_SUCCESS=$((FAST_PATH_SUCCESS + 1))
        echo -e "${GREEN}✅ Fast-path mode detected${NC}"
    fi
    
    # Parse and display classification result
    if command -v jq &> /dev/null; then
        echo "Classification:"
        echo "$body" | jq -r '.classifications // .classification // "N/A"' 2>/dev/null || echo "Response received"
    fi
    
    PASSED_TESTS=$((PASSED_TESTS + 1))
    return 0
}

# Test Phase 2 & 3: Content Quality Thresholds
echo -e "\n${BLUE}📊 Phase 2 & 3: Testing Content Quality Thresholds${NC}"
echo "----------------------------------------"

test_content_quality() {
    local test_name="$1"
    local request_body="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${YELLOW}Testing: ${test_name}${NC}"
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST "${CLASSIFICATION_URL}" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        --max-time 30 \
        -d "${request_body}" 2>&1)
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" != "200" ]; then
        echo -e "${RED}❌ FAIL: HTTP ${http_code}${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Check if classification was successful (indicates content quality passed)
    if echo "$body" | grep -qi "classification\|industry\|naics\|sic\|mcc"; then
        CONTENT_QUALITY_PASS=$((CONTENT_QUALITY_PASS + 1))
        echo -e "${GREEN}✅ Content quality check passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${YELLOW}⚠️  WARN: No classification data in response${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Run Phase 1 tests
echo -e "${BLUE}Running Phase 1 tests (ML Timeout)...${NC}"
for i in "${!TEST_CASES[@]}"; do
    test_ml_timeout "Test Case $((i+1))" "${TEST_CASES[$i]}"
    sleep 1  # Small delay between requests
done

# Run Phase 2 & 3 tests
echo -e "\n${BLUE}Running Phase 2 & 3 tests (Content Quality)...${NC}"
for i in "${!TEST_CASES[@]}"; do
    test_content_quality "Test Case $((i+1))" "${TEST_CASES[$i]}"
    sleep 1  # Small delay between requests
done

# Summary
echo -e "\n${BLUE}=========================================="
echo -e "📊 Test Summary${NC}"
echo "=========================================="
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
echo ""
echo -e "${BLUE}Phase 1 Metrics:${NC}"
echo -e "  Timeout Errors: ${TIMEOUT_ERRORS}"
echo -e "  Fast-Path Success: ${FAST_PATH_SUCCESS}/${TOTAL_TESTS}"
echo ""
echo -e "${BLUE}Phase 2 & 3 Metrics:${NC}"
echo -e "  Content Quality Pass: ${CONTENT_QUALITY_PASS}/${TOTAL_TESTS}"
echo ""

# Success criteria check
SUCCESS=true

if [ $TIMEOUT_ERRORS -gt 0 ]; then
    echo -e "${RED}❌ Phase 1 FAIL: Timeout errors detected${NC}"
    SUCCESS=false
else
    echo -e "${GREEN}✅ Phase 1 PASS: No timeout errors${NC}"
fi

if [ $FAST_PATH_SUCCESS -lt $((TOTAL_TESTS / 2)) ]; then
    echo -e "${YELLOW}⚠️  Phase 1 WARN: Low fast-path success rate${NC}"
else
    echo -e "${GREEN}✅ Phase 1 PASS: Fast-path working${NC}"
fi

if [ $CONTENT_QUALITY_PASS -lt $((TOTAL_TESTS / 2)) ]; then
    echo -e "${YELLOW}⚠️  Phase 2 & 3 WARN: Low content quality pass rate${NC}"
else
    echo -e "${GREEN}✅ Phase 2 & 3 PASS: Content quality checks passing${NC}"
fi

if [ "$SUCCESS" = true ] && [ $PASSED_TESTS -ge $((TOTAL_TESTS * 8 / 10)) ]; then
    echo -e "\n${GREEN}✅ Overall: Tests PASSED${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  Overall: Some tests need attention${NC}"
    exit 1
fi

