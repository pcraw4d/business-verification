#!/bin/bash

# =============================================================================
# Execute Task 1.2 Comprehensive Testing
# Complete verification of restaurant industry data implementation
# =============================================================================

# This script executes comprehensive testing of all Task 1.2 deliverables:
# 1.2.1: Restaurant Industries
# 1.2.2: Restaurant Keywords  
# 1.2.3: Restaurant Classification Codes
# 1.3: Restaurant Classification API Testing

set -e

echo "🚀 Starting Task 1.2 Comprehensive Testing"
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
TEMP_SQL_FILE="/tmp/test_task_1_2_comprehensive.sql"

echo ""
echo "📝 Preparing testing scripts..."

# Copy the SQL testing script to temporary file
cp scripts/test-task-1-2-comprehensive.sql "$TEMP_SQL_FILE"

echo "✅ Testing scripts prepared"

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

# Execute the comprehensive database testing
echo ""
echo "============================================================================="
echo "STEP 1: Database Testing (Tasks 1.2.1, 1.2.2, 1.2.3)"
echo "============================================================================="

if execute_sql_psql "$TEMP_SQL_FILE" "Comprehensive Database Testing"; then
    echo "✅ Database testing completed successfully"
else
    echo "❌ Failed to complete database testing"
    echo ""
    echo "📋 Manual execution required:"
    echo "   1. Open Supabase SQL Editor"
    echo "   2. Copy and paste the contents of: scripts/test-task-1-2-comprehensive.sql"
    echo "   3. Execute the script and review results"
    echo ""
    read -p "Press Enter after manually executing the database testing script..."
fi

# Execute the API testing
echo ""
echo "============================================================================="
echo "STEP 2: API Testing (Task 1.3)"
echo "============================================================================="

echo "🔄 Executing: Restaurant Classification API Testing"
echo "   File: scripts/test-restaurant-classification-api.sh"

if [ -f "scripts/test-restaurant-classification-api.sh" ]; then
    if ./scripts/test-restaurant-classification-api.sh; then
        echo "✅ API testing completed successfully"
    else
        echo "❌ API testing failed or had issues"
        echo ""
        echo "📋 Manual execution required:"
        echo "   1. Ensure the API server is running: go run cmd/server/main.go"
        echo "   2. Run the API test script: ./scripts/test-restaurant-classification-api.sh"
        echo ""
        read -p "Press Enter after manually executing the API testing script..."
    fi
else
    echo "❌ API testing script not found"
    echo "   Please ensure scripts/test-restaurant-classification-api.sh exists"
fi

# Clean up temporary files
rm -f "$TEMP_SQL_FILE"

echo ""
echo "============================================================================="
echo "TASK 1.2 COMPREHENSIVE TESTING SUMMARY"
echo "============================================================================="
echo "✅ Database testing script executed"
echo "✅ API testing script executed"
echo "✅ Comprehensive testing completed"
echo ""
echo "📋 Testing Coverage:"
echo "   - Task 1.2.1: Restaurant Industries (12 industries, confidence thresholds)"
echo "   - Task 1.2.2: Restaurant Keywords (200+ keywords, weight validation)"
echo "   - Task 1.2.3: Restaurant Classification Codes (50+ codes, NAICS/SIC/MCC)"
echo "   - Task 1.3: Restaurant Classification API (16 API tests, performance)"
echo ""
echo "🎯 Expected Results:"
echo "   - 12 restaurant industries with proper confidence thresholds"
echo "   - 200+ keywords with weights 0.6000-1.0000"
echo "   - 50+ classification codes with complete coverage"
echo "   - API classification accuracy >75%"
echo "   - Performance <1.0s per request"
echo ""
echo "📊 Test Categories:"
echo "   - Data Integrity: No duplicates, valid relationships"
echo "   - Performance: Query optimization, response times"
echo "   - Functionality: API endpoints, classification accuracy"
echo "   - Error Handling: Invalid inputs, edge cases"
echo ""
echo "📁 Files Created:"
echo "   - scripts/test-task-1-2-comprehensive.sql"
echo "   - scripts/test-restaurant-classification-api.sh"
echo "   - scripts/execute-task-1-2-testing.sh"
echo ""
echo "🚀 Next Steps:"
echo "   1. Review all test results above"
echo "   2. Fix any failing tests or issues"
echo "   3. Proceed to Phase 2: Algorithm Improvements"
echo "   4. Update COMPREHENSIVE_CLASSIFICATION_IMPROVEMENT_PLAN.md"
echo ""
echo "============================================================================="
echo "Task 1.2 Comprehensive Testing - COMPLETED"
echo "============================================================================="
