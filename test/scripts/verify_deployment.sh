#!/bin/bash

# Verification script to check if fixes are deployed to Railway production
# Checks for specific code patterns and response structures

set -e

CLASSIFICATION_API_URL="https://classification-service-production.up.railway.app"

echo "🔍 Verifying Railway Production Deployment"
echo "=========================================="
echo ""

# Test 1: Check if service is accessible
echo "Test 1: Service Health Check"
if curl -s -f --max-time 10 "$CLASSIFICATION_API_URL/health" > /dev/null 2>&1; then
    echo "✅ Service is accessible"
else
    echo "❌ Service is not accessible"
    exit 1
fi
echo ""

# Test 2: Make a test request and check response structure
echo "Test 2: Response Structure Verification"
TEST_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company", "description": "Test description", "website_url": "https://example.com"}' \
  --max-time 30 2>&1)

if echo "$TEST_RESPONSE" | grep -q "from_cache"; then
    echo "✅ Response includes 'from_cache' field"
else
    echo "❌ Response missing 'from_cache' field"
    echo "Response preview: ${TEST_RESPONSE:0:200}"
fi

if echo "$TEST_RESPONSE" | grep -q "metadata"; then
    echo "✅ Response includes 'metadata' field"
else
    echo "❌ Response missing 'metadata' field"
fi

if echo "$TEST_RESPONSE" | grep -q "classification"; then
    echo "✅ Response includes 'classification' field"
else
    echo "❌ Response missing 'classification' field"
fi

if echo "$TEST_RESPONSE" | grep -q "primary_industry"; then
    echo "✅ Response includes 'primary_industry' field"
else
    echo "❌ Response missing 'primary_industry' field"
fi
echo ""

# Test 3: Check for cache key prefix in logs (if we can access logs)
echo "Test 3: Cache Key Format Verification"
echo "⚠️  Cannot directly verify cache key format without log access"
echo "   Expected format: 'classification:...' (64 char hex hash)"
echo ""

# Test 4: Check error response structure
echo "Test 4: Error Response Structure Verification"
echo "⚠️  Cannot test error responses without triggering errors"
echo "   Expected: All error responses include required frontend fields"
echo ""

# Test 5: Make duplicate request to test cache
echo "Test 5: Cache Functionality Test"
echo "Making first request..."
FIRST_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Cache Test Company", "description": "Cache test", "website_url": "https://cachetest.com"}' \
  --max-time 30 2>&1)

sleep 2

echo "Making duplicate request..."
SECOND_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Cache Test Company", "description": "Cache test", "website_url": "https://cachetest.com"}' \
  --max-time 30 2>&1)

if echo "$SECOND_RESPONSE" | grep -q '"from_cache":\s*true'; then
    echo "✅ Cache is working - second request hit cache"
else
    echo "⚠️  Cache may not be working - second request did not hit cache"
    echo "   This could be normal if cache TTL expired or cache is disabled"
fi
echo ""

echo "=========================================="
echo "Verification Complete"
echo ""
echo "Note: Full verification requires access to Railway logs"
echo "      to check for specific log messages like:"
echo "      - '✅ [CACHE-SET] Stored in Redis cache'"
echo "      - '⏱️ [TIMEOUT] Calculated adaptive timeout'"
echo "      - Cache keys starting with 'classification:'"

