#!/bin/bash

# Drop Businesses Table Script
# Subtask 2.2.4: Remove Redundant Tables
# Date: January 19, 2025
# Purpose: Safely drop the businesses table after successful migration to merchants table

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DROP_LOG="$PROJECT_ROOT/logs/businesses-table-drop-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$DROP_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$DROP_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$DROP_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$DROP_LOG"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$DROP_LOG")"

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

log_info "Starting businesses table removal process"
log_info "Drop log: $DROP_LOG"
log_info "Supabase URL: $SUPABASE_URL"

# Function to execute SQL via Supabase API
execute_sql() {
    local sql="$1"
    local description="$2"
    
    log_info "Executing: $description"
    
    local response=$(curl -s -X POST \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -H "Prefer: return=minimal" \
        -d "{\"query\": \"$sql\"}" \
        "$SUPABASE_URL/rest/v1/rpc/exec" 2>&1)
    
    if [ $? -eq 0 ]; then
        log_success "$description completed successfully"
        echo "$response" >> "$DROP_LOG"
    else
        log_error "$description failed: $response"
        return 1
    fi
}

# Function to get count from table
get_count() {
    local table="$1"
    local sql="SELECT COUNT(*) as count FROM $table"
    
    local response=$(curl -s -X POST \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$sql\"}" \
        "$SUPABASE_URL/rest/v1/rpc/exec")
    
    echo "$response" | grep -o '"count":[0-9]*' | cut -d':' -f2
}

# Function to check if table exists
table_exists() {
    local table="$1"
    local sql="SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = '$table'
    );"
    
    local response=$(curl -s -X POST \
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$sql\"}" \
        "$SUPABASE_URL/rest/v1/rpc/exec")
    
    echo "$response" | grep -o '"exists":[a-z]*' | cut -d':' -f2
}

# Pre-removal validation
log_info "=== PRE-REMOVAL VALIDATION ==="

# Check if businesses table exists
businesses_exists=$(table_exists "businesses")
log_info "Businesses table exists: $businesses_exists"

if [ "$businesses_exists" != "true" ]; then
    log_warning "Businesses table does not exist. Nothing to remove."
    exit 0
fi

# Check if merchants table exists and has data
merchants_exists=$(table_exists "merchants")
log_info "Merchants table exists: $merchants_exists"

if [ "$merchants_exists" != "true" ]; then
    log_error "Merchants table does not exist. Cannot safely remove businesses table."
    exit 1
fi

# Get counts for comparison
businesses_count=$(get_count "businesses")
merchants_count=$(get_count "merchants")

log_info "Businesses table count: $businesses_count"
log_info "Merchants table count: $merchants_count"

# Validate that migration was successful
if [ "$businesses_count" -gt 0 ] && [ "$merchants_count" -eq 0 ]; then
    log_error "Businesses table has data but merchants table is empty. Migration may have failed."
    log_error "Please run the migration script first: scripts/migrate-businesses-to-merchants.sh"
    exit 1
fi

# Check for any remaining references to businesses table
log_info "Checking for remaining references to businesses table..."

# Check for foreign key references
fk_check_sql="SELECT 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
    AND ccu.table_name = 'businesses';"

fk_references=$(curl -s -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"query\": \"$fk_check_sql\"}" \
    "$SUPABASE_URL/rest/v1/rpc/exec")

if echo "$fk_references" | grep -q "businesses"; then
    log_warning "Found foreign key references to businesses table:"
    echo "$fk_references" | tee -a "$DROP_LOG"
    log_warning "These references need to be updated before dropping the businesses table."
    exit 1
else
    log_success "No foreign key references to businesses table found"
fi

# Final confirmation
log_warning "=== FINAL CONFIRMATION ==="
log_warning "You are about to permanently drop the businesses table."
log_warning "This action cannot be undone."
log_warning "Make sure you have:"
log_warning "1. Completed the migration from businesses to merchants table"
log_warning "2. Updated all application code to use merchants table"
log_warning "3. Tested all business-related functionality"
log_warning "4. Created a backup of the businesses table (if needed)"

read -p "Are you sure you want to drop the businesses table? (yes/NO): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    log_info "Operation cancelled by user"
    exit 0
fi

# Step 1: Create backup of businesses table (optional)
log_info "=== STEP 1: CREATE BACKUP (OPTIONAL) ==="
read -p "Do you want to create a backup of the businesses table before dropping it? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Creating backup of businesses table..."
    
    # Create backup table
    execute_sql "CREATE TABLE businesses_backup_$(date +%Y%m%d_%H%M%S) AS SELECT * FROM businesses;" "Create backup of businesses table"
    
    # Verify backup
    backup_count=$(get_count "businesses_backup_$(date +%Y%m%d_%H%M%S)")
    log_success "Backup created with $backup_count records"
fi

# Step 2: Drop businesses table
log_info "=== STEP 2: DROP BUSINESSES TABLE ==="

# Drop the businesses table
execute_sql "DROP TABLE IF EXISTS businesses CASCADE;" "Drop businesses table"

# Verify table is dropped
businesses_exists_after=$(table_exists "businesses")
if [ "$businesses_exists_after" = "false" ]; then
    log_success "Businesses table successfully dropped"
else
    log_error "Failed to drop businesses table"
    exit 1
fi

# Step 3: Clean up any remaining references
log_info "=== STEP 3: CLEAN UP REFERENCES ==="

# Check for any remaining references in application code
log_info "Checking for remaining references in application code..."

# Search for businesses table references in Go files
businesses_refs=$(find "$PROJECT_ROOT" -name "*.go" -type f -exec grep -l "businesses" {} \; 2>/dev/null || true)

if [ -n "$businesses_refs" ]; then
    log_warning "Found references to 'businesses' in the following Go files:"
    echo "$businesses_refs" | tee -a "$DROP_LOG"
    log_warning "Please review and update these files to use 'merchants' table instead."
else
    log_success "No references to 'businesses' table found in Go files"
fi

# Search for businesses table references in SQL files
businesses_sql_refs=$(find "$PROJECT_ROOT" -name "*.sql" -type f -exec grep -l "businesses" {} \; 2>/dev/null || true)

if [ -n "$businesses_sql_refs" ]; then
    log_warning "Found references to 'businesses' in the following SQL files:"
    echo "$businesses_sql_refs" | tee -a "$DROP_LOG"
    log_warning "Please review and update these files to use 'merchants' table instead."
else
    log_success "No references to 'businesses' table found in SQL files"
fi

# Step 4: Final validation
log_info "=== STEP 4: FINAL VALIDATION ==="

# Verify merchants table is still intact
merchants_count_after=$(get_count "merchants")
log_info "Merchants table count after businesses table removal: $merchants_count_after"

if [ "$merchants_count_after" -eq "$merchants_count" ]; then
    log_success "Merchants table data integrity maintained"
else
    log_error "Merchants table data may have been affected"
    exit 1
fi

# Test a simple query on merchants table
test_query_result=$(curl -s -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -d '{"query": "SELECT COUNT(*) as count FROM merchants LIMIT 1;"}' \
    "$SUPABASE_URL/rest/v1/rpc/exec")

if echo "$test_query_result" | grep -q '"count"'; then
    log_success "Merchants table is accessible and functional"
else
    log_error "Merchants table may not be accessible"
    exit 1
fi

# Step 5: Update documentation
log_info "=== STEP 5: UPDATE DOCUMENTATION ==="

# Create a summary of the removal
log_info "Creating removal summary..."

cat >> "$DROP_LOG" << EOF

=== BUSINESSES TABLE REMOVAL SUMMARY ===
Date: $(date)
Action: Dropped businesses table
Reason: Table consolidation - data migrated to merchants table
Records in businesses table before removal: $businesses_count
Records in merchants table after removal: $merchants_count_after
Backup created: $([ -n "$backup_count" ] && echo "Yes ($backup_count records)" || echo "No")
Status: SUCCESS

=== NEXT STEPS ===
1. Update any remaining application code references
2. Update documentation to reflect the new table structure
3. Test all business-related functionality
4. Monitor system performance and functionality

=== FILES TO REVIEW ===
Go files with 'businesses' references:
$businesses_refs

SQL files with 'businesses' references:
$businesses_sql_refs

EOF

# Final summary
log_info "=== REMOVAL SUMMARY ==="
log_success "Businesses table removal completed successfully!"
log_info "Original businesses count: $businesses_count"
log_info "Final merchants count: $merchants_count_after"
log_info "Removal log saved to: $DROP_LOG"

log_info "Next steps:"
log_info "1. Review and update any remaining code references"
log_info "2. Test all business-related functionality"
log_info "3. Update documentation"
log_info "4. Monitor system performance"

log_success "Businesses table removal completed successfully!"

exit 0
