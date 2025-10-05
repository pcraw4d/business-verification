#!/bin/bash
echo '🧪 Testing API Gateway after manual restart...'
echo ''

# Test health endpoint
echo '1. Testing Health Endpoint:'
curl -s 'https://api-gateway-service-production.up.railway.app/health' | jq .
echo ''

# Test classification endpoint
echo '2. Testing Classification Endpoint:'
curl -s -X POST -H 'Content-Type: application/json' -d '{"business_name": "Test Company"}' 'https://api-gateway-service-production.up.railway.app/api/v1/classify' | jq . | head -10
echo ''

# Test merchant endpoint  
echo '3. Testing Merchant Endpoint:'
curl -s 'https://api-gateway-service-production.up.railway.app/api/v1/merchants' | jq . | head -5
echo ''

echo '✅ Testing complete!'
