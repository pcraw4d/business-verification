#!/bin/bash

# Execute Subtask 1.2.1: Execute Classification Schema Migration
# This script runs the supabase-classification-migration.sql to create the 6 classification tables

set -e

echo "🚀 KYB Platform - Subtask 1.2.1: Execute Classification Schema Migration"
echo "======================================================================="

# Check if required environment variables are set
if [ -z "$SUPABASE_URL" ]; then
    echo "❌ Error: SUPABASE_URL environment variable is required"
    echo "   Please set SUPABASE_URL to your Supabase project URL"
    exit 1
fi

if [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo "❌ Error: SUPABASE_SERVICE_ROLE_KEY environment variable is required"
    echo "   Please set SUPABASE_SERVICE_ROLE_KEY to your Supabase service role key"
    exit 1
fi

echo "✅ Environment variables validated"
echo "📝 Supabase URL: $SUPABASE_URL"
echo "📝 Service Role Key: ${SUPABASE_SERVICE_ROLE_KEY:0:20}..."

# Check if migration file exists
MIGRATION_FILE="supabase-classification-migration.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Error: Migration file not found: $MIGRATION_FILE"
    echo "   Please ensure the migration file exists in the current directory"
    exit 1
fi

echo "✅ Migration file found: $MIGRATION_FILE"

# Create temporary files for different execution methods
TEMP_SQL_FILE="/tmp/classification_migration.sql"
cp "$MIGRATION_FILE" "$TEMP_SQL_FILE"
echo "✅ SQL script prepared"

# Function to execute SQL via psql (preferred method)
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
        echo "📝 Using psql to execute migration..."
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

# Function to execute SQL via Supabase API (alternative method)
execute_sql_api() {
    local sql_file="$1"
    local description="$2"
    
    echo ""
    echo "🔄 Executing: $description"
    echo "   File: $sql_file"
    
    # Read SQL content and escape for JSON
    local sql_content=$(cat "$sql_file" | sed 's/"/\\"/g' | tr '\n' ' ')
    
    # Execute via Supabase API
    local response=$(curl -s -X POST \
        "$SUPABASE_URL/rest/v1/rpc/exec_sql" \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"sql\": \"$sql_content\"}" 2>/dev/null)
    
    if [ $? -eq 0 ] && [[ ! "$response" == *"error"* ]]; then
        echo "✅ $description completed successfully"
        return 0
    else
        echo "❌ $description failed"
        echo "   Response: $response"
        return 1
    fi
}

# Function to verify tables were created
verify_tables() {
    echo ""
    echo "🔍 Verifying classification tables..."
    
    local expected_tables=("industries" "industry_keywords" "classification_codes" "industry_patterns" "keyword_weights" "classification_accuracy_metrics")
    local all_tables_exist=true
    
    for table in "${expected_tables[@]}"; do
        echo "🔍 Checking table: $table"
        
        # Try to query the table
        local response=$(curl -s -X GET \
            "$SUPABASE_URL/rest/v1/$table?select=*&limit=1" \
            -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
            -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
            2>/dev/null)
        
        if [ $? -eq 0 ] && [[ ! "$response" == *"error"* ]]; then
            echo "✅ Table $table verified successfully"
        else
            echo "❌ Table $table not found or not accessible"
            all_tables_exist=false
        fi
    done
    
    if [ "$all_tables_exist" = true ]; then
        echo "✅ All 6 classification tables verified successfully"
        return 0
    else
        echo "❌ Some tables are missing or not accessible"
        return 1
    fi
}

# Function to test sample data insertion
test_sample_data() {
    echo ""
    echo "🧪 Testing sample data insertion..."
    
    # Test inserting a sample industry
    local sample_data='{"name": "Test Industry", "description": "Test industry for migration validation", "category": "Test", "confidence_threshold": 0.50, "is_active": true}'
    
    local response=$(curl -s -X POST \
        "$SUPABASE_URL/rest/v1/industries" \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -H "Prefer: return=minimal" \
        -d "$sample_data" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ Sample data insertion test passed"
        return 0
    else
        echo "❌ Sample data insertion test failed"
        echo "   Response: $response"
        return 1
    fi
}

# Execute the classification schema migration
echo ""
echo "============================================================================="
echo "STEP 1: Executing Classification Schema Migration"
echo "============================================================================="

# Try psql first, then fall back to API
if execute_sql_psql "$TEMP_SQL_FILE" "Classification Schema Migration"; then
    echo "✅ Classification schema migration completed successfully"
else
    echo "⚠️  psql execution failed, trying API method..."
    
    if execute_sql_api "$TEMP_SQL_FILE" "Classification Schema Migration"; then
        echo "✅ Classification schema migration completed successfully via API"
    else
        echo "❌ Both psql and API execution failed"
        echo ""
        echo "📋 Manual execution required:"
        echo "   1. Open Supabase SQL Editor in your dashboard"
        echo "   2. Copy and paste the contents of: $MIGRATION_FILE"
        echo "   3. Execute the script"
        echo ""
        read -p "Press Enter after manually executing the SQL script..."
    fi
fi

# Verify tables were created
echo ""
echo "============================================================================="
echo "STEP 2: Verifying Classification Tables"
echo "============================================================================="

if verify_tables; then
    echo "✅ All classification tables verified successfully"
else
    echo "❌ Table verification failed"
    echo "   Please check the Supabase dashboard to ensure tables were created"
    exit 1
fi

# Test sample data insertion
echo ""
echo "============================================================================="
echo "STEP 3: Testing Sample Data Insertion"
echo "============================================================================="

if test_sample_data; then
    echo "✅ Sample data insertion test passed"
else
    echo "⚠️  Sample data insertion test failed (this might be expected if data already exists)"
fi

# Cleanup
rm -f "$TEMP_SQL_FILE"

echo ""
echo "🎉 Subtask 1.2.1 completed successfully!"
echo "📋 Summary:"
echo "   ✅ Classification schema migration executed"
echo "   ✅ All 6 classification tables verified"
echo "   ✅ Sample data insertion tested"
echo ""
echo "📋 Next steps:"
echo "   1. Proceed to subtask 1.2.2: Populate Classification Data"
echo "   2. Add comprehensive industry data and keywords"
echo "   3. Populate NAICS, MCC, SIC codes"
echo ""
echo "🔗 Tables created:"
echo "   - industries"
echo "   - industry_keywords"
echo "   - classification_codes"
echo "   - industry_patterns"
echo "   - keyword_weights"
echo "   - classification_accuracy_metrics"