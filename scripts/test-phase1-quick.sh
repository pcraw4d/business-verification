#!/bin/bash

# Quick Phase 1 Test - Single website
# Tests one website and shows response

CLASSIFICATION_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "Testing classification service..."
echo "URL: $CLASSIFICATION_URL"
echo ""

# Test with example.com (simple static site)
echo "Test: Example.com (should use SimpleHTTP or BrowserHeaders strategy)"
echo ""

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME:%{time_total}" \
    -X POST "$CLASSIFICATION_URL/v1/classify" \
    -H "Content-Type: application/json" \
    -d '{
        "business_name": "Test Restaurant",
        "website_url": "https://example.com"
    }')

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
TIME=$(echo "$RESPONSE" | grep "TIME:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_CODE:" | grep -v "TIME:")

echo "Response Time: ${TIME}s"
echo "HTTP Status: $HTTP_CODE"
echo ""
echo "Response:"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
echo ""
echo "Note: Check Railway logs for scraping strategy and quality metrics"

