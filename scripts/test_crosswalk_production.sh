#!/bin/bash
# Production test script for hybrid crosswalk approach
# Tests crosswalk usage in code generation

set -e

API_URL="${RAILWAY_API_URL:-https://classification-service-production.up.railway.app}"

echo "=========================================="
echo "CROSSWALK PRODUCTION TEST"
echo "=========================================="
echo "API URL: $API_URL"
echo ""

# Test cases with codes that should have crosswalks
test_cases=(
    '{"business_name": "Convenience Store", "website_url": "https://www.7-eleven.com"}'
    '{"business_name": "Software Development", "website_url": "https://www.microsoft.com"}'
    '{"business_name": "Restaurant", "website_url": "https://www.mcdonalds.com"}'
)

success_count=0
total_tests=${#test_cases[@]}

for i in "${!test_cases[@]}"; do
    test_num=$((i + 1))
    echo "Test $test_num/$total_tests:"
    echo "Request: ${test_cases[$i]}"
    echo ""
    
    start_time=$(date +%s%N)
    
    response=$(curl -s -X POST "$API_URL/v1/classify" \
        -H "Content-Type: application/json" \
        -d "${test_cases[$i]}" \
        -w "\nHTTP_STATUS:%{http_code}\nTIME_TOTAL:%{time_total}")
    
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d: -f2)
    time_total=$(echo "$response" | grep "TIME_TOTAL:" | cut -d: -f2)
    json_response=$(echo "$response" | grep -v "HTTP_STATUS:" | grep -v "TIME_TOTAL:")
    
    echo "HTTP Status: $http_status"
    echo "Response Time: ${time_total}s"
    echo ""
    
    if [ "$http_status" = "200" ]; then
        # Extract code counts from classification object
        mcc_count=$(echo "$json_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('classification', {}).get('mcc_codes', [])))" 2>/dev/null || echo "0")
        naics_count=$(echo "$json_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('classification', {}).get('naics_codes', [])))" 2>/dev/null || echo "0")
        sic_count=$(echo "$json_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('classification', {}).get('sic_codes', [])))" 2>/dev/null || echo "0")
        
        echo "📊 Results:"
        echo "   MCC codes: $mcc_count"
        echo "   NAICS codes: $naics_count"
        echo "   SIC codes: $sic_count"
        
        # Check if crosswalks might have been used (multiple code types)
        if [ "$mcc_count" -gt 0 ] && ([ "$naics_count" -gt 0 ] || [ "$sic_count" -gt 0 ]); then
            echo "   ✅ Crosswalks likely used (codes from multiple types)"
            success_count=$((success_count + 1))
        elif [ "$mcc_count" -gt 0 ] || [ "$naics_count" -gt 0 ] || [ "$sic_count" -gt 0 ]; then
            echo "   ⚠️ Only single code type generated"
        else
            echo "   ❌ No codes generated"
        fi
    else
        echo "❌ Request failed with status $http_status"
        echo "Response: $json_response"
    fi
    
    echo ""
    echo "----------------------------------------"
    echo ""
done

echo "=========================================="
echo "SUMMARY"
echo "=========================================="
echo "Tests passed: $success_count/$total_tests"
echo ""

if [ $success_count -eq $total_tests ]; then
    echo "✅ All tests passed"
    exit 0
else
    echo "⚠️ Some tests failed or showed warnings"
    exit 1
fi

