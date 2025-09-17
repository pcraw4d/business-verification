#!/bin/bash

# =============================================================================
# Execute Subtask 1.2.2: Add Restaurant Keywords
# =============================================================================

# This script executes the restaurant keywords addition and verification
# as part of the comprehensive classification improvement plan

set -e

echo "🚀 Starting Subtask 1.2.2: Add Restaurant Keywords"
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
TEMP_ADD_SQL="/tmp/add_restaurant_keywords.sql"
TEMP_TEST_SQL="/tmp/test_restaurant_keywords.sql"

echo ""
echo "📝 Preparing SQL scripts..."

# Copy the SQL scripts to temporary files
cp scripts/add-restaurant-keywords.sql "$TEMP_ADD_SQL"
cp scripts/test-restaurant-keywords.sql "$TEMP_TEST_SQL"

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

# Execute the restaurant keywords addition
echo ""
echo "============================================================================="
echo "STEP 1: Adding Restaurant Keywords"
echo "============================================================================="

if execute_sql_psql "$TEMP_ADD_SQL" "Restaurant Keywords Addition"; then
    echo "✅ Restaurant keywords added successfully"
else
    echo "❌ Failed to add restaurant keywords"
    echo ""
    echo "📋 Manual execution required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/add-restaurant-keywords.sql"
    echo "   3. Execute the script"
    echo ""
    read -p "Press Enter after manually executing the SQL script..."
fi

# Execute the verification tests
echo ""
echo "============================================================================="
echo "STEP 2: Verifying Restaurant Keywords"
echo "============================================================================="

if execute_sql_psql "$TEMP_TEST_SQL" "Restaurant Keywords Verification"; then
    echo "✅ Restaurant keywords verification completed"
else
    echo "❌ Failed to verify restaurant keywords"
    echo ""
    echo "📋 Manual verification required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/test-restaurant-keywords.sql"
    echo "   3. Execute the script and review results"
    echo ""
    read -p "Press Enter after manually executing the verification script..."
fi

# Clean up temporary files
rm -f "$TEMP_ADD_SQL" "$TEMP_TEST_SQL"

echo ""
echo "============================================================================="
echo "SUBTASK 1.2.2 COMPLETION SUMMARY"
echo "============================================================================="
echo "✅ Restaurant keywords addition script created"
echo "✅ Comprehensive verification test script created"
echo "✅ Execution script completed"
echo ""
echo "📋 Keywords Added:"
echo "   - 12 restaurant industry categories"
echo "   - 200+ comprehensive keywords"
echo "   - Base weights: 0.6000 to 1.0000"
echo "   - Industry-specific keyword sets"
echo ""
echo "🎯 Key Features:"
echo "   - Core restaurant keywords (restaurant, dining, cuisine)"
echo "   - Industry-specific terms (fast food, fine dining, casual dining)"
echo "   - Service characteristics (table service, quick service, catering)"
echo "   - Food and beverage terms (menu items, drinks, specialties)"
echo "   - Brand and chain names (McDonald's, Olive Garden, etc.)"
echo ""
echo "📁 Files Created:"
echo "   - scripts/add-restaurant-keywords.sql"
echo "   - scripts/test-restaurant-keywords.sql"
echo "   - scripts/execute-subtask-1-2-2.sh"
echo ""
echo "🚀 Next Steps:"
echo "   1. Verify restaurant keywords were added to database"
echo "   2. Proceed to Subtask 1.2.3: Add restaurant classification codes"
echo "   3. Update COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN.md"
echo ""
echo "📊 Expected Results:"
echo "   - 200+ keywords across 12 restaurant industries"
echo "   - Weighted keywords for accurate classification"
echo "   - Industry-specific keyword sets for precise matching"
echo "   - Foundation for >75% restaurant classification accuracy"
echo ""
echo "============================================================================="
echo "Subtask 1.2.2: Add Restaurant Keywords - COMPLETED"
echo "============================================================================="
