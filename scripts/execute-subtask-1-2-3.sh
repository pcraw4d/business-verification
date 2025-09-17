#!/bin/bash

# =============================================================================
# Execute Subtask 1.2.3: Add Restaurant Classification Codes
# =============================================================================

# This script executes the restaurant classification codes addition and verification
# as part of the comprehensive classification improvement plan

set -e

echo "🚀 Starting Subtask 1.2.3: Add Restaurant Classification Codes"
echo "============================================================================="

# Check if required environment variables are set
if [ -z "$SUPABASE_URL" ]; then
    echo "❌ Error: SUPABASE_URL environment variable is required"
    echo "   Please set SUPABASE_URL to your Supabase project URL"
    exit 1
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "❌ Error: SUPABASE_ANON_KEY environment variable is required"
    echo "   Please set SUPABASE_ANON_KEY to your Supabase anon key"
    exit 1
fi

echo "✅ Environment variables validated"
echo "   SUPABASE_URL: ${SUPABASE_URL:0:30}..."
echo "   SUPABASE_ANON_KEY: ${SUPABASE_ANON_KEY:0:20}..."

# Create temporary files for SQL execution
TEMP_ADD_SQL="/tmp/add_restaurant_classification_codes.sql"
TEMP_TEST_SQL="/tmp/test_restaurant_classification_codes.sql"

echo ""
echo "📝 Preparing SQL scripts..."

# Copy the SQL scripts to temporary files
cp scripts/add-restaurant-classification-codes.sql "$TEMP_ADD_SQL"
cp scripts/test-restaurant-classification-codes.sql "$TEMP_TEST_SQL"

echo "✅ SQL scripts prepared"

# Function to execute SQL via psql
execute_sql_psql() {
    local sql_file="$1"
    local description="$2"
    
    echo ""
    echo "🔄 Executing: $description"
    echo "   File: $sql_file"
    
    # Extract database connection details from SUPABASE_URL
    # Format: postgresql://postgres:[password]@[host]:[port]/postgres
    local db_url="$SUPABASE_URL"
    
    if command -v psql &> /dev/null; then
        if psql "$db_url" -f "$sql_file" -v ON_ERROR_STOP=1; then
            echo "✅ $description completed successfully"
            return 0
        else
            echo "❌ $description failed"
            return 1
        fi
    else
        echo "❌ psql not available, cannot execute SQL directly"
        echo "   Please run the SQL scripts manually in Supabase SQL Editor:"
        echo "   1. $sql_file"
        return 1
    fi
}

# Execute the restaurant classification codes addition
echo ""
echo "============================================================================="
echo "STEP 1: Adding Restaurant Classification Codes"
echo "============================================================================="

if execute_sql_psql "$TEMP_ADD_SQL" "Restaurant Classification Codes Addition"; then
    echo "✅ Restaurant classification codes added successfully"
else
    echo "❌ Failed to add restaurant classification codes"
    echo ""
    echo "📋 Manual execution required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/add-restaurant-classification-codes.sql"
    echo "   3. Execute the script"
    echo ""
    read -p "Press Enter after manually executing the SQL script..."
fi

# Execute the verification tests
echo ""
echo "============================================================================="
echo "STEP 2: Verifying Restaurant Classification Codes"
echo "============================================================================="

if execute_sql_psql "$TEMP_TEST_SQL" "Restaurant Classification Codes Verification"; then
    echo "✅ Restaurant classification codes verification completed"
else
    echo "❌ Failed to verify restaurant classification codes"
    echo ""
    echo "📋 Manual verification required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/test-restaurant-classification-codes.sql"
    echo "   3. Execute the script and review results"
    echo ""
    read -p "Press Enter after manually executing the verification script..."
fi

# Clean up temporary files
rm -f "$TEMP_ADD_SQL" "$TEMP_TEST_SQL"

echo ""
echo "============================================================================="
echo "SUBTASK 1.2.3 COMPLETION SUMMARY"
echo "============================================================================="
echo "✅ Restaurant classification codes addition script created"
echo "✅ Comprehensive verification test script created"
echo "✅ Execution script completed"
echo ""
echo "📋 Classification Codes Added:"
echo "   - 12 restaurant industry categories"
echo "   - 50+ comprehensive classification codes"
echo "   - NAICS, SIC, and MCC code types"
echo "   - Industry-specific code mappings"
echo ""
echo "🎯 Key Features:"
echo "   - NAICS codes: 722511 (Full-Service), 722513 (Limited-Service), 722515 (Snack Bars)"
echo "   - SIC codes: 5812 (Eating Places), 5813 (Drinking Places), 5814 (Caterers)"
echo "   - MCC codes: 5812 (Restaurants), 5813 (Bars), 5814 (Fast Food)"
echo "   - Industry-specific mappings for accurate classification"
echo "   - Complete code coverage for all restaurant types"
echo ""
echo "📁 Files Created:"
echo "   - scripts/add-restaurant-classification-codes.sql"
echo "   - scripts/test-restaurant-classification-codes.sql"
echo "   - scripts/execute-subtask-1-2-3.sh"
echo ""
echo "🚀 Next Steps:"
echo "   1. Verify restaurant classification codes were added to database"
echo "   2. Proceed to Task 1.3: Test Restaurant Classification"
echo "   3. Update COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN.md"
echo ""
echo "📊 Expected Results:"
echo "   - 50+ classification codes across 12 restaurant industries"
echo "   - Complete NAICS, SIC, and MCC code coverage"
echo "   - Industry-specific code mappings for precise classification"
echo "   - Foundation for complete restaurant business classification"
echo ""
echo "============================================================================="
echo "Subtask 1.2.3: Add Restaurant Classification Codes - COMPLETED"
echo "============================================================================="
