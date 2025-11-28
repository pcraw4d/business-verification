#!/bin/bash

# Script to verify Python ML service integration
# Checks both services and their connection

set -e

echo "🔍 Verifying Python ML Service Integration"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo -e "${RED}❌ Railway CLI is not installed${NC}"
    echo "   Install it from: https://docs.railway.app/develop/cli"
    exit 1
fi

# Check if logged in
if ! railway whoami &> /dev/null; then
    echo -e "${RED}❌ Not logged in to Railway${NC}"
    echo "   Run: railway login"
    exit 1
fi

echo "📋 Checking Python ML Service..."
echo ""

# Check Python ML service URL
PYTHON_ML_URL=$(railway variables --service python-ml-service 2>/dev/null | grep -i "RAILWAY_PUBLIC_DOMAIN\|SERVICE_URL" | head -1 | awk -F'=' '{print $2}' | tr -d ' ' || echo "")

if [ -z "$PYTHON_ML_URL" ]; then
    echo -e "${YELLOW}⚠️  Could not get Python ML service URL from Railway CLI${NC}"
    echo "   Checking service status..."
    PYTHON_ML_URL=$(railway status --service python-ml-service --json 2>/dev/null | jq -r '.service.url // empty' || echo "")
fi

if [ -z "$PYTHON_ML_URL" ]; then
    echo -e "${YELLOW}⚠️  Could not automatically detect Python ML service URL${NC}"
    echo "   Please provide it manually:"
    read -p "Enter Python ML Service URL: " PYTHON_ML_URL
fi

if [ -z "$PYTHON_ML_URL" ]; then
    echo -e "${RED}❌ Python ML service URL is required${NC}"
    exit 1
fi

# Remove trailing slash
PYTHON_ML_URL=$(echo "$PYTHON_ML_URL" | sed 's|/$||')

echo -e "${GREEN}✅ Python ML Service URL: $PYTHON_ML_URL${NC}"
echo ""

# Test Python ML service endpoints
echo "🧪 Testing Python ML Service Endpoints..."
echo ""

# Test /ping
echo -n "  Testing /ping... "
if curl -s -f -m 5 "$PYTHON_ML_URL/ping" > /dev/null 2>&1; then
    PING_RESPONSE=$(curl -s "$PYTHON_ML_URL/ping")
    if echo "$PING_RESPONSE" | grep -q "ok\|running"; then
        echo -e "${GREEN}✅ OK${NC}"
    else
        echo -e "${RED}❌ Unexpected response: $PING_RESPONSE${NC}"
    fi
else
    echo -e "${RED}❌ Failed (connection error or timeout)${NC}"
fi

# Test /health
echo -n "  Testing /health... "
if curl -s -f -m 5 "$PYTHON_ML_URL/health" > /dev/null 2>&1; then
    HEALTH_RESPONSE=$(curl -s "$PYTHON_ML_URL/health")
    if echo "$HEALTH_RESPONSE" | grep -q "healthy\|status"; then
        echo -e "${GREEN}✅ OK${NC}"
        echo "    Response: $HEALTH_RESPONSE"
    else
        echo -e "${RED}❌ Unexpected response: $HEALTH_RESPONSE${NC}"
    fi
else
    echo -e "${RED}❌ Failed (connection error or timeout)${NC}"
fi

echo ""
echo "📋 Checking Classification Service Configuration..."
echo ""

# Check if PYTHON_ML_SERVICE_URL is set in classification service
CLASSIFICATION_PYTHON_ML_URL=$(railway variables --service classification-service 2>/dev/null | grep "PYTHON_ML_SERVICE_URL" | awk -F'=' '{print $2}' | tr -d ' ' || echo "")

if [ -z "$CLASSIFICATION_PYTHON_ML_URL" ]; then
    echo -e "${RED}❌ PYTHON_ML_SERVICE_URL is NOT set in classification-service${NC}"
    echo ""
    echo "🔧 Setting PYTHON_ML_SERVICE_URL..."
    railway variables set "PYTHON_ML_SERVICE_URL=$PYTHON_ML_URL" --service classification-service
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Successfully set PYTHON_ML_SERVICE_URL=$PYTHON_ML_URL${NC}"
        echo ""
        echo "⏳ Classification service will auto-redeploy..."
        echo "   Wait 1-2 minutes, then check logs for:"
        echo "   - '🐍 Initializing Python ML Service'"
        echo "   - '✅ Python ML Service initialized successfully'"
    else
        echo -e "${RED}❌ Failed to set environment variable${NC}"
        echo ""
        echo "💡 Manual setup:"
        echo "   1. Go to Railway Dashboard → classification-service → Variables"
        echo "   2. Add: PYTHON_ML_SERVICE_URL = $PYTHON_ML_URL"
        exit 1
    fi
else
    # Remove trailing slash for comparison
    CLASSIFICATION_PYTHON_ML_URL=$(echo "$CLASSIFICATION_PYTHON_ML_URL" | sed 's|/$||')
    PYTHON_ML_URL_NO_SLASH=$(echo "$PYTHON_ML_URL" | sed 's|/$||')
    
    if [ "$CLASSIFICATION_PYTHON_ML_URL" = "$PYTHON_ML_URL_NO_SLASH" ]; then
        echo -e "${GREEN}✅ PYTHON_ML_SERVICE_URL is set correctly${NC}"
        echo "   Value: $CLASSIFICATION_PYTHON_ML_URL"
    else
        echo -e "${YELLOW}⚠️  PYTHON_ML_SERVICE_URL is set but doesn't match${NC}"
        echo "   Current: $CLASSIFICATION_PYTHON_ML_URL"
        echo "   Expected: $PYTHON_ML_URL_NO_SLASH"
        echo ""
        read -p "Update to correct URL? (y/n): " UPDATE
        if [ "$UPDATE" = "y" ]; then
            railway variables set "PYTHON_ML_SERVICE_URL=$PYTHON_ML_URL" --service classification-service
            echo -e "${GREEN}✅ Updated${NC}"
        fi
    fi
fi

echo ""
echo "📝 Next Steps:"
echo "   1. Wait for classification service to redeploy (1-2 minutes)"
echo "   2. Check classification service logs for Python ML initialization"
echo "   3. Test enhanced classification with a request that includes website URL"
echo ""
echo "✅ Verification complete!"

