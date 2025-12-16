#!/bin/bash

# Quick Manual Test for Classification Service
# This is a simple test you can run immediately

set -e

# Get service URL from environment or use default
SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"

echo "🧪 Quick Classification Test"
echo "============================"
echo "Service URL: ${SERVICE_URL}"
echo ""

# Test 1: Simple case (Layer 1)
echo "Test 1: Simple Technology Company (should use Layer 1)"
echo "--------------------------------------------------------"
curl -s -X POST "${SERVICE_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "TechCorp Solutions",
    "description": "Software development and cloud computing services"
  }' | jq -r '{
    industry: .primary_industry // .classification.industry,
    confidence: .confidence_score // .confidence,
    method: .metadata.method // "unknown",
    request_id: .request_id
  }' || echo "Response received (install jq for formatted output)"

echo ""
echo ""

# Test 2: Ambiguous case (should trigger LLM - Layer 3)
echo "Test 2: Ambiguous Multi-Industry Business (should use Layer 3 - LLM)"
echo "----------------------------------------------------------------------"
curl -s -X POST "${SERVICE_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Global Innovations Group",
    "description": "A diversified company providing technology consulting, healthcare data analytics, financial advisory services, and sustainable energy solutions. We operate across multiple sectors with a focus on digital transformation and innovation.",
    "website_url": "https://example.com"
  }' | jq -r '{
    industry: .primary_industry // .classification.industry,
    confidence: .confidence_score // .confidence,
    method: .metadata.method // "unknown",
    has_explanation: (.classification.explanation != null),
    request_id: .request_id
  }' || echo "Response received (install jq for formatted output)"

echo ""
echo ""
echo "✅ Quick test complete!"
echo ""
echo "To see full response, remove the '| jq ...' part from the curl commands"
echo "Or run the full test suite: ./test_e2e_classification.sh"

