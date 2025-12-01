#!/bin/bash

# Simple Streaming Response Test
# Tests the streaming endpoint with a basic request

set -e

# Configuration
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
ENDPOINT="${CLASSIFICATION_SERVICE_URL}/v1/classify"

echo "=========================================="
echo "Testing Streaming Responses"
echo "=========================================="
echo ""
echo "Service URL: ${CLASSIFICATION_SERVICE_URL}"
echo "Endpoint: ${ENDPOINT}?stream=true"
echo ""

# Test request
REQUEST_BODY='{
  "business_name": "Microsoft Corporation",
  "description": "Software development and cloud computing services",
  "website_url": "https://microsoft.com"
}'

echo "Request Body:"
echo "$REQUEST_BODY" | jq '.'
echo ""
echo "=========================================="
echo "Streaming Response (NDJSON):"
echo "=========================================="
echo ""

# Make streaming request
curl -s -X POST "${ENDPOINT}?stream=true" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY" | while IFS= read -r line; do
    if [ -n "$line" ]; then
      # Pretty print each JSON line
      echo "$line" | jq '.'
      echo ""
    fi
  done

echo "=========================================="
echo "Test Complete"
echo "=========================================="

