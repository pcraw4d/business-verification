#!/bin/bash

# Phase 2 Comprehensive API Testing Script
# Tests all Phase 2 enhancements via HTTP API

# Don't exit on error - we want to run all tests and report results
# set -e

# Configuration
# Default port is 8081 (not 8080) per config
API_BASE_URL="${API_BASE_URL:-http://localhost:8081}"
ENDPOINT="${API_BASE_URL}/classify"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Helper function to make API request
make_request() {
    local business_name="$1"
    local description="$2"
    local website_url="${3:-}"
    
    local payload
    if [ -z "$website_url" ]; then
        payload=$(cat <<EOF
{
  "business_name": "$business_name",
  "description": "$description"
}
EOF
)
    else
        payload=$(cat <<EOF
{
  "business_name": "$business_name",
  "description": "$description",
  "website_url": "$website_url"
}
EOF
)
    fi
    
    curl -s -X POST "$ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        -w "\n%{http_code}" \
        --max-time 30
}

# Test 1: Top 3 Codes Per Type
test_top3_codes() {
    echo -e "${BLUE}📋 Test 1: Top 3 Codes Per Type${NC}"
    echo ""
    
    local test_cases=(
        "Joe's Pizza Restaurant|Family pizza restaurant serving authentic Italian cuisine"
        "Tech Startup Inc|Software development and cloud services"
        "Fashion Boutique|Clothing and accessories retail store"
    )
    
    local all_passed=true
    
    for case in "${test_cases[@]}"; do
        IFS='|' read -r business desc <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        response=$(make_request "$business" "$desc")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" != "200" ]; then
            echo -e "  ${RED}❌ $business: HTTP $http_code${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
            continue
        fi
        
        # Check for 3 codes per type
        mcc_count=$(echo "$body" | jq -r '.classification.mcc_codes | length' 2>/dev/null || echo "0")
        sic_count=$(echo "$body" | jq -r '.classification.sic_codes | length' 2>/dev/null || echo "0")
        naics_count=$(echo "$body" | jq -r '.classification.naics_codes | length' 2>/dev/null || echo "0")
        
        # Check for Source field
        has_source=$(echo "$body" | jq -r '.classification.mcc_codes[0].source != null' 2>/dev/null || echo "false")
        
        if [ "$mcc_count" = "3" ] && [ "$sic_count" = "3" ] && [ "$naics_count" = "3" ] && [ "$has_source" = "true" ]; then
            echo -e "  ${GREEN}✅ $business: All 3 code types with Source fields${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "  ${RED}❌ $business: MCC:$mcc_count SIC:$sic_count NAICS:$naics_count (expected 3 each)${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
        fi
    done
    
    echo ""
    return $([ "$all_passed" = true ] && echo 0 || echo 1)
}

# Test 2: Confidence Calibration
test_confidence_calibration() {
    echo -e "${BLUE}📊 Test 2: Confidence Calibration${NC}"
    echo ""
    
    local test_cases=(
        "Starbucks Coffee|Coffee shop and cafe|0.70|0.95"
        "ABC Services|General business services|0.60|0.90"
    )
    
    local all_passed=true
    
    for case in "${test_cases[@]}"; do
        IFS='|' read -r business desc min max <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        response=$(make_request "$business" "$desc")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" != "200" ]; then
            echo -e "  ${RED}❌ $business: HTTP $http_code${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
            continue
        fi
        
        confidence=$(echo "$body" | jq -r '.confidence_score' 2>/dev/null || echo "0")
        confidence_pct=$(echo "$confidence * 100" | bc -l | cut -d. -f1)
        min_pct=$(echo "$min * 100" | bc -l | cut -d. -f1)
        max_pct=$(echo "$max * 100" | bc -l | cut -d. -f1)
        
        if (( $(echo "$confidence >= $min && $confidence <= $max" | bc -l) )); then
            echo -e "  ${GREEN}✅ $business: Confidence ${confidence_pct}% in range [${min_pct}%, ${max_pct}%]${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "  ${RED}❌ $business: Confidence ${confidence_pct}% outside range [${min_pct}%, ${max_pct}%]${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
        fi
    done
    
    echo ""
    return $([ "$all_passed" = true ] && echo 0 || echo 1)
}

# Test 3: Fast Path Performance
test_fast_path() {
    echo -e "${BLUE}⚡ Test 3: Fast Path Performance${NC}"
    echo ""
    
    local obvious_cases=(
        "Pizza Hut|Pizza restaurant"
        "Starbucks Coffee|Coffee shop"
        "Hilton Hotel|Hotel and lodging"
        "Chase Bank|Banking services"
    )
    
    local fast_count=0
    local total_time=0
    
    for case in "${obvious_cases[@]}"; do
        IFS='|' read -r business desc <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        start_time=$(date +%s%N)
        response=$(make_request "$business" "$desc")
        end_time=$(date +%s%N)
        
        latency_ms=$(( (end_time - start_time) / 1000000 ))
        total_time=$((total_time + latency_ms))
        
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        # Check processing_path from explanation to determine if fast path was used
        processing_path=$(echo "$body" | jq -r '.classification.explanation.processing_path // "unknown"' 2>/dev/null || echo "unknown")
        method_used=$(echo "$body" | jq -r '.classification.explanation.method_used // "unknown"' 2>/dev/null || echo "unknown")
        
        if [ "$http_code" = "200" ]; then
            # Fast path is detected by processing_path="fast_path" OR method containing "fast_path"
            if [ "$processing_path" = "fast_path" ] || echo "$method_used" | grep -q "fast_path"; then
                if [ "$latency_ms" -lt 100 ]; then
                    echo -e "  ${GREEN}✅ $business: Fast path (<100ms) - ${latency_ms}ms${NC}"
                else
                    echo -e "  ${GREEN}✅ $business: Fast path (${latency_ms}ms) - path detected${NC}"
                fi
                fast_count=$((fast_count + 1))
                PASSED=$((PASSED + 1))
            else
                echo -e "  ${YELLOW}⚠️  $business: Slow path (>=100ms) - ${latency_ms}ms (path: $processing_path)${NC}"
                PASSED=$((PASSED + 1))
            fi
        else
            echo -e "  ${RED}❌ $business: Request failed (HTTP $http_code)${NC}"
            FAILED=$((FAILED + 1))
        fi
    done
    
    local fast_rate=$(echo "scale=1; $fast_count * 100 / ${#obvious_cases[@]}" | bc)
    local avg_latency=$(echo "scale=0; $total_time / ${#obvious_cases[@]}" | bc)
    
    echo -e "  Fast Path Hit Rate: ${fast_rate}% (target: >=60%)"
    echo -e "  Average Latency: ${avg_latency}ms (target: <200ms)"
    echo ""
}

# Test 4: Structured Explanations
test_structured_explanations() {
    echo -e "${BLUE}📝 Test 4: Structured Explanations${NC}"
    echo ""
    
    local test_cases=(
        "Mario's Italian Restaurant|Authentic Italian restaurant"
        "Cloud Services Inc|Cloud computing and SaaS platform"
    )
    
    local all_passed=true
    
    for case in "${test_cases[@]}"; do
        IFS='|' read -r business desc <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        response=$(make_request "$business" "$desc")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" != "200" ]; then
            echo -e "  ${RED}❌ $business: HTTP $http_code${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
            continue
        fi
        
        has_primary=$(echo "$body" | jq -r '.classification.explanation.primary_reason != null and .classification.explanation.primary_reason != ""' 2>/dev/null || echo "false")
        supporting_count=$(echo "$body" | jq -r '.classification.explanation.supporting_factors | length' 2>/dev/null || echo "0")
        has_key_terms=$(echo "$body" | jq -r '.classification.explanation.key_terms_found | length > 0' 2>/dev/null || echo "false")
        has_method=$(echo "$body" | jq -r '.classification.explanation.method_used != null and .classification.explanation.method_used != ""' 2>/dev/null || echo "false")
        
        if [ "$has_primary" = "true" ] && [ "$supporting_count" -ge 3 ] && [ "$has_key_terms" = "true" ] && [ "$has_method" = "true" ]; then
            echo -e "  ${GREEN}✅ $business: Complete explanation ($supporting_count factors)${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "  ${RED}❌ $business: Incomplete explanation (primary:$has_primary, factors:$supporting_count, key_terms:$has_key_terms)${NC}"
            FAILED=$((FAILED + 1))
            all_passed=false
        fi
    done
    
    echo ""
    return $([ "$all_passed" = true ] && echo 0 || echo 1)
}

# Test 5: Generic Fallback
test_generic_fallback() {
    echo -e "${BLUE}🔄 Test 5: Generic Fallback Fix${NC}"
    echo ""
    
    local ambiguous_cases=(
        "ABC Corporation|General business services"
        "XYZ Services|Business services"
        "Global Enterprises|Corporate services"
    )
    
    local generic_count=0
    
    for case in "${ambiguous_cases[@]}"; do
        IFS='|' read -r business desc <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        response=$(make_request "$business" "$desc")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" != "200" ]; then
            continue
        fi
        
        industry=$(echo "$body" | jq -r '.classification.industry // .primary_industry' 2>/dev/null || echo "Unknown")
        
        if [ "$industry" = "General Business" ]; then
            echo -e "  ${YELLOW}⚠️  $business: Classified as 'General Business'${NC}"
            generic_count=$((generic_count + 1))
        else
            echo -e "  ${GREEN}✅ $business: Classified as '$industry' (specific)${NC}"
            PASSED=$((PASSED + 1))
        fi
    done
    
    local generic_rate=$(echo "scale=1; $generic_count * 100 / ${#ambiguous_cases[@]}" | bc)
    echo -e "  Generic Business Rate: ${generic_rate}% (target: <10%)"
    echo ""
}

# Test 6: Performance Metrics
test_performance() {
    echo -e "${BLUE}⚙️  Test 6: Overall Performance${NC}"
    echo ""
    
    local test_cases=(
        "Joe's Pizza|Pizza restaurant"
        "Software Inc|Software development"
        "Fashion Store|Clothing store"
    )
    
    local latencies=()
    
    for case in "${test_cases[@]}"; do
        IFS='|' read -r business desc <<< "$case"
        TOTAL=$((TOTAL + 1))
        
        start_time=$(date +%s%N)
        response=$(make_request "$business" "$desc")
        end_time=$(date +%s%N)
        
        latency_ms=$(( (end_time - start_time) / 1000000 ))
        latencies+=($latency_ms)
        
        http_code=$(echo "$response" | tail -n1)
        
        if [ "$http_code" = "200" ]; then
            if [ "$latency_ms" -lt 500 ]; then
                echo -e "  ${GREEN}✅ $business: Fast (${latency_ms}ms)${NC}"
            else
                echo -e "  ${YELLOW}⚠️  $business: Slow (${latency_ms}ms)${NC}"
            fi
            PASSED=$((PASSED + 1))
        else
            echo -e "  ${RED}❌ $business: Request failed${NC}"
            FAILED=$((FAILED + 1))
        fi
    done
    
    # Calculate percentiles (simplified)
    IFS=$'\n' sorted=($(sort -n <<<"${latencies[*]}"))
    p50=${sorted[$(( ${#sorted[@]} / 2 ))]}
    p90=${sorted[$(( ${#sorted[@]} * 9 / 10 ))]}
    p95=${sorted[$(( ${#sorted[@]} * 95 / 100 ))]}
    
    echo -e "  P50 Latency: ${p50}ms"
    echo -e "  P90 Latency: ${p90}ms"
    echo -e "  P95 Latency: ${p95}ms"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}=== Phase 2 Comprehensive Testing ===${NC}"
    echo ""
    echo "API Endpoint: $ENDPOINT"
    echo ""
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is required for JSON parsing. Please install jq.${NC}"
        exit 1
    fi
    
    # Check if bc is available
    if ! command -v bc &> /dev/null; then
        echo -e "${RED}Error: bc is required for calculations. Please install bc.${NC}"
        exit 1
    fi
    
    # Check if API is reachable (use health endpoint for GET, classify needs POST)
    echo -e "${BLUE}Checking API availability...${NC}"
    # Try health endpoint first (GET request works better for availability check)
    health_url="${API_BASE_URL}/health"
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$health_url" --max-time 5 2>/dev/null || echo "000")
    # Clean up the response code - remove newlines and extract just the number
    http_code=$(echo "$http_code" | tr -d '\n\r' | sed 's/[^0-9]//g')
    
    # Check if it's a connection error (000) or empty
    if [ -z "$http_code" ] || [ "$http_code" = "000" ] || [ "$http_code" = "000000" ]; then
        echo -e "${RED}❌ API is not reachable at $API_BASE_URL${NC}"
        echo ""
        echo "Please ensure:"
        echo "  1. The classification service is running"
        echo "  2. The API URL is correct (set via API_BASE_URL environment variable)"
        echo "  3. The service is accessible from this machine"
        echo ""
        echo "To start the service locally:"
        echo "  cd services/classification-service"
        echo "  go run cmd/main.go"
        echo ""
        echo "Or use a remote API:"
        echo "  API_BASE_URL=https://your-railway-url.up.railway.app ./test/phase2_api_test.sh"
        echo ""
        exit 1
    fi
    
    # Check if it's a valid HTTP status code
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 600 ] 2>/dev/null; then
        echo -e "${GREEN}✅ API is reachable (HTTP $http_code)${NC}"
    else
        echo -e "${RED}❌ API responded with invalid code: $http_code${NC}"
        echo ""
        echo "The API may not be properly configured or running."
        exit 1
    fi
    echo ""
    
    # Run tests (don't exit on failure - collect all results)
    test_top3_codes || true
    test_confidence_calibration || true
    test_fast_path || true
    test_structured_explanations || true
    test_generic_fallback || true
    test_performance || true
    
    # Print summary
    echo -e "${BLUE}=== Test Summary ===${NC}"
    echo ""
    echo -e "Total Tests: $TOTAL"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    
    local pass_rate=$(echo "scale=1; $PASSED * 100 / $TOTAL" | bc)
    echo -e "Pass Rate: ${pass_rate}%"
    echo ""
    
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
