#!/bin/bash
# Test script for Layer 2 (Embedding) Integration

CLASSIFICATION_URL="https://classification-service-production.up.railway.app"

echo "=========================================="
echo "Layer 2 Integration Test"
echo "=========================================="
echo ""

# Test 1: High confidence (should use Layer 1 only)
echo "Test 1: High confidence case (should use Layer 1)"
echo "Expected: Layer 1 only, no Layer 2"
curl -s --max-time 30 -X POST "$CLASSIFICATION_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Joe'\''s Pizza Restaurant",
    "description": "Italian restaurant serving pizza and pasta",
    "website_url": "https://joespizza.com"
  }' | python3 -m json.tool 2>/dev/null | grep -E '"method"|"confidence"|"primary_industry"' | head -5
echo ""

# Test 2: Low confidence (should trigger Layer 2)
echo "Test 2: Low confidence case (should trigger Layer 2)"
echo "Expected: Layer 1 confidence < 0.80, then Layer 2"
curl -s --max-time 60 -X POST "$CLASSIFICATION_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "CloudOps Solutions",
    "description": "Container orchestration and microservices architecture consulting",
    "website_url": "https://example.com"
  }' | python3 -m json.tool 2>/dev/null | grep -E '"method"|"confidence"|"primary_industry"|"status"' | head -10
echo ""

echo "=========================================="
echo "Check Railway logs for Layer 2 activity:"
echo "railway logs -s classification-service | grep -i 'layer\|embedding\|phase 3'"
echo "=========================================="

