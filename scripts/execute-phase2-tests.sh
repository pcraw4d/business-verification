#!/bin/bash

# Execute Phase 2 Manual Testing Checklist
# This script helps verify Phase 2 implementation by checking:
# - Error codes are implemented
# - Error messages include codes
# - CTAs are present
# - Type guards are in place
# Usage: ./scripts/execute-phase2-tests.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Phase 2 Testing Checklist Execution${NC}"
echo ""
echo -e "${YELLOW}This script will help verify Phase 2 implementation.${NC}"
echo -e "${YELLOW}Some tests require manual browser verification.${NC}"
echo ""

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Function to test error codes in codebase
test_error_codes() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 1: Error Code Implementation${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check if error-codes.ts exists
    if [ ! -f "frontend/lib/error-codes.ts" ]; then
        echo -e "${RED}✗ Error codes file not found${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo -e "${GREEN}✓ Error codes file exists${NC}"
    
    # Check for required error codes
    REQUIRED_CODES=(
        "PC-001" "PC-002" "PC-003" "PC-004" "PC-005"
        "RS-001" "RS-002" "RS-003"
        "AC-001" "AC-002" "AC-003" "AC-004" "AC-005"
        "RB-001" "RB-002" "RB-003" "RB-004" "RB-005"
    )
    
    MISSING_CODES=()
    for code in "${REQUIRED_CODES[@]}"; do
        if ! grep -q "$code" frontend/lib/error-codes.ts; then
            MISSING_CODES+=("$code")
        fi
    done
    
    if [ ${#MISSING_CODES[@]} -eq 0 ]; then
        echo -e "${GREEN}✓ All required error codes found${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Missing error codes: ${MISSING_CODES[*]}${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo ""
}

# Function to test formatErrorWithCode usage
test_error_formatting() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 2: Error Message Formatting${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    COMPONENTS=(
        "frontend/components/merchant/PortfolioComparisonCard.tsx"
        "frontend/components/merchant/RiskScoreCard.tsx"
        "frontend/components/merchant/AnalyticsComparison.tsx"
        "frontend/components/merchant/RiskBenchmarkComparison.tsx"
    )
    
    MISSING_FORMAT=()
    for component in "${COMPONENTS[@]}"; do
        if [ -f "$component" ]; then
            if ! grep -q "formatErrorWithCode" "$component"; then
                MISSING_FORMAT+=("$component")
            fi
        fi
    done
    
    if [ ${#MISSING_FORMAT[@]} -eq 0 ]; then
        echo -e "${GREEN}✓ All components use formatErrorWithCode${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ Components missing formatErrorWithCode:${NC}"
        for comp in "${MISSING_FORMAT[@]}"; do
            echo -e "${RED}   - $comp${NC}"
        done
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo ""
}

# Function to test CTA buttons
test_cta_buttons() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 3: CTA Buttons in Error States${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check for CTA buttons in components
    CTA_PATTERNS=(
        "Run Risk Assessment"
        "Start Risk Assessment"
        "Refresh Data"
        "Retry"
        "Enrich Data"
    )
    
    FOUND_CTAS=0
    for pattern in "${CTA_PATTERNS[@]}"; do
        if grep -r "$pattern" frontend/components/merchant/*.tsx 2>/dev/null | head -1 > /dev/null; then
            FOUND_CTAS=$((FOUND_CTAS + 1))
        fi
    done
    
    if [ $FOUND_CTAS -ge 3 ]; then
        echo -e "${GREEN}✓ CTA buttons found in components (found $FOUND_CTAS patterns)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Limited CTA buttons found (found $FOUND_CTAS patterns)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo ""
}

# Function to test type guards
test_type_guards() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 4: Type Guards and Validation${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    TYPE_GUARD_PATTERNS=(
        "isValidRiskScore"
        "hasValidPortfolioStats"
        "hasValidMerchantRiskScore"
        "typeof.*==="
        "instanceof"
    )
    
    FOUND_GUARDS=0
    for pattern in "${TYPE_GUARD_PATTERNS[@]}"; do
        if grep -r "$pattern" frontend/components/merchant/*.tsx 2>/dev/null | head -1 > /dev/null; then
            FOUND_GUARDS=$((FOUND_GUARDS + 1))
        fi
    done
    
    if [ $FOUND_GUARDS -ge 2 ]; then
        echo -e "${GREEN}✓ Type guards found in components (found $FOUND_GUARDS patterns)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Limited type guards found (found $FOUND_GUARDS patterns)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo ""
}

# Function to test safeFetch usage
test_safefetch() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 5: safeFetch Implementation${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check if safeFetch exists
    if ! grep -q "safeFetch" frontend/lib/api.ts; then
        echo -e "${RED}✗ safeFetch not found in api.ts${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Check if safeFetch is used
    SAFEFETCH_COUNT=$(grep -c "safeFetch" frontend/lib/api.ts || echo "0")
    if [ "$SAFEFETCH_COUNT" -gt 5 ]; then
        echo -e "${GREEN}✓ safeFetch implemented and used ($SAFEFETCH_COUNT occurrences)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${YELLOW}⚠ safeFetch found but limited usage ($SAFEFETCH_COUNT occurrences)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    echo ""
}

# Function to check services are running
test_services() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}Test 6: Services Running${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Check frontend
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Frontend running on http://localhost:3000${NC}"
    else
        echo -e "${RED}✗ Frontend not running on http://localhost:3000${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Check API Gateway
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ API Gateway running on http://localhost:8080${NC}"
    else
        echo -e "${RED}✗ API Gateway not running on http://localhost:8080${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Check Merchant Service
    if curl -s http://localhost:8083/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Merchant Service running on http://localhost:8083${NC}"
    else
        echo -e "${YELLOW}⚠ Merchant Service not running on http://localhost:8083${NC}"
    fi
    
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo ""
}

# Run all tests
echo -e "${BLUE}Starting Phase 2 automated tests...${NC}"
echo ""

test_error_codes
test_error_formatting
test_cta_buttons
test_type_guards
test_safefetch
test_services

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Phase 2 Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All automated tests passed!${NC}"
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Next Steps: Manual Browser Testing${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "1. Open browser: http://localhost:3000"
    echo "2. Navigate to a merchant details page"
    echo "3. Open DevTools console (F12)"
    echo "4. Test error states and verify:"
    echo "   - Error messages include codes (PC-001, RS-001, etc.)"
    echo "   - CTA buttons are visible and functional"
    echo "   - No console errors"
    echo "   - Loading states display correctly"
    echo ""
    echo "See: .cursor/plans/phase2-manual-test-checklist.md for detailed checklist"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Some tests failed!${NC}"
    echo -e "${YELLOW}   Please fix the issues above before proceeding with manual testing.${NC}"
    exit 1
fi

