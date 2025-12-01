#!/bin/bash
# Diagnose and Reset ML Service Circuit Breaker
# This script helps diagnose circuit breaker issues and provides options to reset

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🔍 ML Service Circuit Breaker Diagnostic Tool"
echo "=============================================="
echo ""

# Configuration
PYTHON_ML_SERVICE_URL="${PYTHON_ML_SERVICE_URL:-https://python-ml-service-production-a6b8.up.railway.app}"
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "📡 Service URLs:"
echo "   Python ML Service: $PYTHON_ML_SERVICE_URL"
echo "   Classification Service: $CLASSIFICATION_SERVICE_URL"
echo ""

# Step 1: Check Python ML Service Health
echo "Step 1: Checking Python ML Service Health..."
echo "--------------------------------------------"

HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "${PYTHON_ML_SERVICE_URL}/health" 2>/dev/null || echo "ERROR")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ Python ML Service is healthy${NC}"
    if command -v jq &> /dev/null; then
        STATUS=$(echo "$BODY" | jq -r '.status // "unknown"')
        MODELS=$(echo "$BODY" | jq -r '.models_status // "unknown"')
        echo "   Status: $STATUS"
        echo "   Models: $MODELS"
    else
        echo "   Response: $BODY"
    fi
else
    echo -e "${RED}❌ Python ML Service health check failed (HTTP $HTTP_CODE)${NC}"
    echo "   Response: $BODY"
    exit 1
fi

echo ""

# Step 2: Test Python ML Service Classification
echo "Step 2: Testing Python ML Service Classification..."
echo "---------------------------------------------------"

TEST_REQUEST='{"business_name":"Test Technology Company","description":"Software development and consulting services","website_url":"https://example.com","max_results":3}'
CLASSIFY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$TEST_REQUEST" \
    "${PYTHON_ML_SERVICE_URL}/classify-enhanced" 2>/dev/null || echo "ERROR")

CLASSIFY_HTTP=$(echo "$CLASSIFY_RESPONSE" | tail -n1)
CLASSIFY_BODY=$(echo "$CLASSIFY_RESPONSE" | sed '$d')

if [ "$CLASSIFY_HTTP" == "200" ]; then
    echo -e "${GREEN}✅ Python ML Service classification working${NC}"
    if command -v jq &> /dev/null; then
        SUCCESS=$(echo "$CLASSIFY_BODY" | jq -r '.success // false')
        CLASS_COUNT=$(echo "$CLASSIFY_BODY" | jq -r '(.classifications | length) // 0')
        if [ "$SUCCESS" == "true" ] && [ "$CLASS_COUNT" -gt 0 ]; then
            PRIMARY=$(echo "$CLASSIFY_BODY" | jq -r '.classifications[0].label // "none"')
            CONFIDENCE=$(echo "$CLASSIFY_BODY" | jq -r '.classifications[0].confidence // 0')
            echo "   Success: $SUCCESS"
            echo "   Classifications: $CLASS_COUNT"
            echo "   Primary Industry: $PRIMARY"
            echo "   Confidence: $(printf "%.2f" $CONFIDENCE)"
        else
            echo -e "${YELLOW}⚠️  Classification returned but no results${NC}"
        fi
    fi
else
    echo -e "${RED}❌ Python ML Service classification failed (HTTP $CLASSIFY_HTTP)${NC}"
    echo "   Response: $CLASSIFY_BODY"
fi

echo ""

# Step 3: Check Classification Service Health
echo "Step 3: Checking Classification Service Health..."
echo "------------------------------------------------"

CLASS_HEALTH=$(curl -s "${CLASSIFICATION_SERVICE_URL}/health" 2>/dev/null || echo "ERROR")
if [[ "$CLASS_HEALTH" != *"ERROR"* ]]; then
    echo -e "${GREEN}✅ Classification Service is healthy${NC}"
    if command -v jq &> /dev/null; then
        ML_ENABLED=$(echo "$CLASS_HEALTH" | jq -r '.features.ml_enabled // false')
        STATUS=$(echo "$CLASS_HEALTH" | jq -r '.status // "unknown"')
        echo "   Status: $STATUS"
        echo "   ML Enabled: $ML_ENABLED"
    fi
else
    echo -e "${YELLOW}⚠️  Could not check classification service health${NC}"
fi

echo ""

# Step 4: Test Classification Service with ML
echo "Step 4: Testing Classification Service (should use ML)..."
echo "--------------------------------------------------------"

TEST_CLASSIFY='{"business_name":"Acme Technology Corp","description":"Software development and cloud services","website_url":"https://www.acme.com"}'
CLASS_SERVICE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$TEST_CLASSIFY" \
    "${CLASSIFICATION_SERVICE_URL}/v1/classify" 2>/dev/null || echo "ERROR")

CLASS_SERVICE_HTTP=$(echo "$CLASS_SERVICE_RESPONSE" | tail -n1)
CLASS_SERVICE_BODY=$(echo "$CLASS_SERVICE_RESPONSE" | sed '$d')

if [ "$CLASS_SERVICE_HTTP" == "200" ]; then
    echo -e "${GREEN}✅ Classification Service request successful${NC}"
    if command -v jq &> /dev/null; then
        INDUSTRY=$(echo "$CLASS_SERVICE_BODY" | jq -r '.industry_name // "null"')
        CONFIDENCE=$(echo "$CLASS_SERVICE_BODY" | jq -r '.confidence_score // 0')
        METHOD=$(echo "$CLASS_SERVICE_BODY" | jq -r '.classification_method // "unknown"')
        echo "   Industry: $INDUSTRY"
        echo "   Confidence: $(printf "%.2f" $CONFIDENCE)"
        echo "   Method: $METHOD"
        
        if [ "$METHOD" == "ml_distilbart" ] || [ "$METHOD" == "ml" ]; then
            echo -e "   ${GREEN}✅ ML service is being used!${NC}"
        elif [ "$METHOD" == "ml_fallback" ] || [ "$METHOD" == "keyword" ]; then
            echo -e "   ${YELLOW}⚠️  Using fallback method (not ML)${NC}"
            echo "   This suggests the circuit breaker may be open"
        fi
    fi
else
    echo -e "${RED}❌ Classification Service request failed (HTTP $CLASS_SERVICE_HTTP)${NC}"
    echo "   Response: $CLASS_SERVICE_BODY"
fi

echo ""

# Step 5: Recommendations
echo "Step 5: Recommendations"
echo "-----------------------"

if [ "$CLASS_SERVICE_HTTP" == "200" ] && command -v jq &> /dev/null; then
    METHOD=$(echo "$CLASS_SERVICE_BODY" | jq -r '.classification_method // "unknown"')
    if [[ "$METHOD" != "ml"* ]] && [[ "$METHOD" != "ml_distilbart" ]]; then
        echo -e "${YELLOW}⚠️  Circuit Breaker Issue Detected${NC}"
        echo ""
        echo "The classification service is not using the ML service."
        echo "This typically means the circuit breaker is OPEN."
        echo ""
        echo "Possible causes:"
        echo "  1. Circuit breaker opened due to too many failures"
        echo "  2. ML service was slow/unresponsive during previous requests"
        echo "  3. Network issues between services"
        echo ""
        echo "Solutions:"
        echo "  1. Wait for circuit breaker timeout (60s) to expire"
        echo "  2. Redeploy classification service to reset circuit breaker"
        echo "  3. Check Railway logs for ML service errors"
        echo "  4. Verify ML service is accessible from classification service"
    else
        echo -e "${GREEN}✅ ML Service is working correctly!${NC}"
        echo "   The circuit breaker is CLOSED and ML is being used."
    fi
fi

echo ""
echo "💡 Next Steps:"
echo "   1. Check Railway deployment logs for both services"
echo "   2. Monitor circuit breaker state over time"
echo "   3. If circuit breaker is stuck open, consider redeploying classification service"
echo "   4. Run accuracy tests again after verifying ML service is working"
echo ""

