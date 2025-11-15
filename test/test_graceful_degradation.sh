#!/bin/bash

# Graceful Degradation Test Script
# Tests that the system works when database/Redis are unavailable

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Graceful Degradation Test"
echo "=========================================="
echo "This test verifies that the system gracefully handles"
echo "missing database and Redis connections."
echo ""

# Test 1: Health check should show database status
echo -e "${YELLOW}Test 1: Health Check${NC}"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health/detailed")
echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"

DB_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.checks.postgres.status // "unknown"' 2>/dev/null || echo "unknown")
REDIS_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.checks.redis.status // "unknown"' 2>/dev/null || echo "unknown")

echo "Database status: $DB_STATUS"
echo "Redis status: $REDIS_STATUS"

# Test 2: Threshold endpoints should work even without database (in-memory fallback)
echo -e "\n${YELLOW}Test 2: Threshold Endpoints (In-Memory Fallback)${NC}"

# GET thresholds - should work with empty response if no database
GET_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/risk/thresholds")
HTTP_CODE=$(echo "$GET_RESPONSE" | tail -n1)
BODY=$(echo "$GET_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ GET thresholds works (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY" | jq '.' 2>/dev/null || echo "Response: $BODY"
else
    echo -e "${RED}✗ GET thresholds failed (HTTP $HTTP_CODE)${NC}"
fi

# CREATE threshold - should work with in-memory storage if no database
THRESHOLD_DATA='{
  "name": "In-Memory Test Threshold",
  "category": "financial",
  "risk_levels": {
    "low": 25.0,
    "medium": 50.0,
    "high": 75.0,
    "critical": 90.0
  },
  "is_active": true
}'

CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/admin/risk/thresholds" \
  -H "Content-Type: application/json" \
  -d "$THRESHOLD_DATA")

CREATE_HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
CREATE_BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if [ "$CREATE_HTTP_CODE" -eq 201 ] || [ "$CREATE_HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ CREATE threshold works (HTTP $CREATE_HTTP_CODE)${NC}"
    echo "Response: $CREATE_BODY" | jq '.' 2>/dev/null || echo "Response: $CREATE_BODY"
else
    echo -e "${YELLOW}⚠ CREATE threshold returned HTTP $CREATE_HTTP_CODE${NC}"
    echo "This may be expected if database is required for persistence"
    echo "Response: $CREATE_BODY"
fi

# Test 3: Classification endpoint should work without Redis (no caching)
echo -e "\n${YELLOW}Test 3: Classification Endpoint (Without Redis)${NC}"
CLASSIFY_DATA='{
  "name": "Test Business",
  "description": "Test business for classification"
}'

CLASSIFY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "$CLASSIFY_DATA")

CLASSIFY_HTTP_CODE=$(echo "$CLASSIFY_RESPONSE" | tail -n1)
CLASSIFY_BODY=$(echo "$CLASSIFY_RESPONSE" | sed '$d')

if [ "$CLASSIFY_HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Classification works (HTTP $CLASSIFY_HTTP_CODE)${NC}"
    CACHE_HEADER=$(curl -s -I -X POST "$BASE_URL/v1/classify" \
      -H "Content-Type: application/json" \
      -d "$CLASSIFY_DATA" | grep -i "X-Cache" || echo "")
    if [ -n "$CACHE_HEADER" ]; then
        echo "Cache status: $CACHE_HEADER"
    else
        echo "No cache header (Redis not available or not configured)"
    fi
else
    echo -e "${RED}✗ Classification failed (HTTP $CLASSIFY_HTTP_CODE)${NC}"
fi

echo -e "\n${GREEN}✓ Graceful degradation test complete${NC}"
echo ""
echo "Summary:"
echo "- System should work with or without database (in-memory fallback)"
echo "- System should work with or without Redis (no caching)"
echo "- Health check should report status of all services"

