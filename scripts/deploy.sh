#!/bin/bash

# KYB Platform - Deployment Script
# This script handles deployments to different environments

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
DOCKER_REGISTRY="ghcr.io"
IMAGE_NAME="kyb-platform"
AWS_REGION="us-east-1"

# Default values
ENVIRONMENT="staging"
VERSION=""
FORCE_DEPLOY=false
SKIP_TESTS=false
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

Deploy KYB Platform to specified environment.

Options:
    -e, --environment ENV     Environment to deploy to (staging|production) [default: staging]
    -v, --version VERSION     Specific version to deploy (optional)
    -f, --force              Force deployment even if tests fail
    -s, --skip-tests         Skip running tests before deployment
    -d, --dry-run            Show what would be deployed without actually deploying
    -h, --help               Show this help message

Examples:
    $0 -e staging                    # Deploy to staging
    $0 -e production -v v1.2.3      # Deploy specific version to production
    $0 -e staging -f                # Force deploy to staging
    $0 -e production --dry-run      # Show production deployment plan

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
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_DEPLOY=true
                shift
                ;;
            -s|--skip-tests)
                SKIP_TESTS=true
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

# Validate environment
validate_environment() {
    case $ENVIRONMENT in
        staging|production)
            log_info "Deploying to environment: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'"
            exit 1
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
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
    
    # Check if kubectl is installed (for Kubernetes deployments)
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not found - Kubernetes deployments will be skipped"
    fi
    
    log_success "Prerequisites check completed"
}

# Get current version
get_current_version() {
    if [ -z "$VERSION" ]; then
        # Get version from git
        VERSION=$(git rev-parse --short HEAD)
        log_info "Using git commit as version: $VERSION"
    else
        log_info "Using specified version: $VERSION"
    fi
}

# Build Docker image
build_image() {
    log_info "Building Docker image..."
    
    local image_tag="$DOCKER_REGISTRY/$IMAGE_NAME:$VERSION"
    local latest_tag="$DOCKER_REGISTRY/$IMAGE_NAME:latest"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would build image $image_tag"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    # Build image
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        -t "$image_tag" \
        -t "$latest_tag" \
        .
    
    log_success "Docker image built successfully: $image_tag"
}

# Push Docker image
push_image() {
    log_info "Pushing Docker image to registry..."
    
    local image_tag="$DOCKER_REGISTRY/$IMAGE_NAME:$VERSION"
    local latest_tag="$DOCKER_REGISTRY/$IMAGE_NAME:latest"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would push image $image_tag"
        return 0
    fi
    
    # Push versioned tag
    docker push "$image_tag"
    
    # Push latest tag
    docker push "$latest_tag"
    
    log_success "Docker image pushed successfully"
}

# Run pre-deployment tests
run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        log_warning "Skipping tests as requested"
        return 0
    fi
    
    log_info "Running pre-deployment tests..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would run tests"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests
    if ! go test -v ./internal/... ./pkg/...; then
        if [ "$FORCE_DEPLOY" = true ]; then
            log_warning "Tests failed but continuing due to force flag"
        else
            log_error "Tests failed. Use --force to deploy anyway"
            exit 1
        fi
    fi
    
    log_success "Pre-deployment tests completed"
}

# Deploy to AWS ECS
deploy_to_ecs() {
    log_info "Deploying to AWS ECS..."
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    local service_name="kyb-platform-api"
    local image_uri="$DOCKER_REGISTRY/$IMAGE_NAME:$VERSION"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would deploy to ECS cluster: $cluster_name"
        log_info "DRY RUN: Would update service: $service_name"
        log_info "DRY RUN: Would use image: $image_uri"
        return 0
    fi
    
    # Update ECS service
    aws ecs update-service \
        --cluster "$cluster_name" \
        --service "$service_name" \
        --force-new-deployment
    
    # Wait for deployment to complete
    log_info "Waiting for deployment to complete..."
    aws ecs wait services-stable \
        --cluster "$cluster_name" \
        --services "$service_name"
    
    log_success "ECS deployment completed"
}

# Deploy to Kubernetes
deploy_to_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not available - skipping Kubernetes deployment"
        return 0
    fi
    
    local namespace="kyb-platform-$ENVIRONMENT"
    local image_uri="$DOCKER_REGISTRY/$IMAGE_NAME:$VERSION"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would deploy to Kubernetes namespace: $namespace"
        log_info "DRY RUN: Would use image: $image_uri"
        return 0
    fi
    
    # Create namespace if it doesn't exist
    kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f -
    
    # Update deployment
    kubectl set image deployment/kyb-platform-api \
        kyb-platform-api="$image_uri" \
        -n "$namespace"
    
    # Wait for rollout to complete
    kubectl rollout status deployment/kyb-platform-api -n "$namespace"
    
    log_success "Kubernetes deployment completed"
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."
    
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
        log_info "DRY RUN: Would check health at: $health_url"
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
            return 0
        else
            log_warning "Health check failed (attempt $attempt/$max_attempts)"
            if [ $attempt -eq $max_attempts ]; then
                log_error "Health check failed after $max_attempts attempts"
                return 1
            fi
            sleep 30
            attempt=$((attempt + 1))
        fi
    done
}

# Run smoke tests
run_smoke_tests() {
    log_info "Running smoke tests..."
    
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
        log_info "DRY RUN: Would run smoke tests against: $base_url"
        return 0
    fi
    
    # Test health endpoint
    curl -f "$base_url/health" || {
        log_error "Health endpoint test failed"
        return 1
    }
    
    # Test status endpoint
    curl -f "$base_url/status" || {
        log_error "Status endpoint test failed"
        return 1
    }
    
    # Test metrics endpoint
    curl -f "$base_url/metrics" || {
        log_error "Metrics endpoint test failed"
        return 1
    }
    
    log_success "Smoke tests passed"
}

# Create deployment record
create_deployment_record() {
    log_info "Creating deployment record..."
    
    local deployment_file="$PROJECT_ROOT/deployments/deployment-$ENVIRONMENT-$VERSION.json"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would create deployment record: $deployment_file"
        return 0
    fi
    
    # Create deployments directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/deployments"
    
    # Create deployment record
    cat > "$deployment_file" << EOF
{
    "environment": "$ENVIRONMENT",
    "version": "$VERSION",
    "deployed_at": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
    "deployed_by": "$(whoami)",
    "git_commit": "$(git rev-parse HEAD)",
    "git_branch": "$(git rev-parse --abbrev-ref HEAD)",
    "docker_image": "$DOCKER_REGISTRY/$IMAGE_NAME:$VERSION"
}
EOF
    
    log_success "Deployment record created: $deployment_file"
}

# Send notifications
send_notifications() {
    log_info "Sending deployment notifications..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would send notifications"
        return 0
    fi
    
    # This would integrate with your notification system
    # (Slack, Teams, email, etc.)
    
    local message="🚀 KYB Platform deployed to $ENVIRONMENT"
    message="$message\nVersion: $VERSION"
    message="$message\nDeployed by: $(whoami)"
    message="$message\nTime: $(date)"
    
    # Example: Send to Slack
    # curl -X POST -H 'Content-type: application/json' \
    #     --data "{\"text\":\"$message\"}" \
    #     "$SLACK_WEBHOOK_URL"
    
    log_success "Notifications sent"
}

# Main deployment function
main() {
    echo "=========================================="
    echo "      KYB Platform Deployment"
    echo "=========================================="
    echo "Environment: $ENVIRONMENT"
    echo "Version: $VERSION"
    echo "Force Deploy: $FORCE_DEPLOY"
    echo "Skip Tests: $SKIP_TESTS"
    echo "Dry Run: $DRY_RUN"
    echo "=========================================="
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Validate environment
    validate_environment
    
    # Check prerequisites
    check_prerequisites
    
    # Get current version
    get_current_version
    
    # Run pre-deployment tests
    run_tests
    
    # Build Docker image
    build_image
    
    # Push Docker image
    push_image
    
    # Deploy to ECS
    deploy_to_ecs
    
    # Deploy to Kubernetes (if available)
    deploy_to_kubernetes
    
    # Run health checks
    run_health_checks
    
    # Run smoke tests
    run_smoke_tests
    
    # Create deployment record
    create_deployment_record
    
    # Send notifications
    send_notifications
    
    echo
    echo "=========================================="
    echo "           DEPLOYMENT SUMMARY"
    echo "=========================================="
    echo "✅ Environment: $ENVIRONMENT"
    echo "✅ Version: $VERSION"
    echo "✅ Status: SUCCESS"
    echo "✅ Deployed by: $(whoami)"
    echo "✅ Time: $(date)"
    echo "=========================================="
    echo
    
    log_success "Deployment completed successfully!"
}

# Run main function
main "$@"
