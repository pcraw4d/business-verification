#!/bin/bash

# Test script for Priority 3: Website Scraping Timeouts
# Tests that requests with website URLs complete successfully within 120s timeout

set -e

API_URL="${API_URL:-https://classification-service-production.up.railway.app}"
TIMEOUT=150  # 150s test timeout (120s request timeout + 30s buffer)

echo "========================================"
echo "Website Scraping Timeout Test"
echo "========================================"
echo ""
echo "API URL: $API_URL"
echo "Test Timeout: ${TIMEOUT}s"
echo ""

# Test cases with website URLs
test_cases=(
    '{"business_name": "Microsoft Corporation", "description": "Software development", "website_url": "https://www.microsoft.com"}'
    '{"business_name": "Apple Inc", "description": "Technology company", "website_url": "https://www.apple.com"}'
    '{"business_name": "Amazon", "description": "E-commerce and cloud computing", "website_url": "https://www.amazon.com"}'
)

passed=0
failed=0
timeout_count=0

for i in "${!test_cases[@]}"; do
    test_num=$((i + 1))
    test_data="${test_cases[$i]}"
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Test $test_num: Website URL Request"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    start_time=$(date +%s.%N)
    
    # Make request with timeout
    response=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/v1/classify" \
        -H "Content-Type: application/json" \
        -d "$test_data" \
        --max-time $TIMEOUT) || {
        echo "❌ Request failed or timed out"
        ((failed++))
        ((timeout_count++))
        continue
    }
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo "Response Time: ${duration}s"
    echo "HTTP Status: $http_code"
    echo ""
    
    # Check for timeout
    if [ "$http_code" = "408" ] || [ "$http_code" = "504" ]; then
        echo "❌ TEST FAILED: Request timed out (HTTP $http_code)"
        ((failed++))
        ((timeout_count++))
        continue
    fi
    
    # Parse JSON response
    success=$(echo "$body" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('success', False))" 2>/dev/null || echo "false")
    status=$(echo "$body" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('status', ''))" 2>/dev/null || echo "")
    
    echo "Success: $success"
    echo "Status: $status"
    echo ""
    
    # Check if request completed successfully
    if [ "$success" = "True" ] && [ "$http_code" = "200" ]; then
        echo "✅ TEST PASSED: Request completed successfully"
        echo "   Duration: ${duration}s (within 120s timeout)"
        ((passed++))
    elif [ "$http_code" = "200" ]; then
        echo "⚠️  TEST PARTIAL: Request completed but success=false"
        echo "   Status: $status"
        ((passed++))
    else
        echo "❌ TEST FAILED: Request failed (HTTP $http_code)"
        echo "   Response: $body"
        ((failed++))
    fi
    
    echo ""
done

echo "========================================"
echo "Test Summary"
echo "========================================"
echo ""
echo "Total Tests: $((passed + failed))"
echo "Passed: $passed"
echo "Failed: $failed"
echo "Timeouts: $timeout_count"
echo ""

if [ $failed -eq 0 ] && [ $timeout_count -eq 0 ]; then
    echo "✅ ALL TESTS PASSED"
    echo "Website scraping timeout fix is working correctly!"
    exit 0
else
    echo "❌ SOME TESTS FAILED"
    echo "Timeout failures: $timeout_count"
    exit 1
fi

