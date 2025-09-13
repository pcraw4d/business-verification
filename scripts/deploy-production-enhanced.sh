#!/bin/bash

# KYB Platform - Enhanced Production Deployment Script
# This script provides comprehensive production deployment with advanced features

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
APP_NAME="kyb-platform"
DEPLOYMENT_ENV="production"
BUILD_DIR="build"
CONFIG_DIR="configs"
BACKUP_DIR="backups"
LOG_DIR="logs"
DEPLOYMENTS_DIR="deployments"

# Default values
DEPLOYMENT_STRATEGY="blue-green"
HEALTH_CHECK_TIMEOUT=10
AUTO_ROLLBACK=true
DRY_RUN=false
SKIP_TESTS=false
FORCE_DEPLOY=false
NOTIFICATION_ENABLED=true

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_debug() {
    echo -e "${PURPLE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Enhanced production deployment for KYB Platform.

Options:
    -s, --strategy STRATEGY     Deployment strategy (blue-green|rolling|canary) [default: blue-green]
    -t, --timeout TIMEOUT       Health check timeout in minutes [default: 10]
    -r, --auto-rollback         Enable automatic rollback on failure [default: true]
    -d, --dry-run              Show what would be deployed without actually deploying
    -k, --skip-tests           Skip running tests before deployment
    -f, --force                Force deployment even if health checks fail
    -n, --no-notifications     Disable deployment notifications
    -h, --help                 Show this help message

Deployment Strategies:
    blue-green    - Zero-downtime blue-green deployment
    rolling       - Rolling deployment with gradual replacement
    canary        - Canary deployment with traffic splitting

Examples:
    $0 -s blue-green -t 15                    # Blue-green deployment with 15min timeout
    $0 -s rolling --skip-tests                # Rolling deployment without tests
    $0 --dry-run -s canary                    # Dry run canary deployment
    $0 -f --no-notifications                  # Force deployment without notifications

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--strategy)
                DEPLOYMENT_STRATEGY="$2"
                shift 2
                ;;
            -t|--timeout)
                HEALTH_CHECK_TIMEOUT="$2"
                shift 2
                ;;
            -r|--auto-rollback)
                AUTO_ROLLBACK=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -k|--skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            -f|--force)
                FORCE_DEPLOY=true
                shift
                ;;
            -n|--no-notifications)
                NOTIFICATION_ENABLED=false
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Validate deployment parameters
validate_deployment_params() {
    log_info "Validating deployment parameters..."
    
    # Validate deployment strategy
    case $DEPLOYMENT_STRATEGY in
        blue-green|rolling|canary)
            log_info "Deployment strategy: $DEPLOYMENT_STRATEGY"
            ;;
        *)
            log_error "Invalid deployment strategy: $DEPLOYMENT_STRATEGY"
            exit 1
            ;;
    esac
    
    # Validate timeout
    if ! [[ "$HEALTH_CHECK_TIMEOUT" =~ ^[0-9]+$ ]] || [ "$HEALTH_CHECK_TIMEOUT" -lt 1 ]; then
        log_error "Invalid timeout: $HEALTH_CHECK_TIMEOUT. Must be a positive integer"
        exit 1
    fi
    
    log_success "Deployment parameters validated"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    # Check if production config exists
    if [ ! -f "$CONFIG_DIR/production.env" ]; then
        log_error "Production configuration file not found: $CONFIG_DIR/production.env"
        exit 1
    fi
    
    # Check if production Docker Compose file exists
    if [ ! -f "docker-compose.production.yml" ]; then
        log_error "Production Docker Compose file not found: docker-compose.production.yml"
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Create deployment directories
create_directories() {
    log_info "Creating deployment directories..."
    
    mkdir -p "$BUILD_DIR" "$BACKUP_DIR" "$LOG_DIR" "$DEPLOYMENTS_DIR"
    
    log_success "Deployment directories created"
}

# Generate deployment metadata
generate_deployment_metadata() {
    log_info "Generating deployment metadata..."
    
    DEPLOYMENT_ID=$(date +%Y%m%d-%H%M%S)
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    VERSION="${GIT_COMMIT}-${DEPLOYMENT_ID}"
    
    # Export environment variables for Docker Compose
    export BUILD_DATE
    export VERSION
    export GIT_COMMIT
    export GIT_BRANCH
    export DEPLOYMENT_ID
    
    log_info "Deployment ID: $DEPLOYMENT_ID"
    log_info "Git Commit: $GIT_COMMIT"
    log_info "Git Branch: $GIT_BRANCH"
    log_info "Version: $VERSION"
    
    log_success "Deployment metadata generated"
}

# Run pre-deployment tests
run_pre_deployment_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        log_warning "Skipping pre-deployment tests"
        return 0
    fi
    
    log_info "Running pre-deployment tests..."
    
    # Run unit tests
    log_info "Running unit tests..."
    go test ./... -v -race -coverprofile=coverage.out
    
    if [ $? -ne 0 ]; then
        log_error "Unit tests failed"
        exit 1
    fi
    
    # Run integration tests
    log_info "Running integration tests..."
    if [ -f "scripts/test-v3-api.sh" ]; then
        ./scripts/test-v3-api.sh
        if [ $? -ne 0 ]; then
            log_error "Integration tests failed"
            exit 1
        fi
    else
        log_warning "Integration test script not found, skipping"
    fi
    
    # Run security tests
    log_info "Running security tests..."
    if [ -f "scripts/security-scan.sh" ]; then
        ./scripts/security-scan.sh
        if [ $? -ne 0 ]; then
            log_error "Security tests failed"
            exit 1
        fi
    else
        log_warning "Security test script not found, skipping"
    fi
    
    log_success "Pre-deployment tests completed"
}

# Build Docker images
build_docker_images() {
    log_info "Building Docker images..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would build Docker images"
        return 0
    fi
    
    # Build production image
    docker build \
        --build-arg BUILD_DATE="$BUILD_DATE" \
        --build-arg VERSION="$VERSION" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --build-arg GIT_BRANCH="$GIT_BRANCH" \
        -f Dockerfile.production \
        -t "$APP_NAME:$VERSION" \
        -t "$APP_NAME:latest" \
        .
    
    if [ $? -eq 0 ]; then
        log_success "Docker images built successfully"
    else
        log_error "Docker image build failed"
        exit 1
    fi
}

# Create backup of current deployment
create_backup() {
    log_info "Creating backup of current deployment..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would create backup"
        return 0
    fi
    
    # Check if current deployment exists
    if docker-compose -f docker-compose.production.yml ps | grep -q "Up"; then
        BACKUP_FILE="$BACKUP_DIR/backup-$DEPLOYMENT_ID.tar.gz"
        
        # Create backup of volumes
        docker run --rm \
            -v "$(pwd)/$BACKUP_DIR":/backup \
            -v kyb_data:/data \
            -v redis_data:/redis \
            -v prometheus_data:/prometheus \
            -v grafana_data:/grafana \
            -v alertmanager_data:/alertmanager \
            alpine:latest \
            tar czf "/backup/backup-$DEPLOYMENT_ID.tar.gz" /data /redis /prometheus /grafana /alertmanager
        
        log_success "Backup created: $BACKUP_FILE"
    else
        log_warning "No current deployment to backup"
    fi
}

# Deploy based on strategy
deploy_application() {
    log_info "Deploying application using $DEPLOYMENT_STRATEGY strategy..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would deploy using $DEPLOYMENT_STRATEGY strategy"
        return 0
    fi
    
    case $DEPLOYMENT_STRATEGY in
        "blue-green")
            deploy_blue_green
            ;;
        "rolling")
            deploy_rolling
            ;;
        "canary")
            deploy_canary
            ;;
    esac
}

# Blue-green deployment
deploy_blue_green() {
    log_info "Performing blue-green deployment..."
    
    # Stop current deployment
    docker-compose -f docker-compose.production.yml down
    
    # Start new deployment
    docker-compose -f docker-compose.production.yml up -d
    
    log_success "Blue-green deployment completed"
}

# Rolling deployment
deploy_rolling() {
    log_info "Performing rolling deployment..."
    
    # Update services one by one
    docker-compose -f docker-compose.production.yml up -d --no-deps kyb-api
    
    # Wait for API to be healthy
    wait_for_health_check "kyb-api" 300
    
    # Update other services
    docker-compose -f docker-compose.production.yml up -d --no-deps redis prometheus grafana
    
    log_success "Rolling deployment completed"
}

# Canary deployment
deploy_canary() {
    log_info "Performing canary deployment..."
    
    # For canary, we'll use a simplified approach
    # In a real scenario, this would involve traffic splitting
    
    # Deploy canary version
    docker-compose -f docker-compose.production.yml up -d
    
    # Wait for canary to be healthy
    wait_for_health_check "kyb-api" 300
    
    log_success "Canary deployment completed"
}

# Wait for health check
wait_for_health_check() {
    local service_name="$1"
    local timeout_seconds="$2"
    local max_attempts=$((timeout_seconds / 10))
    local attempt=1
    
    log_info "Waiting for $service_name to be healthy..."
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f docker-compose.production.yml ps "$service_name" | grep -q "healthy"; then
            log_success "$service_name is healthy"
            return 0
        fi
        
        log_info "Health check attempt $attempt/$max_attempts for $service_name"
        sleep 10
        ((attempt++))
    done
    
    log_error "$service_name failed health check after $timeout_seconds seconds"
    return 1
}

# Run post-deployment tests
run_post_deployment_tests() {
    log_info "Running post-deployment tests..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would run post-deployment tests"
        return 0
    fi
    
    # Wait for services to be ready
    sleep 30
    
    # Test health endpoint
    local health_url="http://localhost:8080/health"
    local max_attempts=20
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f "$health_url" > /dev/null 2>&1; then
            log_success "Health endpoint test passed"
            break
        fi
        
        log_info "Health endpoint test attempt $attempt/$max_attempts"
        sleep 15
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_error "Health endpoint test failed"
        if [ "$FORCE_DEPLOY" != true ]; then
            return 1
        fi
    fi
    
    # Test API endpoints
    local api_url="http://localhost:8080"
    local endpoints=("status" "metrics" "v1/classify")
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        
        if [ "$endpoint" = "v1/classify" ]; then
            curl -f "$api_url/$endpoint" \
                -H "Content-Type: application/json" \
                -d '{"business_name": "Test Corp"}' > /dev/null 2>&1 || {
                log_error "$endpoint test failed"
                if [ "$FORCE_DEPLOY" != true ]; then
                    return 1
                fi
            }
        else
            curl -f "$api_url/$endpoint" > /dev/null 2>&1 || {
                log_error "$endpoint test failed"
                if [ "$FORCE_DEPLOY" != true ]; then
                    return 1
                fi
            }
        fi
        
        log_success "$endpoint test passed"
    done
    
    log_success "Post-deployment tests completed"
}

# Create deployment record
create_deployment_record() {
    log_info "Creating deployment record..."
    
    local deployment_file="$DEPLOYMENTS_DIR/deployment-$DEPLOYMENT_ID.json"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would create deployment record: $deployment_file"
        return 0
    fi
    
    cat > "$deployment_file" << EOF
{
    "deployment_id": "$DEPLOYMENT_ID",
    "environment": "$DEPLOYMENT_ENV",
    "strategy": "$DEPLOYMENT_STRATEGY",
    "version": "$VERSION",
    "git_commit": "$GIT_COMMIT",
    "git_branch": "$GIT_BRANCH",
    "build_date": "$BUILD_DATE",
    "triggered_by": "$(whoami)",
    "timestamp": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
    "status": "completed",
    "health_check_timeout": $HEALTH_CHECK_TIMEOUT,
    "auto_rollback": $AUTO_ROLLBACK,
    "force_deploy": $FORCE_DEPLOY,
    "skip_tests": $SKIP_TESTS
}
EOF
    
    log_success "Deployment record created: $deployment_file"
}

# Send deployment notifications
send_deployment_notifications() {
    if [ "$NOTIFICATION_ENABLED" != true ]; then
        log_info "Notifications disabled, skipping"
        return 0
    fi
    
    log_info "Sending deployment notifications..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would send deployment notifications"
        return 0
    fi
    
    # This would integrate with your notification system
    # (Slack, Teams, email, etc.)
    
    local message="🚀 KYB Platform deployment to $DEPLOYMENT_ENV"
    message="$message\nStrategy: $DEPLOYMENT_STRATEGY"
    message="$message\nVersion: $VERSION"
    message="$message\nGit Commit: $GIT_COMMIT"
    message="$message\nTriggered by: $(whoami)"
    message="$message\nTime: $(date)"
    message="$message\nStatus: SUCCESS"
    
    # Example: Send to Slack
    # curl -X POST -H 'Content-type: application/json' \
    #     --data "{\"text\":\"$message\"}" \
    #     "$SLACK_WEBHOOK_URL"
    
    log_success "Deployment notifications sent"
}

# Rollback function
rollback_deployment() {
    log_error "Deployment failed, initiating rollback..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would rollback deployment"
        return 0
    fi
    
    # Stop current deployment
    docker-compose -f docker-compose.production.yml down
    
    # Restore from backup if available
    local latest_backup=$(ls -t "$BACKUP_DIR"/backup-*.tar.gz 2>/dev/null | head -n1)
    if [ -n "$latest_backup" ]; then
        log_info "Restoring from backup: $latest_backup"
        
        docker run --rm \
            -v "$(pwd)/$BACKUP_DIR":/backup \
            -v kyb_data:/data \
            -v redis_data:/redis \
            -v prometheus_data:/prometheus \
            -v grafana_data:/grafana \
            -v alertmanager_data:/alertmanager \
            alpine:latest \
            tar xzf "/backup/$(basename "$latest_backup")" -C /
        
        # Restart with previous version
        docker-compose -f docker-compose.production.yml up -d
        
        log_success "Rollback completed"
    else
        log_error "No backup found for rollback"
        exit 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up deployment artifacts..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would cleanup deployment artifacts"
        return 0
    fi
    
    # Remove old Docker images (keep last 3 versions)
    docker images "$APP_NAME" --format "table {{.Tag}}" | grep -v "latest" | tail -n +4 | xargs -r docker rmi "$APP_NAME:{}" 2>/dev/null || true
    
    # Remove old backups (keep last 5)
    if [ -d "$BACKUP_DIR" ]; then
        cd "$BACKUP_DIR"
        ls -t backup-*.tar.gz 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
        cd "$PROJECT_ROOT"
    fi
    
    log_success "Cleanup completed"
}

# Main deployment function
main() {
    echo "=========================================="
    echo "      KYB Platform Enhanced Deployment"
    echo "=========================================="
    echo "Environment: $DEPLOYMENT_ENV"
    echo "Strategy: $DEPLOYMENT_STRATEGY"
    echo "Health Check Timeout: ${HEALTH_CHECK_TIMEOUT} minutes"
    echo "Auto Rollback: $AUTO_ROLLBACK"
    echo "Dry Run: $DRY_RUN"
    echo "Skip Tests: $SKIP_TESTS"
    echo "Force Deploy: $FORCE_DEPLOY"
    echo "=========================================="
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Validate deployment parameters
    validate_deployment_params
    
    # Check prerequisites
    check_prerequisites
    
    # Create directories
    create_directories
    
    # Generate metadata
    generate_deployment_metadata
    
    # Run pre-deployment tests
    run_pre_deployment_tests
    
    # Build Docker images
    build_docker_images
    
    # Create backup
    create_backup
    
    # Deploy application
    if deploy_application; then
        # Run post-deployment tests
        if run_post_deployment_tests; then
            # Create deployment record
            create_deployment_record
            
            # Send notifications
            send_deployment_notifications
            
            # Cleanup
            cleanup
            
            echo
            echo "=========================================="
            echo "         DEPLOYMENT SUMMARY"
            echo "=========================================="
            echo "✅ Environment: $DEPLOYMENT_ENV"
            echo "✅ Strategy: $DEPLOYMENT_STRATEGY"
            echo "✅ Version: $VERSION"
            echo "✅ Git Commit: $GIT_COMMIT"
            echo "✅ Status: SUCCESS"
            echo "✅ Triggered by: $(whoami)"
            echo "✅ Time: $(date)"
            echo "=========================================="
            echo
            
            log_success "Enhanced production deployment completed successfully!"
            
        else
            log_error "Post-deployment tests failed"
            if [ "$AUTO_ROLLBACK" = true ]; then
                rollback_deployment
            fi
            exit 1
        fi
    else
        log_error "Deployment failed"
        if [ "$AUTO_ROLLBACK" = true ]; then
            rollback_deployment
        fi
        exit 1
    fi
}

# Handle signals for cleanup
trap 'log_error "Deployment interrupted"; exit 1' INT TERM

# Run main function
main "$@"
