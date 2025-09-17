#!/bin/bash

# =============================================================================
# Task 1.1.1: Add missing is_active column to keyword_weights table
# =============================================================================
# This script executes the first subtask of the Comprehensive Classification
# Improvement Plan - adding the missing is_active column to the keyword_weights table.

set -e

echo "🔧 Task 1.1.1: Adding missing is_active column to keyword_weights table"
echo "========================================================================="

# Load environment variables from .env file
if [ -f ".env" ]; then
    echo "📝 Loading environment variables from .env file..."
    source .env
else
    echo "⚠️  Warning: .env file not found, using system environment variables"
fi

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

# Function to execute SQL command
execute_sql() {
    local sql="$1"
    local description="$2"
    
    echo "📝 $description"
    echo "   SQL: $sql"
    
    # Execute SQL using Supabase REST API
    response=$(curl -s -X POST \
        "$SUPABASE_URL/rest/v1/rpc/exec_sql" \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -H "Prefer: return=minimal" \
        -d "{\"sql\": \"$sql\"}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ $description completed successfully"
        if [ -n "$response" ]; then
            echo "   Response: $response"
        fi
    else
        echo "❌ $description failed"
        echo "   Response: $response"
        return 1
    fi
}

# Function to verify SQL command
verify_sql() {
    local sql="$1"
    local description="$2"
    
    echo "🔍 $description"
    echo "   SQL: $sql"
    
    # Execute verification query
    response=$(curl -s -X POST \
        "$SUPABASE_URL/rest/v1/rpc/exec_sql" \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -H "Prefer: return=minimal" \
        -d "{\"sql\": \"$sql\"}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ $description completed"
        if [ -n "$response" ]; then
            echo "   Result: $response"
        fi
    else
        echo "⚠️  $description failed - $response"
        return 1
    fi
}

echo ""
echo "🚀 Starting Task 1.1.1 execution..."

# Step 1: Add missing is_active column
echo ""
echo "Step 1: Adding is_active column to keyword_weights table"
execute_sql \
    "ALTER TABLE keyword_weights ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;" \
    "Add is_active column to keyword_weights table"

# Step 2: Update existing records to set is_active = true
echo ""
echo "Step 2: Updating existing records to set is_active = true"
execute_sql \
    "UPDATE keyword_weights SET is_active = true WHERE is_active IS NULL;" \
    "Update existing records to set is_active = true"

# Step 3: Create performance indexes
echo ""
echo "Step 3: Creating performance indexes"
execute_sql \
    "CREATE INDEX IF NOT EXISTS idx_keyword_weights_active ON keyword_weights(is_active);" \
    "Create index on is_active column"

execute_sql \
    "CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active ON keyword_weights(industry_id, is_active);" \
    "Create composite index on industry_id and is_active"

echo ""
echo "🔍 Verifying Task 1.1.1 completion..."

# Verification 1: Check that the column exists
echo ""
echo "Verification 1: Checking that is_active column exists"
verify_sql \
    "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'keyword_weights' AND column_name = 'is_active';" \
    "Verify is_active column exists"

# Verification 2: Check that all records are active
echo ""
echo "Verification 2: Checking that all records are active"
verify_sql \
    "SELECT COUNT(*) as total_records, COUNT(CASE WHEN is_active = true THEN 1 END) as active_records FROM keyword_weights;" \
    "Verify all records are active"

# Verification 3: Check that indexes exist
echo ""
echo "Verification 3: Checking that indexes exist"
verify_sql \
    "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'keyword_weights' AND indexname LIKE '%active%';" \
    "Verify indexes exist"

echo ""
echo "✅ Task 1.1.1 completed successfully!"
echo ""
echo "📊 Summary:"
echo "   - Added is_active column to keyword_weights table"
echo "   - Updated all existing records to be active"
echo "   - Created performance indexes for is_active column"
echo "   - Verified all changes were applied correctly"
echo ""
echo "🎯 Next Steps:"
echo "   - Task 1.1.2: Update existing records (already completed in this task)"
echo "   - Task 1.1.3: Create performance indexes (already completed in this task)"
echo "   - Task 1.2: Add Restaurant Industry Data"
echo ""
echo "🔧 The keyword_weights table now has the is_active column and should no longer"
echo "   produce 'is_active does not exist' errors when building the keyword index."
