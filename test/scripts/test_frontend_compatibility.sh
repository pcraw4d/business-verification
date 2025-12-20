#!/bin/bash

# Test script for Priority 4: Frontend Compatibility
# Tests that all responses include required frontend fields

set -e

API_URL="${API_URL:-https://classification-service-production.up.railway.app}"

echo "========================================"
echo "Frontend Compatibility Test"
echo "========================================"
echo ""
echo "API URL: $API_URL"
echo ""

# Required fields for frontend compatibility
REQUIRED_FIELDS=(
    "request_id"
    "business_name"
    "primary_industry"
    "classification"
    "confidence_score"
    "explanation"
    "status"
    "success"
    "timestamp"
    "metadata"
)

# Required classification fields
REQUIRED_CLASSIFICATION_FIELDS=(
    "industry"
    "mcc_codes"
    "naics_codes"
    "sic_codes"
)

passed=0
failed=0

# Test 1: Success response
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 1: Success Response"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

response=$(curl -s -X POST "${API_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Microsoft Corporation", "description": "Software development"}' \
  --max-time 60)

# Check if response is valid JSON
if ! echo "$response" | python3 -c "import sys, json; json.load(sys.stdin)" 2>/dev/null; then
    echo "❌ TEST FAILED: Invalid JSON response"
    ((failed++))
else
    echo "✅ Response is valid JSON"
    
    # Check required fields
    missing_fields=()
    for field in "${REQUIRED_FIELDS[@]}"; do
        if ! echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert '$field' in d, 'Missing field: $field'" 2>/dev/null; then
            missing_fields+=("$field")
        fi
    done
    
    # Check classification fields
    for field in "${REQUIRED_CLASSIFICATION_FIELDS[@]}"; do
        if ! echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert d.get('classification') and '$field' in d['classification'], 'Missing classification field: $field'" 2>/dev/null; then
            missing_fields+=("classification.$field")
        fi
    done
    
    # Check that arrays are not null
    if echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert d.get('classification', {}).get('mcc_codes') is not None, 'mcc_codes is null'" 2>/dev/null; then
        echo "✅ mcc_codes is not null"
    else
        missing_fields+=("classification.mcc_codes (null)")
    fi
    
    if echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert d.get('classification', {}).get('naics_codes') is not None, 'naics_codes is null'" 2>/dev/null; then
        echo "✅ naics_codes is not null"
    else
        missing_fields+=("classification.naics_codes (null)")
    fi
    
    if echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert d.get('classification', {}).get('sic_codes') is not None, 'sic_codes is null'" 2>/dev/null; then
        echo "✅ sic_codes is not null"
    else
        missing_fields+=("classification.sic_codes (null)")
    fi
    
    # Check that metadata is not null
    if echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert d.get('metadata') is not None, 'metadata is null'" 2>/dev/null; then
        echo "✅ metadata is not null"
    else
        missing_fields+=("metadata (null)")
    fi
    
    if [ ${#missing_fields[@]} -eq 0 ]; then
        echo "✅ TEST PASSED: All required fields present"
        ((passed++))
    else
        echo "❌ TEST FAILED: Missing fields: ${missing_fields[*]}"
        ((failed++))
    fi
fi

echo ""

# Test 2: Error response (invalid request)
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 2: Error Response"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

response=$(curl -s -X POST "${API_URL}/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": ""}' \
  --max-time 60)

# Check if response is valid JSON
if ! echo "$response" | python3 -c "import sys, json; json.load(sys.stdin)" 2>/dev/null; then
    echo "❌ TEST FAILED: Invalid JSON response"
    ((failed++))
else
    echo "✅ Response is valid JSON"
    
    # Check required fields (error responses should also have them)
    missing_fields=()
    for field in "${REQUIRED_FIELDS[@]}"; do
        if ! echo "$response" | python3 -c "import sys, json; d=json.load(sys.stdin); assert '$field' in d, 'Missing field: $field'" 2>/dev/null; then
            missing_fields+=("$field")
        fi
    done
    
    if [ ${#missing_fields[@]} -eq 0 ]; then
        echo "✅ TEST PASSED: All required fields present in error response"
        ((passed++))
    else
        echo "❌ TEST FAILED: Missing fields: ${missing_fields[*]}"
        ((failed++))
    fi
fi

echo ""

echo "========================================"
echo "Test Summary"
echo "========================================"
echo ""
echo "Total Tests: $((passed + failed))"
echo "Passed: $passed"
echo "Failed: $failed"
echo ""

if [ $failed -eq 0 ]; then
    echo "✅ ALL TESTS PASSED"
    echo "Frontend compatibility fix is working correctly!"
    exit 0
else
    echo "❌ SOME TESTS FAILED"
    exit 1
fi

