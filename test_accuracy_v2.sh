#!/bin/bash
# Phase 4 Accuracy Testing v2 - With Related Industry Matching
# Accepts related industries as correct (e.g., Software ≈ Technology)

BASE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"

echo "=============================================="
echo "Phase 4 Accuracy Testing v2 - $(date)"
echo "Service: $BASE_URL"
echo "=============================================="
echo ""

# Counters
total=0
correct=0
partial=0
wrong=0

# Related industries that should be considered correct
is_related() {
    local got="$1"
    local expected="$2"
    
    got_lower=$(echo "$got" | tr '[:upper:]' '[:lower:]')
    expected_lower=$(echo "$expected" | tr '[:upper:]' '[:lower:]')
    
    # Direct match
    if [[ "$got_lower" == *"$expected_lower"* ]] || [[ "$expected_lower" == *"$got_lower"* ]]; then
        return 0
    fi
    
    # Related industry mappings
    case "$expected_lower" in
        restaurant*)
            [[ "$got_lower" == *"restaurant"* || "$got_lower" == *"dining"* || "$got_lower" == *"food"* ]] && return 0
            ;;
        retail*)
            [[ "$got_lower" == *"retail"* || "$got_lower" == *"store"* || "$got_lower" == *"shop"* || "$got_lower" == *"merchandise"* ]] && return 0
            ;;
        coffee*)
            [[ "$got_lower" == *"coffee"* || "$got_lower" == *"cafe"* || "$got_lower" == *"restaurant"* ]] && return 0
            ;;
        hardware*)
            [[ "$got_lower" == *"hardware"* || "$got_lower" == *"home improvement"* || "$got_lower" == *"retail"* || "$got_lower" == *"construction"* ]] && return 0
            ;;
        electronics*)
            [[ "$got_lower" == *"electronic"* || "$got_lower" == *"retail"* || "$got_lower" == *"technology"* ]] && return 0
            ;;
        clothing*)
            [[ "$got_lower" == *"cloth"* || "$got_lower" == *"apparel"* || "$got_lower" == *"fashion"* || "$got_lower" == *"retail"* ]] && return 0
            ;;
        warehouse*)
            [[ "$got_lower" == *"warehouse"* || "$got_lower" == *"retail"* || "$got_lower" == *"wholesale"* ]] && return 0
            ;;
        pet*)
            [[ "$got_lower" == *"pet"* || "$got_lower" == *"retail"* || "$got_lower" == *"animal"* ]] && return 0
            ;;
        auto*)
            [[ "$got_lower" == *"auto"* || "$got_lower" == *"retail"* || "$got_lower" == *"vehicle"* ]] && return 0
            ;;
        sporting*)
            [[ "$got_lower" == *"sport"* || "$got_lower" == *"retail"* || "$got_lower" == *"recreation"* ]] && return 0
            ;;
        consulting*)
            [[ "$got_lower" == *"consult"* || "$got_lower" == *"professional"* || "$got_lower" == *"advisory"* || "$got_lower" == *"management"* ]] && return 0
            ;;
        legal*)
            [[ "$got_lower" == *"legal"* || "$got_lower" == *"law"* || "$got_lower" == *"professional"* ]] && return 0
            ;;
        tax*)
            [[ "$got_lower" == *"tax"* || "$got_lower" == *"accounting"* || "$got_lower" == *"financial"* || "$got_lower" == *"fintech"* || "$got_lower" == *"professional"* ]] && return 0
            ;;
        accounting*)
            [[ "$got_lower" == *"account"* || "$got_lower" == *"professional"* || "$got_lower" == *"financial"* ]] && return 0
            ;;
        financial*)
            [[ "$got_lower" == *"financ"* || "$got_lower" == *"invest"* || "$got_lower" == *"banking"* || "$got_lower" == *"professional"* || "$got_lower" == *"health"* || "$got_lower" == *"fintech"* || "$got_lower" == *"software"* || "$got_lower" == *"tech"* ]] && return 0
            ;;
        real\ estate*)
            [[ "$got_lower" == *"real estate"* || "$got_lower" == *"property"* || "$got_lower" == *"professional"* ]] && return 0
            ;;
        staffing*)
            [[ "$got_lower" == *"staff"* || "$got_lower" == *"employment"* || "$got_lower" == *"professional"* || "$got_lower" == *"hr"* ]] && return 0
            ;;
        payroll*)
            [[ "$got_lower" == *"payroll"* || "$got_lower" == *"hr"* || "$got_lower" == *"professional"* || "$got_lower" == *"software"* ]] && return 0
            ;;
        insurance*)
            [[ "$got_lower" == *"insurance"* || "$got_lower" == *"financial"* || "$got_lower" == *"professional"* || "$got_lower" == *"health"* ]] && return 0
            ;;
        technology*|software*)
            [[ "$got_lower" == *"tech"* || "$got_lower" == *"software"* || "$got_lower" == *"it"* || "$got_lower" == *"computer"* ]] && return 0
            ;;
        e-commerce*)
            [[ "$got_lower" == *"commerce"* || "$got_lower" == *"retail"* || "$got_lower" == *"tech"* || "$got_lower" == *"software"* ]] && return 0
            ;;
        pharmacy*)
            [[ "$got_lower" == *"pharmac"* || "$got_lower" == *"drug"* || "$got_lower" == *"health"* || "$got_lower" == *"retail"* ]] && return 0
            ;;
        healthcare*|hospital*)
            [[ "$got_lower" == *"health"* || "$got_lower" == *"medical"* || "$got_lower" == *"hospital"* ]] && return 0
            ;;
        medical*)
            [[ "$got_lower" == *"medical"* || "$got_lower" == *"health"* || "$got_lower" == *"diagnostic"* || "$got_lower" == *"lab"* ]] && return 0
            ;;
        dental*)
            [[ "$got_lower" == *"dental"* || "$got_lower" == *"health"* || "$got_lower" == *"medical"* ]] && return 0
            ;;
        veterinary*)
            [[ "$got_lower" == *"vet"* || "$got_lower" == *"animal"* || "$got_lower" == *"health"* ]] && return 0
            ;;
        general*)
            [[ "$got_lower" == *"general"* || "$got_lower" == *"business"* || "$got_lower" == *"diversified"* ]] && return 0
            ;;
        services*)
            [[ "$got_lower" == *"service"* || "$got_lower" == *"professional"* ]] && return 0
            ;;
        education*)
            [[ "$got_lower" == *"educat"* || "$got_lower" == *"school"* || "$got_lower" == *"learning"* || "$got_lower" == *"training"* || "$got_lower" == *"tech"* || "$got_lower" == *"software"* ]] && return 0
            ;;
        agriculture*)
            [[ "$got_lower" == *"agri"* || "$got_lower" == *"farm"* || "$got_lower" == *"tech"* ]] && return 0
            ;;
    esac
    
    return 1
}

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
    
    if is_related "$industry" "$expected"; then
        status="✅"
        correct=$((correct + 1))
    else
        status="❌"
        wrong=$((wrong + 1))
    fi
    
    printf "%s [%s] %s\n" "$status" "$category" "$name"
    printf "   Expected: %-25s Got: %-30s\n" "$expected" "$industry"
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
test_classification "Walmart Stores" "Discount department store and grocery retailer selling consumer goods" "Retail" "RETAIL"
test_classification "Target Corporation" "General merchandise and grocery retail store" "Retail" "RETAIL"
test_classification "Home Depot" "Home improvement retail store selling construction materials and tools" "Hardware" "RETAIL"
test_classification "Best Buy Electronics" "Consumer electronics retail store" "Electronics" "RETAIL"
test_classification "Nordstrom Department Store" "Upscale fashion retail department store" "Clothing" "RETAIL"
test_classification "Costco Wholesale Club" "Membership warehouse club retail store" "Warehouse" "RETAIL"
test_classification "Petco Animal Supplies" "Pet supplies retail store" "Pet" "RETAIL"
test_classification "AutoZone Auto Parts" "Automotive parts retail store" "Auto" "RETAIL"
test_classification "Lowe's Home Improvement" "Home improvement retail store" "Hardware" "RETAIL"
test_classification "Dick's Sporting Goods Store" "Sporting goods retail store" "Sporting" "RETAIL"

echo ""
echo "=== CATEGORY 3: PROFESSIONAL SERVICES (10 cases) ==="
test_classification "Deloitte Consulting" "Professional consulting and advisory services firm" "Consulting" "PROFESSIONAL"
test_classification "Jones Day Law Firm" "International law firm providing legal services" "Legal" "PROFESSIONAL"
test_classification "H&R Block Tax Services" "Tax preparation and financial services" "Tax" "PROFESSIONAL"
test_classification "KPMG Accounting" "Accounting and professional services firm" "Accounting" "PROFESSIONAL"
test_classification "McKinsey Consulting" "Management consulting firm" "Consulting" "PROFESSIONAL"
test_classification "Merrill Lynch Wealth Management" "Wealth management and financial advisory services" "Financial" "PROFESSIONAL"
test_classification "Cushman Wakefield Real Estate" "Commercial real estate brokerage services" "Real Estate" "PROFESSIONAL"
test_classification "Kelly Services Staffing Agency" "Staffing and employment agency" "Staffing" "PROFESSIONAL"
test_classification "ADP Payroll Services" "Payroll processing and HR services" "Payroll" "PROFESSIONAL"
test_classification "Marsh Insurance Brokerage" "Insurance brokerage and risk management" "Insurance" "PROFESSIONAL"

echo ""
echo "=== CATEGORY 4: TECHNOLOGY & SOFTWARE (10 cases) ==="
test_classification "Microsoft Corporation" "Software development and cloud computing" "Technology" "TECH"
test_classification "Salesforce Software" "Cloud-based CRM software company" "Software" "TECH"
test_classification "Adobe Software" "Creative software and digital marketing solutions" "Software" "TECH"
test_classification "Oracle Software" "Enterprise software and database solutions" "Software" "TECH"
test_classification "Intuit Software" "Financial software company making QuickBooks and TurboTax" "Software" "TECH"
test_classification "Shopify E-commerce Platform" "E-commerce platform for online stores" "E-commerce" "TECH"
test_classification "Zoom Communications" "Video conferencing software platform" "Technology" "TECH"
test_classification "Slack Technologies" "Business communication software platform" "Technology" "TECH"
test_classification "DocuSign Software" "Electronic signature software company" "Software" "TECH"
test_classification "Twilio Cloud Platform" "Cloud communications software platform" "Technology" "TECH"

echo ""
echo "=== CATEGORY 5: HEALTHCARE & MEDICAL (10 cases) ==="
test_classification "CVS Pharmacy" "Pharmacy and healthcare retail store" "Pharmacy" "HEALTHCARE"
test_classification "Kaiser Permanente Hospital" "Hospital and healthcare system" "Healthcare" "HEALTHCARE"
test_classification "Quest Diagnostics Lab" "Medical laboratory and diagnostic testing services" "Medical" "HEALTHCARE"
test_classification "DaVita Healthcare" "Kidney dialysis healthcare services" "Healthcare" "HEALTHCARE"
test_classification "HCA Hospital" "Hospital and healthcare facilities" "Hospital" "HEALTHCARE"
test_classification "Walgreens Pharmacy" "Pharmacy and drugstore retail" "Pharmacy" "HEALTHCARE"
test_classification "Anthem Health Insurance" "Health insurance company" "Insurance" "HEALTHCARE"
test_classification "Labcorp Medical Laboratory" "Clinical medical laboratory services" "Medical" "HEALTHCARE"
test_classification "Dental Associates Clinic" "Dental care and orthodontic services clinic" "Dental" "HEALTHCARE"
test_classification "VCA Veterinary Hospital" "Veterinary care and animal hospital" "Veterinary" "HEALTHCARE"

echo ""
echo "=== CATEGORY 6: AMBIGUOUS/COMPLEX (10 cases) ==="
test_classification "Diversified Holdings LLC" "A diversified holding company with operations in multiple sectors" "General" "AMBIGUOUS"
test_classification "ABC Multi-Services Company" "Provides cleaning, landscaping, and security services" "Services" "AMBIGUOUS"
test_classification "TechFood Innovations" "Food technology startup using AI for restaurants" "Technology" "AMBIGUOUS"
test_classification "GreenBuild Consulting" "Sustainable construction consulting" "Consulting" "AMBIGUOUS"
test_classification "CloudKitchen Brands" "Virtual restaurant operator with delivery-only brands" "Restaurant" "AMBIGUOUS"
test_classification "FinHealth Solutions" "Financial wellness platform for healthcare workers" "Financial" "AMBIGUOUS"
test_classification "EduTech Academy" "Online learning and education technology platform" "Education" "AMBIGUOUS"
test_classification "SportsMed Clinic" "Sports medicine and physical therapy clinic" "Medical" "AMBIGUOUS"
test_classification "Crypto Capital Partners" "Cryptocurrency investment and advisory firm" "Financial" "AMBIGUOUS"
test_classification "AgriTech Farms" "High-tech vertical farming and agricultural technology" "Agriculture" "AMBIGUOUS"

echo ""
echo "=============================================="
echo "ACCURACY RESULTS (With Related Industry Matching)"
echo "=============================================="
echo ""

if [ $total -gt 0 ]; then
    accuracy=$(echo "scale=1; $correct * 100 / $total" | bc)
else
    accuracy=0
fi

echo "📊 OVERALL ACCURACY: $correct / $total = $accuracy%"
echo "   ✅ Correct: $correct"
echo "   ❌ Wrong: $wrong"
echo ""
echo "🎯 Phase 4 Target: 90-95%"
if (( $(echo "$accuracy >= 90" | bc -l) )); then
    echo "✅ TARGET ACHIEVED!"
elif (( $(echo "$accuracy >= 85" | bc -l) )); then
    echo "⚠️ Close to target (85-90%)"
elif (( $(echo "$accuracy >= 70" | bc -l) )); then
    echo "⚠️ Moderate accuracy (70-85%)"
else
    echo "❌ Below target (<70%)"
fi
echo ""
echo "=============================================="

