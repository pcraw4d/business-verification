#!/bin/bash

# Railway Modular Deployment Script
# This script deploys the modular microservices architecture to Railway

set -e

# Configuration
RAILWAY_PROJECT_NAME="business-verification-modular"
RAILWAY_SERVICE_NAME="business-verification-modular"
DOCKERFILE_PATH="Dockerfile.modular"
RAILWAY_CONFIG_PATH="railway.modular.json"
ENVIRONMENT=${1:-production}

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI is not installed. Please install it first."
        log_info "Installation: npm install -g @railway/cli"
        exit 1
    fi
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    # Check if required files exist
    if [ ! -f "$DOCKERFILE_PATH" ]; then
        log_error "Dockerfile not found: $DOCKERFILE_PATH"
        exit 1
    fi
    
    if [ ! -f "$RAILWAY_CONFIG_PATH" ]; then
        log_error "Railway config not found: $RAILWAY_CONFIG_PATH"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Build the application
build_application() {
    log_info "Building modular application..."
    
    # Clean previous builds
    log_info "Cleaning previous builds..."
    docker system prune -f
    
    # Build the modular application
    log_info "Building Docker image..."
    docker build \
        --file "$DOCKERFILE_PATH" \
        --tag "$RAILWAY_SERVICE_NAME:latest" \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        .
    
    log_success "Application built successfully"
}

# Deploy to Railway
deploy_to_railway() {
    log_info "Deploying to Railway..."
    
    # Login to Railway (if not already logged in)
    log_info "Checking Railway authentication..."
    if ! railway whoami &> /dev/null; then
        log_info "Logging in to Railway..."
        railway login
    fi
    
    # Link to Railway project
    log_info "Linking to Railway project..."
    railway link --project "$RAILWAY_PROJECT_NAME"
    
    # Set environment variables
    log_info "Setting environment variables for $ENVIRONMENT..."
    set_environment_variables
    
    # Deploy the application
    log_info "Deploying application..."
    railway up --service "$RAILWAY_SERVICE_NAME"
    
    log_success "Deployment completed successfully"
}

# Set environment variables
set_environment_variables() {
    log_info "Setting environment variables..."
    
    # Core application variables
    railway variables set ENVIRONMENT="$ENVIRONMENT"
    railway variables set PORT="8080"
    railway variables set HOST="0.0.0.0"
    
    # Module enablement
    railway variables set ENABLE_MODULES="all"
    railway variables set ENABLE_WEBSITE_ANALYSIS="true"
    railway variables set ENABLE_WEB_SEARCH_ANALYSIS="true"
    railway variables set ENABLE_ML_CLASSIFICATION="true"
    railway variables set ENABLE_KEYWORD_CLASSIFICATION="true"
    railway variables set ENABLE_VERIFICATION="true"
    railway variables set ENABLE_DATA_EXTRACTION="true"
    railway variables set ENABLE_RISK_ASSESSMENT="true"
    railway variables set ENABLE_INTELLIGENT_ROUTING="true"
    
    # Error resilience configuration
    railway variables set CIRCUIT_BREAKER_ENABLED="true"
    railway variables set RETRY_POLICY_ENABLED="true"
    railway variables set FALLBACK_STRATEGY_ENABLED="true"
    railway variables set DEGRADATION_POLICY_ENABLED="true"
    
    # Observability configuration
    railway variables set LOG_LEVEL="info"
    railway variables set LOG_FORMAT="json"
    railway variables set METRICS_ENABLED="true"
    railway variables set HEALTH_CHECK_ENABLED="true"
    railway variables set PERFORMANCE_MONITORING_ENABLED="true"
    
    # Cache configuration
    railway variables set CACHE_ENABLED="true"
    railway variables set CACHE_TYPE="memory"
    
    # Security configuration
    railway variables set SECURITY_ENABLED="true"
    railway variables set SECURITY_CORS_ENABLED="true"
    railway variables set SECURITY_RATE_LIMITING_ENABLED="true"
    
    # Module-specific configuration
    railway variables set WEBSITE_ANALYSIS_TIMEOUT="30s"
    railway variables set WEB_SEARCH_ANALYSIS_TIMEOUT="20s"
    railway variables set ML_CLASSIFICATION_CONFIDENCE_THRESHOLD="0.7"
    railway variables set VERIFICATION_TIMEOUT="45s"
    railway variables set DATA_EXTRACTION_TIMEOUT="60s"
    railway variables set RISK_ASSESSMENT_TIMEOUT="30s"
    
    # Microservices configuration
    railway variables set SERVICE_DISCOVERY_ENABLED="true"
    railway variables set SERVICE_ISOLATION_ENABLED="true"
    railway variables set SERVICE_ISOLATION_LEVEL="enhanced"
    
    # Railway-specific variables
    railway variables set RAILWAY_ENVIRONMENT="$ENVIRONMENT"
    railway variables set RAILWAY_SERVICE_NAME="$RAILWAY_SERVICE_NAME"
    
    log_success "Environment variables set successfully"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    sleep 30
    
    # Get the deployment URL
    DEPLOYMENT_URL=$(railway status --service "$RAILWAY_SERVICE_NAME" --json | jq -r '.url')
    
    if [ -z "$DEPLOYMENT_URL" ] || [ "$DEPLOYMENT_URL" = "null" ]; then
        log_error "Failed to get deployment URL"
        return 1
    fi
    
    log_info "Deployment URL: $DEPLOYMENT_URL"
    
    # Test health endpoint
    log_info "Testing health endpoint..."
    for i in {1..10}; do
        if curl -f -s "$DEPLOYMENT_URL/health" > /dev/null; then
            log_success "Health check passed"
            break
        else
            log_warning "Health check failed, attempt $i/10"
            if [ $i -eq 10 ]; then
                log_error "Health check failed after 10 attempts"
                return 1
            fi
            sleep 10
        fi
    done
    
    # Test readiness endpoint
    log_info "Testing readiness endpoint..."
    if curl -f -s "$DEPLOYMENT_URL/ready" > /dev/null; then
        log_success "Readiness check passed"
    else
        log_error "Readiness check failed"
        return 1
    fi
    
    # Test liveness endpoint
    log_info "Testing liveness endpoint..."
    if curl -f -s "$DEPLOYMENT_URL/live" > /dev/null; then
        log_success "Liveness check passed"
    else
        log_error "Liveness check failed"
        return 1
    fi
    
    log_success "Deployment verification completed successfully"
    log_info "Application is available at: $DEPLOYMENT_URL"
}

# Monitor deployment
monitor_deployment() {
    log_info "Starting deployment monitoring..."
    
    DEPLOYMENT_URL=$(railway status --service "$RAILWAY_SERVICE_NAME" --json | jq -r '.url')
    
    if [ -z "$DEPLOYMENT_URL" ] || [ "$DEPLOYMENT_URL" = "null" ]; then
        log_error "Failed to get deployment URL for monitoring"
        return 1
    fi
    
    log_info "Monitoring deployment at: $DEPLOYMENT_URL"
    log_info "Press Ctrl+C to stop monitoring"
    
    # Monitor health for 5 minutes
    for i in {1..30}; do
        if curl -f -s "$DEPLOYMENT_URL/health" > /dev/null; then
            echo -e "${GREEN}✓${NC} Health check passed ($i/30)"
        else
            echo -e "${RED}✗${NC} Health check failed ($i/30)"
        fi
        sleep 10
    done
    
    log_success "Monitoring completed"
}

# Rollback deployment
rollback_deployment() {
    log_info "Rolling back deployment..."
    
    # Get previous deployment
    PREVIOUS_DEPLOYMENT=$(railway deployments --service "$RAILWAY_SERVICE_NAME" --json | jq -r '.[1].id')
    
    if [ -z "$PREVIOUS_DEPLOYMENT" ] || [ "$PREVIOUS_DEPLOYMENT" = "null" ]; then
        log_error "No previous deployment found for rollback"
        return 1
    fi
    
    log_info "Rolling back to deployment: $PREVIOUS_DEPLOYMENT"
    railway rollback --service "$RAILWAY_SERVICE_NAME" --deployment "$PREVIOUS_DEPLOYMENT"
    
    log_success "Rollback completed successfully"
}

# Show deployment status
show_status() {
    log_info "Showing deployment status..."
    
    railway status --service "$RAILWAY_SERVICE_NAME"
    
    # Show recent deployments
    log_info "Recent deployments:"
    railway deployments --service "$RAILWAY_SERVICE_NAME" --limit 5
}

# Show logs
show_logs() {
    log_info "Showing deployment logs..."
    
    railway logs --service "$RAILWAY_SERVICE_NAME" --follow
}

# Main deployment function
main() {
    log_info "Starting Railway modular deployment..."
    log_info "Environment: $ENVIRONMENT"
    log_info "Service: $RAILWAY_SERVICE_NAME"
    
    # Check prerequisites
    check_prerequisites
    
    # Build application
    build_application
    
    # Deploy to Railway
    deploy_to_railway
    
    # Verify deployment
    verify_deployment
    
    log_success "Railway modular deployment completed successfully!"
}

# Command line argument handling
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "build")
        check_prerequisites
        build_application
        ;;
    "verify")
        verify_deployment
        ;;
    "monitor")
        monitor_deployment
        ;;
    "rollback")
        rollback_deployment
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs
        ;;
    "help"|"-h"|"--help")
        echo "Railway Modular Deployment Script"
        echo ""
        echo "Usage: $0 [COMMAND] [ENVIRONMENT]"
        echo ""
        echo "Commands:"
        echo "  deploy    Deploy the application (default)"
        echo "  build     Build the application only"
        echo "  verify    Verify the deployment"
        echo "  monitor   Monitor the deployment"
        echo "  rollback  Rollback to previous deployment"
        echo "  status    Show deployment status"
        echo "  logs      Show deployment logs"
        echo "  help      Show this help message"
        echo ""
        echo "Environments:"
        echo "  production (default)"
        echo "  staging"
        echo "  development"
        echo ""
        echo "Examples:"
        echo "  $0 deploy production"
        echo "  $0 verify"
        echo "  $0 monitor"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
