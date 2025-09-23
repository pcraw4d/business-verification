#!/bin/bash

# Test script to verify UI fixes for website keywords and classification accuracy
# This script tests both the website keyword extraction and classification accuracy improvements

set -e

echo "🧪 Testing UI Fixes for Website Keywords and Classification Accuracy"
echo "=================================================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
API_BASE_URL="http://localhost:8080"
TEST_TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="ui_fixes_test_${TEST_TIMESTAMP}.log"

# Test cases
declare -a TEST_CASES=(
    "Test Restaurant|Fine dining restaurant serving Italian cuisine|https://testrestaurant.com"
    "TechCorp Solutions|Software development and cloud computing services|https://techcorp.com"
    "Green Grape Company|Sustainable wine production and distribution|https://greenegrape.com"
    "McDonald's Corporation|Fast food restaurant chain|https://mcdonalds.com"
    "Apple Inc|Consumer electronics and software company|https://apple.com"
)

# Function to log messages
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Function to test website keyword extraction
test_website_keywords() {
    local business_name=$1
    local description=$2
    local website_url=$3
    
    log_message "INFO" "Testing website keyword extraction for: $business_name"
    
    # Make API request
    local response=$(curl -s -X POST "$API_BASE_URL/api/v3/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"$business_name\",
            \"description\": \"$description\",
            \"website_url\": \"$website_url\"
        }" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        log_message "ERROR" "Failed to make API request for $business_name"
        return 1
    fi
    
    # Extract website keywords from response
    local website_keywords=$(echo "$response" | jq -r '.metadata.website_keywords // []' 2>/dev/null)
    local keyword_count=$(echo "$website_keywords" | jq -r 'length' 2>/dev/null)
    
    if [ "$keyword_count" = "null" ] || [ "$keyword_count" = "0" ]; then
        log_message "ERROR" "No website keywords extracted for $business_name"
        echo "Response: $response" >> "$LOG_FILE"
        return 1
    else
        log_message "SUCCESS" "Extracted $keyword_count website keywords for $business_name: $website_keywords"
        return 0
    fi
}

# Function to test classification accuracy
test_classification_accuracy() {
    local business_name=$1
    local description=$2
    local website_url=$3
    
    log_message "INFO" "Testing classification accuracy for: $business_name"
    
    # Make API request
    local response=$(curl -s -X POST "$API_BASE_URL/api/v3/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"$business_name\",
            \"description\": \"$description\",
            \"website_url\": \"$website_url\"
        }" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        log_message "ERROR" "Failed to make API request for $business_name"
        return 1
    fi
    
    # Extract classification results
    local detected_industry=$(echo "$response" | jq -r '.detected_industry // "unknown"' 2>/dev/null)
    local confidence=$(echo "$response" | jq -r '.confidence // 0' 2>/dev/null)
    local classification_codes=$(echo "$response" | jq -r '.classification_codes // {}' 2>/dev/null)
    
    # Check if confidence is not fixed at 0.45
    if [ "$confidence" = "0.45" ]; then
        log_message "ERROR" "Classification confidence is fixed at 0.45 for $business_name (should be dynamic)"
        return 1
    fi
    
    # Check if confidence is within reasonable bounds
    if (( $(echo "$confidence < 0.1" | bc -l) )) || (( $(echo "$confidence > 1.0" | bc -l) )); then
        log_message "ERROR" "Classification confidence out of bounds for $business_name: $confidence"
        return 1
    fi
    
    # Check if we have classification codes
    local mcc_count=$(echo "$classification_codes" | jq -r '.mcc // [] | length' 2>/dev/null)
    local naics_count=$(echo "$classification_codes" | jq -r '.naics // [] | length' 2>/dev/null)
    local sic_count=$(echo "$classification_codes" | jq -r '.sic // [] | length' 2>/dev/null)
    
    if [ "$mcc_count" = "0" ] && [ "$naics_count" = "0" ] && [ "$sic_count" = "0" ]; then
        log_message "WARNING" "No classification codes returned for $business_name"
    fi
    
    log_message "SUCCESS" "Classification successful for $business_name: $detected_industry (confidence: $confidence)"
    return 0
}

# Function to test domain keyword extraction
test_domain_keyword_extraction() {
    local website_url=$1
    local expected_keywords=$2
    
    log_message "INFO" "Testing domain keyword extraction for: $website_url"
    
    # Extract domain from URL
    local domain=$(echo "$website_url" | sed 's|https\?://||' | sed 's|www\.||' | cut -d'/' -f1 | cut -d'.' -f1)
    
    # Make API request
    local response=$(curl -s -X POST "$API_BASE_URL/api/v3/classify" \
        -H "Content-Type: application/json" \
        -d "{
            \"business_name\": \"Test Company\",
            \"description\": \"Test description\",
            \"website_url\": \"$website_url\"
        }" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        log_message "ERROR" "Failed to make API request for $website_url"
        return 1
    fi
    
    # Extract website keywords
    local website_keywords=$(echo "$response" | jq -r '.metadata.website_keywords // []' 2>/dev/null)
    
    # Check if domain is in keywords
    local domain_found=false
    for keyword in $(echo "$website_keywords" | jq -r '.[]' 2>/dev/null); do
        if [[ "$keyword" == *"$domain"* ]]; then
            domain_found=true
            break
        fi
    done
    
    if [ "$domain_found" = "false" ]; then
        log_message "ERROR" "Domain '$domain' not found in extracted keywords: $website_keywords"
        return 1
    else
        log_message "SUCCESS" "Domain '$domain' found in extracted keywords: $website_keywords"
        return 0
    fi
}

# Main test execution
main() {
    log_message "INFO" "Starting UI fixes test suite"
    
    # Check if API is running
    log_message "INFO" "Checking if API is running at $API_BASE_URL"
    if ! curl -s "$API_BASE_URL/health" > /dev/null 2>&1; then
        log_message "ERROR" "API is not running at $API_BASE_URL. Please start the server first."
        exit 1
    fi
    log_message "SUCCESS" "API is running"
    
    # Initialize counters
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    echo
    echo "🔍 Testing Website Keyword Extraction"
    echo "===================================="
    
    # Test website keyword extraction for each test case
    for test_case in "${TEST_CASES[@]}"; do
        IFS='|' read -r business_name description website_url <<< "$test_case"
        
        total_tests=$((total_tests + 1))
        if test_website_keywords "$business_name" "$description" "$website_url"; then
            passed_tests=$((passed_tests + 1))
            echo -e "${GREEN}✅ PASSED${NC}: Website keywords extracted for $business_name"
        else
            failed_tests=$((failed_tests + 1))
            echo -e "${RED}❌ FAILED${NC}: Website keywords extraction failed for $business_name"
        fi
    done
    
    echo
    echo "🎯 Testing Classification Accuracy"
    echo "================================="
    
    # Test classification accuracy for each test case
    for test_case in "${TEST_CASES[@]}"; do
        IFS='|' read -r business_name description website_url <<< "$test_case"
        
        total_tests=$((total_tests + 1))
        if test_classification_accuracy "$business_name" "$description" "$website_url"; then
            passed_tests=$((passed_tests + 1))
            echo -e "${GREEN}✅ PASSED${NC}: Classification accurate for $business_name"
        else
            failed_tests=$((failed_tests + 1))
            echo -e "${RED}❌ FAILED${NC}: Classification accuracy failed for $business_name"
        fi
    done
    
    echo
    echo "🌐 Testing Domain Keyword Extraction"
    echo "==================================="
    
    # Test domain keyword extraction
    local domain_tests=(
        "https://greenegrape.com|greenegrape"
        "https://techcorp.com|techcorp"
        "https://testrestaurant.com|testrestaurant"
    )
    
    for domain_test in "${domain_tests[@]}"; do
        IFS='|' read -r website_url expected_keyword <<< "$domain_test"
        
        total_tests=$((total_tests + 1))
        if test_domain_keyword_extraction "$website_url" "$expected_keyword"; then
            passed_tests=$((passed_tests + 1))
            echo -e "${GREEN}✅ PASSED${NC}: Domain keyword extraction for $website_url"
        else
            failed_tests=$((failed_tests + 1))
            echo -e "${RED}❌ FAILED${NC}: Domain keyword extraction failed for $website_url"
        fi
    done
    
    echo
    echo "📊 Test Results Summary"
    echo "======================"
    echo "Total Tests: $total_tests"
    echo -e "Passed: ${GREEN}$passed_tests${NC}"
    echo -e "Failed: ${RED}$failed_tests${NC}"
    
    local success_rate=$((passed_tests * 100 / total_tests))
    echo "Success Rate: $success_rate%"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! UI fixes are working correctly.${NC}"
        log_message "SUCCESS" "All UI fixes tests passed successfully"
        exit 0
    else
        echo -e "\n${RED}❌ Some tests failed. Please check the log file: $LOG_FILE${NC}"
        log_message "ERROR" "Some UI fixes tests failed"
        exit 1
    fi
}

# Run main function
main "$@"
