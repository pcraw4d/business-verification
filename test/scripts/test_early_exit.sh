#!/bin/bash

# Test Early Exit Functionality
# Tests if early exit is working correctly after fixes

set -e

CLASSIFICATION_API_URL="https://classification-service-production.up.railway.app"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Early Exit Functionality Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${CYAN}$CLASSIFICATION_API_URL${NC}"
echo ""

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test function
test_early_exit() {
    local test_name="$1"
    local request_data="$2"
    local expected_early_exit="$3"
    local expected_confidence_min="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test: $test_name${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    echo -e "Request Data:"
    echo "$request_data" | python3 -m json.tool 2>/dev/null || echo "$request_data"
    echo ""
    
    START_TIME=$(date +%s.%N)
    HTTP_CODE=$(curl -s -o /tmp/response.json -w "%{http_code}" -X POST "$CLASSIFICATION_API_URL/v1/classify" \
      -H "Content-Type: application/json" \
      -d "$request_data" \
      --max-time 60 2>&1)
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    
    RESPONSE_BODY=$(cat /tmp/response.json 2>/dev/null || echo "{}")
    
    echo -e "Response Time: ${CYAN}${DURATION}s${NC}"
    echo ""
    
    # Parse response (handle both True/true and False/false)
    SUCCESS=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); s=d.get('success', False); print('True' if s else 'False')" 2>/dev/null || echo "False")
    EARLY_EXIT=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); ee=d.get('metadata', {}).get('early_exit', False); print('True' if ee else 'False')" 2>/dev/null || echo "False")
    SCRAPING_STRATEGY=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('metadata', {}).get('scraping_strategy', '') or '')" 2>/dev/null || echo "")
    PROCESSING_PATH=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('processing_path', '') or '')" 2>/dev/null || echo "")
    CONFIDENCE=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('confidence_score', 0))" 2>/dev/null || echo "0")
    INDUSTRY=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('primary_industry', 'N/A'))" 2>/dev/null || echo "N/A")
    
    echo -e "Success: ${CYAN}$SUCCESS${NC}"
    echo -e "Early Exit: ${CYAN}$EARLY_EXIT${NC}"
    echo -e "Scraping Strategy: ${CYAN}$SCRAPING_STRATEGY${NC}"
    echo -e "Processing Path: ${CYAN}$PROCESSING_PATH${NC}"
    echo -e "Confidence Score: ${CYAN}$CONFIDENCE${NC}"
    echo -e "Primary Industry: ${CYAN}$INDUSTRY${NC}"
    echo ""
    
    # Validate results
    TEST_PASSED=true
    FAILURE_REASONS=()
    
    if [ "$SUCCESS" != "True" ]; then
        TEST_PASSED=false
        FAILURE_REASONS+=("Request failed (success=$SUCCESS)")
        # If request failed, skip other validations
        echo ""
        if [ "$TEST_PASSED" = "true" ]; then
            echo -e "${GREEN}✅ TEST PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ TEST FAILED${NC}"
            for reason in "${FAILURE_REASONS[@]}"; do
                echo -e "  ${RED}- $reason${NC}"
            done
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        echo ""
        return
    fi
    
    if [ "$expected_early_exit" = "true" ]; then
        if [ "$EARLY_EXIT" != "True" ]; then
            TEST_PASSED=false
            FAILURE_REASONS+=("Early exit not set (expected: True, got: $EARLY_EXIT) - Fixes may not be deployed")
        fi
        
        if [ "$SCRAPING_STRATEGY" != "early_exit" ]; then
            TEST_PASSED=false
            FAILURE_REASONS+=("Scraping strategy not 'early_exit' (got: $SCRAPING_STRATEGY)")
        fi
        
        if [ "$PROCESSING_PATH" != "layer1" ]; then
            TEST_PASSED=false
            FAILURE_REASONS+=("Processing path not 'layer1' (got: $PROCESSING_PATH)")
        fi
    fi
    
    if [ -n "$expected_confidence_min" ]; then
        CONFIDENCE_FLOAT=$(echo "$CONFIDENCE" | python3 -c "import sys; print(float(sys.stdin.read()))" 2>/dev/null || echo "0")
        MIN_FLOAT=$(echo "$expected_confidence_min" | python3 -c "import sys; print(float(sys.stdin.read()))" 2>/dev/null || echo "0")
        if (( $(echo "$CONFIDENCE_FLOAT < $MIN_FLOAT" | bc -l) )); then
            TEST_PASSED=false
            FAILURE_REASONS+=("Confidence too low (expected: >= $expected_confidence_min, got: $CONFIDENCE)")
        fi
    fi
    
    # Report results
    if [ "$TEST_PASSED" = "true" ]; then
        echo -e "${GREEN}✅ TEST PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ TEST FAILED${NC}"
        for reason in "${FAILURE_REASONS[@]}"; do
            echo -e "  ${RED}- $reason${NC}"
        done
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo ""
}

# Test 1: High Confidence Request (should trigger early exit)
test_early_exit \
    "High Confidence - Software Development" \
    '{"business_name": "Microsoft Corporation", "description": "Software development, cloud computing, and technology services"}' \
    "true" \
    "0.85"

# Test 2: High Confidence - Technology Keywords
test_early_exit \
    "High Confidence - Technology Keywords" \
    '{"business_name": "Tech Startup Inc", "description": "Software development and cloud consulting services"}' \
    "true" \
    "0.85"

# Test 3: Healthcare (should have high confidence)
test_early_exit \
    "High Confidence - Healthcare" \
    '{"business_name": "City Hospital", "description": "Medical services, patient care, and healthcare"}' \
    "true" \
    "0.80"

# Test 4: Financial Services (should have high confidence)
test_early_exit \
    "High Confidence - Financial Services" \
    '{"business_name": "Bank of America", "description": "Banking services, loans, and financial products"}' \
    "true" \
    "0.80"

# Test 5: Retail (should have high confidence)
test_early_exit \
    "High Confidence - Retail" \
    '{"business_name": "Walmart Store", "description": "Retail store selling groceries and general merchandise"}' \
    "true" \
    "0.80"

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Total Tests: ${CYAN}$TOTAL_TESTS${NC}"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
    echo -e "Early exit functionality is working correctly!"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    echo -e "Early exit functionality needs fixes or deployment."
    exit 1
fi

