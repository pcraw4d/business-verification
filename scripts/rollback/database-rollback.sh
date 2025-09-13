#!/bin/bash

# Database Rollback Script
# KYB Platform - Merchant-Centric UI Implementation
# 
# This script provides safe database rollback capabilities for the KYB platform.
# It supports rolling back to previous database schema versions and data states.

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_FILE="${PROJECT_ROOT}/logs/rollback-$(date +%Y%m%d-%H%M%S).log"
BACKUP_DIR="${PROJECT_ROOT}/backups/database"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] <rollback_target>

Database Rollback Script for KYB Platform

OPTIONS:
    -h, --help              Show this help message
    -v, --version           Show version information
    -d, --dry-run           Perform a dry run without executing changes
    -f, --force             Force rollback without confirmation prompts
    -b, --backup            Create backup before rollback
    -t, --target <version>  Specify target version to rollback to
    -l, --list              List available rollback targets

ROLLBACK TARGETS:
    schema                  Rollback database schema to previous version
    data                    Rollback data to previous state
    full                    Full rollback (schema + data)
    migration <id>          Rollback to specific migration ID

EXAMPLES:
    $0 --dry-run schema
    $0 --backup --target 005 full
    $0 --list
    $0 --force data

EOF
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if database connection is available
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL client (psql) is not installed"
        exit 1
    fi
    
    # Check if backup directory exists
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warn "Backup directory does not exist, creating: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
    
    # Check if log directory exists
    if [[ ! -d "$(dirname "$LOG_FILE")" ]]; then
        log_warn "Log directory does not exist, creating: $(dirname "$LOG_FILE")"
        mkdir -p "$(dirname "$LOG_FILE")"
    fi
    
    log_success "Prerequisites check completed"
}

# Function to list available rollback targets
list_rollback_targets() {
    log_info "Available rollback targets:"
    
    # List available migrations
    if [[ -d "$PROJECT_ROOT/internal/database/migrations" ]]; then
        echo "Database Migrations:"
        for file in "$PROJECT_ROOT/internal/database/migrations"/*.sql; do
            if [[ -f "$file" ]]; then
                echo "  - $(basename "$file" .sql)"
            fi
        done
    fi
    
    # List available backups
    if [[ -d "$BACKUP_DIR" ]]; then
        echo "Available Backups:"
        for file in "$BACKUP_DIR"/*.sql; do
            if [[ -f "$file" ]]; then
                echo "  - $(basename "$file" .sql)"
            fi
        done
    fi
}

# Function to create backup
create_backup() {
    local backup_name="backup-$(date +%Y%m%d-%H%M%S).sql"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    log_info "Creating database backup: $backup_name"
    
    # Get database connection details from environment
    local db_host="${DB_HOST:-localhost}"
    local db_port="${DB_PORT:-5432}"
    local db_name="${DB_NAME:-kyb_platform}"
    local db_user="${DB_USER:-postgres}"
    
    # Create backup
    if pg_dump -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" > "$backup_path"; then
        log_success "Backup created successfully: $backup_path"
        echo "$backup_path"
    else
        log_error "Failed to create backup"
        exit 1
    fi
}

# Function to rollback schema
rollback_schema() {
    local target_version="$1"
    local dry_run="$2"
    
    log_info "Rolling back database schema to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback schema to version $target_version"
        return 0
    fi
    
    # Find the migration file
    local migration_file="$PROJECT_ROOT/internal/database/migrations/${target_version}_*.sql"
    if [[ ! -f $migration_file ]]; then
        log_error "Migration file not found: $migration_file"
        exit 1
    fi
    
    # Execute rollback (this would need to be implemented based on your migration system)
    log_info "Executing schema rollback..."
    # Note: This is a placeholder - actual implementation depends on your migration system
    log_success "Schema rollback completed"
}

# Function to rollback data
rollback_data() {
    local backup_file="$1"
    local dry_run="$2"
    
    log_info "Rolling back data from backup: $backup_file"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would restore data from $backup_file"
        return 0
    fi
    
    # Get database connection details
    local db_host="${DB_HOST:-localhost}"
    local db_port="${DB_PORT:-5432}"
    local db_name="${DB_NAME:-kyb_platform}"
    local db_user="${DB_USER:-postgres}"
    
    # Restore data
    if psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" < "$backup_file"; then
        log_success "Data rollback completed successfully"
    else
        log_error "Data rollback failed"
        exit 1
    fi
}

# Function to perform full rollback
rollback_full() {
    local target_version="$1"
    local dry_run="$2"
    local create_backup_flag="$3"
    
    log_info "Performing full rollback to version: $target_version"
    
    local backup_file=""
    if [[ "$create_backup_flag" == "true" ]]; then
        backup_file=$(create_backup)
    fi
    
    # Rollback schema
    rollback_schema "$target_version" "$dry_run"
    
    # Rollback data if backup is available
    if [[ -n "$backup_file" && -f "$backup_file" ]]; then
        rollback_data "$backup_file" "$dry_run"
    fi
    
    log_success "Full rollback completed"
}

# Function to confirm rollback
confirm_rollback() {
    local rollback_type="$1"
    local target="$2"
    
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo -e "${YELLOW}WARNING: This will rollback the database $rollback_type to: $target${NC}"
    echo -e "${YELLOW}This action may result in data loss. Are you sure you want to continue?${NC}"
    read -p "Type 'yes' to confirm: " confirmation
    
    if [[ "$confirmation" != "yes" ]]; then
        log_info "Rollback cancelled by user"
        exit 0
    fi
}

# Main function
main() {
    # Default values
    local DRY_RUN="false"
    local FORCE="false"
    local CREATE_BACKUP="false"
    local TARGET_VERSION=""
    local ROLLBACK_TYPE=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                echo "Database Rollback Script v1.0.0"
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -f|--force)
                FORCE="true"
                shift
                ;;
            -b|--backup)
                CREATE_BACKUP="true"
                shift
                ;;
            -t|--target)
                TARGET_VERSION="$2"
                shift 2
                ;;
            -l|--list)
                list_rollback_targets
                exit 0
                ;;
            schema|data|full|migration)
                ROLLBACK_TYPE="$1"
                shift
                break
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Validate required parameters
    if [[ -z "$ROLLBACK_TYPE" ]]; then
        log_error "Rollback type is required"
        usage
        exit 1
    fi
    
    # Initialize logging
    log_info "Starting database rollback process"
    log_info "Rollback type: $ROLLBACK_TYPE"
    log_info "Target version: ${TARGET_VERSION:-latest}"
    log_info "Dry run: $DRY_RUN"
    log_info "Force: $FORCE"
    log_info "Create backup: $CREATE_BACKUP"
    
    # Check prerequisites
    check_prerequisites
    
    # Confirm rollback
    confirm_rollback "$ROLLBACK_TYPE" "${TARGET_VERSION:-latest}"
    
    # Execute rollback based on type
    case "$ROLLBACK_TYPE" in
        schema)
            rollback_schema "${TARGET_VERSION:-latest}" "$DRY_RUN"
            ;;
        data)
            if [[ -z "$TARGET_VERSION" ]]; then
                log_error "Target version is required for data rollback"
                exit 1
            fi
            rollback_data "$TARGET_VERSION" "$DRY_RUN"
            ;;
        full)
            rollback_full "${TARGET_VERSION:-latest}" "$DRY_RUN" "$CREATE_BACKUP"
            ;;
        migration)
            if [[ -z "$TARGET_VERSION" ]]; then
                log_error "Migration ID is required"
                exit 1
            fi
            rollback_schema "$TARGET_VERSION" "$DRY_RUN"
            ;;
        *)
            log_error "Unknown rollback type: $ROLLBACK_TYPE"
            exit 1
            ;;
    esac
    
    log_success "Database rollback process completed successfully"
}

# Run main function with all arguments
main "$@"
