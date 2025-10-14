#!/bin/bash

# Risk Assessment Service Database Migration Runner
# This script applies database migrations in the correct order with proper error handling

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/supabase-migrations"
LOG_FILE="$PROJECT_ROOT/logs/migration-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Default values
ENVIRONMENT="development"
DRY_RUN=false
ROLLBACK=false
VERBOSE=false
FORCE=false

# Migration files in order
MIGRATIONS=(
    "risk-assessment-schema.sql"
    "risk-assessment-indexes.sql"
    "risk-assessment-rls.sql"
)

# Help function
show_help() {
    cat << EOF
Risk Assessment Service Database Migration Runner

Usage: $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Environment (development, staging, production) [default: development]
    -d, --dry-run           Show what would be executed without running
    -r, --rollback          Rollback the last migration
    -v, --verbose           Verbose output
    -f, --force             Force migration even if errors occur
    -h, --help              Show this help message

ENVIRONMENT VARIABLES:
    DATABASE_URL            PostgreSQL connection string
    SUPABASE_URL            Supabase project URL
    SUPABASE_SERVICE_ROLE_KEY Supabase service role key

EXAMPLES:
    $0 --environment production
    $0 --dry-run --verbose
    $0 --rollback --environment staging

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -r|--rollback)
                ROLLBACK=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate environment
validate_environment() {
    case $ENVIRONMENT in
        development|staging|production)
            log_info "Environment: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT. Must be one of: development, staging, production"
            exit 1
            ;;
    esac
}

# Get database connection parameters
get_db_config() {
    if [[ -n "${DATABASE_URL:-}" ]]; then
        log_info "Using DATABASE_URL from environment"
        DB_URL="$DATABASE_URL"
    elif [[ -n "${SUPABASE_URL:-}" && -n "${SUPABASE_SERVICE_ROLE_KEY:-}" ]]; then
        log_info "Using Supabase configuration"
        # Extract database URL from Supabase URL
        SUPABASE_PROJECT_ID=$(echo "$SUPABASE_URL" | sed 's/.*\/\/\([^.]*\)\..*/\1/')
        DB_URL="postgresql://postgres:${SUPABASE_SERVICE_ROLE_KEY}@db.${SUPABASE_PROJECT_ID}.supabase.co:5432/postgres"
    else
        log_error "Database configuration not found. Please set DATABASE_URL or SUPABASE_URL + SUPABASE_SERVICE_ROLE_KEY"
        exit 1
    fi
}

# Test database connection
test_db_connection() {
    log_info "Testing database connection..."
    
    if $DRY_RUN; then
        log_warning "Dry run mode - skipping connection test"
        return 0
    fi
    
    if psql "$DB_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "Database connection successful"
    else
        log_error "Failed to connect to database"
        exit 1
    fi
}

# Create migration tracking table
create_migration_table() {
    log_info "Creating migration tracking table..."
    
    local sql="
    CREATE TABLE IF NOT EXISTS schema_migrations (
        id SERIAL PRIMARY KEY,
        filename VARCHAR(255) NOT NULL UNIQUE,
        applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        checksum VARCHAR(64),
        environment VARCHAR(50) NOT NULL,
        applied_by VARCHAR(100) DEFAULT current_user
    );
    "
    
    if $DRY_RUN; then
        log_info "Would execute: $sql"
        return 0
    fi
    
    if psql "$DB_URL" -c "$sql" >> "$LOG_FILE" 2>&1; then
        log_success "Migration tracking table created/verified"
    else
        log_error "Failed to create migration tracking table"
        exit 1
    fi
}

# Calculate file checksum
calculate_checksum() {
    local file="$1"
    if command -v sha256sum > /dev/null; then
        sha256sum "$file" | cut -d' ' -f1
    elif command -v shasum > /dev/null; then
        shasum -a 256 "$file" | cut -d' ' -f1
    else
        log_warning "No checksum tool available, skipping checksum validation"
        echo "no-checksum"
    fi
}

# Check if migration was already applied
is_migration_applied() {
    local filename="$1"
    local count
    
    count=$(psql "$DB_URL" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE filename = '$filename' AND environment = '$ENVIRONMENT';" 2>/dev/null | tr -d ' ')
    
    if [[ "$count" -gt 0 ]]; then
        return 0  # Migration already applied
    else
        return 1  # Migration not applied
    fi
}

# Apply a single migration
apply_migration() {
    local migration_file="$1"
    local migration_path="$MIGRATIONS_DIR/$migration_file"
    
    if [[ ! -f "$migration_path" ]]; then
        log_error "Migration file not found: $migration_path"
        return 1
    fi
    
    log_info "Applying migration: $migration_file"
    
    # Check if migration was already applied
    if is_migration_applied "$migration_file"; then
        log_warning "Migration $migration_file already applied, skipping"
        return 0
    fi
    
    # Calculate checksum
    local checksum
    checksum=$(calculate_checksum "$migration_path")
    
    if $DRY_RUN; then
        log_info "Would apply migration: $migration_file"
        log_info "Checksum: $checksum"
        return 0
    fi
    
    # Start transaction
    local temp_sql="/tmp/migration_$$.sql"
    cat > "$temp_sql" << EOF
BEGIN;

-- Apply migration
\i $migration_path

-- Record migration
INSERT INTO schema_migrations (filename, checksum, environment, applied_by)
VALUES ('$migration_file', '$checksum', '$ENVIRONMENT', current_user);

COMMIT;
EOF
    
    # Apply migration
    if psql "$DB_URL" -f "$temp_sql" >> "$LOG_FILE" 2>&1; then
        log_success "Migration $migration_file applied successfully"
        rm -f "$temp_sql"
        return 0
    else
        log_error "Failed to apply migration: $migration_file"
        rm -f "$temp_sql"
        
        if ! $FORCE; then
            log_error "Migration failed. Use --force to continue or check logs: $LOG_FILE"
            exit 1
        else
            log_warning "Migration failed but continuing due to --force flag"
            return 1
        fi
    fi
}

# Rollback last migration
rollback_migration() {
    log_info "Rolling back last migration..."
    
    if $DRY_RUN; then
        log_warning "Dry run mode - would rollback last migration"
        return 0
    fi
    
    # Get last applied migration
    local last_migration
    last_migration=$(psql "$DB_URL" -t -c "SELECT filename FROM schema_migrations WHERE environment = '$ENVIRONMENT' ORDER BY applied_at DESC LIMIT 1;" 2>/dev/null | tr -d ' ')
    
    if [[ -z "$last_migration" ]]; then
        log_warning "No migrations to rollback"
        return 0
    fi
    
    log_warning "Rollback functionality not implemented for safety. Manual rollback required."
    log_info "Last applied migration: $last_migration"
    log_info "To rollback manually:"
    log_info "1. Connect to database: psql '$DB_URL'"
    log_info "2. Review migration: $MIGRATIONS_DIR/$last_migration"
    log_info "3. Apply reverse operations manually"
    log_info "4. Remove from tracking: DELETE FROM schema_migrations WHERE filename = '$last_migration' AND environment = '$ENVIRONMENT';"
}

# Show migration status
show_migration_status() {
    log_info "Migration status for environment: $ENVIRONMENT"
    
    if $DRY_RUN; then
        log_warning "Dry run mode - showing expected status"
        return 0
    fi
    
    local status_sql="
    SELECT 
        filename,
        applied_at,
        checksum,
        applied_by
    FROM schema_migrations 
    WHERE environment = '$ENVIRONMENT'
    ORDER BY applied_at;
    "
    
    if psql "$DB_URL" -c "$status_sql" >> "$LOG_FILE" 2>&1; then
        log_success "Migration status displayed"
    else
        log_error "Failed to get migration status"
        return 1
    fi
}

# Validate migration files
validate_migrations() {
    log_info "Validating migration files..."
    
    local missing_files=()
    
    for migration in "${MIGRATIONS[@]}"; do
        local migration_path="$MIGRATIONS_DIR/$migration"
        if [[ ! -f "$migration_path" ]]; then
            missing_files+=("$migration")
        fi
    done
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        log_error "Missing migration files:"
        for file in "${missing_files[@]}"; do
            log_error "  - $file"
        done
        exit 1
    fi
    
    log_success "All migration files found"
}

# Main execution function
main() {
    log_info "Starting Risk Assessment Service database migration"
    log_info "Log file: $LOG_FILE"
    
    # Parse arguments
    parse_args "$@"
    
    # Validate environment
    validate_environment
    
    # Get database configuration
    get_db_config
    
    # Validate migration files
    validate_migrations
    
    # Test database connection
    test_db_connection
    
    # Create migration tracking table
    create_migration_table
    
    if $ROLLBACK; then
        rollback_migration
        exit 0
    fi
    
    # Apply migrations
    local failed_migrations=()
    
    for migration in "${MIGRATIONS[@]}"; do
        if ! apply_migration "$migration"; then
            failed_migrations+=("$migration")
            if ! $FORCE; then
                break
            fi
        fi
    done
    
    # Show final status
    show_migration_status
    
    # Report results
    if [[ ${#failed_migrations[@]} -eq 0 ]]; then
        log_success "All migrations completed successfully"
        exit 0
    else
        log_error "Some migrations failed: ${failed_migrations[*]}"
        exit 1
    fi
}

# Cleanup function
cleanup() {
    # Remove temporary files
    rm -f /tmp/migration_*.sql
}

# Set up signal handlers
trap cleanup EXIT

# Run main function
main "$@"