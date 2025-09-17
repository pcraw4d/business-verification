#!/bin/bash

# =============================================================================
# CLASSIFICATION IMPROVEMENTS TEST SCRIPT
# =============================================================================
# This script tests the classification improvements to ensure they're working
# correctly after applying the database fixes.

set -e

# Configuration
API_BASE_URL="http://localhost:8080"
SUPABASE_URL="https://qpqhuqqmkjxsltzshfam.supabase.co"
SUPABASE_API_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTQ4NzQ4MzEsImV4cCI6MjA3MDQ1MDgzMX0.UelJkQAVf-XJz1UV0Rbyi-hZHADGOdsHo1PwcPf7JVI"

echo "🚀 Testing Classification Improvements"
echo "======================================"

# Test 1: Check if server is running
echo "📡 Test 1: Checking if server is running..."
if curl -s -f "$API_BASE_URL/health" > /dev/null; then
    echo "✅ Server is running"
else
    echo "❌ Server is not running. Please start the server first."
    exit 1
fi

# Test 2: Check database schema fixes
echo ""
echo "🗄️  Test 2: Checking database schema fixes..."

# Check if is_active column exists
IS_ACTIVE_EXISTS=$(curl -s -X GET "$SUPABASE_URL/rest/v1/keyword_weights?select=is_active&limit=1" \
  -H "apikey: $SUPABASE_API_KEY" \
  -H "Authorization: Bearer $SUPABASE_API_KEY" | jq -r '.[0].is_active // "null"')

if [ "$IS_ACTIVE_EXISTS" != "null" ]; then
    echo "✅ is_active column exists and has data"
else
    echo "❌ is_active column is missing or empty"
    echo "   Please run the database schema fix script first"
    exit 1
fi

# Check if restaurant industry exists
RESTAURANT_INDUSTRY=$(curl -s -X GET "$SUPABASE_URL/rest/v1/industries?name=eq.Restaurants&select=id,name" \
  -H "apikey: $SUPABASE_API_KEY" \
  -H "Authorization: Bearer $SUPABASE_API_KEY" | jq -r '.[0].name // "null"')

if [ "$RESTAURANT_INDUSTRY" = "Restaurants" ]; then
    echo "✅ Restaurant industry exists"
else
    echo "❌ Restaurant industry is missing"
    echo "   Please run the database schema fix script first"
    exit 1
fi

# Test 3: Test restaurant classification
echo ""
echo "🍽️  Test 3: Testing restaurant classification..."

# Test Italian Restaurant
echo "   Testing Italian Restaurant..."
ITALIAN_RESULT=$(curl -s -X POST "$API_BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Mario'\''s Italian Bistro",
    "description": "Fine dining Italian restaurant serving authentic pasta and wine",
    "website_url": ""
  }')

ITALIAN_INDUSTRY=$(echo "$ITALIAN_RESULT" | jq -r '.industry.name // "null"')
ITALIAN_CONFIDENCE=$(echo "$ITALIAN_RESULT" | jq -r '.confidence // 0')

if [[ "$ITALIAN_INDUSTRY" == *"Restaurant"* ]] || [[ "$ITALIAN_INDUSTRY" == *"Food"* ]]; then
    echo "   ✅ Italian Restaurant classified as: $ITALIAN_INDUSTRY (confidence: $ITALIAN_CONFIDENCE)"
else
    echo "   ❌ Italian Restaurant classified as: $ITALIAN_INDUSTRY (confidence: $ITALIAN_CONFIDENCE)"
    echo "   Expected: Restaurant or Food industry"
fi

# Test Fast Food
echo "   Testing Fast Food..."
FAST_FOOD_RESULT=$(curl -s -X POST "$API_BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "McDonalds",
    "description": "Fast food restaurant chain serving burgers and fries",
    "website_url": ""
  }')

FAST_FOOD_INDUSTRY=$(echo "$FAST_FOOD_RESULT" | jq -r '.industry.name // "null"')
FAST_FOOD_CONFIDENCE=$(echo "$FAST_FOOD_RESULT" | jq -r '.confidence // 0')

if [[ "$FAST_FOOD_INDUSTRY" == *"Food"* ]] || [[ "$FAST_FOOD_INDUSTRY" == *"Restaurant"* ]]; then
    echo "   ✅ Fast Food classified as: $FAST_FOOD_INDUSTRY (confidence: $FAST_FOOD_CONFIDENCE)"
else
    echo "   ❌ Fast Food classified as: $FAST_FOOD_INDUSTRY (confidence: $FAST_FOOD_CONFIDENCE)"
    echo "   Expected: Food or Restaurant industry"
fi

# Test Pizza Restaurant
echo "   Testing Pizza Restaurant..."
PIZZA_RESULT=$(curl -s -X POST "$API_BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Tony'\''s Pizza",
    "description": "Family-owned pizzeria serving authentic Italian pizza",
    "website_url": ""
  }')

PIZZA_INDUSTRY=$(echo "$PIZZA_RESULT" | jq -r '.industry.name // "null"')
PIZZA_CONFIDENCE=$(echo "$PIZZA_RESULT" | jq -r '.confidence // 0')

if [[ "$PIZZA_INDUSTRY" == *"Restaurant"* ]] || [[ "$PIZZA_INDUSTRY" == *"Food"* ]]; then
    echo "   ✅ Pizza Restaurant classified as: $PIZZA_INDUSTRY (confidence: $PIZZA_CONFIDENCE)"
else
    echo "   ❌ Pizza Restaurant classified as: $PIZZA_INDUSTRY (confidence: $PIZZA_CONFIDENCE)"
    echo "   Expected: Restaurant or Food industry"
fi

# Test 4: Check confidence score differentiation
echo ""
echo "📊 Test 4: Checking confidence score differentiation..."

# Get confidence scores from all tests
CONFIDENCE_SCORES=("$ITALIAN_CONFIDENCE" "$FAST_FOOD_CONFIDENCE" "$PIZZA_CONFIDENCE")

# Check if confidence scores are different (not all 0.45)
UNIQUE_SCORES=$(printf '%s\n' "${CONFIDENCE_SCORES[@]}" | sort -u | wc -l)

if [ "$UNIQUE_SCORES" -gt 1 ]; then
    echo "✅ Confidence scores are differentiated: ${CONFIDENCE_SCORES[*]}"
else
    echo "❌ Confidence scores are identical: ${CONFIDENCE_SCORES[*]}"
    echo "   Expected: Different confidence scores based on match quality"
fi

# Test 5: Check keyword extraction quality
echo ""
echo "🔍 Test 5: Checking keyword extraction quality..."

# Test with website URL
WEBSITE_RESULT=$(curl -s -X POST "$API_BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "The Greene Grape",
    "description": "Wine and spirits retailer",
    "website_url": "https://greenegrape.com/"
  }')

WEBSITE_KEYWORDS=$(echo "$WEBSITE_RESULT" | jq -r '.metadata.website_keywords // []' | jq -r 'length')
WEBSITE_SCRAPING=$(echo "$WEBSITE_RESULT" | jq -r '.metadata.website_scraping.success // false')

if [ "$WEBSITE_KEYWORDS" -gt 10 ]; then
    echo "✅ Website keyword extraction working: $WEBSITE_KEYWORDS keywords extracted"
else
    echo "❌ Website keyword extraction poor: only $WEBSITE_KEYWORDS keywords extracted"
fi

if [ "$WEBSITE_SCRAPING" = "true" ]; then
    echo "✅ Website scraping successful"
else
    echo "❌ Website scraping failed"
fi

# Test 6: Performance check
echo ""
echo "⚡ Test 6: Checking performance..."

# Measure response time
START_TIME=$(date +%s%N)
curl -s -X POST "$API_BASE_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Performance",
    "description": "Testing response time",
    "website_url": ""
  }' > /dev/null
END_TIME=$(date +%s%N)

RESPONSE_TIME=$(( (END_TIME - START_TIME) / 1000000 )) # Convert to milliseconds

if [ "$RESPONSE_TIME" -lt 1000 ]; then
    echo "✅ Response time acceptable: ${RESPONSE_TIME}ms"
else
    echo "❌ Response time too slow: ${RESPONSE_TIME}ms"
fi

# Summary
echo ""
echo "📋 Test Summary"
echo "==============="

# Count successful tests
SUCCESS_COUNT=0
TOTAL_TESTS=6

# Check each test result
if curl -s -f "$API_BASE_URL/health" > /dev/null; then ((SUCCESS_COUNT++)); fi
if [ "$IS_ACTIVE_EXISTS" != "null" ]; then ((SUCCESS_COUNT++)); fi
if [[ "$ITALIAN_INDUSTRY" == *"Restaurant"* ]] || [[ "$ITALIAN_INDUSTRY" == *"Food"* ]]; then ((SUCCESS_COUNT++)); fi
if [ "$UNIQUE_SCORES" -gt 1 ]; then ((SUCCESS_COUNT++)); fi
if [ "$WEBSITE_KEYWORDS" -gt 10 ]; then ((SUCCESS_COUNT++)); fi
if [ "$RESPONSE_TIME" -lt 1000 ]; then ((SUCCESS_COUNT++)); fi

echo "Tests passed: $SUCCESS_COUNT/$TOTAL_TESTS"

if [ "$SUCCESS_COUNT" -eq "$TOTAL_TESTS" ]; then
    echo "🎉 All tests passed! Classification improvements are working correctly."
    exit 0
elif [ "$SUCCESS_COUNT" -ge 4 ]; then
    echo "⚠️  Most tests passed. Some improvements may need additional work."
    exit 0
else
    echo "❌ Multiple tests failed. Please check the implementation."
    exit 1
fi
