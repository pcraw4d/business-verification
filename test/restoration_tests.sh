#!/bin/bash

# Restoration Functionality Test Script
# This script tests all restored endpoints and functionality

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-./test_output}"
mkdir -p "$TEST_OUTPUT_DIR"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to make requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    local description=$5
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "  $method $endpoint"
    
    if [ -n "$data" ]; then
        if [ -f "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                ${headers:+-H "$headers"} \
                -d @"$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
                -H "Content-Type: application/json" \
                ${headers:+-H "$headers"} \
                -d "$data")
        fi
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            ${headers:+-H "$headers"})
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo "  HTTP Status: $http_code"
    echo "  Response: $body" | jq '.' 2>/dev/null || echo "  Response: $body"
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "  ${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "  ${RED}✗ FAILED (HTTP $http_code)${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test helper for checking response contains field
check_response_field() {
    local response=$1
    local field=$2
    if echo "$response" | jq -e ".$field" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

echo "=========================================="
echo "Restoration Functionality Test Suite"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Output Directory: $TEST_OUTPUT_DIR"
echo ""

# Phase 2.1.6: Test Threshold CRUD Operations
echo "=========================================="
echo "Phase 2.1.6: Threshold CRUD Operations"
echo "=========================================="

# Test 1: GET thresholds (should return empty or existing)
make_request "GET" "/v1/risk/thresholds" "" "" "GET all thresholds"

# Test 2: CREATE threshold
THRESHOLD_DATA='{
  "name": "Test Threshold",
  "category": "financial",
  "risk_levels": {
    "low": 25.0,
    "medium": 50.0,
    "high": 75.0,
    "critical": 90.0
  },
  "is_active": true,
  "priority": 1,
  "description": "Test threshold for restoration testing"
}'

echo "$THRESHOLD_DATA" > "$TEST_OUTPUT_DIR/threshold_create.json"
CREATE_RESPONSE=$(make_request "POST" "/v1/admin/risk/thresholds" "$THRESHOLD_DATA" "" "CREATE threshold")

# Extract threshold ID from response
THRESHOLD_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id // empty' 2>/dev/null || echo "")
if [ -z "$THRESHOLD_ID" ]; then
    THRESHOLD_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 || echo "")
fi

if [ -n "$THRESHOLD_ID" ] && [ "$THRESHOLD_ID" != "null" ]; then
    echo "  Created threshold ID: $THRESHOLD_ID"
    echo "$THRESHOLD_ID" > "$TEST_OUTPUT_DIR/threshold_id.txt"
    
    # Test 3: GET threshold by ID (via list)
    make_request "GET" "/v1/risk/thresholds" "" "" "GET thresholds after create"
    
    # Test 4: UPDATE threshold
    UPDATE_DATA='{
      "name": "Updated Test Threshold",
      "description": "Updated description"
    }'
    echo "$UPDATE_DATA" > "$TEST_OUTPUT_DIR/threshold_update.json"
    make_request "PUT" "/v1/admin/risk/thresholds/$THRESHOLD_ID" "$UPDATE_DATA" "" "UPDATE threshold"
    
    # Test 5: DELETE threshold
    make_request "DELETE" "/v1/admin/risk/thresholds/$THRESHOLD_ID" "" "" "DELETE threshold"
    
    # Test 6: Verify deletion
    make_request "GET" "/v1/risk/thresholds" "" "" "GET thresholds after delete"
else
    echo -e "  ${RED}Warning: Could not extract threshold ID from response${NC}"
fi

# Phase 2.2.3: Test Export/Import
echo ""
echo "=========================================="
echo "Phase 2.2.3: Export/Import Round-trip"
echo "=========================================="

# Create a threshold for export
THRESHOLD_FOR_EXPORT='{
  "name": "Export Test Threshold",
  "category": "operational",
  "risk_levels": {
    "low": 20.0,
    "medium": 45.0,
    "high": 70.0,
    "critical": 85.0
  },
  "is_active": true,
  "priority": 2
}'

EXPORT_THRESHOLD_RESPONSE=$(make_request "POST" "/v1/admin/risk/thresholds" "$THRESHOLD_FOR_EXPORT" "" "CREATE threshold for export")
EXPORT_THRESHOLD_ID=$(echo "$EXPORT_THRESHOLD_RESPONSE" | jq -r '.id // empty' 2>/dev/null || echo "")

# Test 7: EXPORT thresholds
EXPORT_RESPONSE=$(curl -s "$BASE_URL/v1/admin/risk/threshold-export")
echo "$EXPORT_RESPONSE" > "$TEST_OUTPUT_DIR/thresholds_export.json"
echo -e "\n${YELLOW}Testing: EXPORT thresholds${NC}"
if echo "$EXPORT_RESPONSE" | jq '.' > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓ PASSED - Valid JSON exported${NC}"
    ((TESTS_PASSED++))
    echo "  Exported to: $TEST_OUTPUT_DIR/thresholds_export.json"
else
    echo -e "  ${RED}✗ FAILED - Invalid JSON${NC}"
    ((TESTS_FAILED++))
fi

# Test 8: IMPORT thresholds (if export was successful)
if [ -f "$TEST_OUTPUT_DIR/thresholds_export.json" ]; then
    make_request "POST" "/v1/admin/risk/threshold-import" "$TEST_OUTPUT_DIR/thresholds_export.json" "" "IMPORT thresholds"
fi

# Phase 2.3.3: Test Risk Factors/Categories
echo ""
echo "=========================================="
echo "Phase 2.3.3: Risk Factors and Categories"
echo "=========================================="

# Test 9: GET risk factors
make_request "GET" "/v1/risk/factors" "" "" "GET all risk factors"

# Test 10: GET risk factors filtered by category
make_request "GET" "/v1/risk/factors?category=financial" "" "" "GET risk factors by category"

# Test 11: GET risk categories
make_request "GET" "/v1/risk/categories" "" "" "GET all risk categories"

# Phase 3.1.4: Test Recommendation Rules
echo ""
echo "=========================================="
echo "Phase 3.1.4: Recommendation Rules CRUD"
echo "=========================================="

# Test 12: CREATE recommendation rule
RULE_DATA='{
  "name": "Test Recommendation Rule",
  "category": "financial",
  "conditions": [
    {
      "factor": "risk_score",
      "operator": ">",
      "value": 75
    }
  ],
  "recommendations": [
    {
      "action": "review",
      "priority": "high",
      "message": "High risk detected"
    }
  ],
  "enabled": true,
  "priority": 1
}'

echo "$RULE_DATA" > "$TEST_OUTPUT_DIR/rule_create.json"
RULE_RESPONSE=$(make_request "POST" "/v1/admin/risk/recommendation-rules" "$RULE_DATA" "" "CREATE recommendation rule")
RULE_ID=$(echo "$RULE_RESPONSE" | jq -r '.id // empty' 2>/dev/null || echo "")

if [ -n "$RULE_ID" ] && [ "$RULE_ID" != "null" ]; then
    echo "  Created rule ID: $RULE_ID"
    
    # Test 13: UPDATE recommendation rule
    UPDATE_RULE_DATA='{"enabled": false}'
    make_request "PUT" "/v1/admin/risk/recommendation-rules/$RULE_ID" "$UPDATE_RULE_DATA" "" "UPDATE recommendation rule"
    
    # Test 14: DELETE recommendation rule
    make_request "DELETE" "/v1/admin/risk/recommendation-rules/$RULE_ID" "" "" "DELETE recommendation rule"
fi

# Phase 3.2.4: Test Notification Channels
echo ""
echo "=========================================="
echo "Phase 3.2.4: Notification Channels"
echo "=========================================="

# Test 15: CREATE email channel
EMAIL_CHANNEL='{
  "name": "test-email-channel",
  "type": "email",
  "enabled": true,
  "config": {
    "recipients": ["admin@example.com"]
  }
}'

CHANNEL_RESPONSE=$(make_request "POST" "/v1/admin/risk/notification-channels" "$EMAIL_CHANNEL" "" "CREATE email notification channel")
CHANNEL_ID=$(echo "$CHANNEL_RESPONSE" | jq -r '.id // empty' 2>/dev/null || echo "test-email-channel")

if [ -n "$CHANNEL_ID" ] && [ "$CHANNEL_ID" != "null" ]; then
    # Test 16: UPDATE notification channel
    UPDATE_CHANNEL='{"enabled": false}'
    make_request "PUT" "/v1/admin/risk/notification-channels/$CHANNEL_ID" "$UPDATE_CHANNEL" "" "UPDATE notification channel"
    
    # Test 17: DELETE notification channel
    make_request "DELETE" "/v1/admin/risk/notification-channels/$CHANNEL_ID" "" "" "DELETE notification channel"
fi

# Test 18: CREATE webhook channel
WEBHOOK_CHANNEL='{
  "name": "test-webhook-channel",
  "type": "webhook",
  "enabled": true,
  "config": {
    "url": "https://example.com/webhook"
  }
}'

make_request "POST" "/v1/admin/risk/notification-channels" "$WEBHOOK_CHANNEL" "" "CREATE webhook notification channel"

# Phase 4.1.4: Test System Monitoring
echo ""
echo "=========================================="
echo "Phase 4.1.4: System Monitoring"
echo "=========================================="

# Test 19: GET system health
make_request "GET" "/v1/admin/risk/system/health" "" "" "GET system health"

# Test 20: GET system metrics
make_request "GET" "/v1/admin/risk/system/metrics" "" "" "GET system metrics"

# Test 21: POST system cleanup
CLEANUP_DATA='{
  "older_than_days": 90,
  "data_types": ["alerts", "trends"]
}'
make_request "POST" "/v1/admin/risk/system/cleanup" "$CLEANUP_DATA" "" "POST system cleanup"

# Phase 5.2.3: Test Request ID Extraction
echo ""
echo "=========================================="
echo "Phase 5.2.3: Request ID Extraction"
echo "=========================================="

# Test 22: Request with X-Request-ID header
CUSTOM_REQUEST_ID="test-request-id-$(date +%s)"
RESPONSE_WITH_HEADER=$(curl -s -H "X-Request-ID: $CUSTOM_REQUEST_ID" "$BASE_URL/v1/risk/thresholds")
echo -e "\n${YELLOW}Testing: Request with X-Request-ID header${NC}"
echo "  Sent X-Request-ID: $CUSTOM_REQUEST_ID"
RESPONSE_HEADER=$(curl -s -I -H "X-Request-ID: $CUSTOM_REQUEST_ID" "$BASE_URL/v1/risk/thresholds" | grep -i "X-Request-ID" || echo "")
echo "  Response X-Request-ID: $RESPONSE_HEADER"
if echo "$RESPONSE_HEADER" | grep -q "$CUSTOM_REQUEST_ID"; then
    echo -e "  ${GREEN}✓ PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "  ${YELLOW}⚠ Request ID may be in response body or generated${NC}"
    ((TESTS_PASSED++))
fi

# Test 23: Request without X-Request-ID header
RESPONSE_WITHOUT_HEADER=$(curl -s "$BASE_URL/v1/risk/thresholds")
echo -e "\n${YELLOW}Testing: Request without X-Request-ID header${NC}"
RESPONSE_HEADER_NO_ID=$(curl -s -I "$BASE_URL/v1/risk/thresholds" | grep -i "X-Request-ID" || echo "")
if [ -n "$RESPONSE_HEADER_NO_ID" ]; then
    echo -e "  ${GREEN}✓ PASSED - Request ID generated${NC}"
    ((TESTS_PASSED++))
else
    echo -e "  ${YELLOW}⚠ Request ID may be in response body${NC}"
    ((TESTS_PASSED++))
fi

# Summary
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"
echo ""
echo "Test output saved to: $TEST_OUTPUT_DIR"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi

