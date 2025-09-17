#!/bin/bash

# =============================================================================
# Restaurant Classification API Testing Script
# Task 1.3: Test Restaurant Classification
# =============================================================================

# This script tests the restaurant classification API endpoints
# to verify that the restaurant industry data is working correctly

set -e

echo "🚀 Starting Task 1.3: Test Restaurant Classification"
echo "============================================================================="

# Configuration
API_BASE_URL="http://localhost:8080"
API_ENDPOINT="/v1/classify"
CONTENT_TYPE="Content-Type: application/json"

# Check if server is running
echo "🔍 Checking if API server is running..."
if ! curl -s --connect-timeout 5 "$API_BASE_URL/health" > /dev/null 2>&1; then
    echo "❌ API server is not running at $API_BASE_URL"
    echo "   Please start the server first:"
    echo "   go run cmd/server/main.go"
    echo ""
    read -p "Press Enter after starting the server..."
fi

echo "✅ API server is running"

# Function to test classification
test_classification() {
    local test_name="$1"
    local business_name="$2"
    local description="$3"
    local website_url="$4"
    local expected_industry="$5"
    local min_confidence="$6"
    
    echo ""
    echo "🧪 Testing: $test_name"
    echo "   Business: $business_name"
    echo "   Description: $description"
    echo "   Expected Industry: $expected_industry"
    echo "   Min Confidence: $min_confidence"
    
    # Create JSON payload
    local json_payload=$(cat <<EOF
{
    "business_name": "$business_name",
    "description": "$description",
    "website_url": "$website_url"
}
EOF
)
    
    # Make API request
    local response=$(curl -s -X POST "$API_BASE_URL$API_ENDPOINT" \
        -H "$CONTENT_TYPE" \
        -d "$json_payload" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        echo "   ❌ API request failed"
        return 1
    fi
    
    # Parse response
    local actual_industry=$(echo "$response" | jq -r '.industry.name // "null"' 2>/dev/null)
    local confidence=$(echo "$response" | jq -r '.confidence // "null"' 2>/dev/null)
    local keywords=$(echo "$response" | jq -r '.keywords // []' 2>/dev/null)
    local error=$(echo "$response" | jq -r '.error // "null"' 2>/dev/null)
    
    # Check for errors
    if [ "$error" != "null" ] && [ "$error" != "" ]; then
        echo "   ❌ API Error: $error"
        return 1
    fi
    
    # Validate response
    local test_passed=true
    
    if [ "$actual_industry" = "null" ] || [ "$actual_industry" = "" ]; then
        echo "   ❌ No industry returned"
        test_passed=false
    elif [ "$actual_industry" != "$expected_industry" ]; then
        echo "   ⚠️  Industry mismatch: expected '$expected_industry', got '$actual_industry'"
        test_passed=false
    else
        echo "   ✅ Industry correct: $actual_industry"
    fi
    
    if [ "$confidence" = "null" ] || [ "$confidence" = "" ]; then
        echo "   ❌ No confidence score returned"
        test_passed=false
    else
        local confidence_num=$(echo "$confidence" | sed 's/[^0-9.]//g')
        if (( $(echo "$confidence_num >= $min_confidence" | bc -l) )); then
            echo "   ✅ Confidence score: $confidence (>= $min_confidence)"
        else
            echo "   ❌ Confidence too low: $confidence (< $min_confidence)"
            test_passed=false
        fi
    fi
    
    if [ "$keywords" != "[]" ] && [ "$keywords" != "null" ]; then
        local keyword_count=$(echo "$keywords" | jq 'length' 2>/dev/null)
        echo "   ✅ Keywords extracted: $keyword_count keywords"
        echo "   📝 Keywords: $keywords"
    else
        echo "   ⚠️  No keywords extracted"
    fi
    
    if [ "$test_passed" = true ]; then
        echo "   🎉 TEST PASSED"
        return 0
    else
        echo "   💥 TEST FAILED"
        return 1
    fi
}

# Function to test error handling
test_error_handling() {
    local test_name="$1"
    local json_payload="$2"
    
    echo ""
    echo "🧪 Testing Error Handling: $test_name"
    
    local response=$(curl -s -X POST "$API_BASE_URL$API_ENDPOINT" \
        -H "$CONTENT_TYPE" \
        -d "$json_payload" 2>/dev/null)
    
    local error=$(echo "$response" | jq -r '.error // "null"' 2>/dev/null)
    
    if [ "$error" != "null" ] && [ "$error" != "" ]; then
        echo "   ✅ Error handled correctly: $error"
        return 0
    else
        echo "   ❌ Error not handled properly"
        return 1
    fi
}

echo ""
echo "============================================================================="
echo "RESTAURANT CLASSIFICATION API TESTS"
echo "============================================================================="

# Test 1: Basic restaurant classification
test_classification \
    "Italian Restaurant" \
    "Mario's Italian Bistro" \
    "Fine dining Italian restaurant serving authentic pasta and wine" \
    "" \
    "Restaurants" \
    "0.75"

# Test 2: Fast food classification
test_classification \
    "Fast Food Chain" \
    "McDonalds" \
    "Fast food restaurant chain serving burgers and fries" \
    "" \
    "Fast Food" \
    "0.80"

# Test 3: Fine dining classification
test_classification \
    "Fine Dining Restaurant" \
    "Le Bernardin" \
    "Upscale fine dining restaurant with wine pairing and tasting menu" \
    "" \
    "Fine Dining" \
    "0.85"

# Test 4: Casual dining classification
test_classification \
    "Casual Dining" \
    "Olive Garden" \
    "Family-friendly casual dining restaurant with Italian cuisine" \
    "" \
    "Casual Dining" \
    "0.75"

# Test 5: Quick service classification
test_classification \
    "Quick Service" \
    "Chipotle" \
    "Fast casual restaurant with customizable burritos and bowls" \
    "" \
    "Quick Service" \
    "0.80"

# Test 6: Cafe classification
test_classification \
    "Coffee Shop" \
    "Starbucks" \
    "Coffee shop serving specialty coffee drinks and light food" \
    "" \
    "Cafes & Coffee Shops" \
    "0.70"

# Test 7: Bar classification
test_classification \
    "Sports Bar" \
    "Buffalo Wild Wings" \
    "Sports bar and restaurant serving wings, beer, and cocktails" \
    "" \
    "Bars & Pubs" \
    "0.75"

# Test 8: Brewery classification
test_classification \
    "Craft Brewery" \
    "Dogfish Head Brewery" \
    "Craft brewery with tasting room and beer production" \
    "" \
    "Breweries" \
    "0.80"

# Test 9: Winery classification
test_classification \
    "Winery" \
    "Napa Valley Winery" \
    "Winery with wine tasting room and vineyard tours" \
    "" \
    "Wineries" \
    "0.80"

# Test 10: Food truck classification
test_classification \
    "Food Truck" \
    "Taco Truck" \
    "Mobile food truck serving authentic Mexican tacos" \
    "" \
    "Food Trucks" \
    "0.75"

# Test 11: Catering classification
test_classification \
    "Catering Service" \
    "Event Catering Co" \
    "Full-service catering company for weddings and corporate events" \
    "" \
    "Catering" \
    "0.70"

# Test 12: General food & beverage
test_classification \
    "Food Service" \
    "Food Service Inc" \
    "General food and beverage service company" \
    "" \
    "Food & Beverage" \
    "0.70"

echo ""
echo "============================================================================="
echo "ERROR HANDLING TESTS"
echo "============================================================================="

# Test 13: Empty request
test_error_handling \
    "Empty Request" \
    '{}'

# Test 14: Invalid JSON
test_error_handling \
    "Invalid JSON" \
    '{"business_name": "Test", "description": "Test", "invalid": }'

# Test 15: Missing required fields
test_error_handling \
    "Missing Business Name" \
    '{"description": "Test description"}'

echo ""
echo "============================================================================="
echo "PERFORMANCE TESTS"
echo "============================================================================="

# Test 16: Performance test
echo "🧪 Testing API Performance"
echo "   Running 10 classification requests..."

start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s -X POST "$API_BASE_URL$API_ENDPOINT" \
        -H "$CONTENT_TYPE" \
        -d '{"business_name": "Test Restaurant '${i}'", "description": "Test restaurant for performance testing", "website_url": ""}' \
        > /dev/null 2>&1
done
end_time=$(date +%s.%N)

duration=$(echo "$end_time - $start_time" | bc)
avg_duration=$(echo "scale=3; $duration / 10" | bc)

echo "   ✅ Performance test completed"
echo "   📊 Total time: ${duration}s"
echo "   📊 Average time per request: ${avg_duration}s"

if (( $(echo "$avg_duration < 1.0" | bc -l) )); then
    echo "   🎉 Performance test PASSED (< 1.0s per request)"
else
    echo "   ⚠️  Performance test WARNING (> 1.0s per request)"
fi

echo ""
echo "============================================================================="
echo "TASK 1.3 TESTING COMPLETED"
echo "============================================================================="
echo "✅ Restaurant classification API testing completed"
echo "📊 Tests performed: 16 total tests"
echo "🎯 Expected results:"
echo "   - Restaurant businesses classified correctly"
echo "   - Confidence scores > 75%"
echo "   - Business-relevant keywords extracted"
echo "   - No database errors"
echo "   - Performance < 1.0s per request"
echo ""
echo "🚀 Next Steps:"
echo "   1. Review test results above"
echo "   2. Fix any failing tests"
echo "   3. Proceed to Phase 2: Algorithm Improvements"
echo ""
echo "============================================================================="
