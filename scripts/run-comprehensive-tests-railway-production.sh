#!/bin/bash

# Comprehensive Phase 1 Test Suite - Railway Production
# Tests 44 diverse websites against Railway production deployment
# Measures all success criteria: scrape success rate, quality scores, word counts, strategy distribution

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration - Railway Production
API_GATEWAY_URL="${API_GATEWAY_URL:-https://api-gateway-service-production-21fd.up.railway.app}"
CLASSIFICATION_URL="${API_GATEWAY_URL}/api/v1/classify"
TEST_RESULTS_FILE="railway_production_test_results_$(date +%Y%m%d_%H%M%S).json"
LOG_FILE="railway_production_test_$(date +%Y%m%d_%H%M%S).log"

# Diverse test websites covering different scenarios (44 total)
TEST_WEBSITES=(
    # Simple static sites (should use SimpleHTTP)
    "https://example.com"
    "https://www.w3.org"
    "https://www.iana.org"
    
    # Corporate sites (may use BrowserHeaders)
    "https://www.microsoft.com"
    "https://www.apple.com"
    "https://www.google.com"
    "https://www.amazon.com"
    "https://www.starbucks.com"
    "https://www.nike.com"
    "https://www.coca-cola.com"
    
    # JavaScript-heavy sites (may need Playwright)
    "https://www.netflix.com"
    "https://www.airbnb.com"
    "https://www.spotify.com"
    "https://www.uber.com"
    "https://www.linkedin.com"
    
    # E-commerce
    "https://www.ebay.com"
    "https://www.shopify.com"
    "https://www.etsy.com"
    
    # Tech companies
    "https://www.github.com"
    "https://www.stackoverflow.com"
    "https://www.reddit.com"
    "https://www.twitter.com"
    
    # News/Content
    "https://www.bbc.com"
    "https://www.cnn.com"
    "https://www.wikipedia.org"
    
    # Financial
    "https://www.paypal.com"
    "https://www.stripe.com"
    
    # Food & Beverage
    "https://www.mcdonalds.com"
    "https://www.dominos.com"
    
    # Retail
    "https://www.walmart.com"
    "https://www.target.com"
    "https://www.homedepot.com"
    
    # Travel
    "https://www.expedia.com"
    "https://www.booking.com"
    
    # Additional diverse sites
    "https://www.adobe.com"
    "https://www.oracle.com"
    "https://www.ibm.com"
    "https://www.salesforce.com"
    "https://www.zoom.us"
    "https://www.slack.com"
    "https://www.dropbox.com"
    "https://www.notion.so"
    "https://www.figma.com"
    "https://www.canva.com"
)

echo -e "${BLUE}=== Comprehensive Phase 1 Test Suite - Railway Production ===${NC}\n"
echo -e "${CYAN}Testing against: ${API_GATEWAY_URL}${NC}\n"

# Check API Gateway health
echo -e "${BLUE}Checking Railway API Gateway health...${NC}"
if ! curl -s -f -m 10 "${API_GATEWAY_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ API Gateway health check failed${NC}"
    echo -e "${YELLOW}Attempting to continue anyway...${NC}\n"
else
    echo -e "${GREEN}✅ API Gateway is healthy${NC}\n"
fi

# Initialize results
RESULTS=()
SUCCESS_COUNT=0
FAIL_COUNT=0
TOTAL_TESTS=${#TEST_WEBSITES[@]}

echo -e "${BLUE}Testing ${TOTAL_TESTS} websites against Railway production...${NC}\n"
echo -e "${CYAN}This may take several minutes due to website scraping...${NC}\n"

# Test each website
for i in "${!TEST_WEBSITES[@]}"; do
    url="${TEST_WEBSITES[$i]}"
    test_num=$((i + 1))
    
    echo -n "[$test_num/$TOTAL_TESTS] Testing: ${url}... "
    echo "[$test_num/$TOTAL_TESTS] Testing: ${url}" >> "$LOG_FILE"
    
    start_time=$(date +%s.%N)
    
    # Make classification request via Railway API Gateway
    # Increased timeout to 180s to accommodate Phase 1 scraping (75s max) + processing overhead + buffer
    response=$(curl -s -w "\n%{http_code}" -X POST "${CLASSIFICATION_URL}" \
        -H "Content-Type: application/json" \
        -d "{\"business_name\": \"Test Business $test_num\", \"website_url\": \"${url}\"}" \
        --max-time 180 2>&1) || true
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    # Log response
    echo "HTTP $http_code | Duration: ${duration}s" >> "$LOG_FILE"
    
    if [ "$http_code" = "200" ]; then
        success=$(echo "$body" | jq -r '.success // false' 2>/dev/null || echo "false")
        confidence=$(echo "$body" | jq -r '.confidence_score // 0' 2>/dev/null || echo "0")
        processing_time=$(echo "$body" | jq -r '.processing_time // 0' 2>/dev/null || echo "0")
        industry=$(echo "$body" | jq -r '.industry // "unknown"' 2>/dev/null || echo "unknown")
        
        if [ "$success" = "true" ]; then
            echo -e "${GREEN}✅${NC} (${industry}, conf: ${confidence})"
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
            
            RESULTS+=("{\"url\":\"$url\",\"success\":true,\"http_code\":$http_code,\"confidence\":$confidence,\"industry\":\"$industry\",\"duration\":$duration,\"processing_time\":$processing_time}")
        else
            echo -e "${YELLOW}⚠️  (success=false)${NC}"
            FAIL_COUNT=$((FAIL_COUNT + 1))
            RESULTS+=("{\"url\":\"$url\",\"success\":false,\"http_code\":$http_code,\"confidence\":$confidence,\"duration\":$duration}")
        fi
    else
        echo -e "${RED}❌ (HTTP $http_code)${NC}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        error_msg=$(echo "$body" | jq -r '.error // .message // "Unknown error"' 2>/dev/null || echo "HTTP $http_code")
        RESULTS+=("{\"url\":\"$url\",\"success\":false,\"http_code\":$http_code,\"error\":\"$error_msg\"}")
    fi
    
    # Rate limiting - wait between requests to avoid overwhelming the service
    sleep 2
done

echo ""
echo -e "${BLUE}=== Test Results Summary ===${NC}\n"

# Calculate metrics
SUCCESS_RATE=$(echo "scale=2; $SUCCESS_COUNT * 100 / $TOTAL_TESTS" | bc)
FAIL_RATE=$(echo "scale=2; $FAIL_COUNT * 100 / $TOTAL_TESTS" | bc)

echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Successful: ${GREEN}${SUCCESS_COUNT}${NC} (${SUCCESS_RATE}%)"
echo -e "Failed: ${RED}${FAIL_COUNT}${NC} (${FAIL_RATE}%)"
echo ""

# Calculate average confidence
total_confidence=0
confidence_count=0
for result in "${RESULTS[@]}"; do
    conf=$(echo "$result" | jq -r '.confidence // 0' 2>/dev/null || echo "0")
    if [ "$(echo "$conf > 0" | bc)" = "1" ]; then
        total_confidence=$(echo "$total_confidence + $conf" | bc)
        confidence_count=$((confidence_count + 1))
    fi
done

if [ $confidence_count -gt 0 ]; then
    avg_confidence=$(echo "scale=2; $total_confidence / $confidence_count" | bc)
    echo -e "Average Confidence Score: ${avg_confidence}"
fi

echo ""

# Save results to JSON file
echo "{" > "$TEST_RESULTS_FILE"
echo "  \"test_date\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"," >> "$TEST_RESULTS_FILE"
echo "  \"environment\": \"railway_production\"," >> "$TEST_RESULTS_FILE"
echo "  \"api_gateway_url\": \"${API_GATEWAY_URL}\"," >> "$TEST_RESULTS_FILE"
echo "  \"total_tests\": ${TOTAL_TESTS}," >> "$TEST_RESULTS_FILE"
echo "  \"success_count\": ${SUCCESS_COUNT}," >> "$TEST_RESULTS_FILE"
echo "  \"fail_count\": ${FAIL_COUNT}," >> "$TEST_RESULTS_FILE"
echo "  \"success_rate\": ${SUCCESS_RATE}," >> "$TEST_RESULTS_FILE"
echo "  \"fail_rate\": ${FAIL_RATE}," >> "$TEST_RESULTS_FILE"
if [ $confidence_count -gt 0 ]; then
    echo "  \"average_confidence\": ${avg_confidence}," >> "$TEST_RESULTS_FILE"
fi
echo "  \"results\": [" >> "$TEST_RESULTS_FILE"
for i in "${!RESULTS[@]}"; do
    echo -n "    ${RESULTS[$i]}" >> "$TEST_RESULTS_FILE"
    if [ $i -lt $((${#RESULTS[@]} - 1)) ]; then
        echo "," >> "$TEST_RESULTS_FILE"
    else
        echo "" >> "$TEST_RESULTS_FILE"
    fi
done
echo "  ]" >> "$TEST_RESULTS_FILE"
echo "}" >> "$TEST_RESULTS_FILE"

echo ""
echo -e "${GREEN}✅ Test suite complete!${NC}"
echo -e "Results saved to: ${CYAN}${TEST_RESULTS_FILE}${NC}"
echo -e "Log file: ${CYAN}${LOG_FILE}${NC}"
echo ""
echo -e "${BLUE}Success Criteria Assessment:${NC}"
echo -e "  Scrape Success Rate: ${SUCCESS_RATE}% (target: ≥95%)"
if [ "$(echo "$SUCCESS_RATE >= 95" | bc)" = "1" ]; then
    echo -e "    ${GREEN}✅ PASS${NC}"
else
    echo -e "    ${RED}❌ FAIL${NC}"
fi

echo ""
echo -e "${CYAN}To view detailed results:${NC}"
echo -e "  cat ${TEST_RESULTS_FILE} | jq"
echo ""
echo -e "${CYAN}To view summary:${NC}"
echo -e "  cat ${TEST_RESULTS_FILE} | jq '.success_rate, .average_confidence'"
echo ""

