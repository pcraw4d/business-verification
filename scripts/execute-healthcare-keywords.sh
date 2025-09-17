#!/bin/bash

# =============================================================================
# HEALTHCARE KEYWORDS EXECUTION SCRIPT
# =============================================================================
# This script executes the comprehensive healthcare keywords addition for all
# 4 healthcare industries to achieve >85% classification accuracy.
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

echo "🏥 Starting Healthcare Keywords Implementation..."
echo "================================================================================"
echo "Adding comprehensive healthcare keywords for 4 industries:"
echo "1. Medical Practices: 50+ keywords (family medicine, specialists, clinical services)"
echo "2. Healthcare Services: 50+ keywords (hospitals, clinics, medical facilities)"
echo "3. Mental Health: 50+ keywords (counseling, therapy, psychological services)"
echo "4. Healthcare Technology: 50+ keywords (medical devices, health IT, digital health)"
echo "Total: 200+ healthcare-specific keywords with base weights 0.50-1.00"
echo "================================================================================"

# Execute the healthcare keywords addition
echo "🏥 Adding healthcare keywords..."
psql "$DB_URL" -f scripts/add-healthcare-keywords.sql

if [ $? -eq 0 ]; then
    echo "✅ Healthcare keywords added successfully"
else
    echo "❌ Error: Failed to add healthcare keywords"
    exit 1
fi

# Run comprehensive testing
echo "🧪 Running healthcare keywords testing..."
psql "$DB_URL" -f scripts/test-healthcare-keywords.sql

if [ $? -eq 0 ]; then
    echo "✅ Healthcare keywords testing completed successfully"
else
    echo "❌ Error: Healthcare keywords testing failed"
    exit 1
fi

echo "================================================================================"
echo "🎉 HEALTHCARE KEYWORDS IMPLEMENTATION COMPLETED SUCCESSFULLY"
echo "================================================================================"
echo "✅ All 4 healthcare industries now have comprehensive keyword coverage"
echo "✅ 200+ healthcare-specific keywords added with appropriate weights"
echo "✅ Keywords are active and ready for classification testing"
echo "✅ System ready for >85% healthcare classification accuracy"
echo "================================================================================"
echo "Next Steps:"
echo "1. Test healthcare classification with sample businesses"
echo "2. Verify classification accuracy meets >85% target"
echo "3. Proceed to next subtask (3.2.3: Technology Keywords)"
echo "================================================================================"
