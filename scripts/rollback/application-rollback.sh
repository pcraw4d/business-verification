#!/bin/bash

# Application Rollback Script
# KYB Platform - Merchant-Centric UI Implementation
# 
# This script provides safe application rollback capabilities for the KYB platform.
# It supports rolling back to previous application versions, configurations, and deployments.

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_FILE="${PROJECT_ROOT}/logs/app-rollback-$(date +%Y%m%d-%H%M%S).log"
BACKUP_DIR="${PROJECT_ROOT}/backups/application"
DEPLOYMENT_DIR="${PROJECT_ROOT}/deployments"

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

Application Rollback Script for KYB Platform

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
    binary                  Rollback application binary
    config                  Rollback configuration files
    full                    Full rollback (binary + config)
    deployment              Rollback deployment configuration
    docker                  Rollback Docker containers and images

EXAMPLES:
    $0 --dry-run --environment production binary
    $0 --backup --target v1.2.3 full
    $0 --list
    $0 --force --environment staging config

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
    
    if [[ ! -d "$DEPLOYMENT_DIR" ]]; then
        log_warn "Deployment directory does not exist, creating: $DEPLOYMENT_DIR"
        mkdir -p "$DEPLOYMENT_DIR"
    fi
    
    # Check if log directory exists
    if [[ ! -d "$(dirname "$LOG_FILE")" ]]; then
        log_warn "Log directory does not exist, creating: $(dirname "$LOG_FILE")"
        mkdir -p "$(dirname "$LOG_FILE")"
    fi
    
    # Check if Docker is available (for Docker rollbacks)
    if command -v docker &> /dev/null; then
        log_info "Docker is available"
    else
        log_warn "Docker is not available - Docker rollbacks will be skipped"
    fi
    
    log_success "Prerequisites check completed"
}

# Function to list available rollback targets
list_rollback_targets() {
    log_info "Available rollback targets:"
    
    # List available application versions
    if [[ -d "$BACKUP_DIR" ]]; then
        echo "Application Backups:"
        for file in "$BACKUP_DIR"/*.tar.gz; do
            if [[ -f "$file" ]]; then
                echo "  - $(basename "$file" .tar.gz)"
            fi
        done
    fi
    
    # List available deployments
    if [[ -d "$DEPLOYMENT_DIR" ]]; then
        echo "Deployment Configurations:"
        for file in "$DEPLOYMENT_DIR"/*.yaml; do
            if [[ -f "$file" ]]; then
                echo "  - $(basename "$file" .yaml)"
            fi
        done
    fi
    
    # List Docker images
    if command -v docker &> /dev/null; then
        echo "Docker Images:"
        docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.CreatedAt}}" | grep kyb-platform || echo "  No KYB platform images found"
    fi
}

# Function to create application backup
create_application_backup() {
    local backup_name="app-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    log_info "Creating application backup: $backup_name"
    
    # Create backup of current application
    if tar -czf "$backup_path" \
        --exclude="node_modules" \
        --exclude=".git" \
        --exclude="logs" \
        --exclude="tmp" \
        --exclude="coverage" \
        -C "$PROJECT_ROOT" .; then
        log_success "Application backup created successfully: $backup_path"
        echo "$backup_path"
    else
        log_error "Failed to create application backup"
        exit 1
    fi
}

# Function to rollback application binary
rollback_binary() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back application binary to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback binary to version $target_version"
        return 0
    fi
    
    # Stop current application
    log_info "Stopping current application..."
    if command -v systemctl &> /dev/null; then
        sudo systemctl stop kyb-platform || log_warn "Failed to stop service via systemctl"
    fi
    
    # Kill any running processes
    pkill -f "kyb-platform" || log_warn "No running KYB platform processes found"
    
    # Find the backup file
    local backup_file="$BACKUP_DIR/app-backup-${target_version}.tar.gz"
    if [[ ! -f "$backup_file" ]]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    # Extract backup
    log_info "Extracting backup: $backup_file"
    if tar -xzf "$backup_file" -C "$PROJECT_ROOT"; then
        log_success "Backup extracted successfully"
    else
        log_error "Failed to extract backup"
        exit 1
    fi
    
    # Rebuild application if needed
    if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
        log_info "Rebuilding Go application..."
        cd "$PROJECT_ROOT"
        go build -o kyb-platform ./cmd/server
    fi
    
    # Start application
    log_info "Starting application..."
    if command -v systemctl &> /dev/null; then
        sudo systemctl start kyb-platform
    else
        # Start in background
        nohup "$PROJECT_ROOT/kyb-platform" > "$PROJECT_ROOT/logs/app.log" 2>&1 &
    fi
    
    log_success "Application binary rollback completed"
}

# Function to rollback configuration
rollback_config() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back configuration to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback configuration to version $target_version"
        return 0
    fi
    
    # Find the configuration backup
    local config_backup="$BACKUP_DIR/config-backup-${target_version}.tar.gz"
    if [[ ! -f "$config_backup" ]]; then
        log_error "Configuration backup not found: $config_backup"
        exit 1
    fi
    
    # Backup current configuration
    local current_config_backup="$BACKUP_DIR/config-current-$(date +%Y%m%d-%H%M%S).tar.gz"
    tar -czf "$current_config_backup" -C "$PROJECT_ROOT" configs/ || log_warn "Failed to backup current configuration"
    
    # Extract configuration backup
    log_info "Extracting configuration backup: $config_backup"
    if tar -xzf "$config_backup" -C "$PROJECT_ROOT"; then
        log_success "Configuration rollback completed"
    else
        log_error "Failed to extract configuration backup"
        exit 1
    fi
    
    # Restart application to apply new configuration
    log_info "Restarting application to apply new configuration..."
    if command -v systemctl &> /dev/null; then
        sudo systemctl restart kyb-platform
    else
        pkill -f "kyb-platform"
        sleep 2
        nohup "$PROJECT_ROOT/kyb-platform" > "$PROJECT_ROOT/logs/app.log" 2>&1 &
    fi
}

# Function to rollback Docker deployment
rollback_docker() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    
    log_info "Rolling back Docker deployment to version: $target_version"
    
    if [[ "$dry_run" == "true" ]]; then
        log_info "DRY RUN: Would rollback Docker deployment to version $target_version"
        return 0
    fi
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not available"
        exit 1
    fi
    
    # Stop current containers
    log_info "Stopping current containers..."
    docker-compose -f "$PROJECT_ROOT/docker-compose.${environment}.yml" down || log_warn "Failed to stop containers"
    
    # Pull previous image version
    log_info "Pulling previous image version: kyb-platform:$target_version"
    docker pull "kyb-platform:$target_version" || log_error "Failed to pull image"
    
    # Update docker-compose file to use previous version
    local compose_file="$PROJECT_ROOT/docker-compose.${environment}.yml"
    if [[ -f "$compose_file" ]]; then
        sed -i.bak "s/image: kyb-platform:.*/image: kyb-platform:$target_version/" "$compose_file"
    fi
    
    # Start containers with previous version
    log_info "Starting containers with previous version..."
    docker-compose -f "$compose_file" up -d
    
    log_success "Docker deployment rollback completed"
}

# Function to perform full rollback
rollback_full() {
    local target_version="$1"
    local dry_run="$2"
    local environment="$3"
    local create_backup_flag="$4"
    
    log_info "Performing full application rollback to version: $target_version"
    
    local backup_file=""
    if [[ "$create_backup_flag" == "true" ]]; then
        backup_file=$(create_application_backup)
    fi
    
    # Rollback binary
    rollback_binary "$target_version" "$dry_run" "$environment"
    
    # Rollback configuration
    rollback_config "$target_version" "$dry_run" "$environment"
    
    log_success "Full application rollback completed"
}

# Function to confirm rollback
confirm_rollback() {
    local rollback_type="$1"
    local target="$2"
    local environment="$3"
    
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo -e "${YELLOW}WARNING: This will rollback the application $rollback_type to: $target${NC}"
    echo -e "${YELLOW}Environment: $environment${NC}"
    echo -e "${YELLOW}This action may result in service downtime. Are you sure you want to continue?${NC}"
    read -p "Type 'yes' to confirm: " confirmation
    
    if [[ "$confirmation" != "yes" ]]; then
        log_info "Rollback cancelled by user"
        exit 0
    fi
}

# Function to verify rollback
verify_rollback() {
    local environment="$1"
    
    log_info "Verifying rollback..."
    
    # Check if application is running
    if pgrep -f "kyb-platform" > /dev/null; then
        log_success "Application is running"
    else
        log_warn "Application is not running"
    fi
    
    # Check health endpoint
    local health_url="http://localhost:8080/health"
    if command -v curl &> /dev/null; then
        if curl -f -s "$health_url" > /dev/null; then
            log_success "Health check passed"
        else
            log_warn "Health check failed"
        fi
    fi
    
    # Check Docker containers if applicable
    if command -v docker &> /dev/null; then
        local running_containers=$(docker ps --filter "name=kyb-platform" --format "{{.Names}}" | wc -l)
        if [[ "$running_containers" -gt 0 ]]; then
            log_success "Docker containers are running: $running_containers"
        else
            log_warn "No Docker containers are running"
        fi
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
                echo "Application Rollback Script v1.0.0"
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
            binary|config|full|deployment|docker)
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
    log_info "Starting application rollback process"
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
        binary)
            rollback_binary "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        config)
            rollback_config "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        full)
            rollback_full "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT" "$CREATE_BACKUP"
            ;;
        deployment)
            rollback_config "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        docker)
            rollback_docker "${TARGET_VERSION:-latest}" "$DRY_RUN" "$ENVIRONMENT"
            ;;
        *)
            log_error "Unknown rollback type: $ROLLBACK_TYPE"
            exit 1
            ;;
    esac
    
    # Verify rollback
    verify_rollback "$ENVIRONMENT"
    
    log_success "Application rollback process completed successfully"
}

# Run main function with all arguments
main "$@"
