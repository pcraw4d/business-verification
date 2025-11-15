#!/bin/bash

# Database Persistence Test Script
# Tests that thresholds persist across server restarts

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-./test_output}"
mkdir -p "$TEST_OUTPUT_DIR"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Database Persistence Test"
echo "=========================================="
echo "This test verifies that thresholds persist in the database"
echo "and survive server restarts."
echo ""

# Step 1: Create a threshold
echo -e "${YELLOW}Step 1: Creating a threshold${NC}"
THRESHOLD_DATA='{
  "name": "Persistence Test Threshold",
  "category": "financial",
  "risk_levels": {
    "low": 25.0,
    "medium": 50.0,
    "high": 75.0,
    "critical": 90.0
  },
  "is_active": true,
  "priority": 1,
  "description": "This threshold should persist after server restart"
}'

CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/admin/risk/thresholds" \
  -H "Content-Type: application/json" \
  -d "$THRESHOLD_DATA")

THRESHOLD_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id // empty' 2>/dev/null || echo "")

if [ -z "$THRESHOLD_ID" ] || [ "$THRESHOLD_ID" = "null" ]; then
    echo -e "${RED}✗ Failed to create threshold${NC}"
    echo "Response: $CREATE_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Created threshold with ID: $THRESHOLD_ID${NC}"
echo "$THRESHOLD_ID" > "$TEST_OUTPUT_DIR/persistence_threshold_id.txt"

# Step 2: Verify threshold exists
echo -e "\n${YELLOW}Step 2: Verifying threshold exists${NC}"
GET_RESPONSE=$(curl -s "$BASE_URL/v1/risk/thresholds")
THRESHOLD_FOUND=$(echo "$GET_RESPONSE" | jq -r ".thresholds[] | select(.category == \"financial\") | .category" 2>/dev/null || echo "")

if [ -n "$THRESHOLD_FOUND" ]; then
    echo -e "${GREEN}✓ Threshold found in database${NC}"
else
    echo -e "${RED}✗ Threshold not found${NC}"
    exit 1
fi

# Step 3: Get threshold details
echo -e "\n${YELLOW}Step 3: Getting threshold details${NC}"
THRESHOLD_DETAILS=$(curl -s "$BASE_URL/v1/risk/thresholds" | jq ".thresholds[] | select(.category == \"financial\")" 2>/dev/null || echo "")
echo "Threshold details:"
echo "$THRESHOLD_DETAILS" | jq '.' 2>/dev/null || echo "$THRESHOLD_DETAILS"

# Step 4: Instructions for manual server restart test
echo -e "\n${YELLOW}Step 4: Manual Server Restart Test${NC}"
echo "To complete the persistence test:"
echo "1. Stop the server (Ctrl+C or kill process)"
echo "2. Restart the server"
echo "3. Run the verification script:"
echo "   ./test/verify_persistence.sh"
echo ""
echo "Or run this command to verify:"
echo "curl -s $BASE_URL/v1/risk/thresholds | jq '.thresholds[] | select(.category == \"financial\")'"

echo -e "\n${GREEN}✓ Database persistence test setup complete${NC}"
echo "Threshold ID saved to: $TEST_OUTPUT_DIR/persistence_threshold_id.txt"

