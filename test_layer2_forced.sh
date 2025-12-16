#!/bin/bash
# Test script to force Layer 2 by using ambiguous business descriptions

CLASSIFICATION_URL="https://classification-service-production.up.railway.app"
REQUEST_ID="test_$(date +%s)"

echo "=========================================="
echo "Layer 2 Forced Test - Ambiguous Cases"
echo "=========================================="
echo ""

# Test case designed to have low Layer 1 confidence
# Using novel/ambiguous terminology that Layer 1 might struggle with
echo "Test: Ambiguous business with novel terminology"
echo "Expected: Layer 1 confidence < 0.80, triggering Layer 2"
echo ""

curl -s --max-time 90 -X POST "$CLASSIFICATION_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "{
    \"business_name\": \"Quantum Analytics Platform\",
    \"description\": \"Leveraging quantum-inspired algorithms for predictive analytics and real-time decision optimization in distributed systems\",
    \"website_url\": \"https://www.ibm.com\"
  }" | python3 -m json.tool 2>/dev/null | grep -E '"method"|"confidence"|"primary_industry"|"request_id"|"status"' | head -10

echo ""
echo "=========================================="
echo "Check logs for Layer 2 activity:"
echo "railway logs -s classification-service | grep -i 'layer 2\|embedding\|phase 3.*confidence.*< 80'"
echo "=========================================="

