#!/bin/bash
# Phase 5 Accuracy Validation Script
# Validates classification accuracy with 50-100 test cases

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
ENDPOINT="${API_URL}/v1/classify"

echo "🎯 Phase 5 Accuracy Validation Test"
echo "====================================="
echo ""
echo "📊 Configuration:"
echo "   API URL: $ENDPOINT"
echo ""

# Test cases: business_name, description, expected_industry, expected_mcc (optional)
# Format: "business_name|description|expected_industry|expected_mcc"
declare -a TEST_CASES=(
    # Restaurants & Food Service
    "McDonalds|Fast food restaurant chain|Restaurants|5814"
    "Starbucks|Coffee shop chain|Restaurants|5814"
    "Pizza Hut|Pizza restaurant|Restaurants|5814"
    "Subway|Sandwich shop|Restaurants|5814"
    "KFC|Fried chicken restaurant|Restaurants|5814"
    
    # Technology
    "Microsoft|Software and cloud computing|Technology|5734"
    "Apple|Consumer electronics and software|Technology|5734"
    "Google|Internet search and technology|Technology|5734"
    "Amazon|E-commerce and cloud services|Technology|5734"
    "IBM|Enterprise software and services|Technology|5734"
    
    # Financial Services
    "Bank of America|Commercial banking services|Financial Services|6011"
    "JPMorgan Chase|Investment banking|Financial Services|6011"
    "Wells Fargo|Banking and financial services|Financial Services|6011"
    "Goldman Sachs|Investment banking|Financial Services|6011"
    "Citibank|Banking services|Financial Services|6011"
    
    # Retail
    "Walmart|Retail department store|Retail|5311"
    "Target|Retail department store|Retail|5311"
    "Costco|Wholesale retail|Retail|5311"
    "Home Depot|Home improvement retail|Retail|5211"
    "Best Buy|Electronics retail|Retail|5732"
    
    # Healthcare
    "Mayo Clinic|Hospital and medical center|Healthcare|8062"
    "Cleveland Clinic|Hospital and medical center|Healthcare|8062"
    "Johns Hopkins|Hospital and medical research|Healthcare|8062"
    "CVS Health|Pharmacy and healthcare|Healthcare|5912"
    "Walgreens|Pharmacy chain|Healthcare|5912"
    
    # Automotive
    "Ford Motor Company|Automobile manufacturer|Automotive|5511"
    "General Motors|Automobile manufacturer|Automotive|5511"
    "Tesla|Electric vehicle manufacturer|Automotive|5511"
    "Toyota|Automobile manufacturer|Automotive|5511"
    "BMW|Luxury automobile manufacturer|Automotive|5511"
    
    # Airlines & Travel
    "American Airlines|Airline services|Airlines|4511"
    "Delta Air Lines|Airline services|Airlines|4511"
    "United Airlines|Airline services|Airlines|4511"
    "Marriott|Hotel chain|Hotels|7011"
    "Hilton|Hotel chain|Hotels|7011"
    
    # Energy & Utilities
    "ExxonMobil|Oil and gas|Energy|5542"
    "Chevron|Oil and gas|Energy|5542"
    "Shell|Oil and gas|Energy|5542"
    "Duke Energy|Electric utility|Utilities|4900"
    "Southern Company|Electric utility|Utilities|4900"
    
    # Telecommunications
    "AT&T|Telecommunications|Telecommunications|4814"
    "Verizon|Telecommunications|Telecommunications|4814"
    "T-Mobile|Wireless telecommunications|Telecommunications|4814"
    "Comcast|Cable and internet|Telecommunications|4814"
    "Sprint|Wireless telecommunications|Telecommunications|4814"
    
    # Manufacturing
    "Boeing|Aerospace manufacturer|Manufacturing|3721"
    "Caterpillar|Heavy machinery|Manufacturing|5082"
    "3M|Industrial manufacturing|Manufacturing|5082"
    "General Electric|Industrial manufacturing|Manufacturing|5082"
    "Honeywell|Industrial manufacturing|Manufacturing|5082"
    
    # Entertainment & Media
    "Disney|Entertainment and media|Entertainment|7829"
    "Netflix|Streaming entertainment|Entertainment|7829"
    "Warner Bros|Film production|Entertainment|7829"
    "Sony|Entertainment and electronics|Entertainment|5734"
    "Nintendo|Video game company|Entertainment|5735"
    
    # Education
    "Harvard University|Higher education|Education|8299"
    "MIT|Higher education|Education|8299"
    "Stanford University|Higher education|Education|8299"
    "Khan Academy|Online education|Education|8299"
    "Coursera|Online education platform|Education|8299"
    
    # Insurance
    "State Farm|Insurance services|Insurance|6300"
    "Allstate|Insurance services|Insurance|6300"
    "Progressive|Insurance services|Insurance|6300"
    "Geico|Insurance services|Insurance|6300"
    "USAA|Insurance and banking|Insurance|6300"
    
    # Real Estate
    "Keller Williams|Real estate brokerage|Real Estate|6513"
    "RE/MAX|Real estate brokerage|Real Estate|6513"
    "Century 21|Real estate brokerage|Real Estate|6513"
    "Coldwell Banker|Real estate brokerage|Real Estate|6513"
    "Zillow|Real estate technology|Real Estate|6513"
    
    # Transportation & Logistics
    "FedEx|Package delivery|Transportation|4215"
    "UPS|Package delivery|Transportation|4215"
    "DHL|Package delivery|Transportation|4215"
    "Uber|Ride sharing|Transportation|4121"
    "Lyft|Ride sharing|Transportation|4121"
    
    # Food & Beverage Manufacturing
    "Coca-Cola|Beverage manufacturer|Food & Beverage|5441"
    "PepsiCo|Beverage and snack manufacturer|Food & Beverage|5441"
    "Nestle|Food and beverage|Food & Beverage|5441"
    "Kraft Heinz|Food manufacturer|Food & Beverage|5441"
    "General Mills|Food manufacturer|Food & Beverage|5441"
    
    # Fashion & Apparel
    "Nike|Athletic apparel|Apparel|5651"
    "Adidas|Athletic apparel|Apparel|5651"
    "Zara|Fashion retail|Apparel|5651"
    "H&M|Fashion retail|Apparel|5651"
    "Gap|Apparel retail|Apparel|5651"
    
    # Pharmaceuticals
    "Pfizer|Pharmaceutical company|Pharmaceuticals|5122"
    "Johnson & Johnson|Pharmaceutical company|Pharmaceuticals|5122"
    "Merck|Pharmaceutical company|Pharmaceuticals|5122"
    "Novartis|Pharmaceutical company|Pharmaceuticals|5122"
    "Roche|Pharmaceutical company|Pharmaceuticals|5122"
)

# Counters
total=0
correct=0
incorrect=0
errors=0
layer1_total=0
layer1_correct=0
layer2_total=0
layer2_correct=0
layer3_total=0
layer3_correct=0
high_confidence=0
medium_confidence=0
low_confidence=0
cache_hits=0

# Results file
RESULTS_FILE=$(mktemp)
echo "Test Case|Business Name|Expected|Got|Match|Layer|Confidence|Cache|Time(ms)" > "$RESULTS_FILE"

# Function to test a single classification
test_classification() {
    local test_case="$1"
    IFS='|' read -r business_name description expected_industry expected_mcc <<< "$test_case"
    
    total=$((total + 1))
    
    # Make request
    local start_time=$(date +%s%N)
    local response=$(curl -s --max-time 60 -X POST "$ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "{\"business_name\": \"$business_name\", \"description\": \"$description\"}" 2>&1)
    local end_time=$(date +%s%N)
    local duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    # Parse response
    local got_industry=$(echo "$response" | jq -r '.primary_industry // .classification.industry // "ERROR"' 2>/dev/null)
    local confidence=$(echo "$response" | jq -r '.confidence_score // .classification.confidence // 0' 2>/dev/null)
    local processing_path=$(echo "$response" | jq -r '.processing_path // .classification.explanation.layer_used // "unknown"' 2>/dev/null)
    local from_cache=$(echo "$response" | jq -r '.from_cache // false' 2>/dev/null)
    
    # Check for errors
    if echo "$response" | jq -e '.error' > /dev/null 2>&1 || [ "$got_industry" = "ERROR" ]; then
        errors=$((errors + 1))
        echo "$total|$business_name|$expected_industry|ERROR|NO|unknown|0|false|$duration_ms" >> "$RESULTS_FILE"
        echo -e "${RED}❌ $business_name: ERROR${NC}"
        return 1
    fi
    
    # Normalize for comparison
    got_lower=$(echo "$got_industry" | tr '[:upper:]' '[:lower:]')
    expected_lower=$(echo "$expected_industry" | tr '[:upper:]' '[:lower:]')
    
    # Check match (case-insensitive, partial match)
    match="NO"
    if [[ "$got_lower" == *"$expected_lower"* ]] || [[ "$expected_lower" == *"$got_lower"* ]]; then
        match="YES"
        correct=$((correct + 1))
        status="${GREEN}✅${NC}"
    else
        incorrect=$((incorrect + 1))
        status="${RED}❌${NC}"
    fi
    
    # Track by layer
    if [[ "$processing_path" == *"layer1"* ]]; then
        layer1_total=$((layer1_total + 1))
        if [ "$match" = "YES" ]; then
            layer1_correct=$((layer1_correct + 1))
        fi
    elif [[ "$processing_path" == *"layer2"* ]]; then
        layer2_total=$((layer2_total + 1))
        if [ "$match" = "YES" ]; then
            layer2_correct=$((layer2_correct + 1))
        fi
    elif [[ "$processing_path" == *"layer3"* ]]; then
        layer3_total=$((layer3_total + 1))
        if [ "$match" = "YES" ]; then
            layer3_correct=$((layer3_correct + 1))
        fi
    fi
    
    # Track confidence distribution
    if [ "$(echo "$confidence >= 0.90" | bc -l)" -eq 1 ]; then
        high_confidence=$((high_confidence + 1))
    elif [ "$(echo "$confidence >= 0.70" | bc -l)" -eq 1 ]; then
        medium_confidence=$((medium_confidence + 1))
    else
        low_confidence=$((low_confidence + 1))
    fi
    
    # Track cache hits
    if [ "$from_cache" = "true" ]; then
        cache_hits=$((cache_hits + 1))
    fi
    
    # Write to results file
    echo "$total|$business_name|$expected_industry|$got_industry|$match|$processing_path|$confidence|$from_cache|$duration_ms" >> "$RESULTS_FILE"
    
    # Print result
    echo -e "$status $business_name: Expected '$expected_industry', Got '$got_industry' (${processing_path}, ${confidence})"
    
    return 0
}

# Run all test cases
echo "Running accuracy validation tests..."
echo ""

for test_case in "${TEST_CASES[@]}"; do
    test_classification "$test_case"
    # Small delay to avoid rate limiting
    sleep 0.1
done

echo ""
echo "=========================================="
echo "📊 Accuracy Validation Results"
echo "=========================================="
echo ""

# Calculate percentages
if [ $total -gt 0 ]; then
    accuracy=$(echo "scale=1; $correct * 100 / $total" | bc)
else
    accuracy=$(echo "scale=1; 0" | bc)
fi

echo "Overall Accuracy:"
echo "   Total Tests: $total"
echo "   Correct: $correct"
echo "   Incorrect: $incorrect"
echo "   Errors: $errors"
echo "   Accuracy: ${accuracy}%"
echo ""

# Layer-specific accuracy
echo "Accuracy by Layer:"
if [ $layer1_total -gt 0 ]; then
    layer1_accuracy=$(echo "scale=1; $layer1_correct * 100 / $layer1_total" | bc)
    echo "   Layer 1 (Keyword): $layer1_correct/$layer1_total (${layer1_accuracy}%)"
else
    echo "   Layer 1 (Keyword): 0/0 (N/A)"
fi

if [ $layer2_total -gt 0 ]; then
    layer2_accuracy=$(echo "scale=1; $layer2_correct * 100 / $layer2_total" | bc)
    echo "   Layer 2 (Embedding): $layer2_correct/$layer2_total (${layer2_accuracy}%)"
else
    echo "   Layer 2 (Embedding): 0/0 (N/A)"
fi

if [ $layer3_total -gt 0 ]; then
    layer3_accuracy=$(echo "scale=1; $layer3_correct * 100 / $layer3_total" | bc)
    echo "   Layer 3 (LLM): $layer3_correct/$layer3_total (${layer3_accuracy}%)"
else
    echo "   Layer 3 (LLM): 0/0 (N/A)"
fi
echo ""

# Confidence distribution
echo "Confidence Distribution:"
echo "   High (>= 0.90): $high_confidence"
echo "   Medium (0.70-0.90): $medium_confidence"
echo "   Low (< 0.70): $low_confidence"
echo ""

# Cache statistics
if [ $total -gt 0 ]; then
    cache_hit_rate=$(echo "scale=1; $cache_hits * 100 / $total" | bc)
    echo "Cache Statistics:"
    echo "   Cache Hits: $cache_hits"
    echo "   Cache Hit Rate: ${cache_hit_rate}%"
    echo ""
fi

# Success criteria
echo "Success Criteria:"
TARGET_ACCURACY=85
if [ "$(echo "$accuracy >= $TARGET_ACCURACY" | bc -l)" -eq 1 ]; then
    echo -e "   Overall Accuracy: ${GREEN}✅ PASS${NC} (>= ${TARGET_ACCURACY}%)"
else
    echo -e "   Overall Accuracy: ${RED}❌ FAIL${NC} (< ${TARGET_ACCURACY}%)"
fi

echo ""
echo "Detailed results saved to: $RESULTS_FILE"
echo ""
echo "=========================================="
echo "✅ Accuracy validation completed"
echo "=========================================="

