#!/bin/bash

# Railway Deployment Script for Risk Assessment Service
# This script handles the deployment of the LSTM-enhanced risk assessment service to Railway

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="risk-assessment-service"
RAILWAY_PROJECT="kyb-platform"
ENVIRONMENT=${1:-"production"}
DOCKERFILE_PATH="Dockerfile"
RAILWAY_CONFIG="railway.json"

# Functions
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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI is not installed. Please install it first:"
        echo "npm install -g @railway/cli"
        exit 1
    fi
    
    # Check if we're logged into Railway
    if ! railway whoami &> /dev/null; then
        log_error "Not logged into Railway. Please run: railway login"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "$DOCKERFILE_PATH" ]; then
        log_error "Dockerfile not found. Please run this script from the service root directory."
        exit 1
    fi
    
    if [ ! -f "$RAILWAY_CONFIG" ]; then
        log_error "Railway configuration file not found: $RAILWAY_CONFIG"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

login_railway() {
    log_info "Logging into Railway..."
    
    if ! railway login; then
        log_error "Failed to login to Railway"
        exit 1
    fi
    
    log_success "Successfully logged into Railway"
}

setup_project() {
    log_info "Setting up Railway project..."
    
    # Check if project exists
    if ! railway status &> /dev/null; then
        log_info "Creating new Railway project..."
        if ! railway init; then
            log_error "Failed to initialize Railway project"
            exit 1
        fi
    else
        log_info "Using existing Railway project"
    fi
    
    log_success "Railway project setup complete"
}

build_and_test() {
    log_info "Validating application code..."
    
    # Check if Go is available for basic validation
    if command -v go &> /dev/null; then
        log_info "Running Go module validation..."
        if ! go mod tidy; then
            log_error "Go module validation failed"
            exit 1
        fi
        
        log_info "Running Go build validation..."
        if ! go build -o /tmp/risk-assessment-service ./cmd/main.go; then
            log_error "Go build validation failed"
            exit 1
        fi
        
        # Clean up test binary
        rm -f /tmp/risk-assessment-service
    else
        log_warning "Go not available for local validation, skipping build test"
    fi
    
    log_success "Code validation completed successfully"
}

deploy_to_railway() {
    log_info "Deploying to Railway ($ENVIRONMENT environment)..."
    
    # Set environment variables
    export RAILWAY_ENVIRONMENT=$ENVIRONMENT
    
    # Deploy using Railway CLI
    if ! railway up --detach; then
        log_error "Railway deployment failed"
        exit 1
    fi
    
    log_success "Deployment initiated successfully"
}

wait_for_deployment() {
    log_info "Waiting for deployment to complete..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Checking deployment status (attempt $attempt/$max_attempts)..."
        
        # Get service URL
        local service_url=$(railway domain 2>/dev/null || echo "")
        
        if [ -n "$service_url" ]; then
            # Test health endpoint
            if curl -f -s "https://$service_url/health" > /dev/null 2>&1; then
                log_success "Deployment completed successfully!"
                log_success "Service URL: https://$service_url"
                return 0
            fi
        fi
        
        sleep 10
        attempt=$((attempt + 1))
    done
    
    log_error "Deployment timeout - service may not be ready"
    return 1
}

run_smoke_tests() {
    log_info "Running smoke tests..."
    
    local service_url=$(railway domain 2>/dev/null || echo "")
    if [ -z "$service_url" ]; then
        log_error "Could not get service URL"
        return 1
    fi
    
    local base_url="https://$service_url"
    
    # Test health endpoint
    log_info "Testing health endpoint..."
    if ! curl -f -s "$base_url/health" > /dev/null; then
        log_error "Health endpoint test failed"
        return 1
    fi
    
    # Test metrics endpoint
    log_info "Testing metrics endpoint..."
    if ! curl -f -s "$base_url/metrics" > /dev/null; then
        log_warning "Metrics endpoint test failed (may not be implemented yet)"
    fi
    
    # Test API documentation
    log_info "Testing API documentation..."
    if ! curl -f -s "$base_url/docs" > /dev/null; then
        log_warning "API documentation test failed (may not be implemented yet)"
    fi
    
    log_success "Smoke tests completed"
}

show_deployment_info() {
    log_info "Deployment Information:"
    echo "========================"
    
    local service_url=$(railway domain 2>/dev/null || echo "Not available")
    echo "Service URL: https://$service_url"
    echo "Environment: $ENVIRONMENT"
    echo "Project: $RAILWAY_PROJECT"
    echo "Service: $SERVICE_NAME"
    echo ""
    
    log_info "Useful commands:"
    echo "  railway logs          # View logs"
    echo "  railway status        # Check status"
    echo "  railway variables     # View environment variables"
    echo "  railway shell         # Access service shell"
    echo "  railway redeploy      # Redeploy service"
    echo ""
    
    log_info "API Endpoints:"
    echo "  Health: https://$service_url/health"
    echo "  Metrics: https://$service_url/metrics"
    echo "  API Docs: https://$service_url/docs"
    echo "  Risk Assessment: https://$service_url/api/v1/assess"
    echo "  Advanced Prediction: https://$service_url/api/v1/risk/predict-advanced"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    # Clean up any temporary files if needed
}

# Main execution
main() {
    log_info "Starting Railway deployment for $SERVICE_NAME"
    log_info "Environment: $ENVIRONMENT"
    echo ""
    
    # Trap cleanup on exit
    trap cleanup EXIT
    
    # Execute deployment steps
    check_prerequisites
    login_railway
    setup_project
    build_and_test
    deploy_to_railway
    
    if wait_for_deployment; then
        run_smoke_tests
        show_deployment_info
        log_success "Deployment completed successfully!"
    else
        log_error "Deployment failed or timed out"
        exit 1
    fi
}

# Help function
show_help() {
    echo "Railway Deployment Script for Risk Assessment Service"
    echo ""
    echo "Usage: $0 [ENVIRONMENT]"
    echo ""
    echo "Arguments:"
    echo "  ENVIRONMENT    Deployment environment (default: production)"
    echo "                 Options: production, staging, development"
    echo ""
    echo "Examples:"
    echo "  $0                    # Deploy to production"
    echo "  $0 staging           # Deploy to staging"
    echo "  $0 development       # Deploy to development"
    echo ""
    echo "Prerequisites:"
    echo "  - Railway CLI installed (npm install -g @railway/cli)"
    echo "  - Railway account and project access"
    echo "  - Go 1.22+ (for local validation, optional)"
    echo ""
    echo "Environment Variables:"
    echo "  RAILWAY_TOKEN        Railway API token (optional, will prompt for login)"
    echo "  RAILWAY_PROJECT      Railway project name (default: kyb-platform)"
}

# Check for help flag
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Run main function
main "$@"
