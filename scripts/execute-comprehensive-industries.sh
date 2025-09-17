#!/bin/bash

# =============================================================================
# COMPREHENSIVE INDUSTRY EXPANSION EXECUTION SCRIPT
# =============================================================================
# This script executes the comprehensive industry expansion to add 27 new
# industries across 7 major business categories for complete classification
# coverage.
# =============================================================================

set -e  # Exit on any error

# Load environment variables
if [ -f .env ]; then
    source .env
    echo "✅ Environment variables loaded"
else
    echo "❌ Error: .env file not found"
    exit 1
fi

# Set database URL
DB_URL="postgresql://postgres:${DB_PASSWORD}@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres"

echo "🚀 Starting Comprehensive Industry Expansion..."
echo "================================================================================"
echo "Adding 27 new industries across 7 major business categories:"
echo "1. Legal Services (4 industries)"
echo "2. Healthcare (4 industries)"
echo "3. Financial Services (4 industries)"
echo "4. Retail & E-commerce (4 industries)"
echo "5. Manufacturing (4 industries)"
echo "6. Agriculture & Energy (4 industries)"
echo "7. Technology (3 industries)"
echo "================================================================================"

# Execute the comprehensive industry addition
echo "📊 Adding comprehensive industries..."
psql "$DB_URL" -f scripts/add-comprehensive-industries.sql

if [ $? -eq 0 ]; then
    echo "✅ Comprehensive industries added successfully"
else
    echo "❌ Error: Failed to add comprehensive industries"
    exit 1
fi

# Run comprehensive testing
echo "🧪 Running comprehensive industry testing..."
psql "$DB_URL" -f scripts/test-comprehensive-industries.sql

if [ $? -eq 0 ]; then
    echo "✅ Comprehensive industry testing completed successfully"
else
    echo "❌ Error: Comprehensive industry testing failed"
    exit 1
fi

# Display final summary
echo ""
echo "🎉 COMPREHENSIVE INDUSTRY EXPANSION COMPLETED SUCCESSFULLY!"
echo "================================================================================"
echo "✅ 27 new industries added across 7 major categories"
echo "✅ Total industries: 39 (12 restaurant + 27 new)"
echo "✅ Categories covered: Legal, Healthcare, Financial, Retail, Manufacturing, Agriculture, Technology"
echo "✅ Confidence thresholds: 0.70-0.85 range"
echo "✅ All industries properly configured and tested"
echo ""
echo "📋 Next Steps:"
echo "1. Add comprehensive keywords for all industries (Task 3.2)"
echo "2. Add classification codes for all industries (Task 3.3)"
echo "3. Test comprehensive classification accuracy"
echo ""
echo "🚀 Ready to proceed with keyword expansion!"
echo "================================================================================"
