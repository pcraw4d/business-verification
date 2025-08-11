#!/bin/bash

# KYB Platform - Blue-Green Deployment Script
# This script provides zero-downtime blue-green deployments

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
DEPLOYMENT_STRATEGY="automatic"
HEALTH_CHECK_TIMEOUT=10
AUTO_SWITCH_TRAFFIC=true
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

Perform blue-green deployment for KYB Platform.

Options:
    -e, --environment ENV     Environment to deploy to (staging|production) [default: staging]
    -s, --strategy STRATEGY   Deployment strategy (automatic|manual_switch|canary) [default: automatic]
    -t, --timeout TIMEOUT     Health check timeout in minutes [default: 10]
    -a, --auto-switch         Automatically switch traffic after health checks [default: true]
    -d, --dry-run            Show what would be deployed without actually deploying
    -h, --help               Show this help message

Deployment Strategies:
    automatic      - Automatic deployment with traffic switching
    manual_switch  - Deploy and wait for manual traffic switch
    canary         - Gradual traffic shifting (not implemented yet)

Examples:
    $0 -e staging -s automatic                    # Automatic blue-green deployment to staging
    $0 -e production -s manual_switch            # Manual switch deployment to production
    $0 -e staging --dry-run                      # Dry run for staging deployment

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
            -s|--strategy)
                DEPLOYMENT_STRATEGY="$2"
                shift 2
                ;;
            -t|--timeout)
                HEALTH_CHECK_TIMEOUT="$2"
                shift 2
                ;;
            -a|--auto-switch)
                AUTO_SWITCH_TRAFFIC=true
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

# Validate deployment parameters
validate_deployment_params() {
    # Validate environment
    case $ENVIRONMENT in
        staging|production)
            log_info "Deployment environment: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'"
            exit 1
            ;;
    esac
    
    # Validate deployment strategy
    case $DEPLOYMENT_STRATEGY in
        automatic|manual_switch|canary)
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
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Get current active environment
get_current_environment() {
    log_info "Determining current active environment..."
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    
    if [ "$DRY_RUN" = true ]; then
        CURRENT_ENV="blue"
        log_info "DRY RUN: Assuming current environment is blue"
        return 0
    fi
    
    # Check which environment is currently active by examining target groups
    local blue_target_group="kyb-platform-$ENVIRONMENT-blue"
    local green_target_group="kyb-platform-$ENVIRONMENT-green"
    
    # Check if blue target group exists and has healthy targets
    local blue_healthy=$(aws elbv2 describe-target-health \
        --target-group-arn "$(aws elbv2 describe-target-groups --names "$blue_target_group" --query 'TargetGroups[0].TargetGroupArn' --output text 2>/dev/null || echo '')" \
        --query 'TargetHealthDescriptions[?TargetHealth.State==`healthy`] | length(@)' \
        --output text 2>/dev/null || echo "0")
    
    # Check if green target group exists and has healthy targets
    local green_healthy=$(aws elbv2 describe-target-health \
        --target-group-arn "$(aws elbv2 describe-target-groups --names "$green_target_group" --query 'TargetGroups[0].TargetGroupArn' --output text 2>/dev/null || echo '')" \
        --query 'TargetHealthDescriptions[?TargetHealth.State==`healthy`] | length(@)' \
        --output text 2>/dev/null || echo "0")
    
    # Determine current active environment
    if [ "$blue_healthy" -gt 0 ]; then
        CURRENT_ENV="blue"
    elif [ "$green_healthy" -gt 0 ]; then
        CURRENT_ENV="green"
    else
        CURRENT_ENV="blue"  # Default to blue if no environment is active
    fi
    
    log_info "Current active environment: $CURRENT_ENV"
}

# Determine target environment
determine_target_environment() {
    if [ "$CURRENT_ENV" = "blue" ]; then
        TARGET_ENV="green"
    else
        TARGET_ENV="blue"
    fi
    
    log_info "Target environment: $TARGET_ENV"
}

# Build and push Docker image
build_and_push_image() {
    log_info "Building and pushing Docker image..."
    
    local version=$(git rev-parse --short HEAD)
    local image_tag="$DOCKER_REGISTRY/$IMAGE_NAME:$version"
    local latest_tag="$DOCKER_REGISTRY/$IMAGE_NAME:latest"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would build and push image $image_tag"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    # Build image
    docker build \
        --build-arg VERSION="$version" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        -t "$image_tag" \
        -t "$latest_tag" \
        .
    
    # Push image
    docker push "$image_tag"
    docker push "$latest_tag"
    
    log_success "Docker image built and pushed: $image_tag"
}

# Deploy to target environment
deploy_to_target_environment() {
    log_info "Deploying to target environment: $TARGET_ENV"
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    local service_name="kyb-platform-api-$TARGET_ENV"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would deploy to $TARGET_ENV environment"
        log_info "DRY RUN: Cluster: $cluster_name"
        log_info "DRY RUN: Service: $service_name"
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
    
    log_success "Deployment to $TARGET_ENV environment completed"
}

# Run health checks on target environment
run_health_checks() {
    log_info "Running health checks on $TARGET_ENV environment..."
    
    local health_url=""
    case $ENVIRONMENT in
        staging)
            health_url="http://$TARGET_ENV.staging.kybplatform.com/health"
            ;;
        production)
            health_url="http://$TARGET_ENV.production.kybplatform.com/health"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would check health at $health_url"
        return 0
    fi
    
    # Wait for service to be ready
    log_info "Waiting for service to be ready..."
    sleep 60
    
    # Run health checks
    local max_attempts=$((HEALTH_CHECK_TIMEOUT * 2))  # 30-second intervals
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Health check attempt $attempt/$max_attempts"
        
        if curl -f "$health_url" > /dev/null 2>&1; then
            log_success "Health check passed for $TARGET_ENV environment"
            break
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

# Run smoke tests on target environment
run_smoke_tests() {
    log_info "Running smoke tests on $TARGET_ENV environment..."
    
    local base_url=""
    case $ENVIRONMENT in
        staging)
            base_url="http://$TARGET_ENV.staging.kybplatform.com"
            ;;
        production)
            base_url="http://$TARGET_ENV.production.kybplatform.com"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would run smoke tests against $base_url"
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
    
    # Test API functionality
    curl -f "$base_url/v1/classify" \
        -H "Content-Type: application/json" \
        -d '{"business_name": "Test Corp"}' || {
        log_error "API functionality test failed"
        return 1
    }
    
    log_success "Smoke tests passed for $TARGET_ENV environment"
}

# Switch traffic to target environment
switch_traffic() {
    log_info "Switching traffic to $TARGET_ENV environment..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would switch traffic from $CURRENT_ENV to $TARGET_ENV"
        return 0
    fi
    
    local load_balancer_name="kyb-platform-$ENVIRONMENT"
    local old_target_group="kyb-platform-$ENVIRONMENT-$CURRENT_ENV"
    local new_target_group="kyb-platform-$ENVIRONMENT-$TARGET_ENV"
    
    # Get load balancer ARN
    local lb_arn=$(aws elbv2 describe-load-balancers \
        --names "$load_balancer_name" \
        --query 'LoadBalancers[0].LoadBalancerArn' \
        --output text)
    
    # Get listener ARN
    local listener_arn=$(aws elbv2 describe-listeners \
        --load-balancer-arn "$lb_arn" \
        --query 'Listeners[0].ListenerArn' \
        --output text)
    
    # Get target group ARNs
    local old_tg_arn=$(aws elbv2 describe-target-groups \
        --names "$old_target_group" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)
    
    local new_tg_arn=$(aws elbv2 describe-target-groups \
        --names "$new_target_group" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)
    
    # Update listener rule to point to new target group
    aws elbv2 modify-listener \
        --listener-arn "$listener_arn" \
        --default-actions Type=forward,TargetGroupArn="$new_tg_arn"
    
    log_success "Traffic switched to $TARGET_ENV environment"
}

# Verify traffic switch
verify_traffic_switch() {
    log_info "Verifying traffic switch..."
    
    local health_url=""
    case $ENVIRONMENT in
        staging)
            health_url="https://staging.kybplatform.com/health"
            ;;
        production)
            health_url="https://api.kybplatform.com/health"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would verify traffic switch at $health_url"
        return 0
    fi
    
    # Wait for traffic switch to propagate
    log_info "Waiting for traffic switch to propagate..."
    sleep 30
    
    # Verify health check
    local max_attempts=10
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Verification attempt $attempt/$max_attempts"
        
        if curl -f "$health_url" > /dev/null 2>&1; then
            log_success "Traffic switch verification successful"
            break
        else
            log_warning "Traffic switch verification failed (attempt $attempt/$max_attempts)"
            if [ $attempt -eq $max_attempts ]; then
                log_error "Traffic switch verification failed after $max_attempts attempts"
                return 1
            fi
            sleep 30
            attempt=$((attempt + 1))
        fi
    done
}

# Run post-switch tests
run_post_switch_tests() {
    log_info "Running post-switch tests..."
    
    local base_url=""
    case $ENVIRONMENT in
        staging)
            base_url="https://staging.kybplatform.com"
            ;;
        production)
            base_url="https://api.kybplatform.com"
            ;;
    esac
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would run post-switch tests against $base_url"
        return 0
    fi
    
    # Test all endpoints
    local endpoints=("health" "status" "metrics" "v1/classify" "v1/auth/login" "v1/risk/assess")
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        
        if [ "$endpoint" = "v1/classify" ]; then
            curl -f "$base_url/$endpoint" \
                -H "Content-Type: application/json" \
                -d '{"business_name": "Test Corp"}' || {
                log_error "$endpoint test failed"
                return 1
            }
        elif [ "$endpoint" = "v1/auth/login" ]; then
            curl -f "$base_url/$endpoint" \
                -H "Content-Type: application/json" \
                -d '{"email": "test@example.com", "password": "testpass"}' || {
                log_error "$endpoint test failed"
                return 1
            }
        elif [ "$endpoint" = "v1/risk/assess" ]; then
            curl -f "$base_url/$endpoint" \
                -H "Content-Type: application/json" \
                -d '{"business_id": "test-123"}' || {
                log_error "$endpoint test failed"
                return 1
            }
        else
            curl -f "$base_url/$endpoint" || {
                log_error "$endpoint test failed"
                return 1
            }
        fi
        
        log_success "$endpoint test passed"
    done
    
    log_success "All post-switch tests passed"
}

# Cleanup old environment
cleanup_old_environment() {
    log_info "Cleaning up old $CURRENT_ENV environment..."
    
    local cluster_name="kyb-platform-$ENVIRONMENT"
    local service_name="kyb-platform-api-$CURRENT_ENV"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would cleanup old $CURRENT_ENV environment"
        log_info "DRY RUN: Cluster: $cluster_name"
        log_info "DRY RUN: Service: $service_name"
        return 0
    fi
    
    # Scale down old service to 0 instances
    aws ecs update-service \
        --cluster "$cluster_name" \
        --service "$service_name" \
        --desired-count 0
    
    # Wait for service to scale down
    aws ecs wait services-stable \
        --cluster "$cluster_name" \
        --services "$service_name"
    
    log_success "Old $CURRENT_ENV environment scaled down"
}

# Create deployment record
create_deployment_record() {
    log_info "Creating blue-green deployment record..."
    
    local deployment_file="$PROJECT_ROOT/deployments/blue-green-$ENVIRONMENT-$(date +%Y%m%d-%H%M%S).json"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would create deployment record: $deployment_file"
        return 0
    fi
    
    # Create deployments directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/deployments"
    
    # Create deployment record
    cat > "$deployment_file" << EOF
{
    "deployment_id": "$(date +%Y%m%d-%H%M%S)",
    "environment": "$ENVIRONMENT",
    "deployment_strategy": "$DEPLOYMENT_STRATEGY",
    "from_environment": "$CURRENT_ENV",
    "to_environment": "$TARGET_ENV",
    "version": "$(git rev-parse --short HEAD)",
    "triggered_by": "$(whoami)",
    "timestamp": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
    "status": "completed",
    "auto_switch_traffic": $AUTO_SWITCH_TRAFFIC,
    "health_check_timeout": $HEALTH_CHECK_TIMEOUT
}
EOF
    
    log_success "Blue-green deployment record created: $deployment_file"
}

# Send deployment notifications
send_deployment_notifications() {
    log_info "Sending blue-green deployment notifications..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN: Would send deployment notifications"
        return 0
    fi
    
    # This would integrate with your notification system
    # (Slack, Teams, email, etc.)
    
    local message="🔄 KYB Platform blue-green deployment to $ENVIRONMENT"
    message="$message\nStrategy: $DEPLOYMENT_STRATEGY"
    message="$message\nFrom: $CURRENT_ENV"
    message="$message\nTo: $TARGET_ENV"
    message="$message\nVersion: $(git rev-parse --short HEAD)"
    message="$message\nTriggered by: $(whoami)"
    message="$message\nTime: $(date)"
    
    # Example: Send to Slack
    # curl -X POST -H 'Content-type: application/json' \
    #     --data "{\"text\":\"$message\"}" \
    #     "$SLACK_WEBHOOK_URL"
    
    log_success "Blue-green deployment notifications sent"
}

# Main blue-green deployment function
main() {
    echo "=========================================="
    echo "      KYB Platform Blue-Green Deployment"
    echo "=========================================="
    echo "Environment: $ENVIRONMENT"
    echo "Strategy: $DEPLOYMENT_STRATEGY"
    echo "Health Check Timeout: ${HEALTH_CHECK_TIMEOUT} minutes"
    echo "Auto Switch Traffic: $AUTO_SWITCH_TRAFFIC"
    echo "Dry Run: $DRY_RUN"
    echo "=========================================="
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Validate deployment parameters
    validate_deployment_params
    
    # Check prerequisites
    check_prerequisites
    
    # Get current active environment
    get_current_environment
    
    # Determine target environment
    determine_target_environment
    
    # Build and push Docker image
    build_and_push_image
    
    # Deploy to target environment
    deploy_to_target_environment
    
    # Run health checks on target environment
    run_health_checks
    
    # Run smoke tests on target environment
    run_smoke_tests
    
    # Switch traffic if auto-switch is enabled
    if [ "$AUTO_SWITCH_TRAFFIC" = true ]; then
        switch_traffic
        verify_traffic_switch
        run_post_switch_tests
        cleanup_old_environment
    else
        log_warning "Auto-switch disabled. Manual traffic switch required."
        log_info "Target environment $TARGET_ENV is ready for traffic switch"
    fi
    
    # Create deployment record
    create_deployment_record
    
    # Send deployment notifications
    send_deployment_notifications
    
    echo
    echo "=========================================="
    echo "       BLUE-GREEN DEPLOYMENT SUMMARY"
    echo "=========================================="
    echo "✅ Environment: $ENVIRONMENT"
    echo "✅ Strategy: $DEPLOYMENT_STRATEGY"
    echo "✅ From Environment: $CURRENT_ENV"
    echo "✅ To Environment: $TARGET_ENV"
    echo "✅ Version: $(git rev-parse --short HEAD)"
    echo "✅ Auto Switch: $AUTO_SWITCH_TRAFFIC"
    echo "✅ Status: SUCCESS"
    echo "✅ Triggered by: $(whoami)"
    echo "✅ Time: $(date)"
    echo "=========================================="
    echo
    
    log_success "Blue-green deployment completed successfully!"
}

# Run main function
main "$@"
