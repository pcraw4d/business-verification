#!/bin/bash

# Configuration Rollback Script
# KYB Platform - Merchant-Centric UI Implementation
# 
# This script provides safe configuration rollback capabilities for the KYB platform.
# It supports rolling back environment configurations, feature flags, and system settings.

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_FILE="${PROJECT_ROOT}/logs/config-rollback-$(date +%Y%m%d-%H%M%S).log"
BACKUP_DIR="${PROJECT_ROOT}/backups/configuration"
CONFIG_DIR="${PROJECT_ROOT}/configs"

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

Configuration Rollback Script for KYB Platform

OPTIONS:
    -h, --help              Show this help message
    -v, --version           Show version information
    -d, --dry-run           Perform a dry run without executing changes
    -f, --force             Force rollback without confirmation prompts
    -b, --backup            Create backup before rollback
    -t, --target <version>  Specify target version to rollback to
    -l, --list              List available rollback targets
    -e, --environment       Specify environment (dev, staging, production)

ROLLBACK TARGETS:
    env                     Rollback environment variables
    features                Rollback feature flags
    database                Rollback database configuration
    api                     Rollback API configuration
    security                Rollback security settings
    full                    Full rollback (all configurations)

EXAMPLES:
    $0 --dry-run --environment production env
    $0 --backup --target v1.2.3 features
    $0 --list
    $0 --force --environment staging full

EOF
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if required directories exist
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_warn "Backup directory does not exist, creating: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
    
    if [[ ! -d "$CONFIG_DIR" ]]; then
        log_warn "Configuration directory does not exist, creating: $CONFIG_DIR"
        mkdir -p "$CONFIG_DIR"
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
    
    # List available configuration backups
    if [[ -d "$BACKUP_DIR" ]]; then
        echo "Configuration Backups:"
        for file in "$BACKUP_DIR"/*.tar.gz; do
            if [[ -f "$file" ]]; then
                echo "  - $(basename "$file" .tar.gz)"
            fi
        done
    fi
    
    # List current configuration files
    if [[ -d "$CONFIG_DIR" ]]; then
        echo "Current Configuration Files:"
        find "$CONFIG_DIR" -name "*.yaml" -o -name "*.yml" -o -name "*.json" -o -name "*.env" | while read -r line; do
            echo "  - $(basename "$line")"
        done
    fi
}

# Function to create configuration backup
create_configuration_backup() {
    local backup_name="config-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    log_info "Creating configuration backup: $backup_name"
    
    # Create backup of current configuration
    if tar -czf "$backup_path" -C "$PROJECT_ROOT" configs/; then
        log_success "Configuration backup created successfully: $backup_path"
        echo "$backup_path"
    else
        log_error "Failed to create configuration backup"
        exit 1
    fi
}

# Function to rollback environment variables
rollback_env() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back environment variables to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback environment variables to version $target_version"
        return 0
    fi
    
    # Find the environment backup
    local env_backup="$BACKUP_DIR/env-backup-${target_version}.tar.gz"
    if [[ ! -f "$env_backup" ]]; then
        log_error "Environment backup not found: $env_backup"
        exit 1
    fi
    
    # Backup current environment
    local current_env_backup="$BACKUP_DIR/env-current-$(date +%Y%m%d-%H%M%S).tar.gz"
    tar -czf "$current_env_backup" -C "$PROJECT_ROOT" .env* || log_warn "Failed to backup current environment"
    
    # Extract environment backup
    log_info "Extracting environment backup: $env_backup"
    if tar -xzf "$env_backup" -C "$PROJECT_ROOT"; then
        log_success "Environment variables rollback completed"
    else
        log_error "Failed to extract environment backup"
        exit 1
    fi
}

# Function to rollback feature flags
rollback_features() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back feature flags to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback feature flags to version $target_version"
        return 0
    fi
    
    # Find the feature flags backup
    local features_backup="$BACKUP_DIR/features-backup-${target_version}.json"
    if [[ ! -f "$features_backup" ]]; then
        log_error "Feature flags backup not found: $features_backup"
        exit 1
    fi
    
    # Backup current feature flags
    local current_features_backup="$BACKUP_DIR/features-current-$(date +%Y%m%d-%H%M%S).json"
    cp "$CONFIG_DIR/features.json" "$current_features_backup" 2>/dev/null || log_warn "No current features.json found"
    
    # Restore feature flags
    log_info "Restoring feature flags: $features_backup"
    if cp "$features_backup" "$CONFIG_DIR/features.json"; then
        log_success "Feature flags rollback completed"
    else
        log_error "Failed to restore feature flags"
        exit 1
    fi
}

# Function to rollback database configuration
rollback_database_config() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back database configuration to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback database configuration to version $target_version"
        return 0
    fi
    
    # Find the database configuration backup
    local db_backup="$BACKUP_DIR/database-backup-${target_version}.yaml"
    if [[ ! -f "$db_backup" ]]; then
        log_error "Database configuration backup not found: $db_backup"
        exit 1
    fi
    
    # Backup current database configuration
    local current_db_backup="$BACKUP_DIR/database-current-$(date +%Y%m%d-%H%M%S).yaml"
    cp "$CONFIG_DIR/database.yaml" "$current_db_backup" 2>/dev/null || log_warn "No current database.yaml found"
    
    # Restore database configuration
    log_info "Restoring database configuration: $db_backup"
    if cp "$db_backup" "$CONFIG_DIR/database.yaml"; then
        log_success "Database configuration rollback completed"
    else
        log_error "Failed to restore database configuration"
        exit 1
    fi
}

# Function to rollback API configuration
rollback_api_config() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back API configuration to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback API configuration to version $target_version"
        return 0
    fi
    
    # Find the API configuration backup
    local api_backup="$BACKUP_DIR/api-backup-${target_version}.yaml"
    if [[ ! -f "$api_backup" ]]; then
        log_error "API configuration backup not found: $api_backup"
        exit 1
    fi
    
    # Backup current API configuration
    local current_api_backup="$BACKUP_DIR/api-current-$(date +%Y%m%d-%H%M%S).yaml"
    cp "$CONFIG_DIR/api.yaml" "$current_api_backup" 2>/dev/null || log_warn "No current api.yaml found"
    
    # Restore API configuration
    log_info "Restoring API configuration: $api_backup"
    if cp "$api_backup" "$CONFIG_DIR/api.yaml"; then
        log_success "API configuration rollback completed"
    else
        log_error "Failed to restore API configuration"
        exit 1
    fi
}

# Function to rollback security settings
rollback_security() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back security settings to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback security settings to version $target_version"
        return 0
    fi
    
    # Find the security configuration backup
    local security_backup="$BACKUP_DIR/security-backup-${target_version}.yaml"
    if [[ ! -f "$security_backup" ]]; then
        log_error "Security configuration backup not found: $security_backup"
        exit 1
    fi
    
    # Backup current security configuration
    local current_security_backup="$BACKUP_DIR/security-current-$(date +%Y%m%d-%H%M%S).yaml"
    cp "$CONFIG_DIR/security.yaml" "$current_security_backup" 2>/dev/null || log_warn "No current security.yaml found"
    
    # Restore security configuration
    log_info "Restoring security configuration: $security_backup"
    if cp "$security_backup" "$CONFIG_DIR/security.yaml"; then
        log_success "Security settings rollback completed"
    else
        log_error "Failed to restore security configuration"
        exit 1
    fi
}

# Function to perform full configuration rollback
rollback_full() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    local create_backup_flag="$4"
    
    log_info "Performing full configuration rollback to version: $target_version"
    
    local backup_file=""
    if [[ "$create_backup_flag" == "true" ]]; then
        backup_file=$(create_configuration_backup)
    fi
    
    # Rollback all configuration types
    rollback_env "$target_version" "$dry_run" "$environment"
    rollback_features "$target_version" "$dry_run" "$environment"
    rollback_database_config "$target_version" "$dry_run" "$environment"
    rollback_api_config "$target_version" "$dry_run" "$environment"
    rollback_security "$target_version" "$dry_run" "$environment"
    
    log_success "Full configuration rollback completed"
}

# Function to confirm rollback
confirm_rollback() {
    local rollback_type="$1"
    local target="$2"
    local environment="$3"
    
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo -e "${YELLOW}WARNING: This will rollback the configuration $rollback_type to: $target${NC}"
    echo -e "${YELLOW}Environment: $environment${NC}"
    echo -e "${YELLOW}This action may affect system behavior. Are you sure you want to continue?${NC}"
    read -p "Type 'yes' to confirm: " confirmation
    
    if [[ "$confirmation" != "yes" ]]; then
        log_info "Rollback cancelled by user"
        exit 0
    fi
}

# Function to verify rollback
verify_rollback() {
    local environment="$1"
    
    log_info "Verifying configuration rollback..."
    
    # Check if configuration files exist and are valid
    local config_files=("$CONFIG_DIR/database.yaml" "$CONFIG_DIR/api.yaml" "$CONFIG_DIR/security.yaml" "$CONFIG_DIR/features.json")
    
    for config_file in "${config_files[@]}"; do
        if [[ -f "$config_file" ]]; then
            log_success "Configuration file exists: $(basename "$config_file")"
            
            # Basic validation for YAML files
            if [[ "$config_file" == *.yaml ]] || [[ "$config_file" == *.yml ]]; then
                if command -v yq &> /dev/null; then
                    if yq eval '.' "$config_file" > /dev/null 2>&1; then
                        log_success "YAML syntax is valid: $(basename "$config_file")"
                    else
                        log_warn "YAML syntax may be invalid: $(basename "$config_file")"
                    fi
                fi
            fi
            
            # Basic validation for JSON files
            if [[ "$config_file" == *.json ]]; then
                if command -v jq &> /dev/null; then
                    if jq empty "$config_file" 2>/dev/null; then
                        log_success "JSON syntax is valid: $(basename "$config_file")"
                    else
                        log_warn "JSON syntax may be invalid: $(basename "$config_file")"
                    fi
                fi
            fi
        else
            log_warn "Configuration file missing: $(basename "$config_file")"
        fi
    done
    
    # Check environment variables
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        log_success "Environment file exists"
    else
        log_warn "Environment file missing"
    fi
}

# Main function
main() {
    # Default values
    local DRY_RUN="false"
    local FORCE="false"
    local CREATE_BACKUP="false"
    local TARGET_VERSION=""
    local ENVIRONMENT="production"
    local ROLLBACK_TYPE=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--version)
                echo "Configuration Rollback Script v1.0.0"
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
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -l|--list)
                list_rollback_targets
                exit 0
                ;;
            env|features|database|api|security|full)
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
    log_info "Starting configuration rollback process"
    log_info "Rollback type: $ROLLBACK_TYPE"
    log_info "Target version: ${TARGET_VERSION:-latest}"
    log_info "Environment: $ENVIRONMENT"
    log_info "Dry run: $DRY_RUN"
    log_info "Force: $FORCE"
    log_info "Create backup: $CREATE_BACKUP"
    
    # Check prerequisites
    check_prerequisites
    
    # Confirm rollback
    confirm_rollback "$ROLLBACK_TYPE" "${TARGET_VERSION:-latest}" "$ENVIRONMENT"
    
    # Execute rollback based on type
    case "$ROLLBACK_TYPE" in
        env)
            rollback_env "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        features)
            rollback_features "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        database)
            rollback_database_config "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        api)
            rollback_api_config "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        security)
            rollback_security "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        full)
            rollback_full "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT" "$CREATE_BACKUP"
            ;;
        *)
            log_error "Unknown rollback type: $ROLLBACK_TYPE"
            exit 1
            ;;
    esac
    
    # Verify rollback
    verify_rollback "$ENVIRONMENT"
    
    log_success "Configuration rollback process completed successfully"
}

# Run main function with all arguments
main "$@"
