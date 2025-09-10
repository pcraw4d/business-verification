#!/bin/bash

# Populate Supabase Database with Sample Data for Keyword Classification System
# This script adds sample industries, keywords, and classification codes

set -e

echo "🚀 Populating Supabase Database with Sample Data"

# Check if required environment variables are set
if [ -z "$SUPABASE_URL" ]; then
    echo "❌ Error: SUPABASE_URL environment variable is required"
    exit 1
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "❌ Error: SUPABASE_ANON_KEY environment variable is required"
    exit 1
fi

echo "✅ Environment variables validated"

# Function to make API request
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    curl -s -X "$method" \
      "$SUPABASE_URL/rest/v1/$endpoint" \
      -H "apikey: $SUPABASE_ANON_KEY" \
      -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
      -H "Content-Type: application/json" \
      -H "Prefer: resolution=ignore-duplicates" \
      ${data:+-d "$data"}
}

# Insert sample industries
echo "📝 Inserting sample industries..."
make_request "POST" "industries" '{"name": "Technology", "description": "Technology and software companies", "category": "traditional", "confidence_threshold": 0.80}'
make_request "POST" "industries" '{"name": "Healthcare", "description": "Healthcare and medical services", "category": "traditional", "confidence_threshold": 0.85}'
make_request "POST" "industries" '{"name": "Financial Services", "description": "Banking, finance, and investment services", "category": "traditional", "confidence_threshold": 0.90}'
make_request "POST" "industries" '{"name": "Retail", "description": "Retail and consumer goods", "category": "traditional", "confidence_threshold": 0.75}'
make_request "POST" "industries" '{"name": "Manufacturing", "description": "Manufacturing and industrial production", "category": "traditional", "confidence_threshold": 0.80}'

echo "✅ Sample industries inserted"

# Insert sample keywords
echo "📝 Inserting sample keywords..."

# Technology keywords
make_request "POST" "industry_keywords" '{"industry_id": 1, "keyword": "software", "weight": 0.9, "context": "technical", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 1, "keyword": "technology", "weight": 0.8, "context": "general", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 1, "keyword": "development", "weight": 0.7, "context": "technical", "is_primary": false}'
make_request "POST" "industry_keywords" '{"industry_id": 1, "keyword": "platform", "weight": 0.6, "context": "technical", "is_primary": false}'
make_request "POST" "industry_keywords" '{"industry_id": 1, "keyword": "digital", "weight": 0.5, "context": "general", "is_primary": false}'

# Healthcare keywords
make_request "POST" "industry_keywords" '{"industry_id": 2, "keyword": "medical", "weight": 0.9, "context": "business", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 2, "keyword": "healthcare", "weight": 0.8, "context": "business", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 2, "keyword": "patient", "weight": 0.6, "context": "business", "is_primary": false}'
make_request "POST" "industry_keywords" '{"industry_id": 2, "keyword": "clinic", "weight": 0.7, "context": "business", "is_primary": false}'

# Financial Services keywords
make_request "POST" "industry_keywords" '{"industry_id": 3, "keyword": "bank", "weight": 0.9, "context": "business", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 3, "keyword": "finance", "weight": 0.8, "context": "business", "is_primary": true}'
make_request "POST" "industry_keywords" '{"industry_id": 3, "keyword": "credit", "weight": 0.7, "context": "business", "is_primary": false}'
make_request "POST" "industry_keywords" '{"industry_id": 3, "keyword": "investment", "weight": 0.6, "context": "business", "is_primary": false}'

echo "✅ Sample keywords inserted"

# Insert sample classification codes
echo "📝 Inserting sample classification codes..."

# Technology codes
make_request "POST" "classification_codes" '{"industry_id": 1, "code_type": "NAICS", "code": "541511", "description": "Custom Computer Programming Services", "confidence": 0.9, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 1, "code_type": "NAICS", "code": "541512", "description": "Computer Systems Design Services", "confidence": 0.85, "is_primary": false}'
make_request "POST" "classification_codes" '{"industry_id": 1, "code_type": "MCC", "code": "5734", "description": "Computer Software Stores", "confidence": 0.8, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 1, "code_type": "SIC", "code": "7372", "description": "Prepackaged Software", "confidence": 0.85, "is_primary": true}'

# Healthcare codes
make_request "POST" "classification_codes" '{"industry_id": 2, "code_type": "NAICS", "code": "621111", "description": "Offices of Physicians", "confidence": 0.9, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 2, "code_type": "NAICS", "code": "621112", "description": "Offices of Physicians, Mental Health Specialists", "confidence": 0.85, "is_primary": false}'
make_request "POST" "classification_codes" '{"industry_id": 2, "code_type": "MCC", "code": "8099", "description": "Health Practitioners, Not Elsewhere Classified", "confidence": 0.8, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 2, "code_type": "SIC", "code": "8011", "description": "Offices and Clinics of Doctors of Medicine", "confidence": 0.85, "is_primary": true}'

# Financial Services codes
make_request "POST" "classification_codes" '{"industry_id": 3, "code_type": "NAICS", "code": "522110", "description": "Commercial Banking", "confidence": 0.9, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 3, "code_type": "NAICS", "code": "522120", "description": "Savings Institutions", "confidence": 0.85, "is_primary": false}'
make_request "POST" "classification_codes" '{"industry_id": 3, "code_type": "MCC", "code": "6011", "description": "Automated Teller Machine Services", "confidence": 0.8, "is_primary": true}'
make_request "POST" "classification_codes" '{"industry_id": 3, "code_type": "SIC", "code": "6021", "description": "National Commercial Banks", "confidence": 0.85, "is_primary": true}'

echo "✅ Sample classification codes inserted"

echo "🎉 Database population completed successfully!"
echo "📊 Sample data includes:"
echo "   - 5 industries (Technology, Healthcare, Financial Services, Retail, Manufacturing)"
echo "   - 15 keywords with weights and contexts"
echo "   - 12 classification codes (NAICS, MCC, SIC) across 3 industries"
echo ""
echo "✅ Database is ready for testing the new keyword classification system!"
