#!/bin/bash

# Classification Service Test Script
# Tests the classification service with diverse business types after Railway deployment

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API Gateway URL (update if different)
API_GATEWAY_URL="${API_GATEWAY_URL:-https://api-gateway-service-production-21fd.up.railway.app}"

echo -e "${BLUE}🧪 Classification Service Test Suite${NC}"
echo "=========================================="
echo "API Gateway: $API_GATEWAY_URL"
echo ""

# Test counter
PASSED=0
FAILED=0
TOTAL=0

# Function to test classification
test_classification() {
    local test_name="$1"
    local business_name="$2"
    local description="$3"
    local expected_industry="$4"
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "${BLUE}Test $TOTAL: $test_name${NC}"
    echo "  Business: $business_name"
    echo "  Description: $description"
    echo "  Expected Industry: $expected_industry"
    
    # Make API call
    response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"$business_name\",
            \"description\": \"$description\"
        }")
    
    # Extract industry from response
    industry=$(echo "$response" | jq -r '.classification.industry // .classification.primary_industry // "unknown"' 2>/dev/null || echo "error")
    
    if [ "$industry" = "error" ] || [ -z "$industry" ]; then
        echo -e "  ${RED}❌ FAILED: Could not parse response${NC}"
        echo "  Response: $response"
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    echo "  Actual Industry: $industry"
    
    # Check if industry matches expected (case-insensitive)
    if echo "$industry" | grep -qi "$expected_industry"; then
        echo -e "  ${GREEN}✅ PASSED${NC}"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "  ${RED}❌ FAILED: Expected industry containing '$expected_industry', got '$industry'${NC}"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Function to test that it's NOT Food & Beverage (for non-restaurant businesses)
test_not_food_beverage() {
    local test_name="$1"
    local business_name="$2"
    local description="$3"
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "${BLUE}Test $TOTAL: $test_name${NC}"
    echo "  Business: $business_name"
    echo "  Description: $description"
    echo "  Should NOT be: Food & Beverage"
    
    # Make API call
    response=$(curl -s -X POST "$API_GATEWAY_URL/api/v1/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"$business_name\",
            \"description\": \"$description\"
        }")
    
    # Extract industry from response
    industry=$(echo "$response" | jq -r '.classification.industry // .classification.primary_industry // "unknown"' 2>/dev/null || echo "error")
    
    if [ "$industry" = "error" ] || [ -z "$industry" ]; then
        echo -e "  ${RED}❌ FAILED: Could not parse response${NC}"
        echo "  Response: $response"
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    echo "  Actual Industry: $industry"
    
    # Check if it's NOT Food & Beverage
    if echo "$industry" | grep -qi "food.*beverage\|beverage.*food"; then
        echo -e "  ${RED}❌ FAILED: Got 'Food & Beverage' for non-restaurant business${NC}"
        FAILED=$((FAILED + 1))
        return 1
    else
        echo -e "  ${GREEN}✅ PASSED: Not Food & Beverage${NC}"
        PASSED=$((FAILED + 1))
        return 0
    fi
}

# Run tests
echo -e "${YELLOW}Running classification tests...${NC}"
echo ""

# Test 1: Software Development Company
test_classification \
    "Software Development Company" \
    "Acme Software Solutions" \
    "Custom software development and cloud infrastructure services" \
    "Technology\|Software\|IT"

# Test 2: Medical Clinic
test_classification \
    "Medical Clinic" \
    "City Medical Clinic" \
    "Primary care medical clinic providing healthcare services and patient care" \
    "Healthcare\|Medical\|Health"

# Test 3: Financial Services
test_classification \
    "Financial Services" \
    "Global Financial Advisors" \
    "Investment advisory and wealth management services" \
    "Financial\|Finance\|Banking"

# Test 4: Retail Store
test_classification \
    "Retail Store" \
    "Fashion Forward Boutique" \
    "Retail store selling trendy clothing and accessories" \
    "Retail\|Apparel\|Fashion"

# Test 5: Restaurant (should be Food & Beverage)
test_classification \
    "Restaurant" \
    "The Gourmet Bistro" \
    "Fine dining restaurant serving French cuisine" \
    "Food.*Beverage\|Restaurant\|Food"

# Test 6: Tech Startup
test_classification \
    "Tech Startup" \
    "TechStart Innovations" \
    "AI-powered software solutions for business automation" \
    "Technology\|Software\|AI\|Tech"

# Test 7: Verify NOT Food & Beverage for Software
test_not_food_beverage \
    "Software Company (Not Food)" \
    "TechCorp Software" \
    "Enterprise software solutions"

# Test 8: Verify NOT Food & Beverage for Medical
test_not_food_beverage \
    "Medical Practice (Not Food)" \
    "HealthCare Plus" \
    "Medical practice providing healthcare services"

# Summary
echo ""
echo "=========================================="
echo -e "${BLUE}Test Summary${NC}"
echo "=========================================="
echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

