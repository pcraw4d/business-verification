#!/bin/bash

# Deployment Verification Script
# Verifies that the API Gateway validation fix is deployed

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

API_BASE_URL="${API_BASE_URL:-https://api-gateway-service-production-21fd.up.railway.app}"

echo "=========================================="
echo "API Gateway Deployment Verification"
echo "=========================================="
echo "API Base URL: $API_BASE_URL"
echo ""

# Test 1: Health check
echo "Test 1: Health Check"
health_response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL/health")
health_code=$(echo "$health_response" | tail -n1)
if [ "$health_code" == "200" ]; then
    echo -e "${GREEN}✓ PASSED${NC} - Health check working"
else
    echo -e "${RED}✗ FAILED${NC} - Health check returned $health_code"
    exit 1
fi
echo ""

# Test 2: Missing required field (should return 400 after fix)
echo "Test 2: Missing Required Field Validation"
validation_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -w "\n%{http_code}" \
    -d '{"description":"Test without business_name"}' \
    "$API_BASE_URL/api/v1/classify")
validation_code=$(echo "$validation_response" | tail -n1)
validation_body=$(echo "$validation_response" | sed '$d')

if [ "$validation_code" == "400" ]; then
    echo -e "${GREEN}✓ PASSED${NC} - Validation fix deployed (returns 400)"
    echo "  Response: $validation_body" | head -c 200
    echo ""
elif [ "$validation_code" == "503" ]; then
    echo -e "${YELLOW}⚠ PENDING${NC} - Fix not yet deployed (still returns 503)"
    echo "  Deployment may still be in progress"
    echo "  Expected: 400 Bad Request"
    echo "  Actual: 503 Service Unavailable"
    exit 1
else
    echo -e "${RED}✗ FAILED${NC} - Unexpected status code: $validation_code"
    echo "  Response: $validation_body" | head -c 200
    exit 1
fi
echo ""

# Test 3: Valid request (should still work)
echo "Test 3: Valid Request"
valid_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -w "\n%{http_code}" \
    -d '{"business_name":"Test Company"}' \
    "$API_BASE_URL/api/v1/classify")
valid_code=$(echo "$valid_response" | tail -n1)

if [ "$valid_code" == "200" ]; then
    echo -e "${GREEN}✓ PASSED${NC} - Valid requests still work"
else
    echo -e "${RED}✗ FAILED${NC} - Valid request returned $valid_code"
    exit 1
fi
echo ""

echo "=========================================="
echo -e "${GREEN}✓ All verification tests passed!${NC}"
echo "=========================================="
echo "Deployment verification complete."
echo "The validation fix is successfully deployed."

