#!/bin/bash

# End-to-End Classification Test Suite
# Tests the full 3-layer classification system including LLM (Layer 3)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-http://localhost:8080}"
LLM_SERVICE_URL="${LLM_SERVICE_URL:-}"

echo -e "${BLUE}🧪 End-to-End Classification Test Suite${NC}"
echo "=========================================="
echo -e "Classification Service: ${YELLOW}${CLASSIFICATION_SERVICE_URL}${NC}"
if [ -n "$LLM_SERVICE_URL" ]; then
    echo -e "LLM Service: ${YELLOW}${LLM_SERVICE_URL}${NC}"
else
    echo -e "LLM Service: ${YELLOW}Not configured${NC}"
fi
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to make API calls
test_classification() {
    local test_name="$1"
    local business_name="$2"
    local description="$3"
    local website_url="${4:-}"
    local expected_layer="${5:-}"  # "layer1", "layer2", "layer3", or empty for any
    
    echo -e "\n${BLUE}📋 Test: ${test_name}${NC}"
    echo "  Business: ${business_name}"
    echo "  Description: ${description}"
    if [ -n "$website_url" ]; then
        echo "  Website: ${website_url}"
    fi
    if [ -n "$expected_layer" ]; then
        echo "  Expected Layer: ${expected_layer}"
    fi
    
    # Build request body
    local request_body="{\"business_name\": \"${business_name}\", \"description\": \"${description}\""
    if [ -n "$website_url" ]; then
        request_body="${request_body}, \"website_url\": \"${website_url}\""
    fi
    request_body="${request_body}}"
    
    # Make API call (try /v1/classify first, fallback to /classify)
    local start_time=$(date +%s%N)
    local response=$(curl -s -w "\n%{http_code}" -X POST "${CLASSIFICATION_SERVICE_URL}/v1/classify" \
        -H "Content-Type: application/json" \
        -d "${request_body}" \
        --max-time 120)
    
    # If /v1/classify fails, try /classify
    local http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" = "404" ] || [ "$http_code" = "000" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "${CLASSIFICATION_SERVICE_URL}/classify" \
            -H "Content-Type: application/json" \
            -d "${request_body}" \
            --max-time 120)
    fi
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
    
    # Extract HTTP status code (last line)
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    # Check HTTP status
    if [ "$http_code" != "200" ]; then
        echo -e "  ${RED}❌ FAILED: HTTP ${http_code}${NC}"
        echo "  Response: ${body}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
    
    # Parse response using jq if available, otherwise use grep
    if command -v jq &> /dev/null; then
        local industry=$(echo "$body" | jq -r '.primary_industry // .classification.industry // "N/A"')
        local confidence=$(echo "$body" | jq -r '.confidence_score // .confidence // 0')
        local method=$(echo "$body" | jq -r '.metadata.method // .classification.explanation.method_used // "unknown"')
        local request_id=$(echo "$body" | jq -r '.request_id // "N/A"')
        local has_explanation=$(echo "$body" | jq -r '.classification.explanation != null // false')
        
        echo -e "  ${GREEN}✅ SUCCESS${NC}"
        echo "  Industry: ${industry}"
        echo "  Confidence: ${confidence}"
        echo "  Method: ${method}"
        echo "  Request ID: ${request_id}"
        echo "  Processing Time: ${duration}ms"
        echo "  Has Explanation: ${has_explanation}"
        
        # Check if expected layer was used
        if [ -n "$expected_layer" ]; then
            case "$expected_layer" in
                "layer3"|"llm")
                    if [[ "$method" == *"llm"* ]] || [[ "$method" == *"reasoning"* ]]; then
                        echo -e "  ${GREEN}✅ Layer 3 (LLM) was used as expected${NC}"
                    else
                        echo -e "  ${YELLOW}⚠️  Expected Layer 3 (LLM) but got: ${method}${NC}"
                        echo "  (This might be OK if confidence was high enough for Layer 1/2)"
                    fi
                    ;;
                "layer2"|"embedding")
                    if [[ "$method" == *"embedding"* ]] || [[ "$method" == *"vector"* ]]; then
                        echo -e "  ${GREEN}✅ Layer 2 (Embedding) was used as expected${NC}"
                    else
                        echo -e "  ${YELLOW}⚠️  Expected Layer 2 but got: ${method}${NC}"
                    fi
                    ;;
                "layer1"|"multi")
                    if [[ "$method" == *"multi"* ]] || [[ "$method" == *"keyword"* ]]; then
                        echo -e "  ${GREEN}✅ Layer 1 (Multi-Strategy) was used as expected${NC}"
                    else
                        echo -e "  ${YELLOW}⚠️  Expected Layer 1 but got: ${method}${NC}"
                    fi
                    ;;
            esac
        fi
        
        # Validate response structure
        local has_codes=$(echo "$body" | jq -r '.classification.mcc_codes != null or .classification.naics_codes != null // false')
        if [ "$has_codes" = "true" ]; then
            echo -e "  ${GREEN}✅ Response includes industry codes${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Response missing industry codes${NC}"
        fi
        
    else
        # Fallback if jq is not available
        echo -e "  ${GREEN}✅ SUCCESS (HTTP 200)${NC}"
        echo "  Response length: $(echo "$body" | wc -c) bytes"
        echo "  Processing Time: ${duration}ms"
        if echo "$body" | grep -q "primary_industry\|classification"; then
            echo -e "  ${GREEN}✅ Response structure looks valid${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Response structure may be invalid${NC}"
        fi
    fi
    
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

# Test 1: Simple, clear case (should use Layer 1)
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Suite 1: Layer 1 (Multi-Strategy) Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

test_classification \
    "Clear Technology Company" \
    "TechCorp Solutions" \
    "Software development and cloud computing services" \
    "" \
    "layer1"

test_classification \
    "Restaurant Business" \
    "Mama Mia's Italian Restaurant" \
    "Authentic Italian restaurant serving pasta, pizza, and fine wines" \
    "" \
    "layer1"

# Test 2: Ambiguous/Complex cases (should trigger Layer 3 - LLM)
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Suite 2: Layer 3 (LLM) Tests - Ambiguous Cases${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

test_classification \
    "Ambiguous Multi-Industry Business" \
    "Global Innovations Group" \
    "A diversified company providing technology consulting, healthcare data analytics, financial advisory services, and sustainable energy solutions. We operate across multiple sectors with a focus on digital transformation and innovation." \
    "https://example.com" \
    "layer3"

test_classification \
    "Novel Business Model" \
    "QuantumLeap Dynamics" \
    "We combine quantum computing research with traditional software development to create next-generation solutions for industries that don't yet exist. Our platform enables businesses to prepare for future technological paradigms." \
    "https://example.com" \
    "layer3"

test_classification \
    "Vague Description" \
    "Synergy Partners" \
    "We help businesses grow and succeed through innovative solutions and strategic partnerships." \
    "https://example.com" \
    "layer3"

# Test 3: With Website URL (may trigger Layer 2 or Layer 3)
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Suite 3: Website Content Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

test_classification \
    "E-commerce with Website" \
    "ShopSmart Online" \
    "Online retail platform selling consumer electronics, home goods, and fashion accessories" \
    "https://www.amazon.com" \
    ""

test_classification \
    "Healthcare Provider with Website" \
    "Wellness Medical Center" \
    "Comprehensive medical services including primary care, specialty consultations, and preventive health programs" \
    "https://www.mayoclinic.org" \
    ""

# Test 4: Edge Cases
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Suite 4: Edge Cases${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

test_classification \
    "Minimal Information" \
    "ABC Corp" \
    "Business services" \
    "" \
    ""

test_classification \
    "Long Description" \
    "Comprehensive Solutions Inc" \
    "$(cat <<'EOF'
A full-service technology consulting firm specializing in enterprise software development, 
cloud infrastructure, cybersecurity, data analytics, artificial intelligence, machine learning, 
IoT solutions, blockchain technology, and digital transformation services. We serve clients across 
multiple industries including healthcare, finance, retail, manufacturing, and government sectors. 
Our team of expert engineers and consultants work closely with clients to deliver innovative 
solutions that drive business growth and operational efficiency.
EOF
)" \
    "" \
    ""

# Summary
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi

