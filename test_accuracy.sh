#!/bin/bash
# Phase 4 Accuracy Testing - 50+ Test Cases
# Tests Layer 1, 2, and 3 classification accuracy

BASE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "=============================================="
echo "Phase 4 Accuracy Testing - $(date)"
echo "Service: $BASE_URL"
echo "=============================================="
echo ""

# Counters
total=0
correct=0
layer1_total=0
layer1_correct=0
layer2_total=0
layer2_correct=0
layer3_total=0
layer3_correct=0

# Test function
test_classification() {
    local name="$1"
    local description="$2"
    local expected="$3"
    local category="$4"
    
    total=$((total + 1))
    
    result=$(curl -s --max-time 60 -X POST "$BASE_URL/v1/classify" \
      -H "Content-Type: application/json" \
      -d "{\"business_name\": \"$name\", \"description\": \"$description\"}" 2>/dev/null)
    
    industry=$(echo "$result" | jq -r '.primary_industry // .classification.industry // "ERROR"' 2>/dev/null)
    confidence=$(echo "$result" | jq -r '.confidence_score // .classification.explanation.confidence_factors.overall_confidence // 0' 2>/dev/null)
    method=$(echo "$result" | jq -r '.classification.explanation.method_used // "unknown"' 2>/dev/null)
    llm_status=$(echo "$result" | jq -r '.llm_status // "none"' 2>/dev/null)
    
    # Normalize for comparison (case-insensitive, partial match)
    industry_lower=$(echo "$industry" | tr '[:upper:]' '[:lower:]')
    expected_lower=$(echo "$expected" | tr '[:upper:]' '[:lower:]')
    
    # Check if expected is contained in industry or vice versa
    if [[ "$industry_lower" == *"$expected_lower"* ]] || [[ "$expected_lower" == *"$industry_lower"* ]]; then
        status="✅"
        correct=$((correct + 1))
        
        # Track by layer
        if [[ "$llm_status" == "processing" ]]; then
            layer3_total=$((layer3_total + 1))
            layer3_correct=$((layer3_correct + 1))
        elif [[ "$method" == *"layer2"* ]] || [[ "$method" == *"embedding"* ]]; then
            layer2_total=$((layer2_total + 1))
            layer2_correct=$((layer2_correct + 1))
        else
            layer1_total=$((layer1_total + 1))
            layer1_correct=$((layer1_correct + 1))
        fi
    else
        status="❌"
        
        # Track by layer
        if [[ "$llm_status" == "processing" ]]; then
            layer3_total=$((layer3_total + 1))
        elif [[ "$method" == *"layer2"* ]] || [[ "$method" == *"embedding"* ]]; then
            layer2_total=$((layer2_total + 1))
        else
            layer1_total=$((layer1_total + 1))
        fi
    fi
    
    printf "%s [%s] %s\n" "$status" "$category" "$name"
    printf "   Expected: %-25s Got: %-25s (%.0f%% conf)\n" "$expected" "$industry" "$(echo "$confidence * 100" | bc 2>/dev/null || echo 0)"
}

echo "=== CATEGORY 1: RESTAURANTS & FOOD SERVICE (10 cases) ==="
test_classification "Texas Roadhouse" "Casual dining steakhouse restaurant chain" "Restaurant" "RESTAURANT"
test_classification "Chipotle Mexican Grill" "Fast casual Mexican restaurant serving burritos and tacos" "Restaurant" "RESTAURANT"
test_classification "Starbucks Coffee" "Coffee shop and cafe serving espresso drinks and pastries" "Coffee" "RESTAURANT"
test_classification "Domino's Pizza" "Pizza delivery and carryout restaurant" "Restaurant" "RESTAURANT"
test_classification "The Cheesecake Factory" "Upscale casual dining restaurant" "Restaurant" "RESTAURANT"
test_classification "Five Guys Burgers" "Fast casual hamburger restaurant" "Restaurant" "RESTAURANT"
test_classification "Panda Express" "Chinese fast food restaurant chain" "Restaurant" "RESTAURANT"
test_classification "Olive Garden" "Italian casual dining restaurant" "Restaurant" "RESTAURANT"
test_classification "Sweetgreen" "Fast casual salad restaurant" "Restaurant" "RESTAURANT"
test_classification "Joe's Crab Shack" "Seafood restaurant specializing in crab and shrimp" "Restaurant" "RESTAURANT"

echo ""
echo "=== CATEGORY 2: RETAIL & SHOPPING (10 cases) ==="
test_classification "Walmart" "Discount department store and grocery retailer" "Retail" "RETAIL"
test_classification "Target" "General merchandise and grocery retailer" "Retail" "RETAIL"
test_classification "Home Depot" "Home improvement and construction materials retailer" "Hardware" "RETAIL"
test_classification "Best Buy" "Consumer electronics retailer" "Electronics" "RETAIL"
test_classification "Nordstrom" "Upscale fashion department store" "Clothing" "RETAIL"
test_classification "Costco Wholesale" "Membership warehouse club retailer" "Warehouse" "RETAIL"
test_classification "Petco" "Pet supplies and services retailer" "Pet" "RETAIL"
test_classification "AutoZone" "Automotive parts and accessories retailer" "Auto" "RETAIL"
test_classification "Lowe's" "Home improvement and appliance retailer" "Hardware" "RETAIL"
test_classification "Dick's Sporting Goods" "Sporting goods and outdoor equipment retailer" "Sporting" "RETAIL"

echo ""
echo "=== CATEGORY 3: PROFESSIONAL SERVICES (10 cases) ==="
test_classification "Deloitte" "Professional services firm providing audit, consulting, and advisory" "Consulting" "PROFESSIONAL"
test_classification "Jones Day Law Firm" "International law firm providing legal services" "Legal" "PROFESSIONAL"
test_classification "H&R Block" "Tax preparation and financial services" "Tax" "PROFESSIONAL"
test_classification "KPMG" "Accounting and professional services firm" "Accounting" "PROFESSIONAL"
test_classification "McKinsey & Company" "Management consulting firm" "Consulting" "PROFESSIONAL"
test_classification "Merrill Lynch" "Wealth management and financial advisory" "Financial" "PROFESSIONAL"
test_classification "Cushman & Wakefield" "Commercial real estate services" "Real Estate" "PROFESSIONAL"
test_classification "Kelly Services" "Staffing and workforce solutions" "Staffing" "PROFESSIONAL"
test_classification "ADP" "Payroll and human resources services" "Payroll" "PROFESSIONAL"
test_classification "Marsh McLennan" "Insurance brokerage and risk management" "Insurance" "PROFESSIONAL"

echo ""
echo "=== CATEGORY 4: TECHNOLOGY & SOFTWARE (10 cases) ==="
test_classification "Microsoft" "Software development and cloud computing services" "Technology" "TECH"
test_classification "Salesforce" "Cloud-based CRM and enterprise software" "Software" "TECH"
test_classification "Adobe" "Creative software and digital marketing solutions" "Software" "TECH"
test_classification "Oracle" "Enterprise software and database solutions" "Software" "TECH"
test_classification "Intuit" "Financial software including QuickBooks and TurboTax" "Software" "TECH"
test_classification "Shopify" "E-commerce platform for online stores" "E-commerce" "TECH"
test_classification "Zoom Video" "Video conferencing and communications platform" "Technology" "TECH"
test_classification "Slack Technologies" "Business communication and collaboration platform" "Technology" "TECH"
test_classification "DocuSign" "Electronic signature and agreement cloud" "Software" "TECH"
test_classification "Twilio" "Cloud communications platform APIs" "Technology" "TECH"

echo ""
echo "=== CATEGORY 5: HEALTHCARE & MEDICAL (10 cases) ==="
test_classification "CVS Health" "Pharmacy and healthcare services" "Pharmacy" "HEALTHCARE"
test_classification "Kaiser Permanente" "Integrated healthcare and hospital system" "Healthcare" "HEALTHCARE"
test_classification "Quest Diagnostics" "Medical laboratory and diagnostic testing" "Medical" "HEALTHCARE"
test_classification "DaVita" "Kidney dialysis services and healthcare" "Healthcare" "HEALTHCARE"
test_classification "HCA Healthcare" "Hospital and healthcare facilities" "Hospital" "HEALTHCARE"
test_classification "Walgreens" "Pharmacy and health products retailer" "Pharmacy" "HEALTHCARE"
test_classification "Anthem Blue Cross" "Health insurance and managed care" "Insurance" "HEALTHCARE"
test_classification "Labcorp" "Clinical laboratory services" "Medical" "HEALTHCARE"
test_classification "Dental Associates" "Dental care and orthodontic services" "Dental" "HEALTHCARE"
test_classification "VCA Animal Hospitals" "Veterinary care and animal hospitals" "Veterinary" "HEALTHCARE"

echo ""
echo "=== CATEGORY 6: AMBIGUOUS/COMPLEX CASES (10 cases) ==="
test_classification "Diversified Holdings LLC" "A diversified holding company with operations in multiple unrelated sectors" "General" "AMBIGUOUS"
test_classification "ABC Multi-Services" "Provides cleaning, landscaping, and security services to commercial clients" "Services" "AMBIGUOUS"
test_classification "TechFood Innovations" "Food technology startup using AI to optimize restaurant operations" "Technology" "AMBIGUOUS"
test_classification "GreenBuild Consulting" "Sustainable construction consulting and green building certification" "Consulting" "AMBIGUOUS"
test_classification "CloudKitchen Brands" "Virtual restaurant operator running delivery-only food brands" "Restaurant" "AMBIGUOUS"
test_classification "FinHealth Solutions" "Financial wellness platform for healthcare workers" "Financial" "AMBIGUOUS"
test_classification "EduTech Academy" "Online learning platform with K-12 and professional courses" "Education" "AMBIGUOUS"
test_classification "SportsMed Clinic" "Sports medicine clinic and physical therapy center" "Medical" "AMBIGUOUS"
test_classification "Crypto Capital Partners" "Cryptocurrency investment and blockchain consulting" "Financial" "AMBIGUOUS"
test_classification "AgriTech Farms" "High-tech vertical farming and agricultural technology" "Agriculture" "AMBIGUOUS"

echo ""
echo "=============================================="
echo "ACCURACY RESULTS"
echo "=============================================="
echo ""

# Calculate percentages
# Ensure accuracy is always a decimal string for consistent comparisons
if [ $total -gt 0 ]; then
    accuracy=$(echo "scale=1; $correct * 100 / $total" | bc)
else
    accuracy="0.0"
fi

# Ensure accuracy is set to a valid decimal string (handle bc failures or empty results)
if [ -z "$accuracy" ]; then
    accuracy="0.0"
else
    # Normalize to ensure it's a valid decimal format
    accuracy=$(echo "scale=1; $accuracy" | bc 2>/dev/null || echo "0.0")
fi

if [ $layer1_total -gt 0 ]; then
    layer1_acc=$(echo "scale=1; $layer1_correct * 100 / $layer1_total" | bc)
else
    layer1_acc="N/A"
fi

if [ $layer2_total -gt 0 ]; then
    layer2_acc=$(echo "scale=1; $layer2_correct * 100 / $layer2_total" | bc)
else
    layer2_acc="N/A"
fi

if [ $layer3_total -gt 0 ]; then
    layer3_acc=$(echo "scale=1; $layer3_correct * 100 / $layer3_total" | bc)
else
    layer3_acc="N/A"
fi

echo "📊 OVERALL ACCURACY: $correct / $total = $accuracy%"
echo ""
echo "📈 BY LAYER:"
echo "   Layer 1 (Multi-Strategy): $layer1_correct / $layer1_total = $layer1_acc%"
echo "   Layer 2 (Embeddings):     $layer2_correct / $layer2_total = $layer2_acc%"
echo "   Layer 3 (LLM):            $layer3_correct / $layer3_total = $layer3_acc%"
echo ""
echo "🎯 Phase 4 Target: 90-95%"
# Use bc -l for consistent floating-point comparison
if [ -n "$accuracy" ] && (( $(echo "$accuracy >= 90" | bc -l) )); then
    echo "✅ TARGET ACHIEVED!"
elif [ -n "$accuracy" ] && (( $(echo "$accuracy >= 85" | bc -l) )); then
    echo "⚠️ Close to target (85-90%)"
else
    echo "❌ Below target (<85%)"
fi
echo ""
echo "=============================================="

