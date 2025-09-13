#!/bin/bash

# KYB Platform - Deployment Automation Script
# This script provides comprehensive deployment automation with CI/CD integration

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

# Default values
ENVIRONMENT="staging"
DEPLOYMENT_TYPE="automated"
TRIGGER_SOURCE="manual"
BUILD_NUMBER=""
COMMIT_SHA=""
BRANCH_NAME=""
PULL_REQUEST_NUMBER=""
NOTIFICATION_CHANNEL="deployment"
SLACK_WEBHOOK_URL=""
TEAMS_WEBHOOK_URL=""
EMAIL_RECIPIENTS=""

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

Automated deployment for KYB Platform with CI/CD integration.

Options:
    -e, --environment ENV       Environment to deploy to (staging|production) [default: staging]
    -t, --type TYPE            Deployment type (automated|manual|hotfix) [default: automated]
    -s, --source SOURCE        Trigger source (github|gitlab|jenkins|manual) [default: manual]
    -b, --build BUILD          Build number or identifier
    -c, --commit COMMIT        Git commit SHA
    -r, --branch BRANCH        Git branch name
    -p, --pr PR                Pull request number
    -n, --notify CHANNEL       Notification channel (slack|teams|email|all) [default: deployment]
    -w, --webhook URL          Slack webhook URL
    -m, --teams URL            Teams webhook URL
    -a, --email EMAILS         Email recipients (comma-separated)
    -h, --help                 Show this help message

Environment Variables:
    GITHUB_ACTIONS             Set to 'true' when running in GitHub Actions
    GITLAB_CI                  Set to 'true' when running in GitLab CI
    JENKINS_URL                Set when running in Jenkins
    BUILD_NUMBER               Build number from CI system
    GITHUB_SHA                 Git commit SHA from GitHub
    GITHUB_REF                 Git reference from GitHub
    GITLAB_COMMIT_SHA          Git commit SHA from GitLab
    GITLAB_REF_NAME            Git reference from GitLab
    SLACK_WEBHOOK_URL          Slack webhook URL for notifications
    TEAMS_WEBHOOK_URL          Teams webhook URL for notifications
    EMAIL_RECIPIENTS           Email recipients for notifications

Examples:
    $0 -e staging -s github -b 123 -c abc123 -r main
    $0 -e production -t hotfix -s manual -c def456
    $0 -e staging --notify all --webhook https://hooks.slack.com/...

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
                DEPLOYMENT_TYPE="$2"
                shift 2
                ;;
            -s|--source)
                TRIGGER_SOURCE="$2"
                shift 2
                ;;
            -b|--build)
                BUILD_NUMBER="$2"
                shift 2
                ;;
            -c|--commit)
                COMMIT_SHA="$2"
                shift 2
                ;;
            -r|--branch)
                BRANCH_NAME="$2"
                shift 2
                ;;
            -p|--pr)
                PULL_REQUEST_NUMBER="$2"
                shift 2
                ;;
            -n|--notify)
                NOTIFICATION_CHANNEL="$2"
                shift 2
                ;;
            -w|--webhook)
                SLACK_WEBHOOK_URL="$2"
                shift 2
                ;;
            -m|--teams)
                TEAMS_WEBHOOK_URL="$2"
                shift 2
                ;;
            -a|--email)
                EMAIL_RECIPIENTS="$2"
                shift 2
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

# Detect CI/CD environment
detect_ci_environment() {
    log_info "Detecting CI/CD environment..."
    
    if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
        TRIGGER_SOURCE="github"
        BUILD_NUMBER="${GITHUB_RUN_NUMBER:-$BUILD_NUMBER}"
        COMMIT_SHA="${GITHUB_SHA:-$COMMIT_SHA}"
        BRANCH_NAME="${GITHUB_REF_NAME:-$BRANCH_NAME}"
        log_info "Detected GitHub Actions environment"
    elif [ "${GITLAB_CI:-}" = "true" ]; then
        TRIGGER_SOURCE="gitlab"
        BUILD_NUMBER="${CI_PIPELINE_ID:-$BUILD_NUMBER}"
        COMMIT_SHA="${GITLAB_COMMIT_SHA:-$COMMIT_SHA}"
        BRANCH_NAME="${GITLAB_REF_NAME:-$BRANCH_NAME}"
        log_info "Detected GitLab CI environment"
    elif [ -n "${JENKINS_URL:-}" ]; then
        TRIGGER_SOURCE="jenkins"
        BUILD_NUMBER="${BUILD_NUMBER:-${BUILD_ID:-}}"
        COMMIT_SHA="${GIT_COMMIT:-$COMMIT_SHA}"
        BRANCH_NAME="${GIT_BRANCH:-$BRANCH_NAME}"
        log_info "Detected Jenkins environment"
    else
        log_info "No CI/CD environment detected, using manual mode"
    fi
    
    # Set default values if not provided
    BUILD_NUMBER="${BUILD_NUMBER:-$(date +%Y%m%d%H%M%S)}"
    COMMIT_SHA="${COMMIT_SHA:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
    BRANCH_NAME="${BRANCH_NAME:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')}"
    
    log_info "Trigger source: $TRIGGER_SOURCE"
    log_info "Build number: $BUILD_NUMBER"
    log_info "Commit SHA: $COMMIT_SHA"
    log_info "Branch: $BRANCH_NAME"
}

# Validate deployment parameters
validate_deployment_params() {
    log_info "Validating deployment parameters..."
    
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
    
    # Validate deployment type
    case $DEPLOYMENT_TYPE in
        automated|manual|hotfix)
            log_info "Deployment type: $DEPLOYMENT_TYPE"
            ;;
        *)
            log_error "Invalid deployment type: $DEPLOYMENT_TYPE"
            exit 1
            ;;
    esac
    
    # Validate trigger source
    case $TRIGGER_SOURCE in
        github|gitlab|jenkins|manual)
            log_info "Trigger source: $TRIGGER_SOURCE"
            ;;
        *)
            log_error "Invalid trigger source: $TRIGGER_SOURCE"
            exit 1
            ;;
    esac
    
    # Validate notification channel
    case $NOTIFICATION_CHANNEL in
        slack|teams|email|all|deployment)
            log_info "Notification channel: $NOTIFICATION_CHANNEL"
            ;;
        *)
            log_error "Invalid notification channel: $NOTIFICATION_CHANNEL"
            exit 1
            ;;
    esac
    
    log_success "Deployment parameters validated"
}

# Check deployment prerequisites
check_prerequisites() {
    log_info "Checking deployment prerequisites..."
    
    # Check if required tools are installed
    local required_tools=("docker" "docker-compose" "curl" "jq")
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "Required tool not found: $tool"
            exit 1
        fi
    done
    
    # Check if configuration files exist
    local config_files=("configs/$ENVIRONMENT.env" "docker-compose.production.yml")
    
    for config_file in "${config_files[@]}"; do
        if [ ! -f "$config_file" ]; then
            log_error "Configuration file not found: $config_file"
            exit 1
        fi
    done
    
    # Check Docker daemon
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Generate deployment metadata
generate_deployment_metadata() {
    log_info "Generating deployment metadata..."
    
    DEPLOYMENT_ID="deploy-${ENVIRONMENT}-${BUILD_NUMBER}-$(date +%Y%m%d%H%M%S)"
    VERSION="${COMMIT_SHA}-${BUILD_NUMBER}"
    BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    
    # Export environment variables for deployment scripts
    export DEPLOYMENT_ID
    export VERSION
    export BUILD_DATE
    export COMMIT_SHA
    export BRANCH_NAME
    export BUILD_NUMBER
    export ENVIRONMENT
    export DEPLOYMENT_TYPE
    export TRIGGER_SOURCE
    
    log_info "Deployment ID: $DEPLOYMENT_ID"
    log_info "Version: $VERSION"
    log_info "Build date: $BUILD_DATE"
    
    log_success "Deployment metadata generated"
}

# Run pre-deployment checks
run_pre_deployment_checks() {
    log_info "Running pre-deployment checks..."
    
    # Check if deployment is allowed for this branch
    if [ "$ENVIRONMENT" = "production" ] && [ "$BRANCH_NAME" != "main" ] && [ "$DEPLOYMENT_TYPE" != "hotfix" ]; then
        log_error "Production deployments are only allowed from main branch or hotfix type"
        exit 1
    fi
    
    # Check if there are uncommitted changes
    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
        log_warning "Uncommitted changes detected"
        if [ "$DEPLOYMENT_TYPE" != "hotfix" ]; then
            log_error "Deployment aborted due to uncommitted changes"
            exit 1
        fi
    fi
    
    # Check if tests are passing (if available)
    if [ -f "scripts/run_tests.sh" ]; then
        log_info "Running tests..."
        ./scripts/run_tests.sh
        if [ $? -ne 0 ]; then
            log_error "Tests failed, deployment aborted"
            exit 1
        fi
    fi
    
    log_success "Pre-deployment checks completed"
}

# Send deployment start notification
send_deployment_start_notification() {
    log_info "Sending deployment start notification..."
    
    local message="🚀 KYB Platform deployment started"
    message="$message\nEnvironment: $ENVIRONMENT"
    message="$message\nType: $DEPLOYMENT_TYPE"
    message="$message\nSource: $TRIGGER_SOURCE"
    message="$message\nVersion: $VERSION"
    message="$message\nBranch: $BRANCH_NAME"
    message="$message\nCommit: $COMMIT_SHA"
    message="$message\nBuild: $BUILD_NUMBER"
    message="$message\nTriggered by: $(whoami)"
    message="$message\nTime: $(date)"
    
    send_notification "$message" "info"
}

# Execute deployment
execute_deployment() {
    log_info "Executing deployment..."
    
    # Choose deployment strategy based on environment and type
    local deployment_strategy="blue-green"
    
    if [ "$ENVIRONMENT" = "staging" ]; then
        deployment_strategy="rolling"
    elif [ "$DEPLOYMENT_TYPE" = "hotfix" ]; then
        deployment_strategy="rolling"
    fi
    
    log_info "Using deployment strategy: $deployment_strategy"
    
    # Run the appropriate deployment script
    case $deployment_strategy in
        "blue-green")
            ./scripts/deploy-production-enhanced.sh -s blue-green -t 15
            ;;
        "rolling")
            ./scripts/deploy-production-enhanced.sh -s rolling -t 10
            ;;
        *)
            ./scripts/deploy-production-enhanced.sh -s blue-green -t 15
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        log_success "Deployment executed successfully"
        return 0
    else
        log_error "Deployment execution failed"
        return 1
    fi
}

# Run post-deployment verification
run_post_deployment_verification() {
    log_info "Running post-deployment verification..."
    
    # Wait for services to be ready
    sleep 30
    
    # Check service health
    local health_url="http://localhost:8080/health"
    local max_attempts=20
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f "$health_url" > /dev/null 2>&1; then
            log_success "Service health check passed"
            break
        fi
        
        log_info "Health check attempt $attempt/$max_attempts"
        sleep 15
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_error "Service health check failed"
        return 1
    fi
    
    # Run smoke tests
    if [ -f "scripts/smoke-tests.sh" ]; then
        log_info "Running smoke tests..."
        ./scripts/smoke-tests.sh
        if [ $? -ne 0 ]; then
            log_error "Smoke tests failed"
            return 1
        fi
    fi
    
    log_success "Post-deployment verification completed"
}

# Send deployment success notification
send_deployment_success_notification() {
    log_info "Sending deployment success notification..."
    
    local message="✅ KYB Platform deployment successful"
    message="$message\nEnvironment: $ENVIRONMENT"
    message="$message\nType: $DEPLOYMENT_TYPE"
    message="$message\nVersion: $VERSION"
    message="$message\nBranch: $BRANCH_NAME"
    message="$message\nCommit: $COMMIT_SHA"
    message="$message\nBuild: $BUILD_NUMBER"
    message="$message\nDeployment ID: $DEPLOYMENT_ID"
    message="$message\nTime: $(date)"
    
    send_notification "$message" "success"
}

# Send deployment failure notification
send_deployment_failure_notification() {
    log_info "Sending deployment failure notification..."
    
    local message="❌ KYB Platform deployment failed"
    message="$message\nEnvironment: $ENVIRONMENT"
    message="$message\nType: $DEPLOYMENT_TYPE"
    message="$message\nVersion: $VERSION"
    message="$message\nBranch: $BRANCH_NAME"
    message="$message\nCommit: $COMMIT_SHA"
    message="$message\nBuild: $BUILD_NUMBER"
    message="$message\nDeployment ID: $DEPLOYMENT_ID"
    message="$message\nTime: $(date)"
    
    send_notification "$message" "error"
}

# Send notification to configured channels
send_notification() {
    local message="$1"
    local level="$2"
    
    # Send to Slack
    if [ "$NOTIFICATION_CHANNEL" = "slack" ] || [ "$NOTIFICATION_CHANNEL" = "all" ] || [ "$NOTIFICATION_CHANNEL" = "deployment" ]; then
        send_slack_notification "$message" "$level"
    fi
    
    # Send to Teams
    if [ "$NOTIFICATION_CHANNEL" = "teams" ] || [ "$NOTIFICATION_CHANNEL" = "all" ]; then
        send_teams_notification "$message" "$level"
    fi
    
    # Send to Email
    if [ "$NOTIFICATION_CHANNEL" = "email" ] || [ "$NOTIFICATION_CHANNEL" = "all" ]; then
        send_email_notification "$message" "$level"
    fi
}

# Send Slack notification
send_slack_notification() {
    local message="$1"
    local level="$2"
    
    local webhook_url="${SLACK_WEBHOOK_URL:-${SLACK_WEBHOOK_URL}}"
    
    if [ -z "$webhook_url" ]; then
        log_warning "Slack webhook URL not configured, skipping Slack notification"
        return 0
    fi
    
    local color="good"
    case $level in
        "error")
            color="danger"
            ;;
        "warning")
            color="warning"
            ;;
        "info")
            color="#36a64f"
            ;;
        "success")
            color="good"
            ;;
    esac
    
    local payload=$(cat << EOF
{
    "attachments": [
        {
            "color": "$color",
            "text": "$message",
            "footer": "KYB Platform Deployment",
            "ts": $(date +%s)
        }
    ]
}
EOF
)
    
    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$webhook_url" > /dev/null 2>&1 || {
        log_warning "Failed to send Slack notification"
    }
}

# Send Teams notification
send_teams_notification() {
    local message="$1"
    local level="$2"
    
    local webhook_url="${TEAMS_WEBHOOK_URL:-${TEAMS_WEBHOOK_URL}}"
    
    if [ -z "$webhook_url" ]; then
        log_warning "Teams webhook URL not configured, skipping Teams notification"
        return 0
    fi
    
    local color="00ff00"
    case $level in
        "error")
            color="ff0000"
            ;;
        "warning")
            color="ffff00"
            ;;
        "info")
            color="0099ff"
            ;;
        "success")
            color="00ff00"
            ;;
    esac
    
    local payload=$(cat << EOF
{
    "@type": "MessageCard",
    "@context": "http://schema.org/extensions",
    "themeColor": "$color",
    "summary": "KYB Platform Deployment",
    "sections": [
        {
            "activityTitle": "KYB Platform Deployment",
            "activitySubtitle": "$(date)",
            "text": "$message",
            "markdown": true
        }
    ]
}
EOF
)
    
    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$webhook_url" > /dev/null 2>&1 || {
        log_warning "Failed to send Teams notification"
    }
}

# Send email notification
send_email_notification() {
    local message="$1"
    local level="$2"
    
    local recipients="${EMAIL_RECIPIENTS:-${EMAIL_RECIPIENTS}}"
    
    if [ -z "$recipients" ]; then
        log_warning "Email recipients not configured, skipping email notification"
        return 0
    fi
    
    local subject="KYB Platform Deployment - $level"
    case $level in
        "error")
            subject="❌ KYB Platform Deployment Failed"
            ;;
        "success")
            subject="✅ KYB Platform Deployment Successful"
            ;;
        "info")
            subject="🚀 KYB Platform Deployment Started"
            ;;
    esac
    
    # This would integrate with your email system
    # (SendGrid, SES, SMTP, etc.)
    log_info "Would send email to: $recipients"
    log_info "Subject: $subject"
    log_info "Message: $message"
}

# Create deployment record
create_deployment_record() {
    log_info "Creating deployment record..."
    
    local deployment_file="deployments/deployment-$DEPLOYMENT_ID.json"
    mkdir -p deployments
    
    cat > "$deployment_file" << EOF
{
    "deployment_id": "$DEPLOYMENT_ID",
    "environment": "$ENVIRONMENT",
    "deployment_type": "$DEPLOYMENT_TYPE",
    "trigger_source": "$TRIGGER_SOURCE",
    "version": "$VERSION",
    "build_number": "$BUILD_NUMBER",
    "commit_sha": "$COMMIT_SHA",
    "branch_name": "$BRANCH_NAME",
    "pull_request_number": "$PULL_REQUEST_NUMBER",
    "build_date": "$BUILD_DATE",
    "triggered_by": "$(whoami)",
    "timestamp": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
    "status": "completed",
    "notification_channel": "$NOTIFICATION_CHANNEL"
}
EOF
    
    log_success "Deployment record created: $deployment_file"
}

# Main deployment function
main() {
    echo "=========================================="
    echo "      KYB Platform Deployment Automation"
    echo "=========================================="
    echo "Environment: $ENVIRONMENT"
    echo "Type: $DEPLOYMENT_TYPE"
    echo "Source: $TRIGGER_SOURCE"
    echo "Build: $BUILD_NUMBER"
    echo "Commit: $COMMIT_SHA"
    echo "Branch: $BRANCH_NAME"
    echo "Notification: $NOTIFICATION_CHANNEL"
    echo "=========================================="
    echo
    
    # Parse arguments
    parse_args "$@"
    
    # Detect CI/CD environment
    detect_ci_environment
    
    # Validate deployment parameters
    validate_deployment_params
    
    # Check prerequisites
    check_prerequisites
    
    # Generate metadata
    generate_deployment_metadata
    
    # Run pre-deployment checks
    run_pre_deployment_checks
    
    # Send start notification
    send_deployment_start_notification
    
    # Execute deployment
    if execute_deployment; then
        # Run post-deployment verification
        if run_post_deployment_verification; then
            # Create deployment record
            create_deployment_record
            
            # Send success notification
            send_deployment_success_notification
            
            echo
            echo "=========================================="
            echo "         DEPLOYMENT SUMMARY"
            echo "=========================================="
            echo "✅ Environment: $ENVIRONMENT"
            echo "✅ Type: $DEPLOYMENT_TYPE"
            echo "✅ Version: $VERSION"
            echo "✅ Build: $BUILD_NUMBER"
            echo "✅ Status: SUCCESS"
            echo "✅ Deployment ID: $DEPLOYMENT_ID"
            echo "✅ Time: $(date)"
            echo "=========================================="
            echo
            
            log_success "Automated deployment completed successfully!"
            
        else
            log_error "Post-deployment verification failed"
            send_deployment_failure_notification
            exit 1
        fi
    else
        log_error "Deployment execution failed"
        send_deployment_failure_notification
        exit 1
    fi
}

# Handle signals for cleanup
trap 'log_error "Deployment interrupted"; send_deployment_failure_notification; exit 1' INT TERM

# Run main function
main "$@"
