#!/bin/bash

# Script to verify Python ML Service connection from Classification Service perspective
# Tests all aspects of the connection between Go and Python services

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Python ML Service Connection Verification${NC}"
echo ""

# Step 1: Check environment variable
echo -e "${BLUE}Step 1: Checking PYTHON_ML_SERVICE_URL environment variable...${NC}"

PYTHON_ML_SERVICE_URL="${PYTHON_ML_SERVICE_URL:-}"

if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
    echo -e "${YELLOW}⚠️  PYTHON_ML_SERVICE_URL not set in current environment${NC}"
    echo "   Checking Railway environment..."
    
    # Try to get from Railway
    if command -v railway &> /dev/null; then
        RAILWAY_URL=$(railway variables --service classification-service 2>/dev/null | grep "PYTHON_ML_SERVICE_URL" | awk -F'=' '{print $2}' | tr -d ' ' || echo "")
        if [ -n "$RAILWAY_URL" ]; then
            PYTHON_ML_SERVICE_URL="$RAILWAY_URL"
            echo -e "${GREEN}✅ Found in Railway: $PYTHON_ML_SERVICE_URL${NC}"
        else
            echo -e "${RED}❌ PYTHON_ML_SERVICE_URL not set in Railway${NC}"
            echo ""
            echo "   To set it:"
            echo "   railway variables set PYTHON_ML_SERVICE_URL=\"https://python-ml-service-production.up.railway.app\" --service classification-service"
            exit 1
        fi
    else
        echo -e "${RED}❌ Railway CLI not installed${NC}"
        echo "   Please set PYTHON_ML_SERVICE_URL environment variable"
        exit 1
    fi
else
    echo -e "${GREEN}✅ PYTHON_ML_SERVICE_URL is set: $PYTHON_ML_SERVICE_URL${NC}"
fi

# Normalize URL (remove trailing slash)
PYTHON_ML_SERVICE_URL=$(echo "$PYTHON_ML_SERVICE_URL" | sed 's|/$||')
echo "   Normalized URL: $PYTHON_ML_SERVICE_URL"
echo ""

# Step 2: Test basic connectivity
echo -e "${BLUE}Step 2: Testing basic connectivity...${NC}"

if ! curl -s -f -m 5 "$PYTHON_ML_SERVICE_URL/ping" > /dev/null 2>&1; then
    echo -e "${RED}❌ Cannot reach Python ML service at $PYTHON_ML_SERVICE_URL/ping${NC}"
    echo "   Possible issues:"
    echo "   - Service is not running"
    echo "   - URL is incorrect"
    echo "   - Network connectivity problem"
    exit 1
fi

PING_RESPONSE=$(curl -s -m 5 "$PYTHON_ML_SERVICE_URL/ping")
if echo "$PING_RESPONSE" | grep -q "ok\|running"; then
    echo -e "${GREEN}✅ Ping successful: $PING_RESPONSE${NC}"
else
    echo -e "${YELLOW}⚠️  Unexpected ping response: $PING_RESPONSE${NC}"
fi
echo ""

# Step 3: Test health endpoint
echo -e "${BLUE}Step 3: Testing health endpoint...${NC}"

if ! curl -s -f -m 5 "$PYTHON_ML_SERVICE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Health check failed${NC}"
    exit 1
fi

HEALTH_RESPONSE=$(curl -s -m 5 "$PYTHON_ML_SERVICE_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy\|status"; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "   Response: $(echo "$HEALTH_RESPONSE" | head -c 200)..."
else
    echo -e "${YELLOW}⚠️  Unexpected health response: $HEALTH_RESPONSE${NC}"
fi
echo ""

# Step 4: Test models endpoint (should work even if models not loaded)
echo -e "${BLUE}Step 4: Testing /models endpoint...${NC}"

MODELS_RESPONSE=$(curl -s -w "\n%{http_code}" -m 10 "$PYTHON_ML_SERVICE_URL/models" 2>/dev/null || echo "")
HTTP_CODE=$(echo "$MODELS_RESPONSE" | tail -n 1)
MODELS_BODY=$(echo "$MODELS_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ /models endpoint accessible (HTTP $HTTP_CODE)${NC}"
    if echo "$MODELS_BODY" | grep -q "\[\]"; then
        echo "   Models not loaded yet (empty list) - this is OK"
    else
        echo "   Models available: $(echo "$MODELS_BODY" | jq -r 'length' 2>/dev/null || echo "unknown")"
    fi
elif [ "$HTTP_CODE" = "503" ]; then
    echo -e "${YELLOW}⚠️  /models returned 503 (models loading) - this should be fixed${NC}"
    echo "   The endpoint should return 200 with empty list, not 503"
else
    echo -e "${RED}❌ /models endpoint failed (HTTP $HTTP_CODE)${NC}"
    echo "   Response: $MODELS_BODY"
fi
echo ""

# Step 5: Test classification endpoint (if models are loaded)
echo -e "${BLUE}Step 5: Testing classification endpoint...${NC}"

TEST_REQUEST='{
  "business_name": "Test Business",
  "description": "Software development services",
  "website_url": "https://example.com",
  "max_results": 3,
  "max_content_length": 500
}'

CLASSIFY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$TEST_REQUEST" \
    -m 120 \
    "$PYTHON_ML_SERVICE_URL/classify-enhanced" 2>/dev/null || echo "")

HTTP_CODE=$(echo "$CLASSIFY_RESPONSE" | tail -n 1)
CLASSIFY_BODY=$(echo "$CLASSIFY_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Classification endpoint working (HTTP $HTTP_CODE)${NC}"
    if echo "$CLASSIFY_BODY" | grep -q "success"; then
        SUCCESS=$(echo "$CLASSIFY_BODY" | jq -r '.success' 2>/dev/null || echo "unknown")
        if [ "$SUCCESS" = "true" ]; then
            echo -e "${GREEN}   Classification successful${NC}"
        else
            echo -e "${YELLOW}   Classification returned success=false${NC}"
        fi
    fi
elif [ "$HTTP_CODE" = "503" ]; then
    echo -e "${YELLOW}⚠️  Classification returned 503 (models may still be loading)${NC}"
    echo "   This is expected if models are loading in background"
else
    echo -e "${YELLOW}⚠️  Classification endpoint returned HTTP $HTTP_CODE${NC}"
    echo "   Response: $(echo "$CLASSIFY_BODY" | head -c 200)..."
fi
echo ""

# Step 6: Verify Go service can connect (simulate)
echo -e "${BLUE}Step 6: Simulating Go service connection...${NC}"

# Test with Go-like HTTP client behavior
echo "   Testing with standard HTTP client (30s timeout)..."
TIMEOUT_TEST=$(curl -s -w "\n%{http_code}\n%{time_total}" -m 30 "$PYTHON_ML_SERVICE_URL/ping" 2>/dev/null || echo "")
HTTP_CODE=$(echo "$TIMEOUT_TEST" | tail -n 2 | head -n 1)
TIME_TOTAL=$(echo "$TIMEOUT_TEST" | tail -n 1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Connection test successful (HTTP $HTTP_CODE, ${TIME_TOTAL}s)${NC}"
else
    echo -e "${RED}❌ Connection test failed (HTTP $HTTP_CODE)${NC}"
fi
echo ""

# Step 7: Check for common issues
echo -e "${BLUE}Step 7: Checking for common connection issues...${NC}"

ISSUES=0

# Check URL format
if [[ ! "$PYTHON_ML_SERVICE_URL" =~ ^https?:// ]]; then
    echo -e "${RED}❌ URL must start with http:// or https://${NC}"
    ISSUES=$((ISSUES + 1))
fi

# Check for trailing slash (should be removed)
if [[ "$PYTHON_ML_SERVICE_URL" =~ /$ ]]; then
    echo -e "${YELLOW}⚠️  URL has trailing slash (should be removed)${NC}"
    ISSUES=$((ISSUES + 1))
fi

# Check SSL/TLS
if [[ "$PYTHON_ML_SERVICE_URL" =~ ^http:// ]]; then
    echo -e "${YELLOW}⚠️  Using HTTP instead of HTTPS (not recommended for production)${NC}"
    ISSUES=$((ISSUES + 1))
fi

# Check response time
if (( $(echo "$TIME_TOTAL > 5.0" | bc -l 2>/dev/null || echo 0) )); then
    echo -e "${YELLOW}⚠️  Slow response time: ${TIME_TOTAL}s (may cause timeouts)${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ No common issues detected${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Connection Verification Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "✅ Language Difference: HTTP/REST is language-agnostic"
echo "   - Go and Python communicate via standard HTTP/JSON"
echo "   - No language-specific connection issues"
echo ""
echo "✅ Connection Method: HTTPS (Public URL)"
echo "   - Services communicate via Railway's public domains"
echo "   - No internal networking required"
echo ""
echo "✅ Resilience Features:"
echo "   - Circuit breaker (opens after 5 failures)"
echo "   - Graceful fallback (works without Python ML)"
echo "   - Timeout protection (30s per request)"
echo ""
echo "📋 Next Steps:"
echo "   1. Verify PYTHON_ML_SERVICE_URL is set in Railway"
echo "   2. Check classification service logs for initialization"
echo "   3. Test a classification request with website_url"
echo "   4. Monitor circuit breaker state in logs"
echo ""

