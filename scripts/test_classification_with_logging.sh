#!/bin/bash
# Test script to trigger a classification request and capture enhanced logging
# This script makes a real classification request and monitors Railway logs for enhanced logging indicators

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "🧪 Testing Classification Request with Enhanced Logging"
echo "======================================================"
echo ""

# Load Railway environment variables
if [ -f "railway.env" ]; then
    echo "📋 Loading Railway environment variables from railway.env..."
    set -a
    source railway.env
    set +a
    export SUPABASE_URL
    export SUPABASE_ANON_KEY
    export SUPABASE_SERVICE_ROLE_KEY
    export DATABASE_URL
    echo -e "${GREEN}✅ Environment variables loaded${NC}"
else
    echo -e "${YELLOW}⚠️  No railway.env found${NC}"
    echo "   Using environment variables from current shell"
fi

# Get classification service URL
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

# Try to get from Railway CLI if available
if command -v railway &> /dev/null; then
    RAILWAY_URL=$(railway variables --service classification-service 2>/dev/null | grep "RAILWAY_PUBLIC_DOMAIN" | awk -F'=' '{print $2}' | tr -d ' ' || echo "")
    if [ -n "$RAILWAY_URL" ]; then
        CLASSIFICATION_SERVICE_URL="https://$RAILWAY_URL"
        echo -e "${GREEN}✅ Found classification service: $CLASSIFICATION_SERVICE_URL${NC}"
    fi
fi

echo ""
echo "🔍 Classification Service URL: $CLASSIFICATION_SERVICE_URL"
echo ""

# Verify service is accessible
echo "🔍 Verifying service is accessible..."
if curl -s -f -m 5 "$CLASSIFICATION_SERVICE_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Service is accessible${NC}"
else
    echo -e "${RED}❌ Service is NOT accessible${NC}"
    echo "   URL: $CLASSIFICATION_SERVICE_URL"
    exit 1
fi

# Test business data - using a simpler case that should complete faster
TEST_BUSINESS_NAME="Local Coffee Shop"
TEST_WEBSITE_URL="https://example.com"
TEST_DESCRIPTION="Small local coffee shop and cafe"

echo ""
echo "📝 Test Request Details:"
echo "   Business Name: $TEST_BUSINESS_NAME"
echo "   Website URL: $TEST_WEBSITE_URL"
echo "   Description: $TEST_DESCRIPTION"
echo ""

# Create request payload
REQUEST_PAYLOAD=$(cat <<EOF
{
  "business_name": "$TEST_BUSINESS_NAME",
  "website_url": "$TEST_WEBSITE_URL",
  "description": "$TEST_DESCRIPTION"
}
EOF
)

echo "🚀 Making classification request..."
echo ""

# Start monitoring Railway logs in background
LOG_FILE="/tmp/classification_test_logs_$(date +%Y%m%d_%H%M%S).log"
echo "📊 Monitoring Railway logs (saving to: $LOG_FILE)"
echo "   Note: Logs will be fetched after request completes"

# Make the classification request
echo "📤 Sending POST request to $CLASSIFICATION_SERVICE_URL/v1/classify..."
echo "   (Timeout: 60 seconds)"
echo ""

# Make request with timeout, but capture timeout separately
RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 30 -X POST "$CLASSIFICATION_SERVICE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_PAYLOAD" 2>&1)

CURL_EXIT_CODE=$?
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

# Handle timeout
if [ $CURL_EXIT_CODE -eq 28 ]; then
    HTTP_CODE="TIMEOUT"
    echo -e "${YELLOW}⚠️  Request timed out after 30 seconds${NC}"
    echo "   This is expected for complex website scraping"
    echo "   We'll still check logs for any enhanced logging indicators"
fi

# Wait a moment for logs to flush
sleep 3

# Fetch recent logs from Railway
echo ""
echo "📥 Fetching recent Railway logs..."
railway logs --service classification-service --json > "$LOG_FILE" 2>&1 || railway logs --json > "$LOG_FILE" 2>&1 || echo "⚠️  Could not fetch Railway logs"

echo ""
echo "📊 Response:"
echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Request successful${NC}"
    echo ""
    echo "   Response body (first 500 chars):"
    echo "$RESPONSE_BODY" | head -c 500
    echo ""
else
    echo -e "${RED}❌ Request failed${NC}"
    echo ""
    echo "   Response body:"
    echo "$RESPONSE_BODY"
fi

echo ""
echo "🔍 Analyzing logs for enhanced logging indicators..."
echo ""

# Check for enhanced logging indicators
if [ -f "$LOG_FILE" ]; then
    echo "📋 Enhanced Logging Indicators Found:"
    echo "====================================="
    
    FAST_PATH_COUNT=$(grep -c "\[FAST-PATH\]" "$LOG_FILE" 2>/dev/null || echo "0")
    PARALLEL_COUNT=$(grep -c "\[PARALLEL\]" "$LOG_FILE" 2>/dev/null || echo "0")
    CONTENT_CHECK_COUNT=$(grep -c "\[ContentCheck\]" "$LOG_FILE" 2>/dev/null || echo "0")
    REGULAR_COUNT=$(grep -c "\[REGULAR\]" "$LOG_FILE" 2>/dev/null || echo "0")
    
    # Convert to integers, handling potential empty strings
    FAST_PATH_COUNT=${FAST_PATH_COUNT:-0}
    PARALLEL_COUNT=${PARALLEL_COUNT:-0}
    CONTENT_CHECK_COUNT=${CONTENT_CHECK_COUNT:-0}
    REGULAR_COUNT=${REGULAR_COUNT:-0}
    
    echo ""
    echo "   [FAST-PATH]: $FAST_PATH_COUNT occurrences"
    echo "   [PARALLEL]: $PARALLEL_COUNT occurrences"
    echo "   [ContentCheck]: $CONTENT_CHECK_COUNT occurrences"
    echo "   [REGULAR]: $REGULAR_COUNT occurrences"
    echo ""
    
    TOTAL_ENHANCED=$((FAST_PATH_COUNT + PARALLEL_COUNT + CONTENT_CHECK_COUNT + REGULAR_COUNT))
    
    if [ "$TOTAL_ENHANCED" -gt 0 ]; then
        echo -e "${GREEN}✅ Enhanced logging is working! Found $TOTAL_ENHANCED total indicators${NC}"
        echo ""
        echo "📋 Sample Enhanced Log Entries:"
        echo "================================"
        grep -E "\[FAST-PATH\]|\[PARALLEL\]|\[ContentCheck\]|\[REGULAR\]" "$LOG_FILE" | head -20
    else
        echo -e "${YELLOW}⚠️  No enhanced logging indicators found${NC}"
        echo ""
        echo "   This could mean:"
        echo "   1. Website scraping wasn't triggered (single-page content was sufficient)"
        echo "   2. Request didn't reach the website scraping code path"
        echo "   3. Logs haven't flushed yet"
        echo ""
        echo "   Checking for any website scraping activity..."
        grep -iE "scraping|crawl|website|keyword" "$LOG_FILE" | head -10 || echo "   No website scraping activity found"
    fi
    
    echo ""
    echo "📋 All Classification-Related Logs:"
    echo "===================================="
    grep -iE "classify|classification|POST|business|industry|website" "$LOG_FILE" | head -30 || echo "   No classification-related logs found"
    
    echo ""
    echo "💾 Full log file saved to: $LOG_FILE"
else
    echo -e "${RED}❌ Log file not found: $LOG_FILE${NC}"
fi

echo ""
echo "✅ Test complete"
echo ""

