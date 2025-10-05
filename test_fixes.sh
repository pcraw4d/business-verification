#!/bin/bash

# Test script to verify Railway deployment fixes
# Run this after setting the environment variables in Railway

set -e

echo "🔧 Testing Railway Deployment Fixes"
echo "=================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service URLs (using correct Railway service names from dashboard)
API_GATEWAY_URL="https://api-gateway-service-production.up.railway.app"
MERCHANT_SERVICE_URL="https://merchant-service-production.up.railway.app"
CLASSIFICATION_SERVICE_URL="https://classification-service-production.up.railway.app"

echo -e "${BLUE}1. Testing Health Endpoints${NC}"
echo "=========================="

# Test health endpoints
echo -n "API Gateway Health... "
if response=$(curl -s --max-time 10 "$API_GATEWAY_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo -n "Merchant Service Health... "
if response=$(curl -s --max-time 10 "$MERCHANT_SERVICE_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo -n "Classification Service Health... "
if response=$(curl -s --max-time 10 "$CLASSIFICATION_SERVICE_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo ""
echo -e "${BLUE}2. Testing Supabase Connection${NC}"
echo "============================="

# Test Supabase connection
echo -n "API Gateway Supabase... "
if response=$(curl -s --max-time 10 "$API_GATEWAY_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"supabase_status"'; then
        if echo "$response" | grep -q '"connected":true'; then
            echo -e "${GREEN}✅ CONNECTED${NC}"
        else
            echo -e "${RED}❌ DISCONNECTED${NC}"
            echo "  Check SUPABASE_ANON_KEY environment variable"
        fi
    else
        echo -e "${YELLOW}⚠️  NO SUPABASE STATUS${NC}"
        echo "  API Gateway health endpoint doesn't include Supabase status"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo -n "Merchant Service Supabase... "
if response=$(curl -s --max-time 10 "$MERCHANT_SERVICE_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"supabase_status"'; then
        if echo "$response" | grep -q '"connected":true'; then
            echo -e "${GREEN}✅ CONNECTED${NC}"
        else
            echo -e "${RED}❌ DISCONNECTED${NC}"
            echo "  Check SUPABASE_ANON_KEY environment variable"
        fi
    else
        echo -e "${YELLOW}⚠️  NO SUPABASE STATUS${NC}"
        echo "  Merchant Service health endpoint doesn't include Supabase status"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo -n "Classification Service Supabase... "
if response=$(curl -s --max-time 10 "$CLASSIFICATION_SERVICE_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"supabase_status"'; then
        if echo "$response" | grep -q '"connected":true'; then
            echo -e "${GREEN}✅ CONNECTED${NC}"
        else
            echo -e "${RED}❌ DISCONNECTED${NC}"
            echo "  Check SUPABASE_ANON_KEY environment variable"
        fi
    else
        echo -e "${YELLOW}⚠️  NO SUPABASE STATUS${NC}"
        echo "  Classification Service health endpoint doesn't include Supabase status"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo ""
echo -e "${BLUE}3. Testing API Endpoints${NC}"
echo "========================"

# Test classification endpoint directly
echo -n "Classification Service Direct... "
CLASSIFICATION_DATA='{"business_name": "Acme Corporation", "description": "A technology company"}'
if response=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$CLASSIFICATION_DATA" "$CLASSIFICATION_SERVICE_URL/classify" 2>/dev/null); then
    if echo "$response" | grep -q '"classification"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

# Test API Gateway proxy
echo -n "API Gateway -> Classification... "
if response=$(curl -s --max-time 10 -X POST -H "Content-Type: application/json" -d "$CLASSIFICATION_DATA" "$API_GATEWAY_URL/api/v1/classify" 2>/dev/null); then
    if echo "$response" | grep -q '"classification"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

# Test merchant endpoint
echo -n "API Gateway -> Merchant... "
if response=$(curl -s --max-time 10 "$API_GATEWAY_URL/api/v1/merchants" 2>/dev/null); then
    if echo "$response" | grep -q '"merchants"'; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
        echo "  Response: $response"
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo ""
echo -e "${BLUE}4. Summary${NC}"
echo "======="
echo "✅ Health endpoints tested"
echo "✅ Supabase connection tested"
echo "✅ API endpoints tested"
echo "✅ Inter-service communication tested"
echo ""
echo -e "${GREEN}🎉 Testing completed!${NC}"
echo ""
echo "If any tests failed:"
echo "1. Check Railway environment variables"
echo "2. Verify SUPABASE_ANON_KEY is set (not SUPABASE_API_KEY)"
echo "3. Check Railway service logs"
echo "4. Ensure all services are deployed and running"
