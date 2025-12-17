#!/bin/bash
# Verify Dashboard Endpoints Deployment
# Checks if dashboard endpoints are available after Railway deployment

set -e

API_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "🔍 Verifying Dashboard Endpoints Deployment"
echo "==========================================="
echo ""
echo "Service URL: $API_URL"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test dashboard summary endpoint
echo "Testing /api/dashboard/summary..."
SUMMARY_RESPONSE=$(curl -s --max-time 10 "${API_URL}/api/dashboard/summary?days=7" 2>&1)
SUMMARY_STATUS=$(echo "$SUMMARY_RESPONSE" | head -1)

if echo "$SUMMARY_RESPONSE" | jq . > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Dashboard summary endpoint is working!${NC}"
    echo "$SUMMARY_RESPONSE" | jq '.' | head -20
elif [ "$SUMMARY_STATUS" = "404" ] || echo "$SUMMARY_RESPONSE" | grep -q "404"; then
    echo -e "${YELLOW}⚠️  Dashboard summary endpoint not found (404)${NC}"
    echo "   This means the code hasn't been deployed yet or routes aren't registered"
else
    echo -e "${RED}❌ Dashboard summary endpoint error${NC}"
    echo "$SUMMARY_RESPONSE" | head -5
fi

echo ""

# Test dashboard timeseries endpoint
echo "Testing /api/dashboard/timeseries..."
TIMESERIES_RESPONSE=$(curl -s --max-time 10 "${API_URL}/api/dashboard/timeseries?days=7" 2>&1)
TIMESERIES_STATUS=$(echo "$TIMESERIES_RESPONSE" | head -1)

if echo "$TIMESERIES_RESPONSE" | jq . > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Dashboard timeseries endpoint is working!${NC}"
    echo "$TIMESERIES_RESPONSE" | jq '.' | head -20
elif [ "$TIMESERIES_STATUS" = "404" ] || echo "$TIMESERIES_RESPONSE" | grep -q "404"; then
    echo -e "${YELLOW}⚠️  Dashboard timeseries endpoint not found (404)${NC}"
    echo "   This means the code hasn't been deployed yet or routes aren't registered"
else
    echo -e "${RED}❌ Dashboard timeseries endpoint error${NC}"
    echo "$TIMESERIES_RESPONSE" | head -5
fi

echo ""
echo "==========================================="
echo "Deployment Status Check Complete"
echo "==========================================="
echo ""
echo "If endpoints return 404:"
echo "  1. Check Railway deployment logs"
echo "  2. Verify code was pushed to GitHub"
echo "  3. Wait a few minutes for Railway to deploy"
echo "  4. Check Railway dashboard for deployment status"
echo ""

