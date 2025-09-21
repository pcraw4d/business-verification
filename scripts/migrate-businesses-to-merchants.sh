#!/bin/bash

# Business to Merchants Migration Script
# Subtask 2.2.2: Enhance Merchants Table - Data Migration
# Date: January 19, 2025
# Purpose: Migrate data from businesses table to enhanced merchants table

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATION_LOG="$PROJECT_ROOT/logs/business-migration-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$MIGRATION_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$MIGRATION_LOG"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$MIGRATION_LOG")"

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

log_info "Starting business to merchants migration"
log_info "Migration log: $MIGRATION_LOG"
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
        echo "$response" >> "$MIGRATION_LOG"
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

# Pre-migration validation
log_info "=== PRE-MIGRATION VALIDATION ==="

# Check if businesses table exists and has data
businesses_count=$(get_count "businesses")
log_info "Businesses table count: $businesses_count"

if [ "$businesses_count" -eq 0 ]; then
    log_warning "No data in businesses table to migrate"
    exit 0
fi

# Check if merchants table exists
merchants_count=$(get_count "merchants")
log_info "Current merchants table count: $merchants_count"

# Check if portfolio_types table exists and has data
portfolio_types_count=$(get_count "portfolio_types")
log_info "Portfolio types count: $portfolio_types_count"

if [ "$portfolio_types_count" -eq 0 ]; then
    log_warning "No portfolio types found, creating default portfolio types"
    
    # Create default portfolio types
    execute_sql "INSERT INTO portfolio_types (name, description, is_active) VALUES 
        ('prospective', 'Prospective merchants under evaluation', true),
        ('onboarded', 'Successfully onboarded merchants', true),
        ('deactivated', 'Deactivated merchants', true),
        ('pending', 'Pending approval merchants', true)
        ON CONFLICT (name) DO NOTHING;" "Create default portfolio types"
fi

# Check if risk_levels table exists and has data
risk_levels_count=$(get_count "risk_levels")
log_info "Risk levels count: $risk_levels_count"

if [ "$risk_levels_count" -eq 0 ]; then
    log_warning "No risk levels found, creating default risk levels"
    
    # Create default risk levels
    execute_sql "INSERT INTO risk_levels (name, description, level, is_active) VALUES 
        ('low', 'Low risk merchants', 1, true),
        ('medium', 'Medium risk merchants', 2, true),
        ('high', 'High risk merchants', 3, true),
        ('critical', 'Critical risk merchants', 4, true)
        ON CONFLICT (name) DO NOTHING;" "Create default risk levels"
fi

# Step 1: Run the merchants table enhancement migration
log_info "=== STEP 1: ENHANCE MERCHANTS TABLE ==="

# Read and execute the enhancement migration
if [ -f "$PROJECT_ROOT/internal/database/migrations/008_enhance_merchants_table.sql" ]; then
    log_info "Executing merchants table enhancement migration"
    
    # Read the SQL file and execute it
    sql_content=$(cat "$PROJECT_ROOT/internal/database/migrations/008_enhance_merchants_table.sql")
    execute_sql "$sql_content" "Enhance merchants table with missing fields"
else
    log_error "Migration file not found: $PROJECT_ROOT/internal/database/migrations/008_enhance_merchants_table.sql"
    exit 1
fi

# Step 2: Execute data migration
log_info "=== STEP 2: MIGRATE DATA FROM BUSINESSES TO MERCHANTS ==="

# Execute the migration function
migration_result=$(curl -s -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -d '{"query": "SELECT migrate_businesses_to_merchants();"}' \
    "$SUPABASE_URL/rest/v1/rpc/exec")

log_info "Migration result: $migration_result"

# Extract the number of migrated records
migrated_count=$(echo "$migration_result" | grep -o '[0-9]*' | head -1)
log_success "Migrated $migrated_count records from businesses to merchants"

# Step 3: Validate migration
log_info "=== STEP 3: VALIDATE MIGRATION ==="

# Get new merchants count
new_merchants_count=$(get_count "merchants")
log_info "New merchants table count: $new_merchants_count"

# Execute validation function
validation_result=$(curl -s -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -d '{"query": "SELECT * FROM validate_merchants_migration();"}' \
    "$SUPABASE_URL/rest/v1/rpc/exec")

log_info "Validation results:"
echo "$validation_result" | tee -a "$MIGRATION_LOG"

# Check for validation failures
if echo "$validation_result" | grep -q '"status":"FAIL"'; then
    log_error "Migration validation failed. Check the validation results above."
    log_error "Consider running rollback if needed: SELECT rollback_merchants_enhancement();"
    exit 1
else
    log_success "Migration validation passed"
fi

# Step 4: Update foreign key constraints
log_info "=== STEP 4: UPDATE FOREIGN KEY CONSTRAINTS ==="

# Restore NOT NULL constraints after successful migration
execute_sql "ALTER TABLE merchants ALTER COLUMN registration_number SET NOT NULL;" "Restore registration_number NOT NULL constraint"
execute_sql "ALTER TABLE merchants ALTER COLUMN legal_name SET NOT NULL;" "Restore legal_name NOT NULL constraint"
execute_sql "ALTER TABLE merchants ALTER COLUMN portfolio_type_id SET NOT NULL;" "Restore portfolio_type_id NOT NULL constraint"
execute_sql "ALTER TABLE merchants ALTER COLUMN risk_level_id SET NOT NULL;" "Restore risk_level_id NOT NULL constraint"
execute_sql "ALTER TABLE merchants ALTER COLUMN created_by SET NOT NULL;" "Restore created_by NOT NULL constraint"

# Add foreign key constraints if they don't exist
execute_sql "ALTER TABLE merchants ADD CONSTRAINT fk_merchants_portfolio_type 
    FOREIGN KEY (portfolio_type_id) REFERENCES portfolio_types(id) ON DELETE RESTRICT;" "Add portfolio_type foreign key constraint"

execute_sql "ALTER TABLE merchants ADD CONSTRAINT fk_merchants_risk_level 
    FOREIGN KEY (risk_level_id) REFERENCES risk_levels(id) ON DELETE RESTRICT;" "Add risk_level foreign key constraint"

execute_sql "ALTER TABLE merchants ADD CONSTRAINT fk_merchants_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT;" "Add created_by foreign key constraint"

# Step 5: Create additional indexes for performance
log_info "=== STEP 5: CREATE PERFORMANCE INDEXES ==="

execute_sql "CREATE INDEX IF NOT EXISTS idx_merchants_portfolio_type ON merchants(portfolio_type_id);" "Create portfolio_type index"
execute_sql "CREATE INDEX IF NOT EXISTS idx_merchants_risk_level ON merchants(risk_level_id);" "Create risk_level index"
execute_sql "CREATE INDEX IF NOT EXISTS idx_merchants_created_by ON merchants(created_by);" "Create created_by index"
execute_sql "CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);" "Create status index"
execute_sql "CREATE INDEX IF NOT EXISTS idx_merchants_compliance_status ON merchants(compliance_status);" "Create compliance_status index"

# Step 6: Final validation
log_info "=== STEP 6: FINAL VALIDATION ==="

# Run final validation
final_validation=$(curl -s -X POST \
    -H "apikey: $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY" \
    -H "Content-Type: application/json" \
    -d '{"query": "SELECT * FROM validate_merchants_migration();"}' \
    "$SUPABASE_URL/rest/v1/rpc/exec")

log_info "Final validation results:"
echo "$final_validation" | tee -a "$MIGRATION_LOG"

# Migration summary
log_info "=== MIGRATION SUMMARY ==="
log_success "Migration completed successfully!"
log_info "Original businesses count: $businesses_count"
log_info "Final merchants count: $new_merchants_count"
log_info "Records migrated: $migrated_count"
log_info "Migration log saved to: $MIGRATION_LOG"

# Cleanup: Remove user_id column after successful migration (optional)
read -p "Do you want to remove the temporary user_id column from merchants table? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Removing temporary user_id column"
    execute_sql "ALTER TABLE merchants DROP COLUMN IF EXISTS user_id;" "Remove temporary user_id column"
    execute_sql "DROP INDEX IF EXISTS idx_merchants_user_id;" "Remove user_id index"
    log_success "Temporary user_id column removed"
fi

log_success "Business to merchants migration completed successfully!"
log_info "Next steps:"
log_info "1. Update application code to use merchants table instead of businesses table"
log_info "2. Test all business-related functionality"
log_info "3. Remove businesses table after confirming everything works (Task 2.2.4)"

exit 0
