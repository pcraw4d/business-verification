#!/bin/bash

# KYB Tool Classification API Test Suite
echo "🧪 Testing KYB Classification API"
echo "=================================="

BASE_URL="http://localhost:8080"

# Test 1: Software Company
echo -e "\n1️⃣ Testing Software Company:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "CodeCraft Software", "business_type": "technology", "description": "Custom software development"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

# Test 2: Marketing Agency
echo -e "\n2️⃣ Testing Marketing Agency:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Creative Marketing Solutions", "business_type": "advertising", "description": "Digital marketing and advertising"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

# Test 3: Consulting Firm
echo -e "\n3️⃣ Testing Consulting Firm:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Strategic Consulting Group", "business_type": "consulting", "description": "Management and strategy consulting"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

# Test 4: Restaurant
echo -e "\n4️⃣ Testing Restaurant:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Mama Mias Italian Bistro", "business_type": "restaurant", "description": "Fine dining Italian restaurant"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

# Test 5: Healthcare Provider
echo -e "\n5️⃣ Testing Healthcare Provider:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Downtown Medical Center", "business_type": "healthcare", "description": "Primary care medical practice"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

# Test 6: Batch Classification
echo -e "\n6️⃣ Testing Batch Classification:"
curl -s -X POST $BASE_URL/v1/classify/batch \
  -H "Content-Type: application/json" \
  -d '{
    "businesses": [
      {"business_name": "TechStart Software", "business_type": "technology"},
      {"business_name": "Law Office Associates", "business_type": "legal"},
      {"business_name": "Fresh Market Grocery", "business_type": "retail"}
    ]
  }' | jq '{total_processed: .total_processed, success_count: .success_count, results: [.results[] | {name: .raw_data.request.business_name, code: .primary_classification.industry_code, classification: .primary_classification.industry_name}]}'

# Test 7: Edge Cases
echo -e "\n7️⃣ Testing Edge Cases:"
echo "Empty business name:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "", "business_type": "technology"}' \
  | jq '.error // .primary_classification.industry_name'

echo -e "\nMinimal data:"
curl -s -X POST $BASE_URL/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "XYZ Corp"}' \
  | jq '.primary_classification | {code: .industry_code, name: .industry_name, confidence: .confidence_score}'

echo -e "\n✅ Test Suite Complete!"
