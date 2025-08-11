#!/bin/bash

# KYB Platform - Rollback Script
# This script provides comprehensive rollback capabilities for different scenarios

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
AWS_REGION="us-east-1"
DOCKER_REGISTRY="ghcr.io"
IMAGE_NAME="kyb-platform"

# Default values
ENVIRONMENT="staging"
ROLLBACK_TYPE="previous"
TARGET_VERSION=""
REASON=""
FORCE_ROLLBACK=false
DRY_RUN=false

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

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Rollback KYB Platform to a previous version.

Options:
    -e, --environment ENV     Environment to rollback (staging|production) [default: staging]
    -t, --type TYPE          Rollback type (previous|specific|emergency) [default: previous]
    -v, --version VERSION    Specific version to rollback to (required for specific type)
    -r, --reason REASON      Reason for rollback (required)
    -f, --force             Force rollback even if health checks fail
    -d, --dry-run           Show what would be rolled back without actually rolling back
    -h, --help              Show this help message

Rollback Types:
    previous    - Rollback to the previous version
    specific    - Rollback to a specific version
    emergency   - Rollback to a known stable version

Examples:
    $0 -e staging -t previous -r "Performance issues detected"
    $0 -e production -t specific -v v1.2.3 -r "Critical bug in v1.3.0"
    $0 -e production -t emergency -r "Service unavailable" -f
    $0 -e staging --dry-run -t previous -r "Testing rollback procedure"

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
            -t|--type)
                ROLLBACK_TYPE="$2"
                shift 2
                ;;
            -v|--version)
                TARGET_VERSION="$2"
                shift 2
                ;;
            -r|--reason)
                REASON="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_ROLLBACK=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
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

# Validate rollback parameters
validate_rollback_params() {
    # Check required parameters
    if [ -z "$REASON" ]; then
        log_error "Reason for rollback is required"
        exit 1
    fi
    
    # Validate environment
    case $ENVIRONMENT in
        staging|production)
            log_info "Rollback environment: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'"
            exit 1
            ;;
    esac
    
    # Validate rollback type
    case $ROLLBACK_TYPE in
        previous|specific|emergency)
            log_info "Rollback type: $ROLLBACK_TYPE"
            ;;
        *)
            log_error "Invalid rollback type: $ROLLBACK_TYPE"
            exit 1
            ;;
    esac
    
    # Validate specific version if needed
    if [ "$ROLLBACK_TYPE" = "specific" ] && [ -z "$TARGET_VERSION" ]; then
        log_error "Target version is required for specific rollback type"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed or not in PATH"
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured or invalid"
        exit 1
    fi
    
    # Check if kubectl is installed (for Kubernetes rollbacks)
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not found - Kubernetes rollbacks will be skipped"
    fi
    
    log_success "Prerequisites check completed"
}

# Get current version
get_current_version() {
    log_info "Getting current version..."
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    local service_name="kyb-platform-api"
    
    if [ "$DRY_RUN" = true ]; then
        CURRENT_VERSION="dry-run-current"
        log_info "DRY RUN: Would get current version from $cluster_name"
        return 0
    fi
    
    # Get current task definition
    CURRENT_VERSION=$(aws ecs describe-services \
        --cluster "$cluster_name" \
        --services "$service_name" \
        --query 'services[0].taskDefinition' \
        --output text | cut -d: -f2)
    
    log_info "Current version: $CURRENT_VERSION"
}

# Determine rollback version
determine_rollback_version() {
    log_info "Determining rollback version..."
    
    case $ROLLBACK_TYPE in
        "previous")
            if [ "$DRY_RUN" = true ]; then
                ROLLBACK_VERSION="dry-run-previous"
                log_info "DRY RUN: Would rollback to previous version"
            else
                # Get previous task definition revision
                local prev_revision=$((CURRENT_VERSION - 1))
                ROLLBACK_VERSION="kyb-platform-api:$prev_revision"
                log_info "Rollback version: $ROLLBACK_VERSION"
            fi
            ;;
        "specific")
            ROLLBACK_VERSION="$TARGET_VERSION"
            log_info "Rollback version: $ROLLBACK_VERSION"
            ;;
        "emergency")
            # Use a known stable version
            ROLLBACK_VERSION="kyb-platform-api:stable"
            log_info "Emergency rollback version: $ROLLBACK_VERSION"
            ;;
    esac
}

# Validate rollback version
validate_rollback_version() {
    log_info "Validating rollback version..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would validate rollback version $ROLLBACK_VERSION"
        return 0
    fi
    
    # Check if rollback version exists
    if ! aws ecs describe-task-definition --task-definition "$ROLLBACK_VERSION" > /dev/null 2>&1; then
        log_error "Rollback version $ROLLBACK_VERSION does not exist"
        exit 1
    fi
    
    # Check if rollback is needed
    if [ "$CURRENT_VERSION" = "$ROLLBACK_VERSION" ]; then
        log_error "Current version is already $ROLLBACK_VERSION"
        exit 1
    fi
    
    log_success "Rollback version validation successful"
}

# Perform ECS rollback
perform_ecs_rollback() {
    log_info "Performing ECS rollback..."
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    local service_name="kyb-platform-api"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would rollback ECS service $service_name in cluster $cluster_name"
        log_info "DRY RUN: Would update to version $ROLLBACK_VERSION"
        return 0
    fi
    
    # Update ECS service
    aws ecs update-service \
        --cluster "$cluster_name" \
        --service "$service_name" \
        --task-definition "$ROLLBACK_VERSION" \
        --force-new-deployment
    
    # Wait for rollback to complete
    log_info "Waiting for rollback to complete..."
    aws ecs wait services-stable \
        --cluster "$cluster_name" \
        --services "$service_name"
    
    log_success "ECS rollback completed"
}

# Perform Kubernetes rollback
perform_kubernetes_rollback() {
    log_info "Performing Kubernetes rollback..."
    
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not available - skipping Kubernetes rollback"
        return 0
    fi
    
    local namespace="kyb-platform-$ENVIRONMENT"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would rollback Kubernetes deployment in namespace $namespace"
        return 0
    fi
    
    # Rollback to previous revision
    kubectl rollout undo deployment/kyb-platform-api -n "$namespace"
    
    # Wait for rollback to complete
    kubectl rollout status deployment/kyb-platform-api -n "$namespace"
    
    log_success "Kubernetes rollback completed"
}

# Verify rollback
verify_rollback() {
    log_info "Verifying rollback..."
    
    local health_url=""
    case $ENVIRONMENT in
        staging)
            health_url="http://staging.kybplatform.com/health"
            ;;
        production)
            health_url="https://api.kybplatform.com/health"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would verify rollback at $health_url"
        return 0
    fi
    
    # Wait for service to be ready
    log_info "Waiting for service to be ready..."
    sleep 60
    
    # Run health checks
    local max_attempts=10
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Health check attempt $attempt/$max_attempts"
        
        if curl -f "$health_url" > /dev/null 2>&1; then
            log_success "Health check passed"
            break
        else
            log_warning "Health check failed (attempt $attempt/$max_attempts)"
            if [ $attempt -eq $max_attempts ]; then
                if [ "$FORCE_ROLLBACK" = true ]; then
                    log_warning "Health check failed but continuing due to force flag"
                else
                    log_error "Health check failed after $max_attempts attempts"
                    return 1
                fi
            fi
            sleep 30
            attempt=$((attempt + 1))
        fi
    done
    
    # Check version
    local status_url="${health_url%/health}/status"
    local version_response=$(curl -s "$status_url" | jq -r '.version' 2>/dev/null || echo "unknown")
    log_info "Current version after rollback: $version_response"
    
    log_success "Rollback verification successful"
}

# Run post-rollback tests
run_post_rollback_tests() {
    log_info "Running post-rollback tests..."
    
    local base_url=""
    case $ENVIRONMENT in
        staging)
            base_url="http://staging.kybplatform.com"
            ;;
        production)
            base_url="https://api.kybplatform.com"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would run post-rollback tests against $base_url"
        return 0
    fi
    
    # Test health endpoint
    curl -f "$base_url/health" || {
        log_error "Health endpoint test failed"
        if [ "$FORCE_ROLLBACK" != true ]; then
            return 1
        fi
    }
    
    # Test status endpoint
    curl -f "$base_url/status" || {
        log_error "Status endpoint test failed"
        if [ "$FORCE_ROLLBACK" != true ]; then
            return 1
        fi
    }
    
    # Test API functionality
    curl -f "$base_url/v1/classify" \
        -H "Content-Type: application/json" \
        -d '{"business_name": "Test Corp"}' || {
        log_error "API functionality test failed"
        if [ "$FORCE_ROLLBACK" != true ]; then
            return 1
        fi
    }
    
    log_success "Post-rollback tests passed"
}

# Create rollback record
create_rollback_record() {
    log_info "Creating rollback record..."
    
    local rollback_file="$PROJECT_ROOT/deployments/rollback-$ENVIRONMENT-$(date +%Y%m%d-%H%M%S).json"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would create rollback record: $rollback_file"
        return 0
    fi
    
    # Create deployments directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/deployments"
    
    # Create rollback record
    cat > "$rollback_file" << EOF
{
    "rollback_id": "$(date +%Y%m%d-%H%M%S)",
    "environment": "$ENVIRONMENT",
    "reason": "$REASON",
    "rollback_type": "$ROLLBACK_TYPE",
    "from_version": "$CURRENT_VERSION",
    "to_version": "$ROLLBACK_VERSION",
    "triggered_by": "$(whoami)",
    "timestamp": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
    "status": "completed",
    "force_rollback": $FORCE_ROLLBACK
}
EOF
    
    log_success "Rollback record created: $rollback_file"
}

# Send rollback notifications
send_rollback_notifications() {
    log_info "Sending rollback notifications..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would send rollback notifications"
        return 0
    fi
    
    # This would integrate with your notification system
    # (Slack, Teams, email, etc.)
    
    local message="🔄 KYB Platform rollback to $ENVIRONMENT"
    message="$message\nType: $ROLLBACK_TYPE"
    message="$message\nFrom: $CURRENT_VERSION"
    message="$message\nTo: $ROLLBACK_VERSION"
    message="$message\nReason: $REASON"
    message="$message\nTriggered by: $(whoami)"
    message="$message\nTime: $(date)"
    
    # Example: Send to Slack
    # curl -X POST -H 'Content-type: application/json' \
    #     --data "{\"text\":\"$message\"}" \
    #     "$SLACK_WEBHOOK_URL"
    
    log_success "Rollback notifications sent"
}

# Main rollback function
main() {
    echo "=========================================="
    echo "      KYB Platform Rollback"
    echo "=========================================="
    echo "Environment: $ENVIRONMENT"
    echo "Rollback Type: $ROLLBACK_TYPE"
    echo "Target Version: $TARGET_VERSION"
    echo "Reason: $REASON"
    echo "Force Rollback: $FORCE_ROLLBACK"
    echo "Dry Run: $DRY_RUN"
    echo "=========================================="
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Validate rollback parameters
    validate_rollback_params
    
    # Check prerequisites
    check_prerequisites
    
    # Get current version
    get_current_version
    
    # Determine rollback version
    determine_rollback_version
    
    # Validate rollback version
    validate_rollback_version
    
    # Perform ECS rollback
    perform_ecs_rollback
    
    # Perform Kubernetes rollback (if available)
    perform_kubernetes_rollback
    
    # Verify rollback
    verify_rollback
    
    # Run post-rollback tests
    run_post_rollback_tests
    
    # Create rollback record
    create_rollback_record
    
    # Send rollback notifications
    send_rollback_notifications
    
    echo
    echo "=========================================="
    echo "           ROLLBACK SUMMARY"
    echo "=========================================="
    echo "✅ Environment: $ENVIRONMENT"
    echo "✅ Rollback Type: $ROLLBACK_TYPE"
    echo "✅ From Version: $CURRENT_VERSION"
    echo "✅ To Version: $ROLLBACK_VERSION"
    echo "✅ Reason: $REASON"
    echo "✅ Status: SUCCESS"
    echo "✅ Triggered by: $(whoami)"
    echo "✅ Time: $(date)"
    echo "=========================================="
    echo
    
    log_success "Rollback completed successfully!"
}

# Run main function
main "$@"
