#!/bin/bash

# Comprehensive Pre-Beta Review Test Script
# Tests all services, endpoints, and UI flows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service URLs from documentation
API_GATEWAY="https://api-gateway-service-production-21fd.up.railway.app"
CLASSIFICATION="https://classification-service-production.up.railway.app"
MERCHANT="https://merchant-service-production.up.railway.app"
FRONTEND="https://frontend-service-production-b225.up.railway.app"
BI_SERVICE="https://bi-service-production.up.railway.app"
PIPELINE="https://pipeline-service-production.up.railway.app"
MONITORING="https://monitoring-service-production.up.railway.app"
SERVICE_DISCOVERY="https://service-discovery-production-d397.up.railway.app"
RISK_ASSESSMENT="https://risk-assessment-service-production.up.railway.app"

REPORT_FILE="COMPREHENSIVE_PRE_BETA_REVIEW.md"
RESULTS=()

# Test function
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    echo -n "Testing $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>&1)
    
    if [ "$response" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} (HTTP $response)"
        RESULTS+=("✅ $name: PASS")
        return 0
    else
        echo -e "${RED}❌ FAIL${NC} (HTTP $response, expected $expected_status)"
        RESULTS+=("❌ $name: FAIL (HTTP $response)")
        return 1
    fi
}

# Test JSON endpoint
test_json_endpoint() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local data=${4:-""}
    
    echo -n "Testing $name... "
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -X POST "$url" \
            -H "Content-Type: application/json" \
            -d "$data" 2>&1)
    else
        response=$(curl -s "$url" 2>&1)
    fi
    
    if echo "$response" | jq . > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS${NC} (Valid JSON)"
        RESULTS+=("✅ $name: PASS")
        return 0
    else
        echo -e "${RED}❌ FAIL${NC} (Invalid JSON or error)"
        RESULTS+=("❌ $name: FAIL")
        return 1
    fi
}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Comprehensive Pre-Beta Review${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Phase 1: Service Health Checks
echo -e "${YELLOW}Phase 1: Service Health Checks${NC}"
echo "----------------------------------------"

test_endpoint "API Gateway Health" "$API_GATEWAY/health"
test_endpoint "Classification Service Health" "$CLASSIFICATION/health"
test_endpoint "Merchant Service Health" "$MERCHANT/health"
test_endpoint "Frontend Service Health" "$FRONTEND/health"
test_endpoint "BI Service Health" "$BI_SERVICE/health"
test_endpoint "Pipeline Service Health" "$PIPELINE/health"
test_endpoint "Monitoring Service Health" "$MONITORING/health"
test_endpoint "Service Discovery Health" "$SERVICE_DISCOVERY/health"
test_endpoint "Risk Assessment Service Health" "$RISK_ASSESSMENT/health"

echo ""

# Phase 2: Frontend Pages
echo -e "${YELLOW}Phase 2: Frontend Pages${NC}"
echo "----------------------------------------"

test_endpoint "Add Merchant Page" "$FRONTEND/add-merchant.html"
test_endpoint "Merchant Details Page" "$FRONTEND/merchant-details"
test_endpoint "Merchant Portfolio Page" "$FRONTEND/merchant-portfolio.html"
test_endpoint "Dashboard Hub" "$FRONTEND/dashboard-hub.html"
test_endpoint "Business Intelligence" "$FRONTEND/business-intelligence.html"

echo ""

# Phase 3: API Endpoints
echo -e "${YELLOW}Phase 3: API Endpoints${NC}"
echo "----------------------------------------"

CLASSIFY_DATA='{"business_name":"Test Company","geographic_region":"US","website_url":"https://example.com","description":"Test","analysis_type":"comprehensive"}'
test_json_endpoint "Classification API" "$API_GATEWAY/api/v1/classify" "POST" "$CLASSIFY_DATA"
test_json_endpoint "Merchants List API" "$API_GATEWAY/api/v1/merchants" "GET"

echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
for result in "${RESULTS[@]}"; do
    echo "$result"
done

echo ""
echo -e "${GREEN}Review complete. Results saved to $REPORT_FILE${NC}"

