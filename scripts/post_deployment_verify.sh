#!/bin/bash

# Post-Deployment Verification Script
# Verifies deployment success by testing endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API URL (default to localhost, can be overridden)
API_URL="${API_URL:-http://localhost:8080}"

ERRORS=0

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Post-Deployment Verification${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""
echo -e "Testing API at: ${BLUE}$API_URL${NC}"
echo ""

# Check 1: Health Endpoint
echo -e "${GREEN}Check 1: Health Endpoint${NC}"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/health/detailed" || echo "ERROR")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
HEALTH_BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ Health endpoint accessible (200 OK)${NC}"
    
    # Check for database status
    if echo "$HEALTH_BODY" | grep -qi "database.*ok\|database.*healthy\|postgres.*ok"; then
        echo -e "  ${GREEN}✅ Database health check passed${NC}"
    else
        echo -e "  ${YELLOW}⚠️  Database status unclear${NC}"
    fi
else
    echo -e "  ${RED}❌ Health endpoint failed (HTTP $HTTP_CODE)${NC}"
    ((ERRORS++))
fi

# Check 2: Get Thresholds Endpoint
echo ""
echo -e "${GREEN}Check 2: Get Thresholds Endpoint${NC}"
THRESHOLDS_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/v1/risk/thresholds" || echo "ERROR")
HTTP_CODE=$(echo "$THRESHOLDS_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ Get thresholds endpoint works (200 OK)${NC}"
else
    echo -e "  ${RED}❌ Get thresholds endpoint failed (HTTP $HTTP_CODE)${NC}"
    ((ERRORS++))
fi

# Check 3: Risk Factors Endpoint
echo ""
echo -e "${GREEN}Check 3: Risk Factors Endpoint${NC}"
FACTORS_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/v1/risk/factors" || echo "ERROR")
HTTP_CODE=$(echo "$FACTORS_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ Risk factors endpoint works (200 OK)${NC}"
else
    echo -e "  ${RED}❌ Risk factors endpoint failed (HTTP $HTTP_CODE)${NC}"
    ((ERRORS++))
fi

# Check 4: Risk Categories Endpoint
echo ""
echo -e "${GREEN}Check 4: Risk Categories Endpoint${NC}"
CATEGORIES_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/v1/risk/categories" || echo "ERROR")
HTTP_CODE=$(echo "$CATEGORIES_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ Risk categories endpoint works (200 OK)${NC}"
else
    echo -e "  ${RED}❌ Risk categories endpoint failed (HTTP $HTTP_CODE)${NC}"
    ((ERRORS++))
fi

# Check 5: System Health Endpoint
echo ""
echo -e "${GREEN}Check 5: System Health Endpoint${NC}"
SYSTEM_HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/v1/admin/risk/system/health" || echo "ERROR")
HTTP_CODE=$(echo "$SYSTEM_HEALTH_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ System health endpoint works (200 OK)${NC}"
else
    echo -e "  ${YELLOW}⚠️  System health endpoint returned (HTTP $HTTP_CODE)${NC}"
    # Not critical, just a warning
fi

# Check 6: API Documentation
echo ""
echo -e "${GREEN}Check 6: API Documentation${NC}"
DOCS_RESPONSE=$(curl -s -w "\n%{http_code}" "${API_URL}/docs" || echo "ERROR")
HTTP_CODE=$(echo "$DOCS_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✅ API documentation accessible (200 OK)${NC}"
else
    echo -e "  ${YELLOW}⚠️  API documentation not accessible (HTTP $HTTP_CODE)${NC}"
    # Not critical
fi

# Summary
echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All critical checks passed! Deployment successful.${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    echo ""
    echo -e "${GREEN}Next steps:${NC}"
    echo "  1. Monitor logs for the next 15 minutes"
    echo "  2. Run full test suite: ./test/restoration_tests.sh"
    echo "  3. Test database persistence: ./test/test_database_persistence.sh"
    echo "  4. Monitor health endpoint: curl ${API_URL}/health/detailed"
    exit 0
else
    echo -e "${RED}❌ Deployment verification failed: $ERRORS error(s)${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    echo ""
    echo -e "${RED}Action required:${NC}"
    echo "  1. Check server logs for errors"
    echo "  2. Verify environment variables are set"
    echo "  3. Verify database connection (if configured)"
    echo "  4. Review deployment checklist: docs/DEPLOYMENT_CHECKLIST.md"
    exit 1
fi

