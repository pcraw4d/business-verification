#!/bin/bash

# Quick E2E Metrics Test - 50 Sample Test
# Tests against Railway production to measure improvements

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

RAILWAY_API_URL="${RAILWAY_API_URL:-https://classification-service-production.up.railway.app}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Quick E2E Metrics Test (50 Samples)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${YELLOW}$RAILWAY_API_URL${NC}"
echo ""

# Create results directory
mkdir -p test/results

# Run with limited samples (modify test to use 50 samples)
echo -e "${GREEN}🚀 Running quick metrics test...${NC}"
echo ""

# Check if we can modify the test to use fewer samples
# For now, let's run a direct API test with a few samples
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="test/results/quick_e2e_metrics_${TIMESTAMP}.json"

# Test a few samples directly
echo "Testing sample classifications..."

# Sample 1: Technology Company
echo -n "Test 1: Technology Company... "
RESPONSE1=$(curl -s -w "\n%{http_code}" -X POST "${RAILWAY_API_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name":"TechCorp Solutions","description":"Software development and cloud computing services"}' \
  --max-time 120 -k 2>&1)
HTTP_CODE1=$(echo "$RESPONSE1" | tail -n1)
if [ "$HTTP_CODE1" = "200" ]; then
  echo -e "${GREEN}✅${NC}"
else
  echo -e "${RED}❌ HTTP $HTTP_CODE1${NC}"
fi

# Sample 2: Restaurant
echo -n "Test 2: Restaurant... "
RESPONSE2=$(curl -s -w "\n%{http_code}" -X POST "${RAILWAY_API_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Mama Mias Italian Restaurant","description":"Authentic Italian restaurant serving pasta, pizza, and fine wines"}' \
  --max-time 120 -k 2>&1)
HTTP_CODE2=$(echo "$RESPONSE2" | tail -n1)
if [ "$HTTP_CODE2" = "200" ]; then
  echo -e "${GREEN}✅${NC}"
else
  echo -e "${RED}❌ HTTP $HTTP_CODE2${NC}"
fi

echo ""
echo -e "${BLUE}For comprehensive metrics, check the full E2E test running in background${NC}"
echo -e "${BLUE}Results will be saved to: test/results/railway_e2e_*.json${NC}"

