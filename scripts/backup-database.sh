#!/bin/bash

# Supabase Database Backup Script
# This script creates a complete backup of the Supabase database before any schema changes

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${PROJECT_ROOT}/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="supabase_backup_${TIMESTAMP}"

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if required environment variables are set
check_environment() {
    log_info "Checking environment variables..."
    
    local missing_vars=()
    
    if [[ -z "${SUPABASE_URL:-}" ]]; then
        missing_vars+=("SUPABASE_URL")
    fi
    
    if [[ -z "${SUPABASE_API_KEY:-}" ]] && [[ -z "${SUPABASE_ANON_KEY:-}" ]]; then
        missing_vars+=("SUPABASE_API_KEY or SUPABASE_ANON_KEY")
    fi
    
    if [[ -z "${SUPABASE_SERVICE_ROLE_KEY:-}" ]]; then
        missing_vars+=("SUPABASE_SERVICE_ROLE_KEY")
    fi
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        echo ""
        echo "Please set these variables in your environment or .env file"
        exit 1
    fi
    
    log_success "Environment variables validated"
}

# Create backup directory
create_backup_directory() {
    log_info "Creating backup directory: $BACKUP_DIR"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        mkdir -p "$BACKUP_DIR"
        log_success "Backup directory created"
    else
        log_info "Backup directory already exists"
    fi
}

# Run the backup
run_backup() {
    log_info "Starting Supabase database backup..."
    log_info "Backup name: $BACKUP_NAME"
    log_info "Output directory: $BACKUP_DIR"
    
    cd "$PROJECT_ROOT"
    
    # Run the backup tool
    if go run cmd/backup/main.go \
        -output "$BACKUP_DIR" \
        -retention 30 \
        -verify \
        -timeout 30m; then
        log_success "Database backup completed successfully"
    else
        log_error "Database backup failed"
        exit 1
    fi
}

# Verify backup integrity
verify_backup() {
    log_info "Verifying backup integrity..."
    
    local latest_backup=$(find "$BACKUP_DIR" -name "backup_*" -type d | sort | tail -1)
    
    if [[ -z "$latest_backup" ]]; then
        log_error "No backup found to verify"
        exit 1
    fi
    
    log_info "Verifying backup: $(basename "$latest_backup")"
    
    # Check if metadata file exists
    if [[ ! -f "$latest_backup/backup_metadata.json" ]]; then
        log_error "Backup metadata file not found"
        exit 1
    fi
    
    # Check if all table files exist
    local metadata_file="$latest_backup/backup_metadata.json"
    local table_count=$(jq -r '.tables | length' "$metadata_file")
    
    log_info "Expected $table_count table backup files"
    
    for ((i=0; i<table_count; i++)); do
        local table_name=$(jq -r ".tables[$i].name" "$metadata_file")
        local table_file="$latest_backup/${table_name}.json"
        
        if [[ ! -f "$table_file" ]]; then
            log_error "Table backup file not found: $table_file"
            exit 1
        fi
    done
    
    log_success "Backup integrity verified"
}

# Display backup summary
show_backup_summary() {
    log_info "Backup Summary"
    echo "=================="
    
    local latest_backup=$(find "$BACKUP_DIR" -name "backup_*" -type d | sort | tail -1)
    
    if [[ -n "$latest_backup" ]]; then
        local metadata_file="$latest_backup/backup_metadata.json"
        
        if [[ -f "$metadata_file" ]]; then
            echo "Backup ID: $(jq -r '.backup_id' "$metadata_file")"
            echo "Timestamp: $(jq -r '.timestamp' "$metadata_file")"
            echo "Database: $(jq -r '.database_url' "$metadata_file")"
            echo "Total Records: $(jq -r '.total_records' "$metadata_file")"
            echo "Backup Size: $(jq -r '.backup_size' "$metadata_file") bytes"
            echo "Checksum: $(jq -r '.checksum' "$metadata_file")"
            echo "Status: $(jq -r '.status' "$metadata_file")"
            echo "Tables: $(jq -r '.tables | length' "$metadata_file")"
            echo "Backup Location: $latest_backup"
        fi
    fi
    
    echo ""
    log_success "Backup process completed successfully"
}

# List available backups
list_backups() {
    log_info "Available backups:"
    echo "==================="
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warning "No backup directory found"
        return
    fi
    
    local backup_count=0
    
    for backup_dir in "$BACKUP_DIR"/backup_*; do
        if [[ -d "$backup_dir" ]]; then
            local metadata_file="$backup_dir/backup_metadata.json"
            
            if [[ -f "$metadata_file" ]]; then
                local backup_id=$(jq -r '.backup_id' "$metadata_file")
                local timestamp=$(jq -r '.timestamp' "$metadata_file")
                local status=$(jq -r '.status' "$metadata_file")
                local records=$(jq -r '.total_records' "$metadata_file")
                local size=$(jq -r '.backup_size' "$metadata_file")
                
                echo "ID: $backup_id"
                echo "  Timestamp: $timestamp"
                echo "  Status: $status"
                echo "  Records: $records"
                echo "  Size: $size bytes"
                echo "  Location: $backup_dir"
                echo ""
                
                ((backup_count++))
            fi
        fi
    done
    
    if [[ $backup_count -eq 0 ]]; then
        log_warning "No backups found"
    else
        log_success "Found $backup_count backup(s)"
    fi
}

# Clean up old backups
cleanup_backups() {
    log_info "Cleaning up old backups..."
    
    cd "$PROJECT_ROOT"
    
    if go run cmd/backup/main.go -cleanup; then
        log_success "Backup cleanup completed"
    else
        log_warning "Backup cleanup failed"
    fi
}

# Main function
main() {
    echo "🗄️  Supabase Database Backup Script"
    echo "===================================="
    echo ""
    
    # Parse command line arguments
    case "${1:-backup}" in
        "backup")
            check_environment
            create_backup_directory
            run_backup
            verify_backup
            show_backup_summary
            ;;
        "list")
            list_backups
            ;;
        "cleanup")
            cleanup_backups
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  backup   Create a new database backup (default)"
            echo "  list     List available backups"
            echo "  cleanup  Clean up old backups"
            echo "  help     Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  SUPABASE_URL              Supabase project URL"
            echo "  SUPABASE_API_KEY          Supabase API key"
            echo "  SUPABASE_SERVICE_ROLE_KEY Supabase service role key"
            echo "  SUPABASE_JWT_SECRET       Supabase JWT secret"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
