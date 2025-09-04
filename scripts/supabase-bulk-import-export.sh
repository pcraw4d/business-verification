#!/bin/bash

# =============================================================================
# SUPABASE BULK IMPORT/EXPORT SCRIPT
# =============================================================================
# This script provides bulk import/export functionality for keyword management
# using Supabase's REST API and SQL capabilities.
#
# Usage:
#   ./supabase-bulk-import-export.sh [command] [options]
#
# Commands:
#   export-industries     Export all industries to CSV
#   export-keywords       Export all keywords to CSV
#   export-codes          Export all classification codes to CSV
#   import-industries     Import industries from CSV
#   import-keywords       Import keywords from CSV
#   import-codes          Import classification codes from CSV
#   backup-all           Backup all data to JSON files
#   restore-all          Restore all data from JSON files
# =============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_ROOT/data/supabase"
BACKUP_DIR="$PROJECT_ROOT/backups/supabase"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required environment variables are set
check_env() {
    if [[ -z "${SUPABASE_URL:-}" ]]; then
        log_error "SUPABASE_URL environment variable is not set"
        exit 1
    fi
    
    if [[ -z "${SUPABASE_SERVICE_ROLE_KEY:-}" ]]; then
        log_error "SUPABASE_SERVICE_ROLE_KEY environment variable is not set"
        exit 1
    fi
    
    log_info "Environment variables validated"
}

# Create necessary directories
setup_directories() {
    mkdir -p "$DATA_DIR"
    mkdir -p "$BACKUP_DIR"
    log_info "Directories created: $DATA_DIR, $BACKUP_DIR"
}

# Make API request to Supabase
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    
    local url="${SUPABASE_URL}/rest/v1/${endpoint}"
    local headers=(
        -H "apikey: $SUPABASE_SERVICE_ROLE_KEY"
        -H "Authorization: Bearer $SUPABASE_SERVICE_ROLE_KEY"
        -H "Content-Type: application/json"
        -H "Prefer: return=representation"
    )
    
    if [[ -n "$data" ]]; then
        curl -s -X "$method" "$url" "${headers[@]}" -d "$data"
    else
        curl -s -X "$method" "$url" "${headers[@]}"
    fi
}

# Export functions
export_industries() {
    log_info "Exporting industries to CSV..."
    
    local output_file="$DATA_DIR/industries_export_$(date +%Y%m%d_%H%M%S).csv"
    
    # Get data from Supabase
    local data=$(api_request "GET" "industries?select=*")
    
    if [[ -z "$data" ]]; then
        log_error "Failed to fetch industries data"
        return 1
    fi
    
    # Convert JSON to CSV
    echo "$data" | jq -r '
        ["id", "name", "description", "category", "is_active", "created_at", "updated_at"],
        (.[] | [.id, .name, .description, .category, .is_active, .created_at, .updated_at])
        | @csv' > "$output_file"
    
    log_success "Industries exported to: $output_file"
    echo "$output_file"
}

export_keywords() {
    log_info "Exporting keywords to CSV..."
    
    local output_file="$DATA_DIR/keywords_export_$(date +%Y%m%d_%H%M%S).csv"
    
    # Get data with industry names
    local data=$(api_request "GET" "industry_keywords?select=*,industries(name)")
    
    if [[ -z "$data" ]]; then
        log_error "Failed to fetch keywords data"
        return 1
    fi
    
    # Convert JSON to CSV
    echo "$data" | jq -r '
        ["id", "industry_id", "industry_name", "keyword", "weight", "keyword_type", "is_active", "created_at", "updated_at"],
        (.[] | [.id, .industry_id, .industries.name, .keyword, .weight, .keyword_type, .is_active, .created_at, .updated_at])
        | @csv' > "$output_file"
    
    log_success "Keywords exported to: $output_file"
    echo "$output_file"
}

export_codes() {
    log_info "Exporting classification codes to CSV..."
    
    local output_file="$DATA_DIR/codes_export_$(date +%Y%m%d_%H%M%S).csv"
    
    # Get data with industry names
    local data=$(api_request "GET" "classification_codes?select=*,industries(name)")
    
    if [[ -z "$data" ]]; then
        log_error "Failed to fetch classification codes data"
        return 1
    fi
    
    # Convert JSON to CSV
    echo "$data" | jq -r '
        ["id", "code", "description", "code_type", "industry_id", "industry_name", "is_active", "created_at", "updated_at"],
        (.[] | [.id, .code, .description, .code_type, .industry_id, .industries.name, .is_active, .created_at, .updated_at])
        | @csv' > "$output_file"
    
    log_success "Classification codes exported to: $output_file"
    echo "$output_file"
}

# Import functions
import_industries() {
    local csv_file="$1"
    
    if [[ ! -f "$csv_file" ]]; then
        log_error "CSV file not found: $csv_file"
        return 1
    fi
    
    log_info "Importing industries from: $csv_file"
    
    # Convert CSV to JSON
    local json_data=$(tail -n +2 "$csv_file" | while IFS=',' read -r id name description category is_active created_at updated_at; do
        cat << EOF
{
    "name": "$name",
    "description": "$description",
    "category": "$category",
    "is_active": $is_active,
    "created_at": "$created_at",
    "updated_at": "$updated_at"
}
EOF
    done | jq -s '.')
    
    # Import to Supabase
    local result=$(api_request "POST" "industries" "$json_data")
    
    if [[ -n "$result" ]]; then
        log_success "Industries imported successfully"
        echo "$result" | jq '. | length' | xargs -I {} log_info "Imported {} industries"
    else
        log_error "Failed to import industries"
        return 1
    fi
}

import_keywords() {
    local csv_file="$1"
    
    if [[ ! -f "$csv_file" ]]; then
        log_error "CSV file not found: $csv_file"
        return 1
    fi
    
    log_info "Importing keywords from: $csv_file"
    
    # Convert CSV to JSON
    local json_data=$(tail -n +2 "$csv_file" | while IFS=',' read -r id industry_id industry_name keyword weight keyword_type is_active created_at updated_at; do
        cat << EOF
{
    "industry_id": $industry_id,
    "keyword": "$keyword",
    "weight": $weight,
    "keyword_type": "$keyword_type",
    "is_active": $is_active,
    "created_at": "$created_at",
    "updated_at": "$updated_at"
}
EOF
    done | jq -s '.')
    
    # Import to Supabase
    local result=$(api_request "POST" "industry_keywords" "$json_data")
    
    if [[ -n "$result" ]]; then
        log_success "Keywords imported successfully"
        echo "$result" | jq '. | length' | xargs -I {} log_info "Imported {} keywords"
    else
        log_error "Failed to import keywords"
        return 1
    fi
}

import_codes() {
    local csv_file="$1"
    
    if [[ ! -f "$csv_file" ]]; then
        log_error "CSV file not found: $csv_file"
        return 1
    fi
    
    log_info "Importing classification codes from: $csv_file"
    
    # Convert CSV to JSON
    local json_data=$(tail -n +2 "$csv_file" | while IFS=',' read -r id code description code_type industry_id industry_name is_active created_at updated_at; do
        cat << EOF
{
    "code": "$code",
    "description": "$description",
    "code_type": "$code_type",
    "industry_id": $industry_id,
    "is_active": $is_active,
    "created_at": "$created_at",
    "updated_at": "$updated_at"
}
EOF
    done | jq -s '.')
    
    # Import to Supabase
    local result=$(api_request "POST" "classification_codes" "$json_data")
    
    if [[ -n "$result" ]]; then
        log_success "Classification codes imported successfully"
        echo "$result" | jq '. | length' | xargs -I {} log_info "Imported {} classification codes"
    else
        log_error "Failed to import classification codes"
        return 1
    fi
}

# Backup functions
backup_all() {
    local backup_timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="$BACKUP_DIR/backup_$backup_timestamp"
    
    mkdir -p "$backup_dir"
    
    log_info "Creating full backup to: $backup_dir"
    
    # Export all data
    local industries_file=$(export_industries)
    local keywords_file=$(export_keywords)
    local codes_file=$(export_codes)
    
    # Move files to backup directory
    mv "$industries_file" "$backup_dir/"
    mv "$keywords_file" "$backup_dir/"
    mv "$codes_file" "$backup_dir/"
    
    # Create backup manifest
    cat > "$backup_dir/manifest.json" << EOF
{
    "backup_timestamp": "$backup_timestamp",
    "backup_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "files": {
        "industries": "$(basename "$industries_file")",
        "keywords": "$(basename "$keywords_file")",
        "codes": "$(basename "$codes_file")"
    },
    "environment": {
        "supabase_url": "$SUPABASE_URL"
    }
}
EOF
    
    log_success "Backup completed: $backup_dir"
    echo "$backup_dir"
}

restore_all() {
    local backup_dir="$1"
    
    if [[ ! -d "$backup_dir" ]]; then
        log_error "Backup directory not found: $backup_dir"
        return 1
    fi
    
    log_warning "This will restore all data from backup. Continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Restore cancelled"
        return 0
    fi
    
    log_info "Restoring from backup: $backup_dir"
    
    # Check manifest
    if [[ -f "$backup_dir/manifest.json" ]]; then
        log_info "Backup manifest found:"
        cat "$backup_dir/manifest.json" | jq '.'
    fi
    
    # Restore data
    local manifest_file="$backup_dir/manifest.json"
    if [[ -f "$manifest_file" ]]; then
        local industries_file="$backup_dir/$(jq -r '.files.industries' "$manifest_file")"
        local keywords_file="$backup_dir/$(jq -r '.files.keywords' "$manifest_file")"
        local codes_file="$backup_dir/$(jq -r '.files.codes' "$manifest_file")"
        
        if [[ -f "$industries_file" ]]; then
            import_industries "$industries_file"
        fi
        
        if [[ -f "$keywords_file" ]]; then
            import_keywords "$keywords_file"
        fi
        
        if [[ -f "$codes_file" ]]; then
            import_codes "$codes_file"
        fi
    else
        log_warning "No manifest found, attempting to restore from CSV files in directory"
        
        for csv_file in "$backup_dir"/*.csv; do
            if [[ -f "$csv_file" ]]; then
                local filename=$(basename "$csv_file")
                if [[ "$filename" == *"industries"* ]]; then
                    import_industries "$csv_file"
                elif [[ "$filename" == *"keywords"* ]]; then
                    import_keywords "$csv_file"
                elif [[ "$filename" == *"codes"* ]]; then
                    import_codes "$csv_file"
                fi
            fi
        done
    fi
    
    log_success "Restore completed"
}

# Validation functions
validate_csv() {
    local csv_file="$1"
    local expected_columns="$2"
    
    if [[ ! -f "$csv_file" ]]; then
        log_error "CSV file not found: $csv_file"
        return 1
    fi
    
    local header=$(head -n 1 "$csv_file")
    if [[ "$header" != "$expected_columns" ]]; then
        log_error "CSV header mismatch. Expected: $expected_columns, Got: $header"
        return 1
    fi
    
    log_success "CSV validation passed: $csv_file"
    return 0
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        "export-industries")
            check_env
            setup_directories
            export_industries
            ;;
        "export-keywords")
            check_env
            setup_directories
            export_keywords
            ;;
        "export-codes")
            check_env
            setup_directories
            export_codes
            ;;
        "import-industries")
            if [[ -z "${2:-}" ]]; then
                log_error "CSV file path required"
                exit 1
            fi
            check_env
            validate_csv "$2" "id,name,description,category,is_active,created_at,updated_at"
            import_industries "$2"
            ;;
        "import-keywords")
            if [[ -z "${2:-}" ]]; then
                log_error "CSV file path required"
                exit 1
            fi
            check_env
            validate_csv "$2" "id,industry_id,industry_name,keyword,weight,keyword_type,is_active,created_at,updated_at"
            import_keywords "$2"
            ;;
        "import-codes")
            if [[ -z "${2:-}" ]]; then
                log_error "CSV file path required"
                exit 1
            fi
            check_env
            validate_csv "$2" "id,code,description,code_type,industry_id,industry_name,is_active,created_at,updated_at"
            import_codes "$2"
            ;;
        "backup-all")
            check_env
            setup_directories
            backup_all
            ;;
        "restore-all")
            if [[ -z "${2:-}" ]]; then
                log_error "Backup directory path required"
                exit 1
            fi
            check_env
            restore_all "$2"
            ;;
        "help"|*)
            cat << EOF
Supabase Bulk Import/Export Script

Usage: $0 [command] [options]

Commands:
  export-industries     Export all industries to CSV
  export-keywords       Export all keywords to CSV
  export-codes          Export all classification codes to CSV
  import-industries     Import industries from CSV file
  import-keywords       Import keywords from CSV file
  import-codes          Import classification codes from CSV file
  backup-all           Backup all data to timestamped directory
  restore-all          Restore all data from backup directory
  help                 Show this help message

Environment Variables Required:
  SUPABASE_URL              Your Supabase project URL
  SUPABASE_SERVICE_ROLE_KEY Your Supabase service role key

Examples:
  $0 export-industries
  $0 import-keywords /path/to/keywords.csv
  $0 backup-all
  $0 restore-all /path/to/backup_20250119_143022

EOF
            ;;
    esac
}

# Run main function with all arguments
main "$@"
