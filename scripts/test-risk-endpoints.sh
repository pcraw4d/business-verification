#!/bin/bash

# Test script for Risk API endpoints
# Tests benchmarks and predictions endpoints through API Gateway

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_MERCHANT_ID="${TEST_MERCHANT_ID:-test-merchant-123}"

echo -e "${YELLOW}🧪 Testing Risk API Endpoints${NC}"
echo "API Base URL: $API_BASE_URL"
echo "Test Merchant ID: $TEST_MERCHANT_ID"
echo ""

# Test 1: Benchmarks Endpoint
echo -e "${YELLOW}Test 1: GET /api/v1/risk/benchmarks${NC}"
echo "Testing with MCC code 5411..."

RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/v1/risk/benchmarks?mcc=5411")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Benchmarks endpoint returned 200${NC}"
    echo "Response preview:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY" | head -c 200
    echo ""
else
    echo -e "${RED}❌ Benchmarks endpoint returned $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 2: Benchmarks with NAICS
echo -e "${YELLOW}Test 2: GET /api/v1/risk/benchmarks (with NAICS)${NC}"
echo "Testing with NAICS code 541110..."

RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/v1/risk/benchmarks?naics=541110")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Benchmarks endpoint (NAICS) returned 200${NC}"
    echo "Response preview:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY" | head -c 200
    echo ""
else
    echo -e "${RED}❌ Benchmarks endpoint (NAICS) returned $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 3: Benchmarks Error Handling
echo -e "${YELLOW}Test 3: GET /api/v1/risk/benchmarks (error case)${NC}"
echo "Testing without industry codes (should return 400)..."

RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/v1/risk/benchmarks")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 400 ]; then
    echo -e "${GREEN}✅ Error handling works (returned 400 as expected)${NC}"
else
    echo -e "${YELLOW}⚠️  Expected 400, got $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 4: Predictions Endpoint
echo -e "${YELLOW}Test 4: GET /api/v1/risk/predictions/{merchant_id}${NC}"
echo "Testing with merchant ID: $TEST_MERCHANT_ID..."

RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/v1/risk/predictions/${TEST_MERCHANT_ID}?horizons=3,6,12&includeScenarios=true&includeConfidence=true")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Predictions endpoint returned 200${NC}"
    echo "Response preview:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY" | head -c 200
    echo ""
else
    echo -e "${RED}❌ Predictions endpoint returned $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 5: Predictions with Custom Horizons
echo -e "${YELLOW}Test 5: GET /api/v1/risk/predictions (custom horizons)${NC}"
echo "Testing with horizons: 6,12..."

RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/v1/risk/predictions/${TEST_MERCHANT_ID}?horizons=6,12")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Predictions endpoint (custom horizons) returned 200${NC}"
    echo "Response preview:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY" | head -c 200
    echo ""
else
    echo -e "${RED}❌ Predictions endpoint (custom horizons) returned $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Summary
echo -e "${YELLOW}📊 Test Summary${NC}"
echo "All endpoint tests completed."
echo ""
echo "To test with different configuration:"
echo "  API_BASE_URL=https://your-api.com ./scripts/test-risk-endpoints.sh"
echo "  TEST_MERCHANT_ID=your-merchant-id ./scripts/test-risk-endpoints.sh"

