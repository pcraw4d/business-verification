#!/bin/bash

# Verify Phase 2 CTA Buttons Implementation
# This script verifies that all CTA buttons are properly implemented
# Usage: ./scripts/verify-phase2-ctas.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Verifying Phase 2 CTA Buttons Implementation...${NC}"
echo ""

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test 1: Verify Retry buttons are wired to fetch functions
test_retry_buttons() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 1: Retry Buttons Implementation${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check if Retry buttons call fetch functions
    RETRY_BUTTONS=$(grep -r "onClick.*fetch" frontend/components/merchant/*.tsx 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${RETRY_BUTTONS}" -gt 0 ]; then
        echo -e "${GREEN}✅ Found ${RETRY_BUTTONS} Retry/Refresh buttons with onClick handlers${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ No Retry buttons with onClick handlers found${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
}

# Test 2: Verify error messages include error codes
test_error_codes_in_messages() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 2: Error Codes in Error Messages${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check if formatErrorWithCode is used in error displays
    FORMAT_ERROR_USAGE=$(grep -r "formatErrorWithCode" frontend/components/merchant/*.tsx 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${FORMAT_ERROR_USAGE}" -gt 0 ]; then
        echo -e "${GREEN}✅ Found ${FORMAT_ERROR_USAGE} uses of formatErrorWithCode${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ formatErrorWithCode not used in components${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    # Check for error code patterns
    ERROR_CODE_PATTERNS=$(grep -r "ErrorCodes\." frontend/components/merchant/*.tsx 2>/dev/null | wc -l | tr -d ' ')
    echo -e "${GREEN}✅ Found ${ERROR_CODE_PATTERNS} error code references${NC}"
    echo ""
}

# Test 3: Verify all components have CTAs in error states
test_ctas_in_error_states() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 3: CTAs in Error States${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    COMPONENTS=(
        "PortfolioComparisonCard"
        "RiskScoreCard"
        "AnalyticsComparison"
        "RiskBenchmarkComparison"
        "RiskExplainabilitySection"
    )
    
    for component in "${COMPONENTS[@]}"; do
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        FILE="frontend/components/merchant/${component}.tsx"
        
        if [ -f "${FILE}" ]; then
            # Check if component has error state with button
            HAS_ERROR_BUTTON=$(grep -A 20 "if (error)" "${FILE}" 2>/dev/null | grep -i "Button\|onClick" | wc -l | tr -d ' ')
            
            if [ "${HAS_ERROR_BUTTON}" -gt 0 ]; then
                echo -e "${GREEN}✅ ${component}: Has CTA in error state${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo -e "${YELLOW}⚠️  ${component}: May not have CTA in error state${NC}"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            echo -e "${YELLOW}⚠️  ${component}: File not found${NC}"
        fi
    done
    echo ""
}

# Test 4: Verify button text matches expected patterns
test_button_text() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 4: Button Text Verification${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    EXPECTED_BUTTONS=(
        "Retry"
        "Refresh Data"
        "Run Risk Assessment"
        "Start Risk Assessment"
        "Enrich Data"
    )
    
    FOUND_BUTTONS=0
    for button_text in "${EXPECTED_BUTTONS[@]}"; do
        if grep -r "\"${button_text}\"" frontend/components/merchant/*.tsx 2>/dev/null | head -1 > /dev/null; then
            FOUND_BUTTONS=$((FOUND_BUTTONS + 1))
            echo -e "${GREEN}✅ Found: \"${button_text}\"${NC}"
        fi
    done
    
    if [ "${FOUND_BUTTONS}" -eq "${#EXPECTED_BUTTONS[@]}" ]; then
        echo -e "${GREEN}✅ All expected button texts found (${FOUND_BUTTONS}/${#EXPECTED_BUTTONS[@]})${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${YELLOW}⚠️  Found ${FOUND_BUTTONS}/${#EXPECTED_BUTTONS[@]} expected button texts${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
}

# Test 5: Verify error recovery functions exist
test_error_recovery_functions() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 5: Error Recovery Functions${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    RECOVERY_FUNCTIONS=(
        "fetchComparisonData"
        "fetchRiskScore"
        "fetchExplanation"
        "fetchAlerts"
        "fetchRecommendations"
        "loadAssessment"
    )
    
    FOUND_FUNCTIONS=0
    for func in "${RECOVERY_FUNCTIONS[@]}"; do
        if grep -r "${func}" frontend/components/merchant/*.tsx 2>/dev/null | head -1 > /dev/null; then
            FOUND_FUNCTIONS=$((FOUND_FUNCTIONS + 1))
        fi
    done
    
    if [ "${FOUND_FUNCTIONS}" -ge 4 ]; then
        echo -e "${GREEN}✅ Found ${FOUND_FUNCTIONS} error recovery functions${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${YELLOW}⚠️  Found ${FOUND_FUNCTIONS} error recovery functions (expected at least 4)${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
}

# Run all tests
test_retry_buttons
test_error_codes_in_messages
test_ctas_in_error_states
test_button_text
test_error_recovery_functions

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 CTA Verification Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
echo ""

if [ "${FAILED_TESTS}" -eq 0 ]; then
    echo -e "${GREEN}✅ All CTA verification tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some tests failed. Please review the output above.${NC}"
    exit 1
fi

