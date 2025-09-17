#!/bin/bash

# =============================================================================
# Execute Subtask 1.2.1: Add Restaurant Industries
# =============================================================================

# This script executes the restaurant industries addition and verification
# as part of the comprehensive classification improvement plan

set -e

echo "🚀 Starting Subtask 1.2.1: Add Restaurant Industries"
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
TEMP_ADD_SQL="/tmp/add_restaurant_industries.sql"
TEMP_TEST_SQL="/tmp/test_restaurant_industries.sql"

echo ""
echo "📝 Preparing SQL scripts..."

# Copy the SQL scripts to temporary files
cp scripts/add-restaurant-industries.sql "$TEMP_ADD_SQL"
cp scripts/test-restaurant-industries.sql "$TEMP_TEST_SQL"

echo "✅ SQL scripts prepared"

# Function to execute SQL via Supabase API
execute_sql() {
    local sql_file="$1"
    local description="$2"
    
    echo ""
    echo "🔄 Executing: $description"
    echo "   File: $sql_file"
    
    # Read SQL content
    local sql_content=$(cat "$sql_file")
    
    # Execute via Supabase API
    local response=$(curl -s -X POST \
        "$SUPABASE_URL/rest/v1/rpc/execute_sql" \
        -H "apikey: $SUPABASE_ANON_KEY" \
        -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"sql\": \"$sql_content\"}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ $description completed successfully"
        return 0
    else
        echo "❌ $description failed"
        echo "   Response: $response"
        return 1
    fi
}

# Alternative method using psql if available
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

# Execute the restaurant industries addition
echo ""
echo "============================================================================="
echo "STEP 1: Adding Restaurant Industries"
echo "============================================================================="

if execute_sql_psql "$TEMP_ADD_SQL" "Restaurant Industries Addition"; then
    echo "✅ Restaurant industries added successfully"
else
    echo "❌ Failed to add restaurant industries"
    echo ""
    echo "📋 Manual execution required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/add-restaurant-industries.sql"
    echo "   3. Execute the script"
    echo ""
    read -p "Press Enter after manually executing the SQL script..."
fi

# Execute the verification tests
echo ""
echo "============================================================================="
echo "STEP 2: Verifying Restaurant Industries"
echo "============================================================================="

if execute_sql_psql "$TEMP_TEST_SQL" "Restaurant Industries Verification"; then
    echo "✅ Restaurant industries verification completed"
else
    echo "❌ Failed to verify restaurant industries"
    echo ""
    echo "📋 Manual verification required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/test-restaurant-industries.sql"
    echo "   3. Execute the script and review results"
    echo ""
    read -p "Press Enter after manually executing the verification script..."
fi

# Clean up temporary files
rm -f "$TEMP_ADD_SQL" "$TEMP_TEST_SQL"

echo ""
echo "============================================================================="
echo "SUBTASK 1.2.1 COMPLETION SUMMARY"
echo "============================================================================="
echo "✅ Restaurant industries addition script created"
echo "✅ Verification test script created"
echo "✅ Execution script completed"
echo ""
echo "📋 Next Steps:"
echo "   1. Verify restaurant industries were added to database"
echo "   2. Proceed to Subtask 1.2.2: Add restaurant keywords"
echo "   3. Update COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN.md"
echo ""
echo "🎯 Expected Results:"
echo "   - 12 restaurant industry categories added"
echo "   - Confidence thresholds: 0.70 to 0.85"
echo "   - All industries active and ready for keywords"
echo ""
echo "📁 Files Created:"
echo "   - scripts/add-restaurant-industries.sql"
echo "   - scripts/test-restaurant-industries.sql"
echo "   - scripts/execute-subtask-1-2-1.sh"
echo ""
echo "============================================================================="
echo "Subtask 1.2.1: Add Restaurant Industries - COMPLETED"
echo "============================================================================="
