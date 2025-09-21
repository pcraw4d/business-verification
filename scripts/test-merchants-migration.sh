#!/bin/bash

# Merchants Migration Data Integrity Testing Script
# Subtask 2.2.2: Enhance Merchants Table - Data Integrity Testing
# Date: January 19, 2025
# Purpose: Comprehensive testing of merchants table migration and data integrity

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_LOG="$PROJECT_ROOT/logs/merchants-migration-test-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$TEST_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$TEST_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$TEST_LOG"
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1" | tee -a "$TEST_LOG"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$TEST_LOG")"

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
    log_info "Loaded environment variables from .env"
else
    log_warning ".env file not found, using system environment variables"
fi

# Validate required environment variables
required_vars=("SUPABASE_URL" "SUPABASE_SERVICE_ROLE_KEY")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        log_error "Required environment variable $var is not set"
        exit 1
    fi
done

log_info "Starting merchants migration data integrity testing"
log_info "Test log: $TEST_LOG"
log_info "Supabase URL: $SUPABASE_URL"

# Function to execute SQL and get result
execute_sql() {
    local sql="$1"
    local description="$2"
    
    log_test "Executing: $description"
    
    local response=$(curl -s -X POST \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$sql\"}" \
        "$SUPABASE_URL/rest/v1/rpc/exec" 2>&1)
    
    if [ $? -eq 0 ]; then
        echo "$response"
    else
        log_error "SQL execution failed: $response"
        return 1
    fi
}

# Function to run a test
run_test() {
    local test_name="$1"
    local sql="$2"
    local expected_condition="$3"
    local description="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "Running test: $test_name"
    log_test "Description: $description"
    
    local result=$(execute_sql "$sql" "$test_name")
    local count=$(echo "$result" | grep -o '[0-9]*' | head -1)
    
    if [ -z "$count" ]; then
        count=0
    fi
    
    log_test "Result: $count"
    
    case "$expected_condition" in
        "zero")
            if [ "$count" -eq 0 ]; then
                log_success "PASS: $test_name - Found $count records (expected 0)"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                log_error "FAIL: $test_name - Found $count records (expected 0)"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
            ;;
        "positive")
            if [ "$count" -gt 0 ]; then
                log_success "PASS: $test_name - Found $count records (expected > 0)"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                log_error "FAIL: $test_name - Found $count records (expected > 0)"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
            ;;
        "equal")
            local expected_count="$5"
            if [ "$count" -eq "$expected_count" ]; then
                log_success "PASS: $test_name - Found $count records (expected $expected_count)"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                log_error "FAIL: $test_name - Found $count records (expected $expected_count)"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
            ;;
    esac
    
    echo "---" | tee -a "$TEST_LOG"
}

# Function to get count from table
get_count() {
    local table="$1"
    local where_clause="${2:-}"
    
    local sql="SELECT COUNT(*) as count FROM $table"
    if [ -n "$where_clause" ]; then
        sql="$sql WHERE $where_clause"
    fi
    
    local result=$(execute_sql "$sql" "Get count from $table")
    echo "$result" | grep -o '"count":[0-9]*' | cut -d':' -f2
}

# Test 1: Schema Validation
log_info "=== TEST 1: SCHEMA VALIDATION ==="

# Test 1.1: Check if all required columns exist
run_test "schema_metadata_column" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'metadata';" \
    "positive" \
    "Check if metadata column exists in merchants table"

run_test "schema_website_url_column" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'website_url';" \
    "positive" \
    "Check if website_url column exists in merchants table"

run_test "schema_description_column" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'description';" \
    "positive" \
    "Check if description column exists in merchants table"

run_test "schema_user_id_column" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'user_id';" \
    "positive" \
    "Check if user_id column exists in merchants table"

# Test 1.2: Check column data types
run_test "schema_metadata_type" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'metadata' AND data_type = 'jsonb';" \
    "positive" \
    "Check if metadata column is JSONB type"

run_test "schema_name_length" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'name' AND character_maximum_length = 500;" \
    "positive" \
    "Check if name column has correct length (500)"

run_test "schema_industry_length" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'industry' AND character_maximum_length = 255;" \
    "positive" \
    "Check if industry column has correct length (255)"

run_test "schema_industry_code_length" \
    "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'merchants' AND column_name = 'industry_code' AND character_maximum_length = 50;" \
    "positive" \
    "Check if industry_code column has correct length (50)"

# Test 2: Data Integrity Validation
log_info "=== TEST 2: DATA INTEGRITY VALIDATION ==="

# Test 2.1: Check for duplicate registration numbers
run_test "duplicate_registration_numbers" \
    "SELECT COUNT(*) FROM (SELECT registration_number, COUNT(*) as cnt FROM merchants WHERE registration_number IS NOT NULL AND registration_number != '' GROUP BY registration_number HAVING COUNT(*) > 1) as duplicates;" \
    "zero" \
    "Check for duplicate registration numbers"

# Test 2.2: Check for missing required fields
run_test "missing_names" \
    "SELECT COUNT(*) FROM merchants WHERE name IS NULL OR name = '';" \
    "zero" \
    "Check for missing business names"

run_test "missing_legal_names" \
    "SELECT COUNT(*) FROM merchants WHERE legal_name IS NULL OR legal_name = '';" \
    "zero" \
    "Check for missing legal names"

# Test 2.3: Check for invalid foreign key references
run_test "invalid_portfolio_types" \
    "SELECT COUNT(*) FROM merchants m LEFT JOIN portfolio_types pt ON m.portfolio_type_id = pt.id WHERE m.portfolio_type_id IS NOT NULL AND pt.id IS NULL;" \
    "zero" \
    "Check for invalid portfolio type references"

run_test "invalid_risk_levels" \
    "SELECT COUNT(*) FROM merchants m LEFT JOIN risk_levels rl ON m.risk_level_id = rl.id WHERE m.risk_level_id IS NOT NULL AND rl.id IS NULL;" \
    "zero" \
    "Check for invalid risk level references"

run_test "invalid_created_by" \
    "SELECT COUNT(*) FROM merchants m LEFT JOIN users u ON m.created_by = u.id WHERE m.created_by IS NOT NULL AND u.id IS NULL;" \
    "zero" \
    "Check for invalid created_by user references"

# Test 3: Data Migration Validation
log_info "=== TEST 3: DATA MIGRATION VALIDATION ==="

# Get original businesses count for comparison
businesses_count=$(get_count "businesses")
log_info "Original businesses count: $businesses_count"

merchants_count=$(get_count "merchants")
log_info "Current merchants count: $merchants_count"

# Test 3.1: Check if all businesses were migrated
if [ "$businesses_count" -gt 0 ]; then
    run_test "migration_completeness" \
        "SELECT COUNT(*) FROM merchants;" \
        "equal" \
        "Check if all businesses were migrated to merchants" \
        "$businesses_count"
else
    log_warning "No businesses to migrate, skipping migration completeness test"
fi

# Test 3.2: Check data consistency between businesses and merchants
if [ "$businesses_count" -gt 0 ]; then
    run_test "data_consistency_names" \
        "SELECT COUNT(*) FROM businesses b WHERE NOT EXISTS (SELECT 1 FROM merchants m WHERE m.name = b.name);" \
        "zero" \
        "Check if all business names were migrated correctly"
    
    run_test "data_consistency_industries" \
        "SELECT COUNT(*) FROM businesses b JOIN merchants m ON b.name = m.name WHERE b.industry IS NOT NULL AND m.industry IS NULL;" \
        "zero" \
        "Check if industry data was migrated correctly"
fi

# Test 4: Performance and Index Validation
log_info "=== TEST 4: PERFORMANCE AND INDEX VALIDATION ==="

# Test 4.1: Check if required indexes exist
run_test "index_metadata" \
    "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'merchants' AND indexname = 'idx_merchants_metadata';" \
    "positive" \
    "Check if metadata GIN index exists"

run_test "index_website_url" \
    "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'merchants' AND indexname = 'idx_merchants_website_url';" \
    "positive" \
    "Check if website_url index exists"

run_test "index_portfolio_type" \
    "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'merchants' AND indexname = 'idx_merchants_portfolio_type';" \
    "positive" \
    "Check if portfolio_type index exists"

run_test "index_risk_level" \
    "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'merchants' AND indexname = 'idx_merchants_risk_level';" \
    "positive" \
    "Check if risk_level index exists"

# Test 5: Constraint Validation
log_info "=== TEST 5: CONSTRAINT VALIDATION ==="

# Test 5.1: Check NOT NULL constraints
run_test "constraint_registration_number_not_null" \
    "SELECT COUNT(*) FROM information_schema.check_constraints WHERE constraint_name LIKE '%merchants%' AND check_clause LIKE '%registration_number%';" \
    "positive" \
    "Check if registration_number NOT NULL constraint exists"

run_test "constraint_legal_name_not_null" \
    "SELECT COUNT(*) FROM information_schema.check_constraints WHERE constraint_name LIKE '%merchants%' AND check_clause LIKE '%legal_name%';" \
    "positive" \
    "Check if legal_name NOT NULL constraint exists"

# Test 5.2: Check foreign key constraints
run_test "constraint_portfolio_type_fk" \
    "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_name = 'merchants' AND constraint_type = 'FOREIGN KEY' AND constraint_name LIKE '%portfolio_type%';" \
    "positive" \
    "Check if portfolio_type foreign key constraint exists"

run_test "constraint_risk_level_fk" \
    "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_name = 'merchants' AND constraint_type = 'FOREIGN KEY' AND constraint_name LIKE '%risk_level%';" \
    "positive" \
    "Check if risk_level foreign key constraint exists"

# Test 6: Business Logic Validation
log_info "=== TEST 6: BUSINESS LOGIC VALIDATION ==="

# Test 6.1: Check portfolio type distribution
run_test "portfolio_type_distribution" \
    "SELECT COUNT(*) FROM merchants m JOIN portfolio_types pt ON m.portfolio_type_id = pt.id WHERE pt.name IN ('prospective', 'onboarded', 'deactivated', 'pending');" \
    "positive" \
    "Check if merchants have valid portfolio types"

# Test 6.2: Check risk level distribution
run_test "risk_level_distribution" \
    "SELECT COUNT(*) FROM merchants m JOIN risk_levels rl ON m.risk_level_id = rl.id WHERE rl.name IN ('low', 'medium', 'high', 'critical');" \
    "positive" \
    "Check if merchants have valid risk levels"

# Test 6.3: Check compliance status values
run_test "compliance_status_values" \
    "SELECT COUNT(*) FROM merchants WHERE compliance_status IN ('pending', 'approved', 'rejected', 'under_review');" \
    "positive" \
    "Check if merchants have valid compliance status values"

# Test 6.4: Check status values
run_test "status_values" \
    "SELECT COUNT(*) FROM merchants WHERE status IN ('active', 'inactive', 'suspended', 'pending');" \
    "positive" \
    "Check if merchants have valid status values"

# Test 7: Data Quality Validation
log_info "=== TEST 7: DATA QUALITY VALIDATION ==="

# Test 7.1: Check for valid email formats
run_test "valid_email_formats" \
    "SELECT COUNT(*) FROM merchants WHERE contact_email IS NOT NULL AND contact_email !~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$';" \
    "zero" \
    "Check for valid email formats"

# Test 7.2: Check for valid website URLs
run_test "valid_website_urls" \
    "SELECT COUNT(*) FROM merchants WHERE website_url IS NOT NULL AND website_url !~ '^https?://[^\\s]+$';" \
    "zero" \
    "Check for valid website URL formats"

# Test 7.3: Check for reasonable employee counts
run_test "reasonable_employee_counts" \
    "SELECT COUNT(*) FROM merchants WHERE employee_count IS NOT NULL AND (employee_count < 0 OR employee_count > 1000000);" \
    "zero" \
    "Check for reasonable employee counts (0-1,000,000)"

# Test 7.4: Check for reasonable annual revenue
run_test "reasonable_annual_revenue" \
    "SELECT COUNT(*) FROM merchants WHERE annual_revenue IS NOT NULL AND (annual_revenue < 0 OR annual_revenue > 999999999999.99);" \
    "zero" \
    "Check for reasonable annual revenue values"

# Test 8: Migration Function Validation
log_info "=== TEST 8: MIGRATION FUNCTION VALIDATION ==="

# Test 8.1: Check if migration functions exist
run_test "migration_function_exists" \
    "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'migrate_businesses_to_merchants' AND routine_type = 'FUNCTION';" \
    "positive" \
    "Check if migrate_businesses_to_merchants function exists"

run_test "validation_function_exists" \
    "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'validate_merchants_migration' AND routine_type = 'FUNCTION';" \
    "positive" \
    "Check if validate_merchants_migration function exists"

run_test "rollback_function_exists" \
    "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'rollback_merchants_enhancement' AND routine_type = 'FUNCTION';" \
    "positive" \
    "Check if rollback_merchants_enhancement function exists"

# Test 9: Performance Testing
log_info "=== TEST 9: PERFORMANCE TESTING ==="

# Test 9.1: Test query performance on new indexes
log_test "Testing query performance on metadata field"
start_time=$(date +%s%N)
execute_sql "SELECT COUNT(*) FROM merchants WHERE metadata ? 'industry';" "Performance test: metadata query"
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))
log_info "Metadata query took ${duration}ms"

log_test "Testing query performance on website_url field"
start_time=$(date +%s%N)
execute_sql "SELECT COUNT(*) FROM merchants WHERE website_url IS NOT NULL;" "Performance test: website_url query"
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 ))
log_info "Website URL query took ${duration}ms"

# Test Summary
log_info "=== TEST SUMMARY ==="
log_info "Total tests run: $TOTAL_TESTS"
log_success "Passed tests: $PASSED_TESTS"
if [ $FAILED_TESTS -gt 0 ]; then
    log_error "Failed tests: $FAILED_TESTS"
else
    log_success "Failed tests: $FAILED_TESTS"
fi

# Calculate success rate
if [ $TOTAL_TESTS -gt 0 ]; then
    success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    log_info "Success rate: $success_rate%"
    
    if [ $success_rate -ge 95 ]; then
        log_success "EXCELLENT: Migration validation passed with $success_rate% success rate"
        exit 0
    elif [ $success_rate -ge 80 ]; then
        log_warning "GOOD: Migration validation passed with $success_rate% success rate"
        exit 0
    else
        log_error "POOR: Migration validation failed with only $success_rate% success rate"
        exit 1
    fi
else
    log_error "No tests were run"
    exit 1
fi
