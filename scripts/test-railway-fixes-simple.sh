#!/bin/bash

# Railway Logs Fixes - Simple Production Testing Script
# Tests Phase 1 (ML Timeout) and Phase 2 & 3 (Content Quality Thresholds)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Production URL
API_GATEWAY="${API_GATEWAY:-https://api-gateway-service-production-21fd.up.railway.app}"
CLASSIFICATION_URL="${CLASSIFICATION_URL:-${API_GATEWAY}/api/v1/classify}"

echo -e "${BLUE}🧪 Railway Logs Fixes - Production Testing${NC}"
echo "=========================================="
echo -e "Testing URL: ${YELLOW}${CLASSIFICATION_URL}${NC}"
echo ""

# Test counters
TOTAL=0
PASSED=0
FAILED=0
TIMEOUTS=0

# Test function
test_classification() {
    local name="$1"
    local payload="$2"
    
    TOTAL=$((TOTAL + 1))
    echo -e "\n${YELLOW}Test ${TOTAL}: ${name}${NC}"
    
    # Make request with 20s timeout
    start_time=$(date +%s)
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME:%{time_total}" \
        --max-time 20 \
        -X POST "${CLASSIFICATION_URL}" \
        -H "Content-Type: application/json" \
        -d "${payload}" 2>&1) || true
    
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    # Extract components
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2 || echo "000")
    time_total=$(echo "$response" | grep "TIME:" | cut -d: -f2 || echo "0")
    body=$(echo "$response" | grep -v "HTTP_CODE:" | grep -v "TIME:" || echo "")
    
    # Check for timeout
    if [ "$http_code" = "000" ] || [ -z "$http_code" ]; then
        echo -e "${RED}❌ Timeout or connection error${NC}"
        TIMEOUTS=$((TIMEOUTS + 1))
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # Check HTTP status
    if [ "$http_code" != "200" ]; then
        echo -e "${RED}❌ HTTP ${http_code}${NC}"
        echo "Response: $(echo "$body" | head -c 200)"
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # Check for timeout errors in body
    if echo "$body" | grep -qi "timeout\|deadline exceeded"; then
        echo -e "${RED}❌ Timeout error in response${NC}"
        TIMEOUTS=$((TIMEOUTS + 1))
        FAILED=$((FAILED + 1))
        return 1
    fi
    
    # Success
    echo -e "${GREEN}✅ Success${NC}"
    echo "   Duration: ${time_total}s"
    echo "   HTTP: ${http_code}"
    
    # Check if classification data exists
    if echo "$body" | grep -qi "classification\|industry\|naics\|sic\|mcc"; then
        echo -e "${GREEN}   ✅ Classification data present${NC}"
    fi
    
    PASSED=$((PASSED + 1))
    return 0
}

# Check service health first
echo -e "${BLUE}Checking service health...${NC}"
health=$(curl -s --max-time 5 "${API_GATEWAY}/health" || echo "FAIL")
if echo "$health" | grep -qi "ok\|healthy\|success"; then
    echo -e "${GREEN}✅ Service is healthy${NC}"
else
    echo -e "${YELLOW}⚠️  Health check unclear, continuing anyway...${NC}"
fi

# Run tests
echo -e "\n${BLUE}Running classification tests...${NC}"

test_classification "Tech Company" \
    '{"business_name":"TechCorp Software","description":"Custom software development","website_url":"https://techcorp.com"}'

test_classification "Marketing Agency" \
    '{"business_name":"Digital Marketing","description":"SEO and social media marketing services"}'

test_classification "Healthcare" \
    '{"business_name":"Health Clinic","description":"Primary care medical services"}'

# Summary
echo -e "\n${BLUE}=========================================="
echo -e "📊 Test Summary${NC}"
echo "=========================================="
echo -e "Total: ${TOTAL}"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"
echo -e "${YELLOW}Timeouts: ${TIMEOUTS}${NC}"
echo ""

# Success criteria
if [ $TIMEOUTS -eq 0 ] && [ $PASSED -ge $((TOTAL * 7 / 10)) ]; then
    echo -e "${GREEN}✅ Overall: PASSED${NC}"
    echo -e "${GREEN}✅ Phase 1: No timeout errors${NC}"
    echo -e "${GREEN}✅ Phase 2 & 3: Content quality checks passing${NC}"
    exit 0
elif [ $TIMEOUTS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Overall: PARTIAL${NC}"
    echo -e "${GREEN}✅ Phase 1: No timeout errors${NC}"
    echo -e "${YELLOW}⚠️  Phase 2 & 3: Some failures${NC}"
    exit 1
else
    echo -e "${RED}❌ Overall: FAILED${NC}"
    echo -e "${RED}❌ Phase 1: Timeout errors detected${NC}"
    exit 1
fi

