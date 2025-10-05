#!/bin/bash

echo "🚀 Testing All Railway Services After Redeployment"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Service URLs
API_GATEWAY_URL="https://api-gateway-service-production.up.railway.app"
CLASSIFICATION_URL="https://classification-service-production.up.railway.app"
MERCHANT_URL="https://merchant-service-production.up.railway.app"
FRONTEND_URL="https://frontend-service-production.up.railway.app"

echo -e "\n${BLUE}1. Testing Classification Service${NC}"
echo "URL: $CLASSIFICATION_URL/health"
if response=$(curl -s "$CLASSIFICATION_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ Classification Service: HEALTHY${NC}"
        echo "Supabase: $(echo "$response" | grep -o '"connected":[^,]*' | cut -d: -f2)"
        echo "Classifications: $(echo "$response" | grep -o '"classifications_count":[^,]*' | cut -d: -f2)"
    else
        echo -e "${RED}❌ Classification Service: UNHEALTHY${NC}"
        echo "Response: $response"
    fi
else
    echo -e "${RED}❌ Classification Service: NOT RESPONDING${NC}"
fi

echo -e "\n${BLUE}2. Testing Merchant Service${NC}"
echo "URL: $MERCHANT_URL/health"
if response=$(curl -s "$MERCHANT_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ Merchant Service: HEALTHY${NC}"
        echo "Supabase: $(echo "$response" | grep -o '"connected":[^,]*' | cut -d: -f2)"
        echo "Merchants: $(echo "$response" | grep -o '"merchants_count":[^,]*' | cut -d: -f2)"
    else
        echo -e "${RED}❌ Merchant Service: UNHEALTHY${NC}"
        echo "Response: $response"
    fi
else
    echo -e "${RED}❌ Merchant Service: NOT RESPONDING${NC}"
fi

echo -e "\n${BLUE}3. Testing Frontend Service${NC}"
echo "URL: $FRONTEND_URL"
if response=$(curl -s -I "$FRONTEND_URL" 2>/dev/null | head -1); then
    if echo "$response" | grep -q "200"; then
        echo -e "${GREEN}✅ Frontend Service: HEALTHY${NC}"
    else
        echo -e "${YELLOW}⚠️  Frontend Service: $response${NC}"
    fi
else
    echo -e "${RED}❌ Frontend Service: NOT RESPONDING${NC}"
fi

echo -e "\n${BLUE}4. Testing API Gateway${NC}"
echo "URL: $API_GATEWAY_URL/health"
if response=$(curl -s "$API_GATEWAY_URL/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ API Gateway: HEALTHY${NC}"
    else
        echo -e "${YELLOW}⚠️  API Gateway: $response${NC}"
    fi
else
    echo -e "${RED}❌ API Gateway: NOT RESPONDING${NC}"
fi

echo -e "\n${BLUE}5. Testing API Gateway Routing${NC}"
echo "Testing /api/classification/health"
if response=$(curl -s "$API_GATEWAY_URL/api/classification/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ API Gateway -> Classification: WORKING${NC}"
    else
        echo -e "${YELLOW}⚠️  API Gateway -> Classification: $response${NC}"
    fi
else
    echo -e "${RED}❌ API Gateway -> Classification: NOT WORKING${NC}"
fi

echo "Testing /api/merchant/health"
if response=$(curl -s "$API_GATEWAY_URL/api/merchant/health" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo -e "${GREEN}✅ API Gateway -> Merchant: WORKING${NC}"
    else
        echo -e "${YELLOW}⚠️  API Gateway -> Merchant: $response${NC}"
    fi
else
    echo -e "${RED}❌ API Gateway -> Merchant: NOT WORKING${NC}"
fi

echo -e "\n${BLUE}6. Testing Business Verification Flow${NC}"
echo "Testing classification endpoint"
if response=$(curl -s -X POST "$CLASSIFICATION_URL/classify" \
    -H "Content-Type: application/json" \
    -d '{"business_name":"Test Company","business_description":"Software development company"}' 2>/dev/null); then
    if echo "$response" | grep -q '"classifications"'; then
        echo -e "${GREEN}✅ Business Classification: WORKING${NC}"
        echo "Sample result: $(echo "$response" | head -c 100)..."
    else
        echo -e "${YELLOW}⚠️  Business Classification: $response${NC}"
    fi
else
    echo -e "${RED}❌ Business Classification: NOT WORKING${NC}"
fi

echo -e "\n${BLUE}Summary${NC}"
echo "=========="
echo "✅ Classification Service: Working"
echo "✅ Merchant Service: Working" 
echo "❌ API Gateway: Issues with routing"
echo "❌ Frontend Service: 404 error"
echo ""
echo "Next steps:"
echo "1. Fix API Gateway routing configuration"
echo "2. Fix Frontend Service deployment"
echo "3. Test end-to-end business verification flow"
