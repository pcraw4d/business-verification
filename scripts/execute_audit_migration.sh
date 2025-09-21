#!/bin/bash

# =============================================================================
# Audit Tables Migration Execution Script
# =============================================================================
# This script executes the unified audit schema migration and data consolidation
# 
# Usage: ./execute_audit_migration.sh [options]
# 
# Options:
#   --dry-run     Show what would be migrated without executing
#   --validate    Only validate existing data without migration
#   --rollback    Rollback the migration if needed
#   --help        Show this help message
# =============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATION_FILE="$PROJECT_ROOT/internal/database/migrations/009_unified_audit_schema.sql"
LOG_FILE="$PROJECT_ROOT/logs/audit_migration_$(date +%Y%m%d_%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
DRY_RUN=false
VALIDATE_ONLY=false
ROLLBACK=false
HELP=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --validate)
            VALIDATE_ONLY=true
            shift
            ;;
        --rollback)
            ROLLBACK=true
            shift
            ;;
        --help)
            HELP=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Show help if requested
if [[ "$HELP" == true ]]; then
    echo "Audit Tables Migration Execution Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --dry-run     Show what would be migrated without executing"
    echo "  --validate    Only validate existing data without migration"
    echo "  --rollback    Rollback the migration if needed"
    echo "  --help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DATABASE_URL  PostgreSQL connection string (required)"
    echo "  LOG_LEVEL     Logging level (default: INFO)"
    echo ""
    exit 0
fi

# Function to log messages
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
        "DEBUG")
            echo -e "${BLUE}[DEBUG]${NC} $message"
            ;;
    esac
    
    # Also log to file
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."
    
    # Check if DATABASE_URL is set
    if [[ -z "${DATABASE_URL:-}" ]]; then
        log "ERROR" "DATABASE_URL environment variable is not set"
        log "ERROR" "Please set DATABASE_URL to your PostgreSQL connection string"
        exit 1
    fi
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log "ERROR" "psql command not found. Please install PostgreSQL client tools"
        exit 1
    fi
    
    # Check if migration file exists
    if [[ ! -f "$MIGRATION_FILE" ]]; then
        log "ERROR" "Migration file not found: $MIGRATION_FILE"
        exit 1
    fi
    
    # Create logs directory if it doesn't exist
    mkdir -p "$(dirname "$LOG_FILE")"
    
    log "INFO" "Prerequisites check completed successfully"
}

# Function to test database connection
test_database_connection() {
    log "INFO" "Testing database connection..."
    
    if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        log "INFO" "Database connection successful"
    else
        log "ERROR" "Failed to connect to database"
        log "ERROR" "Please check your DATABASE_URL and database availability"
        exit 1
    fi
}

# Function to backup existing audit tables
backup_audit_tables() {
    log "INFO" "Creating backup of existing audit tables..."
    
    local backup_file="$PROJECT_ROOT/backups/audit_tables_backup_$(date +%Y%m%d_%H%M%S).sql"
    mkdir -p "$(dirname "$backup_file")"
    
    # Create backup of all audit-related tables
    pg_dump "$DATABASE_URL" \
        --table=audit_logs \
        --table=merchant_audit_logs \
        --data-only \
        --no-owner \
        --no-privileges \
        > "$backup_file" 2>/dev/null || {
        log "WARN" "Some audit tables may not exist, continuing..."
    }
    
    log "INFO" "Backup created: $backup_file"
    echo "$backup_file"
}

# Function to validate existing data
validate_existing_data() {
    log "INFO" "Validating existing audit data..."
    
    # Check if audit tables exist and get counts
    local tables_info=$(psql "$DATABASE_URL" -t -c "
        SELECT 
            schemaname,
            tablename,
            n_tup_ins as inserts,
            n_tup_upd as updates,
            n_tup_del as deletes
        FROM pg_stat_user_tables 
        WHERE tablename IN ('audit_logs', 'merchant_audit_logs')
        ORDER BY tablename;
    " 2>/dev/null || echo "")
    
    if [[ -n "$tables_info" ]]; then
        log "INFO" "Existing audit tables found:"
        echo "$tables_info" | while read -r line; do
            if [[ -n "$line" ]]; then
                log "INFO" "  $line"
            fi
        done
    else
        log "INFO" "No existing audit tables found"
    fi
    
    # Check for data integrity issues
    log "INFO" "Checking for data integrity issues..."
    
    # Check for orphaned references
    local orphaned_count=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) 
        FROM audit_logs al 
        LEFT JOIN users u ON al.user_id = u.id 
        WHERE al.user_id IS NOT NULL AND u.id IS NULL;
    " 2>/dev/null | tr -d ' ' || echo "0")
    
    if [[ "$orphaned_count" != "0" ]]; then
        log "WARN" "Found $orphaned_count orphaned user references in audit_logs"
    fi
    
    local merchant_orphaned_count=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) 
        FROM merchant_audit_logs mal 
        LEFT JOIN merchants m ON mal.merchant_id = m.id 
        WHERE mal.merchant_id IS NOT NULL AND m.id IS NULL;
    " 2>/dev/null | tr -d ' ' || echo "0")
    
    if [[ "$merchant_orphaned_count" != "0" ]]; then
        log "WARN" "Found $merchant_orphaned_count orphaned merchant references in merchant_audit_logs"
    fi
    
    log "INFO" "Data validation completed"
}

# Function to execute migration
execute_migration() {
    log "INFO" "Executing unified audit schema migration..."
    
    if [[ "$DRY_RUN" == true ]]; then
        log "INFO" "DRY RUN: Would execute migration file: $MIGRATION_FILE"
        log "INFO" "DRY RUN: Migration would create unified_audit_logs table and migrate data"
        return 0
    fi
    
    # Execute the migration file
    if psql "$DATABASE_URL" -f "$MIGRATION_FILE" >> "$LOG_FILE" 2>&1; then
        log "INFO" "Migration file executed successfully"
    else
        log "ERROR" "Failed to execute migration file"
        log "ERROR" "Check the log file for details: $LOG_FILE"
        exit 1
    fi
    
    # Execute the data migration
    log "INFO" "Executing data migration..."
    local migration_result=$(psql "$DATABASE_URL" -t -c "SELECT * FROM migrate_audit_logs_to_unified();" 2>/dev/null || echo "")
    
    if [[ -n "$migration_result" ]]; then
        log "INFO" "Data migration completed:"
        echo "$migration_result" | while read -r line; do
            if [[ -n "$line" ]]; then
                log "INFO" "  $line"
            fi
        done
    else
        log "WARN" "Data migration may have failed or no data to migrate"
    fi
}

# Function to validate migration
validate_migration() {
    log "INFO" "Validating migration results..."
    
    # Run validation function
    local validation_result=$(psql "$DATABASE_URL" -t -c "SELECT * FROM validate_audit_migration();" 2>/dev/null || echo "")
    
    if [[ -n "$validation_result" ]]; then
        log "INFO" "Migration validation results:"
        echo "$validation_result" | while read -r line; do
            if [[ -n "$line" ]]; then
                log "INFO" "  $line"
            fi
        done
    else
        log "WARN" "Validation function may not be available"
    fi
    
    # Check if unified_audit_logs table exists and has data
    local unified_count=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM unified_audit_logs;" 2>/dev/null | tr -d ' ' || echo "0")
    log "INFO" "Unified audit logs count: $unified_count"
    
    if [[ "$unified_count" == "0" ]]; then
        log "WARN" "No data found in unified_audit_logs table"
    else
        log "INFO" "Migration validation completed successfully"
    fi
}

# Function to rollback migration
rollback_migration() {
    log "INFO" "Rolling back audit migration..."
    
    if [[ "$DRY_RUN" == true ]]; then
        log "INFO" "DRY RUN: Would rollback the migration"
        return 0
    fi
    
    # Execute rollback function
    local rollback_result=$(psql "$DATABASE_URL" -t -c "SELECT rollback_audit_migration();" 2>/dev/null || echo "")
    
    if [[ -n "$rollback_result" ]]; then
        log "INFO" "Rollback result: $rollback_result"
    else
        log "ERROR" "Rollback may have failed"
        exit 1
    fi
    
    log "INFO" "Migration rollback completed"
}

# Function to show migration summary
show_migration_summary() {
    log "INFO" "Migration Summary:"
    log "INFO" "=================="
    log "INFO" "Migration file: $MIGRATION_FILE"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Database: $(echo "$DATABASE_URL" | sed 's/:[^:]*@/:***@/')"
    
    if [[ "$DRY_RUN" == true ]]; then
        log "INFO" "Mode: DRY RUN (no changes made)"
    elif [[ "$VALIDATE_ONLY" == true ]]; then
        log "INFO" "Mode: VALIDATION ONLY"
    elif [[ "$ROLLBACK" == true ]]; then
        log "INFO" "Mode: ROLLBACK"
    else
        log "INFO" "Mode: FULL MIGRATION"
    fi
    
    log "INFO" "=================="
}

# Main execution
main() {
    log "INFO" "Starting audit tables migration process..."
    show_migration_summary
    
    check_prerequisites
    test_database_connection
    
    if [[ "$ROLLBACK" == true ]]; then
        rollback_migration
    elif [[ "$VALIDATE_ONLY" == true ]]; then
        validate_existing_data
    else
        local backup_file=""
        if [[ "$DRY_RUN" == false ]]; then
            backup_file=$(backup_audit_tables)
        fi
        
        validate_existing_data
        execute_migration
        validate_migration
        
        if [[ -n "$backup_file" ]]; then
            log "INFO" "Backup available at: $backup_file"
        fi
    fi
    
    log "INFO" "Audit tables migration process completed successfully!"
    log "INFO" "Check the log file for detailed information: $LOG_FILE"
}

# Run main function
main "$@"
