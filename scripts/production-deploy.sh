#!/bin/bash

# KYB Platform - Production Deployment Script
# Automated deployment script for production environment

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PRODUCTION_DIR="/opt/kyb-platform"
BACKUP_DIR="/opt/kyb-platform/backups"
LOG_FILE="/var/log/kyb-platform-deploy.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log_error "$1"
    exit 1
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root"
        exit 1
    fi
}

# Check system requirements
check_requirements() {
    log "Checking system requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error_exit "Docker is not installed"
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        error_exit "Docker Compose is not installed"
    fi
    
    # Check available disk space
    available_space=$(df /opt | awk 'NR==2 {print $4}')
    if [[ $available_space -lt 10485760 ]]; then  # 10GB in KB
        error_exit "Insufficient disk space. Need at least 10GB available"
    fi
    
    # Check available memory
    available_memory=$(free -m | awk 'NR==2{printf "%.0f", $7}')
    if [[ $available_memory -lt 8192 ]]; then  # 8GB
        error_exit "Insufficient memory. Need at least 8GB available"
    fi
    
    log_success "System requirements check passed"
}

# Create backup
create_backup() {
    log "Creating backup..."
    
    local backup_name="kyb-platform-backup-$(date +%Y%m%d_%H%M%S)"
    local backup_file="$BACKUP_DIR/$backup_name.tar.gz"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"
    
    # Create backup
    tar -czf "$backup_file" \
        -C "$PRODUCTION_DIR" \
        configs/ \
        logs/ \
        data/ \
        2>/dev/null || log_warning "Some files could not be backed up"
    
    # Keep only last 10 backups
    find "$BACKUP_DIR" -name "kyb-platform-backup-*.tar.gz" -type f -printf '%T@ %p\n' | \
        sort -rn | tail -n +11 | cut -d' ' -f2- | xargs -r rm -f
    
    log_success "Backup created: $backup_file"
    echo "$backup_file"
}

# Validate configuration
validate_config() {
    log "Validating configuration..."
    
    # Check if production.env exists
    if [[ ! -f "$PRODUCTION_DIR/configs/production.env" ]]; then
        error_exit "Production environment file not found"
    fi
    
    # Check if docker-compose.production.yml exists
    if [[ ! -f "$PRODUCTION_DIR/docker-compose.production.yml" ]]; then
        error_exit "Production Docker Compose file not found"
    fi
    
    # Validate Docker Compose configuration
    if ! docker-compose -f "$PRODUCTION_DIR/docker-compose.production.yml" config >/dev/null 2>&1; then
        error_exit "Invalid Docker Compose configuration"
    fi
    
    # Check required environment variables
    source "$PRODUCTION_DIR/configs/production.env"
    required_vars=(
        "SUPABASE_URL"
        "SUPABASE_ANON_KEY"
        "SUPABASE_SERVICE_ROLE_KEY"
        "JWT_SECRET"
        "REDIS_PASSWORD"
    )
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            error_exit "Required environment variable $var is not set"
        fi
    done
    
    log_success "Configuration validation passed"
}

# Build Docker image
build_image() {
    log "Building Docker image..."
    
    cd "$PROJECT_ROOT"
    
    # Build the production image
    docker build -f Dockerfile.production -t kyb-platform:production . || \
        error_exit "Failed to build Docker image"
    
    # Tag with timestamp
    local timestamp=$(date +%Y%m%d_%H%M%S)
    docker tag kyb-platform:production kyb-platform:$timestamp
    
    log_success "Docker image built successfully"
}

# Deploy services
deploy_services() {
    log "Deploying services..."
    
    cd "$PRODUCTION_DIR"
    
    # Stop existing services
    log "Stopping existing services..."
    docker-compose -f docker-compose.production.yml down --remove-orphans || \
        log_warning "Some services were not running"
    
    # Start infrastructure services first
    log "Starting infrastructure services..."
    docker-compose -f docker-compose.production.yml up -d redis prometheus grafana alertmanager || \
        error_exit "Failed to start infrastructure services"
    
    # Wait for infrastructure services to be ready
    log "Waiting for infrastructure services to be ready..."
    sleep 30
    
    # Start application services
    log "Starting application services..."
    docker-compose -f docker-compose.production.yml up -d kyb-api || \
        error_exit "Failed to start application services"
    
    # Wait for application to be ready
    log "Waiting for application to be ready..."
    sleep 30
    
    log_success "Services deployed successfully"
}

# Health check
health_check() {
    log "Performing health checks..."
    
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        log "Health check attempt $attempt/$max_attempts"
        
        # Check application health
        if curl -f -s http://localhost:8080/health >/dev/null 2>&1; then
            log_success "Application health check passed"
            break
        fi
        
        if [[ $attempt -eq $max_attempts ]]; then
            error_exit "Health check failed after $max_attempts attempts"
        fi
        
        sleep 10
        ((attempt++))
    done
    
    # Check metrics endpoint
    if curl -f -s http://localhost:9090/metrics >/dev/null 2>&1; then
        log_success "Metrics endpoint check passed"
    else
        log_warning "Metrics endpoint check failed"
    fi
    
    # Check Grafana
    if curl -f -s http://localhost:3000/api/health >/dev/null 2>&1; then
        log_success "Grafana health check passed"
    else
        log_warning "Grafana health check failed"
    fi
}

# Update load balancer
update_load_balancer() {
    log "Updating load balancer configuration..."
    
    # Copy nginx configuration
    if [[ -f "$PRODUCTION_DIR/configs/production/nginx.conf" ]]; then
        sudo cp "$PRODUCTION_DIR/configs/production/nginx.conf" /etc/nginx/nginx.conf || \
            log_warning "Failed to copy nginx configuration"
        
        # Test nginx configuration
        if sudo nginx -t >/dev/null 2>&1; then
            sudo systemctl reload nginx || \
                log_warning "Failed to reload nginx"
            log_success "Load balancer updated successfully"
        else
            log_warning "Invalid nginx configuration"
        fi
    else
        log_warning "Nginx configuration file not found"
    fi
}

# Cleanup old images
cleanup_images() {
    log "Cleaning up old Docker images..."
    
    # Remove dangling images
    docker image prune -f >/dev/null 2>&1 || true
    
    # Keep only last 5 versions of kyb-platform images
    docker images kyb-platform --format "table {{.Tag}}\t{{.CreatedAt}}" | \
        grep -E "^[0-9]{8}_[0-9]{6}" | \
        sort -r | tail -n +6 | \
        awk '{print $1}' | \
        xargs -r -I {} docker rmi kyb-platform:{} || true
    
    log_success "Docker images cleaned up"
}

# Send deployment notification
send_notification() {
    local status="$1"
    local message="$2"
    
    # Send Slack notification if webhook URL is configured
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        if [[ "$status" == "error" ]]; then
            color="danger"
        elif [[ "$status" == "warning" ]]; then
            color="warning"
        fi
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"KYB Platform Deployment\",\"attachments\":[{\"color\":\"$color\",\"fields\":[{\"title\":\"Status\",\"value\":\"$status\",\"short\":true},{\"title\":\"Message\",\"value\":\"$message\",\"short\":false}]}]}" \
            "$SLACK_WEBHOOK_URL" >/dev/null 2>&1 || true
    fi
    
    # Send email notification if SMTP is configured
    if [[ -n "${SMTP_HOST:-}" && -n "${ALERT_EMAIL:-}" ]]; then
        echo "KYB Platform Deployment $status: $message" | \
            mail -s "KYB Platform Deployment $status" "$ALERT_EMAIL" 2>/dev/null || true
    fi
}

# Main deployment function
main() {
    log "Starting KYB Platform production deployment..."
    
    # Load environment variables
    if [[ -f "$PRODUCTION_DIR/configs/production.env" ]]; then
        source "$PRODUCTION_DIR/configs/production.env"
    fi
    
    # Pre-deployment checks
    check_root
    check_requirements
    validate_config
    
    # Create backup
    local backup_file
    backup_file=$(create_backup)
    
    # Build and deploy
    build_image
    deploy_services
    
    # Post-deployment checks
    health_check
    update_load_balancer
    cleanup_images
    
    # Send success notification
    send_notification "success" "Deployment completed successfully. Backup: $(basename "$backup_file")"
    
    log_success "KYB Platform production deployment completed successfully!"
    
    # Display service status
    log "Service status:"
    docker-compose -f "$PRODUCTION_DIR/docker-compose.production.yml" ps
}

# Rollback function
rollback() {
    log "Starting rollback procedure..."
    
    # Find latest backup
    local latest_backup
    latest_backup=$(find "$BACKUP_DIR" -name "kyb-platform-backup-*.tar.gz" -type f -printf '%T@ %p\n' | \
        sort -rn | head -1 | cut -d' ' -f2-)
    
    if [[ -z "$latest_backup" ]]; then
        error_exit "No backup found for rollback"
    fi
    
    log "Rolling back to: $(basename "$latest_backup")"
    
    # Stop services
    cd "$PRODUCTION_DIR"
    docker-compose -f docker-compose.production.yml down
    
    # Restore backup
    tar -xzf "$latest_backup" -C "$PRODUCTION_DIR"
    
    # Restart services
    docker-compose -f docker-compose.production.yml up -d
    
    # Health check
    health_check
    
    send_notification "warning" "Rollback completed to $(basename "$latest_backup")"
    
    log_success "Rollback completed successfully"
}

# Script usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -r, --rollback Rollback to previous version"
    echo "  -d, --dry-run  Perform a dry run (validate only)"
    echo ""
    echo "Examples:"
    echo "  $0                # Deploy to production"
    echo "  $0 --rollback     # Rollback to previous version"
    echo "  $0 --dry-run      # Validate configuration only"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -r|--rollback)
            rollback
            exit 0
            ;;
        -d|--dry-run)
            log "Performing dry run..."
            check_root
            check_requirements
            validate_config
            log_success "Dry run completed successfully"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Run main deployment
main
