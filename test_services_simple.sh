#!/bin/bash

# Simple KYB Platform Service Testing Script
# This script tests the deployed services using Railway URLs

set -e

echo "🚀 KYB Platform Service Testing"
echo "=============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to test HTTP endpoint
test_endpoint() {
    local service_name="$1"
    local url="$2"
    local expected_status="${3:-200}"
    
    echo -n "Testing $service_name... "
    
    # Make HTTP request with timeout
    if response=$(curl -s -w "\n%{http_code}" --max-time 10 "$url" 2>/dev/null); then
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" = "$expected_status" ]; then
            echo -e "${GREEN}✅ PASS${NC} (HTTP $http_code)"
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} (HTTP $http_code)"
            echo "  URL: $url"
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Connection failed)"
        echo "  URL: $url"
        return 1
    fi
}

# Function to test POST endpoint
test_post_endpoint() {
    local service_name="$1"
    local url="$2"
    local data="$3"
    local expected_status="${4:-200}"
    
    echo -n "Testing $service_name... "
    
    # Make HTTP POST request with timeout
    if response=$(curl -s -w "\n%{http_code}" --max-time 10 -X POST -H "Content-Type: application/json" -d "$data" "$url" 2>/dev/null); then
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" = "$expected_status" ]; then
            echo -e "${GREEN}✅ PASS${NC} (HTTP $http_code)"
            return 0
        else
            echo -e "${RED}❌ FAIL${NC} (HTTP $http_code)"
            echo "  URL: $url"
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Connection failed)"
        echo "  URL: $url"
        return 1
    fi
}

echo -e "${BLUE}Please provide the Railway URLs for your deployed services:${NC}"
echo "You can find these in your Railway dashboard under each service's 'Deployments' tab"
echo ""

# Prompt for Railway URLs
read -p "API Gateway URL (e.g., https://api-gateway-service-production.up.railway.app): " API_GATEWAY_URL
read -p "Merchant Service URL: " MERCHANT_SERVICE_URL
read -p "Classification Service URL: " CLASSIFICATION_SERVICE_URL
read -p "Pipeline Service URL: " PIPELINE_SERVICE_URL
read -p "Frontend Service URL: " FRONTEND_SERVICE_URL
read -p "Service Discovery URL: " SERVICE_DISCOVERY_URL
read -p "BI Service URL: " BI_SERVICE_URL
read -p "Monitoring Service URL: " MONITORING_SERVICE_URL

# Add https:// protocol if not present
for var in API_GATEWAY_URL MERCHANT_SERVICE_URL CLASSIFICATION_SERVICE_URL PIPELINE_SERVICE_URL FRONTEND_SERVICE_URL SERVICE_DISCOVERY_URL BI_SERVICE_URL MONITORING_SERVICE_URL; do
    eval "url=\$$var"
    if [[ ! "$url" =~ ^https?:// ]]; then
        eval "$var=https://$url"
    fi
done

echo ""
echo -e "${BLUE}Testing Health Endpoints${NC}"
echo "=========================="

# Test health endpoints for all services
test_endpoint "API Gateway Health" "$API_GATEWAY_URL/health"
test_endpoint "Merchant Service Health" "$MERCHANT_SERVICE_URL/health"
test_endpoint "Classification Service Health" "$CLASSIFICATION_SERVICE_URL/health"
test_endpoint "Pipeline Service Health" "$PIPELINE_SERVICE_URL/health"
test_endpoint "Frontend Service Health" "$FRONTEND_SERVICE_URL/health"
test_endpoint "Service Discovery Health" "$SERVICE_DISCOVERY_URL/health"
test_endpoint "BI Service Health" "$BI_SERVICE_URL/health"
test_endpoint "Monitoring Service Health" "$MONITORING_SERVICE_URL/health"

echo ""
echo -e "${BLUE}Testing Core API Endpoints${NC}"
echo "============================="

# Test classification endpoint
CLASSIFICATION_DATA='{"business_name": "Acme Corporation", "description": "A technology company specializing in software development"}'
test_post_endpoint "Classification API" "$CLASSIFICATION_SERVICE_URL/api/v1/classify" "$CLASSIFICATION_DATA"

# Test merchant endpoints
test_endpoint "Merchant List" "$MERCHANT_SERVICE_URL/api/v1/merchants"
test_endpoint "Merchant Analytics" "$MERCHANT_SERVICE_URL/api/v1/merchants/analytics"

# Test API Gateway proxy endpoints
test_endpoint "API Gateway Merchant Proxy" "$API_GATEWAY_URL/api/v1/merchants"

echo ""
echo -e "${BLUE}Testing Business Intelligence Endpoints${NC}"
echo "============================================="

test_endpoint "BI Executive Dashboard" "$BI_SERVICE_URL/dashboard/executive"
test_endpoint "BI KPIs" "$BI_SERVICE_URL/dashboard/kpis"
test_endpoint "BI Reports" "$BI_SERVICE_URL/reports"

echo ""
echo -e "${BLUE}Testing Inter-Service Communication${NC}"
echo "====================================="

# Test API Gateway proxying to other services
echo -n "Testing API Gateway -> Classification Service... "
CLASSIFICATION_RESPONSE=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$CLASSIFICATION_DATA" "$API_GATEWAY_URL/api/v1/classify" 2>/dev/null || echo "ERROR")
if [[ "$CLASSIFICATION_RESPONSE" == *"classification"* ]] || [[ "$CLASSIFICATION_RESPONSE" == *"MCC"* ]] || [[ "$CLASSIFICATION_RESPONSE" == *"NAICS"* ]] || [[ "$CLASSIFICATION_RESPONSE" == *"SIC"* ]]; then
    echo -e "${GREEN}✅ PASS${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
    echo "  Response: $CLASSIFICATION_RESPONSE"
fi

echo ""
echo -e "${GREEN}🎉 Service testing completed!${NC}"
echo ""
echo "If any tests failed, check:"
echo "1. Railway service logs for errors"
echo "2. Environment variables are set correctly"
echo "3. Supabase connection is working"
echo "4. All services are running and healthy"
