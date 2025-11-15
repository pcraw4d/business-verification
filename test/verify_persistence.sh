#!/bin/bash

# Verify Persistence After Server Restart
# This script verifies that thresholds created before restart still exist

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-./test_output}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Verify Persistence After Server Restart"
echo "=========================================="

if [ ! -f "$TEST_OUTPUT_DIR/persistence_threshold_id.txt" ]; then
    echo -e "${RED}✗ No threshold ID file found${NC}"
    echo "Run test_database_persistence.sh first to create a threshold"
    exit 1
fi

THRESHOLD_ID=$(cat "$TEST_OUTPUT_DIR/persistence_threshold_id.txt")

echo "Looking for threshold ID: $THRESHOLD_ID"
echo ""

# Check if threshold exists
GET_RESPONSE=$(curl -s "$BASE_URL/v1/risk/thresholds")
THRESHOLD_FOUND=$(echo "$GET_RESPONSE" | jq -r ".thresholds[] | select(.category == \"financial\") | .category" 2>/dev/null || echo "")

if [ -n "$THRESHOLD_FOUND" ]; then
    echo -e "${GREEN}✓ SUCCESS: Threshold persisted after server restart!${NC}"
    echo ""
    echo "Threshold details:"
    echo "$GET_RESPONSE" | jq ".thresholds[] | select(.category == \"financial\")" 2>/dev/null
    exit 0
else
    echo -e "${RED}✗ FAILED: Threshold not found after server restart${NC}"
    echo ""
    echo "This could mean:"
    echo "1. Database is not configured (using in-memory storage)"
    echo "2. Database connection failed"
    echo "3. Threshold was not properly saved"
    echo ""
    echo "Current thresholds:"
    echo "$GET_RESPONSE" | jq '.' 2>/dev/null || echo "$GET_RESPONSE"
    exit 1
fi

